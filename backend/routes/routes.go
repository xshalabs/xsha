package routes

import (
	"sleep0-backend/handlers"
	"sleep0-backend/middleware"
	"sleep0-backend/services"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes sets up routes
func SetupRoutes(r *gin.Engine, authService services.AuthService, authHandlers *handlers.AuthHandlers, gitCredHandlers *handlers.GitCredentialHandlers, projectHandlers *handlers.ProjectHandlers, operationLogHandlers *handlers.AdminOperationLogHandlers) {
	// Apply global middleware
	r.Use(middleware.I18nMiddleware())
	r.Use(middleware.ErrorHandlerMiddleware())

	// Set 404 and 405 error handlers
	r.NoRoute(middleware.NotFoundHandler())
	r.NoMethod(middleware.MethodNotAllowedHandler())

	// Swagger documentation route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check route
	r.GET("/health", handlers.HealthHandler)

	// Internationalization related routes (no authentication required)
	r.GET("/api/v1/languages", handlers.GetLanguagesHandler)
	r.POST("/api/v1/language", handlers.SetLanguageHandler)

	// Authentication related routes (no authentication required)
	auth := r.Group("/api/v1/auth")
	{
		// Login route applies rate limiting middleware
		auth.POST("/login", middleware.LoginRateLimitMiddleware(), authHandlers.LoginHandler)
	}

	// API route group (authentication required)
	api := r.Group("/api/v1")
	api.Use(middleware.AuthMiddlewareWithService(authService))

	// 添加操作日志记录中间件（在认证中间件之后）
	api.Use(middleware.OperationLogMiddleware(operationLogHandlers.OperationLogService))

	{
		// User information
		api.GET("/user/current", authHandlers.CurrentUserHandler)

		// Logout (requires token)
		api.POST("/auth/logout", authHandlers.LogoutHandler)

		// 管理员功能
		admin := api.Group("/admin")
		{
			admin.GET("/login-logs", authHandlers.GetLoginLogsHandler)

			// 新增：操作日志相关路由
			admin.GET("/operation-logs", operationLogHandlers.GetOperationLogs)    // 获取操作日志列表
			admin.GET("/operation-logs/:id", operationLogHandlers.GetOperationLog) // 获取单个操作日志
			admin.GET("/operation-stats", operationLogHandlers.GetOperationStats)  // 获取操作统计
		}

		// Git凭据管理
		gitCreds := api.Group("/git-credentials")
		{
			gitCreds.POST("", gitCredHandlers.CreateCredential)            // 创建凭据
			gitCreds.GET("", gitCredHandlers.ListCredentials)              // 获取凭据列表
			gitCreds.GET("/:id", gitCredHandlers.GetCredential)            // 获取单个凭据
			gitCreds.PUT("/:id", gitCredHandlers.UpdateCredential)         // 更新凭据
			gitCreds.DELETE("/:id", gitCredHandlers.DeleteCredential)      // 删除凭据
			gitCreds.POST("/:id/toggle", gitCredHandlers.ToggleCredential) // 切换激活状态
			gitCreds.POST("/:id/use", gitCredHandlers.UseCredential)       // 使用凭据
		}

		// 项目管理
		projects := api.Group("/projects")
		{
			projects.POST("", projectHandlers.CreateProject)                            // 创建项目
			projects.GET("", projectHandlers.ListProjects)                              // 获取项目列表
			projects.POST("/parse-url", projectHandlers.ParseRepositoryURL)             // 解析仓库URL
			projects.POST("/branches", projectHandlers.FetchRepositoryBranches)         // 获取仓库分支列表
			projects.POST("/validate-access", projectHandlers.ValidateRepositoryAccess) // 验证仓库访问权限
			projects.GET("/credentials", projectHandlers.GetCompatibleCredentials)      // 获取兼容的凭据列表
			projects.GET("/:id", projectHandlers.GetProject)                            // 获取单个项目
			projects.PUT("/:id", projectHandlers.UpdateProject)                         // 更新项目
			projects.DELETE("/:id", projectHandlers.DeleteProject)                      // 删除项目
			projects.POST("/:id/toggle", projectHandlers.ToggleProject)                 // 切换激活状态
			projects.POST("/:id/use", projectHandlers.UseProject)                       // 使用项目
		}
	}
}
