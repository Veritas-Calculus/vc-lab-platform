// Package model defines the database models for the application.
package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel contains common fields for all models.
type BaseModel struct {
	ID        string         `gorm:"type:char(36);primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate generates a UUID before creating a record.
func (b *BaseModel) BeforeCreate(_ *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return nil
}

// UserSource represents how the user was created.
type UserSource string

// UserSource constants.
const (
	// UserSourceLocal represents a local user with password.
	UserSourceLocal UserSource = "local"
	// UserSourceLDAP represents an LDAP/AD user.
	UserSourceLDAP UserSource = "ldap"
	// UserSourceOIDC represents an OpenID Connect (SSO) user.
	UserSourceOIDC UserSource = "oidc"
	// UserSourceSAML represents a SAML SSO user.
	UserSourceSAML UserSource = "saml"
	// UserSourceOAuth2 represents an OAuth2 provider user.
	UserSourceOAuth2 UserSource = "oauth2"
)

// User represents a platform user.
type User struct {
	BaseModel
	Username     string     `gorm:"type:varchar(64);uniqueIndex;not null" json:"username"`
	Email        string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash string     `gorm:"type:varchar(255);not null" json:"-"`
	DisplayName  string     `gorm:"type:varchar(128)" json:"display_name"`
	Phone        string     `gorm:"type:varchar(20)" json:"phone"`
	Avatar       string     `gorm:"type:varchar(512)" json:"avatar"`
	Source       UserSource `gorm:"type:varchar(20);default:'local';not null" json:"source"` // User source: local, ldap, oidc, saml, oauth2
	ExternalID   string     `gorm:"type:varchar(255)" json:"external_id,omitempty"`          // External ID from SSO provider
	IsSystem     bool       `gorm:"default:false;not null" json:"is_system"`                 // System user (cannot be deleted)
	Status       int8       `gorm:"type:tinyint;default:1;not null" json:"status"`           // 0: disabled, 1: active
	LastLoginAt  *time.Time `json:"last_login_at"`
	LastLoginIP  string     `gorm:"type:varchar(45)" json:"last_login_ip"`
	Roles        []Role     `gorm:"many2many:user_roles;" json:"roles,omitempty"`
}

// TableName returns the table name for User.
func (User) TableName() string {
	return "users"
}

// Role represents a user role for RBAC.
type Role struct {
	BaseModel
	Name        string       `gorm:"type:varchar(64);uniqueIndex;not null" json:"name"`
	Code        string       `gorm:"type:varchar(64);uniqueIndex;not null" json:"code"`
	Description string       `gorm:"type:varchar(255)" json:"description"`
	IsSystem    bool         `gorm:"default:false;not null" json:"is_system"` // System role (cannot be deleted)
	Status      int8         `gorm:"type:tinyint;default:1;not null" json:"status"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
	Users       []User       `gorm:"many2many:user_roles;" json:"users,omitempty"`
}

// TableName returns the table name for Role.
func (Role) TableName() string {
	return "roles"
}

// Permission represents a system permission.
type Permission struct {
	BaseModel
	Name        string `gorm:"type:varchar(64);not null" json:"name"`
	Code        string `gorm:"type:varchar(128);uniqueIndex;not null" json:"code"`
	Description string `gorm:"type:varchar(255)" json:"description"`
	Resource    string `gorm:"type:varchar(64);not null" json:"resource"` // user, role, resource, etc.
	Action      string `gorm:"type:varchar(32);not null" json:"action"`   // create, read, update, delete
	Roles       []Role `gorm:"many2many:role_permissions;" json:"roles,omitempty"`
}

// TableName returns the table name for Permission.
func (Permission) TableName() string {
	return "permissions"
}

// Resource represents a computing resource (VM, container, etc.).
type Resource struct {
	BaseModel
	Name        string     `gorm:"type:varchar(128);not null" json:"name"`
	Type        string     `gorm:"type:varchar(32);not null" json:"type"`                     // vm, container, bare_metal
	Provider    string     `gorm:"type:varchar(32);not null" json:"provider"`                 // pve, vmware, openstack
	Status      string     `gorm:"type:varchar(32);not null;default:'pending'" json:"status"` // pending, provisioning, running, stopped, error
	Spec        string     `gorm:"type:json" json:"spec"`                                     // CPU, memory, disk specs as JSON
	IPAddress   string     `gorm:"type:varchar(45)" json:"ip_address"`
	HostName    string     `gorm:"type:varchar(255)" json:"hostname"`
	OwnerID     string     `gorm:"type:char(36);index;not null" json:"owner_id"`
	Owner       *User      `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Environment string     `gorm:"type:varchar(32);index;not null" json:"environment"` // dev, test, staging, prod
	ExternalID  string     `gorm:"type:varchar(255)" json:"external_id"`               // ID in the external provider
	ExpiresAt   *time.Time `json:"expires_at"`
	Tags        string     `gorm:"type:json" json:"tags"` // JSON array of tags
	Description string     `gorm:"type:text" json:"description"`
}

// TableName returns the table name for Resource.
func (Resource) TableName() string {
	return "resources"
}

// ResourceSpec represents the specification for a resource.
type ResourceSpec struct {
	CPU      int    `json:"cpu"`       // Number of CPU cores
	Memory   int    `json:"memory"`    // Memory in MB
	Disk     int    `json:"disk"`      // Disk size in GB
	DiskType string `json:"disk_type"` // ssd, hdd
	OSType   string `json:"os_type"`   // linux, windows
	OSImage  string `json:"os_image"`  // ubuntu-22.04, centos-7, etc.
	Network  string `json:"network"`   // Network configuration
}

// ResourceRequest represents a resource request/application.
type ResourceRequest struct {
	BaseModel
	Title                string             `gorm:"type:varchar(255);not null" json:"title"`
	Description          string             `gorm:"type:text" json:"description"`
	Spec                 string             `gorm:"type:json;not null" json:"spec"` // Requested spec
	Environment          string             `gorm:"type:varchar(32);not null" json:"environment"`
	Provider             string             `gorm:"type:varchar(32);not null" json:"provider"`
	Type                 string             `gorm:"type:varchar(32);not null" json:"type"` // vm, container, bare_metal
	RegionID             *string            `gorm:"type:char(36)" json:"region_id"`
	Region               *Region            `gorm:"foreignKey:RegionID" json:"region,omitempty"`
	ZoneID               *string            `gorm:"type:char(36)" json:"zone_id"`
	Zone                 *Zone              `gorm:"foreignKey:ZoneID" json:"zone,omitempty"`
	TfProviderID         *string            `gorm:"type:char(36)" json:"tf_provider_id"` // Selected Terraform provider
	TfProvider           *TerraformProvider `gorm:"foreignKey:TfProviderID" json:"tf_provider,omitempty"`
	TfModuleID           *string            `gorm:"type:char(36)" json:"tf_module_id"` // Selected Terraform module
	TfModule             *TerraformModule   `gorm:"foreignKey:TfModuleID" json:"tf_module,omitempty"`
	CredentialID         *string            `gorm:"type:char(36)" json:"credential_id"` // Selected credential for access
	Credential           *Credential        `gorm:"foreignKey:CredentialID" json:"credential,omitempty"`
	NodeConfigID         *string            `gorm:"type:char(36)" json:"node_config_id"` // Link to node configuration in storage repo
	Quantity             int                `gorm:"type:int;default:1;not null" json:"quantity"`
	Status               string             `gorm:"type:varchar(32);not null;default:'pending'" json:"status"` // pending, approved, rejected, provisioning, completed, failed
	RequesterID          string             `gorm:"type:char(36);index;not null" json:"requester_id"`
	Requester            *User              `gorm:"foreignKey:RequesterID" json:"requester,omitempty"`
	ApproverID           *string            `gorm:"type:char(36)" json:"approver_id"`
	Approver             *User              `gorm:"foreignKey:ApproverID" json:"approver,omitempty"`
	ApprovedAt           *time.Time         `json:"approved_at"`
	RejectedAt           *time.Time         `json:"rejected_at"`
	ProvisionStartedAt   *time.Time         `json:"provision_started_at"`
	ProvisionCompletedAt *time.Time         `json:"provision_completed_at"`
	Reason               string             `gorm:"type:text" json:"reason"`          // Reason for approval/rejection
	ProvisionLog         string             `gorm:"type:text" json:"provision_log"`   // Terraform execution log
	TerraformState       string             `gorm:"type:text" json:"terraform_state"` // Terraform state information
	ResourceID           *string            `gorm:"type:char(36)" json:"resource_id"` // Created resource ID
	Resource             *Resource          `gorm:"foreignKey:ResourceID" json:"resource,omitempty"`
	ExpiresAt            *time.Time         `json:"expires_at"`
	ErrorMessage         string             `gorm:"type:text" json:"error_message"` // Error message if provisioning failed
}

// TableName returns the table name for ResourceRequest.
func (ResourceRequest) TableName() string {
	return "resource_requests"
}

// AuditLog represents an audit log entry.
type AuditLog struct {
	ID         string    `gorm:"type:char(36);primaryKey" json:"id"`
	UserID     string    `gorm:"type:char(36);index" json:"user_id"`
	Username   string    `gorm:"type:varchar(64)" json:"username"`
	Action     string    `gorm:"type:varchar(64);not null" json:"action"`
	Resource   string    `gorm:"type:varchar(64);not null" json:"resource"`
	ResourceID string    `gorm:"type:varchar(255)" json:"resource_id"`
	Details    string    `gorm:"type:json" json:"details"`
	IPAddress  string    `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent  string    `gorm:"type:varchar(512)" json:"user_agent"`
	Status     string    `gorm:"type:varchar(32);not null" json:"status"` // success, failure
	CreatedAt  time.Time `gorm:"index" json:"created_at"`
}

// TableName returns the table name for AuditLog.
func (AuditLog) TableName() string {
	return "audit_logs"
}

// BeforeCreate generates a UUID before creating an audit log.
func (a *AuditLog) BeforeCreate(_ *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	return nil
}

// ProviderConfig represents a cloud/infrastructure provider configuration.
type ProviderConfig struct {
	BaseModel
	Name         string      `gorm:"type:varchar(128);not null" json:"name"`
	Type         string      `gorm:"type:varchar(32);not null" json:"type"`      // pve, vmware, openstack, aws, aliyun, gcp, azure
	Endpoint     string      `gorm:"type:varchar(512);not null" json:"endpoint"` // API endpoint URL
	Description  string      `gorm:"type:text" json:"description"`
	Config       string      `gorm:"type:json" json:"config"`                       // Provider-specific config as JSON
	Status       int8        `gorm:"type:tinyint;default:1;not null" json:"status"` // 0: disabled, 1: active
	IsDefault    bool        `gorm:"default:false" json:"is_default"`
	CredentialID *string     `gorm:"type:char(36)" json:"credential_id"` // Link to credential for authentication
	Credential   *Credential `gorm:"foreignKey:CredentialID" json:"credential,omitempty"`
}

// TableName returns the table name for ProviderConfig.
func (ProviderConfig) TableName() string {
	return "provider_configs"
}

// Credential represents cloud/infrastructure credentials.
type Credential struct {
	BaseModel
	Name        string          `gorm:"type:varchar(128);not null" json:"name"`
	Type        string          `gorm:"type:varchar(32);not null" json:"type"` // pve, vmware, openstack, aws, aliyun, gcp, azure
	ProviderID  *string         `gorm:"type:char(36)" json:"provider_id"`      // Optional link to provider config
	Provider    *ProviderConfig `gorm:"foreignKey:ProviderID" json:"provider,omitempty"`
	ZoneID      *string         `gorm:"type:char(36)" json:"zone_id"` // Link to zone this credential is for
	Zone        *Zone           `gorm:"foreignKey:ZoneID" json:"zone,omitempty"`
	Endpoint    string          `gorm:"type:varchar(512)" json:"endpoint"` // API endpoint URL for this credential
	AccessKey   string          `gorm:"type:varchar(512)" json:"-"`        // Encrypted access key / username
	SecretKey   string          `gorm:"type:varchar(512)" json:"-"`        // Encrypted secret key / password
	Token       string          `gorm:"type:text" json:"-"`                // Encrypted token (optional)
	Description string          `gorm:"type:text" json:"description"`
	Status      int8            `gorm:"type:tinyint;default:1;not null" json:"status"` // 0: disabled, 1: active
	LastUsedAt  *time.Time      `json:"last_used_at"`
	CreatedByID string          `gorm:"type:char(36);not null" json:"created_by_id"`
	CreatedBy   *User           `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`
}

// TableName returns the table name for Credential.
func (Credential) TableName() string {
	return "credentials"
}

// GitRepoType represents the type of git repository.
type GitRepoType string

// GitRepoType constants.
const (
	// GitRepoTypeModules represents a Terraform modules repository.
	GitRepoTypeModules GitRepoType = "modules"
	// GitRepoTypeStorage represents a node configuration storage repository.
	GitRepoTypeStorage GitRepoType = "storage"
)

// GitAuthType represents the authentication method for git repository.
type GitAuthType string

// GitAuthType constants.
const (
	// GitAuthTypeNone represents no authentication (public repo).
	GitAuthTypeNone GitAuthType = "none"
	// GitAuthTypeToken represents token/PAT authentication.
	GitAuthTypeToken GitAuthType = "token"
	// GitAuthTypePassword represents username/password authentication.
	GitAuthTypePassword GitAuthType = "password"
	// GitAuthTypeSSHKey represents SSH key authentication.
	GitAuthTypeSSHKey GitAuthType = "ssh_key"
)

// GitRepository represents a git repository for storing terraform modules or node configs.
type GitRepository struct {
	BaseModel
	Name        string      `gorm:"type:varchar(128);not null" json:"name"`
	Type        GitRepoType `gorm:"type:varchar(32);not null" json:"type"` // modules, storage
	URL         string      `gorm:"type:varchar(512);not null" json:"url"` // Git URL (https or ssh)
	Branch      string      `gorm:"type:varchar(128);default:'main'" json:"branch"`
	AuthType    GitAuthType `gorm:"type:varchar(32);default:'none'" json:"auth_type"` // Authentication type
	Username    string      `gorm:"type:varchar(256)" json:"username,omitempty"`      // Git auth username
	Token       string      `gorm:"type:text" json:"-"`                               // Git auth token/password (encrypted)
	SSHKey      string      `gorm:"type:text" json:"-"`                               // SSH private key (encrypted)
	BasePath    string      `gorm:"type:varchar(512);default:'/'" json:"base_path"`   // Base path within repo for configs
	Description string      `gorm:"type:text" json:"description"`
	Status      int8        `gorm:"type:tinyint;default:1;not null" json:"status"` // 0: disabled, 1: active
	IsDefault   bool        `gorm:"default:false" json:"is_default"`
	LastSyncAt  *time.Time  `json:"last_sync_at"`
}

// TableName returns the table name for GitRepository.
func (GitRepository) TableName() string {
	return "git_repositories"
}

// NodeConfigStatus represents the status of a node configuration.
type NodeConfigStatus string

// NodeConfigStatus constants.
const (
	// NodeConfigStatusPending represents a config waiting for approval.
	NodeConfigStatusPending NodeConfigStatus = "pending"
	// NodeConfigStatusApproved represents an approved config ready to provision.
	NodeConfigStatusApproved NodeConfigStatus = "approved"
	// NodeConfigStatusProvisioning represents a config being provisioned.
	NodeConfigStatusProvisioning NodeConfigStatus = "provisioning"
	// NodeConfigStatusActive represents a successfully provisioned config.
	NodeConfigStatusActive NodeConfigStatus = "active"
	// NodeConfigStatusFailed represents a failed provisioning.
	NodeConfigStatusFailed NodeConfigStatus = "failed"
	// NodeConfigStatusDestroying represents a config being destroyed.
	NodeConfigStatusDestroying NodeConfigStatus = "destroying"
	// NodeConfigStatusDestroyed represents a successfully destroyed config.
	NodeConfigStatusDestroyed NodeConfigStatus = "destroyed"
)

// NodeConfig represents a node configuration stored in the storage repository.
type NodeConfig struct {
	BaseModel
	Name              string           `gorm:"type:varchar(128);not null" json:"name"`                  // Node name (e.g., minio-01)
	Path              string           `gorm:"type:varchar(512);not null" json:"path"`                  // Path in storage repo (e.g., proxmox-ve/instance/minio/minio-01)
	ResourceRequestID string           `gorm:"type:char(36);not null;index" json:"resource_request_id"` // Link to resource request
	ResourceRequest   *ResourceRequest `gorm:"foreignKey:ResourceRequestID" json:"resource_request,omitempty"`
	StorageRepoID     string           `gorm:"type:char(36);not null;index" json:"storage_repo_id"` // Link to storage repository
	StorageRepo       *GitRepository   `gorm:"foreignKey:StorageRepoID" json:"storage_repo,omitempty"`
	ModuleRepoID      *string          `gorm:"type:char(36)" json:"module_repo_id"` // Link to modules repository
	ModuleRepo        *GitRepository   `gorm:"foreignKey:ModuleRepoID" json:"module_repo,omitempty"`
	TerragruntConfig  string           `gorm:"type:text" json:"terragrunt_config"` // Generated terragrunt.hcl content
	TerraformVars     string           `gorm:"type:json" json:"terraform_vars"`    // Variables as JSON
	Status            NodeConfigStatus `gorm:"type:varchar(32);default:'pending'" json:"status"`
	CommitSHA         string           `gorm:"type:varchar(64)" json:"commit_sha"`         // Current commit SHA in storage repo
	PendingCommitSHA  string           `gorm:"type:varchar(64)" json:"pending_commit_sha"` // Pending commit SHA (before approval)
	TerraformState    string           `gorm:"type:text" json:"terraform_state"`           // Terraform state (if stored locally)
	ProvisionLog      string           `gorm:"type:text" json:"provision_log"`             // Provisioning log
	ErrorMessage      string           `gorm:"type:text" json:"error_message"`             // Error message if failed
	ProvisionedAt     *time.Time       `json:"provisioned_at"`
	DestroyedAt       *time.Time       `json:"destroyed_at"`
}

// TableName returns the table name for NodeConfig.
func (NodeConfig) TableName() string {
	return "node_configs"
}

// TerraformRegistry represents a Terraform provider registry.
type TerraformRegistry struct {
	BaseModel
	Name        string `gorm:"type:varchar(128);not null" json:"name"`
	Endpoint    string `gorm:"type:varchar(512);not null" json:"endpoint"` // e.g., registry.infra.plz.ac
	Username    string `gorm:"type:varchar(256)" json:"-"`                 // Auth username (encrypted)
	Token       string `gorm:"type:text" json:"-"`                         // Auth token (encrypted)
	Description string `gorm:"type:text" json:"description"`
	Status      int8   `gorm:"type:tinyint;default:1;not null" json:"status"` // 0: disabled, 1: active
	IsDefault   bool   `gorm:"default:false" json:"is_default"`
}

// TableName returns the table name for TerraformRegistry.
func (TerraformRegistry) TableName() string {
	return "terraform_registries"
}

// TerraformProvider represents a Terraform provider from a registry.
type TerraformProvider struct {
	BaseModel
	Name        string             `gorm:"type:varchar(128);not null" json:"name"`          // e.g., proxmox, aws, azurerm
	Namespace   string             `gorm:"type:varchar(128)" json:"namespace"`              // e.g., telmate, hashicorp
	Source      string             `gorm:"type:varchar(512)" json:"source"`                 // Full source path
	Version     string             `gorm:"type:varchar(64)" json:"version"`                 // Version constraint
	RegistryID  string             `gorm:"type:char(36);not null;index" json:"registry_id"` // Link to registry
	Registry    *TerraformRegistry `gorm:"foreignKey:RegistryID" json:"registry,omitempty"`
	Description string             `gorm:"type:text" json:"description"`
	Status      int8               `gorm:"type:tinyint;default:1;not null" json:"status"` // 0: disabled, 1: active
}

// TableName returns the table name for TerraformProvider.
func (TerraformProvider) TableName() string {
	return "terraform_providers"
}

// TerraformModule represents a Terraform module source.
type TerraformModule struct {
	BaseModel
	Name        string             `gorm:"type:varchar(128);not null" json:"name"`   // Display name
	Source      string             `gorm:"type:varchar(512);not null" json:"source"` // Git URL or registry path
	Version     string             `gorm:"type:varchar(64)" json:"version"`          // Version/tag/branch
	RegistryID  *string            `gorm:"type:char(36)" json:"registry_id"`         // Link to registry
	Registry    *TerraformRegistry `gorm:"foreignKey:RegistryID" json:"registry,omitempty"`
	ProviderID  *string            `gorm:"type:char(36)" json:"provider_id"` // Link to required provider
	Provider    *TerraformProvider `gorm:"foreignKey:ProviderID" json:"provider,omitempty"`
	Description string             `gorm:"type:text" json:"description"`
	Variables   string             `gorm:"type:json" json:"variables"`                    // Available variables as JSON
	Status      int8               `gorm:"type:tinyint;default:1;not null" json:"status"` // 0: disabled, 1: active
}

// TableName returns the table name for TerraformModule.
func (TerraformModule) TableName() string {
	return "terraform_modules"
}

// Region represents a geographical region.
type Region struct {
	BaseModel
	Name        string `gorm:"type:varchar(64);uniqueIndex;not null" json:"name"`
	Code        string `gorm:"type:varchar(32);uniqueIndex;not null" json:"code"` // e.g., cn-north, us-east
	DisplayName string `gorm:"type:varchar(128);not null" json:"display_name"`
	Description string `gorm:"type:text" json:"description"`
	Status      int8   `gorm:"type:tinyint;default:1;not null" json:"status"` // 0: disabled, 1: active
	Zones       []Zone `gorm:"foreignKey:RegionID" json:"zones,omitempty"`
}

// TableName returns the table name for Region.
func (Region) TableName() string {
	return "regions"
}

// Zone represents a zone/cluster within a region.
type Zone struct {
	BaseModel
	Name        string  `gorm:"type:varchar(64);not null" json:"name"`
	Code        string  `gorm:"type:varchar(32);not null;uniqueIndex:idx_zone_code" json:"code"` // e.g., zone-a, cluster-1
	DisplayName string  `gorm:"type:varchar(128);not null" json:"display_name"`
	Description string  `gorm:"type:text" json:"description"`
	RegionID    string  `gorm:"type:char(36);not null;index" json:"region_id"`
	Region      *Region `gorm:"foreignKey:RegionID" json:"region,omitempty"`
	Status      int8    `gorm:"type:tinyint;default:1;not null" json:"status"` // 0: disabled, 1: active
	IsDefault   bool    `gorm:"default:false" json:"is_default"`
}

// TableName returns the table name for Zone.
func (Zone) TableName() string {
	return "zones"
}

// SSHKey represents an SSH public key for VM provisioning.
type SSHKey struct {
	BaseModel
	Name        string `gorm:"type:varchar(128);not null" json:"name"`
	PublicKey   string `gorm:"type:text;not null" json:"public_key"`
	Fingerprint string `gorm:"type:varchar(128);uniqueIndex" json:"fingerprint"`
	Description string `gorm:"type:text" json:"description"`
	CreatedByID string `gorm:"type:char(36);not null" json:"created_by_id"`
	CreatedBy   *User  `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`
	IsDefault   bool   `gorm:"default:false" json:"is_default"`
	Status      int8   `gorm:"type:tinyint;default:1;not null" json:"status"` // 0: disabled, 1: active
}

// TableName returns the table name for SSHKey.
func (SSHKey) TableName() string {
	return "ssh_keys"
}

// IPPool represents an IP address pool for IPAM.
type IPPool struct {
	BaseModel
	Name        string `gorm:"type:varchar(128);not null" json:"name"`
	CIDR        string `gorm:"type:varchar(64);not null" json:"cidr"`       // e.g., "10.31.0.0/24"
	Gateway     string `gorm:"type:varchar(45);not null" json:"gateway"`    // e.g., "10.31.0.254"
	DNS         string `gorm:"type:varchar(256)" json:"dns"`                // Comma-separated DNS servers
	VLANTag     int    `gorm:"default:-1" json:"vlan_tag"`                  // -1 means no VLAN
	StartIP     string `gorm:"type:varchar(45);not null" json:"start_ip"`   // Start of usable range
	EndIP       string `gorm:"type:varchar(45);not null" json:"end_ip"`     // End of usable range
	ZoneID      string `gorm:"type:char(36);not null;index" json:"zone_id"` // Associated zone
	Zone        *Zone  `gorm:"foreignKey:ZoneID" json:"zone,omitempty"`
	NetworkType string `gorm:"type:varchar(32);default:'vmbr0'" json:"network_type"` // Bridge name
	Description string `gorm:"type:text" json:"description"`
	Status      int8   `gorm:"type:tinyint;default:1;not null" json:"status"` // 0: disabled, 1: active
}

// TableName returns the table name for IPPool.
func (IPPool) TableName() string {
	return "ip_pools"
}

// IPAllocationStatus represents the status of an IP allocation.
type IPAllocationStatus string

// IPAllocationStatus constants.
const (
	IPStatusAvailable IPAllocationStatus = "available"
	IPStatusReserved  IPAllocationStatus = "reserved"
	IPStatusAllocated IPAllocationStatus = "allocated"
)

// IPAllocation represents an allocated IP address from a pool.
type IPAllocation struct {
	BaseModel
	IPPoolID    string             `gorm:"type:char(36);not null;index" json:"ip_pool_id"`
	IPPool      *IPPool            `gorm:"foreignKey:IPPoolID" json:"ip_pool,omitempty"`
	IPAddress   string             `gorm:"type:varchar(45);not null;uniqueIndex" json:"ip_address"`
	Hostname    string             `gorm:"type:varchar(256)" json:"hostname"`
	ResourceID  *string            `gorm:"type:char(36);index" json:"resource_id"` // Reference to the resource using this IP
	Status      IPAllocationStatus `gorm:"type:varchar(32);default:'available'" json:"status"`
	AllocatedAt *time.Time         `json:"allocated_at"`
	Description string             `gorm:"type:text" json:"description"`
}

// TableName returns the table name for IPAllocation.
func (IPAllocation) TableName() string {
	return "ip_allocations"
}

// VMTemplate represents a VM template for cloud-init provisioning.
type VMTemplate struct {
	BaseModel
	Name         string  `gorm:"type:varchar(128);not null" json:"name"`                      // Display name
	TemplateName string  `gorm:"type:varchar(256);not null;uniqueIndex" json:"template_name"` // Actual template name in provider
	Provider     string  `gorm:"type:varchar(32);not null" json:"provider"`                   // pve, vmware, etc.
	OSType       string  `gorm:"type:varchar(64);not null" json:"os_type"`                    // linux, windows
	OSFamily     string  `gorm:"type:varchar(64)" json:"os_family"`                           // ubuntu, debian, centos, windows-server
	OSVersion    string  `gorm:"type:varchar(32)" json:"os_version"`                          // 22.04, 12, etc.
	ZoneID       *string `gorm:"type:char(36);index" json:"zone_id"`                          // Optional zone restriction
	Zone         *Zone   `gorm:"foreignKey:ZoneID" json:"zone,omitempty"`
	MinCPU       int     `gorm:"default:1" json:"min_cpu"`
	MinMemoryMB  int     `gorm:"default:1024" json:"min_memory_mb"`
	MinDiskGB    int     `gorm:"default:10" json:"min_disk_gb"`
	DefaultUser  string  `gorm:"type:varchar(64);default:'root'" json:"default_user"` // Default SSH user
	CloudInit    bool    `gorm:"default:true" json:"cloud_init"`                      // Supports cloud-init
	Description  string  `gorm:"type:text" json:"description"`
	Status       int8    `gorm:"type:tinyint;default:1;not null" json:"status"` // 0: disabled, 1: active
}

// TableName returns the table name for VMTemplate.
func (VMTemplate) TableName() string {
	return "vm_templates"
}
