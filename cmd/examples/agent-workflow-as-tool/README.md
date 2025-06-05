# Agent Workflow as Tool Example

This example demonstrates a sophisticated multi-stage research pipeline where workflow agents are wrapped as tools and orchestrated by a main LLM agent.

## Architecture Overview

```
Research Coordinator (Main LLM Agent)
├── Web Search Tool
├── Web Fetch Tool
├── Analysis Pipeline Tool (SequentialAgent wrapped as tool)
│   ├── Content Analyzer Agent
│   ├── Fact Checker Agent
│   └── Summary Generator Agent
├── Comparison Tool (ParallelAgent wrapped as tool)
│   ├── Source A Analyzer
│   └── Source B Analyzer
└── File Write Tool
```

## Components

### 1. Research Coordinator (Main LLM Agent)
The main orchestrator that:
- Manages the entire research process
- Decides which tools to use and when
- Coordinates between different analysis workflows
- Produces the final research report

### 2. Analysis Pipeline (Sequential Workflow)
A three-stage sequential processing pipeline:

**Stage 1: Content Analyzer**
- Extracts key points, entities, and topics
- Identifies main arguments and claims
- Analyzes tone and perspective

**Stage 2: Fact Checker**
- Verifies claims against reliable sources
- Distinguishes facts from opinions
- Rates credibility

**Stage 3: Summary Generator**
- Creates executive summary
- Lists verified facts
- Provides overall assessment

### 3. Comparison Tool (Parallel Workflow)
Analyzes two sources simultaneously:
- **Source Analyzer A**: Analyzes first source
- **Source Analyzer B**: Analyzes second source
- **Custom Merge Strategy**: Combines results to show similarities and differences

## Key Features

### Agent-to-Tool Conversion
- Workflow agents are wrapped as tools using `AgentTool`
- Maintains full agent capabilities while exposing tool interface
- Seamless integration with LLM agent's tool-calling mechanism

### State Management
- Each agent maintains its own state
- States are passed through sequential pipeline
- Parallel results are merged intelligently

### Error Handling
- Timeout support for long-running operations
- Graceful error propagation through pipeline
- Fallback strategies for failed operations

## Usage

### Basic Usage
```bash
go run main.go
```

### Custom Research Query
```bash
go run main.go "Research the environmental impact of electric vehicles vs traditional cars"
```

### Environment Variables
Requires LLM provider configuration:
```bash
# OpenAI
export OPENAI_API_KEY="your-key"

# Anthropic
export ANTHROPIC_API_KEY="your-key"

# Google Gemini
export GEMINI_API_KEY="your-key"
```

## Example Flow

1. **User Query**: "Research and compare perspectives on AI safety from two different sources"

2. **Research Coordinator Actions**:
   - Uses `web_search` to find relevant articles
   - Uses `web_fetch` to retrieve content from two authoritative sources
   - Calls `comparison-agent` tool:
     - Both source analyzers work in parallel
     - Results merged showing different perspectives
   - Calls `analysis-pipeline` tool:
     - Content analysis → Fact checking → Summary generation
   - Uses `file_write` to save final report

3. **Output**: Comprehensive research report with:
   - Analyzed perspectives from both sources
   - Fact-checked claims
   - Similarities and differences highlighted
   - Executive summary
   - Saved to file for future reference

## Benefits

1. **Modularity**: Each agent has a specific, focused responsibility
2. **Reusability**: Workflow agents can be reused in different contexts
3. **Scalability**: Easy to add more stages or parallel branches
4. **Performance**: Parallel processing saves time
5. **Flexibility**: Components can be swapped without changing structure
6. **Transparency**: Clear audit trail of analysis steps

## Extending the Example

### Adding More Analysis Stages
```go
sentimentAnalyzer := createSentimentAnalyzer(llmProvider)
analysisPipeline.WithAgents(
    contentAnalyzer,
    sentimentAnalyzer,  // New stage
    factChecker,
    summaryGenerator,
)
```

### Adding More Parallel Branches
```go
sourceAnalyzerC := createSourceAnalyzer("C", llmProvider)
comparisonAgent.WithAgents(
    sourceAnalyzerA,
    sourceAnalyzerB,
    sourceAnalyzerC,  // Third source
)
```

### Custom Tools
The research coordinator can be extended with additional tools:
- Database query tools
- API integration tools
- Specialized analysis tools
- Visualization tools

## Error Handling

The example includes comprehensive error handling:
- Provider initialization errors
- Agent creation failures
- Tool execution errors
- Timeout handling
- Merge strategy errors

## Performance Considerations

- Parallel analysis reduces total execution time
- Timeouts prevent hanging on slow operations
- Concurrency limits prevent resource exhaustion
- State passing minimizes redundant processing