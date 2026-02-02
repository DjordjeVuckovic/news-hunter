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
.PHONY: build build-all clean test fmt vet schema-gen build-benchmark run-benchmark run-benchmark-all

migrate-up:
	@echo "Running database migrations up..."
	@migrate -path $(MIGRATIONS_PATH) -database $(DB_CONN) up
# Build all commands
build-all: build-ds-ingest build-schemagen build-benchmark

build-ds-ingest:
	@echo "Building ds_ingest..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/ds-ingest $(CMD_DIR)/ds_ingest

build-news-api:
	@echo "Building news-api..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/news-api $(CMD_DIR)/news_search

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

run-schemagen: build-schemagen
	@echo "Running schema generator..."
	@./$(BIN_DIR)/schemagen -output=api

# Run data import with default config
run-import-pg: build-ds-ingest
	@echo "Running data import..."
	@ENV_PATHS="cmd/ds_ingest/pg.env" ./$(BIN_DIR)/ds-ingest

# Run data import with default config
run-import-es: build-ds-ingest
	@echo "Running data import..."
	@ENV_PATHS="cmd/ds_ingest/es.env" ./$(BIN_DIR)/ds-ingest

run-search: build-news-api
	@echo "Running news search service..."
	@ENV_PATHS="cmd/news_search/.env" ./$(BIN_DIR)/news-api

run-search-pg: build-news-api
	@echo "Running news search service..."
	@ENV_PATHS="cmd/news_search/pg.env" ./$(BIN_DIR)/news-api

run-search-es: build-news-api
	@echo "Running news search service..."
	@ENV_PATHS="cmd/news_search/es.env" ./$(BIN_DIR)/news-api
# Benchmark commands
build-benchmark:
	@echo "Building benchmark..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/benchmark $(CMD_DIR)/benchmark

run-benchmark: build-benchmark
	@echo "Running FTS quality benchmark..."
	@./$(BIN_DIR)/benchmark --pg $(DB_CONN) --suite configs/benchmark/fts_quality_v1.yaml

run-benchmark-all: build-benchmark
	@echo "Running FTS quality benchmark (PG + ES)..."
	@./$(BIN_DIR)/benchmark --pg $(DB_CONN) --es-addresses "http://localhost:9200" --suite configs/benchmark/fts_quality_v1.yaml

# Development workflow
dev: fmt vet test build-all

.DEFAULT_GOAL := build-all