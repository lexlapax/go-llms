// ABOUTME: Provides current date/time functionality in various formats
// ABOUTME: Supports UTC, local, and specific timezone outputs

package datetime

import (
	"fmt"
	"time"

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

// DateTimeNow returns a tool that gets the current date/time in various formats
func DateTimeNow() agentDomain.Tool {
	return atools.NewTool(
		"datetime_now",
		"Get current date/time in various formats and timezones",
		func(ctx *agentDomain.ToolContext, input DateTimeNowInput) (*DateTimeNowOutput, error) {
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
		},
		dateTimeNowParamSchema,
	)
}

// isLeapYear checks if a year is a leap year
func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

func init() {
	// Register the tool
	if err := registerTool("datetime_now", DateTimeNow()); err != nil {
		panic(fmt.Sprintf("Failed to register datetime_now tool: %v", err))
	}
}
