package routes

import (
	"embed"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
	"xsha-backend/config"
	"xsha-backend/handlers"
	"xsha-backend/middleware"
	"xsha-backend/services"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes(r *gin.Engine, cfg *config.Config, authService services.AuthService, adminService services.AdminService, authHandlers *handlers.AuthHandlers, adminHandlers *handlers.AdminHandlers, adminAvatarHandlers *handlers.AdminAvatarHandlers, gitCredHandlers *handlers.GitCredentialHandlers, projectHandlers *handlers.ProjectHandlers, operationLogHandlers *handlers.AdminOperationLogHandlers, devEnvHandlers *handlers.DevEnvironmentHandlers, taskHandlers *handlers.TaskHandlers, taskConvHandlers *handlers.TaskConversationHandlers, taskExecLogHandlers *handlers.TaskExecutionLogHandlers, attachmentHandlers *handlers.TaskConversationAttachmentHandlers, systemConfigHandlers *handlers.SystemConfigHandlers, dashboardHandlers *handlers.DashboardHandlers, staticFiles *embed.FS) {
	r.Use(middleware.I18nMiddleware())
	r.Use(middleware.ErrorHandlerMiddleware())

	r.NoMethod(middleware.MethodNotAllowedHandler())

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/health", handlers.HealthHandler)

	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/login", middleware.LoginRateLimitMiddleware(), authHandlers.LoginHandler)
	}

	api := r.Group("/api/v1")
	api.Use(middleware.AuthMiddlewareWithService(authService, adminService, cfg))

	api.Use(middleware.OperationLogMiddleware(operationLogHandlers.OperationLogService))

	{
		api.GET("/user/current", authHandlers.CurrentUserHandler)
		api.PUT("/user/change-password", authHandlers.ChangeOwnPasswordHandler)
		api.PUT("/user/update-avatar", authHandlers.UpdateOwnAvatarHandler)
		api.POST("/auth/logout", authHandlers.LogoutHandler)

		admin := api.Group("/admin")
		admin.Use(middleware.RequireSuperAdmin())
		{
			admin.GET("/login-logs", authHandlers.GetLoginLogsHandler)

			admin.GET("/operation-logs", operationLogHandlers.GetOperationLogs)
			admin.GET("/operation-logs/:id", operationLogHandlers.GetOperationLog)
			admin.GET("/operation-stats", operationLogHandlers.GetOperationStats)

			// Admin user management
			admin.POST("/users", adminHandlers.CreateAdminHandler)
			admin.GET("/users", adminHandlers.ListAdminsHandler)
			admin.GET("/users/:id", adminHandlers.GetAdminHandler)
			admin.PUT("/users/:id", adminHandlers.UpdateAdminHandler)
			admin.DELETE("/users/:id", adminHandlers.DeleteAdminHandler)
			admin.PUT("/users/:id/password", adminHandlers.ChangePasswordHandler)
			admin.PUT("/avatar/:uuid", adminAvatarHandlers.UpdateAdminAvatarHandler)

			// Admin avatar management
			admin.POST("/avatar/upload", adminAvatarHandlers.UploadAvatarHandler)
			
			// Role management
			admin.GET("/roles", adminHandlers.GetAvailableRolesHandler)
		}

		// Avatar preview (public endpoint, no auth required)
		r.GET("/api/v1/admin/avatar/preview/:uuid", adminAvatarHandlers.PreviewAvatarHandler)

		gitCreds := api.Group("/credentials")
		{
			gitCreds.POST("", middleware.RequireAdminOrSuperAdmin(), gitCredHandlers.CreateCredential)
			gitCreds.GET("", gitCredHandlers.ListCredentials)
			gitCreds.GET("/:id", gitCredHandlers.GetCredential)
			gitCreds.PUT("/:id", middleware.RequirePermission("credential", "update"), gitCredHandlers.UpdateCredential)
			gitCreds.DELETE("/:id", middleware.RequirePermission("credential", "delete"), gitCredHandlers.DeleteCredential)
		}

		projects := api.Group("/projects")
		{
			projects.POST("", middleware.RequireAdminOrSuperAdmin(), projectHandlers.CreateProject)
			projects.GET("", projectHandlers.ListProjects)
			projects.POST("/parse-url", projectHandlers.ParseRepositoryURL)
			projects.POST("/branches", projectHandlers.FetchRepositoryBranches)
			projects.POST("/validate-access", projectHandlers.ValidateRepositoryAccess)
			projects.GET("/credentials", projectHandlers.GetCompatibleCredentials)
			projects.GET("/:id", projectHandlers.GetProject)
			projects.PUT("/:id", middleware.RequirePermission("project", "update"), projectHandlers.UpdateProject)
			projects.DELETE("/:id", middleware.RequirePermission("project", "delete"), projectHandlers.DeleteProject)
			projects.GET("/:id/kanban", taskHandlers.GetKanbanTasks)
		}

		tasks := api.Group("/tasks")
		{
			tasks.POST("", taskHandlers.CreateTask)
			tasks.GET("", taskHandlers.ListTasks)
			tasks.GET("/:id", taskHandlers.GetTask)
			tasks.PUT("/:id", middleware.RequirePermission("task", "update"), taskHandlers.UpdateTask)
			tasks.PUT("/:id/status", middleware.RequirePermission("task", "update"), taskHandlers.UpdateTaskStatus)
			tasks.PUT("/batch/status", taskHandlers.BatchUpdateTaskStatus)
			tasks.DELETE("/:id", middleware.RequirePermission("task", "delete"), taskHandlers.DeleteTask)
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
			conversations.GET("/:id/details", taskConvHandlers.GetConversationDetails)
			conversations.PUT("/:id", middleware.RequirePermission("conversation", "update"), taskConvHandlers.UpdateConversation)
			conversations.DELETE("/:id", middleware.RequirePermission("conversation", "delete"), taskConvHandlers.DeleteConversation)
			conversations.GET("/:id/git-diff", taskConvHandlers.GetConversationGitDiff)
			conversations.GET("/:id/git-diff/file", taskConvHandlers.GetConversationGitDiffFile)
			conversations.GET("/:id/logs/stream", taskConvHandlers.StreamConversationLogs)
		}

		attachments := api.Group("/attachments")
		{
			attachments.POST("/upload", attachmentHandlers.UploadAttachment)
			attachments.GET("", attachmentHandlers.GetConversationAttachments)

			attachments.GET("/:id", attachmentHandlers.GetAttachment)
			attachments.GET("/:id/download", attachmentHandlers.DownloadAttachment)
			attachments.GET("/:id/preview", attachmentHandlers.PreviewAttachment)
			attachments.DELETE("/:id", attachmentHandlers.DeleteAttachment)
		}

		api.GET("/task-conversations/:conversationId/execution-log", taskExecLogHandlers.GetExecutionLog)
		api.POST("/task-conversations/:conversationId/execution/cancel", taskExecLogHandlers.CancelExecution)
		api.POST("/task-conversations/:conversationId/execution/retry", taskExecLogHandlers.RetryExecution)

		devEnvs := api.Group("/environments")
		{
			devEnvs.POST("", middleware.RequireAdminOrSuperAdmin(), devEnvHandlers.CreateEnvironment)
			devEnvs.GET("", devEnvHandlers.ListEnvironments)
			devEnvs.GET("/available-images", devEnvHandlers.GetAvailableImages)
			devEnvs.GET("/:id", devEnvHandlers.GetEnvironment)
			devEnvs.PUT("/:id", middleware.RequirePermission("environment", "update"), devEnvHandlers.UpdateEnvironment)
			devEnvs.DELETE("/:id", middleware.RequirePermission("environment", "delete"), devEnvHandlers.DeleteEnvironment)
		}

		systemConfigs := api.Group("/settings")
		systemConfigs.Use(middleware.RequireSuperAdmin())
		{
			systemConfigs.GET("", systemConfigHandlers.ListAllConfigs)
			systemConfigs.PUT("", systemConfigHandlers.BatchUpdateConfigs)
		}

		dashboard := api.Group("/dashboard")
		{
			dashboard.GET("/stats", dashboardHandlers.GetDashboardStats)
			dashboard.GET("/recent-tasks", dashboardHandlers.GetRecentTasks)
		}
	}

	// Setup static file serving for frontend
	setupStaticRoutes(r, staticFiles)
}

func setupStaticRoutes(r *gin.Engine, embeddedFiles *embed.FS) {
	// Try to get embedded filesystem first
	var staticFS fs.FS
	var assetsFS fs.FS
	var err error

	if embeddedFiles != nil {
		staticFS, err = fs.Sub(*embeddedFiles, "static")
		if err != nil {
			staticFS = nil
		} else {
			// Create assets subdirectory filesystem
			assetsFS, err = fs.Sub(staticFS, "assets")
			if err != nil {
				staticFS = nil
				assetsFS = nil
			}
		}
	}

	// If embed fails or is nil, fallback to file system
	if staticFS == nil || assetsFS == nil {
		setupFallbackStaticRoutes(r)
		return
	}

	// Serve static assets from embedded assets filesystem
	r.StaticFS("/assets", http.FS(assetsFS))

	// Serve individual static files
	serveSingleFile(r, "/favicon.ico", staticFS)
	serveSingleFile(r, "/vite.svg", staticFS)

	// Serve index.html for all non-API routes (SPA support)
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// Skip API routes and static assets
		if strings.HasPrefix(path, "/api") ||
			strings.HasPrefix(path, "/assets") ||
			strings.HasPrefix(path, "/swagger") ||
			strings.HasPrefix(path, "/health") ||
			path == "/favicon.ico" ||
			path == "/vite.svg" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
			return
		}

		// For all other routes, serve the React app
		indexData, err := staticFS.Open("index.html")
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Frontend not built"})
			return
		}
		defer indexData.Close()

		stat, err := indexData.Stat()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get file info"})
			return
		}

		c.DataFromReader(http.StatusOK, stat.Size(), "text/html; charset=utf-8", indexData, nil)
	})
}

// Helper function to serve single files from embedded filesystem
func serveSingleFile(r *gin.Engine, path string, staticFS fs.FS) {
	r.GET(path, func(c *gin.Context) {
		file, err := staticFS.Open(strings.TrimPrefix(path, "/"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}
		defer file.Close()

		stat, err := file.Stat()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get file info"})
			return
		}

		// Determine content type based on file extension
		contentType := "application/octet-stream"
		ext := filepath.Ext(path)
		switch ext {
		case ".ico":
			contentType = "image/x-icon"
		case ".svg":
			contentType = "image/svg+xml"
		}

		c.DataFromReader(http.StatusOK, stat.Size(), contentType, file, nil)
	})
}

// Fallback function for development mode when static files are not embedded
func setupFallbackStaticRoutes(r *gin.Engine) {
	// Static files directory
	staticDir := "static"

	// Serve static assets (CSS, JS, images, etc.)
	r.Static("/assets", staticDir+"/assets")
	r.StaticFile("/favicon.ico", staticDir+"/favicon.ico")
	r.StaticFile("/vite.svg", staticDir+"/vite.svg")

	// Serve index.html for all non-API routes (SPA support)
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// Skip API routes and static assets
		if strings.HasPrefix(path, "/api") ||
			strings.HasPrefix(path, "/assets") ||
			strings.HasPrefix(path, "/swagger") ||
			strings.HasPrefix(path, "/health") ||
			path == "/favicon.ico" ||
			path == "/vite.svg" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
			return
		}

		// For all other routes, serve the React app
		indexPath := filepath.Join(staticDir, "index.html")
		c.File(indexPath)
	})
}
