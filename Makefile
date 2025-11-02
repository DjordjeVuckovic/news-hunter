# Build variables
APP_NAME := news-hunter
CMD_DIR := ./cmd
BIN_DIR := ./bin
PKG := github.com/DjordjeVuckovic/news-hunter
MIGRATIONS_PATH := ./db/migrations
DB_CONN := "postgresql://news_user:news_password@localhost:54320/news_db?sslmode=disable"

# Go variables
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# Build commands
.PHONY: build build-all clean test fmt vet schema-gen

migrate-up:
	@echo "Running database migrations up..."
	@migrate -path $(MIGRATIONS_PATH) -database $(DB_CONN) up
# Build all commands
build-all: build-data-import build-schemagen

build-data-import:
	@echo "Building data-import..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/data-import $(CMD_DIR)/data_import

build-news-search:
	@echo "Building news-searcher..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/news-search $(CMD_DIR)/news_search

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
dc-up:
	@echo "Starting database..."
	@docker-compose up -d

dc-down:
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
	@ENV_PATHS="cmd/data_import/pg.env" ./$(BIN_DIR)/data-import

run-schemagen: build-schemagen
	@echo "Running schema generator..."
	@./$(BIN_DIR)/schemagen -output=api

run-search: build-news-search
	@echo "Running news search service..."
	@ENV_PATHS="cmd/news_search/.env" ./$(BIN_DIR)/news-search

run-search-pg: build-news-search
	@echo "Running news search service..."
	@ENV_PATHS="cmd/news_search/pg.env" ./$(BIN_DIR)/news-search

run-search-es: build-news-search
	@echo "Running news search service..."
	@ENV_PATHS="cmd/news_search/es.env" ./$(BIN_DIR)/news-search
# Development workflow
dev: fmt vet test schema-gen build-all

.DEFAULT_GOAL := build-all