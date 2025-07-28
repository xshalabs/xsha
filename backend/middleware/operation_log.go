package middleware

import (
	"bytes"
	"io"
	"strconv"
	"strings"
	"xsha-backend/services"

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
		} else if strings.Contains(path, "/parse-url") {
			operation = "read"
			description = "解析仓库URL"
		} else if strings.Contains(path, "/branches") {
			operation = "read"
			description = "获取仓库分支列表"
		} else if strings.Contains(path, "/validate-access") {
			operation = "read"
			description = "验证仓库访问权限"
		} else {
			description = "创建" + getResourceDisplayName(resource)
		}
	case "PUT":
		operation = "update"
		description = "更新" + getResourceDisplayName(resource)
		resourceID = id
	case "DELETE":
		operation = "delete"
		description = "删除" + getResourceDisplayName(resource)
		resourceID = id
	case "GET":
		operation = "read"
		if id != "" {
			description = "查看" + getResourceDisplayName(resource) + "详情"
			resourceID = id
		} else {
			description = "查看" + getResourceDisplayName(resource) + "列表"
		}
	}

	// 特殊路径处理和资源名称规范化
	resource = normalizeResourceName(path, resource)

	return
}

// getResourceDisplayName 获取资源的显示名称
func getResourceDisplayName(resource string) string {
	displayNames := map[string]string{
		"admin":            "管理员",
		"auth":             "认证",
		"git-credentials":  "Git凭据",
		"projects":         "项目",
		"tasks":            "任务",
		"conversations":    "任务对话",
		"dev-environments": "开发环境",
		"operation-logs":   "操作日志",
		"login-logs":       "登录日志",
		"logs":             "日志",
		"user":             "用户",
	}

	if displayName, exists := displayNames[resource]; exists {
		return displayName
	}
	return resource
}

// normalizeResourceName 规范化资源名称
func normalizeResourceName(path, resource string) string {
	// 特殊路径处理，按优先级顺序
	if strings.Contains(path, "/admin/operation-logs") {
		return "operation-logs"
	}
	if strings.Contains(path, "/admin/login-logs") {
		return "login-logs"
	}
	if strings.Contains(path, "/admin") {
		return "admin"
	}
	if strings.Contains(path, "/auth") {
		return "auth"
	}
	if strings.Contains(path, "/git-credentials") {
		return "git-credentials"
	}
	if strings.Contains(path, "/projects") {
		return "projects"
	}
	if strings.Contains(path, "/tasks") {
		return "tasks"
	}
	if strings.Contains(path, "/conversations") {
		return "conversations"
	}
	if strings.Contains(path, "/dev-environments") {
		return "dev-environments"
	}
	if strings.Contains(path, "/logs") {
		return "logs"
	}
	if strings.Contains(path, "/user") {
		return "user"
	}

	return resource
}
