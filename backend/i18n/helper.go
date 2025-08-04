package i18n

import (
	appErrors "xsha-backend/errors"

	"github.com/gin-gonic/gin"
)

type Helper struct {
	lang string
}

func NewHelper(lang string) *Helper {
	return &Helper{lang: lang}
}

func (h *Helper) T(key string, args ...interface{}) string {
	return T(h.lang, key, args...)
}

func (h *Helper) GetLang() string {
	return h.lang
}

func MapErrorToLocalizedMessage(err error, lang string) string {
	if err == nil {
		return ""
	}

	if i18nErr, ok := err.(*appErrors.I18nError); ok {
		if i18nErr.Params != nil {
			return T(lang, i18nErr.Key, i18nErr.Params)
		}
		return T(lang, i18nErr.Key)
	}

	return err.Error()
}

func (h *Helper) ErrorResponseFromError(c *gin.Context, statusCode int, err error) {
	if i18nErr, ok := err.(*appErrors.I18nError); ok {
		response := gin.H{
			"error": h.T(i18nErr.Key, i18nErr.Params),
		}
		if i18nErr.Details != "" {
			response["details"] = i18nErr.Details
		}
		c.JSON(statusCode, response)
		return
	}

	c.JSON(statusCode, gin.H{
		"error": MapErrorToLocalizedMessage(err, h.GetLang()),
	})
}

func MapErrorToI18nKey(err error, lang string) string {
	return MapErrorToLocalizedMessage(err, lang)
}
