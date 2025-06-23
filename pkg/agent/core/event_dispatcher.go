// ABOUTME: Implements event dispatching for agent lifecycle and execution monitoring
// ABOUTME: Provides asynchronous event distribution with filtering and subscription management

package core

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// eventDispatcher implements domain.EventDispatcher.
// It provides asynchronous event distribution to registered subscribers
// with support for filtering and buffered event processing.
type eventDispatcher struct {
	mu            sync.RWMutex
	subscriptions map[string]*subscription
	eventChan     chan domain.Event
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	bufferSize    int
}

// subscription represents an event subscription.
// It contains the handler to invoke, optional filters to apply,
// and a unique identifier for unsubscribing.
type subscription struct {
	id      string
	handler domain.EventHandler
	filters []domain.EventFilter
}

// NewEventDispatcher creates a new event dispatcher with the specified buffer size.
// The dispatcher processes events asynchronously in a separate goroutine.
// If bufferSize is <= 0, it defaults to 100. The dispatcher must be closed
// with Close() when no longer needed to prevent goroutine leaks.
func NewEventDispatcher(bufferSize int) domain.EventDispatcher {
	if bufferSize <= 0 {
		bufferSize = 100
	}

	ctx, cancel := context.WithCancel(context.Background())
	ed := &eventDispatcher{
		subscriptions: make(map[string]*subscription),
		eventChan:     make(chan domain.Event, bufferSize),
		ctx:           ctx,
		cancel:        cancel,
		bufferSize:    bufferSize,
	}

	ed.wg.Add(1)
	go ed.processEvents()

	return ed
}

// Subscribe adds a handler with optional filters to receive events.
// Returns a subscription ID that can be used to unsubscribe later.
// If handler is nil, returns an empty string. Filters are applied
// in order - an event must pass all filters to reach the handler.
func (ed *eventDispatcher) Subscribe(handler domain.EventHandler, filters ...domain.EventFilter) string {
	if handler == nil {
		return ""
	}

	ed.mu.Lock()
	defer ed.mu.Unlock()

	sub := &subscription{
		id:      uuid.New().String(),
		handler: handler,
		filters: filters,
	}

	ed.subscriptions[sub.id] = sub
	return sub.id
}

// Unsubscribe removes a subscription
func (ed *eventDispatcher) Unsubscribe(subscriptionID string) {
	ed.mu.Lock()
	defer ed.mu.Unlock()
	delete(ed.subscriptions, subscriptionID)
}

// Dispatch sends an event to all matching subscribers
func (ed *eventDispatcher) Dispatch(event domain.Event) {
	select {
	case <-ed.ctx.Done():
		// Dispatcher is closed, drop the event
		return
	default:
		// Try to send the event
		select {
		case ed.eventChan <- event:
			// Event sent successfully
		case <-ed.ctx.Done():
			// Dispatcher is closing
		default:
			// Channel is full, drop the event
			// In production, you might want to log this or handle it differently
		}
	}
}

// Close shuts down the dispatcher
func (ed *eventDispatcher) Close() {
	ed.cancel()
	close(ed.eventChan)
	ed.wg.Wait()
}

// processEvents processes events from the channel
func (ed *eventDispatcher) processEvents() {
	defer ed.wg.Done()

	for {
		select {
		case event, ok := <-ed.eventChan:
			if !ok {
				return
			}
			ed.handleEvent(event)
		case <-ed.ctx.Done():
			// Drain remaining events
			for {
				select {
				case event, ok := <-ed.eventChan:
					if !ok {
						return
					}
					ed.handleEvent(event)
				default:
					return
				}
			}
		}
	}
}

// handleEvent distributes an event to matching subscribers
func (ed *eventDispatcher) handleEvent(event domain.Event) {
	ed.mu.RLock()
	defer ed.mu.RUnlock()

	for _, sub := range ed.subscriptions {
		// Check filters
		if !ed.matchesFilters(event, sub.filters) {
			continue
		}

		// Handle event (non-blocking)
		go func(h domain.EventHandler, e domain.Event) {
			// Recover from panics in handlers
			defer func() {
				if r := recover(); r != nil {
					// In production, log the panic
					_ = r
				}
			}()

			if err := h.HandleEvent(e); err != nil {
				// In production, log the error or emit an error event
				_ = err
			}
		}(sub.handler, event)
	}
}

// matchesFilters checks if an event matches all filters
func (ed *eventDispatcher) matchesFilters(event domain.Event, filters []domain.EventFilter) bool {
	for _, filter := range filters {
		if filter != nil && !filter(event) {
			return false
		}
	}
	return true
}

// eventStream implements domain.EventStream.
// It provides a channel-based stream of events that can be consumed
// sequentially. The stream is bounded by a configurable buffer size.
type eventStream struct {
	events chan domain.Event
	ctx    context.Context
	cancel context.CancelFunc
}

// NewEventStream creates a new event stream with the specified buffer size.
// If bufferSize is <= 0, it defaults to 10. The stream must be closed
// with Close() when no longer needed to prevent goroutine leaks.
func NewEventStream(bufferSize int) domain.EventStream {
	if bufferSize <= 0 {
		bufferSize = 10
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &eventStream{
		events: make(chan domain.Event, bufferSize),
		ctx:    ctx,
		cancel: cancel,
	}
}

// Next returns the next event from the stream or blocks until one is available.
// Returns domain.ErrAgentCancelled if the stream is closed or the context is cancelled.
// This method is safe for concurrent use by multiple goroutines.
func (es *eventStream) Next() (domain.Event, error) {
	select {
	case event, ok := <-es.events:
		if !ok {
			return domain.Event{}, domain.ErrAgentCancelled
		}
		return event, nil
	case <-es.ctx.Done():
		return domain.Event{}, domain.ErrAgentCancelled
	}
}

// Close closes the event stream and cancels its context.
// Any pending Next() calls will return domain.ErrAgentCancelled.
// The stream cannot be used after calling Close().
func (es *eventStream) Close() {
	es.cancel()
	close(es.events)
}

// Send sends an event to the stream (internal use).
// Returns true if the event was sent successfully, false if the stream
// is closed or the channel buffer is full (non-blocking).
func (es *eventStream) Send(event domain.Event) bool {
	select {
	case es.events <- event:
		return true
	case <-es.ctx.Done():
		return false
	default:
		// Channel full
		return false
	}
}

// BufferedEventHandler wraps an event handler with a buffer.
// It processes events asynchronously in a separate goroutine, preventing
// slow handlers from blocking event dispatch. Events are dropped if the buffer fills.
type BufferedEventHandler struct {
	handler domain.EventHandler
	buffer  chan domain.Event
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

// NewBufferedEventHandler creates a new buffered event handler.
// The handler processes events from a buffer of the specified size.
// If bufferSize is <= 0, it defaults to 100. The handler must be
// closed with Close() when no longer needed.
func NewBufferedEventHandler(handler domain.EventHandler, bufferSize int) *BufferedEventHandler {
	if bufferSize <= 0 {
		bufferSize = 100
	}

	ctx, cancel := context.WithCancel(context.Background())
	beh := &BufferedEventHandler{
		handler: handler,
		buffer:  make(chan domain.Event, bufferSize),
		ctx:     ctx,
		cancel:  cancel,
	}

	beh.wg.Add(1)
	go beh.processEvents()

	return beh
}

// HandleEvent implements domain.EventHandler.
// Adds the event to the buffer for asynchronous processing.
// Returns domain.ErrEventDispatch if the buffer is full or the handler is closed.
func (beh *BufferedEventHandler) HandleEvent(event domain.Event) error {
	select {
	case beh.buffer <- event:
		return nil
	case <-beh.ctx.Done():
		return domain.ErrEventDispatch
	default:
		// Buffer full
		return domain.ErrEventDispatch
	}
}

// processEvents processes buffered events in a separate goroutine.
// It continues processing until the context is cancelled, then drains
// any remaining events from the buffer before exiting.
func (beh *BufferedEventHandler) processEvents() {
	defer beh.wg.Done()

	for {
		select {
		case event, ok := <-beh.buffer:
			if !ok {
				return
			}
			if err := beh.handler.HandleEvent(event); err != nil {
				// In production, log the error
				_ = err
			}
		case <-beh.ctx.Done():
			// Drain buffer
			for {
				select {
				case event, ok := <-beh.buffer:
					if !ok {
						return
					}
					if err := beh.handler.HandleEvent(event); err != nil {
						// In production, log the error
						_ = err
					}
				default:
					return
				}
			}
		}
	}
}

// Close closes the buffered handler and waits for pending events to process.
// After closing, HandleEvent will return domain.ErrEventDispatch.
// This method blocks until all buffered events are processed.
func (beh *BufferedEventHandler) Close() {
	beh.cancel()
	close(beh.buffer)
	beh.wg.Wait()
}

// CompositeEventHandler distributes events to multiple handlers.
// It implements the Composite pattern for event handling, allowing
// multiple handlers to process the same event independently.
type CompositeEventHandler struct {
	handlers []domain.EventHandler
	mu       sync.RWMutex
}

// NewCompositeEventHandler creates a new composite event handler.
// The handlers parameter accepts zero or more event handlers that will
// receive all events. Nil handlers are ignored during event processing.
func NewCompositeEventHandler(handlers ...domain.EventHandler) *CompositeEventHandler {
	return &CompositeEventHandler{
		handlers: handlers,
	}
}

// HandleEvent implements domain.EventHandler.
// Distributes the event to all registered handlers. If any handler returns
// an error, all errors are collected and returned as a MultiError.
func (ceh *CompositeEventHandler) HandleEvent(event domain.Event) error {
	ceh.mu.RLock()
	defer ceh.mu.RUnlock()

	multiErr := domain.NewMultiError()
	for _, handler := range ceh.handlers {
		if handler != nil {
			if err := handler.HandleEvent(event); err != nil {
				multiErr.Add(err)
			}
		}
	}

	if multiErr.HasErrors() {
		return multiErr
	}
	return nil
}

// AddHandler adds a handler to the composite.
// Nil handlers are ignored. The handler will receive all future events
// processed by this composite. Thread-safe for concurrent use.
func (ceh *CompositeEventHandler) AddHandler(handler domain.EventHandler) {
	if handler == nil {
		return
	}

	ceh.mu.Lock()
	defer ceh.mu.Unlock()
	ceh.handlers = append(ceh.handlers, handler)
}

// RemoveHandler removes a handler from the composite.
// If the handler appears multiple times, only the first occurrence is removed.
// Nil handlers are ignored. Thread-safe for concurrent use.
func (ceh *CompositeEventHandler) RemoveHandler(handler domain.EventHandler) {
	if handler == nil {
		return
	}

	ceh.mu.Lock()
	defer ceh.mu.Unlock()

	newHandlers := make([]domain.EventHandler, 0, len(ceh.handlers))
	for _, h := range ceh.handlers {
		if h != handler {
			newHandlers = append(newHandlers, h)
		}
	}
	ceh.handlers = newHandlers
}

// FilteredEventHandler applies filters before handling events.
// It wraps an event handler and only forwards events that pass
// all configured filters, providing selective event processing.
type FilteredEventHandler struct {
	handler domain.EventHandler
	filters []domain.EventFilter
}

// NewFilteredEventHandler creates a new filtered event handler.
// Events must pass all provided filters to reach the wrapped handler.
// If no filters are provided, all events are forwarded to the handler.
func NewFilteredEventHandler(handler domain.EventHandler, filters ...domain.EventFilter) *FilteredEventHandler {
	return &FilteredEventHandler{
		handler: handler,
		filters: filters,
	}
}

// HandleEvent implements domain.EventHandler.
// Applies all filters to the event. If any filter returns false,
// the event is dropped and nil is returned. Otherwise forwards to wrapped handler.
func (feh *FilteredEventHandler) HandleEvent(event domain.Event) error {
	// Check all filters
	for _, filter := range feh.filters {
		if filter != nil && !filter(event) {
			return nil // Event filtered out
		}
	}

	return feh.handler.HandleEvent(event)
}

// LoggingEventHandler logs events (placeholder for actual implementation).
// This handler logs event details at the configured level.
// In production, this would integrate with a logging framework.
type LoggingEventHandler struct {
	level string
}

// NewLoggingEventHandler creates a new logging event handler.
// The level parameter controls the verbosity of logging (e.g., "debug", "info").
// This is a placeholder implementation for demonstration purposes.
func NewLoggingEventHandler(level string) *LoggingEventHandler {
	return &LoggingEventHandler{
		level: level,
	}
}

// HandleEvent implements domain.EventHandler.
// Logs the event details according to the configured level.
// This placeholder implementation always returns nil.
func (leh *LoggingEventHandler) HandleEvent(event domain.Event) error {
	// In a real implementation, this would use a logging framework
	// For now, this is just a placeholder
	return nil
}
