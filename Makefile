.PHONY: run build clean test deps hot-reload help migrate-up migrate-down migrate-status migrate-create migrate-force

# Default target
all: deps build

# Run the server
run:
	@echo "Starting time tracker server..."
	go run main.go

# Build the application
build:
	@echo "Building time tracker..."
	go build -o bin/time-tracker .

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod tidy
	go mod download

# Run with hot reload (requires air)
hot-reload:
	@echo "Starting server with hot reload..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not installed. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Falling back to normal run..."; \
		make run; \
	fi

# Development setup
dev-setup: deps
	@echo "Setting up development environment..."
	@if ! command -v air > /dev/null; then \
		echo "Installing air for hot reload..."; \
		go install github.com/cosmtrek/air@latest; \
	fi

# Start development server
dev: dev-setup hot-reload

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Migration commands
# Load environment variables from .env file
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Install golang-migrate if not present
install-migrate:
	@which migrate > /dev/null || (echo "Installing golang-migrate..." && go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest)

# Use full path to migrate if not in PATH
MIGRATE_CMD := $(shell which migrate 2>/dev/null || echo "$(shell go env GOPATH)/bin/migrate")

# Run all pending migrations
migrate-up: install-migrate
	@echo "Running migrations up..."
	$(MIGRATE_CMD) -path migrations -database "$(DATABASE_URL)" up

# Rollback the last migration
migrate-down: install-migrate
	@echo "Rolling back last migration..."
	$(MIGRATE_CMD) -path migrations -database "$(DATABASE_URL)" down 1

# Rollback all migrations
migrate-down-all: install-migrate
	@echo "Rolling back all migrations..."
	$(MIGRATE_CMD) -path migrations -database "$(DATABASE_URL)" down

# Show current migration status
migrate-status: install-migrate
	@echo "Current migration status:"
	$(MIGRATE_CMD) -path migrations -database "$(DATABASE_URL)" version

# Create a new migration file
# Usage: make migrate-create NAME=add_user_table
migrate-create: install-migrate
	@if [ -z "$(NAME)" ]; then echo "Usage: make migrate-create NAME=migration_name"; exit 1; fi
	@echo "Creating migration: $(NAME)"
	$(MIGRATE_CMD) create -ext sql -dir migrations -seq $(NAME)

# Force migration to specific version (use carefully!)
# Usage: make migrate-force VERSION=1
migrate-force: install-migrate
	@if [ -z "$(VERSION)" ]; then echo "Usage: make migrate-force VERSION=version_number"; exit 1; fi
	@echo "Forcing migration to version $(VERSION)..."
	$(MIGRATE_CMD) -path migrations -database "$(DATABASE_URL)" force $(VERSION)

# Migrate to specific version
# Usage: make migrate-to VERSION=1
migrate-to: install-migrate
	@if [ -z "$(VERSION)" ]; then echo "Usage: make migrate-to VERSION=version_number"; exit 1; fi
	@echo "Migrating to version $(VERSION)..."
	$(MIGRATE_CMD) -path migrations -database "$(DATABASE_URL)" goto $(VERSION)

# Development helpers
db-setup: install-migrate migrate-up
	@echo "Database setup complete!"

db-reset: migrate-down-all migrate-up
	@echo "Database reset complete!"

# Show help
help:
	@echo "Available targets:"
	@echo ""
	@echo "Development:"
	@echo "  run         - Run the server"
	@echo "  build       - Build the application"
	@echo "  clean       - Clean build artifacts"
	@echo "  test        - Run tests"
	@echo "  deps        - Download dependencies"
	@echo "  dev         - Start development server with hot reload"
	@echo "  hot-reload  - Run with hot reload (requires air)"
	@echo "  dev-setup   - Setup development environment"
	@echo "  fmt         - Format code"
	@echo "  lint        - Lint code (requires golangci-lint)"
	@echo ""
	@echo "Database Migrations:"
	@echo "  migrate-up              - Run all pending migrations"
	@echo "  migrate-down            - Rollback the last migration"
	@echo "  migrate-down-all        - Rollback all migrations"
	@echo "  migrate-status          - Show current migration version"
	@echo "  migrate-create NAME=... - Create new migration file"
	@echo "  migrate-force VERSION=. - Force to specific version (dangerous!)"
	@echo "  migrate-to VERSION=...  - Migrate to specific version"
	@echo "  db-setup                - Setup database with migrations"
	@echo "  db-reset                - Reset and rerun all migrations"
	@echo ""
	@echo "  help        - Show this help message"