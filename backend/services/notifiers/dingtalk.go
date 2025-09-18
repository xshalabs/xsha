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
	"net/url"
	"strconv"
	"time"
	"xsha-backend/database"
)

// DingTalkProvider implements NotificationProvider for DingTalk webhook
type DingTalkProvider struct {
	webhookURL string
	secret     string
	timeout    time.Duration
	httpClient *http.Client
}

// DingTalkTextMessage represents text message payload for DingTalk webhook
type DingTalkTextMessage struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
}

// NewDingTalkProvider creates a new DingTalk provider
func NewDingTalkProvider(config map[string]interface{}) (*DingTalkProvider, error) {
	webhookURL, ok := config["webhook_url"].(string)
	if !ok || webhookURL == "" {
		return nil, &ProviderError{
			Type:    "dingtalk",
			Message: "webhook_url is required",
		}
	}

	secret, _ := config["secret"].(string) // Optional

	timeout := 5 * time.Second

	return &DingTalkProvider{
		webhookURL: webhookURL,
		secret:     secret,
		timeout:    timeout,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// GetName returns the provider name
func (p *DingTalkProvider) GetName() string {
	return "DingTalk"
}

// ValidateConfig validates the DingTalk configuration
func (p *DingTalkProvider) ValidateConfig(config map[string]interface{}) error {
	webhookURL, ok := config["webhook_url"].(string)
	if !ok || webhookURL == "" {
		return &ProviderError{
			Type:    "dingtalk",
			Message: "webhook_url is required",
		}
	}

	return nil
}

// Send sends a notification message via DingTalk webhook
func (p *DingTalkProvider) Send(title, content, projectName string, status database.ConversationStatus, lang string) error {
	// Format message content using localized helper
	message := FormatNotificationMessage(title, content, projectName, status, lang)

	// Create message payload
	payload := DingTalkTextMessage{
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
func (p *DingTalkProvider) Test(lang string) error {
	payload := DingTalkTextMessage{
		MsgType: "text",
		Text: struct {
			Content string `json:"content"`
		}{
			Content: FormatTestMessage(lang),
		},
	}

	return p.sendMessage(payload)
}

// sendMessage sends the actual HTTP request to DingTalk webhook
func (p *DingTalkProvider) sendMessage(payload DingTalkTextMessage) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return &ProviderError{
			Type:    "dingtalk",
			Message: "failed to marshal payload",
			Err:     err,
		}
	}

	// Build URL with signature if secret is provided
	requestURL := p.webhookURL
	if p.secret != "" {
		timestamp := time.Now().UnixNano() / 1e6
		sign := p.generateSignature(timestamp)

		parsedURL, err := url.Parse(p.webhookURL)
		if err != nil {
			return &ProviderError{
				Type:    "dingtalk",
				Message: "invalid webhook URL",
				Err:     err,
			}
		}

		query := parsedURL.Query()
		query.Set("timestamp", strconv.FormatInt(timestamp, 10))
		query.Set("sign", sign)
		parsedURL.RawQuery = query.Encode()
		requestURL = parsedURL.String()
	}

	// Create request with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", requestURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return &ProviderError{
			Type:    "dingtalk",
			Message: "failed to create request",
			Err:     err,
		}
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return &ProviderError{
			Type:    "dingtalk",
			Message: "failed to send request",
			Err:     err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &ProviderError{
			Type:    "dingtalk",
			Message: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}
	}

	return nil
}

// generateSignature generates HMAC-SHA256 signature for DingTalk
func (p *DingTalkProvider) generateSignature(timestamp int64) string {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, p.secret)
	h := hmac.New(sha256.New, []byte(p.secret))
	h.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// formatMessage formats the notification message for DingTalk
