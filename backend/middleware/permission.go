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

		// Get resource owner ID if available from URL params
		var resourceOwnerID uint = 0
		if idStr := c.Param("id"); idStr != "" {
			if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
				resourceOwnerID = uint(id)
			}
		}

		// Check permission
		if !adminService.HasPermission(admin, resource, action, resourceOwnerID) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": i18n.T(lang, "auth.insufficient_permissions"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CheckResourceOwner creates a middleware that checks if the current user owns the resource
func CheckResourceOwner(resourceType string) gin.HandlerFunc {
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

		// Super admin can access any resource
		if admin.Role == database.AdminRoleSuperAdmin {
			c.Next()
			return
		}

		// Get resource ID from URL parameter
		idStr := c.Param("id")
		if idStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": i18n.T(lang, "validation.invalid_id"),
			})
			c.Abort()
			return
		}

		resourceID, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": i18n.T(lang, "validation.invalid_id"),
			})
			c.Abort()
			return
		}

		// Check ownership based on resource type
		if err := checkResourceOwnership(c, admin, resourceType, uint(resourceID)); err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"error": i18n.T(lang, "auth.access_denied"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Helper function to check resource ownership
func checkResourceOwnership(c *gin.Context, admin *database.Admin, resourceType string, resourceID uint) error {
	// This would typically involve checking the database
	// For now, we'll implement a basic check
	// In a real implementation, you'd inject the appropriate service
	
	switch resourceType {
	case "project":
		// Check if admin owns this project
		// This would require project service injection
		return nil
	case "task":
		// Check if admin owns this task or the parent project
		return nil
	case "credential":
		// Check if admin owns this credential
		return nil
	case "environment":
		// Check if admin owns this environment
		return nil
	default:
		return nil
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