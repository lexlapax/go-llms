package datetime

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestDateTimeParse(t *testing.T) {
	ctx := context.Background()
	tool := DateTimeParse()

	t.Run("common formats", func(t *testing.T) {
		testCases := []struct {
			dateString string
			valid      bool
		}{
			{"2024-01-15T10:30:00Z", true},          // RFC3339
			{"2024-01-15", true},                    // ISO Date
			{"2024-01-15 10:30:00", true},           // DateTime
			{"01/15/2024", true},                    // US Date
			{"15-Jan-2024", true},                   // DD-Mon-YYYY
			{"Jan 15, 2024", true},                  // Mon D, YYYY
			{"January 15, 2024", true},              // Month D, YYYY
			{"January 15, 2024 3:04 PM", true},      // With time
			{"2024/01/15", true},                    // YYYY/MM/DD
			{"20240115", true},                      // YYYYMMDD
			{"Mon, 02 Jan 2006 15:04:05 MST", true}, // RFC1123
			{"invalid-date", false},                 // Invalid
		}

		for _, tc := range testCases {
			t.Run(tc.dateString, func(t *testing.T) {
				input := DateTimeParseInput{
					DateString: tc.dateString,
					AutoDetect: true,
				}
				result, err := tool.Execute(ctx, input)
				if err != nil {
					t.Fatalf("Failed to execute tool: %v", err)
				}

				output := result.(*DateTimeParseOutput)
				if output.Valid != tc.valid {
					t.Errorf("Expected valid=%v for %s, got %v", tc.valid, tc.dateString, output.Valid)
				}

				if output.Valid {
					// Verify we got a valid RFC3339 formatted result
					_, err := time.Parse(time.RFC3339, output.Parsed)
					if err != nil {
						t.Errorf("Failed to parse output as RFC3339: %v", err)
					}

					// Check that we detected a format
					if output.DetectedFormat == "" {
						t.Error("Expected detected format to be set")
					}

					// Check Unix timestamp
					if output.UnixTimestamp == 0 {
						t.Error("Expected non-zero Unix timestamp")
					}
				}
			})
		}
	})

	t.Run("custom format", func(t *testing.T) {
		input := DateTimeParseInput{
			DateString: "15/01/2024 10:30",
			Format:     "02/01/2006 15:04",
		}
		result, err := tool.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeParseOutput)
		if !output.Valid {
			t.Error("Expected valid parse with custom format")
		}
		if output.DetectedFormat != "custom format" {
			t.Errorf("Expected 'custom format', got %s", output.DetectedFormat)
		}

		// Verify the parsed date
		parsedTime, _ := time.Parse(time.RFC3339, output.Parsed)
		if parsedTime.Day() != 15 || parsedTime.Month() != 1 || parsedTime.Year() != 2024 {
			t.Errorf("Incorrect parsed date: %s", output.Parsed)
		}
	})

	t.Run("with timezone", func(t *testing.T) {
		input := DateTimeParseInput{
			DateString: "2024-01-15 10:30:00",
			Timezone:   "America/New_York",
		}
		result, err := tool.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeParseOutput)
		if !output.Valid {
			t.Error("Expected valid parse")
		}

		// The parsed time should reflect EST/EDT offset
		if !strings.Contains(output.Parsed, "-05:00") && !strings.Contains(output.Parsed, "-04:00") {
			t.Errorf("Expected Eastern timezone offset in result: %s", output.Parsed)
		}
	})

	t.Run("relative dates", func(t *testing.T) {
		// Use a fixed reference time for consistent testing
		referenceTime := "2024-01-15T10:30:00Z"

		testCases := []struct {
			dateString     string
			expectedDelta  time.Duration
			expectedFormat string
		}{
			{"now", 0, "relative date"},
			{"today", 0, "relative date"}, // Will be start of day
			{"yesterday", -24 * time.Hour, "relative date"},
			{"tomorrow", 24 * time.Hour, "relative date"},
			{"in 3 days", 3 * 24 * time.Hour, "relative date"},
			{"5 days ago", -5 * 24 * time.Hour, "relative date"},
			{"in 2 weeks", 14 * 24 * time.Hour, "relative date"},
			{"1 week ago", -7 * 24 * time.Hour, "relative date"},
			{"in 2 months", 0, "relative date"}, // Months are variable
			{"3 months ago", 0, "relative date"},
			{"in 5 hours", 5 * time.Hour, "relative date"},
			{"2 hours ago", -2 * time.Hour, "relative date"},
			{"next monday", 0, "relative date"}, // Day-specific
			{"last friday", 0, "relative date"},
			{"next week", 7 * 24 * time.Hour, "relative date"},
			{"last week", -7 * 24 * time.Hour, "relative date"},
		}

		refTime, _ := time.Parse(time.RFC3339, referenceTime)

		for _, tc := range testCases {
			t.Run(tc.dateString, func(t *testing.T) {
				input := DateTimeParseInput{
					DateString:    tc.dateString,
					ReferenceTime: referenceTime,
				}
				result, err := tool.Execute(ctx, input)
				if err != nil {
					t.Fatalf("Failed to execute tool: %v", err)
				}

				output := result.(*DateTimeParseOutput)
				if !output.Valid {
					t.Errorf("Expected valid parse for relative date '%s'", tc.dateString)
				}
				if output.DetectedFormat != tc.expectedFormat {
					t.Errorf("Expected format '%s', got '%s'", tc.expectedFormat, output.DetectedFormat)
				}

				// For specific delta checks (not month-based or weekday-based)
				if tc.expectedDelta != 0 && !strings.Contains(tc.dateString, "month") &&
					!strings.Contains(tc.dateString, "monday") && !strings.Contains(tc.dateString, "friday") {
					parsedTime, _ := time.Parse(time.RFC3339, output.Parsed)
					actualDelta := parsedTime.Sub(refTime)

					// For "today", adjust for start of day
					if tc.dateString == "today" {
						expectedTime := time.Date(refTime.Year(), refTime.Month(), refTime.Day(), 0, 0, 0, 0, refTime.Location())
						actualDelta = parsedTime.Sub(expectedTime)
					}

					if actualDelta != tc.expectedDelta {
						t.Errorf("Expected delta %v for '%s', got %v", tc.expectedDelta, tc.dateString, actualDelta)
					}
				}
			})
		}
	})

	t.Run("unix timestamps", func(t *testing.T) {
		testCases := []struct {
			timestamp string
			format    string
		}{
			{"1705315800", "Unix timestamp (seconds)"},         // 2024-01-15 10:30:00 UTC
			{"1705315800000", "Unix timestamp (milliseconds)"}, // Same time in milliseconds
		}

		for _, tc := range testCases {
			t.Run(tc.format, func(t *testing.T) {
				input := DateTimeParseInput{
					DateString: tc.timestamp,
				}
				result, err := tool.Execute(ctx, input)
				if err != nil {
					t.Fatalf("Failed to execute tool: %v", err)
				}

				output := result.(*DateTimeParseOutput)
				if !output.Valid {
					t.Error("Expected valid parse for Unix timestamp")
				}
				if output.DetectedFormat != tc.format {
					t.Errorf("Expected format '%s', got '%s'", tc.format, output.DetectedFormat)
				}

				// Verify the parsed time is correct
				parsedTime, _ := time.Parse(time.RFC3339, output.Parsed)
				if parsedTime.Year() != 2024 || parsedTime.Month() != 1 || parsedTime.Day() != 15 {
					t.Errorf("Incorrect parsed date from Unix timestamp: %s", output.Parsed)
				}
			})
		}
	})

	t.Run("validation errors", func(t *testing.T) {
		input := DateTimeParseInput{
			DateString: "not-a-date",
			Format:     "2006-01-02", // Try with specific format
		}
		result, err := tool.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeParseOutput)
		if output.Valid {
			t.Error("Expected invalid parse")
		}
		if len(output.ValidationErrors) == 0 {
			t.Error("Expected validation errors")
		}
	})

	t.Run("ambiguous date formats", func(t *testing.T) {
		// Test date that could be US or EU format
		input := DateTimeParseInput{
			DateString: "01/02/2024",
		}
		result, err := tool.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeParseOutput)
		if !output.Valid {
			t.Error("Expected valid parse")
		}

		// Should parse as US format (January 2) since it comes first in our list
		parsedTime, _ := time.Parse(time.RFC3339, output.Parsed)
		if parsedTime.Month() != 1 || parsedTime.Day() != 2 {
			t.Errorf("Expected January 2, got %s", output.Parsed)
		}
	})

	t.Run("relative weekdays", func(t *testing.T) {
		// Use Monday Jan 15, 2024 as reference
		referenceTime := "2024-01-15T10:30:00Z"

		testCases := []struct {
			dateString  string
			expectedDay int
		}{
			{"next monday", 22}, // Next Monday is Jan 22
			{"next friday", 19}, // Next Friday is Jan 19
			{"last monday", 8},  // Last Monday was Jan 8
			{"last friday", 12}, // Last Friday was Jan 12
		}

		for _, tc := range testCases {
			t.Run(tc.dateString, func(t *testing.T) {
				input := DateTimeParseInput{
					DateString:    tc.dateString,
					ReferenceTime: referenceTime,
				}
				result, err := tool.Execute(ctx, input)
				if err != nil {
					t.Fatalf("Failed to execute tool: %v", err)
				}

				output := result.(*DateTimeParseOutput)
				if !output.Valid {
					t.Errorf("Expected valid parse for '%s'", tc.dateString)
				}

				parsedTime, _ := time.Parse(time.RFC3339, output.Parsed)
				if parsedTime.Day() != tc.expectedDay {
					t.Errorf("Expected day %d for '%s', got %d", tc.expectedDay, tc.dateString, parsedTime.Day())
				}
			})
		}
	})

	t.Run("disable auto detect", func(t *testing.T) {
		input := DateTimeParseInput{
			DateString: "2024-01-15",
			Format:     "2006/01/02", // Wrong format
			AutoDetect: false,
		}
		result, err := tool.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeParseOutput)
		if output.Valid {
			t.Error("Expected invalid parse when auto-detect is disabled and format doesn't match")
		}
	})

	t.Run("special relative dates", func(t *testing.T) {
		referenceTime := "2024-01-15T10:30:00Z"

		testCases := []string{
			"next month",
			"last month",
			"next year",
			"last year",
		}

		for _, dateString := range testCases {
			t.Run(dateString, func(t *testing.T) {
				input := DateTimeParseInput{
					DateString:    dateString,
					ReferenceTime: referenceTime,
				}
				result, err := tool.Execute(ctx, input)
				if err != nil {
					t.Fatalf("Failed to execute tool: %v", err)
				}

				output := result.(*DateTimeParseOutput)
				if !output.Valid {
					t.Errorf("Expected valid parse for '%s'", dateString)
				}
			})
		}
	})
}

func TestParseRelativeDate(t *testing.T) {
	referenceTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	testCases := []struct {
		input    string
		expected time.Time
		valid    bool
	}{
		{"now", referenceTime, true},
		{"today", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), true},
		{"yesterday", referenceTime.AddDate(0, 0, -1), true},
		{"tomorrow", referenceTime.AddDate(0, 0, 1), true},
		{"in 3 days", referenceTime.AddDate(0, 0, 3), true},
		{"5 days ago", referenceTime.AddDate(0, 0, -5), true},
		{"next week", referenceTime.AddDate(0, 0, 7), true},
		{"last week", referenceTime.AddDate(0, 0, -7), true},
		{"random text", time.Time{}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result, valid := parseRelativeDate(tc.input, referenceTime)
			if valid != tc.valid {
				t.Errorf("Expected valid=%v for '%s', got %v", tc.valid, tc.input, valid)
			}
			if valid && !result.Equal(tc.expected) {
				t.Errorf("Expected %v for '%s', got %v", tc.expected, tc.input, result)
			}
		})
	}
}
