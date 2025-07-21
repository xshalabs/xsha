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

// GitCredentialRepository 定义Git凭据仓库接口
type GitCredentialRepository interface {
	// 基本CRUD操作
	Create(credential *database.GitCredential) error
	GetByID(id uint, createdBy string) (*database.GitCredential, error)
	GetByName(name, createdBy string) (*database.GitCredential, error)
	List(createdBy string, credType *database.GitCredentialType, page, pageSize int) ([]database.GitCredential, int64, error)
	Update(credential *database.GitCredential) error
	Delete(id uint, createdBy string) error

	// 业务操作
	UpdateLastUsed(id uint, createdBy string) error
	SetActive(id uint, createdBy string, isActive bool) error
	ListActive(createdBy string, credType *database.GitCredentialType) ([]database.GitCredential, error)
}
