package middleware

import (
	"fmt"
	"time"
	"xsha-backend/utils"

	"github.com/gin-gonic/gin"
)

// LoggerMiddleware 基于slog的HTTP请求日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	logger := utils.GetLogger()

	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 使用slog记录请求日志
		logger.Info("HTTP Request",
			"client_ip", param.ClientIP,
			"timestamp", param.TimeStamp.Format(time.RFC3339),
			"method", param.Method,
			"path", param.Path,
			"protocol", param.Request.Proto,
			"status_code", param.StatusCode,
			"latency", param.Latency.String(),
			"user_agent", param.Request.UserAgent(),
			"error", param.ErrorMessage,
		)

		// 返回空字符串避免重复输出
		return ""
	})
}

// RequestLogMiddleware 详细的请求日志中间件
func RequestLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 生成请求ID用于追踪
		requestID := fmt.Sprintf("%d-%s", start.UnixNano(), c.ClientIP())
		c.Set("request_id", requestID)

		// 创建带请求上下文的日志记录器
		logger := utils.WithFields(map[string]interface{}{
			"request_id": requestID,
			"method":     c.Request.Method,
			"path":       path,
			"client_ip":  c.ClientIP(),
			"user_agent": c.GetHeader("User-Agent"),
		})

		// 记录请求开始
		if raw != "" {
			path = path + "?" + raw
		}

		logger.Info("Request started",
			"full_path", path,
		)

		// 处理请求
		c.Next()

		// 记录请求完成
		end := time.Now()
		latency := end.Sub(start)
		statusCode := c.Writer.Status()

		// 根据状态码选择日志级别
		logFunc := logger.Info
		if statusCode >= 400 && statusCode < 500 {
			logFunc = logger.Warn
		} else if statusCode >= 500 {
			logFunc = logger.Error
		}

		logFunc("Request completed",
			"status_code", statusCode,
			"latency", latency.String(),
			"response_size", c.Writer.Size(),
		)

		// 如果有错误，记录错误详情
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				logger.Error("Request error",
					"error", err.Error(),
					"type", err.Type,
				)
			}
		}
	}
}
