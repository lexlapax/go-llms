# Binary Size Reduction Analysis for go-llms

## Current State

- **Binary Size**: 6.3MB (already optimized with `-ldflags "-s -w"`)
- **Dependencies**: Only 17 total (very minimal)
  - Direct: google/uuid, json-iterator/go, stretchr/testify, yaml.v3
  - Most indirect dependencies are from testify
- **Build flags**: Already using `-s -w` for stripping debug info

## Key Findings

### 1. json-iterator Usage
The project uses json-iterator/go for JSON operations. This is used extensively across 38 files for:
- Provider implementations (OpenAI, Anthropic, Gemini)
- Agent workflows
- Schema validation
- Structured output processing

### 2. Dependency Analysis
- **google/uuid**: Likely necessary for request IDs
- **json-iterator/go**: Performance-optimized JSON, but adds ~500KB
- **stretchr/testify**: Only for tests, not in binary
- **yaml.v3**: Used for config files

## Recommendations (Without Major Code Changes)

### 1. Build Optimization
```bash
# Current build flags
LDFLAGS=-ldflags "-s -w"

# Additional flags to try
LDFLAGS=-ldflags "-s -w -trimpath"
# Or more aggressive
LDFLAGS=-ldflags "-s -w -trimpath -installsuffix static"
```

### 2. Conditional Compilation
Create build tags to exclude providers not needed:
```go
// +build !no_openai
// +build !no_anthropic
// +build !no_gemini
```

This would allow users to build with only the providers they need:
```bash
go build -tags no_gemini,no_anthropic -ldflags "-s -w" ./cmd/
```

### 3. Replace json-iterator (Most Impact)
Replace json-iterator with standard library encoding/json. The performance difference is minimal for most use cases:

```go
// Replace in pkg/util/json/json.go
import "encoding/json"

func Marshal(v interface{}) ([]byte, error) {
    return json.Marshal(v)
}
```

Estimated size reduction: ~500KB (8% reduction)

### 4. Test Dependency Isolation
Move testify imports to separate test files to ensure they don't accidentally get included:
```go
// +build test
```

### 5. Module Optimization
```bash
# Clean up module cache
go mod tidy
go mod download
```

### 6. UPX Compression (External Tool)
As a last resort, use UPX to compress the binary:
```bash
upx --best bin/go-llms
```
This can reduce size by 50-70% but may affect startup time and antivirus detection.

## Implementation Priority

1. **Replace json-iterator** (High impact, low effort)
   - Create a feature flag to switch between implementations
   - Benchmark to ensure acceptable performance
   
2. **Add build tags** for conditional provider compilation (Medium impact, medium effort)
   
3. **Try additional build flags** (Low impact, no effort)

4. **Consider UPX** only if absolutely necessary (High impact, but with trade-offs)

## Expected Results

Without major refactoring:
- Replacing json-iterator: ~500KB reduction (to ~5.8MB)
- Build tags for providers: ~200-300KB per excluded provider
- Additional build flags: ~50-100KB
- **Total potential reduction**: ~1MB (to ~5.3MB, 16% reduction)

With UPX compression:
- Additional 50-70% reduction (to ~2-3MB)
- But with startup time penalty