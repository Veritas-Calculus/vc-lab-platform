// Package service provides business logic implementations.
package service

import (
	"context"
	"errors"

	"github.com/Veritas-Calculus/vc-lab-platform/internal/constants"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/model"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/repository"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// UserService provides user-related business operations.
type UserService interface {
	Create(ctx context.Context, input *CreateUserInput) (*model.User, error)
	GetByID(ctx context.Context, id string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	List(ctx context.Context, filters UserFilters, page, pageSize int) ([]*model.User, int64, error)
	Update(ctx context.Context, id string, updates map[string]interface{}) (*model.User, error)
	Delete(ctx context.Context, id string) error
	ChangePassword(ctx context.Context, id, oldPassword, newPassword string) error
	ResetPassword(ctx context.Context, id, newPassword string) error
	AssignRole(ctx context.Context, userID, roleID string) error
	RemoveRole(ctx context.Context, userID, roleID string) error
}

// userService implements UserService.
type userService struct {
	userRepo repository.UserRepository
	roleRepo repository.RoleRepository
	logger   *zap.Logger
}

// NewUserService creates a new user service.
func NewUserService(userRepo repository.UserRepository, roleRepo repository.RoleRepository, logger *zap.Logger) UserService {
	return &userService{
		userRepo: userRepo,
		roleRepo: roleRepo,
		logger:   logger,
	}
}

// CreateUserInput represents input for user creation.
type CreateUserInput struct {
	Username    string
	Email       string
	Password    string
	DisplayName string
	Avatar      string
	Phone       string
}

// UserFilters represents filters for user listing.
type UserFilters struct {
	Status *int
	Search string
}

// Create creates a new user.
func (s *userService) Create(ctx context.Context, input *CreateUserInput) (*model.User, error) {
	if input == nil {
		return nil, errors.New("input cannot be nil")
	}

	// Check for existing email
	existing, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("email already exists")
	}

	// Check for existing username
	existing, err = s.userRepo.GetByUsername(ctx, input.Username)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("username already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &model.User{
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
		DisplayName:  input.DisplayName,
		Avatar:       input.Avatar,
		Phone:        input.Phone,
		Status:       1, // Active
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error("failed to create user", zap.Error(err))
		return nil, errors.New("failed to create user")
	}

	return user, nil
}

// GetByID gets a user by ID.
func (s *userService) GetByID(ctx context.Context, id string) (*model.User, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, repository.ErrNotFound
		}
		s.logger.Error("failed to get user", zap.Error(err))
		return nil, errors.New("failed to get user")
	}

	return user, nil
}

// GetByEmail gets a user by email.
func (s *userService) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	if email == "" {
		return nil, errors.New("email cannot be empty")
	}

	return s.userRepo.GetByEmail(ctx, email)
}

// GetByUsername gets a user by username.
func (s *userService) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}

	return s.userRepo.GetByUsername(ctx, username)
}

// List lists users with filters and pagination.
func (s *userService) List(ctx context.Context, _ UserFilters, page, pageSize int) ([]*model.User, int64, error) {
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
	return s.userRepo.List(ctx, offset, pageSize)
}

// Update updates a user.
func (s *userService) Update(ctx context.Context, id string, updates map[string]interface{}) (*model.User, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	// Verify user exists
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	// Apply allowed updates
	if displayName, ok := updates["display_name"].(string); ok {
		user.DisplayName = displayName
	}
	if avatar, ok := updates["avatar"].(string); ok {
		user.Avatar = avatar
	}
	if phone, ok := updates["phone"].(string); ok {
		user.Phone = phone
	}
	if status, ok := updates["status"].(int8); ok {
		user.Status = status
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("failed to update user", zap.Error(err))
		return nil, errors.New("failed to update user")
	}

	return s.userRepo.GetByID(ctx, id)
}

// Delete deletes a user.
func (s *userService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}

	// Verify user exists
	_, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return repository.ErrNotFound
		}
		return err
	}

	if err := s.userRepo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete user", zap.Error(err))
		return errors.New("failed to delete user")
	}

	return nil
}

// ChangePassword changes a user's password.
func (s *userService) ChangePassword(ctx context.Context, id, oldPassword, newPassword string) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}
	if oldPassword == "" {
		return errors.New("old password cannot be empty")
	}
	if newPassword == "" {
		return errors.New("new password cannot be empty")
	}
	if len(newPassword) < constants.MinPasswordLength {
		return errors.New("password must be at least 8 characters")
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return repository.ErrNotFound
		}
		return err
	}

	// Verify old password
	if pwdErr := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); pwdErr != nil {
		return errors.New("incorrect password")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	user.PasswordHash = string(hashedPassword)
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("failed to update password", zap.Error(err))
		return errors.New("failed to update password")
	}

	return nil
}

// ResetPassword resets a user's password (admin function).
func (s *userService) ResetPassword(ctx context.Context, id, newPassword string) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}
	if newPassword == "" {
		return errors.New("new password cannot be empty")
	}
	if len(newPassword) < constants.MinPasswordLength {
		return errors.New("password must be at least 8 characters")
	}

	// Verify user exists
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return repository.ErrNotFound
		}
		return err
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	user.PasswordHash = string(hashedPassword)
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("failed to reset password", zap.Error(err))
		return errors.New("failed to reset password")
	}

	return nil
}

// AssignRole assigns a role to a user.
func (s *userService) AssignRole(ctx context.Context, userID, roleID string) error {
	if userID == "" {
		return errors.New("user ID cannot be empty")
	}
	if roleID == "" {
		return errors.New("role ID cannot be empty")
	}

	// Verify user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify role exists
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return err
	}

	// Check if already assigned
	for _, r := range user.Roles {
		if r.ID == roleID {
			return nil // Already assigned
		}
	}

	// Assign role
	user.Roles = append(user.Roles, *role)
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("failed to assign role", zap.Error(err))
		return errors.New("failed to assign role")
	}

	return nil
}

// RemoveRole removes a role from a user.
func (s *userService) RemoveRole(ctx context.Context, userID, roleID string) error {
	if userID == "" {
		return errors.New("user ID cannot be empty")
	}
	if roleID == "" {
		return errors.New("role ID cannot be empty")
	}

	// Verify user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Remove role
	var newRoles []model.Role
	for _, r := range user.Roles {
		if r.ID != roleID {
			newRoles = append(newRoles, r)
		}
	}

	user.Roles = newRoles
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("failed to remove role", zap.Error(err))
		return errors.New("failed to remove role")
	}

	return nil
}
