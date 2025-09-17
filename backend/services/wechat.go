package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"xsha-backend/database"
	"xsha-backend/utils"
)

// WeChatConfig holds WeChat webhook configuration
type WeChatConfig struct {
	Enabled    bool
	WebhookURL string
	Timeout    time.Duration
}

// WeChatTextMessage represents text message payload for WeChat Work webhook
type WeChatTextMessage struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
}

// WeChatMarkdownMessage represents markdown message payload for WeChat Work webhook
type WeChatMarkdownMessage struct {
	MsgType  string `json:"msgtype"`
	Markdown struct {
		Content string `json:"content"`
	} `json:"markdown"`
}

type wechatService struct {
	systemConfigService SystemConfigService
	httpClient          *http.Client
}

func NewWeChatService(systemConfigService SystemConfigService) WeChatService {
	return &wechatService{
		systemConfigService: systemConfigService,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// normalizeLanguage normalizes language codes to supported formats
func (s *wechatService) normalizeLanguage(lang string) string {
	if lang == "" {
		return "en-US"
	}

	// Map Chinese language variants to zh-CN
	if strings.HasPrefix(lang, "zh") {
		return "zh-CN"
	}

	return lang
}

// sendWeChatMessage is a common method for sending WeChat notification messages
func (s *wechatService) sendWeChatMessage(admin *database.Admin, messageContent string, lang string) error {
	// Check if WeChat service is enabled
	enabled, err := s.isWeChatEnabled()
	if err != nil {
		utils.Error("Failed to check if WeChat service is enabled", "error", err)
		return err
	}

	if !enabled {
		utils.Info("WeChat service is disabled, skipping notification message", "username", admin.Username)
		return nil
	}

	// Load WeChat configuration
	wechatConfig, err := s.loadWeChatConfig()
	if err != nil {
		utils.Error("Failed to load WeChat configuration", "error", err)
		return err
	}

	// Send WeChat message
	return s.sendMessage(wechatConfig, messageContent)
}

func (s *wechatService) sendMessage(config *WeChatConfig, content string) error {
	// Create message payload
	payload := WeChatTextMessage{
		MsgType: "text",
		Text: struct {
			Content string `json:"content"`
		}{
			Content: content,
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create request with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", config.WebhookURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		utils.Error("Failed to send WeChat webhook request", "url", config.WebhookURL, "error", err)
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		utils.Error("WeChat webhook returned non-OK status", "status_code", resp.StatusCode, "url", config.WebhookURL)
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (s *wechatService) loadWeChatConfig() (*WeChatConfig, error) {
	config := &WeChatConfig{}

	// Define all WeChat config keys to fetch in one query
	configKeys := []string{
		"wechat_webhook_enabled",
		"wechat_webhook_url",
		"wechat_webhook_timeout",
	}

	// Batch fetch all WeChat configurations
	configValues, err := s.systemConfigService.GetValuesByKeys(configKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to load WeChat configurations: %v", err)
	}

	// Parse enabled status
	config.Enabled = configValues["wechat_webhook_enabled"] == "true"

	if !config.Enabled {
		return config, nil
	}

	// Parse webhook URL
	config.WebhookURL = configValues["wechat_webhook_url"]

	// Parse timeout
	timeoutStr := configValues["wechat_webhook_timeout"]
	if timeoutStr == "" {
		timeoutStr = "30s"
	}
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return nil, fmt.Errorf("invalid wechat_webhook_timeout: %v", err)
	}
	config.Timeout = timeout

	return config, nil
}

func (s *wechatService) isWeChatEnabled() (bool, error) {
	config, err := s.loadWeChatConfig()
	if err != nil {
		return false, err
	}

	// Check if all required fields are configured
	if !config.Enabled || config.WebhookURL == "" {
		return false, nil
	}

	return true, nil
}

func (s *wechatService) SendTaskConversationCompletedMessage(admin *database.Admin, task *database.Task, conversation *database.TaskConversation, status database.ConversationStatus, completionTime time.Time, errorMsg string, lang string) error {
	go func() {
		// Prepare status display and emoji
		var statusDisplay, statusEmoji string
		switch status {
		case database.ConversationStatusSuccess:
			statusDisplay = "Success"
			statusEmoji = "✅"
			if lang == "zh-CN" {
				statusDisplay = "成功"
			}
		case database.ConversationStatusFailed:
			statusDisplay = "Failed"
			statusEmoji = "❌"
			if lang == "zh-CN" {
				statusDisplay = "失败"
			}
		case database.ConversationStatusCancelled:
			statusDisplay = "Cancelled"
			statusEmoji = "⚠️"
			if lang == "zh-CN" {
				statusDisplay = "已取消"
			}
		default:
			statusDisplay = "Unknown"
			statusEmoji = "❓"
			if lang == "zh-CN" {
				statusDisplay = "未知"
			}
		}

		// Get admin name
		adminName := admin.Name
		if adminName == "" {
			adminName = admin.Username
		}

		// Truncate conversation content if too long
		conversationContent := conversation.Content
		if len(conversationContent) > 100 {
			conversationContent = conversationContent[:100] + "..."
		}

		// Create message content based on language
		var messageContent string
		if lang == "zh-CN" {
			messageContent = fmt.Sprintf("🤖 任务对话执行完成通知\n\n"+
				"📋 任务: %s\n"+
				"👤 执行者: %s\n"+
				"💬 对话内容: %s\n"+
				"📊 执行状态: %s %s\n"+
				"⏰ 完成时间: %s",
				task.Title,
				adminName,
				conversationContent,
				statusEmoji, statusDisplay,
				completionTime.Format("2006-01-02 15:04:05 MST"))

			if errorMsg != "" && status == database.ConversationStatusFailed {
				messageContent += fmt.Sprintf("\n❗ 错误信息: %s", errorMsg)
			}
		} else {
			messageContent = fmt.Sprintf("🤖 Task Conversation Execution Completed\n\n"+
				"📋 Task: %s\n"+
				"👤 Executor: %s\n"+
				"💬 Conversation: %s\n"+
				"📊 Status: %s %s\n"+
				"⏰ Completed: %s",
				task.Title,
				adminName,
				conversationContent,
				statusEmoji, statusDisplay,
				completionTime.Format("2006-01-02 15:04:05 MST"))

			if errorMsg != "" && status == database.ConversationStatusFailed {
				messageContent += fmt.Sprintf("\n❗ Error: %s", errorMsg)
			}
		}

		if err := s.sendWeChatMessage(admin, messageContent, lang); err != nil {
			utils.Error("Failed to send task conversation completion WeChat notification",
				"username", admin.Username,
				"task_id", task.ID,
				"conversation_id", conversation.ID,
				"status", status,
				"error", err)
		}
	}()
	return nil
}