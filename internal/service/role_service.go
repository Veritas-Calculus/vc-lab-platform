// Package service provides business logic implementations.
package service

import (
	"context"
	"errors"

	"github.com/Veritas-Calculus/vc-lab-platform/internal/model"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/repository"
	"go.uber.org/zap"
)

// RoleService provides role-related business operations.
type RoleService interface {
	Create(ctx context.Context, input *CreateRoleInput) (*model.Role, error)
	GetByID(ctx context.Context, id string) (*model.Role, error)
	GetByCode(ctx context.Context, code string) (*model.Role, error)
	List(ctx context.Context, page, pageSize int) ([]*model.Role, int64, error)
	Update(ctx context.Context, id string, updates map[string]interface{}) (*model.Role, error)
	Delete(ctx context.Context, id string) error
}

// roleService implements RoleService.
type roleService struct {
	roleRepo repository.RoleRepository
	logger   *zap.Logger
}

// NewRoleService creates a new role service.
func NewRoleService(roleRepo repository.RoleRepository, logger *zap.Logger) RoleService {
	return &roleService{
		roleRepo: roleRepo,
		logger:   logger,
	}
}

// CreateRoleInput represents input for role creation.
type CreateRoleInput struct {
	Name        string
	Code        string
	Description string
}

// Create creates a new role.
func (s *roleService) Create(ctx context.Context, input *CreateRoleInput) (*model.Role, error) {
	if input == nil {
		return nil, errors.New("input cannot be nil")
	}

	// Check for existing code
	existing, err := s.roleRepo.GetByCode(ctx, input.Code)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("role code already exists")
	}

	role := &model.Role{
		Name:        input.Name,
		Code:        input.Code,
		Description: input.Description,
		Status:      1, // Active
	}

	if err := s.roleRepo.Create(ctx, role); err != nil {
		s.logger.Error("failed to create role", zap.Error(err))
		return nil, err
	}

	return role, nil
}

// GetByID retrieves a role by ID.
func (s *roleService) GetByID(ctx context.Context, id string) (*model.Role, error) {
	return s.roleRepo.GetByID(ctx, id)
}

// GetByCode retrieves a role by code.
func (s *roleService) GetByCode(ctx context.Context, code string) (*model.Role, error) {
	return s.roleRepo.GetByCode(ctx, code)
}

// List returns a paginated list of roles.
func (s *roleService) List(ctx context.Context, page, pageSize int) ([]*model.Role, int64, error) {
	offset := (page - 1) * pageSize
	return s.roleRepo.List(ctx, offset, pageSize)
}

// Update updates a role with the given updates.
func (s *roleService) Update(ctx context.Context, id string, updates map[string]interface{}) (*model.Role, error) {
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok && name != "" {
		role.Name = name
	}
	if description, ok := updates["description"].(string); ok {
		role.Description = description
	}
	if status, ok := updates["status"].(int8); ok {
		role.Status = status
	}

	if err := s.roleRepo.Update(ctx, role); err != nil {
		s.logger.Error("failed to update role", zap.Error(err))
		return nil, err
	}

	return role, nil
}

// Delete deletes a role by ID.
func (s *roleService) Delete(ctx context.Context, id string) error {
	// Check if role exists
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Prevent deletion of system roles
	if role.IsSystem {
		return errors.New("cannot delete system role")
	}

	return s.roleRepo.Delete(ctx, id)
}
