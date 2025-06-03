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

// Span represents a tracing span
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

// Tracer creates spans
type Tracer interface {
	// Start creates a new span
	Start(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span)
}

// Attribute represents a key-value attribute
type Attribute struct {
	Key   string
	Value interface{}
}

// StatusCode represents span status
type StatusCode int

const (
	StatusCodeUnset StatusCode = iota
	StatusCodeOk
	StatusCodeError
)

// SpanOption configures a span
type SpanOption interface {
	Apply(*SpanConfig)
}

// SpanConfig holds span configuration
type SpanConfig struct {
	Attributes []Attribute
}

// WithAttributes returns a SpanOption that sets attributes
func WithAttributes(attrs ...Attribute) SpanOption {
	return &withAttributes{attrs: attrs}
}

type withAttributes struct {
	attrs []Attribute
}

func (w *withAttributes) Apply(cfg *SpanConfig) {
	cfg.Attributes = append(cfg.Attributes, w.attrs...)
}

// TracingHook provides tracing integration for agent lifecycle
type TracingHook struct {
	tracer     Tracer
	tracerName string
}

// NewTracingHook creates a new tracing hook
func NewTracingHook(tracerName string, tracer Tracer) *TracingHook {
	return &TracingHook{
		tracer:     tracer,
		tracerName: tracerName,
	}
}

// BeforeRun starts a new span for agent execution
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

// AfterRun completes the span and records results
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

// BeforeInitialize traces agent initialization
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

// AfterInitialize completes initialization span
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

// BeforeCleanup traces agent cleanup
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

// AfterCleanup completes cleanup span
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

// ToolCallTracingHook traces tool calls
type ToolCallTracingHook struct {
	tracer     Tracer
	tracerName string
}

// NewToolCallTracingHook creates a new tool call tracing hook
func NewToolCallTracingHook(tracerName string, tracer Tracer) *ToolCallTracingHook {
	return &ToolCallTracingHook{
		tracer:     tracer,
		tracerName: tracerName,
	}
}

// BeforeToolCall starts a span for tool execution
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

// AfterToolCall completes the tool call span
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

// EventTracingHook traces event dispatch
type EventTracingHook struct {
	tracer     Tracer
	tracerName string
}

// NewEventTracingHook creates a new event tracing hook
func NewEventTracingHook(tracerName string, tracer Tracer) *EventTracingHook {
	return &EventTracingHook{
		tracer:     tracer,
		tracerName: tracerName,
	}
}

// HandleEvent traces event handling
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

// CompositeTracingHook combines multiple tracing aspects
type CompositeTracingHook struct {
	agentHook    *TracingHook
	toolHook     *ToolCallTracingHook
	eventHandler domain.EventHandler
}

// NewCompositeTracingHook creates a comprehensive tracing solution
func NewCompositeTracingHook(tracerName string, tracer Tracer) *CompositeTracingHook {
	return &CompositeTracingHook{
		agentHook:    NewTracingHook(tracerName, tracer),
		toolHook:     NewToolCallTracingHook(tracerName, tracer),
		eventHandler: NewEventTracingHook(tracerName, tracer),
	}
}

// GetAgentHook returns the agent tracing hook
func (c *CompositeTracingHook) GetAgentHook() *TracingHook {
	return c.agentHook
}

// GetToolHook returns the tool tracing hook
func (c *CompositeTracingHook) GetToolHook() *ToolCallTracingHook {
	return c.toolHook
}

// GetEventHandler returns the event tracing handler
func (c *CompositeTracingHook) GetEventHandler() domain.EventHandler {
	return c.eventHandler
}

// Context management

type spanKey struct{}

// ContextWithSpan returns a new context with the span attached
func ContextWithSpan(ctx context.Context, span Span) context.Context {
	return context.WithValue(ctx, spanKey{}, span)
}

// SpanFromContext returns the span from the context
func SpanFromContext(ctx context.Context) Span {
	if span, ok := ctx.Value(spanKey{}).(Span); ok {
		return span
	}
	return nil
}

// Helper functions for span attributes

// AddStateAttributes adds state information to a span
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

// AddAgentAttributes adds agent information to a span
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

// AddErrorAttributes adds error information to a span
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

// NoOpTracer is a no-op implementation of Tracer
type NoOpTracer struct{}

// Start returns a no-op span
func (t *NoOpTracer) Start(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span) {
	return ctx, &NoOpSpan{}
}

// NoOpSpan is a no-op implementation of Span
type NoOpSpan struct{}

func (s *NoOpSpan) End()                                          {}
func (s *NoOpSpan) SetAttributes(attributes ...Attribute)         {}
func (s *NoOpSpan) RecordError(err error)                         {}
func (s *NoOpSpan) SetStatus(code StatusCode, description string) {}
func (s *NoOpSpan) IsRecording() bool                             { return false }

// MetricsHook collects metrics during agent execution
type MetricsHook struct {
	executionTimes map[string][]time.Duration
	errorCounts    map[string]int
	mu             sync.RWMutex
}

// NewMetricsHook creates a new metrics collection hook
func NewMetricsHook() *MetricsHook {
	return &MetricsHook{
		executionTimes: make(map[string][]time.Duration),
		errorCounts:    make(map[string]int),
	}
}

// RecordExecution records execution time for an agent
func (h *MetricsHook) RecordExecution(agentName string, duration time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.executionTimes[agentName] = append(h.executionTimes[agentName], duration)
}

// RecordError records an error for an agent
func (h *MetricsHook) RecordError(agentName string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.errorCounts[agentName]++
}

// GetMetrics returns collected metrics
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
