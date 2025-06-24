# Workflow Orchestration: Complex Workflow Patterns

> **[Project Root](/) / [Documentation](/docs/) / [User Guide](/docs/user-guide/) / [Advanced Topics](/docs/user-guide/advanced/) / Workflow Orchestration**

Master advanced workflow orchestration patterns in Go-LLMs, including parallel execution, conditional branching, error recovery, state management, and building sophisticated multi-agent systems.

## Workflow Architecture Overview

Go-LLMs provides powerful workflow primitives for orchestrating complex agent interactions:

```go
// Core workflow interfaces
type Workflow interface {
    Execute(ctx context.Context, input interface{}) (interface{}, error)
    GetState() WorkflowState
}

type WorkflowStep interface {
    Name() string
    Execute(ctx context.Context, input interface{}) (StepResult, error)
    CanExecute(state WorkflowState) bool
}

type WorkflowState interface {
    Get(key string) interface{}
    Set(key string, value interface{})
    GetStepResults() map[string]StepResult
    GetCurrentStep() string
    IsCompleted() bool
}
```

---

## Sequential Workflows

### Basic Sequential Execution

```go
// DocumentProcessingWorkflow processes documents through multiple stages
type DocumentProcessingWorkflow struct {
    extractAgent    *core.LLMAgent
    analyzeAgent    *core.LLMAgent
    summarizeAgent  *core.LLMAgent
    state           *WorkflowState
}

func NewDocumentProcessingWorkflow(provider provider.Provider) *DocumentProcessingWorkflow {
    return &DocumentProcessingWorkflow{
        extractAgent: core.NewLLMAgent("extractor", provider,
            core.WithSystemPrompt("Extract key information from documents"),
            core.WithTools(tools.NewFileReadTool(), tools.NewJSONProcessTool()),
        ),
        analyzeAgent: core.NewLLMAgent("analyzer", provider,
            core.WithSystemPrompt("Analyze extracted information for insights"),
            core.WithTools(tools.NewDataAnalysisTool()),
        ),
        summarizeAgent: core.NewLLMAgent("summarizer", provider,
            core.WithSystemPrompt("Create concise summaries of analysis"),
        ),
        state: NewWorkflowState(),
    }
}

func (w *DocumentProcessingWorkflow) Execute(ctx context.Context, documentPath string) (*ProcessingResult, error) {
    // Step 1: Extract information
    extractResult, err := w.extractAgent.Complete(ctx, &core.CompletionRequest{
        Messages: []core.Message{
            {Role: "user", Content: fmt.Sprintf("Extract key information from document: %s", documentPath)},
        },
    })
    if err != nil {
        return nil, fmt.Errorf("extraction failed: %w", err)
    }
    
    w.state.Set("extracted_data", extractResult.Content)
    
    // Step 2: Analyze extracted data
    analyzeResult, err := w.analyzeAgent.Complete(ctx, &core.CompletionRequest{
        Messages: []core.Message{
            {Role: "user", Content: fmt.Sprintf("Analyze this data: %s", extractResult.Content)},
        },
    })
    if err != nil {
        return nil, fmt.Errorf("analysis failed: %w", err)
    }
    
    w.state.Set("analysis", analyzeResult.Content)
    
    // Step 3: Generate summary
    summaryResult, err := w.summarizeAgent.Complete(ctx, &core.CompletionRequest{
        Messages: []core.Message{
            {Role: "user", Content: fmt.Sprintf("Summarize this analysis: %s", analyzeResult.Content)},
        },
    })
    if err != nil {
        return nil, fmt.Errorf("summarization failed: %w", err)
    }
    
    return &ProcessingResult{
        ExtractedData: extractResult.Content,
        Analysis:      analyzeResult.Content,
        Summary:       summaryResult.Content,
        Metadata:      w.generateMetadata(),
    }, nil
}
```

### Advanced Sequential with Checkpointing

```go
// CheckpointedWorkflow supports resumable execution
type CheckpointedWorkflow struct {
    steps       []WorkflowStep
    state       *PersistentState
    checkpointer Checkpointer
}

type Checkpointer interface {
    Save(ctx context.Context, state WorkflowState) error
    Load(ctx context.Context, workflowID string) (WorkflowState, error)
}

func (w *CheckpointedWorkflow) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // Load previous state if exists
    if savedState, err := w.checkpointer.Load(ctx, w.state.WorkflowID); err == nil {
        w.state = savedState.(*PersistentState)
    }
    
    // Find starting point
    startIndex := 0
    if w.state.LastCompletedStep != "" {
        for i, step := range w.steps {
            if step.Name() == w.state.LastCompletedStep {
                startIndex = i + 1
                break
            }
        }
    }
    
    // Execute remaining steps
    for i := startIndex; i < len(w.steps); i++ {
        step := w.steps[i]
        
        // Check if step can execute
        if !step.CanExecute(w.state) {
            return nil, fmt.Errorf("step %s cannot execute: preconditions not met", step.Name())
        }
        
        // Execute step
        result, err := w.executeStepWithRetry(ctx, step, input)
        if err != nil {
            // Save state before failing
            w.checkpointer.Save(ctx, w.state)
            return nil, fmt.Errorf("step %s failed: %w", step.Name(), err)
        }
        
        // Update state
        w.state.SetStepResult(step.Name(), result)
        w.state.LastCompletedStep = step.Name()
        
        // Save checkpoint
        if err := w.checkpointer.Save(ctx, w.state); err != nil {
            log.Printf("Failed to save checkpoint: %v", err)
        }
        
        // Use result as input for next step
        input = result.Output
    }
    
    w.state.Status = WorkflowCompleted
    return input, nil
}

func (w *CheckpointedWorkflow) executeStepWithRetry(ctx context.Context, step WorkflowStep, input interface{}) (StepResult, error) {
    maxRetries := 3
    backoff := time.Second
    
    for attempt := 0; attempt <= maxRetries; attempt++ {
        result, err := step.Execute(ctx, input)
        if err == nil {
            return result, nil
        }
        
        if !isRetryableError(err) {
            return StepResult{}, err
        }
        
        if attempt < maxRetries {
            time.Sleep(backoff)
            backoff *= 2
        }
    }
    
    return StepResult{}, fmt.Errorf("max retries exceeded")
}
```

---

## Parallel Workflows

### Basic Parallel Execution

```go
// ParallelAnalysisWorkflow analyzes data using multiple agents simultaneously
type ParallelAnalysisWorkflow struct {
    agents      map[string]*core.LLMAgent
    aggregator  *core.LLMAgent
    concurrency int
}

func (w *ParallelAnalysisWorkflow) Execute(ctx context.Context, data interface{}) (*AnalysisResult, error) {
    // Create channels for coordination
    tasks := make(chan AnalysisTask, len(w.agents))
    results := make(chan AnalysisOutput, len(w.agents))
    errors := make(chan error, len(w.agents))
    
    // Create worker pool
    var wg sync.WaitGroup
    for i := 0; i < w.concurrency; i++ {
        wg.Add(1)
        go w.worker(ctx, tasks, results, errors, &wg)
    }
    
    // Queue tasks
    for name, agent := range w.agents {
        tasks <- AnalysisTask{
            AgentName: name,
            Agent:     agent,
            Data:      data,
        }
    }
    close(tasks)
    
    // Wait for completion
    go func() {
        wg.Wait()
        close(results)
        close(errors)
    }()
    
    // Collect results
    outputs := make(map[string]interface{})
    var errs []error
    
    for i := 0; i < len(w.agents); i++ {
        select {
        case result := <-results:
            outputs[result.AgentName] = result.Output
            
        case err := <-errors:
            errs = append(errs, err)
            
        case <-ctx.Done():
            return nil, ctx.Err()
        }
    }
    
    if len(errs) > 0 {
        return nil, fmt.Errorf("parallel execution failed: %v", errs)
    }
    
    // Aggregate results
    aggregated, err := w.aggregateResults(ctx, outputs)
    if err != nil {
        return nil, fmt.Errorf("aggregation failed: %w", err)
    }
    
    return &AnalysisResult{
        IndividualAnalyses: outputs,
        AggregatedInsights: aggregated,
        Timestamp:          time.Now(),
    }, nil
}

func (w *ParallelAnalysisWorkflow) worker(ctx context.Context, tasks <-chan AnalysisTask, results chan<- AnalysisOutput, errors chan<- error, wg *sync.WaitGroup) {
    defer wg.Done()
    
    for task := range tasks {
        output, err := task.Agent.Complete(ctx, &core.CompletionRequest{
            Messages: []core.Message{
                {Role: "user", Content: fmt.Sprintf("Analyze this data: %v", task.Data)},
            },
        })
        
        if err != nil {
            errors <- fmt.Errorf("agent %s failed: %w", task.AgentName, err)
            continue
        }
        
        results <- AnalysisOutput{
            AgentName: task.AgentName,
            Output:    output.Content,
        }
    }
}
```

### Map-Reduce Pattern

```go
// MapReduceWorkflow implements distributed processing pattern
type MapReduceWorkflow struct {
    mapper      MapFunction
    reducer     ReduceFunction
    partitioner Partitioner
    workers     int
}

type MapFunction func(ctx context.Context, agent *core.LLMAgent, input interface{}) ([]KeyValue, error)
type ReduceFunction func(ctx context.Context, agent *core.LLMAgent, key string, values []interface{}) (interface{}, error)
type Partitioner func(data interface{}) []interface{}

func (w *MapReduceWorkflow) Execute(ctx context.Context, input interface{}) (map[string]interface{}, error) {
    // Partition input data
    partitions := w.partitioner(input)
    
    // Map phase
    mapResults := make(chan []KeyValue, len(partitions))
    mapErrors := make(chan error, len(partitions))
    
    var mapWg sync.WaitGroup
    semaphore := make(chan struct{}, w.workers)
    
    for _, partition := range partitions {
        mapWg.Add(1)
        go func(p interface{}) {
            defer mapWg.Done()
            
            semaphore <- struct{}{}
            defer func() { <-semaphore }()
            
            agent := core.NewLLMAgent("mapper", provider,
                core.WithSystemPrompt("Process and extract key-value pairs from data"),
            )
            
            kvPairs, err := w.mapper(ctx, agent, p)
            if err != nil {
                mapErrors <- err
                return
            }
            
            mapResults <- kvPairs
        }(partition)
    }
    
    // Wait for map phase
    mapWg.Wait()
    close(mapResults)
    close(mapErrors)
    
    // Check for errors
    if len(mapErrors) > 0 {
        return nil, <-mapErrors
    }
    
    // Shuffle phase - group by key
    shuffled := make(map[string][]interface{})
    for kvPairs := range mapResults {
        for _, kv := range kvPairs {
            shuffled[kv.Key] = append(shuffled[kv.Key], kv.Value)
        }
    }
    
    // Reduce phase
    reduced := make(map[string]interface{})
    reduceMutex := sync.Mutex{}
    
    var reduceWg sync.WaitGroup
    for key, values := range shuffled {
        reduceWg.Add(1)
        go func(k string, v []interface{}) {
            defer reduceWg.Done()
            
            semaphore <- struct{}{}
            defer func() { <-semaphore }()
            
            agent := core.NewLLMAgent("reducer", provider,
                core.WithSystemPrompt("Aggregate and summarize grouped data"),
            )
            
            result, err := w.reducer(ctx, agent, k, v)
            if err != nil {
                log.Printf("Reduce failed for key %s: %v", k, err)
                return
            }
            
            reduceMutex.Lock()
            reduced[k] = result
            reduceMutex.Unlock()
        }(key, values)
    }
    
    reduceWg.Wait()
    
    return reduced, nil
}

// Example: Document analysis map-reduce
documentAnalyzer := &MapReduceWorkflow{
    mapper: func(ctx context.Context, agent *core.LLMAgent, input interface{}) ([]KeyValue, error) {
        doc := input.(Document)
        
        // Extract entities, topics, sentiment
        result, err := agent.Complete(ctx, &core.CompletionRequest{
            Messages: []core.Message{
                {Role: "user", Content: fmt.Sprintf("Extract entities, topics, and sentiment from: %s", doc.Content)},
            },
        })
        
        if err != nil {
            return nil, err
        }
        
        // Parse result into key-value pairs
        return parseAnalysisResult(result.Content), nil
    },
    
    reducer: func(ctx context.Context, agent *core.LLMAgent, key string, values []interface{}) (interface{}, error) {
        // Aggregate findings
        result, err := agent.Complete(ctx, &core.CompletionRequest{
            Messages: []core.Message{
                {Role: "user", Content: fmt.Sprintf("Summarize these %s findings: %v", key, values)},
            },
        })
        
        if err != nil {
            return nil, err
        }
        
        return result.Content, nil
    },
    
    partitioner: func(data interface{}) []interface{} {
        // Split documents into chunks
        docs := data.([]Document)
        return splitIntoChunks(docs, 10)
    },
    
    workers: 5,
}
```

---

## Conditional Workflows

### Decision Tree Workflow

```go
// DecisionTreeWorkflow implements branching logic
type DecisionTreeWorkflow struct {
    root DecisionNode
}

type DecisionNode interface {
    Evaluate(ctx context.Context, state WorkflowState) (string, error)
    GetChildren() map[string]DecisionNode
    Execute(ctx context.Context, state WorkflowState) error
}

type ConditionalNode struct {
    condition  ConditionFunc
    agent      *core.LLMAgent
    children   map[string]DecisionNode
}

func (n *ConditionalNode) Evaluate(ctx context.Context, state WorkflowState) (string, error) {
    // Execute agent to make decision
    result, err := n.agent.Complete(ctx, &core.CompletionRequest{
        Messages: []core.Message{
            {Role: "user", Content: fmt.Sprintf("Evaluate condition with state: %v", state.GetAll())},
        },
    })
    
    if err != nil {
        return "", err
    }
    
    // Parse decision from result
    decision := parseDecision(result.Content)
    
    // Validate decision
    if _, exists := n.children[decision]; !exists {
        return "", fmt.Errorf("invalid decision: %s", decision)
    }
    
    return decision, nil
}

func (w *DecisionTreeWorkflow) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    state := NewWorkflowState()
    state.Set("input", input)
    
    currentNode := w.root
    path := []string{"root"}
    
    for currentNode != nil {
        // Execute current node
        if err := currentNode.Execute(ctx, state); err != nil {
            return nil, fmt.Errorf("node execution failed at %v: %w", path, err)
        }
        
        // Get children
        children := currentNode.GetChildren()
        if len(children) == 0 {
            // Leaf node reached
            break
        }
        
        // Evaluate next step
        decision, err := currentNode.Evaluate(ctx, state)
        if err != nil {
            return nil, fmt.Errorf("decision failed at %v: %w", path, err)
        }
        
        // Move to next node
        nextNode, exists := children[decision]
        if !exists {
            return nil, fmt.Errorf("no child node for decision %s at %v", decision, path)
        }
        
        currentNode = nextNode
        path = append(path, decision)
        state.Set("decision_path", path)
    }
    
    return state.Get("result"), nil
}

// Example: Customer support routing
supportRouter := &DecisionTreeWorkflow{
    root: &ConditionalNode{
        agent: core.NewLLMAgent("classifier", provider,
            core.WithSystemPrompt("Classify customer support requests"),
        ),
        children: map[string]DecisionNode{
            "technical": &ConditionalNode{
                agent: core.NewLLMAgent("technical_classifier", provider,
                    core.WithSystemPrompt("Classify technical issues"),
                ),
                children: map[string]DecisionNode{
                    "bug": &ActionNode{
                        agent: core.NewLLMAgent("bug_handler", provider,
                            core.WithSystemPrompt("Handle bug reports"),
                        ),
                    },
                    "feature": &ActionNode{
                        agent: core.NewLLMAgent("feature_handler", provider,
                            core.WithSystemPrompt("Handle feature requests"),
                        ),
                    },
                },
            },
            "billing": &ActionNode{
                agent: core.NewLLMAgent("billing_handler", provider,
                    core.WithSystemPrompt("Handle billing inquiries"),
                ),
            },
        },
    },
}
```

### Dynamic Routing Workflow

```go
// DynamicRoutingWorkflow routes based on runtime conditions
type DynamicRoutingWorkflow struct {
    router      RouterAgent
    handlers    map[string]Handler
    fallback    Handler
}

type RouterAgent interface {
    Route(ctx context.Context, input interface{}) (string, float64, error)
}

type Handler interface {
    Handle(ctx context.Context, input interface{}) (interface{}, error)
    CanHandle(input interface{}) bool
}

func (w *DynamicRoutingWorkflow) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // Get routing decision with confidence
    route, confidence, err := w.router.Route(ctx, input)
    if err != nil {
        return nil, fmt.Errorf("routing failed: %w", err)
    }
    
    // Log routing decision
    log.Printf("Routing to %s with confidence %.2f", route, confidence)
    
    // Check confidence threshold
    if confidence < 0.7 {
        log.Printf("Low confidence routing, using fallback")
        if w.fallback != nil {
            return w.fallback.Handle(ctx, input)
        }
        return nil, fmt.Errorf("no confident route found")
    }
    
    // Get handler
    handler, exists := w.handlers[route]
    if !exists {
        if w.fallback != nil {
            return w.fallback.Handle(ctx, input)
        }
        return nil, fmt.Errorf("no handler for route: %s", route)
    }
    
    // Verify handler can process input
    if !handler.CanHandle(input) {
        return nil, fmt.Errorf("handler %s cannot process input", route)
    }
    
    // Execute handler with timeout
    handlerCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
    defer cancel()
    
    result, err := handler.Handle(handlerCtx, input)
    if err != nil {
        // Try fallback on error
        if w.fallback != nil && isRecoverableError(err) {
            log.Printf("Handler failed, trying fallback: %v", err)
            return w.fallback.Handle(ctx, input)
        }
        return nil, err
    }
    
    return result, nil
}

// Intelligent router implementation
type IntelligentRouter struct {
    agent   *core.LLMAgent
    history *RoutingHistory
}

func (r *IntelligentRouter) Route(ctx context.Context, input interface{}) (string, float64, error) {
    // Get historical context
    context := r.history.GetContext(input)
    
    // Ask agent to route
    result, err := r.agent.Complete(ctx, &core.CompletionRequest{
        Messages: []core.Message{
            {Role: "system", Content: "Route inputs to appropriate handlers. Return JSON with 'route' and 'confidence' fields."},
            {Role: "user", Content: fmt.Sprintf("Route this input: %v\nContext: %v", input, context)},
        },
    })
    
    if err != nil {
        return "", 0, err
    }
    
    // Parse routing decision
    var decision struct {
        Route      string  `json:"route"`
        Confidence float64 `json:"confidence"`
        Reasoning  string  `json:"reasoning"`
    }
    
    if err := json.Unmarshal([]byte(result.Content), &decision); err != nil {
        return "", 0, fmt.Errorf("failed to parse routing decision: %w", err)
    }
    
    // Record decision
    r.history.Record(input, decision.Route, decision.Confidence)
    
    return decision.Route, decision.Confidence, nil
}
```

---

## State Management

### Distributed State Management

```go
// DistributedWorkflowState manages state across multiple nodes
type DistributedWorkflowState struct {
    store      StateStore
    localCache map[string]interface{}
    mu         sync.RWMutex
    workflowID string
    version    int64
}

type StateStore interface {
    Get(ctx context.Context, workflowID string, key string) (interface{}, error)
    Set(ctx context.Context, workflowID string, key string, value interface{}) error
    GetAll(ctx context.Context, workflowID string) (map[string]interface{}, error)
    Lock(ctx context.Context, workflowID string, ttl time.Duration) (UnlockFunc, error)
}

func (s *DistributedWorkflowState) Get(key string) interface{} {
    s.mu.RLock()
    if val, ok := s.localCache[key]; ok {
        s.mu.RUnlock()
        return val
    }
    s.mu.RUnlock()
    
    // Fetch from distributed store
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    val, err := s.store.Get(ctx, s.workflowID, key)
    if err != nil {
        log.Printf("Failed to get %s from store: %v", key, err)
        return nil
    }
    
    // Update local cache
    s.mu.Lock()
    s.localCache[key] = val
    s.mu.Unlock()
    
    return val
}

func (s *DistributedWorkflowState) Set(key string, value interface{}) error {
    // Update local cache
    s.mu.Lock()
    s.localCache[key] = value
    s.version++
    s.mu.Unlock()
    
    // Persist to distributed store
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    return s.store.Set(ctx, s.workflowID, key, value)
}

func (s *DistributedWorkflowState) Transaction(fn func(tx StateTransaction) error) error {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Acquire distributed lock
    unlock, err := s.store.Lock(ctx, s.workflowID, 30*time.Second)
    if err != nil {
        return fmt.Errorf("failed to acquire lock: %w", err)
    }
    defer unlock()
    
    // Reload state
    allState, err := s.store.GetAll(ctx, s.workflowID)
    if err != nil {
        return fmt.Errorf("failed to load state: %w", err)
    }
    
    s.mu.Lock()
    s.localCache = allState
    s.mu.Unlock()
    
    // Execute transaction
    tx := &stateTransaction{state: s, changes: make(map[string]interface{})}
    if err := fn(tx); err != nil {
        return err
    }
    
    // Apply changes
    for key, value := range tx.changes {
        if err := s.store.Set(ctx, s.workflowID, key, value); err != nil {
            return fmt.Errorf("failed to apply change %s: %w", key, err)
        }
    }
    
    return nil
}
```

### Event-Driven State Updates

```go
// EventDrivenWorkflow reacts to state changes
type EventDrivenWorkflow struct {
    state       *ReactiveState
    handlers    map[EventType]EventHandler
    subscribers []StateSubscriber
}

type ReactiveState struct {
    data      map[string]interface{}
    events    chan StateEvent
    mu        sync.RWMutex
}

type StateEvent struct {
    Type      EventType
    Key       string
    OldValue  interface{}
    NewValue  interface{}
    Timestamp time.Time
}

func (s *ReactiveState) Set(key string, value interface{}) {
    s.mu.Lock()
    oldValue := s.data[key]
    s.data[key] = value
    s.mu.Unlock()
    
    // Emit event
    event := StateEvent{
        Type:      EventTypeUpdate,
        Key:       key,
        OldValue:  oldValue,
        NewValue:  value,
        Timestamp: time.Now(),
    }
    
    select {
    case s.events <- event:
    default:
        log.Printf("Event channel full, dropping event: %+v", event)
    }
}

func (w *EventDrivenWorkflow) Start(ctx context.Context) {
    for {
        select {
        case event := <-w.state.events:
            w.handleEvent(ctx, event)
            
        case <-ctx.Done():
            return
        }
    }
}

func (w *EventDrivenWorkflow) handleEvent(ctx context.Context, event StateEvent) {
    // Execute handlers
    if handler, ok := w.handlers[event.Type]; ok {
        if err := handler.Handle(ctx, event); err != nil {
            log.Printf("Handler failed for event %v: %v", event.Type, err)
        }
    }
    
    // Notify subscribers
    for _, subscriber := range w.subscribers {
        if subscriber.IsInterested(event) {
            go subscriber.Notify(event)
        }
    }
}

// Example: Reactive document processing
reactiveProcessor := &EventDrivenWorkflow{
    state: NewReactiveState(),
    handlers: map[EventType]EventHandler{
        EventTypeDocumentAdded: &DocumentAddedHandler{
            analyzer: core.NewLLMAgent("analyzer", provider),
        },
        EventTypeAnalysisComplete: &AnalysisCompleteHandler{
            notifier: notificationService,
        },
        EventTypeError: &ErrorHandler{
            retryQueue: retryQueue,
        },
    },
    subscribers: []StateSubscriber{
        &MetricsCollector{},
        &AuditLogger{},
        &ProgressTracker{},
    },
}
```

---

## Error Handling and Recovery

### Saga Pattern Implementation

```go
// SagaWorkflow implements distributed transaction pattern
type SagaWorkflow struct {
    steps        []SagaStep
    compensators map[string]Compensator
}

type SagaStep interface {
    Execute(ctx context.Context, state WorkflowState) error
    GetCompensator() Compensator
}

type Compensator interface {
    Compensate(ctx context.Context, state WorkflowState) error
}

func (s *SagaWorkflow) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    state := NewWorkflowState()
    state.Set("input", input)
    
    completedSteps := []string{}
    
    // Execute steps
    for i, step := range s.steps {
        stepName := fmt.Sprintf("step_%d", i)
        
        if err := step.Execute(ctx, state); err != nil {
            // Compensate in reverse order
            if compensateErr := s.compensate(ctx, completedSteps, state); compensateErr != nil {
                return nil, fmt.Errorf("compensation failed: %v (original error: %w)", compensateErr, err)
            }
            
            return nil, fmt.Errorf("saga failed at %s: %w", stepName, err)
        }
        
        completedSteps = append(completedSteps, stepName)
        s.compensators[stepName] = step.GetCompensator()
    }
    
    return state.Get("result"), nil
}

func (s *SagaWorkflow) compensate(ctx context.Context, steps []string, state WorkflowState) error {
    // Compensate in reverse order
    for i := len(steps) - 1; i >= 0; i-- {
        stepName := steps[i]
        
        if compensator, ok := s.compensators[stepName]; ok {
            if err := compensator.Compensate(ctx, state); err != nil {
                return fmt.Errorf("compensation failed for %s: %w", stepName, err)
            }
        }
    }
    
    return nil
}

// Example: Order processing saga
orderSaga := &SagaWorkflow{
    steps: []SagaStep{
        &ReserveInventoryStep{
            inventoryService: inventoryService,
            compensator: &ReleaseInventoryCompensator{
                inventoryService: inventoryService,
            },
        },
        &ChargePaymentStep{
            paymentService: paymentService,
            compensator: &RefundPaymentCompensator{
                paymentService: paymentService,
            },
        },
        &CreateShipmentStep{
            shippingService: shippingService,
            compensator: &CancelShipmentCompensator{
                shippingService: shippingService,
            },
        },
    },
}
```

### Circuit Breaker Workflow

```go
// CircuitBreakerWorkflow prevents cascading failures
type CircuitBreakerWorkflow struct {
    workflow    Workflow
    breaker     *CircuitBreaker
    fallback    Workflow
    metrics     *BreakerMetrics
}

type CircuitBreaker struct {
    maxFailures  int
    timeout      time.Duration
    resetTimeout time.Duration
    state        BreakerState
    failures     int
    lastFailTime time.Time
    mu           sync.RWMutex
}

type BreakerState int

const (
    BreakerClosed BreakerState = iota
    BreakerOpen
    BreakerHalfOpen
)

func (cb *CircuitBreakerWorkflow) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // Check circuit breaker state
    if !cb.breaker.CanExecute() {
        cb.metrics.RecordRejection()
        
        if cb.fallback != nil {
            return cb.fallback.Execute(ctx, input)
        }
        
        return nil, errors.New("circuit breaker is open")
    }
    
    // Execute with timeout
    execCtx, cancel := context.WithTimeout(ctx, cb.breaker.timeout)
    defer cancel()
    
    start := time.Now()
    result, err := cb.workflow.Execute(execCtx, input)
    duration := time.Since(start)
    
    // Update circuit breaker
    if err != nil {
        cb.breaker.RecordFailure()
        cb.metrics.RecordFailure(duration)
        
        if cb.fallback != nil && cb.breaker.IsOpen() {
            return cb.fallback.Execute(ctx, input)
        }
        
        return nil, err
    }
    
    cb.breaker.RecordSuccess()
    cb.metrics.RecordSuccess(duration)
    
    return result, nil
}

func (cb *CircuitBreaker) CanExecute() bool {
    cb.mu.RLock()
    defer cb.mu.RUnlock()
    
    switch cb.state {
    case BreakerClosed:
        return true
        
    case BreakerOpen:
        if time.Since(cb.lastFailTime) > cb.resetTimeout {
            cb.mu.RUnlock()
            cb.mu.Lock()
            cb.state = BreakerHalfOpen
            cb.failures = 0
            cb.mu.Unlock()
            cb.mu.RLock()
            return true
        }
        return false
        
    case BreakerHalfOpen:
        return true
        
    default:
        return false
    }
}

func (cb *CircuitBreaker) RecordFailure() {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    cb.failures++
    cb.lastFailTime = time.Now()
    
    if cb.failures >= cb.maxFailures {
        cb.state = BreakerOpen
        log.Printf("Circuit breaker opened after %d failures", cb.failures)
    }
}

func (cb *CircuitBreaker) RecordSuccess() {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    if cb.state == BreakerHalfOpen {
        cb.state = BreakerClosed
        cb.failures = 0
        log.Printf("Circuit breaker closed after successful execution")
    }
}
```

---

## Advanced Orchestration Patterns

### Hierarchical Workflows

```go
// HierarchicalWorkflow supports nested sub-workflows
type HierarchicalWorkflow struct {
    name         string
    coordinator  *core.LLMAgent
    subWorkflows map[string]Workflow
    strategy     OrchestrationStrategy
}

type OrchestrationStrategy interface {
    Plan(ctx context.Context, coordinator *core.LLMAgent, input interface{}) ([]WorkflowTask, error)
    Merge(ctx context.Context, coordinator *core.LLMAgent, results map[string]interface{}) (interface{}, error)
}

func (h *HierarchicalWorkflow) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // Get execution plan from coordinator
    tasks, err := h.strategy.Plan(ctx, h.coordinator, input)
    if err != nil {
        return nil, fmt.Errorf("planning failed: %w", err)
    }
    
    // Execute sub-workflows
    results := make(map[string]interface{})
    resultsMu := sync.Mutex{}
    
    // Group tasks by dependency level
    levels := h.groupByDependencyLevel(tasks)
    
    for level, levelTasks := range levels {
        log.Printf("Executing level %d with %d tasks", level, len(levelTasks))
        
        var wg sync.WaitGroup
        errors := make(chan error, len(levelTasks))
        
        for _, task := range levelTasks {
            wg.Add(1)
            go func(t WorkflowTask) {
                defer wg.Done()
                
                // Get sub-workflow
                subWorkflow, ok := h.subWorkflows[t.WorkflowName]
                if !ok {
                    errors <- fmt.Errorf("unknown sub-workflow: %s", t.WorkflowName)
                    return
                }
                
                // Prepare input with dependencies
                taskInput := h.prepareInput(t, input, results)
                
                // Execute sub-workflow
                result, err := subWorkflow.Execute(ctx, taskInput)
                if err != nil {
                    errors <- fmt.Errorf("sub-workflow %s failed: %w", t.WorkflowName, err)
                    return
                }
                
                // Store result
                resultsMu.Lock()
                results[t.Name] = result
                resultsMu.Unlock()
            }(task)
        }
        
        wg.Wait()
        
        // Check for errors
        close(errors)
        for err := range errors {
            if err != nil {
                return nil, err
            }
        }
    }
    
    // Merge results
    return h.strategy.Merge(ctx, h.coordinator, results)
}

// Example: Complex analysis workflow
analysisOrchestrator := &HierarchicalWorkflow{
    name: "comprehensive_analysis",
    coordinator: core.NewLLMAgent("orchestrator", provider,
        core.WithSystemPrompt("Coordinate complex multi-stage analysis workflows"),
    ),
    subWorkflows: map[string]Workflow{
        "data_extraction": dataExtractionWorkflow,
        "statistical_analysis": statisticalWorkflow,
        "ml_prediction": mlPredictionWorkflow,
        "report_generation": reportingWorkflow,
    },
    strategy: &AdaptiveOrchestrationStrategy{
        requirementsAnalyzer: requirementsAgent,
        dependencyResolver:   dependencyAgent,
        resultAggregator:     aggregatorAgent,
    },
}
```

### Event Sourcing Workflow

```go
// EventSourcedWorkflow maintains complete execution history
type EventSourcedWorkflow struct {
    workflow    Workflow
    eventStore  EventStore
    projections map[string]Projection
}

type WorkflowEvent struct {
    ID           string
    WorkflowID   string
    Type         string
    Timestamp    time.Time
    Data         interface{}
    Metadata     map[string]string
}

type EventStore interface {
    Append(ctx context.Context, event WorkflowEvent) error
    GetEvents(ctx context.Context, workflowID string, fromVersion int) ([]WorkflowEvent, error)
    GetSnapshot(ctx context.Context, workflowID string) (*WorkflowSnapshot, error)
    SaveSnapshot(ctx context.Context, snapshot WorkflowSnapshot) error
}

func (es *EventSourcedWorkflow) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    workflowID := generateWorkflowID()
    
    // Record start event
    es.recordEvent(workflowID, "workflow_started", map[string]interface{}{
        "input": input,
    })
    
    // Create instrumented context
    instrumentedCtx := es.createInstrumentedContext(ctx, workflowID)
    
    // Execute workflow
    result, err := es.workflow.Execute(instrumentedCtx, input)
    
    if err != nil {
        es.recordEvent(workflowID, "workflow_failed", map[string]interface{}{
            "error": err.Error(),
        })
        return nil, err
    }
    
    es.recordEvent(workflowID, "workflow_completed", map[string]interface{}{
        "result": result,
    })
    
    // Update projections
    es.updateProjections(workflowID)
    
    return result, nil
}

func (es *EventSourcedWorkflow) Replay(ctx context.Context, workflowID string, toVersion int) (*WorkflowSnapshot, error) {
    // Get snapshot if available
    snapshot, err := es.eventStore.GetSnapshot(ctx, workflowID)
    if err != nil {
        snapshot = &WorkflowSnapshot{
            WorkflowID: workflowID,
            Version:    0,
            State:      NewWorkflowState(),
        }
    }
    
    // Get events after snapshot
    events, err := es.eventStore.GetEvents(ctx, workflowID, snapshot.Version)
    if err != nil {
        return nil, err
    }
    
    // Replay events
    for _, event := range events {
        if event.Version > toVersion {
            break
        }
        
        if err := es.applyEvent(snapshot.State, event); err != nil {
            return nil, fmt.Errorf("failed to apply event %s: %w", event.ID, err)
        }
        
        snapshot.Version = event.Version
    }
    
    return snapshot, nil
}

// Time-travel debugging
func (es *EventSourcedWorkflow) Debug(ctx context.Context, workflowID string, timestamp time.Time) (*DebugInfo, error) {
    // Find version at timestamp
    events, err := es.eventStore.GetEvents(ctx, workflowID, 0)
    if err != nil {
        return nil, err
    }
    
    targetVersion := 0
    for _, event := range events {
        if event.Timestamp.After(timestamp) {
            break
        }
        targetVersion = event.Version
    }
    
    // Replay to target version
    snapshot, err := es.Replay(ctx, workflowID, targetVersion)
    if err != nil {
        return nil, err
    }
    
    return &DebugInfo{
        WorkflowID: workflowID,
        Version:    targetVersion,
        Timestamp:  timestamp,
        State:      snapshot.State,
        Events:     events[:targetVersion],
    }, nil
}
```

---

## Performance Optimization

### Workflow Caching

```go
// CachedWorkflow optimizes repeated executions
type CachedWorkflow struct {
    workflow    Workflow
    cache       WorkflowCache
    keyGen      CacheKeyGenerator
    ttl         time.Duration
}

type WorkflowCache interface {
    Get(key string) (interface{}, bool)
    Set(key string, value interface{}, ttl time.Duration)
    InvalidatePattern(pattern string) error
}

func (c *CachedWorkflow) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // Generate cache key
    key := c.keyGen.GenerateKey(input)
    
    // Check cache
    if cached, found := c.cache.Get(key); found {
        log.Printf("Cache hit for key: %s", key)
        return cached, nil
    }
    
    // Execute workflow
    result, err := c.workflow.Execute(ctx, input)
    if err != nil {
        return nil, err
    }
    
    // Cache result
    c.cache.Set(key, result, c.ttl)
    
    return result, nil
}

// Intelligent cache key generation
type SmartCacheKeyGenerator struct {
    relevantFields []string
    normalizer     DataNormalizer
}

func (g *SmartCacheKeyGenerator) GenerateKey(input interface{}) string {
    // Extract relevant fields
    relevant := g.extractRelevantFields(input)
    
    // Normalize data
    normalized := g.normalizer.Normalize(relevant)
    
    // Generate hash
    h := sha256.New()
    data, _ := json.Marshal(normalized)
    h.Write(data)
    
    return hex.EncodeToString(h.Sum(nil))
}
```

### Workflow Profiling

```go
// ProfiledWorkflow collects performance metrics
type ProfiledWorkflow struct {
    workflow Workflow
    profiler WorkflowProfiler
}

type WorkflowProfiler struct {
    spans    []ProfileSpan
    metrics  map[string]float64
    mu       sync.Mutex
}

type ProfileSpan struct {
    Name      string
    StartTime time.Time
    EndTime   time.Time
    Metadata  map[string]interface{}
}

func (p *ProfiledWorkflow) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // Start profiling
    span := p.profiler.StartSpan("workflow_execution")
    defer span.End()
    
    // Measure input size
    p.profiler.RecordMetric("input_size", float64(getSize(input)))
    
    // Execute with profiling context
    profilingCtx := context.WithValue(ctx, "profiler", p.profiler)
    
    start := time.Now()
    result, err := p.workflow.Execute(profilingCtx, input)
    duration := time.Since(start)
    
    // Record metrics
    p.profiler.RecordMetric("execution_time_ms", duration.Milliseconds())
    p.profiler.RecordMetric("output_size", float64(getSize(result)))
    
    if err != nil {
        p.profiler.RecordMetric("error_count", 1)
        return nil, err
    }
    
    p.profiler.RecordMetric("success_count", 1)
    
    // Generate report
    report := p.profiler.GenerateReport()
    log.Printf("Workflow profile: %+v", report)
    
    return result, nil
}

func (p *WorkflowProfiler) StartSpan(name string) *Span {
    span := &Span{
        profiler: p,
        name:     name,
        start:    time.Now(),
        metadata: make(map[string]interface{}),
    }
    
    p.mu.Lock()
    p.spans = append(p.spans, ProfileSpan{
        Name:      name,
        StartTime: span.start,
        Metadata:  span.metadata,
    })
    p.mu.Unlock()
    
    return span
}
```

---

## Monitoring and Observability

### Workflow Telemetry

```go
// TelemetryWorkflow provides comprehensive observability
type TelemetryWorkflow struct {
    workflow  Workflow
    tracer    trace.Tracer
    meter     metric.Meter
    logger    *zap.Logger
}

func (t *TelemetryWorkflow) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // Start trace
    ctx, span := t.tracer.Start(ctx, "workflow.execute",
        trace.WithAttributes(
            attribute.String("workflow.type", fmt.Sprintf("%T", t.workflow)),
            attribute.Int("input.size", getSize(input)),
        ),
    )
    defer span.End()
    
    // Create metrics
    counter, _ := t.meter.Int64Counter("workflow.executions",
        metric.WithDescription("Total workflow executions"),
    )
    
    histogram, _ := t.meter.Float64Histogram("workflow.duration",
        metric.WithDescription("Workflow execution duration"),
        metric.WithUnit("ms"),
    )
    
    // Log execution start
    t.logger.Info("Workflow execution started",
        zap.String("trace_id", span.SpanContext().TraceID().String()),
        zap.Any("input_sample", t.sampleData(input)),
    )
    
    // Execute
    start := time.Now()
    result, err := t.workflow.Execute(ctx, input)
    duration := time.Since(start)
    
    // Record metrics
    labels := []attribute.KeyValue{
        attribute.String("workflow.type", fmt.Sprintf("%T", t.workflow)),
        attribute.Bool("success", err == nil),
    }
    
    counter.Add(ctx, 1, metric.WithAttributes(labels...))
    histogram.Record(ctx, duration.Milliseconds(), metric.WithAttributes(labels...))
    
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        t.logger.Error("Workflow execution failed",
            zap.Error(err),
            zap.Duration("duration", duration),
        )
        return nil, err
    }
    
    span.SetStatus(codes.Ok, "completed")
    t.logger.Info("Workflow execution completed",
        zap.Duration("duration", duration),
        zap.Any("result_sample", t.sampleData(result)),
    )
    
    return result, nil
}
```

### Health Check System

```go
// HealthCheckableWorkflow monitors workflow health
type HealthCheckableWorkflow struct {
    workflow     Workflow
    healthChecks []HealthCheck
    status       *WorkflowHealth
}

type WorkflowHealth struct {
    Status       HealthStatus
    LastCheck    time.Time
    Issues       []HealthIssue
    Metrics      map[string]float64
    mu           sync.RWMutex
}

func (h *HealthCheckableWorkflow) StartHealthMonitoring(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            h.runHealthChecks(ctx)
            
        case <-ctx.Done():
            return
        }
    }
}

func (h *HealthCheckableWorkflow) runHealthChecks(ctx context.Context) {
    h.status.mu.Lock()
    defer h.status.mu.Unlock()
    
    h.status.Issues = []HealthIssue{}
    h.status.LastCheck = time.Now()
    h.status.Status = HealthStatusHealthy
    
    for _, check := range h.healthChecks {
        result := check.Check(ctx)
        
        if result.Status != HealthStatusHealthy {
            h.status.Issues = append(h.status.Issues, HealthIssue{
                CheckName:   check.Name(),
                Status:      result.Status,
                Message:     result.Message,
                Severity:    result.Severity,
                Timestamp:   time.Now(),
            })
            
            if result.Status == HealthStatusUnhealthy {
                h.status.Status = HealthStatusUnhealthy
            } else if h.status.Status == HealthStatusHealthy {
                h.status.Status = HealthStatusDegraded
            }
        }
        
        // Record metrics
        for k, v := range result.Metrics {
            h.status.Metrics[check.Name()+"."+k] = v
        }
    }
}

// Example health checks
healthChecks := []HealthCheck{
    &DependencyHealthCheck{
        name: "llm_provider",
        checker: func(ctx context.Context) error {
            return provider.HealthCheck(ctx)
        },
    },
    &PerformanceHealthCheck{
        name: "response_time",
        threshold: 5 * time.Second,
        window: 5 * time.Minute,
    },
    &ErrorRateHealthCheck{
        name: "error_rate",
        maxErrorRate: 0.05, // 5%
        window: 10 * time.Minute,
    },
}
```

---

## Best Practices

### 1. Workflow Design
- Keep workflows focused and composable
- Use clear naming conventions
- Document expected inputs/outputs
- Design for idempotency where possible

### 2. State Management
- Minimize shared state between steps
- Use immutable state where possible
- Implement proper state versioning
- Consider event sourcing for complex workflows

### 3. Error Handling
- Implement comprehensive error recovery
- Use compensating transactions for distributed workflows
- Add circuit breakers for external dependencies
- Log errors with full context

### 4. Performance
- Profile workflows to identify bottlenecks
- Implement caching for expensive operations
- Use parallel execution where appropriate
- Monitor resource usage

### 5. Observability
- Add comprehensive logging
- Implement distributed tracing
- Collect meaningful metrics
- Create dashboards for monitoring

### 6. Testing
- Unit test individual workflow steps
- Integration test complete workflows
- Test error scenarios and recovery
- Performance test under load

---

## Workflow Configuration Template

```yaml
# workflow-config.yaml
workflows:
  document_processing:
    type: sequential
    timeout: 5m
    retry_policy:
      max_attempts: 3
      backoff: exponential
      initial_delay: 1s
      
    steps:
      - name: extraction
        agent: document_extractor
        timeout: 1m
        
      - name: analysis
        agent: document_analyzer
        timeout: 2m
        depends_on: [extraction]
        
      - name: summary
        agent: document_summarizer
        timeout: 1m
        depends_on: [analysis]
    
    error_handling:
      on_step_failure: compensate
      on_timeout: abort
      
    monitoring:
      metrics: enabled
      tracing: enabled
      health_checks:
        - response_time
        - error_rate
        
  parallel_analysis:
    type: parallel
    concurrency: 5
    timeout: 10m
    
    agents:
      - sentiment_analyzer
      - entity_extractor
      - topic_classifier
      - language_detector
      
    aggregation:
      strategy: merge_all
      timeout: 30s
      
    caching:
      enabled: true
      ttl: 1h
      key_fields: [content_hash, analysis_type]
```

---

## Next Steps

- **[Multi-Agent Systems](/docs/technical/agents/multi-agent-systems.md)** - Advanced agent coordination
- **[Custom Tools](custom-tools.md)** - Build tools for workflows
- **[State Management](/docs/technical/agents/state-management.md)** - Deep dive into state patterns
- **[Production Deployment](production-deployment.md)** - Deploy workflows at scale
- **[API Reference](/docs/technical/api-reference/agents.md)** - Workflow API documentation