// ABOUTME: Example demonstrating ultra-simple LLM agent creation with string-based provider specification
// ABOUTME: Shows the minimal code needed to create and run an AI agent with the new API

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

func main() {
	// Check environment variables
	fmt.Println("Environment check:")
	fmt.Printf("OPENAI_API_KEY set: %v\n", os.Getenv("OPENAI_API_KEY") != "")
	fmt.Printf("GO_LLMS_OPENAI_API_KEY set: %v\n", os.Getenv("GO_LLMS_OPENAI_API_KEY") != "")
	fmt.Printf("ANTHROPIC_API_KEY set: %v\n", os.Getenv("ANTHROPIC_API_KEY") != "")
	fmt.Printf("GO_LLMS_ANTHROPIC_API_KEY set: %v\n", os.Getenv("GO_LLMS_ANTHROPIC_API_KEY") != "")
	fmt.Printf("GEMINI_API_KEY set: %v\n", os.Getenv("GEMINI_API_KEY") != "")
	fmt.Printf("GO_LLMS_GEMINI_API_KEY set: %v\n", os.Getenv("GO_LLMS_GEMINI_API_KEY") != "")
	fmt.Println()

	// Example 1: Create agent with explicit provider/model
	// Using gpt-4o-mini which is faster and cheaper
	agent1, err := core.NewAgentFromString("assistant", "openai/gpt-4o-mini")
	if err != nil {
		// Will suggest setting OPENAI_API_KEY or GO_LLMS_OPENAI_API_KEY
		log.Printf("Failed to create OpenAI agent: %v", err)
		agent1 = nil
	}

	// Example 2: Create agent with alias (ultra-simple!)
	agent2, err := core.NewAgentFromString("claude-agent", "claude")
	if err != nil {
		log.Printf("Failed to create Claude agent: %v", err)
	} else {
		log.Printf("Created Claude agent: %s", agent2.Name())
	}

	// Example 3: Create agent with model inference
	agent3, err := core.NewAgentFromString("gemini-agent", "gemini-2.0-flash")
	if err != nil {
		log.Printf("Failed to create Gemini agent: %v", err)
	} else {
		log.Printf("Created Gemini agent: %s", agent3.Name())
	}

	// Example 4: Mock provider for testing (always works)
	mockAgent, err := core.NewAgentFromString("test-agent", "mock")
	if err != nil {
		log.Fatalf("Failed to create mock agent: %v", err)
	}

	// Run a simple task
	ctx := context.Background()

	// Using the new state-based interface with timeout
	if agent1 != nil {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		state := domain.NewState()
		state.Set("user_input", "What is 2+2?")

		resultState, err := agent1.Run(ctxWithTimeout, state)
		if err != nil {
			log.Printf("Agent1 error: %v", err)
		} else {
			if output, exists := resultState.Get("output"); exists {
				fmt.Printf("Agent1 response: %v\n", output)
			}
		}
	}

	// Using the new state-based interface
	state := domain.NewState()
	state.Set("user_input", "Tell me a short joke")

	resultState, err := mockAgent.Run(ctx, state)
	if err != nil {
		log.Fatalf("Mock agent error: %v", err)
	}

	if joke, exists := resultState.Get("output"); exists {
		fmt.Printf("Mock agent joke: %v\n", joke)
	}

	// Demonstrate the simplicity
	fmt.Println("\n--- Ultra-Simple Agent Creation Examples ---")
	fmt.Println("core.NewAgentFromString(\"my-agent\", \"gpt-4\")")
	fmt.Println("core.NewAgentFromString(\"my-agent\", \"claude\")")
	fmt.Println("core.NewAgentFromString(\"my-agent\", \"gemini\")")
	fmt.Println("core.NewAgentFromString(\"my-agent\", \"openai/gpt-4o-mini\")")
	fmt.Println("core.NewAgentFromString(\"my-agent\", \"anthropic/claude-3-opus-latest\")")
	fmt.Println("core.NewAgentFromString(\"my-agent\", \"gemini/gemini-2.0-flash\")")
}
