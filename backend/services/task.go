package services

import (
	"errors"
	"sleep0-backend/database"
	"sleep0-backend/repository"
	"strings"
)

type taskService struct {
	repo        repository.TaskRepository
	projectRepo repository.ProjectRepository
}

// NewTaskService 创建任务服务实例
func NewTaskService(repo repository.TaskRepository, projectRepo repository.ProjectRepository) TaskService {
	return &taskService{
		repo:        repo,
		projectRepo: projectRepo,
	}
}

// CreateTask 创建任务
func (s *taskService) CreateTask(title, description, startBranch, createdBy string, projectID uint) (*database.Task, error) {
	// 验证输入数据
	if err := s.ValidateTaskData(title, startBranch, projectID, createdBy); err != nil {
		return nil, err
	}

	// 检查项目是否存在且属于当前用户
	project, err := s.projectRepo.GetByID(projectID, createdBy)
	if err != nil {
		return nil, errors.New("project not found or access denied")
	}

	// 创建任务
	task := &database.Task{
		Title:       strings.TrimSpace(title),
		Description: strings.TrimSpace(description),
		StartBranch: strings.TrimSpace(startBranch),
		Status:      database.TaskStatusTodo,
		ProjectID:   projectID,
		CreatedBy:   createdBy,
	}

	if err := s.repo.Create(task); err != nil {
		return nil, err
	}

	// 预加载关联数据
	task.Project = project
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

// UpdateTask 更新任务
func (s *taskService) UpdateTask(id uint, createdBy string, updates map[string]interface{}) error {
	// 获取任务
	task, err := s.repo.GetByID(id, createdBy)
	if err != nil {
		return err
	}

	// 验证更新数据
	if title, ok := updates["title"]; ok {
		if titleStr, ok := title.(string); ok && strings.TrimSpace(titleStr) == "" {
			return errors.New("task title cannot be empty")
		}
	}

	if startBranch, ok := updates["start_branch"]; ok {
		if branchStr, ok := startBranch.(string); ok && strings.TrimSpace(branchStr) == "" {
			return errors.New("start branch cannot be empty")
		}
	}

	// 更新字段
	for key, value := range updates {
		switch key {
		case "title":
			if v, ok := value.(string); ok {
				task.Title = strings.TrimSpace(v)
			}
		case "description":
			if v, ok := value.(string); ok {
				task.Description = strings.TrimSpace(v)
			}
		case "start_branch":
			if v, ok := value.(string); ok {
				task.StartBranch = strings.TrimSpace(v)
			}
		}
	}

	return s.repo.Update(task)
}

// DeleteTask 删除任务
func (s *taskService) DeleteTask(id uint, createdBy string) error {
	// 检查任务是否存在
	_, err := s.repo.GetByID(id, createdBy)
	if err != nil {
		return err
	}

	return s.repo.Delete(id, createdBy)
}

// UpdateTaskStatus 更新任务状态
func (s *taskService) UpdateTaskStatus(id uint, createdBy string, status database.TaskStatus) error {
	// 验证状态值
	validStatuses := map[database.TaskStatus]bool{
		database.TaskStatusTodo:       true,
		database.TaskStatusInProgress: true,
		database.TaskStatusDone:       true,
		database.TaskStatusCancelled:  true,
	}

	if !validStatuses[status] {
		return errors.New("invalid task status")
	}

	return s.repo.UpdateStatus(id, createdBy, status)
}

// UpdatePullRequestStatus 更新PR状态
func (s *taskService) UpdatePullRequestStatus(id uint, createdBy string, hasPullRequest bool) error {
	return s.repo.UpdatePullRequestStatus(id, createdBy, hasPullRequest)
}

// GetTaskStats 获取任务统计
func (s *taskService) GetTaskStats(projectID uint, createdBy string) (map[database.TaskStatus]int64, error) {
	return s.repo.CountByStatus(projectID, createdBy)
}

// ListTasksByProject 根据项目获取任务列表
func (s *taskService) ListTasksByProject(projectID uint, createdBy string) ([]database.Task, error) {
	return s.repo.ListByProject(projectID, createdBy)
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
