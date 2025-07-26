// @title Sleep0 Backend API
// @version 1.0
// @description Sleep0 Backend API服务，提供用户认证、项目管理、Git凭据管理等功能
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
	"log"
	"os"
	"os/signal"
	"sleep0-backend/config"
	"sleep0-backend/database"
	"sleep0-backend/handlers"
	"sleep0-backend/i18n"
	"sleep0-backend/repository"
	"sleep0-backend/routes"
	"sleep0-backend/scheduler"
	"sleep0-backend/services"
	"sleep0-backend/utils"
	"syscall"
	"time"

	_ "sleep0-backend/docs" // 自动生成的swagger docs

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
		log.Fatalf("Failed to initialize database: %v", err)
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

	// Initialize log broadcaster
	logBroadcaster := services.NewLogBroadcaster()
	logBroadcaster.Start()

	// Initialize workspace manager
	workspaceManager := utils.NewWorkspaceManager(cfg.WorkspaceBaseDir)

	// Initialize services
	authService := services.NewAuthService(tokenRepo, loginLogRepo, cfg)
	loginLogService := services.NewLoginLogService(loginLogRepo)
	adminOperationLogService := services.NewAdminOperationLogService(adminOperationLogRepo)
	gitCredService := services.NewGitCredentialService(gitCredRepo, cfg)
	projectService := services.NewProjectService(projectRepo, gitCredRepo, gitCredService, cfg)
	devEnvService := services.NewDevEnvironmentService(devEnvRepo)
	taskService := services.NewTaskService(taskRepo, projectRepo, devEnvRepo, workspaceManager)
	taskConvService := services.NewTaskConversationService(taskConvRepo, taskRepo, execLogRepo)
	aiTaskExecutor := services.NewAITaskExecutorService(taskConvRepo, taskRepo, execLogRepo, gitCredService, cfg, logBroadcaster)

	// Initialize scheduler
	taskProcessor := scheduler.NewTaskProcessor(aiTaskExecutor)

	// 解析定时器间隔
	schedulerInterval, err := time.ParseDuration(cfg.SchedulerInterval)
	if err != nil {
		log.Printf("解析定时器间隔失败，使用默认值30秒: %v", err)
		schedulerInterval = 30 * time.Second
	}

	schedulerManager := scheduler.NewSchedulerManager(taskProcessor, schedulerInterval)

	// Initialize handlers
	authHandlers := handlers.NewAuthHandlers(authService, loginLogService)
	adminOperationLogHandlers := handlers.NewAdminOperationLogHandlers(adminOperationLogService)
	gitCredHandlers := handlers.NewGitCredentialHandlers(gitCredService)
	projectHandlers := handlers.NewProjectHandlers(projectService)
	devEnvHandlers := handlers.NewDevEnvironmentHandlers(devEnvService)
	taskHandlers := handlers.NewTaskHandlers(taskService, taskConvService)
	taskConvHandlers := handlers.NewTaskConversationHandlers(taskConvService)
	taskExecLogHandlers := handlers.NewTaskExecutionLogHandlers(aiTaskExecutor)
	sseLogHandlers := handlers.NewSSELogHandlers(logBroadcaster)

	// Set gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create gin engine
	r := gin.Default()

	// Setup routes - 传递所有处理器实例
	routes.SetupRoutes(r, authService, authHandlers, gitCredHandlers, projectHandlers, adminOperationLogHandlers, devEnvHandlers, taskHandlers, taskConvHandlers, taskExecLogHandlers, sseLogHandlers)

	// Start scheduler
	if err := schedulerManager.Start(); err != nil {
		log.Fatalf("启动定时器失败: %v", err)
	}

	// 设置优雅关闭
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("收到关闭信号，正在停止服务...")

		// 停止定时器
		if err := schedulerManager.Stop(); err != nil {
			log.Printf("停止定时器失败: %v", err)
		}

		os.Exit(0)
	}()

	// Start server
	log.Print(i18nInstance.GetMessage("zh-CN", "server.starting"))
	log.Printf("Server starting on port %s", cfg.Port)

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("%s: %v", i18nInstance.GetMessage("zh-CN", "server.start_failed"), err)
	}
}
