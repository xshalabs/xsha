package repository

import (
	"time"
	"xsha-backend/database"
	"xsha-backend/utils"

	"gorm.io/gorm"
)

type tokenBlacklistRepository struct {
	db *gorm.DB
}

func NewTokenBlacklistRepository(db *gorm.DB) TokenBlacklistRepository {
	return &tokenBlacklistRepository{db: db}
}

func (r *tokenBlacklistRepository) Add(tokenID string, username string, expiresAt time.Time, reason string) error {
	blacklistEntry := database.TokenBlacklist{
		TokenID:   tokenID,
		Username:  username,
		ExpiresAt: expiresAt,
		Reason:    reason,
	}

	return r.db.Create(&blacklistEntry).Error
}

func (r *tokenBlacklistRepository) IsBlacklisted(tokenID string) (bool, error) {
	var count int64
	err := r.db.Model(&database.TokenBlacklist{}).
		Where("token_id = ? AND expires_at > ?", tokenID, utils.Now()).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}


func (r *tokenBlacklistRepository) CleanExpired() error {
	return r.db.Where("expires_at < ?", utils.Now()).Delete(&database.TokenBlacklist{}).Error
}

