package notifiers

import (
	"fmt"
	"time"
	"xsha-backend/database"
	"xsha-backend/i18n"
)

// NotificationContext contains all notification parameters
type NotificationContext struct {
	// Basic information
	Title       string
	Content     string
	ProjectName string
	Status      database.ConversationStatus
	Lang        string

	// ID information
	ProjectID uint
	TaskID    uint
	ConvID    uint

	// Time information
	Timestamp time.Time
}

// NotificationProvider defines the interface that all notification providers must implement
type NotificationProvider interface {
	// Send sends a notification message
	Send(ctx *NotificationContext) error

	// Test sends a test notification to verify the configuration
	Test(lang string) error

	// ValidateConfig validates the provider configuration
	ValidateConfig(config map[string]interface{}) error

	// GetName returns the provider name
	GetName() string
}

// BaseNotificationData contains common data for all notifications
type BaseNotificationData struct {
	Title          string
	Content        string
	ProjectName    string
	Status         database.ConversationStatus
	TaskTitle      string
	AdminName      string
	CompletionTime time.Time
	ErrorMessage   string
	Language       string
}

// NotificationConfig represents the configuration for a notifier
type NotificationConfig struct {
	Type   database.NotifierType
	Config map[string]interface{}
}

// NewProvider creates a new notification provider based on the type and config
func NewProvider(notifierType database.NotifierType, config map[string]interface{}) (NotificationProvider, error) {
	switch notifierType {
	case database.NotifierTypeWeChatWork:
		return NewWeChatWorkProvider(config)
	case database.NotifierTypeDingTalk:
		return NewDingTalkProvider(config)
	case database.NotifierTypeFeishu:
		return NewFeishuProvider(config)
	case database.NotifierTypeSlack:
		return NewSlackProvider(config)
	case database.NotifierTypeDiscord:
		return NewDiscordProvider(config)
	case database.NotifierTypeWebhook:
		return NewWebhookProvider(config)
	default:
		return nil, &ProviderError{
			Type:    string(notifierType),
			Message: "unsupported notification provider type",
		}
	}
}

// ProviderError represents an error from a notification provider
type ProviderError struct {
	Type    string
	Message string
	Err     error
}

func (e *ProviderError) Error() string {
	if e.Err != nil {
		return e.Type + ": " + e.Message + ": " + e.Err.Error()
	}
	return e.Type + ": " + e.Message
}

func (e *ProviderError) Unwrap() error {
	return e.Err
}

// Helper functions for formatting notification content

func FormatStatusEmoji(status database.ConversationStatus) string {
	switch status {
	case database.ConversationStatusSuccess:
		return "✅"
	case database.ConversationStatusFailed:
		return "❌"
	case database.ConversationStatusCancelled:
		return "⚠️"
	default:
		return "❓"
	}
}

func FormatStatusText(status database.ConversationStatus, lang string) string {
	if lang == "zh-CN" {
		switch status {
		case database.ConversationStatusSuccess:
			return "成功"
		case database.ConversationStatusFailed:
			return "失败"
		case database.ConversationStatusCancelled:
			return "已取消"
		default:
			return "未知"
		}
	}

	// English
	switch status {
	case database.ConversationStatusSuccess:
		return "Success"
	case database.ConversationStatusFailed:
		return "Failed"
	case database.ConversationStatusCancelled:
		return "Cancelled"
	default:
		return "Unknown"
	}
}

func TruncateContent(content string, maxLength int) string {
	if len(content) <= maxLength {
		return content
	}
	return content[:maxLength] + "..."
}

// FormatNotificationMessage creates a localized notification message
func FormatNotificationMessage(title, content, projectName string, status database.ConversationStatus, lang string) string {
	statusEmoji := FormatStatusEmoji(status)
	statusText := FormatStatusText(status, lang)

	notificationTitle := i18n.T(lang, "notification.task_execution_title")
	projectLabel := i18n.T(lang, "notification.project_label")
	taskLabel := i18n.T(lang, "notification.task_label")
	statusLabel := i18n.T(lang, "notification.status_label")
	contentLabel := i18n.T(lang, "notification.content_label")
	timeLabel := i18n.T(lang, "notification.time_label")

	var projectLine string
	if projectName != "" {
		projectLine = fmt.Sprintf("🗂️ %s: %s\n", projectLabel, projectName)
	}

	return fmt.Sprintf("🤖 %s\n\n"+
		"%s"+
		"📋 %s: %s\n"+
		"📊 %s: %s %s\n"+
		"💬 %s: %s\n"+
		"⏰ %s: %s",
		notificationTitle,
		projectLine,
		taskLabel, title,
		statusLabel, statusEmoji, statusText,
		contentLabel, TruncateContent(content, 100),
		timeLabel, time.Now().Format("2006-01-02 15:04:05 MST"))
}

// FormatTestMessage creates a localized test message
func FormatTestMessage(lang string) string {
	testMessage := i18n.T(lang, "notification.test_message")
	return "🤖 " + testMessage
}
