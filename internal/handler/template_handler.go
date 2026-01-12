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

// VMTemplateHandler handles VM template related HTTP requests.
type VMTemplateHandler struct {
	templateService service.VMTemplateService
	logger          *zap.Logger
}

// NewVMTemplateHandler creates a new VM template handler.
func NewVMTemplateHandler(templateService service.VMTemplateService, logger *zap.Logger) *VMTemplateHandler {
	return &VMTemplateHandler{
		templateService: templateService,
		logger:          logger,
	}
}

// ListVMTemplates handles listing VM templates.
func (h *VMTemplateHandler) ListVMTemplates(c *gin.Context) {
	provider := c.Query("provider")

	// Check if requesting all templates for a provider (for dropdowns)
	if c.Query("all") == constants.QueryTrue && provider != "" {
		templates, err := h.templateService.ListByProvider(c.Request.Context(), provider)
		if err != nil {
			h.logger.Error("failed to list templates by provider", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list templates"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"templates": templates})
		return
	}

	osType := c.Query("os_type")
	zoneID := c.Query("zone_id")
	page := parseInt(c.DefaultQuery("page", "1"), 1)
	pageSize := parseInt(c.DefaultQuery("page_size", "20"), constants.DefaultPageSize)
	if pageSize > constants.MaxPageSize {
		pageSize = constants.MaxPageSize
	}

	templates, total, err := h.templateService.List(c.Request.Context(), provider, osType, zoneID, page, pageSize)
	if err != nil {
		h.logger.Error("failed to list VM templates", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list VM templates"})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize
	c.JSON(http.StatusOK, gin.H{
		"templates":   templates,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// GetVMTemplate handles getting a VM template by ID.
func (h *VMTemplateHandler) GetVMTemplate(c *gin.Context) {
	id := c.Param("id")
	template, err := h.templateService.Get(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "VM template not found"})
			return
		}
		h.logger.Error("failed to get VM template", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get VM template"})
		return
	}
	c.JSON(http.StatusOK, template)
}

// CreateVMTemplateRequest represents a VM template creation request.
type CreateVMTemplateRequest struct {
	Name         string  `json:"name" binding:"required,min=1,max=64"`
	TemplateName string  `json:"template_name" binding:"required"`
	Provider     string  `json:"provider" binding:"required"`
	OSType       string  `json:"os_type" binding:"required"`
	OSFamily     string  `json:"os_family"`
	OSVersion    string  `json:"os_version"`
	ZoneID       *string `json:"zone_id"`
	MinCPU       int     `json:"min_cpu"`
	MinMemoryMB  int     `json:"min_memory_mb"`
	MinDiskGB    int     `json:"min_disk_gb"`
	DefaultUser  string  `json:"default_user"`
	CloudInit    bool    `json:"cloud_init"`
	Description  string  `json:"description"`
}

// CreateVMTemplate handles creating a VM template.
func (h *VMTemplateHandler) CreateVMTemplate(c *gin.Context) {
	var req CreateVMTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template, err := h.templateService.Create(c.Request.Context(), &service.CreateVMTemplateInput{
		Name:         req.Name,
		TemplateName: req.TemplateName,
		Provider:     req.Provider,
		OSType:       req.OSType,
		OSFamily:     req.OSFamily,
		OSVersion:    req.OSVersion,
		ZoneID:       req.ZoneID,
		MinCPU:       req.MinCPU,
		MinMemoryMB:  req.MinMemoryMB,
		MinDiskGB:    req.MinDiskGB,
		DefaultUser:  req.DefaultUser,
		CloudInit:    req.CloudInit,
		Description:  req.Description,
	})
	if err != nil {
		h.logger.Error("failed to create VM template", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, template)
}

// UpdateVMTemplateRequest represents a VM template update request.
type UpdateVMTemplateRequest struct {
	Name         *string `json:"name"`
	TemplateName *string `json:"template_name"`
	OSType       *string `json:"os_type"`
	OSFamily     *string `json:"os_family"`
	OSVersion    *string `json:"os_version"`
	MinCPU       *int    `json:"min_cpu"`
	MinMemoryMB  *int    `json:"min_memory_mb"`
	MinDiskGB    *int    `json:"min_disk_gb"`
	DefaultUser  *string `json:"default_user"`
	CloudInit    *bool   `json:"cloud_init"`
	Description  *string `json:"description"`
	Status       *int8   `json:"status"`
}

// UpdateVMTemplate handles updating a VM template.
func (h *VMTemplateHandler) UpdateVMTemplate(c *gin.Context) {
	id := c.Param("id")
	var req UpdateVMTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template, err := h.templateService.Update(c.Request.Context(), id, &service.UpdateVMTemplateInput{
		Name:         req.Name,
		TemplateName: req.TemplateName,
		OSType:       req.OSType,
		OSFamily:     req.OSFamily,
		OSVersion:    req.OSVersion,
		MinCPU:       req.MinCPU,
		MinMemoryMB:  req.MinMemoryMB,
		MinDiskGB:    req.MinDiskGB,
		DefaultUser:  req.DefaultUser,
		CloudInit:    req.CloudInit,
		Description:  req.Description,
		Status:       req.Status,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "VM template not found"})
			return
		}
		h.logger.Error("failed to update VM template", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, template)
}

// DeleteVMTemplate handles deleting a VM template.
func (h *VMTemplateHandler) DeleteVMTemplate(c *gin.Context) {
	id := c.Param("id")
	if err := h.templateService.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "VM template not found"})
			return
		}
		h.logger.Error("failed to delete VM template", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete VM template"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "VM template deleted successfully"})
}
