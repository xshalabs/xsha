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

		authHeader := c.GetHeader("Authorization")
		token, err := utils.ExtractTokenFromAuthHeader(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": i18n.T(lang, "auth.unauthorized") + ": " + err.Error(),
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
				"error": i18n.T(lang, "auth.invalid_token") + ": " + err.Error(),
			})
			c.Abort()
			return
		}

		c.Set("username", claims.Username)
		c.Next()
	}
}
