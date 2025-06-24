# Agent Communication: Coordination and Handoffs

> **[Project Root](/) / [Documentation](/docs/) / [User Guide](/docs/user-guide/) / [Guides](/docs/user-guide/guides/) / Agent Communication**

Master the art of multi-agent coordination with sophisticated communication patterns, handoff mechanisms, and collaborative workflows. Build systems where specialized agents work together seamlessly to solve complex problems.

## Why Agent Communication Matters

- **Specialized Expertise** - Leverage different agents for their unique strengths
- **Complex Problem Solving** - Break down tasks across multiple intelligent agents
- **Scalable Architecture** - Distribute workload across agent networks
- **Fault Tolerance** - Graceful degradation when individual agents fail
- **Collaborative Intelligence** - Combine perspectives for better outcomes

## Agent Communication Architecture

![Agent Communication Patterns](../../images/agent-communication.svg)

### Core Communication Patterns
1. **Sequential Handoffs** - Pass results from one agent to the next
2. **Parallel Coordination** - Multiple agents work simultaneously
3. **Hierarchical Delegation** - Manager agents coordinate worker agents
4. **Peer-to-Peer Collaboration** - Direct agent-to-agent communication
5. **Event-Driven Messaging** - Asynchronous communication via events
6. **Consensus Building** - Agents negotiate and agree on solutions

### Communication Mechanisms
| Pattern | Use Case | Complexity | Benefits |
|---------|----------|------------|----------|
| **Direct Transfer** | Simple handoffs | Low | Fast, reliable |
| **Message Passing** | Async coordination | Medium | Scalable, decoupled |
| **Shared State** | Collaborative work | Medium | Consistent, transparent |
| **Event Bus** | Complex workflows | High | Flexible, extensible |
| **Negotiation** | Conflict resolution | High | Intelligent, adaptive |

## Prerequisites

- [Creating Agents completed](creating-agents.md) ✅
- [Agent Tools understanding](agent-tools.md) ✅
- Basic knowledge of concurrency patterns ✅

---

## Level 1: Sequential Agent Handoffs
*Chain agents together for step-by-step processing*

### Basic Agent Pipeline
```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
)

// AgentPipeline represents a sequence of agents that process data in order
type AgentPipeline struct {
    name        string
    agents      []PipelineStage
    globalState *domain.State
}

type PipelineStage struct {
    Name        string
    Agent       domain.BaseAgent
    InputKey    string
    OutputKey   string
    Transform   TransformFunc
    Validator   ValidatorFunc
    Timeout     time.Duration
    Retries     int
}

type TransformFunc func(input interface{}) interface{}
type ValidatorFunc func(output interface{}) error

type PipelineResult struct {
    Success     bool
    StageResults map[string]StageResult
    FinalOutput interface{}
    TotalTime   time.Duration
    Error       error
}

type StageResult struct {
    Success     bool
    Input       interface{}
    Output      interface{}
    Duration    time.Duration
    Attempts    int
    Error       error
}

func NewAgentPipeline(name string) *AgentPipeline {
    return &AgentPipeline{
        name:        name,
        agents:      make([]PipelineStage, 0),
        globalState: domain.NewState(),
    }
}

func (ap *AgentPipeline) AddStage(name, agentProvider, inputKey, outputKey string) *AgentPipeline {
    agent, err := core.NewAgentFromString(fmt.Sprintf("%s-%s", ap.name, name), agentProvider)
    if err != nil {
        log.Printf("Failed to create agent for stage %s: %v", name, err)
        return ap
    }

    stage := PipelineStage{
        Name:      name,
        Agent:     agent,
        InputKey:  inputKey,
        OutputKey: outputKey,
        Timeout:   30 * time.Second,
        Retries:   2,
    }

    ap.agents = append(ap.agents, stage)
    return ap
}

func (ap *AgentPipeline) SetStagePrompt(stageName, prompt string) *AgentPipeline {
    for i, stage := range ap.agents {
        if stage.Name == stageName {
            stage.Agent.SetSystemPrompt(prompt)
            ap.agents[i] = stage
            break
        }
    }
    return ap
}

func (ap *AgentPipeline) SetStageTransform(stageName string, transform TransformFunc) *AgentPipeline {
    for i, stage := range ap.agents {
        if stage.Name == stageName {
            stage.Transform = transform
            ap.agents[i] = stage
            break
        }
    }
    return ap
}

func (ap *AgentPipeline) SetStageValidator(stageName string, validator ValidatorFunc) *AgentPipeline {
    for i, stage := range ap.agents {
        if stage.Name == stageName {
            stage.Validator = validator
            ap.agents[i] = stage
            break
        }
    }
    return ap
}

func (ap *AgentPipeline) Execute(ctx context.Context, initialInput interface{}) (*PipelineResult, error) {
    fmt.Printf("🔄 Starting pipeline: %s with %d stages\n", ap.name, len(ap.agents))
    
    result := &PipelineResult{
        Success:      true,
        StageResults: make(map[string]StageResult),
        TotalTime:    0,
    }

    startTime := time.Now()
    currentInput := initialInput

    // Set initial input in global state
    ap.globalState.Set("pipeline_input", initialInput)
    ap.globalState.Set("pipeline_start_time", startTime)

    for i, stage := range ap.agents {
        fmt.Printf("  📍 Stage %d: %s\n", i+1, stage.Name)
        
        stageResult, err := ap.executeStage(ctx, stage, currentInput)
        result.StageResults[stage.Name] = stageResult

        if err != nil {
            result.Success = false
            result.Error = fmt.Errorf("stage %s failed: %w", stage.Name, err)
            fmt.Printf("  ❌ Stage %s failed: %v\n", stage.Name, err)
            break
        }

        fmt.Printf("  ✅ Stage %s completed in %v\n", stage.Name, stageResult.Duration)

        // Pass output to next stage
        currentInput = stageResult.Output

        // Update global state with stage output
        ap.globalState.Set(stage.OutputKey, stageResult.Output)
        ap.globalState.Set(fmt.Sprintf("%s_duration", stage.Name), stageResult.Duration)
    }

    result.TotalTime = time.Since(startTime)
    result.FinalOutput = currentInput

    if result.Success {
        fmt.Printf("✅ Pipeline %s completed successfully in %v\n", ap.name, result.TotalTime)
    } else {
        fmt.Printf("❌ Pipeline %s failed after %v\n", ap.name, result.TotalTime)
    }

    return result, result.Error
}

func (ap *AgentPipeline) executeStage(ctx context.Context, stage PipelineStage, input interface{}) (StageResult, error) {
    stageResult := StageResult{
        Input:    input,
        Attempts: 0,
    }

    // Apply input transformation if provided
    processedInput := input
    if stage.Transform != nil {
        processedInput = stage.Transform(input)
    }

    // Retry logic
    for attempt := 0; attempt < stage.Retries+1; attempt++ {
        stageResult.Attempts = attempt + 1
        
        // Create timeout context for this stage
        stageCtx, cancel := context.WithTimeout(ctx, stage.Timeout)
        defer cancel()

        startTime := time.Now()
        output, err := ap.runStageAgent(stageCtx, stage, processedInput)
        stageResult.Duration = time.Since(startTime)

        if err != nil {
            stageResult.Error = err
            if attempt < stage.Retries {
                fmt.Printf("    🔄 Retrying stage %s (attempt %d/%d)\n", 
                    stage.Name, attempt+2, stage.Retries+1)
                time.Sleep(time.Duration(attempt+1) * time.Second) // Exponential backoff
                continue
            }
            return stageResult, err
        }

        // Validate output if validator provided
        if stage.Validator != nil {
            if validateErr := stage.Validator(output); validateErr != nil {
                stageResult.Error = validateErr
                if attempt < stage.Retries {
                    fmt.Printf("    ⚠️ Validation failed for stage %s, retrying\n", stage.Name)
                    continue
                }
                return stageResult, validateErr
            }
        }

        stageResult.Success = true
        stageResult.Output = output
        return stageResult, nil
    }

    return stageResult, fmt.Errorf("stage failed after %d attempts", stage.Retries+1)
}

func (ap *AgentPipeline) runStageAgent(ctx context.Context, stage PipelineStage, input interface{}) (interface{}, error) {
    // Create state for this stage
    stageState := domain.NewState()
    
    // Copy relevant global state
    if ap.globalState != nil {
        // Add global context to stage state
        stageState.Set("pipeline_context", ap.globalState.GetAll())
    }

    // Set the stage input
    stageState.Set(stage.InputKey, input)
    stageState.Set("stage_name", stage.Name)
    
    // Execute the agent
    result, err := stage.Agent.Run(ctx, stageState)
    if err != nil {
        return nil, err
    }

    // Extract the output
    output, exists := result.Get(stage.OutputKey)
    if !exists {
        // Try default output key
        if response, exists := result.Get("response"); exists {
            output = response
        } else {
            return nil, fmt.Errorf("stage %s did not produce expected output key: %s", stage.Name, stage.OutputKey)
        }
    }

    return output, nil
}

func (ap *AgentPipeline) GetGlobalState() *domain.State {
    return ap.globalState
}

func main() {
    fmt.Println("🔗 Sequential Agent Handoffs")
    fmt.Println("============================")

    // Create a content creation pipeline
    pipeline := NewAgentPipeline("content-creation").
        AddStage("researcher", "anthropic/claude-3-5-sonnet", "topic", "research_data").
        AddStage("outliner", "openai/gpt-4o-mini", "research_data", "outline").
        AddStage("writer", "anthropic/claude-3-5-sonnet", "outline", "draft").
        AddStage("editor", "openai/gpt-4o-mini", "draft", "final_content")

    // Configure each stage with specialized prompts
    pipeline.SetStagePrompt("researcher", `You are a research specialist.
    Given a topic, conduct thorough research and provide:
    1. Key facts and statistics
    2. Current trends and developments
    3. Expert opinions and quotes
    4. Relevant examples and case studies
    
    Format your research as structured data that can be used by other agents.`)

    pipeline.SetStagePrompt("outliner", `You are a content outliner.
    Given research data, create a detailed outline for a comprehensive article:
    1. Compelling headline
    2. Introduction hook
    3. Main sections with subpoints
    4. Conclusion strategy
    
    Ensure logical flow and engaging structure.`)

    pipeline.SetStagePrompt("writer", `You are a skilled content writer.
    Given an outline, write a complete, engaging article:
    1. Follow the outline structure
    2. Use clear, compelling language
    3. Include specific examples and data
    4. Maintain consistent tone
    
    Write in a professional yet accessible style.`)

    pipeline.SetStagePrompt("editor", `You are a professional editor.
    Given a draft article, improve it by:
    1. Fixing grammar and clarity issues
    2. Enhancing readability and flow
    3. Strengthening arguments
    4. Ensuring consistency
    
    Return the polished final version.`)

    // Add transformations and validators
    pipeline.SetStageTransform("outliner", func(input interface{}) interface{} {
        // Transform research data into outline-friendly format
        return fmt.Sprintf("Research data for outline creation:\n%v\n\nCreate a detailed outline based on this research.", input)
    })

    pipeline.SetStageValidator("researcher", func(output interface{}) error {
        // Validate that research contains key elements
        outputStr := fmt.Sprintf("%v", output)
        if len(outputStr) < 100 {
            return fmt.Errorf("research output too short, expected comprehensive research")
        }
        return nil
    })

    // Test the pipeline with different topics
    topics := []string{
        "The impact of artificial intelligence on remote work productivity",
        "Sustainable energy solutions for urban development",
        "The psychology of user experience design",
    }

    ctx := context.Background()

    for i, topic := range topics {
        fmt.Printf("\n🎯 Pipeline Test %d: %s\n", i+1, topic)
        fmt.Printf("Topic: %s\n", topic)

        result, err := pipeline.Execute(ctx, topic)
        if err != nil {
            fmt.Printf("❌ Pipeline failed: %v\n", err)
            continue
        }

        // Display results
        fmt.Printf("\n📊 Pipeline Results:\n")
        fmt.Printf("Total Time: %v\n", result.TotalTime)
        fmt.Printf("Success: %t\n", result.Success)

        fmt.Printf("\n📋 Stage Breakdown:\n")
        for stageName, stageResult := range result.StageResults {
            status := "✅"
            if !stageResult.Success {
                status = "❌"
            }
            fmt.Printf("  %s %s: %v (attempts: %d)\n", 
                status, stageName, stageResult.Duration, stageResult.Attempts)
        }

        // Show final output preview
        if result.FinalOutput != nil {
            finalStr := fmt.Sprintf("%v", result.FinalOutput)
            if len(finalStr) > 300 {
                finalStr = finalStr[:300] + "..."
            }
            fmt.Printf("\n📝 Final Content Preview:\n%s\n", finalStr)
        }

        // Show global state summary
        globalState := pipeline.GetGlobalState()
        if globalState != nil {
            fmt.Printf("\n🌐 Pipeline State Summary:\n")
            allState := globalState.GetAll()
            for key, value := range allState {
                if key != "pipeline_context" { // Skip recursive display
                    fmt.Printf("  %s: %v\n", key, value)
                }
            }
        }
    }
}
```

### Key Features
✅ **Sequential Processing** - Step-by-step agent coordination  
✅ **State Management** - Global and stage-specific state handling  
✅ **Error Recovery** - Retry logic and graceful failure handling  
✅ **Validation** - Output validation between stages  
✅ **Monitoring** - Detailed execution tracking and metrics  

---

## Level 2: Parallel Agent Coordination
*Coordinate multiple agents working simultaneously*

### Parallel Agent Orchestra
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

// ParallelAgentOrchestra coordinates multiple agents working simultaneously
type ParallelAgentOrchestra struct {
    name           string
    agents         map[string]*AgentWorker
    coordinator    *CoordinatorAgent
    resultAggregator *ResultAggregator
    communicationHub *CommunicationHub
    
    maxConcurrency int
    timeout        time.Duration
    
    mutex          sync.RWMutex
}

type AgentWorker struct {
    ID             string
    Name           string
    Agent          domain.BaseAgent
    Specialization string
    Capabilities   []string
    CurrentTask    *Task
    Status         WorkerStatus
    Metrics        WorkerMetrics
    
    mutex          sync.RWMutex
}

type WorkerStatus string

const (
    StatusIdle    WorkerStatus = "idle"
    StatusWorking WorkerStatus = "working"
    StatusError   WorkerStatus = "error"
    StatusOffline WorkerStatus = "offline"
)

type WorkerMetrics struct {
    TasksCompleted   int64
    AverageTime      time.Duration
    SuccessRate      float64
    LastActive       time.Time
}

type Task struct {
    ID           string
    Type         string
    Priority     int
    Input        interface{}
    Requirements TaskRequirements
    Deadline     time.Time
    AssignedTo   string
    Status       TaskStatus
    Result       *TaskResult
    
    CreatedAt    time.Time
    StartedAt    time.Time
    CompletedAt  time.Time
}

type TaskRequirements struct {
    Specialization []string
    MinCapability  []string
    MaxDuration    time.Duration
    QualityLevel   string
}

type TaskStatus string

const (
    TaskStatusPending    TaskStatus = "pending"
    TaskStatusAssigned   TaskStatus = "assigned"
    TaskStatusInProgress TaskStatus = "in_progress"
    TaskStatusCompleted  TaskStatus = "completed"
    TaskStatusFailed     TaskStatus = "failed"
)

type TaskResult struct {
    Success     bool
    Output      interface{}
    Duration    time.Duration
    Quality     float64
    Confidence  float64
    Metadata    map[string]interface{}
    Error       error
}

// CoordinatorAgent manages task distribution and agent coordination
type CoordinatorAgent struct {
    agent          domain.BaseAgent
    taskQueue      chan *Task
    resultChannel  chan *TaskResult
    assignments    map[string]*Task
    
    mutex          sync.RWMutex
}

// ResultAggregator combines results from multiple parallel agents
type ResultAggregator struct {
    strategies     map[string]AggregationStrategy
    currentStrategy string
}

type AggregationStrategy interface {
    Aggregate(results []*TaskResult) (*AggregatedResult, error)
    Name() string
}

type AggregatedResult struct {
    CombinedOutput  interface{}
    IndividualResults []*TaskResult
    AggregationMethod string
    QualityScore    float64
    Confidence      float64
    Metadata        map[string]interface{}
}

// CommunicationHub handles inter-agent messaging
type CommunicationHub struct {
    messageChannels map[string]chan *Message
    subscribers     map[string][]string
    messageHistory  []*Message
    
    mutex           sync.RWMutex
}

type Message struct {
    ID          string
    From        string
    To          string
    Type        MessageType
    Content     interface{}
    Priority    int
    Timestamp   time.Time
    Response    chan *Message
}

type MessageType string

const (
    MessageTypeRequest     MessageType = "request"
    MessageTypeResponse    MessageType = "response"
    MessageTypeNotification MessageType = "notification"
    MessageTypeBroadcast   MessageType = "broadcast"
)

func NewParallelAgentOrchestra(name string, maxConcurrency int) *ParallelAgentOrchestra {
    orchestra := &ParallelAgentOrchestra{
        name:             name,
        agents:           make(map[string]*AgentWorker),
        resultAggregator: NewResultAggregator(),
        communicationHub: NewCommunicationHub(),
        maxConcurrency:   maxConcurrency,
        timeout:          60 * time.Second,
    }

    // Create coordinator agent
    coordinatorAgent, err := core.NewAgentFromString("coordinator", "anthropic/claude-3-5-sonnet")
    if err != nil {
        log.Printf("Failed to create coordinator agent: %v", err)
    } else {
        coordinatorAgent.SetSystemPrompt(`You are a task coordination specialist.
        Your job is to:
        1. Analyze incoming tasks and requirements
        2. Assign tasks to the most suitable agents
        3. Monitor progress and resolve conflicts
        4. Optimize resource allocation
        5. Ensure quality and deadlines are met
        
        Consider agent specializations, current workload, and task priorities.`)
        
        orchestra.coordinator = &CoordinatorAgent{
            agent:         coordinatorAgent,
            taskQueue:     make(chan *Task, 100),
            resultChannel: make(chan *TaskResult, 100),
            assignments:   make(map[string]*Task),
        }
    }

    return orchestra
}

func (orchestra *ParallelAgentOrchestra) AddWorker(name, agentProvider, specialization string, capabilities []string) error {
    agent, err := core.NewAgentFromString(fmt.Sprintf("%s-worker-%s", orchestra.name, name), agentProvider)
    if err != nil {
        return fmt.Errorf("failed to create worker agent %s: %w", name, err)
    }

    worker := &AgentWorker{
        ID:             fmt.Sprintf("worker-%s-%d", name, time.Now().Unix()),
        Name:           name,
        Agent:          agent,
        Specialization: specialization,
        Capabilities:   capabilities,
        Status:         StatusIdle,
        Metrics: WorkerMetrics{
            LastActive: time.Now(),
        },
    }

    orchestra.mutex.Lock()
    orchestra.agents[worker.ID] = worker
    orchestra.mutex.Unlock()

    fmt.Printf("👥 Added worker: %s (Specialization: %s, Capabilities: %v)\n", 
        name, specialization, capabilities)

    return nil
}

func (orchestra *ParallelAgentOrchestra) SetWorkerPrompt(workerName, prompt string) error {
    orchestra.mutex.RLock()
    defer orchestra.mutex.RUnlock()

    for _, worker := range orchestra.agents {
        if worker.Name == workerName {
            worker.Agent.SetSystemPrompt(prompt)
            return nil
        }
    }

    return fmt.Errorf("worker %s not found", workerName)
}

func (orchestra *ParallelAgentOrchestra) ExecuteParallelTasks(ctx context.Context, tasks []*Task) (*ParallelExecutionResult, error) {
    fmt.Printf("🎭 Starting parallel execution with %d tasks\n", len(tasks))

    result := &ParallelExecutionResult{
        TotalTasks:     len(tasks),
        StartTime:      time.Now(),
        TaskResults:    make(map[string]*TaskResult),
        WorkerMetrics:  make(map[string]WorkerMetrics),
    }

    // Start workers
    orchestra.startWorkers(ctx)

    // Submit tasks
    for _, task := range tasks {
        task.Status = TaskStatusPending
        task.CreatedAt = time.Now()
        orchestra.coordinator.taskQueue <- task
    }

    // Wait for completion or timeout
    completedTasks := 0
    timeoutCtx, cancel := context.WithTimeout(ctx, orchestra.timeout)
    defer cancel()

    for completedTasks < len(tasks) {
        select {
        case taskResult := <-orchestra.coordinator.resultChannel:
            completedTasks++
            result.TaskResults[taskResult.Task.ID] = taskResult
            
            if taskResult.Success {
                result.SuccessfulTasks++
            } else {
                result.FailedTasks++
            }

            fmt.Printf("📋 Task completed: %s (Success: %t, Duration: %v)\n", 
                taskResult.Task.ID, taskResult.Success, taskResult.Duration)

        case <-timeoutCtx.Done():
            result.TimedOut = true
            fmt.Printf("⏱️ Parallel execution timed out after %v\n", orchestra.timeout)
            break
        }
    }

    result.Duration = time.Since(result.StartTime)

    // Aggregate results if multiple successful tasks
    if result.SuccessfulTasks > 1 {
        successfulResults := make([]*TaskResult, 0)
        for _, taskResult := range result.TaskResults {
            if taskResult.Success {
                successfulResults = append(successfulResults, taskResult)
            }
        }

        aggregated, err := orchestra.resultAggregator.AggregateResults(successfulResults)
        if err != nil {
            log.Printf("Result aggregation failed: %v", err)
        } else {
            result.AggregatedResult = aggregated
        }
    }

    // Collect worker metrics
    orchestra.mutex.RLock()
    for id, worker := range orchestra.agents {
        result.WorkerMetrics[id] = worker.Metrics
    }
    orchestra.mutex.RUnlock()

    fmt.Printf("✅ Parallel execution completed: %d/%d successful in %v\n", 
        result.SuccessfulTasks, result.TotalTasks, result.Duration)

    return result, nil
}

func (orchestra *ParallelAgentOrchestra) startWorkers(ctx context.Context) {
    fmt.Printf("🚀 Starting %d workers\n", len(orchestra.agents))

    for _, worker := range orchestra.agents {
        go orchestra.runWorker(ctx, worker)
    }

    // Start coordinator
    go orchestra.runCoordinator(ctx)
}

func (orchestra *ParallelAgentOrchestra) runWorker(ctx context.Context, worker *AgentWorker) {
    for {
        select {
        case <-ctx.Done():
            worker.mutex.Lock()
            worker.Status = StatusOffline
            worker.mutex.Unlock()
            return

        default:
            // Check for assigned tasks
            if worker.CurrentTask != nil && worker.CurrentTask.Status == TaskStatusAssigned {
                orchestra.executeTask(ctx, worker, worker.CurrentTask)
            }
            
            time.Sleep(100 * time.Millisecond) // Small delay to prevent busy waiting
        }
    }
}

func (orchestra *ParallelAgentOrchestra) runCoordinator(ctx context.Context) {
    for {
        select {
        case task := <-orchestra.coordinator.taskQueue:
            // Find best worker for this task
            worker := orchestra.findBestWorker(task)
            if worker != nil {
                orchestra.assignTask(worker, task)
            } else {
                // No suitable worker available, mark task as failed
                result := &TaskResult{
                    Task:    task,
                    Success: false,
                    Error:   fmt.Errorf("no suitable worker available"),
                }
                orchestra.coordinator.resultChannel <- result
            }

        case <-ctx.Done():
            return
        }
    }
}

func (orchestra *ParallelAgentOrchestra) findBestWorker(task *Task) *AgentWorker {
    orchestra.mutex.RLock()
    defer orchestra.mutex.RUnlock()

    var bestWorker *AgentWorker
    bestScore := -1.0

    for _, worker := range orchestra.agents {
        if worker.Status != StatusIdle {
            continue
        }

        score := orchestra.calculateWorkerScore(worker, task)
        if score > bestScore {
            bestScore = score
            bestWorker = worker
        }
    }

    return bestWorker
}

func (orchestra *ParallelAgentOrchestra) calculateWorkerScore(worker *AgentWorker, task *Task) float64 {
    score := 0.0

    // Check specialization match
    for _, reqSpec := range task.Requirements.Specialization {
        if worker.Specialization == reqSpec {
            score += 5.0
        }
    }

    // Check capability match
    for _, reqCap := range task.Requirements.MinCapability {
        for _, workerCap := range worker.Capabilities {
            if workerCap == reqCap {
                score += 2.0
            }
        }
    }

    // Factor in success rate
    score += worker.Metrics.SuccessRate * 3.0

    // Factor in average response time
    if worker.Metrics.AverageTime > 0 && task.Requirements.MaxDuration > 0 {
        if worker.Metrics.AverageTime <= task.Requirements.MaxDuration {
            score += 2.0
        } else {
            score -= 1.0
        }
    }

    return score
}

func (orchestra *ParallelAgentOrchestra) assignTask(worker *AgentWorker, task *Task) {
    worker.mutex.Lock()
    worker.CurrentTask = task
    worker.Status = StatusWorking
    worker.mutex.Unlock()

    task.Status = TaskStatusAssigned
    task.AssignedTo = worker.ID
    task.StartedAt = time.Now()

    orchestra.coordinator.mutex.Lock()
    orchestra.coordinator.assignments[task.ID] = task
    orchestra.coordinator.mutex.Unlock()

    fmt.Printf("📋 Assigned task %s to worker %s\n", task.ID, worker.Name)
}

func (orchestra *ParallelAgentOrchestra) executeTask(ctx context.Context, worker *AgentWorker, task *Task) {
    task.Status = TaskStatusInProgress
    startTime := time.Now()

    // Create execution context
    state := domain.NewState()
    state.Set("task_input", task.Input)
    state.Set("task_type", task.Type)
    state.Set("task_requirements", task.Requirements)
    state.Set("worker_id", worker.ID)
    state.Set("worker_specialization", worker.Specialization)

    // Execute the task
    result, err := worker.Agent.Run(ctx, state)
    duration := time.Since(startTime)

    // Create task result
    taskResult := &TaskResult{
        Task:     task,
        Duration: duration,
        Metadata: make(map[string]interface{}),
    }

    if err != nil {
        taskResult.Success = false
        taskResult.Error = err
        task.Status = TaskStatusFailed
        worker.Status = StatusError
    } else {
        taskResult.Success = true
        taskResult.Output = result.Get("response")
        taskResult.Quality = orchestra.assessQuality(result)
        taskResult.Confidence = orchestra.assessConfidence(result)
        task.Status = TaskStatusCompleted
        worker.Status = StatusIdle
    }

    task.Result = taskResult
    task.CompletedAt = time.Now()

    // Update worker metrics
    worker.mutex.Lock()
    worker.Metrics.TasksCompleted++
    if worker.Metrics.AverageTime == 0 {
        worker.Metrics.AverageTime = duration
    } else {
        worker.Metrics.AverageTime = (worker.Metrics.AverageTime + duration) / 2
    }
    
    successCount := float64(worker.Metrics.TasksCompleted)
    if taskResult.Success {
        worker.Metrics.SuccessRate = (worker.Metrics.SuccessRate*(successCount-1) + 1.0) / successCount
    } else {
        worker.Metrics.SuccessRate = (worker.Metrics.SuccessRate * (successCount - 1)) / successCount
    }
    
    worker.Metrics.LastActive = time.Now()
    worker.CurrentTask = nil
    worker.mutex.Unlock()

    // Send result back
    orchestra.coordinator.resultChannel <- taskResult
}

func (orchestra *ParallelAgentOrchestra) assessQuality(result *domain.State) float64 {
    // Simplified quality assessment
    if response, exists := result.Get("response"); exists {
        responseStr := fmt.Sprintf("%v", response)
        if len(responseStr) > 100 {
            return 0.8
        } else if len(responseStr) > 50 {
            return 0.6
        } else {
            return 0.4
        }
    }
    return 0.0
}

func (orchestra *ParallelAgentOrchestra) assessConfidence(result *domain.State) float64 {
    // Simplified confidence assessment
    return 0.7 + (float64(time.Now().Unix()%30) / 100.0) // Mock confidence 0.7-1.0
}

// Supporting types and implementations
type ParallelExecutionResult struct {
    TotalTasks       int
    SuccessfulTasks  int
    FailedTasks      int
    TimedOut         bool
    StartTime        time.Time
    Duration         time.Duration
    TaskResults      map[string]*TaskResult
    AggregatedResult *AggregatedResult
    WorkerMetrics    map[string]WorkerMetrics
}

// Simple aggregation strategy
type ConsensusAggregation struct{}

func (ca *ConsensusAggregation) Aggregate(results []*TaskResult) (*AggregatedResult, error) {
    if len(results) == 0 {
        return nil, fmt.Errorf("no results to aggregate")
    }

    // Simple majority consensus
    outputs := make(map[string]int)
    var bestOutput string
    maxCount := 0

    for _, result := range results {
        outputStr := fmt.Sprintf("%v", result.Output)
        outputs[outputStr]++
        if outputs[outputStr] > maxCount {
            maxCount = outputs[outputStr]
            bestOutput = outputStr
        }
    }

    aggregated := &AggregatedResult{
        CombinedOutput:    bestOutput,
        IndividualResults: results,
        AggregationMethod: "consensus",
        QualityScore:      float64(maxCount) / float64(len(results)),
        Confidence:        0.8,
        Metadata: map[string]interface{}{
            "total_results": len(results),
            "consensus_count": maxCount,
        },
    }

    return aggregated, nil
}

func (ca *ConsensusAggregation) Name() string {
    return "consensus"
}

func NewResultAggregator() *ResultAggregator {
    aggregator := &ResultAggregator{
        strategies: make(map[string]AggregationStrategy),
        currentStrategy: "consensus",
    }
    
    aggregator.strategies["consensus"] = &ConsensusAggregation{}
    
    return aggregator
}

func (ra *ResultAggregator) AggregateResults(results []*TaskResult) (*AggregatedResult, error) {
    strategy, exists := ra.strategies[ra.currentStrategy]
    if !exists {
        return nil, fmt.Errorf("aggregation strategy %s not found", ra.currentStrategy)
    }
    
    return strategy.Aggregate(results)
}

func NewCommunicationHub() *CommunicationHub {
    return &CommunicationHub{
        messageChannels: make(map[string]chan *Message),
        subscribers:     make(map[string][]string),
        messageHistory:  make([]*Message, 0),
    }
}

func main() {
    fmt.Println("🎭 Parallel Agent Coordination")
    fmt.Println("==============================")

    // Create parallel agent orchestra
    orchestra := NewParallelAgentOrchestra("content-team", 4)

    // Add specialized workers
    err := orchestra.AddWorker("researcher", "anthropic/claude-3-5-sonnet", "research", 
        []string{"web_search", "data_analysis", "fact_checking"})
    if err != nil {
        log.Printf("Failed to add researcher: %v", err)
    }

    err = orchestra.AddWorker("writer", "openai/gpt-4o-mini", "writing", 
        []string{"content_creation", "storytelling", "copywriting"})
    if err != nil {
        log.Printf("Failed to add writer: %v", err)
    }

    err = orchestra.AddWorker("analyst", "gemini/gemini-2.0-flash", "analysis", 
        []string{"data_analysis", "trend_analysis", "insights"})
    if err != nil {
        log.Printf("Failed to add analyst: %v", err)
    }

    err = orchestra.AddWorker("editor", "anthropic/claude-3-5-haiku", "editing", 
        []string{"proofreading", "style_checking", "fact_verification"})
    if err != nil {
        log.Printf("Failed to add editor: %v", err)
    }

    // Set specialized prompts
    orchestra.SetWorkerPrompt("researcher", `You are a research specialist.
    Conduct thorough research on the given topic and provide:
    1. Key facts and current data
    2. Expert opinions and quotes
    3. Relevant statistics and trends
    4. Credible sources and references
    
    Focus on accuracy and comprehensiveness.`)

    orchestra.SetWorkerPrompt("writer", `You are a content writer.
    Create engaging, well-structured content based on the input:
    1. Compelling introduction
    2. Clear main points
    3. Supporting examples
    4. Strong conclusion
    
    Write in an accessible, professional tone.`)

    orchestra.SetWorkerPrompt("analyst", `You are a data analyst.
    Analyze the provided information and deliver:
    1. Key insights and patterns
    2. Trend analysis
    3. Recommendations
    4. Data-driven conclusions
    
    Focus on actionable insights.`)

    orchestra.SetWorkerPrompt("editor", `You are an editor.
    Review and improve the given content:
    1. Fix grammar and clarity
    2. Improve structure and flow
    3. Verify facts and consistency
    4. Enhance readability
    
    Maintain the original voice while improving quality.`)

    // Create parallel tasks
    tasks := []*Task{
        {
            ID:   "research-ai-trends",
            Type: "research",
            Priority: 2,
            Input: "Latest trends in artificial intelligence for 2025",
            Requirements: TaskRequirements{
                Specialization: []string{"research"},
                MinCapability:  []string{"web_search", "data_analysis"},
                MaxDuration:    30 * time.Second,
                QualityLevel:   "high",
            },
            Deadline: time.Now().Add(60 * time.Second),
        },
        {
            ID:   "write-ai-article",
            Type: "writing",
            Priority: 1,
            Input: "Write an article about the benefits of AI in healthcare",
            Requirements: TaskRequirements{
                Specialization: []string{"writing"},
                MinCapability:  []string{"content_creation"},
                MaxDuration:    45 * time.Second,
                QualityLevel:   "high",
            },
            Deadline: time.Now().Add(90 * time.Second),
        },
        {
            ID:   "analyze-market-data",
            Type: "analysis",
            Priority: 2,
            Input: "Analyze the impact of remote work on the tech industry",
            Requirements: TaskRequirements{
                Specialization: []string{"analysis"},
                MinCapability:  []string{"data_analysis", "insights"},
                MaxDuration:    35 * time.Second,
                QualityLevel:   "medium",
            },
            Deadline: time.Now().Add(75 * time.Second),
        },
        {
            ID:   "edit-content",
            Type: "editing",
            Priority: 3,
            Input: "Edit this draft: 'AI is changing the world in many ways. It helps with automation and makes things more efficient. The future looks bright for AI technology.'",
            Requirements: TaskRequirements{
                Specialization: []string{"editing"},
                MinCapability:  []string{"proofreading", "style_checking"},
                MaxDuration:    20 * time.Second,
                QualityLevel:   "high",
            },
            Deadline: time.Now().Add(45 * time.Second),
        },
    }

    // Execute parallel tasks
    ctx := context.Background()
    result, err := orchestra.ExecuteParallelTasks(ctx, tasks)
    if err != nil {
        log.Printf("Parallel execution failed: %v", err)
        return
    }

    // Display results
    fmt.Printf("\n📊 Parallel Execution Results\n")
    fmt.Printf("============================\n")
    fmt.Printf("Total Tasks: %d\n", result.TotalTasks)
    fmt.Printf("Successful: %d\n", result.SuccessfulTasks)
    fmt.Printf("Failed: %d\n", result.FailedTasks)
    fmt.Printf("Timed Out: %t\n", result.TimedOut)
    fmt.Printf("Total Duration: %v\n", result.Duration)

    fmt.Printf("\n📋 Individual Task Results:\n")
    for taskID, taskResult := range result.TaskResults {
        status := "✅"
        if !taskResult.Success {
            status = "❌"
        }
        fmt.Printf("%s %s: %v (Quality: %.2f, Confidence: %.2f)\n",
            status, taskID, taskResult.Duration, taskResult.Quality, taskResult.Confidence)
        
        if taskResult.Output != nil {
            outputStr := fmt.Sprintf("%v", taskResult.Output)
            if len(outputStr) > 100 {
                outputStr = outputStr[:100] + "..."
            }
            fmt.Printf("   Output: %s\n", outputStr)
        }
    }

    // Show aggregated results if available
    if result.AggregatedResult != nil {
        fmt.Printf("\n🔗 Aggregated Result:\n")
        fmt.Printf("Method: %s\n", result.AggregatedResult.AggregationMethod)
        fmt.Printf("Quality Score: %.2f\n", result.AggregatedResult.QualityScore)
        fmt.Printf("Confidence: %.2f\n", result.AggregatedResult.Confidence)
        
        if result.AggregatedResult.CombinedOutput != nil {
            outputStr := fmt.Sprintf("%v", result.AggregatedResult.CombinedOutput)
            if len(outputStr) > 200 {
                outputStr = outputStr[:200] + "..."
            }
            fmt.Printf("Combined Output: %s\n", outputStr)
        }
    }

    // Show worker performance
    fmt.Printf("\n👥 Worker Performance:\n")
    for workerID, metrics := range result.WorkerMetrics {
        fmt.Printf("Worker %s:\n", workerID)
        fmt.Printf("  Tasks Completed: %d\n", metrics.TasksCompleted)
        fmt.Printf("  Success Rate: %.1f%%\n", metrics.SuccessRate*100)
        fmt.Printf("  Average Time: %v\n", metrics.AverageTime)
        fmt.Printf("  Last Active: %v\n", metrics.LastActive.Format("15:04:05"))
        fmt.Println()
    }
}
```

### Advanced Features
✅ **Dynamic Task Assignment** - Intelligent worker selection based on capabilities  
✅ **Real-time Coordination** - Coordinator agent manages resource allocation  
✅ **Result Aggregation** - Combine outputs from multiple agents intelligently  
✅ **Performance Monitoring** - Track worker metrics and success rates  
✅ **Quality Assessment** - Evaluate and score agent outputs  

---

## Level 3: Hierarchical Agent Networks
*Build sophisticated agent hierarchies with delegation and oversight*

### Enterprise Agent Hierarchy
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

// AgentHierarchy represents a hierarchical organization of agents
type AgentHierarchy struct {
    name            string
    rootAgent       *HierarchicalAgent
    agentRegistry   map[string]*HierarchicalAgent
    communicationTree *CommunicationTree
    governanceSystem *GovernanceSystem
    
    mutex           sync.RWMutex
}

type HierarchicalAgent struct {
    ID              string
    Name            string
    Level           int
    Role            AgentRole
    Agent           domain.BaseAgent
    
    // Hierarchy relationships
    Manager         *HierarchicalAgent
    DirectReports   []*HierarchicalAgent
    
    // Capabilities and responsibilities
    Authority       AuthorityLevel
    Responsibilities []string
    Capabilities    []string
    Specialization  string
    
    // State and metrics
    CurrentTasks    []*HierarchicalTask
    CompletedTasks  []*HierarchicalTask
    Performance     PerformanceMetrics
    Status          AgentStatus
    
    mutex           sync.RWMutex
}

type AgentRole string

const (
    RoleExecutive   AgentRole = "executive"    // Top-level strategic decisions
    RoleManager     AgentRole = "manager"      // Team coordination and oversight
    RoleSpecialist  AgentRole = "specialist"   // Domain expertise
    RoleWorker      AgentRole = "worker"       // Task execution
    RoleCoordinator AgentRole = "coordinator"  // Inter-team communication
)

type AuthorityLevel int

const (
    AuthorityLow    AuthorityLevel = 1
    AuthorityMedium AuthorityLevel = 3
    AuthorityHigh   AuthorityLevel = 5
    AuthorityMax    AuthorityLevel = 10
)

type AgentStatus string

const (
    AgentStatusActive    AgentStatus = "active"
    AgentStatusBusy      AgentStatus = "busy"
    AgentStatusWaiting   AgentStatus = "waiting"
    AgentStatusOffline   AgentStatus = "offline"
)

type HierarchicalTask struct {
    ID              string
    Type            TaskType
    Priority        Priority
    Description     string
    Requirements    TaskRequirements
    Subtasks        []*HierarchicalTask
    Dependencies    []string
    
    // Assignment and delegation
    AssignedAgent   string
    DelegatedBy     string
    DelegatedTo     []string
    
    // Execution tracking
    Status          TaskStatus
    Progress        float64
    StartTime       time.Time
    Deadline        time.Time
    CompletionTime  time.Time
    
    // Results and feedback
    Result          *TaskResult
    QualityReview   *QualityReview
    Escalations     []*Escalation
    
    mutex           sync.RWMutex
}

type TaskType string

const (
    TaskTypeStrategic    TaskType = "strategic"
    TaskTypeOperational  TaskType = "operational"
    TaskTypeTactical     TaskType = "tactical"
    TaskTypeResearch     TaskType = "research"
    TaskTypeAnalysis     TaskType = "analysis"
    TaskTypeExecution    TaskType = "execution"
)

type Priority int

const (
    PriorityLow      Priority = 1
    PriorityMedium   Priority = 3
    PriorityHigh     Priority = 5
    PriorityCritical Priority = 10
)

type QualityReview struct {
    ReviewerID      string
    QualityScore    float64
    Feedback        string
    Improvements    []string
    Approved        bool
    ReviewTime      time.Time
}

type Escalation struct {
    Reason          string
    EscalatedBy     string
    EscalatedTo     string
    EscalationTime  time.Time
    Resolution      string
    Resolved        bool
}

// CommunicationTree manages hierarchical communication patterns
type CommunicationTree struct {
    channels        map[string]*CommunicationChannel
    routingRules    []RoutingRule
    messageQueue    chan *HierarchicalMessage
    
    mutex           sync.RWMutex
}

type CommunicationChannel struct {
    ID              string
    Participants    []string
    AccessLevel     AuthorityLevel
    MessageHistory  []*HierarchicalMessage
    IsActive        bool
}

type HierarchicalMessage struct {
    ID              string
    From            string
    To              []string
    Type            MessageType
    Content         interface{}
    Priority        Priority
    RequiresResponse bool
    ResponseDeadline time.Time
    Responses       []*MessageResponse
    Timestamp       time.Time
}

type MessageResponse struct {
    From            string
    Content         interface{}
    Timestamp       time.Time
}

type RoutingRule struct {
    FromLevel       int
    ToLevel         int
    MessageType     MessageType
    RequiresApproval bool
    AutoForward     bool
}

// GovernanceSystem manages policies and compliance
type GovernanceSystem struct {
    policies        []Policy
    approvalChains  map[string][]string
    auditTrail      []*AuditEvent
    complianceRules []ComplianceRule
    
    mutex           sync.RWMutex
}

type Policy struct {
    ID              string
    Name            string
    Description     string
    ApplicableRoles []AgentRole
    Rules           []PolicyRule
    Enforcement     EnforcementLevel
}

type PolicyRule struct {
    Condition       string
    Action          string
    Parameters      map[string]interface{}
}

type EnforcementLevel string

const (
    EnforcementAdvisory  EnforcementLevel = "advisory"
    EnforcementWarning   EnforcementLevel = "warning"
    EnforcementBlocking  EnforcementLevel = "blocking"
)

type AuditEvent struct {
    ID              string
    AgentID         string
    Action          string
    Details         map[string]interface{}
    Timestamp       time.Time
    ComplianceScore float64
}

func NewAgentHierarchy(name string) *AgentHierarchy {
    return &AgentHierarchy{
        name:            name,
        agentRegistry:   make(map[string]*HierarchicalAgent),
        communicationTree: NewCommunicationTree(),
        governanceSystem:  NewGovernanceSystem(),
    }
}

func (ah *AgentHierarchy) CreateRootAgent(name, agentProvider string) error {
    agent, err := core.NewAgentFromString(fmt.Sprintf("%s-root", name), agentProvider)
    if err != nil {
        return fmt.Errorf("failed to create root agent: %w", err)
    }

    rootAgent := &HierarchicalAgent{
        ID:              "root-001",
        Name:            name,
        Level:           0,
        Role:            RoleExecutive,
        Agent:           agent,
        Authority:       AuthorityMax,
        Responsibilities: []string{"strategic_planning", "resource_allocation", "oversight"},
        Capabilities:    []string{"decision_making", "delegation", "evaluation"},
        Specialization:  "executive_leadership",
        DirectReports:   make([]*HierarchicalAgent, 0),
        CurrentTasks:    make([]*HierarchicalTask, 0),
        CompletedTasks:  make([]*HierarchicalTask, 0),
        Status:          AgentStatusActive,
    }

    agent.SetSystemPrompt(`You are the Chief Executive Agent of this hierarchical system.
    Your responsibilities include:
    1. Strategic planning and high-level decision making
    2. Resource allocation and priority setting
    3. Delegating tasks to appropriate subordinates
    4. Monitoring organizational performance
    5. Ensuring compliance with policies and standards
    
    Always consider the broader organizational impact of your decisions.
    Delegate effectively while maintaining oversight and accountability.`)

    ah.rootAgent = rootAgent
    ah.agentRegistry[rootAgent.ID] = rootAgent

    fmt.Printf("👑 Created root agent: %s (Level: %d, Authority: %d)\n", 
        rootAgent.Name, rootAgent.Level, rootAgent.Authority)

    return nil
}

func (ah *AgentHierarchy) AddAgent(name, agentProvider string, level int, role AgentRole, managerID string, specialization string) error {
    agent, err := core.NewAgentFromString(fmt.Sprintf("%s-%s", ah.name, name), agentProvider)
    if err != nil {
        return fmt.Errorf("failed to create agent %s: %w", name, err)
    }

    // Find manager
    manager, exists := ah.agentRegistry[managerID]
    if !exists {
        return fmt.Errorf("manager with ID %s not found", managerID)
    }

    // Create hierarchical agent
    hierarchicalAgent := &HierarchicalAgent{
        ID:              fmt.Sprintf("agent-%s-%d", name, time.Now().Unix()),
        Name:            name,
        Level:           level,
        Role:            role,
        Agent:           agent,
        Manager:         manager,
        Authority:       ah.calculateAuthority(level, role),
        Responsibilities: ah.getDefaultResponsibilities(role),
        Capabilities:    ah.getDefaultCapabilities(role),
        Specialization:  specialization,
        DirectReports:   make([]*HierarchicalAgent, 0),
        CurrentTasks:    make([]*HierarchicalTask, 0),
        CompletedTasks:  make([]*HierarchicalTask, 0),
        Status:          AgentStatusActive,
    }

    // Set role-specific system prompt
    prompt := ah.generateRolePrompt(role, specialization, level)
    agent.SetSystemPrompt(prompt)

    // Add to manager's direct reports
    manager.mutex.Lock()
    manager.DirectReports = append(manager.DirectReports, hierarchicalAgent)
    manager.mutex.Unlock()

    // Register agent
    ah.agentRegistry[hierarchicalAgent.ID] = hierarchicalAgent

    fmt.Printf("👥 Added agent: %s (Level: %d, Role: %s, Manager: %s)\n", 
        name, level, role, manager.Name)

    return nil
}

func (ah *AgentHierarchy) calculateAuthority(level int, role AgentRole) AuthorityLevel {
    baseAuthority := AuthorityMax - AuthorityLevel(level)
    
    switch role {
    case RoleExecutive:
        return baseAuthority
    case RoleManager:
        return baseAuthority - 1
    case RoleCoordinator:
        return baseAuthority - 1
    case RoleSpecialist:
        return baseAuthority - 2
    case RoleWorker:
        return baseAuthority - 3
    default:
        return AuthorityLow
    }
}

func (ah *AgentHierarchy) getDefaultResponsibilities(role AgentRole) []string {
    switch role {
    case RoleExecutive:
        return []string{"strategic_planning", "resource_allocation", "performance_oversight"}
    case RoleManager:
        return []string{"team_coordination", "task_delegation", "performance_monitoring"}
    case RoleCoordinator:
        return []string{"communication_facilitation", "resource_coordination", "status_reporting"}
    case RoleSpecialist:
        return []string{"domain_expertise", "technical_guidance", "quality_assurance"}
    case RoleWorker:
        return []string{"task_execution", "deliverable_creation", "status_reporting"}
    default:
        return []string{"general_support"}
    }
}

func (ah *AgentHierarchy) getDefaultCapabilities(role AgentRole) []string {
    switch role {
    case RoleExecutive:
        return []string{"strategic_thinking", "decision_making", "leadership"}
    case RoleManager:
        return []string{"project_management", "team_leadership", "coordination"}
    case RoleCoordinator:
        return []string{"communication", "organization", "facilitation"}
    case RoleSpecialist:
        return []string{"technical_expertise", "analysis", "consultation"}
    case RoleWorker:
        return []string{"execution", "implementation", "delivery"}
    default:
        return []string{"general_capabilities"}
    }
}

func (ah *AgentHierarchy) generateRolePrompt(role AgentRole, specialization string, level int) string {
    basePrompt := fmt.Sprintf("You are a %s agent at level %d in a hierarchical organization. Your specialization is %s.\n\n", role, level, specialization)
    
    switch role {
    case RoleManager:
        return basePrompt + `Your key responsibilities:
1. Coordinate your team and manage resources effectively
2. Delegate tasks to appropriate subordinates
3. Monitor progress and provide guidance
4. Escalate issues when necessary
5. Report status to your manager
6. Ensure quality standards are met

Always balance efficiency with quality and maintain clear communication.`

    case RoleSpecialist:
        return basePrompt + `Your key responsibilities:
1. Provide expert knowledge in your specialization
2. Ensure technical accuracy and quality
3. Guide and mentor junior team members
4. Stay current with best practices
5. Recommend improvements and innovations
6. Collaborate with other specialists

Focus on excellence in your domain while supporting team objectives.`

    case RoleWorker:
        return basePrompt + `Your key responsibilities:
1. Execute assigned tasks efficiently and accurately
2. Follow established procedures and standards
3. Report progress and issues promptly
4. Seek guidance when needed
5. Collaborate effectively with team members
6. Deliver high-quality work on time

Focus on reliable execution while learning and improving.`

    case RoleCoordinator:
        return basePrompt + `Your key responsibilities:
1. Facilitate communication between teams and levels
2. Coordinate resources and schedules
3. Track progress across multiple workstreams
4. Identify and resolve conflicts
5. Maintain documentation and status reports
6. Support organizational efficiency

Focus on smooth operations and clear communication.`

    default:
        return basePrompt + "Perform your assigned role effectively while supporting organizational objectives."
    }
}

func (ah *AgentHierarchy) ExecuteHierarchicalTask(ctx context.Context, task *HierarchicalTask) (*HierarchicalTaskResult, error) {
    fmt.Printf("🎯 Starting hierarchical task: %s (Type: %s, Priority: %d)\n", 
        task.Description, task.Type, task.Priority)

    result := &HierarchicalTaskResult{
        TaskID:          task.ID,
        StartTime:       time.Now(),
        ExecutionTrace:  make([]*ExecutionStep, 0),
        AgentContributions: make(map[string]*AgentContribution),
    }

    // Start from root agent for delegation
    executionStep, err := ah.delegateTask(ctx, ah.rootAgent, task)
    if err != nil {
        result.Success = false
        result.Error = err
        return result, err
    }

    result.ExecutionTrace = append(result.ExecutionTrace, executionStep)

    // Monitor execution
    err = ah.monitorTaskExecution(ctx, task, result)
    if err != nil {
        result.Success = false
        result.Error = err
        return result, err
    }

    result.Duration = time.Since(result.StartTime)
    result.Success = task.Status == TaskStatusCompleted

    fmt.Printf("✅ Hierarchical task completed: %s (Duration: %v, Success: %t)\n", 
        task.Description, result.Duration, result.Success)

    return result, nil
}

func (ah *AgentHierarchy) delegateTask(ctx context.Context, manager *HierarchicalAgent, task *HierarchicalTask) (*ExecutionStep, error) {
    executionStep := &ExecutionStep{
        AgentID:     manager.ID,
        Action:      "delegation",
        StartTime:   time.Now(),
        Description: fmt.Sprintf("Delegating task to subordinates"),
    }

    // Manager analyzes task and decides on delegation
    delegationState := domain.NewState()
    delegationState.Set("task_description", task.Description)
    delegationState.Set("task_type", task.Type)
    delegationState.Set("task_priority", task.Priority)
    delegationState.Set("available_reports", ah.getDirectReportInfo(manager))
    delegationState.Set("manager_authority", manager.Authority)

    delegationPrompt := fmt.Sprintf(`Analyze this task and determine the best delegation strategy:

Task: %s
Type: %s
Priority: %d
Deadline: %v

Available direct reports and their capabilities:
%s

Decide:
1. Should you handle this personally or delegate?
2. If delegating, which subordinate(s) should receive the task?
3. Should the task be broken into subtasks?
4. What oversight level is needed?

Provide your delegation decision and reasoning.`, 
        task.Description, task.Type, task.Priority, task.Deadline,
        ah.formatDirectReportsInfo(manager))

    delegationState.Set("user_input", delegationPrompt)

    result, err := manager.Agent.Run(ctx, delegationState)
    if err != nil {
        executionStep.Success = false
        executionStep.Error = err
        return executionStep, err
    }

    // Parse delegation decision (simplified)
    delegationDecision := result.Get("response")
    executionStep.Output = delegationDecision
    executionStep.Duration = time.Since(executionStep.StartTime)
    executionStep.Success = true

    // For demonstration, delegate to first available subordinate
    if len(manager.DirectReports) > 0 {
        selectedSubordinate := manager.DirectReports[0]
        task.AssignedAgent = selectedSubordinate.ID
        task.DelegatedBy = manager.ID
        task.Status = TaskStatusAssigned

        fmt.Printf("📋 Task delegated by %s to %s\n", manager.Name, selectedSubordinate.Name)

        // If subordinate has their own reports, they may further delegate
        if len(selectedSubordinate.DirectReports) > 0 && task.Type == TaskTypeStrategic {
            _, err := ah.delegateTask(ctx, selectedSubordinate, task)
            if err != nil {
                log.Printf("Sub-delegation failed: %v", err)
            }
        } else {
            // Execute task at this level
            return ah.executeTaskDirectly(ctx, selectedSubordinate, task)
        }
    } else {
        // No subordinates, manager must execute directly
        return ah.executeTaskDirectly(ctx, manager, task)
    }

    return executionStep, nil
}

func (ah *AgentHierarchy) executeTaskDirectly(ctx context.Context, agent *HierarchicalAgent, task *HierarchicalTask) (*ExecutionStep, error) {
    executionStep := &ExecutionStep{
        AgentID:     agent.ID,
        Action:      "execution",
        StartTime:   time.Now(),
        Description: fmt.Sprintf("Direct task execution"),
    }

    // Agent executes the task
    executionState := domain.NewState()
    executionState.Set("task_description", task.Description)
    executionState.Set("task_type", task.Type)
    executionState.Set("task_requirements", task.Requirements)
    executionState.Set("agent_role", agent.Role)
    executionState.Set("agent_specialization", agent.Specialization)
    executionState.Set("agent_capabilities", agent.Capabilities)

    taskPrompt := fmt.Sprintf(`Execute this task according to your role and capabilities:

Task: %s
Type: %s
Your Role: %s
Your Specialization: %s

Requirements:
- Deliver high-quality results
- Follow organizational standards
- Document your approach
- Report any issues or concerns

Provide your complete task execution and results.`, 
        task.Description, task.Type, agent.Role, agent.Specialization)

    executionState.Set("user_input", taskPrompt)

    result, err := agent.Agent.Run(ctx, executionState)
    if err != nil {
        executionStep.Success = false
        executionStep.Error = err
        task.Status = TaskStatusFailed
        return executionStep, err
    }

    // Create task result
    taskResult := &TaskResult{
        Success:  true,
        Output:   result.Get("response"),
        Duration: time.Since(executionStep.StartTime),
        Quality:  ah.assessTaskQuality(result),
        Metadata: map[string]interface{}{
            "executed_by": agent.ID,
            "agent_role":  agent.Role,
        },
    }

    task.Result = taskResult
    task.Status = TaskStatusCompleted
    task.CompletionTime = time.Now()

    executionStep.Output = taskResult.Output
    executionStep.Duration = time.Since(executionStep.StartTime)
    executionStep.Success = true

    // Update agent metrics
    agent.mutex.Lock()
    agent.CompletedTasks = append(agent.CompletedTasks, task)
    agent.Performance.TasksCompleted++
    agent.mutex.Unlock()

    fmt.Printf("✅ Task executed by %s (%s)\n", agent.Name, agent.Role)

    return executionStep, nil
}

func (ah *AgentHierarchy) monitorTaskExecution(ctx context.Context, task *HierarchicalTask, result *HierarchicalTaskResult) error {
    // Simplified monitoring - in production this would be more sophisticated
    timeout := time.After(60 * time.Second)
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-timeout:
            if task.Status != TaskStatusCompleted {
                task.Status = TaskStatusFailed
                return fmt.Errorf("task execution timed out")
            }
            return nil

        case <-ticker.C:
            if task.Status == TaskStatusCompleted || task.Status == TaskStatusFailed {
                return nil
            }
            
            // Update progress (simplified)
            task.Progress = min(task.Progress+0.1, 0.9)

        case <-ctx.Done():
            return ctx.Err()
        }
    }
}

func (ah *AgentHierarchy) assessTaskQuality(result *domain.State) float64 {
    // Simplified quality assessment
    if response, exists := result.Get("response"); exists {
        responseStr := fmt.Sprintf("%v", response)
        switch {
        case len(responseStr) > 200:
            return 0.9
        case len(responseStr) > 100:
            return 0.7
        case len(responseStr) > 50:
            return 0.5
        default:
            return 0.3
        }
    }
    return 0.0
}

// Supporting types and helper functions
type HierarchicalTaskResult struct {
    TaskID             string
    Success            bool
    StartTime          time.Time
    Duration           time.Duration
    ExecutionTrace     []*ExecutionStep
    AgentContributions map[string]*AgentContribution
    Error              error
}

type ExecutionStep struct {
    AgentID     string
    Action      string
    StartTime   time.Time
    Duration    time.Duration
    Description string
    Output      interface{}
    Success     bool
    Error       error
}

type AgentContribution struct {
    AgentID      string
    Role         AgentRole
    TasksHandled int
    QualityScore float64
    Efficiency   float64
}

type PerformanceMetrics struct {
    TasksCompleted   int64
    AverageQuality   float64
    AverageTime      time.Duration
    SuccessRate      float64
    DelegationCount  int64
}

type ComplianceRule struct {
    ID          string
    Description string
    Condition   string
    Action      string
}

func (ah *AgentHierarchy) getDirectReportInfo(manager *HierarchicalAgent) []map[string]interface{} {
    reports := make([]map[string]interface{}, 0)
    for _, report := range manager.DirectReports {
        reports = append(reports, map[string]interface{}{
            "id":            report.ID,
            "name":          report.Name,
            "role":          report.Role,
            "specialization": report.Specialization,
            "capabilities":  report.Capabilities,
            "authority":     report.Authority,
            "status":        report.Status,
        })
    }
    return reports
}

func (ah *AgentHierarchy) formatDirectReportsInfo(manager *HierarchicalAgent) string {
    if len(manager.DirectReports) == 0 {
        return "No direct reports available"
    }

    info := ""
    for i, report := range manager.DirectReports {
        info += fmt.Sprintf("%d. %s (%s) - %s specialist (Authority: %d, Status: %s)\n",
            i+1, report.Name, report.Role, report.Specialization, report.Authority, report.Status)
    }
    return info
}

func min(a, b float64) float64 {
    if a < b { return a }
    return b
}

func NewCommunicationTree() *CommunicationTree {
    return &CommunicationTree{
        channels:     make(map[string]*CommunicationChannel),
        routingRules: make([]RoutingRule, 0),
        messageQueue: make(chan *HierarchicalMessage, 100),
    }
}

func NewGovernanceSystem() *GovernanceSystem {
    return &GovernanceSystem{
        policies:        make([]Policy, 0),
        approvalChains:  make(map[string][]string),
        auditTrail:      make([]*AuditEvent, 0),
        complianceRules: make([]ComplianceRule, 0),
    }
}

func main() {
    fmt.Println("🏢 Hierarchical Agent Networks")
    fmt.Println("==============================")

    // Create agent hierarchy
    hierarchy := NewAgentHierarchy("enterprise-ai")

    // Create root executive agent
    err := hierarchy.CreateRootAgent("CEO", "anthropic/claude-3-5-sonnet")
    if err != nil {
        log.Fatalf("Failed to create root agent: %v", err)
    }

    // Add management layer
    err = hierarchy.AddAgent("CTO", "openai/gpt-4o", 1, RoleManager, "root-001", "technology")
    if err != nil {
        log.Printf("Failed to add CTO: %v", err)
    }

    err = hierarchy.AddAgent("VP-Marketing", "anthropic/claude-3-5-haiku", 1, RoleManager, "root-001", "marketing")
    if err != nil {
        log.Printf("Failed to add VP-Marketing: %v", err)
    }

    // Add specialist layer under CTO
    ctoID := ""
    for id, agent := range hierarchy.agentRegistry {
        if agent.Name == "CTO" {
            ctoID = id
            break
        }
    }

    if ctoID != "" {
        err = hierarchy.AddAgent("Lead-Developer", "openai/gpt-4o-mini", 2, RoleSpecialist, ctoID, "software_development")
        if err != nil {
            log.Printf("Failed to add Lead-Developer: %v", err)
        }

        err = hierarchy.AddAgent("DevOps-Engineer", "gemini/gemini-2.0-flash", 2, RoleSpecialist, ctoID, "infrastructure")
        if err != nil {
            log.Printf("Failed to add DevOps-Engineer: %v", err)
        }
    }

    // Add worker layer
    leadDevID := ""
    for id, agent := range hierarchy.agentRegistry {
        if agent.Name == "Lead-Developer" {
            leadDevID = id
            break
        }
    }

    if leadDevID != "" {
        err = hierarchy.AddAgent("Junior-Developer", "gemini/gemini-2.0-flash", 3, RoleWorker, leadDevID, "frontend_development")
        if err != nil {
            log.Printf("Failed to add Junior-Developer: %v", err)
        }
    }

    // Display hierarchy structure
    fmt.Printf("\n🏗️ Organization Structure:\n")
    ah.displayHierarchy(hierarchy.rootAgent, 0)

    // Execute hierarchical tasks
    tasks := []*HierarchicalTask{
        {
            ID:          "strategic-001",
            Type:        TaskTypeStrategic,
            Priority:    PriorityHigh,
            Description: "Develop AI strategy for Q2 2025",
            Requirements: TaskRequirements{
                Specialization: []string{"technology", "strategy"},
                MaxDuration:    60 * time.Second,
                QualityLevel:   "high",
            },
            Deadline: time.Now().Add(120 * time.Second),
            Status:   TaskStatusPending,
        },
        {
            ID:          "operational-001", 
            Type:        TaskTypeOperational,
            Priority:    PriorityMedium,
            Description: "Implement new deployment pipeline",
            Requirements: TaskRequirements{
                Specialization: []string{"infrastructure", "software_development"},
                MaxDuration:    45 * time.Second,
                QualityLevel:   "medium",
            },
            Deadline: time.Now().Add(90 * time.Second),
            Status:   TaskStatusPending,
        },
        {
            ID:          "tactical-001",
            Type:        TaskTypeTactical,
            Priority:    PriorityLow,
            Description: "Update user interface components",
            Requirements: TaskRequirements{
                Specialization: []string{"frontend_development"},
                MaxDuration:    30 * time.Second,
                QualityLevel:   "medium",
            },
            Deadline: time.Now().Add(60 * time.Second),
            Status:   TaskStatusPending,
        },
    }

    ctx := context.Background()

    for i, task := range tasks {
        fmt.Printf("\n🎯 Executing Hierarchical Task %d\n", i+1)
        fmt.Printf("Task: %s\n", task.Description)
        fmt.Printf("Type: %s, Priority: %d\n", task.Type, task.Priority)

        result, err := hierarchy.ExecuteHierarchicalTask(ctx, task)
        if err != nil {
            fmt.Printf("❌ Task execution failed: %v\n", err)
            continue
        }

        fmt.Printf("📊 Task Result:\n")
        fmt.Printf("  Success: %t\n", result.Success)
        fmt.Printf("  Duration: %v\n", result.Duration)
        fmt.Printf("  Execution Steps: %d\n", len(result.ExecutionTrace))

        for j, step := range result.ExecutionTrace {
            status := "✅"
            if !step.Success {
                status = "❌"
            }
            agent := hierarchy.agentRegistry[step.AgentID]
            agentName := "Unknown"
            if agent != nil {
                agentName = agent.Name
            }
            fmt.Printf("    %s Step %d: %s by %s (%v)\n", 
                status, j+1, step.Action, agentName, step.Duration)
        }

        if task.Result != nil && task.Result.Output != nil {
            outputStr := fmt.Sprintf("%v", task.Result.Output)
            if len(outputStr) > 200 {
                outputStr = outputStr[:200] + "..."
            }
            fmt.Printf("  Output: %s\n", outputStr)
        }
    }

    // Display final organization metrics
    fmt.Printf("\n📈 Organization Performance:\n")
    for id, agent := range hierarchy.agentRegistry {
        fmt.Printf("Agent %s (%s):\n", agent.Name, agent.Role)
        fmt.Printf("  Completed Tasks: %d\n", len(agent.CompletedTasks))
        fmt.Printf("  Current Tasks: %d\n", len(agent.CurrentTasks))
        fmt.Printf("  Direct Reports: %d\n", len(agent.DirectReports))
        fmt.Printf("  Authority Level: %d\n", agent.Authority)
        fmt.Printf("  Status: %s\n", agent.Status)
        fmt.Println()
    }
}

func (ah *AgentHierarchy) displayHierarchy(agent *HierarchicalAgent, level int) {
    indent := ""
    for i := 0; i < level; i++ {
        indent += "  "
    }
    
    fmt.Printf("%s├─ %s (%s) - Level %d\n", indent, agent.Name, agent.Role, agent.Level)
    
    for _, report := range agent.DirectReports {
        ah.displayHierarchy(report, level+1)
    }
}
```

### Hierarchical Features
✅ **Multi-Level Delegation** - Intelligent task delegation down the hierarchy  
✅ **Role-Based Authority** - Authority levels determine decision-making power  
✅ **Governance System** - Policies and compliance monitoring  
✅ **Communication Tree** - Structured inter-agent messaging  
✅ **Performance Tracking** - Monitor agent performance across the hierarchy  
✅ **Quality Review** - Hierarchical quality assurance processes  

---

## Best Practices for Agent Communication

### Communication Design Patterns

1. **Clear Interfaces**
   - Define explicit input/output contracts
   - Use structured data formats
   - Validate data at boundaries

2. **Error Handling**
   - Implement graceful degradation
   - Provide meaningful error messages
   - Include retry mechanisms

3. **State Management**
   - Use immutable state where possible
   - Track state changes explicitly
   - Implement state validation

4. **Performance Optimization**
   - Cache frequently used results
   - Use async communication when possible
   - Monitor and optimize bottlenecks

## Next Steps

🤝 **Agent communication mastered!** Continue with:

- **[Agent Memory](agent-memory.md)** - State management patterns for agents
- **[Structured Data](structured-data.md)** - Data handling and validation
- **[Data Validation](data-validation.md)** - Ensure data quality in communication
- **[Workflow Orchestration](../advanced/workflow-orchestration.md)** - Advanced coordination patterns

### Quick Reference

- **[Agent Architecture](../technical/agents/overview.md)** - Technical agent details
- **[Communication Patterns](../technical/agents/communication.md)** - Advanced patterns
- **[Best Practices Checklist](../reference/best-practices-checklist.md)** - Communication checklist

---

**Need help with complex coordination?** Check our [multi-agent examples](../examples/multi-agent-projects.md) or join the discussion on [GitHub](https://github.com/lexlapax/go-llms/discussions).