package migrations

import (
	"fmt"
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
	applied, err := checkMigrationStatus(db, migrationName)
	if err != nil {
		return err
	}
	if applied {
		return nil
	}

	utils.Info("Starting admin role migration", "migration", migrationName)

	result := db.Exec("UPDATE admins SET role = 'super_admin' WHERE role IS NULL OR role = '' OR role = 'admin'")
	if result.Error != nil {
		return fmt.Errorf("failed to update existing admin roles: %v", result.Error)
	}

	utils.Info("Updated existing admin users to super_admin role",
		"migration", migrationName,
		"updated_count", result.RowsAffected)

	// Record migration as applied
	return recordMigrationApplied(db, migrationName)
}
