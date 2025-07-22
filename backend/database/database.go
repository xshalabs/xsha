package database

import (
	"log"
	"sleep0-backend/config"

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
			log.Fatal("MySQL DSN not configured, please set SLEEP0_MYSQL_DSN environment variable")
		}
		db, err = gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{})
		if err != nil {
			return nil, err
		}
		log.Println("MySQL database connected successfully")
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(cfg.SQLitePath), &gorm.Config{})
		if err != nil {
			return nil, err
		}
		log.Println("SQLite database connected successfully")
	default:
		log.Fatalf("Unsupported database type: %s", cfg.DatabaseType)
	}

	// Auto-migrate database tables
	if err := db.AutoMigrate(&TokenBlacklist{}, &LoginLog{}, &GitCredential{}, &Project{}, &AdminOperationLog{}, &DevEnvironment{}, &Task{}, &TaskConversation{}); err != nil {
		return nil, err
	}
	log.Println("Database table migration completed")

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
		log.Fatalf("Failed to initialize database: %v", err)
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
