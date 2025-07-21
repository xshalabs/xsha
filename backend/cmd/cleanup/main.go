package main

import (
	"log"
	"sleep0-backend/config"
	"sleep0-backend/database"
)

func main() {
	// 加载配置
	config.Load()

	// 初始化数据库
	database.InitDatabase()

	// 清理30天前的登录日志
	if err := database.CleanOldLoginLogs(30); err != nil {
		log.Fatalf("清理旧登录日志失败: %v", err)
	}

	log.Println("登录日志清理完成")
}
