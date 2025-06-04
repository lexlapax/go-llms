# Agent Simple LLM Example

This example demonstrates the ultra-simple agent creation feature introduced in Phase 2 of the agent architecture restructuring.

## Features

- Create AI agents with minimal code using string specifications
- Support for provider/model format: `"openai/gpt-4"`
- Model aliases for convenience: `"claude"`, `"gemini"`, `"gpt-4"`
- Automatic provider inference from model names
- Automatic API key resolution from environment variables

## Usage

```bash
# Set your API keys
export OPENAI_API_KEY=your-key-here
export ANTHROPIC_API_KEY=your-key-here
export GEMINI_API_KEY=your-key-here

# Build and run
go build -o agent-simple-llm
./agent-simple-llm
```

## Examples

### Ultra-Simple Agent Creation

```go
// Using aliases
agent, _ := core.NewAgentFromString("my-agent", "claude")
agent, _ := core.NewAgentFromString("my-agent", "gemini")
agent, _ := core.NewAgentFromString("my-agent", "gpt-4")

// Using provider/model format
agent, _ := core.NewAgentFromString("my-agent", "openai/gpt-4o-mini")
agent, _ := core.NewAgentFromString("my-agent", "anthropic/claude-3-opus-latest")
agent, _ := core.NewAgentFromString("my-agent", "gemini/gemini-2.0-flash")

// Model inference (provider detected from model name)
agent, _ := core.NewAgentFromString("my-agent", "gpt-4o-mini")    // → openai
agent, _ := core.NewAgentFromString("my-agent", "claude-3-haiku") // → anthropic
agent, _ := core.NewAgentFromString("my-agent", "gemini-2.0-flash") // → gemini
```

### Running the Agent

```go
// State-based interface (the only way now)
state := domain.NewState()
state.Set("prompt", "What is 2+2?")

resultState, err := agent.Run(ctx, state)
if err != nil {
    log.Fatal(err)
}

// Get the result
if result, exists := resultState.Get("result"); exists {
    fmt.Printf("Response: %v\n", result)
}
```

## Supported Aliases

The following aliases are preconfigured for convenience:

### OpenAI
- `"gpt-4"` → `"openai/gpt-4"`
- `"gpt-4o"` → `"openai/gpt-4o"`
- `"o1"` → `"openai/o1"`
- `"o3"` → `"openai/o3"`

### Anthropic
- `"claude"` → `"anthropic/claude-3-7-sonnet-latest"`
- `"opus"` → `"anthropic/claude-3-opus-latest"`
- `"sonnet"` → `"anthropic/claude-3-7-sonnet-latest"`
- `"haiku"` → `"anthropic/claude-3-5-haiku-latest"`

### Google/Gemini
- `"gemini"` → `"google/gemini-2.0-flash"`
- `"gemini-pro"` → `"google/gemini-2.5-pro-preview-05-06"`
- `"flash"` → `"google/gemini-2.0-flash"`

## Environment Variables

The library automatically looks for API keys in these environment variables:

- OpenAI: `OPENAI_API_KEY` or `GO_LLMS_OPENAI_API_KEY`
- Anthropic: `ANTHROPIC_API_KEY` or `GO_LLMS_ANTHROPIC_API_KEY`
- Google: `GEMINI_API_KEY` or `GO_LLMS_GEMINI_API_KEY`

## Error Handling

If an API key is missing, you'll get a helpful error message:

```
Failed to create agent: missing API key for provider 'openai'. Set OPENAI_API_KEY or GO_LLMS_OPENAI_API_KEY environment variable
```