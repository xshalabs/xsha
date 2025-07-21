package handlers

import (
	"net/http"
	"sleep0-backend/i18n"
	"sleep0-backend/middleware"

	"github.com/gin-gonic/gin"
)

// GetLanguagesHandler handles getting supported language list
func GetLanguagesHandler(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	helper := i18n.NewHelper(lang)

	languages := i18n.GetInstance().GetSupportedLanguages()

	helper.Response(c, http.StatusOK, "common.success", gin.H{
		"languages": languages,
		"current":   lang,
	})
}

// SetLanguageHandler handles setting language preference (example)
func SetLanguageHandler(c *gin.Context) {
	helper := i18n.NewHelperFromContext(c)

	var langData struct {
		Language string `json:"language" binding:"required"`
	}

	if err := c.ShouldBindJSON(&langData); err != nil {
		helper.ErrorResponse(c, http.StatusBadRequest, "validation.invalid_format", err.Error())
		return
	}

	// Validate if the language is supported
	supportedLangs := i18n.GetInstance().GetSupportedLanguages()
	isSupported := false
	for _, lang := range supportedLangs {
		if lang == langData.Language {
			isSupported = true
			break
		}
	}

	if !isSupported {
		helper.ErrorResponse(c, http.StatusBadRequest, "validation.invalid_format", "Unsupported language")
		return
	}

	// In real projects, here you can save user's language preference to database or cookies
	helper.SetLang(langData.Language)

	helper.Response(c, http.StatusOK, "common.success", gin.H{
		"language": langData.Language,
	})
}
