package llmutil

import (
	"context"
	"fmt"
	"time"

	agentDomain "github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/tools"
	"github.com/lexlapax/go-llms/pkg/agent/workflow"
	llmDomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// AgentConfig represents a configuration for an agent
type AgentConfig struct {
	Provider      llmDomain.Provider
	SystemPrompt  string
	ModelName     string
	EnableCaching bool
	Tools         []agentDomain.Tool
	Hooks         []agentDomain.Hook
}

// CreateAgent creates a new agent with the specified configuration
func CreateAgent(config AgentConfig) agentDomain.Agent {
	// Create the base agent
	var agent agentDomain.Agent
	
	if config.EnableCaching {
		agent = workflow.NewCachedAgent(config.Provider)
	} else {
		agent = workflow.NewAgent(config.Provider)
	}
	
	// Set system prompt if provided
	if config.SystemPrompt != "" {
		agent.SetSystemPrompt(config.SystemPrompt)
	}
	
	// Set model if provided
	if config.ModelName != "" {
		agent.WithModel(config.ModelName)
	}
	
	// Add tools
	for _, tool := range config.Tools {
		agent.AddTool(tool)
	}
	
	// Add hooks
	for _, hook := range config.Hooks {
		agent.WithHook(hook)
	}
	
	return agent
}

// CreateStandardTools creates a set of standard tools that are commonly used
func CreateStandardTools() []agentDomain.Tool {
	standardTools := []agentDomain.Tool{
		// Current date/time tool
		tools.NewTool(
			"get_current_date",
			"Get the current date and time information",
			func() map[string]string {
				now := time.Now()
				return map[string]string{
					"date":              now.Format("2006-01-02"),
					"time":              now.Format("15:04:05"),
					"year":              fmt.Sprintf("%d", now.Year()),
					"month":             now.Month().String(),
					"day":               fmt.Sprintf("%d", now.Day()),
					"weekday":           now.Weekday().String(),
					"timezone":          now.Location().String(),
					"unix_epoch":        fmt.Sprintf("%d", now.Unix()),
				}
			},
			&schemaDomain.Schema{
				Type:        "object",
				Description: "Returns the current date and time information",
			},
		),
		
		// Simple calculator tool
		tools.NewTool(
			"calculator",
			"Perform mathematical calculations",
			func(params struct {
				Expression string `json:"expression"`
			}) (map[string]interface{}, error) {
				// Implementation omitted for brevity - would call a calculator function
				return map[string]interface{}{
					"success":    true,
					"expression": params.Expression,
					"result":     0, // placeholder
				}, nil
			},
			&schemaDomain.Schema{
				Type: "object",
				Properties: map[string]schemaDomain.Property{
					"expression": {
						Type:        "string",
						Description: "The mathematical expression to evaluate",
					},
				},
				Required: []string{"expression"},
			},
		),
	}
	
	return standardTools
}

// AgentWithMetrics creates an agent with standard metrics monitoring
func AgentWithMetrics(provider llmDomain.Provider, systemPrompt string) agentDomain.Agent {
	// Create a cached agent for better performance
	agent := workflow.NewCachedAgent(provider)
	
	// Set the system prompt
	agent.SetSystemPrompt(systemPrompt)
	
	// Add a metrics hook
	metricsHook := workflow.NewMetricsHook()
	agent.WithHook(metricsHook)
	
	return agent
}

// RunWithTimeout runs an agent with a timeout
func RunWithTimeout(
	agent agentDomain.Agent,
	prompt string,
	timeout time.Duration,
) (interface{}, error) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	// Run the agent with the timeout context
	return agent.Run(ctx, prompt)
}

// RunWithSchema runs an agent with a schema and timeout
func RunWithSchema[T any](
	agent agentDomain.Agent,
	prompt string,
	timeout time.Duration,
) (T, error) {
	var result T
	
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	// This is a placeholder implementation
	// In a real implementation, we would generate schema from the type
	// and convert the result properly
	
	// Run the agent with a simple schema
	schema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{},
	}
	
	_, err := agent.RunWithSchema(ctx, prompt, schema)
	if err != nil {
		return result, err
	}
	
	// Placeholder implementation
	return result, fmt.Errorf("schema generation and conversion not implemented")
}