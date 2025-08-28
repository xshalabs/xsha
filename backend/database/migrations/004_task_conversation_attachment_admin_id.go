package migrations

import (
	"fmt"
	"strings"
	"time"
	"xsha-backend/config"
	"xsha-backend/utils"

	"gorm.io/gorm"
)

// TaskConversationAttachmentAdminIDMigration adds admin_id column to task_conversation_attachments table
type TaskConversationAttachmentAdminIDMigration struct{}

func (m *TaskConversationAttachmentAdminIDMigration) Name() string {
	return "004_task_conversation_attachment_admin_id"
}

func (m *TaskConversationAttachmentAdminIDMigration) Run(db *gorm.DB, cfg *config.Config) error {
	migrationName := m.Name()

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