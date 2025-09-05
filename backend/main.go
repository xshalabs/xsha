// @title XSHA Backend API
// @version 1.0
// @description XSHA Backend API service, providing user authentication, project management, Git credential management and other functions

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

package main

import (
	"embed"
	"os"
	"os/signal"
	"syscall"
	"time"
	"xsha-backend/config"
	"xsha-backend/database"
	"xsha-backend/handlers"
	"xsha-backend/repository"
	"xsha-backend/routes"
	"xsha-backend/scheduler"
	"xsha-backend/services"
	"xsha-backend/services/executor"
	"xsha-backend/utils"

	_ "xsha-backend/docs"

	"github.com/gin-gonic/gin"
)

//go:embed static/*
var StaticFiles embed.FS

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database with new architecture
	dbManager, err := database.NewDatabaseManager(cfg)
	if err != nil {
		utils.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer dbManager.Close()

	// Initialize repositories
	tokenRepo := repository.NewTokenBlacklistRepository(dbManager.GetDB())
	loginLogRepo := repository.NewLoginLogRepository(dbManager.GetDB())
	adminRepo := repository.NewAdminRepository(dbManager.GetDB())
	adminAvatarRepo := repository.NewAdminAvatarRepository(dbManager.GetDB())
	adminOperationLogRepo := repository.NewAdminOperationLogRepository(dbManager.GetDB())
	gitCredRepo := repository.NewGitCredentialRepository(dbManager.GetDB())
	projectRepo := repository.NewProjectRepository(dbManager.GetDB())
	devEnvRepo := repository.NewDevEnvironmentRepository(dbManager.GetDB())
	taskRepo := repository.NewTaskRepository(dbManager.GetDB())
	taskConvRepo := repository.NewTaskConversationRepository(dbManager.GetDB())
	execLogRepo := repository.NewTaskExecutionLogRepository(dbManager.GetDB())
	taskConvResultRepo := repository.NewTaskConversationResultRepository(dbManager.GetDB())
	taskConvAttachmentRepo := repository.NewTaskConversationAttachmentRepository(dbManager.GetDB())
	systemConfigRepo := repository.NewSystemConfigRepository(dbManager.GetDB())
	dashboardRepo := repository.NewDashboardRepository(dbManager.GetDB())

	// Initialize services
	loginLogService := services.NewLoginLogService(loginLogRepo)
	adminOperationLogService := services.NewAdminOperationLogService(adminOperationLogRepo)
	adminService := services.NewAdminService(adminRepo)
	adminAvatarService := services.NewAdminAvatarService(adminAvatarRepo, adminRepo, cfg)
	authService := services.NewAuthService(tokenRepo, loginLogRepo, adminOperationLogService, adminService, adminRepo, cfg)

	// Set up circular dependency - adminService needs authService for session invalidation
	adminService.SetAuthService(authService)
	gitCredService := services.NewGitCredentialService(gitCredRepo, projectRepo, cfg)
	systemConfigService := services.NewSystemConfigService(systemConfigRepo)
	dashboardService := services.NewDashboardService(dashboardRepo)

	// Get git clone timeout from system config
	gitCloneTimeout, err := systemConfigService.GetGitCloneTimeout()
	if err != nil {
		utils.Error("Failed to get git clone timeout from system config, using default", "error", err)
		gitCloneTimeout = 5 * time.Minute
	}

	// Initialize workspace manager
	workspaceManager := utils.NewWorkspaceManager(cfg.WorkspaceBaseDir, gitCloneTimeout)
	devEnvService := services.NewDevEnvironmentService(devEnvRepo, taskRepo, systemConfigService, cfg)

	// Set up circular dependencies - adminService needs devEnvService and gitCredService for permission checks
	adminService.SetDevEnvironmentService(devEnvService)
	adminService.SetGitCredentialService(gitCredService)
	projectService := services.NewProjectService(projectRepo, gitCredRepo, gitCredService, taskRepo, systemConfigService, cfg)
	taskService := services.NewTaskService(taskRepo, projectRepo, devEnvRepo, taskConvRepo, execLogRepo, taskConvResultRepo, taskConvAttachmentRepo, workspaceManager, cfg, gitCredService, systemConfigService)
	taskConvResultService := services.NewTaskConversationResultService(taskConvResultRepo, taskConvRepo, taskRepo, projectRepo)
	taskConvAttachmentService := services.NewTaskConversationAttachmentService(taskConvAttachmentRepo, cfg)
	taskConvService := services.NewTaskConversationService(taskConvRepo, taskRepo, execLogRepo, taskConvResultRepo, taskService, taskConvAttachmentService, workspaceManager)

	// Create shared execution manager
	maxConcurrency := 5
	if cfg.MaxConcurrentTasks > 0 {
		maxConcurrency = cfg.MaxConcurrentTasks
	}
	executionManager := executor.NewExecutionManager(maxConcurrency)

	// Initialize services with shared execution manager
	aiTaskExecutor := executor.NewAITaskExecutorServiceWithManager(taskConvRepo, taskRepo, execLogRepo, taskConvResultRepo, gitCredService, taskConvResultService, taskService, systemConfigService, taskConvAttachmentService, cfg, executionManager)
	logStreamingService := executor.NewLogStreamingService(taskConvRepo, execLogRepo, executionManager)

	// Initialize scheduler
	taskProcessor := scheduler.NewTaskProcessor(aiTaskExecutor)
	schedulerManager := scheduler.NewSchedulerManager(taskProcessor, cfg.SchedulerIntervalDuration)

	// Initialize handlers
	authHandlers := handlers.NewAuthHandlers(authService, loginLogService, adminService, adminAvatarService)
	adminHandlers := handlers.NewAdminHandlers(adminService)
	adminAvatarHandlers := handlers.NewAdminAvatarHandlers(adminAvatarService, adminService)
	adminOperationLogHandlers := handlers.NewAdminOperationLogHandlers(adminOperationLogService)
	gitCredHandlers := handlers.NewGitCredentialHandlers(gitCredService)
	projectHandlers := handlers.NewProjectHandlers(projectService)
	devEnvHandlers := handlers.NewDevEnvironmentHandlers(devEnvService)
	taskHandlers := handlers.NewTaskHandlers(taskService, taskConvService, projectService)
	taskConvHandlers := handlers.NewTaskConversationHandlers(taskConvService, logStreamingService, aiTaskExecutor)
	taskConvAttachmentHandlers := handlers.NewTaskConversationAttachmentHandlers(taskConvAttachmentService)
	systemConfigHandlers := handlers.NewSystemConfigHandlers(systemConfigService)
	dashboardHandlers := handlers.NewDashboardHandlers(dashboardService)

	// Set gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create gin engine
	r := gin.Default()

	// Initialize system configuration default values
	if err := systemConfigService.InitializeDefaultConfigs(); err != nil {
		utils.Error("Failed to initialize default system configurations", "error", err)
		os.Exit(1)
	}

	// Create attachment storage directory
	if err := os.MkdirAll(cfg.AttachmentsDir, 0755); err != nil {
		utils.Error("Failed to create attachment storage directory", "directory", cfg.AttachmentsDir, "error", err)
		os.Exit(1)
	}
	utils.Info("Attachment storage directory initialized", "directory", cfg.AttachmentsDir)

	// Create avatar storage directory
	if err := os.MkdirAll(cfg.AvatarsDir, 0755); err != nil {
		utils.Error("Failed to create avatar storage directory", "directory", cfg.AvatarsDir, "error", err)
		os.Exit(1)
	}
	utils.Info("Avatar storage directory initialized", "directory", cfg.AvatarsDir)

	// Create workspace base directory
	if err := os.MkdirAll(cfg.WorkspaceBaseDir, 0755); err != nil {
		utils.Error("Failed to create workspace base directory", "directory", cfg.WorkspaceBaseDir, "error", err)
		os.Exit(1)
	}
	utils.Info("Workspace base directory initialized", "directory", cfg.WorkspaceBaseDir)

	// Create dev sessions directory
	if err := os.MkdirAll(cfg.DevSessionsDir, 0755); err != nil {
		utils.Error("Failed to create dev sessions directory", "directory", cfg.DevSessionsDir, "error", err)
		os.Exit(1)
	}
	utils.Info("Dev sessions directory initialized", "directory", cfg.DevSessionsDir)

	// Initialize default admin user
	if err := adminService.InitializeDefaultAdmin(); err != nil {
		utils.Error("Failed to initialize default admin user", "error", err)
		os.Exit(1)
	}

	// Setup routes - Pass all handler instances including static files
	routes.SetupRoutes(r, cfg, authService, adminService, authHandlers, adminHandlers, adminAvatarHandlers, gitCredHandlers, projectHandlers, adminOperationLogHandlers, devEnvHandlers, taskHandlers, taskConvHandlers, taskConvAttachmentHandlers, systemConfigHandlers, dashboardHandlers, &StaticFiles, projectService, taskService, taskConvService, gitCredService, devEnvService)

	// Start scheduler
	if err := schedulerManager.Start(); err != nil {
		utils.Error("Failed to start scheduler", "error", err)
		os.Exit(1)
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		utils.Info("Received shutdown signal, stopping service...")

		// Stop scheduler
		if err := schedulerManager.Stop(); err != nil {
			utils.Error("Failed to stop scheduler", "error", err)
		}

		// Sync logger before exit
		if err := utils.Sync(); err != nil {
			utils.Error("Failed to sync logger", "error", err)
		}

		os.Exit(0)
	}()

	// Start server
	utils.Info("Server starting...")
	utils.Info("Server starting on port", "port", cfg.Port)

	if err := r.Run(":" + cfg.Port); err != nil {
		utils.Error("Server start failed", "error", err)
		utils.Sync()
		os.Exit(1)
	}
}
