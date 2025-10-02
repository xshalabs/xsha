package config

import (
	"os"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Port != "8080" {
		t.Errorf("Expected default port to be 8080, got %s", cfg.Port)
	}

	if cfg.Environment != "production" {
		t.Errorf("Expected default environment to be production, got %s", cfg.Environment)
	}

	if cfg.DatabaseType != "sqlite" {
		t.Errorf("Expected default database type to be sqlite, got %s", cfg.DatabaseType)
	}

	if cfg.MaxConcurrentTasks != 8 {
		t.Errorf("Expected default max concurrent tasks to be 8, got %d", cfg.MaxConcurrentTasks)
	}

	if cfg.LogLevel != LevelInfo {
		t.Errorf("Expected default log level to be INFO, got %s", cfg.LogLevel)
	}

	if cfg.LogFormat != FormatJSON {
		t.Errorf("Expected default log format to be JSON, got %s", cfg.LogFormat)
	}
}

func TestLogLevelUnmarshalText(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
		wantErr  bool
	}{
		{"debug", LevelDebug, false},
		{"DEBUG", LevelDebug, false},
		{"info", LevelInfo, false},
		{"INFO", LevelInfo, false},
		{"warn", LevelWarn, false},
		{"WARN", LevelWarn, false},
		{"error", LevelError, false},
		{"ERROR", LevelError, false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var level LogLevel
			err := level.UnmarshalText([]byte(tt.input))

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error for input %s, but got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input %s: %v", tt.input, err)
				}
				if level != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, level)
				}
			}
		})
	}
}

func TestLogFormatUnmarshalText(t *testing.T) {
	tests := []struct {
		input    string
		expected LogFormat
		wantErr  bool
	}{
		{"json", FormatJSON, false},
		{"JSON", FormatJSON, false},
		{"text", FormatText, false},
		{"TEXT", FormatText, false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var format LogFormat
			err := format.UnmarshalText([]byte(tt.input))

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error for input %s, but got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input %s: %v", tt.input, err)
				}
				if format != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, format)
				}
			}
		})
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "Valid config",
			config: &Config{
				Port:                     "8080",
				Environment:              "development",
				DatabaseType:             "sqlite",
				SQLitePath:               "test.db",
				SchedulerIntervalDuration: 5 * time.Second,
				MaxConcurrentTasks:       8,
				LogLevel:                 LevelInfo,
				LogFormat:                FormatJSON,
			},
		},
		{
			name: "Missing port",
			config: &Config{
				Environment:              "development",
				DatabaseType:             "sqlite",
				SQLitePath:               "test.db",
				SchedulerIntervalDuration: 5 * time.Second,
				MaxConcurrentTasks:       8,
				LogLevel:                 LevelInfo,
				LogFormat:                FormatJSON,
			},
			wantErr: true,
		},
		{
			name: "Missing environment",
			config: &Config{
				Port:                     "8080",
				DatabaseType:             "sqlite",
				SQLitePath:               "test.db",
				SchedulerIntervalDuration: 5 * time.Second,
				MaxConcurrentTasks:       8,
				LogLevel:                 LevelInfo,
				LogFormat:                FormatJSON,
			},
			wantErr: true,
		},
		{
			name: "Production without JWT secret",
			config: &Config{
				Port:                     "8080",
				Environment:              "production",
				DatabaseType:             "sqlite",
				SQLitePath:               "test.db",
				SchedulerIntervalDuration: 5 * time.Second,
				MaxConcurrentTasks:       8,
				LogLevel:                 LevelInfo,
				LogFormat:                FormatJSON,
			},
			wantErr: true,
		},
		{
			name: "Invalid database type",
			config: &Config{
				Port:                     "8080",
				Environment:              "development",
				DatabaseType:             "postgres",
				SchedulerIntervalDuration: 5 * time.Second,
				MaxConcurrentTasks:       8,
				LogLevel:                 LevelInfo,
				LogFormat:                FormatJSON,
			},
			wantErr: true,
		},
		{
			name: "SQLite without path",
			config: &Config{
				Port:                     "8080",
				Environment:              "development",
				DatabaseType:             "sqlite",
				SchedulerIntervalDuration: 5 * time.Second,
				MaxConcurrentTasks:       8,
				LogLevel:                 LevelInfo,
				LogFormat:                FormatJSON,
			},
			wantErr: true,
		},
		{
			name: "MySQL without DSN",
			config: &Config{
				Port:                     "8080",
				Environment:              "development",
				DatabaseType:             "mysql",
				SchedulerIntervalDuration: 5 * time.Second,
				MaxConcurrentTasks:       8,
				LogLevel:                 LevelInfo,
				LogFormat:                FormatJSON,
			},
			wantErr: true,
		},
		{
			name: "Invalid scheduler interval",
			config: &Config{
				Port:                     "8080",
				Environment:              "development",
				DatabaseType:             "sqlite",
				SQLitePath:               "test.db",
				SchedulerIntervalDuration: -5 * time.Second,
				MaxConcurrentTasks:       8,
				LogLevel:                 LevelInfo,
				LogFormat:                FormatJSON,
			},
			wantErr: true,
		},
		{
			name: "Invalid max concurrent tasks",
			config: &Config{
				Port:                     "8080",
				Environment:              "development",
				DatabaseType:             "sqlite",
				SQLitePath:               "test.db",
				SchedulerIntervalDuration: 5 * time.Second,
				MaxConcurrentTasks:       0,
				LogLevel:                 LevelInfo,
				LogFormat:                FormatJSON,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadWithEnvironmentVariables(t *testing.T) {
	// Save current environment
	originalEnv := os.Environ()
	defer func() {
		// Restore environment
		os.Clearenv()
		for _, e := range originalEnv {
			pair := splitEnvVar(e)
			os.Setenv(pair[0], pair[1])
		}
	}()

	// Set test environment variables
	os.Setenv("XSHA_PORT", "9090")
	os.Setenv("XSHA_ENVIRONMENT", "development")
	os.Setenv("XSHA_DATABASE_TYPE", "mysql")
	os.Setenv("XSHA_MYSQL_DSN", "test:test@tcp(localhost:3306)/testdb")
	os.Setenv("XSHA_JWT_SECRET", "test-secret")
	os.Setenv("XSHA_MAX_CONCURRENT_TASKS", "16")
	os.Setenv("XSHA_LOG_LEVEL", "debug")
	os.Setenv("XSHA_LOG_FORMAT", "text")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Unexpected error loading config: %v", err)
	}

	if cfg.Port != "9090" {
		t.Errorf("Expected port 9090, got %s", cfg.Port)
	}

	if cfg.Environment != "development" {
		t.Errorf("Expected environment development, got %s", cfg.Environment)
	}

	if cfg.DatabaseType != "mysql" {
		t.Errorf("Expected database type mysql, got %s", cfg.DatabaseType)
	}

	if cfg.MySQLDSN != "test:test@tcp(localhost:3306)/testdb" {
		t.Errorf("Expected MySQL DSN test:test@tcp(localhost:3306)/testdb, got %s", cfg.MySQLDSN)
	}

	if cfg.MaxConcurrentTasks != 16 {
		t.Errorf("Expected max concurrent tasks 16, got %d", cfg.MaxConcurrentTasks)
	}

	if cfg.LogLevel != LevelDebug {
		t.Errorf("Expected log level DEBUG, got %s", cfg.LogLevel)
	}

	if cfg.LogFormat != FormatText {
		t.Errorf("Expected log format TEXT, got %s", cfg.LogFormat)
	}
}

func TestDockerVolumeConfiguration(t *testing.T) {
	// Save current environment
	originalEnv := os.Environ()
	defer func() {
		// Restore environment
		os.Clearenv()
		for _, e := range originalEnv {
			pair := splitEnvVar(e)
			os.Setenv(pair[0], pair[1])
		}
	}()

	// Set test environment variables
	os.Setenv("XSHA_ENVIRONMENT", "development")
	os.Setenv("XSHA_DOCKER_VOLUME_WORKSPACES", "xsha-workspaces-volume")
	os.Setenv("XSHA_DOCKER_VOLUME_SESSIONS", "xsha-sessions-volume")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Unexpected error loading config: %v", err)
	}

	if cfg.DockerVolumeWorkspaces != "xsha-workspaces-volume" {
		t.Errorf("Expected DockerVolumeWorkspaces to be 'xsha-workspaces-volume', got %s", cfg.DockerVolumeWorkspaces)
	}

	if cfg.DockerVolumeSessions != "xsha-sessions-volume" {
		t.Errorf("Expected DockerVolumeSessions to be 'xsha-sessions-volume', got %s", cfg.DockerVolumeSessions)
	}
}

func TestConfigOptions(t *testing.T) {
	// Save current environment
	originalEnv := os.Environ()
	defer func() {
		// Restore environment
		os.Clearenv()
		for _, e := range originalEnv {
			pair := splitEnvVar(e)
			os.Setenv(pair[0], pair[1])
		}
	}()

	// Set minimal environment for valid config
	os.Setenv("XSHA_JWT_SECRET", "test-secret")

	cfg, err := LoadWithOptions(
		WithPort("7070"),
		WithEnvironment("staging"),
		WithJWTSecret("override-secret"),
	)

	if err != nil {
		t.Fatalf("Unexpected error loading config with options: %v", err)
	}

	if cfg.Port != "7070" {
		t.Errorf("Expected port 7070, got %s", cfg.Port)
	}

	if cfg.Environment != "staging" {
		t.Errorf("Expected environment staging, got %s", cfg.Environment)
	}

	if cfg.JWTSecret != "override-secret" {
		t.Errorf("Expected JWT secret override-secret, got %s", cfg.JWTSecret)
	}
}

func TestConfigEnvironmentHelpers(t *testing.T) {
	tests := []struct {
		environment string
		isProd      bool
		isDev       bool
	}{
		{"production", true, false},
		{"prod", true, false}, // "prod" is not recognized as development
		{"development", false, true},
		{"dev", false, true},
		{"staging", true, false},
		{"test", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.environment, func(t *testing.T) {
			cfg := &Config{Environment: tt.environment}

			if cfg.IsProduction() != tt.isProd {
				t.Errorf("IsProduction() = %v, want %v", cfg.IsProduction(), tt.isProd)
			}

			if cfg.IsDevelopment() != tt.isDev {
				t.Errorf("IsDevelopment() = %v, want %v", cfg.IsDevelopment(), tt.isDev)
			}
		})
	}
}

// Helper function to split environment variable
func splitEnvVar(env string) []string {
	for i := 0; i < len(env); i++ {
		if env[i] == '=' {
			return []string{env[:i], env[i+1:]}
		}
	}
	return []string{env, ""}
}