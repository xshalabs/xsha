package repository

import (
	"sleep0-backend/database"
	"time"

	"gorm.io/gorm"
)

type gitCredentialRepository struct {
	db *gorm.DB
}

// NewGitCredentialRepository 创建Git凭据仓库实例
func NewGitCredentialRepository(db *gorm.DB) GitCredentialRepository {
	return &gitCredentialRepository{db: db}
}

// Create 创建Git凭据
func (r *gitCredentialRepository) Create(credential *database.GitCredential) error {
	return r.db.Create(credential).Error
}

// GetByID 根据ID获取Git凭据
func (r *gitCredentialRepository) GetByID(id uint, createdBy string) (*database.GitCredential, error) {
	var credential database.GitCredential
	err := r.db.Where("id = ? AND created_by = ?", id, createdBy).First(&credential).Error
	if err != nil {
		return nil, err
	}
	return &credential, nil
}

// GetByName 根据名称获取Git凭据
func (r *gitCredentialRepository) GetByName(name, createdBy string) (*database.GitCredential, error) {
	var credential database.GitCredential
	err := r.db.Where("name = ? AND created_by = ?", name, createdBy).First(&credential).Error
	if err != nil {
		return nil, err
	}
	return &credential, nil
}

// List 分页获取Git凭据列表
func (r *gitCredentialRepository) List(createdBy string, credType *database.GitCredentialType, page, pageSize int) ([]database.GitCredential, int64, error) {
	var credentials []database.GitCredential
	var total int64

	query := r.db.Model(&database.GitCredential{}).Where("created_by = ?", createdBy)

	if credType != nil {
		query = query.Where("type = ?", *credType)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&credentials).Error; err != nil {
		return nil, 0, err
	}

	return credentials, total, nil
}

// Update 更新Git凭据
func (r *gitCredentialRepository) Update(credential *database.GitCredential) error {
	return r.db.Save(credential).Error
}

// Delete 删除Git凭据
func (r *gitCredentialRepository) Delete(id uint, createdBy string) error {
	return r.db.Where("id = ? AND created_by = ?", id, createdBy).Delete(&database.GitCredential{}).Error
}

// UpdateLastUsed 更新最后使用时间
func (r *gitCredentialRepository) UpdateLastUsed(id uint, createdBy string) error {
	now := time.Now()
	return r.db.Model(&database.GitCredential{}).
		Where("id = ? AND created_by = ?", id, createdBy).
		Update("last_used", now).Error
}

// SetActive 设置激活状态
func (r *gitCredentialRepository) SetActive(id uint, createdBy string, isActive bool) error {
	return r.db.Model(&database.GitCredential{}).
		Where("id = ? AND created_by = ?", id, createdBy).
		Update("is_active", isActive).Error
}

// ListActive 获取激活的凭据列表
func (r *gitCredentialRepository) ListActive(createdBy string, credType *database.GitCredentialType) ([]database.GitCredential, error) {
	var credentials []database.GitCredential

	query := r.db.Where("created_by = ? AND is_active = ?", createdBy, true)
	if credType != nil {
		query = query.Where("type = ?", *credType)
	}

	err := query.Order("created_at DESC").Find(&credentials).Error
	return credentials, err
}
