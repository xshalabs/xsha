package migrations

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"xsha-backend/config"
	"xsha-backend/utils"

	"gorm.io/gorm"
)

// DevEnvironmentSessionDirMigration converts absolute session directories to relative paths
type DevEnvironmentSessionDirMigration struct{}

func (m *DevEnvironmentSessionDirMigration) Name() string {
	return "002_dev_environment_session_dir_relative_paths"
}

func (m *DevEnvironmentSessionDirMigration) Run(db *gorm.DB, cfg *config.Config) error {
	migrationName := m.Name()

	// Check if migration already applied
	var existing Migration
	if err := db.Where("name = ?", migrationName).First(&existing).Error; err == nil {
		utils.Info("Migration already applied, skipping", "migration", migrationName)
		return nil
	} else if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check migration status: %v", err)
	}

	utils.Info("Starting dev environment session dir relative paths migration", "migration", migrationName)

	// Get all dev environments with non-empty session directories
	var devEnvs []DevEnvironment
	if err := db.Where("session_dir != '' AND session_dir IS NOT NULL").Find(&devEnvs).Error; err != nil {
		return fmt.Errorf("failed to fetch dev environments: %v", err)
	}

	utils.Info("Found dev environments to migrate", "count", len(devEnvs))

	// Process each dev environment
	migratedCount := 0
	errorCount := 0

	for _, devEnv := range devEnvs {
		if err := m.migrateDevEnvironmentSessionDir(db, &devEnv, cfg.DevSessionsDir); err != nil {
			utils.Error("Failed to migrate dev environment", "devEnvId", devEnv.ID, "error", err)
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

func (m *DevEnvironmentSessionDirMigration) migrateDevEnvironmentSessionDir(db *gorm.DB, devEnv *DevEnvironment, devSessionsDir string) error {
	originalPath := devEnv.SessionDir

	// Skip if already looks like a relative path (no leading slash)
	if !strings.HasPrefix(originalPath, "/") {
		utils.Debug("DevEnvironment already has relative session dir", "devEnvId", devEnv.ID, "path", originalPath)
		return nil
	}

	// Extract relative path from absolute path
	relativePath := utils.ExtractDevSessionRelativePath(originalPath)
	if relativePath == "" {
		return fmt.Errorf("could not extract relative path from: %s", originalPath)
	}

	utils.Debug("Converting session dir",
		"devEnvId", devEnv.ID,
		"from", originalPath,
		"to", relativePath)

	// Check if old directory exists and move it if needed
	if _, err := os.Stat(originalPath); err == nil {
		// Old directory exists, we might need to move it
		newAbsolutePath := filepath.Join(devSessionsDir, relativePath)

		// Only move if the new location doesn't exist
		if _, err := os.Stat(newAbsolutePath); os.IsNotExist(err) {
			// Ensure parent directory exists
			if err := os.MkdirAll(filepath.Dir(newAbsolutePath), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory: %v", err)
			}

			// Move the directory
			if err := os.Rename(originalPath, newAbsolutePath); err != nil {
				utils.Warn("Failed to move session directory",
					"devEnvId", devEnv.ID,
					"from", originalPath,
					"to", newAbsolutePath,
					"error", err)
				// Continue with database update even if move fails
			} else {
				utils.Debug("Moved session directory",
					"devEnvId", devEnv.ID,
					"from", originalPath,
					"to", newAbsolutePath)
			}
		} else if err != nil {
			// Error checking new location
			utils.Warn("Error checking new session directory location",
				"devEnvId", devEnv.ID,
				"path", newAbsolutePath,
				"error", err)
		} else {
			// New location already exists, keep old one for safety
			utils.Debug("New session directory location already exists",
				"devEnvId", devEnv.ID,
				"path", newAbsolutePath)
		}
	}

	// Update database with relative path
	if err := db.Model(devEnv).Update("session_dir", relativePath).Error; err != nil {
		return fmt.Errorf("failed to update dev environment session dir: %v", err)
	}

	utils.Debug("Successfully migrated dev environment session dir",
		"devEnvId", devEnv.ID,
		"oldPath", originalPath,
		"newPath", relativePath)

	return nil
}
