# Dependency Reduction Results

## Size Comparison
- **Current build** (koanf/kong/kongplete): 9.9MB
- **Optimized build** (stdlib + yaml.v3): 6.3MB (-36%)
- **Minimal build** (CLI only): 1.6MB (baseline)

## Analysis

### Current Dependencies (cmd/)
1. **koanf/v2** + subpackages - Configuration management
2. **kong** - CLI parsing
3. **kongplete** - Shell completion
4. **yaml.v3** - YAML parsing (indirect via koanf)
5. **mapstructure** - Indirect via koanf

### Optimized Approach
By replacing koanf/kong with stdlib alternatives:
- Use **flag** package for CLI parsing
- Use **yaml.v3** directly for config files
- Manual environment variable handling
- No shell completion (or basic manual implementation)

## Code Changes Required

### 1. Remove Koanf (Biggest Impact)
Replace with:
```go
// Direct YAML unmarshaling
data, _ := os.ReadFile(configFile)
yaml.Unmarshal(data, &config)

// Manual env var loading
if val := os.Getenv("GO_LLMS_PROVIDER"); val != "" {
    config.Provider = val
}
```

### 2. Remove Kong/Kongplete
Replace with:
```go
// stdlib flag package
providerFlag := flag.String("provider", "openai", "LLM provider")
modelFlag := flag.String("model", "", "Model to use")
flag.Parse()

// Manual command routing
switch flag.Arg(0) {
case "chat":
    runChat()
case "complete":
    runComplete()
}
```

## Trade-offs

### Pros of Optimization
- 36% binary size reduction (3.6MB saved)
- Fewer dependencies to manage
- Simpler code for basic needs
- Faster compile times

### Cons of Optimization
- Loss of shell completion
- Less flexible configuration system
- Manual command routing
- No automatic help generation
- More code to maintain for env vars

## Recommendation

**Keep current setup if:**
- 9.9MB binary size is acceptable
- You value shell completion
- You need flexible config management
- You want structured CLI with automatic help

**Optimize if:**
- Binary size is critical
- You can sacrifice some CLI features
- Simple config management is sufficient
- You prefer minimal dependencies

## Implementation Path

If you decide to optimize:

1. **Phase 1**: Remove koanf (easiest, biggest impact)
   - Replace with direct YAML + env handling
   - Keep kong for CLI
   - Estimated reduction: ~2MB

2. **Phase 2**: Remove kongplete
   - Remove shell completion or implement basic version
   - Keep kong for CLI structure
   - Additional reduction: ~0.5MB

3. **Phase 3**: Replace kong with stdlib
   - Use flag package
   - Manual command routing
   - Additional reduction: ~1MB

Total potential reduction: ~3.6MB (36%)