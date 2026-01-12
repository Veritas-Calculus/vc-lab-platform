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

// NewExecutor creates a new Terraform executor.
func NewExecutor(logger *zap.Logger) *Executor {
	return &Executor{
		logger: logger,
	}
}

// Init initializes a Terraform working directory.
func (e *Executor) Init(workDir string) error {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "terraform", "init", "-no-color")
	cmd.Dir = workDir

	// Check if .terraformrc exists and set TF_CLI_CONFIG_FILE
	rcPath := filepath.Join(workDir, ".terraformrc")
	if _, err := os.Stat(rcPath); err == nil {
		cmd.Env = append(os.Environ(), fmt.Sprintf("TF_CLI_CONFIG_FILE=%s", rcPath))
		e.logger.Info("using custom terraform config", zap.String("config", rcPath))
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		e.logger.Error("terraform init failed",
			zap.String("stderr", stderr.String()),
			zap.String("stdout", stdout.String()),
			zap.Error(err),
		)
		return fmt.Errorf("terraform init failed: %s", stderr.String())
	}

	e.logger.Info("terraform init completed", zap.String("output", stdout.String()))
	return nil
}

// Plan runs terraform plan.
func (e *Executor) Plan(workDir string) *ExecutionResult {
	start := time.Now()
	result := &ExecutionResult{}

	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "terraform", "plan", "-no-color", "-out=tfplan")
	cmd.Dir = workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result.Duration = time.Since(start)
	result.Output = stdout.String()

	if err != nil {
		e.logger.Error("terraform plan failed",
			zap.Error(err),
			zap.String("stderr", stderr.String()),
		)
		result.Error = stderr.String()
		return result
	}

	result.Success = true
	return result
}

// Apply applies the Terraform plan.
func (e *Executor) Apply(workDir string) *ExecutionResult {
	start := time.Now()
	result := &ExecutionResult{}

	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "terraform", "apply", "-no-color", "-auto-approve", "tfplan")
	cmd.Dir = workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result.Duration = time.Since(start)
	result.Output = stdout.String()

	if err != nil {
		e.logger.Error("terraform apply failed",
			zap.Error(err),
			zap.String("stderr", stderr.String()),
		)
		result.Error = stderr.String()
		return result
	}

	result.Success = true

	// Get outputs
	outputs := e.GetOutputs(workDir)
	result.Outputs = outputs

	return result
}

// Destroy destroys the Terraform-managed infrastructure.
func (e *Executor) Destroy(workDir string) *ExecutionResult {
	start := time.Now()
	result := &ExecutionResult{}

	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "terraform", "destroy", "-no-color", "-auto-approve")
	cmd.Dir = workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result.Duration = time.Since(start)
	result.Output = stdout.String()

	if err != nil {
		e.logger.Error("terraform destroy failed",
			zap.Error(err),
			zap.String("stderr", stderr.String()),
		)
		result.Error = stderr.String()
		return result
	}

	result.Success = true
	return result
}

// GetOutputs retrieves Terraform outputs.
func (e *Executor) GetOutputs(workDir string) map[string]string {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "terraform", "output", "-json")
	cmd.Dir = workDir

	output, err := cmd.Output()
	if err != nil {
		e.logger.Error("failed to get terraform outputs", zap.Error(err))
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

	// Generate main.tf based on whether we have a module or raw provider
	var mainTF string
	var err error
	if config.ModuleSource != "" {
		mainTF = generateModuleBasedTF(config)
	} else {
		mainTF, err = generateMainTF(config)
		if err != nil {
			return err
		}
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
		zap.Bool("use_module", config.ModuleSource != ""),
	)

	return nil
}

// generateTerraformRC generates a .terraformrc file for registry mirror.
func generateTerraformRC(config Config) string {
	var creds string
	if config.RegistryToken != "" {
		creds = fmt.Sprintf(`
credentials "%s" {
  token = "%s"
}`, config.RegistryEndpoint, config.RegistryToken)
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
%s`, config.RegistryEndpoint, creds)
}

// generateTFVars generates terraform.tfvars with credentials and resource specs.
//
//nolint:gocognit,goconst,gocritic,nestif,gocyclo // complexity is inherent to tfvars generation
func generateTFVars(config Config) string {
	var lines []string

	// Add provider-specific credential variables
	switch config.Provider {
	case "pve":
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
		case "pve":
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

// generateModuleBasedTF generates Terraform config using a module.
func generateModuleBasedTF(config Config) string {
	providerSource := config.ProviderSource
	if providerSource == "" {
		// Default providers for common types
		switch config.Provider {
		case "pve":
			providerSource = "bpg/proxmox"
		case "vmware":
			providerSource = "hashicorp/vsphere"
		case "openstack":
			providerSource = "terraform-provider-openstack/openstack"
		default:
			providerSource = "hashicorp/" + config.Provider
		}
	}

	providerVersion := config.ProviderVersion
	if providerVersion == "" {
		providerVersion = ">= 0.1.0"
	}

	moduleVersion := ""
	if config.ModuleVersion != "" {
		moduleVersion = fmt.Sprintf(`  version = "%s"`, config.ModuleVersion) //nolint:gocritic // %q would add extra escaping
	}

	return fmt.Sprintf(`terraform {
  required_providers {
    %s = {
      source  = "%s"
      version = "%s"
    }
  }
}

variable "api_endpoint" {
  description = "API endpoint URL"
  type        = string
}

variable "api_username" {
  description = "API username"
  type        = string
  default     = ""
}

variable "api_password" {
  description = "API password"
  type        = string
  sensitive   = true
  default     = ""
}

variable "api_token" {
  description = "API token"
  type        = string
  sensitive   = true
  default     = ""
}

variable "cpu" {
  description = "Number of CPUs"
  type        = number
  default     = 2
}

variable "memory" {
  description = "Memory in MB"
  type        = number
  default     = 4096
}

variable "disk" {
  description = "Disk size in GB"
  type        = number
  default     = 50
}

variable "vm_name" {
  description = "Name of the virtual machine"
  type        = string
}

variable "network" {
  description = "Network configuration"
  type        = string
  default     = ""
}

variable "os_image" {
  description = "OS image/template to use"
  type        = string
  default     = ""
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "dev"
}

module "resource" {
  source = "%s"
%s

  api_endpoint = var.api_endpoint
  api_username = var.api_username
  api_password = var.api_password
  api_token    = var.api_token
  
  cpu     = var.cpu
  memory  = var.memory
  disk    = var.disk
  vm_name = var.vm_name
  network = var.network
  os_image = var.os_image
  environment = var.environment
}

output "resource_id" {
  description = "ID of the created resource"
  value       = try(module.resource.resource_id, "")
}

output "resource_ip" {
  description = "IP address of the resource"
  value       = try(module.resource.ip_address, "")
}
`, config.Provider, providerSource, providerVersion, config.ModuleSource, moduleVersion)
}

// generateMainTF generates the main Terraform configuration.
func generateMainTF(config Config) (string, error) {
	switch config.Provider {
	case "pve":
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
