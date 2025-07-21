package services

import (
	"sleep0-backend/database"
	"sleep0-backend/repository"
	"time"
)

type adminOperationLogService struct {
	repo repository.AdminOperationLogRepository
}

// NewAdminOperationLogService 创建管理员操作日志服务实例
func NewAdminOperationLogService(repo repository.AdminOperationLogRepository) AdminOperationLogService {
	return &adminOperationLogService{repo: repo}
}

// LogOperation 记录操作日志
func (s *adminOperationLogService) LogOperation(username, operation, resource, resourceID,
	description, details string, success bool, errorMsg, ip, userAgent, method, path string) error {

	log := &database.AdminOperationLog{
		Username:    username,
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

// LogCreate 记录创建操作
func (s *adminOperationLogService) LogCreate(username, resource, resourceID, description,
	ip, userAgent, path string, success bool, errorMsg string) error {
	return s.LogOperation(username, string(database.AdminOperationCreate), resource, resourceID,
		description, "", success, errorMsg, ip, userAgent, "POST", path)
}

// LogUpdate 记录更新操作
func (s *adminOperationLogService) LogUpdate(username, resource, resourceID, description,
	ip, userAgent, path string, success bool, errorMsg string) error {
	return s.LogOperation(username, string(database.AdminOperationUpdate), resource, resourceID,
		description, "", success, errorMsg, ip, userAgent, "PUT", path)
}

// LogDelete 记录删除操作
func (s *adminOperationLogService) LogDelete(username, resource, resourceID, description,
	ip, userAgent, path string, success bool, errorMsg string) error {
	return s.LogOperation(username, string(database.AdminOperationDelete), resource, resourceID,
		description, "", success, errorMsg, ip, userAgent, "DELETE", path)
}

// LogRead 记录查询操作
func (s *adminOperationLogService) LogRead(username, resource, resourceID, description,
	ip, userAgent, path string) error {
	return s.LogOperation(username, string(database.AdminOperationRead), resource, resourceID,
		description, "", true, "", ip, userAgent, "GET", path)
}

// LogLogin 记录登录操作
func (s *adminOperationLogService) LogLogin(username, ip, userAgent string, success bool, errorMsg string) error {
	return s.LogOperation(username, string(database.AdminOperationLogin), "auth", username,
		"用户登录", "", success, errorMsg, ip, userAgent, "POST", "/api/v1/auth/login")
}

// LogLogout 记录登出操作
func (s *adminOperationLogService) LogLogout(username, ip, userAgent string, success bool, errorMsg string) error {
	return s.LogOperation(username, string(database.AdminOperationLogout), "auth", username,
		"用户登出", "", success, errorMsg, ip, userAgent, "POST", "/api/v1/auth/logout")
}

// GetLogs 获取操作日志
func (s *adminOperationLogService) GetLogs(username string, operation *database.AdminOperationType,
	resource string, success *bool, startTime, endTime *time.Time, page, pageSize int) ([]database.AdminOperationLog, int64, error) {
	return s.repo.List(username, operation, resource, success, startTime, endTime, page, pageSize)
}

// GetLog 获取单个操作日志
func (s *adminOperationLogService) GetLog(id uint) (*database.AdminOperationLog, error) {
	return s.repo.GetByID(id)
}

// GetOperationStats 获取操作统计
func (s *adminOperationLogService) GetOperationStats(username string, startTime, endTime time.Time) (map[string]int64, error) {
	return s.repo.GetOperationStats(username, startTime, endTime)
}

// GetResourceStats 获取资源统计
func (s *adminOperationLogService) GetResourceStats(username string, startTime, endTime time.Time) (map[string]int64, error) {
	return s.repo.GetResourceStats(username, startTime, endTime)
}

// CleanOldLogs 清理旧日志
func (s *adminOperationLogService) CleanOldLogs(days int) error {
	return s.repo.CleanOld(days)
}
