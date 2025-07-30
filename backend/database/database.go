package database

import (
	"fmt"
	"xsha-backend/config"
	"xsha-backend/utils"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DatabaseManager struct {
	db *gorm.DB
}

func NewDatabaseManager(cfg *config.Config) (*DatabaseManager, error) {
	var db *gorm.DB
	var err error

	switch cfg.DatabaseType {
	case "mysql":
		if cfg.MySQLDSN == "" {
			utils.Error("MySQL DSN not configured")
			panic("MySQL DSN not configured, please set XSHA_MYSQL_DSN environment variable")
		}
		db, err = gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{})
		if err != nil {
			return nil, err
		}
		utils.Info("MySQL database connected successfully")
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(cfg.SQLitePath), &gorm.Config{})
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

	if err := db.AutoMigrate(&TokenBlacklist{}, &LoginLog{}, &GitCredential{}, &Project{}, &AdminOperationLog{}, &DevEnvironment{}, &Task{}, &TaskConversation{}, &TaskExecutionLog{}, &TaskConversationResult{}, &SystemConfig{}); err != nil {
		return nil, err
	}
	utils.Info("Database table migration completed")

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
