package services

import (
	"strings"
	"xsha-backend/database"
	appErrors "xsha-backend/errors"
	"xsha-backend/repository"
	"xsha-backend/utils"
)

type taskConversationService struct {
	repo        repository.TaskConversationRepository
	taskRepo    repository.TaskRepository
	execLogRepo repository.TaskExecutionLogRepository
	resultRepo  repository.TaskConversationResultRepository
	taskService TaskService
}

func NewTaskConversationService(repo repository.TaskConversationRepository, taskRepo repository.TaskRepository, execLogRepo repository.TaskExecutionLogRepository, resultRepo repository.TaskConversationResultRepository, taskService TaskService) TaskConversationService {
	return &taskConversationService{
		repo:        repo,
		taskRepo:    taskRepo,
		execLogRepo: execLogRepo,
		resultRepo:  resultRepo,
		taskService: taskService,
	}
}

func (s *taskConversationService) CreateConversation(taskID uint, content, createdBy string) (*database.TaskConversation, error) {
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

	conversation := &database.TaskConversation{
		TaskID:    taskID,
		Content:   strings.TrimSpace(content),
		Status:    database.ConversationStatusPending,
		CreatedBy: createdBy,
	}

	if err := s.repo.Create(conversation); err != nil {
		return nil, err
	}

	conversation.Task = task
	return conversation, nil
}

func (s *taskConversationService) GetConversation(id uint) (*database.TaskConversation, error) {
	return s.repo.GetByID(id)
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

func (s *taskConversationService) UpdateConversation(id uint, updates map[string]interface{}) error {
	conversation, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if content, ok := updates["content"]; ok {
		if contentStr, ok := content.(string); ok && strings.TrimSpace(contentStr) == "" {
			return appErrors.ErrRequired
		}
	}

	for key, value := range updates {
		switch key {
		case "content":
			if v, ok := value.(string); ok {
				conversation.Content = strings.TrimSpace(v)
			}
		}
	}

	return s.repo.Update(conversation)
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
		if err := utils.GitResetToPreviousCommit(conversation.Task.WorkspacePath, conversation.CommitHash); err != nil {
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

func (s *taskConversationService) GetLatestConversation(taskID uint) (*database.TaskConversation, error) {
	return s.repo.GetLatestByTask(taskID)
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

	diff, err := utils.GetCommitDiff(task.WorkspacePath, conversation.CommitHash, includeContent)
	if err != nil {
		return nil, err
	}

	return diff, nil
}

func (s *taskConversationService) GetConversationGitDiffFile(conversationID uint, filePath string) (string, error) {
	if filePath == "" {
		return "", appErrors.ErrFilePathEmpty
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

	diffContent, err := utils.GetCommitFileDiff(task.WorkspacePath, conversation.CommitHash, filePath)
	if err != nil {
		return "", err
	}

	return diffContent, nil
}
