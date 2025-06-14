package workflow

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScriptHandlerRegistry(t *testing.T) {
	// Clear any existing handlers
	globalScriptHandlers.mu.Lock()
	globalScriptHandlers.handlers = make(map[string]ScriptHandler)
	globalScriptHandlers.mu.Unlock()

	t.Run("RegisterScriptHandler", func(t *testing.T) {
		handler := &MockJavaScriptHandler{}
		err := RegisterScriptHandler("test-js", handler)
		assert.NoError(t, err)

		// Verify registration
		h, exists := GetScriptHandler("test-js")
		assert.True(t, exists)
		assert.Equal(t, handler, h)
	})

	t.Run("RegisterScriptHandler_EmptyLanguage", func(t *testing.T) {
		err := RegisterScriptHandler("", &MockJavaScriptHandler{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "language cannot be empty")
	})

	t.Run("RegisterScriptHandler_NilHandler", func(t *testing.T) {
		err := RegisterScriptHandler("test", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "handler cannot be nil")
	})

	t.Run("UnregisterScriptHandler", func(t *testing.T) {
		// Register first
		handler := &MockJavaScriptHandler{}
		err := RegisterScriptHandler("test-remove", handler)
		require.NoError(t, err)

		// Verify it exists
		_, exists := GetScriptHandler("test-remove")
		assert.True(t, exists)

		// Unregister
		UnregisterScriptHandler("test-remove")

		// Verify it's gone
		_, exists = GetScriptHandler("test-remove")
		assert.False(t, exists)
	})

	t.Run("ListScriptLanguages", func(t *testing.T) {
		// Clear and add known handlers
		globalScriptHandlers.mu.Lock()
		globalScriptHandlers.handlers = make(map[string]ScriptHandler)
		globalScriptHandlers.mu.Unlock()

		err := RegisterScriptHandler("lang1", &MockJavaScriptHandler{})
		require.NoError(t, err)
		err = RegisterScriptHandler("lang2", &ExpressionHandler{})
		require.NoError(t, err)

		languages := ListScriptLanguages()
		assert.Len(t, languages, 2)
		assert.Contains(t, languages, "lang1")
		assert.Contains(t, languages, "lang2")
	})
}

func TestScriptStep(t *testing.T) {
	// Register test handlers
	err := RegisterDefaultHandlers()
	require.NoError(t, err)

	t.Run("NewScriptStep", func(t *testing.T) {
		step, err := NewScriptStep("test-step", "javascript", "return 'success'")
		assert.NoError(t, err)
		assert.NotNil(t, step)
		assert.Equal(t, "test-step", step.Name())
		assert.Equal(t, "javascript", step.Language())
		assert.Equal(t, "return 'success'", step.Script())
		assert.Equal(t, 30*time.Second, step.Timeout())
	})

	t.Run("NewScriptStep_UnknownLanguage", func(t *testing.T) {
		step, err := NewScriptStep("test-step", "unknown-lang", "code")
		assert.Error(t, err)
		assert.Nil(t, step)
		assert.Contains(t, err.Error(), "no handler registered")
	})

	t.Run("ScriptStep_Execute", func(t *testing.T) {
		step, err := NewScriptStep("test-exec", "javascript", "return 'success'")
		require.NoError(t, err)

		// Create initial state
		state := &WorkflowState{
			State:    nil,
			Metadata: make(map[string]interface{}),
		}

		// Execute
		ctx := context.Background()
		result, err := step.Execute(ctx, state)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.State)

		// Check metadata was added
		assert.Contains(t, result.Metadata, "script_test-exec_executed")
		assert.Contains(t, result.Metadata, "script_test-exec_language")
		assert.Equal(t, "javascript", result.Metadata["script_test-exec_language"])
	})

	t.Run("ScriptStep_Validate", func(t *testing.T) {
		step, err := NewScriptStep("test-validate", "expr", "result = 'done'")
		require.NoError(t, err)

		err = step.Validate()
		assert.NoError(t, err)
	})

	t.Run("ScriptStep_Properties", func(t *testing.T) {
		builder := NewScriptStepBuilder("test-props").
			WithLanguage("javascript").
			WithScript("return true").
			WithDescription("Test description").
			WithTimeout(60*time.Second).
			WithEnvironment("key1", "value1").
			WithMetadata("version", "1.0")

		step, err := builder.Build()
		require.NoError(t, err)

		assert.Equal(t, "test-props", step.Name())
		assert.Equal(t, "javascript", step.Language())
		assert.Equal(t, "return true", step.Script())
		assert.Equal(t, "Test description", step.Description())
		assert.Equal(t, 60*time.Second, step.Timeout())

		env := step.Environment()
		assert.Equal(t, "value1", env["key1"])

		meta := step.Metadata()
		assert.Equal(t, "1.0", meta["version"])
	})
}

func TestScriptStepBuilder(t *testing.T) {
	// Register handlers
	err := RegisterDefaultHandlers()
	require.NoError(t, err)

	t.Run("BuilderSuccess", func(t *testing.T) {
		step, err := NewScriptStepBuilder("builder-test").
			WithLanguage("expr").
			WithScript("x = 10").
			WithDescription("Builder test").
			WithTimeout(5*time.Second).
			WithEnvironment("env1", "val1").
			WithEnvironment("env2", 42).
			WithMetadata("tag", "test").
			Build()

		assert.NoError(t, err)
		assert.NotNil(t, step)
		assert.Equal(t, "builder-test", step.Name())
		assert.Equal(t, "expr", step.Language())
		assert.Equal(t, "x = 10", step.Script())
		assert.Equal(t, "Builder test", step.Description())
		assert.Equal(t, 5*time.Second, step.Timeout())
	})

	t.Run("BuilderMissingLanguage", func(t *testing.T) {
		step, err := NewScriptStepBuilder("no-lang").
			WithScript("code").
			Build()

		assert.Error(t, err)
		assert.Nil(t, step)
		assert.Contains(t, err.Error(), "language is required")
	})

	t.Run("BuilderMissingScript", func(t *testing.T) {
		step, err := NewScriptStepBuilder("no-script").
			WithLanguage("javascript").
			Build()

		assert.Error(t, err)
		assert.Nil(t, step)
		assert.Contains(t, err.Error(), "script is required")
	})

	t.Run("BuilderUnknownLanguage", func(t *testing.T) {
		step, err := NewScriptStepBuilder("unknown").
			WithLanguage("unknown-lang").
			WithScript("code").
			Build()

		assert.Error(t, err)
		assert.Nil(t, step)
		assert.Contains(t, err.Error(), "no handler registered")
	})

	t.Run("BuilderInvalidScript", func(t *testing.T) {
		// Register a handler that validates scripts
		validatingHandler := &ScriptHandlerFunc{
			LanguageFn: func() string { return "validating" },
			ValidateFn: func(script string) error {
				if script == "invalid" {
					return assert.AnError
				}
				return nil
			},
		}
		err := RegisterScriptHandler("validating", validatingHandler)
		require.NoError(t, err)

		step, err := NewScriptStepBuilder("invalid-script").
			WithLanguage("validating").
			WithScript("invalid").
			Build()

		assert.Error(t, err)
		assert.Nil(t, step)
		assert.Contains(t, err.Error(), "script validation failed")
	})
}

func TestScriptHandlerFunc(t *testing.T) {
	t.Run("Execute", func(t *testing.T) {
		called := false
		handler := &ScriptHandlerFunc{
			ExecuteFn: func(ctx context.Context, state *WorkflowState, script string, env map[string]interface{}) (*WorkflowState, error) {
				called = true
				return state, nil
			},
		}

		ctx := context.Background()
		state := &WorkflowState{}
		_, err := handler.Execute(ctx, state, "script", nil)
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("Execute_NotImplemented", func(t *testing.T) {
		handler := &ScriptHandlerFunc{}
		ctx := context.Background()
		state := &WorkflowState{}
		_, err := handler.Execute(ctx, state, "script", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not implemented")
	})

	t.Run("Language", func(t *testing.T) {
		handler := &ScriptHandlerFunc{
			LanguageFn: func() string { return "test-lang" },
		}
		assert.Equal(t, "test-lang", handler.Language())
	})

	t.Run("Language_NotImplemented", func(t *testing.T) {
		handler := &ScriptHandlerFunc{}
		assert.Equal(t, "", handler.Language())
	})

	t.Run("Validate", func(t *testing.T) {
		handler := &ScriptHandlerFunc{
			ValidateFn: func(script string) error {
				if script == "" {
					return assert.AnError
				}
				return nil
			},
		}

		err := handler.Validate("valid")
		assert.NoError(t, err)

		err = handler.Validate("")
		assert.Error(t, err)
	})

	t.Run("Validate_NotImplemented", func(t *testing.T) {
		handler := &ScriptHandlerFunc{}
		err := handler.Validate("script")
		assert.NoError(t, err)
	})
}
