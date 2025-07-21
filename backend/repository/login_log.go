package repository

import (
	"sleep0-backend/database"
	"time"

	"gorm.io/gorm"
)

type loginLogRepository struct {
	db *gorm.DB
}

// NewLoginLogRepository 创建登录日志仓库实例
func NewLoginLogRepository(db *gorm.DB) LoginLogRepository {
	return &loginLogRepository{db: db}
}

// Add 添加登录日志
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

// GetLogs 获取登录日志（支持分页和筛选）
func (r *loginLogRepository) GetLogs(username string, page, pageSize int) ([]database.LoginLog, int64, error) {
	var logs []database.LoginLog
	var total int64

	query := r.db.Model(&database.LoginLog{})

	// 按用户名筛选（可选）
	if username != "" {
		query = query.Where("username = ?", username)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("login_time DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// CleanOld 清理旧的登录日志（保留最近N天）
func (r *loginLogRepository) CleanOld(days int) error {
	cutoffTime := time.Now().AddDate(0, 0, -days)
	return r.db.Where("login_time < ?", cutoffTime).Delete(&database.LoginLog{}).Error
}
