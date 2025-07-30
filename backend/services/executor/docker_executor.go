package executor

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
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

// NewDockerExecutor 创建Docker执行器
func NewDockerExecutor(cfg *config.Config, logAppender LogAppender, configService services.SystemConfigService) DockerExecutor {
	return &dockerExecutor{
		config:        cfg,
		logAppender:   logAppender,
		configService: configService,
	}
}

// CheckAvailability 检查Docker可用性
func (d *dockerExecutor) CheckAvailability() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 检查 Docker 守护进程是否可用
	cmd := exec.CommandContext(ctx, "docker", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker 命令不可用或 docker 守护进程未运行: %v", err)
	}

	return nil
}

// BuildCommand 构建Docker命令
func (d *dockerExecutor) BuildCommand(conv *database.TaskConversation, workspacePath string) string {
	devEnv := conv.Task.DevEnvironment

	// 解析环境变量
	envVars := make(map[string]string)
	if devEnv.EnvVars != "" {
		json.Unmarshal([]byte(devEnv.EnvVars), &envVars)
	}

	// 构建基础命令
	cmd := []string{
		"docker", "run", "--rm",
		fmt.Sprintf("-v %s:/app", workspacePath),
	}

	// 添加资源限制
	if devEnv.CPULimit > 0 {
		cmd = append(cmd, fmt.Sprintf("--cpus=%.2f", devEnv.CPULimit))
	}
	if devEnv.MemoryLimit > 0 {
		cmd = append(cmd, fmt.Sprintf("--memory=%dm", devEnv.MemoryLimit))
	}

	// 添加环境变量
	for key, value := range envVars {
		cmd = append(cmd, fmt.Sprintf("-e %s=%s", key, value))
	}

	// 根据开发环境类型选择镜像和命令
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
			conv.Content,
		}
	case "opencode":
		aiCommand = []string{conv.Content}
	case "gemini_cli":
		aiCommand = []string{conv.Content}
	default:
		// 默认使用 claude-code 命令
		aiCommand = []string{
			"claude",
			"-p",
			"--output-format=stream-json",
			"--dangerously-skip-permissions",
			"--verbose",
			conv.Content,
		}
	}

	// 添加镜像名称
	cmd = append(cmd, imageName)

	// 添加 AI 命令参数
	cmd = append(cmd, aiCommand...)

	return strings.Join(cmd, " ")
}

// getImageNameFromConfig 根据开发环境类型从系统配置获取镜像名称
func (d *dockerExecutor) getImageNameFromConfig(envType database.DevEnvironmentType) string {
	// 从系统配置获取环境类型映射
	envTypesJSON, err := d.configService.GetValue("dev_environment_types")
	if err != nil {
		// 如果获取配置失败，使用默认配置
		return "claude-code:latest"
	}

	// 解析环境类型配置
	var envTypes []map[string]interface{}
	if err := json.Unmarshal([]byte(envTypesJSON), &envTypes); err != nil {
		// 如果解析失败，使用默认配置
		return "claude-code:latest"
	}

	// 根据环境类型配置选择镜像
	for _, envTypeConfig := range envTypes {
		if name, ok := envTypeConfig["name"].(string); ok && name == string(envType) {
			if image, ok := envTypeConfig["image"].(string); ok {
				return image
			}
		}
	}

	// 如果找不到匹配的配置，使用默认配置
	return "claude-code:latest"
}

// BuildCommandForLog 构建用于日志的Docker命令
func (d *dockerExecutor) BuildCommandForLog(conv *database.TaskConversation, workspacePath string) string {
	devEnv := conv.Task.DevEnvironment

	// 解析环境变量
	envVars := make(map[string]string)
	if devEnv.EnvVars != "" {
		json.Unmarshal([]byte(devEnv.EnvVars), &envVars)
	}

	// 构建基础命令
	cmd := []string{
		"docker", "run", "--rm",
		fmt.Sprintf("-v %s:/app", workspacePath),
	}

	// 添加资源限制
	if devEnv.CPULimit > 0 {
		cmd = append(cmd, fmt.Sprintf("--cpus=%.2f", devEnv.CPULimit))
	}
	if devEnv.MemoryLimit > 0 {
		cmd = append(cmd, fmt.Sprintf("--memory=%dm", devEnv.MemoryLimit))
	}

	// 添加环境变量（值已打码）
	for key, value := range envVars {
		maskedValue := utils.MaskSensitiveValue(value)
		cmd = append(cmd, fmt.Sprintf("-e %s=%s", key, maskedValue))
	}

	// 根据开发环境类型选择镜像和命令
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
			conv.Content,
		}
	case "opencode":
		aiCommand = []string{conv.Content}
	case "gemini_cli":
		aiCommand = []string{conv.Content}
	default:
		// 默认使用 claude-code 命令
		aiCommand = []string{
			"claude",
			"-p",
			"--output-format=stream-json",
			"--dangerously-skip-permissions",
			"--verbose",
			conv.Content,
		}
	}

	// 添加镜像名称
	cmd = append(cmd, imageName)

	// 添加 AI 命令参数
	cmd = append(cmd, aiCommand...)

	return strings.Join(cmd, " ")
}

// ExecuteWithContext 执行Docker命令
func (d *dockerExecutor) ExecuteWithContext(ctx context.Context, dockerCmd string, execLogID uint) error {
	// 首先检查 Docker 是否可用
	if err := d.CheckAvailability(); err != nil {
		d.logAppender.AppendLog(execLogID, fmt.Sprintf("❌ Docker 不可用: %v\n", err))
		return fmt.Errorf("docker 不可用: %v", err)
	}

	d.logAppender.AppendLog(execLogID, "✅ Docker 可用性检查通过\n")

	// 解析超时时间
	timeout, err := time.ParseDuration(d.config.DockerExecutionTimeout)
	if err != nil {
		utils.Warn("解析Docker超时时间失败，使用默认值30分钟", "error", err)
		timeout = 30 * time.Minute
	}

	ctx, cancel := context.WithTimeout(ctx, timeout) // 使用传入的上下文
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", dockerCmd)

	// 获取输出管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		return err
	}

	// 实时读取输出和错误信息
	var stderrLines []string
	var mu sync.Mutex

	go d.readPipe(stdout, execLogID, "STDOUT")
	go d.readPipeWithErrorCapture(stderr, execLogID, "STDERR", &stderrLines, &mu)

	// 等待命令完成
	err = cmd.Wait()
	if err != nil && len(stderrLines) > 0 {
		// 将 STDERR 中的错误信息合并作为错误消息
		mu.Lock()
		errorLines := make([]string, len(stderrLines))
		copy(errorLines, stderrLines)
		mu.Unlock()

		if len(errorLines) > 0 {
			errorMsg := strings.Join(errorLines, "\n")
			// 限制错误信息长度，避免过长
			if len(errorMsg) > 1000 {
				errorMsg = errorMsg[:1000] + "..."
			}
			return fmt.Errorf("%s", errorMsg)
		}
	}
	return err
}

// readPipe 读取管道输出
func (d *dockerExecutor) readPipe(pipe interface{}, execLogID uint, prefix string) {
	scanner := bufio.NewScanner(pipe.(interface{ Read([]byte) (int, error) }))
	for scanner.Scan() {
		line := scanner.Text()
		logLine := fmt.Sprintf("[%s] %s: %s\n", time.Now().Format("15:04:05"), prefix, line)
		d.logAppender.AppendLog(execLogID, logLine)
	}
}

// readPipeWithErrorCapture 读取管道输出并捕获错误信息
func (d *dockerExecutor) readPipeWithErrorCapture(pipe interface{}, execLogID uint, prefix string, errorLines *[]string, mu *sync.Mutex) {
	scanner := bufio.NewScanner(pipe.(interface{ Read([]byte) (int, error) }))
	for scanner.Scan() {
		line := scanner.Text()
		logLine := fmt.Sprintf("[%s] %s: %s\n", time.Now().Format("15:04:05"), prefix, line)
		d.logAppender.AppendLog(execLogID, logLine)

		// 如果是 STDERR，捕获错误信息
		if prefix == "STDERR" {
			mu.Lock()
			*errorLines = append(*errorLines, line)
			mu.Unlock()
		}
	}
}
