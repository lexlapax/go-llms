# Multi-Agent Coordination Example

This example demonstrates hierarchical multi-agent coordination patterns, showing how to build sophisticated systems where multiple specialized agents work together under coordinator agents to solve complex tasks.

## Overview

Multi-agent coordination enables building systems where:
- Specialist agents focus on specific domains (research, analysis, reporting)
- Coordinator agents orchestrate work across teams
- Agents can form hierarchical team structures
- Work can be parallelized automatically
- Results are aggregated intelligently

## Agent Architecture

### 1. Specialist Agents
Individual agents with focused capabilities:
```go
// Web research specialist
webResearcher := core.NewLLMAgent("web_researcher", "claude-3-haiku",
    core.WithTools(
        builtins.WebSearch(),
        builtins.WebScrape(),
    ),
)

// Data analysis specialist
dataAnalyst := core.NewLLMAgent("data_analyst", "gpt-4",
    core.WithTools(
        builtins.JSONProcess(),
        builtins.DataTransform(),
    ),
)

// Report writing specialist
reportWriter := core.NewLLMAgent("report_writer", "claude-3-opus",
    core.WithSystemPrompt("You are an expert technical writer..."),
)
```

### 2. Coordinator Agents
Team leads that manage specialist agents:
```go
// Research team coordinator
researchLead := core.NewLLMAgent("research_lead", "gpt-4",
    core.WithSystemPrompt("You coordinate research activities..."),
    core.WithTools(
        tools.AgentTool("web_researcher", webResearcher),
        tools.AgentTool("academic_researcher", academicResearcher),
        tools.AgentTool("news_researcher", newsResearcher),
    ),
)

// Analysis team coordinator
analysisLead := core.NewLLMAgent("analysis_lead", "claude-3-sonnet",
    core.WithSystemPrompt("You coordinate data analysis..."),
    core.WithTools(
        tools.AgentTool("data_analyst", dataAnalyst),
        tools.AgentTool("trend_analyst", trendAnalyst),
        tools.AgentTool("risk_analyst", riskAnalyst),
    ),
)
```

### 3. Team Structures
Hierarchical organization with main coordinator:
```go
// Main project coordinator
mainCoordinator := core.NewLLMAgent("project_coordinator", "gpt-4",
    core.WithSystemPrompt("You manage complex research projects..."),
    core.WithTools(
        tools.AgentTool("research_team", researchLead),
        tools.AgentTool("analysis_team", analysisLead),
        tools.AgentTool("report_writer", reportWriter),
    ),
)
```

## Key Patterns Demonstrated

### Pattern 1: Hierarchical Delegation
```go
// User request flows through hierarchy
ctx := agent.NewRunContext[interface{}](context.Background())
state := agent.NewState()
state.Set("task", "Research AI impact on healthcare industry")

// Main coordinator decides to delegate to research team
// Research lead delegates to specific researchers
// Results bubble up through hierarchy
result, err := mainCoordinator.Run(ctx, state)
```

### Pattern 2: Parallel Execution
Research team can execute searches in parallel:
```go
// Research lead creates parallel workflow
parallelResearch := workflow.NewParallelAgent("parallel_research",
    []domain.BaseAgent{webResearcher, academicResearcher, newsResearcher},
    workflow.WithMergeStrategy(workflow.MergeStrategyAppend),
)
```

### Pattern 3: State Sharing
Agents share context through state:
```go
// Research findings available to analysis team
state.Set("research_findings", researchResults)
state.Set("analysis_context", analysisParams)

// Each agent can access and update shared state
analysisResults, err := analysisLead.Run(ctx, state)
```

### Pattern 4: Conditional Routing
Coordinator decides which specialists to engage:
```go
// Analysis lead chooses analysts based on data type
conditionalAnalysis := workflow.NewConditionalAgent("conditional_analysis",
    []workflow.Branch{
        {
            Name: "quantitative",
            Condition: func(s domain.State) bool {
                return s.Get("data_type") == "quantitative"
            },
            Agent: dataAnalyst,
        },
        {
            Name: "qualitative",
            Condition: func(s domain.State) bool {
                return s.Get("data_type") == "qualitative"
            },
            Agent: trendAnalyst,
        },
    },
)
```

## Code Examples

### Example 1: Research Organization
```go
// Create research organization
researchOrg := createResearchOrganization()

// Execute complex research task
state := agent.NewState()
state.Set("topic", "Quantum computing applications in cryptography")
state.Set("depth", "comprehensive")
state.Set("sources", []string{"academic", "industry", "news"})

ctx := agent.NewRunContext[interface{}](context.Background())
result, err := researchOrg.Run(ctx, state)

// Result contains:
// - Aggregated research from all sources
// - Analysis from multiple perspectives
// - Comprehensive report with citations
```

### Example 2: Customer Service System
```go
// Create customer service platform
customerService := createCustomerServicePlatform()

// Handle customer inquiry
state := agent.NewState()
state.Set("customer_id", "12345")
state.Set("issue", "Cannot access account after password reset")
state.Set("priority", "high")

result, err := customerService.Run(ctx, state)

// System automatically:
// - Triages issue to tech support
// - Engages password reset specialist
// - Escalates if needed
// - Provides resolution
```

### Example 3: Document Processing Pipeline
```go
// Create document processing system
docProcessor := createDocumentProcessor()

// Process complex document
state := agent.NewState()
state.Set("document_path", "/path/to/document.pdf")
state.Set("requirements", []string{"extract", "summarize", "translate"})

result, err := docProcessor.Run(ctx, state)

// Pipeline handles:
// - Document type detection
// - Parallel processing tasks
// - Quality validation
// - Format conversion
```

## Real-World Applications

### 1. Research and Intelligence
- **Market Research**: Teams analyze competitors, trends, regulations
- **Due Diligence**: Parallel investigation of financial, legal, technical aspects
- **Threat Intelligence**: Coordinated monitoring of security threats

### 2. Content Production
- **News Generation**: Research, fact-check, write, edit pipeline
- **Technical Documentation**: Code analysis, API docs, tutorials
- **Marketing Campaigns**: Market research, content creation, optimization

### 3. Business Operations
- **Customer Support**: Triage, specialist routing, escalation
- **HR Processes**: Resume screening, interview scheduling, onboarding
- **Quality Assurance**: Multi-aspect product testing and reporting

### 4. Data Analysis
- **Business Intelligence**: Data collection, processing, visualization
- **Scientific Research**: Experiment design, data analysis, paper writing
- **Financial Analysis**: Market data, risk assessment, report generation

## Architecture Benefits

### 1. Modularity
- Each agent has single responsibility
- Easy to add/remove agents
- Clear separation of concerns
- Reusable agent components

### 2. Scalability
- Parallel execution where possible
- Dynamic team composition
- Horizontal scaling of specialist agents
- Load distribution across teams

### 3. Flexibility
- LLM-driven routing decisions
- Adaptive to different scenarios
- No hardcoded workflows
- Easy reconfiguration

### 4. Maintainability
- Clear hierarchical structure
- Isolated agent logic
- Centralized coordination
- Simplified testing

### 5. Performance
- Concurrent execution
- Efficient state sharing
- Result caching
- Resource optimization

## Running the Example

```bash
# Basic execution
go run main.go

# With custom configuration
GO_LLMS_OPENAI_API_KEY=your-key go run main.go

# Enable debug logging
GO_LLMS_DEBUG=all go run main.go
```

## Example Output

```
=== Multi-Agent Coordination Example ===

Creating research organization...
- Web Researcher (claude-3-haiku)
- Academic Researcher (gpt-3.5-turbo)
- News Researcher (claude-3-haiku)
- Research Lead (gpt-4)
- Data Analyst (gpt-4)
- Trend Analyst (claude-3-sonnet)
- Analysis Lead (claude-3-sonnet)
- Report Writer (claude-3-opus)
- Project Coordinator (gpt-4)

Executing research task: "AI impact on healthcare"

[Research Phase]
Research Lead coordinating parallel searches...
- Web Researcher: Found 15 relevant articles
- Academic Researcher: Found 8 peer-reviewed papers
- News Researcher: Found 12 recent news items

[Analysis Phase]
Analysis Lead processing findings...
- Data Analyst: Identified 5 key trends
- Trend Analyst: Projected 3 future scenarios

[Reporting Phase]
Report Writer creating final document...
- Executive Summary
- Detailed Findings
- Trend Analysis
- Recommendations

Task completed successfully!
Total agents involved: 9
Total execution time: 45.3s
```

## Best Practices

1. **Agent Design**
   - Keep agents focused on specific domains
   - Use clear, descriptive system prompts
   - Provide appropriate tools for each role

2. **Coordination**
   - Use coordinator agents for team management
   - Implement clear delegation patterns
   - Handle failures gracefully

3. **State Management**
   - Share only necessary information
   - Use structured data formats
   - Maintain state consistency

4. **Performance**
   - Parallelize independent tasks
   - Cache intermediate results
   - Monitor resource usage

## Related Examples

- `agent-simple-llm` - Basic agent usage
- `agent-handoff` - Agent delegation patterns
- `workflow-parallel` - Parallel execution patterns
- `workflow-conditional` - Conditional routing
- `agent-workflow-as-tool` - Using workflows as tools