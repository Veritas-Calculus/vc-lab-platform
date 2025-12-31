// Package service provides business logic implementations.
package service

import (
	"context"
	"errors"
	"time"

	"github.com/Veritas-Calculus/vc-lab-platform/internal/constants"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/model"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/repository"
	"go.uber.org/zap"
)

// ErrInvalidRequestStatus indicates an invalid request status transition.
var ErrInvalidRequestStatus = errors.New("invalid request status")

// ResourceService provides resource-related business operations.
type ResourceService interface {
	// Resource operations
	Create(ctx context.Context, input *CreateResourceInput) (*model.Resource, error)
	GetByID(ctx context.Context, id string) (*model.Resource, error)
	List(ctx context.Context, filters ResourceFilters, page, pageSize int) ([]*model.Resource, int64, error)
	Update(ctx context.Context, id string, updates map[string]interface{}) (*model.Resource, error)
	Delete(ctx context.Context, id string) error

	// Resource request operations
	CreateRequest(ctx context.Context, input *CreateRequestInput) (*model.ResourceRequest, error)
	GetRequest(ctx context.Context, id string) (*model.ResourceRequest, error)
	ListRequests(ctx context.Context, filters RequestFilters, page, pageSize int) ([]*model.ResourceRequest, int64, error)
	ApproveRequest(ctx context.Context, id, approverID, reason string) (*model.ResourceRequest, error)
	RejectRequest(ctx context.Context, id, approverID, reason string) (*model.ResourceRequest, error)
}

// resourceService implements ResourceService.
type resourceService struct {
	resourceRepo        repository.ResourceRepository
	resourceRequestRepo repository.ResourceRequestRepository
	logger              *zap.Logger
}

// NewResourceService creates a new resource service.
func NewResourceService(resourceRepo repository.ResourceRepository, resourceRequestRepo repository.ResourceRequestRepository, logger *zap.Logger) ResourceService {
	return &resourceService{
		resourceRepo:        resourceRepo,
		resourceRequestRepo: resourceRequestRepo,
		logger:              logger,
	}
}

// CreateResourceInput represents input for resource creation.
type CreateResourceInput struct {
	Name        string
	Type        string
	Provider    string
	Environment string
	Spec        string
	Description string
	OwnerID     string
}

// ResourceFilters represents filters for resource listing.
type ResourceFilters struct {
	Type        string
	Provider    string
	Status      string
	Environment string
	OwnerID     string
}

// CreateRequestInput represents input for resource request creation.
type CreateRequestInput struct {
	Title       string
	Description string
	Environment string
	Provider    string
	Spec        string
	Quantity    int
	RequesterID string
}

// RequestFilters represents filters for request listing.
type RequestFilters struct {
	Status      string
	Environment string
	RequesterID string
}

// Create creates a new resource.
func (s *resourceService) Create(ctx context.Context, input *CreateResourceInput) (*model.Resource, error) {
	if input == nil {
		return nil, errors.New("input cannot be nil")
	}
	if input.Name == "" {
		return nil, errors.New("name is required")
	}
	if input.Type == "" {
		return nil, errors.New("type is required")
	}
	if input.Provider == "" {
		return nil, errors.New("provider is required")
	}

	resource := &model.Resource{
		Name:        input.Name,
		Type:        input.Type,
		Provider:    input.Provider,
		Environment: input.Environment,
		Spec:        input.Spec,
		Description: input.Description,
		OwnerID:     input.OwnerID,
		Status:      "active",
	}

	if err := s.resourceRepo.Create(ctx, resource); err != nil {
		s.logger.Error("failed to create resource", zap.Error(err))
		return nil, errors.New("failed to create resource")
	}

	return resource, nil
}

// GetByID gets a resource by ID.
func (s *resourceService) GetByID(ctx context.Context, id string) (*model.Resource, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	resource, err := s.resourceRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, repository.ErrNotFound
		}
		s.logger.Error("failed to get resource", zap.Error(err))
		return nil, errors.New("failed to get resource")
	}

	return resource, nil
}

// List lists resources with filters and pagination.
func (s *resourceService) List(ctx context.Context, filters ResourceFilters, page, pageSize int) ([]*model.Resource, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = constants.DefaultPageSize
	}
	if pageSize > constants.MaxPageSize {
		pageSize = constants.MaxPageSize
	}

	offset := (page - 1) * pageSize

	repoFilters := repository.ResourceFilters{
		Type:        filters.Type,
		Provider:    filters.Provider,
		Status:      filters.Status,
		Environment: filters.Environment,
		OwnerID:     filters.OwnerID,
	}

	return s.resourceRepo.List(ctx, repoFilters, offset, pageSize)
}

// Update updates a resource.
func (s *resourceService) Update(ctx context.Context, id string, updates map[string]interface{}) (*model.Resource, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	// Verify resource exists
	resource, err := s.resourceRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	// Filter allowed updates and apply to resource
	if name, ok := updates["name"].(string); ok && name != "" {
		resource.Name = name
	}
	if desc, ok := updates["description"].(string); ok {
		resource.Description = desc
	}
	if spec, ok := updates["spec"].(string); ok {
		resource.Spec = spec
	}
	if status, ok := updates["status"].(string); ok && status != "" {
		resource.Status = status
	}
	if tags, ok := updates["tags"].(string); ok {
		resource.Tags = tags
	}

	if err := s.resourceRepo.Update(ctx, resource); err != nil {
		s.logger.Error("failed to update resource", zap.Error(err))
		return nil, errors.New("failed to update resource")
	}

	return s.resourceRepo.GetByID(ctx, id)
}

// Delete deletes a resource.
func (s *resourceService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}

	// Verify resource exists
	_, err := s.resourceRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return repository.ErrNotFound
		}
		return err
	}

	if err := s.resourceRepo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete resource", zap.Error(err))
		return errors.New("failed to delete resource")
	}

	return nil
}

// CreateRequest creates a resource request.
func (s *resourceService) CreateRequest(ctx context.Context, input *CreateRequestInput) (*model.ResourceRequest, error) {
	if input == nil {
		return nil, errors.New("input cannot be nil")
	}
	if input.Title == "" {
		return nil, errors.New("title is required")
	}
	if input.RequesterID == "" {
		return nil, errors.New("requester ID is required")
	}

	request := &model.ResourceRequest{
		Title:       input.Title,
		Description: input.Description,
		Environment: input.Environment,
		Provider:    input.Provider,
		Spec:        input.Spec,
		Quantity:    input.Quantity,
		RequesterID: input.RequesterID,
		Status:      "pending",
	}

	if err := s.resourceRequestRepo.Create(ctx, request); err != nil {
		s.logger.Error("failed to create request", zap.Error(err))
		return nil, errors.New("failed to create request")
	}

	return request, nil
}

// GetRequest gets a resource request by ID.
func (s *resourceService) GetRequest(ctx context.Context, id string) (*model.ResourceRequest, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	request, err := s.resourceRequestRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, repository.ErrNotFound
		}
		s.logger.Error("failed to get request", zap.Error(err))
		return nil, errors.New("failed to get request")
	}

	return request, nil
}

// ListRequests lists resource requests with filters.
func (s *resourceService) ListRequests(ctx context.Context, filters RequestFilters, page, pageSize int) ([]*model.ResourceRequest, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = constants.DefaultPageSize
	}
	if pageSize > constants.MaxPageSize {
		pageSize = constants.MaxPageSize
	}

	offset := (page - 1) * pageSize

	repoFilters := repository.RequestFilters{
		Status:      filters.Status,
		Environment: filters.Environment,
		RequesterID: filters.RequesterID,
	}

	return s.resourceRequestRepo.List(ctx, repoFilters, offset, pageSize)
}

// ApproveRequest approves a resource request.
func (s *resourceService) ApproveRequest(ctx context.Context, id, approverID, reason string) (*model.ResourceRequest, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}
	if approverID == "" {
		return nil, errors.New("approver ID cannot be empty")
	}

	request, err := s.resourceRequestRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	if request.Status != "pending" {
		return nil, ErrInvalidRequestStatus
	}

	now := time.Now()
	request.Status = "approved"
	request.ApproverID = &approverID
	request.ApprovedAt = &now
	request.Reason = reason

	if err := s.resourceRequestRepo.Update(ctx, request); err != nil {
		s.logger.Error("failed to approve request", zap.Error(err))
		return nil, errors.New("failed to approve request")
	}

	return s.resourceRequestRepo.GetByID(ctx, id)
}

// RejectRequest rejects a resource request.
func (s *resourceService) RejectRequest(ctx context.Context, id, approverID, reason string) (*model.ResourceRequest, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}
	if approverID == "" {
		return nil, errors.New("approver ID cannot be empty")
	}
	if reason == "" {
		return nil, errors.New("reason is required")
	}

	request, err := s.resourceRequestRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	if request.Status != "pending" {
		return nil, ErrInvalidRequestStatus
	}

	now := time.Now()
	request.Status = "rejected"
	request.ApproverID = &approverID
	request.ApprovedAt = &now
	request.Reason = reason

	if err := s.resourceRequestRepo.Update(ctx, request); err != nil {
		s.logger.Error("failed to reject request", zap.Error(err))
		return nil, errors.New("failed to reject request")
	}

	return s.resourceRequestRepo.GetByID(ctx, id)
}
