package datetime

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestDateTimeNow(t *testing.T) {
	ctx := context.Background()
	tool := DateTimeNow()

	t.Run("basic functionality", func(t *testing.T) {
		input := DateTimeNowInput{}
		result, err := tool.Execute(createTestToolContext(ctx), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output, ok := result.(*DateTimeNowOutput)
		if !ok {
			t.Fatalf("Expected *DateTimeNowOutput, got %T", result)
		}

		// Check that UTC and Local times are set
		if output.UTC == "" {
			t.Error("UTC time should not be empty")
		}
		if output.Local == "" {
			t.Error("Local time should not be empty")
		}

		// Verify they can be parsed
		_, err = time.Parse(time.RFC3339, output.UTC)
		if err != nil {
			t.Errorf("Failed to parse UTC time: %v", err)
		}
		_, err = time.Parse(time.RFC3339, output.Local)
		if err != nil {
			t.Errorf("Failed to parse Local time: %v", err)
		}
	})

	t.Run("with specific timezone", func(t *testing.T) {
		input := DateTimeNowInput{
			Timezone: "America/New_York",
		}
		result, err := tool.Execute(createTestToolContext(ctx), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output, ok := result.(*DateTimeNowOutput)
		if !ok {
			t.Fatalf("Expected *DateTimeNowOutput, got %T", result)
		}

		if output.Timezone == "" {
			t.Error("Timezone time should not be empty")
		}
		if output.TimezoneName != "America/New_York" {
			t.Errorf("Expected timezone name 'America/New_York', got %s", output.TimezoneName)
		}
	})

	t.Run("invalid timezone", func(t *testing.T) {
		input := DateTimeNowInput{
			Timezone: "Invalid/Timezone",
		}
		_, err := tool.Execute(createTestToolContext(ctx), input)
		if err == nil {
			t.Error("Expected error for invalid timezone")
		}
	})

	t.Run("with components", func(t *testing.T) {
		input := DateTimeNowInput{
			IncludeComponents: true,
		}
		result, err := tool.Execute(createTestToolContext(ctx), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output, ok := result.(*DateTimeNowOutput)
		if !ok {
			t.Fatalf("Expected *DateTimeNowOutput, got %T", result)
		}

		if output.Components == nil {
			t.Fatal("Components should not be nil")
		}

		now := time.Now()
		if output.Components.Year != now.Year() {
			t.Errorf("Expected year %d, got %d", now.Year(), output.Components.Year)
		}
		if output.Components.Month < 1 || output.Components.Month > 12 {
			t.Errorf("Invalid month: %d", output.Components.Month)
		}
		if output.Components.MonthName == "" {
			t.Error("Month name should not be empty")
		}
		if output.Components.WeekdayName == "" {
			t.Error("Weekday name should not be empty")
		}
	})

	t.Run("with week info", func(t *testing.T) {
		input := DateTimeNowInput{
			IncludeWeekInfo: true,
		}
		result, err := tool.Execute(createTestToolContext(ctx), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output, ok := result.(*DateTimeNowOutput)
		if !ok {
			t.Fatalf("Expected *DateTimeNowOutput, got %T", result)
		}

		if output.WeekInfo == nil {
			t.Fatal("WeekInfo should not be nil")
		}

		if output.WeekInfo.WeekNumber < 1 || output.WeekInfo.WeekNumber > 53 {
			t.Errorf("Invalid week number: %d", output.WeekInfo.WeekNumber)
		}
		if output.WeekInfo.DayOfWeek < 1 || output.WeekInfo.DayOfWeek > 7 {
			t.Errorf("Invalid day of week: %d", output.WeekInfo.DayOfWeek)
		}
		if output.WeekInfo.Quarter < 1 || output.WeekInfo.Quarter > 4 {
			t.Errorf("Invalid quarter: %d", output.WeekInfo.Quarter)
		}
		if output.WeekInfo.DayOfYear < 1 || output.WeekInfo.DayOfYear > 366 {
			t.Errorf("Invalid day of year: %d", output.WeekInfo.DayOfYear)
		}
	})

	t.Run("with timestamps", func(t *testing.T) {
		input := DateTimeNowInput{
			IncludeTimestamps: true,
		}
		result, err := tool.Execute(createTestToolContext(ctx), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output, ok := result.(*DateTimeNowOutput)
		if !ok {
			t.Fatalf("Expected *DateTimeNowOutput, got %T", result)
		}

		if output.Timestamps == nil {
			t.Fatal("Timestamps should not be nil")
		}

		// Verify timestamps are reasonable (within last minute)
		now := time.Now().Unix()
		if output.Timestamps.Unix < now-60 || output.Timestamps.Unix > now+1 {
			t.Errorf("Unix timestamp seems incorrect: %d", output.Timestamps.Unix)
		}
		if output.Timestamps.UnixMilli < output.Timestamps.Unix*1000 {
			t.Error("UnixMilli should be >= Unix * 1000")
		}
		if output.Timestamps.UnixMicro < output.Timestamps.UnixMilli*1000 {
			t.Error("UnixMicro should be >= UnixMilli * 1000")
		}
		if output.Timestamps.UnixNano < output.Timestamps.UnixMicro*1000 {
			t.Error("UnixNano should be >= UnixMicro * 1000")
		}
	})

	t.Run("with custom format", func(t *testing.T) {
		input := DateTimeNowInput{
			Format: "2006-01-02 15:04:05",
		}
		result, err := tool.Execute(createTestToolContext(ctx), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output, ok := result.(*DateTimeNowOutput)
		if !ok {
			t.Fatalf("Expected *DateTimeNowOutput, got %T", result)
		}

		if output.Formatted == "" {
			t.Error("Formatted output should not be empty")
		}

		// Try to parse it back
		_, err = time.Parse("2006-01-02 15:04:05", output.Formatted)
		if err != nil {
			t.Errorf("Failed to parse formatted output: %v", err)
		}
	})

	t.Run("all options", func(t *testing.T) {
		input := DateTimeNowInput{
			Timezone:          "Europe/London",
			IncludeComponents: true,
			IncludeWeekInfo:   true,
			IncludeTimestamps: true,
			Format:            time.Kitchen,
		}
		result, err := tool.Execute(createTestToolContext(ctx), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output, ok := result.(*DateTimeNowOutput)
		if !ok {
			t.Fatalf("Expected *DateTimeNowOutput, got %T", result)
		}

		// Verify all fields are populated
		if output.UTC == "" || output.Local == "" || output.Timezone == "" {
			t.Error("Basic time fields should not be empty")
		}
		if output.TimezoneName != "Europe/London" {
			t.Errorf("Expected timezone name 'Europe/London', got %s", output.TimezoneName)
		}
		if output.Formatted == "" {
			t.Error("Formatted output should not be empty")
		}
		if output.Components == nil {
			t.Error("Components should not be nil")
		}
		if output.WeekInfo == nil {
			t.Error("WeekInfo should not be nil")
		}
		if output.Timestamps == nil {
			t.Error("Timestamps should not be nil")
		}
	})
}

func TestIsLeapYear(t *testing.T) {
	tests := []struct {
		year     int
		expected bool
	}{
		{2000, true},  // Divisible by 400
		{2004, true},  // Divisible by 4, not by 100
		{1900, false}, // Divisible by 100, not by 400
		{2001, false}, // Not divisible by 4
		{2020, true},  // Divisible by 4, not by 100
		{2024, true},  // Divisible by 4, not by 100
		{2100, false}, // Divisible by 100, not by 400
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("year_%d", tt.year), func(t *testing.T) {
			result := isLeapYear(tt.year)
			if result != tt.expected {
				t.Errorf("isLeapYear(%d) = %v, want %v", tt.year, result, tt.expected)
			}
		})
	}
}
