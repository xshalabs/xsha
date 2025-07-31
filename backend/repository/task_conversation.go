package repository

import (
	"xsha-backend/database"

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

func (r *taskConversationRepository) GetByID(id uint, createdBy string) (*database.TaskConversation, error) {
	var conversation database.TaskConversation
	err := r.db.Preload("Task").
		Preload("Task.Project").
		Preload("Task.Project.Credential").
		Preload("Task.DevEnvironment").
		Where("id = ? AND created_by = ?", id, createdBy).First(&conversation).Error
	if err != nil {
		return nil, err
	}
	return &conversation, nil
}

func (r *taskConversationRepository) List(taskID uint, createdBy string, page, pageSize int) ([]database.TaskConversation, int64, error) {
	var conversations []database.TaskConversation
	var total int64

	query := r.db.Model(&database.TaskConversation{}).
		Where("task_id = ? AND created_by = ?", taskID, createdBy)

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

func (r *taskConversationRepository) Delete(id uint, createdBy string) error {
	return r.db.Where("id = ? AND created_by = ?", id, createdBy).Delete(&database.TaskConversation{}).Error
}

func (r *taskConversationRepository) ListByTask(taskID uint, createdBy string) ([]database.TaskConversation, error) {
	var conversations []database.TaskConversation
	err := r.db.Where("task_id = ? AND created_by = ?", taskID, createdBy).
		Order("created_at ASC").Find(&conversations).Error
	return conversations, err
}

func (r *taskConversationRepository) GetLatestByTask(taskID uint, createdBy string) (*database.TaskConversation, error) {
	var conversation database.TaskConversation
	err := r.db.Where("task_id = ? AND created_by = ?", taskID, createdBy).
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
	err := r.db.Preload("Task").
		Preload("Task.Project").
		Preload("Task.Project.Credential").
		Preload("Task.DevEnvironment").
		Where("status = ?", database.ConversationStatusPending).
		Order("created_at ASC").
		Find(&conversations).Error
	return conversations, err
}

func (r *taskConversationRepository) HasPendingOrRunningConversations(taskID uint, createdBy string) (bool, error) {
	var count int64
	err := r.db.Model(&database.TaskConversation{}).
		Where("task_id = ? AND created_by = ? AND status IN (?)",
			taskID, createdBy, []database.ConversationStatus{
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
