package services

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"sleep0-backend/config"
	"sleep0-backend/database"
	"sleep0-backend/repository"
	"sleep0-backend/utils"
	"strings"
	"time"
)

type aiTaskExecutorService struct {
	taskConvRepo     repository.TaskConversationRepository
	execLogRepo      repository.TaskExecutionLogRepository
	workspaceManager *utils.WorkspaceManager
	gitCredService   GitCredentialService
	config           *config.Config
}

// NewAITaskExecutorService 创建AI任务执行服务
func NewAITaskExecutorService(
	taskConvRepo repository.TaskConversationRepository,
	execLogRepo repository.TaskExecutionLogRepository,
	gitCredService GitCredentialService,
	cfg *config.Config,
) AITaskExecutorService {
	return &aiTaskExecutorService{
		taskConvRepo:     taskConvRepo,
		execLogRepo:      execLogRepo,
		workspaceManager: utils.NewWorkspaceManager(cfg.WorkspaceBaseDir),
		gitCredService:   gitCredService,
		config:           cfg,
	}
}

// ProcessPendingConversations 处理待处理的对话
func (s *aiTaskExecutorService) ProcessPendingConversations() error {
	conversations, err := s.taskConvRepo.GetPendingConversationsWithDetails()
	if err != nil {
		return fmt.Errorf("获取待处理对话失败: %v", err)
	}

	log.Printf("发现 %d 个待处理的对话", len(conversations))

	for _, conv := range conversations {
		if err := s.processConversation(&conv); err != nil {
			log.Printf("处理对话 %d 失败: %v", conv.ID, err)
		}
	}

	return nil
}

// GetExecutionLog 获取执行日志
func (s *aiTaskExecutorService) GetExecutionLog(conversationID uint) (*database.TaskExecutionLog, error) {
	return s.execLogRepo.GetByConversationID(conversationID)
}

// CancelExecution 取消执行
func (s *aiTaskExecutorService) CancelExecution(conversationID uint) error {
	log, err := s.execLogRepo.GetByConversationID(conversationID)
	if err != nil {
		return fmt.Errorf("获取执行日志失败: %v", err)
	}

	if log.Status != database.TaskExecStatusPending && log.Status != database.TaskExecStatusRunning {
		return fmt.Errorf("只能取消待处理或执行中的任务")
	}

	// 更新执行状态为已取消
	return s.execLogRepo.UpdateStatus(log.ID, database.TaskExecStatusCancelled)
}

// processConversation 处理单个对话
func (s *aiTaskExecutorService) processConversation(conv *database.TaskConversation) error {
	// 验证关联数据
	if conv.Task == nil || conv.Task.Project == nil || conv.Task.DevEnvironment == nil {
		return fmt.Errorf("对话关联数据不完整")
	}

	// 更新对话状态为 running
	conv.Status = database.ConversationStatusRunning
	if err := s.taskConvRepo.Update(conv); err != nil {
		return fmt.Errorf("更新对话状态失败: %v", err)
	}

	// 创建执行日志
	execLog := &database.TaskExecutionLog{
		ConversationID: conv.ID,
		Status:         database.TaskExecStatusPending,
		CreatedBy:      conv.CreatedBy,
	}
	if err := s.execLogRepo.Create(execLog); err != nil {
		return fmt.Errorf("创建执行日志失败: %v", err)
	}

	// 在协程中执行任务
	go s.executeTask(conv, execLog)

	return nil
}

// executeTask 在协程中执行任务
func (s *aiTaskExecutorService) executeTask(conv *database.TaskConversation, execLog *database.TaskExecutionLog) {
	var finalStatus database.ConversationStatus
	var execStatus database.TaskExecutionStatus
	var errorMsg string
	var commitHash string

	defer func() {
		// 更新对话状态
		conv.Status = finalStatus
		if err := s.taskConvRepo.Update(conv); err != nil {
			log.Printf("更新对话最终状态失败: %v", err)
		}

		// 更新执行日志状态
		execLog.Status = execStatus
		execLog.ErrorMessage = errorMsg
		execLog.CommitHash = commitHash
		if err := s.execLogRepo.Update(execLog); err != nil {
			log.Printf("更新执行日志最终状态失败: %v", err)
		}
	}()

	// 1. 创建临时工作目录
	workspacePath, err := s.workspaceManager.CreateTempWorkspace(conv.ID)
	if err != nil {
		finalStatus = database.ConversationStatusFailed
		execStatus = database.TaskExecStatusFailed
		errorMsg = fmt.Sprintf("创建工作目录失败: %v", err)
		return
	}
	defer s.workspaceManager.CleanupWorkspace(workspacePath)

	execLog.WorkspacePath = workspacePath
	execLog.Status = database.TaskExecStatusRunning
	s.execLogRepo.Update(execLog)

	// 2. 克隆代码
	credential, err := s.prepareGitCredential(conv.Task.Project)
	if err != nil {
		finalStatus = database.ConversationStatusFailed
		execStatus = database.TaskExecStatusFailed
		errorMsg = fmt.Sprintf("准备Git凭据失败: %v", err)
		return
	}

	if err := s.workspaceManager.CloneRepository(
		workspacePath,
		conv.Task.Project.RepoURL,
		conv.Task.StartBranch,
		credential,
	); err != nil {
		finalStatus = database.ConversationStatusFailed
		execStatus = database.TaskExecStatusFailed
		errorMsg = fmt.Sprintf("克隆仓库失败: %v", err)
		return
	}

	s.appendLog(execLog.ID, fmt.Sprintf("✅ 成功克隆仓库到: %s\n", workspacePath))

	// 3. 构建并执行Docker命令
	dockerCmd := s.buildDockerCommand(conv, workspacePath)
	execLog.DockerCommand = dockerCmd
	s.execLogRepo.Update(execLog)

	s.appendLog(execLog.ID, fmt.Sprintf("🚀 开始执行命令: %s\n", dockerCmd))

	if err := s.executeDockerCommand(dockerCmd, execLog.ID); err != nil {
		finalStatus = database.ConversationStatusFailed
		execStatus = database.TaskExecStatusFailed
		errorMsg = fmt.Sprintf("执行Docker命令失败: %v", err)
		return
	}

	// 4. 提交更改
	hash, err := s.workspaceManager.CommitChanges(workspacePath, fmt.Sprintf("AI generated changes for conversation %d", conv.ID))
	if err != nil {
		s.appendLog(execLog.ID, fmt.Sprintf("⚠️ 提交更改失败: %v\n", err))
		// 不设为失败，因为任务可能已经成功执行
	} else {
		commitHash = hash
		s.appendLog(execLog.ID, fmt.Sprintf("✅ 成功提交更改，commit hash: %s\n", hash))
	}

	finalStatus = database.ConversationStatusSuccess
	execStatus = database.TaskExecStatusSuccess
}

// prepareGitCredential 准备Git凭据
func (s *aiTaskExecutorService) prepareGitCredential(project *database.Project) (*utils.GitCredentialInfo, error) {
	if project.Credential == nil {
		return nil, nil
	}

	credential := &utils.GitCredentialInfo{
		Type:     utils.GitCredentialType(project.Credential.Type),
		Username: project.Credential.Username,
	}

	// 解密敏感信息
	if project.Credential.Type == database.GitCredentialTypePassword || project.Credential.Type == database.GitCredentialTypeToken {
		password, err := s.gitCredService.DecryptCredentialSecret(project.Credential, "password")
		if err != nil {
			return nil, err
		}
		credential.Password = password
	} else if project.Credential.Type == database.GitCredentialTypeSSHKey {
		privateKey, err := s.gitCredService.DecryptCredentialSecret(project.Credential, "private_key")
		if err != nil {
			return nil, err
		}
		credential.PrivateKey = privateKey
		credential.PublicKey = project.Credential.PublicKey
	}

	return credential, nil
}

// buildDockerCommand 构建Docker命令
func (s *aiTaskExecutorService) buildDockerCommand(conv *database.TaskConversation, workspacePath string) string {
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
	var imageName string
	var aiCommand []string

	switch devEnv.Type {
	case "claude-code":
		imageName = "claude-code:latest"
		aiCommand = []string{conv.Content}
	case "opencode":
		imageName = "opencode:latest"
		aiCommand = []string{conv.Content}
	case "gemini-cli":
		imageName = "gemini-cli:latest"
		aiCommand = []string{conv.Content}
	default:
		// 默认使用 claude-code
		imageName = "claude-code:latest"
		aiCommand = []string{conv.Content}
	}

	// 添加镜像名称
	cmd = append(cmd, imageName)

	// 添加 AI 命令参数
	cmd = append(cmd, aiCommand...)

	return strings.Join(cmd, " ")
}

// executeDockerCommand 执行Docker命令
func (s *aiTaskExecutorService) executeDockerCommand(dockerCmd string, execLogID uint) error {
	// 首先检查 Docker 是否可用
	if err := s.checkDockerAvailability(); err != nil {
		s.appendLog(execLogID, fmt.Sprintf("❌ Docker 不可用: %v\n", err))
		return fmt.Errorf("Docker 不可用: %v", err)
	}

	s.appendLog(execLogID, "✅ Docker 可用性检查通过\n")

	// 解析超时时间
	timeout, err := time.ParseDuration(s.config.DockerExecutionTimeout)
	if err != nil {
		log.Printf("解析Docker超时时间失败，使用默认值30分钟: %v", err)
		timeout = 30 * time.Minute
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
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

	// 实时读取输出
	go s.readPipe(stdout, execLogID, "STDOUT")
	go s.readPipe(stderr, execLogID, "STDERR")

	// 等待命令完成
	return cmd.Wait()
}

// checkDockerAvailability 检查 Docker 是否可用
func (s *aiTaskExecutorService) checkDockerAvailability() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 检查 Docker 守护进程是否可用
	cmd := exec.CommandContext(ctx, "docker", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Docker 命令不可用或 Docker 守护进程未运行: %v", err)
	}

	return nil
}

// readPipe 读取管道输出
func (s *aiTaskExecutorService) readPipe(pipe interface{}, execLogID uint, prefix string) {
	scanner := bufio.NewScanner(pipe.(interface{ Read([]byte) (int, error) }))
	for scanner.Scan() {
		line := scanner.Text()
		logLine := fmt.Sprintf("[%s] %s: %s\n", time.Now().Format("15:04:05"), prefix, line)
		s.appendLog(execLogID, logLine)
	}
}

// appendLog 追加日志
func (s *aiTaskExecutorService) appendLog(execLogID uint, content string) {
	if err := s.execLogRepo.AppendLog(execLogID, content); err != nil {
		log.Printf("追加日志失败: %v", err)
	}
}
