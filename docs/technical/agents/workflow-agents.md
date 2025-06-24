# Workflow Agents: Sequential, Parallel, Conditional, and Loop Patterns

> **[Project Root](/) / [Documentation](../..) / [Technical Documentation](../../technical) / [Agents](../../technical/agents) / Workflow Agents**

Deep dive into workflow agent patterns in Go-LLMs, covering sequential execution, parallel processing, conditional branching, loop constructs, error handling strategies, and complex workflow orchestration for building sophisticated automated processes.

## Workflow Agent Architecture

### Core Workflow Interfaces

```go
// WorkflowAgent orchestrates multiple steps and agents
type WorkflowAgent interface {
    Agent
    
    // Step management
    AddStep(step WorkflowStep) error
    RemoveStep(name string) error
    GetSteps() []WorkflowStep
    GetStep(name string) (WorkflowStep, error)
    
    // Execution control
    ExecuteStep(ctx context.Context, stepName string, input interface{}) (interface{}, error)
    ExecuteSteps(ctx context.Context, stepNames []string, input interface{}) ([]interface{}, error)
    ExecuteWorkflow(ctx context.Context, input interface{}) (*WorkflowResult, error)
    
    // State and flow control
    GetWorkflowState() WorkflowState
    SetCondition(name string, condition Condition) error
    SetLoop(name string, loop Loop) error
    
    // Monitoring and control
    Pause() error
    Resume() error
    Stop() error
    GetExecutionStatus() ExecutionStatus
}

// WorkflowStep represents a single step in a workflow
type WorkflowStep interface {
    // Metadata
    Name() string
    Description() string
    Type() StepType
    
    // Execution
    Execute(ctx context.Context, input interface{}) (StepResult, error)
    
    // Dependencies and conditions
    GetDependencies() []string
    CanExecute(state WorkflowState) bool
    
    // Configuration
    GetConfig() StepConfig
    SetConfig(config StepConfig) error
    
    // Validation
    Validate() error
}

type StepType string

const (
    StepTypeAction      StepType = "action"      // Execute an action
    StepTypeCondition   StepType = "condition"   // Conditional branching
    StepTypeLoop        StepType = "loop"        // Loop execution
    StepTypeParallel    StepType = "parallel"    // Parallel execution
    StepTypeSubworkflow StepType = "subworkflow" // Nested workflow
    StepTypeJoin        StepType = "join"        // Join parallel branches
    StepTypeGateway     StepType = "gateway"     // Decision gateway
    StepTypeDelay       StepType = "delay"       // Delay/wait step
)

// WorkflowState tracks workflow execution state
type WorkflowState interface {
    // Step results
    GetStepResult(stepName string) (StepResult, bool)
    SetStepResult(stepName string, result StepResult)
    GetAllResults() map[string]StepResult
    
    // Variables
    GetVariable(name string) interface{}
    SetVariable(name string, value interface{})
    GetAllVariables() map[string]interface{}
    
    // Execution context
    GetCurrentStep() string
    SetCurrentStep(stepName string)
    GetExecutionPath() []string
    
    // Status
    IsCompleted() bool
    IsFailed() bool
    GetError() error
    SetError(error error)
    
    // Branching
    GetActiveBranches() []string
    AddBranch(branchName string)
    RemoveBranch(branchName string)
}

// StepResult contains step execution results
type StepResult struct {
    StepName    string                 `json:"step_name"`
    Output      interface{}            `json:"output"`
    Error       error                  `json:"error,omitempty"`
    StartTime   time.Time              `json:"start_time"`
    EndTime     time.Time              `json:"end_time"`
    Duration    time.Duration          `json:"duration"`
    Status      StepStatus             `json:"status"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
    RetryCount  int                    `json:"retry_count"`
}

type StepStatus string

const (
    StepStatusPending   StepStatus = "pending"
    StepStatusRunning   StepStatus = "running"
    StepStatusCompleted StepStatus = "completed"
    StepStatusFailed    StepStatus = "failed"
    StepStatusSkipped   StepStatus = "skipped"
    StepStatusRetrying  StepStatus = "retrying"
)
```

### Base Workflow Agent Implementation

```go
// DefaultWorkflowAgent implements WorkflowAgent
type DefaultWorkflowAgent struct {
    *BaseAgent
    
    // Workflow definition
    steps       map[string]WorkflowStep
    stepOrder   []string
    
    // Execution control
    state       WorkflowState
    executor    WorkflowExecutor
    scheduler   StepScheduler
    
    // Flow control
    conditions  map[string]Condition
    loops       map[string]Loop
    
    // Error handling
    errorHandler WorkflowErrorHandler
    retryPolicy  RetryPolicy
    
    // Monitoring
    monitor     WorkflowMonitor
    metrics     *WorkflowMetrics
    
    // Execution state
    status      ExecutionStatus
    pauseChan   chan struct{}
    stopChan    chan struct{}
}

type ExecutionStatus struct {
    State        ExecutionState `json:"state"`
    CurrentStep  string         `json:"current_step,omitempty"`
    Progress     float64        `json:"progress"`
    StartTime    time.Time      `json:"start_time"`
    EndTime      time.Time      `json:"end_time,omitempty"`
    Duration     time.Duration  `json:"duration"`
    StepsTotal   int            `json:"steps_total"`
    StepsCompleted int          `json:"steps_completed"`
    StepsFailed  int            `json:"steps_failed"`
}

type ExecutionState string

const (
    ExecutionStateIdle    ExecutionState = "idle"
    ExecutionStateRunning ExecutionState = "running"
    ExecutionStatePaused  ExecutionState = "paused"
    ExecutionStateStopped ExecutionState = "stopped"
    ExecutionStateCompleted ExecutionState = "completed"
    ExecutionStateFailed  ExecutionState = "failed"
)

func NewWorkflowAgent(name string, opts ...WorkflowAgentOption) *DefaultWorkflowAgent {
    agent := &DefaultWorkflowAgent{
        BaseAgent:   NewBaseAgent(name, AgentTypeWorkflow),
        steps:       make(map[string]WorkflowStep),
        stepOrder:   make([]string, 0),
        state:       NewWorkflowState(),
        conditions:  make(map[string]Condition),
        loops:       make(map[string]Loop),
        status:      ExecutionStatus{State: ExecutionStateIdle},
        pauseChan:   make(chan struct{}),
        stopChan:    make(chan struct{}),
    }
    
    // Apply options
    for _, opt := range opts {
        opt(agent)
    }
    
    // Set defaults if not provided
    if agent.executor == nil {
        agent.executor = NewDefaultWorkflowExecutor()
    }
    
    if agent.scheduler == nil {
        agent.scheduler = NewSequentialScheduler()
    }
    
    if agent.errorHandler == nil {
        agent.errorHandler = NewDefaultWorkflowErrorHandler()
    }
    
    return agent
}

type WorkflowAgentOption func(*DefaultWorkflowAgent)

func WithWorkflowExecutor(executor WorkflowExecutor) WorkflowAgentOption {
    return func(a *DefaultWorkflowAgent) {
        a.executor = executor
    }
}

func WithStepScheduler(scheduler StepScheduler) WorkflowAgentOption {
    return func(a *DefaultWorkflowAgent) {
        a.scheduler = scheduler
    }
}

func WithWorkflowMonitor(monitor WorkflowMonitor) WorkflowAgentOption {
    return func(a *DefaultWorkflowAgent) {
        a.monitor = monitor
    }
}
```

---

## Sequential Workflow Patterns

### Basic Sequential Execution

```go
// SequentialWorkflow executes steps in order
type SequentialWorkflow struct {
    *DefaultWorkflowAgent
}

func NewSequentialWorkflow(name string) *SequentialWorkflow {
    return &SequentialWorkflow{
        DefaultWorkflowAgent: NewWorkflowAgent(name,
            WithStepScheduler(NewSequentialScheduler()),
        ),
    }
}

func (w *SequentialWorkflow) ExecuteWorkflow(ctx context.Context, input interface{}) (*WorkflowResult, error) {
    w.status.State = ExecutionStateRunning
    w.status.StartTime = time.Now()
    
    result := &WorkflowResult{
        WorkflowID:   w.ID(),
        Input:        input,
        Steps:        make([]StepResult, 0),
        StartTime:    time.Now(),
    }
    
    currentInput := input
    
    for i, stepName := range w.stepOrder {
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        case <-w.stopChan:
            w.status.State = ExecutionStateStopped
            result.Status = WorkflowStatusStopped
            return result, nil
        case <-w.pauseChan:
            // Handle pause
            w.status.State = ExecutionStatePaused
            <-w.pauseChan // Wait for resume
            w.status.State = ExecutionStateRunning
        default:
        }
        
        step, exists := w.steps[stepName]
        if !exists {
            err := fmt.Errorf("step %s not found", stepName)
            result.Error = err
            return result, err
        }
        
        // Check if step can execute
        if !step.CanExecute(w.state) {
            w.logger.Info("Step skipped due to conditions",
                zap.String("workflow_id", w.ID()),
                zap.String("step_name", stepName),
            )
            
            stepResult := StepResult{
                StepName:  stepName,
                Status:    StepStatusSkipped,
                StartTime: time.Now(),
                EndTime:   time.Now(),
            }
            
            result.Steps = append(result.Steps, stepResult)
            w.state.SetStepResult(stepName, stepResult)
            continue
        }
        
        // Execute step
        w.status.CurrentStep = stepName
        w.state.SetCurrentStep(stepName)
        
        stepResult, err := w.executeStepWithRetry(ctx, step, currentInput)
        if err != nil {
            stepResult.Error = err
            stepResult.Status = StepStatusFailed
            w.status.StepsFailed++
            
            // Handle error
            if w.errorHandler.ShouldContinue(err, stepName, w.state) {
                w.logger.Warn("Step failed but continuing",
                    zap.String("workflow_id", w.ID()),
                    zap.String("step_name", stepName),
                    zap.Error(err),
                )
            } else {
                result.Error = err
                result.Status = WorkflowStatusFailed
                w.status.State = ExecutionStateFailed
                return result, err
            }
        } else {
            stepResult.Status = StepStatusCompleted
            w.status.StepsCompleted++
            currentInput = stepResult.Output // Pass output to next step
        }
        
        result.Steps = append(result.Steps, stepResult)
        w.state.SetStepResult(stepName, stepResult)
        
        // Update progress
        w.status.Progress = float64(i+1) / float64(len(w.stepOrder))
        
        // Monitor callback
        if w.monitor != nil {
            w.monitor.OnStepCompleted(w.ID(), stepResult)
        }
    }
    
    result.EndTime = time.Now()
    result.Duration = result.EndTime.Sub(result.StartTime)
    result.Status = WorkflowStatusCompleted
    result.Output = currentInput
    
    w.status.State = ExecutionStateCompleted
    w.status.EndTime = time.Now()
    w.status.Duration = w.status.EndTime.Sub(w.status.StartTime)
    
    return result, nil
}

func (w *SequentialWorkflow) executeStepWithRetry(ctx context.Context, step WorkflowStep, input interface{}) (StepResult, error) {
    var lastResult StepResult
    var lastErr error
    
    maxRetries := w.retryPolicy.GetMaxRetries(step.Name())
    
    for attempt := 0; attempt <= maxRetries; attempt++ {
        if attempt > 0 {
            // Apply backoff
            backoff := w.retryPolicy.GetBackoff(step.Name(), attempt)
            select {
            case <-time.After(backoff):
            case <-ctx.Done():
                return lastResult, ctx.Err()
            }
            
            lastResult.RetryCount = attempt
        }
        
        result, err := step.Execute(ctx, input)
        if err == nil {
            return result, nil
        }
        
        lastResult = result
        lastErr = err
        
        // Check if error is retryable
        if !w.retryPolicy.IsRetryable(step.Name(), err) {
            break
        }
        
        w.logger.Warn("Step execution failed, retrying",
            zap.String("workflow_id", w.ID()),
            zap.String("step_name", step.Name()),
            zap.Int("attempt", attempt+1),
            zap.Int("max_retries", maxRetries),
            zap.Error(err),
        )
    }
    
    return lastResult, lastErr
}

// Pipeline workflow for data processing
type PipelineWorkflow struct {
    *SequentialWorkflow
    transformers []DataTransformer
}

type DataTransformer interface {
    Transform(data interface{}) (interface{}, error)
    GetSchema() TransformSchema
}

func NewPipelineWorkflow(name string) *PipelineWorkflow {
    return &PipelineWorkflow{
        SequentialWorkflow: NewSequentialWorkflow(name),
        transformers:       make([]DataTransformer, 0),
    }
}

func (p *PipelineWorkflow) AddTransformer(transformer DataTransformer) error {
    p.transformers = append(p.transformers, transformer)
    
    // Create workflow step for transformer
    step := &TransformerStep{
        name:        fmt.Sprintf("transform_%d", len(p.transformers)),
        transformer: transformer,
    }
    
    return p.AddStep(step)
}

type TransformerStep struct {
    name        string
    transformer DataTransformer
}

func (s *TransformerStep) Name() string { return s.name }
func (s *TransformerStep) Description() string { return "Data transformation step" }
func (s *TransformerStep) Type() StepType { return StepTypeAction }

func (s *TransformerStep) Execute(ctx context.Context, input interface{}) (StepResult, error) {
    start := time.Now()
    
    output, err := s.transformer.Transform(input)
    
    return StepResult{
        StepName:  s.name,
        Output:    output,
        Error:     err,
        StartTime: start,
        EndTime:   time.Now(),
        Duration:  time.Since(start),
    }, err
}

func (s *TransformerStep) GetDependencies() []string { return nil }
func (s *TransformerStep) CanExecute(state WorkflowState) bool { return true }
func (s *TransformerStep) GetConfig() StepConfig { return StepConfig{} }
func (s *TransformerStep) SetConfig(config StepConfig) error { return nil }
func (s *TransformerStep) Validate() error { return nil }
```

---

## Parallel Workflow Patterns

### Parallel Execution

```go
// ParallelWorkflow executes steps concurrently
type ParallelWorkflow struct {
    *DefaultWorkflowAgent
    maxConcurrency int
    semaphore      chan struct{}
}

func NewParallelWorkflow(name string, maxConcurrency int) *ParallelWorkflow {
    return &ParallelWorkflow{
        DefaultWorkflowAgent: NewWorkflowAgent(name,
            WithStepScheduler(NewParallelScheduler(maxConcurrency)),
        ),
        maxConcurrency: maxConcurrency,
        semaphore:      make(chan struct{}, maxConcurrency),
    }
}

func (w *ParallelWorkflow) ExecuteWorkflow(ctx context.Context, input interface{}) (*WorkflowResult, error) {
    w.status.State = ExecutionStateRunning
    w.status.StartTime = time.Now()
    
    result := &WorkflowResult{
        WorkflowID: w.ID(),
        Input:      input,
        Steps:      make([]StepResult, 0),
        StartTime:  time.Now(),
    }
    
    // Group steps by dependencies
    stepGroups := w.groupStepsByDependencies()
    
    for groupIndex, stepGroup := range stepGroups {
        // Execute steps in current group in parallel
        groupResults, err := w.executeStepGroup(ctx, stepGroup, input, result)
        if err != nil {
            result.Error = err
            result.Status = WorkflowStatusFailed
            return result, err
        }
        
        result.Steps = append(result.Steps, groupResults...)
        
        // Update progress
        progress := float64(groupIndex+1) / float64(len(stepGroups))
        w.status.Progress = progress
        
        w.logger.Info("Step group completed",
            zap.String("workflow_id", w.ID()),
            zap.Int("group_index", groupIndex),
            zap.Int("steps_in_group", len(stepGroup)),
            zap.Float64("progress", progress),
        )
    }
    
    result.EndTime = time.Now()
    result.Duration = result.EndTime.Sub(result.StartTime)
    result.Status = WorkflowStatusCompleted
    
    w.status.State = ExecutionStateCompleted
    w.status.EndTime = time.Now()
    w.status.Duration = w.status.EndTime.Sub(w.status.StartTime)
    
    return result, nil
}

func (w *ParallelWorkflow) executeStepGroup(ctx context.Context, stepNames []string, input interface{}, result *WorkflowResult) ([]StepResult, error) {
    var wg sync.WaitGroup
    results := make([]StepResult, len(stepNames))
    errors := make([]error, len(stepNames))
    
    for i, stepName := range stepNames {
        wg.Add(1)
        
        go func(index int, name string) {
            defer wg.Done()
            
            // Acquire semaphore
            w.semaphore <- struct{}{}
            defer func() { <-w.semaphore }()
            
            step, exists := w.steps[name]
            if !exists {
                errors[index] = fmt.Errorf("step %s not found", name)
                return
            }
            
            // Check execution conditions
            if !step.CanExecute(w.state) {
                results[index] = StepResult{
                    StepName:  name,
                    Status:    StepStatusSkipped,
                    StartTime: time.Now(),
                    EndTime:   time.Now(),
                }
                return
            }
            
            // Execute step
            stepResult, err := w.executeStepWithRetry(ctx, step, input)
            results[index] = stepResult
            errors[index] = err
            
            // Update state
            w.state.SetStepResult(name, stepResult)
            
            if err != nil {
                w.status.StepsFailed++
            } else {
                w.status.StepsCompleted++
            }
            
            // Monitor callback
            if w.monitor != nil {
                w.monitor.OnStepCompleted(w.ID(), stepResult)
            }
        }(i, stepName)
    }
    
    wg.Wait()
    
    // Check for errors
    var firstError error
    for i, err := range errors {
        if err != nil {
            if firstError == nil {
                firstError = err
            }
            
            if !w.errorHandler.ShouldContinue(err, stepNames[i], w.state) {
                return results, fmt.Errorf("step group failed: %w", firstError)
            }
        }
    }
    
    return results, nil
}

func (w *ParallelWorkflow) groupStepsByDependencies() [][]string {
    // Topological sort to group steps by dependency levels
    inDegree := make(map[string]int)
    graph := make(map[string][]string)
    
    // Initialize
    for stepName := range w.steps {
        inDegree[stepName] = 0
        graph[stepName] = make([]string, 0)
    }
    
    // Build dependency graph
    for stepName, step := range w.steps {
        dependencies := step.GetDependencies()
        inDegree[stepName] = len(dependencies)
        
        for _, dep := range dependencies {
            graph[dep] = append(graph[dep], stepName)
        }
    }
    
    var groups [][]string
    queue := make([]string, 0)
    
    // Find steps with no dependencies
    for stepName, degree := range inDegree {
        if degree == 0 {
            queue = append(queue, stepName)
        }
    }
    
    for len(queue) > 0 {
        currentGroup := make([]string, len(queue))
        copy(currentGroup, queue)
        groups = append(groups, currentGroup)
        
        nextQueue := make([]string, 0)
        
        for _, stepName := range queue {
            for _, dependent := range graph[stepName] {
                inDegree[dependent]--
                if inDegree[dependent] == 0 {
                    nextQueue = append(nextQueue, dependent)
                }
            }
        }
        
        queue = nextQueue
    }
    
    return groups
}

// Fork-Join pattern
type ForkJoinWorkflow struct {
    *ParallelWorkflow
    forkSteps []string
    joinStep  WorkflowStep
}

func NewForkJoinWorkflow(name string, forkSteps []string, joinStep WorkflowStep) *ForkJoinWorkflow {
    workflow := &ForkJoinWorkflow{
        ParallelWorkflow: NewParallelWorkflow(name, len(forkSteps)),
        forkSteps:        forkSteps,
        joinStep:         joinStep,
    }
    
    return workflow
}

func (w *ForkJoinWorkflow) ExecuteWorkflow(ctx context.Context, input interface{}) (*WorkflowResult, error) {
    // Execute fork steps in parallel
    forkResults, err := w.executeForkSteps(ctx, input)
    if err != nil {
        return nil, fmt.Errorf("fork execution failed: %w", err)
    }
    
    // Collect results for join step
    joinInput := w.aggregateForkResults(forkResults)
    
    // Execute join step
    joinResult, err := w.joinStep.Execute(ctx, joinInput)
    if err != nil {
        return nil, fmt.Errorf("join execution failed: %w", err)
    }
    
    // Build workflow result
    result := &WorkflowResult{
        WorkflowID: w.ID(),
        Input:      input,
        Steps:      append(forkResults, joinResult),
        Output:     joinResult.Output,
        StartTime:  time.Now(), // This would be tracked properly
        EndTime:    time.Now(),
        Status:     WorkflowStatusCompleted,
    }
    
    return result, nil
}

func (w *ForkJoinWorkflow) executeForkSteps(ctx context.Context, input interface{}) ([]StepResult, error) {
    results := make([]StepResult, len(w.forkSteps))
    errors := make([]error, len(w.forkSteps))
    
    var wg sync.WaitGroup
    
    for i, stepName := range w.forkSteps {
        wg.Add(1)
        
        go func(index int, name string) {
            defer wg.Done()
            
            step, exists := w.steps[name]
            if !exists {
                errors[index] = fmt.Errorf("fork step %s not found", name)
                return
            }
            
            result, err := step.Execute(ctx, input)
            results[index] = result
            errors[index] = err
        }(i, stepName)
    }
    
    wg.Wait()
    
    // Check for errors
    for i, err := range errors {
        if err != nil {
            return nil, fmt.Errorf("fork step %s failed: %w", w.forkSteps[i], err)
        }
    }
    
    return results, nil
}

func (w *ForkJoinWorkflow) aggregateForkResults(results []StepResult) interface{} {
    outputs := make([]interface{}, len(results))
    for i, result := range results {
        outputs[i] = result.Output
    }
    
    return map[string]interface{}{
        "fork_results": outputs,
        "step_names":   w.forkSteps,
    }
}
```

---

## Conditional Workflow Patterns

### Conditional Branching

```go
// Condition interface for workflow branching
type Condition interface {
    Evaluate(state WorkflowState) (bool, error)
    GetExpression() string
    GetDescription() string
}

// SimpleCondition evaluates basic conditions
type SimpleCondition struct {
    expression  string
    description string
    evaluator   ConditionEvaluator
}

type ConditionEvaluator func(state WorkflowState) (bool, error)

func NewSimpleCondition(expression, description string, evaluator ConditionEvaluator) *SimpleCondition {
    return &SimpleCondition{
        expression:  expression,
        description: description,
        evaluator:   evaluator,
    }
}

func (c *SimpleCondition) Evaluate(state WorkflowState) (bool, error) {
    return c.evaluator(state)
}

func (c *SimpleCondition) GetExpression() string { return c.expression }
func (c *SimpleCondition) GetDescription() string { return c.description }

// ConditionalWorkflow supports branching logic
type ConditionalWorkflow struct {
    *DefaultWorkflowAgent
    branches map[string]ConditionalBranch
}

type ConditionalBranch struct {
    Condition Condition
    Steps     []string
    ElseSteps []string
}

func NewConditionalWorkflow(name string) *ConditionalWorkflow {
    return &ConditionalWorkflow{
        DefaultWorkflowAgent: NewWorkflowAgent(name),
        branches:             make(map[string]ConditionalBranch),
    }
}

func (w *ConditionalWorkflow) AddBranch(name string, condition Condition, thenSteps, elseSteps []string) error {
    w.branches[name] = ConditionalBranch{
        Condition: condition,
        Steps:     thenSteps,
        ElseSteps: elseSteps,
    }
    
    return nil
}

func (w *ConditionalWorkflow) ExecuteWorkflow(ctx context.Context, input interface{}) (*WorkflowResult, error) {
    result := &WorkflowResult{
        WorkflowID: w.ID(),
        Input:      input,
        Steps:      make([]StepResult, 0),
        StartTime:  time.Now(),
    }
    
    w.status.State = ExecutionStateRunning
    w.status.StartTime = time.Now()
    
    // Execute workflow with conditional logic
    currentInput := input
    
    for i, stepName := range w.stepOrder {
        // Check if this is a conditional branch
        if branch, exists := w.branches[stepName]; exists {
            branchResult, err := w.executeBranch(ctx, stepName, branch, currentInput)
            if err != nil {
                result.Error = err
                result.Status = WorkflowStatusFailed
                return result, err
            }
            
            result.Steps = append(result.Steps, branchResult...)
            if len(branchResult) > 0 {
                currentInput = branchResult[len(branchResult)-1].Output
            }
        } else {
            // Regular step execution
            step, exists := w.steps[stepName]
            if !exists {
                err := fmt.Errorf("step %s not found", stepName)
                result.Error = err
                return result, err
            }
            
            stepResult, err := step.Execute(ctx, currentInput)
            if err != nil {
                result.Error = err
                result.Status = WorkflowStatusFailed
                return result, err
            }
            
            result.Steps = append(result.Steps, stepResult)
            w.state.SetStepResult(stepName, stepResult)
            currentInput = stepResult.Output
        }
        
        // Update progress
        w.status.Progress = float64(i+1) / float64(len(w.stepOrder))
    }
    
    result.EndTime = time.Now()
    result.Duration = result.EndTime.Sub(result.StartTime)
    result.Status = WorkflowStatusCompleted
    result.Output = currentInput
    
    w.status.State = ExecutionStateCompleted
    
    return result, nil
}

func (w *ConditionalWorkflow) executeBranch(ctx context.Context, branchName string, branch ConditionalBranch, input interface{}) ([]StepResult, error) {
    // Evaluate condition
    conditionResult, err := branch.Condition.Evaluate(w.state)
    if err != nil {
        return nil, fmt.Errorf("condition evaluation failed for branch %s: %w", branchName, err)
    }
    
    w.logger.Info("Branch condition evaluated",
        zap.String("workflow_id", w.ID()),
        zap.String("branch_name", branchName),
        zap.String("condition", branch.Condition.GetExpression()),
        zap.Bool("result", conditionResult),
    )
    
    // Choose steps to execute
    var stepsToExecute []string
    if conditionResult {
        stepsToExecute = branch.Steps
        w.state.AddBranch(branchName + "_then")
    } else {
        stepsToExecute = branch.ElseSteps
        w.state.AddBranch(branchName + "_else")
    }
    
    // Execute chosen steps
    var results []StepResult
    currentInput := input
    
    for _, stepName := range stepsToExecute {
        step, exists := w.steps[stepName]
        if !exists {
            return nil, fmt.Errorf("branch step %s not found", stepName)
        }
        
        stepResult, err := step.Execute(ctx, currentInput)
        if err != nil {
            return nil, fmt.Errorf("branch step %s failed: %w", stepName, err)
        }
        
        results = append(results, stepResult)
        w.state.SetStepResult(stepName, stepResult)
        currentInput = stepResult.Output
    }
    
    return results, nil
}

// Decision Gateway pattern
type DecisionGateway struct {
    name      string
    decisions []Decision
}

type Decision struct {
    Condition   Condition
    NextStep    string
    Description string
}

func NewDecisionGateway(name string) *DecisionGateway {
    return &DecisionGateway{
        name:      name,
        decisions: make([]Decision, 0),
    }
}

func (g *DecisionGateway) AddDecision(condition Condition, nextStep, description string) {
    g.decisions = append(g.decisions, Decision{
        Condition:   condition,
        NextStep:    nextStep,
        Description: description,
}
}

func (g *DecisionGateway) Name() string { return g.name }
func (g *DecisionGateway) Description() string { return "Decision gateway step" }
func (g *DecisionGateway) Type() StepType { return StepTypeGateway }

func (g *DecisionGateway) Execute(ctx context.Context, input interface{}) (StepResult, error) {
    start := time.Now()
    
    // This would be integrated with WorkflowState
    // For now, we'll create a placeholder implementation
    
    result := StepResult{
        StepName:  g.name,
        StartTime: start,
        EndTime:   time.Now(),
        Duration:  time.Since(start),
        Status:    StepStatusCompleted,
        Output:    input, // Pass through input
        Metadata: map[string]interface{}{
            "gateway_type": "decision",
            "decisions":    len(g.decisions),
        },
    }
    
    return result, nil
}

func (g *DecisionGateway) GetDependencies() []string { return nil }
func (g *DecisionGateway) CanExecute(state WorkflowState) bool { return true }
func (g *DecisionGateway) GetConfig() StepConfig { return StepConfig{} }
func (g *DecisionGateway) SetConfig(config StepConfig) error { return nil }
func (g *DecisionGateway) Validate() error { return nil }

// Example conditions
func CreateExampleConditions() map[string]Condition {
    conditions := make(map[string]Condition)
    
    // Variable-based condition
    conditions["has_error"] = NewSimpleCondition(
        "error != nil",
        "Check if previous step had an error",
        func(state WorkflowState) (bool, error) {
            return state.GetError() != nil, nil
        },
    )
    
    // Value comparison condition
    conditions["score_threshold"] = NewSimpleCondition(
        "score > 0.8",
        "Check if score exceeds threshold",
        func(state WorkflowState) (bool, error) {
            score := state.GetVariable("score")
            if scoreValue, ok := score.(float64); ok {
                return scoreValue > 0.8, nil
            }
            return false, fmt.Errorf("score variable not found or invalid type")
        },
    )
    
    // Step result condition
    conditions["processing_success"] = NewSimpleCondition(
        "previous_step.status == 'completed'",
        "Check if previous processing step completed successfully",
        func(state WorkflowState) (bool, error) {
            if result, exists := state.GetStepResult("data_processing"); exists {
                return result.Status == StepStatusCompleted, nil
            }
            return false, nil
        },
    )
    
    return conditions
}
```

---

## Loop Workflow Patterns

### Loop Constructs

```go
// Loop interface for iterative execution
type Loop interface {
    ShouldContinue(state WorkflowState, iteration int) (bool, error)
    GetMaxIterations() int
    GetType() LoopType
    GetSteps() []string
}

type LoopType string

const (
    LoopTypeWhile   LoopType = "while"   // While condition is true
    LoopTypeFor     LoopType = "for"     // For N iterations
    LoopTypeForEach LoopType = "foreach" // For each item in collection
    LoopTypeUntil   LoopType = "until"   // Until condition is true
)

// WhileLoop continues while condition is true
type WhileLoop struct {
    condition     Condition
    steps         []string
    maxIterations int
}

func NewWhileLoop(condition Condition, steps []string, maxIterations int) *WhileLoop {
    return &WhileLoop{
        condition:     condition,
        steps:         steps,
        maxIterations: maxIterations,
    }
}

func (l *WhileLoop) ShouldContinue(state WorkflowState, iteration int) (bool, error) {
    if iteration >= l.maxIterations {
        return false, nil
    }
    
    return l.condition.Evaluate(state)
}

func (l *WhileLoop) GetMaxIterations() int { return l.maxIterations }
func (l *WhileLoop) GetType() LoopType { return LoopTypeWhile }
func (l *WhileLoop) GetSteps() []string { return l.steps }

// ForLoop executes for a fixed number of iterations
type ForLoop struct {
    iterations int
    steps      []string
}

func NewForLoop(iterations int, steps []string) *ForLoop {
    return &ForLoop{
        iterations: iterations,
        steps:      steps,
    }
}

func (l *ForLoop) ShouldContinue(state WorkflowState, iteration int) (bool, error) {
    return iteration < l.iterations, nil
}

func (l *ForLoop) GetMaxIterations() int { return l.iterations }
func (l *ForLoop) GetType() LoopType { return LoopTypeFor }
func (l *ForLoop) GetSteps() []string { return l.steps }

// ForEachLoop iterates over a collection
type ForEachLoop struct {
    collectionPath string
    itemVariable   string
    steps          []string
    collection     []interface{}
}

func NewForEachLoop(collectionPath, itemVariable string, steps []string) *ForEachLoop {
    return &ForEachLoop{
        collectionPath: collectionPath,
        itemVariable:   itemVariable,
        steps:          steps,
    }
}

func (l *ForEachLoop) ShouldContinue(state WorkflowState, iteration int) (bool, error) {
    // Initialize collection if not done yet
    if l.collection == nil {
        collectionValue := state.GetVariable(l.collectionPath)
        if collection, ok := collectionValue.([]interface{}); ok {
            l.collection = collection
        } else {
            return false, fmt.Errorf("collection %s not found or invalid type", l.collectionPath)
        }
    }
    
    if iteration < len(l.collection) {
        // Set current item as variable
        state.SetVariable(l.itemVariable, l.collection[iteration])
        return true, nil
    }
    
    return false, nil
}

func (l *ForEachLoop) GetMaxIterations() int { return len(l.collection) }
func (l *ForEachLoop) GetType() LoopType { return LoopTypeForEach }
func (l *ForEachLoop) GetSteps() []string { return l.steps }

// LoopWorkflow supports iterative execution
type LoopWorkflow struct {
    *DefaultWorkflowAgent
    loops map[string]Loop
}

func NewLoopWorkflow(name string) *LoopWorkflow {
    return &LoopWorkflow{
        DefaultWorkflowAgent: NewWorkflowAgent(name),
        loops:                make(map[string]Loop),
    }
}

func (w *LoopWorkflow) AddLoop(name string, loop Loop) error {
    w.loops[name] = loop
    return nil
}

func (w *LoopWorkflow) ExecuteWorkflow(ctx context.Context, input interface{}) (*WorkflowResult, error) {
    result := &WorkflowResult{
        WorkflowID: w.ID(),
        Input:      input,
        Steps:      make([]StepResult, 0),
        StartTime:  time.Now(),
    }
    
    w.status.State = ExecutionStateRunning
    w.status.StartTime = time.Now()
    
    currentInput := input
    
    for _, stepName := range w.stepOrder {
        // Check if this is a loop
        if loop, exists := w.loops[stepName]; exists {
            loopResults, err := w.executeLoop(ctx, stepName, loop, currentInput)
            if err != nil {
                result.Error = err
                result.Status = WorkflowStatusFailed
                return result, err
            }
            
            result.Steps = append(result.Steps, loopResults...)
            if len(loopResults) > 0 {
                currentInput = loopResults[len(loopResults)-1].Output
            }
        } else {
            // Regular step execution
            step, exists := w.steps[stepName]
            if !exists {
                err := fmt.Errorf("step %s not found", stepName)
                result.Error = err
                return result, err
            }
            
            stepResult, err := step.Execute(ctx, currentInput)
            if err != nil {
                result.Error = err
                result.Status = WorkflowStatusFailed
                return result, err
            }
            
            result.Steps = append(result.Steps, stepResult)
            w.state.SetStepResult(stepName, stepResult)
            currentInput = stepResult.Output
        }
    }
    
    result.EndTime = time.Now()
    result.Duration = result.EndTime.Sub(result.StartTime)
    result.Status = WorkflowStatusCompleted
    result.Output = currentInput
    
    w.status.State = ExecutionStateCompleted
    
    return result, nil
}

func (w *LoopWorkflow) executeLoop(ctx context.Context, loopName string, loop Loop, input interface{}) ([]StepResult, error) {
    var allResults []StepResult
    iteration := 0
    currentInput := input
    
    w.logger.Info("Starting loop execution",
        zap.String("workflow_id", w.ID()),
        zap.String("loop_name", loopName),
        zap.String("loop_type", string(loop.GetType())),
        zap.Int("max_iterations", loop.GetMaxIterations()),
    )
    
    for {
        // Check loop condition
        shouldContinue, err := loop.ShouldContinue(w.state, iteration)
        if err != nil {
            return nil, fmt.Errorf("loop condition evaluation failed: %w", err)
        }
        
        if !shouldContinue {
            w.logger.Info("Loop terminated",
                zap.String("workflow_id", w.ID()),
                zap.String("loop_name", loopName),
                zap.Int("iterations", iteration),
            )
            break
        }
        
        // Check context cancellation
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        default:
        }
        
        w.logger.Debug("Loop iteration starting",
            zap.String("workflow_id", w.ID()),
            zap.String("loop_name", loopName),
            zap.Int("iteration", iteration),
        )
        
        // Execute loop steps
        iterationResults, err := w.executeLoopIteration(ctx, loop, currentInput, iteration)
        if err != nil {
            return nil, fmt.Errorf("loop iteration %d failed: %w", iteration, err)
        }
        
        allResults = append(allResults, iterationResults...)
        
        // Update input for next iteration
        if len(iterationResults) > 0 {
            currentInput = iterationResults[len(iterationResults)-1].Output
        }
        
        // Update loop state
        w.state.SetVariable(fmt.Sprintf("%s_iteration", loopName), iteration)
        w.state.SetVariable(fmt.Sprintf("%s_last_result", loopName), currentInput)
        
        iteration++
    }
    
    return allResults, nil
}

func (w *LoopWorkflow) executeLoopIteration(ctx context.Context, loop Loop, input interface{}, iteration int) ([]StepResult, error) {
    var results []StepResult
    currentInput := input
    
    for _, stepName := range loop.GetSteps() {
        step, exists := w.steps[stepName]
        if !exists {
            return nil, fmt.Errorf("loop step %s not found", stepName)
        }
        
        stepResult, err := step.Execute(ctx, currentInput)
        if err != nil {
            return nil, fmt.Errorf("loop step %s failed in iteration %d: %w", stepName, iteration, err)
        }
        
        // Modify step name to include iteration
        stepResult.StepName = fmt.Sprintf("%s_iter_%d", stepName, iteration)
        stepResult.Metadata = map[string]interface{}{
            "loop_iteration": iteration,
            "original_step":  stepName,
        }
        
        results = append(results, stepResult)
        w.state.SetStepResult(stepResult.StepName, stepResult)
        currentInput = stepResult.Output
    }
    
    return results, nil
}

// Batch processing pattern with loops
type BatchProcessingWorkflow struct {
    *LoopWorkflow
    batchSize    int
    processor    BatchProcessor
}

type BatchProcessor interface {
    ProcessBatch(ctx context.Context, batch []interface{}) ([]interface{}, error)
    GetBatchSize() int
}

func NewBatchProcessingWorkflow(name string, batchSize int, processor BatchProcessor) *BatchProcessingWorkflow {
    return &BatchProcessingWorkflow{
        LoopWorkflow: NewLoopWorkflow(name),
        batchSize:    batchSize,
        processor:    processor,
    }
}

func (w *BatchProcessingWorkflow) ProcessItems(ctx context.Context, items []interface{}) (*WorkflowResult, error) {
    // Create batches
    batches := w.createBatches(items)
    w.state.SetVariable("batches", batches)
    w.state.SetVariable("current_batch_index", 0)
    
    // Create loop for batch processing
    batchLoop := NewForLoop(len(batches), []string{"process_batch"})
    w.AddLoop("batch_processing", batchLoop)
    
    // Add batch processing step
    batchStep := &BatchProcessingStep{
        name:      "process_batch",
        processor: w.processor,
        workflow:  w,
    }
    w.AddStep(batchStep)
    
    return w.ExecuteWorkflow(ctx, items)
}

func (w *BatchProcessingWorkflow) createBatches(items []interface{}) [][]interface{} {
    var batches [][]interface{}
    
    for i := 0; i < len(items); i += w.batchSize {
        end := i + w.batchSize
        if end > len(items) {
            end = len(items)
        }
        batches = append(batches, items[i:end])
    }
    
    return batches
}

type BatchProcessingStep struct {
    name      string
    processor BatchProcessor
    workflow  *BatchProcessingWorkflow
}

func (s *BatchProcessingStep) Name() string { return s.name }
func (s *BatchProcessingStep) Description() string { return "Process a batch of items" }
func (s *BatchProcessingStep) Type() StepType { return StepTypeAction }

func (s *BatchProcessingStep) Execute(ctx context.Context, input interface{}) (StepResult, error) {
    start := time.Now()
    
    // Get current batch
    batches := s.workflow.state.GetVariable("batches").([][]interface{})
    batchIndex := s.workflow.state.GetVariable("current_batch_index").(int)
    
    if batchIndex >= len(batches) {
        return StepResult{
            StepName:  s.name,
            Error:     fmt.Errorf("batch index out of range"),
            StartTime: start,
            EndTime:   time.Now(),
            Status:    StepStatusFailed,
        }, fmt.Errorf("batch index out of range")
    }
    
    currentBatch := batches[batchIndex]
    
    // Process batch
    results, err := s.processor.ProcessBatch(ctx, currentBatch)
    if err != nil {
        return StepResult{
            StepName:  s.name,
            Error:     err,
            StartTime: start,
            EndTime:   time.Now(),
            Status:    StepStatusFailed,
        }, err
    }
    
    // Update batch index for next iteration
    s.workflow.state.SetVariable("current_batch_index", batchIndex+1)
    
    return StepResult{
        StepName:  s.name,
        Output:    results,
        StartTime: start,
        EndTime:   time.Now(),
        Duration:  time.Since(start),
        Status:    StepStatusCompleted,
        Metadata: map[string]interface{}{
            "batch_index": batchIndex,
            "batch_size":  len(currentBatch),
            "results_count": len(results),
        },
    }, nil
}

func (s *BatchProcessingStep) GetDependencies() []string { return nil }
func (s *BatchProcessingStep) CanExecute(state WorkflowState) bool { return true }
func (s *BatchProcessingStep) GetConfig() StepConfig { return StepConfig{} }
func (s *BatchProcessingStep) SetConfig(config StepConfig) error { return nil }
func (s *BatchProcessingStep) Validate() error { return nil }
```

---

## Error Handling and Recovery

### Workflow Error Handling

```go
// WorkflowErrorHandler manages error handling strategies
type WorkflowErrorHandler interface {
    HandleError(err error, stepName string, state WorkflowState) ErrorAction
    ShouldContinue(err error, stepName string, state WorkflowState) bool
    GetRecoveryStrategy(err error, stepName string) RecoveryStrategy
}

type ErrorAction string

const (
    ErrorActionStop     ErrorAction = "stop"     // Stop workflow execution
    ErrorActionContinue ErrorAction = "continue" // Continue with next step
    ErrorActionRetry    ErrorAction = "retry"    // Retry the failed step
    ErrorActionSkip     ErrorAction = "skip"     // Skip the failed step
    ErrorActionFallback ErrorAction = "fallback" // Execute fallback step
)

type RecoveryStrategy interface {
    Recover(ctx context.Context, err error, step WorkflowStep, input interface{}) (interface{}, error)
    CanRecover(err error) bool
}

// DefaultWorkflowErrorHandler provides basic error handling
type DefaultWorkflowErrorHandler struct {
    strategies map[string]ErrorAction
    recovery   map[string]RecoveryStrategy
    logger     *zap.Logger
}

func NewDefaultWorkflowErrorHandler() *DefaultWorkflowErrorHandler {
    return &DefaultWorkflowErrorHandler{
        strategies: map[string]ErrorAction{
            "timeout":           ErrorActionRetry,
            "rate_limit":        ErrorActionRetry,
            "temporary_failure": ErrorActionRetry,
            "validation_error":  ErrorActionStop,
            "permission_denied": ErrorActionStop,
            "not_found":         ErrorActionSkip,
        },
        recovery: make(map[string]RecoveryStrategy),
        logger:   zap.NewNop(),
    }
}

func (h *DefaultWorkflowErrorHandler) HandleError(err error, stepName string, state WorkflowState) ErrorAction {
    // Classify error
    errorType := h.classifyError(err)
    
    // Get action for error type
    if action, exists := h.strategies[errorType]; exists {
        h.logger.Info("Error handling strategy determined",
            zap.String("step_name", stepName),
            zap.String("error_type", errorType),
            zap.String("action", string(action)),
            zap.Error(err),
        )
        return action
    }
    
    // Default action
    return ErrorActionStop
}

func (h *DefaultWorkflowErrorHandler) ShouldContinue(err error, stepName string, state WorkflowState) bool {
    action := h.HandleError(err, stepName, state)
    return action == ErrorActionContinue || action == ErrorActionSkip
}

func (h *DefaultWorkflowErrorHandler) GetRecoveryStrategy(err error, stepName string) RecoveryStrategy {
    errorType := h.classifyError(err)
    if strategy, exists := h.recovery[errorType]; exists {
        return strategy
    }
    return nil
}

func (h *DefaultWorkflowErrorHandler) classifyError(err error) string {
    errMsg := strings.ToLower(err.Error())
    
    switch {
    case strings.Contains(errMsg, "timeout"):
        return "timeout"
    case strings.Contains(errMsg, "rate limit"):
        return "rate_limit"
    case strings.Contains(errMsg, "temporary"):
        return "temporary_failure"
    case strings.Contains(errMsg, "validation"):
        return "validation_error"
    case strings.Contains(errMsg, "permission"):
        return "permission_denied"
    case strings.Contains(errMsg, "not found"):
        return "not_found"
    default:
        return "unknown"
    }
}

// Compensation pattern for workflow rollback
type CompensationWorkflow struct {
    *DefaultWorkflowAgent
    compensations map[string]WorkflowStep
    executed      []string
}

func NewCompensationWorkflow(name string) *CompensationWorkflow {
    return &CompensationWorkflow{
        DefaultWorkflowAgent: NewWorkflowAgent(name),
        compensations:        make(map[string]WorkflowStep),
        executed:             make([]string, 0),
    }
}

func (w *CompensationWorkflow) AddCompensation(stepName string, compensation WorkflowStep) error {
    w.compensations[stepName] = compensation
    return nil
}

func (w *CompensationWorkflow) ExecuteWorkflow(ctx context.Context, input interface{}) (*WorkflowResult, error) {
    result := &WorkflowResult{
        WorkflowID: w.ID(),
        Input:      input,
        Steps:      make([]StepResult, 0),
        StartTime:  time.Now(),
    }
    
    currentInput := input
    
    for _, stepName := range w.stepOrder {
        step, exists := w.steps[stepName]
        if !exists {
            err := fmt.Errorf("step %s not found", stepName)
            // Compensate for executed steps
            w.compensate(ctx, result)
            result.Error = err
            return result, err
        }
        
        stepResult, err := step.Execute(ctx, currentInput)
        if err != nil {
            // Execute compensation for all executed steps
            w.compensate(ctx, result)
            result.Error = err
            result.Status = WorkflowStatusFailed
            return result, err
        }
        
        w.executed = append(w.executed, stepName)
        result.Steps = append(result.Steps, stepResult)
        w.state.SetStepResult(stepName, stepResult)
        currentInput = stepResult.Output
    }
    
    result.EndTime = time.Now()
    result.Duration = result.EndTime.Sub(result.StartTime)
    result.Status = WorkflowStatusCompleted
    result.Output = currentInput
    
    return result, nil
}

func (w *CompensationWorkflow) compensate(ctx context.Context, result *WorkflowResult) {
    w.logger.Info("Starting compensation",
        zap.String("workflow_id", w.ID()),
        zap.Int("steps_to_compensate", len(w.executed)),
    )
    
    // Execute compensations in reverse order
    for i := len(w.executed) - 1; i >= 0; i-- {
        stepName := w.executed[i]
        
        if compensation, exists := w.compensations[stepName]; exists {
            compensationResult, err := compensation.Execute(ctx, nil)
            if err != nil {
                w.logger.Error("Compensation failed",
                    zap.String("workflow_id", w.ID()),
                    zap.String("step_name", stepName),
                    zap.Error(err),
                )
            } else {
                w.logger.Info("Compensation executed",
                    zap.String("workflow_id", w.ID()),
                    zap.String("step_name", stepName),
                )
                
                // Add compensation result to workflow result
                compensationResult.StepName = stepName + "_compensation"
                result.Steps = append(result.Steps, compensationResult)
            }
        }
    }
}

// Circuit breaker pattern for step execution
type CircuitBreakerStep struct {
    step           WorkflowStep
    circuitBreaker *CircuitBreaker
}

type CircuitBreaker struct {
    failureThreshold int
    recoveryTimeout  time.Duration
    failures         int
    lastFailureTime  time.Time
    state            CircuitBreakerState
    mu               sync.Mutex
}

type CircuitBreakerState string

const (
    CircuitBreakerClosed   CircuitBreakerState = "closed"
    CircuitBreakerOpen     CircuitBreakerState = "open"
    CircuitBreakerHalfOpen CircuitBreakerState = "half_open"
)

func NewCircuitBreakerStep(step WorkflowStep, failureThreshold int, recoveryTimeout time.Duration) *CircuitBreakerStep {
    return &CircuitBreakerStep{
        step: step,
        circuitBreaker: &CircuitBreaker{
            failureThreshold: failureThreshold,
            recoveryTimeout:  recoveryTimeout,
            state:            CircuitBreakerClosed,
        },
    }
}

func (s *CircuitBreakerStep) Execute(ctx context.Context, input interface{}) (StepResult, error) {
    s.circuitBreaker.mu.Lock()
    defer s.circuitBreaker.mu.Unlock()
    
    // Check circuit breaker state
    switch s.circuitBreaker.state {
    case CircuitBreakerOpen:
        if time.Since(s.circuitBreaker.lastFailureTime) > s.circuitBreaker.recoveryTimeout {
            s.circuitBreaker.state = CircuitBreakerHalfOpen
        } else {
            return StepResult{
                StepName:  s.step.Name(),
                Error:     fmt.Errorf("circuit breaker is open"),
                StartTime: time.Now(),
                EndTime:   time.Now(),
                Status:    StepStatusFailed,
            }, fmt.Errorf("circuit breaker is open")
        }
    }
    
    // Execute step
    result, err := s.step.Execute(ctx, input)
    
    if err != nil {
        s.circuitBreaker.failures++
        s.circuitBreaker.lastFailureTime = time.Now()
        
        if s.circuitBreaker.failures >= s.circuitBreaker.failureThreshold {
            s.circuitBreaker.state = CircuitBreakerOpen
        }
        
        return result, err
    }
    
    // Success - reset circuit breaker
    s.circuitBreaker.failures = 0
    s.circuitBreaker.state = CircuitBreakerClosed
    
    return result, nil
}

// Delegate methods
func (s *CircuitBreakerStep) Name() string { return s.step.Name() }
func (s *CircuitBreakerStep) Description() string { return s.step.Description() }
func (s *CircuitBreakerStep) Type() StepType { return s.step.Type() }
func (s *CircuitBreakerStep) GetDependencies() []string { return s.step.GetDependencies() }
func (s *CircuitBreakerStep) CanExecute(state WorkflowState) bool { return s.step.CanExecute(state) }
func (s *CircuitBreakerStep) GetConfig() StepConfig { return s.step.GetConfig() }
func (s *CircuitBreakerStep) SetConfig(config StepConfig) error { return s.step.SetConfig(config) }
func (s *CircuitBreakerStep) Validate() error { return s.step.Validate() }
```

---

## Best Practices

### 1. Sequential Workflows
- Keep steps focused and single-purpose
- Use clear naming conventions
- Implement proper error handling
- Design for testability
- Consider step dependencies

### 2. Parallel Workflows
- Manage concurrency limits
- Handle partial failures gracefully
- Design thread-safe steps
- Monitor resource usage
- Plan for synchronization points

### 3. Conditional Workflows
- Use clear, testable conditions
- Document decision logic
- Handle edge cases
- Test all branches
- Consider default paths

### 4. Loop Workflows
- Set reasonable iteration limits
- Implement proper termination conditions
- Handle infinite loop prevention
- Monitor performance impact
- Design stateless iterations when possible

### 5. Error Handling
- Implement comprehensive error strategies
- Use circuit breakers for resilience
- Design compensation patterns
- Monitor error rates
- Plan for graceful degradation

---

## Next Steps

- **[Multi-Agent Systems](multi-agent-systems.md)** - Coordination and communication
- **[State Management](state-management.md)** - Agent state and data flow
- **[LLM Agents](llm-agents.md)** - AI-powered agents with tool support
- **[Agent Overview](overview.md)** - Agent architecture and concepts
- **[Agent API Reference](../../technical/api-reference/agents.md)** - Detailed API documentation