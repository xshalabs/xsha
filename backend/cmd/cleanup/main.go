package main

import (
	"xsha-backend/config"
	"xsha-backend/database"
	"xsha-backend/repository"
	"xsha-backend/services"
	"xsha-backend/utils"
)

func main() {
	cfg := config.Load()

	logConfig := utils.LogConfig{
		Level:  cfg.LogLevel,
		Format: cfg.LogFormat,
		Output: cfg.LogOutput,
	}

	if err := utils.InitLogger(logConfig); err != nil {
		utils.Error("Failed to initialize logger", "error", err.Error())
		return
	}

	logger := utils.WithFields(map[string]interface{}{
		"component": "cleanup",
	})

	dbManager, err := database.NewDatabaseManager(cfg)
	if err != nil {
		logger.Error("Failed to initialize database",
			"error", err.Error(),
		)
		return
	}
	defer dbManager.Close()

	tokenRepo := repository.NewTokenBlacklistRepository(dbManager.GetDB())
	loginLogRepo := repository.NewLoginLogRepository(dbManager.GetDB())
	adminOperationLogRepo := repository.NewAdminOperationLogRepository(dbManager.GetDB())

	adminOperationLogService := services.NewAdminOperationLogService(adminOperationLogRepo)
	authService := services.NewAuthService(tokenRepo, loginLogRepo, adminOperationLogService, cfg)
	loginLogService := services.NewLoginLogService(loginLogRepo)

	logger.Info("Starting cleanup tasks...")

	if err := authService.CleanExpiredTokens(); err != nil {
		logger.Error("Failed to clean expired tokens",
			"error", err.Error(),
		)
	} else {
		logger.Info("Expired tokens cleaned successfully")
	}

	if err := loginLogService.CleanOldLogs(30); err != nil {
		logger.Error("Failed to clean old login logs",
			"error", err.Error(),
		)
	} else {
		logger.Info("Old login logs cleaned successfully")
	}

	logger.Info("All cleanup tasks completed")
}
