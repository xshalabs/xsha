package middleware

import (
	"net/http"
	"sleep0-backend/config"
	"sleep0-backend/database"
	"sleep0-backend/i18n"
	"sleep0-backend/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware authentication middleware
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.Load()
		lang := GetLangFromContext(c)

		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		token, err := utils.ExtractTokenFromAuthHeader(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": i18n.T(lang, "auth.unauthorized") + ": " + err.Error(),
			})
			c.Abort()
			return
		}

		// Check if token is in blacklist
		if database.IsTokenBlacklisted(token) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": i18n.T(lang, "auth.token_blacklisted"),
			})
			c.Abort()
			return
		}

		// Validate JWT token
		claims, err := utils.ValidateJWT(token, cfg.JWTSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": i18n.T(lang, "auth.invalid_token") + ": " + err.Error(),
			})
			c.Abort()
			return
		}

		// Store user information in context for subsequent handlers
		c.Set("username", claims.Username)
		c.Next()
	}
}
