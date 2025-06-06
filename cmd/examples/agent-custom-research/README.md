# Custom Research Assistant Agent Example

This example demonstrates how to create a sophisticated custom agent that extends `BaseAgent` and showcases advanced agent features.

## Features Demonstrated

1. **Custom Agent Implementation**
   - Extends `BaseAgentImpl` 
   - Implements custom `Run` method with complex logic
   - Manages state throughout multi-phase process

2. **Sub-Agent Coordination**
   - Web Searcher Agent - finds relevant sources
   - Summarizer Agent - extracts key points from articles
   - Fact Checker Agent - verifies claims and checks contradictions

3. **Tool Integration**
   - Uses `web_search` tool to find information
   - Uses `web_fetch` tool to retrieve content
   - Demonstrates proper tool context creation

4. **State Management**
   - Tracks research progress through phases
   - Accumulates findings in state
   - Preserves metadata (timing, sources)

5. **Error Handling**
   - Graceful degradation when tools fail
   - Fallback to mock agents when no LLM available
   - Continues research even with partial failures

## Architecture

```
ResearchAssistant (Custom Agent)
├── Web Search Phase
│   └── web_search tool
├── Information Gathering Phase  
│   └── web_fetch tool (multiple calls)
├── Summarization Phase
│   └── Summarizer sub-agent (LLM or mock)
├── Fact Checking Phase
│   └── Fact Checker sub-agent (LLM or mock)
└── Report Synthesis Phase
    └── Internal logic
```

## Running the Example

```bash
# Basic usage (will use mock agents if no API keys)
go run cmd/examples/agent-custom-research/main.go

# With web search API (e.g., Brave Search)
export SEARCH_API_KEY="your-brave-search-api-key"
go run cmd/examples/agent-custom-research/main.go

# With full LLM support for summarization and fact-checking
export OPENAI_API_KEY="your-api-key"  # or ANTHROPIC_API_KEY, GEMINI_API_KEY
export SEARCH_API_KEY="your-search-api-key"
go run cmd/examples/agent-custom-research/main.go
```

## Output

The agent produces a comprehensive research report including:
- Executive summary
- Key findings from multiple sources
- Fact-checked information
- Source citations
- Research metadata (duration, source count)

## Customization

You can customize the research assistant by:

1. **Changing the research depth**: Modify `maxSources` to analyze more/fewer sources
2. **Adding more sub-agents**: Create specialized agents for specific analysis tasks
3. **Enhancing the synthesis**: Improve the report generation logic
4. **Adding persistence**: Save research results to files or databases
5. **Implementing caching**: Cache web fetches to avoid repeated API calls

## Key Concepts

### Custom Agent Pattern
```go
type ResearchAssistant struct {
    *core.BaseAgentImpl
    // Custom fields for sub-agents and configuration
    webSearcher  domain.BaseAgent
    summarizer   domain.BaseAgent
    maxSources   int
}

func (r *ResearchAssistant) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
    // Multi-phase execution logic
    // Coordinate sub-agents
    // Aggregate results
    // Return synthesized state
}
```

### Sub-Agent Coordination
```go
// Create sub-agent
summarizer, err := createSummarizerAgent()

// Use sub-agent in research flow  
result, err := r.summarizer.Run(ctx, summaryState)
```

### Tool Usage in Custom Agent
```go
// Get tool from agent's tool registry
searchTool, ok := r.GetTool("web_search")

// Create tool context
toolCtx := &domain.ToolContext{
    Context: ctx.Context,
    State:   domain.NewStateReader(ctx.State),
    Agent:   ctx.Agent,
    RunID:   ctx.RunID,
}

// Execute tool
result, err := searchTool.Execute(toolCtx, params)
```

This example shows how to build production-ready custom agents that can handle complex, multi-step tasks with proper error handling and state management.