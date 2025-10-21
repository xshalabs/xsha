package handlers

import (
	"net/http"
	"strconv"
	"xsha-backend/i18n"
	"xsha-backend/middleware"
	"xsha-backend/services"

	"github.com/gin-gonic/gin"
)

type ProviderHandlers struct {
	providerService services.ProviderService
}

func NewProviderHandlers(providerService services.ProviderService) *ProviderHandlers {
	return &ProviderHandlers{
		providerService: providerService,
	}
}

// Request/Response structures

type CreateProviderRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Type        string `json:"type" binding:"required"`
	Config      string `json:"config" binding:"required"`
}

type UpdateProviderRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Config      *string `json:"config"`
}

// GetProviderTypes gets available provider types
// @Summary Get available provider types
// @Description Get list of available provider types
// @Tags Providers
// @Accept json
// @Produce json
// @Success 200 {object} object{types=[]string} "Provider types retrieved successfully"
// @Router /providers/types [get]
func (h *ProviderHandlers) GetProviderTypes(c *gin.Context) {
	types := h.providerService.GetProviderTypes()
	c.JSON(http.StatusOK, gin.H{
		"types": types,
	})
}

// CreateProvider creates a new provider
// @Summary Create provider
// @Description Create a new provider (all logged-in users can create)
// @Tags Providers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param provider body CreateProviderRequest true "Provider information"
// @Success 201 {object} object{message=string,provider=object} "Provider created successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Failure 500 {object} object{error=string} "Failed to create provider"
// @Router /providers [post]
func (h *ProviderHandlers) CreateProvider(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	admin := middleware.GetAdminFromContext(c)

	var req CreateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format_with_details", err.Error()),
		})
		return
	}

	provider, err := h.providerService.CreateProvider(req.Name, req.Description, req.Type, req.Config, admin)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.MapErrorToI18nKey(err, lang),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  i18n.T(lang, "provider.create_success"),
		"provider": provider,
	})
}

// GetProvider gets a single provider
// @Summary Get provider details
// @Description Get detailed information of a specified provider by ID
// @Tags Providers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Provider ID"
// @Success 200 {object} object{provider=object} "Provider details"
// @Failure 400 {object} object{error=string} "Invalid provider ID"
// @Failure 404 {object} object{error=string} "Provider not found"
// @Router /providers/{id} [get]
func (h *ProviderHandlers) GetProvider(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	admin := middleware.GetAdminFromContext(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format"),
		})
		return
	}

	provider, err := h.providerService.GetProvider(uint(id), admin)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": i18n.MapErrorToI18nKey(err, lang),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"provider": provider,
	})
}

// ListProviders gets the provider list
// @Summary Get provider list
// @Description Get the provider list, supporting filtering by name, type and pagination
// @Tags Providers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param name query string false "Provider name filter"
// @Param type query string false "Provider type filter"
// @Param page query int false "Page number, defaults to 1"
// @Param page_size query int false "Page size, defaults to 20, maximum 100"
// @Success 200 {object} object{message=string,providers=[]object,total=number,page=number,page_size=number,total_pages=number} "Provider list"
// @Failure 500 {object} object{error=string} "Failed to get provider list"
// @Router /providers [get]
func (h *ProviderHandlers) ListProviders(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	admin := middleware.GetAdminFromContext(c)

	page := 1
	pageSize := 20
	var name *string
	var providerType *string

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
	if n := c.Query("name"); n != "" {
		name = &n
	}
	if t := c.Query("type"); t != "" {
		providerType = &t
	}

	providers, total, err := h.providerService.ListProviders(admin, name, providerType, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.MapErrorToI18nKey(err, lang),
		})
		return
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	c.JSON(http.StatusOK, gin.H{
		"message":     i18n.T(lang, "common.success"),
		"providers":   providers,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// UpdateProvider updates a provider
// @Summary Update provider
// @Description Update information of a specified provider
// @Tags Providers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Provider ID"
// @Param provider body UpdateProviderRequest true "Provider update information"
// @Success 200 {object} object{message=string} "Provider updated successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Failure 404 {object} object{error=string} "Provider not found"
// @Router /providers/{id} [put]
func (h *ProviderHandlers) UpdateProvider(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	admin := middleware.GetAdminFromContext(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format"),
		})
		return
	}

	var req UpdateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format_with_details", err.Error()),
		})
		return
	}

	// Build update data
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

	err = h.providerService.UpdateProvider(uint(id), updates, admin)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.MapErrorToI18nKey(err, lang),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "provider.update_success"),
	})
}

// DeleteProvider deletes a provider
// @Summary Delete provider
// @Description Delete a specified provider
// @Tags Providers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Provider ID"
// @Success 200 {object} object{message=string} "Provider deleted successfully"
// @Failure 400 {object} object{error=string} "Invalid provider ID or provider in use"
// @Failure 404 {object} object{error=string} "Provider not found"
// @Router /providers/{id} [delete]
func (h *ProviderHandlers) DeleteProvider(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	admin := middleware.GetAdminFromContext(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format"),
		})
		return
	}

	err = h.providerService.DeleteProvider(uint(id), admin)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.MapErrorToI18nKey(err, lang),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "provider.delete_success"),
	})
}
