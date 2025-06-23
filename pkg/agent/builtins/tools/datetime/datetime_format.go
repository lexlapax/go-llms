// ABOUTME: Provides date/time formatting functionality with multiple formats
// ABOUTME: Supports standard formats, custom layouts, localized formatting, and relative time

package datetime

import (
	"fmt"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	agentDomain "github.com/lexlapax/go-llms/pkg/agent/domain"
	atools "github.com/lexlapax/go-llms/pkg/agent/tools"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// DateTimeFormatInput defines the input for the datetime_format tool
type DateTimeFormatInput struct {
	// Date/time to format (RFC3339 format preferred)
	DateTime string `json:"datetime"`
	// Format type: "standard", "custom", "relative", "multiple"
	FormatType string `json:"format_type,omitempty"`
	// Standard format name (for format_type="standard")
	StandardFormat string `json:"standard_format,omitempty"`
	// Custom format layout (for format_type="custom")
	CustomFormat string `json:"custom_format,omitempty"`
	// Multiple formats to output (for format_type="multiple")
	Formats []string `json:"formats,omitempty"`
	// Timezone for formatting (default: UTC)
	Timezone string `json:"timezone,omitempty"`
	// Include weekday name in relative format
	IncludeWeekday bool `json:"include_weekday,omitempty"`
	// Localization options
	Locale string `json:"locale,omitempty"`
}

// DateTimeFormatOutput defines the output for the datetime_format tool
type DateTimeFormatOutput struct {
	// Formatted date/time string
	Formatted string `json:"formatted,omitempty"`
	// Multiple formatted outputs (if requested)
	MultipleFormats map[string]string `json:"multiple_formats,omitempty"`
	// Relative time description
	RelativeTime string `json:"relative_time,omitempty"`
	// Localized components (if locale specified)
	Localized *LocalizedComponents `json:"localized,omitempty"`
}

// LocalizedComponents holds localized date/time components
type LocalizedComponents struct {
	MonthName      string `json:"month_name"`
	MonthNameShort string `json:"month_name_short"`
	WeekdayName    string `json:"weekday_name"`
	WeekdayShort   string `json:"weekday_short"`
	Period         string `json:"period,omitempty"` // AM/PM
}

// Standard format mappings
var standardFormats = map[string]string{
	"RFC3339":  time.RFC3339,
	"RFC1123":  time.RFC1123,
	"RFC822":   time.RFC822,
	"Kitchen":  time.Kitchen,
	"Stamp":    time.Stamp,
	"ISO8601":  "2006-01-02T15:04:05Z07:00",
	"UnixDate": time.UnixDate,
}

// Locale data (simplified - in production, use proper i18n library)
var localeData = map[string]struct {
	months        []string
	monthsShort   []string
	weekdays      []string
	weekdaysShort []string
}{
	"es": {
		months:        []string{"", "enero", "febrero", "marzo", "abril", "mayo", "junio", "julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre"},
		monthsShort:   []string{"", "ene", "feb", "mar", "abr", "may", "jun", "jul", "ago", "sep", "oct", "nov", "dic"},
		weekdays:      []string{"domingo", "lunes", "martes", "miércoles", "jueves", "viernes", "sábado"},
		weekdaysShort: []string{"dom", "lun", "mar", "mié", "jue", "vie", "sáb"},
	},
	"fr": {
		months:        []string{"", "janvier", "février", "mars", "avril", "mai", "juin", "juillet", "août", "septembre", "octobre", "novembre", "décembre"},
		monthsShort:   []string{"", "jan", "fév", "mar", "avr", "mai", "juin", "juil", "août", "sep", "oct", "nov", "déc"},
		weekdays:      []string{"dimanche", "lundi", "mardi", "mercredi", "jeudi", "vendredi", "samedi"},
		weekdaysShort: []string{"dim", "lun", "mar", "mer", "jeu", "ven", "sam"},
	},
	"de": {
		months:        []string{"", "Januar", "Februar", "März", "April", "Mai", "Juni", "Juli", "August", "September", "Oktober", "November", "Dezember"},
		monthsShort:   []string{"", "Jan", "Feb", "Mär", "Apr", "Mai", "Jun", "Jul", "Aug", "Sep", "Okt", "Nov", "Dez"},
		weekdays:      []string{"Sonntag", "Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag", "Samstag"},
		weekdaysShort: []string{"So", "Mo", "Di", "Mi", "Do", "Fr", "Sa"},
	},
}

// dateTimeFormatExecute is the execution function for datetime_format
func dateTimeFormatExecute(ctx *agentDomain.ToolContext, input DateTimeFormatInput) (*DateTimeFormatOutput, error) {
	// Emit start event
	if ctx.Events != nil {
		ctx.Events.Emit(agentDomain.EventToolCall, agentDomain.ToolCallEventData{
			ToolName:   "datetime_format",
			Parameters: input,
			RequestID:  ctx.RunID,
		})
	}

	// Parse the input date/time
	parsedTime, err := parseDate(input.DateTime)
	if err != nil {
		return nil, fmt.Errorf("invalid datetime: %w", err)
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

	output := &DateTimeFormatOutput{}

	// Set default format type if not specified
	if input.FormatType == "" {
		if input.CustomFormat != "" {
			input.FormatType = "custom"
		} else if len(input.Formats) > 0 {
			input.FormatType = "multiple"
		} else {
			input.FormatType = "standard"
		}
	}

	switch input.FormatType {
	case "standard":
		formatName := input.StandardFormat
		if formatName == "" {
			formatName = "RFC3339"
		}
		format, ok := standardFormats[formatName]
		if !ok {
			return nil, fmt.Errorf("unknown standard format: %s", formatName)
		}
		output.Formatted = parsedTime.Format(format)

	case "custom":
		if input.CustomFormat == "" {
			return nil, fmt.Errorf("custom_format is required for custom format type")
		}
		output.Formatted = parsedTime.Format(input.CustomFormat)

	case "relative":
		output.RelativeTime = formatRelativeTime(parsedTime, time.Now(), input.IncludeWeekday)
		output.Formatted = output.RelativeTime

	case "multiple":
		if len(input.Formats) == 0 {
			// Default set of formats
			input.Formats = []string{
				"RFC3339",
				"2006-01-02",
				"January 2, 2006",
				"15:04:05",
				"Monday",
			}
		}

		output.MultipleFormats = make(map[string]string)
		for _, formatStr := range input.Formats {
			// Check if it's a standard format name
			if format, ok := standardFormats[formatStr]; ok {
				output.MultipleFormats[formatStr] = parsedTime.Format(format)
			} else {
				// Treat as custom format
				output.MultipleFormats[formatStr] = parsedTime.Format(formatStr)
			}
		}

	default:
		return nil, fmt.Errorf("invalid format_type: %s", input.FormatType)
	}

	// Add localized components if requested
	if input.Locale != "" {
		output.Localized = getLocalizedComponents(parsedTime, input.Locale)
	}

	// Emit result event
	if ctx.Events != nil {
		ctx.Events.Emit(agentDomain.EventToolResult, agentDomain.ToolResultEventData{
			ToolName:  "datetime_format",
			Result:    output,
			RequestID: ctx.RunID,
		})
	}

	return output, nil
}

// DateTimeFormat returns a tool that formats date/time strings in various representations.
// It supports standard formats (RFC3339, RFC1123, etc.), custom Go time layouts, relative time formatting,
// and basic localization for Spanish, French, and German. The tool can output multiple formats
// simultaneously and provides human-readable relative time descriptions like "3 days ago" or "tomorrow".
func DateTimeFormat() agentDomain.Tool {
	// Define parameter schema
	paramSchema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"datetime": {
				Type:        "string",
				Description: "Date/time to format (RFC3339 format preferred)",
			},
			"format_type": {
				Type:        "string",
				Description: "Format type: standard, custom, relative, multiple",
				Enum:        []string{"standard", "custom", "relative", "multiple"},
			},
			"standard_format": {
				Type:        "string",
				Description: "Standard format name",
				Enum:        []string{"RFC3339", "RFC1123", "RFC822", "Kitchen", "Stamp", "ISO8601", "UnixDate"},
			},
			"custom_format": {
				Type:        "string",
				Description: "Custom format layout (Go time format)",
			},
			"formats": {
				Type:        "array",
				Description: "Multiple format strings for multiple output",
				Items: &schemaDomain.Property{
					Type: "string",
				},
			},
			"timezone": {
				Type:        "string",
				Description: "Timezone for formatting",
			},
			"include_weekday": {
				Type:        "boolean",
				Description: "Include weekday name in relative format",
			},
			"locale": {
				Type:        "string",
				Description: "Locale for localized formatting",
			},
		},
		Required: []string{"datetime"},
	}

	// Define output schema
	outputSchema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"formatted": {
				Type:        "string",
				Description: "Formatted date/time string",
			},
			"multiple_formats": {
				Type:        "object",
				Description: "Multiple formatted outputs (if requested)",
			},
			"relative_time": {
				Type:        "string",
				Description: "Relative time description",
			},
			"localized": {
				Type:        "object",
				Description: "Localized components (if locale specified)",
				Properties: map[string]schemaDomain.Property{
					"month_name": {
						Type:        "string",
						Description: "Full month name in the specified locale",
					},
					"month_name_short": {
						Type:        "string",
						Description: "Short month name in the specified locale",
					},
					"weekday_name": {
						Type:        "string",
						Description: "Full weekday name in the specified locale",
					},
					"weekday_short": {
						Type:        "string",
						Description: "Short weekday name in the specified locale",
					},
					"period": {
						Type:        "string",
						Description: "AM/PM period indicator",
					},
				},
			},
		},
	}

	builder := atools.NewToolBuilder("datetime_format", "Format date/time to various string representations").
		WithFunction(dateTimeFormatExecute).
		WithParameterSchema(paramSchema).
		WithOutputSchema(outputSchema).
		WithUsageInstructions(`The datetime_format tool provides flexible date/time formatting capabilities:

Format Types:
1. Standard Formats:
   - RFC3339: "2006-01-02T15:04:05Z07:00" (default)
   - RFC1123: "Mon, 02 Jan 2006 15:04:05 MST"
   - RFC822: "02 Jan 06 15:04 MST"
   - ISO8601: "2006-01-02T15:04:05Z07:00"
   - Kitchen: "3:04PM"
   - Stamp: "Jan _2 15:04:05"
   - UnixDate: "Mon Jan _2 15:04:05 MST 2006"

2. Custom Formats:
   Use Go's time layout syntax with these reference components:
   - Year: "2006" (4 digits), "06" (2 digits)
   - Month: "01" or "1" (number), "Jan" (short name), "January" (full name)
   - Day: "02" or "2" (day of month), "_2" (space-padded)
   - Weekday: "Mon" (short), "Monday" (full)
   - Hour: "15" (24-hour), "03" or "3" (12-hour), "PM" (AM/PM)
   - Minute: "04" or "4"
   - Second: "05" or "5"
   - Millisecond: ".000" (decimal), ".999" (trailing zeros removed)
   - Timezone: "MST" (name), "-0700" (numeric), "Z07:00" (ISO 8601)

3. Relative Time:
   Formats dates as human-readable relative times:
   - "a few seconds ago", "in 3 minutes"
   - "yesterday at 15:04", "tomorrow at 09:30"
   - "3 days ago", "in 2 weeks"
   - "1 month ago", "in 1 year"
   - Optional weekday inclusion: "3 days ago (Monday)"

4. Multiple Formats:
   Output the same date in multiple formats simultaneously.
   Specify format names or custom format strings.

Localization:
Basic support for Spanish (es), French (fr), and German (de):
- Localized month names (full and short)
- Localized weekday names (full and short)  
- AM/PM indicators where applicable

State Integration:
- default_timezone: Default timezone if not specified in input

Examples of Custom Formats:
- "Monday, January 2, 2006": Full date with weekday
- "02/01/2006 15:04:05": European datetime
- "Jan 2 '06 at 3:04pm": Compact friendly format
- "2006-W01-1": ISO week date
- "2006.01.02 AD at 15:04 MST": With era`).
		WithExamples([]agentDomain.ToolExample{
			{
				Name:        "Standard RFC3339 format",
				Description: "Format using the default standard format",
				Scenario:    "When you need a standard, unambiguous date format",
				Input: map[string]interface{}{
					"datetime": "2024-03-15T14:30:45Z",
				},
				Output: map[string]interface{}{
					"formatted": "2024-03-15T14:30:45Z",
				},
				Explanation: "Returns the date in RFC3339 format, which is the default standard format",
			},
			{
				Name:        "Human-readable custom format",
				Description: "Format date in a friendly, readable way",
				Scenario:    "When displaying dates to end users",
				Input: map[string]interface{}{
					"datetime":      "2024-03-15T14:30:45Z",
					"format_type":   "custom",
					"custom_format": "Monday, January 2, 2006 at 3:04 PM",
				},
				Output: map[string]interface{}{
					"formatted": "Friday, March 15, 2024 at 2:30 PM",
				},
				Explanation: "Uses Go's time format syntax to create a human-friendly date string",
			},
			{
				Name:        "Relative time format",
				Description: "Show time relative to now",
				Scenario:    "When showing how long ago something happened",
				Input: map[string]interface{}{
					"datetime":        "2024-03-12T10:00:00Z",
					"format_type":     "relative",
					"include_weekday": true,
				},
				Output: map[string]interface{}{
					"formatted":     "3 days ago (Tuesday)",
					"relative_time": "3 days ago (Tuesday)",
				},
				Explanation: "Shows the date as '3 days ago' with the weekday included",
			},
			{
				Name:        "Multiple format output",
				Description: "Get the same date in multiple formats",
				Scenario:    "When you need different format options for different uses",
				Input: map[string]interface{}{
					"datetime":    "2024-03-15T14:30:45-04:00",
					"format_type": "multiple",
					"formats":     []string{"RFC3339", "2006-01-02", "Kitchen", "Monday"},
				},
				Output: map[string]interface{}{
					"multiple_formats": map[string]string{
						"RFC3339":    "2024-03-15T14:30:45-04:00",
						"2006-01-02": "2024-03-15",
						"Kitchen":    "2:30PM",
						"Monday":     "Friday",
					},
				},
				Explanation: "Returns the date formatted in multiple ways simultaneously",
			},
			{
				Name:        "Localized format with Spanish",
				Description: "Format date with Spanish month and weekday names",
				Scenario:    "When displaying dates for Spanish-speaking users",
				Input: map[string]interface{}{
					"datetime":      "2024-03-15T14:30:45Z",
					"format_type":   "custom",
					"custom_format": "Monday, 2 January 2006",
					"locale":        "es",
				},
				Output: map[string]interface{}{
					"formatted": "Friday, 15 March 2024",
					"localized": map[string]interface{}{
						"month_name":       "marzo",
						"month_name_short": "mar",
						"weekday_name":     "viernes",
						"weekday_short":    "vie",
						"period":           "PM",
					},
				},
				Explanation: "Provides localized components for constructing Spanish date strings",
			},
			{
				Name:        "Format with timezone",
				Description: "Format date in a specific timezone",
				Scenario:    "When displaying times for users in different timezones",
				Input: map[string]interface{}{
					"datetime":      "2024-03-15T14:30:45Z",
					"timezone":      "America/New_York",
					"format_type":   "custom",
					"custom_format": "Jan 2, 2006 3:04 PM MST",
				},
				Output: map[string]interface{}{
					"formatted": "Mar 15, 2024 10:30 AM EDT",
				},
				Explanation: "Converts UTC time to New York timezone before formatting",
			},
			{
				Name:        "Kitchen time format",
				Description: "Simple time format for casual display",
				Scenario:    "When you only need to show the time portion",
				Input: map[string]interface{}{
					"datetime":        "2024-03-15T14:30:45Z",
					"format_type":     "standard",
					"standard_format": "Kitchen",
				},
				Output: map[string]interface{}{
					"formatted": "2:30PM",
				},
				Explanation: "Kitchen format shows time in simple 12-hour format",
			},
			{
				Name:        "Relative time for recent events",
				Description: "Format very recent times",
				Scenario:    "When showing activity that just happened",
				Input: map[string]interface{}{
					"datetime":    "2024-03-15T14:29:30Z",
					"format_type": "relative",
				},
				Output: map[string]interface{}{
					"formatted":     "1 minute ago",
					"relative_time": "1 minute ago",
				},
				Explanation: "Shows very recent times in minutes or seconds",
			},
		}).
		WithConstraints([]string{
			"Date/time must be provided in a parseable format (RFC3339 preferred)",
			"Custom format must use Go time layout syntax",
			"Timezone must be a valid IANA timezone name",
			"Locale support is limited to es, fr, and de",
			"Relative time is calculated from current time",
			"Standard format names are case-sensitive",
			"Multiple formats can mix standard names and custom layouts",
		}).
		WithErrorGuidance(map[string]string{
			"invalid datetime":          "Ensure the datetime is in a valid format. RFC3339 (e.g., '2024-03-15T14:30:45Z') is preferred",
			"invalid timezone":          "Use a valid IANA timezone name like 'America/New_York' or 'UTC'",
			"unknown standard format":   "Use one of: RFC3339, RFC1123, RFC822, Kitchen, Stamp, ISO8601, UnixDate",
			"custom_format is required": "Provide a custom format string when format_type is 'custom'",
			"invalid format_type":       "Use one of: standard, custom, relative, multiple",
			"invalid locale":            "Currently supported locales: es (Spanish), fr (French), de (German)",
		}).
		WithCategory("datetime").
		WithTags([]string{"datetime", "format", "localization", "relative-time", "i18n", "display", "timezone"}).
		WithVersion("2.0.0").
		WithBehavior(false, false, false, "fast") // Non-deterministic due to relative time

	return builder.Build()
}

// formatRelativeTime formats a time relative to now
func formatRelativeTime(t, now time.Time, includeWeekday bool) string {
	duration := now.Sub(t)
	absDuration := duration
	if duration < 0 {
		absDuration = -duration
	}

	// Helper to add weekday if requested
	addWeekday := func(s string) string {
		if includeWeekday {
			return fmt.Sprintf("%s (%s)", s, t.Weekday().String())
		}
		return s
	}

	// Less than a minute
	if absDuration < time.Minute {
		if duration < 0 {
			return "in a few seconds"
		}
		return "a few seconds ago"
	}

	// Less than an hour
	if absDuration < time.Hour {
		minutes := int(absDuration.Minutes())
		if duration < 0 {
			if minutes == 1 {
				return "in 1 minute"
			}
			return fmt.Sprintf("in %d minutes", minutes)
		}
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	}

	// Less than a day
	if absDuration < 24*time.Hour {
		hours := int(absDuration.Hours())
		if duration < 0 {
			if hours == 1 {
				return "in 1 hour"
			}
			return fmt.Sprintf("in %d hours", hours)
		}
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}

	// Check if it's today, yesterday, or tomorrow
	nowDate := now.Truncate(24 * time.Hour)
	tDate := t.Truncate(24 * time.Hour)
	daysDiff := int(nowDate.Sub(tDate).Hours() / 24)

	switch daysDiff {
	case 0:
		return addWeekday(fmt.Sprintf("today at %s", t.Format("15:04")))
	case 1:
		return addWeekday(fmt.Sprintf("yesterday at %s", t.Format("15:04")))
	case -1:
		return addWeekday(fmt.Sprintf("tomorrow at %s", t.Format("15:04")))
	}

	// Less than a week
	if absDuration < 7*24*time.Hour {
		days := int(absDuration.Hours() / 24)
		if duration < 0 {
			if days == 1 {
				return addWeekday("in 1 day")
			}
			return addWeekday(fmt.Sprintf("in %d days", days))
		}
		if days == 1 {
			return addWeekday("1 day ago")
		}
		return addWeekday(fmt.Sprintf("%d days ago", days))
	}

	// Less than a month
	if absDuration < 30*24*time.Hour {
		weeks := int(absDuration.Hours() / (24 * 7))
		if duration < 0 {
			if weeks == 1 {
				return "in 1 week"
			}
			return fmt.Sprintf("in %d weeks", weeks)
		}
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	}

	// Less than a year
	if absDuration < 365*24*time.Hour {
		months := int(absDuration.Hours() / (24 * 30))
		if duration < 0 {
			if months == 1 {
				return "in 1 month"
			}
			return fmt.Sprintf("in %d months", months)
		}
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	}

	// Years
	years := int(absDuration.Hours() / (24 * 365))
	if duration < 0 {
		if years == 1 {
			return "in 1 year"
		}
		return fmt.Sprintf("in %d years", years)
	}
	if years == 1 {
		return "1 year ago"
	}
	return fmt.Sprintf("%d years ago", years)
}

// getLocalizedComponents returns localized date components
func getLocalizedComponents(t time.Time, locale string) *LocalizedComponents {
	// Default to English
	components := &LocalizedComponents{
		MonthName:      t.Month().String(),
		MonthNameShort: t.Month().String()[:3],
		WeekdayName:    t.Weekday().String(),
		WeekdayShort:   t.Weekday().String()[:3],
	}

	// Add period if applicable
	hour := t.Hour()
	if hour < 12 {
		components.Period = "AM"
	} else {
		components.Period = "PM"
	}

	// Apply locale if available
	if data, ok := localeData[strings.ToLower(locale)]; ok {
		month := int(t.Month())
		weekday := int(t.Weekday())

		if month >= 1 && month <= 12 {
			components.MonthName = data.months[month]
			components.MonthNameShort = data.monthsShort[month]
		}

		if weekday >= 0 && weekday < 7 {
			components.WeekdayName = data.weekdays[weekday]
			components.WeekdayShort = data.weekdaysShort[weekday]
		}
	}

	return components
}

func init() {
	tools.MustRegisterTool("datetime_format", DateTimeFormat(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "datetime_format",
			Category:    "datetime",
			Tags:        []string{"datetime", "format", "localization", "relative-time", "i18n", "display"},
			Description: "Format date/time strings with standard formats, custom layouts, relative time, and localization",
			Version:     "2.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Standard format",
					Description: "Format using standard RFC3339",
					Code:        `DateTimeFormat().Execute(ctx, DateTimeFormatInput{DateTime: "2024-03-15T10:30:00Z"})`,
				},
				{
					Name:        "Custom format",
					Description: "Format using custom Go time layout",
					Code:        `DateTimeFormat().Execute(ctx, DateTimeFormatInput{DateTime: "2024-03-15T10:30:00Z", FormatType: "custom", CustomFormat: "Monday, January 2, 2006 at 3:04 PM"})`,
				},
				{
					Name:        "Relative time",
					Description: "Format as relative time (e.g., '3 days ago')",
					Code:        `DateTimeFormat().Execute(ctx, DateTimeFormatInput{DateTime: "2024-03-12T10:30:00Z", FormatType: "relative"})`,
				},
				{
					Name:        "Multiple formats",
					Description: "Output multiple format variations",
					Code:        `DateTimeFormat().Execute(ctx, DateTimeFormatInput{DateTime: "2024-03-15T10:30:00Z", FormatType: "multiple", Formats: ["RFC3339", "2006-01-02", "Kitchen"]})`,
				},
				{
					Name:        "Localized format",
					Description: "Format with Spanish localization",
					Code:        `DateTimeFormat().Execute(ctx, DateTimeFormatInput{DateTime: "2024-03-15T10:30:00Z", Locale: "es"})`,
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
		UsageInstructions: `The datetime_format tool formats date/time values in various ways:
- Standard formats: RFC3339, RFC1123, RFC822, Kitchen, Stamp, ISO8601, UnixDate
- Custom formats: Use Go time layout syntax (e.g., "2006-01-02 15:04:05")
- Relative time: Shows time relative to now (e.g., "3 days ago", "in 2 hours")
- Multiple formats: Output the same date in multiple formats simultaneously
- Localization: Basic support for Spanish (es), French (fr), and German (de)

The tool automatically detects the format type based on provided parameters.`,
		Constraints: []string{
			"Date/time must be provided in a parseable format (RFC3339 preferred)",
			"Custom format must use Go time layout syntax",
			"Timezone must be a valid IANA timezone name",
			"Locale support is limited to es, fr, and de",
			"Relative time is calculated from current time",
		},
		ErrorGuidance: map[string]string{
			"invalid datetime":          "Ensure the datetime is in a valid format. RFC3339 is preferred",
			"invalid timezone":          "Use a valid IANA timezone name like 'America/New_York' or 'UTC'",
			"unknown standard format":   "Use one of: RFC3339, RFC1123, RFC822, Kitchen, Stamp, ISO8601, UnixDate",
			"custom_format is required": "Provide a custom format when format_type is 'custom'",
			"invalid format_type":       "Use one of: standard, custom, relative, multiple",
		},
		IsDeterministic:      false, // Relative time depends on current time
		IsDestructive:        false,
		RequiresConfirmation: false,
		EstimatedLatency:     "instant",
	})
}
