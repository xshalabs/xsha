package services

import (
	"fmt"
	"strings"
	"time"
	"xsha-backend/database"
	appErrors "xsha-backend/errors"
	"xsha-backend/repository"
	"xsha-backend/utils"
)

type taskConversationService struct {
	repo              repository.TaskConversationRepository
	taskRepo          repository.TaskRepository
	execLogRepo       repository.TaskExecutionLogRepository
	resultRepo        repository.TaskConversationResultRepository
	taskService       TaskService
	attachmentService TaskConversationAttachmentService
	workspaceManager  *utils.WorkspaceManager
}

func NewTaskConversationService(repo repository.TaskConversationRepository, taskRepo repository.TaskRepository, execLogRepo repository.TaskExecutionLogRepository, resultRepo repository.TaskConversationResultRepository, taskService TaskService, attachmentService TaskConversationAttachmentService, workspaceManager *utils.WorkspaceManager) TaskConversationService {
	return &taskConversationService{
		repo:              repo,
		taskRepo:          taskRepo,
		execLogRepo:       execLogRepo,
		resultRepo:        resultRepo,
		taskService:       taskService,
		attachmentService: attachmentService,
		workspaceManager:  workspaceManager,
	}
}

func (s *taskConversationService) CreateConversationWithExecutionTime(taskID uint, content, createdBy string, executionTime *time.Time, envParams string, adminID *uint) (*database.TaskConversation, error) {
	if err := s.ValidateConversationData(taskID, content); err != nil {
		return nil, err
	}

	task, err := s.taskRepo.GetByID(taskID)
	if err != nil {
		return nil, appErrors.ErrTaskNotFound
	}

	if task.Status == database.TaskStatusDone || task.Status == database.TaskStatusCancelled {
		return nil, appErrors.ErrConversationTaskCompleted
	}

	hasPendingOrRunning, err := s.repo.HasPendingOrRunningConversations(taskID)
	if err != nil {
		return nil, appErrors.ErrConversationGetFailed
	}
	if hasPendingOrRunning {
		return nil, appErrors.ErrConversationCreateFailed
	}

	// Ensure envParams is valid JSON, default to empty object if not provided
	if envParams == "" {
		envParams = "{}"
	}

	conversation := &database.TaskConversation{
		TaskID:        taskID,
		Content:       strings.TrimSpace(content),
		Status:        database.ConversationStatusPending,
		ExecutionTime: executionTime,
		EnvParams:     envParams,
		AdminID:       adminID,
		CreatedBy:     createdBy,
	}

	if err := s.repo.Create(conversation); err != nil {
		return nil, err
	}

	conversation.Task = task
	return conversation, nil
}

func (s *taskConversationService) CreateConversationWithExecutionTimeAndAttachments(taskID uint, content, createdBy string, executionTime *time.Time, envParams string, attachmentIDs []uint, adminID *uint) (*database.TaskConversation, error) {
	if err := s.ValidateConversationData(taskID, content); err != nil {
		return nil, err
	}

	task, err := s.taskRepo.GetByID(taskID)
	if err != nil {
		return nil, appErrors.ErrTaskNotFound
	}

	if task.Status == database.TaskStatusDone || task.Status == database.TaskStatusCancelled {
		return nil, appErrors.ErrConversationTaskCompleted
	}

	hasPendingOrRunning, err := s.repo.HasPendingOrRunningConversations(taskID)
	if err != nil {
		return nil, appErrors.ErrConversationGetFailed
	}
	if hasPendingOrRunning {
		return nil, appErrors.ErrConversationCreateFailed
	}

	// Ensure envParams is valid JSON, default to empty object if not provided
	if envParams == "" {
		envParams = "{}"
	}

	// Validate and process attachments
	if len(attachmentIDs) > 0 {
		// First create the conversation, then associate attachments
		// For now, we'll handle attachments after conversation creation
	}

	// Process content with attachment tags if attachments exist
	processedContent := strings.TrimSpace(content)
	if len(attachmentIDs) > 0 {
		// We'll update the content after creating attachments
		// For now, keep original content
	}

	conversation := &database.TaskConversation{
		TaskID:        taskID,
		Content:       processedContent,
		Status:        database.ConversationStatusPending,
		ExecutionTime: executionTime,
		EnvParams:     envParams,
		AdminID:       adminID,
		CreatedBy:     createdBy,
	}

	if err := s.repo.Create(conversation); err != nil {
		return nil, err
	}

	// Process attachment associations if provided
	if len(attachmentIDs) > 0 {
		for _, attachmentID := range attachmentIDs {
			// Associate attachment with conversation
			if err := s.attachmentService.AssociateWithConversation(attachmentID, conversation.ID); err != nil {
				utils.Error("Failed to associate attachment with conversation", "attachmentID", attachmentID, "conversationID", conversation.ID, "error", err)
				continue
			}
		}

		// Get all attachments for this conversation after association
		attachments, err := s.attachmentService.GetAttachmentsByConversation(conversation.ID)
		if err != nil {
			utils.Error("Failed to get attachments for conversation", "conversationID", conversation.ID, "error", err)
		} else {
			// Update conversation content with attachment tags
			processedContent = s.attachmentService.ProcessContentWithAttachments(content, attachments, conversation.ID)
			conversation.Content = processedContent

			// Save the updated content
			if err := s.repo.Update(conversation); err != nil {
				utils.Error("Failed to update conversation content with attachment tags", "conversationID", conversation.ID, "error", err)
			}
		}
	}

	conversation.Task = task
	return conversation, nil
}

func (s *taskConversationService) GetConversation(id uint) (*database.TaskConversation, error) {
	return s.repo.GetByID(id)
}

func (s *taskConversationService) GetConversationWithResult(id uint) (map[string]interface{}, error) {
	conversation, result, executionLog, err := s.repo.GetWithResult(id)
	if err != nil {
		return nil, err
	}

	response := map[string]interface{}{
		"conversation":  conversation,
		"result":        result,
		"execution_log": executionLog,
	}

	return response, nil
}

func (s *taskConversationService) ListConversations(taskID uint, page, pageSize int) ([]database.TaskConversation, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return s.repo.List(taskID, page, pageSize)
}

func (s *taskConversationService) DeleteConversation(id uint) error {
	conversation, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if conversation.Status == database.ConversationStatusRunning {
		return appErrors.ErrConversationDeleteFailed
	}

	latestConversation, err := s.repo.GetLatestByTask(conversation.TaskID)
	if err != nil {
		return appErrors.ErrConversationGetFailed
	}

	if conversation.ID != latestConversation.ID {
		return appErrors.ErrConversationDeleteLatestOnly
	}

	if conversation.CommitHash != "" && conversation.Task != nil && conversation.Task.WorkspacePath != "" {
		// Convert relative workspace path to absolute for git operations
		absoluteWorkspacePath := s.workspaceManager.GetAbsolutePath(conversation.Task.WorkspacePath)
		if err := utils.GitResetToPreviousCommit(absoluteWorkspacePath, conversation.CommitHash); err != nil {
			utils.Error("Failed to reset git repository to previous commit",
				"conversation_id", id,
				"commit_hash", conversation.CommitHash,
				"workspace_path", conversation.Task.WorkspacePath,
				"error", err)
			return appErrors.NewI18nError("git.reset_failed", err.Error())
		}
		utils.Info("Successfully reset git repository to previous commit",
			"conversation_id", id,
			"commit_hash", conversation.CommitHash,
			"workspace_path", conversation.Task.WorkspacePath)
	}

	if err := s.execLogRepo.DeleteByConversationID(id); err != nil {
		return appErrors.ErrConversationDeleteFailed
	}

	if err := s.resultRepo.DeleteByConversationID(id); err != nil {
		utils.Warn("Failed to delete conversation result",
			"conversation_id", id,
			"error", err)
	}

	// Delete attachments associated with the conversation
	if err := s.attachmentService.DeleteAttachmentsByConversation(id); err != nil {
		utils.Warn("Failed to delete conversation attachments",
			"conversation_id", id,
			"error", err)
	}

	if err := s.repo.Delete(id); err != nil {
		return err
	}

	latestResult, err := s.resultRepo.GetLatestByTaskID(conversation.TaskID)
	if err != nil {
		if err.Error() != "record not found" {
			utils.Warn("Failed to get latest conversation result for task",
				"task_id", conversation.TaskID,
				"error", err)
		}
		if err := s.taskService.UpdateTaskSessionID(conversation.TaskID, ""); err != nil {
			utils.Warn("Failed to clear task session ID",
				"task_id", conversation.TaskID,
				"error", err)
		}
	} else {
		if err := s.taskService.UpdateTaskSessionID(conversation.TaskID, latestResult.SessionID); err != nil {
			utils.Warn("Failed to update task session ID",
				"task_id", conversation.TaskID,
				"session_id", latestResult.SessionID,
				"error", err)
		} else {
			utils.Info("Successfully updated task session ID after conversation deletion",
				"task_id", conversation.TaskID,
				"old_conversation_id", id,
				"new_session_id", latestResult.SessionID)
		}
	}

	return nil
}

func (s *taskConversationService) ValidateConversationData(taskID uint, content string) error {
	if strings.TrimSpace(content) == "" {
		return appErrors.ErrRequired
	}

	if len(content) > 100000 {
		return appErrors.ErrTooLong
	}

	if taskID == 0 {
		return appErrors.ErrRequired
	}

	return nil
}

func (s *taskConversationService) GetConversationGitDiff(conversationID uint, includeContent bool) (*utils.GitDiffSummary, error) {
	conversation, err := s.repo.GetByID(conversationID)
	if err != nil {
		return nil, appErrors.ErrTaskNotFound
	}

	if conversation.CommitHash == "" {
		return nil, appErrors.ErrNoCommitHash
	}

	task, err := s.taskRepo.GetByID(conversation.TaskID)
	if err != nil {
		return nil, appErrors.ErrTaskNotFound
	}

	if task.WorkspacePath == "" {
		return nil, appErrors.ErrWorkspacePathEmpty
	}

	// Convert relative workspace path to absolute for git operations
	absoluteWorkspacePath := s.workspaceManager.GetAbsolutePath(task.WorkspacePath)

	diff, err := utils.GetCommitDiff(absoluteWorkspacePath, conversation.CommitHash, includeContent)
	if err != nil {
		return nil, err
	}

	return diff, nil
}

func (s *taskConversationService) GetConversationGitDiffFile(conversationID uint, filePath string) (string, error) {
	if filePath == "" {
		return "", appErrors.ErrFilePathEmpty
	}

	// Additional security validation at service layer
	if err := utils.ValidateGitFilePath(filePath); err != nil {
		utils.Warn("Service layer security validation failed",
			"conversationID", conversationID,
			"filePath", filePath,
			"error", err)
		return "", fmt.Errorf("invalid file path: %v", err)
	}

	conversation, err := s.repo.GetByID(conversationID)
	if err != nil {
		return "", appErrors.ErrTaskNotFound
	}

	if conversation.CommitHash == "" {
		return "", appErrors.ErrNoCommitHash
	}

	task, err := s.taskRepo.GetByID(conversation.TaskID)
	if err != nil {
		return "", appErrors.ErrTaskNotFound
	}

	if task.WorkspacePath == "" {
		return "", appErrors.ErrWorkspacePathEmpty
	}

	// Log the file access attempt for security monitoring
	utils.Info("Git diff file access",
		"conversationID", conversationID,
		"taskID", task.ID,
		"filePath", filePath,
		"commitHash", conversation.CommitHash,
		"workspacePath", task.WorkspacePath)

	// Convert relative workspace path to absolute for git operations
	absoluteWorkspacePath := s.workspaceManager.GetAbsolutePath(task.WorkspacePath)

	diffContent, err := utils.GetCommitFileDiff(absoluteWorkspacePath, conversation.CommitHash, filePath)
	if err != nil {
		utils.Error("Failed to get commit file diff",
			"conversationID", conversationID,
			"filePath", filePath,
			"commitHash", conversation.CommitHash,
			"workspacePath", absoluteWorkspacePath,
			"error", err)
		return "", err
	}

	return diffContent, nil
}
