# Go-LLMs Makefile 

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
GOFMT=$(GOCMD) fmt
GOMOD=$(GOCMD) mod
GOGET=$(GOCMD) get

# Project parameters
BINARY_DIR=bin
CMD_DIR=cmd
PACKAGE_DIR=pkg
EXAMPLES_DIR=examples
MAIN_PACKAGE=github.com/lexlapax/go-llms

# Binary names
BINARY_NAME_MAIN=go-llms
BINARY_NAME_EXAMPLES=examples

# Build flags
LDFLAGS=-ldflags "-s -w"
BUILD_FLAGS=-v

# Test flags
TEST_FLAGS=-v -race -coverprofile=coverage.out -covermode=atomic

# Dependency flags
DEP_FLAGS=-v

# Commands
.PHONY: all build clean test test-pkg test-verbose test-verbose-pkg test-race test-race-pkg test-short test-short-pkg test-func benchmark benchmark-pkg coverage coverage-pkg lint fmt vet mod-tidy mod-download help examples examples-all

# Default target
all: clean test build

# Build the main binary
build: 
	$(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_DIR)/$(BINARY_NAME_MAIN) $(LDFLAGS) ./$(CMD_DIR)/main.go

# Build all available example binaries
examples-all:
	@if [ -d "$(CMD_DIR)/$(EXAMPLES_DIR)" ]; then \
		for dir in $(CMD_DIR)/$(EXAMPLES_DIR)/*/; do \
			if [ -d "$$dir" ]; then \
				name=$$(basename $$dir); \
				echo "Building example: $$name"; \
				$(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_DIR)/$$name $(LDFLAGS) ./$(CMD_DIR)/$(EXAMPLES_DIR)/$$name; \
			fi; \
		done; \
	else \
		echo "No examples found in $(CMD_DIR)/$(EXAMPLES_DIR)"; \
	fi

# Build a specific example (usage: make example EXAMPLE=simple)
example:
	@if [ -z "$(EXAMPLE)" ]; then \
		echo "Usage: make example EXAMPLE=<example-name>"; \
		exit 1; \
	fi
	@if [ -d "$(CMD_DIR)/$(EXAMPLES_DIR)/$(EXAMPLE)" ]; then \
		echo "Building example: $(EXAMPLE)"; \
		$(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_DIR)/$(EXAMPLE) $(LDFLAGS) ./$(CMD_DIR)/$(EXAMPLES_DIR)/$(EXAMPLE); \
	else \
		echo "Example $(EXAMPLE) not found in $(CMD_DIR)/$(EXAMPLES_DIR)/$(EXAMPLE)"; \
		exit 1; \
	fi

# Run all tests
test:
	$(GOTEST) $(TEST_FLAGS) ./...

# Run tests for a specific package (usage: make test-pkg PKG=schema/validation)
test-pkg:
	@if [ -z "$(PKG)" ]; then \
		echo "Usage: make test-pkg PKG=<package-path>"; \
		exit 1; \
	fi
	$(GOTEST) $(TEST_FLAGS) ./$(PACKAGE_DIR)/$(PKG)

# Run tests with verbose output
test-verbose:
	$(GOTEST) -v ./...

# Run tests for a specific package with verbose output (usage: make test-verbose-pkg PKG=schema/validation)
test-verbose-pkg:
	@if [ -z "$(PKG)" ]; then \
		echo "Usage: make test-verbose-pkg PKG=<package-path>"; \
		exit 1; \
	fi
	$(GOTEST) -v ./$(PACKAGE_DIR)/$(PKG)

# Run tests with race detection
test-race:
	$(GOTEST) -race ./...

# Run tests for a specific package with race detection (usage: make test-race-pkg PKG=schema/validation)
test-race-pkg:
	@if [ -z "$(PKG)" ]; then \
		echo "Usage: make test-race-pkg PKG=<package-path>"; \
		exit 1; \
	fi
	$(GOTEST) -race ./$(PACKAGE_DIR)/$(PKG)

# Run only short tests (useful for quick checks)
test-short:
	$(GOTEST) -short ./...

# Run only short tests for a specific package (usage: make test-short-pkg PKG=schema/validation)
test-short-pkg:
	@if [ -z "$(PKG)" ]; then \
		echo "Usage: make test-short-pkg PKG=<package-path>"; \
		exit 1; \
	fi
	$(GOTEST) -short ./$(PACKAGE_DIR)/$(PKG)

# Test a specific test function (usage: make test-func PKG=schema/validation FUNC=TestArrayValidation)
test-func:
	@if [ -z "$(PKG)" ] || [ -z "$(FUNC)" ]; then \
		echo "Usage: make test-func PKG=<package-path> FUNC=<function-name>"; \
		exit 1; \
	fi
	$(GOTEST) -v ./$(PACKAGE_DIR)/$(PKG) -run "$(FUNC)"

# Run benchmarks
benchmark:
	$(GOTEST) -bench=. -benchmem ./...

# Run benchmarks for a specific package (usage: make benchmark-pkg PKG=schema/validation)
benchmark-pkg:
	@if [ -z "$(PKG)" ]; then \
		echo "Usage: make benchmark-pkg PKG=<package-path>"; \
		exit 1; \
	fi
	$(GOTEST) -bench=. -benchmem ./$(PACKAGE_DIR)/$(PKG)

# Generate test coverage
coverage: test
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

# Generate test coverage for a specific package
coverage-pkg:
	@if [ -z "$(PKG)" ]; then \
		echo "Usage: make coverage-pkg PKG=<package-path>"; \
		exit 1; \
	fi
	$(GOTEST) -coverprofile=coverage.out -covermode=atomic ./$(PACKAGE_DIR)/$(PKG)
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report for $(PKG) generated at coverage.html"

# Run linting
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "Running golangci-lint..."; \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

# Format code
fmt:
	$(GOFMT) ./...

# Run vet
vet:
	$(GOVET) ./...

# Tidy dependencies
mod-tidy:
	$(GOMOD) tidy

# Download dependencies
mod-download:
	$(GOMOD) download

# Clean build artifacts
clean:
	rm -rf $(BINARY_DIR)/*
	mkdir -p $(BINARY_DIR)
	rm -f coverage.out coverage.html

# Install golangci-lint
install-lint:
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Help message
help:
	@echo "Go-LLMs Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make all              Build and test everything"
	@echo "  make build            Build the main binary"
	@echo "  make example          Build a specific example (usage: make example EXAMPLE=simple)"
	@echo "  make examples-all     Build all example binaries"
	@echo ""
	@echo "Testing:"
	@echo "  make test             Run all tests with race detection and coverage"
	@echo "  make test-pkg         Run tests for a specific package with race detection and coverage"
	@echo "                        (usage: make test-pkg PKG=schema/validation)"
	@echo "  make test-verbose     Run all tests with verbose output"
	@echo "  make test-verbose-pkg Run tests for a specific package with verbose output"
	@echo "                        (usage: make test-verbose-pkg PKG=schema/validation)"
	@echo "  make test-race        Run all tests with race detection"
	@echo "  make test-race-pkg    Run tests for a specific package with race detection"
	@echo "                        (usage: make test-race-pkg PKG=schema/validation)"
	@echo "  make test-short       Run only short tests"
	@echo "  make test-short-pkg   Run only short tests for a specific package"
	@echo "                        (usage: make test-short-pkg PKG=schema/validation)"
	@echo "  make test-func        Run a specific test function"
	@echo "                        (usage: make test-func PKG=schema/validation FUNC=TestArrayValidation)"
	@echo "  make benchmark        Run benchmarks for all packages"
	@echo "  make benchmark-pkg    Run benchmarks for a specific package"
	@echo "                        (usage: make benchmark-pkg PKG=schema/validation)"
	@echo "  make coverage         Generate test coverage report"
	@echo "  make coverage-pkg     Generate test coverage report for a specific package"
	@echo "                        (usage: make coverage-pkg PKG=schema/validation)"
	@echo ""
	@echo "Code quality:"
	@echo "  make lint             Run linters"
	@echo "  make fmt              Format Go code"
	@echo "  make vet              Run Go vet"
	@echo ""
	@echo "Dependencies:"
	@echo "  make mod-tidy         Tidy Go module dependencies"
	@echo "  make mod-download     Download Go module dependencies"
	@echo "  make install-lint     Install golangci-lint"
	@echo ""
	@echo "Maintenance:"
	@echo "  make clean            Clean build artifacts"
	@echo "  make help             Show this help message"