// Package service provides business logic implementations.
package service

import (
	"context"
	"errors"
	"time"

	"github.com/Veritas-Calculus/vc-lab-platform/internal/config"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/model"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/repository"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Common errors.
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserDisabled       = errors.New("user account is disabled")
	ErrTokenBlacklisted   = errors.New("token has been revoked")
)

// AuthService defines the authentication service interface.
type AuthService interface {
	Login(ctx context.Context, username, password, clientIP string) (*TokenPair, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
	Logout(ctx context.Context, accessToken string) error
	ValidateToken(ctx context.Context, tokenString string) (*Claims, error)
}

// TokenPair represents access and refresh tokens.
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// Claims represents the JWT claims.
type Claims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

type authService struct {
	userRepo repository.UserRepository
	rdb      *redis.Client
	cfg      *config.Config
}

// NewAuthService creates a new authentication service.
func NewAuthService(userRepo repository.UserRepository, rdb *redis.Client, cfg *config.Config) AuthService {
	return &authService{
		userRepo: userRepo,
		rdb:      rdb,
		cfg:      cfg,
	}
}

func (s *authService) Login(ctx context.Context, username, password, clientIP string) (*TokenPair, error) {
	// Validate input
	if username == "" || password == "" {
		return nil, ErrInvalidCredentials
	}

	// Get user by username
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Check user status
	if user.Status == 0 {
		return nil, ErrUserDisabled
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate token pair
	tokenPair, err := s.generateTokenPair(user)
	if err != nil {
		return nil, err
	}

	// Update last login
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID, clientIP); err != nil {
		// Log error but don't fail login
		_ = err
	}

	return tokenPair, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	// Parse and validate refresh token
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(s.cfg.JWT.Secret), nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidCredentials
	}

	// Check if token is blacklisted
	if s.isTokenBlacklisted(ctx, refreshToken) {
		return nil, ErrTokenBlacklisted
	}

	// Get fresh user data
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	if user.Status == 0 {
		return nil, ErrUserDisabled
	}

	// Blacklist old refresh token
	s.blacklistToken(ctx, refreshToken, time.Duration(s.cfg.JWT.RefreshTokenTTL)*time.Hour)

	// Generate new token pair
	return s.generateTokenPair(user)
}

func (s *authService) Logout(ctx context.Context, accessToken string) error {
	// Parse token to get expiration
	claims := &Claims{}
	token, _ := jwt.ParseWithClaims(accessToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWT.Secret), nil
	})

	if token != nil && token.Valid {
		// Calculate remaining TTL
		ttl := time.Until(claims.ExpiresAt.Time)
		if ttl > 0 {
			s.blacklistToken(ctx, accessToken, ttl)
		}
	}

	return nil
}

func (s *authService) ValidateToken(ctx context.Context, tokenString string) (*Claims, error) {
	// Check if token is blacklisted
	if s.isTokenBlacklisted(ctx, tokenString) {
		return nil, ErrTokenBlacklisted
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(s.cfg.JWT.Secret), nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidCredentials
	}

	return claims, nil
}

func (s *authService) generateTokenPair(user *model.User) (*TokenPair, error) {
	now := time.Now()
	accessExpiry := now.Add(time.Duration(s.cfg.JWT.AccessTokenTTL) * time.Minute)
	refreshExpiry := now.Add(time.Duration(s.cfg.JWT.RefreshTokenTTL) * time.Hour)

	// Extract role codes
	roles := make([]string, len(user.Roles))
	for i, role := range user.Roles {
		roles[i] = role.Code
	}

	// Generate access token
	accessClaims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.cfg.JWT.Issuer,
			Subject:   user.ID,
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshClaims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.cfg.JWT.Issuer,
			Subject:   user.ID,
			ExpiresAt: jwt.NewNumericDate(refreshExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    accessExpiry,
		TokenType:    "Bearer",
	}, nil
}

func (s *authService) blacklistToken(ctx context.Context, token string, ttl time.Duration) {
	if s.rdb == nil {
		return
	}
	key := "blacklist:" + token
	s.rdb.Set(ctx, key, "1", ttl)
}

func (s *authService) isTokenBlacklisted(ctx context.Context, token string) bool {
	if s.rdb == nil {
		return false
	}
	key := "blacklist:" + token
	result, err := s.rdb.Exists(ctx, key).Result()
	return err == nil && result > 0
}

// HashPassword hashes a password using bcrypt.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword checks if a password matches a hash.
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
