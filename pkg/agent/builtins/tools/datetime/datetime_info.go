// ABOUTME: Provides date information functionality for analyzing dates
// ABOUTME: Includes day/week/month/year info, leap year check, period boundaries

package datetime

import (
	"fmt"
	"time"

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

var dateTimeInfoParamSchema = &schemaDomain.Schema{
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

// DateTimeInfo returns a tool that gets information about a specific date
func DateTimeInfo() agentDomain.Tool {
	return atools.NewTool(
		"datetime_info",
		"Get comprehensive information about a specific date",
		func(ctx *agentDomain.ToolContext, input DateTimeInfoInput) (*DateTimeInfoOutput, error) {
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
		},
		dateTimeInfoParamSchema,
	)
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
	// Register the tool
	if err := registerTool("datetime_info", DateTimeInfo()); err != nil {
		panic(fmt.Sprintf("Failed to register datetime_info tool: %v", err))
	}
}
