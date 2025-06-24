# Provider Setup: Environment Configuration and API Keys

> **[Project Root](/) / [Documentation](/docs/) / [User Guide](/docs/user-guide/) / [Guides](/docs/user-guide/guides/) / Provider Setup**

Master the setup and configuration of all LLM providers. From getting API keys to advanced authentication, environment management, and production deployment strategies.

## Why Provider Setup Matters

- **Security** - Proper API key management and authentication
- **Reliability** - Correct configuration prevents runtime errors
- **Flexibility** - Support multiple providers and environments
- **Production Ready** - Secure, scalable configuration patterns
- **Cost Management** - Optimize usage and billing

## Supported Providers

Go-LLMs supports 6 major LLM providers with unified configuration:

| Provider | Models | Authentication | Special Features |
|----------|--------|----------------|------------------|
| **OpenAI** | GPT-4o, GPT-4 Turbo, GPT-4o-mini | API Key | Function calling, vision, latest models |
| **Anthropic** | Claude 3.5 Sonnet, Claude 3.5 Haiku | API Key | Long context, constitutional AI |
| **Google Gemini** | Gemini 2.0 Flash, Gemini 1.5 Pro | API Key | Multimodal, fast inference |
| **Google Vertex AI** | Gemini + partner models | Service Account | Enterprise features, compliance |
| **Ollama** | Llama, Mistral, CodeLlama | None (local) | Local hosting, privacy, offline |
| **OpenRouter** | 400+ models | API Key | Model variety, cost optimization |

## Prerequisites

- [Installation completed](../getting-started/installation.md) ✅
- Basic understanding of environment variables ✅
- Access to at least one LLM provider ✅

---

## Quick Setup Guide

### 1. OpenAI Setup (Recommended for Beginners)

#### Get API Key
1. Visit [OpenAI Platform](https://platform.openai.com/)
2. Sign up or log in to your account
3. Navigate to [API Keys](https://platform.openai.com/api-keys)
4. Click "Create new secret key"
5. Copy the key (starts with `sk-...`)

#### Environment Configuration
```bash
# Basic setup
export OPENAI_API_KEY="sk-proj-your-key-here"

# Optional: Organization ID (for team accounts)
export OPENAI_ORGANIZATION="org-your-org-id"

# Optional: Project ID (for project isolation)
export OPENAI_PROJECT="proj-your-project-id"
```

#### Usage Example
```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
    llmDomain "github.com/lexlapax/go-llms/pkg/llm/domain"
)

func main() {
    // Method 1: Simple string-based creation
    agent, err := core.NewAgentFromString("assistant", "openai/gpt-4o-mini")
    if err != nil {
        log.Fatalf("Failed to create agent: %v", err)
    }

    // Method 2: Explicit provider creation with options
    apiKey := os.Getenv("OPENAI_API_KEY")
    orgID := os.Getenv("OPENAI_ORGANIZATION")
    
    var options []llmDomain.ProviderOption
    if orgID != "" {
        options = append(options, llmDomain.NewOpenAIOrganizationOption(orgID))
    }

    openaiProvider := provider.NewOpenAIProvider(apiKey, "gpt-4o-mini", options...)
    
    explicitAgent := core.NewLLMAgent("explicit-assistant", "gpt-4o-mini", core.LLMDeps{
        Provider: openaiProvider,
    })

    // Test the setup
    state := domain.NewState()
    state.Set("user_input", "Hello! Can you confirm that OpenAI is working?")

    result, err := agent.Run(context.Background(), state)
    if err != nil {
        log.Fatalf("OpenAI test failed: %v", err)
    }

    if response, exists := result.Get("response"); exists {
        fmt.Printf("✅ OpenAI setup successful!\nResponse: %v\n", response)
    }
}
```

### 2. Anthropic (Claude) Setup

#### Get API Key
1. Visit [Anthropic Console](https://console.anthropic.com/)
2. Sign up or log in
3. Navigate to [API Keys](https://console.anthropic.com/account/keys)
4. Create a new key
5. Copy the key (starts with `sk-ant-...`)

#### Environment Configuration
```bash
# Standard setup
export ANTHROPIC_API_KEY="sk-ant-your-key-here"

# Alternative naming (useful for multi-provider setups)
export GO_LLMS_ANTHROPIC_API_KEY="sk-ant-your-key-here"
```

#### Usage Example
```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
    llmDomain "github.com/lexlapax/go-llms/pkg/llm/domain"
)

func main() {
    // Method 1: String-based creation
    agent, err := core.NewAgentFromString("claude-assistant", "anthropic/claude-3-5-sonnet")
    if err != nil {
        log.Fatalf("Failed to create Claude agent: %v", err)
    }

    // Method 2: Explicit creation with system prompt option
    apiKey := os.Getenv("ANTHROPIC_API_KEY")
    systemPromptOption := llmDomain.NewAnthropicSystemPromptOption(
        "You are Claude, a helpful AI assistant created by Anthropic.",
    )

    anthropicProvider := provider.NewAnthropicProvider(
        apiKey,
        "claude-3-5-sonnet-latest",
        systemPromptOption,
    )

    explicitAgent := core.NewLLMAgent("explicit-claude", "claude-3-5-sonnet-latest", core.LLMDeps{
        Provider: anthropicProvider,
    })

    // Test with a reasoning task (Claude's strength)
    state := domain.NewState()
    state.Set("user_input", "Analyze the pros and cons of remote work vs office work")

    result, err := agent.Run(context.Background(), state)
    if err != nil {
        log.Fatalf("Anthropic test failed: %v", err)
    }

    if response, exists := result.Get("response"); exists {
        fmt.Printf("✅ Anthropic setup successful!\nAnalysis: %v\n", response)
    }
}
```

### 3. Google Gemini Setup

#### Get API Key
1. Visit [Google AI Studio](https://makersuite.google.com/)
2. Sign in with your Google account
3. Click "Get API key" 
4. Create a new API key
5. Copy the key (starts with `AI...`)

#### Environment Configuration
```bash
# Standard setup
export GEMINI_API_KEY="AIza-your-key-here"

# Alternative naming
export GO_LLMS_GEMINI_API_KEY="AIza-your-key-here"
```

#### Usage Example
```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
)

func main() {
    // Create Gemini agent
    agent, err := core.NewAgentFromString("gemini-assistant", "gemini/gemini-2.0-flash-latest")
    if err != nil {
        log.Fatalf("Failed to create Gemini agent: %v", err)
    }

    // Explicit provider creation
    apiKey := os.Getenv("GEMINI_API_KEY")
    geminiProvider := provider.NewGeminiProvider(apiKey, "gemini-2.0-flash-latest")

    explicitAgent := core.NewLLMAgent("explicit-gemini", "gemini-2.0-flash-latest", core.LLMDeps{
        Provider: geminiProvider,
    })

    // Test with a speed-focused task (Gemini's strength)
    state := domain.NewState()
    state.Set("user_input", "Quickly summarize the key benefits of Go programming language")

    result, err := agent.Run(context.Background(), state)
    if err != nil {
        log.Fatalf("Gemini test failed: %v", err)
    }

    if response, exists := result.Get("response"); exists {
        fmt.Printf("✅ Gemini setup successful!\nSummary: %v\n", response)
    }
}
```

### 4. Ollama (Local) Setup

#### Install Ollama
```bash
# macOS/Linux
curl -fsSL https://ollama.ai/install.sh | sh

# Or download from https://ollama.ai/download
```

#### Pull Models
```bash
# Recommended models for development
ollama pull llama3.2:3b      # Fast, good for testing
ollama pull codellama:7b     # Great for code tasks
ollama pull mistral:7b       # General purpose

# List available models
ollama list

# Check if Ollama is running
ollama --version
```

#### Environment Configuration
```bash
# Optional: Custom host (default: http://localhost:11434)
export OLLAMA_HOST="http://localhost:11434"

# Optional: Default model
export OLLAMA_MODEL="llama3.2:3b"
```

#### Usage Example
```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
)

func main() {
    // Check if Ollama is running
    host := os.Getenv("OLLAMA_HOST")
    if host == "" {
        host = "http://localhost:11434"
    }

    // Method 1: String-based creation
    agent, err := core.NewAgentFromString("local-assistant", "ollama/llama3.2:3b")
    if err != nil {
        log.Printf("String-based creation failed: %v", err)
        log.Println("Trying explicit provider creation...")
        
        // Method 2: Explicit provider creation
        ollamaProvider := provider.NewOllamaProvider(host, "llama3.2:3b")
        agent = core.NewLLMAgent("local-assistant", "llama3.2:3b", core.LLMDeps{
            Provider: ollamaProvider,
        })
    }

    // Test local model
    state := domain.NewState()
    state.Set("user_input", "Hello! Are you running locally?")

    result, err := agent.Run(context.Background(), state)
    if err != nil {
        log.Fatalf("Ollama test failed (is Ollama running?): %v", err)
    }

    if response, exists := result.Get("response"); exists {
        fmt.Printf("✅ Ollama setup successful!\nLocal Response: %v\n", response)
    }
}
```

### 5. OpenRouter Setup

#### Get API Key
1. Visit [OpenRouter](https://openrouter.ai/)
2. Sign up or log in
3. Go to [Keys](https://openrouter.ai/keys)
4. Create a new API key
5. Copy the key (starts with `sk-or-...`)

#### Environment Configuration
```bash
# Required: API key
export OPENROUTER_API_KEY="sk-or-your-key-here"

# Optional: Site URL for usage tracking
export OPENROUTER_SITE_URL="https://yoursite.com"

# Optional: App name for tracking
export OPENROUTER_APP_NAME="YourApp"
```

#### Usage Example
```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
)

func main() {
    // Use Claude via OpenRouter
    agent, err := core.NewAgentFromString("openrouter-assistant", "openrouter/anthropic/claude-3.5-sonnet")
    if err != nil {
        log.Fatalf("Failed to create OpenRouter agent: %v", err)
    }

    // Explicit provider creation with metadata
    apiKey := os.Getenv("OPENROUTER_API_KEY")
    siteURL := os.Getenv("OPENROUTER_SITE_URL")
    
    openrouterProvider := provider.NewOpenRouterProvider(apiKey, "anthropic/claude-3.5-sonnet")
    
    // Add optional metadata
    if siteURL != "" {
        // In a real implementation, you'd pass this as an option
        fmt.Printf("Using site URL: %s\n", siteURL)
    }

    explicitAgent := core.NewLLMAgent("explicit-openrouter", "claude-3.5-sonnet", core.LLMDeps{
        Provider: openrouterProvider,
    })

    // Test with model comparison task
    state := domain.NewState()
    state.Set("user_input", "Compare the advantages of different LLM providers")

    result, err := agent.Run(context.Background(), state)
    if err != nil {
        log.Fatalf("OpenRouter test failed: %v", err)
    }

    if response, exists := result.Get("response"); exists {
        fmt.Printf("✅ OpenRouter setup successful!\nComparison: %v\n", response)
    }
}
```

### 6. Google Vertex AI Setup

#### Setup Service Account
1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create or select a project
3. Enable the Vertex AI API
4. Create a service account with Vertex AI permissions
5. Download the service account key JSON file

#### Environment Configuration
```bash
# Method 1: Service account key file
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account.json"

# Method 2: Use gcloud authentication
gcloud auth application-default login

# Required: Project ID
export GOOGLE_CLOUD_PROJECT="your-project-id"

# Optional: Region (default: us-central1)
export GOOGLE_CLOUD_REGION="us-central1"
```

#### Usage Example
```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
)

func main() {
    projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
    if projectID == "" {
        log.Fatal("GOOGLE_CLOUD_PROJECT environment variable required")
    }

    region := os.Getenv("GOOGLE_CLOUD_REGION")
    if region == "" {
        region = "us-central1"
    }

    // Create Vertex AI agent
    vertexProvider := provider.NewVertexAIProvider(projectID, region, "gemini-pro")
    
    agent := core.NewLLMAgent("vertex-assistant", "gemini-pro", core.LLMDeps{
        Provider: vertexProvider,
    })

    // Test enterprise features
    state := domain.NewState()
    state.Set("user_input", "Explain the benefits of enterprise AI deployment")

    result, err := agent.Run(context.Background(), state)
    if err != nil {
        log.Fatalf("Vertex AI test failed: %v", err)
    }

    if response, exists := result.Get("response"); exists {
        fmt.Printf("✅ Vertex AI setup successful!\nEnterprise Guide: %v\n", response)
    }
}
```

---

## Advanced Configuration

### Multi-Provider Environment
```bash
# .env file for development
# OpenAI
OPENAI_API_KEY=sk-proj-your-openai-key
OPENAI_ORGANIZATION=org-your-org

# Anthropic
ANTHROPIC_API_KEY=sk-ant-your-anthropic-key

# Google
GEMINI_API_KEY=AIza-your-gemini-key
GOOGLE_CLOUD_PROJECT=your-project-id
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json

# OpenRouter
OPENROUTER_API_KEY=sk-or-your-openrouter-key
OPENROUTER_SITE_URL=https://yoursite.com

# Ollama
OLLAMA_HOST=http://localhost:11434
OLLAMA_MODEL=llama3.2:3b

# Optional: Go-LLMs prefixed versions
GO_LLMS_OPENAI_API_KEY=sk-proj-alternative-key
GO_LLMS_ANTHROPIC_API_KEY=sk-ant-alternative-key
GO_LLMS_GEMINI_API_KEY=AIza-alternative-key
```

### Configuration Management
```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
)

// ProviderConfig holds provider configuration
type ProviderConfig struct {
    Name      string
    Model     string
    Available bool
    Config    map[string]string
}

// ConfigManager manages provider configurations
type ConfigManager struct {
    providers map[string]ProviderConfig
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() *ConfigManager {
    cm := &ConfigManager{
        providers: make(map[string]ProviderConfig),
    }
    cm.detectProviders()
    return cm
}

// detectProviders automatically detects available providers
func (cm *ConfigManager) detectProviders() {
    // OpenAI
    if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
        cm.providers["openai"] = ProviderConfig{
            Name:      "OpenAI",
            Model:     "gpt-4o-mini",
            Available: true,
            Config: map[string]string{
                "api_key": apiKey,
                "org_id":  os.Getenv("OPENAI_ORGANIZATION"),
            },
        }
    }

    // Anthropic
    if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
        cm.providers["anthropic"] = ProviderConfig{
            Name:      "Anthropic",
            Model:     "claude-3-5-haiku-latest",
            Available: true,
            Config: map[string]string{
                "api_key": apiKey,
            },
        }
    }

    // Gemini
    if apiKey := os.Getenv("GEMINI_API_KEY"); apiKey != "" {
        cm.providers["gemini"] = ProviderConfig{
            Name:      "Google Gemini",
            Model:     "gemini-2.0-flash-latest",
            Available: true,
            Config: map[string]string{
                "api_key": apiKey,
            },
        }
    }

    // Ollama (check if running)
    cm.providers["ollama"] = ProviderConfig{
        Name:      "Ollama",
        Model:     "llama3.2:3b",
        Available: cm.checkOllamaAvailable(),
        Config: map[string]string{
            "host":  os.Getenv("OLLAMA_HOST"),
            "model": os.Getenv("OLLAMA_MODEL"),
        },
    }

    // OpenRouter
    if apiKey := os.Getenv("OPENROUTER_API_KEY"); apiKey != "" {
        cm.providers["openrouter"] = ProviderConfig{
            Name:      "OpenRouter",
            Model:     "anthropic/claude-3.5-haiku",
            Available: true,
            Config: map[string]string{
                "api_key":  apiKey,
                "site_url": os.Getenv("OPENROUTER_SITE_URL"),
            },
        }
    }
}

// checkOllamaAvailable checks if Ollama is running
func (cm *ConfigManager) checkOllamaAvailable() bool {
    // In a real implementation, you'd make an HTTP request to check
    // For now, just check if OLLAMA_HOST is set or default exists
    return os.Getenv("OLLAMA_HOST") != "" || true // Assume available
}

// GetAvailableProviders returns all available providers
func (cm *ConfigManager) GetAvailableProviders() []ProviderConfig {
    var available []ProviderConfig
    for _, config := range cm.providers {
        if config.Available {
            available = append(available, config)
        }
    }
    return available
}

// CreateAgent creates an agent with the best available provider
func (cm *ConfigManager) CreateAgent(name string, preferredProvider string) (domain.BaseAgent, error) {
    // Try preferred provider first
    if config, exists := cm.providers[preferredProvider]; exists && config.Available {
        return cm.createAgentWithProvider(name, preferredProvider, config)
    }

    // Fallback order: OpenAI -> Anthropic -> Gemini -> OpenRouter -> Ollama
    fallbackOrder := []string{"openai", "anthropic", "gemini", "openrouter", "ollama"}
    
    for _, providerName := range fallbackOrder {
        if config, exists := cm.providers[providerName]; exists && config.Available {
            log.Printf("Using fallback provider: %s", config.Name)
            return cm.createAgentWithProvider(name, providerName, config)
        }
    }

    return nil, fmt.Errorf("no providers available")
}

// createAgentWithProvider creates an agent with a specific provider
func (cm *ConfigManager) createAgentWithProvider(name, providerName string, config ProviderConfig) (domain.BaseAgent, error) {
    providerString := fmt.Sprintf("%s/%s", providerName, config.Model)
    return core.NewAgentFromString(name, providerString)
}

func main() {
    fmt.Println("🔧 Advanced Provider Configuration")
    fmt.Println("=================================")

    // Create configuration manager
    configManager := NewConfigManager()

    // Display available providers
    available := configManager.GetAvailableProviders()
    fmt.Printf("Available providers: %d\n", len(available))
    
    for _, config := range available {
        fmt.Printf("  ✓ %s (%s)\n", config.Name, config.Model)
    }

    if len(available) == 0 {
        fmt.Println("❌ No providers configured. Please set up at least one provider.")
        return
    }

    // Create agents with automatic fallback
    testCases := []struct {
        name      string
        preferred string
    }{
        {"fast-agent", "gemini"},
        {"reasoning-agent", "anthropic"},
        {"general-agent", "openai"},
        {"local-agent", "ollama"},
        {"budget-agent", "openrouter"},
    }

    for _, test := range testCases {
        fmt.Printf("\n--- Creating %s (preferred: %s) ---\n", test.name, test.preferred)
        
        agent, err := configManager.CreateAgent(test.name, test.preferred)
        if err != nil {
            fmt.Printf("❌ Failed to create agent: %v\n", err)
            continue
        }

        fmt.Printf("✅ Created agent: %s\n", agent.Name())
    }
}
```

### Production Configuration
```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
    llmDomain "github.com/lexlapax/go-llms/pkg/llm/domain"
)

// ProductionConfig holds production-ready configuration
type ProductionConfig struct {
    Environment     string
    PrimaryProvider string
    FallbackProvider string
    Timeout         time.Duration
    RetryAttempts   int
    RateLimiting    bool
    Monitoring      bool
    Debug           bool
}

// LoadProductionConfig loads configuration from environment
func LoadProductionConfig() *ProductionConfig {
    env := os.Getenv("ENVIRONMENT")
    if env == "" {
        env = "development"
    }

    config := &ProductionConfig{
        Environment:      env,
        PrimaryProvider:  os.Getenv("PRIMARY_LLM_PROVIDER"),
        FallbackProvider: os.Getenv("FALLBACK_LLM_PROVIDER"),
        Timeout:          30 * time.Second,
        RetryAttempts:    3,
        RateLimiting:     env == "production",
        Monitoring:       env == "production",
        Debug:           env == "development",
    }

    // Override defaults from environment
    if timeout := os.Getenv("LLM_TIMEOUT"); timeout != "" {
        if d, err := time.ParseDuration(timeout); err == nil {
            config.Timeout = d
        }
    }

    return config
}

// ProductionAgent wraps agents with production features
type ProductionAgent struct {
    primary  domain.BaseAgent
    fallback domain.BaseAgent
    config   *ProductionConfig
}

// NewProductionAgent creates a production-ready agent
func NewProductionAgent(name string, config *ProductionConfig) (*ProductionAgent, error) {
    // Create primary agent
    primary, err := createAgentForProvider(name+"-primary", config.PrimaryProvider)
    if err != nil {
        return nil, fmt.Errorf("failed to create primary agent: %w", err)
    }

    // Create fallback agent
    var fallback domain.BaseAgent
    if config.FallbackProvider != "" && config.FallbackProvider != config.PrimaryProvider {
        fallback, err = createAgentForProvider(name+"-fallback", config.FallbackProvider)
        if err != nil {
            log.Printf("Warning: fallback agent creation failed: %v", err)
        }
    }

    return &ProductionAgent{
        primary:  primary,
        fallback: fallback,
        config:   config,
    }, nil
}

// createAgentForProvider creates an agent for a specific provider
func createAgentForProvider(name, providerName string) (domain.BaseAgent, error) {
    switch providerName {
    case "openai":
        return core.NewAgentFromString(name, "openai/gpt-4o-mini")
    case "anthropic":
        return core.NewAgentFromString(name, "anthropic/claude-3-5-haiku")
    case "gemini":
        return core.NewAgentFromString(name, "gemini/gemini-2.0-flash")
    case "ollama":
        return core.NewAgentFromString(name, "ollama/llama3.2:3b")
    default:
        return core.NewAgentFromString(name, "openai/gpt-4o-mini") // Default fallback
    }
}

// Run executes the agent with production features
func (pa *ProductionAgent) Run(ctx context.Context, state domain.StateReader) (*domain.State, error) {
    // Add timeout to context
    ctx, cancel := context.WithTimeout(ctx, pa.config.Timeout)
    defer cancel()

    // Try primary agent
    result, err := pa.runWithRetry(ctx, pa.primary, state)
    if err == nil {
        if pa.config.Monitoring {
            log.Printf("Production agent success: primary provider used")
        }
        return result, nil
    }

    if pa.config.Debug {
        log.Printf("Primary agent failed: %v", err)
    }

    // Try fallback agent if available
    if pa.fallback != nil {
        result, fallbackErr := pa.runWithRetry(ctx, pa.fallback, state)
        if fallbackErr == nil {
            if pa.config.Monitoring {
                log.Printf("Production agent success: fallback provider used")
            }
            return result, nil
        }

        if pa.config.Debug {
            log.Printf("Fallback agent also failed: %v", fallbackErr)
        }
    }

    return nil, fmt.Errorf("all agents failed, primary: %v", err)
}

// runWithRetry runs an agent with retry logic
func (pa *ProductionAgent) runWithRetry(ctx context.Context, agent domain.BaseAgent, state domain.StateReader) (*domain.State, error) {
    var lastErr error
    
    for attempt := 0; attempt < pa.config.RetryAttempts; attempt++ {
        if attempt > 0 {
            // Exponential backoff
            backoff := time.Duration(attempt) * time.Second
            select {
            case <-time.After(backoff):
            case <-ctx.Done():
                return nil, ctx.Err()
            }
        }

        result, err := agent.Run(ctx, state)
        if err == nil {
            return result, nil
        }

        lastErr = err
        if pa.config.Debug {
            log.Printf("Attempt %d failed: %v", attempt+1, err)
        }
    }

    return nil, fmt.Errorf("all %d attempts failed, last error: %w", pa.config.RetryAttempts, lastErr)
}

func main() {
    fmt.Println("🏭 Production Provider Setup")
    fmt.Println("============================")

    // Load production configuration
    config := LoadProductionConfig()
    fmt.Printf("Environment: %s\n", config.Environment)
    fmt.Printf("Primary Provider: %s\n", config.PrimaryProvider)
    fmt.Printf("Fallback Provider: %s\n", config.FallbackProvider)

    // Create production agent
    prodAgent, err := NewProductionAgent("production-assistant", config)
    if err != nil {
        log.Fatalf("Failed to create production agent: %v", err)
    }

    // Test production features
    state := domain.NewState()
    state.Set("user_input", "Test production setup with timeout and retry logic")

    startTime := time.Now()
    result, err := prodAgent.Run(context.Background(), state)
    duration := time.Since(startTime)

    if err != nil {
        fmt.Printf("❌ Production test failed: %v\n", err)
        return
    }

    if response, exists := result.Get("response"); exists {
        fmt.Printf("✅ Production setup successful!\n")
        fmt.Printf("Response: %v\n", response)
        fmt.Printf("Duration: %v\n", duration)
    }

    fmt.Println("\n📊 Production Features Enabled:")
    fmt.Printf("  ⏱️ Timeout: %v\n", config.Timeout)
    fmt.Printf("  🔄 Retry Attempts: %d\n", config.RetryAttempts)
    fmt.Printf("  📈 Monitoring: %t\n", config.Monitoring)
    fmt.Printf("  🔍 Debug: %t\n", config.Debug)
}
```

## Security Best Practices

### API Key Management
```bash
# ✅ Good practices
export OPENAI_API_KEY="sk-proj-..."          # Secure environment variable
echo $OPENAI_API_KEY | vault kv put ...     # Store in secrets manager

# ❌ Bad practices
echo "sk-proj-..." > api_key.txt           # Never store in files
const API_KEY = "sk-proj-..."              # Never hardcode in source
```

### Environment Separation
```bash
# Development environment
cat > .env.development << EOF
OPENAI_API_KEY=sk-proj-dev-key
ANTHROPIC_API_KEY=sk-ant-dev-key
ENVIRONMENT=development
DEBUG=true
EOF

# Production environment
cat > .env.production << EOF
OPENAI_API_KEY=sk-proj-prod-key
ANTHROPIC_API_KEY=sk-ant-prod-key
ENVIRONMENT=production
DEBUG=false
MONITORING=true
EOF
```

### Secrets Management
```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/lexlapax/go-llms/pkg/agent/core"
)

// SecretsManager interface for different secret backends
type SecretsManager interface {
    GetSecret(key string) (string, error)
}

// EnvSecretsManager uses environment variables
type EnvSecretsManager struct{}

func (e *EnvSecretsManager) GetSecret(key string) (string, error) {
    value := os.Getenv(key)
    if value == "" {
        return "", fmt.Errorf("secret %s not found", key)
    }
    return value, nil
}

// SecureAgentFactory creates agents using secure secret management
type SecureAgentFactory struct {
    secrets SecretsManager
}

func NewSecureAgentFactory(secrets SecretsManager) *SecureAgentFactory {
    return &SecureAgentFactory{secrets: secrets}
}

func (f *SecureAgentFactory) CreateAgent(name, provider string) (domain.BaseAgent, error) {
    // Map provider to secret key
    secretKey := fmt.Sprintf("%s_API_KEY", strings.ToUpper(provider))
    
    apiKey, err := f.secrets.GetSecret(secretKey)
    if err != nil {
        return nil, fmt.Errorf("failed to get API key for %s: %w", provider, err)
    }

    // Create agent with retrieved secret
    return core.NewAgentFromString(name, fmt.Sprintf("%s/default", provider))
}
```

## Troubleshooting

### Common Issues

#### API Key Not Working
```bash
# Check if key is set
echo $OPENAI_API_KEY

# Test key validity
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY"
```

#### Provider Connection Issues
```go
// Test provider connectivity
func testProviderConnection(providerName string) error {
    agent, err := core.NewAgentFromString("test", providerName)
    if err != nil {
        return fmt.Errorf("failed to create %s agent: %w", providerName, err)
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    state := domain.NewState()
    state.Set("user_input", "test")

    _, err = agent.Run(ctx, state)
    return err
}
```

#### Environment Variable Issues
```bash
# Debug environment variables
env | grep -E "(OPENAI|ANTHROPIC|GEMINI|OLLAMA|OPENROUTER)"

# Check for hidden characters
echo "$OPENAI_API_KEY" | xxd
```

#### Rate Limiting
```go
// Handle rate limiting
func handleRateLimit(err error) bool {
    // Check if error is rate limit related
    return strings.Contains(err.Error(), "rate limit") ||
           strings.Contains(err.Error(), "quota")
}

// Implement exponential backoff
func retryWithBackoff(fn func() error, maxRetries int) error {
    for i := 0; i < maxRetries; i++ {
        err := fn()
        if err == nil {
            return nil
        }
        
        if handleRateLimit(err) {
            backoff := time.Duration(i+1) * time.Second
            time.Sleep(backoff)
            continue
        }
        
        return err
    }
    return fmt.Errorf("max retries exceeded")
}
```

## Next Steps

🔧 **Provider setup complete!** Continue with:

- **[Provider Selection](provider-selection.md)** - Choose the right provider for your use case
- **[Multi-Provider Strategies](multi-provider-strategies.md)** - Use multiple providers together
- **[Local Providers](local-providers.md)** - Deep dive into Ollama and local hosting
- **[Creating Agents](creating-agents.md)** - Build your first agents

### Quick Reference

- **[Configuration Reference](../reference/configuration-reference.md)** - All configuration options
- **[Provider Comparison](../reference/provider-comparison.md)** - Feature matrix and selection
- **[Error Codes](../reference/error-codes-reference.md)** - Common error solutions
- **[Best Practices](../reference/best-practices-checklist.md)** - Production checklist

---

**Need help?** Check our [troubleshooting guide](../advanced/troubleshooting.md) or join the discussion on [GitHub](https://github.com/lexlapax/go-llms/discussions).