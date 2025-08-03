package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type LogLevel string

const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
)

type LogFormat string

const (
	FormatJSON LogFormat = "JSON"
	FormatText LogFormat = "TEXT"
)

type Config struct {
	Port         string
	Environment  string
	DatabaseType string
	SQLitePath   string
	MySQLDSN     string
	JWTSecret    string
	AESKey       string

	SchedulerInterval         string
	SchedulerIntervalDuration time.Duration
	WorkspaceBaseDir          string
	DevSessionsDir            string
	MaxConcurrentTasks        int

	LogLevel  LogLevel
	LogFormat LogFormat
	LogOutput string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		fmt.Printf("No .env file found or failed to load, using environment variables and default values: %v\n", err)
	} else {
		fmt.Println("Successfully loaded .env file")
	}

	config := &Config{
		Port:               getEnv("XSHA_PORT", "8080"),
		Environment:        getEnv("XSHA_ENVIRONMENT", "production"),
		DatabaseType:       getEnv("XSHA_DATABASE_TYPE", "sqlite"),
		SQLitePath:         getEnv("XSHA_SQLITE_PATH", "app.db"),
		MySQLDSN:           getEnv("XSHA_MYSQL_DSN", ""),
		JWTSecret:          getEnv("XSHA_JWT_SECRET", "your-jwt-secret-key-change-this-in-production"),
		AESKey:             normalizeAESKey(getEnv("XSHA_AES_KEY", "default-aes-key-change-in-production")),
		SchedulerInterval:  getEnv("XSHA_SCHEDULER_INTERVAL", "5s"),
		WorkspaceBaseDir:   getEnv("XSHA_WORKSPACE_BASE_DIR", "/tmp/xsha-workspaces"),
		DevSessionsDir:     getEnv("XSHA_DEV_SESSIONS_DIR", "/tmp/xsha-dev-sessions"),
		MaxConcurrentTasks: getEnvInt("XSHA_MAX_CONCURRENT_TASKS", 8),
		LogLevel:           LogLevel(getEnv("XSHA_LOG_LEVEL", "INFO")),
		LogFormat:          LogFormat(getEnv("XSHA_LOG_FORMAT", "JSON")),
		LogOutput:          getEnv("XSHA_LOG_OUTPUT", "stdout"),
	}

	schedulerInterval, err := time.ParseDuration(config.SchedulerInterval)
	if err != nil {
		fmt.Printf("Warning: Failed to parse scheduler interval, using default 30 seconds. interval=%s error=%v\n", config.SchedulerInterval, err)
		schedulerInterval = 30 * time.Second
	}
	config.SchedulerIntervalDuration = schedulerInterval

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		fmt.Printf("Warning: Failed to parse environment variable as integer, using default value. key=%s value=%s default=%d\n", key, value, defaultValue)
	}
	return defaultValue
}

func normalizeAESKey(key string) string {
	if len(key) >= 32 {
		return key[:32]
	}
	normalized := make([]byte, 32)
	copy(normalized, []byte(key))
	return string(normalized)
}
