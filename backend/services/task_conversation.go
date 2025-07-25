package services

import (
	"errors"
	"sleep0-backend/database"
	"sleep0-backend/repository"
	"strings"
)

type taskConversationService struct {
	repo     repository.TaskConversationRepository
	taskRepo repository.TaskRepository
}

// NewTaskConversationService 创建任务对话服务实例
func NewTaskConversationService(repo repository.TaskConversationRepository, taskRepo repository.TaskRepository) TaskConversationService {
	return &taskConversationService{
		repo:     repo,
		taskRepo: taskRepo,
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
	_, err := s.repo.GetByID(id, createdBy)
	if err != nil {
		return err
	}

	return s.repo.Delete(id, createdBy)
}

// ListConversationsByTask 根据任务获取对话列表
func (s *taskConversationService) ListConversationsByTask(taskID uint, createdBy string) ([]database.TaskConversation, error) {
	return s.repo.ListByTask(taskID, createdBy)
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
