# Creating Custom Tools: Build Custom Tools

> **[Project Root](/) / [Documentation](../..) / [Technical Documentation](../../technical) / [Tools](../../technical/tools) / Creating Tools**

Complete guide to designing and implementing custom tools in Go-LLMs, covering tool interfaces, schema definition, validation, testing, packaging, and deployment strategies for building robust, reusable tools that integrate seamlessly with the agent ecosystem.

## Tool Development Lifecycle

### 1. Tool Design Phase

```go
// ToolDesign represents the design specification for a new tool
type ToolDesign struct {
    Name        string             `yaml:"name" json:"name"`
    Category    string             `yaml:"category" json:"category"`
    Description string             `yaml:"description" json:"description"`
    Purpose     string             `yaml:"purpose" json:"purpose"`
    UseCases    []string           `yaml:"use_cases" json:"use_cases"`
    
    // Interface design
    InputSchema  *SchemaDefinition `yaml:"input_schema" json:"input_schema"`
    OutputSchema *SchemaDefinition `yaml:"output_schema" json:"output_schema"`
    
    // Requirements
    Dependencies []string          `yaml:"dependencies" json:"dependencies"`
    Permissions  []Permission      `yaml:"permissions" json:"permissions"`
    Resources    ResourceRequirements `yaml:"resources" json:"resources"`
    
    // Behavioral characteristics
    Capabilities ToolCapabilities  `yaml:"capabilities" json:"capabilities"`
    Constraints  ToolConstraints   `yaml:"constraints" json:"constraints"`
}

type SchemaDefinition struct {
    Type       string                    `yaml:"type" json:"type"`
    Properties map[string]*PropertyDef   `yaml:"properties,omitempty" json:"properties,omitempty"`
    Required   []string                  `yaml:"required,omitempty" json:"required,omitempty"`
    Examples   []interface{}             `yaml:"examples,omitempty" json:"examples,omitempty"`
}

type PropertyDef struct {
    Type        string      `yaml:"type" json:"type"`
    Description string      `yaml:"description,omitempty" json:"description,omitempty"`
    Default     interface{} `yaml:"default,omitempty" json:"default,omitempty"`
    Enum        []string    `yaml:"enum,omitempty" json:"enum,omitempty"`
    Min         *float64    `yaml:"min,omitempty" json:"min,omitempty"`
    Max         *float64    `yaml:"max,omitempty" json:"max,omitempty"`
    Pattern     string      `yaml:"pattern,omitempty" json:"pattern,omitempty"`
}

type ResourceRequirements struct {
    Memory      int64         `yaml:"memory,omitempty" json:"memory,omitempty"`
    CPU         float64       `yaml:"cpu,omitempty" json:"cpu,omitempty"`
    Disk        int64         `yaml:"disk,omitempty" json:"disk,omitempty"`
    Network     bool          `yaml:"network,omitempty" json:"network,omitempty"`
    FileSystem  bool          `yaml:"file_system,omitempty" json:"file_system,omitempty"`
    Timeout     time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

type ToolConstraints struct {
    MaxInputSize   int64         `yaml:"max_input_size,omitempty" json:"max_input_size,omitempty"`
    MaxOutputSize  int64         `yaml:"max_output_size,omitempty" json:"max_output_size,omitempty"`
    MaxDuration    time.Duration `yaml:"max_duration,omitempty" json:"max_duration,omitempty"`
    Concurrency    int           `yaml:"concurrency,omitempty" json:"concurrency,omitempty"`
    RateLimit      *RateLimit    `yaml:"rate_limit,omitempty" json:"rate_limit,omitempty"`
}
```

### 2. Basic Tool Implementation

```go
// BaseTool provides a foundation for custom tools
type BaseTool struct {
    name         string
    description  string
    version      string
    category     string
    config       ToolConfig
    schema       *ToolSchema
    capabilities ToolCapabilities
    logger       Logger
}

// NewBaseTool creates a new base tool with common functionality
func NewBaseTool(name, description string) *BaseTool {
    return &BaseTool{
        name:        name,
        description: description,
        version:     "1.0.0",
        category:    "custom",
        config:      ToolConfig{},
        capabilities: ToolCapabilities{},
        logger:      NewLogger(name),
    }
}

// Implement core Tool interface methods
func (t *BaseTool) Name() string        { return t.name }
func (t *BaseTool) Description() string { return t.description }
func (t *BaseTool) Version() string     { return t.version }

func (t *BaseTool) GetConfig() ToolConfig {
    return t.config
}

func (t *BaseTool) SetConfig(config ToolConfig) error {
    t.config = config
    return t.validateConfig()
}

func (t *BaseTool) GetCapabilities() ToolCapabilities {
    return t.capabilities
}

func (t *BaseTool) Initialize(ctx context.Context) error {
    t.logger.Info("initializing tool", "name", t.name)
    return nil
}

func (t *BaseTool) Cleanup(ctx context.Context) error {
    t.logger.Info("cleaning up tool", "name", t.name)
    return nil
}

// Schema methods
func (t *BaseTool) GetInputSchema() *jsonschema.Schema {
    if t.schema == nil {
        return nil
    }
    return t.schema.InputSchema
}

func (t *BaseTool) GetOutputSchema() *jsonschema.Schema {
    if t.schema == nil {
        return nil
    }
    return t.schema.OutputSchema
}

func (t *BaseTool) ValidateInput(input interface{}) error {
    if t.schema == nil || t.schema.InputSchema == nil {
        return nil // No validation if no schema
    }
    
    return t.schema.InputSchema.Validate(input)
}

// Helper methods for common patterns
func (t *BaseTool) SetSchema(inputSchema, outputSchema *jsonschema.Schema) {
    t.schema = &ToolSchema{
        InputSchema:  inputSchema,
        OutputSchema: outputSchema,
    }
}

func (t *BaseTool) SetCapability(capability string, enabled bool) {
    switch capability {
    case "async":
        t.capabilities.Async = enabled
    case "streaming":
        t.capabilities.Streaming = enabled
    case "batching":
        t.capabilities.Batching = enabled
    case "cancellable":
        t.capabilities.Cancellable = enabled
    }
}

type ToolSchema struct {
    InputSchema  *jsonschema.Schema `json:"input_schema"`
    OutputSchema *jsonschema.Schema `json:"output_schema"`
}
```

### 3. Example: File Processing Tool

```go
// FileProcessorTool demonstrates a complete custom tool implementation
type FileProcessorTool struct {
    *BaseTool
    processor FileProcessor
    validator InputValidator
    monitor   *PerformanceMonitor
}

// FileProcessorConfig configures the file processor
type FileProcessorConfig struct {
    MaxFileSize   int64    `yaml:"max_file_size" json:"max_file_size"`
    AllowedTypes  []string `yaml:"allowed_types" json:"allowed_types"`
    OutputFormat  string   `yaml:"output_format" json:"output_format"`
    Compression   bool     `yaml:"compression" json:"compression"`
    Encoding      string   `yaml:"encoding" json:"encoding"`
}

// FileProcessorInput defines the input structure
type FileProcessorInput struct {
    FilePath    string                 `json:"file_path" validate:"required"`
    Operation   string                 `json:"operation" validate:"required,oneof=read write transform analyze"`
    Parameters  map[string]interface{} `json:"parameters,omitempty"`
    Options     FileProcessOptions     `json:"options,omitempty"`
}

type FileProcessOptions struct {
    Encoding    string `json:"encoding,omitempty"`
    Compression bool   `json:"compression,omitempty"`
    Streaming   bool   `json:"streaming,omitempty"`
    ChunkSize   int    `json:"chunk_size,omitempty"`
}

// FileProcessorOutput defines the output structure
type FileProcessorOutput struct {
    Success     bool                   `json:"success"`
    Result      interface{}            `json:"result,omitempty"`
    Metadata    FileMetadata           `json:"metadata"`
    Performance PerformanceStats       `json:"performance"`
    Errors      []string               `json:"errors,omitempty"`
}

type FileMetadata struct {
    Size         int64     `json:"size"`
    ModTime      time.Time `json:"mod_time"`
    Type         string    `json:"type"`
    Encoding     string    `json:"encoding,omitempty"`
    Checksum     string    `json:"checksum,omitempty"`
    ProcessedAt  time.Time `json:"processed_at"`
}

// NewFileProcessorTool creates a new file processor tool
func NewFileProcessorTool() *FileProcessorTool {
    tool := &FileProcessorTool{
        BaseTool:  NewBaseTool("file_processor", "Advanced file processing and analysis tool"),
        processor: NewFileProcessor(),
        validator: NewInputValidator(),
        monitor:   NewPerformanceMonitor(),
    }
    
    // Set version and category
    tool.version = "2.1.0"
    tool.category = "file_system"
    
    // Configure capabilities
    tool.SetCapability("async", true)
    tool.SetCapability("streaming", true)
    tool.SetCapability("cancellable", true)
    
    // Define schemas
    tool.setupSchemas()
    
    return tool
}

// Execute implements the main tool logic
func (t *FileProcessorTool) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // Start performance monitoring
    monitor := t.monitor.Start("execute")
    defer monitor.End()
    
    // Validate and parse input
    processInput, err := t.parseInput(input)
    if err != nil {
        return nil, fmt.Errorf("input validation failed: %w", err)
    }
    
    // Check file permissions and existence
    if err := t.validateFileAccess(processInput.FilePath); err != nil {
        return nil, fmt.Errorf("file access validation failed: %w", err)
    }
    
    // Get file metadata
    metadata, err := t.getFileMetadata(processInput.FilePath)
    if err != nil {
        return nil, fmt.Errorf("failed to get file metadata: %w", err)
    }
    
    // Validate file size and type
    if err := t.validateFileConstraints(metadata, processInput); err != nil {
        return nil, fmt.Errorf("file constraints validation failed: %w", err)
    }
    
    // Execute the requested operation
    result, err := t.executeOperation(ctx, processInput, metadata)
    if err != nil {
        return nil, fmt.Errorf("operation execution failed: %w", err)
    }
    
    // Prepare output
    output := &FileProcessorOutput{
        Success:     true,
        Result:      result,
        Metadata:    metadata,
        Performance: monitor.GetStats(),
    }
    
    return output, nil
}

// parseInput validates and converts input to structured format
func (t *FileProcessorTool) parseInput(input interface{}) (*FileProcessorInput, error) {
    // Convert input to bytes for JSON unmarshaling
    inputBytes, err := json.Marshal(input)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal input: %w", err)
    }
    
    var processInput FileProcessorInput
    if err := json.Unmarshal(inputBytes, &processInput); err != nil {
        return nil, fmt.Errorf("failed to unmarshal input: %w", err)
    }
    
    // Validate using struct tags
    if err := t.validator.Validate(&processInput); err != nil {
        return nil, fmt.Errorf("input validation failed: %w", err)
    }
    
    return &processInput, nil
}

// executeOperation performs the requested file operation
func (t *FileProcessorTool) executeOperation(ctx context.Context, input *FileProcessorInput, metadata FileMetadata) (interface{}, error) {
    switch input.Operation {
    case "read":
        return t.executeRead(ctx, input, metadata)
    case "write":
        return t.executeWrite(ctx, input, metadata)
    case "transform":
        return t.executeTransform(ctx, input, metadata)
    case "analyze":
        return t.executeAnalyze(ctx, input, metadata)
    default:
        return nil, fmt.Errorf("unsupported operation: %s", input.Operation)
    }
}

// executeRead reads file content
func (t *FileProcessorTool) executeRead(ctx context.Context, input *FileProcessorInput, metadata FileMetadata) (interface{}, error) {
    if input.Options.Streaming {
        return t.executeStreamingRead(ctx, input, metadata)
    }
    
    content, err := t.processor.ReadFile(input.FilePath, FileReadOptions{
        Encoding: input.Options.Encoding,
        MaxSize:  t.config.Parameters["max_file_size"].(int64),
}
    if err != nil {
        return nil, fmt.Errorf("failed to read file: %w", err)
    }
    
    return map[string]interface{}{
        "content": content,
        "size":    len(content),
        "encoding": input.Options.Encoding,
    }, nil
}

// executeStreamingRead implements streaming file reading
func (t *FileProcessorTool) executeStreamingRead(ctx context.Context, input *FileProcessorInput, metadata FileMetadata) (interface{}, error) {
    chunkSize := input.Options.ChunkSize
    if chunkSize == 0 {
        chunkSize = 4096 // Default chunk size
    }
    
    chunks := make([]string, 0)
    reader, err := t.processor.CreateStreamReader(input.FilePath, chunkSize)
    if err != nil {
        return nil, fmt.Errorf("failed to create stream reader: %w", err)
    }
    defer reader.Close()
    
    for {
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        default:
            chunk, err := reader.ReadChunk()
            if err == io.EOF {
                break
            }
            if err != nil {
                return nil, fmt.Errorf("failed to read chunk: %w", err)
            }
            chunks = append(chunks, chunk)
        }
    }
    
    return map[string]interface{}{
        "chunks":      chunks,
        "chunk_count": len(chunks),
        "total_size":  metadata.Size,
    }, nil
}

// executeAnalyze performs file analysis
func (t *FileProcessorTool) executeAnalyze(ctx context.Context, input *FileProcessorInput, metadata FileMetadata) (interface{}, error) {
    analyzer := NewFileAnalyzer()
    
    analysis, err := analyzer.AnalyzeFile(ctx, input.FilePath, AnalysisOptions{
        IncludeContent:  getBoolParam(input.Parameters, "include_content", false),
        IncludeStats:    getBoolParam(input.Parameters, "include_stats", true),
        IncludeChecksum: getBoolParam(input.Parameters, "include_checksum", true),
        DetectType:      getBoolParam(input.Parameters, "detect_type", true),
        SampleSize:      getIntParam(input.Parameters, "sample_size", 1024),
}
    if err != nil {
        return nil, fmt.Errorf("file analysis failed: %w", err)
    }
    
    return analysis, nil
}

// setupSchemas defines JSON schemas for input and output validation
func (t *FileProcessorTool) setupSchemas() {
    // Input schema
    inputSchema := &jsonschema.Schema{
        Type: "object",
        Properties: map[string]*jsonschema.Schema{
            "file_path": {
                Type:        "string",
                Description: "Path to the file to process",
                MinLength:   &[]int{1}[0],
            },
            "operation": {
                Type:        "string",
                Description: "Operation to perform on the file",
                Enum:        []interface{}{"read", "write", "transform", "analyze"},
            },
            "parameters": {
                Type:        "object",
                Description: "Operation-specific parameters",
            },
            "options": {
                Type: "object",
                Properties: map[string]*jsonschema.Schema{
                    "encoding": {
                        Type:    "string",
                        Default: "utf-8",
                    },
                    "compression": {
                        Type:    "boolean",
                        Default: false,
                    },
                    "streaming": {
                        Type:    "boolean",
                        Default: false,
                    },
                    "chunk_size": {
                        Type:    "integer",
                        Minimum: &[]float64{1}[0],
                        Maximum: &[]float64{1048576}[0], // 1MB max chunk
                    },
                },
            },
        },
        Required: []string{"file_path", "operation"},
    }
    
    // Output schema
    outputSchema := &jsonschema.Schema{
        Type: "object",
        Properties: map[string]*jsonschema.Schema{
            "success": {
                Type: "boolean",
            },
            "result": {
                Type:        "object",
                Description: "Operation result data",
            },
            "metadata": {
                Type: "object",
                Properties: map[string]*jsonschema.Schema{
                    "size":         {Type: "integer"},
                    "mod_time":     {Type: "string", Format: "date-time"},
                    "type":         {Type: "string"},
                    "encoding":     {Type: "string"},
                    "checksum":     {Type: "string"},
                    "processed_at": {Type: "string", Format: "date-time"},
                },
            },
            "performance": {
                Type: "object",
                Properties: map[string]*jsonschema.Schema{
                    "duration":     {Type: "number"},
                    "memory_used":  {Type: "integer"},
                    "cpu_time":     {Type: "number"},
                },
            },
            "errors": {
                Type: "array",
                Items: &jsonschema.Schema{Type: "string"},
            },
        },
        Required: []string{"success", "metadata", "performance"},
    }
    
    t.SetSchema(inputSchema, outputSchema)
}

// Helper functions for parameter extraction
func getBoolParam(params map[string]interface{}, key string, defaultValue bool) bool {
    if val, exists := params[key]; exists {
        if boolVal, ok := val.(bool); ok {
            return boolVal
        }
    }
    return defaultValue
}

func getIntParam(params map[string]interface{}, key string, defaultValue int) int {
    if val, exists := params[key]; exists {
        if intVal, ok := val.(int); ok {
            return intVal
        }
        if floatVal, ok := val.(float64); ok {
            return int(floatVal)
        }
    }
    return defaultValue
}
```

## Advanced Tool Patterns

### 4. Async Tool Implementation

```go
// AsyncTool provides asynchronous execution capabilities
type AsyncTool struct {
    *BaseTool
    executor *AsyncExecutor
    jobs     map[string]*AsyncJob
    mu       sync.RWMutex
}

type AsyncJob struct {
    ID        string                 `json:"id"`
    Status    JobStatus              `json:"status"`
    Progress  float64                `json:"progress"`
    Result    interface{}            `json:"result,omitempty"`
    Error     error                  `json:"error,omitempty"`
    StartTime time.Time              `json:"start_time"`
    EndTime   *time.Time             `json:"end_time,omitempty"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type JobStatus string

const (
    JobStatusPending   JobStatus = "pending"
    JobStatusRunning   JobStatus = "running"
    JobStatusCompleted JobStatus = "completed"
    JobStatusFailed    JobStatus = "failed"
    JobStatusCancelled JobStatus = "cancelled"
)

// ExecuteAsync starts asynchronous execution
func (t *AsyncTool) ExecuteAsync(ctx context.Context, input interface{}) (string, error) {
    jobID := generateJobID()
    
    job := &AsyncJob{
        ID:        jobID,
        Status:    JobStatusPending,
        StartTime: time.Now(),
        Metadata:  make(map[string]interface{}),
    }
    
    t.mu.Lock()
    t.jobs[jobID] = job
    t.mu.Unlock()
    
    // Start async execution
    go t.executeAsync(ctx, jobID, input)
    
    return jobID, nil
}

// GetJobStatus returns the status of an async job
func (t *AsyncTool) GetJobStatus(jobID string) (*AsyncJob, error) {
    t.mu.RLock()
    defer t.mu.RUnlock()
    
    job, exists := t.jobs[jobID]
    if !exists {
        return nil, fmt.Errorf("job %s not found", jobID)
    }
    
    return job, nil
}

// CancelJob cancels an async job
func (t *AsyncTool) CancelJob(jobID string) error {
    t.mu.Lock()
    defer t.mu.Unlock()
    
    job, exists := t.jobs[jobID]
    if !exists {
        return fmt.Errorf("job %s not found", jobID)
    }
    
    if job.Status == JobStatusRunning {
        job.Status = JobStatusCancelled
        now := time.Now()
        job.EndTime = &now
        
        // Cancel execution context
        return t.executor.CancelJob(jobID)
    }
    
    return nil
}

// executeAsync performs the actual async execution
func (t *AsyncTool) executeAsync(ctx context.Context, jobID string, input interface{}) {
    job := t.jobs[jobID]
    job.Status = JobStatusRunning
    
    defer func() {
        if r := recover(); r != nil {
            job.Status = JobStatusFailed
            job.Error = fmt.Errorf("panic during execution: %v", r)
            now := time.Now()
            job.EndTime = &now
        }
    }()
    
    // Execute with progress tracking
    result, err := t.executeWithProgress(ctx, input, func(progress float64) {
        t.updateJobProgress(jobID, progress)
}
    
    now := time.Now()
    job.EndTime = &now
    
    if err != nil {
        job.Status = JobStatusFailed
        job.Error = err
    } else {
        job.Status = JobStatusCompleted
        job.Result = result
        job.Progress = 1.0
    }
}

// updateJobProgress updates the progress of a job
func (t *AsyncTool) updateJobProgress(jobID string, progress float64) {
    t.mu.Lock()
    defer t.mu.Unlock()
    
    if job, exists := t.jobs[jobID]; exists {
        job.Progress = progress
    }
}
```

### 5. Streaming Tool Implementation

```go
// StreamingTool provides streaming execution capabilities
type StreamingTool struct {
    *BaseTool
    streamer *DataStreamer
}

type StreamChunk struct {
    ID        string      `json:"id"`
    Sequence  int         `json:"sequence"`
    Data      interface{} `json:"data"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
    IsLast    bool        `json:"is_last"`
    Error     error       `json:"error,omitempty"`
}

// ExecuteStream executes the tool with streaming output
func (t *StreamingTool) ExecuteStream(ctx context.Context, input interface{}) (<-chan StreamChunk, error) {
    // Validate input
    if err := t.ValidateInput(input); err != nil {
        return nil, fmt.Errorf("input validation failed: %w", err)
    }
    
    // Create output channel
    outputChan := make(chan StreamChunk, 100) // Buffered channel
    
    // Start streaming execution
    go t.executeStreaming(ctx, input, outputChan)
    
    return outputChan, nil
}

// executeStreaming performs streaming execution
func (t *StreamingTool) executeStreaming(ctx context.Context, input interface{}, output chan<- StreamChunk) {
    defer close(output)
    
    streamer := t.streamer.NewStream(input)
    sequence := 0
    
    for {
        select {
        case <-ctx.Done():
            output <- StreamChunk{
                ID:       generateChunkID(),
                Sequence: sequence,
                Error:    ctx.Err(),
                IsLast:   true,
            }
            return
            
        default:
            chunk, err := streamer.NextChunk()
            if err == io.EOF {
                output <- StreamChunk{
                    ID:       generateChunkID(),
                    Sequence: sequence,
                    IsLast:   true,
                }
                return
            }
            
            if err != nil {
                output <- StreamChunk{
                    ID:       generateChunkID(),
                    Sequence: sequence,
                    Error:    err,
                    IsLast:   true,
                }
                return
            }
            
            output <- StreamChunk{
                ID:       generateChunkID(),
                Sequence: sequence,
                Data:     chunk,
                Metadata: map[string]interface{}{
                    "size": len(chunk),
                    "timestamp": time.Now(),
                },
                IsLast: false,
            }
            
            sequence++
        }
    }
}
```

### 6. Stateful Tool Implementation

```go
// StatefulTool maintains state across executions
type StatefulTool struct {
    *BaseTool
    state     *ToolState
    stateMgr  StateManager
    mu        sync.RWMutex
}

type ToolState struct {
    ID          string                 `json:"id"`
    Version     int64                  `json:"version"`
    Data        map[string]interface{} `json:"data"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// NewStatefulTool creates a stateful tool
func NewStatefulTool(name, description string, stateMgr StateManager) *StatefulTool {
    tool := &StatefulTool{
        BaseTool: NewBaseTool(name, description),
        stateMgr: stateMgr,
        state: &ToolState{
            ID:        generateStateID(),
            Data:      make(map[string]interface{}),
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        },
    }
    
    tool.SetCapability("stateful", true)
    return tool
}

// GetState returns the current tool state
func (t *StatefulTool) GetState() *ToolState {
    t.mu.RLock()
    defer t.mu.RUnlock()
    
    // Return a copy to prevent external modifications
    stateCopy := *t.state
    stateCopy.Data = make(map[string]interface{})
    for k, v := range t.state.Data {
        stateCopy.Data[k] = v
    }
    
    return &stateCopy
}

// SetState updates the tool state
func (t *StatefulTool) SetState(key string, value interface{}) error {
    t.mu.Lock()
    defer t.mu.Unlock()
    
    t.state.Data[key] = value
    t.state.UpdatedAt = time.Now()
    t.state.Version++
    
    // Persist state
    return t.stateMgr.SaveState(t.state.ID, t.state)
}

// LoadState loads state from storage
func (t *StatefulTool) LoadState(stateID string) error {
    state, err := t.stateMgr.LoadState(stateID)
    if err != nil {
        return fmt.Errorf("failed to load state: %w", err)
    }
    
    t.mu.Lock()
    t.state = state
    t.mu.Unlock()
    
    return nil
}

// Execute with state management
func (t *StatefulTool) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // Validate input
    if err := t.ValidateInput(input); err != nil {
        return nil, fmt.Errorf("input validation failed: %w", err)
    }
    
    // Parse input to include state operations
    stateInput, err := t.parseStateInput(input)
    if err != nil {
        return nil, fmt.Errorf("failed to parse state input: %w", err)
    }
    
    // Apply state updates if requested
    if len(stateInput.StateUpdates) > 0 {
        if err := t.applyStateUpdates(stateInput.StateUpdates); err != nil {
            return nil, fmt.Errorf("failed to apply state updates: %w", err)
        }
    }
    
    // Execute main logic with access to state
    result, err := t.executeWithState(ctx, stateInput)
    if err != nil {
        return nil, err
    }
    
    // Include current state in output if requested
    output := map[string]interface{}{
        "result": result,
    }
    
    if stateInput.IncludeState {
        output["state"] = t.GetState()
    }
    
    return output, nil
}

type StateInput struct {
    Operation     string                    `json:"operation"`
    Parameters    map[string]interface{}    `json:"parameters"`
    StateUpdates  map[string]interface{}    `json:"state_updates,omitempty"`
    IncludeState  bool                      `json:"include_state,omitempty"`
}

// parseStateInput extracts state-related operations from input
func (t *StatefulTool) parseStateInput(input interface{}) (*StateInput, error) {
    inputBytes, err := json.Marshal(input)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal input: %w", err)
    }
    
    var stateInput StateInput
    if err := json.Unmarshal(inputBytes, &stateInput); err != nil {
        return nil, fmt.Errorf("failed to unmarshal state input: %w", err)
    }
    
    return &stateInput, nil
}

// applyStateUpdates applies the requested state updates
func (t *StatefulTool) applyStateUpdates(updates map[string]interface{}) error {
    for key, value := range updates {
        if err := t.SetState(key, value); err != nil {
            return fmt.Errorf("failed to update state key %s: %w", key, err)
        }
    }
    return nil
}
```

## Tool Testing Framework

### 7. Comprehensive Testing

```go
// ToolTester provides comprehensive testing capabilities
type ToolTester struct {
    tool     Tool
    registry ToolRegistry
    executor ToolExecutor
    reporter TestReporter
}

// NewToolTester creates a new tool tester
func NewToolTester(tool Tool) *ToolTester {
    return &ToolTester{
        tool:     tool,
        registry: NewTestRegistry(),
        executor: NewTestExecutor(),
        reporter: NewTestReporter(),
    }
}

// TestSuite represents a complete test suite for a tool
type TestSuite struct {
    Name        string      `yaml:"name" json:"name"`
    Description string      `yaml:"description" json:"description"`
    Setup       []TestStep  `yaml:"setup,omitempty" json:"setup,omitempty"`
    Tests       []TestCase  `yaml:"tests" json:"tests"`
    Teardown    []TestStep  `yaml:"teardown,omitempty" json:"teardown,omitempty"`
}

type TestCase struct {
    Name        string                 `yaml:"name" json:"name"`
    Description string                 `yaml:"description,omitempty" json:"description,omitempty"`
    Input       interface{}            `yaml:"input" json:"input"`
    Expected    ExpectedResult         `yaml:"expected" json:"expected"`
    Timeout     time.Duration          `yaml:"timeout,omitempty" json:"timeout,omitempty"`
    Setup       []TestStep             `yaml:"setup,omitempty" json:"setup,omitempty"`
    Cleanup     []TestStep             `yaml:"cleanup,omitempty" json:"cleanup,omitempty"`
    Tags        []string               `yaml:"tags,omitempty" json:"tags,omitempty"`
    Skip        bool                   `yaml:"skip,omitempty" json:"skip,omitempty"`
    SkipReason  string                 `yaml:"skip_reason,omitempty" json:"skip_reason,omitempty"`
}

type ExpectedResult struct {
    Success     *bool                  `yaml:"success,omitempty" json:"success,omitempty"`
    Output      interface{}            `yaml:"output,omitempty" json:"output,omitempty"`
    Error       *string                `yaml:"error,omitempty" json:"error,omitempty"`
    Validation  []ValidationRule       `yaml:"validation,omitempty" json:"validation,omitempty"`
    Performance *PerformanceExpectation `yaml:"performance,omitempty" json:"performance,omitempty"`
}

type ValidationRule struct {
    Field     string      `yaml:"field" json:"field"`
    Operator  string      `yaml:"operator" json:"operator"`
    Value     interface{} `yaml:"value" json:"value"`
    Message   string      `yaml:"message,omitempty" json:"message,omitempty"`
}

type PerformanceExpectation struct {
    MaxDuration time.Duration `yaml:"max_duration,omitempty" json:"max_duration,omitempty"`
    MaxMemory   int64         `yaml:"max_memory,omitempty" json:"max_memory,omitempty"`
    MaxCPU      float64       `yaml:"max_cpu,omitempty" json:"max_cpu,omitempty"`
}

// RunTestSuite executes a complete test suite
func (t *ToolTester) RunTestSuite(ctx context.Context, suite *TestSuite) (*TestResults, error) {
    results := &TestResults{
        SuiteName:   suite.Name,
        StartTime:   time.Now(),
        TestResults: make([]TestResult, 0, len(suite.Tests)),
    }
    
    // Run setup steps
    if err := t.runTestSteps(ctx, suite.Setup); err != nil {
        return nil, fmt.Errorf("setup failed: %w", err)
    }
    
    defer func() {
        // Run teardown steps
        if err := t.runTestSteps(ctx, suite.Teardown); err != nil {
            t.reporter.ReportError("teardown failed", err)
        }
    }()
    
    // Run individual tests
    for _, testCase := range suite.Tests {
        if testCase.Skip {
            results.TestResults = append(results.TestResults, TestResult{
                Name:     testCase.Name,
                Status:   TestStatusSkipped,
                Message:  testCase.SkipReason,
                Duration: 0,
}
            continue
        }
        
        result := t.runTestCase(ctx, &testCase)
        results.TestResults = append(results.TestResults, result)
    }
    
    // Calculate summary
    results.EndTime = time.Now()
    results.Duration = results.EndTime.Sub(results.StartTime)
    results.Summary = t.calculateSummary(results.TestResults)
    
    return results, nil
}

// runTestCase executes a single test case
func (t *ToolTester) runTestCase(ctx context.Context, testCase *TestCase) TestResult {
    result := TestResult{
        Name:      testCase.Name,
        StartTime: time.Now(),
    }
    
    defer func() {
        result.EndTime = time.Now()
        result.Duration = result.EndTime.Sub(result.StartTime)
    }()
    
    // Run test setup
    if err := t.runTestSteps(ctx, testCase.Setup); err != nil {
        result.Status = TestStatusFailed
        result.Error = fmt.Sprintf("test setup failed: %v", err)
        return result
    }
    
    defer func() {
        // Run test cleanup
        if err := t.runTestSteps(ctx, testCase.Cleanup); err != nil {
            t.reporter.ReportWarning("test cleanup failed", err)
        }
    }()
    
    // Set timeout if specified
    testCtx := ctx
    if testCase.Timeout > 0 {
        var cancel context.CancelFunc
        testCtx, cancel = context.WithTimeout(ctx, testCase.Timeout)
        defer cancel()
    }
    
    // Execute the tool
    output, err := t.executor.Execute(testCtx, t.tool, testCase.Input)
    
    // Validate results
    if err := t.validateResult(output, err, &testCase.Expected); err != nil {
        result.Status = TestStatusFailed
        result.Error = err.Error()
        return result
    }
    
    result.Status = TestStatusPassed
    result.Output = output
    
    return result
}

// validateResult validates the execution result against expectations
func (t *ToolTester) validateResult(output *ExecutionResult, err error, expected *ExpectedResult) error {
    // Check success expectation
    if expected.Success != nil {
        actualSuccess := err == nil
        if *expected.Success != actualSuccess {
            if *expected.Success {
                return fmt.Errorf("expected success but got error: %v", err)
            } else {
                return fmt.Errorf("expected failure but execution succeeded")
            }
        }
    }
    
    // Check error expectation
    if expected.Error != nil {
        if err == nil {
            return fmt.Errorf("expected error %q but execution succeeded", *expected.Error)
        }
        if !strings.Contains(err.Error(), *expected.Error) {
            return fmt.Errorf("expected error containing %q but got %q", *expected.Error, err.Error())
        }
    }
    
    // Check output expectation
    if expected.Output != nil && output != nil {
        if !t.compareValues(expected.Output, output.Output) {
            return fmt.Errorf("output mismatch: expected %v, got %v", expected.Output, output.Output)
        }
    }
    
    // Run validation rules
    for _, rule := range expected.Validation {
        if err := t.validateRule(output, rule); err != nil {
            return fmt.Errorf("validation rule failed: %w", err)
        }
    }
    
    // Check performance expectations
    if expected.Performance != nil && output != nil {
        if err := t.validatePerformance(output, expected.Performance); err != nil {
            return fmt.Errorf("performance expectation failed: %w", err)
        }
    }
    
    return nil
}

type TestResults struct {
    SuiteName   string        `json:"suite_name"`
    StartTime   time.Time     `json:"start_time"`
    EndTime     time.Time     `json:"end_time"`
    Duration    time.Duration `json:"duration"`
    TestResults []TestResult  `json:"test_results"`
    Summary     TestSummary   `json:"summary"`
}

type TestResult struct {
    Name      string        `json:"name"`
    Status    TestStatus    `json:"status"`
    StartTime time.Time     `json:"start_time"`
    EndTime   time.Time     `json:"end_time"`
    Duration  time.Duration `json:"duration"`
    Output    interface{}   `json:"output,omitempty"`
    Error     string        `json:"error,omitempty"`
    Message   string        `json:"message,omitempty"`
}

type TestStatus string

const (
    TestStatusPassed  TestStatus = "passed"
    TestStatusFailed  TestStatus = "failed"
    TestStatusSkipped TestStatus = "skipped"
    TestStatusError   TestStatus = "error"
)

type TestSummary struct {
    Total   int `json:"total"`
    Passed  int `json:"passed"`
    Failed  int `json:"failed"`
    Skipped int `json:"skipped"`
    Errors  int `json:"errors"`
}
```

## Tool Packaging and Distribution

### 8. Tool Packaging

```go
// ToolPackage represents a packaged tool for distribution
type ToolPackage struct {
    Metadata    PackageMetadata    `yaml:"metadata" json:"metadata"`
    Tool        Tool               `yaml:"-" json:"-"`
    Config      ToolConfig         `yaml:"config" json:"config"`
    Dependencies []Dependency      `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`
    Resources   []Resource         `yaml:"resources,omitempty" json:"resources,omitempty"`
    Tests       *TestSuite         `yaml:"tests,omitempty" json:"tests,omitempty"`
    Examples    []ToolExample      `yaml:"examples,omitempty" json:"examples,omitempty"`
    Documentation *Documentation  `yaml:"documentation,omitempty" json:"documentation,omitempty"`
}

type PackageMetadata struct {
    Name        string            `yaml:"name" json:"name"`
    Version     string            `yaml:"version" json:"version"`
    Description string            `yaml:"description" json:"description"`
    Author      string            `yaml:"author" json:"author"`
    License     string            `yaml:"license" json:"license"`
    Homepage    string            `yaml:"homepage,omitempty" json:"homepage,omitempty"`
    Repository  string            `yaml:"repository,omitempty" json:"repository,omitempty"`
    Keywords    []string          `yaml:"keywords,omitempty" json:"keywords,omitempty"`
    Tags        map[string]string `yaml:"tags,omitempty" json:"tags,omitempty"`
}

type Dependency struct {
    Name    string `yaml:"name" json:"name"`
    Version string `yaml:"version" json:"version"`
    Source  string `yaml:"source,omitempty" json:"source,omitempty"`
    Optional bool  `yaml:"optional,omitempty" json:"optional,omitempty"`
}

type Resource struct {
    Name string `yaml:"name" json:"name"`
    Type string `yaml:"type" json:"type"`
    Path string `yaml:"path" json:"path"`
    Size int64  `yaml:"size,omitempty" json:"size,omitempty"`
}

// ToolPackager creates tool packages
type ToolPackager struct {
    builder   *PackageBuilder
    validator *PackageValidator
    publisher *PackagePublisher
}

// CreatePackage creates a new tool package
func (p *ToolPackager) CreatePackage(tool Tool, metadata PackageMetadata) (*ToolPackage, error) {
    pkg := &ToolPackage{
        Metadata: metadata,
        Tool:     tool,
        Config:   tool.GetConfig(),
    }
    
    // Extract dependencies
    deps, err := p.extractDependencies(tool)
    if err != nil {
        return nil, fmt.Errorf("failed to extract dependencies: %w", err)
    }
    pkg.Dependencies = deps
    
    // Generate examples
    examples, err := p.generateExamples(tool)
    if err != nil {
        return nil, fmt.Errorf("failed to generate examples: %w", err)
    }
    pkg.Examples = examples
    
    // Generate documentation
    docs, err := p.generateDocumentation(tool)
    if err != nil {
        return nil, fmt.Errorf("failed to generate documentation: %w", err)
    }
    pkg.Documentation = docs
    
    // Validate package
    if err := p.validator.ValidatePackage(pkg); err != nil {
        return nil, fmt.Errorf("package validation failed: %w", err)
    }
    
    return pkg, nil
}

// BuildPackage builds a distributable package
func (p *ToolPackager) BuildPackage(pkg *ToolPackage, outputPath string) error {
    return p.builder.Build(pkg, BuildOptions{
        OutputPath:     outputPath,
        Format:         "tar.gz",
        IncludeTests:   true,
        IncludeDocs:    true,
        Compression:    true,
        Verification:   true,
}
}

type BuildOptions struct {
    OutputPath     string `json:"output_path"`
    Format         string `json:"format"`
    IncludeTests   bool   `json:"include_tests"`
    IncludeDocs    bool   `json:"include_docs"`
    Compression    bool   `json:"compression"`
    Verification   bool   `json:"verification"`
}
```

This comprehensive guide covers all aspects of creating custom tools in Go-LLMs, from basic implementation to advanced patterns, testing, and packaging. The examples provide complete, production-ready implementations that can serve as templates for building robust, scalable tools that integrate seamlessly with the Go-LLMs ecosystem.