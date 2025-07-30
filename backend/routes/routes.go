package routes

import (
	"xsha-backend/config"
	"xsha-backend/handlers"
	"xsha-backend/middleware"
	"xsha-backend/services"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes sets up routes
func SetupRoutes(r *gin.Engine, cfg *config.Config, authService services.AuthService, authHandlers *handlers.AuthHandlers, gitCredHandlers *handlers.GitCredentialHandlers, projectHandlers *handlers.ProjectHandlers, operationLogHandlers *handlers.AdminOperationLogHandlers, devEnvHandlers *handlers.DevEnvironmentHandlers, taskHandlers *handlers.TaskHandlers, taskConvHandlers *handlers.TaskConversationHandlers, taskConvResultHandlers *handlers.TaskConversationResultHandlers, taskExecLogHandlers *handlers.TaskExecutionLogHandlers, sseLogHandlers *handlers.SSELogHandlers) {
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
	api.Use(middleware.AuthMiddlewareWithService(authService, cfg))

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
			gitCreds.POST("", gitCredHandlers.CreateCredential)       // 创建凭据
			gitCreds.GET("", gitCredHandlers.ListCredentials)         // 获取凭据列表
			gitCreds.GET("/:id", gitCredHandlers.GetCredential)       // 获取单个凭据
			gitCreds.PUT("/:id", gitCredHandlers.UpdateCredential)    // 更新凭据
			gitCreds.DELETE("/:id", gitCredHandlers.DeleteCredential) // 删除凭据
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
		}

		// 任务管理
		tasks := api.Group("/tasks")
		{
			tasks.POST("", taskHandlers.CreateTask)                        // 创建任务
			tasks.GET("", taskHandlers.ListTasks)                          // 获取任务列表
			tasks.GET("/:id", taskHandlers.GetTask)                        // 获取单个任务
			tasks.PUT("/:id", taskHandlers.UpdateTask)                     // 更新任务
			tasks.PUT("/:id/status", taskHandlers.UpdateTaskStatus)        // 更新任务状态
			tasks.PUT("/batch/status", taskHandlers.BatchUpdateTaskStatus) // 批量更新任务状态
			tasks.DELETE("/:id", taskHandlers.DeleteTask)                  // 删除任务
		}

		// 任务对话管理
		conversations := api.Group("/conversations")
		{
			conversations.POST("", taskConvHandlers.CreateConversation)          // 创建对话
			conversations.GET("", taskConvHandlers.ListConversations)            // 获取对话列表
			conversations.GET("/latest", taskConvHandlers.GetLatestConversation) // 获取最新对话
			conversations.GET("/:id", taskConvHandlers.GetConversation)          // 获取单个对话
			conversations.PUT("/:id", taskConvHandlers.UpdateConversation)       // 更新对话
			conversations.DELETE("/:id", taskConvHandlers.DeleteConversation)    // 删除对话
		}

		// 任务对话结果管理
		results := api.Group("/conversation-results")
		{
			results.GET("", taskConvResultHandlers.ListResultsByTaskID)                                        // 根据任务ID获取结果列表
			results.GET("/by-project", taskConvResultHandlers.ListResultsByProjectID)                          // 根据项目ID获取结果列表
			results.GET("/:id", taskConvResultHandlers.GetResult)                                              // 获取单个结果
			results.GET("/by-conversation/:conversation_id", taskConvResultHandlers.GetResultByConversationID) // 根据对话ID获取结果
			results.PUT("/:id", taskConvResultHandlers.UpdateResult)                                           // 更新结果
			results.DELETE("/:id", taskConvResultHandlers.DeleteResult)                                        // 删除结果
		}

		// 统计信息管理
		stats := api.Group("/stats")
		{
			stats.GET("/tasks/:task_id", taskConvResultHandlers.GetTaskStats)          // 获取任务统计
			stats.GET("/projects/:project_id", taskConvResultHandlers.GetProjectStats) // 获取项目统计
		}

		// 任务执行日志管理
		api.GET("/task-conversations/:conversationId/execution-log", taskExecLogHandlers.GetExecutionLog)
		api.POST("/task-conversations/:conversationId/execution/cancel", taskExecLogHandlers.CancelExecution)
		api.POST("/task-conversations/:conversationId/execution/retry", taskExecLogHandlers.RetryExecution)

		// SSE实时日志管理
		logs := api.Group("/logs")
		{
			logs.GET("/stream", sseLogHandlers.StreamLogs)                     // SSE实时日志流
			logs.GET("/stats", sseLogHandlers.GetLogStats)                     // 获取连接统计
			logs.POST("/test/:conversationId", sseLogHandlers.SendTestMessage) // 发送测试消息
		}

		// 开发环境管理
		devEnvs := api.Group("/dev-environments")
		{
			devEnvs.POST("", devEnvHandlers.CreateEnvironment)                 // 创建环境
			devEnvs.GET("", devEnvHandlers.ListEnvironments)                   // 获取环境列表
			devEnvs.GET("/:id", devEnvHandlers.GetEnvironment)                 // 获取单个环境
			devEnvs.PUT("/:id", devEnvHandlers.UpdateEnvironment)              // 更新环境
			devEnvs.DELETE("/:id", devEnvHandlers.DeleteEnvironment)           // 删除环境
			devEnvs.GET("/:id/env-vars", devEnvHandlers.GetEnvironmentVars)    // 获取环境变量
			devEnvs.PUT("/:id/env-vars", devEnvHandlers.UpdateEnvironmentVars) // 更新环境变量
		}
	}
}
