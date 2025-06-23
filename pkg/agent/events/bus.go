// ABOUTME: EventBus implementation for event subscription, publishing, and filtering
// ABOUTME: Provides thread-safe event distribution with pattern matching and bridge integration support

package events

import (
	"context"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// EventBus manages event subscriptions and distribution.
// It provides thread-safe event publishing and subscription with pattern matching,
// filtering, and concurrent event handling. The bus supports both synchronous
// and asynchronous event delivery with configurable buffer sizes.
type EventBus struct {
	mu            sync.RWMutex
	subscriptions map[string]*subscription
	bufferSize    int
	closed        bool
	closeOnce     sync.Once
	wg            sync.WaitGroup
}

// subscription represents an event subscription.
// It contains the handler, filters, and pattern matching configuration
// for processing events delivered through a buffered channel.
type subscription struct {
	id         string
	handler    EventHandler
	filters    []EventFilter
	channel    chan domain.Event
	pattern    *regexp.Regexp
	patternStr string
	ctx        context.Context
	cancel     context.CancelFunc
}

// EventHandler processes events with context support.
// Implementations should handle events efficiently and return
// errors for any processing failures.
type EventHandler interface {
	HandleEvent(ctx context.Context, event domain.Event) error
}

// EventHandlerFunc is a function adapter for EventHandler.
// It allows regular functions to be used as EventHandler implementations.
type EventHandlerFunc func(ctx context.Context, event domain.Event) error

// HandleEvent implements the EventHandler interface.
// It calls the underlying function with the provided context and event.
func (f EventHandlerFunc) HandleEvent(ctx context.Context, event domain.Event) error {
	return f(ctx, event)
}

// EventFilter filters events with enhanced matching.
// Filters are applied after pattern matching to provide
// fine-grained control over which events a handler receives.
type EventFilter interface {
	Match(event domain.Event) bool
}

// EventFilterFunc is a function adapter for EventFilter.
// It allows regular functions to be used as EventFilter implementations.
type EventFilterFunc func(event domain.Event) bool

// Match implements the EventFilter interface.
// It calls the underlying function with the provided event.
func (f EventFilterFunc) Match(event domain.Event) bool {
	return f(event)
}

// EventBusOption configures the EventBus.
// Options can be passed to NewEventBus to customize bus behavior.
type EventBusOption func(*EventBus)

// WithBufferSize sets the buffer size for event channels.
// Larger buffers can prevent event loss under high load but consume more memory.
//
// Parameters:
//   - size: The number of events that can be buffered per subscription
//
// Returns an EventBusOption to configure the bus.
func WithBufferSize(size int) EventBusOption {
	return func(eb *EventBus) {
		eb.bufferSize = size
	}
}

// NewEventBus creates a new EventBus with the specified options.
// The default buffer size is 100 events per subscription.
//
// Parameters:
//   - opts: Optional configuration functions
//
// Returns a new EventBus instance.
func NewEventBus(opts ...EventBusOption) *EventBus {
	eb := &EventBus{
		subscriptions: make(map[string]*subscription),
		bufferSize:    100, // Default buffer size
	}

	for _, opt := range opts {
		opt(eb)
	}

	return eb
}

// Subscribe adds a handler with optional filters to the event bus.
// The handler will receive all events that pass the provided filters.
// Each subscription runs in its own goroutine for concurrent processing.
//
// Parameters:
//   - handler: The event handler to process matching events
//   - filters: Optional filters to apply before sending events to the handler
//
// Returns a unique subscription ID that can be used to unsubscribe.
func (eb *EventBus) Subscribe(handler EventHandler, filters ...EventFilter) string {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.closed {
		return ""
	}

	id := uuid.New().String()
	ctx, cancel := context.WithCancel(context.Background())

	sub := &subscription{
		id:      id,
		handler: handler,
		filters: filters,
		channel: make(chan domain.Event, eb.bufferSize),
		ctx:     ctx,
		cancel:  cancel,
	}

	eb.subscriptions[id] = sub

	// Start handler goroutine
	eb.wg.Add(1)
	go eb.handleSubscription(sub)

	return id
}

// SubscribePattern subscribes to events matching a pattern.
// Patterns use regular expression syntax to match event types.
// Common patterns include "tool.*" for all tool events or "agent.start" for specific events.
//
// Parameters:
//   - pattern: Regular expression pattern to match event types
//   - handler: The event handler to process matching events
//   - filters: Optional additional filters to apply
//
// Returns a subscription ID and nil error on success, or empty string and error if pattern is invalid.
func (eb *EventBus) SubscribePattern(pattern string, handler EventHandler, filters ...EventFilter) (string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("invalid pattern %q: %w", pattern, err)
	}

	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.closed {
		return "", fmt.Errorf("event bus is closed")
	}

	id := uuid.New().String()
	ctx, cancel := context.WithCancel(context.Background())

	sub := &subscription{
		id:         id,
		handler:    handler,
		filters:    filters,
		channel:    make(chan domain.Event, eb.bufferSize),
		pattern:    re,
		patternStr: pattern,
		ctx:        ctx,
		cancel:     cancel,
	}

	eb.subscriptions[id] = sub

	// Start handler goroutine
	eb.wg.Add(1)
	go eb.handleSubscription(sub)

	return id, nil
}

// Unsubscribe removes a subscription from the event bus.
// The subscription's handler goroutine will be terminated gracefully.
//
// Parameters:
//   - subscriptionID: The ID returned from Subscribe or SubscribePattern
func (eb *EventBus) Unsubscribe(subscriptionID string) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if sub, exists := eb.subscriptions[subscriptionID]; exists {
		sub.cancel()
		delete(eb.subscriptions, subscriptionID)
	}
}

// Publish sends an event to all matching subscribers.
// This is a convenience method that uses a background context.
// Events are delivered asynchronously to avoid blocking the publisher.
//
// Parameters:
//   - event: The event to publish
func (eb *EventBus) Publish(event domain.Event) {
	eb.PublishContext(context.Background(), event)
}

// PublishContext sends an event with context to all matching subscribers.
// The context can be used to cancel event delivery if needed.
// Events are sent to subscription channels without blocking; full channels result in dropped events.
//
// Parameters:
//   - ctx: Context for cancellation
//   - event: The event to publish
func (eb *EventBus) PublishContext(ctx context.Context, event domain.Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	if eb.closed {
		return
	}

	for _, sub := range eb.subscriptions {
		// Check if event matches subscription
		if !eb.matchesSubscription(event, sub) {
			continue
		}

		// Try to send event without blocking
		select {
		case sub.channel <- event:
			// Event sent successfully
		case <-ctx.Done():
			// Context canceled
			return
		default:
			// Channel full, drop event (could implement overflow handling here)
			// In production, you might want to log this or handle overflow differently
		}
	}
}

// Close shuts down the event bus gracefully.
// It cancels all subscriptions, waits for handlers to complete,
// and clears all internal state. This method is safe to call multiple times.
func (eb *EventBus) Close() {
	eb.closeOnce.Do(func() {
		eb.mu.Lock()
		eb.closed = true

		// Cancel all subscriptions
		for _, sub := range eb.subscriptions {
			sub.cancel()
		}
		eb.mu.Unlock()

		// Wait for all handlers to finish
		eb.wg.Wait()

		// Clear subscriptions
		eb.mu.Lock()
		eb.subscriptions = make(map[string]*subscription)
		eb.mu.Unlock()
	})
}

// GetSubscriptionCount returns the number of active subscriptions.
// This is useful for monitoring and debugging event bus usage.
//
// Returns the count of active subscriptions.
func (eb *EventBus) GetSubscriptionCount() int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return len(eb.subscriptions)
}

// GetSubscriptionInfo returns information about a subscription.
// This is useful for debugging and monitoring specific subscriptions.
//
// Parameters:
//   - subscriptionID: The ID of the subscription to query
//
// Returns:
//   - pattern: The pattern string if subscription uses pattern matching
//   - filterCount: Number of filters applied to the subscription
//   - found: Whether the subscription exists
func (eb *EventBus) GetSubscriptionInfo(subscriptionID string) (pattern string, filterCount int, found bool) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	sub, exists := eb.subscriptions[subscriptionID]
	if !exists {
		return "", 0, false
	}

	return sub.patternStr, len(sub.filters), true
}

// handleSubscription processes events for a subscription.
// It runs in a separate goroutine and handles events from the subscription's
// channel until the subscription is canceled. Each event is processed with
// a 30-second timeout to prevent hanging handlers.
func (eb *EventBus) handleSubscription(sub *subscription) {
	defer eb.wg.Done()

	for {
		select {
		case event := <-sub.channel:
			// Handle event with timeout
			ctx, cancel := context.WithTimeout(sub.ctx, 30*time.Second)
			err := sub.handler.HandleEvent(ctx, event)
			cancel()

			if err != nil {
				// TODO: Consider adding error handler callback or metrics
				_ = err // Acknowledge error for now
			}

		case <-sub.ctx.Done():
			// Subscription canceled
			return
		}
	}
}

// matchesSubscription checks if an event matches a subscription.
// It first checks pattern matching (if configured), then applies all filters.
// An event must pass both pattern and all filters to match.
//
// Parameters:
//   - event: The event to check
//   - sub: The subscription to match against
//
// Returns true if the event matches the subscription criteria.
func (eb *EventBus) matchesSubscription(event domain.Event, sub *subscription) bool {
	// Check pattern match if pattern is set
	if sub.pattern != nil {
		if !sub.pattern.MatchString(string(event.Type)) {
			return false
		}
	}

	// Check all filters
	for _, filter := range sub.filters {
		if !filter.Match(event) {
			return false
		}
	}

	return true
}

// defaultBus is the global default event bus instance.
// It provides a convenient shared bus for simple use cases.
var defaultBus = NewEventBus()

// GetDefaultBus returns the global default event bus.
// This bus can be used when a shared event bus is sufficient
// and creating a dedicated instance is not necessary.
//
// Returns the default EventBus instance.
func GetDefaultBus() *EventBus {
	return defaultBus
}

// Subscribe adds a handler to the default event bus.
// This is a convenience function that operates on the global bus.
//
// Parameters:
//   - handler: The event handler to process matching events
//   - filters: Optional filters to apply before sending events to the handler
//
// Returns a unique subscription ID that can be used to unsubscribe.
func Subscribe(handler EventHandler, filters ...EventFilter) string {
	return defaultBus.Subscribe(handler, filters...)
}

// SubscribePattern subscribes to events matching a pattern on the default bus.
// This is a convenience function that operates on the global bus.
//
// Parameters:
//   - pattern: Regular expression pattern to match event types
//   - handler: The event handler to process matching events
//   - filters: Optional additional filters to apply
//
// Returns a subscription ID and nil error on success, or empty string and error if pattern is invalid.
func SubscribePattern(pattern string, handler EventHandler, filters ...EventFilter) (string, error) {
	return defaultBus.SubscribePattern(pattern, handler, filters...)
}

// Unsubscribe removes a subscription from the default bus.
// This is a convenience function that operates on the global bus.
//
// Parameters:
//   - subscriptionID: The ID returned from Subscribe or SubscribePattern
func Unsubscribe(subscriptionID string) {
	defaultBus.Unsubscribe(subscriptionID)
}

// Publish sends an event to the default bus.
// This is a convenience function that operates on the global bus.
//
// Parameters:
//   - event: The event to publish
func Publish(event domain.Event) {
	defaultBus.Publish(event)
}

// PublishContext sends an event with context to the default bus.
// This is a convenience function that operates on the global bus.
//
// Parameters:
//   - ctx: Context for cancellation
//   - event: The event to publish
func PublishContext(ctx context.Context, event domain.Event) {
	defaultBus.PublishContext(ctx, event)
}
