package errors

import (
	"fmt"
	"testing"
	"time"
)

// TestNewErrorContext tests creating new error context
func TestNewErrorContext(t *testing.T) {
	ctx := NewErrorContext()

	if ctx == nil {
		t.Fatal("expected non-nil context")
	}

	// Should be empty initially
	all := ctx.GetAll()
	if len(all) != 0 {
		t.Error("expected empty context initially")
	}
}

// TestErrorContextAdd tests adding values to context
func TestErrorContextAdd(t *testing.T) {
	ctx := NewErrorContext()

	// Test chaining
	result := ctx.Add("key1", "value1").Add("key2", 123)
	if result != ctx {
		t.Error("expected Add to return same context for chaining")
	}

	// Verify values
	val, ok := ctx.Get("key1")
	if !ok || val != "value1" {
		t.Error("expected key1 to have value1")
	}

	val, ok = ctx.Get("key2")
	if !ok || val != 123 {
		t.Error("expected key2 to have 123")
	}
}

// TestErrorContextAddAll tests adding multiple values
func TestErrorContextAddAll(t *testing.T) {
	ctx := NewErrorContext()

	values := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
		"key3": true,
	}

	result := ctx.AddAll(values)
	if result != ctx {
		t.Error("expected AddAll to return same context")
	}

	all := ctx.GetAll()
	if len(all) != 3 {
		t.Error("expected 3 values in context")
	}

	for k, v := range values {
		if all[k] != v {
			t.Errorf("expected %s to have %v", k, v)
		}
	}
}

// TestErrorContextGet tests retrieving values
func TestErrorContextGet(t *testing.T) {
	ctx := NewErrorContext()
	ctx.Add("key", "value")

	// Existing key
	val, ok := ctx.Get("key")
	if !ok {
		t.Error("expected to find key")
	}
	if val != "value" {
		t.Error("expected correct value")
	}

	// Non-existing key
	_, ok = ctx.Get("nonexistent")
	if ok {
		t.Error("expected not to find nonexistent key")
	}
}

// TestErrorContextGetAll tests getting all values
func TestErrorContextGetAll(t *testing.T) {
	ctx := NewErrorContext()
	ctx.Add("key1", "value1").Add("key2", "value2")

	all := ctx.GetAll()

	// Should be a copy
	all["key3"] = "value3"

	// Original should not be modified
	_, ok := ctx.Get("key3")
	if ok {
		t.Error("expected GetAll to return a copy")
	}
}

// TestErrorContextWithStackTrace tests adding stack trace
func TestErrorContextWithStackTrace(t *testing.T) {
	ctx := NewErrorContext()

	result := ctx.WithStackTrace()
	if result != ctx {
		t.Error("expected WithStackTrace to return same context")
	}

	val, ok := ctx.Get("stack_trace")
	if !ok {
		t.Error("expected stack_trace to be added")
	}

	if stack, ok := val.([]StackFrame); ok {
		if len(stack) == 0 {
			t.Error("expected non-empty stack trace")
		}
	} else {
		t.Error("expected stack_trace to be []StackFrame")
	}
}

// TestErrorContextWithTimestamp tests adding timestamp
func TestErrorContextWithTimestamp(t *testing.T) {
	ctx := NewErrorContext()
	before := time.Now()

	result := ctx.WithTimestamp()
	if result != ctx {
		t.Error("expected WithTimestamp to return same context")
	}

	val, ok := ctx.Get("timestamp")
	if !ok {
		t.Error("expected timestamp to be added")
	}

	if ts, ok := val.(time.Time); ok {
		if ts.Before(before) {
			t.Error("expected timestamp to be after test start")
		}
	} else {
		t.Error("expected timestamp to be time.Time")
	}
}

// TestErrorContextClone tests cloning context
func TestErrorContextClone(t *testing.T) {
	ctx := NewErrorContext()
	ctx.Add("key1", "value1").Add("key2", "value2")

	clone := ctx.Clone()

	// Should have same values
	all := clone.GetAll()
	if len(all) != 2 {
		t.Error("expected clone to have 2 values")
	}

	// Modify clone
	clone.Add("key3", "value3")

	// Original should not be affected
	_, ok := ctx.Get("key3")
	if ok {
		t.Error("expected original to be unaffected by clone modification")
	}
}

// TestNewContextualError tests creating contextual error
func TestNewContextualError(t *testing.T) {
	err := NewContextualError(fmt.Errorf("test error"))

	if err.Error() != "test error" {
		t.Error("expected error message to match")
	}

	// Should have timestamp
	all := err.context.GetAll()
	if _, ok := all["timestamp"]; !ok {
		t.Error("expected timestamp to be added automatically")
	}
}

// TestContextualErrorMethods tests contextual error methods
func TestContextualErrorMethods(t *testing.T) {
	original := fmt.Errorf("original")
	err := NewContextualError(original)

	// Test Unwrap
	if err.Unwrap() != original {
		t.Error("expected Unwrap to return original error")
	}

	// Test WithContext
	result := err.WithContext("key", "value")
	if result != err {
		t.Error("expected WithContext to return same error")
	}

	val, _ := err.Context().Get("key")
	if val != "value" {
		t.Error("expected context to be updated")
	}

	// Test WithContextMap
	_ = err.WithContextMap(map[string]interface{}{
		"key2": "value2",
		"key3": "value3",
	})

	all := err.Context().GetAll()
	if len(all) < 3 {
		t.Error("expected at least 3 values in context")
	}
}

// TestContextualErrorProvideContext tests ProvideContext
func TestContextualErrorProvideContext(t *testing.T) {
	err := NewContextualError(fmt.Errorf("test")).
		WithContext("key1", "value1").
		WithContext("key2", "value2")

	newCtx := NewErrorContext()
	newCtx.Add("key3", "value3")

	result := err.ProvideContext(newCtx)

	// Should merge contexts
	all := result.GetAll()
	if len(all) < 3 {
		t.Error("expected merged context to have all values")
	}

	if all["key1"] != "value1" || all["key2"] != "value2" || all["key3"] != "value3" {
		t.Error("expected all values to be present")
	}
}

// TestNewErrorBuilder tests error builder creation
func TestNewErrorBuilder(t *testing.T) {
	builder := NewErrorBuilder("test message")

	if builder.message != "test message" {
		t.Error("expected message to be set")
	}

	if builder.context == nil {
		t.Error("expected context to be initialized")
	}
}

// TestErrorBuilderMethods tests builder methods
func TestErrorBuilderMethods(t *testing.T) {
	cause := fmt.Errorf("cause")
	recovery := NewNoRetryStrategy()

	err := NewErrorBuilder("test error").
		WithCause(cause).
		WithCode("ERR_001").
		WithType("TestError").
		WithContext("key", "value").
		WithContextMap(map[string]interface{}{"key2": "value2"}).
		WithRetryable(true).
		WithFatal(false).
		WithRecovery(recovery).
		Build()

	if err.Message != "test error" {
		t.Error("expected message")
	}

	if err.Cause != cause {
		t.Error("expected cause")
	}

	if err.Code != "ERR_001" {
		t.Error("expected code")
	}

	if err.Type != "TestError" {
		t.Error("expected type")
	}

	if err.Context["key"] != "value" || err.Context["key2"] != "value2" {
		t.Error("expected context values")
	}

	if !err.Retryable {
		t.Error("expected retryable")
	}

	if err.Fatal {
		t.Error("expected not fatal")
	}

	if err.Recovery != recovery {
		t.Error("expected recovery strategy")
	}

	if len(err.Stack) == 0 {
		t.Error("expected stack trace")
	}
}

// TestErrorBuilderDefaults tests builder defaults
func TestErrorBuilderDefaults(t *testing.T) {
	err := NewErrorBuilder("test").Build()

	if err.Type != "Error" {
		t.Error("expected default type 'Error'")
	}

	if err.CauseText != "" {
		t.Error("expected empty cause text without cause")
	}
}

// TestEnrichError tests error enrichment
func TestEnrichError(t *testing.T) {
	// Test with nil
	if EnrichError(nil, nil) != nil {
		t.Error("expected nil for nil error")
	}

	// Test with BaseError
	baseErr := NewError("test")
	enriched := EnrichError(baseErr, map[string]interface{}{
		"key": "value",
	})

	if enrichedBase, ok := enriched.(*BaseError); ok {
		if enrichedBase.Context["key"] != "value" {
			t.Error("expected context to be added")
		}
	} else {
		t.Error("expected BaseError type")
	}

	// Test with standard error
	stdErr := fmt.Errorf("standard")
	enriched = EnrichError(stdErr, map[string]interface{}{
		"key": "value",
	})

	if enrichedBase, ok := enriched.(*BaseError); ok {
		if enrichedBase.Context["key"] != "value" {
			t.Error("expected context to be added")
		}
		if enrichedBase.Cause != stdErr {
			t.Error("expected original error to be wrapped")
		}
	} else {
		t.Error("expected BaseError type")
	}
}

// TestEnrichErrorWithOperation tests operation enrichment
func TestEnrichErrorWithOperation(t *testing.T) {
	err := fmt.Errorf("test")
	enriched := EnrichErrorWithOperation(err, "TestOperation")

	if be, ok := enriched.(*BaseError); ok {
		if be.Context["operation"] != "TestOperation" {
			t.Error("expected operation in context")
		}
		if _, hasTimestamp := be.Context["timestamp"]; !hasTimestamp {
			t.Error("expected timestamp in context")
		}
	} else {
		t.Error("expected BaseError")
	}
}

// TestEnrichErrorWithRequest tests request enrichment
func TestEnrichErrorWithRequest(t *testing.T) {
	err := fmt.Errorf("test")
	enriched := EnrichErrorWithRequest(err, "POST", "https://api.example.com", 500)

	if be, ok := enriched.(*BaseError); ok {
		if be.Context["request_method"] != "POST" {
			t.Error("expected request method")
		}
		if be.Context["request_url"] != "https://api.example.com" {
			t.Error("expected request URL")
		}
		if be.Context["status_code"] != 500 {
			t.Error("expected status code")
		}
	} else {
		t.Error("expected BaseError")
	}
}

// TestEnrichErrorWithResource tests resource enrichment
func TestEnrichErrorWithResource(t *testing.T) {
	err := fmt.Errorf("test")
	enriched := EnrichErrorWithResource(err, "User", "12345")

	if be, ok := enriched.(*BaseError); ok {
		if be.Context["resource_type"] != "User" {
			t.Error("expected resource type")
		}
		if be.Context["resource_id"] != "12345" {
			t.Error("expected resource ID")
		}
	} else {
		t.Error("expected BaseError")
	}
}

// TestCaptureRuntimeContext tests runtime context capture
func TestCaptureRuntimeContext(t *testing.T) {
	ctx := CaptureRuntimeContext()

	if ctx.GoVersion == "" {
		t.Error("expected Go version")
	}

	if ctx.GOOS == "" {
		t.Error("expected GOOS")
	}

	if ctx.GOARCH == "" {
		t.Error("expected GOARCH")
	}

	if ctx.NumCPU == 0 {
		t.Error("expected NumCPU > 0")
	}

	if ctx.NumGoroutine == 0 {
		t.Error("expected NumGoroutine > 0")
	}

	if ctx.MemStats == nil {
		t.Error("expected MemStats")
	}
}

// TestEnrichErrorWithRuntime tests runtime enrichment
func TestEnrichErrorWithRuntime(t *testing.T) {
	err := fmt.Errorf("test")
	enriched := EnrichErrorWithRuntime(err)

	if be, ok := enriched.(*BaseError); ok {
		if runtime, ok := be.Context["runtime"]; ok {
			if rtCtx, ok := runtime.(RuntimeContext); ok {
				if rtCtx.GoVersion == "" {
					t.Error("expected runtime context")
				}
			} else {
				t.Error("expected RuntimeContext type")
			}
		} else {
			t.Error("expected runtime in context")
		}
	} else {
		t.Error("expected BaseError")
	}
}

// TestErrorContextConcurrency tests concurrent access
func TestErrorContextConcurrency(t *testing.T) {
	ctx := NewErrorContext()

	// Concurrent writes
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(n int) {
			ctx.Add(fmt.Sprintf("key%d", n), n)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all values
	all := ctx.GetAll()
	if len(all) != 10 {
		t.Errorf("expected 10 values, got %d", len(all))
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func(n int) {
			val, ok := ctx.Get(fmt.Sprintf("key%d", n))
			if !ok || val != n {
				t.Errorf("expected key%d to have value %d", n, n)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
