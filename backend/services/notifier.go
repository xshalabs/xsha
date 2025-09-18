package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"xsha-backend/database"
	appErrors "xsha-backend/errors"
	"xsha-backend/repository"
	"xsha-backend/services/notifiers"
	"xsha-backend/utils"
)

type notifierService struct {
	repo        repository.NotifierRepository
	projectRepo repository.ProjectRepository
}

func NewNotifierService(repo repository.NotifierRepository, projectRepo repository.ProjectRepository) NotifierService {
	return &notifierService{
		repo:        repo,
		projectRepo: projectRepo,
	}
}

// CRUD operations with permission checks

func (s *notifierService) CreateNotifier(name, description string, notifierType database.NotifierType, config map[string]interface{}, admin *database.Admin) (*database.Notifier, error) {
	// Check if admin has permission to create notifiers
	if admin.Role != database.AdminRoleSuperAdmin && admin.Role != database.AdminRoleAdmin {
		return nil, appErrors.NewI18nError("permission_denied", "en-US")
	}

	// Validate input
	if err := s.validateNotifierData(name, notifierType, config); err != nil {
		return nil, err
	}

	// Check if name already exists
	if existing, _ := s.repo.GetByName(name); existing != nil {
		return nil, appErrors.NewI18nError("notifier_name_exists", "en-US")
	}

	// Validate provider configuration
	provider, err := notifiers.NewProvider(notifierType, config)
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %v", err)
	}

	if err := provider.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %v", err)
	}

	// Convert config to JSON string
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %v", err)
	}

	notifier := &database.Notifier{
		Name:        name,
		Description: description,
		Type:        notifierType,
		Config:      string(configJSON),
		IsEnabled:   true,
		AdminID:     &admin.ID,
		CreatedBy:   admin.Username,
	}

	if err := s.repo.Create(notifier); err != nil {
		return nil, err
	}

	return notifier, nil
}

func (s *notifierService) GetNotifier(id uint, admin *database.Admin) (*database.Notifier, error) {
	notifier, err := s.repo.GetByIDWithAdmin(id)
	if err != nil {
		return nil, err
	}

	// Check permission
	canAccess, err := s.CanAdminAccessNotifier(id, admin)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, appErrors.NewI18nError("permission_denied", "en-US")
	}

	return notifier, nil
}

func (s *notifierService) ListNotifiers(admin *database.Admin, name *string, notifierTypes []database.NotifierType, isEnabled *bool, page, pageSize int) ([]database.NotifierListItemResponse, int64, error) {
	// Check if admin has permission to view notifiers
	if admin.Role != database.AdminRoleSuperAdmin && admin.Role != database.AdminRoleAdmin {
		return []database.NotifierListItemResponse{}, 0, nil
	}

	notifiers, total, err := s.repo.ListByAdminAccess(admin.ID, admin.Role, name, notifierTypes, isEnabled, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	responses := database.ToNotifierListItemResponses(notifiers)
	return responses, total, nil
}

func (s *notifierService) UpdateNotifier(id uint, updates map[string]interface{}, admin *database.Admin) error {
	// Check permission
	canAccess, err := s.CanAdminAccessNotifier(id, admin)
	if err != nil {
		return err
	}
	if !canAccess {
		return appErrors.NewI18nError("permission_denied", "en-US")
	}

	notifier, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// Reject type field updates - type cannot be changed after creation
	if _, ok := updates["type"]; ok {
		return appErrors.NewI18nError("notifier_type_cannot_be_modified", "en-US")
	}

	// Update fields
	if name, ok := updates["name"]; ok {
		notifier.Name = name.(string)
	}
	if description, ok := updates["description"]; ok {
		notifier.Description = description.(string)
	}
	if isEnabled, ok := updates["is_enabled"]; ok {
		notifier.IsEnabled = isEnabled.(bool)
	}
	if config, ok := updates["config"]; ok {
		// Validate configuration if updated
		configMap := config.(map[string]interface{})
		provider, err := notifiers.NewProvider(notifier.Type, configMap)
		if err != nil {
			return fmt.Errorf("invalid configuration: %v", err)
		}

		if err := provider.ValidateConfig(configMap); err != nil {
			return fmt.Errorf("configuration validation failed: %v", err)
		}

		configJSON, err := json.Marshal(configMap)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %v", err)
		}
		notifier.Config = string(configJSON)
	}

	return s.repo.Update(notifier)
}

func (s *notifierService) DeleteNotifier(id uint, admin *database.Admin) error {
	// Check permission
	canAccess, err := s.CanAdminAccessNotifier(id, admin)
	if err != nil {
		return err
	}
	if !canAccess {
		return appErrors.NewI18nError("permission_denied", "en-US")
	}

	return s.repo.Delete(id)
}

func (s *notifierService) TestNotifier(id uint, admin *database.Admin) error {
	// Check permission
	canAccess, err := s.CanAdminAccessNotifier(id, admin)
	if err != nil {
		return err
	}
	if !canAccess {
		return appErrors.NewI18nError("permission_denied", "en-US")
	}

	notifier, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// Parse config
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(notifier.Config), &config); err != nil {
		return fmt.Errorf("failed to parse config: %v", err)
	}

	// Create provider and test
	provider, err := notifiers.NewProvider(notifier.Type, config)
	if err != nil {
		return fmt.Errorf("failed to create provider: %v", err)
	}

	return provider.Test(admin.Lang)
}

// Project association methods

func (s *notifierService) AddNotifierToProject(projectID, notifierID uint, admin *database.Admin) error {
	// Check if admin can access the notifier
	canAccessNotifier, err := s.CanAdminAccessNotifier(notifierID, admin)
	if err != nil {
		return err
	}
	if !canAccessNotifier {
		return appErrors.NewI18nError("permission_denied", "en-US")
	}

	// Check if admin can access the project (use existing project service logic)
	project, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		return err
	}

	// For now, allow any admin with notifier access to associate with projects
	// This can be refined later based on project permissions
	_ = project

	return s.repo.AddProject(notifierID, projectID)
}

func (s *notifierService) RemoveNotifierFromProject(projectID, notifierID uint, admin *database.Admin) error {
	// Check if admin can access the notifier
	canAccessNotifier, err := s.CanAdminAccessNotifier(notifierID, admin)
	if err != nil {
		return err
	}
	if !canAccessNotifier {
		return appErrors.NewI18nError("permission_denied", "en-US")
	}

	return s.repo.RemoveProject(notifierID, projectID)
}

func (s *notifierService) GetProjectNotifiers(projectID uint) ([]database.NotifierListItemResponse, error) {
	notifiers, err := s.repo.GetProjectNotifiers(projectID)
	if err != nil {
		return nil, err
	}

	responses := database.ToNotifierListItemResponses(notifiers)
	return responses, nil
}

// Notification sending

func (s *notifierService) SendNotificationForTask(task *database.Task, conversation *database.TaskConversation, status database.ConversationStatus, completionTime time.Time, errorMsg string, adminLang string) error {
	// Get all enabled notifiers associated with the project
	notifiers, err := s.repo.GetEnabledProjectNotifiers(task.ProjectID)
	if err != nil {
		utils.Error("Failed to get enabled project notifiers", "projectID", task.ProjectID, "error", err)
		return err
	}

	if len(notifiers) == 0 {
		utils.Debug("No notifiers configured for project", "projectID", task.ProjectID)
		return nil
	}

	// Send notifications concurrently
	errorChan := make(chan error, len(notifiers))
	for _, notifier := range notifiers {
		go func(n database.Notifier) {
			err := s.sendSingleNotification(n, task, conversation, status, completionTime, errorMsg, adminLang)
			errorChan <- err
		}(notifier)
	}

	// Collect errors (don't fail the entire operation if some notifications fail)
	var errors []string
	for i := 0; i < len(notifiers); i++ {
		if err := <-errorChan; err != nil {
			errors = append(errors, err.Error())
			utils.Error("Failed to send notification", "notifierID", notifiers[i].ID, "error", err)
		}
	}

	if len(errors) > 0 {
		utils.Warn("Some notifications failed to send", "errors", strings.Join(errors, "; "))
	}

	return nil
}

func (s *notifierService) sendSingleNotification(notifier database.Notifier, task *database.Task, conversation *database.TaskConversation, status database.ConversationStatus, completionTime time.Time, errorMsg string, adminLang string) error {
	// Skip if notifier is disabled (safety check)
	if !notifier.IsEnabled {
		return nil
	}

	// Parse config
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(notifier.Config), &config); err != nil {
		return fmt.Errorf("failed to parse notifier config: %v", err)
	}

	// Create provider
	provider, err := notifiers.NewProvider(notifier.Type, config)
	if err != nil {
		return fmt.Errorf("failed to create notification provider: %v", err)
	}

	// Prepare notification content
	title := task.Title
	content := conversation.Content
	if len(content) > 100 {
		content = content[:100] + "..."
	}

	// Add error message if status is failed
	if status == database.ConversationStatusFailed && errorMsg != "" {
		content += fmt.Sprintf("\n\nError: %s", errorMsg)
	}

	// Send notification
	return provider.Send(title, content, status, adminLang)
}

// Permission helpers

func (s *notifierService) CanAdminAccessNotifier(notifierID uint, admin *database.Admin) (bool, error) {
	// Super admin can access all notifiers
	if admin.Role == database.AdminRoleSuperAdmin {
		return true, nil
	}

	// Admin can only access their own notifiers
	if admin.Role == database.AdminRoleAdmin {
		isOwner, err := s.repo.IsOwner(notifierID, admin.ID)
		if err != nil {
			return false, err
		}
		return isOwner, nil
	}

	// Other roles cannot access notifiers
	return false, nil
}

func (s *notifierService) IsNotifierOwner(notifierID uint, admin *database.Admin) (bool, error) {
	return s.repo.IsOwner(notifierID, admin.ID)
}

func (s *notifierService) validateNotifierData(name string, notifierType database.NotifierType, config map[string]interface{}) error {
	if strings.TrimSpace(name) == "" {
		return appErrors.NewI18nError("required_field", "en-US")
	}

	// Validate notifier type
	validTypes := []database.NotifierType{
		database.NotifierTypeWeChatWork,
		database.NotifierTypeDingTalk,
		database.NotifierTypeFeishu,
		database.NotifierTypeSlack,
		database.NotifierTypeDiscord,
		database.NotifierTypeWebhook,
	}

	valid := false
	for _, validType := range validTypes {
		if notifierType == validType {
			valid = true
			break
		}
	}

	if !valid {
		return appErrors.NewI18nError("invalid_notifier_type", "en-US")
	}

	if config == nil || len(config) == 0 {
		return appErrors.NewI18nError("notifier_config_required", "en-US")
	}

	return nil
}
