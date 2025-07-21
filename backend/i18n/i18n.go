package i18n

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// I18n internationalization structure
type I18n struct {
	mu          sync.RWMutex
	messages    map[string]map[string]string // [language code][message key]message content
	defaultLang string
}

var (
	instance *I18n
	once     sync.Once
)

// GetInstance gets singleton instance
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

// loadMessages loads message files
func (i *I18n) loadMessages() {
	// Try to load messages from file
	langDir := "i18n/locales"
	if _, err := os.Stat(langDir); os.IsNotExist(err) {
		log.Printf("Language file directory does not exist, using built-in messages: %s", langDir)
		return
	}

	files, err := os.ReadDir(langDir)
	if err != nil {
		log.Printf("Failed to read language file directory: %v", err)
		return
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			lang := file.Name()[:len(file.Name())-5] // Remove .json extension
			i.loadMessageFile(filepath.Join(langDir, file.Name()), lang)
		}
	}
}

// loadMessageFile loads message file for specified language
func (i *I18n) loadMessageFile(filename, lang string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Printf("Failed to read language file %s: %v", filename, err)
		return
	}

	var messages map[string]string
	if err := json.Unmarshal(data, &messages); err != nil {
		log.Printf("Failed to parse language file %s: %v", filename, err)
		return
	}

	i.mu.Lock()
	i.messages[lang] = messages
	i.mu.Unlock()

	log.Printf("Language file loaded: %s (%d messages)", filename, len(messages))
}

// GetMessage gets message for specified language
func (i *I18n) GetMessage(lang, key string) string {
	i.mu.RLock()
	defer i.mu.RUnlock()

	// Try to get message for specified language
	if messages, exists := i.messages[lang]; exists {
		if message, found := messages[key]; found {
			return message
		}
	}

	// Fall back to default language
	if lang != i.defaultLang {
		if messages, exists := i.messages[i.defaultLang]; exists {
			if message, found := messages[key]; found {
				return message
			}
		}
	}

	// If not found, return key itself
	return key
}

// GetMessageWithArgs gets message with arguments
func (i *I18n) GetMessageWithArgs(lang, key string, args ...interface{}) string {
	message := i.GetMessage(lang, key)
	if len(args) > 0 {
		return fmt.Sprintf(message, args...)
	}
	return message
}

// GetSupportedLanguages gets list of supported languages
func (i *I18n) GetSupportedLanguages() []string {
	i.mu.RLock()
	defer i.mu.RUnlock()

	var langs []string
	for lang := range i.messages {
		langs = append(langs, lang)
	}
	return langs
}

// T simplified translation function
func T(lang, key string, args ...interface{}) string {
	return GetInstance().GetMessageWithArgs(lang, key, args...)
}
