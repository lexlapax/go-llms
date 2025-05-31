// ABOUTME: Example demonstrating usage of built-in components including tools, agents, and workflows
// ABOUTME: Shows how to discover, configure, and use pre-built components from the registry

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	// Import built-in components - this triggers auto-registration
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web" // Import for side effects (registration)
	"github.com/lexlapax/go-llms/pkg/agent/workflow"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
)

func main() {
	// Display registry summary
	fmt.Println("=== Built-in Components Registry ===")
	fmt.Println()

	// List all registered tools
	fmt.Println("Available Tools:")
	for _, entry := range tools.Tools.List() {
		fmt.Printf("  - %s (%s): %s\n", 
			entry.Metadata.Name,
			entry.Metadata.Category,
			entry.Metadata.Description)
		fmt.Printf("    Tags: %v\n", entry.Metadata.Tags)
		fmt.Printf("    Version: %s\n", entry.Metadata.Version)
		fmt.Println()
	}

	// Search for specific tools
	fmt.Println("Searching for 'web' tools:")
	webTools := tools.Tools.Search("web")
	for _, entry := range webTools {
		fmt.Printf("  - %s: %s\n", entry.Metadata.Name, entry.Metadata.Description)
	}
	fmt.Println()

	// List tools by category
	fmt.Println("Tools in 'web' category:")
	categoryTools := tools.Tools.ListByCategory("web")
	for _, entry := range categoryTools {
		fmt.Printf("  - %s\n", entry.Metadata.Name)
	}
	fmt.Println()

	// Demonstrate tool usage
	fmt.Println("=== Using Built-in Tools ===")
	fmt.Println()

	// Get the web_fetch tool
	webFetch, found := tools.GetTool("web_fetch")
	if !found {
		log.Fatal("web_fetch tool not found")
	}

	// Create a simple agent with the built-in tool
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Println("OPENAI_API_KEY not set, skipping agent demonstration")
		return
	}

	// Create provider
	p := provider.NewOpenAIProvider(apiKey, "gpt-4o-mini")

	// Create agent with built-in tool
	agent := workflow.NewAgent(p).
		SetSystemPrompt("You are a helpful assistant that can fetch web content.").
		AddTool(webFetch)

	// Use the agent
	ctx := context.Background()
	result, err := agent.Run(ctx, "What is the content of https://example.com? Use the web_fetch tool to get it.")
	if err != nil {
		log.Printf("Error running agent: %v", err)
		return
	}

	fmt.Printf("Agent response: %v\n", result)

	// Show tool examples
	fmt.Println()
	fmt.Println("=== Tool Examples ===")
	toolEntries := tools.Tools.Search("web_fetch")
	if len(toolEntries) > 0 {
		for _, example := range toolEntries[0].Metadata.Examples {
			fmt.Printf("\nExample: %s\n", example.Name)
			fmt.Printf("Description: %s\n", example.Description)
			fmt.Printf("Code:\n%s\n", example.Code)
		}
	}
}