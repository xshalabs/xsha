package services

import (
	"xsha-backend/database"
	"xsha-backend/repository"
)

type loginLogService struct {
	repo repository.LoginLogRepository
}

// NewLoginLogService 创建登录日志服务实例
func NewLoginLogService(repo repository.LoginLogRepository) LoginLogService {
	return &loginLogService{repo: repo}
}

// GetLogs 获取登录日志
func (s *loginLogService) GetLogs(username string, page, pageSize int) ([]database.LoginLog, int64, error) {
	return s.repo.GetLogs(username, page, pageSize)
}

// CleanOldLogs 清理旧日志
func (s *loginLogService) CleanOldLogs(days int) error {
	return s.repo.CleanOld(days)
}
