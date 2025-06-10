// ABOUTME: Example demonstrating the use of built-in datetime tools
// ABOUTME: Shows various date/time operations like parsing, formatting, and calculations

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	datetime "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/datetime"
	agentDomain "github.com/lexlapax/go-llms/pkg/agent/domain"
)

// Helper types for creating a minimal ToolContext for standalone tool execution

// minimalStateReader implements StateReader interface with empty state
type minimalStateReader struct {
	state *agentDomain.State
}

func (m *minimalStateReader) Get(key string) (interface{}, bool) {
	return m.state.Get(key)
}

func (m *minimalStateReader) Values() map[string]interface{} {
	return m.state.Values()
}

func (m *minimalStateReader) GetArtifact(id string) (*agentDomain.Artifact, bool) {
	return m.state.GetArtifact(id)
}

func (m *minimalStateReader) Artifacts() map[string]*agentDomain.Artifact {
	return m.state.Artifacts()
}

func (m *minimalStateReader) Messages() []agentDomain.Message {
	return m.state.Messages()
}

func (m *minimalStateReader) GetMetadata(key string) (interface{}, bool) {
	return m.state.GetMetadata(key)
}

func (m *minimalStateReader) Has(key string) bool {
	return m.state.Has(key)
}

func (m *minimalStateReader) Keys() []string {
	values := m.state.Values()
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	return keys
}

// minimalEventEmitter implements EventEmitter interface with no-op methods
type minimalEventEmitter struct{}

func (m *minimalEventEmitter) Emit(eventType agentDomain.EventType, data interface{}) {}
func (m *minimalEventEmitter) EmitProgress(current, total int, message string)        {}
func (m *minimalEventEmitter) EmitMessage(message string)                             {}
func (m *minimalEventEmitter) EmitError(err error)                                    {}
func (m *minimalEventEmitter) EmitCustom(eventName string, data interface{})          {}

// createToolContext creates a minimal ToolContext for standalone tool execution
func createToolContext(ctx context.Context) *agentDomain.ToolContext {
	state := agentDomain.NewState()
	stateReader := &minimalStateReader{state: state}

	toolCtx := &agentDomain.ToolContext{
		Context:   ctx,
		State:     stateReader,
		RunID:     "standalone-execution",
		Retry:     0,
		StartTime: time.Now(),
		Events:    &minimalEventEmitter{},
		Agent: agentDomain.AgentInfo{
			ID:          "standalone",
			Name:        "standalone-tool-executor",
			Description: "Minimal agent for standalone tool execution",
			Type:        agentDomain.AgentTypeLLM,
			Metadata:    make(map[string]interface{}),
		},
	}

	return toolCtx
}

func main() {
	ctx := context.Background()
	toolCtx := createToolContext(ctx)

	// List all datetime tools
	fmt.Println("=== Available DateTime Tools ===")
	fmt.Println()
	dateTimeTools := tools.Tools.ListByCategory("datetime")
	fmt.Printf("Total datetime tools: %d\n\n", len(dateTimeTools))
	for _, entry := range dateTimeTools {
		fmt.Printf("• %s: %s\n", entry.Metadata.Name, entry.Metadata.Description)
	}
	fmt.Println()

	// Example 1: Get current date/time in various formats and timezones
	fmt.Println("=== Example 1: Current Date/Time (datetime_now) ===")
	fmt.Println()
	nowTool := tools.MustGetTool("datetime_now")

	// Get current time in New York
	result, err := nowTool.Execute(toolCtx, map[string]interface{}{
		"timezone":           "America/New_York",
		"include_components": true,
		"include_week_info":  true,
		"include_timestamps": true,
		"format":             "Monday, January 2, 2006 at 3:04 PM MST",
	})
	if err != nil {
		log.Fatalf("Failed to get current time: %v", err)
	}

	fmt.Println("Current time in New York with full details:")
	if output, ok := result.(*datetime.DateTimeNowOutput); ok {
		fmt.Printf("  UTC: %s\n", output.UTC)
		fmt.Printf("  Local: %s\n", output.Local)
		fmt.Printf("  Timezone: %s (%s)\n", output.Timezone, output.TimezoneName)
		fmt.Printf("  Formatted: %s\n", output.Formatted)

		if output.Components != nil {
			fmt.Printf("  Components: Year=%d, Month=%d, Day=%d, Hour=%d, Minute=%d\n",
				output.Components.Year, output.Components.Month, output.Components.Day,
				output.Components.Hour, output.Components.Minute)
		}
		if output.WeekInfo != nil {
			fmt.Printf("  Week info: Week=%d, Day=%d\n",
				output.WeekInfo.WeekNumber, output.WeekInfo.DayOfWeek)
		}
		if output.Timestamps != nil {
			fmt.Printf("  Unix timestamp: %d\n", output.Timestamps.Unix)
		}
	}
	fmt.Println()

	// Get current time in multiple timezones
	timezones := []string{"UTC", "Europe/London", "Asia/Tokyo", "Australia/Sydney"}
	fmt.Println("Current time in different timezones:")
	for _, tz := range timezones {
		result, _ := nowTool.Execute(toolCtx, map[string]interface{}{
			"timezone": tz,
			"format":   "15:04 MST on Monday, Jan 2",
		})
		if output, ok := result.(*datetime.DateTimeNowOutput); ok {
			fmt.Printf("  %s: %s\n", tz, output.Formatted)
		}
	}
	fmt.Println()

	// Example 2: Parse various date formats
	fmt.Println("=== Example 2: Parse Dates (datetime_parse) ===")
	fmt.Println()
	parseTool := tools.MustGetTool("datetime_parse")

	// Parse different date formats
	dateStrings := []string{
		"2024-12-25",
		"December 25, 2024",
		"25/12/2024",
		"2024-01-15T10:30:00Z",
		"tomorrow",
		"next Monday",
		"in 3 days",
	}

	fmt.Println("Parsing various date formats:")
	for _, dateStr := range dateStrings {
		parseResult, err := parseTool.Execute(toolCtx, map[string]interface{}{
			"date_string": dateStr,
			"timezone":    "America/Los_Angeles",
		})
		if err != nil {
			fmt.Printf("  ❌ '%s': %v\n", dateStr, err)
		} else {
			if output, ok := parseResult.(*datetime.DateTimeParseOutput); ok {
				fmt.Printf("  ✓ '%s' → %s (format: %s)\n",
					dateStr, output.Parsed, output.DetectedFormat)
			}
		}
	}
	fmt.Println()

	// Example 3: Date calculations and operations
	fmt.Println("=== Example 3: Date Calculations (datetime_calculate) ===")
	fmt.Println()
	calcTool := tools.MustGetTool("datetime_calculate")

	// Add time units
	fmt.Println("Adding time units:")
	calcResult, err := calcTool.Execute(toolCtx, map[string]interface{}{
		"operation":  "add",
		"start_date": "2024-01-15T10:00:00Z",
		"unit":       "days",
		"value":      30,
	})
	if err != nil {
		log.Printf("Failed to calculate date: %v", err)
	} else {
		if output, ok := calcResult.(*datetime.DateTimeCalculateOutput); ok {
			fmt.Printf("  2024-01-15 + 30 days = %s\n", output.Result)
		}
	}

	// Add business days
	fmt.Println("\nBusiness day calculations:")
	businessResult, err := calcTool.Execute(toolCtx, map[string]interface{}{
		"operation":  "add_business_days",
		"start_date": "2024-01-15T10:00:00Z",
		"value":      5,
	})
	if err != nil {
		log.Printf("Failed to calculate business days: %v", err)
	} else {
		if output, ok := businessResult.(*datetime.DateTimeCalculateOutput); ok {
			fmt.Printf("  2024-01-15 (Monday) + 5 business days = %s\n", output.Result)
			fmt.Printf("  Business days added: %d\n", output.BusinessDays)
		}
	}

	// Calculate age
	fmt.Println("\nAge calculation:")
	birthDate := "1990-05-15T00:00:00Z"
	ageResult, err := calcTool.Execute(toolCtx, map[string]interface{}{
		"operation":  "age",
		"start_date": birthDate,
	})
	if err != nil {
		log.Printf("Failed to calculate age: %v", err)
	} else {
		if output, ok := ageResult.(*datetime.DateTimeCalculateOutput); ok {
			if output.Age != nil {
				fmt.Printf("  Born on %s: %d years, %d months, %d days old\n",
					birthDate[:10], output.Age.Years, output.Age.Months, output.Age.Days)
				fmt.Printf("  Total days: %d\n", output.Age.TotalDays)
				fmt.Printf("  Human readable: %s\n", output.Age.HumanReadable)
			}
		}
	}

	// Calculate duration between dates
	fmt.Println("\nDuration calculation:")
	durationResult, err := calcTool.Execute(toolCtx, map[string]interface{}{
		"operation":  "duration",
		"start_date": "2024-01-01T00:00:00Z",
		"end_date":   "2024-12-31T23:59:59Z",
	})
	if err != nil {
		log.Printf("Failed to calculate duration: %v", err)
	} else {
		if output, ok := durationResult.(*datetime.DateTimeCalculateOutput); ok {
			if output.Duration != nil {
				fmt.Printf("  Duration of 2024: %d days, %d hours, %d minutes\n",
					output.Duration.Days, output.Duration.Hours, output.Duration.Minutes)
				fmt.Printf("  Total seconds: %.0f\n", output.Duration.TotalSeconds)
				fmt.Printf("  Human readable: %s\n", output.Duration.HumanReadable)
			}
		}
	}
	fmt.Println()

	// Example 4: Format dates in various ways
	fmt.Println("=== Example 4: Format Dates (datetime_format) ===")
	fmt.Println()
	formatTool := tools.MustGetTool("datetime_format")

	// Format in multiple formats
	formatResult, err := formatTool.Execute(toolCtx, map[string]interface{}{
		"datetime":    "2024-12-25T10:30:00Z",
		"format_type": "multiple",
		"formats":     []string{"RFC3339", "Kitchen", "Monday, January 2, 2006"},
		"locale":      "es", // Spanish
	})
	if err != nil {
		log.Printf("Failed to format date: %v", err)
	} else {
		fmt.Println("Christmas 2024 in multiple formats:")
		if output, ok := formatResult.(*datetime.DateTimeFormatOutput); ok {
			if output.MultipleFormats != nil {
				for format, value := range output.MultipleFormats {
					fmt.Printf("  %s: %s\n", format, value)
				}
			}
			if output.Localized != nil {
				fmt.Printf("  Spanish: %s, %s\n",
					output.Localized.WeekdayName,
					output.Localized.MonthName)
			}
		}
	}

	// Relative time formatting
	fmt.Println("\nRelative time formatting:")
	twoHoursAgo := time.Now().Add(-2 * time.Hour).Format(time.RFC3339)
	relativeResult, _ := formatTool.Execute(toolCtx, map[string]interface{}{
		"datetime":    twoHoursAgo,
		"format_type": "relative",
	})
	if output, ok := relativeResult.(*datetime.DateTimeFormatOutput); ok {
		fmt.Printf("  %s → %s\n", twoHoursAgo[:16], output.RelativeTime)
	}

	// Localized formatting
	fmt.Println("\nLocalized date formatting:")
	locales := []string{"en", "es", "fr", "de", "ja"}
	for _, locale := range locales {
		locResult, _ := formatTool.Execute(toolCtx, map[string]interface{}{
			"datetime":    "2024-07-14T14:00:00Z", // Bastille Day
			"format_type": "standard",
			"locale":      locale,
		})
		if output, ok := locResult.(*datetime.DateTimeFormatOutput); ok {
			if output.Localized != nil {
				fmt.Printf("  %s: %s, %s\n", locale,
					output.Localized.WeekdayName, output.Localized.MonthName)
			}
		}
	}
	fmt.Println()

	// Example 5: Timezone and Unix timestamp conversions
	fmt.Println("=== Example 5: Conversions (datetime_convert) ===")
	fmt.Println()
	convertTool := tools.MustGetTool("datetime_convert")

	// Timezone conversion
	fmt.Println("Timezone conversions:")
	meetingTime := "2024-07-15T15:00:00Z"
	zones := []string{"America/New_York", "Europe/London", "Asia/Tokyo", "Australia/Sydney"}

	fmt.Printf("Meeting at %s UTC:\n", meetingTime[:16])
	for _, zone := range zones {
		result, _ := convertTool.Execute(toolCtx, map[string]interface{}{
			"operation":     "timezone",
			"datetime":      meetingTime,
			"from_timezone": "UTC",
			"to_timezone":   zone,
		})
		if output, ok := result.(*datetime.DateTimeConvertOutput); ok {
			fmt.Printf("  → %s: %s", zone, output.Converted)
			if output.DSTInfo != nil && output.DSTInfo.IsDST {
				fmt.Printf(" (DST active)")
			}
			fmt.Println()
		}
	}

	// Unix timestamp conversion
	fmt.Println("\nUnix timestamp conversions:")
	unixResult, err := convertTool.Execute(toolCtx, map[string]interface{}{
		"operation": "to_timestamp",
		"datetime":  "2024-01-01T00:00:00Z",
	})
	if err != nil {
		log.Printf("Failed to convert to unix: %v", err)
	} else if output, ok := unixResult.(*datetime.DateTimeConvertOutput); ok && output != nil {
		fmt.Printf("  2024-01-01 00:00:00 UTC → Unix: %d\n", output.Timestamp)
		fmt.Printf("  Unix milliseconds: %d\n", output.TimestampMillis)

		// Convert back from unix
		fromUnixResult, _ := convertTool.Execute(toolCtx, map[string]interface{}{
			"operation": "from_timestamp",
			"timestamp": output.Timestamp,
			"timezone":  "UTC",
		})
		if fromOutput, ok := fromUnixResult.(*datetime.DateTimeConvertOutput); ok && fromOutput != nil {
			fmt.Printf("  Unix %d → %s\n", output.Timestamp, fromOutput.Converted)
		}
	}
	fmt.Println()

	// Example 6: Get comprehensive date information
	fmt.Println("=== Example 6: Date Information (datetime_info) ===")
	fmt.Println()
	infoTool := tools.MustGetTool("datetime_info")

	// Get info for leap year date
	leapDate := "2024-02-29T00:00:00Z"
	infoResult, err := infoTool.Execute(toolCtx, map[string]interface{}{
		"date": leapDate, // Changed from "date_time" to "date"
	})
	if err != nil {
		log.Printf("Failed to get date info: %v", err)
	} else {
		fmt.Printf("Information for %s (leap year):\n", leapDate[:10])
		if output, ok := infoResult.(*datetime.DateTimeInfoOutput); ok {
			fmt.Printf("  Day of week: %d (%s)\n", output.DayOfWeek, output.DayOfWeekName)
			fmt.Printf("  Day of year: %d\n", output.DayOfYear)
			fmt.Printf("  Week number: %d\n", output.WeekNumber)
			fmt.Printf("  Is leap year: %v\n", output.IsLeapYear)
			fmt.Printf("  Days in month: %d\n", output.DaysInMonth)
			fmt.Printf("  Quarter: %d\n", output.Quarter)
			fmt.Printf("  Month: %s\n", output.MonthName)
		}
	}

	// Get info for current date
	fmt.Println("\nCurrent date information:")
	currentInfo, _ := infoTool.Execute(toolCtx, map[string]interface{}{
		"date": time.Now().Format(time.RFC3339),
	})
	if output, ok := currentInfo.(*datetime.DateTimeInfoOutput); ok {
		fmt.Printf("  Today is: %s, %s %d, %d\n",
			output.DayOfWeekName, output.MonthName, output.DayOfMonth, output.Year)
		fmt.Printf("  Week: %d of %d\n", output.WeekNumber, output.WeekYear)
	}
	fmt.Println()

	// Example 7: Compare and sort dates
	fmt.Println("=== Example 7: Compare Dates (datetime_compare) ===")
	fmt.Println()
	compareTool := tools.MustGetTool("datetime_compare")

	// Compare two dates
	date1 := "2024-01-15T10:00:00Z"
	date2 := "2024-01-20T15:30:00Z"
	compareResult, err := compareTool.Execute(toolCtx, map[string]interface{}{
		"operation": "compare",
		"date1":     date1,
		"date2":     date2,
	})
	if err != nil {
		log.Printf("Failed to compare dates: %v", err)
	} else {
		fmt.Printf("Comparing dates:\n")
		fmt.Printf("  Date 1: %s\n", date1)
		fmt.Printf("  Date 2: %s\n", date2)
		if output, ok := compareResult.(*datetime.DateTimeCompareOutput); ok {
			fmt.Printf("  Date1 is before Date2: %v\n", output.Before)
			fmt.Printf("  Date1 is after Date2: %v\n", output.After)
			fmt.Printf("  Dates are equal: %v\n", output.Equal)
			if output.Difference != nil {
				fmt.Printf("  Difference: %s\n", output.Difference.HumanReadable)
				fmt.Printf("  Total hours: %.1f\n", output.Difference.TotalHours)
			}
		}
	}

	// Find min/max dates
	dates := []string{
		"2024-03-15T00:00:00Z",
		"2024-01-10T00:00:00Z",
		"2024-06-20T00:00:00Z",
		"2024-02-28T00:00:00Z",
		"2024-12-31T00:00:00Z",
		"2024-07-04T00:00:00Z",
	}

	minResult, _ := compareTool.Execute(toolCtx, map[string]interface{}{
		"operation":    "find_extreme",
		"dates":        dates,
		"extreme_type": "earliest",
	})
	maxResult, _ := compareTool.Execute(toolCtx, map[string]interface{}{
		"operation":    "find_extreme",
		"dates":        dates,
		"extreme_type": "latest",
	})

	fmt.Println("\nFinding min/max from date list:")
	if minOutput, ok := minResult.(*datetime.DateTimeCompareOutput); ok {
		fmt.Printf("  Earliest date: %s\n", minOutput.ExtremeDate)
	}
	if maxOutput, ok := maxResult.(*datetime.DateTimeCompareOutput); ok {
		fmt.Printf("  Latest date: %s\n", maxOutput.ExtremeDate)
	}

	// Sort multiple dates
	fmt.Println("\nSorting dates:")
	sortResult, err := compareTool.Execute(toolCtx, map[string]interface{}{
		"operation":  "sort",
		"dates":      dates,
		"sort_order": "desc",
	})
	if err != nil {
		log.Printf("Failed to sort dates: %v", err)
	} else {
		if output, ok := sortResult.(*datetime.DateTimeCompareOutput); ok {
			fmt.Println("  Sorted (descending):")
			for i, date := range output.SortedDates {
				fmt.Printf("    %d. %s\n", i+1, date[:10])
			}
		}
	}

	// Check date ranges
	fmt.Println("\nDate range checks:")
	rangeResult, _ := compareTool.Execute(toolCtx, map[string]interface{}{
		"operation":   "range_check",
		"date1":       "2024-06-15T00:00:00Z",
		"range_start": "2024-06-01T00:00:00Z",
		"range_end":   "2024-06-30T00:00:00Z",
	})
	if output, ok := rangeResult.(*datetime.DateTimeCompareOutput); ok {
		fmt.Printf("  Is 2024-06-15 between June 1-30? %v\n", output.InRange)
	}
}
