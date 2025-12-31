// Package main provides the entry point for the VC Lab Platform server.
package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Veritas-Calculus/vc-lab-platform/internal/config"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/constants"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/database"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/logger"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/router"
	"go.uber.org/zap"
)

func main() {
	configPath := flag.String("config", "config/config.yaml", "path to config file")
	flag.Parse()

	// Initialize logger
	log, err := logger.New()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer func() {
		if syncErr := log.Sync(); syncErr != nil {
			// Ignore sync errors for stdout/stderr
			_ = syncErr
		}
	}()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Error("failed to load config", zap.Error(err))
		return
	}

	// Initialize database
	db, err := database.New(cfg.Database)
	if err != nil {
		log.Error("failed to connect to database", zap.Error(err))
		return
	}

	// Auto migrate database schemas
	if migrateErr := database.AutoMigrate(db); migrateErr != nil {
		log.Error("failed to migrate database", zap.Error(migrateErr))
		return
	}

	// Initialize Redis
	rdb, err := database.NewRedis(cfg.Redis)
	if err != nil {
		log.Error("failed to connect to redis", zap.Error(err))
		return
	}
	defer func() {
		_ = rdb.Close()
	}()

	// Setup router
	r := router.New(db, rdb, log, cfg)

	// Create HTTP server
	srv := &http.Server{
		Addr:              cfg.Server.Addr,
		Handler:           r,
		ReadTimeout:       time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:      time.Duration(cfg.Server.WriteTimeout) * time.Second,
		ReadHeaderTimeout: constants.ReadHeaderTimeout,
		IdleTimeout:       constants.IdleTimeout,
		MaxHeaderBytes:    constants.MaxHeaderBytes,
	}

	// Start server in goroutine
	go func() {
		log.Info("starting server", zap.String("addr", cfg.Server.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("failed to start server", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), constants.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server forced to shutdown", zap.Error(err))
	}

	log.Info("server exited")
}
