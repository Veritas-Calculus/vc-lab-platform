// Package repository provides resource data access.
package repository

import (
	"context"
	"errors"

	"github.com/Veritas-Calculus/vc-lab-platform/internal/model"
	"gorm.io/gorm"
)

// ResourceRepository defines the interface for resource data access.
type ResourceRepository interface {
	Create(ctx context.Context, resource *model.Resource) error
	GetByID(ctx context.Context, id string) (*model.Resource, error)
	Update(ctx context.Context, resource *model.Resource) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filters ResourceFilters, offset, limit int) ([]*model.Resource, int64, error)
}

// ResourceFilters defines filters for resource queries.
type ResourceFilters struct {
	Type        string
	Provider    string
	Status      string
	Environment string
	OwnerID     string
}

type resourceRepository struct {
	db *gorm.DB
}

// NewResourceRepository creates a new resource repository.
func NewResourceRepository(db *gorm.DB) ResourceRepository {
	return &resourceRepository{db: db}
}

func (r *resourceRepository) Create(ctx context.Context, resource *model.Resource) error {
	result := r.db.WithContext(ctx).Create(resource)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *resourceRepository) GetByID(ctx context.Context, id string) (*model.Resource, error) {
	var resource model.Resource
	result := r.db.WithContext(ctx).Preload("Owner").First(&resource, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, result.Error
	}
	return &resource, nil
}

func (r *resourceRepository) Update(ctx context.Context, resource *model.Resource) error {
	result := r.db.WithContext(ctx).Save(resource)
	return result.Error
}

func (r *resourceRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&model.Resource{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *resourceRepository) List(ctx context.Context, filters ResourceFilters, offset, limit int) ([]*model.Resource, int64, error) {
	var resources []*model.Resource
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Resource{})

	// Apply filters
	if filters.Type != "" {
		query = query.Where("type = ?", filters.Type)
	}
	if filters.Provider != "" {
		query = query.Where("provider = ?", filters.Provider)
	}
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Environment != "" {
		query = query.Where("environment = ?", filters.Environment)
	}
	if filters.OwnerID != "" {
		query = query.Where("owner_id = ?", filters.OwnerID)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	result := query.Preload("Owner").Offset(offset).Limit(limit).Find(&resources)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return resources, total, nil
}

// ResourceRequestRepository defines the interface for resource request data access.
type ResourceRequestRepository interface {
	Create(ctx context.Context, request *model.ResourceRequest) error
	GetByID(ctx context.Context, id string) (*model.ResourceRequest, error)
	Update(ctx context.Context, request *model.ResourceRequest) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filters RequestFilters, offset, limit int) ([]*model.ResourceRequest, int64, error)
}

// RequestFilters defines filters for request queries.
type RequestFilters struct {
	Status      string
	Environment string
	RequesterID string
}

type resourceRequestRepository struct {
	db *gorm.DB
}

// NewResourceRequestRepository creates a new resource request repository.
func NewResourceRequestRepository(db *gorm.DB) ResourceRequestRepository {
	return &resourceRequestRepository{db: db}
}

func (r *resourceRequestRepository) Create(ctx context.Context, request *model.ResourceRequest) error {
	result := r.db.WithContext(ctx).Create(request)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *resourceRequestRepository) GetByID(ctx context.Context, id string) (*model.ResourceRequest, error) {
	var request model.ResourceRequest
	result := r.db.WithContext(ctx).
		Preload("Requester").
		Preload("Approver").
		Preload("Region").
		Preload("Zone").
		Preload("Credential").
		Preload("Credential.Zone").
		Preload("TfProvider").
		Preload("TfProvider.Registry").
		Preload("TfModule").
		Preload("TfModule.Registry").
		Preload("TfModule.Provider").
		First(&request, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, result.Error
	}
	return &request, nil
}

func (r *resourceRequestRepository) Update(ctx context.Context, request *model.ResourceRequest) error {
	result := r.db.WithContext(ctx).Save(request)
	return result.Error
}

func (r *resourceRequestRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&model.ResourceRequest{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *resourceRequestRepository) List(ctx context.Context, filters RequestFilters, offset, limit int) ([]*model.ResourceRequest, int64, error) {
	var requests []*model.ResourceRequest
	var total int64

	query := r.db.WithContext(ctx).Model(&model.ResourceRequest{})

	// Apply filters
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Environment != "" {
		query = query.Where("environment = ?", filters.Environment)
	}
	if filters.RequesterID != "" {
		query = query.Where("requester_id = ?", filters.RequesterID)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results with all related data
	result := query.
		Preload("Requester").
		Preload("Approver").
		Preload("Region").
		Preload("Zone").
		Preload("Credential").
		Preload("Credential.Zone").
		Preload("TfProvider").
		Preload("TfProvider.Registry").
		Preload("TfModule").
		Preload("TfModule.Registry").
		Offset(offset).Limit(limit).Order("created_at DESC").Find(&requests)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return requests, total, nil
}
