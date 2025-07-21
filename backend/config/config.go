package config

import (
	"log"
	"os"
	"sleep0-backend/utils"
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
	}

	// 处理管理员密码：尝试解密，失败则当作明文（向后兼容）
	encryptedPass := getEnv("SLEEP0_ADMIN_PASS", "admin123")
	if decryptedPass, err := utils.DecryptAES(encryptedPass, aesKey); err == nil {
		config.AdminPass = decryptedPass
		log.Println("管理员密码已从加密值加载")
	} else {
		config.AdminPass = encryptedPass
		log.Println("管理员密码作为明文加载（建议加密）")
	}

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
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
