// Package service provides business logic implementations.
package service

import (
	"context"
	"fmt"

	"github.com/Veritas-Calculus/vc-lab-platform/internal/model"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/repository"
	"go.uber.org/zap"
)

// VMTemplateService defines the interface for VM template operations.
type VMTemplateService interface {
	List(ctx context.Context, provider, osType, zoneID string, page, pageSize int) ([]*model.VMTemplate, int64, error)
	ListByProvider(ctx context.Context, provider string) ([]*model.VMTemplate, error)
	Get(ctx context.Context, id string) (*model.VMTemplate, error)
	GetByName(ctx context.Context, templateName, provider string) (*model.VMTemplate, error)
	Create(ctx context.Context, input *CreateVMTemplateInput) (*model.VMTemplate, error)
	Update(ctx context.Context, id string, input *UpdateVMTemplateInput) (*model.VMTemplate, error)
	Delete(ctx context.Context, id string) error
}

// CreateVMTemplateInput represents input for creating a VM template.
type CreateVMTemplateInput struct {
	Name         string
	TemplateName string
	Provider     string
	OSType       string
	OSFamily     string
	OSVersion    string
	ZoneID       *string
	MinCPU       int
	MinMemoryMB  int
	MinDiskGB    int
	DefaultUser  string
	CloudInit    bool
	Description  string
}

// UpdateVMTemplateInput represents input for updating a VM template.
type UpdateVMTemplateInput struct {
	Name         *string
	TemplateName *string
	OSType       *string
	OSFamily     *string
	OSVersion    *string
	MinCPU       *int
	MinMemoryMB  *int
	MinDiskGB    *int
	DefaultUser  *string
	CloudInit    *bool
	Description  *string
	Status       *int8
}

type vmTemplateService struct {
	repo   repository.VMTemplateRepository
	logger *zap.Logger
}

// NewVMTemplateService creates a new VM template service.
func NewVMTemplateService(repo repository.VMTemplateRepository, logger *zap.Logger) VMTemplateService {
	return &vmTemplateService{
		repo:   repo,
		logger: logger,
	}
}

// List retrieves VM templates with pagination.
func (s *vmTemplateService) List(ctx context.Context, provider, osType, zoneID string, page, pageSize int) ([]*model.VMTemplate, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, provider, osType, zoneID, offset, pageSize)
}

// ListByProvider retrieves all active VM templates for a provider.
func (s *vmTemplateService) ListByProvider(ctx context.Context, provider string) ([]*model.VMTemplate, error) {
	return s.repo.ListByProvider(ctx, provider)
}

// Get retrieves a VM template by ID.
func (s *vmTemplateService) Get(ctx context.Context, id string) (*model.VMTemplate, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByName retrieves a VM template by template name and provider.
func (s *vmTemplateService) GetByName(ctx context.Context, templateName, provider string) (*model.VMTemplate, error) {
	return s.repo.GetByName(ctx, templateName, provider)
}

// Create creates a new VM template.
func (s *vmTemplateService) Create(ctx context.Context, input *CreateVMTemplateInput) (*model.VMTemplate, error) {
	template := &model.VMTemplate{
		Name:         input.Name,
		TemplateName: input.TemplateName,
		Provider:     input.Provider,
		OSType:       input.OSType,
		OSFamily:     input.OSFamily,
		OSVersion:    input.OSVersion,
		ZoneID:       input.ZoneID,
		MinCPU:       input.MinCPU,
		MinMemoryMB:  input.MinMemoryMB,
		MinDiskGB:    input.MinDiskGB,
		DefaultUser:  input.DefaultUser,
		CloudInit:    input.CloudInit,
		Description:  input.Description,
		Status:       1, // 1: active
	}

	if err := s.repo.Create(ctx, template); err != nil {
		return nil, fmt.Errorf("failed to create VM template: %w", err)
	}

	return template, nil
}

// Update updates an existing VM template.
func (s *vmTemplateService) Update(ctx context.Context, id string, input *UpdateVMTemplateInput) (*model.VMTemplate, error) {
	template, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		template.Name = *input.Name
	}
	if input.TemplateName != nil {
		template.TemplateName = *input.TemplateName
	}
	if input.OSType != nil {
		template.OSType = *input.OSType
	}
	if input.OSFamily != nil {
		template.OSFamily = *input.OSFamily
	}
	if input.OSVersion != nil {
		template.OSVersion = *input.OSVersion
	}
	if input.MinCPU != nil {
		template.MinCPU = *input.MinCPU
	}
	if input.MinMemoryMB != nil {
		template.MinMemoryMB = *input.MinMemoryMB
	}
	if input.MinDiskGB != nil {
		template.MinDiskGB = *input.MinDiskGB
	}
	if input.DefaultUser != nil {
		template.DefaultUser = *input.DefaultUser
	}
	if input.CloudInit != nil {
		template.CloudInit = *input.CloudInit
	}
	if input.Description != nil {
		template.Description = *input.Description
	}
	if input.Status != nil {
		template.Status = *input.Status
	}

	if err := s.repo.Update(ctx, template); err != nil {
		return nil, fmt.Errorf("failed to update VM template: %w", err)
	}

	return template, nil
}

// Delete deletes a VM template.
func (s *vmTemplateService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
