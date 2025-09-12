package app

import (
	"context"
	"embed"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"xsha-backend/config"
	"xsha-backend/database"
	"xsha-backend/events"
	"xsha-backend/internal/container"
	"xsha-backend/routes"
	"xsha-backend/scheduler"
	"xsha-backend/utils"

	"github.com/gin-gonic/gin"
)

type App struct {
	cfg              *config.Config
	dbManager        *database.DatabaseManager
	eventBus         events.EventBus
	schedulerManager scheduler.Scheduler
	container        *container.Container
	server           *http.Server
	StaticFiles      *embed.FS
}

func New(cfg *config.Config, staticFiles *embed.FS) *App {
	return &App{
		cfg:         cfg,
		StaticFiles: staticFiles,
	}
}

func (a *App) Initialize() error {
	if err := a.initializeDatabase(); err != nil {
		return err
	}

	if err := a.initializeEventSystem(); err != nil {
		return err
	}

	if err := a.initializeContainer(); err != nil {
		return err
	}

	if err := a.initializeDirectories(); err != nil {
		return err
	}

	if err := a.initializeServices(); err != nil {
		return err
	}

	if err := a.initializeScheduler(); err != nil {
		return err
	}

	if err := a.initializeServer(); err != nil {
		return err
	}

	return nil
}

func (a *App) Run() error {
	if err := a.schedulerManager.Start(); err != nil {
		return err
	}

	utils.Info("Server starting on port", "port", a.cfg.Port)
	
	if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	utils.Info("Shutting down gracefully...")

	if a.schedulerManager != nil {
		if err := a.schedulerManager.Stop(); err != nil {
			utils.Error("Failed to stop scheduler", "error", err)
		}
	}

	if a.eventBus != nil {
		if err := a.eventBus.Stop(ctx); err != nil {
			utils.Error("Failed to stop event bus", "error", err)
		}
	}

	if a.server != nil {
		if err := a.server.Shutdown(ctx); err != nil {
			utils.Error("Failed to shutdown HTTP server", "error", err)
		}
	}

	if a.dbManager != nil {
		a.dbManager.Close()
	}

	if err := utils.Sync(); err != nil {
		utils.Error("Failed to sync logger", "error", err)
	}

	return nil
}

func (a *App) WaitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := a.Shutdown(shutdownCtx); err != nil {
		utils.Error("Error during shutdown", "error", err)
		os.Exit(1)
	}
}

func (a *App) initializeDatabase() error {
	dbManager, err := database.NewDatabaseManager(a.cfg)
	if err != nil {
		return err
	}
	a.dbManager = dbManager
	return nil
}

func (a *App) initializeEventSystem() error {
	eventBusConfig := events.DefaultEventBusConfig()
	eventStore := events.NewDBEventStore(a.dbManager.GetDB())
	deadLetterQueue := events.NewDBDeadLetterQueue(a.dbManager.GetDB())
	
	eventBus := events.NewEventBus(eventBusConfig,
		events.WithEventStore(eventStore),
		events.WithDeadLetterQueue(deadLetterQueue),
	)

	if err := eventBus.Start(context.Background()); err != nil {
		return err
	}

	a.eventBus = eventBus
	return nil
}

func (a *App) initializeContainer() error {
	c, err := container.New(a.dbManager, a.eventBus, a.cfg)
	if err != nil {
		return err
	}
	a.container = c
	return nil
}

func (a *App) initializeDirectories() error {
	directories := []string{
		a.cfg.AttachmentsDir,
		a.cfg.AvatarsDir,
		a.cfg.WorkspaceBaseDir,
		a.cfg.DevSessionsDir,
	}

	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

func (a *App) initializeServices() error {
	if err := a.container.SystemConfigService.InitializeDefaultConfigs(); err != nil {
		return err
	}

	if err := a.container.AdminService.InitializeDefaultAdmin(); err != nil {
		return err
	}

	return nil
}

func (a *App) initializeScheduler() error {
	schedulerManager := scheduler.NewSchedulerManager(
		a.container.TaskProcessor, 
		a.cfg.SchedulerIntervalDuration,
	)
	a.schedulerManager = schedulerManager
	return nil
}

func (a *App) initializeServer() error {
	if a.cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	routes.SetupRoutes(r, a.cfg, a.container.AuthService, a.container.AdminService,
		a.container.AuthHandlers, a.container.AdminHandlers, a.container.AdminAvatarHandlers,
		a.container.GitCredHandlers, a.container.ProjectHandlers, a.container.AdminOperationLogHandlers,
		a.container.DevEnvHandlers, a.container.TaskHandlers, a.container.TaskConvHandlers,
		a.container.TaskConvAttachmentHandlers, a.container.SystemConfigHandlers,
		a.container.DashboardHandlers, a.StaticFiles)

	a.server = &http.Server{
		Addr:    ":" + a.cfg.Port,
		Handler: r,
	}

	return nil
}