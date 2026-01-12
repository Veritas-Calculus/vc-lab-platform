// Package repository provides data access layer implementations.
package repository

import (
	"context"
	"errors"

	"github.com/Veritas-Calculus/vc-lab-platform/internal/model"
	"gorm.io/gorm"
)

// ProviderRepository defines the interface for provider config operations.
type ProviderRepository interface {
	Create(ctx context.Context, provider *model.ProviderConfig) error
	GetByID(ctx context.Context, id string) (*model.ProviderConfig, error)
	List(ctx context.Context, providerType string, offset, limit int) ([]*model.ProviderConfig, int64, error)
	Update(ctx context.Context, provider *model.ProviderConfig) error
	Delete(ctx context.Context, id string) error
}

type providerRepository struct {
	db *gorm.DB
}

// NewProviderRepository creates a new provider repository.
func NewProviderRepository(db *gorm.DB) ProviderRepository {
	return &providerRepository{db: db}
}

// Create creates a new provider config.
func (r *providerRepository) Create(ctx context.Context, provider *model.ProviderConfig) error {
	return r.db.WithContext(ctx).Create(provider).Error
}

// GetByID retrieves a provider config by ID.
func (r *providerRepository) GetByID(ctx context.Context, id string) (*model.ProviderConfig, error) {
	var provider model.ProviderConfig
	if err := r.db.WithContext(ctx).Preload("Credential").First(&provider, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &provider, nil
}

// List retrieves provider configs with optional filtering.
func (r *providerRepository) List(ctx context.Context, providerType string, offset, limit int) ([]*model.ProviderConfig, int64, error) {
	var providers []*model.ProviderConfig
	var total int64

	query := r.db.WithContext(ctx).Model(&model.ProviderConfig{})
	if providerType != "" {
		query = query.Where("type = ?", providerType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Credential").Offset(offset).Limit(limit).Order("created_at DESC").Find(&providers).Error; err != nil {
		return nil, 0, err
	}

	return providers, total, nil
}

// Update updates a provider config.
func (r *providerRepository) Update(ctx context.Context, provider *model.ProviderConfig) error {
	return r.db.WithContext(ctx).Save(provider).Error
}

// Delete soft deletes a provider config.
func (r *providerRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.ProviderConfig{}, "id = ?", id).Error
}
