package executor

import (
	"context"
	"fmt"
	"sync"
	"time"
	"xsha-backend/config"
	"xsha-backend/database"
	"xsha-backend/repository"
	"xsha-backend/services"
	"xsha-backend/utils"
)

type aiTaskExecutorService struct {
	taskConvRepo       repository.TaskConversationRepository
	taskRepo           repository.TaskRepository
	execLogRepo        repository.TaskExecutionLogRepository
	taskConvResultRepo repository.TaskConversationResultRepository

	gitCredService        services.GitCredentialService
	taskConvResultService services.TaskConversationResultService
	taskService           services.TaskService

	executionManager *ExecutionManager
	dockerExecutor   DockerExecutor
	resultParser     ResultParser
	workspaceCleaner WorkspaceCleaner
	stateManager     ConversationStateManager

	workspaceManager *utils.WorkspaceManager
	config           *config.Config
}

func NewAITaskExecutorService(
	taskConvRepo repository.TaskConversationRepository,
	taskRepo repository.TaskRepository,
	execLogRepo repository.TaskExecutionLogRepository,
	taskConvResultRepo repository.TaskConversationResultRepository,
	gitCredService services.GitCredentialService,
	taskConvResultService services.TaskConversationResultService,
	taskService services.TaskService,
	systemConfigService services.SystemConfigService,
	cfg *config.Config,
) services.AITaskExecutorService {
	gitCloneTimeout, err := time.ParseDuration(cfg.GitCloneTimeout)
	if err != nil {
		utils.Warn("Failed to parse git clone timeout, using default 5 minutes", "timeout", cfg.GitCloneTimeout, "error", err)
		gitCloneTimeout = 5 * time.Minute
	}
	workspaceManager := utils.NewWorkspaceManager(cfg.WorkspaceBaseDir, gitCloneTimeout)

	logAppender := &logAppenderImpl{
		execLogRepo: execLogRepo,
	}

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
		taskService:           taskService,
		executionManager:      executionManager,
		dockerExecutor:        dockerExecutor,
		resultParser:          resultParser,
		workspaceCleaner:      workspaceCleaner,
		stateManager:          stateManager,
		workspaceManager:      workspaceManager,
		config:                cfg,
	}
}

func (s *aiTaskExecutorService) ProcessPendingConversations() error {
	conversations, err := s.taskConvRepo.GetPendingConversationsWithDetails()
	if err != nil {
		return fmt.Errorf("failed to get pending conversations: %v", err)
	}

	utils.Info("Found pending conversations to process",
		"count", len(conversations),
		"running", s.executionManager.GetRunningCount(),
		"maxConcurrency", s.executionManager.maxConcurrency)

	var wg sync.WaitGroup
	processedCount := 0
	skippedCount := 0

	for _, conv := range conversations {
		if !s.executionManager.CanExecute() {
			skippedCount++
			utils.Warn("Reached maximum concurrency limit, skipping conversation", "conversationId", conv.ID)
			continue
		}

		if s.executionManager.IsRunning(conv.ID) {
			skippedCount++
			utils.Warn("Conversation already in progress, skipping", "conversationId", conv.ID)
			continue
		}

		wg.Add(1)
		processedCount++

		go func(conversation database.TaskConversation) {
			defer wg.Done()
			if err := s.processConversation(&conversation); err != nil {
				utils.Error("Failed to process conversation", "conversationId", conversation.ID, "error", err)
			}
		}(conv)
	}

	wg.Wait()

	utils.Info("Batch conversation processing completed", "processed", processedCount, "skipped", skippedCount)
	return nil
}

func (s *aiTaskExecutorService) GetExecutionLog(conversationID uint) (*database.TaskExecutionLog, error) {
	return s.execLogRepo.GetByConversationID(conversationID)
}

func (s *aiTaskExecutorService) CancelExecution(conversationID uint, createdBy string) error {
	conv, err := s.taskConvRepo.GetByID(conversationID, createdBy)
	if err != nil {
		return fmt.Errorf("failed to get conversation info: %v", err)
	}

	if conv.Status != database.ConversationStatusPending && conv.Status != database.ConversationStatusRunning {
		return fmt.Errorf("can only cancel pending or running conversations")
	}

	if s.executionManager.CancelExecution(conversationID) {
		utils.Info("Force cancelling running conversation",
			"conversation_id", conversationID,
		)
	}

	conv.Status = database.ConversationStatusCancelled
	if err := s.taskConvRepo.Update(conv); err != nil {
		return fmt.Errorf("failed to update conversation status to cancelled: %v", err)
	}

	if conv.Task != nil && conv.Task.WorkspacePath != "" {
		if cleanupErr := s.workspaceCleaner.CleanupOnCancel(conv.Task.ID, conv.Task.WorkspacePath); cleanupErr != nil {
			utils.Error("Failed to cleanup workspace during cancellation", "task_id", conv.Task.ID, "workspace", conv.Task.WorkspacePath, "error", cleanupErr)
		}
	}

	return nil
}

func (s *aiTaskExecutorService) RetryExecution(conversationID uint, createdBy string) error {
	conv, err := s.taskConvRepo.GetByID(conversationID, createdBy)
	if err != nil {
		return fmt.Errorf("failed to get conversation info: %v", err)
	}

	if conv.Status != database.ConversationStatusFailed && conv.Status != database.ConversationStatusCancelled {
		return fmt.Errorf("can only retry failed or cancelled conversations")
	}

	if s.executionManager.IsRunning(conversationID) {
		return fmt.Errorf("conversation is running, cannot retry")
	}

	if !s.executionManager.CanExecute() {
		return fmt.Errorf("reached maximum concurrency limit, please try again later")
	}

	if err := s.execLogRepo.DeleteByConversationID(conversationID); err != nil {
		return fmt.Errorf("failed to delete old execution logs: %v", err)
	}

	conv.Status = database.ConversationStatusPending
	if err := s.taskConvRepo.Update(conv); err != nil {
		return fmt.Errorf("failed to reset conversation status: %v", err)
	}

	if err := s.processConversation(conv); err != nil {
		conv.Status = database.ConversationStatusFailed
		s.taskConvRepo.Update(conv)
		return fmt.Errorf("failed to retry execution: %v", err)
	}

	return nil
}

func (s *aiTaskExecutorService) GetExecutionStatus() map[string]interface{} {
	return map[string]interface{}{
		"running_count":   s.executionManager.GetRunningCount(),
		"max_concurrency": s.executionManager.maxConcurrency,
		"can_execute":     s.executionManager.CanExecute(),
	}
}

func (s *aiTaskExecutorService) processConversation(conv *database.TaskConversation) error {
	if conv.Task == nil {
		s.stateManager.SetFailed(conv, "task information is missing")
		return fmt.Errorf("task information is missing")
	}
	if conv.Task.Project == nil {
		s.stateManager.SetFailed(conv, "project information is missing")
		return fmt.Errorf("project information is missing")
	}
	if conv.Task.DevEnvironment == nil {
		s.stateManager.SetFailed(conv, "task has no development environment configured, cannot execute")
		return fmt.Errorf("task has no development environment configured, cannot execute")
	}

	if conv.Task.Status == database.TaskStatusTodo {
		if err := s.taskService.UpdateTaskStatus(conv.Task.ID, conv.CreatedBy, database.TaskStatusInProgress); err != nil {
			utils.Error("Failed to update task status", "task_id", conv.Task.ID, "error", err)
		} else {
			utils.Info("Task status updated", "task_id", conv.Task.ID, "old_status", "todo", "new_status", "in_progress")
		}
	}

	conv.Status = database.ConversationStatusRunning
	if err := s.taskConvRepo.Update(conv); err != nil {
		s.stateManager.Rollback(conv, fmt.Sprintf("failed to update conversation status: %v", err))
		return fmt.Errorf("failed to update conversation status: %v", err)
	}

	execLog := &database.TaskExecutionLog{
		ConversationID: conv.ID,
		ExecutionLogs:  "",
	}
	if err := s.execLogRepo.Create(execLog); err != nil {
		s.stateManager.Rollback(conv, fmt.Sprintf("failed to create execution log: %v", err))
		return fmt.Errorf("failed to create execution log: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	if !s.executionManager.AddExecution(conv.ID, cancel) {
		s.stateManager.RollbackToState(conv, execLog,
			database.ConversationStatusPending,
			"reached maximum concurrency limit")
		return fmt.Errorf("reached maximum concurrency limit")
	}

	go s.executeTask(ctx, conv, execLog)

	return nil
}

func (s *aiTaskExecutorService) executeTask(ctx context.Context, conv *database.TaskConversation, execLog *database.TaskExecutionLog) {
	var finalStatus database.ConversationStatus
	var errorMsg string
	var commitHash string

	defer func() {
		s.executionManager.RemoveExecution(conv.ID)

		conv.Status = finalStatus
		if err := s.taskConvRepo.Update(conv); err != nil {
			utils.Error("Failed to update conversation final status", "error", err)
		}

		if finalStatus == database.ConversationStatusFailed || finalStatus == database.ConversationStatusCancelled {
			if conv.Task != nil && conv.Task.WorkspacePath != "" {
				if finalStatus == database.ConversationStatusFailed {
					if cleanupErr := s.workspaceCleaner.CleanupOnFailure(conv.Task.ID, conv.Task.WorkspacePath); cleanupErr != nil {
						utils.Error("Error during failed task workspace cleanup", "task_id", conv.Task.ID, "error", cleanupErr)
					}
				} else if finalStatus == database.ConversationStatusCancelled {
					if cleanupErr := s.workspaceCleaner.CleanupOnCancel(conv.Task.ID, conv.Task.WorkspacePath); cleanupErr != nil {
						utils.Error("Error during cancelled task workspace cleanup", "task_id", conv.Task.ID, "error", cleanupErr)
					}
				}
			}
		}

		if commitHash != "" {
			if err := s.taskConvRepo.UpdateCommitHash(conv.ID, commitHash); err != nil {
				utils.Error("Failed to update conversation commit hash", "error", err)
			}
		}

		updates := make(map[string]interface{})

		if errorMsg != "" {
			updates["error_message"] = errorMsg
		}

		now := time.Now()
		updates["completed_at"] = &now

		if err := s.execLogRepo.UpdateMetadata(execLog.ID, updates); err != nil {
			utils.Error("Failed to update execution log metadata", "error", err)
		}

		statusMessage := fmt.Sprintf("Execution completed: %s", string(finalStatus))
		if errorMsg != "" {
			statusMessage += fmt.Sprintf(" - %s", errorMsg)
		}

		latestExecLog, err := s.execLogRepo.GetByID(execLog.ID)
		if err != nil {
			utils.Error("Failed to get latest execution log", "execLogID", execLog.ID, "error", err)
			latestExecLog = execLog // use original object as fallback
		}
		s.resultParser.ParseAndCreate(conv, latestExecLog)

		utils.Info("Conversation execution completed", "conversationId", conv.ID, "status", string(finalStatus))
	}()

	select {
	case <-ctx.Done():
		finalStatus = database.ConversationStatusCancelled
		errorMsg = "conversation cancelled"
		return
	default:
	}

	workspacePath, err := s.workspaceManager.GetOrCreateTaskWorkspace(conv.Task.ID, conv.Task.WorkspacePath)
	if err != nil {
		finalStatus = database.ConversationStatusFailed
		errorMsg = fmt.Sprintf("failed to create workspace: %v", err)
		return
	}

	if conv.Task.WorkspacePath == "" {
		conv.Task.WorkspacePath = workspacePath
		if updateErr := s.taskRepo.Update(conv.Task); updateErr != nil {
			utils.Error("Failed to update task workspace path", "error", updateErr)
		}
	}

	now := time.Now()
	startedUpdates := map[string]interface{}{
		"started_at": &now,
	}
	s.execLogRepo.UpdateMetadata(execLog.ID, startedUpdates)

	select {
	case <-ctx.Done():
		finalStatus = database.ConversationStatusCancelled
		errorMsg = "conversation cancelled"
		return
	default:
	}

	if s.workspaceManager.CheckGitRepositoryExists(workspacePath) {
	} else {
		credential, err := s.prepareGitCredential(conv.Task.Project)
		if err != nil {
			finalStatus = database.ConversationStatusFailed
			errorMsg = fmt.Sprintf("failed to prepare git credential: %v", err)
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
			errorMsg = fmt.Sprintf("failed to clone repository: %v", err)
			return
		}
	}

	select {
	case <-ctx.Done():
		finalStatus = database.ConversationStatusCancelled
		errorMsg = "conversation cancelled"
		return
	default:
	}

	workBranch := conv.Task.WorkBranch
	if workBranch == "" {
		workBranch = utils.GenerateWorkBranchName(conv.Task.Title, conv.Task.CreatedBy)
		conv.Task.WorkBranch = workBranch
		if updateErr := s.taskRepo.Update(conv.Task); updateErr != nil {
			utils.Error("Failed to update task work branch", "taskID", conv.Task.ID, "error", updateErr)
		} else {
			utils.Info("Generated work branch for existing task", "taskID", conv.Task.ID, "workBranch", workBranch)
		}
	}

	if err := s.workspaceManager.CreateAndSwitchToBranch(
		workspacePath,
		workBranch,
		conv.Task.StartBranch,
	); err != nil {
		finalStatus = database.ConversationStatusFailed
		errorMsg = fmt.Sprintf("failed to create or switch to work branch: %v", err)
		return
	}

	dockerCmd := s.dockerExecutor.BuildCommand(conv, workspacePath)
	dockerCmdForLog := s.dockerExecutor.BuildCommandForLog(conv, workspacePath)
	dockerUpdates := map[string]interface{}{
		"docker_command": dockerCmdForLog,
	}
	s.execLogRepo.UpdateMetadata(execLog.ID, dockerUpdates)

	if err := s.dockerExecutor.ExecuteWithContext(ctx, dockerCmd, execLog.ID); err != nil {
		select {
		case <-ctx.Done():
			finalStatus = database.ConversationStatusCancelled
			errorMsg = "conversation cancelled"
		default:
			finalStatus = database.ConversationStatusFailed
			errorMsg = fmt.Sprintf("failed to execute docker command: %v", err)
		}
		return
	}

	hash, err := s.workspaceManager.CommitChanges(workspacePath, fmt.Sprintf("AI generated changes for conversation %d", conv.ID))
	if err != nil {
	} else {
		commitHash = hash
	}

	finalStatus = database.ConversationStatusSuccess
}

func (s *aiTaskExecutorService) prepareGitCredential(project *database.Project) (*utils.GitCredentialInfo, error) {
	if project.Credential == nil {
		return nil, nil
	}

	credential := &utils.GitCredentialInfo{
		Type:     utils.GitCredentialType(project.Credential.Type),
		Username: project.Credential.Username,
	}

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

func (s *aiTaskExecutorService) CleanupWorkspaceOnFailure(taskID uint, workspacePath string) error {
	return s.workspaceCleaner.CleanupOnFailure(taskID, workspacePath)
}

func (s *aiTaskExecutorService) CleanupWorkspaceOnCancel(taskID uint, workspacePath string) error {
	return s.workspaceCleaner.CleanupOnCancel(taskID, workspacePath)
}

type logAppenderImpl struct {
	execLogRepo repository.TaskExecutionLogRepository
}

func (l *logAppenderImpl) AppendLog(execLogID uint, content string) {
	if err := l.execLogRepo.AppendLog(execLogID, content); err != nil {
		utils.Error("Failed to append log", "error", err)
		return
	}

}
