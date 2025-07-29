package repository

import (
	"xsha-backend/database"

	"gorm.io/gorm"
)

type gitCredentialRepository struct {
	db *gorm.DB
}

// NewGitCredentialRepository creates a new Git credential repository instance
func NewGitCredentialRepository(db *gorm.DB) GitCredentialRepository {
	return &gitCredentialRepository{db: db}
}

// Create creates a Git credential
func (r *gitCredentialRepository) Create(credential *database.GitCredential) error {
	return r.db.Create(credential).Error
}

// GetByID gets a Git credential by ID
func (r *gitCredentialRepository) GetByID(id uint, createdBy string) (*database.GitCredential, error) {
	var credential database.GitCredential
	err := r.db.Where("id = ? AND created_by = ?", id, createdBy).First(&credential).Error
	if err != nil {
		return nil, err
	}
	return &credential, nil
}

// GetByName gets a Git credential by name
func (r *gitCredentialRepository) GetByName(name, createdBy string) (*database.GitCredential, error) {
	var credential database.GitCredential
	err := r.db.Where("name = ? AND created_by = ?", name, createdBy).First(&credential).Error
	if err != nil {
		return nil, err
	}
	return &credential, nil
}

// List gets a paginated list of Git credentials
func (r *gitCredentialRepository) List(createdBy string, credType *database.GitCredentialType, page, pageSize int) ([]database.GitCredential, int64, error) {
	var credentials []database.GitCredential
	var total int64

	query := r.db.Model(&database.GitCredential{}).Where("created_by = ?", createdBy)

	if credType != nil {
		query = query.Where("type = ?", *credType)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Paginated query
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&credentials).Error; err != nil {
		return nil, 0, err
	}

	return credentials, total, nil
}

// Update updates a Git credential
func (r *gitCredentialRepository) Update(credential *database.GitCredential) error {
	return r.db.Save(credential).Error
}

// Delete deletes a Git credential
func (r *gitCredentialRepository) Delete(id uint, createdBy string) error {
	return r.db.Where("id = ? AND created_by = ?", id, createdBy).Delete(&database.GitCredential{}).Error
}
