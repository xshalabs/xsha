package middleware

import (
	"fmt"
	"time"
	"xsha-backend/utils"

	"github.com/gin-gonic/gin"
)

func LoggerMiddleware() gin.HandlerFunc {
	logger := utils.GetLogger()

	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
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

		logger := utils.WithFields(map[string]interface{}{
			"request_id": requestID,
			"method":     c.Request.Method,
			"path":       path,
			"client_ip":  c.ClientIP(),
			"user_agent": c.GetHeader("User-Agent"),
		})

		if raw != "" {
			path = path + "?" + raw
		}

		logger.Info("Request started",
			"full_path", path,
		)

		c.Next()

		end := utils.Now()
		latency := end.Sub(start)
		statusCode := c.Writer.Status()

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
