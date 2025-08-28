package repository

import (
	"xsha-backend/database"

	"gorm.io/gorm"
)

type taskConversationResultRepository struct {
	db *gorm.DB
}

func NewTaskConversationResultRepository(db *gorm.DB) TaskConversationResultRepository {
	return &taskConversationResultRepository{db: db}
}

func (r *taskConversationResultRepository) Create(result *database.TaskConversationResult) error {
	return r.db.Create(result).Error
}


func (r *taskConversationResultRepository) ExistsByConversationID(conversationID uint) (bool, error) {
	var count int64
	err := r.db.Model(&database.TaskConversationResult{}).
		Where("conversation_id = ?", conversationID).
		Count(&count).Error
	return count > 0, err
}

func (r *taskConversationResultRepository) DeleteByConversationID(conversationID uint) error {
	return r.db.Where("conversation_id = ?", conversationID).
		Delete(&database.TaskConversationResult{}).Error
}

func (r *taskConversationResultRepository) GetLatestByTaskID(taskID uint) (*database.TaskConversationResult, error) {
	var result database.TaskConversationResult

	subQuery := r.db.Model(&database.TaskConversation{}).
		Select("id").
		Where("task_id = ?", taskID)

	err := r.db.Preload("Conversation").
		Preload("Conversation.Task").
		Where("conversation_id IN (?)", subQuery).
		Order("created_at DESC").
		First(&result).Error

	if err != nil {
		return nil, err
	}
	return &result, nil
}
