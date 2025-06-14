package scenario

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicMatchers(t *testing.T) {
	t.Run("Equals", func(t *testing.T) {
		matcher := Equals("hello")

		ok, msg := matcher.Match("hello")
		assert.True(t, ok)
		assert.Empty(t, msg)

		ok, msg = matcher.Match("world")
		assert.False(t, ok)
		assert.Contains(t, msg, "expected hello, got world")

		// Test with complex types
		matcher2 := Equals(map[string]int{"a": 1, "b": 2})
		ok, _ = matcher2.Match(map[string]int{"a": 1, "b": 2})
		assert.True(t, ok)
	})

	t.Run("Contains", func(t *testing.T) {
		matcher := Contains("world")

		ok, msg := matcher.Match("hello world")
		assert.True(t, ok)
		assert.Empty(t, msg)

		ok, msg = matcher.Match("hello")
		assert.False(t, ok)
		assert.Contains(t, msg, "does not contain")

		// Non-string value
		ok, msg = matcher.Match(123)
		assert.False(t, ok)
		assert.Contains(t, msg, "expected string")
	})

	t.Run("HasField", func(t *testing.T) {
		// Test with struct
		type testStruct struct {
			Name string
			Age  int
		}

		matcher := HasField("Name", Equals("John"))
		obj := testStruct{Name: "John", Age: 30}

		ok, msg := matcher.Match(obj)
		assert.True(t, ok)
		assert.Empty(t, msg)

		// Test with pointer
		ok, _ = matcher.Match(&obj)
		assert.True(t, ok)

		// Test with map
		mapObj := map[string]interface{}{"Name": "John", "Age": 30}
		ok, _ = matcher.Match(mapObj)
		assert.True(t, ok)

		// Field not found
		matcher2 := HasField("Missing", Equals("value"))
		ok, msg = matcher2.Match(obj)
		assert.False(t, ok)
		assert.Contains(t, msg, "not found")
	})

	t.Run("IsNil", func(t *testing.T) {
		matcher := IsNil()

		ok, _ := matcher.Match(nil)
		assert.True(t, ok)

		var ptr *string
		ok, _ = matcher.Match(ptr)
		assert.True(t, ok)

		var slice []int
		ok, _ = matcher.Match(slice)
		assert.True(t, ok)

		ok, msg := matcher.Match("not nil")
		assert.False(t, ok)
		assert.Contains(t, msg, "expected nil")
	})

	t.Run("IsNotNil", func(t *testing.T) {
		matcher := IsNotNil()

		ok, _ := matcher.Match("value")
		assert.True(t, ok)

		str := "hello"
		ok, _ = matcher.Match(&str)
		assert.True(t, ok)

		ok, msg := matcher.Match(nil)
		assert.False(t, ok)
		assert.Contains(t, msg, "expected non-nil")
	})
}

func TestAdvancedMatchers(t *testing.T) {
	t.Run("MatchesJSON", func(t *testing.T) {
		matcher := MatchesJSON(`{"name": "John", "age": 30}`)

		// Test with matching structure
		obj := map[string]interface{}{"name": "John", "age": 30}
		ok, _ := matcher.Match(obj)
		assert.True(t, ok)

		// Test with different structure
		obj2 := map[string]interface{}{"name": "Jane", "age": 25}
		ok, msg := matcher.Match(obj2)
		assert.False(t, ok)
		assert.Contains(t, msg, "does not match")

		// Test with invalid pattern
		badMatcher := MatchesJSON(`{invalid json}`)
		ok2, msg2 := badMatcher.Match(obj)
		assert.False(t, ok2)
		assert.Contains(t, msg2, "invalid JSON pattern")
	})

	t.Run("MatchesRegex", func(t *testing.T) {
		matcher := MatchesRegex(`^hello.*world$`)

		ok, _ := matcher.Match("hello beautiful world")
		assert.True(t, ok)

		ok, msg := matcher.Match("goodbye world")
		assert.False(t, ok)
		assert.Contains(t, msg, "does not match pattern")

		// Non-string value
		ok, msg = matcher.Match(123)
		assert.False(t, ok)
		assert.Contains(t, msg, "expected string")

		// Invalid regex
		badMatcher := MatchesRegex(`[invalid`)
		ok, msg = badMatcher.Match("test")
		assert.False(t, ok)
		assert.Contains(t, msg, "invalid regex")
	})

	t.Run("HasLength", func(t *testing.T) {
		matcher := HasLength(3)

		// Slice
		ok, _ := matcher.Match([]int{1, 2, 3})
		assert.True(t, ok)

		// String
		ok, _ = matcher.Match("abc")
		assert.True(t, ok)

		// Map
		ok, _ = matcher.Match(map[string]int{"a": 1, "b": 2, "c": 3})
		assert.True(t, ok)

		// Wrong length
		ok, msg := matcher.Match([]int{1, 2})
		assert.False(t, ok)
		assert.Contains(t, msg, "expected length 3, got 2")

		// Type without length
		ok, msg = matcher.Match(123)
		assert.False(t, ok)
		assert.Contains(t, msg, "does not have length")
	})

	t.Run("IsEmpty", func(t *testing.T) {
		matcher := IsEmpty()

		ok, _ := matcher.Match([]int{})
		assert.True(t, ok)

		ok, _ = matcher.Match("")
		assert.True(t, ok)

		ok, _ = matcher.Match(map[string]int{})
		assert.True(t, ok)

		ok, msg := matcher.Match([]int{1})
		assert.False(t, ok)
		assert.Contains(t, msg, "expected empty")
	})

	t.Run("IsBetween", func(t *testing.T) {
		matcher := IsBetween(10, 20)

		ok, _ := matcher.Match(15)
		assert.True(t, ok)

		ok, _ = matcher.Match(10)
		assert.True(t, ok)

		ok, _ = matcher.Match(20)
		assert.True(t, ok)

		ok, msg := matcher.Match(5)
		assert.False(t, ok)
		assert.Contains(t, msg, "not between")

		// Non-numeric
		ok, msg = matcher.Match("hello")
		assert.False(t, ok)
		assert.Contains(t, msg, "not numeric")
	})
}

func TestCompositeMatchers(t *testing.T) {
	t.Run("AllOf", func(t *testing.T) {
		matcher := AllOf(
			IsNotNil(),
			HasLength(5),
			Contains("hello"),
		)

		ok, _ := matcher.Match("hello")
		assert.True(t, ok)

		ok, msg := matcher.Match("hi")
		assert.False(t, ok)
		assert.Contains(t, msg, "matcher 2 failed") // HasLength fails

		assert.Contains(t, matcher.Description(), "all of")
	})

	t.Run("AnyOf", func(t *testing.T) {
		matcher := AnyOf(
			Equals("hello"),
			Equals("world"),
			Contains("test"),
		)

		ok, _ := matcher.Match("hello")
		assert.True(t, ok)

		ok, _ = matcher.Match("test string")
		assert.True(t, ok)

		ok, msg := matcher.Match("goodbye")
		assert.False(t, ok)
		assert.Contains(t, msg, "none of the matchers")
	})

	t.Run("Not", func(t *testing.T) {
		matcher := Not(Equals("hello"))

		ok, _ := matcher.Match("world")
		assert.True(t, ok)

		ok, msg := matcher.Match("hello")
		assert.False(t, ok)
		assert.Contains(t, msg, "expected not to match")
	})

	t.Run("Complex composition", func(t *testing.T) {
		// Match a non-nil string that contains "test" but is not exactly "test"
		matcher := AllOf(
			IsNotNil(),
			Contains("test"),
			Not(Equals("test")),
		)

		ok, _ := matcher.Match("testing 123")
		assert.True(t, ok)

		ok, _ = matcher.Match("test")
		assert.False(t, ok)
	})
}

func TestMatcherFunc(t *testing.T) {
	// Custom matcher using MatcherFunc
	isError := MatcherFunc(func(value interface{}) (bool, string) {
		_, ok := value.(error)
		if ok {
			return true, ""
		}
		return false, "expected error type"
	})

	ok, _ := isError.Match(errors.New("test error"))
	assert.True(t, ok)

	ok, msg := isError.Match("not an error")
	assert.False(t, ok)
	assert.Contains(t, msg, "expected error type")
}
