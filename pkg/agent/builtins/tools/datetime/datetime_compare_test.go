package datetime

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestDateTimeCompare(t *testing.T) {
	ctx := context.Background()
	tool := DateTimeCompare()

	t.Run("compare operation", func(t *testing.T) {
		// Test date1 before date2
		input := DateTimeCompareInput{
			Operation: "compare",
			Date1:     "2024-01-15T10:30:00Z",
			Date2:     "2024-01-16T10:30:00Z",
		}
		result, err := tool.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeCompareOutput)
		if !output.Before {
			t.Error("Expected Before to be true")
		}
		if output.After {
			t.Error("Expected After to be false")
		}
		if output.Equal {
			t.Error("Expected Equal to be false")
		}

		// Check difference
		if output.Difference == nil {
			t.Fatal("Expected time difference")
		}
		if output.Difference.Days != 1 {
			t.Errorf("Expected 1 day difference, got %d", output.Difference.Days)
		}
		if output.Difference.HumanReadable != "in 1 day" {
			t.Errorf("Expected 'in 1 day', got '%s'", output.Difference.HumanReadable)
		}

		// Test date1 after date2
		input = DateTimeCompareInput{
			Operation: "compare",
			Date1:     "2024-01-16T10:30:00Z",
			Date2:     "2024-01-15T10:30:00Z",
		}
		result, err = tool.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output = result.(*DateTimeCompareOutput)
		if output.Before {
			t.Error("Expected Before to be false")
		}
		if !output.After {
			t.Error("Expected After to be true")
		}
		if output.Equal {
			t.Error("Expected Equal to be false")
		}

		if output.Difference.HumanReadable != "1 day ago" {
			t.Errorf("Expected '1 day ago', got '%s'", output.Difference.HumanReadable)
		}

		// Test equal dates
		input = DateTimeCompareInput{
			Operation: "compare",
			Date1:     "2024-01-15T10:30:00Z",
			Date2:     "2024-01-15T10:30:00Z",
		}
		result, err = tool.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output = result.(*DateTimeCompareOutput)
		if output.Before {
			t.Error("Expected Before to be false")
		}
		if output.After {
			t.Error("Expected After to be false")
		}
		if !output.Equal {
			t.Error("Expected Equal to be true")
		}
		if output.Difference.HumanReadable != "same time" {
			t.Errorf("Expected 'same time', got '%s'", output.Difference.HumanReadable)
		}
	})

	t.Run("compare with timezone", func(t *testing.T) {
		input := DateTimeCompareInput{
			Operation: "compare",
			Date1:     "2024-01-15T10:30:00Z",
			Date2:     "2024-01-15T10:30:00Z",
			Timezone:  "America/New_York",
		}
		result, err := tool.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeCompareOutput)
		if !output.Equal {
			t.Error("Expected Equal to be true when comparing same times in same timezone")
		}
	})

	t.Run("same period checks", func(t *testing.T) {
		// Same day
		input := DateTimeCompareInput{
			Operation:  "same_period",
			Date1:      "2024-01-15T10:30:00Z",
			Date2:      "2024-01-15T20:45:00Z",
			PeriodType: "day",
		}
		result, err := tool.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeCompareOutput)
		if !output.SamePeriod {
			t.Error("Expected SamePeriod to be true for same day")
		}

		// Different days
		input.Date2 = "2024-01-16T10:30:00Z"
		result, err = tool.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output = result.(*DateTimeCompareOutput)
		if output.SamePeriod {
			t.Error("Expected SamePeriod to be false for different days")
		}

		// Same week
		input = DateTimeCompareInput{
			Operation:  "same_period",
			Date1:      "2024-01-15T10:30:00Z", // Monday
			Date2:      "2024-01-19T10:30:00Z", // Friday
			PeriodType: "week",
		}
		result, err = tool.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output = result.(*DateTimeCompareOutput)
		if !output.SamePeriod {
			t.Error("Expected SamePeriod to be true for same week")
		}

		// Same month
		input = DateTimeCompareInput{
			Operation:  "same_period",
			Date1:      "2024-01-01T10:30:00Z",
			Date2:      "2024-01-31T10:30:00Z",
			PeriodType: "month",
		}
		result, err = tool.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output = result.(*DateTimeCompareOutput)
		if !output.SamePeriod {
			t.Error("Expected SamePeriod to be true for same month")
		}

		// Same year
		input = DateTimeCompareInput{
			Operation:  "same_period",
			Date1:      "2024-01-01T10:30:00Z",
			Date2:      "2024-12-31T10:30:00Z",
			PeriodType: "year",
		}
		result, err = tool.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output = result.(*DateTimeCompareOutput)
		if !output.SamePeriod {
			t.Error("Expected SamePeriod to be true for same year")
		}
	})

	t.Run("range check", func(t *testing.T) {
		// Date within range
		input := DateTimeCompareInput{
			Operation:  "range_check",
			Date1:      "2024-01-15T10:30:00Z",
			RangeStart: "2024-01-01T00:00:00Z",
			RangeEnd:   "2024-01-31T23:59:59Z",
		}
		result, err := tool.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeCompareOutput)
		if !output.InRange {
			t.Error("Expected InRange to be true")
		}

		// Date before range
		input.Date1 = "2023-12-31T10:30:00Z"
		result, err = tool.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output = result.(*DateTimeCompareOutput)
		if output.InRange {
			t.Error("Expected InRange to be false for date before range")
		}

		// Date after range
		input.Date1 = "2024-02-01T10:30:00Z"
		result, err = tool.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output = result.(*DateTimeCompareOutput)
		if output.InRange {
			t.Error("Expected InRange to be false for date after range")
		}

		// Date exactly at range start
		input.Date1 = "2024-01-01T00:00:00Z"
		result, err = tool.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output = result.(*DateTimeCompareOutput)
		if !output.InRange {
			t.Error("Expected InRange to be true for date at range start")
		}
	})

	t.Run("sort dates", func(t *testing.T) {
		dates := []string{
			"2024-01-15T10:30:00Z",
			"2024-01-10T10:30:00Z",
			"2024-01-20T10:30:00Z",
			"2024-01-05T10:30:00Z",
		}

		// Ascending order
		input := DateTimeCompareInput{
			Operation: "sort",
			Dates:     dates,
			SortOrder: "asc",
		}
		result, err := tool.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeCompareOutput)
		expectedAsc := []string{
			"2024-01-05T10:30:00Z",
			"2024-01-10T10:30:00Z",
			"2024-01-15T10:30:00Z",
			"2024-01-20T10:30:00Z",
		}
		for i, expected := range expectedAsc {
			if output.SortedDates[i] != expected {
				t.Errorf("Expected %s at position %d, got %s", expected, i, output.SortedDates[i])
			}
		}

		// Descending order
		input.SortOrder = "desc"
		result, err = tool.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output = result.(*DateTimeCompareOutput)
		expectedDesc := []string{
			"2024-01-20T10:30:00Z",
			"2024-01-15T10:30:00Z",
			"2024-01-10T10:30:00Z",
			"2024-01-05T10:30:00Z",
		}
		for i, expected := range expectedDesc {
			if output.SortedDates[i] != expected {
				t.Errorf("Expected %s at position %d, got %s", expected, i, output.SortedDates[i])
			}
		}

		// Default order (should be ascending)
		input.SortOrder = ""
		result, err = tool.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output = result.(*DateTimeCompareOutput)
		for i, expected := range expectedAsc {
			if output.SortedDates[i] != expected {
				t.Errorf("Expected %s at position %d, got %s", expected, i, output.SortedDates[i])
			}
		}
	})

	t.Run("find extreme dates", func(t *testing.T) {
		dates := []string{
			"2024-01-15T10:30:00Z",
			"2024-01-10T10:30:00Z",
			"2024-01-20T10:30:00Z",
			"2024-01-05T10:30:00Z",
		}

		// Find earliest
		input := DateTimeCompareInput{
			Operation:   "find_extreme",
			Dates:       dates,
			ExtremeType: "earliest",
		}
		result, err := tool.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeCompareOutput)
		if output.ExtremeDate != "2024-01-05T10:30:00Z" {
			t.Errorf("Expected earliest date to be 2024-01-05T10:30:00Z, got %s", output.ExtremeDate)
		}

		// Find latest
		input.ExtremeType = "latest"
		result, err = tool.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output = result.(*DateTimeCompareOutput)
		if output.ExtremeDate != "2024-01-20T10:30:00Z" {
			t.Errorf("Expected latest date to be 2024-01-20T10:30:00Z, got %s", output.ExtremeDate)
		}

		// Default to earliest
		input.ExtremeType = ""
		result, err = tool.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output = result.(*DateTimeCompareOutput)
		if output.ExtremeDate != "2024-01-05T10:30:00Z" {
			t.Errorf("Expected default extreme to be earliest (2024-01-05T10:30:00Z), got %s", output.ExtremeDate)
		}
	})

	t.Run("error cases", func(t *testing.T) {
		// Missing dates for compare
		input := DateTimeCompareInput{
			Operation: "compare",
			Date1:     "2024-01-15T10:30:00Z",
			// Date2 missing
		}
		_, err := tool.Execute(ctx, input)
		if err == nil {
			t.Error("Expected error for missing date2")
		}

		// Missing period type for same_period
		input = DateTimeCompareInput{
			Operation: "same_period",
			Date1:     "2024-01-15T10:30:00Z",
			Date2:     "2024-01-16T10:30:00Z",
			// PeriodType missing
		}
		_, err = tool.Execute(ctx, input)
		if err == nil {
			t.Error("Expected error for missing period_type")
		}

		// Missing dates for range_check
		input = DateTimeCompareInput{
			Operation:  "range_check",
			Date1:      "2024-01-15T10:30:00Z",
			RangeStart: "2024-01-01T00:00:00Z",
			// RangeEnd missing
		}
		_, err = tool.Execute(ctx, input)
		if err == nil {
			t.Error("Expected error for missing range_end")
		}

		// Empty dates array for sort
		input = DateTimeCompareInput{
			Operation: "sort",
			Dates:     []string{},
		}
		_, err = tool.Execute(ctx, input)
		if err == nil {
			t.Error("Expected error for empty dates array")
		}

		// Invalid operation
		input = DateTimeCompareInput{
			Operation: "invalid",
		}
		_, err = tool.Execute(ctx, input)
		if err == nil {
			t.Error("Expected error for invalid operation")
		}

		// Invalid date format
		input = DateTimeCompareInput{
			Operation: "compare",
			Date1:     "invalid-date",
			Date2:     "2024-01-15T10:30:00Z",
		}
		_, err = tool.Execute(ctx, input)
		if err == nil {
			t.Error("Expected error for invalid date format")
		}

		// Invalid timezone
		input = DateTimeCompareInput{
			Operation: "compare",
			Date1:     "2024-01-15T10:30:00Z",
			Date2:     "2024-01-16T10:30:00Z",
			Timezone:  "Invalid/Timezone",
		}
		_, err = tool.Execute(ctx, input)
		if err == nil {
			t.Error("Expected error for invalid timezone")
		}
	})

	t.Run("various date formats", func(t *testing.T) {
		// Test that various date formats work
		formats := []struct {
			date1 string
			date2 string
		}{
			{"2024-01-15", "2024-01-16"},
			{"2024-01-15 10:30:00", "2024-01-16 10:30:00"},
			{"01/15/2024", "01/16/2024"},
			{"15-Jan-2024", "16-Jan-2024"},
		}

		for _, f := range formats {
			t.Run(fmt.Sprintf("%s vs %s", f.date1, f.date2), func(t *testing.T) {
				input := DateTimeCompareInput{
					Operation: "compare",
					Date1:     f.date1,
					Date2:     f.date2,
				}
				result, err := tool.Execute(ctx, input)
				if err != nil {
					t.Fatalf("Failed to parse dates: %v", err)
				}

				output := result.(*DateTimeCompareOutput)
				if !output.Before {
					t.Error("Expected date1 to be before date2")
				}
			})
		}
	})
}

func TestCalculateTimeDifference(t *testing.T) {
	testCases := []struct {
		name          string
		time1         time.Time
		time2         time.Time
		expectedDays  int
		expectedHours int
		expectedHuman string
	}{
		{
			name:          "1 day difference",
			time1:         time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			time2:         time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
			expectedDays:  1,
			expectedHours: 0,
			expectedHuman: "in 1 day",
		},
		{
			name:          "1 day ago",
			time1:         time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
			time2:         time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			expectedDays:  1,
			expectedHours: 0,
			expectedHuman: "1 day ago",
		},
		{
			name:          "hours and minutes",
			time1:         time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			time2:         time.Date(2024, 1, 15, 13, 45, 0, 0, time.UTC),
			expectedDays:  0,
			expectedHours: 3,
			expectedHuman: "in 3 hours and 15 minutes",
		},
		{
			name:          "same time",
			time1:         time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			time2:         time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			expectedDays:  0,
			expectedHours: 0,
			expectedHuman: "same time",
		},
		{
			name:          "complex difference",
			time1:         time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			time2:         time.Date(2024, 1, 17, 14, 45, 30, 0, time.UTC),
			expectedDays:  2,
			expectedHours: 4,
			expectedHuman: "in 2 days, 4 hours, 15 minutes and 30 seconds",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			diff := calculateTimeDifference(tc.time1, tc.time2)

			if diff.Days != tc.expectedDays {
				t.Errorf("Expected %d days, got %d", tc.expectedDays, diff.Days)
			}
			if diff.Hours != tc.expectedHours {
				t.Errorf("Expected %d hours, got %d", tc.expectedHours, diff.Hours)
			}
			if diff.HumanReadable != tc.expectedHuman {
				t.Errorf("Expected '%s', got '%s'", tc.expectedHuman, diff.HumanReadable)
			}
		})
	}
}

func TestAreSamePeriod(t *testing.T) {
	// Monday, January 15, 2024
	time1 := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	testCases := []struct {
		name       string
		time2      time.Time
		periodType string
		expected   bool
	}{
		{
			name:       "same day - true",
			time2:      time.Date(2024, 1, 15, 20, 45, 0, 0, time.UTC),
			periodType: "day",
			expected:   true,
		},
		{
			name:       "different day - false",
			time2:      time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
			periodType: "day",
			expected:   false,
		},
		{
			name:       "same week - true",
			time2:      time.Date(2024, 1, 19, 10, 30, 0, 0, time.UTC), // Friday
			periodType: "week",
			expected:   true,
		},
		{
			name:       "different week - false",
			time2:      time.Date(2024, 1, 22, 10, 30, 0, 0, time.UTC), // Next Monday
			periodType: "week",
			expected:   false,
		},
		{
			name:       "same month - true",
			time2:      time.Date(2024, 1, 31, 10, 30, 0, 0, time.UTC),
			periodType: "month",
			expected:   true,
		},
		{
			name:       "different month - false",
			time2:      time.Date(2024, 2, 1, 10, 30, 0, 0, time.UTC),
			periodType: "month",
			expected:   false,
		},
		{
			name:       "same year - true",
			time2:      time.Date(2024, 12, 31, 10, 30, 0, 0, time.UTC),
			periodType: "year",
			expected:   true,
		},
		{
			name:       "different year - false",
			time2:      time.Date(2025, 1, 1, 10, 30, 0, 0, time.UTC),
			periodType: "year",
			expected:   false,
		},
		{
			name:       "invalid period type",
			time2:      time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			periodType: "invalid",
			expected:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := areSamePeriod(time1, tc.time2, tc.periodType)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestParseDateInLocation(t *testing.T) {
	nyLoc, _ := time.LoadLocation("America/New_York")

	testCases := []struct {
		name     string
		dateStr  string
		loc      *time.Location
		hasError bool
	}{
		{
			name:     "RFC3339 date",
			dateStr:  "2024-01-15T10:30:00Z",
			loc:      time.UTC,
			hasError: false,
		},
		{
			name:     "simple date",
			dateStr:  "2024-01-15",
			loc:      nyLoc,
			hasError: false,
		},
		{
			name:     "date with time",
			dateStr:  "2024-01-15 10:30:00",
			loc:      nyLoc,
			hasError: false,
		},
		{
			name:     "invalid date",
			dateStr:  "invalid-date",
			loc:      time.UTC,
			hasError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseDateInLocation(tc.dateStr, tc.loc)
			if tc.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result.Location() != tc.loc {
					t.Errorf("Expected location %v, got %v", tc.loc, result.Location())
				}
			}
		})
	}
}
