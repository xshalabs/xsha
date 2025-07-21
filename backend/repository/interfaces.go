package repository

import (
	"sleep0-backend/database"
	"time"
)

// TokenBlacklistRepository 定义Token黑名单仓库接口
type TokenBlacklistRepository interface {
	Add(token string, username string, expiresAt time.Time, reason string) error
	IsBlacklisted(token string) (bool, error)
	CleanExpired() error
}

// LoginLogRepository 定义登录日志仓库接口
type LoginLogRepository interface {
	Add(username, ip, userAgent, reason string, success bool) error
	GetLogs(username string, page, pageSize int) ([]database.LoginLog, int64, error)
	CleanOld(days int) error
}
