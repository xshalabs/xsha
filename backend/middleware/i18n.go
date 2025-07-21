package middleware

import (
	"sleep0-backend/i18n"
	"strings"

	"github.com/gin-gonic/gin"
)

// I18nMiddleware 国际化中间件
func I18nMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := detectLanguage(c)
		c.Set("lang", lang)
		c.Next()
	}
}

// detectLanguage 检测请求的语言
func detectLanguage(c *gin.Context) string {
	// 优先级：
	// 1. URL参数中的lang
	// 2. Header中的Accept-Language
	// 3. 默认语言

	// 1. 检查URL参数
	if lang := c.Query("lang"); lang != "" {
		if isValidLanguage(lang) {
			return lang
		}
	}

	// 2. 检查Header
	if acceptLang := c.GetHeader("Accept-Language"); acceptLang != "" {
		lang := parseAcceptLanguage(acceptLang)
		if isValidLanguage(lang) {
			return lang
		}
	}

	// 3. 返回默认语言
	return "zh-CN"
}

// parseAcceptLanguage 解析Accept-Language头
func parseAcceptLanguage(acceptLang string) string {
	// 简单解析，取第一个语言标识
	languages := strings.Split(acceptLang, ",")
	if len(languages) > 0 {
		lang := strings.TrimSpace(strings.Split(languages[0], ";")[0])
		// 标准化语言代码
		return normalizeLanguage(lang)
	}
	return ""
}

// normalizeLanguage 标准化语言代码
func normalizeLanguage(lang string) string {
	lang = strings.ToLower(lang)
	switch {
	case strings.HasPrefix(lang, "zh-cn"), strings.HasPrefix(lang, "zh_cn"), lang == "zh":
		return "zh-CN"
	case strings.HasPrefix(lang, "en-us"), strings.HasPrefix(lang, "en_us"), lang == "en":
		return "en-US"
	default:
		return lang
	}
}

// isValidLanguage 检查是否为支持的语言
func isValidLanguage(lang string) bool {
	supportedLangs := i18n.GetInstance().GetSupportedLanguages()
	for _, supportedLang := range supportedLangs {
		if lang == supportedLang {
			return true
		}
	}
	return false
}

// GetLangFromContext 从context中获取语言
func GetLangFromContext(c *gin.Context) string {
	if lang, exists := c.Get("lang"); exists {
		if langStr, ok := lang.(string); ok {
			return langStr
		}
	}
	return "zh-CN" // 默认语言
}
