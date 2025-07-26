package database

import (
	"fmt"
	"sleep0-backend/config"
	"sleep0-backend/utils"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DatabaseManager 数据库管理器
type DatabaseManager struct {
	db *gorm.DB
}

// NewDatabaseManager 创建数据库管理器实例
func NewDatabaseManager(cfg *config.Config) (*DatabaseManager, error) {
	var db *gorm.DB
	var err error

	switch cfg.DatabaseType {
	case "mysql":
		if cfg.MySQLDSN == "" {
			utils.Error("MySQL DSN not configured")
			panic("MySQL DSN not configured, please set SLEEP0_MYSQL_DSN environment variable")
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

	// Auto-migrate database tables
	if err := db.AutoMigrate(&TokenBlacklist{}, &LoginLog{}, &GitCredential{}, &Project{}, &AdminOperationLog{}, &DevEnvironment{}, &Task{}, &TaskConversation{}, &TaskExecutionLog{}); err != nil {
		return nil, err
	}
	utils.Info("Database table migration completed")

	return &DatabaseManager{db: db}, nil
}

// GetDB 获取数据库连接
func (dm *DatabaseManager) GetDB() *gorm.DB {
	return dm.db
}

// Close 关闭数据库连接
func (dm *DatabaseManager) Close() error {
	sqlDB, err := dm.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// 保持向后兼容的全局变量和函数
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

// Maintain backward compatibility
func InitSQLite() {
	InitDatabase()
}

func GetDB() *gorm.DB {
	return DB
}
