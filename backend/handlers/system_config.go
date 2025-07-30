package handlers

import (
	"net/http"
	"strconv"
	"xsha-backend/i18n"
	"xsha-backend/middleware"
	"xsha-backend/services"

	"github.com/gin-gonic/gin"
)

type SystemConfigHandlers struct {
	configService services.SystemConfigService
}

func NewSystemConfigHandlers(configService services.SystemConfigService) *SystemConfigHandlers {
	return &SystemConfigHandlers{
		configService: configService,
	}
}

type UpdateConfigRequest struct {
	ConfigValue string `json:"config_value"`
	Description string `json:"description"`
	Category    string `json:"category"`
	IsEditable  *bool  `json:"is_editable"`
}

type UpdateDevEnvironmentTypesRequest struct {
	EnvTypes []map[string]interface{} `json:"env_types" binding:"required"`
}

// GetConfig gets a system configuration by ID
// @Summary Get configuration
// @Description Get system configuration by ID
// @Tags System Configuration
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Configuration ID"
// @Success 200 {object} object{config=object} "Configuration information"
// @Failure 404 {object} object{error=string} "Configuration not found"
// @Router /system-configs/{id} [get]
func (h *SystemConfigHandlers) GetConfig(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_id"),
		})
		return
	}

	config, err := h.configService.GetConfig(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": i18n.T(lang, "system_config.not_found"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"config": config,
	})
}

// ListConfigs gets system configuration list
// @Summary Get configuration list
// @Description Get system configuration list with pagination and filtering
// @Tags System Configuration
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number, default is 1"
// @Param page_size query int false "Page size, default is 20"
// @Param category query string false "Configuration category filter"
// @Success 200 {object} object{configs=[]object,total=number} "Configuration list"
// @Router /system-configs [get]
func (h *SystemConfigHandlers) ListConfigs(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	page := 1
	pageSize := 20
	category := c.Query("category")

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

	configs, total, err := h.configService.ListConfigs(category, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.T(lang, "system_config.list_failed"),
		})
		return
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	c.JSON(http.StatusOK, gin.H{
		"message":     i18n.T(lang, "system_config.list_success"),
		"configs":     configs,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// UpdateConfig updates a system configuration
// @Summary Update configuration
// @Description Update system configuration information
// @Tags System Configuration
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Configuration ID"
// @Param config body UpdateConfigRequest true "Configuration information"
// @Success 200 {object} object{message=string} "Configuration updated successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Router /system-configs/{id} [put]
func (h *SystemConfigHandlers) UpdateConfig(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_id"),
		})
		return
	}

	var req UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format") + ": " + err.Error(),
		})
		return
	}

	updates := make(map[string]interface{})
	if req.ConfigValue != "" {
		updates["config_value"] = req.ConfigValue
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Category != "" {
		updates["category"] = req.Category
	}
	if req.IsEditable != nil {
		updates["is_editable"] = *req.IsEditable
	}

	if err := h.configService.UpdateConfig(uint(id), updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.T(lang, "system_config.update_failed") + ": " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "system_config.update_success"),
	})
}

// GetDevEnvironmentTypes gets available development environment types
// @Summary Get development environment types
// @Description Get available development environment types from system configuration
// @Tags System Configuration
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object{env_types=[]object} "Environment types"
// @Router /system-configs/dev-environment-types [get]
func (h *SystemConfigHandlers) GetDevEnvironmentTypes(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	envTypes, err := h.configService.GetDevEnvironmentTypes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.T(lang, "system_config.get_dev_env_types_failed") + ": " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"env_types": envTypes,
	})
}

// UpdateDevEnvironmentTypes updates available development environment types
// @Summary Update development environment types
// @Description Update available development environment types in system configuration
// @Tags System Configuration
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param env_types body UpdateDevEnvironmentTypesRequest true "Environment types"
// @Success 200 {object} object{message=string} "Environment types updated successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Router /system-configs/dev-environment-types [put]
func (h *SystemConfigHandlers) UpdateDevEnvironmentTypes(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	var req UpdateDevEnvironmentTypesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format") + ": " + err.Error(),
		})
		return
	}

	if err := h.configService.UpdateDevEnvironmentTypes(req.EnvTypes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.T(lang, "system_config.update_dev_env_types_failed") + ": " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "system_config.update_dev_env_types_success"),
	})
}
