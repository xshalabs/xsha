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

// DiscordProvider implements NotificationProvider for Discord webhook
type DiscordProvider struct {
	webhookURL string
	timeout    time.Duration
	httpClient *http.Client
}

// DiscordMessage represents message payload for Discord webhook
type DiscordMessage struct {
	Content   string         `json:"content,omitempty"`
	Username  string         `json:"username,omitempty"`
	AvatarURL string         `json:"avatar_url,omitempty"`
	Embeds    []DiscordEmbed `json:"embeds,omitempty"`
}

// DiscordEmbed represents an embed in Discord message
type DiscordEmbed struct {
	Title       string              `json:"title,omitempty"`
	Description string              `json:"description,omitempty"`
	Color       int                 `json:"color,omitempty"`
	Timestamp   string              `json:"timestamp,omitempty"`
	Footer      *DiscordEmbedFooter `json:"footer,omitempty"`
	Fields      []DiscordEmbedField `json:"fields,omitempty"`
}

// DiscordEmbedFooter represents footer in Discord embed
type DiscordEmbedFooter struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url,omitempty"`
}

// DiscordEmbedField represents a field in Discord embed
type DiscordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// NewDiscordProvider creates a new Discord provider
func NewDiscordProvider(config map[string]interface{}) (*DiscordProvider, error) {
	webhookURL, ok := config["webhook_url"].(string)
	if !ok || webhookURL == "" {
		return nil, &ProviderError{
			Type:    "discord",
			Message: "webhook_url is required",
		}
	}

	timeout := 5 * time.Second

	return &DiscordProvider{
		webhookURL: webhookURL,
		timeout:    timeout,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// GetName returns the provider name
func (p *DiscordProvider) GetName() string {
	return "Discord"
}

// ValidateConfig validates the Discord configuration
func (p *DiscordProvider) ValidateConfig(config map[string]interface{}) error {
	webhookURL, ok := config["webhook_url"].(string)
	if !ok || webhookURL == "" {
		return &ProviderError{
			Type:    "discord",
			Message: "webhook_url is required",
		}
	}

	return nil
}

// Send sends a notification message via Discord webhook
func (p *DiscordProvider) Send(title, content, projectName string, status database.ConversationStatus, lang string) error {
	// Create message payload with rich formatting
	payload := p.createMessage(title, content, projectName, status, lang)

	return p.sendMessage(payload)
}

// Test sends a test notification
func (p *DiscordProvider) Test(lang string) error {
	payload := DiscordMessage{
		Username: "xsha",
		Embeds: []DiscordEmbed{
			{
				Title:       fmt.Sprintf("🤖 %s", i18n.T(lang, "notification.test_message")),
				Description: i18n.T(lang, "notification.test_message"),
				Color:       0x00ff00, // Green
				Timestamp:   time.Now().Format(time.RFC3339),
				Footer: &DiscordEmbedFooter{
					Text: "xsha notification system",
				},
				Fields: []DiscordEmbedField{
					{
						Name:   i18n.T(lang, "notification.status_label"),
						Value:  "✅ Connected",
						Inline: true,
					},
					{
						Name:   i18n.T(lang, "notification.time_label"),
						Value:  time.Now().Format("2006-01-02 15:04:05 MST"),
						Inline: true,
					},
				},
			},
		},
	}

	return p.sendMessage(payload)
}

// sendMessage sends the actual HTTP request to Discord webhook
func (p *DiscordProvider) sendMessage(payload DiscordMessage) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return &ProviderError{
			Type:    "discord",
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
			Type:    "discord",
			Message: "failed to create request",
			Err:     err,
		}
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return &ProviderError{
			Type:    "discord",
			Message: "failed to send request",
			Err:     err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return &ProviderError{
			Type:    "discord",
			Message: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}
	}

	return nil
}

// createMessage creates a Discord message with rich formatting
func (p *DiscordProvider) createMessage(title, content, projectName string, status database.ConversationStatus, lang string) DiscordMessage {
	statusEmoji := FormatStatusEmoji(status)
	statusText := FormatStatusText(status, lang)

	// Determine color based on status
	color := 0xffaa00 // Orange for unknown
	switch status {
	case database.ConversationStatusSuccess:
		color = 0x00ff00 // Green
	case database.ConversationStatusFailed:
		color = 0xff0000 // Red
	case database.ConversationStatusCancelled:
		color = 0xffaa00 // Orange
	}

	message := DiscordMessage{
		Username: "xsha",
		Embeds: []DiscordEmbed{
			{
				Title:       fmt.Sprintf("%s %s", statusEmoji, i18n.T(lang, "notification.task_execution_title")),
				Description: fmt.Sprintf("**%s:** %s\n**%s:** %s", i18n.T(lang, "notification.task_label"), title, i18n.T(lang, "notification.content_label"), TruncateContent(content, 200)),
				Color:       color,
				Timestamp:   time.Now().Format(time.RFC3339),
				Footer: &DiscordEmbedFooter{
					Text: "xsha notification system",
				},
				Fields: buildDiscordFields(projectName, statusEmoji, statusText, lang),
			},
		},
	}

	return message
}

// buildDiscordFields builds the fields array for Discord embed, including project name if available
func buildDiscordFields(projectName, statusEmoji, statusText, lang string) []DiscordEmbedField {
	fields := []DiscordEmbedField{}

	// Add project field if project name is provided
	if projectName != "" {
		fields = append(fields, DiscordEmbedField{
			Name:   i18n.T(lang, "notification.project_label"),
			Value:  projectName,
			Inline: true,
		})
	}

	// Add status field
	fields = append(fields, DiscordEmbedField{
		Name:   i18n.T(lang, "notification.status_label"),
		Value:  fmt.Sprintf("%s %s", statusEmoji, statusText),
		Inline: true,
	})

	// Add time field
	fields = append(fields, DiscordEmbedField{
		Name:   i18n.T(lang, "notification.time_label"),
		Value:  time.Now().Format("2006-01-02 15:04:05 MST"),
		Inline: true,
	})

	return fields
}
