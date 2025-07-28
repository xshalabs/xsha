package repository

import (
	"xsha-backend/database"

	"gorm.io/gorm"
)

type taskConversationResultRepository struct {
	db *gorm.DB
}

// NewTaskConversationResultRepository 创建任务对话结果仓库实例
func NewTaskConversationResultRepository(db *gorm.DB) TaskConversationResultRepository {
	return &taskConversationResultRepository{db: db}
}

// Create 创建对话结果
func (r *taskConversationResultRepository) Create(result *database.TaskConversationResult) error {
	return r.db.Create(result).Error
}

// GetByID 根据ID获取对话结果
func (r *taskConversationResultRepository) GetByID(id uint) (*database.TaskConversationResult, error) {
	var result database.TaskConversationResult
	err := r.db.Preload("Conversation").
		Preload("Conversation.Task").
		Preload("Conversation.Task.Project").
		Where("id = ?", id).First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetByConversationID 根据对话ID获取结果
func (r *taskConversationResultRepository) GetByConversationID(conversationID uint) (*database.TaskConversationResult, error) {
	var result database.TaskConversationResult
	err := r.db.Preload("Conversation").
		Preload("Conversation.Task").
		Preload("Conversation.Task.Project").
		Where("conversation_id = ?", conversationID).First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Update 更新对话结果
func (r *taskConversationResultRepository) Update(result *database.TaskConversationResult) error {
	return r.db.Save(result).Error
}

// Delete 删除对话结果
func (r *taskConversationResultRepository) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&database.TaskConversationResult{}).Error
}

// ListByTaskID 根据任务ID分页获取结果列表
func (r *taskConversationResultRepository) ListByTaskID(taskID uint, page, pageSize int) ([]database.TaskConversationResult, int64, error) {
	var results []database.TaskConversationResult
	var total int64

	subQuery := r.db.Model(&database.TaskConversation{}).
		Select("id").
		Where("task_id = ?", taskID)

	query := r.db.Model(&database.TaskConversationResult{}).
		Preload("Conversation").
		Where("conversation_id IN (?)", subQuery)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&results).Error; err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

// ListByProjectID 根据项目ID分页获取结果列表
func (r *taskConversationResultRepository) ListByProjectID(projectID uint, page, pageSize int) ([]database.TaskConversationResult, int64, error) {
	var results []database.TaskConversationResult
	var total int64

	// 构建子查询：先找到项目下的所有任务，再找到任务下的所有对话
	taskSubQuery := r.db.Model(&database.Task{}).
		Select("id").
		Where("project_id = ?", projectID)

	conversationSubQuery := r.db.Model(&database.TaskConversation{}).
		Select("id").
		Where("task_id IN (?)", taskSubQuery)

	query := r.db.Model(&database.TaskConversationResult{}).
		Preload("Conversation").
		Preload("Conversation.Task").
		Where("conversation_id IN (?)", conversationSubQuery)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&results).Error; err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

// GetSuccessRate 获取任务成功率
func (r *taskConversationResultRepository) GetSuccessRate(taskID uint) (float64, error) {
	var totalCount, successCount int64

	subQuery := r.db.Model(&database.TaskConversation{}).
		Select("id").
		Where("task_id = ?", taskID)

	// 获取总数
	if err := r.db.Model(&database.TaskConversationResult{}).
		Where("conversation_id IN (?)", subQuery).
		Count(&totalCount).Error; err != nil {
		return 0, err
	}

	if totalCount == 0 {
		return 0, nil
	}

	// 获取成功数
	if err := r.db.Model(&database.TaskConversationResult{}).
		Where("conversation_id IN (?) AND is_error = ?", subQuery, false).
		Count(&successCount).Error; err != nil {
		return 0, err
	}

	return float64(successCount) / float64(totalCount), nil
}

// GetTotalCost 获取任务总成本
func (r *taskConversationResultRepository) GetTotalCost(taskID uint) (float64, error) {
	var totalCost float64

	subQuery := r.db.Model(&database.TaskConversation{}).
		Select("id").
		Where("task_id = ?", taskID)

	err := r.db.Model(&database.TaskConversationResult{}).
		Where("conversation_id IN (?)", subQuery).
		Select("COALESCE(SUM(total_cost_usd), 0)").
		Scan(&totalCost).Error

	return totalCost, err
}

// GetAverageDuration 获取任务平均执行时间
func (r *taskConversationResultRepository) GetAverageDuration(taskID uint) (float64, error) {
	var avgDuration float64

	subQuery := r.db.Model(&database.TaskConversation{}).
		Select("id").
		Where("task_id = ?", taskID)

	err := r.db.Model(&database.TaskConversationResult{}).
		Where("conversation_id IN (?)", subQuery).
		Select("COALESCE(AVG(duration_ms), 0)").
		Scan(&avgDuration).Error

	return avgDuration, err
}

// ExistsByConversationID 检查对话是否已有结果记录
func (r *taskConversationResultRepository) ExistsByConversationID(conversationID uint) (bool, error) {
	var count int64
	err := r.db.Model(&database.TaskConversationResult{}).
		Where("conversation_id = ?", conversationID).
		Count(&count).Error
	return count > 0, err
}

// DeleteByConversationID 根据对话ID删除结果
func (r *taskConversationResultRepository) DeleteByConversationID(conversationID uint) error {
	return r.db.Where("conversation_id = ?", conversationID).
		Delete(&database.TaskConversationResult{}).Error
}
