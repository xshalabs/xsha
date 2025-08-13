package repository

import (
	"time"
	"xsha-backend/database"
	"xsha-backend/utils"

	"gorm.io/gorm"
)

type adminOperationLogRepository struct {
	db *gorm.DB
}

func NewAdminOperationLogRepository(db *gorm.DB) AdminOperationLogRepository {
	return &adminOperationLogRepository{db: db}
}

func (r *adminOperationLogRepository) Add(log *database.AdminOperationLog) error {
	log.OperationTime = utils.Now()
	return r.db.Create(log).Error
}

func (r *adminOperationLogRepository) GetByID(id uint) (*database.AdminOperationLog, error) {
	var log database.AdminOperationLog
	if err := r.db.First(&log, id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *adminOperationLogRepository) List(username string, operation *database.AdminOperationType,
	resource string, success *bool, startTime, endTime *time.Time, page, pageSize int) ([]database.AdminOperationLog, int64, error) {

	var logs []database.AdminOperationLog
	var total int64

	query := r.db.Model(&database.AdminOperationLog{})

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

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("operation_time DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

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

func (r *adminOperationLogRepository) CleanOld(days int) error {
	cutoffTime := utils.Now().AddDate(0, 0, -days)
	return r.db.Where("operation_time < ?", cutoffTime).Delete(&database.AdminOperationLog{}).Error
}
