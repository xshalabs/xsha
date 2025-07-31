package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"xsha-backend/utils"
)

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
	langDir := "i18n/locales"
	if _, err := os.Stat(langDir); os.IsNotExist(err) {
		utils.Warn("Language file directory does not exist, using built-in messages", "langDir", langDir)
		return
	}

	files, err := os.ReadDir(langDir)
	if err != nil {
		utils.Error("Failed to read language file directory", "error", err)
		return
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			lang := file.Name()[:len(file.Name())-5]
			i.loadMessageFile(filepath.Join(langDir, file.Name()), lang)
		}
	}
}

func (i *I18n) loadMessageFile(filename, lang string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		utils.Error("Failed to read language file", "filename", filename, "error", err)
		return
	}

	var messages map[string]string
	if err := json.Unmarshal(data, &messages); err != nil {
		utils.Error("Failed to parse language file", "filename", filename, "error", err)
		return
	}

	i.mu.Lock()
	i.messages[lang] = messages
	i.mu.Unlock()

	utils.Info("Language file loaded", "filename", filename, "messageCount", len(messages))
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
