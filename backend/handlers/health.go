package handlers

import (
	"net/http"
	"sleep0-backend/i18n"
	"sleep0-backend/middleware"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check
// @Summary 健康检查
// @Description 检查服务器状态
// @Tags 系统
// @Accept json
// @Produce json
// @Success 200 {object} object{status=string,message=string,lang=string} "服务器状态正常"
// @Router /health [get]
func HealthHandler(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": i18n.T(lang, "health.status_ok"),
		"lang":    lang,
	})
}
