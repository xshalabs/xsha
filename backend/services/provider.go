package services

import (
	"encoding/json"
	"strings"
	"xsha-backend/database"
	appErrors "xsha-backend/errors"
	"xsha-backend/repository"
)

type providerService struct {
	repo        repository.ProviderRepository
	devEnvRepo  repository.DevEnvironmentRepository
}

func NewProviderService(repo repository.ProviderRepository, devEnvRepo repository.DevEnvironmentRepository) ProviderService {
	return &providerService{
		repo:       repo,
		devEnvRepo: devEnvRepo,
	}
}

// CRUD operations with permission checks

func (s *providerService) CreateProvider(name, description, providerType, config string, admin *database.Admin) (*database.Provider, error) {
	// All logged-in users (Developer, Admin, SuperAdmin) can create providers
	if admin == nil {
		return nil, appErrors.NewI18nError("permission_denied", "en-US")
	}

	// Validate input
	if err := s.validateProviderData(name, providerType, config); err != nil {
		return nil, err
	}

	// Check if name already exists
	if existing, _ := s.repo.GetByName(name); existing != nil {
		return nil, appErrors.NewI18nError("provider_name_exists", "en-US")
	}

	// Validate config JSON format
	if err := s.ValidateConfig(config); err != nil {
		return nil, err
	}

	provider := &database.Provider{
		Name:        name,
		Description: description,
		Type:        database.ProviderType(providerType),
		Config:      config,
		AdminID:     &admin.ID,
		CreatedBy:   admin.Username,
	}

	if err := s.repo.Create(provider); err != nil {
		return nil, err
	}

	return provider, nil
}

func (s *providerService) GetProvider(id uint, admin *database.Admin) (*database.Provider, error) {
	provider, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Check permission
	canAccess, err := s.CanAdminAccessProvider(id, admin)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, appErrors.NewI18nError("permission_denied", "en-US")
	}

	return provider, nil
}

func (s *providerService) ListProviders(admin *database.Admin, name *string, providerType *string, page, pageSize int) ([]database.ProviderListItemResponse, int64, error) {
	if admin == nil {
		return []database.ProviderListItemResponse{}, 0, appErrors.NewI18nError("permission_denied", "en-US")
	}

	providers, total, err := s.repo.ListByAdminAccess(admin.ID, admin.Role, name, providerType, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	responses := database.ToProviderListItemResponses(providers)
	return responses, total, nil
}

func (s *providerService) UpdateProvider(id uint, updates map[string]interface{}, admin *database.Admin) error {
	// Check permission
	canAccess, err := s.CanAdminAccessProvider(id, admin)
	if err != nil {
		return err
	}
	if !canAccess {
		return appErrors.NewI18nError("permission_denied", "en-US")
	}

	provider, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// Reject type field updates - type cannot be changed after creation
	if _, ok := updates["type"]; ok {
		return appErrors.NewI18nError("provider_type_cannot_be_modified", "en-US")
	}

	// Update fields
	if name, ok := updates["name"]; ok {
		nameStr := name.(string)
		// Check if name already exists (except current provider)
		if existing, _ := s.repo.GetByName(nameStr); existing != nil && existing.ID != id {
			return appErrors.NewI18nError("provider_name_exists", "en-US")
		}
		provider.Name = nameStr
	}
	if description, ok := updates["description"]; ok {
		provider.Description = description.(string)
	}
	if config, ok := updates["config"]; ok {
		configStr := config.(string)
		// Validate configuration if updated
		if err := s.ValidateConfig(configStr); err != nil {
			return err
		}
		provider.Config = configStr
	}

	return s.repo.Update(provider)
}

func (s *providerService) DeleteProvider(id uint, admin *database.Admin) error {
	// Check permission
	canAccess, err := s.CanAdminAccessProvider(id, admin)
	if err != nil {
		return err
	}
	if !canAccess {
		return appErrors.NewI18nError("permission_denied", "en-US")
	}

	// Check if provider is being used by any dev environments
	count, err := s.repo.CountByProviderID(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return appErrors.NewI18nError("provider_in_use", "en-US")
	}

	return s.repo.Delete(id)
}

// Provider-specific methods

func (s *providerService) ValidateConfig(config string) error {
	if strings.TrimSpace(config) == "" {
		return appErrors.NewI18nError("provider_config_required", "en-US")
	}

	// Validate that config is valid JSON
	var configMap map[string]interface{}
	if err := json.Unmarshal([]byte(config), &configMap); err != nil {
		return appErrors.NewI18nError("provider_config_invalid", "en-US")
	}

	return nil
}

func (s *providerService) GetProviderTypes() []string {
	return []string{
		string(database.ProviderTypeClaudeCode),
		// Add more types here as they are implemented
	}
}

// Permission helpers

func (s *providerService) CanAdminAccessProvider(providerID uint, admin *database.Admin) (bool, error) {
	if admin == nil {
		return false, nil
	}

	// Super admin can access all providers
	if admin.Role == database.AdminRoleSuperAdmin {
		return true, nil
	}

	// Admin and Developer can only access their own providers
	isOwner, err := s.repo.IsOwner(providerID, admin.ID)
	if err != nil {
		return false, err
	}
	return isOwner, nil
}

func (s *providerService) IsProviderOwner(providerID uint, admin *database.Admin) (bool, error) {
	if admin == nil {
		return false, nil
	}
	return s.repo.IsOwner(providerID, admin.ID)
}

// Private validation helper

func (s *providerService) validateProviderData(name, providerType, config string) error {
	if strings.TrimSpace(name) == "" {
		return appErrors.NewI18nError("required_field", "en-US")
	}

	if strings.TrimSpace(providerType) == "" {
		return appErrors.NewI18nError("provider_type_required", "en-US")
	}

	// Validate provider type
	validTypes := s.GetProviderTypes()
	valid := false
	for _, validType := range validTypes {
		if providerType == validType {
			valid = true
			break
		}
	}

	if !valid {
		return appErrors.NewI18nError("invalid_provider_type", "en-US")
	}

	if strings.TrimSpace(config) == "" {
		return appErrors.NewI18nError("provider_config_required", "en-US")
	}

	return nil
}
