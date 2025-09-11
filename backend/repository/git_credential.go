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
	if err := query.
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
		Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&credentials).Error; err != nil {
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
	// First, get credential IDs that the admin has access to
	var credentialIDs []uint

	subQuery := r.db.Table("git_credentials").
		Select("DISTINCT git_credentials.id").
		Joins("LEFT JOIN git_credential_admins gca ON git_credentials.id = gca.git_credential_id").
		Where("gca.admin_id = ? OR git_credentials.admin_id = ?", adminID, adminID)

	if name != nil && *name != "" {
		subQuery = subQuery.Where("git_credentials.name LIKE ?", "%"+*name+"%")
	}

	if credType != nil {
		subQuery = subQuery.Where("git_credentials.type = ?", *credType)
	}

	if err := subQuery.Pluck("id", &credentialIDs).Error; err != nil {
		return nil, 0, err
	}

	if len(credentialIDs) == 0 {
		return []database.GitCredential{}, 0, nil
	}

	// Count total
	total := int64(len(credentialIDs))

	// Get paginated results using the credential IDs
	var credentials []database.GitCredential
	offset := (page - 1) * pageSize

	query := r.db.Where("id IN ?", credentialIDs).
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
		Order("created_at DESC").Offset(offset).Limit(pageSize)

	if err := query.Find(&credentials).Error; err != nil {
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

func (r *gitCredentialRepository) DeleteAdminAssociations(credentialID uint) error {
	return r.db.Exec("DELETE FROM git_credential_admins WHERE git_credential_id = ?", credentialID).Error
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

func (r *gitCredentialRepository) IsOwner(credentialID, adminID uint) (bool, error) {
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

	return false, nil
}

// CountByAdminID counts the number of git credentials created by a specific admin
func (r *gitCredentialRepository) CountByAdminID(adminID uint) (int64, error) {
	var count int64
	err := r.db.Model(&database.GitCredential{}).
		Where("admin_id = ?", adminID).
		Count(&count).Error
	return count, err
}
