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
	"sleep0-backend/config"
	"sleep0-backend/database"
	"sleep0-backend/handlers"
	"sleep0-backend/i18n"
	"sleep0-backend/repository"
	"sleep0-backend/routes"
	"sleep0-backend/services"

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

	// Initialize services
	authService := services.NewAuthService(tokenRepo, loginLogRepo, cfg)
	loginLogService := services.NewLoginLogService(loginLogRepo)
	adminOperationLogService := services.NewAdminOperationLogService(adminOperationLogRepo)
	gitCredService := services.NewGitCredentialService(gitCredRepo, cfg)
	projectService := services.NewProjectService(projectRepo, gitCredRepo, gitCredService, cfg)
	devEnvService := services.NewDevEnvironmentService(devEnvRepo)
	taskService := services.NewTaskService(taskRepo, projectRepo, devEnvRepo)
	taskConvService := services.NewTaskConversationService(taskConvRepo, taskRepo)

	// Initialize handlers
	authHandlers := handlers.NewAuthHandlers(authService, loginLogService)
	adminOperationLogHandlers := handlers.NewAdminOperationLogHandlers(adminOperationLogService)
	gitCredHandlers := handlers.NewGitCredentialHandlers(gitCredService)
	projectHandlers := handlers.NewProjectHandlers(projectService)
	devEnvHandlers := handlers.NewDevEnvironmentHandlers(devEnvService)
	taskHandlers := handlers.NewTaskHandlers(taskService)
	taskConvHandlers := handlers.NewTaskConversationHandlers(taskConvService)

	// Set gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create gin engine
	r := gin.Default()

	// Setup routes - 传递所有处理器实例
	routes.SetupRoutes(r, authService, authHandlers, gitCredHandlers, projectHandlers, adminOperationLogHandlers, devEnvHandlers, taskHandlers, taskConvHandlers)

	// Start server
	log.Print(i18nInstance.GetMessage("zh-CN", "server.starting"))
	log.Printf("Server starting on port %s", cfg.Port)

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("%s: %v", i18nInstance.GetMessage("zh-CN", "server.start_failed"), err)
	}
}
