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
	// 初始化内置消息（避免文件丢失时的fallback）
	i.initBuiltinMessages()

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

// initBuiltinMessages 初始化内置消息
func (i *I18n) initBuiltinMessages() {
	// 中文消息
	i.messages["zh-CN"] = map[string]string{
		"login.success":              "登录成功",
		"login.failed":               "用户名或密码错误",
		"login.invalid_request":      "请求数据格式错误",
		"login.token_generate_error": "生成token失败",
		"login.rate_limit":           "登录尝试过于频繁，请稍后再试",
		"logout.success":             "登出成功",
		"logout.failed":              "登出失败",
		"logout.invalid_token":       "无效的token",
		"logout.token_expired":       "token已失效，请重新登录",
		"auth.unauthorized":          "未授权访问",
		"auth.invalid_token":         "无效的token",
		"auth.token_blacklisted":     "token已失效，请重新登录",
		"auth.get_token_exp_error":   "获取token过期时间失败",
		"user.get_info_error":        "无法获取用户信息",
		"user.authenticated":         "已认证",
		"common.internal_error":      "内部服务器错误",
		"health.status_ok":           "服务运行正常",
	}

	// 英文消息
	i.messages["en-US"] = map[string]string{
		"login.success":              "Login successful",
		"login.failed":               "Invalid username or password",
		"login.invalid_request":      "Invalid request data format",
		"login.token_generate_error": "Failed to generate token",
		"login.rate_limit":           "Too many login attempts, please try again later",
		"logout.success":             "Logout successful",
		"logout.failed":              "Logout failed",
		"logout.invalid_token":       "Invalid token",
		"logout.token_expired":       "Token has expired, please login again",
		"auth.unauthorized":          "Unauthorized access",
		"auth.invalid_token":         "Invalid token",
		"auth.token_blacklisted":     "Token has been invalidated, please login again",
		"auth.get_token_exp_error":   "Failed to get token expiration time",
		"user.get_info_error":        "Unable to get user information",
		"user.authenticated":         "Authenticated",
		"common.internal_error":      "Internal server error",
		"health.status_ok":           "Service is running normally",
	}
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
