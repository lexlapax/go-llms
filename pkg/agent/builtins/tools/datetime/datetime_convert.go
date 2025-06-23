// ABOUTME: Provides date/time conversion functionality for timezones and timestamps
// ABOUTME: Supports timezone conversions, unix timestamps, DST info, and timezone listing

package datetime

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	agentDomain "github.com/lexlapax/go-llms/pkg/agent/domain"
	atools "github.com/lexlapax/go-llms/pkg/agent/tools"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// DateTimeConvertInput defines the input for the datetime_convert tool
type DateTimeConvertInput struct {
	// Operation: "timezone", "to_timestamp", "from_timestamp", "list_timezones"
	Operation string `json:"operation"`
	// DateTime to convert (RFC3339 format preferred)
	DateTime string `json:"datetime,omitempty"`
	// Unix timestamp (for from_timestamp operation)
	Timestamp int64 `json:"timestamp,omitempty"`
	// Timestamp unit: "seconds", "milliseconds", "microseconds", "nanoseconds"
	TimestampUnit string `json:"timestamp_unit,omitempty"`
	// Source timezone (for timezone conversion)
	FromTimezone string `json:"from_timezone,omitempty"`
	// Target timezone (for timezone conversion)
	ToTimezone string `json:"to_timezone,omitempty"`
	// Filter for timezone listing (e.g., "America", "Europe")
	TimezoneFilter string `json:"timezone_filter,omitempty"`
	// Include DST information
	IncludeDST bool `json:"include_dst,omitempty"`
}

// DateTimeConvertOutput defines the output for the datetime_convert tool
type DateTimeConvertOutput struct {
	// Converted date/time in RFC3339 format
	Converted string `json:"converted,omitempty"`
	// Unix timestamp
	Timestamp int64 `json:"timestamp,omitempty"`
	// Timestamp in milliseconds
	TimestampMillis int64 `json:"timestamp_millis,omitempty"`
	// Timestamp in microseconds
	TimestampMicros int64 `json:"timestamp_micros,omitempty"`
	// Timestamp in nanoseconds
	TimestampNanos int64 `json:"timestamp_nanos,omitempty"`
	// List of timezones
	Timezones []string `json:"timezones,omitempty"`
	// DST information
	DSTInfo *DSTInfo `json:"dst_info,omitempty"`
	// Timezone information
	TimezoneInfo *TimezoneInfo `json:"timezone_info,omitempty"`
}

// DSTInfo holds daylight saving time information
type DSTInfo struct {
	IsDST          bool   `json:"is_dst"`
	DSTName        string `json:"dst_name,omitempty"`
	StandardName   string `json:"standard_name,omitempty"`
	CurrentOffset  string `json:"current_offset"`  // e.g., "-05:00"
	StandardOffset string `json:"standard_offset"` // e.g., "-05:00"
	DSTOffset      string `json:"dst_offset"`      // e.g., "-04:00"
}

// TimezoneInfo holds timezone information
type TimezoneInfo struct {
	Name          string `json:"name"`           // e.g., "America/New_York"
	Abbreviation  string `json:"abbreviation"`   // e.g., "EST" or "EDT"
	Offset        string `json:"offset"`         // e.g., "-05:00"
	OffsetSeconds int    `json:"offset_seconds"` // e.g., -18000
}

// Common timezones for listing
var commonTimezones = []string{
	"UTC",
	"America/New_York",
	"America/Chicago",
	"America/Denver",
	"America/Los_Angeles",
	"America/Phoenix",
	"America/Toronto",
	"America/Vancouver",
	"America/Mexico_City",
	"America/Sao_Paulo",
	"America/Buenos_Aires",
	"Europe/London",
	"Europe/Paris",
	"Europe/Berlin",
	"Europe/Madrid",
	"Europe/Rome",
	"Europe/Amsterdam",
	"Europe/Brussels",
	"Europe/Vienna",
	"Europe/Stockholm",
	"Europe/Moscow",
	"Asia/Tokyo",
	"Asia/Shanghai",
	"Asia/Hong_Kong",
	"Asia/Singapore",
	"Asia/Seoul",
	"Asia/Taipei",
	"Asia/Bangkok",
	"Asia/Jakarta",
	"Asia/Manila",
	"Asia/Kolkata",
	"Asia/Dubai",
	"Asia/Tel_Aviv",
	"Australia/Sydney",
	"Australia/Melbourne",
	"Australia/Brisbane",
	"Australia/Perth",
	"Pacific/Auckland",
	"Pacific/Fiji",
	"Pacific/Honolulu",
	"Africa/Cairo",
	"Africa/Johannesburg",
	"Africa/Lagos",
	"Africa/Nairobi",
}

// dateTimeConvertExecute is the execution function for datetime_convert
func dateTimeConvertExecute(ctx *agentDomain.ToolContext, input DateTimeConvertInput) (*DateTimeConvertOutput, error) {
	// Emit start event
	if ctx.Events != nil {
		ctx.Events.Emit(agentDomain.EventToolCall, agentDomain.ToolCallEventData{
			ToolName:   "datetime_convert",
			Parameters: input,
			RequestID:  ctx.RunID,
		})
	}

	output := &DateTimeConvertOutput{}

	switch input.Operation {
	case "timezone":
		if input.DateTime == "" {
			return nil, fmt.Errorf("datetime is required for timezone conversion")
		}
		if input.ToTimezone == "" {
			return nil, fmt.Errorf("to_timezone is required for timezone conversion")
		}

		// Parse the input date/time
		parsedTime, err := parseDate(input.DateTime)
		if err != nil {
			return nil, fmt.Errorf("invalid datetime: %w", err)
		}

		// Apply source timezone if specified
		if input.FromTimezone != "" && input.FromTimezone != "UTC" {
			loc, err := time.LoadLocation(input.FromTimezone)
			if err != nil {
				return nil, fmt.Errorf("invalid from_timezone: %w", err)
			}
			// Adjust the time to be in the source timezone
			parsedTime = time.Date(
				parsedTime.Year(), parsedTime.Month(), parsedTime.Day(),
				parsedTime.Hour(), parsedTime.Minute(), parsedTime.Second(),
				parsedTime.Nanosecond(), loc,
			)
		}

		// Convert to target timezone
		targetLoc, err := time.LoadLocation(input.ToTimezone)
		if err != nil {
			return nil, fmt.Errorf("invalid to_timezone: %w", err)
		}
		convertedTime := parsedTime.In(targetLoc)
		output.Converted = convertedTime.Format(time.RFC3339)

		// Add timezone info
		output.TimezoneInfo = getTimezoneInfo(convertedTime)

		// Add DST info if requested
		if input.IncludeDST {
			output.DSTInfo = getDSTInfo(convertedTime, targetLoc)
		}

	case "to_timestamp":
		if input.DateTime == "" {
			return nil, fmt.Errorf("datetime is required for timestamp conversion")
		}

		parsedTime, err := parseDate(input.DateTime)
		if err != nil {
			return nil, fmt.Errorf("invalid datetime: %w", err)
		}

		// Apply timezone if specified
		if input.FromTimezone != "" {
			loc, err := time.LoadLocation(input.FromTimezone)
			if err != nil {
				return nil, fmt.Errorf("invalid from_timezone: %w", err)
			}
			parsedTime = time.Date(
				parsedTime.Year(), parsedTime.Month(), parsedTime.Day(),
				parsedTime.Hour(), parsedTime.Minute(), parsedTime.Second(),
				parsedTime.Nanosecond(), loc,
			)
		} else if ctx.State != nil {
			// Check state for default timezone
			if defaultTZ, exists := ctx.State.Get("default_timezone"); exists {
				if tzStr, ok := defaultTZ.(string); ok && tzStr != "" {
					loc, err := time.LoadLocation(tzStr)
					if err == nil {
						parsedTime = time.Date(
							parsedTime.Year(), parsedTime.Month(), parsedTime.Day(),
							parsedTime.Hour(), parsedTime.Minute(), parsedTime.Second(),
							parsedTime.Nanosecond(), loc,
						)
					}
				}
			}
		}

		// Output all timestamp formats
		output.Timestamp = parsedTime.Unix()
		output.TimestampMillis = parsedTime.UnixMilli()
		output.TimestampMicros = parsedTime.UnixMicro()
		output.TimestampNanos = parsedTime.UnixNano()

	case "from_timestamp":
		if input.Timestamp == 0 {
			return nil, fmt.Errorf("timestamp is required for timestamp conversion")
		}

		var parsedTime time.Time
		unit := input.TimestampUnit
		if unit == "" {
			unit = "seconds"
		}

		switch unit {
		case "seconds":
			parsedTime = time.Unix(input.Timestamp, 0)
		case "milliseconds":
			parsedTime = time.UnixMilli(input.Timestamp)
		case "microseconds":
			parsedTime = time.UnixMicro(input.Timestamp)
		case "nanoseconds":
			seconds := input.Timestamp / 1e9
			nanos := input.Timestamp % 1e9
			parsedTime = time.Unix(seconds, nanos)
		default:
			return nil, fmt.Errorf("invalid timestamp_unit: %s", unit)
		}

		// Apply timezone if specified
		if input.ToTimezone != "" {
			loc, err := time.LoadLocation(input.ToTimezone)
			if err != nil {
				return nil, fmt.Errorf("invalid to_timezone: %w", err)
			}
			parsedTime = parsedTime.In(loc)
		} else {
			parsedTime = parsedTime.UTC()
		}

		output.Converted = parsedTime.Format(time.RFC3339)
		output.TimezoneInfo = getTimezoneInfo(parsedTime)

	case "list_timezones":
		timezones := getFilteredTimezones(input.TimezoneFilter)
		output.Timezones = timezones

	default:
		return nil, fmt.Errorf("invalid operation: %s", input.Operation)
	}

	// Emit result event
	if ctx.Events != nil {
		ctx.Events.Emit(agentDomain.EventToolResult, agentDomain.ToolResultEventData{
			ToolName:  "datetime_convert",
			Result:    output,
			RequestID: ctx.RunID,
		})
	}

	return output, nil
}

// DateTimeConvert returns a tool that converts date/time between timezones and formats.
// It provides comprehensive timezone conversion, unix timestamp conversion, and timezone listing capabilities.
// The tool supports IANA timezone identifiers, handles DST transitions automatically, and can convert
// between various timestamp units (seconds, milliseconds, microseconds, nanoseconds).
func DateTimeConvert() agentDomain.Tool {
	// Define parameter schema
	paramSchema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"operation": {
				Type:        "string",
				Description: "Conversion operation",
				Enum:        []string{"timezone", "to_timestamp", "from_timestamp", "list_timezones"},
			},
			"datetime": {
				Type:        "string",
				Description: "Date/time to convert (RFC3339 format preferred)",
			},
			"timestamp": {
				Type:        "integer",
				Description: "Unix timestamp for conversion",
			},
			"timestamp_unit": {
				Type:        "string",
				Description: "Unit of timestamp",
				Enum:        []string{"seconds", "milliseconds", "microseconds", "nanoseconds"},
			},
			"from_timezone": {
				Type:        "string",
				Description: "Source timezone",
			},
			"to_timezone": {
				Type:        "string",
				Description: "Target timezone",
			},
			"timezone_filter": {
				Type:        "string",
				Description: "Filter for timezone listing",
			},
			"include_dst": {
				Type:        "boolean",
				Description: "Include DST information",
			},
		},
		Required: []string{"operation"},
	}

	// Define output schema
	outputSchema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"converted": {
				Type:        "string",
				Description: "Converted date/time in RFC3339 format",
			},
			"timestamp": {
				Type:        "integer",
				Description: "Unix timestamp in seconds",
			},
			"timestamp_millis": {
				Type:        "integer",
				Description: "Unix timestamp in milliseconds",
			},
			"timestamp_micros": {
				Type:        "integer",
				Description: "Unix timestamp in microseconds",
			},
			"timestamp_nanos": {
				Type:        "integer",
				Description: "Unix timestamp in nanoseconds",
			},
			"timezones": {
				Type:        "array",
				Description: "List of matching timezones",
				Items: &schemaDomain.Property{
					Type: "string",
				},
			},
			"dst_info": {
				Type:        "object",
				Description: "Daylight saving time information",
				Properties: map[string]schemaDomain.Property{
					"is_dst": {
						Type:        "boolean",
						Description: "Whether the time is in DST",
					},
					"dst_name": {
						Type:        "string",
						Description: "DST timezone name (e.g., 'EDT')",
					},
					"standard_name": {
						Type:        "string",
						Description: "Standard timezone name (e.g., 'EST')",
					},
					"current_offset": {
						Type:        "string",
						Description: "Current UTC offset (e.g., '-05:00')",
					},
					"standard_offset": {
						Type:        "string",
						Description: "Standard UTC offset",
					},
					"dst_offset": {
						Type:        "string",
						Description: "DST UTC offset",
					},
				},
			},
			"timezone_info": {
				Type:        "object",
				Description: "Timezone information",
				Properties: map[string]schemaDomain.Property{
					"name": {
						Type:        "string",
						Description: "Full timezone name (e.g., 'America/New_York')",
					},
					"abbreviation": {
						Type:        "string",
						Description: "Timezone abbreviation (e.g., 'EST')",
					},
					"offset": {
						Type:        "string",
						Description: "UTC offset (e.g., '-05:00')",
					},
					"offset_seconds": {
						Type:        "integer",
						Description: "UTC offset in seconds",
					},
				},
			},
		},
	}

	builder := atools.NewToolBuilder("datetime_convert", "Convert date/time between timezones and unix timestamps").
		WithFunction(dateTimeConvertExecute).
		WithParameterSchema(paramSchema).
		WithOutputSchema(outputSchema).
		WithUsageInstructions(`The datetime_convert tool provides comprehensive timezone and timestamp conversion capabilities:

Operations:
1. Timezone Conversion:
   - Convert dates between any IANA timezones
   - Handles DST transitions automatically
   - Preserves the exact moment in time
   - Supports source timezone specification

2. To Timestamp:
   - Convert datetime to unix timestamps
   - Outputs in seconds, milliseconds, microseconds, and nanoseconds
   - Respects timezone information in the input
   - Uses state timezone or UTC as default

3. From Timestamp:
   - Convert unix timestamps to datetime
   - Supports various units: seconds, milliseconds, microseconds, nanoseconds
   - Can output in any specified timezone
   - Defaults to UTC if no timezone specified

4. List Timezones:
   - Get filtered list of IANA timezone identifiers
   - Supports substring filtering
   - Returns common timezones by default
   - Useful for timezone discovery

Timezone Information:
- Full IANA timezone names (e.g., "America/New_York")
- Timezone abbreviations (e.g., "EST", "EDT")
- UTC offsets in "+/-HH:MM" format
- Offset in seconds for calculations

DST (Daylight Saving Time) Information:
- Current DST status
- Standard and DST timezone names
- Standard and DST offsets
- Automatic detection of DST periods

State Integration:
- default_timezone: Used when from_timezone not specified for to_timestamp

Common Timezone Examples:
- Americas: America/New_York, America/Chicago, America/Los_Angeles, America/Toronto
- Europe: Europe/London, Europe/Paris, Europe/Berlin, Europe/Moscow
- Asia: Asia/Tokyo, Asia/Shanghai, Asia/Kolkata, Asia/Dubai
- Pacific: Australia/Sydney, Pacific/Auckland, Pacific/Honolulu`).
		WithExamples([]agentDomain.ToolExample{
			{
				Name:        "Convert between timezones",
				Description: "Convert a New York time to Tokyo time",
				Scenario:    "When coordinating meetings across timezones",
				Input: map[string]interface{}{
					"operation":     "timezone",
					"datetime":      "2024-03-15T10:30:00-04:00",
					"from_timezone": "America/New_York",
					"to_timezone":   "Asia/Tokyo",
				},
				Output: map[string]interface{}{
					"converted": "2024-03-15T23:30:00+09:00",
					"timezone_info": map[string]interface{}{
						"name":           "Asia/Tokyo",
						"abbreviation":   "JST",
						"offset":         "+09:00",
						"offset_seconds": 32400,
					},
				},
				Explanation: "10:30 AM EDT in New York is 11:30 PM JST the same day in Tokyo",
			},
			{
				Name:        "Convert with DST information",
				Description: "Convert time and get DST details",
				Scenario:    "When you need to know if a time falls in daylight saving period",
				Input: map[string]interface{}{
					"operation":   "timezone",
					"datetime":    "2024-07-15T10:00:00Z",
					"to_timezone": "America/New_York",
					"include_dst": true,
				},
				Output: map[string]interface{}{
					"converted": "2024-07-15T06:00:00-04:00",
					"dst_info": map[string]interface{}{
						"is_dst":          true,
						"dst_name":        "EDT",
						"standard_name":   "EST",
						"current_offset":  "-04:00",
						"standard_offset": "-05:00",
						"dst_offset":      "-04:00",
					},
				},
				Explanation: "July 15 is during DST period in New York, so EDT (UTC-4) applies",
			},
			{
				Name:        "Convert datetime to timestamps",
				Description: "Get unix timestamps in various units",
				Scenario:    "When you need timestamps for APIs or databases",
				Input: map[string]interface{}{
					"operation": "to_timestamp",
					"datetime":  "2024-03-15T14:30:45.123Z",
				},
				Output: map[string]interface{}{
					"timestamp":        int64(1710512445),
					"timestamp_millis": int64(1710512445123),
					"timestamp_micros": int64(1710512445123000),
					"timestamp_nanos":  int64(1710512445123000000),
				},
				Explanation: "Converts to unix timestamps in seconds, milliseconds, microseconds, and nanoseconds",
			},
			{
				Name:        "Convert timestamp to datetime",
				Description: "Convert unix timestamp to readable date",
				Scenario:    "When interpreting timestamps from logs or APIs",
				Input: map[string]interface{}{
					"operation":      "from_timestamp",
					"timestamp":      int64(1710512445),
					"timestamp_unit": "seconds",
					"to_timezone":    "Europe/London",
				},
				Output: map[string]interface{}{
					"converted": "2024-03-15T14:30:45Z",
					"timezone_info": map[string]interface{}{
						"name":           "Europe/London",
						"abbreviation":   "GMT",
						"offset":         "+00:00",
						"offset_seconds": 0,
					},
				},
				Explanation: "Converts unix timestamp to human-readable format in London timezone",
			},
			{
				Name:        "Convert millisecond timestamp",
				Description: "Convert JavaScript-style millisecond timestamp",
				Scenario:    "When working with JavaScript Date.now() values",
				Input: map[string]interface{}{
					"operation":      "from_timestamp",
					"timestamp":      int64(1710512445123),
					"timestamp_unit": "milliseconds",
					"to_timezone":    "America/Los_Angeles",
				},
				Output: map[string]interface{}{
					"converted": "2024-03-15T07:30:45.123-07:00",
					"timezone_info": map[string]interface{}{
						"name":           "America/Los_Angeles",
						"abbreviation":   "PDT",
						"offset":         "-07:00",
						"offset_seconds": -25200,
					},
				},
				Explanation: "Millisecond precision timestamp converted to Pacific time",
			},
			{
				Name:        "List European timezones",
				Description: "Find all European timezone identifiers",
				Scenario:    "When you need to show timezone options for Europe",
				Input: map[string]interface{}{
					"operation":       "list_timezones",
					"timezone_filter": "Europe",
				},
				Output: map[string]interface{}{
					"timezones": []string{
						"Europe/Amsterdam",
						"Europe/Athens",
						"Europe/Berlin",
						"Europe/Brussels",
						"Europe/Dublin",
						"Europe/Helsinki",
						"Europe/Lisbon",
						"Europe/London",
						"Europe/Madrid",
						"Europe/Moscow",
						"Europe/Paris",
						"Europe/Rome",
						"Europe/Stockholm",
						"Europe/Vienna",
						"Europe/Warsaw",
					},
				},
				Explanation: "Returns all timezone identifiers containing 'Europe'",
			},
			{
				Name:        "Convert without source timezone",
				Description: "Convert assuming UTC source",
				Scenario:    "When converting UTC times to local timezones",
				Input: map[string]interface{}{
					"operation":   "timezone",
					"datetime":    "2024-03-15T14:30:00Z",
					"to_timezone": "Australia/Sydney",
				},
				Output: map[string]interface{}{
					"converted": "2024-03-16T01:30:00+11:00",
					"timezone_info": map[string]interface{}{
						"name":           "Australia/Sydney",
						"abbreviation":   "AEDT",
						"offset":         "+11:00",
						"offset_seconds": 39600,
					},
				},
				Explanation: "UTC time converted to Sydney time (next day due to timezone difference)",
			},
			{
				Name:        "Timestamp with timezone context",
				Description: "Convert local time to timestamp",
				Scenario:    "When converting a local time to unix timestamp",
				Input: map[string]interface{}{
					"operation":     "to_timestamp",
					"datetime":      "2024-03-15 10:30:00",
					"from_timezone": "America/Chicago",
				},
				Output: map[string]interface{}{
					"timestamp":        int64(1710515400),
					"timestamp_millis": int64(1710515400000),
					"timestamp_micros": int64(1710515400000000),
					"timestamp_nanos":  int64(1710515400000000000),
				},
				Explanation: "Local Chicago time converted to UTC timestamp",
			},
		}).
		WithConstraints([]string{
			"Date/time must be in a parseable format (RFC3339 preferred)",
			"Timezones must be valid IANA timezone identifiers",
			"Timestamps are interpreted as UTC unless timezone specified",
			"Timestamp units default to seconds if not specified",
			"DST information is calculated based on the timezone's rules",
			"Timezone abbreviations may vary by date due to DST",
			"Maximum timestamp precision is nanoseconds",
			"Timezone list filtering is case-insensitive substring matching",
		}).
		WithErrorGuidance(map[string]string{
			"invalid datetime":        "Ensure the datetime is in a valid format. RFC3339 (e.g., '2024-03-15T14:30:00Z') is preferred",
			"invalid from_timezone":   "Use a valid IANA timezone name like 'America/New_York' or 'Europe/London'",
			"invalid to_timezone":     "Use a valid IANA timezone name. Run list_timezones to see available options",
			"datetime is required":    "Provide a datetime string for timezone or to_timestamp operations",
			"to_timezone is required": "Specify the target timezone for timezone conversion",
			"timestamp is required":   "Provide a numeric timestamp for from_timestamp operation",
			"invalid timestamp_unit":  "Use one of: seconds, milliseconds, microseconds, nanoseconds",
			"invalid operation":       "Use one of: timezone, to_timestamp, from_timestamp, list_timezones",
		}).
		WithCategory("datetime").
		WithTags([]string{"datetime", "timezone", "conversion", "timestamp", "unix", "dst", "iana", "utc"}).
		WithVersion("2.0.0").
		WithBehavior(true, false, false, "fast")

	return builder.Build()
}

// getTimezoneInfo returns timezone information for a time
func getTimezoneInfo(t time.Time) *TimezoneInfo {
	name, offsetSeconds := t.Zone()

	location := t.Location().String()
	offsetStr := t.Format("-07:00")

	return &TimezoneInfo{
		Name:          location,
		Abbreviation:  name,
		Offset:        offsetStr,
		OffsetSeconds: offsetSeconds,
	}
}

// getDSTInfo returns DST information for a time and location
func getDSTInfo(t time.Time, loc *time.Location) *DSTInfo {
	// Check if current time is in DST
	_, currentOffset := t.Zone()
	isDST := false

	// Get standard offset by checking January 1st
	jan1 := time.Date(t.Year(), 1, 1, 0, 0, 0, 0, loc)
	_, janOffset := jan1.Zone()

	// Get DST offset by checking July 1st
	jul1 := time.Date(t.Year(), 7, 1, 0, 0, 0, 0, loc)
	_, julOffset := jul1.Zone()

	standardOffset := janOffset
	dstOffset := julOffset

	// In Southern Hemisphere, it might be reversed
	if janOffset > julOffset {
		standardOffset = julOffset
		dstOffset = janOffset
	}

	isDST = currentOffset == dstOffset && dstOffset != standardOffset

	// Format offsets
	currentOffsetStr := formatOffset(currentOffset)
	standardOffsetStr := formatOffset(standardOffset)
	dstOffsetStr := formatOffset(dstOffset)

	info := &DSTInfo{
		IsDST:          isDST,
		CurrentOffset:  currentOffsetStr,
		StandardOffset: standardOffsetStr,
		DSTOffset:      dstOffsetStr,
	}

	// Get zone names
	if isDST {
		info.DSTName = t.Format("MST")
		info.StandardName = jan1.Format("MST")
	} else {
		info.StandardName = t.Format("MST")
		info.DSTName = jul1.Format("MST")
	}

	return info
}

// formatOffset formats an offset in seconds to "+/-HH:MM" format
func formatOffset(offsetSeconds int) string {
	sign := "+"
	if offsetSeconds < 0 {
		sign = "-"
		offsetSeconds = -offsetSeconds
	}
	hours := offsetSeconds / 3600
	minutes := (offsetSeconds % 3600) / 60
	return fmt.Sprintf("%s%02d:%02d", sign, hours, minutes)
}

// getFilteredTimezones returns a filtered list of timezones
func getFilteredTimezones(filter string) []string {
	if filter == "" {
		return commonTimezones
	}

	filter = strings.ToLower(filter)
	var filtered []string

	// First check common timezones
	for _, tz := range commonTimezones {
		if strings.Contains(strings.ToLower(tz), filter) {
			filtered = append(filtered, tz)
		}
	}

	// If we need more, check all available timezones
	if len(filtered) < 10 {
		// Get zones from the time package (this is a simplified approach)
		additionalZones := []string{
			"America/Anchorage", "America/Halifax", "America/Regina",
			"America/Bogota", "America/Lima", "America/Caracas",
			"Europe/Dublin", "Europe/Lisbon", "Europe/Warsaw",
			"Europe/Budapest", "Europe/Athens", "Europe/Helsinki",
			"Asia/Mumbai", "Asia/Karachi", "Asia/Dhaka",
			"Asia/Bangkok", "Asia/Ho_Chi_Minh", "Asia/Jakarta",
			"Australia/Adelaide", "Australia/Darwin", "Australia/Hobart",
			"Pacific/Guam", "Pacific/Port_Moresby", "Pacific/Noumea",
		}

		for _, tz := range additionalZones {
			if strings.Contains(strings.ToLower(tz), filter) {
				// Avoid duplicates
				found := false
				for _, existing := range filtered {
					if existing == tz {
						found = true
						break
					}
				}
				if !found {
					filtered = append(filtered, tz)
				}
			}
		}
	}

	sort.Strings(filtered)
	return filtered
}

func init() {
	tools.MustRegisterTool("datetime_convert", DateTimeConvert(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "datetime_convert",
			Category:    "datetime",
			Tags:        []string{"datetime", "timezone", "conversion", "timestamp", "unix", "dst", "iana"},
			Description: "Convert date/time between timezones, unix timestamps, and provide timezone information",
			Version:     "2.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Timezone conversion",
					Description: "Convert time from New York to Tokyo",
					Code:        `DateTimeConvert().Execute(ctx, DateTimeConvertInput{Operation: "timezone", DateTime: "2024-03-15T10:30:00-05:00", FromTimezone: "America/New_York", ToTimezone: "Asia/Tokyo"})`,
				},
				{
					Name:        "To timestamp",
					Description: "Convert datetime to unix timestamp",
					Code:        `DateTimeConvert().Execute(ctx, DateTimeConvertInput{Operation: "to_timestamp", DateTime: "2024-03-15T10:30:00Z"})`,
				},
				{
					Name:        "From timestamp",
					Description: "Convert unix timestamp to datetime",
					Code:        `DateTimeConvert().Execute(ctx, DateTimeConvertInput{Operation: "from_timestamp", Timestamp: 1710502200, ToTimezone: "Europe/London"})`,
				},
				{
					Name:        "List timezones",
					Description: "List all timezones containing 'america'",
					Code:        `DateTimeConvert().Execute(ctx, DateTimeConvertInput{Operation: "list_timezones", TimezoneFilter: "america"})`,
				},
				{
					Name:        "DST information",
					Description: "Get timezone conversion with DST info",
					Code:        `DateTimeConvert().Execute(ctx, DateTimeConvertInput{Operation: "timezone", DateTime: "2024-07-15T10:30:00Z", ToTimezone: "America/New_York", IncludeDST: true})`,
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
		UsageInstructions: `The datetime_convert tool provides timezone and timestamp conversion capabilities:
- Timezone conversion: Convert dates between any IANA timezones
- To timestamp: Convert datetime to unix timestamps (seconds, milliseconds, microseconds, nanoseconds)
- From timestamp: Convert unix timestamps back to datetime in any timezone
- List timezones: Get a filtered list of available timezones
- DST information: Get detailed daylight saving time information

The tool handles timezone abbreviations, offsets, and DST transitions automatically.`,
		Constraints: []string{
			"Date/time must be in a parseable format (RFC3339 preferred)",
			"Timezones must be valid IANA timezone names",
			"Timestamps are interpreted as UTC unless timezone specified",
			"Timestamp units default to seconds if not specified",
			"DST information is calculated based on the timezone rules",
		},
		ErrorGuidance: map[string]string{
			"invalid datetime":        "Ensure the datetime is in a valid format. RFC3339 is preferred",
			"invalid from_timezone":   "Use a valid IANA timezone name for the source timezone",
			"invalid to_timezone":     "Use a valid IANA timezone name for the target timezone",
			"datetime is required":    "Provide a datetime for timezone or timestamp conversion",
			"to_timezone is required": "Specify the target timezone for conversion",
			"timestamp is required":   "Provide a timestamp for from_timestamp conversion",
			"invalid timestamp_unit":  "Use one of: seconds, milliseconds, microseconds, nanoseconds",
			"invalid operation":       "Use one of: timezone, to_timestamp, from_timestamp, list_timezones",
		},
		IsDeterministic:      true,
		IsDestructive:        false,
		RequiresConfirmation: false,
		EstimatedLatency:     "instant",
	})
}
