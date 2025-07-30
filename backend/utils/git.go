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

// GitDiffFile 表示单个文件的变动信息
type GitDiffFile struct {
	Path        string `json:"path"`
	Status      string `json:"status"`       // added, modified, deleted, renamed
	Additions   int    `json:"additions"`    // 新增行数
	Deletions   int    `json:"deletions"`    // 删除行数
	IsBinary    bool   `json:"is_binary"`    // 是否为二进制文件
	OldPath     string `json:"old_path"`     // 重命名前的路径
	DiffContent string `json:"diff_content"` // 具体的diff内容
}

// GitDiffSummary 表示Git差异摘要
type GitDiffSummary struct {
	TotalFiles     int           `json:"total_files"`
	TotalAdditions int           `json:"total_additions"`
	TotalDeletions int           `json:"total_deletions"`
	Files          []GitDiffFile `json:"files"`
	CommitsBehind  int           `json:"commits_behind"` // 落后的提交数
	CommitsAhead   int           `json:"commits_ahead"`  // 领先的提交数
}

// GetBranchDiff 获取两个分支之间的差异
func GetBranchDiff(workspacePath, baseBranch, compareBranch string, includeContent bool) (*GitDiffSummary, error) {
	if workspacePath == "" {
		return nil, fmt.Errorf("workspace path cannot be empty")
	}

	if baseBranch == "" {
		return nil, fmt.Errorf("base branch cannot be empty")
	}

	if compareBranch == "" {
		return nil, fmt.Errorf("compare branch cannot be empty")
	}

	// 检查工作空间目录是否存在
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("workspace directory does not exist: %s", workspacePath)
	}

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	summary := &GitDiffSummary{
		Files: []GitDiffFile{},
	}

	// 1. 获取提交差异统计
	if err := getBranchCommitDiff(ctx, workspacePath, baseBranch, compareBranch, summary); err != nil {
		return nil, fmt.Errorf("failed to get commit diff: %v", err)
	}

	// 2. 获取文件差异统计
	if err := getBranchFileDiff(ctx, workspacePath, baseBranch, compareBranch, summary, includeContent); err != nil {
		return nil, fmt.Errorf("failed to get file diff: %v", err)
	}

	return summary, nil
}

// getBranchCommitDiff 获取分支间的提交差异
func getBranchCommitDiff(ctx context.Context, workspacePath, baseBranch, compareBranch string, summary *GitDiffSummary) error {
	// 获取compare分支相对于base分支的领先提交数
	aheadCmd := exec.CommandContext(ctx, "git", "rev-list", "--count", fmt.Sprintf("%s..%s", baseBranch, compareBranch))
	aheadCmd.Dir = workspacePath

	aheadOutput, err := aheadCmd.Output()
	if err != nil {
		// 如果命令失败，可能是分支不存在，设置为0
		summary.CommitsAhead = 0
	} else {
		if count, parseErr := fmt.Sscanf(strings.TrimSpace(string(aheadOutput)), "%d", &summary.CommitsAhead); parseErr != nil || count != 1 {
			summary.CommitsAhead = 0
		}
	}

	// 获取compare分支相对于base分支的落后提交数
	behindCmd := exec.CommandContext(ctx, "git", "rev-list", "--count", fmt.Sprintf("%s..%s", compareBranch, baseBranch))
	behindCmd.Dir = workspacePath

	behindOutput, err := behindCmd.Output()
	if err != nil {
		summary.CommitsBehind = 0
	} else {
		if count, parseErr := fmt.Sscanf(strings.TrimSpace(string(behindOutput)), "%d", &summary.CommitsBehind); parseErr != nil || count != 1 {
			summary.CommitsBehind = 0
		}
	}

	return nil
}

// getBranchFileDiff 获取分支间的文件差异
func getBranchFileDiff(ctx context.Context, workspacePath, baseBranch, compareBranch string, summary *GitDiffSummary, includeContent bool) error {
	// 获取文件变动统计
	statCmd := exec.CommandContext(ctx, "git", "-c", "core.quotepath=false", "diff", "--numstat", fmt.Sprintf("%s..%s", baseBranch, compareBranch))
	statCmd.Dir = workspacePath

	statOutput, err := statCmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			errorMessage := string(exitError.Stderr)
			return fmt.Errorf("git diff --numstat failed: %s", errorMessage)
		}
		return fmt.Errorf("failed to execute git diff --numstat: %v", err)
	}

	// 解析numstat输出
	files, totalAdditions, totalDeletions := parseNumstat(string(statOutput))
	summary.Files = files
	summary.TotalFiles = len(files)
	summary.TotalAdditions = totalAdditions
	summary.TotalDeletions = totalDeletions

	// 如果需要包含详细内容，获取每个文件的diff内容
	if includeContent {
		for i := range summary.Files {
			content, err := getFileDiffContent(ctx, workspacePath, baseBranch, compareBranch, summary.Files[i].Path)
			if err != nil {
				// 记录错误但不中断整个流程
				Warn("Failed to get diff content for file", "file", summary.Files[i].Path, "error", err)
				continue
			}
			summary.Files[i].DiffContent = content
		}
	}

	return nil
}

// parseNumstat 解析git diff --numstat输出
func parseNumstat(output string) ([]GitDiffFile, int, int) {
	var files []GitDiffFile
	totalAdditions := 0
	totalDeletions := 0

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) < 3 {
			continue
		}

		file := GitDiffFile{
			Path: parts[2],
		}

		// 解析新增和删除行数
		if parts[0] == "-" || parts[1] == "-" {
			// 二进制文件
			file.IsBinary = true
		} else {
			if additions, err := fmt.Sscanf(parts[0], "%d", &file.Additions); err == nil && additions == 1 {
				totalAdditions += file.Additions
			}
			if deletions, err := fmt.Sscanf(parts[1], "%d", &file.Deletions); err == nil && deletions == 1 {
				totalDeletions += file.Deletions
			}
		}

		// 判断文件状态
		if file.Additions > 0 && file.Deletions == 0 {
			file.Status = "added"
		} else if file.Additions == 0 && file.Deletions > 0 {
			file.Status = "deleted"
		} else {
			file.Status = "modified"
		}

		files = append(files, file)
	}

	return files, totalAdditions, totalDeletions
}

// getFileDiffContent 获取单个文件的diff内容
func getFileDiffContent(ctx context.Context, workspacePath, baseBranch, compareBranch, filePath string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "-c", "core.quotepath=false", "diff", fmt.Sprintf("%s..%s", baseBranch, compareBranch), "--", filePath)
	cmd.Dir = workspacePath

	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			errorMessage := string(exitError.Stderr)
			return "", fmt.Errorf("git diff failed for file %s: %s", filePath, errorMessage)
		}
		return "", fmt.Errorf("failed to execute git diff for file %s: %v", filePath, err)
	}

	return string(output), nil
}

// ValidateBranchExists 验证分支是否存在
func ValidateBranchExists(workspacePath, branchName string) error {
	if workspacePath == "" {
		return fmt.Errorf("workspace path cannot be empty")
	}

	if branchName == "" {
		return fmt.Errorf("branch name cannot be empty")
	}

	// 检查工作空间目录是否存在
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		return fmt.Errorf("workspace directory does not exist: %s", workspacePath)
	}

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 检查分支是否存在
	cmd := exec.CommandContext(ctx, "git", "show-ref", "--verify", "--quiet", fmt.Sprintf("refs/heads/%s", branchName))
	cmd.Dir = workspacePath

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("branch '%s' does not exist", branchName)
	}

	return nil
}

// GetCommitDiff 获取指定提交的变动差异（与其父提交比较）
func GetCommitDiff(workspacePath, commitHash string, includeContent bool) (*GitDiffSummary, error) {
	if workspacePath == "" {
		return nil, fmt.Errorf("workspace path cannot be empty")
	}

	if commitHash == "" {
		return nil, fmt.Errorf("commit hash cannot be empty")
	}

	// 检查工作空间目录是否存在
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("workspace directory does not exist: %s", workspacePath)
	}

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 验证 commit hash 是否存在
	if err := validateCommitExists(ctx, workspacePath, commitHash); err != nil {
		return nil, fmt.Errorf("commit validation failed: %v", err)
	}

	summary := &GitDiffSummary{
		Files: []GitDiffFile{},
	}

	// 获取文件差异统计（与父提交比较）
	if err := getCommitFileDiff(ctx, workspacePath, commitHash, summary, includeContent); err != nil {
		return nil, fmt.Errorf("failed to get commit diff: %v", err)
	}

	return summary, nil
}

// validateCommitExists 验证提交是否存在
func validateCommitExists(ctx context.Context, workspacePath, commitHash string) error {
	cmd := exec.CommandContext(ctx, "git", "cat-file", "-t", commitHash)
	cmd.Dir = workspacePath

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("commit '%s' does not exist", commitHash)
	}

	if strings.TrimSpace(string(output)) != "commit" {
		return fmt.Errorf("'%s' is not a valid commit", commitHash)
	}

	return nil
}

// getCommitFileDiff 获取提交的文件差异
func getCommitFileDiff(ctx context.Context, workspacePath, commitHash string, summary *GitDiffSummary, includeContent bool) error {
	// 获取文件变动统计（与父提交比较）
	statCmd := exec.CommandContext(ctx, "git", "-c", "core.quotepath=false", "diff", "--numstat", commitHash+"^", commitHash)
	statCmd.Dir = workspacePath

	statOutput, err := statCmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			errorMessage := string(exitError.Stderr)
			return fmt.Errorf("git diff --numstat failed: %s", errorMessage)
		}
		return fmt.Errorf("failed to execute git diff --numstat: %v", err)
	}

	// 解析numstat输出
	files, totalAdditions, totalDeletions := parseNumstat(string(statOutput))
	summary.Files = files
	summary.TotalFiles = len(files)
	summary.TotalAdditions = totalAdditions
	summary.TotalDeletions = totalDeletions

	// 如果需要包含详细内容，获取每个文件的diff内容
	if includeContent {
		for i := range summary.Files {
			content, err := getCommitFileDiffContent(ctx, workspacePath, commitHash, summary.Files[i].Path)
			if err != nil {
				// 记录错误但不中断整个流程
				Warn("Failed to get diff content for file", "file", summary.Files[i].Path, "error", err)
				continue
			}
			summary.Files[i].DiffContent = content
		}
	}

	return nil
}

// getCommitFileDiffContent 获取提交中单个文件的diff内容
func getCommitFileDiffContent(ctx context.Context, workspacePath, commitHash, filePath string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "-c", "core.quotepath=false", "diff", commitHash+"^", commitHash, "--", filePath)
	cmd.Dir = workspacePath

	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			errorMessage := string(exitError.Stderr)
			return "", fmt.Errorf("git diff failed for file %s: %s", filePath, errorMessage)
		}
		return "", fmt.Errorf("failed to execute git diff for file %s: %v", filePath, err)
	}

	return string(output), nil
}

// GetCommitFileDiff 获取提交中指定文件的变动内容
func GetCommitFileDiff(workspacePath, commitHash, filePath string) (string, error) {
	if workspacePath == "" {
		return "", fmt.Errorf("workspace path cannot be empty")
	}

	if commitHash == "" {
		return "", fmt.Errorf("commit hash cannot be empty")
	}

	if filePath == "" {
		return "", fmt.Errorf("file path cannot be empty")
	}

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 验证 commit hash 是否存在
	if err := validateCommitExists(ctx, workspacePath, commitHash); err != nil {
		return "", err
	}

	return getCommitFileDiffContent(ctx, workspacePath, commitHash, filePath)
}
