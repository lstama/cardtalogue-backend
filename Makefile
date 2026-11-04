.PHONY: build build-migrate run start test clean dev-setup docker-up docker-down db-up db-down migrate-up migrate-down migrate-rollback migrate-create migrate-version migrate-force help fmt lint install fill-sample delete-db delete-db-docker image image-arm64 sync-contract generate verify-contract

# Variables
BINARY_NAME=cardtalogue-backend
DATABASE_DRIVER ?= sqlite
DATABASE_URL ?= file:./data.db
MIGRATIONS_PATH=migrations
IMAGE_REPO ?= registry.tama.id:30500/cardtalogue-backend
IMAGE_TAG ?= arm64
IMAGE_NAME ?= $(IMAGE_REPO):$(IMAGE_TAG)
CARD_SCAN_API_URL=http://172.17.0.3:5000/api/recognize # check the real local IP first, use localhost if the app is run locally from outside container 

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	@echo "Building..."
	@go build -o bin/$(BINARY_NAME) ./cmd/api

image: ## Build the container image locally with buildx
	@echo "Building image $(IMAGE_NAME)..."
	docker buildx build --platform linux/arm64 --load -t $(IMAGE_NAME) .

image-arm64: image ## Alias for the arm64 image build

build-migrate: ## Build the migration helper
	@echo "Building migration helper..."
	@go build -o bin/migrate ./cmd/migrate

run: ## Run the application (schema is auto-synced on startup)
	@echo "Starting application..."
	@DATABASE_URL="$(DATABASE_URL)" DATABASE_DRIVER="$(DATABASE_DRIVER)" CARD_SCAN_API_URL="$(CARD_SCAN_API_URL)"  go run ./cmd/api

start: docker-up run ## Start the database and application

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

# Database schema sync
migrate-up: ## Sync schema with GORM (this happens automatically on startup)
	@echo "Schema sync is handled automatically by GORM on startup."
	@echo "Run 'make run' or 'make start' to launch the API."

migrate-down: ## Rollback is not used with GORM auto-migration
	@echo "Rollback is not used with GORM auto-migration."

migrate-rollback: ## Rollback specific number of migrations (not used with GORM auto-migration)
	@echo "Rollback is not used with GORM auto-migration."

migrate-create: ## Create a new model change instead of a SQL migration file
	@echo "GORM auto-migration uses your Go models. Update the model struct and rerun the API."

migrate-version: ## Show current migration version (not used with GORM auto-migration)
	@echo "GORM auto-migration does not use versioned migration files."

migrate-force: ## Force migration version (not used with GORM auto-migration)
	@echo "GORM auto-migration does not use versioned migration files."

# Development
dev-setup: ## Setup development environment
	@echo "Setting up development environment..."
	@go mod download
	@go mod tidy

docker-up: ## Start database with docker compose
	@echo "Starting database..."
	@docker compose up -d postgres

docker-down: ## Stop docker services
	@echo "Stopping docker services..."
	@docker compose down

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
	@go install ./cmd/api

fill-sample:
	@PGPASSWORD=password psql -h 172.17.0.1 -U user -d dbname -a -f sample-data/initialize.sql

delete-db:
	@podman ps --format "{{.ID}}" --filter "name=postgres" | xargs podman rm -f
	@podman volume rm -f cardtalogue-backend_postgres_data

delete-db-docker:
	@docker ps --format "{{.ID}}" --filter "name=postgres" | xargs docker rm -f
	@docker volume rm -f cardtalogue-backend_postgres_data

db-up: docker-up

db-down: docker-down

up-db: docker-up

# --- API contract -----------------------------------------------------------
# The backend consumes the pinned contract from the sibling cardtalogue-api-contract
# repo. CONTRACT_PATH overrides the source (defaults to the sibling repo).
# Local dev override: CONTRACT_PATH=../cardtalogue-api-contract make generate
CONTRACT_VERSION := $(shell cat CONTRACT_VERSION 2>/dev/null)
CONTRACT_PATH ?= ../cardtalogue-api-contract
BACKEND_CONTRACT := $(CONTRACT_PATH)/contracts/backend.yaml
GEN_DIR := internal/api/gen

# oapi-codegen lives in GOPATH/bin; put it on PATH so `make generate` works
# without manual PATH setup.
export PATH := $(shell go env GOPATH)/bin:$(PATH)

sync-contract: ## Copy the pinned contract into internal/api/gen/backend.yaml
	@if [ -n "$(CONTRACT_PATH)" ] && [ -f "$(BACKEND_CONTRACT)" ]; then \
		cp "$(BACKEND_CONTRACT)" "$(GEN_DIR)/backend.yaml"; \
		echo "  synced $(GEN_DIR)/backend.yaml from $(CONTRACT_PATH) (v$(CONTRACT_VERSION))"; \
	else \
		echo "CONTRACT_PATH=$(CONTRACT_PATH) not found; fetch tag v$(CONTRACT_VERSION) manually"; \
		exit 1; \
	fi

generate: ## Regenerate Go types + Gin/strict server interface from the contract
	@if [ ! -f "$(GEN_DIR)/backend.yaml" ]; then \
		echo "backend.yaml missing; run 'make sync-contract' first"; exit 1; \
	fi
	oapi-codegen -o $(GEN_DIR)/openapi.go -config $(GEN_DIR)/oapi-codegen.yaml $(GEN_DIR)/backend.yaml

verify-contract: ## Regenerate and confirm the committed code is in sync
	@if [ ! -f "$(GEN_DIR)/backend.yaml" ]; then \
		echo "backend.yaml missing; run 'make sync-contract' first"; exit 1; \
	fi
	@cp $(GEN_DIR)/openapi.go /tmp/openapi.verify.$$
	oapi-codegen -o $(GEN_DIR)/openapi.go -config $(GEN_DIR)/oapi-codegen.yaml $(GEN_DIR)/backend.yaml
	@if diff -q /tmp/openapi.verify.$$ $(GEN_DIR)/openapi.go >/dev/null 2>&1; then \
		echo "verify-contract: OK (generated code in sync with contract)"; \
	else \
		echo "verify-contract: FAIL — committed openapi.go is out of sync; run 'make generate'"; \
		diff /tmp/openapi.verify.$$ $(GEN_DIR)/openapi.go | head; \
		exit 1; \
	fi
	rm -f /tmp/openapi.verify.$$
