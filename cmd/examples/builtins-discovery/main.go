// ABOUTME: Example demonstrating both legacy and new discovery APIs for built-in tools
// ABOUTME: Shows metadata-first discovery for scripting engines and traditional import-based usage

package main

import (
	"context"
	"encoding/json"
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
	discoveryTools "github.com/lexlapax/go-llms/pkg/agent/tools"
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
	// Demonstrate NEW discovery API (v0.3.4+) - NO IMPORTS NEEDED!
	fmt.Println("=== NEW Discovery API (v0.3.4+) - Metadata-First Approach ===")
	fmt.Println("Perfect for scripting engines and dynamic environments!")
	fmt.Println()

	demonstrateNewDiscoveryAPI()

	// Display legacy registry summary
	fmt.Println("=== Legacy Registry API (with imports) ===")
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

// demonstrateNewDiscoveryAPI shows the new metadata-first discovery system
func demonstrateNewDiscoveryAPI() {
	// Get the discovery instance - NO IMPORTS NEEDED!
	discovery := discoveryTools.NewDiscovery()

	// 1. List all available tools without importing them
	fmt.Println("1. Listing all available tools (no imports required):")
	allTools := discovery.ListTools()
	fmt.Printf("   Found %d tools available\n", len(allTools))

	// Show first few tools as example
	fmt.Println("   First 5 tools:")
	for i, tool := range allTools {
		if i >= 5 {
			break
		}
		fmt.Printf("   - %s (%s): %s\n", tool.Name, tool.Category, tool.Description)
	}
	fmt.Println()

	// 2. Search tools by keyword
	fmt.Println("2. Searching tools by keyword:")
	jsonTools := discovery.SearchTools("json")
	fmt.Printf("   Tools related to 'json': %d\n", len(jsonTools))
	for _, tool := range jsonTools {
		fmt.Printf("   - %s: %s (tags: %v)\n", tool.Name, tool.Description, tool.Tags)
	}
	fmt.Println()

	// 3. Filter by category
	fmt.Println("3. Filtering tools by category:")
	mathTools := discovery.ListByCategory("math")
	fmt.Printf("   Math tools: %d\n", len(mathTools))
	for _, tool := range mathTools {
		fmt.Printf("   - %s: %s\n", tool.Name, tool.Description)
	}
	fmt.Println()

	// 4. Get detailed tool schema without creating the tool
	fmt.Println("4. Getting tool schema (no tool creation):")
	schema, err := discovery.GetToolSchema("calculator")
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Tool: %s\n", schema.Name)
		fmt.Printf("   Description: %s\n", schema.Description)
		if schema.Parameters != nil {
			params, _ := json.MarshalIndent(schema.Parameters, "   ", "  ")
			fmt.Printf("   Parameters:\n%s\n", params)
		}
	}
	fmt.Println()

	// 5. Get tool examples
	fmt.Println("5. Getting tool examples:")
	examples, err := discovery.GetToolExamples("calculator")
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Calculator examples: %d\n", len(examples))
		for i, ex := range examples {
			if i >= 2 { // Show only first 2 examples
				break
			}
			fmt.Printf("   - %s: %s\n", ex.Name, ex.Description)
			if ex.Input != nil {
				input, _ := json.Marshal(ex.Input)
				fmt.Printf("     Input: %s\n", input)
			}
			if ex.Output != nil {
				output, _ := json.Marshal(ex.Output)
				fmt.Printf("     Output: %s\n", output)
			}
		}
	}
	fmt.Println()

	// 6. Dynamic tool loading (lazy instantiation)
	fmt.Println("6. Dynamic tool loading:")
	tool, err := discovery.CreateTool("calculator")
	if err != nil {
		// This is expected without build tags
		fmt.Printf("   Tool not loaded (expected): %v\n", err)
		fmt.Println("   To load tools, compile with: go build -tags=tools")
	} else {
		fmt.Printf("   Successfully loaded tool: %s\n", tool.Name())
	}
	fmt.Println()

	// 7. Get comprehensive tool help
	fmt.Println("7. Getting tool help text:")
	help, err := discovery.GetToolHelp("datetime_now")
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Help for datetime_now:\n")
		// Show first few lines of help
		lines := help
		if len(lines) > 200 {
			lines = lines[:200] + "..."
		}
		fmt.Printf("%s\n", lines)
	}
	fmt.Println()

	// 8. Demonstrate scripting bridge use case
	fmt.Println("8. Scripting bridge example (how go-llmspell would use this):")
	metadata := discoveryTools.GetToolMetadata()
	fmt.Printf("   Total metadata entries: %d\n", len(metadata))

	// Simulate exposing to a scripting engine
	scriptData := make(map[string]interface{})
	for name, info := range metadata {
		toolData := map[string]interface{}{
			"name":        name,
			"description": info.Description,
			"category":    info.Category,
			"tags":        info.Tags,
		}

		// Parse schemas for script access
		if len(info.ParameterSchema) > 0 {
			var params interface{}
			if err := json.Unmarshal(info.ParameterSchema, &params); err == nil {
				toolData["parameters"] = params
			}
		}

		scriptData[name] = toolData
	}

	fmt.Printf("   Exposed %d tools to scripting engine\n", len(scriptData))
	fmt.Println("   Each tool includes: name, description, category, tags, parameters")
	fmt.Println()

	// 9. Show category statistics from discovery
	fmt.Println("9. Category statistics from discovery:")
	categoryStats := make(map[string]int)
	for _, tool := range allTools {
		categoryStats[tool.Category]++
	}

	for category, count := range categoryStats {
		fmt.Printf("   %s: %d tools\n", category, count)
	}
	fmt.Println()

	fmt.Println("=== Key Benefits of New Discovery API ===")
	fmt.Println("✓ No imports required - perfect for scripting engines")
	fmt.Println("✓ Metadata-first - explore tools before loading")
	fmt.Println("✓ Lazy loading - create tools only when needed")
	fmt.Println("✓ Search and filter - find tools by keyword, category, tags")
	fmt.Println("✓ Schema access - get parameter/output schemas without tool instances")
	fmt.Println("✓ Examples - see usage examples before tool creation")
	fmt.Println("✓ Help generation - get formatted help text")
	fmt.Println("✓ Bridge-friendly - designed for go-llmspell integration")
	fmt.Println()
}
