package middleware

import (
	"net/http"
	"strconv"
	"xsha-backend/database"
	"xsha-backend/i18n"
	"xsha-backend/services"

	"github.com/gin-gonic/gin"
)

// RequireRole creates a middleware that requires specific roles
func RequireRole(roles ...database.AdminRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := GetLangFromContext(c)
		admin := GetAdminFromContext(c)

		if admin == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": i18n.T(lang, "auth.unauthorized"),
			})
			c.Abort()
			return
		}

		// Check if admin has any of the required roles
		hasRole := false
		for _, role := range roles {
			if admin.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error": i18n.T(lang, "auth.insufficient_permissions"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermission creates a middleware that checks specific permissions
func RequirePermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := GetLangFromContext(c)
		admin := GetAdminFromContext(c)
		adminService := c.MustGet("adminService").(services.AdminService)

		if admin == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": i18n.T(lang, "auth.unauthorized"),
			})
			c.Abort()
			return
		}

		// Get resource ID from URL params
		var resourceID uint = 0
		if idStr := c.Param("id"); idStr != "" {
			if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
				resourceID = uint(id)
			}
		}

		// Check permission (AdminService will handle resource owner lookup internally)
		if !adminService.HasPermission(admin, resource, action, resourceID) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": i18n.T(lang, "auth.insufficient_permissions"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetAdminFromContext retrieves admin from gin context
func GetAdminFromContext(c *gin.Context) *database.Admin {
	if admin, exists := c.Get("admin"); exists {
		return admin.(*database.Admin)
	}
	return nil
}

// RequireSuperAdmin is a convenience middleware for super admin only access
func RequireSuperAdmin() gin.HandlerFunc {
	return RequireRole(database.AdminRoleSuperAdmin)
}

// RequireAdminOrSuperAdmin is a convenience middleware for admin or super admin access
func RequireAdminOrSuperAdmin() gin.HandlerFunc {
	return RequireRole(database.AdminRoleAdmin, database.AdminRoleSuperAdmin)
}
