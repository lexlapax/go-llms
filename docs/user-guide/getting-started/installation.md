# Installation & Environment Setup

> **[Project Root](/) / [Documentation](../..) / [User Guide](../../user-guide) / [Getting Started](../../user-guide/getting-started) / Installation**

Get Go-LLMs up and running in your development environment with this comprehensive setup guide.

## Requirements

### Go Version
- **Go 1.24.1 or later** (required)
- Verify your version: `go version`

```bash
# If you need to update Go
go install golang.org/dl/go1.24.1@latest
go1.24.1 download
```

### Operating System Support
- **Linux** (fully tested)
- **macOS** (fully tested)  
- **Windows** (community tested)

## Installation

### 1. Add Go-LLMs to Your Project

```bash
# Initialize a new Go module (if needed)
go mod init your-project-name

# Add Go-LLMs dependency
go get github.com/lexlapax/go-llms

# Verify installation
go list -m github.com/lexlapax/go-llms
```

### 2. Basic Import Test

Create a simple test to verify installation:

```go
// test_installation.go
package main

import (
    "fmt"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
)

func main() {
    // Test with mock provider (no API key needed)
    mock := provider.NewMockProvider()
    fmt.Println("Go-LLMs installed successfully!")
    fmt.Printf("Mock provider created: %T\n", mock)
}
```

```bash
# Run the test
go run test_installation.go
```

## Environment Configuration

### Provider API Keys

Go-LLMs supports multiple naming conventions for environment variables. Choose the pattern that works best for your setup:

#### OpenAI
```bash
# Option 1: Standard naming
export OPENAI_API_KEY="sk-..."

# Option 2: Go-LLMs prefixed (useful for multi-tool environments)
export GO_LLMS_OPENAI_API_KEY="sk-..."

# Optional: Organization ID
export OPENAI_ORGANIZATION="org-..."
```

#### Anthropic (Claude)
```bash
# Option 1: Standard naming
export ANTHROPIC_API_KEY="sk-ant-..."

# Option 2: Go-LLMs prefixed
export GO_LLMS_ANTHROPIC_API_KEY="sk-ant-..."
```

#### Google Gemini
```bash
# Option 1: Standard naming
export GEMINI_API_KEY="AI..."

# Option 2: Go-LLMs prefixed
export GO_LLMS_GEMINI_API_KEY="AI..."
```

#### Google Vertex AI
```bash
# Service account key path
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account.json"

# Or use gcloud authentication
gcloud auth application-default login
```

#### Ollama (Local Models)
```bash
# Ollama host (optional, defaults to localhost:11434)
export OLLAMA_HOST="http://localhost:11434"

# Default model (optional)
export OLLAMA_MODEL="llama3.2:3b"
```

#### OpenRouter
```bash
# OpenRouter API key
export OPENROUTER_API_KEY="sk-or-..."

# Optional: Site URL for usage tracking
export OPENROUTER_SITE_URL="https://yoursite.com"
```

### Configuration Files

#### Option 1: `.env` File

Create a `.env` file in your project root:

```bash
# .env
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...
GEMINI_API_KEY=AI...
OLLAMA_HOST=http://localhost:11434

# Add to .gitignore
echo ".env" >> .gitignore
```

Load with a package like `godotenv`:
```go
import "github.com/joho/godotenv"

func init() {
    godotenv.Load()
}
```

#### Option 2: Configuration Struct

For production applications:

```go
type Config struct {
    OpenAIKey     string `env:"OPENAI_API_KEY"`
    AnthropicKey  string `env:"ANTHROPIC_API_KEY"`
    GeminiKey     string `env:"GEMINI_API_KEY"`
    OllamaHost    string `env:"OLLAMA_HOST" envDefault:"http://localhost:11434"`
}
```

## Development Environment Setup

### 1. Project Structure

Recommended project structure for Go-LLMs applications:

```
your-project/
├── cmd/                    # Application entry points
│   └── main.go
├── internal/               # Private application code
│   ├── agents/            # Your agent implementations
│   ├── tools/             # Custom tools
│   └── config/            # Configuration management
├── pkg/                   # Public library code (if building a library)
├── tests/                 # Integration and end-to-end tests
├── .env                   # Environment variables (do not commit)
├── .gitignore            # Include .env and other secrets
├── go.mod                # Go module definition
└── README.md             # Project documentation
```

### 2. Development Tools

#### Optional but Recommended

**Air (Hot Reload)**
```bash
go install github.com/air-verse/air@latest
air init  # Creates .air.toml config
air       # Run with hot reload
```

**Delve (Debugging)**
```bash
go install github.com/go-delve/delve/cmd/dlv@latest
dlv debug cmd/main.go
```

**golangci-lint (Code Quality)**
```bash
# Install
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2

# Run
golangci-lint run
```

### 3. Editor Setup

#### VS Code
Install these extensions for the best experience:

- **Go** (golang.go) - Official Go extension
- **Go Outline** - Code outline and symbols
- **Go Test Explorer** - Test runner integration
- **REST Client** - For testing HTTP APIs

#### GoLand/IntelliJ
Enable these features:

- Go plugin (should be built-in)
- Environment variables from .env files
- Code completion and navigation
- Integrated debugger and test runner

## Local Model Setup (Ollama)

For development with local models (privacy, no API costs):

### 1. Install Ollama

```bash
# macOS/Linux
curl -fsSL https://ollama.ai/install.sh | sh

# Or download from https://ollama.ai/download
```

### 2. Pull Models

```bash
# Recommended models for development
ollama pull llama3.2:3b      # Fast, good for testing
ollama pull codellama:7b     # Good for code tasks
ollama pull mistral:7b       # General purpose

# List installed models
ollama list
```

### 3. Verify Ollama Setup

```bash
# Check if Ollama is running
curl http://localhost:11434/api/version

# Test generation
curl http://localhost:11434/api/generate -d '{
  "model": "llama3.2:3b",
  "prompt": "Hello, world!",
  "stream": false
}'
```

## Verification

### 1. Run Example Projects

Test your setup with the included examples:

```bash
# Clone the repository (optional, for examples)
git clone https://github.com/lexlapax/go-llms.git
cd go-llms

# Test with mock provider (no API key needed)
go run cmd/examples/simple/main.go

# Test with OpenAI (requires API key)
export OPENAI_API_KEY="your-key"
go run cmd/examples/provider-openai/main.go

# Test with Ollama (requires local Ollama)
go run cmd/examples/provider-ollama/main.go
```

### 2. Environment Check Tool

Create a verification script:

```go
// verify_setup.go
package main

import (
    "fmt"
    "os"
    "github.com/lexlapax/go-llms/pkg/agent/core"
)

func main() {
    fmt.Println("Go-LLMs Environment Check")
    fmt.Println("========================")
    
    // Check API keys
    providers := map[string]string{
        "OpenAI":    "OPENAI_API_KEY",
        "Anthropic": "ANTHROPIC_API_KEY", 
        "Gemini":    "GEMINI_API_KEY",
        "Ollama":    "OLLAMA_HOST",
    }
    
    for name, envVar := range providers {
        if value := os.Getenv(envVar); value != "" {
            fmt.Printf("✓ %s: %s configured\n", name, envVar)
        } else {
            fmt.Printf("✗ %s: %s not set\n", name, envVar)
        }
    }
    
    // Test agent creation
    fmt.Println("\nTesting agent creation...")
    agent, err := core.NewAgentFromString("test", "mock")
    if err != nil {
        fmt.Printf("✗ Agent creation failed: %v\n", err)
    } else {
        fmt.Printf("✓ Agent created successfully: %s\n", agent.Name())
    }
}
```

## Troubleshooting

### Common Issues

#### Import Path Errors
```bash
# If you see "module not found" errors
go mod tidy
go clean -modcache
go mod download
```

#### API Key Issues
```bash
# Check if environment variables are loaded
go run -ldflags="-X main.debug=true" your-app.go

# Or add debug prints
fmt.Printf("OpenAI Key loaded: %t\n", os.Getenv("OPENAI_API_KEY") != "")
```

#### Network/Proxy Issues
```bash
# For corporate environments
export GOPROXY=direct
export GOSUMDB=off

# Or configure through your proxy
export GOPROXY=https://your-proxy.com
```

#### Ollama Connection Issues
```bash
# Check if Ollama is running
ps aux | grep ollama

# Start Ollama service
ollama serve

# Check port availability
netstat -an | grep 11434
```

### Getting Help

If you encounter issues:

1. **Check the examples** - Run included examples to isolate the issue
2. **Review logs** - Enable debug logging for detailed error information
3. **Community support** - [GitHub Discussions](https://github.com/lexlapax/go-llms/discussions)
4. **Bug reports** - [GitHub Issues](https://github.com/lexlapax/go-llms/issues)

## Next Steps

✅ **Installation complete!** Continue with:

- **[First Steps](first-steps.md)** - Build your first 3 applications
- **[Choosing Providers](choosing-providers.md)** - Select the right provider for your use case
- **[Key Concepts](key-concepts.md)** - Understand the core abstractions
- **[Quick Start](quickstart.md)** - 5-minute interactive guide

---

**Need help?** Check our [troubleshooting guide](../advanced/troubleshooting.md) or ask on [GitHub Discussions](https://github.com/lexlapax/go-llms/discussions).