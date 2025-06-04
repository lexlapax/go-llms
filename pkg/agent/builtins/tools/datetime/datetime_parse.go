// ABOUTME: Provides date/time parsing functionality with format detection
// ABOUTME: Supports common formats, custom layouts, relative dates, and validation

package datetime

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	agentDomain "github.com/lexlapax/go-llms/pkg/agent/domain"
	atools "github.com/lexlapax/go-llms/pkg/agent/tools"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// DateTimeParseInput defines the input for the datetime_parse tool
type DateTimeParseInput struct {
	// Date string to parse
	DateString string `json:"date_string"`
	// Custom format layout (Go time format)
	Format string `json:"format,omitempty"`
	// Timezone to use for parsing (default: UTC)
	Timezone string `json:"timezone,omitempty"`
	// AutoDetect format (default: true)
	AutoDetect bool `json:"auto_detect,omitempty"`
	// Reference time for relative dates (default: now)
	ReferenceTime string `json:"reference_time,omitempty"`
}

// DateTimeParseOutput defines the output for the datetime_parse tool
type DateTimeParseOutput struct {
	// Parsed date/time in RFC3339 format
	Parsed string `json:"parsed"`
	// Whether the date string is valid
	Valid bool `json:"valid"`
	// Detected format (if auto-detect was used)
	DetectedFormat string `json:"detected_format,omitempty"`
	// Unix timestamp
	UnixTimestamp int64 `json:"unix_timestamp"`
	// Validation errors (if any)
	ValidationErrors []string `json:"validation_errors,omitempty"`
}

var dateTimeParseParamSchema = &schemaDomain.Schema{
	Type: "object",
	Properties: map[string]schemaDomain.Property{
		"date_string": {
			Type:        "string",
			Description: "Date string to parse",
		},
		"format": {
			Type:        "string",
			Description: "Custom format layout (Go time format)",
		},
		"timezone": {
			Type:        "string",
			Description: "Timezone to use for parsing",
		},
		"auto_detect": {
			Type:        "boolean",
			Description: "Auto-detect format (default: true)",
		},
		"reference_time": {
			Type:        "string",
			Description: "Reference time for relative dates",
		},
	},
	Required: []string{"date_string"},
}

// Common date formats to try
var commonFormats = []struct {
	format string
	name   string
}{
	{time.RFC3339, "RFC3339"},
	{time.RFC3339Nano, "RFC3339Nano"},
	{time.RFC1123, "RFC1123"},
	{time.RFC1123Z, "RFC1123Z"},
	{time.RFC822, "RFC822"},
	{time.RFC822Z, "RFC822Z"},
	{time.RFC850, "RFC850"},
	{time.ANSIC, "ANSIC"},
	{time.UnixDate, "UnixDate"},
	{time.RubyDate, "RubyDate"},
	{time.Kitchen, "Kitchen"},
	{time.Stamp, "Stamp"},
	{time.StampMilli, "StampMilli"},
	{time.StampMicro, "StampMicro"},
	{time.StampNano, "StampNano"},
	{"2006-01-02", "ISO Date"},
	{"2006-01-02 15:04:05", "DateTime"},
	{"2006-01-02T15:04:05", "ISO DateTime"},
	{"01/02/2006", "US Date"},
	{"02/01/2006", "EU Date"},
	{"02-Jan-2006", "DD-Mon-YYYY"},
	{"Jan 2, 2006", "Mon D, YYYY"},
	{"January 2, 2006", "Month D, YYYY"},
	{"January 2, 2006 3:04 PM", "Month D, YYYY H:MM PM"},
	{"January 2, 2006 15:04:05", "Month D, YYYY HH:MM:SS"},
	{"2006/01/02", "YYYY/MM/DD"},
	{"20060102", "YYYYMMDD"},
	{"20060102150405", "YYYYMMDDHHmmss"},
	{"2006-01-02 15:04:05.000", "DateTime with millis"},
	{"2006-01-02 15:04:05 MST", "DateTime with timezone"},
	{"2006-01-02 15:04:05 -0700", "DateTime with offset"},
}

// DateTimeParse returns a tool that parses date/time strings
func DateTimeParse() agentDomain.Tool {
	return atools.NewTool(
		"datetime_parse",
		"Parse and validate date/time strings in various formats",
		func(ctx *agentDomain.ToolContext, input DateTimeParseInput) (*DateTimeParseOutput, error) {
			// Emit start event
			if ctx.Events != nil {
				ctx.Events.Emit(agentDomain.EventToolCall, agentDomain.ToolCallEventData{
					ToolName:   "datetime_parse",
					Parameters: input,
					RequestID:  ctx.RunID,
				})
			}

			output := &DateTimeParseOutput{
				Valid:            false,
				ValidationErrors: []string{},
			}

			// Set default for auto_detect if not specified
			if input.Format == "" && !input.AutoDetect {
				input.AutoDetect = true
			}

			// Get reference time for relative dates
			var referenceTime time.Time
			if input.ReferenceTime != "" {
				var err error
				referenceTime, err = time.Parse(time.RFC3339, input.ReferenceTime)
				if err != nil {
					return nil, fmt.Errorf("invalid reference time: %w", err)
				}
			} else {
				referenceTime = time.Now()
			}

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

			// Try to parse relative dates first
			if parsedTime, ok := parseRelativeDate(input.DateString, referenceTime); ok {
				output.Parsed = parsedTime.In(loc).Format(time.RFC3339)
				output.Valid = true
				output.DetectedFormat = "relative date"
				output.UnixTimestamp = parsedTime.Unix()

				// Emit result event
				if ctx.Events != nil {
					ctx.Events.Emit(agentDomain.EventToolResult, agentDomain.ToolResultEventData{
						ToolName:  "datetime_parse",
						Result:    output,
						RequestID: ctx.RunID,
					})
				}

				return output, nil
			}

			// If custom format is provided, try it first
			if input.Format != "" {
				if parsedTime, err := time.ParseInLocation(input.Format, input.DateString, loc); err == nil {
					output.Parsed = parsedTime.Format(time.RFC3339)
					output.Valid = true
					output.DetectedFormat = "custom format"
					output.UnixTimestamp = parsedTime.Unix()

					// Emit result event
					if ctx.Events != nil {
						ctx.Events.Emit(agentDomain.EventToolResult, agentDomain.ToolResultEventData{
							ToolName:  "datetime_parse",
							Result:    output,
							RequestID: ctx.RunID,
						})
					}

					return output, nil
				} else {
					output.ValidationErrors = append(output.ValidationErrors, fmt.Sprintf("Failed to parse with custom format: %v", err))
				}
			}

			// Auto-detect format
			if input.AutoDetect {
				for _, format := range commonFormats {
					if parsedTime, err := time.ParseInLocation(format.format, input.DateString, loc); err == nil {
						output.Parsed = parsedTime.Format(time.RFC3339)
						output.Valid = true
						output.DetectedFormat = format.name
						output.UnixTimestamp = parsedTime.Unix()

						// Emit result event
						if ctx.Events != nil {
							ctx.Events.Emit(agentDomain.EventToolResult, agentDomain.ToolResultEventData{
								ToolName:  "datetime_parse",
								Result:    output,
								RequestID: ctx.RunID,
							})
						}

						return output, nil
					}
				}

				// Try Unix timestamp
				if timestamp, err := strconv.ParseInt(input.DateString, 10, 64); err == nil {
					// Check if it's a reasonable timestamp (between 1970 and 2100)
					if timestamp > 0 && timestamp < 4102444800 { // seconds
						parsedTime := time.Unix(timestamp, 0).In(loc)
						output.Parsed = parsedTime.Format(time.RFC3339)
						output.Valid = true
						output.DetectedFormat = "Unix timestamp (seconds)"
						output.UnixTimestamp = timestamp

						// Emit result event
						if ctx.Events != nil {
							ctx.Events.Emit(agentDomain.EventToolResult, agentDomain.ToolResultEventData{
								ToolName:  "datetime_parse",
								Result:    output,
								RequestID: ctx.RunID,
							})
						}

						return output, nil
					} else if timestamp > 1000000000000 && timestamp < 4102444800000 { // milliseconds
						parsedTime := time.Unix(timestamp/1000, (timestamp%1000)*1000000).In(loc)
						output.Parsed = parsedTime.Format(time.RFC3339)
						output.Valid = true
						output.DetectedFormat = "Unix timestamp (milliseconds)"
						output.UnixTimestamp = parsedTime.Unix()

						// Emit result event
						if ctx.Events != nil {
							ctx.Events.Emit(agentDomain.EventToolResult, agentDomain.ToolResultEventData{
								ToolName:  "datetime_parse",
								Result:    output,
								RequestID: ctx.RunID,
							})
						}

						return output, nil
					}
				}
			}

			// If we couldn't parse the date
			output.ValidationErrors = append(output.ValidationErrors, "Unable to parse date string with any known format")

			// Emit result event
			if ctx.Events != nil {
				ctx.Events.Emit(agentDomain.EventToolResult, agentDomain.ToolResultEventData{
					ToolName:  "datetime_parse",
					Result:    output,
					RequestID: ctx.RunID,
				})
			}

			return output, nil
		},
		dateTimeParseParamSchema,
	)
}

// parseRelativeDate parses relative date strings like "yesterday", "tomorrow", "next Monday"
func parseRelativeDate(dateStr string, referenceTime time.Time) (time.Time, bool) {
	lower := strings.ToLower(strings.TrimSpace(dateStr))

	// Simple relative dates
	switch lower {
	case "now":
		return referenceTime, true
	case "today":
		return time.Date(referenceTime.Year(), referenceTime.Month(), referenceTime.Day(), 0, 0, 0, 0, referenceTime.Location()), true
	case "yesterday":
		return referenceTime.AddDate(0, 0, -1), true
	case "tomorrow":
		return referenceTime.AddDate(0, 0, 1), true
	}

	// Relative days pattern: "in X days", "X days ago"
	if matches := regexp.MustCompile(`^in (\d+) days?$`).FindStringSubmatch(lower); matches != nil {
		days, _ := strconv.Atoi(matches[1])
		return referenceTime.AddDate(0, 0, days), true
	}
	if matches := regexp.MustCompile(`^(\d+) days? ago$`).FindStringSubmatch(lower); matches != nil {
		days, _ := strconv.Atoi(matches[1])
		return referenceTime.AddDate(0, 0, -days), true
	}

	// Relative weeks pattern: "in X weeks", "X weeks ago"
	if matches := regexp.MustCompile(`^in (\d+) weeks?$`).FindStringSubmatch(lower); matches != nil {
		weeks, _ := strconv.Atoi(matches[1])
		return referenceTime.AddDate(0, 0, weeks*7), true
	}
	if matches := regexp.MustCompile(`^(\d+) weeks? ago$`).FindStringSubmatch(lower); matches != nil {
		weeks, _ := strconv.Atoi(matches[1])
		return referenceTime.AddDate(0, 0, -weeks*7), true
	}

	// Relative months pattern: "in X months", "X months ago"
	if matches := regexp.MustCompile(`^in (\d+) months?$`).FindStringSubmatch(lower); matches != nil {
		months, _ := strconv.Atoi(matches[1])
		return referenceTime.AddDate(0, months, 0), true
	}
	if matches := regexp.MustCompile(`^(\d+) months? ago$`).FindStringSubmatch(lower); matches != nil {
		months, _ := strconv.Atoi(matches[1])
		return referenceTime.AddDate(0, -months, 0), true
	}

	// Relative hours pattern: "in X hours", "X hours ago"
	if matches := regexp.MustCompile(`^in (\d+) hours?$`).FindStringSubmatch(lower); matches != nil {
		hours, _ := strconv.Atoi(matches[1])
		return referenceTime.Add(time.Duration(hours) * time.Hour), true
	}
	if matches := regexp.MustCompile(`^(\d+) hours? ago$`).FindStringSubmatch(lower); matches != nil {
		hours, _ := strconv.Atoi(matches[1])
		return referenceTime.Add(-time.Duration(hours) * time.Hour), true
	}

	// Next/last weekday pattern
	weekdays := map[string]time.Weekday{
		"sunday":    time.Sunday,
		"monday":    time.Monday,
		"tuesday":   time.Tuesday,
		"wednesday": time.Wednesday,
		"thursday":  time.Thursday,
		"friday":    time.Friday,
		"saturday":  time.Saturday,
	}

	if strings.HasPrefix(lower, "next ") {
		dayName := strings.TrimPrefix(lower, "next ")
		if weekday, ok := weekdays[dayName]; ok {
			return nextWeekday(referenceTime, weekday), true
		}
	}

	if strings.HasPrefix(lower, "last ") {
		dayName := strings.TrimPrefix(lower, "last ")
		if weekday, ok := weekdays[dayName]; ok {
			return previousWeekday(referenceTime, weekday), true
		}
	}

	// Special relative dates
	switch lower {
	case "next week":
		return referenceTime.AddDate(0, 0, 7), true
	case "last week":
		return referenceTime.AddDate(0, 0, -7), true
	case "next month":
		return referenceTime.AddDate(0, 1, 0), true
	case "last month":
		return referenceTime.AddDate(0, -1, 0), true
	case "next year":
		return referenceTime.AddDate(1, 0, 0), true
	case "last year":
		return referenceTime.AddDate(-1, 0, 0), true
	}

	return time.Time{}, false
}

func init() {
	// Register the tool
	if err := registerTool("datetime_parse", DateTimeParse()); err != nil {
		panic(fmt.Sprintf("Failed to register datetime_parse tool: %v", err))
	}
}
