package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/schema/adapter/reflection"
)

func TestTaskSchema(t *testing.T) {
	schema, err := reflection.GenerateSchema(Task{})
	if err != nil {
		t.Fatalf("Failed to generate Task schema: %v", err)
	}

	// Verify schema has required properties
	if schema.Type != "object" {
		t.Errorf("Expected schema type 'object', got '%s'", schema.Type)
	}

	// Check required fields
	expectedRequired := []string{"id", "title", "status", "priority", "created_at"}
	for _, field := range expectedRequired {
		found := false
		for _, required := range schema.Required {
			if required == field {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected field '%s' to be required in schema", field)
		}
	}

	// Check properties exist
	if _, exists := schema.Properties["id"]; !exists {
		t.Error("Expected 'id' property in schema")
	}
	if _, exists := schema.Properties["status"]; !exists {
		t.Error("Expected 'status' property in schema")
	}
	if _, exists := schema.Properties["priority"]; !exists {
		t.Error("Expected 'priority' property in schema")
	}
}

func TestProjectAnalysisSchema(t *testing.T) {
	schema, err := reflection.GenerateSchema(ProjectAnalysis{})
	if err != nil {
		t.Fatalf("Failed to generate ProjectAnalysis schema: %v", err)
	}

	// Verify schema structure
	if schema.Type != "object" {
		t.Errorf("Expected schema type 'object', got '%s'", schema.Type)
	}

	// Check for key properties
	requiredProperties := []string{
		"project_name", "total_tasks", "completed_tasks",
		"completion_rate", "recommendations", "risks", "next_actions",
	}

	for _, prop := range requiredProperties {
		if _, exists := schema.Properties[prop]; !exists {
			t.Errorf("Expected property '%s' in ProjectAnalysis schema", prop)
		}
	}
}

func TestMeetingNotesSchema(t *testing.T) {
	schema, err := reflection.GenerateSchema(MeetingNotes{})
	if err != nil {
		t.Fatalf("Failed to generate MeetingNotes schema: %v", err)
	}

	// Verify nested object handling
	if schema.Type != "object" {
		t.Errorf("Expected schema type 'object', got '%s'", schema.Type)
	}

	// Check action_items property for nested objects
	actionItemsProp, exists := schema.Properties["action_items"]
	if !exists {
		t.Error("Expected 'action_items' property in MeetingNotes schema")
	}

	if actionItemsProp.Type != "array" {
		t.Errorf("Expected action_items type 'array', got '%s'", actionItemsProp.Type)
	}
}

func TestActionItemSchema(t *testing.T) {
	schema, err := reflection.GenerateSchema(ActionItem{})
	if err != nil {
		t.Fatalf("Failed to generate ActionItem schema: %v", err)
	}

	// Check enum validation for status and priority
	statusProp, exists := schema.Properties["status"]
	if !exists {
		t.Error("Expected 'status' property in ActionItem schema")
	}

	priorityProp, exists := schema.Properties["priority"]
	if !exists {
		t.Error("Expected 'priority' property in ActionItem schema")
	}

	// Verify these are string types (enums are represented as strings with constraints)
	if statusProp.Type != "string" {
		t.Errorf("Expected status type 'string', got '%s'", statusProp.Type)
	}
	if priorityProp.Type != "string" {
		t.Errorf("Expected priority type 'string', got '%s'", priorityProp.Type)
	}
}

func TestTaskJSONMarshaling(t *testing.T) {
	now := time.Now()
	task := Task{
		ID:             "test-001",
		Title:          "Test Task",
		Description:    "A test task for validation",
		Status:         TaskStatusPending,
		Priority:       PriorityHigh,
		DueDate:        &now,
		EstimatedHours: 5.5,
		Tags:           []string{"test", "validation"},
		CreatedAt:      now,
	}

	// Test marshaling
	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("Failed to marshal Task: %v", err)
	}

	// Test unmarshaling
	var unmarshaled Task
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal Task: %v", err)
	}

	// Verify data integrity
	if unmarshaled.ID != task.ID {
		t.Errorf("Expected ID '%s', got '%s'", task.ID, unmarshaled.ID)
	}
	if unmarshaled.Status != task.Status {
		t.Errorf("Expected status '%s', got '%s'", task.Status, unmarshaled.Status)
	}
	if unmarshaled.Priority != task.Priority {
		t.Errorf("Expected priority '%s', got '%s'", task.Priority, unmarshaled.Priority)
	}
}

func TestProjectAnalysisValidation(t *testing.T) {
	analysis := ProjectAnalysis{
		ProjectName:       "Test Project",
		TotalTasks:        10,
		CompletedTasks:    3,
		PendingTasks:      5,
		InProgressTasks:   2,
		HighPriorityTasks: 2,
		OverdueTasks:      1,
		EstimatedHours:    45.5,
		CompletionRate:    30.0,
		Recommendations:   []string{"Focus on high priority tasks", "Review overdue items"},
		Risks:             []string{"Resource constraints", "Timeline pressure"},
		NextActions:       []string{"Prioritize overdue tasks", "Allocate additional resources"},
	}

	// Test JSON marshaling
	data, err := json.Marshal(analysis)
	if err != nil {
		t.Fatalf("Failed to marshal ProjectAnalysis: %v", err)
	}

	var unmarshaled ProjectAnalysis
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal ProjectAnalysis: %v", err)
	}

	// Verify calculations make sense
	if unmarshaled.CompletedTasks+unmarshaled.PendingTasks+unmarshaled.InProgressTasks != unmarshaled.TotalTasks {
		t.Error("Task counts don't add up correctly")
	}

	if unmarshaled.CompletionRate < 0 || unmarshaled.CompletionRate > 100 {
		t.Errorf("Completion rate should be between 0-100, got %.2f", unmarshaled.CompletionRate)
	}
}

func TestMeetingNotesComplexStructure(t *testing.T) {
	dueDate := time.Now().Add(24 * time.Hour)
	nextMeeting := time.Now().Add(7 * 24 * time.Hour)

	notes := MeetingNotes{
		MeetingTitle: "Sprint Planning",
		Date:         time.Now(),
		Attendees:    []string{"Alice", "Bob", "Charlie"},
		Duration:     90,
		KeyDiscussions: []string{
			"Sprint goals review",
			"Resource allocation",
			"Technical challenges",
		},
		Decisions: []string{
			"Prioritize authentication feature",
			"Extend sprint by 2 days",
		},
		ActionItems: []ActionItem{
			{
				Description: "Update user stories",
				AssignedTo:  "Alice",
				DueDate:     &dueDate,
				Priority:    PriorityHigh,
				Status:      TaskStatusPending,
			},
			{
				Description: "Create technical design",
				AssignedTo:  "Bob",
				Priority:    PriorityMedium,
				Status:      TaskStatusInProgress,
			},
		},
		NextMeeting: &nextMeeting,
	}

	// Test JSON marshaling of complex nested structure
	data, err := json.Marshal(notes)
	if err != nil {
		t.Fatalf("Failed to marshal MeetingNotes: %v", err)
	}

	var unmarshaled MeetingNotes
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal MeetingNotes: %v", err)
	}

	// Verify nested data
	if len(unmarshaled.ActionItems) != 2 {
		t.Errorf("Expected 2 action items, got %d", len(unmarshaled.ActionItems))
	}

	if len(unmarshaled.Attendees) != 3 {
		t.Errorf("Expected 3 attendees, got %d", len(unmarshaled.Attendees))
	}

	// Verify action item details
	firstAction := unmarshaled.ActionItems[0]
	if firstAction.AssignedTo != "Alice" {
		t.Errorf("Expected first action assigned to 'Alice', got '%s'", firstAction.AssignedTo)
	}
	if firstAction.Priority != PriorityHigh {
		t.Errorf("Expected high priority, got '%s'", firstAction.Priority)
	}
}

func TestEnumValidation(t *testing.T) {
	// Test valid enum values
	validStatuses := []TaskStatus{
		TaskStatusPending,
		TaskStatusInProgress,
		TaskStatusCompleted,
		TaskStatusCancelled,
	}

	for _, status := range validStatuses {
		task := Task{
			ID:        "test",
			Title:     "Test",
			Status:    status,
			Priority:  PriorityMedium,
			CreatedAt: time.Now(),
		}

		data, err := json.Marshal(task)
		if err != nil {
			t.Errorf("Failed to marshal task with status '%s': %v", status, err)
		}

		var unmarshaled Task
		err = json.Unmarshal(data, &unmarshaled)
		if err != nil {
			t.Errorf("Failed to unmarshal task with status '%s': %v", status, err)
		}

		if unmarshaled.Status != status {
			t.Errorf("Status mismatch: expected '%s', got '%s'", status, unmarshaled.Status)
		}
	}
}

func TestOptionalFields(t *testing.T) {
	// Test task without optional fields
	task := Task{
		ID:        "test-minimal",
		Title:     "Minimal Task",
		Status:    TaskStatusPending,
		Priority:  PriorityLow,
		CreatedAt: time.Now(),
		// DueDate is nil (optional)
		// Tags is empty slice
		// Description is empty string
		EstimatedHours: 0,
	}

	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("Failed to marshal minimal task: %v", err)
	}

	var unmarshaled Task
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal minimal task: %v", err)
	}

	// Verify optional fields are properly handled
	if unmarshaled.DueDate != nil {
		t.Error("Expected DueDate to be nil")
	}
	if len(unmarshaled.Tags) != 0 {
		t.Errorf("Expected empty tags slice, got %d items", len(unmarshaled.Tags))
	}
	if unmarshaled.Description != "" {
		t.Errorf("Expected empty description, got '%s'", unmarshaled.Description)
	}
}
