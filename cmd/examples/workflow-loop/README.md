# Loop Workflow Example

This example demonstrates the Loop Workflow agent capabilities for iterative processing patterns. The LoopAgent provides powerful loop control features including while loops, count loops, until loops, and retry logic with proper termination conditions.

## Features Demonstrated

### 1. Count Loop
- Fixed number of iterations for batch processing
- Automatic iteration counting
- State management across iterations

### 2. While Loop
- Continues while a condition remains true
- Convergence-based termination
- Dynamic condition evaluation

### 3. Retry Loop (Until Loop)
- Continues until a success condition is met
- Exponential backoff simulation
- Maximum attempt limits

### 4. Advanced Loop Features
- Result collection from each iteration
- State passthrough control
- Iteration delays
- Error handling strategies

## Loop Types Supported

- **Count Loop**: Execute a fixed number of times
- **While Loop**: Continue while condition is true
- **Until Loop**: Continue until condition becomes true
- **Custom Loop**: Combine multiple termination conditions

## Termination Conditions

The LoopAgent supports multiple termination conditions:

1. **Maximum Iterations**: `WithMaxIterations(n)`
2. **Maximum Duration**: `WithMaxDuration(duration)`
3. **While Condition**: `WithWhileCondition(func)`
4. **Until Condition**: `WithUntilCondition(func)`

## Advanced Features

### Result Collection
```go
loop.WithCollectResults(true)
results := loop.GetIterationResults()
```

### State Management
```go
// Pass state between iterations
loop.WithPassStateThrough(true)

// Fresh state for each iteration
loop.WithPassStateThrough(false)
```

### Error Handling
```go
// Break on first error
loop.WithBreakOnError(true)

// Continue despite errors
loop.WithBreakOnError(false)
```

### Iteration Control
```go
// Add delay between iterations
loop.WithIterationDelay(100 * time.Millisecond)
```

## Mock Agents

The example includes several mock agents that simulate real-world scenarios:

- **Batch Processor**: Simulates data batch processing
- **Optimizer Agent**: Simulates iterative optimization with convergence
- **Retry Agent**: Simulates API calls with success after multiple attempts
- **Survey Agent**: Simulates processing multiple survey responses

## Usage

```bash
# Run the example
go run cmd/examples/workflow-loop/main.go

# Build and run
make build-example EXAMPLE=workflow-loop
./bin/workflow-loop
```

## Real LLM Integration

The example attempts to use real LLM agents first and falls back to mock agents if API keys are not configured:

- **Claude**: For batch processing and retry scenarios
- **GPT-4**: For optimization and survey analysis scenarios

Set the appropriate environment variables to enable real LLM integration:
- `GO_LLMS_ANTHROPIC_API_KEY`
- `GO_LLMS_OPENAI_API_KEY`

## Output Example

```
=== Count Loop Example ===
Processing a batch of items with a fixed number of iterations...

Starting batch processing...
Batch processing completed in 523ms
Final batch result: Batch item processed
Total iterations: 5
Total duration: 523ms

=== While Loop Example ===
Iterative optimization until convergence...

Starting optimization...
Optimization completed in 402ms
Final error rate: 0.0051
Final optimization result: Optimization step completed
Total iterations: 4
Convergence achieved: true

=== Retry Loop Example ===
API call with retry logic and exponential backoff...

Starting API calls with retry logic...
API retry completed in 251ms
API call successful: true
Final API response: API call successful
Total attempts: 3
Max attempts reached: false

=== Advanced Loop Features Example ===
Demonstrating result collection, state management, and loop control...

Starting survey analysis...
Survey analysis completed in 233ms
Surveys analyzed: 3
Generated insights: [Insight 1: Analysis completed for survey response, ...]
Collected 3 iteration results:
  Iteration 0: Duration=61ms, Success=true
  Iteration 1: Duration=62ms, Success=true  
  Iteration 2: Duration=63ms, Success=true
Total iterations: 3
Total duration: 233ms
```

## Architecture

The LoopAgent is built on the BaseWorkflowAgent foundation and provides:

- Thread-safe iteration counting
- Comprehensive event emission
- Hook integration for monitoring
- Flexible condition evaluation
- Robust error handling
- Metadata collection

This makes it suitable for production workflows requiring reliable iterative processing with proper monitoring and control.