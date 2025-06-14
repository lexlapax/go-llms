package errors

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestNewError tests creating a new base error
func TestNewError(t *testing.T) {
	err := NewError("test error")

	if err.Message != "test error" {
		t.Errorf("expected message 'test error', got %s", err.Message)
	}

	if err.Type != "BaseError" {
		t.Errorf("expected type 'BaseError', got %s", err.Type)
	}

	if err.Context == nil {
		t.Error("expected context to be initialized")
	}

	if len(err.Stack) == 0 {
		t.Error("expected stack trace to be captured")
	}
}

// TestNewErrorWithCode tests creating error with code
func TestNewErrorWithCode(t *testing.T) {
	err := NewErrorWithCode("ERR_001", "test error")

	if err.Code != "ERR_001" {
		t.Errorf("expected code 'ERR_001', got %s", err.Code)
	}

	if err.Type != "ERR_001" {
		t.Errorf("expected type to match code, got %s", err.Type)
	}
}

// TestWrap tests error wrapping
func TestWrap(t *testing.T) {
	original := fmt.Errorf("original error")
	wrapped := Wrap(original, "wrapped error")

	if wrapped.Message != "wrapped error" {
		t.Errorf("expected message 'wrapped error', got %s", wrapped.Message)
	}

	if wrapped.Cause != original {
		t.Error("expected cause to be original error")
	}

	if wrapped.CauseText != "original error" {
		t.Errorf("expected cause text 'original error', got %s", wrapped.CauseText)
	}

	// Test wrapping nil
	if Wrap(nil, "test") != nil {
		t.Error("expected nil when wrapping nil error")
	}
}

// TestWrapBaseError tests wrapping another BaseError
func TestWrapBaseError(t *testing.T) {
	original := NewError("original").
		WithCode("ERR_001").
		SetRetryable(true).
		WithContext("key1", "value1")

	wrapped := Wrap(original, "wrapped")

	if wrapped.Code != "ERR_001" {
		t.Error("expected code to be inherited")
	}

	if !wrapped.Retryable {
		t.Error("expected retryable to be inherited")
	}

	if val, ok := wrapped.Context["key1"]; !ok || val != "value1" {
		t.Error("expected context to be merged")
	}
}

// TestErrorInterface tests error interface implementation
func TestErrorInterface(t *testing.T) {
	err := NewError("test error")
	if err.Error() != "test error" {
		t.Errorf("expected 'test error', got %s", err.Error())
	}

	// With cause
	cause := fmt.Errorf("cause error")
	wrapped := Wrap(cause, "wrapped")
	if !strings.Contains(wrapped.Error(), "wrapped") || !strings.Contains(wrapped.Error(), "cause error") {
		t.Errorf("expected error to contain both messages, got %s", wrapped.Error())
	}
}

// TestToJSON tests JSON serialization
func TestToJSON(t *testing.T) {
	err := NewError("test error").
		WithCode("ERR_001").
		WithContext("key", "value").
		SetRetryable(true)

	data, jsonErr := err.ToJSON()
	if jsonErr != nil {
		t.Fatalf("failed to serialize to JSON: %v", jsonErr)
	}

	// Parse back
	var parsed map[string]interface{}
	if jsonErr := json.Unmarshal(data, &parsed); jsonErr != nil {
		t.Fatalf("failed to parse JSON: %v", jsonErr)
	}

	if parsed["message"] != "test error" {
		t.Error("expected message in JSON")
	}

	if parsed["code"] != "ERR_001" {
		t.Error("expected code in JSON")
	}

	if parsed["retryable"] != true {
		t.Error("expected retryable in JSON")
	}

	if context, ok := parsed["context"].(map[string]interface{}); ok {
		if context["key"] != "value" {
			t.Error("expected context value in JSON")
		}
	} else {
		t.Error("expected context in JSON")
	}
}

// TestGetContext tests context retrieval
func TestGetContext(t *testing.T) {
	err := NewError("test")
	ctx := err.GetContext()

	if ctx == nil {
		t.Fatal("expected context to be initialized")
	}

	// Add context
	_ = err.WithContext("key", "value")
	ctx = err.GetContext()

	if ctx["key"] != "value" {
		t.Error("expected context value")
	}
}

// TestWithContext tests context addition
func TestWithContext(t *testing.T) {
	err := NewError("test")

	// Chain context additions
	_ = err.
		WithContext("key1", "value1").
		WithContext("key2", 123).
		WithContextMap(map[string]interface{}{
			"key3": true,
			"key4": []string{"a", "b"},
		})

	ctx := err.GetContext()

	if ctx["key1"] != "value1" {
		t.Error("expected key1")
	}

	if ctx["key2"] != 123 {
		t.Error("expected key2")
	}

	if ctx["key3"] != true {
		t.Error("expected key3")
	}

	if _, ok := ctx["key4"].([]string); !ok {
		t.Error("expected key4")
	}
}

// TestSetters tests various setter methods
func TestSetters(t *testing.T) {
	err := NewError("test")

	_ = err.
		WithCode("CODE_123").
		WithType("CustomError").
		SetRetryable(true).
		SetFatal(true).
		WithRecovery(NewNoRetryStrategy())

	if err.Code != "CODE_123" {
		t.Error("expected code to be set")
	}

	if err.Type != "CustomError" {
		t.Error("expected type to be set")
	}

	if !err.Retryable {
		t.Error("expected retryable to be true")
	}

	if !err.Fatal {
		t.Error("expected fatal to be true")
	}

	if err.Recovery == nil {
		t.Error("expected recovery strategy to be set")
	}
}

// TestUnwrap tests error unwrapping
func TestUnwrap(t *testing.T) {
	original := fmt.Errorf("original")
	wrapped := Wrap(original, "wrapped")

	unwrapped := wrapped.Unwrap()
	if unwrapped != original {
		t.Error("expected unwrapped error to be original")
	}
}

// TestIs tests error matching
func TestIs(t *testing.T) {
	err1 := NewErrorWithCode("ERR_001", "error 1")
	err2 := NewErrorWithCode("ERR_001", "error 2")
	err3 := NewErrorWithCode("ERR_002", "error 3")

	// Same code should match
	if !err1.Is(err2) {
		t.Error("expected errors with same code to match")
	}

	// Different code should not match
	if err1.Is(err3) {
		t.Error("expected errors with different codes not to match")
	}

	// Test with wrapped error
	original := fmt.Errorf("test")
	wrapped := Wrap(original, "wrapped")
	if !wrapped.Is(original) {
		t.Error("expected wrapped error to match original")
	}

	// Test with nil
	if err1.Is(nil) {
		t.Error("expected not to match nil")
	}
}

// TestMarshalJSON tests custom JSON marshaling
func TestMarshalJSON(t *testing.T) {
	err := NewError("test").
		WithRecovery(NewExponentialBackoffStrategy(3, time.Second, time.Minute))

	data, jsonErr := json.Marshal(err)
	if jsonErr != nil {
		t.Fatalf("failed to marshal: %v", jsonErr)
	}

	var parsed map[string]interface{}
	if jsonErr := json.Unmarshal(data, &parsed); jsonErr != nil {
		t.Fatalf("failed to unmarshal: %v", jsonErr)
	}

	// Check timestamp format
	if _, ok := parsed["timestamp_str"].(string); !ok {
		t.Error("expected timestamp_str in JSON")
	}

	// Check recovery strategy name
	if parsed["recovery_strategy"] != "exponential_backoff" {
		t.Error("expected recovery strategy name in JSON")
	}
}

// TestUnmarshalJSON tests JSON deserialization
func TestUnmarshalJSON(t *testing.T) {
	jsonData := `{
		"type": "TestError",
		"message": "test message",
		"code": "TEST_001",
		"cause": "original error",
		"retryable": true,
		"timestamp_str": "2023-01-01T00:00:00Z"
	}`

	var err BaseError
	if jsonErr := json.Unmarshal([]byte(jsonData), &err); jsonErr != nil {
		t.Fatalf("failed to unmarshal: %v", jsonErr)
	}

	if err.Type != "TestError" {
		t.Error("expected type to be deserialized")
	}

	if err.Message != "test message" {
		t.Error("expected message to be deserialized")
	}

	if err.Code != "TEST_001" {
		t.Error("expected code to be deserialized")
	}

	if !err.Retryable {
		t.Error("expected retryable to be deserialized")
	}

	if err.CauseText != "original error" {
		t.Error("expected cause text to be deserialized")
	}

	if err.Cause == nil {
		t.Error("expected cause to be recreated")
	}
}

// TestErrorFromJSON tests creating error from JSON
func TestErrorFromJSON(t *testing.T) {
	original := NewError("test").
		WithCode("TEST_001").
		WithContext("key", "value")

	data, _ := original.ToJSON()

	restored, err := ErrorFromJSON(data)
	if err != nil {
		t.Fatalf("failed to restore from JSON: %v", err)
	}

	if restored.Message != original.Message {
		t.Error("expected message to match")
	}

	if restored.Code != original.Code {
		t.Error("expected code to match")
	}

	if restored.Context["key"] != "value" {
		t.Error("expected context to match")
	}
}

// TestIsRetryableError tests retryable error checking
func TestIsRetryableError(t *testing.T) {
	// Test with nil
	if IsRetryableError(nil) {
		t.Error("expected nil not to be retryable")
	}

	// Test with retryable error
	retryable := NewError("test").SetRetryable(true)
	if !IsRetryableError(retryable) {
		t.Error("expected error to be retryable")
	}

	// Test with non-retryable error
	nonRetryable := NewError("test").SetRetryable(false)
	if IsRetryableError(nonRetryable) {
		t.Error("expected error not to be retryable")
	}

	// Test with wrapped retryable error
	wrapped := Wrap(retryable, "wrapped")
	if !IsRetryableError(wrapped) {
		t.Error("expected wrapped error to be retryable")
	}

	// Test with standard error
	stdErr := fmt.Errorf("standard error")
	if IsRetryableError(stdErr) {
		t.Error("expected standard error not to be retryable")
	}
}

// TestIsFatalError tests fatal error checking
func TestIsFatalError(t *testing.T) {
	// Test with nil
	if IsFatalError(nil) {
		t.Error("expected nil not to be fatal")
	}

	// Test with fatal error
	fatal := NewError("test").SetFatal(true)
	if !IsFatalError(fatal) {
		t.Error("expected error to be fatal")
	}

	// Test with non-fatal error
	nonFatal := NewError("test").SetFatal(false)
	if IsFatalError(nonFatal) {
		t.Error("expected error not to be fatal")
	}
}

// TestGetErrorContext tests context extraction
func TestGetErrorContext(t *testing.T) {
	// Test with nil
	if GetErrorContext(nil) != nil {
		t.Error("expected nil context for nil error")
	}

	// Test with BaseError
	err := NewError("test").WithContext("key", "value")
	ctx := GetErrorContext(err)
	if ctx == nil || ctx["key"] != "value" {
		t.Error("expected context to be extracted")
	}

	// Test with standard error
	stdErr := fmt.Errorf("standard")
	if GetErrorContext(stdErr) != nil {
		t.Error("expected nil context for standard error")
	}
}

// TestAs tests error type assertion
func TestAs(t *testing.T) {
	err := NewError("test")

	var target *BaseError
	if !As(err, &target) {
		t.Error("expected As to succeed")
	}

	if target != err {
		t.Error("expected target to be set")
	}

	// Test with wrapped error
	wrapped := Wrap(err, "wrapped")
	target = nil
	if !As(wrapped, &target) {
		t.Error("expected As to succeed for wrapped error")
	}

	// Test with nil
	if As(nil, &target) {
		t.Error("expected As to fail for nil")
	}

	// Test with standard error
	stdErr := fmt.Errorf("standard")
	target = nil
	if As(stdErr, &target) {
		t.Error("expected As to fail for standard error")
	}
}

// TestStackTrace tests stack trace capture
func TestStackTrace(t *testing.T) {
	err := NewError("test")

	if len(err.Stack) == 0 {
		t.Fatal("expected stack trace to be captured")
	}

	// Check first frame
	firstFrame := err.Stack[0]
	if firstFrame.Function == "" {
		t.Error("expected function name in stack frame")
	}

	if firstFrame.File == "" {
		t.Error("expected file name in stack frame")
	}

	if firstFrame.Line == 0 {
		t.Error("expected line number in stack frame")
	}

	// Test explicit stack trace
	err2 := NewError("test2")
	err2.Stack = nil
	_ = err2.WithStackTrace()

	if len(err2.Stack) == 0 {
		t.Error("expected stack trace to be added")
	}
}

// TestCaptureStackTrace tests stack trace capture function
func TestCaptureStackTrace(t *testing.T) {
	stack := captureStackTrace(0)

	if len(stack) == 0 {
		t.Fatal("expected stack frames")
	}

	// Should not contain runtime frames
	for _, frame := range stack {
		if strings.Contains(frame.Function, "runtime.") {
			t.Error("expected runtime frames to be filtered")
			break
		}
	}

	// Should be limited to 20 frames
	if len(stack) > 20 {
		t.Error("expected stack to be limited to 20 frames")
	}
}
