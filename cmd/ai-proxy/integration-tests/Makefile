# AI Proxy Integration Tests Makefile

.PHONY: help test test-chat test-embeddings test-audio test-images test-responses test-verbose test-concurrent test-quick clean setup

# Default target
.DEFAULT_GOAL := help

## Show help information
help:
	@echo "AI Proxy Integration Tests"
	@echo ""
	@echo "Main targets:"
	@echo "  test                   - Run all tests (chat completions + responses)"
	@echo "  setup                  - Set up test environment (copy .env.example to .env)"
	@echo "  check                  - Check compilation and syntax"
	@echo "  clean                  - Clean test cache"
	@echo ""
	@echo "Individual API tests:"
	@echo "  test-chat-completions - Run chat completions tests (non-stream + stream)"
	@echo "  test-chat-completions-nonstream - Run chat completions non-stream tests"
	@echo "  test-chat-completions-stream - Run chat completions stream tests"
	@echo "  test-audio             - Run audio transcriptions tests"
	@echo "  test-images            - Run image generation tests"
	@echo "  test-embeddings        - Run embeddings tests"
	@echo "  test-responses         - Run responses API tests (non-stream + stream)"
	@echo "  test-responses-nonstream - Run responses non-stream tests"
	@echo "  test-responses-stream  - Run responses stream tests"
	@echo "  test-openai-v1-models  - Run OpenAI v1/models tests"
	@echo "  test-chat              - Alias for test-chat-completions"
	@echo ""
	@echo "Utility targets:"
	@echo "  help                   - Show this help message"
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

## Run all tests
test: test-chat-completions test-audio test-images test-embeddings test-responses test-openai-v1-models
	@echo "✓ All tests completed!"


## ===== Individual API type test targets =====

## Run chat completions non-stream tests
test-chat-completions-nonstream: setup mod-init
	@echo "Running Chat Completions Non-stream tests..."
	go test -v -count=1 -timeout 300s -failfast -run "^TestChatCompletionsNonStream$$" ./...

## Run chat completions stream tests
test-chat-completions-stream: setup mod-init
	@echo "Running Chat Completions Stream tests..."
	go test -v -count=1 -timeout 300s -failfast -run "^TestChatCompletionsStreaming$$" ./...

## Run chat completions tests (non-stream + stream)
test-chat-completions: test-chat-completions-nonstream test-chat-completions-stream
	@echo "✓ All chat completions tests completed!"

## Run audio transcription tests
test-audio: setup mod-init
	@echo "Running Audio Transcriptions tests..."
	@echo "Note: Test audio files will be automatically generated if missing"
	go test -v -count=1 -timeout 300s -failfast -run "^TestAudioTranscriptions" ./...
	@echo "✓ All audio transcriptions tests completed!"

## Run image generation tests
test-images: setup mod-init
	@echo "Running Image Generation tests..."
	go test -v -count=1 -timeout 300s -failfast -run "^TestImageGenerations" ./...
	@echo "✓ All image generation tests completed!"

## Run embeddings tests
test-embeddings: setup mod-init
	@echo "Running Embeddings tests..."
	go test -v -count=1 -timeout 300s -failfast -run "^TestEmbeddings" ./...
	@echo "✓ All embeddings tests completed!"

## Run responses non-stream tests
test-responses-nonstream: setup mod-init
	@echo "Running Responses Non-stream tests..."
	go test -v -count=1 -timeout 300s -failfast -run "^TestResponsesNonStream$$" ./...

## Run responses stream tests
test-responses-stream: setup mod-init
	@echo "Running Responses Stream tests..."
	go test -v -count=1 -timeout 300s -failfast -run "^TestResponsesStreaming$$" ./...

## Run responses tests (non-stream + stream)
test-responses: test-responses-nonstream test-responses-stream
	@echo "✓ All responses tests completed!"

## Run model list tests
test-openai-v1-models: setup mod-init
	@echo "Running OpenAI v1/models tests..."
	go test -v -count=1 -timeout 300s -failfast -run "^TestV1Models" ./...
	@echo "✓ All OpenAI v1/models tests completed!"

## ===== Utility test targets =====

## Clean test cache
clean:
	go clean -testcache

## Run specific test
test-specific: setup
	@if [ -z "$(TEST)" ]; then \
		echo "Usage: make test-specific TEST=TestName"; \
		exit 1; \
	fi
	go test -v -count=1 -timeout 300s -failfast -run $(TEST) ./...

## Run tests and generate coverage report
test-coverage: setup
	go test -v -count=1 -timeout 300s -failfast -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## Run tests continuously (auto-rerun on file changes)
test-watch: setup
	@echo "Watching for file changes... (Press Ctrl+C to stop)"
	@while true; do \
		find . -name "*.go" | entr -d make test-quick; \
	done

## Run stress tests
test-stress: setup
	@echo "Running stress tests..."
	CONCURRENT_TESTS=true go test -v -count=10 -timeout 300s -failfast ./...

## Check compilation
check: mod-init
	@echo "Checking compilation..."
	go vet ./...
	go build ./...
	go test -v -count=1 -run "^$$" ./...
	@echo "✓ All checks passed!"

## Initialize modules
mod-init:
	go mod tidy

## Update dependencies
mod-update:
	go mod tidy
	go mod download

## Check code format
fmt:
	go fmt ./...

## Check code
vet:
	go vet ./...

## Run full CI checks
ci: fmt vet test

## Show test status
status:
	@echo "Test Environment Status:"
	@echo "  .env file: $(if $(wildcard .env),exists,missing)"
	@echo "  Go version: $(shell go version)"
	@echo "  Working directory: $(shell pwd)"
	@echo "  Available test files:"
	@ls -la *_test.go 2>/dev/null || echo "    No test files found"
	@echo "  Audio test files:"
	@if [ -d "testdata" ]; then \
		ls -la testdata/*.mp3 testdata/*.wav testdata/*.m4a 2>/dev/null || echo "    No audio files found (will be auto-generated)"; \
	else \
		echo "    testdata directory missing (will be auto-created)"; \
	fi
