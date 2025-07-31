package repository

import (
	"xsha-backend/database"

	"gorm.io/gorm"
)

type taskExecutionLogRepository struct {
	db *gorm.DB
}

func NewTaskExecutionLogRepository(db *gorm.DB) TaskExecutionLogRepository {
	return &taskExecutionLogRepository{db: db}
}

func (r *taskExecutionLogRepository) Create(log *database.TaskExecutionLog) error {
	return r.db.Create(log).Error
}

func (r *taskExecutionLogRepository) GetByID(id uint) (*database.TaskExecutionLog, error) {
	var log database.TaskExecutionLog
	err := r.db.Preload("Conversation").First(&log, id).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *taskExecutionLogRepository) GetByConversationID(conversationID uint) (*database.TaskExecutionLog, error) {
	var log database.TaskExecutionLog
	err := r.db.Where("conversation_id = ?", conversationID).First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *taskExecutionLogRepository) Update(log *database.TaskExecutionLog) error {
	return r.db.Save(log).Error
}

func (r *taskExecutionLogRepository) AppendLog(id uint, logContent string) error {
	return r.db.Model(&database.TaskExecutionLog{}).
		Where("id = ?", id).
		Update("execution_logs", gorm.Expr("CONCAT(COALESCE(execution_logs, ''), ?)", logContent)).Error
}

func (r *taskExecutionLogRepository) UpdateMetadata(id uint, updates map[string]interface{}) error {
	allowedFields := map[string]bool{
		"error_message":  true,
		"started_at":     true,
		"completed_at":   true,
		"docker_command": true,
	}

	filteredUpdates := make(map[string]interface{})
	for key, value := range updates {
		if allowedFields[key] {
			filteredUpdates[key] = value
		}
	}

	if len(filteredUpdates) == 0 {
		return nil
	}

	return r.db.Model(&database.TaskExecutionLog{}).
		Where("id = ?", id).
		Updates(filteredUpdates).Error
}

func (r *taskExecutionLogRepository) DeleteByConversationID(conversationID uint) error {
	return r.db.Where("conversation_id = ?", conversationID).Delete(&database.TaskExecutionLog{}).Error
}
