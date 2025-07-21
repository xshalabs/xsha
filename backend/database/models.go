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
