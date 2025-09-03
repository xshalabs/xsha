package repository

import (
	"xsha-backend/database"

	"gorm.io/gorm"
)

type gitCredentialRepository struct {
	db *gorm.DB
}

func NewGitCredentialRepository(db *gorm.DB) GitCredentialRepository {
	return &gitCredentialRepository{db: db}
}

func (r *gitCredentialRepository) Create(credential *database.GitCredential) error {
	return r.db.Create(credential).Error
}

func (r *gitCredentialRepository) GetByID(id uint) (*database.GitCredential, error) {
	var credential database.GitCredential
	err := r.db.Where("id = ?", id).First(&credential).Error
	if err != nil {
		return nil, err
	}
	return &credential, nil
}

func (r *gitCredentialRepository) GetByName(name string) (*database.GitCredential, error) {
	var credential database.GitCredential
	err := r.db.Where("name = ?", name).First(&credential).Error
	if err != nil {
		return nil, err
	}
	return &credential, nil
}

func (r *gitCredentialRepository) List(name *string, credType *database.GitCredentialType, page, pageSize int) ([]database.GitCredential, int64, error) {
	var credentials []database.GitCredential
	var total int64

	query := r.db.Model(&database.GitCredential{})

	if name != nil && *name != "" {
		query = query.Where("name LIKE ?", "%"+*name+"%")
	}

	if credType != nil {
		query = query.Where("type = ?", *credType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&credentials).Error; err != nil {
		return nil, 0, err
	}

	return credentials, total, nil
}

func (r *gitCredentialRepository) Update(credential *database.GitCredential) error {
	return r.db.Save(credential).Error
}

func (r *gitCredentialRepository) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&database.GitCredential{}).Error
}

func (r *gitCredentialRepository) GetByIDWithAdmins(id uint) (*database.GitCredential, error) {
	var credential database.GitCredential
	err := r.db.Preload("Admins").Where("id = ?", id).First(&credential).Error
	if err != nil {
		return nil, err
	}
	return &credential, nil
}

func (r *gitCredentialRepository) ListByAdminAccess(adminID uint, name *string, credType *database.GitCredentialType, page, pageSize int) ([]database.GitCredential, int64, error) {
	var credentials []database.GitCredential
	var total int64

	// Base query for filtering
	baseQuery := r.db.Model(&database.GitCredential{}).
		Joins("LEFT JOIN git_credential_admins gca ON git_credentials.id = gca.git_credential_id").
		Where("gca.admin_id = ? OR git_credentials.admin_id = ?", adminID, adminID)

	if name != nil && *name != "" {
		baseQuery = baseQuery.Where("git_credentials.name LIKE ?", "%"+*name+"%")
	}

	if credType != nil {
		baseQuery = baseQuery.Where("git_credentials.type = ?", *credType)
	}

	// Count total distinct credentials
	if err := baseQuery.Distinct("git_credentials.id").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	if err := baseQuery.
		Preload("Admin", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, username, name, email, avatar_id").
				Preload("Avatar", func(db *gorm.DB) *gorm.DB {
					return db.Select("id, uuid, original_name")
				})
		}).
		Preload("Admins", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, username, name, email, avatar_id").
				Preload("Avatar", func(db *gorm.DB) *gorm.DB {
					return db.Select("id, uuid, original_name")
				})
		}).
		Distinct("git_credentials.id").
		Order("git_credentials.created_at DESC").Offset(offset).Limit(pageSize).Find(&credentials).Error; err != nil {
		return nil, 0, err
	}

	return credentials, total, nil
}

func (r *gitCredentialRepository) AddAdmin(credentialID, adminID uint) error {
	// Check if relationship already exists
	var count int64
	if err := r.db.Table("git_credential_admins").
		Where("git_credential_id = ? AND admin_id = ?", credentialID, adminID).
		Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return nil // Already exists, no error
	}

	// Add the relationship
	return r.db.Exec("INSERT INTO git_credential_admins (git_credential_id, admin_id) VALUES (?, ?)", credentialID, adminID).Error
}

func (r *gitCredentialRepository) RemoveAdmin(credentialID, adminID uint) error {
	return r.db.Exec("DELETE FROM git_credential_admins WHERE git_credential_id = ? AND admin_id = ?", credentialID, adminID).Error
}

func (r *gitCredentialRepository) GetAdmins(credentialID uint) ([]database.Admin, error) {
	var admins []database.Admin
	err := r.db.
		Table("admins").
		Select("admins.id, admins.username, admins.name, admins.email, admins.avatar_id").
		Joins("JOIN git_credential_admins gca ON admins.id = gca.admin_id").
		Where("gca.git_credential_id = ?", credentialID).
		Preload("Avatar", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, uuid, original_name")
		}).
		Find(&admins).Error

	return admins, err
}

func (r *gitCredentialRepository) IsAdminForCredential(credentialID, adminID uint) (bool, error) {
	// Check both legacy AdminID and many-to-many relationship
	var count int64

	// Check legacy relationship first
	if err := r.db.Model(&database.GitCredential{}).
		Where("id = ? AND admin_id = ?", credentialID, adminID).
		Count(&count).Error; err != nil {
		return false, err
	}

	if count > 0 {
		return true, nil
	}

	// Check many-to-many relationship
	if err := r.db.Table("git_credential_admins").
		Where("git_credential_id = ? AND admin_id = ?", credentialID, adminID).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}
