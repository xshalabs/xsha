package notifiers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"xsha-backend/database"
	"xsha-backend/i18n"
)

// SlackProvider implements NotificationProvider for Slack webhook
type SlackProvider struct {
	webhookURL string
	timeout    time.Duration
	httpClient *http.Client
}

// SlackMessage represents message payload for Slack webhook
type SlackMessage struct {
	Text        string            `json:"text"`
	Username    string            `json:"username,omitempty"`
	IconEmoji   string            `json:"icon_emoji,omitempty"`
	Channel     string            `json:"channel,omitempty"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
}

// SlackAttachment represents an attachment in Slack message
type SlackAttachment struct {
	Color     string       `json:"color,omitempty"`
	Title     string       `json:"title,omitempty"`
	Text      string       `json:"text,omitempty"`
	Fields    []SlackField `json:"fields,omitempty"`
	Timestamp int64        `json:"ts,omitempty"`
}

// SlackField represents a field in Slack attachment
type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// NewSlackProvider creates a new Slack provider
func NewSlackProvider(config map[string]interface{}) (*SlackProvider, error) {
	webhookURL, ok := config["webhook_url"].(string)
	if !ok || webhookURL == "" {
		return nil, &ProviderError{
			Type:    "slack",
			Message: "webhook_url is required",
		}
	}

	timeout := 5 * time.Second

	return &SlackProvider{
		webhookURL: webhookURL,
		timeout:    timeout,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// GetName returns the provider name
func (p *SlackProvider) GetName() string {
	return "Slack"
}

// ValidateConfig validates the Slack configuration
func (p *SlackProvider) ValidateConfig(config map[string]interface{}) error {
	webhookURL, ok := config["webhook_url"].(string)
	if !ok || webhookURL == "" {
		return &ProviderError{
			Type:    "slack",
			Message: "webhook_url is required",
		}
	}

	return nil
}

// Send sends a notification message via Slack webhook
func (p *SlackProvider) Send(ctx *NotificationContext) error {
	// Create message payload with rich formatting
	payload := p.createMessage(ctx.Title, ctx.Content, ctx.ProjectName, ctx.Status, ctx.Lang)

	return p.sendMessage(payload)
}

// Test sends a test notification
func (p *SlackProvider) Test(lang string) error {
	payload := SlackMessage{
		Text:      FormatTestMessage(lang),
		Username:  "xsha",
		IconEmoji: ":robot_face:",
		Attachments: []SlackAttachment{
			{
				Color: "good",
				Title: "Test Notification",
				Text:  "This is a test message to verify the Slack notification configuration.",
				Fields: []SlackField{
					{
						Title: "Status",
						Value: "✅ Connected",
						Short: true,
					},
					{
						Title: "Time",
						Value: time.Now().Format("2006-01-02 15:04:05 MST"),
						Short: true,
					},
				},
				Timestamp: time.Now().Unix(),
			},
		},
	}

	return p.sendMessage(payload)
}

// sendMessage sends the actual HTTP request to Slack webhook
func (p *SlackProvider) sendMessage(payload SlackMessage) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return &ProviderError{
			Type:    "slack",
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
			Type:    "slack",
			Message: "failed to create request",
			Err:     err,
		}
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return &ProviderError{
			Type:    "slack",
			Message: "failed to send request",
			Err:     err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &ProviderError{
			Type:    "slack",
			Message: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}
	}

	return nil
}

// createMessage creates a Slack message with rich formatting
func (p *SlackProvider) createMessage(title, content, projectName string, status database.ConversationStatus, lang string) SlackMessage {
	statusEmoji := FormatStatusEmoji(status)
	statusText := FormatStatusText(status, lang)

	// Determine color based on status
	color := "warning"
	switch status {
	case database.ConversationStatusSuccess:
		color = "good"
	case database.ConversationStatusFailed:
		color = "danger"
	}

	message := SlackMessage{
		Text:      fmt.Sprintf("🤖 %s", i18n.T(lang, "notification.task_execution_title")),
		Username:  "xsha",
		IconEmoji: ":robot_face:",
		Attachments: []SlackAttachment{
			{
				Color:     color,
				Title:     fmt.Sprintf("%s Task: %s", statusEmoji, title),
				Text:      TruncateContent(content, 200),
				Fields:    buildSlackFields(projectName, statusEmoji, statusText, lang),
				Timestamp: time.Now().Unix(),
			},
		},
	}

	return message
}

// buildSlackFields builds the fields array for Slack message, including project name if available
func buildSlackFields(projectName, statusEmoji, statusText, lang string) []SlackField {
	fields := []SlackField{}

	// Add project field if project name is provided
	if projectName != "" {
		fields = append(fields, SlackField{
			Title: i18n.T(lang, "notification.project_label"),
			Value: projectName,
			Short: true,
		})
	}

	// Add status field
	fields = append(fields, SlackField{
		Title: i18n.T(lang, "notification.status_label"),
		Value: fmt.Sprintf("%s %s", statusEmoji, statusText),
		Short: true,
	})

	// Add time field
	fields = append(fields, SlackField{
		Title: i18n.T(lang, "notification.time_label"),
		Value: time.Now().Format("2006-01-02 15:04:05 MST"),
		Short: true,
	})

	return fields
}
