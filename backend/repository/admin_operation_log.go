package repository

import (
	"sleep0-backend/database"
	"time"

	"gorm.io/gorm"
)

type adminOperationLogRepository struct {
	db *gorm.DB
}

// NewAdminOperationLogRepository 创建管理员操作日志仓库实例
func NewAdminOperationLogRepository(db *gorm.DB) AdminOperationLogRepository {
	return &adminOperationLogRepository{db: db}
}

// Add 添加操作日志
func (r *adminOperationLogRepository) Add(log *database.AdminOperationLog) error {
	log.OperationTime = time.Now()
	return r.db.Create(log).Error
}

// GetByID 根据ID获取操作日志
func (r *adminOperationLogRepository) GetByID(id uint) (*database.AdminOperationLog, error) {
	var log database.AdminOperationLog
	if err := r.db.First(&log, id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

// List 获取操作日志列表（支持多种筛选条件和分页）
func (r *adminOperationLogRepository) List(username string, operation *database.AdminOperationType,
	resource string, success *bool, startTime, endTime *time.Time, page, pageSize int) ([]database.AdminOperationLog, int64, error) {

	var logs []database.AdminOperationLog
	var total int64

	query := r.db.Model(&database.AdminOperationLog{})

	// 条件筛选
	if username != "" {
		query = query.Where("username = ?", username)
	}
	if operation != nil {
		query = query.Where("operation = ?", *operation)
	}
	if resource != "" {
		query = query.Where("resource = ?", resource)
	}
	if success != nil {
		query = query.Where("success = ?", *success)
	}
	if startTime != nil {
		query = query.Where("operation_time >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("operation_time <= ?", *endTime)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("operation_time DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetOperationStats 获取操作类型统计
func (r *adminOperationLogRepository) GetOperationStats(username string, startTime, endTime time.Time) (map[string]int64, error) {
	var results []struct {
		Operation string `json:"operation"`
		Count     int64  `json:"count"`
	}

	query := r.db.Model(&database.AdminOperationLog{}).
		Select("operation, COUNT(*) as count").
		Where("operation_time BETWEEN ? AND ?", startTime, endTime).
		Group("operation")

	if username != "" {
		query = query.Where("username = ?", username)
	}

	if err := query.Scan(&results).Error; err != nil {
		return nil, err
	}

	stats := make(map[string]int64)
	for _, result := range results {
		stats[result.Operation] = result.Count
	}

	return stats, nil
}

// GetResourceStats 获取资源类型统计
func (r *adminOperationLogRepository) GetResourceStats(username string, startTime, endTime time.Time) (map[string]int64, error) {
	var results []struct {
		Resource string `json:"resource"`
		Count    int64  `json:"count"`
	}

	query := r.db.Model(&database.AdminOperationLog{}).
		Select("resource, COUNT(*) as count").
		Where("operation_time BETWEEN ? AND ?", startTime, endTime).
		Group("resource")

	if username != "" {
		query = query.Where("username = ?", username)
	}

	if err := query.Scan(&results).Error; err != nil {
		return nil, err
	}

	stats := make(map[string]int64)
	for _, result := range results {
		stats[result.Resource] = result.Count
	}

	return stats, nil
}

// CleanOld 清理旧的操作日志（保留最近N天）
func (r *adminOperationLogRepository) CleanOld(days int) error {
	cutoffTime := time.Now().AddDate(0, 0, -days)
	return r.db.Where("operation_time < ?", cutoffTime).Delete(&database.AdminOperationLog{}).Error
}
