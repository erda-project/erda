# MCP Proxy Integration Tests Makefile

.PHONY: help test test-connect test-message test-error test-all clean setup mod-init

# Default target
.DEFAULT_GOAL := help

## Show help information
help:
	@echo "MCP Proxy Integration Tests"
	@echo ""
	@echo "Main targets:"
	@echo "  test                   - Run all tests"
	@echo "  test-all               - Run all tests in sequence"
	@echo "  setup                  - Set up test environment (copy .env.example to .env)"
	@echo "  check                  - Check compilation and syntax"
	@echo "  clean                  - Clean test cache"
	@echo ""
	@echo "Individual API tests:"
	@echo "  test-connect           - Run proxy connect tests (GET and POST)"
	@echo "  test-message           - Run proxy message tests"
	@echo "  test-error             - Run error handling tests"
	@echo ""
	@echo "Utility targets:"
	@echo "  help                   - Show this help message"
	@echo "  mod-init               - Initialize Go modules"
	@echo ""

## Set up test environment
setup:
	@if [ ! -f ".env" ]; then \
		echo "Creating .env from .env.example..."; \
		cp .env.example .env; \
		echo "Please edit .env file with your configuration"; \
	else \
		echo ".env file already exists"; \
	fi

## Initialize Go modules
mod-init:
	@echo "Initializing Go modules..."
	go mod tidy
	go mod download

## Run all tests
test: test-all
	@echo "✓ All tests completed!"

## Run all tests in sequence
test-all: setup mod-init
	@echo "Running all MCP Proxy integration tests..."
	go test -v -count=1 -timeout 300s -failfast ./...
	@echo "✓ All tests completed!"

## ===== Individual API test targets =====

## Run proxy connect tests (GET and POST)
test-connect: setup mod-init
	@echo "Running proxy connect tests..."
	go test -v -count=1 -timeout 300s -failfast -run "^TestProxyConnect" ./...

## Run proxy message tests
test-message: setup mod-init
	@echo "Running proxy message tests..."
	go test -v -count=1 -timeout 300s -failfast -run "^TestProxyMessage" ./...

## Run error handling tests
test-error: setup mod-init
	@echo "Running error handling tests..."
	go test -v -count=1 -timeout 300s -failfast -run "^TestProxy.*Invalid|TestProxy.*Empty|TestProxy.*Timeout" ./...

## Check compilation and syntax
check: mod-init
	@echo "Checking compilation and syntax..."
	go build ./...
	go vet ./...
	@echo "✓ Compilation and syntax check passed!"

## Clean test cache
clean:
	@echo "Cleaning test cache..."
	go clean -testcache
	@echo "✓ Test cache cleaned!"

## Run tests with verbose output
test-verbose: setup mod-init
	@echo "Running tests with verbose output..."
	go test -v -count=1 -timeout 300s ./...

## Run tests with race detection
test-race: setup mod-init
	@echo "Running tests with race detection..."
	go test -v -count=1 -timeout 300s -race ./...

## Run tests with coverage
test-coverage: setup mod-init
	@echo "Running tests with coverage..."
	go test -v -count=1 -timeout 300s -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html" 