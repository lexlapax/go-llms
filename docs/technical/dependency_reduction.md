# Dependency Reduction Journey

## Overview

This document chronicles our journey to reduce dependencies and binary size in the go-llms project, from viper/cobra through koanf/kong to a minimal stdlib-based solution.

## The Journey Timeline

### Stage 1: Original Architecture (Viper/Cobra)
**Binary Size**: 11MB  
**Key Dependencies**:
- `github.com/spf13/cobra` - Command-line interface framework
- `github.com/spf13/viper` - Configuration management
- Plus many transitive dependencies

### Stage 2: Migration to Koanf/Kong
**Binary Size**: 14MB (increased from 11MB)  
**Key Dependencies**:
- `github.com/knadh/koanf/v2` - Configuration management
- `github.com/alecthomas/kong` - CLI parsing
- `github.com/willabides/kongplete` - Shell completion

### Stage 3: Optimization with Stdlib
**Binary Size**: 6.3MB (36% reduction from 9.9MB)  
**Key Dependencies**:
- `gopkg.in/yaml.v3` - YAML parsing only
- Everything else from Go stdlib

## Motivation

The initial migration from viper/cobra to koanf/kong was intended to:
1. Reduce dependencies
2. Improve type safety
3. Modernize the codebase

However, the migration actually increased binary size from 11MB to 14MB, prompting a deeper analysis of what could be done to truly reduce dependencies.

## Migration Phase 1: Viper/Cobra → Koanf/Kong

### Original State (Viper/Cobra)
```go
// Command structure with Cobra
var rootCmd = &cobra.Command{
    Use: "go-llms",
    Run: func(cmd *cobra.Command, args []string) {
        provider := viper.GetString("provider")
        // ...
    },
}

// Configuration with Viper
func initConfig() {
    viper.SetConfigFile(cfgFile)
    viper.AutomaticEnv()
    viper.ReadInConfig()
}
```

### Migrated State (Koanf/Kong)
```go
// Command structure with Kong
type CLI struct {
    Provider string `kong:"default='openai',help='LLM provider'"`
    Chat     ChatCmd `kong:"cmd,help='Interactive chat with an LLM'"`
}

// Configuration with Koanf
func InitConfig(configFile string) error {
    k := koanf.New(".")
    k.Load(file.Provider(configFile), yaml.Parser())
    k.Load(env.Provider("GO_LLMS_", ".", nil))
    return nil
}
```

### Migration Plan Features
1. Type-safe CLI parsing with struct tags
2. Better configuration provider system
3. Maintained backward compatibility
4. Shell completion through kongplete

### Results
- ✅ Successfully removed viper and cobra
- ✅ Improved code structure with type safety
- ❌ Binary size increased (11MB → 14MB)
- ❌ Still many dependencies

## Analysis Phase: Understanding the Problem

After the migration, we conducted a detailed analysis to understand why the binary size increased and what could be done.

### Build Size Tests
```bash
# Minimal build (stdlib only)
$ make build-minimal
Binary size: 1.6MB

# Current build (with koanf/kong)
$ make build
Binary size: 9.9MB

# Optimized build (stdlib + yaml)
$ make build-optimized
Binary size: 6.3MB
```

### Key Findings
1. Koanf brought in many indirect dependencies:
   - `mapstructure` for configuration mapping
   - `mitchellh/copystructure` for deep copying
   - Multiple provider packages

2. Kong and kongplete added:
   - Shell completion dependencies
   - Struct tag parsing overhead
   - Additional validation logic

## Migration Phase 2: Koanf/Kong → Stdlib

### The Decision

Based on the analysis, we decided to implement the "optimized approach":
- Replace koanf with direct YAML parsing + manual env var handling
- Replace kong/kongplete with stdlib flag package
- Keep only yaml.v3 as an external dependency

### Implementation Details

#### Replacing Koanf

**Before (Koanf)**:
```go
func InitConfig(configFile string) error {
    k := koanf.New(".")
    
    // Load defaults
    k.Load(structs.Provider(DefaultConfig(), "koanf"), nil)
    
    // Load config file
    k.Load(file.Provider(configFile), yaml.Parser())
    
    // Load environment variables
    k.Load(env.Provider("GO_LLMS_", ".", nil))
    
    return nil
}
```

**After (Direct YAML)**:
```go
func InitOptimizedConfig(configFile string) error {
    if configFile != "" {
        data, err := os.ReadFile(configFile)
        if err != nil {
            return err
        }
        return yaml.Unmarshal(data, &globalConfig)
    }
    
    // Manual environment variable loading
    if provider := os.Getenv("GO_LLMS_PROVIDER"); provider != "" {
        globalConfig.Provider = provider
    }
    
    return nil
}
```

#### Replacing Kong

**Before (Kong)**:
```go
type CLI struct {
    Provider string   `kong:"default='openai',help='LLM provider'"`
    Chat     ChatCmd  `kong:"cmd,help='Interactive chat'"`
    Complete CompleteCmd `kong:"cmd,help='Text completion'"`
}

func main() {
    cli := CLI{}
    ctx := kong.Parse(&cli)
    err := ctx.Run(&cli)
}
```

**After (Stdlib flag)**:
```go
func main() {
    providerFlag := flag.String("p", "", "LLM provider")
    modelFlag := flag.String("m", "", "Model to use")
    configFile := flag.String("c", "", "Config file")
    
    flag.Parse()
    
    command := flag.Arg(0)
    switch command {
    case "chat":
        runChat()
    case "complete":
        runComplete()
    default:
        printUsage()
    }
}
```

### Trade-offs Made

#### Features Lost
1. **Shell Completion**: Removed kongplete dependency
2. **Advanced Config Providers**: No more koanf provider system
3. **Type-safe CLI Parsing**: Lost Kong's struct-based approach
4. **Automatic Help Generation**: Manual help text now

#### Benefits Gained
1. **36% Binary Size Reduction**: 9.9MB → 6.3MB
2. **Fewer Dependencies**: Only 4 direct dependencies remaining
3. **Faster Build Times**: Fewer packages to compile
4. **Simpler Code**: Direct, obvious implementation

## Results Summary

### Binary Size Evolution
1. **Viper/Cobra**: 11MB
2. **Koanf/Kong**: 14MB (increased)
3. **Final (Optimized)**: 6.3MB (43% reduction from peak)

### Dependencies Removed
```
github.com/alecthomas/kong
github.com/willabides/kongplete
github.com/knadh/koanf/v2
github.com/knadh/koanf/parsers/yaml
github.com/knadh/koanf/providers/env
github.com/knadh/koanf/providers/file
github.com/knadh/koanf/providers/structs
github.com/go-viper/mapstructure/v2 (indirect)
github.com/mitchellh/copystructure (indirect)
github.com/fatih/structs (indirect)
```

### Dependencies Kept
```
github.com/google/uuid       # UUID generation
github.com/json-iterator/go  # Fast JSON parsing
github.com/stretchr/testify  # Testing utilities
gopkg.in/yaml.v3            # YAML parsing
```

## Lessons Learned

1. **Measure First**: The initial migration increased binary size despite expectations
2. **Analyze Dependencies**: Understanding indirect dependencies is crucial
3. **Consider Trade-offs**: Feature loss vs. size reduction needs careful balance
4. **Stdlib is Powerful**: Go's standard library can handle most CLI needs
5. **Keep What Matters**: We kept yaml.v3 because it provides real value

## Recommendations for Others

1. **Start with Analysis**: Build size tests before making changes
2. **Map Dependencies**: Use `go mod graph` to understand the full tree
3. **Consider Your Users**: What features actually matter to them?
4. **Test Thoroughly**: Ensure backward compatibility during migration
5. **Document Trade-offs**: Be clear about what's lost and gained

## Future Possibilities

1. **Shell Completion**: Could be reimplemented manually if needed
2. **Further Optimization**: Replace json-iterator with encoding/json
3. **Feature Restoration**: Advanced commands could be added back carefully
4. **Modular Approach**: Optional features as separate binaries

## Conclusion

Our journey from viper/cobra through koanf/kong to a minimal stdlib-based solution demonstrates that dependency reduction requires careful analysis and willingness to make trade-offs. The final 36% reduction in binary size validates the approach, though we sacrificed some developer conveniences for the benefit of a leaner, more maintainable codebase.