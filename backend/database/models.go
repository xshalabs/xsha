package database

import (
	"time"

	"gorm.io/gorm"
)

// TokenBlacklist token blacklist model
type TokenBlacklist struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Token     string         `gorm:"uniqueIndex;not null" json:"token"` // JWT token
	ExpiresAt time.Time      `gorm:"not null" json:"expires_at"`        // Token expiration time
	Username  string         `gorm:"not null" json:"username"`          // Username
	Reason    string         `gorm:"default:'logout'" json:"reason"`    // Reason for adding to blacklist
}

// LoginLog 登录日志模型
type LoginLog struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Username  string         `gorm:"not null;index" json:"username"`   // 尝试登录的用户名
	Success   bool           `gorm:"not null;index" json:"success"`    // 登录是否成功
	IP        string         `gorm:"not null" json:"ip"`               // 客户端IP地址
	UserAgent string         `gorm:"type:text" json:"user_agent"`      // 用户代理字符串
	Reason    string         `gorm:"default:''" json:"reason"`         // 失败原因
	LoginTime time.Time      `gorm:"not null;index" json:"login_time"` // 登录时间
}

// GitCredentialType 定义Git凭据类型
type GitCredentialType string

const (
	GitCredentialTypePassword GitCredentialType = "password" // 用户名密码
	GitCredentialTypeToken    GitCredentialType = "token"    // Personal Access Token
	GitCredentialTypeSSHKey   GitCredentialType = "ssh_key"  // SSH Key
)

// GitCredential Git凭据模型
type GitCredential struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 基本信息
	Name        string            `gorm:"not null;uniqueIndex:idx_name_user" json:"name"` // 凭据名称
	Description string            `gorm:"type:text" json:"description"`                   // 描述
	Type        GitCredentialType `gorm:"not null;index" json:"type"`                     // 凭据类型

	// 认证信息
	Username string `gorm:"default:''" json:"username"` // 用户名（用于password和token类型）

	// 加密存储的敏感信息
	PasswordHash string `gorm:"type:text" json:"-"`          // 加密的密码/token
	PrivateKey   string `gorm:"type:text" json:"-"`          // 加密的SSH私钥
	PublicKey    string `gorm:"type:text" json:"public_key"` // SSH公钥（不敏感，可显示）

	// 元数据
	LastUsed *time.Time `json:"last_used"`                           // 最后使用时间
	IsActive bool       `gorm:"default:true;index" json:"is_active"` // 是否激活

	// 关联用户
	CreatedBy string `gorm:"not null;index;uniqueIndex:idx_name_user" json:"created_by"` // 创建者用户名
}

// GitProtocolType 定义Git协议类型
type GitProtocolType string

const (
	GitProtocolHTTPS GitProtocolType = "https" // HTTPS协议
	GitProtocolSSH   GitProtocolType = "ssh"   // SSH协议
)

// Project 项目模型
type Project struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 基本信息
	Name        string `gorm:"not null;uniqueIndex:idx_project_name_user" json:"name"` // 项目名称
	Description string `gorm:"type:text" json:"description"`                           // 项目描述

	// Git配置
	RepoURL       string          `gorm:"not null" json:"repo_url"`                  // Git仓库地址
	Protocol      GitProtocolType `gorm:"not null;index" json:"protocol"`            // Git协议类型
	DefaultBranch string          `gorm:"default:'main'" json:"default_branch"`      // 默认分支
	CredentialID  *uint           `gorm:"index" json:"credential_id"`                // 绑定的凭据ID
	Credential    *GitCredential  `gorm:"foreignKey:CredentialID" json:"credential"` // 关联的凭据

	// 元数据
	IsActive bool       `gorm:"default:true;index" json:"is_active"` // 是否激活
	LastUsed *time.Time `json:"last_used"`                           // 最后使用时间

	// 关联用户
	CreatedBy string `gorm:"not null;index;uniqueIndex:idx_project_name_user" json:"created_by"` // 创建者用户名
}
