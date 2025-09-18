package notifiers

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"xsha-backend/database"
)

// FeishuProvider implements NotificationProvider for Feishu webhook
type FeishuProvider struct {
	webhookURL string
	secret     string
	timeout    time.Duration
	httpClient *http.Client
}

// FeishuTextMessage represents text message payload for Feishu webhook
type FeishuTextMessage struct {
	MsgType string `json:"msg_type"`
	Content struct {
		Text string `json:"text"`
	} `json:"content"`
	Timestamp string `json:"timestamp,omitempty"`
	Sign      string `json:"sign,omitempty"`
}

// NewFeishuProvider creates a new Feishu provider
func NewFeishuProvider(config map[string]interface{}) (*FeishuProvider, error) {
	webhookURL, ok := config["webhook_url"].(string)
	if !ok || webhookURL == "" {
		return nil, &ProviderError{
			Type:    "feishu",
			Message: "webhook_url is required",
		}
	}

	secret, _ := config["secret"].(string) // Optional

	timeout := 5 * time.Second

	return &FeishuProvider{
		webhookURL: webhookURL,
		secret:     secret,
		timeout:    timeout,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// GetName returns the provider name
func (p *FeishuProvider) GetName() string {
	return "Feishu"
}

// ValidateConfig validates the Feishu configuration
func (p *FeishuProvider) ValidateConfig(config map[string]interface{}) error {
	webhookURL, ok := config["webhook_url"].(string)
	if !ok || webhookURL == "" {
		return &ProviderError{
			Type:    "feishu",
			Message: "webhook_url is required",
		}
	}

	return nil
}

// Send sends a notification message via Feishu webhook
func (p *FeishuProvider) Send(title, content string, status database.ConversationStatus, lang string) error {
	// Format message content using localized helper
	message := FormatNotificationMessage(title, content, status, lang)

	// Create message payload
	payload := FeishuTextMessage{
		MsgType: "text",
		Content: struct {
			Text string `json:"text"`
		}{
			Text: message,
		},
	}

	// Add signature if secret is provided
	if p.secret != "" {
		timestamp := time.Now().Unix()
		timestampStr := strconv.FormatInt(timestamp, 10)
		sign := p.generateSignature(timestampStr)

		payload.Timestamp = timestampStr
		payload.Sign = sign
	}

	return p.sendMessage(payload)
}

// Test sends a test notification
func (p *FeishuProvider) Test(lang string) error {
	payload := FeishuTextMessage{
		MsgType: "text",
		Content: struct {
			Text string `json:"text"`
		}{
			Text: FormatTestMessage(lang),
		},
	}

	// Add signature if secret is provided
	if p.secret != "" {
		timestamp := time.Now().Unix()
		timestampStr := strconv.FormatInt(timestamp, 10)
		sign := p.generateSignature(timestampStr)

		payload.Timestamp = timestampStr
		payload.Sign = sign
	}

	return p.sendMessage(payload)
}

// sendMessage sends the actual HTTP request to Feishu webhook
func (p *FeishuProvider) sendMessage(payload FeishuTextMessage) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return &ProviderError{
			Type:    "feishu",
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
			Type:    "feishu",
			Message: "failed to create request",
			Err:     err,
		}
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return &ProviderError{
			Type:    "feishu",
			Message: "failed to send request",
			Err:     err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &ProviderError{
			Type:    "feishu",
			Message: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}
	}

	return nil
}

// generateSignature generates HMAC-SHA256 signature for Feishu
func (p *FeishuProvider) generateSignature(timestamp string) string {
	// timestamp + key 做sha256, 再进行base64 encode
	stringToSign := timestamp + "\n" + p.secret

	var data []byte
	h := hmac.New(sha256.New, []byte(stringToSign))
	h.Write(data)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// formatMessage formats the notification message for Feishu
