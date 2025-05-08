# Concurrency Patterns in Go-LLMs

> **[Documentation Home](/REFERENCE.md) / [Technical Documentation](/docs/technical/) / Concurrency Patterns**

This document describes the concurrency patterns implemented in Go-LLMs to ensure thread safety, optimize performance, and handle streaming operations effectively. It covers design decisions, implementation details, and best practices.

*Related: [Performance Optimization](performance.md) | [Sync.Pool Implementation](sync-pool.md) | [Caching Mechanisms](caching.md)*

## Overview of Concurrency in Go-LLMs

Go-LLMs uses several concurrency patterns to address different requirements:

1. **Thread-safe data structures**: Protects shared state with mutexes and atomic operations
2. **Streaming with channels**: Efficiently handles token streaming from LLM providers
3. **Fan-out/fan-in patterns**: Processes multiple requests or responses concurrently
4. **Worker pools**: Manages resource utilization for high-throughput scenarios
5. **Singleton initialization**: Ensures thread-safe creation of shared resources

## Thread Safety Implementations

### Message Manager

The `MessageManager` (`pkg/agent/workflow/message_manager.go`) demonstrates comprehensive thread safety:

```go
// MessageManager handles efficient management of conversation message history
type MessageManager struct {
    // Current messages
    messages []ldomain.Message

    // Thread safety
    mu sync.RWMutex

    // Other fields...
}

// AddMessage adds a message to the history
func (m *MessageManager) AddMessage(message ldomain.Message) {
    m.mu.Lock()
    defer m.mu.Unlock()

    // Clone the message to prevent external modification
    newMsg := m.cloneMessage(message)

    // Add the message
    m.messages = append(m.messages, newMsg)
    m.messageTimestamps = append(m.messageTimestamps, time.Now())

    // Calculate and cache token count
    tokens := m.estimateTokens(newMsg.Content)
    m.tokenCounts[newMsg.Content] = tokens

    // Apply truncation if needed
    m.applyTruncation()
}

// GetMessages returns the current message history
func (m *MessageManager) GetMessages() []ldomain.Message {
    m.mu.RLock()
    defer m.mu.RUnlock()

    // Return a deep copy to prevent external modification
    result := make([]ldomain.Message, len(m.messages))
    for i, msg := range m.messages {
        result[i] = m.cloneMessage(msg)
    }

    return result
}
```

Key thread safety features:
1. **Read/write mutex**: Uses `sync.RWMutex` to allow concurrent reads
2. **Defensive copying**: Returns copies of data to prevent external modification
3. **Complete locking**: Ensures all state modifications are protected
4. **Minimized lock scope**: Locks only for the necessary operations

### Validator Thread Safety

The schema validator uses synchronized data structures:

```go
// RegexCache stores compiled regular expressions to avoid recompilation
var RegexCache = sync.Map{}

// Validator implements schema validation with performance enhancements
type Validator struct {
    // errorBufferPool provides reusable string buffers for errors
    errorBufferPool sync.Pool

    // validationResultPool provides reusable validation results
    validationResultPool sync.Pool
    
    // enableCoercion controls whether the validator attempts to coerce values
    enableCoercion bool
}
```

Here, `sync.Map` provides a thread-safe map implementation without explicit locking.

### Response Cache Thread Safety

The `ResponseCache` uses lock upgrading for optimistic reads and minimal contention:

```go
// Get retrieves a cached response for the given messages and options
func (c *ResponseCache) Get(messages []ldomain.Message, options []ldomain.Option) (ldomain.Response, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    // Generate cache key from messages and options
    key := c.generateCacheKey(messages, options)

    entry, found := c.cache[key]
    if !found {
        return ldomain.Response{}, false
    }

    // Check if entry has expired
    if time.Since(entry.Timestamp) > c.ttl {
        return ldomain.Response{}, false
    }

    // Update usage count (requires write lock)
    c.mu.RUnlock()
    c.mu.Lock()
    entry.UsageCount++
    c.cache[key] = entry
    c.mu.Unlock()
    c.mu.RLock()

    return entry.Response, true
}
```

This pattern:
1. Uses a read lock initially for concurrent access
2. Upgrades to a write lock only when necessary
3. Minimizes the duration of the write lock
4. Returns to a read lock for the remainder of the operation

## Singleton Initialization

Go-LLMs uses the `sync.Once` pattern to ensure thread-safe initialization of singletons:

```go
// Global singleton channel pool
var (
    globalChannelPool *ChannelPool
    channelPoolOnce   sync.Once
)

// GetChannelPool returns the singleton global channel pool
func GetChannelPool() *ChannelPool {
    channelPoolOnce.Do(func() {
        globalChannelPool = NewChannelPool()
    })
    return globalChannelPool
}
```

This ensures:
1. Lazy initialization (created only when needed)
2. Thread-safe creation (only initialized once)
3. No locking overhead after initialization

## Streaming with Channels

One of the most important concurrency patterns in Go-LLMs is channel-based streaming for token processing.

### Channel Pool Implementation

```go
// ChannelPool is a pool of channels that can be reused to reduce memory allocations
// This significantly reduces GC pressure in streaming operations
type ChannelPool struct {
    pool sync.Pool
}

// GetResponseStream creates a new response stream using the pool
// The returned channel is cast to ResponseStream (read-only)
// The caller is responsible for closing the channel when done with it
func (p *ChannelPool) GetResponseStream() (ResponseStream, chan Token) {
    ch := p.Get()
    return ch, ch
}
```

### Stream Usage Pattern

The typical pattern for token streaming:

```go
// Producer (typically in a goroutine)
func streamTokens(stream chan Token, text string) {
    defer close(stream)
    
    // Split text into tokens
    tokens := tokenize(text)
    
    for i, token := range tokens {
        select {
        case stream <- Token{
            Text:     token,
            Finished: i == len(tokens)-1,
        }:
            // Successfully sent
        case <-time.After(1 * time.Second):
            // Timeout - consumer might be slow or gone
            return
        }
    }
}

// Consumer
func processStream(stream <-chan Token) string {
    var result strings.Builder
    
    for token := range stream {
        result.WriteString(token.Text)
        if token.Finished {
            break
        }
    }
    
    return result.String()
}
```

### Streaming with Timeouts and Cancellation

For robust streaming implementations, Go-LLMs uses contexts and timeouts:

```go
func streamWithContext(ctx context.Context, stream chan Token, text string) {
    defer close(stream)
    
    tokens := tokenize(text)
    
    for i, token := range tokens {
        select {
        case stream <- Token{
            Text:     token,
            Finished: i == len(tokens)-1,
        }:
            // Successfully sent
        case <-ctx.Done():
            // Context cancelled or timed out
            return
        }
    }
}

// Usage with context
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

stream := make(chan Token, 20)
go streamWithContext(ctx, stream, "Long text to stream...")

// Process the stream...
```

## Fan-Out/Fan-In Pattern

The consensus mechanism in Go-LLMs uses fan-out/fan-in patterns to process multiple provider responses:

```go
// Example fan-out/fan-in pattern (conceptual)
func (p *ConsensusProvider) Complete(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.Response, error) {
    // Create response channel
    responseChannel := make(chan providerResponse, len(p.providers))
    
    // Fan out - send requests to all providers
    for _, provider := range p.providers {
        go func(prov domain.Provider) {
            resp, err := prov.Complete(ctx, messages, options...)
            responseChannel <- providerResponse{Response: resp, Error: err}
        }(provider)
    }
    
    // Fan in - collect responses
    var responses []domain.Response
    for i := 0; i < len(p.providers); i++ {
        select {
        case result := <-responseChannel:
            if result.Error == nil {
                responses = append(responses, result.Response)
            }
        case <-ctx.Done():
            return domain.Response{}, ctx.Err()
        }
    }
    
    // Process the collected responses
    consensus := p.findConsensus(responses)
    return consensus, nil
}
```

This pattern:
1. Spawns concurrent goroutines for each provider request
2. Collects results through a buffered channel
3. Handles timeouts and cancellation with context
4. Processes the aggregated results

## Worker Pools

For high-throughput scenarios, worker pools can be implemented:

```go
// WorkerPool manages a pool of workers for parallel processing
type WorkerPool struct {
    tasks   chan Task
    results chan Result
    workers int
    wg      sync.WaitGroup
}

// NewWorkerPool creates a new worker pool with the specified number of workers
func NewWorkerPool(workers int) *WorkerPool {
    pool := &WorkerPool{
        tasks:   make(chan Task, workers*2),
        results: make(chan Result, workers*2),
        workers: workers,
    }
    
    // Start workers
    pool.Start()
    
    return pool
}

// Start launches the worker goroutines
func (p *WorkerPool) Start() {
    for i := 0; i < p.workers; i++ {
        p.wg.Add(1)
        go p.worker()
    }
}

// worker processes tasks from the task channel
func (p *WorkerPool) worker() {
    defer p.wg.Done()
    
    for task := range p.tasks {
        result := task.Execute()
        p.results <- result
    }
}

// Submit adds a task to the pool
func (p *WorkerPool) Submit(task Task) {
    p.tasks <- task
}

// Results returns the results channel
func (p *WorkerPool) Results() <-chan Result {
    return p.results
}

// Stop closes the task channel and waits for all workers to finish
func (p *WorkerPool) Stop() {
    close(p.tasks)
    p.wg.Wait()
    close(p.results)
}
```

## Concurrency Best Practices

### 1. Always Lock Data Access

When multiple goroutines access shared data, always use synchronization:

```go
// Bad - race condition
func (c *Counter) Increment() {
    c.value++ // Shared variable without synchronization
}

// Good - protected access
func (c *Counter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value++
}
```

### 2. Use Read/Write Mutexes for Read-Heavy Workloads

When reads are more common than writes, use `sync.RWMutex`:

```go
// Configuration object with infrequent writes
type Config struct {
    settings map[string]string
    mu       sync.RWMutex
}

// Read operation - multiple readers allowed
func (c *Config) Get(key string) string {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.settings[key]
}

// Write operation - exclusive lock
func (c *Config) Set(key, value string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.settings[key] = value
}
```

### 3. Keep Lock Scopes as Small as Possible

Minimize the code executed while holding a lock:

```go
// Bad - holds lock during expensive operation
func (c *Cache) GetOrCompute(key string, compute func() interface{}) interface{} {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    if val, ok := c.data[key]; ok {
        return val
    }
    
    // Expensive operation while holding lock
    result := compute()
    c.data[key] = result
    return result
}

// Good - minimizes locked time
func (c *Cache) GetOrCompute(key string, compute func() interface{}) interface{} {
    // Check if in cache with read lock
    c.mu.RLock()
    if val, ok := c.data[key]; ok {
        c.mu.RUnlock()
        return val
    }
    c.mu.RUnlock()
    
    // Compute without lock
    result := compute()
    
    // Update cache with write lock
    c.mu.Lock()
    // Double-check in case another goroutine computed while we were computing
    if val, ok := c.data[key]; ok {
        c.mu.Unlock()
        return val
    }
    c.data[key] = result
    c.mu.Unlock()
    
    return result
}
```

### 4. Use Buffered Channels for Asynchronous Operations

Use buffered channels to prevent blocking in producer-consumer patterns:

```go
// Buffered channel prevents blocking if consumer is temporarily slow
tokenChannel := make(chan Token, 20)

// Producer won't block until buffer is full
go func() {
    defer close(tokenChannel)
    for _, token := range tokens {
        tokenChannel <- token
    }
}()

// Consumer processes at its own pace
for token := range tokenChannel {
    processToken(token)
}
```

### 5. Always Handle Channel Closure

Ensure proper channel handling to prevent panics:

```go
// Sending on a closed channel will panic
// Always protect against this
func SafeSend(ch chan<- Token, token Token) (closed bool) {
    defer func() {
        if r := recover(); r != nil {
            closed = true
        }
    }()
    
    select {
    case ch <- token:
        return false
    default:
        // Channel buffer is full
        return false
    }
}
```

### 6. Use Select with Default for Non-Blocking Operations

Use the `select` statement with a `default` case for non-blocking operations:

```go
// Non-blocking send
func TrySend(ch chan<- Token, token Token) bool {
    select {
    case ch <- token:
        return true
    default:
        // Channel is full or closed
        return false
    }
}

// Non-blocking receive
func TryReceive(ch <-chan Token) (Token, bool) {
    select {
    case token, ok := <-ch:
        return token, ok
    default:
        // Channel is empty
        return Token{}, false
    }
}
```

### 7. Use Context for Cancellation and Timeouts

Always use context for propagating cancellation and timeouts:

```go
func ProcessWithTimeout(input string) (string, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    return ProcessWithContext(ctx, input)
}

func ProcessWithContext(ctx context.Context, input string) (string, error) {
    // Create result channel
    resultCh := make(chan string, 1)
    errCh := make(chan error, 1)
    
    // Process in goroutine
    go func() {
        result, err := expensiveOperation(input)
        if err != nil {
            errCh <- err
            return
        }
        resultCh <- result
    }()
    
    // Wait with context
    select {
    case result := <-resultCh:
        return result, nil
    case err := <-errCh:
        return "", err
    case <-ctx.Done():
        return "", ctx.Err()
    }
}
```

## Avoiding Concurrency Issues

### 1. Deadlocks

Avoid deadlocks by:
- Maintaining consistent lock acquisition order
- Using lock timeouts
- Minimizing nested locks

Example of deadlock-prone code:
```go
func Transfer(from, to *Account, amount float64) {
    from.mu.Lock()
    defer from.mu.Unlock()
    
    to.mu.Lock()
    defer to.mu.Unlock()
    
    from.balance -= amount
    to.balance += amount
}
```

Deadlock-resistant version:
```go
func Transfer(from, to *Account, amount float64) {
    // Lock accounts in a consistent order based on ID
    if from.id < to.id {
        from.mu.Lock()
        to.mu.Lock()
    } else {
        to.mu.Lock()
        from.mu.Lock()
    }
    
    // Use defer in reverse order of acquisition
    defer func() {
        if from.id < to.id {
            to.mu.Unlock()
            from.mu.Unlock()
        } else {
            from.mu.Unlock()
            to.mu.Unlock()
        }
    }()
    
    from.balance -= amount
    to.balance += amount
}
```

### 2. Race Conditions

Use the Go race detector to identify race conditions:

```bash
go test -race ./...
```

Common race conditions:
1. Unprotected access to shared variables
2. Map access without synchronization
3. Slice access or modification from multiple goroutines

### 3. Goroutine Leaks

Prevent goroutine leaks by:
- Using context cancellation
- Setting timeouts
- Ensuring proper channel closing

Example of preventing goroutine leaks:
```go
func ProcessItems(ctx context.Context, items []Item) []Result {
    resultCh := make(chan Result)
    
    // Use a separate context for early cancellation
    processingCtx, cancel := context.WithCancel(ctx)
    
    // Track running goroutines
    var wg sync.WaitGroup
    
    // Launch workers
    for _, item := range items {
        wg.Add(1)
        go func(item Item) {
            defer wg.Done()
            
            select {
            case resultCh <- processItem(item):
                // Result sent
            case <-processingCtx.Done():
                // Processing cancelled
                return
            }
        }(item)
    }
    
    // Close channel when all goroutines are done
    go func() {
        wg.Wait()
        close(resultCh)
    }()
    
    // Collect results with timeout
    var results []Result
    timeout := time.After(30 * time.Second)
    
    for {
        select {
        case result, ok := <-resultCh:
            if !ok {
                // Channel closed, all results collected
                cancel() // Ensure all goroutines exit
                return results
            }
            results = append(results, result)
        case <-timeout:
            // Timeout, cancel processing and return partial results
            cancel()
            return results
        case <-ctx.Done():
            // Parent context cancelled
            cancel()
            return results
        }
    }
}
```

## Conclusion

Go-LLMs implements a variety of concurrency patterns to ensure thread safety, optimize performance, and provide a robust foundation for LLM interactions. By following established Go concurrency best practices and implementing appropriate synchronization mechanisms, the library handles high-throughput scenarios effectively while maintaining correctness.

Understanding these patterns is essential for extending or customizing Go-LLMs in ways that preserve its thread safety and performance characteristics.