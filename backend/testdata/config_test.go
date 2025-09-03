package testdata

import (
	"xsha-backend/config"
	"xsha-backend/database"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GetTestConfig returns a test configuration
func GetTestConfig() *config.Config {
	return &config.Config{
		Port:        "8080",
		DatabaseURL: ":memory:",
		JWTSecret:   "test-jwt-secret-key-for-testing-only",
		LogLevel:    "error", // Reduce log noise during testing
	}
}

// SetupTestDB creates an in-memory SQLite database for testing
func SetupTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Silent mode for testing
	})
	if err != nil {
		return nil, err
	}

	// Run migrations
	err = database.RunMigrations(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// TeardownTestDB closes the test database connection
func TeardownTestDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}