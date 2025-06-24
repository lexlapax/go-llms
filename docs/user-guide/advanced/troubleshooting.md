# Troubleshooting Guide: Problem Diagnosis and Resolution

> **[Project Root](/) / [Documentation](../..) / [User Guide](../../user-guide) / [Advanced Topics](../../user-guide/advanced) / Troubleshooting**

Comprehensive troubleshooting guide for diagnosing and resolving common issues in Go-LLMs applications, including debugging techniques, error analysis, performance problems, and integration challenges.

## Quick Diagnostics

### Health Check Script

```bash
#!/bin/bash
# go-llms-healthcheck.sh

echo "🔍 Go-LLMs Health Check"
echo "======================"

# Check environment
echo -e "\n📋 Environment Variables:"
for var in OPENAI_API_KEY ANTHROPIC_API_KEY GOOGLE_API_KEY; do
    if [ -z "${!var}" ]; then
        echo "❌ $var: NOT SET"
    else
        echo "✅ $var: SET (${#var} chars)"
    fi
done

# Check connectivity
echo -e "\n🌐 API Connectivity:"
urls=(
    "https://api.openai.com/v1/models"
    "https://api.anthropic.com/v1/models"
    "https://generativelanguage.googleapis.com/v1beta/models"
)

for url in "${urls[@]}"; do
    if curl -s -o /dev/null -w "%{http_code}" "$url" | grep -q "401\|200"; then
        echo "✅ $url: REACHABLE"
    else
        echo "❌ $url: UNREACHABLE"
    fi
done

# Check Go installation
echo -e "\n🔧 Go Environment:"
if command -v go &> /dev/null; then
    echo "✅ Go version: $(go version)"
    echo "✅ GOPATH: $GOPATH"
    echo "✅ GOMODCACHE: $(go env GOMODCACHE)"
else
    echo "❌ Go not installed"
fi

# Check module
echo -e "\n📦 Module Status:"
if [ -f "go.mod" ]; then
    echo "✅ go.mod found"
    go mod verify && echo "✅ Module verified" || echo "❌ Module verification failed"
else
    echo "❌ go.mod not found"
fi
```

---

## Common Issues and Solutions

### API Key Issues

#### Problem: "invalid API key" or authentication errors
```
Error: provider error: authentication failed: invalid API key
```

**Diagnosis:**
```go
// Test API key validity
func testAPIKey(provider string, apiKey string) error {
    ctx := context.Background()
    
    switch provider {
    case "openai":
        p, err := provider.NewOpenAIProvider(os.Getenv("OPENAI_API_KEY"), "gpt-4", 
            APIKey: apiKey,
}
        if err != nil {
            return err
        }
        _, err = p.ListModels(ctx)
        return err
        
    case "anthropic":
        // Similar test for Anthropic
    }
    
    return fmt.Errorf("unknown provider: %s", provider)
}
```

**Solutions:**
1. Verify API key format:
   ```bash
   # OpenAI keys start with 'sk-'
   echo $OPENAI_API_KEY | grep -E '^sk-[a-zA-Z0-9]{48}$'
   
   # Anthropic keys start with 'sk-ant-'
   echo $ANTHROPIC_API_KEY | grep -E '^sk-ant-[a-zA-Z0-9]+'
   ```

2. Check environment variable loading:
   ```go
   fmt.Printf("Environment API Key: %s\n", os.Getenv("OPENAI_API_KEY"))
   fmt.Printf("Loaded API Key: %s\n", config.APIKey)
   ```

3. Verify key permissions in provider dashboard

#### Problem: Rate limiting errors
```
Error: rate limit exceeded: 429 Too Many Requests
```

**Diagnosis:**
```go
// Monitor rate limit headers
type RateLimitMonitor struct {
    remaining int
    reset     time.Time
    mu        sync.RWMutex
}

func (m *RateLimitMonitor) ExtractFromResponse(resp *http.Response) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining != "" {
        m.remaining, _ = strconv.Atoi(remaining)
    }
    
    if reset := resp.Header.Get("X-RateLimit-Reset"); reset != "" {
        resetTime, _ := strconv.ParseInt(reset, 10, 64)
        m.reset = time.Unix(resetTime, 0)
    }
    
    log.Printf("Rate limit: %d remaining, resets at %v", m.remaining, m.reset)
}
```

**Solutions:**
1. Implement exponential backoff:
   ```go
   func retryWithBackoff(fn func() error) error {
       backoff := 100 * time.Millisecond
       maxBackoff := 30 * time.Second
       
       for attempt := 0; attempt < 5; attempt++ {
           err := fn()
           if err == nil {
               return nil
           }
           
           var rateLimitErr *RateLimitError
           if !errors.As(err, &rateLimitErr) {
               return err
           }
           
           time.Sleep(backoff)
           backoff = min(backoff*2, maxBackoff)
       }
       
       return fmt.Errorf("max retries exceeded")
   }
   ```

2. Use request queuing:
   ```go
   type RequestQueue struct {
       requests chan Request
       limiter  *rate.Limiter
   }
   
   func (q *RequestQueue) Process(ctx context.Context) {
       for req := range q.requests {
           if err := q.limiter.Wait(ctx); err != nil {
               req.Error <- err
               continue
           }
           
           resp, err := q.execute(req)
           req.Response <- resp
           req.Error <- err
       }
   }
   ```

---

### Connection and Network Issues

#### Problem: Connection timeouts
```
Error: Post "https://api.openai.com/v1/completions": context deadline exceeded
```

**Diagnosis:**
```go
// Network connectivity test
func diagnoseNetwork(provider string) {
    urls := map[string]string{
        "openai":    "https://api.openai.com/v1/models",
        "anthropic": "https://api.anthropic.com/v1/models",
        "gemini":    "https://generativelanguage.googleapis.com/v1beta/models",
    }
    
    url := urls[provider]
    
    // Test DNS resolution
    u, _ := url.Parse(url)
    ips, err := net.LookupIP(u.Hostname())
    if err != nil {
        log.Printf("DNS resolution failed: %v", err)
        return
    }
    log.Printf("Resolved to IPs: %v", ips)
    
    // Test TCP connection
    conn, err := net.DialTimeout("tcp", u.Host+":443", 5*time.Second)
    if err != nil {
        log.Printf("TCP connection failed: %v", err)
        return
    }
    conn.Close()
    
    // Test HTTP request
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Get(url)
    if err != nil {
        log.Printf("HTTP request failed: %v", err)
        return
    }
    defer resp.Body.Close()
    
    log.Printf("HTTP status: %d", resp.StatusCode)
}
```

**Solutions:**
1. Increase timeouts:
   ```go
   httpClient := &http.Client{
       Timeout: 60 * time.Second,
       Transport: &http.Transport{
           DialContext: (&net.Dialer{
               Timeout:   30 * time.Second,
               KeepAlive: 30 * time.Second,
           }).DialContext,
           TLSHandshakeTimeout:   10 * time.Second,
           ResponseHeaderTimeout: 20 * time.Second,
           IdleConnTimeout:       90 * time.Second,
       },
   }
   ```

2. Configure proxy if needed:
   ```go
   proxyURL, _ := url.Parse("http://proxy.company.com:8080")
   httpClient := &http.Client{
       Transport: &http.Transport{
           Proxy: http.ProxyURL(proxyURL),
       },
   }
   ```

3. Use connection pooling:
   ```go
   transport := &http.Transport{
       MaxIdleConns:        100,
       MaxIdleConnsPerHost: 10,
       MaxConnsPerHost:     20,
   }
   ```

#### Problem: SSL/TLS errors
```
Error: x509: certificate signed by unknown authority
```

**Solutions:**
1. Update CA certificates:
   ```bash
   # Ubuntu/Debian
   sudo apt-get update && sudo apt-get install ca-certificates
   
   # macOS
   brew install ca-certificates
   ```

2. Configure custom CA (for corporate environments):
   ```go
   caCert, _ := ioutil.ReadFile("/path/to/ca.crt")
   caCertPool := x509.NewCertPool()
   caCertPool.AppendCertsFromPEM(caCert)
   
   httpClient := &http.Client{
       Transport: &http.Transport{
           TLSClientConfig: &tls.Config{
               RootCAs: caCertPool,
           },
       },
   }
   ```

---

### Model and Provider Issues

#### Problem: Model not found
```
Error: model 'gpt-4' not found
```

**Diagnosis:**
```go
// List available models
func listAvailableModels(p provider.Provider) {
    ctx := context.Background()
    models, err := p.ListModels(ctx)
    if err != nil {
        log.Printf("Failed to list models: %v", err)
        return
    }
    
    fmt.Println("Available models:")
    for _, model := range models {
        fmt.Printf("- %s (context: %d tokens)\n", model.ID, model.Context)
    }
}
```

**Solutions:**
1. Use correct model names:
   ```go
   modelMappings := map[string]map[string]string{
       "openai": {
           "gpt4": "gpt-4",
           "gpt35": "gpt-3.5-turbo",
           "gpt4o": "gpt-4o",
       },
       "anthropic": {
           "claude": "claude-3-haiku",
           "claude2": "claude-2.1",
       },
   }
   ```

2. Check model availability for your account
3. Use fallback models:
   ```go
   models := []string{"gpt-4", "gpt-3.5-turbo"}
   var lastErr error
   
   for _, model := range models {
       req.Model = model
       resp, err := provider.Complete(ctx, req)
       if err == nil {
           return resp, nil
       }
       lastErr = err
   }
   
   return nil, lastErr
   ```

#### Problem: Streaming not working
```
Error: streaming not supported
```

**Solutions:**
1. Check provider support:
   ```go
   type StreamingProvider interface {
       provider.Provider
       CompleteStream(context.Context, *CompletionRequest) (<-chan StreamChunk, error)
   }
   
   if sp, ok := p.(StreamingProvider); ok {
       stream, err := sp.CompleteStream(ctx, req)
       // Handle stream
   } else {
       // Fall back to non-streaming
   }
   ```

2. Handle stream errors:
   ```go
   for chunk := range stream {
       if chunk.Error != nil {
           log.Printf("Stream error: %v", chunk.Error)
           break
       }
       
       fmt.Print(chunk.Content)
   }
   ```

---

### Memory and Performance Issues

#### Problem: High memory usage
```
runtime: out of memory
```

**Diagnosis:**
```go
// Memory profiling
func profileMemory() {
    f, _ := os.Create("mem.prof")
    defer f.Close()
    
    runtime.GC()
    pprof.WriteHeapProfile(f)
}

// Monitor memory usage
func monitorMemory() {
    var m runtime.MemStats
    
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        runtime.ReadMemStats(&m)
        log.Printf("Alloc: %v MB, TotalAlloc: %v MB, Sys: %v MB, NumGC: %v",
            m.Alloc/1024/1024,
            m.TotalAlloc/1024/1024,
            m.Sys/1024/1024,
            m.NumGC,
        )
    }
}
```

**Solutions:**
1. Implement object pooling:
   ```go
   var bufferPool = sync.Pool{
       New: func() interface{} {
           return bytes.NewBuffer(make([]byte, 0, 4096))
       },
   }
   
   func processWithPool(data []byte) {
       buf := bufferPool.Get().(*bytes.Buffer)
       defer func() {
           buf.Reset()
           bufferPool.Put(buf)
       }()
       
       // Use buffer
   }
   ```

2. Stream large responses:
   ```go
   func streamLargeResponse(reader io.Reader) error {
       scanner := bufio.NewScanner(reader)
       scanner.Buffer(make([]byte, 1024), 1024*1024) // 1MB max
       
       for scanner.Scan() {
           processLine(scanner.Text())
       }
       
       return scanner.Err()
   }
   ```

3. Limit concurrent operations:
   ```go
   sem := make(chan struct{}, 10) // Max 10 concurrent
   
   for _, item := range items {
       sem <- struct{}{}
       go func(i Item) {
           defer func() { <-sem }()
           process(i)
       }(item)
   }
   ```

#### Problem: Slow response times
```
Request took 30s to complete
```

**Diagnosis:**
```go
// Performance tracing
func tracePerformance(ctx context.Context, operation string) func() {
    start := time.Now()
    
    return func() {
        duration := time.Since(start)
        log.Printf("%s took %v", operation, duration)
        
        if duration > 5*time.Second {
            log.Printf("SLOW: %s exceeded threshold", operation)
        }
    }
}

// Usage
defer tracePerformance(ctx, "API call")()
```

**Solutions:**
1. Implement caching:
   ```go
   type ResponseCache struct {
       cache *lru.Cache
       ttl   time.Duration
   }
   
   func (c *ResponseCache) Get(key string) (interface{}, bool) {
       if val, ok := c.cache.Get(key); ok {
           entry := val.(*CacheEntry)
           if time.Since(entry.Time) < c.ttl {
               return entry.Data, true
           }
           c.cache.Remove(key)
       }
       return nil, false
   }
   ```

2. Optimize prompts:
   ```go
   func optimizePrompt(prompt string) string {
       // Remove redundant whitespace
       prompt = strings.TrimSpace(prompt)
       prompt = regexp.MustCompile(`\s+`).ReplaceAllString(prompt, " ")
       
       // Truncate if too long
       if len(prompt) > 2000 {
           prompt = prompt[:2000] + "..."
       }
       
       return prompt
   }
   ```

---

### Tool and Agent Issues

#### Problem: Tool execution failures
```
Error: tool 'web_fetch' failed: connection refused
```

**Diagnosis:**
```go
// Tool diagnostics
func diagnoseTool(tool domain.Tool) {
    // Check tool metadata
    fmt.Printf("Tool: %s\n", tool.Name())
    fmt.Printf("Description: %s\n", tool.Description())
    
    // Validate schemas
    inputSchema := tool.InputSchema()
    outputSchema := tool.OutputSchema()
    
    // Test with sample input
    ctx := context.Background()
    testInput := generateTestInput(inputSchema)
    
    result, err := tool.Execute(ctx, testInput)
    if err != nil {
        fmt.Printf("Execution error: %v\n", err)
        
        // Check specific error types
        var netErr net.Error
        if errors.As(err, &netErr) {
            fmt.Printf("Network error: timeout=%v, temporary=%v\n",
                netErr.Timeout(), netErr.Temporary())
        }
    } else {
        fmt.Printf("Success: %+v\n", result)
    }
}
```

**Solutions:**
1. Validate tool configuration:
   ```go
   func validateToolConfig(config map[string]interface{}) error {
       required := []string{"api_key", "timeout", "base_url"}
       
       for _, key := range required {
           if _, ok := config[key]; !ok {
               return fmt.Errorf("missing required config: %s", key)
           }
       }
       
       return nil
   }
   ```

2. Implement tool health checks:
   ```go
   type HealthCheckableTool interface {
       domain.Tool
       HealthCheck(ctx context.Context) error
   }
   
   func monitorToolHealth(tools []domain.Tool) {
       for _, tool := range tools {
           if hc, ok := tool.(HealthCheckableTool); ok {
               if err := hc.HealthCheck(context.Background()); err != nil {
                   log.Printf("Tool %s unhealthy: %v", tool.Name(), err)
               }
           }
       }
   }
   ```

#### Problem: Agent not using tools correctly
```
Agent ignores available tools
```

**Solutions:**
1. Improve system prompt:
   ```go
   systemPrompt := `You are an AI assistant with access to the following tools:
   %s
   
   When you need to perform an action, use the appropriate tool by specifying:
   - tool_name: The name of the tool to use
   - parameters: The input parameters for the tool
   
   Always use tools when available rather than trying to answer without them.`
   
   toolDescriptions := generateToolDescriptions(agent.Tools())
   agent.SetSystemPrompt(fmt.Sprintf(systemPrompt, toolDescriptions))
   ```

2. Add tool usage examples:
   ```go
   type ToolExample struct {
       Description string
       Input       interface{}
       Output      interface{}
   }
   
   func addToolExamples(agent *core.LLMAgent, examples map[string][]ToolExample) {
       for toolName, exs := range examples {
           for _, ex := range exs {
               agent.AddExample(core.Message{
                   Role: "user",
                   Content: ex.Description,
               }, core.Message{
                   Role: "assistant",
                   Content: fmt.Sprintf("I'll use the %s tool: %v", toolName, ex.Input),
               })
           }
       }
   }
   ```

---

### Integration Issues

#### Problem: JSON parsing errors
```
Error: invalid character 'I' looking for beginning of value
```

**Diagnosis:**
```go
// JSON validation
func validateJSON(data []byte) {
    var js json.RawMessage
    if err := json.Unmarshal(data, &js); err != nil {
        // Find error position
        if syntaxErr, ok := err.(*json.SyntaxError); ok {
            start := max(0, syntaxErr.Offset-20)
            end := min(len(data), syntaxErr.Offset+20)
            
            fmt.Printf("JSON error at position %d:\n", syntaxErr.Offset)
            fmt.Printf("Context: ...%s...\n", data[start:end])
            fmt.Printf("         %s^\n", strings.Repeat(" ", int(syntaxErr.Offset-start)))
        }
    }
}
```

**Solutions:**
1. Extract JSON from mixed content:
   ```go
   func extractJSON(content string) ([]byte, error) {
       // Find JSON boundaries
       start := strings.Index(content, "{")
       if start == -1 {
           start = strings.Index(content, "[")
       }
       
       if start == -1 {
           return nil, errors.New("no JSON found")
       }
       
       // Extract and validate
       var depth int
       inString := false
       escape := false
       
       for i := start; i < len(content); i++ {
           char := content[i]
           
           if !escape {
               switch char {
               case '"':
                   if !inString {
                       inString = true
                   } else {
                       inString = false
                   }
               case '{', '[':
                   if !inString {
                       depth++
                   }
               case '}', ']':
                   if !inString {
                       depth--
                       if depth == 0 {
                           return []byte(content[start:i+1]), nil
                       }
                   }
               case '\\':
                   escape = true
                   continue
               }
           }
           escape = false
       }
       
       return nil, errors.New("incomplete JSON")
   }
   ```

2. Use structured output:
   ```go
   request := &CompletionRequest{
       Messages: messages,
       ResponseFormat: &ResponseFormat{
           Type: "json_object",
           Schema: outputSchema,
       },
   }
   ```

#### Problem: Schema validation failures
```
Error: field 'age' must be number, got string
```

**Solutions:**
1. Implement type coercion:
   ```go
   func coerceTypes(data map[string]interface{}, schema Schema) error {
       for field, spec := range schema.Properties {
           if val, ok := data[field]; ok {
               coerced, err := coerceValue(val, spec.Type)
               if err != nil {
                   return fmt.Errorf("field %s: %w", field, err)
               }
               data[field] = coerced
           }
       }
       return nil
   }
   
   func coerceValue(val interface{}, targetType string) (interface{}, error) {
       switch targetType {
       case "number":
           switch v := val.(type) {
           case string:
               return strconv.ParseFloat(v, 64)
           case int:
               return float64(v), nil
           }
           
       case "string":
           return fmt.Sprintf("%v", val), nil
           
       case "boolean":
           switch v := val.(type) {
           case string:
               return strconv.ParseBool(v)
           case int:
               return v != 0, nil
           }
       }
       
       return val, nil
   }
   ```

2. Add validation feedback:
   ```go
   func validateWithFeedback(data interface{}, schema Schema) []ValidationError {
       var errors []ValidationError
       
       // Detailed validation with paths
       validateRecursive("", data, schema, &errors)
       
       return errors
   }
   
   type ValidationError struct {
       Path    string
       Message string
       Value   interface{}
       Rule    string
   }
   ```

---

## Debug Techniques

### Enable Debug Logging

```go
// Debug logger configuration
func setupDebugLogging() *zap.Logger {
    config := zap.NewDevelopmentConfig()
    config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
    config.EncoderConfig.TimeKey = "timestamp"
    config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
    
    logger, _ := config.Build()
    
    // Add debug middleware
    logger = logger.With(
        zap.String("app", "go-llms"),
        zap.String("version", version),
    )
    
    return logger
}

// HTTP request/response debugging
type DebugTransport struct {
    Transport http.RoundTripper
    Logger    *zap.Logger
}

func (d *DebugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    // Log request
    reqBody, _ := httputil.DumpRequestOut(req, true)
    d.Logger.Debug("HTTP Request",
        zap.String("method", req.Method),
        zap.String("url", req.URL.String()),
        zap.ByteString("body", reqBody),
    )
    
    // Execute request
    resp, err := d.Transport.RoundTrip(req)
    if err != nil {
        d.Logger.Error("HTTP Error", zap.Error(err))
        return nil, err
    }
    
    // Log response
    respBody, _ := httputil.DumpResponse(resp, true)
    d.Logger.Debug("HTTP Response",
        zap.Int("status", resp.StatusCode),
        zap.ByteString("body", respBody),
    )
    
    return resp, nil
}
```

### Interactive Debugging

```go
// Debug REPL for testing
func startDebugREPL(provider provider.Provider) {
    scanner := bufio.NewScanner(os.Stdin)
    ctx := context.Background()
    
    fmt.Println("Go-LLMs Debug REPL")
    fmt.Println("Commands: .models, .test, .exit")
    
    for {
        fmt.Print("> ")
        if !scanner.Scan() {
            break
        }
        
        input := scanner.Text()
        
        switch input {
        case ".models":
            models, err := provider.ListModels(ctx)
            if err != nil {
                fmt.Printf("Error: %v\n", err)
            } else {
                for _, m := range models {
                    fmt.Printf("- %s\n", m.ID)
                }
            }
            
        case ".test":
            resp, err := provider.Complete(ctx, &CompletionRequest{
                Model: "gpt-3.5-turbo",
                Messages: []Message{
                    {Role: "user", Content: "Say hello"},
                },
}
            if err != nil {
                fmt.Printf("Error: %v\n", err)
            } else {
                fmt.Printf("Response: %s\n", resp.Content)
            }
            
        case ".exit":
            return
            
        default:
            // Send as prompt
            resp, err := provider.Complete(ctx, &CompletionRequest{
                Messages: []Message{
                    {Role: "user", Content: input},
                },
}
            if err != nil {
                fmt.Printf("Error: %v\n", err)
            } else {
                fmt.Printf("%s\n", resp.Content)
            }
        }
    }
}
```

### Trace Execution

```go
// Execution tracer
type ExecutionTracer struct {
    spans []TraceSpan
    mu    sync.Mutex
}

type TraceSpan struct {
    Name      string
    Start     time.Time
    End       time.Time
    Tags      map[string]interface{}
    Error     error
}

func (t *ExecutionTracer) StartSpan(name string) *Span {
    span := &Span{
        tracer: t,
        Name:   name,
        Start:  time.Now(),
        Tags:   make(map[string]interface{}),
    }
    
    return span
}

func (s *Span) End() {
    s.End = time.Now()
    
    s.tracer.mu.Lock()
    s.tracer.spans = append(s.tracer.spans, TraceSpan{
        Name:  s.Name,
        Start: s.Start,
        End:   s.End,
        Tags:  s.Tags,
        Error: s.Error,
}
    s.tracer.mu.Unlock()
}

func (t *ExecutionTracer) PrintTrace() {
    t.mu.Lock()
    defer t.mu.Unlock()
    
    fmt.Println("Execution Trace:")
    for _, span := range t.spans {
        duration := span.End.Sub(span.Start)
        status := "OK"
        if span.Error != nil {
            status = fmt.Sprintf("ERROR: %v", span.Error)
        }
        
        fmt.Printf("  %s: %v [%s]\n", span.Name, duration, status)
        
        for k, v := range span.Tags {
            fmt.Printf("    %s: %v\n", k, v)
        }
    }
}
```

---

## Error Recovery Strategies

### Automatic Recovery

```go
// Self-healing system
type SelfHealingSystem struct {
    monitors  []HealthMonitor
    healers   map[string]Healer
    alerts    AlertSystem
}

type HealthMonitor interface {
    Check(ctx context.Context) (string, error)
}

type Healer interface {
    Heal(ctx context.Context, issue string) error
}

func (s *SelfHealingSystem) Run(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            s.checkAndHeal(ctx)
            
        case <-ctx.Done():
            return
        }
    }
}

func (s *SelfHealingSystem) checkAndHeal(ctx context.Context) {
    for _, monitor := range s.monitors {
        issue, err := monitor.Check(ctx)
        if err != nil {
            log.Printf("Monitor check failed: %v", err)
            continue
        }
        
        if issue != "" {
            log.Printf("Issue detected: %s", issue)
            
            if healer, ok := s.healers[issue]; ok {
                if err := healer.Heal(ctx, issue); err != nil {
                    s.alerts.Alert(fmt.Sprintf("Healing failed for %s: %v", issue, err))
                } else {
                    log.Printf("Successfully healed: %s", issue)
                }
            } else {
                s.alerts.Alert(fmt.Sprintf("No healer for issue: %s", issue))
            }
        }
    }
}

// Example healers
healers := map[string]Healer{
    "high_memory": &MemoryHealer{
        threshold: 80, // 80% memory usage
        actions: []func(){
            runtime.GC,
            clearCaches,
            reduceWorkers,
        },
    },
    "connection_pool_exhausted": &ConnectionHealer{
        resetPool: true,
        increaseSize: true,
    },
    "rate_limit": &RateLimitHealer{
        backoffDuration: 5 * time.Minute,
        reduceRate: 0.5,
    },
}
```

### Graceful Degradation

```go
// Degradation manager
type DegradationManager struct {
    levels   []DegradationLevel
    current  int
    metrics  *SystemMetrics
}

type DegradationLevel struct {
    Name        string
    Threshold   float64
    Actions     []DegradationAction
}

func (d *DegradationManager) CheckAndAdjust() {
    load := d.metrics.GetSystemLoad()
    
    // Find appropriate level
    newLevel := 0
    for i, level := range d.levels {
        if load > level.Threshold {
            newLevel = i
        }
    }
    
    // Apply degradation if needed
    if newLevel != d.current {
        if newLevel > d.current {
            log.Printf("Degrading from level %d to %d", d.current, newLevel)
            d.applyDegradation(d.levels[newLevel])
        } else {
            log.Printf("Recovering from level %d to %d", d.current, newLevel)
            d.removeDegradation(d.levels[d.current])
        }
        
        d.current = newLevel
    }
}

// Example degradation levels
levels := []DegradationLevel{
    {
        Name: "normal",
        Threshold: 0.0,
        Actions: []DegradationAction{},
    },
    {
        Name: "high_load",
        Threshold: 0.7,
        Actions: []DegradationAction{
            DisableNonEssentialFeatures{},
            ReduceCacheTTL{Duration: 5 * time.Minute},
            LimitConcurrency{Max: 50},
        },
    },
    {
        Name: "critical",
        Threshold: 0.9,
        Actions: []DegradationAction{
            EnableReadOnlyMode{},
            RejectNonPriorityRequests{},
            MinimalResponseMode{},
        },
    },
}
```

---

## Monitoring and Alerting

### Health Dashboard

```go
// Simple health dashboard
func StartHealthDashboard(port int) {
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        health := collectHealthStatus()
        
        w.Header().Set("Content-Type", "application/json")
        if health.Status != "healthy" {
            w.WriteHeader(http.StatusServiceUnavailable)
        }
        
        json.NewEncoder(w).Encode(health)
}
    
    http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
        metrics := collectMetrics()
        
        w.Header().Set("Content-Type", "text/plain")
        for name, value := range metrics {
            fmt.Fprintf(w, "%s %f\n", name, value)
        }
}
    
    log.Printf("Health dashboard started on :%d", port)
    http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

type HealthStatus struct {
    Status     string                 `json:"status"`
    Version    string                 `json:"version"`
    Uptime     time.Duration          `json:"uptime"`
    Checks     map[string]CheckResult `json:"checks"`
    Timestamp  time.Time              `json:"timestamp"`
}

type CheckResult struct {
    Status  string `json:"status"`
    Message string `json:"message,omitempty"`
    Latency int64  `json:"latency_ms"`
}
```

### Alert Configuration

```yaml
# alerts.yaml
alerts:
  - name: high_error_rate
    condition: error_rate > 0.05
    duration: 5m
    severity: warning
    actions:
      - log
      - email
      
  - name: api_timeout
    condition: response_time_p95 > 5000
    duration: 2m
    severity: critical
    actions:
      - log
      - pagerduty
      - slack
      
  - name: memory_leak
    condition: memory_usage_trend > 0.1
    duration: 30m
    severity: warning
    actions:
      - log
      - email
      - auto_restart
```

---

## Troubleshooting Checklist

### Initial Diagnosis
- [ ] Check environment variables are set correctly
- [ ] Verify API keys are valid and have proper permissions
- [ ] Test network connectivity to provider endpoints
- [ ] Review recent code changes or deployments
- [ ] Check system resources (CPU, memory, disk)

### Error Investigation
- [ ] Enable debug logging
- [ ] Collect error stack traces
- [ ] Check provider status pages
- [ ] Review rate limit headers
- [ ] Analyze request/response payloads

### Performance Issues
- [ ] Profile CPU and memory usage
- [ ] Check for memory leaks
- [ ] Analyze response times
- [ ] Review concurrent request counts
- [ ] Check cache hit rates

### Integration Problems
- [ ] Validate JSON schemas
- [ ] Check type conversions
- [ ] Review tool configurations
- [ ] Test with minimal examples
- [ ] Check version compatibility

### Recovery Actions
- [ ] Restart affected services
- [ ] Clear caches if corrupted
- [ ] Reduce load temporarily
- [ ] Switch to fallback providers
- [ ] Apply emergency patches

---

## Getting Help

### Gathering Diagnostic Information

```bash
#!/bin/bash
# collect-diagnostics.sh

OUTPUT_DIR="diagnostics-$(date +%Y%m%d-%H%M%S)"
mkdir -p "$OUTPUT_DIR"

# System information
echo "Collecting system information..."
uname -a > "$OUTPUT_DIR/system.txt"
go version >> "$OUTPUT_DIR/system.txt"
go env >> "$OUTPUT_DIR/system.txt"

# Environment
echo "Collecting environment..."
env | grep -E "(OPENAI|ANTHROPIC|GOOGLE|OLLAMA|GOLLMS)" > "$OUTPUT_DIR/environment.txt"

# Module information
echo "Collecting module information..."
go list -m all > "$OUTPUT_DIR/modules.txt"
go mod graph > "$OUTPUT_DIR/mod-graph.txt"

# Recent logs
echo "Collecting logs..."
if [ -f "app.log" ]; then
    tail -n 1000 app.log > "$OUTPUT_DIR/recent-logs.txt"
fi

# Create archive
tar -czf "diagnostics-$(date +%Y%m%d-%H%M%S).tar.gz" "$OUTPUT_DIR"
rm -rf "$OUTPUT_DIR"

echo "Diagnostics collected in diagnostics-*.tar.gz"
```

### Reporting Issues

When reporting issues, include:
1. Go-LLMs version
2. Provider being used
3. Error messages and stack traces
4. Minimal code example reproducing the issue
5. Diagnostic information from the script above

### Community Resources

- GitHub Issues: Report bugs and request features
- Documentation: Check for updates and examples
- Community Forums: Ask questions and share solutions

---

## Next Steps

- **[Performance Optimization](performance-optimization.md)** - Improve application performance
- **[Security Considerations](security-considerations.md)** - Security troubleshooting
- **[Best Practices Checklist](../../user-guide/reference/best-practices-checklist.md)** - Avoid common issues
- **[Error Codes Reference](../../user-guide/reference/error-codes-reference.md)** - Detailed error information
- **[Configuration Reference](../../user-guide/reference/configuration-reference.md)** - Configuration troubleshooting