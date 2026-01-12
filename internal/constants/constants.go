// Package constants defines application-wide constants.
package constants

import "time"

// Pagination constants.
const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// HTTP server timeouts.
const (
	ReadHeaderTimeout = 5 * time.Second
	IdleTimeout       = 120 * time.Second
	MaxHeaderBytes    = 1 << 20 // 1MB
	ShutdownTimeout   = 30 * time.Second
)

// Database connection timeouts.
const (
	DBConnectionTimeout = 5 * time.Second
)

// Query parameter constants.
const (
	QueryTrue = "true" // Common query parameter value
)

// Provider type constants.
const (
	ProviderTypePVE       = "pve"
	ProviderTypeVMware    = "vmware"
	ProviderTypeOpenStack = "openstack"
	ProviderTypeAWS       = "aws"
	ProviderTypeAliyun    = "aliyun"
	ProviderTypeGCP       = "gcp"
	ProviderTypeAzure     = "azure"
)

// Security constants.
const (
	MinJWTSecretLength = 32
	MinPasswordLength  = 8
	HTTPStatusErrorMin = 400
	DefaultRateLimit   = 100
	BearerParts        = 2 // "Bearer" + token
)
