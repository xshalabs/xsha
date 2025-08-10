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

func (s *projectService) CreateProject(name, description, repoURL, protocol string, credentialID *uint, createdBy string) (*database.Project, error) {
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
		RepoURL:      repoURL,
		Protocol:     protocolType,
		CredentialID: credentialID,
		CreatedBy:    createdBy,
	}

	if err := s.repo.Create(project); err != nil {
		return nil, err
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
		return []ProjectWithTaskCount{}, total, nil
	}

	projectIDs := make([]uint, len(projects))
	for i, project := range projects {
		projectIDs[i] = project.ID
	}

	taskCounts, err := s.repo.GetTaskCounts(projectIDs)
	if err != nil {
		return nil, 0, err
	}

	projectsWithTaskCount := make([]ProjectWithTaskCount, len(projects))
	for i, project := range projects {
		projectsWithTaskCount[i] = ProjectWithTaskCount{
			Project:   &project,
			TaskCount: taskCounts[project.ID],
		}
	}

	return projectsWithTaskCount, total, nil
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
	}

	return s.repo.Update(project)
}

func (s *projectService) DeleteProject(id uint) error {
	project, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	inProgressStatus := database.TaskStatusInProgress
	tasks, _, err := s.taskRepo.List(&project.ID, &inProgressStatus, nil, nil, nil, 1, 1)
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

func (s *projectService) GetCompatibleCredentials(protocol database.GitProtocolType) ([]database.GitCredential, error) {
	switch protocol {
	case database.GitProtocolHTTPS:
		passwordType := database.GitCredentialTypePassword
		passwordCreds, err := s.gitCredService.ListActiveCredentials(&passwordType)
		if err != nil {
			return nil, err
		}

		tokenType := database.GitCredentialTypeToken
		tokenCreds, err := s.gitCredService.ListActiveCredentials(&tokenType)
		if err != nil {
			return nil, err
		}

		credentials := append(passwordCreds, tokenCreds...)
		return credentials, nil

	case database.GitProtocolSSH:
		sshType := database.GitCredentialTypeSSHKey
		return s.gitCredService.ListActiveCredentials(&sshType)

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

func (s *projectService) ValidateRepositoryAccess(repoURL string, credentialID *uint) error {
	result, err := s.FetchRepositoryBranches(repoURL, credentialID)
	if err != nil {
		return err
	}

	if !result.CanAccess {
		return fmt.Errorf(result.ErrorMessage)
	}

	return nil
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
		if !strings.HasPrefix(repoURL, "https://") {
			return appErrors.ErrInvalidFormat
		}
		if _, err := url.Parse(repoURL); err != nil {
			return fmt.Errorf("invalid HTTPS URL format: %v", err)
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
