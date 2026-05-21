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
.PHONY: build build-all clean test fmt vet schema-gen build-bench bench-validate bench-run bench-pool bench-judge-lexical bench-judge-cli bench-judge-api bench-qrels bench-show-spec bench-show-pool bench-show-judgments

migrate-up:
	@echo "Running database migrations up..."
	@migrate -path $(MIGRATIONS_PATH) -database $(DB_CONN) up
# Build all commands
build-all: build-ds-ingest build-news-api build-schemagen build-bench

build-ds-ingest:
	@echo "Building ds_ds-ingest..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/ds-ingest $(CMD_DIR)/ds_ingest

build-news-api:
	@echo "Building news-api..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/news-api $(CMD_DIR)/news_api

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

# Run data ds-ingest with default config
run-ds-ingest-pg: build-ds-ingest
	@echo "Running data ds-ingest..."
	@ENV_PATHS="cmd/ds_ingest/.env,cmd/ds_ingest/pg.env" ./$(BIN_DIR)/ds-ingest

# Run data ds-ingest with default config
run-ds-ingest-es: build-ds-ingest
	@echo "Running data ds-ingest..."
	@ENV_PATHS="cmd/ds_ingest/.env,cmd/ds_ingest/es.env" ./$(BIN_DIR)/ds-ingest

run-api: build-news-api
	@echo "Running news search service..."
	@ENV_PATHS="cmd/news_api/.env" ./$(BIN_DIR)/news-api

run-api-pg: build-news-api
	@echo "Running news search service..."
	@ENV_PATHS="cmd/news_api/.env,cmd/news_api/pg.env" ./$(BIN_DIR)/news-api

run-api-es: build-news-api
	@echo "Running news search service..."
	@ENV_PATHS="cmd/news_api/.env,cmd/news_api/es.env" ./$(BIN_DIR)/news-api
# Benchmark commands
build-bench:
	@echo "Building bench..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/bench $(CMD_DIR)/bench

TRACK ?= fts_quality

bench-validate: build-bench
	@./$(BIN_DIR)/bench validate $(TRACK)

bench-run: build-bench
	@./$(BIN_DIR)/bench run $(TRACK)

bench-pool: build-bench
	@./$(BIN_DIR)/bench pool $(TRACK)

bench-judge-lexical: build-bench
	@./$(BIN_DIR)/bench judge $(TRACK) --strategy lexical

bench-judge-cli: build-bench
	@./$(BIN_DIR)/bench judge $(TRACK) --strategy claude-cli --resume

bench-judge-api: build-bench
	@./$(BIN_DIR)/bench judge $(TRACK) --strategy claude-api --resume

bench-show-spec: build-bench
	@./$(BIN_DIR)/bench show spec $(TRACK)

bench-show-pool: build-bench
	@./$(BIN_DIR)/bench show pool $(TRACK)

bench-show-judgments: build-bench
	@./$(BIN_DIR)/bench show judgments $(TRACK) --strategy lexical

bench-qrels: build-bench
	@./$(BIN_DIR)/bench qrels $(TRACK) --strategy lexical

# Development workflow
dev: fmt vet test build-all

.DEFAULT_GOAL := build-all