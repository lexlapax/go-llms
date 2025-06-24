# Existing Systems: Adding LLM Capabilities

> **[Project Root](/) / [Documentation](/docs/) / [User Guide](/docs/user-guide/) / [Guides](/docs/user-guide/guides/) / Existing Systems**

Transform legacy applications and existing systems by adding AI capabilities. Master gradual integration strategies, wrapper patterns, and bridge architectures that enhance existing functionality without disrupting operations.

## Why Legacy Integration Matters

- **Preservation** - Maintain existing investments and avoid costly rewrites
- **Enhancement** - Add intelligence to proven business processes
- **Risk Mitigation** - Introduce AI capabilities gradually with minimal risk
- **Value Creation** - Unlock new capabilities from existing data and workflows
- **Cost Efficiency** - Extend system lifecycles with modern AI features

## Integration Architecture

![Legacy System Integration](../../images/legacy-integration-architecture.svg)

### Integration Layers
1. **Wrapper Layer** - API and service wrapping without modification
2. **Bridge Layer** - Event-driven communication between systems
3. **Enhancement Layer** - AI-powered data processing and insights
4. **Monitoring Layer** - Observability and performance tracking
5. **Migration Layer** - Gradual transition to AI-enhanced workflows

### Integration Strategies
| Strategy | Risk Level | Effort | Time to Value | Best For |
|----------|------------|--------|---------------|----------|
| **API Wrapping** | Low | Low | Days | REST APIs, microservices |
| **Event Bridging** | Medium | Medium | Weeks | Event-driven systems |
| **Data Enhancement** | Low | Medium | Weeks | Reports, analytics |
| **Workflow Augmentation** | Medium | High | Months | Business processes |
| **Full Migration** | High | High | Quarters | System replacement |

## Prerequisites

- [APIs and Services understanding](apis-and-services.md) ✅
- [Agent Tools completed](agent-tools.md) ✅
- Understanding of existing system architecture ✅

---

## Level 1: API Wrapping and Basic Integration
*Add AI capabilities through non-invasive API wrappers*

### Universal API Wrapper
```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
    "github.com/lexlapax/go-llms/pkg/util/auth"
)

// LegacySystemWrapper wraps existing APIs with AI capabilities
type LegacySystemWrapper struct {
    baseURL     string
    authConfig  *auth.AuthConfig
    agent       *core.LLMAgent
    apiClient   *web.APIClientTool
}

// SystemConfig holds configuration for legacy system
type SystemConfig struct {
    BaseURL      string            `json:"base_url"`
    AuthType     string            `json:"auth_type"`     // api_key, bearer, basic
    Credentials  map[string]string `json:"credentials"`
    Endpoints    map[string]EndpointConfig `json:"endpoints"`
    RateLimit    int               `json:"rate_limit"`
    Timeout      time.Duration     `json:"timeout"`
}

type EndpointConfig struct {
    Path        string            `json:"path"`
    Method      string            `json:"method"`
    Description string            `json:"description"`
    Parameters  map[string]string `json:"parameters"`
    ResponseSchema map[string]interface{} `json:"response_schema"`
}

// AIEnhancedResponse wraps legacy responses with AI insights
type AIEnhancedResponse struct {
    OriginalData   interface{}            `json:"original_data"`
    AIInsights     map[string]interface{} `json:"ai_insights"`
    Summary        string                 `json:"summary"`
    Recommendations []string             `json:"recommendations"`
    Metadata       map[string]interface{} `json:"metadata"`
    ProcessedAt    time.Time              `json:"processed_at"`
}

func NewLegacySystemWrapper(config *SystemConfig) (*LegacySystemWrapper, error) {
    // Create LLM provider
    llm, err := provider.NewOpenAI(
        provider.WithModel("gpt-4"),
        provider.WithMaxTokens(2000),
    )
    if err != nil {
        return nil, err
    }

    // Create agent
    agent := core.NewLLMAgent("legacy-wrapper-agent", llm)

    // Configure authentication
    authConfig := &auth.AuthConfig{
        Type: config.AuthType,
        Credentials: config.Credentials,
    }

    // Create API client tool
    apiClient := web.NewAPIClientTool()

    wrapper := &LegacySystemWrapper{
        baseURL:    config.BaseURL,
        authConfig: authConfig,
        agent:      agent,
        apiClient:  apiClient,
    }

    // Add API client tool to agent
    agent.AddTool(apiClient)

    return wrapper, nil
}

func (lsw *LegacySystemWrapper) CallEndpointWithAI(ctx context.Context, endpoint string, params map[string]interface{}, enhancementType string) (*AIEnhancedResponse, error) {
    // 1. Call legacy endpoint
    originalResponse, err := lsw.callLegacyEndpoint(ctx, endpoint, params)
    if err != nil {
        return nil, fmt.Errorf("legacy endpoint call failed: %w", err)
    }

    // 2. Process with AI based on enhancement type
    aiInsights, err := lsw.enhanceWithAI(ctx, originalResponse, enhancementType)
    if err != nil {
        log.Printf("AI enhancement failed: %v", err)
        // Continue without AI insights rather than fail
        aiInsights = map[string]interface{}{
            "error": "AI enhancement unavailable",
        }
    }

    // 3. Create enhanced response
    response := &AIEnhancedResponse{
        OriginalData: originalResponse,
        AIInsights:   aiInsights,
        Metadata: map[string]interface{}{
            "endpoint": endpoint,
            "enhancement_type": enhancementType,
        },
        ProcessedAt: time.Now(),
    }

    return response, nil
}

func (lsw *LegacySystemWrapper) callLegacyEndpoint(ctx context.Context, endpoint string, params map[string]interface{}) (interface{}, error) {
    // Create state for API client
    state := domain.NewState()
    
    // Set authentication in state
    for key, value := range lsw.authConfig.Credentials {
        state.Set(key, value)
    }

    // Prepare API call parameters
    apiParams := map[string]interface{}{
        "url":    lsw.baseURL + "/" + endpoint,
        "method": "GET", // Default, can be configured per endpoint
        "params": params,
    }

    // Execute API call through agent
    result, err := lsw.agent.RunWithParams(ctx, state, map[string]interface{}{
        "tool": "api_client",
        "params": apiParams,
    })
    
    if err != nil {
        return nil, err
    }

    // Extract data from agent result
    if len(result.Messages) > 0 {
        lastMessage := result.Messages[len(result.Messages)-1].TextContent()
        
        var data interface{}
        if err := json.Unmarshal([]byte(lastMessage), &data); err != nil {
            // Return as text if not JSON
            return lastMessage, nil
        }
        return data, nil
    }

    return nil, fmt.Errorf("no response from legacy endpoint")
}

func (lsw *LegacySystemWrapper) enhanceWithAI(ctx context.Context, data interface{}, enhancementType string) (map[string]interface{}, error) {
    dataJSON, err := json.Marshal(data)
    if err != nil {
        return nil, err
    }

    var prompt string
    switch enhancementType {
    case "analyze":
        prompt = fmt.Sprintf(`Analyze the following data and provide insights:

Data: %s

Please provide:
1. Key findings and patterns
2. Summary of important information
3. Actionable recommendations
4. Potential issues or anomalies

Return your analysis as JSON with structured insights.`, string(dataJSON))

    case "summarize":
        prompt = fmt.Sprintf(`Summarize the following data in a clear, concise manner:

Data: %s

Provide a brief summary highlighting the most important information and key takeaways.`, string(dataJSON))

    case "validate":
        prompt = fmt.Sprintf(`Validate the following data for consistency and completeness:

Data: %s

Check for:
1. Data completeness
2. Logical consistency
3. Format compliance
4. Potential errors

Return validation results with specific findings.`, string(dataJSON))

    case "enrich":
        prompt = fmt.Sprintf(`Enrich the following data with additional context and insights:

Data: %s

Add relevant information such as:
1. Context and background
2. Related information
3. Industry standards or benchmarks
4. Suggestions for improvement`, string(dataJSON))

    default:
        prompt = fmt.Sprintf(`Process the following data and provide useful insights:

Data: %s

Analyze and provide relevant information that would be helpful for decision-making.`, string(dataJSON))
    }

    // Create state for AI processing
    state := domain.NewState()
    state.AddMessage(domain.NewTextMessage(domain.RoleUser, prompt))

    // Process with agent
    result, err := lsw.agent.Run(ctx, state)
    if err != nil {
        return nil, err
    }

    // Extract AI insights
    insights := make(map[string]interface{})
    if len(result.Messages) > 0 {
        aiResponse := result.Messages[len(result.Messages)-1].TextContent()
        
        // Try to parse as JSON
        if err := json.Unmarshal([]byte(aiResponse), &insights); err != nil {
            // Fallback to text format
            insights = map[string]interface{}{
                "analysis": aiResponse,
                "format": "text",
            }
        }
    }

    insights["tokens_used"] = result.TokensUsed
    insights["processing_time"] = result.ProcessingTime

    return insights, nil
}

// Batch processing for multiple legacy operations
func (lsw *LegacySystemWrapper) BatchProcessWithAI(ctx context.Context, requests []BatchRequest) ([]AIEnhancedResponse, error) {
    responses := make([]AIEnhancedResponse, len(requests))
    
    for i, req := range requests {
        response, err := lsw.CallEndpointWithAI(ctx, req.Endpoint, req.Params, req.EnhancementType)
        if err != nil {
            responses[i] = AIEnhancedResponse{
                Metadata: map[string]interface{}{
                    "error": err.Error(),
                    "endpoint": req.Endpoint,
                },
                ProcessedAt: time.Now(),
            }
            continue
        }
        responses[i] = *response
    }

    return responses, nil
}

type BatchRequest struct {
    Endpoint        string                 `json:"endpoint"`
    Params          map[string]interface{} `json:"params"`
    EnhancementType string                 `json:"enhancement_type"`
}

// HTTP API wrapper for external consumption
func (lsw *LegacySystemWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var request struct {
        Endpoint        string                 `json:"endpoint"`
        Params          map[string]interface{} `json:"params"`
        EnhancementType string                 `json:"enhancement_type"`
    }

    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
    defer cancel()

    response, err := lsw.CallEndpointWithAI(ctx, request.Endpoint, request.Params, request.EnhancementType)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func main() {
    // Configuration for legacy system
    config := &SystemConfig{
        BaseURL:  "https://legacy-api.company.com/api/v1",
        AuthType: "api_key",
        Credentials: map[string]string{
            "api_key": "legacy_api_key_123",
        },
        Endpoints: map[string]EndpointConfig{
            "customers": {
                Path:        "customers",
                Method:      "GET",
                Description: "Get customer data",
            },
            "orders": {
                Path:        "orders",
                Method:      "GET",
                Description: "Get order information",
            },
        },
        RateLimit: 100,
        Timeout:   30 * time.Second,
    }

    // Create wrapper
    wrapper, err := NewLegacySystemWrapper(config)
    if err != nil {
        log.Fatal("Failed to create wrapper:", err)
    }

    // Example: Enhance customer data with AI
    ctx := context.Background()
    
    customerParams := map[string]interface{}{
        "customer_id": "12345",
    }

    enhancedResponse, err := wrapper.CallEndpointWithAI(ctx, "customers", customerParams, "analyze")
    if err != nil {
        log.Fatal("Failed to enhance customer data:", err)
    }

    fmt.Printf("Original Data: %+v\n", enhancedResponse.OriginalData)
    fmt.Printf("AI Insights: %+v\n", enhancedResponse.AIInsights)

    // Serve as HTTP API
    http.Handle("/enhance", wrapper)
    log.Println("AI-enhanced legacy API wrapper running on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

---

## Level 2: Event-Driven Integration
*Bridge legacy systems with AI using event-driven architecture*

### Event Bridge for Legacy Systems
```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/agent/events"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
)

// LegacyEventBridge connects legacy systems with AI agents
type LegacyEventBridge struct {
    agent               *core.LLMAgent
    eventBus            *events.EventBus
    legacyConnectors    map[string]LegacyConnector
    processors          map[string]EventProcessor
    config              *BridgeConfig
    metrics             *BridgeMetrics
    mu                  sync.RWMutex
}

type BridgeConfig struct {
    BufferSize          int           `json:"buffer_size"`
    ProcessingTimeout   time.Duration `json:"processing_timeout"`
    RetryAttempts       int           `json:"retry_attempts"`
    EnableMetrics       bool          `json:"enable_metrics"`
    EnableDeadLetter    bool          `json:"enable_dead_letter"`
    DeadLetterPath      string        `json:"dead_letter_path"`
}

type BridgeMetrics struct {
    EventsProcessed     int64
    EventsSucceeded     int64
    EventsFailed        int64
    ProcessingTime      time.Duration
    LastProcessingTime  time.Time
}

// LegacyConnector interface for different legacy system types
type LegacyConnector interface {
    Connect(ctx context.Context) error
    Listen(ctx context.Context, eventChan chan<- LegacyEvent) error
    Send(ctx context.Context, event LegacyEvent) error
    Close() error
    Name() string
}

// EventProcessor handles different types of events
type EventProcessor interface {
    CanProcess(event LegacyEvent) bool
    Process(ctx context.Context, event LegacyEvent, agent *core.LLMAgent) (*ProcessedEvent, error)
    Name() string
}

// LegacyEvent represents an event from a legacy system
type LegacyEvent struct {
    ID          string                 `json:"id"`
    Type        string                 `json:"type"`
    Source      string                 `json:"source"`
    Data        map[string]interface{} `json:"data"`
    Timestamp   time.Time              `json:"timestamp"`
    Metadata    map[string]interface{} `json:"metadata"`
}

// ProcessedEvent represents an AI-processed event
type ProcessedEvent struct {
    OriginalEvent LegacyEvent            `json:"original_event"`
    AIProcessing  map[string]interface{} `json:"ai_processing"`
    Actions       []RecommendedAction    `json:"actions"`
    Insights      []string               `json:"insights"`
    ProcessedAt   time.Time              `json:"processed_at"`
    ProcessingTime time.Duration         `json:"processing_time"`
}

type RecommendedAction struct {
    Type        string                 `json:"type"`
    Description string                 `json:"description"`
    Parameters  map[string]interface{} `json:"parameters"`
    Priority    string                 `json:"priority"` // high, medium, low
    Confidence  float64                `json:"confidence"`
}

func NewLegacyEventBridge(config *BridgeConfig) (*LegacyEventBridge, error) {
    // Create LLM provider
    llm, err := provider.NewOpenAI(
        provider.WithModel("gpt-4"),
    )
    if err != nil {
        return nil, err
    }

    // Create agent
    agent := core.NewLLMAgent("legacy-bridge-agent", llm)

    // Create event bus
    eventBus := events.NewEventBus()

    bridge := &LegacyEventBridge{
        agent:            agent,
        eventBus:         eventBus,
        legacyConnectors: make(map[string]LegacyConnector),
        processors:       make(map[string]EventProcessor),
        config:           config,
        metrics:          &BridgeMetrics{},
    }

    // Add default processors
    bridge.AddProcessor(&CustomerEventProcessor{})
    bridge.AddProcessor(&OrderEventProcessor{})
    bridge.AddProcessor(&SystemEventProcessor{})

    return bridge, nil
}

func (leb *LegacyEventBridge) AddConnector(connector LegacyConnector) {
    leb.mu.Lock()
    defer leb.mu.Unlock()
    leb.legacyConnectors[connector.Name()] = connector
}

func (leb *LegacyEventBridge) AddProcessor(processor EventProcessor) {
    leb.mu.Lock()
    defer leb.mu.Unlock()
    leb.processors[processor.Name()] = processor
}

func (leb *LegacyEventBridge) Start(ctx context.Context) error {
    // Connect to all legacy systems
    for name, connector := range leb.legacyConnectors {
        if err := connector.Connect(ctx); err != nil {
            log.Printf("Failed to connect to %s: %v", name, err)
            continue
        }

        // Start listening for events
        go leb.listenToConnector(ctx, connector)
    }

    log.Println("Legacy Event Bridge started")
    return nil
}

func (leb *LegacyEventBridge) listenToConnector(ctx context.Context, connector LegacyConnector) {
    eventChan := make(chan LegacyEvent, leb.config.BufferSize)
    
    // Start listening
    go func() {
        if err := connector.Listen(ctx, eventChan); err != nil {
            log.Printf("Connector %s listening error: %v", connector.Name(), err)
        }
    }()

    // Process events
    for {
        select {
        case event := <-eventChan:
            go leb.processEvent(ctx, event)
        case <-ctx.Done():
            return
        }
    }
}

func (leb *LegacyEventBridge) processEvent(ctx context.Context, event LegacyEvent) {
    start := time.Now()
    
    // Update metrics
    leb.metrics.EventsProcessed++
    leb.metrics.LastProcessingTime = start

    // Find appropriate processor
    var processor EventProcessor
    for _, p := range leb.processors {
        if p.CanProcess(event) {
            processor = p
            break
        }
    }

    if processor == nil {
        log.Printf("No processor found for event type: %s", event.Type)
        leb.metrics.EventsFailed++
        return
    }

    // Process with timeout
    processCtx, cancel := context.WithTimeout(ctx, leb.config.ProcessingTimeout)
    defer cancel()

    processed, err := processor.Process(processCtx, event, leb.agent)
    if err != nil {
        log.Printf("Event processing failed: %v", err)
        leb.metrics.EventsFailed++
        
        if leb.config.EnableDeadLetter {
            leb.sendToDeadLetter(event, err)
        }
        return
    }

    // Update metrics
    leb.metrics.EventsSucceeded++
    leb.metrics.ProcessingTime = time.Since(start)

    // Publish processed event
    leb.eventBus.Publish(ctx, domain.Event{
        Type: "legacy_event_processed",
        Data: map[string]interface{}{
            "original_event": event,
            "processed_event": processed,
            "processor": processor.Name(),
        },
    })

    // Execute recommended actions
    leb.executeActions(ctx, processed.Actions)
}

func (leb *LegacyEventBridge) executeActions(ctx context.Context, actions []RecommendedAction) {
    for _, action := range actions {
        // Only execute high-confidence actions automatically
        if action.Confidence < 0.8 {
            log.Printf("Skipping low-confidence action: %s (confidence: %.2f)", action.Description, action.Confidence)
            continue
        }

        switch action.Type {
        case "notify":
            leb.sendNotification(ctx, action)
        case "update_record":
            leb.updateRecord(ctx, action)
        case "trigger_workflow":
            leb.triggerWorkflow(ctx, action)
        default:
            log.Printf("Unknown action type: %s", action.Type)
        }
    }
}

func (leb *LegacyEventBridge) sendToDeadLetter(event LegacyEvent, err error) {
    deadLetterEvent := map[string]interface{}{
        "event": event,
        "error": err.Error(),
        "timestamp": time.Now(),
    }

    data, _ := json.Marshal(deadLetterEvent)
    // Write to dead letter queue/file
    log.Printf("Dead letter: %s", string(data))
}

// Customer Event Processor
type CustomerEventProcessor struct{}

func (cep *CustomerEventProcessor) Name() string {
    return "customer_processor"
}

func (cep *CustomerEventProcessor) CanProcess(event LegacyEvent) bool {
    return event.Type == "customer_created" || 
           event.Type == "customer_updated" || 
           event.Type == "customer_deleted"
}

func (cep *CustomerEventProcessor) Process(ctx context.Context, event LegacyEvent, agent *core.LLMAgent) (*ProcessedEvent, error) {
    start := time.Now()

    // Create AI analysis prompt
    prompt := fmt.Sprintf(`Analyze this customer event and provide insights:

Event Type: %s
Customer Data: %s

Please provide:
1. Risk assessment for this customer
2. Recommendations for customer engagement
3. Potential upselling opportunities
4. Compliance considerations

Return your analysis as JSON with specific insights and recommended actions.`,
        event.Type, formatData(event.Data))

    // Process with agent
    state := domain.NewState()
    state.AddMessage(domain.NewTextMessage(domain.RoleUser, prompt))

    result, err := agent.Run(ctx, state)
    if err != nil {
        return nil, err
    }

    // Parse AI response
    var aiAnalysis map[string]interface{}
    if len(result.Messages) > 0 {
        aiResponse := result.Messages[len(result.Messages)-1].TextContent()
        json.Unmarshal([]byte(aiResponse), &aiAnalysis)
    }

    // Extract recommended actions
    actions := extractActionsFromAnalysis(aiAnalysis)

    processed := &ProcessedEvent{
        OriginalEvent:  event,
        AIProcessing:   aiAnalysis,
        Actions:        actions,
        Insights:       extractInsights(aiAnalysis),
        ProcessedAt:    time.Now(),
        ProcessingTime: time.Since(start),
    }

    return processed, nil
}

// Order Event Processor
type OrderEventProcessor struct{}

func (oep *OrderEventProcessor) Name() string {
    return "order_processor"
}

func (oep *OrderEventProcessor) CanProcess(event LegacyEvent) bool {
    return event.Type == "order_created" || 
           event.Type == "order_cancelled" || 
           event.Type == "order_shipped"
}

func (oep *OrderEventProcessor) Process(ctx context.Context, event LegacyEvent, agent *core.LLMAgent) (*ProcessedEvent, error) {
    start := time.Now()

    prompt := fmt.Sprintf(`Analyze this order event for business insights:

Event Type: %s
Order Data: %s

Provide analysis on:
1. Order patterns and trends
2. Revenue impact assessment
3. Inventory implications
4. Customer satisfaction predictions
5. Fraud risk evaluation

Return structured analysis with actionable recommendations.`,
        event.Type, formatData(event.Data))

    state := domain.NewState()
    state.AddMessage(domain.NewTextMessage(domain.RoleUser, prompt))

    result, err := agent.Run(ctx, state)
    if err != nil {
        return nil, err
    }

    var aiAnalysis map[string]interface{}
    if len(result.Messages) > 0 {
        aiResponse := result.Messages[len(result.Messages)-1].TextContent()
        json.Unmarshal([]byte(aiResponse), &aiAnalysis)
    }

    actions := extractActionsFromAnalysis(aiAnalysis)

    processed := &ProcessedEvent{
        OriginalEvent:  event,
        AIProcessing:   aiAnalysis,
        Actions:        actions,
        Insights:       extractInsights(aiAnalysis),
        ProcessedAt:    time.Now(),
        ProcessingTime: time.Since(start),
    }

    return processed, nil
}

// System Event Processor
type SystemEventProcessor struct{}

func (sep *SystemEventProcessor) Name() string {
    return "system_processor"
}

func (sep *SystemEventProcessor) CanProcess(event LegacyEvent) bool {
    return event.Type == "system_error" || 
           event.Type == "performance_alert" || 
           event.Type == "security_event"
}

func (sep *SystemEventProcessor) Process(ctx context.Context, event LegacyEvent, agent *core.LLMAgent) (*ProcessedEvent, error) {
    start := time.Now()

    prompt := fmt.Sprintf(`Analyze this system event for operational insights:

Event Type: %s
System Data: %s

Provide analysis on:
1. Severity assessment
2. Root cause analysis
3. Impact on business operations
4. Recommended remediation steps
5. Prevention strategies

Return structured analysis with prioritized actions.`,
        event.Type, formatData(event.Data))

    state := domain.NewState()
    state.AddMessage(domain.NewTextMessage(domain.RoleUser, prompt))

    result, err := agent.Run(ctx, state)
    if err != nil {
        return nil, err
    }

    var aiAnalysis map[string]interface{}
    if len(result.Messages) > 0 {
        aiResponse := result.Messages[len(result.Messages)-1].TextContent()
        json.Unmarshal([]byte(aiResponse), &aiAnalysis)
    }

    actions := extractActionsFromAnalysis(aiAnalysis)

    processed := &ProcessedEvent{
        OriginalEvent:  event,
        AIProcessing:   aiAnalysis,
        Actions:        actions,
        Insights:       extractInsights(aiAnalysis),
        ProcessedAt:    time.Now(),
        ProcessingTime: time.Since(start),
    }

    return processed, nil
}

// Utility functions
func formatData(data map[string]interface{}) string {
    jsonData, _ := json.MarshalIndent(data, "", "  ")
    return string(jsonData)
}

func extractActionsFromAnalysis(analysis map[string]interface{}) []RecommendedAction {
    // Extract actions from AI analysis
    // This is a simplified implementation
    var actions []RecommendedAction
    
    if recommendations, ok := analysis["recommendations"].([]interface{}); ok {
        for _, rec := range recommendations {
            if recMap, ok := rec.(map[string]interface{}); ok {
                action := RecommendedAction{
                    Type:        "action",
                    Description: fmt.Sprintf("%v", recMap["description"]),
                    Parameters:  make(map[string]interface{}),
                    Priority:    "medium",
                    Confidence:  0.8,
                }
                actions = append(actions, action)
            }
        }
    }
    
    return actions
}

func extractInsights(analysis map[string]interface{}) []string {
    var insights []string
    
    if insightData, ok := analysis["insights"].([]interface{}); ok {
        for _, insight := range insightData {
            insights = append(insights, fmt.Sprintf("%v", insight))
        }
    }
    
    return insights
}

func main() {
    config := &BridgeConfig{
        BufferSize:        1000,
        ProcessingTimeout: 30 * time.Second,
        RetryAttempts:     3,
        EnableMetrics:     true,
        EnableDeadLetter:  true,
        DeadLetterPath:    "/tmp/dead_letters",
    }

    bridge, err := NewLegacyEventBridge(config)
    if err != nil {
        log.Fatal("Failed to create bridge:", err)
    }

    // Add database connector (example)
    // bridge.AddConnector(NewDatabaseConnector("postgres://..."))

    ctx := context.Background()
    
    if err := bridge.Start(ctx); err != nil {
        log.Fatal("Failed to start bridge:", err)
    }

    // Keep running
    select {}
}
```

---

## Level 3: Complete System Transformation
*Full integration with monitoring, migration, and governance*

### Enterprise Integration Platform
```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "go.uber.org/zap"
    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/agent/workflow"
)

// EnterpriseIntegrationPlatform orchestrates legacy system transformation
type EnterpriseIntegrationPlatform struct {
    systems         map[string]*LegacySystem
    workflows       map[string]*workflow.SequentialAgent
    governance      *GovernanceEngine
    monitoring      *MonitoringSystem
    migration       *MigrationManager
    config          *PlatformConfig
    logger          *zap.Logger
    mu              sync.RWMutex
}

type PlatformConfig struct {
    Environment         string            `json:"environment"`
    MaxConcurrentSystems int              `json:"max_concurrent_systems"`
    DefaultTimeout      time.Duration     `json:"default_timeout"`
    EnableAuditLog      bool              `json:"enable_audit_log"`
    EnableCompliance    bool              `json:"enable_compliance"`
    GovernanceRules     []GovernanceRule  `json:"governance_rules"`
    MigrationStrategy   string            `json:"migration_strategy"`
}

type LegacySystem struct {
    ID              string                 `json:"id"`
    Name            string                 `json:"name"`
    Type            string                 `json:"type"` // database, api, mainframe, etc.
    Status          string                 `json:"status"` // active, migrating, deprecated
    Wrapper         *LegacySystemWrapper   `json:"-"`
    AICapabilities  []string               `json:"ai_capabilities"`
    MigrationPhase  int                    `json:"migration_phase"`
    LastActivity    time.Time              `json:"last_activity"`
    Metrics         *SystemMetrics         `json:"metrics"`
}

type SystemMetrics struct {
    RequestCount        int64         `json:"request_count"`
    SuccessRate         float64       `json:"success_rate"`
    AvgResponseTime     time.Duration `json:"avg_response_time"`
    AIEnhancementRate   float64       `json:"ai_enhancement_rate"`
    ErrorRate           float64       `json:"error_rate"`
    LastErrorTime       time.Time     `json:"last_error_time"`
}

type GovernanceEngine struct {
    rules           []GovernanceRule
    auditLog        []AuditEntry
    complianceCheck func(operation string, data interface{}) error
    logger          *zap.Logger
    mu              sync.RWMutex
}

type GovernanceRule struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Type        string            `json:"type"` // data_protection, compliance, security
    Conditions  map[string]interface{} `json:"conditions"`
    Actions     []string          `json:"actions"`
    Enabled     bool              `json:"enabled"`
}

type AuditEntry struct {
    Timestamp   time.Time              `json:"timestamp"`
    SystemID    string                 `json:"system_id"`
    Operation   string                 `json:"operation"`
    UserID      string                 `json:"user_id"`
    Data        map[string]interface{} `json:"data"`
    Result      string                 `json:"result"` // success, failure, blocked
    Reason      string                 `json:"reason,omitempty"`
}

type MonitoringSystem struct {
    metrics     map[string]prometheus.Metric
    alerts      chan Alert
    dashboards  map[string]*Dashboard
    logger      *zap.Logger
}

type Alert struct {
    ID          string                 `json:"id"`
    Level       string                 `json:"level"` // info, warning, critical
    SystemID    string                 `json:"system_id"`
    Message     string                 `json:"message"`
    Data        map[string]interface{} `json:"data"`
    CreatedAt   time.Time              `json:"created_at"`
    ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
}

type Dashboard struct {
    Name        string                 `json:"name"`
    Widgets     []Widget               `json:"widgets"`
    RefreshRate time.Duration          `json:"refresh_rate"`
}

type Widget struct {
    Type        string                 `json:"type"` // chart, table, metric
    Title       string                 `json:"title"`
    Query       string                 `json:"query"`
    Options     map[string]interface{} `json:"options"`
}

type MigrationManager struct {
    strategies  map[string]MigrationStrategy
    phases      []MigrationPhase
    rollback    *RollbackManager
    logger      *zap.Logger
}

type MigrationStrategy interface {
    Name() string
    Plan(system *LegacySystem) (*MigrationPlan, error)
    Execute(ctx context.Context, plan *MigrationPlan) error
    Validate(ctx context.Context, system *LegacySystem) error
}

type MigrationPhase struct {
    ID          int                    `json:"id"`
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Tasks       []MigrationTask        `json:"tasks"`
    Rollback    []RollbackStep         `json:"rollback"`
    Validation  []ValidationCheck      `json:"validation"`
}

type MigrationTask struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Type        string                 `json:"type"`
    Parameters  map[string]interface{} `json:"parameters"`
    Dependencies []string              `json:"dependencies"`
    Timeout     time.Duration          `json:"timeout"`
}

type ValidationCheck struct {
    Name        string                 `json:"name"`
    Type        string                 `json:"type"` // functional, performance, data_integrity
    Criteria    map[string]interface{} `json:"criteria"`
    Required    bool                   `json:"required"`
}

func NewEnterpriseIntegrationPlatform(config *PlatformConfig, logger *zap.Logger) (*EnterpriseIntegrationPlatform, error) {
    // Initialize governance engine
    governance := &GovernanceEngine{
        rules:    config.GovernanceRules,
        auditLog: make([]AuditEntry, 0),
        logger:   logger.Named("governance"),
    }

    // Initialize monitoring
    monitoring := &MonitoringSystem{
        metrics:    make(map[string]prometheus.Metric),
        alerts:     make(chan Alert, 1000),
        dashboards: make(map[string]*Dashboard),
        logger:     logger.Named("monitoring"),
    }

    // Initialize migration manager
    migration := &MigrationManager{
        strategies: make(map[string]MigrationStrategy),
        logger:     logger.Named("migration"),
    }

    platform := &EnterpriseIntegrationPlatform{
        systems:    make(map[string]*LegacySystem),
        workflows:  make(map[string]*workflow.SequentialAgent),
        governance: governance,
        monitoring: monitoring,
        migration:  migration,
        config:     config,
        logger:     logger,
    }

    // Register default migration strategies
    migration.strategies["gradual"] = &GradualMigrationStrategy{}
    migration.strategies["parallel"] = &ParallelMigrationStrategy{}
    migration.strategies["phased"] = &PhasedMigrationStrategy{}

    return platform, nil
}

func (eip *EnterpriseIntegrationPlatform) RegisterSystem(system *LegacySystem) error {
    eip.mu.Lock()
    defer eip.mu.Unlock()

    // Validate system configuration
    if err := eip.governance.ValidateSystemRegistration(system); err != nil {
        return fmt.Errorf("governance validation failed: %w", err)
    }

    // Initialize system metrics
    system.Metrics = &SystemMetrics{
        RequestCount:      0,
        SuccessRate:       1.0,
        AvgResponseTime:   0,
        AIEnhancementRate: 0,
        ErrorRate:         0,
    }

    // Register with monitoring
    eip.monitoring.RegisterSystem(system)

    // Add to systems map
    eip.systems[system.ID] = system

    // Log registration
    eip.governance.LogAudit(AuditEntry{
        Timestamp: time.Now(),
        SystemID:  system.ID,
        Operation: "system_registered",
        Data:      map[string]interface{}{"system": system},
        Result:    "success",
    })

    eip.logger.Info("System registered",
        zap.String("system_id", system.ID),
        zap.String("system_name", system.Name),
        zap.String("system_type", system.Type))

    return nil
}

func (eip *EnterpriseIntegrationPlatform) ExecuteMigration(systemID string) error {
    system, exists := eip.systems[systemID]
    if !exists {
        return fmt.Errorf("system not found: %s", systemID)
    }

    // Get migration strategy
    strategy, exists := eip.migration.strategies[eip.config.MigrationStrategy]
    if !exists {
        return fmt.Errorf("migration strategy not found: %s", eip.config.MigrationStrategy)
    }

    // Create migration plan
    plan, err := strategy.Plan(system)
    if err != nil {
        return fmt.Errorf("migration planning failed: %w", err)
    }

    // Execute migration
    ctx := context.Background()
    if err := strategy.Execute(ctx, plan); err != nil {
        // Trigger rollback if migration fails
        eip.migration.rollback.Execute(ctx, plan)
        return fmt.Errorf("migration execution failed: %w", err)
    }

    // Update system status
    system.Status = "migrated"
    system.MigrationPhase = len(eip.migration.phases)

    eip.logger.Info("Migration completed",
        zap.String("system_id", systemID),
        zap.String("strategy", strategy.Name()))

    return nil
}

// Governance Engine methods
func (ge *GovernanceEngine) ValidateSystemRegistration(system *LegacySystem) error {
    for _, rule := range ge.rules {
        if !rule.Enabled {
            continue
        }

        if rule.Type == "system_registration" {
            if err := ge.applyRule(rule, system); err != nil {
                return err
            }
        }
    }
    return nil
}

func (ge *GovernanceEngine) ValidateOperation(systemID, operation string, data interface{}) error {
    for _, rule := range ge.rules {
        if !rule.Enabled {
            continue
        }

        if rule.Type == "operation_validation" {
            if err := ge.applyRule(rule, map[string]interface{}{
                "system_id": systemID,
                "operation": operation,
                "data":      data,
            }); err != nil {
                return err
            }
        }
    }
    return nil
}

func (ge *GovernanceEngine) applyRule(rule GovernanceRule, data interface{}) error {
    // Simplified rule application
    // In production, this would be more sophisticated
    if rule.Name == "data_protection" {
        // Check for PII data
        if containsPII(data) {
            return fmt.Errorf("operation blocked: contains PII data")
        }
    }
    return nil
}

func (ge *GovernanceEngine) LogAudit(entry AuditEntry) {
    ge.mu.Lock()
    defer ge.mu.Unlock()
    
    ge.auditLog = append(ge.auditLog, entry)
    
    // Log to external audit system if configured
    ge.logger.Info("Audit entry",
        zap.String("system_id", entry.SystemID),
        zap.String("operation", entry.Operation),
        zap.String("result", entry.Result))
}

// Monitoring System methods
func (ms *MonitoringSystem) RegisterSystem(system *LegacySystem) {
    // Create Prometheus metrics for the system
    requestCounter := prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: fmt.Sprintf("legacy_system_%s_requests_total", system.ID),
            Help: fmt.Sprintf("Total requests to legacy system %s", system.Name),
        },
        []string{"status"},
    )

    ms.metrics[system.ID+"_requests"] = requestCounter
    prometheus.MustRegister(requestCounter)
}

func (ms *MonitoringSystem) RecordMetric(systemID, metricType string, value float64) {
    if metric, exists := ms.metrics[systemID+"_"+metricType]; exists {
        if counter, ok := metric.(prometheus.Counter); ok {
            counter.Add(value)
        }
    }
}

func (ms *MonitoringSystem) CreateAlert(alert Alert) {
    select {
    case ms.alerts <- alert:
        ms.logger.Warn("Alert created",
            zap.String("system_id", alert.SystemID),
            zap.String("level", alert.Level),
            zap.String("message", alert.Message))
    default:
        ms.logger.Error("Alert channel full, dropping alert")
    }
}

// Migration strategies
type GradualMigrationStrategy struct{}

func (gms *GradualMigrationStrategy) Name() string {
    return "gradual"
}

func (gms *GradualMigrationStrategy) Plan(system *LegacySystem) (*MigrationPlan, error) {
    // Create a gradual migration plan
    return &MigrationPlan{
        SystemID: system.ID,
        Strategy: "gradual",
        Phases: []MigrationPhase{
            {
                ID:   1,
                Name: "Read-only AI Enhancement",
                Description: "Add AI insights without modifying legacy system",
            },
            {
                ID:   2,
                Name: "Parallel Processing",
                Description: "Run AI and legacy processes in parallel",
            },
            {
                ID:   3,
                Name: "Gradual Cutover",
                Description: "Gradually shift traffic to AI-enhanced system",
            },
        },
    }, nil
}

func (gms *GradualMigrationStrategy) Execute(ctx context.Context, plan *MigrationPlan) error {
    // Execute gradual migration phases
    for _, phase := range plan.Phases {
        log.Printf("Executing migration phase: %s", phase.Name)
        
        // Execute phase tasks
        for _, task := range phase.Tasks {
            if err := gms.executeTask(ctx, task); err != nil {
                return fmt.Errorf("task %s failed: %w", task.Name, err)
            }
        }
        
        // Validate phase completion
        if err := gms.validatePhase(ctx, phase); err != nil {
            return fmt.Errorf("phase %s validation failed: %w", phase.Name, err)
        }
    }
    
    return nil
}

func (gms *GradualMigrationStrategy) Validate(ctx context.Context, system *LegacySystem) error {
    // Validate migration readiness
    return nil
}

func (gms *GradualMigrationStrategy) executeTask(ctx context.Context, task MigrationTask) error {
    // Task execution logic
    return nil
}

func (gms *GradualMigrationStrategy) validatePhase(ctx context.Context, phase MigrationPhase) error {
    // Phase validation logic
    return nil
}

// Utility functions
func containsPII(data interface{}) bool {
    // Simplified PII detection
    // In production, use proper PII detection libraries
    dataStr := fmt.Sprintf("%v", data)
    
    // Check for common PII patterns
    piiPatterns := []string{"ssn", "social", "credit_card", "email", "phone"}
    for _, pattern := range piiPatterns {
        if contains(dataStr, pattern) {
            return true
        }
    }
    
    return false
}

func contains(s, substr string) bool {
    return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && 
        (strings.Contains(strings.ToLower(s), strings.ToLower(substr)))))
}

func main() {
    // Initialize logger
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    // Platform configuration
    config := &PlatformConfig{
        Environment:         "production",
        MaxConcurrentSystems: 10,
        DefaultTimeout:      30 * time.Second,
        EnableAuditLog:      true,
        EnableCompliance:    true,
        MigrationStrategy:   "gradual",
        GovernanceRules: []GovernanceRule{
            {
                ID:      "data_protection",
                Name:    "Data Protection Rule",
                Type:    "data_protection",
                Enabled: true,
            },
        },
    }

    // Create platform
    platform, err := NewEnterpriseIntegrationPlatform(config, logger)
    if err != nil {
        log.Fatal("Failed to create platform:", err)
    }

    // Register legacy system
    legacySystem := &LegacySystem{
        ID:             "legacy-crm-001",
        Name:           "Legacy CRM System",
        Type:           "database",
        Status:         "active",
        AICapabilities: []string{"data_analysis", "customer_insights"},
        MigrationPhase: 0,
        LastActivity:   time.Now(),
    }

    if err := platform.RegisterSystem(legacySystem); err != nil {
        log.Fatal("Failed to register system:", err)
    }

    // Execute migration
    if err := platform.ExecuteMigration("legacy-crm-001"); err != nil {
        log.Printf("Migration failed: %v", err)
    }

    logger.Info("Enterprise Integration Platform started successfully")
}
```

## Migration Strategies

### Recommended Approach
1. **Assessment Phase** - Analyze existing systems and integration points
2. **Wrapper Phase** - Add AI capabilities through non-invasive wrappers
3. **Bridge Phase** - Implement event-driven integration
4. **Enhancement Phase** - Gradual replacement of legacy functionality
5. **Optimization Phase** - Performance tuning and monitoring

### Risk Mitigation
- **Rollback Plans** - Always have rollback procedures ready
- **Parallel Processing** - Run legacy and AI systems in parallel during transition
- **Gradual Cutover** - Migrate functionality incrementally
- **Monitoring** - Comprehensive monitoring and alerting
- **Testing** - Extensive testing at each phase

## Next Steps

- **[Performance Optimization](../advanced/performance-optimization.md)** - Optimize integrated systems
- **[Production Deployment](../advanced/production-deployment.md)** - Deploy to production environments
- **[Security Considerations](../advanced/security-considerations.md)** - Secure legacy integrations

---

*Gold Space, this comprehensive guide provides patterns for adding AI capabilities to existing systems without disruption. Start with simple API wrappers, then progress to event-driven integration and full system transformation as confidence and capability grow.*