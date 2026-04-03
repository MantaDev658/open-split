# Makefile
.PHONY: setup lint test test-race fuzz run-cli

# --- Variables ---
GOBIN = $(shell go env GOPATH)/bin
DB_URL = "postgresql://postgres:password@localhost:5432/opensplit?sslmode=disable"
MIGRATE_PATH = apps/backend/internal/core/infrastructure/postgres/migrations
MODULES = libs/shared apps/backend

setup:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

build:
	@echo "Building Open Split CLI..."
	go build -o bin/opensplit-cli ./apps/backend/cmd/cli
	@echo "✅ Binary compiled to bin/opensplit-cli"
	@echo "Building Open Split API..."
	go build -o bin/opensplit-api ./apps/backend/cmd/api
	@echo "✅ Binary compiled to bin/opensplit-api"

lint:
	@for mod in $(MODULES); do \
		echo "Linting $$mod..."; \
		cd $$mod && $(GOBIN)/golangci-lint run ./... || exit 1; \
		cd - > /dev/null; \
	done

check: lint test test-race

# --- Testing ---

# We isolate integration tests so they don't slow down our standard 'make test'
test-integration: db-up migrate-up
	@echo "Running PostgreSQL Integration Tests..."
	TEST_DB_URL=$(DB_URL) go test -v ./apps/backend/internal/core/infrastructure/postgres/...

test:
	@for mod in $(MODULES); do \
		echo "Testing $$mod..."; \
		cd $$mod && go test -v -cover ./... || exit 1; \
		cd - > /dev/null; \
	done

test-race:
	@for mod in $(MODULES); do \
		echo "Race testing $$mod..."; \
		cd $$mod && go test -race ./... || exit 1; \
		cd - > /dev/null; \
	done

fuzz:
	cd apps/backend && go test -fuzz=Fuzz -fuzztime=30s ./internal/core/domain/...

# --- Run Application ---

run-api: db-up
	@echo "Starting Open Split API..."
	go run ./apps/backend/cmd/api/main.go

run-cli:
	cd apps/backend && go run cmd/cli/main.go -file=../../test_expenses.csv

# --- Database & Infrastructure ---

db-up:
	@echo "Starting PostgreSQL..."
	docker-compose up -d
	@echo "Waiting for database to be ready..."
	sleep 2

db-down:
	@echo "Stopping PostgreSQL..."
	docker-compose down

setup-migrate:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

migrate-up:
	@$(GOBIN)/migrate -path $(MIGRATE_PATH) -database $(DB_URL) -verbose up

migrate-down:
	@$(GOBIN)/migrate -path $(MIGRATE_PATH) -database $(DB_URL) -verbose down