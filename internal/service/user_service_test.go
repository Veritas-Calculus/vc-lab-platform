// Package service provides user service tests.
package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Veritas-Calculus/vc-lab-platform/internal/model"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// MockRoleRepository is a mock implementation of RoleRepository.
type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) Create(ctx context.Context, role *model.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepository) GetByID(ctx context.Context, id string) (*model.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	role, ok := args.Get(0).(*model.Role)
	if !ok {
		return nil, args.Error(1)
	}
	return role, args.Error(1)
}

func (m *MockRoleRepository) GetByCode(ctx context.Context, code string) (*model.Role, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	role, ok := args.Get(0).(*model.Role)
	if !ok {
		return nil, args.Error(1)
	}
	return role, args.Error(1)
}

func (m *MockRoleRepository) Update(ctx context.Context, role *model.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRoleRepository) List(ctx context.Context, offset, limit int) ([]*model.Role, int64, error) {
	args := m.Called(ctx, offset, limit)
	roles, ok := args.Get(0).([]*model.Role)
	if !ok {
		return nil, 0, args.Error(2)
	}
	total, ok := args.Get(1).(int64)
	if !ok {
		return roles, 0, args.Error(2)
	}
	return roles, total, args.Error(2)
}

func (m *MockRoleRepository) AddPermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	args := m.Called(ctx, roleID, permissionIDs)
	return args.Error(0)
}

func (m *MockRoleRepository) RemovePermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	args := m.Called(ctx, roleID, permissionIDs)
	return args.Error(0)
}

// Ensure MockRoleRepository implements repository.RoleRepository.
var _ repository.RoleRepository = (*MockRoleRepository)(nil)

func TestUserService_Create(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name      string
		input     *CreateUserInput
		setupMock func(*MockUserRepository, *MockRoleRepository)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "successful user creation",
			input: &CreateUserInput{
				Username:    "newuser",
				Email:       "new@example.com",
				Password:    "password123",
				DisplayName: "New User",
			},
			setupMock: func(ur *MockUserRepository, _ *MockRoleRepository) {
				ur.On("GetByEmail", mock.Anything, "new@example.com").Return(nil, repository.ErrNotFound)
				ur.On("GetByUsername", mock.Anything, "newuser").Return(nil, repository.ErrNotFound)
				ur.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "email already exists",
			input: &CreateUserInput{
				Username: "newuser",
				Email:    "existing@example.com",
				Password: "password123",
			},
			setupMock: func(ur *MockUserRepository, _ *MockRoleRepository) {
				ur.On("GetByEmail", mock.Anything, "existing@example.com").Return(&model.User{
					BaseModel: model.BaseModel{ID: "existing-id"},
					Email:     "existing@example.com",
				}, nil)
			},
			wantErr: true,
			errMsg:  "email already exists",
		},
		{
			name: "username already exists",
			input: &CreateUserInput{
				Username: "existinguser",
				Email:    "new@example.com",
				Password: "password123",
			},
			setupMock: func(ur *MockUserRepository, _ *MockRoleRepository) {
				ur.On("GetByEmail", mock.Anything, "new@example.com").Return(nil, repository.ErrNotFound)
				ur.On("GetByUsername", mock.Anything, "existinguser").Return(&model.User{
					BaseModel: model.BaseModel{ID: "existing-id"},
					Username:  "existinguser",
				}, nil)
			},
			wantErr: true,
			errMsg:  "username already exists",
		},
		{
			name:      "nil input",
			input:     nil,
			setupMock: func(_ *MockUserRepository, _ *MockRoleRepository) {},
			wantErr:   true,
			errMsg:    "input cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := new(MockUserRepository)
			mockRoleRepo := new(MockRoleRepository)
			tt.setupMock(mockUserRepo, mockRoleRepo)

			svc := NewUserService(mockUserRepo, mockRoleRepo, logger)

			user, err := svc.Create(t.Context(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
			}

			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_GetByID(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name      string
		userID    string
		setupMock func(*MockUserRepository)
		wantErr   bool
	}{
		{
			name:   "successful get",
			userID: "user-123",
			setupMock: func(m *MockUserRepository) {
				m.On("GetByID", mock.Anything, "user-123").Return(&model.User{
					BaseModel: model.BaseModel{ID: "user-123"},
					Username:  "testuser",
				}, nil)
			},
			wantErr: false,
		},
		{
			name:   "user not found",
			userID: "nonexistent",
			setupMock: func(m *MockUserRepository) {
				m.On("GetByID", mock.Anything, "nonexistent").Return(nil, repository.ErrNotFound)
			},
			wantErr: true,
		},
		{
			name:      "empty user ID",
			userID:    "",
			setupMock: func(_ *MockUserRepository) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := new(MockUserRepository)
			mockRoleRepo := new(MockRoleRepository)
			tt.setupMock(mockUserRepo)

			svc := NewUserService(mockUserRepo, mockRoleRepo, logger)

			user, err := svc.GetByID(t.Context(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.userID, user.ID)
			}

			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_ChangePassword(t *testing.T) {
	logger := zap.NewNop()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("oldpassword"), bcrypt.DefaultCost)
	require.NoError(t, err)

	tests := []struct {
		name        string
		userID      string
		oldPassword string
		newPassword string
		setupMock   func(*MockUserRepository)
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "successful password change",
			userID:      "user-123",
			oldPassword: "oldpassword",
			newPassword: "newpassword123",
			setupMock: func(m *MockUserRepository) {
				m.On("GetByID", mock.Anything, "user-123").Return(&model.User{
					BaseModel:    model.BaseModel{ID: "user-123"},
					PasswordHash: string(hashedPassword),
				}, nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:        "incorrect old password",
			userID:      "user-123",
			oldPassword: "wrongpassword",
			newPassword: "newpassword123",
			setupMock: func(m *MockUserRepository) {
				m.On("GetByID", mock.Anything, "user-123").Return(&model.User{
					BaseModel:    model.BaseModel{ID: "user-123"},
					PasswordHash: string(hashedPassword),
				}, nil)
			},
			wantErr: true,
			errMsg:  "incorrect password",
		},
		{
			name:        "password too short",
			userID:      "user-123",
			oldPassword: "oldpassword",
			newPassword: "short",
			setupMock:   func(_ *MockUserRepository) {},
			wantErr:     true,
			errMsg:      "at least 8 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := new(MockUserRepository)
			mockRoleRepo := new(MockRoleRepository)
			tt.setupMock(mockUserRepo)

			svc := NewUserService(mockUserRepo, mockRoleRepo, logger)

			err := svc.ChangePassword(t.Context(), tt.userID, tt.oldPassword, tt.newPassword)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_Delete(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name      string
		userID    string
		setupMock func(*MockUserRepository)
		wantErr   bool
	}{
		{
			name:   "successful delete",
			userID: "user-123",
			setupMock: func(m *MockUserRepository) {
				m.On("GetByID", mock.Anything, "user-123").Return(&model.User{
					BaseModel: model.BaseModel{ID: "user-123"},
				}, nil)
				m.On("Delete", mock.Anything, "user-123").Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "user not found",
			userID: "nonexistent",
			setupMock: func(m *MockUserRepository) {
				m.On("GetByID", mock.Anything, "nonexistent").Return(nil, repository.ErrNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := new(MockUserRepository)
			mockRoleRepo := new(MockRoleRepository)
			tt.setupMock(mockUserRepo)

			svc := NewUserService(mockUserRepo, mockRoleRepo, logger)

			err := svc.Delete(t.Context(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_List(t *testing.T) {
	logger := zap.NewNop()

	t.Run("successful list", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockRoleRepo := new(MockRoleRepository)

		mockUserRepo.On("List", mock.Anything, 0, 20).Return([]*model.User{
			{BaseModel: model.BaseModel{ID: "user-1"}, Username: "user1"},
			{BaseModel: model.BaseModel{ID: "user-2"}, Username: "user2"},
		}, int64(2), nil)

		svc := NewUserService(mockUserRepo, mockRoleRepo, logger)

		users, total, err := svc.List(t.Context(), UserFilters{}, 1, 20)

		assert.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, users, 2)

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("list with invalid page", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockRoleRepo := new(MockRoleRepository)

		// Page < 1 should default to 1
		mockUserRepo.On("List", mock.Anything, 0, 20).Return([]*model.User{}, int64(0), nil)

		svc := NewUserService(mockUserRepo, mockRoleRepo, logger)

		_, _, err := svc.List(t.Context(), UserFilters{}, 0, 20)

		assert.NoError(t, err)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("list error", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockRoleRepo := new(MockRoleRepository)

		mockUserRepo.On("List", mock.Anything, 0, 20).Return([]*model.User(nil), int64(0), errors.New("database error"))

		svc := NewUserService(mockUserRepo, mockRoleRepo, logger)

		_, _, err := svc.List(t.Context(), UserFilters{}, 1, 20)

		assert.Error(t, err)
		mockUserRepo.AssertExpectations(t)
	})
}
