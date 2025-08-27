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
	"xsha-backend/services/executor/result_parser"
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
	systemConfigService   services.SystemConfigService
	attachmentService     services.TaskConversationAttachmentService

	executionManager *ExecutionManager
	dockerExecutor   DockerExecutor
	resultParser     result_parser.Parser
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
	attachmentService services.TaskConversationAttachmentService,
	cfg *config.Config,
) services.AITaskExecutorService {
	return NewAITaskExecutorServiceWithManager(
		taskConvRepo, taskRepo, execLogRepo, taskConvResultRepo,
		gitCredService, taskConvResultService, taskService, systemConfigService,
		attachmentService, cfg, nil,
	)
}

func NewAITaskExecutorServiceWithManager(
	taskConvRepo repository.TaskConversationRepository,
	taskRepo repository.TaskRepository,
	execLogRepo repository.TaskExecutionLogRepository,
	taskConvResultRepo repository.TaskConversationResultRepository,
	gitCredService services.GitCredentialService,
	taskConvResultService services.TaskConversationResultService,
	taskService services.TaskService,
	systemConfigService services.SystemConfigService,
	attachmentService services.TaskConversationAttachmentService,
	cfg *config.Config,
	executionManager *ExecutionManager,
) services.AITaskExecutorService {
	gitCloneTimeout, err := systemConfigService.GetGitCloneTimeout()
	if err != nil {
		utils.Error("Failed to get git clone timeout from system config, using default", "error", err)
		gitCloneTimeout = 5 * time.Minute
	}
	workspaceManager := utils.NewWorkspaceManager(cfg.WorkspaceBaseDir, gitCloneTimeout)

	logAppender := &logAppenderImpl{
		execLogRepo: execLogRepo,
	}

	// Create ExecutionManager if not provided
	if executionManager == nil {
		maxConcurrency := 5
		if cfg.MaxConcurrentTasks > 0 {
			maxConcurrency = cfg.MaxConcurrentTasks
		}
		executionManager = NewExecutionManager(maxConcurrency)
	}
	dockerExecutor := NewDockerExecutor(cfg, logAppender, systemConfigService)
	resultParser := result_parser.NewResultParser(taskConvResultRepo, taskConvResultService, taskService)
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
		systemConfigService:   systemConfigService,
		attachmentService:     attachmentService,
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

	return nil
}

func (s *aiTaskExecutorService) GetExecutionLog(conversationID uint) (*database.TaskExecutionLog, error) {
	return s.execLogRepo.GetByConversationID(conversationID)
}

func (s *aiTaskExecutorService) CancelExecution(conversationID uint, createdBy string) error {
	conv, err := s.taskConvRepo.GetByID(conversationID)
	if err != nil {
		return fmt.Errorf("failed to get conversation info: %v", err)
	}

	if conv.Status != database.ConversationStatusPending && conv.Status != database.ConversationStatusRunning {
		return fmt.Errorf("can only cancel pending or running conversations")
	}

	// Get cancel function and container ID
	cancelFunc, containerID := s.executionManager.CancelExecution(conversationID)
	if cancelFunc != nil {
		utils.Info("Force cancelling running conversation",
			"conversation_id", conversationID,
			"container_id", containerID,
		)

		// Cancel the context
		cancelFunc()

		// If we have a container ID, try to stop and remove it
		if containerID != "" {
			utils.Info("Attempting to stop and remove container", "container_id", containerID)
			if cleanupErr := s.dockerExecutor.StopAndRemoveContainer(containerID); cleanupErr != nil {
				utils.Error("Failed to stop and remove container during cancellation",
					"container_id", containerID,
					"conversation_id", conversationID,
					"error", cleanupErr)
			} else {
				utils.Info("Successfully stopped and removed container", "container_id", containerID)
			}
		}
	}

	conv.Status = database.ConversationStatusCancelled
	if err := s.taskConvRepo.Update(conv); err != nil {
		return fmt.Errorf("failed to update conversation status to cancelled: %v", err)
	}

	// Delete associated execution result if it exists
	if err := s.taskConvResultRepo.DeleteByConversationID(conversationID); err != nil {
		utils.Warn("Failed to delete conversation result during cancellation",
			"conversation_id", conversationID,
			"error", err)
	} else {
		utils.Info("Successfully deleted conversation result during cancellation",
			"conversation_id", conversationID)
	}

	if conv.Task != nil && conv.Task.WorkspacePath != "" {
		if cleanupErr := s.workspaceCleaner.CleanupOnCancel(conv.Task.ID, conv.Task.WorkspacePath); cleanupErr != nil {
			utils.Error("Failed to cleanup workspace during cancellation", "task_id", conv.Task.ID, "workspace", conv.Task.WorkspacePath, "error", cleanupErr)
		}
	}

	return nil
}

func (s *aiTaskExecutorService) RetryExecution(conversationID uint, createdBy string) error {
	conv, err := s.taskConvRepo.GetByID(conversationID)
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

	// Delete associated execution result if it exists
	if err := s.taskConvResultRepo.DeleteByConversationID(conversationID); err != nil {
		return fmt.Errorf("failed to delete old conversation result: %v", err)
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
		if err := s.taskService.UpdateTaskStatus(conv.Task.ID, database.TaskStatusInProgress); err != nil {
			utils.Error("Failed to update task status", "task_id", conv.Task.ID, "error", err)
			s.stateManager.SetFailed(conv, fmt.Sprintf("failed to update task status: %v", err))
			return fmt.Errorf("failed to update task status: %v", err)
		} else {
			utils.Info("Task status updated", "task_id", conv.Task.ID, "old_status", "todo", "new_status", "in_progress")
			// Update the in-memory task object to reflect the database change
			conv.Task.Status = database.TaskStatusInProgress
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

		now := utils.Now()
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

	now := utils.Now()
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

	proxyConfig, err := s.systemConfigService.GetGitProxyConfig()
	if err != nil {
		utils.Warn("Failed to get proxy config, using no proxy", "error", err)
		proxyConfig = nil
	}

	if s.workspaceManager.CheckGitRepositoryExists(workspacePath) {
		if err := s.workspaceCleaner.CleanupBeforeExecution(conv.Task.ID, workspacePath); err != nil {
			finalStatus = database.ConversationStatusFailed
			errorMsg = fmt.Sprintf("failed to cleanup workspace before execution: %v", err)
			return
		}
	} else {
		credential, err := s.prepareGitCredential(conv.Task.Project)
		if err != nil {
			finalStatus = database.ConversationStatusFailed
			errorMsg = fmt.Sprintf("failed to prepare git credential: %v", err)
			return
		}

		gitSSLVerify, err := s.systemConfigService.GetGitSSLVerify()
		if err != nil {
			utils.Warn("Failed to get git SSL verify setting, using default false", "error", err)
			gitSSLVerify = false
		}

		if err := s.workspaceManager.CloneRepositoryWithConfig(
			workspacePath,
			conv.Task.Project.RepoURL,
			conv.Task.StartBranch,
			credential,
			gitSSLVerify,
			proxyConfig,
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
		proxyConfig,
	); err != nil {
		finalStatus = database.ConversationStatusFailed
		errorMsg = fmt.Sprintf("failed to create or switch to work branch: %v", err)
		return
	}

	// Process attachments before building Docker command
	workspaceAttachments, err := s.attachmentService.CopyAttachmentsToWorkspace(conv.ID, workspacePath)
	if err != nil {
		finalStatus = database.ConversationStatusFailed
		errorMsg = fmt.Sprintf("failed to copy attachments to workspace: %v", err)
		return
	}

	// Replace attachment tags in conversation content with workspace paths
	processedContent := s.attachmentService.ReplaceAttachmentTagsWithPaths(conv.Content, workspaceAttachments, workspacePath)

	// Create a temporary conversation with processed content for Docker execution
	tempConv := *conv
	tempConv.Content = processedContent

	dockerCmdForLog := s.dockerExecutor.BuildCommandForLog(&tempConv, workspacePath)
	dockerUpdates := map[string]interface{}{
		"docker_command": dockerCmdForLog,
	}
	s.execLogRepo.UpdateMetadata(execLog.ID, dockerUpdates)

	// Execute with container tracking using processed conversation
	containerID, err := s.dockerExecutor.ExecuteWithContainerTracking(ctx, &tempConv, workspacePath, execLog.ID)
	if containerID != "" {
		// Set the container ID in execution manager for proper cleanup on cancellation
		s.executionManager.SetContainerID(conv.ID, containerID)
	}

	if err != nil {
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

	// Clean up workspace attachments before committing changes
	if cleanupErr := s.attachmentService.CleanupWorkspaceAttachments(workspacePath); cleanupErr != nil {
		utils.Warn("Failed to cleanup workspace attachments before commit", "workspace", workspacePath, "error", cleanupErr)
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
