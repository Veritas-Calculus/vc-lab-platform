// Package logger provides structured logging capabilities for the application.
package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a new zap logger instance.
func New() (*zap.Logger, error) {
	env := os.Getenv("VC_ENV")
	if env == "" {
		env = "development"
	}

	var config zap.Config
	if env == "production" {
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	return config.Build()
}

// NewNop creates a no-op logger for testing.
func NewNop() *zap.Logger {
	return zap.NewNop()
}
