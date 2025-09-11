package migrations

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"xsha-backend/config"
	"xsha-backend/utils"

	"gorm.io/gorm"
)

// WorkspaceRelativePathsMigration converts absolute workspace paths to relative paths
type WorkspaceRelativePathsMigration struct{}

func (m *WorkspaceRelativePathsMigration) Name() string {
	return "001_workspace_relative_paths"
}

func (m *WorkspaceRelativePathsMigration) Run(db *gorm.DB, cfg *config.Config) error {
	migrationName := m.Name()

	// Check if migration already applied
	applied, err := checkMigrationStatus(db, migrationName)
	if err != nil {
		return err
	}
	if applied {
		return nil
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
		if err := m.migrateTaskWorkspacePath(db, &task, cfg.WorkspaceBaseDir); err != nil {
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
	return recordMigrationApplied(db, migrationName)
}

func (m *WorkspaceRelativePathsMigration) migrateTaskWorkspacePath(db *gorm.DB, task *Task, workspaceBaseDir string) error {
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
