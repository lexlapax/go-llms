# Performance: Optimization Strategies and Benchmarking

> **[Project Root](/) / [Documentation](/docs/) / [Technical Documentation](/docs/technical/) / [Advanced Topics](/docs/technical/advanced/) / Performance**

Comprehensive guide to performance optimization in Go-LLMs, covering profiling techniques, memory management, concurrency optimization, caching strategies, request batching, connection pooling, and systematic benchmarking for building high-performance LLM applications.

## Performance Profiling and Analysis

### 1. Built-in Profiling Infrastructure

```go
// PerformanceProfiler provides comprehensive profiling capabilities
type PerformanceProfiler struct {
    enabled     bool
    cpuProfile  *os.File
    memProfile  *os.File
    traceFile   *os.File
    metrics     *MetricsCollector
    mu          sync.RWMutex
}

// NewPerformanceProfiler creates a new profiler instance
func NewPerformanceProfiler(config ProfilerConfig) *PerformanceProfiler {
    return &PerformanceProfiler{
        enabled: config.Enabled,
        metrics: NewMetricsCollector(),
    }
}

type ProfilerConfig struct {
    Enabled        bool   `yaml:"enabled" json:"enabled"`
    CPUProfile     string `yaml:"cpu_profile,omitempty" json:"cpu_profile,omitempty"`
    MemProfile     string `yaml:"mem_profile,omitempty" json:"mem_profile,omitempty"`
    TraceFile      string `yaml:"trace_file,omitempty" json:"trace_file,omitempty"`
    SampleRate     int    `yaml:"sample_rate" json:"sample_rate" default:"100"`
    ProfileDuration time.Duration `yaml:"profile_duration" json:"profile_duration" default:"30s"`
}

// StartProfiling begins performance profiling
func (p *PerformanceProfiler) StartProfiling(ctx context.Context) error {
    if !p.enabled {
        return nil
    }
    
    p.mu.Lock()
    defer p.mu.Unlock()
    
    // Start CPU profiling
    if p.cpuProfile != nil {
        if err := pprof.StartCPUProfile(p.cpuProfile); err != nil {
            return fmt.Errorf("failed to start CPU profile: %w", err)
        }
    }
    
    // Start execution trace
    if p.traceFile != nil {
        if err := trace.Start(p.traceFile); err != nil {
            return fmt.Errorf("failed to start execution trace: %w", err)
        }
    }
    
    // Start runtime metrics collection
    go p.collectRuntimeMetrics(ctx)
    
    return nil
}

// StopProfiling stops profiling and writes results
func (p *PerformanceProfiler) StopProfiling() error {
    if !p.enabled {
        return nil
    }
    
    p.mu.Lock()
    defer p.mu.Unlock()
    
    // Stop CPU profiling
    pprof.StopCPUProfile()
    
    // Stop execution trace
    trace.Stop()
    
    // Write memory profile
    if p.memProfile != nil {
        runtime.GC() // Force GC before memory profile
        if err := pprof.WriteHeapProfile(p.memProfile); err != nil {
            return fmt.Errorf("failed to write memory profile: %w", err)
        }
    }
    
    return nil
}

// collectRuntimeMetrics collects Go runtime metrics
func (p *PerformanceProfiler) collectRuntimeMetrics(ctx context.Context) {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    
    var memStats runtime.MemStats
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            runtime.ReadMemStats(&memStats)
            
            p.metrics.RecordGauge("memory.heap_alloc", float64(memStats.HeapAlloc))
            p.metrics.RecordGauge("memory.heap_sys", float64(memStats.HeapSys))
            p.metrics.RecordGauge("memory.heap_idle", float64(memStats.HeapIdle))
            p.metrics.RecordGauge("memory.heap_inuse", float64(memStats.HeapInuse))
            p.metrics.RecordGauge("memory.stack_inuse", float64(memStats.StackInuse))
            p.metrics.RecordGauge("memory.stack_sys", float64(memStats.StackSys))
            p.metrics.RecordCounter("memory.gc_runs", float64(memStats.NumGC))
            p.metrics.RecordGauge("memory.gc_pause_ns", float64(memStats.PauseNs[(memStats.NumGC+255)%256]))
            p.metrics.RecordGauge("goroutines.count", float64(runtime.NumGoroutine()))
        }
    }
}

// ProfiledProvider wraps a provider with performance monitoring
type ProfiledProvider struct {
    Provider
    profiler *PerformanceProfiler
    metrics  *ProviderMetrics
}

type ProviderMetrics struct {
    RequestCount    *Counter
    RequestDuration *Histogram
    RequestSize     *Histogram
    ResponseSize    *Histogram
    ErrorCount      *Counter
    TokensProcessed *Counter
}

// Complete wraps provider completion with profiling
func (p *ProfiledProvider) Complete(ctx context.Context, request *CompletionRequest) (*CompletionResponse, error) {
    start := time.Now()
    
    // Record request metrics
    p.metrics.RequestCount.Inc()
    requestSize := calculateRequestSize(request)
    p.metrics.RequestSize.Observe(float64(requestSize))
    
    defer func() {
        duration := time.Since(start)
        p.metrics.RequestDuration.Observe(duration.Seconds())
    }()
    
    // Execute request
    response, err := p.Provider.Complete(ctx, request)
    
    if err != nil {
        p.metrics.ErrorCount.Inc()
        return nil, err
    }
    
    // Record response metrics
    responseSize := calculateResponseSize(response)
    p.metrics.ResponseSize.Observe(float64(responseSize))
    
    if response.Usage != nil {
        p.metrics.TokensProcessed.Add(float64(response.Usage.TotalTokens))
    }
    
    return response, nil
}
```

### 2. Request Tracing and Analysis

```go
// RequestTracer provides detailed request tracing
type RequestTracer struct {
    spans    map[string]*TraceSpan
    mu       sync.RWMutex
    exporter TraceExporter
}

type TraceSpan struct {
    ID         string                 `json:"id"`
    ParentID   string                 `json:"parent_id,omitempty"`
    Operation  string                 `json:"operation"`
    StartTime  time.Time              `json:"start_time"`
    EndTime    *time.Time             `json:"end_time,omitempty"`
    Duration   time.Duration          `json:"duration"`
    Tags       map[string]interface{} `json:"tags,omitempty"`
    Logs       []TraceLog             `json:"logs,omitempty"`
    Status     SpanStatus             `json:"status"`
}

type TraceLog struct {
    Timestamp time.Time              `json:"timestamp"`
    Level     string                 `json:"level"`
    Message   string                 `json:"message"`
    Fields    map[string]interface{} `json:"fields,omitempty"`
}

type SpanStatus string

const (
    SpanStatusOK    SpanStatus = "ok"
    SpanStatusError SpanStatus = "error"
)

// StartSpan creates a new trace span
func (t *RequestTracer) StartSpan(operation string, parentID string) *TraceSpan {
    span := &TraceSpan{
        ID:        generateSpanID(),
        ParentID:  parentID,
        Operation: operation,
        StartTime: time.Now(),
        Tags:      make(map[string]interface{}),
        Status:    SpanStatusOK,
    }
    
    t.mu.Lock()
    t.spans[span.ID] = span
    t.mu.Unlock()
    
    return span
}

// FinishSpan completes a trace span
func (t *RequestTracer) FinishSpan(spanID string, status SpanStatus) {
    t.mu.Lock()
    defer t.mu.Unlock()
    
    span, exists := t.spans[spanID]
    if !exists {
        return
    }
    
    now := time.Now()
    span.EndTime = &now
    span.Duration = now.Sub(span.StartTime)
    span.Status = status
    
    // Export completed span
    if t.exporter != nil {
        t.exporter.Export(span)
    }
    
    delete(t.spans, spanID)
}

// TracedAgent wraps an agent with request tracing
type TracedAgent struct {
    Agent
    tracer *RequestTracer
}

func (a *TracedAgent) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    span := a.tracer.StartSpan("agent.execute", "")
    defer func() {
        status := SpanStatusOK
        if err := recover(); err != nil {
            status = SpanStatusError
            panic(err)
        }
        a.tracer.FinishSpan(span.ID, status)
    }()
    
    span.Tags["input_type"] = fmt.Sprintf("%T", input)
    span.Tags["agent_type"] = fmt.Sprintf("%T", a.Agent)
    
    return a.Agent.Execute(ctx, input)
}
```

## Memory Optimization

### 1. Object Pooling

```go
// ObjectPool provides reusable object pooling
type ObjectPool[T any] struct {
    pool    sync.Pool
    factory func() T
    reset   func(T)
    metrics *PoolMetrics
}

type PoolMetrics struct {
    Gets     int64
    Puts     int64
    News     int64
    Misses   int64
    Size     int64
    MaxSize  int64
}

// NewObjectPool creates a new object pool
func NewObjectPool[T any](factory func() T, reset func(T)) *ObjectPool[T] {
    return &ObjectPool[T]{
        pool: sync.Pool{
            New: func() interface{} {
                atomic.AddInt64(&pool.metrics.News, 1)
                return factory()
            },
        },
        factory: factory,
        reset:   reset,
        metrics: &PoolMetrics{},
    }
}

// Get retrieves an object from the pool
func (p *ObjectPool[T]) Get() T {
    atomic.AddInt64(&p.metrics.Gets, 1)
    
    obj := p.pool.Get().(T)
    if p.reset != nil {
        p.reset(obj)
    }
    
    return obj
}

// Put returns an object to the pool
func (p *ObjectPool[T]) Put(obj T) {
    atomic.AddInt64(&p.metrics.Puts, 1)
    p.pool.Put(obj)
}

// Common pools for frequent allocations
var (
    // Buffer pool for byte operations
    BufferPool = NewObjectPool(
        func() *bytes.Buffer {
            return &bytes.Buffer{}
        },
        func(b *bytes.Buffer) {
            b.Reset()
        },
    )
    
    // Request pool for HTTP requests
    RequestPool = NewObjectPool(
        func() *CompletionRequest {
            return &CompletionRequest{
                Messages: make([]Message, 0, 4),
            }
        },
        func(r *CompletionRequest) {
            r.Messages = r.Messages[:0]
            r.Model = ""
            r.Temperature = nil
            r.MaxTokens = nil
        },
    )
    
    // Response pool for HTTP responses
    ResponsePool = NewObjectPool(
        func() *CompletionResponse {
            return &CompletionResponse{}
        },
        func(r *CompletionResponse) {
            r.Content = ""
            r.Model = ""
            r.Usage = nil
            r.ToolCalls = r.ToolCalls[:0]
        },
    )
)

// Optimized provider using pools
type PooledProvider struct {
    Provider
    requestPool  *ObjectPool[*CompletionRequest]
    responsePool *ObjectPool[*CompletionResponse]
    bufferPool   *ObjectPool[*bytes.Buffer]
}

func (p *PooledProvider) Complete(ctx context.Context, request *CompletionRequest) (*CompletionResponse, error) {
    // Get pooled objects
    pooledRequest := p.requestPool.Get()
    pooledResponse := p.responsePool.Get()
    buffer := p.bufferPool.Get()
    
    defer func() {
        p.requestPool.Put(pooledRequest)
        p.responsePool.Put(pooledResponse)
        p.bufferPool.Put(buffer)
    }()
    
    // Copy request data to pooled object
    copyRequest(request, pooledRequest)
    
    // Execute with pooled objects
    result, err := p.Provider.Complete(ctx, pooledRequest)
    if err != nil {
        return nil, err
    }
    
    // Copy result to fresh response object
    response := &CompletionResponse{}
    copyResponse(result, response)
    
    return response, nil
}
```

### 2. Memory-Efficient Data Structures

```go
// CompactMessage uses string interning for common values
type CompactMessage struct {
    Role    InternedString `json:"role"`
    Content InternedString `json:"content"`
}

// StringInterner reduces memory usage for repeated strings
type StringInterner struct {
    strings map[string]string
    mu      sync.RWMutex
}

type InternedString struct {
    value string
}

var globalInterner = &StringInterner{
    strings: make(map[string]string),
}

// Intern returns a canonical representation of the string
func (si *StringInterner) Intern(s string) InternedString {
    si.mu.RLock()
    if interned, exists := si.strings[s]; exists {
        si.mu.RUnlock()
        return InternedString{value: interned}
    }
    si.mu.RUnlock()
    
    si.mu.Lock()
    defer si.mu.Unlock()
    
    // Double-check after acquiring write lock
    if interned, exists := si.strings[s]; exists {
        return InternedString{value: interned}
    }
    
    // Store the string
    si.strings[s] = s
    return InternedString{value: s}
}

func (is InternedString) String() string {
    return is.value
}

// Efficient slice operations
type CompactSlice[T any] struct {
    data     []T
    length   int
    capacity int
}

// NewCompactSlice creates a slice with pre-allocated capacity
func NewCompactSlice[T any](capacity int) *CompactSlice[T] {
    return &CompactSlice[T]{
        data:     make([]T, 0, capacity),
        capacity: capacity,
    }
}

// Append adds elements efficiently
func (cs *CompactSlice[T]) Append(items ...T) {
    needed := cs.length + len(items)
    
    // Grow if necessary
    if needed > cs.capacity {
        newCapacity := cs.capacity * 2
        if newCapacity < needed {
            newCapacity = needed
        }
        
        newData := make([]T, cs.length, newCapacity)
        copy(newData, cs.data[:cs.length])
        cs.data = newData
        cs.capacity = newCapacity
    }
    
    cs.data = cs.data[:needed]
    copy(cs.data[cs.length:], items)
    cs.length = needed
}

// Reset clears the slice without deallocating
func (cs *CompactSlice[T]) Reset() {
    cs.data = cs.data[:0]
    cs.length = 0
}
```

## Concurrency Optimization

### 1. Worker Pool Pattern

```go
// WorkerPool manages concurrent request processing
type WorkerPool struct {
    workers    []*Worker
    jobQueue   chan Job
    resultChan chan Result
    ctx        context.Context
    cancel     context.CancelFunc
    wg         sync.WaitGroup
    metrics    *WorkerPoolMetrics
}

type Job struct {
    ID      string
    Type    JobType
    Payload interface{}
    Context context.Context
}

type Result struct {
    JobID   string
    Data    interface{}
    Error   error
    Metrics JobMetrics
}

type JobType string

const (
    JobTypeCompletion JobType = "completion"
    JobTypeEmbedding  JobType = "embedding"
    JobTypeToolCall   JobType = "tool_call"
)

type WorkerPoolMetrics struct {
    ActiveWorkers   int64
    QueuedJobs     int64
    CompletedJobs  int64
    FailedJobs     int64
    AverageLatency time.Duration
}

// NewWorkerPool creates a worker pool
func NewWorkerPool(size int, queueSize int) *WorkerPool {
    ctx, cancel := context.WithCancel(context.Background())
    
    pool := &WorkerPool{
        workers:    make([]*Worker, size),
        jobQueue:   make(chan Job, queueSize),
        resultChan: make(chan Result, queueSize),
        ctx:        ctx,
        cancel:     cancel,
        metrics:    &WorkerPoolMetrics{},
    }
    
    // Start workers
    for i := 0; i < size; i++ {
        worker := NewWorker(i, pool.jobQueue, pool.resultChan)
        pool.workers[i] = worker
        pool.wg.Add(1)
        go pool.runWorker(worker)
    }
    
    return pool
}

// Submit adds a job to the queue
func (wp *WorkerPool) Submit(job Job) error {
    select {
    case wp.jobQueue <- job:
        atomic.AddInt64(&wp.metrics.QueuedJobs, 1)
        return nil
    case <-wp.ctx.Done():
        return wp.ctx.Err()
    default:
        return fmt.Errorf("job queue is full")
    }
}

// Results returns the result channel
func (wp *WorkerPool) Results() <-chan Result {
    return wp.resultChan
}

// runWorker executes a worker
func (wp *WorkerPool) runWorker(worker *Worker) {
    defer wp.wg.Done()
    atomic.AddInt64(&wp.metrics.ActiveWorkers, 1)
    defer atomic.AddInt64(&wp.metrics.ActiveWorkers, -1)
    
    for {
        select {
        case job := <-wp.jobQueue:
            atomic.AddInt64(&wp.metrics.QueuedJobs, -1)
            
            start := time.Now()
            result := worker.Process(job)
            duration := time.Since(start)
            
            result.Metrics = JobMetrics{
                Duration:  duration,
                WorkerID:  worker.ID,
                StartTime: start,
                EndTime:   time.Now(),
            }
            
            if result.Error != nil {
                atomic.AddInt64(&wp.metrics.FailedJobs, 1)
            } else {
                atomic.AddInt64(&wp.metrics.CompletedJobs, 1)
            }
            
            select {
            case wp.resultChan <- result:
            case <-wp.ctx.Done():
                return
            }
            
        case <-wp.ctx.Done():
            return
        }
    }
}

// Worker processes jobs
type Worker struct {
    ID        int
    processor JobProcessor
}

type JobProcessor interface {
    Process(job Job) Result
}

// LLMJobProcessor handles LLM-related jobs
type LLMJobProcessor struct {
    provider Provider
    tools    ToolRegistry
}

func (p *LLMJobProcessor) Process(job Job) Result {
    switch job.Type {
    case JobTypeCompletion:
        request := job.Payload.(*CompletionRequest)
        response, err := p.provider.Complete(job.Context, request)
        return Result{
            JobID: job.ID,
            Data:  response,
            Error: err,
        }
        
    case JobTypeToolCall:
        toolCall := job.Payload.(*ToolCall)
        tool, err := p.tools.Get(toolCall.Name)
        if err != nil {
            return Result{JobID: job.ID, Error: err}
        }
        
        result, err := tool.Execute(job.Context, toolCall.Arguments)
        return Result{
            JobID: job.ID,
            Data:  result,
            Error: err,
        }
        
    default:
        return Result{
            JobID: job.ID,
            Error: fmt.Errorf("unknown job type: %s", job.Type),
        }
    }
}
```

### 2. Parallel Request Processing

```go
// ParallelProcessor handles concurrent request processing
type ParallelProcessor struct {
    maxConcurrency int
    semaphore      chan struct{}
    metrics        *ParallelMetrics
}

type ParallelMetrics struct {
    ActiveRequests int64
    TotalRequests  int64
    TotalLatency   time.Duration
}

// NewParallelProcessor creates a parallel processor
func NewParallelProcessor(maxConcurrency int) *ParallelProcessor {
    return &ParallelProcessor{
        maxConcurrency: maxConcurrency,
        semaphore:      make(chan struct{}, maxConcurrency),
        metrics:        &ParallelMetrics{},
    }
}

// ProcessBatch processes multiple requests concurrently
func (pp *ParallelProcessor) ProcessBatch(ctx context.Context, requests []*CompletionRequest, provider Provider) ([]*CompletionResponse, error) {
    if len(requests) == 0 {
        return nil, nil
    }
    
    results := make([]*CompletionResponse, len(requests))
    errors := make([]error, len(requests))
    
    var wg sync.WaitGroup
    
    for i, request := range requests {
        wg.Add(1)
        go func(index int, req *CompletionRequest) {
            defer wg.Done()
            
            // Acquire semaphore
            select {
            case pp.semaphore <- struct{}{}:
                defer func() { <-pp.semaphore }()
            case <-ctx.Done():
                errors[index] = ctx.Err()
                return
            }
            
            atomic.AddInt64(&pp.metrics.ActiveRequests, 1)
            atomic.AddInt64(&pp.metrics.TotalRequests, 1)
            
            start := time.Now()
            defer func() {
                duration := time.Since(start)
                atomic.AddInt64(&pp.metrics.ActiveRequests, -1)
                atomic.AddInt64((*int64)(&pp.metrics.TotalLatency), int64(duration))
            }()
            
            response, err := provider.Complete(ctx, req)
            results[index] = response
            errors[index] = err
        }(i, request)
    }
    
    wg.Wait()
    
    // Check for errors
    var firstError error
    for _, err := range errors {
        if err != nil && firstError == nil {
            firstError = err
        }
    }
    
    return results, firstError
}

// StreamingParallelProcessor handles parallel streaming requests
type StreamingParallelProcessor struct {
    maxStreams int
    semaphore  chan struct{}
}

func (spp *StreamingParallelProcessor) ProcessStreams(ctx context.Context, requests []*CompletionRequest, provider StreamingProvider) (<-chan StreamResult, error) {
    resultChan := make(chan StreamResult, len(requests)*10) // Buffer for chunks
    
    var wg sync.WaitGroup
    
    for i, request := range requests {
        wg.Add(1)
        go func(index int, req *CompletionRequest) {
            defer wg.Done()
            
            // Acquire semaphore
            select {
            case spp.semaphore <- struct{}{}:
                defer func() { <-spp.semaphore }()
            case <-ctx.Done():
                resultChan <- StreamResult{
                    RequestIndex: index,
                    Error:        ctx.Err(),
                    Final:        true,
                }
                return
            }
            
            stream, err := provider.CompleteStream(ctx, req)
            if err != nil {
                resultChan <- StreamResult{
                    RequestIndex: index,
                    Error:        err,
                    Final:        true,
                }
                return
            }
            
            // Process stream chunks
            for chunk := range stream {
                select {
                case resultChan <- StreamResult{
                    RequestIndex: index,
                    Chunk:        chunk,
                    Final:        false,
                }:
                case <-ctx.Done():
                    return
                }
            }
            
            // Send final marker
            resultChan <- StreamResult{
                RequestIndex: index,
                Final:        true,
            }
        }(i, request)
    }
    
    go func() {
        wg.Wait()
        close(resultChan)
    }()
    
    return resultChan, nil
}

type StreamResult struct {
    RequestIndex int
    Chunk        StreamChunk
    Error        error
    Final        bool
}
```

## Caching Strategies

### 1. Multi-Level Caching

```go
// CacheManager implements multi-level caching
type CacheManager struct {
    l1Cache *LRUCache     // In-memory cache
    l2Cache *RedisCache   // Distributed cache
    l3Cache *FileCache    // Persistent cache
    config  CacheConfig
    metrics *CacheMetrics
}

type CacheConfig struct {
    L1Size        int           `yaml:"l1_size" json:"l1_size"`
    L2TTL         time.Duration `yaml:"l2_ttl" json:"l2_ttl"`
    L3TTL         time.Duration `yaml:"l3_ttl" json:"l3_ttl"`
    EnableL1      bool          `yaml:"enable_l1" json:"enable_l1"`
    EnableL2      bool          `yaml:"enable_l2" json:"enable_l2"`
    EnableL3      bool          `yaml:"enable_l3" json:"enable_l3"`
    Compression   bool          `yaml:"compression" json:"compression"`
    Serialization string        `yaml:"serialization" json:"serialization"`
}

type CacheMetrics struct {
    L1Hits   int64
    L1Misses int64
    L2Hits   int64
    L2Misses int64
    L3Hits   int64
    L3Misses int64
    Evictions int64
}

// Get retrieves a value from the cache hierarchy
func (cm *CacheManager) Get(key string) (interface{}, bool) {
    // Try L1 cache first
    if cm.config.EnableL1 {
        if value, found := cm.l1Cache.Get(key); found {
            atomic.AddInt64(&cm.metrics.L1Hits, 1)
            return value, true
        }
        atomic.AddInt64(&cm.metrics.L1Misses, 1)
    }
    
    // Try L2 cache
    if cm.config.EnableL2 {
        if value, found := cm.l2Cache.Get(key); found {
            atomic.AddInt64(&cm.metrics.L2Hits, 1)
            
            // Populate L1 cache
            if cm.config.EnableL1 {
                cm.l1Cache.Set(key, value)
            }
            
            return value, true
        }
        atomic.AddInt64(&cm.metrics.L2Misses, 1)
    }
    
    // Try L3 cache
    if cm.config.EnableL3 {
        if value, found := cm.l3Cache.Get(key); found {
            atomic.AddInt64(&cm.metrics.L3Hits, 1)
            
            // Populate higher-level caches
            if cm.config.EnableL2 {
                cm.l2Cache.Set(key, value, cm.config.L2TTL)
            }
            if cm.config.EnableL1 {
                cm.l1Cache.Set(key, value)
            }
            
            return value, true
        }
        atomic.AddInt64(&cm.metrics.L3Misses, 1)
    }
    
    return nil, false
}

// Set stores a value in all enabled cache levels
func (cm *CacheManager) Set(key string, value interface{}) {
    if cm.config.EnableL1 {
        cm.l1Cache.Set(key, value)
    }
    
    if cm.config.EnableL2 {
        cm.l2Cache.Set(key, value, cm.config.L2TTL)
    }
    
    if cm.config.EnableL3 {
        cm.l3Cache.Set(key, value, cm.config.L3TTL)
    }
}

// CachedProvider implements caching for LLM requests
type CachedProvider struct {
    Provider
    cache  *CacheManager
    hasher ContentHasher
}

func (cp *CachedProvider) Complete(ctx context.Context, request *CompletionRequest) (*CompletionResponse, error) {
    // Generate cache key
    cacheKey, err := cp.hasher.Hash(request)
    if err != nil {
        // Fall back to direct request if hashing fails
        return cp.Provider.Complete(ctx, request)
    }
    
    // Check cache
    if cached, found := cp.cache.Get(cacheKey); found {
        response := cached.(*CompletionResponse)
        
        // Add cache metadata
        if response.Metadata == nil {
            response.Metadata = make(map[string]interface{})
        }
        response.Metadata["cached"] = true
        response.Metadata["cache_key"] = cacheKey
        
        return response, nil
    }
    
    // Cache miss - execute request
    response, err := cp.Provider.Complete(ctx, request)
    if err != nil {
        return nil, err
    }
    
    // Cache successful response
    if response != nil {
        cp.cache.Set(cacheKey, response)
    }
    
    return response, nil
}
```

### 2. Semantic Caching

```go
// SemanticCache provides content-aware caching
type SemanticCache struct {
    vectorStore   VectorStore
    cache        *CacheManager
    embedder     EmbeddingProvider
    threshold    float64
    maxResults   int
}

type SemanticCacheConfig struct {
    SimilarityThreshold float64 `yaml:"similarity_threshold" json:"similarity_threshold"`
    MaxResults         int     `yaml:"max_results" json:"max_results"`
    EmbeddingModel     string  `yaml:"embedding_model" json:"embedding_model"`
    VectorDimensions   int     `yaml:"vector_dimensions" json:"vector_dimensions"`
}

// Get retrieves semantically similar cached responses
func (sc *SemanticCache) Get(query string) (*CompletionResponse, float64, error) {
    // Generate embedding for query
    embedding, err := sc.embedder.CreateEmbedding(context.Background(), &EmbeddingRequest{
        Input: []string{query},
    })
    if err != nil {
        return nil, 0, fmt.Errorf("failed to create embedding: %w", err)
    }
    
    // Search for similar vectors
    results, err := sc.vectorStore.Search(embedding.Embeddings[0], sc.maxResults)
    if err != nil {
        return nil, 0, fmt.Errorf("vector search failed: %w", err)
    }
    
    // Check if best match meets threshold
    if len(results) > 0 && results[0].Score >= sc.threshold {
        // Retrieve cached response
        if cached, found := sc.cache.Get(results[0].ID); found {
            response := cached.(*CompletionResponse)
            return response, results[0].Score, nil
        }
    }
    
    return nil, 0, nil
}

// Set stores a response with its semantic vector
func (sc *SemanticCache) Set(query string, response *CompletionResponse) error {
    // Generate embedding
    embedding, err := sc.embedder.CreateEmbedding(context.Background(), &EmbeddingRequest{
        Input: []string{query},
    })
    if err != nil {
        return fmt.Errorf("failed to create embedding: %w", err)
    }
    
    // Generate unique ID
    cacheKey := generateCacheKey(query, response)
    
    // Store in vector database
    if err := sc.vectorStore.Store(cacheKey, embedding.Embeddings[0], map[string]interface{}{
        "query":     query,
        "timestamp": time.Now(),
    }); err != nil {
        return fmt.Errorf("failed to store vector: %w", err)
    }
    
    // Store response in cache
    sc.cache.Set(cacheKey, response)
    
    return nil
}
```

## Benchmarking Framework

### 1. Comprehensive Benchmarking Suite

```go
// BenchmarkSuite provides comprehensive performance testing
type BenchmarkSuite struct {
    providers   map[string]Provider
    testCases   []BenchmarkCase
    metrics     *BenchmarkMetrics
    config      BenchmarkConfig
}

type BenchmarkCase struct {
    Name        string              `yaml:"name" json:"name"`
    Description string              `yaml:"description" json:"description"`
    Request     *CompletionRequest  `yaml:"request" json:"request"`
    Iterations  int                 `yaml:"iterations" json:"iterations"`
    Concurrent  int                 `yaml:"concurrent" json:"concurrent"`
    Timeout     time.Duration       `yaml:"timeout" json:"timeout"`
    WarmupRuns  int                 `yaml:"warmup_runs" json:"warmup_runs"`
}

type BenchmarkMetrics struct {
    TotalRequests    int64
    SuccessfulReqs   int64
    FailedRequests   int64
    TotalLatency     time.Duration
    MinLatency       time.Duration
    MaxLatency       time.Duration
    P50Latency       time.Duration
    P95Latency       time.Duration
    P99Latency       time.Duration
    Throughput       float64 // requests per second
    TokensPerSecond  float64
    ErrorRate        float64
}

// RunBenchmark executes a comprehensive benchmark
func (bs *BenchmarkSuite) RunBenchmark(ctx context.Context, providerName string) (*BenchmarkResult, error) {
    provider, exists := bs.providers[providerName]
    if !exists {
        return nil, fmt.Errorf("provider %s not found", providerName)
    }
    
    result := &BenchmarkResult{
        Provider:  providerName,
        StartTime: time.Now(),
        Cases:     make([]CaseResult, len(bs.testCases)),
    }
    
    for i, testCase := range bs.testCases {
        caseResult, err := bs.runBenchmarkCase(ctx, provider, &testCase)
        if err != nil {
            return nil, fmt.Errorf("benchmark case %s failed: %w", testCase.Name, err)
        }
        
        result.Cases[i] = *caseResult
    }
    
    result.EndTime = time.Now()
    result.TotalDuration = result.EndTime.Sub(result.StartTime)
    result.Summary = bs.calculateSummary(result.Cases)
    
    return result, nil
}

// runBenchmarkCase executes a single benchmark case
func (bs *BenchmarkSuite) runBenchmarkCase(ctx context.Context, provider Provider, testCase *BenchmarkCase) (*CaseResult, error) {
    result := &CaseResult{
        Name:      testCase.Name,
        StartTime: time.Now(),
        Latencies: make([]time.Duration, 0, testCase.Iterations),
    }
    
    // Warmup runs
    if testCase.WarmupRuns > 0 {
        for i := 0; i < testCase.WarmupRuns; i++ {
            _, _ = provider.Complete(ctx, testCase.Request)
        }
    }
    
    // Actual benchmark runs
    if testCase.Concurrent > 1 {
        result = bs.runConcurrentBenchmark(ctx, provider, testCase, result)
    } else {
        result = bs.runSequentialBenchmark(ctx, provider, testCase, result)
    }
    
    result.EndTime = time.Now()
    result.Metrics = bs.calculateMetrics(result.Latencies, result.Errors)
    
    return result, nil
}

// runConcurrentBenchmark executes concurrent benchmark
func (bs *BenchmarkSuite) runConcurrentBenchmark(ctx context.Context, provider Provider, testCase *BenchmarkCase, result *CaseResult) *CaseResult {
    var wg sync.WaitGroup
    latencyMu := sync.Mutex{}
    errorMu := sync.Mutex{}
    
    requestsPerWorker := testCase.Iterations / testCase.Concurrent
    
    for worker := 0; worker < testCase.Concurrent; worker++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            
            for i := 0; i < requestsPerWorker; i++ {
                start := time.Now()
                _, err := provider.Complete(ctx, testCase.Request)
                latency := time.Since(start)
                
                latencyMu.Lock()
                result.Latencies = append(result.Latencies, latency)
                latencyMu.Unlock()
                
                if err != nil {
                    errorMu.Lock()
                    result.Errors = append(result.Errors, err)
                    errorMu.Unlock()
                }
            }
        }()
    }
    
    wg.Wait()
    return result
}

type BenchmarkResult struct {
    Provider      string        `json:"provider"`
    StartTime     time.Time     `json:"start_time"`
    EndTime       time.Time     `json:"end_time"`
    TotalDuration time.Duration `json:"total_duration"`
    Cases         []CaseResult  `json:"cases"`
    Summary       *BenchmarkMetrics `json:"summary"`
}

type CaseResult struct {
    Name      string          `json:"name"`
    StartTime time.Time       `json:"start_time"`
    EndTime   time.Time       `json:"end_time"`
    Latencies []time.Duration `json:"latencies"`
    Errors    []error         `json:"errors"`
    Metrics   *BenchmarkMetrics `json:"metrics"`
}

// calculateMetrics computes performance metrics
func (bs *BenchmarkSuite) calculateMetrics(latencies []time.Duration, errors []error) *BenchmarkMetrics {
    if len(latencies) == 0 {
        return &BenchmarkMetrics{}
    }
    
    // Sort latencies for percentile calculations
    sort.Slice(latencies, func(i, j int) bool {
        return latencies[i] < latencies[j]
    })
    
    var totalLatency time.Duration
    for _, latency := range latencies {
        totalLatency += latency
    }
    
    metrics := &BenchmarkMetrics{
        TotalRequests:  int64(len(latencies)),
        FailedRequests: int64(len(errors)),
        TotalLatency:   totalLatency,
        MinLatency:     latencies[0],
        MaxLatency:     latencies[len(latencies)-1],
        P50Latency:     latencies[len(latencies)/2],
        P95Latency:     latencies[int(float64(len(latencies))*0.95)],
        P99Latency:     latencies[int(float64(len(latencies))*0.99)],
    }
    
    metrics.SuccessfulReqs = metrics.TotalRequests - metrics.FailedRequests
    metrics.ErrorRate = float64(metrics.FailedRequests) / float64(metrics.TotalRequests)
    
    if totalLatency > 0 {
        metrics.Throughput = float64(metrics.TotalRequests) / totalLatency.Seconds()
    }
    
    return metrics
}
```

### 2. Performance Regression Testing

```go
// RegressionTester detects performance regressions
type RegressionTester struct {
    baseline    *BenchmarkResult
    threshold   float64  // Acceptable performance degradation (e.g., 0.1 for 10%)
    metrics     []string // Metrics to compare
}

// CompareResults compares benchmark results against baseline
func (rt *RegressionTester) CompareResults(current *BenchmarkResult) (*RegressionReport, error) {
    if rt.baseline == nil {
        return nil, fmt.Errorf("no baseline results available")
    }
    
    report := &RegressionReport{
        BaselineTime: rt.baseline.StartTime,
        CurrentTime:  current.StartTime,
        Comparisons:  make([]MetricComparison, 0),
        HasRegression: false,
    }
    
    // Compare each test case
    for i, currentCase := range current.Cases {
        if i >= len(rt.baseline.Cases) {
            continue
        }
        
        baselineCase := rt.baseline.Cases[i]
        comparison := rt.compareCase(&baselineCase, &currentCase)
        report.Comparisons = append(report.Comparisons, comparison)
        
        if comparison.Regression {
            report.HasRegression = true
        }
    }
    
    return report, nil
}

type RegressionReport struct {
    BaselineTime  time.Time          `json:"baseline_time"`
    CurrentTime   time.Time          `json:"current_time"`
    Comparisons   []MetricComparison `json:"comparisons"`
    HasRegression bool               `json:"has_regression"`
}

type MetricComparison struct {
    CaseName      string  `json:"case_name"`
    MetricName    string  `json:"metric_name"`
    BaselineValue float64 `json:"baseline_value"`
    CurrentValue  float64 `json:"current_value"`
    Change        float64 `json:"change"`        // Percentage change
    Regression    bool    `json:"regression"`    // Whether this is a regression
    Severity      string  `json:"severity"`     // minor, major, critical
}

// compareCase compares metrics between two test cases
func (rt *RegressionTester) compareCase(baseline, current *CaseResult) MetricComparison {
    // Compare P95 latency as primary metric
    baselineP95 := baseline.Metrics.P95Latency.Seconds()
    currentP95 := current.Metrics.P95Latency.Seconds()
    
    change := (currentP95 - baselineP95) / baselineP95
    regression := change > rt.threshold
    
    severity := "minor"
    if change > 0.2 {
        severity = "major"
    }
    if change > 0.5 {
        severity = "critical"
    }
    
    return MetricComparison{
        CaseName:      current.Name,
        MetricName:    "p95_latency",
        BaselineValue: baselineP95,
        CurrentValue:  currentP95,
        Change:        change,
        Regression:    regression,
        Severity:      severity,
    }
}
```

This comprehensive performance guide provides the tools and strategies needed to build high-performance LLM applications with Go-LLMs, covering everything from profiling and memory optimization to advanced caching and systematic benchmarking.