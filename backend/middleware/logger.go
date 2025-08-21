package middleware

import (
	"fmt"
	"time"
	"xsha-backend/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func LoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		utils.Info("HTTP Request",
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

		return ""
	})
}

func RequestLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := utils.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		requestID := fmt.Sprintf("%d-%s", start.UnixNano(), c.ClientIP())
		c.Set("request_id", requestID)

		// Create logger with request fields using zap
		logger := utils.GetLogger().With(
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.GetHeader("User-Agent")),
		)

		if raw != "" {
			path = path + "?" + raw
		}

		logger.Info("Request started",
			zap.String("full_path", path),
		)

		c.Next()

		end := utils.Now()
		latency := end.Sub(start)
		statusCode := c.Writer.Status()

		// Log request completion with appropriate level
		if statusCode >= 500 {
			logger.Error("Request completed",
				zap.Int("status_code", statusCode),
				zap.String("latency", latency.String()),
				zap.Int("response_size", c.Writer.Size()),
			)
		} else if statusCode >= 400 {
			logger.Warn("Request completed",
				zap.Int("status_code", statusCode),
				zap.String("latency", latency.String()),
				zap.Int("response_size", c.Writer.Size()),
			)
		} else {
			logger.Info("Request completed",
				zap.Int("status_code", statusCode),
				zap.String("latency", latency.String()),
				zap.Int("response_size", c.Writer.Size()),
			)
		}

		// Log any errors that occurred during request processing
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				logger.Error("Request error",
					zap.String("error", err.Error()),
					zap.Uint64("type", uint64(err.Type)),
				)
			}
		}
	}
}