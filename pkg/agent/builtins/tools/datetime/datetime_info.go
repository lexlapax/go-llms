// ABOUTME: Provides date information functionality for analyzing dates
// ABOUTME: Includes day/week/month/year info, leap year check, period boundaries

package datetime

import (
	"fmt"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	agentDomain "github.com/lexlapax/go-llms/pkg/agent/domain"
	atools "github.com/lexlapax/go-llms/pkg/agent/tools"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// DateTimeInfoInput defines the input for the datetime_info tool
type DateTimeInfoInput struct {
	// Date to get information about (RFC3339 format preferred)
	Date string `json:"date"`
	// Timezone for the date (default: UTC)
	Timezone string `json:"timezone,omitempty"`
	// WeekStartDay (0 = Sunday, 1 = Monday)
	WeekStartDay int `json:"week_start_day,omitempty"`
}

// DateTimeInfoOutput defines the output for the datetime_info tool
type DateTimeInfoOutput struct {
	// Original date in RFC3339 format
	Date string `json:"date"`
	// Day information
	DayOfWeek     int    `json:"day_of_week"`      // 0 = Sunday, 6 = Saturday
	DayOfWeekName string `json:"day_of_week_name"` // e.g., "Monday"
	DayOfWeekISO  int    `json:"day_of_week_iso"`  // 1 = Monday, 7 = Sunday
	DayOfMonth    int    `json:"day_of_month"`     // 1-31
	DayOfYear     int    `json:"day_of_year"`      // 1-366
	// Week information
	WeekNumber int `json:"week_number"` // ISO week number (1-53)
	WeekYear   int `json:"week_year"`   // Year of the ISO week
	// Month information
	Month       int    `json:"month"`         // 1-12
	MonthName   string `json:"month_name"`    // e.g., "January"
	DaysInMonth int    `json:"days_in_month"` // 28-31
	// Quarter information
	Quarter int `json:"quarter"` // 1-4
	// Year information
	Year       int  `json:"year"`
	IsLeapYear bool `json:"is_leap_year"`
	// Period boundaries
	StartOfWeek    string `json:"start_of_week"`    // RFC3339 format
	EndOfWeek      string `json:"end_of_week"`      // RFC3339 format
	StartOfMonth   string `json:"start_of_month"`   // RFC3339 format
	EndOfMonth     string `json:"end_of_month"`     // RFC3339 format
	StartOfQuarter string `json:"start_of_quarter"` // RFC3339 format
	EndOfQuarter   string `json:"end_of_quarter"`   // RFC3339 format
	StartOfYear    string `json:"start_of_year"`    // RFC3339 format
	EndOfYear      string `json:"end_of_year"`      // RFC3339 format
}

// dateTimeInfoExecute is the execution function for datetime_info
func dateTimeInfoExecute(ctx *agentDomain.ToolContext, input DateTimeInfoInput) (*DateTimeInfoOutput, error) {
	// Emit start event
	if ctx.Events != nil {
		ctx.Events.Emit(agentDomain.EventToolCall, agentDomain.ToolCallEventData{
			ToolName:   "datetime_info",
			Parameters: input,
			RequestID:  ctx.RunID,
		})
	}

	// Parse the input date
	parsedTime, err := time.Parse(time.RFC3339, input.Date)
	if err != nil {
		// Try other common formats
		for _, format := range []string{
			"2006-01-02",
			"2006-01-02 15:04:05",
			"01/02/2006",
			"02-Jan-2006",
		} {
			if parsed, err2 := time.Parse(format, input.Date); err2 == nil {
				parsedTime = parsed
				err = nil
				break
			}
		}
		if err != nil {
			return nil, fmt.Errorf("invalid date format: %w", err)
		}
	}

	// Apply timezone if specified
	if input.Timezone != "" {
		loc, err := time.LoadLocation(input.Timezone)
		if err != nil {
			return nil, fmt.Errorf("invalid timezone: %w", err)
		}
		parsedTime = parsedTime.In(loc)
	} else if ctx.State != nil {
		// Check state for default timezone
		if defaultTZ, exists := ctx.State.Get("default_timezone"); exists {
			if tzStr, ok := defaultTZ.(string); ok && tzStr != "" {
				loc, err := time.LoadLocation(tzStr)
				if err == nil {
					parsedTime = parsedTime.In(loc)
				}
			}
		}
	}

	// Calculate all the information
	year, week := parsedTime.ISOWeek()
	quarter := (int(parsedTime.Month())-1)/3 + 1

	// ISO day of week: Monday = 1, Sunday = 7
	isoDayOfWeek := int(parsedTime.Weekday())
	if isoDayOfWeek == 0 {
		isoDayOfWeek = 7
	}

	output := &DateTimeInfoOutput{
		Date:          parsedTime.Format(time.RFC3339),
		DayOfWeek:     int(parsedTime.Weekday()),
		DayOfWeekName: parsedTime.Weekday().String(),
		DayOfWeekISO:  isoDayOfWeek,
		DayOfMonth:    parsedTime.Day(),
		DayOfYear:     parsedTime.YearDay(),
		WeekNumber:    week,
		WeekYear:      year,
		Month:         int(parsedTime.Month()),
		MonthName:     parsedTime.Month().String(),
		DaysInMonth:   daysInMonth(parsedTime),
		Quarter:       quarter,
		Year:          parsedTime.Year(),
		IsLeapYear:    isLeapYear(parsedTime.Year()),
	}

	// Calculate period boundaries
	weekStartDay := time.Sunday
	if input.WeekStartDay == 1 {
		weekStartDay = time.Monday
	}

	// Start and end of week
	startOfWeek := startOfWeek(parsedTime, weekStartDay)
	endOfWeek := endOfDay(startOfWeek.AddDate(0, 0, 6))
	output.StartOfWeek = startOfWeek.Format(time.RFC3339)
	output.EndOfWeek = endOfWeek.Format(time.RFC3339)

	// Start and end of month
	startOfMonth := time.Date(parsedTime.Year(), parsedTime.Month(), 1, 0, 0, 0, 0, parsedTime.Location())
	endOfMonth := endOfDay(startOfMonth.AddDate(0, 1, -1))
	output.StartOfMonth = startOfMonth.Format(time.RFC3339)
	output.EndOfMonth = endOfMonth.Format(time.RFC3339)

	// Start and end of quarter
	quarterMonth := time.Month((quarter-1)*3 + 1)
	startOfQuarter := time.Date(parsedTime.Year(), quarterMonth, 1, 0, 0, 0, 0, parsedTime.Location())
	endOfQuarter := endOfDay(startOfQuarter.AddDate(0, 3, -1))
	output.StartOfQuarter = startOfQuarter.Format(time.RFC3339)
	output.EndOfQuarter = endOfQuarter.Format(time.RFC3339)

	// Start and end of year
	startOfYear := time.Date(parsedTime.Year(), 1, 1, 0, 0, 0, 0, parsedTime.Location())
	endOfYear := endOfDay(time.Date(parsedTime.Year(), 12, 31, 0, 0, 0, 0, parsedTime.Location()))
	output.StartOfYear = startOfYear.Format(time.RFC3339)
	output.EndOfYear = endOfYear.Format(time.RFC3339)

	// Emit result event
	if ctx.Events != nil {
		ctx.Events.Emit(agentDomain.EventToolResult, agentDomain.ToolResultEventData{
			ToolName:  "datetime_info",
			Result:    output,
			RequestID: ctx.RunID,
		})
	}

	return output, nil
}

// DateTimeInfo returns a tool that provides comprehensive information about a specific date.
// It extracts day/week/month/year components, calculates ISO week numbers, determines leap years,
// and provides period boundaries (start/end of week, month, quarter, year). The tool supports
// customizable week start days and timezone-aware date analysis.
func DateTimeInfo() agentDomain.Tool {
	// First define the parameter schema
	paramSchema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"date": {
				Type:        "string",
				Description: "Date to get information about (RFC3339 format preferred)",
			},
			"timezone": {
				Type:        "string",
				Description: "Timezone for the date (default: UTC)",
			},
			"week_start_day": {
				Type:        "integer",
				Description: "Week start day (0 = Sunday, 1 = Monday)",
				Minimum:     float64Ptr(0),
				Maximum:     float64Ptr(1),
			},
		},
		Required: []string{"date"},
	}

	// Define the output schema
	outputSchema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"date": {
				Type:        "string",
				Description: "Original date in RFC3339 format",
			},
			"day_of_week": {
				Type:        "integer",
				Description: "Numeric day of week (0 = Sunday, 6 = Saturday)",
			},
			"day_of_week_name": {
				Type:        "string",
				Description: "Day name (e.g., 'Monday')",
			},
			"day_of_week_iso": {
				Type:        "integer",
				Description: "ISO day of week (1 = Monday, 7 = Sunday)",
			},
			"day_of_month": {
				Type:        "integer",
				Description: "Day of the month (1-31)",
			},
			"day_of_year": {
				Type:        "integer",
				Description: "Day of the year (1-366)",
			},
			"week_number": {
				Type:        "integer",
				Description: "ISO week number (1-53)",
			},
			"week_year": {
				Type:        "integer",
				Description: "Year of the ISO week",
			},
			"month": {
				Type:        "integer",
				Description: "Month number (1-12)",
			},
			"month_name": {
				Type:        "string",
				Description: "Month name (e.g., 'January')",
			},
			"days_in_month": {
				Type:        "integer",
				Description: "Number of days in the month (28-31)",
			},
			"quarter": {
				Type:        "integer",
				Description: "Quarter number (1-4)",
			},
			"year": {
				Type:        "integer",
				Description: "Year",
			},
			"is_leap_year": {
				Type:        "boolean",
				Description: "Whether the year is a leap year",
			},
			"start_of_week": {
				Type:        "string",
				Description: "Start of week in RFC3339 format",
			},
			"end_of_week": {
				Type:        "string",
				Description: "End of week in RFC3339 format",
			},
			"start_of_month": {
				Type:        "string",
				Description: "Start of month in RFC3339 format",
			},
			"end_of_month": {
				Type:        "string",
				Description: "End of month in RFC3339 format",
			},
			"start_of_quarter": {
				Type:        "string",
				Description: "Start of quarter in RFC3339 format",
			},
			"end_of_quarter": {
				Type:        "string",
				Description: "End of quarter in RFC3339 format",
			},
			"start_of_year": {
				Type:        "string",
				Description: "Start of year in RFC3339 format",
			},
			"end_of_year": {
				Type:        "string",
				Description: "End of year in RFC3339 format",
			},
		},
		Required: []string{"date", "day_of_week", "day_of_month", "month", "year"},
	}

	builder := atools.NewToolBuilder("datetime_info", "Get comprehensive information about a specific date").
		WithFunction(dateTimeInfoExecute).
		WithParameterSchema(paramSchema).
		WithOutputSchema(outputSchema).
		WithUsageInstructions(`The datetime_info tool provides comprehensive information about a specific date, including:
- Day information: day of week, day of month, day of year
- Week information: ISO week number and week year
- Month information: month number, name, and days in month
- Quarter information
- Year information including leap year status
- Period boundaries: start/end of week, month, quarter, and year

The tool accepts dates in various formats but prefers RFC3339. It can work with different timezones and allows customizing the week start day.`).
		WithExamples([]agentDomain.ToolExample{
			{
				Name:        "Basic date info",
				Description: "Get information about a specific date",
				Input:       map[string]interface{}{"date": "2024-03-15T10:30:00Z"},
				Output: map[string]interface{}{
					"date":             "2024-03-15T10:30:00Z",
					"day_of_week":      5,
					"day_of_week_name": "Friday",
					"day_of_week_iso":  5,
					"day_of_month":     15,
					"day_of_year":      75,
					"month":            3,
					"month_name":       "March",
					"year":             2024,
					"is_leap_year":     true,
				},
			},
			{
				Name:        "Date info with timezone",
				Description: "Get date information in a specific timezone",
				Input:       map[string]interface{}{"date": "2024-07-04", "timezone": "America/New_York"},
				Output: map[string]interface{}{
					"date":             "2024-07-04T00:00:00-04:00",
					"day_of_week":      4,
					"day_of_week_name": "Thursday",
					"month_name":       "July",
					"quarter":          3,
				},
			},
			{
				Name:        "Monday week start",
				Description: "Get date info with Monday as the start of the week",
				Input:       map[string]interface{}{"date": "2024-12-25", "week_start_day": 1},
				Output: map[string]interface{}{
					"date":          "2024-12-25T00:00:00Z",
					"day_of_week":   3,
					"week_number":   52,
					"start_of_week": "2024-12-23T00:00:00Z",
					"end_of_week":   "2024-12-29T23:59:59.999999999Z",
				},
			},
			{
				Name:        "Leap year check",
				Description: "Check if a year is a leap year and get February info",
				Input:       map[string]interface{}{"date": "2024-02-29"},
				Output: map[string]interface{}{
					"date":          "2024-02-29T00:00:00Z",
					"is_leap_year":  true,
					"days_in_month": 29,
					"month_name":    "February",
				},
			},
			{
				Name:        "Quarter boundaries",
				Description: "Get quarter information and boundaries",
				Input:       map[string]interface{}{"date": "2024-05-15"},
				Output: map[string]interface{}{
					"quarter":          2,
					"start_of_quarter": "2024-04-01T00:00:00Z",
					"end_of_quarter":   "2024-06-30T23:59:59.999999999Z",
				},
			},
			{
				Name:        "ISO week numbering",
				Description: "Get ISO week information for edge cases",
				Input:       map[string]interface{}{"date": "2024-01-01"},
				Output: map[string]interface{}{
					"week_number": 1,
					"week_year":   2024,
					"day_of_year": 1,
				},
			},
			{
				Name:        "Component extraction",
				Description: "Extract all date components for analysis",
				Input:       map[string]interface{}{"date": "2024-09-30T14:45:30Z"},
				Output: map[string]interface{}{
					"day_of_month":  30,
					"day_of_year":   274,
					"days_in_month": 30,
					"end_of_month":  "2024-09-30T23:59:59.999999999Z",
					"start_of_year": "2024-01-01T00:00:00Z",
					"end_of_year":   "2024-12-31T23:59:59.999999999Z",
				},
			},
		}).
		WithConstraints([]string{
			"Date must be provided in a valid format (RFC3339 preferred)",
			"Timezone must be a valid IANA timezone name",
			"Week start day must be 0 (Sunday) or 1 (Monday)",
			"All period boundaries are calculated in the specified timezone",
			"ISO week numbering follows ISO 8601 standard",
		}).
		WithErrorGuidance(map[string]string{
			"invalid date format":  "Ensure the date is in a valid format. RFC3339 (e.g., '2024-03-15T10:30:00Z') is preferred, but common formats like '2024-03-15' are also accepted",
			"invalid timezone":     "Use a valid IANA timezone name like 'America/New_York', 'Europe/London', or 'UTC'",
			"week_start_day range": "Week start day must be 0 (Sunday) or 1 (Monday)",
			"parsing error":        "Check the date format. Common formats: RFC3339, YYYY-MM-DD, MM/DD/YYYY",
		}).
		WithCategory("datetime").
		WithTags([]string{"datetime", "calendar", "date-analysis", "timezone", "week", "month", "year", "iso8601"}).
		WithVersion("2.0.0").
		WithBehavior(true, false, false, "fast")

	return builder.Build()
}

// daysInMonth returns the number of days in the month of the given time
func daysInMonth(t time.Time) int {
	// Get the first day of next month, then subtract one day
	firstOfNextMonth := time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, t.Location())
	lastOfMonth := firstOfNextMonth.AddDate(0, 0, -1)
	return lastOfMonth.Day()
}

// startOfWeek returns the start of the week for the given time
func startOfWeek(t time.Time, weekStartDay time.Weekday) time.Time {
	// Calculate days to subtract to get to start of week
	currentWeekday := t.Weekday()
	daysToSubtract := int(currentWeekday - weekStartDay)
	if daysToSubtract < 0 {
		daysToSubtract += 7
	}

	startOfWeek := t.AddDate(0, 0, -daysToSubtract)
	return time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, startOfWeek.Location())
}

// endOfDay returns the end of the day (23:59:59.999999999) for the given time
func endOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// float64Ptr returns a pointer to a float64
func float64Ptr(f float64) *float64 {
	return &f
}

func init() {
	tools.MustRegisterTool("datetime_info", DateTimeInfo(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "datetime_info",
			Category:    "datetime",
			Tags:        []string{"datetime", "calendar", "date-analysis", "timezone", "week", "month", "year", "iso8601"},
			Description: "Get comprehensive information about a specific date including day/week/month/year info and period boundaries",
			Version:     "2.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Basic date info",
					Description: "Get information about a specific date",
					Code:        `DateTimeInfo().Execute(ctx, DateTimeInfoInput{Date: "2024-03-15T10:30:00Z"})`,
				},
				{
					Name:        "Date info with timezone",
					Description: "Get date information in a specific timezone",
					Code:        `DateTimeInfo().Execute(ctx, DateTimeInfoInput{Date: "2024-07-04", Timezone: "America/New_York"})`,
				},
				{
					Name:        "Monday week start",
					Description: "Get date info with Monday as the start of the week",
					Code:        `DateTimeInfo().Execute(ctx, DateTimeInfoInput{Date: "2024-12-25", WeekStartDay: 1})`,
				},
				{
					Name:        "Leap year check",
					Description: "Check if a year is a leap year",
					Code:        `DateTimeInfo().Execute(ctx, DateTimeInfoInput{Date: "2024-02-29"})`,
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
		UsageInstructions: `The datetime_info tool provides comprehensive information about a specific date, including:
- Day information: day of week, day of month, day of year
- Week information: ISO week number and week year
- Month information: month number, name, and days in month
- Quarter information
- Year information including leap year status
- Period boundaries: start/end of week, month, quarter, and year

The tool accepts dates in various formats but prefers RFC3339. It can work with different timezones and allows customizing the week start day.`,
		Constraints: []string{
			"Date must be provided in a valid format (RFC3339 preferred)",
			"Timezone must be a valid IANA timezone name",
			"Week start day must be 0 (Sunday) or 1 (Monday)",
			"All period boundaries are calculated in the specified timezone",
			"ISO week numbering follows ISO 8601 standard",
		},
		ErrorGuidance: map[string]string{
			"invalid date format":  "Ensure the date is in a valid format. RFC3339 (e.g., '2024-03-15T10:30:00Z') is preferred, but common formats like '2024-03-15' are also accepted",
			"invalid timezone":     "Use a valid IANA timezone name like 'America/New_York', 'Europe/London', or 'UTC'",
			"week_start_day range": "Week start day must be 0 (Sunday) or 1 (Monday)",
			"parsing error":        "Check the date format. Common formats: RFC3339, YYYY-MM-DD, MM/DD/YYYY",
		},
		IsDeterministic:      true,
		IsDestructive:        false,
		RequiresConfirmation: false,
		EstimatedLatency:     "instant",
	})
}
