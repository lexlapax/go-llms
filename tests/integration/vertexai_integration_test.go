package integration

// ABOUTME: Integration tests for Google Vertex AI provider implementation
// ABOUTME: Tests authentication methods, region support, and partner models

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func skipIfNoVertexAI(t *testing.T) {
	// Check for either service account or ADC setup
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" && 
	   os.Getenv("VERTEXAI_PROJECT") == "" {
		t.Skip("Skipping Vertex AI integration test: GOOGLE_APPLICATION_CREDENTIALS or VERTEXAI_PROJECT not set")
	}
}

func getVertexAIConfig() (project, region string) {
	project = os.Getenv("VERTEXAI_PROJECT")
	if project == "" {
		project = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}
	
	region = os.Getenv("VERTEXAI_REGION")
	if region == "" {
		region = "us-central1" // Default region
	}
	
	return project, region
}

func TestVertexAIIntegration_ServiceAccountAuth(t *testing.T) {
	skipIfNoVertexAI(t)
	
	// Skip if no service account credentials
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		t.Skip("Skipping service account test: GOOGLE_APPLICATION_CREDENTIALS not set")
	}
	
	project, region := getVertexAIConfig()
	if project == "" {
		t.Skip("Skipping test: project ID not found")
	}
	
	// Create provider with service account authentication
	llm, err := provider.NewVertexAIProvider(project, region, "gemini-1.5-flash")
	require.NoError(t, err)
	
	messages := []domain.Message{
		domain.NewTextMessage(domain.RoleUser, "Say 'Hello from Vertex AI' and nothing else."),
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	response, err := llm.GenerateMessage(ctx, messages, domain.WithMaxTokens(50))
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Contains(t, response.Content, "Hello from Vertex AI")
}

func TestVertexAIIntegration_ADCAuth(t *testing.T) {
	skipIfNoVertexAI(t)
	
	// This test uses Application Default Credentials
	project, region := getVertexAIConfig()
	if project == "" {
		t.Skip("Skipping test: project ID not found")
	}
	
	// Temporarily unset service account to test ADC
	oldCreds := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	_ = os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
	defer func() {
		if oldCreds != "" {
			_ = os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", oldCreds)
		}
	}()
	
	// Create provider with ADC
	llm, err := provider.NewVertexAIProvider(project, region, "gemini-1.5-flash")
	require.NoError(t, err)
	
	messages := []domain.Message{
		domain.NewTextMessage(domain.RoleUser, "What is 2+2?"),
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	response, generateErr := llm.GenerateMessage(ctx, messages, domain.WithMaxTokens(20))
	
	// If ADC is not configured, skip rather than fail
	if generateErr != nil && strings.Contains(generateErr.Error(), "credentials") {
		t.Skip("Skipping ADC test: Application Default Credentials not configured")
	}
	
	require.NoError(t, generateErr)
	require.NotNil(t, response)
	assert.Contains(t, response.Content, "4")
}

func TestVertexAIIntegration_Streaming(t *testing.T) {
	skipIfNoVertexAI(t)
	
	project, region := getVertexAIConfig()
	if project == "" {
		t.Skip("Skipping test: project ID not found")
	}
	
	llm, err := provider.NewVertexAIProvider(project, region, "gemini-1.5-flash")
	require.NoError(t, err)
	
	messages := []domain.Message{
		domain.NewTextMessage(domain.RoleUser, "Count from 1 to 5"),
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	stream, err := llm.StreamMessage(ctx, messages, domain.WithMaxTokens(50))
	require.NoError(t, err)
	require.NotNil(t, stream)
	
	var fullContent string
	chunkCount := 0
	
	for token := range stream {
		if token.Text != "" {
			fullContent += token.Text
			chunkCount++
		}
	}
	
	assert.Greater(t, chunkCount, 1, "Expected multiple chunks in streaming response")
	assert.NotEmpty(t, fullContent)
	t.Logf("Received %d chunks, full content: %s", chunkCount, fullContent)
}

func TestVertexAIIntegration_DifferentModels(t *testing.T) {
	skipIfNoVertexAI(t)
	
	project, region := getVertexAIConfig()
	if project == "" {
		t.Skip("Skipping test: project ID not found")
	}
	
	testCases := []struct {
		name  string
		model string
	}{
		{
			name:  "Gemini 1.5 Flash",
			model: "gemini-1.5-flash",
		},
		{
			name:  "Gemini 1.5 Pro",
			model: "gemini-1.5-pro",
		},
		{
			name:  "Gemini 1.0 Pro",
			model: "gemini-1.0-pro",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			llm, err := provider.NewVertexAIProvider(project, region, tc.model)
			require.NoError(t, err)
			
			messages := []domain.Message{
				domain.NewTextMessage(domain.RoleUser, "What is the capital of France?"),
			}
			
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			
			response, err := llm.GenerateMessage(ctx, messages, domain.WithMaxTokens(50))
			require.NoError(t, err)
			require.NotNil(t, response)
			assert.Contains(t, strings.ToLower(response.Content), "paris")
		})
	}
}

func TestVertexAIIntegration_PartnerModels(t *testing.T) {
	skipIfNoVertexAI(t)
	
	project, region := getVertexAIConfig()
	if project == "" {
		t.Skip("Skipping test: project ID not found")
	}
	
	// Test Claude model if available
	// Note: This requires the Claude model to be enabled in your Vertex AI project
	testCases := []struct {
		name    string
		model   string
		canSkip bool
	}{
		{
			name:    "Claude 3 Haiku",
			model:   "claude-3-haiku@20240307",
			canSkip: true, // Partner models may not be available
		},
		{
			name:    "Claude 3 Sonnet",
			model:   "claude-3-sonnet@20240229",
			canSkip: true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			llm, err := provider.NewVertexAIProvider(project, region, tc.model)
			require.NoError(t, err)
			
			messages := []domain.Message{
				domain.NewTextMessage(domain.RoleUser, "Say hello"),
			}
			
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			
			response, err := llm.GenerateMessage(ctx, messages, domain.WithMaxTokens(50))
			
			if err != nil && tc.canSkip {
				t.Skipf("Partner model %s not available: %v", tc.model, err)
				return
			}
			
			require.NoError(t, err)
			require.NotNil(t, response)
			assert.NotEmpty(t, response.Content)
		})
	}
}

func TestVertexAIIntegration_RegionSpecific(t *testing.T) {
	skipIfNoVertexAI(t)
	
	project, _ := getVertexAIConfig()
	if project == "" {
		t.Skip("Skipping test: project ID not found")
	}
	
	// Test different regions
	regions := []string{
		"us-central1",
		"europe-west4",
		"asia-northeast1",
	}
	
	for _, region := range regions {
		t.Run(region, func(t *testing.T) {
			llm, err := provider.NewVertexAIProvider(project, region, "gemini-1.5-flash")
			require.NoError(t, err)
			
			messages := []domain.Message{
				domain.NewTextMessage(domain.RoleUser, "What region are you in?"),
			}
			
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			
			response, err := llm.GenerateMessage(ctx, messages, domain.WithMaxTokens(100))
			
			// Some regions might not be available, skip if so
			if err != nil && strings.Contains(err.Error(), "region") {
				t.Skipf("Region %s not available: %v", region, err)
				return
			}
			
			require.NoError(t, err)
			require.NotNil(t, response)
			assert.NotEmpty(t, response.Content)
		})
	}
}

func TestVertexAIIntegration_ErrorHandling(t *testing.T) {
	skipIfNoVertexAI(t)
	
	project, region := getVertexAIConfig()
	if project == "" {
		t.Skip("Skipping test: project ID not found")
	}
	
	testCases := []struct {
		name          string
		project       string
		region        string
		model         string
		expectedError string
	}{
		{
			name:          "Invalid Project",
			project:       "invalid-project-12345",
			region:        region,
			model:         "gemini-1.5-flash",
			expectedError: "project",
		},
		{
			name:          "Invalid Model",
			project:       project,
			region:        region,
			model:         "nonexistent-model",
			expectedError: "model",
		},
		{
			name:          "Invalid Region",
			project:       project,
			region:        "invalid-region",
			model:         "gemini-1.5-flash",
			expectedError: "region",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			llm, err := provider.NewVertexAIProvider(tc.project, tc.region, tc.model)
			require.NoError(t, err)
			
			messages := []domain.Message{
				domain.NewTextMessage(domain.RoleUser, "Test"),
			}
			
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			
			_, generateErr := llm.GenerateMessage(ctx, messages)
			require.Error(t, generateErr)
			assert.Contains(t, strings.ToLower(generateErr.Error()), tc.expectedError)
		})
	}
}

func TestVertexAIIntegration_LongContext(t *testing.T) {
	skipIfNoVertexAI(t)
	
	project, region := getVertexAIConfig()
	if project == "" {
		t.Skip("Skipping test: project ID not found")
	}
	
	// Gemini 1.5 models support very long context
	llm, err := provider.NewVertexAIProvider(project, region, "gemini-1.5-flash")
	require.NoError(t, err)
	
	// Create a long message
	longText := strings.Repeat("This is a test sentence. ", 1000) // ~5000 tokens
	
	messages := []domain.Message{
		domain.NewTextMessage(domain.RoleUser, longText + "\n\nSummarize the above in one sentence."),
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	
	response, err := llm.GenerateMessage(ctx, messages, domain.WithMaxTokens(100))
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.NotEmpty(t, response.Content)
	assert.Less(t, len(response.Content), len(longText), "Summary should be shorter than input")
}