# Viper and Cobra Usage Analysis

## Overview
The viper and cobra libraries are only used in the main CLI application (`cmd/main.go` and `cmd/main_test.go`). The rest of the codebase does not use these libraries.

## Cobra Usage

### Features Used:
1. **Command Structure**:
   - Root command with versioning
   - Subcommands: `chat`, `complete`, `agent`, `structured`, `completion`
   - Command descriptions and long descriptions
   - Argument validation (`Args: cobra.MinimumNArgs(1)`, `Args: cobra.MatchAll(...)`)

2. **Flags**:
   - Persistent flags on root command:
     - `--config` (string)
     - `--provider` (string, default: "openai")
     - `--model` (string)
     - `--verbose` / `-v` (bool)
     - `--output` / `-o` (string, default: "text")
   - Command-specific flags:
     - chat: `--system`, `--temperature`, `--max-tokens`, `--stream`, `--no-stream`
     - complete: `--temperature`, `--max-tokens`, `--stream`
     - agent: `--tools`, `--system`, `--schema`, `--temperature`
     - structured: `--schema`, `--output-file`, `--validate`, `--temperature`, `--max-tokens`
   - Flag shortcuts using `cobra.Flags().StringP()`, `cobra.Flags().Float32P()`, etc.
   - Mutually exclusive flags: `cmd.MarkFlagsMutuallyExclusive("stream", "no-stream")`

3. **Shell Completion**:
   - `GenBashCompletion()`
   - `GenZshCompletion()`
   - `GenFishCompletion()`
   - `GenPowerShellCompletionWithDesc()`

4. **Initialization**:
   - `cobra.OnInitialize(initConfig)`
   - `cobra.CheckErr(err)`

5. **Command Execution**:
   - `rootCmd.Execute()`

### API Calls:
```go
// Command creation
cobra.Command{}
rootCmd.AddCommand()

// Flag management
cmd.PersistentFlags().StringVar()
cmd.Flags().StringP()
cmd.Flags().GetString()
cmd.Flags().GetFloat32()
cmd.Flags().GetBool()
cmd.Flags().Changed()
cmd.MarkFlagsMutuallyExclusive()

// Completion
cmd.Root().GenBashCompletion()
cmd.Root().GenZshCompletion()
cmd.Root().GenFishCompletion()
cmd.Root().GenPowerShellCompletionWithDesc()

// Validation
cobra.MinimumNArgs()
cobra.ExactArgs()
cobra.OnlyValidArgs()
cobra.MatchAll()

// Error handling
cobra.CheckErr()

// Initialization
cobra.OnInitialize()
```

## Viper Usage

### Features Used:
1. **Configuration Management**:
   - Config file support (`.go-llms.yaml` in home directory or `.config/go-llms/config.yaml`)
   - Multiple config search paths
   - Environment variable support
   - Setting defaults
   - Reading config file

2. **Flag Binding**:
   - Binding cobra flags to viper configuration keys
   - Automatic precedence (flags > env vars > config file > defaults)

3. **Configuration Reading**:
   - Getting string values with `viper.GetString()`
   - Hierarchical config access (e.g., `providers.openai.default_model`)
   - Setting default values

4. **Config Search Paths**:
   - Home directory
   - Current directory
   - `~/.config/go-llms/`

### API Calls:
```go
// Config file management
viper.SetConfigFile()
viper.AddConfigPath()
viper.SetConfigType()
viper.SetConfigName()
viper.ReadInConfig()
viper.ConfigFileUsed()

// Environment variables
viper.AutomaticEnv()

// Defaults
viper.SetDefault()

// Flag binding
viper.BindPFlag()

// Getting values
viper.GetString()

// Setting values (in tests)
viper.Set()
viper.Reset()
```

## Config Structure Expected by Viper

Based on the code, the expected configuration structure is:

```yaml
provider: openai  # LLM provider to use
model: gpt-4o    # Model to use
verbose: false   # Enable verbose output
output: text     # Output format

providers:
  openai:
    api_key: your-api-key
    default_model: gpt-4o
    
  anthropic:
    api_key: your-api-key
    default_model: claude-3-5-sonnet-latest
    
  gemini:
    api_key: your-api-key
    default_model: gemini-2.0-flash-lite
```

## Tests Dependent on Viper/Cobra

### `cmd/main_test.go`:
1. `TestGetAPIKey`:
   - Tests viper's config management
   - Tests environment variable fallback
   - Uses `viper.Set()` and `viper.Reset()`

2. `TestChatCmdStreamingFlag`:
   - Tests cobra's flag parsing
   - Verifies flag default values
   - Tests explicit flag setting

## Configuration Precedence

The current implementation follows this precedence order:
1. Command-line flags (highest priority)
2. Environment variables (e.g., `OPENAI_API_KEY`)
3. Config file
4. Default values (lowest priority)

## Files Using Viper/Cobra

1. `/Users/spuri/projects/lapaxworks/go-llms/cmd/main.go`:
   - Full CLI implementation with all commands
   - Viper configuration management
   - Cobra command structure

2. `/Users/spuri/projects/lapaxworks/go-llms/cmd/main_test.go`:
   - Tests for API key retrieval
   - Tests for command flags
   - Uses viper for test configuration

## Key Integration Points

1. **initConfig()** function:
   - Called by cobra initialization
   - Sets up viper configuration
   - Binds flags to config

2. **getProvider()** function:
   - Uses viper to get provider and model
   - Implements fallback to default models

3. **getAPIKey()** function:
   - Uses viper for config values
   - Falls back to environment variables

4. **Flag binding** in init():
   - Binds persistent flags to viper keys
   - Enables configuration value override via CLI