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
	gitCredRepo := repository.NewGitCredentialRepository(dbManager.GetDB())
	projectRepo := repository.NewProjectRepository(dbManager.GetDB())

	// Initialize services
	authService := services.NewAuthService(tokenRepo, loginLogRepo, cfg)
	loginLogService := services.NewLoginLogService(loginLogRepo)
	gitCredService := services.NewGitCredentialService(gitCredRepo, cfg)
	projectService := services.NewProjectService(projectRepo, gitCredRepo, gitCredService)

	// Initialize handlers
	authHandlers := handlers.NewAuthHandlers(authService, loginLogService)
	gitCredHandlers := handlers.NewGitCredentialHandlers(gitCredService)
	projectHandlers := handlers.NewProjectHandlers(projectService)

	// Set global handlers for backward compatibility (只保留auth相关)
	handlers.SetAuthHandlers(authHandlers)

	// Set gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create gin engine
	r := gin.Default()

	// Setup routes - 传递所有处理器实例
	routes.SetupRoutes(r, authService, gitCredHandlers, projectHandlers)

	// Start server
	log.Print(i18nInstance.GetMessage("zh-CN", "server.starting"))
	log.Printf("Server starting on port %s", cfg.Port)

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("%s: %v", i18nInstance.GetMessage("zh-CN", "server.start_failed"), err)
	}
}
