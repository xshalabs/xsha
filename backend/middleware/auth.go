package middleware

import (
	"net/http"
	"xsha-backend/config"
	"xsha-backend/i18n"
	"xsha-backend/services"
	"xsha-backend/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddlewareWithService(authService services.AuthService, cfg *config.Config) gin.HandlerFunc {
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

		c.Set("username", claims.Username)
		c.Next()
	}
}
