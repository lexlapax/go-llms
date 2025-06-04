package main

import (
	"context"
	"testing"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
)

func TestOpenAIWithMock(t *testing.T) {
	t.Run("BasicGeneration", func(t *testing.T) {
		// Create a mock provider with organization option
		orgOption := domain.NewOpenAIOrganizationOption("test-org-id")
		mockProvider := provider.NewMockProvider(orgOption)

		// Set a simple response
		mockProvider.WithGenerateFunc(func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
			return "Mocked response for: " + prompt, nil
		})

		// Test basic generation
		response, err := mockProvider.Generate(context.Background(), "Test prompt")
		if err != nil {
			t.Fatalf("Error in Generate: %v", err)
		}
		if response == "" {
			t.Fatalf("Empty response received")
		}
	})

	t.Run("MessageConversation", func(t *testing.T) {
		// Create a mock provider
		mockProvider := provider.NewMockProvider()

		// Set a message response
		mockProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.Response, error) {
			return domain.Response{
				Content: "Mocked response for message conversation",
			}, nil
		})

		// Test message-based conversation
		messages := []domain.Message{
			domain.NewTextMessage(domain.RoleSystem, "You are a helpful assistant."),
			domain.NewTextMessage(domain.RoleUser, "Tell me about Go"),
		}

		response, err := mockProvider.GenerateMessage(context.Background(), messages)
		if err != nil {
			t.Fatalf("Error in GenerateMessage: %v", err)
		}
		if response.Content == "" {
			t.Fatalf("Empty response content received")
		}
	})

	t.Run("Streaming", func(t *testing.T) {
		// Create a mock provider
		mockProvider := provider.NewMockProvider()

		// Set a stream response
		mockProvider.WithStreamFunc(func(ctx context.Context, prompt string, options ...domain.Option) (domain.ResponseStream, error) {
			tokenCh := make(chan domain.Token, 3)

			go func() {
				tokenCh <- domain.Token{Text: "First ", Finished: false}
				tokenCh <- domain.Token{Text: "part. ", Finished: false}
				tokenCh <- domain.Token{Text: "Last part.", Finished: true}
				close(tokenCh)
			}()

			return tokenCh, nil
		})

		// Test streaming
		stream, err := mockProvider.Stream(context.Background(), "Stream test")
		if err != nil {
			t.Fatalf("Error in Stream: %v", err)
		}

		var tokens []string
		for token := range stream {
			tokens = append(tokens, token.Text)
		}

		if len(tokens) == 0 {
			t.Fatalf("No tokens received from stream")
		}
	})
}
