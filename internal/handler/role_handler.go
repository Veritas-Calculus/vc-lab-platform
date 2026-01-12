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

// RoleHandler handles role management requests.
type RoleHandler struct {
	roleService service.RoleService
	logger      *zap.Logger
}

// NewRoleHandler creates a new role handler.
func NewRoleHandler(roleService service.RoleService, logger *zap.Logger) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
		logger:      logger,
	}
}

// List handles listing roles.
func (h *RoleHandler) List(c *gin.Context) {
	page := parseInt(c.DefaultQuery("page", "1"), 1)
	pageSize := parseInt(c.DefaultQuery("page_size", "20"), constants.DefaultPageSize)

	if pageSize > constants.MaxPageSize {
		pageSize = constants.MaxPageSize
	}

	roles, total, err := h.roleService.List(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("failed to list roles", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list roles"})
		return
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	c.JSON(http.StatusOK, gin.H{
		"roles":       roles,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// CreateRoleRequest represents a role creation request.
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=64"`
	Code        string `json:"code" binding:"required,min=2,max=64"`
	Description string `json:"description"`
}

// Create handles role creation.
func (h *RoleHandler) Create(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role, err := h.roleService.Create(c.Request.Context(), &service.CreateRoleInput{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
	})
	if err != nil {
		h.logger.Error("failed to create role", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, role)
}

// GetByID handles getting a role by ID.
func (h *RoleHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role ID required"})
		return
	}

	role, err := h.roleService.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
			return
		}
		h.logger.Error("failed to get role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get role"})
		return
	}

	c.JSON(http.StatusOK, role)
}

// UpdateRoleRequest represents a role update request.
type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      *int8  `json:"status"`
}

// Update handles role updates.
func (h *RoleHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role ID required"})
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	role, err := h.roleService.Update(c.Request.Context(), id, updates)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
			return
		}
		h.logger.Error("failed to update role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role"})
		return
	}

	c.JSON(http.StatusOK, role)
}

// Delete handles role deletion.
func (h *RoleHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role ID required"})
		return
	}

	if err := h.roleService.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
			return
		}
		h.logger.Error("failed to delete role", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
