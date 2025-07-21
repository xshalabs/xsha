package database

import (
	"time"

	"gorm.io/gorm"
)

// TokenBlacklist token黑名单模型
type TokenBlacklist struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Token     string         `gorm:"uniqueIndex;not null" json:"token"` // JWT token
	ExpiresAt time.Time      `gorm:"not null" json:"expires_at"`        // token过期时间
	Username  string         `gorm:"not null" json:"username"`          // 用户名
	Reason    string         `gorm:"default:'logout'" json:"reason"`    // 加入黑名单原因
}
