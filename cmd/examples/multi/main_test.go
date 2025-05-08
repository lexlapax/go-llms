package main

import (
	"context"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	"github.com/lexlapax/go-llms/pkg/testutils"
)

// TestMultiProviderFastest tests the fastest strategy of MultiProvider
func TestMultiProviderFastest(t *testing.T) {
	// Create mock providers with different latencies
	fastProvider := &testutils.TestMockProvider{
		GenerateFunc: func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
			time.Sleep(10 * time.Millisecond)
			return "Fast provider response", nil
		},
	}

	slowProvider := &testutils.TestMockProvider{
		GenerateFunc: func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
			time.Sleep(30 * time.Millisecond)
			return "Slow provider response", nil
		},
	}

	// Create providers with weights
	providers := []provider.ProviderWeight{
		{Provider: slowProvider, Weight: 1.0, Name: "slow"},
		{Provider: fastProvider, Weight: 1.0, Name: "fast"},
	}

	// Create MultiProvider with fastest strategy
	multiProvider := provider.NewMultiProvider(providers, provider.StrategyFastest)

	// Generate response
	result, err := multiProvider.Generate(context.Background(), "test prompt")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Should get response from fast provider
	if result != "Fast provider response" {
		t.Errorf("Expected result from fast provider, got: %s", result)
	}
}

// TestMultiProviderPrimary tests the primary strategy of MultiProvider
func TestMultiProviderPrimary(t *testing.T) {
	// The test is no longer flaky after our fix to use sequential execution for the primary strategy

	// Create mock providers with tracking variables
	var primaryCalled, secondaryCalled bool

	primaryProvider := &testutils.TestMockProvider{
		GenerateFunc: func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
			primaryCalled = true
			return "Primary provider response", nil
		},
	}

	secondaryProvider := &testutils.TestMockProvider{
		GenerateFunc: func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
			secondaryCalled = true
			return "Secondary provider response", nil
		},
	}

	// Create providers with weights
	providers := []provider.ProviderWeight{
		{Provider: primaryProvider, Weight: 1.0, Name: "primary"},
		{Provider: secondaryProvider, Weight: 1.0, Name: "secondary"},
	}

	// Create MultiProvider with primary strategy
	multiProvider := provider.NewMultiProvider(providers, provider.StrategyPrimary)

	// Reset tracking variables
	primaryCalled = false
	secondaryCalled = false

	// Generate response
	result, err := multiProvider.Generate(context.Background(), "test prompt")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Should get response from primary provider
	if result != "Primary provider response" {
		t.Errorf("Expected result from primary provider, got: %s", result)
	}

	// Verify primary was called but secondary wasn't
	if !primaryCalled {
		t.Errorf("Expected primary provider to be called")
	}

	if secondaryCalled {
		t.Errorf("Expected secondary provider not to be called when primary succeeds")
	}
}

// TestMultiProviderStreaming tests the streaming functionality of MultiProvider
func TestMultiProviderStreaming(t *testing.T) {
	// Create mock provider with streaming
	mockProvider := &testutils.TestMockProvider{
		StreamFunc: func(ctx context.Context, prompt string, options ...domain.Option) (domain.ResponseStream, error) {
			ch := make(chan domain.Token)
			go func() {
				defer close(ch)
				ch <- domain.Token{Text: "Token 1", Finished: false}
				time.Sleep(10 * time.Millisecond)
				ch <- domain.Token{Text: "Token 2", Finished: false}
				time.Sleep(10 * time.Millisecond)
				ch <- domain.Token{Text: "Token 3", Finished: true}
			}()
			return ch, nil
		},
	}

	// Create providers with weights
	providers := []provider.ProviderWeight{
		{Provider: mockProvider, Weight: 1.0, Name: "mock"},
	}

	// Create MultiProvider with fastest strategy
	multiProvider := provider.NewMultiProvider(providers, provider.StrategyFastest)

	// Create stream
	stream, err := multiProvider.Stream(context.Background(), "test prompt")
	if err != nil {
		t.Fatalf("Stream failed: %v", err)
	}

	// Collect tokens
	tokens := []string{}
	for token := range stream {
		tokens = append(tokens, token.Text)
	}

	// Should have 3 tokens
	if len(tokens) != 3 {
		t.Errorf("Expected 3 tokens, got %d", len(tokens))
	}
}
