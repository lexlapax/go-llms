# Examples Gallery

Explore working examples to learn go-llms through hands-on code.

## 🚀 Getting Started

### No API Key Required

These examples work without any setup - perfect for getting started!

#### **Hello World with Mock Provider**
📁 [simple](../../cmd/examples/simple/)  
⭐ Difficulty: Beginner

Your first go-llms program using a mock provider.

```bash
cd cmd/examples/simple && go run main.go
```

**What you'll learn:**
- Basic provider setup
- Making your first LLM call
- Understanding responses

**Expected output:**
```
Response: This is a mock response for: Hello, AI!
```

---

## 🤖 Building Your First Agent

### Basic Agent Creation

#### **Minimal LLM Agent**
📁 [agent-simple-llm](../../cmd/examples/agent-simple-llm/)  
⭐ Difficulty: Beginner  
🔑 Requires: OpenAI API key

Create your first agent in just a few lines of code.

```bash
export OPENAI_API_KEY="your-key"
cd cmd/examples/agent-simple-llm && go run main.go
```

**What you'll learn:**
- Creating an LLM agent
- Setting system prompts
- Basic agent interactions

**Try this:** Modify the system prompt to create different agent personalities!

---

## 🛠️ Working with Tools

### Calculator and Math

#### **Natural Language Calculator**
📁 [agent-calculator](../../cmd/examples/agent-calculator/)  
⭐ Difficulty: Beginner  
🔑 Requires: OpenAI API key (for LLM mode)

Let an agent solve math problems using natural language.

```bash
# Agent solves math with natural language
cd cmd/examples/agent-calculator && go run main.go

# Use calculator directly (no LLM needed)
go run main.go direct

# See tool information
go run main.go info
```

**What you'll learn:**
- How agents use tools
- Natural language to function calls
- Direct tool usage vs agent-mediated usage

**Try this:** Ask complex word problems like "If I have 23 apples and give away 40%, how many do I have left?"

### Web Tools

#### **Web Search and Scraping**
📁 [builtins-web-tools](../../cmd/examples/builtins-web-tools/)  
⭐ Difficulty: Intermediate

Search the web, fetch pages, and extract content.

```bash
cd cmd/examples/builtins-web-tools && go run main.go
```

**What you'll learn:**
- Web search with multiple engines
- Content extraction from URLs
- HTML scraping with CSS selectors
- Making HTTP requests

**Features demonstrated:**
- DuckDuckGo search (no API key needed)
- Web page fetching and cleaning
- Structured data extraction
- Custom HTTP requests

#### **REST API Client**
📁 [builtins-web-api-client](../../cmd/examples/builtins-web-api-client/)  
⭐ Difficulty: Intermediate

Make authenticated API calls with automatic error handling.

```bash
cd cmd/examples/builtins-web-api-client && go run main.go
```

**What you'll learn:**
- REST API interactions
- Authentication methods (API key, Bearer, Basic)
- Error handling and retries
- Response parsing

**Try this:** Modify to call your own APIs!

#### **GraphQL Client**
📁 [builtins-graphql-client](../../cmd/examples/builtins-graphql-client/)  
⭐ Difficulty: Advanced  
🔑 Requires: GitHub token (for GitHub GraphQL examples)

Execute GraphQL queries and mutations.

```bash
export GITHUB_TOKEN="your-token"
cd cmd/examples/builtins-graphql-client && go run main.go
```

**What you'll learn:**
- GraphQL query execution
- Schema introspection
- Variable handling
- GraphQL error handling

### File Operations

#### **File Management Suite**
📁 [builtins-file-tools](../../cmd/examples/builtins-file-tools/)  
⭐ Difficulty: Beginner

Complete file operations toolkit.

```bash
cd cmd/examples/builtins-file-tools && go run main.go
```

**What you'll learn:**
- Safe file reading/writing
- Directory listing with filters
- File search with regex
- Atomic file operations

**Safety features:**
- Permission checks
- Atomic writes
- Backup creation
- Path validation

### Data Processing

#### **JSON, CSV, and XML Processing**
📁 [builtins-data-tools](../../cmd/examples/builtins-data-tools/)  
⭐ Difficulty: Intermediate

Transform and query structured data.

```bash
cd cmd/examples/builtins-data-tools && go run main.go
```

**What you'll learn:**
- JSONPath queries
- CSV filtering and statistics
- XML to JSON conversion
- Data transformations (map, filter, reduce)

**Cool features:**
- Extract nested JSON values
- Calculate CSV column statistics
- Transform between formats

---

## 🎯 Structured Output

#### **Type-Safe LLM Responses**
📁 [agent-structured-output](../../cmd/examples/agent-structured-output/)  
⭐ Difficulty: Intermediate  
🔑 Requires: OpenAI API key

Get structured, validated data from LLMs.

```bash
export OPENAI_API_KEY="your-key"
cd cmd/examples/agent-structured-output && go run main.go
```

**What you'll learn:**
- Define schemas for LLM output
- Automatic validation
- Type safety with Go structs
- Error recovery strategies

**Example output:**
```json
{
  "name": "iPhone 15 Pro",
  "price": 999.99,
  "inStock": true,
  "features": ["A17 Pro chip", "Titanium design", "Action button"]
}
```

---

## 🔄 Workflows

### Sequential Processing

#### **Step-by-Step Workflows**
📁 [workflow-sequential](../../cmd/examples/workflow-sequential/)  
⭐ Difficulty: Intermediate

Chain multiple agents in sequence.

```bash
cd cmd/examples/workflow-sequential && go run main.go
```

**What you'll learn:**
- Creating multi-step pipelines
- State passing between agents
- Error propagation
- Step dependencies

**Use case:** Data pipeline: Extract → Transform → Analyze → Report

### Parallel Execution

#### **Concurrent Agent Execution**
📁 [workflow-parallel](../../cmd/examples/workflow-parallel/)  
⭐ Difficulty: Intermediate

Run multiple agents simultaneously.

```bash
cd cmd/examples/workflow-parallel && go run main.go
```

**What you'll learn:**
- Parallel agent execution
- Result merging strategies
- Performance optimization
- Concurrency control

**Use case:** Analyze text for sentiment, keywords, and entities simultaneously

### Conditional Logic

#### **Dynamic Workflow Branching**
📁 [workflow-conditional](../../cmd/examples/workflow-conditional/)  
⭐ Difficulty: Advanced

Route workflows based on conditions.

```bash
cd cmd/examples/workflow-conditional && go run main.go
```

**What you'll learn:**
- Conditional branching
- Dynamic routing
- State-based decisions
- Complex workflow logic

**Use case:** Customer support routing based on issue type

---

## 🌐 Multi-Provider Strategies

#### **Reliability Through Multiple Providers**
📁 [provider-multi](../../cmd/examples/provider-multi/)  
⭐ Difficulty: Advanced  
🔑 Requires: Multiple API keys

Use multiple LLM providers for reliability and performance.

```bash
export OPENAI_API_KEY="your-key"
export ANTHROPIC_API_KEY="your-key"
cd cmd/examples/provider-multi && go run main.go
```

**What you'll learn:**
- Fallback strategies
- Load balancing
- Provider selection
- Cost optimization

**Strategies demonstrated:**
- Primary with fallback
- Fastest response wins
- Round-robin distribution

---

## 🎨 Advanced Patterns

### Custom Research Agent

#### **Multi-Source Research System**
📁 [agent-custom-research](../../cmd/examples/agent-custom-research/)  
⭐ Difficulty: Advanced  
🔑 Requires: OpenAI API key + search API keys

Build a sophisticated research agent.

```bash
export OPENAI_API_KEY="your-key"
export BRAVE_API_KEY="your-key"  # Optional
export TAVILY_API_KEY="your-key"  # Optional
cd cmd/examples/agent-custom-research && go run main.go
```

**What you'll learn:**
- Multi-phase research workflows
- Source aggregation
- Fact checking
- Report generation

**Architecture:**
- Web search phase
- Content analysis phase  
- Synthesis phase
- Report generation

### State Persistence

#### **Save and Resume Agent State**
📁 [agent-state-persistence](../../cmd/examples/agent-state-persistence/)  
⭐ Difficulty: Advanced

Persist agent conversations and state.

```bash
cd cmd/examples/agent-state-persistence && go run main.go
```

**What you'll learn:**
- State serialization
- Session management
- Conversation history
- Resume capabilities

---

## 🖼️ Multimodal Examples

#### **Image, Audio, and Video Processing**
📁 [provider-multimodal](../../cmd/examples/provider-multimodal/)  
⭐ Difficulty: Intermediate  
🔑 Requires: Provider API key

Work with images, audio, and video content.

```bash
export OPENAI_API_KEY="your-key"
cd cmd/examples/provider-multimodal && go run main.go -provider openai -mode image -a image.jpg
```

**What you'll learn:**
- Image analysis
- Audio transcription
- Video understanding
- Mixed media conversations

**Supported modes:**
- `-mode image`: Analyze images
- `-mode audio`: Transcribe audio
- `-mode video`: Process video content

---

## 📚 Learning Paths

### Beginner Path
1. Start with [simple](../../cmd/examples/simple/) - no API key needed
2. Try [agent-calculator](../../cmd/examples/agent-calculator/) - see tools in action
3. Explore [builtins-file-tools](../../cmd/examples/builtins-file-tools/) - practical file operations
4. Build with [agent-simple-llm](../../cmd/examples/agent-simple-llm/) - your first real agent

### Intermediate Path
1. Master [agent-structured-output](../../cmd/examples/agent-structured-output/) - reliable data extraction
2. Learn [workflow-sequential](../../cmd/examples/workflow-sequential/) - multi-step processes
3. Try [workflow-parallel](../../cmd/examples/workflow-parallel/) - concurrent execution
4. Implement [agent-error-handling](../../cmd/examples/agent-error-handling/) - robust applications

### Advanced Path
1. Study [agent-custom-research](../../cmd/examples/agent-custom-research/) - complex orchestration
2. Implement [provider-multi](../../cmd/examples/provider-multi/) - high availability
3. Master [workflow-conditional](../../cmd/examples/workflow-conditional/) - dynamic workflows
4. Build with [agent-multi-coordination](../../cmd/examples/agent-multi-coordination/) - agent teams

---

## 💡 Tips for Success

### Running Examples

1. **Check Requirements**: Each example's README lists required API keys
2. **Start Simple**: Begin with examples that don't need API keys
3. **Read the Code**: Examples are well-commented learning resources
4. **Modify and Experiment**: Change prompts, add tools, adjust workflows
5. **Check Debug Output**: Use `DEBUG=1` for detailed logging

### Environment Setup

```bash
# Basic setup for most examples
export OPENAI_API_KEY="your-key"
export ANTHROPIC_API_KEY="your-key"  
export GEMINI_API_KEY="your-key"

# Optional for specific examples
export BRAVE_API_KEY="your-key"      # Web search
export TAVILY_API_KEY="your-key"     # AI-optimized search
export GITHUB_TOKEN="your-key"       # GitHub API examples

# Enable debug mode
export DEBUG=1
```

### Common Patterns

All examples follow similar patterns:

```go
// 1. Create provider
provider := provider.NewOpenAIProvider(apiKey, model)

// 2. Create agent 
agent := core.NewLLMAgent("my-agent", provider)

// 3. Add tools (optional)
agent.AddTool(tools.MustGetTool("web_search"))

// 4. Run with state
state := domain.NewState()
state.Set("task", "your task here")
result, err := agent.Run(ctx, state)
```

---

## 🛠️ Building Your Own

After exploring these examples:

1. **Copy an example** as your starting point
2. **Modify gradually** - change one thing at a time
3. **Combine patterns** - mix tools, agents, and workflows
4. **Share your creation** - contribute back to the community!

---

Ready to explore? Pick an example and start coding! 🚀