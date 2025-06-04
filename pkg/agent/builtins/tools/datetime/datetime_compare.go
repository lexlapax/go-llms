// ABOUTME: Provides date/time comparison functionality
// ABOUTME: Supports before/after/equal checks, same period comparisons, and sorting

package datetime

import (
	"fmt"
	"sort"
	"time"

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

var dateTimeCompareParamSchema = &schemaDomain.Schema{
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

// DateTimeCompare returns a tool that compares dates and times
func DateTimeCompare() agentDomain.Tool {
	return atools.NewTool(
		"datetime_compare",
		"Compare dates and times with various operations",
		func(ctx *agentDomain.ToolContext, input DateTimeCompareInput) (*DateTimeCompareOutput, error) {
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
		},
		dateTimeCompareParamSchema,
	)
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
	// Register the tool
	if err := registerTool("datetime_compare", DateTimeCompare()); err != nil {
		panic(fmt.Sprintf("Failed to register datetime_compare tool: %v", err))
	}
}
