package container

import (
	"time"
	"xsha-backend/config"
	"xsha-backend/database"
	"xsha-backend/events"
	"xsha-backend/handlers"
	"xsha-backend/repository"
	"xsha-backend/scheduler"
	"xsha-backend/services"
	"xsha-backend/services/executor"
	"xsha-backend/utils"
)

type Container struct {
	// Core dependencies
	cfg      *config.Config
	db       *database.DatabaseManager
	eventBus events.EventBus

	// Repositories
	TokenRepo                  repository.TokenBlacklistRepository
	LoginLogRepo               repository.LoginLogRepository
	AdminRepo                  repository.AdminRepository
	AdminAvatarRepo            repository.AdminAvatarRepository
	AdminOperationLogRepo      repository.AdminOperationLogRepository
	GitCredRepo                repository.GitCredentialRepository
	ProjectRepo                repository.ProjectRepository
	DevEnvRepo                 repository.DevEnvironmentRepository
	TaskRepo                   repository.TaskRepository
	TaskConvRepo               repository.TaskConversationRepository
	ExecLogRepo                repository.TaskExecutionLogRepository
	TaskConvResultRepo         repository.TaskConversationResultRepository
	TaskConvAttachmentRepo     repository.TaskConversationAttachmentRepository
	SystemConfigRepo           repository.SystemConfigRepository
	DashboardRepo              repository.DashboardRepository

	// Services
	LoginLogService            services.LoginLogService
	AdminOperationLogService   services.AdminOperationLogService
	AdminService               services.AdminService
	AdminAvatarService         services.AdminAvatarService
	AuthService                services.AuthService
	GitCredService             services.GitCredentialService
	SystemConfigService        services.SystemConfigService
	DashboardService           services.DashboardService
	DevEnvService              services.DevEnvironmentService
	ProjectService             services.ProjectService
	TaskService                services.TaskService
	TaskConvResultService      services.TaskConversationResultService
	TaskConvAttachmentService  services.TaskConversationAttachmentService
	TaskConvService            services.TaskConversationService

	// Execution components
	WorkspaceManager           *utils.WorkspaceManager
	ExecutionManager           *executor.ExecutionManager
	AITaskExecutor             services.AITaskExecutorService
	LogStreamingService        executor.LogStreamingService

	// Scheduler
	TaskProcessor              scheduler.TaskProcessor

	// Handlers
	AuthHandlers               *handlers.AuthHandlers
	AdminHandlers              *handlers.AdminHandlers
	AdminAvatarHandlers        *handlers.AdminAvatarHandlers
	AdminOperationLogHandlers  *handlers.AdminOperationLogHandlers
	GitCredHandlers            *handlers.GitCredentialHandlers
	ProjectHandlers            *handlers.ProjectHandlers
	DevEnvHandlers             *handlers.DevEnvironmentHandlers
	TaskHandlers               *handlers.TaskHandlers
	TaskConvHandlers           *handlers.TaskConversationHandlers
	TaskConvAttachmentHandlers *handlers.TaskConversationAttachmentHandlers
	SystemConfigHandlers       *handlers.SystemConfigHandlers
	DashboardHandlers          *handlers.DashboardHandlers
}

func New(dbManager *database.DatabaseManager, eventBus events.EventBus, cfg *config.Config) (*Container, error) {
	c := &Container{
		cfg:      cfg,
		db:       dbManager,
		eventBus: eventBus,
	}

	if err := c.initializeRepositories(); err != nil {
		return nil, err
	}

	if err := c.initializeServices(); err != nil {
		return nil, err
	}

	if err := c.initializeExecutors(); err != nil {
		return nil, err
	}

	if err := c.initializeScheduler(); err != nil {
		return nil, err
	}

	if err := c.initializeHandlers(); err != nil {
		return nil, err
	}

	if err := c.wireCircularDependencies(); err != nil {
		return nil, err
	}

	if err := c.registerEventHandlers(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Container) initializeRepositories() error {
	db := c.db.GetDB()
	
	c.TokenRepo = repository.NewTokenBlacklistRepository(db)
	c.LoginLogRepo = repository.NewLoginLogRepository(db)
	c.AdminRepo = repository.NewAdminRepository(db)
	c.AdminAvatarRepo = repository.NewAdminAvatarRepository(db)
	c.AdminOperationLogRepo = repository.NewAdminOperationLogRepository(db)
	c.GitCredRepo = repository.NewGitCredentialRepository(db)
	c.ProjectRepo = repository.NewProjectRepository(db)
	c.DevEnvRepo = repository.NewDevEnvironmentRepository(db)
	c.TaskRepo = repository.NewTaskRepository(db)
	c.TaskConvRepo = repository.NewTaskConversationRepository(db)
	c.ExecLogRepo = repository.NewTaskExecutionLogRepository(db)
	c.TaskConvResultRepo = repository.NewTaskConversationResultRepository(db)
	c.TaskConvAttachmentRepo = repository.NewTaskConversationAttachmentRepository(db)
	c.SystemConfigRepo = repository.NewSystemConfigRepository(db)
	c.DashboardRepo = repository.NewDashboardRepository(db)

	return nil
}

func (c *Container) initializeServices() error {
	c.LoginLogService = services.NewLoginLogService(c.LoginLogRepo)
	c.AdminOperationLogService = services.NewAdminOperationLogService(c.AdminOperationLogRepo)
	c.AdminService = services.NewAdminService(c.AdminRepo, c.eventBus)
	c.AdminAvatarService = services.NewAdminAvatarService(c.AdminAvatarRepo, c.AdminRepo, c.cfg)
	c.AuthService = services.NewAuthService(c.TokenRepo, c.LoginLogRepo, c.AdminOperationLogService, c.AdminService, c.AdminRepo, c.cfg)
	c.GitCredService = services.NewGitCredentialService(c.GitCredRepo, c.ProjectRepo, c.cfg)
	c.SystemConfigService = services.NewSystemConfigService(c.SystemConfigRepo)
	c.DashboardService = services.NewDashboardService(c.DashboardRepo)
	c.DevEnvService = services.NewDevEnvironmentService(c.DevEnvRepo, c.TaskRepo, c.SystemConfigService, c.cfg)

	gitCloneTimeout, err := c.SystemConfigService.GetGitCloneTimeout()
	if err != nil {
		utils.Error("Failed to get git clone timeout from system config, using default", "error", err)
		gitCloneTimeout = 5 * time.Minute
	}

	c.WorkspaceManager = utils.NewWorkspaceManager(c.cfg.WorkspaceBaseDir, gitCloneTimeout)
	c.ProjectService = services.NewProjectService(c.ProjectRepo, c.GitCredRepo, c.GitCredService, c.TaskRepo, c.SystemConfigService, c.cfg)
	c.TaskService = services.NewTaskService(c.TaskRepo, c.ProjectRepo, c.DevEnvRepo, c.TaskConvRepo, c.ExecLogRepo, c.TaskConvResultRepo, c.TaskConvAttachmentRepo, c.WorkspaceManager, c.cfg, c.GitCredService, c.SystemConfigService, c.eventBus)
	c.TaskConvResultService = services.NewTaskConversationResultService(c.TaskConvResultRepo, c.TaskConvRepo, c.TaskRepo, c.ProjectRepo)
	c.TaskConvAttachmentService = services.NewTaskConversationAttachmentService(c.TaskConvAttachmentRepo, c.cfg)
	c.TaskConvService = services.NewTaskConversationService(c.TaskConvRepo, c.TaskRepo, c.ExecLogRepo, c.TaskConvResultRepo, c.TaskService, c.TaskConvAttachmentService, c.WorkspaceManager)

	return nil
}

func (c *Container) initializeExecutors() error {
	maxConcurrency := 5
	if c.cfg.MaxConcurrentTasks > 0 {
		maxConcurrency = c.cfg.MaxConcurrentTasks
	}
	
	c.ExecutionManager = executor.NewExecutionManager(maxConcurrency)
	c.AITaskExecutor = executor.NewAITaskExecutorServiceWithManager(c.TaskConvRepo, c.TaskRepo, c.ExecLogRepo, c.TaskConvResultRepo, c.GitCredService, c.TaskConvResultService, c.TaskService, c.SystemConfigService, c.TaskConvAttachmentService, c.cfg, c.ExecutionManager)
	c.LogStreamingService = executor.NewLogStreamingService(c.TaskConvRepo, c.ExecLogRepo, c.ExecutionManager)

	return nil
}

func (c *Container) initializeScheduler() error {
	c.TaskProcessor = scheduler.NewTaskProcessor(c.AITaskExecutor)
	return nil
}

func (c *Container) initializeHandlers() error {
	c.AuthHandlers = handlers.NewAuthHandlers(c.AuthService, c.LoginLogService, c.AdminService, c.AdminAvatarService)
	c.AdminHandlers = handlers.NewAdminHandlers(c.AdminService)
	c.AdminAvatarHandlers = handlers.NewAdminAvatarHandlers(c.AdminAvatarService, c.AdminService)
	c.AdminOperationLogHandlers = handlers.NewAdminOperationLogHandlers(c.AdminOperationLogService)
	c.GitCredHandlers = handlers.NewGitCredentialHandlers(c.GitCredService)
	c.ProjectHandlers = handlers.NewProjectHandlers(c.ProjectService)
	c.DevEnvHandlers = handlers.NewDevEnvironmentHandlers(c.DevEnvService)
	c.TaskHandlers = handlers.NewTaskHandlers(c.TaskService, c.TaskConvService, c.ProjectService)
	c.TaskConvHandlers = handlers.NewTaskConversationHandlers(c.TaskConvService, c.LogStreamingService, c.AITaskExecutor)
	c.TaskConvAttachmentHandlers = handlers.NewTaskConversationAttachmentHandlers(c.TaskConvAttachmentService)
	c.SystemConfigHandlers = handlers.NewSystemConfigHandlers(c.SystemConfigService)
	c.DashboardHandlers = handlers.NewDashboardHandlers(c.DashboardService)

	return nil
}

func (c *Container) wireCircularDependencies() error {
	c.AdminService.SetAuthService(c.AuthService)
	c.AdminService.SetDevEnvironmentService(c.DevEnvService)
	c.AdminService.SetGitCredentialService(c.GitCredService)
	c.AdminService.SetProjectService(c.ProjectService)
	c.AdminService.SetTaskService(c.TaskService)
	c.AdminService.SetTaskConversationService(c.TaskConvService)

	return nil
}

func (c *Container) registerEventHandlers() error {
	// Import event handlers here to avoid circular imports
	eventHandlers := c.createEventHandlers()

	c.eventBus.Subscribe(events.EventTypeTaskCreated, eventHandlers.Task.HandleTaskCreated)
	c.eventBus.Subscribe(events.EventTypeTaskStatusChanged, eventHandlers.Task.HandleTaskStatusChanged)
	c.eventBus.Subscribe(events.EventTypeTaskCompleted, eventHandlers.Task.HandleTaskCompleted)
	c.eventBus.Subscribe(events.EventTypeTaskFailed, eventHandlers.Task.HandleTaskFailed)
	c.eventBus.Subscribe(events.EventTypeTaskDeleted, eventHandlers.Task.HandleTaskDeleted)
	c.eventBus.Subscribe(events.EventTypeTaskWorkspaceReady, eventHandlers.Task.HandleTaskWorkspaceReady)

	c.eventBus.Subscribe(events.EventTypeAdminCreated, eventHandlers.Admin.HandleAdminCreated)
	c.eventBus.Subscribe(events.EventTypeAdminUpdated, eventHandlers.Admin.HandleAdminUpdated)
	c.eventBus.Subscribe(events.EventTypeAdminDeleted, eventHandlers.Admin.HandleAdminDeleted)
	c.eventBus.Subscribe(events.EventTypeAdminRoleChanged, eventHandlers.Admin.HandleAdminRoleChanged)
	c.eventBus.Subscribe(events.EventTypeAdminLogin, eventHandlers.Admin.HandleAdminLogin)
	c.eventBus.Subscribe(events.EventTypeAdminLogout, eventHandlers.Admin.HandleAdminLogout)
	c.eventBus.Subscribe(events.EventTypePermissionGranted, eventHandlers.Admin.HandlePermissionGranted)
	c.eventBus.Subscribe(events.EventTypePermissionRevoked, eventHandlers.Admin.HandlePermissionRevoked)

	return nil
}