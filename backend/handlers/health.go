package handlers

import (
	"net/http"
	"sleep0-backend/i18n"
	"sleep0-backend/middleware"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check
func HealthHandler(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": i18n.T(lang, "health.status_ok"),
		"lang":    lang,
	})
}
