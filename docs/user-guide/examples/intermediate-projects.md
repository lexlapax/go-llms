# Intermediate Projects: Practical Applications

> **[Project Root](/) / [Documentation](/docs/) / [User Guide](/docs/user-guide/) / [Examples](/docs/user-guide/examples/) / Intermediate Projects**

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
    openaiProvider, err := provider.NewOpenAI(provider.OpenAIOptions{
        APIKey: "your-openai-key",
        Model:  "gpt-4",
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create OpenAI provider: %w", err)
    }

    // Create extractor agent
    extractorAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "content-extractor",
        SystemPrompt: extractorPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create extractor agent: %w", err)
    }

    // Create categorizer agent
    categorizerAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "content-categorizer",
        SystemPrompt: categorizerPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create categorizer agent: %w", err)
    }

    // Create summarizer agent
    summarizerAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "content-summarizer", 
        SystemPrompt: summarizerPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create summarizer agent: %w", err)
    }

    // Create trend analyzer
    trendAnalyzer, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "trend-analyzer",
        SystemPrompt: trendAnalyzerPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create trend analyzer: %w", err)
    }

    // Create workflow
    workflowAgent := workflow.NewSequentialAgent(workflow.SequentialAgentOptions{
        Name: "news-processing-workflow",
        Steps: []domain.Agent{
            extractorAgent,
            categorizerAgent,
            summarizerAgent,
        },
    })

    return &NewsAggregator{
        extractorAgent:   extractorAgent,
        categorizerAgent: categorizerAgent,
        summarizerAgent:  summarizerAgent,
        trendAnalyzer:    trendAnalyzer,
        workflowAgent:    workflowAgent,
        feeds: []string{
            "https://feeds.bbci.co.uk/news/world/rss.xml",
            "https://rss.cnn.com/rss/edition.rss",
            "https://feeds.reuters.com/reuters/topNews",
        },
        userPreferences: UserPreferences{
            Topics:        []string{"technology", "science", "business"},
            Languages:     []string{"en"},
            MinRelevance:  0.7,
            SummaryLength: "medium",
        },
    }, nil
}

func (na *NewsAggregator) ProcessFeed(ctx context.Context, feedURL string) ([]NewsArticle, error) {
    // Step 1: Extract articles from feed
    extractionCtx := domain.NewAgentContext().
        SetInputProperty("feed_url", feedURL).
        SetInputProperty("user_preferences", na.userPreferences)

    result, err := na.workflowAgent.Execute(ctx, extractionCtx)
    if err != nil {
        return nil, fmt.Errorf("failed to process feed: %w", err)
    }

    // Parse processed articles
    var articles []NewsArticle
    if articlesData := result.GetOutputProperty("processed_articles"); articlesData != nil {
        if err := json.Unmarshal([]byte(fmt.Sprintf("%v", articlesData)), &articles); err != nil {
            return nil, fmt.Errorf("failed to parse articles: %w", err)
        }
    }

    return articles, nil
}

func (na *NewsAggregator) AnalyzeTrends(ctx context.Context, articles []NewsArticle) (TrendAnalysis, error) {
    articlesJSON, _ := json.Marshal(articles)
    
    trendCtx := domain.NewAgentContext().
        SetInputProperty("articles", string(articlesJSON)).
        SetInputProperty("analysis_period", "24h")

    result, err := na.trendAnalyzer.Execute(ctx, trendCtx)
    if err != nil {
        return TrendAnalysis{}, fmt.Errorf("failed to analyze trends: %w", err)
    }

    var trends TrendAnalysis
    if trendsData := result.GetOutputProperty("trend_analysis"); trendsData != nil {
        if err := json.Unmarshal([]byte(fmt.Sprintf("%v", trendsData)), &trends); err != nil {
            return TrendAnalysis{}, fmt.Errorf("failed to parse trends: %w", err)
        }
    }

    return trends, nil
}

type TrendAnalysis struct {
    EmergingTopics  []string    `json:"emerging_topics"`
    TrendingKeywords []string   `json:"trending_keywords"`
    SentimentAnalysis map[string]float64 `json:"sentiment_analysis"`
    GeographicTrends map[string][]string `json:"geographic_trends"`
}

// System prompts
const extractorPrompt = `You are a news content extractor. Your task is to:
1. Parse RSS feeds and extract article content
2. Clean and normalize text content
3. Extract metadata (title, source, publication date)
4. Filter out advertisements and irrelevant content
5. Return structured article data

Always return valid JSON with the following structure:
{
  "articles": [{
    "title": "string",
    "content": "string", 
    "source": "string",
    "published_at": "ISO8601 timestamp",
    "url": "string"
  }]
}`

const categorizerPrompt = `You are a news categorizer. Your task is to:
1. Analyze article content for topic classification
2. Assign primary and secondary categories
3. Calculate relevance scores based on user preferences
4. Detect duplicate or similar articles
5. Filter content based on quality and relevance

Categories: technology, science, business, politics, sports, entertainment, health, world, local

Return JSON with categorized articles including relevance scores.`

const summarizerPrompt = `You are a news summarizer. Your task is to:
1. Create concise, informative summaries
2. Extract key insights and important details
3. Maintain factual accuracy
4. Adapt summary length to user preferences
5. Highlight breaking news or urgent information

Always preserve the core facts while making content accessible and engaging.`

const trendAnalyzerPrompt = `You are a trend analyst. Your task is to:
1. Identify emerging topics and patterns
2. Analyze sentiment across different topics
3. Detect geographic trends and regional focus
4. Track keyword frequency and importance
5. Provide insights on story development

Focus on actionable insights and notable patterns in the news landscape.`

func main() {
    ctx := context.Background()
    
    // Initialize news aggregator
    aggregator, err := NewNewsAggregator()
    if err != nil {
        log.Fatalf("Failed to initialize news aggregator: %v", err)
    }

    // Process feeds
    var allArticles []NewsArticle
    for _, feed := range aggregator.feeds {
        fmt.Printf("Processing feed: %s\n", feed)
        articles, err := aggregator.ProcessFeed(ctx, feed)
        if err != nil {
            log.Printf("Error processing feed %s: %v", feed, err)
            continue
        }
        allArticles = append(allArticles, articles...)
        fmt.Printf("Extracted %d articles\n", len(articles))
    }

    // Analyze trends
    fmt.Println("\nAnalyzing trends...")
    trends, err := aggregator.AnalyzeTrends(ctx, allArticles)
    if err != nil {
        log.Printf("Error analyzing trends: %v", err)
    } else {
        fmt.Printf("Emerging topics: %v\n", trends.EmergingTopics)
        fmt.Printf("Trending keywords: %v\n", trends.TrendingKeywords)
    }

    fmt.Printf("\nProcessed %d articles total\n", len(allArticles))
}
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
    openaiProvider, err := provider.NewOpenAI(provider.OpenAIOptions{
        APIKey: "your-openai-key",
        Model:  "gpt-4",
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create OpenAI provider: %w", err)
    }

    // Document parser agent
    parserAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "document-parser",
        SystemPrompt: documentParserPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create parser agent: %w", err)
    }

    // Content extractor agent
    extractorAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "content-extractor",
        SystemPrompt: contentExtractorPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create extractor agent: %w", err)
    }

    // Document classifier agent
    classifierAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "document-classifier",
        SystemPrompt: documentClassifierPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create classifier agent: %w", err)
    }

    // Entity recognition agent
    entityAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "entity-recognizer",
        SystemPrompt: entityRecognitionPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create entity agent: %w", err)
    }

    // Create parallel workflow for simultaneous processing
    workflowAgent := workflow.NewParallelAgent(workflow.ParallelAgentOptions{
        Name: "document-processing-workflow",
        Agents: []domain.Agent{
            extractorAgent,
            classifierAgent,
            entityAgent,
        },
        MergeStrategy: workflow.MergeAll,
    })

    return &DocumentIntelligenceSystem{
        parserAgent:      parserAgent,
        extractorAgent:   extractorAgent,
        classifierAgent:  classifierAgent,
        entityAgent:      entityAgent,
        workflowAgent:    workflowAgent,
        supportedFormats: []string{".pdf", ".docx", ".txt", ".html", ".md"},
    }, nil
}

func (dis *DocumentIntelligenceSystem) ProcessDocument(ctx context.Context, filePath string) (*ProcessedDocument, error) {
    // Validate file format
    ext := strings.ToLower(filepath.Ext(filePath))
    if !dis.isFormatSupported(ext) {
        return nil, fmt.Errorf("unsupported format: %s", ext)
    }

    // Step 1: Parse document content
    parseCtx := domain.NewAgentContext().
        SetInputProperty("file_path", filePath).
        SetInputProperty("format", ext)

    parseResult, err := dis.parserAgent.Execute(ctx, parseCtx)
    if err != nil {
        return nil, fmt.Errorf("failed to parse document: %w", err)
    }

    rawContent := parseResult.GetOutputProperty("content").(string)

    // Step 2: Parallel processing for extraction, classification, and entity recognition
    processCtx := domain.NewAgentContext().
        SetInputProperty("content", rawContent).
        SetInputProperty("file_name", filepath.Base(filePath))

    processResult, err := dis.workflowAgent.Execute(ctx, processCtx)
    if err != nil {
        return nil, fmt.Errorf("failed to process document: %w", err)
    }

    // Combine results
    doc := &ProcessedDocument{
        ID:       generateDocumentID(filePath),
        FileName: filepath.Base(filePath),
        Format:   ext,
        Content:  rawContent,
    }

    // Extract structured data from results
    if structure := processResult.GetOutputProperty("document_structure"); structure != nil {
        // Parse structure...
    }
    if entities := processResult.GetOutputProperty("entities"); entities != nil {
        // Parse entities...
    }
    if classification := processResult.GetOutputProperty("classification"); classification != nil {
        // Parse classification...
    }

    return doc, nil
}

func (dis *DocumentIntelligenceSystem) isFormatSupported(format string) bool {
    for _, supported := range dis.supportedFormats {
        if supported == format {
            return true
        }
    }
    return false
}

func generateDocumentID(filePath string) string {
    // Generate unique ID based on file path and timestamp
    return fmt.Sprintf("doc_%s_%d", filepath.Base(filePath), time.Now().Unix())
}

// System prompts
const documentParserPrompt = `You are a document parser. Your task is to:
1. Extract raw text content from various document formats
2. Preserve document structure and formatting information
3. Handle special elements (tables, images, headers)
4. Clean and normalize extracted text
5. Return structured content data

Support formats: PDF, DOCX, TXT, HTML, Markdown
Always return clean, structured text while preserving important formatting.`

const contentExtractorPrompt = `You are a content structure extractor. Your task is to:
1. Identify document sections, headings, and hierarchy
2. Extract tables, lists, and structured data
3. Locate and catalog images, charts, and figures
4. Identify references, citations, and links
5. Create a comprehensive document structure map

Return JSON with detailed structural information.`

const documentClassifierPrompt = `You are a document classifier. Your task is to:
1. Classify documents into categories (report, contract, manual, etc.)
2. Identify document purpose and domain
3. Assign relevant tags and keywords
4. Assess document formality and target audience
5. Provide confidence scores for classifications

Categories: technical, legal, business, academic, marketing, operational, financial`

const entityRecognitionPrompt = `You are an entity recognition specialist. Your task is to:
1. Identify named entities (people, places, organizations)
2. Extract dates, numbers, and measurements
3. Find technical terms and domain-specific concepts
4. Locate contact information and identifiers
5. Provide context and relationships between entities

Return structured entity data with types, confidence, and contextual information.`

func main() {
    ctx := context.Background()
    
    // Initialize document intelligence system
    system, err := NewDocumentIntelligenceSystem()
    if err != nil {
        log.Fatalf("Failed to initialize system: %v", err)
    }

    // Example document processing
    testDocuments := []string{
        "sample_report.pdf",
        "contract.docx",
        "manual.txt",
        "research_paper.pdf",
    }

    for _, docPath := range testDocuments {
        fmt.Printf("Processing document: %s\n", docPath)
        
        doc, err := system.ProcessDocument(ctx, docPath)
        if err != nil {
            log.Printf("Error processing %s: %v", docPath, err)
            continue
        }

        fmt.Printf("Successfully processed: %s\n", doc.FileName)
        fmt.Printf("Format: %s, Entities: %d\n", doc.Format, len(doc.Entities))
        fmt.Printf("Classification: %s (%.2f confidence)\n", 
            doc.Classification.Category, doc.Classification.Confidence)
        fmt.Println("---")
    }
}
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
    openaiProvider, err := provider.NewOpenAI(provider.OpenAIOptions{
        APIKey: "your-openai-key",
        Model:  "gpt-4",
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create OpenAI provider: %w", err)
    }

    // Task parser agent - converts natural language to structured tasks
    parserAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "task-parser",
        SystemPrompt: taskParserPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create parser agent: %w", err)
    }

    // Priority assessment agent
    priorityAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "priority-assessor",
        SystemPrompt: priorityAssessmentPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create priority agent: %w", err)
    }

    // Task scheduler agent
    schedulerAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "task-scheduler",
        SystemPrompt: taskSchedulingPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create scheduler agent: %w", err)
    }

    // Schedule optimizer agent
    optimizerAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "schedule-optimizer",
        SystemPrompt: scheduleOptimizationPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create optimizer agent: %w", err)
    }

    return &SmartTaskScheduler{
        parserAgent:    parserAgent,
        priorityAgent:  priorityAgent,
        schedulerAgent: schedulerAgent,
        optimizerAgent: optimizerAgent,
        tasks:         make([]Task, 0),
        resources:     make([]Resource, 0),
        constraints:   make([]Constraint, 0),
    }, nil
}

func (sts *SmartTaskScheduler) AddTaskFromNaturalLanguage(ctx context.Context, input string) (*Task, error) {
    // Parse natural language input into structured task
    parseCtx := domain.NewAgentContext().
        SetInputProperty("natural_language_input", input).
        SetInputProperty("existing_tasks", sts.getTaskSummaries())

    parseResult, err := sts.parserAgent.Execute(ctx, parseCtx)
    if err != nil {
        return nil, fmt.Errorf("failed to parse task: %w", err)
    }

    // Convert result to Task struct
    var task Task
    if taskData := parseResult.GetOutputProperty("parsed_task"); taskData != nil {
        taskJSON, _ := json.Marshal(taskData)
        if err := json.Unmarshal(taskJSON, &task); err != nil {
            return nil, fmt.Errorf("failed to unmarshal task: %w", err)
        }
    }

    // Assess priority using AI
    priorityCtx := domain.NewAgentContext().
        SetInputProperty("task", task).
        SetInputProperty("context", sts.getSchedulingContext())

    priorityResult, err := sts.priorityAgent.Execute(ctx, priorityCtx)
    if err != nil {
        return nil, fmt.Errorf("failed to assess priority: %w", err)
    }

    if priority := priorityResult.GetOutputProperty("priority"); priority != nil {
        task.Priority = int(priority.(float64))
    }

    task.ID = sts.generateTaskID()
    task.Status = TaskStatusPending
    sts.tasks = append(sts.tasks, task)

    return &task, nil
}

func (sts *SmartTaskScheduler) OptimizeSchedule(ctx context.Context) (*ScheduleOptimization, error) {
    // Prepare scheduling context
    scheduleCtx := domain.NewAgentContext().
        SetInputProperty("tasks", sts.tasks).
        SetInputProperty("resources", sts.resources).
        SetInputProperty("constraints", sts.constraints).
        SetInputProperty("current_time", time.Now())

    // Generate initial schedule
    scheduleResult, err := sts.schedulerAgent.Execute(ctx, scheduleCtx)
    if err != nil {
        return nil, fmt.Errorf("failed to generate schedule: %w", err)
    }

    // Optimize the generated schedule
    optimizeCtx := domain.NewAgentContext().
        SetInputProperty("initial_schedule", scheduleResult.GetOutputProperty("schedule")).
        SetInputProperty("optimization_goals", sts.getOptimizationGoals())

    optimizeResult, err := sts.optimizerAgent.Execute(ctx, optimizeCtx)
    if err != nil {
        return nil, fmt.Errorf("failed to optimize schedule: %w", err)
    }

    // Parse optimization results
    var optimization ScheduleOptimization
    if optData := optimizeResult.GetOutputProperty("optimization"); optData != nil {
        optJSON, _ := json.Marshal(optData)
        json.Unmarshal(optJSON, &optimization)
    }

    // Apply optimized schedule to tasks
    sts.applyScheduleOptimization(&optimization)

    return &optimization, nil
}

type ScheduleOptimization struct {
    OptimizedTasks    []Task                `json:"optimized_tasks"`
    ResourceUtilization map[string]float64  `json:"resource_utilization"`
    ConflictResolutions []ConflictResolution `json:"conflict_resolutions"`
    Metrics           OptimizationMetrics   `json:"metrics"`
    Recommendations   []string              `json:"recommendations"`
}

type ConflictResolution struct {
    ConflictType  string `json:"conflict_type"`
    Description   string `json:"description"`
    Resolution    string `json:"resolution"`
    AffectedTasks []string `json:"affected_tasks"`
}

type OptimizationMetrics struct {
    TotalTasks           int     `json:"total_tasks"`
    ScheduledTasks       int     `json:"scheduled_tasks"`
    AverageUtilization   float64 `json:"average_utilization"`
    DeadlineMisses       int     `json:"deadline_misses"`
    OverallocatedResources int   `json:"overallocated_resources"`
}

func (sts *SmartTaskScheduler) getTaskSummaries() []string {
    summaries := make([]string, len(sts.tasks))
    for i, task := range sts.tasks {
        summaries[i] = fmt.Sprintf("%s: %s (Priority: %d)", task.ID, task.Title, task.Priority)
    }
    return summaries
}

func (sts *SmartTaskScheduler) getSchedulingContext() map[string]interface{} {
    return map[string]interface{}{
        "total_tasks":      len(sts.tasks),
        "available_resources": len(sts.resources),
        "active_projects":  sts.getActiveProjects(),
        "current_workload": sts.getCurrentWorkload(),
    }
}

func (sts *SmartTaskScheduler) getOptimizationGoals() []string {
    return []string{
        "minimize_deadline_misses",
        "balance_resource_utilization",
        "prioritize_high_impact_tasks",
        "reduce_context_switching",
        "maintain_team_morale",
    }
}

func (sts *SmartTaskScheduler) generateTaskID() string {
    return fmt.Sprintf("task_%d", time.Now().UnixNano())
}

// System prompts
const taskParserPrompt = `You are a task parsing specialist. Your role is to:
1. Convert natural language task descriptions into structured data
2. Extract key information: title, description, requirements, deadlines
3. Identify task dependencies and relationships
4. Estimate effort and complexity
5. Categorize tasks by type and project

Parse input into structured JSON with all relevant task properties.
Be thorough in extracting implicit information and context.`

const priorityAssessmentPrompt = `You are a priority assessment expert. Your role is to:
1. Analyze task impact on business objectives
2. Consider urgency, importance, and dependencies
3. Evaluate resource requirements and availability
4. Factor in stakeholder expectations and deadlines
5. Assign priority scores on a 1-10 scale

Consider both immediate needs and long-term strategic value.
Provide reasoning for priority assignments.`

const taskSchedulingPrompt = `You are a task scheduling coordinator. Your role is to:
1. Create optimal task schedules considering all constraints
2. Balance resource allocation and workload distribution
3. Respect dependencies and deadline requirements
4. Minimize conflicts and maximize efficiency
5. Provide realistic time estimates

Generate comprehensive schedules with detailed resource assignments.
Flag potential conflicts and suggest resolutions.`

const scheduleOptimizationPrompt = `You are a schedule optimization specialist. Your role is to:
1. Analyze existing schedules for improvement opportunities
2. Optimize resource utilization and minimize waste
3. Reduce context switching and improve focus time
4. Balance competing priorities and constraints
5. Provide actionable optimization recommendations

Focus on practical improvements that enhance productivity and meet deadlines.
Consider team dynamics and individual capabilities.`

func main() {
    ctx := context.Background()
    
    // Initialize task scheduler
    scheduler, err := NewSmartTaskScheduler()
    if err != nil {
        log.Fatalf("Failed to initialize scheduler: %v", err)
    }

    // Add sample resources
    scheduler.resources = []Resource{
        {
            ID:     "dev1",
            Name:   "Alice Developer",
            Skills: []string{"go", "python", "frontend"},
            CurrentLoad: 70,
        },
        {
            ID:     "dev2", 
            Name:   "Bob Designer",
            Skills: []string{"ui/ux", "frontend", "graphics"},
            CurrentLoad: 60,
        },
    }

    // Example task inputs
    taskInputs := []string{
        "Implement user authentication system for the mobile app by Friday",
        "Review and update the API documentation, not urgent but should be done this week",
        "Fix the critical bug in the payment processing system ASAP",
        "Design mockups for the new dashboard feature for next sprint",
        "Set up CI/CD pipeline for the new microservice project",
    }

    // Add tasks from natural language
    for _, input := range taskInputs {
        fmt.Printf("Processing: %s\n", input)
        task, err := scheduler.AddTaskFromNaturalLanguage(ctx, input)
        if err != nil {
            log.Printf("Error processing task: %v", err)
            continue
        }
        fmt.Printf("Created task: %s (Priority: %d)\n", task.Title, task.Priority)
    }

    // Optimize schedule
    fmt.Println("\nOptimizing schedule...")
    optimization, err := scheduler.OptimizeSchedule(ctx)
    if err != nil {
        log.Printf("Error optimizing schedule: %v", err)
    } else {
        fmt.Printf("Optimization complete: %d tasks scheduled\n", optimization.Metrics.ScheduledTasks)
        fmt.Printf("Average utilization: %.1f%%\n", optimization.Metrics.AverageUtilization*100)
        
        if len(optimization.Recommendations) > 0 {
            fmt.Println("Recommendations:")
            for _, rec := range optimization.Recommendations {
                fmt.Printf("- %s\n", rec)
            }
        }
    }
}
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
    openaiProvider, err := provider.NewOpenAI(provider.OpenAIOptions{
        APIKey: "your-openai-key",
        Model:  "gpt-4",
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create OpenAI provider: %w", err)
    }

    // Routing intelligence agent
    routingAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "smart-router",
        SystemPrompt: smartRoutingPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create routing agent: %w", err)
    }

    // Request/Response transformation agent
    transformAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "transform-engine",
        SystemPrompt: transformationPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create transform agent: %w", err)
    }

    // Documentation generation agent
    docAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "doc-generator",
        SystemPrompt: documentationPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create doc agent: %w", err)
    }

    // Security analysis agent
    securityAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "security-analyzer",
        SystemPrompt: securityAnalysisPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create security agent: %w", err)
    }

    return &IntelligentAPIGateway{
        routingAgent:   routingAgent,
        transformAgent: transformAgent,
        docAgent:       docAgent,
        securityAgent:  securityAgent,
        routes:        make(map[string]*BackendService),
        cache:         NewIntelligentCache(),
        rateLimiter:   NewSmartRateLimiter(),
    }, nil
}

func (gateway *IntelligentAPIGateway) AddBackendService(service *BackendService) {
    gateway.mu.Lock()
    defer gateway.mu.Unlock()
    gateway.routes[service.ID] = service
}

func (gateway *IntelligentAPIGateway) SmartRoute(ctx context.Context, request *APIRequest) (*BackendService, error) {
    // Use AI to determine optimal routing
    routingCtx := domain.NewAgentContext().
        SetInputProperty("request", request).
        SetInputProperty("available_services", gateway.getServiceSummaries()).
        SetInputProperty("historical_patterns", gateway.getRoutingHistory())

    result, err := gateway.routingAgent.Execute(ctx, routingCtx)
    if err != nil {
        return nil, fmt.Errorf("failed to determine route: %w", err)
    }

    serviceID := result.GetOutputProperty("selected_service").(string)
    
    gateway.mu.RLock()
    service, exists := gateway.routes[serviceID]
    gateway.mu.RUnlock()

    if !exists {
        return nil, fmt.Errorf("service not found: %s", serviceID)
    }

    return service, nil
}

func (gateway *IntelligentAPIGateway) TransformRequest(ctx context.Context, request *APIRequest, config *TransformConfig) (*APIRequest, error) {
    if config == nil || config.RequestTransform == "" {
        return request, nil
    }

    transformCtx := domain.NewAgentContext().
        SetInputProperty("original_request", request).
        SetInputProperty("transform_config", config.RequestTransform)

    result, err := gateway.transformAgent.Execute(ctx, transformCtx)
    if err != nil {
        return nil, fmt.Errorf("failed to transform request: %w", err)
    }

    var transformedRequest APIRequest
    if reqData := result.GetOutputProperty("transformed_request"); reqData != nil {
        reqJSON, _ := json.Marshal(reqData)
        json.Unmarshal(reqJSON, &transformedRequest)
    }

    return &transformedRequest, nil
}

func (gateway *IntelligentAPIGateway) AnalyzeSecurity(ctx context.Context, request *APIRequest) (*SecurityAnalysis, error) {
    securityCtx := domain.NewAgentContext().
        SetInputProperty("request", request).
        SetInputProperty("threat_patterns", gateway.getThreatPatterns()).
        SetInputProperty("security_policies", gateway.getSecurityPolicies())

    result, err := gateway.securityAgent.Execute(ctx, securityCtx)
    if err != nil {
        return nil, fmt.Errorf("failed to analyze security: %w", err)
    }

    var analysis SecurityAnalysis
    if analysisData := result.GetOutputProperty("security_analysis"); analysisData != nil {
        analysisJSON, _ := json.Marshal(analysisData)
        json.Unmarshal(analysisJSON, &analysis)
    }

    return &analysis, nil
}

type SecurityAnalysis struct {
    ThreatLevel   string   `json:"threat_level"` // low, medium, high, critical
    Threats       []string `json:"threats"`
    Blocked       bool     `json:"blocked"`
    Reason        string   `json:"reason,omitempty"`
    Recommendations []string `json:"recommendations"`
}

func (gateway *IntelligentAPIGateway) SetupRoutes() *gin.Engine {
    r := gin.Default()

    // Add intelligent middleware
    r.Use(gateway.SecurityMiddleware())
    r.Use(gateway.SmartCacheMiddleware())
    r.Use(gateway.RateLimitMiddleware())

    // Catch-all route for intelligent routing
    r.Any("/*path", gateway.IntelligentHandler())

    // Admin endpoints
    r.GET("/admin/docs", gateway.GenerateDocumentation())
    r.GET("/admin/routes", gateway.ListRoutes())
    r.POST("/admin/services", gateway.AddServiceHandler())

    return r
}

func (gateway *IntelligentAPIGateway) IntelligentHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := c.Request.Context()

        // Parse request
        request := &APIRequest{
            Method:      c.Request.Method,
            Path:        c.Request.URL.Path,
            Headers:     make(map[string]string),
            QueryParams: make(map[string]string),
            Timestamp:   time.Now(),
            ClientIP:    c.ClientIP(),
        }

        // Copy headers
        for key, values := range c.Request.Header {
            if len(values) > 0 {
                request.Headers[key] = values[0]
            }
        }

        // Copy query parameters
        for key, values := range c.Request.URL.Query() {
            if len(values) > 0 {
                request.QueryParams[key] = values[0]
            }
        }

        // Security analysis
        security, err := gateway.AnalyzeSecurity(ctx, request)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Security analysis failed"})
            return
        }

        if security.Blocked {
            c.JSON(http.StatusForbidden, gin.H{
                "error": "Request blocked",
                "reason": security.Reason,
            })
            return
        }

        // Smart routing
        service, err := gateway.SmartRoute(ctx, request)
        if err != nil {
            c.JSON(http.StatusBadGateway, gin.H{"error": "Routing failed"})
            return
        }

        // Transform request if needed
        transformedRequest, err := gateway.TransformRequest(ctx, request, service.Transform)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Request transformation failed"})
            return
        }

        // Proxy to backend service
        gateway.proxyToBackend(c, service, transformedRequest)
    }
}

func (gateway *IntelligentAPIGateway) proxyToBackend(c *gin.Context, service *BackendService, request *APIRequest) {
    backendURL, _ := url.Parse(service.BaseURL)
    proxy := httputil.NewSingleHostReverseProxy(backendURL)
    
    // Modify request for backend
    director := proxy.Director
    proxy.Director = func(req *http.Request) {
        director(req)
        req.Host = backendURL.Host
        req.URL.Scheme = backendURL.Scheme
        req.URL.Host = backendURL.Host
        
        // Apply any header transformations
        for key, value := range request.Headers {
            req.Header.Set(key, value)
        }
    }

    proxy.ServeHTTP(c.Writer, c.Request)
}

// Helper methods
func (gateway *IntelligentAPIGateway) getServiceSummaries() []string {
    gateway.mu.RLock()
    defer gateway.mu.RUnlock()
    
    summaries := make([]string, 0, len(gateway.routes))
    for _, service := range gateway.routes {
        summaries = append(summaries, fmt.Sprintf("%s: %s", service.ID, service.Name))
    }
    return summaries
}

// System prompts
const smartRoutingPrompt = `You are an intelligent API gateway routing specialist. Your role is to:
1. Analyze incoming requests for optimal backend service selection
2. Consider request content, headers, and patterns
3. Factor in service health, load, and capabilities
4. Apply intelligent load balancing and failover logic
5. Learn from historical routing patterns

Select the best backend service based on request characteristics and service capabilities.
Provide reasoning for routing decisions and confidence scores.`

const transformationPrompt = `You are a request/response transformation engine. Your role is to:
1. Transform request formats between different API specifications
2. Modify headers, parameters, and body content as needed
3. Handle version compatibility and protocol translation
4. Ensure data integrity and type safety
5. Optimize requests for target backend systems

Apply transformations while preserving semantic meaning and functionality.
Handle edge cases and provide error recovery strategies.`

const documentationPrompt = `You are an API documentation generator. Your role is to:
1. Analyze API traffic patterns and generate comprehensive documentation
2. Create OpenAPI specifications from observed behavior
3. Document request/response schemas and examples
4. Generate usage guides and integration examples
5. Maintain up-to-date documentation automatically

Create clear, accurate, and comprehensive API documentation.
Include practical examples and common use cases.`

const securityAnalysisPrompt = `You are a security analysis specialist. Your role is to:
1. Analyze requests for potential security threats
2. Detect injection attacks, suspicious patterns, and anomalies
3. Apply security policies and access controls
4. Provide threat assessment and mitigation recommendations
5. Learn from attack patterns and adapt defenses

Focus on preventing security breaches while minimizing false positives.
Provide clear explanations for security decisions.`

func main() {
    // Initialize intelligent API gateway
    gateway, err := NewIntelligentAPIGateway()
    if err != nil {
        log.Fatalf("Failed to initialize gateway: %v", err)
    }

    // Add sample backend services
    gateway.AddBackendService(&BackendService{
        ID:      "user-service",
        Name:    "User Management Service",
        BaseURL: "http://localhost:8001",
        Patterns: []RoutePattern{
            {Method: "GET", Path: "/users/*", Priority: 1},
            {Method: "POST", Path: "/users", Priority: 1},
        },
    })

    gateway.AddBackendService(&BackendService{
        ID:      "order-service", 
        Name:    "Order Processing Service",
        BaseURL: "http://localhost:8002",
        Patterns: []RoutePattern{
            {Method: "*", Path: "/orders/*", Priority: 1},
        },
    })

    // Setup and start server
    r := gateway.SetupRoutes()
    fmt.Println("Intelligent API Gateway starting on :8080")
    log.Fatal(r.Run(":8080"))
}
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
    openaiProvider, err := provider.NewOpenAI(provider.OpenAIOptions{
        APIKey: "your-openai-key",
        Model:  "gpt-4",
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create OpenAI provider: %w", err)
    }

    // Code analyzer agent
    analyzerAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "code-analyzer",
        SystemPrompt: codeAnalysisPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create analyzer agent: %w", err)
    }

    // Suggestion generator agent
    suggesterAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "code-suggester",
        SystemPrompt: codeSuggestionPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create suggester agent: %w", err)
    }

    // Documentation generator agent
    documentorAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "code-documentor",
        SystemPrompt: codeDocumentationPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create documentor agent: %w", err)
    }

    // Refactoring agent
    refactorAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "code-refactor",
        SystemPrompt: codeRefactoringPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create refactor agent: %w", err)
    }

    // Quality assessment agent
    qualityAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "quality-assessor",
        SystemPrompt: qualityAssessmentPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create quality agent: %w", err)
    }

    // Initialize codebase context
    codebase, err := analyzeCodebaseContext(codebasePath)
    if err != nil {
        return nil, fmt.Errorf("failed to analyze codebase: %w", err)
    }

    return &IntelligentCodeAssistant{
        analyzerAgent:   analyzerAgent,
        suggesterAgent:  suggesterAgent,
        documentorAgent: documentorAgent,
        refactorAgent:   refactorAgent,
        qualityAgent:    qualityAgent,
        codebase:       codebase,
        analysisHistory: make([]AnalysisResult, 0),
    }, nil
}

func (assistant *IntelligentCodeAssistant) AnalyzeFile(ctx context.Context, filePath string) (*AnalysisResult, error) {
    // Read file content
    content, err := readFile(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to read file: %w", err)
    }

    // Parse code structure (for Go files)
    var astInfo *ASTInfo
    if strings.HasSuffix(filePath, ".go") {
        astInfo, err = parseGoFile(filePath, content)
        if err != nil {
            log.Printf("Failed to parse Go file: %v", err)
        }
    }

    // AI-powered code analysis
    analysisCtx := domain.NewAgentContext().
        SetInputProperty("file_path", filePath).
        SetInputProperty("file_content", content).
        SetInputProperty("codebase_context", assistant.codebase).
        SetInputProperty("ast_info", astInfo)

    analysisResult, err := assistant.analyzerAgent.Execute(ctx, analysisCtx)
    if err != nil {
        return nil, fmt.Errorf("failed to analyze code: %w", err)
    }

    // Generate suggestions
    suggestionCtx := domain.NewAgentContext().
        SetInputProperty("analysis_result", analysisResult.GetOutputProperty("analysis")).
        SetInputProperty("file_content", content).
        SetInputProperty("best_practices", assistant.getBestPractices())

    suggestionResult, err := assistant.suggesterAgent.Execute(ctx, suggestionCtx)
    if err != nil {
        return nil, fmt.Errorf("failed to generate suggestions: %w", err)
    }

    // Assess quality
    qualityCtx := domain.NewAgentContext().
        SetInputProperty("file_content", content).
        SetInputProperty("analysis", analysisResult.GetOutputProperty("analysis")).
        SetInputProperty("quality_criteria", assistant.getQualityCriteria())

    qualityResult, err := assistant.qualityAgent.Execute(ctx, qualityCtx)
    if err != nil {
        return nil, fmt.Errorf("failed to assess quality: %w", err)
    }

    // Combine results
    result := &AnalysisResult{
        FilePath:  filePath,
        Timestamp: time.Now(),
    }

    // Parse AI responses into structured data
    if issues := analysisResult.GetOutputProperty("issues"); issues != nil {
        issuesJSON, _ := json.Marshal(issues)
        json.Unmarshal(issuesJSON, &result.Issues)
    }

    if suggestions := suggestionResult.GetOutputProperty("suggestions"); suggestions != nil {
        suggestionsJSON, _ := json.Marshal(suggestions)
        json.Unmarshal(suggestionsJSON, &result.Suggestions)
    }

    if quality := qualityResult.GetOutputProperty("quality_score"); quality != nil {
        result.QualityScore = quality.(float64)
    }

    // Store analysis result
    assistant.analysisHistory = append(assistant.analysisHistory, *result)

    return result, nil
}

func (assistant *IntelligentCodeAssistant) GenerateDocumentation(ctx context.Context, filePath string) (*DocumentationResult, error) {
    content, err := readFile(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to read file: %w", err)
    }

    docCtx := domain.NewAgentContext().
        SetInputProperty("file_content", content).
        SetInputProperty("file_path", filePath).
        SetInputProperty("documentation_style", assistant.codebase.Conventions).
        SetInputProperty("existing_docs", assistant.getExistingDocumentation(filePath))

    result, err := assistant.documentorAgent.Execute(ctx, docCtx)
    if err != nil {
        return nil, fmt.Errorf("failed to generate documentation: %w", err)
    }

    var docResult DocumentationResult
    if docData := result.GetOutputProperty("documentation"); docData != nil {
        docJSON, _ := json.Marshal(docData)
        json.Unmarshal(docJSON, &docResult)
    }

    return &docResult, nil
}

func (assistant *IntelligentCodeAssistant) SuggestRefactoring(ctx context.Context, filePath string, targetPattern string) (*RefactoringResult, error) {
    content, err := readFile(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to read file: %w", err)
    }

    refactorCtx := domain.NewAgentContext().
        SetInputProperty("file_content", content).
        SetInputProperty("target_pattern", targetPattern).
        SetInputProperty("refactoring_rules", assistant.getRefactoringRules()).
        SetInputProperty("codebase_patterns", assistant.codebase.Patterns)

    result, err := assistant.refactorAgent.Execute(ctx, refactorCtx)
    if err != nil {
        return nil, fmt.Errorf("failed to suggest refactoring: %w", err)
    }

    var refactorResult RefactoringResult
    if refactorData := result.GetOutputProperty("refactoring"); refactorData != nil {
        refactorJSON, _ := json.Marshal(refactorData)
        json.Unmarshal(refactorJSON, &refactorResult)
    }

    return &refactorResult, nil
}

type DocumentationResult struct {
    GeneratedDocs   []GeneratedDoc `json:"generated_docs"`
    MissingDocs     []string       `json:"missing_docs"`
    Improvements    []string       `json:"improvements"`
    Coverage        float64        `json:"coverage"`
}

type GeneratedDoc struct {
    Type        string `json:"type"` // function, class, package, etc.
    Location    string `json:"location"`
    Content     string `json:"content"`
    Existing    string `json:"existing,omitempty"`
}

type RefactoringResult struct {
    Suggestions   []RefactoringSuggestion `json:"suggestions"`
    RiskLevel     string                 `json:"risk_level"`
    EstimatedTime string                 `json:"estimated_time"`
    Benefits      []string               `json:"benefits"`
}

type RefactoringSuggestion struct {
    Type          string  `json:"type"`
    Description   string  `json:"description"`
    OriginalCode  string  `json:"original_code"`
    RefactoredCode string `json:"refactored_code"`
    Confidence    float64 `json:"confidence"`
    Impact        string  `json:"impact"`
}

// Helper functions and system prompts
const codeAnalysisPrompt = `You are an expert code analyzer. Your role is to:
1. Identify bugs, security vulnerabilities, and performance issues
2. Detect code smells and maintainability problems
3. Check adherence to coding standards and best practices
4. Analyze code complexity and structure
5. Provide detailed issue reports with severity levels

Focus on actionable insights that improve code quality and maintainability.
Consider the broader codebase context and established patterns.`

const codeSuggestionPrompt = `You are a code improvement specialist. Your role is to:
1. Suggest specific code improvements and optimizations
2. Recommend refactoring opportunities
3. Propose better algorithms and data structures
4. Suggest missing error handling and edge cases
5. Recommend testing strategies and test cases

Provide practical, implementable suggestions with clear reasoning.
Consider performance, readability, and maintainability trade-offs.`

const codeDocumentationPrompt = `You are a documentation specialist. Your role is to:
1. Generate comprehensive code documentation
2. Create clear function and class descriptions
3. Document API endpoints and parameters
4. Generate usage examples and code samples
5. Identify missing or outdated documentation

Create documentation that helps developers understand and use the code effectively.
Follow established documentation conventions and standards.`

const codeRefactoringPrompt = `You are a refactoring expert. Your role is to:
1. Identify refactoring opportunities to improve code structure
2. Suggest design pattern applications
3. Recommend code consolidation and simplification
4. Propose better separation of concerns
5. Suggest performance optimizations

Focus on safe refactorings that preserve functionality while improving design.
Consider the impact on existing code and potential breaking changes.`

const qualityAssessmentPrompt = `You are a code quality assessor. Your role is to:
1. Evaluate overall code quality on multiple dimensions
2. Assess readability, maintainability, and testability
3. Calculate quality scores and metrics
4. Identify areas for improvement
5. Provide quality benchmarking and goals

Provide objective quality assessments with specific improvement recommendations.
Consider industry standards and best practices for the given language and domain.`

func main() {
    ctx := context.Background()
    
    // Initialize code assistant
    assistant, err := NewIntelligentCodeAssistant("./my-project")
    if err != nil {
        log.Fatalf("Failed to initialize code assistant: %v", err)
    }

    // Example file analysis
    testFiles := []string{
        "main.go",
        "handlers/user.go", 
        "models/user.go",
        "utils/helpers.go",
    }

    for _, file := range testFiles {
        fmt.Printf("Analyzing file: %s\n", file)
        
        result, err := assistant.AnalyzeFile(ctx, file)
        if err != nil {
            log.Printf("Error analyzing %s: %v", file, err)
            continue
        }

        fmt.Printf("Quality Score: %.2f\n", result.QualityScore)
        fmt.Printf("Issues Found: %d\n", len(result.Issues))
        fmt.Printf("Suggestions: %d\n", len(result.Suggestions))

        // Show critical issues
        for _, issue := range result.Issues {
            if issue.Severity == SeverityCritical || issue.Severity == SeverityHigh {
                fmt.Printf("  [%s] Line %d: %s\n", issue.Severity, issue.Line, issue.Message)
            }
        }

        // Generate documentation
        fmt.Printf("Generating documentation...\n")
        docs, err := assistant.GenerateDocumentation(ctx, file)
        if err != nil {
            log.Printf("Error generating docs for %s: %v", file, err)
        } else {
            fmt.Printf("Documentation coverage: %.1f%%\n", docs.Coverage*100)
        }

        fmt.Println("---")
    }

    fmt.Printf("Analysis complete. Total files analyzed: %d\n", len(assistant.analysisHistory))
}
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