package executor

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
	"xsha-backend/config"
	"xsha-backend/database"
	"xsha-backend/repository"
	"xsha-backend/services"
	"xsha-backend/utils"
)

type aiTaskExecutorService struct {
	// 仓储层
	taskConvRepo       repository.TaskConversationRepository
	taskRepo           repository.TaskRepository
	execLogRepo        repository.TaskExecutionLogRepository
	taskConvResultRepo repository.TaskConversationResultRepository

	// 外部服务
	gitCredService        services.GitCredentialService
	taskConvResultService services.TaskConversationResultService

	// 内部组件
	executionManager *ExecutionManager
	dockerExecutor   DockerExecutor
	resultParser     ResultParser
	workspaceCleaner WorkspaceCleaner
	stateManager     ConversationStateManager

	// 基础设施
	workspaceManager *utils.WorkspaceManager
	logBroadcaster   *services.LogBroadcaster
	config           *config.Config
}

// NewAITaskExecutorService 创建AI任务执行服务
func NewAITaskExecutorService(
	taskConvRepo repository.TaskConversationRepository,
	taskRepo repository.TaskRepository,
	execLogRepo repository.TaskExecutionLogRepository,
	taskConvResultRepo repository.TaskConversationResultRepository,
	gitCredService services.GitCredentialService,
	taskConvResultService services.TaskConversationResultService,
	systemConfigService services.SystemConfigService,
	cfg *config.Config,
	logBroadcaster *services.LogBroadcaster,
) services.AITaskExecutorService {
	// 创建工作空间管理器
	workspaceManager := utils.NewWorkspaceManager(cfg.WorkspaceBaseDir)

	// 创建日志追加器
	logAppender := &logAppenderImpl{
		execLogRepo:    execLogRepo,
		logBroadcaster: logBroadcaster,
	}

	// 创建内部组件
	maxConcurrency := 5
	if cfg.MaxConcurrentTasks > 0 {
		maxConcurrency = cfg.MaxConcurrentTasks
	}

	executionManager := NewExecutionManager(maxConcurrency)
	dockerExecutor := NewDockerExecutor(cfg, logAppender, systemConfigService)
	resultParser := NewResultParser(taskConvResultRepo, taskConvResultService)
	workspaceCleaner := NewWorkspaceCleaner(workspaceManager)
	stateManager := NewConversationStateManager(taskConvRepo, execLogRepo)

	return &aiTaskExecutorService{
		taskConvRepo:          taskConvRepo,
		taskRepo:              taskRepo,
		execLogRepo:           execLogRepo,
		taskConvResultRepo:    taskConvResultRepo,
		gitCredService:        gitCredService,
		taskConvResultService: taskConvResultService,
		executionManager:      executionManager,
		dockerExecutor:        dockerExecutor,
		resultParser:          resultParser,
		workspaceCleaner:      workspaceCleaner,
		stateManager:          stateManager,
		workspaceManager:      workspaceManager,
		logBroadcaster:        logBroadcaster,
		config:                cfg,
	}
}

// ProcessPendingConversations 处理待处理的对话 - 支持并发执行
func (s *aiTaskExecutorService) ProcessPendingConversations() error {
	conversations, err := s.taskConvRepo.GetPendingConversationsWithDetails()
	if err != nil {
		return fmt.Errorf("获取待处理对话失败: %v", err)
	}

	utils.Info("发现待处理的对话",
		"count", len(conversations),
		"running", s.executionManager.GetRunningCount(),
		"maxConcurrency", s.executionManager.maxConcurrency)

	// 并发处理对话
	var wg sync.WaitGroup
	processedCount := 0
	skippedCount := 0

	for _, conv := range conversations {
		// 检查是否可以执行新任务
		if !s.executionManager.CanExecute() {
			skippedCount++
			utils.Warn("达到最大并发数限制，跳过对话", "conversationId", conv.ID)
			continue
		}

		// 检查是否已经在运行
		if s.executionManager.IsRunning(conv.ID) {
			skippedCount++
			utils.Warn("对话已在运行中，跳过", "conversationId", conv.ID)
			continue
		}

		wg.Add(1)
		processedCount++

		// 并发处理对话
		go func(conversation database.TaskConversation) {
			defer wg.Done()
			if err := s.processConversation(&conversation); err != nil {
				utils.Error("处理对话失败", "conversationId", conversation.ID, "error", err)
			}
		}(conv)
	}

	// 等待所有当前批次的对话开始处理（不等待完成）
	wg.Wait()

	utils.Info("本批次对话处理完成", "processed", processedCount, "skipped", skippedCount)
	return nil
}

// GetExecutionLog 获取执行日志
func (s *aiTaskExecutorService) GetExecutionLog(conversationID uint) (*database.TaskExecutionLog, error) {
	return s.execLogRepo.GetByConversationID(conversationID)
}

// CancelExecution 取消执行 - 支持强制取消正在运行的任务
func (s *aiTaskExecutorService) CancelExecution(conversationID uint, createdBy string) error {
	// 获取对话信息作为主体
	conv, err := s.taskConvRepo.GetByID(conversationID, createdBy)
	if err != nil {
		return fmt.Errorf("获取对话信息失败: %v", err)
	}

	// 检查对话状态是否可以取消
	if conv.Status != database.ConversationStatusPending && conv.Status != database.ConversationStatusRunning {
		return fmt.Errorf("只能取消待处理或执行中的任务")
	}

	// 如果任务正在运行，先取消执行
	if s.executionManager.CancelExecution(conversationID) {
		utils.Info("Force cancelling running conversation",
			"conversation_id", conversationID,
		)
	}

	// 更新对话状态为已取消
	conv.Status = database.ConversationStatusCancelled
	if err := s.taskConvRepo.Update(conv); err != nil {
		return fmt.Errorf("failed to update conversation status to cancelled: %v", err)
	}

	// 清理工作空间（在取消时）
	if conv.Task != nil && conv.Task.WorkspacePath != "" {
		if cleanupErr := s.workspaceCleaner.CleanupOnCancel(conv.Task.ID, conv.Task.WorkspacePath); cleanupErr != nil {
			utils.Error("取消执行时清理工作空间失败", "task_id", conv.Task.ID, "workspace", conv.Task.WorkspacePath, "error", cleanupErr)
			// 不因为清理失败而中断取消操作，但要记录错误
		}
	}

	return nil
}

// RetryExecution 重试执行对话
func (s *aiTaskExecutorService) RetryExecution(conversationID uint, createdBy string) error {
	// 获取对话信息
	conv, err := s.taskConvRepo.GetByID(conversationID, createdBy)
	if err != nil {
		return fmt.Errorf("获取对话信息失败: %v", err)
	}

	// 检查对话状态是否可以重试
	if conv.Status != database.ConversationStatusFailed && conv.Status != database.ConversationStatusCancelled {
		return fmt.Errorf("只能重试失败或已取消的任务")
	}

	// 检查是否有正在运行的执行
	if s.executionManager.IsRunning(conversationID) {
		return fmt.Errorf("任务正在执行中，无法重试")
	}

	// 检查是否可以执行新任务（并发限制）
	if !s.executionManager.CanExecute() {
		return fmt.Errorf("已达到最大并发数限制，请稍后重试")
	}

	// 删除该对话的所有旧执行日志
	if err := s.execLogRepo.DeleteByConversationID(conversationID); err != nil {
		return fmt.Errorf("删除旧执行日志失败: %v", err)
	}

	// 重置对话状态为待处理
	conv.Status = database.ConversationStatusPending
	if err := s.taskConvRepo.Update(conv); err != nil {
		return fmt.Errorf("重置对话状态失败: %v", err)
	}

	// 处理对话（这会创建新的执行日志）
	if err := s.processConversation(conv); err != nil {
		// 如果处理失败，将状态回滚为失败
		conv.Status = database.ConversationStatusFailed
		s.taskConvRepo.Update(conv)
		return fmt.Errorf("重试执行失败: %v", err)
	}

	return nil
}

// GetExecutionStatus 获取执行状态信息
func (s *aiTaskExecutorService) GetExecutionStatus() map[string]interface{} {
	return map[string]interface{}{
		"running_count":   s.executionManager.GetRunningCount(),
		"max_concurrency": s.executionManager.maxConcurrency,
		"can_execute":     s.executionManager.CanExecute(),
	}
}

// processConversation 处理单个对话 - 添加上下文控制
func (s *aiTaskExecutorService) processConversation(conv *database.TaskConversation) error {
	// 验证关联数据
	if conv.Task == nil {
		s.stateManager.SetFailed(conv, "任务信息缺失")
		return fmt.Errorf("任务信息缺失")
	}
	if conv.Task.Project == nil {
		s.stateManager.SetFailed(conv, "项目信息缺失")
		return fmt.Errorf("项目信息缺失")
	}
	if conv.Task.DevEnvironment == nil {
		s.stateManager.SetFailed(conv, "task has no development environment configured, cannot execute")
		return fmt.Errorf("task has no development environment configured, cannot execute")
	}

	// 更新对话状态为 running
	conv.Status = database.ConversationStatusRunning
	if err := s.taskConvRepo.Update(conv); err != nil {
		s.stateManager.Rollback(conv, fmt.Sprintf("failed to update conversation status: %v", err))
		return fmt.Errorf("failed to update conversation status: %v", err)
	}

	// 创建执行日志
	execLog := &database.TaskExecutionLog{
		ConversationID: conv.ID,
		ExecutionLogs:  "", // 初始化为空字符串，避免NULL值问题
	}
	if err := s.execLogRepo.Create(execLog); err != nil {
		s.stateManager.Rollback(conv, fmt.Sprintf("failed to create execution log: %v", err))
		return fmt.Errorf("failed to create execution log: %v", err)
	}

	// 创建上下文和取消函数
	ctx, cancel := context.WithCancel(context.Background())

	// 注册到执行管理器
	if !s.executionManager.AddExecution(conv.ID, cancel) {
		// 如果无法添加到执行管理器，回滚状态
		s.stateManager.RollbackToState(conv, execLog,
			database.ConversationStatusPending,
			"超过最大并发数限制")
		return fmt.Errorf("超过最大并发数限制")
	}

	// 在协程中执行任务
	go s.executeTask(ctx, conv, execLog)

	return nil
}

// executeTask 在协程中执行任务 - 添加上下文控制
func (s *aiTaskExecutorService) executeTask(ctx context.Context, conv *database.TaskConversation, execLog *database.TaskExecutionLog) {
	var finalStatus database.ConversationStatus
	var errorMsg string
	var commitHash string

	// 确保从执行管理器中移除
	defer func() {
		s.executionManager.RemoveExecution(conv.ID)

		// 更新对话状态 (主状态)
		conv.Status = finalStatus
		if err := s.taskConvRepo.Update(conv); err != nil {
			utils.Error("更新对话最终状态失败", "error", err)
		}

		// 清理工作空间（在失败或取消时）
		if finalStatus == database.ConversationStatusFailed || finalStatus == database.ConversationStatusCancelled {
			if conv.Task != nil && conv.Task.WorkspacePath != "" {
				if finalStatus == database.ConversationStatusFailed {
					if cleanupErr := s.workspaceCleaner.CleanupOnFailure(conv.Task.ID, conv.Task.WorkspacePath); cleanupErr != nil {
						utils.Error("清理失败任务工作空间时出错", "task_id", conv.Task.ID, "error", cleanupErr)
					}
				} else if finalStatus == database.ConversationStatusCancelled {
					if cleanupErr := s.workspaceCleaner.CleanupOnCancel(conv.Task.ID, conv.Task.WorkspacePath); cleanupErr != nil {
						utils.Error("清理取消任务工作空间时出错", "task_id", conv.Task.ID, "error", cleanupErr)
					}
				}
			}
		}

		// 更新对话的 commit hash（如果成功）
		if commitHash != "" {
			if err := s.taskConvRepo.UpdateCommitHash(conv.ID, commitHash); err != nil {
				utils.Error("更新对话commit hash失败", "error", err)
			}
		}

		// 准备执行日志元数据更新
		updates := make(map[string]interface{})

		if errorMsg != "" {
			updates["error_message"] = errorMsg
		}

		// 更新完成时间
		now := time.Now()
		updates["completed_at"] = &now

		// 使用 UpdateMetadata 避免覆盖 execution_logs 字段
		if err := s.execLogRepo.UpdateMetadata(execLog.ID, updates); err != nil {
			utils.Error("更新执行日志元数据失败", "error", err)
		}

		// 广播状态变化
		statusMessage := fmt.Sprintf("执行完成: %s", string(finalStatus))
		if errorMsg != "" {
			statusMessage += fmt.Sprintf(" - %s", errorMsg)
		}
		s.logBroadcaster.BroadcastStatus(conv.ID, fmt.Sprintf("%s - %s", string(finalStatus), statusMessage))

		// 尝试解析并创建任务结果记录
		// 重新从数据库获取最新的执行日志数据（包含所有追加的日志内容）
		latestExecLog, err := s.execLogRepo.GetByID(execLog.ID)
		if err != nil {
			utils.Error("获取最新执行日志失败", "execLogID", execLog.ID, "error", err)
			latestExecLog = execLog // 使用原始对象作为后备
		}
		s.resultParser.ParseAndCreate(conv, latestExecLog)

		utils.Info("对话执行完成", "conversationId", conv.ID, "status", string(finalStatus))
	}()

	// 检查是否被取消
	select {
	case <-ctx.Done():
		finalStatus = database.ConversationStatusCancelled
		errorMsg = "任务被取消"
		return
	default:
	}

	// 1. 获取或创建任务级工作目录
	workspacePath, err := s.workspaceManager.GetOrCreateTaskWorkspace(conv.Task.ID, conv.Task.WorkspacePath)
	if err != nil {
		finalStatus = database.ConversationStatusFailed
		errorMsg = fmt.Sprintf("创建工作目录失败: %v", err)
		return
	}

	// 更新任务的工作空间路径（如果尚未设置）
	if conv.Task.WorkspacePath == "" {
		conv.Task.WorkspacePath = workspacePath
		if updateErr := s.taskRepo.Update(conv.Task); updateErr != nil {
			utils.Error("更新任务工作空间路径失败", "error", updateErr)
			// 继续执行，不因为路径更新失败而中断任务
		}
	}

	// 更新开始时间
	now := time.Now()
	startedUpdates := map[string]interface{}{
		"started_at": &now,
	}
	s.execLogRepo.UpdateMetadata(execLog.ID, startedUpdates)

	// 检查是否被取消
	select {
	case <-ctx.Done():
		finalStatus = database.ConversationStatusCancelled
		errorMsg = "任务被取消"
		return
	default:
	}

	// 2. 检查并克隆代码
	if s.workspaceManager.CheckGitRepositoryExists(workspacePath) {
		// 仓库已存在，跳过克隆
		// 这里不再直接调用appendLog，Docker执行器内部会处理日志
	} else {
		// 仓库不存在，执行克隆
		credential, err := s.prepareGitCredential(conv.Task.Project)
		if err != nil {
			finalStatus = database.ConversationStatusFailed
			errorMsg = fmt.Sprintf("准备Git凭据失败: %v", err)
			return
		}

		if err := s.workspaceManager.CloneRepositoryWithConfig(
			workspacePath,
			conv.Task.Project.RepoURL,
			conv.Task.StartBranch,
			credential,
			s.config.GitSSLVerify,
		); err != nil {
			finalStatus = database.ConversationStatusFailed
			errorMsg = fmt.Sprintf("克隆仓库失败: %v", err)
			return
		}
	}

	// 检查是否被取消
	select {
	case <-ctx.Done():
		finalStatus = database.ConversationStatusCancelled
		errorMsg = "任务被取消"
		return
	default:
	}

	// 3. 处理工作分支（为已存在的任务生成工作分支）
	workBranch := conv.Task.WorkBranch
	if workBranch == "" {
		// 为已存在的任务生成工作分支名称
		workBranch = s.generateWorkBranchName(conv.Task.Title, conv.Task.CreatedBy)

		// 更新任务的工作分支字段
		conv.Task.WorkBranch = workBranch
		if updateErr := s.taskRepo.Update(conv.Task); updateErr != nil {
			utils.Error("更新任务工作分支失败", "taskID", conv.Task.ID, "error", updateErr)
			// 继续执行，不因为更新失败而中断任务
		} else {
			utils.Info("为已存在任务生成工作分支", "taskID", conv.Task.ID, "workBranch", workBranch)
		}
	}

	// 4. 创建并切换到工作分支
	if err := s.workspaceManager.CreateAndSwitchToBranch(
		workspacePath,
		workBranch,
		conv.Task.StartBranch,
	); err != nil {
		finalStatus = database.ConversationStatusFailed
		errorMsg = fmt.Sprintf("创建或切换到工作分支失败: %v", err)
		return
	}

	// 5. 构建并执行Docker命令
	dockerCmd := s.dockerExecutor.BuildCommand(conv, workspacePath)
	// 构建用于记录的安全版本（环境变量值已打码）
	dockerCmdForLog := s.dockerExecutor.BuildCommandForLog(conv, workspacePath)
	dockerUpdates := map[string]interface{}{
		"docker_command": dockerCmdForLog,
	}
	s.execLogRepo.UpdateMetadata(execLog.ID, dockerUpdates)

	// 使用上下文控制的Docker执行
	if err := s.dockerExecutor.ExecuteWithContext(ctx, dockerCmd, execLog.ID); err != nil {
		// 检查是否是由于取消导致的错误
		select {
		case <-ctx.Done():
			finalStatus = database.ConversationStatusCancelled
			errorMsg = "任务被取消"
		default:
			finalStatus = database.ConversationStatusFailed
			errorMsg = fmt.Sprintf("执行Docker命令失败: %v", err)
		}
		return
	}

	// 5. 提交更改
	hash, err := s.workspaceManager.CommitChanges(workspacePath, fmt.Sprintf("AI generated changes for conversation %d", conv.ID))
	if err != nil {
		// 不设为失败，因为任务可能已经成功执行
	} else {
		commitHash = hash
	}

	finalStatus = database.ConversationStatusSuccess
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
	switch project.Credential.Type {
	case database.GitCredentialTypePassword, database.GitCredentialTypeToken:
		password, err := s.gitCredService.DecryptCredentialSecret(project.Credential, "password")
		if err != nil {
			return nil, err
		}
		credential.Password = password
	case database.GitCredentialTypeSSHKey:
		privateKey, err := s.gitCredService.DecryptCredentialSecret(project.Credential, "private_key")
		if err != nil {
			return nil, err
		}
		credential.PrivateKey = privateKey
		credential.PublicKey = project.Credential.PublicKey
	}

	return credential, nil
}

// generateWorkBranchName 生成工作分支名称（与 task service 中的逻辑保持一致）
func (s *aiTaskExecutorService) generateWorkBranchName(title, createdBy string) string {
	// 清理标题，只保留字母数字和连字符
	cleanTitle := strings.ToLower(strings.TrimSpace(title))

	// 替换空格和特殊字符为连字符
	cleanTitle = strings.ReplaceAll(cleanTitle, " ", "-")
	cleanTitle = strings.ReplaceAll(cleanTitle, "_", "-")

	// 移除非字母数字和连字符的字符
	var result strings.Builder
	for _, r := range cleanTitle {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	cleanTitle = result.String()

	// 限制长度
	if len(cleanTitle) > 30 {
		cleanTitle = cleanTitle[:30]
	}

	// 去掉开头和结尾的连字符
	cleanTitle = strings.Trim(cleanTitle, "-")

	// 如果清理后为空，使用默认前缀
	if cleanTitle == "" {
		cleanTitle = "task"
	}

	// 生成时间戳
	timestamp := time.Now().Format("20060102-150405")

	// 组合分支名: feature/{user}/{clean-title}-{timestamp}
	return fmt.Sprintf("xsha/%s/%s-%s", createdBy, cleanTitle, timestamp)
}

// CleanupWorkspaceOnFailure 在任务执行失败时清理工作空间
func (s *aiTaskExecutorService) CleanupWorkspaceOnFailure(taskID uint, workspacePath string) error {
	return s.workspaceCleaner.CleanupOnFailure(taskID, workspacePath)
}

// CleanupWorkspaceOnCancel 在任务被取消时清理工作空间
func (s *aiTaskExecutorService) CleanupWorkspaceOnCancel(taskID uint, workspacePath string) error {
	return s.workspaceCleaner.CleanupOnCancel(taskID, workspacePath)
}

// logAppenderImpl 日志追加器实现
type logAppenderImpl struct {
	execLogRepo    repository.TaskExecutionLogRepository
	logBroadcaster *services.LogBroadcaster
}

func (l *logAppenderImpl) AppendLog(execLogID uint, content string) {
	// 追加到数据库
	if err := l.execLogRepo.AppendLog(execLogID, content); err != nil {
		utils.Error("追加日志失败", "error", err)
		return
	}

	// 获取对话ID进行广播
	if execLog, err := l.execLogRepo.GetByID(execLogID); err == nil {
		l.logBroadcaster.BroadcastLog(execLog.ConversationID, content, "log")
	}
}
