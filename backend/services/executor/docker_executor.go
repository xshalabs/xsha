package executor

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
	"xsha-backend/config"
	"xsha-backend/database"
	"xsha-backend/services"
	"xsha-backend/utils"
)

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

// buildDockerCommandOptions represents options for building docker commands
type buildDockerCommandOptions struct {
	containerName    string // empty means no container name
	maskEnvVars      bool   // whether to mask environment variables for logging
	includeStdinFlag bool   // whether to include -i flag
}

// buildDockerCommandCore is the core method that builds docker commands with shared logic
func (d *dockerExecutor) buildDockerCommandCore(conv *database.TaskConversation, workspacePath string, opts buildDockerCommandOptions) string {
	devEnv := conv.Task.DevEnvironment

	envVars := make(map[string]string)
	if devEnv.EnvVars != "" {
		json.Unmarshal([]byte(devEnv.EnvVars), &envVars)
	}

	// Check if running in container
	isInContainer := utils.IsRunningInContainer()

	// Build base command
	cmd := []string{"docker", "run", "--rm"}

	// Add stdin flag if needed
	if opts.includeStdinFlag {
		cmd = append(cmd, "-i")
	}

	// Add container name if specified
	if opts.containerName != "" {
		cmd = append(cmd, fmt.Sprintf("--name=%s", opts.containerName))
	}

	// Add volume mapping and working directory based on environment
	if isInContainer {
		// When running in container, use named volume
		cmd = append(cmd, "-v xsha_workspaces:/app")
		// Add dev sessions volume for container environment
		cmd = append(cmd, "-v xsha_dev_sessions:/xsha_dev_sessions")
		// Set working directory to the specific workspace path
		workspaceRelPath := utils.ExtractWorkspaceRelativePath(workspacePath)
		cmd = append(cmd, fmt.Sprintf("-w /app/%s", workspaceRelPath))
	} else {
		// When running on host, use direct path mapping
		cmd = append(cmd, fmt.Sprintf("-v %s:/app", workspacePath))
		// Add session directory mapping
		if devEnv.SessionDir != "" {
			cmd = append(cmd, fmt.Sprintf("-v %s:/home/xsha", devEnv.SessionDir))
		}
		// Set working directory to /app (the mounted workspace)
		cmd = append(cmd, "-w /app")
	}

	// Add resource limits
	if devEnv.CPULimit > 0 {
		cmd = append(cmd, fmt.Sprintf("--cpus=%.2f", devEnv.CPULimit))
	}
	if devEnv.MemoryLimit > 0 {
		cmd = append(cmd, fmt.Sprintf("--memory=%dm", devEnv.MemoryLimit))
	}

	// Add environment variables
	for key, value := range envVars {
		if opts.maskEnvVars {
			value = utils.MaskSensitiveValue(value)
		}
		cmd = append(cmd, fmt.Sprintf("-e %s=%s", key, value))
	}

	// Get image name from dev environment, fallback to config if empty
	imageName := devEnv.DockerImage
	aiCommand := d.buildAICommand(devEnv.Type, conv.Content, isInContainer, conv.Task, devEnv)

	cmd = append(cmd, imageName)
	cmd = append(cmd, aiCommand...)

	return strings.Join(cmd, " ")
}

// buildAICommand builds the AI-specific command based on environment type
func (d *dockerExecutor) buildAICommand(envType, content string, isInContainer bool, task *database.Task, devEnv *database.DevEnvironment) []string {
	var baseCommand []string

	switch envType {
	case "claude_code":
		// Build claude base command
		claudeCommand := []string{
			"claude",
			"-p",
			"--output-format=stream-json",
			"--dangerously-skip-permissions",
			"--verbose",
		}

		// Add session_id parameter if it exists
		if task.SessionID != "" {
			claudeCommand = append(claudeCommand, "-s", task.SessionID)
		}

		// Add content to the command
		claudeCommand = append(claudeCommand, d.escapeShellArg(content))

		// Add session_dir parameter if running in container and session_dir exists
		if isInContainer && devEnv.SessionDir != "" {
			baseCommand = append(baseCommand, "-d", "/xsha_dev_sessions/"+utils.ExtractDevSessionRelativePath(devEnv.SessionDir))
		}

		// Add --command parameter with claude command as its value
		claudeCommandStr := strings.Join(claudeCommand, " ")
		baseCommand = append(baseCommand, "--command", d.escapeShellArg(claudeCommandStr))
	case "opencode", "gemini_cli":
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

	go d.readPipe(stdout, execLogID, "STDOUT")
	go d.readPipeWithErrorCapture(stderr, execLogID, "STDERR", &stderrLines, &mu)

	err = cmd.Wait()
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

func (d *dockerExecutor) readPipe(pipe interface{}, execLogID uint, prefix string) {
	scanner := bufio.NewScanner(pipe.(interface{ Read([]byte) (int, error) }))
	for scanner.Scan() {
		line := scanner.Text()
		logLine := fmt.Sprintf("[%s] %s: %s\n", time.Now().Format("15:04:05"), prefix, line)
		d.logAppender.AppendLog(execLogID, logLine)
	}
}

func (d *dockerExecutor) readPipeWithErrorCapture(pipe interface{}, execLogID uint, prefix string, errorLines *[]string, mu *sync.Mutex) {
	scanner := bufio.NewScanner(pipe.(interface{ Read([]byte) (int, error) }))
	for scanner.Scan() {
		line := scanner.Text()
		logLine := fmt.Sprintf("[%s] %s: %s\n", time.Now().Format("15:04:05"), prefix, line)
		d.logAppender.AppendLog(execLogID, logLine)

		if prefix == "STDERR" {
			mu.Lock()
			*errorLines = append(*errorLines, line)
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

	go d.readPipe(stdout, execLogID, "STDOUT")
	go d.readPipeWithErrorCapture(stderr, execLogID, "STDERR", &stderrLines, &mu)

	err = cmd.Wait()

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
