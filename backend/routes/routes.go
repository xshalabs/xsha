package routes

import (
	"sleep0-backend/handlers"
	"sleep0-backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up routes
func SetupRoutes(r *gin.Engine) {
	// Apply global middleware
	r.Use(middleware.I18nMiddleware())
	r.Use(middleware.ErrorHandlerMiddleware())

	// Set 404 and 405 error handlers
	r.NoRoute(middleware.NotFoundHandler())
	r.NoMethod(middleware.MethodNotAllowedHandler())

	// Health check route
	r.GET("/health", handlers.HealthHandler)

	// Internationalization related routes (no authentication required)
	r.GET("/api/v1/languages", handlers.GetLanguagesHandler)
	r.POST("/api/v1/language", handlers.SetLanguageHandler)

	// Authentication related routes (no authentication required)
	auth := r.Group("/api/v1/auth")
	{
		// Login route applies rate limiting middleware
		auth.POST("/login", middleware.LoginRateLimitMiddleware(), handlers.LoginHandler)
	}

	// API route group (authentication required)
	api := r.Group("/api/v1")
	api.Use(middleware.AuthMiddleware())
	{
		// User information
		api.GET("/user/current", handlers.CurrentUserHandler)

		// Logout (requires token)
		api.POST("/auth/logout", handlers.LogoutHandler)

		// 新增：管理员功能
		admin := api.Group("/admin")
		{
			admin.GET("/login-logs", handlers.GetLoginLogsHandler)
		}

		// Additional authenticated routes can be added here
	}
}
