package validation

import (
	"testing"
)

func TestCoerceToString(t *testing.T) {
	validator := NewValidator()

	t.Run("number to string", func(t *testing.T) {
		result, ok := validator.Coerce("string", 42.0)
		if !ok {
			t.Errorf("Expected successful coercion from number to string")
		}
		str, isStr := result.(string)
		if !isStr {
			t.Errorf("Expected string result, got %T", result)
		}
		if str != "42" {
			t.Errorf("Expected '42', got '%s'", str)
		}
	})

	t.Run("boolean to string", func(t *testing.T) {
		result, ok := validator.Coerce("string", true)
		if !ok {
			t.Errorf("Expected successful coercion from boolean to string")
		}
		str, isStr := result.(string)
		if !isStr {
			t.Errorf("Expected string result, got %T", result)
		}
		if str != "true" {
			t.Errorf("Expected 'true', got '%s'", str)
		}
	})

	t.Run("no coercion for string", func(t *testing.T) {
		result, ok := validator.Coerce("string", "already a string")
		if !ok {
			t.Errorf("Expected successful handling of string to string")
		}
		str, isStr := result.(string)
		if !isStr {
			t.Errorf("Expected string result, got %T", result)
		}
		if str != "already a string" {
			t.Errorf("Expected 'already a string', got '%s'", str)
		}
	})
}

func TestCoerceToNumber(t *testing.T) {
	validator := NewValidator()

	t.Run("string to number", func(t *testing.T) {
		result, ok := validator.Coerce("number", "42.5")
		if !ok {
			t.Errorf("Expected successful coercion from string to number")
		}
		num, isNum := result.(float64)
		if !isNum {
			t.Errorf("Expected float64 result, got %T", result)
		}
		if num != 42.5 {
			t.Errorf("Expected 42.5, got %f", num)
		}
	})

	t.Run("integer string to number", func(t *testing.T) {
		result, ok := validator.Coerce("number", "42")
		if !ok {
			t.Errorf("Expected successful coercion from string to number")
		}
		num, isNum := result.(float64)
		if !isNum {
			t.Errorf("Expected float64 result, got %T", result)
		}
		if num != 42.0 {
			t.Errorf("Expected 42.0, got %f", num)
		}
	})

	t.Run("boolean to number", func(t *testing.T) {
		result, ok := validator.Coerce("number", true)
		if !ok {
			t.Errorf("Expected successful coercion from boolean to number")
		}
		num, isNum := result.(float64)
		if !isNum {
			t.Errorf("Expected float64 result, got %T", result)
		}
		if num != 1.0 {
			t.Errorf("Expected 1.0, got %f", num)
		}
	})

	t.Run("invalid string to number", func(t *testing.T) {
		_, ok := validator.Coerce("number", "not-a-number")
		if ok {
			t.Errorf("Expected failed coercion from invalid string to number")
		}
	})
}

func TestCoerceToInteger(t *testing.T) {
	validator := NewValidator()

	t.Run("float to integer", func(t *testing.T) {
		result, ok := validator.Coerce("integer", 42.5)
		if !ok {
			t.Errorf("Expected successful coercion from float to integer")
		}
		num, isNum := result.(int64)
		if !isNum {
			t.Errorf("Expected int64 result, got %T", result)
		}
		if num != 42 {
			t.Errorf("Expected 42, got %d", num)
		}
	})

	t.Run("string to integer", func(t *testing.T) {
		result, ok := validator.Coerce("integer", "42")
		if !ok {
			t.Errorf("Expected successful coercion from string to integer")
		}
		num, isNum := result.(int64)
		if !isNum {
			t.Errorf("Expected int64 result, got %T", result)
		}
		if num != 42 {
			t.Errorf("Expected 42, got %d", num)
		}
	})

	t.Run("boolean to integer", func(t *testing.T) {
		result, ok := validator.Coerce("integer", true)
		if !ok {
			t.Errorf("Expected successful coercion from boolean to integer")
		}
		num, isNum := result.(int64)
		if !isNum {
			t.Errorf("Expected int64 result, got %T", result)
		}
		if num != 1 {
			t.Errorf("Expected 1, got %d", num)
		}
	})

	t.Run("invalid string to integer", func(t *testing.T) {
		_, ok := validator.Coerce("integer", "not-an-integer")
		if ok {
			t.Errorf("Expected failed coercion from invalid string to integer")
		}
	})
}

func TestCoerceToBoolean(t *testing.T) {
	validator := NewValidator()

	t.Run("string to boolean - true", func(t *testing.T) {
		result, ok := validator.Coerce("boolean", "true")
		if !ok {
			t.Errorf("Expected successful coercion from string to boolean")
		}
		b, isBool := result.(bool)
		if !isBool {
			t.Errorf("Expected bool result, got %T", result)
		}
		if !b {
			t.Errorf("Expected true, got %v", b)
		}
	})

	t.Run("string to boolean - false", func(t *testing.T) {
		result, ok := validator.Coerce("boolean", "false")
		if !ok {
			t.Errorf("Expected successful coercion from string to boolean")
		}
		b, isBool := result.(bool)
		if !isBool {
			t.Errorf("Expected bool result, got %T", result)
		}
		if b {
			t.Errorf("Expected false, got %v", b)
		}
	})

	t.Run("number to boolean", func(t *testing.T) {
		result, ok := validator.Coerce("boolean", 1.0)
		if !ok {
			t.Errorf("Expected successful coercion from number to boolean")
		}
		b, isBool := result.(bool)
		if !isBool {
			t.Errorf("Expected bool result, got %T", result)
		}
		if !b {
			t.Errorf("Expected true, got %v", b)
		}
	})

	t.Run("number to boolean - zero", func(t *testing.T) {
		result, ok := validator.Coerce("boolean", 0.0)
		if !ok {
			t.Errorf("Expected successful coercion from number to boolean")
		}
		b, isBool := result.(bool)
		if !isBool {
			t.Errorf("Expected bool result, got %T", result)
		}
		if b {
			t.Errorf("Expected false, got %v", b)
		}
	})

	t.Run("invalid string to boolean", func(t *testing.T) {
		_, ok := validator.Coerce("boolean", "not-a-boolean")
		if ok {
			t.Errorf("Expected failed coercion from invalid string to boolean")
		}
	})
}