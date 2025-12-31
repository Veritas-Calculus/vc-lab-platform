// Package model defines the CMDB models for the application.
package model

import (
	"time"

	"gorm.io/gorm"
)

// CMDBAsset represents an asset in the CMDB.
type CMDBAsset struct {
	BaseModel
	Name           string     `gorm:"size:100;not null" json:"name"`
	AssetType      string     `gorm:"size:50;not null;index" json:"asset_type"`
	SerialNumber   string     `gorm:"size:100;uniqueIndex" json:"serial_number"`
	Manufacturer   string     `gorm:"size:100" json:"manufacturer"`
	Model          string     `gorm:"size:100" json:"model"`
	Location       string     `gorm:"size:200" json:"location"`
	DataCenter     string     `gorm:"size:100" json:"data_center"`
	Rack           string     `gorm:"size:50" json:"rack"`
	RackPosition   int        `gorm:"" json:"rack_position"`
	Status         string     `gorm:"size:20;not null;default:'active'" json:"status"`
	IPAddresses    string     `gorm:"type:text" json:"ip_addresses"`  // JSON array
	MACAddresses   string     `gorm:"type:text" json:"mac_addresses"` // JSON array
	PurchaseDate   *time.Time `gorm:"" json:"purchase_date"`
	WarrantyExpiry *time.Time `gorm:"" json:"warranty_expiry"`
	OwnerID        string     `gorm:"size:36;index" json:"owner_id"`
	Owner          *User      `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Tags           string     `gorm:"type:text" json:"tags"`
	Notes          string     `gorm:"type:text" json:"notes"`
	Metadata       string     `gorm:"type:text" json:"metadata"` // JSON object
}

// TableName specifies the table name for CMDBAsset.
func (CMDBAsset) TableName() string {
	return "cmdb_assets"
}

// BeforeCreate hook for CMDBAsset.
func (a *CMDBAsset) BeforeCreate(tx *gorm.DB) error {
	return a.BaseModel.BeforeCreate(tx)
}

// Service represents a service running on infrastructure.
type Service struct {
	BaseModel
	Name          string `gorm:"size:100;not null;uniqueIndex" json:"name"`
	DisplayName   string `gorm:"size:200" json:"display_name"`
	Description   string `gorm:"type:text" json:"description"`
	ServiceType   string `gorm:"size:50;not null" json:"service_type"`
	Status        string `gorm:"size:20;not null;default:'active'" json:"status"`
	Tier          string `gorm:"size:20" json:"tier"` // e.g., tier1, tier2, tier3
	OwnerID       string `gorm:"size:36;index" json:"owner_id"`
	Owner         *User  `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	TeamEmail     string `gorm:"size:100" json:"team_email"`
	SlackChannel  string `gorm:"size:100" json:"slack_channel"`
	Repository    string `gorm:"size:200" json:"repository"`
	Documentation string `gorm:"size:500" json:"documentation"`
	Dependencies  string `gorm:"type:text" json:"dependencies"` // JSON array of service IDs
	Metadata      string `gorm:"type:text" json:"metadata"`     // JSON object
}

// TableName specifies the table name for Service.
func (Service) TableName() string {
	return "services"
}

// BeforeCreate hook for Service.
func (s *Service) BeforeCreate(tx *gorm.DB) error {
	return s.BaseModel.BeforeCreate(tx)
}

// ServiceInstance represents an instance of a service.
type ServiceInstance struct {
	BaseModel
	ServiceID   string     `gorm:"size:36;not null;index" json:"service_id"`
	Service     *Service   `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
	ResourceID  string     `gorm:"size:36;not null;index" json:"resource_id"`
	Resource    *Resource  `gorm:"foreignKey:ResourceID" json:"resource,omitempty"`
	Environment string     `gorm:"size:20;not null" json:"environment"`
	Version     string     `gorm:"size:50" json:"version"`
	Port        int        `gorm:"" json:"port"`
	Status      string     `gorm:"size:20;not null;default:'running'" json:"status"`
	HealthCheck string     `gorm:"size:200" json:"health_check"`
	LastHealthy *time.Time `gorm:"" json:"last_healthy"`
	Metadata    string     `gorm:"type:text" json:"metadata"`
}

// TableName specifies the table name for ServiceInstance.
func (ServiceInstance) TableName() string {
	return "service_instances"
}

// BeforeCreate hook for ServiceInstance.
func (si *ServiceInstance) BeforeCreate(tx *gorm.DB) error {
	return si.BaseModel.BeforeCreate(tx)
}
