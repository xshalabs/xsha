.PHONY: help build run test clean docker-build docker-run dev deps fmt lint vet tidy

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
DOCKER_IMAGE=$(APP_NAME):latest

# Go related commands
build: ## Build application
	@echo "Building $(APP_NAME)..."
	cd $(BACKEND_DIR) && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Build completed: $(BACKEND_DIR)/$(BUILD_DIR)/$(BINARY_NAME)"

build-local: ## Build local version
	@echo "Building local version..."
	cd $(BACKEND_DIR) && go build -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Build completed: $(BACKEND_DIR)/$(BUILD_DIR)/$(BINARY_NAME)"

build-embedded: ## Build application with embedded frontend (Linux) - works with existing static files
	@echo "Building $(APP_NAME) with embedded frontend..."
	cd $(BACKEND_DIR) && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $(BUILD_DIR)/$(BINARY_NAME)-embedded .
	@echo "Embedded build completed: $(BACKEND_DIR)/$(BUILD_DIR)/$(BINARY_NAME)-embedded"
	@echo "This binary contains all frontend assets and can run standalone!"

build-embedded-local: ## Build local version with embedded frontend - works with existing static files
	@echo "Building local version with embedded frontend..."
	cd $(BACKEND_DIR) && go build -o $(BUILD_DIR)/$(BINARY_NAME)-embedded .
	@echo "Embedded local build completed: $(BACKEND_DIR)/$(BUILD_DIR)/$(BINARY_NAME)-embedded"
	@echo "This binary contains all frontend assets and can run standalone!"

build-embedded-production: ## Build production version with embedded frontend - works with existing static files
	@echo "Building production version with embedded frontend..."
	cd $(BACKEND_DIR) && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o $(BUILD_DIR)/$(BINARY_NAME)-embedded .
	@echo "Embedded production build completed: $(BACKEND_DIR)/$(BUILD_DIR)/$(BINARY_NAME)-embedded"
	@echo "This binary contains all frontend assets and can run standalone!"

build-embedded-with-frontend: frontend-build build-embedded-local ## Build frontend then embedded backend in one command
	@echo "Complete embedded build with fresh frontend completed!"
	@echo "Binary: $(BACKEND_DIR)/$(BUILD_DIR)/$(BINARY_NAME)-embedded"

run: ## Run application (development mode)
	@echo "Starting $(APP_NAME)..."
	cd $(BACKEND_DIR) && go run main.go

test: ## Run tests
	@echo "Running tests..."
	cd $(BACKEND_DIR) && go test -v ./...

test-coverage: ## Run tests and generate coverage report
	@echo "Running test coverage analysis..."
	cd $(BACKEND_DIR) && go test -v -coverprofile=coverage.out ./...
	cd $(BACKEND_DIR) && go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: $(BACKEND_DIR)/coverage.html"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	cd $(BACKEND_DIR) && go mod download
	cd $(BACKEND_DIR) && go mod verify

tidy: ## Tidy dependencies
	@echo "Tidying dependencies..."
	cd $(BACKEND_DIR) && go mod tidy

fmt: ## Format code
	@echo "Formatting code..."
	cd $(BACKEND_DIR) && go fmt ./...

vet: ## Run go vet
	@echo "Running vet check..."
	cd $(BACKEND_DIR) && go vet ./...

clean: ## Clean build files
	@echo "Cleaning build files..."
	rm -rf $(BACKEND_DIR)/$(BUILD_DIR)
	rm -rf $(BACKEND_DIR)/coverage.out
	rm -rf $(BACKEND_DIR)/coverage.html
	rm -rf $(BACKEND_DIR)/app.db
	
# Database related
db-reset: ## Reset database (delete SQLite file)
	@echo "Resetting database..."
	rm -f $(BACKEND_DIR)/app.db

# Production deployment
deploy-build: ## Build production version
	@echo "Building production version..."
	cd $(BACKEND_DIR) && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o $(BUILD_DIR)/$(BINARY_NAME) .

# Health check
health: ## Check application health status
	@echo "Checking application health status..."
	curl -f http://localhost:8080/api/health || echo "Application is not running or health check failed"

# Complete check suite
check: fmt vet lint test ## Run all checks (format, vet, lint, test)

# Frontend related commands
frontend-deps: ## Install frontend dependencies
	@echo "Installing frontend dependencies..."
	cd $(FRONTEND_DIR) && pnpm install

frontend-build: frontend-deps ## Build frontend application
	@echo "Building frontend application..."
	cd $(FRONTEND_DIR) && pnpm run build
	@echo "Frontend build completed! Static files are in $(BACKEND_DIR)/static/"

frontend-dev: ## Start frontend development server
	@echo "Starting frontend development server..."
	cd $(FRONTEND_DIR) && pnpm run dev

frontend-clean: ## Clean frontend build files
	@echo "Cleaning frontend build files..."
	rm -rf $(BACKEND_DIR)/static
	rm -rf $(FRONTEND_DIR)/dist
	rm -rf $(FRONTEND_DIR)/node_modules/.cache

# Full-stack build commands
build-fullstack: frontend-build build ## Build complete application (frontend + backend)
	@echo "Full-stack build completed!"
	@echo "Static files: $(BACKEND_DIR)/static/"
	@echo "Backend binary: $(BACKEND_DIR)/$(BUILD_DIR)/$(BINARY_NAME)"

build-fullstack-local: frontend-build build-local ## Build complete application for local development
	@echo "Local full-stack build completed!"

deploy-fullstack: frontend-build deploy-build ## Build production version with frontend
	@echo "Production full-stack build completed!"

deploy-embedded: build-embedded-production ## Build production version with embedded frontend
	@echo "Production embedded build completed!"
	@echo "Single binary ready for deployment: $(BACKEND_DIR)/$(BUILD_DIR)/$(BINARY_NAME)"

run-fullstack: frontend-build ## Run full-stack application
	@echo "Starting full-stack application..."
	cd $(BACKEND_DIR) && go run main.go

# Docker related commands
docker-build: ## Build Docker image with full-stack (current platform)
	@echo "Building Docker image with full-stack..."
	docker build -t $(DOCKER_IMAGE) .
	@echo "Docker image built successfully: $(DOCKER_IMAGE)"

docker-build-multiplatform: ## Build Docker image with multi-platform support
	@echo "Building multi-platform Docker image..."
	docker buildx create --use --name multiplatform-builder 2>/dev/null || true
	docker buildx build --platform linux/amd64,linux/arm64 -t $(DOCKER_IMAGE) --push .
	@echo "Multi-platform Docker image built and pushed successfully!"

docker-build-arm64: ## Build Docker image for ARM64 platform
	@echo "Building Docker image for ARM64..."
	docker buildx create --use --name arm64-builder 2>/dev/null || true
	docker buildx build --platform linux/arm64 -t $(DOCKER_IMAGE)-arm64 --load .
	@echo "ARM64 Docker image built successfully: $(DOCKER_IMAGE)-arm64"

docker-build-amd64: ## Build Docker image for AMD64 platform
	@echo "Building Docker image for AMD64..."
	docker buildx create --use --name amd64-builder 2>/dev/null || true
	docker buildx build --platform linux/amd64 -t $(DOCKER_IMAGE)-amd64 --load .
	@echo "AMD64 Docker image built successfully: $(DOCKER_IMAGE)-amd64"

docker-run: ## Run application in Docker container
	@echo "Running application in Docker container..."
	docker run -p 8080:8080 --rm --name xsha-app $(DOCKER_IMAGE)

docker-stop: ## Stop Docker container
	@echo "Stopping Docker container..."
	docker stop xsha-app || true

docker-clean: ## Clean Docker images and containers
	@echo "Cleaning Docker images and containers..."
	docker stop xsha-app || true
	docker rm xsha-app || true
	docker rmi $(DOCKER_IMAGE) || true

docker-compose-up: ## Start all services with docker-compose
	@echo "Starting services with docker-compose..."
	docker-compose up -d

docker-compose-down: ## Stop all services with docker-compose
	@echo "Stopping services with docker-compose..."
	docker-compose down

docker-compose-logs: ## View docker-compose logs
	@echo "Viewing docker-compose logs..."
	docker-compose logs -f

docker-compose-rebuild: ## Rebuild and restart docker-compose services
	@echo "Rebuilding and restarting services..."
	docker-compose down
	docker-compose build --no-cache
	docker-compose up -d

docker-setup-buildx: ## Setup Docker buildx for multi-platform builds
	@echo "Setting up Docker buildx..."
	docker buildx create --use --name multiplatform-builder --driver docker-container 2>/dev/null || true
	docker buildx inspect --bootstrap
	@echo "Docker buildx setup completed!"

# Complete development workflow
setup: deps frontend-deps install-tools ## Setup development environment
	@echo "Development environment setup completed!"
	@echo "Run 'make dev' to start development server"
	@echo "Run 'make frontend-dev' to start frontend development server"
	@echo "Run 'make build-fullstack' to build complete application"
	@echo "Run 'make build-embedded-local' to build standalone binary with embedded frontend"
	@echo "Run 'make build-embedded-production' to build production standalone binary"
	@echo "Run 'make docker-build' to build Docker image (current platform)"
	@echo "Run 'make docker-build-multiplatform' to build multi-platform image"
	@echo "Run 'make docker-setup-buildx' to setup multi-platform builds"
	@echo "Run 'make docker-compose-up' to start with docker-compose"
	@echo "Run 'make help' to view all available commands" 