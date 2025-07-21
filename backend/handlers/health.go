package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler 健康检查处理
func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "服务运行正常",
	})
}
