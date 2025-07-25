package handlers

import (
	"net/http"
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
	conversationIDStr := c.Param("conversationId")
	conversationID, err := strconv.ParseUint(conversationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的对话ID"})
		return
	}

	log, err := h.aiTaskExecutor.GetExecutionLog(uint(conversationID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "执行日志不存在"})
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
	conversationIDStr := c.Param("conversationId")
	conversationID, err := strconv.ParseUint(conversationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的对话ID"})
		return
	}

	// 获取当前用户名
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无法获取用户信息"})
		return
	}

	createdBy, ok := username.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户信息格式错误"})
		return
	}

	if err := h.aiTaskExecutor.CancelExecution(uint(conversationID), createdBy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "任务执行已取消"})
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
	conversationIDStr := c.Param("conversationId")
	conversationID, err := strconv.ParseUint(conversationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的对话ID"})
		return
	}

	// 获取当前用户名
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无法获取用户信息"})
		return
	}

	createdBy, ok := username.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户信息格式错误"})
		return
	}

	if err := h.aiTaskExecutor.RetryExecution(uint(conversationID), createdBy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "任务重试执行已启动"})
}
