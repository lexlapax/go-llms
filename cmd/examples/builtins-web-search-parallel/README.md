# Built-ins Web Search Parallel Example

This example demonstrates how to use the `engine_api_key` parameter for production scenarios where you need to:
- Use different API keys for different searches
- Perform parallel searches across multiple engines
- Compare results from different search providers
- Implement redundancy and failover

## Features

1. **Explicit API Key Management**
   - No reliance on environment variables
   - Keys injected programmatically from secure storage
   - Different keys for different operations

2. **Parallel Execution**
   - Searches all available engines concurrently
   - Collects results with timing information
   - Handles failures gracefully

3. **Result Comparison**
   - Compare results from different engines
   - Find the fastest responding engine
   - Implement consensus or quality checks

## Usage

```bash
# Set API keys (optional - will use DuckDuckGo if none are set)
export BRAVE_API_KEY="your-brave-api-key"
export TAVILY_API_KEY="your-tavily-api-key"
export SERPERDEV_API_KEY="your-serperdev-api-key"
export SERPAPI_API_KEY="your-serpapi-api-key"

# Run the example
go run cmd/examples/builtins-web-search-parallel/main.go

# Or from the examples directory
cd cmd/examples/builtins-web-search-parallel
go run main.go
```

### Environment Variables

- `BRAVE_API_KEY` - API key for Brave Search (get from https://brave.com/search/api/)
- `TAVILY_API_KEY` - API key for Tavily Search (get from https://tavily.com/)
- `SERPERDEV_API_KEY` - API key for Serper.dev (get from https://serper.dev/)
- `SERPAPI_API_KEY` - API key for Serpapi Search (get from https://serpapi.com/)

Note: In production, keys would typically come from secure storage:
- AWS Secrets Manager
- HashiCorp Vault
- Kubernetes Secrets
- Environment-specific configuration

## Production Patterns

### Multi-Tenant Scenarios
```go
// Different API keys per customer
customerKeys := map[string]SearchConfig{
    "customer1": {BraveAPIKey: "customer1-brave-key"},
    "customer2": {TavilyAPIKey: "customer2-tavily-key"},
}
```

### A/B Testing
```go
// Test different search engines
if useExperiment {
    params["engine"] = "serper"
    params["engine_api_key"] = experimentalKey
} else {
    params["engine"] = "brave"
    params["engine_api_key"] = productionKey
}
```

### Rate Limit Management
```go
// Use different keys to avoid rate limits
keys := []string{primaryKey, secondaryKey, tertiaryKey}
keyIndex := requestCount % len(keys)
params["engine_api_key"] = keys[keyIndex]
```

## Benefits

1. **Security**: API keys never exposed in environment
2. **Flexibility**: Different keys for different use cases
3. **Scalability**: Distribute load across multiple accounts
4. **Reliability**: Failover to backup keys if needed
5. **Auditability**: Track which keys are used where