package config

import (
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
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

	SchedulerInterval         string
	SchedulerIntervalDuration time.Duration
	WorkspaceBaseDir          string
	DevSessionsDir            string
	AttachmentsDir            string
	AvatarsDir                string
	MaxConcurrentTasks        int

	LogLevel  LogLevel
	LogFormat LogFormat
	LogOutput string
}

func Load() *Config {
	// Create a simple logger for config loading since main logger isn't initialized yet
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	if err := godotenv.Load(); err == nil {
		logger.Info("Successfully loaded .env file")
	}

	environment := getEnv("XSHA_ENVIRONMENT", "production")

	// Set default log level and format based on environment
	defaultLogLevel := "INFO"
	defaultLogFormat := "JSON"
	if environment == "development" || environment == "dev" {
		defaultLogLevel = "DEBUG"
		defaultLogFormat = "TEXT"
	}

	config := &Config{
		Port:         getEnv("XSHA_PORT", "8080"),
		Environment:  environment,
		DatabaseType: getEnv("XSHA_DATABASE_TYPE", "sqlite"),
		SQLitePath:   getEnv("XSHA_SQLITE_PATH", "app.db"),
		MySQLDSN:     getEnv("XSHA_MYSQL_DSN", ""),
		JWTSecret:    getEnv("XSHA_JWT_SECRET", "your-jwt-secret-key-change-this-in-production"),

		SchedulerInterval:  getEnv("XSHA_SCHEDULER_INTERVAL", "5s"),
		WorkspaceBaseDir:   getEnv("XSHA_WORKSPACE_BASE_DIR", "_data/workspaces"),
		DevSessionsDir:     getEnv("XSHA_DEV_SESSIONS_DIR", "_data/sessions"),
		AttachmentsDir:     getEnv("XSHA_ATTACHMENTS_DIR", "_data/attachments"),
		AvatarsDir:         getEnv("XSHA_AVATARS_DIR", "_data/avatars"),
		MaxConcurrentTasks: getEnvInt("XSHA_MAX_CONCURRENT_TASKS", 8),
		LogLevel:           LogLevel(getEnv("XSHA_LOG_LEVEL", defaultLogLevel)),
		LogFormat:          LogFormat(getEnv("XSHA_LOG_FORMAT", defaultLogFormat)),
		LogOutput:          getEnv("XSHA_LOG_OUTPUT", "stdout"),
	}

	schedulerInterval, err := time.ParseDuration(config.SchedulerInterval)
	if err != nil {
		logger.Warn("Failed to parse scheduler interval, using default 30 seconds",
			zap.String("interval", config.SchedulerInterval),
			zap.Error(err))
		schedulerInterval = 30 * time.Second
	}
	config.SchedulerIntervalDuration = schedulerInterval

	// Normalize paths to absolute paths for Docker compatibility
	config.WorkspaceBaseDir = normalizeConfigPath(config.WorkspaceBaseDir)
	config.DevSessionsDir = normalizeConfigPath(config.DevSessionsDir)
	config.AttachmentsDir = normalizeConfigPath(config.AttachmentsDir)
	config.AvatarsDir = normalizeConfigPath(config.AvatarsDir)

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
		// Create a simple logger for this warning
		logger, _ := zap.NewDevelopment()
		defer logger.Sync()
		logger.Warn("Failed to parse environment variable as integer, using default value",
			zap.String("key", key),
			zap.String("value", value),
			zap.Int("default", defaultValue))
	}
	return defaultValue
}

// normalizeConfigPath converts relative paths to absolute paths
func normalizeConfigPath(path string) string {
	if path == "" {
		return path
	}

	// Check if path is already absolute
	if filepath.IsAbs(path) {
		return path
	}

	// Convert relative path to absolute
	absPath, err := filepath.Abs(path)
	if err != nil {
		logger, _ := zap.NewDevelopment()
		defer logger.Sync()
		logger.Warn("Failed to convert path to absolute",
			zap.String("path", path),
			zap.Error(err))
		return path // Return original path if conversion fails
	}

	return absPath
}
