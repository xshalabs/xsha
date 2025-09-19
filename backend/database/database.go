package database

import (
	"fmt"
	"strings"
	"time"
	"xsha-backend/config"
	"xsha-backend/database/migrations"
	"xsha-backend/utils"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DatabaseManager struct {
	db *gorm.DB
}

func NewDatabaseManager(cfg *config.Config) (*DatabaseManager, error) {
	var db *gorm.DB
	var err error

	// Configure GORM to use UTC timezone
	gormConfig := &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	switch cfg.DatabaseType {
	case "mysql":
		if cfg.MySQLDSN == "" {
			utils.Error("MySQL DSN not configured")
			panic("MySQL DSN not configured, please set XSHA_MYSQL_DSN environment variable")
		}
		// Ensure MySQL connection uses UTC timezone
		dsn := cfg.MySQLDSN
		if !containsTimeZone(dsn) {
			if containsParams(dsn) {
				dsn += "&time_zone=%27%2B00%3A00%27"
			} else {
				dsn += "?time_zone=%27%2B00%3A00%27"
			}
		}
		db, err = gorm.Open(mysql.Open(dsn), gormConfig)
		if err != nil {
			return nil, err
		}

		// Execute SQL to set session timezone to UTC
		if err := db.Exec("SET time_zone = '+00:00'").Error; err != nil {
			utils.Warn("Failed to set MySQL session timezone to UTC", "error", err)
		}

		utils.Info("MySQL database connected successfully with UTC timezone")
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(cfg.SQLitePath), gormConfig)
		if err != nil {
			return nil, err
		}
		utils.Info("SQLite database connected successfully")
	default:
		utils.Error("Unsupported database type",
			"type", cfg.DatabaseType,
		)
		panic(fmt.Sprintf("Unsupported database type: %s", cfg.DatabaseType))
	}

	// AutoMigrate all tables first to create the base structure
	if err := db.AutoMigrate(&Migration{}, &TokenBlacklistV2{}, &LoginLog{}, &Admin{}, &GitCredential{}, &Project{}, &AdminOperationLog{}, &DevEnvironment{}, &Task{}, &TaskConversation{}, &TaskExecutionLog{}, &TaskConversationResult{}, &TaskConversationAttachment{}, &SystemConfig{}, &AdminAvatar{}, &Notifier{}); err != nil {
		return nil, err
	}

	// Run custom migrations after AutoMigrate (for data migrations and custom changes)
	if err := runMigrations(db, cfg); err != nil {
		utils.Error("Failed to run custom migrations", "error", err)
		return nil, err
	}

	return &DatabaseManager{db: db}, nil
}

func (dm *DatabaseManager) GetDB() *gorm.DB {
	return dm.db
}

func (dm *DatabaseManager) Close() error {
	sqlDB, err := dm.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

var (
	DB *gorm.DB
)

func InitDatabase() {
	cfg := config.Load()
	dm, err := NewDatabaseManager(cfg)
	if err != nil {
		utils.Error("Failed to initialize database",
			"error", err.Error(),
		)
		panic(fmt.Sprintf("Failed to initialize database: %v", err))
	}
	DB = dm.GetDB()
}

func InitSQLite() {
	InitDatabase()
}

func GetDB() *gorm.DB {
	return DB
}

// runMigrations executes custom migrations
func runMigrations(db *gorm.DB, cfg *config.Config) error {
	migrationManager := migrations.NewMigrationManager(db, cfg)
	if err := migrationManager.RunAll(); err != nil {
		return fmt.Errorf("migration manager failed: %v", err)
	}
	return nil
}

// containsTimeZone checks if the DSN already contains timezone information
func containsTimeZone(dsn string) bool {
	return strings.Contains(strings.ToLower(dsn), "time_zone")
}

// containsParams checks if the DSN already contains query parameters
func containsParams(dsn string) bool {
	return strings.Contains(dsn, "?")
}
