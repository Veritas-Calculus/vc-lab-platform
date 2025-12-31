// Package router provides HTTP router setup.
package router

import (
	"time"

	"github.com/Veritas-Calculus/vc-lab-platform/internal/config"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/handler"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/middleware"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/repository"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// New creates a new configured Gin router with all dependencies.
func New(db *gorm.DB, rdb *redis.Client, logger *zap.Logger, cfg *config.Config) *gin.Engine {
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

	// Initialize services
	authService := service.NewAuthService(userRepo, rdb, cfg)
	userService := service.NewUserService(userRepo, roleRepo, logger)
	resourceService := service.NewResourceService(resourceRepo, resourceRequestRepo, logger)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService, logger)
	userHandler := handler.NewUserHandler(userService, logger)
	resourceHandler := handler.NewResourceHandler(resourceService, logger)
	healthHandler := handler.NewHealthHandler(db, logger)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authService, logger)
	rateLimiter := middleware.NewRateLimiter(rdb, logger, 100, time.Minute)
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

	// Apply rate limiting
	if rateLimiter != nil {
		v1.Use(rateLimiter.Limit())
	}

	// Public routes
	auth := v1.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
	}

	// Protected routes
	protected := v1.Group("")
	protected.Use(authMiddleware.Authenticate())
	protected.Use(auditMiddleware.Audit())
	{
		// Auth routes
		protected.POST("/auth/logout", authHandler.Logout)

		// User routes
		users := protected.Group("/users")
		{
			users.GET("", userHandler.List)
			users.POST("", userHandler.Create)
			users.GET("/me", userHandler.GetCurrentUser)
			users.PUT("/me", userHandler.UpdateCurrentUser)
			users.PUT("/me/password", userHandler.ChangePassword)
			users.GET("/:id", userHandler.GetByID)
			users.PUT("/:id", userHandler.Update)
			users.DELETE("/:id", userHandler.Delete)
		}

		// Resource routes
		resources := protected.Group("/resources")
		{
			resources.GET("", resourceHandler.List)
			resources.POST("", resourceHandler.Create)
			resources.GET("/:id", resourceHandler.GetByID)
			resources.PUT("/:id", resourceHandler.Update)
			resources.DELETE("/:id", resourceHandler.Delete)
		}

		// Resource request routes
		requests := protected.Group("/resource-requests")
		{
			requests.GET("", resourceHandler.ListRequests)
			requests.POST("", resourceHandler.CreateRequest)
			requests.GET("/:id", resourceHandler.GetRequest)
			requests.POST("/:id/approve", resourceHandler.ApproveRequest)
			requests.POST("/:id/reject", resourceHandler.RejectRequest)
		}
	}

	return router
}
