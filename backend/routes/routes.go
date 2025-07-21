package routes

import (
	"sleep0-backend/config"
	"sleep0-backend/handlers"
	"sleep0-backend/middleware"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由
func SetupRoutes(r *gin.Engine) {
	cfg := config.Load()

	// 设置 session 存储
	store := cookie.NewStore([]byte(cfg.SessionSecret))
	r.Use(sessions.Sessions("session", store))

	// 健康检查路由
	r.GET("/health", handlers.HealthHandler)

	// 认证相关路由（无需认证）
	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/login", handlers.LoginHandler)
		auth.POST("/logout", handlers.LogoutHandler)
	}

	// API 路由组（需要认证）
	api := r.Group("/api/v1")
	api.Use(middleware.AuthMiddleware())
	{
		// 用户信息
		api.GET("/user/current", handlers.CurrentUserHandler)

		// 这里可以添加其他需要认证的路由
	}
}
