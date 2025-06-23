// ABOUTME: Provides current date/time functionality in various formats
// ABOUTME: Supports UTC, local, and specific timezone outputs

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

// DateTimeNowInput defines the input for the datetime_now tool
type DateTimeNowInput struct {
	// Timezone to get the current time in (e.g., "America/New_York", "Europe/London")
	// If empty, returns both UTC and local time
	Timezone string `json:"timezone,omitempty"`
	// IncludeComponents whether to include individual date/time components
	IncludeComponents bool `json:"include_components,omitempty"`
	// IncludeWeekInfo whether to include week-related information
	IncludeWeekInfo bool `json:"include_week_info,omitempty"`
	// IncludeTimestamps whether to include unix timestamps
	IncludeTimestamps bool `json:"include_timestamps,omitempty"`
	// Format custom format string (Go time format)
	Format string `json:"format,omitempty"`
}

// DateTimeNowOutput defines the output for the datetime_now tool
type DateTimeNowOutput struct {
	// UTC time in ISO 8601 format
	UTC string `json:"utc"`
	// Local time in ISO 8601 format
	Local string `json:"local"`
	// Time in requested timezone (if specified)
	Timezone string `json:"timezone,omitempty"`
	// Timezone name (if specified)
	TimezoneName string `json:"timezone_name,omitempty"`
	// Custom formatted output (if format specified)
	Formatted string `json:"formatted,omitempty"`
	// Date/time components (if requested)
	Components *DateTimeComponents `json:"components,omitempty"`
	// Week information (if requested)
	WeekInfo *WeekInfo `json:"week_info,omitempty"`
	// Unix timestamps (if requested)
	Timestamps *Timestamps `json:"timestamps,omitempty"`
}

// DateTimeComponents holds individual date/time components
type DateTimeComponents struct {
	Year        int    `json:"year"`
	Month       int    `json:"month"`
	MonthName   string `json:"month_name"`
	Day         int    `json:"day"`
	Hour        int    `json:"hour"`
	Minute      int    `json:"minute"`
	Second      int    `json:"second"`
	Nanosecond  int    `json:"nanosecond"`
	Weekday     int    `json:"weekday"` // 0 = Sunday
	WeekdayName string `json:"weekday_name"`
}

// WeekInfo holds week-related information
type WeekInfo struct {
	WeekNumber int  `json:"week_number"` // ISO week number
	DayOfWeek  int  `json:"day_of_week"` // 1 = Monday, 7 = Sunday (ISO)
	DayOfYear  int  `json:"day_of_year"` // 1-366
	Quarter    int  `json:"quarter"`     // 1-4
	IsLeapYear bool `json:"is_leap_year"`
}

// Timestamps holds various timestamp formats
type Timestamps struct {
	Unix      int64 `json:"unix"`       // Seconds since epoch
	UnixMilli int64 `json:"unix_milli"` // Milliseconds since epoch
	UnixMicro int64 `json:"unix_micro"` // Microseconds since epoch
	UnixNano  int64 `json:"unix_nano"`  // Nanoseconds since epoch
}

var dateTimeNowParamSchema = &schemaDomain.Schema{
	Type: "object",
	Properties: map[string]schemaDomain.Property{
		"timezone": {
			Type:        "string",
			Description: "Timezone to get current time in (e.g., 'America/New_York', 'Europe/London')",
		},
		"include_components": {
			Type:        "boolean",
			Description: "Include individual date/time components",
		},
		"include_week_info": {
			Type:        "boolean",
			Description: "Include week-related information",
		},
		"include_timestamps": {
			Type:        "boolean",
			Description: "Include unix timestamps",
		},
		"format": {
			Type:        "string",
			Description: "Custom format string (Go time format)",
		},
	},
}

// DateTimeNow returns a tool that gets the current date/time in various formats and timezones.
// It provides UTC and local time by default, supports any IANA timezone, includes optional components
// (year, month, day, etc.), week information (ISO week numbers), and unix timestamps in multiple units.
// The tool can format output using custom Go time layouts and integrates with state for default settings.
func DateTimeNow() agentDomain.Tool {
	// Create output schema
	outputSchema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"utc": {
				Type:        "string",
				Description: "UTC time in ISO 8601 format",
			},
			"local": {
				Type:        "string",
				Description: "Local time in ISO 8601 format",
			},
			"timezone": {
				Type:        "string",
				Description: "Time in requested timezone (if specified)",
			},
			"timezone_name": {
				Type:        "string",
				Description: "Timezone name (if specified)",
			},
			"formatted": {
				Type:        "string",
				Description: "Custom formatted output (if format specified)",
			},
			"components": {
				Type:        "object",
				Description: "Individual date/time components",
			},
			"week_info": {
				Type:        "object",
				Description: "Week-related information",
			},
			"timestamps": {
				Type:        "object",
				Description: "Unix timestamps in various units",
			},
		},
		Required: []string{"utc", "local"},
	}

	builder := atools.NewToolBuilder("datetime_now", "Get current date/time in various formats and timezones").
		WithFunction(dateTimeNowExecute).
		WithParameterSchema(dateTimeNowParamSchema).
		WithOutputSchema(outputSchema).
		WithUsageInstructions(`Use this tool to get the current date and time in various formats and timezones.

Basic Usage:
- Returns both UTC and local time by default
- Supports any valid IANA timezone (e.g., "America/New_York", "Europe/London", "Asia/Tokyo")
- Can include detailed components, week information, and timestamps

Timezone Support:
- Common US: America/New_York, America/Chicago, America/Denver, America/Los_Angeles
- Europe: Europe/London, Europe/Paris, Europe/Berlin, Europe/Moscow
- Asia: Asia/Tokyo, Asia/Shanghai, Asia/Kolkata, Asia/Dubai
- Pacific: Pacific/Auckland, Australia/Sydney
- Use empty string or omit for UTC and local time only

Format Options:
- Default: RFC3339 (ISO 8601) format
- Custom format using Go time format patterns:
  - "2006-01-02": Date only (YYYY-MM-DD)
  - "15:04:05": Time only (HH:MM:SS)
  - "Mon, 02 Jan 2006 15:04:05 MST": RFC1123 with timezone
  - "January 2, 2006": Human readable date
  - "3:04 PM": 12-hour time
  - "2006-01-02T15:04:05Z07:00": Full ISO 8601

Component Information (include_components):
- Year, Month (number and name), Day
- Hour (24-hour), Minute, Second, Nanosecond
- Weekday (number 0-6, Sunday=0) and name

Week Information (include_week_info):
- ISO week number (1-53)
- ISO day of week (1-7, Monday=1, Sunday=7)
- Day of year (1-366)
- Quarter (1-4)
- Leap year indicator

Timestamps (include_timestamps):
- Unix: Seconds since January 1, 1970 UTC
- UnixMilli: Milliseconds since epoch
- UnixMicro: Microseconds since epoch
- UnixNano: Nanoseconds since epoch

State Integration:
- datetime_default_timezone: Default timezone if not specified
- datetime_default_format: Default format string if not specified`).
		WithExamples([]agentDomain.ToolExample{
			{
				Name:        "Get current time in UTC and local",
				Description: "Basic usage without parameters",
				Scenario:    "When you need the current time in standard formats",
				Input:       map[string]interface{}{},
				Output: map[string]interface{}{
					"utc":   "2024-01-15T14:30:45Z",
					"local": "2024-01-15T09:30:45-05:00",
				},
				Explanation: "Returns current time in both UTC and system local timezone",
			},
			{
				Name:        "Get time in specific timezone",
				Description: "Request time in New York timezone",
				Scenario:    "When you need current time in a specific location",
				Input: map[string]interface{}{
					"timezone": "America/New_York",
				},
				Output: map[string]interface{}{
					"utc":           "2024-01-15T14:30:45Z",
					"local":         "2024-01-15T09:30:45-05:00",
					"timezone":      "2024-01-15T09:30:45-05:00",
					"timezone_name": "America/New_York",
				},
				Explanation: "Includes the requested timezone in addition to UTC and local",
			},
			{
				Name:        "Get time with components",
				Description: "Include individual date/time components",
				Scenario:    "When you need to work with specific parts of the date/time",
				Input: map[string]interface{}{
					"timezone":           "Europe/London",
					"include_components": true,
				},
				Output: map[string]interface{}{
					"utc":           "2024-01-15T14:30:45Z",
					"local":         "2024-01-15T09:30:45-05:00",
					"timezone":      "2024-01-15T14:30:45Z",
					"timezone_name": "Europe/London",
					"components": map[string]interface{}{
						"year":         2024,
						"month":        1,
						"month_name":   "January",
						"day":          15,
						"hour":         14,
						"minute":       30,
						"second":       45,
						"nanosecond":   0,
						"weekday":      1,
						"weekday_name": "Monday",
					},
				},
				Explanation: "Components are extracted from the timezone-specific time",
			},
			{
				Name:        "Get time with week information",
				Description: "Include week-related details",
				Scenario:    "When you need week numbers or day of year",
				Input: map[string]interface{}{
					"include_week_info": true,
				},
				Output: map[string]interface{}{
					"utc":   "2024-03-15T10:00:00Z",
					"local": "2024-03-15T06:00:00-04:00",
					"week_info": map[string]interface{}{
						"week_number":  11,
						"day_of_week":  5,
						"day_of_year":  75,
						"quarter":      1,
						"is_leap_year": true,
					},
				},
				Explanation: "ISO week numbering: week starts on Monday, first week contains January 4th",
			},
			{
				Name:        "Get time with custom format",
				Description: "Format time using custom pattern",
				Scenario:    "When you need time in a specific string format",
				Input: map[string]interface{}{
					"timezone": "Asia/Tokyo",
					"format":   "2006年01月02日 15:04:05",
				},
				Output: map[string]interface{}{
					"utc":           "2024-01-15T14:30:45Z",
					"local":         "2024-01-15T09:30:45-05:00",
					"timezone":      "2024-01-15T23:30:45+09:00",
					"timezone_name": "Asia/Tokyo",
					"formatted":     "2024年01月15日 23:30:45",
				},
				Explanation: "Go uses specific reference time for format patterns: Mon Jan 2 15:04:05 MST 2006",
			},
			{
				Name:        "Get Unix timestamps",
				Description: "Include various Unix timestamp formats",
				Scenario:    "When you need epoch-based timestamps for calculations",
				Input: map[string]interface{}{
					"include_timestamps": true,
				},
				Output: map[string]interface{}{
					"utc":   "2024-01-15T14:30:45Z",
					"local": "2024-01-15T09:30:45-05:00",
					"timestamps": map[string]interface{}{
						"unix":       int64(1705329045),
						"unix_milli": int64(1705329045000),
						"unix_micro": int64(1705329045000000),
						"unix_nano":  int64(1705329045000000000),
					},
				},
				Explanation: "Timestamps are always in UTC regardless of timezone settings",
			},
			{
				Name:        "Get all information",
				Description: "Request all available information",
				Scenario:    "When you need comprehensive date/time data",
				Input: map[string]interface{}{
					"timezone":           "Australia/Sydney",
					"include_components": true,
					"include_week_info":  true,
					"include_timestamps": true,
					"format":             "Monday, January 2, 2006 3:04 PM MST",
				},
				Output: map[string]interface{}{
					"utc":           "2024-07-15T04:30:45Z",
					"local":         "2024-07-15T00:30:45-04:00",
					"timezone":      "2024-07-15T14:30:45+10:00",
					"timezone_name": "Australia/Sydney",
					"formatted":     "Monday, July 15, 2024 2:30 PM AEST",
					"components": map[string]interface{}{
						"year":         2024,
						"month":        7,
						"month_name":   "July",
						"day":          15,
						"hour":         14,
						"minute":       30,
						"second":       45,
						"nanosecond":   0,
						"weekday":      1,
						"weekday_name": "Monday",
					},
					"week_info": map[string]interface{}{
						"week_number":  29,
						"day_of_week":  1,
						"day_of_year":  197,
						"quarter":      3,
						"is_leap_year": true,
					},
					"timestamps": map[string]interface{}{
						"unix":       int64(1721018445),
						"unix_milli": int64(1721018445000),
						"unix_micro": int64(1721018445000000),
						"unix_nano":  int64(1721018445000000000),
					},
				},
				Explanation: "All information is consistent with the same moment in time",
			},
		}).
		WithConstraints([]string{
			"Timezone must be a valid IANA timezone identifier",
			"Custom format uses Go's time format syntax (reference time: Mon Jan 2 15:04:05 MST 2006)",
			"Weekday numbers: Sunday=0, Monday=1, ..., Saturday=6",
			"ISO week numbering: Monday=1, Sunday=7",
			"Week 1 is the first week containing January 4th",
			"Unix timestamps are always in UTC",
			"Nanosecond precision may vary by system",
			"Some timezone abbreviations may be ambiguous (use full names)",
		}).
		WithErrorGuidance(map[string]string{
			"invalid timezone":  "The timezone identifier is not recognized. Use IANA timezone names like 'America/New_York' or 'Europe/London'",
			"unknown time zone": "The specified timezone does not exist. Check the spelling and use format 'Continent/City'",
		}).
		WithCategory("datetime").
		WithTags([]string{"datetime", "time", "timezone", "clock", "timestamp", "calendar"}).
		WithVersion("2.0.0").
		WithBehavior(false, false, false, "fast") // Non-deterministic due to current time

	return builder.Build()
}

// dateTimeNowExecute is the main execution logic
func dateTimeNowExecute(ctx *agentDomain.ToolContext, input DateTimeNowInput) (*DateTimeNowOutput, error) {
	// Emit start event
	if ctx.Events != nil {
		ctx.Events.Emit(agentDomain.EventToolCall, agentDomain.ToolCallEventData{
			ToolName:   "datetime_now",
			Parameters: input,
			RequestID:  ctx.RunID,
		})
	}

	// Check state for default timezone if not provided
	if input.Timezone == "" && ctx.State != nil {
		if val, ok := ctx.State.Get("datetime_default_timezone"); ok {
			if tz, ok := val.(string); ok && tz != "" {
				input.Timezone = tz
			}
		}
	}

	// Check state for default format if not provided
	if input.Format == "" && ctx.State != nil {
		if val, ok := ctx.State.Get("datetime_default_format"); ok {
			if fmt, ok := val.(string); ok && fmt != "" {
				input.Format = fmt
			}
		}
	}

	now := time.Now()

	output := &DateTimeNowOutput{
		UTC:   now.UTC().Format(time.RFC3339),
		Local: now.Local().Format(time.RFC3339),
	}

	// Handle specific timezone if requested
	if input.Timezone != "" {
		loc, err := time.LoadLocation(input.Timezone)
		if err != nil {
			return nil, fmt.Errorf("invalid timezone: %w", err)
		}
		tzTime := now.In(loc)
		output.Timezone = tzTime.Format(time.RFC3339)
		output.TimezoneName = input.Timezone
	}

	// Handle custom format
	if input.Format != "" {
		targetTime := now
		if input.Timezone != "" {
			loc, _ := time.LoadLocation(input.Timezone)
			targetTime = now.In(loc)
		}
		output.Formatted = targetTime.Format(input.Format)
	}

	// Include components if requested
	if input.IncludeComponents {
		targetTime := now
		if input.Timezone != "" {
			loc, _ := time.LoadLocation(input.Timezone)
			targetTime = now.In(loc)
		}
		output.Components = &DateTimeComponents{
			Year:        targetTime.Year(),
			Month:       int(targetTime.Month()),
			MonthName:   targetTime.Month().String(),
			Day:         targetTime.Day(),
			Hour:        targetTime.Hour(),
			Minute:      targetTime.Minute(),
			Second:      targetTime.Second(),
			Nanosecond:  targetTime.Nanosecond(),
			Weekday:     int(targetTime.Weekday()),
			WeekdayName: targetTime.Weekday().String(),
		}
	}

	// Include week info if requested
	if input.IncludeWeekInfo {
		targetTime := now
		if input.Timezone != "" {
			loc, _ := time.LoadLocation(input.Timezone)
			targetTime = now.In(loc)
		}

		year, week := targetTime.ISOWeek()
		isLeapYear := isLeapYear(year)
		quarter := (int(targetTime.Month())-1)/3 + 1
		dayOfYear := targetTime.YearDay()

		// ISO day of week: Monday = 1, Sunday = 7
		isoDayOfWeek := int(targetTime.Weekday())
		if isoDayOfWeek == 0 {
			isoDayOfWeek = 7
		}

		output.WeekInfo = &WeekInfo{
			WeekNumber: week,
			DayOfWeek:  isoDayOfWeek,
			DayOfYear:  dayOfYear,
			Quarter:    quarter,
			IsLeapYear: isLeapYear,
		}
	}

	// Include timestamps if requested
	if input.IncludeTimestamps {
		output.Timestamps = &Timestamps{
			Unix:      now.Unix(),
			UnixMilli: now.UnixMilli(),
			UnixMicro: now.UnixMicro(),
			UnixNano:  now.UnixNano(),
		}
	}

	// Emit result event
	if ctx.Events != nil {
		ctx.Events.Emit(agentDomain.EventToolResult, agentDomain.ToolResultEventData{
			ToolName:  "datetime_now",
			Result:    output,
			RequestID: ctx.RunID,
		})
	}

	return output, nil
}

// isLeapYear checks if a year is a leap year
func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

func init() {
	tools.MustRegisterTool("datetime_now", DateTimeNow(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "datetime_now",
			Category:    "datetime",
			Tags:        []string{"datetime", "current-time", "timezone", "now", "timestamp", "utc", "local"},
			Description: "Get current date/time in various formats with timezone support",
			Version:     "2.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Basic current time",
					Description: "Get current time in UTC",
					Code:        `DateTimeNow().Execute(ctx, DateTimeNowInput{})`,
				},
				{
					Name:        "Time with timezone",
					Description: "Get current time in specific timezone",
					Code:        `DateTimeNow().Execute(ctx, DateTimeNowInput{Timezone: "America/New_York"})`,
				},
				{
					Name:        "Time with components",
					Description: "Get current time with date/time components",
					Code:        `DateTimeNow().Execute(ctx, DateTimeNowInput{IncludeComponents: true})`,
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
		UsageInstructions: `The datetime_now tool returns the current date and time with various options:
- Timezone support (IANA timezone names)
- Multiple format options (RFC3339, custom formats)
- Optional components (year, month, day, hour, minute, second)
- Week information (ISO week number, day of week)
- Unix timestamps (seconds, milliseconds, microseconds, nanoseconds)

The tool can be used to get the current time in any timezone and format.`,
		Constraints: []string{
			"Timezone must be a valid IANA timezone name",
			"Custom format must use Go time layout syntax",
			"Output is non-deterministic (current time)",
			"All timestamps are calculated at the moment of execution",
		},
		ErrorGuidance: map[string]string{
			"invalid timezone": "Use a valid IANA timezone name like 'America/New_York', 'Europe/London', or 'UTC'",
			"invalid format":   "Custom format must use Go time layout syntax (e.g., '2006-01-02 15:04:05')",
		},
		IsDeterministic:      false,
		IsDestructive:        false,
		RequiresConfirmation: false,
		EstimatedLatency:     "fast",
	})
}
