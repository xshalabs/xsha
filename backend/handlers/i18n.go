package handlers

import (
	"net/http"
	"sleep0-backend/i18n"
	"sleep0-backend/middleware"

	"github.com/gin-gonic/gin"
)

// GetLanguagesHandler 获取支持的语言列表
func GetLanguagesHandler(c *gin.Context) {
	lang := middleware.GetLangFromContext(c)
	helper := i18n.NewHelper(lang)

	languages := i18n.GetInstance().GetSupportedLanguages()

	helper.Response(c, http.StatusOK, "common.success", gin.H{
		"languages": languages,
		"current":   lang,
	})
}

// SetLanguageHandler 设置语言偏好（示例）
func SetLanguageHandler(c *gin.Context) {
	helper := i18n.NewHelperFromContext(c)

	var langData struct {
		Language string `json:"language" binding:"required"`
	}

	if err := c.ShouldBindJSON(&langData); err != nil {
		helper.ErrorResponse(c, http.StatusBadRequest, "validation.invalid_format", err.Error())
		return
	}

	// 验证语言是否支持
	supportedLangs := i18n.GetInstance().GetSupportedLanguages()
	isSupported := false
	for _, lang := range supportedLangs {
		if lang == langData.Language {
			isSupported = true
			break
		}
	}

	if !isSupported {
		helper.ErrorResponse(c, http.StatusBadRequest, "validation.invalid_format", "不支持的语言")
		return
	}

	// 在实际项目中，这里可以保存用户的语言偏好到数据库或cookies
	helper.SetLang(langData.Language)

	helper.Response(c, http.StatusOK, "common.success", gin.H{
		"language": langData.Language,
	})
}
