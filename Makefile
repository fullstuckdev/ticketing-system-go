# Ticketing System Makefile

# Variables
APP_NAME=ticketing-system
BINARY_NAME=ticketing-system
MAIN_FILE=main.go
BUILD_DIR=build

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Database
DB_NAME=ticketing_system
DB_USER=root
DB_HOST=localhost
DB_PORT=3306

.PHONY: help build clean test run deps fmt vet swagger docker-build docker-run db-create db-migrate

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

clean: ## Clean build files
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

run: ## Run the application
	@echo "Running $(APP_NAME)..."
	$(GOCMD) run $(MAIN_FILE)

dev: ## Run in development mode with auto-reload (requires air)
	@echo "Running in development mode..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "air not found. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Falling back to normal run..."; \
		make run; \
	fi

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) tidy
	$(GOMOD) download

fmt: ## Format code
	@echo "Formatting code..."
	$(GOFMT) ./...

vet: ## Run go vet
	@echo "Running go vet..."
	$(GOCMD) vet ./...

lint: ## Run golangci-lint (requires golangci-lint)
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

swagger: ## Generate Swagger documentation
	@echo "Generating Swagger docs..."
	@if command -v swag > /dev/null; then \
		swag init -g main.go --output docs; \
	else \
		echo "swag not found. Install with: go install github.com/swaggo/swag/cmd/swag@latest"; \
	fi

db-create: ## Create database
	@echo "Creating database $(DB_NAME)..."
	@mysql -h $(DB_HOST) -P $(DB_PORT) -u $(DB_USER) -p -e "CREATE DATABASE IF NOT EXISTS $(DB_NAME);"
	@echo "Database $(DB_NAME) created or already exists"

db-drop: ## Drop database
	@echo "Dropping database $(DB_NAME)..."
	@mysql -h $(DB_HOST) -P $(DB_PORT) -u $(DB_USER) -p -e "DROP DATABASE IF EXISTS $(DB_NAME);"
	@echo "Database $(DB_NAME) dropped"

db-reset: db-drop db-create ## Reset database (drop and create)
	@echo "Database reset complete"

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):latest .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env $(APP_NAME):latest

docker-compose-up: ## Start services with docker-compose
	@echo "Starting services with docker-compose..."
	docker-compose up -d

docker-compose-down: ## Stop services with docker-compose
	@echo "Stopping services with docker-compose..."
	docker-compose down

install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	@echo "Development tools installed"

setup: deps db-create swagger ## Setup development environment
	@echo "Development environment setup complete"

all: clean deps fmt vet test build ## Run all checks and build

# Development workflow
check: fmt vet lint test ## Run all checks

# Production build
build-prod: ## Build for production
	@echo "Building for production..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) -a -installsuffix cgo -ldflags '-extldflags "-static"' -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "Production build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Help target should be first
.DEFAULT_GOAL := help 