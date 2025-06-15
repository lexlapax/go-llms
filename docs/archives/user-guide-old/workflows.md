# Building Workflows

This guide shows you how to create complex multi-agent workflows using go-llms workflow patterns.

## Overview

Workflows orchestrate multiple agents to accomplish complex tasks. Go-llms provides four core workflow patterns:
- **Sequential**: Execute agents one after another
- **Parallel**: Run multiple agents simultaneously  
- **Conditional**: Branch based on conditions
- **Loop**: Iterate until a condition is met

## Sequential Workflows

Execute agents in order, passing state from one to the next.

### Basic Sequential Flow

```go
import "github.com/lexlapax/go-llms/pkg/agent/workflow"

// Create a data processing pipeline
pipeline := workflow.NewSequentialAgent("data-pipeline")

// Add agents in order
pipeline.AddAgent(dataLoader)      // Step 1: Load data
pipeline.AddAgent(dataValidator)   // Step 2: Validate
pipeline.AddAgent(dataProcessor)   // Step 3: Process
pipeline.AddAgent(reportGenerator) // Step 4: Generate report

// Run the pipeline
result, err := pipeline.Run(context.Background(), initialState)
```

### With Error Handling

```go
// Stop on first error
pipeline.SetStopOnError(true)

// Or continue and collect errors
pipeline.SetStopOnError(false)
results, err := pipeline.Run(ctx, state)
if err != nil {
    // Check which steps failed
    for i, agent := range pipeline.GetAgents() {
        if stepErr := results.GetError(agent.Name()); stepErr != nil {
            fmt.Printf("Step %d (%s) failed: %v\n", i+1, agent.Name(), stepErr)
        }
    }
}
```

### State Flow Control

```go
// Each agent can modify state for the next
extractAgent := createAgent("extract", func(state *domain.State) (*domain.State, error) {
    // Extract data
    data := extractData(state.Get("source"))
    state.Set("extracted_data", data)
    return state, nil
})

transformAgent := createAgent("transform", func(state *domain.State) (*domain.State, error) {
    // Use data from previous step
    data := state.Get("extracted_data")
    transformed := transformData(data)
    state.Set("transformed_data", transformed)
    return state, nil
})
```

## Parallel Workflows

Run multiple agents simultaneously and merge results.

### Basic Parallel Execution

```go
// Create parallel analyzer
analyzer := workflow.NewParallelAgent("multi-analyzer")

// Add parallel tasks
analyzer.AddAgent(sentimentAnalyzer)
analyzer.AddAgent(keywordExtractor)
analyzer.AddAgent(summaryGenerator)
analyzer.AddAgent(topicClassifier)

// All run simultaneously
result, err := analyzer.Run(ctx, textState)

// Results are automatically merged
sentiment := result.Get("sentiment")
keywords := result.Get("keywords")
summary := result.Get("summary")
topics := result.Get("topics")
```

### With Timeout Control

```go
// Set timeout for parallel execution
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := analyzer.Run(ctx, state)
if err == context.DeadlineExceeded {
    // Some agents didn't finish in time
}
```

### Custom Merge Strategies

```go
// Define how to merge results
analyzer.SetMergeStrategy(func(results []*domain.State) *domain.State {
    merged := domain.NewState()
    
    // Custom merge logic
    var allScores []float64
    for _, result := range results {
        if score, ok := result.Get("score").(float64); ok {
            allScores = append(allScores, score)
        }
    }
    
    // Calculate average
    avg := average(allScores)
    merged.Set("average_score", avg)
    
    return merged
})
```

### Handling Partial Failures

```go
// Configure failure tolerance
analyzer.SetRequireAllSuccess(false) // Don't fail if some agents fail

result, err := analyzer.Run(ctx, state)

// Check individual results
successCount := result.GetInt("successful_agents")
failedAgents := result.Get("failed_agents").([]string)
```

## Conditional Workflows

Route execution based on conditions.

### Basic Branching

```go
// Create conditional router
router := workflow.NewConditionalAgent("task-router")

// Add branches with conditions
router.AddBranch(
    "urgent",
    func(state *domain.State) bool {
        priority := state.GetString("priority")
        return priority == "high" || priority == "urgent"
    },
    urgentHandler,
)

router.AddBranch(
    "scheduled",
    func(state *domain.State) bool {
        return state.Has("scheduled_time")
    },
    scheduledHandler,
)

// Default branch (always true)
router.AddBranch(
    "normal",
    func(state *domain.State) bool { return true },
    normalHandler,
)

// Routes to appropriate handler
result, err := router.Run(ctx, taskState)
```

### Complex Conditions

```go
// Multi-criteria routing
router.AddBranch(
    "premium_urgent",
    func(state *domain.State) bool {
        isPremium := state.GetBool("is_premium_user")
        isUrgent := state.GetString("priority") == "urgent"
        hasPayment := state.Has("payment_method")
        
        return isPremium && isUrgent && hasPayment
    },
    premiumUrgentHandler,
)

// Data-based routing
router.AddBranch(
    "large_dataset",
    func(state *domain.State) bool {
        if data, ok := state.Get("data").([]interface{}); ok {
            return len(data) > 10000
        }
        return false
    },
    largeDataHandler,
)
```

### Dynamic Branch Selection

```go
// Select branch based on computed value
router.AddBranch(
    "ml_route",
    func(state *domain.State) bool {
        // Use ML model to decide
        features := extractFeatures(state)
        prediction := mlModel.Predict(features)
        return prediction == "route_a"
    },
    mlRouteHandler,
)
```

## Loop Workflows

Iterate until a condition is met.

### Basic Loop

```go
// Create refinement loop
refiner := workflow.NewLoopAgent("content-refiner")

// Set the agent that runs in each iteration
refiner.SetLoopAgent(refinementAgent)

// Set termination condition
refiner.SetCondition(func(state *domain.State) bool {
    quality := state.GetFloat64("quality_score")
    return quality >= 0.9 // Stop when quality is good enough
})

// Set maximum iterations (safety limit)
refiner.SetMaxIterations(5)

// Run the loop
result, err := refiner.Run(ctx, draftState)
```

### Progressive Refinement

```go
// Agent that improves with each iteration
refinementAgent := createAgent("refine", func(state *domain.State) (*domain.State, error) {
    iteration := state.GetInt("iteration")
    content := state.GetString("content")
    
    // Apply different strategies based on iteration
    var refined string
    switch iteration {
    case 1:
        refined = improveGrammar(content)
    case 2:
        refined = enhanceClarity(refined)
    case 3:
        refined = addDetails(refined)
    default:
        refined = polishFinal(refined)
    }
    
    // Evaluate quality
    quality := evaluateQuality(refined)
    
    state.Set("content", refined)
    state.Set("quality_score", quality)
    state.Set("iteration", iteration + 1)
    
    return state, nil
})
```

### Early Termination

```go
// Multiple termination conditions
refiner.SetCondition(func(state *domain.State) bool {
    quality := state.GetFloat64("quality_score")
    iteration := state.GetInt("iteration")
    userSatisfied := state.GetBool("user_approved")
    
    // Stop if any condition is met
    return quality >= 0.9 || 
           iteration >= 10 || 
           userSatisfied
})
```

## Complex Workflow Patterns

### Nested Workflows

Combine workflow patterns for complex orchestration:

```go
// Main sequential pipeline
mainPipeline := workflow.NewSequentialAgent("main-pipeline")

// Step 1: Parallel data gathering
dataGatherer := workflow.NewParallelAgent("data-gatherer")
dataGatherer.AddAgent(apiDataFetcher)
dataGatherer.AddAgent(databaseReader)
dataGatherer.AddAgent(fileSystemScanner)

// Step 2: Conditional processing
processor := workflow.NewConditionalAgent("processor")
processor.AddBranch("small", isSmallDataset, smallDataProcessor)
processor.AddBranch("large", isLargeDataset, largeDataProcessor)

// Step 3: Iterative refinement
refiner := workflow.NewLoopAgent("refiner")
refiner.SetLoopAgent(qualityImprover)
refiner.SetCondition(meetsQualityStandard)

// Combine into main pipeline
mainPipeline.AddAgent(dataGatherer)
mainPipeline.AddAgent(processor)
mainPipeline.AddAgent(refiner)

// Run the complete workflow
result, err := mainPipeline.Run(ctx, initialState)
```

### Fork-Join Pattern

```go
// Split work, process in parallel, then merge
forkJoin := createForkJoinWorkflow(func(state *domain.State) []*domain.State {
    // Fork: Split data into chunks
    data := state.Get("large_dataset").([]interface{})
    chunks := splitIntoChunks(data, 4)
    
    states := make([]*domain.State, len(chunks))
    for i, chunk := range chunks {
        s := domain.NewState()
        s.Set("chunk", chunk)
        s.Set("chunk_id", i)
        states[i] = s
    }
    return states
})

// Process chunks in parallel
parallel := workflow.NewParallelAgent("chunk-processor")
for i := 0; i < 4; i++ {
    parallel.AddAgent(createChunkProcessor(i))
}

// Join: Merge results
joiner := createAgent("join", func(states []*domain.State) (*domain.State, error) {
    merged := domain.NewState()
    
    var allResults []interface{}
    for _, state := range states {
        results := state.Get("results").([]interface{})
        allResults = append(allResults, results...)
    }
    
    merged.Set("final_results", allResults)
    return merged, nil
})
```

### Map-Reduce Pattern

```go
// Map phase: Process items independently
mapper := workflow.NewParallelAgent("mapper")
for _, item := range items {
    mapper.AddAgent(createMapAgent(item))
}

// Reduce phase: Aggregate results
reducer := createAgent("reducer", func(state *domain.State) (*domain.State, error) {
    mappedResults := state.Get("mapped_results").([]interface{})
    
    // Aggregate
    final := reduce(mappedResults)
    
    state.Set("final_result", final)
    return state, nil
})

// Combine in sequence
mapReduce := workflow.NewSequentialAgent("map-reduce")
mapReduce.AddAgent(mapper)
mapReduce.AddAgent(reducer)
```

## Real-World Examples

### Document Processing Pipeline

```go
// Complete document processing workflow
docPipeline := workflow.NewSequentialAgent("document-processor")

// 1. Extract text from various formats
extractor := workflow.NewConditionalAgent("extractor")
extractor.AddBranch("pdf", isPDF, pdfExtractor)
extractor.AddBranch("docx", isDOCX, docxExtractor)
extractor.AddBranch("html", isHTML, htmlExtractor)

// 2. Parallel analysis
analyzer := workflow.NewParallelAgent("analyzer")
analyzer.AddAgent(languageDetector)
analyzer.AddAgent(sentimentAnalyzer)
analyzer.AddAgent(entityExtractor)
analyzer.AddAgent(summaryGenerator)

// 3. Quality check loop
qualityChecker := workflow.NewLoopAgent("quality-checker")
qualityChecker.SetLoopAgent(qualityImprover)
qualityChecker.SetCondition(func(s *domain.State) bool {
    return s.GetFloat64("quality") >= 0.8
})

// 4. Output generation
outputGen := workflow.NewConditionalAgent("output-generator")
outputGen.AddBranch("report", wantsReport, reportGenerator)
outputGen.AddBranch("api", wantsAPI, apiFormatter)
outputGen.AddBranch("email", wantsEmail, emailFormatter)

// Assemble pipeline
docPipeline.AddAgent(extractor)
docPipeline.AddAgent(analyzer)
docPipeline.AddAgent(qualityChecker)
docPipeline.AddAgent(outputGen)
```

### Customer Support Workflow

```go
// Intelligent support ticket router
supportFlow := workflow.NewSequentialAgent("support-flow")

// 1. Categorize ticket
categorizer := createLLMAgent("categorizer", "Categorize support tickets")

// 2. Route based on category
router := workflow.NewConditionalAgent("router")

// Technical issues
router.AddBranch("technical", 
    func(s *domain.State) bool {
        return s.GetString("category") == "technical"
    },
    createTechnicalFlow(),
)

// Billing issues  
router.AddBranch("billing",
    func(s *domain.State) bool {
        return s.GetString("category") == "billing"
    },
    createBillingFlow(),
)

// General inquiries
router.AddBranch("general",
    func(s *domain.State) bool { return true },
    createGeneralFlow(),
)

// 3. Follow-up loop
followUp := workflow.NewLoopAgent("follow-up")
followUp.SetLoopAgent(customerSatisfactionChecker)
followUp.SetCondition(func(s *domain.State) bool {
    return s.GetBool("issue_resolved") || s.GetInt("attempts") > 3
})

supportFlow.AddAgent(categorizer)
supportFlow.AddAgent(router)
supportFlow.AddAgent(followUp)
```

### Research Workflow

```go
// Multi-phase research workflow
researchFlow := workflow.NewSequentialAgent("research")

// Phase 1: Parallel information gathering
infoGathering := workflow.NewParallelAgent("gather-info")
infoGathering.AddAgent(webSearcher)
infoGathering.AddAgent(academicSearcher)
infoGathering.AddAgent(newsSearcher)
infoGathering.AddAgent(socialMediaSearcher)

// Phase 2: Source validation loop
validator := workflow.NewLoopAgent("validate-sources")
validator.SetLoopAgent(sourceValidator)
validator.SetCondition(func(s *domain.State) bool {
    validSources := s.GetInt("valid_source_count")
    return validSources >= 10
})

// Phase 3: Synthesis
synthesizer := createLLMAgent("synthesizer", "Synthesize research findings")

// Phase 4: Fact checking
factChecker := workflow.NewParallelAgent("fact-check")
factChecker.AddAgent(claimExtractor)
factChecker.AddAgent(evidenceFinder)
factChecker.AddAgent(contradictionDetector)

// Phase 5: Report generation
reporter := createLLMAgent("reporter", "Generate research report")

researchFlow.AddAgent(infoGathering)
researchFlow.AddAgent(validator)
researchFlow.AddAgent(synthesizer)
researchFlow.AddAgent(factChecker)
researchFlow.AddAgent(reporter)
```

## State Management in Workflows

### State Propagation

```go
// State flows through the workflow
initialState := domain.NewState()
initialState.Set("input_data", data)
initialState.Set("config", config)
initialState.Set("user_preferences", prefs)

// Each agent can read and modify state
result, err := workflow.Run(ctx, initialState)

// Final state contains accumulated results
finalData := result.Get("processed_data")
metrics := result.Get("processing_metrics")
logs := result.Get("processing_logs")
```

### State Isolation

```go
// Agents can work with isolated state copies
parallelAgent := workflow.NewParallelAgent("isolated-parallel")
parallelAgent.SetStateIsolation(true) // Each agent gets a copy

// Or share state (careful with concurrent access)
parallelAgent.SetStateIsolation(false) // Agents share state reference
```

### State Checkpointing

```go
// Save state between steps for recovery
type CheckpointedWorkflow struct {
    *workflow.SequentialAgent
    checkpointer StateCheckpointer
}

func (c *CheckpointedWorkflow) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
    for i, agent := range c.GetAgents() {
        // Run agent
        result, err := agent.Run(ctx, state)
        if err != nil {
            return nil, err
        }
        
        // Checkpoint after each step
        if err := c.checkpointer.Save(i, result); err != nil {
            log.Printf("Checkpoint failed: %v", err)
        }
        
        state = result
    }
    
    return state, nil
}
```

## Error Handling and Recovery

### Retry Logic

```go
// Add retry wrapper to any workflow
retryWorkflow := createRetryWrapper(workflow, RetryConfig{
    MaxAttempts: 3,
    Backoff:     time.Second,
    RetryOn: func(err error) bool {
        // Retry on transient errors
        return isTransientError(err)
    },
})
```

### Fallback Workflows

```go
// Primary workflow with fallback
primaryWorkflow := createPrimaryWorkflow()
fallbackWorkflow := createFallbackWorkflow()

result, err := primaryWorkflow.Run(ctx, state)
if err != nil {
    log.Printf("Primary workflow failed: %v, trying fallback", err)
    result, err = fallbackWorkflow.Run(ctx, state)
}
```

### Partial Recovery

```go
// Continue workflow from last successful step
type RecoverableWorkflow struct {
    *workflow.SequentialAgent
    lastSuccessful int
}

func (r *RecoverableWorkflow) RunWithRecovery(ctx context.Context, state *domain.State) (*domain.State, error) {
    agents := r.GetAgents()
    
    // Start from last successful step
    for i := r.lastSuccessful; i < len(agents); i++ {
        result, err := agents[i].Run(ctx, state)
        if err != nil {
            r.lastSuccessful = i
            return nil, fmt.Errorf("failed at step %d: %w", i, err)
        }
        state = result
    }
    
    return state, nil
}
```

## Performance Optimization

### Parallel Execution Limits

```go
// Control parallelism
parallel := workflow.NewParallelAgent("controlled-parallel")
parallel.SetMaxConcurrency(5) // Limit to 5 concurrent agents

// Or use worker pool
pool := createWorkerPool(10)
parallel.SetWorkerPool(pool)
```

### Lazy Evaluation

```go
// Only run agents when needed
type LazyConditional struct {
    *workflow.ConditionalAgent
}

func (l *LazyConditional) AddLazyBranch(name string, condition func(*domain.State) bool, 
    agentFactory func() domain.BaseAgent) {
    
    l.AddBranch(name, condition, &LazyAgent{
        factory: agentFactory,
    })
}

type LazyAgent struct {
    factory func() domain.BaseAgent
    agent   domain.BaseAgent
    once    sync.Once
}

func (l *LazyAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
    l.once.Do(func() {
        l.agent = l.factory()
    })
    return l.agent.Run(ctx, state)
}
```

### Caching Results

```go
// Cache workflow results
type CachedWorkflow struct {
    workflow domain.BaseAgent
    cache    map[string]*domain.State
    mu       sync.RWMutex
}

func (c *CachedWorkflow) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
    key := generateCacheKey(state)
    
    // Check cache
    c.mu.RLock()
    if cached, ok := c.cache[key]; ok {
        c.mu.RUnlock()
        return cached.Clone(), nil
    }
    c.mu.RUnlock()
    
    // Run workflow
    result, err := c.workflow.Run(ctx, state)
    if err != nil {
        return nil, err
    }
    
    // Cache result
    c.mu.Lock()
    c.cache[key] = result.Clone()
    c.mu.Unlock()
    
    return result, nil
}
```

## Monitoring and Debugging

### Workflow Events

```go
// Monitor workflow execution
workflow.OnStepComplete(func(stepName string, duration time.Duration, err error) {
    if err != nil {
        log.Printf("Step %s failed after %v: %v", stepName, duration, err)
    } else {
        log.Printf("Step %s completed in %v", stepName, duration)
    }
})

// Track state changes
workflow.OnStateChange(func(before, after *domain.State) {
    changes := detectChanges(before, after)
    log.Printf("State changes: %v", changes)
})
```

### Workflow Visualization

```go
// Generate workflow diagram
func visualizeWorkflow(w domain.BaseAgent) string {
    switch wf := w.(type) {
    case *workflow.SequentialAgent:
        return visualizeSequential(wf)
    case *workflow.ParallelAgent:
        return visualizeParallel(wf)
    case *workflow.ConditionalAgent:
        return visualizeConditional(wf)
    case *workflow.LoopAgent:
        return visualizeLoop(wf)
    default:
        return w.Name()
    }
}
```

## Best Practices

### 1. Workflow Design
- Keep workflows focused on a single objective
- Use meaningful names for agents and steps
- Document the expected state at each step
- Plan for failure scenarios

### 2. State Management
- Minimize state size between steps
- Use clear, consistent key names
- Document required state keys
- Clean up temporary state data

### 3. Error Handling
- Implement appropriate retry strategies
- Provide fallback options
- Log errors with context
- Set reasonable timeouts

### 4. Performance
- Use parallel workflows when tasks are independent
- Set concurrency limits based on resources
- Cache expensive computations
- Monitor execution times

### 5. Testing
- Test each agent independently
- Test workflow integration
- Test error scenarios
- Test with realistic data volumes

## Next Steps

Now that you understand workflows:
- Check out [Examples Gallery](examples-gallery.md) for complete workflow examples
- See the [API Reference](../api/workflows.md) for detailed documentation
- Explore [Custom Agents](custom-agents.md) for building specialized agents

Ready to orchestrate complex multi-agent systems? Let's compose! 🎭