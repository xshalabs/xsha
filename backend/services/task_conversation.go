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

// NewTaskConversationService 创建任务对话服务实例
func NewTaskConversationService(repo repository.TaskConversationRepository, taskRepo repository.TaskRepository, execLogRepo repository.TaskExecutionLogRepository) TaskConversationService {
	return &taskConversationService{
		repo:        repo,
		taskRepo:    taskRepo,
		execLogRepo: execLogRepo,
	}
}

// CreateConversation 创建对话
func (s *taskConversationService) CreateConversation(taskID uint, content, createdBy string) (*database.TaskConversation, error) {
	// 验证输入数据
	if err := s.ValidateConversationData(taskID, content, createdBy); err != nil {
		return nil, err
	}

	// 检查任务是否存在且属于当前用户
	task, err := s.taskRepo.GetByID(taskID, createdBy)
	if err != nil {
		return nil, errors.New("task not found or access denied")
	}

	// 检查任务状态 - 已结束的任务不能创建对话
	if task.Status == database.TaskStatusDone || task.Status == database.TaskStatusCancelled {
		return nil, errors.New("cannot create conversation for completed or cancelled task")
	}

	// 检查任务是否有pending或running状态的对话
	hasPendingOrRunning, err := s.repo.HasPendingOrRunningConversations(taskID, createdBy)
	if err != nil {
		return nil, errors.New("failed to check conversation status")
	}
	if hasPendingOrRunning {
		return nil, errors.New("cannot create new conversation while there are pending or running conversations")
	}

	// 创建对话
	conversation := &database.TaskConversation{
		TaskID:    taskID,
		Content:   strings.TrimSpace(content),
		Status:    database.ConversationStatusPending,
		CreatedBy: createdBy,
	}

	if err := s.repo.Create(conversation); err != nil {
		return nil, err
	}

	// 预加载关联数据
	conversation.Task = task
	return conversation, nil
}

// GetConversation 获取对话
func (s *taskConversationService) GetConversation(id uint, createdBy string) (*database.TaskConversation, error) {
	return s.repo.GetByID(id, createdBy)
}

// ListConversations 获取对话列表
func (s *taskConversationService) ListConversations(taskID uint, createdBy string, page, pageSize int) ([]database.TaskConversation, int64, error) {
	// 验证分页参数
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return s.repo.List(taskID, createdBy, page, pageSize)
}

// UpdateConversation 更新对话
func (s *taskConversationService) UpdateConversation(id uint, createdBy string, updates map[string]interface{}) error {
	// 获取对话
	conversation, err := s.repo.GetByID(id, createdBy)
	if err != nil {
		return err
	}

	// 验证更新数据
	if content, ok := updates["content"]; ok {
		if contentStr, ok := content.(string); ok && strings.TrimSpace(contentStr) == "" {
			return errors.New("conversation content cannot be empty")
		}
	}

	// 更新字段
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

// DeleteConversation 删除对话
func (s *taskConversationService) DeleteConversation(id uint, createdBy string) error {
	// 检查对话是否存在
	conversation, err := s.repo.GetByID(id, createdBy)
	if err != nil {
		return err
	}

	// 验证对话状态：只有在非 running 状态时才能删除
	if conversation.Status == database.ConversationStatusRunning {
		return errors.New("cannot delete conversation while it is running")
	}

	// 获取该任务的最新对话
	latestConversation, err := s.repo.GetLatestByTask(conversation.TaskID, createdBy)
	if err != nil {
		return errors.New("failed to get latest conversation")
	}

	// 验证只有最新对话才能删除
	if conversation.ID != latestConversation.ID {
		return errors.New("only the latest conversation can be deleted")
	}

	// 如果对话有 commit_hash，则需要将关联任务的工作空间仓库重置到前一个提交
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

	// 先删除关联的任务执行日志
	if err := s.execLogRepo.DeleteByConversationID(id); err != nil {
		return errors.New("failed to delete related execution logs")
	}

	// 再删除对话
	return s.repo.Delete(id, createdBy)
}

// GetLatestConversation 获取最新对话
func (s *taskConversationService) GetLatestConversation(taskID uint, createdBy string) (*database.TaskConversation, error) {
	return s.repo.GetLatestByTask(taskID, createdBy)
}

// ValidateConversationData 验证对话数据
func (s *taskConversationService) ValidateConversationData(taskID uint, content string, createdBy string) error {
	// 验证内容
	if strings.TrimSpace(content) == "" {
		return errors.New("conversation content is required")
	}

	if len(content) > 10000 {
		return errors.New("conversation content too long")
	}

	// 验证任务ID
	if taskID == 0 {
		return errors.New("task ID is required")
	}

	return nil
}

// GetConversationGitDiff 获取对话的Git变动差异
func (s *taskConversationService) GetConversationGitDiff(conversationID uint, createdBy string, includeContent bool) (*utils.GitDiffSummary, error) {
	// 获取对话信息
	conversation, err := s.repo.GetByID(conversationID, createdBy)
	if err != nil {
		return nil, errors.New("conversation not found or access denied")
	}

	// 检查是否有 commit hash
	if conversation.CommitHash == "" {
		return nil, errors.New("conversation has no commit hash")
	}

	// 获取任务信息
	task, err := s.taskRepo.GetByID(conversation.TaskID, createdBy)
	if err != nil {
		return nil, errors.New("task not found or access denied")
	}

	// 检查工作空间路径
	if task.WorkspacePath == "" {
		return nil, errors.New("task workspace path is empty")
	}

	// 获取提交的Git差异
	diff, err := utils.GetCommitDiff(task.WorkspacePath, conversation.CommitHash, includeContent)
	if err != nil {
		return nil, err
	}

	return diff, nil
}

// GetConversationGitDiffFile 获取对话指定文件的Git变动详情
func (s *taskConversationService) GetConversationGitDiffFile(conversationID uint, createdBy string, filePath string) (string, error) {
	if filePath == "" {
		return "", errors.New("file path cannot be empty")
	}

	// 获取对话信息
	conversation, err := s.repo.GetByID(conversationID, createdBy)
	if err != nil {
		return "", errors.New("conversation not found or access denied")
	}

	// 检查是否有 commit hash
	if conversation.CommitHash == "" {
		return "", errors.New("conversation has no commit hash")
	}

	// 获取任务信息
	task, err := s.taskRepo.GetByID(conversation.TaskID, createdBy)
	if err != nil {
		return "", errors.New("task not found or access denied")
	}

	// 检查工作空间路径
	if task.WorkspacePath == "" {
		return "", errors.New("task workspace path is empty")
	}

	// 获取文件的Git差异内容
	diffContent, err := utils.GetCommitFileDiff(task.WorkspacePath, conversation.CommitHash, filePath)
	if err != nil {
		return "", err
	}

	return diffContent, nil
}
