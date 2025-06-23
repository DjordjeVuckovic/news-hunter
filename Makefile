# Build variables
APP_NAME := news-hunter
CMD_DIR := ./cmd
BIN_DIR := ./bin
PKG := github.com/DjordjeVuckovic/news-hunter

# Go variables
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# Build commands
.PHONY: build build-all clean test fmt vet schema-gen

# Build all commands
build-all: build-data-import build-schemagen

build-data-import:
	@echo "Building data-import..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/data-import $(CMD_DIR)/data_import

build-schemagen:
	@echo "Building schema generator..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/schemagen $(CMD_DIR)/schemagen

# Generate schemas from Go structs
schema-gen: build-schemagen
	@echo "Generating schemas..."
	@./$(BIN_DIR)/schemagen -output=api
	@echo "Schemas generated in api/ directory"

# Development commands
test:
	@echo "Running tests..."
	@go test -v ./...

fmt:
	@echo "Formatting code..."
	@go fmt ./...

vet:
	@echo "Vetting code..."
	@go vet ./...

# Database commands
db-up:
	@echo "Starting database..."
	@docker-compose up -d

db-down:
	@echo "Stopping database..."
	@docker-compose down

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BIN_DIR)
	@rm -rf schema/generated

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Run data import with default config
run-import: build-data-import
	@echo "Running data import..."
	@./$(BIN_DIR)/data-import

# Development workflow
dev: fmt vet test schema-gen build-all

.DEFAULT_GOAL := build-all