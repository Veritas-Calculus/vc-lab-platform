// Package handler provides HTTP request handlers.
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

// IPAMHandler handles IP address management related HTTP requests.
type IPAMHandler struct {
	ipamService service.IPAMService
	logger      *zap.Logger
}

// NewIPAMHandler creates a new IPAM handler.
func NewIPAMHandler(ipamService service.IPAMService, logger *zap.Logger) *IPAMHandler {
	return &IPAMHandler{
		ipamService: ipamService,
		logger:      logger,
	}
}

// ListIPPools handles listing IP pools.
func (h *IPAMHandler) ListIPPools(c *gin.Context) {
	// Check if requesting all for dropdowns
	if c.Query("all") == constants.QueryTrue {
		pools, _, err := h.ipamService.ListPools(c.Request.Context(), "", 1, constants.MaxPageSize)
		if err != nil {
			h.logger.Error("failed to list IP pools", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list IP pools"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ip_pools": pools})
		return
	}

	zoneID := c.Query("zone_id")
	page := parseInt(c.DefaultQuery("page", "1"), 1)
	pageSize := parseInt(c.DefaultQuery("page_size", "20"), constants.DefaultPageSize)
	if pageSize > constants.MaxPageSize {
		pageSize = constants.MaxPageSize
	}

	pools, total, err := h.ipamService.ListPools(c.Request.Context(), zoneID, page, pageSize)
	if err != nil {
		h.logger.Error("failed to list IP pools", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list IP pools"})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize
	c.JSON(http.StatusOK, gin.H{
		"ip_pools":    pools,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// GetIPPool handles getting an IP pool by ID.
func (h *IPAMHandler) GetIPPool(c *gin.Context) {
	id := c.Param("id")
	pool, err := h.ipamService.GetPool(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "IP pool not found"})
			return
		}
		h.logger.Error("failed to get IP pool", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get IP pool"})
		return
	}

	// Get available count
	availableCount, err := h.ipamService.GetAvailableCount(c.Request.Context(), id)
	if err != nil {
		h.logger.Warn("failed to get available count", zap.Error(err))
	}

	c.JSON(http.StatusOK, gin.H{
		"pool":            pool,
		"available_count": availableCount,
	})
}

// CreateIPPoolRequest represents an IP pool creation request.
type CreateIPPoolRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=64"`
	CIDR        string `json:"cidr" binding:"required"`
	Gateway     string `json:"gateway" binding:"required"`
	DNS         string `json:"dns"`
	VLANTag     int    `json:"vlan_tag"`
	StartIP     string `json:"start_ip" binding:"required"`
	EndIP       string `json:"end_ip" binding:"required"`
	ZoneID      string `json:"zone_id" binding:"required"`
	NetworkType string `json:"network_type"`
	Description string `json:"description"`
}

// CreateIPPool handles creating an IP pool.
func (h *IPAMHandler) CreateIPPool(c *gin.Context) {
	var req CreateIPPoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pool, err := h.ipamService.CreatePool(c.Request.Context(), &service.CreateIPPoolInput{
		Name:        req.Name,
		CIDR:        req.CIDR,
		Gateway:     req.Gateway,
		DNS:         req.DNS,
		VLANTag:     req.VLANTag,
		StartIP:     req.StartIP,
		EndIP:       req.EndIP,
		ZoneID:      req.ZoneID,
		NetworkType: req.NetworkType,
		Description: req.Description,
	})
	if err != nil {
		h.logger.Error("failed to create IP pool", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, pool)
}

// UpdateIPPoolRequest represents an IP pool update request.
type UpdateIPPoolRequest struct {
	Name        *string `json:"name"`
	Gateway     *string `json:"gateway"`
	DNS         *string `json:"dns"`
	VLANTag     *int    `json:"vlan_tag"`
	Description *string `json:"description"`
	Status      *int8   `json:"status"`
}

// UpdateIPPool handles updating an IP pool.
func (h *IPAMHandler) UpdateIPPool(c *gin.Context) {
	id := c.Param("id")
	var req UpdateIPPoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pool, err := h.ipamService.UpdatePool(c.Request.Context(), id, &service.UpdateIPPoolInput{
		Name:        req.Name,
		Gateway:     req.Gateway,
		DNS:         req.DNS,
		VLANTag:     req.VLANTag,
		Description: req.Description,
		Status:      req.Status,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "IP pool not found"})
			return
		}
		h.logger.Error("failed to update IP pool", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pool)
}

// DeleteIPPool handles deleting an IP pool.
func (h *IPAMHandler) DeleteIPPool(c *gin.Context) {
	id := c.Param("id")
	if err := h.ipamService.DeletePool(c.Request.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "IP pool not found"})
			return
		}
		h.logger.Error("failed to delete IP pool", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "IP pool deleted successfully"})
}

// ListIPAllocations handles listing IP allocations for a pool.
func (h *IPAMHandler) ListIPAllocations(c *gin.Context) {
	poolID := c.Param("id")
	page := parseInt(c.DefaultQuery("page", "1"), 1)
	pageSize := parseInt(c.DefaultQuery("page_size", "50"), constants.DefaultPageSize)
	if pageSize > constants.MaxPageSize {
		pageSize = constants.MaxPageSize
	}

	allocations, total, err := h.ipamService.ListAllocations(c.Request.Context(), poolID, page, pageSize)
	if err != nil {
		h.logger.Error("failed to list IP allocations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list IP allocations"})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize
	c.JSON(http.StatusOK, gin.H{
		"allocations": allocations,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// AllocateIPRequest represents an IP allocation request.
type AllocateIPRequest struct {
	PoolID     string `json:"pool_id" binding:"required"`
	Hostname   string `json:"hostname"`
	ResourceID string `json:"resource_id"`
	IPAddress  string `json:"ip_address"` // Optional: specific IP to allocate
}

// AllocateIP handles allocating an IP address from a pool.
func (h *IPAMHandler) AllocateIP(c *gin.Context) {
	var req AllocateIPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	allocation, err := h.ipamService.AllocateIP(c.Request.Context(), &service.AllocateIPInput{
		PoolID:     req.PoolID,
		Hostname:   req.Hostname,
		ResourceID: req.ResourceID,
		IPAddress:  req.IPAddress,
	})
	if err != nil {
		h.logger.Error("failed to allocate IP", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, allocation)
}

// ReleaseIP handles releasing an allocated IP address.
func (h *IPAMHandler) ReleaseIP(c *gin.Context) {
	id := c.Param("id")
	if err := h.ipamService.ReleaseIP(c.Request.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "IP allocation not found"})
			return
		}
		h.logger.Error("failed to release IP", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to release IP"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "IP released successfully"})
}

// GetAllocationsByResource handles getting IP allocations for a resource.
func (h *IPAMHandler) GetAllocationsByResource(c *gin.Context) {
	resourceID := c.Param("resource_id")
	allocations, err := h.ipamService.GetAllocationsByResource(c.Request.Context(), resourceID)
	if err != nil {
		h.logger.Error("failed to get allocations by resource", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get IP allocations"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"allocations": allocations})
}
