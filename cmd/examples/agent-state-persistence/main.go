// ABOUTME: Example demonstrating state persistence and serialization for agents
// ABOUTME: Shows how to save and restore agent state for multi-session interactions

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// PersistentState wraps domain.State with persistence methods
type PersistentState struct {
	*domain.State
	filename string
}

// NewPersistentState creates a new state that can be saved/loaded
func NewPersistentState(filename string) *PersistentState {
	return &PersistentState{
		State:    domain.NewState(),
		filename: filename,
	}
}

// Save persists the state to a file
func (ps *PersistentState) Save() error {
	// Get all state data
	data := ps.Values()

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write to file
	if err := os.WriteFile(ps.filename, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// Load restores state from a file
func (ps *PersistentState) Load() error {
	// Read file
	jsonData, err := os.ReadFile(ps.filename)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet, that's OK
			return nil
		}
		return fmt.Errorf("failed to read state file: %w", err)
	}

	// Unmarshal JSON
	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return fmt.Errorf("failed to unmarshal state: %w", err)
	}

	// Set all values
	for key, value := range data {
		ps.Set(key, value)
	}

	return nil
}

func main() {
	fmt.Println("=== State Persistence Example ===")

	// Create a mock agent for demonstration
	agent, err := core.NewAgentFromString("persistent-agent", "mock")
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Create persistent state
	state := NewPersistentState("agent_state.json")

	// Load any existing state
	if err := state.Load(); err != nil {
		log.Printf("Warning: Could not load state: %v", err)
	}

	// Check if we have conversation history
	var conversationHistory []map[string]string
	if history, exists := state.Get("conversation_history"); exists {
		if h, ok := history.([]interface{}); ok {
			conversationHistory = make([]map[string]string, 0, len(h))
			for _, item := range h {
				if m, ok := item.(map[string]interface{}); ok {
					entry := make(map[string]string)
					for k, v := range m {
						if s, ok := v.(string); ok {
							entry[k] = s
						}
					}
					conversationHistory = append(conversationHistory, entry)
				}
			}
		}
	} else {
		conversationHistory = []map[string]string{}
	}

	// Display conversation history
	if len(conversationHistory) > 0 {
		fmt.Println("\n--- Previous Conversation ---")
		for _, entry := range conversationHistory {
			fmt.Printf("User: %s\n", entry["user"])
			fmt.Printf("Agent: %s\n\n", entry["agent"])
		}
		fmt.Println("--- Continuing Conversation ---")
	} else {
		fmt.Println("\n--- Starting New Conversation ---")
	}

	// Get user input
	fmt.Print("\nEnter your message (or 'quit' to exit): ")
	var userInput string
	_, _ = fmt.Scanln(&userInput)

	if userInput == "quit" {
		fmt.Println("Goodbye!")
		return
	}

	// Set current input
	state.Set("user_input", userInput)

	// Run agent
	ctx := context.Background()
	resultState, err := agent.Run(ctx, state.State)
	if err != nil {
		log.Fatalf("Agent error: %v", err)
	}

	// Get response
	output, _ := resultState.Get("output")
	response := fmt.Sprintf("%v", output)
	fmt.Printf("\nAgent: %s\n", response)

	// Update conversation history
	conversationHistory = append(conversationHistory, map[string]string{
		"user":  userInput,
		"agent": response,
	})

	// Store updated history in state
	state.Set("conversation_history", conversationHistory)

	// Get session count
	sessionCount := 0
	if count, exists := state.Get("session_count"); exists {
		if c, ok := count.(float64); ok {
			sessionCount = int(c)
		}
	}
	sessionCount++
	state.Set("session_count", sessionCount)

	// Save state
	if err := state.Save(); err != nil {
		log.Printf("Warning: Could not save state: %v", err)
	}

	fmt.Printf("\n--- Session #%d saved ---\n", sessionCount)
	fmt.Println("Run the program again to continue the conversation!")
}
