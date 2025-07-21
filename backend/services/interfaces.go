package services

import (
	"sleep0-backend/database"
)

// AuthService 定义认证服务接口
type AuthService interface {
	// Login 用户登录
	Login(username, password, clientIP, userAgent string) (bool, string, error)

	// Logout 用户登出
	Logout(token, username string) error

	// IsTokenBlacklisted 检查Token是否在黑名单
	IsTokenBlacklisted(token string) (bool, error)

	// CleanExpiredTokens 清理过期Token
	CleanExpiredTokens() error
}

// LoginLogService 定义登录日志服务接口
type LoginLogService interface {
	// GetLogs 获取登录日志
	GetLogs(username string, page, pageSize int) ([]database.LoginLog, int64, error)

	// CleanOldLogs 清理旧日志
	CleanOldLogs(days int) error
}
