// Package middleware provides HTTP middleware components.
package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/Veritas-Calculus/vc-lab-platform/internal/constants"
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

// getStringFromContext safely extracts a string value from gin.Context.
func getStringFromContext(c *gin.Context, key string) string {
	value, exists := c.Get(key)
	if !exists || value == nil {
		return ""
	}
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}

// readRequestBody reads and restores the request body for logging.
func (m *AuditMiddleware) readRequestBody(c *gin.Context) []byte {
	if c.Request.Body == nil {
		return nil
	}
	requestBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		m.logger.Warn("failed to read request body for audit", zap.Error(err))
		return nil
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	return requestBody
}

// isHealthCheckEndpoint checks if the request is for a health check endpoint.
func isHealthCheckEndpoint(path string) bool {
	return path == "/health" || path == "/ready"
}

// determineStatus returns the status string based on HTTP status code.
func determineStatus(statusCode int) string {
	if statusCode >= constants.HTTPStatusErrorMin {
		return "failure"
	}
	return "success"
}

// Audit returns a middleware that logs API requests.
func (m *AuditMiddleware) Audit() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestBody := m.readRequestBody(c)

		// Process request
		c.Next()

		// Skip health check endpoints
		if isHealthCheckEndpoint(c.Request.URL.Path) {
			return
		}

		userIDStr := getStringFromContext(c, "user_id")
		usernameStr := getStringFromContext(c, "username")

		// Create audit log
		auditLog := &model.AuditLog{
			ID:        uuid.New().String(),
			UserID:    userIDStr,
			Username:  usernameStr,
			Action:    c.Request.Method,
			Resource:  c.Request.URL.Path,
			IPAddress: c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
			Status:    determineStatus(c.Writer.Status()),
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
