// Package database provides database connection and management utilities.
package database

import (
	"github.com/Veritas-Calculus/vc-lab-platform/internal/model"
	"gorm.io/gorm"
)

// AutoMigrate runs database migrations for all models.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Role{},
		&model.Permission{},
		&model.Resource{},
		&model.ResourceRequest{},
		&model.AuditLog{},
	)
}
