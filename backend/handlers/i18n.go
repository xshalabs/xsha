package handlers

import (
	"net/http"
	"xsha-backend/i18n"
	"xsha-backend/middleware"

	"github.com/gin-gonic/gin"
)

// GetLanguagesHandler handles getting supported language list
// @Summary 获取支持的语言列表
// @Description 获取系统支持的所有语言列表及当前语言
// @Tags 国际化
// @Accept json
// @Produce json
// @Success 200 {object} object{languages=[]string,current=string} "支持的语言列表"
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
// @Summary 设置语言偏好
// @Description 设置用户的语言偏好
// @Tags 国际化
// @Accept json
// @Produce json
// @Param languageData body object{language=string} true "语言设置"
// @Success 200 {object} object{language=string} "语言设置成功"
// @Failure 400 {object} object{error=string} "请求参数错误"
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
