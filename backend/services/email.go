package services

import (
	"crypto/tls"
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"xsha-backend/database"
	"xsha-backend/utils"

	"gopkg.in/gomail.v2"
	"gorm.io/gorm"
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

type EmailTemplate struct {
	Subject string
	Body    string
}

type emailService struct {
	systemConfigService SystemConfigService
}

func NewEmailService(systemConfigService SystemConfigService) EmailService {
	return &emailService{
		systemConfigService: systemConfigService,
	}
}

func (s *emailService) SendWelcomeEmail(admin *database.Admin, lang string) error {
	// Check if email service is enabled
	enabled, err := s.isEmailEnabled()
	if err != nil {
		utils.Error("Failed to check if email service is enabled", "error", err)
		return err
	}

	if !enabled {
		utils.Info("Email service is disabled, skipping welcome email", "username", admin.Username)
		return nil
	}

	// Check if admin has email
	if admin.Email == "" {
		utils.Info("Admin has no email address, skipping welcome email", "username", admin.Username)
		return nil
	}

	// Load SMTP configuration
	smtpConfig, err := s.loadSMTPConfig()
	if err != nil {
		utils.Error("Failed to load SMTP configuration", "error", err)
		return err
	}

	// Generate email content
	subject, body, err := s.generateWelcomeEmailContent(admin, lang)
	if err != nil {
		utils.Error("Failed to generate welcome email content", "error", err)
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

	// Get enabled status
	enabledStr, err := s.systemConfigService.GetValue("smtp_enabled")
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get smtp_enabled: %v", err)
	}
	config.Enabled = enabledStr == "true"

	if !config.Enabled {
		return config, nil
	}

	// Get host
	config.Host, err = s.systemConfigService.GetValue("smtp_host")
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get smtp_host: %v", err)
	}

	// Get port
	portStr, err := s.systemConfigService.GetValue("smtp_port")
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get smtp_port: %v", err)
	}
	if portStr == "" {
		portStr = "587"
	}
	config.Port, err = strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid smtp_port: %v", err)
	}

	// Get username
	config.Username, err = s.systemConfigService.GetValue("smtp_username")
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get smtp_username: %v", err)
	}

	// Get password
	config.Password, err = s.systemConfigService.GetValue("smtp_password")
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get smtp_password: %v", err)
	}

	// Get from address
	config.From, err = s.systemConfigService.GetValue("smtp_from")
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get smtp_from: %v", err)
	}

	// Get from name
	config.FromName, err = s.systemConfigService.GetValue("smtp_from_name")
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get smtp_from_name: %v", err)
	}
	if config.FromName == "" {
		config.FromName = "XSha Platform"
	}

	// Get TLS setting
	useTLSStr, err := s.systemConfigService.GetValue("smtp_use_tls")
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get smtp_use_tls: %v", err)
	}
	config.UseTLS = useTLSStr != "false"

	// Get skip verify setting
	skipVerifyStr, err := s.systemConfigService.GetValue("smtp_skip_verify")
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get smtp_skip_verify: %v", err)
	}
	config.SkipVerify = skipVerifyStr == "true"

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

func (s *emailService) generateWelcomeEmailContent(admin *database.Admin, lang string) (string, string, error) {
	// Determine language
	if lang == "" {
		lang = "en-US"
	}

	var subject, bodyTemplate string

	if strings.HasPrefix(lang, "zh") {
		subject = "欢迎加入 XSha 平台！"
		bodyTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>欢迎加入 XSha 平台</title>
    <style>
        body { font-family: 'Microsoft YaHei', Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #007bff; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #f8f9fa; padding: 30px; border-radius: 0 0 8px 8px; }
        .info-box { background: white; padding: 20px; border-radius: 8px; margin: 20px 0; border-left: 4px solid #007bff; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 14px; }
        .btn { display: inline-block; background: #007bff; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; margin: 10px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎉 欢迎加入 XSha 平台！</h1>
        </div>
        <div class="content">
            <p>亲爱的 <strong>{{.Name}}</strong>，</p>
            <p>欢迎您加入 XSha AI 驱动的项目管理和开发平台！您的账户已成功创建。</p>

            <div class="info-box">
                <h3>📋 您的账户信息</h3>
                <p><strong>用户名：</strong>{{.Username}}</p>
                <p><strong>邮箱：</strong>{{.Email}}</p>
                <p><strong>角色：</strong>{{.Role}}</p>
            </div>

            <div class="info-box">
                <h3>🚀 开始使用</h3>
                <p>您现在可以使用您的用户名和密码登录平台：</p>
                <a href="#" class="btn">立即登录</a>
            </div>

            <div class="info-box">
                <h3>💡 平台特性</h3>
                <ul>
                    <li>AI 驱动的任务开发和管理</li>
                    <li>与 Claude Code 的无缝集成</li>
                    <li>Git 仓库和凭证管理</li>
                    <li>Docker 容器化的开发环境</li>
                </ul>
            </div>

            <p>如果您有任何问题或需要帮助，请随时联系我们的支持团队。</p>
            <p>祝您使用愉快！</p>
        </div>
        <div class="footer">
            <p>此邮件由 XSha 平台自动发送，请勿回复。</p>
        </div>
    </div>
</body>
</html>`
	} else {
		subject = "Welcome to XSha Platform!"
		bodyTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Welcome to XSha Platform</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #007bff; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #f8f9fa; padding: 30px; border-radius: 0 0 8px 8px; }
        .info-box { background: white; padding: 20px; border-radius: 8px; margin: 20px 0; border-left: 4px solid #007bff; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 14px; }
        .btn { display: inline-block; background: #007bff; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; margin: 10px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎉 Welcome to XSha Platform!</h1>
        </div>
        <div class="content">
            <p>Dear <strong>{{.Name}}</strong>,</p>
            <p>Welcome to XSha, the AI-driven project management and development platform! Your account has been successfully created.</p>

            <div class="info-box">
                <h3>📋 Your Account Information</h3>
                <p><strong>Username:</strong> {{.Username}}</p>
                <p><strong>Email:</strong> {{.Email}}</p>
                <p><strong>Role:</strong> {{.Role}}</p>
            </div>

            <div class="info-box">
                <h3>🚀 Get Started</h3>
                <p>You can now log in to the platform using your username and password:</p>
                <a href="#" class="btn">Login Now</a>
            </div>

            <div class="info-box">
                <h3>💡 Platform Features</h3>
                <ul>
                    <li>AI-driven task development and management</li>
                    <li>Seamless integration with Claude Code</li>
                    <li>Git repository and credential management</li>
                    <li>Dockerized development environments</li>
                </ul>
            </div>

            <p>If you have any questions or need assistance, please don't hesitate to contact our support team.</p>
            <p>Happy coding!</p>
        </div>
        <div class="footer">
            <p>This email was sent automatically by XSha Platform. Please do not reply.</p>
        </div>
    </div>
</body>
</html>`
	}

	// Parse template
	tmpl, err := template.New("welcome").Parse(bodyTemplate)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse email template: %v", err)
	}

	// Execute template
	var bodyBuilder strings.Builder
	err = tmpl.Execute(&bodyBuilder, admin)
	if err != nil {
		return "", "", fmt.Errorf("failed to execute email template: %v", err)
	}

	return subject, bodyBuilder.String(), nil
}