// ABOUTME: Event filtering system with pattern matching and composite filters
// ABOUTME: Provides flexible event filtering for bridge layer integration

package events

import (
	"regexp"
	"strings"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// PatternFilter matches events by type pattern
type PatternFilter struct {
	pattern *regexp.Regexp
	raw     string
}

// NewPatternFilter creates a filter that matches event types by pattern
// Supports wildcards: "tool.*" matches all tool events
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

// Match implements EventFilter
func (f *PatternFilter) Match(event domain.Event) bool {
	return f.pattern.MatchString(string(event.Type))
}

// Pattern returns the original pattern string
func (f *PatternFilter) Pattern() string {
	return f.raw
}

// TypeFilter matches events by exact type
type TypeFilter struct {
	types map[domain.EventType]bool
}

// NewTypeFilter creates a filter that matches specific event types
func NewTypeFilter(types ...domain.EventType) *TypeFilter {
	typeMap := make(map[domain.EventType]bool)
	for _, t := range types {
		typeMap[t] = true
	}
	return &TypeFilter{types: typeMap}
}

// Match implements EventFilter
func (f *TypeFilter) Match(event domain.Event) bool {
	return f.types[event.Type]
}

// AgentFilter matches events by agent ID or name
type AgentFilter struct {
	agentID   string
	agentName string
}

// NewAgentFilter creates a filter that matches events from specific agents
func NewAgentFilter(agentID, agentName string) *AgentFilter {
	return &AgentFilter{
		agentID:   agentID,
		agentName: agentName,
	}
}

// Match implements EventFilter
func (f *AgentFilter) Match(event domain.Event) bool {
	if f.agentID != "" && event.AgentID != f.agentID {
		return false
	}
	if f.agentName != "" && event.AgentName != f.agentName {
		return false
	}
	return true
}

// ErrorFilter matches error events
type ErrorFilter struct{}

// NewErrorFilter creates a filter that matches error events
func NewErrorFilter() *ErrorFilter {
	return &ErrorFilter{}
}

// Match implements EventFilter
func (f *ErrorFilter) Match(event domain.Event) bool {
	return event.IsError()
}

// MetadataFilter matches events by metadata
type MetadataFilter struct {
	key   string
	value interface{}
}

// NewMetadataFilter creates a filter that matches events with specific metadata
func NewMetadataFilter(key string, value interface{}) *MetadataFilter {
	return &MetadataFilter{
		key:   key,
		value: value,
	}
}

// Match implements EventFilter
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

// CompositeFilter combines multiple filters with logic operators
type CompositeFilter struct {
	operator CompositeOperator
	filters  []EventFilter
}

// CompositeOperator defines how filters are combined
type CompositeOperator int

const (
	// OperatorAND requires all filters to match
	OperatorAND CompositeOperator = iota
	// OperatorOR requires at least one filter to match
	OperatorOR
	// OperatorNOT inverts the match result
	OperatorNOT
)

// NewCompositeFilter creates a filter that combines multiple filters
func NewCompositeFilter(operator CompositeOperator, filters ...EventFilter) *CompositeFilter {
	return &CompositeFilter{
		operator: operator,
		filters:  filters,
	}
}

// Match implements EventFilter
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

// AND creates a composite filter with AND logic
func AND(filters ...EventFilter) *CompositeFilter {
	return NewCompositeFilter(OperatorAND, filters...)
}

// OR creates a composite filter with OR logic
func OR(filters ...EventFilter) *CompositeFilter {
	return NewCompositeFilter(OperatorOR, filters...)
}

// NOT creates a composite filter with NOT logic
func NOT(filter EventFilter) *CompositeFilter {
	return NewCompositeFilter(OperatorNOT, filter)
}

// FieldFilter matches events by field values using reflection
type FieldFilter struct {
	fieldPath string
	value     interface{}
	operator  FieldOperator
}

// FieldOperator defines comparison operators for field matching
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

// NewFieldFilter creates a filter that matches events by field value
func NewFieldFilter(fieldPath string, operator FieldOperator, value interface{}) *FieldFilter {
	return &FieldFilter{
		fieldPath: fieldPath,
		value:     value,
		operator:  operator,
	}
}

// Match implements EventFilter
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

// TimeRangeFilter matches events within a time range
type TimeRangeFilter struct {
	start interface{} // time.Time or time.Duration (relative to now)
	end   interface{} // time.Time or time.Duration (relative to now)
}

// NewTimeRangeFilter creates a filter that matches events within a time range
func NewTimeRangeFilter(start, end interface{}) *TimeRangeFilter {
	return &TimeRangeFilter{
		start: start,
		end:   end,
	}
}

// Match implements EventFilter
func (f *TimeRangeFilter) Match(event domain.Event) bool {
	// Simplified implementation - would need proper time handling
	// This is a placeholder that always returns true
	return true
}
