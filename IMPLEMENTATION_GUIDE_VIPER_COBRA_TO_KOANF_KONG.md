# Implementation Guide: Viper/Cobra to Koanf/Kong Migration

This guide provides step-by-step instructions with code examples for migrating from viper/cobra to koanf/kong.

## Step 1: Update Dependencies

Update `go.mod`:
```go
// Remove
// github.com/spf13/cobra v1.9.1
// github.com/spf13/viper v1.20.1

// Add
require (
    github.com/knadh/koanf/v2 v2.0.1
    github.com/knadh/koanf/parsers/yaml v0.1.0
    github.com/knadh/koanf/providers/env v0.1.0
    github.com/knadh/koanf/providers/file v0.1.0
    github.com/knadh/koanf/providers/basicflag v0.1.0
    github.com/alecthomas/kong v0.8.1
    github.com/willabides/kongplete v0.4.0
    github.com/posener/complete v1.2.3
)
```

## Step 2: Create New Config Structure

Create `cmd/config.go`:
```go
package main

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    
    "github.com/knadh/koanf/v2"
    "github.com/knadh/koanf/parsers/yaml"
    "github.com/knadh/koanf/providers/env"
    "github.com/knadh/koanf/providers/file"
    "github.com/knadh/koanf/providers/structs"
)

// Global koanf instance
var k = koanf.New(".")

// Config represents our application configuration
type Config struct {
    Provider string `koanf:"provider" json:"provider"`
    Model    string `koanf:"model" json:"model"`
    Verbose  bool   `koanf:"verbose" json:"verbose"`
    Output   string `koanf:"output" json:"output"`
    
    Providers struct {
        OpenAI struct {
            APIKey       string `koanf:"api_key" json:"api_key"`
            DefaultModel string `koanf:"default_model" json:"default_model"`
        } `koanf:"openai" json:"openai"`
        
        Anthropic struct {
            APIKey       string `koanf:"api_key" json:"api_key"`
            DefaultModel string `koanf:"default_model" json:"default_model"`
        } `koanf:"anthropic" json:"anthropic"`
        
        Gemini struct {
            APIKey       string `koanf:"api_key" json:"api_key"`
            DefaultModel string `koanf:"default_model" json:"default_model"`
        } `koanf:"gemini" json:"gemini"`
    } `koanf:"providers" json:"providers"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
    return Config{
        Provider: "openai",
        Output:   "text",
        Providers: struct {
            OpenAI struct {
                APIKey       string `koanf:"api_key" json:"api_key"`
                DefaultModel string `koanf:"default_model" json:"default_model"`
            } `koanf:"openai" json:"openai"`
            Anthropic struct {
                APIKey       string `koanf:"api_key" json:"api_key"`
                DefaultModel string `koanf:"default_model" json:"default_model"`
            } `koanf:"anthropic" json:"anthropic"`
            Gemini struct {
                APIKey       string `koanf:"api_key" json:"api_key"`
                DefaultModel string `koanf:"default_model" json:"default_model"`
            } `koanf:"gemini" json:"gemini"`
        }{
            OpenAI: struct {
                APIKey       string `koanf:"api_key" json:"api_key"`
                DefaultModel string `koanf:"default_model" json:"default_model"`
            }{
                DefaultModel: "gpt-4o",
            },
            Anthropic: struct {
                APIKey       string `koanf:"api_key" json:"api_key"`
                DefaultModel string `koanf:"default_model" json:"default_model"`
            }{
                DefaultModel: "claude-3-5-sonnet-latest",
            },
            Gemini: struct {
                APIKey       string `koanf:"api_key" json:"api_key"`
                DefaultModel string `koanf:"default_model" json:"default_model"`
            }{
                DefaultModel: "gemini-2.0-flash-lite",
            },
        },
    }
}

// InitConfig loads configuration from various sources
func InitConfig(configFile string) error {
    // Load defaults
    if err := k.Load(structs.Provider(DefaultConfig(), "koanf"), nil); err != nil {
        return fmt.Errorf("error loading default config: %w", err)
    }
    
    // Load from config file if specified
    if configFile != "" {
        if err := k.Load(file.Provider(configFile), yaml.Parser()); err != nil {
            fmt.Printf("Using config file: %s\n", configFile)
        }
    } else {
        // Try standard locations
        home, _ := os.UserHomeDir()
        configPaths := []string{
            filepath.Join(home, ".go-llms.yaml"),
            ".go-llms.yaml",
            filepath.Join(home, ".config", "go-llms", "config.yaml"),
        }
        
        for _, path := range configPaths {
            if _, err := os.Stat(path); err == nil {
                if err := k.Load(file.Provider(path), yaml.Parser()); err == nil {
                    fmt.Printf("Using config file: %s\n", path)
                    break
                }
            }
        }
    }
    
    // Load environment variables
    // Map MYAPP_PROVIDER to provider, MYAPP_PROVIDERS_OPENAI_API_KEY to providers.openai.api_key
    if err := k.Load(env.Provider("GO_LLMS_", ".", func(s string) string {
        return strings.Replace(
            strings.ToLower(strings.TrimPrefix(s, "GO_LLMS_")), 
            "_", ".", -1)
    }), nil); err != nil {
        return fmt.Errorf("error loading environment variables: %w", err)
    }
    
    return nil
}

// GetAPIKey retrieves the API key for a provider
func GetAPIKey(provider string) (string, error) {
    key := k.String(fmt.Sprintf("providers.%s.api_key", provider))
    if key == "" {
        // Try environment variable as fallback
        envVar := fmt.Sprintf("%s_API_KEY", strings.ToUpper(provider))
        key = os.Getenv(envVar)
        if key == "" {
            return "", fmt.Errorf("no API key configured for provider %s", provider)
        }
    }
    return key, nil
}
```

## Step 3: Create New CLI Structure with Kong

Create `cmd/cli.go`:
```go
package main

import (
    "context"
    "fmt"
    "os"
    
    "github.com/alecthomas/kong"
    "github.com/willabides/kongplete"
    "github.com/posener/complete"
)

// CLI represents our command-line interface
type CLI struct {
    Config   string `kong:"type='path',help='Config file location'"`
    Provider string `kong:"default='openai',help='LLM provider to use (openai, anthropic, gemini, mock)'"`
    Model    string `kong:"help='Model to use with the provider'"`
    Verbose  bool   `kong:"short='v',help='Enable verbose output'"`
    Output   string `kong:"short='o',default='text',enum='text,json',help='Output format'"`
    Version  bool   `kong:"short='V',help='Print version information'"`
    
    Chat       ChatCmd       `kong:"cmd,help='Interactive chat with an LLM'"`
    Complete   CompleteCmd   `kong:"cmd,help='One-shot text completion with an LLM'"`
    Agent      AgentCmd      `kong:"cmd,help='Run an agent with tools'"`
    Structured StructuredCmd `kong:"cmd,help='Extract structured data from text'"`
    Completion CompletionCmd `kong:"cmd,help='Generate shell completions'"`
    
    // For shell completion
    InstallCompletions kongplete.InstallCompletions `kong:"cmd,help='Install shell completions'"`
}

// ChatCmd represents the chat command
type ChatCmd struct {
    System     string  `kong:"short='s',help='System prompt for the chat'"`
    Temperature float32 `kong:"short='t',default='0.7',help='Temperature for generation'"`
    MaxTokens  int     `kong:"default='1000',help='Maximum tokens to generate'"`
    Stream     bool    `kong:"help='Stream responses in real-time'"`
    NoStream   bool    `kong:"help='Disable streaming'"`
}

// CompleteCmd represents the complete command
type CompleteCmd struct {
    Temperature float32  `kong:"short='t',default='0.7',help='Temperature for generation'"`
    MaxTokens   int      `kong:"default='1000',help='Maximum tokens to generate'"`
    Stream      bool     `kong:"help='Stream responses'"`
    Prompt      []string `kong:"arg,required,help='Prompt for completion'"`
}

// AgentCmd represents the agent command  
type AgentCmd struct {
    Tools       []string `kong:"name='tool',short='t',help='Enable specific tools'"`
    System      string   `kong:"short='s',help='System prompt for the agent'"`
    Schema      string   `kong:"help='Path to schema file for tool responses'"`
    Temperature float32  `kong:"default='0.7',help='Temperature for generation'"`
    Prompt      []string `kong:"arg,required,help='Initial prompt for the agent'"`
}

// StructuredCmd represents the structured command
type StructuredCmd struct {
    Schema      string  `kong:"name='schema',short='s',help='Path to JSON schema file',type='existingfile'"`
    OutputFile  string  `kong:"short='o',help='Output file for extracted data',type='path'"`
    Validate    bool    `kong:"name='validate',help='Validate against schema without LLM'"`
    Temperature float32 `kong:"short='t',default='0.7',help='Temperature for generation'"`
    MaxTokens   int     `kong:"default='1000',help='Maximum tokens to generate'"`
    Prompt      string  `kong:"arg,required,help='Prompt or file containing prompt'"`
}

// CompletionCmd represents the completion command
type CompletionCmd struct {
    Shell string `kong:"arg,required,enum='bash,zsh,fish,powershell',help='Shell to generate completions for'"`
}

// Run implements the chat command
func (c *ChatCmd) Run(ctx *Context) error {
    // Implementation similar to original newChatCmd
    if c.Stream && c.NoStream {
        return fmt.Errorf("cannot specify both --stream and --no-stream")
    }
    
    // Get provider configuration
    provider, model, err := ctx.GetProvider()
    if err != nil {
        return err
    }
    
    // Original chat implementation...
    return nil
}

// Context holds shared state for commands
type Context struct {
    Config *Config
    CLI    *CLI
}

// GetProvider returns the configured provider and model
func (ctx *Context) GetProvider() (string, string, error) {
    provider := ctx.CLI.Provider
    if provider == "" {
        provider = ctx.Config.Provider
    }
    
    model := ctx.CLI.Model
    if model == "" {
        model = k.String(fmt.Sprintf("providers.%s.default_model", provider))
        if model == "" {
            return "", "", fmt.Errorf("no model specified for provider %s", provider)
        }
    }
    
    return provider, model, nil
}
```

## Step 4: Update main.go

Replace the current `main.go`:
```go
package main

import (
    "context"
    "fmt"
    "os"
    
    "github.com/alecthomas/kong"
    "github.com/willabides/kongplete"
)

var version = "v0.2.0"

func main() {
    // Parse CLI arguments
    cli := CLI{}
    parser := kong.Must(&cli,
        kong.Name("go-llms"),
        kong.Description("A Go library for LLM-powered applications"),
        kong.UsageOnError(),
        kong.ConfigureHelp(kong.HelpOptions{
            Compact: true,
        }),
        kong.Vars{
            "version": version,
        },
    )
    
    // Setup shell completion
    kongplete.Complete(parser)
    
    // Parse arguments
    ctx, err := parser.Parse(os.Args[1:])
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    
    // Handle version flag
    if cli.Version {
        fmt.Printf("go-llms %s\n", version)
        os.Exit(0)
    }
    
    // Initialize configuration
    if err := InitConfig(cli.Config); err != nil {
        fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
        os.Exit(1)
    }
    
    // Load configuration into struct
    var config Config
    if err := k.Unmarshal("", &config); err != nil {
        fmt.Fprintf(os.Stderr, "Error parsing config: %v\n", err)
        os.Exit(1)
    }
    
    // Create context
    cmdCtx := &Context{
        Config: &config,
        CLI:    &cli,
    }
    
    // Execute command
    if err := ctx.Run(cmdCtx); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

## Step 5: Update Tests

Update `cmd/main_test.go`:
```go
package main

import (
    "os"
    "testing"
    
    "github.com/knadh/koanf/v2"
    "github.com/knadh/koanf/parsers/yaml"
    "github.com/knadh/koanf/providers/structs"
)

func TestGetAPIKey(t *testing.T) {
    // Save old environment variables
    oldOpenAIKey := os.Getenv("OPENAI_API_KEY")
    oldAnthropicKey := os.Getenv("ANTHROPIC_API_KEY")
    
    // Cleanup function
    defer func() {
        os.Setenv("OPENAI_API_KEY", oldOpenAIKey)
        os.Setenv("ANTHROPIC_API_KEY", oldAnthropicKey)
        k = koanf.New(".") // Reset koanf instance
    }()
    
    // Clear environment variables
    os.Unsetenv("OPENAI_API_KEY")
    os.Unsetenv("ANTHROPIC_API_KEY")
    
    // Test with environment variable
    t.Run("GetFromEnvOpenAI", func(t *testing.T) {
        os.Setenv("OPENAI_API_KEY", "test-openai-key")
        
        key, err := GetAPIKey("openai")
        if err != nil {
            t.Fatalf("Expected no error, got: %v", err)
        }
        
        if key != "test-openai-key" {
            t.Errorf("Expected 'test-openai-key', got: %s", key)
        }
    })
    
    // Test with koanf configuration
    t.Run("GetFromConfig", func(t *testing.T) {
        // Reset koanf and set config value
        k = koanf.New(".")
        config := DefaultConfig()
        config.Providers.OpenAI.APIKey = "config-api-key"
        
        k.Load(structs.Provider(config, "koanf"), nil)
        
        key, err := GetAPIKey("openai")
        if err != nil {
            t.Fatalf("Expected no error, got: %v", err)
        }
        
        if key != "config-api-key" {
            t.Errorf("Expected 'config-api-key', got: %s", key)
        }
    })
}
```

## Step 6: Shell Completion Script

Create `scripts/install-completion.sh`:
```bash
#!/bin/bash
# Install shell completions for go-llms

case "$SHELL" in
  */bash)
    if [[ -d ~/.bash_completion.d ]]; then
      go-llms completion bash > ~/.bash_completion.d/go-llms
      echo "Bash completion installed to ~/.bash_completion.d/go-llms"
    else
      echo "Run: go-llms completion bash >> ~/.bashrc"
    fi
    ;;
  */zsh)
    if [[ -d ~/.zsh/completions ]]; then
      go-llms completion zsh > ~/.zsh/completions/_go-llms
      echo "Zsh completion installed to ~/.zsh/completions/_go-llms"
    else
      echo "Run: go-llms completion zsh >> ~/.zshrc"
    fi
    ;;
  */fish)
    if [[ -d ~/.config/fish/completions ]]; then
      go-llms completion fish > ~/.config/fish/completions/go-llms.fish
      echo "Fish completion installed"
    else
      echo "Run: go-llms completion fish > ~/.config/fish/completions/go-llms.fish"
    fi
    ;;
  *)
    echo "Shell not recognized. Run 'go-llms completion <shell>' manually"
    ;;
esac
```

## Migration Checklist

1. [ ] Update go.mod with new dependencies
2. [ ] Create config.go with koanf implementation
3. [ ] Create cli.go with kong structures
4. [ ] Update main.go to use new structure
5. [ ] Implement all command Run methods
6. [ ] Update tests to use koanf
7. [ ] Test all commands
8. [ ] Test configuration loading
9. [ ] Test shell completions
10. [ ] Update documentation
11. [ ] Create migration notes for users

## Common Issues and Solutions

### Issue 1: Flag Precedence
Koanf loads in order, so ensure CLI flags override config file values:
```go
// Load in this order:
// 1. Defaults
// 2. Config file  
// 3. Environment variables
// 4. CLI flags (highest priority)
```

### Issue 2: Command Context
Kong uses Run methods instead of cobra's anonymous functions:
```go
// Convert cobra command functions to kong Run methods
func (c *ChatCmd) Run(ctx *Context) error {
    // Implementation
}
```

### Issue 3: Shell Completion
Kongplete requires the complete package:
```go
import (
    "github.com/willabides/kongplete"
    "github.com/posener/complete"
)
```

## Testing the Migration

1. Build the application
2. Test each command
3. Test configuration loading
4. Test environment variables
5. Test shell completions
6. Run existing tests