// Package handler provides HTTP request handlers.
package handler

import (
	"errors"
	"net/http"

	"github.com/Veritas-Calculus/vc-lab-platform/internal/repository"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ResourceHandler handles resource management requests.
type ResourceHandler struct {
	resourceService service.ResourceService
	logger          *zap.Logger
}

// NewResourceHandler creates a new resource handler.
func NewResourceHandler(resourceService service.ResourceService, logger *zap.Logger) *ResourceHandler {
	return &ResourceHandler{
		resourceService: resourceService,
		logger:          logger,
	}
}

// List handles listing resources.
func (h *ResourceHandler) List(c *gin.Context) {
	page := parseInt(c.DefaultQuery("page", "1"), 1)
	pageSize := parseInt(c.DefaultQuery("page_size", "20"), 20)

	if pageSize > 100 {
		pageSize = 100
	}

	filters := service.ResourceFilters{
		Type:        c.Query("type"),
		Provider:    c.Query("provider"),
		Status:      c.Query("status"),
		Environment: c.Query("environment"),
		OwnerID:     c.Query("owner_id"),
	}

	resources, total, err := h.resourceService.List(c.Request.Context(), filters, page, pageSize)
	if err != nil {
		h.logger.Error("failed to list resources", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list resources"})
		return
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	c.JSON(http.StatusOK, gin.H{
		"resources":   resources,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// CreateResourceRequest represents a resource creation request.
type CreateResourceRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Type        string `json:"type" binding:"required,oneof=vm container bare_metal"`
	Provider    string `json:"provider" binding:"required,oneof=pve vmware openstack aws aliyun"`
	Environment string `json:"environment" binding:"required,oneof=dev test staging prod"`
	Spec        string `json:"spec"`
	Description string `json:"description"`
}

// Create handles resource creation.
func (h *ResourceHandler) Create(c *gin.Context) {
	var req CreateResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")

	resource, err := h.resourceService.Create(c.Request.Context(), &service.CreateResourceInput{
		Name:        req.Name,
		Type:        req.Type,
		Provider:    req.Provider,
		Environment: req.Environment,
		Spec:        req.Spec,
		Description: req.Description,
		OwnerID:     userID.(string),
	})
	if err != nil {
		h.logger.Error("failed to create resource", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create resource"})
		return
	}

	c.JSON(http.StatusCreated, resource)
}

// GetByID handles getting a resource by ID.
func (h *ResourceHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Resource ID required"})
		return
	}

	resource, err := h.resourceService.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
			return
		}
		h.logger.Error("failed to get resource", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get resource"})
		return
	}

	c.JSON(http.StatusOK, resource)
}

// Update handles resource updates.
func (h *ResourceHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Resource ID required"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resource, err := h.resourceService.Update(c.Request.Context(), id, req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
			return
		}
		h.logger.Error("failed to update resource", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update resource"})
		return
	}

	c.JSON(http.StatusOK, resource)
}

// Delete handles resource deletion.
func (h *ResourceHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Resource ID required"})
		return
	}

	if err := h.resourceService.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
			return
		}
		h.logger.Error("failed to delete resource", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete resource"})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListRequests handles listing resource requests.
func (h *ResourceHandler) ListRequests(c *gin.Context) {
	page := parseInt(c.DefaultQuery("page", "1"), 1)
	pageSize := parseInt(c.DefaultQuery("page_size", "20"), 20)

	if pageSize > 100 {
		pageSize = 100
	}

	filters := service.RequestFilters{
		Status:      c.Query("status"),
		Environment: c.Query("environment"),
		RequesterID: c.Query("requester_id"),
	}

	requests, total, err := h.resourceService.ListRequests(c.Request.Context(), filters, page, pageSize)
	if err != nil {
		h.logger.Error("failed to list requests", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list requests"})
		return
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	c.JSON(http.StatusOK, gin.H{
		"requests":    requests,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// CreateRequestRequest represents a resource request creation.
type CreateRequestRequest struct {
	Title       string `json:"title" binding:"required,min=1,max=200"`
	Description string `json:"description"`
	Environment string `json:"environment" binding:"required,oneof=dev test staging prod"`
	Provider    string `json:"provider" binding:"required,oneof=pve vmware openstack aws aliyun"`
	Spec        string `json:"spec"`
	Quantity    int    `json:"quantity"`
}

// CreateRequest handles resource request creation.
func (h *ResourceHandler) CreateRequest(c *gin.Context) {
	var req CreateRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")

	quantity := req.Quantity
	if quantity < 1 {
		quantity = 1
	}

	request, err := h.resourceService.CreateRequest(c.Request.Context(), &service.CreateRequestInput{
		Title:       req.Title,
		Description: req.Description,
		Environment: req.Environment,
		Provider:    req.Provider,
		Spec:        req.Spec,
		Quantity:    quantity,
		RequesterID: userID.(string),
	})
	if err != nil {
		h.logger.Error("failed to create request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	c.JSON(http.StatusCreated, request)
}

// GetRequest handles getting a resource request by ID.
func (h *ResourceHandler) GetRequest(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request ID required"})
		return
	}

	request, err := h.resourceService.GetRequest(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Request not found"})
			return
		}
		h.logger.Error("failed to get request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get request"})
		return
	}

	c.JSON(http.StatusOK, request)
}

// ApproveRequestBody represents an approval request body.
type ApproveRequestBody struct {
	Reason string `json:"reason"`
}

// ApproveRequest handles request approval.
func (h *ResourceHandler) ApproveRequest(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request ID required"})
		return
	}

	var body ApproveRequestBody
	c.ShouldBindJSON(&body)

	userID, _ := c.Get("user_id")

	request, err := h.resourceService.ApproveRequest(c.Request.Context(), id, userID.(string), body.Reason)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Request not found"})
			return
		}
		if errors.Is(err, service.ErrInvalidRequestStatus) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Request cannot be approved"})
			return
		}
		h.logger.Error("failed to approve request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve request"})
		return
	}

	c.JSON(http.StatusOK, request)
}

// RejectRequestBody represents a rejection request body.
type RejectRequestBody struct {
	Reason string `json:"reason" binding:"required"`
}

// RejectRequest handles request rejection.
func (h *ResourceHandler) RejectRequest(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request ID required"})
		return
	}

	var body RejectRequestBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reason is required"})
		return
	}

	userID, _ := c.Get("user_id")

	request, err := h.resourceService.RejectRequest(c.Request.Context(), id, userID.(string), body.Reason)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Request not found"})
			return
		}
		if errors.Is(err, service.ErrInvalidRequestStatus) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Request cannot be rejected"})
			return
		}
		h.logger.Error("failed to reject request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject request"})
		return
	}

	c.JSON(http.StatusOK, request)
}
