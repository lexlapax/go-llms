// ABOUTME: Registry functions for datetime tools
// ABOUTME: Handles registration of all datetime-related tools

package datetime

import (
	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// registerTool registers a datetime tool with the tools registry
func registerTool(name string, tool domain.Tool) error {
	metadata := builtins.Metadata{
		Name:        name,
		Category:    "datetime",
		Description: tool.Description(),
		Version:     "1.0.0",
		Tags:        []string{"datetime", "time", "date", "timezone"},
		Examples:    getExamplesForTool(name),
	}

	return tools.Tools.Register(name, tool, metadata)
}

// getExamplesForTool returns examples for each datetime tool
func getExamplesForTool(name string) []builtins.Example {
	switch name {
	case "datetime_now":
		return []builtins.Example{
			{
				Name:        "Basic Usage",
				Description: "Get current time in UTC and local timezone",
				Code:        `result := datetime_now({})`,
			},
			{
				Name:        "With Timezone",
				Description: "Get current time in specific timezone with components",
				Code:        `result := datetime_now({"timezone": "America/New_York", "include_components": true, "include_week_info": true})`,
			},
		}
	case "datetime_info":
		return []builtins.Example{
			{
				Name:        "Date Information",
				Description: "Get information about a specific date",
				Code:        `result := datetime_info({"date": "2024-01-15"})`,
			},
		}
	case "datetime_calculate":
		return []builtins.Example{
			{
				Name:        "Add Days",
				Description: "Add 5 days to a date",
				Code:        `result := datetime_calculate({"date": "2024-01-15", "operation": "add", "unit": "days", "value": 5})`,
			},
		}
	case "datetime_parse":
		return []builtins.Example{
			{
				Name:        "Parse Date",
				Description: "Parse a date string",
				Code:        `result := datetime_parse({"date_string": "January 15, 2024"})`,
			},
		}
	case "datetime_format":
		return []builtins.Example{
			{
				Name:        "Custom Format",
				Description: "Format a date to custom format",
				Code:        `result := datetime_format({"date": "2024-01-15T10:30:00Z", "format": "Monday, January 2, 2006"})`,
			},
		}
	case "datetime_convert":
		return []builtins.Example{
			{
				Name:        "Timezone Conversion",
				Description: "Convert time between timezones",
				Code:        `result := datetime_convert({"time": "2024-01-15T10:30:00Z", "from_timezone": "UTC", "to_timezone": "America/New_York"})`,
			},
		}
	case "datetime_compare":
		return []builtins.Example{
			{
				Name:        "Compare Dates",
				Description: "Compare two dates",
				Code:        `result := datetime_compare({"date1": "2024-01-15", "date2": "2024-01-20"})`,
			},
		}
	default:
		return []builtins.Example{}
	}
}

// Package initialization is handled by individual tool files
