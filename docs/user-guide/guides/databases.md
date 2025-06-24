# Databases: Storing LLM Interactions

> **[Project Root](/) / [Documentation](/docs/) / [User Guide](/docs/user-guide/) / [Guides](/docs/user-guide/guides/) / Databases**

Master persistent storage for LLM interactions, conversation history, and agent state. Learn to integrate with SQL databases, implement caching strategies, and build robust data persistence layers for production AI applications.

## Why Database Integration Matters

- **Persistence** - Store conversations and agent state across sessions
- **History** - Track user interactions and AI responses over time
- **Analytics** - Analyze usage patterns and improve AI performance
- **Compliance** - Meet data retention and audit requirements
- **Scalability** - Handle multiple users and concurrent sessions
- **Recovery** - Restore conversations and state after failures

## Database Architecture

![Database Integration Architecture](../../images/database-architecture.svg)

### Storage Layers
1. **Conversation Storage** - Messages, context, and session data
2. **Agent State** - Persistent memory and workflow status
3. **Event Storage** - Audit logs and interaction tracking
4. **Schema Storage** - Validation rules and data structures
5. **Cache Layer** - Performance optimization and quick access

### Database Types
| Database | Use Case | Benefits | Considerations |
|----------|----------|----------|----------------|
| **PostgreSQL** | Structured data, ACID | Strong consistency, SQL | Complex setup |
| **SQLite** | Embedded apps, development | Simple, file-based | Single writer |
| **MongoDB** | Flexible schemas, JSON | Schema evolution | Eventually consistent |
| **Redis** | Caching, sessions | High performance | Memory limits |
| **DynamoDB** | Serverless, scale | Auto-scaling | AWS lock-in |

## Prerequisites

- [Agent Memory completed](agent-memory.md) ✅
- [APIs and Services understanding](apis-and-services.md) ✅
- Basic database knowledge ✅

---

## Level 1: Basic Database Integration
*Store conversations and basic agent state*

### SQLite Conversation Storage
```go
package main

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "time"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
    _ "github.com/mattn/go-sqlite3"
)

// ConversationStore handles database operations for conversations
type ConversationStore struct {
    db *sql.DB
}

// Conversation represents a stored conversation
type Conversation struct {
    ID        string                 `json:"id" db:"id"`
    UserID    string                 `json:"user_id" db:"user_id"`
    Title     string                 `json:"title" db:"title"`
    Messages  []domain.Message       `json:"messages" db:"messages"`
    State     map[string]interface{} `json:"state" db:"state"`
    CreatedAt time.Time              `json:"created_at" db:"created_at"`
    UpdatedAt time.Time              `json:"updated_at" db:"updated_at"`
}

// Message represents a stored message
type Message struct {
    ID             string                 `json:"id" db:"id"`
    ConversationID string                 `json:"conversation_id" db:"conversation_id"`
    Role           string                 `json:"role" db:"role"`
    Content        string                 `json:"content" db:"content"`
    Metadata       map[string]interface{} `json:"metadata" db:"metadata"`
    TokensUsed     int                    `json:"tokens_used" db:"tokens_used"`
    CreatedAt      time.Time              `json:"created_at" db:"created_at"`
}

// AgentSession combines agent with persistence
type AgentSession struct {
    agent  *core.LLMAgent
    store  *ConversationStore
    convID string
    userID string
}

func NewConversationStore(dbPath string) (*ConversationStore, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, err
    }

    store := &ConversationStore{db: db}
    if err := store.createTables(); err != nil {
        return nil, err
    }

    return store, nil
}

func (cs *ConversationStore) createTables() error {
    queries := []string{
        `CREATE TABLE IF NOT EXISTS conversations (
            id TEXT PRIMARY KEY,
            user_id TEXT NOT NULL,
            title TEXT NOT NULL,
            state TEXT,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )`,
        `CREATE TABLE IF NOT EXISTS messages (
            id TEXT PRIMARY KEY,
            conversation_id TEXT NOT NULL,
            role TEXT NOT NULL,
            content TEXT NOT NULL,
            metadata TEXT,
            tokens_used INTEGER DEFAULT 0,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (conversation_id) REFERENCES conversations(id)
        )`,
        `CREATE INDEX IF NOT EXISTS idx_conversations_user_id ON conversations(user_id)`,
        `CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id)`,
        `CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at)`,
    }

    for _, query := range queries {
        if _, err := cs.db.Exec(query); err != nil {
            return err
        }
    }

    return nil
}

func (cs *ConversationStore) CreateConversation(ctx context.Context, conv *Conversation) error {
    stateJSON, err := json.Marshal(conv.State)
    if err != nil {
        return err
    }

    query := `INSERT INTO conversations (id, user_id, title, state, created_at, updated_at) 
              VALUES (?, ?, ?, ?, ?, ?)`
    
    _, err = cs.db.ExecContext(ctx, query, 
        conv.ID, conv.UserID, conv.Title, string(stateJSON), 
        conv.CreatedAt, conv.UpdatedAt)
    
    return err
}

func (cs *ConversationStore) GetConversation(ctx context.Context, id string) (*Conversation, error) {
    query := `SELECT id, user_id, title, state, created_at, updated_at 
              FROM conversations WHERE id = ?`
    
    row := cs.db.QueryRowContext(ctx, query, id)
    
    var conv Conversation
    var stateJSON string
    
    err := row.Scan(&conv.ID, &conv.UserID, &conv.Title, &stateJSON, 
                   &conv.CreatedAt, &conv.UpdatedAt)
    if err != nil {
        return nil, err
    }

    if err := json.Unmarshal([]byte(stateJSON), &conv.State); err != nil {
        return nil, err
    }

    // Load messages
    messages, err := cs.GetMessages(ctx, id)
    if err != nil {
        return nil, err
    }

    // Convert to domain messages
    conv.Messages = make([]domain.Message, len(messages))
    for i, msg := range messages {
        conv.Messages[i] = domain.NewTextMessage(
            domain.Role(msg.Role), 
            msg.Content,
        )
    }

    return &conv, nil
}

func (cs *ConversationStore) AddMessage(ctx context.Context, msg *Message) error {
    metadataJSON, err := json.Marshal(msg.Metadata)
    if err != nil {
        return err
    }

    query := `INSERT INTO messages (id, conversation_id, role, content, metadata, tokens_used, created_at) 
              VALUES (?, ?, ?, ?, ?, ?, ?)`
    
    _, err = cs.db.ExecContext(ctx, query,
        msg.ID, msg.ConversationID, msg.Role, msg.Content, 
        string(metadataJSON), msg.TokensUsed, msg.CreatedAt)
    
    return err
}

func (cs *ConversationStore) GetMessages(ctx context.Context, conversationID string) ([]Message, error) {
    query := `SELECT id, conversation_id, role, content, metadata, tokens_used, created_at 
              FROM messages WHERE conversation_id = ? ORDER BY created_at ASC`
    
    rows, err := cs.db.QueryContext(ctx, query, conversationID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var messages []Message
    for rows.Next() {
        var msg Message
        var metadataJSON string
        
        err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.Role, 
                        &msg.Content, &metadataJSON, &msg.TokensUsed, &msg.CreatedAt)
        if err != nil {
            return nil, err
        }

        if err := json.Unmarshal([]byte(metadataJSON), &msg.Metadata); err != nil {
            return nil, err
        }

        messages = append(messages, msg)
    }

    return messages, nil
}

func (cs *ConversationStore) UpdateConversationState(ctx context.Context, id string, state map[string]interface{}) error {
    stateJSON, err := json.Marshal(state)
    if err != nil {
        return err
    }

    query := `UPDATE conversations SET state = ?, updated_at = ? WHERE id = ?`
    _, err = cs.db.ExecContext(ctx, query, string(stateJSON), time.Now(), id)
    return err
}

func (cs *ConversationStore) GetUserConversations(ctx context.Context, userID string, limit, offset int) ([]Conversation, error) {
    query := `SELECT id, user_id, title, state, created_at, updated_at 
              FROM conversations WHERE user_id = ? 
              ORDER BY updated_at DESC LIMIT ? OFFSET ?`
    
    rows, err := cs.db.QueryContext(ctx, query, userID, limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var conversations []Conversation
    for rows.Next() {
        var conv Conversation
        var stateJSON string
        
        err := rows.Scan(&conv.ID, &conv.UserID, &conv.Title, &stateJSON, 
                        &conv.CreatedAt, &conv.UpdatedAt)
        if err != nil {
            return nil, err
        }

        if err := json.Unmarshal([]byte(stateJSON), &conv.State); err != nil {
            return nil, err
        }

        conversations = append(conversations, conv)
    }

    return conversations, nil
}

func NewAgentSession(userID string, store *ConversationStore) (*AgentSession, error) {
    // Create LLM provider
    llm, err := provider.NewOpenAI(
        provider.WithModel("gpt-4"),
    )
    if err != nil {
        return nil, err
    }

    // Create agent
    agent := core.NewLLMAgent("persistent-agent", llm)

    // Generate conversation ID
    convID := fmt.Sprintf("conv_%d_%s", time.Now().Unix(), userID)

    return &AgentSession{
        agent:  agent,
        store:  store,
        convID: convID,
        userID: userID,
    }, nil
}

func (as *AgentSession) SendMessage(ctx context.Context, text string) (*domain.State, error) {
    // Load existing conversation or create new one
    conv, err := as.store.GetConversation(ctx, as.convID)
    if err != nil {
        // Create new conversation
        conv = &Conversation{
            ID:        as.convID,
            UserID:    as.userID,
            Title:     "New Conversation",
            State:     make(map[string]interface{}),
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        }
        
        if err := as.store.CreateConversation(ctx, conv); err != nil {
            return nil, err
        }
    }

    // Create state from conversation
    state := domain.NewState()
    
    // Restore state data
    for k, v := range conv.State {
        state.Set(k, v)
    }

    // Add existing messages to state
    for _, msg := range conv.Messages {
        state.AddMessage(msg)
    }

    // Add new user message
    userMessage := domain.NewTextMessage(domain.RoleUser, text)
    state.AddMessage(userMessage)

    // Store user message
    msgID := fmt.Sprintf("msg_%d", time.Now().UnixNano())
    userMsg := &Message{
        ID:             msgID,
        ConversationID: as.convID,
        Role:           string(domain.RoleUser),
        Content:        text,
        Metadata:       make(map[string]interface{}),
        CreatedAt:      time.Now(),
    }
    
    if err := as.store.AddMessage(ctx, userMsg); err != nil {
        return nil, err
    }

    // Run agent
    result, err := as.agent.Run(ctx, state)
    if err != nil {
        return nil, err
    }

    // Store AI response
    if len(result.Messages) > 0 {
        lastMessage := result.Messages[len(result.Messages)-1]
        aiMsgID := fmt.Sprintf("msg_%d", time.Now().UnixNano())
        
        aiMsg := &Message{
            ID:             aiMsgID,
            ConversationID: as.convID,
            Role:           string(domain.RoleAssistant),
            Content:        lastMessage.TextContent(),
            Metadata: map[string]interface{}{
                "model": "gpt-4",
                "provider": "openai",
            },
            TokensUsed: result.TokensUsed,
            CreatedAt:  time.Now(),
        }
        
        if err := as.store.AddMessage(ctx, aiMsg); err != nil {
            return nil, err
        }
    }

    // Update conversation state
    stateData := make(map[string]interface{})
    for k, v := range result.Data {
        stateData[k] = v
    }
    
    if err := as.store.UpdateConversationState(ctx, as.convID, stateData); err != nil {
        return nil, err
    }

    return result, nil
}

func main() {
    // Initialize database
    store, err := NewConversationStore("conversations.db")
    if err != nil {
        log.Fatal("Failed to create conversation store:", err)
    }

    // Create agent session
    session, err := NewAgentSession("user123", store)
    if err != nil {
        log.Fatal("Failed to create agent session:", err)
    }

    ctx := context.Background()

    // Example conversation
    response1, err := session.SendMessage(ctx, "Hello, can you help me with some analysis?")
    if err != nil {
        log.Fatal("Failed to send message:", err)
    }

    fmt.Printf("AI Response: %s\n", response1.Messages[len(response1.Messages)-1].TextContent())

    // Continue conversation
    response2, err := session.SendMessage(ctx, "What kinds of analysis can you perform?")
    if err != nil {
        log.Fatal("Failed to send message:", err)
    }

    fmt.Printf("AI Response: %s\n", response2.Messages[len(response2.Messages)-1].TextContent())

    // List user conversations
    conversations, err := store.GetUserConversations(ctx, "user123", 10, 0)
    if err != nil {
        log.Fatal("Failed to get conversations:", err)
    }

    fmt.Printf("User has %d conversations\n", len(conversations))
}
```

---

## Level 2: Advanced Database Features
*Implement caching, indexing, and performance optimization*

### PostgreSQL with Advanced Features
```go
package main

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "time"

    "github.com/lib/pq"
    "github.com/jmoiron/sqlx"
    "github.com/go-redis/redis/v8"
    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/agent/events"
)

// AdvancedConversationStore with PostgreSQL and Redis caching
type AdvancedConversationStore struct {
    db    *sqlx.DB
    redis *redis.Client
    cache *ConversationCache
}

type ConversationCache struct {
    redis *redis.Client
    ttl   time.Duration
}

// Enhanced conversation with vector embeddings
type EnhancedConversation struct {
    ID          string                 `json:"id" db:"id"`
    UserID      string                 `json:"user_id" db:"user_id"`
    Title       string                 `json:"title" db:"title"`
    Summary     string                 `json:"summary" db:"summary"`
    State       map[string]interface{} `json:"state" db:"state"`
    Tags        []string               `json:"tags" db:"tags"`
    Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
    CreatedAt   time.Time              `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
    LastActivity time.Time             `json:"last_activity" db:"last_activity"`
}

// Enhanced message with embeddings and analytics
type EnhancedMessage struct {
    ID             string                 `json:"id" db:"id"`
    ConversationID string                 `json:"conversation_id" db:"conversation_id"`
    Role           string                 `json:"role" db:"role"`
    Content        string                 `json:"content" db:"content"`
    ContentType    string                 `json:"content_type" db:"content_type"`
    Metadata       map[string]interface{} `json:"metadata" db:"metadata"`
    TokensUsed     int                    `json:"tokens_used" db:"tokens_used"`
    ProcessingTime int64                  `json:"processing_time" db:"processing_time"` // milliseconds
    Provider       string                 `json:"provider" db:"provider"`
    Model          string                 `json:"model" db:"model"`
    Embedding      []float64              `json:"embedding" db:"embedding"`
    CreatedAt      time.Time              `json:"created_at" db:"created_at"`
}

// Analytics and reporting
type ConversationAnalytics struct {
    UserID         string    `json:"user_id"`
    TotalMessages  int       `json:"total_messages"`
    TotalTokens    int       `json:"total_tokens"`
    AvgResponseTime float64  `json:"avg_response_time"`
    TopicDistribution map[string]int `json:"topic_distribution"`
    LastActivity   time.Time `json:"last_activity"`
}

func NewAdvancedConversationStore(dbURL, redisAddr string) (*AdvancedConversationStore, error) {
    // Connect to PostgreSQL
    db, err := sqlx.Connect("postgres", dbURL)
    if err != nil {
        return nil, err
    }

    // Connect to Redis
    rdb := redis.NewClient(&redis.Options{
        Addr: redisAddr,
    })

    cache := &ConversationCache{
        redis: rdb,
        ttl:   30 * time.Minute,
    }

    store := &AdvancedConversationStore{
        db:    db,
        redis: rdb,
        cache: cache,
    }

    if err := store.createTables(); err != nil {
        return nil, err
    }

    return store, nil
}

func (acs *AdvancedConversationStore) createTables() error {
    queries := []string{
        `CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,
        `CREATE EXTENSION IF NOT EXISTS "vector"`, // For embeddings
        
        `CREATE TABLE IF NOT EXISTS conversations (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            user_id VARCHAR(255) NOT NULL,
            title TEXT NOT NULL,
            summary TEXT,
            state JSONB,
            tags TEXT[],
            metadata JSONB,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            last_activity TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        )`,
        
        `CREATE TABLE IF NOT EXISTS messages (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
            role VARCHAR(50) NOT NULL,
            content TEXT NOT NULL,
            content_type VARCHAR(50) DEFAULT 'text',
            metadata JSONB,
            tokens_used INTEGER DEFAULT 0,
            processing_time BIGINT DEFAULT 0,
            provider VARCHAR(100),
            model VARCHAR(100),
            embedding VECTOR(1536), -- OpenAI embedding dimension
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        )`,
        
        `CREATE TABLE IF NOT EXISTS conversation_events (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
            event_type VARCHAR(100) NOT NULL,
            event_data JSONB,
            user_id VARCHAR(255),
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        )`,
        
        // Indexes for performance
        `CREATE INDEX IF NOT EXISTS idx_conversations_user_id ON conversations(user_id)`,
        `CREATE INDEX IF NOT EXISTS idx_conversations_last_activity ON conversations(last_activity DESC)`,
        `CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id)`,
        `CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at DESC)`,
        `CREATE INDEX IF NOT EXISTS idx_messages_embedding ON messages USING ivfflat (embedding vector_cosine_ops)`,
        `CREATE INDEX IF NOT EXISTS idx_events_conversation_id ON conversation_events(conversation_id)`,
        `CREATE INDEX IF NOT EXISTS idx_events_type_created ON conversation_events(event_type, created_at DESC)`,
        
        // Full-text search
        `CREATE INDEX IF NOT EXISTS idx_messages_content_search ON messages USING gin(to_tsvector('english', content))`,
    }

    for _, query := range queries {
        if _, err := acs.db.Exec(query); err != nil {
            log.Printf("Warning: Failed to execute query: %s, Error: %v", query, err)
            // Continue with other queries - some extensions might not be available
        }
    }

    return nil
}

func (acs *AdvancedConversationStore) CreateConversationWithCache(ctx context.Context, conv *EnhancedConversation) error {
    // Insert into database
    query := `INSERT INTO conversations (id, user_id, title, summary, state, tags, metadata, created_at, updated_at, last_activity) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
    
    stateJSON, _ := json.Marshal(conv.State)
    metadataJSON, _ := json.Marshal(conv.Metadata)
    
    _, err := acs.db.ExecContext(ctx, query,
        conv.ID, conv.UserID, conv.Title, conv.Summary,
        stateJSON, pq.Array(conv.Tags), metadataJSON,
        conv.CreatedAt, conv.UpdatedAt, conv.LastActivity)
    
    if err != nil {
        return err
    }

    // Cache the conversation
    return acs.cache.Set(ctx, conv.ID, conv)
}

func (acs *AdvancedConversationStore) GetConversationWithCache(ctx context.Context, id string) (*EnhancedConversation, error) {
    // Try cache first
    if conv, err := acs.cache.Get(ctx, id); err == nil {
        return conv, nil
    }

    // Fallback to database
    query := `SELECT id, user_id, title, summary, state, tags, metadata, created_at, updated_at, last_activity 
              FROM conversations WHERE id = $1`
    
    row := acs.db.QueryRowContext(ctx, query, id)
    
    var conv EnhancedConversation
    var stateJSON, metadataJSON []byte
    var tags pq.StringArray
    
    err := row.Scan(&conv.ID, &conv.UserID, &conv.Title, &conv.Summary,
                   &stateJSON, &tags, &metadataJSON,
                   &conv.CreatedAt, &conv.UpdatedAt, &conv.LastActivity)
    if err != nil {
        return nil, err
    }

    json.Unmarshal(stateJSON, &conv.State)
    json.Unmarshal(metadataJSON, &conv.Metadata)
    conv.Tags = []string(tags)

    // Cache for future use
    acs.cache.Set(ctx, id, &conv)

    return &conv, nil
}

func (acs *AdvancedConversationStore) AddMessageWithEmbedding(ctx context.Context, msg *EnhancedMessage) error {
    query := `INSERT INTO messages 
              (id, conversation_id, role, content, content_type, metadata, tokens_used, 
               processing_time, provider, model, embedding, created_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
    
    metadataJSON, _ := json.Marshal(msg.Metadata)
    
    _, err := acs.db.ExecContext(ctx, query,
        msg.ID, msg.ConversationID, msg.Role, msg.Content, msg.ContentType,
        metadataJSON, msg.TokensUsed, msg.ProcessingTime,
        msg.Provider, msg.Model, pq.Array(msg.Embedding), msg.CreatedAt)
    
    if err != nil {
        return err
    }

    // Update conversation last activity
    _, err = acs.db.ExecContext(ctx,
        `UPDATE conversations SET last_activity = $1 WHERE id = $2`,
        time.Now(), msg.ConversationID)
    
    // Invalidate cache
    acs.cache.Delete(ctx, msg.ConversationID)
    
    return err
}

func (acs *AdvancedConversationStore) SearchSimilarMessages(ctx context.Context, embedding []float64, limit int) ([]EnhancedMessage, error) {
    query := `SELECT id, conversation_id, role, content, content_type, metadata, 
                     tokens_used, processing_time, provider, model, embedding, created_at,
                     1 - (embedding <=> $1) as similarity
              FROM messages 
              WHERE embedding IS NOT NULL
              ORDER BY embedding <=> $1
              LIMIT $2`
    
    rows, err := acs.db.QueryContext(ctx, query, pq.Array(embedding), limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var messages []EnhancedMessage
    for rows.Next() {
        var msg EnhancedMessage
        var metadataJSON []byte
        var embeddingArray pq.Float64Array
        var similarity float64
        
        err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.Role, &msg.Content,
                        &msg.ContentType, &metadataJSON, &msg.TokensUsed,
                        &msg.ProcessingTime, &msg.Provider, &msg.Model,
                        &embeddingArray, &msg.CreatedAt, &similarity)
        if err != nil {
            return nil, err
        }

        json.Unmarshal(metadataJSON, &msg.Metadata)
        msg.Embedding = []float64(embeddingArray)
        
        // Add similarity to metadata
        if msg.Metadata == nil {
            msg.Metadata = make(map[string]interface{})
        }
        msg.Metadata["similarity"] = similarity

        messages = append(messages, msg)
    }

    return messages, nil
}

func (acs *AdvancedConversationStore) GetConversationAnalytics(ctx context.Context, userID string, days int) (*ConversationAnalytics, error) {
    query := `SELECT 
                COUNT(m.id) as total_messages,
                COALESCE(SUM(m.tokens_used), 0) as total_tokens,
                COALESCE(AVG(m.processing_time), 0) as avg_response_time,
                MAX(c.last_activity) as last_activity
              FROM conversations c
              LEFT JOIN messages m ON c.id = m.conversation_id
              WHERE c.user_id = $1 
                AND c.created_at >= NOW() - INTERVAL '%d days'
              GROUP BY c.user_id`
    
    row := acs.db.QueryRowContext(ctx, fmt.Sprintf(query, days), userID)
    
    var analytics ConversationAnalytics
    err := row.Scan(&analytics.TotalMessages, &analytics.TotalTokens,
                   &analytics.AvgResponseTime, &analytics.LastActivity)
    if err != nil {
        return nil, err
    }

    analytics.UserID = userID

    // Get topic distribution (simplified)
    topicQuery := `SELECT tags, COUNT(*) as count
                   FROM conversations
                   WHERE user_id = $1 AND tags IS NOT NULL
                   GROUP BY tags`
    
    rows, err := acs.db.QueryContext(ctx, topicQuery, userID)
    if err == nil {
        defer rows.Close()
        
        analytics.TopicDistribution = make(map[string]int)
        for rows.Next() {
            var tags pq.StringArray
            var count int
            
            if err := rows.Scan(&tags, &count); err == nil {
                for _, tag := range tags {
                    analytics.TopicDistribution[tag] += count
                }
            }
        }
    }

    return &analytics, nil
}

// Cache implementation
func (cc *ConversationCache) Set(ctx context.Context, key string, conv *EnhancedConversation) error {
    data, err := json.Marshal(conv)
    if err != nil {
        return err
    }
    
    return cc.redis.Set(ctx, "conv:"+key, data, cc.ttl).Err()
}

func (cc *ConversationCache) Get(ctx context.Context, key string) (*EnhancedConversation, error) {
    data, err := cc.redis.Get(ctx, "conv:"+key).Result()
    if err != nil {
        return nil, err
    }
    
    var conv EnhancedConversation
    if err := json.Unmarshal([]byte(data), &conv); err != nil {
        return nil, err
    }
    
    return &conv, nil
}

func (cc *ConversationCache) Delete(ctx context.Context, key string) error {
    return cc.redis.Del(ctx, "conv:"+key).Err()
}

// Event logging for audit trail
func (acs *AdvancedConversationStore) LogEvent(ctx context.Context, conversationID, eventType, userID string, data map[string]interface{}) error {
    query := `INSERT INTO conversation_events (conversation_id, event_type, event_data, user_id, created_at)
              VALUES ($1, $2, $3, $4, $5)`
    
    eventData, _ := json.Marshal(data)
    _, err := acs.db.ExecContext(ctx, query, conversationID, eventType, eventData, userID, time.Now())
    
    return err
}
```

---

## Level 3: Production Database Systems
*Build enterprise-grade persistence with monitoring and scalability*

### Enterprise Database Layer
```go
package main

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/jmoiron/sqlx"
    "github.com/prometheus/client_golang/prometheus"
    "go.uber.org/zap"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/agent/events"
)

// Metrics for monitoring
var (
    dbOperations = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "db_operations_total",
            Help: "Total database operations",
        },
        []string{"operation", "table", "status"},
    )

    dbDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "db_operation_duration_seconds",
            Help: "Database operation duration",
        },
        []string{"operation", "table"},
    )

    dbConnections = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "db_connections",
            Help: "Database connections",
        },
        []string{"status"}, // active, idle, total
    )
)

func init() {
    prometheus.MustRegister(dbOperations)
    prometheus.MustRegister(dbDuration)
    prometheus.MustRegister(dbConnections)
}

// DatabaseManager handles multiple database connections and operations
type DatabaseManager struct {
    primary   *sqlx.DB
    replica   *sqlx.DB
    cache     CacheInterface
    logger    *zap.Logger
    config    *DatabaseConfig
    metrics   *DatabaseMetrics
    pool      *ConnectionPool
}

type DatabaseConfig struct {
    PrimaryURL     string        `json:"primary_url"`
    ReplicaURL     string        `json:"replica_url"`
    MaxConnections int           `json:"max_connections"`
    MaxIdle        int           `json:"max_idle"`
    ConnLifetime   time.Duration `json:"conn_lifetime"`
    QueryTimeout   time.Duration `json:"query_timeout"`
    RetryAttempts  int           `json:"retry_attempts"`
    EnableMetrics  bool          `json:"enable_metrics"`
}

type DatabaseMetrics struct {
    mu            sync.RWMutex
    Operations    map[string]int64
    AvgDuration   map[string]time.Duration
    ErrorCount    int64
    LastError     time.Time
}

type ConnectionPool struct {
    primary *ConnectionStats
    replica *ConnectionStats
}

type ConnectionStats struct {
    Active int
    Idle   int
    Total  int
}

// CacheInterface for multiple cache backends
type CacheInterface interface {
    Get(ctx context.Context, key string) ([]byte, error)
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Clear(ctx context.Context, pattern string) error
}

// Repository pattern for clean data access
type ConversationRepository struct {
    db      *DatabaseManager
    cache   CacheInterface
    logger  *zap.Logger
}

type MessageRepository struct {
    db      *DatabaseManager
    cache   CacheInterface
    logger  *zap.Logger
}

type AnalyticsRepository struct {
    db      *DatabaseManager
    logger  *zap.Logger
}

func NewDatabaseManager(config *DatabaseConfig, logger *zap.Logger) (*DatabaseManager, error) {
    // Connect to primary database
    primary, err := sqlx.Connect("postgres", config.PrimaryURL)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to primary database: %w", err)
    }

    // Configure connection pool
    primary.SetMaxOpenConns(config.MaxConnections)
    primary.SetMaxIdleConns(config.MaxIdle)
    primary.SetConnMaxLifetime(config.ConnLifetime)

    // Connect to replica database (optional)
    var replica *sqlx.DB
    if config.ReplicaURL != "" {
        replica, err = sqlx.Connect("postgres", config.ReplicaURL)
        if err != nil {
            logger.Warn("Failed to connect to replica database", zap.Error(err))
        } else {
            replica.SetMaxOpenConns(config.MaxConnections / 2)
            replica.SetMaxIdleConns(config.MaxIdle / 2)
            replica.SetConnMaxLifetime(config.ConnLifetime)
        }
    }

    dm := &DatabaseManager{
        primary: primary,
        replica: replica,
        logger:  logger,
        config:  config,
        metrics: &DatabaseMetrics{
            Operations:  make(map[string]int64),
            AvgDuration: make(map[string]time.Duration),
        },
        pool: &ConnectionPool{
            primary: &ConnectionStats{},
            replica: &ConnectionStats{},
        },
    }

    // Start metrics collection
    if config.EnableMetrics {
        go dm.collectMetrics()
    }

    return dm, nil
}

func (dm *DatabaseManager) collectMetrics() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        // Collect primary DB stats
        if stats := dm.primary.Stats(); stats != nil {
            dbConnections.WithLabelValues("active").Set(float64(stats.InUse))
            dbConnections.WithLabelValues("idle").Set(float64(stats.Idle))
            dbConnections.WithLabelValues("total").Set(float64(stats.OpenConnections))
        }

        // Collect replica DB stats
        if dm.replica != nil {
            if stats := dm.replica.Stats(); stats != nil {
                // Additional replica metrics could go here
            }
        }
    }
}

func (dm *DatabaseManager) executeWithMetrics(ctx context.Context, operation, table string, fn func() error) error {
    start := time.Now()
    
    err := fn()
    
    duration := time.Since(start)
    
    // Record metrics
    status := "success"
    if err != nil {
        status = "error"
        dm.metrics.mu.Lock()
        dm.metrics.ErrorCount++
        dm.metrics.LastError = time.Now()
        dm.metrics.mu.Unlock()
    }
    
    dbOperations.WithLabelValues(operation, table, status).Inc()
    dbDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
    
    // Update internal metrics
    dm.metrics.mu.Lock()
    dm.metrics.Operations[operation]++
    dm.metrics.AvgDuration[operation] = (dm.metrics.AvgDuration[operation] + duration) / 2
    dm.metrics.mu.Unlock()
    
    return err
}

func (dm *DatabaseManager) Query(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
    // Use replica for read operations if available
    db := dm.primary
    if dm.replica != nil && isReadQuery(query) {
        db = dm.replica
    }

    var rows *sqlx.Rows
    var err error
    
    err = dm.executeWithMetrics(ctx, "query", extractTable(query), func() error {
        ctx, cancel := context.WithTimeout(ctx, dm.config.QueryTimeout)
        defer cancel()
        
        rows, err = db.QueryxContext(ctx, query, args...)
        return err
    })

    return rows, err
}

func (dm *DatabaseManager) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
    // Use primary for write operations
    var result sql.Result
    var err error
    
    err = dm.executeWithMetrics(ctx, "exec", extractTable(query), func() error {
        ctx, cancel := context.WithTimeout(ctx, dm.config.QueryTimeout)
        defer cancel()
        
        result, err = dm.primary.ExecContext(ctx, query, args...)
        return err
    })

    return result, err
}

func (dm *DatabaseManager) Transaction(ctx context.Context, fn func(*sqlx.Tx) error) error {
    return dm.executeWithMetrics(ctx, "transaction", "multiple", func() error {
        tx, err := dm.primary.BeginTxx(ctx, nil)
        if err != nil {
            return err
        }

        if err := fn(tx); err != nil {
            tx.Rollback()
            return err
        }

        return tx.Commit()
    })
}

// Repository implementations
func NewConversationRepository(dm *DatabaseManager, cache CacheInterface, logger *zap.Logger) *ConversationRepository {
    return &ConversationRepository{
        db:     dm,
        cache:  cache,
        logger: logger,
    }
}

func (cr *ConversationRepository) Create(ctx context.Context, conv *EnhancedConversation) error {
    return cr.db.Transaction(ctx, func(tx *sqlx.Tx) error {
        query := `INSERT INTO conversations 
                  (id, user_id, title, summary, state, tags, metadata, created_at, updated_at, last_activity) 
                  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
        
        stateJSON, _ := json.Marshal(conv.State)
        metadataJSON, _ := json.Marshal(conv.Metadata)
        
        _, err := tx.ExecContext(ctx, query,
            conv.ID, conv.UserID, conv.Title, conv.Summary,
            stateJSON, pq.Array(conv.Tags), metadataJSON,
            conv.CreatedAt, conv.UpdatedAt, conv.LastActivity)
        
        if err != nil {
            cr.logger.Error("Failed to create conversation", 
                zap.String("id", conv.ID), zap.Error(err))
            return err
        }

        // Cache the conversation
        if cr.cache != nil {
            convData, _ := json.Marshal(conv)
            cr.cache.Set(ctx, "conv:"+conv.ID, convData, 30*time.Minute)
        }

        return nil
    })
}

func (cr *ConversationRepository) GetByID(ctx context.Context, id string) (*EnhancedConversation, error) {
    // Try cache first
    if cr.cache != nil {
        if data, err := cr.cache.Get(ctx, "conv:"+id); err == nil {
            var conv EnhancedConversation
            if json.Unmarshal(data, &conv) == nil {
                return &conv, nil
            }
        }
    }

    // Fallback to database
    query := `SELECT id, user_id, title, summary, state, tags, metadata, 
                     created_at, updated_at, last_activity 
              FROM conversations WHERE id = $1`
    
    rows, err := cr.db.Query(ctx, query, id)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    if !rows.Next() {
        return nil, sql.ErrNoRows
    }

    var conv EnhancedConversation
    var stateJSON, metadataJSON []byte
    var tags pq.StringArray
    
    err = rows.Scan(&conv.ID, &conv.UserID, &conv.Title, &conv.Summary,
                   &stateJSON, &tags, &metadataJSON,
                   &conv.CreatedAt, &conv.UpdatedAt, &conv.LastActivity)
    if err != nil {
        return nil, err
    }

    json.Unmarshal(stateJSON, &conv.State)
    json.Unmarshal(metadataJSON, &conv.Metadata)
    conv.Tags = []string(tags)

    // Cache for future use
    if cr.cache != nil {
        convData, _ := json.Marshal(conv)
        cr.cache.Set(ctx, "conv:"+id, convData, 30*time.Minute)
    }

    return &conv, nil
}

func (cr *ConversationRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]EnhancedConversation, error) {
    query := `SELECT id, user_id, title, summary, state, tags, metadata, 
                     created_at, updated_at, last_activity 
              FROM conversations 
              WHERE user_id = $1 
              ORDER BY last_activity DESC 
              LIMIT $2 OFFSET $3`
    
    rows, err := cr.db.Query(ctx, query, userID, limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var conversations []EnhancedConversation
    for rows.Next() {
        var conv EnhancedConversation
        var stateJSON, metadataJSON []byte
        var tags pq.StringArray
        
        err := rows.Scan(&conv.ID, &conv.UserID, &conv.Title, &conv.Summary,
                        &stateJSON, &tags, &metadataJSON,
                        &conv.CreatedAt, &conv.UpdatedAt, &conv.LastActivity)
        if err != nil {
            cr.logger.Error("Failed to scan conversation", zap.Error(err))
            continue
        }

        json.Unmarshal(stateJSON, &conv.State)
        json.Unmarshal(metadataJSON, &conv.Metadata)
        conv.Tags = []string(tags)

        conversations = append(conversations, conv)
    }

    return conversations, nil
}

// Health check for the database system
func (dm *DatabaseManager) HealthCheck(ctx context.Context) error {
    // Check primary database
    if err := dm.primary.PingContext(ctx); err != nil {
        return fmt.Errorf("primary database unhealthy: %w", err)
    }

    // Check replica database if configured
    if dm.replica != nil {
        if err := dm.replica.PingContext(ctx); err != nil {
            dm.logger.Warn("Replica database unhealthy", zap.Error(err))
            // Don't fail health check for replica issues
        }
    }

    return nil
}

// Utility functions
func isReadQuery(query string) bool {
    query = strings.ToUpper(strings.TrimSpace(query))
    return strings.HasPrefix(query, "SELECT") || 
           strings.HasPrefix(query, "WITH")
}

func extractTable(query string) string {
    // Simple table extraction - could be enhanced
    words := strings.Fields(strings.ToUpper(query))
    for i, word := range words {
        if (word == "FROM" || word == "UPDATE" || word == "INSERT") && i+1 < len(words) {
            return strings.ToLower(words[i+1])
        }
    }
    return "unknown"
}

func main() {
    // Initialize logger
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    // Database configuration
    config := &DatabaseConfig{
        PrimaryURL:     "postgres://user:pass@localhost:5432/llm_db?sslmode=disable",
        ReplicaURL:     "postgres://user:pass@replica:5432/llm_db?sslmode=disable",
        MaxConnections: 25,
        MaxIdle:        5,
        ConnLifetime:   30 * time.Minute,
        QueryTimeout:   30 * time.Second,
        RetryAttempts:  3,
        EnableMetrics:  true,
    }

    // Create database manager
    dbManager, err := NewDatabaseManager(config, logger)
    if err != nil {
        log.Fatal("Failed to create database manager:", err)
    }

    // Create repositories
    conversationRepo := NewConversationRepository(dbManager, nil, logger)

    // Example usage
    ctx := context.Background()
    
    conv := &EnhancedConversation{
        ID:        "conv-123",
        UserID:    "user-456",
        Title:     "Test Conversation",
        Summary:   "A test conversation",
        State:     make(map[string]interface{}),
        Tags:      []string{"test", "example"},
        Metadata:  make(map[string]interface{}),
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
        LastActivity: time.Now(),
    }

    if err := conversationRepo.Create(ctx, conv); err != nil {
        logger.Error("Failed to create conversation", zap.Error(err))
    }

    // Health check
    if err := dbManager.HealthCheck(ctx); err != nil {
        logger.Error("Database health check failed", zap.Error(err))
    } else {
        logger.Info("Database health check passed")
    }
}
```

## Performance Considerations

### Optimization Strategies
1. **Connection Pooling** - Manage database connections efficiently
2. **Query Optimization** - Use indexes and query analysis
3. **Caching** - Cache frequently accessed data
4. **Read Replicas** - Distribute read operations
5. **Partitioning** - Partition large tables by date or user

### Monitoring and Alerting
- **Metrics Collection** - Track query performance and errors
- **Health Checks** - Monitor database connectivity
- **Alert Thresholds** - Set up alerts for performance degradation
- **Dashboard** - Visualize database metrics

## Security Best Practices

1. **Access Control** - Use proper database permissions
2. **Encryption** - Encrypt data at rest and in transit
3. **SQL Injection** - Use parameterized queries
4. **Audit Logging** - Track database operations
5. **Backup Strategy** - Regular backups and recovery testing

## Next Steps

- **[Existing Systems](existing-systems.md)** - Integrate with legacy systems
- **[Production Deployment](../advanced/production-deployment.md)** - Deploy database systems to production
- **[Performance Optimization](../advanced/performance-optimization.md)** - Advanced optimization techniques

---

*Gold Space, this comprehensive database guide covers everything from basic SQLite storage to enterprise PostgreSQL systems with caching, monitoring, and high availability. Start with Level 1 for simple persistence needs, then scale up to production-ready systems.*