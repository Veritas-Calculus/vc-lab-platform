// Package terraform provides Terraform execution utilities.
package terraform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// Executor handles Terraform operations.
type Executor struct {
	logger *zap.Logger
}

// ExecutionResult contains the result of a Terraform execution.
type ExecutionResult struct {
	Success  bool              `json:"success"`
	Output   string            `json:"output"`
	Error    string            `json:"error"`
	Duration time.Duration     `json:"duration"`
	Outputs  map[string]string `json:"outputs"`
}

// Config represents configuration for terraform file generation.
type Config struct {
	Provider    string                 `json:"provider"`    // pve, vmware, openstack
	Environment string                 `json:"environment"` // dev, test, staging, prod
	Spec        map[string]interface{} `json:"spec"`        // Resource specifications

	// Git authentication for module downloads
	GitHost     string `json:"git_host"`     // Git server host (e.g., git.example.com)
	GitUsername string `json:"git_username"` // Git username
	GitToken    string `json:"git_token"`    // Git access token

	// Zone configuration
	RegistryEndpoint string `json:"registry_endpoint"` // Registry mirror URL
	RegistryToken    string `json:"registry_token"`    // Registry auth token

	// Provider configuration from Zone.TfProvider
	ProviderSource    string `json:"provider_source"`    // e.g., bpg/proxmox
	ProviderNamespace string `json:"provider_namespace"` // e.g., bpg
	ProviderVersion   string `json:"provider_version"`   // e.g., ~> 0.38.0

	// Module configuration from Zone.TfModule
	ModuleSource  string `json:"module_source"`  // Git URL or registry path
	ModuleVersion string `json:"module_version"` // Version/tag

	// Cluster credentials
	ClusterEndpoint string `json:"cluster_endpoint"` // API endpoint
	ClusterUsername string `json:"cluster_username"` // Username/AccessKey
	ClusterPassword string `json:"cluster_password"` // Password/SecretKey
	ClusterToken    string `json:"cluster_token"`    // Optional token
}

// File permission constants.
const (
	dirPerm  = 0o750 // Directory permissions (rwxr-x---)
	filePerm = 0o644 // File permissions (rw-r--r--)
)

// ansiRegex matches ANSI escape sequences.
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// stripANSI removes ANSI escape codes from a string.
func stripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

// NewExecutor creates a new Terraform executor.
func NewExecutor(logger *zap.Logger) *Executor {
	return &Executor{
		logger: logger,
	}
}

// InitWithConfig initializes a Terraform working directory with Git credentials.
func (e *Executor) InitWithConfig(workDir string, config Config) error {
	// Configure Git credentials if provided
	if config.GitHost != "" && config.GitToken != "" {
		if err := e.configureGitCredentials(workDir, config); err != nil {
			e.logger.Warn("failed to configure git credentials", zap.Error(err))
		}
	}

	return e.Init(workDir)
}

// netrcFilePermission is the permission for .netrc files.
const netrcFilePermission = 0o600

// configureGitCredentials sets up Git credentials for module downloads.
func (e *Executor) configureGitCredentials(workDir string, config Config) error {
	// Create .netrc file for HTTPS authentication
	netrcContent := fmt.Sprintf("machine %s\nlogin %s\npassword %s\n",
		config.GitHost,
		config.GitUsername,
		config.GitToken,
	)
	netrcPath := filepath.Join(workDir, ".netrc")
	if err := os.WriteFile(netrcPath, []byte(netrcContent), netrcFilePermission); err != nil {
		return fmt.Errorf("failed to write .netrc: %w", err)
	}

	// Set HOME to workDir so git uses our .netrc
	if err := os.Setenv("HOME", workDir); err != nil {
		return fmt.Errorf("failed to set HOME: %w", err)
	}
	e.logger.Info("configured git credentials", zap.String("host", config.GitHost))
	return nil
}

// isTerragrunt checks if the working directory uses Terragrunt.
func (e *Executor) isTerragrunt(workDir string) bool {
	hclPath := filepath.Join(workDir, "terragrunt.hcl")
	_, err := os.Stat(hclPath)
	return err == nil
}

// buildEnv builds environment variables for Terraform/Terragrunt execution.
func (e *Executor) buildEnv(workDir string) []string {
	env := os.Environ()

	// Check if .terraformrc exists and set TF_CLI_CONFIG_FILE
	rcPath := filepath.Join(workDir, ".terraformrc")
	if _, err := os.Stat(rcPath); err == nil {
		env = append(env, fmt.Sprintf("TF_CLI_CONFIG_FILE=%s", rcPath))
		e.logger.Info("using custom terraform config", zap.String("config", rcPath))
	}

	// Check if .netrc exists and set HOME to workDir
	netrcPath := filepath.Join(workDir, ".netrc")
	if _, err := os.Stat(netrcPath); err == nil {
		env = append(env, fmt.Sprintf("HOME=%s", workDir))
		e.logger.Info("using .netrc for git authentication", zap.String("path", netrcPath))
	}

	return env
}

// Init initializes a Terraform/Terragrunt working directory.
func (e *Executor) Init(workDir string) error {
	ctx := context.Background()

	var cmd *exec.Cmd
	if e.isTerragrunt(workDir) {
		cmd = exec.CommandContext(ctx, "terragrunt", "init", "--terragrunt-non-interactive")
		e.logger.Info("using terragrunt init")
	} else {
		cmd = exec.CommandContext(ctx, "terraform", "init", "-no-color")
		e.logger.Info("using terraform init")
	}
	cmd.Dir = workDir
	cmd.Env = e.buildEnv(workDir)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		e.logger.Error("init failed",
			zap.String("stderr", stderr.String()),
			zap.String("stdout", stdout.String()),
			zap.Error(err),
		)
		return fmt.Errorf("init failed: %s", stripANSI(stderr.String()))
	}

	e.logger.Info("init completed", zap.String("output", stripANSI(stdout.String())))
	return nil
}

// runCommand executes a terraform/terragrunt command and returns the result.
func (e *Executor) runCommand(workDir, operation string, tfArgs, tgArgs []string) *ExecutionResult {
	start := time.Now()
	result := &ExecutionResult{}
	ctx := context.Background()

	var cmd *exec.Cmd
	if e.isTerragrunt(workDir) {
		cmd = exec.CommandContext(ctx, "terragrunt", tgArgs...)
	} else {
		cmd = exec.CommandContext(ctx, "terraform", tfArgs...)
	}
	cmd.Dir = workDir
	cmd.Env = e.buildEnv(workDir)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result.Duration = time.Since(start)
	result.Output = stripANSI(stdout.String())

	if err != nil {
		e.logger.Error(operation+" failed",
			zap.Error(err),
			zap.String("stderr", stripANSI(stderr.String())),
		)
		result.Error = stripANSI(stderr.String())
		return result
	}

	result.Success = true
	return result
}

// Plan runs terraform/terragrunt plan.
func (e *Executor) Plan(workDir string) *ExecutionResult {
	return e.runCommand(workDir, "plan",
		[]string{"plan", "-no-color", "-out=tfplan"},
		[]string{"plan", "--terragrunt-non-interactive", "-out=tfplan"},
	)
}

// Apply applies the Terraform/Terragrunt plan.
func (e *Executor) Apply(workDir string) *ExecutionResult {
	result := e.runCommand(workDir, "apply",
		[]string{"apply", "-no-color", "-auto-approve", "tfplan"},
		[]string{"apply", "--terragrunt-non-interactive", "-auto-approve", "tfplan"},
	)
	if result.Success {
		result.Outputs = e.GetOutputs(workDir)
	}
	return result
}

// Destroy destroys the Terraform/Terragrunt-managed infrastructure.
func (e *Executor) Destroy(workDir string) *ExecutionResult {
	return e.runCommand(workDir, "destroy",
		[]string{"destroy", "-no-color", "-auto-approve"},
		[]string{"destroy", "--terragrunt-non-interactive", "-auto-approve"},
	)
}

// GetOutputs retrieves Terraform/Terragrunt outputs.
func (e *Executor) GetOutputs(workDir string) map[string]string {
	ctx := context.Background()

	var cmd *exec.Cmd
	if e.isTerragrunt(workDir) {
		cmd = exec.CommandContext(ctx, "terragrunt", "output", "-json")
	} else {
		cmd = exec.CommandContext(ctx, "terraform", "output", "-json")
	}
	cmd.Dir = workDir
	cmd.Env = e.buildEnv(workDir)

	output, err := cmd.Output()
	if err != nil {
		e.logger.Error("failed to get outputs", zap.Error(err))
		return nil
	}

	var rawOutputs map[string]interface{}
	if err := json.Unmarshal(output, &rawOutputs); err != nil {
		return nil
	}

	outputs := make(map[string]string)
	for key, val := range rawOutputs {
		if valMap, ok := val.(map[string]interface{}); ok {
			if value, ok := valMap["value"].(string); ok {
				outputs[key] = value
			}
		}
	}

	return outputs
}

// GenerateTFFiles generates Terraform configuration files for a resource.
func (e *Executor) GenerateTFFiles(workDir string, config Config) error {
	// Create work directory
	if err := os.MkdirAll(workDir, dirPerm); err != nil {
		return fmt.Errorf("failed to create work directory: %w", err)
	}

	// Generate .terraformrc for registry mirror if configured
	if config.RegistryEndpoint != "" {
		terraformRC := generateTerraformRC(config)
		rcPath := filepath.Join(workDir, ".terraformrc")
		if err := os.WriteFile(rcPath, []byte(terraformRC), filePerm); err != nil {
			return fmt.Errorf("failed to write .terraformrc: %w", err)
		}
		e.logger.Info("generated .terraformrc for registry mirror",
			zap.String("registry", config.RegistryEndpoint),
		)
	}

	// Determine whether to use Terragrunt or pure Terraform
	if config.ModuleSource != "" {
		// Use Terragrunt for module-based deployments
		return e.generateTerragruntFiles(workDir, config)
	}

	// Pure Terraform for raw provider configurations
	var mainTF string
	var err error
	mainTF, err = generateMainTF(config)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(workDir, "main.tf"), []byte(mainTF), filePerm); err != nil {
		return err
	}

	// Generate terraform.tfvars with credentials
	tfvars := generateTFVars(config)
	if err := os.WriteFile(filepath.Join(workDir, "terraform.tfvars"), []byte(tfvars), filePerm); err != nil {
		return fmt.Errorf("failed to write terraform.tfvars: %w", err)
	}

	e.logger.Info("generated terraform files",
		zap.String("work_dir", workDir),
		zap.String("provider", config.Provider),
	)

	return nil
}

// generateTerragruntFiles generates Terragrunt configuration files.
func (e *Executor) generateTerragruntFiles(workDir string, config Config) error {
	// Format module source as git::https://... if needed
	moduleSource := formatModuleSource(config.ModuleSource, config.ModuleVersion)

	// Generate terragrunt.hcl
	terragruntHCL := generateTerragruntHCL(config, moduleSource)
	hclPath := filepath.Join(workDir, "terragrunt.hcl")
	if err := os.WriteFile(hclPath, []byte(terragruntHCL), filePerm); err != nil {
		return fmt.Errorf("failed to write terragrunt.hcl: %w", err)
	}

	e.logger.Info("generated terragrunt files",
		zap.String("work_dir", workDir),
		zap.String("module_source", moduleSource),
	)

	return nil
}

// formatModuleSource converts a module URL to git::https:// format if needed.
func formatModuleSource(source, version string) string {
	// If already in git:: format, add version if needed and return
	if strings.HasPrefix(source, "git::") {
		return addVersionRef(source, version)
	}

	// Convert https:// to git::https://
	if strings.HasPrefix(source, "https://") {
		gitSource := convertHTTPSToGit(source)
		return addVersionRef(gitSource, version)
	}

	// Return as-is for other formats (registry modules, etc.)
	return source
}

// convertHTTPSToGit converts an https:// URL to git::https:// format.
// Example: https://git.example.com/repo//subpath -> git::https://git.example.com/repo.git//subpath
func convertHTTPSToGit(source string) string {
	// Already has .git suffix, just add git:: prefix
	if strings.Contains(source, ".git") {
		return "git::" + source
	}

	// Find the subpath separator (//) - but skip the https:// part
	// We need to find // after the host part
	withoutScheme := strings.TrimPrefix(source, "https://")

	// Find the first // in the path (subpath separator)
	if idx := strings.Index(withoutScheme, "//"); idx > 0 {
		// Split into repo part and subpath
		repoPart := withoutScheme[:idx]
		subPath := withoutScheme[idx:]
		// Add .git before the subpath
		return "git::https://" + repoPart + ".git" + subPath
	}

	// No subpath, just append .git
	return "git::" + source + ".git"
}

// addVersionRef adds ?ref=version to a source URL if not already present.
func addVersionRef(source, version string) string {
	if version != "" && !strings.Contains(source, "?ref=") {
		return source + "?ref=" + version
	}
	return source
}

// providerPVE is the constant for Proxmox VE provider type.
const providerPVE = "pve"

// generateTerragruntHCL generates a terragrunt.hcl file.
func generateTerragruntHCL(config Config, moduleSource string) string {
	inputs := buildTerragruntInputs(config)

	return fmt.Sprintf(`# Generated by VC Lab Platform
# Provider: %s
# Environment: %s

terraform {
  source = "%s"
}

inputs = {
%s
}
`, config.Provider, config.Environment, moduleSource, strings.Join(inputs, "\n"))
}

// buildTerragruntInputs builds the inputs block for terragrunt.hcl.
func buildTerragruntInputs(config Config) []string {
	var inputs []string

	// Add provider credentials based on provider type
	switch config.Provider {
	case providerPVE:
		inputs = append(inputs, buildPVEInputs(config)...)
	default:
		inputs = append(inputs, buildGenericInputs(config)...)
	}

	// Add spec values
	for key, value := range config.Spec {
		inputs = append(inputs, formatInputValue(key, value))
	}

	return inputs
}

// buildPVEInputs builds Proxmox VE specific inputs.
func buildPVEInputs(config Config) []string {
	var inputs []string
	if config.ClusterEndpoint != "" {
		inputs = append(inputs, fmt.Sprintf("  proxmox_host = %q", extractHost(config.ClusterEndpoint)))
	}
	if config.ClusterUsername != "" {
		inputs = append(inputs, fmt.Sprintf("  pm_user = %q", config.ClusterUsername))
	}
	if config.ClusterToken != "" {
		inputs = append(inputs, fmt.Sprintf("  pm_api_token = %q", config.ClusterToken))
	} else if config.ClusterPassword != "" {
		inputs = append(inputs, fmt.Sprintf("  pm_password = %q", config.ClusterPassword))
	}
	return inputs
}

// buildGenericInputs builds generic provider inputs.
func buildGenericInputs(config Config) []string {
	var inputs []string
	if config.ClusterEndpoint != "" {
		inputs = append(inputs, fmt.Sprintf("  api_endpoint = %q", config.ClusterEndpoint))
	}
	if config.ClusterUsername != "" {
		inputs = append(inputs, fmt.Sprintf("  api_username = %q", config.ClusterUsername))
	}
	if config.ClusterToken != "" {
		inputs = append(inputs, fmt.Sprintf("  api_token = %q", config.ClusterToken))
	}
	return inputs
}

// extractHost extracts hostname from URL.
func extractHost(endpoint string) string {
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")
	if idx := strings.Index(endpoint, "/"); idx > 0 {
		return endpoint[:idx]
	}
	if idx := strings.Index(endpoint, ":"); idx > 0 {
		return endpoint[:idx]
	}
	return endpoint
}

// formatInputValue formats a value for HCL.
func formatInputValue(key string, value interface{}) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("  %s = %q", key, v)
	case float64:
		if v == float64(int64(v)) {
			return fmt.Sprintf("  %s = %d", key, int64(v))
		}
		return fmt.Sprintf("  %s = %f", key, v)
	case bool:
		return fmt.Sprintf("  %s = %t", key, v)
	case int, int64:
		return fmt.Sprintf("  %s = %d", key, v)
	default:
		jsonVal, _ := json.Marshal(v) //nolint:errcheck // will not fail
		return fmt.Sprintf("  %s = %s", key, string(jsonVal))
	}
}

// generateTerraformRC generates a .terraformrc file for registry mirror.
func generateTerraformRC(config Config) string {
	// Normalize registry endpoint - remove https:// prefix if present
	endpoint := config.RegistryEndpoint
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimSuffix(endpoint, "/")

	var creds string
	if config.RegistryToken != "" {
		creds = fmt.Sprintf(`
credentials "%s" {
  token = "%s"
}`, endpoint, config.RegistryToken)
	}

	return fmt.Sprintf(`provider_installation {
  network_mirror {
    url = "https://%s/v1/providers/"
    include = ["*/*"]
  }
  direct {
    exclude = ["*/*"]
  }
}
%s`, endpoint, creds)
}

// generateTFVars generates terraform.tfvars with credentials and resource specs.
//
//nolint:gocognit,goconst,gocritic,nestif,gocyclo // complexity is inherent to tfvars generation
func generateTFVars(config Config) string {
	var lines []string

	// Add provider-specific credential variables
	switch config.Provider {
	case providerPVE:
		// Proxmox-specific variable names
		if config.ClusterEndpoint != "" {
			lines = append(lines, fmt.Sprintf(`proxmox_api_url = "%s"`, config.ClusterEndpoint))
		}
		if config.ClusterUsername != "" {
			lines = append(lines, fmt.Sprintf(`proxmox_user = "%s"`, config.ClusterUsername))
		}
		if config.ClusterPassword != "" {
			lines = append(lines, fmt.Sprintf(`proxmox_password = "%s"`, config.ClusterPassword))
		}
	case "vmware":
		// VMware-specific variable names
		if config.ClusterEndpoint != "" {
			lines = append(lines, fmt.Sprintf(`vsphere_server = "%s"`, config.ClusterEndpoint))
		}
		if config.ClusterUsername != "" {
			lines = append(lines, fmt.Sprintf(`vsphere_user = "%s"`, config.ClusterUsername))
		}
		if config.ClusterPassword != "" {
			lines = append(lines, fmt.Sprintf(`vsphere_password = "%s"`, config.ClusterPassword))
		}
	case "openstack":
		// OpenStack-specific variable names
		if config.ClusterEndpoint != "" {
			lines = append(lines, fmt.Sprintf(`openstack_auth_url = "%s"`, config.ClusterEndpoint))
		}
		if config.ClusterUsername != "" {
			lines = append(lines, fmt.Sprintf(`openstack_user = "%s"`, config.ClusterUsername))
		}
		if config.ClusterPassword != "" {
			lines = append(lines, fmt.Sprintf(`openstack_password = "%s"`, config.ClusterPassword))
		}
	default:
		// Generic variable names for module-based deployments
		if config.ClusterEndpoint != "" {
			lines = append(lines, fmt.Sprintf(`api_endpoint = "%s"`, config.ClusterEndpoint))
		}
		if config.ClusterUsername != "" {
			lines = append(lines, fmt.Sprintf(`api_username = "%s"`, config.ClusterUsername))
		}
		if config.ClusterPassword != "" {
			lines = append(lines, fmt.Sprintf(`api_password = "%s"`, config.ClusterPassword))
		}
		if config.ClusterToken != "" {
			lines = append(lines, fmt.Sprintf(`api_token = "%s"`, config.ClusterToken))
		}
	}

	// Add resource spec values
	if config.Spec != nil {
		if cpu, ok := config.Spec["cpu"]; ok {
			lines = append(lines, fmt.Sprintf(`cpu = %v`, cpu))
		}
		if memory, ok := config.Spec["memory"]; ok {
			lines = append(lines, fmt.Sprintf(`memory = %v`, memory))
		}
		if disk, ok := config.Spec["disk"]; ok {
			lines = append(lines, fmt.Sprintf(`disk = %v`, disk))
		}
		if vmName, ok := config.Spec["name"]; ok {
			lines = append(lines, fmt.Sprintf(`vm_name = "%v"`, vmName))
		} else {
			lines = append(lines, fmt.Sprintf(`vm_name = "%s-vm"`, config.Environment))
		}
		if network, ok := config.Spec["network"]; ok {
			lines = append(lines, fmt.Sprintf(`network = "%v"`, network))
		}
		if osImage, ok := config.Spec["os_image"]; ok {
			lines = append(lines, fmt.Sprintf(`os_image = "%v"`, osImage))
		}

		// Provider-specific spec values
		switch config.Provider {
		case providerPVE:
			if targetNode, ok := config.Spec["target_node"]; ok {
				lines = append(lines, fmt.Sprintf(`target_node = "%v"`, targetNode))
			} else {
				lines = append(lines, `target_node = "pve"`)
			}
			if templateName, ok := config.Spec["template_name"]; ok {
				lines = append(lines, fmt.Sprintf(`template_name = "%v"`, templateName))
			} else {
				lines = append(lines, `template_name = "ubuntu-template"`)
			}
			if storagePool, ok := config.Spec["storage_pool"]; ok {
				lines = append(lines, fmt.Sprintf(`storage_pool = "%v"`, storagePool))
			}
			if networkBridge, ok := config.Spec["network_bridge"]; ok {
				lines = append(lines, fmt.Sprintf(`network_bridge = "%v"`, networkBridge))
			}
		case "vmware":
			if datacenter, ok := config.Spec["datacenter"]; ok {
				lines = append(lines, fmt.Sprintf(`datacenter = "%v"`, datacenter))
			}
			if cluster, ok := config.Spec["cluster"]; ok {
				lines = append(lines, fmt.Sprintf(`cluster = "%v"`, cluster))
			}
			if datastore, ok := config.Spec["datastore"]; ok {
				lines = append(lines, fmt.Sprintf(`datastore = "%v"`, datastore))
			}
			if networkName, ok := config.Spec["network_name"]; ok {
				lines = append(lines, fmt.Sprintf(`network_name = "%v"`, networkName))
			}
			if templateName, ok := config.Spec["template_name"]; ok {
				lines = append(lines, fmt.Sprintf(`template_name = "%v"`, templateName))
			}
		case "openstack":
			if flavorName, ok := config.Spec["flavor_name"]; ok {
				lines = append(lines, fmt.Sprintf(`flavor_name = "%v"`, flavorName))
			}
			if imageName, ok := config.Spec["image_name"]; ok {
				lines = append(lines, fmt.Sprintf(`image_name = "%v"`, imageName))
			}
			if networkName, ok := config.Spec["network_name"]; ok {
				lines = append(lines, fmt.Sprintf(`network_name = "%v"`, networkName))
			}
			if tenantName, ok := config.Spec["tenant_name"]; ok {
				lines = append(lines, fmt.Sprintf(`tenant_name = "%v"`, tenantName))
			}
		}
	}

	// Add environment tag
	lines = append(lines, fmt.Sprintf(`environment = "%s"`, config.Environment))

	return strings.Join(lines, "\n") + "\n"
}

// generateMainTF generates the main Terraform configuration.
func generateMainTF(config Config) (string, error) {
	switch config.Provider {
	case providerPVE:
		return generateProxmoxTF(config), nil
	case "vmware":
		return generateVMwareTF(config), nil
	case "openstack":
		return generateOpenStackTF(config), nil
	default:
		return "", fmt.Errorf("unsupported provider: %s", config.Provider)
	}
}

// generateProxmoxTF generates Proxmox provider configuration.
//
//nolint:mnd // default values for VM configuration
func generateProxmoxTF(config Config) string {
	vmName := fmt.Sprintf("%s-vm", config.Environment)
	cpu := getIntValue(config.Spec, "cpu", 2)
	memory := getIntValue(config.Spec, "memory", 4096)
	disk := getIntValue(config.Spec, "disk", 50)

	return fmt.Sprintf(`terraform {
  required_providers {
    proxmox = {
      source  = "telmate/proxmox"
      version = "~> 2.9"
    }
  }
}

provider "proxmox" {
  pm_api_url = var.proxmox_api_url
  pm_user    = var.proxmox_user
  pm_password = var.proxmox_password
}

resource "proxmox_vm_qemu" "%s" {
  name        = var.vm_name
  target_node = var.target_node
  clone       = var.template_name
  
  cores    = %d
  memory   = %d
  
  disk {
    size    = "%dG"
    type    = "scsi"
    storage = var.storage_pool
  }
  
  network {
    model  = "virtio"
    bridge = var.network_bridge
  }

  tags = var.tags
}

variable "proxmox_api_url" {
  description = "Proxmox API URL"
  type        = string
}

variable "proxmox_user" {
  description = "Proxmox user"
  type        = string
}

variable "proxmox_password" {
  description = "Proxmox password"
  type        = string
  sensitive   = true
}

variable "vm_name" {
  description = "Name of the virtual machine"
  type        = string
}

variable "target_node" {
  description = "Target Proxmox node"
  type        = string
}

variable "template_name" {
  description = "Template to clone"
  type        = string
}

variable "storage_pool" {
  description = "Storage pool"
  type        = string
  default     = "local-lvm"
}

variable "network_bridge" {
  description = "Network bridge"
  type        = string
  default     = "vmbr0"
}

variable "tags" {
  description = "Tags for the VM"
  type        = string
  default     = ""
}

output "vm_id" {
  description = "ID of the created VM"
  value       = proxmox_vm_qemu.%s.vmid
}

output "vm_ip" {
  description = "IP address of the VM"
  value       = proxmox_vm_qemu.%s.default_ipv4_address
}
`, vmName, cpu, memory, disk, vmName, vmName)
}

// generateVMwareTF generates VMware provider configuration.
//
//nolint:mnd // default values for VM configuration
func generateVMwareTF(config Config) string {
	vmName := fmt.Sprintf("%s-vm", config.Environment)
	cpu := getIntValue(config.Spec, "cpu", 2)
	memory := getIntValue(config.Spec, "memory", 4096)
	disk := getIntValue(config.Spec, "disk", 50)

	return fmt.Sprintf(`terraform {
  required_providers {
    vsphere = {
      source  = "hashicorp/vsphere"
      version = "~> 2.4"
    }
  }
}

provider "vsphere" {
  user                 = var.vsphere_user
  password             = var.vsphere_password
  vsphere_server       = var.vsphere_server
  allow_unverified_ssl = true
}

data "vsphere_datacenter" "dc" {
  name = var.datacenter
}

data "vsphere_compute_cluster" "cluster" {
  name          = var.cluster
  datacenter_id = data.vsphere_datacenter.dc.id
}

data "vsphere_network" "network" {
  name          = var.network
  datacenter_id = data.vsphere_datacenter.dc.id
}

data "vsphere_datastore" "datastore" {
  name          = var.datastore
  datacenter_id = data.vsphere_datacenter.dc.id
}

data "vsphere_virtual_machine" "template" {
  name          = var.template_name
  datacenter_id = data.vsphere_datacenter.dc.id
}

resource "vsphere_virtual_machine" "%s" {
  name             = var.vm_name
  resource_pool_id = data.vsphere_compute_cluster.cluster.resource_pool_id
  datastore_id     = data.vsphere_datastore.datastore.id
  num_cpus         = %d
  memory           = %d

  network_interface {
    network_id = data.vsphere_network.network.id
  }

  disk {
    label            = "disk0"
    size             = %d
    thin_provisioned = true
  }

  clone {
    template_uuid = data.vsphere_virtual_machine.template.id
  }
}

variable "vsphere_user" {
  description = "vSphere user"
  type        = string
}

variable "vsphere_password" {
  description = "vSphere password"
  type        = string
  sensitive   = true
}

variable "vsphere_server" {
  description = "vSphere server"
  type        = string
}

variable "datacenter" {
  description = "Datacenter"
  type        = string
}

variable "cluster" {
  description = "Compute cluster"
  type        = string
}

variable "network" {
  description = "Network name"
  type        = string
}

variable "datastore" {
  description = "Datastore name"
  type        = string
}

variable "template_name" {
  description = "Template name"
  type        = string
}

variable "vm_name" {
  description = "Name of the VM"
  type        = string
}

output "vm_id" {
  description = "ID of the created VM"
  value       = vsphere_virtual_machine.%s.id
}

output "vm_ip" {
  description = "IP address of the VM"
  value       = vsphere_virtual_machine.%s.default_ip_address
}
`, vmName, cpu, memory, disk, vmName, vmName)
}

// generateOpenStackTF generates OpenStack provider configuration.
func generateOpenStackTF(config Config) string {
	instanceName := fmt.Sprintf("%s-instance", config.Environment)

	return fmt.Sprintf(`terraform {
  required_providers {
    openstack = {
      source  = "terraform-provider-openstack/openstack"
      version = "~> 1.51"
    }
  }
}

provider "openstack" {
  user_name   = var.os_username
  password    = var.os_password
  auth_url    = var.os_auth_url
  tenant_name = var.os_tenant_name
  region      = var.os_region
}

resource "openstack_compute_instance_v2" "%s" {
  name            = var.instance_name
  image_name      = var.image_name
  flavor_name     = var.flavor_name
  key_pair        = var.key_pair
  security_groups = var.security_groups

  network {
    name = var.network_name
  }
}

variable "os_username" {
  description = "OpenStack username"
  type        = string
}

variable "os_password" {
  description = "OpenStack password"
  type        = string
  sensitive   = true
}

variable "os_auth_url" {
  description = "OpenStack auth URL"
  type        = string
}

variable "os_tenant_name" {
  description = "OpenStack tenant name"
  type        = string
}

variable "os_region" {
  description = "OpenStack region"
  type        = string
}

variable "instance_name" {
  description = "Name of the instance"
  type        = string
}

variable "image_name" {
  description = "Image name"
  type        = string
}

variable "flavor_name" {
  description = "Flavor name"
  type        = string
}

variable "key_pair" {
  description = "Key pair name"
  type        = string
}

variable "security_groups" {
  description = "Security groups"
  type        = list(string)
  default     = ["default"]
}

variable "network_name" {
  description = "Network name"
  type        = string
}

output "instance_id" {
  description = "ID of the created instance"
  value       = openstack_compute_instance_v2.%s.id
}

output "instance_ip" {
  description = "IP address of the instance"
  value       = openstack_compute_instance_v2.%s.access_ip_v4
}
`, instanceName, instanceName, instanceName)
}

// getIntValue safely extracts an integer value from a map.
func getIntValue(m map[string]interface{}, key string, defaultVal int) int {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			// Try to parse string to int
			var i int
			if _, err := fmt.Sscanf(v, "%d", &i); err == nil {
				return i
			}
		}
	}
	return defaultVal
}
