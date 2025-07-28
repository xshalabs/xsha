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

// DevEnvironmentHandlers 开发环境处理器结构体
type DevEnvironmentHandlers struct {
	devEnvService services.DevEnvironmentService
}

// NewDevEnvironmentHandlers 创建开发环境处理器实例
func NewDevEnvironmentHandlers(devEnvService services.DevEnvironmentService) *DevEnvironmentHandlers {
	return &DevEnvironmentHandlers{
		devEnvService: devEnvService,
	}
}

// CreateEnvironmentRequest 创建开发环境请求结构
type CreateEnvironmentRequest struct {
	Name        string            `json:"name" binding:"required"`
	Description string            `json:"description"`
	Type        string            `json:"type" binding:"required,oneof=claude_code gemini_cli opencode"`
	CPULimit    float64           `json:"cpu_limit" binding:"min=0.1,max=16"`
	MemoryLimit int64             `json:"memory_limit" binding:"min=128,max=32768"`
	EnvVars     map[string]string `json:"env_vars"`
}

// UpdateEnvironmentRequest 更新开发环境请求结构
type UpdateEnvironmentRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	CPULimit    float64           `json:"cpu_limit"`
	MemoryLimit int64             `json:"memory_limit"`
	EnvVars     map[string]string `json:"env_vars"`
}

// CreateEnvironment 创建开发环境
// @Summary 创建开发环境
// @Description 创建一个新的开发环境
// @Tags 开发环境
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param environment body CreateEnvironmentRequest true "环境信息"
// @Success 201 {object} object{message=string,environment=object} "环境创建成功"
// @Failure 400 {object} object{error=string} "请求参数错误"
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
			"error": "创建环境失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "环境创建成功",
		"environment": env,
	})
}

// GetEnvironment 获取单个开发环境
// @Summary 获取环境详情
// @Description 根据环境ID获取开发环境详细信息
// @Tags 开发环境
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "环境ID"
// @Success 200 {object} object{environment=object} "环境详情"
// @Failure 404 {object} object{error=string} "环境不存在"
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
			"error": "环境不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"environment": env,
	})
}

// ListEnvironments 获取开发环境列表
// @Summary 获取环境列表
// @Description 获取当前用户的开发环境列表，支持分页和筛选
// @Tags 开发环境
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码，默认为1"
// @Param page_size query int false "每页数量，默认为20"
// @Param type query string false "环境类型筛选"
// @Param status query string false "状态筛选"
// @Success 200 {object} object{environments=[]object,total=number} "环境列表"
// @Router /dev-environments [get]
func (h *DevEnvironmentHandlers) ListEnvironments(c *gin.Context) {
	username, _ := c.Get("username")

	// 解析查询参数
	page := 1
	pageSize := 20
	var envType *database.DevEnvironmentType

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

	environments, total, err := h.devEnvService.ListEnvironments(username.(string), envType, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取环境列表失败",
		})
		return
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	c.JSON(http.StatusOK, gin.H{
		"message":      "获取成功",
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
	username, _ := c.Get("username")

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的环境ID",
		})
		return
	}

	var req UpdateEnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 构建更新数据
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
			"error": "更新环境失败: " + err.Error(),
		})
		return
	}

	// 如果有环境变量更新
	if req.EnvVars != nil {
		err = h.devEnvService.UpdateEnvironmentVars(uint(id), username.(string), req.EnvVars)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "更新环境变量失败: " + err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "环境更新成功",
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
	username, _ := c.Get("username")

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的环境ID",
		})
		return
	}

	err = h.devEnvService.DeleteEnvironment(uint(id), username.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "删除环境失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "环境删除成功",
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
	username, _ := c.Get("username")

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的环境ID",
		})
		return
	}

	envVars, err := h.devEnvService.GetEnvironmentVars(uint(id), username.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "获取环境变量失败: " + err.Error(),
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
	username, _ := c.Get("username")

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的环境ID",
		})
		return
	}

	var envVars map[string]string
	if err := c.ShouldBindJSON(&envVars); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: " + err.Error(),
		})
		return
	}

	err = h.devEnvService.UpdateEnvironmentVars(uint(id), username.(string), envVars)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "更新环境变量失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "环境变量更新成功",
	})
}
