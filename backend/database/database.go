package database

import (
	"log"
	"sleep0-backend/config"
	"time"

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

	// 自动迁移数据库表
	if err := DB.AutoMigrate(&TokenBlacklist{}); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	log.Println("数据库表迁移完成")
}

// 保留向后兼容性
func InitSQLite() {
	InitDatabase()
}

func GetDB() *gorm.DB {
	return DB
}

// AddTokenToBlacklist 将token添加到黑名单
func AddTokenToBlacklist(token string, username string, expiresAt time.Time, reason string) error {
	blacklistEntry := TokenBlacklist{
		Token:     token,
		Username:  username,
		ExpiresAt: expiresAt,
		Reason:    reason,
	}

	return DB.Create(&blacklistEntry).Error
}

// IsTokenBlacklisted 检查token是否在黑名单中
func IsTokenBlacklisted(token string) bool {
	var count int64
	DB.Model(&TokenBlacklist{}).Where("token = ? AND expires_at > ?", token, time.Now()).Count(&count)
	return count > 0
}

// CleanExpiredTokens 清理过期的黑名单token（可以通过定时任务调用）
func CleanExpiredTokens() error {
	return DB.Where("expires_at < ?", time.Now()).Delete(&TokenBlacklist{}).Error
}
