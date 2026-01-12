// Package handler provides HTTP request handlers.
//
//nolint:dupl // similar handler patterns are acceptable
package handler

import (
	"errors"
	"net/http"

	"github.com/Veritas-Calculus/vc-lab-platform/internal/constants"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/repository"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// InfraHandler handles infrastructure management requests.
type InfraHandler struct {
	infraService service.InfraService
	logger       *zap.Logger
}

// NewInfraHandler creates a new infrastructure handler.
func NewInfraHandler(infraService service.InfraService, logger *zap.Logger) *InfraHandler {
	return &InfraHandler{
		infraService: infraService,
		logger:       logger,
	}
}

// ListRegions handles listing regions.
func (h *InfraHandler) ListRegions(c *gin.Context) {
	// Check if requesting all regions (for dropdowns)
	if c.Query("all") == constants.QueryTrue {
		regions, err := h.infraService.ListAllRegions(c.Request.Context())
		if err != nil {
			h.logger.Error("failed to list all regions", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list regions"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"regions": regions})
		return
	}

	page := parseInt(c.DefaultQuery("page", "1"), 1)
	pageSize := parseInt(c.DefaultQuery("page_size", "20"), constants.DefaultPageSize)
	if pageSize > constants.MaxPageSize {
		pageSize = constants.MaxPageSize
	}

	regions, total, err := h.infraService.ListRegions(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("failed to list regions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list regions"})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize
	c.JSON(http.StatusOK, gin.H{
		"regions":     regions,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// CreateRegionRequest represents a region creation request.
type CreateRegionRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=64"`
	Code        string `json:"code" binding:"required,min=1,max=32"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
}

// CreateRegion handles creating a region.
func (h *InfraHandler) CreateRegion(c *gin.Context) {
	var req CreateRegionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	region, err := h.infraService.CreateRegion(c.Request.Context(), &service.CreateRegionInput{
		Name:        req.Name,
		Code:        req.Code,
		DisplayName: req.DisplayName,
		Description: req.Description,
	})
	if err != nil {
		h.logger.Error("failed to create region", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, region)
}

// GetRegion handles getting a region by ID.
func (h *InfraHandler) GetRegion(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Region ID required"})
		return
	}

	region, err := h.infraService.GetRegion(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Region not found"})
			return
		}
		h.logger.Error("failed to get region", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get region"})
		return
	}

	c.JSON(http.StatusOK, region)
}

// UpdateRegionRequest represents a region update request.
type UpdateRegionRequest struct {
	Name        *string `json:"name"`
	DisplayName *string `json:"display_name"`
	Description *string `json:"description"`
	Status      *int8   `json:"status"`
}

// UpdateRegion handles updating a region.
func (h *InfraHandler) UpdateRegion(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Region ID required"})
		return
	}

	var req UpdateRegionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	region, err := h.infraService.UpdateRegion(c.Request.Context(), id, &service.UpdateRegionInput{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Status:      req.Status,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Region not found"})
			return
		}
		h.logger.Error("failed to update region", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update region"})
		return
	}

	c.JSON(http.StatusOK, region)
}

// DeleteRegion handles deleting a region.
func (h *InfraHandler) DeleteRegion(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Region ID required"})
		return
	}

	if err := h.infraService.DeleteRegion(c.Request.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Region not found"})
			return
		}
		h.logger.Error("failed to delete region", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete region"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Region deleted successfully"})
}

// ListZones handles listing zones.
func (h *InfraHandler) ListZones(c *gin.Context) {
	// Check if filtering by region
	regionID := c.Query("region_id")
	if regionID != "" {
		zones, err := h.infraService.ListZonesByRegion(c.Request.Context(), regionID)
		if err != nil {
			h.logger.Error("failed to list zones by region", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list zones"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"zones": zones})
		return
	}

	page := parseInt(c.DefaultQuery("page", "1"), 1)
	pageSize := parseInt(c.DefaultQuery("page_size", "20"), constants.DefaultPageSize)
	if pageSize > constants.MaxPageSize {
		pageSize = constants.MaxPageSize
	}

	zones, total, err := h.infraService.ListZones(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("failed to list zones", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list zones"})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize
	c.JSON(http.StatusOK, gin.H{
		"zones":       zones,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// CreateZoneRequest represents a zone creation request.
type CreateZoneRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=64"`
	Code        string `json:"code" binding:"required,min=1,max=32"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	RegionID    string `json:"region_id" binding:"required"`
	IsDefault   bool   `json:"is_default"`
}

// CreateZone handles creating a zone.
func (h *InfraHandler) CreateZone(c *gin.Context) {
	var req CreateZoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	zone, err := h.infraService.CreateZone(c.Request.Context(), &service.CreateZoneInput{
		Name:        req.Name,
		Code:        req.Code,
		DisplayName: req.DisplayName,
		Description: req.Description,
		RegionID:    req.RegionID,
		IsDefault:   req.IsDefault,
	})
	if err != nil {
		h.logger.Error("failed to create zone", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, zone)
}

// GetZone handles getting a zone by ID.
func (h *InfraHandler) GetZone(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Zone ID required"})
		return
	}

	zone, err := h.infraService.GetZone(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Zone not found"})
			return
		}
		h.logger.Error("failed to get zone", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get zone"})
		return
	}

	c.JSON(http.StatusOK, zone)
}

// UpdateZoneRequest represents a zone update request.
type UpdateZoneRequest struct {
	Name        *string `json:"name"`
	DisplayName *string `json:"display_name"`
	Description *string `json:"description"`
	Status      *int8   `json:"status"`
	IsDefault   *bool   `json:"is_default"`
}

// UpdateZone handles updating a zone.
func (h *InfraHandler) UpdateZone(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Zone ID required"})
		return
	}

	var req UpdateZoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	zone, err := h.infraService.UpdateZone(c.Request.Context(), id, &service.UpdateZoneInput{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Status:      req.Status,
		IsDefault:   req.IsDefault,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Zone not found"})
			return
		}
		h.logger.Error("failed to update zone", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update zone"})
		return
	}

	c.JSON(http.StatusOK, zone)
}

// DeleteZone handles deleting a zone.
func (h *InfraHandler) DeleteZone(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Zone ID required"})
		return
	}

	if err := h.infraService.DeleteZone(c.Request.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Zone not found"})
			return
		}
		h.logger.Error("failed to delete zone", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete zone"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Zone deleted successfully"})
}

// Terraform Registry handlers

// ListRegistries handles listing terraform registries.
func (h *InfraHandler) ListRegistries(c *gin.Context) {
	if c.Query("all") == constants.QueryTrue {
		registries, err := h.infraService.ListAllRegistries(c.Request.Context())
		if err != nil {
			h.logger.Error("failed to list all registries", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list registries"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"registries": registries})
		return
	}

	page := parseInt(c.DefaultQuery("page", "1"), 1)
	pageSize := parseInt(c.DefaultQuery("page_size", "20"), constants.DefaultPageSize)
	if pageSize > constants.MaxPageSize {
		pageSize = constants.MaxPageSize
	}

	registries, total, err := h.infraService.ListRegistries(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("failed to list registries", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list registries"})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize
	c.JSON(http.StatusOK, gin.H{
		"registries":  registries,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// CreateRegistryRequest represents a registry creation request.
type CreateRegistryRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=128"`
	Endpoint    string `json:"endpoint" binding:"required"`
	Username    string `json:"username"`
	Token       string `json:"token"`
	Description string `json:"description"`
	IsDefault   bool   `json:"is_default"`
}

// CreateRegistry handles creating a terraform registry.
func (h *InfraHandler) CreateRegistry(c *gin.Context) {
	var req CreateRegistryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	registry, err := h.infraService.CreateRegistry(c.Request.Context(), &service.CreateRegistryInput{
		Name:        req.Name,
		Endpoint:    req.Endpoint,
		Username:    req.Username,
		Token:       req.Token,
		Description: req.Description,
		IsDefault:   req.IsDefault,
	})
	if err != nil {
		h.logger.Error("failed to create registry", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, registry)
}

// GetRegistry handles getting a registry by ID.
func (h *InfraHandler) GetRegistry(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Registry ID required"})
		return
	}

	registry, err := h.infraService.GetRegistry(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Registry not found"})
			return
		}
		h.logger.Error("failed to get registry", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get registry"})
		return
	}

	c.JSON(http.StatusOK, registry)
}

// UpdateRegistryRequest represents a registry update request.
type UpdateRegistryRequest struct {
	Name        *string `json:"name"`
	Endpoint    *string `json:"endpoint"`
	Username    *string `json:"username"`
	Token       *string `json:"token"`
	Description *string `json:"description"`
	Status      *int8   `json:"status"`
	IsDefault   *bool   `json:"is_default"`
}

// UpdateRegistry handles updating a registry.
func (h *InfraHandler) UpdateRegistry(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Registry ID required"})
		return
	}

	var req UpdateRegistryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	registry, err := h.infraService.UpdateRegistry(c.Request.Context(), id, &service.UpdateRegistryInput{
		Name:        req.Name,
		Endpoint:    req.Endpoint,
		Username:    req.Username,
		Token:       req.Token,
		Description: req.Description,
		Status:      req.Status,
		IsDefault:   req.IsDefault,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Registry not found"})
			return
		}
		h.logger.Error("failed to update registry", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update registry"})
		return
	}

	c.JSON(http.StatusOK, registry)
}

// DeleteRegistry handles deleting a registry.
func (h *InfraHandler) DeleteRegistry(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Registry ID required"})
		return
	}

	if err := h.infraService.DeleteRegistry(c.Request.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Registry not found"})
			return
		}
		h.logger.Error("failed to delete registry", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete registry"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Registry deleted successfully"})
}

// Terraform Provider handlers

// ListProviders handles listing terraform providers.
func (h *InfraHandler) ListProviders(c *gin.Context) {
	registryID := c.Query("registry_id")
	if registryID != "" {
		providers, err := h.infraService.ListProvidersByRegistry(c.Request.Context(), registryID)
		if err != nil {
			h.logger.Error("failed to list providers by registry", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list providers"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"providers": providers})
		return
	}

	page := parseInt(c.DefaultQuery("page", "1"), 1)
	pageSize := parseInt(c.DefaultQuery("page_size", "20"), constants.DefaultPageSize)
	if pageSize > constants.MaxPageSize {
		pageSize = constants.MaxPageSize
	}

	providers, total, err := h.infraService.ListProviders(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("failed to list providers", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list providers"})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize
	c.JSON(http.StatusOK, gin.H{
		"providers":   providers,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// CreateTfProviderRequest represents a provider creation request.
type CreateTfProviderRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=128"`
	Namespace   string `json:"namespace"`
	Source      string `json:"source"`
	Version     string `json:"version"`
	RegistryID  string `json:"registry_id" binding:"required"`
	Description string `json:"description"`
}

// CreateProvider handles creating a terraform provider.
func (h *InfraHandler) CreateProvider(c *gin.Context) {
	var req CreateTfProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	provider, err := h.infraService.CreateProvider(c.Request.Context(), &service.CreateTfProviderInput{
		Name:        req.Name,
		Namespace:   req.Namespace,
		Source:      req.Source,
		Version:     req.Version,
		RegistryID:  req.RegistryID,
		Description: req.Description,
	})
	if err != nil {
		h.logger.Error("failed to create provider", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, provider)
}

// GetProvider handles getting a provider by ID.
func (h *InfraHandler) GetProvider(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provider ID required"})
		return
	}

	provider, err := h.infraService.GetProvider(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Provider not found"})
			return
		}
		h.logger.Error("failed to get provider", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get provider"})
		return
	}

	c.JSON(http.StatusOK, provider)
}

// UpdateTfProviderRequest represents a provider update request.
type UpdateTfProviderRequest struct {
	Name        *string `json:"name"`
	Namespace   *string `json:"namespace"`
	Source      *string `json:"source"`
	Version     *string `json:"version"`
	Description *string `json:"description"`
	Status      *int8   `json:"status"`
}

// UpdateProvider handles updating a provider.
func (h *InfraHandler) UpdateProvider(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provider ID required"})
		return
	}

	var req UpdateTfProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	provider, err := h.infraService.UpdateProvider(c.Request.Context(), id, &service.UpdateTfProviderInput{
		Name:        req.Name,
		Namespace:   req.Namespace,
		Source:      req.Source,
		Version:     req.Version,
		Description: req.Description,
		Status:      req.Status,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Provider not found"})
			return
		}
		h.logger.Error("failed to update provider", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update provider"})
		return
	}

	c.JSON(http.StatusOK, provider)
}

// DeleteProvider handles deleting a provider.
func (h *InfraHandler) DeleteProvider(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provider ID required"})
		return
	}

	if err := h.infraService.DeleteProvider(c.Request.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Provider not found"})
			return
		}
		h.logger.Error("failed to delete provider", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete provider"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Provider deleted successfully"})
}

// Terraform Module handlers

// ListModules handles listing terraform modules.
func (h *InfraHandler) ListModules(c *gin.Context) {
	if c.Query("all") == constants.QueryTrue {
		modules, err := h.infraService.ListAllModules(c.Request.Context())
		if err != nil {
			h.logger.Error("failed to list all modules", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list modules"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"modules": modules})
		return
	}

	page := parseInt(c.DefaultQuery("page", "1"), 1)
	pageSize := parseInt(c.DefaultQuery("page_size", "20"), constants.DefaultPageSize)
	if pageSize > constants.MaxPageSize {
		pageSize = constants.MaxPageSize
	}

	modules, total, err := h.infraService.ListModules(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("failed to list modules", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list modules"})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize
	c.JSON(http.StatusOK, gin.H{
		"modules":     modules,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// CreateModuleRequest represents a module creation request.
type CreateModuleRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=128"`
	Source      string  `json:"source" binding:"required"`
	Version     string  `json:"version"`
	RegistryID  *string `json:"registry_id"`
	ProviderID  *string `json:"provider_id"`
	Description string  `json:"description"`
	Variables   string  `json:"variables"`
}

// CreateModule handles creating a terraform module.
func (h *InfraHandler) CreateModule(c *gin.Context) {
	var req CreateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	module, err := h.infraService.CreateModule(c.Request.Context(), &service.CreateModuleInput{
		Name:        req.Name,
		Source:      req.Source,
		Version:     req.Version,
		RegistryID:  req.RegistryID,
		ProviderID:  req.ProviderID,
		Description: req.Description,
		Variables:   req.Variables,
	})
	if err != nil {
		h.logger.Error("failed to create module", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, module)
}

// GetModule handles getting a module by ID.
func (h *InfraHandler) GetModule(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Module ID required"})
		return
	}

	module, err := h.infraService.GetModule(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Module not found"})
			return
		}
		h.logger.Error("failed to get module", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get module"})
		return
	}

	c.JSON(http.StatusOK, module)
}

// UpdateModuleRequest represents a module update request.
type UpdateModuleRequest struct {
	Name        *string `json:"name"`
	Source      *string `json:"source"`
	Version     *string `json:"version"`
	RegistryID  *string `json:"registry_id"`
	ProviderID  *string `json:"provider_id"`
	Description *string `json:"description"`
	Variables   *string `json:"variables"`
	Status      *int8   `json:"status"`
}

// UpdateModule handles updating a module.
func (h *InfraHandler) UpdateModule(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Module ID required"})
		return
	}

	var req UpdateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	module, err := h.infraService.UpdateModule(c.Request.Context(), id, &service.UpdateModuleInput{
		Name:        req.Name,
		Source:      req.Source,
		Version:     req.Version,
		RegistryID:  req.RegistryID,
		ProviderID:  req.ProviderID,
		Description: req.Description,
		Variables:   req.Variables,
		Status:      req.Status,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Module not found"})
			return
		}
		h.logger.Error("failed to update module", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update module"})
		return
	}

	c.JSON(http.StatusOK, module)
}

// DeleteModule handles deleting a module.
func (h *InfraHandler) DeleteModule(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Module ID required"})
		return
	}

	if err := h.infraService.DeleteModule(c.Request.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Module not found"})
			return
		}
		h.logger.Error("failed to delete module", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete module"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Module deleted successfully"})
}
