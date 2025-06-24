# Intermediate Projects: Practical Applications

> **[Project Root](/) / [Documentation](../..) / [User Guide](../../user-guide) / [Examples](../../user-guide/examples) / Intermediate Projects**

Five comprehensive intermediate-level projects that demonstrate practical Go-LLMs applications. Each project builds on core concepts while introducing more sophisticated patterns and real-world integration scenarios.

---

## Project 1: Smart News Aggregator

Build an intelligent news aggregation system that collects, filters, and summarizes news from multiple sources.

### Features
- RSS feed monitoring and content extraction
- AI-powered topic categorization and duplicate detection
- Personalized content filtering based on user interests
- Automated summary generation with key insights
- Trend analysis and emerging topic detection

### Implementation

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/agent/workflow"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
)

// News aggregator system
type NewsAggregator struct {
    extractorAgent      *core.LLMAgent
    categorizerAgent    *core.LLMAgent  
    summarizerAgent     *core.LLMAgent
    trendAnalyzer       *core.LLMAgent
    workflowAgent       *workflow.SequentialAgent
    feeds               []string
    userPreferences     UserPreferences
}

type UserPreferences struct {
    Topics          []string `json:"topics"`
    Languages       []string `json:"languages"`
    MinRelevance    float64  `json:"min_relevance"`
    SummaryLength   string   `json:"summary_length"` // short, medium, long
}

type NewsArticle struct {
    ID          string    `json:"id"`
    Title       string    `json:"title"`
    Content     string    `json:"content"`
    Source      string    `json:"source"`
    Category    string    `json:"category"`
    PublishedAt time.Time `json:"published_at"`
    Relevance   float64   `json:"relevance"`
    Summary     string    `json:"summary"`
    KeyInsights []string  `json:"key_insights"`
}

func NewNewsAggregator() (*NewsAggregator, error) {
    // Initialize OpenAI provider
provider := provider.NewOpenAIProvider(
)
```

---

## Project 2: Document Intelligence System

Create a system that processes various document formats and extracts structured information using AI.

### Features
- Multi-format document processing (PDF, DOCX, TXT, HTML)
- Intelligent content extraction and structuring
- Entity recognition and relationship mapping
- Document classification and tagging
- Searchable knowledge base creation

### Implementation

```go
package main

import (
    "context"
    "fmt"
    "log"
    "path/filepath"
    "strings"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/agent/workflow"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
)

type DocumentIntelligenceSystem struct {
    parserAgent       *core.LLMAgent
    extractorAgent    *core.LLMAgent
    classifierAgent   *core.LLMAgent
    entityAgent       *core.LLMAgent
    workflowAgent     *workflow.ParallelAgent
    supportedFormats  []string
}

type ProcessedDocument struct {
    ID           string                 `json:"id"`
    FileName     string                 `json:"file_name"`
    Format       string                 `json:"format"`
    Content      string                 `json:"content"`
    Structure    DocumentStructure      `json:"structure"`
    Entities     []Entity              `json:"entities"`
    Classification DocumentClass        `json:"classification"`
    Metadata     map[string]interface{} `json:"metadata"`
    Summary      string                 `json:"summary"`
}

type DocumentStructure struct {
    Title       string    `json:"title"`
    Sections    []Section `json:"sections"`
    Tables      []Table   `json:"tables"`
    Images      []Image   `json:"images"`
    References  []string  `json:"references"`
}

type Section struct {
    Heading string `json:"heading"`
    Content string `json:"content"`
    Level   int    `json:"level"`
}

type Entity struct {
    Text       string  `json:"text"`
    Type       string  `json:"type"` // PERSON, ORGANIZATION, LOCATION, DATE, etc.
    Confidence float64 `json:"confidence"`
    Context    string  `json:"context"`
}

type DocumentClass struct {
    Category    string  `json:"category"`
    SubCategory string  `json:"sub_category"`
    Confidence  float64 `json:"confidence"`
    Tags        []string `json:"tags"`
}

func NewDocumentIntelligenceSystem() (*DocumentIntelligenceSystem, error) {
provider := provider.NewOpenAIProvider(
)
```

---

## Project 3: Smart Task Scheduler

Build an intelligent task scheduling system that uses AI to optimize task prioritization and resource allocation.

### Features
- Natural language task input and parsing
- Smart priority assignment based on context
- Deadline and dependency management
- Resource conflict detection and resolution
- Automated scheduling optimization

### Implementation

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "sort"
    "time"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
)

type SmartTaskScheduler struct {
    parserAgent      *core.LLMAgent
    priorityAgent    *core.LLMAgent
    schedulerAgent   *core.LLMAgent
    optimizerAgent   *core.LLMAgent
    tasks           []Task
    resources       []Resource
    constraints     []Constraint
}

type Task struct {
    ID              string          `json:"id"`
    Title           string          `json:"title"`
    Description     string          `json:"description"`
    Priority        int             `json:"priority"` // 1-10 scale
    EstimatedHours  float64         `json:"estimated_hours"`
    Deadline        *time.Time      `json:"deadline,omitempty"`
    Dependencies    []string        `json:"dependencies"`
    RequiredSkills  []string        `json:"required_skills"`
    AssignedTo      string          `json:"assigned_to,omitempty"`
    Status          TaskStatus      `json:"status"`
    Context         TaskContext     `json:"context"`
    ScheduledStart  *time.Time      `json:"scheduled_start,omitempty"`
    ScheduledEnd    *time.Time      `json:"scheduled_end,omitempty"`
}

type TaskStatus string

const (
    TaskStatusPending    TaskStatus = "pending"
    TaskStatusScheduled  TaskStatus = "scheduled"
    TaskStatusInProgress TaskStatus = "in_progress"
    TaskStatusCompleted  TaskStatus = "completed"
    TaskStatusBlocked    TaskStatus = "blocked"
)

type TaskContext struct {
    Project      string   `json:"project"`
    Category     string   `json:"category"`
    Tags         []string `json:"tags"`
    BusinessImpact string `json:"business_impact"`
}

type Resource struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Skills      []string  `json:"skills"`
    Availability []TimeSlot `json:"availability"`
    CurrentLoad int       `json:"current_load"` // 0-100 percentage
}

type TimeSlot struct {
    Start time.Time `json:"start"`
    End   time.Time `json:"end"`
}

type Constraint struct {
    Type        ConstraintType `json:"type"`
    Description string         `json:"description"`
    TaskIDs     []string       `json:"task_ids,omitempty"`
    ResourceIDs []string       `json:"resource_ids,omitempty"`
}

type ConstraintType string

const (
    ConstraintDependency   ConstraintType = "dependency"
    ConstraintResource     ConstraintType = "resource"
    ConstraintTime         ConstraintType = "time"
    ConstraintPriority     ConstraintType = "priority"
)

func NewSmartTaskScheduler() (*SmartTaskScheduler, error) {
provider := provider.NewOpenAIProvider(
)
```

---

## Project 4: API Gateway with Intelligence

Build an intelligent API gateway that provides smart routing, request transformation, and automated API documentation.

### Features
- Smart request routing based on content analysis
- Automatic request/response transformation
- Real-time API documentation generation
- Intelligent caching and rate limiting
- Security threat detection and mitigation

### Implementation

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "net/http/httputil"
    "net/url"
    "strings"
    "sync"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
)

type IntelligentAPIGateway struct {
    routingAgent        *core.LLMAgent
    transformAgent      *core.LLMAgent
    docAgent           *core.LLMAgent
    securityAgent      *core.LLMAgent
    routes             map[string]*BackendService
    cache              *IntelligentCache
    rateLimiter        *SmartRateLimiter
    middleware         []gin.HandlerFunc
    mu                 sync.RWMutex
}

type BackendService struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    BaseURL     string            `json:"base_url"`
    Patterns    []RoutePattern    `json:"patterns"`
    Transform   *TransformConfig  `json:"transform,omitempty"`
    Cache       *CacheConfig      `json:"cache,omitempty"`
    RateLimit   *RateLimitConfig  `json:"rate_limit,omitempty"`
    Auth        *AuthConfig       `json:"auth,omitempty"`
}

type RoutePattern struct {
    Method      string            `json:"method"`
    Path        string            `json:"path"`
    ContentType string            `json:"content_type,omitempty"`
    Headers     map[string]string `json:"headers,omitempty"`
    Priority    int               `json:"priority"`
}

type TransformConfig struct {
    RequestTransform  string `json:"request_transform,omitempty"`
    ResponseTransform string `json:"response_transform,omitempty"`
}

type APIRequest struct {
    Method      string              `json:"method"`
    Path        string              `json:"path"`
    Headers     map[string]string   `json:"headers"`
    Body        string              `json:"body,omitempty"`
    QueryParams map[string]string   `json:"query_params"`
    Timestamp   time.Time           `json:"timestamp"`
    ClientIP    string              `json:"client_ip"`
}

func NewIntelligentAPIGateway() (*IntelligentAPIGateway, error) {
provider := provider.NewOpenAIProvider(
    domain.NewBaseURLOption("http://localhost:8001"),
    domain.NewBaseURLOption("http://localhost:8002"),
)
```

---

## Project 5: Intelligent Code Assistant

Create a development assistant that provides intelligent code analysis, suggestions, and automated code generation.

### Features
- Real-time code analysis and suggestions
- Automated code documentation generation
- Intelligent refactoring recommendations
- Code quality assessment and improvement
- Pattern detection and best practice enforcement

### Implementation

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "go/ast"
    "go/parser"
    "go/token"
    "log"
    "path/filepath"
    "strings"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain" 
    "github.com/lexlapax/go-llms/pkg/llm/provider"
)

type IntelligentCodeAssistant struct {
    analyzerAgent     *core.LLMAgent
    suggesterAgent    *core.LLMAgent
    documentorAgent   *core.LLMAgent
    refactorAgent     *core.LLMAgent
    qualityAgent      *core.LLMAgent
    codebase         *CodebaseContext
    analysisHistory  []AnalysisResult
}

type CodebaseContext struct {
    RootPath      string              `json:"root_path"`
    Language      string              `json:"language"`
    Framework     string              `json:"framework,omitempty"`
    Dependencies  []string            `json:"dependencies"`
    Patterns      []CodePattern       `json:"patterns"`
    Conventions   CodingConventions   `json:"conventions"`
}

type CodingConventions struct {
    NamingConvention string   `json:"naming_convention"`
    IndentationStyle string   `json:"indentation_style"`
    MaxLineLength    int      `json:"max_line_length"`
    RequiredHeaders  []string `json:"required_headers"`
    ForbiddenPatterns []string `json:"forbidden_patterns"`
}

type CodePattern struct {
    Name        string `json:"name"`
    Pattern     string `json:"pattern"`
    Description string `json:"description"`
    Category    string `json:"category"`
}

type AnalysisResult struct {
    FilePath        string                `json:"file_path"`
    Timestamp       time.Time             `json:"timestamp"`
    Issues          []CodeIssue           `json:"issues"`
    Suggestions     []CodeSuggestion      `json:"suggestions"`
    QualityScore    float64               `json:"quality_score"`
    Complexity      ComplexityMetrics     `json:"complexity"`
    Documentation   DocumentationAnalysis `json:"documentation"`
}

type CodeIssue struct {
    Type        IssueType `json:"type"`
    Severity    Severity  `json:"severity"`
    Line        int       `json:"line"`
    Column      int       `json:"column"`
    Message     string    `json:"message"`
    Suggestion  string    `json:"suggestion,omitempty"`
    RuleName    string    `json:"rule_name"`
}

type IssueType string

const (
    IssueTypeBug          IssueType = "bug"
    IssueTypePerformance  IssueType = "performance"
    IssueTypeSecurity     IssueType = "security"
    IssueTypeStyle        IssueType = "style"
    IssueTypeMaintainability IssueType = "maintainability"
)

type Severity string

const (
    SeverityLow      Severity = "low"
    SeverityMedium   Severity = "medium"
    SeverityHigh     Severity = "high"
    SeverityCritical Severity = "critical"
)

type CodeSuggestion struct {
    Type          SuggestionType `json:"type"`
    Description   string         `json:"description"`
    OriginalCode  string         `json:"original_code"`
    SuggestedCode string         `json:"suggested_code"`
    Reasoning     string         `json:"reasoning"`
    Confidence    float64        `json:"confidence"`
}

type SuggestionType string

const (
    SuggestionRefactor      SuggestionType = "refactor"
    SuggestionOptimize      SuggestionType = "optimize"
    SuggestionDocument      SuggestionType = "document"
    SuggestionTest          SuggestionType = "test"
    SuggestionSecurity      SuggestionType = "security"
)

func NewIntelligentCodeAssistant(codebasePath string) (*IntelligentCodeAssistant, error) {
provider := provider.NewOpenAIProvider(
)
```

---

## Next Steps

Each of these intermediate projects demonstrates key Go-LLMs patterns:

1. **Multi-Agent Coordination** - Using multiple specialized agents working together
2. **Real-World Integration** - Connecting with databases, APIs, and external services  
3. **Intelligent Processing** - Leveraging AI for decision-making and optimization
4. **Production Patterns** - Error handling, monitoring, and scalability considerations
5. **Domain-Specific Applications** - Tailored solutions for specific business needs

Choose a project that aligns with your interests and build upon these foundations to create more sophisticated applications.

> **Next:** [Advanced Projects](advanced-projects.md) - Complex multi-agent systems and enterprise applications