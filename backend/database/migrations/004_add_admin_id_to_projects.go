package migrations

import (
	"fmt"
	"time"
	"xsha-backend/config"
	"xsha-backend/utils"

	"gorm.io/gorm"
)

// AddAdminIdToProjectsMigration adds admin_id column to projects table
type AddAdminIdToProjectsMigration struct{}

func (m *AddAdminIdToProjectsMigration) Name() string {
	return "004_add_admin_id_to_projects"
}

func (m *AddAdminIdToProjectsMigration) Run(db *gorm.DB, cfg *config.Config) error {
	migrationName := m.Name()

	// Check if migration already applied
	var existing Migration
	if err := db.Where("name = ?", migrationName).First(&existing).Error; err == nil {
		utils.Info("Migration already applied, skipping", "migration", migrationName)
		return nil
	} else if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check migration status: %v", err)
	}

	utils.Info("Starting add admin_id to projects migration", "migration", migrationName)

	// Add admin_id column to projects table
	if err := db.Exec("ALTER TABLE projects ADD COLUMN admin_id BIGINT UNSIGNED NULL").Error; err != nil {
		// Check if column already exists
		if db.Exec("SELECT admin_id FROM projects LIMIT 1").Error == nil {
			utils.Info("Column admin_id already exists in projects table")
		} else {
			return fmt.Errorf("failed to add admin_id column to projects table: %v", err)
		}
	}

	// Add index on admin_id
	if err := db.Exec("CREATE INDEX idx_projects_admin_id ON projects(admin_id)").Error; err != nil {
		// Index might already exist, log but don't fail
		utils.Warn("Failed to create index on admin_id, may already exist", "error", err)
	}

	// Add foreign key constraint (if using MySQL)
	if cfg.DatabaseType == "mysql" {
		if err := db.Exec("ALTER TABLE projects ADD CONSTRAINT fk_projects_admin FOREIGN KEY (admin_id) REFERENCES admins(id) ON DELETE SET NULL").Error; err != nil {
			// Constraint might already exist, log but don't fail
			utils.Warn("Failed to add foreign key constraint for admin_id, may already exist", "error", err)
		}
	}

	// Update existing projects to set admin_id based on created_by username
	// First, get all existing projects
	var projects []struct {
		ID        uint   `gorm:"column:id"`
		CreatedBy string `gorm:"column:created_by"`
	}
	
	if err := db.Table("projects").Select("id, created_by").Find(&projects).Error; err != nil {
		return fmt.Errorf("failed to fetch existing projects: %v", err)
	}

	// For each project, find the admin with matching username and set admin_id
	for _, project := range projects {
		var admin struct {
			ID uint `gorm:"column:id"`
		}
		
		// Find admin with matching username
		if err := db.Table("admins").Select("id").Where("username = ?", project.CreatedBy).First(&admin).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				utils.Warn("No admin found for project creator", 
					"project_id", project.ID, 
					"created_by", project.CreatedBy)
				continue
			}
			return fmt.Errorf("failed to find admin for project %d: %v", project.ID, err)
		}

		// Update project with admin_id
		if err := db.Exec("UPDATE projects SET admin_id = ? WHERE id = ?", admin.ID, project.ID).Error; err != nil {
			utils.Warn("Failed to update project admin_id", 
				"project_id", project.ID, 
				"admin_id", admin.ID,
				"error", err)
		} else {
			utils.Info("Updated project admin_id", 
				"project_id", project.ID, 
				"admin_id", admin.ID)
		}
	}

	utils.Info("Completed admin_id migration for projects table", "migration", migrationName)

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