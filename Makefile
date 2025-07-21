.PHONY: help build run test clean docker-build docker-run dev deps fmt lint vet tidy

# Default target
help: ## Show help information
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

# Project configuration
APP_NAME=sleep0-backend
BUILD_DIR=./build
BACKEND_DIR=./backend
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

run: ## Run application (development mode)
	@echo "Starting $(APP_NAME)..."
	cd $(BACKEND_DIR) && go run main.go

dev: deps ## Start development environment
	@echo "Starting development environment..."
	cd $(BACKEND_DIR) && air || go run main.go

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

lint: ## Run golint
	@echo "Running lint check..."
	cd $(BACKEND_DIR) && golangci-lint run || echo "Please install golangci-lint: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"

vet: ## Run go vet
	@echo "Running vet check..."
	cd $(BACKEND_DIR) && go vet ./...

clean: ## Clean build files
	@echo "Cleaning build files..."
	rm -rf $(BACKEND_DIR)/$(BUILD_DIR)
	rm -rf $(BACKEND_DIR)/coverage.out
	rm -rf $(BACKEND_DIR)/coverage.html
	rm -rf $(BACKEND_DIR)/app.db

# Docker related commands
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## Run Docker container
	@echo "Starting Docker container..."
	docker run --rm -p 8080:8080 --name $(APP_NAME) $(DOCKER_IMAGE)

docker-run-dev: ## Run development Docker container
	@echo "Starting development Docker container..."
	docker run --rm -p 8080:8080 -v $(PWD)/backend:/app --name $(APP_NAME)-dev $(DOCKER_IMAGE)

docker-compose-up: ## Start docker-compose services
	@echo "Starting docker-compose services..."
	docker-compose up -d

docker-compose-down: ## Stop docker-compose services
	@echo "Stopping docker-compose services..."
	docker-compose down

docker-compose-logs: ## View docker-compose logs
	docker-compose logs -f

docker-clean: ## Clean Docker resources
	@echo "Cleaning Docker resources..."
	docker rmi $(DOCKER_IMAGE) 2>/dev/null || true
	docker system prune -f

# Database related
db-reset: ## Reset database (delete SQLite file)
	@echo "Resetting database..."
	rm -f $(BACKEND_DIR)/app.db

# Install development tools
install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

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

# Complete development workflow
setup: deps install-tools ## Setup development environment
	@echo "Development environment setup completed!"
	@echo "Run 'make dev' to start development server"
	@echo "Run 'make help' to view all available commands" 