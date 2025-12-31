// Package middleware provides HTTP middleware components.
package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/Veritas-Calculus/vc-lab-platform/internal/model"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AuditMiddleware provides audit logging middleware.
type AuditMiddleware struct {
	auditRepo repository.AuditRepository
	logger    *zap.Logger
}

// NewAuditMiddleware creates a new audit middleware.
func NewAuditMiddleware(auditRepo repository.AuditRepository, logger *zap.Logger) *AuditMiddleware {
	return &AuditMiddleware{
		auditRepo: auditRepo,
		logger:    logger,
	}
}

// Audit returns a middleware that logs API requests.
func (m *AuditMiddleware) Audit() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Read request body
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Process request
		c.Next()

		// Skip health check endpoints
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/ready" {
			return
		}

		// Get user info from context
		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")

		userIDStr := ""
		if userID != nil {
			userIDStr = userID.(string)
		}
		usernameStr := ""
		if username != nil {
			usernameStr = username.(string)
		}

		// Determine action from method
		action := c.Request.Method
		resource := c.Request.URL.Path

		// Determine status
		status := "success"
		if c.Writer.Status() >= 400 {
			status = "failure"
		}

		// Create audit log
		auditLog := &model.AuditLog{
			ID:        uuid.New().String(),
			UserID:    userIDStr,
			Username:  usernameStr,
			Action:    action,
			Resource:  resource,
			IPAddress: c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
			Status:    status,
			Details:   string(requestBody),
			CreatedAt: time.Now(),
		}

		// Log asynchronously to avoid blocking requests
		go func(log *model.AuditLog) {
			if err := m.auditRepo.Create(c.Request.Context(), log); err != nil {
				m.logger.Error("failed to create audit log", zap.Error(err))
			}
		}(auditLog)

		// Log the request
		m.logger.Info("request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", time.Since(start)),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_id", userIDStr),
		)
	}
}
