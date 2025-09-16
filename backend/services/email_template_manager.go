package services

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
	"xsha-backend/utils"
)

//go:embed templates/email/**/*
var emailTemplatesFS embed.FS

type EmailTemplateData struct {
	Subject string
	Body    string
}

type EmailTemplateManager struct {
	mu        sync.RWMutex
	templates map[string]*EmailTemplateData // key: "templateName/language"
	bodyTmpls map[string]*template.Template // key: "templateName/language"
}

var (
	emailTemplateManager *EmailTemplateManager
	emailTemplateOnce    sync.Once
)

// GetEmailTemplateManager returns the singleton email template manager
func GetEmailTemplateManager() *EmailTemplateManager {
	emailTemplateOnce.Do(func() {
		emailTemplateManager = &EmailTemplateManager{
			templates: make(map[string]*EmailTemplateData),
			bodyTmpls: make(map[string]*template.Template),
		}
		emailTemplateManager.loadTemplates()
	})
	return emailTemplateManager
}

// loadTemplates loads all email templates from embedded filesystem
func (m *EmailTemplateManager) loadTemplates() {
	err := fs.WalkDir(emailTemplatesFS, "templates/email", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Parse path to extract template name and language
		// Expected format: templates/email/{templateName}/{language}.{extension}
		relativePath := strings.TrimPrefix(path, "templates/email/")
		parts := strings.Split(relativePath, "/")
		if len(parts) != 2 {
			return nil // Skip files not in expected structure
		}

		templateName := parts[0]
		fileName := parts[1]

		// Extract language and extension
		ext := filepath.Ext(fileName)
		language := strings.TrimSuffix(fileName, ext)

		key := fmt.Sprintf("%s/%s", templateName, language)

		switch ext {
		case ".subject":
			if err := m.loadSubjectTemplate(path, key); err != nil {
				utils.Error("Failed to load subject template", "path", path, "error", err)
			}
		case ".html":
			if err := m.loadBodyTemplate(path, key); err != nil {
				utils.Error("Failed to load body template", "path", path, "error", err)
			}
		}

		return nil
	})

	if err != nil {
		utils.Error("Failed to walk email templates directory", "error", err)
	}
}

// loadSubjectTemplate loads a subject template
func (m *EmailTemplateManager) loadSubjectTemplate(path, key string) error {
	content, err := fs.ReadFile(emailTemplatesFS, path)
	if err != nil {
		return fmt.Errorf("failed to read subject template %s: %v", path, err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.templates[key] == nil {
		m.templates[key] = &EmailTemplateData{}
	}
	m.templates[key].Subject = strings.TrimSpace(string(content))

	utils.Info("Loaded subject template", "key", key, "path", path)
	return nil
}

// loadBodyTemplate loads a body template
func (m *EmailTemplateManager) loadBodyTemplate(path, key string) error {
	content, err := fs.ReadFile(emailTemplatesFS, path)
	if err != nil {
		return fmt.Errorf("failed to read body template %s: %v", path, err)
	}

	tmpl, err := template.New(key).Parse(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse body template %s: %v", path, err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.templates[key] == nil {
		m.templates[key] = &EmailTemplateData{}
	}
	m.templates[key].Body = string(content)
	m.bodyTmpls[key] = tmpl

	utils.Info("Loaded body template", "key", key, "path", path)
	return nil
}

// GetTemplate returns the email template for the given template name and language
func (m *EmailTemplateManager) GetTemplate(templateName, language string) (*EmailTemplateData, error) {
	key := fmt.Sprintf("%s/%s", templateName, language)

	m.mu.RLock()
	defer m.mu.RUnlock()

	template, exists := m.templates[key]
	if !exists {
		// Try fallback to en-US if requested language doesn't exist
		if language != "en-US" {
			fallbackKey := fmt.Sprintf("%s/en-US", templateName)
			if fallbackTemplate, fallbackExists := m.templates[fallbackKey]; fallbackExists {
				utils.Info("Using fallback language", "requested", language, "fallback", "en-US", "template", templateName)
				return fallbackTemplate, nil
			}
		}
		return nil, fmt.Errorf("template not found: %s", key)
	}

	return template, nil
}

// RenderTemplate renders the email template with the given data
func (m *EmailTemplateManager) RenderTemplate(templateName, language string, data interface{}) (subject string, body string, err error) {
	template, err := m.GetTemplate(templateName, language)
	if err != nil {
		return "", "", err
	}

	subject = template.Subject

	// Render body template
	key := fmt.Sprintf("%s/%s", templateName, language)

	m.mu.RLock()
	bodyTmpl, exists := m.bodyTmpls[key]
	m.mu.RUnlock()

	if !exists {
		// Try fallback to en-US
		if language != "en-US" {
			fallbackKey := fmt.Sprintf("%s/en-US", templateName)
			m.mu.RLock()
			fallbackBodyTmpl, fallbackExists := m.bodyTmpls[fallbackKey]
			m.mu.RUnlock()

			if fallbackExists {
				bodyTmpl = fallbackBodyTmpl
			} else {
				return "", "", fmt.Errorf("body template not found: %s", key)
			}
		} else {
			return "", "", fmt.Errorf("body template not found: %s", key)
		}
	}

	var bodyBuilder strings.Builder
	err = bodyTmpl.Execute(&bodyBuilder, data)
	if err != nil {
		return "", "", fmt.Errorf("failed to execute body template: %v", err)
	}

	body = bodyBuilder.String()
	return subject, body, nil
}

// ListAvailableTemplates returns a list of all available templates
func (m *EmailTemplateManager) ListAvailableTemplates() map[string][]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string][]string)
	for key := range m.templates {
		parts := strings.Split(key, "/")
		if len(parts) == 2 {
			templateName, language := parts[0], parts[1]
			if _, exists := result[templateName]; !exists {
				result[templateName] = []string{}
			}
			// Add language if not already in list
			found := false
			for _, lang := range result[templateName] {
				if lang == language {
					found = true
					break
				}
			}
			if !found {
				result[templateName] = append(result[templateName], language)
			}
		}
	}

	return result
}