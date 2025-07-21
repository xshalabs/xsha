package database

import (
	"log"
	"sleep0-backend/config"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	DB *gorm.DB
)

func InitDatabase() {
	cfg := config.Load()

	var err error
	switch cfg.DatabaseType {
	case "mysql":
		if cfg.MySQLDSN == "" {
			log.Fatal("MySQL DSN 未配置，请设置 SLEEP0_MYSQL_DSN 环境变量")
		}
		DB, err = gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{})
		if err != nil {
			log.Fatalf("连接 MySQL 数据库失败: %v", err)
		}
		log.Println("MySQL 数据库连接成功")
	case "sqlite":
		DB, err = gorm.Open(sqlite.Open(cfg.SQLitePath), &gorm.Config{})
		if err != nil {
			log.Fatalf("连接 SQLite 数据库失败: %v", err)
		}
		log.Println("SQLite 数据库连接成功")
	default:
		log.Fatalf("不支持的数据库类型: %s", cfg.DatabaseType)
	}
}

// 保留向后兼容性
func InitSQLite() {
	InitDatabase()
}

func GetDB() *gorm.DB {
	return DB
}
