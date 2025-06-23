// ABOUTME: Event filtering system with pattern matching and composite filters
// ABOUTME: Provides flexible event filtering for bridge layer integration

package events

import (
	"regexp"
	"strings"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// PatternFilter matches events by type pattern using regular expressions.
// It supports simple wildcard patterns that are converted to regex internally.
type PatternFilter struct {
	pattern *regexp.Regexp
	raw     string
}

// NewPatternFilter creates a filter that matches event types by pattern.
// The pattern supports simple wildcards where '*' matches any characters.
//
// Examples:
//   - "tool.*" matches all tool events
//   - "agent.start" matches exact type
//   - "*.error" matches any error events
//
// Parameters:
//   - pattern: The pattern string with optional wildcards
//
// Returns a PatternFilter and nil on success, or nil and error if pattern is invalid.
func NewPatternFilter(pattern string) (*PatternFilter, error) {
	// Convert simple wildcard to regex
	regexPattern := strings.ReplaceAll(pattern, ".", "\\.")
	regexPattern = strings.ReplaceAll(regexPattern, "*", ".*")
	regexPattern = "^" + regexPattern + "$"

	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, err
	}

	return &PatternFilter{
		pattern: re,
		raw:     pattern,
	}, nil
}

// Match implements the EventFilter interface.
// It returns true if the event type matches the pattern.
func (f *PatternFilter) Match(event domain.Event) bool {
	return f.pattern.MatchString(string(event.Type))
}

// Pattern returns the original pattern string before regex conversion.
// This is useful for debugging and display purposes.
func (f *PatternFilter) Pattern() string {
	return f.raw
}

// TypeFilter matches events by exact type comparison.
// It maintains a set of allowed event types for efficient matching.
type TypeFilter struct {
	types map[domain.EventType]bool
}

// NewTypeFilter creates a filter that matches specific event types.
// Only events with types in the provided list will match.
//
// Parameters:
//   - types: One or more event types to match
//
// Returns a new TypeFilter instance.
func NewTypeFilter(types ...domain.EventType) *TypeFilter {
	typeMap := make(map[domain.EventType]bool)
	for _, t := range types {
		typeMap[t] = true
	}
	return &TypeFilter{types: typeMap}
}

// Match implements the EventFilter interface.
// It returns true if the event type is in the allowed set.
func (f *TypeFilter) Match(event domain.Event) bool {
	return f.types[event.Type]
}

// AgentFilter matches events by agent ID or name.
// Both criteria can be specified; if both are provided, both must match.
type AgentFilter struct {
	agentID   string
	agentName string
}

// NewAgentFilter creates a filter that matches events from specific agents.
// Either agentID or agentName can be empty to ignore that criterion.
//
// Parameters:
//   - agentID: The agent ID to match (empty string ignores ID)
//   - agentName: The agent name to match (empty string ignores name)
//
// Returns a new AgentFilter instance.
func NewAgentFilter(agentID, agentName string) *AgentFilter {
	return &AgentFilter{
		agentID:   agentID,
		agentName: agentName,
	}
}

// Match implements the EventFilter interface.
// It returns true if the event matches all non-empty criteria.
func (f *AgentFilter) Match(event domain.Event) bool {
	if f.agentID != "" && event.AgentID != f.agentID {
		return false
	}
	if f.agentName != "" && event.AgentName != f.agentName {
		return false
	}
	return true
}

// ErrorFilter matches error events.
// It identifies events that represent errors or failures.
type ErrorFilter struct{}

// NewErrorFilter creates a filter that matches error events.
// This filter uses the event's IsError() method to determine matches.
//
// Returns a new ErrorFilter instance.
func NewErrorFilter() *ErrorFilter {
	return &ErrorFilter{}
}

// Match implements the EventFilter interface.
// It returns true if the event represents an error.
func (f *ErrorFilter) Match(event domain.Event) bool {
	return event.IsError()
}

// MetadataFilter matches events by metadata key-value pairs.
// It performs exact matching on metadata values.
type MetadataFilter struct {
	key   string
	value interface{}
}

// NewMetadataFilter creates a filter that matches events with specific metadata.
// The filter checks for exact equality of the metadata value.
//
// Parameters:
//   - key: The metadata key to check
//   - value: The expected value for the key
//
// Returns a new MetadataFilter instance.
func NewMetadataFilter(key string, value interface{}) *MetadataFilter {
	return &MetadataFilter{
		key:   key,
		value: value,
	}
}

// Match implements the EventFilter interface.
// It returns true if the event has the specified metadata key with the expected value.
func (f *MetadataFilter) Match(event domain.Event) bool {
	if event.Metadata == nil {
		return false
	}
	val, exists := event.Metadata[f.key]
	if !exists {
		return false
	}
	return val == f.value
}

// CompositeFilter combines multiple filters with logic operators.
// It allows building complex filter expressions using AND, OR, and NOT operations.
type CompositeFilter struct {
	operator CompositeOperator
	filters  []EventFilter
}

// CompositeOperator defines how filters are combined in a CompositeFilter.
// It determines the logic operation applied to child filters.
type CompositeOperator int

const (
	// OperatorAND requires all filters to match
	OperatorAND CompositeOperator = iota
	// OperatorOR requires at least one filter to match
	OperatorOR
	// OperatorNOT inverts the match result
	OperatorNOT
)

// NewCompositeFilter creates a filter that combines multiple filters.
// The operator determines how the filters are combined.
//
// Parameters:
//   - operator: The logic operator (AND, OR, NOT)
//   - filters: The filters to combine
//
// Returns a new CompositeFilter instance.
func NewCompositeFilter(operator CompositeOperator, filters ...EventFilter) *CompositeFilter {
	return &CompositeFilter{
		operator: operator,
		filters:  filters,
	}
}

// Match implements the EventFilter interface.
// It applies the logic operator to combine results from child filters.
// For NOT operator, only the first filter is considered.
func (f *CompositeFilter) Match(event domain.Event) bool {
	switch f.operator {
	case OperatorAND:
		for _, filter := range f.filters {
			if !filter.Match(event) {
				return false
			}
		}
		return true

	case OperatorOR:
		for _, filter := range f.filters {
			if filter.Match(event) {
				return true
			}
		}
		return false

	case OperatorNOT:
		if len(f.filters) == 0 {
			return true
		}
		// NOT applies to the first filter only
		return !f.filters[0].Match(event)

	default:
		return false
	}
}

// AND creates a composite filter with AND logic.
// All provided filters must match for the composite to match.
//
// Parameters:
//   - filters: Filters to combine with AND logic
//
// Returns a new CompositeFilter with AND operator.
func AND(filters ...EventFilter) *CompositeFilter {
	return NewCompositeFilter(OperatorAND, filters...)
}

// OR creates a composite filter with OR logic.
// At least one filter must match for the composite to match.
//
// Parameters:
//   - filters: Filters to combine with OR logic
//
// Returns a new CompositeFilter with OR operator.
func OR(filters ...EventFilter) *CompositeFilter {
	return NewCompositeFilter(OperatorOR, filters...)
}

// NOT creates a composite filter with NOT logic.
// The filter inverts the match result of the provided filter.
//
// Parameters:
//   - filter: The filter to invert
//
// Returns a new CompositeFilter with NOT operator.
func NOT(filter EventFilter) *CompositeFilter {
	return NewCompositeFilter(OperatorNOT, filter)
}

// FieldFilter matches events by field values using reflection.
// Currently supports a limited set of top-level event fields.
type FieldFilter struct {
	fieldPath string
	value     interface{}
	operator  FieldOperator
}

// FieldOperator defines comparison operators for field matching.
// It determines how field values are compared in FieldFilter.
type FieldOperator int

const (
	// OpEqual checks if field equals value
	OpEqual FieldOperator = iota
	// OpNotEqual checks if field does not equal value
	OpNotEqual
	// OpContains checks if field contains value (for strings)
	OpContains
	// OpGreaterThan checks if field is greater than value
	OpGreaterThan
	// OpLessThan checks if field is less than value
	OpLessThan
)

// NewFieldFilter creates a filter that matches events by field value.
// Currently supports top-level fields: Type, AgentID, AgentName, ID.
//
// Parameters:
//   - fieldPath: The field to check (e.g., "Type", "AgentID")
//   - operator: The comparison operator
//   - value: The value to compare against
//
// Returns a new FieldFilter instance.
func NewFieldFilter(fieldPath string, operator FieldOperator, value interface{}) *FieldFilter {
	return &FieldFilter{
		fieldPath: fieldPath,
		value:     value,
		operator:  operator,
	}
}

// Match implements the EventFilter interface.
// It compares the specified field value using the configured operator.
// Currently limited to top-level event fields.
func (f *FieldFilter) Match(event domain.Event) bool {
	// This is a simplified implementation
	// In a full implementation, you would use reflection to navigate the field path
	// For now, we'll just support top-level fields

	var fieldValue interface{}

	switch f.fieldPath {
	case "Type":
		fieldValue = event.Type
	case "AgentID":
		fieldValue = event.AgentID
	case "AgentName":
		fieldValue = event.AgentName
	case "ID":
		fieldValue = event.ID
	default:
		return false
	}

	switch f.operator {
	case OpEqual:
		return fieldValue == f.value
	case OpNotEqual:
		return fieldValue != f.value
	case OpContains:
		if str, ok := fieldValue.(string); ok {
			if searchStr, ok := f.value.(string); ok {
				return strings.Contains(str, searchStr)
			}
		}
	}

	return false
}

// TimeRangeFilter matches events within a time range.
// It supports both absolute time.Time values and relative time.Duration values.
type TimeRangeFilter struct {
	start interface{} // time.Time or time.Duration (relative to now)
	end   interface{} // time.Time or time.Duration (relative to now)
}

// NewTimeRangeFilter creates a filter that matches events within a time range.
// Start and end can be either time.Time for absolute times or time.Duration
// for times relative to now.
//
// Parameters:
//   - start: Start of time range (time.Time or time.Duration)
//   - end: End of time range (time.Time or time.Duration)
//
// Returns a new TimeRangeFilter instance.
func NewTimeRangeFilter(start, end interface{}) *TimeRangeFilter {
	return &TimeRangeFilter{
		start: start,
		end:   end,
	}
}

// Match implements the EventFilter interface.
// Note: This is a simplified placeholder implementation that always returns true.
// A full implementation would properly handle time comparisons.
func (f *TimeRangeFilter) Match(event domain.Event) bool {
	// Simplified implementation - would need proper time handling
	// This is a placeholder that always returns true
	return true
}
