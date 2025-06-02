package datetime

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestDateTimeInfo(t *testing.T) {
	ctx := context.Background()
	tool := DateTimeInfo()

	t.Run("basic date info", func(t *testing.T) {
		// Use a specific date for consistent testing
		input := DateTimeInfoInput{
			Date: "2024-01-15T10:30:00Z",
		}
		result, err := tool.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output, ok := result.(*DateTimeInfoOutput)
		if !ok {
			t.Fatalf("Expected *DateTimeInfoOutput, got %T", result)
		}

		// Verify basic fields
		if output.Year != 2024 {
			t.Errorf("Expected year 2024, got %d", output.Year)
		}
		if output.Month != 1 {
			t.Errorf("Expected month 1, got %d", output.Month)
		}
		if output.MonthName != "January" {
			t.Errorf("Expected month name January, got %s", output.MonthName)
		}
		if output.DayOfMonth != 15 {
			t.Errorf("Expected day 15, got %d", output.DayOfMonth)
		}
		if output.DayOfWeekName != "Monday" {
			t.Errorf("Expected Monday, got %s", output.DayOfWeekName)
		}
		if output.DayOfWeek != 1 { // Monday = 1
			t.Errorf("Expected day of week 1, got %d", output.DayOfWeek)
		}
		if output.DayOfWeekISO != 1 { // Monday = 1 in ISO
			t.Errorf("Expected ISO day of week 1, got %d", output.DayOfWeekISO)
		}
	})

	t.Run("leap year check", func(t *testing.T) {
		testCases := []struct {
			date       string
			isLeapYear bool
		}{
			{"2024-01-01T00:00:00Z", true},  // 2024 is a leap year
			{"2023-01-01T00:00:00Z", false}, // 2023 is not a leap year
			{"2000-01-01T00:00:00Z", true},  // 2000 is a leap year (divisible by 400)
			{"1900-01-01T00:00:00Z", false}, // 1900 is not a leap year (divisible by 100 but not 400)
		}

		for _, tc := range testCases {
			t.Run(tc.date[:4], func(t *testing.T) {
				input := DateTimeInfoInput{Date: tc.date}
				result, err := tool.Execute(ctx, input)
				if err != nil {
					t.Fatalf("Failed to execute tool: %v", err)
				}

				output := result.(*DateTimeInfoOutput)
				if output.IsLeapYear != tc.isLeapYear {
					t.Errorf("Expected IsLeapYear=%v for %s, got %v", tc.isLeapYear, tc.date, output.IsLeapYear)
				}
			})
		}
	})

	t.Run("days in month", func(t *testing.T) {
		testCases := []struct {
			date string
			days int
		}{
			{"2024-01-15T00:00:00Z", 31}, // January
			{"2024-02-15T00:00:00Z", 29}, // February in leap year
			{"2023-02-15T00:00:00Z", 28}, // February in non-leap year
			{"2024-04-15T00:00:00Z", 30}, // April
		}

		for _, tc := range testCases {
			t.Run(tc.date[:7], func(t *testing.T) {
				input := DateTimeInfoInput{Date: tc.date}
				result, err := tool.Execute(ctx, input)
				if err != nil {
					t.Fatalf("Failed to execute tool: %v", err)
				}

				output := result.(*DateTimeInfoOutput)
				if output.DaysInMonth != tc.days {
					t.Errorf("Expected %d days in month for %s, got %d", tc.days, tc.date, output.DaysInMonth)
				}
			})
		}
	})

	t.Run("quarter calculation", func(t *testing.T) {
		testCases := []struct {
			date    string
			quarter int
		}{
			{"2024-01-15T00:00:00Z", 1},
			{"2024-04-15T00:00:00Z", 2},
			{"2024-07-15T00:00:00Z", 3},
			{"2024-10-15T00:00:00Z", 4},
		}

		for _, tc := range testCases {
			t.Run(tc.date[:7], func(t *testing.T) {
				input := DateTimeInfoInput{Date: tc.date}
				result, err := tool.Execute(ctx, input)
				if err != nil {
					t.Fatalf("Failed to execute tool: %v", err)
				}

				output := result.(*DateTimeInfoOutput)
				if output.Quarter != tc.quarter {
					t.Errorf("Expected quarter %d for %s, got %d", tc.quarter, tc.date, output.Quarter)
				}
			})
		}
	})

	t.Run("week boundaries with Sunday start", func(t *testing.T) {
		input := DateTimeInfoInput{
			Date:         "2024-01-15T10:30:00Z", // Monday
			WeekStartDay: 0,                      // Sunday
		}
		result, err := tool.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeInfoOutput)

		// Week should start on Sunday Jan 14
		if !strings.HasPrefix(output.StartOfWeek, "2024-01-14T00:00:00") {
			t.Errorf("Expected week to start on 2024-01-14, got %s", output.StartOfWeek)
		}
		// Week should end on Saturday Jan 20
		if !strings.HasPrefix(output.EndOfWeek, "2024-01-20T23:59:59") {
			t.Errorf("Expected week to end on 2024-01-20, got %s", output.EndOfWeek)
		}
	})

	t.Run("week boundaries with Monday start", func(t *testing.T) {
		input := DateTimeInfoInput{
			Date:         "2024-01-15T10:30:00Z", // Monday
			WeekStartDay: 1,                      // Monday
		}
		result, err := tool.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeInfoOutput)

		// Week should start on Monday Jan 15
		if !strings.HasPrefix(output.StartOfWeek, "2024-01-15T00:00:00") {
			t.Errorf("Expected week to start on 2024-01-15, got %s", output.StartOfWeek)
		}
		// Week should end on Sunday Jan 21
		if !strings.HasPrefix(output.EndOfWeek, "2024-01-21T23:59:59") {
			t.Errorf("Expected week to end on 2024-01-21, got %s", output.EndOfWeek)
		}
	})

	t.Run("month boundaries", func(t *testing.T) {
		input := DateTimeInfoInput{
			Date: "2024-01-15T10:30:00Z",
		}
		result, err := tool.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeInfoOutput)

		if !strings.HasPrefix(output.StartOfMonth, "2024-01-01T00:00:00") {
			t.Errorf("Expected month to start on 2024-01-01, got %s", output.StartOfMonth)
		}
		if !strings.HasPrefix(output.EndOfMonth, "2024-01-31T23:59:59") {
			t.Errorf("Expected month to end on 2024-01-31, got %s", output.EndOfMonth)
		}
	})

	t.Run("quarter boundaries", func(t *testing.T) {
		input := DateTimeInfoInput{
			Date: "2024-02-15T10:30:00Z", // Q1
		}
		result, err := tool.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeInfoOutput)

		if !strings.HasPrefix(output.StartOfQuarter, "2024-01-01T00:00:00") {
			t.Errorf("Expected Q1 to start on 2024-01-01, got %s", output.StartOfQuarter)
		}
		if !strings.HasPrefix(output.EndOfQuarter, "2024-03-31T23:59:59") {
			t.Errorf("Expected Q1 to end on 2024-03-31, got %s", output.EndOfQuarter)
		}
	})

	t.Run("year boundaries", func(t *testing.T) {
		input := DateTimeInfoInput{
			Date: "2024-06-15T10:30:00Z",
		}
		result, err := tool.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeInfoOutput)

		if !strings.HasPrefix(output.StartOfYear, "2024-01-01T00:00:00") {
			t.Errorf("Expected year to start on 2024-01-01, got %s", output.StartOfYear)
		}
		if !strings.HasPrefix(output.EndOfYear, "2024-12-31T23:59:59") {
			t.Errorf("Expected year to end on 2024-12-31, got %s", output.EndOfYear)
		}
	})

	t.Run("with timezone", func(t *testing.T) {
		input := DateTimeInfoInput{
			Date:     "2024-01-15T10:30:00Z",
			Timezone: "America/New_York",
		}
		result, err := tool.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeInfoOutput)

		// The output date should be in EST/EDT timezone
		if !strings.Contains(output.Date, "-05:00") && !strings.Contains(output.Date, "-04:00") {
			t.Errorf("Expected date to be in Eastern timezone, got %s", output.Date)
		}
	})

	t.Run("alternative date formats", func(t *testing.T) {
		formats := []string{
			"2024-01-15",
			"2024-01-15 10:30:00",
			"01/15/2024",
			"15-Jan-2024",
		}

		for _, format := range formats {
			t.Run(format, func(t *testing.T) {
				input := DateTimeInfoInput{Date: format}
				result, err := tool.Execute(ctx, input)
				if err != nil {
					t.Fatalf("Failed to parse date format %s: %v", format, err)
				}

				output := result.(*DateTimeInfoOutput)
				if output.Year != 2024 || output.Month != 1 || output.DayOfMonth != 15 {
					t.Errorf("Incorrectly parsed date %s", format)
				}
			})
		}
	})

	t.Run("invalid date", func(t *testing.T) {
		input := DateTimeInfoInput{
			Date: "invalid-date",
		}
		_, err := tool.Execute(ctx, input)
		if err == nil {
			t.Error("Expected error for invalid date")
		}
	})

	t.Run("sunday date with ISO week", func(t *testing.T) {
		input := DateTimeInfoInput{
			Date: "2024-01-07T10:30:00Z", // Sunday
		}
		result, err := tool.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeInfoOutput)

		if output.DayOfWeek != 0 { // Sunday = 0
			t.Errorf("Expected day of week 0 for Sunday, got %d", output.DayOfWeek)
		}
		if output.DayOfWeekISO != 7 { // Sunday = 7 in ISO
			t.Errorf("Expected ISO day of week 7 for Sunday, got %d", output.DayOfWeekISO)
		}
	})

	t.Run("week number and year", func(t *testing.T) {
		// Test a date that's in week 1 of the next year
		input := DateTimeInfoInput{
			Date: "2023-12-31T00:00:00Z", // This might be week 52/53 of 2023 or week 1 of 2024
		}
		result, err := tool.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeInfoOutput)

		// Verify week number is valid
		if output.WeekNumber < 1 || output.WeekNumber > 53 {
			t.Errorf("Invalid week number: %d", output.WeekNumber)
		}

		// Week year might be different from calendar year
		if output.WeekYear != 2023 && output.WeekYear != 2024 {
			t.Errorf("Unexpected week year: %d", output.WeekYear)
		}
	})
}

func TestDaysInMonth(t *testing.T) {
	testCases := []struct {
		date     string
		expected int
	}{
		{"2024-01-15", 31}, // January
		{"2024-02-15", 29}, // February in leap year
		{"2023-02-15", 28}, // February in non-leap year
		{"2024-03-15", 31}, // March
		{"2024-04-15", 30}, // April
		{"2024-05-15", 31}, // May
		{"2024-06-15", 30}, // June
		{"2024-07-15", 31}, // July
		{"2024-08-15", 31}, // August
		{"2024-09-15", 30}, // September
		{"2024-10-15", 31}, // October
		{"2024-11-15", 30}, // November
		{"2024-12-15", 31}, // December
	}

	for _, tc := range testCases {
		t.Run(tc.date, func(t *testing.T) {
			date, _ := time.Parse("2006-01-02", tc.date)
			days := daysInMonth(date)
			if days != tc.expected {
				t.Errorf("Expected %d days for %s, got %d", tc.expected, tc.date, days)
			}
		})
	}
}

func TestStartOfWeek(t *testing.T) {
	// Test date: Wednesday, Jan 17, 2024
	testDate, _ := time.Parse("2006-01-02", "2024-01-17")

	t.Run("Sunday start", func(t *testing.T) {
		start := startOfWeek(testDate, time.Sunday)
		expected, _ := time.Parse("2006-01-02", "2024-01-14")
		if !start.Equal(expected) {
			t.Errorf("Expected week to start on %s, got %s", expected.Format("2006-01-02"), start.Format("2006-01-02"))
		}
	})

	t.Run("Monday start", func(t *testing.T) {
		start := startOfWeek(testDate, time.Monday)
		expected, _ := time.Parse("2006-01-02", "2024-01-15")
		if !start.Equal(expected) {
			t.Errorf("Expected week to start on %s, got %s", expected.Format("2006-01-02"), start.Format("2006-01-02"))
		}
	})

	// Test edge case: when test date is the start day
	t.Run("date is start day", func(t *testing.T) {
		monday, _ := time.Parse("2006-01-02", "2024-01-15")
		start := startOfWeek(monday, time.Monday)
		if !start.Equal(monday) {
			t.Errorf("Expected week to start on same day, got %s", start.Format("2006-01-02"))
		}
	})
}
