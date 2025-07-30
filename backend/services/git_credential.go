package services

import (
	"errors"
	"fmt"
	"strings"
	"xsha-backend/config"
	"xsha-backend/database"
	"xsha-backend/repository"
	"xsha-backend/utils"
)

type gitCredentialService struct {
	repo        repository.GitCredentialRepository
	projectRepo repository.ProjectRepository
	config      *config.Config
}

// NewGitCredentialService creates a Git credential service instance
func NewGitCredentialService(repo repository.GitCredentialRepository, projectRepo repository.ProjectRepository, cfg *config.Config) GitCredentialService {
	return &gitCredentialService{
		repo:        repo,
		projectRepo: projectRepo,
		config:      cfg,
	}
}

// CreateCredential creates a Git credential
func (s *gitCredentialService) CreateCredential(name, description, credType, username, createdBy string, secretData map[string]string) (*database.GitCredential, error) {
	// Validate input
	if err := s.ValidateCredentialData(credType, secretData); err != nil {
		return nil, err
	}

	// Check if name already exists
	if existing, _ := s.repo.GetByName(name, createdBy); existing != nil {
		return nil, errors.New("credential name already exists")
	}

	// Create credential object
	credential := &database.GitCredential{
		Name:        name,
		Description: description,
		Type:        database.GitCredentialType(credType),
		Username:    username,
		CreatedBy:   createdBy,
	}

	// Encrypt sensitive data
	switch database.GitCredentialType(credType) {
	case database.GitCredentialTypePassword, database.GitCredentialTypeToken:
		if password, ok := secretData["password"]; ok {
			encrypted, err := utils.EncryptAES(password, s.config.AESKey)
			if err != nil {
				return nil, fmt.Errorf("failed to encrypt password: %v", err)
			}
			credential.PasswordHash = encrypted
		}
	case database.GitCredentialTypeSSHKey:
		if privateKey, ok := secretData["private_key"]; ok {
			encrypted, err := utils.EncryptAES(privateKey, s.config.AESKey)
			if err != nil {
				return nil, fmt.Errorf("failed to encrypt private key: %v", err)
			}
			credential.PrivateKey = encrypted
		}
		if publicKey, ok := secretData["public_key"]; ok {
			credential.PublicKey = publicKey
		}
	}

	// Save to database
	if err := s.repo.Create(credential); err != nil {
		return nil, err
	}

	return credential, nil
}

// GetCredential gets a Git credential
func (s *gitCredentialService) GetCredential(id uint, createdBy string) (*database.GitCredential, error) {
	return s.repo.GetByID(id, createdBy)
}

// ListCredentials gets credential list
func (s *gitCredentialService) ListCredentials(createdBy string, credType *database.GitCredentialType, page, pageSize int) ([]database.GitCredential, int64, error) {
	return s.repo.List(createdBy, credType, page, pageSize)
}

// UpdateCredential updates a credential
func (s *gitCredentialService) UpdateCredential(id uint, createdBy string, updates map[string]interface{}, secretData map[string]string) error {
	credential, err := s.repo.GetByID(id, createdBy)
	if err != nil {
		return err
	}

	// Update basic information
	if name, ok := updates["name"]; ok {
		credential.Name = name.(string)
	}
	if description, ok := updates["description"]; ok {
		credential.Description = description.(string)
	}
	if username, ok := updates["username"]; ok {
		credential.Username = username.(string)
	}

	if len(secretData) > 0 {
		switch credential.Type {
		case database.GitCredentialTypePassword, database.GitCredentialTypeToken:
			if password, ok := secretData["password"]; ok {
				encrypted, err := utils.EncryptAES(password, s.config.AESKey)
				if err != nil {
					return fmt.Errorf("failed to encrypt password: %v", err)
				}
				credential.PasswordHash = encrypted
			}
		case database.GitCredentialTypeSSHKey:
			if privateKey, ok := secretData["private_key"]; ok {
				encrypted, err := utils.EncryptAES(privateKey, s.config.AESKey)
				if err != nil {
					return fmt.Errorf("failed to encrypt private key: %v", err)
				}
				credential.PrivateKey = encrypted
			}
			if publicKey, ok := secretData["public_key"]; ok {
				credential.PublicKey = publicKey
			}
		}
	}

	return s.repo.Update(credential)
}

// DeleteCredential deletes a credential
func (s *gitCredentialService) DeleteCredential(id uint, createdBy string) error {
	// 检查凭据是否存在
	credential, err := s.repo.GetByID(id, createdBy)
	if err != nil {
		return err
	}

	// 检查是否有项目使用此凭据
	projects, _, err := s.projectRepo.List(createdBy, "", nil, 1, 1000)
	if err != nil {
		return fmt.Errorf("failed to check credential usage: %v", err)
	}

	for _, project := range projects {
		if project.CredentialID != nil && *project.CredentialID == credential.ID {
			return errors.New("git_credential.delete_used_by_projects")
		}
	}

	return s.repo.Delete(id, createdBy)
}

// ListActiveCredentials gets credential list (now returns all credentials since we no longer distinguish active status)
func (s *gitCredentialService) ListActiveCredentials(createdBy string, credType *database.GitCredentialType) ([]database.GitCredential, error) {
	// Use a large page size to get all credentials
	credentials, _, err := s.repo.List(createdBy, credType, 1, 1000)
	return credentials, err
}

// DecryptCredentialSecret decrypts credential sensitive information
func (s *gitCredentialService) DecryptCredentialSecret(credential *database.GitCredential, secretType string) (string, error) {
	switch secretType {
	case "password":
		if credential.PasswordHash == "" {
			return "", errors.New("password not set")
		}
		return utils.DecryptAES(credential.PasswordHash, s.config.AESKey)
	case "private_key":
		if credential.PrivateKey == "" {
			return "", errors.New("private key not set")
		}
		return utils.DecryptAES(credential.PrivateKey, s.config.AESKey)
	default:
		return "", errors.New("unsupported secret type")
	}
}

// ValidateCredentialData validates credential data
func (s *gitCredentialService) ValidateCredentialData(credType string, data map[string]string) error {
	switch database.GitCredentialType(credType) {
	case database.GitCredentialTypePassword:
		if data["password"] == "" {
			return errors.New("password is required for password type")
		}
	case database.GitCredentialTypeToken:
		if data["password"] == "" { // token stored in password field
			return errors.New("token is required for token type")
		}
	case database.GitCredentialTypeSSHKey:
		if data["private_key"] == "" {
			return errors.New("private key is required for SSH key type")
		}
		// Validate SSH key format
		if !strings.Contains(data["private_key"], "BEGIN") {
			return errors.New("invalid private key format")
		}
	default:
		return errors.New("unsupported credential type")
	}
	return nil
}
