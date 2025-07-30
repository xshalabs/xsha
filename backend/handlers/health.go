package handlers

import (
	"net/http"
	"xsha-backend/i18n"
	"xsha-backend/middleware"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check
// @Summary Health check
// @Description Check server status
// @Tags System
// @Accept json
// @Produce json
// @Success 200 {object} object{status=string,message=string,lang=string} "Server status is normal"
// @Router /health [get]
func HealthHandler(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": i18n.T(lang, "health.status_ok"),
		"lang":    lang,
	})
}
