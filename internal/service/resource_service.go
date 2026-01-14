// Package service provides business logic implementations.
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/Veritas-Calculus/vc-lab-platform/internal/constants"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/model"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/notification"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/repository"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/terraform"
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
	RetryRequest(ctx context.Context, id, userID string) (*model.ResourceRequest, error)
	DeleteRequest(ctx context.Context, id, userID string) error
}

// resourceService implements ResourceService.
type resourceService struct {
	resourceRepo        repository.ResourceRepository
	resourceRequestRepo repository.ResourceRequestRepository
	gitRepoRepo         repository.GitRepoRepository
	terraformExecutor   *terraform.Executor
	notificationService notification.Service
	logger              *zap.Logger
}

// NewResourceService creates a new resource service.
func NewResourceService(
	resourceRepo repository.ResourceRepository,
	resourceRequestRepo repository.ResourceRequestRepository,
	gitRepoRepo repository.GitRepoRepository,
	terraformExecutor *terraform.Executor,
	notificationService notification.Service,
	logger *zap.Logger,
) ResourceService {
	return &resourceService{
		resourceRepo:        resourceRepo,
		resourceRequestRepo: resourceRequestRepo,
		gitRepoRepo:         gitRepoRepo,
		terraformExecutor:   terraformExecutor,
		notificationService: notificationService,
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
	Title        string
	Description  string
	Type         string // vm, container, bare_metal
	Environment  string
	Provider     string
	RegionID     *string
	ZoneID       *string
	TfProviderID *string // Selected Terraform provider
	TfModuleID   *string // Selected Terraform module
	CredentialID *string // Selected credential for access
	Spec         string
	Quantity     int
	RequesterID  string
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
	if input.Type == "" {
		return nil, errors.New("type is required")
	}

	request := &model.ResourceRequest{
		Title:        input.Title,
		Description:  input.Description,
		Type:         input.Type,
		Environment:  input.Environment,
		Provider:     input.Provider,
		RegionID:     input.RegionID,
		ZoneID:       input.ZoneID,
		TfProviderID: input.TfProviderID,
		TfModuleID:   input.TfModuleID,
		CredentialID: input.CredentialID,
		Spec:         input.Spec,
		Quantity:     input.Quantity,
		RequesterID:  input.RequesterID,
		Status:       "pending",
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

// ApproveRequest approves a resource request and triggers provisioning.
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

	// Send approval notification
	if err := s.notificationService.NotifyResourceRequestApproved(ctx, request.RequesterID, request.ID, request.Title, reason); err != nil {
		s.logger.Error("failed to send approval notification", zap.Error(err))
	}

	// Start provisioning asynchronously
	go func() { //nolint:contextcheck // intentionally using background context for async operation
		bgCtx := context.Background()
		if err := s.provisionResource(bgCtx, request); err != nil {
			s.logger.Error("failed to provision resource", zap.String("request_id", request.ID), zap.Error(err))
		}
	}()

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
	request.RejectedAt = &now
	request.Reason = reason

	if err := s.resourceRequestRepo.Update(ctx, request); err != nil {
		s.logger.Error("failed to reject request", zap.Error(err))
		return nil, errors.New("failed to reject request")
	}

	// Send rejection notification
	if err := s.notificationService.NotifyResourceRequestRejected(ctx, request.RequesterID, request.ID, request.Title, reason); err != nil {
		s.logger.Error("failed to send rejection notification", zap.Error(err))
	}

	return s.resourceRequestRepo.GetByID(ctx, id)
}

// RetryRequest retries a failed resource request.
func (s *resourceService) RetryRequest(ctx context.Context, id, userID string) (*model.ResourceRequest, error) {
	if id == "" {
		return nil, errors.New("request ID cannot be empty")
	}

	request, err := s.resourceRequestRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	// Only failed requests can be retried
	if request.Status != "failed" {
		return nil, ErrInvalidRequestStatus
	}

	// Reset the request status to approved and clear error
	request.Status = "approved"
	request.ErrorMessage = ""
	request.ProvisionLog = ""
	request.ProvisionStartedAt = nil
	request.ProvisionCompletedAt = nil

	if err := s.resourceRequestRepo.Update(ctx, request); err != nil {
		s.logger.Error("failed to reset request for retry", zap.Error(err))
		return nil, errors.New("failed to reset request for retry")
	}

	s.logger.Info("retrying resource provisioning",
		zap.String("request_id", id),
		zap.String("user_id", userID),
	)

	// Start provisioning in background
	go func() { //nolint:contextcheck // intentionally using background context for async operation
		bgCtx := context.Background()
		if err := s.provisionResource(bgCtx, request); err != nil {
			s.logger.Error("resource provisioning retry failed",
				zap.String("request_id", id),
				zap.Error(err),
			)
		}
	}()

	return s.resourceRequestRepo.GetByID(ctx, id)
}

// DeleteRequest deletes a resource request.
func (s *resourceService) DeleteRequest(ctx context.Context, id, userID string) error {
	if id == "" {
		return errors.New("request ID cannot be empty")
	}

	request, err := s.resourceRequestRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return repository.ErrNotFound
		}
		return err
	}

	// Only pending, rejected, or failed requests can be deleted
	// Completed or provisioning requests cannot be deleted (resource already exists)
	if request.Status == "provisioning" || request.Status == "completed" {
		return errors.New("cannot delete request in current status")
	}

	s.logger.Info("deleting resource request",
		zap.String("request_id", id),
		zap.String("user_id", userID),
		zap.String("status", request.Status),
	)

	return s.resourceRequestRepo.Delete(ctx, id)
}

// provisionResource handles the Terraform provisioning workflow.
func (s *resourceService) provisionResource(ctx context.Context, request *model.ResourceRequest) error {
	s.logger.Info("starting resource provisioning", zap.String("request_id", request.ID))

	// Re-fetch the request with all relationships to ensure we have complete data
	fullRequest, err := s.resourceRequestRepo.GetByID(ctx, request.ID)
	if err != nil {
		s.logger.Error("failed to fetch request for provisioning", zap.Error(err))
		return s.handleProvisioningError(ctx, request, fmt.Errorf("failed to fetch request: %w", err))
	}
	request = fullRequest

	// Update status to provisioning
	now := time.Now()
	request.Status = "provisioning"
	request.ProvisionStartedAt = &now
	if err := s.resourceRequestRepo.Update(ctx, request); err != nil {
		s.logger.Error("failed to update request status to provisioning", zap.Error(err))
		return err
	}

	// Parse spec to get resource configuration
	var spec map[string]interface{}
	if err := json.Unmarshal([]byte(request.Spec), &spec); err != nil {
		return s.handleProvisioningError(ctx, request, fmt.Errorf("failed to parse spec: %w", err))
	}

	// Build Terraform config from request configuration
	tfConfig := s.buildTerraformConfig(ctx, request, spec)

	// Log the final TerraformConfig for debugging
	s.logger.Info("terraform config prepared",
		zap.String("registry_endpoint", tfConfig.RegistryEndpoint),
		zap.String("provider_source", tfConfig.ProviderSource),
		zap.String("module_source", tfConfig.ModuleSource),
		zap.String("cluster_endpoint", tfConfig.ClusterEndpoint),
		zap.String("git_host", tfConfig.GitHost),
	)

	// Execute Terraform workflow
	return s.executeTerraformWorkflow(ctx, request, tfConfig)
}

// buildTerraformConfig creates a Terraform configuration from the request.
func (s *resourceService) buildTerraformConfig(ctx context.Context, request *model.ResourceRequest, spec map[string]interface{}) terraform.Config {
	tfConfig := terraform.Config{
		Provider:    request.Provider,
		Environment: request.Environment,
		Spec:        spec,
	}

	s.logger.Info("provisioning configuration",
		zap.String("request_id", request.ID),
		zap.Bool("has_zone", request.Zone != nil),
		zap.Bool("has_tf_provider", request.TfProvider != nil),
		zap.Bool("has_tf_module", request.TfModule != nil),
	)

	// Get Provider configuration
	if request.TfProvider != nil {
		tfConfig.ProviderSource = request.TfProvider.Source
		tfConfig.ProviderNamespace = request.TfProvider.Namespace
		tfConfig.ProviderVersion = request.TfProvider.Version
		s.logger.Info("using tf provider",
			zap.String("source", request.TfProvider.Source),
			zap.String("version", request.TfProvider.Version),
		)
		if request.TfProvider.Registry != nil {
			tfConfig.RegistryEndpoint = request.TfProvider.Registry.Endpoint
			tfConfig.RegistryToken = request.TfProvider.Registry.Token
		}
	}

	// Get Module configuration
	if request.TfModule != nil {
		tfConfig.ModuleSource = request.TfModule.Source
		tfConfig.ModuleVersion = request.TfModule.Version
		s.logger.Info("using tf module",
			zap.String("source", request.TfModule.Source),
			zap.String("version", request.TfModule.Version),
		)
		if request.TfModule.Registry != nil && tfConfig.RegistryEndpoint == "" {
			tfConfig.RegistryEndpoint = request.TfModule.Registry.Endpoint
			tfConfig.RegistryToken = request.TfModule.Registry.Token
		}
	}

	// Get Credential configuration
	if request.Credential != nil {
		tfConfig.ClusterEndpoint = request.Credential.Endpoint
		tfConfig.ClusterUsername = request.Credential.AccessKey
		tfConfig.ClusterPassword = request.Credential.SecretKey
		tfConfig.ClusterToken = request.Credential.Token
	}

	// Configure Git authentication for module download
	if tfConfig.ModuleSource != "" {
		if err := s.configureGitAuth(ctx, &tfConfig); err != nil {
			s.logger.Warn("failed to configure git auth", zap.Error(err))
		}
	}

	return tfConfig
}

// executeTerraformWorkflow runs the Terraform init, plan, apply workflow.
//
//nolint:contextcheck // terraform executor methods don't use context
func (s *resourceService) executeTerraformWorkflow(ctx context.Context, request *model.ResourceRequest, tfConfig terraform.Config) error {
	workDir := fmt.Sprintf("/tmp/terraform/%s", request.ID)

	// Generate Terraform files
	if err := s.terraformExecutor.GenerateTFFiles(workDir, tfConfig); err != nil {
		return s.handleProvisioningError(ctx, request, fmt.Errorf("failed to generate terraform files: %w", err))
	}

	// Initialize Terraform with Git credentials
	if err := s.terraformExecutor.InitWithConfig(workDir, tfConfig); err != nil {
		return s.handleProvisioningError(ctx, request, fmt.Errorf("terraform init failed: %w", err))
	}

	// Plan
	planResult := s.terraformExecutor.Plan(workDir)
	provisionLog := fmt.Sprintf("=== Terraform Plan ===\n%s\n", planResult.Output)
	if !planResult.Success {
		return s.handleProvisioningError(ctx, request, fmt.Errorf("terraform plan failed: %s", planResult.Error))
	}

	// Apply
	applyResult := s.terraformExecutor.Apply(workDir)
	provisionLog += fmt.Sprintf("\n=== Terraform Apply ===\n%s\n", applyResult.Output)
	if !applyResult.Success {
		return s.handleProvisioningError(ctx, request, fmt.Errorf("terraform apply failed: %s", applyResult.Error))
	}

	// Get outputs and create resource record
	outputs := s.terraformExecutor.GetOutputs(workDir)
	outputsJSON, _ := json.Marshal(outputs) //nolint:errcheck // will not fail with map

	resourceName := fmt.Sprintf("%s-%s", request.Title, request.ID[:8])
	resource := &model.Resource{
		Name:        resourceName,
		Type:        request.Type,
		Provider:    request.Provider,
		Environment: request.Environment,
		Spec:        string(outputsJSON),
		Description: request.Description,
		OwnerID:     request.RequesterID,
		Status:      "running",
	}

	if err := s.resourceRepo.Create(ctx, resource); err != nil {
		s.logger.Error("failed to create resource record", zap.Error(err))
	}

	// Update request with completion status
	completedAt := time.Now()
	request.Status = "completed"
	request.ProvisionCompletedAt = &completedAt
	request.ProvisionLog = provisionLog
	request.TerraformState = "applied"
	request.ResourceID = &resource.ID

	if err := s.resourceRequestRepo.Update(ctx, request); err != nil {
		s.logger.Error("failed to update request completion status", zap.Error(err))
		return err
	}

	// Send success notification
	if err := s.notificationService.NotifyResourceProvisioned(ctx, request.RequesterID, request.ID, resourceName, outputs); err != nil {
		s.logger.Error("failed to send provisioning success notification", zap.Error(err))
	}

	s.logger.Info("resource provisioning completed", zap.String("request_id", request.ID), zap.String("resource_id", resource.ID))
	return nil
}

// configureGitAuth extracts Git host from module source and finds matching repository credentials.
// maxGitReposToSearch is the maximum number of git repos to search for credentials.
const maxGitReposToSearch = 100

func (s *resourceService) configureGitAuth(ctx context.Context, tfConfig *terraform.Config) error {
	if s.gitRepoRepo == nil {
		return nil // No git repo configured
	}

	moduleSource := tfConfig.ModuleSource
	if moduleSource == "" {
		return nil
	}

	// Extract host from module source URL using net/url
	host := extractHostFromURL(moduleSource)
	if host == "" {
		s.logger.Debug("could not extract host from module source", zap.String("source", moduleSource))
		return nil
	}

	s.logger.Info("looking for git credentials", zap.String("host", host))

	repos, _, err := s.gitRepoRepo.List(ctx, 1, maxGitReposToSearch)
	if err != nil {
		return fmt.Errorf("failed to list git repos: %w", err)
	}

	for _, repo := range repos {
		if repo.URL == "" {
			continue
		}
		repoHost := extractHostFromURL(repo.URL)
		if repoHost != host {
			continue
		}
		tfConfig.GitHost = host
		tfConfig.GitUsername = repo.Username
		if tfConfig.GitUsername == "" {
			tfConfig.GitUsername = "git" // Default username for token auth
		}
		tfConfig.GitToken = repo.Token
		s.logger.Info("found git credentials for module download",
			zap.String("host", host),
			zap.String("repo_name", repo.Name),
			zap.Bool("has_token", repo.Token != ""),
		)
		return nil
	}

	s.logger.Warn("no git credentials found for module host", zap.String("host", host))
	return nil
}

// extractHostFromURL extracts the host from a URL string.
func extractHostFromURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return parsed.Host
}

// handleProvisioningError updates the request with error status and sends notification.
func (s *resourceService) handleProvisioningError(ctx context.Context, request *model.ResourceRequest, err error) error {
	s.logger.Error("provisioning failed", zap.String("request_id", request.ID), zap.Error(err))

	request.Status = "failed"
	request.ErrorMessage = err.Error()
	if updateErr := s.resourceRequestRepo.Update(ctx, request); updateErr != nil {
		s.logger.Error("failed to update request error status", zap.Error(updateErr))
	}

	// Send failure notification
	if notifyErr := s.notificationService.NotifyResourceProvisioningFailed(ctx, request.RequesterID, request.ID, request.Title, err.Error()); notifyErr != nil {
		s.logger.Error("failed to send provisioning failure notification", zap.Error(notifyErr))
	}

	return err
}
