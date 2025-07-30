package services

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
	"xsha-backend/config"
	"xsha-backend/database"
	"xsha-backend/repository"
	"xsha-backend/utils"
)

type taskService struct {
	repo             repository.TaskRepository
	projectRepo      repository.ProjectRepository
	devEnvRepo       repository.DevEnvironmentRepository
	workspaceManager *utils.WorkspaceManager
	config           *config.Config
	gitCredService   GitCredentialService
}

// NewTaskService 创建任务服务实例
func NewTaskService(repo repository.TaskRepository, projectRepo repository.ProjectRepository, devEnvRepo repository.DevEnvironmentRepository, workspaceManager *utils.WorkspaceManager, cfg *config.Config, gitCredService GitCredentialService) TaskService {
	return &taskService{
		repo:             repo,
		projectRepo:      projectRepo,
		devEnvRepo:       devEnvRepo,
		workspaceManager: workspaceManager,
		config:           cfg,
		gitCredService:   gitCredService,
	}
}

// CreateTask 创建任务
func (s *taskService) CreateTask(title, startBranch, createdBy string, projectID uint, devEnvironmentID *uint) (*database.Task, error) {
	// 验证输入数据
	if err := s.ValidateTaskData(title, startBranch, projectID, createdBy); err != nil {
		return nil, err
	}

	// 检查项目是否存在且属于当前用户
	project, err := s.projectRepo.GetByID(projectID, createdBy)
	if err != nil {
		return nil, errors.New("project not found or access denied")
	}

	// 如果指定了开发环境，验证其存在性和权限
	var devEnv *database.DevEnvironment
	if devEnvironmentID != nil {
		devEnv, err = s.devEnvRepo.GetByID(*devEnvironmentID, createdBy)
		if err != nil {
			return nil, errors.New("development environment not found or access denied")
		}
	}

	// 生成工作分支名称
	workBranch := s.generateWorkBranchName(title, createdBy)

	// 创建任务
	task := &database.Task{
		Title:            strings.TrimSpace(title),
		StartBranch:      strings.TrimSpace(startBranch),
		WorkBranch:       workBranch,
		Status:           database.TaskStatusTodo,
		ProjectID:        projectID,
		DevEnvironmentID: devEnvironmentID,
		CreatedBy:        createdBy,
	}

	if err := s.repo.Create(task); err != nil {
		return nil, err
	}

	// 预加载关联数据
	task.Project = project
	task.DevEnvironment = devEnv
	return task, nil
}

// generateWorkBranchName 生成工作分支名称
func (s *taskService) generateWorkBranchName(title, createdBy string) string {
	// 清理标题，只保留字母数字和连字符
	cleanTitle := strings.ToLower(strings.TrimSpace(title))

	// 替换空格和特殊字符为连字符
	cleanTitle = strings.ReplaceAll(cleanTitle, " ", "-")
	cleanTitle = strings.ReplaceAll(cleanTitle, "_", "-")

	// 移除非字母数字和连字符的字符
	var result strings.Builder
	for _, r := range cleanTitle {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	cleanTitle = result.String()

	// 限制长度
	if len(cleanTitle) > 30 {
		cleanTitle = cleanTitle[:30]
	}

	// 去掉开头和结尾的连字符
	cleanTitle = strings.Trim(cleanTitle, "-")

	// 如果清理后为空，使用默认前缀
	if cleanTitle == "" {
		cleanTitle = "task"
	}

	// 生成时间戳
	timestamp := time.Now().Format("20060102-150405")

	// 组合分支名: feature/{user}/{clean-title}-{timestamp}
	return fmt.Sprintf("feature/%s/%s-%s", createdBy, cleanTitle, timestamp)
}

// GetTask 获取任务
func (s *taskService) GetTask(id uint, createdBy string) (*database.Task, error) {
	return s.repo.GetByID(id, createdBy)
}

// ListTasks 获取任务列表
func (s *taskService) ListTasks(projectID *uint, createdBy string, status *database.TaskStatus, title *string, branch *string, devEnvID *uint, page, pageSize int) ([]database.Task, int64, error) {
	// 验证分页参数
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return s.repo.List(projectID, createdBy, status, title, branch, devEnvID, page, pageSize)
}

// UpdateTask 更新任务（只允许更新标题）
func (s *taskService) UpdateTask(id uint, createdBy string, updates map[string]interface{}) error {
	// 获取任务
	task, err := s.repo.GetByID(id, createdBy)
	if err != nil {
		return err
	}

	// 只允许更新标题
	title, ok := updates["title"]
	if !ok {
		return errors.New("no updates provided")
	}

	titleStr, ok := title.(string)
	if !ok {
		return errors.New("invalid title format")
	}

	if strings.TrimSpace(titleStr) == "" {
		return errors.New("task title cannot be empty")
	}

	// 更新标题
	task.Title = strings.TrimSpace(titleStr)

	return s.repo.Update(task)
}

// UpdateTaskStatus 更新任务状态
func (s *taskService) UpdateTaskStatus(id uint, createdBy string, status database.TaskStatus) error {
	// 获取任务
	task, err := s.repo.GetByID(id, createdBy)
	if err != nil {
		return err
	}

	oldStatus := task.Status
	task.Status = status

	// 更新任务状态
	if err := s.repo.Update(task); err != nil {
		return err
	}

	utils.Info("Task status updated",
		"task_id", id,
		"created_by", createdBy,
		"old_status", string(oldStatus),
		"new_status", string(status),
	)
	return nil
}

// DeleteTask 删除任务
func (s *taskService) DeleteTask(id uint, createdBy string) error {
	// 检查任务是否存在
	task, err := s.repo.GetByID(id, createdBy)
	if err != nil {
		return err
	}

	// 如果任务有工作空间，先清理工作空间
	if task.WorkspacePath != "" {
		if err := s.workspaceManager.CleanupTaskWorkspace(task.WorkspacePath); err != nil {
			utils.Error("Failed to cleanup task workspace",
				"task_id", id,
				"workspace_path", task.WorkspacePath,
				"error", err.Error(),
			)
			// 不返回错误，避免因清理失败影响任务删除
		} else {
			utils.Info("Task workspace cleaned up",
				"task_id", id,
				"workspace_path", task.WorkspacePath,
			)
		}
	}

	// 删除任务记录
	if err := s.repo.Delete(id, createdBy); err != nil {
		return err
	}

	utils.Info("Task deleted",
		"task_id", id,
		"created_by", createdBy,
	)
	return nil
}

// ValidateTaskData 验证任务数据
func (s *taskService) ValidateTaskData(title, startBranch string, projectID uint, createdBy string) error {
	// 验证标题
	if strings.TrimSpace(title) == "" {
		return errors.New("task title is required")
	}

	if len(title) > 200 {
		return errors.New("task title too long")
	}

	// 验证分支名
	if strings.TrimSpace(startBranch) == "" {
		return errors.New("start branch is required")
	}

	// 验证项目ID
	if projectID == 0 {
		return errors.New("project ID is required")
	}

	return nil
}

// UpdateTaskStatusBatch 批量更新任务状态
func (s *taskService) UpdateTaskStatusBatch(taskIDs []uint, createdBy string, status database.TaskStatus) ([]uint, []uint, error) {
	if len(taskIDs) == 0 {
		return nil, nil, errors.New("task IDs cannot be empty")
	}

	if len(taskIDs) > 100 {
		return nil, nil, errors.New("cannot update more than 100 tasks at once")
	}

	var successIDs []uint
	var failedIDs []uint

	for _, taskID := range taskIDs {
		// 获取任务
		task, err := s.repo.GetByID(taskID, createdBy)
		if err != nil {
			// 任务不存在或没有权限
			failedIDs = append(failedIDs, taskID)
			utils.Warn("Failed to get task for batch status update",
				"task_id", taskID,
				"created_by", createdBy,
				"error", err.Error(),
			)
			continue
		}

		oldStatus := task.Status
		task.Status = status

		// 更新任务状态
		if err := s.repo.Update(task); err != nil {
			failedIDs = append(failedIDs, taskID)
			utils.Error("Failed to update task status in batch",
				"task_id", taskID,
				"created_by", createdBy,
				"error", err.Error(),
			)
			continue
		}

		successIDs = append(successIDs, taskID)
		utils.Info("Task status updated in batch",
			"task_id", taskID,
			"created_by", createdBy,
			"old_status", string(oldStatus),
			"new_status", string(status),
		)
	}

	return successIDs, failedIDs, nil
}

// GetTaskGitDiff 获取任务的Git变动差异
func (s *taskService) GetTaskGitDiff(task *database.Task, includeContent bool) (*utils.GitDiffSummary, error) {
	if task == nil {
		return nil, fmt.Errorf("task cannot be nil")
	}

	// 检查必要的字段
	if task.WorkspacePath == "" {
		return nil, fmt.Errorf("task workspace path is empty")
	}

	if task.StartBranch == "" {
		return nil, fmt.Errorf("task start branch is empty")
	}

	if task.WorkBranch == "" {
		return nil, fmt.Errorf("task work branch is empty")
	}

	// 验证分支是否存在
	if err := utils.ValidateBranchExists(task.WorkspacePath, task.StartBranch); err != nil {
		return nil, fmt.Errorf("start branch validation failed: %v", err)
	}

	if err := utils.ValidateBranchExists(task.WorkspacePath, task.WorkBranch); err != nil {
		return nil, fmt.Errorf("work branch validation failed: %v", err)
	}

	// 获取分支差异
	diff, err := utils.GetBranchDiff(task.WorkspacePath, task.StartBranch, task.WorkBranch, includeContent)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch diff: %v", err)
	}

	return diff, nil
}

// GetTaskGitDiffFile 获取任务指定文件的Git变动详情
func (s *taskService) GetTaskGitDiffFile(task *database.Task, filePath string) (string, error) {
	if task == nil {
		return "", fmt.Errorf("task cannot be nil")
	}

	if filePath == "" {
		return "", fmt.Errorf("file path cannot be empty")
	}

	// 检查必要的字段
	if task.WorkspacePath == "" {
		return "", fmt.Errorf("task workspace path is empty")
	}

	if task.StartBranch == "" {
		return "", fmt.Errorf("task start branch is empty")
	}

	if task.WorkBranch == "" {
		return "", fmt.Errorf("task work branch is empty")
	}

	// 使用utils中的函数获取文件diff内容
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "-c", "core.quotepath=false", "diff", fmt.Sprintf("%s..%s", task.StartBranch, task.WorkBranch), "--", filePath)
	cmd.Dir = task.WorkspacePath

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

// PushTaskBranch 推送任务分支到远程仓库
func (s *taskService) PushTaskBranch(id uint, createdBy string) (string, error) {
	// 获取任务详情
	task, err := s.GetTask(id, createdBy)
	if err != nil {
		return "", err
	}

	// 检查任务状态
	if task.Status == database.TaskStatusCancelled {
		return "", fmt.Errorf("无法推送已取消的任务")
	}

	// 检查必要的字段
	if task.WorkBranch == "" {
		return "", fmt.Errorf("任务工作分支不存在")
	}

	if task.WorkspacePath == "" {
		return "", fmt.Errorf("任务工作空间路径为空")
	}

	if task.Project == nil {
		return "", fmt.Errorf("任务关联项目信息不完整")
	}

	// 检查项目是否关联了 Git Credential
	if task.Project.CredentialID == nil {
		return "", fmt.Errorf("项目未关联Git凭据，请先关联Git Credential后再推送")
	}

	// 准备Git凭据
	var credential *utils.GitCredentialInfo
	if task.Project.CredentialID != nil {
		cred, err := s.gitCredService.GetCredential(*task.Project.CredentialID, createdBy)
		if err != nil {
			return "", fmt.Errorf("获取Git凭据失败: %v", err)
		}

		// 解密凭据信息
		credential = &utils.GitCredentialInfo{
			Type:     utils.GitCredentialType(cred.Type),
			Username: cred.Username,
		}

		switch cred.Type {
		case database.GitCredentialTypePassword:
			password, err := s.gitCredService.DecryptCredentialSecret(cred, "password")
			if err != nil {
				return "", fmt.Errorf("解密密码失败: %v", err)
			}
			credential.Password = password

		case database.GitCredentialTypeToken:
			token, err := s.gitCredService.DecryptCredentialSecret(cred, "token")
			if err != nil {
				return "", fmt.Errorf("解密令牌失败: %v", err)
			}
			credential.Password = token

		case database.GitCredentialTypeSSHKey:
			privateKey, err := s.gitCredService.DecryptCredentialSecret(cred, "private_key")
			if err != nil {
				return "", fmt.Errorf("解密SSH私钥失败: %v", err)
			}
			credential.PrivateKey = privateKey
			credential.PublicKey = cred.PublicKey
		}
	}

	// 执行推送
	output, err := s.workspaceManager.PushBranch(
		task.WorkspacePath,
		task.WorkBranch,
		task.Project.RepoURL,
		credential,
		s.config.GitSSLVerify,
	)

	if err != nil {
		utils.Error("推送任务分支失败", "taskID", id, "branch", task.WorkBranch, "error", err)
		return output, err
	}

	utils.Info("成功推送任务分支", "taskID", id, "branch", task.WorkBranch)
	return output, nil
}
