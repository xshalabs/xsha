package repository

import (
	"xsha-backend/database"
	"xsha-backend/utils"

	"gorm.io/gorm"
)

type taskConversationRepository struct {
	db *gorm.DB
}

func NewTaskConversationRepository(db *gorm.DB) TaskConversationRepository {
	return &taskConversationRepository{db: db}
}

func (r *taskConversationRepository) Create(conversation *database.TaskConversation) error {
	return r.db.Create(conversation).Error
}

func (r *taskConversationRepository) GetByID(id uint) (*database.TaskConversation, error) {
	var conversation database.TaskConversation
	err := r.db.Preload("Task").
		Preload("Task.Project").
		Preload("Task.Project.Credential").
		Preload("Task.DevEnvironment").
		Where("id = ?", id).First(&conversation).Error
	if err != nil {
		return nil, err
	}
	return &conversation, nil
}

func (r *taskConversationRepository) GetWithResult(id uint) (*database.TaskConversation, *database.TaskConversationResult, *database.TaskExecutionLog, error) {
	var conversation database.TaskConversation
	err := r.db.Preload("Task").
		Preload("Task.Project").
		Preload("Task.Project.Credential").
		Preload("Task.DevEnvironment").
		Where("id = ?", id).First(&conversation).Error
	if err != nil {
		return nil, nil, nil, err
	}

	// Get conversation result if exists
	var result database.TaskConversationResult
	err = r.db.Where("conversation_id = ?", id).First(&result).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, nil, nil, err
	}

	// Get execution log if exists
	var executionLog database.TaskExecutionLog
	err = r.db.Where("conversation_id = ?", id).First(&executionLog).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, nil, nil, err
	}

	// Return conversation with result and execution log (both can be nil if not found)
	var resultPtr *database.TaskConversationResult
	if result.ID != 0 {
		resultPtr = &result
	}

	var logPtr *database.TaskExecutionLog
	if executionLog.ID != 0 {
		logPtr = &executionLog
	}

	return &conversation, resultPtr, logPtr, nil
}

func (r *taskConversationRepository) List(taskID uint, page, pageSize int) ([]database.TaskConversation, int64, error) {
	var conversations []database.TaskConversation
	var total int64

	query := r.db.Model(&database.TaskConversation{}).
		Where("task_id = ?", taskID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at ASC").Offset(offset).Limit(pageSize).Find(&conversations).Error; err != nil {
		return nil, 0, err
	}

	return conversations, total, nil
}

func (r *taskConversationRepository) Update(conversation *database.TaskConversation) error {
	return r.db.Save(conversation).Error
}

func (r *taskConversationRepository) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&database.TaskConversation{}).Error
}

func (r *taskConversationRepository) ListByTask(taskID uint) ([]database.TaskConversation, error) {
	var conversations []database.TaskConversation
	err := r.db.Where("task_id = ?", taskID).
		Order("created_at ASC").Find(&conversations).Error
	return conversations, err
}

func (r *taskConversationRepository) GetLatestByTask(taskID uint) (*database.TaskConversation, error) {
	var conversation database.TaskConversation
	err := r.db.Where("task_id = ?", taskID).
		Order("created_at DESC").First(&conversation).Error
	if err != nil {
		return nil, err
	}
	return &conversation, nil
}

func (r *taskConversationRepository) ListByStatus(status database.ConversationStatus) ([]database.TaskConversation, error) {
	var conversations []database.TaskConversation
	err := r.db.Where("status = ?", status).
		Order("created_at ASC").Find(&conversations).Error
	return conversations, err
}

func (r *taskConversationRepository) GetPendingConversationsWithDetails() ([]database.TaskConversation, error) {
	var conversations []database.TaskConversation
	now := utils.Now()

	// 先查询所有 pending 状态的对话数量，用于日志统计
	var totalPendingCount int64
	r.db.Model(&database.TaskConversation{}).
		Where("status = ?", database.ConversationStatusPending).
		Count(&totalPendingCount)

	// 查询条件：状态为 pending，且执行时间为空或执行时间已到
	err := r.db.Preload("Task").
		Preload("Task.Project").
		Preload("Task.Project.Credential").
		Preload("Task.DevEnvironment").
		Where("status = ? AND (execution_time IS NULL OR execution_time <= ?)",
			database.ConversationStatusPending, now).
		Order("created_at ASC").
		Find(&conversations).Error

	return conversations, err
}

func (r *taskConversationRepository) HasPendingOrRunningConversations(taskID uint) (bool, error) {
	var count int64
	err := r.db.Model(&database.TaskConversation{}).
		Where("task_id = ? AND status IN (?)",
			taskID, []database.ConversationStatus{
				database.ConversationStatusPending,
				database.ConversationStatusRunning,
			}).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *taskConversationRepository) UpdateCommitHash(id uint, commitHash string) error {
	return r.db.Model(&database.TaskConversation{}).
		Where("id = ?", id).
		Update("commit_hash", commitHash).Error
}
