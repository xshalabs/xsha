package middleware

import (
	"net/http"
	"sleep0-backend/i18n"

	"github.com/gin-gonic/gin"
)

// ErrorHandlerMiddleware 错误处理中间件
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 处理错误
		if len(c.Errors) > 0 {
			lang := GetLangFromContext(c)

			// 获取最后一个错误
			err := c.Errors.Last()

			// 根据错误类型返回相应的HTTP状态码和消息
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

// NotFoundHandler 404错误处理
func NotFoundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := GetLangFromContext(c)
		c.JSON(http.StatusNotFound, gin.H{
			"error": i18n.T(lang, "api.not_found"),
		})
	}
}

// MethodNotAllowedHandler 405错误处理
func MethodNotAllowedHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := GetLangFromContext(c)
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"error": i18n.T(lang, "api.method_not_allowed"),
		})
	}
}
