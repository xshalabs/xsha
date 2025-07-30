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
func SetupRoutes(r *gin.Engine, cfg *config.Config, authService services.AuthService, authHandlers *handlers.AuthHandlers, gitCredHandlers *handlers.GitCredentialHandlers, projectHandlers *handlers.ProjectHandlers, operationLogHandlers *handlers.AdminOperationLogHandlers, devEnvHandlers *handlers.DevEnvironmentHandlers, taskHandlers *handlers.TaskHandlers, taskConvHandlers *handlers.TaskConversationHandlers, taskConvResultHandlers *handlers.TaskConversationResultHandlers, taskExecLogHandlers *handlers.TaskExecutionLogHandlers, sseLogHandlers *handlers.SSELogHandlers, systemConfigHandlers *handlers.SystemConfigHandlers) {
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

	// Add operation log middleware (after authentication middleware)
	api.Use(middleware.OperationLogMiddleware(operationLogHandlers.OperationLogService))

	{
		// User information
		api.GET("/user/current", authHandlers.CurrentUserHandler)

		// Logout (requires token)
		api.POST("/auth/logout", authHandlers.LogoutHandler)

		// Admin functions
		admin := api.Group("/admin")
		{
			admin.GET("/login-logs", authHandlers.GetLoginLogsHandler)

			// Operation log related routes
			admin.GET("/operation-logs", operationLogHandlers.GetOperationLogs)    // Get operation log list
			admin.GET("/operation-logs/:id", operationLogHandlers.GetOperationLog) // Get single operation log
			admin.GET("/operation-stats", operationLogHandlers.GetOperationStats)  // Get operation statistics
		}

		// Git credential management
		gitCreds := api.Group("/git-credentials")
		{
			gitCreds.POST("", gitCredHandlers.CreateCredential)       // Create credential
			gitCreds.GET("", gitCredHandlers.ListCredentials)         // Get credential list
			gitCreds.GET("/:id", gitCredHandlers.GetCredential)       // Get single credential
			gitCreds.PUT("/:id", gitCredHandlers.UpdateCredential)    // Update credential
			gitCreds.DELETE("/:id", gitCredHandlers.DeleteCredential) // Delete credential
		}

		// Project management
		projects := api.Group("/projects")
		{
			projects.POST("", projectHandlers.CreateProject)                            // Create project
			projects.GET("", projectHandlers.ListProjects)                              // Get project list
			projects.POST("/parse-url", projectHandlers.ParseRepositoryURL)             // Parse repository URL
			projects.POST("/branches", projectHandlers.FetchRepositoryBranches)         // Get repository branch list
			projects.POST("/validate-access", projectHandlers.ValidateRepositoryAccess) // Validate repository access
			projects.GET("/credentials", projectHandlers.GetCompatibleCredentials)      // Get compatible credential list
			projects.GET("/:id", projectHandlers.GetProject)                            // Get single project
			projects.PUT("/:id", projectHandlers.UpdateProject)                         // Update project
			projects.DELETE("/:id", projectHandlers.DeleteProject)                      // Delete project
		}

		// Task management
		tasks := api.Group("/tasks")
		{
			tasks.POST("", taskHandlers.CreateTask)                          // Create task
			tasks.GET("", taskHandlers.ListTasks)                            // Get task list
			tasks.GET("/:id", taskHandlers.GetTask)                          // Get single task
			tasks.PUT("/:id", taskHandlers.UpdateTask)                       // Update task
			tasks.PUT("/:id/status", taskHandlers.UpdateTaskStatus)          // Update task status
			tasks.PUT("/batch/status", taskHandlers.BatchUpdateTaskStatus)   // Batch update task status
			tasks.DELETE("/:id", taskHandlers.DeleteTask)                    // Delete task
			tasks.GET("/:id/git-diff", taskHandlers.GetTaskGitDiff)          // Get task Git changes
			tasks.GET("/:id/git-diff/file", taskHandlers.GetTaskGitDiffFile) // Get task specific file Git changes
			tasks.POST("/:id/push", taskHandlers.PushTaskBranch)             // Push task branch to remote repository
		}

		// Task conversation management
		conversations := api.Group("/conversations")
		{
			conversations.POST("", taskConvHandlers.CreateConversation)                          // Create conversation
			conversations.GET("", taskConvHandlers.ListConversations)                            // Get conversation list
			conversations.GET("/latest", taskConvHandlers.GetLatestConversation)                 // Get latest conversation
			conversations.GET("/:id", taskConvHandlers.GetConversation)                          // Get single conversation
			conversations.PUT("/:id", taskConvHandlers.UpdateConversation)                       // Update conversation
			conversations.DELETE("/:id", taskConvHandlers.DeleteConversation)                    // Delete conversation
			conversations.GET("/:id/git-diff", taskConvHandlers.GetConversationGitDiff)          // Get conversation Git changes
			conversations.GET("/:id/git-diff/file", taskConvHandlers.GetConversationGitDiffFile) // Get conversation specific file Git changes
		}

		// Task conversation result management
		results := api.Group("/conversation-results")
		{
			results.GET("", taskConvResultHandlers.ListResultsByTaskID)                                        // Get result list by task ID
			results.GET("/by-project", taskConvResultHandlers.ListResultsByProjectID)                          // Get result list by project ID
			results.GET("/:id", taskConvResultHandlers.GetResult)                                              // Get single result
			results.GET("/by-conversation/:conversation_id", taskConvResultHandlers.GetResultByConversationID) // Get result by conversation ID
			results.PUT("/:id", taskConvResultHandlers.UpdateResult)                                           // Update result
			results.DELETE("/:id", taskConvResultHandlers.DeleteResult)                                        // Delete result
		}

		// Statistics management
		stats := api.Group("/stats")
		{
			stats.GET("/tasks/:task_id", taskConvResultHandlers.GetTaskStats)          // Get task statistics
			stats.GET("/projects/:project_id", taskConvResultHandlers.GetProjectStats) // Get project statistics
		}

		// Task execution log management
		api.GET("/task-conversations/:conversationId/execution-log", taskExecLogHandlers.GetExecutionLog)
		api.POST("/task-conversations/:conversationId/execution/cancel", taskExecLogHandlers.CancelExecution)
		api.POST("/task-conversations/:conversationId/execution/retry", taskExecLogHandlers.RetryExecution)

		// SSE real-time log management
		logs := api.Group("/logs")
		{
			logs.GET("/stream", sseLogHandlers.StreamLogs)                     // SSE real-time log stream
			logs.GET("/stats", sseLogHandlers.GetLogStats)                     // Get connection statistics
			logs.POST("/test/:conversationId", sseLogHandlers.SendTestMessage) // Send test message
		}

		// Development environment management
		devEnvs := api.Group("/dev-environments")
		{
			devEnvs.POST("", devEnvHandlers.CreateEnvironment)                 // Create environment
			devEnvs.GET("", devEnvHandlers.ListEnvironments)                   // Get environment list
			devEnvs.GET("/available-types", devEnvHandlers.GetAvailableTypes)  // Get available environment types
			devEnvs.GET("/:id", devEnvHandlers.GetEnvironment)                 // Get single environment
			devEnvs.PUT("/:id", devEnvHandlers.UpdateEnvironment)              // Update environment
			devEnvs.DELETE("/:id", devEnvHandlers.DeleteEnvironment)           // Delete environment
			devEnvs.GET("/:id/env-vars", devEnvHandlers.GetEnvironmentVars)    // Get environment variables
			devEnvs.PUT("/:id/env-vars", devEnvHandlers.UpdateEnvironmentVars) // Update environment variables
		}

		// System configuration management
		systemConfigs := api.Group("/system-configs")
		{
			systemConfigs.GET("", systemConfigHandlers.ListAllConfigs)     // Get all configurations
			systemConfigs.PUT("", systemConfigHandlers.BatchUpdateConfigs) // Batch update configurations
		}
	}
}
