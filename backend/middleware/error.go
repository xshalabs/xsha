package middleware

import (
	"net/http"
	"xsha-backend/i18n"

	"github.com/gin-gonic/gin"
)

func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			lang := GetLangFromContext(c)

			err := c.Errors.Last()

			switch err.Type {
			case gin.ErrorTypeBind:
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   i18n.T(lang, "validation.invalid_format"),
					"details": err.Error(),
				})
			case gin.ErrorTypePublic:
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": i18n.T(lang, "common.internal_error"),
				})
			}
		}
	}
}

func NotFoundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := GetLangFromContext(c)
		c.JSON(http.StatusNotFound, gin.H{
			"error": i18n.T(lang, "api.not_found"),
		})
	}
}

func MethodNotAllowedHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := GetLangFromContext(c)
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"error": i18n.T(lang, "api.method_not_allowed"),
		})
	}
}
