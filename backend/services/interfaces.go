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

// GitCredentialService 定义Git凭据服务接口
type GitCredentialService interface {
	// 凭据管理
	CreateCredential(name, description, credType, username, createdBy string, secretData map[string]string) (*database.GitCredential, error)
	GetCredential(id uint, createdBy string) (*database.GitCredential, error)
	GetCredentialByName(name, createdBy string) (*database.GitCredential, error)
	ListCredentials(createdBy string, credType *database.GitCredentialType, page, pageSize int) ([]database.GitCredential, int64, error)
	UpdateCredential(id uint, createdBy string, updates map[string]interface{}, secretData map[string]string) error
	DeleteCredential(id uint, createdBy string) error

	// 凭据操作
	UseCredential(id uint, createdBy string) (*database.GitCredential, error)
	ToggleCredential(id uint, createdBy string, isActive bool) error
	ListActiveCredentials(createdBy string, credType *database.GitCredentialType) ([]database.GitCredential, error)

	// 凭据验证和解密
	DecryptCredentialSecret(credential *database.GitCredential, secretType string) (string, error)
	ValidateCredentialData(credType string, data map[string]string) error
}

// ProjectService 定义项目服务接口
type ProjectService interface {
	// 项目管理
	CreateProject(name, description, repoURL, protocol, defaultBranch, createdBy string, credentialID *uint) (*database.Project, error)
	GetProject(id uint, createdBy string) (*database.Project, error)
	GetProjectByName(name, createdBy string) (*database.Project, error)
	ListProjects(createdBy string, protocol *database.GitProtocolType, page, pageSize int) ([]database.Project, int64, error)
	UpdateProject(id uint, createdBy string, updates map[string]interface{}) error
	DeleteProject(id uint, createdBy string) error

	// 项目操作
	UseProject(id uint, createdBy string) (*database.Project, error)
	ToggleProject(id uint, createdBy string, isActive bool) error
	ListActiveProjects(createdBy string, protocol *database.GitProtocolType) ([]database.Project, error)

	// 协议和凭据验证
	ValidateProtocolCredential(protocol database.GitProtocolType, credentialID *uint, createdBy string) error
	GetCompatibleCredentials(protocol database.GitProtocolType, createdBy string) ([]database.GitCredential, error)
}
