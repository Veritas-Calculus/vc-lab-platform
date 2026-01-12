// Package repository provides data access layer implementations.
package repository

import (
	"context"
	"errors"

	"github.com/Veritas-Calculus/vc-lab-platform/internal/model"
	"gorm.io/gorm"
)

// VMTemplateRepository defines the interface for VM template operations.
type VMTemplateRepository interface {
	Create(ctx context.Context, template *model.VMTemplate) error
	GetByID(ctx context.Context, id string) (*model.VMTemplate, error)
	GetByName(ctx context.Context, templateName, provider string) (*model.VMTemplate, error)
	List(ctx context.Context, provider, osType, zoneID string, offset, limit int) ([]*model.VMTemplate, int64, error)
	Update(ctx context.Context, template *model.VMTemplate) error
	Delete(ctx context.Context, id string) error
	ListByProvider(ctx context.Context, provider string) ([]*model.VMTemplate, error)
}

type vmTemplateRepository struct {
	db *gorm.DB
}

// NewVMTemplateRepository creates a new VM template repository.
func NewVMTemplateRepository(db *gorm.DB) VMTemplateRepository {
	return &vmTemplateRepository{db: db}
}

// Create creates a new VM template.
func (r *vmTemplateRepository) Create(ctx context.Context, template *model.VMTemplate) error {
	return r.db.WithContext(ctx).Create(template).Error
}

// GetByID retrieves a VM template by ID.
func (r *vmTemplateRepository) GetByID(ctx context.Context, id string) (*model.VMTemplate, error) {
	var template model.VMTemplate
	if err := r.db.WithContext(ctx).Preload("Zone").First(&template, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &template, nil
}

// GetByName retrieves a VM template by template name and provider.
func (r *vmTemplateRepository) GetByName(ctx context.Context, templateName, provider string) (*model.VMTemplate, error) {
	var template model.VMTemplate
	if err := r.db.WithContext(ctx).Preload("Zone").First(&template, "template_name = ? AND provider = ?", templateName, provider).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &template, nil
}

// List retrieves VM templates with optional filtering.
func (r *vmTemplateRepository) List(ctx context.Context, provider, osType, zoneID string, offset, limit int) ([]*model.VMTemplate, int64, error) {
	var templates []*model.VMTemplate
	var total int64

	query := r.db.WithContext(ctx).Model(&model.VMTemplate{})
	if provider != "" {
		query = query.Where("provider = ?", provider)
	}
	if osType != "" {
		query = query.Where("os_type = ?", osType)
	}
	if zoneID != "" {
		query = query.Where("zone_id = ?", zoneID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Zone").Offset(offset).Limit(limit).Order("name ASC").Find(&templates).Error; err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

// Update updates an existing VM template.
func (r *vmTemplateRepository) Update(ctx context.Context, template *model.VMTemplate) error {
	return r.db.WithContext(ctx).Save(template).Error
}

// Delete deletes a VM template by ID.
func (r *vmTemplateRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&model.VMTemplate{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ListByProvider retrieves all VM templates for a specific provider.
func (r *vmTemplateRepository) ListByProvider(ctx context.Context, provider string) ([]*model.VMTemplate, error) {
	var templates []*model.VMTemplate
	if err := r.db.WithContext(ctx).Preload("Zone").Where("provider = ? AND status = ?", provider, "active").Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}
