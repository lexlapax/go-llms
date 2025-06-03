// ABOUTME: Tests for state transformation functions including filter, map, and update operations
// ABOUTME: covering message manipulation, metadata transforms, and chainable operations

package core

import (
	"strings"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

func TestFilterMessages(t *testing.T) {
	state := domain.NewState()
	state.AddMessage(domain.NewMessage(domain.RoleUser, "Hello"))
	state.AddMessage(domain.NewMessage(domain.RoleAssistant, "Hi there"))
	state.AddMessage(domain.NewMessage(domain.RoleSystem, "System prompt"))
	state.AddMessage(domain.NewMessage(domain.RoleUser, "How are you?"))

	// Filter only user messages
	transformed := FilterMessages(state, func(msg domain.Message) bool {
		return msg.Role == domain.RoleUser
	})

	messages := transformed.Messages()
	if len(messages) != 2 {
		t.Errorf("Expected 2 user messages, got %d", len(messages))
	}

	for _, msg := range messages {
		if msg.Role != domain.RoleUser {
			t.Errorf("Expected only user messages, got role: %s", msg.Role)
		}
	}
}

func TestMapMessages(t *testing.T) {
	state := domain.NewState()
	state.AddMessage(domain.NewMessage(domain.RoleUser, "hello world"))
	state.AddMessage(domain.NewMessage(domain.RoleAssistant, "hi there"))

	// Transform messages to uppercase
	transformed := MapMessages(state, func(msg domain.Message) domain.Message {
		newMsg := msg
		newMsg.Content = strings.ToUpper(msg.Content)
		return newMsg
	})

	messages := transformed.Messages()
	expected := []string{"HELLO WORLD", "HI THERE"}

	for i, msg := range messages {
		if msg.Content != expected[i] {
			t.Errorf("Message %d: expected %s, got %s", i, expected[i], msg.Content)
		}
	}
}

func TestFilterMetadata(t *testing.T) {
	state := domain.NewState()
	state.Set("keep1", "value1")
	state.Set("remove1", "value2")
	state.Set("keep2", "value3")
	state.Set("remove2", "value4")

	// Keep only keys that start with "keep"
	transformed := FilterMetadata(state, func(key string, value interface{}) bool {
		return strings.HasPrefix(key, "keep")
	})

	// Check kept keys
	if val, ok := transformed.Get("keep1"); !ok || val != "value1" {
		t.Error("Expected keep1 to be preserved")
	}
	if val, ok := transformed.Get("keep2"); !ok || val != "value3" {
		t.Error("Expected keep2 to be preserved")
	}

	// Check removed keys
	if _, ok := transformed.Get("remove1"); ok {
		t.Error("Expected remove1 to be filtered out")
	}
	if _, ok := transformed.Get("remove2"); ok {
		t.Error("Expected remove2 to be filtered out")
	}
}

func TestMapMetadata(t *testing.T) {
	state := domain.NewState()
	state.Set("count", 5)
	state.Set("name", "test")
	state.Set("active", true)

	// Transform metadata values
	transformed := MapMetadata(state, func(key string, value interface{}) interface{} {
		switch v := value.(type) {
		case int:
			return v * 2
		case string:
			return strings.ToUpper(v)
		case bool:
			return !v
		default:
			return value
		}
	})

	// Check transformations
	if val, _ := transformed.Get("count"); val != 10 {
		t.Errorf("Expected count to be 10, got %v", val)
	}
	if val, _ := transformed.Get("name"); val != "TEST" {
		t.Errorf("Expected name to be TEST, got %v", val)
	}
	if val, _ := transformed.Get("active"); val != false {
		t.Errorf("Expected active to be false, got %v", val)
	}
}

func TestUpdateMetadata(t *testing.T) {
	state := domain.NewState()
	state.Set("existing", "old")
	state.Set("keep", "value")

	updates := map[string]interface{}{
		"existing": "new",
		"added":    "fresh",
	}

	transformed := UpdateMetadata(state, updates)

	// Check updates
	if val, _ := transformed.Get("existing"); val != "new" {
		t.Errorf("Expected existing to be updated to 'new', got %v", val)
	}
	if val, _ := transformed.Get("added"); val != "fresh" {
		t.Errorf("Expected added to be 'fresh', got %v", val)
	}
	if val, _ := transformed.Get("keep"); val != "value" {
		t.Errorf("Expected keep to remain 'value', got %v", val)
	}
}

func TestRemoveMetadataKeys(t *testing.T) {
	state := domain.NewState()
	state.Set("keep1", "value1")
	state.Set("remove1", "value2")
	state.Set("keep2", "value3")
	state.Set("remove2", "value4")

	transformed := RemoveMetadataKeys(state, "remove1", "remove2", "nonexistent")

	// Check kept keys
	if _, ok := transformed.Get("keep1"); !ok {
		t.Error("Expected keep1 to be preserved")
	}
	if _, ok := transformed.Get("keep2"); !ok {
		t.Error("Expected keep2 to be preserved")
	}

	// Check removed keys
	if _, ok := transformed.Get("remove1"); ok {
		t.Error("Expected remove1 to be removed")
	}
	if _, ok := transformed.Get("remove2"); ok {
		t.Error("Expected remove2 to be removed")
	}
}

func TestTruncateMessages(t *testing.T) {
	state := domain.NewState()
	for i := 0; i < 10; i++ {
		state.AddMessage(domain.NewMessage(domain.RoleUser, "Message"))
	}

	// Keep only last 3 messages
	transformed := TruncateMessages(state, 3)

	messages := transformed.Messages()
	if len(messages) != 3 {
		t.Errorf("Expected 3 messages after truncation, got %d", len(messages))
	}
}

func TestSortMessages(t *testing.T) {
	state := domain.NewState()

	// Add messages with timestamps
	now := time.Now()
	msg1 := domain.NewMessage(domain.RoleUser, "First")
	msg1.Timestamp = now.Add(-3 * time.Minute)

	msg2 := domain.NewMessage(domain.RoleAssistant, "Second")
	msg2.Timestamp = now.Add(-1 * time.Minute)

	msg3 := domain.NewMessage(domain.RoleUser, "Third")
	msg3.Timestamp = now.Add(-2 * time.Minute)

	state.AddMessage(msg1)
	state.AddMessage(msg2)
	state.AddMessage(msg3)

	// Sort by timestamp
	transformed := SortMessages(state, func(a, b domain.Message) bool {
		return a.Timestamp.Before(b.Timestamp)
	})

	messages := transformed.Messages()
	expectedOrder := []string{"First", "Third", "Second"}

	for i, msg := range messages {
		if msg.Content != expectedOrder[i] {
			t.Errorf("Message %d: expected %s, got %s", i, expectedOrder[i], msg.Content)
		}
	}
}

func TestAddMessagePrefix(t *testing.T) {
	state := domain.NewState()
	state.AddMessage(domain.NewMessage(domain.RoleUser, "Hello"))
	state.AddMessage(domain.NewMessage(domain.RoleAssistant, "Hi"))

	transformed := AddMessagePrefix(state, "[ARCHIVED] ")

	messages := transformed.Messages()
	for i, msg := range messages {
		if !strings.HasPrefix(msg.Content, "[ARCHIVED] ") {
			t.Errorf("Message %d missing prefix: %s", i, msg.Content)
		}
	}
}

func TestAddMessageSuffix(t *testing.T) {
	state := domain.NewState()
	state.AddMessage(domain.NewMessage(domain.RoleUser, "Hello"))
	state.AddMessage(domain.NewMessage(domain.RoleAssistant, "Hi"))

	transformed := AddMessageSuffix(state, " [END]")

	messages := transformed.Messages()
	for i, msg := range messages {
		if !strings.HasSuffix(msg.Content, " [END]") {
			t.Errorf("Message %d missing suffix: %s", i, msg.Content)
		}
	}
}

func TestMergeStates(t *testing.T) {
	state1 := domain.NewState()
	state1.SetMetadata("key1", "value1")
	state1.SetMetadata("shared", "state1")
	state1.AddMessage(domain.NewMessage(domain.RoleUser, "From state1"))

	state2 := domain.NewState()
	state2.SetMetadata("key2", "value2")
	state2.SetMetadata("shared", "state2")
	state2.AddMessage(domain.NewMessage(domain.RoleAssistant, "From state2"))

	// Merge with state2 taking precedence
	merged := MergeStates(state1, state2)

	// Check metadata merge
	if val, _ := merged.GetMetadata("key1"); val != "value1" {
		t.Errorf("Expected key1 from state1, got %v", val)
	}
	if val, _ := merged.GetMetadata("key2"); val != "value2" {
		t.Errorf("Expected key2 from state2, got %v", val)
	}
	if val, _ := merged.GetMetadata("shared"); val != "state2" {
		t.Errorf("Expected shared key from state2 (precedence), got %v", val)
	}

	// Check messages merge
	messages := merged.Messages()
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages after merge, got %d", len(messages))
	}
}

func TestCloneWithModifications(t *testing.T) {
	original := domain.NewState()
	original.SetMetadata("key1", "value1")
	original.SetMetadata("key2", "value2")
	original.AddMessage(domain.NewMessage(domain.RoleUser, "Hello"))

	modified := CloneWithModifications(original, func(s *domain.State) {
		s.SetMetadata("key2", "modified")
		s.SetMetadata("key3", "added")
		s.AddMessage(domain.NewMessage(domain.RoleAssistant, "Response"))
	})

	// Check original is unchanged
	if val, _ := original.GetMetadata("key2"); val != "value2" {
		t.Error("Original state should not be modified")
	}
	if _, ok := original.GetMetadata("key3"); ok {
		t.Error("Original state should not have key3")
	}
	if len(original.Messages()) != 1 {
		t.Error("Original state should still have 1 message")
	}

	// Check modifications
	if val, _ := modified.GetMetadata("key2"); val != "modified" {
		t.Errorf("Expected key2 to be modified, got %v", val)
	}
	if val, _ := modified.GetMetadata("key3"); val != "added" {
		t.Errorf("Expected key3 to be added, got %v", val)
	}
	if len(modified.Messages()) != 2 {
		t.Error("Modified state should have 2 messages")
	}
}

func TestConditionalTransform(t *testing.T) {
	// Transform that only applies if condition is met
	transform := ConditionalUtilityTransform(
		func(s *domain.State) bool {
			count, ok := s.Get("message_count")
			return ok && count.(int) > 5
		},
		func(s *domain.State) *domain.State {
			return AddMessagePrefix(s, "[ARCHIVED] ")
		},
	)

	// State with few messages (condition not met)
	state1 := domain.NewState()
	state1.Set("message_count", 3)
	state1.AddMessage(domain.NewMessage(domain.RoleUser, "Hello"))

	result1 := transform(state1)
	if strings.HasPrefix(result1.Messages()[0].Content, "[ARCHIVED]") {
		t.Error("Transform should not apply when condition is not met")
	}

	// State with many messages (condition met)
	state2 := domain.NewState()
	state2.Set("message_count", 10)
	state2.AddMessage(domain.NewMessage(domain.RoleUser, "Hello"))

	result2 := transform(state2)
	if !strings.HasPrefix(result2.Messages()[0].Content, "[ARCHIVED]") {
		t.Error("Transform should apply when condition is met")
	}
}

func TestChainTransforms(t *testing.T) {
	state := domain.NewState()
	state.Set("count", 5)
	state.AddMessage(domain.NewMessage(domain.RoleUser, "hello"))
	state.AddMessage(domain.NewMessage(domain.RoleAssistant, "world"))

	// Chain multiple transforms
	transformed := ChainUtilityTransforms(
		// 1. Double the count
		func(s *domain.State) *domain.State {
			return MapMetadata(s, func(k string, v interface{}) interface{} {
				if k == "count" && v != nil {
					return v.(int) * 2
				}
				return v
			})
		},
		// 2. Uppercase messages
		func(s *domain.State) *domain.State {
			return MapMessages(s, func(m domain.Message) domain.Message {
				newMsg := m
				newMsg.Content = strings.ToUpper(m.Content)
				return newMsg
			})
		},
		// 3. Add prefix
		func(s *domain.State) *domain.State {
			return AddMessagePrefix(s, "[PROCESSED] ")
		},
	)(state)

	// Verify all transforms applied
	count, _ := transformed.Get("count")
	if count != 10 {
		t.Errorf("Expected count to be 10, got %v", count)
	}

	messages := transformed.Messages()
	expected := []string{"[PROCESSED] HELLO", "[PROCESSED] WORLD"}
	for i, msg := range messages {
		if msg.Content != expected[i] {
			t.Errorf("Message %d: expected %s, got %s", i, expected[i], msg.Content)
		}
	}
}

func TestWithTimestamp(t *testing.T) {
	state := domain.NewState()

	// Add timestamp
	transformed := WithTimestamp(state)

	timestamp, ok := transformed.Get("timestamp")
	if !ok {
		t.Error("Expected timestamp to be added")
	}

	ts, ok := timestamp.(time.Time)
	if !ok {
		t.Error("Expected timestamp to be time.Time")
	}

	// Verify timestamp is recent
	if time.Since(ts) > 1*time.Second {
		t.Error("Timestamp should be recent")
	}
}

func TestWithMessageCount(t *testing.T) {
	state := domain.NewState()
	state.AddMessage(domain.NewMessage(domain.RoleUser, "1"))
	state.AddMessage(domain.NewMessage(domain.RoleAssistant, "2"))
	state.AddMessage(domain.NewMessage(domain.RoleUser, "3"))

	transformed := WithMessageCount(state)

	count, ok := transformed.Get("message_count")
	if !ok {
		t.Error("Expected message_count to be added")
	}

	if count != 3 {
		t.Errorf("Expected message_count to be 3, got %v", count)
	}
}

func TestComplexTransformScenario(t *testing.T) {
	// Simulate a conversation archiving scenario
	state := domain.NewState()
	state.Set("user_id", "user123")
	state.Set("session_start", time.Now().Add(-1*time.Hour))

	// Add conversation messages
	for i := 0; i < 20; i++ {
		if i%2 == 0 {
			state.AddMessage(domain.NewMessage(domain.RoleUser, "User message"))
		} else {
			state.AddMessage(domain.NewMessage(domain.RoleAssistant, "Assistant response"))
		}
	}

	// Complex transform pipeline
	archived := ChainUtilityTransforms(
		// 1. Add timestamp
		WithTimestamp,

		// 2. Keep only last 10 messages
		func(s *domain.State) *domain.State {
			return TruncateMessages(s, 10)
		},

		// 3. Add archive prefix to messages
		func(s *domain.State) *domain.State {
			return AddMessagePrefix(s, "[ARCHIVED] ")
		},

		// 4. Add message count after truncation
		WithMessageCount,

		// 5. Add archive metadata
		func(s *domain.State) *domain.State {
			return UpdateMetadata(s, map[string]interface{}{
				"archived":       true,
				"archive_date":   time.Now(),
				"original_count": 20,
			})
		},

		// 6. Remove sensitive data
		func(s *domain.State) *domain.State {
			return RemoveMetadataKeys(s, "session_start")
		},
	)(state)

	// Verify transformations
	if count, _ := archived.Get("message_count"); count != 10 {
		t.Errorf("Expected 10 messages after truncation, got %v", count)
	}

	if archived, _ := archived.Get("archived"); archived != true {
		t.Error("Expected archived flag to be true")
	}

	if _, ok := archived.Get("session_start"); ok {
		t.Error("Expected session_start to be removed")
	}

	messages := archived.Messages()
	for _, msg := range messages {
		if !strings.HasPrefix(msg.Content, "[ARCHIVED] ") {
			t.Error("Expected all messages to have archive prefix")
		}
	}
}
