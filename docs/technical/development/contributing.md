# Contributing: Code Organization and Style Guide

> **[Project Root](/) / [Documentation](/docs/) / [Technical Documentation](/docs/technical/) / [Development](/docs/technical/development/) / Contributing**

Comprehensive guide to contributing to Go-LLMs, covering code organization principles, style guidelines, development workflows, testing requirements, documentation standards, and project governance for maintaining high-quality, consistent codebase contributions.

## Project Architecture and Organization

### Repository Structure

```
go-llms/
├── pkg/                    # Library packages (public API)
│   ├── llm/               # LLM provider implementations
│   │   ├── provider/      # Individual provider implementations
│   │   ├── core/          # Core interfaces and types
│   │   └── util/          # LLM utilities
│   ├── agent/             # Agent framework
│   │   ├── core/          # Agent interfaces and base types
│   │   ├── workflow/      # Workflow orchestration
│   │   ├── tools/         # Tool integration
│   │   └── state/         # State management
│   ├── schema/            # JSON schema validation
│   ├── structured/        # Structured output parsing
│   ├── errors/            # Error handling system
│   └── testutils/         # Testing utilities and fixtures
├── cmd/                   # Command-line applications
│   ├── cli/               # Main CLI application
│   └── examples/          # Example applications
├── tests/                 # Integration and end-to-end tests
│   ├── integration/       # Integration test suites
│   ├── e2e/              # End-to-end test scenarios
│   └── fixtures/         # Test data and fixtures
├── docs/                  # Documentation
│   ├── user-guide/        # User-facing documentation
│   ├── technical/         # Technical documentation
│   └── api/               # API documentation
├── scripts/               # Build and development scripts
├── .github/               # GitHub workflows and templates
└── internal/              # Internal packages (private)
    ├── build/             # Build utilities
    ├── testutils/         # Internal testing utilities
    └── tools/             # Development tools
```

### Package Organization Principles

#### 1. Public vs Private APIs

```go
// pkg/ - Public API packages
// - Stable interfaces exposed to users
// - Comprehensive documentation required
// - Backward compatibility guarantees
// - Semantic versioning applies

// internal/ - Private implementation packages
// - Implementation details
// - Can change without notice
// - No compatibility guarantees
// - Used by cmd/ and pkg/ packages
```

#### 2. Package Naming Conventions

```go
// Good package names
pkg/llm/provider/openai     // Clear, descriptive, lowercase
pkg/agent/workflow          // Logical grouping
pkg/schema/validation       // Specific purpose

// Avoid
pkg/llm/OpenAI             // Capitalized
pkg/agent/workflowStuff    // Vague naming
pkg/misc                   // Generic names
pkg/utils                  // Overly broad
```

#### 3. Dependency Direction

```
┌─────────────┐    ┌─────────────┐
│     cmd/    │───▶│    pkg/     │
└─────────────┘    └─────────────┘
       │                  │
       ▼                  ▼
┌─────────────┐    ┌─────────────┐
│  internal/  │◀───│  external   │
└─────────────┘    └─────────────┘
```

**Rules:**
- `cmd/` can import `pkg/` and `internal/`
- `pkg/` should NOT import `internal/` (use interfaces)
- `internal/` can import `pkg/` for interfaces
- Avoid circular dependencies
- External dependencies should be isolated behind interfaces

## Code Style and Standards

### Go Code Style

#### 1. Formatting and Layout

```go
// File header - REQUIRED for all Go files
// ABOUTME: This file implements the OpenAI provider for LLM communication
// ABOUTME: Handles authentication, request formatting, and response parsing

package openai

import (
    "context"
    "fmt"
    "net/http"
    
    "github.com/lexlapax/go-llms/pkg/llm/core"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
)

// Package-level constants
const (
    DefaultBaseURL = "https://api.openai.com/v1"
    DefaultModel   = "gpt-3.5-turbo"
    MaxTokens      = 4096
)

// Package-level variables
var (
    ErrInvalidAPIKey = errors.New("invalid API key")
    ErrRateLimit     = errors.New("rate limit exceeded")
)
```

#### 2. Interface Design

```go
// Good interface design - focused and cohesive
type Provider interface {
    // Core functionality grouped logically
    Complete(ctx context.Context, request *CompletionRequest) (*CompletionResponse, error)
    CompleteStream(ctx context.Context, request *CompletionRequest) (<-chan StreamChunk, error)
    
    // Configuration
    GetConfig() ProviderConfig
    SetConfig(config ProviderConfig) error
    
    // Metadata
    GetModels(ctx context.Context) ([]Model, error)
    GetCapabilities() Capabilities
}

// Avoid large, unfocused interfaces
type BadInterface interface {
    // Too many responsibilities
    Complete(ctx context.Context, request *CompletionRequest) (*CompletionResponse, error)
    ValidateAPIKey(key string) error
    LogRequest(request *CompletionRequest)
    ParseResponse(data []byte) (*CompletionResponse, error)
    CalculateCost(tokens int) float64
    SendMetrics(metrics Metrics)
    // ... many more methods
}
```

#### 3. Error Handling

```go
// Custom error types with context
type ProviderError struct {
    Provider string `json:"provider"`
    Code     string `json:"code"`
    Message  string `json:"message"`
    Details  map[string]interface{} `json:"details,omitempty"`
    Cause    error  `json:"-"`
}

func (e *ProviderError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %s (caused by: %v)", e.Provider, e.Message, e.Cause)
    }
    return fmt.Sprintf("%s: %s", e.Provider, e.Message)
}

func (e *ProviderError) Unwrap() error {
    return e.Cause
}

// Error creation helpers
func NewProviderError(provider, code, message string) *ProviderError {
    return &ProviderError{
        Provider: provider,
        Code:     code,
        Message:  message,
        Details:  make(map[string]interface{}),
    }
}

// Usage example
func (p *OpenAIProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
    if req.APIKey == "" {
        return nil, NewProviderError("openai", "AUTH_MISSING", "API key is required")
    }
    
    resp, err := p.client.Post(ctx, req)
    if err != nil {
        // Wrap external errors
        return nil, &ProviderError{
            Provider: "openai",
            Code:     "REQUEST_FAILED",
            Message:  "failed to complete request",
            Cause:    err,
        }
    }
    
    return resp, nil
}
```

#### 4. Configuration Patterns

```go
// Configuration structs with validation
type ProviderConfig struct {
    APIKey      string        `yaml:"api_key" json:"api_key" validate:"required"`
    BaseURL     string        `yaml:"base_url" json:"base_url" validate:"url"`
    Model       string        `yaml:"model" json:"model" validate:"required"`
    Temperature *float64      `yaml:"temperature,omitempty" json:"temperature,omitempty" validate:"min=0,max=2"`
    MaxTokens   *int          `yaml:"max_tokens,omitempty" json:"max_tokens,omitempty" validate:"min=1"`
    Timeout     time.Duration `yaml:"timeout" json:"timeout" validate:"min=1s"`
    RetryConfig *RetryConfig  `yaml:"retry,omitempty" json:"retry,omitempty"`
}

// Validation method
func (c *ProviderConfig) Validate() error {
    if c.APIKey == "" {
        return fmt.Errorf("api_key is required")
    }
    
    if c.Temperature != nil && (*c.Temperature < 0 || *c.Temperature > 2) {
        return fmt.Errorf("temperature must be between 0 and 2")
    }
    
    if c.BaseURL != "" {
        if _, err := url.Parse(c.BaseURL); err != nil {
            return fmt.Errorf("invalid base_url: %w", err)
        }
    }
    
    return nil
}

// Default configuration
func DefaultProviderConfig() ProviderConfig {
    return ProviderConfig{
        BaseURL:     DefaultBaseURL,
        Model:       DefaultModel,
        Temperature: &[]float64{0.7}[0], // Pointer to value
        MaxTokens:   &[]int{1000}[0],
        Timeout:     30 * time.Second,
        RetryConfig: DefaultRetryConfig(),
    }
}
```

#### 5. Concurrency Patterns

```go
// Safe concurrent access with proper synchronization
type ThreadSafeRegistry struct {
    providers map[string]Provider
    mu        sync.RWMutex
    closed    chan struct{}
    once      sync.Once
}

func (r *ThreadSafeRegistry) Register(name string, provider Provider) error {
    select {
    case <-r.closed:
        return fmt.Errorf("registry is closed")
    default:
    }
    
    r.mu.Lock()
    defer r.mu.Unlock()
    
    if _, exists := r.providers[name]; exists {
        return fmt.Errorf("provider %s already registered", name)
    }
    
    r.providers[name] = provider
    return nil
}

func (r *ThreadSafeRegistry) Get(name string) (Provider, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    provider, exists := r.providers[name]
    if !exists {
        return nil, fmt.Errorf("provider %s not found", name)
    }
    
    return provider, nil
}

func (r *ThreadSafeRegistry) Close() error {
    r.once.Do(func() {
        close(r.closed)
    })
    return nil
}
```

### Documentation Standards

#### 1. Godoc Comments

```go
// Package openai provides an OpenAI API client implementation for the Go-LLMs library.
//
// This package implements the core Provider interface and provides comprehensive
// support for OpenAI's Chat Completions API, including streaming responses,
// function calling, and error handling.
//
// Example usage:
//
//	provider := openai.New(openai.Config{
//		APIKey: "your-api-key",
//		Model:  "gpt-4",
//	})
//	
//	response, err := provider.Complete(ctx, &core.CompletionRequest{
//		Messages: []core.Message{{Role: "user", Content: "Hello!"}},
//	})
//
// The package handles authentication, rate limiting, and error recovery
// automatically. For streaming responses, use CompleteStream method.
package openai

// Provider implements the core.Provider interface for OpenAI API.
//
// It provides thread-safe access to OpenAI's Chat Completions API with
// support for streaming, function calling, and comprehensive error handling.
//
// The provider automatically handles:
//   - Authentication via API key
//   - Request/response serialization
//   - Rate limiting and retries
//   - Error classification and wrapping
//
// Configuration is managed through the ProviderConfig struct, which
// supports all OpenAI API parameters including model selection,
// temperature, and token limits.
type Provider struct {
    config ProviderConfig
    client *http.Client
    mu     sync.RWMutex
}

// Complete sends a completion request to the OpenAI API.
//
// The method handles authentication, request formatting, and response parsing
// automatically. It returns a structured response containing the completion
// text and metadata.
//
// Parameters:
//   - ctx: Context for request cancellation and timeout control
//   - request: Completion request containing messages and options
//
// Returns:
//   - CompletionResponse: Structured response with completion and metadata
//   - error: Wrapped error with provider context if the request fails
//
// Common errors:
//   - ErrInvalidAPIKey: When API key is missing or invalid
//   - ErrRateLimit: When rate limits are exceeded
//   - ErrModelNotFound: When specified model is not available
//
// Example:
//
//	response, err := provider.Complete(ctx, &CompletionRequest{
//		Messages: []Message{{Role: "user", Content: "Explain quantum computing"}},
//		Temperature: &[]float64{0.7}[0],
//		MaxTokens: &[]int{500}[0],
//	})
//	if err != nil {
//		return fmt.Errorf("completion failed: %w", err)
//	}
//	fmt.Println(response.Content)
func (p *Provider) Complete(ctx context.Context, request *core.CompletionRequest) (*core.CompletionResponse, error) {
    // Implementation details...
}
```

#### 2. README Documentation

Each package should include a README.md with:

```markdown
# Package Name

Brief description of the package purpose and functionality.

## Features

- Feature 1 with brief description
- Feature 2 with brief description
- Feature 3 with brief description

## Installation

```go
import "github.com/lexlapax/go-llms/pkg/package/name"
```

## Quick Start

```go
// Basic usage example
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/lexlapax/go-llms/pkg/package/name"
)

func main() {
    // Create instance
    instance := name.New(name.Config{
        // Configuration options
    })
    
    // Use the instance
    result, err := instance.DoSomething(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(result)
}
```

## Configuration

### Required Options

- `option1`: Description of required option
- `option2`: Description of required option

### Optional Options

- `option3`: Description with default value
- `option4`: Description with default value

## Examples

### Example 1: Basic Usage

[Detailed example with explanation]

### Example 2: Advanced Configuration

[Advanced example with explanation]

## Error Handling

Common errors and how to handle them:

- `ErrType1`: When this occurs and how to handle
- `ErrType2`: When this occurs and how to handle

## Contributing

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for development guidelines.

## License

[License information]
```

## Testing Guidelines

### 1. Test Organization

```go
// provider_test.go - Unit tests for provider functionality
package openai

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "github.com/lexlapax/go-llms/pkg/testutils"
)

// TestProvider_Complete tests the Complete method
func TestProvider_Complete(t *testing.T) {
    tests := []struct {
        name     string
        config   ProviderConfig
        request  *core.CompletionRequest
        want     *core.CompletionResponse
        wantErr  bool
        errType  error
    }{
        {
            name: "successful_completion",
            config: ProviderConfig{
                APIKey: "test-api-key",
                Model:  "gpt-3.5-turbo",
            },
            request: &core.CompletionRequest{
                Messages: []core.Message{
                    {Role: "user", Content: "Hello!"},
                },
            },
            want: &core.CompletionResponse{
                Content: "Hello! How can I help you today?",
                Model:   "gpt-3.5-turbo",
            },
            wantErr: false,
        },
        {
            name: "missing_api_key",
            config: ProviderConfig{
                Model: "gpt-3.5-turbo",
            },
            request: &core.CompletionRequest{
                Messages: []core.Message{
                    {Role: "user", Content: "Hello!"},
                },
            },
            wantErr: true,
            errType: ErrInvalidAPIKey,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            provider := New(tt.config)
            if !tt.wantErr {
                // Mock successful API response
                testutils.MockHTTPResponse(t, provider.client, 200, mockSuccessResponse)
            }
            
            // Execute
            ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
            defer cancel()
            
            got, err := provider.Complete(ctx, tt.request)
            
            // Assert
            if tt.wantErr {
                require.Error(t, err)
                if tt.errType != nil {
                    assert.ErrorIs(t, err, tt.errType)
                }
                return
            }
            
            require.NoError(t, err)
            assert.Equal(t, tt.want.Content, got.Content)
            assert.Equal(t, tt.want.Model, got.Model)
        })
    }
}

// Benchmark tests
func BenchmarkProvider_Complete(b *testing.B) {
    provider := New(ProviderConfig{
        APIKey: "test-api-key",
        Model:  "gpt-3.5-turbo",
    })
    
    request := &core.CompletionRequest{
        Messages: []core.Message{
            {Role: "user", Content: "Hello!"},
        },
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := provider.Complete(context.Background(), request)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### 2. Integration Tests

```go
// integration_test.go - Integration tests with real APIs
// +build integration

package openai_test

import (
    "context"
    "os"
    "testing"
    
    "github.com/stretchr/testify/require"
    
    "github.com/lexlapax/go-llms/pkg/llm/provider/openai"
)

func TestProvider_Integration(t *testing.T) {
    // Skip if no API key provided
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        t.Skip("OPENAI_API_KEY not set, skipping integration tests")
    }
    
    provider := openai.New(openai.ProviderConfig{
        APIKey: apiKey,
        Model:  "gpt-3.5-turbo",
    })
    
    t.Run("real_api_completion", func(t *testing.T) {
        response, err := provider.Complete(context.Background(), &core.CompletionRequest{
            Messages: []core.Message{
                {Role: "user", Content: "Say 'integration test successful'"},
            },
        })
        
        require.NoError(t, err)
        require.NotEmpty(t, response.Content)
        require.Contains(t, response.Content, "integration test successful")
    })
}
```

### 3. Test Utilities

```go
// pkg/testutils/providers.go - Reusable test utilities
package testutils

// MockProvider creates a mock provider for testing
func MockProvider(responses map[string]*core.CompletionResponse) core.Provider {
    return &mockProvider{
        responses: responses,
    }
}

type mockProvider struct {
    responses map[string]*core.CompletionResponse
}

func (m *mockProvider) Complete(ctx context.Context, req *core.CompletionRequest) (*core.CompletionResponse, error) {
    key := req.Messages[0].Content
    if resp, exists := m.responses[key]; exists {
        return resp, nil
    }
    return nil, fmt.Errorf("no mock response for: %s", key)
}

// HTTPResponseMocker helps mock HTTP responses
type HTTPResponseMocker struct {
    responses map[string]HTTPResponse
}

type HTTPResponse struct {
    StatusCode int
    Body       string
    Headers    map[string]string
}

func NewHTTPMocker() *HTTPResponseMocker {
    return &HTTPResponseMocker{
        responses: make(map[string]HTTPResponse),
    }
}

func (m *HTTPResponseMocker) AddResponse(url string, response HTTPResponse) {
    m.responses[url] = response
}

// Test fixtures
const (
    MockOpenAIResponse = `{
        "choices": [{
            "message": {
                "role": "assistant",
                "content": "Hello! How can I help you today?"
            },
            "finish_reason": "stop"
        }],
        "model": "gpt-3.5-turbo",
        "usage": {
            "prompt_tokens": 10,
            "completion_tokens": 12,
            "total_tokens": 22
        }
    }`
)
```

## Development Workflow

### 1. Git Workflow

```bash
# Feature development
git checkout -b feature/provider-anthropic
git add .
git commit -m "feat: add Anthropic provider implementation

- Implement core Provider interface
- Add streaming support
- Include comprehensive error handling
- Add unit and integration tests

Closes #123"

# Code review and merge
git push origin feature/provider-anthropic
# Create pull request
# Address review comments
# Merge to main
```

### 2. Commit Message Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples:**
```
feat(llm): add Anthropic provider support

Implement complete Anthropic provider with:
- Claude model support
- Streaming responses
- Function calling
- Comprehensive error handling

Includes unit tests and integration tests.

Closes #456
```

### 3. Pull Request Process

1. **Create Feature Branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Development Checklist**
   - [ ] Code follows style guidelines
   - [ ] All tests pass (`make test`)
   - [ ] Code is properly documented
   - [ ] No linting errors (`make lint`)
   - [ ] Integration tests pass if applicable
   - [ ] Examples updated if needed

3. **Pre-submission**
   ```bash
   make fmt      # Format code
   make lint     # Check for issues
   make test     # Run all tests
   make generate # Update generated files
   ```

4. **Pull Request Template**
   ```markdown
   ## Description
   Brief description of changes
   
   ## Type of Change
   - [ ] Bug fix
   - [ ] New feature
   - [ ] Breaking change
   - [ ] Documentation update
   
   ## Testing
   - [ ] Unit tests pass
   - [ ] Integration tests pass
   - [ ] Manual testing completed
   
   ## Checklist
   - [ ] Code follows style guidelines
   - [ ] Self-review completed
   - [ ] Documentation updated
   - [ ] No breaking changes (or documented)
   ```

### 4. Code Review Guidelines

**For Reviewers:**
- Focus on design, correctness, and maintainability
- Check for proper error handling
- Verify test coverage
- Ensure documentation is complete
- Look for potential security issues

**Review Checklist:**
- [ ] Code is readable and well-structured
- [ ] Error handling is comprehensive
- [ ] Tests cover edge cases
- [ ] Documentation is clear and complete
- [ ] Performance considerations addressed
- [ ] Security implications considered
- [ ] Backward compatibility maintained

## Performance and Optimization

### 1. Performance Guidelines

```go
// Use appropriate data structures
type FastLookup struct {
    data map[string]*Item // O(1) lookup
    mu   sync.RWMutex
}

// Pool expensive objects
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 0, 1024)
    },
}

func ProcessData(data []byte) []byte {
    buf := bufferPool.Get().([]byte)
    defer bufferPool.Put(buf[:0]) // Reset length but keep capacity
    
    // Process data using pooled buffer
    return append(buf, processedData...)
}

// Avoid unnecessary allocations
func BuildQuery(parts []string) string {
    // Good: Pre-allocate capacity
    var builder strings.Builder
    builder.Grow(estimateSize(parts))
    
    for _, part := range parts {
        builder.WriteString(part)
    }
    return builder.String()
}
```

### 2. Memory Management

```go
// Resource cleanup patterns
type ResourceManager struct {
    resources []io.Closer
    mu        sync.Mutex
}

func (rm *ResourceManager) Add(resource io.Closer) {
    rm.mu.Lock()
    defer rm.mu.Unlock()
    rm.resources = append(rm.resources, resource)
}

func (rm *ResourceManager) Close() error {
    rm.mu.Lock()
    defer rm.mu.Unlock()
    
    var firstErr error
    for _, resource := range rm.resources {
        if err := resource.Close(); err != nil && firstErr == nil {
            firstErr = err
        }
    }
    rm.resources = rm.resources[:0] // Clear slice
    return firstErr
}

// Context-based cancellation
func LongRunningOperation(ctx context.Context) error {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            // Do work
            if err := doWork(); err != nil {
                return err
            }
        }
    }
}
```

## Security Considerations

### 1. Input Validation

```go
// Validate all inputs
func ValidateAPIKey(key string) error {
    if key == "" {
        return fmt.Errorf("API key cannot be empty")
    }
    
    if len(key) < 10 {
        return fmt.Errorf("API key too short")
    }
    
    // Check for valid characters
    if !regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`).MatchString(key) {
        return fmt.Errorf("API key contains invalid characters")
    }
    
    return nil
}

// Sanitize file paths
func SafeFilePath(path string) (string, error) {
    // Clean the path
    cleaned := filepath.Clean(path)
    
    // Check for directory traversal
    if strings.Contains(cleaned, "..") {
        return "", fmt.Errorf("directory traversal not allowed")
    }
    
    // Ensure absolute path
    if !filepath.IsAbs(cleaned) {
        return "", fmt.Errorf("relative paths not allowed")
    }
    
    return cleaned, nil
}
```

### 2. Secrets Management

```go
// Never log sensitive data
func (p *Provider) logRequest(req *Request) {
    // Sanitize request before logging
    sanitized := *req
    sanitized.APIKey = "[REDACTED]"
    sanitized.Headers = sanitizeHeaders(req.Headers)
    
    log.Printf("Request: %+v", sanitized)
}

// Use secure defaults
type SecurityConfig struct {
    TLSEnabled     bool          `yaml:"tls_enabled" default:"true"`
    TLSMinVersion  string        `yaml:"tls_min_version" default:"1.2"`
    Timeout        time.Duration `yaml:"timeout" default:"30s"`
    MaxRequestSize int64         `yaml:"max_request_size" default:"10485760"` // 10MB
}
```

This comprehensive contributing guide provides all the necessary information for developers to contribute effectively to the Go-LLMs project while maintaining high code quality and consistency standards.