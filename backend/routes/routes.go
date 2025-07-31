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

func SetupRoutes(r *gin.Engine, cfg *config.Config, authService services.AuthService, authHandlers *handlers.AuthHandlers, gitCredHandlers *handlers.GitCredentialHandlers, projectHandlers *handlers.ProjectHandlers, operationLogHandlers *handlers.AdminOperationLogHandlers, devEnvHandlers *handlers.DevEnvironmentHandlers, taskHandlers *handlers.TaskHandlers, taskConvHandlers *handlers.TaskConversationHandlers, taskConvResultHandlers *handlers.TaskConversationResultHandlers, taskExecLogHandlers *handlers.TaskExecutionLogHandlers, systemConfigHandlers *handlers.SystemConfigHandlers) {
	r.Use(middleware.I18nMiddleware())
	r.Use(middleware.ErrorHandlerMiddleware())

	r.NoRoute(middleware.NotFoundHandler())
	r.NoMethod(middleware.MethodNotAllowedHandler())

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/health", handlers.HealthHandler)

	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/login", middleware.LoginRateLimitMiddleware(), authHandlers.LoginHandler)
	}

	api := r.Group("/api/v1")
	api.Use(middleware.AuthMiddlewareWithService(authService, cfg))

	api.Use(middleware.OperationLogMiddleware(operationLogHandlers.OperationLogService))

	{
		api.GET("/user/current", authHandlers.CurrentUserHandler)
		api.POST("/auth/logout", authHandlers.LogoutHandler)

		admin := api.Group("/admin")
		{
			admin.GET("/login-logs", authHandlers.GetLoginLogsHandler)

			admin.GET("/operation-logs", operationLogHandlers.GetOperationLogs)
			admin.GET("/operation-logs/:id", operationLogHandlers.GetOperationLog)
			admin.GET("/operation-stats", operationLogHandlers.GetOperationStats)
		}

		gitCreds := api.Group("/git-credentials")
		{
			gitCreds.POST("", gitCredHandlers.CreateCredential)
			gitCreds.GET("", gitCredHandlers.ListCredentials)
			gitCreds.GET("/:id", gitCredHandlers.GetCredential)
			gitCreds.PUT("/:id", gitCredHandlers.UpdateCredential)
			gitCreds.DELETE("/:id", gitCredHandlers.DeleteCredential)
		}

		projects := api.Group("/projects")
		{
			projects.POST("", projectHandlers.CreateProject)
			projects.GET("", projectHandlers.ListProjects)
			projects.POST("/parse-url", projectHandlers.ParseRepositoryURL)
			projects.POST("/branches", projectHandlers.FetchRepositoryBranches)
			projects.POST("/validate-access", projectHandlers.ValidateRepositoryAccess)
			projects.GET("/credentials", projectHandlers.GetCompatibleCredentials)
			projects.GET("/:id", projectHandlers.GetProject)
			projects.PUT("/:id", projectHandlers.UpdateProject)
			projects.DELETE("/:id", projectHandlers.DeleteProject)
		}

		tasks := api.Group("/tasks")
		{
			tasks.POST("", taskHandlers.CreateTask)
			tasks.GET("", taskHandlers.ListTasks)
			tasks.GET("/:id", taskHandlers.GetTask)
			tasks.PUT("/:id", taskHandlers.UpdateTask)
			tasks.PUT("/:id/status", taskHandlers.UpdateTaskStatus)
			tasks.PUT("/batch/status", taskHandlers.BatchUpdateTaskStatus)
			tasks.DELETE("/:id", taskHandlers.DeleteTask)
			tasks.GET("/:id/git-diff", taskHandlers.GetTaskGitDiff)
			tasks.GET("/:id/git-diff/file", taskHandlers.GetTaskGitDiffFile)
			tasks.POST("/:id/push", taskHandlers.PushTaskBranch)
		}

		conversations := api.Group("/conversations")
		{
			conversations.POST("", taskConvHandlers.CreateConversation)
			conversations.GET("", taskConvHandlers.ListConversations)
			conversations.GET("/latest", taskConvHandlers.GetLatestConversation)
			conversations.GET("/:id", taskConvHandlers.GetConversation)
			conversations.PUT("/:id", taskConvHandlers.UpdateConversation)
			conversations.DELETE("/:id", taskConvHandlers.DeleteConversation)
			conversations.GET("/:id/git-diff", taskConvHandlers.GetConversationGitDiff)
			conversations.GET("/:id/git-diff/file", taskConvHandlers.GetConversationGitDiffFile)
		}

		results := api.Group("/conversation-results")
		{
			results.GET("", taskConvResultHandlers.ListResultsByTaskID)
			results.GET("/by-project", taskConvResultHandlers.ListResultsByProjectID)
			results.GET("/:id", taskConvResultHandlers.GetResult)
			results.GET("/by-conversation/:conversation_id", taskConvResultHandlers.GetResultByConversationID)
			results.PUT("/:id", taskConvResultHandlers.UpdateResult)
			results.DELETE("/:id", taskConvResultHandlers.DeleteResult)
		}

		stats := api.Group("/stats")
		{
			stats.GET("/tasks/:task_id", taskConvResultHandlers.GetTaskStats)
			stats.GET("/projects/:project_id", taskConvResultHandlers.GetProjectStats)
		}

		api.GET("/task-conversations/:conversationId/execution-log", taskExecLogHandlers.GetExecutionLog)
		api.POST("/task-conversations/:conversationId/execution/cancel", taskExecLogHandlers.CancelExecution)
		api.POST("/task-conversations/:conversationId/execution/retry", taskExecLogHandlers.RetryExecution)

		devEnvs := api.Group("/dev-environments")
		{
			devEnvs.POST("", devEnvHandlers.CreateEnvironment)
			devEnvs.GET("", devEnvHandlers.ListEnvironments)
			devEnvs.GET("/available-types", devEnvHandlers.GetAvailableTypes)
			devEnvs.GET("/:id", devEnvHandlers.GetEnvironment)
			devEnvs.PUT("/:id", devEnvHandlers.UpdateEnvironment)
			devEnvs.DELETE("/:id", devEnvHandlers.DeleteEnvironment)
			devEnvs.GET("/:id/env-vars", devEnvHandlers.GetEnvironmentVars)
			devEnvs.PUT("/:id/env-vars", devEnvHandlers.UpdateEnvironmentVars)
		}

		systemConfigs := api.Group("/system-configs")
		{
			systemConfigs.GET("", systemConfigHandlers.ListAllConfigs)
			systemConfigs.PUT("", systemConfigHandlers.BatchUpdateConfigs)
		}
	}
}
