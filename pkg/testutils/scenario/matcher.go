// ABOUTME: Matcher interface and implementations for test assertions
// ABOUTME: Provides flexible value matching for scenario-based testing
// Package scenario provides a framework for building and executing complex test scenarios.
// It includes a comprehensive set of matchers for flexible test assertions and
// scenario builders for constructing multi-step test cases.
package scenario

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// Matcher defines the interface for matching values in tests
type Matcher interface {
	// Match checks if the value matches and returns success and explanation
	Match(value interface{}) (bool, string)
	// Description returns a human-readable description of the matcher
	Description() string
}

// MatcherFunc is a function type that implements Matcher interface
type MatcherFunc func(value interface{}) (bool, string)

// Match implements Matcher interface
func (f MatcherFunc) Match(value interface{}) (bool, string) {
	return f(value)
}

// Description implements Matcher interface
func (f MatcherFunc) Description() string {
	return "custom matcher function"
}

// Basic Matchers

// Equals matches exact equality
type equalsMatcher struct {
	expected interface{}
}

// Equals creates a matcher that checks for exact equality
func Equals(expected interface{}) Matcher {
	return &equalsMatcher{expected: expected}
}

func (m *equalsMatcher) Match(value interface{}) (bool, string) {
	if reflect.DeepEqual(value, m.expected) {
		return true, ""
	}
	return false, fmt.Sprintf("expected %v, got %v", m.expected, value)
}

func (m *equalsMatcher) Description() string {
	return fmt.Sprintf("equals %v", m.expected)
}

// Contains matches substring presence
type containsMatcher struct {
	substring string
}

// Contains creates a matcher that checks for substring presence
func Contains(substring string) Matcher {
	return &containsMatcher{substring: substring}
}

func (m *containsMatcher) Match(value interface{}) (bool, string) {
	str, ok := value.(string)
	if !ok {
		return false, fmt.Sprintf("expected string, got %T", value)
	}
	if strings.Contains(str, m.substring) {
		return true, ""
	}
	return false, fmt.Sprintf("string %q does not contain %q", str, m.substring)
}

func (m *containsMatcher) Description() string {
	return fmt.Sprintf("contains %q", m.substring)
}

// HasField matches field presence and value
type hasFieldMatcher struct {
	field        string
	valueMatcher Matcher
}

// HasField creates a matcher that checks for field presence and value
func HasField(field string, valueMatcher Matcher) Matcher {
	return &hasFieldMatcher{field: field, valueMatcher: valueMatcher}
}

func (m *hasFieldMatcher) Match(value interface{}) (bool, string) {
	// Try to access field via reflection
	rv := reflect.ValueOf(value)

	// Handle pointers
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	// Check if it's a struct
	if rv.Kind() == reflect.Struct {
		field := rv.FieldByName(m.field)
		if !field.IsValid() {
			return false, fmt.Sprintf("field %q not found", m.field)
		}
		return m.valueMatcher.Match(field.Interface())
	}

	// Check if it's a map
	if rv.Kind() == reflect.Map {
		if rv.Type().Key().Kind() != reflect.String {
			return false, fmt.Sprintf("map key type must be string, got %v", rv.Type().Key())
		}
		key := reflect.ValueOf(m.field)
		val := rv.MapIndex(key)
		if !val.IsValid() {
			return false, fmt.Sprintf("key %q not found in map", m.field)
		}
		return m.valueMatcher.Match(val.Interface())
	}

	return false, fmt.Sprintf("expected struct or map, got %T", value)
}

func (m *hasFieldMatcher) Description() string {
	return fmt.Sprintf("has field %q matching %v", m.field, m.valueMatcher.Description())
}

// IsNil matches nil values
type isNilMatcher struct{}

// IsNil creates a matcher that checks for nil
func IsNil() Matcher {
	return &isNilMatcher{}
}

func (m *isNilMatcher) Match(value interface{}) (bool, string) {
	if value == nil {
		return true, ""
	}

	// Check for nil pointer, slice, map, etc.
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func, reflect.Interface:
		if rv.IsNil() {
			return true, ""
		}
	}

	return false, fmt.Sprintf("expected nil, got %v", value)
}

func (m *isNilMatcher) Description() string {
	return "is nil"
}

// IsNotNil matches non-nil values
type isNotNilMatcher struct{}

// IsNotNil creates a matcher that checks for non-nil
func IsNotNil() Matcher {
	return &isNotNilMatcher{}
}

func (m *isNotNilMatcher) Match(value interface{}) (bool, string) {
	if value == nil {
		return false, "expected non-nil value, got nil"
	}

	// Check for nil pointer, slice, map, etc.
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func, reflect.Interface:
		if rv.IsNil() {
			return false, fmt.Sprintf("expected non-nil, got nil %v", rv.Type())
		}
	}

	return true, ""
}

func (m *isNotNilMatcher) Description() string {
	return "is not nil"
}

// Advanced Matchers

// MatchesJSON matches JSON structure
type matchesJSONMatcher struct {
	pattern string
}

// MatchesJSON creates a matcher that checks JSON structure
func MatchesJSON(pattern string) Matcher {
	return &matchesJSONMatcher{pattern: pattern}
}

func (m *matchesJSONMatcher) Match(value interface{}) (bool, string) {
	// Convert value to JSON
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Sprintf("failed to marshal value to JSON: %v", err)
	}

	// Parse pattern
	var patternData interface{}
	if err := json.Unmarshal([]byte(m.pattern), &patternData); err != nil {
		return false, fmt.Sprintf("invalid JSON pattern: %v", err)
	}

	// Parse value JSON
	var valueData interface{}
	if err := json.Unmarshal(valueJSON, &valueData); err != nil {
		return false, fmt.Sprintf("failed to parse value JSON: %v", err)
	}

	// Compare structures
	if reflect.DeepEqual(patternData, valueData) {
		return true, ""
	}

	return false, "JSON structure does not match pattern"
}

func (m *matchesJSONMatcher) Description() string {
	return fmt.Sprintf("matches JSON %s", m.pattern)
}

// MatchesRegex matches regular expression
type matchesRegexMatcher struct {
	pattern *regexp.Regexp
}

// MatchesRegex creates a matcher that checks regex pattern
func MatchesRegex(pattern string) Matcher {
	re, err := regexp.Compile(pattern)
	if err != nil {
		// Return a matcher that always fails with error
		return MatcherFunc(func(value interface{}) (bool, string) {
			return false, fmt.Sprintf("invalid regex pattern: %v", err)
		})
	}
	return &matchesRegexMatcher{pattern: re}
}

func (m *matchesRegexMatcher) Match(value interface{}) (bool, string) {
	str, ok := value.(string)
	if !ok {
		return false, fmt.Sprintf("expected string, got %T", value)
	}

	if m.pattern.MatchString(str) {
		return true, ""
	}

	return false, fmt.Sprintf("string %q does not match pattern %q", str, m.pattern.String())
}

func (m *matchesRegexMatcher) Description() string {
	return fmt.Sprintf("matches regex %q", m.pattern.String())
}

// HasLength matches collection length
type hasLengthMatcher struct {
	expected int
}

// HasLength creates a matcher that checks collection length
func HasLength(expected int) Matcher {
	return &hasLengthMatcher{expected: expected}
}

func (m *hasLengthMatcher) Match(value interface{}) (bool, string) {
	rv := reflect.ValueOf(value)

	switch rv.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map, reflect.String, reflect.Chan:
		if rv.Len() == m.expected {
			return true, ""
		}
		return false, fmt.Sprintf("expected length %d, got %d", m.expected, rv.Len())
	default:
		return false, fmt.Sprintf("type %T does not have length", value)
	}
}

func (m *hasLengthMatcher) Description() string {
	return fmt.Sprintf("has length %d", m.expected)
}

// IsEmpty matches empty collections
type isEmptyMatcher struct{}

// IsEmpty creates a matcher that checks for empty collections
func IsEmpty() Matcher {
	return &isEmptyMatcher{}
}

func (m *isEmptyMatcher) Match(value interface{}) (bool, string) {
	rv := reflect.ValueOf(value)

	switch rv.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map, reflect.String, reflect.Chan:
		if rv.Len() == 0 {
			return true, ""
		}
		return false, fmt.Sprintf("expected empty collection, got length %d", rv.Len())
	default:
		return false, fmt.Sprintf("type %T cannot be empty", value)
	}
}

func (m *isEmptyMatcher) Description() string {
	return "is empty"
}

// IsBetween matches numeric ranges
type isBetweenMatcher struct {
	min, max interface{}
}

// IsBetween creates a matcher that checks numeric ranges
func IsBetween(min, max interface{}) Matcher {
	return &isBetweenMatcher{min: min, max: max}
}

func (m *isBetweenMatcher) Match(value interface{}) (bool, string) {
	// Convert all to float64 for comparison
	val, err := toFloat64(value)
	if err != nil {
		return false, fmt.Sprintf("value is not numeric: %v", err)
	}

	minVal, err := toFloat64(m.min)
	if err != nil {
		return false, fmt.Sprintf("min is not numeric: %v", err)
	}

	maxVal, err := toFloat64(m.max)
	if err != nil {
		return false, fmt.Sprintf("max is not numeric: %v", err)
	}

	if val >= minVal && val <= maxVal {
		return true, ""
	}

	return false, fmt.Sprintf("value %v is not between %v and %v", value, m.min, m.max)
}

func (m *isBetweenMatcher) Description() string {
	return fmt.Sprintf("is between %v and %v", m.min, m.max)
}

// Composite Matchers

// AllOf matches when all matchers match
type allOfMatcher struct {
	matchers []Matcher
}

// AllOf creates a matcher that requires all matchers to match
func AllOf(matchers ...Matcher) Matcher {
	return &allOfMatcher{matchers: matchers}
}

func (m *allOfMatcher) Match(value interface{}) (bool, string) {
	for i, matcher := range m.matchers {
		if ok, msg := matcher.Match(value); !ok {
			return false, fmt.Sprintf("matcher %d failed: %s", i+1, msg)
		}
	}
	return true, ""
}

func (m *allOfMatcher) Description() string {
	descriptions := make([]string, len(m.matchers))
	for i, matcher := range m.matchers {
		descriptions[i] = matcher.Description()
	}
	return fmt.Sprintf("all of [%s]", strings.Join(descriptions, ", "))
}

// AnyOf matches when any matcher matches
type anyOfMatcher struct {
	matchers []Matcher
}

// AnyOf creates a matcher that requires any matcher to match
func AnyOf(matchers ...Matcher) Matcher {
	return &anyOfMatcher{matchers: matchers}
}

func (m *anyOfMatcher) Match(value interface{}) (bool, string) {
	var failures []string
	for _, matcher := range m.matchers {
		if ok, msg := matcher.Match(value); ok {
			return true, ""
		} else if msg != "" {
			failures = append(failures, msg)
		}
	}
	return false, fmt.Sprintf("none of the matchers matched: %s", strings.Join(failures, "; "))
}

func (m *anyOfMatcher) Description() string {
	descriptions := make([]string, len(m.matchers))
	for i, matcher := range m.matchers {
		descriptions[i] = matcher.Description()
	}
	return fmt.Sprintf("any of [%s]", strings.Join(descriptions, ", "))
}

// Not inverts a matcher
type notMatcher struct {
	matcher Matcher
}

// Not creates a matcher that inverts another matcher
func Not(matcher Matcher) Matcher {
	return &notMatcher{matcher: matcher}
}

func (m *notMatcher) Match(value interface{}) (bool, string) {
	if ok, _ := m.matcher.Match(value); ok {
		return false, fmt.Sprintf("expected not to match %s", m.matcher.Description())
	}
	return true, ""
}

func (m *notMatcher) Description() string {
	return fmt.Sprintf("not %s", m.matcher.Description())
}

// Helper functions

func toFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int8:
		return float64(val), nil
	case int16:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case uint:
		return float64(val), nil
	case uint8:
		return float64(val), nil
	case uint16:
		return float64(val), nil
	case uint32:
		return float64(val), nil
	case uint64:
		return float64(val), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}
