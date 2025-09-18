package notifiers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"xsha-backend/database"
)

// WeChatWorkProvider implements NotificationProvider for WeChat Work webhook
type WeChatWorkProvider struct {
	webhookURL string
	timeout    time.Duration
	httpClient *http.Client
}

// WeChatTextMessage represents text message payload for WeChat Work webhook
type WeChatTextMessage struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
}

// NewWeChatWorkProvider creates a new WeChat Work provider
func NewWeChatWorkProvider(config map[string]interface{}) (*WeChatWorkProvider, error) {
	webhookURL, ok := config["webhook_url"].(string)
	if !ok || webhookURL == "" {
		return nil, &ProviderError{
			Type:    "wechat_work",
			Message: "webhook_url is required",
		}
	}

	timeout := 5 * time.Second

	return &WeChatWorkProvider{
		webhookURL: webhookURL,
		timeout:    timeout,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// GetName returns the provider name
func (p *WeChatWorkProvider) GetName() string {
	return "WeChat Work"
}

// ValidateConfig validates the WeChat Work configuration
func (p *WeChatWorkProvider) ValidateConfig(config map[string]interface{}) error {
	webhookURL, ok := config["webhook_url"].(string)
	if !ok || webhookURL == "" {
		return &ProviderError{
			Type:    "wechat_work",
			Message: "webhook_url is required",
		}
	}

	return nil
}

// Send sends a notification message via WeChat Work webhook
func (p *WeChatWorkProvider) Send(title, content, projectName string, status database.ConversationStatus, lang string) error {
	// Format message content using localized helper
	message := FormatNotificationMessage(title, content, projectName, status, lang)

	// Create message payload
	payload := WeChatTextMessage{
		MsgType: "text",
		Text: struct {
			Content string `json:"content"`
		}{
			Content: message,
		},
	}

	return p.sendMessage(payload)
}

// Test sends a test notification
func (p *WeChatWorkProvider) Test(lang string) error {
	payload := WeChatTextMessage{
		MsgType: "text",
		Text: struct {
			Content string `json:"content"`
		}{
			Content: FormatTestMessage(lang),
		},
	}

	return p.sendMessage(payload)
}

// sendMessage sends the actual HTTP request to WeChat webhook
func (p *WeChatWorkProvider) sendMessage(payload WeChatTextMessage) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return &ProviderError{
			Type:    "wechat_work",
			Message: "failed to marshal payload",
			Err:     err,
		}
	}

	// Create request with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", p.webhookURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return &ProviderError{
			Type:    "wechat_work",
			Message: "failed to create request",
			Err:     err,
		}
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return &ProviderError{
			Type:    "wechat_work",
			Message: "failed to send request",
			Err:     err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &ProviderError{
			Type:    "wechat_work",
			Message: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}
	}

	return nil
}
