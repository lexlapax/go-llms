# Provider Registry: Dynamic Registration and Discovery

> **[Project Root](/) / [Documentation](../..) / [Technical Documentation](../../technical) / [Providers](../../technical/providers) / Provider Registry**

The provider registry system in Go-LLMs enables dynamic registration, discovery, and management of LLM providers at runtime. This document covers the registry architecture, registration mechanisms, discovery patterns, and advanced provider management features.

## Registry Architecture

### Core Components

```go
// Provider registry manages all available providers
type ProviderRegistry struct {
    providers    map[string]ProviderFactory
    metadata     map[string]ProviderMetadata
    validators   map[string]ConfigValidator
    mu           sync.RWMutex
    hooks        RegistryHooks
    discoverers  []ProviderDiscoverer
}

// ProviderFactory creates provider instances
type ProviderFactory func(config map[string]interface{}) (Provider, error)

// ProviderMetadata describes provider capabilities
type ProviderMetadata struct {
    Name            string
    DisplayName     string
    Description     string
    Version         string
    Author          string
    SupportedModels []ModelInfo
    Capabilities    []Capability
    ConfigSchema    *jsonschema.Schema
    Documentation   string
    Examples        []UsageExample
}

// ConfigValidator validates provider configuration
type ConfigValidator func(config map[string]interface{}) error

// RegistryHooks allow customization of registry behavior
type RegistryHooks struct {
    OnRegister   func(name string, metadata ProviderMetadata) error
    OnDeregister func(name string) error
    OnCreate     func(name string, provider Provider) error
    OnError      func(name string, err error)
}
```

### Registry Interface

```go
// Registry defines the provider registry interface
type Registry interface {
    // Registration
    Register(name string, factory ProviderFactory, metadata ProviderMetadata) error
    Deregister(name string) error
    
    // Discovery
    List() []ProviderInfo
    Get(name string) (ProviderFactory, bool)
    GetMetadata(name string) (ProviderMetadata, bool)
    
    // Creation
    Create(name string, config map[string]interface{}) (Provider, error)
    CreateFromURL(url string) (Provider, error)
    
    // Validation
    ValidateConfig(name string, config map[string]interface{}) error
    
    // Search and filtering
    Search(query string) []ProviderInfo
    Filter(predicate func(ProviderInfo) bool) []ProviderInfo
}
```

---

## Provider Registration

### Static Registration

```go
// Static registration using init functions
package openai

import (
    "github.com/lexlapax/go-llms/pkg/llm/provider/registry"
)

func init() {
    // Register OpenAI provider
    registry.Register("openai", 
        NewOpenAIProvider,
        ProviderMetadata{
            Name:        "openai",
            DisplayName: "OpenAI",
            Description: "OpenAI GPT models provider",
            Version:     "1.0.0",
            Author:      "go-llms",
            Capabilities: []Capability{
                CapabilityTextGeneration,
                CapabilityFunctionCalling,
                CapabilityVision,
                CapabilityStreaming,
            },
            SupportedModels: []ModelInfo{
                {ID: "gpt-4o", MaxTokens: 128000, SupportsVision: true},
                {ID: "gpt-4o-mini", MaxTokens: 128000, SupportsVision: true},
                {ID: "gpt-3.5-turbo", MaxTokens: 16384},
            },
            ConfigSchema: &jsonschema.Schema{
                Type: "object",
                Properties: map[string]*jsonschema.Schema{
                    "api_key": {
                        Type:        "string",
                        Description: "OpenAI API key",
                        Required:    true,
                    },
                    "organization": {
                        Type:        "string",
                        Description: "OpenAI organization ID",
                    },
                    "base_url": {
                        Type:        "string",
                        Description: "API base URL",
                        Default:     "https://api.openai.com",
                    },
                },
            },
        },
    )
}

// Factory function
func NewOpenAIProvider(config map[string]interface{}) (Provider, error) {
    opts := OpenAIOptions{}
    
    // Parse configuration
    if apiKey, ok := config["api_key"].(string); ok {
        opts.APIKey = apiKey
    } else {
        return nil, errors.New("api_key is required")
    }
    
    if org, ok := config["organization"].(string); ok {
        opts.Organization = org
    }
    
    if baseURL, ok := config["base_url"].(string); ok {
        opts.BaseURL = baseURL
    }
    
    return NewOpenAI(opts)
}
```

### Dynamic Registration

```go
// Dynamic provider registration at runtime
type DynamicRegistry struct {
    *ProviderRegistry
    loader ProviderLoader
}

// ProviderLoader loads providers from external sources
type ProviderLoader interface {
    Load(source string) ([]ProviderDefinition, error)
}

// ProviderDefinition defines a dynamically loaded provider
type ProviderDefinition struct {
    Name     string
    Type     string // "plugin", "wasm", "script", "remote"
    Source   string // path, URL, or script content
    Metadata ProviderMetadata
    Config   map[string]interface{}
}

// Load and register providers dynamically
func (r *DynamicRegistry) LoadProviders(source string) error {
    definitions, err := r.loader.Load(source)
    if err != nil {
        return fmt.Errorf("failed to load providers: %w", err)
    }
    
    for _, def := range definitions {
        factory, err := r.createFactory(def)
        if err != nil {
            log.Printf("Failed to create factory for %s: %v", def.Name, err)
            continue
        }
        
        if err := r.Register(def.Name, factory, def.Metadata); err != nil {
            log.Printf("Failed to register %s: %v", def.Name, err)
            continue
        }
        
        log.Printf("Successfully registered provider: %s", def.Name)
    }
    
    return nil
}

// Plugin-based provider loading
func (r *DynamicRegistry) LoadPlugin(path string) error {
    // Open plugin
    plug, err := plugin.Open(path)
    if err != nil {
        return fmt.Errorf("failed to open plugin: %w", err)
    }
    
    // Look for provider symbol
    providerSym, err := plug.Lookup("Provider")
    if err != nil {
        return fmt.Errorf("plugin missing Provider symbol: %w", err)
    }
    
    // Assert provider interface
    provider, ok := providerSym.(ProviderPlugin)
    if !ok {
        return fmt.Errorf("invalid provider plugin interface")
    }
    
    // Register provider
    return r.Register(
        provider.Name(),
        provider.Factory(),
        provider.Metadata(),
    )
}

// ProviderPlugin interface for plugins
type ProviderPlugin interface {
    Name() string
    Factory() ProviderFactory
    Metadata() ProviderMetadata
}
```

### Configuration-Based Registration

```go
// YAML-based provider configuration
type ProviderConfig struct {
    Providers []struct {
        Name     string                 `yaml:"name"`
        Type     string                 `yaml:"type"`
        Enabled  bool                   `yaml:"enabled"`
        Config   map[string]interface{} `yaml:"config"`
        Metadata struct {
            DisplayName  string   `yaml:"display_name"`
            Description  string   `yaml:"description"`
            Capabilities []string `yaml:"capabilities"`
        } `yaml:"metadata"`
    } `yaml:"providers"`
}

// Load providers from configuration file
func LoadProvidersFromConfig(configPath string) error {
    data, err := os.ReadFile(configPath)
    if err != nil {
        return err
    }
    
    var config ProviderConfig
    if err := yaml.Unmarshal(data, &config); err != nil {
        return err
    }
    
    registry := GetGlobalRegistry()
    
    for _, providerCfg := range config.Providers {
        if !providerCfg.Enabled {
            continue
        }
        
        // Get factory for provider type
        factory, err := GetFactoryForType(providerCfg.Type)
        if err != nil {
            log.Printf("Unknown provider type %s: %v", providerCfg.Type, err)
            continue
        }
        
        // Create metadata
        metadata := ProviderMetadata{
            Name:         providerCfg.Name,
            DisplayName:  providerCfg.Metadata.DisplayName,
            Description:  providerCfg.Metadata.Description,
            Capabilities: parseCapabilities(providerCfg.Metadata.Capabilities),
        }
        
        // Register provider
        if err := registry.Register(providerCfg.Name, factory, metadata); err != nil {
            log.Printf("Failed to register %s: %v", providerCfg.Name, err)
        }
    }
    
    return nil
}
```

---

## Provider Discovery

### Basic Discovery

```go
// Discover available providers
func (r *ProviderRegistry) DiscoverProviders() []ProviderInfo {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    var providers []ProviderInfo
    
    for name, metadata := range r.metadata {
        providers = append(providers, ProviderInfo{
            Name:         name,
            Metadata:     metadata,
            Available:    r.checkAvailability(name),
            HealthStatus: r.checkHealth(name),
}
    }
    
    // Sort by name for consistent ordering
    sort.Slice(providers, func(i, j int) bool {
        return providers[i].Name < providers[j].Name
}
    
    return providers
}

// Check provider availability
func (r *ProviderRegistry) checkAvailability(name string) bool {
    // Check if required environment variables are set
    switch name {
    case "openai":
        return os.Getenv("OPENAI_API_KEY") != ""
    case "anthropic":
        return os.Getenv("ANTHROPIC_API_KEY") != ""
    case "ollama":
        // Check if Ollama is running
        return r.checkOllamaConnection()
    default:
        return true
    }
}

// Health check for providers
func (r *ProviderRegistry) checkHealth(name string) HealthStatus {
    factory, ok := r.providers[name]
    if !ok {
        return HealthStatusUnknown
    }
    
    // Create temporary instance for health check
    provider, err := factory(map[string]interface{}{
        "api_key": os.Getenv(strings.ToUpper(name) + "_API_KEY"),
}
    if err != nil {
        return HealthStatusUnhealthy
    }
    
    // Check if provider implements HealthCheckable
    if hc, ok := provider.(HealthCheckable); ok {
        if err := hc.HealthCheck(context.Background()); err != nil {
            return HealthStatusUnhealthy
        }
    }
    
    return HealthStatusHealthy
}

// HealthCheckable interface for providers
type HealthCheckable interface {
    HealthCheck(ctx context.Context) error
}
```

### Advanced Discovery

```go
// Capability-based discovery
func (r *ProviderRegistry) DiscoverByCapability(capabilities ...Capability) []ProviderInfo {
    return r.Filter(func(info ProviderInfo) bool {
        for _, required := range capabilities {
            found := false
            for _, cap := range info.Metadata.Capabilities {
                if cap == required {
                    found = true
                    break
                }
            }
            if !found {
                return false
            }
        }
        return true
}
}

// Model-based discovery
func (r *ProviderRegistry) DiscoverByModel(modelID string) []ProviderInfo {
    return r.Filter(func(info ProviderInfo) bool {
        for _, model := range info.Metadata.SupportedModels {
            if model.ID == modelID {
                return true
            }
        }
        return false
}
}

// Cost-based discovery
func (r *ProviderRegistry) DiscoverByCost(maxCostPerToken float64) []ProviderInfo {
    providers := r.Filter(func(info ProviderInfo) bool {
        for _, model := range info.Metadata.SupportedModels {
            if model.CostPerToken <= maxCostPerToken {
                return true
            }
        }
        return false
}
    
    // Sort by cost
    sort.Slice(providers, func(i, j int) bool {
        return getLowestCost(providers[i]) < getLowestCost(providers[j])
}
    
    return providers
}

// Composite discovery with scoring
type DiscoveryQuery struct {
    RequiredCapabilities []Capability
    PreferredModels      []string
    MaxCost              float64
    MinReliability       float64
    PreferLocal          bool
}

func (r *ProviderRegistry) DiscoverWithScoring(query DiscoveryQuery) []ScoredProvider {
    providers := r.List()
    var scored []ScoredProvider
    
    for _, provider := range providers {
        score := r.scoreProvider(provider, query)
        if score > 0 {
            scored = append(scored, ScoredProvider{
                Provider: provider,
                Score:    score,
                Reasons:  r.explainScore(provider, query),
}
        }
    }
    
    // Sort by score descending
    sort.Slice(scored, func(i, j int) bool {
        return scored[i].Score > scored[j].Score
}
    
    return scored
}
```

### Service Discovery Integration

```go
// Consul-based provider discovery
type ConsulDiscoverer struct {
    client   *consul.Client
    service  string
    registry *ProviderRegistry
}

func (d *ConsulDiscoverer) Discover(ctx context.Context) error {
    // Query Consul for provider services
    services, _, err := d.client.Health().Service(d.service, "", true, nil)
    if err != nil {
        return fmt.Errorf("consul query failed: %w", err)
    }
    
    for _, service := range services {
        // Extract provider metadata from service tags
        metadata := d.parseServiceMetadata(service)
        
        // Create remote provider factory
        factory := d.createRemoteFactory(service.Service.Address, service.Service.Port)
        
        // Register provider
        if err := d.registry.Register(metadata.Name, factory, metadata); err != nil {
            log.Printf("Failed to register discovered provider %s: %v", metadata.Name, err)
        }
    }
    
    return nil
}

// Kubernetes-based provider discovery
type K8sDiscoverer struct {
    client    kubernetes.Interface
    namespace string
    selector  labels.Selector
    registry  *ProviderRegistry
}

func (d *K8sDiscoverer) Discover(ctx context.Context) error {
    // List services matching selector
    services, err := d.client.CoreV1().Services(d.namespace).List(ctx, metav1.ListOptions{
        LabelSelector: d.selector.String(),
}
    if err != nil {
        return fmt.Errorf("k8s list failed: %w", err)
    }
    
    for _, svc := range services.Items {
        // Check if service has provider annotation
        if providerType, ok := svc.Annotations["llm-provider/type"]; ok {
            metadata := d.parseServiceAnnotations(svc.Annotations)
            
            // Create provider factory for k8s service
            factory := d.createK8sFactory(&svc, providerType)
            
            // Register provider
            if err := d.registry.Register(metadata.Name, factory, metadata); err != nil {
                log.Printf("Failed to register k8s provider %s: %v", metadata.Name, err)
            }
        }
    }
    
    return nil
}
```

---

## Provider Lifecycle Management

### Provider Creation

```go
// Enhanced provider creation with lifecycle hooks
func (r *ProviderRegistry) Create(name string, config map[string]interface{}) (Provider, error) {
    r.mu.RLock()
    factory, exists := r.providers[name]
    validator := r.validators[name]
    r.mu.RUnlock()
    
    if !exists {
        return nil, fmt.Errorf("provider %s not found", name)
    }
    
    // Validate configuration
    if validator != nil {
        if err := validator(config); err != nil {
            return nil, fmt.Errorf("invalid configuration: %w", err)
        }
    }
    
    // Pre-creation hook
    if r.hooks.OnCreate != nil {
        if err := r.hooks.OnCreate(name, nil); err != nil {
            return nil, fmt.Errorf("pre-creation hook failed: %w", err)
        }
    }
    
    // Create provider instance
    provider, err := factory(config)
    if err != nil {
        if r.hooks.OnError != nil {
            r.hooks.OnError(name, err)
        }
        return nil, err
    }
    
    // Wrap with lifecycle management
    wrapped := &LifecycleProvider{
        Provider:  provider,
        name:      name,
        config:    config,
        registry:  r,
        createdAt: time.Now(),
        metrics:   NewProviderMetrics(name),
    }
    
    // Post-creation hook
    if r.hooks.OnCreate != nil {
        if err := r.hooks.OnCreate(name, wrapped); err != nil {
            // Cleanup on hook failure
            if closer, ok := provider.(io.Closer); ok {
                closer.Close()
            }
            return nil, fmt.Errorf("post-creation hook failed: %w", err)
        }
    }
    
    return wrapped, nil
}

// LifecycleProvider wraps providers with lifecycle management
type LifecycleProvider struct {
    Provider
    name      string
    config    map[string]interface{}
    registry  *ProviderRegistry
    createdAt time.Time
    metrics   *ProviderMetrics
    mu        sync.RWMutex
}

// Override methods to add lifecycle tracking
func (p *LifecycleProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
    start := time.Now()
    p.metrics.RecordRequest()
    
    resp, err := p.Provider.Complete(ctx, req)
    
    duration := time.Since(start)
    p.metrics.RecordDuration(duration)
    
    if err != nil {
        p.metrics.RecordError(err)
        return nil, err
    }
    
    p.metrics.RecordTokens(resp.Usage.PromptTokens, resp.Usage.CompletionTokens)
    return resp, nil
}

// Graceful shutdown
func (p *LifecycleProvider) Shutdown(ctx context.Context) error {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    // Notify registry
    if deregister := p.registry.hooks.OnDeregister; deregister != nil {
        if err := deregister(p.name); err != nil {
            log.Printf("Deregister hook failed: %v", err)
        }
    }
    
    // Close provider if it implements io.Closer
    if closer, ok := p.Provider.(io.Closer); ok {
        return closer.Close()
    }
    
    return nil
}
```

### Provider Pool Management

```go
// ProviderPool manages a pool of provider instances
type ProviderPool struct {
    name        string
    factory     ProviderFactory
    config      map[string]interface{}
    minSize     int
    maxSize     int
    idleTimeout time.Duration
    providers   chan Provider
    metrics     *PoolMetrics
    mu          sync.Mutex
}

func NewProviderPool(name string, factory ProviderFactory, config PoolConfig) *ProviderPool {
    pool := &ProviderPool{
        name:        name,
        factory:     factory,
        config:      config.ProviderConfig,
        minSize:     config.MinSize,
        maxSize:     config.MaxSize,
        idleTimeout: config.IdleTimeout,
        providers:   make(chan Provider, config.MaxSize),
        metrics:     NewPoolMetrics(name),
    }
    
    // Pre-create minimum providers
    for i := 0; i < pool.minSize; i++ {
        if provider, err := pool.createProvider(); err == nil {
            pool.providers <- provider
        }
    }
    
    // Start maintenance routine
    go pool.maintain()
    
    return pool
}

// Get provider from pool
func (p *ProviderPool) Get(ctx context.Context) (Provider, error) {
    select {
    case provider := <-p.providers:
        p.metrics.RecordCheckout()
        return &PooledProvider{
            Provider: provider,
            pool:     p,
            checkedOut: time.Now(),
        }, nil
        
    case <-ctx.Done():
        return nil, ctx.Err()
        
    default:
        // Create new provider if under max size
        p.mu.Lock()
        if len(p.providers) < p.maxSize {
            p.mu.Unlock()
            provider, err := p.createProvider()
            if err != nil {
                return nil, err
            }
            p.metrics.RecordCheckout()
            return &PooledProvider{
                Provider: provider,
                pool:     p,
                checkedOut: time.Now(),
            }, nil
        }
        p.mu.Unlock()
        
        // Wait for available provider
        select {
        case provider := <-p.providers:
            p.metrics.RecordCheckout()
            return &PooledProvider{
                Provider: provider,
                pool:     p,
                checkedOut: time.Now(),
            }, nil
            
        case <-ctx.Done():
            return nil, ctx.Err()
        }
    }
}

// Return provider to pool
func (p *ProviderPool) Put(provider Provider) {
    // Unwrap if pooled
    if pooled, ok := provider.(*PooledProvider); ok {
        provider = pooled.Provider
        p.metrics.RecordCheckin(time.Since(pooled.checkedOut))
    }
    
    select {
    case p.providers <- provider:
        // Provider returned to pool
    default:
        // Pool full, close provider
        if closer, ok := provider.(io.Closer); ok {
            closer.Close()
        }
    }
}

// Maintain pool health
func (p *ProviderPool) maintain() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        p.mu.Lock()
        currentSize := len(p.providers)
        
        // Ensure minimum size
        for currentSize < p.minSize {
            if provider, err := p.createProvider(); err == nil {
                p.providers <- provider
                currentSize++
            } else {
                log.Printf("Failed to create provider for pool: %v", err)
                break
            }
        }
        
        // Remove idle providers over max idle time
        for currentSize > p.minSize {
            select {
            case provider := <-p.providers:
                if pooled, ok := provider.(*PooledProvider); ok {
                    if time.Since(pooled.lastUsed) > p.idleTimeout {
                        if closer, ok := provider.(io.Closer); ok {
                            closer.Close()
                        }
                        currentSize--
                        continue
                    }
                }
                // Put back
                p.providers <- provider
                
            default:
                break
            }
        }
        
        p.mu.Unlock()
        
        // Update metrics
        p.metrics.UpdatePoolSize(currentSize)
    }
}
```

---

## Registry Extensions

### Provider Aliasing

```go
// Alias support for provider compatibility
func (r *ProviderRegistry) RegisterAlias(alias, target string) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    // Check target exists
    if _, exists := r.providers[target]; !exists {
        return fmt.Errorf("target provider %s not found", target)
    }
    
    // Create alias
    r.providers[alias] = r.providers[target]
    r.metadata[alias] = r.metadata[target]
    
    // Update metadata to indicate alias
    aliasedMeta := r.metadata[alias]
    aliasedMeta.Name = alias
    aliasedMeta.Description = fmt.Sprintf("Alias for %s. %s", target, aliasedMeta.Description)
    r.metadata[alias] = aliasedMeta
    
    return nil
}

// Common aliases
func RegisterCommonAliases(r *ProviderRegistry) {
    aliases := map[string]string{
        "gpt":     "openai",
        "claude":  "anthropic",
        "palm":    "google",
        "bard":    "google",
        "local":   "ollama",
    }
    
    for alias, target := range aliases {
        if err := r.RegisterAlias(alias, target); err != nil {
            log.Printf("Failed to register alias %s -> %s: %v", alias, target, err)
        }
    }
}
```

### Provider Composition

```go
// CompositeProvider combines multiple providers
type CompositeProvider struct {
    providers []Provider
    strategy  SelectionStrategy
    fallback  FallbackStrategy
}

type SelectionStrategy interface {
    Select(providers []Provider, request *CompletionRequest) Provider
}

type FallbackStrategy interface {
    ShouldFallback(err error) bool
    NextProvider(current Provider, providers []Provider) Provider
}

// Register composite providers
func RegisterCompositeProvider(r *ProviderRegistry) {
    r.Register("multi", 
        func(config map[string]interface{}) (Provider, error) {
            // Parse provider list
            providerNames, ok := config["providers"].([]string)
            if !ok {
                return nil, errors.New("providers list required")
            }
            
            // Create provider instances
            var providers []Provider
            for _, name := range providerNames {
                provider, err := r.Create(name, config)
                if err != nil {
                    log.Printf("Failed to create provider %s: %v", name, err)
                    continue
                }
                providers = append(providers, provider)
            }
            
            if len(providers) == 0 {
                return nil, errors.New("no providers available")
            }
            
            // Create selection strategy
            strategy := &LoadBalancingStrategy{
                method: config["balance_method"].(string),
            }
            
            // Create fallback strategy
            fallback := &SmartFallbackStrategy{
                maxRetries: 3,
                backoff:    time.Second,
            }
            
            return &CompositeProvider{
                providers: providers,
                strategy:  strategy,
                fallback:  fallback,
            }, nil
        },
        ProviderMetadata{
            Name:        "multi",
            DisplayName: "Multi-Provider",
            Description: "Composite provider with fallback and load balancing",
            Capabilities: []Capability{
                CapabilityTextGeneration,
                CapabilityHighAvailability,
                CapabilityLoadBalancing,
            },
        },
    )
}
```

### Provider Middleware

```go
// Middleware wraps providers with additional functionality
type ProviderMiddleware func(Provider) Provider

// Apply middleware to all providers
func (r *ProviderRegistry) UseMiddleware(middleware ...ProviderMiddleware) {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    for name, factory := range r.providers {
        originalFactory := factory
        
        r.providers[name] = func(config map[string]interface{}) (Provider, error) {
            provider, err := originalFactory(config)
            if err != nil {
                return nil, err
            }
            
            // Apply middleware in order
            for _, mw := range middleware {
                provider = mw(provider)
            }
            
            return provider, nil
        }
    }
}

// Example middleware
func LoggingMiddleware(logger *zap.Logger) ProviderMiddleware {
    return func(provider Provider) Provider {
        return &LoggingProvider{
            Provider: provider,
            logger:   logger,
        }
    }
}

func RateLimitingMiddleware(limiter *rate.Limiter) ProviderMiddleware {
    return func(provider Provider) Provider {
        return &RateLimitedProvider{
            Provider: provider,
            limiter:  limiter,
        }
    }
}

func CachingMiddleware(cache Cache) ProviderMiddleware {
    return func(provider Provider) Provider {
        return &CachedProvider{
            Provider: provider,
            cache:    cache,
        }
    }
}

// Apply middleware
registry.UseMiddleware(
    LoggingMiddleware(logger),
    RateLimitingMiddleware(limiter),
    CachingMiddleware(cache),
)
```

---

## Configuration Examples

### Registry Configuration

```yaml
# provider-registry.yaml
registry:
  # Discovery settings
  discovery:
    enabled: true
    sources:
      - type: static
        path: /etc/go-llms/providers
      - type: consul
        address: consul.service.consul:8500
        service: llm-provider
      - type: kubernetes
        namespace: llm-system
        selector: app=llm-provider
    interval: 60s
  
  # Pool settings
  pooling:
    enabled: true
    default:
      min_size: 1
      max_size: 10
      idle_timeout: 5m
    per_provider:
      openai:
        min_size: 2
        max_size: 20
      ollama:
        min_size: 1
        max_size: 5
  
  # Middleware
  middleware:
    - type: logging
      level: info
    - type: metrics
      enabled: true
    - type: rate_limiting
      default_rps: 100
    - type: caching
      ttl: 5m
  
  # Aliases
  aliases:
    gpt: openai
    claude: anthropic
    local: ollama
  
  # Composite providers
  composites:
    high_availability:
      providers:
        - openai
        - anthropic
        - gemini
      strategy: failover
      health_check_interval: 30s
    
    load_balanced:
      providers:
        - openai
        - openai  # Multiple instances
        - anthropic
      strategy: round_robin
```

### Provider Definition Files

```yaml
# providers/custom-provider.yaml
name: custom_llm
type: http
enabled: true

metadata:
  display_name: "Custom LLM Provider"
  description: "Internal LLM service"
  version: "1.0.0"
  author: "Internal Team"
  capabilities:
    - text_generation
    - streaming
  
config:
  endpoint: https://llm.internal.company.com
  auth:
    type: bearer
    token: ${CUSTOM_LLM_TOKEN}
  timeout: 30s
  retry:
    max_attempts: 3
    backoff: exponential
  
models:
  - id: custom-base
    max_tokens: 4096
    cost_per_token: 0.00001
  - id: custom-large
    max_tokens: 8192
    cost_per_token: 0.00005
    
health_check:
  endpoint: /health
  interval: 30s
  timeout: 5s
```

---

## Best Practices

### 1. Registration
- Use static registration for built-in providers
- Implement proper error handling in factory functions
- Include comprehensive metadata for discovery
- Validate configurations before creating instances

### 2. Discovery
- Cache discovery results to reduce overhead
- Implement health checks for availability
- Use capability-based discovery for flexibility
- Consider service mesh integration for dynamic environments

### 3. Lifecycle Management
- Use provider pools for frequently used providers
- Implement graceful shutdown procedures
- Monitor provider health and performance
- Handle provider rotation for key updates

### 4. Security
- Validate provider sources for dynamic loading
- Implement authentication for remote providers
- Use secure configuration storage
- Audit provider access and usage

### 5. Performance
- Pool provider instances to reduce creation overhead
- Implement caching at the registry level
- Use lazy loading for rarely used providers
- Monitor and optimize discovery operations

---

## Monitoring and Diagnostics

### Registry Metrics

```go
// Registry metrics collection
type RegistryMetrics struct {
    registrations   prometheus.Counter
    deregistrations prometheus.Counter
    creations       prometheus.Counter
    discoveries     prometheus.Counter
    errors          *prometheus.CounterVec
    providerCount   prometheus.Gauge
}

func (r *ProviderRegistry) ExportMetrics() *RegistryMetrics {
    return &RegistryMetrics{
        registrations: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "provider_registry_registrations_total",
            Help: "Total number of provider registrations",
        }),
        deregistrations: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "provider_registry_deregistrations_total",
            Help: "Total number of provider deregistrations",
        }),
        creations: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "provider_registry_creations_total",
            Help: "Total number of provider instances created",
        }),
        discoveries: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "provider_registry_discoveries_total",
            Help: "Total number of discovery operations",
        }),
        errors: prometheus.NewCounterVec(prometheus.CounterOpts{
            Name: "provider_registry_errors_total",
            Help: "Total number of registry errors",
        }, []string{"operation", "provider"}),
        providerCount: prometheus.NewGauge(prometheus.GaugeOpts{
            Name: "provider_registry_providers",
            Help: "Current number of registered providers",
        }),
    }
}
```

### Health Endpoint

```go
// Registry health check endpoint
func (r *ProviderRegistry) HealthHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, req *http.Request) {
        health := r.CheckHealth()
        
        status := http.StatusOK
        if health.Status != "healthy" {
            status = http.StatusServiceUnavailable
        }
        
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(status)
        json.NewEncoder(w).Encode(health)
    }
}

type RegistryHealth struct {
    Status    string                     `json:"status"`
    Providers map[string]ProviderHealth  `json:"providers"`
    Timestamp time.Time                  `json:"timestamp"`
}

type ProviderHealth struct {
    Available bool   `json:"available"`
    Healthy   bool   `json:"healthy"`
    Message   string `json:"message,omitempty"`
}
```

---

## Next Steps

- **[Provider Metadata](metadata.md)** - Capabilities and configuration schemas
- **[Provider Interface](../../technical/api-reference/providers.md)** - Core provider API
- **[Custom Providers](../../user-guide/advanced/custom-providers.md)** - Building custom providers
- **[Provider Setup Guide](../../user-guide/guides/provider-setup.md)** - Configuration guide
- **[Multi-Provider Strategies](../../user-guide/guides/multi-provider-strategies.md)** - Advanced patterns