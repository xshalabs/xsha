package main

import (
	"sleep0-backend/config"
	"sleep0-backend/database"
	"sleep0-backend/repository"
	"sleep0-backend/services"
	"sleep0-backend/utils"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化日志
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

	// 初始化数据库（新架构）
	dbManager, err := database.NewDatabaseManager(cfg)
	if err != nil {
		logger.Error("Failed to initialize database",
			"error", err.Error(),
		)
		return
	}
	defer dbManager.Close()

	// 初始化仓库层
	tokenRepo := repository.NewTokenBlacklistRepository(dbManager.GetDB())
	loginLogRepo := repository.NewLoginLogRepository(dbManager.GetDB())
	adminOperationLogRepo := repository.NewAdminOperationLogRepository(dbManager.GetDB())

	// 初始化服务层
	adminOperationLogService := services.NewAdminOperationLogService(adminOperationLogRepo)
	authService := services.NewAuthService(tokenRepo, loginLogRepo, adminOperationLogService, cfg)
	loginLogService := services.NewLoginLogService(loginLogRepo)

	logger.Info("Starting cleanup tasks...")

	// 清理过期的黑名单Token
	if err := authService.CleanExpiredTokens(); err != nil {
		logger.Error("Failed to clean expired tokens",
			"error", err.Error(),
		)
	} else {
		logger.Info("Expired tokens cleaned successfully")
	}

	// 清理30天前的登录日志
	if err := loginLogService.CleanOldLogs(30); err != nil {
		logger.Error("Failed to clean old login logs",
			"error", err.Error(),
		)
	} else {
		logger.Info("Old login logs cleaned successfully")
	}

	logger.Info("All cleanup tasks completed")
}
