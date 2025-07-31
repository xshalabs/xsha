package repository

import (
	"time"
	"xsha-backend/database"

	"gorm.io/gorm"
)

type tokenBlacklistRepository struct {
	db *gorm.DB
}

func NewTokenBlacklistRepository(db *gorm.DB) TokenBlacklistRepository {
	return &tokenBlacklistRepository{db: db}
}

func (r *tokenBlacklistRepository) Add(token string, username string, expiresAt time.Time, reason string) error {
	blacklistEntry := database.TokenBlacklist{
		Token:     token,
		Username:  username,
		ExpiresAt: expiresAt,
		Reason:    reason,
	}

	return r.db.Create(&blacklistEntry).Error
}

func (r *tokenBlacklistRepository) IsBlacklisted(token string) (bool, error) {
	var count int64
	err := r.db.Model(&database.TokenBlacklist{}).
		Where("token = ? AND expires_at > ?", token, time.Now()).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *tokenBlacklistRepository) CleanExpired() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&database.TokenBlacklist{}).Error
}
