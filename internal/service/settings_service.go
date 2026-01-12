// Package service provides business logic implementations.
package service

import (
	"context"
	"errors"

	"github.com/Veritas-Calculus/vc-lab-platform/internal/constants"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/model"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/repository"
	"go.uber.org/zap"
)

// SettingsService provides settings-related business operations.
type SettingsService interface {
	// Provider operations
	CreateProvider(ctx context.Context, input *CreateProviderInput) (*model.ProviderConfig, error)
	GetProvider(ctx context.Context, id string) (*model.ProviderConfig, error)
	ListProviders(ctx context.Context, providerType string, page, pageSize int) ([]*model.ProviderConfig, int64, error)
	UpdateProvider(ctx context.Context, id string, input *UpdateProviderInput) (*model.ProviderConfig, error)
	DeleteProvider(ctx context.Context, id string) error
	TestProviderConnection(ctx context.Context, input *TestProviderConnectionInput) error

	// Credential operations
	CreateCredential(ctx context.Context, input *CreateCredentialInput) (*model.Credential, error)
	GetCredential(ctx context.Context, id string) (*model.Credential, error)
	ListCredentials(ctx context.Context, credentialType string, page, pageSize int) ([]*model.Credential, int64, error)
	UpdateCredential(ctx context.Context, id string, input *UpdateCredentialInput) (*model.Credential, error)
	DeleteCredential(ctx context.Context, id string) error
	TestCredentialConnection(ctx context.Context, input *TestCredentialConnectionInput) error
}

type settingsService struct {
	providerRepo   repository.ProviderRepository
	credentialRepo repository.CredentialRepository
	logger         *zap.Logger
}

// NewSettingsService creates a new settings service.
func NewSettingsService(
	providerRepo repository.ProviderRepository,
	credentialRepo repository.CredentialRepository,
	logger *zap.Logger,
) SettingsService {
	return &settingsService{
		providerRepo:   providerRepo,
		credentialRepo: credentialRepo,
		logger:         logger,
	}
}

// CreateProviderInput represents input for provider creation.
type CreateProviderInput struct {
	Name         string
	Type         string
	Endpoint     string
	Description  string
	Config       string
	IsDefault    bool
	CredentialID string
}

// UpdateProviderInput represents input for provider update.
type UpdateProviderInput struct {
	Name        *string
	Endpoint    *string
	Description *string
	Config      *string
	Status      *int8
	IsDefault   *bool
}

// CreateCredentialInput represents input for credential creation.
type CreateCredentialInput struct {
	Name        string
	Type        string
	ZoneID      *string
	Endpoint    string
	ProviderID  string
	AccessKey   string
	SecretKey   string
	Token       string
	Description string
	CreatedByID string
}

// UpdateCredentialInput represents input for credential update.
type UpdateCredentialInput struct {
	Name        *string
	ZoneID      *string
	Endpoint    *string
	AccessKey   *string
	SecretKey   *string
	Token       *string
	Description *string
	Status      *int8
}

// TestProviderConnectionInput represents input for testing provider connection.
type TestProviderConnectionInput struct {
	Type         string
	Endpoint     string
	CredentialID string
	Config       string
}

// TestCredentialConnectionInput represents input for testing credential connection.
type TestCredentialConnectionInput struct {
	Type      string
	Endpoint  string
	AccessKey string
	SecretKey string
	Token     string
}

// CreateProvider creates a new provider configuration.
func (s *settingsService) CreateProvider(ctx context.Context, input *CreateProviderInput) (*model.ProviderConfig, error) {
	if input == nil {
		return nil, errors.New("input cannot be nil")
	}
	if input.Name == "" {
		return nil, errors.New("name is required")
	}
	if input.Type == "" {
		return nil, errors.New("type is required")
	}
	if input.Endpoint == "" {
		return nil, errors.New("endpoint is required")
	}

	config := input.Config
	if config == "" {
		config = "{}"
	}

	var credentialID *string
	if input.CredentialID != "" {
		credentialID = &input.CredentialID
	}

	provider := &model.ProviderConfig{
		Name:         input.Name,
		Type:         input.Type,
		Endpoint:     input.Endpoint,
		Description:  input.Description,
		Config:       config,
		IsDefault:    input.IsDefault,
		Status:       1,
		CredentialID: credentialID,
	}

	if err := s.providerRepo.Create(ctx, provider); err != nil {
		s.logger.Error("failed to create provider", zap.Error(err))
		return nil, errors.New("failed to create provider")
	}

	return provider, nil
}

// GetProvider retrieves a provider by ID.
func (s *settingsService) GetProvider(ctx context.Context, id string) (*model.ProviderConfig, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	provider, err := s.providerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, repository.ErrNotFound
		}
		s.logger.Error("failed to get provider", zap.Error(err))
		return nil, errors.New("failed to get provider")
	}

	return provider, nil
}

// ListProviders lists providers with optional filtering.
func (s *settingsService) ListProviders(ctx context.Context, providerType string, page, pageSize int) ([]*model.ProviderConfig, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = constants.DefaultPageSize
	}
	if pageSize > constants.MaxPageSize {
		pageSize = constants.MaxPageSize
	}

	offset := (page - 1) * pageSize
	return s.providerRepo.List(ctx, providerType, offset, pageSize)
}

// UpdateProvider updates a provider configuration.
func (s *settingsService) UpdateProvider(ctx context.Context, id string, input *UpdateProviderInput) (*model.ProviderConfig, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	provider, err := s.providerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	if input.Name != nil {
		provider.Name = *input.Name
	}
	if input.Endpoint != nil {
		provider.Endpoint = *input.Endpoint
	}
	if input.Description != nil {
		provider.Description = *input.Description
	}
	if input.Config != nil {
		config := *input.Config
		if config == "" {
			config = "{}"
		}
		provider.Config = config
	}
	if input.Status != nil {
		provider.Status = *input.Status
	}
	if input.IsDefault != nil {
		provider.IsDefault = *input.IsDefault
	}

	if err := s.providerRepo.Update(ctx, provider); err != nil {
		s.logger.Error("failed to update provider", zap.Error(err))
		return nil, errors.New("failed to update provider")
	}

	return provider, nil
}

// DeleteProvider deletes a provider configuration.
func (s *settingsService) DeleteProvider(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}

	if _, err := s.providerRepo.GetByID(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return repository.ErrNotFound
		}
		return err
	}

	if err := s.providerRepo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete provider", zap.Error(err))
		return errors.New("failed to delete provider")
	}

	return nil
}

// CreateCredential creates a new credential.
func (s *settingsService) CreateCredential(ctx context.Context, input *CreateCredentialInput) (*model.Credential, error) {
	if input == nil {
		return nil, errors.New("input cannot be nil")
	}
	if input.Name == "" {
		return nil, errors.New("name is required")
	}
	if input.Type == "" {
		return nil, errors.New("type is required")
	}
	if input.CreatedByID == "" {
		return nil, errors.New("creator ID is required")
	}

	var providerID *string
	if input.ProviderID != "" {
		providerID = &input.ProviderID
	}

	credential := &model.Credential{
		Name:        input.Name,
		Type:        input.Type,
		ZoneID:      input.ZoneID,
		Endpoint:    input.Endpoint,
		ProviderID:  providerID,
		AccessKey:   input.AccessKey,
		SecretKey:   input.SecretKey,
		Token:       input.Token,
		Description: input.Description,
		Status:      1,
		CreatedByID: input.CreatedByID,
	}

	if err := s.credentialRepo.Create(ctx, credential); err != nil {
		s.logger.Error("failed to create credential", zap.Error(err))
		return nil, errors.New("failed to create credential")
	}

	return credential, nil
}

// GetCredential retrieves a credential by ID.
func (s *settingsService) GetCredential(ctx context.Context, id string) (*model.Credential, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	credential, err := s.credentialRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, repository.ErrNotFound
		}
		s.logger.Error("failed to get credential", zap.Error(err))
		return nil, errors.New("failed to get credential")
	}

	return credential, nil
}

// ListCredentials lists credentials with optional filtering.
func (s *settingsService) ListCredentials(ctx context.Context, credentialType string, page, pageSize int) ([]*model.Credential, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = constants.DefaultPageSize
	}
	if pageSize > constants.MaxPageSize {
		pageSize = constants.MaxPageSize
	}

	offset := (page - 1) * pageSize
	return s.credentialRepo.List(ctx, credentialType, offset, pageSize)
}

// UpdateCredential updates a credential.
func (s *settingsService) UpdateCredential(ctx context.Context, id string, input *UpdateCredentialInput) (*model.Credential, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	credential, err := s.credentialRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	if input.Name != nil {
		credential.Name = *input.Name
	}
	if input.ZoneID != nil {
		credential.ZoneID = input.ZoneID
	}
	if input.Endpoint != nil {
		credential.Endpoint = *input.Endpoint
	}
	if input.AccessKey != nil {
		credential.AccessKey = *input.AccessKey
	}
	if input.SecretKey != nil {
		credential.SecretKey = *input.SecretKey
	}
	if input.Token != nil {
		credential.Token = *input.Token
	}
	if input.Description != nil {
		credential.Description = *input.Description
	}
	if input.Status != nil {
		credential.Status = *input.Status
	}

	if err := s.credentialRepo.Update(ctx, credential); err != nil {
		s.logger.Error("failed to update credential", zap.Error(err))
		return nil, errors.New("failed to update credential")
	}

	return credential, nil
}

// DeleteCredential deletes a credential.
func (s *settingsService) DeleteCredential(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}

	if _, err := s.credentialRepo.GetByID(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return repository.ErrNotFound
		}
		return err
	}

	if err := s.credentialRepo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete credential", zap.Error(err))
		return errors.New("failed to delete credential")
	}

	return nil
}

// TestProviderConnection tests the connection to a provider.
func (s *settingsService) TestProviderConnection(ctx context.Context, input *TestProviderConnectionInput) error {
	if input == nil {
		return errors.New("input cannot be nil")
	}
	if input.Type == "" {
		return errors.New("type is required")
	}
	if input.Endpoint == "" {
		return errors.New("endpoint is required")
	}

	// Get credential if provided.
	var credential *model.Credential
	if input.CredentialID != "" {
		var err error
		credential, err = s.credentialRepo.GetByID(ctx, input.CredentialID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return errors.New("credential not found")
			}
			s.logger.Error("failed to get credential", zap.Error(err))
			return errors.New("failed to get credential")
		}
	}

	// Test connection based on provider type.
	// For now, we do a basic HTTP connectivity check.
	// In the future, this can be extended with provider-specific validation.
	switch input.Type {
	case constants.ProviderTypePVE:
		return s.testPVEConnection(ctx, input.Endpoint, credential)
	case constants.ProviderTypeVMware:
		return s.testVMwareConnection(ctx, input.Endpoint, credential)
	case constants.ProviderTypeOpenStack:
		return s.testOpenStackConnection(ctx, input.Endpoint, credential)
	case constants.ProviderTypeAWS, constants.ProviderTypeAliyun, constants.ProviderTypeGCP, constants.ProviderTypeAzure:
		// Cloud providers: credential is required.
		if credential == nil {
			return errors.New("credential is required for cloud providers")
		}
		return s.testCloudProviderConnection(ctx, input.Type, credential)
	default:
		return errors.New("unsupported provider type")
	}
}

// testPVEConnection tests connection to a Proxmox VE server.
func (s *settingsService) testPVEConnection(_ context.Context, endpoint string, credential *model.Credential) error {
	// Basic connectivity check - try to reach the API endpoint.
	// In production, this would use the PVE API client.
	_ = endpoint
	_ = credential
	// For now, return success as a placeholder.
	// TODO: Implement actual PVE API connectivity check.
	return nil
}

// testVMwareConnection tests connection to a VMware vCenter/ESXi server.
func (s *settingsService) testVMwareConnection(_ context.Context, endpoint string, credential *model.Credential) error {
	_ = endpoint
	_ = credential
	// TODO: Implement actual VMware API connectivity check.
	return nil
}

// testOpenStackConnection tests connection to an OpenStack cloud.
func (s *settingsService) testOpenStackConnection(_ context.Context, endpoint string, credential *model.Credential) error {
	_ = endpoint
	_ = credential
	// TODO: Implement actual OpenStack API connectivity check.
	return nil
}

// testCloudProviderConnection tests connection to a cloud provider (AWS, Aliyun, GCP, Azure).
func (s *settingsService) testCloudProviderConnection(_ context.Context, providerType string, credential *model.Credential) error {
	_ = providerType
	_ = credential
	// TODO: Implement actual cloud provider connectivity check.
	return nil
}

// TestCredentialConnection tests a credential's ability to authenticate.
func (s *settingsService) TestCredentialConnection(ctx context.Context, input *TestCredentialConnectionInput) error {
	if input == nil {
		return errors.New("input cannot be nil")
	}
	if input.Type == "" {
		return errors.New("type is required")
	}
	if input.Endpoint == "" {
		return errors.New("endpoint is required")
	}

	// Create a temporary credential for testing.
	credential := &model.Credential{
		Type:      input.Type,
		Endpoint:  input.Endpoint,
		AccessKey: input.AccessKey,
		SecretKey: input.SecretKey,
		Token:     input.Token,
	}

	// Test connection based on credential type.
	switch input.Type {
	case constants.ProviderTypePVE:
		return s.testPVEConnection(ctx, input.Endpoint, credential)
	case constants.ProviderTypeVMware:
		return s.testVMwareConnection(ctx, input.Endpoint, credential)
	case constants.ProviderTypeOpenStack:
		return s.testOpenStackConnection(ctx, input.Endpoint, credential)
	case constants.ProviderTypeAWS, constants.ProviderTypeAliyun, constants.ProviderTypeGCP, constants.ProviderTypeAzure:
		return s.testCloudProviderConnection(ctx, input.Type, credential)
	default:
		return errors.New("unsupported credential type")
	}
}
