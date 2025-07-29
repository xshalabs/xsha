package handlers

import (
	"net/http"
	"strconv"
	"xsha-backend/database"
	"xsha-backend/i18n"
	"xsha-backend/middleware"
	"xsha-backend/services"

	"github.com/gin-gonic/gin"
)

type DevEnvironmentHandlers struct {
	devEnvService services.DevEnvironmentService
}

func NewDevEnvironmentHandlers(devEnvService services.DevEnvironmentService) *DevEnvironmentHandlers {
	return &DevEnvironmentHandlers{
		devEnvService: devEnvService,
	}
}

type CreateEnvironmentRequest struct {
	Name        string            `json:"name" binding:"required"`
	Description string            `json:"description"`
	Type        string            `json:"type" binding:"required,oneof=claude_code gemini_cli opencode"`
	CPULimit    float64           `json:"cpu_limit" binding:"min=0.1,max=16"`
	MemoryLimit int64             `json:"memory_limit" binding:"min=128,max=32768"`
	EnvVars     map[string]string `json:"env_vars"`
}

type UpdateEnvironmentRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	CPULimit    float64           `json:"cpu_limit"`
	MemoryLimit int64             `json:"memory_limit"`
	EnvVars     map[string]string `json:"env_vars"`
}

// CreateEnvironment creates a development environment
// @Summary Create development environment
// @Description Create a new development environment
// @Tags Development Environment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param environment body CreateEnvironmentRequest true "Environment information"
// @Success 201 {object} object{message=string,environment=object} "Environment created successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Router /dev-environments [post]
func (h *DevEnvironmentHandlers) CreateEnvironment(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	username, _ := c.Get("username")

	var req CreateEnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format") + ": " + err.Error(),
		})
		return
	}

	if req.EnvVars == nil {
		req.EnvVars = make(map[string]string)
	}

	env, err := h.devEnvService.CreateEnvironment(
		req.Name, req.Description, req.Type, username.(string),
		req.CPULimit, req.MemoryLimit, req.EnvVars,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "dev_environment.create_failed") + ": " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     i18n.T(lang, "dev_environment.create_success"),
		"environment": env,
	})
}

// GetEnvironment gets a single development environment
// @Summary Get environment details
// @Description Get detailed information of a development environment by ID
// @Tags Development Environment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Environment ID"
// @Success 200 {object} object{environment=object} "Environment details"
// @Failure 404 {object} object{error=string} "Environment not found"
// @Router /dev-environments/{id} [get]
func (h *DevEnvironmentHandlers) GetEnvironment(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	username, _ := c.Get("username")

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format"),
		})
		return
	}

	env, err := h.devEnvService.GetEnvironment(uint(id), username.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": i18n.T(lang, "dev_environment.not_found"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"environment": env,
	})
}

// ListEnvironments gets development environment list
// @Summary Get environment list
// @Description Get current user's development environment list with pagination and filtering
// @Tags Development Environment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number, default is 1"
// @Param page_size query int false "Page size, default is 10"
// @Param type query string false "Environment type filter"
// @Param name query string false "Environment name filter"
// @Success 200 {object} object{environments=[]object,total=number} "Environment list"
// @Router /dev-environments [get]
func (h *DevEnvironmentHandlers) ListEnvironments(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	username, _ := c.Get("username")

	page := 1
	pageSize := 10
	var envType *database.DevEnvironmentType
	var name *string

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
	if t := c.Query("type"); t != "" {
		typeValue := database.DevEnvironmentType(t)
		envType = &typeValue
	}
	if n := c.Query("name"); n != "" {
		name = &n
	}

	environments, total, err := h.devEnvService.ListEnvironments(username.(string), envType, name, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.T(lang, "dev_environment.list_failed"),
		})
		return
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	c.JSON(http.StatusOK, gin.H{
		"message":      i18n.T(lang, "dev_environment.list_success"),
		"environments": environments,
		"total":        total,
		"page":         page,
		"page_size":    pageSize,
		"total_pages":  totalPages,
	})
}

// UpdateEnvironment 更新开发环境
// @Summary 更新环境
// @Description 更新指定开发环境的信息
// @Tags 开发环境
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "环境ID"
// @Param environment body UpdateEnvironmentRequest true "环境更新信息"
// @Success 200 {object} object{message=string} "环境更新成功"
// @Failure 400 {object} object{error=string} "请求参数错误"
// @Router /dev-environments/{id} [put]
func (h *DevEnvironmentHandlers) UpdateEnvironment(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	username, _ := c.Get("username")

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "dev_environment.invalid_id"),
		})
		return
	}

	var req UpdateEnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "dev_environment.invalid_request") + ": " + err.Error(),
		})
		return
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.CPULimit > 0 {
		updates["cpu_limit"] = req.CPULimit
	}
	if req.MemoryLimit > 0 {
		updates["memory_limit"] = req.MemoryLimit
	}

	err = h.devEnvService.UpdateEnvironment(uint(id), username.(string), updates)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "dev_environment.update_failed") + ": " + err.Error(),
		})
		return
	}

	if req.EnvVars != nil {
		err = h.devEnvService.UpdateEnvironmentVars(uint(id), username.(string), req.EnvVars)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": i18n.T(lang, "dev_environment.env_vars_update_failed") + ": " + err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "dev_environment.update_success"),
	})
}

// DeleteEnvironment 删除开发环境
// @Summary 删除环境
// @Description 删除指定的开发环境
// @Tags 开发环境
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "环境ID"
// @Success 200 {object} object{message=string} "环境删除成功"
// @Failure 400 {object} object{error=string} "删除失败"
// @Router /dev-environments/{id} [delete]
func (h *DevEnvironmentHandlers) DeleteEnvironment(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	username, _ := c.Get("username")

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "dev_environment.invalid_id"),
		})
		return
	}

	err = h.devEnvService.DeleteEnvironment(uint(id), username.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "dev_environment.delete_failed") + ": " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "dev_environment.delete_success"),
	})
}

// GetEnvironmentVars 获取环境变量
// @Summary 获取环境变量
// @Description 获取指定环境的环境变量
// @Tags 开发环境
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "环境ID"
// @Success 200 {object} object{env_vars=object} "环境变量"
// @Failure 400 {object} object{error=string} "获取失败"
// @Router /dev-environments/{id}/env-vars [get]
func (h *DevEnvironmentHandlers) GetEnvironmentVars(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	username, _ := c.Get("username")

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "dev_environment.invalid_id"),
		})
		return
	}

	envVars, err := h.devEnvService.GetEnvironmentVars(uint(id), username.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "dev_environment.env_vars_get_failed") + ": " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"env_vars": envVars,
	})
}

// UpdateEnvironmentVars 更新环境变量
// @Summary 更新环境变量
// @Description 更新指定环境的环境变量
// @Tags 开发环境
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "环境ID"
// @Param env_vars body map[string]string true "环境变量"
// @Success 200 {object} object{message=string} "更新成功"
// @Failure 400 {object} object{error=string} "更新失败"
// @Router /dev-environments/{id}/env-vars [put]
func (h *DevEnvironmentHandlers) UpdateEnvironmentVars(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	username, _ := c.Get("username")

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "dev_environment.invalid_id"),
		})
		return
	}

	var envVars map[string]string
	if err := c.ShouldBindJSON(&envVars); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "dev_environment.invalid_request") + ": " + err.Error(),
		})
		return
	}

	err = h.devEnvService.UpdateEnvironmentVars(uint(id), username.(string), envVars)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "dev_environment.env_vars_update_failed") + ": " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "dev_environment.env_vars_update_success"),
	})
}
