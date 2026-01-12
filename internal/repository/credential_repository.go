// Package repository provides data access layer implementations.
package repository

import (
	"context"
	"errors"

	"github.com/Veritas-Calculus/vc-lab-platform/internal/model"
	"gorm.io/gorm"
)

// CredentialRepository defines the interface for credential operations.
type CredentialRepository interface {
	Create(ctx context.Context, credential *model.Credential) error
	GetByID(ctx context.Context, id string) (*model.Credential, error)
	List(ctx context.Context, credentialType string, offset, limit int) ([]*model.Credential, int64, error)
	Update(ctx context.Context, credential *model.Credential) error
	Delete(ctx context.Context, id string) error
}

type credentialRepository struct {
	db *gorm.DB
}

// NewCredentialRepository creates a new credential repository.
func NewCredentialRepository(db *gorm.DB) CredentialRepository {
	return &credentialRepository{db: db}
}

// Create creates a new credential.
func (r *credentialRepository) Create(ctx context.Context, credential *model.Credential) error {
	return r.db.WithContext(ctx).Create(credential).Error
}

// GetByID retrieves a credential by ID.
func (r *credentialRepository) GetByID(ctx context.Context, id string) (*model.Credential, error) {
	var credential model.Credential
	if err := r.db.WithContext(ctx).Preload("Zone").Preload("Provider").Preload("CreatedBy").First(&credential, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &credential, nil
}

// List retrieves credentials with optional filtering.
func (r *credentialRepository) List(ctx context.Context, credentialType string, offset, limit int) ([]*model.Credential, int64, error) {
	var credentials []*model.Credential
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Credential{})
	if credentialType != "" {
		query = query.Where("type = ?", credentialType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Zone").Preload("Provider").Preload("CreatedBy").Offset(offset).Limit(limit).Order("created_at DESC").Find(&credentials).Error; err != nil {
		return nil, 0, err
	}

	return credentials, total, nil
}

// Update updates a credential.
func (r *credentialRepository) Update(ctx context.Context, credential *model.Credential) error {
	return r.db.WithContext(ctx).Save(credential).Error
}

// Delete soft deletes a credential.
func (r *credentialRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.Credential{}, "id = ?", id).Error
}
