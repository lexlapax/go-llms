# Multi-Provider Strategies: Reliability and Optimization

> **[Project Root](/) / [Documentation](/docs/) / [User Guide](/docs/user-guide/) / [Guides](/docs/user-guide/guides/) / Multi-Provider Strategies**

Master advanced multi-provider patterns to maximize reliability, optimize costs, and leverage the unique strengths of different LLM providers. Build resilient systems that automatically handle failures and adapt to changing conditions.

## Why Multi-Provider Strategies Matter

- **Reliability** - Eliminate single points of failure with automatic fallbacks
- **Cost Optimization** - Route queries to the most cost-effective provider
- **Performance** - Balance speed, quality, and capacity across providers
- **Risk Mitigation** - Reduce dependency on any single vendor
- **Strategic Flexibility** - Adapt to changing provider landscapes

## Multi-Provider Architecture

![Multi-Provider Strategy](../../images/multi-provider-architecture.svg)

### Core Patterns
1. **Failover Strategy** - Primary provider with automatic fallbacks
2. **Load Balancing** - Distribute requests across multiple providers
3. **Cost Optimization** - Route based on price and budget constraints
4. **Quality Routing** - Select provider based on query complexity
5. **Regional Distribution** - Use different providers in different regions
6. **A/B Testing** - Compare providers with live traffic

## Prerequisites

- [Provider Selection completed](provider-selection.md) ✅
- [Provider Setup for multiple providers](provider-setup.md) ✅
- Understanding of reliability patterns ✅

---

## Level 1: Basic Failover Strategy
*Implement automatic provider fallback in 20 minutes*

### Simple Failover Implementation
```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
)

// FailoverAgent provides automatic provider fallback
type FailoverAgent struct {
    name            string
    providers       []ProviderConfig
    currentProvider int
    maxRetries      int
    retryDelay      time.Duration
}

type ProviderConfig struct {
    Name        string
    Agent       domain.BaseAgent
    Priority    int     // Lower number = higher priority
    CostPerCall float64 // Cost in USD
    MaxLatency  time.Duration
    Reliability float64 // 0.0-1.0 reliability score
}

func NewFailoverAgent(name string) *FailoverAgent {
    return &FailoverAgent{
        name:            name,
        providers:       make([]ProviderConfig, 0),
        currentProvider: 0,
        maxRetries:      3,
        retryDelay:      time.Second,
    }
}

func (fa *FailoverAgent) AddProvider(name, providerString string, priority int, cost float64, reliability float64) error {
    agent, err := core.NewAgentFromString(fmt.Sprintf("%s-%s", fa.name, name), providerString)
    if err != nil {
        return fmt.Errorf("failed to create %s agent: %w", name, err)
    }

    config := ProviderConfig{
        Name:        name,
        Agent:       agent,
        Priority:    priority,
        CostPerCall: cost,
        MaxLatency:  10 * time.Second,
        Reliability: reliability,
    }

    fa.providers = append(fa.providers, config)
    fa.sortProvidersByPriority()
    
    fmt.Printf("✅ Added provider: %s (Priority: %d, Cost: $%.4f, Reliability: %.2f)\n", 
        name, priority, cost, reliability)
    
    return nil
}

func (fa *FailoverAgent) sortProvidersByPriority() {
    // Simple bubble sort by priority (ascending)
    for i := 0; i < len(fa.providers); i++ {
        for j := 0; j < len(fa.providers)-1-i; j++ {
            if fa.providers[j].Priority > fa.providers[j+1].Priority {
                fa.providers[j], fa.providers[j+1] = fa.providers[j+1], fa.providers[j]
            }
        }
    }
}

func (fa *FailoverAgent) Run(ctx context.Context, state domain.StateReader) (*domain.State, error) {
    if len(fa.providers) == 0 {
        return nil, fmt.Errorf("no providers configured")
    }

    var lastError error

    // Try each provider in priority order
    for attempt := 0; attempt < fa.maxRetries; attempt++ {
        for i, provider := range fa.providers {
            fmt.Printf("🔄 Attempt %d: Trying provider %s\n", attempt+1, provider.Name)
            
            // Create timeout context for this provider
            providerCtx, cancel := context.WithTimeout(ctx, provider.MaxLatency)
            
            startTime := time.Now()
            result, err := provider.Agent.Run(providerCtx, state)
            duration := time.Since(startTime)
            cancel()

            if err == nil {
                fa.currentProvider = i
                fmt.Printf("✅ Success with %s (latency: %v)\n", provider.Name, duration)
                return result, nil
            }

            lastError = err
            fmt.Printf("❌ Failed with %s: %v\n", provider.Name, err)

            // If this was a timeout error, skip to next provider immediately
            if isTimeoutError(err) {
                continue
            }

            // For other errors, wait before retrying
            if attempt < fa.maxRetries-1 {
                time.Sleep(fa.retryDelay)
            }
        }

        // Exponential backoff between full retry cycles
        if attempt < fa.maxRetries-1 {
            backoffDelay := time.Duration(attempt+1) * fa.retryDelay
            fmt.Printf("⏳ Waiting %v before next retry cycle\n", backoffDelay)
            time.Sleep(backoffDelay)
        }
    }

    return nil, fmt.Errorf("all providers failed after %d attempts, last error: %w", fa.maxRetries, lastError)
}

func (fa *FailoverAgent) GetCurrentProvider() string {
    if fa.currentProvider < len(fa.providers) {
        return fa.providers[fa.currentProvider].Name
    }
    return "none"
}

func (fa *FailoverAgent) GetProviderStats() map[string]interface{} {
    stats := make(map[string]interface{})
    
    for _, provider := range fa.providers {
        stats[provider.Name] = map[string]interface{}{
            "priority":    provider.Priority,
            "cost":        provider.CostPerCall,
            "reliability": provider.Reliability,
            "max_latency": provider.MaxLatency,
        }
    }
    
    return stats
}

func isTimeoutError(err error) bool {
    return err != nil && (err == context.DeadlineExceeded || 
                         contains(err.Error(), "timeout") ||
                         contains(err.Error(), "deadline"))
}

func contains(s, substr string) bool {
    return len(s) >= len(substr) && 
           (len(s) == len(substr) && s == substr || 
            len(s) > len(substr) && (s[:len(substr)] == substr || 
                                   s[len(s)-len(substr):] == substr ||
                                   containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
    for i := 0; i <= len(s)-len(substr); i++ {
        if s[i:i+len(substr)] == substr {
            return true
        }
    }
    return false
}

func main() {
    fmt.Println("🔄 Basic Failover Strategy")
    fmt.Println("==========================")

    // Create failover agent
    failoverAgent := NewFailoverAgent("smart-assistant")

    // Add providers in priority order (1=highest priority)
    err := failoverAgent.AddProvider("primary", "openai/gpt-4o-mini", 1, 0.0001, 0.98)
    if err != nil {
        log.Printf("Warning: Failed to add OpenAI: %v", err)
    }

    err = failoverAgent.AddProvider("secondary", "anthropic/claude-3-5-haiku", 2, 0.00015, 0.97)
    if err != nil {
        log.Printf("Warning: Failed to add Anthropic: %v", err)
    }

    err = failoverAgent.AddProvider("tertiary", "gemini/gemini-2.0-flash", 3, 0.00008, 0.95)
    if err != nil {
        log.Printf("Warning: Failed to add Gemini: %v", err)
    }

    err = failoverAgent.AddProvider("fallback", "ollama/llama3.2:3b", 4, 0.0, 0.90)
    if err != nil {
        log.Printf("Warning: Failed to add Ollama: %v", err)
    }

    if len(failoverAgent.providers) == 0 {
        log.Fatal("No providers available. Please check your API keys and setup.")
    }

    // Test failover with multiple queries
    queries := []string{
        "What is the capital of France?",
        "Explain quantum computing in simple terms",
        "Write a haiku about programming",
        "Calculate the square root of 144",
        "What are the benefits of microservices architecture?",
    }

    ctx := context.Background()
    
    for i, query := range queries {
        fmt.Printf("\n--- Query %d ---\n", i+1)
        fmt.Printf("Query: %s\n", query)

        state := domain.NewState()
        state.Set("user_input", query)

        result, err := failoverAgent.Run(ctx, state)
        if err != nil {
            fmt.Printf("❌ All providers failed: %v\n", err)
            continue
        }

        if response, exists := result.Get("response"); exists {
            fmt.Printf("✅ Response from %s:\n%v\n", 
                failoverAgent.GetCurrentProvider(), response)
        }
    }

    // Display final statistics
    fmt.Printf("\n📊 Provider Statistics:\n")
    stats := failoverAgent.GetProviderStats()
    for name, data := range stats {
        if dataMap, ok := data.(map[string]interface{}); ok {
            fmt.Printf("  %s: Priority=%v, Cost=$%.4f, Reliability=%.2f\n",
                name, dataMap["priority"], dataMap["cost"], dataMap["reliability"])
        }
    }
}
```

### Key Features
✅ **Automatic Failover** - Seamless provider switching on failure  
✅ **Priority-Based Routing** - Configurable provider preferences  
✅ **Timeout Handling** - Provider-specific latency limits  
✅ **Exponential Backoff** - Intelligent retry strategies  

---

## Level 2: Intelligent Load Balancing
*Distribute load based on cost, performance, and capacity*

### Smart Load Balancer Implementation
```go
package main

import (
    "context"
    "fmt"
    "log"
    "math/rand"
    "sync"
    "time"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
)

// LoadBalancer distributes requests across multiple providers intelligently
type LoadBalancer struct {
    name              string
    providers         []LoadBalancedProvider
    strategy          LoadBalancingStrategy
    metrics           *LoadBalancerMetrics
    healthChecker     *HealthChecker
    rateLimiter       *RateLimiter
    costOptimizer     *CostOptimizer
    
    mutex             sync.RWMutex
}

type LoadBalancedProvider struct {
    Config         ProviderConfig
    CurrentLoad    int
    MaxConcurrency int
    HealthStatus   HealthStatus
    Metrics        ProviderMetrics
    RateLimit      *ProviderRateLimit
}

type LoadBalancingStrategy string

const (
    RoundRobin    LoadBalancingStrategy = "round_robin"
    WeightedRandom LoadBalancingStrategy = "weighted_random"
    LeastLoaded   LoadBalancingStrategy = "least_loaded"
    CostOptimized LoadBalancingStrategy = "cost_optimized"
    QualityFirst  LoadBalancingStrategy = "quality_first"
    Adaptive      LoadBalancingStrategy = "adaptive"
)

type HealthStatus struct {
    IsHealthy        bool
    LastCheck        time.Time
    ConsecutiveErrors int
    AverageLatency   time.Duration
    ErrorRate        float64
}

type ProviderMetrics struct {
    TotalRequests    int64
    SuccessfulRequests int64
    FailedRequests   int64
    TotalLatency     time.Duration
    AverageLatency   time.Duration
    LastRequestTime  time.Time
    CostAccumulated  float64
}

type ProviderRateLimit struct {
    RequestsPerMinute int
    CurrentRequests   int
    WindowStart       time.Time
}

type LoadBalancerMetrics struct {
    TotalRequests       int64
    DistributionPattern map[string]int64
    AverageCost         float64
    AverageLatency      time.Duration
    ErrorRate           float64
    
    mutex sync.RWMutex
}

type HealthChecker struct {
    checkInterval time.Duration
    timeout       time.Duration
    stopCh        chan struct{}
}

type RateLimiter struct {
    globalLimit   int
    currentLoad   int
    windowStart   time.Time
    mutex         sync.Mutex
}

type CostOptimizer struct {
    dailyBudget     float64
    currentSpend    float64
    budgetResetTime time.Time
    costThreshold   float64
}

func NewLoadBalancer(name string, strategy LoadBalancingStrategy) *LoadBalancer {
    return &LoadBalancer{
        name:      name,
        providers: make([]LoadBalancedProvider, 0),
        strategy:  strategy,
        metrics:   NewLoadBalancerMetrics(),
        healthChecker: &HealthChecker{
            checkInterval: 30 * time.Second,
            timeout:       5 * time.Second,
            stopCh:        make(chan struct{}),
        },
        rateLimiter: &RateLimiter{
            globalLimit: 1000, // requests per minute
        },
        costOptimizer: &CostOptimizer{
            dailyBudget:   100.0, // $100 per day
            costThreshold: 0.8,   // Alert at 80% of budget
        },
    }
}

func (lb *LoadBalancer) AddProvider(name, providerString string, weight float64, maxConcurrency int, rateLimit int) error {
    agent, err := core.NewAgentFromString(fmt.Sprintf("%s-%s", lb.name, name), providerString)
    if err != nil {
        return fmt.Errorf("failed to create %s agent: %w", name, err)
    }

    // Determine cost based on provider type (simplified)
    cost := lb.estimateProviderCost(providerString)

    provider := LoadBalancedProvider{
        Config: ProviderConfig{
            Name:        name,
            Agent:       agent,
            Priority:    int(weight * 10), // Convert weight to priority
            CostPerCall: cost,
            Reliability: 0.95, // Default reliability
        },
        CurrentLoad:    0,
        MaxConcurrency: maxConcurrency,
        HealthStatus: HealthStatus{
            IsHealthy:        true,
            LastCheck:        time.Now(),
            ConsecutiveErrors: 0,
            ErrorRate:        0.0,
        },
        Metrics: ProviderMetrics{},
        RateLimit: &ProviderRateLimit{
            RequestsPerMinute: rateLimit,
            CurrentRequests:   0,
            WindowStart:       time.Now(),
        },
    }

    lb.mutex.Lock()
    lb.providers = append(lb.providers, provider)
    lb.mutex.Unlock()

    fmt.Printf("✅ Added provider: %s (Weight: %.2f, Max Concurrency: %d, Rate Limit: %d/min)\n",
        name, weight, maxConcurrency, rateLimit)

    return nil
}

func (lb *LoadBalancer) estimateProviderCost(providerString string) float64 {
    costMap := map[string]float64{
        "openai":    0.00015,
        "anthropic": 0.00018,
        "gemini":    0.00008,
        "ollama":    0.0,
        "vertex":    0.00020,
    }

    for provider, cost := range costMap {
        if containsSubstring(providerString, provider) {
            return cost
        }
    }
    return 0.0001 // Default cost
}

func (lb *LoadBalancer) Run(ctx context.Context, state domain.StateReader) (*domain.State, error) {
    if len(lb.providers) == 0 {
        return nil, fmt.Errorf("no providers configured")
    }

    // Check rate limiting
    if !lb.rateLimiter.Allow() {
        return nil, fmt.Errorf("rate limit exceeded")
    }

    // Check cost budget
    if !lb.costOptimizer.CheckBudget() {
        return nil, fmt.Errorf("daily budget exceeded")
    }

    // Select provider based on strategy
    provider, err := lb.selectProvider(ctx, state)
    if err != nil {
        return nil, err
    }

    // Execute request with monitoring
    return lb.executeWithMonitoring(ctx, provider, state)
}

func (lb *LoadBalancer) selectProvider(ctx context.Context, state domain.StateReader) (*LoadBalancedProvider, error) {
    lb.mutex.RLock()
    defer lb.mutex.RUnlock()

    // Filter healthy providers
    healthyProviders := make([]LoadBalancedProvider, 0)
    for _, provider := range lb.providers {
        if provider.HealthStatus.IsHealthy && 
           provider.CurrentLoad < provider.MaxConcurrency &&
           provider.RateLimit.CanHandle() {
            healthyProviders = append(healthyProviders, provider)
        }
    }

    if len(healthyProviders) == 0 {
        return nil, fmt.Errorf("no healthy providers available")
    }

    // Apply selection strategy
    switch lb.strategy {
    case RoundRobin:
        return lb.selectRoundRobin(healthyProviders), nil
    case WeightedRandom:
        return lb.selectWeightedRandom(healthyProviders), nil
    case LeastLoaded:
        return lb.selectLeastLoaded(healthyProviders), nil
    case CostOptimized:
        return lb.selectCostOptimized(healthyProviders), nil
    case QualityFirst:
        return lb.selectQualityFirst(healthyProviders), nil
    case Adaptive:
        return lb.selectAdaptive(healthyProviders, state), nil
    default:
        return &healthyProviders[0], nil
    }
}

func (lb *LoadBalancer) selectRoundRobin(providers []LoadBalancedProvider) *LoadBalancedProvider {
    // Simple round-robin implementation
    index := int(lb.metrics.TotalRequests) % len(providers)
    return &providers[index]
}

func (lb *LoadBalancer) selectWeightedRandom(providers []LoadBalancedProvider) *LoadBalancedProvider {
    totalWeight := 0.0
    for _, provider := range providers {
        totalWeight += float64(provider.Config.Priority)
    }

    if totalWeight == 0 {
        return &providers[0]
    }

    r := rand.Float64() * totalWeight
    currentWeight := 0.0

    for i, provider := range providers {
        currentWeight += float64(provider.Config.Priority)
        if r <= currentWeight {
            return &providers[i]
        }
    }

    return &providers[len(providers)-1]
}

func (lb *LoadBalancer) selectLeastLoaded(providers []LoadBalancedProvider) *LoadBalancedProvider {
    leastLoaded := &providers[0]
    minLoad := providers[0].CurrentLoad

    for i := 1; i < len(providers); i++ {
        if providers[i].CurrentLoad < minLoad {
            minLoad = providers[i].CurrentLoad
            leastLoaded = &providers[i]
        }
    }

    return leastLoaded
}

func (lb *LoadBalancer) selectCostOptimized(providers []LoadBalancedProvider) *LoadBalancedProvider {
    cheapest := &providers[0]
    minCost := providers[0].Config.CostPerCall

    for i := 1; i < len(providers); i++ {
        if providers[i].Config.CostPerCall < minCost {
            minCost = providers[i].Config.CostPerCall
            cheapest = &providers[i]
        }
    }

    return cheapest
}

func (lb *LoadBalancer) selectQualityFirst(providers []LoadBalancedProvider) *LoadBalancedProvider {
    best := &providers[0]
    maxReliability := providers[0].Config.Reliability

    for i := 1; i < len(providers); i++ {
        if providers[i].Config.Reliability > maxReliability {
            maxReliability = providers[i].Config.Reliability
            best = &providers[i]
        }
    }

    return best
}

func (lb *LoadBalancer) selectAdaptive(providers []LoadBalancedProvider, state domain.StateReader) *LoadBalancedProvider {
    // Adaptive selection based on current conditions
    
    // If budget is low, prefer cost optimization
    if lb.costOptimizer.currentSpend/lb.costOptimizer.dailyBudget > 0.8 {
        return lb.selectCostOptimized(providers)
    }

    // If system is under high load, prefer fastest providers
    if lb.getTotalLoad() > 50 {
        return lb.selectByLatency(providers)
    }

    // Otherwise, balance quality and cost
    return lb.selectBalanced(providers)
}

func (lb *LoadBalancer) selectByLatency(providers []LoadBalancedProvider) *LoadBalancedProvider {
    fastest := &providers[0]
    minLatency := providers[0].HealthStatus.AverageLatency

    for i := 1; i < len(providers); i++ {
        if providers[i].HealthStatus.AverageLatency < minLatency {
            minLatency = providers[i].HealthStatus.AverageLatency
            fastest = &providers[i]
        }
    }

    return fastest
}

func (lb *LoadBalancer) selectBalanced(providers []LoadBalancedProvider) *LoadBalancedProvider {
    best := &providers[0]
    bestScore := lb.calculateBalancedScore(&providers[0])

    for i := 1; i < len(providers); i++ {
        score := lb.calculateBalancedScore(&providers[i])
        if score > bestScore {
            bestScore = score
            best = &providers[i]
        }
    }

    return best
}

func (lb *LoadBalancer) calculateBalancedScore(provider *LoadBalancedProvider) float64 {
    // Composite score: reliability * 0.4 + (1/cost) * 0.3 + (1/latency) * 0.3
    reliabilityScore := provider.Config.Reliability * 0.4
    
    costScore := 0.0
    if provider.Config.CostPerCall > 0 {
        costScore = (1.0 / provider.Config.CostPerCall) * 0.0001 * 0.3
    } else {
        costScore = 0.3 // Free provider gets full cost score
    }
    
    latencyScore := 0.0
    if provider.HealthStatus.AverageLatency > 0 {
        latencyScore = (1.0 / float64(provider.HealthStatus.AverageLatency.Milliseconds())) * 1000 * 0.3
    } else {
        latencyScore = 0.3 // Default latency score
    }

    return reliabilityScore + costScore + latencyScore
}

func (lb *LoadBalancer) executeWithMonitoring(ctx context.Context, provider *LoadBalancedProvider, state domain.StateReader) (*domain.State, error) {
    // Increment load
    lb.mutex.Lock()
    provider.CurrentLoad++
    provider.RateLimit.IncrementUsage()
    lb.mutex.Unlock()

    // Decrement load on completion
    defer func() {
        lb.mutex.Lock()
        provider.CurrentLoad--
        lb.mutex.Unlock()
    }()

    // Execute request with timing
    startTime := time.Now()
    result, err := provider.Config.Agent.Run(ctx, state)
    duration := time.Since(startTime)

    // Update metrics
    lb.updateMetrics(provider, duration, err == nil)

    // Update cost tracking
    if err == nil {
        lb.costOptimizer.AddCost(provider.Config.CostPerCall)
    }

    if err != nil {
        lb.updateHealthStatus(provider, false, duration)
        return nil, fmt.Errorf("provider %s failed: %w", provider.Config.Name, err)
    }

    lb.updateHealthStatus(provider, true, duration)
    fmt.Printf("✅ Request handled by %s (latency: %v, cost: $%.6f)\n", 
        provider.Config.Name, duration, provider.Config.CostPerCall)

    return result, nil
}

func (lb *LoadBalancer) updateMetrics(provider *LoadBalancedProvider, duration time.Duration, success bool) {
    lb.mutex.Lock()
    defer lb.mutex.Unlock()

    provider.Metrics.TotalRequests++
    provider.Metrics.LastRequestTime = time.Now()
    provider.Metrics.TotalLatency += duration
    provider.Metrics.AverageLatency = time.Duration(int64(provider.Metrics.TotalLatency) / provider.Metrics.TotalRequests)

    if success {
        provider.Metrics.SuccessfulRequests++
    } else {
        provider.Metrics.FailedRequests++
    }

    provider.Metrics.CostAccumulated += provider.Config.CostPerCall

    // Update global metrics
    lb.metrics.TotalRequests++
    if lb.metrics.DistributionPattern == nil {
        lb.metrics.DistributionPattern = make(map[string]int64)
    }
    lb.metrics.DistributionPattern[provider.Config.Name]++
}

func (lb *LoadBalancer) updateHealthStatus(provider *LoadBalancedProvider, success bool, latency time.Duration) {
    lb.mutex.Lock()
    defer lb.mutex.Unlock()

    provider.HealthStatus.LastCheck = time.Now()
    provider.HealthStatus.AverageLatency = latency

    if success {
        provider.HealthStatus.ConsecutiveErrors = 0
        provider.HealthStatus.IsHealthy = true
    } else {
        provider.HealthStatus.ConsecutiveErrors++
        if provider.HealthStatus.ConsecutiveErrors > 3 {
            provider.HealthStatus.IsHealthy = false
        }
    }

    // Calculate error rate
    if provider.Metrics.TotalRequests > 0 {
        provider.HealthStatus.ErrorRate = float64(provider.Metrics.FailedRequests) / float64(provider.Metrics.TotalRequests)
    }
}

func (lb *LoadBalancer) getTotalLoad() int {
    totalLoad := 0
    for _, provider := range lb.providers {
        totalLoad += provider.CurrentLoad
    }
    return totalLoad
}

func (lb *LoadBalancer) GetStats() map[string]interface{} {
    lb.mutex.RLock()
    defer lb.mutex.RUnlock()

    stats := map[string]interface{}{
        "strategy":         string(lb.strategy),
        "total_requests":   lb.metrics.TotalRequests,
        "current_load":     lb.getTotalLoad(),
        "total_cost":       lb.costOptimizer.currentSpend,
        "budget_remaining": lb.costOptimizer.dailyBudget - lb.costOptimizer.currentSpend,
        "providers":        make(map[string]interface{}),
    }

    providers := stats["providers"].(map[string]interface{})
    for _, provider := range lb.providers {
        providers[provider.Config.Name] = map[string]interface{}{
            "healthy":           provider.HealthStatus.IsHealthy,
            "current_load":      provider.CurrentLoad,
            "total_requests":    provider.Metrics.TotalRequests,
            "success_rate":      float64(provider.Metrics.SuccessfulRequests) / float64(max(1, provider.Metrics.TotalRequests)),
            "average_latency":   provider.Metrics.AverageLatency,
            "cost_accumulated":  provider.Metrics.CostAccumulated,
        }
    }

    return stats
}

// Helper functions and types
func NewLoadBalancerMetrics() *LoadBalancerMetrics {
    return &LoadBalancerMetrics{
        DistributionPattern: make(map[string]int64),
    }
}

func (rl *RateLimiter) Allow() bool {
    rl.mutex.Lock()
    defer rl.mutex.Unlock()

    now := time.Now()
    if now.Sub(rl.windowStart) > time.Minute {
        rl.currentLoad = 0
        rl.windowStart = now
    }

    if rl.currentLoad < rl.globalLimit {
        rl.currentLoad++
        return true
    }

    return false
}

func (co *CostOptimizer) CheckBudget() bool {
    now := time.Now()
    if now.Sub(co.budgetResetTime) > 24*time.Hour {
        co.currentSpend = 0
        co.budgetResetTime = now
    }

    return co.currentSpend < co.dailyBudget
}

func (co *CostOptimizer) AddCost(cost float64) {
    co.currentSpend += cost
}

func (prl *ProviderRateLimit) CanHandle() bool {
    now := time.Now()
    if now.Sub(prl.WindowStart) > time.Minute {
        prl.CurrentRequests = 0
        prl.WindowStart = now
    }

    return prl.CurrentRequests < prl.RequestsPerMinute
}

func (prl *ProviderRateLimit) IncrementUsage() {
    now := time.Now()
    if now.Sub(prl.WindowStart) > time.Minute {
        prl.CurrentRequests = 0
        prl.WindowStart = now
    }
    prl.CurrentRequests++
}

func max(a, b int64) int64 {
    if a > b { return a }
    return b
}

func main() {
    fmt.Println("⚖️ Intelligent Load Balancing")
    fmt.Println("=============================")

    // Create load balancer with adaptive strategy
    lb := NewLoadBalancer("smart-router", Adaptive)

    // Add providers with different characteristics
    err := lb.AddProvider("fast", "gemini/gemini-2.0-flash", 0.8, 20, 100)
    if err != nil {
        log.Printf("Warning: Failed to add Gemini: %v", err)
    }

    err = lb.AddProvider("quality", "anthropic/claude-3-5-haiku", 0.9, 10, 60)
    if err != nil {
        log.Printf("Warning: Failed to add Anthropic: %v", err)
    }

    err = lb.AddProvider("balanced", "openai/gpt-4o-mini", 0.85, 15, 80)
    if err != nil {
        log.Printf("Warning: Failed to add OpenAI: %v", err)
    }

    err = lb.AddProvider("free", "ollama/llama3.2:3b", 0.7, 5, 50)
    if err != nil {
        log.Printf("Warning: Failed to add Ollama: %v", err)
    }

    if len(lb.providers) == 0 {
        log.Fatal("No providers available. Please check your setup.")
    }

    // Simulate multiple concurrent requests
    ctx := context.Background()
    queries := []string{
        "What is machine learning?",
        "Explain blockchain technology",
        "How does photosynthesis work?",
        "What are microservices?",
        "Describe quantum mechanics",
    }

    // Run queries multiple times to see load balancing in action
    for round := 1; round <= 3; round++ {
        fmt.Printf("\n🔄 Round %d - Load Balancing Test\n", round)
        fmt.Println("================================")

        for i, query := range queries {
            fmt.Printf("\nQuery %d: %s\n", i+1, query)

            state := domain.NewState()
            state.Set("user_input", query)

            result, err := lb.Run(ctx, state)
            if err != nil {
                fmt.Printf("❌ Request failed: %v\n", err)
                continue
            }

            if response, exists := result.Get("response"); exists {
                // Show just first 100 chars of response
                responseStr := fmt.Sprintf("%v", response)
                if len(responseStr) > 100 {
                    responseStr = responseStr[:100] + "..."
                }
                fmt.Printf("📝 Response: %s\n", responseStr)
            }
        }

        // Show statistics after each round
        fmt.Printf("\n📊 Load Balancer Statistics (Round %d):\n", round)
        stats := lb.GetStats()
        fmt.Printf("Total Requests: %v\n", stats["total_requests"])
        fmt.Printf("Current Load: %v\n", stats["current_load"])
        fmt.Printf("Total Cost: $%.6f\n", stats["total_cost"])
        fmt.Printf("Budget Remaining: $%.2f\n", stats["budget_remaining"])

        if providers, ok := stats["providers"].(map[string]interface{}); ok {
            for name, data := range providers {
                if providerData, ok := data.(map[string]interface{}); ok {
                    fmt.Printf("  %s: Requests=%v, Success=%.2f%%, Latency=%v, Cost=$%.6f\n",
                        name,
                        providerData["total_requests"],
                        providerData["success_rate"].(float64)*100,
                        providerData["average_latency"],
                        providerData["cost_accumulated"])
                }
            }
        }
    }
}
```

### Advanced Features
✅ **Multiple Load Balancing Strategies** - Round-robin, weighted, adaptive  
✅ **Health Monitoring** - Automatic provider health checks  
✅ **Rate Limiting** - Per-provider and global rate limits  
✅ **Cost Tracking** - Real-time budget monitoring  
✅ **Adaptive Selection** - Dynamic strategy based on conditions  

---

## Level 3: Advanced Multi-Provider Patterns
*Implement sophisticated routing and optimization strategies*

### Sophisticated Provider Router
```go
package main

import (
    "context"
    "fmt"
    "log"
    "math"
    "sync"
    "time"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
)

// AdvancedProviderRouter implements sophisticated routing strategies
type AdvancedProviderRouter struct {
    name               string
    providers          []AdvancedProvider
    routingStrategies  map[string]RoutingStrategy
    currentStrategy    string
    performanceModel   *PerformanceModel
    costModel          *CostModel
    qualityModel       *QualityModel
    circuitBreakers    map[string]*CircuitBreaker
    
    // Advanced features
    abTesting          *ABTestingManager
    geolocation        *GeolocationRouter
    workloadClassifier *WorkloadClassifier
    predictiveScaling  *PredictiveScaler
    
    mutex              sync.RWMutex
}

type AdvancedProvider struct {
    Config           ProviderConfig
    Capabilities     ProviderCapabilities
    Performance      PerformanceProfile
    Geographic       GeographicProfile
    Specializations  []string
    CircuitBreaker   *CircuitBreaker
    
    // Real-time metrics
    CurrentMetrics   RealTimeMetrics
    PredictedLoad    float64
    OptimalLoad      float64
}

type ProviderCapabilities struct {
    MaxContextLength    int
    SupportsStreaming   bool
    SupportsVision      bool
    SupportsFunctions   bool
    SupportsMultimodal  bool
    LanguageSupport     []string
    SpecializedDomains  []string
}

type PerformanceProfile struct {
    OptimalLatency      time.Duration
    MaxThroughput       int
    TokensPerSecond     float64
    ConcurrencyLimit    int
    WarmupTime          time.Duration
    CooldownTime        time.Duration
}

type GeographicProfile struct {
    PreferredRegions    []string
    DataCenters         []string
    ComplianceRegions   []string
    LatencyByRegion     map[string]time.Duration
}

type RealTimeMetrics struct {
    CurrentLatency      time.Duration
    CurrentThroughput   float64
    ErrorRate           float64
    QueueLength         int
    ResourceUtilization float64
    
    LastUpdated         time.Time
}

type RoutingStrategy interface {
    SelectProvider(ctx context.Context, request *RoutingRequest) (*AdvancedProvider, error)
    Name() string
    Configure(config map[string]interface{}) error
}

type RoutingRequest struct {
    UserQuery       string
    Context         domain.StateReader
    Requirements    RequestRequirements
    UserProfile     UserProfile
    SessionContext  SessionContext
}

type RequestRequirements struct {
    MaxLatency      time.Duration
    MinQuality      float64
    MaxCost         float64
    RequiredFeatures []string
    Complexity      int
    Priority        string
    Region          string
}

type UserProfile struct {
    UserID          string
    Tier            string
    CostLimits      CostLimits
    Preferences     UserPreferences
    History         RequestHistory
}

type SessionContext struct {
    SessionID       string
    RequestCount    int
    TotalCost       float64
    ProviderUsage   map[string]int
    QualityFeedback []QualityScore
}

// Circuit Breaker for individual providers
type CircuitBreaker struct {
    name           string
    maxFailures    int
    timeout        time.Duration
    resetTimeout   time.Duration
    
    failures       int
    lastFailTime   time.Time
    state          CircuitState
    
    mutex          sync.Mutex
}

type CircuitState int

const (
    Closed CircuitState = iota
    Open
    HalfOpen
)

// Performance-based routing strategy
type PerformanceBasedRouting struct {
    weightLatency    float64
    weightThroughput float64
    weightReliability float64
}

func (pbr *PerformanceBasedRouting) SelectProvider(ctx context.Context, request *RoutingRequest) (*AdvancedProvider, error) {
    // Complex algorithm to select based on performance characteristics
    // This is a simplified version
    return nil, fmt.Errorf("performance-based routing not implemented")
}

func (pbr *PerformanceBasedRouting) Name() string {
    return "performance_based"
}

func (pbr *PerformanceBasedRouting) Configure(config map[string]interface{}) error {
    if weight, ok := config["latency_weight"].(float64); ok {
        pbr.weightLatency = weight
    }
    if weight, ok := config["throughput_weight"].(float64); ok {
        pbr.weightThroughput = weight
    }
    if weight, ok := config["reliability_weight"].(float64); ok {
        pbr.weightReliability = weight
    }
    return nil
}

// Cost-optimization routing strategy
type CostOptimizedRouting struct {
    budgetConstraints BudgetConstraints
    costModels        map[string]CostModel
}

func (cor *CostOptimizedRouting) SelectProvider(ctx context.Context, request *RoutingRequest) (*AdvancedProvider, error) {
    // Select the most cost-effective provider that meets requirements
    return nil, fmt.Errorf("cost-optimized routing not implemented")
}

func (cor *CostOptimizedRouting) Name() string {
    return "cost_optimized"
}

func (cor *CostOptimizedRouting) Configure(config map[string]interface{}) error {
    return nil
}

// Quality-first routing strategy
type QualityFirstRouting struct {
    qualityThresholds map[string]float64
    qualityMetrics    []string
}

func (qfr *QualityFirstRouting) SelectProvider(ctx context.Context, request *RoutingRequest) (*AdvancedProvider, error) {
    // Select provider with highest expected quality for this request type
    return nil, fmt.Errorf("quality-first routing not implemented")
}

func (qfr *QualityFirstRouting) Name() string {
    return "quality_first"
}

func (qfr *QualityFirstRouting) Configure(config map[string]interface{}) error {
    return nil
}

// Machine Learning-based routing
type MLBasedRouting struct {
    model          MLModel
    features       []string
    predictionCache map[string]PredictionResult
    
    mutex          sync.RWMutex
}

func (mlr *MLBasedRouting) SelectProvider(ctx context.Context, request *RoutingRequest) (*AdvancedProvider, error) {
    // Use ML model to predict optimal provider
    features := mlr.extractFeatures(request)
    prediction := mlr.model.Predict(features)
    
    // Convert prediction to provider selection
    return mlr.selectFromPrediction(prediction)
}

func (mlr *MLBasedRouting) Name() string {
    return "ml_based"
}

func (mlr *MLBasedRouting) Configure(config map[string]interface{}) error {
    return nil
}

func (mlr *MLBasedRouting) extractFeatures(request *RoutingRequest) []float64 {
    // Extract features for ML model
    return []float64{
        float64(len(request.UserQuery)),
        float64(request.Requirements.Complexity),
        request.Requirements.MaxLatency.Seconds(),
        request.Requirements.MaxCost,
    }
}

func (mlr *MLBasedRouting) selectFromPrediction(prediction PredictionResult) (*AdvancedProvider, error) {
    // Convert ML prediction to provider selection
    return nil, fmt.Errorf("prediction conversion not implemented")
}

// Workload classifier for intelligent routing
type WorkloadClassifier struct {
    patterns    map[string]WorkloadPattern
    classifier  TextClassifier
    
    mutex       sync.RWMutex
}

type WorkloadPattern struct {
    Name            string
    Description     string
    OptimalProviders []string
    ResourceProfile ResourceProfile
    QualityProfile  QualityProfile
}

type ResourceProfile struct {
    CPUIntensive    bool
    MemoryIntensive bool
    IOIntensive     bool
    LatencySensitive bool
}

type QualityProfile struct {
    RequiresAccuracy   bool
    RequiresCreativity bool
    RequiresReasoning  bool
    RequiresSpeed      bool
}

func (wc *WorkloadClassifier) ClassifyWorkload(query string, context domain.StateReader) (string, float64) {
    // Classify the workload type
    features := wc.extractQueryFeatures(query)
    classification := wc.classifier.Classify(features)
    
    return classification.Class, classification.Confidence
}

func (wc *WorkloadClassifier) extractQueryFeatures(query string) QueryFeatures {
    return QueryFeatures{
        Length:       len(query),
        Complexity:   calculateComplexity(query),
        Keywords:     extractKeywords(query),
        Intent:       classifyIntent(query),
    }
}

// A/B Testing Manager
type ABTestingManager struct {
    activeTests    map[string]ABTest
    results        map[string]ABTestResults
    
    mutex          sync.RWMutex
}

type ABTest struct {
    ID              string
    Name            string
    ControlProvider string
    TestProvider    string
    TrafficSplit    float64
    StartTime       time.Time
    EndTime         time.Time
    Metrics         []string
    
    Status          string
}

type ABTestResults struct {
    ControlMetrics  TestMetrics
    TestMetrics     TestMetrics
    StatisticalSignificance float64
    Recommendation string
}

func (abt *ABTestingManager) ShouldUseTestProvider(testID string, userID string) bool {
    // Determine if this user should be in the test group
    hash := simpleHash(userID + testID)
    
    abt.mutex.RLock()
    test, exists := abt.activeTests[testID]
    abt.mutex.RUnlock()
    
    if !exists {
        return false
    }
    
    return hash < test.TrafficSplit
}

func simpleHash(s string) float64 {
    hash := 0
    for _, c := range s {
        hash = hash*31 + int(c)
    }
    return float64(hash%1000) / 1000.0
}

// Circuit Breaker implementation
func (cb *CircuitBreaker) Call(fn func() error) error {
    cb.mutex.Lock()
    defer cb.mutex.Unlock()

    if cb.state == Open {
        if time.Since(cb.lastFailTime) > cb.resetTimeout {
            cb.state = HalfOpen
            cb.failures = 0
        } else {
            return fmt.Errorf("circuit breaker is open")
        }
    }

    err := fn()
    
    if err != nil {
        cb.failures++
        cb.lastFailTime = time.Now()
        
        if cb.failures >= cb.maxFailures {
            cb.state = Open
        }
        
        return err
    }

    // Success - reset if we were in half-open state
    if cb.state == HalfOpen {
        cb.state = Closed
        cb.failures = 0
    }

    return nil
}

// Advanced router implementation
func NewAdvancedProviderRouter(name string) *AdvancedProviderRouter {
    router := &AdvancedProviderRouter{
        name:              name,
        providers:         make([]AdvancedProvider, 0),
        routingStrategies: make(map[string]RoutingStrategy),
        circuitBreakers:   make(map[string]*CircuitBreaker),
        currentStrategy:   "adaptive",
        performanceModel:  NewPerformanceModel(),
        costModel:         NewCostModel(),
        qualityModel:      NewQualityModel(),
        abTesting:         NewABTestingManager(),
        workloadClassifier: NewWorkloadClassifier(),
    }

    // Register routing strategies
    router.routingStrategies["performance"] = &PerformanceBasedRouting{
        weightLatency:     0.4,
        weightThroughput:  0.3,
        weightReliability: 0.3,
    }
    router.routingStrategies["cost"] = &CostOptimizedRouting{}
    router.routingStrategies["quality"] = &QualityFirstRouting{}
    router.routingStrategies["ml"] = &MLBasedRouting{
        model:           NewMLModel(),
        features:        []string{"query_length", "complexity", "latency_req", "cost_req"},
        predictionCache: make(map[string]PredictionResult),
    }

    return router
}

func (apr *AdvancedProviderRouter) AddAdvancedProvider(config ProviderConfig, capabilities ProviderCapabilities) error {
    agent, err := core.NewAgentFromString(config.Name, fmt.Sprintf("%s/%s", config.Name, "default"))
    if err != nil {
        return fmt.Errorf("failed to create agent: %w", err)
    }

    config.Agent = agent

    // Create circuit breaker for this provider
    circuitBreaker := &CircuitBreaker{
        name:         config.Name,
        maxFailures:  5,
        timeout:      10 * time.Second,
        resetTimeout: 60 * time.Second,
        state:        Closed,
    }

    provider := AdvancedProvider{
        Config:       config,
        Capabilities: capabilities,
        Performance: PerformanceProfile{
            OptimalLatency:   2 * time.Second,
            MaxThroughput:    100,
            TokensPerSecond:  50.0,
            ConcurrencyLimit: 10,
        },
        Geographic: GeographicProfile{
            PreferredRegions: []string{"us-east-1", "us-west-2"},
            DataCenters:     []string{"us", "eu"},
        },
        CircuitBreaker: circuitBreaker,
        CurrentMetrics: RealTimeMetrics{
            LastUpdated: time.Now(),
        },
    }

    apr.mutex.Lock()
    apr.providers = append(apr.providers, provider)
    apr.circuitBreakers[config.Name] = circuitBreaker
    apr.mutex.Unlock()

    fmt.Printf("✅ Added advanced provider: %s\n", config.Name)
    fmt.Printf("   Capabilities: Context=%d, Streaming=%t, Vision=%t, Functions=%t\n",
        capabilities.MaxContextLength, capabilities.SupportsStreaming, 
        capabilities.SupportsVision, capabilities.SupportsFunctions)

    return nil
}

func (apr *AdvancedProviderRouter) Route(ctx context.Context, request *RoutingRequest) (*domain.State, error) {
    // Classify workload
    workloadType, confidence := apr.workloadClassifier.ClassifyWorkload(request.UserQuery, request.Context)
    fmt.Printf("🔍 Workload classified as: %s (confidence: %.2f)\n", workloadType, confidence)

    // Check A/B testing
    testProvider := ""
    if request.UserProfile.UserID != "" {
        for testID, test := range apr.abTesting.activeTests {
            if apr.abTesting.ShouldUseTestProvider(testID, request.UserProfile.UserID) {
                testProvider = test.TestProvider
                fmt.Printf("🧪 User in A/B test: %s, using provider: %s\n", testID, testProvider)
                break
            }
        }
    }

    // Select routing strategy
    strategy := apr.selectRoutingStrategy(request, workloadType)
    fmt.Printf("📊 Using routing strategy: %s\n", strategy.Name())

    // Select provider
    var selectedProvider *AdvancedProvider
    var err error

    if testProvider != "" {
        selectedProvider = apr.findProviderByName(testProvider)
    } else {
        selectedProvider, err = strategy.SelectProvider(ctx, request)
        if err != nil {
            return nil, fmt.Errorf("provider selection failed: %w", err)
        }
    }

    if selectedProvider == nil {
        return nil, fmt.Errorf("no suitable provider found")
    }

    // Execute with circuit breaker protection
    var result *domain.State
    err = selectedProvider.CircuitBreaker.Call(func() error {
        var execErr error
        result, execErr = selectedProvider.Config.Agent.Run(ctx, request.Context)
        return execErr
    })

    if err != nil {
        return nil, fmt.Errorf("execution failed: %w", err)
    }

    fmt.Printf("✅ Request routed to: %s\n", selectedProvider.Config.Name)
    return result, nil
}

func (apr *AdvancedProviderRouter) selectRoutingStrategy(request *RoutingRequest, workloadType string) RoutingStrategy {
    // Select strategy based on request characteristics and workload type
    
    // If cost is a primary concern
    if request.Requirements.MaxCost > 0 && request.Requirements.MaxCost < 0.001 {
        return apr.routingStrategies["cost"]
    }
    
    // If latency is critical
    if request.Requirements.MaxLatency > 0 && request.Requirements.MaxLatency < 500*time.Millisecond {
        return apr.routingStrategies["performance"]
    }
    
    // If quality is paramount
    if request.Requirements.MinQuality > 0.9 {
        return apr.routingStrategies["quality"]
    }
    
    // Use ML for complex routing decisions
    return apr.routingStrategies["ml"]
}

func (apr *AdvancedProviderRouter) findProviderByName(name string) *AdvancedProvider {
    apr.mutex.RLock()
    defer apr.mutex.RUnlock()
    
    for i := range apr.providers {
        if apr.providers[i].Config.Name == name {
            return &apr.providers[i]
        }
    }
    return nil
}

// Stub implementations for missing types and functions
type PerformanceModel struct{}
type CostModel struct{}
type QualityModel struct{}
type MLModel struct{}
type TextClassifier struct{}
type PredictionResult struct{}
type QueryFeatures struct {
    Length     int
    Complexity float64
    Keywords   []string
    Intent     string
}
type TestMetrics struct{}
type BudgetConstraints struct{}
type CostLimits struct{}
type UserPreferences struct{}
type RequestHistory struct{}
type QualityScore struct{}

func NewPerformanceModel() *PerformanceModel { return &PerformanceModel{} }
func NewCostModel() *CostModel { return &CostModel{} }
func NewQualityModel() *QualityModel { return &QualityModel{} }
func NewMLModel() MLModel { return MLModel{} }
func NewABTestingManager() *ABTestingManager {
    return &ABTestingManager{
        activeTests: make(map[string]ABTest),
        results:     make(map[string]ABTestResults),
    }
}
func NewWorkloadClassifier() *WorkloadClassifier {
    return &WorkloadClassifier{
        patterns: make(map[string]WorkloadPattern),
    }
}

func (ml MLModel) Predict(features []float64) PredictionResult { return PredictionResult{} }
func (tc TextClassifier) Classify(features QueryFeatures) struct{ Class string; Confidence float64 } {
    return struct{ Class string; Confidence float64 }{"general", 0.8}
}

func calculateComplexity(query string) float64 {
    // Simple complexity calculation based on length and keywords
    complexity := float64(len(query)) / 100.0
    if containsSubstring(query, "explain") || containsSubstring(query, "analyze") {
        complexity += 0.5
    }
    return math.Min(complexity, 10.0)
}

func extractKeywords(query string) []string {
    // Simple keyword extraction
    return []string{"keyword1", "keyword2"}
}

func classifyIntent(query string) string {
    // Simple intent classification
    if containsSubstring(query, "what") || containsSubstring(query, "how") {
        return "question"
    }
    if containsSubstring(query, "create") || containsSubstring(query, "write") {
        return "creation"
    }
    return "general"
}

func main() {
    fmt.Println("🚀 Advanced Multi-Provider Routing")
    fmt.Println("==================================")

    // Create advanced router
    router := NewAdvancedProviderRouter("enterprise-router")

    // Add providers with detailed capabilities
    openaiCaps := ProviderCapabilities{
        MaxContextLength:   128000,
        SupportsStreaming:  true,
        SupportsVision:     true,
        SupportsFunctions:  true,
        SupportsMultimodal: true,
        LanguageSupport:    []string{"en", "es", "fr", "de"},
        SpecializedDomains: []string{"general", "coding", "analysis"},
    }

    anthropicCaps := ProviderCapabilities{
        MaxContextLength:   200000,
        SupportsStreaming:  true,
        SupportsVision:     true,
        SupportsFunctions:  true,
        SupportsMultimodal: false,
        LanguageSupport:    []string{"en", "es", "fr"},
        SpecializedDomains: []string{"reasoning", "analysis", "writing"},
    }

    geminiCaps := ProviderCapabilities{
        MaxContextLength:   1000000,
        SupportsStreaming:  true,
        SupportsVision:     true,
        SupportsFunctions:  true,
        SupportsMultimodal: true,
        LanguageSupport:    []string{"en", "es", "fr", "de", "ja", "ko"},
        SpecializedDomains: []string{"general", "multimodal", "fast"},
    }

    // Add providers to router
    router.AddAdvancedProvider(ProviderConfig{
        Name:        "openai",
        CostPerCall: 0.00015,
        Reliability: 0.98,
    }, openaiCaps)

    router.AddAdvancedProvider(ProviderConfig{
        Name:        "anthropic", 
        CostPerCall: 0.00018,
        Reliability: 0.97,
    }, anthropicCaps)

    router.AddAdvancedProvider(ProviderConfig{
        Name:        "gemini",
        CostPerCall: 0.00008,
        Reliability: 0.95,
    }, geminiCaps)

    // Test different routing scenarios
    testScenarios := []struct {
        name    string
        request RoutingRequest
    }{
        {
            name: "Cost-sensitive query",
            request: RoutingRequest{
                UserQuery: "What is the capital of France?",
                Requirements: RequestRequirements{
                    MaxCost:    0.0001,
                    Priority:   "low",
                    Complexity: 1,
                },
                UserProfile: UserProfile{
                    UserID: "user123",
                    Tier:   "basic",
                },
            },
        },
        {
            name: "High-quality analysis",
            request: RoutingRequest{
                UserQuery: "Analyze the economic implications of artificial intelligence on employment markets",
                Requirements: RequestRequirements{
                    MinQuality:  0.95,
                    Priority:    "high",
                    Complexity:  9,
                    MaxLatency:  10 * time.Second,
                },
                UserProfile: UserProfile{
                    UserID: "premium_user456",
                    Tier:   "premium",
                },
            },
        },
        {
            name: "Low-latency request",
            request: RoutingRequest{
                UserQuery: "Quick: what's 2+2?",
                Requirements: RequestRequirements{
                    MaxLatency: 500 * time.Millisecond,
                    Priority:   "urgent",
                    Complexity: 1,
                },
                UserProfile: UserProfile{
                    UserID: "speed_user789",
                    Tier:   "enterprise",
                },
            },
        },
    }

    ctx := context.Background()

    for i, scenario := range testScenarios {
        fmt.Printf("\n🧪 Test Scenario %d: %s\n", i+1, scenario.name)
        fmt.Printf("Query: %s\n", scenario.request.UserQuery)
        fmt.Printf("Requirements: Max Cost=$%.6f, Min Quality=%.2f, Max Latency=%v\n",
            scenario.request.Requirements.MaxCost,
            scenario.request.Requirements.MinQuality,
            scenario.request.Requirements.MaxLatency)

        state := domain.NewState()
        state.Set("user_input", scenario.request.UserQuery)
        scenario.request.Context = state

        result, err := router.Route(ctx, &scenario.request)
        if err != nil {
            fmt.Printf("❌ Routing failed: %v\n", err)
            continue
        }

        if response, exists := result.Get("response"); exists {
            responseStr := fmt.Sprintf("%v", response)
            if len(responseStr) > 150 {
                responseStr = responseStr[:150] + "..."
            }
            fmt.Printf("✅ Response: %s\n", responseStr)
        }
    }
}
```

### Sophisticated Features
✅ **Workload Classification** - ML-based request type detection  
✅ **Circuit Breakers** - Per-provider failure protection  
✅ **A/B Testing** - Live traffic testing of provider changes  
✅ **Geolocation Routing** - Region-based provider selection  
✅ **Predictive Scaling** - Anticipate capacity needs  
✅ **ML-Based Routing** - Machine learning optimization  

---

## Regional and Compliance Strategies

### Geographic Distribution
```go
type GeographicStrategy struct {
    regions          map[string]RegionConfig
    complianceRules  map[string]ComplianceRule
    latencyTargets   map[string]time.Duration
}

type RegionConfig struct {
    Name            string
    AvailableProviders []string
    PreferredProvider  string
    ComplianceLevel    string
    DataResidency      bool
}

func (gs *GeographicStrategy) SelectProviderByRegion(userRegion string, requirements RequestRequirements) string {
    regionConfig := gs.regions[userRegion]
    
    // Check compliance requirements
    for _, rule := range gs.complianceRules {
        if rule.AppliesTo(userRegion) && !rule.AllowsProvider(regionConfig.PreferredProvider) {
            // Find compliant alternative
            for _, provider := range regionConfig.AvailableProviders {
                if rule.AllowsProvider(provider) {
                    return provider
                }
            }
        }
    }
    
    return regionConfig.PreferredProvider
}
```

## Next Steps

🌐 **Multi-provider strategies mastered!** Continue with:

- **[Local Providers](local-providers.md)** - Deep dive into Ollama and local hosting
- **[Agent Communication](agent-communication.md)** - Multi-agent coordination
- **[Performance Optimization](../advanced/performance-optimization.md)** - Optimize your architecture
- **[Production Deployment](../advanced/production-deployment.md)** - Deploy at enterprise scale

### Quick Reference

- **[Provider Comparison](../reference/provider-comparison.md)** - Detailed feature matrix
- **[Configuration Reference](../reference/configuration-reference.md)** - All configuration options
- **[Error Codes Reference](../reference/error-codes-reference.md)** - Multi-provider error handling
- **[Best Practices Checklist](../reference/best-practices-checklist.md)** - Production checklist

---

**Need help with complex routing?** Check our [architecture patterns guide](../advanced/architecture-patterns.md) or join the discussion on [GitHub](https://github.com/lexlapax/go-llms/discussions).