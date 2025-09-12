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
	"xsha-backend/config"
	"xsha-backend/internal/app"
	"xsha-backend/utils"

	_ "xsha-backend/docs"
)

//go:embed static/*
var StaticFiles embed.FS

func main() {
	cfg := config.Load()
	
	application := app.New(cfg, &StaticFiles)
	
	if err := application.Initialize(); err != nil {
		utils.Error("Failed to initialize application", "error", err)
		os.Exit(1)
	}

	go application.WaitForShutdown()

	if err := application.Run(); err != nil {
		utils.Error("Application run failed", "error", err)
		utils.Sync()
		os.Exit(1)
	}
}
