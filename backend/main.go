package main

import (
	"log"
	"sleep0-backend/config"
	"sleep0-backend/database"
	"sleep0-backend/i18n"
	"sleep0-backend/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize internationalization
	i18nInstance := i18n.GetInstance()

	// Load configuration
	cfg := config.Load()

	// Initialize database
	database.InitDatabase()

	// Set gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create gin engine
	r := gin.Default()

	// Setup routes
	routes.SetupRoutes(r)

	// Start server
	log.Print(i18nInstance.GetMessage("zh-CN", "server.starting"))
	log.Printf("Server starting on port %s", cfg.Port)

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("%s: %v", i18nInstance.GetMessage("zh-CN", "server.start_failed"), err)
	}
}
