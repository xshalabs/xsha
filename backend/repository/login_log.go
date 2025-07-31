package repository

import (
	"time"
	"xsha-backend/database"

	"gorm.io/gorm"
)

type loginLogRepository struct {
	db *gorm.DB
}

func NewLoginLogRepository(db *gorm.DB) LoginLogRepository {
	return &loginLogRepository{db: db}
}

func (r *loginLogRepository) Add(username, ip, userAgent, reason string, success bool) error {
	loginLog := database.LoginLog{
		Username:  username,
		Success:   success,
		IP:        ip,
		UserAgent: userAgent,
		Reason:    reason,
		LoginTime: time.Now(),
	}

	return r.db.Create(&loginLog).Error
}

func (r *loginLogRepository) GetLogs(username string, page, pageSize int) ([]database.LoginLog, int64, error) {
	var logs []database.LoginLog
	var total int64

	query := r.db.Model(&database.LoginLog{})

	if username != "" {
		query = query.Where("username = ?", username)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("login_time DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *loginLogRepository) CleanOld(days int) error {
	cutoffTime := time.Now().AddDate(0, 0, -days)
	return r.db.Where("login_time < ?", cutoffTime).Delete(&database.LoginLog{}).Error
}
