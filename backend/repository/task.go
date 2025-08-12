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

func (r *taskRepository) List(projectID *uint, statuses []database.TaskStatus, title *string, branch *string, devEnvID *uint, sortBy, sortDirection string, page, pageSize int) ([]database.Task, int64, error) {
	var tasks []database.Task
	var total int64

	query := r.db.Model(&database.Task{})

	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}

	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
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

	// Handle sorting
	var orderClause string
	switch sortBy {
	case "title":
		orderClause = "title " + sortDirection
	case "start_branch":
		orderClause = "start_branch " + sortDirection
	case "created_at":
		orderClause = "created_at " + sortDirection
	case "updated_at":
		orderClause = "updated_at " + sortDirection
	case "status":
		orderClause = "status " + sortDirection
	case "conversation_count":
		// For conversation_count sorting, we need to join with conversation counts
		subQuery := r.db.Table("task_conversations").
			Select("task_id, COUNT(*) as conversation_count").
			Where("deleted_at IS NULL").
			Group("task_id")

		query = query.
			Select("tasks.*, COALESCE(conversation_counts.conversation_count, 0) as conversation_count").
			Joins("LEFT JOIN (?) as conversation_counts ON tasks.id = conversation_counts.task_id", subQuery).
			Order("conversation_count " + sortDirection + ", tasks.created_at DESC")
	case "dev_environment_name":
		// For dev_environment_name sorting, we need to join with dev_environments table
		query = query.
			Select("tasks.*").
			Joins("LEFT JOIN dev_environments ON tasks.dev_environment_id = dev_environments.id").
			Order("dev_environments.name " + sortDirection + ", tasks.created_at DESC")
	default:
		orderClause = "created_at " + sortDirection
	}

	if sortBy != "conversation_count" && sortBy != "dev_environment_name" {
		query = query.Order(orderClause)
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Project").Preload("DevEnvironment").Offset(offset).Limit(pageSize).Find(&tasks).Error; err != nil {
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
