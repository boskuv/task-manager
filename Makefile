.PHONY: help build run test test-cover migrate migrate-down migrate-create clean lint fmt

APP_NAME    := task-manager
BINARY      := bin/$(APP_NAME)
MAIN        := ./cmd/api
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

test: ## Run all tests
	go test ./... -v -race -count=1

test-cover: ## Run tests with coverage report
	go test ./... -race -coverprofile=coverage.out -covermode=atomic
	go tool cover -func=coverage.out

migrate: ## Apply database migrations
	$(GOOSE) -dir $(MIGRATIONS) mysql "$(DATABASE_DSN)" up

migrate-down: ## Rollback last migration
	$(GOOSE) -dir $(MIGRATIONS) mysql "$(DATABASE_DSN)" down

migrate-create: ## Create new migration (usage: make migrate-create NAME=create_users)
	$(GOOSE) -dir $(MIGRATIONS) create $(NAME) sql

lint: ## Run golangci-lint
	golangci-lint run ./...

fmt: ## Format Go code
	go fmt ./...
	gofmt -s -w .

clean: ## Remove build artifacts
	rm -rf bin/ coverage.out coverage.html
