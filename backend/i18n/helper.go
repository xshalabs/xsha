package i18n

import "github.com/gin-gonic/gin"

// Helper 国际化助手结构
type Helper struct {
	lang string
}

// NewHelper 创建新的助手实例
func NewHelper(lang string) *Helper {
	return &Helper{lang: lang}
}

// NewHelperFromContext 从Gin上下文创建助手实例
func NewHelperFromContext(c *gin.Context) *Helper {
	lang := "zh-CN" // 默认语言
	if l, exists := c.Get("lang"); exists {
		if langStr, ok := l.(string); ok {
			lang = langStr
		}
	}
	return &Helper{lang: lang}
}

// T 翻译函数
func (h *Helper) T(key string, args ...interface{}) string {
	return T(h.lang, key, args...)
}

// GetLang 获取当前语言
func (h *Helper) GetLang() string {
	return h.lang
}

// SetLang 设置语言
func (h *Helper) SetLang(lang string) {
	h.lang = lang
}

// Response 国际化响应助手
func (h *Helper) Response(c *gin.Context, statusCode int, messageKey string, data ...interface{}) {
	response := gin.H{
		"message": h.T(messageKey),
	}

	// 如果有额外数据，添加到响应中
	if len(data) > 0 {
		if dataMap, ok := data[0].(gin.H); ok {
			for key, value := range dataMap {
				response[key] = value
			}
		} else if dataMap, ok := data[0].(map[string]interface{}); ok {
			for key, value := range dataMap {
				response[key] = value
			}
		}
	}

	c.JSON(statusCode, response)
}

// ErrorResponse 错误响应助手
func (h *Helper) ErrorResponse(c *gin.Context, statusCode int, errorKey string, details ...string) {
	response := gin.H{
		"error": h.T(errorKey),
	}

	if len(details) > 0 {
		response["details"] = details[0]
	}

	c.JSON(statusCode, response)
}
