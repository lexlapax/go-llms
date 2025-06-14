package helpers

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

func TestEventCapture(t *testing.T) {
	t.Run("basic capture", func(t *testing.T) {
		ec := NewEventCapture()

		event1 := domain.NewEvent(domain.EventAgentStart, "agent1", "Agent 1", nil)
		event2 := domain.NewEvent(domain.EventToolCall, "agent1", "Agent 1", map[string]interface{}{"tool": "calc"})

		ec.Capture(event1)
		ec.Capture(event2)

		events := ec.GetEvents()
		assert.Len(t, events, 2)
		assert.Equal(t, domain.EventAgentStart, events[0].Type)
		assert.Equal(t, domain.EventToolCall, events[1].Type)
	})

	t.Run("filter by type", func(t *testing.T) {
		ec := NewEventCapture()

		ec.Capture(domain.NewEvent(domain.EventAgentStart, "agent1", "Agent 1", nil))
		ec.Capture(domain.NewEvent(domain.EventToolCall, "agent1", "Agent 1", nil))
		ec.Capture(domain.NewEvent(domain.EventToolCall, "agent1", "Agent 1", nil))
		ec.Capture(domain.NewEvent(domain.EventAgentComplete, "agent1", "Agent 1", nil))

		toolCalls := ec.FilterByType(domain.EventToolCall)
		assert.Len(t, toolCalls, 2)

		starts := ec.FilterByType(domain.EventAgentStart)
		assert.Len(t, starts, 1)
	})

	t.Run("filter by data", func(t *testing.T) {
		ec := NewEventCapture()

		ec.Capture(domain.NewEvent(domain.EventToolCall, "agent1", "Agent 1", map[string]interface{}{"tool": "calc"}))
		ec.Capture(domain.NewEvent(domain.EventToolCall, "agent1", "Agent 1", map[string]interface{}{"tool": "search"}))
		ec.Capture(domain.NewEvent(domain.EventMessage, "agent1", "Agent 1", map[string]interface{}{"message": "hello"}))

		calcCalls := ec.FilterByData(func(data interface{}) bool {
			if m, ok := data.(map[string]interface{}); ok {
				return m["tool"] == "calc"
			}
			return false
		})

		assert.Len(t, calcCalls, 1)
		assert.Equal(t, domain.EventToolCall, calcCalls[0].Type)
	})

	t.Run("filter by time range", func(t *testing.T) {
		ec := NewEventCapture()

		start := time.Now()

		event1 := domain.NewEvent(domain.EventAgentStart, "agent1", "Agent 1", nil)
		ec.Capture(event1)

		time.Sleep(10 * time.Millisecond)
		mid := time.Now()

		event2 := domain.NewEvent(domain.EventToolCall, "agent1", "Agent 1", nil)
		ec.Capture(event2)

		time.Sleep(10 * time.Millisecond)

		event3 := domain.NewEvent(domain.EventAgentComplete, "agent1", "Agent 1", nil)
		ec.Capture(event3)

		end := time.Now()

		// Should get all events
		allEvents := ec.FilterByTimeRange(start, end)
		assert.Len(t, allEvents, 3)

		// Should only get middle event
		midEvents := ec.FilterByTimeRange(mid.Add(-5*time.Millisecond), mid.Add(5*time.Millisecond))
		assert.Len(t, midEvents, 1)
		assert.Equal(t, domain.EventToolCall, midEvents[0].Type)
	})

	t.Run("clear", func(t *testing.T) {
		ec := NewEventCapture()

		ec.Capture(domain.NewEvent(domain.EventAgentStart, "agent1", "Agent 1", nil))
		ec.Capture(domain.NewEvent(domain.EventToolCall, "agent1", "Agent 1", nil))

		assert.Len(t, ec.GetEvents(), 2)

		ec.Clear()
		assert.Len(t, ec.GetEvents(), 0)
	})
}

func TestEventAssertion(t *testing.T) {
	t.Run("has count", func(t *testing.T) {
		events := []domain.Event{
			domain.NewEvent(domain.EventAgentStart, "agent1", "Agent 1", nil),
			domain.NewEvent(domain.EventToolCall, "agent1", "Agent 1", nil),
		}

		ea := AssertEvents(events)
		ea.HasCount(2)
		assert.True(t, ea.IsValid())
		assert.Empty(t, ea.GetErrors())

		ea2 := AssertEvents(events)
		ea2.HasCount(3)
		assert.False(t, ea2.IsValid())
		assert.Len(t, ea2.GetErrors(), 1)
	})

	t.Run("has type", func(t *testing.T) {
		events := []domain.Event{
			domain.NewEvent(domain.EventAgentStart, "agent1", "Agent 1", nil),
			domain.NewEvent(domain.EventToolCall, "agent1", "Agent 1", nil),
		}

		ea := AssertEvents(events)
		ea.HasType(domain.EventAgentStart)
		assert.True(t, ea.IsValid())

		ea2 := AssertEvents(events)
		ea2.HasType(domain.EventAgentError)
		assert.False(t, ea2.IsValid())
	})

	t.Run("has type count", func(t *testing.T) {
		events := []domain.Event{
			domain.NewEvent(domain.EventToolCall, "agent1", "Agent 1", nil),
			domain.NewEvent(domain.EventToolCall, "agent1", "Agent 1", nil),
			domain.NewEvent(domain.EventAgentComplete, "agent1", "Agent 1", nil),
		}

		ea := AssertEvents(events)
		ea.HasTypeCount(domain.EventToolCall, 2)
		assert.True(t, ea.IsValid())

		ea2 := AssertEvents(events)
		ea2.HasTypeCount(domain.EventToolCall, 1)
		assert.False(t, ea2.IsValid())
	})

	t.Run("in order", func(t *testing.T) {
		events := []domain.Event{
			domain.NewEvent(domain.EventAgentStart, "agent1", "Agent 1", nil),
			domain.NewEvent(domain.EventToolCall, "agent1", "Agent 1", nil),
			domain.NewEvent(domain.EventToolResult, "agent1", "Agent 1", nil),
			domain.NewEvent(domain.EventAgentComplete, "agent1", "Agent 1", nil),
		}

		ea := AssertEvents(events)
		ea.InOrder(domain.EventAgentStart, domain.EventToolCall, domain.EventAgentComplete)
		assert.True(t, ea.IsValid())

		ea2 := AssertEvents(events)
		ea2.InOrder(domain.EventAgentComplete, domain.EventAgentStart)
		assert.False(t, ea2.IsValid())
	})

	t.Run("no errors", func(t *testing.T) {
		events := []domain.Event{
			domain.NewEvent(domain.EventAgentStart, "agent1", "Agent 1", nil),
			domain.NewEvent(domain.EventToolCall, "agent1", "Agent 1", nil),
		}

		ea := AssertEvents(events)
		ea.NoErrors()
		assert.True(t, ea.IsValid())

		// Test with error event
		errorEvents := []domain.Event{
			domain.NewEvent(domain.EventAgentStart, "agent1", "Agent 1", nil),
			domain.NewEvent(domain.EventAgentError, "agent1", "Agent 1", nil),
		}

		ea2 := AssertEvents(errorEvents)
		ea2.NoErrors()
		assert.False(t, ea2.IsValid())

		// Test with event containing error
		eventWithError := domain.NewEvent(domain.EventMessage, "agent1", "Agent 1", nil).WithError(errors.New("test error"))
		eventsWithError := []domain.Event{eventWithError}

		ea3 := AssertEvents(eventsWithError)
		ea3.NoErrors()
		assert.False(t, ea3.IsValid())
	})

	t.Run("with data", func(t *testing.T) {
		events := []domain.Event{
			domain.NewEvent(domain.EventToolCall, "agent1", "Agent 1", map[string]interface{}{
				"tool":  "calculator",
				"input": "2+2",
			}),
			domain.NewEvent(domain.EventMessage, "agent1", "Agent 1", map[string]interface{}{
				"message": "hello",
			}),
		}

		ea := AssertEvents(events)
		ea.WithData("tool", "calculator")
		assert.True(t, ea.IsValid())

		ea2 := AssertEvents(events)
		ea2.WithData("tool", "search")
		assert.False(t, ea2.IsValid())
	})

	t.Run("chained assertions", func(t *testing.T) {
		events := []domain.Event{
			domain.NewEvent(domain.EventAgentStart, "agent1", "Agent 1", nil),
			domain.NewEvent(domain.EventToolCall, "agent1", "Agent 1", map[string]interface{}{"tool": "calc"}),
			domain.NewEvent(domain.EventAgentComplete, "agent1", "Agent 1", nil),
		}

		ea := AssertEvents(events).
			HasCount(3).
			HasType(domain.EventAgentStart).
			HasTypeCount(domain.EventToolCall, 1).
			WithData("tool", "calc").
			NoErrors()

		assert.True(t, ea.IsValid())
		assert.Empty(t, ea.GetErrors())
	})
}

func TestEventTimeline(t *testing.T) {
	t.Run("empty timeline", func(t *testing.T) {
		et := NewEventTimeline([]domain.Event{})
		assert.Equal(t, "No events to visualize", et.Visualize())
		assert.Equal(t, time.Duration(0), et.GetDuration())
	})

	t.Run("timeline with events", func(t *testing.T) {
		events := []domain.Event{
			domain.NewEvent(domain.EventAgentStart, "agent1", "Agent 1", nil),
			domain.NewEvent(domain.EventToolCall, "agent1", "Agent 1", map[string]interface{}{"tool": "calc"}),
			domain.NewEvent(domain.EventAgentComplete, "agent1", "Agent 1", nil),
		}

		// Set timestamps manually for predictable output
		now := time.Now()
		events[0].Timestamp = now
		events[1].Timestamp = now.Add(100 * time.Millisecond)
		events[2].Timestamp = now.Add(200 * time.Millisecond)

		et := NewEventTimeline(events)

		visualization := et.Visualize()
		assert.Contains(t, visualization, "Event Timeline:")
		assert.Contains(t, visualization, string(domain.EventAgentStart))
		assert.Contains(t, visualization, string(domain.EventToolCall))
		assert.Contains(t, visualization, string(domain.EventAgentComplete))
		assert.Contains(t, visualization, "Total duration:")
		assert.Contains(t, visualization, "Total events: 3")

		assert.Equal(t, 200*time.Millisecond, et.GetDuration())
	})

	t.Run("group by type", func(t *testing.T) {
		events := []domain.Event{
			domain.NewEvent(domain.EventToolCall, "agent1", "Agent 1", nil),
			domain.NewEvent(domain.EventToolCall, "agent2", "Agent 2", nil),
			domain.NewEvent(domain.EventAgentStart, "agent1", "Agent 1", nil),
			domain.NewEvent(domain.EventToolCall, "agent1", "Agent 1", nil),
		}

		et := NewEventTimeline(events)
		groups := et.GroupByType()

		assert.Len(t, groups, 2)
		assert.Len(t, groups[domain.EventToolCall], 3)
		assert.Len(t, groups[domain.EventAgentStart], 1)
	})

	t.Run("timeline with error event", func(t *testing.T) {
		errorEvent := domain.NewEvent(domain.EventAgentError, "agent1", "Agent 1", nil).
			WithError(errors.New("test error"))

		events := []domain.Event{
			domain.NewEvent(domain.EventAgentStart, "agent1", "Agent 1", nil),
			errorEvent,
		}

		et := NewEventTimeline(events)
		visualization := et.Visualize()

		assert.Contains(t, visualization, "Error: test error")
	})
}

func TestEventHelpers(t *testing.T) {
	t.Run("create progress event", func(t *testing.T) {
		event := CreateProgressEvent("agent1", "Agent 1", 50, 100, "Processing...")

		assert.Equal(t, domain.EventProgress, event.Type)
		assert.Equal(t, "agent1", event.AgentID)
		assert.Equal(t, "Agent 1", event.AgentName)

		data, ok := event.Data.(domain.ProgressEventData)
		require.True(t, ok)
		assert.Equal(t, 50, data.Current)
		assert.Equal(t, 100, data.Total)
		assert.Equal(t, "Processing...", data.Message)
	})

	t.Run("create tool call event", func(t *testing.T) {
		input := map[string]interface{}{"expression": "2+2"}
		event := CreateToolCallEvent("agent1", "Agent 1", "calculator", input)

		assert.Equal(t, domain.EventToolCall, event.Type)

		data, ok := event.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "calculator", data["tool"])
		assert.Equal(t, input, data["input"])
	})

	t.Run("create error event", func(t *testing.T) {
		err := errors.New("test error")
		event := CreateErrorEvent("agent1", "Agent 1", err)

		assert.Equal(t, domain.EventAgentError, event.Type)
		assert.Equal(t, err, event.Error)
	})
}
