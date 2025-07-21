package repository

import (
	"sleep0-backend/database"
	"time"

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
func (r *projectRepository) List(createdBy string, protocol *database.GitProtocolType, page, pageSize int) ([]database.Project, int64, error) {
	var projects []database.Project
	var total int64

	query := r.db.Model(&database.Project{}).Where("created_by = ?", createdBy)

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

// SetActive 设置激活状态
func (r *projectRepository) SetActive(id uint, createdBy string, isActive bool) error {
	return r.db.Model(&database.Project{}).
		Where("id = ? AND created_by = ?", id, createdBy).
		Update("is_active", isActive).Error
}

// ListActive 获取激活的项目列表
func (r *projectRepository) ListActive(createdBy string, protocol *database.GitProtocolType) ([]database.Project, error) {
	var projects []database.Project

	query := r.db.Preload("Credential").Where("created_by = ? AND is_active = ?", createdBy, true)
	if protocol != nil {
		query = query.Where("protocol = ?", *protocol)
	}

	err := query.Order("created_at DESC").Find(&projects).Error
	return projects, err
}

// GetByCredentialID 根据凭据ID获取关联的项目
func (r *projectRepository) GetByCredentialID(credentialID uint, createdBy string) ([]database.Project, error) {
	var projects []database.Project
	err := r.db.Where("credential_id = ? AND created_by = ?", credentialID, createdBy).Find(&projects).Error
	return projects, err
}
