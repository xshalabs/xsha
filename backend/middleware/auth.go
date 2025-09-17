package middleware

import (
	"net/http"
	"xsha-backend/config"
	"xsha-backend/database"
	"xsha-backend/i18n"
	"xsha-backend/services"
	"xsha-backend/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddlewareWithService(authService services.AuthService, adminService services.AdminService, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := GetLangFromContext(c)

		// First try to get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		token, err := utils.ExtractTokenFromAuthHeader(authHeader)

		// If header auth fails, try to get token from query parameter (for SSE requests)
		if err != nil {
			queryToken := c.Query("token")
			if queryToken != "" {
				token = queryToken
				err = nil
			}
		}

		if err != nil || token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": i18n.T(lang, "auth.unauthorized_with_details", "missing Authorization header or token parameter"),
			})
			c.Abort()
			return
		}

		isBlacklisted, err := authService.IsTokenBlacklisted(token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": i18n.T(lang, "auth.server_error"),
			})
			c.Abort()
			return
		}

		if isBlacklisted {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": i18n.T(lang, "auth.token_blacklisted"),
			})
			c.Abort()
			return
		}

		claims, err := utils.ValidateJWT(token, cfg.JWTSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": i18n.T(lang, "auth.invalid_token_with_details", err.Error()),
			})
			c.Abort()
			return
		}

		// Check if admin is still active
		isActive, err := authService.CheckAdminStatus(claims.AdminID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": i18n.T(lang, "auth.server_error"),
			})
			c.Abort()
			return
		}

		if !isActive {
			// Get admin for username (for logging purposes)
			admin, adminErr := adminService.GetAdmin(claims.AdminID)
			username := "unknown"
			if adminErr == nil {
				username = admin.Username
			}

			// Add token to blacklist with reason "admin_deactivated"
			go func() {
				if blacklistErr := authService.Logout(token, username, c.ClientIP(), c.GetHeader("User-Agent")); blacklistErr != nil {
					utils.Error("Failed to blacklist token for deactivated admin",
						"admin_id", claims.AdminID,
						"error", blacklistErr.Error(),
					)
				}
			}()

			c.JSON(http.StatusUnauthorized, gin.H{
				"error": i18n.T(lang, "auth.admin_deactivated"),
			})
			c.Abort()
			return
		}

		// Get admin details to set in context
		admin, err := adminService.GetAdmin(claims.AdminID)
		if err != nil {
			utils.Error("Failed to get admin details for context",
				"admin_id", claims.AdminID,
				"error", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": i18n.T(lang, "auth.server_error"),
			})
			c.Abort()
			return
		}

		c.Set("username", admin.Username)
		c.Set("admin_id", claims.AdminID)
		c.Set("admin", admin)
		c.Set("adminService", adminService)

		// Update admin language preference if current language is different from stored preference
		updateAdminLanguagePreference(c, admin, adminService)

		c.Next()
	}
}

func updateAdminLanguagePreference(c *gin.Context, admin *database.Admin, adminService services.AdminService) {
	currentLang := GetLangFromContext(c)

	// Only update if the current language is different from stored preference
	if admin.Lang != currentLang {
		go func() {
			if err := adminService.UpdateAdminLanguage(admin.ID, currentLang); err != nil {
				utils.Error("Failed to update admin language preference",
					"admin_id", admin.ID,
					"new_lang", currentLang,
					"error", err.Error(),
				)
			}
		}()
	}
}
