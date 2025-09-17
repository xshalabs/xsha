package repository

import (
	"xsha-backend/database"

	"gorm.io/gorm"
)

type devEnvironmentRepository struct {
	db *gorm.DB
}

func NewDevEnvironmentRepository(db *gorm.DB) DevEnvironmentRepository {
	return &devEnvironmentRepository{db: db}
}

func (r *devEnvironmentRepository) Create(env *database.DevEnvironment) error {
	return r.db.Create(env).Error
}

func (r *devEnvironmentRepository) GetByID(id uint) (*database.DevEnvironment, error) {
	var env database.DevEnvironment
	err := r.db.Where("id = ?", id).First(&env).Error
	if err != nil {
		return nil, err
	}
	return &env, nil
}

func (r *devEnvironmentRepository) GetByIDWithAdmins(id uint) (*database.DevEnvironment, error) {
	var env database.DevEnvironment
	err := r.db.Preload("Admins").Where("id = ?", id).First(&env).Error
	if err != nil {
		return nil, err
	}
	return &env, nil
}

func (r *devEnvironmentRepository) GetByName(name string) (*database.DevEnvironment, error) {
	var env database.DevEnvironment
	err := r.db.Where("name = ?", name).First(&env).Error
	if err != nil {
		return nil, err
	}
	return &env, nil
}

func (r *devEnvironmentRepository) List(name *string, dockerImage *string, page, pageSize int) ([]database.DevEnvironment, int64, error) {
	var environments []database.DevEnvironment
	var total int64

	query := r.db.Model(&database.DevEnvironment{})

	if name != nil && *name != "" {
		query = query.Where("name LIKE ?", "%"+*name+"%")
	}

	if dockerImage != nil && *dockerImage != "" {
		query = query.Where("docker_image LIKE ?", "%"+*dockerImage+"%")
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
		Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&environments).Error; err != nil {
		return nil, 0, err
	}

	return environments, total, nil
}

func (r *devEnvironmentRepository) ListByAdminAccess(adminID uint, name *string, dockerImage *string, page, pageSize int) ([]database.DevEnvironment, int64, error) {
	var environments []database.DevEnvironment
	var total int64

	// Base query for filtering
	baseQuery := r.db.Model(&database.DevEnvironment{}).
		Joins("LEFT JOIN dev_environment_admins dea ON dev_environments.id = dea.dev_environment_id").
		Where("dea.admin_id = ? OR dev_environments.admin_id = ?", adminID, adminID)

	if name != nil && *name != "" {
		baseQuery = baseQuery.Where("dev_environments.name LIKE ?", "%"+*name+"%")
	}

	if dockerImage != nil && *dockerImage != "" {
		baseQuery = baseQuery.Where("dev_environments.docker_image LIKE ?", "%"+*dockerImage+"%")
	}

	// Count distinct environments using a subquery to avoid GROUP BY issues
	countQuery := r.db.Model(&database.DevEnvironment{}).
		Where("id IN (?)", baseQuery.Select("DISTINCT dev_environments.id"))

	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get the actual results
	offset := (page - 1) * pageSize
	if err := baseQuery.Select("dev_environments.*").Group("dev_environments.id").
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
		Order("dev_environments.created_at DESC").Offset(offset).Limit(pageSize).Find(&environments).Error; err != nil {
		return nil, 0, err
	}

	return environments, total, nil
}

func (r *devEnvironmentRepository) Update(env *database.DevEnvironment) error {
	return r.db.Save(env).Error
}

func (r *devEnvironmentRepository) Delete(id uint) error {
	// Use transaction to ensure both operations succeed or fail together
	return r.db.Transaction(func(tx *gorm.DB) error {
		// First, delete all admin relationships for this environment
		if err := tx.Exec("DELETE FROM dev_environment_admins WHERE dev_environment_id = ?", id).Error; err != nil {
			return err
		}

		// Then delete the environment itself
		return tx.Where("id = ?", id).Delete(&database.DevEnvironment{}).Error
	})
}

// AddAdmin adds an admin to the environment's admin list
func (r *devEnvironmentRepository) AddAdmin(envID, adminID uint) error {
	// Check if the relationship already exists
	var count int64
	err := r.db.Table("dev_environment_admins").
		Where("dev_environment_id = ? AND admin_id = ?", envID, adminID).
		Count(&count).Error
	if err != nil {
		return err
	}

	// If relationship doesn't exist, create it
	if count == 0 {
		return r.db.Exec("INSERT INTO dev_environment_admins (dev_environment_id, admin_id) VALUES (?, ?)", envID, adminID).Error
	}

	return nil
}

// RemoveAdmin removes an admin from the environment's admin list
func (r *devEnvironmentRepository) RemoveAdmin(envID, adminID uint) error {
	// Use direct SQL to delete from the many-to-many relationship table
	return r.db.Exec("DELETE FROM dev_environment_admins WHERE dev_environment_id = ? AND admin_id = ?", envID, adminID).Error
}

// GetAdmins retrieves all admins for a specific environment
func (r *devEnvironmentRepository) GetAdmins(envID uint) ([]database.Admin, error) {
	var admins []database.Admin
	err := r.db.Table("admins").
		Preload("Avatar", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, uuid, original_name")
		}).
		Joins("JOIN dev_environment_admins dea ON admins.id = dea.admin_id").
		Where("dea.dev_environment_id = ?", envID).
		Find(&admins).Error

	return admins, err
}

// IsOwner checks if an admin has access to a specific environment
func (r *devEnvironmentRepository) IsOwner(envID, adminID uint) (bool, error) {
	var count int64

	err := r.db.Table("dev_environments").
		Where("dev_environments.id = ? AND dev_environments.admin_id = ?", envID, adminID).
		Count(&count).Error

	return count > 0, err
}

// CountByAdminID counts the number of dev environments created by a specific admin
func (r *devEnvironmentRepository) CountByAdminID(adminID uint) (int64, error) {
	var count int64
	err := r.db.Model(&database.DevEnvironment{}).
		Where("admin_id = ?", adminID).
		Count(&count).Error
	return count, err
}
