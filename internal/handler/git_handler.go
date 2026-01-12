// Package handler provides HTTP request handlers.
package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Veritas-Calculus/vc-lab-platform/internal/constants"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/model"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/repository"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GitHandler handles git repository management requests.
type GitHandler struct {
	gitService service.GitService
	logger     *zap.Logger
}

// NewGitHandler creates a new git handler.
func NewGitHandler(gitService service.GitService, logger *zap.Logger) *GitHandler {
	return &GitHandler{
		gitService: gitService,
		logger:     logger,
	}
}

// ListRepositories handles listing git repositories.
func (h *GitHandler) ListRepositories(c *gin.Context) {
	page := parseInt(c.DefaultQuery("page", "1"), 1)
	pageSize := parseInt(c.DefaultQuery("page_size", "20"), constants.DefaultPageSize)
	if pageSize > constants.MaxPageSize {
		pageSize = constants.MaxPageSize
	}

	repos, total, err := h.gitService.ListRepositories(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("failed to list git repositories", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list repositories"})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize
	c.JSON(http.StatusOK, gin.H{
		"repositories": repos,
		"total":        total,
		"page":         page,
		"page_size":    pageSize,
		"total_pages":  totalPages,
	})
}

// CreateGitRepoRequest represents a git repository creation request.
type CreateGitRepoRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=128"`
	Type        string `json:"type" binding:"required,oneof=modules storage"`
	URL         string `json:"url" binding:"required"`
	Branch      string `json:"branch"`
	AuthType    string `json:"auth_type"` // none, token, password, ssh_key
	Username    string `json:"username"`
	Token       string `json:"token"`
	SSHKey      string `json:"ssh_key"`
	BasePath    string `json:"base_path"`
	Description string `json:"description"`
	IsDefault   bool   `json:"is_default"`
}

// CreateRepository handles creating a git repository.
func (h *GitHandler) CreateRepository(c *gin.Context) {
	var req CreateGitRepoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	repo, err := h.gitService.CreateRepository(c.Request.Context(), &service.CreateGitRepoInput{
		Name:        req.Name,
		Type:        model.GitRepoType(req.Type),
		URL:         req.URL,
		Branch:      req.Branch,
		AuthType:    model.GitAuthType(req.AuthType),
		Username:    req.Username,
		Token:       req.Token,
		SSHKey:      req.SSHKey,
		BasePath:    req.BasePath,
		Description: req.Description,
		IsDefault:   req.IsDefault,
	})
	if err != nil {
		h.logger.Error("failed to create git repository", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, repo)
}

// GetRepository handles getting a git repository by ID.
func (h *GitHandler) GetRepository(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Repository ID required"})
		return
	}

	repo, err := h.gitService.GetRepository(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
			return
		}
		h.logger.Error("failed to get git repository", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		return
	}

	c.JSON(http.StatusOK, repo)
}

// UpdateGitRepoRequest represents a git repository update request.
type UpdateGitRepoRequest struct {
	Name        *string `json:"name"`
	URL         *string `json:"url"`
	Branch      *string `json:"branch"`
	AuthType    *string `json:"auth_type"`
	Username    *string `json:"username"`
	Token       *string `json:"token"`
	SSHKey      *string `json:"ssh_key"`
	BasePath    *string `json:"base_path"`
	Description *string `json:"description"`
	Status      *int8   `json:"status"`
	IsDefault   *bool   `json:"is_default"`
}

// UpdateRepository handles updating a git repository.
func (h *GitHandler) UpdateRepository(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Repository ID required"})
		return
	}

	var req UpdateGitRepoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert auth_type string pointer to GitAuthType pointer
	var authType *model.GitAuthType
	if req.AuthType != nil {
		at := model.GitAuthType(*req.AuthType)
		authType = &at
	}

	repo, err := h.gitService.UpdateRepository(c.Request.Context(), id, &service.UpdateGitRepoInput{
		Name:        req.Name,
		URL:         req.URL,
		Branch:      req.Branch,
		AuthType:    authType,
		Username:    req.Username,
		Token:       req.Token,
		SSHKey:      req.SSHKey,
		BasePath:    req.BasePath,
		Description: req.Description,
		Status:      req.Status,
		IsDefault:   req.IsDefault,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
			return
		}
		h.logger.Error("failed to update git repository", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update repository"})
		return
	}

	c.JSON(http.StatusOK, repo)
}

// DeleteRepository handles deleting a git repository.
func (h *GitHandler) DeleteRepository(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Repository ID required"})
		return
	}

	if err := h.gitService.DeleteRepository(c.Request.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
			return
		}
		h.logger.Error("failed to delete git repository", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete repository"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Repository deleted"})
}

// TestConnection handles testing the connection to a git repository.
func (h *GitHandler) TestConnection(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Repository ID required"})
		return
	}

	if err := h.gitService.TestConnection(c.Request.Context(), id); err != nil {
		h.logger.Error("git connection test failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Connection successful"})
}

// TestConnectionDirectRequest represents a direct connection test request.
type TestConnectionDirectRequest struct {
	URL      string `json:"url" binding:"required"`
	Branch   string `json:"branch"`
	AuthType string `json:"auth_type"` // none, token, password, ssh_key
	Username string `json:"username"`
	Token    string `json:"token"`
	SSHKey   string `json:"ssh_key"`
}

// TestConnectionDirect handles testing a git connection without saving.
func (h *GitHandler) TestConnectionDirect(c *gin.Context) {
	var req TestConnectionDirectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.gitService.TestConnectionDirect(c.Request.Context(), &service.TestConnectionInput{
		URL:      req.URL,
		Branch:   req.Branch,
		AuthType: model.GitAuthType(req.AuthType),
		Username: req.Username,
		Token:    req.Token,
		SSHKey:   req.SSHKey,
	}); err != nil {
		h.logger.Error("git connection test failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Connection successful"})
}

// ListNodeConfigs handles listing node configurations.
func (h *GitHandler) ListNodeConfigs(c *gin.Context) {
	repoID := c.Query("repo_id")
	if repoID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "repo_id is required"})
		return
	}

	page := parseInt(c.DefaultQuery("page", "1"), 1)
	pageSize := parseInt(c.DefaultQuery("page_size", "20"), constants.DefaultPageSize)
	if pageSize > constants.MaxPageSize {
		pageSize = constants.MaxPageSize
	}

	configs, total, err := h.gitService.ListNodeConfigs(c.Request.Context(), repoID, page, pageSize)
	if err != nil {
		h.logger.Error("failed to list node configs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list node configs"})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize
	c.JSON(http.StatusOK, gin.H{
		"node_configs": configs,
		"total":        total,
		"page":         page,
		"page_size":    pageSize,
		"total_pages":  totalPages,
	})
}

// GetNodeConfig handles getting a node configuration by ID.
func (h *GitHandler) GetNodeConfig(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Node config ID required"})
		return
	}

	config, err := h.gitService.GetNodeConfig(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Node config not found"})
			return
		}
		h.logger.Error("failed to get node config", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get node config"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// GetNodeConfigByRequest handles getting a node configuration by resource request ID.
func (h *GitHandler) GetNodeConfigByRequest(c *gin.Context) {
	requestID := c.Param("request_id")
	if requestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request ID required"})
		return
	}

	config, err := h.gitService.GetNodeConfigByRequest(c.Request.Context(), requestID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Node config not found"})
			return
		}
		h.logger.Error("failed to get node config", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get node config"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// CommitNodeConfigRequest represents a request to commit a node configuration.
type CommitNodeConfigRequest struct {
	Message string `json:"message" binding:"required"`
}

// CommitNodeConfig handles committing a node configuration to the storage repository.
func (h *GitHandler) CommitNodeConfig(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Node config ID required"})
		return
	}

	var req CommitNodeConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	commitSHA, err := h.gitService.CommitNodeConfig(c.Request.Context(), id, req.Message)
	if err != nil {
		h.logger.Error("failed to commit node config", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Node config committed successfully",
		"commit_sha": commitSHA,
	})
}

// ListModulesFromGit handles listing Terraform modules from the default modules git repository.
func (h *GitHandler) ListModulesFromGit(c *gin.Context) {
	modules, err := h.gitService.ListModulesFromGit(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to list modules from git", zap.Error(err))
		// Check if the error is about missing repository configuration
		if strings.Contains(err.Error(), "no default modules repository configured") {
			c.JSON(http.StatusOK, gin.H{
				"modules": []interface{}{},
				"total":   0,
				"warning": "No default modules repository configured. Please configure a modules repository in Git Ops settings.",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"modules": modules,
		"total":   len(modules),
	})
}

// SyncModulesFromGit handles syncing (refreshing) Terraform modules from the git repository.
func (h *GitHandler) SyncModulesFromGit(c *gin.Context) {
	modules, err := h.gitService.SyncModulesFromGit(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to sync modules from git", zap.Error(err))
		// Check if the error is about missing repository configuration
		if strings.Contains(err.Error(), "no default modules repository configured") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "No default modules repository configured. Please configure a modules repository in Git Ops settings first.",
				"code":    "NO_MODULES_REPO",
				"modules": []interface{}{},
				"total":   0,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"modules": modules,
		"total":   len(modules),
		"message": "Modules synced successfully",
	})
}
