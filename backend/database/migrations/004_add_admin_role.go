package migrations

import (
	"fmt"
	"time"
	"xsha-backend/config"
	"xsha-backend/utils"

	"gorm.io/gorm"
)

// AddAdminRoleMigration adds role field to Admin table and sets existing admins as super_admin
type AddAdminRoleMigration struct{}

func (m *AddAdminRoleMigration) Name() string {
	return "004_add_admin_role"
}

func (m *AddAdminRoleMigration) Run(db *gorm.DB, cfg *config.Config) error {
	migrationName := m.Name()

	// Check if migration already applied
	var existing Migration
	if err := db.Where("name = ?", migrationName).First(&existing).Error; err == nil {
		utils.Info("Migration already applied, skipping", "migration", migrationName)
		return nil
	} else if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check migration status: %v", err)
	}

	utils.Info("Starting admin role migration", "migration", migrationName)

	// Add role column to admins table using raw SQL to avoid import cycle
	if err := db.Exec("ALTER TABLE admins ADD COLUMN IF NOT EXISTS role VARCHAR(50) DEFAULT 'admin'").Error; err != nil {
		// Try without IF NOT EXISTS for databases that don't support it
		if err := db.Exec("ALTER TABLE admins ADD COLUMN role VARCHAR(50) DEFAULT 'admin'").Error; err != nil {
			// Column might already exist, check if we can continue
			utils.Info("Failed to add role column, it might already exist", "error", err.Error())
		}
	}

	// Add index on role column
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_admins_role ON admins(role)").Error; err != nil {
		// Try without IF NOT EXISTS
		if err := db.Exec("CREATE INDEX idx_admins_role ON admins(role)").Error; err != nil {
			utils.Info("Failed to create role index, it might already exist", "error", err.Error())
		}
	}

	// Set all existing admins to super_admin role (for backward compatibility)
	result := db.Exec("UPDATE admins SET role = 'super_admin' WHERE role IS NULL OR role = '' OR role = 'admin'")
	if result.Error != nil {
		return fmt.Errorf("failed to update existing admin roles: %v", result.Error)
	}

	utils.Info("Updated existing admin users to super_admin role",
		"migration", migrationName,
		"updated_count", result.RowsAffected)

	// Record migration as applied
	migration := Migration{
		Name:      migrationName,
		AppliedAt: time.Now(),
	}
	if err := db.Create(&migration).Error; err != nil {
		return fmt.Errorf("failed to record migration: %v", err)
	}

	utils.Info("Admin role migration completed successfully", "migration", migrationName)
	return nil
}