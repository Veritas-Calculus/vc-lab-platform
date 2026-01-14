// Package notification provides notification services for the application.
package notification

import (
	"context"
	"fmt"
	"time"

	"github.com/Veritas-Calculus/vc-lab-platform/internal/sanitize"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Type represents the type of notification.
type Type string

const (
	// TypeEmail represents email notification.
	TypeEmail Type = "email"
	// TypeInApp represents in-app notification.
	TypeInApp Type = "in_app"
	// TypeWebhook represents webhook notification.
	TypeWebhook Type = "webhook"
)

// Notification represents a notification message.
type Notification struct {
	ID        string                 `json:"id"`
	Type      Type                   `json:"type"`
	UserID    string                 `json:"user_id"`
	Title     string                 `json:"title"`
	Content   string                 `json:"content"`
	Data      map[string]interface{} `json:"data"`
	Read      bool                   `json:"read"`
	CreatedAt time.Time              `json:"created_at"`
}

// Service provides notification capabilities.
type Service interface {
	// Send sends a notification to a user.
	Send(ctx context.Context, notification *Notification) error
	// SendBatch sends multiple notifications.
	SendBatch(ctx context.Context, notifications []*Notification) error

	// NotifyResourceRequestApproved notifies user about resource request approval.
	NotifyResourceRequestApproved(ctx context.Context, userID, requestID, requestTitle, reason string) error
	// NotifyResourceRequestRejected notifies user about resource request rejection.
	NotifyResourceRequestRejected(ctx context.Context, userID, requestID, requestTitle, reason string) error
	// NotifyResourceProvisioned notifies user about resource provisioning completion.
	NotifyResourceProvisioned(ctx context.Context, userID, resourceID, resourceName string, outputs map[string]string) error
	// NotifyResourceProvisioningFailed notifies user about resource provisioning failure.
	NotifyResourceProvisioningFailed(ctx context.Context, userID, requestID, requestTitle, errorMsg string) error
}

// service implements Service.
type service struct {
	db     *gorm.DB
	logger *zap.Logger
	// In production, add email client, webhook client, etc.
}

// NewService creates a new notification service.
func NewService(db *gorm.DB, logger *zap.Logger) Service {
	return &service{
		db:     db,
		logger: logger,
	}
}

// Send sends a notification to a user.
func (s *service) Send(ctx context.Context, notification *Notification) error {
	// In production, implement actual notification sending
	// For now, just log the notification
	s.logger.Info("sending notification",
		zap.String("type", string(notification.Type)),
		zap.String("user_id", notification.UserID),
		zap.String("title", sanitize.Content(notification.Title)),
		zap.String("content", sanitize.Content(notification.Content)),
	)

	// TODO: Implement actual notification delivery based on type
	switch notification.Type {
	case TypeEmail:
		return s.sendEmail(ctx, notification)
	case TypeInApp:
		return s.sendInApp(ctx, notification)
	case TypeWebhook:
		return s.sendWebhook(ctx, notification)
	default:
		return fmt.Errorf("unsupported notification type: %s", notification.Type)
	}
}

// SendBatch sends multiple notifications.
func (s *service) SendBatch(ctx context.Context, notifications []*Notification) error {
	for _, notification := range notifications {
		if err := s.Send(ctx, notification); err != nil {
			s.logger.Error("failed to send notification",
				zap.String("notification_id", notification.ID),
				zap.Error(err),
			)
			// Continue sending other notifications even if one fails
		}
	}
	return nil
}

// NotifyResourceRequestApproved notifies user about resource request approval.
func (s *service) NotifyResourceRequestApproved(ctx context.Context, userID, requestID, requestTitle, reason string) error {
	notification := &Notification{
		Type:    TypeInApp,
		UserID:  userID,
		Title:   "Resource Request Approved",
		Content: fmt.Sprintf("Your resource request '%s' has been approved and is being provisioned.", requestTitle),
		Data: map[string]interface{}{
			"request_id": requestID,
			"status":     "approved",
			"reason":     reason,
		},
		CreatedAt: time.Now(),
	}
	return s.Send(ctx, notification)
}

// NotifyResourceRequestRejected notifies user about resource request rejection.
func (s *service) NotifyResourceRequestRejected(ctx context.Context, userID, requestID, requestTitle, reason string) error {
	notification := &Notification{
		Type:    TypeInApp,
		UserID:  userID,
		Title:   "Resource Request Rejected",
		Content: fmt.Sprintf("Your resource request '%s' has been rejected. Reason: %s", requestTitle, reason),
		Data: map[string]interface{}{
			"request_id": requestID,
			"status":     "rejected",
			"reason":     reason,
		},
		CreatedAt: time.Now(),
	}
	return s.Send(ctx, notification)
}

// NotifyResourceProvisioned notifies user about resource provisioning completion.
func (s *service) NotifyResourceProvisioned(ctx context.Context, userID, resourceID, resourceName string, outputs map[string]string) error {
	notification := &Notification{
		Type:    TypeInApp,
		UserID:  userID,
		Title:   "Resource Provisioned",
		Content: fmt.Sprintf("Your resource '%s' has been successfully provisioned and is ready to use.", resourceName),
		Data: map[string]interface{}{
			"resource_id":   resourceID,
			"resource_name": resourceName,
			"outputs":       outputs,
			"status":        "completed",
		},
		CreatedAt: time.Now(),
	}
	return s.Send(ctx, notification)
}

// NotifyResourceProvisioningFailed notifies user about resource provisioning failure.
func (s *service) NotifyResourceProvisioningFailed(ctx context.Context, userID, requestID, requestTitle, errorMsg string) error {
	notification := &Notification{
		Type:    TypeInApp,
		UserID:  userID,
		Title:   "Resource Provisioning Failed",
		Content: fmt.Sprintf("Failed to provision resource for request '%s'. Error: %s", requestTitle, errorMsg),
		Data: map[string]interface{}{
			"request_id": requestID,
			"status":     "failed",
			"error":      errorMsg,
		},
		CreatedAt: time.Now(),
	}
	return s.Send(ctx, notification)
}

// sendEmail sends an email notification.
func (s *service) sendEmail(_ context.Context, notification *Notification) error {
	// TODO: Implement email sending using SMTP or email service provider
	s.logger.Info("would send email notification",
		zap.String("user_id", notification.UserID),
		zap.String("title", notification.Title),
	)
	return nil
}

// sendInApp sends an in-app notification.
func (s *service) sendInApp(_ context.Context, notification *Notification) error {
	// TODO: Store notification in database for user to view in app
	s.logger.Info("would send in-app notification",
		zap.String("user_id", notification.UserID),
		zap.String("title", notification.Title),
	)
	return nil
}

// sendWebhook sends a webhook notification.
func (s *service) sendWebhook(_ context.Context, notification *Notification) error {
	// TODO: Send HTTP POST to webhook URL
	s.logger.Info("would send webhook notification",
		zap.String("user_id", notification.UserID),
		zap.String("title", notification.Title),
	)
	return nil
}
