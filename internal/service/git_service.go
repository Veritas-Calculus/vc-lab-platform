// Package service provides business logic implementations.
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/Veritas-Calculus/vc-lab-platform/internal/model"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/repository"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/sanitize"
	"go.uber.org/zap"
)

// File permission constants for git operations.
const (
	dirPerm  = 0o750 // Directory permissions (rwxr-x---)
	filePerm = 0o644 // File permissions (rw-r--r--)
)

// GitService defines the interface for git operations.
type GitService interface {
	// Repository management
	ListRepositories(ctx context.Context, page, pageSize int) ([]model.GitRepository, int64, error)
	GetRepository(ctx context.Context, id string) (*model.GitRepository, error)
	GetDefaultRepository(ctx context.Context, repoType model.GitRepoType) (*model.GitRepository, error)
	CreateRepository(ctx context.Context, input *CreateGitRepoInput) (*model.GitRepository, error)
	UpdateRepository(ctx context.Context, id string, input *UpdateGitRepoInput) (*model.GitRepository, error)
	DeleteRepository(ctx context.Context, id string) error
	TestConnection(ctx context.Context, id string) error
	TestConnectionDirect(ctx context.Context, input *TestConnectionInput) error

	// Node config management
	CreateNodeConfig(ctx context.Context, request *model.ResourceRequest) (*model.NodeConfig, error)
	UpdateNodeConfigStatus(ctx context.Context, configID string, status model.NodeConfigStatus, log string) error
	CommitNodeConfig(ctx context.Context, configID string, message string) (string, error)
	GetNodeConfig(ctx context.Context, id string) (*model.NodeConfig, error)
	GetNodeConfigByRequest(ctx context.Context, requestID string) (*model.NodeConfig, error)
	ListNodeConfigs(ctx context.Context, repoID string, page, pageSize int) ([]model.NodeConfig, int64, error)

	// Git operations
	CloneRepository(ctx context.Context, repo *model.GitRepository, targetPath string) error
	PullChanges(ctx context.Context, repoPath string) error
	CommitAndPush(ctx context.Context, repoPath string, files []string, message string) (string, error)

	// Module operations
	ListModulesFromGit(ctx context.Context) ([]GitModule, error)
	SyncModulesFromGit(ctx context.Context) ([]GitModule, error)
}

// GitModule represents a Terraform module discovered from a git repository.
type GitModule struct {
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	Description string   `json:"description,omitempty"`
	Source      string   `json:"source"`
	Variables   []string `json:"variables,omitempty"`
	Outputs     []string `json:"outputs,omitempty"`
}

// CreateGitRepoInput represents input for creating a git repository.
type CreateGitRepoInput struct {
	Name        string
	Type        model.GitRepoType
	URL         string
	Branch      string
	AuthType    model.GitAuthType
	Username    string
	Token       string
	SSHKey      string
	BasePath    string
	Description string
	IsDefault   bool
}

// UpdateGitRepoInput represents input for updating a git repository.
type UpdateGitRepoInput struct {
	Name        *string
	URL         *string
	Branch      *string
	AuthType    *model.GitAuthType
	Username    *string
	Token       *string
	SSHKey      *string
	BasePath    *string
	Description *string
	Status      *int8
	IsDefault   *bool
}

// TestConnectionInput represents input for testing a git connection.
type TestConnectionInput struct {
	URL      string
	Branch   string
	AuthType model.GitAuthType
	Username string
	Token    string
	SSHKey   string
}

type gitService struct {
	gitRepoRepo    repository.GitRepoRepository
	nodeConfigRepo repository.NodeConfigRepository
	tfModuleRepo   repository.TerraformModuleRepository
	logger         *zap.Logger
	workDir        string // Base directory for git operations
}

// NewGitService creates a new git service.
func NewGitService(
	gitRepoRepo repository.GitRepoRepository,
	nodeConfigRepo repository.NodeConfigRepository,
	tfModuleRepo repository.TerraformModuleRepository,
	logger *zap.Logger,
) GitService {
	workDir := os.Getenv("GIT_WORK_DIR")
	if workDir == "" {
		workDir = "/tmp/git-repos"
	}
	return &gitService{
		gitRepoRepo:    gitRepoRepo,
		nodeConfigRepo: nodeConfigRepo,
		tfModuleRepo:   tfModuleRepo,
		logger:         logger,
		workDir:        workDir,
	}
}

// ListRepositories lists all git repositories with pagination.
func (s *gitService) ListRepositories(ctx context.Context, page, pageSize int) ([]model.GitRepository, int64, error) {
	return s.gitRepoRepo.List(ctx, page, pageSize)
}

// GetRepository retrieves a git repository by ID.
func (s *gitService) GetRepository(ctx context.Context, id string) (*model.GitRepository, error) {
	return s.gitRepoRepo.GetByID(ctx, id)
}

// GetDefaultRepository retrieves the default repository for a given type.
func (s *gitService) GetDefaultRepository(ctx context.Context, repoType model.GitRepoType) (*model.GitRepository, error) {
	return s.gitRepoRepo.GetDefaultByType(ctx, repoType)
}

// CreateRepository creates a new git repository.
func (s *gitService) CreateRepository(ctx context.Context, input *CreateGitRepoInput) (*model.GitRepository, error) {
	if input == nil {
		return nil, errors.New("input cannot be nil")
	}
	if input.Name == "" {
		return nil, errors.New("name is required")
	}
	if input.URL == "" {
		return nil, errors.New("url is required")
	}

	branch := input.Branch
	if branch == "" {
		branch = "main"
	}

	basePath := input.BasePath
	if basePath == "" {
		basePath = "/"
	}

	authType := input.AuthType
	if authType == "" {
		authType = model.GitAuthTypeNone
	}

	repo := &model.GitRepository{
		Name:        input.Name,
		Type:        input.Type,
		URL:         input.URL,
		Branch:      branch,
		AuthType:    authType,
		Username:    input.Username,
		Token:       input.Token,
		SSHKey:      input.SSHKey,
		BasePath:    basePath,
		Description: input.Description,
		IsDefault:   input.IsDefault,
		Status:      1,
	}

	if err := s.gitRepoRepo.Create(ctx, repo); err != nil {
		s.logger.Error("failed to create git repository", zap.Error(err))
		return nil, errors.New("failed to create git repository")
	}

	return repo, nil
}

// UpdateRepository updates a git repository.
func (s *gitService) UpdateRepository(ctx context.Context, id string, input *UpdateGitRepoInput) (*model.GitRepository, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	repo, err := s.gitRepoRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		repo.Name = *input.Name
	}
	if input.URL != nil {
		repo.URL = *input.URL
	}
	if input.Branch != nil {
		repo.Branch = *input.Branch
	}
	if input.AuthType != nil {
		repo.AuthType = *input.AuthType
	}
	if input.Username != nil {
		repo.Username = *input.Username
	}
	if input.Token != nil {
		repo.Token = *input.Token
	}
	if input.SSHKey != nil {
		repo.SSHKey = *input.SSHKey
	}
	if input.BasePath != nil {
		repo.BasePath = *input.BasePath
	}
	if input.Description != nil {
		repo.Description = *input.Description
	}
	if input.Status != nil {
		repo.Status = *input.Status
	}
	if input.IsDefault != nil {
		repo.IsDefault = *input.IsDefault
	}

	if err := s.gitRepoRepo.Update(ctx, repo); err != nil {
		s.logger.Error("failed to update git repository", zap.Error(err))
		return nil, errors.New("failed to update git repository")
	}

	return repo, nil
}

// DeleteRepository deletes a git repository.
func (s *gitService) DeleteRepository(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}

	if _, err := s.gitRepoRepo.GetByID(ctx, id); err != nil {
		return err
	}

	return s.gitRepoRepo.Delete(ctx, id)
}

// TestConnection tests the connection to a git repository.
func (s *gitService) TestConnection(ctx context.Context, id string) error {
	repo, err := s.gitRepoRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Validate URL and branch from stored repository
	if _, urlErr := sanitize.ValidateGitURL(repo.URL); urlErr != nil {
		return fmt.Errorf("invalid repository URL: %w", urlErr)
	}
	branch, branchErr := sanitize.ValidateGitBranch(repo.Branch)
	if branchErr != nil {
		return fmt.Errorf("invalid branch name: %w", branchErr)
	}

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir) //nolint:errcheck // best effort cleanup

	// Try to clone with depth 1 to test connection
	cloneURL := s.buildAuthenticatedURL(repo)
	args := []string{"clone", "--depth", "1", "--branch", branch, cloneURL, tempDir}

	cmd := exec.CommandContext(ctx, "git", args...) // #nosec G204 --  URL and branch validated above
	output, err := cmd.CombinedOutput()
	if err != nil {
		s.logger.Error("git clone test failed",
			zap.String("repo", sanitize.ForLog(repo.Name)),
			zap.String("output", sanitize.CommandOutput(string(output))),
			zap.Error(err),
		)
		return fmt.Errorf("failed to connect to repository: %s", sanitize.CommandOutput(string(output)))
	}

	// Update last sync time
	now := time.Now()
	repo.LastSyncAt = &now
	if err := s.gitRepoRepo.Update(ctx, repo); err != nil {
		s.logger.Warn("failed to update last sync time", zap.Error(err))
	}

	return nil
}

// TestConnectionDirect tests a git connection without saving the repository.
func (s *gitService) TestConnectionDirect(ctx context.Context, input *TestConnectionInput) error {
	if input == nil {
		return errors.New("input cannot be nil")
	}
	if input.URL == "" {
		return errors.New("url is required")
	}

	// Validate the URL to prevent command injection
	validatedURL, err := sanitize.ValidateGitURL(input.URL)
	if err != nil {
		return fmt.Errorf("invalid repository URL: %w", err)
	}

	// Validate and sanitize the branch name
	branch, err := sanitize.ValidateGitBranch(input.Branch)
	if err != nil {
		return fmt.Errorf("invalid branch name: %w", err)
	}

	// Create a temporary repository object for testing
	repo := &model.GitRepository{
		URL:      validatedURL,
		Branch:   branch,
		AuthType: input.AuthType,
		Username: input.Username,
		Token:    input.Token,
		SSHKey:   input.SSHKey,
	}

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "git-test-direct-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir) //nolint:errcheck // best effort cleanup

	// Try to clone with depth 1 to test connection
	// Note: cloneURL is built from validated URL with credentials added
	cloneURL := s.buildAuthenticatedURL(repo)
	args := []string{"clone", "--depth", "1", "--branch", branch, cloneURL, tempDir}

	cmd := exec.CommandContext(ctx, "git", args...) // #nosec G204 --  URL and branch validated above
	output, err := cmd.CombinedOutput()
	if err != nil {
		s.logger.Error("git clone test failed",
			zap.String("url", sanitize.URL(input.URL)),
			zap.String("output", sanitize.CommandOutput(string(output))),
			zap.Error(err),
		)
		return fmt.Errorf("failed to connect to repository: %s", sanitize.CommandOutput(string(output)))
	}

	return nil
}

// CreateNodeConfig creates a new node configuration for a resource request.
func (s *gitService) CreateNodeConfig(ctx context.Context, request *model.ResourceRequest) (*model.NodeConfig, error) {
	if request == nil {
		return nil, errors.New("request cannot be nil")
	}

	// Get the default storage repository
	storageRepo, err := s.gitRepoRepo.GetDefaultByType(ctx, model.GitRepoTypeStorage)
	if err != nil {
		return nil, fmt.Errorf("no default storage repository configured: %w", err)
	}

	// Get the default modules repository (optional)
	var moduleRepoID *string
	moduleRepo, err := s.gitRepoRepo.GetDefaultByType(ctx, model.GitRepoTypeModules)
	if err == nil {
		moduleRepoID = &moduleRepo.ID
	}

	// Generate the config path
	configPath := s.generateConfigPath(request, storageRepo)

	// Generate terragrunt config
	terragruntConfig, err := s.generateTerragruntConfig(request, moduleRepo)
	if err != nil {
		return nil, fmt.Errorf("failed to generate terragrunt config: %w", err)
	}

	// Create the node config record
	config := &model.NodeConfig{
		Name:              s.generateNodeName(request),
		Path:              configPath,
		ResourceRequestID: request.ID,
		StorageRepoID:     storageRepo.ID,
		ModuleRepoID:      moduleRepoID,
		TerragruntConfig:  terragruntConfig,
		TerraformVars:     request.Spec,
		Status:            model.NodeConfigStatusPending,
	}

	if createErr := s.nodeConfigRepo.Create(ctx, config); createErr != nil {
		s.logger.Error("failed to create node config", zap.Error(createErr))
		return nil, errors.New("failed to create node config")
	}

	// Commit the pending config to storage repo
	commitSHA, commitErr := s.commitPendingConfig(ctx, config, storageRepo)
	if commitErr != nil {
		s.logger.Warn("failed to commit pending config", zap.Error(commitErr))
		// Don't fail the creation, just log
	} else {
		config.PendingCommitSHA = commitSHA
		if updateErr := s.nodeConfigRepo.Update(ctx, config); updateErr != nil {
			s.logger.Warn("failed to update commit SHA", zap.Error(updateErr))
		}
	}

	return config, nil
}

// UpdateNodeConfigStatus updates the status of a node configuration.
func (s *gitService) UpdateNodeConfigStatus(ctx context.Context, configID string, status model.NodeConfigStatus, log string) error {
	config, err := s.nodeConfigRepo.GetByID(ctx, configID)
	if err != nil {
		return err
	}

	config.Status = status
	if log != "" {
		config.ProvisionLog = log
	}

	//nolint:staticcheck // if-else chain is clearer here
	if status == model.NodeConfigStatusActive {
		now := time.Now()
		config.ProvisionedAt = &now
	} else if status == model.NodeConfigStatusDestroyed {
		now := time.Now()
		config.DestroyedAt = &now
	}

	return s.nodeConfigRepo.Update(ctx, config)
}

// CommitNodeConfig commits the node configuration to the storage repository.
//
//nolint:gocritic // paramTypeCombine suggestion is less readable
func (s *gitService) CommitNodeConfig(ctx context.Context, configID string, message string) (string, error) {
	config, err := s.nodeConfigRepo.GetByID(ctx, configID)
	if err != nil {
		return "", err
	}

	storageRepo, err := s.gitRepoRepo.GetByID(ctx, config.StorageRepoID)
	if err != nil {
		return "", err
	}

	// Clone the repo
	repoPath := filepath.Join(s.workDir, storageRepo.ID)
	if cloneErr := s.CloneRepository(ctx, storageRepo, repoPath); cloneErr != nil {
		return "", fmt.Errorf("failed to clone repository: %w", cloneErr)
	}
	defer os.RemoveAll(repoPath) //nolint:errcheck // best effort cleanup

	// Write the config file
	configFilePath := filepath.Join(repoPath, storageRepo.BasePath, config.Path, "terragrunt.hcl")
	if mkdirErr := os.MkdirAll(filepath.Dir(configFilePath), dirPerm); mkdirErr != nil {
		return "", fmt.Errorf("failed to create directory: %w", mkdirErr)
	}

	if writeErr := os.WriteFile(configFilePath, []byte(config.TerragruntConfig), filePerm); writeErr != nil {
		return "", fmt.Errorf("failed to write config file: %w", writeErr)
	}

	// Commit and push
	commitSHA, err := s.CommitAndPush(ctx, repoPath, []string{configFilePath}, message)
	if err != nil {
		return "", err
	}

	// Update the config with the commit SHA
	config.CommitSHA = commitSHA
	if err := s.nodeConfigRepo.Update(ctx, config); err != nil {
		s.logger.Warn("failed to update commit SHA", zap.Error(err))
	}

	return commitSHA, nil
}

// GetNodeConfig retrieves a node configuration by ID.
func (s *gitService) GetNodeConfig(ctx context.Context, id string) (*model.NodeConfig, error) {
	return s.nodeConfigRepo.GetByID(ctx, id)
}

// GetNodeConfigByRequest retrieves a node configuration by resource request ID.
func (s *gitService) GetNodeConfigByRequest(ctx context.Context, requestID string) (*model.NodeConfig, error) {
	return s.nodeConfigRepo.GetByResourceRequestID(ctx, requestID)
}

// ListNodeConfigs lists node configurations for a storage repository.
func (s *gitService) ListNodeConfigs(ctx context.Context, repoID string, page, pageSize int) ([]model.NodeConfig, int64, error) {
	return s.nodeConfigRepo.ListByStorageRepo(ctx, repoID, page, pageSize)
}

// CloneRepository clones a git repository to the target path.
func (s *gitService) CloneRepository(ctx context.Context, repo *model.GitRepository, targetPath string) error {
	// Validate URL and branch
	if _, urlErr := sanitize.ValidateGitURL(repo.URL); urlErr != nil {
		return fmt.Errorf("invalid repository URL: %w", urlErr)
	}
	branch, branchErr := sanitize.ValidateGitBranch(repo.Branch)
	if branchErr != nil {
		return fmt.Errorf("invalid branch name: %w", branchErr)
	}

	// Remove existing directory if exists
	if rmErr := os.RemoveAll(targetPath); rmErr != nil {
		return fmt.Errorf("failed to remove existing directory: %w", rmErr)
	}

	// Build the clone command
	cloneURL := s.buildAuthenticatedURL(repo)
	args := []string{"clone", "--branch", branch, "--single-branch", cloneURL, targetPath}

	cmd := exec.CommandContext(ctx, "git", args...) // #nosec G204 --  URL and branch validated above
	output, err := cmd.CombinedOutput()
	if err != nil {
		s.logger.Error("git clone failed",
			zap.String("repo", sanitize.ForLog(repo.Name)),
			zap.String("output", sanitize.CommandOutput(string(output))),
			zap.Error(err),
		)
		return fmt.Errorf("failed to clone repository: %s", sanitize.CommandOutput(string(output)))
	}

	return nil
}

// PullChanges pulls the latest changes from the remote repository.
func (s *gitService) PullChanges(ctx context.Context, repoPath string) error {
	cmd := exec.CommandContext(ctx, "git", "pull")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		s.logger.Error("git pull failed",
			zap.String("path", sanitize.Path(repoPath)),
			zap.String("output", sanitize.CommandOutput(string(output))),
			zap.Error(err),
		)
		return fmt.Errorf("failed to pull changes: %s", sanitize.CommandOutput(string(output)))
	}
	return nil
}

// CommitAndPush commits changes and pushes to the remote repository.
func (s *gitService) CommitAndPush(ctx context.Context, repoPath string, files []string, message string) (string, error) {
	// Add files
	for _, file := range files {
		relPath, err := filepath.Rel(repoPath, file)
		if err != nil {
			relPath = file
		}
		cmd := exec.CommandContext(ctx, "git", "add", relPath) // #nosec G204 --  args are controlled internally
		cmd.Dir = repoPath
		if output, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("failed to add file %s: %s", relPath, string(output))
		}
	}

	// Commit
	cmd := exec.CommandContext(ctx, "git", "commit", "-m", message)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to commit: %s", string(output))
	}

	// Get the commit SHA
	cmd = exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit SHA: %w", err)
	}
	commitSHA := strings.TrimSpace(string(output))

	// Push
	cmd = exec.CommandContext(ctx, "git", "push")
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to push: %s", string(output))
	}

	return commitSHA, nil
}

// Helper functions

//nolint:nestif // auth type handling requires nested checks
func (s *gitService) buildAuthenticatedURL(repo *model.GitRepository) string {
	// If using SSH key auth type, return the URL as-is (assuming SSH URL format)
	if repo.AuthType == model.GitAuthTypeSSHKey && repo.SSHKey != "" {
		return repo.URL
	}

	// If no auth or public repo, return URL as-is
	if repo.AuthType == model.GitAuthTypeNone || repo.AuthType == "" {
		return repo.URL
	}

	// If using token or password auth, embed credentials in HTTPS URL
	if (repo.AuthType == model.GitAuthTypeToken || repo.AuthType == model.GitAuthTypePassword) && repo.Token != "" {
		// Parse the URL and add credentials
		if strings.HasPrefix(repo.URL, "https://") {
			parts := strings.SplitN(repo.URL, "://", 2)
			if len(parts) == 2 {
				username := repo.Username
				if username == "" {
					// For token auth, use a default username (works for GitHub, GitLab, etc.)
					username = "git"
				}
				return fmt.Sprintf("https://%s:%s@%s", username, repo.Token, parts[1])
			}
		}
	}

	return repo.URL
}

func (s *gitService) generateConfigPath(request *model.ResourceRequest, _ *model.GitRepository) string {
	// Generate path like: proxmox-ve/instance/{type}/{name}
	provider := request.Provider
	if provider == "" {
		provider = "default"
	}

	resourceType := request.Type
	if resourceType == "" {
		resourceType = "vm"
	}

	nodeName := s.generateNodeName(request)

	return filepath.Join(provider, "instance", resourceType, nodeName)
}

func (s *gitService) generateNodeName(request *model.ResourceRequest) string {
	// Generate a unique node name from the request
	// Format: {title-slug}-{short-id}
	title := strings.ToLower(request.Title)
	title = strings.ReplaceAll(title, " ", "-")
	// Remove special characters
	var result []rune
	for _, r := range title {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result = append(result, r)
		}
	}
	slug := string(result)
	if len(slug) > 20 { //nolint:mnd // reasonable max length for slug
		slug = slug[:20]
	}

	shortID := request.ID
	if len(shortID) > 8 { //nolint:mnd // 8 chars of UUID is enough for uniqueness
		shortID = shortID[:8]
	}

	return fmt.Sprintf("%s-%s", slug, shortID)
}

func (s *gitService) generateTerragruntConfig(request *model.ResourceRequest, moduleRepo *model.GitRepository) (string, error) {
	// Parse the spec to get variables
	var vars map[string]interface{}
	if err := json.Unmarshal([]byte(request.Spec), &vars); err != nil {
		vars = make(map[string]interface{})
	}

	// Build the module source
	moduleSource := ""
	if request.TfModule != nil {
		moduleSource = request.TfModule.Source
		if request.TfModule.Version != "" {
			moduleSource = fmt.Sprintf("%s?ref=%s", moduleSource, request.TfModule.Version)
		}
	} else if moduleRepo != nil {
		// Use default modules repo
		moduleSource = moduleRepo.URL
	}

	// Template for terragrunt.hcl
	tmpl := `# Terragrunt configuration for {{ .NodeName }}
# Generated by VC Lab Platform
# Request ID: {{ .RequestID }}
# Created: {{ .CreatedAt }}

terraform {
  source = "{{ .ModuleSource }}"
}

include "root" {
  path = find_in_parent_folders()
}

inputs = {
{{- range $key, $value := .Vars }}
  {{ $key }} = {{ $value | formatValue }}
{{- end }}
}
`

	funcMap := template.FuncMap{
		"formatValue": func(v interface{}) string {
			switch val := v.(type) {
			case string:
				return fmt.Sprintf("%q", val)
			case float64:
				if val == float64(int(val)) {
					return fmt.Sprintf("%d", int(val))
				}
				return fmt.Sprintf("%f", val)
			case bool:
				return fmt.Sprintf("%t", val)
			default:
				b, _ := json.Marshal(val) //nolint:errcheck // will not fail
				return string(b)
			}
		},
	}

	t, err := template.New("terragrunt").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return "", err
	}

	data := map[string]interface{}{
		"NodeName":     s.generateNodeName(request),
		"RequestID":    request.ID,
		"CreatedAt":    time.Now().Format(time.RFC3339),
		"ModuleSource": moduleSource,
		"Vars":         vars,
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (s *gitService) commitPendingConfig(ctx context.Context, config *model.NodeConfig, storageRepo *model.GitRepository) (string, error) {
	// Clone the repo
	repoPath := filepath.Join(s.workDir, storageRepo.ID, fmt.Sprintf("pending-%s", config.ID))
	if err := s.CloneRepository(ctx, storageRepo, repoPath); err != nil {
		return "", fmt.Errorf("failed to clone repository: %w", err)
	}
	defer os.RemoveAll(repoPath) //nolint:errcheck // best effort cleanup

	// Write the config file
	configFilePath := filepath.Join(repoPath, storageRepo.BasePath, config.Path, "terragrunt.hcl")
	if err := os.MkdirAll(filepath.Dir(configFilePath), dirPerm); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(configFilePath, []byte(config.TerragruntConfig), filePerm); err != nil {
		return "", fmt.Errorf("failed to write config file: %w", err)
	}

	// Commit and push
	message := fmt.Sprintf("Add pending node config: %s", config.Name)
	return s.CommitAndPush(ctx, repoPath, []string{configFilePath}, message)
}

// ListModulesFromGit lists Terraform modules from the default modules git repository.
func (s *gitService) ListModulesFromGit(ctx context.Context) ([]GitModule, error) {
	return s.scanModulesFromGit(ctx, false)
}

// SyncModulesFromGit forces a refresh of modules from the git repository.
func (s *gitService) SyncModulesFromGit(ctx context.Context) ([]GitModule, error) {
	return s.scanModulesFromGit(ctx, true)
}

// Constants for module scanning.
const (
	maxDescriptionLength = 200
	splitParts           = 3
)

// scanModulesFromGit scans the modules repository for Terraform modules.
func (s *gitService) scanModulesFromGit(ctx context.Context, forceRefresh bool) ([]GitModule, error) {
	// Get the default modules repository
	moduleRepo, err := s.gitRepoRepo.GetDefaultByType(ctx, model.GitRepoTypeModules)
	if err != nil {
		return nil, fmt.Errorf("no default modules repository configured: %w", err)
	}

	// Determine the local path for the cloned repository
	repoPath := filepath.Join(s.workDir, "modules", moduleRepo.ID)

	// Check if we need to clone or pull
	if _, statErr := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(statErr) || forceRefresh {
		// Clone the repository
		if cloneErr := s.CloneRepository(ctx, moduleRepo, repoPath); cloneErr != nil {
			return nil, fmt.Errorf("failed to clone modules repository: %w", cloneErr)
		}
	} else {
		// Pull latest changes
		if pullErr := s.PullChanges(ctx, repoPath); pullErr != nil {
			s.logger.Warn("failed to pull changes, using cached version", zap.Error(pullErr))
		}
	}

	// Update last sync time
	now := time.Now()
	moduleRepo.LastSyncAt = &now
	if updateErr := s.gitRepoRepo.Update(ctx, moduleRepo); updateErr != nil {
		s.logger.Warn("failed to update last sync time", zap.Error(updateErr))
	}

	// Scan for Terraform modules in the repository
	basePath := filepath.Join(repoPath, moduleRepo.BasePath)
	modules, err := s.scanTerraformModules(basePath, moduleRepo.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to scan modules: %w", err)
	}

	// Sync discovered modules to database (terraform_modules table)
	if syncErr := s.syncModulesToDatabase(ctx, modules); syncErr != nil {
		s.logger.Warn("failed to sync modules to database", zap.Error(syncErr))
	}

	return modules, nil
}

// scanTerraformModules scans a directory for Terraform modules.
func (s *gitService) scanTerraformModules(basePath, repoURL string) ([]GitModule, error) {
	var modules []GitModule

	// Walk through the directory structure
	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		}

		// Skip non-directories and the base path itself
		if !info.IsDir() || path == basePath {
			return nil
		}

		// Check if this directory contains .tf files (is a Terraform module)
		module, skipDir := s.processModuleDirectory(path, basePath, repoURL, info)
		if module != nil {
			modules = append(modules, *module)
		}
		if skipDir {
			return filepath.SkipDir
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return modules, nil
}

// processModuleDirectory checks if a directory is a Terraform module and returns the module info.
func (s *gitService) processModuleDirectory(path, basePath, repoURL string, info os.FileInfo) (*GitModule, bool) {
	hasTfFiles, modErr := s.hasTerraformFiles(path)
	if modErr != nil {
		s.logger.Warn("error checking terraform files", zap.String("path", sanitize.Path(path)), zap.Error(modErr))
		return nil, false
	}

	if !hasTfFiles {
		return nil, false
	}

	relPath, relErr := filepath.Rel(basePath, path)
	if relErr != nil {
		s.logger.Warn("error getting relative path", zap.String("path", sanitize.Path(path)), zap.Error(relErr))
		return nil, false
	}

	module := &GitModule{
		Name:        info.Name(),
		Path:        relPath,
		Source:      fmt.Sprintf("%s//%s", repoURL, relPath),
		Description: s.extractModuleDescription(path),
		Variables:   s.extractVariableNames(path),
		Outputs:     s.extractOutputNames(path),
	}

	// Don't recurse into module subdirectories (modules don't contain modules)
	return module, true
}

// hasTerraformFiles checks if a directory contains Terraform (.tf) files.
func (s *gitService) hasTerraformFiles(dirPath string) (bool, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return false, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".tf") {
			return true, nil
		}
	}

	return false, nil
}

// extractModuleDescription tries to extract a description from README.md or module header.
func (s *gitService) extractModuleDescription(modulePath string) string {
	// Try README.md first
	readmePath := filepath.Join(modulePath, "README.md")
	content, err := os.ReadFile(readmePath) // #nosec G304 --  path is constructed from controlled input
	if err != nil {
		return ""
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip headers
		if strings.HasPrefix(line, "#") {
			continue
		}
		// Return first non-empty paragraph
		if line != "" {
			if len(line) > maxDescriptionLength {
				return line[:maxDescriptionLength] + "..."
			}
			return line
		}
	}

	return ""
}

// extractVariableNames extracts variable names from variables.tf or *.tf files.
func (s *gitService) extractVariableNames(modulePath string) []string {
	var variables []string

	// Check for variables.tf first
	entries, err := os.ReadDir(modulePath)
	if err != nil {
		return variables
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tf") {
			continue
		}

		filePath := filepath.Join(modulePath, entry.Name())
		content, err := os.ReadFile(filePath) // #nosec G304 --  path is constructed from controlled input
		if err != nil {
			continue
		}

		// Simple regex-free variable extraction
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "variable ") {
				// Extract variable name from 'variable "name" {'
				parts := strings.SplitN(line, "\"", splitParts)
				if len(parts) >= 2 {
					variables = append(variables, parts[1])
				}
			}
		}
	}

	return variables
}

// extractOutputNames extracts output names from outputs.tf or *.tf files.
func (s *gitService) extractOutputNames(modulePath string) []string {
	var outputs []string

	entries, err := os.ReadDir(modulePath)
	if err != nil {
		return outputs
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tf") {
			continue
		}

		filePath := filepath.Join(modulePath, entry.Name())
		content, err := os.ReadFile(filePath) // #nosec G304 --  path is constructed from controlled input
		if err != nil {
			continue
		}

		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "output ") {
				// Extract output name from 'output "name" {'
				parts := strings.SplitN(line, "\"", splitParts)
				if len(parts) >= 2 {
					outputs = append(outputs, parts[1])
				}
			}
		}
	}

	return outputs
}

// syncModulesToDatabase syncs discovered git modules to the terraform_modules database table.
//
//nolint:unparam // error return is for future use and consistency
func (s *gitService) syncModulesToDatabase(ctx context.Context, gitModules []GitModule) error {
	for _, gm := range gitModules {
		// Check if module with this source already exists
		existingModule, err := s.tfModuleRepo.GetBySource(ctx, gm.Source)
		if err == nil && existingModule != nil {
			// Module exists, update it
			existingModule.Name = gm.Name
			existingModule.Description = gm.Description
			variablesJSON, _ := json.Marshal(gm.Variables) //nolint:errcheck // will not fail with slice
			existingModule.Variables = string(variablesJSON)
			if updateErr := s.tfModuleRepo.Update(ctx, existingModule); updateErr != nil {
				s.logger.Warn("failed to update terraform module",
					zap.String("name", sanitize.ForLog(gm.Name)),
					zap.Error(updateErr),
				)
			}
			continue
		}

		// Create new module
		variablesJSON, _ := json.Marshal(gm.Variables) //nolint:errcheck // will not fail with slice
		newModule := &model.TerraformModule{
			Name:        gm.Name,
			Source:      gm.Source,
			Description: gm.Description,
			Variables:   string(variablesJSON),
			Status:      1, // active
		}
		if createErr := s.tfModuleRepo.Create(ctx, newModule); createErr != nil {
			s.logger.Warn("failed to create terraform module",
				zap.String("name", sanitize.ForLog(gm.Name)),
				zap.Error(createErr),
			)
		} else {
			s.logger.Info("synced terraform module to database",
				zap.String("name", sanitize.ForLog(gm.Name)),
				zap.String("source", sanitize.URL(gm.Source)),
			)
		}
	}
	return nil
}
