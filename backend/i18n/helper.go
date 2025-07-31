package i18n

import "github.com/gin-gonic/gin"

type Helper struct {
	lang string
}

func NewHelper(lang string) *Helper {
	return &Helper{lang: lang}
}

func NewHelperFromContext(c *gin.Context) *Helper {
	lang := "zh-CN"
	if l, exists := c.Get("lang"); exists {
		if langStr, ok := l.(string); ok {
			lang = langStr
		}
	}
	return &Helper{lang: lang}
}

func (h *Helper) T(key string, args ...interface{}) string {
	return T(h.lang, key, args...)
}

func (h *Helper) GetLang() string {
	return h.lang
}

func (h *Helper) SetLang(lang string) {
	h.lang = lang
}

func (h *Helper) Response(c *gin.Context, statusCode int, messageKey string, data ...interface{}) {
	response := gin.H{
		"message": h.T(messageKey),
	}

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

func (h *Helper) ErrorResponse(c *gin.Context, statusCode int, errorKey string, details ...string) {
	response := gin.H{
		"error": h.T(errorKey),
	}

	if len(details) > 0 {
		response["details"] = details[0]
	}

	c.JSON(statusCode, response)
}
