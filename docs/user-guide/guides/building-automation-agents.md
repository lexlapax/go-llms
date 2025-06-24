# Building Automation Agents: Task Automation Workflows

> **[Project Root](/) / [Documentation](../..) / [User Guide](../../user-guide) / [Guides](../../user-guide/guides) / Building Automation Agents**

Master the creation of intelligent automation agents that can execute system commands, process files, call APIs, and orchestrate complex workflows. Build reliable systems that handle business processes, data pipelines, and infrastructure automation.

## Why Automation Agents Matter

- **Process Automation** - Eliminate repetitive manual tasks with intelligent workflows
- **System Integration** - Connect disparate systems and data sources seamlessly
- **Error Recovery** - Built-in retry logic and graceful failure handling
- **Scalable Orchestration** - From simple scripts to complex multi-agent systems
- **Production Ready** - Monitoring, logging, and operational excellence

## Automation Architecture

![Automation Agent Flow](../../images/automation-flow.svg)

### Core Components
1. **System Integration** - Execute commands, manage processes, handle environment
2. **File Operations** - Automated file processing, monitoring, and transformations
3. **Data Pipelines** - Process CSV, JSON, XML data with validation
4. **Web Automation** - API calls, web scraping, notification systems
5. **Workflow Orchestration** - Sequential, parallel, and conditional task coordination

### Tool Categories
| Category | Tools | Purpose |
|----------|-------|---------|
| **System** | Execute, Environment, Process | Command execution and system management |
| **File** | Read, Write, Search, Monitor | File system automation |
| **Data** | JSON, CSV, XML processing | Data transformation pipelines |
| **Web** | HTTP, Fetch, API clients | External system integration |
| **Workflow** | Sequential, Parallel, Conditional | Task orchestration |
| **DateTime** | Parse, Calculate, Format | Time-based automation |

## Prerequisites

- [Creating Agents guide completed](creating-agents.md) ✅
- [Agent Tools guide helpful](agent-tools.md) ✅
- Basic understanding of system administration ✅

---

## Level 1: Simple Task Automation
*Automate basic system tasks in 15 minutes*

### File Processing Automation
```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/file"
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/system"
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/data"
)

func main() {
    fmt.Println("🤖 Simple Automation Agent")
    fmt.Println("==========================")

    // Create automation agent
    agent, err := core.NewAgentFromString("file-processor", "anthropic/claude-3-5-sonnet")
    if err != nil {
        log.Fatalf("Failed to create agent: %v", err)
    }

    agent.SetSystemPrompt(`You are a file processing automation agent.
    You can:
    1. Read and analyze files
    2. Process data in various formats
    3. Execute safe system commands
    4. Write processed results
    
    Always validate inputs and provide detailed progress updates.
    Be careful with file operations and confirm before destructive actions.`)

    // Add file processing tools
    agent.AddTool(file.NewReadFileTool())
    agent.AddTool(file.NewWriteFileTool())
    agent.AddTool(file.NewSearchFileTool())
    
    // Add data processing tools
    agent.AddTool(data.NewJSONProcessTool())
    agent.AddTool(data.NewCSVProcessTool())
    
    // Add safe system execution
    agent.AddTool(system.NewExecuteCommandTool())

    // Automation tasks
    tasks := []string{
        "Find all JSON files in the current directory and count their total lines",
        "Read the package.json file and extract all dependencies into a CSV",
        "Create a backup directory and copy all .md files there",
        "Search for TODO comments in all Go files and create a summary report",
    }

    for i, task := range tasks {
        fmt.Printf("\n--- Automation Task %d ---\n", i+1)
        fmt.Printf("Task: %s\n", task)

        // Create task state
        state := domain.NewState()
        state.Set("user_input", task)
        
        // Configure safety settings
        state.Set("command_safe_mode", true)
        state.Set("file_restricted_paths", []string{"/tmp", "."})
        state.Set("max_file_size", 10*1024*1024) // 10MB limit

        // Execute automation
        result, err := agent.Run(context.Background(), state)
        if err != nil {
            log.Printf("Task failed: %v", err)
            continue
        }

        if response, exists := result.Get("response"); exists {
            fmt.Printf("\n✅ Task Result:\n%v\n", response)
        }
        
        // Show any files created
        if files, exists := result.Get("files_created"); exists {
            fmt.Printf("\n📁 Files Created: %v\n", files)
        }
    }
}
```

### System Monitoring Automation
```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/system"
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/datetime"
)

// SystemMonitor automates system health checks
type SystemMonitor struct {
    agent       domain.BaseAgent
    checkInterval time.Duration
    alertWebhook  string
}

func NewSystemMonitor(webhookURL string) (*SystemMonitor, error) {
    agent, err := core.NewAgentFromString("system-monitor", "gemini/gemini-2.0-flash")
    if err != nil {
        return nil, err
    }

    agent.SetSystemPrompt(`You are a system monitoring agent.
    Monitor system health by checking:
    1. CPU and memory usage
    2. Disk space availability
    3. Running processes
    4. System uptime and load
    
    Generate alerts for critical conditions and provide actionable recommendations.`)

    // Add monitoring tools
    agent.AddTool(system.NewExecuteCommandTool())
    agent.AddTool(system.NewProcessListTool())
    agent.AddTool(system.NewSystemInfoTool())
    agent.AddTool(web.NewHTTPRequestTool())
    agent.AddTool(datetime.NewDateTimeNowTool())

    return &SystemMonitor{
        agent:         agent,
        checkInterval: 5 * time.Minute,
        alertWebhook:  webhookURL,
    }, nil
}

func (sm *SystemMonitor) StartMonitoring(ctx context.Context) {
    fmt.Printf("📊 Starting system monitoring (interval: %v)\n", sm.checkInterval)
    
    ticker := time.NewTicker(sm.checkInterval)
    defer ticker.Stop()

    // Initial check
    sm.performHealthCheck(ctx)

    for {
        select {
        case <-ticker.C:
            sm.performHealthCheck(ctx)
        case <-ctx.Done():
            fmt.Println("🛑 Monitoring stopped")
            return
        }
    }
}

func (sm *SystemMonitor) performHealthCheck(ctx context.Context) {
    fmt.Printf("\n🔍 Performing health check at %v\n", time.Now().Format("15:04:05"))

    state := domain.NewState()
    state.Set("user_input", `Perform a comprehensive system health check:
    
    1. Check CPU and memory usage
    2. Check disk space on all mounted filesystems
    3. List top 10 processes by CPU usage
    4. Check system uptime and load average
    5. Alert if any metrics exceed thresholds:
       - CPU > 80%
       - Memory > 85%
       - Disk > 90%
       - Load average > number of CPUs
    
    If any thresholds are exceeded, format an alert message and send to webhook.`)
    
    // Configure monitoring parameters
    state.Set("alert_webhook", sm.alertWebhook)
    state.Set("cpu_threshold", 80.0)
    state.Set("memory_threshold", 85.0)
    state.Set("disk_threshold", 90.0)

    result, err := sm.agent.Run(ctx, state)
    if err != nil {
        log.Printf("Health check failed: %v", err)
        return
    }

    if status, exists := result.Get("response"); exists {
        fmt.Printf("📋 System Status:\n%v\n", status)
    }

    if alerts, exists := result.Get("alerts_sent"); exists {
        if alertCount, ok := alerts.(int); ok && alertCount > 0 {
            fmt.Printf("🚨 %d alerts sent\n", alertCount)
        }
    }
}

func main() {
    fmt.Println("📊 System Monitoring Automation")
    fmt.Println("==============================")

    webhookURL := os.Getenv("MONITORING_WEBHOOK_URL")
    if webhookURL == "" {
        webhookURL = "https://hooks.slack.com/services/your/webhook/url"
        log.Printf("Using default webhook URL: %s", webhookURL)
    }

    monitor, err := NewSystemMonitor(webhookURL)
    if err != nil {
        log.Fatalf("Failed to create monitor: %v", err)
    }

    // Run monitoring for a limited time (in production, this would run indefinitely)
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
    defer cancel()

    monitor.StartMonitoring(ctx)
}
```

### Key Features
✅ **Safe Execution** - Sandboxed command execution with allowlists  
✅ **File Automation** - Read, write, search, and transform files  
✅ **System Monitoring** - CPU, memory, disk, and process tracking  
✅ **Alert Integration** - Webhook notifications for critical events  

---

## Level 2: Data Pipeline Automation
*Build intelligent data processing workflows*

### Multi-Source Data Pipeline
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
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/data"
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/file"
    "github.com/lexlapax/go-llms/pkg/agent/workflow"
)

// DataPipelineAgent orchestrates complex data processing workflows
type DataPipelineAgent struct {
    extractorAgent   domain.BaseAgent
    transformerAgent domain.BaseAgent
    validatorAgent   domain.BaseAgent
    loaderAgent      domain.BaseAgent
    
    webTools        []domain.Tool
    dataTools       []domain.Tool
    fileTools       []domain.Tool
    
    config          *PipelineConfig
}

type PipelineConfig struct {
    InputSources     []DataSource
    OutputTargets    []DataTarget
    TransformRules   []TransformRule
    ValidationSchema map[string]interface{}
    BatchSize        int
    RetryAttempts    int
    ErrorThreshold   float64
}

type DataSource struct {
    Type        string                 `json:"type"`        // "api", "csv", "json", "rss"
    URL         string                 `json:"url"`
    Format      string                 `json:"format"`
    Headers     map[string]string      `json:"headers"`
    Auth        map[string]interface{} `json:"auth"`
    Schedule    string                 `json:"schedule"`    // cron-like
}

type DataTarget struct {
    Type     string            `json:"type"`     // "file", "api", "database"
    Path     string            `json:"path"`
    Format   string            `json:"format"`
    Headers  map[string]string `json:"headers"`
    Append   bool              `json:"append"`
}

type TransformRule struct {
    Field     string      `json:"field"`
    Operation string      `json:"operation"`  // "map", "filter", "aggregate", "validate"
    Config    interface{} `json:"config"`
}

func NewDataPipelineAgent(config *PipelineConfig) (*DataPipelineAgent, error) {
    // Create specialized agents for each pipeline stage
    extractor, err := core.NewAgentFromString("data-extractor", "gemini/gemini-2.0-flash")
    if err != nil {
        return nil, err
    }
    extractor.SetSystemPrompt(`You are a data extraction specialist. Extract structured data from various sources (APIs, files, feeds) and ensure consistent formatting.`)

    transformer, err := core.NewAgentFromString("data-transformer", "anthropic/claude-3-5-sonnet")
    if err != nil {
        return nil, err
    }
    transformer.SetSystemPrompt(`You are a data transformation expert. Apply business rules, clean data, and convert between formats while maintaining data integrity.`)

    validator, err := core.NewAgentFromString("data-validator", "openai/gpt-4o-mini")
    if err != nil {
        return nil, err
    }
    validator.SetSystemPrompt(`You are a data quality specialist. Validate data against schemas, identify anomalies, and ensure data meets quality standards.`)

    loader, err := core.NewAgentFromString("data-loader", "gemini/gemini-2.0-flash")
    if err != nil {
        return nil, err
    }
    loader.SetSystemPrompt(`You are a data loading specialist. Efficiently write data to target systems with proper error handling and validation.`)

    return &DataPipelineAgent{
        extractorAgent:   extractor,
        transformerAgent: transformer,
        validatorAgent:   validator,
        loaderAgent:      loader,
        webTools:         []domain.Tool{web.NewHTTPRequestTool(), web.NewWebFetchTool()},
        dataTools:        []domain.Tool{data.NewJSONProcessTool(), data.NewCSVProcessTool(), data.NewXMLProcessTool()},
        fileTools:        []domain.Tool{file.NewReadFileTool(), file.NewWriteFileTool()},
        config:           config,
    }, nil
}

func (dpa *DataPipelineAgent) ExecutePipeline(ctx context.Context) error {
    fmt.Printf("🔄 Starting data pipeline execution\n")
    fmt.Printf("Sources: %d, Targets: %d\n", len(dpa.config.InputSources), len(dpa.config.OutputTargets))

    // Phase 1: Extract data from all sources
    extractedData, err := dpa.extractData(ctx)
    if err != nil {
        return fmt.Errorf("extraction failed: %w", err)
    }
    fmt.Printf("📥 Extracted %d data batches\n", len(extractedData))

    // Phase 2: Transform data according to rules
    transformedData, err := dpa.transformData(ctx, extractedData)
    if err != nil {
        return fmt.Errorf("transformation failed: %w", err)
    }
    fmt.Printf("🔄 Transformed %d data batches\n", len(transformedData))

    // Phase 3: Validate data quality
    validatedData, err := dpa.validateData(ctx, transformedData)
    if err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    fmt.Printf("✅ Validated %d data batches\n", len(validatedData))

    // Phase 4: Load data to targets
    err = dpa.loadData(ctx, validatedData)
    if err != nil {
        return fmt.Errorf("loading failed: %w", err)
    }
    fmt.Printf("💾 Successfully loaded data to %d targets\n", len(dpa.config.OutputTargets))

    return nil
}

func (dpa *DataPipelineAgent) extractData(ctx context.Context) ([]DataBatch, error) {
    var mu sync.Mutex
    var wg sync.WaitGroup
    var allBatches []DataBatch

    // Add tools to extractor
    for _, tool := range dpa.webTools {
        dpa.extractorAgent.AddTool(tool)
    }
    for _, tool := range dpa.dataTools {
        dpa.extractorAgent.AddTool(tool)
    }
    for _, tool := range dpa.fileTools {
        dpa.extractorAgent.AddTool(tool)
    }

    // Process each source concurrently
    for _, source := range dpa.config.InputSources {
        wg.Add(1)
        go func(src DataSource) {
            defer wg.Done()

            batch, err := dpa.extractFromSource(ctx, src)
            if err != nil {
                log.Printf("Failed to extract from %s: %v", src.URL, err)
                return
            }

            mu.Lock()
            allBatches = append(allBatches, batch)
            mu.Unlock()
        }(source)
    }

    wg.Wait()
    return allBatches, nil
}

func (dpa *DataPipelineAgent) extractFromSource(ctx context.Context, source DataSource) (DataBatch, error) {
    state := domain.NewState()
    
    extractPrompt := fmt.Sprintf(`Extract data from %s source:
    
    Source Details:
    - Type: %s
    - URL: %s
    - Format: %s
    
    Task:
    1. Fetch data from the source using appropriate method
    2. Parse the data according to format
    3. Extract structured records
    4. Return as consistent JSON array
    
    Handle errors gracefully and provide partial results if possible.`,
        source.Type, source.Type, source.URL, source.Format)

    state.Set("user_input", extractPrompt)
    state.Set("source_config", source)
    state.Set("batch_size", dpa.config.BatchSize)

    result, err := dpa.extractorAgent.Run(ctx, state)
    if err != nil {
        return DataBatch{}, err
    }

    // Process extraction result
    batch := DataBatch{
        SourceID:  source.URL,
        Timestamp: time.Now(),
        Records:   []map[string]interface{}{}, // Extract from result
        Metadata: map[string]interface{}{
            "source_type": source.Type,
            "format":      source.Format,
        },
    }

    if data, exists := result.Get("extracted_data"); exists {
        if records, ok := data.([]interface{}); ok {
            for _, record := range records {
                if recordMap, ok := record.(map[string]interface{}); ok {
                    batch.Records = append(batch.Records, recordMap)
                }
            }
        }
    }

    return batch, nil
}

func (dpa *DataPipelineAgent) transformData(ctx context.Context, batches []DataBatch) ([]DataBatch, error) {
    // Add data processing tools to transformer
    for _, tool := range dpa.dataTools {
        dpa.transformerAgent.AddTool(tool)
    }

    var transformedBatches []DataBatch

    for _, batch := range batches {
        state := domain.NewState()
        
        transformPrompt := fmt.Sprintf(`Transform data batch according to business rules:
        
        Input Records: %d
        Transform Rules: %v
        
        Apply these transformations:
        1. Data cleaning and normalization
        2. Field mapping and renaming
        3. Type conversions
        4. Business rule validation
        5. Enrichment with derived fields
        
        Maintain data lineage and report any issues.`,
            len(batch.Records), dpa.config.TransformRules)

        state.Set("user_input", transformPrompt)
        state.Set("input_batch", batch)
        state.Set("transform_rules", dpa.config.TransformRules)

        result, err := dpa.transformerAgent.Run(ctx, state)
        if err != nil {
            log.Printf("Transformation failed for batch %s: %v", batch.SourceID, err)
            continue
        }

        // Create transformed batch
        transformedBatch := DataBatch{
            SourceID:  batch.SourceID,
            Timestamp: time.Now(),
            Records:   []map[string]interface{}{}, // Extract from result
            Metadata: map[string]interface{}{
                "original_count": len(batch.Records),
                "transformed":    true,
            },
        }

        if data, exists := result.Get("transformed_data"); exists {
            if records, ok := data.([]interface{}); ok {
                for _, record := range records {
                    if recordMap, ok := record.(map[string]interface{}); ok {
                        transformedBatch.Records = append(transformedBatch.Records, recordMap)
                    }
                }
            }
        }

        transformedBatches = append(transformedBatches, transformedBatch)
    }

    return transformedBatches, nil
}

func (dpa *DataPipelineAgent) validateData(ctx context.Context, batches []DataBatch) ([]DataBatch, error) {
    // Add validation tools
    for _, tool := range dpa.dataTools {
        dpa.validatorAgent.AddTool(tool)
    }

    var validatedBatches []DataBatch

    for _, batch := range batches {
        state := domain.NewState()
        
        validatePrompt := fmt.Sprintf(`Validate data quality for batch:
        
        Records: %d
        Schema: %v
        Error Threshold: %.2f%%
        
        Validation Steps:
        1. Schema compliance checking
        2. Data type validation
        3. Business rule validation
        4. Completeness assessment
        5. Anomaly detection
        
        Flag invalid records and calculate quality score.`,
            len(batch.Records), dpa.config.ValidationSchema, dpa.config.ErrorThreshold*100)

        state.Set("user_input", validatePrompt)
        state.Set("input_batch", batch)
        state.Set("validation_schema", dpa.config.ValidationSchema)
        state.Set("error_threshold", dpa.config.ErrorThreshold)

        result, err := dpa.validatorAgent.Run(ctx, state)
        if err != nil {
            log.Printf("Validation failed for batch %s: %v", batch.SourceID, err)
            continue
        }

        // Check validation results
        if qualityScore, exists := result.Get("quality_score"); exists {
            if score, ok := qualityScore.(float64); ok && score < dpa.config.ErrorThreshold {
                log.Printf("Batch %s failed quality check (score: %.2f)", batch.SourceID, score)
                continue
            }
        }

        validatedBatch := DataBatch{
            SourceID:  batch.SourceID,
            Timestamp: time.Now(),
            Records:   batch.Records, // Valid records only
            Metadata: map[string]interface{}{
                "validated":      true,
                "quality_score": result.Get("quality_score"),
                "issues_found":  result.Get("validation_issues"),
            },
        }

        validatedBatches = append(validatedBatches, validatedBatch)
    }

    return validatedBatches, nil
}

func (dpa *DataPipelineAgent) loadData(ctx context.Context, batches []DataBatch) error {
    // Add loading tools
    for _, tool := range dpa.fileTools {
        dpa.loaderAgent.AddTool(tool)
    }
    for _, tool := range dpa.webTools {
        dpa.loaderAgent.AddTool(tool)
    }

    for _, batch := range batches {
        for _, target := range dpa.config.OutputTargets {
            err := dpa.loadToTarget(ctx, batch, target)
            if err != nil {
                log.Printf("Failed to load batch %s to target %s: %v", batch.SourceID, target.Path, err)
                continue
            }
        }
    }

    return nil
}

func (dpa *DataPipelineAgent) loadToTarget(ctx context.Context, batch DataBatch, target DataTarget) error {
    state := domain.NewState()
    
    loadPrompt := fmt.Sprintf(`Load data batch to target:
    
    Target Type: %s
    Target Path: %s
    Format: %s
    Records: %d
    Append Mode: %t
    
    Loading Steps:
    1. Format data according to target requirements
    2. Ensure data integrity during write
    3. Handle conflicts and duplicates
    4. Verify successful write
    5. Update metadata and logs
    
    Provide confirmation of successful load.`,
        target.Type, target.Path, target.Format, len(batch.Records), target.Append)

    state.Set("user_input", loadPrompt)
    state.Set("data_batch", batch)
    state.Set("target_config", target)

    result, err := dpa.loaderAgent.Run(ctx, state)
    if err != nil {
        return err
    }

    // Verify load success
    if success, exists := result.Get("load_success"); exists {
        if loaded, ok := success.(bool); !loaded {
            return fmt.Errorf("data load failed for target %s", target.Path)
        }
    }

    return nil
}

// Supporting types
type DataBatch struct {
    SourceID  string
    Timestamp time.Time
    Records   []map[string]interface{}
    Metadata  map[string]interface{}
}

func main() {
    fmt.Println("🔄 Data Pipeline Automation")
    fmt.Println("===========================")

    // Configure data pipeline
    config := &PipelineConfig{
        InputSources: []DataSource{
            {
                Type:   "api",
                URL:    "https://api.example.com/users",
                Format: "json",
                Headers: map[string]string{
                    "Authorization": "Bearer " + os.Getenv("API_TOKEN"),
                    "Content-Type":  "application/json",
                },
            },
            {
                Type:   "csv",
                URL:    "./data/customers.csv",
                Format: "csv",
            },
            {
                Type:   "rss",
                URL:    "https://example.com/news.xml",
                Format: "xml",
            },
        },
        OutputTargets: []DataTarget{
            {
                Type:   "file",
                Path:   "./output/processed_data.json",
                Format: "json",
                Append: false,
            },
            {
                Type:   "file",
                Path:   "./output/summary_report.csv",
                Format: "csv",
                Append: false,
            },
        },
        TransformRules: []TransformRule{
            {
                Field:     "email",
                Operation: "validate",
                Config:    map[string]interface{}{"pattern": "^[^@]+@[^@]+\\.[^@]+$"},
            },
            {
                Field:     "created_date",
                Operation: "map",
                Config:    map[string]interface{}{"from_format": "2006-01-02", "to_format": "01/02/2006"},
            },
        },
        ValidationSchema: map[string]interface{}{
            "type": "object",
            "required": []string{"id", "email", "name"},
            "properties": map[string]interface{}{
                "id":    map[string]interface{}{"type": "integer"},
                "email": map[string]interface{}{"type": "string"},
                "name":  map[string]interface{}{"type": "string"},
            },
        },
        BatchSize:      100,
        RetryAttempts:  3,
        ErrorThreshold: 0.95, // 95% quality required
    }

    // Create and execute pipeline
    pipeline, err := NewDataPipelineAgent(config)
    if err != nil {
        log.Fatalf("Failed to create pipeline: %v", err)
    }

    ctx := context.Background()
    err = pipeline.ExecutePipeline(ctx)
    if err != nil {
        log.Fatalf("Pipeline execution failed: %v", err)
    }

    fmt.Println("✅ Data pipeline completed successfully")
}
```

### Advanced Features
✅ **Multi-Source Extraction** - APIs, files, feeds processed in parallel  
✅ **Intelligent Transformation** - LLM-powered data cleaning and mapping  
✅ **Quality Validation** - Schema compliance and anomaly detection  
✅ **Flexible Loading** - Multiple output formats and targets  
✅ **Error Recovery** - Graceful handling of partial failures  

---

## Level 3: Workflow Orchestration
*Coordinate complex multi-agent automation*

### Enterprise Workflow System
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
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/file"
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/data"
)

// WorkflowOrchestrator manages complex multi-agent automation workflows
type WorkflowOrchestrator struct {
    workflows       map[string]domain.BaseAgent
    eventBus        *EventBus
    scheduler       *TaskScheduler
    monitor         *WorkflowMonitor
    stateManager    *StateManager
    
    config          *OrchestratorConfig
}

type OrchestratorConfig struct {
    MaxConcurrentWorkflows int
    DefaultTimeout         time.Duration
    RetryPolicy           *RetryPolicy
    MonitoringEnabled     bool
    EventBufferSize       int
}

type RetryPolicy struct {
    MaxAttempts     int
    BackoffStrategy string // "linear", "exponential", "constant"
    BaseDelay       time.Duration
    MaxDelay        time.Duration
    Jitter          bool
}

// EventBus handles workflow event communication
type EventBus struct {
    subscribers map[string][]chan domain.Event
    mutex       sync.RWMutex
}

// TaskScheduler handles workflow timing and dependencies
type TaskScheduler struct {
    scheduledTasks map[string]*ScheduledTask
    cronScheduler  *CronScheduler
    mutex          sync.RWMutex
}

type ScheduledTask struct {
    ID           string
    WorkflowName string
    Schedule     string // cron expression
    State        domain.StateReader
    LastRun      time.Time
    NextRun      time.Time
    Enabled      bool
}

// WorkflowMonitor tracks execution metrics and health
type WorkflowMonitor struct {
    metrics     map[string]*WorkflowMetrics
    alerts      chan Alert
    dashboards  []Dashboard
    mutex       sync.RWMutex
}

type WorkflowMetrics struct {
    TotalRuns        int64
    SuccessfulRuns   int64
    FailedRuns       int64
    AverageDuration  time.Duration
    LastRun          time.Time
    LastSuccess      time.Time
    LastFailure      time.Time
    ErrorRate        float64
}

type Alert struct {
    Level     string
    Message   string
    Workflow  string
    Timestamp time.Time
    Metadata  map[string]interface{}
}

// StateManager handles workflow state persistence and sharing
type StateManager struct {
    storage     map[string]*WorkflowState
    subscribers map[string][]chan *WorkflowState
    mutex       sync.RWMutex
}

type WorkflowState struct {
    ID           string
    WorkflowName string
    Status       string
    Data         map[string]interface{}
    StartTime    time.Time
    UpdateTime   time.Time
    Version      int
}

func NewWorkflowOrchestrator(config *OrchestratorConfig) *WorkflowOrchestrator {
    return &WorkflowOrchestrator{
        workflows:    make(map[string]domain.BaseAgent),
        eventBus:     NewEventBus(config.EventBufferSize),
        scheduler:    NewTaskScheduler(),
        monitor:      NewWorkflowMonitor(),
        stateManager: NewStateManager(),
        config:       config,
    }
}

// RegisterWorkflow adds a new workflow to the orchestrator
func (wo *WorkflowOrchestrator) RegisterWorkflow(name string, workflow domain.BaseAgent) error {
    fmt.Printf("📋 Registering workflow: %s\n", name)
    
    wo.workflows[name] = workflow
    wo.monitor.InitializeMetrics(name)
    
    // Set up event monitoring for the workflow
    wo.setupWorkflowMonitoring(name, workflow)
    
    return nil
}

func (wo *WorkflowOrchestrator) setupWorkflowMonitoring(name string, workflow domain.BaseAgent) {
    // Add event hooks to track workflow execution
    if hookable, ok := workflow.(interface {
        WithHook(func(domain.Event))
    }); ok {
        hookable.WithHook(func(event domain.Event) {
            wo.handleWorkflowEvent(name, event)
}
    }
}

func (wo *WorkflowOrchestrator) handleWorkflowEvent(workflowName string, event domain.Event) {
    // Update metrics based on event type
    wo.monitor.UpdateMetrics(workflowName, event)
    
    // Publish event to event bus
    wo.eventBus.Publish(workflowName, event)
    
    // Check for alert conditions
    if wo.shouldAlert(workflowName, event) {
        alert := wo.createAlert(workflowName, event)
        wo.monitor.SendAlert(alert)
    }
}

// CreateBusinessProcessWorkflow creates a comprehensive business process automation
func (wo *WorkflowOrchestrator) CreateBusinessProcessWorkflow() error {
    fmt.Println("🏢 Creating Business Process Workflow")
    
    // Create specialized agents for business process automation
    orderProcessor, err := wo.createOrderProcessingAgent()
    if err != nil {
        return err
    }
    
    inventoryManager, err := wo.createInventoryManagementAgent()
    if err != nil {
        return err
    }
    
    notificationAgent, err := wo.createNotificationAgent()
    if err != nil {
        return err
    }
    
    reportingAgent, err := wo.createReportingAgent()
    if err != nil {
        return err
    }

    // Create main business process workflow
    businessWorkflow := workflow.NewSequentialAgent("business-process").
        WithStopOnError(false).
        WithMaxRetries(3).
        WithTimeout(30 * time.Minute)

    // Phase 1: Order Processing (parallel for multiple orders)
    orderPhase := workflow.NewParallelAgent("order-processing").
        WithMaxConcurrency(5).
        WithMergeStrategy(workflow.MergeAll).
        AddAgent(orderProcessor)

    // Phase 2: Inventory Management
    inventoryPhase := workflow.NewConditionalAgent("inventory-management").
        AddAgent("low_stock", func(state *domain.State) bool {
            return wo.checkLowStockCondition(state)
        }, inventoryManager)

    // Phase 3: Customer Communication
    notificationPhase := workflow.NewParallelAgent("notifications").
        WithMergeStrategy(workflow.MergeAll).
        AddAgent(notificationAgent)

    // Phase 4: Business Reporting
    reportingPhase := workflow.NewSequentialAgent("reporting").
        AddAgent(reportingAgent)

    // Combine phases into complete workflow
    businessWorkflow.
        AddAgent(orderPhase).
        AddAgent(inventoryPhase).
        AddAgent(notificationPhase).
        AddAgent(reportingPhase)

    // Register the complete workflow
    return wo.RegisterWorkflow("business-process", businessWorkflow)
}

func (wo *WorkflowOrchestrator) createOrderProcessingAgent() (domain.BaseAgent, error) {
    agent, err := core.NewAgentFromString("order-processor", "anthropic/claude-3-5-sonnet")
    if err != nil {
        return nil, err
    }

    agent.SetSystemPrompt(`You are an order processing specialist.
    Handle incoming orders by:
    1. Validating order data and customer information
    2. Checking product availability and pricing
    3. Calculating taxes, shipping, and totals
    4. Processing payment authorizations
    5. Creating order confirmations and tracking
    
    Ensure accuracy and provide detailed status updates.`)

    // Add relevant tools
    agent.AddTool(data.NewJSONProcessTool())
    agent.AddTool(web.NewHTTPRequestTool())
    agent.AddTool(file.NewWriteFileTool())

    return agent, nil
}

func (wo *WorkflowOrchestrator) createInventoryManagementAgent() (domain.BaseAgent, error) {
    agent, err := core.NewAgentFromString("inventory-manager", "openai/gpt-4o-mini")
    if err != nil {
        return nil, err
    }

    agent.SetSystemPrompt(`You are an inventory management specialist.
    Monitor and manage inventory by:
    1. Tracking stock levels and product availability
    2. Identifying low stock and reorder points
    3. Managing supplier relationships and orders
    4. Optimizing inventory allocation
    5. Generating inventory reports and alerts
    
    Maintain optimal stock levels and prevent stockouts.`)

    agent.AddTool(data.NewCSVProcessTool())
    agent.AddTool(web.NewHTTPRequestTool())
    agent.AddTool(file.NewReadFileTool())

    return agent, nil
}

func (wo *WorkflowOrchestrator) createNotificationAgent() (domain.BaseAgent, error) {
    agent, err := core.NewAgentFromString("notification-sender", "gemini/gemini-2.0-flash")
    if err != nil {
        return nil, err
    }

    agent.SetSystemPrompt(`You are a customer communication specialist.
    Send targeted notifications by:
    1. Crafting personalized messages for different customer segments
    2. Choosing optimal communication channels (email, SMS, push)
    3. Timing delivery for maximum engagement
    4. Tracking delivery and engagement metrics
    5. Managing subscription preferences and opt-outs
    
    Ensure professional, helpful, and timely communications.`)

    agent.AddTool(web.NewHTTPRequestTool())
    agent.AddTool(data.NewJSONProcessTool())

    return agent, nil
}

func (wo *WorkflowOrchestrator) createReportingAgent() (domain.BaseAgent, error) {
    agent, err := core.NewAgentFromString("business-reporter", "anthropic/claude-3-5-sonnet")
    if err != nil {
        return nil, err
    }

    agent.SetSystemPrompt(`You are a business intelligence specialist.
    Generate comprehensive reports by:
    1. Collecting data from multiple business systems
    2. Analyzing trends, patterns, and performance metrics
    3. Creating visualizations and executive summaries
    4. Identifying insights and actionable recommendations
    5. Formatting reports for different stakeholder audiences
    
    Provide clear, accurate, and actionable business intelligence.`)

    agent.AddTool(data.NewJSONProcessTool())
    agent.AddTool(data.NewCSVProcessTool())
    agent.AddTool(file.NewWriteFileTool())

    return agent, nil
}

// ExecuteWorkflow runs a registered workflow with full monitoring
func (wo *WorkflowOrchestrator) ExecuteWorkflow(ctx context.Context, name string, initialState domain.StateReader) (*domain.State, error) {
    workflow, exists := wo.workflows[name]
    if !exists {
        return nil, fmt.Errorf("workflow '%s' not found", name)
    }

    fmt.Printf("🚀 Executing workflow: %s\n", name)
    
    // Create workflow state
    workflowState := &WorkflowState{
        ID:           fmt.Sprintf("%s-%d", name, time.Now().Unix()),
        WorkflowName: name,
        Status:       "running",
        Data:         make(map[string]interface{}),
        StartTime:    time.Now(),
        Version:      1,
    }

    // Store initial state
    wo.stateManager.SaveState(workflowState)

    // Execute with timeout and monitoring
    ctx, cancel := context.WithTimeout(ctx, wo.config.DefaultTimeout)
    defer cancel()

    startTime := time.Now()
    result, err := workflow.Run(ctx, initialState)
    duration := time.Since(startTime)

    // Update workflow state
    workflowState.Status = "completed"
    if err != nil {
        workflowState.Status = "failed"
    }
    workflowState.UpdateTime = time.Now()
    workflowState.Version++

    // Update metrics
    wo.monitor.RecordExecution(name, duration, err == nil)

    // Store final state
    wo.stateManager.SaveState(workflowState)

    if err != nil {
        fmt.Printf("❌ Workflow '%s' failed: %v\n", name, err)
        return nil, err
    }

    fmt.Printf("✅ Workflow '%s' completed in %v\n", name, duration)
    return result, nil
}

// ScheduleWorkflow adds a workflow to the scheduler
func (wo *WorkflowOrchestrator) ScheduleWorkflow(name, schedule string, state domain.StateReader) error {
    if _, exists := wo.workflows[name]; !exists {
        return fmt.Errorf("workflow '%s' not registered", name)
    }

    task := &ScheduledTask{
        ID:           fmt.Sprintf("%s-scheduled", name),
        WorkflowName: name,
        Schedule:     schedule,
        State:        state,
        LastRun:      time.Time{},
        NextRun:      wo.calculateNextRun(schedule),
        Enabled:      true,
    }

    return wo.scheduler.AddTask(task)
}

// StartScheduler begins processing scheduled workflows
func (wo *WorkflowOrchestrator) StartScheduler(ctx context.Context) {
    fmt.Println("⏰ Starting workflow scheduler")
    
    ticker := time.NewTicker(1 * time.Minute) // Check every minute
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            wo.processScheduledTasks(ctx)
        case <-ctx.Done():
            fmt.Println("🛑 Scheduler stopped")
            return
        }
    }
}

func (wo *WorkflowOrchestrator) processScheduledTasks(ctx context.Context) {
    now := time.Now()
    tasks := wo.scheduler.GetDueTasks(now)

    for _, task := range tasks {
        if !task.Enabled {
            continue
        }

        fmt.Printf("⏰ Executing scheduled workflow: %s\n", task.WorkflowName)
        
        go func(t *ScheduledTask) {
            _, err := wo.ExecuteWorkflow(ctx, t.WorkflowName, t.State)
            if err != nil {
                log.Printf("Scheduled workflow failed: %s - %v", t.WorkflowName, err)
            }

            // Update task schedule
            t.LastRun = now
            t.NextRun = wo.calculateNextRun(t.Schedule)
            wo.scheduler.UpdateTask(t)
        }(task)
    }
}

// Helper methods and supporting implementations
func (wo *WorkflowOrchestrator) checkLowStockCondition(state *domain.State) bool {
    // Check if any products are below reorder point
    if inventory, exists := state.Get("inventory_levels"); exists {
        if levels, ok := inventory.(map[string]interface{}); ok {
            for _, level := range levels {
                if stock, ok := level.(float64); ok && stock < 10 {
                    return true
                }
            }
        }
    }
    return false
}

func (wo *WorkflowOrchestrator) shouldAlert(workflowName string, event domain.Event) bool {
    return event.Type == domain.EventAgentError || 
           (event.Type == domain.EventAgentEnd && wo.monitor.GetErrorRate(workflowName) > 0.1)
}

func (wo *WorkflowOrchestrator) createAlert(workflowName string, event domain.Event) Alert {
    return Alert{
        Level:     "high",
        Message:   fmt.Sprintf("Workflow %s encountered issue: %v", workflowName, event.Data),
        Workflow:  workflowName,
        Timestamp: time.Now(),
        Metadata:  event.Data.(map[string]interface{}),
    }
}

func (wo *WorkflowOrchestrator) calculateNextRun(schedule string) time.Time {
    // Simplified cron parsing - in production, use proper cron library
    return time.Now().Add(1 * time.Hour)
}

// Supporting type implementations (simplified)
func NewEventBus(bufferSize int) *EventBus {
    return &EventBus{
        subscribers: make(map[string][]chan domain.Event),
    }
}

func (eb *EventBus) Publish(topic string, event domain.Event) {
    eb.mutex.RLock()
    defer eb.mutex.RUnlock()
    
    if subscribers, exists := eb.subscribers[topic]; exists {
        for _, ch := range subscribers {
            select {
            case ch <- event:
            default: // Don't block if channel is full
            }
        }
    }
}

func NewTaskScheduler() *TaskScheduler {
    return &TaskScheduler{
        scheduledTasks: make(map[string]*ScheduledTask),
    }
}

func (ts *TaskScheduler) AddTask(task *ScheduledTask) error {
    ts.mutex.Lock()
    defer ts.mutex.Unlock()
    
    ts.scheduledTasks[task.ID] = task
    return nil
}

func (ts *TaskScheduler) GetDueTasks(now time.Time) []*ScheduledTask {
    ts.mutex.RLock()
    defer ts.mutex.RUnlock()
    
    var dueTasks []*ScheduledTask
    for _, task := range ts.scheduledTasks {
        if task.NextRun.Before(now) || task.NextRun.Equal(now) {
            dueTasks = append(dueTasks, task)
        }
    }
    return dueTasks
}

func (ts *TaskScheduler) UpdateTask(task *ScheduledTask) {
    ts.mutex.Lock()
    defer ts.mutex.Unlock()
    
    ts.scheduledTasks[task.ID] = task
}

func NewWorkflowMonitor() *WorkflowMonitor {
    return &WorkflowMonitor{
        metrics: make(map[string]*WorkflowMetrics),
        alerts:  make(chan Alert, 100),
    }
}

func (wm *WorkflowMonitor) InitializeMetrics(workflowName string) {
    wm.mutex.Lock()
    defer wm.mutex.Unlock()
    
    wm.metrics[workflowName] = &WorkflowMetrics{}
}

func (wm *WorkflowMonitor) UpdateMetrics(workflowName string, event domain.Event) {
    wm.mutex.Lock()
    defer wm.mutex.Unlock()
    
    if metrics, exists := wm.metrics[workflowName]; exists {
        switch event.Type {
        case domain.EventAgentError:
            metrics.FailedRuns++
        case domain.EventAgentEnd:
            metrics.SuccessfulRuns++
        }
        
        metrics.TotalRuns = metrics.SuccessfulRuns + metrics.FailedRuns
        if metrics.TotalRuns > 0 {
            metrics.ErrorRate = float64(metrics.FailedRuns) / float64(metrics.TotalRuns)
        }
    }
}

func (wm *WorkflowMonitor) RecordExecution(workflowName string, duration time.Duration, success bool) {
    wm.mutex.Lock()
    defer wm.mutex.Unlock()
    
    if metrics, exists := wm.metrics[workflowName]; exists {
        metrics.LastRun = time.Now()
        metrics.AverageDuration = (metrics.AverageDuration + duration) / 2
        
        if success {
            metrics.LastSuccess = time.Now()
        } else {
            metrics.LastFailure = time.Now()
        }
    }
}

func (wm *WorkflowMonitor) GetErrorRate(workflowName string) float64 {
    wm.mutex.RLock()
    defer wm.mutex.RUnlock()
    
    if metrics, exists := wm.metrics[workflowName]; exists {
        return metrics.ErrorRate
    }
    return 0
}

func (wm *WorkflowMonitor) SendAlert(alert Alert) {
    select {
    case wm.alerts <- alert:
    default:
        log.Printf("Alert buffer full, dropping alert: %s", alert.Message)
    }
}

func NewStateManager() *StateManager {
    return &StateManager{
        storage:     make(map[string]*WorkflowState),
        subscribers: make(map[string][]chan *WorkflowState),
    }
}

func (sm *StateManager) SaveState(state *WorkflowState) {
    sm.mutex.Lock()
    defer sm.mutex.Unlock()
    
    sm.storage[state.ID] = state
}

func main() {
    fmt.Println("🏢 Enterprise Workflow Orchestration")
    fmt.Println("===================================")

    // Create orchestrator configuration
    config := &OrchestratorConfig{
        MaxConcurrentWorkflows: 10,
        DefaultTimeout:         30 * time.Minute,
        RetryPolicy: &RetryPolicy{
            MaxAttempts:     3,
            BackoffStrategy: "exponential",
            BaseDelay:       1 * time.Second,
            MaxDelay:        30 * time.Second,
            Jitter:          true,
        },
        MonitoringEnabled: true,
        EventBufferSize:   1000,
    }

    // Create orchestrator
    orchestrator := NewWorkflowOrchestrator(config)

    // Create and register business process workflow
    err := orchestrator.CreateBusinessProcessWorkflow()
    if err != nil {
        log.Fatalf("Failed to create business workflow: %v", err)
    }

    // Schedule the workflow to run every hour
    initialState := domain.NewState()
    initialState.Set("business_hours", true)
    initialState.Set("max_orders_per_batch", 50)

    err = orchestrator.ScheduleWorkflow("business-process", "0 * * * *", initialState) // Every hour
    if err != nil {
        log.Fatalf("Failed to schedule workflow: %v", err)
    }

    // Execute workflow immediately for demonstration
    ctx := context.Background()
    result, err := orchestrator.ExecuteWorkflow(ctx, "business-process", initialState)
    if err != nil {
        log.Fatalf("Workflow execution failed: %v", err)
    }

    fmt.Printf("✅ Business process completed successfully\n")
    if summary, exists := result.Get("execution_summary"); exists {
        fmt.Printf("📊 Summary: %v\n", summary)
    }

    // Start scheduler (would run indefinitely in production)
    fmt.Println("⏰ Starting scheduler for 1 minute...")
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
    defer cancel()

    orchestrator.StartScheduler(ctx)
}
```

### Orchestration Features
✅ **Multi-Agent Coordination** - Sequential, parallel, and conditional workflows  
✅ **Event-Driven Architecture** - Real-time monitoring and communication  
✅ **Task Scheduling** - Cron-like scheduling with dependency management  
✅ **State Management** - Persistent workflow state across executions  
✅ **Performance Monitoring** - Metrics, alerts, and health tracking  
✅ **Error Recovery** - Intelligent retry with exponential backoff  

---

## Level 4: Production Automation Platform
*Enterprise-scale automation with monitoring and governance*

### Production Automation Infrastructure
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
)

// ProductionAutomationPlatform provides enterprise automation capabilities
type ProductionAutomationPlatform struct {
    orchestrator    *WorkflowOrchestrator
    auditLogger     *AuditLogger
    security        *SecurityManager
    governance      *GovernanceEngine
    compliance      *ComplianceChecker
    performance     *PerformanceOptimizer
    
    config          *PlatformConfig
}

type PlatformConfig struct {
    SecurityEnabled     bool
    AuditingEnabled     bool
    ComplianceEnabled   bool
    PerformanceOptimization bool
    MaxWorkflowRuntime  time.Duration
    ResourceLimits      *ResourceLimits
    NotificationSettings *NotificationSettings
}

type ResourceLimits struct {
    MaxCPUPercent    float64
    MaxMemoryMB      int64
    MaxDiskSpaceMB   int64
    MaxNetworkMbps   int64
    MaxConcurrentJobs int
}

// AuditLogger tracks all automation activities for compliance
type AuditLogger struct {
    entries     []AuditEntry
    storage     AuditStorage
    encryption  EncryptionService
    mutex       sync.RWMutex
}

type AuditEntry struct {
    ID          string
    Timestamp   time.Time
    WorkflowID  string
    Action      string
    User        string
    Details     map[string]interface{}
    Sensitive   bool
    Encrypted   bool
}

// SecurityManager handles authentication, authorization, and secure execution
type SecurityManager struct {
    rbac            *RoleBasedAccessControl
    encryption      EncryptionService
    secretsVault    SecretsVault
    sandboxManager  SandboxManager
}

// GovernanceEngine enforces automation policies and standards
type GovernanceEngine struct {
    policies        []AutomationPolicy
    approvalWorkflows map[string]ApprovalWorkflow
    riskAssessment  RiskAssessmentEngine
}

type AutomationPolicy struct {
    ID           string
    Name         string
    Description  string
    Rules        []PolicyRule
    Enforcement  string // "advisory", "warning", "blocking"
    Scope        []string
}

type PolicyRule struct {
    Type        string // "resource_limit", "execution_time", "approval_required"
    Condition   string
    Action      string
    Parameters  map[string]interface{}
}

// ComplianceChecker ensures automation meets regulatory requirements
type ComplianceChecker struct {
    frameworks   []ComplianceFramework
    scanners     []ComplianceScanner
    reports      []ComplianceReport
}

type ComplianceFramework struct {
    Name         string // "SOX", "GDPR", "HIPAA", "PCI-DSS"
    Requirements []Requirement
    CheckFrequency time.Duration
}

// PerformanceOptimizer monitors and improves automation performance
type PerformanceOptimizer struct {
    metrics      PerformanceMetrics
    optimizer    OptimizationEngine
    profiler     AutomationProfiler
    tuner        ParameterTuner
}

func NewProductionAutomationPlatform(config *PlatformConfig) *ProductionAutomationPlatform {
    return &ProductionAutomationPlatform{
        orchestrator: NewWorkflowOrchestrator(&OrchestratorConfig{
            MaxConcurrentWorkflows: config.ResourceLimits.MaxConcurrentJobs,
            DefaultTimeout:         config.MaxWorkflowRuntime,
            MonitoringEnabled:      true,
        }),
        auditLogger:     NewAuditLogger(config.AuditingEnabled),
        security:        NewSecurityManager(config.SecurityEnabled),
        governance:      NewGovernanceEngine(),
        compliance:      NewComplianceChecker(config.ComplianceEnabled),
        performance:     NewPerformanceOptimizer(config.PerformanceOptimization),
        config:          config,
    }
}

// DeployAutomationSuite creates a comprehensive automation environment
func (pap *ProductionAutomationPlatform) DeployAutomationSuite(ctx context.Context) error {
    fmt.Println("🏭 Deploying Production Automation Suite")
    fmt.Println("========================================")

    // Initialize security infrastructure
    err := pap.initializeSecurity(ctx)
    if err != nil {
        return fmt.Errorf("security initialization failed: %w", err)
    }

    // Set up governance policies
    err = pap.setupGovernancePolicies()
    if err != nil {
        return fmt.Errorf("governance setup failed: %w", err)
    }

    // Configure compliance monitoring
    err = pap.configureComplianceMonitoring()
    if err != nil {
        return fmt.Errorf("compliance configuration failed: %w", err)
    }

    // Deploy core automation workflows
    err = pap.deployCore workflows(ctx)
    if err != nil {
        return fmt.Errorf("core workflow deployment failed: %w", err)
    }

    // Start monitoring and optimization services
    go pap.startMonitoringServices(ctx)
    
    fmt.Println("✅ Production automation suite deployed successfully")
    return nil
}

func (pap *ProductionAutomationPlatform) initializeSecurity(ctx context.Context) error {
    fmt.Println("🔒 Initializing security infrastructure")

    // Set up role-based access control
    err := pap.security.InitializeRBAC()
    if err != nil {
        return err
    }

    // Configure secrets management
    err = pap.security.SetupSecretsVault()
    if err != nil {
        return err
    }

    // Initialize sandbox environment for safe execution
    err = pap.security.CreateSandboxEnvironment()
    if err != nil {
        return err
    }

    return nil
}

func (pap *ProductionAutomationPlatform) setupGovernancePolicies() error {
    fmt.Println("📋 Setting up governance policies")

    // Resource limit policy
    resourcePolicy := AutomationPolicy{
        ID:          "resource-limits",
        Name:        "Resource Usage Limits",
        Description: "Enforce CPU, memory, and execution time limits",
        Rules: []PolicyRule{
            {
                Type:       "resource_limit",
                Condition:  "cpu_usage > 80",
                Action:     "throttle",
                Parameters: map[string]interface{}{"max_cpu": 80},
            },
            {
                Type:       "execution_time",
                Condition:  "runtime > 30m",
                Action:     "terminate",
                Parameters: map[string]interface{}{"max_time": "30m"},
            },
        },
        Enforcement: "blocking",
        Scope:       []string{"*"},
    }

    // High-risk operation policy
    approvalPolicy := AutomationPolicy{
        ID:          "high-risk-approval",
        Name:        "High Risk Operation Approval",
        Description: "Require approval for high-risk automation tasks",
        Rules: []PolicyRule{
            {
                Type:       "approval_required",
                Condition:  "risk_score > 7",
                Action:     "require_approval",
                Parameters: map[string]interface{}{"approver_role": "automation_admin"},
            },
        },
        Enforcement: "blocking",
        Scope:       []string{"system", "finance", "security"},
    }

    pap.governance.AddPolicy(resourcePolicy)
    pap.governance.AddPolicy(approvalPolicy)

    return nil
}

func (pap *ProductionAutomationPlatform) configureComplianceMonitoring() error {
    fmt.Println("📊 Configuring compliance monitoring")

    // SOX compliance for financial automation
    soxFramework := ComplianceFramework{
        Name: "SOX",
        Requirements: []Requirement{
            {
                ID:          "sox-audit-trail",
                Description: "Maintain complete audit trail of financial automation",
                CheckType:   "audit_completeness",
                Severity:    "high",
            },
            {
                ID:          "sox-segregation",
                Description: "Enforce segregation of duties in financial processes",
                CheckType:   "access_control",
                Severity:    "critical",
            },
        },
        CheckFrequency: 24 * time.Hour,
    }

    // GDPR compliance for data processing automation
    gdprFramework := ComplianceFramework{
        Name: "GDPR",
        Requirements: []Requirement{
            {
                ID:          "gdpr-consent",
                Description: "Verify consent for automated personal data processing",
                CheckType:   "consent_validation",
                Severity:    "high",
            },
            {
                ID:          "gdpr-retention",
                Description: "Enforce data retention limits in automation",
                CheckType:   "data_retention",
                Severity:    "medium",
            },
        },
        CheckFrequency: 8 * time.Hour,
    }

    pap.compliance.AddFramework(soxFramework)
    pap.compliance.AddFramework(gdprFramework)

    return nil
}

func (pap *ProductionAutomationPlatform) deployCoreWorkflows(ctx context.Context) error {
    fmt.Println("⚙️ Deploying core automation workflows")

    // Infrastructure monitoring workflow
    infraWorkflow, err := pap.createInfrastructureMonitoringWorkflow()
    if err != nil {
        return err
    }
    pap.orchestrator.RegisterWorkflow("infrastructure-monitoring", infraWorkflow)

    // Security incident response workflow
    securityWorkflow, err := pap.createSecurityIncidentWorkflow()
    if err != nil {
        return err
    }
    pap.orchestrator.RegisterWorkflow("security-incident-response", securityWorkflow)

    // Backup and disaster recovery workflow
    backupWorkflow, err := pap.createBackupRecoveryWorkflow()
    if err != nil {
        return err
    }
    pap.orchestrator.RegisterWorkflow("backup-recovery", backupWorkflow)

    // Compliance audit workflow
    complianceWorkflow, err := pap.createComplianceAuditWorkflow()
    if err != nil {
        return err
    }
    pap.orchestrator.RegisterWorkflow("compliance-audit", complianceWorkflow)

    return nil
}

func (pap *ProductionAutomationPlatform) createInfrastructureMonitoringWorkflow() (domain.BaseAgent, error) {
    monitoringAgent, err := core.NewAgentFromString("infra-monitor", "anthropic/claude-3-5-sonnet")
    if err != nil {
        return nil, err
    }

    monitoringAgent.SetSystemPrompt(`You are an infrastructure monitoring specialist for production automation.
    
    Monitor and maintain:
    1. System health and performance metrics
    2. Application availability and response times
    3. Resource utilization and capacity planning
    4. Network connectivity and security
    5. Automated remediation for common issues
    
    Priorities:
    - Prevent service disruptions
    - Maintain optimal performance
    - Ensure security compliance
    - Provide actionable insights
    
    Always follow governance policies and maintain audit trails.`)

    return monitoringAgent, nil
}

func (pap *ProductionAutomationPlatform) createSecurityIncidentWorkflow() (domain.BaseAgent, error) {
    securityAgent, err := core.NewAgentFromString("security-responder", "openai/gpt-4o")
    if err != nil {
        return nil, err
    }

    securityAgent.SetSystemPrompt(`You are a security incident response automation specialist.
    
    Handle security incidents by:
    1. Detecting and classifying security threats
    2. Implementing immediate containment measures
    3. Collecting forensic evidence
    4. Notifying appropriate stakeholders
    5. Coordinating remediation efforts
    
    Critical requirements:
    - Respond within SLA timeframes
    - Preserve evidence integrity
    - Follow incident response procedures
    - Maintain communication protocols
    
    Escalate critical incidents immediately and maintain detailed logs.`)

    return securityAgent, nil
}

func (pap *ProductionAutomationPlatform) createBackupRecoveryWorkflow() (domain.BaseAgent, error) {
    backupAgent, err := core.NewAgentFromString("backup-manager", "gemini/gemini-2.0-flash")
    if err != nil {
        return nil, err
    }

    backupAgent.SetSystemPrompt(`You are a backup and disaster recovery automation specialist.
    
    Manage data protection by:
    1. Executing scheduled backup operations
    2. Verifying backup integrity and completeness
    3. Managing retention policies
    4. Testing recovery procedures
    5. Coordinating disaster recovery scenarios
    
    Ensure:
    - Data integrity and consistency
    - Recovery time objectives (RTO)
    - Recovery point objectives (RPO)
    - Compliance with retention policies
    
    Test recovery procedures regularly and maintain recovery documentation.`)

    return backupAgent, nil
}

func (pap *ProductionAutomationPlatform) createComplianceAuditWorkflow() (domain.BaseAgent, error) {
    auditAgent, err := core.NewAgentFromString("compliance-auditor", "anthropic/claude-3-5-sonnet")
    if err != nil {
        return nil, err
    }

    auditAgent.SetSystemPrompt(`You are a compliance audit automation specialist.
    
    Conduct automated compliance assessments by:
    1. Reviewing automation workflows against policies
    2. Validating security controls and access permissions
    3. Checking audit trail completeness
    4. Assessing risk and control effectiveness
    5. Generating compliance reports and recommendations
    
    Focus areas:
    - Regulatory compliance (SOX, GDPR, etc.)
    - Internal policy adherence
    - Security control validation
    - Audit trail integrity
    
    Provide clear findings with risk ratings and remediation recommendations.`)

    return auditAgent, nil
}

func (pap *ProductionAutomationPlatform) startMonitoringServices(ctx context.Context) {
    fmt.Println("📊 Starting monitoring and optimization services")

    // Start performance monitoring
    go pap.performance.StartPerformanceMonitoring(ctx)

    // Start compliance checking
    go pap.compliance.StartContinuousMonitoring(ctx)

    // Start audit logging
    go pap.auditLogger.StartAuditProcessing(ctx)

    // Start governance policy enforcement
    go pap.governance.StartPolicyEnforcement(ctx)
}

// ExecuteSecureAutomation runs automation with full security and governance
func (pap *ProductionAutomationPlatform) ExecuteSecureAutomation(ctx context.Context, workflowName string, request AutomationRequest) (*AutomationResult, error) {
    fmt.Printf("🔐 Executing secure automation: %s\n", workflowName)

    // Audit the request
    auditEntry := AuditEntry{
        ID:         fmt.Sprintf("exec-%d", time.Now().Unix()),
        Timestamp:  time.Now(),
        WorkflowID: workflowName,
        Action:     "execute_automation",
        User:       request.UserID,
        Details: map[string]interface{}{
            "request": request,
        },
        Sensitive: pap.security.IsSensitiveWorkflow(workflowName),
    }
    pap.auditLogger.LogEntry(auditEntry)

    // Security checks
    authorized, err := pap.security.AuthorizeExecution(request.UserID, workflowName, request.Permissions)
    if err != nil {
        return nil, fmt.Errorf("authorization check failed: %w", err)
    }
    if !authorized {
        return nil, fmt.Errorf("user %s not authorized for workflow %s", request.UserID, workflowName)
    }

    // Governance policy evaluation
    policyResult, err := pap.governance.EvaluatePolicies(workflowName, request)
    if err != nil {
        return nil, fmt.Errorf("policy evaluation failed: %w", err)
    }
    if policyResult.RequiresApproval {
        return nil, fmt.Errorf("workflow requires approval: %s", policyResult.Reason)
    }

    // Performance resource allocation
    resources, err := pap.performance.AllocateResources(workflowName, request.ResourceRequirements)
    if err != nil {
        return nil, fmt.Errorf("resource allocation failed: %w", err)
    }
    defer pap.performance.ReleaseResources(resources)

    // Execute in secure sandbox
    result, err := pap.security.ExecuteInSandbox(ctx, func() (*domain.State, error) {
        return pap.orchestrator.ExecuteWorkflow(ctx, workflowName, request.State)
}

    // Log execution result
    resultAudit := AuditEntry{
        ID:         fmt.Sprintf("result-%d", time.Now().Unix()),
        Timestamp:  time.Now(),
        WorkflowID: workflowName,
        Action:     "automation_result",
        User:       request.UserID,
        Details: map[string]interface{}{
            "success": err == nil,
            "error":   err,
        },
    }
    pap.auditLogger.LogEntry(resultAudit)

    if err != nil {
        return nil, fmt.Errorf("automation execution failed: %w", err)
    }

    return &AutomationResult{
        WorkflowID:    workflowName,
        ExecutionID:   auditEntry.ID,
        Status:        "completed",
        Result:        result,
        ExecutionTime: time.Since(auditEntry.Timestamp),
        ResourcesUsed: resources,
        AuditTrail:    []string{auditEntry.ID, resultAudit.ID},
    }, nil
}

// Supporting types and simplified implementations
type AutomationRequest struct {
    UserID               string
    Permissions          []string
    State                domain.StateReader
    ResourceRequirements *ResourceLimits
    Priority             string
}

type AutomationResult struct {
    WorkflowID    string
    ExecutionID   string
    Status        string
    Result        *domain.State
    ExecutionTime time.Duration
    ResourcesUsed *ResourceLimits
    AuditTrail    []string
}

type Requirement struct {
    ID          string
    Description string
    CheckType   string
    Severity    string
}

// Simplified implementation stubs
func NewAuditLogger(enabled bool) *AuditLogger { return &AuditLogger{} }
func NewSecurityManager(enabled bool) *SecurityManager { return &SecurityManager{} }
func NewGovernanceEngine() *GovernanceEngine { return &GovernanceEngine{} }
func NewComplianceChecker(enabled bool) *ComplianceChecker { return &ComplianceChecker{} }
func NewPerformanceOptimizer(enabled bool) *PerformanceOptimizer { return &PerformanceOptimizer{} }

func (al *AuditLogger) LogEntry(entry AuditEntry) {}
func (sm *SecurityManager) InitializeRBAC() error { return nil }
func (sm *SecurityManager) SetupSecretsVault() error { return nil }
func (sm *SecurityManager) CreateSandboxEnvironment() error { return nil }
func (sm *SecurityManager) IsSensitiveWorkflow(name string) bool { return true }
func (sm *SecurityManager) AuthorizeExecution(userID, workflow string, perms []string) (bool, error) { return true, nil }
func (sm *SecurityManager) ExecuteInSandbox(ctx context.Context, fn func() (*domain.State, error)) (*domain.State, error) { return fn() }

func (ge *GovernanceEngine) AddPolicy(policy AutomationPolicy) {}
func (ge *GovernanceEngine) EvaluatePolicies(workflow string, request AutomationRequest) (*PolicyResult, error) {
    return &PolicyResult{RequiresApproval: false}, nil
}
func (ge *GovernanceEngine) StartPolicyEnforcement(ctx context.Context) {}

func (cc *ComplianceChecker) AddFramework(framework ComplianceFramework) {}
func (cc *ComplianceChecker) StartContinuousMonitoring(ctx context.Context) {}

func (po *PerformanceOptimizer) AllocateResources(workflow string, req *ResourceLimits) (*ResourceLimits, error) {
    return req, nil
}
func (po *PerformanceOptimizer) ReleaseResources(resources *ResourceLimits) {}
func (po *PerformanceOptimizer) StartPerformanceMonitoring(ctx context.Context) {}

func (al *AuditLogger) StartAuditProcessing(ctx context.Context) {}

type PolicyResult struct {
    RequiresApproval bool
    Reason          string
}

func main() {
    fmt.Println("🏭 Production Automation Platform")
    fmt.Println("=================================")

    // Create production configuration
    config := &PlatformConfig{
        SecurityEnabled:   true,
        AuditingEnabled:   true,
        ComplianceEnabled: true,
        PerformanceOptimization: true,
        MaxWorkflowRuntime: 2 * time.Hour,
        ResourceLimits: &ResourceLimits{
            MaxCPUPercent:     80,
            MaxMemoryMB:       8192,
            MaxDiskSpaceMB:    10240,
            MaxNetworkMbps:    100,
            MaxConcurrentJobs: 20,
        },
    }

    // Initialize production platform
    platform := NewProductionAutomationPlatform(config)

    // Deploy automation suite
    ctx := context.Background()
    err := platform.DeployAutomationSuite(ctx)
    if err != nil {
        log.Fatalf("Platform deployment failed: %v", err)
    }

    // Execute secure automation
    request := AutomationRequest{
        UserID:      "admin@company.com",
        Permissions: []string{"automation_execute", "infrastructure_monitor"},
        State:       domain.NewState(),
        ResourceRequirements: &ResourceLimits{
            MaxCPUPercent:     50,
            MaxMemoryMB:       2048,
            MaxConcurrentJobs: 5,
        },
        Priority: "high",
    }

    result, err := platform.ExecuteSecureAutomation(ctx, "infrastructure-monitoring", request)
    if err != nil {
        log.Fatalf("Secure automation failed: %v", err)
    }

    fmt.Printf("✅ Production automation completed\n")
    fmt.Printf("Execution ID: %s\n", result.ExecutionID)
    fmt.Printf("Duration: %v\n", result.ExecutionTime)
    fmt.Printf("Status: %s\n", result.Status)
}
```

### Production Features
✅ **Enterprise Security** - RBAC, secrets vault, sandboxed execution  
✅ **Comprehensive Auditing** - Full audit trails for compliance  
✅ **Governance Policies** - Automated policy enforcement and approval workflows  
✅ **Compliance Monitoring** - SOX, GDPR, HIPAA compliance frameworks  
✅ **Performance Optimization** - Resource management and monitoring  
✅ **Risk Management** - Risk assessment and mitigation strategies  

---

## Best Practices

### Automation Agent Design Patterns

1. **Security First**
   - Validate all inputs
   - Use least privilege access
   - Audit all operations
   - Encrypt sensitive data

2. **Reliability Engineering**
   - Implement circuit breakers
   - Use graceful degradation
   - Plan for partial failures
   - Monitor everything

3. **Scalability Design**
   - Use horizontal scaling
   - Implement rate limiting
   - Cache expensive operations
   - Design for concurrency

4. **Operational Excellence**
   - Comprehensive logging
   - Performance monitoring
   - Error alerting
   - Recovery procedures

## Next Steps

🤖 **Automation mastered!** Continue with:

- **[Provider Selection](provider-selection.md)** - Choose optimal providers for automation
- **[Agent Communication](agent-communication.md)** - Multi-agent coordination patterns
- **[Data Validation](data-validation.md)** - Ensure data quality in automation
- **[Production Deployment](../advanced/production-deployment.md)** - Deploy at scale

### Quick Reference

- **[Built-in Tools Reference](../reference/built-in-tools-reference.md)** - Complete automation tool catalog
- **[Configuration Reference](../reference/configuration-reference.md)** - All configuration options
- **[Best Practices Checklist](../reference/best-practices-checklist.md)** - Production checklist

---

**Need help?** Check our [troubleshooting guide](../advanced/troubleshooting.md) or join the discussion on [GitHub](https://github.com/lexlapax/go-llms/discussions).