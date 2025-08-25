package executor

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"xsha-backend/config"
	"xsha-backend/database"
	"xsha-backend/services"
	"xsha-backend/utils"
)

// BatchLogAppender provides batched log appending to reduce database operations
type BatchLogAppender struct {
	execLogID   uint
	logAppender LogAppender
	buffer      []string
	mutex       sync.Mutex
	ticker      *time.Ticker
	done        chan bool
	wg          sync.WaitGroup
	batchSize   int
}

func NewBatchLogAppender(execLogID uint, logAppender LogAppender) *BatchLogAppender {
	bla := &BatchLogAppender{
		execLogID:   execLogID,
		logAppender: logAppender,
		buffer:      make([]string, 0, 100),
		ticker:      time.NewTicker(1 * time.Second), // Flush every second
		done:        make(chan bool),
		batchSize:   50, // Flush after 50 lines
	}

	// Start the background flushing goroutine
	bla.wg.Add(1)
	go bla.flushRoutine()

	return bla
}

func (bla *BatchLogAppender) AppendLog(content string) {
	bla.mutex.Lock()
	defer bla.mutex.Unlock()

	bla.buffer = append(bla.buffer, content)

	// Flush if buffer is full
	if len(bla.buffer) >= bla.batchSize {
		bla.flushLocked()
	}
}

func (bla *BatchLogAppender) flushLocked() {
	if len(bla.buffer) == 0 {
		return
	}

	// Join all buffered log entries
	combined := strings.Join(bla.buffer, "")

	// Protect against extremely large combined logs (10MB limit for batch)
	const maxBatchSize = 10 * 1024 * 1024 // 10MB
	if len(combined) > maxBatchSize {
		combined = combined[:maxBatchSize-100] + "... [BATCH TRUNCATED DUE TO SIZE]\n"
		utils.Warn("Truncated large log batch", "original_size", len(strings.Join(bla.buffer, "")), "truncated_size", len(combined))
	}

	bla.logAppender.AppendLog(bla.execLogID, combined)

	// Clear buffer
	bla.buffer = bla.buffer[:0]
}

func (bla *BatchLogAppender) flushRoutine() {
	defer bla.wg.Done()

	for {
		select {
		case <-bla.ticker.C:
			bla.mutex.Lock()
			bla.flushLocked()
			bla.mutex.Unlock()
		case <-bla.done:
			// Final flush before exit
			bla.mutex.Lock()
			bla.flushLocked()
			bla.mutex.Unlock()
			return
		}
	}
}

func (bla *BatchLogAppender) Close() {
	bla.ticker.Stop()
	close(bla.done)
	bla.wg.Wait()

	// Final flush
	bla.mutex.Lock()
	bla.flushLocked()
	bla.mutex.Unlock()
}

type dockerExecutor struct {
	config        *config.Config
	logAppender   LogAppender
	configService services.SystemConfigService
}

func NewDockerExecutor(cfg *config.Config, logAppender LogAppender, configService services.SystemConfigService) DockerExecutor {
	return &dockerExecutor{
		config:        cfg,
		logAppender:   logAppender,
		configService: configService,
	}
}

func (d *dockerExecutor) CheckAvailability() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker command unavailable or docker daemon not running: %v", err)
	}

	return nil
}

func (d *dockerExecutor) escapeShellArg(arg string) string {
	return strconv.Quote(arg)
}

type buildDockerCommandOptions struct {
	containerName    string
	maskEnvVars      bool
	includeStdinFlag bool
}

func (d *dockerExecutor) buildDockerCommandCore(conv *database.TaskConversation, workspacePath string, opts buildDockerCommandOptions) string {
	devEnv := conv.Task.DevEnvironment

	envVars := make(map[string]string)
	if devEnv.EnvVars != "" {
		json.Unmarshal([]byte(devEnv.EnvVars), &envVars)
	}

	isInContainer := utils.IsRunningInContainer()

	cmd := []string{"docker", "run", "--rm"}

	if opts.includeStdinFlag {
		cmd = append(cmd, "-i")
	}

	if opts.containerName != "" {
		cmd = append(cmd, fmt.Sprintf("--name=%s", opts.containerName))
	}

	if isInContainer {
		cmd = append(cmd, "-v xsha_workspaces:/app")
		cmd = append(cmd, "-v xsha_dev_sessions:/xsha_dev_sessions")
		// workspacePath is now already relative, use it directly
		cmd = append(cmd, fmt.Sprintf("-w /app/%s", workspacePath))
	} else {
		// workspacePath is now relative, need to convert to absolute for volume mounting
		absoluteWorkspacePath := filepath.Join(d.config.WorkspaceBaseDir, workspacePath)
		cmd = append(cmd, fmt.Sprintf("-v %s:/app/%s", absoluteWorkspacePath, workspacePath))
		if devEnv.SessionDir != "" {
			// SessionDir is now also relative, convert to absolute for volume mounting
			absoluteSessionDir := filepath.Join(d.config.DevSessionsDir, devEnv.SessionDir)
			cmd = append(cmd, fmt.Sprintf("-v %s:/home/xsha", absoluteSessionDir))
		}
		cmd = append(cmd, fmt.Sprintf("-w /app/%s", workspacePath))
	}

	if devEnv.CPULimit > 0 {
		cmd = append(cmd, fmt.Sprintf("--cpus=%.2f", devEnv.CPULimit))
	}
	if devEnv.MemoryLimit > 0 {
		cmd = append(cmd, fmt.Sprintf("--memory=%dm", devEnv.MemoryLimit))
	}

	for key, value := range envVars {
		if opts.maskEnvVars {
			value = utils.MaskSensitiveValue(value)
		}
		cmd = append(cmd, fmt.Sprintf("-e %s=%s", key, value))
	}

	imageName := devEnv.DockerImage
	aiCommand := d.buildAICommand(devEnv.Type, conv.Content, isInContainer, conv.Task, devEnv, conv)

	cmd = append(cmd, imageName)
	cmd = append(cmd, aiCommand...)

	return strings.Join(cmd, " ")
}

func (d *dockerExecutor) buildAICommand(envType, content string, isInContainer bool, task *database.Task, devEnv *database.DevEnvironment, conv *database.TaskConversation) []string {
	var baseCommand []string

	switch envType {
	case "claude-code":
		claudeCommand := []string{
			"claude",
			"-p",
			"--output-format=stream-json",
			"--dangerously-skip-permissions",
			"--verbose",
		}

		if task.SessionID != "" {
			claudeCommand = append(claudeCommand, "-r", task.SessionID)
		}

		// Parse env_params to check for model and plan mode parameters
		if conv != nil && conv.EnvParams != "" && conv.EnvParams != "{}" {
			var envParams map[string]interface{}
			if err := json.Unmarshal([]byte(conv.EnvParams), &envParams); err == nil {
				if model, exists := envParams["model"]; exists {
					if modelStr, ok := model.(string); ok && modelStr != "default" {
						claudeCommand = append(claudeCommand, "--model", modelStr)
					}
				}

				if isPlanMode, exists := envParams["is_plan_mode"]; exists {
					if isPlanModeBool, ok := isPlanMode.(bool); ok && isPlanModeBool {
						claudeCommand = append(claudeCommand, "--permission-mode plan")
					}
				}
			}
		}

		// Add system prompt if project has one
		if task.Project != nil && task.Project.SystemPrompt != "" {
			claudeCommand = append(claudeCommand, "--append-system-prompt", d.escapeShellArg(task.Project.SystemPrompt))
		}

		// Add system prompt if dev environment has one
		if devEnv != nil && devEnv.SystemPrompt != "" {
			claudeCommand = append(claudeCommand, "--append-system-prompt", d.escapeShellArg(devEnv.SystemPrompt))
		}

		claudeCommand = append(claudeCommand, d.escapeShellArg(content))

		if isInContainer && devEnv.SessionDir != "" {
			// SessionDir is now already relative, use it directly
			baseCommand = append(baseCommand, "-d", "/xsha_dev_sessions/"+devEnv.SessionDir)
		}

		claudeCommandStr := strings.Join(claudeCommand, " ")
		baseCommand = append(baseCommand, "--command", d.escapeShellArg(claudeCommandStr))
	case "opencode", "gemini-cli":
		baseCommand = []string{d.escapeShellArg(content)}
	}

	return baseCommand
}

func (d *dockerExecutor) BuildCommandForLog(conv *database.TaskConversation, workspacePath string) string {
	return d.buildDockerCommandCore(conv, workspacePath, buildDockerCommandOptions{
		containerName:    "",
		maskEnvVars:      true,
		includeStdinFlag: false,
	})
}

func (d *dockerExecutor) ExecuteWithContext(ctx context.Context, dockerCmd string, execLogID uint) error {
	if err := d.CheckAvailability(); err != nil {
		d.logAppender.AppendLog(execLogID, fmt.Sprintf("❌ Docker unavailable: %v\n", err))
		return fmt.Errorf("docker unavailable: %v", err)
	}

	d.logAppender.AppendLog(execLogID, "✅ Docker availability check passed\n")

	timeout, err := d.configService.GetDockerTimeout()
	if err != nil {
		utils.Warn("Failed to get Docker timeout from system config, using default 120 minutes", "error", err)
		timeout = 120 * time.Minute
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", dockerCmd)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	var stderrLines []string
	var mu sync.Mutex

	// Create batch log appenders for better performance with large logs
	stdoutBatcher := NewBatchLogAppender(execLogID, d.logAppender)
	stderrBatcher := NewBatchLogAppender(execLogID, d.logAppender)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		defer stdoutBatcher.Close()
		d.readPipeWithBatcher(stdout, stdoutBatcher, "STDOUT")
	}()

	go func() {
		defer wg.Done()
		defer stderrBatcher.Close()
		d.readPipeWithErrorCaptureAndBatcher(stderr, stderrBatcher, "STDERR", &stderrLines, &mu)
	}()

	err = cmd.Wait()

	// Wait for all log processing to complete before returning
	wg.Wait()
	if err != nil && len(stderrLines) > 0 {
		mu.Lock()
		errorLines := make([]string, len(stderrLines))
		copy(errorLines, stderrLines)
		mu.Unlock()

		if len(errorLines) > 0 {
			errorMsg := strings.Join(errorLines, "\n")
			if len(errorMsg) > 1000 {
				errorMsg = errorMsg[:1000] + "..."
			}
			return fmt.Errorf("%s", errorMsg)
		}
	}
	return err
}

func (d *dockerExecutor) readPipeWithBatcher(pipe interface{}, batcher *BatchLogAppender, prefix string) {
	scanner := bufio.NewScanner(pipe.(interface{ Read([]byte) (int, error) }))
	// Set larger buffer to handle large log outputs (1MB buffer)
	const maxCapacity = 1024 * 1024 // 1MB
	buf := make([]byte, 0, 64*1024) // Start with 64KB
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		line := scanner.Text()

		// Protect against extremely large log lines
		if len(line) > maxCapacity {
			line = line[:maxCapacity-3] + "..."
			utils.Warn("Truncated extremely large log line", "prefix", prefix, "original_length", len(scanner.Text()))
		}

		logLine := fmt.Sprintf("[%s] %s: %s\n", utils.Now().Format("15:04:05"), prefix, line)
		batcher.AppendLog(logLine)
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		errorLine := fmt.Sprintf("[%s] %s: ERROR - Scanner failed: %v\n", utils.Now().Format("15:04:05"), prefix, err)
		batcher.AppendLog(errorLine)
		utils.Error("Log scanner failed", "prefix", prefix, "error", err)
	}
}

func (d *dockerExecutor) readPipeWithErrorCaptureAndBatcher(pipe interface{}, batcher *BatchLogAppender, prefix string, errorLines *[]string, mu *sync.Mutex) {
	scanner := bufio.NewScanner(pipe.(interface{ Read([]byte) (int, error) }))
	// Set larger buffer to handle large log outputs (1MB buffer)
	const maxCapacity = 1024 * 1024 // 1MB
	buf := make([]byte, 0, 64*1024) // Start with 64KB
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		line := scanner.Text()

		// Protect against extremely large log lines
		if len(line) > maxCapacity {
			line = line[:maxCapacity-3] + "..."
			utils.Warn("Truncated extremely large log line", "prefix", prefix, "original_length", len(scanner.Text()))
		}

		logLine := fmt.Sprintf("[%s] %s: %s\n", utils.Now().Format("15:04:05"), prefix, line)
		batcher.AppendLog(logLine)

		if prefix == "STDERR" {
			mu.Lock()
			*errorLines = append(*errorLines, line)
			mu.Unlock()
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		errorLine := fmt.Sprintf("[%s] %s: ERROR - Scanner failed: %v\n", utils.Now().Format("15:04:05"), prefix, err)
		batcher.AppendLog(errorLine)
		utils.Error("Log scanner failed", "prefix", prefix, "error", err)

		// If this is STDERR scanner and it failed, add the error to errorLines too
		if prefix == "STDERR" {
			mu.Lock()
			*errorLines = append(*errorLines, fmt.Sprintf("Scanner failed: %v", err))
			mu.Unlock()
		}
	}
}

// generateContainerName creates a unique container name for the conversation
func (d *dockerExecutor) generateContainerName(conv *database.TaskConversation) string {
	return fmt.Sprintf("xsha-task-%d-conv-%d", conv.TaskID, conv.ID)
}

// BuildCommandWithContainerName builds the docker command with a specific container name
func (d *dockerExecutor) BuildCommandWithContainerName(conv *database.TaskConversation, workspacePath string) string {
	containerName := d.generateContainerName(conv)
	return d.buildDockerCommandCore(conv, workspacePath, buildDockerCommandOptions{
		containerName:    containerName,
		maskEnvVars:      false,
		includeStdinFlag: true,
	})
}

// ExecuteWithContainerTracking executes docker command with container tracking for proper cleanup
func (d *dockerExecutor) ExecuteWithContainerTracking(ctx context.Context, conv *database.TaskConversation, workspacePath string, execLogID uint) (string, error) {
	if err := d.CheckAvailability(); err != nil {
		d.logAppender.AppendLog(execLogID, fmt.Sprintf("❌ Docker unavailable: %v\n", err))
		return "", fmt.Errorf("docker unavailable: %v", err)
	}

	d.logAppender.AppendLog(execLogID, "✅ Docker availability check passed\n")

	timeout, err := d.configService.GetDockerTimeout()
	if err != nil {
		utils.Warn("Failed to get Docker timeout from system config, using default 120 minutes", "error", err)
		timeout = 120 * time.Minute
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	containerName := d.generateContainerName(conv)
	dockerCmd := d.BuildCommandWithContainerName(conv, workspacePath)

	d.logAppender.AppendLog(execLogID, fmt.Sprintf("🐳 Starting container: %s\n", containerName))

	cmd := exec.CommandContext(ctx, "sh", "-c", dockerCmd)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}

	if err := cmd.Start(); err != nil {
		return "", err
	}

	var stderrLines []string
	var mu sync.Mutex

	// Create batch log appenders for better performance with large logs
	stdoutBatcher := NewBatchLogAppender(execLogID, d.logAppender)
	stderrBatcher := NewBatchLogAppender(execLogID, d.logAppender)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		defer stdoutBatcher.Close()
		d.readPipeWithBatcher(stdout, stdoutBatcher, "STDOUT")
	}()

	go func() {
		defer wg.Done()
		defer stderrBatcher.Close()
		d.readPipeWithErrorCaptureAndBatcher(stderr, stderrBatcher, "STDERR", &stderrLines, &mu)
	}()

	err = cmd.Wait()

	// Wait for all log processing to complete before updating status
	wg.Wait()

	// If context was cancelled, ensure container cleanup
	select {
	case <-ctx.Done():
		d.logAppender.AppendLog(execLogID, fmt.Sprintf("⚠️ Execution cancelled, cleaning up container: %s\n", containerName))
		if cleanupErr := d.StopAndRemoveContainer(containerName); cleanupErr != nil {
			d.logAppender.AppendLog(execLogID, fmt.Sprintf("❌ Failed to cleanup container: %v\n", cleanupErr))
			utils.Error("Failed to cleanup cancelled container", "container", containerName, "error", cleanupErr)
		} else {
			d.logAppender.AppendLog(execLogID, fmt.Sprintf("✅ Container cleaned up successfully: %s\n", containerName))
		}
	default:
	}

	if err != nil && len(stderrLines) > 0 {
		mu.Lock()
		errorLines := make([]string, len(stderrLines))
		copy(errorLines, stderrLines)
		mu.Unlock()

		if len(errorLines) > 0 {
			errorMsg := strings.Join(errorLines, "\n")
			if len(errorMsg) > 1000 {
				errorMsg = errorMsg[:1000] + "..."
			}
			return containerName, fmt.Errorf("%s", errorMsg)
		}
	}
	return containerName, err
}

// StopAndRemoveContainer stops and removes a Docker container by name or ID
func (d *dockerExecutor) StopAndRemoveContainer(containerID string) error {
	// First try to stop the container gracefully
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer stopCancel()

	stopCmd := exec.CommandContext(stopCtx, "docker", "stop", containerID)
	if err := stopCmd.Run(); err != nil {
		utils.Warn("Failed to stop container gracefully, will try force removal", "container", containerID, "error", err)
	}

	// Then remove the container (force remove if needed)
	removeCtx, removeCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer removeCancel()

	removeCmd := exec.CommandContext(removeCtx, "docker", "rm", "-f", containerID)
	if err := removeCmd.Run(); err != nil {
		// Check if container doesn't exist (which is fine)
		if strings.Contains(err.Error(), "No such container") {
			utils.Info("Container already removed or doesn't exist", "container", containerID)
			return nil
		}
		return fmt.Errorf("failed to remove container %s: %v", containerID, err)
	}

	utils.Info("Container stopped and removed successfully", "container", containerID)
	return nil
}
