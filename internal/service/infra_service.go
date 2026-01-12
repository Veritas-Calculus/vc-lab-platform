// Package service provides business logic implementations.
package service

import (
	"context"
	"errors"

	"github.com/Veritas-Calculus/vc-lab-platform/internal/model"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/repository"
	"go.uber.org/zap"
)

// InfraService defines the interface for infrastructure management.
type InfraService interface {
	// Region operations
	ListRegions(ctx context.Context, page, pageSize int) ([]model.Region, int64, error)
	ListAllRegions(ctx context.Context) ([]model.Region, error)
	GetRegion(ctx context.Context, id string) (*model.Region, error)
	CreateRegion(ctx context.Context, input *CreateRegionInput) (*model.Region, error)
	UpdateRegion(ctx context.Context, id string, input *UpdateRegionInput) (*model.Region, error)
	DeleteRegion(ctx context.Context, id string) error

	// Zone operations
	ListZones(ctx context.Context, page, pageSize int) ([]model.Zone, int64, error)
	ListZonesByRegion(ctx context.Context, regionID string) ([]model.Zone, error)
	GetZone(ctx context.Context, id string) (*model.Zone, error)
	CreateZone(ctx context.Context, input *CreateZoneInput) (*model.Zone, error)
	UpdateZone(ctx context.Context, id string, input *UpdateZoneInput) (*model.Zone, error)
	DeleteZone(ctx context.Context, id string) error

	// Terraform Registry operations
	ListRegistries(ctx context.Context, page, pageSize int) ([]model.TerraformRegistry, int64, error)
	ListAllRegistries(ctx context.Context) ([]model.TerraformRegistry, error)
	GetRegistry(ctx context.Context, id string) (*model.TerraformRegistry, error)
	CreateRegistry(ctx context.Context, input *CreateRegistryInput) (*model.TerraformRegistry, error)
	UpdateRegistry(ctx context.Context, id string, input *UpdateRegistryInput) (*model.TerraformRegistry, error)
	DeleteRegistry(ctx context.Context, id string) error

	// Terraform Provider operations
	ListProviders(ctx context.Context, page, pageSize int) ([]model.TerraformProvider, int64, error)
	ListProvidersByRegistry(ctx context.Context, registryID string) ([]model.TerraformProvider, error)
	GetProvider(ctx context.Context, id string) (*model.TerraformProvider, error)
	CreateProvider(ctx context.Context, input *CreateTfProviderInput) (*model.TerraformProvider, error)
	UpdateProvider(ctx context.Context, id string, input *UpdateTfProviderInput) (*model.TerraformProvider, error)
	DeleteProvider(ctx context.Context, id string) error

	// Terraform Module operations
	ListModules(ctx context.Context, page, pageSize int) ([]model.TerraformModule, int64, error)
	ListAllModules(ctx context.Context) ([]model.TerraformModule, error)
	GetModule(ctx context.Context, id string) (*model.TerraformModule, error)
	CreateModule(ctx context.Context, input *CreateModuleInput) (*model.TerraformModule, error)
	UpdateModule(ctx context.Context, id string, input *UpdateModuleInput) (*model.TerraformModule, error)
	DeleteModule(ctx context.Context, id string) error
}

// CreateRegionInput represents input for creating a region.
type CreateRegionInput struct {
	Name        string
	Code        string
	DisplayName string
	Description string
}

// UpdateRegionInput represents input for updating a region.
type UpdateRegionInput struct {
	Name        *string
	DisplayName *string
	Description *string
	Status      *int8
}

// CreateZoneInput represents input for creating a zone.
type CreateZoneInput struct {
	Name        string
	Code        string
	DisplayName string
	Description string
	RegionID    string
	IsDefault   bool
}

// UpdateZoneInput represents input for updating a zone.
type UpdateZoneInput struct {
	Name        *string
	DisplayName *string
	Description *string
	Status      *int8
	IsDefault   *bool
}

// CreateRegistryInput represents input for creating a terraform registry.
type CreateRegistryInput struct {
	Name        string
	Endpoint    string
	Username    string
	Token       string
	Description string
	IsDefault   bool
}

// UpdateRegistryInput represents input for updating a terraform registry.
type UpdateRegistryInput struct {
	Name        *string
	Endpoint    *string
	Username    *string
	Token       *string
	Description *string
	Status      *int8
	IsDefault   *bool
}

// CreateTfProviderInput represents input for creating a terraform provider.
type CreateTfProviderInput struct {
	Name        string
	Namespace   string
	Source      string
	Version     string
	RegistryID  string
	Description string
}

// UpdateTfProviderInput represents input for updating a terraform provider.
type UpdateTfProviderInput struct {
	Name        *string
	Namespace   *string
	Source      *string
	Version     *string
	Description *string
	Status      *int8
}

// CreateModuleInput represents input for creating a terraform module.
type CreateModuleInput struct {
	Name        string
	Source      string
	Version     string
	RegistryID  *string
	ProviderID  *string
	Description string
	Variables   string
}

// UpdateModuleInput represents input for updating a terraform module.
type UpdateModuleInput struct {
	Name        *string
	Source      *string
	Version     *string
	RegistryID  *string
	ProviderID  *string
	Description *string
	Variables   *string
	Status      *int8
}

type infraService struct {
	regionRepo   repository.RegionRepository
	zoneRepo     repository.ZoneRepository
	registryRepo repository.TerraformRegistryRepository
	providerRepo repository.TerraformProviderRepository
	moduleRepo   repository.TerraformModuleRepository
	logger       *zap.Logger
}

// NewInfraService creates a new infrastructure service.
func NewInfraService(
	regionRepo repository.RegionRepository,
	zoneRepo repository.ZoneRepository,
	registryRepo repository.TerraformRegistryRepository,
	providerRepo repository.TerraformProviderRepository,
	moduleRepo repository.TerraformModuleRepository,
	logger *zap.Logger,
) InfraService {
	return &infraService{
		regionRepo:   regionRepo,
		zoneRepo:     zoneRepo,
		registryRepo: registryRepo,
		providerRepo: providerRepo,
		moduleRepo:   moduleRepo,
		logger:       logger,
	}
}

// ListRegions retrieves regions with pagination.
func (s *infraService) ListRegions(ctx context.Context, page, pageSize int) ([]model.Region, int64, error) {
	return s.regionRepo.List(ctx, page, pageSize)
}

// ListAllRegions retrieves all active regions.
func (s *infraService) ListAllRegions(ctx context.Context) ([]model.Region, error) {
	return s.regionRepo.ListAll(ctx)
}

// GetRegion retrieves a region by ID.
func (s *infraService) GetRegion(ctx context.Context, id string) (*model.Region, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}
	return s.regionRepo.GetByID(ctx, id)
}

// CreateRegion creates a new region.
func (s *infraService) CreateRegion(ctx context.Context, input *CreateRegionInput) (*model.Region, error) {
	if input.Name == "" {
		return nil, errors.New("name is required")
	}
	if input.Code == "" {
		return nil, errors.New("code is required")
	}
	if input.DisplayName == "" {
		input.DisplayName = input.Name
	}

	// Check if code already exists
	if _, err := s.regionRepo.GetByCode(ctx, input.Code); err == nil {
		return nil, errors.New("region code already exists")
	}

	region := &model.Region{
		Name:        input.Name,
		Code:        input.Code,
		DisplayName: input.DisplayName,
		Description: input.Description,
		Status:      1,
	}

	if err := s.regionRepo.Create(ctx, region); err != nil {
		s.logger.Error("failed to create region", zap.Error(err))
		return nil, errors.New("failed to create region")
	}

	return region, nil
}

// UpdateRegion updates a region.
func (s *infraService) UpdateRegion(ctx context.Context, id string, input *UpdateRegionInput) (*model.Region, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	region, err := s.regionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		region.Name = *input.Name
	}
	if input.DisplayName != nil {
		region.DisplayName = *input.DisplayName
	}
	if input.Description != nil {
		region.Description = *input.Description
	}
	if input.Status != nil {
		region.Status = *input.Status
	}

	if err := s.regionRepo.Update(ctx, region); err != nil {
		s.logger.Error("failed to update region", zap.Error(err))
		return nil, errors.New("failed to update region")
	}

	return region, nil
}

// DeleteRegion deletes a region.
func (s *infraService) DeleteRegion(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}

	if _, err := s.regionRepo.GetByID(ctx, id); err != nil {
		return err
	}

	if err := s.regionRepo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete region", zap.Error(err))
		return errors.New("failed to delete region")
	}

	return nil
}

// ListZones retrieves zones with pagination.
func (s *infraService) ListZones(ctx context.Context, page, pageSize int) ([]model.Zone, int64, error) {
	return s.zoneRepo.List(ctx, page, pageSize)
}

// ListZonesByRegion retrieves zones by region ID.
func (s *infraService) ListZonesByRegion(ctx context.Context, regionID string) ([]model.Zone, error) {
	if regionID == "" {
		return nil, errors.New("region_id cannot be empty")
	}
	return s.zoneRepo.ListByRegion(ctx, regionID)
}

// GetZone retrieves a zone by ID.
func (s *infraService) GetZone(ctx context.Context, id string) (*model.Zone, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}
	return s.zoneRepo.GetByID(ctx, id)
}

// CreateZone creates a new zone.
func (s *infraService) CreateZone(ctx context.Context, input *CreateZoneInput) (*model.Zone, error) {
	if input.Name == "" {
		return nil, errors.New("name is required")
	}
	if input.Code == "" {
		return nil, errors.New("code is required")
	}
	if input.RegionID == "" {
		return nil, errors.New("region_id is required")
	}
	if input.DisplayName == "" {
		input.DisplayName = input.Name
	}

	// Verify region exists
	if _, err := s.regionRepo.GetByID(ctx, input.RegionID); err != nil {
		return nil, errors.New("region not found")
	}

	// Check if code already exists
	if _, err := s.zoneRepo.GetByCode(ctx, input.Code); err == nil {
		return nil, errors.New("zone code already exists")
	}

	zone := &model.Zone{
		Name:        input.Name,
		Code:        input.Code,
		DisplayName: input.DisplayName,
		Description: input.Description,
		RegionID:    input.RegionID,
		IsDefault:   input.IsDefault,
		Status:      1,
	}

	if err := s.zoneRepo.Create(ctx, zone); err != nil {
		s.logger.Error("failed to create zone", zap.Error(err))
		return nil, errors.New("failed to create zone")
	}

	return zone, nil
}

// UpdateZone updates a zone.
func (s *infraService) UpdateZone(ctx context.Context, id string, input *UpdateZoneInput) (*model.Zone, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	zone, err := s.zoneRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		zone.Name = *input.Name
	}
	if input.DisplayName != nil {
		zone.DisplayName = *input.DisplayName
	}
	if input.Description != nil {
		zone.Description = *input.Description
	}
	if input.Status != nil {
		zone.Status = *input.Status
	}
	if input.IsDefault != nil {
		zone.IsDefault = *input.IsDefault
	}

	if err := s.zoneRepo.Update(ctx, zone); err != nil {
		s.logger.Error("failed to update zone", zap.Error(err))
		return nil, errors.New("failed to update zone")
	}

	return zone, nil
}

// DeleteZone deletes a zone.
func (s *infraService) DeleteZone(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}

	if _, err := s.zoneRepo.GetByID(ctx, id); err != nil {
		return err
	}

	if err := s.zoneRepo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete zone", zap.Error(err))
		return errors.New("failed to delete zone")
	}

	return nil
}

// Terraform Registry operations

func (s *infraService) ListRegistries(ctx context.Context, page, pageSize int) ([]model.TerraformRegistry, int64, error) {
	return s.registryRepo.List(ctx, page, pageSize)
}

func (s *infraService) ListAllRegistries(ctx context.Context) ([]model.TerraformRegistry, error) {
	return s.registryRepo.ListAll(ctx)
}

func (s *infraService) GetRegistry(ctx context.Context, id string) (*model.TerraformRegistry, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}
	return s.registryRepo.GetByID(ctx, id)
}

func (s *infraService) CreateRegistry(ctx context.Context, input *CreateRegistryInput) (*model.TerraformRegistry, error) {
	if input.Name == "" {
		return nil, errors.New("name is required")
	}
	if input.Endpoint == "" {
		return nil, errors.New("endpoint is required")
	}

	registry := &model.TerraformRegistry{
		Name:        input.Name,
		Endpoint:    input.Endpoint,
		Username:    input.Username,
		Token:       input.Token,
		Description: input.Description,
		IsDefault:   input.IsDefault,
		Status:      1,
	}

	if err := s.registryRepo.Create(ctx, registry); err != nil {
		s.logger.Error("failed to create registry", zap.Error(err))
		return nil, errors.New("failed to create registry")
	}

	return registry, nil
}

func (s *infraService) UpdateRegistry(ctx context.Context, id string, input *UpdateRegistryInput) (*model.TerraformRegistry, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	registry, err := s.registryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		registry.Name = *input.Name
	}
	if input.Endpoint != nil {
		registry.Endpoint = *input.Endpoint
	}
	if input.Username != nil {
		registry.Username = *input.Username
	}
	if input.Token != nil {
		registry.Token = *input.Token
	}
	if input.Description != nil {
		registry.Description = *input.Description
	}
	if input.Status != nil {
		registry.Status = *input.Status
	}
	if input.IsDefault != nil {
		registry.IsDefault = *input.IsDefault
	}

	if err := s.registryRepo.Update(ctx, registry); err != nil {
		s.logger.Error("failed to update registry", zap.Error(err))
		return nil, errors.New("failed to update registry")
	}

	return registry, nil
}

func (s *infraService) DeleteRegistry(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}

	if _, err := s.registryRepo.GetByID(ctx, id); err != nil {
		return err
	}

	if err := s.registryRepo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete registry", zap.Error(err))
		return errors.New("failed to delete registry")
	}

	return nil
}

// Terraform Provider operations

func (s *infraService) ListProviders(ctx context.Context, page, pageSize int) ([]model.TerraformProvider, int64, error) {
	return s.providerRepo.List(ctx, page, pageSize)
}

func (s *infraService) ListProvidersByRegistry(ctx context.Context, registryID string) ([]model.TerraformProvider, error) {
	if registryID == "" {
		return nil, errors.New("registry_id cannot be empty")
	}
	return s.providerRepo.ListByRegistry(ctx, registryID)
}

func (s *infraService) GetProvider(ctx context.Context, id string) (*model.TerraformProvider, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}
	return s.providerRepo.GetByID(ctx, id)
}

func (s *infraService) CreateProvider(ctx context.Context, input *CreateTfProviderInput) (*model.TerraformProvider, error) {
	if input.Name == "" {
		return nil, errors.New("name is required")
	}
	if input.RegistryID == "" {
		return nil, errors.New("registry_id is required")
	}

	// Verify registry exists
	if _, err := s.registryRepo.GetByID(ctx, input.RegistryID); err != nil {
		return nil, errors.New("registry not found")
	}

	provider := &model.TerraformProvider{
		Name:        input.Name,
		Namespace:   input.Namespace,
		Source:      input.Source,
		Version:     input.Version,
		RegistryID:  input.RegistryID,
		Description: input.Description,
		Status:      1,
	}

	if err := s.providerRepo.Create(ctx, provider); err != nil {
		s.logger.Error("failed to create provider", zap.Error(err))
		return nil, errors.New("failed to create provider")
	}

	return provider, nil
}

func (s *infraService) UpdateProvider(ctx context.Context, id string, input *UpdateTfProviderInput) (*model.TerraformProvider, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	provider, err := s.providerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		provider.Name = *input.Name
	}
	if input.Namespace != nil {
		provider.Namespace = *input.Namespace
	}
	if input.Source != nil {
		provider.Source = *input.Source
	}
	if input.Version != nil {
		provider.Version = *input.Version
	}
	if input.Description != nil {
		provider.Description = *input.Description
	}
	if input.Status != nil {
		provider.Status = *input.Status
	}

	if err := s.providerRepo.Update(ctx, provider); err != nil {
		s.logger.Error("failed to update provider", zap.Error(err))
		return nil, errors.New("failed to update provider")
	}

	return provider, nil
}

func (s *infraService) DeleteProvider(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}

	if _, err := s.providerRepo.GetByID(ctx, id); err != nil {
		return err
	}

	if err := s.providerRepo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete provider", zap.Error(err))
		return errors.New("failed to delete provider")
	}

	return nil
}

// Terraform Module operations

func (s *infraService) ListModules(ctx context.Context, page, pageSize int) ([]model.TerraformModule, int64, error) {
	return s.moduleRepo.List(ctx, page, pageSize)
}

func (s *infraService) ListAllModules(ctx context.Context) ([]model.TerraformModule, error) {
	return s.moduleRepo.ListAll(ctx)
}

func (s *infraService) GetModule(ctx context.Context, id string) (*model.TerraformModule, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}
	return s.moduleRepo.GetByID(ctx, id)
}

func (s *infraService) CreateModule(ctx context.Context, input *CreateModuleInput) (*model.TerraformModule, error) {
	if input.Name == "" {
		return nil, errors.New("name is required")
	}
	if input.Source == "" {
		return nil, errors.New("source is required")
	}

	module := &model.TerraformModule{
		Name:        input.Name,
		Source:      input.Source,
		Version:     input.Version,
		RegistryID:  input.RegistryID,
		ProviderID:  input.ProviderID,
		Description: input.Description,
		Variables:   input.Variables,
		Status:      1,
	}

	if err := s.moduleRepo.Create(ctx, module); err != nil {
		s.logger.Error("failed to create module", zap.Error(err))
		return nil, errors.New("failed to create module")
	}

	return module, nil
}

func (s *infraService) UpdateModule(ctx context.Context, id string, input *UpdateModuleInput) (*model.TerraformModule, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	module, err := s.moduleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		module.Name = *input.Name
	}
	if input.Source != nil {
		module.Source = *input.Source
	}
	if input.Version != nil {
		module.Version = *input.Version
	}
	if input.RegistryID != nil {
		module.RegistryID = input.RegistryID
	}
	if input.ProviderID != nil {
		module.ProviderID = input.ProviderID
	}
	if input.Description != nil {
		module.Description = *input.Description
	}
	if input.Variables != nil {
		module.Variables = *input.Variables
	}
	if input.Status != nil {
		module.Status = *input.Status
	}

	if err := s.moduleRepo.Update(ctx, module); err != nil {
		s.logger.Error("failed to update module", zap.Error(err))
		return nil, errors.New("failed to update module")
	}

	return module, nil
}

func (s *infraService) DeleteModule(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}

	if _, err := s.moduleRepo.GetByID(ctx, id); err != nil {
		return err
	}

	if err := s.moduleRepo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete module", zap.Error(err))
		return errors.New("failed to delete module")
	}

	return nil
}
