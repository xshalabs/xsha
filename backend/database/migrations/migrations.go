package migrations

import (
	"fmt"
	"xsha-backend/config"
	"xsha-backend/utils"

	"gorm.io/gorm"
)

// MigrationInterface represents a database migration
type MigrationInterface interface {
	Name() string
	Run(db *gorm.DB, cfg *config.Config) error
}

// MigrationManager manages and runs database migrations
type MigrationManager struct {
	db          *gorm.DB
	cfg         *config.Config
	migrations  []MigrationInterface
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(db *gorm.DB, cfg *config.Config) *MigrationManager {
	manager := &MigrationManager{
		db:         db,
		cfg:        cfg,
		migrations: make([]MigrationInterface, 0),
	}
	
	// Register all migrations in order
	manager.registerMigrations()
	return manager
}

// registerMigrations registers all available migrations
func (m *MigrationManager) registerMigrations() {
	m.migrations = []MigrationInterface{
		&WorkspaceRelativePathsMigration{},
		&DevEnvironmentSessionDirMigration{},
		&RemoveAdminCredentialsMigration{},
		&TaskConversationAttachmentAdminIDMigration{},
		&AdminOperationLogAdminIDMigration{},
		&DevEnvironmentAdminIDMigration{},
	}
}

// RunAll runs all registered migrations
func (m *MigrationManager) RunAll() error {
	utils.Info("Starting migration manager")
	
	for _, migration := range m.migrations {
		utils.Info("Running migration", "name", migration.Name())
		if err := migration.Run(m.db, m.cfg); err != nil {
			return fmt.Errorf("migration %s failed: %v", migration.Name(), err)
		}
	}
	
	utils.Info("All migrations completed successfully")
	return nil
}