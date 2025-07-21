package middleware

import (
	"net/http"
	"sleep0-backend/config"
	"sleep0-backend/database"
	"sleep0-backend/i18n"
	"sleep0-backend/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.Load()
		lang := GetLangFromContext(c)

		// 从Authorization header中获取token
		authHeader := c.GetHeader("Authorization")
		token, err := utils.ExtractTokenFromAuthHeader(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": i18n.T(lang, "auth.unauthorized") + "：" + err.Error(),
			})
			c.Abort()
			return
		}

		// 检查token是否在黑名单中
		if database.IsTokenBlacklisted(token) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": i18n.T(lang, "auth.token_blacklisted"),
			})
			c.Abort()
			return
		}

		// 验证JWT token
		claims, err := utils.ValidateJWT(token, cfg.JWTSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": i18n.T(lang, "auth.invalid_token") + "：" + err.Error(),
			})
			c.Abort()
			return
		}

		// 将用户信息存储到context中供后续处理器使用
		c.Set("username", claims.Username)
		c.Next()
	}
}
