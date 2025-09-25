.PHONY: build run test clean migrate migrate-up migrate-down migrate-create migrate-version help

# Variables
BINARY_NAME=cardtalogue-backend
DATABASE_URL=postgres://user:password@localhost:5432/dbname?sslmode=disable
MIGRATIONS_PATH=migrations

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	@echo "Building..."
	@go build -o bin/$(BINARY_NAME) cmd/api/main.go

run: ## Run the application
	@echo "Running application..."
	@go run cmd/api/main.go

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out

clean: ## Clean build files
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out

# Database Migration Commands
migrate-up: ## Run all migrations
	@echo "Running migrations..."
	@go run cmd/migrate/main.go up

migrate-down: ## Rollback last migration
	@echo "Rolling back last migration..."
	@go run cmd/migrate/main.go down 1

migrate-rollback: ## Rollback specific number of migrations (usage: make migrate-rollback STEPS=2)
	@echo "Rolling back $(or $(STEPS),1) migration(s)..."
	@go run cmd/migrate/main.go down $(or $(STEPS),1)

migrate-create: ## Create new migration file (usage: make migrate-create NAME=add_user_status)
	@if [ -z "$(NAME)" ]; then \
		echo "Please provide a name: make migrate-create NAME=your_migration_name"; \
		exit 1; \
	fi
	@echo "Creating migration: $(NAME)"
	@go run cmd/migrate/main.go create $(NAME)

migrate-version: ## Show current migration version
	@echo "Current migration version:"
	@go run cmd/migrate/main.go version

migrate-force: ## Force migration version (usage: make migrate-force VERSION=1)
	@if [ -z "$(VERSION)" ]; then \
		echo "Please provide a version: make migrate-force VERSION=1"; \
		exit 1; \
	fi
	@echo "Forcing migration version to $(VERSION)..."
	@go run cmd/migrate/main.go force $(VERSION)

# Development
dev-setup: ## Setup development environment
	@echo "Setting up development environment..."
	@go mod download
	@go mod tidy

docker-up: ## Start database with docker-compose
	@echo "Starting database..."
	@docker-compose up -d postgres

docker-down: ## Stop docker services
	@echo "Stopping docker services..."
	@docker-compose down

# Linting and formatting
fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run

# Installation
install: ## Install the binary
	@echo "Installing..."
	@go install cmd/api/main.go
