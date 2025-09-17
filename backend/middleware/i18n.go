package middleware

import (
	"strings"
	"xsha-backend/database"
	"xsha-backend/i18n"
	"xsha-backend/services"

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
	// Priority 1: Query parameter
	if lang := c.Query("lang"); lang != "" {
		if isValidLanguage(lang) {
			// Update admin language preference if authenticated and different
			updateAdminLanguagePreference(c, lang)
			return lang
		}
	}

	// Priority 2: Admin's stored language preference (if authenticated)
	if admin, exists := c.Get("admin"); exists {
		if adminObj, ok := admin.(*database.Admin); ok && adminObj.Lang != "" {
			if isValidLanguage(adminObj.Lang) {
				return adminObj.Lang
			}
		}
	}

	// Priority 3: Accept-Language header
	if acceptLang := c.GetHeader("Accept-Language"); acceptLang != "" {
		lang := parseAcceptLanguage(acceptLang)
		if isValidLanguage(lang) {
			// Update admin language preference if authenticated and different
			updateAdminLanguagePreference(c, lang)
			return lang
		}
	}

	// Priority 4: Default
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

func updateAdminLanguagePreference(c *gin.Context, lang string) {
	// Check if admin is authenticated
	admin, adminExists := c.Get("admin")
	adminService, serviceExists := c.Get("adminService")

	if !adminExists || !serviceExists {
		return
	}

	adminObj, ok := admin.(*database.Admin)
	if !ok {
		return
	}

	adminSvc, ok := adminService.(services.AdminService)
	if !ok {
		return
	}

	// Only update if the language is different from current preference
	if adminObj.Lang != lang {
		// Update asynchronously to avoid blocking the request
		go func() {
			if err := adminSvc.UpdateAdminLanguage(adminObj.ID, lang); err != nil {
				// Log error but don't fail the request
				// Note: We could use a logger here, but since this is middleware
				// we'll keep it silent to avoid dependency issues
			}
		}()
	}
}

func GetLangFromContext(c *gin.Context) string {
	if lang, exists := c.Get("lang"); exists {
		if langStr, ok := lang.(string); ok {
			return langStr
		}
	}
	return "en-US"
}
