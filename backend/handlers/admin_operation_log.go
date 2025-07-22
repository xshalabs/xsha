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
// @Summary 获取操作日志列表
// @Description 获取管理员操作日志列表，支持多条件筛选和分页
// @Tags 管理员日志
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param username query string false "用户名筛选"
// @Param resource query string false "资源类型筛选"
// @Param operation query string false "操作类型筛选"
// @Param success query bool false "操作成功状态筛选"
// @Param start_time query string false "开始时间筛选 (YYYY-MM-DD)"
// @Param end_time query string false "结束时间筛选 (YYYY-MM-DD)"
// @Param page query int false "页码，默认为1"
// @Param page_size query int false "每页数量，默认为20，最大100"
// @Success 200 {object} object{message=string,logs=[]object,total=number,page=number,page_size=number,total_pages=number} "操作日志列表"
// @Failure 500 {object} object{error=string} "获取操作日志失败"
// @Router /admin/operation-logs [get]
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
// @Summary 获取操作日志详情
// @Description 根据ID获取单个操作日志的详细信息
// @Tags 管理员日志
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "日志ID"
// @Success 200 {object} object{message=string,log=object} "操作日志详情"
// @Failure 400 {object} object{error=string} "无效的日志ID"
// @Failure 404 {object} object{error=string} "日志不存在"
// @Router /admin/operation-logs/{id} [get]
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
// @Summary 获取操作统计
// @Description 获取指定时间范围内的操作统计信息
// @Tags 管理员日志
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param username query string false "用户名筛选"
// @Param start_time query string false "开始时间 (YYYY-MM-DD)，默认30天前"
// @Param end_time query string false "结束时间 (YYYY-MM-DD)，默认今天"
// @Success 200 {object} object{message=string,operation_stats=object,resource_stats=object,start_time=string,end_time=string} "操作统计信息"
// @Failure 500 {object} object{error=string} "获取操作统计失败"
// @Router /admin/operation-stats [get]
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
