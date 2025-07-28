package handlers

import (
	"net/http"
	"strconv"
	"xsha-backend/database"
	"xsha-backend/i18n"
	"xsha-backend/middleware"
	"xsha-backend/services"

	"xsha-backend/utils"

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
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	RepoURL      string `json:"repo_url" binding:"required"`
	Protocol     string `json:"protocol" binding:"required,oneof=https ssh"`
	CredentialID *uint  `json:"credential_id"`
}

// UpdateProjectRequest 更新项目请求结构
// @Description 更新项目的请求参数
type UpdateProjectRequest struct {
	Name         string `json:"name" example:"更新的项目名称"`
	Description  string `json:"description" example:"更新的项目描述"`
	RepoURL      string `json:"repo_url" example:"https://github.com/user/repo.git"`
	CredentialID *uint  `json:"credential_id" example:"1"`
}

// CreateProject 创建项目
// @Summary 创建项目
// @Description 创建一个新的项目
// @Tags 项目
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param project body CreateProjectRequest true "项目信息"
// @Success 201 {object} object{id=number,message=string} "项目创建成功"
// @Failure 400 {object} object{error=string} "请求参数错误"
// @Failure 500 {object} object{error=string} "创建项目失败"
// @Router /projects [post]
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
		username.(string), req.CredentialID,
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
// @Summary 获取项目详情
// @Description 根据项目ID获取项目详细信息
// @Tags 项目
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "项目ID"
// @Success 200 {object} object{project=object} "项目详情"
// @Failure 400 {object} object{error=string} "无效的项目ID"
// @Failure 404 {object} object{error=string} "项目不存在"
// @Router /projects/{id} [get]
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
// @Summary 获取项目列表
// @Description 获取当前用户的项目列表，支持分页
// @Tags 项目
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码，默认为1"
// @Param page_size query int false "每页数量，默认为20"
// @Success 200 {object} object{projects=[]object,total=number,page=number,page_size=number} "项目列表"
// @Failure 500 {object} object{error=string} "获取项目列表失败"
// @Router /projects [get]
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
// @Summary 更新项目
// @Description 更新指定项目的信息
// @Tags 项目
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "项目ID"
// @Param project body UpdateProjectRequest true "项目更新信息"
// @Success 200 {object} object{message=string} "项目更新成功"
// @Failure 400 {object} object{error=string} "请求参数错误"
// @Failure 404 {object} object{error=string} "项目不存在"
// @Router /projects/{id} [put]
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
// @Summary 删除项目
// @Description 删除指定的项目
// @Tags 项目
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "项目ID"
// @Success 200 {object} object{message=string} "项目删除成功"
// @Failure 400 {object} object{error=string} "无效的项目ID"
// @Failure 404 {object} object{error=string} "项目不存在"
// @Router /projects/{id} [delete]
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

// GetCompatibleCredentials 获取与协议兼容的凭据列表
// @Summary 获取兼容凭据
// @Description 根据协议类型获取兼容的Git凭据列表
// @Tags 项目
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param protocol query string true "协议类型 (https/ssh)"
// @Success 200 {object} object{message=string,credentials=[]object} "获取凭据列表成功"
// @Failure 400 {object} object{error=string} "请求参数错误"
// @Router /projects/credentials [get]
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

// ParseRepositoryURLRequest 解析仓库URL请求结构
// @Description 解析Git仓库URL的请求参数
type ParseRepositoryURLRequest struct {
	RepoURL string `json:"repo_url" binding:"required" example:"https://github.com/user/repo.git"`
}

// ParseRepositoryURLResponse 解析仓库URL响应结构
// @Description 解析Git仓库URL的响应
type ParseRepositoryURLResponse struct {
	Protocol string `json:"protocol" example:"https"`
	Host     string `json:"host" example:"github.com"`
	Owner    string `json:"owner" example:"user"`
	Repo     string `json:"repo" example:"repo"`
	IsValid  bool   `json:"is_valid" example:"true"`
}

// ParseRepositoryURL 解析仓库URL
// @Summary 解析Git仓库URL
// @Description 根据输入的Git仓库URL自动检测协议类型并解析URL信息
// @Tags 项目
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ParseRepositoryURLRequest true "仓库URL"
// @Success 200 {object} object{message=string,result=ParseRepositoryURLResponse} "解析成功"
// @Failure 400 {object} object{error=string} "请求参数错误"
// @Router /projects/parse-url [post]
func (h *ProjectHandlers) ParseRepositoryURL(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	var req ParseRepositoryURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format") + ": " + err.Error(),
		})
		return
	}

	// 使用工具函数解析URL
	urlInfo := utils.ParseGitURL(req.RepoURL)

	// 构建响应
	response := ParseRepositoryURLResponse{
		Protocol: string(urlInfo.Protocol),
		Host:     urlInfo.Host,
		Owner:    urlInfo.Owner,
		Repo:     urlInfo.Repo,
		IsValid:  urlInfo.IsValid,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "common.success"),
		"result":  response,
	})
}

// FetchRepositoryBranchesRequest 获取仓库分支请求结构
// @Description 获取Git仓库分支列表的请求参数
type FetchRepositoryBranchesRequest struct {
	RepoURL      string `json:"repo_url" binding:"required" example:"https://github.com/user/repo.git"`
	CredentialID *uint  `json:"credential_id" example:"1"`
}

// FetchRepositoryBranchesResponse 获取仓库分支响应结构
// @Description 获取Git仓库分支列表的响应
type FetchRepositoryBranchesResponse struct {
	CanAccess    bool     `json:"can_access" example:"true"`
	ErrorMessage string   `json:"error_message" example:""`
	Branches     []string `json:"branches" example:"[\"main\",\"develop\",\"feature-1\"]"`
}

// FetchRepositoryBranches 获取仓库分支列表
// @Summary 获取Git仓库分支列表
// @Description 使用提供的凭据获取Git仓库的分支列表，同时验证访问权限
// @Tags 项目
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body FetchRepositoryBranchesRequest true "仓库信息"
// @Success 200 {object} object{message=string,result=FetchRepositoryBranchesResponse} "获取分支列表成功"
// @Failure 400 {object} object{error=string} "请求参数错误"
// @Failure 500 {object} object{error=string} "获取分支列表失败"
// @Router /projects/branches [post]
func (h *ProjectHandlers) FetchRepositoryBranches(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	username, _ := c.Get("username")

	var req FetchRepositoryBranchesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format") + ": " + err.Error(),
		})
		return
	}

	// 获取分支列表
	result, err := h.projectService.FetchRepositoryBranches(req.RepoURL, req.CredentialID, username.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.T(lang, "project.fetch_branches_failed") + ": " + err.Error(),
		})
		return
	}

	// 构建响应
	response := FetchRepositoryBranchesResponse{
		CanAccess:    result.CanAccess,
		ErrorMessage: result.ErrorMessage,
		Branches:     result.Branches,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "common.success"),
		"result":  response,
	})
}

// ValidateRepositoryAccessRequest 验证仓库访问请求结构
// @Description 验证Git仓库访问权限的请求参数
type ValidateRepositoryAccessRequest struct {
	RepoURL      string `json:"repo_url" binding:"required" example:"https://github.com/user/repo.git"`
	CredentialID *uint  `json:"credential_id" example:"1"`
}

// ValidateRepositoryAccess 验证仓库访问权限
// @Summary 验证Git仓库访问权限
// @Description 使用提供的凭据验证是否能够访问指定的Git仓库
// @Tags 项目
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ValidateRepositoryAccessRequest true "仓库信息"
// @Success 200 {object} object{message=string,can_access=bool} "验证成功"
// @Failure 400 {object} object{error=string} "请求参数错误或验证失败"
// @Router /projects/validate-access [post]
func (h *ProjectHandlers) ValidateRepositoryAccess(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	username, _ := c.Get("username")

	var req ValidateRepositoryAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format") + ": " + err.Error(),
		})
		return
	}

	// 验证仓库访问权限
	err := h.projectService.ValidateRepositoryAccess(req.RepoURL, req.CredentialID, username.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      i18n.T(lang, "project.access_validation_failed") + ": " + err.Error(),
			"can_access": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    i18n.T(lang, "project.access_validation_success"),
		"can_access": true,
	})
}
