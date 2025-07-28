package middleware

import (
	"net/http"
	"xsha-backend/i18n"

	"github.com/gin-gonic/gin"
)

// ErrorHandlerMiddleware error handling middleware
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Handle errors
		if len(c.Errors) > 0 {
			lang := GetLangFromContext(c)

			// Get the last error
			err := c.Errors.Last()

			// Return appropriate HTTP status code and message based on error type
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

// NotFoundHandler handles 404 errors
func NotFoundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := GetLangFromContext(c)
		c.JSON(http.StatusNotFound, gin.H{
			"error": i18n.T(lang, "api.not_found"),
		})
	}
}

// MethodNotAllowedHandler handles 405 errors
func MethodNotAllowedHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := GetLangFromContext(c)
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"error": i18n.T(lang, "api.method_not_allowed"),
		})
	}
}
