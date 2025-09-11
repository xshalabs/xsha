package services

import (
	"context"
	"fmt"
	"os"
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
	repo                           repository.TaskRepository
	projectRepo                    repository.ProjectRepository
	devEnvRepo                     repository.DevEnvironmentRepository
	taskConversationRepo           repository.TaskConversationRepository
	taskExecutionLogRepo           repository.TaskExecutionLogRepository
	taskConversationResultRepo     repository.TaskConversationResultRepository
	taskConversationAttachmentRepo repository.TaskConversationAttachmentRepository
	workspaceManager               *utils.WorkspaceManager
	config                         *config.Config
	gitCredService                 GitCredentialService
	systemConfigService            SystemConfigService
}

func NewTaskService(repo repository.TaskRepository, projectRepo repository.ProjectRepository, devEnvRepo repository.DevEnvironmentRepository, taskConversationRepo repository.TaskConversationRepository, taskExecutionLogRepo repository.TaskExecutionLogRepository, taskConversationResultRepo repository.TaskConversationResultRepository, taskConversationAttachmentRepo repository.TaskConversationAttachmentRepository, workspaceManager *utils.WorkspaceManager, cfg *config.Config, gitCredService GitCredentialService, systemConfigService SystemConfigService) TaskService {
	return &taskService{
		repo:                           repo,
		projectRepo:                    projectRepo,
		devEnvRepo:                     devEnvRepo,
		taskConversationRepo:           taskConversationRepo,
		taskExecutionLogRepo:           taskExecutionLogRepo,
		taskConversationResultRepo:     taskConversationResultRepo,
		taskConversationAttachmentRepo: taskConversationAttachmentRepo,
		workspaceManager:               workspaceManager,
		config:                         cfg,
		gitCredService:                 gitCredService,
		systemConfigService:            systemConfigService,
	}
}

func (s *taskService) CreateTask(title, startBranch string, projectID uint, devEnvironmentID *uint, adminID *uint, createdBy string) (*database.Task, error) {
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
		AdminID:          adminID,
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

func (s *taskService) GetTaskByIDAndProject(taskID, projectID uint) (*database.Task, error) {
	return s.repo.GetByIDAndProjectID(taskID, projectID)
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

	// First, delete all conversations and their related data
	conversations, err := s.taskConversationRepo.ListByTask(id)
	if err != nil {
		utils.Error("Failed to get task conversations for deletion",
			"task_id", id,
			"error", err.Error(),
		)
		// Continue with task deletion even if we can't get conversations
	} else {
		// Delete each conversation and its related data
		for _, conv := range conversations {
			if err := s.deleteConversationCascade(conv.ID, task.WorkspacePath); err != nil {
				utils.Error("Failed to delete task conversation during task deletion",
					"task_id", id,
					"conversation_id", conv.ID,
					"error", err.Error(),
				)
				// Continue deleting other conversations even if one fails
			} else {
				utils.Info("Task conversation deleted during task deletion",
					"task_id", id,
					"conversation_id", conv.ID,
				)
			}
		}
	}

	// Clean up workspace
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

	// Finally, delete the task record
	if err := s.repo.Delete(id); err != nil {
		return err
	}

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

func (s *taskService) UpdateTaskStatusBatch(taskIDs []uint, status database.TaskStatus, projectID uint) ([]uint, []uint, error) {
	if len(taskIDs) == 0 {
		return nil, nil, appErrors.ErrTaskIDsEmpty
	}

	if len(taskIDs) > 100 {
		return nil, nil, appErrors.ErrTooManyTasksForBatch
	}

	// Use repository batch update with project ID filter
	successIDs, failedIDs, err := s.repo.UpdateStatusBatch(taskIDs, status, projectID)
	if err != nil {
		return nil, taskIDs, err
	}

	// Log successful updates
	for _, taskID := range successIDs {
		utils.Info("Task status updated in batch",
			"task_id", taskID,
			"project_id", projectID,
			"created_by", "admin",
			"new_status", string(status),
		)
	}

	// Log failed updates
	for _, taskID := range failedIDs {
		utils.Warn("Failed to update task status in batch (task not found in project or does not exist)",
			"task_id", taskID,
			"project_id", projectID,
			"created_by", "admin",
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

	// Convert relative workspace path to absolute for git operations
	absoluteWorkspacePath := s.workspaceManager.GetAbsolutePath(task.WorkspacePath)

	if err := utils.ValidateBranchExists(absoluteWorkspacePath, task.StartBranch); err != nil {
		return nil, fmt.Errorf("start branch validation failed: %v", err)
	}

	if err := utils.ValidateBranchExists(absoluteWorkspacePath, task.WorkBranch); err != nil {
		return nil, fmt.Errorf("work branch validation failed: %v", err)
	}

	diff, err := utils.GetBranchDiff(absoluteWorkspacePath, task.StartBranch, task.WorkBranch, includeContent)
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

	// Convert relative workspace path to absolute for git operations
	absoluteWorkspacePath := s.workspaceManager.GetAbsolutePath(task.WorkspacePath)

	cmd := exec.CommandContext(ctx, "git", "-c", "core.quotepath=false", "diff", fmt.Sprintf("%s..%s", task.StartBranch, task.WorkBranch), "--", filePath)
	cmd.Dir = absoluteWorkspacePath

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

	return output, nil
}

func (s *taskService) GetKanbanTasks(projectID uint) (map[database.TaskStatus][]database.Task, error) {
	// Validate project exists
	_, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		return nil, appErrors.ErrProjectNotFound
	}

	// Get all tasks for the project
	tasks, err := s.repo.ListByProject(projectID)
	if err != nil {
		return nil, err
	}

	// Get additional task metadata
	if len(tasks) > 0 {
		taskIDs := make([]uint, len(tasks))
		for i, task := range tasks {
			taskIDs[i] = task.ID
		}

		// Get conversation counts
		conversationCounts, err := s.repo.GetConversationCounts(taskIDs)
		if err != nil {
			utils.Error("Failed to get conversation counts", "error", err)
		} else {
			for i := range tasks {
				if count, exists := conversationCounts[tasks[i].ID]; exists {
					tasks[i].ConversationCount = count
				}
			}
		}

		// Get latest execution times
		latestTimes, err := s.repo.GetLatestExecutionTimes(taskIDs)
		if err != nil {
			utils.Error("Failed to get latest execution times", "error", err)
		} else {
			for i := range tasks {
				if execTime, exists := latestTimes[tasks[i].ID]; exists {
					tasks[i].LatestExecutionTime = execTime
				}
			}
		}
	}

	// Group tasks by status
	kanbanData := make(map[database.TaskStatus][]database.Task)

	// Initialize all status groups to ensure they exist even if empty
	kanbanData[database.TaskStatusTodo] = []database.Task{}
	kanbanData[database.TaskStatusInProgress] = []database.Task{}
	kanbanData[database.TaskStatusDone] = []database.Task{}
	kanbanData[database.TaskStatusCancelled] = []database.Task{}

	// Group tasks by their status
	for _, task := range tasks {
		kanbanData[task.Status] = append(kanbanData[task.Status], task)
	}

	return kanbanData, nil
}

// deleteConversationCascade deletes a conversation and all its related data
// This method is used internally by DeleteTask to avoid circular dependencies
func (s *taskService) deleteConversationCascade(conversationID uint, workspacePath string) error {
	// Get conversation details
	conversation, err := s.taskConversationRepo.GetByID(conversationID)
	if err != nil {
		return err
	}

	// Handle git repository reset if needed
	if conversation.CommitHash != "" && workspacePath != "" {
		// Convert relative workspace path to absolute for git operations
		absoluteWorkspacePath := s.workspaceManager.GetAbsolutePath(workspacePath)
		if err := utils.GitResetToPreviousCommit(absoluteWorkspacePath, conversation.CommitHash); err != nil {
			utils.Error("Failed to reset git repository to previous commit",
				"conversation_id", conversationID,
				"commit_hash", conversation.CommitHash,
				"workspace_path", workspacePath,
				"error", err)
			// Don't fail the entire operation for git reset errors
		} else {
			utils.Info("Successfully reset git repository to previous commit",
				"conversation_id", conversationID,
				"commit_hash", conversation.CommitHash,
				"workspace_path", workspacePath)
		}
	}

	// Delete execution logs
	if err := s.taskExecutionLogRepo.DeleteByConversationID(conversationID); err != nil {
		utils.Error("Failed to delete execution logs for conversation",
			"conversation_id", conversationID,
			"error", err)
		// Continue with deletion even if this fails
	}

	// Delete conversation results
	if err := s.taskConversationResultRepo.DeleteByConversationID(conversationID); err != nil {
		utils.Warn("Failed to delete conversation result",
			"conversation_id", conversationID,
			"error", err)
		// Continue with deletion even if this fails
	}

	// Delete attachments (both records and physical files)
	attachments, err := s.taskConversationAttachmentRepo.GetByConversationID(conversationID)
	if err != nil {
		utils.Warn("Failed to get attachments for conversation deletion",
			"conversation_id", conversationID,
			"error", err)
	} else {
		// Delete physical files first
		for _, attachment := range attachments {
			if err := os.Remove(attachment.FilePath); err != nil {
				// Log error but continue with other files
				utils.Warn("Failed to delete physical attachment file",
					"conversation_id", conversationID,
					"attachment_id", attachment.ID,
					"file_path", attachment.FilePath,
					"error", err)
			}
		}

		// Delete attachment records
		if err := s.taskConversationAttachmentRepo.DeleteByConversationID(conversationID); err != nil {
			utils.Warn("Failed to delete attachment records for conversation",
				"conversation_id", conversationID,
				"error", err)
		}
	}

	// Finally, delete the conversation record
	if err := s.taskConversationRepo.Delete(conversationID); err != nil {
		return err
	}

	return nil
}

func (s *taskService) getGitProxyConfig() (*utils.GitProxyConfig, error) {
	return s.systemConfigService.GetGitProxyConfig()
}

// CountByAdminID counts the number of tasks created by a specific admin
func (s *taskService) CountByAdminID(adminID uint) (int64, error) {
	return s.repo.CountByAdminID(adminID)
}
