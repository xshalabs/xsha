package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		isAuthenticated := session.Get("authenticated")

		if isAuthenticated != true {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "未授权访问，请先登录",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
