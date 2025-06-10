// ABOUTME: Demonstrates bidirectional agent-tool conversion utilities
// ABOUTME: Shows registry integration, event forwarding, and conversion patterns

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/tools"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

func main() {
	// Check command line arguments
	if len(os.Args) < 2 {
		fmt.Println("Usage: agent-tools-conversion <example>")
		fmt.Println("\nAvailable examples:")
		fmt.Println("  basic          - Basic agent-tool conversion")
		fmt.Println("  registry       - Registry integration")
		fmt.Println("  events         - Event forwarding from tools")
		fmt.Println("  schema         - Automatic schema mapping")
		fmt.Println("  chain          - Creating tool chains from agents")
		fmt.Println("  mapping        - Advanced parameter mapping")
		fmt.Println("  all            - Run all examples")
		os.Exit(1)
	}

	example := os.Args[1]

	switch example {
	case "basic":
		demonstrateBasicConversion()
	case "registry":
		demonstrateRegistryIntegration()
	case "events":
		demonstrateEventForwarding()
	case "schema":
		demonstrateSchemaMapping()
	case "chain":
		demonstrateToolChain()
	case "mapping":
		demonstrateAdvancedMapping()
	case "all":
		fmt.Println("=== Basic Conversion ===")
		demonstrateBasicConversion()
		fmt.Println("\n=== Registry Integration ===")
		demonstrateRegistryIntegration()
		fmt.Println("\n=== Event Forwarding ===")
		demonstrateEventForwarding()
		fmt.Println("\n=== Schema Mapping ===")
		demonstrateSchemaMapping()
		fmt.Println("\n=== Tool Chain ===")
		demonstrateToolChain()
		fmt.Println("\n=== Advanced Mapping ===")
		demonstrateAdvancedMapping()
	default:
		fmt.Printf("Unknown example: %s\n", example)
		os.Exit(1)
	}
}

// demonstrateBasicConversion shows basic agent-tool conversion
func demonstrateBasicConversion() {
	fmt.Println("Demonstrating basic agent-tool conversion...")

	// Create a simple calculator agent
	calculatorAgent := &customAgent{
		BaseAgentImpl: core.NewBaseAgent("calculator", "Performs calculations", domain.AgentTypeCustom),
		runFunc: func(ctx context.Context, state *domain.State) (*domain.State, error) {
			op, _ := state.Get("operation")
			a, _ := state.Get("a")
			b, _ := state.Get("b")

			var result float64
			switch op {
			case "add":
				result = toFloat(a) + toFloat(b)
			case "multiply":
				result = toFloat(a) * toFloat(b)
			default:
				result = 0
			}

			newState := state.Clone()
			newState.Set("result", result)
			return newState, nil
		},
	}

	// Convert agent to tool
	agentTool := tools.NewAgentTool(calculatorAgent)
	fmt.Printf("Created tool: %s - %s\n", agentTool.Name(), agentTool.Description())

	// Use the tool
	agentInfo := domain.AgentInfo{
		ID:          "demo-agent",
		Name:        "Demo Agent",
		Description: "Agent for demonstration",
		Type:        domain.AgentTypeCustom,
	}
	state := domain.NewState()
	ctx := &domain.ToolContext{
		Context: context.Background(),
		State:   domain.NewStateReader(state),
		Agent:   agentInfo,
		RunID:   "example-run",
	}

	result, err := agentTool.Execute(ctx, map[string]interface{}{
		"operation": "add",
		"a":         5,
		"b":         3,
	})
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Tool execution result: %v\n", result)
}

// demonstrateRegistryIntegration shows registry integration
func demonstrateRegistryIntegration() {
	fmt.Println("Demonstrating registry integration...")

	// Create a registry
	registry := builtins.NewRegistry[domain.Tool]()

	// Create some agents
	agents := []domain.BaseAgent{
		core.NewBaseAgent("summarizer", "Summarizes text", domain.AgentTypeLLM),
		core.NewBaseAgent("translator", "Translates text", domain.AgentTypeLLM),
		core.NewBaseAgent("analyzer", "Analyzes text", domain.AgentTypeLLM),
	}

	// Register agents as tools with a prefix
	err := tools.RegisterAgentsAsTools(agents, registry, tools.ConversionOptions{
		NamePrefix:          "agent_",
		AutoGenerateMappers: true,
	})
	if err != nil {
		log.Printf("Registration error: %v", err)
		return
	}

	// List all registered tools
	fmt.Println("Registered tools:")
	entries := registry.List()
	for _, entry := range entries {
		fmt.Printf("  - %s: %s\n", entry.Metadata.Name, entry.Metadata.Description)
	}

	// Search for specific tools
	fmt.Println("\nSearching for 'trans' tools:")
	results := registry.Search("trans")
	for _, entry := range results {
		fmt.Printf("  - Found: %s\n", entry.Metadata.Name)
	}
}

// demonstrateEventForwarding shows event forwarding
func demonstrateEventForwarding() {
	fmt.Println("Demonstrating event forwarding from tools...")

	// Create an event dispatcher
	dispatcher := core.NewEventDispatcher(100)

	// Subscribe to events
	eventCount := 0
	dispatcher.Subscribe(domain.EventHandlerFunc(func(event domain.Event) error {
		eventCount++
		fmt.Printf("Event %d: [%s] %s", eventCount, event.Type, event.AgentName)
		if event.Type == domain.EventMessage {
			fmt.Printf(" - %v", event.Data)
		}
		fmt.Println()
		return nil
	}))

	// Create a tool that emits events
	eventTool := &eventEmittingTool{
		name:        "event-demo",
		description: "Demonstrates event emission",
	}

	// Wrap as agent with event support
	toolAgent := tools.NewToolAgentWithEvents(eventTool, dispatcher)

	// Run the agent
	state := domain.NewState()
	state.Set("message", "Hello from event-forwarding example!")

	fmt.Println("Running agent (watch for events)...")
	ctx := context.Background()
	result, err := toolAgent.Run(ctx, state)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	// Give time for events to process
	time.Sleep(100 * time.Millisecond)

	if status, ok := result.Get("status"); ok {
		fmt.Printf("Final status: %v\n", status)
	}
}

// demonstrateSchemaMapping shows automatic schema mapping
func demonstrateSchemaMapping() {
	fmt.Println("Demonstrating automatic schema mapping...")

	// Create an agent with schemas
	agent := &schemaAgent{
		BaseAgentImpl: core.NewBaseAgent("greeter", "Greets users", domain.AgentTypeCustom),
		inputSchema: &sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				"name":     {Type: "string", Description: "User's name"},
				"language": {Type: "string", Description: "Greeting language (default: english)"},
			},
			Required: []string{"name"},
		},
		outputSchema: &sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				"greeting": {Type: "string", Description: "The greeting message"},
				"formal":   {Type: "boolean", Description: "Whether greeting is formal"},
			},
			Required: []string{"greeting"},
		},
	}

	// Convert to tool with automatic mapper generation
	agentTool := tools.NewAgentTool(agent)

	// Derive parameter schema
	paramSchema := tools.DeriveToolSchemaFromAgent(agent)
	fmt.Printf("Derived parameter schema:\n")
	fmt.Printf("  Type: %s\n", paramSchema.Type)
	fmt.Printf("  Required fields: %v\n", paramSchema.Required)
	fmt.Printf("  Properties:\n")
	for name, prop := range paramSchema.Properties {
		fmt.Printf("    - %s: %s\n", name, prop.Description)
	}

	// Use the tool
	agentInfo := domain.AgentInfo{
		ID:          "schema-test",
		Name:        "Schema Test Agent",
		Description: "Testing schema mapping",
		Type:        domain.AgentTypeCustom,
	}
	state := domain.NewState()
	ctx := &domain.ToolContext{
		Context: context.Background(),
		State:   domain.NewStateReader(state),
		Agent:   agentInfo,
		RunID:   "schema-test",
	}
	result, err := agentTool.Execute(ctx, map[string]interface{}{
		"name":     "Alice",
		"language": "spanish",
	})
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("\nExecution result: %v\n", result)
}

// demonstrateToolChain shows creating tool chains
func demonstrateToolChain() {
	fmt.Println("Demonstrating tool chains from agents...")

	// Create a chain of transformation agents
	agents := []domain.BaseAgent{
		&transformAgent{
			BaseAgentImpl: core.NewBaseAgent("uppercase", "Converts to uppercase", domain.AgentTypeCustom),
			transform: func(input string) string {
				return fmt.Sprintf("[UPPER] %s", input)
			},
		},
		&transformAgent{
			BaseAgentImpl: core.NewBaseAgent("reverse", "Reverses text", domain.AgentTypeCustom),
			transform: func(input string) string {
				runes := []rune(input)
				for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
					runes[i], runes[j] = runes[j], runes[i]
				}
				return string(runes)
			},
		},
		&transformAgent{
			BaseAgentImpl: core.NewBaseAgent("annotate", "Adds annotation", domain.AgentTypeCustom),
			transform: func(input string) string {
				return fmt.Sprintf("%s [PROCESSED]", input)
			},
		},
	}

	// Create a tool that chains the agents
	chainTool := tools.CreateToolChainFromAgents(agents...)

	// Use the chain tool
	agentInfo := domain.AgentInfo{
		ID:          "chain-test",
		Name:        "Chain Test Agent",
		Description: "Testing tool chains",
		Type:        domain.AgentTypeCustom,
	}
	state := domain.NewState()
	ctx := &domain.ToolContext{
		Context: context.Background(),
		State:   domain.NewStateReader(state),
		Agent:   agentInfo,
		RunID:   "chain-test",
	}
	result, err := chainTool.Execute(ctx, map[string]interface{}{
		"input": "Hello World",
	})
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Chain processing result: %v\n", result)
}

// demonstrateAdvancedMapping shows advanced parameter mapping
func demonstrateAdvancedMapping() {
	fmt.Println("Demonstrating advanced parameter mapping...")

	// Create path-based mapper for nested data
	pathMapper := tools.CreatePathMapper(map[string]string{
		"userName":    "user.profile.name",
		"userEmail":   "user.profile.email",
		"companyName": "user.company.name",
		"companySize": "user.company.employees",
	})

	// Create type conversion mapper
	typeMapper := tools.CreateTypeConversionMapper(map[string]func(interface{}) interface{}{
		"age": func(v interface{}) interface{} {
			// Convert string to int
			switch val := v.(type) {
			case string:
				// Simple conversion for demo
				return len(val) * 3 // Dummy logic
			case float64:
				return int(val)
			default:
				return v
			}
		},
		"active": func(v interface{}) interface{} {
			// Convert to boolean
			switch val := v.(type) {
			case string:
				return val == "true" || val == "yes" || val == "1"
			case int:
				return val > 0
			default:
				return false
			}
		},
	})

	// Test path mapper with nested data
	fmt.Println("Path-based extraction:")
	nestedData := map[string]interface{}{
		"user": map[string]interface{}{
			"profile": map[string]interface{}{
				"name":  "Bob Smith",
				"email": "bob@example.com",
			},
			"company": map[string]interface{}{
				"name":      "TechCorp",
				"employees": 500,
			},
		},
	}

	state, err := pathMapper(context.Background(), nestedData)
	if err != nil {
		log.Printf("Path mapper error: %v", err)
		return
	}

	fmt.Printf("  Extracted user name: %v\n", mustGet(state, "userName"))
	fmt.Printf("  Extracted email: %v\n", mustGet(state, "userEmail"))
	fmt.Printf("  Extracted company: %v\n", mustGet(state, "companyName"))
	fmt.Printf("  Extracted size: %v\n", mustGet(state, "companySize"))

	// Test type conversion
	fmt.Println("\nType conversion:")
	dataWithTypes := map[string]interface{}{
		"age":    "twenty-five",
		"active": "yes",
		"score":  95.5,
	}

	convertedState, err := typeMapper(context.Background(), dataWithTypes)
	if err != nil {
		log.Printf("Type mapper error: %v", err)
		return
	}

	age, _ := convertedState.Get("age")
	active, _ := convertedState.Get("active")
	score, _ := convertedState.Get("score")

	fmt.Printf("  Age: %v (type: %T)\n", age, age)
	fmt.Printf("  Active: %v (type: %T)\n", active, active)
	fmt.Printf("  Score: %v (type: %T)\n", score, score)

	// Test nested state mapper
	fmt.Println("\nNested state flattening:")
	flattenMapper := tools.CreateNestedStateMapper(true)

	flatState, err := flattenMapper(context.Background(), nestedData)
	if err != nil {
		log.Printf("Flatten mapper error: %v", err)
		return
	}

	fmt.Println("  Flattened keys:")
	values := flatState.Values()
	for key := range values {
		fmt.Printf("    - %s = %v\n", key, values[key])
	}
}

// Helper types and functions

type customAgent struct {
	*core.BaseAgentImpl
	runFunc func(context.Context, *domain.State) (*domain.State, error)
}

func (a *customAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	if a.runFunc != nil {
		return a.runFunc(ctx, state)
	}
	return state, nil
}

type eventEmittingTool struct {
	name        string
	description string
}

func (t *eventEmittingTool) Name() string                     { return t.name }
func (t *eventEmittingTool) Description() string              { return t.description }
func (t *eventEmittingTool) ParameterSchema() *sdomain.Schema { return nil }
func (t *eventEmittingTool) Category() string                 { return "demo" }
func (t *eventEmittingTool) Version() string                  { return "1.0.0" }
func (t *eventEmittingTool) OutputSchema() *sdomain.Schema    { return nil }
func (t *eventEmittingTool) UsageInstructions() string        { return "Demo tool for event emission" }
func (t *eventEmittingTool) Examples() []domain.ToolExample   { return nil }
func (t *eventEmittingTool) Constraints() []string            { return nil }
func (t *eventEmittingTool) ErrorGuidance() map[string]string { return nil }
func (t *eventEmittingTool) BehavioralHints() []string        { return nil }
func (t *eventEmittingTool) IsDeterministic() bool            { return true }
func (t *eventEmittingTool) IsDestructive() bool              { return false }
func (t *eventEmittingTool) RequiresConfirmation() bool       { return false }
func (t *eventEmittingTool) EstimatedLatency() string         { return "fast" }
func (t *eventEmittingTool) Tags() []string                   { return []string{"demo", "events"} }
func (t *eventEmittingTool) ToMCPDefinition() domain.MCPToolDefinition {
	return domain.MCPToolDefinition{
		Name:        t.name,
		Description: t.description,
		InputSchema: nil,
	}
}

func (t *eventEmittingTool) Execute(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
	// Emit events if available
	if ctx.Events != nil {
		ctx.Events.EmitMessage("Starting tool execution")

		if msg, ok := params.(map[string]interface{})["message"]; ok {
			ctx.Events.EmitMessage(fmt.Sprintf("Processing message: %v", msg))
		}

		ctx.Events.EmitProgress(50, 100, "Halfway done")
		ctx.Events.EmitMessage("Completing execution")
		ctx.Events.EmitProgress(100, 100, "Done")
	}

	return map[string]interface{}{
		"status":    "completed",
		"input":     params,
		"timestamp": time.Now().Format(time.RFC3339),
	}, nil
}

type schemaAgent struct {
	*core.BaseAgentImpl
	inputSchema  *sdomain.Schema
	outputSchema *sdomain.Schema
}

func (a *schemaAgent) InputSchema() *sdomain.Schema  { return a.inputSchema }
func (a *schemaAgent) OutputSchema() *sdomain.Schema { return a.outputSchema }

func (a *schemaAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	name, _ := state.Get("name")
	language, _ := state.Get("language")

	var greeting string
	var formal bool

	switch language {
	case "spanish":
		greeting = fmt.Sprintf("¡Hola, %v!", name)
		formal = false
	case "french":
		greeting = fmt.Sprintf("Bonjour, %v!", name)
		formal = true
	case "japanese":
		greeting = fmt.Sprintf("こんにちは、%vさん!", name)
		formal = true
	default:
		greeting = fmt.Sprintf("Hello, %v!", name)
		formal = false
	}

	result := domain.NewState()
	result.Set("greeting", greeting)
	result.Set("formal", formal)

	return result, nil
}

type transformAgent struct {
	*core.BaseAgentImpl
	transform func(string) string
}

func (a *transformAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	input, _ := state.Get("input")
	if str, ok := input.(string); ok && a.transform != nil {
		state.Set("input", a.transform(str))
	}
	return state, nil
}

// Helper functions

func toFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case string:
		// Simple conversion for demo
		return float64(len(val))
	default:
		return 0
	}
}

func mustGet(state *domain.State, key string) interface{} {
	val, _ := state.Get(key)
	return val
}
