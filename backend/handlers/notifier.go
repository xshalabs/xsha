package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"xsha-backend/database"
	appErrors "xsha-backend/errors"
	"xsha-backend/i18n"
	"xsha-backend/middleware"
	"xsha-backend/services"

	"github.com/gin-gonic/gin"
)

type NotifierHandlers struct {
	service        services.NotifierService
	projectService services.ProjectService
}

func NewNotifierHandlers(service services.NotifierService, projectService services.ProjectService) *NotifierHandlers {
	return &NotifierHandlers{
		service:        service,
		projectService: projectService,
	}
}

// CreateNotifier creates a new notifier
func (h *NotifierHandlers) CreateNotifier(c *gin.Context) {
	admin := c.MustGet("admin").(*database.Admin)

	var req struct {
		Name        string                 `json:"name" binding:"required"`
		Description string                 `json:"description"`
		Type        database.NotifierType  `json:"type" binding:"required"`
		Config      map[string]interface{} `json:"config" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	notifier, err := h.service.CreateNotifier(req.Name, req.Description, req.Type, req.Config, admin)
	if err != nil {
		if appErr, ok := err.(*appErrors.I18nError); ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": appErr.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, notifier)
}

// GetNotifier retrieves a notifier by ID
func (h *NotifierHandlers) GetNotifier(c *gin.Context) {
	admin := c.MustGet("admin").(*database.Admin)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notifier ID"})
		return
	}

	notifier, err := h.service.GetNotifier(uint(id), admin)
	if err != nil {
		if appErr, ok := err.(*appErrors.I18nError); ok {
			c.JSON(http.StatusForbidden, gin.H{"error": appErr.Error()})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "notifier not found"})
		}
		return
	}

	c.JSON(http.StatusOK, notifier)
}

// ListNotifiers lists notifiers with filtering and pagination
func (h *NotifierHandlers) ListNotifiers(c *gin.Context) {
	admin := c.MustGet("admin").(*database.Admin)

	// Parse query parameters
	name := c.Query("name")
	var namePtr *string
	if name != "" {
		namePtr = &name
	}

	var notifierTypes []database.NotifierType
	if typeStr := c.Query("type"); typeStr != "" {
		// Support comma-separated multiple types
		typeList := strings.Split(typeStr, ",")
		for _, t := range typeList {
			trimmed := strings.TrimSpace(t)
			if trimmed != "" {
				notifierTypes = append(notifierTypes, database.NotifierType(trimmed))
			}
		}
	}

	var isEnabledPtr *bool
	if enabledStr := c.Query("is_enabled"); enabledStr != "" {
		if enabled, err := strconv.ParseBool(enabledStr); err == nil {
			isEnabledPtr = &enabled
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	notifiers, total, err := h.service.ListNotifiers(admin, namePtr, notifierTypes, isEnabledPtr, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      notifiers,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// UpdateNotifier updates a notifier
func (h *NotifierHandlers) UpdateNotifier(c *gin.Context) {
	admin := c.MustGet("admin").(*database.Admin)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notifier ID"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.service.UpdateNotifier(uint(id), updates, admin)
	if err != nil {
		if appErr, ok := err.(*appErrors.I18nError); ok {
			c.JSON(http.StatusForbidden, gin.H{"error": appErr.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "notifier updated successfully"})
}

// DeleteNotifier deletes a notifier
func (h *NotifierHandlers) DeleteNotifier(c *gin.Context) {
	admin := c.MustGet("admin").(*database.Admin)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notifier ID"})
		return
	}

	err = h.service.DeleteNotifier(uint(id), admin)
	if err != nil {
		if appErr, ok := err.(*appErrors.I18nError); ok {
			c.JSON(http.StatusForbidden, gin.H{"error": appErr.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "notifier deleted successfully"})
}

// TestNotifier tests a notifier configuration
func (h *NotifierHandlers) TestNotifier(c *gin.Context) {
	admin := c.MustGet("admin").(*database.Admin)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notifier ID"})
		return
	}

	err = h.service.TestNotifier(uint(id), admin)
	if err != nil {
		if appErr, ok := err.(*appErrors.I18nError); ok {
			c.JSON(http.StatusForbidden, gin.H{"error": appErr.Error()})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "test notification sent successfully"})
}

// GetProjectNotifiers gets notifiers associated with a project
func (h *NotifierHandlers) GetProjectNotifiers(c *gin.Context) {
	idStr := c.Param("id")
	projectID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	notifiers, err := h.service.GetProjectNotifiers(uint(projectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": notifiers})
}

// AddNotifierToProject associates a notifier with a project
func (h *NotifierHandlers) AddNotifierToProject(c *gin.Context) {
	admin := c.MustGet("admin").(*database.Admin)

	idStr := c.Param("id")
	projectID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	var req struct {
		NotifierID uint `json:"notifier_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.service.AddNotifierToProject(uint(projectID), req.NotifierID, admin)
	if err != nil {
		if appErr, ok := err.(*appErrors.I18nError); ok {
			c.JSON(http.StatusForbidden, gin.H{"error": appErr.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "notifier added to project successfully"})
}

// RemoveNotifierFromProject removes a notifier from a project
func (h *NotifierHandlers) RemoveNotifierFromProject(c *gin.Context) {
	admin := c.MustGet("admin").(*database.Admin)

	idStr := c.Param("id")
	projectID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	notifierIDStr := c.Param("notifier_id")
	notifierID, err := strconv.ParseUint(notifierIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notifier ID"})
		return
	}

	err = h.service.RemoveNotifierFromProject(uint(projectID), uint(notifierID), admin)
	if err != nil {
		if appErr, ok := err.(*appErrors.I18nError); ok {
			c.JSON(http.StatusForbidden, gin.H{"error": appErr.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "notifier removed from project successfully"})
}

// GetNotifierTypes returns available notifier types and their configuration schemas
func (h *NotifierHandlers) GetNotifierTypes(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	types := []gin.H{
		{
			"type":        "wechat_work",
			"name":        i18n.T(lang, "notifier.type.wechat_work"),
			"description": i18n.T(lang, "notifier.type.wechat_work.description"),
			"config_schema": []gin.H{
				{
					"name":        "webhook_url",
					"type":        "string",
					"required":    true,
					"description": i18n.T(lang, "notifier.config.wechat_work.webhook_url"),
				},
			},
		},
		{
			"type":        "dingtalk",
			"name":        i18n.T(lang, "notifier.type.dingtalk"),
			"description": i18n.T(lang, "notifier.type.dingtalk.description"),
			"config_schema": []gin.H{
				{
					"name":        "webhook_url",
					"type":        "string",
					"required":    true,
					"description": i18n.T(lang, "notifier.config.dingtalk.webhook_url"),
				},
				{
					"name":        "secret",
					"type":        "string",
					"required":    false,
					"description": i18n.T(lang, "notifier.config.dingtalk.secret"),
				},
			},
		},
		{
			"type":        "feishu",
			"name":        i18n.T(lang, "notifier.type.feishu"),
			"description": i18n.T(lang, "notifier.type.feishu.description"),
			"config_schema": []gin.H{
				{
					"name":        "webhook_url",
					"type":        "string",
					"required":    true,
					"description": i18n.T(lang, "notifier.config.feishu.webhook_url"),
				},
				{
					"name":        "secret",
					"type":        "string",
					"required":    false,
					"description": i18n.T(lang, "notifier.config.feishu.secret"),
				},
			},
		},
		{
			"type":        "slack",
			"name":        i18n.T(lang, "notifier.type.slack"),
			"description": i18n.T(lang, "notifier.type.slack.description"),
			"config_schema": []gin.H{
				{
					"name":        "webhook_url",
					"type":        "string",
					"required":    true,
					"description": i18n.T(lang, "notifier.config.slack.webhook_url"),
				},
			},
		},
		{
			"type":        "discord",
			"name":        i18n.T(lang, "notifier.type.discord"),
			"description": i18n.T(lang, "notifier.type.discord.description"),
			"config_schema": []gin.H{
				{
					"name":        "webhook_url",
					"type":        "string",
					"required":    true,
					"description": i18n.T(lang, "notifier.config.discord.webhook_url"),
				},
			},
		},
		{
			"type":        "webhook",
			"name":        i18n.T(lang, "notifier.type.webhook"),
			"description": i18n.T(lang, "notifier.type.webhook.description"),
			"config_schema": []gin.H{
				{
					"name":        "url",
					"type":        "string",
					"required":    true,
					"description": i18n.T(lang, "notifier.config.webhook.url"),
				},
				{
					"name":        "method",
					"type":        "string",
					"required":    false,
					"default":     "POST",
					"description": i18n.T(lang, "notifier.config.webhook.method"),
				},
				{
					"name":        "headers",
					"type":        "object",
					"required":    false,
					"description": i18n.T(lang, "notifier.config.webhook.headers"),
				},
				{
					"name":        "body_template",
					"type":        "string",
					"required":    false,
					"description": i18n.T(lang, "notifier.config.webhook.body_template"),
				},
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{"data": types})
}
