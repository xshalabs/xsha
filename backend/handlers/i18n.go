package handlers

import (
	"net/http"
	"xsha-backend/i18n"
	"xsha-backend/middleware"

	"github.com/gin-gonic/gin"
)

// GetLanguagesHandler handles getting supported language list
// @Summary Get supported language list
// @Description Get all supported languages and current language from the system
// @Tags Internationalization
// @Accept json
// @Produce json
// @Success 200 {object} object{languages=[]string,current=string} "Supported language list"
// @Router /languages [get]
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
// @Summary Set language preference
// @Description Set user's language preference
// @Tags Internationalization
// @Accept json
// @Produce json
// @Param languageData body object{language=string} true "Language setting"
// @Success 200 {object} object{language=string} "Language set successfully"
// @Failure 400 {object} object{error=string} "Request parameter error"
// @Router /language [post]
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
