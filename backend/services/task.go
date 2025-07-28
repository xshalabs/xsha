package services

import (
	"errors"
	"strings"
	"xsha-backend/database"
	"xsha-backend/repository"
	"xsha-backend/utils"
)

type taskService struct {
	repo             repository.TaskRepository
	projectRepo      repository.ProjectRepository
	devEnvRepo       repository.DevEnvironmentRepository
	workspaceManager *utils.WorkspaceManager
}

// NewTaskService 创建任务服务实例
func NewTaskService(repo repository.TaskRepository, projectRepo repository.ProjectRepository, devEnvRepo repository.DevEnvironmentRepository, workspaceManager *utils.WorkspaceManager) TaskService {
	return &taskService{
		repo:             repo,
		projectRepo:      projectRepo,
		devEnvRepo:       devEnvRepo,
		workspaceManager: workspaceManager,
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

	// 创建任务
	task := &database.Task{
		Title:            strings.TrimSpace(title),
		StartBranch:      strings.TrimSpace(startBranch),
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

// GetTask 获取任务
func (s *taskService) GetTask(id uint, createdBy string) (*database.Task, error) {
	return s.repo.GetByID(id, createdBy)
}

// ListTasks 获取任务列表
func (s *taskService) ListTasks(projectID *uint, createdBy string, status *database.TaskStatus, page, pageSize int) ([]database.Task, int64, error) {
	// 验证分页参数
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return s.repo.List(projectID, createdBy, status, page, pageSize)
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

// GetTaskStats 获取任务统计
func (s *taskService) GetTaskStats(projectID uint, createdBy string) (map[database.TaskStatus]int64, error) {
	return s.repo.CountByStatus(projectID, createdBy)
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
