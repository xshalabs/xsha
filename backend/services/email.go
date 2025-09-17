package services

import (
	"crypto/tls"
	"fmt"
	"strconv"
	"strings"
	"time"
	"xsha-backend/database"
	"xsha-backend/utils"

	"gopkg.in/gomail.v2"
)

type SMTPConfig struct {
	Enabled    bool
	Host       string
	Port       int
	Username   string
	Password   string
	From       string
	FromName   string
	UseTLS     bool
	SkipVerify bool
}

type emailService struct {
	systemConfigService SystemConfigService
}

func NewEmailService(systemConfigService SystemConfigService) EmailService {
	return &emailService{
		systemConfigService: systemConfigService,
	}
}

// normalizeLanguage normalizes language codes to supported formats
func (s *emailService) normalizeLanguage(lang string) string {
	if lang == "" {
		return "en-US"
	}

	// Map Chinese language variants to zh-CN
	if strings.HasPrefix(lang, "zh") {
		return "zh-CN"
	}

	return lang
}

// sendNotificationEmail is a common method for sending notification emails
func (s *emailService) sendNotificationEmail(admin *database.Admin, templateName, lang string, templateData interface{}) error {
	// Check if email service is enabled
	enabled, err := s.isEmailEnabled()
	if err != nil {
		utils.Error("Failed to check if email service is enabled", "error", err)
		return err
	}

	if !enabled {
		utils.Info("Email service is disabled, skipping notification email", "username", admin.Username, "template", templateName)
		return nil
	}

	// Check if admin has email
	if admin.Email == "" {
		utils.Info("Admin has no email address, skipping notification email", "username", admin.Username, "template", templateName)
		return nil
	}

	// Load SMTP configuration
	smtpConfig, err := s.loadSMTPConfig()
	if err != nil {
		utils.Error("Failed to load SMTP configuration", "error", err)
		return err
	}

	// Normalize language
	normalizedLang := s.normalizeLanguage(lang)

	// Generate email content using template manager
	templateManager := GetEmailTemplateManager()
	subject, body, err := templateManager.RenderTemplate(templateName, normalizedLang, templateData)
	if err != nil {
		utils.Error("Failed to generate email content", "template", templateName, "error", err)
		return err
	}

	// Send email
	return s.sendEmail(smtpConfig, admin.Email, subject, body)
}

func (s *emailService) sendEmail(config *SMTPConfig, to, subject, body string) error {
	m := gomail.NewMessage()

	// Set headers
	m.SetHeader("From", m.FormatAddress(config.From, config.FromName))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	// Create dialer
	d := gomail.NewDialer(config.Host, config.Port, config.Username, config.Password)

	if config.UseTLS {
		d.TLSConfig = &tls.Config{
			InsecureSkipVerify: config.SkipVerify,
			ServerName:         config.Host,
		}
	}

	// Send email
	if err := d.DialAndSend(m); err != nil {
		utils.Error("Failed to send email", "to", to, "subject", subject, "error", err)
		return fmt.Errorf("failed to send email: %v", err)
	}

	utils.Info("Email sent successfully", "to", to, "subject", subject)
	return nil
}

func (s *emailService) loadSMTPConfig() (*SMTPConfig, error) {
	config := &SMTPConfig{}

	// Define all SMTP config keys to fetch in one query
	configKeys := []string{
		"smtp_enabled",
		"smtp_host",
		"smtp_port",
		"smtp_username",
		"smtp_password",
		"smtp_from",
		"smtp_from_name",
		"smtp_use_tls",
		"smtp_skip_verify",
	}

	// Batch fetch all SMTP configurations
	configValues, err := s.systemConfigService.GetValuesByKeys(configKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to load SMTP configurations: %v", err)
	}

	// Parse enabled status
	config.Enabled = configValues["smtp_enabled"] == "true"

	if !config.Enabled {
		return config, nil
	}

	// Parse host
	config.Host = configValues["smtp_host"]

	// Parse port
	portStr := configValues["smtp_port"]
	if portStr == "" {
		portStr = "587"
	}
	config.Port, err = strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid smtp_port: %v", err)
	}

	// Parse username and password
	config.Username = configValues["smtp_username"]
	config.Password = configValues["smtp_password"]

	// Parse from address
	config.From = configValues["smtp_from"]

	// Parse from name
	config.FromName = configValues["smtp_from_name"]
	if config.FromName == "" {
		config.FromName = "xsha Platform"
	}

	// Parse TLS setting
	config.UseTLS = configValues["smtp_use_tls"] != "false"

	// Parse skip verify setting
	config.SkipVerify = configValues["smtp_skip_verify"] == "true"

	return config, nil
}

func (s *emailService) isEmailEnabled() (bool, error) {
	config, err := s.loadSMTPConfig()
	if err != nil {
		return false, err
	}

	// Check if all required fields are configured
	if !config.Enabled || config.Host == "" || config.Username == "" || config.Password == "" || config.From == "" {
		return false, nil
	}

	return true, nil
}

func (s *emailService) SendWelcomeEmail(admin *database.Admin, lang string) error {
	go func() {
		if err := s.sendNotificationEmail(admin, "welcome", lang, admin); err != nil {
			utils.Error("Failed to send welcome email", "username", admin.Username, "email", admin.Email, "error", err)
		} else {
			utils.Info("Welcome email sent successfully", "username", admin.Username, "email", admin.Email)
		}
	}()
	return nil
}

func (s *emailService) SendLoginNotificationEmail(admin *database.Admin, clientIP, userAgent, lang string) error {
	go func() {
		loginData := struct {
			*database.Admin
			IPAddress string
			UserAgent string
			LoginTime string
		}{
			Admin:     admin,
			IPAddress: clientIP,
			UserAgent: userAgent,
			LoginTime: time.Now().Format("2006-01-02 15:04:05 MST"),
		}

		if err := s.sendNotificationEmail(admin, "login", lang, loginData); err != nil {
			utils.Error("Failed to send login notification email", "username", admin.Username, "client_ip", clientIP, "error", err)
		}
	}()
	return nil
}

func (s *emailService) SendPasswordChangeEmail(admin *database.Admin, clientIP, userAgent, lang string) error {
	go func() {
		passwordChangeData := struct {
			*database.Admin
			IPAddress  string
			UserAgent  string
			ChangeTime string
		}{
			Admin:      admin,
			IPAddress:  clientIP,
			UserAgent:  userAgent,
			ChangeTime: time.Now().Format("2006-01-02 15:04:05 MST"),
		}

		if err := s.sendNotificationEmail(admin, "password_change", lang, passwordChangeData); err != nil {
			utils.Error("Failed to send password change notification email", "username", admin.Username, "error", err)
		}
	}()
	return nil
}
