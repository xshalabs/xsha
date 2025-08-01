package repository

import (
	"xsha-backend/database"

	"gorm.io/gorm"
)

type taskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(task *database.Task) error {
	return r.db.Create(task).Error
}

func (r *taskRepository) GetByID(id uint) (*database.Task, error) {
	var task database.Task
	err := r.db.Preload("Project").Preload("DevEnvironment").Preload("Conversations").
		Where("id = ?", id).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *taskRepository) List(projectID *uint, status *database.TaskStatus, title *string, branch *string, devEnvID *uint, page, pageSize int) ([]database.Task, int64, error) {
	var tasks []database.Task
	var total int64

	query := r.db.Model(&database.Task{})

	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if title != nil && *title != "" {
		query = query.Where("title LIKE ?", "%"+*title+"%")
	}

	if branch != nil && *branch != "" {
		query = query.Where("start_branch = ?", *branch)
	}

	if devEnvID != nil {
		query = query.Where("dev_environment_id = ?", *devEnvID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Project").Preload("DevEnvironment").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

func (r *taskRepository) Update(task *database.Task) error {
	return r.db.Save(task).Error
}

func (r *taskRepository) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&database.Task{}).Error
}

func (r *taskRepository) ListByProject(projectID uint) ([]database.Task, error) {
	var tasks []database.Task
	err := r.db.Where("project_id = ?", projectID).
		Order("created_at DESC").Find(&tasks).Error
	return tasks, err
}

func (r *taskRepository) GetConversationCounts(taskIDs []uint) (map[uint]int64, error) {
	if len(taskIDs) == 0 {
		return make(map[uint]int64), nil
	}

	type ConversationCountResult struct {
		TaskID uint  `gorm:"column:task_id"`
		Count  int64 `gorm:"column:count"`
	}

	var results []ConversationCountResult
	err := r.db.Table("task_conversations").
		Select("task_id, COUNT(*) as count").
		Where("task_id IN ? AND deleted_at IS NULL", taskIDs).
		Group("task_id").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	conversationCounts := make(map[uint]int64)
	for _, taskID := range taskIDs {
		conversationCounts[taskID] = 0
	}

	for _, result := range results {
		conversationCounts[result.TaskID] = result.Count
	}

	return conversationCounts, nil
}
