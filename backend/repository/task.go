package repository

import (
	"sleep0-backend/database"

	"gorm.io/gorm"
)

type taskRepository struct {
	db *gorm.DB
}

// NewTaskRepository 创建任务仓库实例
func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{db: db}
}

// Create 创建任务
func (r *taskRepository) Create(task *database.Task) error {
	return r.db.Create(task).Error
}

// GetByID 根据ID获取任务
func (r *taskRepository) GetByID(id uint, createdBy string) (*database.Task, error) {
	var task database.Task
	err := r.db.Preload("Project").Preload("DevEnvironment").Preload("Conversations").
		Where("id = ? AND created_by = ?", id, createdBy).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// List 分页获取任务列表
func (r *taskRepository) List(projectID *uint, createdBy string, status *database.TaskStatus, page, pageSize int) ([]database.Task, int64, error) {
	var tasks []database.Task
	var total int64

	query := r.db.Model(&database.Task{}).Where("created_by = ?", createdBy)

	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Preload("Project").Preload("DevEnvironment").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// Update 更新任务
func (r *taskRepository) Update(task *database.Task) error {
	return r.db.Save(task).Error
}

// Delete 删除任务
func (r *taskRepository) Delete(id uint, createdBy string) error {
	return r.db.Where("id = ? AND created_by = ?", id, createdBy).Delete(&database.Task{}).Error
}

// ListByProject 根据项目ID获取任务列表
func (r *taskRepository) ListByProject(projectID uint, createdBy string) ([]database.Task, error) {
	var tasks []database.Task
	err := r.db.Where("project_id = ? AND created_by = ?", projectID, createdBy).
		Order("created_at DESC").Find(&tasks).Error
	return tasks, err
}

// CountByStatus 统计各状态的任务数量
func (r *taskRepository) CountByStatus(projectID uint, createdBy string) (map[database.TaskStatus]int64, error) {
	type StatusCount struct {
		Status database.TaskStatus
		Count  int64
	}

	var results []StatusCount
	err := r.db.Model(&database.Task{}).
		Select("status, count(*) as count").
		Where("project_id = ? AND created_by = ?", projectID, createdBy).
		Group("status").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	counts := make(map[database.TaskStatus]int64)
	for _, result := range results {
		counts[result.Status] = result.Count
	}

	return counts, nil
}
