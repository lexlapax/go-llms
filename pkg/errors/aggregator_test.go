package errors

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// TestNewErrorAggregator tests creating new aggregator
func TestNewErrorAggregator(t *testing.T) {
	agg := NewErrorAggregator()

	if agg == nil {
		t.Fatal("expected non-nil aggregator")
	}

	if agg.HasErrors() {
		t.Error("expected no errors initially")
	}

	if agg.Error() != nil {
		t.Error("expected nil error initially")
	}
}

// TestErrorAggregatorAdd tests adding errors
func TestErrorAggregatorAdd(t *testing.T) {
	agg := NewErrorAggregator()

	// Add nil (should be ignored)
	agg.Add(nil)
	if agg.HasErrors() {
		t.Error("expected nil to be ignored")
	}

	// Add errors
	err1 := fmt.Errorf("error 1")
	err2 := fmt.Errorf("error 2")

	agg.Add(err1)
	agg.Add(err2)

	if !agg.HasErrors() {
		t.Error("expected errors to be present")
	}

	errs := agg.Errors()
	if len(errs) != 2 {
		t.Errorf("expected 2 errors, got %d", len(errs))
	}

	if errs[0] != err1 || errs[1] != err2 {
		t.Error("expected errors to match")
	}
}

// TestErrorAggregatorAddWithContext tests adding errors with context
func TestErrorAggregatorAddWithContext(t *testing.T) {
	agg := NewErrorAggregator()

	// Add with context
	err := fmt.Errorf("test error")
	context := map[string]interface{}{
		"key": "value",
		"num": 123,
	}

	agg.AddWithContext(err, context)

	// Test with BaseError that has existing context
	baseErr := NewError("base error").WithContext("existing", "value")
	agg.AddWithContext(baseErr, map[string]interface{}{"new": "value"})

	errs := agg.Errors()
	if len(errs) != 2 {
		t.Error("expected 2 errors")
	}
}

// TestErrorAggregatorError tests aggregated error message
func TestErrorAggregatorError(t *testing.T) {
	agg := NewErrorAggregator()

	// Empty aggregator
	if agg.Error() != nil {
		t.Error("expected nil for empty aggregator")
	}

	// Single error
	err1 := fmt.Errorf("single error")
	agg.Add(err1)

	if agg.Error() != err1 {
		t.Error("expected single error to be returned directly")
	}

	// Multiple errors
	err2 := fmt.Errorf("second error")
	agg.Add(err2)

	aggErr := agg.Error()
	if aggErr == nil {
		t.Fatal("expected non-nil aggregated error")
	}

	errMsg := aggErr.Error()
	if !strings.Contains(errMsg, "multiple errors occurred (2)") {
		t.Error("expected error count in message")
	}

	if !strings.Contains(errMsg, "single error") || !strings.Contains(errMsg, "second error") {
		t.Error("expected all error messages to be included")
	}
}

// TestErrorAggregatorClear tests clearing errors
func TestErrorAggregatorClear(t *testing.T) {
	agg := NewErrorAggregator()

	agg.Add(fmt.Errorf("error 1"))
	agg.Add(fmt.Errorf("error 2"))

	if !agg.HasErrors() {
		t.Error("expected errors before clear")
	}

	agg.Clear()

	if agg.HasErrors() {
		t.Error("expected no errors after clear")
	}

	if len(agg.Errors()) != 0 {
		t.Error("expected empty errors list after clear")
	}
}

// TestErrorAggregatorToSerializable tests serializable conversion
func TestErrorAggregatorToSerializable(t *testing.T) {
	agg := NewErrorAggregator()

	// Empty aggregator
	if agg.ToSerializable() != nil {
		t.Error("expected nil for empty aggregator")
	}

	// Add various errors
	agg.Add(fmt.Errorf("standard error"))
	agg.Add(NewError("base error").SetRetryable(true))
	agg.Add(NewError("fatal error").SetFatal(true))

	serializable := agg.ToSerializable()
	if serializable == nil {
		t.Fatal("expected non-nil serializable error")
	}

	// Check type
	baseErr, ok := serializable.(*BaseError)
	if !ok {
		t.Fatal("expected BaseError type")
	}

	if baseErr.Type != "AggregatedError" {
		t.Error("expected AggregatedError type")
	}

	// Check context
	if errorCount, ok := baseErr.Context["error_count"]; !ok || errorCount != 3 {
		t.Error("expected error count in context")
	}

	// Check retryable and fatal flags
	if !baseErr.Retryable {
		t.Error("expected retryable to be true")
	}

	if !baseErr.Fatal {
		t.Error("expected fatal to be true")
	}

	// Check JSON serialization
	data, err := baseErr.ToJSON()
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	// Check errors array in context
	if context, ok := parsed["context"].(map[string]interface{}); ok {
		if errors, ok := context["errors"].([]interface{}); ok {
			if len(errors) != 3 {
				t.Error("expected 3 errors in context")
			}
		} else {
			t.Error("expected errors array in context")
		}
	} else {
		t.Error("expected context in JSON")
	}
}

// TestAggregateErrors tests helper function
func TestAggregateErrors(t *testing.T) {
	err1 := fmt.Errorf("error 1")
	err2 := fmt.Errorf("error 2")
	err3 := fmt.Errorf("error 3")

	aggErr := AggregateErrors(err1, err2, err3)

	if aggErr == nil {
		t.Fatal("expected non-nil aggregated error")
	}

	errMsg := aggErr.Error()
	if !strings.Contains(errMsg, "multiple errors occurred (3)") {
		t.Error("expected error count")
	}

	// Test with single error
	singleErr := AggregateErrors(err1)
	if singleErr != err1 {
		t.Error("expected single error to be returned directly")
	}

	// Test with no errors
	noErr := AggregateErrors()
	if noErr != nil {
		t.Error("expected nil for no errors")
	}
}

// TestAggregateErrorsWithContext tests helper with context
func TestAggregateErrorsWithContext(t *testing.T) {
	context := map[string]interface{}{
		"operation": "test",
		"user":      "test-user",
	}

	err1 := fmt.Errorf("error 1")
	err2 := fmt.Errorf("error 2")

	aggErr := AggregateErrorsWithContext(context, err1, err2)

	if aggErr == nil {
		t.Fatal("expected non-nil error")
	}

	if !strings.Contains(aggErr.Error(), "multiple errors occurred") {
		t.Error("expected aggregated error message")
	}
}

// TestErrorList tests error list
func TestErrorList(t *testing.T) {
	err1 := fmt.Errorf("error 1")
	err2 := fmt.Errorf("error 2")

	// Create with initial errors
	list := NewErrorList(err1, err2)

	if list.IsEmpty() {
		t.Error("expected non-empty list")
	}

	if len(list.Errors()) != 2 {
		t.Error("expected 2 errors")
	}

	// Add more errors
	list.Add(fmt.Errorf("error 3"))
	list.Add(nil) // Should be ignored

	if len(list.Errors()) != 3 {
		t.Error("expected 3 errors after adding")
	}

	// Test Error() method
	errMsg := list.Error()
	if !strings.Contains(errMsg, "multiple errors") {
		t.Error("expected multiple errors message")
	}

	// Test First and Last
	if list.First() != err1 {
		t.Error("expected first error to match")
	}

	if list.Last().Error() != "error 3" {
		t.Error("expected last error to match")
	}
}

// TestErrorListEmpty tests empty error list
func TestErrorListEmpty(t *testing.T) {
	list := NewErrorList()

	if !list.IsEmpty() {
		t.Error("expected empty list")
	}

	if list.Error() != "no errors" {
		t.Error("expected 'no errors' message")
	}

	if list.First() != nil {
		t.Error("expected nil first error")
	}

	if list.Last() != nil {
		t.Error("expected nil last error")
	}
}

// TestErrorListSingle tests single error in list
func TestErrorListSingle(t *testing.T) {
	err := fmt.Errorf("single error")
	list := NewErrorList(err)

	if list.Error() != "single error" {
		t.Error("expected single error message")
	}

	if list.First() != err || list.Last() != err {
		t.Error("expected first and last to be same")
	}
}

// TestErrorAggregatorConcurrency tests concurrent access
func TestErrorAggregatorConcurrency(t *testing.T) {
	agg := NewErrorAggregator()

	// Concurrent adds
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(n int) {
			agg.Add(fmt.Errorf("error %d", n))
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Check results
	errs := agg.Errors()
	if len(errs) != 10 {
		t.Errorf("expected 10 errors, got %d", len(errs))
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			_ = agg.HasErrors()
			_ = agg.Error()
			_ = agg.Errors()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestSerializableErrorInAggregator tests serializable errors in aggregator
func TestSerializableErrorInAggregator(t *testing.T) {
	agg := NewErrorAggregator()

	// Add a serializable error
	serErr := NewError("serializable").
		WithCode("SER_001").
		WithContext("key", "value")

	agg.Add(serErr)

	// Convert to serializable
	aggSer := agg.ToSerializable()
	if aggSer == nil {
		t.Fatal("expected non-nil serializable")
	}

	// Check if individual error details are preserved
	baseErr := aggSer.(*BaseError)
	if errors, ok := baseErr.Context["errors"].([]map[string]interface{}); ok {
		if len(errors) != 1 {
			t.Fatal("expected 1 error detail")
		}

		// Check if serialized version is included
		if _, hasSerialized := errors[0]["serialized"]; !hasSerialized {
			t.Error("expected serialized error details")
		}
	} else {
		t.Error("expected errors in context")
	}
}
