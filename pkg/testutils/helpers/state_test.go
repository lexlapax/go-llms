package helpers

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

func TestStateDiff(t *testing.T) {
	t.Run("no differences", func(t *testing.T) {
		state1 := domain.NewState()
		state1.Set("key", "value")
		state1.Set("number", 42)

		state2 := domain.NewState()
		state2.Set("key", "value")
		state2.Set("number", 42)

		diff := DiffStates(state1, state2)
		assert.True(t, diff.IsEmpty())
		assert.Equal(t, "No differences", diff.String())
	})

	t.Run("added values", func(t *testing.T) {
		state1 := domain.NewState()
		state1.Set("existing", "value")

		state2 := domain.NewState()
		state2.Set("existing", "value")
		state2.Set("new1", "value1")
		state2.Set("new2", "value2")

		diff := DiffStates(state1, state2)
		assert.False(t, diff.IsEmpty())
		assert.Len(t, diff.Added, 2)
		assert.Equal(t, "value1", diff.Added["new1"])
		assert.Equal(t, "value2", diff.Added["new2"])

		diffStr := diff.String()
		assert.Contains(t, diffStr, "Added:")
		assert.Contains(t, diffStr, "+ new1: value1")
		assert.Contains(t, diffStr, "+ new2: value2")
	})

	t.Run("modified values", func(t *testing.T) {
		state1 := domain.NewState()
		state1.Set("key", "old value")
		state1.Set("number", 42)

		state2 := domain.NewState()
		state2.Set("key", "new value")
		state2.Set("number", 100)

		diff := DiffStates(state1, state2)
		assert.False(t, diff.IsEmpty())
		assert.Len(t, diff.Modified, 2)
		assert.Equal(t, "old value", diff.Modified["key"].Old)
		assert.Equal(t, "new value", diff.Modified["key"].New)
		assert.Equal(t, 42, diff.Modified["number"].Old)
		assert.Equal(t, 100, diff.Modified["number"].New)

		diffStr := diff.String()
		assert.Contains(t, diffStr, "Modified:")
		assert.Contains(t, diffStr, "~ key: old value -> new value")
		assert.Contains(t, diffStr, "~ number: 42 -> 100")
	})

	t.Run("removed values", func(t *testing.T) {
		state1 := domain.NewState()
		state1.Set("keep", "value")
		state1.Set("remove1", "value1")
		state1.Set("remove2", "value2")

		state2 := domain.NewState()
		state2.Set("keep", "value")

		diff := DiffStates(state1, state2)
		assert.False(t, diff.IsEmpty())
		assert.Len(t, diff.Removed, 2)
		assert.Contains(t, diff.Removed, "remove1")
		assert.Contains(t, diff.Removed, "remove2")

		diffStr := diff.String()
		assert.Contains(t, diffStr, "Removed:")
		assert.Contains(t, diffStr, "- remove1")
		assert.Contains(t, diffStr, "- remove2")
	})

	t.Run("mixed changes", func(t *testing.T) {
		state1 := domain.NewState()
		state1.Set("keep", "value")
		state1.Set("modify", "old")
		state1.Set("remove", "gone")

		state2 := domain.NewState()
		state2.Set("keep", "value")
		state2.Set("modify", "new")
		state2.Set("add", "added")

		diff := DiffStates(state1, state2)
		assert.False(t, diff.IsEmpty())
		assert.Len(t, diff.Added, 1)
		assert.Len(t, diff.Modified, 1)
		assert.Len(t, diff.Removed, 1)
	})
}

func TestStateSnapshot(t *testing.T) {
	t.Run("capture snapshot", func(t *testing.T) {
		state := domain.NewState()
		state.Set("key", "value")
		state.Set("number", 42)

		artifact := domain.NewArtifact("art1", domain.ArtifactTypeDocument, []byte("content"))
		state.AddArtifact(artifact)

		message := domain.NewMessage(domain.RoleUser, "Hello")
		state.AddMessage(message)

		state.SetMetadata("meta", "data")

		snapshot := CaptureSnapshot(state)

		assert.Len(t, snapshot.Values, 2)
		assert.Equal(t, "value", snapshot.Values["key"])
		assert.Equal(t, 42, snapshot.Values["number"])

		assert.Len(t, snapshot.Artifacts, 1)
		// Note: artifact ID will be generated, so we can't check for "art1"
		for _, art := range snapshot.Artifacts {
			assert.NotNil(t, art)
			assert.Equal(t, domain.ArtifactTypeDocument, art.Type)
		}

		assert.Len(t, snapshot.Messages, 1)
		assert.Equal(t, domain.RoleUser, snapshot.Messages[0].Role)
	})

	t.Run("compare snapshots", func(t *testing.T) {
		// Create first snapshot
		state1 := domain.NewState()
		state1.Set("key", "value1")
		snapshot1 := CaptureSnapshot(state1)

		// Create second snapshot with changes
		state2 := domain.NewState()
		state2.Set("key", "value2")
		state2.Set("new", "value")
		state2.AddMessage(domain.NewMessage(domain.RoleAssistant, "Response"))
		snapshot2 := CaptureSnapshot(state2)

		comparison := CompareSnapshots(snapshot1, snapshot2)

		assert.NotNil(t, comparison.ValuesDiff)
		assert.Len(t, comparison.ValuesDiff.Added, 1)
		assert.Len(t, comparison.ValuesDiff.Modified, 1)

		assert.NotNil(t, comparison.MessagesDiff)
		assert.Equal(t, 1, comparison.MessagesDiff.Added)
		assert.Equal(t, 1, comparison.MessagesDiff.Total)
	})
}

func TestStateMutator(t *testing.T) {
	t.Run("set operations", func(t *testing.T) {
		state := domain.NewState()

		MutateState(state).
			Set("key1", "value1").
			Set("key2", "value2").
			SetMultiple(map[string]interface{}{
				"key3": "value3",
				"key4": 4,
			})

		val, exists := state.Get("key1")
		assert.True(t, exists)
		assert.Equal(t, "value1", val)

		val, exists = state.Get("key4")
		assert.True(t, exists)
		assert.Equal(t, 4, val)
	})

	t.Run("delete operations", func(t *testing.T) {
		state := domain.NewState()
		state.Set("key1", "value1")
		state.Set("key2", "value2")

		MutateState(state).Delete("key1")

		_, exists := state.Get("key1")
		assert.False(t, exists)

		_, exists = state.Get("key2")
		assert.True(t, exists)
	})

	t.Run("artifact operations", func(t *testing.T) {
		state := domain.NewState()

		artifact := domain.NewArtifact("art1", domain.ArtifactTypeDocument, []byte("content"))

		MutateState(state).AddArtifact(artifact)

		// Note: artifact ID will be generated, so we need to get it from the artifact
		retrieved, exists := state.GetArtifact(artifact.ID)
		assert.True(t, exists)
		assert.Equal(t, artifact.ID, retrieved.ID)
		assert.Equal(t, artifact.Type, retrieved.Type)
	})

	t.Run("message operations", func(t *testing.T) {
		state := domain.NewState()

		MutateState(state).
			AddMessage(domain.RoleUser, "Hello").
			AddMessage(domain.RoleAssistant, "Hi there!")

		messages := state.Messages()
		assert.Len(t, messages, 2)
		assert.Equal(t, domain.RoleUser, messages[0].Role)
		assert.Equal(t, "Hello", messages[0].Content)
	})

	t.Run("metadata operations", func(t *testing.T) {
		state := domain.NewState()

		MutateState(state).
			SetMetadata("version", "1.0").
			SetMetadata("author", "test")

		val, exists := state.GetMetadata("version")
		assert.True(t, exists)
		assert.Equal(t, "1.0", val)
	})

	t.Run("clear operations", func(t *testing.T) {
		state := domain.NewState()
		state.Set("key1", "value1")
		state.Set("key2", "value2")

		MutateState(state).Clear()

		assert.Len(t, state.Keys(), 0)
	})

	t.Run("chained operations", func(t *testing.T) {
		state := domain.NewState()

		finalState := MutateState(state).
			Set("name", "test").
			Set("count", 10).
			AddMessage(domain.RoleUser, "Start").
			SetMetadata("timestamp", "2024-01-01").
			Done()

		assert.Equal(t, state, finalState)

		val, exists := finalState.Get("name")
		assert.True(t, exists)
		assert.Equal(t, "test", val)

		assert.Len(t, finalState.Messages(), 1)
	})
}

func TestStateValidator(t *testing.T) {
	t.Run("has key", func(t *testing.T) {
		state := domain.NewState()
		state.Set("key1", "value1")
		state.Set("key2", "value2")

		sv := ValidateState(state).HasKey("key1")
		assert.True(t, sv.IsValid())

		sv2 := ValidateState(state).HasKey("missing")
		assert.False(t, sv2.IsValid())
		assert.Contains(t, sv2.String(), "missing required key: missing")
	})

	t.Run("has keys", func(t *testing.T) {
		state := domain.NewState()
		state.Set("key1", "value1")
		state.Set("key2", "value2")

		sv := ValidateState(state).HasKeys("key1", "key2")
		assert.True(t, sv.IsValid())

		sv2 := ValidateState(state).HasKeys("key1", "key2", "key3")
		assert.False(t, sv2.IsValid())
		assert.Len(t, sv2.GetErrors(), 1)
	})

	t.Run("has value", func(t *testing.T) {
		state := domain.NewState()
		state.Set("name", "test")
		state.Set("count", 42)

		sv := ValidateState(state).
			HasValue("name", "test").
			HasValue("count", 42)
		assert.True(t, sv.IsValid())

		sv2 := ValidateState(state).HasValue("name", "wrong")
		assert.False(t, sv2.IsValid())
		assert.Contains(t, sv2.String(), "expected wrong, got test")
	})

	t.Run("has type", func(t *testing.T) {
		state := domain.NewState()
		state.Set("string", "value")
		state.Set("number", 42)
		state.Set("bool", true)

		sv := ValidateState(state).
			HasType("string", reflect.TypeOf("")).
			HasType("number", reflect.TypeOf(0)).
			HasType("bool", reflect.TypeOf(true))
		assert.True(t, sv.IsValid())

		sv2 := ValidateState(state).HasType("string", reflect.TypeOf(0))
		assert.False(t, sv2.IsValid())
		assert.Contains(t, sv2.String(), "expected type int, got string")
	})

	t.Run("has artifact", func(t *testing.T) {
		state := domain.NewState()
		artifact := domain.NewArtifact("art1", domain.ArtifactTypeDocument, []byte("content"))
		state.AddArtifact(artifact)

		sv := ValidateState(state).HasArtifact(artifact.ID)
		assert.True(t, sv.IsValid())

		sv2 := ValidateState(state).HasArtifact("missing")
		assert.False(t, sv2.IsValid())
		assert.Contains(t, sv2.String(), "missing artifact: missing")
	})

	t.Run("has message count", func(t *testing.T) {
		state := domain.NewState()
		state.AddMessage(domain.NewMessage(domain.RoleUser, "Hello"))
		state.AddMessage(domain.NewMessage(domain.RoleAssistant, "Hi"))

		sv := ValidateState(state).HasMessageCount(2)
		assert.True(t, sv.IsValid())

		sv2 := ValidateState(state).HasMessageCount(3)
		assert.False(t, sv2.IsValid())
		assert.Contains(t, sv2.String(), "expected 3 messages, got 2")
	})

	t.Run("chained validations", func(t *testing.T) {
		state := domain.NewState()
		state.Set("name", "test")
		state.Set("count", 42)
		state.AddMessage(domain.NewMessage(domain.RoleUser, "Hello"))

		sv := ValidateState(state).
			HasKeys("name", "count").
			HasValue("name", "test").
			HasType("count", reflect.TypeOf(0)).
			HasMessageCount(1)

		assert.True(t, sv.IsValid())
		assert.Equal(t, "State validation passed", sv.String())
	})
}

func TestStateHelpers(t *testing.T) {
	t.Run("create state with data", func(t *testing.T) {
		data := map[string]interface{}{
			"key1": "value1",
			"key2": 42,
			"key3": true,
		}

		state := CreateStateWithData(data)

		for k, v := range data {
			val, exists := state.Get(k)
			require.True(t, exists)
			assert.Equal(t, v, val)
		}
	})

	t.Run("create state with messages", func(t *testing.T) {
		messages := []domain.Message{
			domain.NewMessage(domain.RoleUser, "Question"),
			domain.NewMessage(domain.RoleAssistant, "Answer"),
			domain.NewMessage(domain.RoleUser, "Follow-up"),
		}

		state := CreateStateWithMessages(messages...)

		stateMessages := state.Messages()
		assert.Len(t, stateMessages, 3)
		for i, msg := range stateMessages {
			assert.Equal(t, messages[i].Role, msg.Role)
			assert.Equal(t, messages[i].Content, msg.Content)
		}
	})

	t.Run("create state with artifacts", func(t *testing.T) {
		artifacts := []*domain.Artifact{
			domain.NewArtifact("art1", domain.ArtifactTypeDocument, []byte("content1")),
			domain.NewArtifact("art2", domain.ArtifactTypeData, []byte(`{"key": "value"}`)),
		}

		state := CreateStateWithArtifacts(artifacts...)

		for _, art := range artifacts {
			retrieved, exists := state.GetArtifact(art.ID)
			require.True(t, exists)
			assert.Equal(t, art.ID, retrieved.ID)
			assert.Equal(t, art.Type, retrieved.Type)
			assert.Equal(t, art.Name, retrieved.Name)
		}
	})
}
