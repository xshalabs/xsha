package services

import (
	"xsha-backend/database"
	"xsha-backend/repository"
)

type loginLogService struct {
	repo repository.LoginLogRepository
}

func NewLoginLogService(repo repository.LoginLogRepository) LoginLogService {
	return &loginLogService{repo: repo}
}

func (s *loginLogService) GetLogs(username string, page, pageSize int) ([]database.LoginLog, int64, error) {
	return s.repo.GetLogs(username, page, pageSize)
}

func (s *loginLogService) CleanOldLogs(days int) error {
	return s.repo.CleanOld(days)
}
