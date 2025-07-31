package middleware

import (
	"strings"
	"xsha-backend/i18n"

	"github.com/gin-gonic/gin"
)

func I18nMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := detectLanguage(c)
		c.Set("lang", lang)
		c.Next()
	}
}

func detectLanguage(c *gin.Context) string {
	if lang := c.Query("lang"); lang != "" {
		if isValidLanguage(lang) {
			return lang
		}
	}

	if acceptLang := c.GetHeader("Accept-Language"); acceptLang != "" {
		lang := parseAcceptLanguage(acceptLang)
		if isValidLanguage(lang) {
			return lang
		}
	}

	return "en-US"
}

func parseAcceptLanguage(acceptLang string) string {
	languages := strings.Split(acceptLang, ",")
	if len(languages) > 0 {
		lang := strings.TrimSpace(strings.Split(languages[0], ";")[0])
		return normalizeLanguage(lang)
	}
	return ""
}

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

func isValidLanguage(lang string) bool {
	supportedLangs := i18n.GetInstance().GetSupportedLanguages()
	for _, supportedLang := range supportedLangs {
		if lang == supportedLang {
			return true
		}
	}
	return false
}

func GetLangFromContext(c *gin.Context) string {
	if lang, exists := c.Get("lang"); exists {
		if langStr, ok := lang.(string); ok {
			return langStr
		}
	}
	return "en-US"
}
