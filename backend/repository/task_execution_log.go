package repository

import (
	"sleep0-backend/database"
	"time"

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

// UpdateStatus 更新执行状态
func (r *taskExecutionLogRepository) UpdateStatus(id uint, status database.TaskExecutionStatus) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if status == database.TaskExecStatusRunning {
		updates["started_at"] = time.Now()
	} else if status == database.TaskExecStatusSuccess || status == database.TaskExecStatusFailed || status == database.TaskExecStatusCancelled {
		updates["completed_at"] = time.Now()
	}

	return r.db.Model(&database.TaskExecutionLog{}).Where("id = ?", id).Updates(updates).Error
}

// AppendLog 追加执行日志
func (r *taskExecutionLogRepository) AppendLog(id uint, logContent string) error {
	return r.db.Model(&database.TaskExecutionLog{}).
		Where("id = ?", id).
		Update("execution_logs", gorm.Expr("CONCAT(execution_logs, ?)", logContent)).Error
}

// ListByStatus 根据状态获取执行日志列表
func (r *taskExecutionLogRepository) ListByStatus(status database.TaskExecutionStatus, limit int) ([]database.TaskExecutionLog, error) {
	var logs []database.TaskExecutionLog
	query := r.db.Where("status = ?", status).Order("created_at ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&logs).Error
	return logs, err
}

// DeleteByConversationID 删除指定对话ID的所有执行日志
func (r *taskExecutionLogRepository) DeleteByConversationID(conversationID uint) error {
	return r.db.Where("conversation_id = ?", conversationID).Delete(&database.TaskExecutionLog{}).Error
}
