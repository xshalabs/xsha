package handlers

import (
	"net/http"
	"strconv"
	"xsha-backend/i18n"
	"xsha-backend/middleware"
	"xsha-backend/services"
	"xsha-backend/utils"

	"github.com/gin-gonic/gin"
)

// TaskConversationHandlers task conversation handler struct
type TaskConversationHandlers struct {
	conversationService services.TaskConversationService
}

// NewTaskConversationHandlers creates a task conversation handler instance
func NewTaskConversationHandlers(conversationService services.TaskConversationService) *TaskConversationHandlers {
	return &TaskConversationHandlers{
		conversationService: conversationService,
	}
}

// CreateConversationRequest request structure for creating conversations
type CreateConversationRequest struct {
	TaskID  uint   `json:"task_id" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// UpdateConversationRequest request structure for updating conversations
type UpdateConversationRequest struct {
	Content string `json:"content"`
}

// CreateConversation creates a conversation
func (h *TaskConversationHandlers) CreateConversation(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	var req CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "validation.invalid_format") + ": " + err.Error()})
		return
	}

	// Get current user
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.T(lang, "auth.unauthorized")})
		return
	}

	// Create conversation
	conversation, err := h.conversationService.CreateConversation(req.TaskID, req.Content, username.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.MapErrorToI18nKey(err, lang)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": i18n.T(lang, "taskConversation.create_success"),
		"data":    conversation,
	})
}

// GetConversation gets conversation details
func (h *TaskConversationHandlers) GetConversation(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	// Get current user
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.T(lang, "auth.unauthorized")})
		return
	}

	// 获取对话
	conversation, err := h.conversationService.GetConversation(uint(id), username.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.T(lang, "taskConversation.not_found")})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "taskConversation.get_success"),
		"data":    conversation,
	})
}

// ListConversations 获取对话列表
func (h *TaskConversationHandlers) ListConversations(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	taskIDStr := c.Query("task_id")
	if taskIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "task.project_id_required")})
		return
	}

	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	// Get current user
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.T(lang, "auth.unauthorized")})
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 获取对话列表
	conversations, total, err := h.conversationService.ListConversations(uint(taskID), username.(string), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(lang, "common.internal_error")})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "taskConversation.get_success"),
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
	lang := middleware.GetLangFromContext(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	var req UpdateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "validation.invalid_format") + ": " + err.Error()})
		return
	}

	// Get current user
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.T(lang, "auth.unauthorized")})
		return
	}

	// 构建更新数据
	updates := make(map[string]interface{})
	if req.Content != "" {
		updates["content"] = req.Content
	}

	// 更新对话
	if err := h.conversationService.UpdateConversation(uint(id), username.(string), updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.MapErrorToI18nKey(err, lang)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": i18n.T(lang, "taskConversation.update_success")})
}

// DeleteConversation 删除对话
func (h *TaskConversationHandlers) DeleteConversation(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	// Get current user
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.T(lang, "auth.unauthorized")})
		return
	}

	// 删除对话
	if err := h.conversationService.DeleteConversation(uint(id), username.(string)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.MapErrorToI18nKey(err, lang)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": i18n.T(lang, "taskConversation.update_success")})
}

// GetLatestConversation 获取最新对话
func (h *TaskConversationHandlers) GetLatestConversation(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	taskIDStr := c.Query("task_id")
	if taskIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "task.project_id_required")})
		return
	}

	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(lang, "common.invalid_id")})
		return
	}

	// Get current user
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.T(lang, "auth.unauthorized")})
		return
	}

	// 获取最新对话
	conversation, err := h.conversationService.GetLatestConversation(uint(taskID), username.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.T(lang, "taskConversation.not_found")})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "taskConversation.get_success"),
		"data":    conversation,
	})
}

// GetConversationGitDiff 获取对话Git变动
func (h *TaskConversationHandlers) GetConversationGitDiff(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	// 获取对话ID
	conversationIDStr := c.Param("id")
	conversationID, err := strconv.ParseUint(conversationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_id"),
		})
		return
	}

	// 获取查询参数
	includeContent := c.DefaultQuery("include_content", "false") == "true"

	// Get current user
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.T(lang, "auth.unauthorized")})
		return
	}

	// 获取对话Git差异
	diff, err := h.conversationService.GetConversationGitDiff(uint(conversationID), username.(string), includeContent)
	if err != nil {
		utils.Error("Failed to get conversation Git diff", "conversationID", conversationID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.T(lang, "taskConversation.git_diff_failed"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": diff,
	})
}

// GetConversationGitDiffFile 获取对话指定文件的Git变动详情
func (h *TaskConversationHandlers) GetConversationGitDiffFile(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	// 获取对话ID
	conversationIDStr := c.Param("id")
	conversationID, err := strconv.ParseUint(conversationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_id"),
		})
		return
	}

	// 获取文件路径
	filePath := c.Query("file_path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.file_path_required"),
		})
		return
	}

	// Get current user
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.T(lang, "auth.unauthorized")})
		return
	}

	// 获取文件的Git差异内容
	diffContent, err := h.conversationService.GetConversationGitDiffFile(uint(conversationID), username.(string), filePath)
	if err != nil {
		utils.Error("Failed to get conversation file Git diff", "conversationID", conversationID, "filePath", filePath, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.T(lang, "taskConversation.git_diff_file_failed"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"file_path":    filePath,
			"diff_content": diffContent,
		},
	})
}
