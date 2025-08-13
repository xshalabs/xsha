package services

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
	"xsha-backend/config"
	"xsha-backend/database"
	appErrors "xsha-backend/errors"
	"xsha-backend/repository"
	"xsha-backend/utils"
)

type taskService struct {
	repo                repository.TaskRepository
	projectRepo         repository.ProjectRepository
	devEnvRepo          repository.DevEnvironmentRepository
	workspaceManager    *utils.WorkspaceManager
	config              *config.Config
	gitCredService      GitCredentialService
	systemConfigService SystemConfigService
}

func NewTaskService(repo repository.TaskRepository, projectRepo repository.ProjectRepository, devEnvRepo repository.DevEnvironmentRepository, workspaceManager *utils.WorkspaceManager, cfg *config.Config, gitCredService GitCredentialService, systemConfigService SystemConfigService) TaskService {
	return &taskService{
		repo:                repo,
		projectRepo:         projectRepo,
		devEnvRepo:          devEnvRepo,
		workspaceManager:    workspaceManager,
		config:              cfg,
		gitCredService:      gitCredService,
		systemConfigService: systemConfigService,
	}
}

func (s *taskService) CreateTask(title, startBranch string, projectID uint, devEnvironmentID *uint, createdBy string) (*database.Task, error) {
	if err := s.ValidateTaskData(title, startBranch, projectID); err != nil {
		return nil, err
	}

	project, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		return nil, appErrors.ErrProjectNotFound
	}

	var devEnv *database.DevEnvironment
	if devEnvironmentID != nil {
		devEnv, err = s.devEnvRepo.GetByID(*devEnvironmentID)
		if err != nil {
			return nil, appErrors.ErrDevEnvironmentNotFound
		}
	}

	workBranch := utils.GenerateWorkBranchName(title, createdBy)

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

	task.Project = project
	task.DevEnvironment = devEnv
	return task, nil
}

func (s *taskService) GetTask(id uint) (*database.Task, error) {
	return s.repo.GetByID(id)
}

func (s *taskService) ListTasks(projectID *uint, statuses []database.TaskStatus, title *string, branch *string, devEnvID *uint, sortBy, sortDirection string, page, pageSize int) ([]database.Task, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	tasks, total, err := s.repo.List(projectID, statuses, title, branch, devEnvID, sortBy, sortDirection, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	if len(tasks) == 0 {
		return tasks, total, nil
	}

	taskIDs := make([]uint, len(tasks))
	for i, task := range tasks {
		taskIDs[i] = task.ID
	}

	conversationCounts, err := s.repo.GetConversationCounts(taskIDs)
	if err != nil {
		utils.Error("Failed to get conversation counts", "error", err)
		return tasks, total, nil
	}

	executionTimes, err := s.repo.GetLatestExecutionTimes(taskIDs)
	if err != nil {
		utils.Error("Failed to get latest execution times", "error", err)
		return tasks, total, nil
	}

	for i := range tasks {
		tasks[i].ConversationCount = conversationCounts[tasks[i].ID]
		tasks[i].LatestExecutionTime = executionTimes[tasks[i].ID]
	}

	return tasks, total, nil
}

func (s *taskService) UpdateTask(id uint, updates map[string]interface{}) error {
	task, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	title, ok := updates["title"]
	if !ok {
		return appErrors.ErrRequired
	}

	titleStr, ok := title.(string)
	if !ok {
		return appErrors.ErrInvalidFormat
	}

	if strings.TrimSpace(titleStr) == "" {
		return appErrors.ErrTaskTitleRequired
	}

	task.Title = strings.TrimSpace(titleStr)

	return s.repo.Update(task)
}

func (s *taskService) UpdateTaskStatus(id uint, status database.TaskStatus) error {
	task, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	oldStatus := task.Status
	task.Status = status

	if err := s.repo.Update(task); err != nil {
		return err
	}

	utils.Info("Task status updated",
		"task_id", id,
		"created_by", "admin",
		"old_status", string(oldStatus),
		"new_status", string(status),
	)
	return nil
}

func (s *taskService) UpdateTaskSessionID(id uint, sessionID string) error {
	task, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	oldSessionID := task.SessionID
	task.SessionID = sessionID

	if err := s.repo.Update(task); err != nil {
		return err
	}

	utils.Info("Task session ID updated",
		"task_id", id,
		"old_session_id", oldSessionID,
		"new_session_id", sessionID,
	)
	return nil
}

func (s *taskService) DeleteTask(id uint) error {
	task, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if task.WorkspacePath != "" {
		if err := s.workspaceManager.CleanupTaskWorkspace(task.WorkspacePath); err != nil {
			utils.Error("Failed to cleanup task workspace",
				"task_id", id,
				"workspace_path", task.WorkspacePath,
				"error", err.Error(),
			)
		} else {
			utils.Info("Task workspace cleaned up",
				"task_id", id,
				"workspace_path", task.WorkspacePath,
			)
		}
	}

	if err := s.repo.Delete(id); err != nil {
		return err
	}

	utils.Info("Task deleted",
		"task_id", id,
		"created_by", "admin",
	)
	return nil
}

func (s *taskService) ValidateTaskData(title, startBranch string, projectID uint) error {
	if strings.TrimSpace(title) == "" {
		return appErrors.ErrTaskTitleRequired
	}

	if len(title) > 200 {
		return appErrors.ErrTaskTitleTooLong
	}

	if strings.TrimSpace(startBranch) == "" {
		return appErrors.ErrStartBranchRequired
	}

	if projectID == 0 {
		return appErrors.ErrProjectIDRequired
	}

	return nil
}

func (s *taskService) UpdateTaskStatusBatch(taskIDs []uint, status database.TaskStatus) ([]uint, []uint, error) {
	if len(taskIDs) == 0 {
		return nil, nil, appErrors.ErrTaskIDsEmpty
	}

	if len(taskIDs) > 100 {
		return nil, nil, appErrors.ErrTooManyTasksForBatch
	}

	var successIDs []uint
	var failedIDs []uint

	for _, taskID := range taskIDs {
		task, err := s.repo.GetByID(taskID)
		if err != nil {
			failedIDs = append(failedIDs, taskID)
			utils.Warn("Failed to get task for batch status update",
				"task_id", taskID,
				"created_by", "admin",
				"error", err.Error(),
			)
			continue
		}

		oldStatus := task.Status
		task.Status = status

		if err := s.repo.Update(task); err != nil {
			failedIDs = append(failedIDs, taskID)
			utils.Error("Failed to update task status in batch",
				"task_id", taskID,
				"created_by", "admin",
				"error", err.Error(),
			)
			continue
		}

		successIDs = append(successIDs, taskID)
		utils.Info("Task status updated in batch",
			"task_id", taskID,
			"created_by", "admin",
			"old_status", string(oldStatus),
			"new_status", string(status),
		)
	}

	return successIDs, failedIDs, nil
}

func (s *taskService) GetTaskGitDiff(task *database.Task, includeContent bool) (*utils.GitDiffSummary, error) {
	if task == nil {
		return nil, fmt.Errorf("task cannot be nil")
	}

	if task.WorkspacePath == "" {
		return nil, fmt.Errorf("task workspace path is empty")
	}

	if task.StartBranch == "" {
		return nil, fmt.Errorf("task start branch is empty")
	}

	if task.WorkBranch == "" {
		return nil, fmt.Errorf("task work branch is empty")
	}

	if err := utils.ValidateBranchExists(task.WorkspacePath, task.StartBranch); err != nil {
		return nil, fmt.Errorf("start branch validation failed: %v", err)
	}

	if err := utils.ValidateBranchExists(task.WorkspacePath, task.WorkBranch); err != nil {
		return nil, fmt.Errorf("work branch validation failed: %v", err)
	}

	diff, err := utils.GetBranchDiff(task.WorkspacePath, task.StartBranch, task.WorkBranch, includeContent)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch diff: %v", err)
	}

	return diff, nil
}

func (s *taskService) GetTaskGitDiffFile(task *database.Task, filePath string) (string, error) {
	if task == nil {
		return "", fmt.Errorf("task cannot be nil")
	}

	if filePath == "" {
		return "", fmt.Errorf("file path cannot be empty")
	}

	if task.WorkspacePath == "" {
		return "", fmt.Errorf("task workspace path is empty")
	}

	if task.StartBranch == "" {
		return "", fmt.Errorf("task start branch is empty")
	}

	if task.WorkBranch == "" {
		return "", fmt.Errorf("task work branch is empty")
	}

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

func (s *taskService) PushTaskBranch(id uint, forcePush bool) (string, error) {
	task, err := s.GetTask(id)
	if err != nil {
		return "", err
	}

	if task.Status == database.TaskStatusCancelled {
		return "", fmt.Errorf("cannot push cancelled task")
	}

	if task.WorkBranch == "" {
		return "", fmt.Errorf("task work branch does not exist")
	}

	if task.WorkspacePath == "" {
		return "", appErrors.ErrWorkspacePathEmpty
	}

	if task.Project == nil {
		return "", appErrors.NewI18nError("task.project_info_incomplete")
	}

	if task.Project.CredentialID == nil {
		return "", appErrors.ErrProjectNotAssociatedWithCredential
	}

	var credential *utils.GitCredentialInfo
	if task.Project.CredentialID != nil {
		cred, err := s.gitCredService.GetCredential(*task.Project.CredentialID)
		if err != nil {
			return "", fmt.Errorf("failed to get Git credential: %v", err)
		}

		credential = &utils.GitCredentialInfo{
			Type:     utils.GitCredentialType(cred.Type),
			Username: cred.Username,
		}

		switch cred.Type {
		case database.GitCredentialTypePassword:
			password, err := s.gitCredService.DecryptCredentialSecret(cred, "password")
			if err != nil {
				return "", fmt.Errorf("failed to decrypt password: %v", err)
			}
			credential.Password = password

		case database.GitCredentialTypeToken:
			token, err := s.gitCredService.DecryptCredentialSecret(cred, "token")
			if err != nil {
				return "", fmt.Errorf("failed to decrypt token: %v", err)
			}
			credential.Password = token

		case database.GitCredentialTypeSSHKey:
			privateKey, err := s.gitCredService.DecryptCredentialSecret(cred, "private_key")
			if err != nil {
				return "", fmt.Errorf("failed to decrypt SSH private key: %v", err)
			}
			credential.PrivateKey = privateKey
			credential.PublicKey = cred.PublicKey
		}
	}

	proxyConfig, err := s.getGitProxyConfig()
	if err != nil {
		utils.Warn("Failed to get proxy config for push, using no proxy", "error", err)
		proxyConfig = nil
	}

	gitSSLVerify, err := s.systemConfigService.GetGitSSLVerify()
	if err != nil {
		utils.Warn("Failed to get git SSL verify setting, using default false", "error", err)
		gitSSLVerify = false
	}

	output, err := s.workspaceManager.PushBranch(
		task.WorkspacePath,
		task.WorkBranch,
		task.Project.RepoURL,
		credential,
		gitSSLVerify,
		proxyConfig,
		forcePush,
	)

	if err != nil {
		utils.Error("Failed to push task branch", "taskID", id, "branch", task.WorkBranch, "error", err)
		return output, err
	}

	utils.Info("Successfully pushed task branch", "taskID", id, "branch", task.WorkBranch)
	return output, nil
}

func (s *taskService) getGitProxyConfig() (*utils.GitProxyConfig, error) {
	return s.systemConfigService.GetGitProxyConfig()
}
