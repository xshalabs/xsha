package database

import (
	"fmt"
	"time"
	"xsha-backend/utils"

	"gorm.io/gorm"
)

// runTokenBlacklistTokenIDMigration migrates TokenBlacklist table to use TokenID instead of Token
func runTokenBlacklistTokenIDMigration(db *gorm.DB) error {
	migrationName := "004_token_blacklist_token_id"

	// Check if migration already applied
	var existing Migration
	if err := db.Where("name = ?", migrationName).First(&existing).Error; err == nil {
		utils.Info("Migration already applied, skipping", "migration", migrationName)
		return nil
	} else if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check migration status: %v", err)
	}

	utils.Info("Starting token blacklist token_id migration", "migration", migrationName)

	// Check if the table exists and has the old structure
	if db.Migrator().HasTable(&TokenBlacklist{}) {
		// Check if the token column exists (old structure)
		if db.Migrator().HasColumn(&TokenBlacklist{}, "token") {
			utils.Info("Found token_blacklists table with old structure, recreating...")

			// Clear all existing entries since they can't be converted
			result := db.Delete(&TokenBlacklist{}, "1 = 1")
			if result.Error != nil {
				utils.Warn("Failed to clear old token blacklist entries", "error", result.Error)
			} else {
				utils.Info("Cleared old token blacklist entries", "count", result.RowsAffected)
			}

			// Drop the old table
			if err := db.Migrator().DropTable(&TokenBlacklist{}); err != nil {
				return fmt.Errorf("failed to drop old token_blacklists table: %v", err)
			}
		} else if db.Migrator().HasColumn(&TokenBlacklist{}, "token_id") {
			// Already has new structure
			utils.Info("token_blacklists table already has new structure")
			return nil
		}
	}

	// Create the table with new structure
	if err := db.Migrator().CreateTable(&TokenBlacklist{}); err != nil {
		return fmt.Errorf("failed to create token_blacklists table with new structure: %v", err)
	}

	utils.Info("Successfully created token_blacklists table with token_id field")

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
