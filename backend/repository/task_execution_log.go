package repository

import (
	"sleep0-backend/database"

	"gorm.io/gorm"
)

type taskExecutionLogRepository struct {
	db *gorm.DB
}

// NewTaskExecutionLogRepository 创建任务执行日志仓库实例
func NewTaskExecutionLogRepository(db *gorm.DB) TaskExecutionLogRepository {
	return &taskExecutionLogRepository{db: db}
}

// Create 创建执行日志
func (r *taskExecutionLogRepository) Create(log *database.TaskExecutionLog) error {
	return r.db.Create(log).Error
}

// GetByID 根据ID获取执行日志
func (r *taskExecutionLogRepository) GetByID(id uint) (*database.TaskExecutionLog, error) {
	var log database.TaskExecutionLog
	err := r.db.Preload("Conversation").First(&log, id).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// GetByConversationID 根据对话ID获取执行日志
func (r *taskExecutionLogRepository) GetByConversationID(conversationID uint) (*database.TaskExecutionLog, error) {
	var log database.TaskExecutionLog
	err := r.db.Where("conversation_id = ?", conversationID).First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// Update 更新执行日志
func (r *taskExecutionLogRepository) Update(log *database.TaskExecutionLog) error {
	return r.db.Save(log).Error
}

// AppendLog 追加执行日志
func (r *taskExecutionLogRepository) AppendLog(id uint, logContent string) error {
	return r.db.Model(&database.TaskExecutionLog{}).
		Where("id = ?", id).
		Update("execution_logs", gorm.Expr("CONCAT(COALESCE(execution_logs, ''), ?)", logContent)).Error
}

// UpdateMetadata 更新执行日志的元数据信息，不覆盖 execution_logs 字段
func (r *taskExecutionLogRepository) UpdateMetadata(id uint, updates map[string]interface{}) error {
	// 明确排除 execution_logs 字段，避免意外覆盖
	allowedFields := map[string]bool{
		"error_message":  true,
		"commit_hash":    true,
		"started_at":     true,
		"completed_at":   true,
		"workspace_path": true,
		"docker_command": true,
	}

	filteredUpdates := make(map[string]interface{})
	for key, value := range updates {
		if allowedFields[key] {
			filteredUpdates[key] = value
		}
	}

	if len(filteredUpdates) == 0 {
		return nil // 没有需要更新的字段
	}

	return r.db.Model(&database.TaskExecutionLog{}).
		Where("id = ?", id).
		Updates(filteredUpdates).Error
}

// DeleteByConversationID 删除指定对话ID的所有执行日志
func (r *taskExecutionLogRepository) DeleteByConversationID(conversationID uint) error {
	return r.db.Where("conversation_id = ?", conversationID).Delete(&database.TaskExecutionLog{}).Error
}
