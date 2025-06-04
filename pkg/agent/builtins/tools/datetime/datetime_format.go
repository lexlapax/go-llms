// ABOUTME: Provides date/time formatting functionality with multiple formats
// ABOUTME: Supports standard formats, custom layouts, localized formatting, and relative time

package datetime

import (
	"fmt"
	"strings"
	"time"

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

var dateTimeFormatParamSchema = &schemaDomain.Schema{
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

// DateTimeFormat returns a tool that formats date/time strings
func DateTimeFormat() agentDomain.Tool {
	return atools.NewTool(
		"datetime_format",
		"Format date/time to various string representations",
		func(ctx *agentDomain.ToolContext, input DateTimeFormatInput) (*DateTimeFormatOutput, error) {
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
		},
		dateTimeFormatParamSchema,
	)
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

	if daysDiff == 0 {
		return addWeekday(fmt.Sprintf("today at %s", t.Format("15:04")))
	} else if daysDiff == 1 {
		return addWeekday(fmt.Sprintf("yesterday at %s", t.Format("15:04")))
	} else if daysDiff == -1 {
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
	// Register the tool
	if err := registerTool("datetime_format", DateTimeFormat()); err != nil {
		panic(fmt.Sprintf("Failed to register datetime_format tool: %v", err))
	}
}
