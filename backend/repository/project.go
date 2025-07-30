package repository

import (
	"time"
	"xsha-backend/database"

	"gorm.io/gorm"
)

type projectRepository struct {
	db *gorm.DB
}

// NewProjectRepository 创建项目仓库实例
func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{db: db}
}

// Create 创建项目
func (r *projectRepository) Create(project *database.Project) error {
	return r.db.Create(project).Error
}

// GetByID 根据ID获取项目
func (r *projectRepository) GetByID(id uint, createdBy string) (*database.Project, error) {
	var project database.Project
	err := r.db.Preload("Credential").Where("id = ? AND created_by = ?", id, createdBy).First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// GetByName 根据名称获取项目
func (r *projectRepository) GetByName(name, createdBy string) (*database.Project, error) {
	var project database.Project
	err := r.db.Preload("Credential").Where("name = ? AND created_by = ?", name, createdBy).First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// List 分页获取项目列表
func (r *projectRepository) List(createdBy string, name string, protocol *database.GitProtocolType, page, pageSize int) ([]database.Project, int64, error) {
	var projects []database.Project
	var total int64

	query := r.db.Model(&database.Project{}).Where("created_by = ?", createdBy)

	// 如果提供了名称筛选条件，添加模糊查询
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	if protocol != nil {
		query = query.Where("protocol = ?", *protocol)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询，预加载凭据信息
	offset := (page - 1) * pageSize
	if err := query.Preload("Credential").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&projects).Error; err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

// Update 更新项目
func (r *projectRepository) Update(project *database.Project) error {
	return r.db.Save(project).Error
}

// Delete 删除项目
func (r *projectRepository) Delete(id uint, createdBy string) error {
	return r.db.Where("id = ? AND created_by = ?", id, createdBy).Delete(&database.Project{}).Error
}

// UpdateLastUsed 更新最后使用时间
func (r *projectRepository) UpdateLastUsed(id uint, createdBy string) error {
	now := time.Now()
	return r.db.Model(&database.Project{}).
		Where("id = ? AND created_by = ?", id, createdBy).
		Update("last_used", now).Error
}

// GetByCredentialID 根据凭据ID获取关联的项目
func (r *projectRepository) GetByCredentialID(credentialID uint, createdBy string) ([]database.Project, error) {
	var projects []database.Project
	err := r.db.Where("credential_id = ? AND created_by = ?", credentialID, createdBy).Find(&projects).Error
	return projects, err
}

// GetTaskCounts 批量获取项目的任务统计数量
func (r *projectRepository) GetTaskCounts(projectIDs []uint, createdBy string) (map[uint]int64, error) {
	if len(projectIDs) == 0 {
		return make(map[uint]int64), nil
	}

	type TaskCountResult struct {
		ProjectID uint  `gorm:"column:project_id"`
		Count     int64 `gorm:"column:count"`
	}

	var results []TaskCountResult
	err := r.db.Table("tasks").
		Select("project_id, COUNT(*) as count").
		Where("project_id IN ? AND created_by = ?", projectIDs, createdBy).
		Group("project_id").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	// 构建返回的map，确保所有项目都有统计数据（包括任务数为0的项目）
	taskCounts := make(map[uint]int64)
	for _, projectID := range projectIDs {
		taskCounts[projectID] = 0 // 默认为0
	}

	// 填充实际的统计数据
	for _, result := range results {
		taskCounts[result.ProjectID] = result.Count
	}

	return taskCounts, nil
}
