package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"xsha-backend/utils"

	"github.com/joho/godotenv"
)

// Type aliases for backward compatibility
type LogLevel = utils.LogLevel
type LogFormat = utils.LogFormat

// Re-export constants for backward compatibility
const (
	LevelDebug = utils.LevelDebug
	LevelInfo  = utils.LevelInfo
	LevelWarn  = utils.LevelWarn
	LevelError = utils.LevelError

	FormatJSON = utils.FormatJSON
	FormatText = utils.FormatText
)

// Config holds all application configuration
type Config struct {
	// Server configuration
	Port        string
	Environment string

	// Database configuration
	DatabaseType string
	SQLitePath   string
	MySQLDSN     string

	// Security
	JWTSecret string

	// Scheduler configuration
	SchedulerInterval         string
	SchedulerIntervalDuration time.Duration

	// Storage paths
	WorkspaceBaseDir string
	DevSessionsDir   string
	AttachmentsDir   string
	AvatarsDir       string

	// Execution limits
	MaxConcurrentTasks int

	// Logging configuration
	LogLevel  LogLevel
	LogFormat LogFormat
	LogOutput string

	// Docker volume configuration
	DockerVolumeWorkspaces     string // Docker workspaces volume name from environment variable
	DockerVolumeWorkspacesPath string // Resolved real path of the Docker workspaces volume
	DockerVolumeSessions       string // Docker sessions volume name from environment variable
	DockerVolumeSessionsPath   string // Resolved real path of the Docker sessions volume
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Port:                   "8080",
		Environment:            "production",
		DatabaseType:           "sqlite",
		SQLitePath:             "app.db",
		MySQLDSN:               "",
		JWTSecret:              "your-jwt-secret-key-change-this-in-production",
		SchedulerInterval:      "5s",
		WorkspaceBaseDir:       "_data/workspaces",
		DevSessionsDir:         "_data/sessions",
		AttachmentsDir:         "_data/attachments",
		AvatarsDir:             "_data/avatars",
		MaxConcurrentTasks:     8,
		LogLevel:               LevelInfo,
		LogFormat:              FormatJSON,
		LogOutput:              "stdout",
		DockerVolumeWorkspaces: "",
		DockerVolumeSessions:   "",
	}
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if not found)
	_ = godotenv.Load()

	// Start with default configuration
	cfg := DefaultConfig()

	// Determine environment
	cfg.Environment = getEnv("XSHA_ENVIRONMENT", cfg.Environment)

	// Adjust defaults based on environment
	if isDevelopmentEnvironment(cfg.Environment) {
		cfg.LogLevel = LevelDebug
		cfg.LogFormat = FormatText
	}

	// Load configuration from environment variables
	cfg.Port = getEnv("XSHA_PORT", cfg.Port)
	cfg.DatabaseType = getEnv("XSHA_DATABASE_TYPE", cfg.DatabaseType)
	cfg.SQLitePath = getEnv("XSHA_SQLITE_PATH", cfg.SQLitePath)
	cfg.MySQLDSN = getEnv("XSHA_MYSQL_DSN", cfg.MySQLDSN)
	cfg.JWTSecret = getEnv("XSHA_JWT_SECRET", cfg.JWTSecret)
	cfg.SchedulerInterval = getEnv("XSHA_SCHEDULER_INTERVAL", cfg.SchedulerInterval)
	cfg.WorkspaceBaseDir = getEnv("XSHA_WORKSPACE_BASE_DIR", cfg.WorkspaceBaseDir)
	cfg.DevSessionsDir = getEnv("XSHA_DEV_SESSIONS_DIR", cfg.DevSessionsDir)
	cfg.AttachmentsDir = getEnv("XSHA_ATTACHMENTS_DIR", cfg.AttachmentsDir)
	cfg.AvatarsDir = getEnv("XSHA_AVATARS_DIR", cfg.AvatarsDir)
	cfg.MaxConcurrentTasks = getEnvInt("XSHA_MAX_CONCURRENT_TASKS", cfg.MaxConcurrentTasks)
	cfg.LogOutput = getEnv("XSHA_LOG_OUTPUT", cfg.LogOutput)
	cfg.DockerVolumeWorkspaces = getEnv("XSHA_DOCKER_VOLUME_WORKSPACES", cfg.DockerVolumeWorkspaces)
	cfg.DockerVolumeSessions = getEnv("XSHA_DOCKER_VOLUME_SESSIONS", cfg.DockerVolumeSessions)

	// Parse log level and format
	if levelStr := getEnv("XSHA_LOG_LEVEL", ""); levelStr != "" {
		if err := cfg.LogLevel.UnmarshalText([]byte(levelStr)); err != nil {
			return nil, fmt.Errorf("invalid log level: %w", err)
		}
	}

	if formatStr := getEnv("XSHA_LOG_FORMAT", ""); formatStr != "" {
		if err := cfg.LogFormat.UnmarshalText([]byte(formatStr)); err != nil {
			return nil, fmt.Errorf("invalid log format: %w", err)
		}
	}

	// Parse scheduler interval
	schedulerInterval, err := time.ParseDuration(cfg.SchedulerInterval)
	if err != nil {
		// Use default interval on parse error
		schedulerInterval = 30 * time.Second
		cfg.SchedulerInterval = "30s"
	}
	cfg.SchedulerIntervalDuration = schedulerInterval

	// Normalize paths to absolute paths
	if err := cfg.normalizePaths(); err != nil {
		return nil, fmt.Errorf("failed to normalize paths: %w", err)
	}

	// Handle Docker volume configuration
	if err := cfg.resolveDockerVolume(); err != nil {
		// Docker volume resolution failure is not fatal
		// Just log it when logger is available
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// LoadWithOptions loads configuration with custom options
func LoadWithOptions(opts ...Option) (*Config, error) {
	cfg, err := Load()
	if err != nil {
		return nil, err
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, fmt.Errorf("failed to apply configuration option: %w", err)
		}
	}

	// Re-validate after applying options
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed after applying options: %w", err)
	}

	return cfg, nil
}

// Option is a functional option for configuration
type Option func(*Config) error

// WithPort sets the server port
func WithPort(port string) Option {
	return func(c *Config) error {
		c.Port = port
		return nil
	}
}

// WithEnvironment sets the environment
func WithEnvironment(env string) Option {
	return func(c *Config) error {
		c.Environment = env
		return nil
	}
}

// WithJWTSecret sets the JWT secret
func WithJWTSecret(secret string) Option {
	return func(c *Config) error {
		if secret == "" {
			return errors.New("JWT secret cannot be empty")
		}
		c.JWTSecret = secret
		return nil
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate required fields
	if c.Port == "" {
		return errors.New("port is required")
	}

	if c.Environment == "" {
		return errors.New("environment is required")
	}

	// Validate JWT secret in production
	if !isDevelopmentEnvironment(c.Environment) && c.JWTSecret == "" {
		return errors.New("JWT secret is required in production environment")
	}

	// Validate database configuration
	switch c.DatabaseType {
	case "sqlite":
		if c.SQLitePath == "" {
			return errors.New("SQLite path is required when using SQLite database")
		}
	case "mysql":
		if c.MySQLDSN == "" {
			return errors.New("MySQL DSN is required when using MySQL database")
		}
	default:
		return fmt.Errorf("unsupported database type: %s", c.DatabaseType)
	}

	// Validate log configuration
	if !c.LogLevel.IsValid() {
		return fmt.Errorf("invalid log level: %s", c.LogLevel)
	}

	if !c.LogFormat.IsValid() {
		return fmt.Errorf("invalid log format: %s", c.LogFormat)
	}

	// Validate scheduler interval
	if c.SchedulerIntervalDuration <= 0 {
		return errors.New("scheduler interval must be positive")
	}

	// Validate max concurrent tasks
	if c.MaxConcurrentTasks <= 0 {
		return errors.New("max concurrent tasks must be positive")
	}

	return nil
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return !isDevelopmentEnvironment(c.Environment)
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return isDevelopmentEnvironment(c.Environment)
}

// normalizePaths converts relative paths to absolute paths
func (c *Config) normalizePaths() error {
	paths := []*string{
		&c.WorkspaceBaseDir,
		&c.DevSessionsDir,
		&c.AttachmentsDir,
		&c.AvatarsDir,
	}

	for _, path := range paths {
		normalized, err := normalizeConfigPath(*path)
		if err != nil {
			return fmt.Errorf("failed to normalize path %s: %w", *path, err)
		}
		*path = normalized
	}

	// Special handling for SQLite path if using SQLite
	if c.DatabaseType == "sqlite" && c.SQLitePath != "" {
		normalized, err := normalizeConfigPath(c.SQLitePath)
		if err != nil {
			return fmt.Errorf("failed to normalize SQLite path %s: %w", c.SQLitePath, err)
		}
		c.SQLitePath = normalized
	}

	return nil
}

// resolveDockerVolume resolves Docker volume path if running in container
func (c *Config) resolveDockerVolume() error {
	// Only resolve if running in Docker
	docker := utils.NewDockerDetector()
	if !docker.IsRunningInDocker() {
		return nil
	}

	resolver := utils.NewDockerVolumeResolver()

	// Resolve workspaces volume
	if c.DockerVolumeWorkspaces != "" {
		realPath, err := resolver.GetVolumeRealPath(c.DockerVolumeWorkspaces)
		if err != nil {
			return fmt.Errorf("failed to resolve Docker workspaces volume: %w", err)
		}
		c.DockerVolumeWorkspacesPath = realPath
	}

	// Resolve sessions volume
	if c.DockerVolumeSessions != "" {
		realPath, err := resolver.GetVolumeRealPath(c.DockerVolumeSessions)
		if err != nil {
			return fmt.Errorf("failed to resolve Docker sessions volume: %w", err)
		}
		c.DockerVolumeSessionsPath = realPath
	}

	return nil
}

// Helper functions

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
	}
	return defaultValue
}

func normalizeConfigPath(path string) (string, error) {
	if path == "" {
		return path, nil
	}

	// Check if path is already absolute
	if filepath.IsAbs(path) {
		return path, nil
	}

	// Convert relative path to absolute
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to convert path to absolute: %w", err)
	}

	return absPath, nil
}

func isDevelopmentEnvironment(env string) bool {
	env = strings.ToLower(env)
	return env == "development" || env == "dev"
}
