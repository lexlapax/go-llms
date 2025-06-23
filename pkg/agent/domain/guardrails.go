// ABOUTME: Defines guardrails for validating agent inputs and outputs
// ABOUTME: Provides interfaces for sync/async validation and guardrail chains

package domain

import (
	"context"
	"fmt"
	"time"
)

// GuardrailType represents when the guardrail is applied.
// Guardrails can validate inputs, outputs, or both depending on the type.
type GuardrailType string

const (
	GuardrailTypeInput  GuardrailType = "input"
	GuardrailTypeOutput GuardrailType = "output"
	GuardrailTypeBoth   GuardrailType = "both"
)

// Guardrail validates agent inputs and/or outputs to ensure safety and compliance.
// Guardrails can perform both synchronous and asynchronous validation
// with configurable timeouts for complex validation scenarios.
type Guardrail interface {
	Name() string
	Type() GuardrailType

	// Validation
	Validate(ctx context.Context, state *State) error

	// Async validation with timeout
	ValidateAsync(ctx context.Context, state *State, timeout time.Duration) <-chan error
}

// GuardrailFunc is a function adapter for simple guardrails
type GuardrailFunc func(ctx context.Context, state *State) error

// guardrailFuncImpl wraps a function as a Guardrail
type guardrailFuncImpl struct {
	name       string
	guardType  GuardrailType
	validateFn GuardrailFunc
}

// NewGuardrailFunc creates a guardrail from a function
func NewGuardrailFunc(name string, guardType GuardrailType, fn GuardrailFunc) Guardrail {
	return &guardrailFuncImpl{
		name:       name,
		guardType:  guardType,
		validateFn: fn,
	}
}

func (g *guardrailFuncImpl) Name() string {
	return g.name
}

func (g *guardrailFuncImpl) Type() GuardrailType {
	return g.guardType
}

func (g *guardrailFuncImpl) Validate(ctx context.Context, state *State) error {
	return g.validateFn(ctx, state)
}

func (g *guardrailFuncImpl) ValidateAsync(ctx context.Context, state *State, timeout time.Duration) <-chan error {
	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)

		// Create timeout context
		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		// Run validation
		done := make(chan error, 1)
		go func() {
			done <- g.validateFn(timeoutCtx, state)
		}()

		// Wait for completion or timeout
		select {
		case err := <-done:
			errCh <- err
		case <-timeoutCtx.Done():
			errCh <- fmt.Errorf("guardrail %s validation timeout after %v", g.name, timeout)
		}
	}()

	return errCh
}

// GuardrailChain runs multiple guardrails
type GuardrailChain struct {
	guardrails []Guardrail
	failFast   bool
	name       string
	guardType  GuardrailType
}

// NewGuardrailChain creates a new guardrail chain
func NewGuardrailChain(name string, guardType GuardrailType, failFast bool) *GuardrailChain {
	return &GuardrailChain{
		name:       name,
		guardType:  guardType,
		guardrails: make([]Guardrail, 0),
		failFast:   failFast,
	}
}

// Add adds a guardrail to the chain
func (gc *GuardrailChain) Add(guardrail Guardrail) *GuardrailChain {
	gc.guardrails = append(gc.guardrails, guardrail)
	return gc
}

// Name returns the chain name
func (gc *GuardrailChain) Name() string {
	return gc.name
}

// Type returns the guardrail type
func (gc *GuardrailChain) Type() GuardrailType {
	return gc.guardType
}

// Validate runs all guardrails in the chain
func (gc *GuardrailChain) Validate(ctx context.Context, state *State) error {
	var errors []error

	for _, g := range gc.guardrails {
		// Skip if guardrail type doesn't match
		if gc.guardType != GuardrailTypeBoth && g.Type() != GuardrailTypeBoth && g.Type() != gc.guardType {
			continue
		}

		if err := g.Validate(ctx, state); err != nil {
			if gc.failFast {
				return fmt.Errorf("guardrail %s failed: %w", g.Name(), err)
			}
			errors = append(errors, fmt.Errorf("guardrail %s: %w", g.Name(), err))
		}
	}

	if len(errors) > 0 {
		return &MultiError{Errors: errors}
	}

	return nil
}

// ValidateAsync runs all guardrails asynchronously
func (gc *GuardrailChain) ValidateAsync(ctx context.Context, state *State, timeout time.Duration) <-chan error {
	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)

		// Create timeout context
		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		// Run validation
		done := make(chan error, 1)
		go func() {
			done <- gc.Validate(timeoutCtx, state)
		}()

		// Wait for completion or timeout
		select {
		case err := <-done:
			errCh <- err
		case <-timeoutCtx.Done():
			errCh <- fmt.Errorf("guardrail chain %s validation timeout after %v", gc.name, timeout)
		}
	}()

	return errCh
}

// Common guardrails

// RequiredKeysGuardrail ensures required keys exist in state
func RequiredKeysGuardrail(name string, keys ...string) Guardrail {
	return NewGuardrailFunc(name, GuardrailTypeInput, func(ctx context.Context, state *State) error {
		var missing []string
		for _, key := range keys {
			if _, ok := state.Get(key); !ok {
				missing = append(missing, key)
			}
		}
		if len(missing) > 0 {
			return fmt.Errorf("missing required keys: %v", missing)
		}
		return nil
	})
}

// MaxStateSizeGuardrail ensures state doesn't exceed size limit
func MaxStateSizeGuardrail(name string, maxBytes int64) Guardrail {
	return NewGuardrailFunc(name, GuardrailTypeBoth, func(ctx context.Context, state *State) error {
		// Serialize state to check size
		data, err := state.MarshalJSON()
		if err != nil {
			return fmt.Errorf("failed to check state size: %w", err)
		}

		size := int64(len(data))
		if size > maxBytes {
			return fmt.Errorf("state size %d bytes exceeds limit of %d bytes", size, maxBytes)
		}

		return nil
	})
}

// MessageCountGuardrail limits the number of messages
func MessageCountGuardrail(name string, maxMessages int) Guardrail {
	return NewGuardrailFunc(name, GuardrailTypeBoth, func(ctx context.Context, state *State) error {
		count := len(state.Messages())
		if count > maxMessages {
			return fmt.Errorf("message count %d exceeds limit of %d", count, maxMessages)
		}
		return nil
	})
}

// ContentModerationGuardrail checks for prohibited content
func ContentModerationGuardrail(name string, prohibitedWords []string) Guardrail {
	// Create a map for faster lookup
	prohibited := make(map[string]bool)
	for _, word := range prohibitedWords {
		prohibited[word] = true
	}

	return NewGuardrailFunc(name, GuardrailTypeBoth, func(ctx context.Context, state *State) error {
		// Check messages
		for _, msg := range state.Messages() {
			if containsProhibited(msg.Content, prohibited) {
				return fmt.Errorf("message contains prohibited content")
			}
		}

		// Check string values in state
		for key, value := range state.Values() {
			if str, ok := value.(string); ok {
				if containsProhibited(str, prohibited) {
					return fmt.Errorf("state key %s contains prohibited content", key)
				}
			}
		}

		return nil
	})
}

// containsProhibited checks if text contains prohibited words
func containsProhibited(text string, prohibited map[string]bool) bool {
	// Simple word-based check - could be enhanced with more sophisticated NLP
	words := tokenize(text)
	for _, word := range words {
		if prohibited[word] {
			return true
		}
	}
	return false
}

// tokenize splits text into words (simplified)
func tokenize(text string) []string {
	// This is a simplified tokenizer - production code might use a proper NLP tokenizer
	var words []string
	current := ""

	for _, r := range text {
		if isWordChar(r) {
			current += string(r)
		} else if current != "" {
			words = append(words, current)
			current = ""
		}
	}

	if current != "" {
		words = append(words, current)
	}

	return words
}

// isWordChar checks if a rune is part of a word
func isWordChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}
