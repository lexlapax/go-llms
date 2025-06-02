# Built-in DateTime Tools Example

This example demonstrates the comprehensive date/time tools available in the go-llms library.

## Overview

The built-in datetime tools provide functionality for:
- Getting current date/time in various formats and timezones
- Parsing dates from various formats including relative dates
- Performing date arithmetic and business day calculations
- Formatting dates with localization support
- Converting between timezones with DST handling
- Getting detailed date information
- Comparing and sorting dates

## Running the Example

```bash
go run main.go
```

## Available DateTime Tools

1. **datetime_now** - Get current date/time with components, timestamps, and custom formats
2. **datetime_parse** - Parse dates from strings with auto-detection and relative date support
3. **datetime_calculate** - Perform date arithmetic including business days and age calculations
4. **datetime_format** - Format dates to various strings with localization
5. **datetime_convert** - Convert between timezones and handle Unix timestamps
6. **datetime_info** - Get comprehensive date information including leap year detection
7. **datetime_compare** - Compare dates and find differences

## Example Usage

### Get Current Time with Details
```go
nowTool := tools.MustGetTool("datetime_now")
result, _ := nowTool.Execute(ctx, map[string]interface{}{
    "timezone": "America/New_York",
    "include_components": true,
    "include_timestamps": true,
})
```

### Parse Relative Dates
```go
parseTool := tools.MustGetTool("datetime_parse")
result, _ := parseTool.Execute(ctx, map[string]interface{}{
    "date_string": "next Monday at 3pm",
    "timezone": "America/Los_Angeles",
})
```

### Business Day Calculations
```go
calcTool := tools.MustGetTool("datetime_calculate")
result, _ := calcTool.Execute(ctx, map[string]interface{}{
    "operation": "add_business_days",
    "date_time": "2024-01-15T10:00:00Z",
    "business_days": 5,
})
```

### Format with Localization
```go
formatTool := tools.MustGetTool("datetime_format")
result, _ := formatTool.Execute(ctx, map[string]interface{}{
    "datetime": "2024-12-25T10:30:00Z",
    "format_type": "multiple",
    "formats": []string{"RFC3339", "Kitchen"},
    "locale": "es", // Spanish
})
```

### Timezone Conversion with DST
```go
convertTool := tools.MustGetTool("datetime_convert")
result, _ := convertTool.Execute(ctx, map[string]interface{}{
    "operation": "timezone",
    "datetime": "2024-07-15T15:00:00Z",
    "to_timezone": "Asia/Tokyo",
    "include_dst": true,
})
```

## Key Features

- **Comprehensive Parsing**: Auto-detects common formats and handles relative dates
- **Business Logic**: Includes business day calculations excluding weekends
- **Localization**: Supports multiple languages for month and day names
- **Timezone Aware**: All operations can be timezone-specific
- **Rich Metadata**: Tools provide detailed information about dates
- **Flexible Formatting**: Multiple output formats in a single call
- **DST Handling**: Proper daylight saving time support

## Integration with Agents

These tools can be easily integrated with agents for time-based workflows:

```go
agent := workflow.NewAgent(
    "scheduler-agent",
    provider,
    workflow.WithTools(
        tools.MustGetTool("datetime_now"),
        tools.MustGetTool("datetime_calculate"),
        tools.MustGetTool("datetime_format"),
    ),
)
```