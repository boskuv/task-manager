.PHONY: help build run test test-integration test-cover test-cover-core test-cover-html migrate migrate-down migrate-create clean lint fmt

APP_NAME    := task-manager
BINARY      := bin/$(APP_NAME)
MAIN        := ./cmd/api
PACKAGES    := ./cmd/... ./internal/...
COVER_CORE_PKGS := ./internal/usecase/... ./internal/repository/...
COVERAGE_MIN    := 85
MIGRATIONS  := migrations
GOOSE       ?= go tool goose
DATABASE_DSN ?= $(shell grep -E '^\s+dsn:' configs/config.yaml 2>/dev/null | head -1 | awk '{print $$2}' | tr -d '"')

help: ## Show available targets
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

build: ## Build API binary
	@mkdir -p bin
	go build -o $(BINARY) $(MAIN)

run: ## Run API locally
	go run $(MAIN)

test: ## Run unit tests
	go test $(PACKAGES) -v -race -count=1

test-integration: ## Run integration tests (requires Docker)
	go test $(PACKAGES) -tags=integration -v -race -count=1

test-cover: ## Run all packages with coverage report
	go test $(PACKAGES) -race -coverprofile=coverage.out -covermode=atomic -count=1
	go tool cover -func=coverage.out

test-cover-core: ## Run usecase+repo tests with coverage (requires Docker, min $(COVERAGE_MIN)%)
	go test $(COVER_CORE_PKGS) -tags=integration -coverprofile=coverage.out -covermode=atomic -count=1
	@echo "--- Coverage summary (usecase + repository) ---"
	go tool cover -func=coverage.out
	@total=$$(go tool cover -func=coverage.out | awk '/^total:/ {gsub("%","",$$3); print $$3}'); \
	echo "Total: $$total% (minimum $(COVERAGE_MIN)%)"; \
	awk -v total="$$total" -v min="$(COVERAGE_MIN)" 'BEGIN { if (total+0 < min+0) { print "coverage below minimum"; exit 1 } }'

test-cover-html: test-cover-core ## Generate HTML coverage report for usecase+repo
	go tool cover -html=coverage.out -o coverage.html
	@echo "Wrote coverage.html"

migrate: ## Apply database migrations
	$(GOOSE) -dir $(MIGRATIONS) mysql "$(DATABASE_DSN)" up

migrate-down: ## Rollback last migration
	$(GOOSE) -dir $(MIGRATIONS) mysql "$(DATABASE_DSN)" down

migrate-create: ## Create new migration (usage: make migrate-create NAME=create_users)
	$(GOOSE) -dir $(MIGRATIONS) create $(NAME) sql

lint: ## Run golangci-lint
	golangci-lint run $(PACKAGES)

fmt: ## Format Go code
	go fmt ./...
	gofmt -s -w .

clean: ## Remove build artifacts
	rm -rf bin/ coverage.out coverage.html coverage.txt
