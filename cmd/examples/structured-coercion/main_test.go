package main

import (
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/schema/validation"
)

func TestCoercionFunctions(t *testing.T) {
	// Test UUID coercion
	t.Run("UUID coercion", func(t *testing.T) {
		uuidStr := "f47ac10b-58cc-0372-8567-0e02b2c3d479"
		uuidVal, ok := validation.CoerceToUUID(uuidStr)
		if !ok {
			t.Errorf("Failed to coerce UUID from valid string")
		}
		if uuidVal.String() != uuidStr {
			t.Errorf("Coerced UUID value doesn't match input: got %s, want %s", uuidVal.String(), uuidStr)
		}

		// Test invalid UUID
		_, ok = validation.CoerceToUUID("not-a-uuid")
		if ok {
			t.Errorf("Expected failure when coercing invalid UUID")
		}
	})

	// Test Date coercion
	t.Run("Date coercion", func(t *testing.T) {
		// Test ISO date format
		dateStr := "2023-09-15"
		dateVal, ok := validation.CoerceToDate(dateStr)
		if !ok {
			t.Errorf("Failed to coerce date from valid string")
		}
		expected := time.Date(2023, 9, 15, 0, 0, 0, 0, time.UTC)
		if !dateVal.Equal(expected) {
			t.Errorf("Coerced date value doesn't match expected: got %v, want %v", dateVal, expected)
		}

		// Test Unix timestamp
		timestamp := int64(1694883600)
		dateVal, ok = validation.CoerceToDate(timestamp)
		if !ok {
			t.Errorf("Failed to coerce date from valid timestamp")
		}
		if dateVal.Unix() != timestamp {
			t.Errorf("Coerced date value doesn't match timestamp: got %d, want %d", dateVal.Unix(), timestamp)
		}
	})

	// Test Email coercion
	t.Run("Email coercion", func(t *testing.T) {
		// Test simple email
		emailStr := "john.doe@example.com"
		emailVal, ok := validation.CoerceToEmail(emailStr)
		if !ok {
			t.Errorf("Failed to coerce email from valid string")
		}
		if emailVal != emailStr {
			t.Errorf("Coerced email value doesn't match input: got %s, want %s", emailVal, emailStr)
		}

		// Test email with display name
		emailWithName := "John Doe <john.doe@example.com>"
		emailVal, ok = validation.CoerceToEmail(emailWithName)
		if !ok {
			t.Errorf("Failed to coerce email from string with display name")
		}
		if emailVal != "john.doe@example.com" {
			t.Errorf("Coerced email value doesn't match expected: got %s, want %s", emailVal, "john.doe@example.com")
		}
	})

	// Test URL coercion
	t.Run("URL coercion", func(t *testing.T) {
		// Test URL without scheme
		urlStr := "example.com"
		urlVal, ok := validation.CoerceToURL(urlStr)
		if !ok {
			t.Errorf("Failed to coerce URL from valid string without scheme")
		}
		if urlVal.String() != "http://example.com" {
			t.Errorf("Coerced URL value doesn't match expected: got %s, want %s", urlVal.String(), "http://example.com")
		}

		// Test URL with scheme
		urlWithScheme := "https://example.com/path?query=value"
		urlVal, ok = validation.CoerceToURL(urlWithScheme)
		if !ok {
			t.Errorf("Failed to coerce URL from valid string with scheme")
		}
		if urlVal.String() != urlWithScheme {
			t.Errorf("Coerced URL value doesn't match input: got %s, want %s", urlVal.String(), urlWithScheme)
		}
	})

	// Test Duration coercion
	t.Run("Duration coercion", func(t *testing.T) {
		// Test Go duration format
		durationStr := "1h30m"
		expected := 90 * time.Minute
		durationVal, ok := validation.CoerceToDuration(durationStr)
		if !ok {
			t.Errorf("Failed to coerce duration from valid Go format")
		}
		if durationVal != expected {
			t.Errorf("Coerced duration value doesn't match expected: got %v, want %v", durationVal, expected)
		}

		// Test HH:MM format
		durationStr = "1:30"
		durationVal, ok = validation.CoerceToDuration(durationStr)
		if !ok {
			t.Errorf("Failed to coerce duration from valid HH:MM format")
		}
		if durationVal != expected {
			t.Errorf("Coerced duration value doesn't match expected: got %v, want %v", durationVal, expected)
		}

		// Test natural language format
		durationStr = "2 days"
		expected = 48 * time.Hour
		durationVal, ok = validation.CoerceToDuration(durationStr)
		if !ok {
			t.Errorf("Failed to coerce duration from valid natural language format")
		}
		if durationVal != expected {
			t.Errorf("Coerced duration value doesn't match expected: got %v, want %v", durationVal, expected)
		}
	})

	// Test Array coercion
	t.Run("Array coercion", func(t *testing.T) {
		// Test comma-separated string
		arrayStr := "item1, item2, item3"
		arrayVal, ok := validation.CoerceToArray(arrayStr)
		if !ok {
			t.Errorf("Failed to coerce array from comma-separated string")
		}
		if len(arrayVal) != 3 {
			t.Errorf("Coerced array has wrong length: got %d, want %d", len(arrayVal), 3)
		}

		// Test JSON array string
		jsonArrayStr := `["item1", "item2", "item3"]`
		arrayVal, ok = validation.CoerceToArray(jsonArrayStr)
		if !ok {
			t.Errorf("Failed to coerce array from JSON array string")
		}
		if len(arrayVal) != 3 {
			t.Errorf("Coerced array has wrong length: got %d, want %d", len(arrayVal), 3)
		}
	})

	// Test Object coercion
	t.Run("Object coercion", func(t *testing.T) {
		// Test JSON object string
		objectStr := `{"name": "test", "value": 123}`
		objectVal, ok := validation.CoerceToObject(objectStr)
		if !ok {
			t.Errorf("Failed to coerce object from JSON object string")
		}
		if len(objectVal) != 2 {
			t.Errorf("Coerced object has wrong number of properties: got %d, want %d", len(objectVal), 2)
		}
		if name, exists := objectVal["name"]; !exists || name != "test" {
			t.Errorf("Coerced object has wrong 'name' property: got %v", name)
		}
	})
}

func TestSchemaValidationWithCoercion(t *testing.T) {
	// Create schema
	schema := createEventSchema()
	validator := validation.NewValidator(validation.WithCoercion(true))

	// Test well-formed event
	t.Run("Well-formed event", func(t *testing.T) {
		eventJSON := `{
			"id": "f47ac10b-58cc-0372-8567-0e02b2c3d479",
			"name": "Go Conference 2023",
			"description": "Annual conference for Go developers",
			"startDate": "2023-09-15T09:00:00Z",
			"endDate": "2023-09-16T18:00:00Z",
			"location": "San Francisco, CA",
			"url": "https://goconf.example.com",
			"duration": "2 days",
			"maxAttendees": 500,
			"categories": ["programming", "conference", "go"],
			"tags": ["golang", "developer", "tech"],
			"isFree": false,
			"organizer": {
				"name": "Go Developer Community",
				"email": "organizers@goconf.example.com"
			}
		}`

		result, err := validator.Validate(schema, eventJSON)
		if err != nil {
			t.Errorf("Validation error: %v", err)
		}
		if !result.Valid {
			t.Errorf("Expected valid result, got invalid: %v", result.Errors)
		}
	})

	// Test event with type mismatches that require coercion
	t.Run("Event with type mismatches", func(t *testing.T) {
		eventWithTypeMismatch := `{
			"id": "f47ac10b-58cc-0372-8567-0e02b2c3d479",
			"name": "Go Conference 2023",
			"description": "Annual conference for Go developers",
			"startDate": "2023-09-15",
			"endDate": 1694883600,
			"location": "San Francisco, CA",
			"url": "goconf.example.com",
			"duration": "1:30:00",
			"maxAttendees": "500",
			"categories": "programming, conference, go",
			"tags": ["golang", "developer", "tech"],
			"isFree": "false",
			"organizer": {
				"name": "Go Developer Community",
				"email": "organizers@goconf.example.com"
			}
		}`

		result, err := validator.Validate(schema, eventWithTypeMismatch)
		if err != nil {
			t.Errorf("Validation error: %v", err)
		}

		// The validation should succeed because of coercion
		if !result.Valid {
			t.Errorf("Expected valid result after coercion, got invalid: %v", result.Errors)
		}
	})
}
