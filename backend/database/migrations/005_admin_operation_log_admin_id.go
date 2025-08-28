package migrations

import (
	"fmt"
	"strings"
	"time"
	"xsha-backend/config"
	"xsha-backend/utils"

	"gorm.io/gorm"
)

// AdminOperationLogAdminIDMigration adds admin_id column to admin_operation_logs table
type AdminOperationLogAdminIDMigration struct{}

func (m *AdminOperationLogAdminIDMigration) Name() string {
	return "005_admin_operation_log_admin_id"
}

func (m *AdminOperationLogAdminIDMigration) Run(db *gorm.DB, cfg *config.Config) error {
	migrationName := m.Name()

	// Check if migration already applied
	var existing Migration
	if err := db.Where("name = ?", migrationName).First(&existing).Error; err == nil {
		utils.Info("Migration already applied, skipping", "migration", migrationName)
		return nil
	} else if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check migration status: %v", err)
	}

	utils.Info("Starting admin operation log admin_id migration", "migration", migrationName)

	// Add admin_id column to admin_operation_logs table
	if err := db.Exec("ALTER TABLE admin_operation_logs ADD COLUMN admin_id INTEGER").Error; err != nil {
		// Check if column already exists
		if !strings.Contains(err.Error(), "duplicate column") && !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to add admin_id column: %v", err)
		}
		utils.Info("admin_id column already exists")
	}

	// Add index for admin_id column
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_admin_operation_logs_admin_id ON admin_operation_logs(admin_id)").Error; err != nil {
		utils.Warn("Failed to create index on admin_id", "error", err)
	}

	// Populate admin_id for existing records
	var logs []AdminOperationLog
	if err := db.Find(&logs).Error; err != nil {
		return fmt.Errorf("failed to fetch operation logs: %v", err)
	}

	utils.Info("Found operation logs to migrate", "count", len(logs))

	migratedCount := 0
	errorCount := 0

	for _, log := range logs {
		if log.AdminID != nil {
			// Already has admin_id, skip
			continue
		}

		// Find admin by username
		var admin Admin
		if err := db.Where("username = ?", log.Username).First(&admin).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				utils.Warn("Admin not found for operation log",
					"logId", log.ID,
					"username", log.Username)
				errorCount++
				continue
			}
			utils.Error("Failed to find admin for operation log",
				"logId", log.ID,
				"username", log.Username,
				"error", err)
			errorCount++
			continue
		}

		// Update log with admin_id
		if err := db.Model(&log).Update("admin_id", admin.ID).Error; err != nil {
			utils.Error("Failed to update operation log admin_id",
				"logId", log.ID,
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
		"total", len(logs))

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