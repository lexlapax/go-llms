# Data Pipelines: End-to-End Processing Workflows

> **[Project Root](/) / [Documentation](/docs/) / [User Guide](/docs/user-guide/) / [Guides](/docs/user-guide/guides/) / Data Pipelines**

Build sophisticated data processing pipelines that combine LLM capabilities with traditional data workflows. Master the art of creating scalable, reliable, and maintainable data transformation systems.

## Why Data Pipelines Matter

- **Automation** - Process large volumes of data without manual intervention
- **Consistency** - Apply the same transformations reliably across all data
- **Scalability** - Handle growing data volumes with parallel processing
- **Observability** - Monitor and debug complex data flows
- **Integration** - Connect LLMs with existing data infrastructure

## Pipeline Architecture

![Data Pipeline Architecture](../../images/data-pipeline-architecture.svg)

### Core Components
1. **Sources** - Data ingestion from files, APIs, databases
2. **Processors** - LLM-powered transformations and enrichment
3. **Validators** - Quality checks and error handling
4. **Routers** - Conditional logic and flow control
5. **Sinks** - Output to storage systems and APIs

### Pipeline Patterns
| Pattern | Use Case | Benefits |
|---------|----------|----------|
| **Linear** | Simple transformations | Easy to understand |
| **Branching** | Conditional processing | Flexible routing |
| **Parallel** | Independent operations | High throughput |
| **Streaming** | Real-time data | Low latency |
| **Batch** | Periodic processing | Resource efficient |

## Prerequisites

- [Structured Data completed](structured-data.md) ✅
- [Data Validation understanding](data-validation.md) ✅
- Basic knowledge of data processing concepts ✅

---

## Level 1: Basic Data Pipelines
*Build simple linear data processing workflows*

### Document Processing Pipeline
```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "os"
    "path/filepath"
    "sync"
    "time"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
    "github.com/lexlapax/go-llms/pkg/schema/validation"
    "github.com/lexlapax/go-llms/pkg/structured/processor"
)

// Pipeline represents a data processing pipeline
type Pipeline struct {
    name      string
    stages    []Stage
    config    PipelineConfig
    metrics   *PipelineMetrics
    errorChan chan PipelineError
}

// Stage represents a processing stage in the pipeline
type Stage interface {
    Name() string
    Process(ctx context.Context, data interface{}) (interface{}, error)
    Configure(config map[string]interface{}) error
}

// PipelineConfig holds pipeline configuration
type PipelineConfig struct {
    MaxConcurrency   int           `json:"max_concurrency"`
    BatchSize        int           `json:"batch_size"`
    ErrorStrategy    string        `json:"error_strategy"` // "stop", "skip", "retry"
    RetryAttempts    int           `json:"retry_attempts"`
    RetryDelay       time.Duration `json:"retry_delay"`
    Timeout          time.Duration `json:"timeout"`
    EnableMetrics    bool          `json:"enable_metrics"`
    EnableCheckpoint bool          `json:"enable_checkpoint"`
}

// PipelineMetrics tracks pipeline performance
type PipelineMetrics struct {
    mu               sync.RWMutex
    TotalProcessed   int64
    SuccessCount     int64
    ErrorCount       int64
    SkippedCount     int64
    TotalDuration    time.Duration
    StageMetrics     map[string]*StageMetrics
}

type StageMetrics struct {
    ProcessedCount int64
    ErrorCount     int64
    AvgDuration    time.Duration
    LastExecution  time.Time
}

// PipelineError represents an error in the pipeline
type PipelineError struct {
    Stage     string
    Data      interface{}
    Error     error
    Timestamp time.Time
    Retries   int
}

// Document represents a document to process
type Document struct {
    ID       string                 `json:"id"`
    Path     string                 `json:"path"`
    Content  string                 `json:"content"`
    Metadata map[string]interface{} `json:"metadata"`
}

// ProcessedDocument represents the output
type ProcessedDocument struct {
    Document
    ExtractedData map[string]interface{} `json:"extracted_data"`
    Summary       string                 `json:"summary"`
    Categories    []string               `json:"categories"`
    Entities      []Entity               `json:"entities"`
    ProcessedAt   time.Time             `json:"processed_at"`
}

type Entity struct {
    Type  string `json:"type"`
    Value string `json:"value"`
    Count int    `json:"count"`
}

// FileSourceStage reads documents from files
type FileSourceStage struct {
    directory string
    pattern   string
    processed map[string]bool
    mu        sync.Mutex
}

func NewFileSourceStage(directory, pattern string) *FileSourceStage {
    return &FileSourceStage{
        directory: directory,
        pattern:   pattern,
        processed: make(map[string]bool),
    }
}

func (fs *FileSourceStage) Name() string { return "file_source" }

func (fs *FileSourceStage) Configure(config map[string]interface{}) error {
    if dir, ok := config["directory"].(string); ok {
        fs.directory = dir
    }
    if pattern, ok := config["pattern"].(string); ok {
        fs.pattern = pattern
    }
    return nil
}

func (fs *FileSourceStage) Process(ctx context.Context, data interface{}) (interface{}, error) {
    // Find files matching pattern
    files, err := filepath.Glob(filepath.Join(fs.directory, fs.pattern))
    if err != nil {
        return nil, fmt.Errorf("failed to list files: %w", err)
    }

    documents := []Document{}
    
    for _, file := range files {
        fs.mu.Lock()
        if fs.processed[file] {
            fs.mu.Unlock()
            continue
        }
        fs.processed[file] = true
        fs.mu.Unlock()

        content, err := os.ReadFile(file)
        if err != nil {
            log.Printf("Failed to read file %s: %v", file, err)
            continue
        }

        doc := Document{
            ID:      filepath.Base(file),
            Path:    file,
            Content: string(content),
            Metadata: map[string]interface{}{
                "size":         len(content),
                "extension":    filepath.Ext(file),
                "modified":     getFileModTime(file),
            },
        }
        
        documents = append(documents, doc)
    }

    return documents, nil
}

// LLMExtractionStage extracts structured data using LLM
type LLMExtractionStage struct {
    agent     domain.BaseAgent
    schema    *schemaDomain.Schema
    processor *processor.StructuredProcessor
}

func NewLLMExtractionStage(agent domain.BaseAgent, schema *schemaDomain.Schema) *LLMExtractionStage {
    validator := validation.NewValidator()
    structProcessor := processor.NewStructuredProcessor(validator)
    
    return &LLMExtractionStage{
        agent:     agent,
        schema:    schema,
        processor: structProcessor,
    }
}

func (le *LLMExtractionStage) Name() string { return "llm_extraction" }

func (le *LLMExtractionStage) Configure(config map[string]interface{}) error {
    return nil
}

func (le *LLMExtractionStage) Process(ctx context.Context, data interface{}) (interface{}, error) {
    documents, ok := data.([]Document)
    if !ok {
        return nil, fmt.Errorf("expected []Document, got %T", data)
    }

    processed := []ProcessedDocument{}
    
    for _, doc := range documents {
        // Extract structured data
        extractPrompt := fmt.Sprintf(`Extract the following information from this document:
- Key topics and themes
- People mentioned (names)
- Organizations mentioned
- Locations mentioned
- Dates and time references
- Key facts and figures

Document:
%s

Return the extracted information as structured JSON.`, doc.Content)

        state := domain.NewState()
        state.Set("user_input", extractPrompt)

        result, err := le.agent.Run(ctx, state)
        if err != nil {
            log.Printf("Extraction failed for %s: %v", doc.ID, err)
            continue
        }

        response, _ := result.Get("response")
        
        // Parse extracted data
        var extractedData map[string]interface{}
        if err := json.Unmarshal([]byte(response.(string)), &extractedData); err != nil {
            log.Printf("Failed to parse extraction result: %v", err)
            extractedData = map[string]interface{}{"raw": response}
        }

        // Create processed document
        processed = append(processed, ProcessedDocument{
            Document:      doc,
            ExtractedData: extractedData,
            ProcessedAt:   time.Now(),
        })
    }

    return processed, nil
}

// SummarizationStage generates summaries
type SummarizationStage struct {
    agent      domain.BaseAgent
    maxLength  int
    style      string
}

func NewSummarizationStage(agent domain.BaseAgent) *SummarizationStage {
    return &SummarizationStage{
        agent:     agent,
        maxLength: 200,
        style:     "concise",
    }
}

func (ss *SummarizationStage) Name() string { return "summarization" }

func (ss *SummarizationStage) Configure(config map[string]interface{}) error {
    if length, ok := config["max_length"].(int); ok {
        ss.maxLength = length
    }
    if style, ok := config["style"].(string); ok {
        ss.style = style
    }
    return nil
}

func (ss *SummarizationStage) Process(ctx context.Context, data interface{}) (interface{}, error) {
    documents, ok := data.([]ProcessedDocument)
    if !ok {
        return nil, fmt.Errorf("expected []ProcessedDocument, got %T", data)
    }

    for i := range documents {
        summaryPrompt := fmt.Sprintf(`Summarize this document in a %s style, maximum %d words:

%s

Focus on the main points and key information.`, ss.style, ss.maxLength, documents[i].Content)

        state := domain.NewState()
        state.Set("user_input", summaryPrompt)

        result, err := ss.agent.Run(ctx, state)
        if err != nil {
            log.Printf("Summarization failed for %s: %v", documents[i].ID, err)
            continue
        }

        if response, exists := result.Get("response"); exists {
            documents[i].Summary = response.(string)
        }
    }

    return documents, nil
}

// CategorizationStage categorizes documents
type CategorizationStage struct {
    agent      domain.BaseAgent
    categories []string
}

func NewCategorizationStage(agent domain.BaseAgent, categories []string) *CategorizationStage {
    return &CategorizationStage{
        agent:      agent,
        categories: categories,
    }
}

func (cs *CategorizationStage) Name() string { return "categorization" }

func (cs *CategorizationStage) Configure(config map[string]interface{}) error {
    if cats, ok := config["categories"].([]string); ok {
        cs.categories = cats
    }
    return nil
}

func (cs *CategorizationStage) Process(ctx context.Context, data interface{}) (interface{}, error) {
    documents, ok := data.([]ProcessedDocument)
    if !ok {
        return nil, fmt.Errorf("expected []ProcessedDocument, got %T", data)
    }

    for i := range documents {
        categoryPrompt := fmt.Sprintf(`Categorize this document into one or more of these categories:
%v

Document Summary:
%s

Return only the applicable category names as a JSON array.`, cs.categories, documents[i].Summary)

        state := domain.NewState()
        state.Set("user_input", categoryPrompt)

        result, err := cs.agent.Run(ctx, state)
        if err != nil {
            log.Printf("Categorization failed for %s: %v", documents[i].ID, err)
            continue
        }

        if response, exists := result.Get("response"); exists {
            var categories []string
            if err := json.Unmarshal([]byte(response.(string)), &categories); err == nil {
                documents[i].Categories = categories
            }
        }
    }

    return documents, nil
}

// JSONOutputStage writes results to JSON files
type JSONOutputStage struct {
    outputDir string
}

func NewJSONOutputStage(outputDir string) *JSONOutputStage {
    os.MkdirAll(outputDir, 0755)
    return &JSONOutputStage{
        outputDir: outputDir,
    }
}

func (jo *JSONOutputStage) Name() string { return "json_output" }

func (jo *JSONOutputStage) Configure(config map[string]interface{}) error {
    if dir, ok := config["output_dir"].(string); ok {
        jo.outputDir = dir
        os.MkdirAll(jo.outputDir, 0755)
    }
    return nil
}

func (jo *JSONOutputStage) Process(ctx context.Context, data interface{}) (interface{}, error) {
    documents, ok := data.([]ProcessedDocument)
    if !ok {
        return nil, fmt.Errorf("expected []ProcessedDocument, got %T", data)
    }

    for _, doc := range documents {
        outputPath := filepath.Join(jo.outputDir, doc.ID+".json")
        
        jsonData, err := json.MarshalIndent(doc, "", "  ")
        if err != nil {
            log.Printf("Failed to marshal document %s: %v", doc.ID, err)
            continue
        }

        if err := os.WriteFile(outputPath, jsonData, 0644); err != nil {
            log.Printf("Failed to write output for %s: %v", doc.ID, err)
            continue
        }

        fmt.Printf("✅ Processed: %s -> %s\n", doc.ID, outputPath)
    }

    return documents, nil
}

// Pipeline implementation
func NewPipeline(name string, config PipelineConfig) *Pipeline {
    return &Pipeline{
        name:      name,
        stages:    []Stage{},
        config:    config,
        metrics:   NewPipelineMetrics(),
        errorChan: make(chan PipelineError, 100),
    }
}

func NewPipelineMetrics() *PipelineMetrics {
    return &PipelineMetrics{
        StageMetrics: make(map[string]*StageMetrics),
    }
}

func (p *Pipeline) AddStage(stage Stage) {
    p.stages = append(p.stages, stage)
    p.metrics.StageMetrics[stage.Name()] = &StageMetrics{}
}

func (p *Pipeline) Run(ctx context.Context) error {
    fmt.Printf("🚀 Starting pipeline: %s\n", p.name)
    startTime := time.Now()

    // Start error handler
    go p.handleErrors()

    // Execute pipeline
    var data interface{}
    var err error

    for i, stage := range p.stages {
        stageStart := time.Now()
        fmt.Printf("⚙️  Stage %d/%d: %s\n", i+1, len(p.stages), stage.Name())

        // Apply timeout if configured
        stageCtx := ctx
        if p.config.Timeout > 0 {
            var cancel context.CancelFunc
            stageCtx, cancel = context.WithTimeout(ctx, p.config.Timeout)
            defer cancel()
        }

        // Process with retry logic
        data, err = p.processWithRetry(stageCtx, stage, data)
        
        // Update metrics
        p.updateStageMetrics(stage.Name(), time.Since(stageStart), err)

        if err != nil {
            p.handleStageError(stage, data, err)
            
            switch p.config.ErrorStrategy {
            case "stop":
                return fmt.Errorf("pipeline stopped at stage %s: %w", stage.Name(), err)
            case "skip":
                fmt.Printf("⚠️  Skipping stage %s due to error: %v\n", stage.Name(), err)
                continue
            }
        }
    }

    p.metrics.mu.Lock()
    p.metrics.TotalDuration = time.Since(startTime)
    p.metrics.mu.Unlock()

    fmt.Printf("✅ Pipeline completed in %v\n", time.Since(startTime))
    p.printMetrics()

    return nil
}

func (p *Pipeline) processWithRetry(ctx context.Context, stage Stage, data interface{}) (interface{}, error) {
    var result interface{}
    var err error

    for attempt := 0; attempt <= p.config.RetryAttempts; attempt++ {
        if attempt > 0 {
            fmt.Printf("  Retry attempt %d/%d\n", attempt, p.config.RetryAttempts)
            time.Sleep(p.config.RetryDelay)
        }

        result, err = stage.Process(ctx, data)
        if err == nil {
            return result, nil
        }
    }

    return nil, err
}

func (p *Pipeline) handleStageError(stage Stage, data interface{}, err error) {
    pipelineErr := PipelineError{
        Stage:     stage.Name(),
        Data:      data,
        Error:     err,
        Timestamp: time.Now(),
    }

    select {
    case p.errorChan <- pipelineErr:
    default:
        log.Printf("Error channel full, dropping error: %v", err)
    }
}

func (p *Pipeline) handleErrors() {
    for err := range p.errorChan {
        log.Printf("Pipeline error in stage %s: %v", err.Stage, err.Error)
        
        p.metrics.mu.Lock()
        p.metrics.ErrorCount++
        p.metrics.mu.Unlock()
    }
}

func (p *Pipeline) updateStageMetrics(stageName string, duration time.Duration, err error) {
    p.metrics.mu.Lock()
    defer p.metrics.mu.Unlock()

    stageMetrics := p.metrics.StageMetrics[stageName]
    stageMetrics.ProcessedCount++
    stageMetrics.LastExecution = time.Now()

    if err != nil {
        stageMetrics.ErrorCount++
        p.metrics.ErrorCount++
    } else {
        p.metrics.SuccessCount++
    }

    // Update average duration
    if stageMetrics.AvgDuration == 0 {
        stageMetrics.AvgDuration = duration
    } else {
        stageMetrics.AvgDuration = (stageMetrics.AvgDuration + duration) / 2
    }

    p.metrics.TotalProcessed++
}

func (p *Pipeline) printMetrics() {
    p.metrics.mu.RLock()
    defer p.metrics.mu.RUnlock()

    fmt.Printf("\n📊 Pipeline Metrics:\n")
    fmt.Printf("Total Processed: %d\n", p.metrics.TotalProcessed)
    fmt.Printf("Success: %d | Errors: %d | Skipped: %d\n", 
        p.metrics.SuccessCount, p.metrics.ErrorCount, p.metrics.SkippedCount)
    fmt.Printf("Total Duration: %v\n", p.metrics.TotalDuration)

    fmt.Printf("\nStage Metrics:\n")
    for stageName, metrics := range p.metrics.StageMetrics {
        fmt.Printf("  %s:\n", stageName)
        fmt.Printf("    Processed: %d | Errors: %d\n", metrics.ProcessedCount, metrics.ErrorCount)
        fmt.Printf("    Avg Duration: %v\n", metrics.AvgDuration)
    }
}

// Helper functions
func getFileModTime(path string) time.Time {
    info, err := os.Stat(path)
    if err != nil {
        return time.Time{}
    }
    return info.ModTime()
}

func main() {
    fmt.Println("📄 Data Pipeline - Document Processing")
    fmt.Println("=====================================")

    // Create agents
    extractionAgent, err := core.NewAgentFromString("extractor", "openai/gpt-4o-mini")
    if err != nil {
        log.Fatalf("Failed to create extraction agent: %v", err)
    }

    summaryAgent, err := core.NewAgentFromString("summarizer", "anthropic/claude-3-5-haiku")
    if err != nil {
        log.Fatalf("Failed to create summary agent: %v", err)
    }

    // Create pipeline
    config := PipelineConfig{
        MaxConcurrency: 5,
        BatchSize:      10,
        ErrorStrategy:  "skip",
        RetryAttempts:  2,
        RetryDelay:     time.Second,
        Timeout:        30 * time.Second,
        EnableMetrics:  true,
    }

    pipeline := NewPipeline("document-processor", config)

    // Add stages
    pipeline.AddStage(NewFileSourceStage("./documents", "*.txt"))
    pipeline.AddStage(NewLLMExtractionStage(extractionAgent, nil))
    pipeline.AddStage(NewSummarizationStage(summaryAgent))
    pipeline.AddStage(NewCategorizationStage(summaryAgent, []string{
        "Technical", "Business", "Legal", "Medical", "Educational", "News", "Other",
    }))
    pipeline.AddStage(NewJSONOutputStage("./processed"))

    // Create sample documents
    createSampleDocuments()

    // Run pipeline
    ctx := context.Background()
    if err := pipeline.Run(ctx); err != nil {
        log.Fatalf("Pipeline failed: %v", err)
    }
}

func createSampleDocuments() {
    os.MkdirAll("./documents", 0755)
    
    samples := []struct {
        name    string
        content string
    }{
        {
            "technical-doc.txt",
            `Technical Specification: Cloud Storage System
            
            This document outlines the architecture for our new cloud storage system.
            The system will use distributed object storage with S3-compatible API.
            Key requirements include 99.99% availability, encryption at rest, and
            support for files up to 5TB. The system should handle 10,000 requests
            per second with sub-100ms latency.`,
        },
        {
            "business-report.txt",
            `Q4 2023 Business Report
            
            Revenue increased by 25% compared to Q3, reaching $5.2 million.
            Customer acquisition cost decreased to $120 per customer.
            The marketing team launched three successful campaigns resulting
            in 2,000 new sign-ups. Churn rate remains stable at 5%.`,
        },
        {
            "news-article.txt",
            `Breaking: Major Tech Company Announces AI Initiative
            
            Silicon Valley giant TechCorp announced today a $1 billion investment
            in artificial intelligence research. CEO Jane Smith stated that the
            company plans to hire 500 AI researchers over the next year.
            The initiative will focus on healthcare and education applications.`,
        },
    }

    for _, sample := range samples {
        path := filepath.Join("./documents", sample.name)
        os.WriteFile(path, []byte(sample.content), 0644)
    }
}
```

---

## Level 2: Advanced Pipeline Patterns
*Implement branching, parallel processing, and streaming*

### Streaming Data Pipeline
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
)

// StreamingPipeline processes data in real-time streams
type StreamingPipeline struct {
    name         string
    stages       []StreamStage
    config       StreamConfig
    inputChan    chan StreamData
    metrics      *StreamMetrics
    errorHandler ErrorHandler
}

// StreamStage processes streaming data
type StreamStage interface {
    Name() string
    ProcessStream(ctx context.Context, input <-chan StreamData) (<-chan StreamData, error)
    Close() error
}

// StreamData represents data flowing through the pipeline
type StreamData struct {
    ID        string                 `json:"id"`
    Type      string                 `json:"type"`
    Payload   interface{}           `json:"payload"`
    Metadata  map[string]interface{} `json:"metadata"`
    Timestamp time.Time             `json:"timestamp"`
}

// StreamConfig configures the streaming pipeline
type StreamConfig struct {
    BufferSize       int           `json:"buffer_size"`
    MaxConcurrency   int           `json:"max_concurrency"`
    FlushInterval    time.Duration `json:"flush_interval"`
    BackpressureMode string        `json:"backpressure_mode"` // "drop", "block", "buffer"
    Checkpointing    bool          `json:"checkpointing"`
}

// StreamMetrics tracks streaming performance
type StreamMetrics struct {
    mu                sync.RWMutex
    MessagesProcessed int64
    MessagesDropped   int64
    Throughput        float64
    Latency           time.Duration
    BackpressureCount int64
}

// BranchingStage routes data to different paths
type BranchingStage struct {
    name       string
    predicates map[string]Predicate
    branches   map[string]chan StreamData
}

type Predicate func(data StreamData) bool

func NewBranchingStage(name string) *BranchingStage {
    return &BranchingStage{
        name:       name,
        predicates: make(map[string]Predicate),
        branches:   make(map[string]chan StreamData),
    }
}

func (bs *BranchingStage) Name() string { return bs.name }

func (bs *BranchingStage) AddBranch(name string, predicate Predicate, bufferSize int) <-chan StreamData {
    branch := make(chan StreamData, bufferSize)
    bs.predicates[name] = predicate
    bs.branches[name] = branch
    return branch
}

func (bs *BranchingStage) ProcessStream(ctx context.Context, input <-chan StreamData) (<-chan StreamData, error) {
    output := make(chan StreamData, len(bs.branches)*100)

    go func() {
        defer close(output)
        defer bs.closeAllBranches()

        for {
            select {
            case data, ok := <-input:
                if !ok {
                    return
                }

                // Route to appropriate branches
                routed := false
                for branchName, predicate := range bs.predicates {
                    if predicate(data) {
                        select {
                        case bs.branches[branchName] <- data:
                            routed = true
                        case <-ctx.Done():
                            return
                        default:
                            log.Printf("Branch %s full, dropping message", branchName)
                        }
                    }
                }

                // Forward to output if routed
                if routed {
                    select {
                    case output <- data:
                    case <-ctx.Done():
                        return
                    }
                }

            case <-ctx.Done():
                return
            }
        }
    }()

    return output, nil
}

func (bs *BranchingStage) Close() error {
    bs.closeAllBranches()
    return nil
}

func (bs *BranchingStage) closeAllBranches() {
    for _, branch := range bs.branches {
        close(branch)
    }
}

// ParallelProcessingStage processes data in parallel
type ParallelProcessingStage struct {
    name      string
    workers   int
    processor func(context.Context, StreamData) (StreamData, error)
}

func NewParallelProcessingStage(name string, workers int, processor func(context.Context, StreamData) (StreamData, error)) *ParallelProcessingStage {
    return &ParallelProcessingStage{
        name:      name,
        workers:   workers,
        processor: processor,
    }
}

func (ps *ParallelProcessingStage) Name() string { return ps.name }

func (ps *ParallelProcessingStage) ProcessStream(ctx context.Context, input <-chan StreamData) (<-chan StreamData, error) {
    output := make(chan StreamData, ps.workers*2)
    
    var wg sync.WaitGroup
    
    // Start workers
    for i := 0; i < ps.workers; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            
            for {
                select {
                case data, ok := <-input:
                    if !ok {
                        return
                    }

                    // Process data
                    processed, err := ps.processor(ctx, data)
                    if err != nil {
                        log.Printf("Worker %d processing error: %v", workerID, err)
                        continue
                    }

                    select {
                    case output <- processed:
                    case <-ctx.Done():
                        return
                    }

                case <-ctx.Done():
                    return
                }
            }
        }(i)
    }

    // Close output when all workers done
    go func() {
        wg.Wait()
        close(output)
    }()

    return output, nil
}

func (ps *ParallelProcessingStage) Close() error {
    return nil
}

// WindowedAggregationStage aggregates data over time windows
type WindowedAggregationStage struct {
    name           string
    windowDuration time.Duration
    aggregator     func([]StreamData) StreamData
}

func NewWindowedAggregationStage(name string, duration time.Duration, aggregator func([]StreamData) StreamData) *WindowedAggregationStage {
    return &WindowedAggregationStage{
        name:           name,
        windowDuration: duration,
        aggregator:     aggregator,
    }
}

func (wa *WindowedAggregationStage) Name() string { return wa.name }

func (wa *WindowedAggregationStage) ProcessStream(ctx context.Context, input <-chan StreamData) (<-chan StreamData, error) {
    output := make(chan StreamData, 10)
    
    go func() {
        defer close(output)
        
        window := []StreamData{}
        ticker := time.NewTicker(wa.windowDuration)
        defer ticker.Stop()

        for {
            select {
            case data, ok := <-input:
                if !ok {
                    // Process remaining window
                    if len(window) > 0 {
                        aggregated := wa.aggregator(window)
                        select {
                        case output <- aggregated:
                        case <-ctx.Done():
                        }
                    }
                    return
                }
                
                window = append(window, data)

            case <-ticker.C:
                if len(window) > 0 {
                    aggregated := wa.aggregator(window)
                    select {
                    case output <- aggregated:
                    case <-ctx.Done():
                        return
                    }
                    window = []StreamData{}
                }

            case <-ctx.Done():
                return
            }
        }
    }()

    return output, nil
}

func (wa *WindowedAggregationStage) Close() error {
    return nil
}

// EnrichmentStage enriches data with additional information
type EnrichmentStage struct {
    name     string
    agent    domain.BaseAgent
    enricher func(context.Context, StreamData, domain.BaseAgent) (StreamData, error)
}

func NewEnrichmentStage(name string, agent domain.BaseAgent, enricher func(context.Context, StreamData, domain.BaseAgent) (StreamData, error)) *EnrichmentStage {
    return &EnrichmentStage{
        name:     name,
        agent:    agent,
        enricher: enricher,
    }
}

func (es *EnrichmentStage) Name() string { return es.name }

func (es *EnrichmentStage) ProcessStream(ctx context.Context, input <-chan StreamData) (<-chan StreamData, error) {
    output := make(chan StreamData, 100)
    
    go func() {
        defer close(output)
        
        for {
            select {
            case data, ok := <-input:
                if !ok {
                    return
                }

                enriched, err := es.enricher(ctx, data, es.agent)
                if err != nil {
                    log.Printf("Enrichment error: %v", err)
                    // Still pass through unenriched data
                    enriched = data
                }

                select {
                case output <- enriched:
                case <-ctx.Done():
                    return
                }

            case <-ctx.Done():
                return
            }
        }
    }()

    return output, nil
}

func (es *EnrichmentStage) Close() error {
    return nil
}

// ComplexPipelineExample demonstrates advanced patterns
func ComplexPipelineExample() {
    fmt.Println("🌊 Streaming Pipeline - Advanced Patterns")
    fmt.Println("========================================")

    // Create agent for enrichment
    agent, err := core.NewAgentFromString("enricher", "openai/gpt-4o-mini")
    if err != nil {
        log.Fatalf("Failed to create agent: %v", err)
    }

    // Create streaming pipeline
    config := StreamConfig{
        BufferSize:       1000,
        MaxConcurrency:   10,
        FlushInterval:    5 * time.Second,
        BackpressureMode: "buffer",
        Checkpointing:    true,
    }

    pipeline := NewStreamingPipeline("analytics-pipeline", config)

    // Stage 1: Branching by data type
    branchStage := NewBranchingStage("type-router")
    
    // Add branches for different data types
    eventBranch := branchStage.AddBranch("events", func(data StreamData) bool {
        return data.Type == "event"
    }, 100)
    
    metricBranch := branchStage.AddBranch("metrics", func(data StreamData) bool {
        return data.Type == "metric"
    }, 100)
    
    logBranch := branchStage.AddBranch("logs", func(data StreamData) bool {
        return data.Type == "log"
    }, 100)

    // Stage 2: Parallel processing for events
    eventProcessor := NewParallelProcessingStage("event-processor", 5, 
        func(ctx context.Context, data StreamData) (StreamData, error) {
            // Simulate event processing
            time.Sleep(100 * time.Millisecond)
            
            if metadata, ok := data.Metadata["processed"]; !ok || !metadata.(bool) {
                data.Metadata["processed"] = true
                data.Metadata["processor_id"] = "event-proc-1"
            }
            
            return data, nil
        })

    // Stage 3: Windowed aggregation for metrics
    metricAggregator := NewWindowedAggregationStage("metric-aggregator", 
        10*time.Second,
        func(window []StreamData) StreamData {
            // Aggregate metrics in window
            aggregated := StreamData{
                ID:        fmt.Sprintf("agg-%d", time.Now().Unix()),
                Type:      "aggregated_metric",
                Timestamp: time.Now(),
                Metadata: map[string]interface{}{
                    "window_size": len(window),
                    "start_time":  window[0].Timestamp,
                    "end_time":    window[len(window)-1].Timestamp,
                },
            }
            
            // Calculate aggregations
            values := []float64{}
            for _, data := range window {
                if val, ok := data.Payload.(float64); ok {
                    values = append(values, val)
                }
            }
            
            if len(values) > 0 {
                sum := 0.0
                for _, v := range values {
                    sum += v
                }
                aggregated.Payload = map[string]float64{
                    "sum":   sum,
                    "avg":   sum / float64(len(values)),
                    "count": float64(len(values)),
                }
            }
            
            return aggregated
        })

    // Stage 4: AI enrichment for logs
    logEnricher := NewEnrichmentStage("log-enricher", agent,
        func(ctx context.Context, data StreamData, agent domain.BaseAgent) (StreamData, error) {
            if logContent, ok := data.Payload.(string); ok {
                prompt := fmt.Sprintf(`Analyze this log entry and extract:
1. Severity level (debug, info, warning, error, critical)
2. Component/service name
3. Key information or errors
4. Suggested action (if any)

Log: %s

Return as JSON with fields: severity, component, key_info, action`, logContent)

                state := domain.NewState()
                state.Set("user_input", prompt)

                result, err := agent.Run(ctx, state)
                if err != nil {
                    return data, err
                }

                if response, exists := result.Get("response"); exists {
                    var analysis map[string]interface{}
                    if err := json.Unmarshal([]byte(response.(string)), &analysis); err == nil {
                        data.Metadata["ai_analysis"] = analysis
                    }
                }
            }
            
            return data, nil
        })

    // Connect stages
    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()

    // Process each branch separately
    processedEvents, _ := eventProcessor.ProcessStream(ctx, eventBranch)
    aggregatedMetrics, _ := metricAggregator.ProcessStream(ctx, metricBranch)
    enrichedLogs, _ := logEnricher.ProcessStream(ctx, logBranch)

    // Merge results
    merger := NewStreamMerger("result-merger")
    mergedOutput := merger.Merge(ctx, processedEvents, aggregatedMetrics, enrichedLogs)

    // Generate sample data
    go generateStreamData(pipeline.inputChan)

    // Process results
    processResults(mergedOutput)
}

// StreamMerger combines multiple streams
type StreamMerger struct {
    name string
}

func NewStreamMerger(name string) *StreamMerger {
    return &StreamMerger{name: name}
}

func (sm *StreamMerger) Merge(ctx context.Context, streams ...<-chan StreamData) <-chan StreamData {
    output := make(chan StreamData, 100)
    
    var wg sync.WaitGroup
    
    // Fan-in from all streams
    for i, stream := range streams {
        wg.Add(1)
        go func(streamID int, input <-chan StreamData) {
            defer wg.Done()
            
            for {
                select {
                case data, ok := <-input:
                    if !ok {
                        return
                    }
                    
                    data.Metadata["stream_id"] = streamID
                    
                    select {
                    case output <- data:
                    case <-ctx.Done():
                        return
                    }
                    
                case <-ctx.Done():
                    return
                }
            }
        }(i, stream)
    }
    
    // Close output when all streams done
    go func() {
        wg.Wait()
        close(output)
    }()
    
    return output
}

// StreamingPipeline implementation
func NewStreamingPipeline(name string, config StreamConfig) *StreamingPipeline {
    return &StreamingPipeline{
        name:      name,
        stages:    []StreamStage{},
        config:    config,
        inputChan: make(chan StreamData, config.BufferSize),
        metrics:   &StreamMetrics{},
    }
}

func generateStreamData(output chan<- StreamData) {
    defer close(output)
    
    // Generate different types of data
    for i := 0; i < 100; i++ {
        var data StreamData
        
        switch i % 3 {
        case 0: // Event
            data = StreamData{
                ID:        fmt.Sprintf("event-%d", i),
                Type:      "event",
                Payload:   fmt.Sprintf("User action: clicked button %d", i),
                Timestamp: time.Now(),
                Metadata:  map[string]interface{}{"user_id": fmt.Sprintf("user-%d", i%10)},
            }
        case 1: // Metric
            data = StreamData{
                ID:        fmt.Sprintf("metric-%d", i),
                Type:      "metric",
                Payload:   float64(i * 10 + rand.Intn(20)),
                Timestamp: time.Now(),
                Metadata:  map[string]interface{}{"metric_name": "cpu_usage"},
            }
        case 2: // Log
            severity := []string{"INFO", "WARNING", "ERROR"}[i%3]
            data = StreamData{
                ID:        fmt.Sprintf("log-%d", i),
                Type:      "log",
                Payload:   fmt.Sprintf("[%s] Service health check %d completed", severity, i),
                Timestamp: time.Now(),
                Metadata:  map[string]interface{}{"source": "health-checker"},
            }
        }
        
        output <- data
        time.Sleep(100 * time.Millisecond)
    }
}

func processResults(results <-chan StreamData) {
    counts := map[string]int{}
    
    for result := range results {
        counts[result.Type]++
        
        // Print sample results
        if counts[result.Type] <= 3 {
            fmt.Printf("\n📦 %s Result:\n", result.Type)
            fmt.Printf("  ID: %s\n", result.ID)
            fmt.Printf("  Payload: %v\n", result.Payload)
            if len(result.Metadata) > 0 {
                fmt.Printf("  Metadata: %v\n", result.Metadata)
            }
        }
    }
    
    fmt.Printf("\n📊 Processing Summary:\n")
    for dataType, count := range counts {
        fmt.Printf("  %s: %d processed\n", dataType, count)
    }
}
```

---

## Level 3: Production Data Pipeline Platform
*Enterprise-grade pipeline orchestration and monitoring*

### Data Pipeline Platform
```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/google/uuid"
    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
)

// DataPipelinePlatform orchestrates multiple pipelines
type DataPipelinePlatform struct {
    pipelines       map[string]*ManagedPipeline
    scheduler       *PipelineScheduler
    monitor         *PipelineMonitor
    registry        *StageRegistry
    checkpointer    *Checkpointer
    alertManager    *AlertManager
    mu              sync.RWMutex
}

// ManagedPipeline wraps a pipeline with management features
type ManagedPipeline struct {
    ID              string                 `json:"id"`
    Name            string                 `json:"name"`
    Pipeline        *Pipeline              `json:"-"`
    Schedule        *Schedule              `json:"schedule"`
    Status          PipelineStatus         `json:"status"`
    LastRun         *PipelineRun          `json:"last_run"`
    Configuration   map[string]interface{} `json:"configuration"`
    Tags            []string              `json:"tags"`
    Owner           string                `json:"owner"`
    CreatedAt       time.Time             `json:"created_at"`
    UpdatedAt       time.Time             `json:"updated_at"`
}

type PipelineStatus string

const (
    StatusIdle       PipelineStatus = "idle"
    StatusRunning    PipelineStatus = "running"
    StatusPaused     PipelineStatus = "paused"
    StatusFailed     PipelineStatus = "failed"
    StatusCompleted  PipelineStatus = "completed"
)

// PipelineRun represents a single execution
type PipelineRun struct {
    ID            string                 `json:"id"`
    PipelineID    string                 `json:"pipeline_id"`
    StartTime     time.Time             `json:"start_time"`
    EndTime       *time.Time            `json:"end_time"`
    Status        PipelineStatus         `json:"status"`
    InputCount    int                   `json:"input_count"`
    OutputCount   int                   `json:"output_count"`
    ErrorCount    int                   `json:"error_count"`
    Checkpoints   []Checkpoint          `json:"checkpoints"`
    Metrics       map[string]interface{} `json:"metrics"`
}

// Schedule defines when pipelines run
type Schedule struct {
    Type        string        `json:"type"` // "cron", "interval", "manual"
    Expression  string        `json:"expression"`
    Interval    time.Duration `json:"interval"`
    NextRun     time.Time     `json:"next_run"`
    MaxRuns     int           `json:"max_runs"`
    RunCount    int           `json:"run_count"`
}

// PipelineScheduler manages pipeline execution schedules
type PipelineScheduler struct {
    schedules map[string]*Schedule
    ticker    *time.Ticker
    mu        sync.RWMutex
    platform  *DataPipelinePlatform
}

// PipelineMonitor tracks pipeline health and performance
type PipelineMonitor struct {
    metrics      map[string]*PipelineHealthMetrics
    alerts       chan Alert
    mu           sync.RWMutex
}

type PipelineHealthMetrics struct {
    RunCount            int64
    SuccessRate         float64
    AvgDuration         time.Duration
    AvgThroughput       float64
    LastHealthCheck     time.Time
    HealthScore         float64
    ResourceUsage       ResourceMetrics
}

type ResourceMetrics struct {
    CPUUsage    float64
    MemoryUsage float64
    DiskIO      float64
    NetworkIO   float64
}

// StageRegistry manages reusable pipeline stages
type StageRegistry struct {
    stages    map[string]StageFactory
    metadata  map[string]StageMetadata
    mu        sync.RWMutex
}

type StageFactory func(config map[string]interface{}) (Stage, error)

type StageMetadata struct {
    Name         string                 `json:"name"`
    Description  string                 `json:"description"`
    Category     string                 `json:"category"`
    Version      string                 `json:"version"`
    InputSchema  interface{}           `json:"input_schema"`
    OutputSchema interface{}           `json:"output_schema"`
    Config       map[string]ConfigParam `json:"config"`
}

type ConfigParam struct {
    Type        string      `json:"type"`
    Description string      `json:"description"`
    Required    bool        `json:"required"`
    Default     interface{} `json:"default"`
}

// Checkpointer handles pipeline state persistence
type Checkpointer struct {
    storage CheckpointStorage
    mu      sync.Mutex
}

type Checkpoint struct {
    ID         string                 `json:"id"`
    PipelineID string                 `json:"pipeline_id"`
    RunID      string                 `json:"run_id"`
    Stage      string                 `json:"stage"`
    State      map[string]interface{} `json:"state"`
    Timestamp  time.Time             `json:"timestamp"`
}

type CheckpointStorage interface {
    Save(checkpoint Checkpoint) error
    Load(pipelineID, runID string) ([]Checkpoint, error)
    Delete(pipelineID, runID string) error
}

// AlertManager handles pipeline alerts
type AlertManager struct {
    rules     []AlertRule
    handlers  map[string]AlertHandler
    mu        sync.RWMutex
}

type Alert struct {
    ID         string                 `json:"id"`
    PipelineID string                 `json:"pipeline_id"`
    Type       string                 `json:"type"`
    Severity   string                 `json:"severity"`
    Message    string                 `json:"message"`
    Details    map[string]interface{} `json:"details"`
    Timestamp  time.Time             `json:"timestamp"`
}

type AlertRule struct {
    Name      string
    Condition func(metrics PipelineHealthMetrics) bool
    Severity  string
    Message   string
}

type AlertHandler interface {
    Handle(alert Alert) error
}

// PipelineBuilder provides fluent API for pipeline construction
type PipelineBuilder struct {
    name      string
    stages    []StageConfig
    config    PipelineConfig
    schedule  *Schedule
    tags      []string
    owner     string
}

type StageConfig struct {
    Type   string                 `json:"type"`
    Name   string                 `json:"name"`
    Config map[string]interface{} `json:"config"`
}

func NewDataPipelinePlatform() *DataPipelinePlatform {
    platform := &DataPipelinePlatform{
        pipelines:    make(map[string]*ManagedPipeline),
        registry:     NewStageRegistry(),
        monitor:      NewPipelineMonitor(),
        checkpointer: NewCheckpointer(NewMemoryCheckpointStorage()),
        alertManager: NewAlertManager(),
    }
    
    platform.scheduler = NewPipelineScheduler(platform)
    
    // Register default stages
    platform.registerDefaultStages()
    
    // Set up default alert rules
    platform.setupDefaultAlerts()
    
    return platform
}

func (dpp *DataPipelinePlatform) registerDefaultStages() {
    // Register LLM processing stage
    dpp.registry.Register("llm_processor", 
        StageMetadata{
            Name:        "LLM Processor",
            Description: "Process data using LLM",
            Category:    "processing",
            Version:     "1.0",
            Config: map[string]ConfigParam{
                "provider": {
                    Type:        "string",
                    Description: "LLM provider string",
                    Required:    true,
                },
                "prompt_template": {
                    Type:        "string",
                    Description: "Prompt template with placeholders",
                    Required:    true,
                },
                "max_tokens": {
                    Type:        "integer",
                    Description: "Maximum tokens in response",
                    Default:     1000,
                },
            },
        },
        func(config map[string]interface{}) (Stage, error) {
            provider := config["provider"].(string)
            template := config["prompt_template"].(string)
            
            agent, err := core.NewAgentFromString("processor", provider)
            if err != nil {
                return nil, err
            }
            
            return NewLLMProcessorStage(agent, template), nil
        },
    )

    // Register data validation stage
    dpp.registry.Register("validator",
        StageMetadata{
            Name:        "Data Validator",
            Description: "Validate data against schema",
            Category:    "validation",
            Version:     "1.0",
            Config: map[string]ConfigParam{
                "schema": {
                    Type:        "object",
                    Description: "JSON Schema for validation",
                    Required:    true,
                },
                "strict": {
                    Type:        "boolean",
                    Description: "Strict validation mode",
                    Default:     false,
                },
            },
        },
        func(config map[string]interface{}) (Stage, error) {
            // Implementation
            return nil, nil
        },
    )

    // Register data sink stages
    dpp.registry.Register("database_sink",
        StageMetadata{
            Name:        "Database Sink",
            Description: "Write data to database",
            Category:    "sink",
            Version:     "1.0",
            Config: map[string]ConfigParam{
                "connection_string": {
                    Type:        "string",
                    Description: "Database connection string",
                    Required:    true,
                },
                "table": {
                    Type:        "string",
                    Description: "Target table name",
                    Required:    true,
                },
                "batch_size": {
                    Type:        "integer",
                    Description: "Batch insert size",
                    Default:     100,
                },
            },
        },
        func(config map[string]interface{}) (Stage, error) {
            // Implementation
            return nil, nil
        },
    )
}

func (dpp *DataPipelinePlatform) setupDefaultAlerts() {
    // High error rate alert
    dpp.alertManager.AddRule(AlertRule{
        Name: "high_error_rate",
        Condition: func(metrics PipelineHealthMetrics) bool {
            return metrics.SuccessRate < 0.8 // Less than 80% success
        },
        Severity: "high",
        Message:  "Pipeline success rate below 80%",
    })

    // Performance degradation alert
    dpp.alertManager.AddRule(AlertRule{
        Name: "performance_degradation",
        Condition: func(metrics PipelineHealthMetrics) bool {
            // Alert if throughput drops by 50%
            return metrics.AvgThroughput < metrics.AvgThroughput*0.5
        },
        Severity: "medium",
        Message:  "Pipeline throughput degraded by 50%",
    })

    // Resource usage alert
    dpp.alertManager.AddRule(AlertRule{
        Name: "high_resource_usage",
        Condition: func(metrics PipelineHealthMetrics) bool {
            return metrics.ResourceUsage.CPUUsage > 0.9 || 
                   metrics.ResourceUsage.MemoryUsage > 0.9
        },
        Severity: "high",
        Message:  "High resource usage detected",
    })
}

// Pipeline CRUD operations
func (dpp *DataPipelinePlatform) CreatePipeline(builder *PipelineBuilder) (*ManagedPipeline, error) {
    dpp.mu.Lock()
    defer dpp.mu.Unlock()

    // Build pipeline from configuration
    pipeline := NewPipeline(builder.name, builder.config)
    
    // Add stages
    for _, stageConfig := range builder.stages {
        factory, err := dpp.registry.GetFactory(stageConfig.Type)
        if err != nil {
            return nil, fmt.Errorf("unknown stage type: %s", stageConfig.Type)
        }
        
        stage, err := factory(stageConfig.Config)
        if err != nil {
            return nil, fmt.Errorf("failed to create stage %s: %w", stageConfig.Name, err)
        }
        
        pipeline.AddStage(stage)
    }

    // Create managed pipeline
    managed := &ManagedPipeline{
        ID:            uuid.New().String(),
        Name:          builder.name,
        Pipeline:      pipeline,
        Schedule:      builder.schedule,
        Status:        StatusIdle,
        Configuration: make(map[string]interface{}),
        Tags:          builder.tags,
        Owner:         builder.owner,
        CreatedAt:     time.Now(),
        UpdatedAt:     time.Now(),
    }

    dpp.pipelines[managed.ID] = managed

    // Schedule if needed
    if managed.Schedule != nil {
        dpp.scheduler.Schedule(managed.ID, managed.Schedule)
    }

    log.Printf("✅ Created pipeline: %s (%s)", managed.Name, managed.ID)
    return managed, nil
}

func (dpp *DataPipelinePlatform) GetPipeline(id string) (*ManagedPipeline, error) {
    dpp.mu.RLock()
    defer dpp.mu.RUnlock()

    pipeline, exists := dpp.pipelines[id]
    if !exists {
        return nil, fmt.Errorf("pipeline not found: %s", id)
    }

    return pipeline, nil
}

func (dpp *DataPipelinePlatform) ListPipelines(filter PipelineFilter) []*ManagedPipeline {
    dpp.mu.RLock()
    defer dpp.mu.RUnlock()

    pipelines := []*ManagedPipeline{}
    
    for _, pipeline := range dpp.pipelines {
        if filter.Matches(pipeline) {
            pipelines = append(pipelines, pipeline)
        }
    }

    return pipelines
}

type PipelineFilter struct {
    Tags   []string
    Owner  string
    Status PipelineStatus
}

func (pf PipelineFilter) Matches(pipeline *ManagedPipeline) bool {
    // Check tags
    if len(pf.Tags) > 0 {
        hasTag := false
        for _, filterTag := range pf.Tags {
            for _, pipelineTag := range pipeline.Tags {
                if filterTag == pipelineTag {
                    hasTag = true
                    break
                }
            }
            if hasTag {
                break
            }
        }
        if !hasTag {
            return false
        }
    }

    // Check owner
    if pf.Owner != "" && pipeline.Owner != pf.Owner {
        return false
    }

    // Check status
    if pf.Status != "" && pipeline.Status != pf.Status {
        return false
    }

    return true
}

// Pipeline execution
func (dpp *DataPipelinePlatform) RunPipeline(id string, input interface{}) (*PipelineRun, error) {
    pipeline, err := dpp.GetPipeline(id)
    if err != nil {
        return nil, err
    }

    if pipeline.Status == StatusRunning {
        return nil, fmt.Errorf("pipeline already running")
    }

    // Create run record
    run := &PipelineRun{
        ID:         uuid.New().String(),
        PipelineID: id,
        StartTime:  time.Now(),
        Status:     StatusRunning,
        Metrics:    make(map[string]interface{}),
    }

    // Update pipeline status
    pipeline.Status = StatusRunning
    pipeline.LastRun = run

    // Execute pipeline
    go func() {
        ctx := context.Background()
        
        // Run with monitoring
        err := dpp.runWithMonitoring(ctx, pipeline, run)
        
        // Update run status
        endTime := time.Now()
        run.EndTime = &endTime
        
        if err != nil {
            run.Status = StatusFailed
            pipeline.Status = StatusFailed
            
            // Send alert
            dpp.alertManager.SendAlert(Alert{
                ID:         uuid.New().String(),
                PipelineID: pipeline.ID,
                Type:       "pipeline_failure",
                Severity:   "high",
                Message:    fmt.Sprintf("Pipeline %s failed: %v", pipeline.Name, err),
                Timestamp:  time.Now(),
            })
        } else {
            run.Status = StatusCompleted
            pipeline.Status = StatusCompleted
        }
        
        // Update metrics
        dpp.monitor.UpdateMetrics(pipeline.ID, run)
    }()

    return run, nil
}

func (dpp *DataPipelinePlatform) runWithMonitoring(ctx context.Context, pipeline *ManagedPipeline, run *PipelineRun) error {
    // Set up monitoring
    monitorCtx, cancel := context.WithCancel(ctx)
    defer cancel()

    // Monitor resource usage
    go dpp.monitorResources(monitorCtx, pipeline.ID)

    // Execute pipeline with checkpointing
    checkpointHandler := func(stageName string, data interface{}) {
        checkpoint := Checkpoint{
            ID:         uuid.New().String(),
            PipelineID: pipeline.ID,
            RunID:      run.ID,
            Stage:      stageName,
            State:      map[string]interface{}{"data": data},
            Timestamp:  time.Now(),
        }
        
        if err := dpp.checkpointer.Save(checkpoint); err != nil {
            log.Printf("Checkpoint save failed: %v", err)
        } else {
            run.Checkpoints = append(run.Checkpoints, checkpoint)
        }
    }

    // Run pipeline
    return pipeline.Pipeline.RunWithCheckpoints(ctx, checkpointHandler)
}

func (dpp *DataPipelinePlatform) monitorResources(ctx context.Context, pipelineID string) {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            // Collect resource metrics (simplified)
            metrics := ResourceMetrics{
                CPUUsage:    0.5,  // In practice, get real metrics
                MemoryUsage: 0.6,
                DiskIO:      0.3,
                NetworkIO:   0.2,
            }
            
            dpp.monitor.UpdateResourceMetrics(pipelineID, metrics)

        case <-ctx.Done():
            return
        }
    }
}

// Pipeline lifecycle management
func (dpp *DataPipelinePlatform) PausePipeline(id string) error {
    pipeline, err := dpp.GetPipeline(id)
    if err != nil {
        return err
    }

    if pipeline.Status != StatusRunning {
        return fmt.Errorf("pipeline not running")
    }

    pipeline.Status = StatusPaused
    // Implementation: Signal pipeline to pause
    
    return nil
}

func (dpp *DataPipelinePlatform) ResumePipeline(id string) error {
    pipeline, err := dpp.GetPipeline(id)
    if err != nil {
        return err
    }

    if pipeline.Status != StatusPaused {
        return fmt.Errorf("pipeline not paused")
    }

    pipeline.Status = StatusRunning
    // Implementation: Signal pipeline to resume
    
    return nil
}

func (dpp *DataPipelinePlatform) DeletePipeline(id string) error {
    dpp.mu.Lock()
    defer dpp.mu.Unlock()

    pipeline, exists := dpp.pipelines[id]
    if !exists {
        return fmt.Errorf("pipeline not found")
    }

    if pipeline.Status == StatusRunning {
        return fmt.Errorf("cannot delete running pipeline")
    }

    // Remove from scheduler
    dpp.scheduler.Unschedule(id)

    // Delete pipeline
    delete(dpp.pipelines, id)

    return nil
}

// Helper implementations
func NewStageRegistry() *StageRegistry {
    return &StageRegistry{
        stages:   make(map[string]StageFactory),
        metadata: make(map[string]StageMetadata),
    }
}

func (sr *StageRegistry) Register(name string, metadata StageMetadata, factory StageFactory) {
    sr.mu.Lock()
    defer sr.mu.Unlock()
    
    sr.stages[name] = factory
    sr.metadata[name] = metadata
}

func (sr *StageRegistry) GetFactory(name string) (StageFactory, error) {
    sr.mu.RLock()
    defer sr.mu.RUnlock()
    
    factory, exists := sr.stages[name]
    if !exists {
        return nil, fmt.Errorf("stage not found: %s", name)
    }
    
    return factory, nil
}

func NewPipelineMonitor() *PipelineMonitor {
    return &PipelineMonitor{
        metrics: make(map[string]*PipelineHealthMetrics),
        alerts:  make(chan Alert, 100),
    }
}

func (pm *PipelineMonitor) UpdateMetrics(pipelineID string, run *PipelineRun) {
    pm.mu.Lock()
    defer pm.mu.Unlock()

    if _, exists := pm.metrics[pipelineID]; !exists {
        pm.metrics[pipelineID] = &PipelineHealthMetrics{}
    }

    metrics := pm.metrics[pipelineID]
    metrics.RunCount++
    
    if run.Status == StatusCompleted {
        metrics.SuccessRate = (metrics.SuccessRate*float64(metrics.RunCount-1) + 1) / float64(metrics.RunCount)
    } else {
        metrics.SuccessRate = metrics.SuccessRate * float64(metrics.RunCount-1) / float64(metrics.RunCount)
    }

    if run.EndTime != nil {
        duration := run.EndTime.Sub(run.StartTime)
        if metrics.AvgDuration == 0 {
            metrics.AvgDuration = duration
        } else {
            metrics.AvgDuration = (metrics.AvgDuration + duration) / 2
        }
    }

    metrics.LastHealthCheck = time.Now()
    metrics.HealthScore = metrics.SuccessRate * 0.5 + (1-metrics.ResourceUsage.CPUUsage)*0.25 + (1-metrics.ResourceUsage.MemoryUsage)*0.25
}

func (pm *PipelineMonitor) UpdateResourceMetrics(pipelineID string, resources ResourceMetrics) {
    pm.mu.Lock()
    defer pm.mu.Unlock()

    if _, exists := pm.metrics[pipelineID]; !exists {
        pm.metrics[pipelineID] = &PipelineHealthMetrics{}
    }

    pm.metrics[pipelineID].ResourceUsage = resources
}

func NewCheckpointer(storage CheckpointStorage) *Checkpointer {
    return &Checkpointer{
        storage: storage,
    }
}

func (c *Checkpointer) Save(checkpoint Checkpoint) error {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    return c.storage.Save(checkpoint)
}

func NewAlertManager() *AlertManager {
    return &AlertManager{
        rules:    []AlertRule{},
        handlers: make(map[string]AlertHandler),
    }
}

func (am *AlertManager) AddRule(rule AlertRule) {
    am.mu.Lock()
    defer am.mu.Unlock()
    
    am.rules = append(am.rules, rule)
}

func (am *AlertManager) SendAlert(alert Alert) {
    am.mu.RLock()
    defer am.mu.RUnlock()

    // Send to all handlers
    for name, handler := range am.handlers {
        if err := handler.Handle(alert); err != nil {
            log.Printf("Alert handler %s failed: %v", name, err)
        }
    }
}

func NewPipelineScheduler(platform *DataPipelinePlatform) *PipelineScheduler {
    scheduler := &PipelineScheduler{
        schedules: make(map[string]*Schedule),
        platform:  platform,
    }
    
    // Start scheduler loop
    go scheduler.run()
    
    return scheduler
}

func (ps *PipelineScheduler) Schedule(pipelineID string, schedule *Schedule) {
    ps.mu.Lock()
    defer ps.mu.Unlock()
    
    ps.schedules[pipelineID] = schedule
    
    // Calculate next run time
    switch schedule.Type {
    case "interval":
        schedule.NextRun = time.Now().Add(schedule.Interval)
    case "cron":
        // Parse cron expression and calculate next run
        // Implementation depends on cron library
    }
}

func (ps *PipelineScheduler) Unschedule(pipelineID string) {
    ps.mu.Lock()
    defer ps.mu.Unlock()
    
    delete(ps.schedules, pipelineID)
}

func (ps *PipelineScheduler) run() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()

    for range ticker.C {
        ps.checkSchedules()
    }
}

func (ps *PipelineScheduler) checkSchedules() {
    ps.mu.RLock()
    defer ps.mu.RUnlock()

    now := time.Now()
    
    for pipelineID, schedule := range ps.schedules {
        if now.After(schedule.NextRun) {
            // Run pipeline
            _, err := ps.platform.RunPipeline(pipelineID, nil)
            if err != nil {
                log.Printf("Scheduled pipeline run failed: %v", err)
            }
            
            // Update schedule
            schedule.RunCount++
            
            // Calculate next run
            switch schedule.Type {
            case "interval":
                schedule.NextRun = now.Add(schedule.Interval)
            }
            
            // Check max runs
            if schedule.MaxRuns > 0 && schedule.RunCount >= schedule.MaxRuns {
                ps.Unschedule(pipelineID)
            }
        }
    }
}

// Memory checkpoint storage (for demo)
type MemoryCheckpointStorage struct {
    checkpoints map[string][]Checkpoint
    mu          sync.RWMutex
}

func NewMemoryCheckpointStorage() *MemoryCheckpointStorage {
    return &MemoryCheckpointStorage{
        checkpoints: make(map[string][]Checkpoint),
    }
}

func (mcs *MemoryCheckpointStorage) Save(checkpoint Checkpoint) error {
    mcs.mu.Lock()
    defer mcs.mu.Unlock()
    
    key := checkpoint.PipelineID + ":" + checkpoint.RunID
    mcs.checkpoints[key] = append(mcs.checkpoints[key], checkpoint)
    
    return nil
}

func (mcs *MemoryCheckpointStorage) Load(pipelineID, runID string) ([]Checkpoint, error) {
    mcs.mu.RLock()
    defer mcs.mu.RUnlock()
    
    key := pipelineID + ":" + runID
    return mcs.checkpoints[key], nil
}

func (mcs *MemoryCheckpointStorage) Delete(pipelineID, runID string) error {
    mcs.mu.Lock()
    defer mcs.mu.Unlock()
    
    key := pipelineID + ":" + runID
    delete(mcs.checkpoints, key)
    
    return nil
}

// Pipeline builder
func NewPipelineBuilder(name string) *PipelineBuilder {
    return &PipelineBuilder{
        name:   name,
        stages: []StageConfig{},
        config: PipelineConfig{
            MaxConcurrency: 10,
            BatchSize:      100,
            ErrorStrategy:  "skip",
            RetryAttempts:  3,
            RetryDelay:     time.Second,
            Timeout:        5 * time.Minute,
            EnableMetrics:  true,
        },
    }
}

func (pb *PipelineBuilder) AddStage(stageType, name string, config map[string]interface{}) *PipelineBuilder {
    pb.stages = append(pb.stages, StageConfig{
        Type:   stageType,
        Name:   name,
        Config: config,
    })
    return pb
}

func (pb *PipelineBuilder) WithSchedule(schedule *Schedule) *PipelineBuilder {
    pb.schedule = schedule
    return pb
}

func (pb *PipelineBuilder) WithTags(tags ...string) *PipelineBuilder {
    pb.tags = tags
    return pb
}

func (pb *PipelineBuilder) WithOwner(owner string) *PipelineBuilder {
    pb.owner = owner
    return pb
}

func (pb *PipelineBuilder) WithConfig(config PipelineConfig) *PipelineBuilder {
    pb.config = config
    return pb
}

// Example LLM processor stage
type LLMProcessorStage struct {
    agent    domain.BaseAgent
    template string
}

func NewLLMProcessorStage(agent domain.BaseAgent, template string) *LLMProcessorStage {
    return &LLMProcessorStage{
        agent:    agent,
        template: template,
    }
}

func (lps *LLMProcessorStage) Name() string { return "llm_processor" }

func (lps *LLMProcessorStage) Configure(config map[string]interface{}) error {
    return nil
}

func (lps *LLMProcessorStage) Process(ctx context.Context, data interface{}) (interface{}, error) {
    // Process data with LLM
    return data, nil
}

// RunWithCheckpoints extends Pipeline to support checkpointing
func (p *Pipeline) RunWithCheckpoints(ctx context.Context, checkpointHandler func(string, interface{})) error {
    var data interface{}
    
    for _, stage := range p.stages {
        result, err := stage.Process(ctx, data)
        if err != nil {
            return err
        }
        
        // Save checkpoint
        checkpointHandler(stage.Name(), result)
        
        data = result
    }
    
    return nil
}

// Example usage
func main() {
    fmt.Println("🏭 Enterprise Data Pipeline Platform")
    fmt.Println("===================================")

    // Create platform
    platform := NewDataPipelinePlatform()

    // Build a pipeline
    builder := NewPipelineBuilder("customer-insights-pipeline").
        WithOwner("data-team").
        WithTags("production", "customer", "analytics").
        AddStage("file_source", "customer_data_source", map[string]interface{}{
            "directory": "./customer_data",
            "pattern":   "*.json",
        }).
        AddStage("llm_processor", "insight_extractor", map[string]interface{}{
            "provider":        "openai/gpt-4o",
            "prompt_template": "Extract customer insights from: {{.data}}",
            "max_tokens":      500,
        }).
        AddStage("validator", "insight_validator", map[string]interface{}{
            "schema": map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "customer_id": map[string]interface{}{"type": "string"},
                    "insights":    map[string]interface{}{"type": "array"},
                    "sentiment":   map[string]interface{}{"type": "string"},
                },
                "required": []string{"customer_id", "insights"},
            },
        }).
        AddStage("database_sink", "postgres_writer", map[string]interface{}{
            "connection_string": "postgres://localhost/insights",
            "table":            "customer_insights",
            "batch_size":       50,
        }).
        WithSchedule(&Schedule{
            Type:     "interval",
            Interval: 1 * time.Hour,
            MaxRuns:  24, // Run for 24 hours
        })

    // Create pipeline
    pipeline, err := platform.CreatePipeline(builder)
    if err != nil {
        log.Fatalf("Failed to create pipeline: %v", err)
    }

    fmt.Printf("✅ Created pipeline: %s\n", pipeline.ID)

    // List pipelines
    pipelines := platform.ListPipelines(PipelineFilter{
        Tags: []string{"production"},
    })

    fmt.Printf("\n📋 Production Pipelines:\n")
    for _, p := range pipelines {
        fmt.Printf("  - %s (%s): %s\n", p.Name, p.ID, p.Status)
    }

    // Run pipeline manually
    fmt.Printf("\n🚀 Running pipeline...\n")
    run, err := platform.RunPipeline(pipeline.ID, nil)
    if err != nil {
        log.Printf("Failed to run pipeline: %v", err)
    } else {
        fmt.Printf("Pipeline run started: %s\n", run.ID)
    }

    // Wait for completion (in practice, use proper synchronization)
    time.Sleep(5 * time.Second)

    // Get pipeline metrics
    fmt.Printf("\n📊 Pipeline Health:\n")
    if metrics, exists := platform.monitor.metrics[pipeline.ID]; exists {
        fmt.Printf("  Success Rate: %.2f%%\n", metrics.SuccessRate*100)
        fmt.Printf("  Avg Duration: %v\n", metrics.AvgDuration)
        fmt.Printf("  Health Score: %.2f\n", metrics.HealthScore)
    }
}
```

## Pipeline Best Practices

### 1. Design Principles
- **Single Responsibility** - Each stage does one thing well
- **Idempotency** - Stages can be safely retried
- **Statelessness** - Avoid state between runs
- **Composability** - Build complex pipelines from simple stages

### 2. Error Handling
- **Graceful Degradation** - Continue processing valid data
- **Dead Letter Queues** - Store failed items for analysis
- **Circuit Breakers** - Prevent cascade failures
- **Retry Strategies** - Exponential backoff with jitter

### 3. Performance Optimization
- **Parallel Processing** - Use worker pools for CPU-bound tasks
- **Batch Operations** - Process data in chunks
- **Streaming** - Handle large datasets without loading all in memory
- **Caching** - Cache expensive operations

### 4. Monitoring and Observability
- **Metrics Collection** - Track throughput, latency, errors
- **Distributed Tracing** - Trace data through pipeline stages
- **Logging** - Structured logs with correlation IDs
- **Alerting** - Proactive notification of issues

## Common Pipeline Patterns

### ETL (Extract, Transform, Load)
- Extract from multiple sources
- Transform with LLM enrichment
- Load to data warehouse

### Real-time Processing
- Stream ingestion
- Windowed aggregation
- Real-time analytics

### Batch Processing
- Scheduled execution
- Large-scale transformation
- Report generation

### Event-Driven
- Event sourcing
- CQRS patterns
- Async processing

## Troubleshooting

### Common Issues

**"Pipeline stuck" problems**
- Check for blocking operations
- Verify timeout configurations
- Monitor resource usage
- Review stage dependencies

**Memory issues**
- Implement streaming for large data
- Use batch processing
- Clear intermediate results
- Profile memory usage

**Performance bottlenecks**
- Identify slow stages
- Add parallel processing
- Optimize LLM calls
- Cache repeated operations

**Data quality issues**
- Add validation stages
- Implement error recovery
- Use dead letter queues
- Monitor data anomalies

## Next Steps

🚀 **Pipeline mastery achieved!** Continue with:

- **[Web Applications](web-applications.md)** - Integrate pipelines with web apps
- **[APIs and Services](apis-and-services.md)** - Build pipeline APIs
- **[Production Deployment](../advanced/production-deployment.md)** - Deploy pipelines at scale
- **[Performance Optimization](../advanced/performance-optimization.md)** - Optimize pipeline performance

### Quick Reference

- **[Configuration Reference](../reference/configuration-reference.md)** - Pipeline configuration
- **[Best Practices Checklist](../reference/best-practices-checklist.md)** - Pipeline checklist
- **[Error Codes Reference](../reference/error-codes-reference.md)** - Common pipeline errors

---

**Need help with pipelines?** Check our [pipeline examples](../examples/pipeline-patterns.md) or join the discussion on [GitHub](https://github.com/lexlapax/go-llms/discussions).