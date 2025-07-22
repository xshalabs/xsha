package services

import (
	"errors"
	"fmt"
	"net/url"
	"sleep0-backend/config"
	"sleep0-backend/database"
	"sleep0-backend/repository"
	"sleep0-backend/utils"
	"strings"
)

type projectService struct {
	repo           repository.ProjectRepository
	gitCredRepo    repository.GitCredentialRepository
	gitCredService GitCredentialService
	config         *config.Config
}

// NewProjectService 创建项目服务实例
func NewProjectService(repo repository.ProjectRepository, gitCredRepo repository.GitCredentialRepository, gitCredService GitCredentialService, cfg *config.Config) ProjectService {
	return &projectService{
		repo:           repo,
		gitCredRepo:    gitCredRepo,
		gitCredService: gitCredService,
		config:         cfg,
	}
}

// CreateProject 创建项目
func (s *projectService) CreateProject(name, description, repoURL, protocol, defaultBranch, createdBy string, credentialID *uint) (*database.Project, error) {
	// 验证输入
	if err := s.validateProjectData(name, repoURL, protocol, defaultBranch); err != nil {
		return nil, err
	}

	// 检查项目名称是否已存在
	if existing, _ := s.repo.GetByName(name, createdBy); existing != nil {
		return nil, errors.New("project name already exists")
	}

	// 解析并验证协议
	protocolType := database.GitProtocolType(protocol)
	if err := s.validateRepositoryURL(repoURL, protocolType); err != nil {
		return nil, err
	}

	// 验证协议和凭据的兼容性
	if err := s.ValidateProtocolCredential(protocolType, credentialID, createdBy); err != nil {
		return nil, err
	}

	// 创建项目对象
	project := &database.Project{
		Name:          name,
		Description:   description,
		RepoURL:       repoURL,
		Protocol:      protocolType,
		DefaultBranch: defaultBranch,
		CredentialID:  credentialID,
		CreatedBy:     createdBy,
		IsActive:      true,
	}

	// 设置默认分支
	if project.DefaultBranch == "" {
		project.DefaultBranch = "main"
	}

	// 保存到数据库
	if err := s.repo.Create(project); err != nil {
		return nil, err
	}

	return project, nil
}

// GetProject 获取项目
func (s *projectService) GetProject(id uint, createdBy string) (*database.Project, error) {
	return s.repo.GetByID(id, createdBy)
}

// GetProjectByName 根据名称获取项目
func (s *projectService) GetProjectByName(name, createdBy string) (*database.Project, error) {
	return s.repo.GetByName(name, createdBy)
}

// ListProjects 获取项目列表
func (s *projectService) ListProjects(createdBy string, protocol *database.GitProtocolType, page, pageSize int) ([]database.Project, int64, error) {
	return s.repo.List(createdBy, protocol, page, pageSize)
}

// UpdateProject 更新项目
func (s *projectService) UpdateProject(id uint, createdBy string, updates map[string]interface{}) error {
	project, err := s.repo.GetByID(id, createdBy)
	if err != nil {
		return err
	}

	// 更新基本信息
	if name, ok := updates["name"]; ok {
		project.Name = name.(string)
	}
	if description, ok := updates["description"]; ok {
		project.Description = description.(string)
	}
	if repoURL, ok := updates["repo_url"]; ok {
		project.RepoURL = repoURL.(string)
		// 重新验证仓库URL
		if err := s.validateRepositoryURL(project.RepoURL, project.Protocol); err != nil {
			return err
		}
	}
	if defaultBranch, ok := updates["default_branch"]; ok {
		project.DefaultBranch = defaultBranch.(string)
	}
	if credentialID, ok := updates["credential_id"]; ok {
		if credentialID == nil {
			project.CredentialID = nil
		} else {
			id := credentialID.(uint)
			project.CredentialID = &id
		}
		// 验证协议和凭据的兼容性
		if err := s.ValidateProtocolCredential(project.Protocol, project.CredentialID, createdBy); err != nil {
			return err
		}
	}

	return s.repo.Update(project)
}

// DeleteProject 删除项目
func (s *projectService) DeleteProject(id uint, createdBy string) error {
	return s.repo.Delete(id, createdBy)
}

// UseProject 使用项目（更新最后使用时间）
func (s *projectService) UseProject(id uint, createdBy string) (*database.Project, error) {
	project, err := s.repo.GetByID(id, createdBy)
	if err != nil {
		return nil, err
	}

	if !project.IsActive {
		return nil, errors.New("project is not active")
	}

	// 更新最后使用时间
	if err := s.repo.UpdateLastUsed(id, createdBy); err != nil {
		return nil, err
	}

	return project, nil
}

// ToggleProject 切换项目激活状态
func (s *projectService) ToggleProject(id uint, createdBy string, isActive bool) error {
	return s.repo.SetActive(id, createdBy, isActive)
}

// ListActiveProjects 获取激活的项目列表
func (s *projectService) ListActiveProjects(createdBy string, protocol *database.GitProtocolType) ([]database.Project, error) {
	return s.repo.ListActive(createdBy, protocol)
}

// ValidateProtocolCredential 验证协议和凭据的兼容性
func (s *projectService) ValidateProtocolCredential(protocol database.GitProtocolType, credentialID *uint, createdBy string) error {
	// 如果没有绑定凭据，则跳过验证
	if credentialID == nil {
		return nil
	}

	// 获取凭据信息
	credential, err := s.gitCredRepo.GetByID(*credentialID, createdBy)
	if err != nil {
		return fmt.Errorf("credential not found: %v", err)
	}

	if !credential.IsActive {
		return errors.New("credential is not active")
	}

	// 验证协议和凭据类型的兼容性
	switch protocol {
	case database.GitProtocolHTTPS:
		if credential.Type != database.GitCredentialTypePassword && credential.Type != database.GitCredentialTypeToken {
			return errors.New("HTTPS protocol only supports password or token credentials")
		}
	case database.GitProtocolSSH:
		if credential.Type != database.GitCredentialTypeSSHKey {
			return errors.New("SSH protocol only supports SSH key credentials")
		}
	default:
		return errors.New("unsupported protocol type")
	}

	return nil
}

// GetCompatibleCredentials 获取与协议兼容的凭据列表
func (s *projectService) GetCompatibleCredentials(protocol database.GitProtocolType, createdBy string) ([]database.GitCredential, error) {
	switch protocol {
	case database.GitProtocolHTTPS:
		// HTTPS协议支持password和token类型，这里我们需要分别查询
		passwordType := database.GitCredentialTypePassword
		passwordCreds, err := s.gitCredService.ListActiveCredentials(createdBy, &passwordType)
		if err != nil {
			return nil, err
		}

		tokenType := database.GitCredentialTypeToken
		tokenCreds, err := s.gitCredService.ListActiveCredentials(createdBy, &tokenType)
		if err != nil {
			return nil, err
		}

		// 合并结果
		credentials := append(passwordCreds, tokenCreds...)
		return credentials, nil

	case database.GitProtocolSSH:
		sshType := database.GitCredentialTypeSSHKey
		return s.gitCredService.ListActiveCredentials(createdBy, &sshType)

	default:
		return nil, errors.New("unsupported protocol type")
	}
}

// FetchRepositoryBranches 获取仓库分支列表
func (s *projectService) FetchRepositoryBranches(repoURL string, credentialID *uint, createdBy string) (*utils.GitAccessResult, error) {
	// 验证仓库URL格式
	if err := utils.ValidateGitURL(repoURL); err != nil {
		return &utils.GitAccessResult{
			CanAccess:    false,
			ErrorMessage: fmt.Sprintf("仓库URL格式无效: %v", err),
		}, nil
	}

	// 获取凭据信息（如果提供了凭据ID）
	var credentialInfo *utils.GitCredentialInfo
	if credentialID != nil {
		credential, err := s.gitCredRepo.GetByID(*credentialID, createdBy)
		if err != nil {
			return &utils.GitAccessResult{
				CanAccess:    false,
				ErrorMessage: fmt.Sprintf("获取凭据失败: %v", err),
			}, nil
		}

		if !credential.IsActive {
			return &utils.GitAccessResult{
				CanAccess:    false,
				ErrorMessage: "凭据未激活",
			}, nil
		}

		// 解密凭据信息
		credentialInfo = &utils.GitCredentialInfo{
			Type:     utils.GitCredentialType(credential.Type),
			Username: credential.Username,
		}

		switch credential.Type {
		case database.GitCredentialTypePassword, database.GitCredentialTypeToken:
			if credential.PasswordHash != "" {
				password, err := utils.DecryptAES(credential.PasswordHash, s.config.AESKey)
				if err != nil {
					return &utils.GitAccessResult{
						CanAccess:    false,
						ErrorMessage: fmt.Sprintf("解密凭据失败: %v", err),
					}, nil
				}
				credentialInfo.Password = password
			}
		case database.GitCredentialTypeSSHKey:
			if credential.PrivateKey != "" {
				privateKey, err := utils.DecryptAES(credential.PrivateKey, s.config.AESKey)
				if err != nil {
					return &utils.GitAccessResult{
						CanAccess:    false,
						ErrorMessage: fmt.Sprintf("解密SSH私钥失败: %v", err),
					}, nil
				}
				credentialInfo.PrivateKey = privateKey
				credentialInfo.PublicKey = credential.PublicKey
			}
		}
	}

	// 使用Git工具获取分支信息
	return utils.FetchRepositoryBranches(repoURL, credentialInfo)
}

// ValidateRepositoryAccess 验证仓库访问权限
func (s *projectService) ValidateRepositoryAccess(repoURL string, credentialID *uint, createdBy string) error {
	result, err := s.FetchRepositoryBranches(repoURL, credentialID, createdBy)
	if err != nil {
		return err
	}

	if !result.CanAccess {
		return fmt.Errorf(result.ErrorMessage)
	}

	return nil
}

// validateProjectData 验证项目数据
func (s *projectService) validateProjectData(name, repoURL, protocol, defaultBranch string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("project name is required")
	}
	if strings.TrimSpace(repoURL) == "" {
		return errors.New("repository URL is required")
	}
	if protocol != string(database.GitProtocolHTTPS) && protocol != string(database.GitProtocolSSH) {
		return errors.New("unsupported protocol type")
	}
	if strings.TrimSpace(defaultBranch) == "" {
		return errors.New("default branch is required")
	}
	return nil
}

// validateRepositoryURL 验证仓库URL格式
func (s *projectService) validateRepositoryURL(repoURL string, protocol database.GitProtocolType) error {
	switch protocol {
	case database.GitProtocolHTTPS:
		if !strings.HasPrefix(repoURL, "https://") {
			return errors.New("HTTPS protocol requires URL to start with 'https://'")
		}
		if _, err := url.Parse(repoURL); err != nil {
			return fmt.Errorf("invalid HTTPS URL format: %v", err)
		}
	case database.GitProtocolSSH:
		if !strings.Contains(repoURL, "@") || !strings.Contains(repoURL, ":") {
			return errors.New("SSH protocol requires URL in format 'user@host:path' or 'ssh://user@host/path'")
		}
	default:
		return errors.New("unsupported protocol type")
	}
	return nil
}
