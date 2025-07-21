package middleware

import (
	"bytes"
	"io"
	"sleep0-backend/services"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// OperationLogMiddleware 操作日志记录中间件
func OperationLogMiddleware(operationLogService services.AdminOperationLogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户信息
		username, exists := c.Get("username")
		if !exists {
			// 如果没有用户信息，跳过日志记录
			c.Next()
			return
		}

		// 获取客户端信息
		clientIP := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")
		method := c.Request.Method
		path := c.Request.URL.Path

		// 只记录需要记录的操作
		if !shouldLogOperation(method, path) {
			c.Next()
			return
		}

		// 读取请求体（如果需要）
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 处理请求
		c.Next()

		// 记录操作日志
		go func() {
			// 确定操作类型和资源
			operation, resource, resourceID, description := determineOperationInfo(method, path, c.Param("id"))

			// 判断操作是否成功
			success := c.Writer.Status() < 400
			errorMsg := ""
			if !success {
				// 从响应中提取错误信息（如果需要）
				errorMsg = "HTTP " + strconv.Itoa(c.Writer.Status())
			}

			// 记录日志
			switch operation {
			case "create":
				operationLogService.LogCreate(username.(string), resource, resourceID,
					description, clientIP, userAgent, path, success, errorMsg)
			case "update":
				operationLogService.LogUpdate(username.(string), resource, resourceID,
					description, clientIP, userAgent, path, success, errorMsg)
			case "delete":
				operationLogService.LogDelete(username.(string), resource, resourceID,
					description, clientIP, userAgent, path, success, errorMsg)
			case "read":
				if success { // 只记录成功的查询操作
					operationLogService.LogRead(username.(string), resource, resourceID,
						description, clientIP, userAgent, path)
				}
			}
		}()
	}
}

// shouldLogOperation 判断是否需要记录操作日志
func shouldLogOperation(method, path string) bool {
	// 跳过健康检查和静态资源
	if strings.HasPrefix(path, "/health") ||
		strings.HasPrefix(path, "/static") ||
		strings.HasPrefix(path, "/api/v1/languages") ||
		strings.HasPrefix(path, "/api/v1/language") {
		return false
	}

	// 只记录API操作
	if !strings.HasPrefix(path, "/api/v1/") {
		return false
	}

	// 记录增删改查操作
	return method == "POST" || method == "PUT" || method == "DELETE" || method == "GET"
}

// determineOperationInfo 根据HTTP方法和路径确定操作信息
func determineOperationInfo(method, path, id string) (operation, resource, resourceID, description string) {
	// 提取资源类型
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	if len(pathParts) >= 3 {
		resource = pathParts[2] // api/v1/[resource]
	}

	// 根据HTTP方法确定操作类型
	switch method {
	case "POST":
		operation = "create"
		if strings.Contains(path, "/toggle") {
			operation = "update"
			description = "切换状态"
		} else if strings.Contains(path, "/use") {
			operation = "read"
			description = "使用资源"
		} else {
			description = "创建" + resource
		}
	case "PUT":
		operation = "update"
		description = "更新" + resource
		resourceID = id
	case "DELETE":
		operation = "delete"
		description = "删除" + resource
		resourceID = id
	case "GET":
		operation = "read"
		if id != "" {
			description = "查看" + resource + "详情"
			resourceID = id
		} else {
			description = "查看" + resource + "列表"
		}
	}

	// 特殊路径处理
	if strings.Contains(path, "admin") {
		resource = "admin"
	}
	if strings.Contains(path, "auth") {
		resource = "auth"
	}
	if strings.Contains(path, "git-credentials") {
		resource = "git-credential"
	}
	if strings.Contains(path, "projects") {
		resource = "project"
	}

	return
}
