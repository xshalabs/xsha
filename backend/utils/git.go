package utils

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// GitProtocolType 定义Git协议类型
type GitProtocolType string

const (
	GitProtocolHTTPS GitProtocolType = "https" // HTTPS协议
	GitProtocolSSH   GitProtocolType = "ssh"   // SSH协议
)

// GitCredentialType 定义Git凭据类型
type GitCredentialType string

const (
	GitCredentialTypePassword GitCredentialType = "password" // 用户名密码
	GitCredentialTypeToken    GitCredentialType = "token"    // 访问令牌
	GitCredentialTypeSSHKey   GitCredentialType = "ssh_key"  // SSH密钥
)

// GitURLInfo Git URL 信息
type GitURLInfo struct {
	Protocol GitProtocolType `json:"protocol"`
	Host     string          `json:"host"`
	Owner    string          `json:"owner"`
	Repo     string          `json:"repo"`
	IsValid  bool            `json:"is_valid"`
}

// GitBranch 分支信息
type GitBranch struct {
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
	CommitID  string `json:"commit_id"`
}

// GitAccessResult 访问验证结果
type GitAccessResult struct {
	CanAccess    bool     `json:"can_access"`
	ErrorMessage string   `json:"error_message"`
	Branches     []string `json:"branches"`
}

// GitCredentialInfo 凭据信息（用于传递给Git操作函数）
type GitCredentialInfo struct {
	Type       GitCredentialType `json:"type"`
	Username   string            `json:"username"`
	Password   string            `json:"password"`    // 明文密码或token
	PrivateKey string            `json:"private_key"` // 明文SSH私钥
	PublicKey  string            `json:"public_key"`  // SSH公钥
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

// FetchRepositoryBranchesWithConfig 获取仓库分支列表（带配置）
func FetchRepositoryBranchesWithConfig(repoURL string, credential *GitCredentialInfo, sslVerify bool) (*GitAccessResult, error) {
	// 创建临时目录
	tempDir, err := ioutil.TempDir("", "git-repo-*")
	if err != nil {
		return &GitAccessResult{
			CanAccess:    false,
			ErrorMessage: fmt.Sprintf("创建临时目录失败: %v", err),
		}, nil
	}
	defer os.RemoveAll(tempDir)

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	var envVars []string

	// 根据协议类型和凭据类型设置认证
	if credential != nil {
		switch credential.Type {
		case GitCredentialTypePassword:
			// HTTPS 用户名密码认证
			if credential.Password == "" {
				return &GitAccessResult{
					CanAccess:    false,
					ErrorMessage: "密码不能为空",
				}, nil
			}

			// 构建带认证信息的URL
			parsedURL, err := url.Parse(repoURL)
			if err != nil {
				return &GitAccessResult{
					CanAccess:    false,
					ErrorMessage: fmt.Sprintf("解析URL失败: %v", err),
				}, nil
			}

			parsedURL.User = url.UserPassword(credential.Username, credential.Password)
			authenticatedURL := parsedURL.String()

			cmd = exec.CommandContext(ctx, "git", "ls-remote", "--heads", authenticatedURL)

		case GitCredentialTypeToken:
			// HTTPS Token认证
			if credential.Password == "" {
				return &GitAccessResult{
					CanAccess:    false,
					ErrorMessage: "Token不能为空",
				}, nil
			}

			// 构建带Token的URL (GitHub风格)
			parsedURL, err := url.Parse(repoURL)
			if err != nil {
				return &GitAccessResult{
					CanAccess:    false,
					ErrorMessage: fmt.Sprintf("解析URL失败: %v", err),
				}, nil
			}

			parsedURL.User = url.UserPassword(credential.Password, "x-oauth-basic")
			authenticatedURL := parsedURL.String()

			cmd = exec.CommandContext(ctx, "git", "ls-remote", "--heads", authenticatedURL)

		case GitCredentialTypeSSHKey:
			// SSH密钥认证
			if credential.PrivateKey == "" {
				return &GitAccessResult{
					CanAccess:    false,
					ErrorMessage: "SSH私钥不能为空",
				}, nil
			}

			// 创建临时SSH密钥文件
			keyFile := filepath.Join(tempDir, "ssh_key")
			if err := ioutil.WriteFile(keyFile, []byte(credential.PrivateKey), 0600); err != nil {
				return &GitAccessResult{
					CanAccess:    false,
					ErrorMessage: fmt.Sprintf("创建SSH密钥文件失败: %v", err),
				}, nil
			}

			// 设置SSH环境变量
			envVars = append(os.Environ(),
				fmt.Sprintf("GIT_SSH_COMMAND=ssh -i %s -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no", keyFile),
			)

			cmd = exec.CommandContext(ctx, "git", "ls-remote", "--heads", repoURL)
			cmd.Env = envVars

		default:
			return &GitAccessResult{
				CanAccess:    false,
				ErrorMessage: "不支持的凭据类型",
			}, nil
		}
	} else {
		// 无凭据，尝试匿名访问
		cmd = exec.CommandContext(ctx, "git", "ls-remote", "--heads", repoURL)
	}

	// 设置Git环境变量
	if cmd.Env == nil {
		cmd.Env = os.Environ()
	}

	// 根据配置决定是否禁用SSL验证
	if !sslVerify {
		cmd.Env = append(cmd.Env, "GIT_SSL_NO_VERIFY=true")
	}

	// 执行Git命令
	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			errorMessage := string(exitError.Stderr)
			// 如果是SSL相关错误且当前启用了SSL验证，建议禁用SSL验证
			if sslVerify && (strings.Contains(errorMessage, "SSL") || strings.Contains(errorMessage, "TLS") || strings.Contains(errorMessage, "certificate")) {
				return &GitAccessResult{
					CanAccess:    false,
					ErrorMessage: fmt.Sprintf("仓库访问验证失败: %s\n建议: 可尝试设置环境变量 XSHA_GIT_SSL_VERIFY=false 禁用SSL验证", errorMessage),
				}, nil
			}
			return &GitAccessResult{
				CanAccess:    false,
				ErrorMessage: fmt.Sprintf("仓库访问验证失败: %s", errorMessage),
			}, nil
		}
		return &GitAccessResult{
			CanAccess:    false,
			ErrorMessage: fmt.Sprintf("执行Git命令失败: %v", err),
		}, nil
	}

	// 解析分支信息
	branches := parseBranchesFromLsRemote(string(output))

	return &GitAccessResult{
		CanAccess: true,
		Branches:  branches,
	}, nil
}

// FetchRepositoryBranches 获取仓库分支列表（保持向后兼容）
func FetchRepositoryBranches(repoURL string, credential *GitCredentialInfo) (*GitAccessResult, error) {
	// 默认禁用SSL验证以解决兼容性问题
	return FetchRepositoryBranchesWithConfig(repoURL, credential, false)
}

// parseBranchesFromLsRemote 从git ls-remote输出解析分支名称
func parseBranchesFromLsRemote(output string) []string {
	var branches []string
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		// git ls-remote --heads 输出格式: <commit_hash>\trefs/heads/<branch_name>
		parts := strings.Split(line, "\t")
		if len(parts) == 2 && strings.HasPrefix(parts[1], "refs/heads/") {
			branchName := strings.TrimPrefix(parts[1], "refs/heads/")
			if branchName != "" {
				branches = append(branches, branchName)
			}
		}
	}

	return branches
}

// ValidateRepositoryAccess 验证仓库访问权限（不获取分支列表，仅验证访问）
func ValidateRepositoryAccess(repoURL string, credential *GitCredentialInfo) error {
	result, err := FetchRepositoryBranches(repoURL, credential)
	if err != nil {
		return err
	}

	if !result.CanAccess {
		return fmt.Errorf(result.ErrorMessage)
	}

	return nil
}

// GitResetToPreviousCommit 将仓库重置到指定 commit 的前一个提交
func GitResetToPreviousCommit(workspacePath, commitHash string) error {
	if workspacePath == "" {
		return fmt.Errorf("workspace path cannot be empty")
	}

	if commitHash == "" {
		return fmt.Errorf("commit hash cannot be empty")
	}

	// 检查工作空间目录是否存在
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		return fmt.Errorf("workspace directory does not exist: %s", workspacePath)
	}

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 首先获取指定 commit 的前一个提交
	getPrevCmd := exec.CommandContext(ctx, "git", "rev-parse", commitHash+"^")
	getPrevCmd.Dir = workspacePath

	prevCommitOutput, err := getPrevCmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			errorMessage := string(exitError.Stderr)
			return fmt.Errorf("failed to get previous commit: %s", errorMessage)
		}
		return fmt.Errorf("failed to execute git rev-parse command: %v", err)
	}

	prevCommitHash := strings.TrimSpace(string(prevCommitOutput))
	if prevCommitHash == "" {
		return fmt.Errorf("failed to get previous commit hash")
	}

	// 执行 git reset --hard 到前一个提交
	resetCmd := exec.CommandContext(ctx, "git", "reset", "--hard", prevCommitHash)
	resetCmd.Dir = workspacePath

	_, err = resetCmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			errorMessage := string(exitError.Stderr)
			return fmt.Errorf("failed to reset to previous commit: %s", errorMessage)
		}
		return fmt.Errorf("failed to execute git reset command: %v", err)
	}

	return nil
}
