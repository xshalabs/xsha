package handlers

import (
	"net/http"
	"sleep0-backend/database"
	"sleep0-backend/i18n"
	"sleep0-backend/middleware"
	"sleep0-backend/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ProjectHandlers 项目处理器结构体
type ProjectHandlers struct {
	projectService services.ProjectService
}

// NewProjectHandlers 创建项目处理器实例
func NewProjectHandlers(projectService services.ProjectService) *ProjectHandlers {
	return &ProjectHandlers{
		projectService: projectService,
	}
}

// CreateProjectRequest 创建项目请求结构
type CreateProjectRequest struct {
	Name          string `json:"name" binding:"required"`
	Description   string `json:"description"`
	RepoURL       string `json:"repo_url" binding:"required"`
	Protocol      string `json:"protocol" binding:"required,oneof=https ssh"`
	DefaultBranch string `json:"default_branch"`
	CredentialID  *uint  `json:"credential_id"`
}

// UpdateProjectRequest 更新项目请求结构
type UpdateProjectRequest struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	RepoURL       string `json:"repo_url"`
	DefaultBranch string `json:"default_branch"`
	CredentialID  *uint  `json:"credential_id"`
}

// CreateProject 创建项目
func (h *ProjectHandlers) CreateProject(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	username, _ := c.Get("username")

	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format") + ": " + err.Error(),
		})
		return
	}

	project, err := h.projectService.CreateProject(
		req.Name, req.Description, req.RepoURL, req.Protocol,
		req.DefaultBranch, username.(string), req.CredentialID,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "project.create_failed") + ": " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": i18n.T(lang, "project.create_success"),
		"project": project,
	})
}

// GetProject 获取单个项目
func (h *ProjectHandlers) GetProject(c *gin.Context) {
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

	project, err := h.projectService.GetProject(uint(id), username.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": i18n.T(lang, "project.not_found"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"project": project,
	})
}

// ListProjects 获取项目列表
func (h *ProjectHandlers) ListProjects(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	username, _ := c.Get("username")

	// 解析查询参数
	page := 1
	pageSize := 20
	var protocol *database.GitProtocolType

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
	if proto := c.Query("protocol"); proto != "" {
		protocolValue := database.GitProtocolType(proto)
		protocol = &protocolValue
	}

	projects, total, err := h.projectService.ListProjects(username.(string), protocol, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.T(lang, "common.internal_error"),
		})
		return
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	c.JSON(http.StatusOK, gin.H{
		"message":     i18n.T(lang, "common.success"),
		"projects":    projects,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// UpdateProject 更新项目
func (h *ProjectHandlers) UpdateProject(c *gin.Context) {
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

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format") + ": " + err.Error(),
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
	if req.RepoURL != "" {
		updates["repo_url"] = req.RepoURL
	}
	if req.DefaultBranch != "" {
		updates["default_branch"] = req.DefaultBranch
	}
	// 处理凭据ID（包括设置为null的情况）
	updates["credential_id"] = req.CredentialID

	err = h.projectService.UpdateProject(uint(id), username.(string), updates)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "project.update_failed") + ": " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "project.update_success"),
	})
}

// DeleteProject 删除项目
func (h *ProjectHandlers) DeleteProject(c *gin.Context) {
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

	err = h.projectService.DeleteProject(uint(id), username.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": i18n.T(lang, "project.delete_failed") + ": " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "project.delete_success"),
	})
}

// ToggleProject 切换项目激活状态
func (h *ProjectHandlers) ToggleProject(c *gin.Context) {
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

	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format") + ": " + err.Error(),
		})
		return
	}

	err = h.projectService.ToggleProject(uint(id), username.(string), req.IsActive)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "project.toggle_failed") + ": " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "project.toggle_success"),
	})
}

// UseProject 使用项目
func (h *ProjectHandlers) UseProject(c *gin.Context) {
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

	project, err := h.projectService.UseProject(uint(id), username.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": i18n.T(lang, "project.use_failed") + ": " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "project.use_success"),
		"project": project,
	})
}

// GetCompatibleCredentials 获取与协议兼容的凭据列表
func (h *ProjectHandlers) GetCompatibleCredentials(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	username, _ := c.Get("username")

	protocol := c.Query("protocol")
	if protocol == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.required") + ": protocol",
		})
		return
	}

	protocolType := database.GitProtocolType(protocol)
	credentials, err := h.projectService.GetCompatibleCredentials(protocolType, username.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "project.get_credentials_failed") + ": " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     i18n.T(lang, "common.success"),
		"credentials": credentials,
	})
}
