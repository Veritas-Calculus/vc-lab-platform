// Package repository provides data access layer implementations.
package repository

import (
	"context"
	"errors"

	"github.com/Veritas-Calculus/vc-lab-platform/internal/model"
	"gorm.io/gorm"
)

// RegionRepository defines the interface for region data access.
type RegionRepository interface {
	Create(ctx context.Context, region *model.Region) error
	GetByID(ctx context.Context, id string) (*model.Region, error)
	GetByCode(ctx context.Context, code string) (*model.Region, error)
	List(ctx context.Context, page, pageSize int) ([]model.Region, int64, error)
	ListAll(ctx context.Context) ([]model.Region, error)
	Update(ctx context.Context, region *model.Region) error
	Delete(ctx context.Context, id string) error
}

type regionRepository struct {
	db *gorm.DB
}

// NewRegionRepository creates a new region repository.
func NewRegionRepository(db *gorm.DB) RegionRepository {
	return &regionRepository{db: db}
}

// Create creates a new region.
func (r *regionRepository) Create(ctx context.Context, region *model.Region) error {
	return r.db.WithContext(ctx).Create(region).Error
}

// GetByID retrieves a region by ID.
func (r *regionRepository) GetByID(ctx context.Context, id string) (*model.Region, error) {
	var region model.Region
	if err := r.db.WithContext(ctx).Preload("Zones").First(&region, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &region, nil
}

// GetByCode retrieves a region by code.
func (r *regionRepository) GetByCode(ctx context.Context, code string) (*model.Region, error) {
	var region model.Region
	if err := r.db.WithContext(ctx).Preload("Zones").First(&region, "code = ?", code).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &region, nil
}

// List retrieves regions with pagination.
func (r *regionRepository) List(ctx context.Context, page, pageSize int) ([]model.Region, int64, error) {
	var regions []model.Region
	var total int64

	if err := r.db.WithContext(ctx).Model(&model.Region{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).Preload("Zones").
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&regions).Error; err != nil {
		return nil, 0, err
	}

	return regions, total, nil
}

// ListAll retrieves all active regions.
func (r *regionRepository) ListAll(ctx context.Context) ([]model.Region, error) {
	var regions []model.Region
	if err := r.db.WithContext(ctx).Preload("Zones").
		Where("status = ?", 1).
		Order("name ASC").
		Find(&regions).Error; err != nil {
		return nil, err
	}
	return regions, nil
}

// Update updates a region.
func (r *regionRepository) Update(ctx context.Context, region *model.Region) error {
	return r.db.WithContext(ctx).Save(region).Error
}

// Delete soft deletes a region.
func (r *regionRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.Region{}, "id = ?", id).Error
}

// ZoneRepository defines the interface for zone data access.
type ZoneRepository interface {
	Create(ctx context.Context, zone *model.Zone) error
	GetByID(ctx context.Context, id string) (*model.Zone, error)
	GetByCode(ctx context.Context, code string) (*model.Zone, error)
	List(ctx context.Context, page, pageSize int) ([]model.Zone, int64, error)
	ListByRegion(ctx context.Context, regionID string) ([]model.Zone, error)
	Update(ctx context.Context, zone *model.Zone) error
	Delete(ctx context.Context, id string) error
}

type zoneRepository struct {
	db *gorm.DB
}

// NewZoneRepository creates a new zone repository.
func NewZoneRepository(db *gorm.DB) ZoneRepository {
	return &zoneRepository{db: db}
}

// Create creates a new zone.
func (r *zoneRepository) Create(ctx context.Context, zone *model.Zone) error {
	return r.db.WithContext(ctx).Create(zone).Error
}

// GetByID retrieves a zone by ID.
func (r *zoneRepository) GetByID(ctx context.Context, id string) (*model.Zone, error) {
	var zone model.Zone
	if err := r.db.WithContext(ctx).
		Preload("Region").
		First(&zone, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &zone, nil
}

// GetByCode retrieves a zone by code.
func (r *zoneRepository) GetByCode(ctx context.Context, code string) (*model.Zone, error) {
	var zone model.Zone
	if err := r.db.WithContext(ctx).
		Preload("Region").
		First(&zone, "code = ?", code).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &zone, nil
}

// List retrieves zones with pagination.
func (r *zoneRepository) List(ctx context.Context, page, pageSize int) ([]model.Zone, int64, error) {
	var zones []model.Zone
	var total int64

	if err := r.db.WithContext(ctx).Model(&model.Zone{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).
		Preload("Region").
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&zones).Error; err != nil {
		return nil, 0, err
	}

	return zones, total, nil
}

// ListByRegion retrieves all active zones in a region.
func (r *zoneRepository) ListByRegion(ctx context.Context, regionID string) ([]model.Zone, error) {
	var zones []model.Zone
	if err := r.db.WithContext(ctx).
		Preload("Region").
		Where("region_id = ? AND status = ?", regionID, 1).
		Order("name ASC").
		Find(&zones).Error; err != nil {
		return nil, err
	}
	return zones, nil
}

// Update updates a zone.
func (r *zoneRepository) Update(ctx context.Context, zone *model.Zone) error {
	return r.db.WithContext(ctx).Save(zone).Error
}

// Delete soft deletes a zone.
func (r *zoneRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.Zone{}, "id = ?", id).Error
}

// TerraformRegistryRepository defines the interface for terraform registry data access.
type TerraformRegistryRepository interface {
	Create(ctx context.Context, registry *model.TerraformRegistry) error
	GetByID(ctx context.Context, id string) (*model.TerraformRegistry, error)
	List(ctx context.Context, page, pageSize int) ([]model.TerraformRegistry, int64, error)
	ListAll(ctx context.Context) ([]model.TerraformRegistry, error)
	Update(ctx context.Context, registry *model.TerraformRegistry) error
	Delete(ctx context.Context, id string) error
}

type terraformRegistryRepository struct {
	db *gorm.DB
}

// NewTerraformRegistryRepository creates a new terraform registry repository.
func NewTerraformRegistryRepository(db *gorm.DB) TerraformRegistryRepository {
	return &terraformRegistryRepository{db: db}
}

func (r *terraformRegistryRepository) Create(ctx context.Context, registry *model.TerraformRegistry) error {
	return r.db.WithContext(ctx).Create(registry).Error
}

func (r *terraformRegistryRepository) GetByID(ctx context.Context, id string) (*model.TerraformRegistry, error) {
	var registry model.TerraformRegistry
	if err := r.db.WithContext(ctx).First(&registry, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &registry, nil
}

func (r *terraformRegistryRepository) List(ctx context.Context, page, pageSize int) ([]model.TerraformRegistry, int64, error) {
	var registries []model.TerraformRegistry
	var total int64

	if err := r.db.WithContext(ctx).Model(&model.TerraformRegistry{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&registries).Error; err != nil {
		return nil, 0, err
	}

	return registries, total, nil
}

func (r *terraformRegistryRepository) ListAll(ctx context.Context) ([]model.TerraformRegistry, error) {
	var registries []model.TerraformRegistry
	if err := r.db.WithContext(ctx).
		Where("status = ?", 1).
		Order("name ASC").
		Find(&registries).Error; err != nil {
		return nil, err
	}
	return registries, nil
}

func (r *terraformRegistryRepository) Update(ctx context.Context, registry *model.TerraformRegistry) error {
	return r.db.WithContext(ctx).Save(registry).Error
}

func (r *terraformRegistryRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.TerraformRegistry{}, "id = ?", id).Error
}

// TerraformProviderRepository defines the interface for terraform provider data access.
type TerraformProviderRepository interface {
	Create(ctx context.Context, provider *model.TerraformProvider) error
	GetByID(ctx context.Context, id string) (*model.TerraformProvider, error)
	List(ctx context.Context, page, pageSize int) ([]model.TerraformProvider, int64, error)
	ListByRegistry(ctx context.Context, registryID string) ([]model.TerraformProvider, error)
	Update(ctx context.Context, provider *model.TerraformProvider) error
	Delete(ctx context.Context, id string) error
}

type terraformProviderRepository struct {
	db *gorm.DB
}

// NewTerraformProviderRepository creates a new terraform provider repository.
func NewTerraformProviderRepository(db *gorm.DB) TerraformProviderRepository {
	return &terraformProviderRepository{db: db}
}

func (r *terraformProviderRepository) Create(ctx context.Context, provider *model.TerraformProvider) error {
	return r.db.WithContext(ctx).Create(provider).Error
}

func (r *terraformProviderRepository) GetByID(ctx context.Context, id string) (*model.TerraformProvider, error) {
	var provider model.TerraformProvider
	if err := r.db.WithContext(ctx).Preload("Registry").First(&provider, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &provider, nil
}

func (r *terraformProviderRepository) List(ctx context.Context, page, pageSize int) ([]model.TerraformProvider, int64, error) {
	var providers []model.TerraformProvider
	var total int64

	if err := r.db.WithContext(ctx).Model(&model.TerraformProvider{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).Preload("Registry").
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&providers).Error; err != nil {
		return nil, 0, err
	}

	return providers, total, nil
}

func (r *terraformProviderRepository) ListByRegistry(ctx context.Context, registryID string) ([]model.TerraformProvider, error) {
	var providers []model.TerraformProvider
	if err := r.db.WithContext(ctx).Preload("Registry").
		Where("registry_id = ? AND status = ?", registryID, 1).
		Order("name ASC").
		Find(&providers).Error; err != nil {
		return nil, err
	}
	return providers, nil
}

func (r *terraformProviderRepository) Update(ctx context.Context, provider *model.TerraformProvider) error {
	return r.db.WithContext(ctx).Save(provider).Error
}

func (r *terraformProviderRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.TerraformProvider{}, "id = ?", id).Error
}

// TerraformModuleRepository defines the interface for terraform module data access.
type TerraformModuleRepository interface {
	Create(ctx context.Context, module *model.TerraformModule) error
	GetByID(ctx context.Context, id string) (*model.TerraformModule, error)
	GetBySource(ctx context.Context, source string) (*model.TerraformModule, error)
	List(ctx context.Context, page, pageSize int) ([]model.TerraformModule, int64, error)
	ListAll(ctx context.Context) ([]model.TerraformModule, error)
	Update(ctx context.Context, module *model.TerraformModule) error
	Delete(ctx context.Context, id string) error
}

type terraformModuleRepository struct {
	db *gorm.DB
}

// NewTerraformModuleRepository creates a new terraform module repository.
func NewTerraformModuleRepository(db *gorm.DB) TerraformModuleRepository {
	return &terraformModuleRepository{db: db}
}

func (r *terraformModuleRepository) Create(ctx context.Context, module *model.TerraformModule) error {
	return r.db.WithContext(ctx).Create(module).Error
}

func (r *terraformModuleRepository) GetByID(ctx context.Context, id string) (*model.TerraformModule, error) {
	var module model.TerraformModule
	if err := r.db.WithContext(ctx).First(&module, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &module, nil
}

func (r *terraformModuleRepository) GetBySource(ctx context.Context, source string) (*model.TerraformModule, error) {
	var module model.TerraformModule
	if err := r.db.WithContext(ctx).First(&module, "source = ?", source).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &module, nil
}

func (r *terraformModuleRepository) List(ctx context.Context, page, pageSize int) ([]model.TerraformModule, int64, error) {
	var modules []model.TerraformModule
	var total int64

	if err := r.db.WithContext(ctx).Model(&model.TerraformModule{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&modules).Error; err != nil {
		return nil, 0, err
	}

	return modules, total, nil
}

func (r *terraformModuleRepository) ListAll(ctx context.Context) ([]model.TerraformModule, error) {
	var modules []model.TerraformModule
	if err := r.db.WithContext(ctx).
		Where("status = ?", 1).
		Order("name ASC").
		Find(&modules).Error; err != nil {
		return nil, err
	}
	return modules, nil
}

func (r *terraformModuleRepository) Update(ctx context.Context, module *model.TerraformModule) error {
	return r.db.WithContext(ctx).Save(module).Error
}

func (r *terraformModuleRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.TerraformModule{}, "id = ?", id).Error
}
