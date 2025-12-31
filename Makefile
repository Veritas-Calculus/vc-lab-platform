# ========================================
# VC Lab Platform Makefile
# ========================================

.PHONY: all build run test lint clean help setup dev

# Variables
BINARY_NAME=vc-lab-server
MAIN_PATH=./cmd/server
BUILD_DIR=./bin
GO=go
NPM=npm

# Build flags
LDFLAGS=-ldflags "-s -w"

# Default target
all: lint test build

# ========================================
# Development
# ========================================

## setup: Install all dependencies and setup hooks
setup:
	@echo "🔧 Setting up development environment..."
	@chmod +x scripts/setup-hooks.sh
	@./scripts/setup-hooks.sh

## dev: Run the development server with hot reload
dev:
	@echo "🚀 Starting development server..."
	@air -c .air.toml || $(GO) run $(MAIN_PATH)

## dev-frontend: Run frontend development server
dev-frontend:
	@echo "🚀 Starting frontend development server..."
	@cd web && $(NPM) run dev

# ========================================
# Build
# ========================================

## build: Build the Go backend
build:
	@echo "🔨 Building backend..."
	@mkdir -p $(BUILD_DIR)
	@$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

## build-frontend: Build the frontend
build-frontend:
	@echo "🔨 Building frontend..."
	@cd web && $(NPM) run build
	@echo "✅ Frontend build complete"

## build-all: Build both backend and frontend
build-all: build build-frontend

## run: Run the server
run: build
	@echo "🚀 Running server..."
	@$(BUILD_DIR)/$(BINARY_NAME)

# ========================================
# Testing
# ========================================

## test: Run all tests
test: test-backend test-frontend

## test-backend: Run backend tests
test-backend:
	@echo "🧪 Running backend tests..."
	@$(GO) test ./... -v -count=1 -race -coverprofile=coverage.out
	@$(GO) tool cover -func=coverage.out | tail -1

## test-frontend: Run frontend tests
test-frontend:
	@echo "🧪 Running frontend tests..."
	@cd web && $(NPM) test -- --run

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "📊 Generating coverage report..."
	@$(GO) test ./... -coverprofile=coverage.out -covermode=atomic
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: coverage.html"

## test-security: Run security tests
test-security:
	@echo "🔒 Running security tests..."
	@$(GO) test ./internal/security/... -v

# ========================================
# Linting
# ========================================

## lint: Run all linters
lint: lint-backend lint-frontend

## lint-backend: Run Go linters
lint-backend:
	@echo "🔍 Running Go linters..."
	@golangci-lint run --timeout 5m ./...

## lint-frontend: Run frontend linters
lint-frontend:
	@echo "🔍 Running frontend linters..."
	@cd web && $(NPM) run lint

## lint-fix: Fix linting issues
lint-fix:
	@echo "🔧 Fixing Go linting issues..."
	@golangci-lint run --fix ./...
	@echo "🔧 Fixing frontend linting issues..."
	@cd web && $(NPM) run lint -- --fix

## fmt: Format code
fmt:
	@echo "📝 Formatting Go code..."
	@$(GO) fmt ./...
	@goimports -w .
	@echo "📝 Formatting frontend code..."
	@cd web && $(NPM) run format || true

# ========================================
# Security
# ========================================

## security: Run security scanners
security:
	@echo "🔒 Running security scanners..."
	@gosec -exclude-dir=vendor ./...

## audit: Audit dependencies
audit:
	@echo "🔍 Auditing Go dependencies..."
	@$(GO) list -json -m all | nancy sleuth || true
	@echo "🔍 Auditing frontend dependencies..."
	@cd web && $(NPM) audit || true

# ========================================
# Database
# ========================================

## migrate: Run database migrations
migrate:
	@echo "🗃️  Running database migrations..."
	@$(GO) run $(MAIN_PATH) migrate

## migrate-down: Rollback last migration
migrate-down:
	@echo "🗃️  Rolling back migration..."
	@$(GO) run $(MAIN_PATH) migrate down

# ========================================
# Docker
# ========================================

## docker-build: Build Docker image
docker-build:
	@echo "🐳 Building Docker image..."
	@docker build -t vc-lab-platform:latest .

## docker-run: Run Docker container
docker-run:
	@echo "🐳 Running Docker container..."
	@docker run -p 8080:8080 vc-lab-platform:latest

## docker-compose-up: Start all services with docker-compose
docker-compose-up:
	@echo "🐳 Starting services..."
	@docker-compose up -d

## docker-compose-down: Stop all services
docker-compose-down:
	@echo "🐳 Stopping services..."
	@docker-compose down

# ========================================
# Cleanup
# ========================================

## clean: Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -rf coverage.out coverage.html
	@rm -rf web/dist web/node_modules/.cache
	@$(GO) clean -cache -testcache
	@echo "✅ Clean complete"

# ========================================
# Dependencies
# ========================================

## deps: Download Go dependencies
deps:
	@echo "📦 Downloading Go dependencies..."
	@$(GO) mod download
	@$(GO) mod tidy

## deps-frontend: Install frontend dependencies
deps-frontend:
	@echo "📦 Installing frontend dependencies..."
	@cd web && $(NPM) install

## deps-update: Update all dependencies
deps-update:
	@echo "📦 Updating Go dependencies..."
	@$(GO) get -u ./...
	@$(GO) mod tidy
	@echo "📦 Updating frontend dependencies..."
	@cd web && $(NPM) update

# ========================================
# Pre-commit
# ========================================

## pre-commit: Run pre-commit hooks
pre-commit:
	@pre-commit run --all-files

## pre-commit-install: Install pre-commit hooks
pre-commit-install:
	@pre-commit install
	@pre-commit install --hook-type commit-msg

# ========================================
# Help
# ========================================

## help: Show this help message
help:
	@echo "VC Lab Platform - Available Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
