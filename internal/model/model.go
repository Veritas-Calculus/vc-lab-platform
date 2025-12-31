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

// User represents a platform user.
type User struct {
	BaseModel
	Username     string  `gorm:"type:varchar(64);uniqueIndex;not null" json:"username"`
	Email        string  `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash string  `gorm:"type:varchar(255);not null" json:"-"`
	DisplayName  string  `gorm:"type:varchar(128)" json:"display_name"`
	Phone        string  `gorm:"type:varchar(20)" json:"phone"`
	Avatar       string  `gorm:"type:varchar(512)" json:"avatar"`
	Status       int8    `gorm:"type:tinyint;default:1;not null" json:"status"` // 0: disabled, 1: active
	LastLoginAt  *time.Time `json:"last_login_at"`
	LastLoginIP  string  `gorm:"type:varchar(45)" json:"last_login_ip"`
	Roles        []Role  `gorm:"many2many:user_roles;" json:"roles,omitempty"`
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
	Name          string    `gorm:"type:varchar(128);not null" json:"name"`
	Type          string    `gorm:"type:varchar(32);not null" json:"type"` // vm, container, bare_metal
	Provider      string    `gorm:"type:varchar(32);not null" json:"provider"` // pve, vmware, openstack
	Status        string    `gorm:"type:varchar(32);not null;default:'pending'" json:"status"` // pending, provisioning, running, stopped, error
	Spec          string    `gorm:"type:json" json:"spec"` // CPU, memory, disk specs as JSON
	IPAddress     string    `gorm:"type:varchar(45)" json:"ip_address"`
	HostName      string    `gorm:"type:varchar(255)" json:"hostname"`
	OwnerID       string    `gorm:"type:char(36);index;not null" json:"owner_id"`
	Owner         *User     `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Environment   string    `gorm:"type:varchar(32);index;not null" json:"environment"` // dev, test, staging, prod
	ExternalID    string    `gorm:"type:varchar(255)" json:"external_id"` // ID in the external provider
	ExpiresAt     *time.Time `json:"expires_at"`
	Tags          string    `gorm:"type:json" json:"tags"` // JSON array of tags
	Description   string    `gorm:"type:text" json:"description"`
}

// TableName returns the table name for Resource.
func (Resource) TableName() string {
	return "resources"
}

// ResourceSpec represents the specification for a resource.
type ResourceSpec struct {
	CPU       int    `json:"cpu"`        // Number of CPU cores
	Memory    int    `json:"memory"`     // Memory in MB
	Disk      int    `json:"disk"`       // Disk size in GB
	DiskType  string `json:"disk_type"`  // ssd, hdd
	OSType    string `json:"os_type"`    // linux, windows
	OSImage   string `json:"os_image"`   // ubuntu-22.04, centos-7, etc.
	Network   string `json:"network"`    // Network configuration
}

// ResourceRequest represents a resource request/application.
type ResourceRequest struct {
	BaseModel
	Title       string     `gorm:"type:varchar(255);not null" json:"title"`
	Description string     `gorm:"type:text" json:"description"`
	Spec        string     `gorm:"type:json;not null" json:"spec"` // Requested spec
	Environment string     `gorm:"type:varchar(32);not null" json:"environment"`
	Provider    string     `gorm:"type:varchar(32);not null" json:"provider"`
	Quantity    int        `gorm:"type:int;default:1;not null" json:"quantity"`
	Status      string     `gorm:"type:varchar(32);not null;default:'pending'" json:"status"` // pending, approved, rejected, provisioning, completed
	RequesterID string     `gorm:"type:char(36);index;not null" json:"requester_id"`
	Requester   *User      `gorm:"foreignKey:RequesterID" json:"requester,omitempty"`
	ApproverID  *string    `gorm:"type:char(36)" json:"approver_id"`
	Approver    *User      `gorm:"foreignKey:ApproverID" json:"approver,omitempty"`
	ApprovedAt  *time.Time `json:"approved_at"`
	Reason      string     `gorm:"type:text" json:"reason"` // Reason for approval/rejection
	ExpiresAt   *time.Time `json:"expires_at"`
}

// TableName returns the table name for ResourceRequest.
func (ResourceRequest) TableName() string {
	return "resource_requests"
}

// AuditLog represents an audit log entry.
type AuditLog struct {
	ID        string    `gorm:"type:char(36);primaryKey" json:"id"`
	UserID    string    `gorm:"type:char(36);index" json:"user_id"`
	Username  string    `gorm:"type:varchar(64)" json:"username"`
	Action    string    `gorm:"type:varchar(64);not null" json:"action"`
	Resource  string    `gorm:"type:varchar(64);not null" json:"resource"`
	ResourceID string   `gorm:"type:varchar(255)" json:"resource_id"`
	Details   string    `gorm:"type:json" json:"details"`
	IPAddress string    `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent string    `gorm:"type:varchar(512)" json:"user_agent"`
	Status    string    `gorm:"type:varchar(32);not null" json:"status"` // success, failure
	CreatedAt time.Time `gorm:"index" json:"created_at"`
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
