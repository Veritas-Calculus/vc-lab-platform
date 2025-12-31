#!/bin/bash
# ========================================
# Git Hooks Setup Script
# ========================================
# This script sets up pre-commit hooks for the vc-lab-platform project.
# Run this script once after cloning the repository.

set -e

echo "🔧 Setting up Git hooks for vc-lab-platform..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if we're in a git repository
if [ ! -d ".git" ]; then
    echo -e "${RED}Error: Not a git repository. Please run this script from the project root.${NC}"
    exit 1
fi

# Check for required tools
echo "📦 Checking required tools..."

# Check Go
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed. Please install Go 1.21+${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Go $(go version | awk '{print $3}')${NC}"

# Check Node.js
if ! command -v node &> /dev/null; then
    echo -e "${RED}Error: Node.js is not installed. Please install Node.js 18+${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Node.js $(node --version)${NC}"

# Check npm
if ! command -v npm &> /dev/null; then
    echo -e "${RED}Error: npm is not installed.${NC}"
    exit 1
fi
echo -e "${GREEN}✓ npm $(npm --version)${NC}"

# Install pre-commit if not installed
if ! command -v pre-commit &> /dev/null; then
    echo -e "${YELLOW}Installing pre-commit...${NC}"
    if command -v pip3 &> /dev/null; then
        pip3 install pre-commit
    elif command -v pip &> /dev/null; then
        pip install pre-commit
    elif command -v brew &> /dev/null; then
        brew install pre-commit
    else
        echo -e "${RED}Error: Cannot install pre-commit. Please install pip or brew first.${NC}"
        exit 1
    fi
fi
echo -e "${GREEN}✓ pre-commit $(pre-commit --version)${NC}"

# Install golangci-lint if not installed
if ! command -v golangci-lint &> /dev/null; then
    echo -e "${YELLOW}Installing golangci-lint...${NC}"
    if command -v brew &> /dev/null; then
        brew install golangci-lint
    else
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2
    fi
fi
echo -e "${GREEN}✓ golangci-lint $(golangci-lint --version | head -1)${NC}"

# Install gosec if not installed
if ! command -v gosec &> /dev/null; then
    echo -e "${YELLOW}Installing gosec...${NC}"
    go install github.com/securego/gosec/v2/cmd/gosec@latest
fi
echo -e "${GREEN}✓ gosec installed${NC}"

# Install Go dependencies
echo ""
echo "📥 Installing Go dependencies..."
go mod download
go mod tidy
echo -e "${GREEN}✓ Go dependencies installed${NC}"

# Install frontend dependencies
echo ""
echo "📥 Installing frontend dependencies..."
cd web
npm install
cd ..
echo -e "${GREEN}✓ Frontend dependencies installed${NC}"

# Install pre-commit hooks
echo ""
echo "🔗 Installing pre-commit hooks..."
pre-commit install
pre-commit install --hook-type commit-msg
echo -e "${GREEN}✓ Pre-commit hooks installed${NC}"

# Run pre-commit on all files (optional, can take a while)
echo ""
read -p "Run pre-commit on all files now? This may take a few minutes. (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "🔍 Running pre-commit on all files..."
    pre-commit run --all-files || true
fi

echo ""
echo -e "${GREEN}✅ Setup complete!${NC}"
echo ""
echo "Usage:"
echo "  - Commits will automatically be checked by pre-commit hooks"
echo "  - Run 'pre-commit run --all-files' to check all files manually"
echo "  - Run 'make lint' to run linters manually"
echo "  - Run 'make test' to run all tests"
echo ""
