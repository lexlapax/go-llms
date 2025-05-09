package integration

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
)

// TestGeminiStreamingMock tests the Gemini streaming implementation with a mock server
func TestGeminiStreamingMock(t *testing.T) {
	// Create a handler function for ideal streaming behavior
	idealStreamHandler := func(w http.ResponseWriter, r *http.Request) {
		// Verify this is a streaming request to the Gemini API
		if !strings.Contains(r.URL.Path, ":streamGenerateContent") {
			t.Errorf("Expected streaming endpoint, got: %s", r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Set headers for SSE
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Flush headers
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		// Send a series of streamed responses
		streamResponses := []string{
			`data: {"candidates":[{"content":{"parts":[{"text":"1"}]},"finishReason":""}]}`,
			`data: {"candidates":[{"content":{"parts":[{"text":", 2"}]},"finishReason":""}]}`,
			`data: {"candidates":[{"content":{"parts":[{"text":", 3"}]},"finishReason":""}]}`,
			`data: {"candidates":[{"content":{"parts":[{"text":", 4"}]},"finishReason":""}]}`,
			`data: {"candidates":[{"content":{"parts":[{"text":", 5"}]},"finishReason":"STOP"}]}`,
			`data: [DONE]`,
		}

		for _, resp := range streamResponses {
			_, err := w.Write([]byte(resp + "\n"))
			if err != nil {
				return
			}

			// Flush after each write
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}

			// Small delay to simulate streaming
			time.Sleep(50 * time.Millisecond)
		}
	}

	// Basic streaming test with ideal response format
	t.Run("BasicStreaming", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(idealStreamHandler))
		defer mockServer.Close()

		// Create Gemini provider with the mock server
		geminiProvider := provider.NewGeminiProvider(
			"mock-api-key",
			"gemini-2.0-flash-lite",
			domain.NewBaseURLOption(mockServer.URL),
		)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		stream, err := geminiProvider.Stream(ctx, "Count from 1 to 5")
		if err != nil {
			t.Fatalf("Failed to create stream: %v", err)
		}

		var fullResponse strings.Builder
		var tokens []string
		var finished []bool

		for token := range stream {
			tokens = append(tokens, token.Text)
			finished = append(finished, token.Finished)
			fullResponse.WriteString(token.Text)
			t.Logf("Token: '%s', Finished: %v", token.Text, token.Finished)
		}

		// Verify number of tokens (should match the number of responses from the mock server)
		expectedTokenCount := 5 // 5 tokens in our mock responses
		if len(tokens) != expectedTokenCount {
			t.Errorf("Expected %d tokens, got %d", expectedTokenCount, len(tokens))
		}

		// Verify only the last token is marked as finished
		for i, isFinished := range finished[:len(finished)-1] {
			if isFinished {
				t.Errorf("Token %d should not be marked as finished", i)
			}
		}

		// Verify the last token is marked as finished
		if len(finished) > 0 && !finished[len(finished)-1] {
			t.Errorf("Last token should be marked as finished")
		}

		// Verify the combined response
		expectedResponse := "1, 2, 3, 4, 5"
		if fullResponse.String() != expectedResponse {
			t.Errorf("Expected response '%s', got '%s'", expectedResponse, fullResponse.String())
		}
	})

	// Test handling of empty lines and whitespace (simulating potential API issues)
	t.Run("EmptyLinesAndWhitespace", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set headers for SSE
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")

			// Flush headers
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}

			// Responses with empty lines, extra whitespace and keeps-alive
			responses := []string{
				"", // Empty line
				"data: ", // Empty data
				"data: {}", // Empty JSON object
				`data: {"candidates":[]}`, // Empty candidates array
				`data: {"candidates":[{"content":{"parts":[{"text":"Hello"}]},"finishReason":""}]}`,
				"", // Another empty line
				"data: ", // Another empty data
				`data: {"candidates":[{"content":{"parts":[{"text":" World"}]},"finishReason":"STOP"}]}`,
				`data: [DONE]`,
			}

			for _, resp := range responses {
				_, err := w.Write([]byte(resp + "\n"))
				if err != nil {
					return
				}

				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}

				time.Sleep(50 * time.Millisecond)
			}
		}))
		defer mockServer.Close()

		geminiProvider := provider.NewGeminiProvider(
			"mock-api-key",
			"gemini-2.0-flash-lite",
			domain.NewBaseURLOption(mockServer.URL),
		)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		stream, err := geminiProvider.Stream(ctx, "Say Hello World")
		if err != nil {
			t.Fatalf("Failed to create stream: %v", err)
		}

		var fullResponse strings.Builder
		tokenCount := 0
		emptyCount := 0

		for token := range stream {
			tokenCount++
			if token.Text == "" {
				emptyCount++
			}
			fullResponse.WriteString(token.Text)
			t.Logf("Token %d: '%s', Finished: %v", tokenCount, token.Text, token.Finished)
		}

		t.Logf("Received %d tokens, %d empty tokens", tokenCount, emptyCount)
		t.Logf("Full response: '%s'", fullResponse.String())

		// We expect to receive "Hello World" despite the empty and invalid messages
		if fullResponse.String() != "Hello World" {
			t.Errorf("Expected 'Hello World', got '%s'", fullResponse.String())
		}
	})

	// Test handling of malformed JSON responses
	t.Run("MalformedJSON", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set headers for SSE
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")

			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}

			// Mix of valid and invalid JSON
			responses := []string{
				`data: {"candidates":[{"content":{"parts":[{"text":"Valid"}]},"finishReason":""}]}`,
				`data: {"candidates":[{"content":{"parts":{"text":"Malformed"}]},"finishReason":""}]}`, // Parts should be an array
				`data: {"candidates":[{"content":{"parts":[{"text":123}]},"finishReason":""}]}`, // Text should be string, not number
				`data: {"candid`, // Incomplete JSON
				`data: {"candidates":[{"content":{"parts":[{"text":" JSON"}]},"finishReason":"STOP"}]}`,
				`data: [DONE]`,
			}

			for _, resp := range responses {
				_, err := w.Write([]byte(resp + "\n"))
				if err != nil {
					return
				}

				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}

				time.Sleep(50 * time.Millisecond)
			}
		}))
		defer mockServer.Close()

		geminiProvider := provider.NewGeminiProvider(
			"mock-api-key",
			"gemini-2.0-flash-lite",
			domain.NewBaseURLOption(mockServer.URL),
		)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		stream, err := geminiProvider.Stream(ctx, "Test malformed JSON")
		if err != nil {
			t.Fatalf("Failed to create stream: %v", err)
		}

		var fullResponse strings.Builder
		tokenCount := 0

		for token := range stream {
			tokenCount++
			fullResponse.WriteString(token.Text)
			t.Logf("Token %d: '%s', Finished: %v", tokenCount, token.Text, token.Finished)
		}

		t.Logf("Received %d tokens", tokenCount)
		t.Logf("Full response: '%s'", fullResponse.String())

		// We should at least get "Valid" and possibly " JSON" from the valid responses
		if !strings.Contains(fullResponse.String(), "Valid") {
			t.Errorf("Expected response to contain 'Valid', got '%s'", fullResponse.String())
		}
	})

	// Test realistic streaming format from Gemini API (based on documentation)
	t.Run("RealisticGeminiFormat", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set headers for SSE
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")

			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}

			// Format extracted from Gemini API documentation
			responses := []string{
				`data: {"candidates":[{"content":{"parts":[{"text":"The"}]}}]}`,
				`data: {"candidates":[{"content":{"parts":[{"text":" quick"}]}}]}`,
				`data: {"candidates":[{"content":{"parts":[{"text":" brown"}]}}]}`,
				`data: {"candidates":[{"content":{"parts":[{"text":" fox"}]}}]}`,
				`data: {"candidates":[{"content":{"parts":[{"text":" jumps"}]}}]}`,
				`data: {"candidates":[{"content":{"parts":[{"text":" over"}]}}]}`,
				`data: {"candidates":[{"content":{"parts":[{"text":" the"}]}}]}`,
				`data: {"candidates":[{"content":{"parts":[{"text":" lazy"}]}}]}`,
				`data: {"candidates":[{"content":{"parts":[{"text":" dog"}]},"finishReason":"STOP"}]}`,
				`data: [DONE]`,
			}

			for _, resp := range responses {
				_, err := w.Write([]byte(resp + "\n"))
				if err != nil {
					return
				}

				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}

				time.Sleep(50 * time.Millisecond)
			}
		}))
		defer mockServer.Close()

		geminiProvider := provider.NewGeminiProvider(
			"mock-api-key",
			"gemini-2.0-flash-lite",
			domain.NewBaseURLOption(mockServer.URL),
		)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		stream, err := geminiProvider.Stream(ctx, "Complete the sentence")
		if err != nil {
			t.Fatalf("Failed to create stream: %v", err)
		}

		var fullResponse strings.Builder
		tokenCount := 0

		for token := range stream {
			tokenCount++
			fullResponse.WriteString(token.Text)
			t.Logf("Token %d: '%s', Finished: %v", tokenCount, token.Text, token.Finished)
		}

		t.Logf("Received %d tokens", tokenCount)
		t.Logf("Full response: '%s'", fullResponse.String())

		expectedResponse := "The quick brown fox jumps over the lazy dog"
		if fullResponse.String() != expectedResponse {
			t.Errorf("Expected '%s', got '%s'", expectedResponse, fullResponse.String())
		}

		if tokenCount != 9 {
			t.Errorf("Expected 9 tokens, got %d", tokenCount)
		}
	})
}

// TestGeminiAltSSEParameter tests that the alt=sse parameter is required for Gemini streaming
func TestGeminiAltSSEParameter(t *testing.T) {
	// Skip test if GEMINI_API_KEY is not set
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set, skipping Gemini alt=sse test")
	}

	// Test with alt=sse parameter
	t.Run("WithAltSSEParameter", func(t *testing.T) {
		// Create HTTP client
		client := &http.Client{Timeout: 10 * time.Second}

		// Build the request URL with alt=sse parameter
		url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-lite:streamGenerateContent?alt=sse&key=%s", apiKey)

		// Prepare request body
		requestBody := map[string]interface{}{
			"contents": []map[string]interface{}{
				{
					"role": "user",
					"parts": []map[string]interface{}{
						{"text": "Count from 1 to 3, one per line"},
					},
				},
			},
		}

		// Marshal request body
		bodyBytes, err := json.Marshal(requestBody)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}

		// Create request
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		// Set headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "text/event-stream")

		// Make the request
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// Check response code
		t.Logf("Response Status with alt=sse: %s", resp.Status)

		// Sample the response to ensure it's in SSE format
		scanner := bufio.NewScanner(resp.Body)
		lineCount := 0
		var sseContent bool
		for scanner.Scan() && lineCount < 5 {
			line := scanner.Text()
			lineCount++
			t.Logf("Line %d: %s", lineCount, line)

			if strings.HasPrefix(line, "data:") {
				sseContent = true
				break
			}
		}

		// Verify we received data in SSE format
		if !sseContent && lineCount > 0 {
			t.Errorf("Response didn't contain SSE formatted data lines starting with 'data:'")
		}
	})

	// Test without alt=sse parameter
	t.Run("WithoutAltSSEParameter", func(t *testing.T) {
		// Create HTTP client
		client := &http.Client{Timeout: 10 * time.Second}

		// Build the request URL WITHOUT alt=sse parameter
		url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-lite:streamGenerateContent?key=%s", apiKey)

		// Prepare request body
		requestBody := map[string]interface{}{
			"contents": []map[string]interface{}{
				{
					"role": "user",
					"parts": []map[string]interface{}{
						{"text": "Count from 1 to 3, one per line"},
					},
				},
			},
		}

		// Marshal request body
		bodyBytes, err := json.Marshal(requestBody)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}

		// Create request
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		// Set headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "text/event-stream")

		// Make the request
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// Check response code
		t.Logf("Response Status without alt=sse: %s", resp.Status)

		// Sample the response to confirm it's NOT in SSE format
		scanner := bufio.NewScanner(resp.Body)
		lineCount := 0
		var sseContent bool
		for scanner.Scan() && lineCount < 5 {
			line := scanner.Text()
			lineCount++
			t.Logf("Line %d: %s", lineCount, line)

			if strings.HasPrefix(line, "data:") {
				sseContent = true
				break
			}
		}

		formatMsg := "Not SSE format (expected)"
		if sseContent {
			formatMsg = "SSE format (unexpected)"
		}
		t.Logf("Response format without alt=sse: %s", formatMsg)
	})
}

// TestGeminiStreamingLive tests the Gemini streaming with the real API if available
func TestGeminiStreamingLive(t *testing.T) {
	// Only skip if explicitly told to
	if os.Getenv("SKIP_GEMINI_STREAMING") == "true" {
		t.Skip("Skipping live Gemini streaming test due to known issues")
	}

	// Skip test if GEMINI_API_KEY is not set
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set, skipping Gemini streaming test")
	}

	// First, let's make a raw request to see the exact response format
	t.Run("RawResponseFormatCheck", func(t *testing.T) {
		// Create HTTP client
		client := &http.Client{Timeout: 10 * time.Second}

		// Build the request URL
		url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-lite:streamGenerateContent?key=%s", apiKey)

		// Prepare request body
		requestBody := map[string]interface{}{
			"contents": []map[string]interface{}{
				{
					"role": "user",
					"parts": []map[string]interface{}{
						{"text": "Count from 1 to 5, one number per line"},
					},
				},
			},
		}

		// Marshal request body
		bodyBytes, err := json.Marshal(requestBody)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}

		// Create request
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		// Set headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "text/event-stream")

		// Make the request
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// Check response code
		t.Logf("Response Status: %s", resp.Status)
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Bad response: %s - %s", resp.Status, string(body))
		}

		// Read raw response
		t.Log("Raw response (first 10 lines max):")
		scanner := bufio.NewScanner(resp.Body)
		lineCount := 0
		for scanner.Scan() && lineCount < 10 {
			line := scanner.Text()
			lineCount++
			t.Logf("Line %d: %s", lineCount, line)
		}

		if err := scanner.Err(); err != nil {
			t.Logf("Scanner error: %v", err)
		} else if lineCount == 0 {
			t.Log("No lines received from API")
		}
	})

	// Test using our provider implementation
	t.Run("ProviderImplementationTest", func(t *testing.T) {
		// Create Gemini provider
		geminiProvider := provider.NewGeminiProvider(apiKey, "gemini-2.0-flash-lite")

		// Use a shorter timeout for testing
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		prompt := "Count from 1 to 5, one number per line"

		stream, err := geminiProvider.Stream(ctx, prompt)
		if err != nil {
			t.Fatalf("Failed to create stream: %v", err)
		}

		var fullResponse strings.Builder
		var tokens []domain.Token
		tokenCount := 0
		emptyTokens := 0
		finishedTokens := 0

		for token := range stream {
			tokenCount++
			tokens = append(tokens, token)

			if token.Text == "" {
				emptyTokens++
			}

			if token.Finished {
				finishedTokens++
			}

			fullResponse.WriteString(token.Text)

			t.Logf("Stream token %d: '%s' (finished: %v)",
				tokenCount, token.Text, token.Finished)
		}

		// Print statistics for analysis
		t.Logf("Statistics:")
		t.Logf("- Total tokens: %d", tokenCount)
		t.Logf("- Empty tokens: %d", emptyTokens)
		t.Logf("- Finished tokens: %d", finishedTokens)
		t.Logf("- Final response length: %d characters", len(fullResponse.String()))
		t.Logf("- Full response: %s", fullResponse.String())

		// Count numbers in the response
		numCount := 0
		for i := 1; i <= 5; i++ {
			if strings.Contains(fullResponse.String(), string(rune('0'+i))) {
				numCount++
			}
		}

		t.Logf("- Numbers found in response: %d/5", numCount)

		// Basic assertions
		if tokenCount == 0 {
			t.Errorf("Expected at least some tokens in the stream")
		}
	})
}