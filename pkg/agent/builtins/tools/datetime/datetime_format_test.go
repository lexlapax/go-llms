package datetime

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestDateTimeFormat(t *testing.T) {
	ctx := context.Background()
	tool := DateTimeFormat()

	t.Run("standard formats", func(t *testing.T) {
		input := DateTimeFormatInput{
			DateTime:       "2024-01-15T10:30:00Z",
			FormatType:     "standard",
			StandardFormat: "RFC3339",
		}
		result, err := tool.Execute(createTestToolContext(ctx), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeFormatOutput)
		if output.Formatted != "2024-01-15T10:30:00Z" {
			t.Errorf("Expected RFC3339 format, got %s", output.Formatted)
		}

		// Test other standard formats
		standardFormats := []string{"RFC1123", "RFC822", "Kitchen", "ISO8601"}
		for _, format := range standardFormats {
			t.Run(format, func(t *testing.T) {
				input := DateTimeFormatInput{
					DateTime:       "2024-01-15T10:30:00Z",
					FormatType:     "standard",
					StandardFormat: format,
				}
				result, err := tool.Execute(createTestToolContext(ctx), input)
				if err != nil {
					t.Fatalf("Failed to execute tool: %v", err)
				}

				output := result.(*DateTimeFormatOutput)
				if output.Formatted == "" {
					t.Error("Expected non-empty formatted output")
				}
			})
		}
	})

	t.Run("custom format", func(t *testing.T) {
		input := DateTimeFormatInput{
			DateTime:     "2024-01-15T10:30:00Z",
			FormatType:   "custom",
			CustomFormat: "Monday, January 2, 2006 at 3:04 PM",
		}
		result, err := tool.Execute(createTestToolContext(ctx), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeFormatOutput)
		expected := "Monday, January 15, 2024 at 10:30 AM"
		if output.Formatted != expected {
			t.Errorf("Expected '%s', got '%s'", expected, output.Formatted)
		}
	})

	t.Run("relative time format", func(t *testing.T) {
		// Test various relative times
		testCases := []struct {
			name     string
			dateTime time.Time
			expected []string // possible expected outputs
		}{
			{
				name:     "a few seconds ago",
				dateTime: time.Now().Add(-30 * time.Second),
				expected: []string{"a few seconds ago"},
			},
			{
				name:     "minutes ago",
				dateTime: time.Now().Add(-5 * time.Minute),
				expected: []string{"5 minutes ago"},
			},
			{
				name:     "hours ago",
				dateTime: time.Now().Add(-2 * time.Hour),
				expected: []string{"2 hours ago"},
			},
			{
				name:     "yesterday",
				dateTime: time.Now().Add(-24 * time.Hour),
				expected: []string{"yesterday at"},
			},
			{
				name:     "tomorrow",
				dateTime: time.Now().Add(24 * time.Hour),
				expected: []string{"tomorrow at", "in 23 hours", "in 24 hours", "in 1 day"},
			},
			{
				name:     "in a few days",
				dateTime: time.Now().Add(3 * 24 * time.Hour),
				expected: []string{"in 3 days", "in 2 days"},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				input := DateTimeFormatInput{
					DateTime:   tc.dateTime.Format(time.RFC3339),
					FormatType: "relative",
				}
				result, err := tool.Execute(createTestToolContext(ctx), input)
				if err != nil {
					t.Fatalf("Failed to execute tool: %v", err)
				}

				output := result.(*DateTimeFormatOutput)
				if output.RelativeTime == "" {
					t.Error("Expected non-empty relative time")
				}

				// Check if output contains expected string
				found := false
				for _, expected := range tc.expected {
					if strings.Contains(output.RelativeTime, expected) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected relative time to contain one of %v, got '%s'", tc.expected, output.RelativeTime)
				}
			})
		}
	})

	t.Run("relative time with weekday", func(t *testing.T) {
		yesterday := time.Now().Add(-24 * time.Hour)
		input := DateTimeFormatInput{
			DateTime:       yesterday.Format(time.RFC3339),
			FormatType:     "relative",
			IncludeWeekday: true,
		}
		result, err := tool.Execute(createTestToolContext(ctx), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeFormatOutput)
		// Should include weekday name in parentheses
		if !strings.Contains(output.RelativeTime, "(") || !strings.Contains(output.RelativeTime, ")") {
			t.Errorf("Expected weekday in parentheses, got '%s'", output.RelativeTime)
		}
	})

	t.Run("multiple formats", func(t *testing.T) {
		input := DateTimeFormatInput{
			DateTime:   "2024-01-15T10:30:00Z",
			FormatType: "multiple",
			Formats:    []string{"RFC3339", "2006-01-02", "Kitchen"},
		}
		result, err := tool.Execute(createTestToolContext(ctx), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeFormatOutput)
		if output.MultipleFormats == nil {
			t.Fatal("Expected multiple formats output")
		}

		// Check specific formats
		if output.MultipleFormats["RFC3339"] != "2024-01-15T10:30:00Z" {
			t.Errorf("Incorrect RFC3339 format: %s", output.MultipleFormats["RFC3339"])
		}
		if output.MultipleFormats["2006-01-02"] != "2024-01-15" {
			t.Errorf("Incorrect date format: %s", output.MultipleFormats["2006-01-02"])
		}
		if output.MultipleFormats["Kitchen"] != "10:30AM" {
			t.Errorf("Incorrect Kitchen format: %s", output.MultipleFormats["Kitchen"])
		}
	})

	t.Run("with timezone", func(t *testing.T) {
		input := DateTimeFormatInput{
			DateTime:     "2024-01-15T10:30:00Z",
			FormatType:   "custom",
			CustomFormat: "15:04 MST",
			Timezone:     "America/New_York",
		}
		result, err := tool.Execute(createTestToolContext(ctx), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeFormatOutput)
		// Should show EST time
		if !strings.Contains(output.Formatted, "EST") && !strings.Contains(output.Formatted, "EDT") {
			t.Errorf("Expected Eastern timezone in output, got '%s'", output.Formatted)
		}
	})

	t.Run("localized components", func(t *testing.T) {
		testCases := []struct {
			locale      string
			monthName   string
			weekdayName string
		}{
			{"es", "enero", "lunes"},   // Spanish
			{"fr", "janvier", "lundi"}, // French
			{"de", "Januar", "Montag"}, // German
		}

		for _, tc := range testCases {
			t.Run(tc.locale, func(t *testing.T) {
				input := DateTimeFormatInput{
					DateTime:   "2024-01-15T10:30:00Z", // Monday, January 15
					FormatType: "standard",
					Locale:     tc.locale,
				}
				result, err := tool.Execute(createTestToolContext(ctx), input)
				if err != nil {
					t.Fatalf("Failed to execute tool: %v", err)
				}

				output := result.(*DateTimeFormatOutput)
				if output.Localized == nil {
					t.Fatal("Expected localized components")
				}

				if output.Localized.MonthName != tc.monthName {
					t.Errorf("Expected month name '%s' for locale %s, got '%s'",
						tc.monthName, tc.locale, output.Localized.MonthName)
				}
				if output.Localized.WeekdayName != tc.weekdayName {
					t.Errorf("Expected weekday name '%s' for locale %s, got '%s'",
						tc.weekdayName, tc.locale, output.Localized.WeekdayName)
				}
			})
		}
	})

	t.Run("auto detect format type", func(t *testing.T) {
		// With custom format
		input := DateTimeFormatInput{
			DateTime:     "2024-01-15T10:30:00Z",
			CustomFormat: "2006-01-02",
			// FormatType not specified
		}
		result, err := tool.Execute(createTestToolContext(ctx), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeFormatOutput)
		if output.Formatted != "2024-01-15" {
			t.Errorf("Expected auto-detected custom format, got %s", output.Formatted)
		}

		// With multiple formats
		input = DateTimeFormatInput{
			DateTime: "2024-01-15T10:30:00Z",
			Formats:  []string{"RFC3339", "Kitchen"},
			// FormatType not specified
		}
		result, err = tool.Execute(createTestToolContext(ctx), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output = result.(*DateTimeFormatOutput)
		if output.MultipleFormats == nil {
			t.Error("Expected auto-detected multiple format type")
		}
	})

	t.Run("default multiple formats", func(t *testing.T) {
		input := DateTimeFormatInput{
			DateTime:   "2024-01-15T10:30:00Z",
			FormatType: "multiple",
			// No formats specified
		}
		result, err := tool.Execute(createTestToolContext(ctx), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeFormatOutput)
		if output.MultipleFormats == nil {
			t.Fatal("Expected multiple formats output")
		}

		// Should have default formats
		expectedKeys := []string{"RFC3339", "2006-01-02", "January 2, 2006", "15:04:05", "Monday"}
		for _, key := range expectedKeys {
			if _, ok := output.MultipleFormats[key]; !ok {
				t.Errorf("Expected default format key '%s'", key)
			}
		}
	})

	t.Run("invalid format type", func(t *testing.T) {
		input := DateTimeFormatInput{
			DateTime:   "2024-01-15T10:30:00Z",
			FormatType: "invalid",
		}
		_, err := tool.Execute(createTestToolContext(ctx), input)
		if err == nil {
			t.Error("Expected error for invalid format type")
		}
	})

	t.Run("missing custom format", func(t *testing.T) {
		input := DateTimeFormatInput{
			DateTime:   "2024-01-15T10:30:00Z",
			FormatType: "custom",
			// CustomFormat not provided
		}
		_, err := tool.Execute(createTestToolContext(ctx), input)
		if err == nil {
			t.Error("Expected error for missing custom format")
		}
	})

	t.Run("invalid datetime", func(t *testing.T) {
		input := DateTimeFormatInput{
			DateTime:   "invalid-date",
			FormatType: "standard",
		}
		_, err := tool.Execute(createTestToolContext(ctx), input)
		if err == nil {
			t.Error("Expected error for invalid datetime")
		}
	})
}

func TestFormatRelativeTime(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name             string
		time             time.Time
		expectedContains string
		includeWeekday   bool
	}{
		{
			name:             "seconds ago",
			time:             now.Add(-30 * time.Second),
			expectedContains: "seconds ago",
		},
		{
			name:             "1 minute ago",
			time:             now.Add(-1 * time.Minute),
			expectedContains: "1 minute ago",
		},
		{
			name:             "5 minutes ago",
			time:             now.Add(-5 * time.Minute),
			expectedContains: "5 minutes ago",
		},
		{
			name:             "1 hour ago",
			time:             now.Add(-1 * time.Hour),
			expectedContains: "1 hour ago",
		},
		{
			name:             "2 hours ago",
			time:             now.Add(-2 * time.Hour),
			expectedContains: "2 hours ago",
		},
		{
			name:             "yesterday",
			time:             now.Add(-24 * time.Hour),
			expectedContains: "yesterday at",
		},
		{
			name:             "tomorrow",
			time:             now.Add(24 * time.Hour),
			expectedContains: "tomorrow at",
		},
		{
			name:             "in 3 days",
			time:             now.Add(3 * 24 * time.Hour),
			expectedContains: "in 3 days",
		},
		{
			name:             "1 week ago",
			time:             now.Add(-7 * 24 * time.Hour),
			expectedContains: "1 week ago",
		},
		{
			name:             "in 2 weeks",
			time:             now.Add(14 * 24 * time.Hour),
			expectedContains: "in 2 weeks",
		},
		{
			name:             "1 month ago",
			time:             now.Add(-31 * 24 * time.Hour),
			expectedContains: "1 month ago",
		},
		{
			name:             "in 1 year",
			time:             now.Add(365 * 24 * time.Hour),
			expectedContains: "in 1 year",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := formatRelativeTime(tc.time, now, tc.includeWeekday)

			// For "today at" case, check more specifically
			if false && tc.expectedContains == "today at" {
				// If the time difference is less than a day but crosses midnight, it might not be "today"
				nowDate := now.Truncate(24 * time.Hour)
				tcDate := tc.time.Truncate(24 * time.Hour)
				if nowDate.Equal(tcDate) {
					if !strings.Contains(result, tc.expectedContains) {
						t.Errorf("Expected result to contain '%s', got '%s'", tc.expectedContains, result)
					}
				}
			} else {
				if !strings.Contains(result, tc.expectedContains) {
					t.Errorf("Expected result to contain '%s', got '%s'", tc.expectedContains, result)
				}
			}
		})
	}

	// Test with weekday
	t.Run("with weekday", func(t *testing.T) {
		yesterday := now.Add(-24 * time.Hour)
		result := formatRelativeTime(yesterday, now, true)

		// Should contain weekday in parentheses
		if !strings.Contains(result, "(") || !strings.Contains(result, ")") {
			t.Errorf("Expected weekday in parentheses, got '%s'", result)
		}
	})
}

func TestGetLocalizedComponents(t *testing.T) {
	// Monday, January 15, 2024 10:30 AM
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	testCases := []struct {
		locale      string
		monthName   string
		weekdayName string
	}{
		{"en", "January", "Monday"},      // Default English
		{"es", "enero", "lunes"},         // Spanish
		{"fr", "janvier", "lundi"},       // French
		{"de", "Januar", "Montag"},       // German
		{"unknown", "January", "Monday"}, // Unknown locale defaults to English
	}

	for _, tc := range testCases {
		t.Run(tc.locale, func(t *testing.T) {
			components := getLocalizedComponents(testTime, tc.locale)

			if components.MonthName != tc.monthName {
				t.Errorf("Expected month name '%s' for locale %s, got '%s'",
					tc.monthName, tc.locale, components.MonthName)
			}
			if components.WeekdayName != tc.weekdayName {
				t.Errorf("Expected weekday name '%s' for locale %s, got '%s'",
					tc.weekdayName, tc.locale, components.WeekdayName)
			}
			if components.Period != "AM" {
				t.Errorf("Expected AM period, got '%s'", components.Period)
			}
		})
	}

	// Test PM period
	pmTime := time.Date(2024, 1, 15, 15, 30, 0, 0, time.UTC)
	components := getLocalizedComponents(pmTime, "en")
	if components.Period != "PM" {
		t.Errorf("Expected PM period, got '%s'", components.Period)
	}
}
