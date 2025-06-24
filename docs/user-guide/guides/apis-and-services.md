# APIs and Services: Building LLM-Powered APIs

> **[Project Root](/) / [Documentation](../..) / [User Guide](../../user-guide) / [Guides](../../user-guide/guides) / APIs and Services**

Build robust, scalable APIs and microservices powered by AI agents. Master the art of creating LLM-friendly APIs, handling authentication, implementing proper error handling, and building production-ready services that leverage AI capabilities.

## Why LLM-Powered APIs Matter

- **Intelligence Layer** - Add AI capabilities to existing systems and workflows
- **Automation** - Create APIs that can understand and process natural language
- **Integration** - Bridge the gap between traditional services and AI capabilities
- **Scalability** - Handle multiple AI requests with proper resource management
- **Composability** - Build APIs that can be chained and orchestrated

## API Architecture

![LLM-Powered API Architecture](../../images/api-architecture.svg)

### Core Components
1. **Request Handler** - Process incoming API requests
2. **Agent Layer** - Execute AI processing with tools and context
3. **Authentication** - Secure access with multiple auth methods
4. **Validation** - Schema-based request and response validation
5. **Response Processor** - Format and return structured responses

### Service Patterns
| Pattern | Use Case | Benefits |
|---------|----------|----------|
| **Request-Response** | Simple AI operations | Easy to implement |
| **Streaming** | Long-running AI tasks | Real-time feedback |
| **Batch Processing** | Multiple items | Efficient resource usage |
| **Pipeline** | Multi-step workflows | Complex orchestration |
| **Event-Driven** | Asynchronous processing | Scalable and resilient |

## Prerequisites

- [Web Applications understanding](web-applications.md) ✅
- [Structured Data completed](structured-data.md) ✅
- Basic REST API knowledge ✅

---

## Level 1: Basic API Services
*Build simple LLM-powered API endpoints*

### REST API with AI Processing
```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/gorilla/mux"
    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
    "github.com/lexlapax/go-llms/pkg/schema/validation"
    schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
    "github.com/lexlapax/go-llms/pkg/structured/processor"
)

// APIService provides LLM-powered API endpoints
type APIService struct {
    agent             *core.LLMAgent
    validator         *validation.Validator
    structProcessor   *processor.StructuredProcessor
    timeout           time.Duration
}

// Request schemas for different endpoints
type TextAnalysisRequest struct {
    Text     string   `json:"text" validate:"required,min=1,max=10000"`
    Analysis []string `json:"analysis" validate:"required"`
}

type TextAnalysisResponse struct {
    Results   map[string]interface{} `json:"results"`
    Timestamp time.Time              `json:"timestamp"`
    Duration  float64                `json:"duration_ms"`
}

type SummarizationRequest struct {
    Text      string `json:"text" validate:"required,min=1"`
    MaxLength int    `json:"max_length" validate:"min=50,max=500"`
    Style     string `json:"style" validate:"oneof=concise detailed technical"`
}

type SummarizationResponse struct {
    Summary   string    `json:"summary"`
    WordCount int       `json:"word_count"`
    Timestamp time.Time `json:"timestamp"`
}

type DataExtractionRequest struct {
    Text   string                 `json:"text" validate:"required"`
    Schema map[string]interface{} `json:"schema" validate:"required"`
}

type DataExtractionResponse struct {
    Data      interface{} `json:"data"`
    Valid     bool        `json:"valid"`
    Errors    []string    `json:"errors,omitempty"`
    Timestamp time.Time   `json:"timestamp"`
}

func NewAPIService() (*APIService, error) {
    // Create LLM provider
    llm, err := provider.NewOpenAIProvider(
        provider.WithModel("gpt-4"),
        provider.WithMaxTokens(2000),
    )
    if err != nil {
        return nil, err
    }

    // Create agent
    agent := core.NewLLMAgent("api-service-agent", llm)

    // Create validator
    validator := validation.NewValidator()

    // Create structured processor
    structProcessor := processor.NewStructuredProcessor()

    return &APIService{
        agent:           agent,
        validator:       validator,
        structProcessor: structProcessor,
        timeout:         30 * time.Second,
    }, nil
}

func (s *APIService) analyzeTextHandler(w http.ResponseWriter, r *http.Request) {
    start := time.Now()

    // Parse request
    var req TextAnalysisRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        s.sendError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
        return
    }

    // Validate request
    if err := s.validator.Struct(req); err != nil {
        s.sendError(w, http.StatusBadRequest, "Validation failed: "+err.Error())
        return
    }

    // Create analysis prompt
    prompt := fmt.Sprintf(`Analyze the following text for: %v

Text: %s

Provide a JSON response with each analysis type as a key and detailed results as values.`,
        req.Analysis, req.Text)

    // Create state
    state := domain.NewState()
    state.AddMessage(domain.NewTextMessage(domain.RoleUser, prompt))

    // Process with timeout
    ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
    defer cancel()

    result, err := s.agent.Run(ctx, state)
    if err != nil {
        s.sendError(w, http.StatusInternalServerError, "Analysis failed: "+err.Error())
        return
    }

    // Extract response
    var analysisResults map[string]interface{}
    if len(result.Messages) > 0 {
        lastMessage := result.Messages[len(result.Messages)-1].TextContent()
        
        // Try to parse as JSON
        if err := json.Unmarshal([]byte(lastMessage), &analysisResults); err != nil {
            // Fallback to structured text
            analysisResults = map[string]interface{}{
                "raw_analysis": lastMessage,
            }
        }
    }

    response := TextAnalysisResponse{
        Results:   analysisResults,
        Timestamp: time.Now(),
        Duration:  float64(time.Since(start).Nanoseconds()) / 1e6,
    }

    s.sendJSON(w, http.StatusOK, response)
}

func (s *APIService) summarizeHandler(w http.ResponseWriter, r *http.Request) {
    var req SummarizationRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        s.sendError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
        return
    }

    if err := s.validator.Struct(req); err != nil {
        s.sendError(w, http.StatusBadRequest, "Validation failed: "+err.Error())
        return
    }

    // Create summarization prompt
    styleInstructions := map[string]string{
        "concise":   "Provide a brief, concise summary in 2-3 sentences.",
        "detailed":  "Provide a detailed summary with key points and context.",
        "technical": "Provide a technical summary focusing on methodology and results.",
    }

    prompt := fmt.Sprintf(`%s The summary should be approximately %d words.

Text to summarize: %s`,
        styleInstructions[req.Style], req.MaxLength/4, req.Text) // Rough word count estimate

    state := domain.NewState()
    state.AddMessage(domain.NewTextMessage(domain.RoleUser, prompt))

    ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
    defer cancel()

    result, err := s.agent.Run(ctx, state)
    if err != nil {
        s.sendError(w, http.StatusInternalServerError, "Summarization failed: "+err.Error())
        return
    }

    var summary string
    if len(result.Messages) > 0 {
        summary = result.Messages[len(result.Messages)-1].TextContent()
    }

    // Simple word count
    wordCount := len(strings.Fields(summary))

    response := SummarizationResponse{
        Summary:   summary,
        WordCount: wordCount,
        Timestamp: time.Now(),
    }

    s.sendJSON(w, http.StatusOK, response)
}

func (s *APIService) extractDataHandler(w http.ResponseWriter, r *http.Request) {
    var req DataExtractionRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        s.sendError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
        return
    }

    if err := s.validator.Struct(req); err != nil {
        s.sendError(w, http.StatusBadRequest, "Validation failed: "+err.Error())
        return
    }

    // Convert schema to JSON Schema
    schemaBytes, err := json.Marshal(req.Schema)
    if err != nil {
        s.sendError(w, http.StatusBadRequest, "Invalid schema format")
        return
    }

    schema, err := schemaDomain.NewFromJSON(schemaBytes)
    if err != nil {
        s.sendError(w, http.StatusBadRequest, "Schema parsing failed: "+err.Error())
        return
    }

    // Create extraction prompt
    prompt := fmt.Sprintf(`Extract structured data from the following text according to the provided schema.
Return ONLY valid JSON that matches the schema exactly.

Schema: %s

Text: %s`, string(schemaBytes), req.Text)

    state := domain.NewState()
    state.AddMessage(domain.NewTextMessage(domain.RoleUser, prompt))

    ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
    defer cancel()

    result, err := s.agent.Run(ctx, state)
    if err != nil {
        s.sendError(w, http.StatusInternalServerError, "Data extraction failed: "+err.Error())
        return
    }

    var extractedData interface{}
    var isValid bool
    var validationErrors []string

    if len(result.Messages) > 0 {
        output := result.Messages[len(result.Messages)-1].TextContent()
        
        // Process with structured processor
        err := s.structProcessor.ProcessTyped(schema, output, &extractedData)
        if err != nil {
            validationErrors = append(validationErrors, err.Error())
            isValid = false
            
            // Try basic JSON parsing as fallback
            if jsonErr := json.Unmarshal([]byte(output), &extractedData); jsonErr != nil {
                extractedData = map[string]interface{}{
                    "raw_text": output,
                }
                validationErrors = append(validationErrors, "JSON parsing failed: "+jsonErr.Error())
            }
        } else {
            isValid = true
        }
    }

    response := DataExtractionResponse{
        Data:      extractedData,
        Valid:     isValid,
        Errors:    validationErrors,
        Timestamp: time.Now(),
    }

    s.sendJSON(w, http.StatusOK, response)
}

func (s *APIService) healthHandler(w http.ResponseWriter, r *http.Request) {
    health := map[string]interface{}{
        "status":    "healthy",
        "timestamp": time.Now(),
        "service":   "llm-api-service",
        "version":   "1.0.0",
    }
    s.sendJSON(w, http.StatusOK, health)
}

func (s *APIService) sendJSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

func (s *APIService) sendError(w http.ResponseWriter, status int, message string) {
    errorResp := map[string]interface{}{
        "error":     message,
        "timestamp": time.Now(),
        "status":    status,
    }
    s.sendJSON(w, status, errorResp)
}

func main() {
    service, err := NewAPIService()
    if err != nil {
        log.Fatal("Failed to create API service:", err)
    }

    r := mux.NewRouter()

    // API routes
    api := r.PathPrefix("/api/v1").Subrouter()
    api.HandleFunc("/analyze", service.analyzeTextHandler).Methods("POST")
    api.HandleFunc("/summarize", service.summarizeHandler).Methods("POST")
    api.HandleFunc("/extract", service.extractDataHandler).Methods("POST")
    api.HandleFunc("/health", service.healthHandler).Methods("GET")

    // Add CORS middleware
    r.Use(func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Access-Control-Allow-Origin", "*")
            w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
            w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
            
            if r.Method == "OPTIONS" {
                w.WriteHeader(http.StatusOK)
                return
            }
            
            next.ServeHTTP(w, r)
}
}

    log.Println("API service starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}
```

---

## Level 2: Advanced API Features
*Implement authentication, streaming, and batch processing*

### Authenticated API with Multiple Auth Methods
```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
    "time"

    "github.com/golang-jwt/jwt/v4"
    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/util/auth"
)

// AuthenticatedAPIService provides secure LLM API endpoints
type AuthenticatedAPIService struct {
    agent      *core.LLMAgent
    jwtSecret  []byte
    apiKeys    map[string]UserInfo
    rateLimits map[string]*RateLimit
}

type UserInfo struct {
    ID          string   `json:"id"`
    Email       string   `json:"email"`
    Tier        string   `json:"tier"` // free, premium, enterprise
    Permissions []string `json:"permissions"`
}

type RateLimit struct {
    RequestsPerMinute int
    TokensPerMinute   int
    LastReset         time.Time
    CurrentRequests   int
    CurrentTokens     int
}

type AuthContext struct {
    User        UserInfo `json:"user"`
    AuthMethod  string   `json:"auth_method"`
    RequestID   string   `json:"request_id"`
}

func NewAuthenticatedAPIService(jwtSecret []byte) (*AuthenticatedAPIService, error) {
    llm, err := provider.NewOpenAIProvider(provider.WithModel("gpt-4"))
    if err != nil {
        return nil, err
    }

    agent := core.NewLLMAgent("auth-api-agent", llm)

    // Initialize API keys (in production, load from database)
    apiKeys := map[string]UserInfo{
        "ak_test_123": {
            ID:          "user_1",
            Email:       "test@example.com",
            Tier:        "free",
            Permissions: []string{"analyze", "summarize"},
        },
        "ak_premium_456": {
            ID:          "user_2",
            Email:       "premium@example.com",
            Tier:        "premium",
            Permissions: []string{"analyze", "summarize", "extract", "batch"},
        },
    }

    return &AuthenticatedAPIService{
        agent:      agent,
        jwtSecret:  jwtSecret,
        apiKeys:    apiKeys,
        rateLimits: make(map[string]*RateLimit),
    }, nil
}

// Authentication middleware
func (s *AuthenticatedAPIService) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        authCtx, err := s.authenticate(r)
        if err != nil {
            s.sendError(w, http.StatusUnauthorized, err.Error())
            return
        }

        // Check rate limits
        if !s.checkRateLimit(authCtx.User.ID, authCtx.User.Tier) {
            s.sendError(w, http.StatusTooManyRequests, "Rate limit exceeded")
            return
        }

        // Add auth context to request
        ctx := context.WithValue(r.Context(), "auth", authCtx)
        next.ServeHTTP(w, r.WithContext(ctx))
    }
}

func (s *AuthenticatedAPIService) authenticate(r *http.Request) (*AuthContext, error) {
    // Try API Key authentication first
    if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
        if user, exists := s.apiKeys[apiKey]; exists {
            return &AuthContext{
                User:       user,
                AuthMethod: "api_key",
                RequestID:  generateRequestID(),
            }, nil
        }
        return nil, fmt.Errorf("invalid API key")
    }

    // Try Bearer token authentication
    authHeader := r.Header.Get("Authorization")
    if strings.HasPrefix(authHeader, "Bearer ") {
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("unexpected signing method")
            }
            return s.jwtSecret, nil
}

        if err != nil {
            return nil, fmt.Errorf("invalid JWT token: %v", err)
        }

        if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
            user := UserInfo{
                ID:          claims["user_id"].(string),
                Email:       claims["email"].(string),
                Tier:        claims["tier"].(string),
                Permissions: strings.Split(claims["permissions"].(string), ","),
            }

            return &AuthContext{
                User:       user,
                AuthMethod: "jwt",
                RequestID:  generateRequestID(),
            }, nil
        }
        return nil, fmt.Errorf("invalid JWT claims")
    }

    return nil, fmt.Errorf("no valid authentication method found")
}

func (s *AuthenticatedAPIService) checkRateLimit(userID, tier string) bool {
    limits := map[string]RateLimit{
        "free":       {RequestsPerMinute: 10, TokensPerMinute: 5000},
        "premium":    {RequestsPerMinute: 100, TokensPerMinute: 50000},
        "enterprise": {RequestsPerMinute: 1000, TokensPerMinute: 500000},
    }

    limit, exists := limits[tier]
    if !exists {
        return false
    }

    now := time.Now()
    userLimit, exists := s.rateLimits[userID]
    if !exists || now.Sub(userLimit.LastReset) > time.Minute {
        s.rateLimits[userID] = &RateLimit{
            RequestsPerMinute: limit.RequestsPerMinute,
            TokensPerMinute:   limit.TokensPerMinute,
            LastReset:         now,
            CurrentRequests:   1,
            CurrentTokens:     0,
        }
        return true
    }

    if userLimit.CurrentRequests >= userLimit.RequestsPerMinute {
        return false
    }

    userLimit.CurrentRequests++
    return true
}

// Permission-based endpoint
func (s *AuthenticatedAPIService) secureAnalyzeHandler(w http.ResponseWriter, r *http.Request) {
    authCtx := r.Context().Value("auth").(*AuthContext)
    
    // Check permissions
    if !hasPermission(authCtx.User.Permissions, "analyze") {
        s.sendError(w, http.StatusForbidden, "Insufficient permissions")
        return
    }

    // Process request (similar to previous example but with user context)
    var req TextAnalysisRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        s.sendError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
        return
    }

    // Add user context to agent state
    state := domain.NewState()
    state.Set("user_id", authCtx.User.ID)
    state.Set("user_tier", authCtx.User.Tier)
    state.Set("request_id", authCtx.RequestID)

    // Process with agent
    result, err := s.processWithContext(state, req.Text, req.Analysis)
    if err != nil {
        s.sendError(w, http.StatusInternalServerError, "Processing failed: "+err.Error())
        return
    }

    // Update rate limit with token usage
    s.updateTokenUsage(authCtx.User.ID, result.TokensUsed)

    response := map[string]interface{}{
        "results":    result.Data,
        "request_id": authCtx.RequestID,
        "user_id":    authCtx.User.ID,
        "tokens_used": result.TokensUsed,
        "timestamp":  time.Now(),
    }

    s.sendJSON(w, http.StatusOK, response)
}

func hasPermission(permissions []string, required string) bool {
    for _, perm := range permissions {
        if perm == required {
            return true
        }
    }
    return false
}

func generateRequestID() string {
    return fmt.Sprintf("req_%d", time.Now().UnixNano())
}
```

### Streaming API Endpoint
```go
func (s *AuthenticatedAPIService) streamAnalysisHandler(w http.ResponseWriter, r *http.Request) {
    authCtx := r.Context().Value("auth").(*AuthContext)
    
    if !hasPermission(authCtx.User.Permissions, "stream") {
        s.sendError(w, http.StatusForbidden, "Streaming not allowed")
        return
    }

    // Set up SSE headers
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    w.Header().Set("Access-Control-Allow-Origin", "*")

    var req TextAnalysisRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
        return
    }

    flusher, ok := w.(http.Flusher)
    if !ok {
        s.sendError(w, http.StatusInternalServerError, "Streaming not supported")
        return
    }

    // Send initial event
    fmt.Fprintf(w, "event: start\ndata: {\"request_id\":\"%s\"}\n\n", authCtx.RequestID)
    flusher.Flush()

    // Process analysis step by step
    for i, analysisType := range req.Analysis {
        fmt.Fprintf(w, "event: progress\ndata: {\"step\":%d,\"total\":%d,\"type\":\"%s\"}\n\n", 
            i+1, len(req.Analysis), analysisType)
        flusher.Flush()

        // Perform individual analysis
        result, err := s.performSingleAnalysis(req.Text, analysisType)
        if err != nil {
            fmt.Fprintf(w, "event: error\ndata: {\"type\":\"%s\",\"error\":\"%s\"}\n\n", 
                analysisType, err.Error())
            flusher.Flush()
            continue
        }

        // Send result
        resultJSON, _ := json.Marshal(result)
        fmt.Fprintf(w, "event: result\ndata: {\"type\":\"%s\",\"result\":%s}\n\n", 
            analysisType, string(resultJSON))
        flusher.Flush()

        // Small delay to demonstrate streaming
        time.Sleep(500 * time.Millisecond)
    }

    // Send completion event
    fmt.Fprintf(w, "event: complete\ndata: {\"request_id\":\"%s\"}\n\n", authCtx.RequestID)
    flusher.Flush()
}
```

### Batch Processing API
```go
type BatchRequest struct {
    Items   []BatchItem `json:"items" validate:"required,min=1,max=100"`
    Options BatchOptions `json:"options"`
}

type BatchItem struct {
    ID       string                 `json:"id" validate:"required"`
    Type     string                 `json:"type" validate:"required,oneof=analyze summarize extract"`
    Data     map[string]interface{} `json:"data" validate:"required"`
}

type BatchOptions struct {
    Parallel     bool `json:"parallel"`
    MaxRetries   int  `json:"max_retries"`
    FailFast     bool `json:"fail_fast"`
}

type BatchResponse struct {
    Results   []BatchResult `json:"results"`
    Summary   BatchSummary  `json:"summary"`
    RequestID string        `json:"request_id"`
}

type BatchResult struct {
    ID       string      `json:"id"`
    Status   string      `json:"status"` // success, error, skipped
    Data     interface{} `json:"data,omitempty"`
    Error    string      `json:"error,omitempty"`
    Duration float64     `json:"duration_ms"`
}

type BatchSummary struct {
    Total     int     `json:"total"`
    Success   int     `json:"success"`
    Errors    int     `json:"errors"`
    Duration  float64 `json:"total_duration_ms"`
    TokensUsed int    `json:"tokens_used"`
}

func (s *AuthenticatedAPIService) batchProcessHandler(w http.ResponseWriter, r *http.Request) {
    authCtx := r.Context().Value("auth").(*AuthContext)
    
    if !hasPermission(authCtx.User.Permissions, "batch") {
        s.sendError(w, http.StatusForbidden, "Batch processing not allowed")
        return
    }

    var req BatchRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        s.sendError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
        return
    }

    if err := s.validator.Struct(req); err != nil {
        s.sendError(w, http.StatusBadRequest, "Validation failed: "+err.Error())
        return
    }

    start := time.Now()
    results := make([]BatchResult, len(req.Items))
    summary := BatchSummary{Total: len(req.Items)}

    if req.Options.Parallel {
        // Process items in parallel
        results = s.processBatchParallel(req.Items, &summary)
    } else {
        // Process items sequentially
        for i, item := range req.Items {
            result := s.processBatchItem(item, req.Options)
            results[i] = result
            
            if result.Status == "success" {
                summary.Success++
            } else {
                summary.Errors++
                if req.Options.FailFast {
                    break
                }
            }
        }
    }

    summary.Duration = float64(time.Since(start).Nanoseconds()) / 1e6

    response := BatchResponse{
        Results:   results,
        Summary:   summary,
        RequestID: authCtx.RequestID,
    }

    s.sendJSON(w, http.StatusOK, response)
}

func (s *AuthenticatedAPIService) processBatchParallel(items []BatchItem, summary *BatchSummary) []BatchResult {
    results := make([]BatchResult, len(items))
    var wg sync.WaitGroup
    var mu sync.Mutex

    for i, item := range items {
        wg.Add(1)
        go func(index int, item BatchItem) {
            defer wg.Done()
            
            result := s.processBatchItem(item, BatchOptions{})
            
            mu.Lock()
            results[index] = result
            if result.Status == "success" {
                summary.Success++
            } else {
                summary.Errors++
            }
            mu.Unlock()
        }(i, item)
    }

    wg.Wait()
    return results
}
```

---

## Level 3: Production API Services
*Build enterprise-grade APIs with monitoring and scalability*

### Microservice with gRPC and REST
```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net"
    "net/http"
    "time"

    "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
    
    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/agent/events"
)

// Metrics
var (
    apiRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "api_requests_total",
            Help: "Total number of API requests",
        },
        []string{"method", "endpoint", "status"},
    )

    apiRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "api_request_duration_seconds",
            Help: "API request duration in seconds",
        },
        []string{"method", "endpoint"},
    )

    llmTokensUsed = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "llm_tokens_used_total",
            Help: "Total LLM tokens used",
        },
        []string{"provider", "model", "user_tier"},
    )

    activeConnections = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "active_connections",
            Help: "Number of active connections",
        },
    )
)

func init() {
    prometheus.MustRegister(apiRequestsTotal)
    prometheus.MustRegister(apiRequestDuration)
    prometheus.MustRegister(llmTokensUsed)
    prometheus.MustRegister(activeConnections)
}

// LLMAPIServer provides both gRPC and REST APIs
type LLMAPIServer struct {
    agent           *core.LLMAgent
    eventBus        *events.EventBus
    config          *ServerConfig
}

type ServerConfig struct {
    GRPCPort      int    `json:"grpc_port"`
    HTTPPort      int    `json:"http_port"`
    MetricsPort   int    `json:"metrics_port"`
    MaxConcurrent int    `json:"max_concurrent"`
    Timeout       time.Duration `json:"timeout"`
    EnableTracing bool   `json:"enable_tracing"`
    LogLevel      string `json:"log_level"`
}

// Circuit breaker for external dependencies
type CircuitBreaker struct {
    maxFailures int
    timeout     time.Duration
    failures    int
    lastFailure time.Time
    state       string // closed, open, half-open
}

func NewLLMAPIServer(config *ServerConfig) (*LLMAPIServer, error) {
    // Create LLM provider with retry logic
    llm, err := provider.NewOpenAIProvider(
        provider.WithModel("gpt-4"),
        provider.WithTimeout(config.Timeout),
        provider.WithRetries(3),
    )
    if err != nil {
        return nil, err
    }

    // Create agent with error handling
    agent := core.NewLLMAgent("production-api-agent", llm)

    // Create event bus for monitoring
    eventBus := events.NewEventBus()
    
    // Subscribe to events for monitoring
    eventBus.Subscribe(events.EventHandlerFunc(func(ctx context.Context, event domain.Event) error {
        // Log important events
        log.Printf("Event: %s, Data: %+v", event.Type, event.Data)
        return nil
    }))

    return &LLMAPIServer{
        agent:    agent,
        eventBus: eventBus,
        config:   config,
    }, nil
}

func (s *LLMAPIServer) Start() error {
    // Start gRPC server
    go func() {
        if err := s.startGRPCServer(); err != nil {
            log.Fatalf("gRPC server failed: %v", err)
        }
    }()

    // Start HTTP gateway
    go func() {
        if err := s.startHTTPGateway(); err != nil {
            log.Fatalf("HTTP gateway failed: %v", err)
        }
    }()

    // Start metrics server
    go func() {
        if err := s.startMetricsServer(); err != nil {
            log.Fatalf("Metrics server failed: %v", err)
        }
    }()

    return nil
}

func (s *LLMAPIServer) startGRPCServer() error {
    lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.GRPCPort))
    if err != nil {
        return err
    }

    grpcServer := grpc.NewServer(
        grpc.UnaryInterceptor(s.unaryInterceptor),
        grpc.StreamInterceptor(s.streamInterceptor),
    )

    // Register services
    // pb.RegisterLLMServiceServer(grpcServer, s)

    // Enable reflection for development
    reflection.Register(grpcServer)

    log.Printf("gRPC server listening on :%d", s.config.GRPCPort)
    return grpcServer.Serve(lis)
}

func (s *LLMAPIServer) startHTTPGateway() error {
    ctx := context.Background()
    mux := runtime.NewServeMux()

    opts := []grpc.DialOption{grpc.WithInsecure()}
    
    // Register gRPC gateway
    // err := pb.RegisterLLMServiceHandlerFromEndpoint(ctx, mux, 
    //     fmt.Sprintf("localhost:%d", s.config.GRPCPort), opts)
    // if err != nil {
    //     return err
    // }

    // Add CORS and other middleware
    handler := s.corsMiddleware(s.loggingMiddleware(mux))

    log.Printf("HTTP gateway listening on :%d", s.config.HTTPPort)
    return http.ListenAndServe(fmt.Sprintf(":%d", s.config.HTTPPort), handler)
}

func (s *LLMAPIServer) startMetricsServer() error {
    mux := http.NewServeMux()
    mux.Handle("/metrics", promhttp.Handler())
    mux.HandleFunc("/health", s.healthHandler)
    mux.HandleFunc("/ready", s.readinessHandler)

    log.Printf("Metrics server listening on :%d", s.config.MetricsPort)
    return http.ListenAndServe(fmt.Sprintf(":%d", s.config.MetricsPort), mux)
}

// Middleware implementations
func (s *LLMAPIServer) unaryInterceptor(ctx context.Context, req interface{}, 
    info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
    
    start := time.Now()
    activeConnections.Inc()
    defer activeConnections.Dec()

    resp, err := handler(ctx, req)

    // Record metrics
    duration := time.Since(start)
    status := "success"
    if err != nil {
        status = "error"
    }

    apiRequestsTotal.WithLabelValues("grpc", info.FullMethod, status).Inc()
    apiRequestDuration.WithLabelValues("grpc", info.FullMethod).Observe(duration.Seconds())

    return resp, err
}

func (s *LLMAPIServer) streamInterceptor(srv interface{}, ss grpc.ServerStream, 
    info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
    
    activeConnections.Inc()
    defer activeConnections.Dec()

    return handler(srv, ss)
}

func (s *LLMAPIServer) corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")

        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
}
}

func (s *LLMAPIServer) loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        next.ServeHTTP(w, r)
        
        log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
}
}

func (s *LLMAPIServer) healthHandler(w http.ResponseWriter, r *http.Request) {
    health := map[string]interface{}{
        "status":    "healthy",
        "timestamp": time.Now(),
        "version":   "1.0.0",
        "uptime":    time.Since(startTime).String(),
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(health)
}

func (s *LLMAPIServer) readinessHandler(w http.ResponseWriter, r *http.Request) {
    // Check dependencies (database, LLM provider, etc.)
    ready := s.checkDependencies()
    
    status := map[string]interface{}{
        "ready":     ready,
        "timestamp": time.Now(),
    }
    
    if ready {
        w.WriteHeader(http.StatusOK)
    } else {
        w.WriteHeader(http.StatusServiceUnavailable)
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(status)
}

func (s *LLMAPIServer) checkDependencies() bool {
    // Implement health checks for:
    // - Database connectivity
    // - LLM provider availability
    // - External service dependencies
    // - Resource availability (memory, disk)
    return true
}

var startTime = time.Now()

func main() {
    config := &ServerConfig{
        GRPCPort:      8080,
        HTTPPort:      8081,
        MetricsPort:   8082,
        MaxConcurrent: 100,
        Timeout:       30 * time.Second,
        EnableTracing: true,
        LogLevel:      "info",
    }

    server, err := NewLLMAPIServer(config)
    if err != nil {
        log.Fatal("Failed to create server:", err)
    }

    log.Println("Starting LLM API Server...")
    if err := server.Start(); err != nil {
        log.Fatal("Server failed to start:", err)
    }

    // Keep the main goroutine alive
    select {}
}
```

## API Testing Framework

### Comprehensive Testing Suite
```go
package main

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/stretchr/testify/suite"
)

type APITestSuite struct {
    suite.Suite
    server  *httptest.Server
    service *APIService
    client  *http.Client
}

func (s *APITestSuite) SetupSuite() {
    // Create test service
    service, err := NewAPIService()
    require.NoError(s.T(), err)
    
    // Create test server
    handler := s.createTestRouter(service)
    s.server = httptest.NewServer(handler)
    s.service = service
    s.client = &http.Client{Timeout: 10 * time.Second}
}

func (s *APITestSuite) TearDownSuite() {
    s.server.Close()
}

func (s *APITestSuite) TestTextAnalysis() {
    req := TextAnalysisRequest{
        Text:     "This is a sample text for analysis.",
        Analysis: []string{"sentiment", "keywords"},
    }

    resp, err := s.makeRequest("POST", "/api/v1/analyze", req)
    require.NoError(s.T(), err)
    defer resp.Body.Close()

    assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

    var response TextAnalysisResponse
    err = json.NewDecoder(resp.Body).Decode(&response)
    require.NoError(s.T(), err)

    assert.NotEmpty(s.T(), response.Results)
    assert.NotZero(s.T(), response.Duration)
}

func (s *APITestSuite) TestRateLimiting() {
    // Create requests exceeding rate limit
    req := TextAnalysisRequest{
        Text:     "Test text",
        Analysis: []string{"sentiment"},
    }

    // Should succeed within limit
    for i := 0; i < 5; i++ {
        resp, err := s.makeRequest("POST", "/api/v1/analyze", req)
        require.NoError(s.T(), err)
        resp.Body.Close()
        assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
    }

    // Should fail after exceeding limit
    resp, err := s.makeRequest("POST", "/api/v1/analyze", req)
    require.NoError(s.T(), err)
    resp.Body.Close()
    assert.Equal(s.T(), http.StatusTooManyRequests, resp.StatusCode)
}

func (s *APITestSuite) TestAuthenticationRequired() {
    req := TextAnalysisRequest{
        Text:     "Test text",
        Analysis: []string{"sentiment"},
    }

    // Request without authentication should fail
    resp, err := s.makeRequestWithoutAuth("POST", "/api/v1/secure/analyze", req)
    require.NoError(s.T(), err)
    defer resp.Body.Close()

    assert.Equal(s.T(), http.StatusUnauthorized, resp.StatusCode)
}

func (s *APITestSuite) TestBatchProcessing() {
    req := BatchRequest{
        Items: []BatchItem{
            {
                ID:   "item1",
                Type: "analyze",
                Data: map[string]interface{}{
                    "text":     "First text",
                    "analysis": []string{"sentiment"},
                },
            },
            {
                ID:   "item2",
                Type: "summarize",
                Data: map[string]interface{}{
                    "text":       "Second text to summarize",
                    "max_length": 100,
                    "style":      "concise",
                },
            },
        },
        Options: BatchOptions{
            Parallel: true,
        },
    }

    resp, err := s.makeRequestWithAuth("POST", "/api/v1/batch", req)
    require.NoError(s.T(), err)
    defer resp.Body.Close()

    assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

    var response BatchResponse
    err = json.NewDecoder(resp.Body).Decode(&response)
    require.NoError(s.T(), err)

    assert.Len(s.T(), response.Results, 2)
    assert.Equal(s.T(), 2, response.Summary.Total)
}

func (s *APITestSuite) makeRequest(method, path string, body interface{}) (*http.Response, error) {
    jsonBody, _ := json.Marshal(body)
    req, _ := http.NewRequest(method, s.server.URL+path, bytes.NewBuffer(jsonBody))
    req.Header.Set("Content-Type", "application/json")
    return s.client.Do(req)
}

func (s *APITestSuite) makeRequestWithAuth(method, path string, body interface{}) (*http.Response, error) {
    jsonBody, _ := json.Marshal(body)
    req, _ := http.NewRequest(method, s.server.URL+path, bytes.NewBuffer(jsonBody))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-API-Key", "ak_premium_456") // Use premium test key
    return s.client.Do(req)
}

func TestAPITestSuite(t *testing.T) {
    suite.Run(t, new(APITestSuite))
}
```

## Performance Optimization

### Key Optimization Strategies
1. **Connection Pooling** - Reuse HTTP connections
2. **Caching** - Cache frequent LLM responses
3. **Request Batching** - Combine multiple requests
4. **Async Processing** - Non-blocking operations
5. **Resource Pooling** - Reuse expensive resources

### Monitoring and Observability
- **Metrics Collection** - Prometheus integration
- **Distributed Tracing** - Request flow tracking
- **Health Checks** - Service health monitoring
- **Alerting** - Automated incident response

## Next Steps

- **[Databases](databases.md)** - Integrate with databases for persistent storage
- **[Existing Systems](existing-systems.md)** - Add LLM capabilities to legacy systems
- **[Production Deployment](../advanced/production-deployment.md)** - Deploy APIs to production

---

*Gold Space, this comprehensive guide covers building production-ready LLM-powered APIs from basic endpoints to enterprise microservices with full authentication, monitoring, and scalability features.*