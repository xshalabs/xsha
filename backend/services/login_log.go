package services

import (
	"xsha-backend/database"
	"xsha-backend/repository"
)

type loginLogService struct {
	repo repository.LoginLogRepository
}

// NewLoginLogService creates a login log service instance
func NewLoginLogService(repo repository.LoginLogRepository) LoginLogService {
	return &loginLogService{repo: repo}
}

// GetLogs gets login logs
func (s *loginLogService) GetLogs(username string, page, pageSize int) ([]database.LoginLog, int64, error) {
	return s.repo.GetLogs(username, page, pageSize)
}

// CleanOldLogs cleans old logs
func (s *loginLogService) CleanOldLogs(days int) error {
	return s.repo.CleanOld(days)
}
