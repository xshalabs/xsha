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
	repo   repository.GitCredentialRepository
	config *config.Config
}

// NewGitCredentialService 创建Git凭据服务实例
func NewGitCredentialService(repo repository.GitCredentialRepository, cfg *config.Config) GitCredentialService {
	return &gitCredentialService{
		repo:   repo,
		config: cfg,
	}
}

// CreateCredential 创建Git凭据
func (s *gitCredentialService) CreateCredential(name, description, credType, username, createdBy string, secretData map[string]string) (*database.GitCredential, error) {
	// 验证输入
	if err := s.ValidateCredentialData(credType, secretData); err != nil {
		return nil, err
	}

	// 检查名称是否已存在
	if existing, _ := s.repo.GetByName(name, createdBy); existing != nil {
		return nil, errors.New("credential name already exists")
	}

	// 创建凭据对象
	credential := &database.GitCredential{
		Name:        name,
		Description: description,
		Type:        database.GitCredentialType(credType),
		Username:    username,
		CreatedBy:   createdBy,
		IsActive:    true,
	}

	// 加密敏感数据
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

	// 保存到数据库
	if err := s.repo.Create(credential); err != nil {
		return nil, err
	}

	return credential, nil
}

// GetCredential 获取Git凭据
func (s *gitCredentialService) GetCredential(id uint, createdBy string) (*database.GitCredential, error) {
	return s.repo.GetByID(id, createdBy)
}

// ListCredentials 获取凭据列表
func (s *gitCredentialService) ListCredentials(createdBy string, credType *database.GitCredentialType, page, pageSize int) ([]database.GitCredential, int64, error) {
	return s.repo.List(createdBy, credType, page, pageSize)
}

// UpdateCredential 更新凭据
func (s *gitCredentialService) UpdateCredential(id uint, createdBy string, updates map[string]interface{}, secretData map[string]string) error {
	credential, err := s.repo.GetByID(id, createdBy)
	if err != nil {
		return err
	}

	// 更新基本信息
	if name, ok := updates["name"]; ok {
		credential.Name = name.(string)
	}
	if description, ok := updates["description"]; ok {
		credential.Description = description.(string)
	}
	if username, ok := updates["username"]; ok {
		credential.Username = username.(string)
	}

	// 更新敏感数据
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

// DeleteCredential 删除凭据
func (s *gitCredentialService) DeleteCredential(id uint, createdBy string) error {
	return s.repo.Delete(id, createdBy)
}

// ListActiveCredentials 获取激活的凭据列表
func (s *gitCredentialService) ListActiveCredentials(createdBy string, credType *database.GitCredentialType) ([]database.GitCredential, error) {
	return s.repo.ListActive(createdBy, credType)
}

// DecryptCredentialSecret 解密凭据敏感信息
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

// ValidateCredentialData 验证凭据数据
func (s *gitCredentialService) ValidateCredentialData(credType string, data map[string]string) error {
	switch database.GitCredentialType(credType) {
	case database.GitCredentialTypePassword:
		if data["password"] == "" {
			return errors.New("password is required for password type")
		}
	case database.GitCredentialTypeToken:
		if data["password"] == "" { // token存储在password字段
			return errors.New("token is required for token type")
		}
	case database.GitCredentialTypeSSHKey:
		if data["private_key"] == "" {
			return errors.New("private key is required for SSH key type")
		}
		// 验证SSH密钥格式
		if !strings.Contains(data["private_key"], "BEGIN") {
			return errors.New("invalid private key format")
		}
	default:
		return errors.New("unsupported credential type")
	}
	return nil
}
