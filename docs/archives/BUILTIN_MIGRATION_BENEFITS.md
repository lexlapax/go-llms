# Built-in Components Migration Benefits

## Comparison: common_tools.go vs Built-in Tools

### Before (common_tools.go)
```go
// Direct instantiation only
tool := tools.WebFetch()

// No discovery mechanism
// No metadata about the tool
// No resource usage information
// No versioning
```

### After (Built-in Tools)
```go
// Multiple ways to access
tool := tools.MustGetTool("web_fetch")         // By name
webTools := tools.Tools.ListByCategory("web")  // By category
netTools := tools.Tools.Search("network")       // By search

// Rich metadata
entry, _ := tools.Tools.Get("web_fetch")
fmt.Println(entry.Metadata.Description)
fmt.Println(entry.Metadata.Version)
fmt.Println(entry.Metadata.Tags)
fmt.Println(entry.Metadata.Examples)

// Resource usage information
// - Memory requirements
// - Network usage
// - File system access
// - Concurrency safety

// Permission tracking
// Tools declare what permissions they need
```

## Key Benefits

### 1. **Discoverability**
- List all available tools
- Filter by category or tags
- Search by name or description
- No need to know exact import paths

### 2. **Metadata and Documentation**
- Every tool has description and examples
- Version tracking for compatibility
- Deprecation warnings
- Resource usage declarations

### 3. **Registry Pattern**
- Consistent registration mechanism
- Auto-registration on import
- Central access point
- Extensible for custom tools

### 4. **Enhanced Functionality**
During migration, tools are enhanced with:
- Better error handling
- Context awareness
- Resource limits
- Security considerations

### 5. **Type Safety**
- Structured parameter schemas
- Typed result objects
- Compile-time safety where possible

## Migration Path

### Phase 1: Coexistence
- Both common_tools.go and built-ins available
- Gradual migration of tools
- Update examples to use built-ins

### Phase 2: Deprecation
- Mark common_tools.go as deprecated
- Provide migration guide
- Update all documentation

### Phase 3: Removal
- Remove common_tools.go in next major version
- All tools available through registry
- Breaking change documented

## Example: WebFetch Migration

### Old Implementation
```go
func WebFetch() domain.Tool {
    return NewTool(
        "web_fetch",
        "Fetches content from a URL",
        func(ctx context.Context, params WebFetchParams) (*WebFetchResult, error) {
            // Basic implementation
            // Fixed 30-second timeout
            // No header capture
            // No resource tracking
        },
        WebFetchParamSchema,
    )
}
```

### New Implementation
```go
func init() {
    tools.MustRegisterTool("web_fetch", WebFetch(), tools.ToolMetadata{
        Metadata: builtins.Metadata{
            Name:        "web_fetch",
            Category:    "web",
            Tags:        []string{"http", "fetch", "download", "web", "network"},
            Description: "Fetches content from a URL with customizable timeout",
            Version:     "1.0.0",
            Examples:    []builtins.Example{...},
        },
        RequiredPermissions: []string{"network:access"},
        ResourceUsage: tools.ResourceInfo{
            Memory:      "low",
            Network:     true,
            FileSystem:  false,
            Concurrency: true,
        },
    })
}

func WebFetch() domain.Tool {
    // Enhanced implementation with:
    // - Customizable timeout
    // - Header capture
    // - User-agent identification
    // - Better error messages
}
```

## For Library Users

The migration provides:
1. Better tool discovery
2. Richer documentation
3. Resource awareness
4. Version compatibility
5. Consistent patterns

## For Contributors

The new structure enables:
1. Clear contribution guidelines
2. Consistent tool patterns
3. Automated registration
4. Rich metadata requirements
5. Testability improvements