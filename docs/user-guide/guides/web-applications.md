# Web Applications: Framework Integration

> **[Project Root](/) / [Documentation](/docs/) / [User Guide](/docs/user-guide/) / [Guides](/docs/user-guide/guides/) / Web Applications**

Build web applications powered by AI agents. Master the integration of Go-LLMs with popular web frameworks, handle concurrent requests, manage user sessions, and create real-time AI-powered experiences.

## Why Web Integration Matters

- **Scalability** - Handle multiple users and concurrent requests efficiently
- **Real-time** - Provide instant AI responses with streaming and WebSocket support
- **Stateful** - Maintain user sessions and conversation context
- **Integration** - Connect AI agents with existing web infrastructure
- **Production** - Deploy robust, monitored AI-powered web services

## Web Integration Architecture

![Web Application Architecture](../../images/web-integration-architecture.svg)

### Integration Layers
1. **HTTP Layer** - Route requests and handle responses
2. **Agent Layer** - Process requests with AI capabilities
3. **State Layer** - Manage user sessions and conversation context
4. **Event Layer** - Handle real-time updates and notifications
5. **Persistence Layer** - Store conversations and user data

### Framework Support
| Framework | Complexity | Best For | Integration Pattern |
|-----------|------------|----------|---------------------|
| **net/http** | Simple | APIs, microservices | Direct handler integration |
| **Gin** | Medium | REST APIs, SPAs | Middleware and service layers |
| **Echo** | Medium | High-performance APIs | Context-aware handlers |
| **Fiber** | Medium | Express-like APIs | Middleware chains |
| **Chi** | Simple | Composable APIs | Route-based handlers |

## Prerequisites

- [Agent Creation completed](creating-agents.md) ✅
- [State Management understanding](agent-memory.md) ✅
- Basic web framework knowledge ✅

---

## Level 1: Basic Web Integration
*Integrate AI agents into simple web applications*

### HTTP Handler with AI Agent
```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
)

// AIService handles AI-powered web requests
type AIService struct {
    agent *core.LLMAgent
    state *domain.State
}

// Request represents an AI request
type Request struct {
    Message string `json:"message"`
    UserID  string `json:"user_id,omitempty"`
}

// Response represents an AI response
type Response struct {
    Reply     string    `json:"reply"`
    Timestamp time.Time `json:"timestamp"`
    TokensUsed int      `json:"tokens_used,omitempty"`
}

func NewAIService() (*AIService, error) {
    // Create OpenAI provider
    llm, err := provider.NewOpenAI(
        provider.WithModel("gpt-4"),
        provider.WithMaxTokens(1000),
    )
    if err != nil {
        return nil, err
    }

    // Create agent
    agent := core.NewLLMAgent("web-assistant", llm)

    // Create initial state
    state := domain.NewState()
    
    return &AIService{
        agent: agent,
        state: state,
    }, nil
}

func (s *AIService) handleChat(w http.ResponseWriter, r *http.Request) {
    // Parse request
    var req Request
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    // Set user context
    userState := s.state.Clone()
    if req.UserID != "" {
        userState.Set("user_id", req.UserID)
    }

    // Add user message
    userState.AddMessage(domain.NewTextMessage(domain.RoleUser, req.Message))

    // Run agent
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    result, err := s.agent.Run(ctx, userState)
    if err != nil {
        http.Error(w, "AI processing failed", http.StatusInternalServerError)
        return
    }

    // Extract response
    var reply string
    if len(result.Messages) > 0 {
        reply = result.Messages[len(result.Messages)-1].TextContent()
    }

    // Create response
    response := Response{
        Reply:     reply,
        Timestamp: time.Now(),
        TokensUsed: result.TokensUsed,
    }

    // Send response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func main() {
    // Create AI service
    service, err := NewAIService()
    if err != nil {
        log.Fatal("Failed to create AI service:", err)
    }

    // Setup routes
    http.HandleFunc("/chat", service.handleChat)
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        fmt.Fprint(w, "OK")
    })

    // Start server
    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### Gin Framework Integration
```go
package main

import (
    "context"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
)

// AIController handles AI endpoints
type AIController struct {
    agent *core.LLMAgent
}

func NewAIController() (*AIController, error) {
    // Create provider
    llm, err := provider.NewOpenAI(
        provider.WithModel("gpt-4"),
    )
    if err != nil {
        return nil, err
    }

    // Create agent
    agent := core.NewLLMAgent("gin-assistant", llm)

    return &AIController{agent: agent}, nil
}

func (c *AIController) Chat(ctx *gin.Context) {
    var req Request
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Create state
    state := domain.NewState()
    state.Set("user_id", req.UserID)
    state.AddMessage(domain.NewTextMessage(domain.RoleUser, req.Message))

    // Run agent
    reqCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    result, err := c.agent.Run(reqCtx, state)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "AI processing failed"})
        return
    }

    // Extract response
    var reply string
    if len(result.Messages) > 0 {
        reply = result.Messages[len(result.Messages)-1].TextContent()
    }

    response := Response{
        Reply:     reply,
        Timestamp: time.Now(),
        TokensUsed: result.TokensUsed,
    }

    ctx.JSON(http.StatusOK, response)
}

func main() {
    // Create controller
    controller, err := NewAIController()
    if err != nil {
        log.Fatal("Failed to create controller:", err)
    }

    // Setup Gin
    r := gin.Default()

    // Middleware
    r.Use(gin.Logger())
    r.Use(gin.Recovery())

    // Routes
    api := r.Group("/api/v1")
    {
        api.POST("/chat", controller.Chat)
    }

    // Start server
    r.Run(":8080")
}
```

---

## Level 2: Advanced Web Features
*Implement real-time communication and session management*

### WebSocket Real-time Chat
```go
package main

import (
    "context"
    "log"
    "net/http"
    "sync"
    "time"

    "github.com/gorilla/websocket"
    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/agent/events"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // Configure appropriately for production
    },
}

// WebSocketServer manages real-time AI chat
type WebSocketServer struct {
    agent     *core.LLMAgent
    eventBus  *events.EventBus
    clients   map[*websocket.Conn]*ClientSession
    mutex     sync.RWMutex
}

// ClientSession represents a connected client
type ClientSession struct {
    conn   *websocket.Conn
    userID string
    state  *domain.State
    send   chan []byte
}

// Message types for WebSocket communication
type WSMessage struct {
    Type    string      `json:"type"`
    Content interface{} `json:"content"`
}

type ChatMessage struct {
    Text   string `json:"text"`
    UserID string `json:"user_id"`
}

type AIResponse struct {
    Reply      string    `json:"reply"`
    Timestamp  time.Time `json:"timestamp"`
    TokensUsed int       `json:"tokens_used"`
}

func NewWebSocketServer() (*WebSocketServer, error) {
    // Create provider
    llm, err := provider.NewOpenAI(
        provider.WithModel("gpt-4"),
    )
    if err != nil {
        return nil, err
    }

    // Create agent with streaming support
    agent := core.NewLLMAgent("websocket-assistant", llm)

    // Create event bus for real-time updates
    eventBus := events.NewEventBus()

    return &WebSocketServer{
        agent:    agent,
        eventBus: eventBus,
        clients:  make(map[*websocket.Conn]*ClientSession),
    }, nil
}

func (s *WebSocketServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("WebSocket upgrade error:", err)
        return
    }
    defer conn.Close()

    // Create client session
    session := &ClientSession{
        conn:   conn,
        userID: r.URL.Query().Get("user_id"),
        state:  domain.NewState(),
        send:   make(chan []byte, 256),
    }

    // Register client
    s.mutex.Lock()
    s.clients[conn] = session
    s.mutex.Unlock()

    // Start goroutines
    go s.writePump(session)
    go s.readPump(session)
}

func (s *WebSocketServer) readPump(session *ClientSession) {
    defer func() {
        s.mutex.Lock()
        delete(s.clients, session.conn)
        s.mutex.Unlock()
        session.conn.Close()
    }()

    session.conn.SetReadLimit(512)
    session.conn.SetReadDeadline(time.Now().Add(60 * time.Second))

    for {
        var msg WSMessage
        if err := session.conn.ReadJSON(&msg); err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                log.Printf("WebSocket error: %v", err)
            }
            break
        }

        switch msg.Type {
        case "chat":
            s.handleChatMessage(session, msg.Content)
        case "ping":
            s.sendMessage(session, WSMessage{Type: "pong"})
        }
    }
}

func (s *WebSocketServer) writePump(session *ClientSession) {
    ticker := time.NewTicker(54 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case message, ok := <-session.send:
            session.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
            if !ok {
                session.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }

            if err := session.conn.WriteMessage(websocket.TextMessage, message); err != nil {
                return
            }

        case <-ticker.C:
            session.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
            if err := session.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}

func (s *WebSocketServer) handleChatMessage(session *ClientSession, content interface{}) {
    // Parse chat message
    msgData, ok := content.(map[string]interface{})
    if !ok {
        return
    }

    text, ok := msgData["text"].(string)
    if !ok {
        return
    }

    // Add user message to state
    session.state.AddMessage(domain.NewTextMessage(domain.RoleUser, text))

    // Send typing indicator
    s.sendMessage(session, WSMessage{
        Type:    "typing",
        Content: map[string]bool{"typing": true},
    })

    // Process with agent
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    result, err := s.agent.Run(ctx, session.state)
    if err != nil {
        s.sendMessage(session, WSMessage{
            Type:    "error",
            Content: map[string]string{"error": "AI processing failed"},
        })
        return
    }

    // Stop typing indicator
    s.sendMessage(session, WSMessage{
        Type:    "typing",
        Content: map[string]bool{"typing": false},
    })

    // Extract response
    var reply string
    if len(result.Messages) > 0 {
        reply = result.Messages[len(result.Messages)-1].TextContent()
    }

    // Send AI response
    response := AIResponse{
        Reply:      reply,
        Timestamp:  time.Now(),
        TokensUsed: result.TokensUsed,
    }

    s.sendMessage(session, WSMessage{
        Type:    "ai_response",
        Content: response,
    })

    // Update session state
    session.state = result
}

func (s *WebSocketServer) sendMessage(session *ClientSession, msg WSMessage) {
    data, err := json.Marshal(msg)
    if err != nil {
        return
    }

    select {
    case session.send <- data:
    default:
        close(session.send)
        s.mutex.Lock()
        delete(s.clients, session.conn)
        s.mutex.Unlock()
    }
}

func main() {
    server, err := NewWebSocketServer()
    if err != nil {
        log.Fatal("Failed to create WebSocket server:", err)
    }

    http.HandleFunc("/ws", server.handleWebSocket)
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "index.html")
    })

    log.Println("WebSocket server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### Session Management with Redis
```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/go-redis/redis/v8"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
)

// SessionManager handles user sessions with Redis
type SessionManager struct {
    redis *redis.Client
    ttl   time.Duration
}

func NewSessionManager(redisAddr string) *SessionManager {
    rdb := redis.NewClient(&redis.Options{
        Addr: redisAddr,
    })

    return &SessionManager{
        redis: rdb,
        ttl:   24 * time.Hour, // 24 hour session expiry
    }
}

func (sm *SessionManager) SaveSession(ctx context.Context, sessionID string, state *domain.State) error {
    // Serialize state
    data, err := json.Marshal(state)
    if err != nil {
        return err
    }

    // Save to Redis
    key := fmt.Sprintf("session:%s", sessionID)
    return sm.redis.Set(ctx, key, data, sm.ttl).Err()
}

func (sm *SessionManager) LoadSession(ctx context.Context, sessionID string) (*domain.State, error) {
    // Get from Redis
    key := fmt.Sprintf("session:%s", sessionID)
    data, err := sm.redis.Get(ctx, key).Result()
    if err != nil {
        if err == redis.Nil {
            // Session doesn't exist, create new
            return domain.NewState(), nil
        }
        return nil, err
    }

    // Deserialize state
    var state domain.State
    if err := json.Unmarshal([]byte(data), &state); err != nil {
        return nil, err
    }

    return &state, nil
}

func (sm *SessionManager) DeleteSession(ctx context.Context, sessionID string) error {
    key := fmt.Sprintf("session:%s", sessionID)
    return sm.redis.Del(ctx, key).Err()
}

func (sm *SessionManager) ExtendSession(ctx context.Context, sessionID string) error {
    key := fmt.Sprintf("session:%s", sessionID)
    return sm.redis.Expire(ctx, key, sm.ttl).Err()
}
```

---

## Level 3: Production Web Applications
*Build scalable, monitored AI-powered web services*

### Enterprise Web Service
```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "sync"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/agent/events"
    "github.com/lexlapax/go-llms/pkg/agent/workflow"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
)

// Metrics for monitoring
var (
    requestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "ai_requests_total",
            Help: "Total number of AI requests",
        },
        []string{"endpoint", "status"},
    )

    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "ai_request_duration_seconds",
            Help:    "AI request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"endpoint"},
    )

    tokensUsed = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "ai_tokens_used_total",
            Help: "Total number of tokens used",
        },
        []string{"provider", "model"},
    )
)

func init() {
    prometheus.MustRegister(requestsTotal)
    prometheus.MustRegister(requestDuration)
    prometheus.MustRegister(tokensUsed)
}

// AIWebService is a production-ready AI web service
type AIWebService struct {
    agent           *core.LLMAgent
    parallelAgent   *workflow.ParallelAgent
    sessionManager  *SessionManager
    eventBus        *events.EventBus
    rateLimiter     *RateLimiter
    mutex           sync.RWMutex
    config          *Config
}

type Config struct {
    Port              int           `json:"port"`
    MaxConcurrency    int           `json:"max_concurrency"`
    RequestTimeout    time.Duration `json:"request_timeout"`
    RateLimit         int           `json:"rate_limit"` // requests per minute
    MaxTokens         int           `json:"max_tokens"`
    EnableMetrics     bool          `json:"enable_metrics"`
    EnableEvents      bool          `json:"enable_events"`
    RedisAddr         string        `json:"redis_addr"`
}

// Advanced request with features
type AdvancedRequest struct {
    Message     string            `json:"message"`
    UserID      string            `json:"user_id"`
    SessionID   string            `json:"session_id"`
    Context     map[string]interface{} `json:"context,omitempty"`
    MaxTokens   int               `json:"max_tokens,omitempty"`
    Temperature float64           `json:"temperature,omitempty"`
    Tools       []string          `json:"tools,omitempty"`
    Streaming   bool              `json:"streaming,omitempty"`
}

// Advanced response with metrics
type AdvancedResponse struct {
    Reply       string                 `json:"reply"`
    Timestamp   time.Time              `json:"timestamp"`
    TokensUsed  int                    `json:"tokens_used"`
    Duration    float64                `json:"duration_ms"`
    SessionID   string                 `json:"session_id"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

func NewAIWebService(config *Config) (*AIWebService, error) {
    // Create provider with dynamic configuration
    llm, err := provider.NewOpenAI(
        provider.WithModel("gpt-4"),
        provider.WithMaxTokens(config.MaxTokens),
    )
    if err != nil {
        return nil, err
    }

    // Create main agent
    agent := core.NewLLMAgent("web-service-agent", llm)

    // Create parallel agent for concurrent processing
    parallelAgent := workflow.NewParallelAgent("concurrent-processor").
        WithMaxConcurrency(config.MaxConcurrency).
        WithTimeout(config.RequestTimeout).
        WithMergeStrategy(workflow.MergeFirst)

    // Create session manager
    sessionManager := NewSessionManager(config.RedisAddr)

    // Create event bus for real-time updates
    var eventBus *events.EventBus
    if config.EnableEvents {
        eventBus = events.NewEventBus()
    }

    // Create rate limiter
    rateLimiter := NewRateLimiter(config.RateLimit)

    return &AIWebService{
        agent:          agent,
        parallelAgent:  parallelAgent,
        sessionManager: sessionManager,
        eventBus:       eventBus,
        rateLimiter:    rateLimiter,
        config:         config,
    }, nil
}

func (s *AIWebService) chatHandler(c *gin.Context) {
    start := time.Now()
    endpoint := "chat"

    // Rate limiting
    userID := c.GetHeader("X-User-ID")
    if !s.rateLimiter.Allow(userID) {
        requestsTotal.WithLabelValues(endpoint, "rate_limited").Inc()
        c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
        return
    }

    // Parse request
    var req AdvancedRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        requestsTotal.WithLabelValues(endpoint, "bad_request").Inc()
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Load or create session
    ctx := context.Background()
    state, err := s.sessionManager.LoadSession(ctx, req.SessionID)
    if err != nil {
        log.Printf("Session load error: %v", err)
        state = domain.NewState()
    }

    // Update state with request context
    state.Set("user_id", req.UserID)
    state.Set("session_id", req.SessionID)
    for k, v := range req.Context {
        state.Set(k, v)
    }

    // Add user message
    state.AddMessage(domain.NewTextMessage(domain.RoleUser, req.Message))

    // Process with timeout
    reqCtx, cancel := context.WithTimeout(ctx, s.config.RequestTimeout)
    defer cancel()

    result, err := s.agent.Run(reqCtx, state)
    if err != nil {
        requestsTotal.WithLabelValues(endpoint, "error").Inc()
        c.JSON(http.StatusInternalServerError, gin.H{"error": "AI processing failed"})
        return
    }

    // Save session
    if err := s.sessionManager.SaveSession(ctx, req.SessionID, result); err != nil {
        log.Printf("Session save error: %v", err)
    }

    // Extract response
    var reply string
    if len(result.Messages) > 0 {
        reply = result.Messages[len(result.Messages)-1].TextContent()
    }

    // Record metrics
    duration := time.Since(start)
    requestsTotal.WithLabelValues(endpoint, "success").Inc()
    requestDuration.WithLabelValues(endpoint).Observe(duration.Seconds())
    tokensUsed.WithLabelValues("openai", "gpt-4").Add(float64(result.TokensUsed))

    // Publish event
    if s.eventBus != nil {
        s.eventBus.Publish(ctx, domain.Event{
            Type: "chat_response",
            Data: map[string]interface{}{
                "user_id":     req.UserID,
                "session_id":  req.SessionID,
                "tokens_used": result.TokensUsed,
                "duration":    duration.Milliseconds(),
            },
        })
    }

    // Create response
    response := AdvancedResponse{
        Reply:      reply,
        Timestamp:  time.Now(),
        TokensUsed: result.TokensUsed,
        Duration:   float64(duration.Milliseconds()),
        SessionID:  req.SessionID,
        Metadata: map[string]interface{}{
            "model": "gpt-4",
            "provider": "openai",
        },
    }

    c.JSON(http.StatusOK, response)
}

func (s *AIWebService) healthHandler(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status":    "healthy",
        "timestamp": time.Now(),
        "version":   "1.0.0",
    })
}

func (s *AIWebService) metricsHandler() gin.HandlerFunc {
    return gin.WrapH(promhttp.Handler())
}

func main() {
    // Load configuration
    config := &Config{
        Port:           8080,
        MaxConcurrency: 10,
        RequestTimeout: 30 * time.Second,
        RateLimit:      60, // 60 requests per minute
        MaxTokens:      1000,
        EnableMetrics:  true,
        EnableEvents:   true,
        RedisAddr:      "localhost:6379",
    }

    // Create service
    service, err := NewAIWebService(config)
    if err != nil {
        log.Fatal("Failed to create AI web service:", err)
    }

    // Setup Gin
    gin.SetMode(gin.ReleaseMode)
    r := gin.New()

    // Middleware
    r.Use(gin.Logger())
    r.Use(gin.Recovery())
    r.Use(func(c *gin.Context) {
        c.Header("X-Service", "ai-web-service")
        c.Next()
    })

    // Routes
    api := r.Group("/api/v1")
    {
        api.POST("/chat", service.chatHandler)
        api.GET("/health", service.healthHandler)
        
        if config.EnableMetrics {
            api.GET("/metrics", service.metricsHandler())
        }
    }

    // Start server
    log.Printf("AI Web Service starting on port %d", config.Port)
    log.Fatal(r.Run(fmt.Sprintf(":%d", config.Port)))
}
```

### Rate Limiting Implementation
```go
package main

import (
    "sync"
    "time"
)

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
    rate       int               // requests per minute
    buckets    map[string]*Bucket
    mutex      sync.RWMutex
    cleanupDur time.Duration
}

type Bucket struct {
    tokens   int
    lastSeen time.Time
}

func NewRateLimiter(rate int) *RateLimiter {
    rl := &RateLimiter{
        rate:       rate,
        buckets:    make(map[string]*Bucket),
        cleanupDur: 5 * time.Minute,
    }

    // Start cleanup goroutine
    go rl.cleanup()

    return rl
}

func (rl *RateLimiter) Allow(userID string) bool {
    rl.mutex.Lock()
    defer rl.mutex.Unlock()

    now := time.Now()
    bucket, exists := rl.buckets[userID]
    
    if !exists {
        bucket = &Bucket{
            tokens:   rl.rate - 1,
            lastSeen: now,
        }
        rl.buckets[userID] = bucket
        return true
    }

    // Refill tokens based on time passed
    elapsed := now.Sub(bucket.lastSeen)
    tokensToAdd := int(elapsed.Minutes()) * rl.rate
    bucket.tokens += tokensToAdd
    if bucket.tokens > rl.rate {
        bucket.tokens = rl.rate
    }

    bucket.lastSeen = now

    if bucket.tokens > 0 {
        bucket.tokens--
        return true
    }

    return false
}

func (rl *RateLimiter) cleanup() {
    ticker := time.NewTicker(rl.cleanupDur)
    defer ticker.Stop()

    for range ticker.C {
        rl.mutex.Lock()
        now := time.Now()
        for userID, bucket := range rl.buckets {
            if now.Sub(bucket.lastSeen) > rl.cleanupDur {
                delete(rl.buckets, userID)
            }
        }
        rl.mutex.Unlock()
    }
}
```

## Testing Web Applications

### Integration Tests
```go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func setupTestServer(t *testing.T) *gin.Engine {
    gin.SetMode(gin.TestMode)
    
    config := &Config{
        MaxConcurrency: 5,
        RequestTimeout: 10 * time.Second,
        RateLimit:      60,
        MaxTokens:      500,
    }

    service, err := NewAIWebService(config)
    require.NoError(t, err)

    r := gin.New()
    api := r.Group("/api/v1")
    api.POST("/chat", service.chatHandler)
    api.GET("/health", service.healthHandler)

    return r
}

func TestChatEndpoint(t *testing.T) {
    router := setupTestServer(t)

    // Test request
    req := AdvancedRequest{
        Message:   "Hello, how are you?",
        UserID:    "test-user",
        SessionID: "test-session",
    }

    body, err := json.Marshal(req)
    require.NoError(t, err)

    w := httptest.NewRecorder()
    httpReq, _ := http.NewRequest("POST", "/api/v1/chat", bytes.NewBuffer(body))
    httpReq.Header.Set("Content-Type", "application/json")
    
    router.ServeHTTP(w, httpReq)

    assert.Equal(t, http.StatusOK, w.Code)

    var response AdvancedResponse
    err = json.Unmarshal(w.Body.Bytes(), &response)
    require.NoError(t, err)

    assert.NotEmpty(t, response.Reply)
    assert.Equal(t, "test-session", response.SessionID)
    assert.Greater(t, response.TokensUsed, 0)
}

func TestHealthEndpoint(t *testing.T) {
    router := setupTestServer(t)

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/api/v1/health", nil)
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    require.NoError(t, err)

    assert.Equal(t, "healthy", response["status"])
}

func TestRateLimiting(t *testing.T) {
    limiter := NewRateLimiter(5) // 5 requests per minute

    userID := "test-user"

    // First 5 requests should succeed
    for i := 0; i < 5; i++ {
        assert.True(t, limiter.Allow(userID))
    }

    // 6th request should be rate limited
    assert.False(t, limiter.Allow(userID))
}
```

## Performance Considerations

### Optimization Strategies
1. **Connection Pooling** - Reuse HTTP connections and database connections
2. **Caching** - Cache frequent responses and session data
3. **Async Processing** - Use goroutines for non-blocking operations
4. **Load Balancing** - Distribute requests across multiple instances
5. **Circuit Breakers** - Prevent cascade failures

### Monitoring and Observability
- **Metrics** - Track request latency, error rates, token usage
- **Logging** - Structured logging with request tracing
- **Health Checks** - Monitor service health and dependencies
- **Alerting** - Set up alerts for critical metrics

## Security Best Practices

1. **Authentication** - Implement proper user authentication
2. **Authorization** - Control access to AI capabilities
3. **Input Validation** - Sanitize all user inputs
4. **Rate Limiting** - Prevent abuse and DoS attacks
5. **HTTPS** - Encrypt all communications
6. **API Keys** - Secure storage and rotation of provider keys

## Next Steps

- **[APIs and Services](apis-and-services.md)** - Build dedicated LLM-powered APIs
- **[Databases](databases.md)** - Integrate with databases for persistent storage
- **[Production Deployment](../advanced/production-deployment.md)** - Deploy to production environments

---

*Gold Space, this guide provides comprehensive patterns for integrating Go-LLMs with web frameworks. Start with Level 1 for basic integration, then progress to real-time features and production-ready patterns.*