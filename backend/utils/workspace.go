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
		baseDir = "/tmp/xsha-workspaces"
	}
	return &WorkspaceManager{baseDir: baseDir}
}

// CreateTaskWorkspace 创建任务级工作目录
func (w *WorkspaceManager) CreateTaskWorkspace(taskID uint) (string, error) {
	// 创建基础目录
	if err := os.MkdirAll(w.baseDir, 0755); err != nil {
		return "", fmt.Errorf("创建基础目录失败: %v", err)
	}

	// 生成任务工作目录名
	dirName := fmt.Sprintf("task-%d-%d", taskID, time.Now().Unix())
	workspacePath := filepath.Join(w.baseDir, dirName)

	// 创建工作目录
	if err := os.MkdirAll(workspacePath, 0755); err != nil {
		return "", fmt.Errorf("创建工作目录失败: %v", err)
	}

	return workspacePath, nil
}

// GetOrCreateTaskWorkspace 获取或创建任务工作目录
func (w *WorkspaceManager) GetOrCreateTaskWorkspace(taskID uint, existingPath string) (string, error) {
	// 如果已有工作空间路径且目录存在，直接返回
	if existingPath != "" && w.CheckWorkspaceExists(existingPath) {
		return existingPath, nil
	}

	// 否则创建新的工作空间
	return w.CreateTaskWorkspace(taskID)
}

// CleanupTaskWorkspace 清理任务工作目录
func (w *WorkspaceManager) CleanupTaskWorkspace(workspacePath string) error {
	if workspacePath == "" {
		return nil
	}
	return os.RemoveAll(workspacePath)
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
	configCmd1 := exec.CommandContext(ctx, "git", "config", "user.name", "XSHA AI")
	configCmd1.Dir = workspacePath
	if err := configCmd1.Run(); err != nil {
		return "", fmt.Errorf("配置Git用户名失败: %v", err)
	}

	configCmd2 := exec.CommandContext(ctx, "git", "config", "user.email", "ai@xsha.dev")
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

// CheckGitRepositoryExists 检查工作空间中是否已存在git仓库
func (w *WorkspaceManager) CheckGitRepositoryExists(workspacePath string) bool {
	if workspacePath == "" {
		return false
	}

	// 检查.git目录是否存在
	gitDir := filepath.Join(workspacePath, ".git")
	info, err := os.Stat(gitDir)
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

		// 检查目录是否符合工作空间命名模式
		// conversation-* : 旧格式，保留用于向后兼容和清理历史数据
		// task-* : 新格式，当前使用的任务级工作空间
		if !strings.HasPrefix(entry.Name(), "conversation-") && !strings.HasPrefix(entry.Name(), "task-") {
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

// ResetWorkspaceToCleanState 重置工作空间到干净状态，清理所有未提交的变更
// 用于在任务对话被取消或执行失败时防止文件污染
func (w *WorkspaceManager) ResetWorkspaceToCleanState(workspacePath string) error {
	if workspacePath == "" {
		return fmt.Errorf("工作空间路径不能为空")
	}

	// 检查工作空间是否存在
	if !w.CheckWorkspaceExists(workspacePath) {
		return fmt.Errorf("工作空间不存在: %s", workspacePath)
	}

	// 检查是否为 Git 仓库
	if !w.CheckGitRepositoryExists(workspacePath) {
		// 如果不是 Git 仓库，直接清理整个目录并重新创建
		if err := os.RemoveAll(workspacePath); err != nil {
			return fmt.Errorf("清理非Git工作空间失败: %v", err)
		}
		if err := os.MkdirAll(workspacePath, 0755); err != nil {
			return fmt.Errorf("重新创建工作空间失败: %v", err)
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// 1. 重置所有已暂存的更改
	resetStagedCmd := exec.CommandContext(ctx, "git", "reset", "HEAD", ".")
	resetStagedCmd.Dir = workspacePath
	if err := resetStagedCmd.Run(); err != nil {
		// 如果没有staged文件，git reset会返回错误，但这是正常的
		Info("重置暂存区", "workspace", workspacePath, "note", "可能没有暂存的文件")
	}

	// 2. 重置工作目录到最后一次提交的状态
	resetHardCmd := exec.CommandContext(ctx, "git", "reset", "--hard", "HEAD")
	resetHardCmd.Dir = workspacePath
	if err := resetHardCmd.Run(); err != nil {
		return fmt.Errorf("重置工作目录失败: %v", err)
	}

	// 3. 清理所有未跟踪的文件和目录
	cleanCmd := exec.CommandContext(ctx, "git", "clean", "-fd")
	cleanCmd.Dir = workspacePath
	if err := cleanCmd.Run(); err != nil {
		return fmt.Errorf("清理未跟踪文件失败: %v", err)
	}

	// 4. 清理忽略的文件（可选，更彻底的清理）
	cleanIgnoredCmd := exec.CommandContext(ctx, "git", "clean", "-fdx")
	cleanIgnoredCmd.Dir = workspacePath
	if err := cleanIgnoredCmd.Run(); err != nil {
		// 清理忽略文件失败不应该中断整个流程，只记录警告
		Warn("清理忽略文件失败", "workspace", workspacePath, "error", err.Error())
	}

	Info("工作空间已重置到干净状态", "workspace", workspacePath)
	return nil
}

// CheckWorkspaceIsDirty 检查工作空间是否有未提交的更改
func (w *WorkspaceManager) CheckWorkspaceIsDirty(workspacePath string) (bool, error) {
	if workspacePath == "" {
		return false, fmt.Errorf("工作空间路径不能为空")
	}

	// 检查工作空间是否存在
	if !w.CheckWorkspaceExists(workspacePath) {
		return false, fmt.Errorf("工作空间不存在: %s", workspacePath)
	}

	// 检查是否为 Git 仓库
	if !w.CheckGitRepositoryExists(workspacePath) {
		return false, fmt.Errorf("不是Git仓库: %s", workspacePath)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 检查Git状态
	statusCmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	statusCmd.Dir = workspacePath
	output, err := statusCmd.Output()
	if err != nil {
		return false, fmt.Errorf("检查Git状态失败: %v", err)
	}

	// 如果有输出，说明有未提交的更改
	return len(strings.TrimSpace(string(output))) > 0, nil
}

// CreateAndSwitchToBranch 创建新分支并切换到该分支
func (w *WorkspaceManager) CreateAndSwitchToBranch(workspacePath, branchName, baseBranch string) error {
	if workspacePath == "" {
		return fmt.Errorf("工作空间路径不能为空")
	}

	if branchName == "" {
		return fmt.Errorf("分支名不能为空")
	}

	if baseBranch == "" {
		baseBranch = "main" // 默认基于 main 分支
	}

	// 检查工作空间是否存在
	if !w.CheckWorkspaceExists(workspacePath) {
		return fmt.Errorf("工作空间不存在: %s", workspacePath)
	}

	// 检查是否为 Git 仓库
	if !w.CheckGitRepositoryExists(workspacePath) {
		return fmt.Errorf("不是Git仓库: %s", workspacePath)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// 1. 确保在基础分支上
	switchCmd := exec.CommandContext(ctx, "git", "checkout", baseBranch)
	switchCmd.Dir = workspacePath
	if err := switchCmd.Run(); err != nil {
		return fmt.Errorf("切换到基础分支 %s 失败: %v", baseBranch, err)
	}

	// 2. 拉取最新代码
	pullCmd := exec.CommandContext(ctx, "git", "pull", "origin", baseBranch)
	pullCmd.Dir = workspacePath
	if err := pullCmd.Run(); err != nil {
		// 忽略拉取错误，可能是没有远程分支或网络问题
		Warn("拉取最新代码失败", "workspace", workspacePath, "baseBranch", baseBranch, "error", err)
	}

	// 3. 检查分支是否已存在
	exists, err := w.CheckBranchExists(workspacePath, branchName)
	if err != nil {
		return fmt.Errorf("检查分支是否存在失败: %v", err)
	}

	if exists {
		// 分支已存在，直接切换
		switchExistingCmd := exec.CommandContext(ctx, "git", "checkout", branchName)
		switchExistingCmd.Dir = workspacePath
		if err := switchExistingCmd.Run(); err != nil {
			return fmt.Errorf("切换到已存在的分支 %s 失败: %v", branchName, err)
		}
		Info("切换到已存在的工作分支", "workspace", workspacePath, "branch", branchName)
	} else {
		// 分支不存在，创建新分支并切换
		createCmd := exec.CommandContext(ctx, "git", "checkout", "-b", branchName)
		createCmd.Dir = workspacePath
		if err := createCmd.Run(); err != nil {
			return fmt.Errorf("创建并切换到分支 %s 失败: %v", branchName, err)
		}
		Info("创建并切换到新工作分支", "workspace", workspacePath, "branch", branchName, "baseBranch", baseBranch)
	}

	return nil
}

// SwitchBranch 切换到指定分支
func (w *WorkspaceManager) SwitchBranch(workspacePath, branchName string) error {
	if workspacePath == "" {
		return fmt.Errorf("工作空间路径不能为空")
	}

	if branchName == "" {
		return fmt.Errorf("分支名不能为空")
	}

	// 检查工作空间是否存在
	if !w.CheckWorkspaceExists(workspacePath) {
		return fmt.Errorf("工作空间不存在: %s", workspacePath)
	}

	// 检查是否为 Git 仓库
	if !w.CheckGitRepositoryExists(workspacePath) {
		return fmt.Errorf("不是Git仓库: %s", workspacePath)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// 切换分支
	switchCmd := exec.CommandContext(ctx, "git", "checkout", branchName)
	switchCmd.Dir = workspacePath
	if err := switchCmd.Run(); err != nil {
		return fmt.Errorf("切换到分支 %s 失败: %v", branchName, err)
	}

	Info("成功切换分支", "workspace", workspacePath, "branch", branchName)
	return nil
}

// CheckBranchExists 检查分支是否存在（本地分支）
func (w *WorkspaceManager) CheckBranchExists(workspacePath, branchName string) (bool, error) {
	if workspacePath == "" {
		return false, fmt.Errorf("工作空间路径不能为空")
	}

	if branchName == "" {
		return false, fmt.Errorf("分支名不能为空")
	}

	// 检查工作空间是否存在
	if !w.CheckWorkspaceExists(workspacePath) {
		return false, fmt.Errorf("工作空间不存在: %s", workspacePath)
	}

	// 检查是否为 Git 仓库
	if !w.CheckGitRepositoryExists(workspacePath) {
		return false, fmt.Errorf("不是Git仓库: %s", workspacePath)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 检查本地分支是否存在
	branchCmd := exec.CommandContext(ctx, "git", "branch", "--list", branchName)
	branchCmd.Dir = workspacePath
	output, err := branchCmd.Output()
	if err != nil {
		return false, fmt.Errorf("检查分支失败: %v", err)
	}

	// 如果输出不为空，说明分支存在
	return len(strings.TrimSpace(string(output))) > 0, nil
}

// GetCurrentBranch 获取当前分支名
func (w *WorkspaceManager) GetCurrentBranch(workspacePath string) (string, error) {
	if workspacePath == "" {
		return "", fmt.Errorf("工作空间路径不能为空")
	}

	// 检查工作空间是否存在
	if !w.CheckWorkspaceExists(workspacePath) {
		return "", fmt.Errorf("工作空间不存在: %s", workspacePath)
	}

	// 检查是否为 Git 仓库
	if !w.CheckGitRepositoryExists(workspacePath) {
		return "", fmt.Errorf("不是Git仓库: %s", workspacePath)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 获取当前分支名
	branchCmd := exec.CommandContext(ctx, "git", "branch", "--show-current")
	branchCmd.Dir = workspacePath
	output, err := branchCmd.Output()
	if err != nil {
		return "", fmt.Errorf("获取当前分支失败: %v", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// PushBranch 推送分支到远程仓库
func (w *WorkspaceManager) PushBranch(workspacePath, branchName, repoURL string, credential *GitCredentialInfo, sslVerify bool) (string, error) {
	if workspacePath == "" {
		return "", fmt.Errorf("工作空间路径不能为空")
	}

	if branchName == "" {
		return "", fmt.Errorf("分支名不能为空")
	}

	// 检查工作空间是否存在
	if !w.CheckWorkspaceExists(workspacePath) {
		return "", fmt.Errorf("工作空间不存在: %s", workspacePath)
	}

	// 检查是否为 Git 仓库
	if !w.CheckGitRepositoryExists(workspacePath) {
		return "", fmt.Errorf("不是Git仓库: %s", workspacePath)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	var cmd *exec.Cmd
	var envVars []string
	output := ""

	// 准备推送命令和认证
	if credential != nil {
		switch credential.Type {
		case GitCredentialTypePassword, GitCredentialTypeToken:
			// HTTPS 认证：配置认证URL
			authenticatedURL, err := w.buildAuthenticatedURL(repoURL, credential)
			if err != nil {
				return "", fmt.Errorf("构建认证URL失败: %v", err)
			}

			// 设置远程仓库URL
			setURLCmd := exec.CommandContext(ctx, "git", "remote", "set-url", "origin", authenticatedURL)
			setURLCmd.Dir = workspacePath

			// 设置Git环境变量
			if !sslVerify {
				setURLCmd.Env = append(os.Environ(), "GIT_SSL_NO_VERIFY=true")
			}

			if err := setURLCmd.Run(); err != nil {
				return "", fmt.Errorf("设置远程仓库URL失败: %v", err)
			}

			// 执行推送
			cmd = exec.CommandContext(ctx, "git", "push", "origin", branchName)
			cmd.Dir = workspacePath

			if !sslVerify {
				cmd.Env = append(os.Environ(), "GIT_SSL_NO_VERIFY=true")
			}

		case GitCredentialTypeSSHKey:
			// SSH 认证
			keyFile := filepath.Join(workspacePath, ".ssh_key_push")
			if err := ioutil.WriteFile(keyFile, []byte(credential.PrivateKey), 0600); err != nil {
				return "", fmt.Errorf("创建SSH密钥文件失败: %v", err)
			}
			defer os.Remove(keyFile)

			envVars = append(os.Environ(),
				fmt.Sprintf("GIT_SSH_COMMAND=ssh -i %s -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no", keyFile),
			)

			cmd = exec.CommandContext(ctx, "git", "push", "origin", branchName)
			cmd.Dir = workspacePath
			cmd.Env = envVars
		}
	} else {
		// 无认证推送
		cmd = exec.CommandContext(ctx, "git", "push", "origin", branchName)
		cmd.Dir = workspacePath
	}

	// 执行推送命令并捕获输出
	var outputBuilder strings.Builder
	cmd.Stdout = &outputBuilder
	cmd.Stderr = &outputBuilder

	err := cmd.Run()
	output = outputBuilder.String()

	if err != nil {
		return output, fmt.Errorf("推送分支失败: %v", err)
	}

	Info("成功推送分支", "workspace", workspacePath, "branch", branchName)
	return output, nil
}
