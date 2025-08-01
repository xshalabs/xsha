.PHONY: help clean build-amd64 build-arm64 build-arm build-all build-embedded-amd64 build-embedded-arm64 build-embedded-arm build-embedded-all frontend-build

# Default target
help: ## Show help information
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

# Project configuration
APP_NAME=xsha
BUILD_DIR=./build
BACKEND_DIR=./backend
FRONTEND_DIR=./frontend
BINARY_NAME=$(APP_NAME)

# Build flags for production
BUILD_FLAGS=-ldflags="-w -s"
CGO_FLAGS=CGO_ENABLED=0

clean: ## Clean build files
	@echo "Cleaning build files..."
	rm -rf $(BACKEND_DIR)/$(BUILD_DIR)
	rm -rf $(BACKEND_DIR)/static
	rm -rf $(FRONTEND_DIR)/dist

# Frontend build
frontend-build: ## Build frontend application
	@echo "Building frontend application..."
	cd $(FRONTEND_DIR) && pnpm install && pnpm run build
	@echo "Frontend build completed! Static files are in $(BACKEND_DIR)/static/"

# Single architecture builds
build-amd64: ## Build production version for Linux AMD64
	@echo "Building $(APP_NAME) for Linux AMD64..."
	cd $(BACKEND_DIR) && $(CGO_FLAGS) GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	@echo "Build completed: $(BACKEND_DIR)/$(BUILD_DIR)/$(BINARY_NAME)-linux-amd64"

build-arm64: ## Build production version for Linux ARM64
	@echo "Building $(APP_NAME) for Linux ARM64..."
	cd $(BACKEND_DIR) && $(CGO_FLAGS) GOOS=linux GOARCH=arm64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	@echo "Build completed: $(BACKEND_DIR)/$(BUILD_DIR)/$(BINARY_NAME)-linux-arm64"

build-arm: ## Build production version for Linux ARM
	@echo "Building $(APP_NAME) for Linux ARM..."
	cd $(BACKEND_DIR) && $(CGO_FLAGS) GOOS=linux GOARCH=arm go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm .
	@echo "Build completed: $(BACKEND_DIR)/$(BUILD_DIR)/$(BINARY_NAME)-linux-arm"

build-all: build-amd64 build-arm64 build-arm ## Build for all architectures
	@echo "All architecture builds completed!"
	@ls -la $(BACKEND_DIR)/$(BUILD_DIR)/

# Embedded builds (with frontend)
build-embedded-amd64: frontend-build ## Build production version with embedded frontend for Linux AMD64
	@echo "Building $(APP_NAME) with embedded frontend for Linux AMD64..."
	cd $(BACKEND_DIR) && $(CGO_FLAGS) GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-embedded-linux-amd64 .
	@echo "Embedded build completed: $(BACKEND_DIR)/$(BUILD_DIR)/$(BINARY_NAME)-embedded-linux-amd64"
	@echo "This binary contains all frontend assets and can run standalone!"

build-embedded-arm64: frontend-build ## Build production version with embedded frontend for Linux ARM64
	@echo "Building $(APP_NAME) with embedded frontend for Linux ARM64..."
	cd $(BACKEND_DIR) && $(CGO_FLAGS) GOOS=linux GOARCH=arm64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-embedded-linux-arm64 .
	@echo "Embedded build completed: $(BACKEND_DIR)/$(BUILD_DIR)/$(BINARY_NAME)-embedded-linux-arm64"
	@echo "This binary contains all frontend assets and can run standalone!"

build-embedded-arm: frontend-build ## Build production version with embedded frontend for Linux ARM
	@echo "Building $(APP_NAME) with embedded frontend for Linux ARM..."
	cd $(BACKEND_DIR) && $(CGO_FLAGS) GOOS=linux GOARCH=arm go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-embedded-linux-arm .
	@echo "Embedded build completed: $(BACKEND_DIR)/$(BUILD_DIR)/$(BINARY_NAME)-embedded-linux-arm"
	@echo "This binary contains all frontend assets and can run standalone!"

build-embedded-all: build-embedded-amd64 build-embedded-arm64 build-embedded-arm ## Build embedded version for all architectures
	@echo "All embedded architecture builds completed!"
	@ls -la $(BACKEND_DIR)/$(BUILD_DIR)/

# Production deployment shortcuts
deploy: build-embedded-amd64 ## Build production embedded version for AMD64 (most common)
	@echo "Production deployment binary ready: $(BACKEND_DIR)/$(BUILD_DIR)/$(BINARY_NAME)-embedded-linux-amd64"

deploy-all: build-embedded-all ## Build production embedded versions for all architectures
	@echo "All production deployment binaries ready!"
	@echo "Available binaries:"
	@ls -la $(BACKEND_DIR)/$(BUILD_DIR)/ | grep embedded 