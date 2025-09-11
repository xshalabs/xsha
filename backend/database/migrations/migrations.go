package migrations

import (
	"fmt"
	"time"
	"xsha-backend/config"

	"gorm.io/gorm"
)

// MigrationInterface represents a database migration
type MigrationInterface interface {
	Name() string
	Run(db *gorm.DB, cfg *config.Config) error
}

// MigrationManager manages and runs database migrations
type MigrationManager struct {
	db         *gorm.DB
	cfg        *config.Config
	migrations []MigrationInterface
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
		&AddAdminRoleMigration{},
	}
}

// RunAll runs all registered migrations
func (m *MigrationManager) RunAll() error {
	for _, migration := range m.migrations {
		if err := migration.Run(m.db, m.cfg); err != nil {
			return fmt.Errorf("migration %s failed: %v", migration.Name(), err)
		}
	}
	return nil
}

// checkMigrationStatus checks if a migration has already been applied
// Returns true if already applied, false if not applied, error if database error
func checkMigrationStatus(db *gorm.DB, migrationName string) (bool, error) {
	var existing Migration
	if err := db.Where("name = ?", migrationName).First(&existing).Error; err == nil {
		return true, nil
	} else if err != gorm.ErrRecordNotFound {
		return false, fmt.Errorf("failed to check migration status: %v", err)
	}
	return false, nil
}

// recordMigrationApplied records a migration as successfully applied
func recordMigrationApplied(db *gorm.DB, migrationName string) error {
	migration := Migration{
		Name:      migrationName,
		AppliedAt: time.Now(),
	}
	if err := db.Create(&migration).Error; err != nil {
		return fmt.Errorf("failed to record migration: %v", err)
	}
	return nil
}
