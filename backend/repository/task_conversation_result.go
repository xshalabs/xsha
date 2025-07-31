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

func (r *taskConversationResultRepository) Update(result *database.TaskConversationResult) error {
	return r.db.Save(result).Error
}

func (r *taskConversationResultRepository) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&database.TaskConversationResult{}).Error
}

func (r *taskConversationResultRepository) ListByTaskID(taskID uint, page, pageSize int) ([]database.TaskConversationResult, int64, error) {
	var results []database.TaskConversationResult
	var total int64

	subQuery := r.db.Model(&database.TaskConversation{}).
		Select("id").
		Where("task_id = ?", taskID)

	query := r.db.Model(&database.TaskConversationResult{}).
		Preload("Conversation").
		Where("conversation_id IN (?)", subQuery)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&results).Error; err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

func (r *taskConversationResultRepository) ListByProjectID(projectID uint, page, pageSize int) ([]database.TaskConversationResult, int64, error) {
	var results []database.TaskConversationResult
	var total int64

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

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&results).Error; err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

func (r *taskConversationResultRepository) GetSuccessRate(taskID uint) (float64, error) {
	var totalCount, successCount int64

	subQuery := r.db.Model(&database.TaskConversation{}).
		Select("id").
		Where("task_id = ?", taskID)

	if err := r.db.Model(&database.TaskConversationResult{}).
		Where("conversation_id IN (?)", subQuery).
		Count(&totalCount).Error; err != nil {
		return 0, err
	}

	if totalCount == 0 {
		return 0, nil
	}

	if err := r.db.Model(&database.TaskConversationResult{}).
		Where("conversation_id IN (?) AND is_error = ?", subQuery, false).
		Count(&successCount).Error; err != nil {
		return 0, err
	}

	return float64(successCount) / float64(totalCount), nil
}

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
