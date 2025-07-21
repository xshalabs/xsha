package i18n

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// I18n 国际化结构
type I18n struct {
	mu          sync.RWMutex
	messages    map[string]map[string]string // [语言代码][消息键]消息内容
	defaultLang string
}

var (
	instance *I18n
	once     sync.Once
)

// GetInstance 获取单例实例
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

// loadMessages 加载消息文件
func (i *I18n) loadMessages() {
	// 尝试从文件加载消息
	langDir := "i18n/locales"
	if _, err := os.Stat(langDir); os.IsNotExist(err) {
		log.Printf("语言文件目录不存在，使用内置消息: %s", langDir)
		return
	}

	files, err := os.ReadDir(langDir)
	if err != nil {
		log.Printf("读取语言文件目录失败: %v", err)
		return
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			lang := file.Name()[:len(file.Name())-5] // 移除.json扩展名
			i.loadMessageFile(filepath.Join(langDir, file.Name()), lang)
		}
	}
}

// loadMessageFile 加载指定语言的消息文件
func (i *I18n) loadMessageFile(filename, lang string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Printf("读取语言文件失败 %s: %v", filename, err)
		return
	}

	var messages map[string]string
	if err := json.Unmarshal(data, &messages); err != nil {
		log.Printf("解析语言文件失败 %s: %v", filename, err)
		return
	}

	i.mu.Lock()
	i.messages[lang] = messages
	i.mu.Unlock()

	log.Printf("已加载语言文件: %s (%d 条消息)", filename, len(messages))
}

// GetMessage 获取指定语言的消息
func (i *I18n) GetMessage(lang, key string) string {
	i.mu.RLock()
	defer i.mu.RUnlock()

	// 尝试获取指定语言的消息
	if messages, exists := i.messages[lang]; exists {
		if message, found := messages[key]; found {
			return message
		}
	}

	// 回退到默认语言
	if lang != i.defaultLang {
		if messages, exists := i.messages[i.defaultLang]; exists {
			if message, found := messages[key]; found {
				return message
			}
		}
	}

	// 如果都没找到，返回key本身
	return key
}

// GetMessageWithArgs 获取带参数的消息
func (i *I18n) GetMessageWithArgs(lang, key string, args ...interface{}) string {
	message := i.GetMessage(lang, key)
	if len(args) > 0 {
		return fmt.Sprintf(message, args...)
	}
	return message
}

// GetSupportedLanguages 获取支持的语言列表
func (i *I18n) GetSupportedLanguages() []string {
	i.mu.RLock()
	defer i.mu.RUnlock()

	var langs []string
	for lang := range i.messages {
		langs = append(langs, lang)
	}
	return langs
}

// T 简化的翻译函数
func T(lang, key string, args ...interface{}) string {
	return GetInstance().GetMessageWithArgs(lang, key, args...)
}
