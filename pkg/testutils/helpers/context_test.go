package helpers

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

func TestCreateTestToolContext(t *testing.T) {
	t.Run("default context", func(t *testing.T) {
		ctx := CreateTestToolContext()

		assert.NotNil(t, ctx)
		assert.NotEmpty(t, ctx.RunID)
		assert.NotNil(t, ctx.State)
		assert.NotNil(t, ctx.Events)
		assert.Equal(t, "test-agent", ctx.Agent.ID)
		assert.Equal(t, 0, ctx.Retry)
		assert.NotNil(t, ctx.Context)
	})

	t.Run("with options", func(t *testing.T) {
		state := domain.NewState()
		state.Set("key", "value")

		ctx := CreateTestToolContext(
			WithTestState(state),
			WithTestRunID("custom-run-id"),
			WithTestRetry(3),
		)

		assert.Equal(t, "custom-run-id", ctx.RunID)
		assert.Equal(t, 3, ctx.Retry)

		val, exists := ctx.State.Get("key")
		assert.True(t, exists)
		assert.Equal(t, "value", val)
	})

	t.Run("with timeout", func(t *testing.T) {
		ctx := CreateToolContextWithTimeout(100 * time.Millisecond)

		deadline, ok := ctx.Deadline()
		assert.True(t, ok)
		assert.WithinDuration(t, time.Now().Add(100*time.Millisecond), deadline, 10*time.Millisecond)
	})

	t.Run("with error context", func(t *testing.T) {
		ctx := CreateToolContextWithError()

		assert.NotNil(t, ctx.Context)
		assert.Equal(t, 3, ctx.Retry)

		// Context should be canceled
		select {
		case <-ctx.Done():
			// Expected
		default:
			t.Fatal("expected context to be canceled")
		}
	})
}

func TestCreateTestAgentContext(t *testing.T) {
	type TestDeps struct {
		DB     string
		Config map[string]string
	}

	t.Run("with dependencies", func(t *testing.T) {
		deps := TestDeps{
			DB:     "test-db",
			Config: map[string]string{"key": "value"},
		}

		ctx := CreateAgentContextWithDeps(deps)

		assert.NotNil(t, ctx)
		assert.NotEmpty(t, ctx.RunID)
		assert.NotNil(t, ctx.State)
		assert.Equal(t, deps, ctx.Deps())
	})

	t.Run("with state", func(t *testing.T) {
		deps := TestDeps{DB: "test-db"}
		data := map[string]interface{}{
			"input": "test",
			"count": 42,
		}

		ctx := CreateAgentContextWithState(deps, data)

		val, exists := ctx.State.Get("input")
		require.True(t, exists)
		assert.Equal(t, "test", val)

		val, exists = ctx.State.Get("count")
		require.True(t, exists)
		assert.Equal(t, 42, val)
	})

	t.Run("with timeout", func(t *testing.T) {
		deps := TestDeps{DB: "test-db"}
		ctx := CreateAgentContextWithTimeout(deps, 200*time.Millisecond)

		deadline, ok := ctx.Context().Deadline()
		assert.True(t, ok)
		assert.WithinDuration(t, time.Now().Add(200*time.Millisecond), deadline, 10*time.Millisecond)
	})
}

func TestContextOptions(t *testing.T) {
	t.Run("agent info option", func(t *testing.T) {
		info := domain.AgentInfo{
			ID:          "custom-agent",
			Name:        "Custom Agent",
			Description: "Custom test agent",
			Type:        domain.AgentTypeLLM,
			Metadata: map[string]interface{}{
				"version": "1.0",
			},
		}

		ctx := CreateTestToolContext(WithTestAgentInfo(info))

		assert.Equal(t, info.ID, ctx.Agent.ID)
		assert.Equal(t, info.Name, ctx.Agent.Name)
		assert.Equal(t, info.Description, ctx.Agent.Description)
		assert.Equal(t, info.Type, ctx.Agent.Type)
		assert.Equal(t, info.Metadata, ctx.Agent.Metadata)
	})

	t.Run("context option", func(t *testing.T) {
		type ctxKey string
		const testKey ctxKey = "test-key"
		customCtx := context.WithValue(context.Background(), testKey, "test-value")

		toolCtx := CreateTestToolContext(WithTestContext(customCtx))

		assert.Equal(t, "test-value", toolCtx.Value(testKey))
	})
}
