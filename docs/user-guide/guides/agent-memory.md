# Agent Memory: State Management Patterns

> **[Project Root](/) / [Documentation](../..) / [User Guide](../../user-guide) / [Guides](../../user-guide/guides) / Agent Memory**

Master sophisticated memory management patterns for intelligent agents. Learn to implement conversation history, context persistence, hierarchical memory, and custom memory strategies to build agents that remember, learn, and adapt over time.

## Why Agent Memory Matters

- **Continuity** - Maintain context across conversations and sessions
- **Learning** - Accumulate knowledge and improve responses over time
- **Personalization** - Remember user preferences and interaction patterns
- **Context Awareness** - Understand conversational flow and dependencies
- **Performance** - Optimize memory usage while preserving important information

## Memory Architecture Overview

![Agent Memory System](../../images/agent-memory-architecture.svg)

### Core Components
1. **Working Memory** - Active state values and current context
2. **Conversation Memory** - Message history and dialogue flow
3. **Persistent Memory** - Long-term storage via events and artifacts
4. **Shared Memory** - Hierarchical state sharing between agents
5. **Memory Strategies** - Sliding windows, episodic storage, selective retention

### Memory Types
| Type | Scope | Lifespan | Use Cases |
|------|-------|----------|-----------|
| **Working Memory** | Current execution | Single run | Variables, intermediate results |
| **Conversation Memory** | Session/conversation | Until cleared | Chat history, dialogue context |
| **Episodic Memory** | Task/session | Long-term | Important experiences, outcomes |
| **Semantic Memory** | Agent knowledge | Permanent | Facts, procedures, learned patterns |
| **Shared Memory** | Multi-agent | Variable | Coordination, shared context |

## Prerequisites

- [Creating Agents completed](creating-agents.md) ✅
- [Agent Communication understanding](agent-communication.md) ✅
- Basic knowledge of state management concepts ✅

---

## Level 1: Basic State Management
*Implement working memory and conversation history*

### Understanding Agent State
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

// MemoryAgent demonstrates basic state management
type MemoryAgent struct {
    name           string
    baseAgent      domain.BaseAgent
    stateManager   *core.StateManager
    currentState   *domain.State
    memoryStrategy MemoryStrategy
}

type MemoryStrategy interface {
    ProcessState(state *domain.State) *domain.State
    ShouldPersist(state *domain.State) bool
    GetRetentionPolicy() RetentionPolicy
}

type RetentionPolicy struct {
    MaxMessages    int           // Maximum conversation messages
    MaxAge         time.Duration // Maximum age for state data
    CleanupInterval time.Duration // How often to clean up
    CriticalKeys   []string      // Keys that should never be deleted
}

func NewMemoryAgent(name, provider string) (*MemoryAgent, error) {
    baseAgent, err := core.NewAgentFromString(name, provider)
    if err != nil {
        return nil, err
    }

    return &MemoryAgent{
        name:         name,
        baseAgent:    baseAgent,
        stateManager: core.NewStateManager(),
        currentState: domain.NewState(),
        memoryStrategy: NewSlidingWindowStrategy(50), // Keep last 50 messages
    }, nil
}

func (ma *MemoryAgent) SetSystemPrompt(prompt string) {
    ma.baseAgent.SetSystemPrompt(prompt)
    
    // Store system prompt in state for memory
    ma.currentState.Set("system_prompt", prompt)
    ma.currentState.Set("agent_name", ma.name)
}

func (ma *MemoryAgent) ProcessInput(ctx context.Context, input string) (string, error) {
    fmt.Printf("🧠 [%s] Processing input with memory context\n", ma.name)
    
    // Add user input to conversation memory
    userMessage := domain.NewMessage(domain.RoleUser, input)
    userMessage.Metadata = map[string]interface{}{
        "timestamp": time.Now(),
        "input_length": len(input),
    }
    ma.currentState.AddMessage(userMessage)
    
    // Apply memory strategy (e.g., sliding window)
    processedState := ma.memoryStrategy.ProcessState(ma.currentState)
    
    // Set conversation context for agent
    processedState.Set("user_input", input)
    processedState.Set("conversation_turn", len(processedState.GetMessages()))
    
    // Run agent with memory context
    result, err := ma.baseAgent.Run(ctx, processedState)
    if err != nil {
        return "", fmt.Errorf("agent execution failed: %w", err)
    }
    
    // Extract response
    response, exists := result.Get("response")
    if !exists {
        return "", fmt.Errorf("no response in result")
    }
    
    responseStr := response.(string)
    
    // Add assistant response to memory
    assistantMessage := domain.NewMessage(domain.RoleAssistant, responseStr)
    assistantMessage.Metadata = map[string]interface{}{
        "timestamp": time.Now(),
        "response_length": len(responseStr),
        "processing_time": time.Since(userMessage.Timestamp),
    }
    ma.currentState.AddMessage(assistantMessage)
    
    // Update current state with results
    ma.currentState = result
    
    // Persist state if needed
    if ma.memoryStrategy.ShouldPersist(ma.currentState) {
        err = ma.stateManager.SaveState(ma.currentState)
        if err != nil {
            log.Printf("Failed to persist state: %v", err)
        }
    }
    
    fmt.Printf("💾 Memory state: %d messages, %d values\n", 
        len(ma.currentState.GetMessages()), 
        len(ma.currentState.GetAllValues()))
    
    return responseStr, nil
}

func (ma *MemoryAgent) GetConversationHistory() []domain.Message {
    return ma.currentState.GetMessages()
}

func (ma *MemoryAgent) GetMemoryStats() MemoryStats {
    messages := ma.currentState.GetMessages()
    values := ma.currentState.GetAllValues()
    
    var totalMessageLength int
    for _, msg := range messages {
        totalMessageLength += len(msg.Content)
    }
    
    return MemoryStats{
        MessageCount:       len(messages),
        TotalMessageLength: totalMessageLength,
        StateValueCount:    len(values),
        StateAge:          time.Since(ma.currentState.Created()),
        LastModified:      ma.currentState.Modified(),
    }
}

type MemoryStats struct {
    MessageCount       int
    TotalMessageLength int
    StateValueCount    int
    StateAge          time.Duration
    LastModified      time.Time
}

// Sliding Window Memory Strategy
type SlidingWindowStrategy struct {
    maxMessages int
}

func NewSlidingWindowStrategy(maxMessages int) *SlidingWindowStrategy {
    return &SlidingWindowStrategy{maxMessages: maxMessages}
}

func (sws *SlidingWindowStrategy) ProcessState(state *domain.State) *domain.State {
    messages := state.GetMessages()
    
    if len(messages) <= sws.maxMessages {
        return state
    }
    
    // Keep system messages and recent messages
    var filteredMessages []domain.Message
    var systemMessages []domain.Message
    var recentMessages []domain.Message
    
    // Separate system messages
    for _, msg := range messages {
        if msg.Role == domain.RoleSystem {
            systemMessages = append(systemMessages, msg)
        } else {
            recentMessages = append(recentMessages, msg)
        }
    }
    
    // Keep only recent non-system messages
    startIdx := len(recentMessages) - (sws.maxMessages - len(systemMessages))
    if startIdx < 0 {
        startIdx = 0
    }
    
    filteredMessages = append(filteredMessages, systemMessages...)
    filteredMessages = append(filteredMessages, recentMessages[startIdx:]...)
    
    // Create new state with filtered messages
    newState := state.Clone()
    newState.ClearMessages()
    for _, msg := range filteredMessages {
        newState.AddMessage(msg)
    }
    
    return newState
}

func (sws *SlidingWindowStrategy) ShouldPersist(state *domain.State) bool {
    // Persist every 10 messages or when memory is getting full
    messageCount := len(state.GetMessages())
    return messageCount%10 == 0 || messageCount > sws.maxMessages-5
}

func (sws *SlidingWindowStrategy) GetRetentionPolicy() RetentionPolicy {
    return RetentionPolicy{
        MaxMessages:    sws.maxMessages,
        MaxAge:         24 * time.Hour,
        CleanupInterval: time.Hour,
        CriticalKeys:   []string{"system_prompt", "agent_name", "user_id"},
    }
}

func main() {
    fmt.Println("🧠 Agent Memory - Basic State Management")
    fmt.Println("======================================")

    // Create memory-aware agent
    agent, err := NewMemoryAgent("memory-assistant", "openai/gpt-4o-mini")
    if err != nil {
        log.Fatalf("Failed to create agent: %v", err)
    }

    agent.SetSystemPrompt(`You are a helpful assistant with memory. 
    You can remember our conversation history and refer to previous exchanges.
    Always acknowledge when you're using information from our conversation history.`)

    ctx := context.Background()

    // Conversation with memory
    conversations := []string{
        "Hi, my name is Sarah and I'm a software engineer.",
        "What programming languages do you think I should learn?",
        "I mentioned my profession earlier. What was it?",
        "Based on our conversation, what do you know about me?",
    }

    for i, input := range conversations {
        fmt.Printf("\n--- Turn %d ---\n", i+1)
        fmt.Printf("👤 User: %s\n", input)
        
        response, err := agent.ProcessInput(ctx, input)
        if err != nil {
            log.Printf("Error: %v", err)
            continue
        }
        
        fmt.Printf("🤖 Assistant: %s\n", response)
        
        // Show memory stats
        stats := agent.GetMemoryStats()
        fmt.Printf("📊 Memory: %d messages, %d values, age: %v\n", 
            stats.MessageCount, stats.StateValueCount, stats.StateAge.Round(time.Second))
    }

    // Display full conversation history
    fmt.Printf("\n📚 Full Conversation History:\n")
    fmt.Printf("============================\n")
    history := agent.GetConversationHistory()
    for i, msg := range history {
        fmt.Printf("%d. [%s] %s\n", i+1, msg.Role, msg.Content)
    }
}
```

---

## Level 2: Persistent Memory and Session Management
*Implement long-term memory with persistence*

### Session-Based Memory System
```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "time"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
)

// PersistentMemoryAgent with session management
type PersistentMemoryAgent struct {
    name           string
    baseAgent      domain.BaseAgent
    sessionManager *SessionManager
    currentSession *Session
    memoryStore    MemoryStore
}

type Session struct {
    ID          string                 `json:"id"`
    UserID      string                 `json:"user_id"`
    StartTime   time.Time             `json:"start_time"`
    LastActive  time.Time             `json:"last_active"`
    State       *domain.State         `json:"state"`
    Metadata    map[string]interface{} `json:"metadata"`
    Summary     string                `json:"summary"`
    Tags        []string              `json:"tags"`
}

type SessionManager struct {
    sessionsDir string
    maxAge      time.Duration
    autoSave    bool
}

type MemoryStore interface {
    SaveSession(session *Session) error
    LoadSession(sessionID string) (*Session, error)
    ListSessions(userID string) ([]SessionSummary, error)
    DeleteSession(sessionID string) error
    SearchSessions(query SessionQuery) ([]SessionSummary, error)
}

type SessionSummary struct {
    ID         string    `json:"id"`
    UserID     string    `json:"user_id"`
    StartTime  time.Time `json:"start_time"`
    LastActive time.Time `json:"last_active"`
    Summary    string    `json:"summary"`
    Tags       []string  `json:"tags"`
    MessageCount int     `json:"message_count"`
}

type SessionQuery struct {
    UserID     string
    StartDate  *time.Time
    EndDate    *time.Time
    Tags       []string
    Keywords   []string
    Limit      int
}

// File-based memory store implementation
type FileMemoryStore struct {
    baseDir string
}

func NewFileMemoryStore(baseDir string) *FileMemoryStore {
    os.MkdirAll(baseDir, 0755)
    return &FileMemoryStore{baseDir: baseDir}
}

func (fms *FileMemoryStore) SaveSession(session *Session) error {
    sessionPath := filepath.Join(fms.baseDir, session.UserID, session.ID+".json")
    
    // Create user directory if it doesn't exist
    userDir := filepath.Dir(sessionPath)
    if err := os.MkdirAll(userDir, 0755); err != nil {
        return fmt.Errorf("failed to create user directory: %w", err)
    }
    
    // Update last active time
    session.LastActive = time.Now()
    
    // Serialize session
    data, err := json.MarshalIndent(session, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal session: %w", err)
    }
    
    // Write to file
    if err := os.WriteFile(sessionPath, data, 0644); err != nil {
        return fmt.Errorf("failed to write session file: %w", err)
    }
    
    fmt.Printf("💾 Session saved: %s\n", sessionPath)
    return nil
}

func (fms *FileMemoryStore) LoadSession(sessionID string) (*Session, error) {
    // Find session file (search all user directories)
    var sessionPath string
    err := filepath.WalkDir(fms.baseDir, func(path string, d os.DirEntry, err error) error {
        if err != nil {
            return err
        }
        if d.IsDir() {
            return nil
        }
        if d.Name() == sessionID+".json" {
            sessionPath = path
            return filepath.SkipAll
        }
        return nil
}
    
    if err != nil {
        return nil, err
    }
    
    if sessionPath == "" {
        return nil, fmt.Errorf("session not found: %s", sessionID)
    }
    
    // Read and deserialize session
    data, err := os.ReadFile(sessionPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read session file: %w", err)
    }
    
    var session Session
    if err := json.Unmarshal(data, &session); err != nil {
        return nil, fmt.Errorf("failed to unmarshal session: %w", err)
    }
    
    fmt.Printf("📂 Session loaded: %s\n", sessionPath)
    return &session, nil
}

func (fms *FileMemoryStore) ListSessions(userID string) ([]SessionSummary, error) {
    userDir := filepath.Join(fms.baseDir, userID)
    
    if _, err := os.Stat(userDir); os.IsNotExist(err) {
        return []SessionSummary{}, nil
    }
    
    entries, err := os.ReadDir(userDir)
    if err != nil {
        return nil, err
    }
    
    var summaries []SessionSummary
    for _, entry := range entries {
        if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
            sessionID := entry.Name()[:len(entry.Name())-5] // Remove .json
            session, err := fms.LoadSession(sessionID)
            if err != nil {
                continue
            }
            
            summaries = append(summaries, SessionSummary{
                ID:          session.ID,
                UserID:      session.UserID,
                StartTime:   session.StartTime,
                LastActive:  session.LastActive,
                Summary:     session.Summary,
                Tags:        session.Tags,
                MessageCount: len(session.State.GetMessages()),
}
        }
    }
    
    return summaries, nil
}

func (fms *FileMemoryStore) DeleteSession(sessionID string) error {
    // Find and delete session file
    var sessionPath string
    err := filepath.WalkDir(fms.baseDir, func(path string, d os.DirEntry, err error) error {
        if err != nil {
            return err
        }
        if d.IsDir() {
            return nil
        }
        if d.Name() == sessionID+".json" {
            sessionPath = path
            return filepath.SkipAll
        }
        return nil
}
    
    if err != nil {
        return err
    }
    
    if sessionPath == "" {
        return fmt.Errorf("session not found: %s", sessionID)
    }
    
    return os.Remove(sessionPath)
}

func (fms *FileMemoryStore) SearchSessions(query SessionQuery) ([]SessionSummary, error) {
    summaries, err := fms.ListSessions(query.UserID)
    if err != nil {
        return nil, err
    }
    
    var filtered []SessionSummary
    for _, summary := range summaries {
        // Apply filters
        if query.StartDate != nil && summary.StartTime.Before(*query.StartDate) {
            continue
        }
        if query.EndDate != nil && summary.LastActive.After(*query.EndDate) {
            continue
        }
        
        // Tag filtering
        if len(query.Tags) > 0 {
            hasTag := false
            for _, queryTag := range query.Tags {
                for _, sessionTag := range summary.Tags {
                    if sessionTag == queryTag {
                        hasTag = true
                        break
                    }
                }
                if hasTag {
                    break
                }
            }
            if !hasTag {
                continue
            }
        }
        
        filtered = append(filtered, summary)
        
        // Apply limit
        if query.Limit > 0 && len(filtered) >= query.Limit {
            break
        }
    }
    
    return filtered, nil
}

func NewSessionManager(sessionsDir string) *SessionManager {
    return &SessionManager{
        sessionsDir: sessionsDir,
        maxAge:      30 * 24 * time.Hour, // 30 days
        autoSave:    true,
    }
}

func (sm *SessionManager) CreateSession(userID string, metadata map[string]interface{}) *Session {
    sessionID := fmt.Sprintf("session_%d", time.Now().Unix())
    
    return &Session{
        ID:         sessionID,
        UserID:     userID,
        StartTime:  time.Now(),
        LastActive: time.Now(),
        State:      domain.NewState(),
        Metadata:   metadata,
        Tags:       []string{},
    }
}

func NewPersistentMemoryAgent(name, provider, sessionsDir string) (*PersistentMemoryAgent, error) {
    baseAgent, err := core.NewAgentFromString(name, provider)
    if err != nil {
        return nil, err
    }

    return &PersistentMemoryAgent{
        name:           name,
        baseAgent:      baseAgent,
        sessionManager: NewSessionManager(sessionsDir),
        memoryStore:    NewFileMemoryStore(sessionsDir),
    }, nil
}

func (pma *PersistentMemoryAgent) StartSession(userID string, metadata map[string]interface{}) (*Session, error) {
    session := pma.sessionManager.CreateSession(userID, metadata)
    pma.currentSession = session
    
    // Initialize agent state
    session.State.Set("user_id", userID)
    session.State.Set("session_id", session.ID)
    session.State.Set("agent_name", pma.name)
    
    fmt.Printf("🆕 New session started: %s for user %s\n", session.ID, userID)
    return session, nil
}

func (pma *PersistentMemoryAgent) LoadSession(sessionID string) error {
    session, err := pma.memoryStore.LoadSession(sessionID)
    if err != nil {
        return err
    }
    
    pma.currentSession = session
    fmt.Printf("📁 Session loaded: %s\n", sessionID)
    return nil
}

func (pma *PersistentMemoryAgent) ProcessInput(ctx context.Context, input string) (string, error) {
    if pma.currentSession == nil {
        return "", fmt.Errorf("no active session")
    }
    
    // Add user message to session state
    userMessage := domain.NewMessage(domain.RoleUser, input)
    userMessage.Metadata = map[string]interface{}{
        "timestamp": time.Now(),
        "session_id": pma.currentSession.ID,
    }
    pma.currentSession.State.AddMessage(userMessage)
    
    // Set input context
    pma.currentSession.State.Set("user_input", input)
    
    // Process with agent
    result, err := pma.baseAgent.Run(ctx, pma.currentSession.State)
    if err != nil {
        return "", err
    }
    
    response, exists := result.Get("response")
    if !exists {
        return "", fmt.Errorf("no response generated")
    }
    
    responseStr := response.(string)
    
    // Add assistant response
    assistantMessage := domain.NewMessage(domain.RoleAssistant, responseStr)
    assistantMessage.Metadata = map[string]interface{}{
        "timestamp": time.Now(),
        "session_id": pma.currentSession.ID,
    }
    pma.currentSession.State.AddMessage(assistantMessage)
    
    // Update session with result state
    pma.currentSession.State = result
    pma.currentSession.LastActive = time.Now()
    
    // Auto-save if enabled
    if pma.sessionManager.autoSave {
        err = pma.memoryStore.SaveSession(pma.currentSession)
        if err != nil {
            log.Printf("Failed to auto-save session: %v", err)
        }
    }
    
    return responseStr, nil
}

func (pma *PersistentMemoryAgent) SaveSession() error {
    if pma.currentSession == nil {
        return fmt.Errorf("no active session")
    }
    
    return pma.memoryStore.SaveSession(pma.currentSession)
}

func (pma *PersistentMemoryAgent) GetSessionSummary(ctx context.Context) (string, error) {
    if pma.currentSession == nil {
        return "", fmt.Errorf("no active session")
    }
    
    messages := pma.currentSession.State.GetMessages()
    if len(messages) == 0 {
        return "Empty session", nil
    }
    
    // Create summarization prompt
    var conversationText string
    for _, msg := range messages {
        if msg.Role != domain.RoleSystem {
            conversationText += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
        }
    }
    
    // Use the agent itself to summarize
    summarizeState := domain.NewState()
    summarizeState.Set("user_input", fmt.Sprintf(`Please provide a brief summary of this conversation:

%s

Summary should be 1-2 sentences highlighting the main topics and outcomes.`, conversationText))
    
    result, err := pma.baseAgent.Run(ctx, summarizeState)
    if err != nil {
        return "", err
    }
    
    summary, exists := result.Get("response")
    if !exists {
        return "Unable to generate summary", nil
    }
    
    summaryStr := summary.(string)
    pma.currentSession.Summary = summaryStr
    
    return summaryStr, nil
}

func (pma *PersistentMemoryAgent) ListUserSessions(userID string) ([]SessionSummary, error) {
    return pma.memoryStore.ListSessions(userID)
}

func main() {
    fmt.Println("🧠 Agent Memory - Persistent Memory and Sessions")
    fmt.Println("=============================================")

    // Create persistent memory agent
    agent, err := NewPersistentMemoryAgent("persistent-assistant", "anthropic/claude-3-5-haiku", "./sessions")
    if err != nil {
        log.Fatalf("Failed to create agent: %v", err)
    }

    ctx := context.Background()
    userID := "user123"

    // Start new session
    sessionMetadata := map[string]interface{}{
        "user_name": "Alice",
        "topic": "Go programming help",
        "session_type": "learning",
    }
    
    session, err := agent.StartSession(userID, sessionMetadata)
    if err != nil {
        log.Fatalf("Failed to start session: %v", err)
    }

    fmt.Printf("📝 Session %s started for user %s\n", session.ID, userID)

    // Have a conversation
    conversations := []string{
        "Hi, I'm learning Go programming. Can you help me understand interfaces?",
        "Can you show me a simple interface example?",
        "How do I implement that interface?",
        "Thank you! This is very helpful.",
    }

    for i, input := range conversations {
        fmt.Printf("\n--- Turn %d ---\n", i+1)
        fmt.Printf("👤 User: %s\n", input)
        
        response, err := agent.ProcessInput(ctx, input)
        if err != nil {
            log.Printf("Error: %v", err)
            continue
        }
        
        fmt.Printf("🤖 Assistant: %s\n", response)
    }

    // Generate session summary
    summary, err := agent.GetSessionSummary(ctx)
    if err != nil {
        log.Printf("Failed to generate summary: %v", err)
    } else {
        fmt.Printf("\n📄 Session Summary: %s\n", summary)
    }

    // Save session explicitly
    if err := agent.SaveSession(); err != nil {
        log.Printf("Failed to save session: %v", err)
    }

    // List all sessions for user
    sessions, err := agent.ListUserSessions(userID)
    if err != nil {
        log.Printf("Failed to list sessions: %v", err)
    } else {
        fmt.Printf("\n📚 User Sessions:\n")
        for i, s := range sessions {
            fmt.Printf("%d. %s - %s (%d messages)\n", i+1, s.ID, s.Summary, s.MessageCount)
        }
    }

    // Demonstrate session reload
    fmt.Printf("\n🔄 Reloading session...\n")
    newAgent, _ := NewPersistentMemoryAgent("reloaded-assistant", "anthropic/claude-3-5-haiku", "./sessions")
    
    if err := newAgent.LoadSession(session.ID); err != nil {
        log.Printf("Failed to reload session: %v", err)
    } else {
        fmt.Printf("✅ Session reloaded successfully\n")
        
        // Continue conversation from memory
        response, err := newAgent.ProcessInput(ctx, "What were we discussing earlier?")
        if err != nil {
            log.Printf("Error: %v", err)
        } else {
            fmt.Printf("🤖 Reloaded Agent: %s\n", response)
        }
    }
}
```

---

## Level 3: Advanced Memory Patterns
*Implement hierarchical memory and custom strategies*

### Episodic Memory and Hierarchical State
```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/google/uuid"
    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
)

// AdvancedMemoryAgent with episodic and hierarchical memory
type AdvancedMemoryAgent struct {
    name             string
    baseAgent        domain.BaseAgent
    workingMemory    *domain.State           // Current active state
    episodicMemory   *EpisodicMemorySystem   // Long-term experiences
    semanticMemory   *SemanticMemorySystem   // Learned facts and procedures
    sharedMemory     *domain.SharedStateContext // Hierarchical shared context
    memoryManager    *AdvancedMemoryManager
}

// Episodic Memory System - stores experiences and episodes
type EpisodicMemorySystem struct {
    episodes    map[string]*Episode
    index       *EpisodeIndex
    maxEpisodes int
}

type Episode struct {
    ID          string                 `json:"id"`
    StartTime   time.Time             `json:"start_time"`
    EndTime     time.Time             `json:"end_time"`
    Context     map[string]interface{} `json:"context"`
    Trigger     string                `json:"trigger"`
    Outcome     string                `json:"outcome"`
    Importance  float64               `json:"importance"` // 0.0-1.0
    Messages    []domain.Message      `json:"messages"`
    Tags        []string              `json:"tags"`
    Metadata    map[string]interface{} `json:"metadata"`
}

type EpisodeIndex struct {
    byTime      []*Episode
    byTag       map[string][]*Episode
    byImportance []*Episode
    byContext    map[string][]*Episode
}

// Semantic Memory System - stores learned facts and procedures
type SemanticMemorySystem struct {
    facts       map[string]*Fact
    procedures  map[string]*Procedure
    associations map[string][]string
    confidence   map[string]float64
}

type Fact struct {
    ID          string                 `json:"id"`
    Subject     string                `json:"subject"`
    Predicate   string                `json:"predicate"`
    Object      string                `json:"object"`
    Confidence  float64               `json:"confidence"`
    Source      string                `json:"source"`
    Timestamp   time.Time             `json:"timestamp"`
    Metadata    map[string]interface{} `json:"metadata"`
}

type Procedure struct {
    ID          string                 `json:"id"`
    Name        string                `json:"name"`
    Description string                `json:"description"`
    Steps       []string              `json:"steps"`
    Conditions  []string              `json:"conditions"`
    Examples    []string              `json:"examples"`
    SuccessRate float64               `json:"success_rate"`
    Metadata    map[string]interface{} `json:"metadata"`
}

// Advanced Memory Manager coordinates different memory systems
type AdvancedMemoryManager struct {
    workingMemoryLimit    int           // Max items in working memory
    episodeRetentionDays  int           // How long to keep episodes
    factConfidenceThreshold float64      // Minimum confidence for facts
    consolidationInterval time.Duration // How often to consolidate memories
}

func NewEpisodicMemorySystem(maxEpisodes int) *EpisodicMemorySystem {
    return &EpisodicMemorySystem{
        episodes:    make(map[string]*Episode),
        index:       &EpisodeIndex{
            byTag:       make(map[string][]*Episode),
            byContext:   make(map[string][]*Episode),
        },
        maxEpisodes: maxEpisodes,
    }
}

func (ems *EpisodicMemorySystem) CreateEpisode(trigger string, context map[string]interface{}) *Episode {
    episode := &Episode{
        ID:          uuid.New().String(),
        StartTime:   time.Now(),
        Context:     context,
        Trigger:     trigger,
        Importance:  0.5, // Default importance
        Messages:    []domain.Message{},
        Tags:        []string{},
        Metadata:    make(map[string]interface{}),
    }
    
    ems.episodes[episode.ID] = episode
    ems.updateIndex(episode)
    
    fmt.Printf("📝 New episode created: %s (trigger: %s)\n", episode.ID, trigger)
    return episode
}

func (ems *EpisodicMemorySystem) CompleteEpisode(episodeID string, outcome string, importance float64) error {
    episode, exists := ems.episodes[episodeID]
    if !exists {
        return fmt.Errorf("episode not found: %s", episodeID)
    }
    
    episode.EndTime = time.Now()
    episode.Outcome = outcome
    episode.Importance = importance
    
    // Update importance-based index
    ems.rebuildImportanceIndex()
    
    fmt.Printf("✅ Episode completed: %s (importance: %.2f)\n", episodeID, importance)
    return nil
}

func (ems *EpisodicMemorySystem) AddMessageToEpisode(episodeID string, message domain.Message) error {
    episode, exists := ems.episodes[episodeID]
    if !exists {
        return fmt.Errorf("episode not found: %s", episodeID)
    }
    
    episode.Messages = append(episode.Messages, message)
    return nil
}

func (ems *EpisodicMemorySystem) SearchEpisodes(query EpisodeQuery) []*Episode {
    var results []*Episode
    
    for _, episode := range ems.episodes {
        matches := true
        
        // Time range filter
        if query.StartTime != nil && episode.StartTime.Before(*query.StartTime) {
            matches = false
        }
        if query.EndTime != nil && episode.EndTime.After(*query.EndTime) {
            matches = false
        }
        
        // Tag filter
        if len(query.Tags) > 0 {
            hasTag := false
            for _, queryTag := range query.Tags {
                for _, episodeTag := range episode.Tags {
                    if episodeTag == queryTag {
                        hasTag = true
                        break
                    }
                }
                if hasTag {
                    break
                }
            }
            if !hasTag {
                matches = false
            }
        }
        
        // Importance filter
        if query.MinImportance > 0 && episode.Importance < query.MinImportance {
            matches = false
        }
        
        // Context filter
        if len(query.Context) > 0 {
            for key, value := range query.Context {
                if episodeValue, exists := episode.Context[key]; !exists || episodeValue != value {
                    matches = false
                    break
                }
            }
        }
        
        if matches {
            results = append(results, episode)
        }
        
        // Apply limit
        if query.Limit > 0 && len(results) >= query.Limit {
            break
        }
    }
    
    return results
}

func (ems *EpisodicMemorySystem) updateIndex(episode *Episode) {
    // Add to time index
    ems.index.byTime = append(ems.index.byTime, episode)
    
    // Add to tag index
    for _, tag := range episode.Tags {
        ems.index.byTag[tag] = append(ems.index.byTag[tag], episode)
    }
    
    // Add to context index
    for key := range episode.Context {
        ems.index.byContext[key] = append(ems.index.byContext[key], episode)
    }
}

func (ems *EpisodicMemorySystem) rebuildImportanceIndex() {
    ems.index.byImportance = make([]*Episode, 0, len(ems.episodes))
    for _, episode := range ems.episodes {
        ems.index.byImportance = append(ems.index.byImportance, episode)
    }
    
    // Sort by importance (descending)
    for i := 0; i < len(ems.index.byImportance); i++ {
        for j := 0; j < len(ems.index.byImportance)-1-i; j++ {
            if ems.index.byImportance[j].Importance < ems.index.byImportance[j+1].Importance {
                ems.index.byImportance[j], ems.index.byImportance[j+1] = 
                    ems.index.byImportance[j+1], ems.index.byImportance[j]
            }
        }
    }
}

type EpisodeQuery struct {
    StartTime     *time.Time
    EndTime       *time.Time
    Tags          []string
    MinImportance float64
    Context       map[string]interface{}
    Limit         int
}

func NewSemanticMemorySystem() *SemanticMemorySystem {
    return &SemanticMemorySystem{
        facts:       make(map[string]*Fact),
        procedures:  make(map[string]*Procedure),
        associations: make(map[string][]string),
        confidence:   make(map[string]float64),
    }
}

func (sms *SemanticMemorySystem) StoreFact(subject, predicate, object string, confidence float64, source string) *Fact {
    fact := &Fact{
        ID:         uuid.New().String(),
        Subject:    subject,
        Predicate:  predicate,
        Object:     object,
        Confidence: confidence,
        Source:     source,
        Timestamp:  time.Now(),
        Metadata:   make(map[string]interface{}),
    }
    
    sms.facts[fact.ID] = fact
    
    // Create associations
    sms.addAssociation(subject, object)
    sms.addAssociation(object, subject)
    
    fmt.Printf("🧠 Fact stored: %s %s %s (confidence: %.2f)\n", subject, predicate, object, confidence)
    return fact
}

func (sms *SemanticMemorySystem) StoreProcedure(name, description string, steps []string, conditions []string) *Procedure {
    procedure := &Procedure{
        ID:          uuid.New().String(),
        Name:        name,
        Description: description,
        Steps:       steps,
        Conditions:  conditions,
        Examples:    []string{},
        SuccessRate: 0.5,
        Metadata:    make(map[string]interface{}),
    }
    
    sms.procedures[procedure.ID] = procedure
    
    fmt.Printf("⚙️ Procedure stored: %s (%d steps)\n", name, len(steps))
    return procedure
}

func (sms *SemanticMemorySystem) addAssociation(from, to string) {
    if associations, exists := sms.associations[from]; exists {
        // Check if association already exists
        for _, assoc := range associations {
            if assoc == to {
                return
            }
        }
        sms.associations[from] = append(associations, to)
    } else {
        sms.associations[from] = []string{to}
    }
}

func (sms *SemanticMemorySystem) GetRelatedConcepts(concept string) []string {
    return sms.associations[concept]
}

func (sms *SemanticMemorySystem) QueryFacts(subject, predicate, object string) []*Fact {
    var results []*Fact
    
    for _, fact := range sms.facts {
        matches := true
        
        if subject != "" && fact.Subject != subject {
            matches = false
        }
        if predicate != "" && fact.Predicate != predicate {
            matches = false
        }
        if object != "" && fact.Object != object {
            matches = false
        }
        
        if matches {
            results = append(results, fact)
        }
    }
    
    return results
}

func NewAdvancedMemoryAgent(name, provider string, parentState *domain.State) (*AdvancedMemoryAgent, error) {
    baseAgent, err := core.NewAgentFromString(name, provider)
    if err != nil {
        return nil, err
    }

    workingMemory := domain.NewState()
    
    var sharedMemory *domain.SharedStateContext
    if parentState != nil {
        sharedMemory = domain.NewSharedStateContext(domain.NewStateReader(parentState))
    }

    return &AdvancedMemoryAgent{
        name:             name,
        baseAgent:        baseAgent,
        workingMemory:    workingMemory,
        episodicMemory:   NewEpisodicMemorySystem(100),
        semanticMemory:   NewSemanticMemorySystem(),
        sharedMemory:     sharedMemory,
        memoryManager:    &AdvancedMemoryManager{
            workingMemoryLimit:    20,
            episodeRetentionDays:  30,
            factConfidenceThreshold: 0.7,
            consolidationInterval: time.Hour,
        },
    }, nil
}

func (ama *AdvancedMemoryAgent) StartEpisode(trigger string, context map[string]interface{}) *Episode {
    return ama.episodicMemory.CreateEpisode(trigger, context)
}

func (ama *AdvancedMemoryAgent) ProcessInput(ctx context.Context, input string, episodeID string) (string, error) {
    fmt.Printf("🔄 Processing input with advanced memory (episode: %s)\n", episodeID)
    
    // Add user message to working memory
    userMessage := domain.NewMessage(domain.RoleUser, input)
    ama.workingMemory.AddMessage(userMessage)
    
    // Add to current episode if specified
    if episodeID != "" {
        ama.episodicMemory.AddMessageToEpisode(episodeID, userMessage)
    }
    
    // Retrieve relevant episodic memories
    relevantEpisodes := ama.retrieveRelevantEpisodes(input)
    if len(relevantEpisodes) > 0 {
        episodeContext := ama.buildEpisodeContext(relevantEpisodes)
        ama.workingMemory.Set("relevant_episodes", episodeContext)
    }
    
    // Retrieve relevant semantic memories
    relatedConcepts := ama.extractConcepts(input)
    semanticContext := ama.buildSemanticContext(relatedConcepts)
    if len(semanticContext) > 0 {
        ama.workingMemory.Set("semantic_context", semanticContext)
    }
    
    // Set current input
    ama.workingMemory.Set("user_input", input)
    
    // Use shared memory if available
    var executionState *domain.State
    if ama.sharedMemory != nil {
        executionState = ama.sharedMemory.Clone()
        
        // Merge working memory into shared context
        for key, value := range ama.workingMemory.GetAllValues() {
            executionState.Set(key, value)
        }
        
        // Merge messages
        for _, msg := range ama.workingMemory.GetMessages() {
            executionState.AddMessage(msg)
        }
    } else {
        executionState = ama.workingMemory
    }
    
    // Run agent
    result, err := ama.baseAgent.Run(ctx, executionState)
    if err != nil {
        return "", err
    }
    
    response, exists := result.Get("response")
    if !exists {
        return "", fmt.Errorf("no response generated")
    }
    
    responseStr := response.(string)
    
    // Add assistant response to working memory
    assistantMessage := domain.NewMessage(domain.RoleAssistant, responseStr)
    ama.workingMemory.AddMessage(assistantMessage)
    
    // Add to current episode
    if episodeID != "" {
        ama.episodicMemory.AddMessageToEpisode(episodeID, assistantMessage)
    }
    
    // Extract and store new knowledge
    ama.extractAndStoreKnowledge(input, responseStr)
    
    // Update working memory with results
    ama.workingMemory = result
    
    // Manage memory limits
    ama.manageMemoryLimits()
    
    return responseStr, nil
}

func (ama *AdvancedMemoryAgent) retrieveRelevantEpisodes(input string) []*Episode {
    // Simple keyword-based retrieval (in practice, you'd use embeddings)
    keywords := ama.extractKeywords(input)
    
    var relevantEpisodes []*Episode
    for _, episode := range ama.episodicMemory.episodes {
        relevance := ama.calculateEpisodeRelevance(episode, keywords)
        if relevance > 0.3 { // Threshold for relevance
            relevantEpisodes = append(relevantEpisodes, episode)
        }
    }
    
    // Sort by importance and limit results
    if len(relevantEpisodes) > 3 {
        relevantEpisodes = relevantEpisodes[:3]
    }
    
    return relevantEpisodes
}

func (ama *AdvancedMemoryAgent) extractConcepts(input string) []string {
    // Simple concept extraction (in practice, use NLP)
    return []string{"programming", "Go", "interfaces"}
}

func (ama *AdvancedMemoryAgent) extractKeywords(input string) []string {
    // Simple keyword extraction (in practice, use NLP)
    return []string{"help", "learn", "understand"}
}

func (ama *AdvancedMemoryAgent) calculateEpisodeRelevance(episode *Episode, keywords []string) float64 {
    // Simple relevance calculation
    var matches int
    for _, keyword := range keywords {
        if ama.containsKeyword(episode.Trigger, keyword) || ama.containsKeyword(episode.Outcome, keyword) {
            matches++
        }
    }
    
    relevance := float64(matches) / float64(len(keywords))
    return relevance * episode.Importance
}

func (ama *AdvancedMemoryAgent) containsKeyword(text, keyword string) bool {
    // Simple contains check (in practice, use proper text analysis)
    return len(text) > 0 && len(keyword) > 0
}

func (ama *AdvancedMemoryAgent) buildEpisodeContext(episodes []*Episode) string {
    var context string
    for i, episode := range episodes {
        context += fmt.Sprintf("Previous experience %d: %s -> %s\n", i+1, episode.Trigger, episode.Outcome)
    }
    return context
}

func (ama *AdvancedMemoryAgent) buildSemanticContext(concepts []string) map[string]interface{} {
    context := make(map[string]interface{})
    
    for _, concept := range concepts {
        // Get related concepts
        related := ama.semanticMemory.GetRelatedConcepts(concept)
        if len(related) > 0 {
            context[concept] = related
        }
        
        // Get facts about the concept
        facts := ama.semanticMemory.QueryFacts(concept, "", "")
        if len(facts) > 0 {
            context[concept+"_facts"] = facts
        }
    }
    
    return context
}

func (ama *AdvancedMemoryAgent) extractAndStoreKnowledge(input, response string) {
    // Simple knowledge extraction (in practice, use sophisticated NLP)
    if len(input) > 0 && len(response) > 0 {
        // Store a fact about this interaction
        ama.semanticMemory.StoreFact(
            "user",
            "asked_about",
            ama.extractMainTopic(input),
            0.8,
            "conversation",
        )
    }
}

func (ama *AdvancedMemoryAgent) extractMainTopic(input string) string {
    // Simple topic extraction
    return "programming"
}

func (ama *AdvancedMemoryAgent) manageMemoryLimits() {
    // Limit working memory messages
    messages := ama.workingMemory.GetMessages()
    if len(messages) > ama.memoryManager.workingMemoryLimit {
        // Keep system messages and recent messages
        var systemMessages []domain.Message
        var recentMessages []domain.Message
        
        for _, msg := range messages {
            if msg.Role == domain.RoleSystem {
                systemMessages = append(systemMessages, msg)
            } else {
                recentMessages = append(recentMessages, msg)
            }
        }
        
        keepCount := ama.memoryManager.workingMemoryLimit - len(systemMessages)
        if keepCount > 0 && len(recentMessages) > keepCount {
            recentMessages = recentMessages[len(recentMessages)-keepCount:]
        }
        
        ama.workingMemory.ClearMessages()
        for _, msg := range systemMessages {
            ama.workingMemory.AddMessage(msg)
        }
        for _, msg := range recentMessages {
            ama.workingMemory.AddMessage(msg)
        }
    }
}

func (ama *AdvancedMemoryAgent) GetMemoryStatus() AdvancedMemoryStatus {
    return AdvancedMemoryStatus{
        WorkingMemorySize:  len(ama.workingMemory.GetAllValues()),
        EpisodicMemorySize: len(ama.episodicMemory.episodes),
        SemanticFactCount:  len(ama.semanticMemory.facts),
        SemanticProcCount:  len(ama.semanticMemory.procedures),
        HasSharedMemory:    ama.sharedMemory != nil,
    }
}

type AdvancedMemoryStatus struct {
    WorkingMemorySize  int  `json:"working_memory_size"`
    EpisodicMemorySize int  `json:"episodic_memory_size"`
    SemanticFactCount  int  `json:"semantic_fact_count"`
    SemanticProcCount  int  `json:"semantic_proc_count"`
    HasSharedMemory    bool `json:"has_shared_memory"`
}

func main() {
    fmt.Println("🧠 Agent Memory - Advanced Memory Patterns")
    fmt.Println("========================================")

    // Create parent state for hierarchical memory
    parentState := domain.NewState()
    parentState.Set("organization", "AcmeCorp")
    parentState.Set("domain", "software_development")
    parentState.Set("global_context", "enterprise_learning_platform")

    // Create advanced memory agent
    agent, err := NewAdvancedMemoryAgent("advanced-assistant", "anthropic/claude-3-5-sonnet", parentState)
    if err != nil {
        log.Fatalf("Failed to create agent: %v", err)
    }

    ctx := context.Background()

    // Start a learning episode
    episode := agent.StartEpisode("user_learning_go_interfaces", map[string]interface{}{
        "topic": "Go interfaces",
        "user_level": "beginner",
        "goal": "understand interface implementation",
}

    fmt.Printf("📚 Learning episode started: %s\n", episode.ID)

    // Simulate a learning conversation
    conversations := []string{
        "Hi, I'm new to Go. Can you explain what interfaces are?",
        "Can you show me a practical example of interface usage?",
        "How do I know if my type implements an interface correctly?",
        "What are some common interface patterns in Go?",
    }

    for i, input := range conversations {
        fmt.Printf("\n--- Turn %d ---\n", i+1)
        fmt.Printf("👤 User: %s\n", input)
        
        response, err := agent.ProcessInput(ctx, input, episode.ID)
        if err != nil {
            log.Printf("Error: %v", err)
            continue
        }
        
        fmt.Printf("🤖 Assistant: %s\n", response)
        
        // Show memory status
        status := agent.GetMemoryStatus()
        fmt.Printf("🧠 Memory: Working=%d, Episodes=%d, Facts=%d, Procedures=%d, Shared=%v\n",
            status.WorkingMemorySize, status.EpisodicMemorySize, 
            status.SemanticFactCount, status.SemanticProcCount, status.HasSharedMemory)
    }

    // Complete the episode
    err = agent.episodicMemory.CompleteEpisode(episode.ID, "user_learned_go_interfaces", 0.9)
    if err != nil {
        log.Printf("Failed to complete episode: %v", err)
    }

    // Search for episodes
    query := EpisodeQuery{
        MinImportance: 0.8,
        Limit: 5,
    }
    
    importantEpisodes := agent.episodicMemory.SearchEpisodes(query)
    fmt.Printf("\n🔍 Important Episodes Found: %d\n", len(importantEpisodes))
    
    for i, ep := range importantEpisodes {
        fmt.Printf("%d. %s: %s -> %s (importance: %.2f)\n", 
            i+1, ep.ID, ep.Trigger, ep.Outcome, ep.Importance)
    }

    // Show semantic knowledge
    fmt.Printf("\n🧠 Semantic Knowledge:\n")
    facts := agent.semanticMemory.QueryFacts("user", "", "")
    for i, fact := range facts {
        fmt.Printf("%d. %s %s %s (confidence: %.2f)\n", 
            i+1, fact.Subject, fact.Predicate, fact.Object, fact.Confidence)
    }

    // Demonstrate knowledge retrieval in new conversation
    fmt.Printf("\n💡 Testing Knowledge Retrieval:\n")
    response, err := agent.ProcessInput(ctx, "What have we been discussing in our previous conversations?", episode.ID)
    if err != nil {
        log.Printf("Error: %v", err)
    } else {
        fmt.Printf("🤖 Assistant (with memory): %s\n", response)
    }
}
```

---

## Memory Strategy Patterns

### 1. Sliding Window Strategy
- **Use Case**: Real-time chat applications
- **Implementation**: Keep last N messages, preserve system prompts
- **Pros**: Constant memory usage, fast access
- **Cons**: Loses historical context

### 2. Episodic Memory Strategy  
- **Use Case**: Learning systems, complex problem solving
- **Implementation**: Store important experiences as episodes
- **Pros**: Preserves critical interactions, enables experience-based learning
- **Cons**: Requires episode importance scoring

### 3. Hierarchical Memory Strategy
- **Use Case**: Multi-agent systems, organizational contexts
- **Implementation**: Parent-child state inheritance
- **Pros**: Shared context, reduced duplication
- **Cons**: Complexity in state management

### 4. Semantic Memory Strategy
- **Use Case**: Knowledge-based systems, fact retention
- **Implementation**: Extract and store facts/procedures
- **Pros**: Accumulates knowledge, enables reasoning
- **Cons**: Requires knowledge extraction capabilities

## Best Practices

### Memory Performance Optimization
1. **Lazy Loading** - Load memory content only when needed
2. **Compression** - Compress old episodes and conversations
3. **Indexing** - Build efficient indices for memory search
4. **Caching** - Cache frequently accessed memories
5. **Pruning** - Regularly clean up low-importance memories

### Memory Security Considerations
1. **Encryption** - Encrypt sensitive memory content
2. **Access Control** - Restrict memory access by user/role
3. **Audit Trails** - Log memory access and modifications
4. **Data Retention** - Implement proper data retention policies
5. **Anonymization** - Remove PII from stored memories

### Error Handling and Recovery
1. **Graceful Degradation** - Function with reduced memory
2. **Backup Strategies** - Regular memory backups
3. **Corruption Detection** - Validate memory integrity
4. **Recovery Procedures** - Restore from backups
5. **Failover Memory** - Alternative memory stores

## Next Steps

🧠 **Agent memory mastered!** Continue with:

- **[Agent Communication](agent-communication.md)** - Coordinate agents with shared memory
- **[Structured Data](structured-data.md)** - Validate and structure memory content
- **[Performance Optimization](../advanced/performance-optimization.md)** - Optimize memory performance
- **[Production Deployment](../advanced/production-deployment.md)** - Deploy memory systems at scale

### Quick Reference

- **[Configuration Reference](../reference/configuration-reference.md)** - Memory configuration options
- **[Best Practices Checklist](../reference/best-practices-checklist.md)** - Memory management best practices
- **[Troubleshooting](../advanced/troubleshooting.md)** - Common memory issues and solutions

---

**Need help with memory strategies?** Check our [memory patterns guide](../examples/memory-patterns.md) or join the discussion on [GitHub](https://github.com/lexlapax/go-llms/discussions).