// Package repository provides audit log data access.
package repository

import (
	"context"

	"github.com/Veritas-Calculus/vc-lab-platform/internal/model"
	"gorm.io/gorm"
)

// AuditRepository defines the interface for audit log data access.
type AuditRepository interface {
	Create(ctx context.Context, log *model.AuditLog) error
	List(ctx context.Context, filters AuditFilters, offset, limit int) ([]*model.AuditLog, int64, error)
}

// AuditFilters defines filters for audit log queries.
type AuditFilters struct {
	UserID       string
	Action       string
	ResourceType string
	StartTime    string
	EndTime      string
}

type auditRepository struct {
	db *gorm.DB
}

// NewAuditRepository creates a new audit repository.
func NewAuditRepository(db *gorm.DB) AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) Create(ctx context.Context, log *model.AuditLog) error {
	result := r.db.WithContext(ctx).Create(log)
	return result.Error
}

func (r *auditRepository) List(ctx context.Context, filters AuditFilters, offset, limit int) ([]*model.AuditLog, int64, error) {
	var logs []*model.AuditLog
	var total int64

	query := r.db.WithContext(ctx).Model(&model.AuditLog{})

	// Apply filters
	if filters.UserID != "" {
		query = query.Where("user_id = ?", filters.UserID)
	}
	if filters.Action != "" {
		query = query.Where("action = ?", filters.Action)
	}
	if filters.ResourceType != "" {
		query = query.Where("resource_type = ?", filters.ResourceType)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	result := query.Preload("User").Offset(offset).Limit(limit).Order("created_at DESC").Find(&logs)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return logs, total, nil
}
