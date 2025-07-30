package services

import (
	"errors"
	"strings"
	"xsha-backend/database"
	"xsha-backend/repository"
	"xsha-backend/utils"
)

type taskConversationService struct {
	repo        repository.TaskConversationRepository
	taskRepo    repository.TaskRepository
	execLogRepo repository.TaskExecutionLogRepository
}

func NewTaskConversationService(repo repository.TaskConversationRepository, taskRepo repository.TaskRepository, execLogRepo repository.TaskExecutionLogRepository) TaskConversationService {
	return &taskConversationService{
		repo:        repo,
		taskRepo:    taskRepo,
		execLogRepo: execLogRepo,
	}
}

func (s *taskConversationService) CreateConversation(taskID uint, content, createdBy string) (*database.TaskConversation, error) {
	if err := s.ValidateConversationData(taskID, content, createdBy); err != nil {
		return nil, err
	}

	task, err := s.taskRepo.GetByID(taskID, createdBy)
	if err != nil {
		return nil, errors.New("task not found or access denied")
	}

	if task.Status == database.TaskStatusDone || task.Status == database.TaskStatusCancelled {
		return nil, errors.New("cannot create conversation for completed or cancelled task")
	}

	hasPendingOrRunning, err := s.repo.HasPendingOrRunningConversations(taskID, createdBy)
	if err != nil {
		return nil, errors.New("failed to check conversation status")
	}
	if hasPendingOrRunning {
		return nil, errors.New("cannot create new conversation while there are pending or running conversations")
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

func (s *taskConversationService) GetConversation(id uint, createdBy string) (*database.TaskConversation, error) {
	return s.repo.GetByID(id, createdBy)
}

func (s *taskConversationService) ListConversations(taskID uint, createdBy string, page, pageSize int) ([]database.TaskConversation, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return s.repo.List(taskID, createdBy, page, pageSize)
}

func (s *taskConversationService) UpdateConversation(id uint, createdBy string, updates map[string]interface{}) error {
	conversation, err := s.repo.GetByID(id, createdBy)
	if err != nil {
		return err
	}

	if content, ok := updates["content"]; ok {
		if contentStr, ok := content.(string); ok && strings.TrimSpace(contentStr) == "" {
			return errors.New("conversation content cannot be empty")
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

func (s *taskConversationService) DeleteConversation(id uint, createdBy string) error {
	conversation, err := s.repo.GetByID(id, createdBy)
	if err != nil {
		return err
	}

	if conversation.Status == database.ConversationStatusRunning {
		return errors.New("cannot delete conversation while it is running")
	}

	latestConversation, err := s.repo.GetLatestByTask(conversation.TaskID, createdBy)
	if err != nil {
		return errors.New("failed to get latest conversation")
	}

	if conversation.ID != latestConversation.ID {
		return errors.New("only the latest conversation can be deleted")
	}

	if conversation.CommitHash != "" && conversation.Task != nil && conversation.Task.WorkspacePath != "" {
		if err := utils.GitResetToPreviousCommit(conversation.Task.WorkspacePath, conversation.CommitHash); err != nil {
			utils.Error("Failed to reset git repository to previous commit",
				"conversation_id", id,
				"commit_hash", conversation.CommitHash,
				"workspace_path", conversation.Task.WorkspacePath,
				"error", err)
			return errors.New("failed to reset git repository to previous commit: " + err.Error())
		}
		utils.Info("Successfully reset git repository to previous commit",
			"conversation_id", id,
			"commit_hash", conversation.CommitHash,
			"workspace_path", conversation.Task.WorkspacePath)
	}

	if err := s.execLogRepo.DeleteByConversationID(id); err != nil {
		return errors.New("failed to delete related execution logs")
	}

	return s.repo.Delete(id, createdBy)
}

func (s *taskConversationService) GetLatestConversation(taskID uint, createdBy string) (*database.TaskConversation, error) {
	return s.repo.GetLatestByTask(taskID, createdBy)
}

func (s *taskConversationService) ValidateConversationData(taskID uint, content string, createdBy string) error {
	if strings.TrimSpace(content) == "" {
		return errors.New("conversation content is required")
	}

	if len(content) > 10000 {
		return errors.New("conversation content too long")
	}

	if taskID == 0 {
		return errors.New("task ID is required")
	}

	return nil
}

func (s *taskConversationService) GetConversationGitDiff(conversationID uint, createdBy string, includeContent bool) (*utils.GitDiffSummary, error) {
	conversation, err := s.repo.GetByID(conversationID, createdBy)
	if err != nil {
		return nil, errors.New("conversation not found or access denied")
	}

	if conversation.CommitHash == "" {
		return nil, errors.New("conversation has no commit hash")
	}

	task, err := s.taskRepo.GetByID(conversation.TaskID, createdBy)
	if err != nil {
		return nil, errors.New("task not found or access denied")
	}

	if task.WorkspacePath == "" {
		return nil, errors.New("task workspace path is empty")
	}

	diff, err := utils.GetCommitDiff(task.WorkspacePath, conversation.CommitHash, includeContent)
	if err != nil {
		return nil, err
	}

	return diff, nil
}

func (s *taskConversationService) GetConversationGitDiffFile(conversationID uint, createdBy string, filePath string) (string, error) {
	if filePath == "" {
		return "", errors.New("file path cannot be empty")
	}

	conversation, err := s.repo.GetByID(conversationID, createdBy)
	if err != nil {
		return "", errors.New("conversation not found or access denied")
	}

	if conversation.CommitHash == "" {
		return "", errors.New("conversation has no commit hash")
	}

	task, err := s.taskRepo.GetByID(conversation.TaskID, createdBy)
	if err != nil {
		return "", errors.New("task not found or access denied")
	}

	if task.WorkspacePath == "" {
		return "", errors.New("task workspace path is empty")
	}

	diffContent, err := utils.GetCommitFileDiff(task.WorkspacePath, conversation.CommitHash, filePath)
	if err != nil {
		return "", err
	}

	return diffContent, nil
}
