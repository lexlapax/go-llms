# Provider Metadata: Capabilities and Configuration

> **[Project Root](/) / [Documentation](../..) / [Technical Documentation](../../technical) / [Providers](../../technical/providers) / Provider Metadata**

Provider metadata defines capabilities, configuration schemas, and runtime characteristics for LLM providers in Go-LLMs. This document covers metadata structures, capability definitions, configuration validation, and dynamic metadata discovery.

## Metadata Architecture

### Core Metadata Structure

```go
// ProviderMetadata defines comprehensive provider information
type ProviderMetadata struct {
    // Basic information
    Name           string                 `json:"name"`
    DisplayName    string                 `json:"display_name"`
    Description    string                 `json:"description"`
    Version        string                 `json:"version"`
    Author         string                 `json:"author"`
    License        string                 `json:"license,omitempty"`
    Homepage       string                 `json:"homepage,omitempty"`
    Documentation  string                 `json:"documentation,omitempty"`
    
    // Capabilities
    Capabilities   []Capability           `json:"capabilities"`
    SupportedModels []ModelInfo           `json:"supported_models"`
    Features       map[string]interface{} `json:"features"`
    
    // Configuration
    ConfigSchema   *jsonschema.Schema     `json:"config_schema"`
    DefaultConfig  map[string]interface{} `json:"default_config,omitempty"`
    Examples       []ConfigExample        `json:"examples,omitempty"`
    
    // Runtime characteristics
    Performance    PerformanceMetrics     `json:"performance"`
    Reliability    ReliabilityMetrics     `json:"reliability"`
    Cost           CostMetrics            `json:"cost"`
    
    // Integration
    Tags           []string               `json:"tags,omitempty"`
    Categories     []string               `json:"categories,omitempty"`
    Dependencies   []Dependency           `json:"dependencies,omitempty"`
    
    // Lifecycle
    Status         ProviderStatus         `json:"status"`
    Deprecation    *DeprecationInfo       `json:"deprecation,omitempty"`
    Migration      *MigrationInfo         `json:"migration,omitempty"`
}

// ModelInfo describes supported model capabilities
type ModelInfo struct {
    ID              string                 `json:"id"`
    Name            string                 `json:"name,omitempty"`
    Description     string                 `json:"description,omitempty"`
    MaxTokens       int                    `json:"max_tokens"`
    ContextWindow   int                    `json:"context_window"`
    InputCost       float64               `json:"input_cost_per_token"`
    OutputCost      float64               `json:"output_cost_per_token"`
    Capabilities    []ModelCapability     `json:"capabilities"`
    Languages       []string              `json:"languages,omitempty"`
    Domains         []string              `json:"domains,omitempty"`
    TrainingCutoff  *time.Time            `json:"training_cutoff,omitempty"`
    ReleaseDate     *time.Time            `json:"release_date,omitempty"`
    Deprecated      bool                  `json:"deprecated"`
    Replacement     string                `json:"replacement,omitempty"`
}
```

### Capability Definitions

```go
// Capability represents a provider or model capability
type Capability string

const (
    // Text capabilities
    CapabilityTextGeneration    Capability = "text_generation"
    CapabilityTextCompletion    Capability = "text_completion"
    CapabilityTextSummarization Capability = "text_summarization"
    CapabilityTextTranslation   Capability = "text_translation"
    CapabilityTextClassification Capability = "text_classification"
    
    // Conversation capabilities
    CapabilityChatCompletion    Capability = "chat_completion"
    CapabilityConversation      Capability = "conversation"
    CapabilityContextMemory     Capability = "context_memory"
    
    // Function calling
    CapabilityFunctionCalling   Capability = "function_calling"
    CapabilityToolUse           Capability = "tool_use"
    CapabilityAgentWorkflows    Capability = "agent_workflows"
    
    // Media capabilities
    CapabilityVision           Capability = "vision"
    CapabilityImageGeneration  Capability = "image_generation"
    CapabilityAudioProcessing  Capability = "audio_processing"
    CapabilitySpeechToText     Capability = "speech_to_text"
    CapabilityTextToSpeech     Capability = "text_to_speech"
    
    // Streaming and real-time
    CapabilityStreaming        Capability = "streaming"
    CapabilityServerSentEvents Capability = "server_sent_events"
    CapabilityWebSockets       Capability = "websockets"
    
    // Code capabilities
    CapabilityCodeGeneration   Capability = "code_generation"
    CapabilityCodeCompletion   Capability = "code_completion"
    CapabilityCodeAnalysis     Capability = "code_analysis"
    CapabilityCodeExecution    Capability = "code_execution"
    
    // Advanced features
    CapabilityEmbeddings       Capability = "embeddings"
    CapabilityFineTuning       Capability = "fine_tuning"
    CapabilityModeration       Capability = "moderation"
    CapabilityLogitBias        Capability = "logit_bias"
    CapabilityTokenization     Capability = "tokenization"
    
    // Infrastructure
    CapabilityBatching         Capability = "batching"
    CapabilityCaching          Capability = "caching"
    CapabilityRatelimiting     Capability = "rate_limiting"
    CapabilityRetryLogic       Capability = "retry_logic"
    CapabilityHealthChecks     Capability = "health_checks"
    
    // Deployment
    CapabilityCloudHosted      Capability = "cloud_hosted"
    CapabilitySelfHosted       Capability = "self_hosted"
    CapabilityOnPremise        Capability = "on_premise"
    CapabilityEdgeDeployment   Capability = "edge_deployment"
)

// ModelCapability represents model-specific capabilities
type ModelCapability string

const (
    ModelCapabilityTextOnly          ModelCapability = "text_only"
    ModelCapabilityMultimodal        ModelCapability = "multimodal"
    ModelCapabilityVisionInput       ModelCapability = "vision_input"
    ModelCapabilityImageOutput       ModelCapability = "image_output"
    ModelCapabilityAudioInput        ModelCapability = "audio_input"
    ModelCapabilityAudioOutput       ModelCapability = "audio_output"
    ModelCapabilityFunctionCalling   ModelCapability = "function_calling"
    ModelCapabilityJSONMode          ModelCapability = "json_mode"
    ModelCapabilityStructuredOutput  ModelCapability = "structured_output"
    ModelCapabilitySystemPrompt      ModelCapability = "system_prompt"
    ModelCapabilityTemperatureControl ModelCapability = "temperature_control"
    ModelCapabilityTopPControl       ModelCapability = "top_p_control"
    ModelCapabilityStopSequences     ModelCapability = "stop_sequences"
    ModelCapabilityMaxTokensControl  ModelCapability = "max_tokens_control"
)

// CapabilityInfo provides detailed capability information
type CapabilityInfo struct {
    Name         Capability             `json:"name"`
    Description  string                 `json:"description"`
    Version      string                 `json:"version,omitempty"`
    Status       CapabilityStatus       `json:"status"`
    Requirements []string               `json:"requirements,omitempty"`
    Limitations  []string               `json:"limitations,omitempty"`
    Examples     []CapabilityExample    `json:"examples,omitempty"`
}

type CapabilityStatus string

const (
    CapabilityStatusStable      CapabilityStatus = "stable"
    CapabilityStatusBeta        CapabilityStatus = "beta"
    CapabilityStatusAlpha       CapabilityStatus = "alpha"
    CapabilityStatusDeprecated  CapabilityStatus = "deprecated"
    CapabilityStatusExperimental CapabilityStatus = "experimental"
)
```

---

## Configuration Schema System

### Schema Definition

```go
// ConfigSchema defines provider configuration requirements
type ConfigSchema struct {
    Schema      *jsonschema.Schema      `json:"schema"`
    Validation  ValidationRules         `json:"validation"`
    Transforms  []ConfigTransform       `json:"transforms,omitempty"`
    Secrets     []SecretField           `json:"secrets,omitempty"`
    Environment []EnvironmentVariable   `json:"environment,omitempty"`
}

// ValidationRules define custom validation logic
type ValidationRules struct {
    Required    []string                `json:"required"`
    Conditional []ConditionalRule       `json:"conditional,omitempty"`
    Cross       []CrossFieldValidation  `json:"cross_field,omitempty"`
    Custom      []CustomValidator       `json:"custom,omitempty"`
}

// ConditionalRule defines conditional requirements
type ConditionalRule struct {
    If      map[string]interface{} `json:"if"`
    Then    []string               `json:"then"`
    Else    []string               `json:"else,omitempty"`
    Message string                 `json:"message,omitempty"`
}

// Example: OpenAI provider schema
func OpenAIConfigSchema() *ConfigSchema {
    return &ConfigSchema{
        Schema: &jsonschema.Schema{
            Type: "object",
            Properties: map[string]*jsonschema.Schema{
                "api_key": {
                    Type:        "string",
                    Description: "OpenAI API key",
                    Pattern:     "^sk-[a-zA-Z0-9]{48}$",
                    MinLength:   51,
                    MaxLength:   51,
                },
                "organization": {
                    Type:        "string",
                    Description: "OpenAI organization ID",
                    Pattern:     "^org-[a-zA-Z0-9]+$",
                },
                "base_url": {
                    Type:        "string",
                    Description: "API base URL",
                    Format:      "uri",
                    Default:     "https://api.openai.com/v1",
                },
                "timeout": {
                    Type:        "integer",
                    Description: "Request timeout in seconds",
                    Minimum:     1,
                    Maximum:     300,
                    Default:     30,
                },
                "max_retries": {
                    Type:        "integer",
                    Description: "Maximum number of retries",
                    Minimum:     0,
                    Maximum:     10,
                    Default:     3,
                },
                "model_preferences": {
                    Type: "object",
                    Properties: map[string]*jsonschema.Schema{
                        "default_model": {
                            Type:        "string",
                            Description: "Default model to use",
                            Enum:        []interface{}{"gpt-4o", "gpt-4o-mini", "gpt-3.5-turbo"},
                            Default:     "gpt-4o-mini",
                        },
                        "fallback_models": {
                            Type: "array",
                            Items: &jsonschema.Schema{
                                Type: "string",
                            },
                            Description: "Fallback models in order of preference",
                        },
                    },
                },
                "features": {
                    Type: "object",
                    Properties: map[string]*jsonschema.Schema{
                        "streaming": {
                            Type:        "boolean",
                            Description: "Enable streaming responses",
                            Default:     true,
                        },
                        "function_calling": {
                            Type:        "boolean",
                            Description: "Enable function calling",
                            Default:     true,
                        },
                        "vision": {
                            Type:        "boolean",
                            Description: "Enable vision capabilities",
                            Default:     false,
                        },
                    },
                },
            },
            Required: []string{"api_key"},
        },
        Validation: ValidationRules{
            Required: []string{"api_key"},
            Conditional: []ConditionalRule{
                {
                    If:      map[string]interface{}{"features.vision": true},
                    Then:    []string{"model_preferences.default_model"},
                    Message: "Vision requires a vision-capable model",
                },
            },
        },
        Secrets: []SecretField{
            {
                Field:       "api_key",
                Environment: "OPENAI_API_KEY",
                Required:    true,
                Masked:      true,
            },
        },
        Environment: []EnvironmentVariable{
            {
                Name:        "OPENAI_API_KEY",
                Description: "OpenAI API key",
                Required:    true,
                Sensitive:   true,
            },
            {
                Name:        "OPENAI_ORGANIZATION",
                Description: "OpenAI organization ID",
                Required:    false,
                Sensitive:   false,
            },
        },
    }
}
```

### Configuration Validation

```go
// ConfigValidator validates provider configurations
type ConfigValidator struct {
    schemas map[string]*ConfigSchema
    cache   *ValidationCache
}

func NewConfigValidator() *ConfigValidator {
    return &ConfigValidator{
        schemas: make(map[string]*ConfigSchema),
        cache:   NewValidationCache(time.Hour),
    }
}

// Validate validates configuration against provider schema
func (v *ConfigValidator) Validate(provider string, config map[string]interface{}) (*ValidationResult, error) {
    // Check cache first
    cacheKey := fmt.Sprintf("%s:%s", provider, hashConfig(config))
    if result, ok := v.cache.Get(cacheKey); ok {
        return result.(*ValidationResult), nil
    }
    
    schema, ok := v.schemas[provider]
    if !ok {
        return nil, fmt.Errorf("no schema found for provider: %s", provider)
    }
    
    result := &ValidationResult{
        Valid:    true,
        Errors:   []ValidationError{},
        Warnings: []ValidationWarning{},
        Metadata: make(map[string]interface{}),
    }
    
    // JSON Schema validation
    if err := v.validateJSONSchema(config, schema.Schema, result); err != nil {
        return nil, err
    }
    
    // Custom validation rules
    if err := v.validateCustomRules(config, schema.Validation, result); err != nil {
        return nil, err
    }
    
    // Secret field validation
    if err := v.validateSecrets(config, schema.Secrets, result); err != nil {
        return nil, err
    }
    
    // Environment variable validation
    if err := v.validateEnvironment(schema.Environment, result); err != nil {
        return nil, err
    }
    
    // Apply transformations
    if len(schema.Transforms) > 0 {
        transformed, err := v.applyTransforms(config, schema.Transforms)
        if err != nil {
            result.AddWarning("transformation_failed", err.Error())
        } else {
            result.Metadata["transformed_config"] = transformed
        }
    }
    
    // Cache result
    v.cache.Set(cacheKey, result, time.Hour)
    
    return result, nil
}

// ValidationResult contains validation outcome
type ValidationResult struct {
    Valid       bool                    `json:"valid"`
    Errors      []ValidationError       `json:"errors,omitempty"`
    Warnings    []ValidationWarning     `json:"warnings,omitempty"`
    Metadata    map[string]interface{}  `json:"metadata,omitempty"`
}

type ValidationError struct {
    Field   string `json:"field"`
    Code    string `json:"code"`
    Message string `json:"message"`
    Value   interface{} `json:"value,omitempty"`
}

type ValidationWarning struct {
    Field   string `json:"field"`
    Code    string `json:"code"`
    Message string `json:"message"`
    Value   interface{} `json:"value,omitempty"`
}

// Custom validation implementations
func (v *ConfigValidator) validateCustomRules(config map[string]interface{}, rules ValidationRules, result *ValidationResult) error {
    // Conditional validation
    for _, rule := range rules.Conditional {
        if v.evaluateCondition(config, rule.If) {
            for _, field := range rule.Then {
                if !v.hasField(config, field) {
                    result.AddError(field, "conditional_required", rule.Message)
                }
            }
        } else if len(rule.Else) > 0 {
            for _, field := range rule.Else {
                if !v.hasField(config, field) {
                    result.AddError(field, "conditional_required", rule.Message)
                }
            }
        }
    }
    
    // Cross-field validation
    for _, crossRule := range rules.Cross {
        if err := v.validateCrossField(config, crossRule, result); err != nil {
            return err
        }
    }
    
    return nil
}

func (v *ConfigValidator) validateSecrets(config map[string]interface{}, secrets []SecretField, result *ValidationResult) error {
    for _, secret := range secrets {
        value := v.getFieldValue(config, secret.Field)
        
        if secret.Required && value == nil {
            // Check environment variable
            if envValue := os.Getenv(secret.Environment); envValue == "" {
                result.AddError(secret.Field, "secret_required", 
                    fmt.Sprintf("Required secret field %s not provided via config or environment variable %s", 
                        secret.Field, secret.Environment))
            }
        }
        
        if value != nil && secret.Masked {
            // Mask secret in logs
            result.Metadata[secret.Field+"_masked"] = v.maskSecret(fmt.Sprintf("%v", value))
        }
    }
    
    return nil
}
```

---

## Performance and Reliability Metrics

### Performance Characteristics

```go
// PerformanceMetrics define provider performance characteristics
type PerformanceMetrics struct {
    Latency          LatencyMetrics      `json:"latency"`
    Throughput       ThroughputMetrics   `json:"throughput"`
    Concurrency      ConcurrencyMetrics  `json:"concurrency"`
    ResourceUsage    ResourceMetrics     `json:"resource_usage"`
    Scaling          ScalingMetrics      `json:"scaling"`
}

type LatencyMetrics struct {
    Typical          time.Duration `json:"typical"`           // P50
    Best             time.Duration `json:"best"`              // P10
    Worst            time.Duration `json:"worst"`             // P99
    TimeToFirstToken time.Duration `json:"time_to_first_token,omitempty"`
    StreamingLatency time.Duration `json:"streaming_latency,omitempty"`
}

type ThroughputMetrics struct {
    RequestsPerSecond     float64 `json:"requests_per_second"`
    TokensPerSecond       float64 `json:"tokens_per_second"`
    MaxConcurrentRequests int     `json:"max_concurrent_requests"`
    BatchSize             int     `json:"batch_size,omitempty"`
}

type ConcurrencyMetrics struct {
    MaxConnections       int           `json:"max_connections"`
    ConnectionPoolSize   int           `json:"connection_pool_size"`
    KeepAliveTimeout     time.Duration `json:"keep_alive_timeout"`
    ConnectionTimeout    time.Duration `json:"connection_timeout"`
}

type ResourceMetrics struct {
    CPUUsage      float64 `json:"cpu_usage_percent"`
    MemoryUsage   int64   `json:"memory_usage_mb"`
    NetworkBandwidth int64 `json:"network_bandwidth_mbps"`
    DiskIO        int64   `json:"disk_io_mbps,omitempty"`
}

type ScalingMetrics struct {
    AutoScaling        bool          `json:"auto_scaling"`
    ScaleUpTime        time.Duration `json:"scale_up_time,omitempty"`
    ScaleDownTime      time.Duration `json:"scale_down_time,omitempty"`
    MinInstances       int           `json:"min_instances"`
    MaxInstances       int           `json:"max_instances"`
    ScalingTriggers    []string      `json:"scaling_triggers,omitempty"`
}

// Example: OpenAI performance metrics
func OpenAIPerformanceMetrics() PerformanceMetrics {
    return PerformanceMetrics{
        Latency: LatencyMetrics{
            Typical:          800 * time.Millisecond,
            Best:             200 * time.Millisecond,
            Worst:            5 * time.Second,
            TimeToFirstToken: 300 * time.Millisecond,
            StreamingLatency: 50 * time.Millisecond,
        },
        Throughput: ThroughputMetrics{
            RequestsPerSecond:     10,
            TokensPerSecond:       50,
            MaxConcurrentRequests: 100,
        },
        Concurrency: ConcurrencyMetrics{
            MaxConnections:     50,
            ConnectionPoolSize: 10,
            KeepAliveTimeout:   30 * time.Second,
            ConnectionTimeout:  10 * time.Second,
        },
        ResourceUsage: ResourceMetrics{
            CPUUsage:         5.0,
            MemoryUsage:      50,
            NetworkBandwidth: 10,
        },
        Scaling: ScalingMetrics{
            AutoScaling:     true,
            ScaleUpTime:     30 * time.Second,
            ScaleDownTime:   60 * time.Second,
            MinInstances:    1,
            MaxInstances:    10,
            ScalingTriggers: []string{"request_rate", "error_rate"},
        },
    }
}
```

### Reliability Metrics

```go
// ReliabilityMetrics define provider reliability characteristics
type ReliabilityMetrics struct {
    Availability      AvailabilityMetrics   `json:"availability"`
    ErrorRates        ErrorRateMetrics      `json:"error_rates"`
    Recovery          RecoveryMetrics       `json:"recovery"`
    Monitoring        MonitoringMetrics     `json:"monitoring"`
    SLA               SLAMetrics            `json:"sla"`
}

type AvailabilityMetrics struct {
    Uptime                float64       `json:"uptime_percent"`
    MTBF                  time.Duration `json:"mtbf"` // Mean Time Between Failures
    MTTR                  time.Duration `json:"mttr"` // Mean Time To Recovery
    PlannedDowntime       time.Duration `json:"planned_downtime_per_month"`
    UnplannedDowntime     time.Duration `json:"unplanned_downtime_per_month"`
    AvailabilityZones     []string      `json:"availability_zones,omitempty"`
}

type ErrorRateMetrics struct {
    OverallErrorRate      float64            `json:"overall_error_rate"`
    ErrorRatesByType      map[string]float64 `json:"error_rates_by_type"`
    TransientErrorRate    float64            `json:"transient_error_rate"`
    PermanentErrorRate    float64            `json:"permanent_error_rate"`
    TimeoutRate           float64            `json:"timeout_rate"`
    RateLimitErrorRate    float64            `json:"rate_limit_error_rate"`
}

type RecoveryMetrics struct {
    AutoRetry            bool          `json:"auto_retry"`
    MaxRetries           int           `json:"max_retries"`
    RetryBackoff         string        `json:"retry_backoff"` // "linear", "exponential", "custom"
    CircuitBreaker       bool          `json:"circuit_breaker"`
    FailoverTime         time.Duration `json:"failover_time,omitempty"`
    RecoveryTime         time.Duration `json:"recovery_time"`
}

type MonitoringMetrics struct {
    HealthChecks         bool          `json:"health_checks"`
    HealthCheckInterval  time.Duration `json:"health_check_interval"`
    Alerting             bool          `json:"alerting"`
    Logging              bool          `json:"logging"`
    Metrics              bool          `json:"metrics"`
    Tracing              bool          `json:"tracing"`
}

type SLAMetrics struct {
    ResponseTime         time.Duration `json:"response_time_sla"`
    Availability         float64       `json:"availability_sla"`
    Throughput           float64       `json:"throughput_sla"`
    ErrorRate            float64       `json:"error_rate_sla"`
    Support              SupportLevel  `json:"support_level"`
}

type SupportLevel string

const (
    SupportLevelCommunity  SupportLevel = "community"
    SupportLevelBasic      SupportLevel = "basic"
    SupportLevelStandard   SupportLevel = "standard"
    SupportLevelPremium    SupportLevel = "premium"
    SupportLevelEnterprise SupportLevel = "enterprise"
)
```

---

## Cost Metrics

### Cost Structure

```go
// CostMetrics define provider cost characteristics
type CostMetrics struct {
    Pricing          PricingModel    `json:"pricing"`
    TokenCosts       TokenCostInfo   `json:"token_costs"`
    Usage            UsageMetrics    `json:"usage"`
    Billing          BillingInfo     `json:"billing"`
    Optimization     CostOptimization `json:"optimization"`
}

type PricingModel struct {
    Type             string               `json:"type"` // "pay_per_use", "subscription", "hybrid"
    Currency         string               `json:"currency"`
    FreeTier         *FreeTierInfo        `json:"free_tier,omitempty"`
    Subscriptions    []SubscriptionTier   `json:"subscriptions,omitempty"`
    UsageBased       *UsageBasedPricing   `json:"usage_based,omitempty"`
}

type TokenCostInfo struct {
    InputTokens      CostPerToken  `json:"input_tokens"`
    OutputTokens     CostPerToken  `json:"output_tokens"`
    VisionTokens     *CostPerToken `json:"vision_tokens,omitempty"`
    AudioTokens      *CostPerToken `json:"audio_tokens,omitempty"`
    FunctionCalls    *CostPerCall  `json:"function_calls,omitempty"`
}

type CostPerToken struct {
    Cost        float64               `json:"cost"`
    Unit        string                `json:"unit"` // "per_1k_tokens", "per_token"
    ModelTiers  map[string]float64    `json:"model_tiers,omitempty"`
    VolumeTiers []VolumePricingTier   `json:"volume_tiers,omitempty"`
}

type VolumePricingTier struct {
    MinTokens   int64   `json:"min_tokens"`
    MaxTokens   int64   `json:"max_tokens,omitempty"`
    Cost        float64 `json:"cost"`
    Discount    float64 `json:"discount_percent,omitempty"`
}

type UsageMetrics struct {
    EstimatedMonthlyCost  float64              `json:"estimated_monthly_cost"`
    CostBreakdown         map[string]float64   `json:"cost_breakdown"`
    PeakUsageCost         float64              `json:"peak_usage_cost"`
    AverageRequestCost    float64              `json:"average_request_cost"`
    CostEfficiency        CostEfficiencyMetrics `json:"cost_efficiency"`
}

type CostEfficiencyMetrics struct {
    CostPerRequest       float64 `json:"cost_per_request"`
    CostPerToken         float64 `json:"cost_per_token"`
    CostPerSecond        float64 `json:"cost_per_second"`
    UtilizationRate      float64 `json:"utilization_rate"`
    WastePercentage      float64 `json:"waste_percentage"`
}

// Example: OpenAI cost metrics
func OpenAICostMetrics() CostMetrics {
    return CostMetrics{
        Pricing: PricingModel{
            Type:     "pay_per_use",
            Currency: "USD",
            UsageBased: &UsageBasedPricing{
                MeteredBy: "tokens",
                BillingCycle: "monthly",
            },
        },
        TokenCosts: TokenCostInfo{
            InputTokens: CostPerToken{
                Cost: 0.0015,
                Unit: "per_1k_tokens",
                ModelTiers: map[string]float64{
                    "gpt-4o":      0.005,
                    "gpt-4o-mini": 0.00015,
                    "gpt-3.5-turbo": 0.0015,
                },
                VolumeTiers: []VolumePricingTier{
                    {MinTokens: 0, MaxTokens: 1000000, Cost: 0.0015},
                    {MinTokens: 1000000, MaxTokens: 10000000, Cost: 0.0012},
                    {MinTokens: 10000000, Cost: 0.001},
                },
            },
            OutputTokens: CostPerToken{
                Cost: 0.002,
                Unit: "per_1k_tokens",
                ModelTiers: map[string]float64{
                    "gpt-4o":      0.015,
                    "gpt-4o-mini": 0.0006,
                    "gpt-3.5-turbo": 0.002,
                },
            },
        },
        Usage: UsageMetrics{
            EstimatedMonthlyCost: 50.0,
            CostBreakdown: map[string]float64{
                "input_tokens":  15.0,
                "output_tokens": 30.0,
                "api_calls":     5.0,
            },
            AverageRequestCost: 0.05,
            CostEfficiency: CostEfficiencyMetrics{
                CostPerRequest:  0.05,
                CostPerToken:    0.000025,
                UtilizationRate: 0.75,
                WastePercentage: 10.0,
            },
        },
    }
}
```

---

## Dynamic Metadata Discovery

### Metadata Discovery Service

```go
// MetadataDiscovery discovers and updates provider metadata
type MetadataDiscovery struct {
    sources    []MetadataSource
    cache      *MetadataCache
    updater    *MetadataUpdater
    validator  *MetadataValidator
}

type MetadataSource interface {
    Discover(ctx context.Context) ([]ProviderMetadata, error)
    SupportsRealtime() bool
    GetSourceInfo() SourceInfo
}

// HTTP-based metadata discovery
type HTTPMetadataSource struct {
    baseURL    string
    client     *http.Client
    headers    map[string]string
    transformer MetadataTransformer
}

func (s *HTTPMetadataSource) Discover(ctx context.Context) ([]ProviderMetadata, error) {
    // Discover providers from HTTP endpoint
    resp, err := s.client.Get(s.baseURL + "/providers")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var raw []map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
        return nil, err
    }
    
    var metadata []ProviderMetadata
    for _, item := range raw {
        meta, err := s.transformer.Transform(item)
        if err != nil {
            log.Printf("Failed to transform metadata: %v", err)
            continue
        }
        metadata = append(metadata, meta)
    }
    
    return metadata, nil
}

// OpenAPI-based discovery
type OpenAPIMetadataSource struct {
    specURL    string
    client     *http.Client
    extractor  OpenAPIExtractor
}

func (s *OpenAPIMetadataSource) Discover(ctx context.Context) ([]ProviderMetadata, error) {
    // Fetch OpenAPI spec
    resp, err := s.client.Get(s.specURL)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var spec openapi3.T
    if err := spec.ReadFromResponse(resp); err != nil {
        return nil, err
    }
    
    // Extract provider metadata from OpenAPI spec
    return s.extractor.ExtractMetadata(&spec)
}

// Real-time metadata updates
type RealTimeMetadataSource struct {
    websocketURL string
    conn         *websocket.Conn
    updates      chan ProviderMetadata
}

func (s *RealTimeMetadataSource) StartUpdates(ctx context.Context) error {
    conn, _, err := websocket.DefaultDialer.Dial(s.websocketURL, nil)
    if err != nil {
        return err
    }
    s.conn = conn
    
    go s.listenForUpdates(ctx)
    return nil
}

func (s *RealTimeMetadataSource) listenForUpdates(ctx context.Context) {
    defer s.conn.Close()
    
    for {
        select {
        case <-ctx.Done():
            return
        default:
            var update ProviderMetadata
            if err := s.conn.ReadJSON(&update); err != nil {
                log.Printf("Error reading metadata update: %v", err)
                return
            }
            
            select {
            case s.updates <- update:
            case <-ctx.Done():
                return
            }
        }
    }
}
```

### Metadata Validation and Updates

```go
// MetadataValidator validates discovered metadata
type MetadataValidator struct {
    schemas    map[string]*jsonschema.Schema
    rules      []ValidationRule
    sanitizer  MetadataSanitizer
}

func (v *MetadataValidator) Validate(metadata ProviderMetadata) (*ValidationResult, error) {
    result := &ValidationResult{Valid: true}
    
    // Schema validation
    if err := v.validateSchema(metadata, result); err != nil {
        return nil, err
    }
    
    // Business rules validation
    if err := v.validateRules(metadata, result); err != nil {
        return nil, err
    }
    
    // Security validation
    if err := v.validateSecurity(metadata, result); err != nil {
        return nil, err
    }
    
    // Sanitize if needed
    if !result.Valid {
        sanitized, err := v.sanitizer.Sanitize(metadata)
        if err != nil {
            return result, err
        }
        result.Metadata["sanitized"] = sanitized
    }
    
    return result, nil
}

// MetadataUpdater manages metadata updates
type MetadataUpdater struct {
    store      MetadataStore
    notifier   UpdateNotifier
    scheduler  *UpdateScheduler
    versioning VersionManager
}

func (u *MetadataUpdater) UpdateMetadata(metadata ProviderMetadata) error {
    // Version the update
    versioned := u.versioning.CreateVersion(metadata)
    
    // Store update
    if err := u.store.Store(versioned); err != nil {
        return err
    }
    
    // Notify subscribers
    u.notifier.NotifyUpdate(UpdateEvent{
        Type:     "metadata_update",
        Provider: metadata.Name,
        Version:  versioned.Version,
        Changes:  u.detectChanges(metadata),
}
    
    return nil
}

// Metadata change detection
func (u *MetadataUpdater) detectChanges(new ProviderMetadata) []Change {
    existing, err := u.store.Get(new.Name)
    if err != nil {
        return []Change{{Type: "created", Field: "*"}}
    }
    
    var changes []Change
    
    // Capability changes
    if !slicesEqual(existing.Capabilities, new.Capabilities) {
        changes = append(changes, Change{
            Type:     "capability_change",
            Field:    "capabilities",
            OldValue: existing.Capabilities,
            NewValue: new.Capabilities,
}
    }
    
    // Model changes
    if !modelsEqual(existing.SupportedModels, new.SupportedModels) {
        changes = append(changes, Change{
            Type:     "model_change",
            Field:    "supported_models",
            OldValue: existing.SupportedModels,
            NewValue: new.SupportedModels,
}
    }
    
    // Cost changes
    if !costsEqual(existing.Cost, new.Cost) {
        changes = append(changes, Change{
            Type:     "cost_change",
            Field:    "cost",
            OldValue: existing.Cost,
            NewValue: new.Cost,
}
    }
    
    return changes
}

type Change struct {
    Type     string      `json:"type"`
    Field    string      `json:"field"`
    OldValue interface{} `json:"old_value,omitempty"`
    NewValue interface{} `json:"new_value,omitempty"`
}
```

---

## Metadata Query and Search

### Query Interface

```go
// MetadataQuery provides advanced metadata querying
type MetadataQuery struct {
    store   MetadataStore
    index   SearchIndex
    filters []QueryFilter
}

// Query DSL for metadata search
type Query struct {
    Filters    []Filter         `json:"filters,omitempty"`
    Sort       []SortCriteria   `json:"sort,omitempty"`
    Pagination *Pagination      `json:"pagination,omitempty"`
    Facets     []string         `json:"facets,omitempty"`
}

type Filter struct {
    Field    string      `json:"field"`
    Operator string      `json:"operator"` // "eq", "ne", "in", "nin", "gt", "lt", "contains", "regex"
    Value    interface{} `json:"value"`
    Logic    string      `json:"logic,omitempty"` // "and", "or", "not"
}

// Example queries
func ExampleQueries() {
    // Find providers with streaming capability
    streamingQuery := Query{
        Filters: []Filter{
            {
                Field:    "capabilities",
                Operator: "contains",
                Value:    "streaming",
            },
        },
    }
    
    // Find cost-effective providers
    costEffectiveQuery := Query{
        Filters: []Filter{
            {
                Field:    "cost.token_costs.input_tokens.cost",
                Operator: "lt",
                Value:    0.001,
            },
            {
                Field:    "reliability.availability.uptime",
                Operator: "gt",
                Value:    99.5,
                Logic:    "and",
            },
        },
        Sort: []SortCriteria{
            {Field: "cost.token_costs.input_tokens.cost", Direction: "asc"},
        },
    }
    
    // Find vision-capable models
    visionQuery := Query{
        Filters: []Filter{
            {
                Field:    "supported_models.capabilities",
                Operator: "contains",
                Value:    "vision_input",
            },
        },
        Facets: []string{"supported_models.id", "capabilities"},
    }
}

// Advanced search implementation
func (mq *MetadataQuery) Search(query Query) (*SearchResult, error) {
    // Build search criteria
    criteria := mq.buildCriteria(query)
    
    // Execute search with index
    results, err := mq.index.Search(criteria)
    if err != nil {
        return nil, err
    }
    
    // Apply post-processing
    processed := mq.postProcess(results, query)
    
    // Generate facets
    facets := make(map[string][]FacetValue)
    for _, facetField := range query.Facets {
        facets[facetField] = mq.generateFacets(results, facetField)
    }
    
    return &SearchResult{
        Results:     processed,
        Total:       len(results),
        Facets:      facets,
        Pagination:  query.Pagination,
        ExecutionTime: time.Since(start),
    }, nil
}

type SearchResult struct {
    Results       []ProviderMetadata    `json:"results"`
    Total         int                   `json:"total"`
    Facets        map[string][]FacetValue `json:"facets,omitempty"`
    Pagination    *Pagination           `json:"pagination,omitempty"`
    ExecutionTime time.Duration         `json:"execution_time"`
}

type FacetValue struct {
    Value string `json:"value"`
    Count int    `json:"count"`
}
```

---

## Integration Examples

### Provider Registration with Metadata

```go
// Register provider with comprehensive metadata
func RegisterProviderWithMetadata() {
    registry := provider.GetGlobalRegistry()
    
    // Create provider metadata
    metadata := ProviderMetadata{
        Name:        "openai",
        DisplayName: "OpenAI",
        Description: "OpenAI GPT models provider with advanced capabilities",
        Version:     "1.0.0",
        Author:      "go-llms",
        License:     "MIT",
        Homepage:    "https://openai.com",
        Documentation: "https://platform.openai.com/docs",
        
        Capabilities: []Capability{
            CapabilityTextGeneration,
            CapabilityFunctionCalling,
            CapabilityVision,
            CapabilityStreaming,
            CapabilityJSONMode,
        },
        
        SupportedModels: []ModelInfo{
            {
                ID:           "gpt-4o",
                Name:         "GPT-4 Omni",
                Description:  "Most advanced multimodal model",
                MaxTokens:    128000,
                ContextWindow: 128000,
                InputCost:    0.005,
                OutputCost:   0.015,
                Capabilities: []ModelCapability{
                    ModelCapabilityMultimodal,
                    ModelCapabilityVisionInput,
                    ModelCapabilityFunctionCalling,
                    ModelCapabilityJSONMode,
                },
                Languages:    []string{"en", "es", "fr", "de", "zh", "ja"},
                ReleaseDate:  &time.Time{}, // Set actual date
            },
        },
        
        ConfigSchema:   OpenAIConfigSchema().Schema,
        DefaultConfig: map[string]interface{}{
            "timeout":     30,
            "max_retries": 3,
            "base_url":    "https://api.openai.com/v1",
        },
        
        Performance: OpenAIPerformanceMetrics(),
        Reliability: OpenAIReliabilityMetrics(),
        Cost:        OpenAICostMetrics(),
        
        Tags:       []string{"commercial", "cloud", "popular"},
        Categories: []string{"llm", "multimodal", "enterprise"},
        
        Status: ProviderStatusStable,
    }
    
    // Register with factory function
    factory := func(config map[string]interface{}) (provider.Provider, error) {
provider := provider.NewOpenAIProvider(return provider.NewOpenAIProvider(os.Getenv("OPENAI_API_KEY"), "gpt-4",
    domain.NewOpenAIOrganizationOption(getStringValue(config, "organization")),
    domain.NewBaseURLOption(getStringValue(config, "base_url")),
    domain.NewTimeoutOption(time.Duration(getIntValue(config, "timeout")) * time.Second),
)
```

---

## Best Practices

### 1. Metadata Completeness
- Include comprehensive capability information
- Provide accurate performance and cost metrics
- Document configuration schemas thoroughly
- Keep model information up to date

### 2. Schema Design
- Use standard JSON Schema for validation
- Include helpful descriptions and examples
- Implement proper error messages
- Support conditional validation

### 3. Performance Monitoring
- Track actual vs. expected performance
- Update metrics based on real usage
- Monitor for performance degradation
- Implement alerting for SLA violations

### 4. Cost Management
- Track actual costs vs. estimates
- Implement cost optimization recommendations
- Monitor for cost anomalies
- Provide cost projections

### 5. Metadata Evolution
- Version metadata changes
- Provide migration paths
- Maintain backward compatibility
- Document breaking changes

---

## Next Steps

- **[Provider Registry](provider-registry.md)** - Dynamic registration and discovery
- **[Agent Architecture](../../technical/agents/overview.md)** - Agent system design
- **[Custom Providers](../../user-guide/advanced/custom-providers.md)** - Building custom providers
- **[Provider Setup Guide](../../user-guide/guides/provider-setup.md)** - Configuration guide
- **[API Reference](../../technical/api-reference/providers.md)** - Provider interface documentation