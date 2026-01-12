// Package service provides business logic implementations.
package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/Veritas-Calculus/vc-lab-platform/internal/config"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/model"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/repository"
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

// tokenBlacklist provides in-memory token blacklisting.
type tokenBlacklist struct {
	mu     sync.RWMutex
	tokens map[string]time.Time
}

func newTokenBlacklist() *tokenBlacklist {
	tb := &tokenBlacklist{
		tokens: make(map[string]time.Time),
	}
	go tb.cleanup()
	return tb
}

func (tb *tokenBlacklist) add(token string, expiry time.Time) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	// Store hash of token to save memory
	hash := sha256.Sum256([]byte(token))
	tb.tokens[hex.EncodeToString(hash[:])] = expiry
}

func (tb *tokenBlacklist) contains(token string) bool {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	hash := sha256.Sum256([]byte(token))
	_, exists := tb.tokens[hex.EncodeToString(hash[:])]
	return exists
}

func (tb *tokenBlacklist) cleanup() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		tb.mu.Lock()
		now := time.Now()
		for hash, expiry := range tb.tokens {
			if now.After(expiry) {
				delete(tb.tokens, hash)
			}
		}
		tb.mu.Unlock()
	}
}

type authService struct {
	userRepo  repository.UserRepository
	cfg       *config.Config
	blacklist *tokenBlacklist
}

// NewAuthService creates a new authentication service.
func NewAuthService(userRepo repository.UserRepository, cfg *config.Config) AuthService {
	return &authService{
		userRepo:  userRepo,
		cfg:       cfg,
		blacklist: newTokenBlacklist(),
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
	if pwdErr := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); pwdErr != nil {
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
	if s.blacklist.contains(refreshToken) {
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
	s.blacklist.add(refreshToken, claims.ExpiresAt.Time)

	// Generate new token pair
	return s.generateTokenPair(user)
}

func (s *authService) Logout(_ context.Context, accessToken string) error {
	// Parse token to get expiration
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(accessToken, claims, func(_ *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWT.Secret), nil
	})
	// Ignore parse error - we still try to blacklist if token is partially valid
	_ = err

	if token != nil && token.Valid {
		s.blacklist.add(accessToken, claims.ExpiresAt.Time)
	}

	return nil
}

func (s *authService) ValidateToken(_ context.Context, tokenString string) (*Claims, error) {
	// Check if token is blacklisted
	if s.blacklist.contains(tokenString) {
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
