package handlers

import (
	"net/http"
	"strconv"
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
// @Description Create a new administrator user
// @Tags Admin Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param adminData body object{username=string,password=string,name=string,email=string} true "Admin user information"
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

	admin, err := h.adminService.CreateAdmin(adminData.Username, adminData.Password, adminData.Name, adminData.Email, createdBy.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.MapErrorToI18nKey(err, lang),
		})
		return
	}

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

	admins, total, err := h.adminService.ListAdmins(search, isActive, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.T(lang, "common.internal_error"),
		})
		return
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	c.JSON(http.StatusOK, gin.H{
		"message":     i18n.T(lang, "admin.list_success"),
		"admins":      admins,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// UpdateAdminHandler updates admin user information
// @Summary Update admin user
// @Description Update administrator user information
// @Tags Admin Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Admin ID"
// @Param adminData body object{username=string,name=string,email=string,is_active=boolean} true "Admin update information"
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

	if err := h.adminService.UpdateAdmin(uint(id), updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.MapErrorToI18nKey(err, lang),
		})
		return
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
