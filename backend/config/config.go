package config

import (
	"os"
)

type Config struct {
	Port         string
	Environment  string
	DatabaseType string
	SQLitePath   string
	MySQLDSN     string
	AdminUser    string
	AdminPass    string
	JWTSecret    string
}

func Load() *Config {
	return &Config{
		Port:         getEnv("SLEEP0_PORT", "8080"),
		Environment:  getEnv("SLEEP0_ENVIRONMENT", "development"),
		DatabaseType: getEnv("SLEEP0_DATABASE_TYPE", "sqlite"), // sqlite 或 mysql
		SQLitePath:   getEnv("SLEEP0_SQLITE_PATH", "app.db"),
		MySQLDSN:     getEnv("SLEEP0_MYSQL_DSN", ""),
		AdminUser:    getEnv("SLEEP0_ADMIN_USER", "admin"),
		AdminPass:    getEnv("SLEEP0_ADMIN_PASS", "admin123"),
		JWTSecret:    getEnv("SLEEP0_JWT_SECRET", "your-jwt-secret-key-change-this-in-production"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
