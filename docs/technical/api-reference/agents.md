# Agent Interface Documentation

> **[Project Root](/) / [Documentation](../..) / [Technical Documentation](../../technical) / [API Reference](../../technical/api-reference) / Agents**

Complete API reference for agent interfaces and implementations in Go-LLMs, covering core agent interfaces, specialized agent types, workflow orchestration, state management, and multi-agent coordination patterns.

## Core Agent Interfaces

### BaseAgent Interface

The core interface that all agent types must implement (`pkg/agent/domain`):

```go
// BaseAgent defines the core interface for all agent types.
type BaseAgent interface {
    // Identification
    ID() string
    Name() string
    Description() string
    Type() AgentType

    // Hierarchy Management
    Parent() BaseAgent
    SetParent(parent BaseAgent) error
    SubAgents() []BaseAgent
    AddSubAgent(agent BaseAgent) error
    RemoveSubAgent(name string) error
    FindAgent(name string) BaseAgent
    FindSubAgent(name string) BaseAgent

    // Execution
    Run(ctx context.Context, input *State) (*State, error)
    RunAsync(ctx context.Context, input *State) (<-chan Event, error)

    // Lifecycle Hooks
    Initialize(ctx context.Context) error
    BeforeRun(ctx context.Context, state *State) error
    AfterRun(ctx context.Context, state *State, result *State, err error) error
    Cleanup(ctx context.Context) error

    // Schema Definition
    InputSchema() *schema.Schema
    OutputSchema() *schema.Schema

    // Configuration
    Config() AgentConfig
    WithConfig(config AgentConfig) BaseAgent
    Validate() error

    // Metadata
    Metadata() map[string]interface{}
    SetMetadata(key string, value interface{})
}
```

`AgentType` constants: `"llm"`, `"sequential"`, `"parallel"`, `"conditional"`, `"loop"`, `"custom"`

#### Key Methods

##### Run

```go
Run(ctx context.Context, input *State) (*State, error)
```

Executes the agent synchronously. Input and output are both `*domain.State` key-value maps.

**Example:**
```go
import agentdomain "github.com/lexlapax/go-llms/pkg/agent/domain"

state := agentdomain.NewState()
state.Set("user_input", "Summarize this document")
result, err := agent.Run(ctx, state)
if response, ok := result.Get("response"); ok {
    fmt.Println(response)
}
```

##### RunAsync

```go
RunAsync(ctx context.Context, input *State) (<-chan Event, error)
```

Executes the agent asynchronously, streaming events as they occur.

##### AddSubAgent / SubAgents

```go
AddSubAgent(agent BaseAgent) error
SubAgents() []BaseAgent
```

Manage child agents for workflow orchestration (sequential, parallel, etc.).

## Agent Types and Implementations

### SimpleAgent

Basic agent implementation:

```go
// SimpleAgent provides a basic agent implementation
type SimpleAgent struct {
    config   AgentConfig
    metadata AgentMetadata
    handler  ExecutionHandler
}

// ExecutionHandler defines the execution logic
type ExecutionHandler func(ctx context.Context, input interface{}) (interface{}, error)

// NewSimpleAgent creates a new simple agent
func NewSimpleAgent(config AgentConfig, handler ExecutionHandler) *SimpleAgent
```

#### Usage Example

```go
agent := agent.NewSimpleAgent(
    agent.AgentConfig{
        Name:        "data-processor",
        Description: "Processes data files",
    },
    func(ctx context.Context, input interface{}) (interface{}, error) {
        // Agent logic here
        data := input.(map[string]interface{})
        // Process data
        return processedData, nil
    },
)

result, err := agent.Execute(ctx, inputData)
```

### LLMAgentImpl

Language model-powered agent:

```go
// LLMAgentImpl implements LLM-based agent
type LLMAgentImpl struct {
    *ToolEnabledAgentImpl
    provider        Provider
    systemPrompt    string
    history         []Message
    maxHistory      int
    temperature     float64
}

// NewLLMAgent creates a new LLM agent
func NewLLMAgent(config LLMAgentConfig) *LLMAgentImpl

// LLMAgentConfig configures an LLM agent
type LLMAgentConfig struct {
    AgentConfig
    Provider        Provider      `json:"-"`
    SystemPrompt    string        `json:"system_prompt"`
    MaxHistory      int           `json:"max_history"`
    Temperature     float64       `json:"temperature"`
    Model           string        `json:"model"`
    Tools           []string      `json:"tools"`
}
```

#### Usage Example

```go
llmAgent := agent.NewLLMAgent(agent.LLMAgentConfig{
    AgentConfig: agent.AgentConfig{
        Name:        "assistant",
        Description: "AI-powered assistant",
    },
    Provider:     openaiProvider,
    SystemPrompt: "You are a helpful AI assistant.",
    MaxHistory:   10,
    Temperature:  0.7,
    Model:        "gpt-4",
    Tools:        []string{"http_request", "file_reader"},
}

// Register tools
for _, toolName := range config.Tools {
    tool := tools.GetTool(toolName)
    llmAgent.RegisterTool(tool)
}

// Execute
result, err := llmAgent.Execute(ctx, "Find information about Go programming")
```

### WorkflowAgentImpl

Workflow orchestration agent:

```go
// WorkflowAgentImpl implements workflow orchestration
type WorkflowAgentImpl struct {
    *BaseAgent
    steps       []WorkflowStep
    state       WorkflowState
    executor    StepExecutor
    validator   StepValidator
}

// NewWorkflowAgent creates a new workflow agent
func NewWorkflowAgent(config WorkflowAgentConfig) *WorkflowAgentImpl

// WorkflowAgentConfig configures a workflow agent
type WorkflowAgentConfig struct {
    AgentConfig
    Steps           []WorkflowStep `json:"steps"`
    ErrorHandling   ErrorHandling  `json:"error_handling"`
    Parallelization bool           `json:"parallelization"`
    MaxConcurrency  int            `json:"max_concurrency"`
}
```

## Agent Configuration

### AgentConfig

Base configuration for all agents:

```go
// AgentConfig provides base configuration for agents
type AgentConfig struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Version     string                 `json:"version"`
    Tags        []string               `json:"tags"`
    Timeout     time.Duration          `json:"timeout"`
    RetryPolicy RetryPolicy            `json:"retry_policy"`
    Parameters  map[string]interface{} `json:"parameters"`
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
    MaxAttempts    int           `json:"max_attempts"`
    InitialDelay   time.Duration `json:"initial_delay"`
    MaxDelay       time.Duration `json:"max_delay"`
    BackoffFactor  float64       `json:"backoff_factor"`
}
```

### AgentMetadata

Metadata describing agent capabilities:

```go
// AgentMetadata describes an agent
type AgentMetadata struct {
    ID           string            `json:"id"`
    Name         string            `json:"name"`
    Description  string            `json:"description"`
    Version      string            `json:"version"`
    Author       string            `json:"author"`
    CreatedAt    time.Time         `json:"created_at"`
    UpdatedAt    time.Time         `json:"updated_at"`
    Tags         []string          `json:"tags"`
    Capabilities []string          `json:"capabilities"`
    Requirements []string          `json:"requirements"`
    Schema       *AgentSchema      `json:"schema"`
}

// AgentSchema defines input/output schemas
type AgentSchema struct {
    InputSchema  *jsonschema.Schema `json:"input_schema"`
    OutputSchema *jsonschema.Schema `json:"output_schema"`
}
```

## Workflow Components

### WorkflowStep

Represents a step in a workflow:

```go
// WorkflowStep defines a single workflow step
type WorkflowStep struct {
    Name         string                 `json:"name"`
    Type         StepType               `json:"type"`
    Description  string                 `json:"description"`
    Agent        string                 `json:"agent,omitempty"`
    Tool         string                 `json:"tool,omitempty"`
    Input        interface{}            `json:"input,omitempty"`
    InputMapping map[string]string      `json:"input_mapping,omitempty"`
    Output       string                 `json:"output,omitempty"`
    Dependencies []string               `json:"dependencies,omitempty"`
    Condition    string                 `json:"condition,omitempty"`
    OnError      ErrorAction            `json:"on_error,omitempty"`
    Timeout      time.Duration          `json:"timeout,omitempty"`
    RetryPolicy  *RetryPolicy           `json:"retry_policy,omitempty"`
}

// StepType defines the type of workflow step
type StepType string

const (
    StepTypeAgent     StepType = "agent"
    StepTypeTool      StepType = "tool"
    StepTypeCondition StepType = "condition"
    StepTypeLoop      StepType = "loop"
    StepTypeParallel  StepType = "parallel"
    StepTypeSubflow   StepType = "subflow"
)

// ErrorAction defines how to handle step errors
type ErrorAction string

const (
    ErrorActionFail     ErrorAction = "fail"
    ErrorActionContinue ErrorAction = "continue"
    ErrorActionRetry    ErrorAction = "retry"
    ErrorActionFallback ErrorAction = "fallback"
)
```

### WorkflowState

Maintains workflow execution state:

```go
// WorkflowState tracks workflow execution
type WorkflowState interface {
    // GetStepResult gets the result of a step
    GetStepResult(stepName string) (StepResult, bool)
    
    // SetStepResult sets the result of a step
    SetStepResult(stepName string, result StepResult)
    
    // GetVariable gets a workflow variable
    GetVariable(name string) interface{}
    
    // SetVariable sets a workflow variable
    SetVariable(name string, value interface{})
    
    // GetCurrentStep returns the currently executing step
    GetCurrentStep() string
    
    // GetExecutionPath returns the execution path
    GetExecutionPath() []string
    
    // IsCompleted checks if the workflow is complete
    IsCompleted() bool
    
    // GetError returns any workflow error
    GetError() error
}

// StepResult represents the result of a step execution
type StepResult struct {
    StepName    string        `json:"step_name"`
    Success     bool          `json:"success"`
    Output      interface{}   `json:"output"`
    Error       error         `json:"error"`
    StartTime   time.Time     `json:"start_time"`
    EndTime     time.Time     `json:"end_time"`
    Duration    time.Duration `json:"duration"`
    Attempts    int           `json:"attempts"`
}
```

### WorkflowResult

Final result of workflow execution:

```go
// WorkflowResult represents the complete workflow result
type WorkflowResult struct {
    WorkflowID   string                 `json:"workflow_id"`
    Success      bool                   `json:"success"`
    Steps        []StepResult           `json:"steps"`
    Output       interface{}            `json:"output"`
    Error        error                  `json:"error"`
    StartTime    time.Time              `json:"start_time"`
    EndTime      time.Time              `json:"end_time"`
    Duration     time.Duration          `json:"duration"`
    State        map[string]interface{} `json:"state"`
}
```

## Multi-Agent Systems

### MultiAgentSystem Interface

Coordinates multiple agents:

```go
// MultiAgentSystem manages multiple cooperating agents
type MultiAgentSystem interface {
    // RegisterAgent adds an agent to the system
    RegisterAgent(agent Agent) error
    
    // UnregisterAgent removes an agent from the system
    UnregisterAgent(name string) error
    
    // GetAgent retrieves an agent by name
    GetAgent(name string) (Agent, error)
    
    // ListAgents returns all registered agents
    ListAgents() []AgentInfo
    
    // Coordinate coordinates task execution among agents
    Coordinate(ctx context.Context, task Task) (*CoordinationResult, error)
    
    // Broadcast sends a message to all agents
    Broadcast(ctx context.Context, message Message) error
    
    // SendMessage sends a message to a specific agent
    SendMessage(ctx context.Context, to string, message Message) error
}

// Task represents a task to be coordinated
type Task struct {
    ID           string                 `json:"id"`
    Type         string                 `json:"type"`
    Description  string                 `json:"description"`
    Requirements []string               `json:"requirements"`
    Input        interface{}            `json:"input"`
    Constraints  TaskConstraints        `json:"constraints"`
}

// CoordinationResult represents the result of coordination
type CoordinationResult struct {
    TaskID       string                 `json:"task_id"`
    Success      bool                   `json:"success"`
    Assignments  []AgentAssignment      `json:"assignments"`
    Results      map[string]interface{} `json:"results"`
    Timeline     []TimelineEvent        `json:"timeline"`
}
```

### AgentCommunication

Inter-agent communication:

```go
// Message represents an inter-agent message
type Message struct {
    ID          string                 `json:"id"`
    From        string                 `json:"from"`
    To          string                 `json:"to"`
    Type        MessageType            `json:"type"`
    Content     interface{}            `json:"content"`
    Timestamp   time.Time              `json:"timestamp"`
    InReplyTo   string                 `json:"in_reply_to,omitempty"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// MessageType defines message types
type MessageType string

const (
    MessageTypeRequest    MessageType = "request"
    MessageTypeResponse   MessageType = "response"
    MessageTypeNotify     MessageType = "notify"
    MessageTypeBroadcast  MessageType = "broadcast"
    MessageTypeError      MessageType = "error"
)
```

## State Management

### StatefulAgent Interface

Agent with persistent state:

```go
// StatefulAgent maintains state across executions
type StatefulAgent interface {
    Agent
    
    // GetState returns the current agent state
    GetState() AgentState
    
    // SetState updates the agent state
    SetState(state AgentState) error
    
    // SaveState persists the state
    SaveState(ctx context.Context) error
    
    // LoadState loads persisted state
    LoadState(ctx context.Context) error
    
    // ResetState resets to initial state
    ResetState() error
}

// AgentState represents agent state
type AgentState interface {
    // Get retrieves a state value
    Get(key string) (interface{}, bool)
    
    // Set stores a state value
    Set(key string, value interface{}) error
    
    // Delete removes a state value
    Delete(key string) error
    
    // GetAll returns all state data
    GetAll() map[string]interface{}
    
    // Clear removes all state data
    Clear() error
    
    // Version returns the state version
    Version() int64
}
```

### Memory Systems

Agent memory implementations:

```go
// AgentMemory provides memory capabilities
type AgentMemory interface {
    // Store stores information in memory
    Store(ctx context.Context, key string, value interface{}) error
    
    // Retrieve retrieves information from memory
    Retrieve(ctx context.Context, key string) (interface{}, error)
    
    // Search searches memory by query
    Search(ctx context.Context, query string) ([]MemoryItem, error)
    
    // Forget removes information from memory
    Forget(ctx context.Context, key string) error
    
    // GetStats returns memory statistics
    GetStats() MemoryStats
}

// MemoryItem represents an item in memory
type MemoryItem struct {
    Key          string                 `json:"key"`
    Value        interface{}            `json:"value"`
    StoredAt     time.Time              `json:"stored_at"`
    AccessCount  int                    `json:"access_count"`
    LastAccessed time.Time              `json:"last_accessed"`
    Tags         []string               `json:"tags"`
    Metadata     map[string]interface{} `json:"metadata"`
}
```

## Agent Lifecycle

### Initialization

```go
// Initialize prepares an agent for execution
func (a *BaseAgent) Initialize(ctx context.Context) error {
    // Validate configuration
    if err := a.validateConfig(); err != nil {
        return fmt.Errorf("invalid configuration: %w", err)
    }
    
    // Initialize resources
    if err := a.initializeResources(ctx); err != nil {
        return fmt.Errorf("resource initialization failed: %w", err)
    }
    
    // Load state if stateful
    if stateful, ok := a.(StatefulAgent); ok {
        if err := stateful.LoadState(ctx); err != nil {
            return fmt.Errorf("state loading failed: %w", err)
        }
    }
    
    a.initialized = true
    return nil
}
```

### Shutdown

```go
// Shutdown cleans up agent resources
func (a *BaseAgent) Shutdown(ctx context.Context) error {
    if !a.initialized {
        return nil
    }
    
    // Save state if stateful
    if stateful, ok := a.(StatefulAgent); ok {
        if err := stateful.SaveState(ctx); err != nil {
            // Log error but continue shutdown
            a.logger.Error("failed to save state", "error", err)
        }
    }
    
    // Cleanup resources
    if err := a.cleanupResources(ctx); err != nil {
        return fmt.Errorf("resource cleanup failed: %w", err)
    }
    
    a.initialized = false
    return nil
}
```

## Error Handling

### Agent Errors

```go
// AgentError represents an agent-specific error
type AgentError struct {
    Agent    string                 `json:"agent"`
    Code     string                 `json:"code"`
    Message  string                 `json:"message"`
    Details  map[string]interface{} `json:"details"`
    Cause    error                  `json:"-"`
}

// Common error codes
const (
    ErrCodeInitialization = "initialization_failed"
    ErrCodeExecution      = "execution_failed"
    ErrCodeTimeout        = "timeout"
    ErrCodeInvalidInput   = "invalid_input"
    ErrCodeToolNotFound   = "tool_not_found"
    ErrCodeStateError     = "state_error"
    ErrCodeCoordination   = "coordination_failed"
)
```

## Best Practices

### 1. Agent Design

Design agents with single responsibility:

```go
// Good: Focused agent
type DataValidatorAgent struct {
    *BaseAgent
    schema *jsonschema.Schema
}

func (a *DataValidatorAgent) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // Validate data against schema
    if err := a.schema.Validate(input); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    return input, nil
}
```

### 2. Error Handling

Implement comprehensive error handling:

```go
func (a *MyAgent) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // Validate input
    if err := a.validateInput(input); err != nil {
        return nil, &AgentError{
            Agent:   a.GetMetadata().Name,
            Code:    ErrCodeInvalidInput,
            Message: "input validation failed",
            Cause:   err,
        }
    }
    
    // Execute with timeout
    execCtx, cancel := context.WithTimeout(ctx, a.config.Timeout)
    defer cancel()
    
    result, err := a.executeCore(execCtx, input)
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            return nil, &AgentError{
                Agent:   a.GetMetadata().Name,
                Code:    ErrCodeTimeout,
                Message: "execution timeout",
            }
        }
        return nil, err
    }
    
    return result, nil
}
```

### 3. State Management

Implement proper state management:

```go
type StatefulAgentImpl struct {
    *BaseAgent
    state      AgentState
    stateStore StateStore
}

func (a *StatefulAgentImpl) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // Load latest state
    if err := a.LoadState(ctx); err != nil {
        return nil, fmt.Errorf("failed to load state: %w", err)
    }
    
    // Execute with state
    result, err := a.executeWithState(ctx, input)
    if err != nil {
        return nil, err
    }
    
    // Save updated state
    if err := a.SaveState(ctx); err != nil {
        return nil, fmt.Errorf("failed to save state: %w", err)
    }
    
    return result, nil
}
```

### 4. Tool Integration

Properly integrate tools:

```go
func (a *ToolEnabledAgentImpl) ExecuteWithTools(ctx context.Context, task string) (interface{}, error) {
    // Determine required tools
    requiredTools := a.analyzeTask(task)
    
    // Verify tools are available
    for _, toolName := range requiredTools {
        if _, err := a.toolTools.Get(toolName); err != nil {
            return nil, fmt.Errorf("required tool %s not found: %w", toolName, err)
        }
    }
    
    // Execute task with tools
    results := make(map[string]interface{})
    for _, toolName := range requiredTools {
        result, err := a.ExecuteTool(ctx, toolName, task)
        if err != nil {
            return nil, fmt.Errorf("tool %s execution failed: %w", toolName, err)
        }
        results[toolName] = result
    }
    
    return results, nil
}
```

### 5. Workflow Best Practices

Design robust workflows:

```go
workflow := &WorkflowAgentImpl{
    steps: []WorkflowStep{
        {
            Name: "validate",
            Type: StepTypeTool,
            Tool: "validator",
            OnError: ErrorActionFail,
        },
        {
            Name: "process",
            Type: StepTypeAgent,
            Agent: "processor",
            Dependencies: []string{"validate"},
            RetryPolicy: &RetryPolicy{
                MaxAttempts: 3,
                InitialDelay: time.Second,
            },
        },
        {
            Name: "save",
            Type: StepTypeTool,
            Tool: "file_writer",
            Dependencies: []string{"process"},
            OnError: ErrorActionRetry,
        },
    },
}
```

This comprehensive agent API documentation provides all the necessary information for building sophisticated agent-based applications with Go-LLMs.