package datetime

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestDateTimeConvert(t *testing.T) {
	ctx := context.Background()
	tool := DateTimeConvert()

	t.Run("timezone conversion", func(t *testing.T) {
		input := DateTimeConvertInput{
			Operation:    "timezone",
			DateTime:     "2024-01-15T10:30:00Z",
			FromTimezone: "UTC",
			ToTimezone:   "America/New_York",
		}
		result, err := tool.Execute(createTestToolContext(ctx), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeConvertOutput)
		if output.Converted == "" {
			t.Error("Expected converted datetime")
		}

		// Should be in EST (-05:00) for January
		if !strings.Contains(output.Converted, "-05:00") {
			t.Errorf("Expected EST timezone offset, got %s", output.Converted)
		}

		// Check timezone info
		if output.TimezoneInfo == nil {
			t.Error("Expected timezone info")
		} else {
			if output.TimezoneInfo.Name != "America/New_York" {
				t.Errorf("Expected timezone name America/New_York, got %s", output.TimezoneInfo.Name)
			}
			if output.TimezoneInfo.Offset != "-05:00" {
				t.Errorf("Expected offset -05:00, got %s", output.TimezoneInfo.Offset)
			}
		}
	})

	t.Run("timezone conversion with DST info", func(t *testing.T) {
		input := DateTimeConvertInput{
			Operation:  "timezone",
			DateTime:   "2024-07-15T10:30:00Z", // July - DST period
			ToTimezone: "America/New_York",
			IncludeDST: true,
		}
		result, err := tool.Execute(createTestToolContext(ctx), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeConvertOutput)
		if output.DSTInfo == nil {
			t.Fatal("Expected DST info")
		}

		// In July, New York should be in DST (-04:00)
		if !output.DSTInfo.IsDST {
			t.Error("Expected IsDST to be true for July in New York")
		}
		if output.DSTInfo.CurrentOffset != "-04:00" {
			t.Errorf("Expected current offset -04:00 for EDT, got %s", output.DSTInfo.CurrentOffset)
		}
	})

	t.Run("to timestamp", func(t *testing.T) {
		input := DateTimeConvertInput{
			Operation: "to_timestamp",
			DateTime:  "2024-01-15T10:30:00Z",
		}
		result, err := tool.Execute(createTestToolContext(ctx), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeConvertOutput)

		// Verify timestamp values
		expectedTimestamp := int64(1705314600) // 2024-01-15 10:30:00 UTC
		if output.Timestamp != expectedTimestamp {
			t.Errorf("Expected timestamp %d, got %d", expectedTimestamp, output.Timestamp)
		}
		if output.TimestampMillis != expectedTimestamp*1000 {
			t.Errorf("Expected timestamp millis %d, got %d", expectedTimestamp*1000, output.TimestampMillis)
		}
		if output.TimestampMicros != expectedTimestamp*1000000 {
			t.Errorf("Expected timestamp micros %d, got %d", expectedTimestamp*1000000, output.TimestampMicros)
		}
		if output.TimestampNanos != expectedTimestamp*1000000000 {
			t.Errorf("Expected timestamp nanos %d, got %d", expectedTimestamp*1000000000, output.TimestampNanos)
		}
	})

	t.Run("from timestamp seconds", func(t *testing.T) {
		input := DateTimeConvertInput{
			Operation:     "from_timestamp",
			Timestamp:     1705314600, // 2024-01-15 10:30:00 UTC
			TimestampUnit: "seconds",
		}
		result, err := tool.Execute(createTestToolContext(ctx), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeConvertOutput)
		if output.Converted != "2024-01-15T10:30:00Z" {
			t.Errorf("Expected 2024-01-15T10:30:00Z, got %s", output.Converted)
		}
	})

	t.Run("from timestamp milliseconds", func(t *testing.T) {
		input := DateTimeConvertInput{
			Operation:     "from_timestamp",
			Timestamp:     1705314600000, // 2024-01-15 10:30:00 UTC in milliseconds
			TimestampUnit: "milliseconds",
		}
		result, err := tool.Execute(createTestToolContext(ctx), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeConvertOutput)
		if output.Converted != "2024-01-15T10:30:00Z" {
			t.Errorf("Expected 2024-01-15T10:30:00Z, got %s", output.Converted)
		}
	})

	t.Run("from timestamp with timezone", func(t *testing.T) {
		input := DateTimeConvertInput{
			Operation:     "from_timestamp",
			Timestamp:     1705314600,
			TimestampUnit: "seconds",
			ToTimezone:    "America/Los_Angeles",
		}
		result, err := tool.Execute(createTestToolContext(ctx), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeConvertOutput)
		// Should be in PST (-08:00)
		if !strings.Contains(output.Converted, "-08:00") {
			t.Errorf("Expected PST timezone offset, got %s", output.Converted)
		}
	})

	t.Run("list timezones", func(t *testing.T) {
		input := DateTimeConvertInput{
			Operation: "list_timezones",
		}
		result, err := tool.Execute(createTestToolContext(ctx), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeConvertOutput)
		if len(output.Timezones) == 0 {
			t.Error("Expected timezone list")
		}

		// Check for some common timezones
		expectedZones := []string{"UTC", "America/New_York", "Europe/London", "Asia/Tokyo"}
		for _, expected := range expectedZones {
			found := false
			for _, tz := range output.Timezones {
				if tz == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected timezone %s in list", expected)
			}
		}
	})

	t.Run("list timezones with filter", func(t *testing.T) {
		input := DateTimeConvertInput{
			Operation:      "list_timezones",
			TimezoneFilter: "America",
		}
		result, err := tool.Execute(createTestToolContext(ctx), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeConvertOutput)
		if len(output.Timezones) == 0 {
			t.Error("Expected filtered timezone list")
		}

		// All results should contain "America"
		for _, tz := range output.Timezones {
			if !strings.Contains(tz, "America") {
				t.Errorf("Timezone %s doesn't match filter 'America'", tz)
			}
		}
	})

	t.Run("timezone conversion without from_timezone", func(t *testing.T) {
		input := DateTimeConvertInput{
			Operation:  "timezone",
			DateTime:   "2024-01-15T10:30:00Z",
			ToTimezone: "Asia/Tokyo",
		}
		result, err := tool.Execute(createTestToolContext(ctx), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeConvertOutput)
		// Should be in JST (+09:00)
		if !strings.Contains(output.Converted, "+09:00") {
			t.Errorf("Expected JST timezone offset, got %s", output.Converted)
		}
	})

	t.Run("error cases", func(t *testing.T) {
		// Missing datetime for timezone conversion
		input := DateTimeConvertInput{
			Operation:  "timezone",
			ToTimezone: "America/New_York",
		}
		_, err := tool.Execute(createTestToolContext(ctx), input)
		if err == nil {
			t.Error("Expected error for missing datetime")
		}

		// Missing to_timezone
		input = DateTimeConvertInput{
			Operation: "timezone",
			DateTime:  "2024-01-15T10:30:00Z",
		}
		_, err = tool.Execute(createTestToolContext(ctx), input)
		if err == nil {
			t.Error("Expected error for missing to_timezone")
		}

		// Invalid timezone
		input = DateTimeConvertInput{
			Operation:  "timezone",
			DateTime:   "2024-01-15T10:30:00Z",
			ToTimezone: "Invalid/Timezone",
		}
		_, err = tool.Execute(createTestToolContext(ctx), input)
		if err == nil {
			t.Error("Expected error for invalid timezone")
		}

		// Missing timestamp for from_timestamp
		input = DateTimeConvertInput{
			Operation: "from_timestamp",
		}
		_, err = tool.Execute(createTestToolContext(ctx), input)
		if err == nil {
			t.Error("Expected error for missing timestamp")
		}

		// Invalid operation
		input = DateTimeConvertInput{
			Operation: "invalid",
		}
		_, err = tool.Execute(createTestToolContext(ctx), input)
		if err == nil {
			t.Error("Expected error for invalid operation")
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
				input := DateTimeConvertInput{
					Operation:  "timezone",
					DateTime:   format,
					ToTimezone: "UTC",
				}
				result, err := tool.Execute(createTestToolContext(ctx), input)
				if err != nil {
					t.Fatalf("Failed to parse date format %s: %v", format, err)
				}

				output := result.(*DateTimeConvertOutput)
				if output.Converted == "" {
					t.Error("Expected converted datetime")
				}
			})
		}
	})

	t.Run("nanosecond timestamp", func(t *testing.T) {
		nanoTimestamp := int64(1705314600000000000) // 2024-01-15 10:30:00 UTC in nanoseconds
		input := DateTimeConvertInput{
			Operation:     "from_timestamp",
			Timestamp:     nanoTimestamp,
			TimestampUnit: "nanoseconds",
		}
		result, err := tool.Execute(createTestToolContext(ctx), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		output := result.(*DateTimeConvertOutput)
		if output.Converted != "2024-01-15T10:30:00Z" {
			t.Errorf("Expected 2024-01-15T10:30:00Z, got %s", output.Converted)
		}
	})
}

func TestGetTimezoneInfo(t *testing.T) {
	// Create a time in New York timezone
	loc, _ := time.LoadLocation("America/New_York")
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, loc)

	info := getTimezoneInfo(testTime)
	if info.Name != "America/New_York" {
		t.Errorf("Expected timezone name America/New_York, got %s", info.Name)
	}
	if info.Abbreviation != "EST" {
		t.Errorf("Expected abbreviation EST, got %s", info.Abbreviation)
	}
	if info.Offset != "-05:00" {
		t.Errorf("Expected offset -05:00, got %s", info.Offset)
	}
	if info.OffsetSeconds != -18000 {
		t.Errorf("Expected offset seconds -18000, got %d", info.OffsetSeconds)
	}
}

func TestGetDSTInfo(t *testing.T) {
	loc, _ := time.LoadLocation("America/New_York")

	t.Run("winter (no DST)", func(t *testing.T) {
		winterTime := time.Date(2024, 1, 15, 10, 30, 0, 0, loc)
		info := getDSTInfo(winterTime, loc)

		if info.IsDST {
			t.Error("Expected IsDST to be false in winter")
		}
		if info.CurrentOffset != "-05:00" {
			t.Errorf("Expected current offset -05:00, got %s", info.CurrentOffset)
		}
		if info.StandardOffset != "-05:00" {
			t.Errorf("Expected standard offset -05:00, got %s", info.StandardOffset)
		}
		if info.DSTOffset != "-04:00" {
			t.Errorf("Expected DST offset -04:00, got %s", info.DSTOffset)
		}
	})

	t.Run("summer (DST)", func(t *testing.T) {
		summerTime := time.Date(2024, 7, 15, 10, 30, 0, 0, loc)
		info := getDSTInfo(summerTime, loc)

		if !info.IsDST {
			t.Error("Expected IsDST to be true in summer")
		}
		if info.CurrentOffset != "-04:00" {
			t.Errorf("Expected current offset -04:00, got %s", info.CurrentOffset)
		}
	})
}

func TestFormatOffset(t *testing.T) {
	testCases := []struct {
		offsetSeconds int
		expected      string
	}{
		{0, "+00:00"},
		{3600, "+01:00"},
		{-3600, "-01:00"},
		{19800, "+05:30"},  // India
		{-18000, "-05:00"}, // EST
		{-14400, "-04:00"}, // EDT
		{32400, "+09:00"},  // Japan
		{-28800, "-08:00"}, // PST
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := formatOffset(tc.offsetSeconds)
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestGetFilteredTimezones(t *testing.T) {
	t.Run("no filter", func(t *testing.T) {
		zones := getFilteredTimezones("")
		if len(zones) == 0 {
			t.Error("Expected timezones with no filter")
		}

		// Should include common timezones
		found := false
		for _, tz := range zones {
			if tz == "UTC" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected UTC in unfiltered list")
		}
	})

	t.Run("filter by continent", func(t *testing.T) {
		zones := getFilteredTimezones("europe")
		for _, tz := range zones {
			if !strings.Contains(strings.ToLower(tz), "europe") {
				t.Errorf("Timezone %s doesn't match filter 'europe'", tz)
			}
		}
	})

	t.Run("filter by city", func(t *testing.T) {
		zones := getFilteredTimezones("york")
		found := false
		for _, tz := range zones {
			if tz == "America/New_York" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected America/New_York in results for 'york' filter")
		}
	})
}
