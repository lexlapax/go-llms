// ABOUTME: Example demonstrating usage of built-in components including tools, agents, and workflows
// ABOUTME: Shows how to discover, configure, and use pre-built components from the registry

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	// Import built-in components - this triggers auto-registration
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/data"     // Import for side effects (registration)
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/datetime" // Import for side effects (registration)
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/feed"     // Import for side effects (registration)
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/file"     // Import for side effects (registration)
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/system"   // Import for side effects (registration)
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"      // Import for side effects (registration)
	"github.com/lexlapax/go-llms/pkg/agent/core"
	agentDomain "github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
)

// Helper types for creating a minimal ToolContext for standalone tool execution

// minimalStateReader implements StateReader interface with empty state
type minimalStateReader struct {
	state *agentDomain.State
}

func (m *minimalStateReader) Get(key string) (interface{}, bool) {
	return m.state.Get(key)
}

func (m *minimalStateReader) Values() map[string]interface{} {
	return m.state.Values()
}

func (m *minimalStateReader) GetArtifact(id string) (*agentDomain.Artifact, bool) {
	return m.state.GetArtifact(id)
}

func (m *minimalStateReader) Artifacts() map[string]*agentDomain.Artifact {
	return m.state.Artifacts()
}

func (m *minimalStateReader) Messages() []agentDomain.Message {
	return m.state.Messages()
}

func (m *minimalStateReader) GetMetadata(key string) (interface{}, bool) {
	return m.state.GetMetadata(key)
}

func (m *minimalStateReader) Has(key string) bool {
	return m.state.Has(key)
}

func (m *minimalStateReader) Keys() []string {
	values := m.state.Values()
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	return keys
}

// minimalEventEmitter implements EventEmitter interface with no-op methods
type minimalEventEmitter struct{}

func (m *minimalEventEmitter) Emit(eventType agentDomain.EventType, data interface{}) {}
func (m *minimalEventEmitter) EmitProgress(current, total int, message string)        {}
func (m *minimalEventEmitter) EmitMessage(message string)                             {}
func (m *minimalEventEmitter) EmitError(err error)                                    {}
func (m *minimalEventEmitter) EmitCustom(eventName string, data interface{})          {}

// createToolContext creates a minimal ToolContext for standalone tool execution
func createToolContext(ctx context.Context) *agentDomain.ToolContext {
	state := agentDomain.NewState()
	stateReader := &minimalStateReader{state: state}

	toolCtx := &agentDomain.ToolContext{
		Context:   ctx,
		State:     stateReader,
		RunID:     "standalone-execution",
		Retry:     0,
		StartTime: time.Now(),
		Events:    &minimalEventEmitter{},
		Agent: agentDomain.AgentInfo{
			ID:          "standalone",
			Name:        "standalone-tool-executor",
			Description: "Minimal agent for standalone tool execution",
			Type:        agentDomain.AgentTypeLLM,
			Metadata:    make(map[string]interface{}),
		},
	}

	return toolCtx
}

func main() {
	// Display registry summary
	fmt.Println("=== Built-in Components Registry ===")
	fmt.Println()

	// List all registered tools
	fmt.Println("Available Tools:")
	allTools := tools.Tools.List()
	fmt.Printf("Total tools registered: %d\n\n", len(allTools))

	// Group tools by category for better display
	categories := make(map[string][]string)
	for _, entry := range allTools {
		categories[entry.Metadata.Category] = append(categories[entry.Metadata.Category], entry.Metadata.Name)
	}

	// Display tools by category
	for category, toolNames := range categories {
		fmt.Printf("%s Tools (%d):\n", category, len(toolNames))
		for _, name := range toolNames {
			tool, _ := tools.GetTool(name)
			fmt.Printf("  - %s: %s\n", name, tool.Description())
		}
		fmt.Println()
	}

	// Search for specific tools
	fmt.Println("=== Tool Discovery Examples ===")
	fmt.Println()

	// Search by keyword
	fmt.Println("Searching for 'fetch' tools:")
	fetchTools := tools.Tools.Search("fetch")
	for _, entry := range fetchTools {
		fmt.Printf("  - %s: %s\n", entry.Metadata.Name, entry.Metadata.Description)
	}
	fmt.Println()

	// List tools by different categories
	categoryNames := map[string]string{
		"web":      "Web Tools",
		"file":     "File Tools",
		"system":   "System Tools",
		"data":     "Data Processing Tools",
		"datetime": "Date/Time Tools",
		"feed":     "Feed Processing Tools",
	}

	for cat, name := range categoryNames {
		fmt.Printf("%s:\n", name)
		categoryTools := tools.Tools.ListByCategory(cat)
		for _, entry := range categoryTools {
			fmt.Printf("  - %s\n", entry.Metadata.Name)
		}
		fmt.Println()
	}

	// Demonstrate tool usage
	fmt.Println("=== Using Built-in Tools ===")
	fmt.Println()

	// Demonstrate different tool usage examples
	ctx := context.Background()
	toolCtx := createToolContext(ctx)

	// Example 1: Using a file tool
	fmt.Println("Example 1: Using File Tool")
	readTool, found := tools.GetTool("read_file")
	if found {
		// Show tool schema
		fmt.Printf("Tool: %s\n", readTool.Name())
		fmt.Printf("Description: %s\n", readTool.Description())
		fmt.Printf("Parameters: %+v\n", readTool.ParameterSchema())
		fmt.Println()
	}

	// Example 2: Using a datetime tool
	fmt.Println("Example 2: Using DateTime Tool")
	dateTool, found := tools.GetTool("datetime_now")
	if found {
		result, err := dateTool.Execute(toolCtx, map[string]interface{}{
			"timezone": "UTC",
			"format":   "RFC3339",
		})
		if err != nil {
			log.Printf("Error executing datetime_now: %v", err)
		} else {
			fmt.Printf("Current time: %v\n", result)
		}
		fmt.Println()
	}

	// Example 3: Using a data transformation tool
	fmt.Println("Example 3: Using Data Transform Tool")
	transformTool, found := tools.GetTool("data_transform")
	if found {
		result, err := transformTool.Execute(toolCtx, map[string]interface{}{
			"data":      []interface{}{5, 2, 8, 1, 9},
			"operation": "sort",
			"options": map[string]interface{}{
				"reverse": true,
			},
		})
		if err != nil {
			log.Printf("Error executing data_transform: %v", err)
		} else {
			fmt.Printf("Sorted data (descending): %v\n", result)
		}
		fmt.Println()
	}

	// Create a simple agent with multiple built-in tools
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Println("OPENAI_API_KEY not set, skipping agent demonstration")
		fmt.Println()
		fmt.Println("To see agent examples, set your OPENAI_API_KEY environment variable")
	} else {
		// Create provider
		p := provider.NewOpenAIProvider(apiKey, "gpt-4o-mini")

		// Get multiple tools
		webFetch, _ := tools.GetTool("web_fetch")
		jsonProcess, _ := tools.GetTool("json_process")
		dateCalc, _ := tools.GetTool("datetime_calculate")

		// Create agent with multiple built-in tools
		agent := core.NewAgent("discovery-agent", p)
		agent.SetSystemPrompt("You are a helpful assistant with access to web, JSON, and date/time tools.")
		agent.AddTool(webFetch)
		agent.AddTool(jsonProcess)
		agent.AddTool(dateCalc)

		// Use the agent
		state := agentDomain.NewState()
		state.Set("prompt", "What day of the week will it be 30 days from now? Use the datetime_calculate tool.")
		resultState, err := agent.Run(ctx, state)
		if err != nil {
			log.Printf("Error running agent: %v", err)
		} else {
			if result, exists := resultState.Get("result"); exists {
				fmt.Printf("Agent response: %v\n", result)
			}
		}
	}

	// Show tool statistics
	fmt.Println()
	fmt.Println("=== Tool Statistics ===")
	toolsByCategory := make(map[string]int)
	for _, entry := range tools.Tools.List() {
		toolsByCategory[entry.Metadata.Category]++
	}

	fmt.Printf("Total tools available: %d\n", len(tools.Tools.List()))
	for category, count := range toolsByCategory {
		fmt.Printf("  %s: %d tools\n", category, count)
	}

	// Demonstrate migration from custom tools
	fmt.Println()
	fmt.Println("=== Migration from Custom Tools ===")
	fmt.Println()
	fmt.Println("Before (creating custom tools):")
	fmt.Println(`  tool := tools.NewTool("web_fetch", "Fetch web content", fetchFunc, schema)`)
	fmt.Println()
	fmt.Println("After (using built-in tools):")
	fmt.Println(`  tool, _ := tools.GetTool("web_fetch")`)
	fmt.Println()
	fmt.Println("Benefits of built-in tools:")
	fmt.Println("  - Standardized interfaces across all tools")
	fmt.Println("  - Enhanced features (timeouts, streaming, etc.)")
	fmt.Println("  - Automatic registration and discovery")
	fmt.Println("  - Consistent error handling")
	fmt.Println("  - Comprehensive documentation")
	fmt.Println("  - Regular updates and maintenance")
	fmt.Println()

	// Show how to find tools by tags
	fmt.Println("=== Finding Tools by Tags ===")
	fmt.Println()
	fmt.Println("Tools tagged with 'json':")
	for _, entry := range tools.Tools.List() {
		for _, tag := range entry.Metadata.Tags {
			if tag == "json" {
				fmt.Printf("  - %s\n", entry.Metadata.Name)
				break
			}
		}
	}
}
