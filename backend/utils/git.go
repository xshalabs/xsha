package utils

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// GitProtocolType 定义Git协议类型
type GitProtocolType string

const (
	GitProtocolHTTPS GitProtocolType = "https" // HTTPS协议
	GitProtocolSSH   GitProtocolType = "ssh"   // SSH协议
)

// GitURLInfo Git URL 信息
type GitURLInfo struct {
	Protocol GitProtocolType `json:"protocol"`
	Host     string          `json:"host"`
	Owner    string          `json:"owner"`
	Repo     string          `json:"repo"`
	IsValid  bool            `json:"is_valid"`
}

// DetectGitProtocol 根据 Git URL 自动检测协议类型
func DetectGitProtocol(repoURL string) GitProtocolType {
	// 去除首尾空格
	repoURL = strings.TrimSpace(repoURL)

	if repoURL == "" {
		return GitProtocolHTTPS // 默认返回 HTTPS
	}

	// 检测 HTTPS 协议
	if strings.HasPrefix(repoURL, "https://") {
		return GitProtocolHTTPS
	}

	// 检测 HTTP 协议（也归类为 HTTPS）
	if strings.HasPrefix(repoURL, "http://") {
		return GitProtocolHTTPS
	}

	// 检测 SSH 协议格式：ssh://user@host/path
	if strings.HasPrefix(repoURL, "ssh://") {
		return GitProtocolSSH
	}

	// 检测 SSH 协议格式：user@host:path
	sshPattern := regexp.MustCompile(`^[a-zA-Z0-9_.-]+@[a-zA-Z0-9.-]+:[a-zA-Z0-9/._-]+`)
	if sshPattern.MatchString(repoURL) {
		return GitProtocolSSH
	}

	// 如果都不匹配，默认返回 HTTPS
	return GitProtocolHTTPS
}

// ParseGitURL 解析 Git URL 并提取详细信息
func ParseGitURL(repoURL string) *GitURLInfo {
	info := &GitURLInfo{
		IsValid: false,
	}

	// 去除首尾空格
	repoURL = strings.TrimSpace(repoURL)

	// 检测协议类型（即使是空字符串也要设置默认协议）
	info.Protocol = DetectGitProtocol(repoURL)

	if repoURL == "" {
		return info
	}

	switch info.Protocol {
	case GitProtocolHTTPS:
		return parseHTTPSURL(repoURL, info)
	case GitProtocolSSH:
		return parseSSHURL(repoURL, info)
	default:
		return info
	}
}

// parseHTTPSURL 解析 HTTPS 格式的 Git URL
func parseHTTPSURL(repoURL string, info *GitURLInfo) *GitURLInfo {
	// 解析 URL
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return info
	}

	info.Host = parsedURL.Host

	// 解析路径：通常格式为 /owner/repo.git 或 /owner/repo
	path := strings.Trim(parsedURL.Path, "/")

	// 移除 .git 后缀
	if strings.HasSuffix(path, ".git") {
		path = strings.TrimSuffix(path, ".git")
	}

	// 分割路径
	parts := strings.Split(path, "/")
	if len(parts) >= 2 {
		info.Owner = parts[0]
		info.Repo = parts[1]
		info.IsValid = true
	}

	return info
}

// parseSSHURL 解析 SSH 格式的 Git URL
func parseSSHURL(repoURL string, info *GitURLInfo) *GitURLInfo {
	// 处理 ssh://user@host/path 格式
	if strings.HasPrefix(repoURL, "ssh://") {
		parsedURL, err := url.Parse(repoURL)
		if err != nil {
			return info
		}

		info.Host = parsedURL.Host

		// 解析路径
		path := strings.Trim(parsedURL.Path, "/")

		// 移除 .git 后缀
		if strings.HasSuffix(path, ".git") {
			path = strings.TrimSuffix(path, ".git")
		}

		// 分割路径
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			info.Owner = parts[0]
			info.Repo = parts[1]
			info.IsValid = true
		}

		return info
	}

	// 处理 user@host:path 格式
	sshPattern := regexp.MustCompile(`^([a-zA-Z0-9_.-]+)@([a-zA-Z0-9.-]+):(.+)$`)
	matches := sshPattern.FindStringSubmatch(repoURL)
	if len(matches) == 4 {
		info.Host = matches[2]
		path := matches[3]

		// 移除 .git 后缀
		if strings.HasSuffix(path, ".git") {
			path = strings.TrimSuffix(path, ".git")
		}

		// 分割路径
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			info.Owner = parts[0]
			info.Repo = parts[1]
			info.IsValid = true
		}
	}

	return info
}

// ValidateGitURL 验证 Git URL 的有效性
func ValidateGitURL(repoURL string) error {
	info := ParseGitURL(repoURL)
	if !info.IsValid {
		return fmt.Errorf("invalid Git URL format")
	}
	return nil
}

// IsGitURL 检查字符串是否像是一个 Git URL
func IsGitURL(str string) bool {
	str = strings.TrimSpace(str)

	// 检查 HTTPS 格式
	if strings.HasPrefix(str, "https://") || strings.HasPrefix(str, "http://") {
		return true
	}

	// 检查 SSH 格式
	if strings.HasPrefix(str, "ssh://") {
		return true
	}

	// 检查 SSH 简化格式 user@host:path
	sshPattern := regexp.MustCompile(`^[a-zA-Z0-9_.-]+@[a-zA-Z0-9.-]+:[a-zA-Z0-9/._-]+`)
	return sshPattern.MatchString(str)
}
