package handlers

import (
	"net/http"
	"sleep0-backend/database"
	"sleep0-backend/i18n"
	"sleep0-backend/middleware"
	"sleep0-backend/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// AdminOperationLogHandlers 管理员操作日志处理器结构体
type AdminOperationLogHandlers struct {
	OperationLogService services.AdminOperationLogService
}

// NewAdminOperationLogHandlers 创建管理员操作日志处理器实例
func NewAdminOperationLogHandlers(operationLogService services.AdminOperationLogService) *AdminOperationLogHandlers {
	return &AdminOperationLogHandlers{
		OperationLogService: operationLogService,
	}
}

// GetOperationLogs 获取操作日志列表
func (h *AdminOperationLogHandlers) GetOperationLogs(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	// 获取查询参数
	username := c.Query("username")
	resource := c.Query("resource")
	operationStr := c.Query("operation")
	successStr := c.Query("success")
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	page := 1
	pageSize := 20

	// 解析分页参数
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	// 解析操作类型
	var operation *database.AdminOperationType
	if operationStr != "" {
		op := database.AdminOperationType(operationStr)
		operation = &op
	}

	// 解析成功状态
	var success *bool
	if successStr != "" {
		if parsed, err := strconv.ParseBool(successStr); err == nil {
			success = &parsed
		}
	}

	// 解析时间范围
	var startTime, endTime *time.Time
	if startTimeStr != "" {
		if parsed, err := time.Parse("2006-01-02", startTimeStr); err == nil {
			startTime = &parsed
		}
	}
	if endTimeStr != "" {
		if parsed, err := time.Parse("2006-01-02", endTimeStr); err == nil {
			// 设置为当天的23:59:59
			endOfDay := parsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			endTime = &endOfDay
		}
	}

	// 使用操作日志服务获取日志
	logs, total, err := h.OperationLogService.GetLogs(username, operation, resource, success,
		startTime, endTime, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.T(lang, "common.internal_error"),
		})
		return
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	c.JSON(http.StatusOK, gin.H{
		"message":     i18n.T(lang, "common.success"),
		"logs":        logs,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// GetOperationLog 获取单个操作日志
func (h *AdminOperationLogHandlers) GetOperationLog(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	// 解析ID参数
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "common.invalid_id"),
		})
		return
	}

	// 获取操作日志
	log, err := h.OperationLogService.GetLog(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": i18n.T(lang, "common.not_found"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "common.success"),
		"log":     log,
	})
}

// GetOperationStats 获取操作统计
func (h *AdminOperationLogHandlers) GetOperationStats(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	// 获取查询参数
	username := c.Query("username")
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	// 设置默认时间范围（最近30天）
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -30)

	// 解析时间参数
	if startTimeStr != "" {
		if parsed, err := time.Parse("2006-01-02", startTimeStr); err == nil {
			startTime = parsed
		}
	}
	if endTimeStr != "" {
		if parsed, err := time.Parse("2006-01-02", endTimeStr); err == nil {
			endTime = parsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}
	}

	// 获取操作统计
	operationStats, err := h.OperationLogService.GetOperationStats(username, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.T(lang, "common.internal_error"),
		})
		return
	}

	// 获取资源统计
	resourceStats, err := h.OperationLogService.GetResourceStats(username, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.T(lang, "common.internal_error"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         i18n.T(lang, "common.success"),
		"operation_stats": operationStats,
		"resource_stats":  resourceStats,
		"start_time":      startTime.Format("2006-01-02"),
		"end_time":        endTime.Format("2006-01-02"),
	})
}

// 全局实例用于向后兼容
var globalAdminOperationLogHandlers *AdminOperationLogHandlers

// SetAdminOperationLogHandlers 设置全局管理员操作日志处理器实例
func SetAdminOperationLogHandlers(handlers *AdminOperationLogHandlers) {
	globalAdminOperationLogHandlers = handlers
}

// GetOperationLogsHandler 全局操作日志处理器（向后兼容）
func GetOperationLogsHandler(c *gin.Context) {
	globalAdminOperationLogHandlers.GetOperationLogs(c)
}

// GetOperationLogHandler 全局单个操作日志处理器（向后兼容）
func GetOperationLogHandler(c *gin.Context) {
	globalAdminOperationLogHandlers.GetOperationLog(c)
}

// GetOperationStatsHandler 全局操作统计处理器（向后兼容）
func GetOperationStatsHandler(c *gin.Context) {
	globalAdminOperationLogHandlers.GetOperationStats(c)
}
