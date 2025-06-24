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
BENCHMARKS_DIR=tests/benchmarks
TESTS_DIR=tests

# Binary names
BINARY_NAME=go-llms

# Build flags
LDFLAGS=-ldflags "-s -w"
BUILD_FLAGS=-v
DEBUG_BUILD_FLAGS=-v -tags debug
DEBUG_BUILD_FLAGS_VERBOSE=-v -tags debug -gcflags="all=-N -l"

# Test flags
TEST_FLAGS=-v -race -coverprofile=coverage.out -covermode=atomic
TEST_VERBOSE_FLAGS=-v
TEST_RACE_FLAGS=-race
TEST_SHORT_FLAGS=-short
DEBUG_TEST_FLAGS=-v -tags debug

# Benchmark flags
BENCH_FLAGS=-bench=. -benchmem

# Color definitions for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

# Declare PHONY targets (organized by category)
.PHONY: all help \
	build build-debug build-all build-examples build-example \
	test test-quick test-all test-debug test-pkg test-func test-short test-integration test-stress \
	bench bench-all bench-pkg bench-specific \
	profile profile-cpu profile-mem \
	coverage coverage-pkg coverage-view \
	lint fmt vet quality check generate docs-api docs-validate \
	deps clean clean-all \
	dev watch

#=============================================================================
# DEFAULT TARGET
#=============================================================================
# Default target shows help
all: help

#=============================================================================
# QUICK TARGETS (Most commonly used)
#=============================================================================

# Quick build - just the main binary
build:
	@echo "$(GREEN)Building main binary...$(NC)"
	$(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_DIR)/$(BINARY_NAME) $(LDFLAGS) ./$(CMD_DIR)/

# Quick test - runs unit tests only (no integration/stress tests)
test:
	@echo "$(GREEN)Running unit tests...$(NC)"
	$(GOTEST) $(TEST_FLAGS) `$(GOCMD) list ./... | grep -v -E '(integration|multi_provider|stress|profiling|metrics|benchmarks)'`

# Quick check - format, vet, and lint
check: fmt vet lint
	@echo "$(GREEN)✓ All checks passed!$(NC)"

# Development mode - build and test
dev: clean check test build
	@echo "$(GREEN)✓ Development build complete!$(NC)"

#=============================================================================
# BUILD TARGETS
#=============================================================================

# Build with debug logging enabled
build-debug:
	@echo "$(YELLOW)Building with debug logging enabled...$(NC)"
	@echo "$(YELLOW)Set GO_LLMS_DEBUG=all or GO_LLMS_DEBUG=component1,component2$(NC)"
	$(GOBUILD) $(DEBUG_BUILD_FLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-debug $(LDFLAGS) ./$(CMD_DIR)/

# Build all binaries (main + examples)
build-all: build build-examples
	@echo "$(GREEN)✓ All binaries built!$(NC)"

# Build all example binaries
build-examples:
	@echo "$(GREEN)Building example binaries...$(NC)"
	@if [ -d "$(CMD_DIR)/$(EXAMPLES_DIR)" ]; then \
		for dir in $(CMD_DIR)/$(EXAMPLES_DIR)/*/; do \
			if [ -d "$$dir" ]; then \
				name=$$(basename $$dir); \
				echo "  Building $$name..."; \
				$(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_DIR)/$$name $(LDFLAGS) ./$(CMD_DIR)/$(EXAMPLES_DIR)/$$name; \
			fi; \
		done; \
		echo "$(GREEN)✓ Examples built!$(NC)"; \
	else \
		echo "$(RED)No examples found in $(CMD_DIR)/$(EXAMPLES_DIR)$(NC)"; \
	fi

# Build a specific example (usage: make build-example EXAMPLE=simple)
build-example:
	@if [ -z "$(EXAMPLE)" ]; then \
		echo "$(RED)Usage: make build-example EXAMPLE=<example-name>$(NC)"; \
		exit 1; \
	fi
	@if [ -d "$(CMD_DIR)/$(EXAMPLES_DIR)/$(EXAMPLE)" ]; then \
		echo "$(GREEN)Building example: $(EXAMPLE)$(NC)"; \
		$(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_DIR)/$(EXAMPLE) $(LDFLAGS) ./$(CMD_DIR)/$(EXAMPLES_DIR)/$(EXAMPLE); \
	else \
		echo "$(RED)Example $(EXAMPLE) not found$(NC)"; \
		exit 1; \
	fi

#=============================================================================
# TEST TARGETS
#=============================================================================

# Quick test - unit tests only
test-quick:
	@echo "$(GREEN)Running quick tests (unit tests only)...$(NC)"
	$(GOTEST) $(TEST_SHORT_FLAGS) `$(GOCMD) list ./... | grep -v -E '(integration|multi_provider|stress)'`

# Run all tests (unit + integration + stress)
test-all: test test-integration test-stress
	@echo "$(GREEN)✓ All tests passed!$(NC)"

# Run tests with debug logging
test-debug:
	@echo "$(YELLOW)Running tests with debug logging...$(NC)"
	@echo "$(YELLOW)Set GO_LLMS_DEBUG=all or GO_LLMS_DEBUG=component1,component2$(NC)"
	$(GOTEST) $(DEBUG_TEST_FLAGS) `$(GOCMD) list ./... | grep -v -E '(integration|multi_provider|stress)'`

# Run tests for a specific package
test-pkg:
	@if [ -z "$(PKG)" ]; then \
		echo "$(RED)Usage: make test-pkg PKG=<package-path>$(NC)"; \
		echo "$(YELLOW)Example: make test-pkg PKG=schema/validation$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Testing package: $(PKG)$(NC)"
	$(GOTEST) $(TEST_FLAGS) ./$(PACKAGE_DIR)/$(PKG)

# Test a specific function
test-func:
	@if [ -z "$(PKG)" ] || [ -z "$(FUNC)" ]; then \
		echo "$(RED)Usage: make test-func PKG=<package-path> FUNC=<function-name>$(NC)"; \
		echo "$(YELLOW)Example: make test-func PKG=schema/validation FUNC=TestArrayValidation$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Testing function $(FUNC) in package $(PKG)$(NC)"
	$(GOTEST) $(TEST_VERBOSE_FLAGS) ./$(PACKAGE_DIR)/$(PKG) -run "$(FUNC)"

# Run only short tests
test-short:
	@echo "$(GREEN)Running short tests...$(NC)"
	$(GOTEST) $(TEST_SHORT_FLAGS) ./...

# Run integration tests (requires API keys)
test-integration:
	@echo "$(YELLOW)Running integration tests (requires API keys)...$(NC)"
	$(GOTEST) $(TEST_VERBOSE_FLAGS) ./$(TESTS_DIR)/integration/...

# Run stress tests
test-stress:
	@echo "$(YELLOW)Running stress tests...$(NC)"
	$(GOTEST) $(TEST_VERBOSE_FLAGS) ./$(TESTS_DIR)/stress/...

#=============================================================================
# BENCHMARK TARGETS
#=============================================================================

# Run benchmarks
bench:
	@echo "$(GREEN)Running benchmarks...$(NC)"
	$(GOTEST) $(BENCH_FLAGS) ./$(BENCHMARKS_DIR)/...

# Run all benchmarks (including in packages)
bench-all:
	@echo "$(GREEN)Running all benchmarks...$(NC)"
	$(GOTEST) $(BENCH_FLAGS) ./...

# Run benchmarks for a specific package
bench-pkg:
	@if [ -z "$(PKG)" ]; then \
		echo "$(RED)Usage: make bench-pkg PKG=<package-path>$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Running benchmarks for package: $(PKG)$(NC)"
	$(GOTEST) $(BENCH_FLAGS) ./$(PACKAGE_DIR)/$(PKG)

# Run a specific benchmark
bench-specific:
	@if [ -z "$(BENCH)" ]; then \
		echo "$(RED)Usage: make bench-specific BENCH=<benchmark-name>$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Running benchmark: $(BENCH)$(NC)"
	$(GOTEST) -bench=$(BENCH) $(BENCH_FLAGS) ./$(BENCHMARKS_DIR)/...

#=============================================================================
# PROFILING TARGETS
#=============================================================================

# Profile CPU usage
profile-cpu:
	@echo "$(GREEN)Profiling CPU usage...$(NC)"
	$(GOTEST) $(BENCH_FLAGS) -cpuprofile=cpu.prof ./$(BENCHMARKS_DIR)/...
	@echo "$(YELLOW)View profile with: go tool pprof cpu.prof$(NC)"

# Profile memory usage
profile-mem:
	@echo "$(GREEN)Profiling memory usage...$(NC)"
	$(GOTEST) $(BENCH_FLAGS) -memprofile=mem.prof ./$(BENCHMARKS_DIR)/...
	@echo "$(YELLOW)View profile with: go tool pprof mem.prof$(NC)"

# Combined profiling
profile: profile-cpu profile-mem

#=============================================================================
# COVERAGE TARGETS
#=============================================================================

# Generate test coverage
coverage:
	@echo "$(GREEN)Generating coverage report...$(NC)"
	@$(MAKE) test
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report generated at coverage.html$(NC)"

# Generate coverage for specific package
coverage-pkg:
	@if [ -z "$(PKG)" ]; then \
		echo "$(RED)Usage: make coverage-pkg PKG=<package-path>$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Generating coverage for package: $(PKG)$(NC)"
	$(GOTEST) -coverprofile=coverage.out -covermode=atomic ./$(PACKAGE_DIR)/$(PKG)
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report generated at coverage.html$(NC)"

# View coverage report
coverage-view: coverage
	@if [ "$(shell uname)" = "Darwin" ]; then \
		open coverage.html; \
	elif [ "$(shell uname)" = "Linux" ]; then \
		xdg-open coverage.html 2>/dev/null || echo "$(YELLOW)Could not open coverage.html automatically$(NC)"; \
	else \
		echo "$(GREEN)Coverage report generated at coverage.html$(NC)"; \
	fi

#=============================================================================
# CODE QUALITY TARGETS
#=============================================================================

# Run all quality checks
quality: fmt vet lint
	@echo "$(GREEN)✓ Code quality checks passed!$(NC)"

# Generate code (tool metadata)
generate:
	@echo "$(GREEN)Generating tool metadata...$(NC)"
	$(GOCMD) generate ./pkg/agent/tools
	@echo "$(GREEN)✓ Tool metadata generated successfully!$(NC)"

# Generate API documentation
docs-api:
	@echo "$(GREEN)Generating API documentation...$(NC)"
	$(GOCMD) run scripts/generate-api-docs.go
	@echo "$(GREEN)✓ API documentation generated in docs/api/$(NC)"

# Validate documentation links
docs-validate:
	@echo "$(GREEN)Validating documentation links...$(NC)"
	$(GOCMD) run scripts/validate-doc-links.go
	@echo "$(GREEN)✓ Documentation links validated$(NC)"

# Format code
fmt:
	@echo "$(GREEN)Formatting code...$(NC)"
	$(GOFMT) ./...

# Run vet
vet:
	@echo "$(GREEN)Running go vet...$(NC)"
	$(GOVET) ./...

# Run linting
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "$(GREEN)Running golangci-lint...$(NC)"; \
		golangci-lint run ./...; \
	else \
		echo "$(YELLOW)golangci-lint not installed. Install with:$(NC)"; \
		echo "$(BLUE)  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(NC)"; \
		exit 1; \
	fi

#=============================================================================
# DEPENDENCY MANAGEMENT
#=============================================================================

# Manage dependencies
deps:
	@echo "$(GREEN)Managing dependencies...$(NC)"
	$(GOMOD) tidy
	$(GOMOD) download
	@echo "$(GREEN)✓ Dependencies updated!$(NC)"

#=============================================================================
# CLEANUP TARGETS
#=============================================================================

# Clean build artifacts
clean:
	@echo "$(GREEN)Cleaning build artifacts...$(NC)"
	rm -rf $(BINARY_DIR)/*
	mkdir -p $(BINARY_DIR)
	rm -f coverage.out coverage.html *.prof
	@echo "$(GREEN)✓ Clean complete!$(NC)"

# Clean everything including Go cache
clean-all: clean
	@echo "$(YELLOW)Cleaning Go cache...$(NC)"
	$(GOCMD) clean -cache -testcache -modcache
	@echo "$(GREEN)✓ Deep clean complete!$(NC)"

#=============================================================================
# DEVELOPMENT HELPERS
#=============================================================================

# Watch for changes and rebuild (requires entr)
watch:
	@if command -v entr >/dev/null 2>&1; then \
		echo "$(GREEN)Watching for changes...$(NC)"; \
		find . -name '*.go' | entr -c make dev; \
	else \
		echo "$(YELLOW)entr not installed. Install with:$(NC)"; \
		echo "$(BLUE)  macOS: brew install entr$(NC)"; \
		echo "$(BLUE)  Linux: apt-get install entr$(NC)"; \
		exit 1; \
	fi

#=============================================================================
# HELP TARGET
#=============================================================================

help:
	@echo "$(BLUE)Go-LLMs Makefile$(NC)"
	@echo ""
	@echo "$(GREEN)Quick Targets:$(NC)"
	@echo "  $(YELLOW)make build$(NC)         Build the main binary"
	@echo "  $(YELLOW)make test$(NC)          Run unit tests"
	@echo "  $(YELLOW)make check$(NC)         Run format, vet, and lint"
	@echo "  $(YELLOW)make dev$(NC)           Run check, test, and build"
	@echo ""
	@echo "$(GREEN)Build Targets:$(NC)"
	@echo "  $(YELLOW)make build-debug$(NC)   Build with debug logging (use GO_LLMS_DEBUG=all)"
	@echo "  $(YELLOW)make build-all$(NC)     Build main binary and all examples"
	@echo "  $(YELLOW)make build-example EXAMPLE=simple$(NC)  Build specific example"
	@echo ""
	@echo "$(GREEN)Test Targets:$(NC)"
	@echo "  $(YELLOW)make test-quick$(NC)    Run quick unit tests"
	@echo "  $(YELLOW)make test-all$(NC)      Run all tests (unit + integration + stress)"
	@echo "  $(YELLOW)make test-debug$(NC)    Run tests with debug logging"
	@echo "  $(YELLOW)make test-pkg PKG=schema/validation$(NC)  Test specific package"
	@echo "  $(YELLOW)make test-func PKG=schema/validation FUNC=TestArray$(NC)  Test specific function"
	@echo ""
	@echo "$(GREEN)Benchmark & Profile:$(NC)"
	@echo "  $(YELLOW)make bench$(NC)         Run benchmarks"
	@echo "  $(YELLOW)make profile-cpu$(NC)   Profile CPU usage"
	@echo "  $(YELLOW)make profile-mem$(NC)   Profile memory usage"
	@echo ""
	@echo "$(GREEN)Code Quality:$(NC)"
	@echo "  $(YELLOW)make quality$(NC)       Run all code quality checks"
	@echo "  $(YELLOW)make generate$(NC)      Generate tool metadata"
	@echo "  $(YELLOW)make docs-api$(NC)      Generate API documentation"
	@echo "  $(YELLOW)make docs-validate$(NC) Validate documentation links"
	@echo "  $(YELLOW)make coverage$(NC)      Generate coverage report"
	@echo "  $(YELLOW)make coverage-view$(NC) Generate and open coverage report"
	@echo ""
	@echo "$(GREEN)Other:$(NC)"
	@echo "  $(YELLOW)make deps$(NC)          Tidy and download dependencies"
	@echo "  $(YELLOW)make clean$(NC)         Clean build artifacts"
	@echo "  $(YELLOW)make watch$(NC)         Watch for changes and rebuild (requires entr)"
	@echo ""
	@echo "$(BLUE)Debug Logging:$(NC)"
	@echo "  Build with debug: $(YELLOW)make build-debug$(NC)"
	@echo "  Test with debug:  $(YELLOW)make test-debug$(NC)"
	@echo "  Set GO_LLMS_DEBUG=all or GO_LLMS_DEBUG=param_cache,schema"
	@echo ""
	@echo "$(BLUE)Examples:$(NC)"
	@echo "  $(YELLOW)make test-pkg PKG=llm/provider$(NC)"
	@echo "  $(YELLOW)make bench-specific BENCH=BenchmarkConsensus$(NC)"
	@echo "  $(YELLOW)GO_LLMS_DEBUG=param_cache make test-debug$(NC)"