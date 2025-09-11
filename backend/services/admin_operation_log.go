package services

import (
	"time"
	"xsha-backend/database"
	"xsha-backend/repository"
)

type adminOperationLogService struct {
	repo repository.AdminOperationLogRepository
}

func NewAdminOperationLogService(repo repository.AdminOperationLogRepository) AdminOperationLogService {
	return &adminOperationLogService{repo: repo}
}

func (s *adminOperationLogService) LogOperation(username string, adminID *uint, operation, resource, resourceID,
	description, details string, success bool, errorMsg, ip, userAgent, method, path string) error {

	log := &database.AdminOperationLog{
		Username:    username,
		AdminID:     adminID,
		Operation:   database.AdminOperationType(operation),
		Resource:    resource,
		ResourceID:  resourceID,
		Description: description,
		Details:     details,
		Success:     success,
		ErrorMsg:    errorMsg,
		IP:          ip,
		UserAgent:   userAgent,
		Method:      method,
		Path:        path,
	}

	return s.repo.Add(log)
}

func (s *adminOperationLogService) LogCreate(username string, adminID *uint, resource, resourceID, description,
	ip, userAgent, path string, success bool, errorMsg string) error {
	return s.LogOperation(username, adminID, string(database.AdminOperationCreate), resource, resourceID,
		description, "", success, errorMsg, ip, userAgent, "POST", path)
}

func (s *adminOperationLogService) LogUpdate(username string, adminID *uint, resource, resourceID, description,
	ip, userAgent, path string, success bool, errorMsg string) error {
	return s.LogOperation(username, adminID, string(database.AdminOperationUpdate), resource, resourceID,
		description, "", success, errorMsg, ip, userAgent, "PUT", path)
}

func (s *adminOperationLogService) LogDelete(username string, adminID *uint, resource, resourceID, description,
	ip, userAgent, path string, success bool, errorMsg string) error {
	return s.LogOperation(username, adminID, string(database.AdminOperationDelete), resource, resourceID,
		description, "", success, errorMsg, ip, userAgent, "DELETE", path)
}

func (s *adminOperationLogService) LogRead(username string, adminID *uint, resource, resourceID, description,
	ip, userAgent, path string) error {
	return s.LogOperation(username, adminID, string(database.AdminOperationRead), resource, resourceID,
		description, "", true, "", ip, userAgent, "GET", path)
}

func (s *adminOperationLogService) LogLogin(username string, adminID *uint, ip, userAgent string, success bool, errorMsg string) error {
	return s.LogOperation(username, adminID, string(database.AdminOperationLogin), "auth", username,
		"user login", "", success, errorMsg, ip, userAgent, "POST", "/api/v1/auth/login")
}

func (s *adminOperationLogService) LogLogout(username string, adminID *uint, ip, userAgent string, success bool, errorMsg string) error {
	return s.LogOperation(username, adminID, string(database.AdminOperationLogout), "auth", username,
		"user logout", "", success, errorMsg, ip, userAgent, "POST", "/api/v1/auth/logout")
}

func (s *adminOperationLogService) GetLogs(username string, operation *database.AdminOperationType,
	resource string, success *bool, startTime, endTime *time.Time, page, pageSize int) ([]database.AdminOperationLog, int64, error) {
	return s.repo.List(username, operation, resource, success, startTime, endTime, page, pageSize)
}

func (s *adminOperationLogService) GetLog(id uint) (*database.AdminOperationLog, error) {
	return s.repo.GetByID(id)
}

func (s *adminOperationLogService) GetOperationStats(username string, startTime, endTime time.Time) (map[string]int64, error) {
	return s.repo.GetOperationStats(username, startTime, endTime)
}

func (s *adminOperationLogService) GetResourceStats(username string, startTime, endTime time.Time) (map[string]int64, error) {
	return s.repo.GetResourceStats(username, startTime, endTime)
}

func (s *adminOperationLogService) CleanOldLogs(days int) error {
	return s.repo.CleanOld(days)
}
