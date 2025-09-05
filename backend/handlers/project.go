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

type ProjectHandlers struct {
	projectService services.ProjectService
}

func NewProjectHandlers(projectService services.ProjectService) *ProjectHandlers {
	return &ProjectHandlers{
		projectService: projectService,
	}
}

// @Description Create project request
type CreateProjectRequest struct {
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	SystemPrompt string `json:"system_prompt"`
	RepoURL      string `json:"repo_url" binding:"required"`
	Protocol     string `json:"protocol" binding:"omitempty,oneof=https ssh"`
	CredentialID *uint  `json:"credential_id"`
}

// @Description Update project request
type UpdateProjectRequest struct {
	Name         string `json:"name" example:"Updated project name"`
	Description  string `json:"description" example:"Updated project description"`
	SystemPrompt string `json:"system_prompt" example:"Custom system prompt"`
	RepoURL      string `json:"repo_url" example:"https://github.com/user/repo.git"`
	CredentialID *uint  `json:"credential_id" example:"1"`
}

// CreateProject creates project
// @Summary Create project
// @Description Create a new project
// @Tags Project
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param project body CreateProjectRequest true "Project information"
// @Success 201 {object} object{id=number,message=string} "Project created successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Failure 500 {object} object{error=string} "Project creation failed"
// @Router /projects [post]
func (h *ProjectHandlers) CreateProject(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.T(lang, "auth.unauthorized"),
		})
		return
	}

	adminIDVal, adminExists := c.Get("admin_id")
	if !adminExists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.T(lang, "auth.unauthorized"),
		})
		return
	}
	adminID := adminIDVal.(uint)

	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format_with_details", err.Error()),
		})
		return
	}

	// Auto-detect protocol from repository URL
	urlInfo := utils.ParseGitURL(req.RepoURL)
	if urlInfo.IsValid {
		// Use detected protocol instead of user input
		req.Protocol = string(urlInfo.Protocol)
	}

	project, err := h.projectService.CreateProject(
		req.Name, req.Description, req.SystemPrompt, req.RepoURL, req.Protocol,
		req.CredentialID, &adminID, username.(string),
	)
	if err != nil {
		helper := i18n.NewHelper(lang)
		helper.ErrorResponseFromError(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": i18n.T(lang, "project.create_success"),
		"project": project,
	})
}

// GetProject gets single project
// @Summary Get project details
// @Description Get project detailed information by project ID
// @Tags Project
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Project ID"
// @Success 200 {object} object{project=object} "Project details"
// @Failure 400 {object} object{error=string} "Invalid project ID"
// @Failure 404 {object} object{error=string} "Project not found"
// @Router /projects/{id} [get]
func (h *ProjectHandlers) GetProject(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format"),
		})
		return
	}

	project, err := h.projectService.GetProject(uint(id))
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

// ListProjects gets project list
// @Summary Get project list
// @Description Get current user's project list with pagination, filtering and sorting
// @Tags Project
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param name query string false "Project name filter (fuzzy matching)"
// @Param protocol query string false "Protocol type filter (https/ssh)"
// @Param page query int false "Page number, defaults to 1"
// @Param page_size query int false "Page size, defaults to 20"
// @Param sort_by query string false "Sort field: name, task_count, created_at (defaults to created_at)"
// @Param sort_direction query string false "Sort direction: asc, desc (defaults to desc)"
// @Success 200 {object} object{projects=[]object,total=number,page=number,page_size=number} "Project list"
// @Failure 500 {object} object{error=string} "Failed to get project list"
// @Router /projects [get]
func (h *ProjectHandlers) ListProjects(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	// Parse query parameters
	page := 1
	pageSize := 20
	var protocol *database.GitProtocolType
	name := c.Query("name")
	sortBy := c.Query("sort_by")
	sortDirection := c.Query("sort_direction")

	// Default sort values
	if sortBy == "" {
		sortBy = "created_at"
	}
	if sortDirection == "" {
		sortDirection = "desc"
	}

	// Validate sort_by field
	validSortFields := map[string]bool{
		"name":       true,
		"task_count": true,
		"created_at": true,
	}
	if !validSortFields[sortBy] {
		sortBy = "created_at"
	}

	// Validate sort_direction
	if sortDirection != "asc" && sortDirection != "desc" {
		sortDirection = "desc"
	}

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

	admin := middleware.GetAdminFromContext(c)

	var projects interface{}
	var total int64
	var err error
	if admin.Role == database.AdminRoleSuperAdmin {
		projects, total, err = h.projectService.ListProjectsWithTaskCount(name, protocol, sortBy, sortDirection, page, pageSize)
	} else {
		projects, total, err = h.projectService.ListProjectsByAdminAccess(admin.ID, name, protocol, sortBy, sortDirection, page, pageSize)
	}
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

// UpdateProject updates project
// @Summary Update project
// @Description Update specified project information
// @Tags Project
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Project ID"
// @Param project body UpdateProjectRequest true "Project update information"
// @Success 200 {object} object{message=string} "Project updated successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Failure 404 {object} object{error=string} "Project not found"
// @Router /projects/{id} [put]
func (h *ProjectHandlers) UpdateProject(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

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
			"error": i18n.T(lang, "validation.invalid_format_with_details", err.Error()),
		})
		return
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}

	updates["description"] = req.Description
	updates["system_prompt"] = req.SystemPrompt

	if req.RepoURL != "" {
		updates["repo_url"] = req.RepoURL
	}

	updates["credential_id"] = req.CredentialID

	err = h.projectService.UpdateProject(uint(id), updates)
	if err != nil {
		helper := i18n.NewHelper(lang)
		helper.ErrorResponseFromError(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "project.update_success"),
	})
}

// DeleteProject deletes project
// @Summary Delete project
// @Description Delete specified project
// @Tags Project
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Project ID"
// @Success 200 {object} object{message=string} "Project deleted successfully"
// @Failure 400 {object} object{error=string} "Invalid project ID"
// @Failure 404 {object} object{error=string} "Project not found"
// @Router /projects/{id} [delete]
func (h *ProjectHandlers) DeleteProject(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format"),
		})
		return
	}

	err = h.projectService.DeleteProject(uint(id))
	if err != nil {
		helper := i18n.NewHelper(lang)
		helper.ErrorResponseFromError(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "project.delete_success"),
	})
}

// GetCompatibleCredentials gets credential list compatible with protocol
// @Summary Get compatible credentials
// @Description Get Git credential list compatible with protocol type, auto-detected from repo URL or specified protocol
// @Tags Project
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param repo_url query string false "Repository URL for protocol auto-detection"
// @Param protocol query string false "Protocol type (https/ssh) - fallback if repo_url not provided"
// @Success 200 {object} object{message=string,credentials=[]object} "Get credential list successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Router /projects/credentials [get]
func (h *ProjectHandlers) GetCompatibleCredentials(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	repoURL := c.Query("repo_url")
	protocol := c.Query("protocol")

	var protocolType database.GitProtocolType

	if repoURL != "" {
		// Auto-detect protocol from repository URL
		urlInfo := utils.ParseGitURL(repoURL)
		if urlInfo.IsValid {
			protocolType = database.GitProtocolType(urlInfo.Protocol)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": i18n.T(lang, "validation.invalid_repo_url"),
			})
			return
		}
	} else if protocol != "" {
		// Use provided protocol as fallback
		protocolType = database.GitProtocolType(protocol)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.required_repo_url_or_protocol"),
		})
		return
	}

	admin := middleware.GetAdminFromContext(c)

	credentials, err := h.projectService.GetCompatibleCredentials(protocolType, admin)
	if err != nil {
		helper := i18n.NewHelper(lang)
		helper.ErrorResponseFromError(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     i18n.T(lang, "common.success"),
		"credentials": credentials,
	})
}


// @Description Request parameters for fetching Git repository branch list
type FetchRepositoryBranchesRequest struct {
	RepoURL      string `json:"repo_url" binding:"required" example:"https://github.com/user/repo.git"`
	CredentialID *uint  `json:"credential_id" example:"1"`
}

// @Description Response for fetching Git repository branch list
type FetchRepositoryBranchesResponse struct {
	CanAccess    bool     `json:"can_access" example:"true"`
	ErrorMessage string   `json:"error_message" example:""`
	Branches     []string `json:"branches" example:"main,develop,feature-1"`
}

// FetchRepositoryBranches fetches repository branch list
// @Summary Fetch Git repository branch list
// @Description Fetch Git repository branch list using provided credentials and verify access permissions
// @Tags Project
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body FetchRepositoryBranchesRequest true "Repository information"
// @Success 200 {object} object{message=string,result=FetchRepositoryBranchesResponse} "Fetch branch list successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Failure 500 {object} object{error=string} "Failed to fetch branch list"
// @Router /projects/branches [post]
func (h *ProjectHandlers) FetchRepositoryBranches(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	var req FetchRepositoryBranchesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format_with_details", err.Error()),
		})
		return
	}

	result, err := h.projectService.FetchRepositoryBranches(req.RepoURL, req.CredentialID)
	if err != nil {
		helper := i18n.NewHelper(lang)
		helper.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

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

// @Description Add admin to project request
type AddAdminToProjectRequest struct {
	AdminID uint `json:"admin_id" binding:"required"`
}

// @Description Project admins response
type ProjectAdminsResponse struct {
	Admins []database.AdminListResponse `json:"admins"`
}

// GetProjectAdmins gets all admins for a specific project
// @Summary Get project admins
// @Description Get all admins that have access to a specific project
// @Tags Project
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Project ID"
// @Success 200 {object} ProjectAdminsResponse "Project admins"
// @Failure 404 {object} object{error=string} "Project not found"
// @Router /projects/{id}/admins [get]
func (h *ProjectHandlers) GetProjectAdmins(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "project.invalid_id"),
		})
		return
	}

	admins, err := h.projectService.GetProjectAdmins(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": i18n.MapErrorToI18nKey(err, lang),
		})
		return
	}

	// Transform to list response with minimal avatar data
	adminResponses := database.ToAdminListResponses(admins)
	c.JSON(http.StatusOK, ProjectAdminsResponse{
		Admins: adminResponses,
	})
}

// AddAdminToProject adds an admin to the project
// @Summary Add admin to project
// @Description Add an admin to the project's admin list
// @Tags Project
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Project ID"
// @Param admin body AddAdminToProjectRequest true "Admin information"
// @Success 200 {object} object{message=string} "Admin added successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Failure 404 {object} object{error=string} "Project not found"
// @Router /projects/{id}/admins [post]
func (h *ProjectHandlers) AddAdminToProject(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idStr := c.Param("id")
	projectID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "project.invalid_id"),
		})
		return
	}

	var req AddAdminToProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format_with_details", err.Error()),
		})
		return
	}

	err = h.projectService.AddAdminToProject(uint(projectID), req.AdminID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.MapErrorToI18nKey(err, lang),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "project.admin_added_success"),
	})
}

// RemoveAdminFromProject removes an admin from the project
// @Summary Remove admin from project
// @Description Remove an admin from the project's admin list
// @Tags Project
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Project ID"
// @Param admin_id path int true "Admin ID"
// @Success 200 {object} object{message=string} "Admin removed successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Failure 404 {object} object{error=string} "Project not found"
// @Router /projects/{id}/admins/{admin_id} [delete]
func (h *ProjectHandlers) RemoveAdminFromProject(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idStr := c.Param("id")
	projectID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "project.invalid_id"),
		})
		return
	}

	adminIDStr := c.Param("admin_id")
	adminID, err := strconv.ParseUint(adminIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "admin.invalid_id"),
		})
		return
	}

	err = h.projectService.RemoveAdminFromProject(uint(projectID), uint(adminID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.MapErrorToI18nKey(err, lang),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "project.admin_removed_success"),
	})
}


