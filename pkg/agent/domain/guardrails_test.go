// ABOUTME: Tests for the Guardrails interface and implementations that validate agent inputs and outputs
// ABOUTME: including sync/async validation, composable guardrails, and pre-built validators

package domain

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestGuardrailFunc(t *testing.T) {
	tests := []struct {
		name        string
		guardrail   Guardrail
		state       *State
		expectError bool
	}{
		{
			name: "allow all guardrail passes",
			guardrail: NewGuardrailFunc("allow-all", GuardrailTypeBoth, func(ctx context.Context, state *State) error {
				return nil
			}),
			state:       NewState(),
			expectError: false,
		},
		{
			name: "deny all guardrail fails",
			guardrail: NewGuardrailFunc("deny-all", GuardrailTypeBoth, func(ctx context.Context, state *State) error {
				return errors.New("denied")
			}),
			state:       NewState(),
			expectError: true,
		},
		{
			name: "check state value guardrail",
			guardrail: NewGuardrailFunc("check-value", GuardrailTypeInput, func(ctx context.Context, state *State) error {
				if val, ok := state.Get("required"); !ok || val != "present" {
					return errors.New("required value not present")
				}
				return nil
			}),
			state: func() *State {
				s := NewState()
				s.Set("required", "present")
				return s
			}(),
			expectError: false,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.guardrail.Validate(ctx, tt.state)
			if (err != nil) != tt.expectError {
				t.Errorf("Validate() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestGuardrailType(t *testing.T) {
	tests := []struct {
		name         string
		guardrail    Guardrail
		expectedType GuardrailType
		expectedName string
	}{
		{
			name:         "input guardrail",
			guardrail:    NewGuardrailFunc("test-input", GuardrailTypeInput, nil),
			expectedType: GuardrailTypeInput,
			expectedName: "test-input",
		},
		{
			name:         "output guardrail",
			guardrail:    NewGuardrailFunc("test-output", GuardrailTypeOutput, nil),
			expectedType: GuardrailTypeOutput,
			expectedName: "test-output",
		},
		{
			name:         "both guardrail",
			guardrail:    NewGuardrailFunc("test-both", GuardrailTypeBoth, nil),
			expectedType: GuardrailTypeBoth,
			expectedName: "test-both",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.guardrail.Type() != tt.expectedType {
				t.Errorf("Type() = %v, want %v", tt.guardrail.Type(), tt.expectedType)
			}
			if tt.guardrail.Name() != tt.expectedName {
				t.Errorf("Name() = %v, want %v", tt.guardrail.Name(), tt.expectedName)
			}
		})
	}
}

func TestGuardrailChain(t *testing.T) {
	// Create a chain that fails fast
	failFastChain := NewGuardrailChain("fail-fast-chain", GuardrailTypeBoth, true)
	failFastChain.
		Add(NewGuardrailFunc("pass", GuardrailTypeBoth, func(ctx context.Context, state *State) error {
			return nil
		})).
		Add(NewGuardrailFunc("fail", GuardrailTypeBoth, func(ctx context.Context, state *State) error {
			return errors.New("failed")
		})).
		Add(NewGuardrailFunc("never-reached", GuardrailTypeBoth, func(ctx context.Context, state *State) error {
			t.Error("This guardrail should not be reached in fail-fast mode")
			return nil
		}))

	// Create a chain that collects all errors
	collectAllChain := NewGuardrailChain("collect-all-chain", GuardrailTypeBoth, false)
	collectAllChain.
		Add(NewGuardrailFunc("fail1", GuardrailTypeBoth, func(ctx context.Context, state *State) error {
			return errors.New("error1")
		})).
		Add(NewGuardrailFunc("fail2", GuardrailTypeBoth, func(ctx context.Context, state *State) error {
			return errors.New("error2")
		})).
		Add(NewGuardrailFunc("pass", GuardrailTypeBoth, func(ctx context.Context, state *State) error {
			return nil
		}))

	ctx := context.Background()
	state := NewState()

	// Test fail-fast chain
	err := failFastChain.Validate(ctx, state)
	if err == nil {
		t.Error("Expected fail-fast chain to return an error")
	}
	if !strings.Contains(err.Error(), "guardrail fail failed") {
		t.Errorf("Expected error to mention 'fail' guardrail, got: %v", err)
	}

	// Test collect-all chain
	err = collectAllChain.Validate(ctx, state)
	if err == nil {
		t.Error("Expected collect-all chain to return an error")
	}
	multiErr, ok := err.(*MultiError)
	if !ok {
		t.Error("Expected MultiError from collect-all chain")
	}
	if len(multiErr.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(multiErr.Errors))
	}
}

func TestGuardrailChain_TypeFiltering(t *testing.T) {
	// Create a chain with mixed types
	chain := NewGuardrailChain("mixed-chain", GuardrailTypeInput, false)

	inputCalled := false
	outputCalled := false
	bothCalled := false

	chain.
		Add(NewGuardrailFunc("input-only", GuardrailTypeInput, func(ctx context.Context, state *State) error {
			inputCalled = true
			return nil
		})).
		Add(NewGuardrailFunc("output-only", GuardrailTypeOutput, func(ctx context.Context, state *State) error {
			outputCalled = true
			return nil
		})).
		Add(NewGuardrailFunc("both", GuardrailTypeBoth, func(ctx context.Context, state *State) error {
			bothCalled = true
			return nil
		}))

	ctx := context.Background()
	state := NewState()

	// Run the chain (configured for input type)
	err := chain.Validate(ctx, state)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check which guardrails were called
	if !inputCalled {
		t.Error("Input guardrail should have been called")
	}
	if outputCalled {
		t.Error("Output guardrail should not have been called")
	}
	if !bothCalled {
		t.Error("Both guardrail should have been called")
	}
}

func TestRequiredKeysGuardrail(t *testing.T) {
	guardrail := RequiredKeysGuardrail("required-keys", "key1", "key2", "key3")

	tests := []struct {
		name        string
		state       *State
		expectError bool
	}{
		{
			name: "all keys present",
			state: func() *State {
				s := NewState()
				s.Set("key1", "value1")
				s.Set("key2", "value2")
				s.Set("key3", "value3")
				return s
			}(),
			expectError: false,
		},
		{
			name: "missing one key",
			state: func() *State {
				s := NewState()
				s.Set("key1", "value1")
				s.Set("key3", "value3")
				return s
			}(),
			expectError: true,
		},
		{
			name:        "missing all keys",
			state:       NewState(),
			expectError: true,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := guardrail.Validate(ctx, tt.state)
			if (err != nil) != tt.expectError {
				t.Errorf("Validate() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestMaxStateSizeGuardrail(t *testing.T) {
	// Create a small state
	smallState := NewState()
	smallState.Set("key", "value")

	// Create a large state
	largeState := NewState()
	largeValue := strings.Repeat("x", 10000)
	largeState.Set("large", largeValue)

	// Small limit guardrail
	smallLimit := MaxStateSizeGuardrail("small-limit", 1000)

	// Large limit guardrail
	largeLimit := MaxStateSizeGuardrail("large-limit", 100000)

	ctx := context.Background()

	// Small state should pass both
	if err := smallLimit.Validate(ctx, smallState); err != nil {
		t.Errorf("Small state should pass small limit: %v", err)
	}
	if err := largeLimit.Validate(ctx, smallState); err != nil {
		t.Errorf("Small state should pass large limit: %v", err)
	}

	// Large state should fail small limit
	if err := smallLimit.Validate(ctx, largeState); err == nil {
		t.Error("Large state should fail small limit")
	}

	// Large state should pass large limit
	if err := largeLimit.Validate(ctx, largeState); err != nil {
		t.Errorf("Large state should pass large limit: %v", err)
	}
}

func TestMessageCountGuardrail(t *testing.T) {
	guardrail := MessageCountGuardrail("max-messages", 5)

	tests := []struct {
		name         string
		messageCount int
		expectError  bool
	}{
		{
			name:         "under limit",
			messageCount: 3,
			expectError:  false,
		},
		{
			name:         "at limit",
			messageCount: 5,
			expectError:  false,
		},
		{
			name:         "over limit",
			messageCount: 6,
			expectError:  true,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewState()
			for i := 0; i < tt.messageCount; i++ {
				state.AddMessage(NewMessage(RoleUser, fmt.Sprintf("message %d", i)))
			}

			err := guardrail.Validate(ctx, state)
			if (err != nil) != tt.expectError {
				t.Errorf("Validate() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestContentModerationGuardrail(t *testing.T) {
	prohibited := []string{"badword", "inappropriate", "forbidden"}
	guardrail := ContentModerationGuardrail("content-filter", prohibited)

	tests := []struct {
		name        string
		state       *State
		expectError bool
	}{
		{
			name: "clean content",
			state: func() *State {
				s := NewState()
				s.AddMessage(NewMessage(RoleUser, "This is clean content"))
				s.Set("description", "A nice description")
				return s
			}(),
			expectError: false,
		},
		{
			name: "prohibited in message",
			state: func() *State {
				s := NewState()
				s.AddMessage(NewMessage(RoleUser, "This contains badword content"))
				return s
			}(),
			expectError: true,
		},
		{
			name: "prohibited in state value",
			state: func() *State {
				s := NewState()
				s.Set("text", "Something inappropriate here")
				return s
			}(),
			expectError: true,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := guardrail.Validate(ctx, tt.state)
			if (err != nil) != tt.expectError {
				t.Errorf("Validate() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestGuardrailValidateAsync(t *testing.T) {
	// Create a slow guardrail
	slowGuardrail := NewGuardrailFunc("slow", GuardrailTypeBoth, func(ctx context.Context, state *State) error {
		select {
		case <-time.After(100 * time.Millisecond):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	// Create a fast guardrail
	fastGuardrail := NewGuardrailFunc("fast", GuardrailTypeBoth, func(ctx context.Context, state *State) error {
		return errors.New("fast fail")
	})

	ctx := context.Background()
	state := NewState()

	// Test timeout
	errCh := slowGuardrail.ValidateAsync(ctx, state, 50*time.Millisecond)
	err := <-errCh
	if err == nil {
		t.Error("Expected timeout error")
	}
	if !strings.Contains(err.Error(), "timeout") {
		t.Errorf("Expected timeout error, got: %v", err)
	}

	// Test successful completion
	errCh = slowGuardrail.ValidateAsync(ctx, state, 200*time.Millisecond)
	err = <-errCh
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Test fast failure
	errCh = fastGuardrail.ValidateAsync(ctx, state, 100*time.Millisecond)
	err = <-errCh
	if err == nil {
		t.Error("Expected fast fail error")
	}
	if err.Error() != "fast fail" {
		t.Errorf("Expected 'fast fail' error, got: %v", err)
	}
}

func TestGuardrailChainValidateAsync(t *testing.T) {
	chain := NewGuardrailChain("async-chain", GuardrailTypeBoth, false)

	// Add a mix of fast and slow guardrails
	chain.
		Add(NewGuardrailFunc("fast-pass", GuardrailTypeBoth, func(ctx context.Context, state *State) error {
			return nil
		})).
		Add(NewGuardrailFunc("slow", GuardrailTypeBoth, func(ctx context.Context, state *State) error {
			select {
			case <-time.After(100 * time.Millisecond):
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}))

	ctx := context.Background()
	state := NewState()

	// Test with adequate timeout
	errCh := chain.ValidateAsync(ctx, state, 200*time.Millisecond)
	err := <-errCh
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Test with short timeout
	errCh = chain.ValidateAsync(ctx, state, 50*time.Millisecond)
	err = <-errCh
	if err == nil {
		t.Error("Expected timeout error")
	}
	if !strings.Contains(err.Error(), "timeout") {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}
