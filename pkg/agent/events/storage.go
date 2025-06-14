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

// EventStorage defines the interface for event persistence
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

// EventQuery defines criteria for querying events
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

// MemoryStorage implements in-memory event storage
type MemoryStorage struct {
	mu     sync.RWMutex
	events []domain.Event
	closed bool
}

// NewMemoryStorage creates a new in-memory event storage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		events: make([]domain.Event, 0),
	}
}

// Store implements EventStorage
func (s *MemoryStorage) Store(ctx context.Context, event domain.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return fmt.Errorf("storage is closed")
	}

	s.events = append(s.events, event)
	return nil
}

// StoreBatch implements EventStorage
func (s *MemoryStorage) StoreBatch(ctx context.Context, events []domain.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return fmt.Errorf("storage is closed")
	}

	s.events = append(s.events, events...)
	return nil
}

// Query implements EventStorage
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

// Stream implements EventStorage
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

// Count implements EventStorage
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

// Close implements EventStorage
func (s *MemoryStorage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.closed = true
	s.events = nil
	return nil
}

// matchesQuery checks if an event matches the query criteria
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

// EventRecorder records events to storage
type EventRecorder struct {
	storage EventStorage
	bus     *EventBus
	subID   string
	wg      sync.WaitGroup
	cancel  context.CancelFunc
}

// NewEventRecorder creates a new event recorder
func NewEventRecorder(storage EventStorage, bus *EventBus) *EventRecorder {
	return &EventRecorder{
		storage: storage,
		bus:     bus,
	}
}

// Start begins recording events
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

// Stop stops recording events
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

// EventReplayer replays stored events
type EventReplayer struct {
	storage EventStorage
	bus     *EventBus
}

// NewEventReplayer creates a new event replayer
func NewEventReplayer(storage EventStorage, bus *EventBus) *EventReplayer {
	return &EventReplayer{
		storage: storage,
		bus:     bus,
	}
}

// ReplayOptions configures event replay
type ReplayOptions struct {
	// Speed multiplier (1.0 = real-time, 2.0 = 2x speed, 0 = instant)
	Speed float64

	// Filter to apply during replay
	Filter EventFilter

	// Transformer to modify events during replay
	Transformer EventTransformer
}

// EventTransformer modifies events during replay
type EventTransformer func(event domain.Event) domain.Event

// Replay replays events from storage
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

// FileStorage implements file-based event storage
type FileStorage struct {
	writer     io.WriteCloser
	serializer EventSerializer
	mu         sync.Mutex
}

// NewFileStorage creates a new file-based event storage
func NewFileStorage(w io.WriteCloser, serializer EventSerializer) *FileStorage {
	return &FileStorage{
		writer:     w,
		serializer: serializer,
	}
}

// Store implements EventStorage
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

// StoreBatch implements EventStorage
func (s *FileStorage) StoreBatch(ctx context.Context, events []domain.Event) error {
	for _, event := range events {
		if err := s.Store(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

// Query implements EventStorage (not supported for file storage)
func (s *FileStorage) Query(ctx context.Context, query EventQuery) ([]domain.Event, error) {
	return nil, fmt.Errorf("query not supported for file storage")
}

// Stream implements EventStorage (not supported for file storage)
func (s *FileStorage) Stream(ctx context.Context, query EventQuery) (<-chan domain.Event, error) {
	return nil, fmt.Errorf("stream not supported for file storage")
}

// Count implements EventStorage (not supported for file storage)
func (s *FileStorage) Count(ctx context.Context, query EventQuery) (int64, error) {
	return 0, fmt.Errorf("count not supported for file storage")
}

// Close implements EventStorage
func (s *FileStorage) Close() error {
	return s.writer.Close()
}
