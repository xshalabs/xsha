package migrations

import (
	"fmt"
	"strings"
	"time"
	"xsha-backend/config"
	"xsha-backend/utils"

	"gorm.io/gorm"
)

// DevEnvironmentAdminIDMigration adds admin_id column to dev_environments table
type DevEnvironmentAdminIDMigration struct{}

func (m *DevEnvironmentAdminIDMigration) Name() string {
	return "006_dev_environment_admin_id"
}

func (m *DevEnvironmentAdminIDMigration) Run(db *gorm.DB, cfg *config.Config) error {
	migrationName := m.Name()

	// Check if migration already applied
	var existing Migration
	if err := db.Where("name = ?", migrationName).First(&existing).Error; err == nil {
		utils.Info("Migration already applied, skipping", "migration", migrationName)
		return nil
	} else if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check migration status: %v", err)
	}

	utils.Info("Starting dev environment admin_id migration", "migration", migrationName)

	// Add admin_id column to dev_environments table
	if err := db.Exec("ALTER TABLE dev_environments ADD COLUMN admin_id INTEGER").Error; err != nil {
		// Check if column already exists
		if !strings.Contains(err.Error(), "duplicate column") && !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to add admin_id column: %v", err)
		}
		utils.Info("admin_id column already exists")
	}

	// Add index for admin_id column
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_dev_environments_admin_id ON dev_environments(admin_id)").Error; err != nil {
		utils.Warn("Failed to create index on admin_id", "error", err)
	}

	// Populate admin_id for existing records
	var devEnvs []DevEnvironment
	if err := db.Find(&devEnvs).Error; err != nil {
		return fmt.Errorf("failed to fetch dev environments: %v", err)
	}

	utils.Info("Found dev environments to migrate", "count", len(devEnvs))

	migratedCount := 0
	errorCount := 0

	for _, devEnv := range devEnvs {
		if devEnv.AdminID != nil {
			// Already has admin_id, skip
			continue
		}

		// Find admin by username
		var admin Admin
		if err := db.Where("username = ?", devEnv.CreatedBy).First(&admin).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				utils.Warn("Admin not found for dev environment",
					"devEnvId", devEnv.ID,
					"createdBy", devEnv.CreatedBy)
				errorCount++
				continue
			}
			utils.Error("Failed to find admin for dev environment",
				"devEnvId", devEnv.ID,
				"createdBy", devEnv.CreatedBy,
				"error", err)
			errorCount++
			continue
		}

		// Update dev environment with admin_id
		if err := db.Model(&devEnv).Update("admin_id", admin.ID).Error; err != nil {
			utils.Error("Failed to update dev environment admin_id",
				"devEnvId", devEnv.ID,
				"adminId", admin.ID,
				"error", err)
			errorCount++
			continue
		}

		migratedCount++
	}

	utils.Info("Migration completed",
		"migration", migrationName,
		"migrated", migratedCount,
		"errors", errorCount,
		"total", len(devEnvs))

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