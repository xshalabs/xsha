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
			log.Fatal("MySQL DSN not configured, please set SLEEP0_MYSQL_DSN environment variable")
		}
		DB, err = gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{})
		if err != nil {
			log.Fatalf("Failed to connect to MySQL database: %v", err)
		}
		log.Println("MySQL database connected successfully")
	case "sqlite":
		DB, err = gorm.Open(sqlite.Open(cfg.SQLitePath), &gorm.Config{})
		if err != nil {
			log.Fatalf("Failed to connect to SQLite database: %v", err)
		}
		log.Println("SQLite database connected successfully")
	default:
		log.Fatalf("Unsupported database type: %s", cfg.DatabaseType)
	}

	// Auto-migrate database tables
	if err := DB.AutoMigrate(&TokenBlacklist{}); err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}
	log.Println("Database table migration completed")
}

// Maintain backward compatibility
func InitSQLite() {
	InitDatabase()
}

func GetDB() *gorm.DB {
	return DB
}

// AddTokenToBlacklist adds token to blacklist
func AddTokenToBlacklist(token string, username string, expiresAt time.Time, reason string) error {
	blacklistEntry := TokenBlacklist{
		Token:     token,
		Username:  username,
		ExpiresAt: expiresAt,
		Reason:    reason,
	}

	return DB.Create(&blacklistEntry).Error
}

// IsTokenBlacklisted checks if token is in blacklist
func IsTokenBlacklisted(token string) bool {
	var count int64
	DB.Model(&TokenBlacklist{}).Where("token = ? AND expires_at > ?", token, time.Now()).Count(&count)
	return count > 0
}

// CleanExpiredTokens cleans expired blacklist tokens (can be called by scheduled tasks)
func CleanExpiredTokens() error {
	return DB.Where("expires_at < ?", time.Now()).Delete(&TokenBlacklist{}).Error
}
