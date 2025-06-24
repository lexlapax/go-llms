# Performance Optimization: Tuning and Optimization Guide

> **[Project Root](/) / [Documentation](../..) / [User Guide](../../user-guide) / [Advanced Topics](../../user-guide/advanced) / Performance Optimization**

Comprehensive guide to optimizing Go-LLMs applications for maximum performance, efficiency, and scalability. Learn proven techniques for reducing latency, improving throughput, and minimizing costs.

## Performance Overview

Performance optimization in LLM applications involves multiple layers:
- **Request Optimization** - Minimize API calls and token usage
- **Concurrency Management** - Efficient parallel processing
- **Caching Strategies** - Intelligent response caching
- **Resource Utilization** - Memory and CPU optimization
- **Network Efficiency** - Connection pooling and batching

---

## Benchmarking and Profiling

### Setting Up Benchmarks

```go
package main

import (
    "context"
    "testing"
    "time"
    
    "github.com/lexlapax/go-llms/pkg/llm/provider"
)

// Benchmark single request performance
func BenchmarkSingleRequest(b *testing.B) {
    provider, _ := provider.NewOpenAIProvider(os.Getenv("OPENAI_API_KEY"), "gpt-4", 
    // APIKey: os.Getenv("OPENAI_API_KEY"), // Moved to constructor parameters
}
    
    request := &CompletionRequest{
        Messages: []Message{{Role: "user", Content: "Hello"}},
        Model:    "gpt-4o-mini",
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := provider.Complete(context.Background(), request)
        if err != nil {
            b.Fatal(err)
        }
    }
}

// Benchmark concurrent requests
func BenchmarkConcurrentRequests(b *testing.B) {
    provider, _ := provider.NewOpenAIProvider(os.Getenv("OPENAI_API_KEY"), "gpt-4", 
    // APIKey: os.Getenv("OPENAI_API_KEY"), // Moved to constructor parameters
}
    
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            _, err := provider.Complete(context.Background(), request)
            if err != nil {
                b.Fatal(err)
            }
        }
}
}
```

### CPU and Memory Profiling

```go
import (
    "net/http"
    _ "net/http/pprof"
    "runtime/pprof"
)

// Enable profiling endpoint
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()

// CPU profiling
func profileCPU() {
    f, _ := os.Create("cpu.prof")
    defer f.Close()
    
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()
    
    // Run your workload
    runWorkload()
}

// Memory profiling
func profileMemory() {
    f, _ := os.Create("mem.prof")
    defer f.Close()
    
    runtime.GC()
    pprof.WriteHeapProfile(f)
}

// Analyze profiles
// go tool pprof cpu.prof
// go tool pprof -http=:8080 mem.prof
```

---

## Request Optimization

### Prompt Engineering for Performance

```go
// Optimize prompts for token efficiency
type PromptOptimizer struct {
    maxTokens int
    cache     *PromptCache
}

func (po *PromptOptimizer) Optimize(prompt string) string {
    // Remove redundant whitespace
    prompt = strings.TrimSpace(prompt)
    prompt = regexp.MustCompile(`\s+`).ReplaceAllString(prompt, " ")
    
    // Use abbreviations for common phrases
    replacements := map[string]string{
        "for example": "e.g.",
        "that is": "i.e.",
        "and so on": "etc.",
    }
    
    for long, short := range replacements {
        prompt = strings.ReplaceAll(prompt, long, short)
    }
    
    // Compress system prompts
    if strings.HasPrefix(prompt, "You are") {
        prompt = po.compressSystemPrompt(prompt)
    }
    
    return prompt
}

// Token counting for cost optimization
func (po *PromptOptimizer) EstimateTokens(text string) int {
    // Rough estimation: 1 token ≈ 4 characters
    return len(text) / 4
}

// Dynamic model selection based on complexity
func selectOptimalModel(prompt string, requiresAccuracy bool) string {
    tokens := EstimateTokens(prompt)
    
    if tokens < 500 && !requiresAccuracy {
        return "gpt-3.5-turbo" // Fast and cheap
    } else if tokens < 2000 {
        return "gpt-4o-mini"   // Balanced
    } else {
        return "gpt-4o"        // High capacity
    }
}
```

### Request Batching

```go
// Batch multiple requests for efficiency
type RequestBatcher struct {
    batchSize    int
    batchTimeout time.Duration
    requests     chan BatchRequest
    provider     Provider
}

type BatchRequest struct {
    Request  *CompletionRequest
    Response chan *CompletionResponse
    Error    chan error
}

func (rb *RequestBatcher) Start(ctx context.Context) {
    batch := make([]*BatchRequest, 0, rb.batchSize)
    timer := time.NewTimer(rb.batchTimeout)
    
    for {
        select {
        case req := <-rb.requests:
            batch = append(batch, &req)
            
            if len(batch) >= rb.batchSize {
                rb.processBatch(ctx, batch)
                batch = batch[:0]
                timer.Reset(rb.batchTimeout)
            }
            
        case <-timer.C:
            if len(batch) > 0 {
                rb.processBatch(ctx, batch)
                batch = batch[:0]
            }
            timer.Reset(rb.batchTimeout)
            
        case <-ctx.Done():
            return
        }
    }
}

func (rb *RequestBatcher) processBatch(ctx context.Context, batch []*BatchRequest) {
    // Process requests in parallel with rate limiting
    limiter := rate.NewLimiter(rate.Limit(10), 1)
    var wg sync.WaitGroup
    
    for _, req := range batch {
        wg.Add(1)
        go func(br *BatchRequest) {
            defer wg.Done()
            
            limiter.Wait(ctx)
            resp, err := rb.provider.Complete(ctx, br.Request)
            
            if err != nil {
                br.Error <- err
            } else {
                br.Response <- resp
            }
        }(req)
    }
    
    wg.Wait()
}
```

---

## Caching Strategies

### Multi-Level Cache Implementation

```go
// Hierarchical cache system
type MultiLevelCache struct {
    l1Cache *MemoryCache  // In-memory, fastest
    l2Cache *RedisCache   // Distributed, shared
    l3Cache *DiskCache    // Persistent, large capacity
}

func (mlc *MultiLevelCache) Get(key string) (*CachedResponse, error) {
    // Check L1 (memory)
    if resp, ok := mlc.l1Cache.Get(key); ok {
        return resp, nil
    }
    
    // Check L2 (Redis)
    resp, err := mlc.l2Cache.Get(key)
    if err == nil && resp != nil {
        // Promote to L1
        mlc.l1Cache.Set(key, resp)
        return resp, nil
    }
    
    // Check L3 (disk)
    resp, err = mlc.l3Cache.Get(key)
    if err == nil && resp != nil {
        // Promote to L1 and L2
        mlc.l2Cache.Set(key, resp)
        mlc.l1Cache.Set(key, resp)
        return resp, nil
    }
    
    return nil, errors.New("cache miss")
}

// Intelligent cache key generation
func generateCacheKey(req *CompletionRequest) string {
    // Normalize request for better cache hits
    normalized := NormalizeRequest(req)
    
    h := sha256.New()
    h.Write([]byte(normalized.Model))
    h.Write([]byte(fmt.Sprintf("%.2f", normalized.Temperature)))
    
    for _, msg := range normalized.Messages {
        h.Write([]byte(msg.Role))
        h.Write([]byte(msg.Content))
    }
    
    return hex.EncodeToString(h.Sum(nil))
}

// Cache warming strategies
func (mlc *MultiLevelCache) WarmCache(ctx context.Context, patterns []string) {
    for _, pattern := range patterns {
        // Pre-generate common responses
        requests := generateCommonRequests(pattern)
        
        for _, req := range requests {
            if _, err := mlc.Get(generateCacheKey(req)); err != nil {
                // Cache miss, generate and store
                resp, err := generateResponse(ctx, req)
                if err == nil {
                    mlc.Set(generateCacheKey(req), resp)
                }
            }
        }
    }
}
```

### Semantic Caching

```go
// Cache based on semantic similarity
type SemanticCache struct {
    embedder     Embedder
    vectorStore  VectorStore
    threshold    float64
}

func (sc *SemanticCache) Get(query string) (*CachedResponse, error) {
    // Generate embedding for query
    embedding, err := sc.embedder.Embed(query)
    if err != nil {
        return nil, err
    }
    
    // Search for similar queries
    results, err := sc.vectorStore.Search(embedding, 5)
    if err != nil {
        return nil, err
    }
    
    // Check if any result is similar enough
    for _, result := range results {
        if result.Similarity >= sc.threshold {
            return result.Response, nil
        }
    }
    
    return nil, errors.New("no similar cache entry found")
}

func (sc *SemanticCache) Set(query string, response *CompletionResponse) error {
    embedding, err := sc.embedder.Embed(query)
    if err != nil {
        return err
    }
    
    return sc.vectorStore.Insert(embedding, &CachedResponse{
        Query:     query,
        Response:  response,
        Timestamp: time.Now(),
}
}
```

---

## Concurrency and Parallelism

### Optimized Worker Pool

```go
// High-performance worker pool
type WorkerPool struct {
    workers   int
    taskQueue chan Task
    results   chan Result
    limiter   *rate.Limiter
    metrics   *PoolMetrics
}

type Task struct {
    ID      string
    Request *CompletionRequest
    Context context.Context
}

type Result struct {
    ID       string
    Response *CompletionResponse
    Error    error
    Duration time.Duration
}

func (wp *WorkerPool) Start() {
    for i := 0; i < wp.workers; i++ {
        go wp.worker(i)
    }
}

func (wp *WorkerPool) worker(id int) {
    for task := range wp.taskQueue {
        start := time.Now()
        
        // Rate limiting
        wp.limiter.Wait(task.Context)
        
        // Process request
        response, err := processRequest(task.Context, task.Request)
        
        // Update metrics
        duration := time.Since(start)
        wp.metrics.RecordDuration(duration)
        
        if err != nil {
            wp.metrics.RecordError(err)
        }
        
        wp.results <- Result{
            ID:       task.ID,
            Response: response,
            Error:    err,
            Duration: duration,
        }
    }
}

// Dynamic worker scaling
func (wp *WorkerPool) AutoScale(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            queueSize := len(wp.taskQueue)
            avgLatency := wp.metrics.AverageLatency()
            
            if queueSize > 100 && avgLatency > 2*time.Second {
                // Scale up
                wp.AddWorkers(5)
            } else if queueSize < 10 && avgLatency < 500*time.Millisecond {
                // Scale down
                wp.RemoveWorkers(2)
            }
            
        case <-ctx.Done():
            return
        }
    }
}
```

### Pipeline Processing

```go
// Efficient pipeline for multi-stage processing
type Pipeline struct {
    stages []Stage
}

type Stage interface {
    Process(context.Context, interface{}) (interface{}, error)
}

func (p *Pipeline) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    result := input
    
    for i, stage := range p.stages {
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        default:
            var err error
            result, err = stage.Process(ctx, result)
            if err != nil {
                return nil, fmt.Errorf("stage %d failed: %w", i, err)
            }
        }
    }
    
    return result, nil
}

// Parallel pipeline execution
func (p *Pipeline) ExecuteParallel(ctx context.Context, inputs []interface{}) []Result {
    results := make([]Result, len(inputs))
    var wg sync.WaitGroup
    
    // Use semaphore for concurrency control
    sem := make(chan struct{}, runtime.NumCPU())
    
    for i, input := range inputs {
        wg.Add(1)
        go func(idx int, in interface{}) {
            defer wg.Done()
            
            sem <- struct{}{}
            defer func() { <-sem }()
            
            output, err := p.Execute(ctx, in)
            results[idx] = Result{
                Output: output,
                Error:  err,
            }
        }(i, input)
    }
    
    wg.Wait()
    return results
}
```

---

## Memory Optimization

### Object Pooling

```go
// Reuse objects to reduce GC pressure
var (
    requestPool = sync.Pool{
        New: func() interface{} {
            return &CompletionRequest{
                Messages: make([]Message, 0, 10),
            }
        },
    }
    
    responsePool = sync.Pool{
        New: func() interface{} {
            return &CompletionResponse{}
        },
    }
)

func GetRequest() *CompletionRequest {
    return requestPool.Get().(*CompletionRequest)
}

func PutRequest(req *CompletionRequest) {
    // Reset request
    req.Messages = req.Messages[:0]
    req.Model = ""
    req.Temperature = 0
    req.MaxTokens = 0
    
    requestPool.Put(req)
}

// Buffer pooling for large operations
var bufferPool = sync.Pool{
    New: func() interface{} {
        return bytes.NewBuffer(make([]byte, 0, 4096))
    },
}

func processWithPooledBuffer(data []byte) ([]byte, error) {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufferPool.Put(buf)
    }()
    
    // Use buffer for processing
    if err := processData(buf, data); err != nil {
        return nil, err
    }
    
    return buf.Bytes(), nil
}
```

### Memory-Efficient Data Structures

```go
// Streaming response handler
type StreamingHandler struct {
    bufferSize int
    processor  func(chunk []byte) error
}

func (sh *StreamingHandler) HandleResponse(reader io.Reader) error {
    buffer := make([]byte, sh.bufferSize)
    
    for {
        n, err := reader.Read(buffer)
        if n > 0 {
            if err := sh.processor(buffer[:n]); err != nil {
                return err
            }
        }
        
        if err == io.EOF {
            break
        }
        
        if err != nil {
            return err
        }
    }
    
    return nil
}

// Circular buffer for message history
type CircularMessageBuffer struct {
    messages []Message
    capacity int
    head     int
    size     int
    mu       sync.RWMutex
}

func (cmb *CircularMessageBuffer) Add(msg Message) {
    cmb.mu.Lock()
    defer cmb.mu.Unlock()
    
    if cmb.size < cmb.capacity {
        cmb.messages[cmb.size] = msg
        cmb.size++
    } else {
        cmb.messages[cmb.head] = msg
        cmb.head = (cmb.head + 1) % cmb.capacity
    }
}

func (cmb *CircularMessageBuffer) GetRecent(n int) []Message {
    cmb.mu.RLock()
    defer cmb.mu.RUnlock()
    
    if n > cmb.size {
        n = cmb.size
    }
    
    result := make([]Message, n)
    for i := 0; i < n; i++ {
        idx := (cmb.head + cmb.size - n + i) % cmb.capacity
        result[i] = cmb.messages[idx]
    }
    
    return result
}
```

---

## Network Optimization

### Connection Management

```go
// Optimized HTTP client configuration
func createOptimizedClient() *http.Client {
    return &http.Client{
        Transport: &http.Transport{
            // Connection pooling
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 10,
            MaxConnsPerHost:     20,
            IdleConnTimeout:     90 * time.Second,
            
            // Timeouts
            DialContext: (&net.Dialer{
                Timeout:   30 * time.Second,
                KeepAlive: 30 * time.Second,
            }).DialContext,
            
            // TLS configuration
            TLSHandshakeTimeout: 10 * time.Second,
            TLSClientConfig: &tls.Config{
                MinVersion: tls.VersionTLS12,
            },
            
            // HTTP/2 support
            ForceAttemptHTTP2: true,
            
            // Compression
            DisableCompression: false,
        },
        
        Timeout: 60 * time.Second,
    }
}

// Request compression
func compressRequest(data []byte) ([]byte, error) {
    var buf bytes.Buffer
    gz := gzip.NewWriter(&buf)
    
    if _, err := gz.Write(data); err != nil {
        return nil, err
    }
    
    if err := gz.Close(); err != nil {
        return nil, err
    }
    
    return buf.Bytes(), nil
}

// Keep-alive monitoring
type ConnectionMonitor struct {
    client    *http.Client
    checkInterval time.Duration
}

func (cm *ConnectionMonitor) Start(ctx context.Context) {
    ticker := time.NewTicker(cm.checkInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            cm.checkConnections()
        case <-ctx.Done():
            return
        }
    }
}
```

### Request Deduplication

```go
// Prevent duplicate requests
type RequestDeduplicator struct {
    inFlight map[string][]chan Result
    mu       sync.Mutex
}

func (rd *RequestDeduplicator) Execute(key string, fn func() (interface{}, error)) (interface{}, error) {
    rd.mu.Lock()
    
    if callbacks, exists := rd.inFlight[key]; exists {
        // Request in flight, wait for result
        ch := make(chan Result, 1)
        rd.inFlight[key] = append(callbacks, ch)
        rd.mu.Unlock()
        
        result := <-ch
        return result.Data, result.Error
    }
    
    // First request, execute
    rd.inFlight[key] = make([]chan Result, 0)
    rd.mu.Unlock()
    
    data, err := fn()
    
    rd.mu.Lock()
    callbacks := rd.inFlight[key]
    delete(rd.inFlight, key)
    rd.mu.Unlock()
    
    // Notify waiting requests
    result := Result{Data: data, Error: err}
    for _, ch := range callbacks {
        ch <- result
        close(ch)
    }
    
    return data, err
}
```

---

## Cost Optimization

### Token Usage Optimization

```go
// Token-aware request optimization
type TokenOptimizer struct {
    maxTokens      int
    costPerToken   float64
    summarizer     Summarizer
}

func (to *TokenOptimizer) OptimizeRequest(req *CompletionRequest) (*CompletionRequest, error) {
    totalTokens := 0
    
    // Count tokens in messages
    for i, msg := range req.Messages {
        tokens := EstimateTokens(msg.Content)
        totalTokens += tokens
        
        // Compress if too long
        if tokens > to.maxTokens/2 {
            summary, err := to.summarizer.Summarize(msg.Content, to.maxTokens/4)
            if err != nil {
                return nil, err
            }
            req.Messages[i].Content = summary
        }
    }
    
    // Trim conversation history if needed
    if totalTokens > to.maxTokens {
        req.Messages = to.trimConversation(req.Messages, to.maxTokens)
    }
    
    return req, nil
}

// Cost tracking and budgeting
type CostTracker struct {
    usage    map[string]*UsageStats
    budgets  map[string]float64
    mu       sync.RWMutex
}

func (ct *CostTracker) Track(userID string, tokens int, model string) error {
    ct.mu.Lock()
    defer ct.mu.Unlock()
    
    if ct.usage[userID] == nil {
        ct.usage[userID] = &UsageStats{}
    }
    
    cost := calculateCost(tokens, model)
    ct.usage[userID].Tokens += tokens
    ct.usage[userID].Cost += cost
    
    // Check budget
    if budget, ok := ct.budgets[userID]; ok {
        if ct.usage[userID].Cost > budget {
            return errors.New("budget exceeded")
        }
    }
    
    return nil
}
```

---

## Monitoring and Metrics

### Performance Metrics Collection

```go
// Comprehensive metrics collector
type MetricsCollector struct {
    requestDuration   *prometheus.HistogramVec
    requestsTotal     *prometheus.CounterVec
    tokensUsed        *prometheus.CounterVec
    cacheHitRate      *prometheus.GaugeVec
    activeConnections *prometheus.GaugeVec
    errorRate         *prometheus.CounterVec
}

func NewMetricsCollector() *MetricsCollector {
    return &MetricsCollector{
        requestDuration: prometheus.NewHistogramVec(
            prometheus.HistogramOpts{
                Name:    "llm_request_duration_seconds",
                Help:    "Duration of LLM requests",
                Buckets: prometheus.ExponentialBuckets(0.1, 2, 10),
            },
            []string{"provider", "model", "cache_hit"},
        ),
        
        tokensUsed: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "llm_tokens_total",
                Help: "Total tokens used",
            },
            []string{"provider", "model", "type"},
        ),
        
        cacheHitRate: prometheus.NewGaugeVec(
            prometheus.GaugeOpts{
                Name: "llm_cache_hit_rate",
                Help: "Cache hit rate",
            },
            []string{"cache_level"},
        ),
    }
}

// Real-time performance monitoring
func (mc *MetricsCollector) RecordRequest(provider, model string, duration time.Duration, cacheHit bool) {
    labels := prometheus.Labels{
        "provider":  provider,
        "model":     model,
        "cache_hit": fmt.Sprintf("%t", cacheHit),
    }
    
    mc.requestDuration.With(labels).Observe(duration.Seconds())
    mc.requestsTotal.With(labels).Inc()
}
```

### Performance Alerting

```go
// Alert on performance degradation
type PerformanceMonitor struct {
    thresholds Thresholds
    alerter    Alerter
}

type Thresholds struct {
    MaxLatency      time.Duration
    MinCacheHitRate float64
    MaxErrorRate    float64
    MaxMemoryUsage  uint64
}

func (pm *PerformanceMonitor) Check() {
    // Check latency
    if avgLatency := getAverageLatency(); avgLatency > pm.thresholds.MaxLatency {
        pm.alerter.Alert("High latency detected", map[string]interface{}{
            "current":   avgLatency,
            "threshold": pm.thresholds.MaxLatency,
}
    }
    
    // Check cache performance
    if hitRate := getCacheHitRate(); hitRate < pm.thresholds.MinCacheHitRate {
        pm.alerter.Alert("Low cache hit rate", map[string]interface{}{
            "current":   hitRate,
            "threshold": pm.thresholds.MinCacheHitRate,
}
    }
    
    // Check memory usage
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    if m.Alloc > pm.thresholds.MaxMemoryUsage {
        pm.alerter.Alert("High memory usage", map[string]interface{}{
            "current":   m.Alloc,
            "threshold": pm.thresholds.MaxMemoryUsage,
}
    }
}
```

---

## Best Practices Summary

### 1. Measure Before Optimizing
- Always profile and benchmark before making changes
- Focus on bottlenecks identified by profiling
- Set clear performance targets

### 2. Optimize at Multiple Levels
- Request level: Prompt optimization, batching
- Application level: Caching, pooling
- System level: Connection management, resource limits

### 3. Monitor Continuously
- Track key metrics in production
- Set up alerts for performance degradation
- Regular performance reviews

### 4. Balance Trade-offs
- Performance vs. cost
- Latency vs. throughput
- Memory vs. CPU usage

### 5. Test Under Load
- Simulate production workloads
- Test with concurrent users
- Verify optimization effectiveness

---

## Performance Tuning Checklist

- [ ] **Profiling Setup**
  - [ ] CPU profiling enabled
  - [ ] Memory profiling configured
  - [ ] Tracing instrumented
  - [ ] Benchmarks written

- [ ] **Request Optimization**
  - [ ] Prompt compression implemented
  - [ ] Token counting accurate
  - [ ] Model selection optimized
  - [ ] Batching configured

- [ ] **Caching Strategy**
  - [ ] Multi-level cache deployed
  - [ ] Cache keys optimized
  - [ ] TTLs configured appropriately
  - [ ] Cache warming implemented

- [ ] **Concurrency Tuning**
  - [ ] Worker pool sized correctly
  - [ ] Rate limiting configured
  - [ ] Connection pooling optimized
  - [ ] Deadlock prevention verified

- [ ] **Resource Management**
  - [ ] Object pooling implemented
  - [ ] Memory limits set
  - [ ] GC tuning applied
  - [ ] Resource monitoring active

- [ ] **Network Optimization**
  - [ ] HTTP/2 enabled
  - [ ] Keep-alive configured
  - [ ] Compression enabled
  - [ ] Timeouts appropriate

- [ ] **Monitoring & Alerts**
  - [ ] Metrics collection active
  - [ ] Performance alerts configured
  - [ ] Dashboards created
  - [ ] SLOs defined

---

## Next Steps

- **[Production Deployment](production-deployment.md)** - Deploy optimized applications
- **[Security Considerations](security-considerations.md)** - Security and performance balance
- **[Troubleshooting Guide](troubleshooting.md)** - Debug performance issues
- **[Best Practices Checklist](../../user-guide/reference/best-practices-checklist.md)** - Complete production checklist
- **[Provider Comparison](../../user-guide/reference/provider-comparison.md)** - Provider performance characteristics