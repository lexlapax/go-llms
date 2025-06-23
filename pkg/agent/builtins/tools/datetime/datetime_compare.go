// ABOUTME: Provides date/time comparison functionality
// ABOUTME: Supports before/after/equal checks, same period comparisons, and sorting

package datetime

import (
	"fmt"
	"sort"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	agentDomain "github.com/lexlapax/go-llms/pkg/agent/domain"
	atools "github.com/lexlapax/go-llms/pkg/agent/tools"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// DateTimeCompareInput defines the input for the datetime_compare tool
type DateTimeCompareInput struct {
	// Operation: "compare", "same_period", "range_check", "sort", "find_extreme"
	Operation string `json:"operation"`
	// First date/time (RFC3339 format preferred)
	Date1 string `json:"date1,omitempty"`
	// Second date/time (RFC3339 format preferred)
	Date2 string `json:"date2,omitempty"`
	// List of dates for sort/find_extreme operations
	Dates []string `json:"dates,omitempty"`
	// Period type for same_period: "day", "week", "month", "year"
	PeriodType string `json:"period_type,omitempty"`
	// Range start date
	RangeStart string `json:"range_start,omitempty"`
	// Range end date
	RangeEnd string `json:"range_end,omitempty"`
	// Sort order: "asc" or "desc"
	SortOrder string `json:"sort_order,omitempty"`
	// Extreme type: "earliest" or "latest"
	ExtremeType string `json:"extreme_type,omitempty"`
	// Timezone for comparisons
	Timezone string `json:"timezone,omitempty"`
}

// DateTimeCompareOutput defines the output for the datetime_compare tool
type DateTimeCompareOutput struct {
	// Comparison results
	Before bool `json:"before,omitempty"`
	After  bool `json:"after,omitempty"`
	Equal  bool `json:"equal,omitempty"`
	// Same period result
	SamePeriod bool `json:"same_period,omitempty"`
	// Range check result
	InRange bool `json:"in_range,omitempty"`
	// Sorted dates
	SortedDates []string `json:"sorted_dates,omitempty"`
	// Extreme date
	ExtremeDate string `json:"extreme_date,omitempty"`
	// Time difference
	Difference *TimeDifference `json:"difference,omitempty"`
}

// TimeDifference holds the difference between two times
type TimeDifference struct {
	Days         int     `json:"days"`
	Hours        int     `json:"hours"`
	Minutes      int     `json:"minutes"`
	Seconds      int     `json:"seconds"`
	TotalHours   float64 `json:"total_hours"`
	TotalMinutes float64 `json:"total_minutes"`
	TotalSeconds float64 `json:"total_seconds"`
	// Human-readable format
	HumanReadable string `json:"human_readable"`
}

// dateTimeCompareExecute is the execution function for datetime_compare
func dateTimeCompareExecute(ctx *agentDomain.ToolContext, input DateTimeCompareInput) (*DateTimeCompareOutput, error) {
	// Emit start event
	if ctx.Events != nil {
		ctx.Events.Emit(agentDomain.EventToolCall, agentDomain.ToolCallEventData{
			ToolName:   "datetime_compare",
			Parameters: input,
			RequestID:  ctx.RunID,
		})
	}

	output := &DateTimeCompareOutput{}

	// Apply timezone if specified
	loc := time.UTC
	if input.Timezone != "" {
		var err error
		loc, err = time.LoadLocation(input.Timezone)
		if err != nil {
			return nil, fmt.Errorf("invalid timezone: %w", err)
		}
	} else if ctx.State != nil {
		// Check state for default timezone
		if defaultTZ, exists := ctx.State.Get("default_timezone"); exists {
			if tzStr, ok := defaultTZ.(string); ok && tzStr != "" {
				var err error
				loc, err = time.LoadLocation(tzStr)
				if err != nil {
					// Fall back to UTC if timezone is invalid
					loc = time.UTC
				}
			}
		}
	}

	switch input.Operation {
	case "compare":
		if input.Date1 == "" || input.Date2 == "" {
			return nil, fmt.Errorf("date1 and date2 are required for compare operation")
		}

		time1, err := parseDateInLocation(input.Date1, loc)
		if err != nil {
			return nil, fmt.Errorf("invalid date1: %w", err)
		}

		time2, err := parseDateInLocation(input.Date2, loc)
		if err != nil {
			return nil, fmt.Errorf("invalid date2: %w", err)
		}

		output.Before = time1.Before(time2)
		output.After = time1.After(time2)
		output.Equal = time1.Equal(time2)
		output.Difference = calculateTimeDifference(time1, time2)

	case "same_period":
		if input.Date1 == "" || input.Date2 == "" {
			return nil, fmt.Errorf("date1 and date2 are required for same_period operation")
		}
		if input.PeriodType == "" {
			return nil, fmt.Errorf("period_type is required for same_period operation")
		}

		time1, err := parseDateInLocation(input.Date1, loc)
		if err != nil {
			return nil, fmt.Errorf("invalid date1: %w", err)
		}

		time2, err := parseDateInLocation(input.Date2, loc)
		if err != nil {
			return nil, fmt.Errorf("invalid date2: %w", err)
		}

		output.SamePeriod = areSamePeriod(time1, time2, input.PeriodType)

	case "range_check":
		if input.Date1 == "" || input.RangeStart == "" || input.RangeEnd == "" {
			return nil, fmt.Errorf("date1, range_start, and range_end are required for range_check operation")
		}

		date, err := parseDateInLocation(input.Date1, loc)
		if err != nil {
			return nil, fmt.Errorf("invalid date1: %w", err)
		}

		rangeStart, err := parseDateInLocation(input.RangeStart, loc)
		if err != nil {
			return nil, fmt.Errorf("invalid range_start: %w", err)
		}

		rangeEnd, err := parseDateInLocation(input.RangeEnd, loc)
		if err != nil {
			return nil, fmt.Errorf("invalid range_end: %w", err)
		}

		output.InRange = !date.Before(rangeStart) && !date.After(rangeEnd)

	case "sort":
		if len(input.Dates) == 0 {
			return nil, fmt.Errorf("dates array is required for sort operation")
		}

		// Parse all dates
		parsedDates := make([]time.Time, 0, len(input.Dates))
		for _, dateStr := range input.Dates {
			parsed, err := parseDateInLocation(dateStr, loc)
			if err != nil {
				return nil, fmt.Errorf("invalid date '%s': %w", dateStr, err)
			}
			parsedDates = append(parsedDates, parsed)
		}

		// Sort the dates
		if input.SortOrder == "desc" {
			sort.Slice(parsedDates, func(i, j int) bool {
				return parsedDates[i].After(parsedDates[j])
			})
		} else {
			sort.Slice(parsedDates, func(i, j int) bool {
				return parsedDates[i].Before(parsedDates[j])
			})
		}

		// Convert back to strings
		output.SortedDates = make([]string, len(parsedDates))
		for i, t := range parsedDates {
			output.SortedDates[i] = t.Format(time.RFC3339)
		}

	case "find_extreme":
		if len(input.Dates) == 0 {
			return nil, fmt.Errorf("dates array is required for find_extreme operation")
		}
		if input.ExtremeType == "" {
			input.ExtremeType = "earliest"
		}

		// Parse all dates
		var extreme time.Time
		for i, dateStr := range input.Dates {
			parsed, err := parseDateInLocation(dateStr, loc)
			if err != nil {
				return nil, fmt.Errorf("invalid date '%s': %w", dateStr, err)
			}

			if i == 0 {
				extreme = parsed
			} else {
				if input.ExtremeType == "earliest" && parsed.Before(extreme) {
					extreme = parsed
				} else if input.ExtremeType == "latest" && parsed.After(extreme) {
					extreme = parsed
				}
			}
		}

		output.ExtremeDate = extreme.Format(time.RFC3339)

	default:
		return nil, fmt.Errorf("invalid operation: %s", input.Operation)
	}

	// Emit result event
	if ctx.Events != nil {
		ctx.Events.Emit(agentDomain.EventToolResult, agentDomain.ToolResultEventData{
			ToolName:  "datetime_compare",
			Result:    output,
			RequestID: ctx.RunID,
		})
	}

	return output, nil
}

// DateTimeCompare returns a tool that compares dates and times
// This tool provides various comparison operations including before/after checks,
// same period comparisons, range checking, sorting, and finding extreme dates.
// It supports timezone-aware comparisons and detailed time difference calculations.
func DateTimeCompare() agentDomain.Tool {
	// Define parameter schema
	paramSchema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"operation": {
				Type:        "string",
				Description: "Comparison operation",
				Enum:        []string{"compare", "same_period", "range_check", "sort", "find_extreme"},
			},
			"date1": {
				Type:        "string",
				Description: "First date/time (RFC3339 format preferred)",
			},
			"date2": {
				Type:        "string",
				Description: "Second date/time (RFC3339 format preferred)",
			},
			"dates": {
				Type:        "array",
				Description: "List of dates for sort/find_extreme operations",
				Items: &schemaDomain.Property{
					Type: "string",
				},
			},
			"period_type": {
				Type:        "string",
				Description: "Period type for same_period operation",
				Enum:        []string{"day", "week", "month", "year"},
			},
			"range_start": {
				Type:        "string",
				Description: "Range start date",
			},
			"range_end": {
				Type:        "string",
				Description: "Range end date",
			},
			"sort_order": {
				Type:        "string",
				Description: "Sort order",
				Enum:        []string{"asc", "desc"},
			},
			"extreme_type": {
				Type:        "string",
				Description: "Extreme type",
				Enum:        []string{"earliest", "latest"},
			},
			"timezone": {
				Type:        "string",
				Description: "Timezone for comparisons",
			},
		},
		Required: []string{"operation"},
	}

	// Define output schema
	outputSchema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"before": {
				Type:        "boolean",
				Description: "Whether date1 is before date2",
			},
			"after": {
				Type:        "boolean",
				Description: "Whether date1 is after date2",
			},
			"equal": {
				Type:        "boolean",
				Description: "Whether date1 equals date2",
			},
			"same_period": {
				Type:        "boolean",
				Description: "Whether dates are in the same period",
			},
			"in_range": {
				Type:        "boolean",
				Description: "Whether date is within the specified range",
			},
			"sorted_dates": {
				Type:        "array",
				Description: "Sorted list of dates",
				Items: &schemaDomain.Property{
					Type: "string",
				},
			},
			"extreme_date": {
				Type:        "string",
				Description: "The earliest or latest date from the list",
			},
			"difference": {
				Type:        "object",
				Description: "Time difference between two dates",
				Properties: map[string]schemaDomain.Property{
					"days": {
						Type:        "integer",
						Description: "Number of complete days",
					},
					"hours": {
						Type:        "integer",
						Description: "Remaining hours after days",
					},
					"minutes": {
						Type:        "integer",
						Description: "Remaining minutes after hours",
					},
					"seconds": {
						Type:        "integer",
						Description: "Remaining seconds after minutes",
					},
					"total_hours": {
						Type:        "number",
						Description: "Total difference in hours",
					},
					"total_minutes": {
						Type:        "number",
						Description: "Total difference in minutes",
					},
					"total_seconds": {
						Type:        "number",
						Description: "Total difference in seconds",
					},
					"human_readable": {
						Type:        "string",
						Description: "Human-readable time difference",
					},
				},
			},
		},
	}

	builder := atools.NewToolBuilder("datetime_compare", "Compare dates and times with various operations").
		WithFunction(dateTimeCompareExecute).
		WithParameterSchema(paramSchema).
		WithOutputSchema(outputSchema).
		WithUsageInstructions(`The datetime_compare tool provides comprehensive date/time comparison capabilities:

Operations:
1. Compare:
   - Check if date1 is before, after, or equal to date2
   - Calculate detailed time difference between dates
   - Provides human-readable difference format
   - Handles timezones correctly

2. Same Period:
   - Check if two dates fall within the same period
   - Supported periods: day, week, month, year
   - Week comparisons use ISO week numbering
   - Timezone-aware comparisons

3. Range Check:
   - Verify if a date falls within a specified range
   - Inclusive of both start and end dates
   - Useful for deadline validation
   - Date range filtering

4. Sort:
   - Sort multiple dates in ascending or descending order
   - Handles any parseable date format
   - Returns dates in RFC3339 format
   - Efficient for large date lists

5. Find Extreme:
   - Find the earliest or latest date from a list
   - Useful for finding min/max dates
   - Deadline calculations
   - Event scheduling

Time Difference Output:
- Days, hours, minutes, seconds breakdown
- Total time in hours, minutes, and seconds
- Human-readable format (e.g., "2 days, 3 hours and 15 minutes")
- Direction indicator (ago/in the future)

State Integration:
- default_timezone: Used when timezone not specified in input

Common Use Cases:
- Deadline validation and monitoring
- Event scheduling and conflict detection
- Age calculations and date arithmetic
- Historical data analysis
- Time period grouping and filtering`).
		WithExamples([]agentDomain.ToolExample{
			{
				Name:        "Basic date comparison",
				Description: "Compare two dates and get time difference",
				Scenario:    "When you need to know the relationship between two dates",
				Input: map[string]interface{}{
					"operation": "compare",
					"date1":     "2024-03-15T10:30:00Z",
					"date2":     "2024-03-20T15:45:00Z",
				},
				Output: map[string]interface{}{
					"before": true,
					"after":  false,
					"equal":  false,
					"difference": map[string]interface{}{
						"days":           5,
						"hours":          5,
						"minutes":        15,
						"seconds":        0,
						"total_hours":    125.25,
						"total_minutes":  7515.0,
						"total_seconds":  450900.0,
						"human_readable": "in 5 days, 5 hours and 15 minutes",
					},
				},
				Explanation: "Date1 is before date2 by 5 days, 5 hours, and 15 minutes",
			},
			{
				Name:        "Check same month",
				Description: "Verify if two dates are in the same month",
				Scenario:    "When grouping events by month",
				Input: map[string]interface{}{
					"operation":   "same_period",
					"date1":       "2024-03-05",
					"date2":       "2024-03-25",
					"period_type": "month",
				},
				Output: map[string]interface{}{
					"same_period": true,
				},
				Explanation: "Both dates are in March 2024",
			},
			{
				Name:        "Check same week",
				Description: "Check if dates fall in the same ISO week",
				Scenario:    "When organizing weekly meetings or reports",
				Input: map[string]interface{}{
					"operation":   "same_period",
					"date1":       "2024-03-18",
					"date2":       "2024-03-22",
					"period_type": "week",
				},
				Output: map[string]interface{}{
					"same_period": true,
				},
				Explanation: "Both dates are in ISO week 12 of 2024",
			},
			{
				Name:        "Date range validation",
				Description: "Check if a date falls within a range",
				Scenario:    "When validating if an event date is within a valid period",
				Input: map[string]interface{}{
					"operation":   "range_check",
					"date1":       "2024-03-15",
					"range_start": "2024-03-01",
					"range_end":   "2024-03-31",
				},
				Output: map[string]interface{}{
					"in_range": true,
				},
				Explanation: "March 15 is within the March 2024 range",
			},
			{
				Name:        "Sort dates chronologically",
				Description: "Sort a list of dates in ascending order",
				Scenario:    "When organizing events or deadlines chronologically",
				Input: map[string]interface{}{
					"operation":  "sort",
					"dates":      []string{"2024-12-25", "2024-01-01", "2024-07-04", "2024-03-15"},
					"sort_order": "asc",
				},
				Output: map[string]interface{}{
					"sorted_dates": []string{
						"2024-01-01T00:00:00Z",
						"2024-03-15T00:00:00Z",
						"2024-07-04T00:00:00Z",
						"2024-12-25T00:00:00Z",
					},
				},
				Explanation: "Dates sorted from earliest to latest",
			},
			{
				Name:        "Sort dates descending",
				Description: "Sort dates from newest to oldest",
				Scenario:    "When showing most recent items first",
				Input: map[string]interface{}{
					"operation":  "sort",
					"dates":      []string{"2024-01-15", "2024-03-10", "2024-02-20"},
					"sort_order": "desc",
				},
				Output: map[string]interface{}{
					"sorted_dates": []string{
						"2024-03-10T00:00:00Z",
						"2024-02-20T00:00:00Z",
						"2024-01-15T00:00:00Z",
					},
				},
				Explanation: "Dates sorted from latest to earliest",
			},
			{
				Name:        "Find earliest date",
				Description: "Find the earliest date from a list",
				Scenario:    "When finding the first deadline or oldest record",
				Input: map[string]interface{}{
					"operation":    "find_extreme",
					"dates":        []string{"2024-06-15", "2024-03-01", "2024-12-31", "2024-01-10"},
					"extreme_type": "earliest",
				},
				Output: map[string]interface{}{
					"extreme_date": "2024-01-10T00:00:00Z",
				},
				Explanation: "January 10 is the earliest date in the list",
			},
			{
				Name:        "Find latest date",
				Description: "Find the most recent date",
				Scenario:    "When finding the last update or newest entry",
				Input: map[string]interface{}{
					"operation":    "find_extreme",
					"dates":        []string{"2024-02-28", "2024-03-15", "2024-01-01"},
					"extreme_type": "latest",
				},
				Output: map[string]interface{}{
					"extreme_date": "2024-03-15T00:00:00Z",
				},
				Explanation: "March 15 is the latest date in the list",
			},
			{
				Name:        "Compare with timezone",
				Description: "Compare dates in specific timezone",
				Scenario:    "When working with dates in different timezones",
				Input: map[string]interface{}{
					"operation": "compare",
					"date1":     "2024-03-15 10:00:00",
					"date2":     "2024-03-15 14:00:00",
					"timezone":  "America/New_York",
				},
				Output: map[string]interface{}{
					"before": true,
					"after":  false,
					"equal":  false,
					"difference": map[string]interface{}{
						"days":           0,
						"hours":          4,
						"minutes":        0,
						"seconds":        0,
						"total_hours":    4.0,
						"total_minutes":  240.0,
						"total_seconds":  14400.0,
						"human_readable": "in 4 hours",
					},
				},
				Explanation: "Comparing times in New York timezone, 4 hours apart",
			},
			{
				Name:        "Compare past dates",
				Description: "Compare dates where date1 is after date2",
				Scenario:    "When calculating time elapsed since an event",
				Input: map[string]interface{}{
					"operation": "compare",
					"date1":     "2024-03-20",
					"date2":     "2024-03-15",
				},
				Output: map[string]interface{}{
					"before": false,
					"after":  true,
					"equal":  false,
					"difference": map[string]interface{}{
						"days":           5,
						"hours":          0,
						"minutes":        0,
						"seconds":        0,
						"total_hours":    120.0,
						"total_minutes":  7200.0,
						"total_seconds":  432000.0,
						"human_readable": "5 days ago",
					},
				},
				Explanation: "Date1 is 5 days after date2, shown as '5 days ago'",
			},
		}).
		WithConstraints([]string{
			"Dates must be in a parseable format (RFC3339 preferred)",
			"Timezone must be a valid IANA timezone name",
			"Period type must be one of: day, week, month, year",
			"Sort order must be 'asc' or 'desc' (default: asc)",
			"Extreme type must be 'earliest' or 'latest' (default: earliest)",
			"Range checks are inclusive of both start and end dates",
			"Week comparisons use ISO week numbering (Monday as first day)",
			"Time differences show direction (ago for past, in for future)",
		}).
		WithErrorGuidance(map[string]string{
			"invalid date1":                                  "Ensure date1 is in a valid format. RFC3339 (e.g., '2024-03-15T14:30:00Z') is preferred",
			"invalid date2":                                  "Ensure date2 is in a valid format. RFC3339 is preferred",
			"invalid timezone":                               "Use a valid IANA timezone name like 'America/New_York' or 'UTC'",
			"date1 and date2 are required":                   "Provide both date1 and date2 for comparison operations",
			"period_type is required":                        "Specify period_type as one of: day, week, month, year",
			"dates array is required":                        "Provide an array of date strings for sort or find_extreme operations",
			"invalid operation":                              "Use one of: compare, same_period, range_check, sort, find_extreme",
			"date1, range_start, and range_end are required": "All three dates are needed for range_check operation",
			"invalid date in array":                          "Ensure all dates in the array are in valid formats",
		}).
		WithCategory("datetime").
		WithTags([]string{"datetime", "comparison", "before", "after", "range", "sort", "period", "difference", "timezone"}).
		WithVersion("2.0.0").
		WithBehavior(true, false, false, "fast")

	return builder.Build()
}

// parseDateInLocation parses a date string and applies the given location
func parseDateInLocation(dateStr string, loc *time.Location) (time.Time, error) {
	// First try to parse the date
	parsed, err := parseDate(dateStr)
	if err != nil {
		return time.Time{}, err
	}

	// If the parsed time doesn't have a timezone, apply the given location
	if parsed.Location() == time.UTC || parsed.Location() == time.Local {
		parsed = time.Date(
			parsed.Year(), parsed.Month(), parsed.Day(),
			parsed.Hour(), parsed.Minute(), parsed.Second(),
			parsed.Nanosecond(), loc,
		)
	}

	return parsed, nil
}

// calculateTimeDifference calculates the difference between two times
func calculateTimeDifference(time1, time2 time.Time) *TimeDifference {
	duration := time2.Sub(time1)
	absDuration := duration
	if duration < 0 {
		absDuration = -duration
	}

	totalSeconds := absDuration.Seconds()
	totalMinutes := absDuration.Minutes()
	totalHours := absDuration.Hours()

	days := int(totalHours / 24)
	hours := int(totalHours) % 24
	minutes := int(totalMinutes) % 60
	seconds := int(totalSeconds) % 60

	// Create human-readable format
	var humanParts []string
	if days != 0 {
		if days == 1 {
			humanParts = append(humanParts, "1 day")
		} else {
			humanParts = append(humanParts, fmt.Sprintf("%d days", days))
		}
	}
	if hours != 0 {
		if hours == 1 {
			humanParts = append(humanParts, "1 hour")
		} else {
			humanParts = append(humanParts, fmt.Sprintf("%d hours", hours))
		}
	}
	if minutes != 0 {
		if minutes == 1 {
			humanParts = append(humanParts, "1 minute")
		} else {
			humanParts = append(humanParts, fmt.Sprintf("%d minutes", minutes))
		}
	}
	if seconds != 0 || len(humanParts) == 0 {
		if seconds == 1 {
			humanParts = append(humanParts, "1 second")
		} else {
			humanParts = append(humanParts, fmt.Sprintf("%d seconds", seconds))
		}
	}

	humanReadable := ""
	for i, part := range humanParts {
		if i > 0 && i == len(humanParts)-1 {
			humanReadable += " and "
		} else if i > 0 {
			humanReadable += ", "
		}
		humanReadable += part
	}

	// Add direction
	if duration < 0 {
		humanReadable += " ago"
	} else if duration > 0 {
		humanReadable = "in " + humanReadable
	} else {
		humanReadable = "same time"
	}

	return &TimeDifference{
		Days:          days,
		Hours:         hours,
		Minutes:       minutes,
		Seconds:       seconds,
		TotalHours:    totalHours,
		TotalMinutes:  totalMinutes,
		TotalSeconds:  totalSeconds,
		HumanReadable: humanReadable,
	}
}

// areSamePeriod checks if two times are in the same period
func areSamePeriod(time1, time2 time.Time, periodType string) bool {
	switch periodType {
	case "day":
		return time1.Year() == time2.Year() &&
			time1.Month() == time2.Month() &&
			time1.Day() == time2.Day()
	case "week":
		year1, week1 := time1.ISOWeek()
		year2, week2 := time2.ISOWeek()
		return year1 == year2 && week1 == week2
	case "month":
		return time1.Year() == time2.Year() &&
			time1.Month() == time2.Month()
	case "year":
		return time1.Year() == time2.Year()
	default:
		return false
	}
}

func init() {
	tools.MustRegisterTool("datetime_compare", DateTimeCompare(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "datetime_compare",
			Category:    "datetime",
			Tags:        []string{"datetime", "comparison", "before", "after", "range", "sort", "period", "difference"},
			Description: "Compare dates and times with operations like before/after, same period, range checks, and sorting",
			Version:     "2.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Basic comparison",
					Description: "Check if date1 is before date2",
					Code:        `DateTimeCompare().Execute(ctx, DateTimeCompareInput{Operation: "compare", Date1: "2024-03-15", Date2: "2024-03-16"})`,
				},
				{
					Name:        "Same period check",
					Description: "Check if two dates are in the same month",
					Code:        `DateTimeCompare().Execute(ctx, DateTimeCompareInput{Operation: "same_period", Date1: "2024-03-15", Date2: "2024-03-25", PeriodType: "month"})`,
				},
				{
					Name:        "Range check",
					Description: "Check if a date falls within a range",
					Code:        `DateTimeCompare().Execute(ctx, DateTimeCompareInput{Operation: "range_check", Date1: "2024-03-15", RangeStart: "2024-03-01", RangeEnd: "2024-03-31"})`,
				},
				{
					Name:        "Sort dates",
					Description: "Sort multiple dates in ascending order",
					Code:        `DateTimeCompare().Execute(ctx, DateTimeCompareInput{Operation: "sort", Dates: ["2024-03-15", "2024-01-01", "2024-12-31"], SortOrder: "asc"})`,
				},
				{
					Name:        "Find earliest",
					Description: "Find the earliest date from a list",
					Code:        `DateTimeCompare().Execute(ctx, DateTimeCompareInput{Operation: "find_extreme", Dates: ["2024-03-15", "2024-01-01", "2024-12-31"], ExtremeType: "earliest"})`,
				},
			},
		},
		RequiredPermissions: []string{},
		ResourceUsage: tools.ResourceInfo{
			Memory:      "low",
			Network:     false,
			FileSystem:  false,
			Concurrency: true,
		},
		UsageInstructions: `The datetime_compare tool provides various date/time comparison operations:
- compare: Check if date1 is before, after, or equal to date2, with time difference
- same_period: Check if two dates fall within the same day, week, month, or year
- range_check: Verify if a date falls within a specified range (inclusive)
- sort: Sort multiple dates in ascending or descending order
- find_extreme: Find the earliest or latest date from a list

All operations support timezone specification for accurate comparisons.`,
		Constraints: []string{
			"Dates must be in a parseable format (RFC3339 preferred)",
			"Timezone must be a valid IANA timezone name",
			"Period type must be one of: day, week, month, year",
			"Sort order must be 'asc' or 'desc' (default: asc)",
			"Extreme type must be 'earliest' or 'latest' (default: earliest)",
			"Range checks are inclusive of both start and end dates",
		},
		ErrorGuidance: map[string]string{
			"invalid date1":                "Ensure date1 is in a valid format. RFC3339 is preferred",
			"invalid date2":                "Ensure date2 is in a valid format. RFC3339 is preferred",
			"invalid timezone":             "Use a valid IANA timezone name like 'America/New_York' or 'UTC'",
			"date1 and date2 are required": "Provide both date1 and date2 for comparison operations",
			"period_type is required":      "Specify period_type (day, week, month, year) for same_period operation",
			"dates array is required":      "Provide an array of dates for sort or find_extreme operations",
			"invalid operation":            "Use one of: compare, same_period, range_check, sort, find_extreme",
			"range dates required":         "Provide date1, range_start, and range_end for range_check operation",
		},
		IsDeterministic:      true,
		IsDestructive:        false,
		RequiresConfirmation: false,
		EstimatedLatency:     "instant",
	})
}
