package handlers

import (
	"net/http"
	"sleep0-backend/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

// TaskConversationHandlers 任务对话处理器结构体
type TaskConversationHandlers struct {
	conversationService services.TaskConversationService
}

// NewTaskConversationHandlers 创建任务对话处理器实例
func NewTaskConversationHandlers(conversationService services.TaskConversationService) *TaskConversationHandlers {
	return &TaskConversationHandlers{
		conversationService: conversationService,
	}
}

// CreateConversationRequest 创建对话请求结构
type CreateConversationRequest struct {
	TaskID  uint   `json:"task_id" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// UpdateConversationRequest 更新对话请求结构
type UpdateConversationRequest struct {
	Content string `json:"content"`
}

// CreateConversation 创建对话
func (h *TaskConversationHandlers) CreateConversation(c *gin.Context) {
	var req CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取当前用户
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 创建对话
	conversation, err := h.conversationService.CreateConversation(req.TaskID, req.Content, username.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Conversation created successfully",
		"data":    conversation,
	})
}

// GetConversation 获取对话详情
func (h *TaskConversationHandlers) GetConversation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	// 获取当前用户
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 获取对话
	conversation, err := h.conversationService.GetConversation(uint(id), username.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Conversation retrieved successfully",
		"data":    conversation,
	})
}

// ListConversations 获取对话列表
func (h *TaskConversationHandlers) ListConversations(c *gin.Context) {
	taskIDStr := c.Query("task_id")
	if taskIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task ID is required"})
		return
	}

	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	// 获取当前用户
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 获取对话列表
	conversations, total, err := h.conversationService.ListConversations(uint(taskID), username.(string), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Conversations retrieved successfully",
		"data": gin.H{
			"conversations": conversations,
			"total":         total,
			"page":          page,
			"page_size":     pageSize,
		},
	})
}

// UpdateConversation 更新对话
func (h *TaskConversationHandlers) UpdateConversation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	var req UpdateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取当前用户
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 构建更新数据
	updates := make(map[string]interface{})
	if req.Content != "" {
		updates["content"] = req.Content
	}

	// 更新对话
	if err := h.conversationService.UpdateConversation(uint(id), username.(string), updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conversation updated successfully"})
}

// DeleteConversation 删除对话
func (h *TaskConversationHandlers) DeleteConversation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	// 获取当前用户
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 删除对话
	if err := h.conversationService.DeleteConversation(uint(id), username.(string)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conversation deleted successfully"})
}

// GetLatestConversation 获取最新对话
func (h *TaskConversationHandlers) GetLatestConversation(c *gin.Context) {
	taskIDStr := c.Query("task_id")
	if taskIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task ID is required"})
		return
	}

	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	// 获取当前用户
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 获取最新对话
	conversation, err := h.conversationService.GetLatestConversation(uint(taskID), username.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No conversation found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Latest conversation retrieved successfully",
		"data":    conversation,
	})
}
