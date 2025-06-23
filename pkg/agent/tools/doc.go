// Package tools provides built-in tool implementations for LLM agents.
//
// This package contains a collection of ready-to-use tools that can be added
// to agents for common functionality such as file operations, web access,
// mathematical calculations, and system interactions.
//
// Tool Categories:
//
// File System Tools:
//   - FileReader: Read contents of files
//   - FileWriter: Write content to files
//   - DirectoryLister: List directory contents
//
// Web Tools:
//   - WebFetcher: Fetch content from URLs
//   - WebSearcher: Search the web using various search engines
//
// Math Tools:
//   - Calculator: Perform mathematical calculations
//   - StatisticalAnalyzer: Compute statistical measures
//
// System Tools:
//   - CommandRunner: Execute system commands
//   - EnvironmentReader: Read environment variables
//
// Each tool implements the domain.Tool interface and provides:
//   - Structured parameter schemas for validation
//   - Clear output schemas for predictable results
//   - Usage instructions for LLMs
//   - Error handling guidance
//   - Example usage patterns
//
// Example Usage:
//
//	agent := core.NewAgent("my-agent")
//	agent.AddTool(tools.NewCalculator())
//	agent.AddTool(tools.NewFileReader())
//
//	// The agent can now use these tools during execution
//	result, err := agent.Run(ctx, "Calculate the sum of numbers in data.txt")
//
// Custom Tools:
//
// To create custom tools, implement the domain.Tool interface:
//
//	type MyTool struct{}
//
//	func (t *MyTool) Name() string { return "my_tool" }
//	func (t *MyTool) Description() string { return "Does something useful" }
//	func (t *MyTool) Execute(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
//	    // Implementation
//	}
//	// ... other required methods
//
// Tool Discovery:
//
// Tools can be discovered dynamically using the tool registry:
//
//	registry := tools.GetGlobalToolRegistry()
//	availableTools := registry.ListTools()
//
//	// Register custom tools
//	registry.Register(myCustomTool)
package tools
