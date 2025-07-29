package handlers

import (
	"net/http"
	"strconv"
	"time"
	"xsha-backend/database"
	"xsha-backend/i18n"
	"xsha-backend/middleware"
	"xsha-backend/services"

	"github.com/gin-gonic/gin"
)

type AdminOperationLogHandlers struct {
	OperationLogService services.AdminOperationLogService
}

func NewAdminOperationLogHandlers(operationLogService services.AdminOperationLogService) *AdminOperationLogHandlers {
	return &AdminOperationLogHandlers{
		OperationLogService: operationLogService,
	}
}

// GetOperationLogs gets operation log list
// @Summary Get operation log list
// @Description Get administrator operation log list with multi-condition filtering and pagination
// @Tags Admin Logs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param username query string false "Username filter"
// @Param resource query string false "Resource type filter"
// @Param operation query string false "Operation type filter"
// @Param success query bool false "Operation success status filter"
// @Param start_time query string false "Start time filter (YYYY-MM-DD)"
// @Param end_time query string false "End time filter (YYYY-MM-DD)"
// @Param page query int false "Page number, default is 1"
// @Param page_size query int false "Page size, default is 20, maximum is 100"
// @Success 200 {object} object{message=string,logs=[]object,total=number,page=number,page_size=number,total_pages=number} "Operation log list"
// @Failure 500 {object} object{error=string} "Failed to get operation logs"
// @Router /admin/operation-logs [get]
func (h *AdminOperationLogHandlers) GetOperationLogs(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	username := c.Query("username")
	resource := c.Query("resource")
	operationStr := c.Query("operation")
	successStr := c.Query("success")
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	page := 1
	pageSize := 20

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

	var operation *database.AdminOperationType
	if operationStr != "" {
		op := database.AdminOperationType(operationStr)
		operation = &op
	}

	var success *bool
	if successStr != "" {
		if parsed, err := strconv.ParseBool(successStr); err == nil {
			success = &parsed
		}
	}

	var startTime, endTime *time.Time
	if startTimeStr != "" {
		if parsed, err := time.Parse("2006-01-02", startTimeStr); err == nil {
			startTime = &parsed
		}
	}
	if endTimeStr != "" {
		if parsed, err := time.Parse("2006-01-02", endTimeStr); err == nil {
			endOfDay := parsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			endTime = &endOfDay
		}
	}

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

// GetOperationLog gets a single operation log
// @Summary Get operation log details
// @Description Get detailed information of a single operation log by ID
// @Tags Admin Logs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Log ID"
// @Success 200 {object} object{message=string,log=object} "Operation log details"
// @Failure 400 {object} object{error=string} "Invalid log ID"
// @Failure 404 {object} object{error=string} "Log not found"
// @Router /admin/operation-logs/{id} [get]
func (h *AdminOperationLogHandlers) GetOperationLog(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "common.invalid_id"),
		})
		return
	}

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

// GetOperationStats gets operation statistics
// @Summary Get operation statistics
// @Description Get operation statistics information within a specified time range
// @Tags Admin Logs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param username query string false "Username filter"
// @Param start_time query string false "Start time (YYYY-MM-DD), default is 30 days ago"
// @Param end_time query string false "End time (YYYY-MM-DD), default is today"
// @Success 200 {object} object{message=string,operation_stats=object,resource_stats=object,start_time=string,end_time=string} "Operation statistics information"
// @Failure 500 {object} object{error=string} "Failed to get operation statistics"
// @Router /admin/operation-stats [get]
func (h *AdminOperationLogHandlers) GetOperationStats(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	username := c.Query("username")
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -30)

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

	operationStats, err := h.OperationLogService.GetOperationStats(username, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.T(lang, "common.internal_error"),
		})
		return
	}

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
