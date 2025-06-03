// ABOUTME: Tests for the StateValidator interface with built-in validators
// ABOUTME: including composable validation patterns and specialized validators

package domain

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestStateValidator_RequiredKeysValidator(t *testing.T) {
	validator := RequiredKeysValidator("name", "age", "email")

	tests := []struct {
		name        string
		state       *State
		expectError bool
	}{
		{
			name: "all required keys present",
			state: func() *State {
				s := NewState()
				s.Set("name", "John")
				s.Set("age", 30)
				s.Set("email", "john@example.com")
				return s
			}(),
			expectError: false,
		},
		{
			name: "missing required key",
			state: func() *State {
				s := NewState()
				s.Set("name", "John")
				s.Set("age", 30)
				// missing email
				return s
			}(),
			expectError: true,
		},
		{
			name:        "empty state",
			state:       NewState(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.state)
			if (err != nil) != tt.expectError {
				t.Errorf("Validate() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestStateValidator_MaxMessageCountValidator(t *testing.T) {
	validator := MaxMessageCountValidator(3)

	tests := []struct {
		name        string
		state       *State
		expectError bool
	}{
		{
			name: "within limit",
			state: func() *State {
				s := NewState()
				s.AddMessage(NewMessage(RoleUser, "Hello"))
				s.AddMessage(NewMessage(RoleAssistant, "Hi there"))
				return s
			}(),
			expectError: false,
		},
		{
			name: "at limit",
			state: func() *State {
				s := NewState()
				s.AddMessage(NewMessage(RoleUser, "1"))
				s.AddMessage(NewMessage(RoleAssistant, "2"))
				s.AddMessage(NewMessage(RoleUser, "3"))
				return s
			}(),
			expectError: false,
		},
		{
			name: "exceeds limit",
			state: func() *State {
				s := NewState()
				s.AddMessage(NewMessage(RoleUser, "1"))
				s.AddMessage(NewMessage(RoleAssistant, "2"))
				s.AddMessage(NewMessage(RoleUser, "3"))
				s.AddMessage(NewMessage(RoleAssistant, "4"))
				return s
			}(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.state)
			if (err != nil) != tt.expectError {
				t.Errorf("Validate() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestStateValidator_ValidRolesValidator(t *testing.T) {
	validator := ValidRolesValidator(RoleUser, RoleAssistant)

	tests := []struct {
		name        string
		state       *State
		expectError bool
	}{
		{
			name: "all valid roles",
			state: func() *State {
				s := NewState()
				s.AddMessage(NewMessage(RoleUser, "Hello"))
				s.AddMessage(NewMessage(RoleAssistant, "Hi"))
				return s
			}(),
			expectError: false,
		},
		{
			name: "contains invalid role",
			state: func() *State {
				s := NewState()
				s.AddMessage(NewMessage(RoleUser, "Hello"))
				s.AddMessage(NewMessage(RoleSystem, "System message")) // Not in allowed roles
				return s
			}(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.state)
			if (err != nil) != tt.expectError {
				t.Errorf("Validate() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestStateValidator_NoEmptyMessagesValidator(t *testing.T) {
	validator := NoEmptyMessagesValidator()

	tests := []struct {
		name        string
		state       *State
		expectError bool
	}{
		{
			name: "all messages non-empty",
			state: func() *State {
				s := NewState()
				s.AddMessage(NewMessage(RoleUser, "Hello"))
				s.AddMessage(NewMessage(RoleAssistant, "Hi there"))
				return s
			}(),
			expectError: false,
		},
		{
			name: "contains empty message",
			state: func() *State {
				s := NewState()
				s.AddMessage(NewMessage(RoleUser, "Hello"))
				s.AddMessage(NewMessage(RoleAssistant, ""))
				return s
			}(),
			expectError: true,
		},
		{
			name: "contains whitespace-only message",
			state: func() *State {
				s := NewState()
				s.AddMessage(NewMessage(RoleUser, "   \t\n   "))
				return s
			}(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.state)
			if (err != nil) != tt.expectError {
				t.Errorf("Validate() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

// Skip SchemaValidator test for now since it requires a schema validator instance
// and the schema package integration. This would be better tested with a mock
// or actual schema validator instance.

func TestStateValidator_CustomValidator(t *testing.T) {
	// Custom validator that checks if state has been active for less than 1 hour
	validator := CustomValidator("age-check", func(state *State) error {
		created, ok := state.Get("created_at")
		if !ok {
			return errors.New("missing created_at timestamp")
		}

		createdTime, ok := created.(time.Time)
		if !ok {
			return errors.New("created_at is not a time.Time")
		}

		if time.Since(createdTime) > 1*time.Hour {
			return errors.New("state is too old (> 1 hour)")
		}

		return nil
	})

	tests := []struct {
		name        string
		state       *State
		expectError bool
	}{
		{
			name: "recent state",
			state: func() *State {
				s := NewState()
				s.Set("created_at", time.Now())
				return s
			}(),
			expectError: false,
		},
		{
			name: "old state",
			state: func() *State {
				s := NewState()
				s.Set("created_at", time.Now().Add(-2*time.Hour))
				return s
			}(),
			expectError: true,
		},
		{
			name:        "missing timestamp",
			state:       NewState(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.state)
			if (err != nil) != tt.expectError {
				t.Errorf("Validate() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestStateValidator_CompositeValidator(t *testing.T) {
	// Combine multiple validators with AND logic
	validator := CompositeValidator(
		RequiredKeysValidator("user_id", "session_id"),
		MaxMessageCountValidator(10),
		NoEmptyMessagesValidator(),
	)

	tests := []struct {
		name        string
		state       *State
		expectError bool
		errorDesc   string
	}{
		{
			name: "passes all validators",
			state: func() *State {
				s := NewState()
				s.Set("user_id", "123")
				s.Set("session_id", "abc")
				s.AddMessage(NewMessage(RoleUser, "Hello"))
				return s
			}(),
			expectError: false,
		},
		{
			name: "fails required keys",
			state: func() *State {
				s := NewState()
				s.Set("user_id", "123")
				// missing session_id
				s.AddMessage(NewMessage(RoleUser, "Hello"))
				return s
			}(),
			expectError: true,
			errorDesc:   "missing required key",
		},
		{
			name: "fails message count",
			state: func() *State {
				s := NewState()
				s.Set("user_id", "123")
				s.Set("session_id", "abc")
				// Add 11 messages (exceeds limit of 10)
				for i := 0; i < 11; i++ {
					s.AddMessage(NewMessage(RoleUser, "Message"))
				}
				return s
			}(),
			expectError: true,
			errorDesc:   "exceeds max messages",
		},
		{
			name: "fails empty message",
			state: func() *State {
				s := NewState()
				s.Set("user_id", "123")
				s.Set("session_id", "abc")
				s.AddMessage(NewMessage(RoleUser, ""))
				return s
			}(),
			expectError: true,
			errorDesc:   "empty message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.state)
			if (err != nil) != tt.expectError {
				t.Errorf("Validate() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestStateValidator_AnyOfValidator(t *testing.T) {
	// Passes if ANY validator passes
	validator := AnyOfValidator(
		RequiredKeysValidator("api_key"),     // Option 1: has api_key
		RequiredKeysValidator("oauth_token"), // Option 2: has oauth_token
	)

	tests := []struct {
		name        string
		state       *State
		expectError bool
	}{
		{
			name: "has api_key",
			state: func() *State {
				s := NewState()
				s.Set("api_key", "secret")
				return s
			}(),
			expectError: false,
		},
		{
			name: "has oauth_token",
			state: func() *State {
				s := NewState()
				s.Set("oauth_token", "token123")
				return s
			}(),
			expectError: false,
		},
		{
			name: "has both",
			state: func() *State {
				s := NewState()
				s.Set("api_key", "secret")
				s.Set("oauth_token", "token123")
				return s
			}(),
			expectError: false,
		},
		{
			name: "has neither",
			state: func() *State {
				s := NewState()
				s.Set("username", "john")
				return s
			}(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.state)
			if (err != nil) != tt.expectError {
				t.Errorf("Validate() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestStateValidator_NotValidator(t *testing.T) {
	// Inverts the validation result
	validator := NotValidator(
		RequiredKeysValidator("disabled"),
	)

	tests := []struct {
		name        string
		state       *State
		expectError bool
	}{
		{
			name: "does not have disabled key (passes)",
			state: func() *State {
				s := NewState()
				s.Set("enabled", true)
				return s
			}(),
			expectError: false,
		},
		{
			name: "has disabled key (fails)",
			state: func() *State {
				s := NewState()
				s.Set("disabled", true)
				return s
			}(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.state)
			if (err != nil) != tt.expectError {
				t.Errorf("Validate() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestStateValidator_ConditionalValidator(t *testing.T) {
	// Apply validator only if condition is met
	validator := ConditionalValidator(
		// Condition: if state has "validate_strict" = true
		StateValidatorFunc(func(state *State) error {
			strict, ok := state.Get("validate_strict")
			if ok && strict == true {
				return nil // Condition met
			}
			return errors.New("condition not met")
		}),
		// Then apply this validator
		RequiredKeysValidator("user_id", "session_id", "timestamp"),
		nil, // No else validator
	)

	tests := []struct {
		name        string
		state       *State
		expectError bool
	}{
		{
			name: "strict validation enabled and valid",
			state: func() *State {
				s := NewState()
				s.Set("validate_strict", true)
				s.Set("user_id", "123")
				s.Set("session_id", "abc")
				s.Set("timestamp", time.Now())
				return s
			}(),
			expectError: false,
		},
		{
			name: "strict validation enabled but invalid",
			state: func() *State {
				s := NewState()
				s.Set("validate_strict", true)
				s.Set("user_id", "123")
				// missing required fields
				return s
			}(),
			expectError: true,
		},
		{
			name: "strict validation disabled (skips validation)",
			state: func() *State {
				s := NewState()
				s.Set("validate_strict", false)
				s.Set("user_id", "123")
				// missing other fields, but validation is skipped
				return s
			}(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.state)
			if (err != nil) != tt.expectError {
				t.Errorf("Validate() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestStateValidator_ComplexScenario(t *testing.T) {
	// Complex validation scenario for a chatbot
	validator := CompositeValidator(
		// Basic requirements
		RequiredKeysValidator("user_id", "session_id"),

		// Authentication: either API key or OAuth token
		AnyOfValidator(
			RequiredKeysValidator("api_key"),
			RequiredKeysValidator("oauth_token"),
		),

		// Message validation
		MaxMessageCountValidator(100),
		NoEmptyMessagesValidator(),
		ValidRolesValidator(RoleUser, RoleAssistant, RoleSystem),

		// Conditional rate limiting check
		ConditionalValidator(
			StateValidatorFunc(func(state *State) error {
				// Apply rate limit check only for free tier users
				tier, _ := state.Get("user_tier")
				if tier == "free" {
					return nil // Condition met
				}
				return errors.New("not free tier")
			}),
			CustomValidator("rate-limit", func(state *State) error {
				// Check message rate
				msgCount := len(state.Messages())
				if msgCount > 10 {
					return errors.New("free tier limited to 10 messages")
				}
				return nil
			}),
			nil, // No else validator
		),
	)

	// Test cases

	// Valid premium user
	premiumState := NewState()
	premiumState.Set("user_id", "premium123")
	premiumState.Set("session_id", "sess456")
	premiumState.Set("api_key", "key789")
	premiumState.Set("user_tier", "premium")
	for i := 0; i < 20; i++ {
		premiumState.AddMessage(NewMessage(RoleUser, "Message"))
	}

	if err := validator.Validate(premiumState); err != nil {
		t.Errorf("Premium user validation failed: %v", err)
	}

	// Invalid free user (too many messages)
	freeState := NewState()
	freeState.Set("user_id", "free123")
	freeState.Set("session_id", "sess789")
	freeState.Set("oauth_token", "token123")
	freeState.Set("user_tier", "free")
	for i := 0; i < 15; i++ {
		freeState.AddMessage(NewMessage(RoleUser, "Message"))
	}

	if err := validator.Validate(freeState); err == nil {
		t.Error("Free user with too many messages should fail validation")
	}
}

func TestStateValidator_TypeValidator(t *testing.T) {
	validator := TypeValidator("age", reflect.TypeOf(0))

	tests := []struct {
		name        string
		state       *State
		expectError bool
	}{
		{
			name: "correct type",
			state: func() *State {
				s := NewState()
				s.Set("age", 30)
				return s
			}(),
			expectError: false,
		},
		{
			name: "incorrect type",
			state: func() *State {
				s := NewState()
				s.Set("age", "thirty")
				return s
			}(),
			expectError: true,
		},
		{
			name:        "key doesn't exist",
			state:       NewState(),
			expectError: false, // TypeValidator allows missing keys
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.state)
			if (err != nil) != tt.expectError {
				t.Errorf("Validate() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestStateValidator_RangeValidator(t *testing.T) {
	validator := RangeValidator("temperature", 0.0, 2.0)

	tests := []struct {
		name        string
		state       *State
		expectError bool
	}{
		{
			name: "within range",
			state: func() *State {
				s := NewState()
				s.Set("temperature", 1.0)
				return s
			}(),
			expectError: false,
		},
		{
			name: "below range",
			state: func() *State {
				s := NewState()
				s.Set("temperature", -0.5)
				return s
			}(),
			expectError: true,
		},
		{
			name: "above range",
			state: func() *State {
				s := NewState()
				s.Set("temperature", 2.5)
				return s
			}(),
			expectError: true,
		},
		{
			name: "integer value within range",
			state: func() *State {
				s := NewState()
				s.Set("temperature", 1)
				return s
			}(),
			expectError: false,
		},
		{
			name:        "key doesn't exist",
			state:       NewState(),
			expectError: false, // RangeValidator allows missing keys
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.state)
			if (err != nil) != tt.expectError {
				t.Errorf("Validate() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestStateValidator_RegexValidator(t *testing.T) {
	validator := RegexValidator("email", `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	tests := []struct {
		name        string
		state       *State
		expectError bool
	}{
		{
			name: "valid email",
			state: func() *State {
				s := NewState()
				s.Set("email", "test@example.com")
				return s
			}(),
			expectError: false,
		},
		{
			name: "invalid email",
			state: func() *State {
				s := NewState()
				s.Set("email", "not-an-email")
				return s
			}(),
			expectError: true,
		},
		{
			name: "non-string value",
			state: func() *State {
				s := NewState()
				s.Set("email", 123)
				return s
			}(),
			expectError: true,
		},
		{
			name:        "key doesn't exist",
			state:       NewState(),
			expectError: false, // RegexValidator allows missing keys
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.state)
			if (err != nil) != tt.expectError {
				t.Errorf("Validate() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestStateValidator_LengthValidator(t *testing.T) {
	validator := LengthValidator("username", 3, 20)

	tests := []struct {
		name        string
		state       *State
		expectError bool
	}{
		{
			name: "valid length",
			state: func() *State {
				s := NewState()
				s.Set("username", "john_doe")
				return s
			}(),
			expectError: false,
		},
		{
			name: "too short",
			state: func() *State {
				s := NewState()
				s.Set("username", "ab")
				return s
			}(),
			expectError: true,
		},
		{
			name: "too long",
			state: func() *State {
				s := NewState()
				s.Set("username", "this_username_is_way_too_long")
				return s
			}(),
			expectError: true,
		},
		{
			name: "slice within bounds",
			state: func() *State {
				s := NewState()
				s.Set("username", []string{"a", "b", "c", "d"})
				return s
			}(),
			expectError: false,
		},
		{
			name:        "key doesn't exist",
			state:       NewState(),
			expectError: false, // LengthValidator allows missing keys
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.state)
			if (err != nil) != tt.expectError {
				t.Errorf("Validate() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestStateValidator_EnumValidator(t *testing.T) {
	validator := EnumValidator("status", "active", "inactive", "pending")

	tests := []struct {
		name        string
		state       *State
		expectError bool
	}{
		{
			name: "valid enum value",
			state: func() *State {
				s := NewState()
				s.Set("status", "active")
				return s
			}(),
			expectError: false,
		},
		{
			name: "invalid enum value",
			state: func() *State {
				s := NewState()
				s.Set("status", "unknown")
				return s
			}(),
			expectError: true,
		},
		{
			name:        "key doesn't exist",
			state:       NewState(),
			expectError: false, // EnumValidator allows missing keys
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.state)
			if (err != nil) != tt.expectError {
				t.Errorf("Validate() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestStateValidator_DependencyValidator(t *testing.T) {
	validator := DependencyValidator("billing_address", "first_name", "last_name", "zip_code")

	tests := []struct {
		name        string
		state       *State
		expectError bool
	}{
		{
			name: "all dependencies satisfied",
			state: func() *State {
				s := NewState()
				s.Set("billing_address", "123 Main St")
				s.Set("first_name", "John")
				s.Set("last_name", "Doe")
				s.Set("zip_code", "12345")
				return s
			}(),
			expectError: false,
		},
		{
			name: "missing dependencies",
			state: func() *State {
				s := NewState()
				s.Set("billing_address", "123 Main St")
				s.Set("first_name", "John")
				// missing last_name and zip_code
				return s
			}(),
			expectError: true,
		},
		{
			name: "primary key doesn't exist",
			state: func() *State {
				s := NewState()
				s.Set("first_name", "John")
				s.Set("last_name", "Doe")
				// no billing_address, so dependencies not required
				return s
			}(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.state)
			if (err != nil) != tt.expectError {
				t.Errorf("Validate() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestStateValidator_AllOfValidator(t *testing.T) {
	// All validators must pass
	validator := AllOfValidator(
		RequiredKeysValidator("name"),
		LengthValidator("name", 2, 50),
		RegexValidator("name", `^[a-zA-Z\s]+$`),
	)

	tests := []struct {
		name        string
		state       *State
		expectError bool
	}{
		{
			name: "all validators pass",
			state: func() *State {
				s := NewState()
				s.Set("name", "John Doe")
				return s
			}(),
			expectError: false,
		},
		{
			name: "missing required key",
			state: func() *State {
				s := NewState()
				// no name key
				return s
			}(),
			expectError: true,
		},
		{
			name: "name too short",
			state: func() *State {
				s := NewState()
				s.Set("name", "J")
				return s
			}(),
			expectError: true,
		},
		{
			name: "name contains numbers",
			state: func() *State {
				s := NewState()
				s.Set("name", "John123")
				return s
			}(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.state)
			if (err != nil) != tt.expectError {
				t.Errorf("Validate() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestStateValidator_MessageValidator(t *testing.T) {
	validator := MessageValidator(func(messages []Message) error {
		// Custom message validation: ensure alternating roles
		for i := 1; i < len(messages); i++ {
			if messages[i].Role == messages[i-1].Role {
				return errors.New("consecutive messages have the same role")
			}
		}
		return nil
	})

	tests := []struct {
		name        string
		state       *State
		expectError bool
	}{
		{
			name: "alternating roles",
			state: func() *State {
				s := NewState()
				s.AddMessage(NewMessage(RoleUser, "Hello"))
				s.AddMessage(NewMessage(RoleAssistant, "Hi"))
				s.AddMessage(NewMessage(RoleUser, "How are you?"))
				return s
			}(),
			expectError: false,
		},
		{
			name: "same consecutive roles",
			state: func() *State {
				s := NewState()
				s.AddMessage(NewMessage(RoleUser, "Hello"))
				s.AddMessage(NewMessage(RoleUser, "Are you there?"))
				return s
			}(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.state)
			if (err != nil) != tt.expectError {
				t.Errorf("Validate() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestStateValidator_ArtifactValidator(t *testing.T) {
	validator := ArtifactValidator(func(artifacts map[string]*Artifact) error {
		// Ensure all artifacts have non-empty data
		for id, artifact := range artifacts {
			data, err := artifact.Data()
			if err != nil || len(data) == 0 {
				return errors.New("artifact " + id + " has empty content")
			}
		}
		return nil
	})

	tests := []struct {
		name        string
		state       *State
		expectError bool
	}{
		{
			name: "all artifacts have content",
			state: func() *State {
				s := NewState()
				artifact := NewArtifact("doc1", ArtifactTypeDocument, []byte("Document content"))
				artifact.WithMimeType("text/plain")
				s.AddArtifact(artifact)
				return s
			}(),
			expectError: false,
		},
		{
			name: "artifact with empty content",
			state: func() *State {
				s := NewState()
				artifact := NewArtifact("empty", ArtifactTypeDocument, []byte{})
				artifact.WithMimeType("text/plain")
				s.AddArtifact(artifact)
				return s
			}(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.state)
			if (err != nil) != tt.expectError {
				t.Errorf("Validate() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}
