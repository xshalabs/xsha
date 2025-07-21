package main

import (
	"log"
	"sleep0-backend/config"
	"sleep0-backend/database"
	"sleep0-backend/repository"
	"sleep0-backend/services"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化数据库（新架构）
	dbManager, err := database.NewDatabaseManager(cfg)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer dbManager.Close()

	// 初始化仓库层
	tokenRepo := repository.NewTokenBlacklistRepository(dbManager.GetDB())
	loginLogRepo := repository.NewLoginLogRepository(dbManager.GetDB())

	// 初始化服务层
	authService := services.NewAuthService(tokenRepo, loginLogRepo, cfg)
	loginLogService := services.NewLoginLogService(loginLogRepo)

	// 清理过期的黑名单Token
	if err := authService.CleanExpiredTokens(); err != nil {
		log.Printf("清理过期Token失败: %v", err)
	} else {
		log.Println("过期Token清理完成")
	}

	// 清理30天前的登录日志
	if err := loginLogService.CleanOldLogs(30); err != nil {
		log.Printf("清理旧登录日志失败: %v", err)
	} else {
		log.Println("登录日志清理完成")
	}

	log.Println("清理任务全部完成")
}
