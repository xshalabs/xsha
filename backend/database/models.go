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
