.PHONY: help test lint fmt vet coverage clean install-tools

# Default target
help:
	@echo "Available targets:"
	@echo "  make test          - Run tests"
	@echo "  make test-race     - Run tests with race detector"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo "  make lint          - Run linters"
	@echo "  make fmt           - Format code"
	@echo "  make vet           - Run go vet"
	@echo "  make tidy          - Tidy go modules"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make install-tools - Install development tools"

# Run tests
test:
	go test -v ./...

# Run tests with race detector
test-race:
	go test -v -race ./...

# Run tests with coverage
test-coverage:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linters
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...
	goimports -w .

# Run go vet
vet:
	go vet ./...

# Tidy go modules
tidy:
	go mod tidy
	go mod verify

# Clean build artifacts
clean:
	rm -f coverage.txt coverage.html
	rm -rf dist/
	go clean -cache

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install golang.org/x/tools/cmd/goimports@latest
	@echo "Note: Install golangci-lint from https://golangci-lint.run/usage/install/"

# Run all checks (for CI or pre-commit)
check: fmt vet lint test

# Quick validation before commit
pre-commit: tidy fmt vet test-race
	@echo "All checks passed!"
