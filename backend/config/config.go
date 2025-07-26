package config

import (
	"os"
	"sleep0-backend/utils"
	"strconv"
)

type Config struct {
	Port         string
	Environment  string
	DatabaseType string
	SQLitePath   string
	MySQLDSN     string
	AdminUser    string
	AdminPass    string // 解密后的明文密码
	JWTSecret    string
	AESKey       string // 新增：AES密钥

	// 定时器配置
	SchedulerInterval      string // 定时器间隔
	WorkspaceBaseDir       string // 工作目录基础路径
	DockerExecutionTimeout string // Docker执行超时时间
	MaxConcurrentTasks     int    // 最大并发任务数

	// Git配置
	GitSSLVerify bool // Git SSL验证开关

	// 日志配置
	LogLevel  utils.LogLevel  // 日志级别
	LogFormat utils.LogFormat // 日志格式
	LogOutput string          // 日志输出
}

func Load() *Config {
	// 获取AES密钥
	aesKey := normalizeAESKey(getEnv("SLEEP0_AES_KEY", "default-aes-key-change-in-production"))

	config := &Config{
		Port:         getEnv("SLEEP0_PORT", "8080"),
		Environment:  getEnv("SLEEP0_ENVIRONMENT", "development"),
		DatabaseType: getEnv("SLEEP0_DATABASE_TYPE", "sqlite"),
		SQLitePath:   getEnv("SLEEP0_SQLITE_PATH", "app.db"),
		MySQLDSN:     getEnv("SLEEP0_MYSQL_DSN", ""),
		AdminUser:    getEnv("SLEEP0_ADMIN_USER", "admin"),
		JWTSecret:    getEnv("SLEEP0_JWT_SECRET", "your-jwt-secret-key-change-this-in-production"),
		AESKey:       aesKey,

		// 定时器配置
		SchedulerInterval:      getEnv("SLEEP0_SCHEDULER_INTERVAL", "30s"),
		WorkspaceBaseDir:       getEnv("SLEEP0_WORKSPACE_BASE_DIR", "/tmp/sleep0-workspaces"),
		DockerExecutionTimeout: getEnv("SLEEP0_DOCKER_TIMEOUT", "30m"),
		MaxConcurrentTasks:     getEnvInt("SLEEP0_MAX_CONCURRENT_TASKS", 5),

		// Git配置 - 默认禁用SSL验证以解决兼容性问题
		GitSSLVerify: getEnvBool("SLEEP0_GIT_SSL_VERIFY", false),

		// 日志配置
		LogLevel:  utils.LogLevel(getEnv("SLEEP0_LOG_LEVEL", "INFO")),
		LogFormat: utils.LogFormat(getEnv("SLEEP0_LOG_FORMAT", "JSON")),
		LogOutput: getEnv("SLEEP0_LOG_OUTPUT", "stdout"),
	}

	// 处理管理员密码：尝试解密，失败则当作明文（向后兼容）
	encryptedPass := getEnv("SLEEP0_ADMIN_PASS", "admin123")
	if decryptedPass, err := utils.DecryptAES(encryptedPass, aesKey); err == nil {
		config.AdminPass = decryptedPass
		utils.Info("Administrator password loaded from encrypted value")
	} else {
		config.AdminPass = encryptedPass
		utils.Info("管理员密码作为明文加载（建议加密）")
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
		utils.Warn("警告：无法解析环境变量的值为整数，使用默认值", "key", key, "value", value, "default", defaultValue)
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
		utils.Warn("警告：无法解析环境变量的值为布尔值，使用默认值", "key", key, "value", value, "default", defaultValue)
	}
	return defaultValue
}

// normalizeAESKey 标准化AES密钥为32字节
func normalizeAESKey(key string) string {
	if len(key) >= 32 {
		return key[:32]
	}
	normalized := make([]byte, 32)
	copy(normalized, []byte(key))
	return string(normalized)
}
