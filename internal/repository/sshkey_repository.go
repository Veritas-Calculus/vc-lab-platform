// Package repository provides data access layer implementations.
package repository

import (
	"context"
	"errors"

	"github.com/Veritas-Calculus/vc-lab-platform/internal/model"
	"gorm.io/gorm"
)

// SSHKeyRepository defines the interface for SSH key operations.
type SSHKeyRepository interface {
	Create(ctx context.Context, sshKey *model.SSHKey) error
	GetByID(ctx context.Context, id string) (*model.SSHKey, error)
	List(ctx context.Context, offset, limit int) ([]*model.SSHKey, int64, error)
	Update(ctx context.Context, sshKey *model.SSHKey) error
	Delete(ctx context.Context, id string) error
	GetDefault(ctx context.Context) (*model.SSHKey, error)
	SetDefault(ctx context.Context, id string) error
}

type sshKeyRepository struct {
	db *gorm.DB
}

// NewSSHKeyRepository creates a new SSH key repository.
func NewSSHKeyRepository(db *gorm.DB) SSHKeyRepository {
	return &sshKeyRepository{db: db}
}

// Create creates a new SSH key.
func (r *sshKeyRepository) Create(ctx context.Context, sshKey *model.SSHKey) error {
	return r.db.WithContext(ctx).Create(sshKey).Error
}

// GetByID retrieves an SSH key by ID.
func (r *sshKeyRepository) GetByID(ctx context.Context, id string) (*model.SSHKey, error) {
	var sshKey model.SSHKey
	if err := r.db.WithContext(ctx).Preload("CreatedBy").First(&sshKey, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &sshKey, nil
}

// List retrieves SSH keys with pagination.
func (r *sshKeyRepository) List(ctx context.Context, offset, limit int) ([]*model.SSHKey, int64, error) {
	var sshKeys []*model.SSHKey
	var total int64

	query := r.db.WithContext(ctx).Model(&model.SSHKey{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("CreatedBy").Offset(offset).Limit(limit).Order("created_at DESC").Find(&sshKeys).Error; err != nil {
		return nil, 0, err
	}

	return sshKeys, total, nil
}

// Update updates an existing SSH key.
func (r *sshKeyRepository) Update(ctx context.Context, sshKey *model.SSHKey) error {
	return r.db.WithContext(ctx).Save(sshKey).Error
}

// Delete deletes an SSH key by ID.
func (r *sshKeyRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&model.SSHKey{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// GetDefault retrieves the default SSH key.
func (r *sshKeyRepository) GetDefault(ctx context.Context) (*model.SSHKey, error) {
	var sshKey model.SSHKey
	if err := r.db.WithContext(ctx).Preload("CreatedBy").First(&sshKey, "is_default = ?", true).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &sshKey, nil
}

// SetDefault sets an SSH key as the default.
func (r *sshKeyRepository) SetDefault(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Unset all defaults
		if err := tx.Model(&model.SSHKey{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
			return err
		}
		// Set the new default
		result := tx.Model(&model.SSHKey{}).Where("id = ?", id).Update("is_default", true)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}
