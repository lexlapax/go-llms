# Creating Agents: From Simple to Complex AI Systems

> **[Project Root](/) / [Documentation](../..) / [User Guide](../../user-guide) / [Guides](../../user-guide/guides) / Creating Agents**

Master the art of building AI agents that can think, use tools, and coordinate with other agents. Learn progressive patterns from simple conversational agents to sophisticated multi-agent systems.

## Why Build Agents?

- **Conversational Intelligence** - Create AI that remembers context and maintains state
- **Tool Integration** - Agents that can search the web, read files, and call APIs
- **Complex Reasoning** - Break down problems into steps and coordinate solutions
- **Scalable Architecture** - From single agents to multi-agent systems
- **Production Ready** - Built-in error handling, monitoring, and recovery

## Learning Path

![Agent Architecture](../../images/agent-architecture.svg)

1. **Simple Agent** - Basic conversational AI (5 min)
2. **Tool-Enabled Agent** - Agent with built-in capabilities (10 min)
3. **Stateful Agent** - Memory and context management (15 min)
4. **Custom Agent** - Specialized behavior and logic (20 min)
5. **Multi-Agent System** - Coordinated agent collaboration (25 min)

## Prerequisites

- [First Steps completed](../getting-started/first-steps.md) ✅
- [Provider setup](../getting-started/choosing-providers.md) ✅
- Basic understanding of Go interfaces ✅

---

## Level 1: Simple Conversational Agent
*Create your first AI agent in 30 seconds*

### The Pattern
A simple agent wraps an LLM provider with conversation management, system prompts, and state handling.

### Implementation
```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
)

func main() {
    // Create agent with simple string-based API
    agent, err := core.NewAgentFromString("assistant", "openai/gpt-4o-mini")
    if err != nil {
        log.Fatalf("Failed to create agent: %v", err)
    }

    // Configure the agent's personality and behavior
    agent.SetSystemPrompt(`You are a helpful coding assistant specializing in Go.
    Provide clear, concise answers with practical examples.
    Always include error handling in code examples.`)

    // Create conversation state
    state := domain.NewState()
    
    // Have a conversation
    questions := []string{
        "How do I handle errors in Go?",
        "What's the difference between make and new?",
        "Show me a simple HTTP server example",
    }

    for i, question := range questions {
        fmt.Printf("\n--- Question %d ---\n", i+1)
        fmt.Printf("You: %s\n", question)

        // Set user input
        state.Set("user_input", question)

        // Get response
        result, err := agent.Run(context.Background(), state)
        if err != nil {
            log.Printf("Error: %v", err)
            continue
        }

        if response, exists := result.Get("response"); exists {
            fmt.Printf("Agent: %v\n", response)
        }
    }
}
```

### Key Concepts
✅ **String-based Creation** - `NewAgentFromString` for quick setup  
✅ **System Prompts** - Define agent personality and behavior  
✅ **State Management** - Persistent conversation context  
✅ **Error Handling** - Graceful failure handling  

---

## Level 2: Tool-Enabled Agent
*Give your agent superpowers with built-in tools*

### The Pattern
Tool-enabled agents can interact with the real world through pre-built or custom tools.

### Implementation
```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    
    // Import tool categories
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/math"
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/file"
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/system"
)

func main() {
    // Create a research assistant agent
    agent, err := core.NewAgentFromString("research-assistant", "anthropic/claude-3-5-sonnet")
    if err != nil {
        log.Fatalf("Failed to create agent: %v", err)
    }

    // Configure agent for research tasks
    agent.SetSystemPrompt(`You are a research assistant with access to tools.
    When asked questions:
    1. Use web search to find current information
    2. Use calculator for any math operations
    3. Save important findings to files
    4. Always cite your sources
    
    Be thorough but concise. Use tools when they help provide better answers.`)

    // Add web search capability (requires API key)
    if searchKey := os.Getenv("SEARCH_API_KEY"); searchKey != "" {
        webSearch := web.NewWebSearchTool(searchKey)
        agent.AddTool(webSearch)
        fmt.Println("✓ Web search enabled")
    } else {
        fmt.Println("⚠️ Web search disabled (SEARCH_API_KEY not set)")
    }

    // Add calculator for math operations
    calculator := math.NewCalculatorTool()
    agent.AddTool(calculator)
    fmt.Println("✓ Calculator enabled")

    // Add file operations
    fileWriter := file.NewFileWriteTool()
    fileReader := file.NewFileReadTool()
    agent.AddTool(fileWriter)
    agent.AddTool(fileReader)
    fmt.Println("✓ File operations enabled")

    // Add system information
    sysInfo := system.NewSystemInfoTool()
    agent.AddTool(sysInfo)
    fmt.Println("✓ System info enabled")

    // Research tasks that utilize different tools
    tasks := []string{
        "What's the current population of Tokyo and how does it compare to New York? Save the comparison to a file called population_comparison.txt",
        "Calculate the compound interest on $10,000 invested at 7% annually for 10 years",
        "What version of Go am I running and when was it released?",
        "Research the latest developments in quantum computing and summarize key breakthroughs",
    }

    state := domain.NewState()

    for i, task := range tasks {
        fmt.Printf("\n--- Research Task %d ---\n", i+1)
        fmt.Printf("Task: %s\n", task)

        state.Set("user_input", task)

        result, err := agent.Run(context.Background(), state)
        if err != nil {
            fmt.Printf("❌ Error: %v\n", err)
            continue
        }

        if response, exists := result.Get("response"); exists {
            fmt.Printf("Result: %v\n", response)
        }

        // Show which tools were used
        if toolCalls, exists := result.Get("tool_calls"); exists {
            fmt.Printf("🔧 Tools used: %v\n", toolCalls)
        }
    }
}
```

### Tool Categories
✅ **Web Tools** - Search, fetch, scrape, HTTP requests, APIs  
✅ **Math Tools** - Calculator, statistics, financial calculations  
✅ **File Tools** - Read, write, search, list, delete files  
✅ **System Tools** - Environment variables, process info, execution  
✅ **Data Tools** - JSON, CSV, XML processing and transformation  
✅ **DateTime Tools** - Date parsing, formatting, calculations  

---

## Level 3: Stateful Agent with Memory
*Build agents that remember and learn from interactions*

### The Pattern
Stateful agents maintain conversation history, context, and can persist data between runs.

### Implementation
```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "time"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
)

// CustomerProfile represents a customer's information
type CustomerProfile struct {
    Name         string            `json:"name"`
    Email        string            `json:"email"`
    Preferences  map[string]string `json:"preferences"`
    PurchaseHistory []Purchase     `json:"purchase_history"`
    LastContact  time.Time         `json:"last_contact"`
    Notes        []string          `json:"notes"`
}

// Purchase represents a customer purchase
type Purchase struct {
    Product     string    `json:"product"`
    Amount      float64   `json:"amount"`
    Date        time.Time `json:"date"`
    Satisfaction int      `json:"satisfaction"` // 1-5 rating
}

// CustomerSupportAgent maintains customer context and history
type CustomerSupportAgent struct {
    agent         domain.BaseAgent
    profiles      map[string]*CustomerProfile // email -> profile
    sessionState  *domain.State
    agentMemory   map[string]interface{} // persistent memory
}

func NewCustomerSupportAgent() (*CustomerSupportAgent, error) {
    // Create the underlying LLM agent
    agent, err := core.NewAgentFromString("support-agent", "openai/gpt-4o")
    if err != nil {
        return nil, fmt.Errorf("failed to create agent: %w", err)
    }

    // Configure for customer support
    agent.SetSystemPrompt(`You are a professional customer support agent.
    
    Your capabilities:
    - Access customer purchase history and preferences
    - Remember previous conversations
    - Provide personalized recommendations
    - Handle complaints professionally
    - Update customer notes and preferences
    
    Always:
    - Greet returning customers by name
    - Reference their previous purchases when relevant
    - Ask clarifying questions to better help
    - Update their profile with new information
    - End with asking if there's anything else you can help with`)

    return &CustomerSupportAgent{
        agent:        agent,
        profiles:     make(map[string]*CustomerProfile),
        sessionState: domain.NewState(),
        agentMemory:  make(map[string]interface{}),
    }, nil
}

// LoadCustomerProfile loads or creates a customer profile
func (csa *CustomerSupportAgent) LoadCustomerProfile(email string) *CustomerProfile {
    if profile, exists := csa.profiles[email]; exists {
        profile.LastContact = time.Now()
        return profile
    }

    // Create new profile
    profile := &CustomerProfile{
        Email:           email,
        Preferences:     make(map[string]string),
        PurchaseHistory: make([]Purchase, 0),
        LastContact:     time.Now(),
        Notes:           make([]string, 0),
    }
    
    csa.profiles[email] = profile
    return profile
}

// HandleCustomerInteraction processes a customer interaction with context
func (csa *CustomerSupportAgent) HandleCustomerInteraction(ctx context.Context, customerEmail, message string) (string, error) {
    // Load customer profile
    profile := csa.LoadCustomerProfile(customerEmail)

    // Build context for the agent
    contextInfo := fmt.Sprintf(`Customer Context:
- Name: %s
- Email: %s
- Last Contact: %s
- Total Purchases: %d
- Preferences: %v
- Recent Notes: %v

Customer Message: %s`,
        profile.Name,
        profile.Email,
        profile.LastContact.Format("2006-01-02"),
        len(profile.PurchaseHistory),
        profile.Preferences,
        getRecentNotes(profile.Notes, 3),
        message,
    )

    // Add customer context to state
    csa.sessionState.Set("customer_context", contextInfo)
    csa.sessionState.Set("customer_email", customerEmail)
    csa.sessionState.Set("user_input", message)

    // Run the agent
    result, err := csa.agent.Run(ctx, csa.sessionState)
    if err != nil {
        return "", fmt.Errorf("agent interaction failed: %w", err)
    }

    // Extract response
    var response string
    if resp, exists := result.Get("response"); exists {
        response = fmt.Sprintf("%v", resp)
    }

    // Update customer profile based on interaction
    csa.updateCustomerProfile(profile, message, response)

    return response, nil
}

// updateCustomerProfile updates the customer profile based on the interaction
func (csa *CustomerSupportAgent) updateCustomerProfile(profile *CustomerProfile, message, response string) {
    // Add interaction note
    note := fmt.Sprintf("%s: Customer said: %s", time.Now().Format("2006-01-02 15:04"), message)
    profile.Notes = append(profile.Notes, note)

    // Extract customer name if mentioned and not already set
    if profile.Name == "" {
        // Simple name extraction (in production, use more sophisticated NLP)
        if contains(message, "my name is") || contains(message, "I'm") {
            // Extract name logic here
        }
    }

    // Update last contact
    profile.LastContact = time.Now()

    // Keep only last 10 notes to prevent memory bloat
    if len(profile.Notes) > 10 {
        profile.Notes = profile.Notes[len(profile.Notes)-10:]
    }
}

// AddPurchase adds a purchase to the customer's history
func (csa *CustomerSupportAgent) AddPurchase(customerEmail string, purchase Purchase) {
    profile := csa.LoadCustomerProfile(customerEmail)
    profile.PurchaseHistory = append(profile.PurchaseHistory, purchase)
}

// SetCustomerPreference sets a customer preference
func (csa *CustomerSupportAgent) SetCustomerPreference(customerEmail, key, value string) {
    profile := csa.LoadCustomerProfile(customerEmail)
    profile.Preferences[key] = value
}

// GetCustomerInsights provides insights about a customer
func (csa *CustomerSupportAgent) GetCustomerInsights(customerEmail string) map[string]interface{} {
    profile := csa.LoadCustomerProfile(customerEmail)
    
    totalSpent := 0.0
    avgSatisfaction := 0.0
    recentPurchases := 0

    for _, purchase := range profile.PurchaseHistory {
        totalSpent += purchase.Amount
        avgSatisfaction += float64(purchase.Satisfaction)
        
        if time.Since(purchase.Date) < 30*24*time.Hour {
            recentPurchases++
        }
    }

    if len(profile.PurchaseHistory) > 0 {
        avgSatisfaction /= float64(len(profile.PurchaseHistory))
    }

    return map[string]interface{}{
        "total_spent":          totalSpent,
        "total_purchases":      len(profile.PurchaseHistory),
        "average_satisfaction": avgSatisfaction,
        "recent_purchases":     recentPurchases,
        "last_contact":         profile.LastContact,
        "preferences":          profile.Preferences,
    }
}

func main() {
    fmt.Println("🤝 Customer Support Agent - Stateful Interactions")
    fmt.Println("================================================")

    // Create customer support agent
    supportAgent, err := NewCustomerSupportAgent()
    if err != nil {
        log.Fatalf("Failed to create support agent: %v", err)
    }

    // Simulate customer data
    customerEmail := "sarah.johnson@techstart.com"
    
    // Add some purchase history
    supportAgent.AddPurchase(customerEmail, Purchase{
        Product:      "Enterprise CRM License",
        Amount:       2500.00,
        Date:         time.Now().AddDate(0, -2, 0),
        Satisfaction: 4,
}
    
    supportAgent.AddPurchase(customerEmail, Purchase{
        Product:      "Additional User Licenses (5x)",
        Amount:       500.00,
        Date:         time.Now().AddDate(0, -1, -15),
        Satisfaction: 5,
}

    // Set customer preferences
    supportAgent.SetCustomerPreference(customerEmail, "communication_style", "technical")
    supportAgent.SetCustomerPreference(customerEmail, "preferred_contact_time", "business_hours")
    supportAgent.SetCustomerPreference(customerEmail, "industry", "fintech")

    // Simulate customer interactions
    interactions := []string{
        "Hi, I'm Sarah Johnson from TechStart. I'm having trouble with user permissions in the CRM system.",
        "The issue is that new team members can't access the customer database, even though I added them to the right group.",
        "Yes, I checked the admin panel. The permissions look correct there.",
        "That worked! Thank you. By the way, are there any new features coming that might help with sales pipeline automation?",
        "Sounds great! Can you send me information about the upcoming automation features?",
    }

    ctx := context.Background()

    fmt.Printf("Customer: %s\n", customerEmail)
    fmt.Println("---")

    for i, message := range interactions {
        fmt.Printf("\n💬 Interaction %d:\n", i+1)
        fmt.Printf("Customer: %s\n", message)

        response, err := supportAgent.HandleCustomerInteraction(ctx, customerEmail, message)
        if err != nil {
            fmt.Printf("❌ Error: %v\n", err)
            continue
        }

        fmt.Printf("Agent: %s\n", response)
    }

    // Show customer insights
    fmt.Println("\n📊 Customer Insights:")
    insights := supportAgent.GetCustomerInsights(customerEmail)
    for key, value := range insights {
        fmt.Printf("  %s: %v\n", key, value)
    }

    // Show customer profile
    profile := supportAgent.LoadCustomerProfile(customerEmail)
    fmt.Println("\n👤 Customer Profile:")
    profileJSON, _ := json.MarshalIndent(profile, "", "  ")
    fmt.Println(string(profileJSON))
}

// Helper functions
func getRecentNotes(notes []string, limit int) []string {
    if len(notes) <= limit {
        return notes
    }
    return notes[len(notes)-limit:]
}

func contains(str, substr string) bool {
    return len(str) >= len(substr) && 
           (str == substr || 
            (len(str) > len(substr) && 
             (str[:len(substr)] == substr || 
              str[len(str)-len(substr):] == substr ||
              (len(str) > len(substr)*2 && 
               findSubstring(str, substr)))))
}

func findSubstring(str, substr string) bool {
    for i := 0; i <= len(str)-len(substr); i++ {
        if str[i:i+len(substr)] == substr {
            return true
        }
    }
    return false
}
```

### Memory Features
✅ **Customer Profiles** - Persistent customer information and history  
✅ **Conversation Context** - Remember previous interactions  
✅ **Preferences** - Learn and remember customer preferences  
✅ **Purchase History** - Track customer transaction history  
✅ **Insights** - Generate customer analytics and patterns  

---

## Level 4: Custom Agent with Specialized Logic
*Build agents with custom behavior and advanced orchestration*

### The Pattern
Custom agents extend the base agent implementation to add specialized logic, validation, and coordination.

### Implementation
```go
package main

import (
    "context"
    "fmt"
    "log"
    "strings"
    "sync"
    "time"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
)

// ResearchQuality represents the quality of research findings
type ResearchQuality int

const (
    QualityPoor ResearchQuality = iota
    QualityFair
    QualityGood
    QualityExcellent
)

// ResearchFinding represents a single research finding
type ResearchFinding struct {
    Topic       string            `json:"topic"`
    Source      string            `json:"source"`
    Content     string            `json:"content"`
    Reliability float64           `json:"reliability"` // 0-1 score
    Timestamp   time.Time         `json:"timestamp"`
    Keywords    []string          `json:"keywords"`
    Quality     ResearchQuality   `json:"quality"`
    Metadata    map[string]string `json:"metadata"`
}

// ResearchReport represents a comprehensive research report
type ResearchReport struct {
    Topic          string             `json:"topic"`
    Findings       []ResearchFinding  `json:"findings"`
    Summary        string             `json:"summary"`
    Conclusions    []string           `json:"conclusions"`
    Recommendations []string          `json:"recommendations"`
    Confidence     float64            `json:"confidence"`
    Sources        []string           `json:"sources"`
    GeneratedAt    time.Time          `json:"generated_at"`
}

// IntelligentResearchAgent performs multi-stage research with quality validation
type IntelligentResearchAgent struct {
    *core.BaseAgentImpl
    searcher        domain.BaseAgent
    analyzer        domain.BaseAgent
    summarizer      domain.BaseAgent
    validator       domain.BaseAgent
    findings        []ResearchFinding
    qualityThreshold float64
    maxRetries      int
    mu              sync.RWMutex
}

// NewIntelligentResearchAgent creates a new intelligent research agent
func NewIntelligentResearchAgent(name string) (*IntelligentResearchAgent, error) {
    // Create base agent
    base := core.NewBaseAgent(name, "An intelligent research agent with quality validation", domain.AgentTypeCustom)

    // Create specialized sub-agents
    searcher, err := core.NewAgentFromString("searcher", "openai/gpt-4o-mini")
    if err != nil {
        return nil, fmt.Errorf("failed to create searcher: %w", err)
    }
    searcher.SetSystemPrompt(`You are a research search specialist.
    Generate comprehensive search queries for the given topic.
    Return 3-5 diverse search queries that will find different aspects of the topic.
    Include specific terms, academic sources, and recent developments.`)

    analyzer, err := core.NewAgentFromString("analyzer", "anthropic/claude-3-5-sonnet")
    if err != nil {
        return nil, fmt.Errorf("failed to create analyzer: %w", err)
    }
    analyzer.SetSystemPrompt(`You are a research content analyzer.
    Analyze search results and extract key insights.
    Focus on:
    - Factual accuracy and source credibility
    - Novel insights and important findings
    - Contradictions or controversies
    - Practical implications
    
    Rate the reliability of each finding from 0-1.`)

    summarizer, err := core.NewAgentFromString("summarizer", "openai/gpt-4o")
    if err != nil {
        return nil, fmt.Errorf("failed to create summarizer: %w", err)
    }
    summarizer.SetSystemPrompt(`You are a research report writer.
    Create comprehensive summaries from research findings.
    Include:
    - Executive summary
    - Key findings with evidence
    - Conclusions and implications
    - Actionable recommendations
    
    Be objective and cite sources appropriately.`)

    validator, err := core.NewAgentFromString("validator", "anthropic/claude-3-5-haiku")
    if err != nil {
        return nil, fmt.Errorf("failed to create validator: %w", err)
    }
    validator.SetSystemPrompt(`You are a research quality validator.
    Assess the quality and completeness of research findings.
    Check for:
    - Source diversity and credibility
    - Factual consistency
    - Bias or gaps in coverage
    - Evidence quality
    
    Provide a quality score and improvement suggestions.`)

    return &IntelligentResearchAgent{
        BaseAgentImpl:    base,
        searcher:         searcher,
        analyzer:         analyzer,
        summarizer:       summarizer,
        validator:        validator,
        findings:         make([]ResearchFinding, 0),
        qualityThreshold: 0.7,
        maxRetries:       3,
    }, nil
}

// Run performs intelligent research on a given topic
func (ira *IntelligentResearchAgent) Run(ctx context.Context, state domain.StateReader) (*domain.State, error) {
    ira.mu.Lock()
    defer ira.mu.Unlock()

    topic, exists := state.Get("user_input")
    if !exists {
        return nil, fmt.Errorf("no research topic provided")
    }

    topicStr := fmt.Sprintf("%v", topic)
    
    log.Printf("🔍 Starting intelligent research on: %s", topicStr)

    // Stage 1: Generate search queries
    searchQueries, err := ira.generateSearchQueries(ctx, topicStr)
    if err != nil {
        return nil, fmt.Errorf("failed to generate search queries: %w", err)
    }

    // Stage 2: Perform searches and collect findings
    for _, query := range searchQueries {
        findings, err := ira.searchAndAnalyze(ctx, query, topicStr)
        if err != nil {
            log.Printf("⚠️ Search failed for query '%s': %v", query, err)
            continue
        }
        ira.findings = append(ira.findings, findings...)
    }

    // Stage 3: Validate quality and retry if needed
    quality := ira.assessQuality()
    retries := 0
    
    for quality < ira.qualityThreshold && retries < ira.maxRetries {
        log.Printf("📊 Research quality (%.2f) below threshold (%.2f), retrying...", quality, ira.qualityThreshold)
        
        // Identify gaps and research further
        gaps := ira.identifyKnowledgeGaps(topicStr)
        for _, gap := range gaps {
            findings, err := ira.searchAndAnalyze(ctx, gap, topicStr)
            if err != nil {
                continue
            }
            ira.findings = append(ira.findings, findings...)
        }
        
        quality = ira.assessQuality()
        retries++
    }

    // Stage 4: Generate comprehensive report
    report, err := ira.generateReport(ctx, topicStr)
    if err != nil {
        return nil, fmt.Errorf("failed to generate report: %w", err)
    }

    // Return results
    result := domain.NewState()
    result.Set("response", report.Summary)
    result.Set("research_report", report)
    result.Set("findings_count", len(ira.findings))
    result.Set("research_quality", quality)
    result.Set("retries_used", retries)

    return result, nil
}

// generateSearchQueries creates diverse search queries for the topic
func (ira *IntelligentResearchAgent) generateSearchQueries(ctx context.Context, topic string) ([]string, error) {
    state := domain.NewState()
    state.Set("user_input", fmt.Sprintf("Generate 4 diverse search queries for researching: %s", topic))

    result, err := ira.searcher.Run(ctx, state)
    if err != nil {
        return nil, err
    }

    response, exists := result.Get("response")
    if !exists {
        return nil, fmt.Errorf("no search queries generated")
    }

    // Parse queries from response (simplified)
    queries := strings.Split(strings.TrimSpace(fmt.Sprintf("%v", response)), "\n")
    filteredQueries := make([]string, 0)
    
    for _, query := range queries {
        query = strings.TrimSpace(query)
        if len(query) > 10 { // Basic filter
            filteredQueries = append(filteredQueries, query)
        }
    }

    return filteredQueries, nil
}

// searchAndAnalyze performs search and analysis for a query
func (ira *IntelligentResearchAgent) searchAndAnalyze(ctx context.Context, query, topic string) ([]ResearchFinding, error) {
    // In a real implementation, this would use web search tools
    // For this example, we'll simulate the process
    
    findings := make([]ResearchFinding, 0)
    
    // Simulate search results
    mockResults := []string{
        fmt.Sprintf("Research finding about %s from academic source", topic),
        fmt.Sprintf("Industry analysis of %s trends", topic),
        fmt.Sprintf("Recent developments in %s technology", topic),
    }

    for i, content := range mockResults {
        // Analyze each result
        state := domain.NewState()
        state.Set("user_input", fmt.Sprintf("Analyze this research content for topic '%s': %s", topic, content))

        result, err := ira.analyzer.Run(ctx, state)
        if err != nil {
            continue
        }

        if response, exists := result.Get("response"); exists {
            finding := ResearchFinding{
                Topic:       topic,
                Source:      fmt.Sprintf("source-%d", i+1),
                Content:     fmt.Sprintf("%v", response),
                Reliability: 0.8, // Would be calculated based on source analysis
                Timestamp:   time.Now(),
                Keywords:    []string{topic, "research", "analysis"},
                Quality:     QualityGood,
                Metadata:    map[string]string{"query": query},
            }
            findings = append(findings, finding)
        }
    }

    return findings, nil
}

// assessQuality calculates the overall quality of collected research
func (ira *IntelligentResearchAgent) assessQuality() float64 {
    if len(ira.findings) == 0 {
        return 0.0
    }

    totalReliability := 0.0
    diversitySources := make(map[string]bool)
    
    for _, finding := range ira.findings {
        totalReliability += finding.Reliability
        diversitySources[finding.Source] = true
    }

    avgReliability := totalReliability / float64(len(ira.findings))
    diversityScore := float64(len(diversitySources)) / float64(len(ira.findings))
    
    // Combined quality score
    return (avgReliability * 0.7) + (diversityScore * 0.3)
}

// identifyKnowledgeGaps finds areas that need more research
func (ira *IntelligentResearchAgent) identifyKnowledgeGaps(topic string) []string {
    // Simplified gap identification
    gaps := []string{
        fmt.Sprintf("%s recent developments", topic),
        fmt.Sprintf("%s case studies", topic),
        fmt.Sprintf("%s expert opinions", topic),
    }
    return gaps
}

// generateReport creates a comprehensive research report
func (ira *IntelligentResearchAgent) generateReport(ctx context.Context, topic string) (*ResearchReport, error) {
    // Combine all findings into a summary request
    findingsText := ""
    sources := make([]string, 0)
    
    for _, finding := range ira.findings {
        findingsText += fmt.Sprintf("Finding: %s (Source: %s, Reliability: %.2f)\n", 
                                   finding.Content, finding.Source, finding.Reliability)
        sources = append(sources, finding.Source)
    }

    state := domain.NewState()
    state.Set("user_input", fmt.Sprintf(`Create a comprehensive research report for: %s

Research Findings:
%s

Include:
1. Executive summary
2. Key findings with evidence
3. Conclusions
4. Practical recommendations`, topic, findingsText))

    result, err := ira.summarizer.Run(ctx, state)
    if err != nil {
        return nil, err
    }

    summary := ""
    if response, exists := result.Get("response"); exists {
        summary = fmt.Sprintf("%v", response)
    }

    // Create report
    report := &ResearchReport{
        Topic:          topic,
        Findings:       ira.findings,
        Summary:        summary,
        Conclusions:    []string{"Based on research findings...", "Evidence suggests..."},
        Recommendations: []string{"Recommend further investigation...", "Consider implementation..."},
        Confidence:     ira.assessQuality(),
        Sources:        sources,
        GeneratedAt:    time.Now(),
    }

    return report, nil
}

func main() {
    fmt.Println("🧠 Intelligent Research Agent - Custom Logic")
    fmt.Println("===========================================")

    // Create intelligent research agent
    researchAgent, err := NewIntelligentResearchAgent("research-ai")
    if err != nil {
        log.Fatalf("Failed to create research agent: %v", err)
    }

    // Research topics
    topics := []string{
        "artificial intelligence in healthcare",
        "quantum computing applications",
        "sustainable energy technologies",
    }

    for i, topic := range topics {
        fmt.Printf("\n--- Research Project %d ---\n", i+1)
        fmt.Printf("Topic: %s\n", topic)

        state := domain.NewState()
        state.Set("user_input", topic)

        startTime := time.Now()
        result, err := researchAgent.Run(context.Background(), state)
        duration := time.Since(startTime)

        if err != nil {
            fmt.Printf("❌ Research failed: %v\n", err)
            continue
        }

        // Display results
        if summary, exists := result.Get("response"); exists {
            fmt.Printf("📋 Summary: %v\n", summary)
        }

        if count, exists := result.Get("findings_count"); exists {
            fmt.Printf("📊 Findings collected: %v\n", count)
        }

        if quality, exists := result.Get("research_quality"); exists {
            fmt.Printf("⭐ Research quality: %.2f\n", quality)
        }

        if retries, exists := result.Get("retries_used"); exists {
            fmt.Printf("🔄 Retries used: %v\n", retries)
        }

        fmt.Printf("⏱️ Duration: %v\n", duration)
    }
}
```

### Custom Agent Features
✅ **Multi-Stage Processing** - Orchestrate multiple specialized sub-agents  
✅ **Quality Validation** - Assess and improve research quality automatically  
✅ **Adaptive Behavior** - Retry and improve based on quality metrics  
✅ **State Management** - Maintain complex internal state and findings  
✅ **Specialized Logic** - Custom business logic and validation rules  

---

## Level 5: Multi-Agent System
*Coordinate multiple agents for complex problem solving*

### The Pattern
Multi-agent systems distribute work across specialized agents that communicate and coordinate to solve complex problems.

### Implementation
```go
package main

import (
    "context"
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/agent/workflow"
)

// TaskType represents different types of tasks in the system
type TaskType string

const (
    TaskTypeAnalysis     TaskType = "analysis"
    TaskTypeResearch     TaskType = "research"
    TaskTypeWriting      TaskType = "writing"
    TaskTypeReview       TaskType = "review"
    TaskTypeCoordination TaskType = "coordination"
)

// TaskPriority represents task priority levels
type TaskPriority int

const (
    PriorityLow TaskPriority = iota
    PriorityNormal
    PriorityHigh
    PriorityUrgent
)

// Task represents a unit of work in the multi-agent system
type Task struct {
    ID          string       `json:"id"`
    Type        TaskType     `json:"type"`
    Priority    TaskPriority `json:"priority"`
    Description string       `json:"description"`
    Input       interface{}  `json:"input"`
    Output      interface{}  `json:"output"`
    Status      string       `json:"status"`
    AssignedTo  string       `json:"assigned_to"`
    CreatedAt   time.Time    `json:"created_at"`
    CompletedAt *time.Time   `json:"completed_at,omitempty"`
    Dependencies []string    `json:"dependencies"`
}

// AgentCapability represents what an agent can do
type AgentCapability struct {
    TaskType    TaskType `json:"task_type"`
    MaxConcurrency int  `json:"max_concurrency"`
    AvgDuration time.Duration `json:"avg_duration"`
    SuccessRate float64 `json:"success_rate"`
}

// MultiAgentOrchestrator coordinates multiple specialized agents
type MultiAgentOrchestrator struct {
    agents           map[string]domain.BaseAgent
    capabilities     map[string][]AgentCapability
    taskQueue        chan *Task
    completedTasks   map[string]*Task
    activeTasks      map[string]*Task
    coordinatorAgent domain.BaseAgent
    mu               sync.RWMutex
    shutdown         chan bool
    wg               sync.WaitGroup
}

// NewMultiAgentOrchestrator creates a new multi-agent orchestrator
func NewMultiAgentOrchestrator() (*MultiAgentOrchestrator, error) {
    orchestrator := &MultiAgentOrchestrator{
        agents:         make(map[string]domain.BaseAgent),
        capabilities:   make(map[string][]AgentCapability),
        taskQueue:      make(chan *Task, 100),
        completedTasks: make(map[string]*Task),
        activeTasks:    make(map[string]*Task),
        shutdown:       make(chan bool),
    }

    // Create coordinator agent
    coordinator, err := core.NewAgentFromString("coordinator", "anthropic/claude-3-5-sonnet")
    if err != nil {
        return nil, fmt.Errorf("failed to create coordinator: %w", err)
    }
    coordinator.SetSystemPrompt(`You are the coordinator of a multi-agent system.
    Your responsibilities:
    - Break down complex requests into smaller tasks
    - Assign tasks to appropriate specialists
    - Coordinate task dependencies
    - Synthesize results from multiple agents
    - Ensure quality and consistency
    
    Available agent types:
    - analyst: Data analysis and insights
    - researcher: Information gathering and research
    - writer: Content creation and documentation
    - reviewer: Quality assurance and validation`)

    orchestrator.coordinatorAgent = coordinator

    // Create specialized agents
    err = orchestrator.createSpecializedAgents()
    if err != nil {
        return nil, fmt.Errorf("failed to create specialized agents: %w", err)
    }

    // Start task processing
    orchestrator.startTaskProcessing()

    return orchestrator, nil
}

// createSpecializedAgents creates the specialized agent workforce
func (mao *MultiAgentOrchestrator) createSpecializedAgents() error {
    // Data Analyst Agent
    analyst, err := core.NewAgentFromString("analyst", "openai/gpt-4o")
    if err != nil {
        return err
    }
    analyst.SetSystemPrompt(`You are a data analysis specialist.
    Your expertise includes:
    - Statistical analysis and pattern recognition
    - Data visualization recommendations
    - Trend identification and forecasting
    - Performance metrics and KPIs
    - Hypothesis testing and validation
    
    Provide clear, actionable insights with supporting evidence.`)

    mao.agents["analyst"] = analyst
    mao.capabilities["analyst"] = []AgentCapability{
        {TaskType: TaskTypeAnalysis, MaxConcurrency: 3, AvgDuration: 2 * time.Minute, SuccessRate: 0.92},
    }

    // Research Specialist Agent
    researcher, err := core.NewAgentFromString("researcher", "anthropic/claude-3-5-haiku")
    if err != nil {
        return err
    }
    researcher.SetSystemPrompt(`You are a research specialist.
    Your expertise includes:
    - Information gathering and verification
    - Source evaluation and credibility assessment
    - Competitive analysis and market research
    - Academic and industry research
    - Fact-checking and validation
    
    Provide comprehensive, well-sourced research findings.`)

    mao.agents["researcher"] = researcher
    mao.capabilities["researcher"] = []AgentCapability{
        {TaskType: TaskTypeResearch, MaxConcurrency: 2, AvgDuration: 3 * time.Minute, SuccessRate: 0.88},
    }

    // Content Writer Agent
    writer, err := core.NewAgentFromString("writer", "openai/gpt-4o")
    if err != nil {
        return err
    }
    writer.SetSystemPrompt(`You are a professional content writer.
    Your expertise includes:
    - Clear, engaging content creation
    - Technical documentation and user guides
    - Marketing copy and communications
    - Report writing and summarization
    - Content adaptation for different audiences
    
    Create well-structured, audience-appropriate content.`)

    mao.agents["writer"] = writer
    mao.capabilities["writer"] = []AgentCapability{
        {TaskType: TaskTypeWriting, MaxConcurrency: 2, AvgDuration: 4 * time.Minute, SuccessRate: 0.95},
    }

    // Quality Reviewer Agent
    reviewer, err := core.NewAgentFromString("reviewer", "anthropic/claude-3-5-sonnet")
    if err != nil {
        return err
    }
    reviewer.SetSystemPrompt(`You are a quality assurance specialist.
    Your expertise includes:
    - Content review and fact-checking
    - Consistency and style validation
    - Error detection and correction
    - Compliance and standard verification
    - Final quality assessment
    
    Ensure all deliverables meet high standards.`)

    mao.agents["reviewer"] = reviewer
    mao.capabilities["reviewer"] = []AgentCapability{
        {TaskType: TaskTypeReview, MaxConcurrency: 4, AvgDuration: 1 * time.Minute, SuccessRate: 0.97},
    }

    return nil
}

// startTaskProcessing starts the task processing workers
func (mao *MultiAgentOrchestrator) startTaskProcessing() {
    for agentName := range mao.agents {
        capabilities := mao.capabilities[agentName]
        for _, capability := range capabilities {
            // Start workers for each capability
            for i := 0; i < capability.MaxConcurrency; i++ {
                mao.wg.Add(1)
                go mao.taskWorker(agentName, capability.TaskType)
            }
        }
    }
}

// taskWorker processes tasks for a specific agent and task type
func (mao *MultiAgentOrchestrator) taskWorker(agentName string, taskType TaskType) {
    defer mao.wg.Done()

    agent := mao.agents[agentName]

    for {
        select {
        case task := <-mao.taskQueue:
            if task.Type != taskType {
                // Put back in queue if not our task type
                mao.taskQueue <- task
                continue
            }

            // Process the task
            mao.processTask(agent, task)

        case <-mao.shutdown:
            return
        }
    }
}

// processTask processes a single task with the assigned agent
func (mao *MultiAgentOrchestrator) processTask(agent domain.BaseAgent, task *Task) {
    mao.mu.Lock()
    task.Status = "processing"
    task.AssignedTo = agent.Name()
    mao.activeTasks[task.ID] = task
    mao.mu.Unlock()

    log.Printf("🏃 Processing task %s (%s) with agent %s", task.ID, task.Type, agent.Name())

    // Create state for the task
    state := domain.NewState()
    state.Set("user_input", task.Description)
    state.Set("task_input", task.Input)

    // Run the agent
    result, err := agent.Run(context.Background(), state)
    
    mao.mu.Lock()
    defer mao.mu.Unlock()

    if err != nil {
        task.Status = "failed"
        task.Output = fmt.Sprintf("Error: %v", err)
        log.Printf("❌ Task %s failed: %v", task.ID, err)
    } else {
        task.Status = "completed"
        if output, exists := result.Get("response"); exists {
            task.Output = output
        }
        now := time.Now()
        task.CompletedAt = &now
        log.Printf("✅ Task %s completed", task.ID)
    }

    // Move from active to completed
    delete(mao.activeTasks, task.ID)
    mao.completedTasks[task.ID] = task
}

// ProcessRequest processes a complex request by breaking it into tasks
func (mao *MultiAgentOrchestrator) ProcessRequest(ctx context.Context, request string) (string, error) {
    log.Printf("📝 Processing complex request: %s", request)

    // Use coordinator to break down the request
    state := domain.NewState()
    state.Set("user_input", fmt.Sprintf(`Break down this request into specific tasks for our agents:
    
Request: %s

Create a task breakdown with:
1. Research tasks for information gathering
2. Analysis tasks for data processing
3. Writing tasks for content creation
4. Review tasks for quality assurance

Format each task with: TaskType | Description | Dependencies`, request))

    result, err := mao.coordinatorAgent.Run(ctx, state)
    if err != nil {
        return "", fmt.Errorf("coordinator failed: %w", err)
    }

    var breakdown string
    if response, exists := result.Get("response"); exists {
        breakdown = fmt.Sprintf("%v", response)
    }

    // Parse the breakdown and create tasks (simplified for example)
    tasks := mao.parseTaskBreakdown(breakdown)

    // Submit tasks to the queue
    for _, task := range tasks {
        mao.taskQueue <- task
    }

    // Wait for all tasks to complete
    mao.waitForTaskCompletion(tasks)

    // Coordinate final results
    finalResult, err := mao.synthesizeResults(ctx, request, tasks)
    if err != nil {
        return "", fmt.Errorf("result synthesis failed: %w", err)
    }

    return finalResult, nil
}

// parseTaskBreakdown parses the coordinator's task breakdown (simplified)
func (mao *MultiAgentOrchestrator) parseTaskBreakdown(breakdown string) []*Task {
    // In a real implementation, this would be more sophisticated
    tasks := []*Task{
        {
            ID:          "task-1",
            Type:        TaskTypeResearch,
            Priority:    PriorityNormal,
            Description: "Research current market trends",
            CreatedAt:   time.Now(),
            Status:      "pending",
        },
        {
            ID:          "task-2",
            Type:        TaskTypeAnalysis,
            Priority:    PriorityNormal,
            Description: "Analyze research data for patterns",
            Dependencies: []string{"task-1"},
            CreatedAt:   time.Now(),
            Status:      "pending",
        },
        {
            ID:          "task-3",
            Type:        TaskTypeWriting,
            Priority:    PriorityNormal,
            Description: "Create summary report",
            Dependencies: []string{"task-2"},
            CreatedAt:   time.Now(),
            Status:      "pending",
        },
        {
            ID:          "task-4",
            Type:        TaskTypeReview,
            Priority:    PriorityHigh,
            Description: "Review and validate final report",
            Dependencies: []string{"task-3"},
            CreatedAt:   time.Now(),
            Status:      "pending",
        },
    }

    return tasks
}

// waitForTaskCompletion waits for all tasks to complete
func (mao *MultiAgentOrchestrator) waitForTaskCompletion(tasks []*Task) {
    for {
        allCompleted := true
        
        mao.mu.RLock()
        for _, task := range tasks {
            if completedTask, exists := mao.completedTasks[task.ID]; !exists || completedTask.Status != "completed" {
                allCompleted = false
                break
            }
        }
        mao.mu.RUnlock()

        if allCompleted {
            break
        }

        time.Sleep(500 * time.Millisecond)
    }
}

// synthesizeResults combines results from all tasks
func (mao *MultiAgentOrchestrator) synthesizeResults(ctx context.Context, originalRequest string, tasks []*Task) (string, error) {
    // Gather all task outputs
    var allOutputs []string
    
    mao.mu.RLock()
    for _, task := range tasks {
        if completedTask, exists := mao.completedTasks[task.ID]; exists {
            allOutputs = append(allOutputs, fmt.Sprintf("Task %s (%s): %v", task.ID, task.Type, completedTask.Output))
        }
    }
    mao.mu.RUnlock()

    // Use coordinator to synthesize final result
    state := domain.NewState()
    state.Set("user_input", fmt.Sprintf(`Synthesize these task results into a comprehensive response for the original request:

Original Request: %s

Task Results:
%s

Create a coherent, well-structured final response.`, originalRequest, allOutputs))

    result, err := mao.coordinatorAgent.Run(ctx, state)
    if err != nil {
        return "", err
    }

    if response, exists := result.Get("response"); exists {
        return fmt.Sprintf("%v", response), nil
    }

    return "No response generated", nil
}

// Shutdown gracefully shuts down the orchestrator
func (mao *MultiAgentOrchestrator) Shutdown() {
    close(mao.shutdown)
    mao.wg.Wait()
}

func main() {
    fmt.Println("🤝 Multi-Agent Orchestrator - Collaborative Intelligence")
    fmt.Println("=======================================================")

    // Create multi-agent orchestrator
    orchestrator, err := NewMultiAgentOrchestrator()
    if err != nil {
        log.Fatalf("Failed to create orchestrator: %v", err)
    }
    defer orchestrator.Shutdown()

    // Complex requests that require multiple agents
    requests := []string{
        "Analyze the current state of artificial intelligence in healthcare, including market trends, key players, challenges, and create a strategic recommendations report",
        "Research sustainable energy technologies, analyze their economic viability, and write a business case for adoption",
        "Study the impact of remote work on productivity, gather supporting data, and create implementation guidelines for companies",
    }

    for i, request := range requests {
        fmt.Printf("\n--- Complex Request %d ---\n", i+1)
        fmt.Printf("Request: %s\n", request)
        fmt.Println("---")

        startTime := time.Now()
        response, err := orchestrator.ProcessRequest(context.Background(), request)
        duration := time.Since(startTime)

        if err != nil {
            fmt.Printf("❌ Request failed: %v\n", err)
            continue
        }

        fmt.Printf("📋 Final Response:\n%s\n", response)
        fmt.Printf("⏱️ Total Duration: %v\n", duration)
    }
}
```

### Multi-Agent Features
✅ **Task Orchestration** - Break complex problems into manageable tasks  
✅ **Specialized Agents** - Different agents with specific capabilities  
✅ **Dependency Management** - Handle task dependencies and sequencing  
✅ **Concurrent Processing** - Multiple agents working simultaneously  
✅ **Result Synthesis** - Combine outputs into coherent final results  

## Best Practices

### 1. Agent Design
- **Single Responsibility** - Each agent should have a clear, focused purpose
- **System Prompts** - Write detailed, specific system prompts for behavior
- **Error Handling** - Implement graceful error handling and recovery
- **State Management** - Use state objects for conversation context

### 2. Tool Integration
- **Progressive Enhancement** - Start simple, add tools as needed
- **Environment Checks** - Gracefully handle missing API keys or tools
- **Tool Selection** - Choose tools that match your use case
- **Custom Tools** - Build custom tools for specialized needs

### 3. Performance
- **Provider Selection** - Choose appropriate models for different tasks
- **Concurrency** - Use concurrent processing for independent tasks
- **Caching** - Cache common operations and responses
- **Monitoring** - Track performance and error rates

### 4. Production
- **Validation** - Validate inputs and outputs rigorously
- **Logging** - Comprehensive logging for debugging and monitoring
- **Configuration** - Make agents configurable for different environments
- **Testing** - Test agent behavior with various inputs and scenarios

## Troubleshooting

### Common Issues

**Agent Not Responding**
- Check provider API keys and connection
- Verify system prompts are clear and specific
- Ensure input state has required fields
- Check for rate limiting or quota issues

**Poor Agent Performance**
- Improve system prompt specificity
- Add examples and constraints
- Use more capable models for complex tasks
- Implement validation and retry logic

**Tool Integration Problems**
- Verify tool requirements and dependencies
- Check environment variables and configuration
- Test tools independently before integration
- Handle tool failures gracefully

**Memory/State Issues**
- Clear state between unrelated conversations
- Implement state validation and cleanup
- Monitor memory usage in long-running agents
- Use appropriate data structures for state

## Next Steps

🚀 **Ready to build amazing agents?** Explore these advanced topics:

- **[Agent Tools](agent-tools.md)** - Master built-in and custom tools
- **[Agent Communication](agent-communication.md)** - Multi-agent coordination
- **[Agent Memory](agent-memory.md)** - Advanced state management
- **[Building Research Agents](building-research-agents.md)** - Specialized research patterns

### Related Examples

- **[Agent Simple LLM](../../cmd/examples/agent-simple-llm/)** - Basic agent patterns
- **[Agent Built-in Tools](../../cmd/examples/agent-llm-builtin-tools/)** - Tool integration examples
- **[Agent Multi-Coordination](../../cmd/examples/agent-multi-coordination/)** - Multi-agent systems
- **[Agent Custom Research](../../cmd/examples/agent-custom-research/)** - Custom agent implementation

---

**Need help?** Check our [troubleshooting guide](../advanced/troubleshooting.md) or join the discussion on [GitHub](https://github.com/lexlapax/go-llms/discussions).