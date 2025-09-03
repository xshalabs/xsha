package repository

import (
	"time"
	"xsha-backend/database"
	"xsha-backend/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type adminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &adminRepository{db: db}
}

func (r *adminRepository) Create(admin *database.Admin) error {
	return r.db.Create(admin).Error
}

func (r *adminRepository) GetByID(id uint) (*database.Admin, error) {
	var admin database.Admin
	err := r.db.Preload("Avatar").First(&admin, id).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *adminRepository) GetByUsername(username string) (*database.Admin, error) {
	var admin database.Admin
	err := r.db.Preload("Avatar").Where("username = ?", username).First(&admin).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *adminRepository) List(search *string, isActive *bool, page, pageSize int) ([]database.Admin, int64, error) {
	var admins []database.Admin
	var total int64

	query := r.db.Model(&database.Admin{})

	if search != nil && *search != "" {
		searchPattern := "%" + *search + "%"
		query = query.Where("username LIKE ? OR name LIKE ? OR email LIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and fetch records with minimal avatar fields
	offset := (page - 1) * pageSize
	err := query.Preload("Avatar", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, uuid, original_name")
	}).Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&admins).Error
	return admins, total, err
}

func (r *adminRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&database.Admin{}).Where("id = ?", id).Updates(updates).Error
}

func (r *adminRepository) Delete(id uint) error {
	return r.db.Delete(&database.Admin{}, id).Error
}

func (r *adminRepository) UpdateLastLogin(username, ip string) error {
	now := time.Now()
	return r.db.Model(&database.Admin{}).
		Where("username = ?", username).
		Updates(map[string]interface{}{
			"last_login_at": &now,
			"last_login_ip": ip,
		}).Error
}

func (r *adminRepository) CountAdmins() (int64, error) {
	var count int64
	err := r.db.Model(&database.Admin{}).Where("is_active = ?", true).Count(&count).Error
	return count, err
}

func (r *adminRepository) InitializeDefaultAdmin() error {
	// Check if any admin exists
	count, err := r.CountAdmins()
	if err != nil {
		return err
	}

	if count > 0 {
		utils.Info("Admin users already exist, skipping default admin creation")
		return nil
	}

	// Create default admin with xshauser/xshapass
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("xshapass"), bcrypt.DefaultCost)
	if err != nil {
		utils.Error("Failed to hash default admin password", "error", err)
		return err
	}

	defaultAdmin := &database.Admin{
		Username:     "xshauser",
		PasswordHash: string(passwordHash),
		Name:         "XSha Administrator",
		Email:        "",
		Role:         database.AdminRoleSuperAdmin,
		IsActive:     true,
		CreatedBy:    "system",
	}

	if err := r.Create(defaultAdmin); err != nil {
		utils.Error("Failed to create default admin user", "error", err)
		return err
	}

	utils.Info("Default admin user created successfully", "username", "xshauser")
	return nil
}

// CountActiveAdminsByRole counts active admins by specific role
func (r *adminRepository) CountActiveAdminsByRole(role database.AdminRole) (int64, error) {
	var count int64
	err := r.db.Model(&database.Admin{}).Where("is_active = ? AND role = ?", true, role).Count(&count).Error
	return count, err
}
