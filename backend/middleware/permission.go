package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"xsha-backend/database"
	"xsha-backend/i18n"
	"xsha-backend/services"
	"xsha-backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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

		// Get the actual resource owner ID based on resource type and resource ID
		resourceOwnerID, err := getResourceOwnerID(c, resource, resourceID)
		if err != nil {
			utils.Info("error", "err", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": i18n.T(lang, "auth.permission_check_failed"),
			})
			c.Abort()
			return
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

// getResourceOwnerID gets the actual owner ID for a given resource using services
func getResourceOwnerID(c *gin.Context, resourceType string, resourceID uint) (uint, error) {
	// If no resource ID provided, return 0 (no specific owner)
	if resourceID == 0 {
		return 0, nil
	}

	switch resourceType {
	case "project":
		projectService, exists := c.Get("projectService")
		if !exists {
			return 0, fmt.Errorf("projectService not found in context")
		}
		project, err := projectService.(services.ProjectService).GetProject(resourceID)
		if err != nil {
			// Check if it's a not found error (most services return gorm.ErrRecordNotFound)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return 0, nil // Resource not found, no owner
			}
			return 0, fmt.Errorf("failed to get project: %v", err)
		}
		if project.AdminID != nil {
			return *project.AdminID, nil
		}
		return 0, nil

	case "task":
		taskService, exists := c.Get("taskService")
		if !exists {
			return 0, fmt.Errorf("taskService not found in context")
		}
		task, err := taskService.(services.TaskService).GetTask(resourceID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return 0, nil // Resource not found, no owner
			}
			return 0, fmt.Errorf("failed to get task: %v", err)
		}
		if task.AdminID != nil {
			return *task.AdminID, nil
		}
		return 0, nil

	case "conversation":
		taskConvService, exists := c.Get("taskConvService")
		if !exists {
			return 0, fmt.Errorf("taskConvService not found in context")
		}
		conversationData, err := taskConvService.(services.TaskConversationService).GetConversationWithResult(resourceID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return 0, nil // Resource not found, no owner
			}
			return 0, fmt.Errorf("failed to get conversation: %v", err)
		}
		if conversationObj, ok := conversationData["conversation"]; ok {
			if conversation, ok := conversationObj.(*database.TaskConversation); ok {
				if conversation.AdminID != nil {
					return *conversation.AdminID, nil
				}
			}
		}
		return 0, nil

	case "credential":
		gitCredService, exists := c.Get("gitCredService")
		if !exists {
			return 0, fmt.Errorf("gitCredService not found in context")
		}
		credential, err := gitCredService.(services.GitCredentialService).GetCredential(resourceID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return 0, nil // Resource not found, no owner
			}
			return 0, fmt.Errorf("failed to get credential: %v", err)
		}
		if credential.AdminID != nil {
			return *credential.AdminID, nil
		}
		return 0, nil

	case "environment":
		devEnvService, exists := c.Get("devEnvService")
		if !exists {
			return 0, fmt.Errorf("devEnvService not found in context")
		}
		environment, err := devEnvService.(services.DevEnvironmentService).GetEnvironment(resourceID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return 0, nil // Resource not found, no owner
			}
			return 0, fmt.Errorf("failed to get environment: %v", err)
		}
		if environment.AdminID != nil {
			return *environment.AdminID, nil
		}
		return 0, nil

	default:
		// Unknown resource type, return 0
		return 0, nil
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
