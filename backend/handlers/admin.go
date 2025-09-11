package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"xsha-backend/database"
	"xsha-backend/i18n"
	"xsha-backend/middleware"
	"xsha-backend/services"

	"github.com/gin-gonic/gin"
)

type AdminHandlers struct {
	adminService services.AdminService
}

func NewAdminHandlers(adminService services.AdminService) *AdminHandlers {
	return &AdminHandlers{
		adminService: adminService,
	}
}

// CreateAdminHandler creates a new admin user
// @Summary Create admin user
// @Description Create a new administrator user with optional role assignment
// @Tags Admin Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param adminData body object{username=string,password=string,name=string,email=string,role=string} true "Admin user information (role is optional, defaults to 'admin')"
// @Success 200 {object} object{message=string,admin=object} "Admin created successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Failure 500 {object} object{error=string} "Internal server error"
// @Router /admin/users [post]
func (h *AdminHandlers) CreateAdminHandler(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	var adminData struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email"`
		Role     string `json:"role" binding:"omitempty,oneof=super_admin admin developer"`
	}

	if err := c.ShouldBindJSON(&adminData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format_with_details", err.Error()),
		})
		return
	}

	// Get current user as createdBy
	createdBy, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.T(lang, "common.internal_error"),
		})
		return
	}

	// Determine role, default to 'admin' if not specified
	role := database.AdminRoleAdmin
	if adminData.Role != "" {
		role = database.AdminRole(adminData.Role)
	}

	admin, err := h.adminService.CreateAdminWithRole(adminData.Username, adminData.Password, adminData.Name, adminData.Email, role, createdBy.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.MapErrorToI18nKey(err, lang),
		})
		return
	}

	// Hide password hash
	admin.PasswordHash = ""

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "admin.create_success"),
		"admin":   admin,
	})
}

// GetAdminHandler gets admin user by ID
// @Summary Get admin user
// @Description Get administrator user by ID
// @Tags Admin Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Admin ID"
// @Success 200 {object} object{admin=object} "Admin information"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Failure 404 {object} object{error=string} "Admin not found"
// @Router /admin/users/{id} [get]
func (h *AdminHandlers) GetAdminHandler(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_id"),
		})
		return
	}

	admin, err := h.adminService.GetAdmin(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": i18n.MapErrorToI18nKey(err, lang),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"admin": admin,
	})
}

// ListAdminsHandler lists admin users with pagination and filtering
// @Summary List admin users
// @Description List all administrator users with pagination and filtering
// @Tags Admin Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param username query string false "Username filter"
// @Param is_active query boolean false "Active status filter"
// @Param role query string false "Role filter (comma-separated, e.g., 'admin,super_admin')"
// @Param page query int false "Page number, defaults to 1"
// @Param page_size query int false "Page size, defaults to 20, maximum 100"
// @Success 200 {object} object{message=string,admins=[]object,total=number,page=number,page_size=number,total_pages=number} "Admin list"
// @Failure 500 {object} object{error=string} "Failed to get admin list"
// @Router /admin/users [get]
func (h *AdminHandlers) ListAdminsHandler(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	// Parse query parameters
	page := 1
	pageSize := 20
	var search *string
	var isActive *bool
	var roles []string

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

	// Support both 'search' parameter and legacy 'username' parameter for backward compatibility
	if s := c.Query("search"); s != "" {
		search = &s
	} else if u := c.Query("username"); u != "" {
		search = &u
	}

	if ia := c.Query("is_active"); ia != "" {
		if parsed, err := strconv.ParseBool(ia); err == nil {
			isActive = &parsed
		}
	}

	// Parse role parameter - support comma-separated roles
	if r := c.Query("role"); r != "" {
		roles = strings.Split(r, ",")
		// Trim whitespace from each role
		for i, role := range roles {
			roles[i] = strings.TrimSpace(role)
		}
	}

	admins, total, err := h.adminService.ListAdmins(search, isActive, roles, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.T(lang, "common.internal_error"),
		})
		return
	}

	// Transform to list response with minimal avatar data
	adminResponses := database.ToAdminListResponses(admins)

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	c.JSON(http.StatusOK, gin.H{
		"message":     i18n.T(lang, "admin.list_success"),
		"admins":      adminResponses,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// UpdateAdminHandler updates admin user information
// @Summary Update admin user
// @Description Update administrator user information including role
// @Tags Admin Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Admin ID"
// @Param adminData body object{username=string,name=string,email=string,is_active=boolean,role=string} true "Admin update information (role: super_admin|admin|developer)"
// @Success 200 {object} object{message=string} "Admin updated successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Failure 404 {object} object{error=string} "Admin not found"
// @Router /admin/users/{id} [put]
func (h *AdminHandlers) UpdateAdminHandler(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_id"),
		})
		return
	}

	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format_with_details", err.Error()),
		})
		return
	}

	// Handle role update separately if provided
	if role, exists := updateData["role"]; exists {
		if roleStr, ok := role.(string); ok {
			// Validate role value
			validRoles := []string{"super_admin", "admin", "developer"}
			isValid := false
			for _, validRole := range validRoles {
				if roleStr == validRole {
					isValid = true
					break
				}
			}
			if !isValid {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": i18n.T(lang, "validation.invalid_role"),
				})
				return
			}

			// Update role using dedicated method
			if err := h.adminService.UpdateAdminRole(uint(id), database.AdminRole(roleStr)); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": i18n.MapErrorToI18nKey(err, lang),
				})
				return
			}

			// Remove role from updateData to avoid double processing
			delete(updateData, "role")
		}
	}

	// Update other fields if any remain
	if len(updateData) > 0 {
		if err := h.adminService.UpdateAdmin(uint(id), updateData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": i18n.MapErrorToI18nKey(err, lang),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "admin.update_success"),
	})
}

// DeleteAdminHandler deletes admin user
// @Summary Delete admin user
// @Description Delete administrator user (cannot delete the last admin)
// @Tags Admin Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Admin ID"
// @Success 200 {object} object{message=string} "Admin deleted successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Failure 404 {object} object{error=string} "Admin not found"
// @Router /admin/users/{id} [delete]
func (h *AdminHandlers) DeleteAdminHandler(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_id"),
		})
		return
	}

	if err := h.adminService.DeleteAdmin(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.MapErrorToI18nKey(err, lang),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "admin.delete_success"),
	})
}

// ChangePasswordHandler changes admin user password
// @Summary Change admin password
// @Description Change administrator user password
// @Tags Admin Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Admin ID"
// @Param passwordData body object{new_password=string} true "New password"
// @Success 200 {object} object{message=string} "Password changed successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Failure 404 {object} object{error=string} "Admin not found"
// @Router /admin/users/{id}/password [put]
func (h *AdminHandlers) ChangePasswordHandler(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_id"),
		})
		return
	}

	var passwordData struct {
		NewPassword string `json:"new_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&passwordData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(lang, "validation.invalid_format_with_details", err.Error()),
		})
		return
	}

	if err := h.adminService.ChangePassword(uint(id), passwordData.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.MapErrorToI18nKey(err, lang),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(lang, "admin.password_change_success"),
	})
}


// PublicListAdminsHandler lists admin users with authentication
// @Summary List admin users (authenticated)
// @Description Get list of all administrator users (requires authentication)
// @Tags Admin Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object{admins=[]object} "Admin list"
// @Failure 500 {object} object{error=string} "Failed to get admin list"
// @Router /api/v1/admins [get]
func (h *AdminHandlers) PublicListAdminsHandler(c *gin.Context) {
	// Get all active admins
	admins, _, err := h.adminService.ListAdmins(nil, nil, nil, 1, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get admin list",
		})
		return
	}

	// Transform to minimal response format
	adminResponses := database.ToMinimalAdminResponses(admins)

	c.JSON(http.StatusOK, gin.H{
		"admins": adminResponses,
	})
}
