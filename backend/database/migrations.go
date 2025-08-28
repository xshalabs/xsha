package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
	"xsha-backend/utils"
)

// runWorkspaceRelativePathsMigration converts absolute workspace paths to relative paths
func runWorkspaceRelativePathsMigration(db *gorm.DB, workspaceBaseDir string) error {
	migrationName := "001_workspace_relative_paths"

	// Check if migration already applied
	var existing Migration
	if err := db.Where("name = ?", migrationName).First(&existing).Error; err == nil {
		utils.Info("Migration already applied, skipping", "migration", migrationName)
		return nil
	} else if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check migration status: %v", err)
	}

	utils.Info("Starting workspace relative paths migration", "migration", migrationName)

	// Get all tasks with non-empty workspace paths
	var tasks []Task
	if err := db.Where("workspace_path != '' AND workspace_path IS NOT NULL").Find(&tasks).Error; err != nil {
		return fmt.Errorf("failed to fetch tasks: %v", err)
	}

	utils.Info("Found tasks to migrate", "count", len(tasks))

	// Process each task
	migratedCount := 0
	errorCount := 0

	for _, task := range tasks {
		if err := migrateTaskWorkspacePath(db, &task, workspaceBaseDir); err != nil {
			utils.Error("Failed to migrate task", "taskId", task.ID, "error", err)
			errorCount++
			continue
		}
		migratedCount++
	}

	utils.Info("Migration completed",
		"migration", migrationName,
		"migrated", migratedCount,
		"errors", errorCount,
		"total", len(tasks))

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

func migrateTaskWorkspacePath(db *gorm.DB, task *Task, workspaceBaseDir string) error {
	originalPath := task.WorkspacePath

	// Skip if already looks like a relative path (no leading slash)
	if !strings.HasPrefix(originalPath, "/") {
		utils.Debug("Task already has relative path", "taskId", task.ID, "path", originalPath)
		return nil
	}

	// Extract relative path from absolute path
	relativePath := utils.ExtractWorkspaceRelativePath(originalPath)
	if relativePath == "" {
		return fmt.Errorf("could not extract relative path from: %s", originalPath)
	}

	utils.Debug("Converting workspace path",
		"taskId", task.ID,
		"from", originalPath,
		"to", relativePath)

	// Check if old directory exists and move it if needed
	if _, err := os.Stat(originalPath); err == nil {
		// Old directory exists, we might need to move it
		newAbsolutePath := filepath.Join(workspaceBaseDir, relativePath)

		// Only move if the new location doesn't exist
		if _, err := os.Stat(newAbsolutePath); os.IsNotExist(err) {
			// Ensure parent directory exists
			if err := os.MkdirAll(filepath.Dir(newAbsolutePath), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory: %v", err)
			}

			// Move the directory
			if err := os.Rename(originalPath, newAbsolutePath); err != nil {
				utils.Warn("Failed to move workspace directory",
					"taskId", task.ID,
					"from", originalPath,
					"to", newAbsolutePath,
					"error", err)
				// Continue with database update even if move fails
			} else {
				utils.Debug("Moved workspace directory",
					"taskId", task.ID,
					"from", originalPath,
					"to", newAbsolutePath)
			}
		} else if err != nil {
			// Error checking new location
			utils.Warn("Error checking new workspace location",
				"taskId", task.ID,
				"path", newAbsolutePath,
				"error", err)
		} else {
			// New location already exists, keep old one for safety
			utils.Debug("New workspace location already exists",
				"taskId", task.ID,
				"path", newAbsolutePath)
		}
	}

	// Update database with relative path
	if err := db.Model(task).Update("workspace_path", relativePath).Error; err != nil {
		return fmt.Errorf("failed to update task workspace path: %v", err)
	}

	utils.Debug("Successfully migrated task workspace path",
		"taskId", task.ID,
		"oldPath", originalPath,
		"newPath", relativePath)

	return nil
}

// runDevEnvironmentSessionDirMigration converts absolute session directories to relative paths
func runDevEnvironmentSessionDirMigration(db *gorm.DB, devSessionsDir string) error {
	migrationName := "002_dev_environment_session_dir_relative_paths"

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
		if err := migrateDevEnvironmentSessionDir(db, &devEnv, devSessionsDir); err != nil {
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

func migrateDevEnvironmentSessionDir(db *gorm.DB, devEnv *DevEnvironment, devSessionsDir string) error {
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

// runRemoveAdminCredentialsSystemConfigMigration removes admin_user and admin_password system configs
func runRemoveAdminCredentialsSystemConfigMigration(db *gorm.DB) error {
	migrationName := "003_remove_admin_credentials_system_config"

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

// runTaskConversationAttachmentAdminIDMigration adds admin_id column to task_conversation_attachments table
func runTaskConversationAttachmentAdminIDMigration(db *gorm.DB) error {
	migrationName := "004_task_conversation_attachment_admin_id"

	// Check if migration already applied
	var existing Migration
	if err := db.Where("name = ?", migrationName).First(&existing).Error; err == nil {
		utils.Info("Migration already applied, skipping", "migration", migrationName)
		return nil
	} else if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check migration status: %v", err)
	}

	utils.Info("Starting task conversation attachment admin_id migration", "migration", migrationName)

	// Add admin_id column to task_conversation_attachments table
	if err := db.Exec("ALTER TABLE task_conversation_attachments ADD COLUMN admin_id INTEGER").Error; err != nil {
		// Check if column already exists
		if !strings.Contains(err.Error(), "duplicate column") && !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to add admin_id column: %v", err)
		}
		utils.Info("admin_id column already exists")
	}

	// Add index for admin_id column
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_task_conversation_attachments_admin_id ON task_conversation_attachments(admin_id)").Error; err != nil {
		utils.Warn("Failed to create index on admin_id", "error", err)
	}

	// Populate admin_id for existing records
	var attachments []TaskConversationAttachment
	if err := db.Find(&attachments).Error; err != nil {
		return fmt.Errorf("failed to fetch attachments: %v", err)
	}

	utils.Info("Found attachments to migrate", "count", len(attachments))

	migratedCount := 0
	errorCount := 0

	for _, attachment := range attachments {
		if attachment.AdminID != nil {
			// Already has admin_id, skip
			continue
		}

		// Find admin by username
		var admin Admin
		if err := db.Where("username = ?", attachment.CreatedBy).First(&admin).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				utils.Warn("Admin not found for attachment",
					"attachmentId", attachment.ID,
					"createdBy", attachment.CreatedBy)
				errorCount++
				continue
			}
			utils.Error("Failed to find admin for attachment",
				"attachmentId", attachment.ID,
				"createdBy", attachment.CreatedBy,
				"error", err)
			errorCount++
			continue
		}

		// Update attachment with admin_id
		if err := db.Model(&attachment).Update("admin_id", admin.ID).Error; err != nil {
			utils.Error("Failed to update attachment admin_id",
				"attachmentId", attachment.ID,
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
		"total", len(attachments))

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

// runAdminOperationLogAdminIDMigration adds admin_id column to admin_operation_logs table
func runAdminOperationLogAdminIDMigration(db *gorm.DB) error {
	migrationName := "005_admin_operation_log_admin_id"

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
