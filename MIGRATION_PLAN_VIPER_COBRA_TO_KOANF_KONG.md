# Migration Plan: Viper/Cobra to Koanf/Kong

## Overview
This document outlines the migration strategy from Spf13's viper/cobra to knadh/koanf and alecthomas/kong/kongplet.

## Current State Analysis

### Viper Usage
- **Config file format**: YAML
- **Config paths**: `~/.go-llms.yaml`, `./.go-llms.yaml`, `~/.config/go-llms/config.yaml`
- **Features used**:
  - YAML config loading
  - Environment variable binding
  - Default values
  - Flag binding with cobra
  - Hierarchical configuration

### Cobra Usage
- **Commands**: Root + 5 subcommands (chat, complete, agent, structured, completion)
- **Features used**:
  - Persistent and local flags
  - Flag shortcuts
  - Shell completion
  - Argument validation
  - Mutually exclusive flags

## Migration Strategy

### Phase 1: Viper to Koanf

#### Key Mappings
```go
// Viper -> Koanf

// Config setup
viper.SetConfigFile()         -> koanf.Load(file.Provider())
viper.AddConfigPath()         -> Built into file.Provider()
viper.SetConfigType()         -> Built into file.Provider()
viper.AutomaticEnv()          -> koanf.Load(env.Provider())
viper.SetDefault()            -> koanf.Load(structs.Provider())

// Getting values
viper.GetString()             -> k.String()
viper.GetBool()              -> k.Bool()
viper.GetInt()               -> k.Int()

// Flag binding
viper.BindPFlag()            -> koanf.Load(basicflag.Provider())
```

#### Implementation Steps
1. Replace viper imports with koanf
2. Create koanf config struct for defaults
3. Update config loading logic
4. Replace value getters
5. Update flag binding mechanism

### Phase 2: Cobra to Kong/Kongplet

#### Key Mappings
```go
// Cobra -> Kong

// Command structure
&cobra.Command{}             -> struct with kong tags
cmd.Use                      -> `kong:"cmd"`
cmd.Short                    -> `kong:"help"`
cmd.Run                      -> Run() method

// Flags
cmd.PersistentFlags()        -> Embed in parent struct
cmd.Flags()                  -> Struct fields with kong tags
GetString/GetBool/etc        -> Direct struct field access

// Validation
cobra.MinimumNArgs()         -> `kong:"arg,min=1"`
cobra.ExactArgs()            -> `kong:"arg,count=1"`

// Completion
GenBashCompletion()          -> kongplet generation
```

#### Implementation Steps
1. Define CLI struct hierarchy
2. Convert commands to structs
3. Replace flag definitions with struct tags
4. Implement Run() methods
5. Set up kongplet for shell completion

## Code Structure Changes

### Before (Cobra/Viper)
```go
// main.go
var rootCmd = &cobra.Command{
    Use: "go-llms",
    Run: func(cmd *cobra.Command, args []string) {
        provider := viper.GetString("provider")
        // ...
    },
}

func init() {
    rootCmd.PersistentFlags().String("provider", "openai", "LLM provider")
    viper.BindPFlag("provider", rootCmd.PersistentFlags().Lookup("provider"))
}
```

### After (Kong/Koanf)
```go
// main.go
type CLI struct {
    Provider string `kong:"default='openai',help='LLM provider'"`
    
    Chat       ChatCmd       `kong:"cmd,help='Interactive chat with an LLM'"`
    Complete   CompleteCmd   `kong:"cmd,help='One-shot text completion'"`
    // ...
}

type ChatCmd struct {
    System      string  `kong:"short='s',help='System prompt'"`
    Temperature float32 `kong:"short='t',default='0.7',help='Temperature'"`
}

func (c *ChatCmd) Run(ctx *Context) error {
    // Implementation
}
```

## Dependencies to Add

```go
// go.mod additions
require (
    github.com/knadh/koanf/v2 v2.0.1
    github.com/alecthomas/kong v0.8.1
    github.com/willabides/kongplete v0.4.0
)
```

## Testing Considerations

1. Update `main_test.go` to use koanf instead of viper
2. Add kong parser tests
3. Test shell completion with kongplet
4. Ensure backward compatibility with config files

## Benefits of Migration

1. **Better type safety**: Kong uses struct tags for type-safe CLI parsing
2. **Simpler configuration**: Koanf has a cleaner API than viper
3. **Better performance**: Both libraries are more lightweight
4. **Modern APIs**: More idiomatic Go code
5. **Better documentation**: Kong's struct-based approach is self-documenting

## Risk Mitigation

1. Keep existing config file format (YAML)
2. Maintain same CLI interface for users
3. Test thoroughly with existing config files
4. Create migration guide for users
5. Update all documentation

## Implementation Timeline

1. Week 1: Set up koanf, migrate config loading
2. Week 2: Convert CLI structure to Kong
3. Week 3: Implement shell completion with kongplet
4. Week 4: Testing, documentation, and cleanup

## Before/After Comparison

### Before Migration (Viper/Cobra)
- **Binary Size**: 11MB
- **Go Version**: go1.24.3 darwin/arm64
- **Build Date**: May 16, 2025
- **Dependencies**:
  ```
  github.com/spf13/cobra v1.9.1
  github.com/spf13/viper v1.20.1
  github.com/go-viper/mapstructure/v2 v2.2.1
  github.com/spf13/pflag v1.0.5
  github.com/fsnotify/fsnotify v1.8.0
  ```
- **Total Dependencies**: 5 (direct) + many transitive dependencies
- **Key Features**: 
  - Command-line parsing with cobra
  - Configuration management with viper
  - Shell completions via cobra

### After Migration (Koanf/Kong) - COMPLETED
- **Binary Size**: 14MB (increased from 11MB)
- **Go Version**: go1.24.3 darwin/arm64
- **Build Date**: May 16, 2025
- **Dependencies**:
  ```
  github.com/alecthomas/kong v1.11.0
  github.com/knadh/koanf/maps v0.1.2
  github.com/knadh/koanf/parsers/yaml v0.1.0
  github.com/knadh/koanf/providers/env v0.1.0
  github.com/knadh/koanf/providers/file v0.1.0
  github.com/knadh/koanf/providers/structs v0.1.0
  github.com/knadh/koanf/v2 v2.2.0
  github.com/willabides/kongplete v0.4.0
  ```
- **Total Dependencies**: 8 (direct koanf/kong related) but fewer transitive dependencies
- **Key Features**:
  - Struct-based command-line parsing with kong
  - Flexible configuration management with koanf
  - Shell completions via kongplete
  
### Actual Results
- **Binary size**: Increased slightly (14MB vs 11MB) 
- **Code structure**: Significantly improved with type-safe structs
- **Configuration**: More flexible provider system
- **CLI parsing**: Self-documenting through struct tags
- **Maintenance**: Better organized with clear separation of concerns
- **Testing**: Easier to test with struct-based approach

### Migration Achievements
1. ✅ Complete removal of viper and cobra dependencies
2. ✅ Backward compatibility with existing YAML config files
3. ✅ Maintained same CLI interface for users
4. ✅ All unit tests passing
5. ✅ Build system functional with updated Makefile
6. ✅ Performance benchmarks show no regression

### Code Structure Improvements
- Replaced command functions with struct methods
- Type-safe access to configuration values
- Better error handling and validation
- Cleaner separation between CLI and config logic

## Notes

- Consider creating a config migration tool for users
- Shell completion might need custom implementation for some features
- Kong's struct-based approach will make code more maintainable
- Koanf's provider system is more flexible than viper's approach