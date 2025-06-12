package main

// ABOUTME: Example of using the Google Vertex AI provider
// ABOUTME: Demonstrates authentication, model usage, and streaming with Vertex AI

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
)

func main() {
	// Get configuration from environment
	projectID := os.Getenv("VERTEX_AI_PROJECT_ID")
	if projectID == "" {
		log.Fatal("Please set VERTEX_AI_PROJECT_ID environment variable")
	}

	location := os.Getenv("VERTEX_AI_LOCATION")
	if location == "" {
		location = "us-central1" // Default location
		fmt.Printf("VERTEX_AI_LOCATION not set, using default: %s\n", location)
	}

	model := os.Getenv("VERTEX_AI_MODEL")
	if model == "" {
		model = "gemini-1.5-flash-001" // Default model
		fmt.Printf("VERTEX_AI_MODEL not set, using default: %s\n", model)
	}

	// Example 1: Basic usage with ADC (Application Default Credentials)
	fmt.Println("=== Example 1: Basic Generation with ADC ===")
	if err := basicGeneration(projectID, location, model); err != nil {
		log.Printf("Basic generation error: %v", err)
	}

	// Example 2: Using service account
	fmt.Println("\n=== Example 2: Generation with Service Account ===")
	if err := serviceAccountGeneration(projectID, location, model); err != nil {
		log.Printf("Service account generation error: %v", err)
	}

	// Example 3: Streaming response
	fmt.Println("\n=== Example 3: Streaming Generation ===")
	if err := streamingGeneration(projectID, location, model); err != nil {
		log.Printf("Streaming generation error: %v", err)
	}

	// Example 4: Using partner models (Claude)
	fmt.Println("\n=== Example 4: Partner Models (Claude) ===")
	if err := partnerModelGeneration(projectID, location); err != nil {
		log.Printf("Partner model error: %v", err)
	}

	// Example 5: Multimodal (image) input
	fmt.Println("\n=== Example 5: Multimodal Input ===")
	if err := multimodalGeneration(projectID, location, model); err != nil {
		log.Printf("Multimodal generation error: %v", err)
	}
}

func basicGeneration(projectID, location, model string) error {
	// Create provider with ADC
	llm, err := provider.NewVertexAIProvider(projectID, location, model)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	// Create messages
	messages := []domain.Message{
		domain.NewTextMessage(domain.RoleUser, "What are the benefits of using Vertex AI for enterprise applications? List 3 key points."),
	}

	// Generate response
	ctx := context.Background()
	response, err := llm.GenerateMessage(ctx, messages, domain.WithMaxTokens(200))
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	fmt.Printf("Response from %s:\n%s\n", model, response.Content)
	return nil
}

func serviceAccountGeneration(projectID, location, model string) error {
	// Check if service account file is provided
	serviceAccountPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if serviceAccountPath == "" {
		fmt.Println("GOOGLE_APPLICATION_CREDENTIALS not set, skipping service account example")
		return nil
	}

	// Create provider with service account
	llm, err := provider.NewVertexAIProvider(
		projectID,
		location,
		model,
		domain.NewVertexAIServiceAccountOption(serviceAccountPath),
	)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	// Generate response
	ctx := context.Background()
	response, err := llm.Generate(ctx, "Explain the concept of machine learning in one sentence.")
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	fmt.Printf("Response: %s\n", response)
	return nil
}

func streamingGeneration(projectID, location, model string) error {
	// Create provider
	llm, err := provider.NewVertexAIProvider(projectID, location, model)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	// Create messages
	messages := []domain.Message{
		domain.NewTextMessage(domain.RoleUser, "Write a haiku about cloud computing."),
	}

	// Generate streaming response
	ctx := context.Background()
	stream, err := llm.StreamMessage(ctx, messages, domain.WithMaxTokens(100))
	if err != nil {
		return fmt.Errorf("streaming failed: %w", err)
	}

	fmt.Print("Streaming response: ")
	// Read the stream
	for token := range stream {
		if token.Text != "" {
			fmt.Print(token.Text)
		}
	}
	fmt.Println()

	return nil
}

func partnerModelGeneration(projectID, location string) error {
	// Use Claude model through Vertex AI
	claudeModel := "claude-3-5-haiku@20241022"

	llm, err := provider.NewVertexAIProvider(projectID, location, claudeModel)
	if err != nil {
		// Claude models might not be available in all regions
		if err.Error() == "model claude-3-5-haiku@20241022 is not available in region "+location {
			fmt.Printf("Claude models are not available in region %s\n", location)
			fmt.Println("Try using us-central1 or europe-west4 for Claude models")
			return nil
		}
		return fmt.Errorf("failed to create provider: %w", err)
	}

	// Generate response
	ctx := context.Background()
	response, err := llm.Generate(
		ctx,
		"What makes Claude unique compared to other language models?",
		domain.WithMaxTokens(150),
	)
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	fmt.Printf("Response from Claude via Vertex AI:\n%s\n", response)
	return nil
}

func multimodalGeneration(projectID, location, model string) error {
	// Only Gemini models support multimodal input
	if !strings.Contains(model, "gemini") {
		fmt.Printf("Model %s does not support multimodal input, skipping\n", model)
		return nil
	}

	// Create provider
	llm, err := provider.NewVertexAIProvider(projectID, location, model)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	// Create a simple test image (1x1 red pixel PNG)
	// In a real application, you would load an actual image file
	imageData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, 0xDE, 0x00, 0x00, 0x00,
		0x0C, 0x49, 0x44, 0x41, 0x54, 0x08, 0xD7, 0x63, 0xF8, 0xCF, 0xC0, 0x00,
		0x00, 0x03, 0x01, 0x01, 0x00, 0x18, 0xDD, 0x8D, 0xB4, 0x00, 0x00, 0x00,
		0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
	}

	// Create message with image
	imageMsg := domain.NewImageMessage(
		domain.RoleUser,
		imageData,
		"image/png",
		"What color is this image?",
	)

	// Generate response
	ctx := context.Background()
	response, err := llm.GenerateMessage(ctx, []domain.Message{imageMsg})
	if err != nil {
		return fmt.Errorf("multimodal generation failed: %w", err)
	}

	fmt.Printf("Multimodal response: %s\n", response.Content)
	return nil
}
