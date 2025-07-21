package middleware

import (
	"sleep0-backend/i18n"
	"strings"

	"github.com/gin-gonic/gin"
)

// I18nMiddleware internationalization middleware
func I18nMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := detectLanguage(c)
		c.Set("lang", lang)
		c.Next()
	}
}

// detectLanguage detects request language
func detectLanguage(c *gin.Context) string {
	// Priority:
	// 1. lang parameter in URL
	// 2. Accept-Language in Header
	// 3. Default language

	// 1. Check URL parameter
	if lang := c.Query("lang"); lang != "" {
		if isValidLanguage(lang) {
			return lang
		}
	}

	// 2. Check Header
	if acceptLang := c.GetHeader("Accept-Language"); acceptLang != "" {
		lang := parseAcceptLanguage(acceptLang)
		if isValidLanguage(lang) {
			return lang
		}
	}

	// 3. Return default language
	return "zh-CN"
}

// parseAcceptLanguage parses Accept-Language header
func parseAcceptLanguage(acceptLang string) string {
	// Simple parsing, take the first language identifier
	languages := strings.Split(acceptLang, ",")
	if len(languages) > 0 {
		lang := strings.TrimSpace(strings.Split(languages[0], ";")[0])
		// Normalize language code
		return normalizeLanguage(lang)
	}
	return ""
}

// normalizeLanguage normalizes language code
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

// isValidLanguage checks if the language is supported
func isValidLanguage(lang string) bool {
	supportedLangs := i18n.GetInstance().GetSupportedLanguages()
	for _, supportedLang := range supportedLangs {
		if lang == supportedLang {
			return true
		}
	}
	return false
}

// GetLangFromContext gets language from context
func GetLangFromContext(c *gin.Context) string {
	if lang, exists := c.Get("lang"); exists {
		if langStr, ok := lang.(string); ok {
			return langStr
		}
	}
	return "zh-CN" // Default language
}
