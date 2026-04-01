# Makefile
.PHONY: setup lint test test-race fuzz run-cli

MODULES = libs/go-core apps/backend
GOBIN = $(shell go env GOPATH)/bin

setup:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

lint:
	@for mod in $(MODULES); do \
		echo "Linting $$mod..."; \
		cd $$mod && $(GOBIN)/golangci-lint run ./... || exit 1; \
		cd - > /dev/null; \
	done

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
	cd apps/backend && go test -fuzz=Fuzz -fuzztime=30s ./internal/expense/domain/...

run-cli:
	cd apps/backend && go run cmd/cli/main.go -file=../../test_expenses.csv

all: lint test-race
