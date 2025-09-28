package handlers

import (
	"net/http"
	"strconv"
	"xsha-backend/i18n"
	"xsha-backend/middleware"
	"xsha-backend/services"

	"github.com/gin-gonic/gin"
)

type MCPHandlers struct {
	mcpService services.MCPService
}

func NewMCPHandlers(mcpService services.MCPService) *MCPHandlers {
	return &MCPHandlers{
		mcpService: mcpService,
	}
}

// Request/Response DTOs

type CreateMCPRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Config      string `json:"config" binding:"required"`
	Enabled     *bool  `json:"enabled,omitempty"`
}

type UpdateMCPRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Config      *string `json:"config,omitempty"`
	Enabled     *bool   `json:"enabled,omitempty"`
}

type AssociateMCPRequest struct {
	MCPID uint `json:"mcp_id" binding:"required"`
}

// CRUD Handlers

// CreateMCP creates a new MCP configuration
// @Summary Create MCP
// @Description Create a new MCP configuration
// @Tags MCP
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param mcp body CreateMCPRequest true "MCP information"
// @Success 201 {object} object{message=string,mcp=object} "MCP created successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Router /mcp [post]
func (h *MCPHandlers) CreateMCP(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	admin := middleware.GetAdminFromContext(c)

	var req CreateMCPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format_with_details", err.Error()),
		})
		return
	}

	// Default enabled to true if not provided
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	mcp, err := h.mcpService.CreateMCP(req.Name, req.Description, req.Config, enabled, admin)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.MapErrorToI18nKey(err, lang),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": i18n.T(lang, "mcp.create_success"),
		"mcp":     mcp,
	})
}

// GetMCP gets a single MCP configuration
// @Summary Get MCP details
// @Description Get detailed information of an MCP configuration by ID
// @Tags MCP
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "MCP ID"
// @Success 200 {object} object{mcp=object} "MCP details"
// @Failure 404 {object} object{error=string} "MCP not found"
// @Router /mcp/{id} [get]
func (h *MCPHandlers) GetMCP(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	admin := middleware.GetAdminFromContext(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_id"),
		})
		return
	}

	mcp, err := h.mcpService.GetMCP(uint(id), admin)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": i18n.MapErrorToI18nKey(err, lang),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"mcp": mcp,
	})
}

// ListMCPs lists MCP configurations
// @Summary List MCP configurations
// @Description Get a paginated list of MCP configurations
// @Tags MCP
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param name query string false "Filter by name"
// @Param enabled query bool false "Filter by enabled status"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} object{mcps=array,total=int,page=int,page_size=int} "List of MCPs"
// @Router /mcp [get]
func (h *MCPHandlers) ListMCPs(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	admin := middleware.GetAdminFromContext(c)

	// Parse query parameters
	name := c.Query("name")
	var namePtr *string
	if name != "" {
		namePtr = &name
	}

	enabledStr := c.Query("enabled")
	var enabledPtr *bool
	if enabledStr != "" {
		if enabled, err := strconv.ParseBool(enabledStr); err == nil {
			enabledPtr = &enabled
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	mcps, total, err := h.mcpService.ListMCPs(admin, namePtr, enabledPtr, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.T(lang, "common.internal_error"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"mcps":      mcps,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// UpdateMCP updates an MCP configuration
// @Summary Update MCP
// @Description Update an MCP configuration
// @Tags MCP
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "MCP ID"
// @Param mcp body UpdateMCPRequest true "MCP update information"
// @Success 200 {object} object{message=string} "MCP updated successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Router /mcp/{id} [put]
func (h *MCPHandlers) UpdateMCP(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	admin := middleware.GetAdminFromContext(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_id"),
		})
		return
	}

	var req UpdateMCPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format_with_details", err.Error()),
		})
		return
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Config != nil {
		updates["config"] = *req.Config
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	err = h.mcpService.UpdateMCP(uint(id), updates, admin)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.MapErrorToI18nKey(err, lang),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "mcp.update_success"),
	})
}

// DeleteMCP deletes an MCP configuration
// @Summary Delete MCP
// @Description Delete an MCP configuration
// @Tags MCP
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "MCP ID"
// @Success 200 {object} object{message=string} "MCP deleted successfully"
// @Failure 404 {object} object{error=string} "MCP not found"
// @Router /mcp/{id} [delete]
func (h *MCPHandlers) DeleteMCP(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	admin := middleware.GetAdminFromContext(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_id"),
		})
		return
	}

	err = h.mcpService.DeleteMCP(uint(id), admin)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.MapErrorToI18nKey(err, lang),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "mcp.delete_success"),
	})
}

// Project Association Handlers

// GetProjectMCPs gets MCPs associated with a project
// @Summary Get project MCPs
// @Description Get all MCPs associated with a project
// @Tags MCP
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Project ID"
// @Success 200 {object} object{mcps=array} "List of MCPs"
// @Router /projects/{id}/mcp [get]
func (h *MCPHandlers) GetProjectMCPs(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idStr := c.Param("id")
	projectID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_id"),
		})
		return
	}

	mcps, err := h.mcpService.GetProjectMCPs(uint(projectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.T(lang, "common.internal_error"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"mcps": mcps,
	})
}

// AddMCPToProject associates an MCP with a project
// @Summary Associate MCP with project
// @Description Associate an MCP configuration with a project
// @Tags MCP
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Project ID"
// @Param mcp body AssociateMCPRequest true "MCP association information"
// @Success 200 {object} object{message=string} "MCP associated successfully"
// @Router /projects/{id}/mcp [post]
func (h *MCPHandlers) AddMCPToProject(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	admin := middleware.GetAdminFromContext(c)

	idStr := c.Param("id")
	projectID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_id"),
		})
		return
	}

	var req AssociateMCPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format_with_details", err.Error()),
		})
		return
	}

	err = h.mcpService.AddMCPToProject(uint(projectID), req.MCPID, admin)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.MapErrorToI18nKey(err, lang),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "mcp.associate_project_success"),
	})
}

// RemoveMCPFromProject removes an MCP association from a project
// @Summary Remove MCP from project
// @Description Remove an MCP association from a project
// @Tags MCP
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Project ID"
// @Param mcp_id path int true "MCP ID"
// @Success 200 {object} object{message=string} "MCP disassociated successfully"
// @Router /projects/{id}/mcp/{mcp_id} [delete]
func (h *MCPHandlers) RemoveMCPFromProject(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	admin := middleware.GetAdminFromContext(c)

	idStr := c.Param("id")
	projectID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_id"),
		})
		return
	}

	mcpIDStr := c.Param("mcp_id")
	mcpID, err := strconv.ParseUint(mcpIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_id"),
		})
		return
	}

	err = h.mcpService.RemoveMCPFromProject(uint(projectID), uint(mcpID), admin)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.MapErrorToI18nKey(err, lang),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "mcp.disassociate_project_success"),
	})
}

// Environment Association Handlers

// GetEnvironmentMCPs gets MCPs associated with an environment
// @Summary Get environment MCPs
// @Description Get all MCPs associated with an environment
// @Tags MCP
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Environment ID"
// @Success 200 {object} object{mcps=array} "List of MCPs"
// @Router /environments/{id}/mcp [get]
func (h *MCPHandlers) GetEnvironmentMCPs(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idStr := c.Param("id")
	envID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_id"),
		})
		return
	}

	mcps, err := h.mcpService.GetEnvironmentMCPs(uint(envID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.T(lang, "common.internal_error"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"mcps": mcps,
	})
}

// AddMCPToEnvironment associates an MCP with an environment
// @Summary Associate MCP with environment
// @Description Associate an MCP configuration with an environment
// @Tags MCP
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Environment ID"
// @Param mcp body AssociateMCPRequest true "MCP association information"
// @Success 200 {object} object{message=string} "MCP associated successfully"
// @Router /environments/{id}/mcp [post]
func (h *MCPHandlers) AddMCPToEnvironment(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	admin := middleware.GetAdminFromContext(c)

	idStr := c.Param("id")
	envID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_id"),
		})
		return
	}

	var req AssociateMCPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format_with_details", err.Error()),
		})
		return
	}

	err = h.mcpService.AddMCPToEnvironment(uint(envID), req.MCPID, admin)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.MapErrorToI18nKey(err, lang),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "mcp.associate_environment_success"),
	})
}

// RemoveMCPFromEnvironment removes an MCP association from an environment
// @Summary Remove MCP from environment
// @Description Remove an MCP association from an environment
// @Tags MCP
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Environment ID"
// @Param mcp_id path int true "MCP ID"
// @Success 200 {object} object{message=string} "MCP disassociated successfully"
// @Router /environments/{id}/mcp/{mcp_id} [delete]
func (h *MCPHandlers) RemoveMCPFromEnvironment(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	admin := middleware.GetAdminFromContext(c)

	idStr := c.Param("id")
	envID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_id"),
		})
		return
	}

	mcpIDStr := c.Param("mcp_id")
	mcpID, err := strconv.ParseUint(mcpIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_id"),
		})
		return
	}

	err = h.mcpService.RemoveMCPFromEnvironment(uint(envID), uint(mcpID), admin)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.MapErrorToI18nKey(err, lang),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "mcp.disassociate_environment_success"),
	})
}

// Additional utility handlers

// GetMCPProjects gets projects associated with an MCP
// @Summary Get MCP projects
// @Description Get all projects associated with an MCP
// @Tags MCP
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "MCP ID"
// @Success 200 {object} object{projects=array} "List of projects"
// @Router /mcp/{id}/projects [get]
func (h *MCPHandlers) GetMCPProjects(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	admin := middleware.GetAdminFromContext(c)

	idStr := c.Param("id")
	mcpID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_id"),
		})
		return
	}

	projects, err := h.mcpService.GetMCPProjects(uint(mcpID), admin)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.MapErrorToI18nKey(err, lang),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"projects": projects,
	})
}

// GetMCPEnvironments gets environments associated with an MCP
// @Summary Get MCP environments
// @Description Get all environments associated with an MCP
// @Tags MCP
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "MCP ID"
// @Success 200 {object} object{environments=array} "List of environments"
// @Router /mcp/{id}/environments [get]
func (h *MCPHandlers) GetMCPEnvironments(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	admin := middleware.GetAdminFromContext(c)

	idStr := c.Param("id")
	mcpID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_id"),
		})
		return
	}

	environments, err := h.mcpService.GetMCPEnvironments(uint(mcpID), admin)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.MapErrorToI18nKey(err, lang),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"environments": environments,
	})
}