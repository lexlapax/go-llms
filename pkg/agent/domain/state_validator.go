// ABOUTME: Defines validators for agent state validation
// ABOUTME: Provides built-in validators and composable validation patterns

package domain

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// StateValidator validates state contents and structure.
// Validators can check for required fields, value constraints,
// and business rules to ensure state integrity during agent execution.
type StateValidator interface {
	Validate(state *State) error
}

// StateValidatorFunc is a function adapter that implements StateValidator.
// Allows plain functions to be used as state validators.
type StateValidatorFunc func(state *State) error

func (f StateValidatorFunc) Validate(state *State) error {
	return f(state)
}

// Built-in validators

// RequiredKeysValidator ensures required keys exist in the state.
// Validates that all specified keys are present with non-nil values.
func RequiredKeysValidator(keys ...string) StateValidator {
	return StateValidatorFunc(func(state *State) error {
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

// SchemaValidator validates against JSON schema
// Note: This requires a validator instance to be provided
func SchemaValidator(validator sdomain.Validator, schema *sdomain.Schema) StateValidator {
	return StateValidatorFunc(func(state *State) error {
		// Convert state values to JSON for validation
		data, err := json.Marshal(state.Values())
		if err != nil {
			return fmt.Errorf("failed to marshal state values: %w", err)
		}

		result, err := validator.Validate(schema, string(data))
		if err != nil {
			return fmt.Errorf("validation error: %w", err)
		}

		if !result.Valid {
			return fmt.Errorf("validation failed: %v", result.Errors)
		}

		return nil
	})
}

// TypeValidator ensures values are of correct type
func TypeValidator(key string, expectedType reflect.Type) StateValidator {
	return StateValidatorFunc(func(state *State) error {
		val, ok := state.Get(key)
		if !ok {
			return nil // Key doesn't exist, not a type error
		}

		actualType := reflect.TypeOf(val)
		if actualType != expectedType {
			return fmt.Errorf("key %s: expected type %s, got %s", key, expectedType, actualType)
		}
		return nil
	})
}

// RangeValidator ensures numeric values are within range
func RangeValidator(key string, min, max float64) StateValidator {
	return StateValidatorFunc(func(state *State) error {
		val, ok := state.Get(key)
		if !ok {
			return nil // Key doesn't exist
		}

		// Try to convert to float64
		var numVal float64
		switch v := val.(type) {
		case float64:
			numVal = v
		case float32:
			numVal = float64(v)
		case int:
			numVal = float64(v)
		case int32:
			numVal = float64(v)
		case int64:
			numVal = float64(v)
		default:
			return fmt.Errorf("key %s: value is not numeric", key)
		}

		if numVal < min || numVal > max {
			return fmt.Errorf("key %s: value %f is outside range [%f, %f]", key, numVal, min, max)
		}
		return nil
	})
}

// RegexValidator ensures string values match a pattern
func RegexValidator(key string, pattern string) StateValidator {
	re := regexp.MustCompile(pattern)
	return StateValidatorFunc(func(state *State) error {
		val, ok := state.Get(key)
		if !ok {
			return nil // Key doesn't exist
		}

		str, ok := val.(string)
		if !ok {
			return fmt.Errorf("key %s: value is not a string", key)
		}

		if !re.MatchString(str) {
			return fmt.Errorf("key %s: value '%s' does not match pattern %s", key, str, pattern)
		}
		return nil
	})
}

// LengthValidator ensures string/slice length is within bounds
func LengthValidator(key string, minLen, maxLen int) StateValidator {
	return StateValidatorFunc(func(state *State) error {
		val, ok := state.Get(key)
		if !ok {
			return nil // Key doesn't exist
		}

		var length int
		switch v := val.(type) {
		case string:
			length = len(v)
		case []interface{}:
			length = len(v)
		case []string:
			length = len(v)
		case []int:
			length = len(v)
		default:
			// Try reflection for other slice types
			rv := reflect.ValueOf(val)
			if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
				length = rv.Len()
			} else {
				return fmt.Errorf("key %s: value does not have a length", key)
			}
		}

		if length < minLen {
			return fmt.Errorf("key %s: length %d is less than minimum %d", key, length, minLen)
		}
		if maxLen >= 0 && length > maxLen {
			return fmt.Errorf("key %s: length %d exceeds maximum %d", key, length, maxLen)
		}
		return nil
	})
}

// EnumValidator ensures value is one of allowed values
func EnumValidator(key string, allowedValues ...interface{}) StateValidator {
	// Create a map for faster lookup
	allowed := make(map[interface{}]bool)
	for _, v := range allowedValues {
		allowed[v] = true
	}

	return StateValidatorFunc(func(state *State) error {
		val, ok := state.Get(key)
		if !ok {
			return nil // Key doesn't exist
		}

		if !allowed[val] {
			return fmt.Errorf("key %s: value %v is not in allowed values %v", key, val, allowedValues)
		}
		return nil
	})
}

// DependencyValidator ensures if one key exists, dependent keys must also exist
func DependencyValidator(key string, dependentKeys ...string) StateValidator {
	return StateValidatorFunc(func(state *State) error {
		// Check if the primary key exists
		if _, ok := state.Get(key); !ok {
			return nil // Primary key doesn't exist, no dependencies required
		}

		// Check all dependent keys
		var missing []string
		for _, depKey := range dependentKeys {
			if _, ok := state.Get(depKey); !ok {
				missing = append(missing, depKey)
			}
		}

		if len(missing) > 0 {
			return fmt.Errorf("key %s requires dependent keys: %v", key, missing)
		}
		return nil
	})
}

// ConditionalValidator applies validation based on a condition
func ConditionalValidator(condition StateValidatorFunc, thenValidator, elseValidator StateValidator) StateValidator {
	return StateValidatorFunc(func(state *State) error {
		// Check condition
		if err := condition(state); err == nil {
			// Condition met, apply then validator
			if thenValidator != nil {
				return thenValidator.Validate(state)
			}
		} else {
			// Condition not met, apply else validator
			if elseValidator != nil {
				return elseValidator.Validate(state)
			}
		}
		return nil
	})
}

// CustomValidator creates a validator from a custom function
func CustomValidator(name string, fn func(state *State) error) StateValidator {
	return StateValidatorFunc(func(state *State) error {
		if err := fn(state); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		return nil
	})
}

// CompositeValidator combines multiple validators
func CompositeValidator(validators ...StateValidator) StateValidator {
	return StateValidatorFunc(func(state *State) error {
		var errors []error
		for _, v := range validators {
			if err := v.Validate(state); err != nil {
				errors = append(errors, err)
			}
		}
		if len(errors) > 0 {
			return &MultiError{Errors: errors}
		}
		return nil
	})
}

// AllOfValidator ensures all validators pass
func AllOfValidator(validators ...StateValidator) StateValidator {
	return StateValidatorFunc(func(state *State) error {
		for _, v := range validators {
			if err := v.Validate(state); err != nil {
				return err
			}
		}
		return nil
	})
}

// AnyOfValidator ensures at least one validator passes
func AnyOfValidator(validators ...StateValidator) StateValidator {
	return StateValidatorFunc(func(state *State) error {
		var errors []error
		for _, v := range validators {
			if err := v.Validate(state); err == nil {
				return nil // At least one passed
			} else {
				errors = append(errors, err)
			}
		}
		return fmt.Errorf("none of the validators passed: %v", errors)
	})
}

// NotValidator negates a validator
func NotValidator(validator StateValidator) StateValidator {
	return StateValidatorFunc(func(state *State) error {
		if err := validator.Validate(state); err == nil {
			return fmt.Errorf("validation should have failed but passed")
		}
		return nil
	})
}

// MessageValidator validates message content
func MessageValidator(validator func(messages []Message) error) StateValidator {
	return StateValidatorFunc(func(state *State) error {
		messages := state.Messages()
		return validator(messages)
	})
}

// ArtifactValidator validates artifacts
func ArtifactValidator(validator func(artifacts map[string]*Artifact) error) StateValidator {
	return StateValidatorFunc(func(state *State) error {
		artifacts := state.Artifacts()
		return validator(artifacts)
	})
}

// MaxMessageCountValidator limits message count
func MaxMessageCountValidator(maxCount int) StateValidator {
	return MessageValidator(func(messages []Message) error {
		if len(messages) > maxCount {
			return fmt.Errorf("message count %d exceeds maximum %d", len(messages), maxCount)
		}
		return nil
	})
}

// NoEmptyMessagesValidator ensures no empty messages
func NoEmptyMessagesValidator() StateValidator {
	return MessageValidator(func(messages []Message) error {
		for i, msg := range messages {
			if strings.TrimSpace(msg.Content) == "" {
				return fmt.Errorf("message at index %d is empty", i)
			}
		}
		return nil
	})
}

// ValidRolesValidator ensures messages have valid roles
func ValidRolesValidator(validRoles ...Role) StateValidator {
	roleSet := make(map[Role]bool)
	for _, role := range validRoles {
		roleSet[role] = true
	}

	return MessageValidator(func(messages []Message) error {
		for i, msg := range messages {
			if !roleSet[msg.Role] {
				return fmt.Errorf("message at index %d has invalid role: %s", i, msg.Role)
			}
		}
		return nil
	})
}
