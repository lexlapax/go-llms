// ABOUTME: Event storage and replay system for persistence and debugging
// ABOUTME: Provides event recording, storage interfaces, and replay capabilities

package events

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// EventStorage defines the interface for event persistence.
// Implementations can provide various backends like memory, file, or database storage.
type EventStorage interface {
	// Store saves an event
	Store(ctx context.Context, event domain.Event) error

	// StoreBatch saves multiple events
	StoreBatch(ctx context.Context, events []domain.Event) error

	// Query retrieves events based on criteria
	Query(ctx context.Context, query EventQuery) ([]domain.Event, error)

	// Stream returns a channel of events matching the query
	Stream(ctx context.Context, query EventQuery) (<-chan domain.Event, error)

	// Count returns the number of events matching the query
	Count(ctx context.Context, query EventQuery) (int64, error)

	// Close closes the storage
	Close() error
}

// EventQuery defines criteria for querying events from storage.
// It supports filtering by time range, agent information, event types,
// and provides pagination and ordering options.
type EventQuery struct {
	// Time range
	StartTime *time.Time
	EndTime   *time.Time

	// Filters
	AgentID    string
	AgentName  string
	EventTypes []domain.EventType

	// Pagination
	Offset int
	Limit  int

	// Ordering
	OrderBy    string // "timestamp", "type", "agent_id"
	Descending bool
}

// MemoryStorage implements in-memory event storage.
// It provides fast access but data is lost on process termination.
// Suitable for testing and temporary event storage.
type MemoryStorage struct {
	mu     sync.RWMutex
	events []domain.Event
	closed bool
}

// NewMemoryStorage creates a new in-memory event storage.
// The storage starts empty and grows as events are added.
//
// Returns a new MemoryStorage instance.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		events: make([]domain.Event, 0),
	}
}

// Store implements EventStorage interface.
// It appends the event to the in-memory slice in a thread-safe manner.
func (s *MemoryStorage) Store(ctx context.Context, event domain.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return fmt.Errorf("storage is closed")
	}

	s.events = append(s.events, event)
	return nil
}

// StoreBatch implements EventStorage interface.
// It efficiently stores multiple events in a single operation.
func (s *MemoryStorage) StoreBatch(ctx context.Context, events []domain.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return fmt.Errorf("storage is closed")
	}

	s.events = append(s.events, events...)
	return nil
}

// Query implements EventStorage interface.
// It filters events based on the query criteria and applies pagination.
func (s *MemoryStorage) Query(ctx context.Context, query EventQuery) ([]domain.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, fmt.Errorf("storage is closed")
	}

	// Filter events
	filtered := make([]domain.Event, 0)
	for _, event := range s.events {
		if s.matchesQuery(event, query) {
			filtered = append(filtered, event)
		}
	}

	// Apply pagination
	start := query.Offset
	if start > len(filtered) {
		return []domain.Event{}, nil
	}

	end := start + query.Limit
	if end > len(filtered) || query.Limit == 0 {
		end = len(filtered)
	}

	return filtered[start:end], nil
}

// Stream implements EventStorage interface.
// It returns a channel that emits events matching the query criteria.
// The channel is closed when all matching events have been sent.
func (s *MemoryStorage) Stream(ctx context.Context, query EventQuery) (<-chan domain.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, fmt.Errorf("storage is closed")
	}

	ch := make(chan domain.Event)

	go func() {
		defer close(ch)

		s.mu.RLock()
		events := make([]domain.Event, len(s.events))
		copy(events, s.events)
		s.mu.RUnlock()

		for _, event := range events {
			if s.matchesQuery(event, query) {
				select {
				case ch <- event:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return ch, nil
}

// Count implements EventStorage interface.
// It returns the number of events matching the query criteria.
func (s *MemoryStorage) Count(ctx context.Context, query EventQuery) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return 0, fmt.Errorf("storage is closed")
	}

	var count int64
	for _, event := range s.events {
		if s.matchesQuery(event, query) {
			count++
		}
	}

	return count, nil
}

// Close implements EventStorage interface.
// It marks the storage as closed and clears all stored events.
func (s *MemoryStorage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.closed = true
	s.events = nil
	return nil
}

// matchesQuery checks if an event matches the query criteria.
// It evaluates time range, agent filters, and event type filters.
func (s *MemoryStorage) matchesQuery(event domain.Event, query EventQuery) bool {
	// Check time range
	if query.StartTime != nil && event.Timestamp.Before(*query.StartTime) {
		return false
	}
	if query.EndTime != nil && event.Timestamp.After(*query.EndTime) {
		return false
	}

	// Check agent filters
	if query.AgentID != "" && event.AgentID != query.AgentID {
		return false
	}
	if query.AgentName != "" && event.AgentName != query.AgentName {
		return false
	}

	// Check event types
	if len(query.EventTypes) > 0 {
		found := false
		for _, t := range query.EventTypes {
			if event.Type == t {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// EventRecorder records events from an event bus to storage.
// It subscribes to events and persists them for later retrieval or analysis.
type EventRecorder struct {
	storage EventStorage
	bus     *EventBus
	subID   string
	wg      sync.WaitGroup
	cancel  context.CancelFunc
}

// NewEventRecorder creates a new event recorder.
//
// Parameters:
//   - storage: The storage backend to record events to
//   - bus: The event bus to record events from
//
// Returns a new EventRecorder instance.
func NewEventRecorder(storage EventStorage, bus *EventBus) *EventRecorder {
	return &EventRecorder{
		storage: storage,
		bus:     bus,
	}
}

// Start begins recording events that match the specified filters.
// Only one recording session can be active at a time.
//
// Parameters:
//   - filters: Optional filters to apply to recorded events
//
// Returns an error if the recorder is already started.
func (r *EventRecorder) Start(filters ...EventFilter) error {
	if r.cancel != nil {
		return fmt.Errorf("recorder already started")
	}

	_, cancel := context.WithCancel(context.Background())
	r.cancel = cancel

	// Subscribe to events
	handler := EventHandlerFunc(func(ctx context.Context, event domain.Event) error {
		return r.storage.Store(ctx, event)
	})

	r.subID = r.bus.Subscribe(handler, filters...)

	return nil
}

// Stop stops recording events and unsubscribes from the event bus.
// It waits for any pending events to be stored before returning.
func (r *EventRecorder) Stop() {
	if r.cancel != nil {
		r.cancel()
		r.cancel = nil
	}

	if r.subID != "" {
		r.bus.Unsubscribe(r.subID)
		r.subID = ""
	}

	r.wg.Wait()
}

// EventReplayer replays stored events to an event bus.
// It can replay events at different speeds and apply transformations.
type EventReplayer struct {
	storage EventStorage
	bus     *EventBus
}

// NewEventReplayer creates a new event replayer.
//
// Parameters:
//   - storage: The storage to read events from
//   - bus: The event bus to replay events to
//
// Returns a new EventReplayer instance.
func NewEventReplayer(storage EventStorage, bus *EventBus) *EventReplayer {
	return &EventReplayer{
		storage: storage,
		bus:     bus,
	}
}

// ReplayOptions configures event replay behavior.
// It controls replay speed, filtering, and event transformation.
type ReplayOptions struct {
	// Speed multiplier (1.0 = real-time, 2.0 = 2x speed, 0 = instant)
	Speed float64

	// Filter to apply during replay
	Filter EventFilter

	// Transformer to modify events during replay
	Transformer EventTransformer
}

// EventTransformer modifies events during replay.
// It can be used to update timestamps, agent IDs, or other event properties.
type EventTransformer func(event domain.Event) domain.Event

// Replay replays events from storage according to the specified options.
// Events are published to the bus with timing that simulates the original event flow.
//
// Parameters:
//   - ctx: Context for cancellation
//   - query: Query to select events to replay
//   - opts: Options controlling replay behavior
//
// Returns an error if event retrieval fails or context is cancelled.
func (r *EventReplayer) Replay(ctx context.Context, query EventQuery, opts ReplayOptions) error {
	// Get events from storage
	events, err := r.storage.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query events: %w", err)
	}

	if len(events) == 0 {
		return nil
	}

	// Calculate timing
	var lastTimestamp time.Time
	if opts.Speed > 0 {
		lastTimestamp = events[0].Timestamp
	}

	// Replay events
	for _, event := range events {
		// Apply filter if specified
		if opts.Filter != nil && !opts.Filter.Match(event) {
			continue
		}

		// Apply transformer if specified
		if opts.Transformer != nil {
			event = opts.Transformer(event)
		}

		// Calculate delay for realistic replay
		if opts.Speed > 0 && !lastTimestamp.IsZero() {
			delay := event.Timestamp.Sub(lastTimestamp)
			if opts.Speed != 1.0 {
				delay = time.Duration(float64(delay) / opts.Speed)
			}

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// Publish event
		r.bus.PublishContext(ctx, event)

		lastTimestamp = event.Timestamp
	}

	return nil
}

// FileStorage implements file-based event storage.
// It writes events as serialized lines to a file, suitable for append-only logs.
// This implementation only supports writing; querying requires a separate reader.
type FileStorage struct {
	writer     io.WriteCloser
	serializer EventSerializer
	mu         sync.Mutex
}

// NewFileStorage creates a new file-based event storage.
//
// Parameters:
//   - w: Writer for event data (typically a file)
//   - serializer: Serializer for converting events to bytes
//
// Returns a new FileStorage instance.
func NewFileStorage(w io.WriteCloser, serializer EventSerializer) *FileStorage {
	return &FileStorage{
		writer:     w,
		serializer: serializer,
	}
}

// Store implements EventStorage interface.
// It serializes the event and writes it as a line to the file.
func (s *FileStorage) Store(ctx context.Context, event domain.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.serializer.Serialize(event)
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	// Write as JSON lines
	if _, err := s.writer.Write(data); err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}

	if _, err := s.writer.Write([]byte("\n")); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}

	return nil
}

// StoreBatch implements EventStorage interface.
// It stores multiple events by calling Store for each event.
func (s *FileStorage) StoreBatch(ctx context.Context, events []domain.Event) error {
	for _, event := range events {
		if err := s.Store(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

// Query implements EventStorage interface.
// This operation is not supported for append-only file storage.
func (s *FileStorage) Query(ctx context.Context, query EventQuery) ([]domain.Event, error) {
	return nil, fmt.Errorf("query not supported for file storage")
}

// Stream implements EventStorage interface.
// This operation is not supported for append-only file storage.
func (s *FileStorage) Stream(ctx context.Context, query EventQuery) (<-chan domain.Event, error) {
	return nil, fmt.Errorf("stream not supported for file storage")
}

// Count implements EventStorage interface.
// This operation is not supported for append-only file storage.
func (s *FileStorage) Count(ctx context.Context, query EventQuery) (int64, error) {
	return 0, fmt.Errorf("count not supported for file storage")
}

// Close implements EventStorage interface.
// It closes the underlying writer, flushing any buffered data.
func (s *FileStorage) Close() error {
	return s.writer.Close()
}
