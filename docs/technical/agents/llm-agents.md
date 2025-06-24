# LLM Agents: AI-Powered Agents with Tool Support

> **[Project Root](/) / [Documentation](/docs/) / [Technical Documentation](/docs/technical/) / [Agents](/docs/technical/agents/) / LLM Agents**

Comprehensive guide to building and deploying LLM-powered agents in Go-LLMs, covering agent types, tool integration, prompt engineering, conversation management, function calling, and advanced AI agent patterns.

## LLM Agent Architecture

### Core LLM Agent Interface

```go
// LLMAgent extends the base Agent interface with LLM-specific capabilities
type LLMAgent interface {
    Agent
    
    // Core LLM operations
    Complete(ctx context.Context, request *CompletionRequest) (*CompletionResponse, error)
    CompleteStream(ctx context.Context, request *CompletionRequest) (<-chan StreamChunk, error)
    
    // Provider management
    GetProvider() provider.Provider
    SetProvider(provider provider.Provider) error
    
    // Tool management
    GetTools() []domain.Tool
    AddTool(tool domain.Tool) error
    RemoveTool(name string) error
    ListAvailableTools() []ToolInfo
    
    // Prompt management
    SetSystemPrompt(prompt string)
    GetSystemPrompt() string
    AddPromptTemplate(name string, template *PromptTemplate) error
    GetPromptTemplate(name string) (*PromptTemplate, error)
    
    // Conversation management
    GetConversationHistory() []core.Message
    ClearConversationHistory()
    SetMaxHistoryLength(length int)
    
    // Function calling
    EnableFunctionCalling(enabled bool)
    IsFunctionCallingEnabled() bool
    RegisterFunction(fn Function) error
    UnregisterFunction(name string) error
    
    // Memory and context
    GetMemory() AgentMemory
    SetMemory(memory AgentMemory)
    GetContextWindow() int
    
    // Configuration
    SetTemperature(temperature float64)
    SetMaxTokens(maxTokens int)
    SetTopP(topP float64)
    GetModelConfig() ModelConfig
}

// LLMAgentConfig configures LLM agent behavior
type LLMAgentConfig struct {
    BaseConfig     AgentConfig                `yaml:"base" json:"base"`
    Provider       ProviderConfig             `yaml:"provider" json:"provider"`
    Model          ModelConfig                `yaml:"model" json:"model"`
    SystemPrompt   string                     `yaml:"system_prompt,omitempty" json:"system_prompt,omitempty"`
    Tools          []string                   `yaml:"tools,omitempty" json:"tools,omitempty"`
    Functions      []FunctionConfig           `yaml:"functions,omitempty" json:"functions,omitempty"`
    Memory         MemoryConfig               `yaml:"memory,omitempty" json:"memory,omitempty"`
    Conversation   ConversationConfig         `yaml:"conversation,omitempty" json:"conversation,omitempty"`
    Templates      map[string]*PromptTemplate `yaml:"templates,omitempty" json:"templates,omitempty"`
}

type ModelConfig struct {
    Name         string    `yaml:"name" json:"name"`
    Temperature  *float64  `yaml:"temperature,omitempty" json:"temperature,omitempty"`
    MaxTokens    *int      `yaml:"max_tokens,omitempty" json:"max_tokens,omitempty"`
    TopP         *float64  `yaml:"top_p,omitempty" json:"top_p,omitempty"`
    TopK         *int      `yaml:"top_k,omitempty" json:"top_k,omitempty"`
    StopSequences []string `yaml:"stop_sequences,omitempty" json:"stop_sequences,omitempty"`
    ResponseFormat *ResponseFormat `yaml:"response_format,omitempty" json:"response_format,omitempty"`
}

type ConversationConfig struct {
    MaxHistory    int           `yaml:"max_history,omitempty" json:"max_history,omitempty"`
    PersistHistory bool         `yaml:"persist_history,omitempty" json:"persist_history,omitempty"`
    ContextCompression bool     `yaml:"context_compression,omitempty" json:"context_compression,omitempty"`
    MemoryStrategy string       `yaml:"memory_strategy,omitempty" json:"memory_strategy,omitempty"`
}
```

### Basic LLM Agent Implementation

```go
// DefaultLLMAgent implements the LLMAgent interface
type DefaultLLMAgent struct {
    *BaseAgent
    
    // LLM provider
    provider provider.Provider
    
    // Configuration
    modelConfig  ModelConfig
    systemPrompt string
    
    // Tools and functions
    tools     map[string]domain.Tool
    functions map[string]Function
    
    // Conversation management
    history       []core.Message
    maxHistory    int
    memory        AgentMemory
    
    // Prompt templates
    templates map[string]*PromptTemplate
    
    // Function calling
    functionCallingEnabled bool
    
    // Metrics
    requestCount     int64
    totalTokens      int64
    averageLatency   time.Duration
    errorRate        float64
}

func NewLLMAgent(name string, provider provider.Provider, opts ...LLMAgentOption) *DefaultLLMAgent {
    agent := &DefaultLLMAgent{
        BaseAgent:   NewBaseAgent(name, AgentTypeLLM),
        provider:    provider,
        tools:       make(map[string]domain.Tool),
        functions:   make(map[string]Function),
        history:     make([]core.Message, 0),
        maxHistory:  50,
        templates:   make(map[string]*PromptTemplate),
        modelConfig: ModelConfig{
            Temperature: &[]float64{0.7}[0],
            MaxTokens:   &[]int{1000}[0],
        },
        functionCallingEnabled: true,
    }
    
    // Apply options
    for _, opt := range opts {
        opt(agent)
    }
    
    return agent
}

// LLMAgentOption configures LLM agent
type LLMAgentOption func(*DefaultLLMAgent)

func WithSystemPrompt(prompt string) LLMAgentOption {
    return func(a *DefaultLLMAgent) {
        a.systemPrompt = prompt
    }
}

func WithTools(tools ...domain.Tool) LLMAgentOption {
    return func(a *DefaultLLMAgent) {
        for _, tool := range tools {
            a.tools[tool.Name()] = tool
        }
    }
}

func WithModelConfig(config ModelConfig) LLMAgentOption {
    return func(a *DefaultLLMAgent) {
        a.modelConfig = config
    }
}

func WithMemory(memory AgentMemory) LLMAgentOption {
    return func(a *DefaultLLMAgent) {
        a.memory = memory
    }
}

func WithMaxHistory(maxHistory int) LLMAgentOption {
    return func(a *DefaultLLMAgent) {
        a.maxHistory = maxHistory
    }
}
```

---

## Tool Integration

### Tool Management

```go
// Tool integration for LLM agents
func (a *DefaultLLMAgent) AddTool(tool domain.Tool) error {
    a.mu.Lock()
    defer a.mu.Unlock()
    
    // Validate tool
    if err := a.validateTool(tool); err != nil {
        return fmt.Errorf("invalid tool: %w", err)
    }
    
    // Register tool
    a.tools[tool.Name()] = tool
    
    a.logger.Info("Tool added to agent",
        zap.String("agent_id", a.ID()),
        zap.String("tool_name", tool.Name()),
        zap.String("tool_description", tool.Description()),
    )
    
    return nil
}

func (a *DefaultLLMAgent) validateTool(tool domain.Tool) error {
    // Check required methods
    if tool.Name() == "" {
        return errors.New("tool name cannot be empty")
    }
    
    if tool.Description() == "" {
        return errors.New("tool description cannot be empty")
    }
    
    // Validate schema
    inputSchema := tool.InputSchema()
    if inputSchema == nil {
        return errors.New("tool must provide input schema")
    }
    
    outputSchema := tool.OutputSchema()
    if outputSchema == nil {
        return errors.New("tool must provide output schema")
    }
    
    return nil
}

func (a *DefaultLLMAgent) RemoveTool(name string) error {
    a.mu.Lock()
    defer a.mu.Unlock()
    
    if _, exists := a.tools[name]; !exists {
        return fmt.Errorf("tool %s not found", name)
    }
    
    delete(a.tools, name)
    
    a.logger.Info("Tool removed from agent",
        zap.String("agent_id", a.ID()),
        zap.String("tool_name", name),
    )
    
    return nil
}

func (a *DefaultLLMAgent) GetTools() []domain.Tool {
    a.mu.RLock()
    defer a.mu.RUnlock()
    
    tools := make([]domain.Tool, 0, len(a.tools))
    for _, tool := range a.tools {
        tools = append(tools, tool)
    }
    
    return tools
}

func (a *DefaultLLMAgent) ListAvailableTools() []ToolInfo {
    a.mu.RLock()
    defer a.mu.RUnlock()
    
    toolInfos := make([]ToolInfo, 0, len(a.tools))
    for _, tool := range a.tools {
        toolInfos = append(toolInfos, ToolInfo{
            Name:         tool.Name(),
            Description:  tool.Description(),
            InputSchema:  tool.InputSchema(),
            OutputSchema: tool.OutputSchema(),
            Available:    true,
        })
    }
    
    return toolInfos
}

type ToolInfo struct {
    Name         string      `json:"name"`
    Description  string      `json:"description"`
    InputSchema  interface{} `json:"input_schema"`
    OutputSchema interface{} `json:"output_schema"`
    Available    bool        `json:"available"`
    LastUsed     *time.Time  `json:"last_used,omitempty"`
    UsageCount   int64       `json:"usage_count"`
    ErrorCount   int64       `json:"error_count"`
}
```

### Tool Execution

```go
// executeTool handles tool execution within agent context
func (a *DefaultLLMAgent) executeTool(ctx context.Context, toolName string, input interface{}) (interface{}, error) {
    tool, exists := a.tools[toolName]
    if !exists {
        return nil, fmt.Errorf("tool %s not available", toolName)
    }
    
    // Record metrics
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        a.recordToolUsage(toolName, duration)
    }()
    
    // Execute tool with context
    result, err := tool.Execute(ctx, input)
    if err != nil {
        a.recordToolError(toolName, err)
        return nil, fmt.Errorf("tool execution failed: %w", err)
    }
    
    a.logger.Debug("Tool executed successfully",
        zap.String("agent_id", a.ID()),
        zap.String("tool_name", toolName),
        zap.Any("input", input),
        zap.Any("result", result),
    )
    
    return result, nil
}

func (a *DefaultLLMAgent) recordToolUsage(toolName string, duration time.Duration) {
    // Update metrics
    if a.metrics != nil {
        a.metrics.RecordToolUsage(toolName, duration)
    }
    
    // Update tool info
    // Implementation depends on tool registry
}

func (a *DefaultLLMAgent) recordToolError(toolName string, err error) {
    // Record error metrics
    if a.metrics != nil {
        a.metrics.RecordToolError(toolName, err)
    }
    
    a.logger.Error("Tool execution error",
        zap.String("agent_id", a.ID()),
        zap.String("tool_name", toolName),
        zap.Error(err),
    )
}

// ToolExecutionContext provides context for tool execution
type ToolExecutionContext struct {
    AgentID     string                 `json:"agent_id"`
    RequestID   string                 `json:"request_id"`
    UserID      string                 `json:"user_id,omitempty"`
    SessionID   string                 `json:"session_id,omitempty"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
    Timeout     time.Duration          `json:"timeout,omitempty"`
    Permissions []string               `json:"permissions,omitempty"`
}

// Enhanced tool execution with context
func (a *DefaultLLMAgent) ExecuteToolWithContext(ctx context.Context, toolName string, input interface{}, execCtx ToolExecutionContext) (interface{}, error) {
    // Check permissions
    if err := a.checkToolPermissions(toolName, execCtx.Permissions); err != nil {
        return nil, fmt.Errorf("permission denied: %w", err)
    }
    
    // Set timeout if specified
    if execCtx.Timeout > 0 {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(ctx, execCtx.Timeout)
        defer cancel()
    }
    
    // Add execution context to context
    ctx = context.WithValue(ctx, "tool_execution_context", execCtx)
    
    return a.executeTool(ctx, toolName, input)
}

func (a *DefaultLLMAgent) checkToolPermissions(toolName string, permissions []string) error {
    // Implementation depends on security requirements
    // This is a placeholder for permission checking logic
    return nil
}
```

---

## Function Calling

### Function Management

```go
// Function represents a callable function for LLM agents
type Function struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Parameters  *jsonschema.Schema     `json:"parameters"`
    Handler     FunctionHandler        `json:"-"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type FunctionHandler func(ctx context.Context, args map[string]interface{}) (interface{}, error)

// Function calling implementation
func (a *DefaultLLMAgent) RegisterFunction(fn Function) error {
    a.mu.Lock()
    defer a.mu.Unlock()
    
    // Validate function
    if err := a.validateFunction(fn); err != nil {
        return fmt.Errorf("invalid function: %w", err)
    }
    
    a.functions[fn.Name] = fn
    
    a.logger.Info("Function registered",
        zap.String("agent_id", a.ID()),
        zap.String("function_name", fn.Name),
        zap.String("function_description", fn.Description),
    )
    
    return nil
}

func (a *DefaultLLMAgent) validateFunction(fn Function) error {
    if fn.Name == "" {
        return errors.New("function name cannot be empty")
    }
    
    if fn.Description == "" {
        return errors.New("function description cannot be empty")
    }
    
    if fn.Parameters == nil {
        return errors.New("function parameters schema required")
    }
    
    if fn.Handler == nil {
        return errors.New("function handler required")
    }
    
    return nil
}

func (a *DefaultLLMAgent) executeFunction(ctx context.Context, functionName string, args map[string]interface{}) (interface{}, error) {
    function, exists := a.functions[functionName]
    if !exists {
        return nil, fmt.Errorf("function %s not found", functionName)
    }
    
    // Validate arguments against schema
    if err := a.validateFunctionArgs(args, function.Parameters); err != nil {
        return nil, fmt.Errorf("invalid function arguments: %w", err)
    }
    
    // Execute function
    result, err := function.Handler(ctx, args)
    if err != nil {
        return nil, fmt.Errorf("function execution failed: %w", err)
    }
    
    return result, nil
}

func (a *DefaultLLMAgent) validateFunctionArgs(args map[string]interface{}, schema *jsonschema.Schema) error {
    // Implement JSON schema validation
    // This is a simplified version
    return nil
}

// Built-in function examples
func (a *DefaultLLMAgent) registerBuiltinFunctions() {
    // Get current time function
    a.RegisterFunction(Function{
        Name:        "get_current_time",
        Description: "Get the current date and time",
        Parameters: &jsonschema.Schema{
            Type: "object",
            Properties: map[string]*jsonschema.Schema{
                "timezone": {
                    Type:        "string",
                    Description: "Timezone (optional, defaults to UTC)",
                    Default:     "UTC",
                },
                "format": {
                    Type:        "string",
                    Description: "Time format (optional)",
                    Default:     "2006-01-02 15:04:05",
                },
            },
        },
        Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
            timezone := "UTC"
            format := "2006-01-02 15:04:05"
            
            if tz, ok := args["timezone"].(string); ok {
                timezone = tz
            }
            
            if fmt, ok := args["format"].(string); ok {
                format = fmt
            }
            
            loc, err := time.LoadLocation(timezone)
            if err != nil {
                return nil, fmt.Errorf("invalid timezone: %w", err)
            }
            
            return time.Now().In(loc).Format(format), nil
        },
    })
    
    // Calculate function
    a.RegisterFunction(Function{
        Name:        "calculate",
        Description: "Perform mathematical calculations",
        Parameters: &jsonschema.Schema{
            Type: "object",
            Properties: map[string]*jsonschema.Schema{
                "expression": {
                    Type:        "string",
                    Description: "Mathematical expression to evaluate",
                },
            },
            Required: []string{"expression"},
        },
        Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
            expression, ok := args["expression"].(string)
            if !ok {
                return nil, errors.New("expression must be a string")
            }
            
            // Simple calculator implementation
            result, err := evaluateExpression(expression)
            if err != nil {
                return nil, fmt.Errorf("calculation error: %w", err)
            }
            
            return result, nil
        },
    })
}

// Simple expression evaluator (placeholder)
func evaluateExpression(expression string) (float64, error) {
    // This would use a proper math expression parser
    // For now, just return a placeholder
    return 42.0, nil
}
```

### Function Calling in Completions

```go
// Enhanced completion with function calling
func (a *DefaultLLMAgent) Complete(ctx context.Context, request *CompletionRequest) (*CompletionResponse, error) {
    // Prepare request with functions
    enhancedRequest := a.prepareCompletionRequest(request)
    
    // Execute completion
    response, err := a.provider.Complete(ctx, enhancedRequest)
    if err != nil {
        return nil, err
    }
    
    // Handle function calls if present
    if a.functionCallingEnabled && response.FunctionCall != nil {
        return a.handleFunctionCall(ctx, response)
    }
    
    // Add to conversation history
    a.addToHistory(request.Messages...)
    a.addToHistory(core.Message{
        Role:    "assistant",
        Content: response.Content,
    })
    
    return response, nil
}

func (a *DefaultLLMAgent) prepareCompletionRequest(request *CompletionRequest) *CompletionRequest {
    enhanced := *request // Copy request
    
    // Add system prompt if not present
    if len(enhanced.Messages) == 0 || enhanced.Messages[0].Role != "system" {
        systemMsg := core.Message{
            Role:    "system",
            Content: a.getEffectiveSystemPrompt(),
        }
        enhanced.Messages = append([]core.Message{systemMsg}, enhanced.Messages...)
    }
    
    // Add conversation history
    if len(a.history) > 0 {
        // Insert history before the last user message
        if len(enhanced.Messages) > 0 && enhanced.Messages[len(enhanced.Messages)-1].Role == "user" {
            lastMsg := enhanced.Messages[len(enhanced.Messages)-1]
            enhanced.Messages = enhanced.Messages[:len(enhanced.Messages)-1]
            enhanced.Messages = append(enhanced.Messages, a.history...)
            enhanced.Messages = append(enhanced.Messages, lastMsg)
        } else {
            enhanced.Messages = append(enhanced.Messages, a.history...)
        }
    }
    
    // Add function definitions if function calling is enabled
    if a.functionCallingEnabled && len(a.functions) > 0 {
        enhanced.Functions = a.getFunctionDefinitions()
    }
    
    // Apply model configuration
    a.applyModelConfig(&enhanced)
    
    return &enhanced
}

func (a *DefaultLLMAgent) getFunctionDefinitions() []FunctionDefinition {
    definitions := make([]FunctionDefinition, 0, len(a.functions))
    
    for _, function := range a.functions {
        definitions = append(definitions, FunctionDefinition{
            Name:        function.Name,
            Description: function.Description,
            Parameters:  function.Parameters,
        })
    }
    
    return definitions
}

func (a *DefaultLLMAgent) handleFunctionCall(ctx context.Context, response *CompletionResponse) (*CompletionResponse, error) {
    functionCall := response.FunctionCall
    
    a.logger.Info("Function call requested",
        zap.String("agent_id", a.ID()),
        zap.String("function_name", functionCall.Name),
        zap.Any("arguments", functionCall.Arguments),
    )
    
    // Execute function
    result, err := a.executeFunction(ctx, functionCall.Name, functionCall.Arguments)
    if err != nil {
        return nil, fmt.Errorf("function call failed: %w", err)
    }
    
    // Create function result message
    functionResult := core.Message{
        Role:         "function",
        Content:      fmt.Sprintf("%v", result),
        FunctionCall: functionCall,
    }
    
    // Add to history
    a.addToHistory(core.Message{
        Role:         "assistant",
        Content:      response.Content,
        FunctionCall: functionCall,
    })
    a.addToHistory(functionResult)
    
    // Continue conversation with function result
    continueRequest := &CompletionRequest{
        Messages: []core.Message{functionResult},
    }
    
    return a.Complete(ctx, continueRequest)
}

func (a *DefaultLLMAgent) applyModelConfig(request *CompletionRequest) {
    if a.modelConfig.Name != "" {
        request.Model = a.modelConfig.Name
    }
    
    if a.modelConfig.Temperature != nil {
        request.Temperature = *a.modelConfig.Temperature
    }
    
    if a.modelConfig.MaxTokens != nil {
        request.MaxTokens = *a.modelConfig.MaxTokens
    }
    
    if a.modelConfig.TopP != nil {
        request.TopP = *a.modelConfig.TopP
    }
    
    if len(a.modelConfig.StopSequences) > 0 {
        request.Stop = a.modelConfig.StopSequences
    }
    
    if a.modelConfig.ResponseFormat != nil {
        request.ResponseFormat = a.modelConfig.ResponseFormat
    }
}
```

---

## Prompt Engineering

### Prompt Templates

```go
// PromptTemplate represents a reusable prompt template
type PromptTemplate struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Template    string                 `json:"template"`
    Variables   []TemplateVariable     `json:"variables"`
    Examples    []PromptExample        `json:"examples,omitempty"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type TemplateVariable struct {
    Name        string      `json:"name"`
    Type        string      `json:"type"` // "string", "number", "boolean", "array", "object"
    Description string      `json:"description"`
    Required    bool        `json:"required"`
    Default     interface{} `json:"default,omitempty"`
    Validation  *VariableValidation `json:"validation,omitempty"`
}

type VariableValidation struct {
    MinLength *int     `json:"min_length,omitempty"`
    MaxLength *int     `json:"max_length,omitempty"`
    Pattern   *string  `json:"pattern,omitempty"`
    Enum      []string `json:"enum,omitempty"`
}

type PromptExample struct {
    Variables map[string]interface{} `json:"variables"`
    Expected  string                 `json:"expected"`
}

// Prompt template engine
type PromptTemplateEngine struct {
    templates map[string]*PromptTemplate
    functions map[string]TemplateFunction
    mu        sync.RWMutex
}

type TemplateFunction func(args ...interface{}) (string, error)

func NewPromptTemplateEngine() *PromptTemplateEngine {
    engine := &PromptTemplateEngine{
        templates: make(map[string]*PromptTemplate),
        functions: make(map[string]TemplateFunction),
    }
    
    // Register built-in functions
    engine.registerBuiltinFunctions()
    
    return engine
}

func (e *PromptTemplateEngine) RegisterTemplate(template *PromptTemplate) error {
    e.mu.Lock()
    defer e.mu.Unlock()
    
    if err := e.validateTemplate(template); err != nil {
        return fmt.Errorf("invalid template: %w", err)
    }
    
    e.templates[template.Name] = template
    return nil
}

func (e *PromptTemplateEngine) Render(templateName string, variables map[string]interface{}) (string, error) {
    e.mu.RLock()
    template, exists := e.templates[templateName]
    e.mu.RUnlock()
    
    if !exists {
        return "", fmt.Errorf("template %s not found", templateName)
    }
    
    // Validate variables
    if err := e.validateVariables(template, variables); err != nil {
        return "", fmt.Errorf("variable validation failed: %w", err)
    }
    
    // Render template
    return e.renderTemplate(template.Template, variables)
}

func (e *PromptTemplateEngine) renderTemplate(template string, variables map[string]interface{}) (string, error) {
    // Use Go's text/template for rendering
    tmpl := texttemplate.New("prompt")
    
    // Add custom functions
    tmpl.Funcs(texttemplate.FuncMap(e.functions))
    
    // Parse template
    tmpl, err := tmpl.Parse(template)
    if err != nil {
        return "", fmt.Errorf("template parse error: %w", err)
    }
    
    // Execute template
    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, variables); err != nil {
        return "", fmt.Errorf("template execution error: %w", err)
    }
    
    return buf.String(), nil
}

func (e *PromptTemplateEngine) registerBuiltinFunctions() {
    e.functions["upper"] = func(args ...interface{}) (string, error) {
        if len(args) != 1 {
            return "", errors.New("upper requires exactly one argument")
        }
        
        str, ok := args[0].(string)
        if !ok {
            return "", errors.New("upper requires string argument")
        }
        
        return strings.ToUpper(str), nil
    }
    
    e.functions["lower"] = func(args ...interface{}) (string, error) {
        if len(args) != 1 {
            return "", errors.New("lower requires exactly one argument")
        }
        
        str, ok := args[0].(string)
        if !ok {
            return "", errors.New("lower requires string argument")
        }
        
        return strings.ToLower(str), nil
    }
    
    e.functions["format_list"] = func(args ...interface{}) (string, error) {
        if len(args) == 0 {
            return "", nil
        }
        
        var items []string
        for _, arg := range args {
            items = append(items, fmt.Sprintf("- %v", arg))
        }
        
        return strings.Join(items, "\n"), nil
    }
}

// Prompt template examples
func CreateDocumentAnalysisTemplate() *PromptTemplate {
    return &PromptTemplate{
        Name:        "document_analysis",
        Description: "Analyze a document and extract key information",
        Template: `You are an expert document analyst. Please analyze the following document and extract the key information.

Document Type: {{.document_type}}
Document Content:
{{.content}}

Please provide your analysis in the following format:

## Summary
{{if .include_summary}}
Provide a brief summary of the document.
{{end}}

## Key Points
Extract the main points from the document.

## Entities
{{if .extract_entities}}
List important entities mentioned (people, organizations, dates, etc.).
{{end}}

## Recommendations
{{if .include_recommendations}}
Provide any recommendations based on the analysis.
{{end}}

Focus on {{.focus_areas | format_list}} and ensure accuracy in your analysis.`,
        
        Variables: []TemplateVariable{
            {
                Name:        "document_type",
                Type:        "string",
                Description: "Type of document being analyzed",
                Required:    true,
                Validation: &VariableValidation{
                    Enum: []string{"contract", "report", "email", "article", "manual"},
                },
            },
            {
                Name:        "content",
                Type:        "string",
                Description: "Document content to analyze",
                Required:    true,
                Validation: &VariableValidation{
                    MinLength: &[]int{10}[0],
                    MaxLength: &[]int{50000}[0],
                },
            },
            {
                Name:        "include_summary",
                Type:        "boolean",
                Description: "Whether to include a summary section",
                Required:    false,
                Default:     true,
            },
            {
                Name:        "extract_entities",
                Type:        "boolean",
                Description: "Whether to extract entities",
                Required:    false,
                Default:     false,
            },
            {
                Name:        "include_recommendations",
                Type:        "boolean",
                Description: "Whether to include recommendations",
                Required:    false,
                Default:     false,
            },
            {
                Name:        "focus_areas",
                Type:        "array",
                Description: "Specific areas to focus on",
                Required:    false,
                Default:     []string{"accuracy", "completeness"},
            },
        },
        
        Examples: []PromptExample{
            {
                Variables: map[string]interface{}{
                    "document_type":          "contract",
                    "content":                "This is a sample contract...",
                    "include_summary":        true,
                    "extract_entities":       true,
                    "include_recommendations": false,
                    "focus_areas":            []string{"legal obligations", "payment terms"},
                },
                Expected: "Rendered template with contract analysis focus",
            },
        },
    }
}
```

### System Prompt Management

```go
// Advanced system prompt management
func (a *DefaultLLMAgent) getEffectiveSystemPrompt() string {
    var promptParts []string
    
    // Base system prompt
    if a.systemPrompt != "" {
        promptParts = append(promptParts, a.systemPrompt)
    }
    
    // Tool descriptions
    if len(a.tools) > 0 {
        toolsSection := a.generateToolsSection()
        promptParts = append(promptParts, toolsSection)
    }
    
    // Function descriptions
    if a.functionCallingEnabled && len(a.functions) > 0 {
        functionsSection := a.generateFunctionsSection()
        promptParts = append(promptParts, functionsSection)
    }
    
    // Context and guidelines
    contextSection := a.generateContextSection()
    promptParts = append(promptParts, contextSection)
    
    return strings.Join(promptParts, "\n\n")
}

func (a *DefaultLLMAgent) generateToolsSection() string {
    var sb strings.Builder
    sb.WriteString("You have access to the following tools:")
    
    for _, tool := range a.tools {
        sb.WriteString(fmt.Sprintf("\n- %s: %s", tool.Name(), tool.Description()))
    }
    
    sb.WriteString("\n\nTo use a tool, specify the tool name and provide the required parameters.")
    
    return sb.String()
}

func (a *DefaultLLMAgent) generateFunctionsSection() string {
    var sb strings.Builder
    sb.WriteString("You can call the following functions:")
    
    for _, function := range a.functions {
        sb.WriteString(fmt.Sprintf("\n- %s: %s", function.Name, function.Description))
    }
    
    sb.WriteString("\n\nCall functions when you need to perform specific actions or get current information.")
    
    return sb.String()
}

func (a *DefaultLLMAgent) generateContextSection() string {
    return `Guidelines:
1. Be helpful, accurate, and concise in your responses
2. Use tools and functions when appropriate to provide better assistance
3. If you're unsure about something, ask for clarification
4. Always maintain a professional and friendly tone
5. Provide step-by-step explanations for complex tasks`
}

// Dynamic prompt modification
type PromptModifier interface {
    ModifyPrompt(originalPrompt string, context PromptContext) string
    Priority() int
}

type PromptContext struct {
    UserID        string                 `json:"user_id,omitempty"`
    SessionID     string                 `json:"session_id,omitempty"`
    Conversation  []core.Message         `json:"conversation,omitempty"`
    Metadata      map[string]interface{} `json:"metadata,omitempty"`
    RequestType   string                 `json:"request_type,omitempty"`
}

// Personality modifier
type PersonalityModifier struct {
    personality string
    priority    int
}

func NewPersonalityModifier(personality string) *PersonalityModifier {
    return &PersonalityModifier{
        personality: personality,
        priority:    10,
    }
}

func (m *PersonalityModifier) ModifyPrompt(originalPrompt string, context PromptContext) string {
    personalityPrompt := fmt.Sprintf("You should respond with a %s personality. ", m.personality)
    return personalityPrompt + originalPrompt
}

func (m *PersonalityModifier) Priority() int {
    return m.priority
}

// Context-aware modifier
type ContextAwareModifier struct {
    contextRules map[string]string
}

func NewContextAwareModifier(rules map[string]string) *ContextAwareModifier {
    return &ContextAwareModifier{
        contextRules: rules,
    }
}

func (m *ContextAwareModifier) ModifyPrompt(originalPrompt string, context PromptContext) string {
    if rule, exists := m.contextRules[context.RequestType]; exists {
        return rule + "\n\n" + originalPrompt
    }
    
    return originalPrompt
}

func (m *ContextAwareModifier) Priority() int {
    return 5
}
```

---

## Conversation Management

### Conversation History

```go
// Advanced conversation management
func (a *DefaultLLMAgent) addToHistory(messages ...core.Message) {
    a.mu.Lock()
    defer a.mu.Unlock()
    
    for _, message := range messages {
        a.history = append(a.history, message)
    }
    
    // Trim history if it exceeds max length
    if len(a.history) > a.maxHistory {
        // Keep recent messages and important context
        a.history = a.smartTrimHistory(a.history, a.maxHistory)
    }
    
    // Persist history if configured
    if a.memory != nil {
        a.memory.StoreMessages(messages...)
    }
}

func (a *DefaultLLMAgent) smartTrimHistory(history []core.Message, maxLength int) []core.Message {
    if len(history) <= maxLength {
        return history
    }
    
    // Always keep system messages
    var systemMessages []core.Message
    var otherMessages []core.Message
    
    for _, msg := range history {
        if msg.Role == "system" {
            systemMessages = append(systemMessages, msg)
        } else {
            otherMessages = append(otherMessages, msg)
        }
    }
    
    // Calculate available space for non-system messages
    availableSpace := maxLength - len(systemMessages)
    
    if availableSpace <= 0 {
        return systemMessages
    }
    
    // Keep the most recent messages
    if len(otherMessages) > availableSpace {
        otherMessages = otherMessages[len(otherMessages)-availableSpace:]
    }
    
    // Combine system messages and recent messages
    result := append(systemMessages, otherMessages...)
    
    return result
}

func (a *DefaultLLMAgent) GetConversationHistory() []core.Message {
    a.mu.RLock()
    defer a.mu.RUnlock()
    
    // Return a copy to prevent external modification
    history := make([]core.Message, len(a.history))
    copy(history, a.history)
    
    return history
}

func (a *DefaultLLMAgent) ClearConversationHistory() {
    a.mu.Lock()
    defer a.mu.Unlock()
    
    a.history = make([]core.Message, 0)
    
    if a.memory != nil {
        a.memory.Clear()
    }
    
    a.logger.Info("Conversation history cleared",
        zap.String("agent_id", a.ID()),
    )
}

func (a *DefaultLLMAgent) SetMaxHistoryLength(length int) {
    a.mu.Lock()
    defer a.mu.Unlock()
    
    a.maxHistory = length
    
    // Trim current history if necessary
    if len(a.history) > length {
        a.history = a.smartTrimHistory(a.history, length)
    }
}
```

### Agent Memory

```go
// AgentMemory provides persistent memory for agents
type AgentMemory interface {
    // Message storage
    StoreMessages(messages ...core.Message) error
    GetMessages(limit int) ([]core.Message, error)
    SearchMessages(query string) ([]core.Message, error)
    
    // Key-value storage
    Store(key string, value interface{}) error
    Retrieve(key string) (interface{}, error)
    Delete(key string) error
    
    // Semantic memory
    StoreKnowledge(knowledge Knowledge) error
    QueryKnowledge(query string) ([]Knowledge, error)
    
    // Memory management
    Clear() error
    Compact() error
    GetStats() MemoryStats
}

type Knowledge struct {
    ID          string                 `json:"id"`
    Content     string                 `json:"content"`
    Type        string                 `json:"type"`
    Importance  float64               `json:"importance"`
    Timestamp   time.Time             `json:"timestamp"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
    Embedding   []float64             `json:"embedding,omitempty"`
}

type MemoryStats struct {
    MessageCount     int64     `json:"message_count"`
    KnowledgeCount   int64     `json:"knowledge_count"`
    StorageSize      int64     `json:"storage_size_bytes"`
    LastAccessed     time.Time `json:"last_accessed"`
    LastCompacted    time.Time `json:"last_compacted"`
}

// In-memory implementation
type InMemoryAgentMemory struct {
    messages    []core.Message
    keyValue    map[string]interface{}
    knowledge   []Knowledge
    mu          sync.RWMutex
    maxSize     int
    logger      *zap.Logger
}

func NewInMemoryAgentMemory(maxSize int) *InMemoryAgentMemory {
    return &InMemoryAgentMemory{
        messages:  make([]core.Message, 0),
        keyValue:  make(map[string]interface{}),
        knowledge: make([]Knowledge, 0),
        maxSize:   maxSize,
        logger:    zap.NewNop(),
    }
}

func (m *InMemoryAgentMemory) StoreMessages(messages ...core.Message) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    m.messages = append(m.messages, messages...)
    
    // Trim if necessary
    if len(m.messages) > m.maxSize {
        excess := len(m.messages) - m.maxSize
        m.messages = m.messages[excess:]
    }
    
    return nil
}

func (m *InMemoryAgentMemory) GetMessages(limit int) ([]core.Message, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    if limit <= 0 || limit > len(m.messages) {
        limit = len(m.messages)
    }
    
    // Return most recent messages
    start := len(m.messages) - limit
    result := make([]core.Message, limit)
    copy(result, m.messages[start:])
    
    return result, nil
}

func (m *InMemoryAgentMemory) SearchMessages(query string) ([]core.Message, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    var results []core.Message
    queryLower := strings.ToLower(query)
    
    for _, message := range m.messages {
        if strings.Contains(strings.ToLower(message.Content), queryLower) {
            results = append(results, message)
        }
    }
    
    return results, nil
}

func (m *InMemoryAgentMemory) Store(key string, value interface{}) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    m.keyValue[key] = value
    return nil
}

func (m *InMemoryAgentMemory) Retrieve(key string) (interface{}, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    value, exists := m.keyValue[key]
    if !exists {
        return nil, fmt.Errorf("key %s not found", key)
    }
    
    return value, nil
}

func (m *InMemoryAgentMemory) StoreKnowledge(knowledge Knowledge) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    // Add timestamp if not set
    if knowledge.Timestamp.IsZero() {
        knowledge.Timestamp = time.Now()
    }
    
    // Generate ID if not set
    if knowledge.ID == "" {
        knowledge.ID = generateKnowledgeID(knowledge)
    }
    
    m.knowledge = append(m.knowledge, knowledge)
    
    return nil
}

func (m *InMemoryAgentMemory) QueryKnowledge(query string) ([]Knowledge, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    var results []Knowledge
    queryLower := strings.ToLower(query)
    
    for _, k := range m.knowledge {
        if strings.Contains(strings.ToLower(k.Content), queryLower) {
            results = append(results, k)
        }
    }
    
    // Sort by importance and recency
    sort.Slice(results, func(i, j int) bool {
        if results[i].Importance != results[j].Importance {
            return results[i].Importance > results[j].Importance
        }
        return results[i].Timestamp.After(results[j].Timestamp)
    })
    
    return results, nil
}

func generateKnowledgeID(knowledge Knowledge) string {
    content := fmt.Sprintf("%s-%s-%d", knowledge.Content, knowledge.Type, knowledge.Timestamp.Unix())
    hash := sha256.Sum256([]byte(content))
    return fmt.Sprintf("%x", hash[:8])
}
```

---

## Advanced LLM Agent Patterns

### Multi-Step Reasoning Agent

```go
// ReasoningAgent implements step-by-step reasoning
type ReasoningAgent struct {
    *DefaultLLMAgent
    reasoningSteps []ReasoningStep
    maxSteps       int
    validator      ResponseValidator
}

type ReasoningStep struct {
    Name        string `json:"name"`
    Description string `json:"description"`
    Prompt      string `json:"prompt"`
    Required    bool   `json:"required"`
}

type ResponseValidator interface {
    Validate(response string) (bool, string, error)
}

func NewReasoningAgent(name string, provider provider.Provider, steps []ReasoningStep) *ReasoningAgent {
    return &ReasoningAgent{
        DefaultLLMAgent: NewLLMAgent(name, provider),
        reasoningSteps:  steps,
        maxSteps:        len(steps),
        validator:       NewDefaultResponseValidator(),
    }
}

func (a *ReasoningAgent) ExecuteReasoning(ctx context.Context, query string) (*ReasoningResult, error) {
    result := &ReasoningResult{
        Query:     query,
        Steps:     make([]StepResult, 0),
        StartTime: time.Now(),
    }
    
    currentContext := query
    
    for i, step := range a.reasoningSteps {
        stepResult, err := a.executeReasoningStep(ctx, step, currentContext)
        if err != nil {
            if step.Required {
                return nil, fmt.Errorf("required step %s failed: %w", step.Name, err)
            }
            
            // Skip optional step
            stepResult = &StepResult{
                Name:    step.Name,
                Skipped: true,
                Error:   err.Error(),
            }
        }
        
        result.Steps = append(result.Steps, *stepResult)
        
        // Update context for next step
        if stepResult.Output != "" {
            currentContext = stepResult.Output
        }
        
        // Check for early termination
        if i < len(a.reasoningSteps)-1 && a.shouldTerminateEarly(stepResult) {
            break
        }
    }
    
    result.EndTime = time.Now()
    result.Duration = result.EndTime.Sub(result.StartTime)
    
    return result, nil
}

func (a *ReasoningAgent) executeReasoningStep(ctx context.Context, step ReasoningStep, context string) (*StepResult, error) {
    start := time.Now()
    
    // Prepare prompt for this step
    prompt := fmt.Sprintf("%s\n\nContext: %s", step.Prompt, context)
    
    // Execute completion
    response, err := a.Complete(ctx, &CompletionRequest{
        Messages: []core.Message{
            {Role: "user", Content: prompt},
        },
    })
    if err != nil {
        return nil, err
    }
    
    // Validate response
    valid, feedback, err := a.validator.Validate(response.Content)
    if err != nil {
        return nil, fmt.Errorf("validation error: %w", err)
    }
    
    return &StepResult{
        Name:       step.Name,
        Input:      context,
        Output:     response.Content,
        Duration:   time.Since(start),
        Valid:      valid,
        Feedback:   feedback,
        TokenUsage: response.Usage,
    }, nil
}

type ReasoningResult struct {
    Query     string        `json:"query"`
    Steps     []StepResult  `json:"steps"`
    StartTime time.Time     `json:"start_time"`
    EndTime   time.Time     `json:"end_time"`
    Duration  time.Duration `json:"duration"`
}

type StepResult struct {
    Name       string       `json:"name"`
    Input      string       `json:"input"`
    Output     string       `json:"output"`
    Duration   time.Duration `json:"duration"`
    Valid      bool         `json:"valid"`
    Feedback   string       `json:"feedback,omitempty"`
    Skipped    bool         `json:"skipped,omitempty"`
    Error      string       `json:"error,omitempty"`
    TokenUsage TokenUsage   `json:"token_usage,omitempty"`
}
```

### Collaborative Agent

```go
// CollaborativeAgent works with other agents
type CollaborativeAgent struct {
    *DefaultLLMAgent
    collaborators map[string]Agent
    coordinator   AgentCoordinator
    consensus     ConsensusStrategy
}

type ConsensusStrategy interface {
    ReachConsensus(responses []AgentResponse) (AgentResponse, error)
}

func NewCollaborativeAgent(name string, provider provider.Provider, coordinator AgentCoordinator) *CollaborativeAgent {
    return &CollaborativeAgent{
        DefaultLLMAgent: NewLLMAgent(name, provider),
        collaborators:   make(map[string]Agent),
        coordinator:     coordinator,
        consensus:       NewMajorityConsensus(),
    }
}

func (a *CollaborativeAgent) AddCollaborator(agent Agent) error {
    a.mu.Lock()
    defer a.mu.Unlock()
    
    a.collaborators[agent.ID()] = agent
    
    a.logger.Info("Collaborator added",
        zap.String("agent_id", a.ID()),
        zap.String("collaborator_id", agent.ID()),
    )
    
    return nil
}

func (a *CollaborativeAgent) CollaborativeExecute(ctx context.Context, input interface{}) (*CollaborativeResult, error) {
    // Execute on all collaborators including self
    agents := make([]string, 0, len(a.collaborators)+1)
    agents = append(agents, a.ID())
    
    for id := range a.collaborators {
        agents = append(agents, id)
    }
    
    // Execute in parallel
    responses, err := a.coordinator.ExecuteParallel(ctx, agents, input)
    if err != nil {
        return nil, fmt.Errorf("collaborative execution failed: %w", err)
    }
    
    // Convert to agent responses
    agentResponses := make([]AgentResponse, len(responses))
    for i, response := range responses {
        agentResponses[i] = AgentResponse{
            ID:        fmt.Sprintf("response_%d", i),
            AgentID:   agents[i],
            Content:   response,
            Timestamp: time.Now(),
            Final:     true,
        }
    }
    
    // Reach consensus
    consensusResponse, err := a.consensus.ReachConsensus(agentResponses)
    if err != nil {
        return nil, fmt.Errorf("consensus failed: %w", err)
    }
    
    return &CollaborativeResult{
        Input:           input,
        Responses:       agentResponses,
        Consensus:       consensusResponse,
        ParticipantCount: len(agents),
        ExecutionTime:   time.Since(time.Now()), // This would be tracked properly
    }, nil
}

type CollaborativeResult struct {
    Input            interface{}     `json:"input"`
    Responses        []AgentResponse `json:"responses"`
    Consensus        AgentResponse   `json:"consensus"`
    ParticipantCount int             `json:"participant_count"`
    ExecutionTime    time.Duration   `json:"execution_time"`
}

// Majority consensus implementation
type MajorityConsensus struct{}

func NewMajorityConsensus() *MajorityConsensus {
    return &MajorityConsensus{}
}

func (c *MajorityConsensus) ReachConsensus(responses []AgentResponse) (AgentResponse, error) {
    if len(responses) == 0 {
        return AgentResponse{}, errors.New("no responses provided")
    }
    
    if len(responses) == 1 {
        return responses[0], nil
    }
    
    // Simple implementation: find most common response
    responseCounts := make(map[string]int)
    responseMap := make(map[string]AgentResponse)
    
    for _, response := range responses {
        content := fmt.Sprintf("%v", response.Content)
        responseCounts[content]++
        responseMap[content] = response
    }
    
    // Find majority
    var majorityContent string
    var maxCount int
    
    for content, count := range responseCounts {
        if count > maxCount {
            maxCount = count
            majorityContent = content
        }
    }
    
    // Check if we have a clear majority
    if maxCount <= len(responses)/2 {
        // No clear majority, create combined response
        return c.createCombinedResponse(responses), nil
    }
    
    return responseMap[majorityContent], nil
}

func (c *MajorityConsensus) createCombinedResponse(responses []AgentResponse) AgentResponse {
    var combinedContent strings.Builder
    combinedContent.WriteString("Combined response from multiple agents:\n\n")
    
    for i, response := range responses {
        combinedContent.WriteString(fmt.Sprintf("Agent %s: %v\n", response.AgentID, response.Content))
        if i < len(responses)-1 {
            combinedContent.WriteString("\n")
        }
    }
    
    return AgentResponse{
        ID:        "consensus_" + fmt.Sprintf("%d", time.Now().Unix()),
        AgentID:   "collaborative_consensus",
        Content:   combinedContent.String(),
        Timestamp: time.Now(),
        Final:     true,
        Metadata: map[string]interface{}{
            "consensus_type": "combined",
            "participant_count": len(responses),
        },
    }
}
```

---

## Best Practices

### 1. Tool Integration
- Keep tools focused and single-purpose
- Implement proper error handling in tools
- Use clear, descriptive tool names and descriptions
- Validate tool inputs and outputs
- Monitor tool usage and performance

### 2. Function Calling
- Design functions with clear, unambiguous names
- Provide comprehensive parameter schemas
- Handle edge cases gracefully
- Implement proper timeout and error handling
- Log function calls for debugging

### 3. Prompt Engineering
- Use clear, specific instructions
- Provide examples when appropriate
- Structure prompts logically
- Test prompts with different inputs
- Version and track prompt performance

### 4. Memory Management
- Balance memory size with performance
- Implement efficient search and retrieval
- Consider privacy and data retention policies
- Use appropriate persistence strategies
- Monitor memory usage and clean up regularly

### 5. Collaboration
- Design clear communication protocols
- Implement proper consensus mechanisms
- Handle agent failures gracefully
- Monitor collaborative performance
- Plan for scalability

---

## Next Steps

- **[Workflow Agents](workflow-agents.md)** - Sequential, parallel, and conditional patterns
- **[Multi-Agent Systems](multi-agent-systems.md)** - Coordination and communication
- **[State Management](state-management.md)** - Agent state and data flow
- **[Tool Development](/docs/technical/tools/creating-tools.md)** - Building custom tools
- **[Agent API Reference](/docs/technical/api-reference/agents.md)** - Detailed API documentation