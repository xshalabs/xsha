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

func (d *dockerExecutor) BuildCommand(conv *database.TaskConversation, workspacePath string) string {
	devEnv := conv.Task.DevEnvironment

	envVars := make(map[string]string)
	if devEnv.EnvVars != "" {
		json.Unmarshal([]byte(devEnv.EnvVars), &envVars)
	}

	cmd := []string{
		"docker", "run", "--rm", "-i",
		fmt.Sprintf("-v %s:/app", workspacePath),
	}

	if devEnv.CPULimit > 0 {
		cmd = append(cmd, fmt.Sprintf("--cpus=%.2f", devEnv.CPULimit))
	}
	if devEnv.MemoryLimit > 0 {
		cmd = append(cmd, fmt.Sprintf("--memory=%dm", devEnv.MemoryLimit))
	}

	for key, value := range envVars {
		cmd = append(cmd, fmt.Sprintf("-e %s=%s", key, value))
	}

	imageName := d.getImageNameFromConfig(devEnv.Type)
	var aiCommand []string

	switch devEnv.Type {
	case "claude_code":
		aiCommand = []string{
			"claude",
			"-p",
			"--output-format=stream-json",
			"--dangerously-skip-permissions",
			"--verbose",
			d.escapeShellArg(conv.Content),
		}
	case "opencode":
		aiCommand = []string{d.escapeShellArg(conv.Content)}
	case "gemini_cli":
		aiCommand = []string{d.escapeShellArg(conv.Content)}
	default:
		aiCommand = []string{
			"claude",
			"-p",
			"--output-format=stream-json",
			"--dangerously-skip-permissions",
			"--verbose",
			d.escapeShellArg(conv.Content),
		}
	}

	cmd = append(cmd, imageName)

	cmd = append(cmd, aiCommand...)

	return strings.Join(cmd, " ")
}

func (d *dockerExecutor) getImageNameFromConfig(envType string) string {
	envTypesJSON, err := d.configService.GetValue("dev_environment_types")
	if err != nil {
		return "claude-code:latest"
	}

	var envTypes []map[string]interface{}
	if err := json.Unmarshal([]byte(envTypesJSON), &envTypes); err != nil {
		return "claude-code:latest"
	}

	for _, envTypeConfig := range envTypes {
		if key, ok := envTypeConfig["key"].(string); ok && key == envType {
			if image, ok := envTypeConfig["image"].(string); ok {
				return image
			}
		}
	}

	return "claude-code:latest"
}

func (d *dockerExecutor) BuildCommandForLog(conv *database.TaskConversation, workspacePath string) string {
	devEnv := conv.Task.DevEnvironment

	envVars := make(map[string]string)
	if devEnv.EnvVars != "" {
		json.Unmarshal([]byte(devEnv.EnvVars), &envVars)
	}

	cmd := []string{
		"docker", "run", "--rm",
		fmt.Sprintf("-v %s:/app", workspacePath),
	}

	if devEnv.CPULimit > 0 {
		cmd = append(cmd, fmt.Sprintf("--cpus=%.2f", devEnv.CPULimit))
	}
	if devEnv.MemoryLimit > 0 {
		cmd = append(cmd, fmt.Sprintf("--memory=%dm", devEnv.MemoryLimit))
	}

	for key, value := range envVars {
		maskedValue := utils.MaskSensitiveValue(value)
		cmd = append(cmd, fmt.Sprintf("-e %s=%s", key, maskedValue))
	}

	imageName := d.getImageNameFromConfig(devEnv.Type)
	var aiCommand []string

	switch devEnv.Type {
	case "claude_code":
		aiCommand = []string{
			"claude",
			"-p",
			"--output-format=stream-json",
			"--dangerously-skip-permissions",
			"--verbose",
			d.escapeShellArg(conv.Content),
		}
	case "opencode":
		aiCommand = []string{d.escapeShellArg(conv.Content)}
	case "gemini_cli":
		aiCommand = []string{d.escapeShellArg(conv.Content)}
	default:
		aiCommand = []string{
			"claude",
			"-p",
			"--output-format=stream-json",
			"--dangerously-skip-permissions",
			"--verbose",
			d.escapeShellArg(conv.Content),
		}
	}

	cmd = append(cmd, imageName)

	cmd = append(cmd, aiCommand...)

	return strings.Join(cmd, " ")
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
