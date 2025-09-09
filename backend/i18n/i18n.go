package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"sync"
	"xsha-backend/utils"
)

//go:embed locales/*.json
var localesFS embed.FS

type I18n struct {
	mu          sync.RWMutex
	messages    map[string]map[string]string
	defaultLang string
}

var (
	instance *I18n
	once     sync.Once
)

func GetInstance() *I18n {
	once.Do(func() {
		instance = &I18n{
			messages:    make(map[string]map[string]string),
			defaultLang: "zh-CN",
		}
		instance.loadMessages()
	})
	return instance
}

func (i *I18n) loadMessages() {
	files, err := fs.ReadDir(localesFS, "locales")
	if err != nil {
		utils.Error("Failed to read embedded language files", "error", err)
		return
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			lang := file.Name()[:len(file.Name())-5]
			i.loadEmbeddedMessageFile(filepath.Join("locales", file.Name()), lang)
		}
	}
}

func (i *I18n) loadEmbeddedMessageFile(filename, lang string) {
	data, err := localesFS.ReadFile(filename)
	if err != nil {
		utils.Error("Failed to read embedded language file", "filename", filename, "error", err)
		return
	}

	var messages map[string]string
	if err := json.Unmarshal(data, &messages); err != nil {
		utils.Error("Failed to parse embedded language file", "filename", filename, "error", err)
		return
	}

	i.mu.Lock()
	i.messages[lang] = messages
	i.mu.Unlock()
}

func (i *I18n) GetMessage(lang, key string) string {
	i.mu.RLock()
	defer i.mu.RUnlock()

	if messages, exists := i.messages[lang]; exists {
		if message, found := messages[key]; found {
			return message
		}
	}

	if lang != i.defaultLang {
		if messages, exists := i.messages[i.defaultLang]; exists {
			if message, found := messages[key]; found {
				return message
			}
		}
	}

	return key
}

func (i *I18n) GetMessageWithArgs(lang, key string, args ...interface{}) string {
	message := i.GetMessage(lang, key)
	if len(args) > 0 {
		return fmt.Sprintf(message, args...)
	}
	return message
}

func (i *I18n) GetSupportedLanguages() []string {
	i.mu.RLock()
	defer i.mu.RUnlock()

	var langs []string
	for lang := range i.messages {
		langs = append(langs, lang)
	}
	return langs
}

func T(lang, key string, args ...interface{}) string {
	return GetInstance().GetMessageWithArgs(lang, key, args...)
}
