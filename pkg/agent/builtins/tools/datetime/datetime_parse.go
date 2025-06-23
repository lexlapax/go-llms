// ABOUTME: Provides date/time parsing functionality with format detection
// ABOUTME: Supports common formats, custom layouts, relative dates, and validation

package datetime

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
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

// dateTimeParseExecute is the execution function for datetime_parse
func dateTimeParseExecute(ctx *agentDomain.ToolContext, input DateTimeParseInput) (*DateTimeParseOutput, error) {
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
}

// DateTimeParse returns a tool that parses and validates date/time strings in various formats.
// It features automatic format detection for 30+ common formats, custom format support using Go time layouts,
// relative date parsing ("tomorrow", "next Monday", "in 3 days"), and unix timestamp recognition.
// The tool provides validation feedback and can parse dates in specific timezones.
func DateTimeParse() agentDomain.Tool {
	// Define parameter schema
	paramSchema := &schemaDomain.Schema{
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

	// Define output schema
	outputSchema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"parsed": {
				Type:        "string",
				Description: "Parsed date/time in RFC3339 format",
			},
			"valid": {
				Type:        "boolean",
				Description: "Whether the date string is valid",
			},
			"detected_format": {
				Type:        "string",
				Description: "Detected format (if auto-detect was used)",
			},
			"unix_timestamp": {
				Type:        "integer",
				Description: "Unix timestamp",
			},
			"validation_errors": {
				Type:        "array",
				Description: "Validation errors (if any)",
				Items: &schemaDomain.Property{
					Type: "string",
				},
			},
		},
		Required: []string{"parsed", "valid", "unix_timestamp"},
	}

	builder := atools.NewToolBuilder("datetime_parse", "Parse and validate date/time strings in various formats").
		WithFunction(dateTimeParseExecute).
		WithParameterSchema(paramSchema).
		WithOutputSchema(outputSchema).
		WithUsageInstructions(`The datetime_parse tool intelligently parses date/time strings from various formats:

Format Detection:
- Automatic detection of 30+ common date formats (ISO, RFC, US, EU, etc.)
- Custom format specification using Go time layout patterns
- Unix timestamp support (both seconds and milliseconds)
- Relative date parsing (today, tomorrow, yesterday, next Monday, etc.)

Go Time Format Reference:
- Year: "2006" (4 digits), "06" (2 digits)
- Month: "01" or "1" (number), "Jan" (short name), "January" (full name)
- Day: "02" or "2" (day of month), "Mon" (weekday short), "Monday" (weekday full)
- Hour: "15" (24-hour), "03" or "3" (12-hour)
- Minute: "04" or "4"
- Second: "05" or "5"
- AM/PM: "PM" or "pm"
- Timezone: "MST" (name), "-0700" (numeric offset), "Z07:00" (ISO 8601)

Common Format Examples:
- "2006-01-02": ISO date (YYYY-MM-DD)
- "02/01/2006": EU date (DD/MM/YYYY)
- "01/02/2006": US date (MM/DD/YYYY)
- "2006-01-02 15:04:05": Date time with 24-hour format
- "Jan 2, 2006 3:04 PM": Human readable with 12-hour format
- "20060102": Compact date (YYYYMMDD)
- "20060102150405": Compact datetime

Relative Date Support:
- Simple: now, today, yesterday, tomorrow
- Days: "in 3 days", "5 days ago"
- Weeks: "in 2 weeks", "1 week ago", "next week", "last week"
- Months: "in 6 months", "2 months ago", "next month", "last month"
- Years: "next year", "last year"
- Hours: "in 4 hours", "3 hours ago"
- Weekdays: "next Monday", "last Friday"

State Integration:
- default_timezone: Default timezone if not specified in input

Auto-detection tries formats in order of specificity to ensure accurate parsing.`).
		WithExamples([]agentDomain.ToolExample{
			{
				Name:        "Parse ISO date",
				Description: "Parse a standard ISO formatted date",
				Scenario:    "When you have a date in ISO 8601 format",
				Input: map[string]interface{}{
					"date_string": "2024-03-15",
				},
				Output: map[string]interface{}{
					"parsed":          "2024-03-15T00:00:00Z",
					"valid":           true,
					"detected_format": "ISO Date",
					"unix_timestamp":  int64(1710460800),
				},
				Explanation: "Automatically detects and parses ISO date format, assuming midnight UTC",
			},
			{
				Name:        "Parse with custom format",
				Description: "Parse using a specific date format",
				Scenario:    "When you know the exact format of your date string",
				Input: map[string]interface{}{
					"date_string": "15/03/2024 14:30",
					"format":      "02/01/2006 15:04",
				},
				Output: map[string]interface{}{
					"parsed":          "2024-03-15T14:30:00Z",
					"valid":           true,
					"detected_format": "custom format",
					"unix_timestamp":  int64(1710512400),
				},
				Explanation: "Uses the provided Go time format to parse the date precisely",
			},
			{
				Name:        "Parse relative date",
				Description: "Parse natural language relative dates",
				Scenario:    "When working with human-friendly date descriptions",
				Input: map[string]interface{}{
					"date_string": "tomorrow",
				},
				Output: map[string]interface{}{
					"parsed":          "2024-03-16T00:00:00Z",
					"valid":           true,
					"detected_format": "relative date",
					"unix_timestamp":  int64(1710547200),
				},
				Explanation: "Interprets 'tomorrow' relative to current date (assuming today is March 15, 2024)",
			},
			{
				Name:        "Parse Unix timestamp",
				Description: "Convert Unix timestamp to date",
				Scenario:    "When working with system-generated timestamps",
				Input: map[string]interface{}{
					"date_string": "1710460800",
				},
				Output: map[string]interface{}{
					"parsed":          "2024-03-15T00:00:00Z",
					"valid":           true,
					"detected_format": "Unix timestamp (seconds)",
					"unix_timestamp":  int64(1710460800),
				},
				Explanation: "Recognizes numeric string as Unix timestamp and converts to RFC3339",
			},
			{
				Name:        "Parse with timezone",
				Description: "Parse date in specific timezone",
				Scenario:    "When the date should be interpreted in a specific timezone",
				Input: map[string]interface{}{
					"date_string": "2024-03-15 15:30:00",
					"timezone":    "America/New_York",
				},
				Output: map[string]interface{}{
					"parsed":          "2024-03-15T15:30:00-04:00",
					"valid":           true,
					"detected_format": "DateTime",
					"unix_timestamp":  int64(1710530400),
				},
				Explanation: "Parses the date assuming it's in New York timezone (EDT in March)",
			},
			{
				Name:        "Parse complex relative date",
				Description: "Parse relative date with reference time",
				Scenario:    "When calculating dates relative to a specific point in time",
				Input: map[string]interface{}{
					"date_string":    "next Monday",
					"reference_time": "2024-03-15T10:00:00Z",
				},
				Output: map[string]interface{}{
					"parsed":          "2024-03-18T00:00:00Z",
					"valid":           true,
					"detected_format": "relative date",
					"unix_timestamp":  int64(1710720000),
				},
				Explanation: "Finds the next Monday after March 15, 2024 (which is March 18)",
			},
			{
				Name:        "Parse ambiguous date with auto-detect",
				Description: "Let the tool detect ambiguous date format",
				Scenario:    "When the date format is unclear (DD/MM vs MM/DD)",
				Input: map[string]interface{}{
					"date_string": "03/04/2024",
					"auto_detect": true,
				},
				Output: map[string]interface{}{
					"parsed":          "2024-03-04T00:00:00Z",
					"valid":           true,
					"detected_format": "US Date",
					"unix_timestamp":  int64(1709510400),
				},
				Explanation: "Auto-detection prefers US format (MM/DD/YYYY) when ambiguous",
			},
			{
				Name:        "Parse with validation errors",
				Description: "Handle invalid date string",
				Scenario:    "When the date string cannot be parsed",
				Input: map[string]interface{}{
					"date_string": "not a date",
				},
				Output: map[string]interface{}{
					"parsed":            "",
					"valid":             false,
					"detected_format":   "",
					"unix_timestamp":    int64(0),
					"validation_errors": []string{"Unable to parse date string with any known format"},
				},
				Explanation: "Returns validation errors when the string doesn't match any known format",
			},
		}).
		WithConstraints([]string{
			"Date string must be provided",
			"Custom format must use Go time layout syntax",
			"Timezone must be a valid IANA timezone name",
			"Reference time must be in RFC3339 format",
			"Unix timestamps must be reasonable (between 1970 and 2100)",
			"Auto-detection tries formats in order of specificity",
			"Relative dates are calculated from reference time or current time",
			"Ambiguous dates (01/02/03) may be parsed differently than expected",
		}).
		WithErrorGuidance(map[string]string{
			"invalid reference time": "Reference time must be in RFC3339 format (e.g., '2024-03-15T10:30:00Z')",
			"invalid timezone":       "Use a valid IANA timezone name like 'America/New_York' or 'UTC'",
			"parse failure":          "Check the date format. Try providing a custom format or ensure auto_detect is enabled",
			"custom format error":    "Ensure the format uses Go time layout (e.g., '2006-01-02' for YYYY-MM-DD)",
			"ambiguous format":       "For dates like '01/02/03', specify a custom format to avoid ambiguity",
			"timestamp out of range": "Unix timestamps should be between 0 and 4102444800 (year 2100)",
		}).
		WithCategory("datetime").
		WithTags([]string{"datetime", "parse", "validation", "format-detection", "relative-dates", "timestamp", "timezone"}).
		WithVersion("2.0.0").
		WithBehavior(false, false, false, "fast") // Non-deterministic due to relative dates

	return builder.Build()
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
	tools.MustRegisterTool("datetime_parse", DateTimeParse(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "datetime_parse",
			Category:    "datetime",
			Tags:        []string{"datetime", "parse", "validation", "format-detection", "relative-dates", "timestamp"},
			Description: "Parse and validate date/time strings with automatic format detection and relative date support",
			Version:     "2.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Parse ISO date",
					Description: "Parse an ISO formatted date",
					Code:        `DateTimeParse().Execute(ctx, DateTimeParseInput{DateString: "2024-03-15"})`,
				},
				{
					Name:        "Parse with custom format",
					Description: "Parse using a custom Go time format",
					Code:        `DateTimeParse().Execute(ctx, DateTimeParseInput{DateString: "15/03/2024", Format: "02/01/2006"})`,
				},
				{
					Name:        "Parse relative date",
					Description: "Parse a relative date like 'tomorrow'",
					Code:        `DateTimeParse().Execute(ctx, DateTimeParseInput{DateString: "tomorrow"})`,
				},
				{
					Name:        "Parse Unix timestamp",
					Description: "Parse a Unix timestamp",
					Code:        `DateTimeParse().Execute(ctx, DateTimeParseInput{DateString: "1710460800"})`,
				},
				{
					Name:        "Parse with timezone",
					Description: "Parse a date in a specific timezone",
					Code:        `DateTimeParse().Execute(ctx, DateTimeParseInput{DateString: "2024-03-15 15:30:00", Timezone: "America/New_York"})`,
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
		UsageInstructions: `The datetime_parse tool parses date/time strings in various formats:
- Automatic format detection for common date formats
- Support for custom format specifications using Go time layouts
- Relative date parsing (today, tomorrow, yesterday, next Monday, in 3 days, etc.)
- Unix timestamp parsing (both seconds and milliseconds)
- Timezone-aware parsing
- Validation with detailed error messages

The tool tries multiple formats automatically unless a specific format is provided.`,
		Constraints: []string{
			"Date string must be provided",
			"Custom format must use Go time layout syntax",
			"Timezone must be a valid IANA timezone name",
			"Reference time must be in RFC3339 format",
			"Unix timestamps must be reasonable (between 1970 and 2100)",
		},
		ErrorGuidance: map[string]string{
			"invalid reference time": "Reference time must be in RFC3339 format (e.g., '2024-03-15T10:30:00Z')",
			"invalid timezone":       "Use a valid IANA timezone name like 'America/New_York' or 'UTC'",
			"parse failure":          "Check the date format. Try providing a custom format or ensure auto_detect is enabled",
			"custom format error":    "Ensure the format uses Go time layout (e.g., '2006-01-02' for YYYY-MM-DD)",
		},
		IsDeterministic:      false, // Because relative dates depend on current time
		IsDestructive:        false,
		RequiresConfirmation: false,
		EstimatedLatency:     "instant",
	})
}
