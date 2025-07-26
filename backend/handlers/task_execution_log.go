package handlers

import (
	"net/http"
	"sleep0-backend/i18n"
	"sleep0-backend/middleware"
	"sleep0-backend/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TaskExecutionLogHandlers struct {
	aiTaskExecutor services.AITaskExecutorService
}

// NewTaskExecutionLogHandlers 创建任务执行日志处理器
func NewTaskExecutionLogHandlers(aiTaskExecutor services.AITaskExecutorService) *TaskExecutionLogHandlers {
	return &TaskExecutionLogHandlers{
		aiTaskExecutor: aiTaskExecutor,
	}
}

// GetExecutionLog 获取执行日志
// @Summary 获取任务对话的执行日志
// @Description 根据对话ID获取AI任务执行的详细日志
// @Tags 任务执行日志
// @Accept json
// @Produce json
// @Param conversationId path int true "对话ID"
// @Success 200 {object} database.TaskExecutionLog
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /task-conversations/{conversationId}/execution-log [get]
func (h *TaskExecutionLogHandlers) GetExecutionLog(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	conversationIDStr := c.Param("conversationId")
	conversationID, err := strconv.ParseUint(conversationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	log, err := h.aiTaskExecutor.GetExecutionLog(uint(conversationID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.T(lang, "task_execution_log.not_found")})
		return
	}

	c.JSON(http.StatusOK, log)
}

// CancelExecution 取消任务执行
// @Summary 取消任务执行
// @Description 取消正在执行或待执行的AI任务
// @Tags 任务执行日志
// @Accept json
// @Produce json
// @Param conversationId path int true "对话ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /task-conversations/{conversationId}/execution/cancel [post]
func (h *TaskExecutionLogHandlers) CancelExecution(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	conversationIDStr := c.Param("conversationId")
	conversationID, err := strconv.ParseUint(conversationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	// 获取当前用户名
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.T(lang, "auth.unauthorized")})
		return
	}

	createdBy, ok := username.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(lang, "user.info_format_error")})
		return
	}

	if err := h.aiTaskExecutor.CancelExecution(uint(conversationID), createdBy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": i18n.T(lang, "task_execution_log.cancel_success")})
}

// RetryExecution 重试任务执行
// @Summary 重试任务执行
// @Description 重试失败或已取消的AI任务
// @Tags 任务执行日志
// @Accept json
// @Produce json
// @Param conversationId path int true "对话ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /task-conversations/{conversationId}/execution/retry [post]
func (h *TaskExecutionLogHandlers) RetryExecution(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	conversationIDStr := c.Param("conversationId")
	conversationID, err := strconv.ParseUint(conversationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	// 获取当前用户名
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.T(lang, "auth.unauthorized")})
		return
	}

	createdBy, ok := username.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(lang, "user.info_format_error")})
		return
	}

	if err := h.aiTaskExecutor.RetryExecution(uint(conversationID), createdBy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": i18n.T(lang, "task_execution_log.retry_success")})
}
