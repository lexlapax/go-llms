# Configuration Reference: All Configuration Options

> **[Project Root](/) / [Documentation](/docs/) / [User Guide](/docs/user-guide/) / [Reference](/docs/user-guide/reference/) / Configuration**

Complete reference for all configuration options in Go-LLMs, including environment variables, provider settings, agent options, and system configurations.

## Environment Variables

### Global Settings

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `GOLLMS_LOG_LEVEL` | Logging level | `info` | `debug`, `info`, `warn`, `error` |
| `GOLLMS_TIMEOUT` | Global timeout (seconds) | `300` | `600` |
| `GOLLMS_MAX_RETRIES` | Maximum retry attempts | `3` | `5` |
| `GOLLMS_RETRY_DELAY` | Initial retry delay (ms) | `1000` | `2000` |
| `GOLLMS_CACHE_DIR` | Cache directory path | `~/.gollms/cache` | `/var/cache/gollms` |
| `GOLLMS_CONFIG_PATH` | Config file location | `~/.gollms/config.yaml` | `/etc/gollms/config.yaml` |

### Provider API Keys

| Variable | Provider | Required | Example |
|----------|----------|----------|---------|
| `OPENAI_API_KEY` | OpenAI | Yes | `sk-...` |
| `ANTHROPIC_API_KEY` | Anthropic | Yes | `sk-ant-...` |
| `GOOGLE_API_KEY` | Google Gemini | Yes | `AIza...` |
| `GOOGLE_CLOUD_PROJECT` | Vertex AI | Yes | `my-project-123` |
| `OPENROUTER_API_KEY` | OpenRouter | Yes | `sk-or-...` |
| `OLLAMA_HOST` | Ollama | No | `http://localhost:11434` |

### Provider-Specific Settings

#### OpenAI
| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `OPENAI_ORGANIZATION` | Organization ID | - | `org-...` |
| `OPENAI_BASE_URL` | Custom endpoint | `https://api.openai.com` | `https://custom.openai.com` |
| `OPENAI_DEFAULT_MODEL` | Default model | `gpt-4o-mini` | `gpt-4-turbo` |

#### Anthropic
| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `ANTHROPIC_BASE_URL` | Custom endpoint | `https://api.anthropic.com` | `https://custom.anthropic.com` |
| `ANTHROPIC_DEFAULT_MODEL` | Default model | `claude-3-haiku` | `claude-3-opus` |

#### Google Gemini
| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `GOOGLE_REGION` | API region | `us-central1` | `europe-west1` |
| `GEMINI_DEFAULT_MODEL` | Default model | `gemini-1.5-flash` | `gemini-1.5-pro` |

#### Vertex AI
| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `GOOGLE_APPLICATION_CREDENTIALS` | Service account path | - | `/path/to/key.json` |
| `VERTEX_LOCATION` | Deployment location | `us-central1` | `europe-west4` |
| `VERTEX_DEFAULT_MODEL` | Default model | `gemini-1.5-flash` | `gemini-1.5-pro` |

#### OpenRouter
| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `OPENROUTER_SITE_URL` | Your site URL | - | `https://myapp.com` |
| `OPENROUTER_SITE_NAME` | Your site name | - | `My Application` |

---

## Configuration Files

### YAML Configuration

Create `~/.gollms/config.yaml`:

```yaml
# Global settings
global:
  log_level: debug
  timeout: 600
  max_retries: 5
  cache:
    enabled: true
    directory: ~/.gollms/cache
    ttl: 3600

# Provider configurations
providers:
  openai:
    api_key: ${OPENAI_API_KEY}
    organization: org-123
    default_model: gpt-4o-mini
    options:
      temperature: 0.7
      max_tokens: 2000
      
  anthropic:
    api_key: ${ANTHROPIC_API_KEY}
    default_model: claude-3-haiku
    options:
      max_tokens: 4000
      
  gemini:
    api_key: ${GOOGLE_API_KEY}
    default_model: gemini-1.5-flash
    safety_settings:
      harassment: BLOCK_MEDIUM_AND_ABOVE
      
  vertex:
    project_id: my-project
    location: us-central1
    credentials: /path/to/service-account.json
    
  ollama:
    host: http://localhost:11434
    timeout: 120
    models:
      - llama3
      - mistral
      
  openrouter:
    api_key: ${OPENROUTER_API_KEY}
    site_url: https://myapp.com
    site_name: My Application

# Agent defaults
agents:
  default_timeout: 300
  max_concurrent_tools: 10
  memory:
    type: conversation
    max_messages: 100
    
# Tool configurations
tools:
  file_operations:
    allowed_paths:
      - /home/user/data
      - /tmp
    max_file_size: 100MB
    
  web_operations:
    user_agent: "GoLLMs/1.0"
    timeout: 30
    max_redirects: 5
    
  system_operations:
    allowed_commands:
      - ls
      - grep
      - cat
    command_timeout: 60
```

### JSON Configuration

Alternative `~/.gollms/config.json`:

```json
{
  "global": {
    "log_level": "debug",
    "timeout": 600,
    "max_retries": 5
  },
  "providers": {
    "openai": {
      "api_key": "${OPENAI_API_KEY}",
      "organization": "org-123",
      "default_model": "gpt-4o-mini"
    }
  }
}
```

---

## Provider Options

### OpenAI Options

```go
provider, err := provider.NewOpenAI(provider.OpenAIOptions{
    APIKey:       "sk-...",
    Organization: "org-...",
    BaseURL:      "https://api.openai.com",
    HTTPClient:   customClient,
    RetryConfig: RetryConfig{
        MaxAttempts: 5,
        InitialDelay: time.Second,
        MaxDelay: 30 * time.Second,
    },
    DefaultOptions: map[string]interface{}{
        "temperature": 0.7,
        "max_tokens": 2000,
        "top_p": 0.9,
        "frequency_penalty": 0.0,
        "presence_penalty": 0.0,
        "logit_bias": map[string]float64{},
        "user": "user-123",
    },
})
```

### Anthropic Options

```go
provider, err := provider.NewAnthropic(provider.AnthropicOptions{
    APIKey:  "sk-ant-...",
    BaseURL: "https://api.anthropic.com",
    DefaultOptions: map[string]interface{}{
        "temperature": 0.7,
        "max_tokens": 4000,
        "top_p": 0.9,
        "top_k": 40,
        "metadata": map[string]string{
            "user_id": "123",
        },
    },
})
```

### Gemini Options

```go
provider, err := provider.NewGemini(provider.GeminiOptions{
    APIKey: "AIza...",
    SafetySettings: map[string]string{
        "HARM_CATEGORY_HARASSMENT": "BLOCK_MEDIUM_AND_ABOVE",
        "HARM_CATEGORY_HATE_SPEECH": "BLOCK_MEDIUM_AND_ABOVE",
        "HARM_CATEGORY_SEXUALLY_EXPLICIT": "BLOCK_MEDIUM_AND_ABOVE",
        "HARM_CATEGORY_DANGEROUS_CONTENT": "BLOCK_MEDIUM_AND_ABOVE",
    },
    DefaultOptions: map[string]interface{}{
        "temperature": 0.9,
        "top_p": 0.95,
        "top_k": 40,
        "max_output_tokens": 8192,
    },
})
```

### Vertex AI Options

```go
provider, err := provider.NewVertexAI(provider.VertexAIOptions{
    ProjectID:    "my-project",
    Location:     "us-central1",
    Credentials:  "/path/to/credentials.json",
    Endpoint:     "custom-endpoint.googleapis.com", // Optional
    DefaultOptions: map[string]interface{}{
        "temperature": 0.7,
        "max_output_tokens": 2048,
        "top_p": 0.95,
        "top_k": 40,
    },
})
```

### Ollama Options

```go
provider, err := provider.NewOllama(provider.OllamaOptions{
    Host:    "http://localhost:11434",
    Timeout: 120 * time.Second,
    Models: []string{
        "llama3",
        "mistral",
        "codellama",
    },
    DefaultOptions: map[string]interface{}{
        "temperature": 0.8,
        "num_predict": 128,
        "top_k": 40,
        "top_p": 0.9,
        "repeat_penalty": 1.1,
    },
})
```

### OpenRouter Options

```go
provider, err := provider.NewOpenRouter(provider.OpenRouterOptions{
    APIKey:   "sk-or-...",
    SiteURL:  "https://myapp.com",
    SiteName: "My Application",
    BaseURL:  "https://openrouter.ai/api",
    DefaultModel: "openai/gpt-3.5-turbo",
    Preferences: map[string]interface{}{
        "cost_preference": "lowest",
        "speed_preference": "fastest",
        "quality_preference": "balanced",
    },
})
```

---

## Agent Configuration

### LLM Agent Options

```go
agent := core.NewLLMAgent("assistant", provider,
    // Basic options
    core.WithModel("gpt-4o-mini"),
    core.WithSystemPrompt("You are a helpful assistant."),
    core.WithTemperature(0.7),
    core.WithMaxTokens(2000),
    
    // Tool configuration
    core.WithTools(tool1, tool2, tool3),
    core.WithMaxConcurrentTools(5),
    core.WithToolTimeout(30 * time.Second),
    
    // Memory configuration
    core.WithMemoryType(core.ConversationMemory),
    core.WithMaxMemoryMessages(100),
    core.WithMemoryPersistence("/path/to/memory.db"),
    
    // Retry configuration
    core.WithRetryAttempts(3),
    core.WithRetryDelay(time.Second),
    core.WithRetryBackoff(2.0),
    
    // Streaming configuration
    core.WithStreaming(true),
    core.WithStreamingCallback(func(chunk string) {
        fmt.Print(chunk)
    }),
    
    // Safety configuration
    core.WithSafetyLevel(core.SafetyMedium),
    core.WithContentFilters([]string{"violence", "explicit"}),
)
```

### Workflow Agent Options

```go
// Sequential workflow
workflow := workflow.NewSequentialAgent("processor",
    workflow.WithAgents(agent1, agent2, agent3),
    workflow.WithTimeout(5 * time.Minute),
    workflow.WithErrorHandling(workflow.ContinueOnError),
)

// Parallel workflow
parallel := workflow.NewParallelAgent("parallel-processor",
    workflow.WithAgents(agent1, agent2, agent3),
    workflow.WithMaxConcurrency(10),
    workflow.WithAggregationStrategy(workflow.MergeResults),
)

// Conditional workflow
conditional := workflow.NewConditionalAgent("router",
    workflow.WithConditions(map[string]workflow.Condition{
        "agent1": func(input interface{}) bool {
            return input.(map[string]interface{})["type"] == "A"
        },
        "agent2": func(input interface{}) bool {
            return input.(map[string]interface{})["type"] == "B"
        },
    }),
    workflow.WithDefaultAgent(fallbackAgent),
)
```

---

## Tool Configuration

### File Tool Settings

```go
fileTool := tools.NewFileReadTool(
    tools.WithAllowedPaths([]string{"/data", "/tmp"}),
    tools.WithMaxFileSize(100 * 1024 * 1024), // 100MB
    tools.WithFileTypes([]string{".txt", ".json", ".csv"}),
    tools.WithEncoding("utf-8"),
)
```

### Web Tool Settings

```go
webTool := tools.NewHTTPRequestTool(
    tools.WithUserAgent("GoLLMs/1.0"),
    tools.WithTimeout(30 * time.Second),
    tools.WithMaxRedirects(5),
    tools.WithRateLimit(10), // requests per second
    tools.WithRetryConfig(tools.RetryConfig{
        MaxAttempts: 3,
        InitialDelay: time.Second,
    }),
)
```

### System Tool Settings

```go
cmdTool := tools.NewCommandExecutorTool(
    tools.WithAllowedCommands([]string{"ls", "grep", "cat"}),
    tools.WithWorkingDirectory("/home/user"),
    tools.WithEnvironment(map[string]string{
        "PATH": "/usr/local/bin:/usr/bin:/bin",
    }),
    tools.WithTimeout(60 * time.Second),
    tools.WithMaxOutputSize(10 * 1024 * 1024), // 10MB
)
```

---

## Memory Configuration

### Conversation Memory

```go
memory := memory.NewConversationMemory(
    memory.WithMaxMessages(100),
    memory.WithCompression(true),
    memory.WithSummarization(provider),
    memory.WithPersistence(memory.SQLitePersistence{
        Path: "/path/to/memory.db",
    }),
)
```

### Working Memory

```go
memory := memory.NewWorkingMemory(
    memory.WithCapacity(50),
    memory.WithTTL(30 * time.Minute),
    memory.WithEvictionPolicy(memory.LRU),
)
```

### Vector Memory

```go
memory := memory.NewVectorMemory(
    memory.WithEmbeddingProvider(provider),
    memory.WithVectorDB(memory.ChromaDB{
        URL: "http://localhost:8000",
        Collection: "agent_memory",
    }),
    memory.WithSimilarityThreshold(0.8),
)
```

---

## Structured Output Configuration

### JSON Schema Validation

```go
processor := processor.NewSchemaProcessor(
    processor.WithSchema(schema),
    processor.WithValidation(processor.StrictValidation),
    processor.WithCoercion(true),
    processor.WithDefaults(true),
)
```

### Type Conversion

```go
converter := processor.NewTypeConverter(
    processor.WithDateFormat("2006-01-02"),
    processor.WithTimeFormat("15:04:05"),
    processor.WithNumberPrecision(2),
    processor.WithBooleanStrings([]string{"yes", "no", "true", "false"}),
)
```

---

## Performance Tuning

### Connection Pooling

```go
poolConfig := &PoolConfig{
    MaxConnections:     100,
    MaxIdleConnections: 10,
    IdleTimeout:        30 * time.Second,
    MaxLifetime:        5 * time.Minute,
}
```

### Cache Configuration

```go
cacheConfig := &CacheConfig{
    Type:            "redis",
    ConnectionURL:   "redis://localhost:6379",
    TTL:             time.Hour,
    MaxEntries:      10000,
    EvictionPolicy:  "lru",
    Compression:     true,
}
```

### Rate Limiting

```go
rateLimiter := &RateLimiterConfig{
    RequestsPerSecond: 100,
    BurstSize:         200,
    WaitTimeout:       30 * time.Second,
    BackoffStrategy:   "exponential",
}
```

---

## Security Configuration

### API Key Management

```go
keyManager := &KeyManagerConfig{
    Storage:         "vault",
    VaultURL:        "https://vault.example.com",
    RotationPeriod:  30 * 24 * time.Hour,
    Encryption:      true,
}
```

### SSL/TLS Settings

```go
tlsConfig := &tls.Config{
    MinVersion:               tls.VersionTLS12,
    PreferServerCipherSuites: true,
    InsecureSkipVerify:       false,
    ClientAuth:               tls.RequireAndVerifyClientCert,
}
```

### Request Signing

```go
signingConfig := &SigningConfig{
    Algorithm:   "HMAC-SHA256",
    Secret:      os.Getenv("SIGNING_SECRET"),
    HeaderName:  "X-Signature",
    IncludeTime: true,
}
```

---

## Monitoring Configuration

### Metrics

```go
metricsConfig := &MetricsConfig{
    Enabled:         true,
    Endpoint:        "/metrics",
    Port:            9090,
    IncludeRuntime:  true,
    IncludeProvider: true,
    Labels: map[string]string{
        "service": "gollms",
        "env":     "production",
    },
}
```

### Logging

```go
loggingConfig := &LoggingConfig{
    Level:      "info",
    Format:     "json",
    Output:     "stdout",
    File:       "/var/log/gollms/app.log",
    MaxSize:    100, // MB
    MaxBackups: 5,
    MaxAge:     30, // days
    Compress:   true,
}
```

### Tracing

```go
tracingConfig := &TracingConfig{
    Enabled:      true,
    Provider:     "jaeger",
    Endpoint:     "http://localhost:14268/api/traces",
    ServiceName:  "gollms",
    SampleRate:   0.1,
}
```

---

## Configuration Best Practices

### Environment Management

1. **Development Environment**
   ```bash
   # .env.development
   GOLLMS_LOG_LEVEL=debug
   GOLLMS_TIMEOUT=600
   OPENAI_API_KEY=sk-dev-...
   ```

2. **Production Environment**
   ```bash
   # .env.production
   GOLLMS_LOG_LEVEL=info
   GOLLMS_TIMEOUT=300
   OPENAI_API_KEY=${VAULT_OPENAI_KEY}
   ```

### Configuration Hierarchy

1. **Default values** (hardcoded)
2. **Configuration files** (YAML/JSON)
3. **Environment variables** (override files)
4. **Command-line arguments** (highest priority)

### Validation

```go
func ValidateConfig(cfg *Config) error {
    if cfg.Providers.OpenAI.APIKey == "" {
        return fmt.Errorf("OpenAI API key is required")
    }
    
    if cfg.Global.Timeout < 10 {
        return fmt.Errorf("timeout must be at least 10 seconds")
    }
    
    return nil
}
```

### Secret Management

- Never commit API keys to version control
- Use environment variables or secret managers
- Rotate keys regularly
- Implement key encryption at rest
- Use different keys for different environments

---

## Next Steps

- **[Error Codes Reference](error-codes-reference.md)** - Complete error handling guide
- **[Best Practices Checklist](best-practices-checklist.md)** - Production readiness
- **[Provider Setup Guide](/docs/user-guide/guides/provider-setup.md)** - Step-by-step setup
- **[Environment Variables Guide](/docs/technical/configuration/environment-variables.md)** - Detailed env var documentation
- **[Security Guide](/docs/user-guide/advanced/security-considerations.md)** - Security best practices