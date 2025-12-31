// Package repository provides role data access.
package repository

import (
	"context"
	"errors"

	"github.com/Veritas-Calculus/vc-lab-platform/internal/model"
	"gorm.io/gorm"
)

// RoleRepository defines the interface for role data access.
type RoleRepository interface {
	Create(ctx context.Context, role *model.Role) error
	GetByID(ctx context.Context, id string) (*model.Role, error)
	GetByCode(ctx context.Context, code string) (*model.Role, error)
	Update(ctx context.Context, role *model.Role) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]*model.Role, int64, error)
	AddPermissions(ctx context.Context, roleID string, permissionIDs []string) error
	RemovePermissions(ctx context.Context, roleID string, permissionIDs []string) error
}

type roleRepository struct {
	db *gorm.DB
}

// NewRoleRepository creates a new role repository.
func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) Create(ctx context.Context, role *model.Role) error {
	result := r.db.WithContext(ctx).Create(role)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return ErrDuplicateKey
		}
		return result.Error
	}
	return nil
}

func (r *roleRepository) GetByID(ctx context.Context, id string) (*model.Role, error) {
	var role model.Role
	result := r.db.WithContext(ctx).Preload("Permissions").First(&role, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, result.Error
	}
	return &role, nil
}

func (r *roleRepository) GetByCode(ctx context.Context, code string) (*model.Role, error) {
	var role model.Role
	result := r.db.WithContext(ctx).Preload("Permissions").First(&role, "code = ?", code)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, result.Error
	}
	return &role, nil
}

func (r *roleRepository) Update(ctx context.Context, role *model.Role) error {
	result := r.db.WithContext(ctx).Save(role)
	return result.Error
}

func (r *roleRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&model.Role{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *roleRepository) List(ctx context.Context, offset, limit int) ([]*model.Role, int64, error) {
	var roles []*model.Role
	var total int64

	if err := r.db.WithContext(ctx).Model(&model.Role{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	result := r.db.WithContext(ctx).Preload("Permissions").Offset(offset).Limit(limit).Find(&roles)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return roles, total, nil
}

func (r *roleRepository) AddPermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	role, err := r.GetByID(ctx, roleID)
	if err != nil {
		return err
	}

	var permissions []*model.Permission
	if err := r.db.WithContext(ctx).Where("id IN ?", permissionIDs).Find(&permissions).Error; err != nil {
		return err
	}

	return r.db.WithContext(ctx).Model(role).Association("Permissions").Append(permissions)
}

func (r *roleRepository) RemovePermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	role, err := r.GetByID(ctx, roleID)
	if err != nil {
		return err
	}

	var permissions []*model.Permission
	if err := r.db.WithContext(ctx).Where("id IN ?", permissionIDs).Find(&permissions).Error; err != nil {
		return err
	}

	return r.db.WithContext(ctx).Model(role).Association("Permissions").Delete(permissions)
}
