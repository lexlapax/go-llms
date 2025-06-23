// ABOUTME: Provides tracing integration interfaces for agent execution monitoring
// ABOUTME: Defines hooks for distributed tracing of agent workflows and tool calls

package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// Span represents a tracing span for distributed tracing integration.
// It tracks the execution of a specific operation and can record
// attributes, errors, and status information.
type Span interface {
	// End completes the span
	End()

	// SetAttributes sets span attributes
	SetAttributes(attributes ...Attribute)

	// RecordError records an error
	RecordError(err error)

	// SetStatus sets the span status
	SetStatus(code StatusCode, description string)

	// IsRecording returns true if span is recording
	IsRecording() bool
}

// Tracer creates spans for distributed tracing.
// It provides the main entry point for instrumenting code with tracing
// and integrates with OpenTelemetry or similar tracing systems.
type Tracer interface {
	// Start creates a new span
	Start(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span)
}

// Attribute represents a key-value attribute for spans.
// Attributes provide additional context and metadata about the operation
// being traced, such as agent IDs, message counts, or error details.
type Attribute struct {
	Key   string
	Value interface{}
}

// StatusCode represents the status of a span.
// It indicates whether the operation completed successfully or encountered errors.
type StatusCode int

const (
	StatusCodeUnset StatusCode = iota
	StatusCodeOk
	StatusCodeError
)

// SpanOption configures a span during creation.
// Options can set initial attributes or other span properties.
type SpanOption interface {
	Apply(*SpanConfig)
}

// SpanConfig holds span configuration options.
// It is modified by SpanOptions during span creation.
type SpanConfig struct {
	Attributes []Attribute
}

// WithAttributes returns a SpanOption that sets attributes on a span.
// These attributes are added when the span is created, providing
// initial context about the operation being traced.
func WithAttributes(attrs ...Attribute) SpanOption {
	return &withAttributes{attrs: attrs}
}

type withAttributes struct {
	attrs []Attribute
}

func (w *withAttributes) Apply(cfg *SpanConfig) {
	cfg.Attributes = append(cfg.Attributes, w.attrs...)
}

// TracingHook provides tracing integration for agent lifecycle.
// It creates spans for agent operations like initialization, execution,
// and cleanup, enabling distributed tracing of agent workflows.
type TracingHook struct {
	tracer     Tracer
	tracerName string
}

// NewTracingHook creates a new tracing hook with the specified tracer.
// The tracerName identifies the instrumentation library in trace data.
// The hook will create spans for agent lifecycle events.
func NewTracingHook(tracerName string, tracer Tracer) *TracingHook {
	return &TracingHook{
		tracer:     tracer,
		tracerName: tracerName,
	}
}

// BeforeRun starts a new span for agent execution.
// It creates a span named "agent.<name>.run" with attributes about the agent and state.
// The span is added to the context for use by nested operations.
func (h *TracingHook) BeforeRun(ctx context.Context, agent domain.BaseAgent, state *domain.State) (context.Context, error) {
	ctx, span := h.tracer.Start(ctx, fmt.Sprintf("agent.%s.run", agent.Name()),
		WithAttributes(
			Attribute{Key: "agent.id", Value: agent.ID()},
			Attribute{Key: "agent.type", Value: string(agent.Type())},
			Attribute{Key: "agent.name", Value: agent.Name()},
			Attribute{Key: "state.id", Value: state.ID()},
		),
	)

	// Add state size as attribute
	span.SetAttributes(
		Attribute{Key: "state.values.count", Value: len(state.Values())},
		Attribute{Key: "state.messages.count", Value: len(state.Messages())},
		Attribute{Key: "state.artifacts.count", Value: len(state.Artifacts())},
	)

	return ctx, nil
}

// AfterRun completes the execution span and records results.
// It sets the span status based on success/failure and adds attributes
// about the result state if the operation succeeded.
func (h *TracingHook) AfterRun(ctx context.Context, agent domain.BaseAgent, state *domain.State, result *domain.State, err error) error {
	span := SpanFromContext(ctx)
	if span == nil || !span.IsRecording() {
		return nil
	}

	if err != nil {
		span.RecordError(err)
		span.SetStatus(StatusCodeError, err.Error())
	} else {
		span.SetStatus(StatusCodeOk, "")
		if result != nil {
			span.SetAttributes(
				Attribute{Key: "result.id", Value: result.ID()},
				Attribute{Key: "result.values.count", Value: len(result.Values())},
				Attribute{Key: "result.messages.count", Value: len(result.Messages())},
				Attribute{Key: "result.artifacts.count", Value: len(result.Artifacts())},
			)
		}
	}

	span.End()
	return nil
}

// BeforeInitialize starts a span for agent initialization.
// It creates a span named "agent.<name>.initialize" with agent attributes.
// The span tracks the initialization process duration and success.
func (h *TracingHook) BeforeInitialize(ctx context.Context, agent domain.BaseAgent) (context.Context, error) {
	ctx, _ = h.tracer.Start(ctx, fmt.Sprintf("agent.%s.initialize", agent.Name()),
		WithAttributes(
			Attribute{Key: "agent.id", Value: agent.ID()},
			Attribute{Key: "agent.type", Value: string(agent.Type())},
			Attribute{Key: "agent.name", Value: agent.Name()},
		),
	)
	return ctx, nil
}

// AfterInitialize completes the initialization span.
// It records any initialization errors and sets the appropriate span status.
// The span is ended regardless of success or failure.
func (h *TracingHook) AfterInitialize(ctx context.Context, agent domain.BaseAgent, err error) error {
	span := SpanFromContext(ctx)
	if span == nil || !span.IsRecording() {
		return nil
	}

	if err != nil {
		span.RecordError(err)
		span.SetStatus(StatusCodeError, err.Error())
	} else {
		span.SetStatus(StatusCodeOk, "")
	}

	span.End()
	return nil
}

// BeforeCleanup starts a span for agent cleanup.
// It creates a span named "agent.<name>.cleanup" to track resource cleanup.
// This helps identify cleanup performance and potential resource leaks.
func (h *TracingHook) BeforeCleanup(ctx context.Context, agent domain.BaseAgent) (context.Context, error) {
	ctx, _ = h.tracer.Start(ctx, fmt.Sprintf("agent.%s.cleanup", agent.Name()),
		WithAttributes(
			Attribute{Key: "agent.id", Value: agent.ID()},
			Attribute{Key: "agent.type", Value: string(agent.Type())},
			Attribute{Key: "agent.name", Value: agent.Name()},
		),
	)
	return ctx, nil
}

// AfterCleanup completes the cleanup span.
// It records any cleanup errors, helping identify resource cleanup failures.
// The span is ended regardless of cleanup success.
func (h *TracingHook) AfterCleanup(ctx context.Context, agent domain.BaseAgent, err error) error {
	span := SpanFromContext(ctx)
	if span == nil || !span.IsRecording() {
		return nil
	}

	if err != nil {
		span.RecordError(err)
		span.SetStatus(StatusCodeError, err.Error())
	} else {
		span.SetStatus(StatusCodeOk, "")
	}

	span.End()
	return nil
}

// ToolCallTracingHook traces tool calls made by agents.
// It creates spans for each tool invocation, tracking parameters,
// execution time, and results or errors.
type ToolCallTracingHook struct {
	tracer     Tracer
	tracerName string
}

// NewToolCallTracingHook creates a new tool call tracing hook.
// The hook creates spans named "tool.<name>.call" for each tool invocation,
// enabling performance analysis of tool usage.
func NewToolCallTracingHook(tracerName string, tracer Tracer) *ToolCallTracingHook {
	return &ToolCallTracingHook{
		tracer:     tracer,
		tracerName: tracerName,
	}
}

// BeforeToolCall starts a span for tool execution.
// It records the tool name and attempts to extract parameter count if possible.
// The span tracks tool execution duration and success.
func (h *ToolCallTracingHook) BeforeToolCall(ctx context.Context, toolName string, params interface{}) (context.Context, error) {
	ctx, span := h.tracer.Start(ctx, fmt.Sprintf("tool.%s.call", toolName),
		WithAttributes(
			Attribute{Key: "tool.name", Value: toolName},
		),
	)

	// Try to add parameter count
	if paramsMap, ok := params.(map[string]interface{}); ok {
		span.SetAttributes(
			Attribute{Key: "tool.params.count", Value: len(paramsMap)},
		)
	}

	return ctx, nil
}

// AfterToolCall completes the tool call span.
// It records the result type (if successful) or error details (if failed).
// This helps identify tool performance and reliability issues.
func (h *ToolCallTracingHook) AfterToolCall(ctx context.Context, toolName string, result interface{}, err error) error {
	span := SpanFromContext(ctx)
	if span == nil || !span.IsRecording() {
		return nil
	}

	span.SetAttributes(
		Attribute{Key: "tool.name", Value: toolName},
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(StatusCodeError, err.Error())
	} else {
		span.SetStatus(StatusCodeOk, "")

		// Try to add result type
		if result != nil {
			span.SetAttributes(
				Attribute{Key: "tool.result.type", Value: fmt.Sprintf("%T", result)},
			)
		}
	}

	span.End()
	return nil
}

// EventTracingHook traces event dispatch in the agent system.
// It creates spans for each event, tracking event flow through the system
// and helping identify event processing bottlenecks.
type EventTracingHook struct {
	tracer     Tracer
	tracerName string
}

// NewEventTracingHook creates a new event tracing hook.
// The hook creates spans named "event.<type>" for each event processed,
// providing visibility into the event-driven aspects of agent execution.
func NewEventTracingHook(tracerName string, tracer Tracer) *EventTracingHook {
	return &EventTracingHook{
		tracer:     tracer,
		tracerName: tracerName,
	}
}

// HandleEvent implements domain.EventHandler with tracing.
// It creates a span for the event with attributes about the event type,
// agent, and any associated error information.
func (h *EventTracingHook) HandleEvent(event domain.Event) error {
	ctx := context.Background()
	_, span := h.tracer.Start(ctx, fmt.Sprintf("event.%s", event.Type),
		WithAttributes(
			Attribute{Key: "event.id", Value: event.ID},
			Attribute{Key: "event.type", Value: string(event.Type)},
			Attribute{Key: "event.agent.id", Value: event.AgentID},
			Attribute{Key: "event.agent.name", Value: event.AgentName},
			Attribute{Key: "event.timestamp", Value: event.Timestamp.String()},
		),
	)
	defer span.End()

	if event.Error != nil {
		span.RecordError(event.Error)
		span.SetStatus(StatusCodeError, event.Error.Error())
	} else {
		span.SetStatus(StatusCodeOk, "")
	}

	return nil
}

// CompositeTracingHook combines multiple tracing aspects into one.
// It provides agent lifecycle tracing, tool call tracing, and event tracing
// through a single convenient interface.
type CompositeTracingHook struct {
	agentHook    *TracingHook
	toolHook     *ToolCallTracingHook
	eventHandler domain.EventHandler
}

// NewCompositeTracingHook creates a comprehensive tracing solution.
// It initializes hooks for agent operations, tool calls, and events,
// providing complete tracing coverage for agent systems.
func NewCompositeTracingHook(tracerName string, tracer Tracer) *CompositeTracingHook {
	return &CompositeTracingHook{
		agentHook:    NewTracingHook(tracerName, tracer),
		toolHook:     NewToolCallTracingHook(tracerName, tracer),
		eventHandler: NewEventTracingHook(tracerName, tracer),
	}
}

// GetAgentHook returns the agent tracing hook.
// Use this hook to trace agent lifecycle operations like
// initialization, execution, and cleanup.
func (c *CompositeTracingHook) GetAgentHook() *TracingHook {
	return c.agentHook
}

// GetToolHook returns the tool tracing hook.
// Use this hook to trace tool invocations made by agents,
// including parameters and results.
func (c *CompositeTracingHook) GetToolHook() *ToolCallTracingHook {
	return c.toolHook
}

// GetEventHandler returns the event tracing handler.
// Use this handler to trace events flowing through the agent system,
// providing visibility into asynchronous operations.
func (c *CompositeTracingHook) GetEventHandler() domain.EventHandler {
	return c.eventHandler
}

// Context management

type spanKey struct{}

// ContextWithSpan returns a new context with the span attached.
// The span can be retrieved later using SpanFromContext.
// This enables passing spans through call chains without explicit parameters.
func ContextWithSpan(ctx context.Context, span Span) context.Context {
	return context.WithValue(ctx, spanKey{}, span)
}

// SpanFromContext returns the span from the context.
// Returns nil if no span is present in the context.
// Use this to access the current span for adding attributes or events.
func SpanFromContext(ctx context.Context) Span {
	if span, ok := ctx.Value(spanKey{}).(Span); ok {
		return span
	}
	return nil
}

// Helper functions for span attributes

// AddStateAttributes adds state information to a span.
// It records the state ID, version, and counts of values, messages, and artifacts.
// Does nothing if span is nil or not recording.
func AddStateAttributes(span Span, state *domain.State) {
	if state == nil || span == nil || !span.IsRecording() {
		return
	}

	span.SetAttributes(
		Attribute{Key: "state.id", Value: state.ID()},
		Attribute{Key: "state.version", Value: state.Version()},
		Attribute{Key: "state.values.count", Value: len(state.Values())},
		Attribute{Key: "state.messages.count", Value: len(state.Messages())},
		Attribute{Key: "state.artifacts.count", Value: len(state.Artifacts())},
	)
}

// AddAgentAttributes adds agent information to a span.
// It records agent ID, name, type, description, parent info, and sub-agent count.
// Does nothing if span is nil or not recording.
func AddAgentAttributes(span Span, agent domain.BaseAgent) {
	if agent == nil || span == nil || !span.IsRecording() {
		return
	}

	span.SetAttributes(
		Attribute{Key: "agent.id", Value: agent.ID()},
		Attribute{Key: "agent.name", Value: agent.Name()},
		Attribute{Key: "agent.type", Value: string(agent.Type())},
		Attribute{Key: "agent.description", Value: agent.Description()},
	)

	// Add parent info if available
	if parent := agent.Parent(); parent != nil {
		span.SetAttributes(
			Attribute{Key: "agent.parent.id", Value: parent.ID()},
			Attribute{Key: "agent.parent.name", Value: parent.Name()},
		)
	}

	// Add sub-agent count
	span.SetAttributes(
		Attribute{Key: "agent.subagents.count", Value: len(agent.SubAgents())},
	)
}

// AddErrorAttributes adds error information to a span.
// It records the error and extracts additional attributes for known error types
// like AgentError, ToolError, and ValidationError.
func AddErrorAttributes(span Span, err error) {
	if err == nil || span == nil || !span.IsRecording() {
		return
	}

	span.RecordError(err)

	// Check for specific error types
	switch e := err.(type) {
	case *domain.AgentError:
		span.SetAttributes(
			Attribute{Key: "error.type", Value: "agent_error"},
			Attribute{Key: "error.agent.id", Value: e.AgentID},
			Attribute{Key: "error.agent.name", Value: e.AgentName},
			Attribute{Key: "error.phase", Value: e.Phase},
		)
	case *domain.ToolError:
		span.SetAttributes(
			Attribute{Key: "error.type", Value: "tool_error"},
			Attribute{Key: "error.tool.name", Value: e.ToolName},
			Attribute{Key: "error.phase", Value: e.Phase},
		)
	case *domain.ValidationError:
		span.SetAttributes(
			Attribute{Key: "error.type", Value: "validation_error"},
			Attribute{Key: "error.field", Value: e.Field},
		)
	default:
		span.SetAttributes(
			Attribute{Key: "error.type", Value: fmt.Sprintf("%T", err)},
		)
	}
}

// NoOpTracer is a no-op implementation of Tracer.
// It creates no-op spans that do nothing, useful for testing
// or when tracing is disabled.
type NoOpTracer struct{}

// Start returns a no-op span that performs no operations.
// The returned context is unchanged from the input context.
// All operations on the returned span are no-ops.
func (t *NoOpTracer) Start(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span) {
	return ctx, &NoOpSpan{}
}

// NoOpSpan is a no-op implementation of Span.
// All methods do nothing and IsRecording always returns false.
// Useful for testing or when tracing is disabled.
type NoOpSpan struct{}

func (s *NoOpSpan) End()                                          {}
func (s *NoOpSpan) SetAttributes(attributes ...Attribute)         {}
func (s *NoOpSpan) RecordError(err error)                         {}
func (s *NoOpSpan) SetStatus(code StatusCode, description string) {}
func (s *NoOpSpan) IsRecording() bool                             { return false }

// MetricsHook collects metrics during agent execution.
// It tracks execution times and error counts for agents,
// providing basic performance monitoring capabilities.
type MetricsHook struct {
	executionTimes map[string][]time.Duration
	errorCounts    map[string]int
	mu             sync.RWMutex
}

// NewMetricsHook creates a new metrics collection hook.
// The hook starts with empty metrics that accumulate as agents execute.
// Use GetMetrics to retrieve collected statistics.
func NewMetricsHook() *MetricsHook {
	return &MetricsHook{
		executionTimes: make(map[string][]time.Duration),
		errorCounts:    make(map[string]int),
	}
}

// RecordExecution records execution time for an agent.
// Multiple executions are tracked to calculate average execution times.
// Thread-safe for concurrent use.
func (h *MetricsHook) RecordExecution(agentName string, duration time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.executionTimes[agentName] = append(h.executionTimes[agentName], duration)
}

// RecordError records an error for an agent.
// Increments the error count for the specified agent.
// Thread-safe for concurrent use.
func (h *MetricsHook) RecordError(agentName string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.errorCounts[agentName]++
}

// GetMetrics returns collected metrics as a map.
// Includes average execution times and error counts per agent.
// The returned map is safe to modify without affecting internal state.
func (h *MetricsHook) GetMetrics() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	metrics := make(map[string]interface{})

	// Calculate average execution times
	avgTimes := make(map[string]float64)
	for agent, times := range h.executionTimes {
		if len(times) > 0 {
			var total time.Duration
			for _, t := range times {
				total += t
			}
			avgTimes[agent] = float64(total) / float64(len(times))
		}
	}

	metrics["average_execution_times"] = avgTimes
	metrics["error_counts"] = h.errorCounts

	return metrics
}
