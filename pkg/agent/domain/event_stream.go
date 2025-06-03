// ABOUTME: Defines functional stream operations for processing agent events
// ABOUTME: Provides filter, map, reduce, and other stream operations on event sequences

package domain

import (
	"context"
	"sync"
	"time"
)

// FunctionalEventStream provides functional operations on event streams
type FunctionalEventStream interface {
	// Core operations
	Filter(predicate EventPredicate) FunctionalEventStream
	Map(transform EventTransform) FunctionalEventStream
	Reduce(reducer EventReducer, initial interface{}) interface{}

	// Stream control
	Take(n int) FunctionalEventStream
	TakeUntil(predicate EventPredicate) FunctionalEventStream
	Timeout(duration time.Duration) FunctionalEventStream

	// Consumption
	ForEach(handler EventHandler) error
	Collect() ([]Event, error)
	First() (Event, error)
}

// EventPredicate filters events
type EventPredicate func(Event) bool

// EventTransform transforms events
type EventTransform func(Event) Event

// EventReducer reduces events to a single value
type EventReducer func(interface{}, Event) interface{}

// Common predicates
var (
	// IsError checks if event is an error
	IsError EventPredicate = func(e Event) bool {
		return e.Type == EventAgentError || e.Type == EventToolError
	}

	// IsComplete checks if event is a completion
	IsComplete EventPredicate = func(e Event) bool {
		return e.Type == EventAgentComplete
	}

	// ByType creates a predicate for event type
	ByType = func(eventType EventType) EventPredicate {
		return func(e Event) bool {
			return e.Type == eventType
		}
	}

	// ByAgent creates a predicate for agent name
	ByAgent = func(agentName string) EventPredicate {
		return func(e Event) bool {
			return e.AgentName == agentName
		}
	}

	// ByAgentID creates a predicate for agent ID
	ByAgentID = func(agentID string) EventPredicate {
		return func(e Event) bool {
			return e.AgentID == agentID
		}
	}

	// HasError checks if event contains an error
	HasError EventPredicate = func(e Event) bool {
		return e.Error != nil
	}

	// And combines predicates with AND logic
	And = func(predicates ...EventPredicate) EventPredicate {
		return func(e Event) bool {
			for _, p := range predicates {
				if !p(e) {
					return false
				}
			}
			return true
		}
	}

	// Or combines predicates with OR logic
	Or = func(predicates ...EventPredicate) EventPredicate {
		return func(e Event) bool {
			for _, p := range predicates {
				if p(e) {
					return true
				}
			}
			return false
		}
	}

	// Not negates a predicate
	Not = func(predicate EventPredicate) EventPredicate {
		return func(e Event) bool {
			return !predicate(e)
		}
	}
)

// Common transforms
var (
	// WithTimestamp adds current timestamp
	WithTimestamp EventTransform = func(e Event) Event {
		e.Timestamp = time.Now()
		return e
	}

	// WithMetadata adds metadata
	WithMetadata = func(key string, value interface{}) EventTransform {
		return func(e Event) Event {
			if e.Metadata == nil {
				e.Metadata = make(map[string]interface{})
			}
			e.Metadata[key] = value
			return e
		}
	}

	// StripMetadata removes all metadata
	StripMetadata EventTransform = func(e Event) Event {
		e.Metadata = nil
		return e
	}
)

// eventStream is the default implementation
type eventStream struct {
	source  <-chan Event
	ctx     context.Context
	cancel  context.CancelFunc
	timeout time.Duration
	once    sync.Once
}

// Close cancels the stream's context (only called once)
func (s *eventStream) Close() {
	s.once.Do(func() {
		if s.cancel != nil {
			s.cancel()
		}
	})
}

// NewFunctionalEventStream creates a new functional event stream from a channel
func NewFunctionalEventStream(ctx context.Context, source <-chan Event) FunctionalEventStream {
	streamCtx, cancel := context.WithCancel(ctx)
	return &eventStream{
		source: source,
		ctx:    streamCtx,
		cancel: cancel,
	}
}

// Filter returns a new stream with filtered events
func (s *eventStream) Filter(predicate EventPredicate) FunctionalEventStream {
	filtered := make(chan Event)
	newCtx, cancel := context.WithCancel(s.ctx)

	go func() {
		defer close(filtered)

		for {
			select {
			case event, ok := <-s.source:
				if !ok {
					return
				}
				if predicate(event) {
					select {
					case filtered <- event:
					case <-s.ctx.Done():
						return
					}
				}
			case <-s.ctx.Done():
				return
			}
		}
	}()

	return &eventStream{
		source: filtered,
		ctx:    newCtx,
		cancel: cancel,
	}
}

// Map returns a new stream with transformed events
func (s *eventStream) Map(transform EventTransform) FunctionalEventStream {
	mapped := make(chan Event)
	newCtx, cancel := context.WithCancel(s.ctx)

	go func() {
		defer close(mapped)

		for {
			select {
			case event, ok := <-s.source:
				if !ok {
					return
				}
				transformed := transform(event)
				select {
				case mapped <- transformed:
				case <-s.ctx.Done():
					return
				}
			case <-s.ctx.Done():
				return
			}
		}
	}()

	return &eventStream{
		source: mapped,
		ctx:    newCtx,
		cancel: cancel,
	}
}

// Reduce accumulates events into a single value
func (s *eventStream) Reduce(reducer EventReducer, initial interface{}) interface{} {
	result := initial

	for {
		select {
		case event, ok := <-s.source:
			if !ok {
				return result
			}
			result = reducer(result, event)
		case <-s.ctx.Done():
			return result
		}
	}
}

// Take returns a stream that emits at most n events
func (s *eventStream) Take(n int) FunctionalEventStream {
	taken := make(chan Event)
	newCtx, cancel := context.WithCancel(s.ctx)

	go func() {
		defer close(taken)

		count := 0
		for {
			if count >= n {
				return
			}

			select {
			case event, ok := <-s.source:
				if !ok {
					return
				}
				select {
				case taken <- event:
					count++
				case <-s.ctx.Done():
					return
				}
			case <-s.ctx.Done():
				return
			}
		}
	}()

	return &eventStream{
		source: taken,
		ctx:    newCtx,
		cancel: cancel,
	}
}

// TakeUntil returns a stream that emits until predicate is true
func (s *eventStream) TakeUntil(predicate EventPredicate) FunctionalEventStream {
	taken := make(chan Event)
	newCtx, cancel := context.WithCancel(s.ctx)

	go func() {
		defer close(taken)

		for {
			select {
			case event, ok := <-s.source:
				if !ok {
					return
				}

				// Check predicate before emitting
				if predicate(event) {
					// Emit the final event that matched
					select {
					case taken <- event:
					case <-s.ctx.Done():
					}
					return
				}

				select {
				case taken <- event:
				case <-s.ctx.Done():
					return
				}
			case <-s.ctx.Done():
				return
			}
		}
	}()

	return &eventStream{
		source: taken,
		ctx:    newCtx,
		cancel: cancel,
	}
}

// Timeout returns a stream that times out after duration
func (s *eventStream) Timeout(duration time.Duration) FunctionalEventStream {
	newCtx, cancel := context.WithTimeout(s.ctx, duration)
	return &eventStream{
		source:  s.source,
		ctx:     newCtx,
		cancel:  cancel,
		timeout: duration,
	}
}

// ForEach applies handler to each event
func (s *eventStream) ForEach(handler EventHandler) error {
	for {
		select {
		case event, ok := <-s.source:
			if !ok {
				return nil
			}
			if err := handler.HandleEvent(event); err != nil {
				return err
			}
		case <-s.ctx.Done():
			if s.timeout > 0 {
				return ErrExecutionTimeout
			}
			return s.ctx.Err()
		}
	}
}

// Collect gathers all events into a slice
func (s *eventStream) Collect() ([]Event, error) {
	var events []Event

	for {
		select {
		case event, ok := <-s.source:
			if !ok {
				return events, nil
			}
			events = append(events, event)
		case <-s.ctx.Done():
			if s.timeout > 0 {
				return events, ErrExecutionTimeout
			}
			return events, s.ctx.Err()
		}
	}
}

// First returns the first event
func (s *eventStream) First() (Event, error) {
	select {
	case event, ok := <-s.source:
		if !ok {
			return Event{}, ErrEventDispatch
		}
		return event, nil
	case <-s.ctx.Done():
		if s.timeout > 0 {
			return Event{}, ErrExecutionTimeout
		}
		return Event{}, s.ctx.Err()
	}
}

// Helper functions for creating event streams

// EventsFromSlice creates a stream from a slice of events
func EventsFromSlice(ctx context.Context, events []Event) FunctionalEventStream {
	ch := make(chan Event, len(events))
	for _, e := range events {
		ch <- e
	}
	close(ch)
	return NewFunctionalEventStream(ctx, ch)
}

// MergeFunctionalEventStreams merges multiple functional streams into one
func MergeFunctionalEventStreams(ctx context.Context, streams ...FunctionalEventStream) FunctionalEventStream {
	merged := make(chan Event)
	var wg sync.WaitGroup

	// Collect from each stream
	for _, stream := range streams {
		wg.Add(1)
		go func(s FunctionalEventStream) {
			defer wg.Done()
			_ = s.ForEach(EventHandlerFunc(func(e Event) error {
				select {
				case merged <- e:
				case <-ctx.Done():
					return ctx.Err()
				}
				return nil
			}))
		}(stream)
	}

	// Close merged channel when all streams are done
	go func() {
		wg.Wait()
		close(merged)
	}()

	return NewFunctionalEventStream(ctx, merged)
}
