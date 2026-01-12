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

// SSHKeyHandler handles SSH key related HTTP requests.
type SSHKeyHandler struct {
	sshKeyService service.SSHKeyService
	logger        *zap.Logger
}

// NewSSHKeyHandler creates a new SSH key handler.
func NewSSHKeyHandler(sshKeyService service.SSHKeyService, logger *zap.Logger) *SSHKeyHandler {
	return &SSHKeyHandler{
		sshKeyService: sshKeyService,
		logger:        logger,
	}
}

// ListSSHKeys handles listing SSH keys.
func (h *SSHKeyHandler) ListSSHKeys(c *gin.Context) {
	// Check if requesting all for dropdowns
	if c.Query("all") == constants.QueryTrue {
		sshKeys, _, err := h.sshKeyService.List(c.Request.Context(), 1, constants.MaxPageSize)
		if err != nil {
			h.logger.Error("failed to list SSH keys", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list SSH keys"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ssh_keys": sshKeys})
		return
	}

	page := parseInt(c.DefaultQuery("page", "1"), 1)
	pageSize := parseInt(c.DefaultQuery("page_size", "20"), constants.DefaultPageSize)
	if pageSize > constants.MaxPageSize {
		pageSize = constants.MaxPageSize
	}

	sshKeys, total, err := h.sshKeyService.List(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("failed to list SSH keys", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list SSH keys"})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize
	c.JSON(http.StatusOK, gin.H{
		"ssh_keys":    sshKeys,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// GetSSHKey handles getting an SSH key by ID.
func (h *SSHKeyHandler) GetSSHKey(c *gin.Context) {
	id := c.Param("id")
	sshKey, err := h.sshKeyService.Get(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "SSH key not found"})
			return
		}
		h.logger.Error("failed to get SSH key", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get SSH key"})
		return
	}
	c.JSON(http.StatusOK, sshKey)
}

// GetDefaultSSHKey handles getting the default SSH key.
func (h *SSHKeyHandler) GetDefaultSSHKey(c *gin.Context) {
	sshKey, err := h.sshKeyService.GetDefault(c.Request.Context())
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "No default SSH key found"})
			return
		}
		h.logger.Error("failed to get default SSH key", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get default SSH key"})
		return
	}
	c.JSON(http.StatusOK, sshKey)
}

// CreateSSHKeyRequest represents an SSH key creation request.
type CreateSSHKeyRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=128"`
	PublicKey   string `json:"public_key" binding:"required"`
	Description string `json:"description"`
	IsDefault   bool   `json:"is_default"`
}

// CreateSSHKey handles creating an SSH key.
func (h *SSHKeyHandler) CreateSSHKey(c *gin.Context) {
	var req CreateSSHKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return
	}

	sshKey, err := h.sshKeyService.Create(c.Request.Context(), &service.CreateSSHKeyInput{
		Name:        req.Name,
		PublicKey:   req.PublicKey,
		Description: req.Description,
		CreatedByID: userIDStr,
		IsDefault:   req.IsDefault,
	})
	if err != nil {
		h.logger.Error("failed to create SSH key", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, sshKey)
}

// UpdateSSHKeyRequest represents an SSH key update request.
type UpdateSSHKeyRequest struct {
	Name        *string `json:"name"`
	PublicKey   *string `json:"public_key"`
	Description *string `json:"description"`
	IsDefault   *bool   `json:"is_default"`
}

// UpdateSSHKey handles updating an SSH key.
func (h *SSHKeyHandler) UpdateSSHKey(c *gin.Context) {
	id := c.Param("id")
	var req UpdateSSHKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sshKey, err := h.sshKeyService.Update(c.Request.Context(), id, &service.UpdateSSHKeyInput{
		Name:        req.Name,
		PublicKey:   req.PublicKey,
		Description: req.Description,
		IsDefault:   req.IsDefault,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "SSH key not found"})
			return
		}
		h.logger.Error("failed to update SSH key", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sshKey)
}

// DeleteSSHKey handles deleting an SSH key.
func (h *SSHKeyHandler) DeleteSSHKey(c *gin.Context) {
	id := c.Param("id")
	if err := h.sshKeyService.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "SSH key not found"})
			return
		}
		h.logger.Error("failed to delete SSH key", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete SSH key"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "SSH key deleted successfully"})
}

// SetDefaultSSHKey handles setting an SSH key as default.
func (h *SSHKeyHandler) SetDefaultSSHKey(c *gin.Context) {
	id := c.Param("id")
	if err := h.sshKeyService.SetDefault(c.Request.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "SSH key not found"})
			return
		}
		h.logger.Error("failed to set default SSH key", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set default SSH key"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "SSH key set as default"})
}
