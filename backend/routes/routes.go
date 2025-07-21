package routes

import (
	"sleep0-backend/handlers"
	"sleep0-backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由
func SetupRoutes(r *gin.Engine) {
	// 应用全局中间件
	r.Use(middleware.I18nMiddleware())

	// 健康检查路由
	r.GET("/health", handlers.HealthHandler)

	// 认证相关路由（无需认证）
	auth := r.Group("/api/v1/auth")
	{
		// 登录路由应用限流中间件
		auth.POST("/login", middleware.LoginRateLimitMiddleware(), handlers.LoginHandler)
	}

	// API 路由组（需要认证）
	api := r.Group("/api/v1")
	api.Use(middleware.AuthMiddleware())
	{
		// 用户信息
		api.GET("/user/current", handlers.CurrentUserHandler)

		// 登出（需要token）
		api.POST("/auth/logout", handlers.LogoutHandler)

		// 这里可以添加其他需要认证的路由
	}
}
