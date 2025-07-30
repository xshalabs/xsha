// @title XSHA Backend API
// @version 1.0
// @description XSHA Backend API service, providing user authentication, project management, Git credential management and other functions
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"
	"xsha-backend/config"
	"xsha-backend/database"
	"xsha-backend/handlers"
	"xsha-backend/i18n"
	"xsha-backend/repository"
	"xsha-backend/routes"
	"xsha-backend/scheduler"
	"xsha-backend/services"
	"xsha-backend/services/executor"
	"xsha-backend/utils"

	_ "xsha-backend/docs" // Auto-generated swagger docs

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize internationalization
	i18nInstance := i18n.GetInstance()

	// Load configuration
	cfg := config.Load()

	// Initialize database with new architecture
	dbManager, err := database.NewDatabaseManager(cfg)
	if err != nil {
		utils.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer dbManager.Close()

	// Initialize database (for backward compatibility)
	database.InitDatabase()

	// Initialize repositories
	tokenRepo := repository.NewTokenBlacklistRepository(dbManager.GetDB())
	loginLogRepo := repository.NewLoginLogRepository(dbManager.GetDB())
	adminOperationLogRepo := repository.NewAdminOperationLogRepository(dbManager.GetDB())
	gitCredRepo := repository.NewGitCredentialRepository(dbManager.GetDB())
	projectRepo := repository.NewProjectRepository(dbManager.GetDB())
	devEnvRepo := repository.NewDevEnvironmentRepository(dbManager.GetDB())
	taskRepo := repository.NewTaskRepository(dbManager.GetDB())
	taskConvRepo := repository.NewTaskConversationRepository(dbManager.GetDB())
	execLogRepo := repository.NewTaskExecutionLogRepository(dbManager.GetDB())
	taskConvResultRepo := repository.NewTaskConversationResultRepository(dbManager.GetDB())
	systemConfigRepo := repository.NewSystemConfigRepository(dbManager.GetDB())

	// Initialize log broadcaster
	logBroadcaster := services.NewLogBroadcaster()
	logBroadcaster.Start()

	// Initialize workspace manager
	workspaceManager := utils.NewWorkspaceManager(cfg.WorkspaceBaseDir)

	// Initialize services
	loginLogService := services.NewLoginLogService(loginLogRepo)
	adminOperationLogService := services.NewAdminOperationLogService(adminOperationLogRepo)
	authService := services.NewAuthService(tokenRepo, loginLogRepo, adminOperationLogService, cfg)
	gitCredService := services.NewGitCredentialService(gitCredRepo, projectRepo, cfg)
	systemConfigService := services.NewSystemConfigService(systemConfigRepo)
	devEnvService := services.NewDevEnvironmentService(devEnvRepo, taskRepo, systemConfigService)
	projectService := services.NewProjectService(projectRepo, gitCredRepo, gitCredService, taskRepo, cfg)
	taskService := services.NewTaskService(taskRepo, projectRepo, devEnvRepo, workspaceManager, cfg, gitCredService)
	taskConvService := services.NewTaskConversationService(taskConvRepo, taskRepo, execLogRepo)
	taskConvResultService := services.NewTaskConversationResultService(taskConvResultRepo, taskConvRepo, taskRepo, projectRepo)
	aiTaskExecutor := executor.NewAITaskExecutorService(taskConvRepo, taskRepo, execLogRepo, taskConvResultRepo, gitCredService, taskConvResultService, taskService, systemConfigService, cfg, logBroadcaster)

	// Initialize scheduler
	taskProcessor := scheduler.NewTaskProcessor(aiTaskExecutor)

	// Parse scheduler interval
	schedulerInterval, err := time.ParseDuration(cfg.SchedulerInterval)
	if err != nil {
		utils.Warn("Failed to parse scheduler interval, using default value of 30 seconds", "error", err)
		schedulerInterval = 30 * time.Second
	}

	schedulerManager := scheduler.NewSchedulerManager(taskProcessor, schedulerInterval)

	// Initialize handlers
	authHandlers := handlers.NewAuthHandlers(authService, loginLogService)
	adminOperationLogHandlers := handlers.NewAdminOperationLogHandlers(adminOperationLogService)
	gitCredHandlers := handlers.NewGitCredentialHandlers(gitCredService)
	projectHandlers := handlers.NewProjectHandlers(projectService)
	devEnvHandlers := handlers.NewDevEnvironmentHandlers(devEnvService)
	taskHandlers := handlers.NewTaskHandlers(taskService, taskConvService, projectService)
	taskConvHandlers := handlers.NewTaskConversationHandlers(taskConvService)
	taskConvResultHandlers := handlers.NewTaskConversationResultHandlers(taskConvResultService)
	taskExecLogHandlers := handlers.NewTaskExecutionLogHandlers(aiTaskExecutor)
	sseLogHandlers := handlers.NewSSELogHandlers(logBroadcaster)
	systemConfigHandlers := handlers.NewSystemConfigHandlers(systemConfigService)

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

	// Setup routes - Pass all handler instances
	routes.SetupRoutes(r, cfg, authService, authHandlers, gitCredHandlers, projectHandlers, adminOperationLogHandlers, devEnvHandlers, taskHandlers, taskConvHandlers, taskConvResultHandlers, taskExecLogHandlers, sseLogHandlers, systemConfigHandlers)

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

		os.Exit(0)
	}()

	// Start server
	utils.Info(i18nInstance.GetMessage("zh-CN", "server.starting"))
	utils.Info("Server starting on port", "port", cfg.Port)

	if err := r.Run(":" + cfg.Port); err != nil {
		utils.Error(i18nInstance.GetMessage("zh-CN", "server.start_failed"), "error", err)
		os.Exit(1)
	}
}
