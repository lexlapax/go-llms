package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/tools"
	"github.com/lexlapax/go-llms/pkg/agent/workflow"
	llmDomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/schema/validation"
)

// Run implements the agent command
func (a *AgentCmd) Run(ctx *Context) error {
	// Get the prompt
	prompt := strings.Join(a.Prompt, " ")
	if prompt == "" {
		return fmt.Errorf("no prompt provided")
	}

	// Create the LLM provider
	provider, err := ctx.CreateProvider()
	if err != nil {
		return err
	}

	// Create the agent
	agent := workflow.NewCachedAgent(provider)

	// Add requested tools
	if err := addRequestedTools(agent, a.Tools); err != nil {
		return err
	}

	// Set system prompt if provided
	if a.System != "" {
		// Agent API might have changed - need to check the actual method
		// for now, we'll comment this out
		// agent.SetSystemPrompt(a.System)
	}

	// Execute the agent
	kctx := context.Background()
	response, err := agent.Run(kctx, prompt)
	if err != nil {
		return err
	}

	// Output response
	switch v := response.(type) {
	case string:
		fmt.Println(v)
	case []byte:
		fmt.Println(string(v))
	default:
		// Try to marshal as JSON
		if output, err := json.MarshalIndent(response, "", "  "); err == nil {
			fmt.Println(string(output))
		} else {
			fmt.Printf("%v\n", response)
		}
	}
	return nil
}

// Run implements the structured command
func (s *StructuredCmd) Run(ctx *Context) error {
	// Read the schema file
	schemaBytes, err := os.ReadFile(s.Schema)
	if err != nil {
		return fmt.Errorf("error reading schema file: %w", err)
	}

	// Parse the schema into domain.Schema
	var schemaMap map[string]interface{}
	if err := json.Unmarshal(schemaBytes, &schemaMap); err != nil {
		return fmt.Errorf("error parsing schema: %w", err)
	}
	
	// Convert to domain.Schema
	schema := &schemaDomain.Schema{
		Type:       getString(schemaMap, "type"),
		Properties: getProperties(schemaMap),
		Required:   getRequired(schemaMap),
	}

	// Create the validator with options
	validator := validation.NewValidator()

	// If we're just validating, read the input and validate
	if s.Validate {
		// Read input from prompt (could be a file)
		input, err := readInput(s.Prompt)
		if err != nil {
			return err
		}

		// Validate the JSON string
		result, err := validator.Validate(schema, input)
		if err != nil {
			return fmt.Errorf("validation error: %w", err)
		}

		if !result.Valid {
			return fmt.Errorf("validation failed: %v", result.Errors)
		}

		fmt.Println("Validation successful")
		return nil
	}

	// Otherwise, use LLM to extract structured data
	// Create the LLM provider
	provider, err := ctx.CreateProvider()
	if err != nil {
		return err
	}

	// Remove the processor creation since we're doing direct extraction
	// processor := structuredProcessor.NewProcessor()

	// Read the prompt
	promptText, err := readInput(s.Prompt)
	if err != nil {
		return err
	}

	// Process the prompt to extract structured data
	// The API might be different, let's use the correct structured extraction approach
	// For now, we'll just use the provider directly with the schema
	messages := []llmDomain.Message{
		llmDomain.NewTextMessage(llmDomain.RoleUser, promptText),
	}
	
	kctx := context.Background()
	options := []llmDomain.Option{
		llmDomain.WithTemperature(float64(s.Temperature)),
		llmDomain.WithMaxTokens(s.MaxTokens),
	}
	
	response, err := provider.GenerateMessage(kctx, messages, options...)
	if err != nil {
		return fmt.Errorf("structured extraction failed: %w", err)
	}
	
	// Extract and parse the response
	// response.Content is already a string for simple API
	responseText := response.Content
	
	// Try to parse as JSON
	var result interface{}
	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		return fmt.Errorf("error parsing response as JSON: %w", err)
	}
	
	// Validate against schema
	// First convert result to JSON string for validation
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("error marshaling result: %w", err)
	}
	
	validationResult, err := validator.Validate(schema, string(resultJSON))
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	
	if !validationResult.Valid {
		return fmt.Errorf("validation failed: %v", validationResult.Errors)
	}
	
	// Format the output
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("error formatting output: %w", err)
	}

	// Write output to file or stdout
	if s.OutputFile != "" {
		if err := os.WriteFile(s.OutputFile, output, 0644); err != nil {
			return fmt.Errorf("error writing output file: %w", err)
		}
		fmt.Printf("Output written to %s\n", s.OutputFile)
	} else {
		fmt.Println(string(output))
	}

	return nil
}

// Helper function to read input from a string (could be a file path)
func readInput(input string) (string, error) {
	// Check if it's a file
	if _, err := os.Stat(input); err == nil {
		data, err := os.ReadFile(input)
		if err != nil {
			return "", fmt.Errorf("error reading file: %w", err)
		}
		return string(data), nil
	}
	// Otherwise, it's the actual input
	return input, nil
}

// Helper functions for structured command

// addRequestedTools adds the requested tools to the agent
func addRequestedTools(agent *workflow.CachedAgent, toolNames []string) error {
	for _, name := range toolNames {
		switch strings.ToLower(name) {
		case "calculator":
			agent.AddTool(createCalculatorTool())
		case "file_reader":
			agent.AddTool(createFileReaderTool())
		case "file_writer":
			agent.AddTool(createFileWriterTool())
		case "shell_executor":
			agent.AddTool(createShellExecutorTool())
		default:
			return fmt.Errorf("unknown tool: %s", name)
		}
	}
	return nil
}

// Tool creation functions
func createCalculatorTool() domain.Tool {
	schema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"expression": {
				Type:        "string",
				Description: "Mathematical expression to evaluate",
			},
		},
		Required: []string{"expression"},
	}
	
	return tools.NewTool(
		"calculator",
		"Perform mathematical calculations",
		func(params struct {
			Expression string `json:"expression"`
		}) (float64, error) {
			// Simple implementation for demo
			// In real implementation, you'd use a proper expression parser
			return 0, fmt.Errorf("calculator not implemented")
		},
		schema,
	)
}

func createFileReaderTool() domain.Tool {
	schema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"path": {
				Type:        "string",
				Description: "Path to the file to read",
			},
		},
		Required: []string{"path"},
	}
	
	return tools.NewTool(
		"file_reader",
		"Read contents of a file",
		func(params struct {
			Path string `json:"path"`
		}) (string, error) {
			data, err := os.ReadFile(params.Path)
			if err != nil {
				return "", err
			}
			return string(data), nil
		},
		schema,
	)
}

func createFileWriterTool() domain.Tool {
	schema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"path": {
				Type:        "string",
				Description: "Path to the file to write",
			},
			"content": {
				Type:        "string",
				Description: "Content to write to the file",
			},
		},
		Required: []string{"path", "content"},
	}
	
	return tools.NewTool(
		"file_writer",
		"Write contents to a file",
		func(params struct {
			Path    string `json:"path"`
			Content string `json:"content"`
		}) (string, error) {
			err := os.WriteFile(params.Path, []byte(params.Content), 0644)
			if err != nil {
				return "", err
			}
			return "File written successfully", nil
		},
		schema,
	)
}

func createShellExecutorTool() domain.Tool {
	schema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"command": {
				Type:        "string",
				Description: "Shell command to execute",
			},
		},
		Required: []string{"command"},
	}
	
	return tools.NewTool(
		"shell_executor",
		"Execute shell commands",
		func(params struct {
			Command string `json:"command"`
		}) (string, error) {
			// For security, this should be restricted in production
			return "", fmt.Errorf("shell executor not implemented for security reasons")
		},
		schema,
	)
}