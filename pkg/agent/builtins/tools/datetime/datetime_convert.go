// ABOUTME: Provides date/time conversion functionality for timezones and timestamps
// ABOUTME: Supports timezone conversions, unix timestamps, DST info, and timezone listing

package datetime

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

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

var dateTimeConvertParamSchema = &schemaDomain.Schema{
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

// DateTimeConvert returns a tool that converts date/time between timezones and formats
func DateTimeConvert() agentDomain.Tool {
	return atools.NewTool(
		"datetime_convert",
		"Convert date/time between timezones and unix timestamps",
		func(ctx context.Context, input DateTimeConvertInput) (*DateTimeConvertOutput, error) {
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

			return output, nil
		},
		dateTimeConvertParamSchema,
	)
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
	// Register the tool
	if err := registerTool("datetime_convert", DateTimeConvert()); err != nil {
		panic(fmt.Sprintf("Failed to register datetime_convert tool: %v", err))
	}
}
