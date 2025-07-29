package config

import (
	"os"
	"strconv"
	"xsha-backend/utils"

	"github.com/joho/godotenv"
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
	AESKey       string

	SchedulerInterval      string
	WorkspaceBaseDir       string
	DockerExecutionTimeout string
	MaxConcurrentTasks     int

	GitSSLVerify bool

	LogLevel  utils.LogLevel
	LogFormat utils.LogFormat
	LogOutput string
}

func Load() *Config {

	if err := godotenv.Load(); err != nil {
		utils.Info("No .env file found or failed to load, using environment variables and default values", "error", err.Error())
	} else {
		utils.Info("Successfully loaded .env file")
	}

	aesKey := normalizeAESKey(getEnv("XSHA_AES_KEY", "default-aes-key-change-in-production"))

	config := &Config{
		Port:         getEnv("XSHA_PORT", "8080"),
		Environment:  getEnv("XSHA_ENVIRONMENT", "development"),
		DatabaseType: getEnv("XSHA_DATABASE_TYPE", "sqlite"),
		SQLitePath:   getEnv("XSHA_SQLITE_PATH", "app.db"),
		MySQLDSN:     getEnv("XSHA_MYSQL_DSN", ""),
		AdminUser:    getEnv("XSHA_ADMIN_USER", "admin"),
		AdminPass:    getEnv("XSHA_ADMIN_PASS", "admin123"),
		JWTSecret:    getEnv("XSHA_JWT_SECRET", "your-jwt-secret-key-change-this-in-production"),
		AESKey:       aesKey,

		SchedulerInterval:      getEnv("XSHA_SCHEDULER_INTERVAL", "30s"),
		WorkspaceBaseDir:       getEnv("XSHA_WORKSPACE_BASE_DIR", "/tmp/xsha-workspaces"),
		DockerExecutionTimeout: getEnv("XSHA_DOCKER_TIMEOUT", "30m"),
		MaxConcurrentTasks:     getEnvInt("XSHA_MAX_CONCURRENT_TASKS", 5),

		GitSSLVerify: getEnvBool("XSHA_GIT_SSL_VERIFY", false),

		LogLevel:  utils.LogLevel(getEnv("XSHA_LOG_LEVEL", "INFO")),
		LogFormat: utils.LogFormat(getEnv("XSHA_LOG_FORMAT", "JSON")),
		LogOutput: getEnv("XSHA_LOG_OUTPUT", "stdout"),
	}

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
		utils.Warn("Warning: Failed to parse environment variable as integer, using default value", "key", key, "value", value, "default", defaultValue)
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
		utils.Warn("Warning: Failed to parse environment variable as boolean, using default value", "key", key, "value", value, "default", defaultValue)
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
