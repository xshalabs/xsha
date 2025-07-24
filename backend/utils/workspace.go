package utils

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// WorkspaceManager 工作目录管理器
type WorkspaceManager struct {
	baseDir string
}

// NewWorkspaceManager 创建工作目录管理器
func NewWorkspaceManager(baseDir string) *WorkspaceManager {
	if baseDir == "" {
		baseDir = "/tmp/sleep0-workspaces"
	}
	return &WorkspaceManager{baseDir: baseDir}
}

// CreateTempWorkspace 创建临时工作目录
func (w *WorkspaceManager) CreateTempWorkspace(conversationID uint) (string, error) {
	// 创建基础目录
	if err := os.MkdirAll(w.baseDir, 0755); err != nil {
		return "", fmt.Errorf("创建基础目录失败: %v", err)
	}

	// 生成唯一目录名
	dirName := fmt.Sprintf("conversation-%d-%d", conversationID, time.Now().Unix())
	workspacePath := filepath.Join(w.baseDir, dirName)

	// 创建工作目录
	if err := os.MkdirAll(workspacePath, 0755); err != nil {
		return "", fmt.Errorf("创建工作目录失败: %v", err)
	}

	return workspacePath, nil
}

// CloneRepositoryWithConfig 克隆仓库到工作目录（带配置）
func (w *WorkspaceManager) CloneRepositoryWithConfig(workspacePath, repoURL, branch string, credential *GitCredentialInfo, sslVerify bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	var cmd *exec.Cmd
	var envVars []string

	if credential != nil {
		switch credential.Type {
		case GitCredentialTypePassword, GitCredentialTypeToken:
			// HTTPS 认证
			authenticatedURL, err := w.buildAuthenticatedURL(repoURL, credential)
			if err != nil {
				return err
			}
			cmd = exec.CommandContext(ctx, "git", "clone", "-b", branch, authenticatedURL, workspacePath)

		case GitCredentialTypeSSHKey:
			// SSH 认证
			keyFile := filepath.Join(workspacePath, ".ssh_key")
			if err := ioutil.WriteFile(keyFile, []byte(credential.PrivateKey), 0600); err != nil {
				return fmt.Errorf("创建SSH密钥文件失败: %v", err)
			}
			defer os.Remove(keyFile)

			envVars = append(os.Environ(),
				fmt.Sprintf("GIT_SSH_COMMAND=ssh -i %s -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no", keyFile),
			)
			cmd = exec.CommandContext(ctx, "git", "clone", "-b", branch, repoURL, workspacePath)
			cmd.Env = envVars
		}
	} else {
		// 无认证
		cmd = exec.CommandContext(ctx, "git", "clone", "-b", branch, repoURL, workspacePath)
	}

	// 设置Git环境变量
	if cmd.Env == nil {
		cmd.Env = os.Environ()
	}

	// 根据配置决定是否禁用SSL验证
	if !sslVerify {
		cmd.Env = append(cmd.Env, "GIT_SSL_NO_VERIFY=true")
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("克隆仓库失败: %v", err)
	}

	return nil
}

// CloneRepository 克隆仓库到工作目录（保持向后兼容）
func (w *WorkspaceManager) CloneRepository(workspacePath, repoURL, branch string, credential *GitCredentialInfo) error {
	// 默认禁用SSL验证以解决兼容性问题
	return w.CloneRepositoryWithConfig(workspacePath, repoURL, branch, credential, false)
}

// CommitChanges 提交更改并返回commit hash
func (w *WorkspaceManager) CommitChanges(workspacePath, message string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// 配置Git用户信息（临时）
	configCmd1 := exec.CommandContext(ctx, "git", "config", "user.name", "Sleep0 AI")
	configCmd1.Dir = workspacePath
	if err := configCmd1.Run(); err != nil {
		return "", fmt.Errorf("配置Git用户名失败: %v", err)
	}

	configCmd2 := exec.CommandContext(ctx, "git", "config", "user.email", "ai@sleep0.dev")
	configCmd2.Dir = workspacePath
	if err := configCmd2.Run(); err != nil {
		return "", fmt.Errorf("配置Git邮箱失败: %v", err)
	}

	// 添加所有更改
	addCmd := exec.CommandContext(ctx, "git", "add", ".")
	addCmd.Dir = workspacePath
	if err := addCmd.Run(); err != nil {
		return "", fmt.Errorf("添加更改失败: %v", err)
	}

	// 检查是否有更改需要提交
	statusCmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	statusCmd.Dir = workspacePath
	statusOutput, err := statusCmd.Output()
	if err != nil {
		return "", fmt.Errorf("检查Git状态失败: %v", err)
	}

	if len(strings.TrimSpace(string(statusOutput))) == 0 {
		return "", fmt.Errorf("没有更改需要提交")
	}

	// 提交更改
	commitCmd := exec.CommandContext(ctx, "git", "commit", "-m", message)
	commitCmd.Dir = workspacePath
	if err := commitCmd.Run(); err != nil {
		return "", fmt.Errorf("提交更改失败: %v", err)
	}

	// 获取commit hash
	hashCmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	hashCmd.Dir = workspacePath
	output, err := hashCmd.Output()
	if err != nil {
		return "", fmt.Errorf("获取commit hash失败: %v", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// CleanupWorkspace 清理工作目录
func (w *WorkspaceManager) CleanupWorkspace(workspacePath string) error {
	return os.RemoveAll(workspacePath)
}

// buildAuthenticatedURL 构建带认证信息的URL
func (w *WorkspaceManager) buildAuthenticatedURL(repoURL string, credential *GitCredentialInfo) (string, error) {
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return "", fmt.Errorf("解析URL失败: %v", err)
	}

	switch credential.Type {
	case GitCredentialTypePassword:
		if credential.Password == "" {
			return "", fmt.Errorf("密码不能为空")
		}
		parsedURL.User = url.UserPassword(credential.Username, credential.Password)

	case GitCredentialTypeToken:
		if credential.Password == "" {
			return "", fmt.Errorf("Token不能为空")
		}
		// GitHub风格：token作为用户名，密码为空或x-oauth-basic
		parsedURL.User = url.UserPassword(credential.Password, "x-oauth-basic")

	default:
		return "", fmt.Errorf("不支持的凭据类型用于URL构建")
	}

	return parsedURL.String(), nil
}

// CheckWorkspaceExists 检查工作目录是否存在
func (w *WorkspaceManager) CheckWorkspaceExists(workspacePath string) bool {
	if workspacePath == "" {
		return false
	}

	info, err := os.Stat(workspacePath)
	return err == nil && info.IsDir()
}

// GetWorkspaceSize 获取工作目录大小（MB）
func (w *WorkspaceManager) GetWorkspaceSize(workspacePath string) (int64, error) {
	var size int64
	err := filepath.Walk(workspacePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		size += info.Size()
		return nil
	})

	if err != nil {
		return 0, err
	}

	// 转换为MB
	return size / (1024 * 1024), nil
}

// CleanupOldWorkspaces 清理超过指定天数的工作目录
func (w *WorkspaceManager) CleanupOldWorkspaces(days int) error {
	if days <= 0 {
		return fmt.Errorf("天数必须大于0")
	}

	cutoffTime := time.Now().AddDate(0, 0, -days)

	entries, err := os.ReadDir(w.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 目录不存在，无需清理
		}
		return fmt.Errorf("读取基础目录失败: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// 检查目录是否符合命名模式 conversation-*
		if !strings.HasPrefix(entry.Name(), "conversation-") {
			continue
		}

		dirPath := filepath.Join(w.baseDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// 如果目录修改时间早于截止时间，则删除
		if info.ModTime().Before(cutoffTime) {
			if err := os.RemoveAll(dirPath); err != nil {
				fmt.Printf("清理目录失败 %s: %v\n", dirPath, err)
			}
		}
	}

	return nil
}
