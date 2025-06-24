# Customer Support: Complete Support System

> **[Project Root](/) / [Documentation](/docs/) / [User Guide](/docs/user-guide/) / [Examples](/docs/user-guide/examples/) / Customer Support**

Build a complete AI-powered customer support system with multi-channel communication, intelligent routing, knowledge base integration, and escalation management. This example demonstrates real-world application of multiple Go-LLMs capabilities in a production-ready system.

## System Overview

This customer support system provides:

- **Multi-channel Support** - Email, chat, API integration
- **Intelligent Ticket Routing** - AI-powered classification and assignment
- **Knowledge Base Integration** - Automatic answer generation from documentation
- **Escalation Management** - Smart escalation to human agents
- **Analytics and Reporting** - Comprehensive support metrics
- **Customer History** - Persistent conversation tracking

## Architecture

![Customer Support System Architecture](../../images/customer-support-architecture.svg)

### Components
1. **Intake Handler** - Processes incoming support requests
2. **Classification Agent** - Categorizes and prioritizes tickets
3. **Knowledge Agent** - Searches and retrieves relevant information
4. **Response Agent** - Generates contextual responses
5. **Escalation Manager** - Handles complex cases requiring human intervention
6. **Analytics Engine** - Tracks metrics and generates insights

---

## Complete Implementation

```go
package main

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strconv"
    "strings"
    "sync"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    
    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
    "github.com/lexlapax/go-llms/pkg/agent/workflow"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
    "github.com/lexlapax/go-llms/pkg/structured/processor"
    schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// CustomerSupportSystem is the main system orchestrator
type CustomerSupportSystem struct {
    db                  *sqlx.DB
    intakeAgent         *core.LLMAgent
    classificationAgent *core.LLMAgent
    knowledgeAgent      *core.LLMAgent
    responseAgent       *core.LLMAgent
    escalationManager   *EscalationManager
    knowledgeBase       *KnowledgeBase
    analytics          *AnalyticsEngine
    config             *SupportConfig
    metrics            *SupportMetrics
    mu                 sync.RWMutex
}

type SupportConfig struct {
    DatabaseURL           string        `json:"database_url"`
    OpenAIKey            string        `json:"openai_key"`
    MaxResponseTime      time.Duration `json:"max_response_time"`
    EscalationThreshold  float64       `json:"escalation_threshold"`
    KnowledgeBaseURL     string        `json:"knowledge_base_url"`
    EmailConfig          EmailConfig   `json:"email_config"`
    SlackConfig          SlackConfig   `json:"slack_config"`
}

type EmailConfig struct {
    SMTPHost     string `json:"smtp_host"`
    SMTPPort     int    `json:"smtp_port"`
    Username     string `json:"username"`
    Password     string `json:"password"`
    FromAddress  string `json:"from_address"`
}

type SlackConfig struct {
    WebhookURL   string `json:"webhook_url"`
    Channel      string `json:"channel"`
    BotToken     string `json:"bot_token"`
}

// Support ticket models
type SupportTicket struct {
    ID              int                    `json:"id" db:"id"`
    CustomerID      string                 `json:"customer_id" db:"customer_id"`
    Subject         string                 `json:"subject" db:"subject"`
    Description     string                 `json:"description" db:"description"`
    Status          string                 `json:"status" db:"status"` // open, in_progress, resolved, escalated
    Priority        string                 `json:"priority" db:"priority"` // low, medium, high, urgent
    Category        string                 `json:"category" db:"category"`
    Channel         string                 `json:"channel" db:"channel"` // email, chat, api
    AssignedAgent   *string                `json:"assigned_agent" db:"assigned_agent"`
    CustomerEmail   string                 `json:"customer_email" db:"customer_email"`
    CustomerName    string                 `json:"customer_name" db:"customer_name"`
    CreatedAt       time.Time              `json:"created_at" db:"created_at"`
    UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
    ResolvedAt      *time.Time             `json:"resolved_at" db:"resolved_at"`
    FirstResponse   *time.Time             `json:"first_response" db:"first_response"`
    Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
    SentimentScore  float64                `json:"sentiment_score" db:"sentiment_score"`
    ComplexityScore float64                `json:"complexity_score" db:"complexity_score"`
}

type TicketMessage struct {
    ID        int                    `json:"id" db:"id"`
    TicketID  int                    `json:"ticket_id" db:"ticket_id"`
    Sender    string                 `json:"sender" db:"sender"` // customer, agent, system
    Content   string                 `json:"content" db:"content"`
    MessageType string               `json:"message_type" db:"message_type"` // text, attachment, auto_response
    Metadata  map[string]interface{} `json:"metadata" db:"metadata"`
    CreatedAt time.Time              `json:"created_at" db:"created_at"`
}

type KnowledgeBaseEntry struct {
    ID          int      `json:"id" db:"id"`
    Title       string   `json:"title" db:"title"`
    Content     string   `json:"content" db:"content"`
    Category    string   `json:"category" db:"category"`
    Tags        []string `json:"tags" db:"tags"`
    Confidence  float64  `json:"confidence" db:"confidence"`
    UseCount    int      `json:"use_count" db:"use_count"`
    LastUpdated time.Time `json:"last_updated" db:"last_updated"`
}

// Classification result
type TicketClassification struct {
    Category        string  `json:"category"`
    Priority        string  `json:"priority"`
    Urgency         string  `json:"urgency"`
    Sentiment       float64 `json:"sentiment"`
    Complexity      float64 `json:"complexity"`
    RequiresHuman   bool    `json:"requires_human"`
    SuggestedAgent  string  `json:"suggested_agent,omitempty"`
    Confidence      float64 `json:"confidence"`
    Reasoning       string  `json:"reasoning"`
}

// AI Response
type SupportResponse struct {
    Content            string   `json:"content"`
    Confidence         float64  `json:"confidence"`
    KnowledgeUsed      []int    `json:"knowledge_used"`
    SuggestedActions   []string `json:"suggested_actions"`
    RequiresEscalation bool     `json:"requires_escalation"`
    FollowUpRequired   bool     `json:"follow_up_required"`
    EstimatedResolution string  `json:"estimated_resolution"`
}

// Metrics
type SupportMetrics struct {
    TicketsCreated         prometheus.Counter
    TicketsResolved        prometheus.Counter
    TicketsEscalated       prometheus.Counter
    ResponseTime           prometheus.Histogram
    FirstResponseTime      prometheus.Histogram
    CustomerSatisfaction   prometheus.Histogram
    AgentUtilization      prometheus.Gauge
    KnowledgeBaseHitRate  prometheus.Histogram
}

func NewCustomerSupportSystem(config *SupportConfig) (*CustomerSupportSystem, error) {
    // Initialize database
    db, err := sqlx.Connect("postgres", config.DatabaseURL)
    if err != nil {
        return nil, fmt.Errorf("database connection failed: %w", err)
    }

    // Create LLM provider
    llm, err := provider.NewOpenAI(
        provider.WithModel("gpt-4"),
        provider.WithMaxTokens(2000),
    )
    if err != nil {
        return nil, fmt.Errorf("LLM provider creation failed: %w", err)
    }

    // Create specialized agents
    intakeAgent := core.NewLLMAgent("intake-agent", llm)
    classificationAgent := core.NewLLMAgent("classification-agent", llm)
    knowledgeAgent := core.NewLLMAgent("knowledge-agent", llm)
    responseAgent := core.NewLLMAgent("response-agent", llm)

    // Add tools to agents
    webTool := web.NewWebFetchTool()
    knowledgeAgent.AddTool(webTool)

    // Initialize components
    escalationManager := NewEscalationManager()
    knowledgeBase := NewKnowledgeBase(db, config.KnowledgeBaseURL)
    analytics := NewAnalyticsEngine(db)
    metrics := initializeMetrics()

    system := &CustomerSupportSystem{
        db:                  db,
        intakeAgent:         intakeAgent,
        classificationAgent: classificationAgent,
        knowledgeAgent:      knowledgeAgent,
        responseAgent:       responseAgent,
        escalationManager:   escalationManager,
        knowledgeBase:       knowledgeBase,
        analytics:          analytics,
        config:             config,
        metrics:            metrics,
    }

    // Initialize database schema
    if err := system.initializeSchema(); err != nil {
        return nil, fmt.Errorf("schema initialization failed: %w", err)
    }

    return system, nil
}

func (css *CustomerSupportSystem) initializeSchema() error {
    schema := `
    CREATE TABLE IF NOT EXISTS support_tickets (
        id SERIAL PRIMARY KEY,
        customer_id VARCHAR(255) NOT NULL,
        subject TEXT NOT NULL,
        description TEXT NOT NULL,
        status VARCHAR(50) DEFAULT 'open',
        priority VARCHAR(20) DEFAULT 'medium',
        category VARCHAR(100),
        channel VARCHAR(50) NOT NULL,
        assigned_agent VARCHAR(255),
        customer_email VARCHAR(255) NOT NULL,
        customer_name VARCHAR(255) NOT NULL,
        created_at TIMESTAMP DEFAULT NOW(),
        updated_at TIMESTAMP DEFAULT NOW(),
        resolved_at TIMESTAMP,
        first_response TIMESTAMP,
        metadata JSONB,
        sentiment_score DECIMAL(3,2) DEFAULT 0.0,
        complexity_score DECIMAL(3,2) DEFAULT 0.0
    );

    CREATE TABLE IF NOT EXISTS ticket_messages (
        id SERIAL PRIMARY KEY,
        ticket_id INTEGER REFERENCES support_tickets(id),
        sender VARCHAR(50) NOT NULL,
        content TEXT NOT NULL,
        message_type VARCHAR(50) DEFAULT 'text',
        metadata JSONB,
        created_at TIMESTAMP DEFAULT NOW()
    );

    CREATE TABLE IF NOT EXISTS knowledge_base (
        id SERIAL PRIMARY KEY,
        title TEXT NOT NULL,
        content TEXT NOT NULL,
        category VARCHAR(100),
        tags TEXT[],
        confidence DECIMAL(3,2) DEFAULT 1.0,
        use_count INTEGER DEFAULT 0,
        last_updated TIMESTAMP DEFAULT NOW()
    );

    CREATE INDEX IF NOT EXISTS idx_tickets_status ON support_tickets(status);
    CREATE INDEX IF NOT EXISTS idx_tickets_customer ON support_tickets(customer_id);
    CREATE INDEX IF NOT EXISTS idx_tickets_created ON support_tickets(created_at);
    CREATE INDEX IF NOT EXISTS idx_messages_ticket ON ticket_messages(ticket_id);
    `

    _, err := css.db.Exec(schema)
    return err
}

// Main ticket processing workflow
func (css *CustomerSupportSystem) ProcessTicket(ctx context.Context, ticket *SupportTicket) (*SupportResponse, error) {
    start := time.Now()
    css.metrics.TicketsCreated.Inc()

    // Step 1: Store ticket in database
    if err := css.storeTicket(ctx, ticket); err != nil {
        return nil, fmt.Errorf("failed to store ticket: %w", err)
    }

    // Step 2: Classify ticket
    classification, err := css.classifyTicket(ctx, ticket)
    if err != nil {
        log.Printf("Classification failed: %v", err)
        // Continue with default classification
        classification = &TicketClassification{
            Category:   "general",
            Priority:   "medium",
            Sentiment:  0.0,
            Complexity: 0.5,
            Confidence: 0.3,
        }
    }

    // Update ticket with classification
    ticket.Category = classification.Category
    ticket.Priority = classification.Priority
    ticket.SentimentScore = classification.Sentiment
    ticket.ComplexityScore = classification.Complexity

    // Step 3: Check if immediate escalation is needed
    if classification.RequiresHuman || classification.Complexity > css.config.EscalationThreshold {
        return css.escalateTicket(ctx, ticket, classification)
    }

    // Step 4: Search knowledge base
    knowledgeEntries, err := css.knowledgeBase.Search(ctx, ticket.Description, ticket.Category)
    if err != nil {
        log.Printf("Knowledge base search failed: %v", err)
        knowledgeEntries = []KnowledgeBaseEntry{}
    }

    // Step 5: Generate response
    response, err := css.generateResponse(ctx, ticket, classification, knowledgeEntries)
    if err != nil {
        return nil, fmt.Errorf("response generation failed: %w", err)
    }

    // Step 6: Store response message
    message := &TicketMessage{
        TicketID:    ticket.ID,
        Sender:      "system",
        Content:     response.Content,
        MessageType: "auto_response",
        Metadata: map[string]interface{}{
            "confidence":       response.Confidence,
            "knowledge_used":   response.KnowledgeUsed,
            "ai_generated":     true,
        },
        CreatedAt: time.Now(),
    }

    if err := css.storeMessage(ctx, message); err != nil {
        log.Printf("Failed to store response message: %v", err)
    }

    // Step 7: Update ticket status
    ticket.Status = "in_progress"
    if ticket.FirstResponse == nil {
        now := time.Now()
        ticket.FirstResponse = &now
    }
    ticket.UpdatedAt = time.Now()

    if err := css.updateTicket(ctx, ticket); err != nil {
        log.Printf("Failed to update ticket: %v", err)
    }

    // Step 8: Check if escalation is needed based on response confidence
    if response.RequiresEscalation || response.Confidence < 0.6 {
        return css.escalateTicket(ctx, ticket, classification)
    }

    // Record metrics
    css.metrics.ResponseTime.Observe(time.Since(start).Seconds())
    if response.Confidence > 0.8 {
        css.metrics.TicketsResolved.Inc()
        ticket.Status = "resolved"
        now := time.Now()
        ticket.ResolvedAt = &now
        css.updateTicket(ctx, ticket)
    }

    // Update knowledge base usage
    css.knowledgeBase.UpdateUsage(ctx, response.KnowledgeUsed)

    return response, nil
}

func (css *CustomerSupportSystem) classifyTicket(ctx context.Context, ticket *SupportTicket) (*TicketClassification, error) {
    prompt := fmt.Sprintf(`Analyze this customer support ticket and provide classification:

Subject: %s
Description: %s
Customer: %s <%s>
Channel: %s

Please classify this ticket and return a JSON response with:
1. category (billing, technical, account, product, general)
2. priority (low, medium, high, urgent)
3. urgency (low, medium, high, critical)
4. sentiment (score from -1.0 to 1.0, where -1 is very negative, 0 is neutral, 1 is very positive)
5. complexity (score from 0.0 to 1.0, where 0 is simple, 1 is very complex)
6. requires_human (boolean - true if this needs human intervention)
7. suggested_agent (if specific expertise is needed)
8. confidence (score from 0.0 to 1.0 for classification accuracy)
9. reasoning (brief explanation of the classification)

Consider factors like:
- Technical complexity
- Customer emotion and urgency
- Policy requirements
- Potential for automated resolution`,
        ticket.Subject, ticket.Description, ticket.CustomerName, ticket.CustomerEmail, ticket.Channel)

    state := domain.NewState()
    state.AddMessage(domain.NewTextMessage(domain.RoleUser, prompt))

    result, err := css.classificationAgent.Run(ctx, state)
    if err != nil {
        return nil, err
    }

    var classification TicketClassification
    if len(result.Messages) > 0 {
        response := result.Messages[len(result.Messages)-1].TextContent()
        if err := json.Unmarshal([]byte(response), &classification); err != nil {
            return nil, fmt.Errorf("failed to parse classification: %w", err)
        }
    }

    return &classification, nil
}

func (css *CustomerSupportSystem) generateResponse(ctx context.Context, ticket *SupportTicket, classification *TicketClassification, knowledge []KnowledgeBaseEntry) (*SupportResponse, error) {
    // Prepare knowledge context
    var knowledgeContext strings.Builder
    var knowledgeIDs []int

    for _, entry := range knowledge {
        knowledgeContext.WriteString(fmt.Sprintf("\n[KB_%d] %s: %s", entry.ID, entry.Title, entry.Content))
        knowledgeIDs = append(knowledgeIDs, entry.ID)
    }

    prompt := fmt.Sprintf(`You are a helpful customer support agent. Generate a response to this support ticket:

Customer: %s <%s>
Subject: %s
Description: %s
Category: %s
Priority: %s
Customer Sentiment: %.2f

Available Knowledge Base Information:%s

Please provide a helpful, professional response that:
1. Acknowledges the customer's concern
2. Provides a clear solution or next steps
3. Uses relevant information from the knowledge base when applicable
4. Maintains a friendly and professional tone
5. Offers additional assistance if needed

After your response, provide a JSON object with:
{
    "content": "your response here",
    "confidence": 0.0-1.0,
    "knowledge_used": [list of KB IDs used],
    "suggested_actions": ["action1", "action2"],
    "requires_escalation": true/false,
    "follow_up_required": true/false,
    "estimated_resolution": "immediate/1-2_hours/1-2_days/complex"
}`,
        ticket.CustomerName, ticket.CustomerEmail, ticket.Subject, ticket.Description,
        classification.Category, classification.Priority, classification.Sentiment,
        knowledgeContext.String())

    state := domain.NewState()
    state.AddMessage(domain.NewTextMessage(domain.RoleUser, prompt))

    result, err := css.responseAgent.Run(ctx, state)
    if err != nil {
        return nil, err
    }

    if len(result.Messages) == 0 {
        return nil, fmt.Errorf("no response generated")
    }

    responseText := result.Messages[len(result.Messages)-1].TextContent()
    
    // Extract JSON from response
    jsonStart := strings.Index(responseText, "{")
    jsonEnd := strings.LastIndex(responseText, "}")
    
    if jsonStart == -1 || jsonEnd == -1 {
        // Fallback response
        return &SupportResponse{
            Content:            responseText,
            Confidence:         0.5,
            KnowledgeUsed:      knowledgeIDs,
            RequiresEscalation: classification.RequiresHuman,
        }, nil
    }

    var response SupportResponse
    jsonPart := responseText[jsonStart : jsonEnd+1]
    if err := json.Unmarshal([]byte(jsonPart), &response); err != nil {
        // Fallback response
        content := responseText[:jsonStart]
        if content == "" {
            content = responseText
        }
        
        return &SupportResponse{
            Content:            strings.TrimSpace(content),
            Confidence:         0.5,
            KnowledgeUsed:      knowledgeIDs,
            RequiresEscalation: classification.RequiresHuman,
        }, nil
    }

    return &response, nil
}

func (css *CustomerSupportSystem) escalateTicket(ctx context.Context, ticket *SupportTicket, classification *TicketClassification) (*SupportResponse, error) {
    css.metrics.TicketsEscalated.Inc()

    ticket.Status = "escalated"
    
    // Find available human agent
    agent, err := css.escalationManager.FindAvailableAgent(classification.Category, classification.Priority)
    if err != nil {
        log.Printf("No available agents for escalation: %v", err)
        ticket.AssignedAgent = nil
    } else {
        ticket.AssignedAgent = &agent.ID
    }

    // Update ticket
    if err := css.updateTicket(ctx, ticket); err != nil {
        log.Printf("Failed to update escalated ticket: %v", err)
    }

    // Notify escalation
    if err := css.notifyEscalation(ctx, ticket, classification); err != nil {
        log.Printf("Failed to notify escalation: %v", err)
    }

    response := &SupportResponse{
        Content: fmt.Sprintf("Thank you for contacting us, %s. Your ticket has been escalated to our specialized team for immediate attention. A human agent will respond within our service level agreement timeframe. Your ticket ID is #%d for reference.",
            ticket.CustomerName, ticket.ID),
        Confidence:         1.0,
        RequiresEscalation: true,
        EstimatedResolution: "1-2_hours",
    }

    return response, nil
}

// Database operations
func (css *CustomerSupportSystem) storeTicket(ctx context.Context, ticket *SupportTicket) error {
    query := `INSERT INTO support_tickets 
              (customer_id, subject, description, status, priority, category, channel, 
               customer_email, customer_name, metadata, sentiment_score, complexity_score)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
              RETURNING id, created_at`

    metadataJSON, _ := json.Marshal(ticket.Metadata)
    
    err := css.db.QueryRowContext(ctx, query,
        ticket.CustomerID, ticket.Subject, ticket.Description, ticket.Status,
        ticket.Priority, ticket.Category, ticket.Channel, ticket.CustomerEmail,
        ticket.CustomerName, metadataJSON, ticket.SentimentScore, ticket.ComplexityScore,
    ).Scan(&ticket.ID, &ticket.CreatedAt)

    return err
}

func (css *CustomerSupportSystem) updateTicket(ctx context.Context, ticket *SupportTicket) error {
    query := `UPDATE support_tickets 
              SET status = $1, priority = $2, category = $3, assigned_agent = $4,
                  updated_at = $5, resolved_at = $6, first_response = $7,
                  sentiment_score = $8, complexity_score = $9
              WHERE id = $10`

    _, err := css.db.ExecContext(ctx, query,
        ticket.Status, ticket.Priority, ticket.Category, ticket.AssignedAgent,
        ticket.UpdatedAt, ticket.ResolvedAt, ticket.FirstResponse,
        ticket.SentimentScore, ticket.ComplexityScore, ticket.ID)

    return err
}

func (css *CustomerSupportSystem) storeMessage(ctx context.Context, message *TicketMessage) error {
    query := `INSERT INTO ticket_messages 
              (ticket_id, sender, content, message_type, metadata)
              VALUES ($1, $2, $3, $4, $5)
              RETURNING id, created_at`

    metadataJSON, _ := json.Marshal(message.Metadata)
    
    err := css.db.QueryRowContext(ctx, query,
        message.TicketID, message.Sender, message.Content,
        message.MessageType, metadataJSON,
    ).Scan(&message.ID, &message.CreatedAt)

    return err
}

// HTTP API endpoints
func (css *CustomerSupportSystem) CreateTicketHandler(c *gin.Context) {
    var request struct {
        CustomerID    string                 `json:"customer_id" binding:"required"`
        CustomerName  string                 `json:"customer_name" binding:"required"`
        CustomerEmail string                 `json:"customer_email" binding:"required,email"`
        Subject       string                 `json:"subject" binding:"required"`
        Description   string                 `json:"description" binding:"required"`
        Channel       string                 `json:"channel" binding:"required"`
        Metadata      map[string]interface{} `json:"metadata"`
    }

    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    ticket := &SupportTicket{
        CustomerID:    request.CustomerID,
        CustomerName:  request.CustomerName,
        CustomerEmail: request.CustomerEmail,
        Subject:       request.Subject,
        Description:   request.Description,
        Channel:       request.Channel,
        Status:        "open",
        Priority:      "medium",
        Metadata:      request.Metadata,
    }

    ctx, cancel := context.WithTimeout(c.Request.Context(), css.config.MaxResponseTime)
    defer cancel()

    response, err := css.ProcessTicket(ctx, ticket)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "ticket":   ticket,
        "response": response,
    })
}

func (css *CustomerSupportSystem) GetTicketHandler(c *gin.Context) {
    ticketID, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket ID"})
        return
    }

    var ticket SupportTicket
    query := `SELECT * FROM support_tickets WHERE id = $1`
    
    err = css.db.GetContext(c.Request.Context(), &ticket, query, ticketID)
    if err != nil {
        if err == sql.ErrNoRows {
            c.JSON(http.StatusNotFound, gin.H{"error": "Ticket not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // Get ticket messages
    var messages []TicketMessage
    messageQuery := `SELECT * FROM ticket_messages WHERE ticket_id = $1 ORDER BY created_at ASC`
    err = css.db.SelectContext(c.Request.Context(), &messages, messageQuery, ticketID)
    if err != nil {
        log.Printf("Failed to fetch messages: %v", err)
        messages = []TicketMessage{}
    }

    c.JSON(http.StatusOK, gin.H{
        "ticket":   ticket,
        "messages": messages,
    })
}

func (css *CustomerSupportSystem) AddMessageHandler(c *gin.Context) {
    ticketID, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket ID"})
        return
    }

    var request struct {
        Sender      string                 `json:"sender" binding:"required"`
        Content     string                 `json:"content" binding:"required"`
        MessageType string                 `json:"message_type"`
        Metadata    map[string]interface{} `json:"metadata"`
    }

    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    message := &TicketMessage{
        TicketID:    ticketID,
        Sender:      request.Sender,
        Content:     request.Content,
        MessageType: request.MessageType,
        Metadata:    request.Metadata,
    }

    if message.MessageType == "" {
        message.MessageType = "text"
    }

    if err := css.storeMessage(c.Request.Context(), message); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"message": message})
}

// Supporting components
type EscalationManager struct {
    agents []Agent
    mu     sync.RWMutex
}

type Agent struct {
    ID           string   `json:"id"`
    Name         string   `json:"name"`
    Email        string   `json:"email"`
    Specialties  []string `json:"specialties"`
    Available    bool     `json:"available"`
    CurrentLoad  int      `json:"current_load"`
    MaxLoad      int      `json:"max_load"`
}

func NewEscalationManager() *EscalationManager {
    return &EscalationManager{
        agents: []Agent{
            {
                ID:          "agent_001",
                Name:        "Sarah Chen",
                Email:       "sarah.chen@company.com",
                Specialties: []string{"technical", "billing"},
                Available:   true,
                MaxLoad:     10,
            },
            {
                ID:          "agent_002",
                Name:        "Mike Rodriguez",
                Email:       "mike.rodriguez@company.com",
                Specialties: []string{"account", "product"},
                Available:   true,
                MaxLoad:     8,
            },
        },
    }
}

func (em *EscalationManager) FindAvailableAgent(category, priority string) (*Agent, error) {
    em.mu.RLock()
    defer em.mu.RUnlock()

    for i := range em.agents {
        agent := &em.agents[i]
        if agent.Available && agent.CurrentLoad < agent.MaxLoad {
            // Check if agent has relevant specialty
            for _, specialty := range agent.Specialties {
                if specialty == category {
                    agent.CurrentLoad++
                    return agent, nil
                }
            }
        }
    }

    // Fallback to any available agent
    for i := range em.agents {
        agent := &em.agents[i]
        if agent.Available && agent.CurrentLoad < agent.MaxLoad {
            agent.CurrentLoad++
            return agent, nil
        }
    }

    return nil, fmt.Errorf("no available agents")
}

type KnowledgeBase struct {
    db      *sqlx.DB
    baseURL string
}

func NewKnowledgeBase(db *sqlx.DB, baseURL string) *KnowledgeBase {
    return &KnowledgeBase{
        db:      db,
        baseURL: baseURL,
    }
}

func (kb *KnowledgeBase) Search(ctx context.Context, query, category string) ([]KnowledgeBaseEntry, error) {
    searchQuery := `
        SELECT id, title, content, category, tags, confidence, use_count, last_updated
        FROM knowledge_base 
        WHERE category = $1 OR $1 = '' 
        AND (content ILIKE $2 OR title ILIKE $2)
        ORDER BY confidence DESC, use_count DESC
        LIMIT 5`

    var entries []KnowledgeBaseEntry
    searchPattern := "%" + query + "%"
    
    err := kb.db.SelectContext(ctx, &entries, searchQuery, category, searchPattern)
    return entries, err
}

func (kb *KnowledgeBase) UpdateUsage(ctx context.Context, entryIDs []int) error {
    if len(entryIDs) == 0 {
        return nil
    }

    query := `UPDATE knowledge_base SET use_count = use_count + 1 WHERE id = ANY($1)`
    _, err := kb.db.ExecContext(ctx, query, pq.Array(entryIDs))
    return err
}

type AnalyticsEngine struct {
    db *sqlx.DB
}

func NewAnalyticsEngine(db *sqlx.DB) *AnalyticsEngine {
    return &AnalyticsEngine{db: db}
}

func (ae *AnalyticsEngine) GetSupportMetrics(ctx context.Context, days int) (map[string]interface{}, error) {
    query := `
        SELECT 
            COUNT(*) as total_tickets,
            COUNT(CASE WHEN status = 'resolved' THEN 1 END) as resolved_tickets,
            COUNT(CASE WHEN status = 'escalated' THEN 1 END) as escalated_tickets,
            AVG(EXTRACT(EPOCH FROM (first_response - created_at))/60) as avg_first_response_minutes,
            AVG(EXTRACT(EPOCH FROM (resolved_at - created_at))/3600) as avg_resolution_hours,
            AVG(sentiment_score) as avg_sentiment
        FROM support_tickets 
        WHERE created_at >= NOW() - INTERVAL '%d days'`

    var metrics struct {
        TotalTickets            int     `db:"total_tickets"`
        ResolvedTickets         int     `db:"resolved_tickets"`
        EscalatedTickets        int     `db:"escalated_tickets"`
        AvgFirstResponseMinutes float64 `db:"avg_first_response_minutes"`
        AvgResolutionHours      float64 `db:"avg_resolution_hours"`
        AvgSentiment           float64 `db:"avg_sentiment"`
    }

    err := ae.db.GetContext(ctx, &metrics, fmt.Sprintf(query, days))
    if err != nil {
        return nil, err
    }

    result := map[string]interface{}{
        "total_tickets":              metrics.TotalTickets,
        "resolved_tickets":           metrics.ResolvedTickets,
        "escalated_tickets":          metrics.EscalatedTickets,
        "resolution_rate":           float64(metrics.ResolvedTickets) / float64(metrics.TotalTickets),
        "escalation_rate":           float64(metrics.EscalatedTickets) / float64(metrics.TotalTickets),
        "avg_first_response_minutes": metrics.AvgFirstResponseMinutes,
        "avg_resolution_hours":       metrics.AvgResolutionHours,
        "avg_sentiment":             metrics.AvgSentiment,
    }

    return result, nil
}

func (css *CustomerSupportSystem) notifyEscalation(ctx context.Context, ticket *SupportTicket, classification *TicketClassification) error {
    // Send notification to Slack/email
    message := fmt.Sprintf("🚨 Ticket #%d escalated\nCustomer: %s\nSubject: %s\nPriority: %s\nCategory: %s",
        ticket.ID, ticket.CustomerName, ticket.Subject, ticket.Priority, ticket.Category)

    // Implementation would send to Slack/email based on config
    log.Printf("Escalation notification: %s", message)
    return nil
}

func initializeMetrics() *SupportMetrics {
    metrics := &SupportMetrics{
        TicketsCreated: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "support_tickets_created_total",
            Help: "Total number of support tickets created",
        }),
        TicketsResolved: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "support_tickets_resolved_total",
            Help: "Total number of support tickets resolved",
        }),
        TicketsEscalated: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "support_tickets_escalated_total",
            Help: "Total number of support tickets escalated",
        }),
        ResponseTime: prometheus.NewHistogram(prometheus.HistogramOpts{
            Name: "support_response_time_seconds",
            Help: "Time taken to generate initial response",
        }),
        FirstResponseTime: prometheus.NewHistogram(prometheus.HistogramOpts{
            Name: "support_first_response_time_seconds",
            Help: "Time to first response for tickets",
        }),
        CustomerSatisfaction: prometheus.NewHistogram(prometheus.HistogramOpts{
            Name: "support_customer_satisfaction",
            Help: "Customer satisfaction scores",
        }),
        AgentUtilization: prometheus.NewGauge(prometheus.GaugeOpts{
            Name: "support_agent_utilization",
            Help: "Current agent utilization percentage",
        }),
        KnowledgeBaseHitRate: prometheus.NewHistogram(prometheus.HistogramOpts{
            Name: "support_knowledge_base_hit_rate",
            Help: "Knowledge base hit rate for queries",
        }),
    }

    prometheus.MustRegister(
        metrics.TicketsCreated,
        metrics.TicketsResolved,
        metrics.TicketsEscalated,
        metrics.ResponseTime,
        metrics.FirstResponseTime,
        metrics.CustomerSatisfaction,
        metrics.AgentUtilization,
        metrics.KnowledgeBaseHitRate,
    )

    return metrics
}

func main() {
    config := &SupportConfig{
        DatabaseURL:         "postgres://user:pass@localhost/support_db?sslmode=disable",
        MaxResponseTime:     30 * time.Second,
        EscalationThreshold: 0.7,
        KnowledgeBaseURL:    "https://docs.company.com",
        EmailConfig: EmailConfig{
            SMTPHost:    "smtp.company.com",
            SMTPPort:    587,
            FromAddress: "support@company.com",
        },
        SlackConfig: SlackConfig{
            Channel: "#support-escalations",
        },
    }

    system, err := NewCustomerSupportSystem(config)
    if err != nil {
        log.Fatal("Failed to create support system:", err)
    }

    // Setup HTTP server
    r := gin.Default()

    // API routes
    api := r.Group("/api/v1")
    {
        api.POST("/tickets", system.CreateTicketHandler)
        api.GET("/tickets/:id", system.GetTicketHandler)
        api.POST("/tickets/:id/messages", system.AddMessageHandler)
        api.GET("/metrics", gin.WrapH(promhttp.Handler()))
    }

    // Health check
    r.GET("/health", func(c *gin.Context) {
        metrics, err := system.analytics.GetSupportMetrics(c.Request.Context(), 1)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        
        c.JSON(http.StatusOK, gin.H{
            "status":  "healthy",
            "metrics": metrics,
        })
    })

    log.Println("Customer Support System starting on :8080")
    log.Fatal(r.Run(":8080"))
}
```

## Usage Examples

### Creating a Support Ticket

```bash
curl -X POST http://localhost:8080/api/v1/tickets \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "cust_12345",
    "customer_name": "John Doe",
    "customer_email": "john.doe@example.com",
    "subject": "Unable to access my account",
    "description": "I have been trying to log into my account for the past hour but keep getting an error message saying my credentials are invalid. I am sure I am using the correct password.",
    "channel": "email",
    "metadata": {
      "user_agent": "Mozilla/5.0...",
      "ip_address": "192.168.1.100"
    }
  }'
```

### Adding a Follow-up Message

```bash
curl -X POST http://localhost:8080/api/v1/tickets/123/messages \
  -H "Content-Type: application/json" \
  -d '{
    "sender": "customer",
    "content": "I tried the suggested steps but still cannot access my account.",
    "message_type": "text"
  }'
```

### Getting Ticket Information

```bash
curl http://localhost:8080/api/v1/tickets/123
```

## Key Features Demonstrated

1. **Multi-Agent Architecture** - Specialized agents for different tasks
2. **Knowledge Base Integration** - Automatic answer generation from documentation
3. **Intelligent Classification** - AI-powered ticket categorization and routing
4. **Escalation Management** - Smart escalation to human agents when needed
5. **Comprehensive Monitoring** - Metrics and analytics for system performance
6. **Database Integration** - Persistent storage for tickets and conversations
7. **RESTful API** - Clean API for external integrations
8. **Production Ready** - Error handling, logging, and monitoring

## Deployment Considerations

- **Database Setup** - PostgreSQL with proper indexing for performance
- **Environment Variables** - Secure configuration management
- **Load Balancing** - Multiple instances for high availability
- **Monitoring** - Prometheus metrics and alerting
- **Backup Strategy** - Regular database backups
- **Security** - API authentication and rate limiting

This customer support system demonstrates how to build a complete, production-ready AI-powered application using Go-LLMs, showcasing integration of multiple agents, databases, monitoring, and real-world business logic.