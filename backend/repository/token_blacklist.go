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
		Where("token = ? AND expires_at > ?", token, utils.Now()).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *tokenBlacklistRepository) IsUserDeactivated(username string) (bool, error) {
	var count int64
	err := r.db.Model(&database.TokenBlacklist{}).
		Where("username = ? AND token LIKE ? AND expires_at > ? AND reason = ?", 
			username, "USER_DEACTIVATED_%", utils.Now(), "admin_deactivated").
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *tokenBlacklistRepository) CleanExpired() error {
	return r.db.Where("expires_at < ?", utils.Now()).Delete(&database.TokenBlacklist{}).Error
}

func (r *tokenBlacklistRepository) InvalidateAllUserTokens(username string, reason string) error {
	// Get all active tokens for this user that are not already blacklisted
	var existingTokens []database.TokenBlacklist
	err := r.db.Where("username = ? AND expires_at > ?", username, utils.Now()).Find(&existingTokens).Error
	if err != nil {
		return err
	}

	// For this implementation, we'll create a special blacklist entry to invalidate all tokens for the user
	// Since we don't have access to all active tokens, we'll use a different approach:
	// We'll create a marker entry with a special token format that the middleware can check
	blacklistEntry := database.TokenBlacklist{
		Token:     "USER_DEACTIVATED_" + username + "_" + utils.Now().Format(time.RFC3339),
		Username:  username,
		ExpiresAt: utils.Now().Add(365 * 24 * time.Hour), // Long expiration to cover all possible tokens
		Reason:    reason,
	}

	return r.db.Create(&blacklistEntry).Error
}
