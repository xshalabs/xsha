package repository

import (
	"time"
	"xsha-backend/database"

	"gorm.io/gorm"
)

type projectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) Create(project *database.Project) error {
	return r.db.Create(project).Error
}

func (r *projectRepository) GetByID(id uint) (*database.Project, error) {
	var project database.Project
	err := r.db.Preload("Credential").Where("id = ?", id).First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *projectRepository) GetByName(name string) (*database.Project, error) {
	var project database.Project
	err := r.db.Preload("Credential").Where("name = ?", name).First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *projectRepository) List(name string, protocol *database.GitProtocolType, page, pageSize int) ([]database.Project, int64, error) {
	var projects []database.Project
	var total int64

	query := r.db.Model(&database.Project{})

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	if protocol != nil {
		query = query.Where("protocol = ?", *protocol)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Credential").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&projects).Error; err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

func (r *projectRepository) Update(project *database.Project) error {
	return r.db.Save(project).Error
}

func (r *projectRepository) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&database.Project{}).Error
}

func (r *projectRepository) UpdateLastUsed(id uint) error {
	now := time.Now()
	return r.db.Model(&database.Project{}).
		Where("id = ?", id).
		Update("last_used", now).Error
}

func (r *projectRepository) GetByCredentialID(credentialID uint) ([]database.Project, error) {
	var projects []database.Project
	err := r.db.Where("credential_id = ?", credentialID).Find(&projects).Error
	return projects, err
}

func (r *projectRepository) GetTaskCounts(projectIDs []uint) (map[uint]int64, error) {
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
		Where("project_id IN ? AND deleted_at IS NULL", projectIDs).
		Group("project_id").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	taskCounts := make(map[uint]int64)
	for _, projectID := range projectIDs {
		taskCounts[projectID] = 0
	}

	for _, result := range results {
		taskCounts[result.ProjectID] = result.Count
	}

	return taskCounts, nil
}
