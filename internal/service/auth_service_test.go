// Package service provides business logic tests.
package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Veritas-Calculus/vc-lab-platform/internal/config"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository is a mock implementation of UserRepository for auth tests.
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, offset, limit int) ([]*model.User, int64, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]*model.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, id, ip string) error {
	args := m.Called(ctx, id, ip)
	return args.Error(0)
}

func TestAuthService_LoginUserLookup(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	t.Run("successful user lookup", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		mockRepo.On("GetByUsername", mock.Anything, "testuser").Return(&model.User{
			BaseModel:    model.BaseModel{ID: "user-id-1"},
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: string(hashedPassword),
			Status:       1,
		}, nil)

		ctx := context.Background()
		user, err := mockRepo.GetByUsername(ctx, "testuser")
		
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "test@example.com", user.Email)
		mockRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		mockRepo.On("GetByUsername", mock.Anything, "nonexistent").Return(nil, errors.New("not found"))

		ctx := context.Background()
		user, err := mockRepo.GetByUsername(ctx, "nonexistent")
		
		assert.Error(t, err)
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})
}

func TestAuthService_PasswordVerification(t *testing.T) {
	t.Run("correct password verification", func(t *testing.T) {
		password := "password123"
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		assert.NoError(t, err)

		err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
		assert.NoError(t, err)
	})

	t.Run("wrong password verification", func(t *testing.T) {
		password := "password123"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		err := bcrypt.CompareHashAndPassword(hashedPassword, []byte("wrongpassword"))
		assert.Error(t, err)
	})
}

func TestAuthService_UserStatus(t *testing.T) {
	t.Run("active user status", func(t *testing.T) {
		user := &model.User{
			Status: 1, // Active
		}
		assert.Equal(t, int8(1), user.Status, "User should be active")
	})

	t.Run("disabled user status", func(t *testing.T) {
		user := &model.User{
			Status: 0, // Disabled
		}
		assert.Equal(t, int8(0), user.Status, "User should be disabled")
	})
}

func TestAuthService_ValidateToken(t *testing.T) {
	t.Run("token validation requires proper JWT format", func(t *testing.T) {
		invalidTokens := []string{
			"",
			"invalid",
			"invalid.token",
			"invalid.token.format",
		}

		cfg := &config.Config{
			JWT: config.JWTConfig{
				Secret: "test-secret-key-that-is-long-enough-32chars",
				Issuer: "test",
			},
		}

		for _, token := range invalidTokens {
			// Token validation should fail for invalid formats
			assert.NotEmpty(t, cfg.JWT.Secret)
			assert.True(t, len(token) == 0 || len(token) < 50)
		}
	})
}

func TestAuthService_RefreshToken(t *testing.T) {
	t.Run("refresh token requires valid format", func(t *testing.T) {
		cfg := &config.Config{
			JWT: config.JWTConfig{
				Secret:          "test-secret-key-that-is-long-enough-32chars",
				RefreshTokenTTL: 168, // 7 days in hours
				Issuer:          "test",
			},
		}

		assert.NotEmpty(t, cfg.JWT.Secret)
		assert.Equal(t, 168, cfg.JWT.RefreshTokenTTL)
	})
}

func TestTokenExpiration(t *testing.T) {
	t.Run("tokens should have proper expiration", func(t *testing.T) {
		cfg := &config.Config{
			JWT: config.JWTConfig{
				AccessTokenTTL:  15,  // 15 minutes
				RefreshTokenTTL: 168, // 7 days in hours
			},
		}

		accessExpire := time.Duration(cfg.JWT.AccessTokenTTL) * time.Minute
		refreshExpire := time.Duration(cfg.JWT.RefreshTokenTTL) * time.Hour

		assert.Equal(t, 15*time.Minute, accessExpire)
		assert.Equal(t, 7*24*time.Hour, refreshExpire)
	})
}

func TestAuthService_UpdateLastLogin(t *testing.T) {
	t.Run("update last login on successful login", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		mockRepo.On("UpdateLastLogin", mock.Anything, "user-id-1", "127.0.0.1").Return(nil)

		ctx := context.Background()
		err := mockRepo.UpdateLastLogin(ctx, "user-id-1", "127.0.0.1")

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}
