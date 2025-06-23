// ABOUTME: Provides date/time arithmetic operations and calculations
// ABOUTME: Includes add/subtract, duration calculation, business days, age calculation

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

// DateTimeCalculateInput defines the input for the datetime_calculate tool
type DateTimeCalculateInput struct {
	// Operation to perform: "add", "subtract", "duration", "age", "next_weekday", "previous_weekday", "add_business_days", "subtract_business_days"
	Operation string `json:"operation"`
	// Start date/time (RFC3339 format preferred)
	StartDate string `json:"start_date"`
	// End date/time for duration calculations (RFC3339 format preferred)
	EndDate string `json:"end_date,omitempty"`
	// Unit for add/subtract: "years", "months", "days", "hours", "minutes", "seconds"
	Unit string `json:"unit,omitempty"`
	// Value to add/subtract
	Value int `json:"value,omitempty"`
	// Target weekday for next/previous operations (0 = Sunday, 6 = Saturday)
	TargetWeekday int `json:"target_weekday,omitempty"`
	// Timezone for calculations (default: UTC)
	Timezone string `json:"timezone,omitempty"`
	// Include weekends in business day calculations (default: false)
	IncludeWeekends bool `json:"include_weekends,omitempty"`
}

// DateTimeCalculateOutput defines the output for the datetime_calculate tool
type DateTimeCalculateOutput struct {
	// Result date/time in RFC3339 format
	Result string `json:"result,omitempty"`
	// Duration result (for duration operation)
	Duration *DurationInfo `json:"duration,omitempty"`
	// Age result (for age operation)
	Age *AgeInfo `json:"age,omitempty"`
	// Business days count (for business day operations)
	BusinessDays int `json:"business_days,omitempty"`
}

// DurationInfo holds duration information
type DurationInfo struct {
	TotalSeconds float64 `json:"total_seconds"`
	Days         int     `json:"days"`
	Hours        int     `json:"hours"`
	Minutes      int     `json:"minutes"`
	Seconds      int     `json:"seconds"`
	Milliseconds int     `json:"milliseconds"`
	// Human-readable format
	HumanReadable string `json:"human_readable"`
}

// AgeInfo holds age information
type AgeInfo struct {
	Years  int `json:"years"`
	Months int `json:"months"`
	Days   int `json:"days"`
	// Total days
	TotalDays int `json:"total_days"`
	// Human-readable format
	HumanReadable string `json:"human_readable"`
}

// DateTimeCalculate returns a tool that performs date/time arithmetic
// This tool provides various date/time calculations including adding/subtracting
// time periods, calculating durations between dates, age calculations,
// business day calculations, and finding next/previous weekdays.
func DateTimeCalculate() agentDomain.Tool {
	// Create output schema
	outputSchema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"result": {
				Type:        "string",
				Description: "Result date/time in RFC3339 format",
			},
			"duration": {
				Type:        "object",
				Description: "Duration result for duration operations",
				Properties: map[string]schemaDomain.Property{
					"total_seconds": {
						Type:        "number",
						Description: "Total seconds in the duration",
					},
					"days": {
						Type:        "integer",
						Description: "Number of days",
					},
					"hours": {
						Type:        "integer",
						Description: "Number of hours (0-23)",
					},
					"minutes": {
						Type:        "integer",
						Description: "Number of minutes (0-59)",
					},
					"seconds": {
						Type:        "integer",
						Description: "Number of seconds (0-59)",
					},
					"milliseconds": {
						Type:        "integer",
						Description: "Number of milliseconds (0-999)",
					},
					"human_readable": {
						Type:        "string",
						Description: "Human-readable format of the duration",
					},
				},
			},
			"age": {
				Type:        "object",
				Description: "Age result for age operations",
				Properties: map[string]schemaDomain.Property{
					"years": {
						Type:        "integer",
						Description: "Number of years",
					},
					"months": {
						Type:        "integer",
						Description: "Number of months (0-11)",
					},
					"days": {
						Type:        "integer",
						Description: "Number of days",
					},
					"total_days": {
						Type:        "integer",
						Description: "Total number of days",
					},
					"human_readable": {
						Type:        "string",
						Description: "Human-readable format of the age",
					},
				},
			},
			"business_days": {
				Type:        "integer",
				Description: "Business days count for business day operations",
			},
		},
	}

	builder := atools.NewToolBuilder("datetime_calculate", "Perform date/time arithmetic operations").
		WithFunction(dateTimeCalculateExecute).
		WithParameterSchema(&schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"operation": {
					Type:        "string",
					Description: "Operation to perform",
					Enum:        []string{"add", "subtract", "duration", "age", "next_weekday", "previous_weekday", "add_business_days", "subtract_business_days"},
				},
				"start_date": {
					Type:        "string",
					Description: "Start date/time (RFC3339 format preferred)",
				},
				"end_date": {
					Type:        "string",
					Description: "End date/time for duration calculations",
				},
				"unit": {
					Type:        "string",
					Description: "Unit for add/subtract operations",
					Enum:        []string{"years", "months", "days", "hours", "minutes", "seconds"},
				},
				"value": {
					Type:        "integer",
					Description: "Value to add/subtract",
				},
				"target_weekday": {
					Type:        "integer",
					Description: "Target weekday (0 = Sunday, 6 = Saturday)",
					Minimum:     float64Ptr(0),
					Maximum:     float64Ptr(6),
				},
				"timezone": {
					Type:        "string",
					Description: "Timezone for calculations",
				},
				"include_weekends": {
					Type:        "boolean",
					Description: "Include weekends in business day calculations",
				},
			},
			Required: []string{"operation", "start_date"},
		}).
		WithOutputSchema(outputSchema).
		WithUsageInstructions(`The datetime_calculate tool performs various date/time arithmetic operations:
- add/subtract: Add or subtract years, months, days, hours, minutes, or seconds
- duration: Calculate the duration between two dates
- age: Calculate age from birth date to current date or specified date
- next_weekday/previous_weekday: Find the next or previous occurrence of a specific weekday
- add_business_days/subtract_business_days: Add or subtract business days (excluding weekends)

All operations support timezone specification and use RFC3339 format for dates.`).
		WithExamples([]agentDomain.ToolExample{
			{
				Name:        "Add days",
				Description: "Add 30 days to a date",
				Scenario:    "When you need to calculate a future date by adding a specific number of days",
				Input: map[string]interface{}{
					"operation":  "add",
					"start_date": "2024-01-15",
					"unit":       "days",
					"value":      30,
				},
				Output: map[string]interface{}{
					"result": "2024-02-14T00:00:00Z",
				},
				Explanation: "Adds 30 days to January 15, 2024, resulting in February 14, 2024",
			},
			{
				Name:        "Calculate age",
				Description: "Calculate age from birth date",
				Scenario:    "When you need to determine someone's current age from their birth date",
				Input: map[string]interface{}{
					"operation":  "age",
					"start_date": "1990-05-15",
				},
				Output: map[string]interface{}{
					"age": map[string]interface{}{
						"years":          34,
						"months":         7,
						"days":           25,
						"total_days":     12653,
						"human_readable": "34 years, 7 months and 25 days",
					},
				},
				Explanation: "Calculates the age from birth date to current date, broken down into years, months, and days",
			},
			{
				Name:        "Duration between dates",
				Description: "Calculate duration between two dates",
				Scenario:    "When you need to know the time difference between two specific dates",
				Input: map[string]interface{}{
					"operation":  "duration",
					"start_date": "2024-01-01",
					"end_date":   "2024-12-31",
				},
				Output: map[string]interface{}{
					"duration": map[string]interface{}{
						"total_seconds":  31536000,
						"days":           365,
						"hours":          0,
						"minutes":        0,
						"seconds":        0,
						"milliseconds":   0,
						"human_readable": "365 days",
					},
				},
				Explanation: "Calculates the exact duration between January 1 and December 31, 2024",
			},
			{
				Name:        "Add business days",
				Description: "Add 10 business days to a date",
				Scenario:    "When you need to calculate a deadline excluding weekends",
				Input: map[string]interface{}{
					"operation":  "add_business_days",
					"start_date": "2024-03-01",
					"value":      10,
				},
				Output: map[string]interface{}{
					"result":        "2024-03-15T00:00:00Z",
					"business_days": 10,
				},
				Explanation: "Adds 10 business days (Monday-Friday) to March 1, 2024, skipping weekends",
			},
			{
				Name:        "Next weekday",
				Description: "Find next Monday from a given date",
				Scenario:    "When you need to find the next occurrence of a specific day of the week",
				Input: map[string]interface{}{
					"operation":      "next_weekday",
					"start_date":     "2024-03-15",
					"target_weekday": 1,
				},
				Output: map[string]interface{}{
					"result": "2024-03-18T00:00:00Z",
				},
				Explanation: "Finds the next Monday (weekday 1) after March 15, 2024 (Friday)",
			},
		}).
		WithConstraints([]string{
			"Date formats must be valid (RFC3339 preferred)",
			"Unit must be specified for add/subtract operations",
			"End date is required for duration calculations",
			"Target weekday must be 0-6 for weekday operations",
			"Business day calculations exclude weekends by default",
			"Timezone must be a valid IANA timezone name",
		}).
		WithErrorGuidance(map[string]string{
			"invalid start date":          "Ensure the start date is in a valid format. RFC3339 is preferred",
			"invalid end date":            "Ensure the end date is in a valid format for duration/age calculations",
			"unit is required":            "Specify a unit (years, months, days, hours, minutes, seconds) for add/subtract",
			"invalid unit":                "Use one of: years, months, days, hours, minutes, seconds",
			"invalid operation":           "Use one of: add, subtract, duration, age, next_weekday, previous_weekday, add_business_days, subtract_business_days",
			"end_date is required":        "Provide an end date for duration calculations",
			"invalid timezone":            "Use a valid IANA timezone name like 'America/New_York' or 'UTC'",
			"target weekday out of range": "Target weekday must be 0-6 (0=Sunday, 6=Saturday)",
		}).
		WithCategory("datetime").
		WithTags([]string{"datetime", "arithmetic", "duration", "age", "business-days", "weekday", "calendar"}).
		WithVersion("2.0.0").
		WithBehavior(true, false, false, "fast")

	return builder.Build()
}

// dateTimeCalculateExecute is the main execution logic
func dateTimeCalculateExecute(ctx *agentDomain.ToolContext, input DateTimeCalculateInput) (*DateTimeCalculateOutput, error) {
	// Emit start event
	if ctx.Events != nil {
		ctx.Events.Emit(agentDomain.EventToolCall, agentDomain.ToolCallEventData{
			ToolName:   "datetime_calculate",
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

	// Parse the start date
	startTime, err := parseDate(input.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}

	// Apply timezone if specified
	if input.Timezone != "" {
		loc, err := time.LoadLocation(input.Timezone)
		if err != nil {
			return nil, fmt.Errorf("invalid timezone: %w", err)
		}
		startTime = startTime.In(loc)
	}

	output := &DateTimeCalculateOutput{}

	switch input.Operation {
	case "add":
		if input.Unit == "" {
			return nil, fmt.Errorf("unit is required for add operation")
		}
		result, err := addTime(startTime, input.Unit, input.Value)
		if err != nil {
			return nil, err
		}
		output.Result = result.Format(time.RFC3339)

	case "subtract":
		if input.Unit == "" {
			return nil, fmt.Errorf("unit is required for subtract operation")
		}
		result, err := addTime(startTime, input.Unit, -input.Value)
		if err != nil {
			return nil, err
		}
		output.Result = result.Format(time.RFC3339)

	case "duration":
		if input.EndDate == "" {
			return nil, fmt.Errorf("end_date is required for duration operation")
		}
		endTime, err := parseDate(input.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end date: %w", err)
		}
		if input.Timezone != "" {
			loc, _ := time.LoadLocation(input.Timezone)
			endTime = endTime.In(loc)
		}
		output.Duration = calculateDuration(startTime, endTime)

	case "age":
		if input.EndDate == "" {
			// If no end date, use current time
			endTime := time.Now()
			if input.Timezone != "" {
				loc, _ := time.LoadLocation(input.Timezone)
				endTime = endTime.In(loc)
			}
			output.Age = calculateAge(startTime, endTime)
		} else {
			endTime, err := parseDate(input.EndDate)
			if err != nil {
				return nil, fmt.Errorf("invalid end date: %w", err)
			}
			if input.Timezone != "" {
				loc, _ := time.LoadLocation(input.Timezone)
				endTime = endTime.In(loc)
			}
			output.Age = calculateAge(startTime, endTime)
		}

	case "next_weekday":
		result := nextWeekday(startTime, time.Weekday(input.TargetWeekday))
		output.Result = result.Format(time.RFC3339)

	case "previous_weekday":
		result := previousWeekday(startTime, time.Weekday(input.TargetWeekday))
		output.Result = result.Format(time.RFC3339)

	case "add_business_days":
		result, businessDays := addBusinessDays(startTime, input.Value, input.IncludeWeekends)
		output.Result = result.Format(time.RFC3339)
		output.BusinessDays = businessDays

	case "subtract_business_days":
		result, businessDays := addBusinessDays(startTime, -input.Value, input.IncludeWeekends)
		output.Result = result.Format(time.RFC3339)
		output.BusinessDays = businessDays

	default:
		return nil, fmt.Errorf("invalid operation: %s", input.Operation)
	}

	// Emit result event
	if ctx.Events != nil {
		ctx.Events.Emit(agentDomain.EventToolResult, agentDomain.ToolResultEventData{
			ToolName:  "datetime_calculate",
			Result:    output,
			RequestID: ctx.RunID,
		})
	}

	return output, nil
}

// parseDate tries to parse a date string in various formats
func parseDate(dateStr string) (time.Time, error) {
	// Try RFC3339 first
	if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
		return t, nil
	}

	// Try other common formats
	formats := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		"01/02/2006",
		"02-Jan-2006",
		"Jan 2, 2006",
		"January 2, 2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// addTime adds a duration to a time based on unit
func addTime(t time.Time, unit string, value int) (time.Time, error) {
	switch unit {
	case "years":
		return t.AddDate(value, 0, 0), nil
	case "months":
		return t.AddDate(0, value, 0), nil
	case "days":
		return t.AddDate(0, 0, value), nil
	case "hours":
		return t.Add(time.Duration(value) * time.Hour), nil
	case "minutes":
		return t.Add(time.Duration(value) * time.Minute), nil
	case "seconds":
		return t.Add(time.Duration(value) * time.Second), nil
	default:
		return t, fmt.Errorf("invalid unit: %s", unit)
	}
}

// calculateDuration calculates the duration between two times
func calculateDuration(start, end time.Time) *DurationInfo {
	duration := end.Sub(start)
	totalSeconds := duration.Seconds()

	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60
	milliseconds := int(duration.Milliseconds()) % 1000

	// Create human-readable format
	var humanParts []string
	if days != 0 {
		humanParts = append(humanParts, fmt.Sprintf("%d days", days))
	}
	if hours != 0 {
		humanParts = append(humanParts, fmt.Sprintf("%d hours", hours))
	}
	if minutes != 0 {
		humanParts = append(humanParts, fmt.Sprintf("%d minutes", minutes))
	}
	if seconds != 0 || len(humanParts) == 0 {
		humanParts = append(humanParts, fmt.Sprintf("%d seconds", seconds))
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

	return &DurationInfo{
		TotalSeconds:  totalSeconds,
		Days:          days,
		Hours:         hours,
		Minutes:       minutes,
		Seconds:       seconds,
		Milliseconds:  milliseconds,
		HumanReadable: humanReadable,
	}
}

// calculateAge calculates age between two dates
func calculateAge(birthDate, currentDate time.Time) *AgeInfo {
	years := currentDate.Year() - birthDate.Year()
	months := int(currentDate.Month()) - int(birthDate.Month())
	days := currentDate.Day() - birthDate.Day()

	// Adjust for negative months
	if months < 0 {
		years--
		months += 12
	}

	// Adjust for negative days
	if days < 0 {
		months--
		if months < 0 {
			years--
			months += 12
		}
		// Get days in previous month
		prevMonth := currentDate.AddDate(0, -1, 0)
		days += daysInMonth(prevMonth)
	}

	// Calculate total days
	totalDays := int(currentDate.Sub(birthDate).Hours() / 24)

	// Create human-readable format
	var parts []string
	if years > 0 {
		parts = append(parts, fmt.Sprintf("%d years", years))
	}
	if months > 0 {
		parts = append(parts, fmt.Sprintf("%d months", months))
	}
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%d days", days))
	}

	humanReadable := ""
	for i, part := range parts {
		if i > 0 && i == len(parts)-1 {
			humanReadable += " and "
		} else if i > 0 {
			humanReadable += ", "
		}
		humanReadable += part
	}

	return &AgeInfo{
		Years:         years,
		Months:        months,
		Days:          days,
		TotalDays:     totalDays,
		HumanReadable: humanReadable,
	}
}

// nextWeekday finds the next occurrence of a specific weekday
func nextWeekday(from time.Time, weekday time.Weekday) time.Time {
	daysUntil := int(weekday) - int(from.Weekday())
	if daysUntil <= 0 {
		daysUntil += 7
	}
	return from.AddDate(0, 0, daysUntil)
}

// previousWeekday finds the previous occurrence of a specific weekday
func previousWeekday(from time.Time, weekday time.Weekday) time.Time {
	daysSince := int(from.Weekday()) - int(weekday)
	if daysSince <= 0 {
		daysSince += 7
	}
	return from.AddDate(0, 0, -daysSince)
}

// addBusinessDays adds business days to a date
func addBusinessDays(start time.Time, days int, includeWeekends bool) (time.Time, int) {
	if includeWeekends {
		return start.AddDate(0, 0, days), days
	}

	result := start
	businessDays := 0
	direction := 1
	if days < 0 {
		direction = -1
		days = -days
	}

	for businessDays < days {
		result = result.AddDate(0, 0, direction)
		weekday := result.Weekday()
		if weekday != time.Saturday && weekday != time.Sunday {
			businessDays++
		}
	}

	return result, businessDays * direction
}

func init() {
	tools.MustRegisterTool("datetime_calculate", DateTimeCalculate(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "datetime_calculate",
			Category:    "datetime",
			Tags:        []string{"datetime", "arithmetic", "duration", "age", "business-days", "weekday", "calendar"},
			Description: "Perform date/time arithmetic operations including add/subtract, duration, age, and business days",
			Version:     "2.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Add days",
					Description: "Add 30 days to a date",
					Code:        `DateTimeCalculate().Execute(ctx, DateTimeCalculateInput{Operation: "add", StartDate: "2024-01-15", Unit: "days", Value: 30})`,
				},
				{
					Name:        "Calculate age",
					Description: "Calculate age from birth date",
					Code:        `DateTimeCalculate().Execute(ctx, DateTimeCalculateInput{Operation: "age", StartDate: "1990-05-15"})`,
				},
				{
					Name:        "Duration between dates",
					Description: "Calculate duration between two dates",
					Code:        `DateTimeCalculate().Execute(ctx, DateTimeCalculateInput{Operation: "duration", StartDate: "2024-01-01", EndDate: "2024-12-31"})`,
				},
				{
					Name:        "Add business days",
					Description: "Add 10 business days to a date",
					Code:        `DateTimeCalculate().Execute(ctx, DateTimeCalculateInput{Operation: "add_business_days", StartDate: "2024-03-01", Value: 10})`,
				},
				{
					Name:        "Next weekday",
					Description: "Find next Monday from a given date",
					Code:        `DateTimeCalculate().Execute(ctx, DateTimeCalculateInput{Operation: "next_weekday", StartDate: "2024-03-15", TargetWeekday: 1})`,
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
		UsageInstructions: `The datetime_calculate tool performs various date/time arithmetic operations:
- add/subtract: Add or subtract years, months, days, hours, minutes, or seconds
- duration: Calculate the duration between two dates
- age: Calculate age from birth date to current date or specified date
- next_weekday/previous_weekday: Find the next or previous occurrence of a specific weekday
- add_business_days/subtract_business_days: Add or subtract business days (excluding weekends)

All operations support timezone specification and use RFC3339 format for dates.`,
		Constraints: []string{
			"Date formats must be valid (RFC3339 preferred)",
			"Unit must be specified for add/subtract operations",
			"End date is required for duration calculations",
			"Target weekday must be 0-6 for weekday operations",
			"Business day calculations exclude weekends by default",
			"Timezone must be a valid IANA timezone name",
		},
		ErrorGuidance: map[string]string{
			"invalid start date":          "Ensure the start date is in a valid format. RFC3339 is preferred",
			"invalid end date":            "Ensure the end date is in a valid format for duration/age calculations",
			"unit is required":            "Specify a unit (years, months, days, hours, minutes, seconds) for add/subtract",
			"invalid unit":                "Use one of: years, months, days, hours, minutes, seconds",
			"invalid operation":           "Use one of: add, subtract, duration, age, next_weekday, previous_weekday, add_business_days, subtract_business_days",
			"end_date is required":        "Provide an end date for duration calculations",
			"invalid timezone":            "Use a valid IANA timezone name like 'America/New_York' or 'UTC'",
			"target weekday out of range": "Target weekday must be 0-6 (0=Sunday, 6=Saturday)",
		},
		IsDeterministic:      true,
		IsDestructive:        false,
		RequiresConfirmation: false,
		EstimatedLatency:     "instant",
	})
}
