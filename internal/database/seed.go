// Package database provides database connection and management utilities.
package database

import (
	"errors"
	"log"

	"github.com/Veritas-Calculus/vc-lab-platform/internal/config"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Seed initializes default data including roles and admin user.
func Seed(db *gorm.DB, cfg *config.Config) error {
	if err := seedRoles(db); err != nil {
		return err
	}
	if err := seedAdminUser(db, cfg); err != nil {
		return err
	}
	return nil
}

func seedRoles(db *gorm.DB) error {
	roles := []model.Role{
		{
			Name:        "Administrator",
			Code:        "admin",
			Description: "System administrator with full access",
			IsSystem:    true,
			Status:      1,
		},
		{
			Name:        "User",
			Code:        "user",
			Description: "Regular user with limited access",
			IsSystem:    true,
			Status:      1,
		},
		{
			Name:        "Viewer",
			Code:        "viewer",
			Description: "Read-only access",
			IsSystem:    true,
			Status:      1,
		},
	}

	for _, role := range roles {
		var existing model.Role
		result := db.Where("code = ?", role.Code).First(&existing)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			if err := db.Create(&role).Error; err != nil {
				return err
			}
			log.Printf("Created role: %s", role.Code)
		} else if result.Error == nil && !existing.IsSystem {
			// Update existing role to mark as system role
			db.Model(&existing).Update("is_system", true)
		}
	}
	return nil
}

func seedAdminUser(db *gorm.DB, cfg *config.Config) error {
	// Check if admin user already exists
	var existing model.User
	result := db.Where("username = ?", cfg.Admin.Username).First(&existing)
	if result.Error == nil {
		// Admin user already exists, ensure it's marked as system user
		if !existing.IsSystem || existing.Source == "" {
			db.Model(&existing).Updates(map[string]interface{}{
				"is_system": true,
				"source":    model.UserSourceLocal,
			})
		}
		return nil
	}
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cfg.Admin.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Get admin role
	var adminRole model.Role
	if err := db.Where("code = ?", "admin").First(&adminRole).Error; err != nil {
		return err
	}

	// Create admin user
	adminUser := model.User{
		Username:     cfg.Admin.Username,
		Email:        cfg.Admin.Email,
		PasswordHash: string(hashedPassword),
		DisplayName:  "Administrator",
		Source:       model.UserSourceLocal,
		IsSystem:     true,
		Status:       1,
		Roles:        []model.Role{adminRole},
	}

	if err := db.Create(&adminUser).Error; err != nil {
		return err
	}

	log.Printf("Created admin user: %s (password: %s)", cfg.Admin.Username, cfg.Admin.Password)
	return nil
}
