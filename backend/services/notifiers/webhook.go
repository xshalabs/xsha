package notifiers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"xsha-backend/i18n"
)

// WebhookProvider implements NotificationProvider for generic webhook
type WebhookProvider struct {
	url          string
	method       string
	headers      map[string]string
	bodyTemplate string
	timeout      time.Duration
	httpClient   *http.Client
}

// NewWebhookProvider creates a new generic webhook provider
func NewWebhookProvider(config map[string]interface{}) (*WebhookProvider, error) {
	url, ok := config["url"].(string)
	if !ok || url == "" {
		return nil, &ProviderError{
			Type:    "webhook",
			Message: "url is required",
		}
	}

	method, ok := config["method"].(string)
	if !ok || method == "" {
		method = "POST"
	}
	method = strings.ToUpper(method)

	bodyTemplate, _ := config["body_template"].(string)
	if bodyTemplate == "" {
		bodyTemplate = `{
			"title": "{{.Title}}",
			"content": "{{.Content}}",
			"project_name": "{{.ProjectName}}",
			"project_id": {{.ProjectID}},
			"task_id": {{.TaskID}},
			"conv_id": {{.ConvID}},
			"status": "{{.Status}}",
			"timestamp": "{{.Timestamp}}"
		}`
	}

	// Parse headers
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json" // Default
	if headersConfig, ok := config["headers"].(map[string]interface{}); ok {
		for k, v := range headersConfig {
			if str, ok := v.(string); ok {
				headers[k] = str
			}
		}
	}

	timeout := 5 * time.Second

	return &WebhookProvider{
		url:          url,
		method:       method,
		headers:      headers,
		bodyTemplate: bodyTemplate,
		timeout:      timeout,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// GetName returns the provider name
func (p *WebhookProvider) GetName() string {
	return "Generic Webhook"
}

// ValidateConfig validates the webhook configuration
func (p *WebhookProvider) ValidateConfig(config map[string]interface{}) error {
	url, ok := config["url"].(string)
	if !ok || url == "" {
		return &ProviderError{
			Type:    "webhook",
			Message: "url is required",
		}
	}

	// Validate method if provided
	if method, ok := config["method"].(string); ok && method != "" {
		method = strings.ToUpper(method)
		validMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
		valid := false
		for _, validMethod := range validMethods {
			if method == validMethod {
				valid = true
				break
			}
		}
		if !valid {
			return &ProviderError{
				Type:    "webhook",
				Message: "invalid HTTP method",
			}
		}
	}

	return nil
}

// Send sends a notification message via generic webhook
func (p *WebhookProvider) Send(ctx *NotificationContext) error {
	// Prepare template data
	data := struct {
		Title       string
		Content     string
		ProjectName string
		Status      string
		Timestamp   string
		ProjectID   uint
		TaskID      uint
		ConvID      uint
	}{
		Title:       ctx.Title,
		Content:     ctx.Content,
		ProjectName: ctx.ProjectName,
		Status:      string(ctx.Status),
		Timestamp:   ctx.Timestamp.Format(time.RFC3339),
		ProjectID:   ctx.ProjectID,
		TaskID:      ctx.TaskID,
		ConvID:      ctx.ConvID,
	}

	// Create request body
	bodyMap, err := p.renderTemplate(data)
	if err != nil {
		return &ProviderError{
			Type:    "webhook",
			Message: "failed to render body template",
			Err:     err,
		}
	}

	// Serialize to JSON string
	bodyBytes, err := json.Marshal(bodyMap)
	if err != nil {
		return &ProviderError{
			Type:    "webhook",
			Message: "failed to serialize body to JSON",
			Err:     err,
		}
	}

	return p.sendRequest(string(bodyBytes))
}

// Test sends a test notification
func (p *WebhookProvider) Test(lang string) error {
	// Prepare test data
	data := struct {
		Title       string
		Content     string
		ProjectName string
		Status      string
		Timestamp   string
		ProjectID   uint
		TaskID      uint
		ConvID      uint
	}{
		Title:       i18n.T(lang, "notification.test_message"),
		Content:     i18n.T(lang, "notification.test_message"),
		ProjectName: "Test Project",
		Status:      "test",
		Timestamp:   time.Now().Format(time.RFC3339),
		ProjectID:   0,
		TaskID:      0,
		ConvID:      0,
	}

	// Create request body
	bodyMap, err := p.renderTemplate(data)
	if err != nil {
		return &ProviderError{
			Type:    "webhook",
			Message: "failed to render body template",
			Err:     err,
		}
	}

	// Serialize to JSON string
	bodyBytes, err := json.Marshal(bodyMap)
	if err != nil {
		return &ProviderError{
			Type:    "webhook",
			Message: "failed to serialize body to JSON",
			Err:     err,
		}
	}

	return p.sendRequest(string(bodyBytes))
}

// sendRequest sends the actual HTTP request to the webhook
func (p *WebhookProvider) sendRequest(body string) error {
	// Create request with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	var req *http.Request
	var err error

	if p.method == "GET" {
		req, err = http.NewRequestWithContext(ctx, p.method, p.url, nil)
	} else {
		req, err = http.NewRequestWithContext(ctx, p.method, p.url, bytes.NewBufferString(body))
	}

	if err != nil {
		return &ProviderError{
			Type:    "webhook",
			Message: "failed to create request",
			Err:     err,
		}
	}

	// Set headers
	for key, value := range p.headers {
		req.Header.Set(key, value)
	}

	// Send request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return &ProviderError{
			Type:    "webhook",
			Message: "failed to send request",
			Err:     err,
		}
	}
	defer resp.Body.Close()

	// Check response status (accept 2xx status codes)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &ProviderError{
			Type:    "webhook",
			Message: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}
	}

	return nil
}

// renderTemplate renders the body template with the provided data
func (p *WebhookProvider) renderTemplate(data interface{}) (map[string]interface{}, error) {
	// Convert data to map for easier processing
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var dataMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &dataMap); err != nil {
		return nil, err
	}

	// Build default message structure
	defaultMessage := map[string]interface{}{
		"title":        dataMap["Title"],
		"content":      dataMap["Content"],
		"project_name": dataMap["ProjectName"],
		"project_id":   dataMap["ProjectID"],
		"task_id":      dataMap["TaskID"],
		"conv_id":      dataMap["ConvID"],
		"status":       dataMap["Status"],
		"timestamp":    dataMap["Timestamp"],
	}

	// If no custom body template, return default structure
	if p.bodyTemplate == "" {
		return defaultMessage, nil
	}

	// Try to parse custom body template as JSON
	var customTemplate map[string]interface{}

	// First, try to render template placeholders
	renderedTemplate := p.bodyTemplate
	for key, value := range dataMap {
		placeholder := fmt.Sprintf("{{.%s}}", key)
		if str, ok := value.(string); ok {
			renderedTemplate = strings.ReplaceAll(renderedTemplate, placeholder, str)
		}
	}

	// Parse rendered template as JSON
	if err := json.Unmarshal([]byte(renderedTemplate), &customTemplate); err != nil {
		// If custom template is invalid JSON, log warning and use default
		fmt.Printf("Warning: Invalid JSON in body_template, using default structure: %v\n", err)
		return defaultMessage, nil
	}

	// Merge custom template with default message
	// Custom fields override default fields
	result := make(map[string]interface{})

	// Start with default message
	for key, value := range defaultMessage {
		result[key] = value
	}

	// Override with custom template values
	for key, value := range customTemplate {
		result[key] = value
	}

	return result, nil
}
