package migrations

import (
	"fmt"
	"time"
	"xsha-backend/config"
	"xsha-backend/utils"

	"gorm.io/gorm"
)

// RemoveAdminCredentialsMigration removes admin_user and admin_password system configs
type RemoveAdminCredentialsMigration struct{}

func (m *RemoveAdminCredentialsMigration) Name() string {
	return "003_remove_admin_credentials_system_config"
}

func (m *RemoveAdminCredentialsMigration) Run(db *gorm.DB, cfg *config.Config) error {
	migrationName := m.Name()

	// Check if migration already applied
	var existing Migration
	if err := db.Where("name = ?", migrationName).First(&existing).Error; err == nil {
		utils.Info("Migration already applied, skipping", "migration", migrationName)
		return nil
	} else if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check migration status: %v", err)
	}

	utils.Info("Starting admin credentials system config removal migration", "migration", migrationName)

	// Remove admin_user and admin_password system configs
	result := db.Where("config_key IN (?)", []string{"admin_user", "admin_password"}).Delete(&SystemConfig{})
	if result.Error != nil {
		return fmt.Errorf("failed to remove admin credentials system configs: %v", result.Error)
	}

	utils.Info("Removed admin credentials system configs",
		"migration", migrationName,
		"deleted_count", result.RowsAffected)

	// Record migration as applied
	migration := Migration{
		Name:      migrationName,
		AppliedAt: time.Now(),
	}
	if err := db.Create(&migration).Error; err != nil {
		return fmt.Errorf("failed to record migration: %v", err)
	}

	return nil
}