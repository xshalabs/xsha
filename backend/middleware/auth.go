package middleware

import (
	"net/http"
	"sleep0-backend/config"
	"sleep0-backend/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.Load()

		// 从Authorization header中获取token
		authHeader := c.GetHeader("Authorization")
		token, err := utils.ExtractTokenFromAuthHeader(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "未授权访问：" + err.Error(),
			})
			c.Abort()
			return
		}

		// 验证JWT token
		claims, err := utils.ValidateJWT(token, cfg.JWTSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "无效的token：" + err.Error(),
			})
			c.Abort()
			return
		}

		// 将用户信息存储到context中供后续处理器使用
		c.Set("username", claims.Username)
		c.Next()
	}
}
