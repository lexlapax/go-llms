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
BENCHMARKS_DIR=benchmarks
TESTS_DIR=tests

# Binary names
BINARY_NAME=go-llms

# Build flags
LDFLAGS=-ldflags "-s -w"
BUILD_FLAGS=-v

# Test flags
TEST_FLAGS=-v -race -coverprofile=coverage.out -covermode=atomic
TEST_VERBOSE_FLAGS=-v
TEST_RACE_FLAGS=-race
TEST_SHORT_FLAGS=-short

# Benchmark flags
BENCH_FLAGS=-bench=. -benchmem

# Declare PHONY targets
.PHONY: all help \
	build build-all build-examples build-example \
	test test-all test-pkg test-func test-short test-short-pkg test-cmd test-examples \
	test-integration test-integration-mock test-multi-provider test-stress test-stress-provider \
	test-stress-agent test-stress-structured test-stress-pool \
	benchmark benchmark-all benchmark-pkg benchmark-specific \
	profile profile-cpu profile-mem profile-block \
	coverage coverage-pkg coverage-view \
	lint install-lint fmt vet \
	deps deps-tidy deps-download \
	clean clean-all

# Default target
all: clean test build build-examples

# Main binary build
build:
	$(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_DIR)/$(BINARY_NAME) $(LDFLAGS) ./$(CMD_DIR)/main.go

# Build all binaries
build-all: build build-examples

# Build all example binaries
build-examples:
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

# Build a specific example (usage: make build-example EXAMPLE=simple)
build-example:
	@if [ -z "$(EXAMPLE)" ]; then \
		echo "Usage: make build-example EXAMPLE=<example-name>"; \
		exit 1; \
	fi
	@if [ -d "$(CMD_DIR)/$(EXAMPLES_DIR)/$(EXAMPLE)" ]; then \
		echo "Building example: $(EXAMPLE)"; \
		$(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_DIR)/$(EXAMPLE) $(LDFLAGS) ./$(CMD_DIR)/$(EXAMPLES_DIR)/$(EXAMPLE); \
	else \
		echo "Example $(EXAMPLE) not found in $(CMD_DIR)/$(EXAMPLES_DIR)/$(EXAMPLE)"; \
		exit 1; \
	fi

# Test targets
# Run all tests (excluding integration, multi-provider, and stress tests)
test:
	$(GOTEST) $(TEST_FLAGS) `$(GOCMD) list ./... | grep -v github.com/lexlapax/go-llms/$(TESTS_DIR)/integration | grep -v github.com/lexlapax/go-llms/$(TESTS_DIR)/multi_provider | grep -v github.com/lexlapax/go-llms/$(TESTS_DIR)/stress`

# Run all tests including integration, multi-provider, and stress tests
test-all: test test-integration test-multi-provider test-stress

# Run tests for a specific package (usage: make test-pkg PKG=schema/validation)
test-pkg:
	@if [ -z "$(PKG)" ]; then \
		echo "Usage: make test-pkg PKG=<package-path>"; \
		exit 1; \
	fi
	$(GOTEST) $(TEST_FLAGS) ./$(PACKAGE_DIR)/$(PKG)

# Test a specific test function (usage: make test-func PKG=schema/validation FUNC=TestArrayValidation)
test-func:
	@if [ -z "$(PKG)" ] || [ -z "$(FUNC)" ]; then \
		echo "Usage: make test-func PKG=<package-path> FUNC=<function-name>"; \
		exit 1; \
	fi
	$(GOTEST) $(TEST_VERBOSE_FLAGS) ./$(PACKAGE_DIR)/$(PKG) -run "$(FUNC)"

# Run only short tests
test-short:
	$(GOTEST) $(TEST_SHORT_FLAGS) ./...

# Run only short tests for a specific package
test-short-pkg:
	@if [ -z "$(PKG)" ]; then \
		echo "Usage: make test-short-pkg PKG=<package-path>"; \
		exit 1; \
	fi
	$(GOTEST) $(TEST_SHORT_FLAGS) ./$(PACKAGE_DIR)/$(PKG)

# Test the command line client
test-cmd:
	$(GOTEST) $(TEST_VERBOSE_FLAGS) ./$(CMD_DIR)

# Test examples
test-examples:
	@if [ -z "$(EXAMPLE)" ]; then \
		echo "Testing all examples..."; \
		for dir in $(CMD_DIR)/$(EXAMPLES_DIR)/*/; do \
			if [ -d "$$dir" ]; then \
				name=$$(basename $$dir); \
				echo "Testing example: $$name"; \
				$(GOTEST) $(TEST_VERBOSE_FLAGS) ./$(CMD_DIR)/$(EXAMPLES_DIR)/$$name; \
			fi; \
		done; \
	else \
		echo "Testing example: $(EXAMPLE)"; \
		$(GOTEST) $(TEST_VERBOSE_FLAGS) ./$(CMD_DIR)/$(EXAMPLES_DIR)/$(EXAMPLE); \
	fi

# Run integration tests (requires API keys)
test-integration:
	$(GOTEST) $(TEST_VERBOSE_FLAGS) ./$(TESTS_DIR)/integration/...

# Run mock-only integration tests (doesn't require API keys)
test-integration-mock:
	$(GOTEST) $(TEST_VERBOSE_FLAGS) ./$(TESTS_DIR)/integration/validation_test.go ./$(TESTS_DIR)/integration/agent_test.go

# Run multi-provider tests
test-multi-provider:
	$(GOTEST) $(TEST_VERBOSE_FLAGS) ./$(TESTS_DIR)/multi_provider/...

# Run all stress tests
test-stress:
	$(GOTEST) $(TEST_VERBOSE_FLAGS) ./$(TESTS_DIR)/stress/...

# Run provider stress tests
test-stress-provider:
	$(GOTEST) $(TEST_VERBOSE_FLAGS) ./$(TESTS_DIR)/stress/provider_stress_test.go

# Run agent stress tests
test-stress-agent:
	$(GOTEST) $(TEST_VERBOSE_FLAGS) ./$(TESTS_DIR)/stress/agent_stress_test.go

# Run structured output processor stress tests
test-stress-structured:
	$(GOTEST) $(TEST_VERBOSE_FLAGS) ./$(TESTS_DIR)/stress/structured_stress_test.go

# Run memory pool stress tests
test-stress-pool:
	$(GOTEST) $(TEST_VERBOSE_FLAGS) ./$(TESTS_DIR)/stress/pool_stress_test.go

# Benchmark targets
# Run all benchmarks
benchmark:
	$(GOTEST) $(BENCH_FLAGS) ./$(BENCHMARKS_DIR)/...

# Run benchmarks for all components
benchmark-all:
	$(GOTEST) $(BENCH_FLAGS) ./...
	$(GOTEST) $(BENCH_FLAGS) ./$(BENCHMARKS_DIR)/...

# Run benchmarks for a specific package
benchmark-pkg:
	@if [ -z "$(PKG)" ]; then \
		echo "Usage: make benchmark-pkg PKG=<package-path>"; \
		exit 1; \
	fi
	$(GOTEST) $(BENCH_FLAGS) ./$(PACKAGE_DIR)/$(PKG)

# Run a specific benchmark (usage: make benchmark-specific BENCH=BenchmarkConsensus)
benchmark-specific:
	@if [ -z "$(BENCH)" ]; then \
		echo "Usage: make benchmark-specific BENCH=<benchmark-name>"; \
		exit 1; \
	fi
	$(GOTEST) -bench=$(BENCH) $(BENCH_FLAGS) ./$(BENCHMARKS_DIR)/...

# Profiling targets
# Profile CPU usage (creates cpu.prof)
profile-cpu:
	$(GOTEST) $(BENCH_FLAGS) -cpuprofile=cpu.prof ./$(BENCHMARKS_DIR)/...
	@echo "View profile with: go tool pprof cpu.prof"

# Profile memory usage (creates mem.prof)
profile-mem:
	$(GOTEST) $(BENCH_FLAGS) -memprofile=mem.prof ./$(BENCHMARKS_DIR)/...
	@echo "View profile with: go tool pprof mem.prof"

# Profile blocking operations (creates block.prof)
profile-block:
	$(GOTEST) $(BENCH_FLAGS) -blockprofile=block.prof ./$(BENCHMARKS_DIR)/...
	@echo "View profile with: go tool pprof block.prof"

# Combined profiling target
profile: profile-cpu profile-mem profile-block

# Coverage targets
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

# View coverage report
coverage-view: coverage
	@if [ "$(shell uname)" = "Darwin" ]; then \
		open coverage.html; \
	elif [ "$(shell uname)" = "Linux" ]; then \
		xdg-open coverage.html 2>/dev/null || echo "Could not open coverage.html automatically"; \
	else \
		echo "Coverage report generated at coverage.html"; \
	fi

# Code quality targets
# Run linting
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "Running golangci-lint..."; \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Run: make install-lint"; \
		exit 1; \
	fi

# Install golangci-lint
install-lint:
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "golangci-lint installed successfully"

# Format code
fmt:
	$(GOFMT) ./...

# Run vet
vet:
	$(GOVET) ./...

# Dependency targets
# Tidy dependencies
deps-tidy:
	$(GOMOD) tidy

# Download dependencies
deps-download:
	$(GOMOD) download

# Combined dependency management
deps: deps-tidy deps-download

# Clean targets
# Clean build artifacts
clean:
	rm -rf $(BINARY_DIR)/*
	mkdir -p $(BINARY_DIR)
	rm -f coverage.out coverage.html *.prof

# Clean everything including Go cache
clean-all: clean
	$(GOCMD) clean -cache -testcache -modcache
	@echo "Cleaned all build artifacts and Go cache"

# Help message
help:
	@echo "Go-LLMs Makefile"
	@echo ""
	@echo "Building:"
	@echo "  make build            Build the main binary"
	@echo "  make build-examples   Build all example binaries"
	@echo "  make build-example    Build a specific example (usage: make build-example EXAMPLE=simple)"
	@echo "  make build-all        Build main binary and all examples"
	@echo ""
	@echo "Testing:"
	@echo "  make test             Run all tests (excluding integration, multi-provider, and stress tests)"
	@echo "  make test-all         Run all tests including integration, multi-provider, and stress tests"
	@echo "  make test-pkg         Run tests for a specific package (usage: make test-pkg PKG=schema/validation)"
	@echo "  make test-func        Run a specific test function (usage: make test-func PKG=schema/validation FUNC=TestValidation)"
	@echo "  make test-short       Run only short tests (useful for quick checks)"
	@echo "  make test-short-pkg   Run only short tests for a specific package (usage: make test-short-pkg PKG=schema/validation)"
	@echo "  make test-cmd         Test the command line client"
	@echo "  make test-examples    Test all examples (or specific with EXAMPLE=name)"
	@echo "  make test-integration Run all integration tests (requires API keys)"
	@echo "  make test-integration-mock Run integration tests that don't require API keys"
	@echo "  make test-multi-provider Run multi-provider tests"
	@echo "  make test-stress      Run all stress tests"
	@echo "  make test-stress-provider Run provider stress tests"
	@echo "  make test-stress-agent Run agent workflow stress tests"
	@echo "  make test-stress-structured Run structured output processor stress tests"
	@echo "  make test-stress-pool Run memory pool stress tests"
	@echo ""
	@echo "Benchmarking:"
	@echo "  make benchmark        Run benchmarks in the benchmarks directory"
	@echo "  make benchmark-all    Run all benchmarks across the codebase"
	@echo "  make benchmark-pkg    Run benchmarks for a specific package (usage: make benchmark-pkg PKG=schema/validation)"
	@echo "  make benchmark-specific Run a specific benchmark (usage: make benchmark-specific BENCH=BenchmarkConsensus)"
	@echo ""
	@echo "Profiling:"
	@echo "  make profile          Run all profile types (CPU, memory, and block)"
	@echo "  make profile-cpu      Run benchmarks with CPU profiling"
	@echo "  make profile-mem      Run benchmarks with memory profiling"
	@echo "  make profile-block    Run benchmarks with block profiling (for concurrency issues)"
	@echo ""
	@echo "Coverage:"
	@echo "  make coverage         Generate test coverage report"
	@echo "  make coverage-pkg     Generate test coverage report for a specific package (usage: make coverage-pkg PKG=schema/validation)"
	@echo "  make coverage-view    Generate coverage report and open it in a browser"
	@echo ""
	@echo "Code Quality:"
	@echo "  make lint             Run linters"
	@echo "  make install-lint     Install golangci-lint"
	@echo "  make fmt              Format Go code"
	@echo "  make vet              Run Go vet"
	@echo ""
	@echo "Dependencies:"
	@echo "  make deps             Manage all dependencies (tidy and download)"
	@echo "  make deps-tidy        Tidy Go module dependencies"
	@echo "  make deps-download    Download Go module dependencies"
	@echo ""
	@echo "Maintenance:"
	@echo "  make clean            Clean build artifacts"
	@echo "  make clean-all        Clean everything including Go cache"
	@echo "  make all              Default target: clean, test, build, and build examples"
	@echo "  make help             Show this help message"