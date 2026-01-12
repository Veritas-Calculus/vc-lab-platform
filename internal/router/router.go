// Package router provides HTTP router setup.
package router

import (
	"github.com/Veritas-Calculus/vc-lab-platform/internal/config"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/handler"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/middleware"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/notification"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/repository"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/service"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/terraform"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// New creates a new configured Gin router with all dependencies.
func New(db *gorm.DB, logger *zap.Logger, cfg *config.Config) *gin.Engine {
	// Set Gin mode based on environment
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	resourceRepo := repository.NewResourceRepository(db)
	resourceRequestRepo := repository.NewResourceRequestRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	providerRepo := repository.NewProviderRepository(db)
	credentialRepo := repository.NewCredentialRepository(db)
	regionRepo := repository.NewRegionRepository(db)
	zoneRepo := repository.NewZoneRepository(db)
	tfRegistryRepo := repository.NewTerraformRegistryRepository(db)
	tfProviderRepo := repository.NewTerraformProviderRepository(db)
	tfModuleRepo := repository.NewTerraformModuleRepository(db)
	gitRepoRepo := repository.NewGitRepoRepository(db)
	nodeConfigRepo := repository.NewNodeConfigRepository(db)
	sshKeyRepo := repository.NewSSHKeyRepository(db)
	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipAllocationRepo := repository.NewIPAllocationRepository(db)
	vmTemplateRepo := repository.NewVMTemplateRepository(db)

	// Initialize Terraform executor
	terraformExecutor := terraform.NewExecutor(logger)

	// Initialize notification service
	notificationService := notification.NewService(db, logger)

	// Initialize services
	authService := service.NewAuthService(userRepo, cfg)
	userService := service.NewUserService(userRepo, roleRepo, logger)
	resourceService := service.NewResourceService(resourceRepo, resourceRequestRepo, terraformExecutor, notificationService, logger)
	roleService := service.NewRoleService(roleRepo, logger)
	settingsService := service.NewSettingsService(providerRepo, credentialRepo, logger)
	infraService := service.NewInfraService(regionRepo, zoneRepo, tfRegistryRepo, tfProviderRepo, tfModuleRepo, logger)
	gitService := service.NewGitService(gitRepoRepo, nodeConfigRepo, tfModuleRepo, logger)
	sshKeyService := service.NewSSHKeyService(sshKeyRepo, logger)
	ipamService := service.NewIPAMService(ipPoolRepo, ipAllocationRepo, logger)
	vmTemplateService := service.NewVMTemplateService(vmTemplateRepo, logger)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService, logger)
	userHandler := handler.NewUserHandler(userService, logger)
	resourceHandler := handler.NewResourceHandler(resourceService, logger)
	roleHandler := handler.NewRoleHandler(roleService, logger)
	healthHandler := handler.NewHealthHandler(db, logger)
	settingsHandler := handler.NewSettingsHandler(settingsService, logger)
	gitHandler := handler.NewGitHandler(gitService, logger)
	infraHandler := handler.NewInfraHandler(infraService, logger)
	sshKeyHandler := handler.NewSSHKeyHandler(sshKeyService, logger)
	ipamHandler := handler.NewIPAMHandler(ipamService, logger)
	vmTemplateHandler := handler.NewVMTemplateHandler(vmTemplateService, logger)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authService, logger)
	auditMiddleware := middleware.NewAuditMiddleware(auditRepo, logger)

	// Setup router
	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.SecureHeaders())

	// Health check endpoints (no auth required)
	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Ready)

	// API v1 group
	v1 := router.Group("/api/v1")

	// Public routes
	auth := v1.Group("/auth")
	auth.POST("/login", authHandler.Login)
	auth.POST("/refresh", authHandler.RefreshToken)

	// Protected routes
	protected := v1.Group("")
	protected.Use(authMiddleware.Authenticate())
	protected.Use(auditMiddleware.Audit())

	// Auth routes
	protected.POST("/auth/logout", authHandler.Logout)

	// User routes
	users := protected.Group("/users")
	users.GET("", userHandler.List)
	users.POST("", userHandler.Create)
	users.GET("/me", userHandler.GetCurrentUser)
	users.PUT("/me", userHandler.UpdateCurrentUser)
	users.PUT("/me/password", userHandler.ChangePassword)
	users.GET("/:id", userHandler.GetByID)
	users.PUT("/:id", userHandler.Update)
	users.DELETE("/:id", userHandler.Delete)

	// Role routes
	roles := protected.Group("/roles")
	roles.GET("", roleHandler.List)
	roles.POST("", roleHandler.Create)
	roles.GET("/:id", roleHandler.GetByID)
	roles.PUT("/:id", roleHandler.Update)
	roles.DELETE("/:id", roleHandler.Delete)

	// Resource routes
	resources := protected.Group("/resources")
	resources.GET("", resourceHandler.List)
	resources.POST("", resourceHandler.Create)
	resources.GET("/:id", resourceHandler.GetByID)
	resources.PUT("/:id", resourceHandler.Update)
	resources.DELETE("/:id", resourceHandler.Delete)

	// Resource request routes
	requests := protected.Group("/resource-requests")
	requests.GET("", resourceHandler.ListRequests)
	requests.POST("", resourceHandler.CreateRequest)
	requests.GET("/:id", resourceHandler.GetRequest)
	requests.POST("/:id/approve", resourceHandler.ApproveRequest)
	requests.POST("/:id/reject", resourceHandler.RejectRequest)
	requests.POST("/:id/retry", resourceHandler.RetryRequest)
	requests.DELETE("/:id", resourceHandler.DeleteRequest)

	// Settings routes - providers
	providers := protected.Group("/settings/providers")
	providers.GET("", settingsHandler.ListProviders)
	providers.POST("", settingsHandler.CreateProvider)
	providers.POST("/test-connection", settingsHandler.TestProviderConnection)
	providers.GET("/:id", settingsHandler.GetProvider)
	providers.PUT("/:id", settingsHandler.UpdateProvider)
	providers.DELETE("/:id", settingsHandler.DeleteProvider)

	// Settings routes - credentials
	credentials := protected.Group("/settings/credentials")
	credentials.GET("", settingsHandler.ListCredentials)
	credentials.POST("", settingsHandler.CreateCredential)
	credentials.POST("/test-connection", settingsHandler.TestCredentialConnection)
	credentials.GET("/:id", settingsHandler.GetCredential)
	credentials.PUT("/:id", settingsHandler.UpdateCredential)
	credentials.DELETE("/:id", settingsHandler.DeleteCredential)

	// Infrastructure routes - regions
	regions := protected.Group("/infra/regions")
	regions.GET("", infraHandler.ListRegions)
	regions.POST("", infraHandler.CreateRegion)
	regions.GET("/:id", infraHandler.GetRegion)
	regions.PUT("/:id", infraHandler.UpdateRegion)
	regions.DELETE("/:id", infraHandler.DeleteRegion)

	// Infrastructure routes - zones
	zones := protected.Group("/infra/zones")
	zones.GET("", infraHandler.ListZones)
	zones.POST("", infraHandler.CreateZone)
	zones.GET("/:id", infraHandler.GetZone)
	zones.PUT("/:id", infraHandler.UpdateZone)
	zones.DELETE("/:id", infraHandler.DeleteZone)

	// Infrastructure routes - terraform registries
	registries := protected.Group("/infra/registries")
	registries.GET("", infraHandler.ListRegistries)
	registries.POST("", infraHandler.CreateRegistry)
	registries.GET("/:id", infraHandler.GetRegistry)
	registries.PUT("/:id", infraHandler.UpdateRegistry)
	registries.DELETE("/:id", infraHandler.DeleteRegistry)

	// Infrastructure routes - terraform providers
	tfProviders := protected.Group("/infra/providers")
	tfProviders.GET("", infraHandler.ListProviders)
	tfProviders.POST("", infraHandler.CreateProvider)
	tfProviders.GET("/:id", infraHandler.GetProvider)
	tfProviders.PUT("/:id", infraHandler.UpdateProvider)
	tfProviders.DELETE("/:id", infraHandler.DeleteProvider)

	// Infrastructure routes - terraform modules
	modules := protected.Group("/infra/modules")
	modules.GET("", infraHandler.ListModules)
	modules.POST("", infraHandler.CreateModule)
	modules.GET("/:id", infraHandler.GetModule)
	modules.PUT("/:id", infraHandler.UpdateModule)
	modules.DELETE("/:id", infraHandler.DeleteModule)

	// Git repository routes
	gitRepos := protected.Group("/git/repositories")
	gitRepos.GET("", gitHandler.ListRepositories)
	gitRepos.POST("", gitHandler.CreateRepository)
	gitRepos.POST("/test-connection", gitHandler.TestConnectionDirect)
	gitRepos.GET("/:id", gitHandler.GetRepository)
	gitRepos.PUT("/:id", gitHandler.UpdateRepository)
	gitRepos.DELETE("/:id", gitHandler.DeleteRepository)
	gitRepos.POST("/:id/test", gitHandler.TestConnection)

	// Git modules routes (scan Terraform modules from git repository)
	gitModules := protected.Group("/git/modules")
	gitModules.GET("", gitHandler.ListModulesFromGit)
	gitModules.POST("/sync", gitHandler.SyncModulesFromGit)

	// Node config routes
	nodeConfigs := protected.Group("/git/node-configs")
	nodeConfigs.GET("", gitHandler.ListNodeConfigs)
	nodeConfigs.GET("/:id", gitHandler.GetNodeConfig)
	nodeConfigs.GET("/by-request/:request_id", gitHandler.GetNodeConfigByRequest)
	nodeConfigs.POST("/:id/commit", gitHandler.CommitNodeConfig)

	// SSH Key routes
	sshKeys := protected.Group("/settings/ssh-keys")
	sshKeys.GET("", sshKeyHandler.ListSSHKeys)
	sshKeys.POST("", sshKeyHandler.CreateSSHKey)
	sshKeys.GET("/default", sshKeyHandler.GetDefaultSSHKey)
	sshKeys.GET("/:id", sshKeyHandler.GetSSHKey)
	sshKeys.PUT("/:id", sshKeyHandler.UpdateSSHKey)
	sshKeys.DELETE("/:id", sshKeyHandler.DeleteSSHKey)
	sshKeys.POST("/:id/set-default", sshKeyHandler.SetDefaultSSHKey)

	// IPAM routes - IP pools
	ipPools := protected.Group("/ipam/pools")
	ipPools.GET("", ipamHandler.ListIPPools)
	ipPools.POST("", ipamHandler.CreateIPPool)
	ipPools.GET("/:id", ipamHandler.GetIPPool)
	ipPools.PUT("/:id", ipamHandler.UpdateIPPool)
	ipPools.DELETE("/:id", ipamHandler.DeleteIPPool)
	ipPools.GET("/:id/allocations", ipamHandler.ListIPAllocations)

	// IPAM routes - IP allocations
	ipAllocations := protected.Group("/ipam/allocations")
	ipAllocations.POST("", ipamHandler.AllocateIP)
	ipAllocations.DELETE("/:id", ipamHandler.ReleaseIP)
	ipAllocations.GET("/resource/:resource_id", ipamHandler.GetAllocationsByResource)

	// VM Template routes
	vmTemplates := protected.Group("/infra/vm-templates")
	vmTemplates.GET("", vmTemplateHandler.ListVMTemplates)
	vmTemplates.POST("", vmTemplateHandler.CreateVMTemplate)
	vmTemplates.GET("/:id", vmTemplateHandler.GetVMTemplate)
	vmTemplates.PUT("/:id", vmTemplateHandler.UpdateVMTemplate)
	vmTemplates.DELETE("/:id", vmTemplateHandler.DeleteVMTemplate)

	return router
}
