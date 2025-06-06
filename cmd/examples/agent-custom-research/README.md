# Agent Custom Research Example

This example demonstrates how to build a sophisticated custom agent by extending `BaseAgentImpl` and orchestrating multiple specialized agents through code-based coordination.

## Architecture

The example implements a research pipeline with the following architecture:

```
ResearchAgent (extends BaseAgentImpl)
├── Code-based orchestration of:
│   ├── MultiSearchAgent (extends BaseAgentImpl)
│   │   └── Runs parallel searches: Tavily, Brave, Serpapi, Serper.dev, DuckDuckGo
│   ├── DuplicateFilterAgent (uses LLMAgent with custom prompt)
│   │   └── Intelligent deduplication and relevance scoring
│   ├── ContentAnalyzerAgent (uses LLMAgent with custom prompt)  
│   │   └── Extracts key insights from deduplicated results
│   └── ReportGeneratorAgent (uses LLMAgent with custom prompt)
│       └── Synthesizes final research report
```

## Key Concepts Demonstrated

### 1. Custom Agent Development
- **Extending BaseAgentImpl**: Shows how to build agents with full control over execution
- **Phase-based orchestration**: Implements a 4-phase research pipeline
- **Custom state management**: Manages complex state flow between phases
- **Event emission**: Tracks progress and errors through events

### 2. Code-Based Orchestration
- **No library sub-agents**: Uses direct method calls instead of framework features
- **Explicit control flow**: Clear, debuggable execution path
- **Error handling**: Graceful degradation when components fail
- **Progress tracking**: Real-time updates on research phases

### 3. Multi-Engine Search
- **Parallel execution**: Searches multiple engines simultaneously
- **Query variations**: Different queries for comprehensive coverage
- **API key management**: Flexible key injection via environment or state
- **Result aggregation**: Combines results with source metadata

### 4. LLMAgent Usage Pattern
Instead of extending LLMAgent, the example shows the correct pattern:
```go
// Create specialized LLMAgent instances
duplicateFilter, _ := core.NewAgentFromString("filter", "claude")
duplicateFilter.SetSystemPrompt("You are a deduplication expert...")

// Use them as components
result, _ := duplicateFilter.Run(ctx, state)
```

## Running the Example

### Basic Usage
```bash
# Run with default topic
go run main.go

# Run with custom topic
go run main.go "quantum computing applications"

# Run with debug logging enabled
DEBUG=1 go run main.go
```

### Environment Variables

For web search (at least one recommended):
- `BRAVE_API_KEY` - Brave Search API key
- `TAVILY_API_KEY` - Tavily API key (best results)
- `SERPAPI_API_KEY` - SerpAPI key
- `SERPERDEV_API_KEY` - Serper.dev API key

For LLM processing (optional, uses mocks if not set):
- `OPENAI_API_KEY` or `ANTHROPIC_API_KEY` - For LLM sub-agents
- `LLM_PROVIDER` - Set to "openai" or "anthropic" (default: "claude")

For debugging:
- `DEBUG=1` - Enable debug logging with detailed LLM agent activity

### Example Output
```
=== Advanced Research Agent Example ===

This example demonstrates:
- Custom agent extending BaseAgentImpl (not LLMAgent)
- Code-based orchestration of multiple agents
- Parallel search across multiple engines
- LLMAgent instances for intelligent processing
- State management and error handling

Research Topic: quantum computing applications in cryptography
------------------------------------------------------------
🔍 Phase 1: Executing parallel searches for 'quantum computing applications in cryptography'
  🔎 Searching duckduckgo (overview): quantum computing...
  🔎 Searching brave (latest): quantum computing...
  🔎 Searching tavily (expert): quantum computing...
  ✅ Found 120 total results across all engines
🔄 Phase 2: Deduplicating and ranking results
  ✅ Reduced to 45 unique results
📊 Phase 3: Analyzing content and extracting insights
📝 Phase 4: Generating comprehensive report

============================================================
RESEARCH COMPLETE
============================================================

# Research Report: Quantum Computing Applications in Cryptography

## Executive Summary
[Generated report content...]
```

## Implementation Details

### ResearchAgent
- Extends `BaseAgentImpl` for full control
- Implements custom `Run()` method with 4 phases
- Manages state flow and error handling
- Emits events for monitoring

### MultiSearchAgent
- Also extends `BaseAgentImpl`
- Executes parallel searches with goroutines
- Uses different query variations per engine
- Aggregates results with metadata

### LLM Sub-Agents
- Created as instances, not by extending
- Each has a specialized system prompt
- Falls back to mocks when LLM unavailable
- Processes specific aspects of research

## Benefits of This Architecture

1. **Flexibility**: Full control over execution flow
2. **Debuggability**: Clear code path without framework magic
3. **Scalability**: Easy to add new search engines or processing steps
4. **Reliability**: Graceful degradation with mocks
5. **Performance**: Parallel search execution

## Extending the Example

To add new capabilities:

1. **Add a new search engine**: Update `MultiSearchAgent.engines`
2. **Add a new processing phase**: Create new LLMAgent with custom prompt
3. **Customize search queries**: Modify the `queries` map
4. **Change report format**: Update the report generator prompt

## Common Patterns

### Creating Specialized LLM Agents
```go
agent, _ := core.NewAgentFromString("name", "provider")
agent.SetSystemPrompt("Specialized instructions...")
```

### Code-Based Orchestration
```go
// Phase 1
results1, err := r.phase1(ctx, input)
if err != nil {
    // Handle error
}

// Phase 2 - uses results from Phase 1
results2, err := r.phase2(ctx, results1)
```

### Parallel Execution Without Workflows
```go
var wg sync.WaitGroup
for _, item := range items {
    wg.Add(1)
    go func(i Item) {
        defer wg.Done()
        // Process item
    }(item)
}
wg.Wait()
```

This example provides a blueprint for building sophisticated multi-agent systems while maintaining code clarity and control.