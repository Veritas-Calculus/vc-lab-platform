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

// SettingsHandler handles settings-related HTTP requests.
type SettingsHandler struct {
	settingsService service.SettingsService
	logger          *zap.Logger
}

// NewSettingsHandler creates a new settings handler.
func NewSettingsHandler(settingsService service.SettingsService, logger *zap.Logger) *SettingsHandler {
	return &SettingsHandler{
		settingsService: settingsService,
		logger:          logger,
	}
}

// CreateProviderRequest represents the request body for creating a provider.
type CreateProviderRequest struct {
	Name         string `json:"name" binding:"required,min=1,max=128"`
	Type         string `json:"type" binding:"required,oneof=pve vmware openstack aws aliyun gcp azure"`
	Endpoint     string `json:"endpoint" binding:"required,url"`
	Description  string `json:"description"`
	Config       string `json:"config"`
	IsDefault    bool   `json:"is_default"`
	CredentialID string `json:"credential_id"`
}

// UpdateProviderRequest represents the request body for updating a provider.
type UpdateProviderRequest struct {
	Name        *string `json:"name"`
	Endpoint    *string `json:"endpoint"`
	Description *string `json:"description"`
	Config      *string `json:"config"`
	Status      *int8   `json:"status"`
	IsDefault   *bool   `json:"is_default"`
}

// ListProviders lists all providers.
func (h *SettingsHandler) ListProviders(c *gin.Context) {
	page := parseInt(c.DefaultQuery("page", "1"), 1)
	pageSize := parseInt(c.DefaultQuery("page_size", "20"), constants.DefaultPageSize)
	if pageSize > constants.MaxPageSize {
		pageSize = constants.MaxPageSize
	}
	providerType := c.Query("type")

	providers, total, err := h.settingsService.ListProviders(c.Request.Context(), providerType, page, pageSize)
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

// CreateProvider creates a new provider.
func (h *SettingsHandler) CreateProvider(c *gin.Context) {
	var req CreateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	provider, err := h.settingsService.CreateProvider(c.Request.Context(), &service.CreateProviderInput{
		Name:         req.Name,
		Type:         req.Type,
		Endpoint:     req.Endpoint,
		Description:  req.Description,
		Config:       req.Config,
		IsDefault:    req.IsDefault,
		CredentialID: req.CredentialID,
	})
	if err != nil {
		h.logger.Error("failed to create provider", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create provider"})
		return
	}

	c.JSON(http.StatusCreated, provider)
}

// GetProvider gets a provider by ID.
func (h *SettingsHandler) GetProvider(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provider ID required"})
		return
	}

	provider, err := h.settingsService.GetProvider(c.Request.Context(), id)
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

// UpdateProvider updates a provider.
func (h *SettingsHandler) UpdateProvider(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provider ID required"})
		return
	}

	var req UpdateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	provider, err := h.settingsService.UpdateProvider(c.Request.Context(), id, &service.UpdateProviderInput{
		Name:        req.Name,
		Endpoint:    req.Endpoint,
		Description: req.Description,
		Config:      req.Config,
		Status:      req.Status,
		IsDefault:   req.IsDefault,
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

// DeleteProvider deletes a provider.
func (h *SettingsHandler) DeleteProvider(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provider ID required"})
		return
	}

	if err := h.settingsService.DeleteProvider(c.Request.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Provider not found"})
			return
		}
		h.logger.Error("failed to delete provider", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete provider"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Provider deleted"})
}

// TestProviderConnectionRequest represents the request body for testing provider connection.
type TestProviderConnectionRequest struct {
	Type         string `json:"type" binding:"required,oneof=pve vmware openstack aws aliyun gcp azure"`
	Endpoint     string `json:"endpoint" binding:"required,url"`
	CredentialID string `json:"credential_id"`
	Config       string `json:"config"`
}

// TestProviderConnection tests a provider connection.
func (h *SettingsHandler) TestProviderConnection(c *gin.Context) {
	var req TestProviderConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.settingsService.TestProviderConnection(c.Request.Context(), &service.TestProviderConnectionInput{
		Type:         req.Type,
		Endpoint:     req.Endpoint,
		CredentialID: req.CredentialID,
		Config:       req.Config,
	}); err != nil {
		h.logger.Error("failed to test provider connection", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Connection successful"})
}

// TestCredentialConnectionRequest represents the request body for testing credential connection.
type TestCredentialConnectionRequest struct {
	Type      string `json:"type" binding:"required,oneof=pve vmware openstack aws aliyun gcp azure"`
	Endpoint  string `json:"endpoint" binding:"required,url"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Token     string `json:"token"`
}

// TestCredentialConnection tests a credential connection.
func (h *SettingsHandler) TestCredentialConnection(c *gin.Context) {
	var req TestCredentialConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.settingsService.TestCredentialConnection(c.Request.Context(), &service.TestCredentialConnectionInput{
		Type:      req.Type,
		Endpoint:  req.Endpoint,
		AccessKey: req.AccessKey,
		SecretKey: req.SecretKey,
		Token:     req.Token,
	}); err != nil {
		h.logger.Error("failed to test credential connection", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Authentication successful"})
}

// CreateCredentialRequest represents the request body for creating a credential.
type CreateCredentialRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=128"`
	Type        string  `json:"type" binding:"required,oneof=pve vmware openstack aws aliyun gcp azure"`
	ZoneID      *string `json:"zone_id"`
	Endpoint    string  `json:"endpoint"`
	ProviderID  string  `json:"provider_id"`
	AccessKey   string  `json:"access_key"`
	SecretKey   string  `json:"secret_key"`
	Token       string  `json:"token"`
	Description string  `json:"description"`
}

// UpdateCredentialRequest represents the request body for updating a credential.
type UpdateCredentialRequest struct {
	Name        *string `json:"name"`
	ZoneID      *string `json:"zone_id"`
	Endpoint    *string `json:"endpoint"`
	AccessKey   *string `json:"access_key"`
	SecretKey   *string `json:"secret_key"`
	Token       *string `json:"token"`
	Description *string `json:"description"`
	Status      *int8   `json:"status"`
}

// ListCredentials lists all credentials.
func (h *SettingsHandler) ListCredentials(c *gin.Context) {
	page := parseInt(c.DefaultQuery("page", "1"), 1)
	pageSize := parseInt(c.DefaultQuery("page_size", "20"), constants.DefaultPageSize)
	if pageSize > constants.MaxPageSize {
		pageSize = constants.MaxPageSize
	}
	credentialType := c.Query("type")

	credentials, total, err := h.settingsService.ListCredentials(c.Request.Context(), credentialType, page, pageSize)
	if err != nil {
		h.logger.Error("failed to list credentials", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list credentials"})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize
	c.JSON(http.StatusOK, gin.H{
		"credentials": credentials,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// CreateCredential creates a new credential.
func (h *SettingsHandler) CreateCredential(c *gin.Context) {
	var req CreateCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	credential, err := h.settingsService.CreateCredential(c.Request.Context(), &service.CreateCredentialInput{
		Name:        req.Name,
		Type:        req.Type,
		ZoneID:      req.ZoneID,
		Endpoint:    req.Endpoint,
		ProviderID:  req.ProviderID,
		AccessKey:   req.AccessKey,
		SecretKey:   req.SecretKey,
		Token:       req.Token,
		Description: req.Description,
		CreatedByID: userID,
	})
	if err != nil {
		h.logger.Error("failed to create credential", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create credential"})
		return
	}

	c.JSON(http.StatusCreated, credential)
}

// GetCredential gets a credential by ID.
func (h *SettingsHandler) GetCredential(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Credential ID required"})
		return
	}

	credential, err := h.settingsService.GetCredential(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Credential not found"})
			return
		}
		h.logger.Error("failed to get credential", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get credential"})
		return
	}

	c.JSON(http.StatusOK, credential)
}

// UpdateCredential updates a credential.
func (h *SettingsHandler) UpdateCredential(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Credential ID required"})
		return
	}

	var req UpdateCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	credential, err := h.settingsService.UpdateCredential(c.Request.Context(), id, &service.UpdateCredentialInput{
		Name:        req.Name,
		ZoneID:      req.ZoneID,
		Endpoint:    req.Endpoint,
		AccessKey:   req.AccessKey,
		SecretKey:   req.SecretKey,
		Token:       req.Token,
		Description: req.Description,
		Status:      req.Status,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Credential not found"})
			return
		}
		h.logger.Error("failed to update credential", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update credential"})
		return
	}

	c.JSON(http.StatusOK, credential)
}

// DeleteCredential deletes a credential.
func (h *SettingsHandler) DeleteCredential(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Credential ID required"})
		return
	}

	if err := h.settingsService.DeleteCredential(c.Request.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Credential not found"})
			return
		}
		h.logger.Error("failed to delete credential", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete credential"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Credential deleted"})
}
