// Package service provides business logic implementations.
package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/Veritas-Calculus/vc-lab-platform/internal/model"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/repository"
	"go.uber.org/zap"
)

// SSHKeyService defines the interface for SSH key operations.
type SSHKeyService interface {
	List(ctx context.Context, page, pageSize int) ([]*model.SSHKey, int64, error)
	Get(ctx context.Context, id string) (*model.SSHKey, error)
	GetDefault(ctx context.Context) (*model.SSHKey, error)
	Create(ctx context.Context, input *CreateSSHKeyInput) (*model.SSHKey, error)
	Update(ctx context.Context, id string, input *UpdateSSHKeyInput) (*model.SSHKey, error)
	Delete(ctx context.Context, id string) error
	SetDefault(ctx context.Context, id string) error
}

// CreateSSHKeyInput represents input for creating an SSH key.
type CreateSSHKeyInput struct {
	Name        string
	PublicKey   string
	Description string
	CreatedByID string
	IsDefault   bool
}

// UpdateSSHKeyInput represents input for updating an SSH key.
type UpdateSSHKeyInput struct {
	Name        *string
	PublicKey   *string
	Description *string
	IsDefault   *bool
}

type sshKeyService struct {
	repo   repository.SSHKeyRepository
	logger *zap.Logger
}

// NewSSHKeyService creates a new SSH key service.
func NewSSHKeyService(repo repository.SSHKeyRepository, logger *zap.Logger) SSHKeyService {
	return &sshKeyService{
		repo:   repo,
		logger: logger,
	}
}

// List retrieves SSH keys with pagination.
func (s *sshKeyService) List(ctx context.Context, page, pageSize int) ([]*model.SSHKey, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, offset, pageSize)
}

// Get retrieves an SSH key by ID.
func (s *sshKeyService) Get(ctx context.Context, id string) (*model.SSHKey, error) {
	return s.repo.GetByID(ctx, id)
}

// GetDefault retrieves the default SSH key.
func (s *sshKeyService) GetDefault(ctx context.Context) (*model.SSHKey, error) {
	return s.repo.GetDefault(ctx)
}

// Create creates a new SSH key.
func (s *sshKeyService) Create(ctx context.Context, input *CreateSSHKeyInput) (*model.SSHKey, error) {
	// Validate the public key format
	if err := validateSSHPublicKey(input.PublicKey); err != nil {
		return nil, err
	}

	// Calculate fingerprint
	fingerprint, err := calculateSSHFingerprint(input.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate fingerprint: %w", err)
	}

	sshKey := &model.SSHKey{
		Name:        input.Name,
		PublicKey:   input.PublicKey,
		Fingerprint: fingerprint,
		Description: input.Description,
		CreatedByID: input.CreatedByID,
		IsDefault:   input.IsDefault,
		Status:      1, // 1: active
	}

	if err := s.repo.Create(ctx, sshKey); err != nil {
		return nil, fmt.Errorf("failed to create SSH key: %w", err)
	}

	// If this key should be default, set it
	if input.IsDefault {
		if err := s.repo.SetDefault(ctx, sshKey.ID); err != nil {
			s.logger.Warn("Failed to set SSH key as default", zap.Error(err))
		}
	}

	return sshKey, nil
}

// Update updates an existing SSH key.
func (s *sshKeyService) Update(ctx context.Context, id string, input *UpdateSSHKeyInput) (*model.SSHKey, error) {
	sshKey, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		sshKey.Name = *input.Name
	}
	if input.PublicKey != nil {
		if err := validateSSHPublicKey(*input.PublicKey); err != nil {
			return nil, err
		}
		fingerprint, err := calculateSSHFingerprint(*input.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate fingerprint: %w", err)
		}
		sshKey.PublicKey = *input.PublicKey
		sshKey.Fingerprint = fingerprint
	}
	if input.Description != nil {
		sshKey.Description = *input.Description
	}
	if input.IsDefault != nil {
		sshKey.IsDefault = *input.IsDefault
	}

	if err := s.repo.Update(ctx, sshKey); err != nil {
		return nil, fmt.Errorf("failed to update SSH key: %w", err)
	}

	// If this key should be default, set it
	if input.IsDefault != nil && *input.IsDefault {
		if err := s.repo.SetDefault(ctx, sshKey.ID); err != nil {
			s.logger.Warn("Failed to set SSH key as default", zap.Error(err))
		}
	}

	return sshKey, nil
}

// Delete deletes an SSH key.
func (s *sshKeyService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// SetDefault sets an SSH key as the default.
func (s *sshKeyService) SetDefault(ctx context.Context, id string) error {
	return s.repo.SetDefault(ctx, id)
}

// validateSSHPublicKey validates the format of an SSH public key.
func validateSSHPublicKey(publicKey string) error {
	parts := strings.Fields(publicKey)
	if len(parts) < 2 {
		return errors.New("invalid SSH public key format")
	}

	keyType := parts[0]
	validTypes := []string{"ssh-rsa", "ssh-ed25519", "ecdsa-sha2-nistp256", "ecdsa-sha2-nistp384", "ecdsa-sha2-nistp521", "ssh-dss"}
	isValidType := false
	for _, vt := range validTypes {
		if keyType == vt {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return errors.New("unsupported SSH key type")
	}

	// Try to decode the key data
	_, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return errors.New("invalid SSH public key encoding")
	}

	return nil
}

// calculateSSHFingerprint calculates the SHA256 fingerprint of an SSH public key.
func calculateSSHFingerprint(publicKey string) (string, error) {
	parts := strings.Fields(publicKey)
	if len(parts) < 2 {
		return "", errors.New("invalid SSH public key format")
	}

	keyData, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", errors.New("invalid SSH public key encoding")
	}

	hash := sha256.Sum256(keyData)
	fingerprint := base64.StdEncoding.EncodeToString(hash[:])
	return "SHA256:" + fingerprint, nil
}
