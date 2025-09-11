package services

import (
	"fmt"
	"net/url"
	"strings"
	"xsha-backend/config"
	"xsha-backend/database"
	appErrors "xsha-backend/errors"
	"xsha-backend/repository"
	"xsha-backend/utils"
)

type projectService struct {
	repo                repository.ProjectRepository
	gitCredRepo         repository.GitCredentialRepository
	gitCredService      GitCredentialService
	taskRepo            repository.TaskRepository
	systemConfigService SystemConfigService
	config              *config.Config
}

type ProjectWithTaskCount struct {
	*database.Project
	TaskCount int64 `json:"task_count"`
}

type ProjectListItemWithCounts struct {
	*database.ProjectListItemResponse
	TaskCount int64 `json:"task_count"`
}

func NewProjectService(repo repository.ProjectRepository, gitCredRepo repository.GitCredentialRepository, gitCredService GitCredentialService, taskRepo repository.TaskRepository, systemConfigService SystemConfigService, cfg *config.Config) ProjectService {
	return &projectService{
		repo:                repo,
		gitCredRepo:         gitCredRepo,
		gitCredService:      gitCredService,
		taskRepo:            taskRepo,
		systemConfigService: systemConfigService,
		config:              cfg,
	}
}

func (s *projectService) CreateProject(name, description, systemPrompt, repoURL, protocol string, credentialID *uint, adminID *uint, createdBy string) (*database.Project, error) {
	if err := s.validateProjectData(name, repoURL, protocol); err != nil {
		return nil, err
	}

	if existing, _ := s.repo.GetByName(name); existing != nil {
		return nil, appErrors.ErrProjectNameExists
	}

	protocolType := database.GitProtocolType(protocol)
	if err := s.validateRepositoryURL(repoURL, protocolType); err != nil {
		return nil, err
	}

	if err := s.ValidateProtocolCredential(protocolType, credentialID); err != nil {
		return nil, err
	}

	project := &database.Project{
		Name:         name,
		Description:  description,
		SystemPrompt: systemPrompt,
		RepoURL:      repoURL,
		Protocol:     protocolType,
		CredentialID: credentialID,
		AdminID:      adminID,
		CreatedBy:    createdBy,
	}

	if err := s.repo.Create(project); err != nil {
		return nil, err
	}

	// Add creator as admin to the project
	if adminID != nil {
		if err := s.repo.AddAdmin(project.ID, *adminID); err != nil {
			utils.Error("Failed to add creator as admin to project", "projectID", project.ID, "adminID", *adminID, "error", err)
		}
	}

	return project, nil
}

func (s *projectService) GetProject(id uint) (*database.Project, error) {
	return s.repo.GetByID(id)
}

func (s *projectService) ListProjects(name string, protocol *database.GitProtocolType, page, pageSize int) ([]database.Project, int64, error) {
	return s.repo.List(name, protocol, "created_at", "desc", page, pageSize)
}

func (s *projectService) ListProjectsWithTaskCount(name string, protocol *database.GitProtocolType, sortBy, sortDirection string, page, pageSize int) (interface{}, int64, error) {
	projects, total, err := s.repo.List(name, protocol, sortBy, sortDirection, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	if len(projects) == 0 {
		return []database.ProjectListItemResponse{}, total, nil
	}

	projectIDs := make([]uint, len(projects))
	for i, project := range projects {
		projectIDs[i] = project.ID
	}

	// Get task counts and admin counts for these projects
	taskCounts, err := s.repo.GetTaskCounts(projectIDs)
	if err != nil {
		return nil, 0, err
	}

	adminCounts, err := s.repo.GetAdminCounts(projectIDs)
	if err != nil {
		return nil, 0, err
	}

	// Convert to ProjectListItemResponse with counts
	responses := database.ToProjectListItemResponses(projects)
	for i := range responses {
		responses[i].AdminCount = adminCounts[responses[i].ID]
	}

	// Create response with task counts (keeping existing structure for compatibility)
	projectsWithCounts := make([]ProjectListItemWithCounts, len(responses))
	for i, response := range responses {
		projectsWithCounts[i] = ProjectListItemWithCounts{
			ProjectListItemResponse: &response,
			TaskCount:               taskCounts[response.ID],
		}
	}

	return projectsWithCounts, total, nil
}

func (s *projectService) UpdateProject(id uint, updates map[string]interface{}) error {
	project, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if name, ok := updates["name"]; ok {
		project.Name = name.(string)
	}
	if description, ok := updates["description"]; ok {
		project.Description = description.(string)
	}
	if systemPrompt, ok := updates["system_prompt"]; ok {
		project.SystemPrompt = systemPrompt.(string)
	}
	if repoURL, ok := updates["repo_url"]; ok {
		project.RepoURL = repoURL.(string)
		if err := s.validateRepositoryURL(project.RepoURL, project.Protocol); err != nil {
			return err
		}
	}

	if credentialID, ok := updates["credential_id"]; ok {
		if credentialID == nil {
			project.CredentialID = nil
		} else {
			if idPtr, ok := credentialID.(*uint); ok {
				project.CredentialID = idPtr
			} else {
				if id, ok := credentialID.(uint); ok {
					project.CredentialID = &id
				} else {
					return fmt.Errorf("invalid credential_id type")
				}
			}
		}
		if err := s.ValidateProtocolCredential(project.Protocol, project.CredentialID); err != nil {
			return err
		}
		// Clear the stale credential relationship since credential_id has changed
		project.Credential = nil
	}

	if err := s.repo.Update(project); err != nil {
		return err
	}

	return nil
}

func (s *projectService) DeleteProject(id uint) error {
	project, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	inProgressStatuses := []database.TaskStatus{database.TaskStatusInProgress}
	tasks, _, err := s.taskRepo.List(&project.ID, inProgressStatuses, nil, nil, nil, "created_at", "desc", 1, 1)
	if err != nil {
		return fmt.Errorf("failed to check project tasks: %v", err)
	}
	if len(tasks) > 0 {
		return appErrors.ErrProjectHasInProgressTasks
	}

	return s.repo.Delete(id)
}

func (s *projectService) ValidateProtocolCredential(protocol database.GitProtocolType, credentialID *uint) error {
	if credentialID == nil {
		return nil
	}

	credential, err := s.gitCredRepo.GetByID(*credentialID)
	if err != nil {
		return fmt.Errorf("credential not found: %v", err)
	}

	switch protocol {
	case database.GitProtocolHTTPS:
		if credential.Type != database.GitCredentialTypePassword && credential.Type != database.GitCredentialTypeToken {
			return appErrors.ErrIncompatibleCredential
		}
	case database.GitProtocolSSH:
		if credential.Type != database.GitCredentialTypeSSHKey {
			return appErrors.ErrIncompatibleCredential
		}
	default:
		return appErrors.ErrInvalidProtocol
	}

	return nil
}

func (s *projectService) GetCompatibleCredentials(protocol database.GitProtocolType, admin *database.Admin) ([]database.GitCredential, error) {
	switch protocol {
	case database.GitProtocolHTTPS:
		passwordType := database.GitCredentialTypePassword
		tokenType := database.GitCredentialTypeToken

		var passwordCreds, tokenCreds []database.GitCredential
		var err error

		if admin.Role == database.AdminRoleSuperAdmin {
			// Super admin can see all credentials
			passwordCreds, err = s.gitCredService.ListActiveCredentials(&passwordType)
			if err != nil {
				return nil, err
			}

			tokenCreds, err = s.gitCredService.ListActiveCredentials(&tokenType)
			if err != nil {
				return nil, err
			}
		} else {
			// Regular admin can only see credentials they have access to
			passwordCreds, err = s.gitCredService.ListActiveCredentialsByAdminAccess(admin.ID, &passwordType)
			if err != nil {
				return nil, err
			}

			tokenCreds, err = s.gitCredService.ListActiveCredentialsByAdminAccess(admin.ID, &tokenType)
			if err != nil {
				return nil, err
			}
		}

		credentials := append(passwordCreds, tokenCreds...)
		return credentials, nil

	case database.GitProtocolSSH:
		sshType := database.GitCredentialTypeSSHKey

		if admin.Role == database.AdminRoleSuperAdmin {
			// Super admin can see all credentials
			return s.gitCredService.ListActiveCredentials(&sshType)
		} else {
			// Regular admin can only see credentials they have access to
			return s.gitCredService.ListActiveCredentialsByAdminAccess(admin.ID, &sshType)
		}

	default:
		return nil, appErrors.ErrInvalidProtocol
	}
}

func (s *projectService) FetchRepositoryBranches(repoURL string, credentialID *uint) (*utils.GitAccessResult, error) {
	if err := utils.ValidateGitURL(repoURL); err != nil {
		return &utils.GitAccessResult{
			CanAccess:    false,
			ErrorMessage: fmt.Sprintf("invalid repository URL format: %v", err),
		}, nil
	}

	var credentialInfo *utils.GitCredentialInfo
	if credentialID != nil {
		credential, err := s.gitCredRepo.GetByID(*credentialID)
		if err != nil {
			return &utils.GitAccessResult{
				CanAccess:    false,
				ErrorMessage: fmt.Sprintf("failed to get credential: %v", err),
			}, nil
		}

		credentialInfo = &utils.GitCredentialInfo{
			Type:     utils.GitCredentialType(credential.Type),
			Username: credential.Username,
		}

		switch credential.Type {
		case database.GitCredentialTypePassword, database.GitCredentialTypeToken:
			if credential.PasswordHash != "" {
				credentialInfo.Password = credential.PasswordHash
			}
		case database.GitCredentialTypeSSHKey:
			if credential.PrivateKey != "" {
				credentialInfo.PrivateKey = credential.PrivateKey
				credentialInfo.PublicKey = credential.PublicKey
			}
		}
	}

	proxyConfig, err := s.getGitProxyConfig()
	if err != nil {
		utils.Warn("Failed to get proxy config, using no proxy", "error", err)
		proxyConfig = nil
	}

	gitSSLVerify, err := s.systemConfigService.GetGitSSLVerify()
	if err != nil {
		utils.Warn("Failed to get git SSL verify setting, using default false", "error", err)
		gitSSLVerify = false
	}

	return utils.FetchRepositoryBranchesWithConfig(repoURL, credentialInfo, gitSSLVerify, proxyConfig)
}

func (s *projectService) getGitProxyConfig() (*utils.GitProxyConfig, error) {
	return s.systemConfigService.GetGitProxyConfig()
}

func (s *projectService) validateProjectData(name, repoURL, protocol string) error {
	if strings.TrimSpace(name) == "" {
		return appErrors.ErrRequired
	}
	if strings.TrimSpace(repoURL) == "" {
		return appErrors.ErrRequired
	}
	if protocol != string(database.GitProtocolHTTPS) && protocol != string(database.GitProtocolSSH) {
		return appErrors.ErrInvalidProtocol
	}
	return nil
}

func (s *projectService) validateRepositoryURL(repoURL string, protocol database.GitProtocolType) error {
	switch protocol {
	case database.GitProtocolHTTPS:
		if !strings.HasPrefix(repoURL, "https://") && !strings.HasPrefix(repoURL, "http://") {
			return appErrors.ErrInvalidFormat
		}
		if _, err := url.Parse(repoURL); err != nil {
			return fmt.Errorf("invalid HTTP/HTTPS URL format: %v", err)
		}
	case database.GitProtocolSSH:
		if !strings.Contains(repoURL, "@") || !strings.Contains(repoURL, ":") {
			return appErrors.ErrInvalidFormat
		}
	default:
		return appErrors.ErrInvalidProtocol
	}
	return nil
}

// GetProjectWithAdmins retrieves a project with its admin relationships preloaded
func (s *projectService) GetProjectWithAdmins(id uint) (*database.Project, error) {
	return s.repo.GetByIDWithAdmins(id)
}

// ListProjectsByAdminAccess lists projects that an admin has access to with task counts
func (s *projectService) ListProjectsByAdminAccess(adminID uint, name string, protocol *database.GitProtocolType, sortBy, sortDirection string, page, pageSize int) (interface{}, int64, error) {
	projects, total, err := s.repo.ListByAdminAccess(adminID, name, protocol, sortBy, sortDirection, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	if len(projects) == 0 {
		return []database.ProjectListItemResponse{}, total, nil
	}

	// Get project IDs for count lookups
	projectIDs := make([]uint, len(projects))
	for i, project := range projects {
		projectIDs[i] = project.ID
	}

	// Get task counts and admin counts for these projects
	taskCounts, err := s.repo.GetTaskCounts(projectIDs)
	if err != nil {
		return nil, 0, err
	}

	adminCounts, err := s.repo.GetAdminCounts(projectIDs)
	if err != nil {
		return nil, 0, err
	}

	// Convert to ProjectListItemResponse with counts
	responses := database.ToProjectListItemResponses(projects)
	for i := range responses {
		responses[i].AdminCount = adminCounts[responses[i].ID]
	}

	// Create response with task counts (keeping existing structure for compatibility)
	result := make([]ProjectListItemWithCounts, len(responses))
	for i, response := range responses {
		result[i] = ProjectListItemWithCounts{
			ProjectListItemResponse: &response,
			TaskCount:               taskCounts[response.ID],
		}
	}

	return result, total, nil
}

// AddAdminToProject adds an admin to the project's admin list
func (s *projectService) AddAdminToProject(projectID, adminID uint) error {
	_, err := s.repo.GetByID(projectID)
	if err != nil {
		return appErrors.ErrProjectNotFound
	}
	return s.repo.AddAdmin(projectID, adminID)
}

// RemoveAdminFromProject removes an admin from the project's admin list
func (s *projectService) RemoveAdminFromProject(projectID, adminID uint) error {
	// Check if project exists
	project, err := s.repo.GetByID(projectID)
	if err != nil {
		return appErrors.ErrProjectNotFound
	}

	// Check if trying to remove the primary admin
	if project.AdminID != nil && *project.AdminID == adminID {
		return appErrors.ErrProjectCannotRemovePrimaryAdmin
	}

	return s.repo.RemoveAdmin(projectID, adminID)
}

// GetProjectAdmins gets all admins for a specific project
func (s *projectService) GetProjectAdmins(projectID uint) ([]database.Admin, error) {
	_, err := s.repo.GetByID(projectID)
	if err != nil {
		return nil, appErrors.ErrProjectNotFound
	}
	return s.repo.GetAdmins(projectID)
}

// CanAdminAccessProject checks if an admin can access a project
func (s *projectService) CanAdminAccessProject(projectID, adminID uint) (bool, error) {
	project, err := s.repo.GetByIDWithAdmins(projectID)
	if err != nil {
		return false, err
	}

	// Check if admin is the primary admin
	if project.AdminID != nil && *project.AdminID == adminID {
		return true, nil
	}

	// Check if admin is in the many-to-many relationship
	for _, admin := range project.Admins {
		if admin.ID == adminID {
			return true, nil
		}
	}

	return false, nil
}

func (s *projectService) IsOwner(projectID, adminID uint) (bool, error) {
	project, err := s.repo.GetByID(projectID)
	if err != nil {
		return false, err
	}

	if project.AdminID != nil && *project.AdminID == adminID {
		return true, nil
	}

	return false, nil
}
