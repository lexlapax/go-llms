package datetime

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestDateTimeCalculate(t *testing.T) {
	ctx := context.Background()
	tool := DateTimeCalculate()

	t.Run("add operations", func(t *testing.T) {
		testCases := []struct {
			name      string
			startDate string
			unit      string
			value     int
			expected  string
		}{
			{"add days", "2024-01-15T10:30:00Z", "days", 5, "2024-01-20T10:30:00Z"},
			{"add months", "2024-01-15T10:30:00Z", "months", 2, "2024-03-15T10:30:00Z"},
			{"add years", "2024-01-15T10:30:00Z", "years", 1, "2025-01-15T10:30:00Z"},
			{"add hours", "2024-01-15T10:30:00Z", "hours", 3, "2024-01-15T13:30:00Z"},
			{"add minutes", "2024-01-15T10:30:00Z", "minutes", 45, "2024-01-15T11:15:00Z"},
			{"add seconds", "2024-01-15T10:30:00Z", "seconds", 30, "2024-01-15T10:30:30Z"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				input := DateTimeCalculateInput{
					Operation: "add",
					StartDate: tc.startDate,
					Unit:      tc.unit,
					Value:     tc.value,
				}
				result, err := tool.Execute(createTestToolContext(ctx), input)
				if err != nil {
					t.Fatalf("Failed to execute tool: %v", err)
				}

				output := result.(*DateTimeCalculateOutput)
				if output.Result != tc.expected {
					t.Errorf("Expected %s, got %s", tc.expected, output.Result)
				}
			})
		}
	})

	t.Run("subtract operations", func(t *testing.T) {
		testCases := []struct {
			name      string
			startDate string
			unit      string
			value     int
			expected  string
		}{
			{"subtract days", "2024-01-20T10:30:00Z", "days", 5, "2024-01-15T10:30:00Z"},
			{"subtract months", "2024-03-15T10:30:00Z", "months", 2, "2024-01-15T10:30:00Z"},
			{"subtract years", "2025-01-15T10:30:00Z", "years", 1, "2024-01-15T10:30:00Z"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				input := DateTimeCalculateInput{
					Operation: "subtract",
					StartDate: tc.startDate,
					Unit:      tc.unit,
					Value:     tc.value,
				}
				result, err := tool.Execute(createTestToolContext(ctx), input)
				if err != nil {
					t.Fatalf("Failed to execute tool: %v", err)
				}

				output := result.(*DateTimeCalculateOutput)
				if output.Result != tc.expected {
					t.Errorf("Expected %s, got %s", tc.expected, output.Result)
				}
			})
		}
	})

	t.Run("duration calculation", func(t *testing.T) {
		input := DateTimeCalculateInput{
			Operation: "duration",
			StartDate: "2024-01-15T10:30:00Z",
			EndDate:   "2024-01-20T15:45:30Z",
		}
		result, err := tool.Execute(createTestToolContext(ctx), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeCalculateOutput)
		if output.Duration == nil {
			t.Fatal("Expected duration output")
		}

		// Check duration components
		if output.Duration.Days != 5 {
			t.Errorf("Expected 5 days, got %d", output.Duration.Days)
		}
		if output.Duration.Hours != 5 {
			t.Errorf("Expected 5 hours, got %d", output.Duration.Hours)
		}
		if output.Duration.Minutes != 15 {
			t.Errorf("Expected 15 minutes, got %d", output.Duration.Minutes)
		}
		if output.Duration.Seconds != 30 {
			t.Errorf("Expected 30 seconds, got %d", output.Duration.Seconds)
		}

		// Check human readable format
		if !strings.Contains(output.Duration.HumanReadable, "5 days") {
			t.Errorf("Human readable should contain '5 days': %s", output.Duration.HumanReadable)
		}
	})

	t.Run("age calculation", func(t *testing.T) {
		input := DateTimeCalculateInput{
			Operation: "age",
			StartDate: "2000-01-15T00:00:00Z",
			EndDate:   "2024-03-20T00:00:00Z",
		}
		result, err := tool.Execute(createTestToolContext(ctx), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeCalculateOutput)
		if output.Age == nil {
			t.Fatal("Expected age output")
		}

		if output.Age.Years != 24 {
			t.Errorf("Expected 24 years, got %d", output.Age.Years)
		}
		if output.Age.Months != 2 {
			t.Errorf("Expected 2 months, got %d", output.Age.Months)
		}
		if output.Age.Days != 5 {
			t.Errorf("Expected 5 days, got %d", output.Age.Days)
		}

		// Check human readable format
		if !strings.Contains(output.Age.HumanReadable, "24 years") {
			t.Errorf("Human readable should contain '24 years': %s", output.Age.HumanReadable)
		}
	})

	t.Run("age calculation without end date", func(t *testing.T) {
		input := DateTimeCalculateInput{
			Operation: "age",
			StartDate: "2000-01-15T00:00:00Z",
		}
		result, err := tool.Execute(createTestToolContext(ctx), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeCalculateOutput)
		if output.Age == nil {
			t.Fatal("Expected age output")
		}

		// Should calculate age from start date to now
		if output.Age.Years < 24 {
			t.Errorf("Expected at least 24 years, got %d", output.Age.Years)
		}
	})

	t.Run("next weekday", func(t *testing.T) {
		// Start from Monday Jan 15, 2024
		input := DateTimeCalculateInput{
			Operation:     "next_weekday",
			StartDate:     "2024-01-15T10:30:00Z",
			TargetWeekday: 5, // Friday
		}
		result, err := tool.Execute(createTestToolContext(ctx), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeCalculateOutput)
		// Next Friday should be Jan 19
		if !strings.HasPrefix(output.Result, "2024-01-19T10:30:00") {
			t.Errorf("Expected next Friday to be 2024-01-19, got %s", output.Result)
		}
	})

	t.Run("previous weekday", func(t *testing.T) {
		// Start from Monday Jan 15, 2024
		input := DateTimeCalculateInput{
			Operation:     "previous_weekday",
			StartDate:     "2024-01-15T10:30:00Z",
			TargetWeekday: 5, // Friday
		}
		result, err := tool.Execute(createTestToolContext(ctx), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeCalculateOutput)
		// Previous Friday should be Jan 12
		if !strings.HasPrefix(output.Result, "2024-01-12T10:30:00") {
			t.Errorf("Expected previous Friday to be 2024-01-12, got %s", output.Result)
		}
	})

	t.Run("add business days", func(t *testing.T) {
		// Start from Monday Jan 15, 2024
		input := DateTimeCalculateInput{
			Operation: "add_business_days",
			StartDate: "2024-01-15T10:30:00Z",
			Value:     5,
		}
		result, err := tool.Execute(createTestToolContext(ctx), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeCalculateOutput)
		// 5 business days from Monday should be Monday Jan 22 (skipping weekend)
		if !strings.HasPrefix(output.Result, "2024-01-22T10:30:00") {
			t.Errorf("Expected 5 business days later to be 2024-01-22, got %s", output.Result)
		}
		if output.BusinessDays != 5 {
			t.Errorf("Expected 5 business days, got %d", output.BusinessDays)
		}
	})

	t.Run("subtract business days", func(t *testing.T) {
		// Start from Monday Jan 22, 2024
		input := DateTimeCalculateInput{
			Operation: "subtract_business_days",
			StartDate: "2024-01-22T10:30:00Z",
			Value:     5,
		}
		result, err := tool.Execute(createTestToolContext(ctx), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeCalculateOutput)
		// 5 business days before Monday should be Monday Jan 15 (skipping weekend)
		if !strings.HasPrefix(output.Result, "2024-01-15T10:30:00") {
			t.Errorf("Expected 5 business days earlier to be 2024-01-15, got %s", output.Result)
		}
		if output.BusinessDays != -5 {
			t.Errorf("Expected -5 business days, got %d", output.BusinessDays)
		}
	})

	t.Run("business days with weekends", func(t *testing.T) {
		input := DateTimeCalculateInput{
			Operation:       "add_business_days",
			StartDate:       "2024-01-15T10:30:00Z",
			Value:           5,
			IncludeWeekends: true,
		}
		result, err := tool.Execute(createTestToolContext(ctx), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeCalculateOutput)
		// With weekends included, 5 days from Monday is Saturday Jan 20
		if !strings.HasPrefix(output.Result, "2024-01-20T10:30:00") {
			t.Errorf("Expected 5 days later (with weekends) to be 2024-01-20, got %s", output.Result)
		}
	})

	t.Run("with timezone", func(t *testing.T) {
		input := DateTimeCalculateInput{
			Operation: "add",
			StartDate: "2024-01-15T10:30:00Z",
			Unit:      "hours",
			Value:     5,
			Timezone:  "America/New_York",
		}
		result, err := tool.Execute(createTestToolContext(ctx), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeCalculateOutput)
		// Result should be in Eastern timezone
		if !strings.Contains(output.Result, "-05:00") && !strings.Contains(output.Result, "-04:00") {
			t.Errorf("Expected result to be in Eastern timezone, got %s", output.Result)
		}
	})

	t.Run("month end handling", func(t *testing.T) {
		// Test adding months when start date is end of month
		input := DateTimeCalculateInput{
			Operation: "add",
			StartDate: "2024-01-31T00:00:00Z",
			Unit:      "months",
			Value:     1,
		}
		result, err := tool.Execute(createTestToolContext(ctx), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeCalculateOutput)
		// Should handle February correctly (Feb doesn't have 31 days)
		// Go's AddDate handles this by going to March
		if !strings.Contains(output.Result, "-03-") {
			t.Errorf("Expected result to be in March (due to Feb overflow), got %s", output.Result)
		}
	})

	t.Run("invalid operation", func(t *testing.T) {
		input := DateTimeCalculateInput{
			Operation: "invalid",
			StartDate: "2024-01-15T10:30:00Z",
		}
		_, err := tool.Execute(createTestToolContext(ctx), input)
		if err == nil {
			t.Error("Expected error for invalid operation")
		}
	})

	t.Run("missing required fields", func(t *testing.T) {
		// Missing unit for add operation
		input := DateTimeCalculateInput{
			Operation: "add",
			StartDate: "2024-01-15T10:30:00Z",
			Value:     5,
		}
		_, err := tool.Execute(createTestToolContext(ctx), input)
		if err == nil {
			t.Error("Expected error for missing unit")
		}

		// Missing end_date for duration
		input = DateTimeCalculateInput{
			Operation: "duration",
			StartDate: "2024-01-15T10:30:00Z",
		}
		_, err = tool.Execute(createTestToolContext(ctx), input)
		if err == nil {
			t.Error("Expected error for missing end_date")
		}
	})

	t.Run("alternative date formats", func(t *testing.T) {
		formats := []string{
			"2024-01-15",
			"2024-01-15 10:30:00",
			"01/15/2024",
			"15-Jan-2024",
			"Jan 15, 2024",
			"January 15, 2024",
		}

		for _, format := range formats {
			t.Run(format, func(t *testing.T) {
				input := DateTimeCalculateInput{
					Operation: "add",
					StartDate: format,
					Unit:      "days",
					Value:     1,
				}
				result, err := tool.Execute(createTestToolContext(ctx), input)
				if err != nil {
					t.Fatalf("Failed to parse date format %s: %v", format, err)
				}

				output := result.(*DateTimeCalculateOutput)
				if output.Result == "" {
					t.Error("Expected non-empty result")
				}
			})
		}
	})
}

func TestCalculateDuration(t *testing.T) {
	start := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	end := time.Date(2024, 1, 20, 15, 45, 30, 500000000, time.UTC)

	duration := calculateDuration(start, end)

	if duration.Days != 5 {
		t.Errorf("Expected 5 days, got %d", duration.Days)
	}
	if duration.Hours != 5 {
		t.Errorf("Expected 5 hours, got %d", duration.Hours)
	}
	if duration.Minutes != 15 {
		t.Errorf("Expected 15 minutes, got %d", duration.Minutes)
	}
	if duration.Seconds != 30 {
		t.Errorf("Expected 30 seconds, got %d", duration.Seconds)
	}
	if duration.Milliseconds != 500 {
		t.Errorf("Expected 500 milliseconds, got %d", duration.Milliseconds)
	}

	// Test negative duration
	negativeDuration := calculateDuration(end, start)
	if negativeDuration.TotalSeconds >= 0 {
		t.Error("Expected negative total seconds for reverse duration")
	}
}

func TestCalculateAge(t *testing.T) {
	testCases := []struct {
		name           string
		birthDate      string
		currentDate    string
		expectedYears  int
		expectedMonths int
		expectedDays   int
	}{
		{
			name:           "simple age",
			birthDate:      "2000-01-15",
			currentDate:    "2024-03-20",
			expectedYears:  24,
			expectedMonths: 2,
			expectedDays:   5,
		},
		{
			name:           "same month",
			birthDate:      "2000-01-15",
			currentDate:    "2024-01-20",
			expectedYears:  24,
			expectedMonths: 0,
			expectedDays:   5,
		},
		{
			name:           "day before birthday",
			birthDate:      "2000-01-15",
			currentDate:    "2024-01-14",
			expectedYears:  23,
			expectedMonths: 11,
			expectedDays:   30, // December has 31 days, so 30 days from Dec 15 to Jan 14
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			birthDate, _ := time.Parse("2006-01-02", tc.birthDate)
			currentDate, _ := time.Parse("2006-01-02", tc.currentDate)

			age := calculateAge(birthDate, currentDate)

			if age.Years != tc.expectedYears {
				t.Errorf("Expected %d years, got %d", tc.expectedYears, age.Years)
			}
			if age.Months != tc.expectedMonths {
				t.Errorf("Expected %d months, got %d", tc.expectedMonths, age.Months)
			}
			if age.Days != tc.expectedDays {
				t.Errorf("Expected %d days, got %d", tc.expectedDays, age.Days)
			}
		})
	}
}

func TestNextPreviousWeekday(t *testing.T) {
	// Test date: Monday, Jan 15, 2024
	monday := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	t.Run("next weekday", func(t *testing.T) {
		// Next Friday from Monday should be Jan 19
		nextFri := nextWeekday(monday, time.Friday)
		if nextFri.Day() != 19 {
			t.Errorf("Expected next Friday to be Jan 19, got %d", nextFri.Day())
		}

		// Next Monday from Monday should be next week (Jan 22)
		nextMon := nextWeekday(monday, time.Monday)
		if nextMon.Day() != 22 {
			t.Errorf("Expected next Monday to be Jan 22, got %d", nextMon.Day())
		}
	})

	t.Run("previous weekday", func(t *testing.T) {
		// Previous Friday from Monday should be Jan 12
		prevFri := previousWeekday(monday, time.Friday)
		if prevFri.Day() != 12 {
			t.Errorf("Expected previous Friday to be Jan 12, got %d", prevFri.Day())
		}

		// Previous Monday from Monday should be previous week (Jan 8)
		prevMon := previousWeekday(monday, time.Monday)
		if prevMon.Day() != 8 {
			t.Errorf("Expected previous Monday to be Jan 8, got %d", prevMon.Day())
		}
	})
}

func TestAddBusinessDays(t *testing.T) {
	// Test date: Monday, Jan 15, 2024
	monday := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	t.Run("add business days", func(t *testing.T) {
		// Add 5 business days should skip weekend
		result, count := addBusinessDays(monday, 5, false)
		if result.Day() != 22 { // Next Monday
			t.Errorf("Expected Jan 22, got %d", result.Day())
		}
		if count != 5 {
			t.Errorf("Expected count 5, got %d", count)
		}
	})

	t.Run("subtract business days", func(t *testing.T) {
		// Subtract 5 business days should skip weekend
		result, count := addBusinessDays(monday, -5, false)
		if result.Day() != 8 { // Previous Monday
			t.Errorf("Expected Jan 8, got %d", result.Day())
		}
		if count != -5 {
			t.Errorf("Expected count -5, got %d", count)
		}
	})

	t.Run("with weekends included", func(t *testing.T) {
		result, count := addBusinessDays(monday, 5, true)
		if result.Day() != 20 { // Saturday
			t.Errorf("Expected Jan 20, got %d", result.Day())
		}
		if count != 5 {
			t.Errorf("Expected count 5, got %d", count)
		}
	})
}
