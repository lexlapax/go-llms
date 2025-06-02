# Built-in DateTime Tools Example

This comprehensive example demonstrates all 7 datetime tools available in the go-llms library, showcasing their full capabilities with practical examples.

## Overview

The built-in datetime tools provide a complete suite of date/time operations:
- **Current Time**: Get current date/time in multiple timezones with rich metadata
- **Parsing**: Parse dates from various formats including natural language
- **Calculations**: Perform date arithmetic, business day calculations, and age computation
- **Formatting**: Format dates with localization support in multiple languages
- **Conversions**: Convert between timezones and Unix timestamps with DST awareness
- **Information**: Get comprehensive date details including leap year and week numbers
- **Comparisons**: Compare, sort, and find min/max dates with range checking

## Running the Example

```bash
go run main.go
```

## Available DateTime Tools

### 1. datetime_now - Current Date/Time
- Get current time in any timezone
- Include date components (year, month, day, hour, minute, second)
- Week information (week number, ISO week, day of week)
- Multiple timestamps (Unix, Unix milliseconds)
- Custom formatting with Go time layout strings

### 2. datetime_parse - Parse Date Strings
- Auto-detect common date formats (ISO 8601, RFC3339, US/EU formats)
- Natural language parsing ("next Monday", "tomorrow at noon", "in 3 days")
- Relative date parsing ("last week", "2 hours ago")
- Timezone-aware parsing
- Validates parsed dates

### 3. datetime_calculate - Date Arithmetic
- Add/subtract duration (years, months, days, hours, minutes, seconds)
- Business day calculations (skip weekends and holidays)
- Age calculations with years, months, and days breakdown
- Date differences in various units
- ISO 8601 duration support (P1Y2M3DT4H5M6S)

### 4. datetime_format - Format Dates
- Standard formats (RFC3339, Kitchen, Unix date)
- Custom formats using Go time layout
- Relative time formatting ("2 hours ago", "in 3 days")
- Multiple formats in one call
- Localization support (month/day names in different languages)

### 5. datetime_convert - Conversions
- Timezone conversions with DST handling
- Convert to/from Unix timestamps (operations: "to_timestamp", "from_timestamp")
- List available timezones with filtering
- Preserves precision during conversions
- Shows DST status and offset information

### 6. datetime_info - Date Information
- Comprehensive date metadata
- Day of week, day of year, week number
- ISO week information
- Quarter identification
- Leap year detection
- Days in month
- Timezone offset and name

### 7. datetime_compare - Compare and Sort
- Compare two dates (before, after, equal)
- Calculate differences between dates
- Sort multiple dates (ascending/descending)
- Find min/max from date list
- Check if date is within range
- Bulk comparisons

## Example Usage Highlights

### 1. Current Time in Multiple Timezones
```go
nowTool := tools.MustGetTool("datetime_now")
// Get current time in different timezones simultaneously
timezones := []string{"UTC", "Europe/London", "Asia/Tokyo", "Australia/Sydney"}
for _, tz := range timezones {
    result, _ := nowTool.Execute(ctx, map[string]interface{}{
        "timezone": tz,
        "format":   "15:04 MST on Monday, Jan 2",
    })
}
```

### 2. Parse Natural Language Dates
```go
parseTool := tools.MustGetTool("datetime_parse")
// Parse various date formats including natural language
dateStrings := []string{
    "tomorrow",
    "next Monday",
    "in 3 days",
    "2024-12-25",
    "December 25, 2024",
}
for _, dateStr := range dateStrings {
    result, _ := parseTool.Execute(ctx, map[string]interface{}{
        "date_string": dateStr,
        "timezone":    "America/Los_Angeles",
    })
}
```

### 3. Complex Date Calculations
```go
calcTool := tools.MustGetTool("datetime_calculate")
// Add business days to a date
result, _ := calcTool.Execute(ctx, map[string]interface{}{
    "operation":  "add_business_days",
    "start_date": "2024-01-15T10:00:00Z",
    "value":      5,
})
// Also supports: age calculation, duration between dates, add/subtract operations
```

### 4. Localized Date Formatting
```go
formatTool := tools.MustGetTool("datetime_format")
// Format dates in multiple languages
locales := []string{"en", "es", "fr", "de", "ja"}
for _, locale := range locales {
    result, _ := formatTool.Execute(ctx, map[string]interface{}{
        "datetime":    "2024-07-14T14:00:00Z",
        "format_type": "standard",
        "locale":      locale,
    })
}
```

### 5. Timezone Conversion Matrix
```go
convertTool := tools.MustGetTool("datetime_convert")
// Convert meeting time to multiple timezones
zones := []string{"America/New_York", "Europe/London", "Australia/Sydney"}
for _, zone := range zones {
    result, _ := convertTool.Execute(ctx, map[string]interface{}{
        "operation":     "timezone",
        "datetime":      meetingTime,
        "from_timezone": "UTC",
        "to_timezone":   zone,
    })
}
```

### 6. Comprehensive Date Information
```go
infoTool := tools.MustGetTool("datetime_info")
// Get detailed information about a leap year date
result, _ := infoTool.Execute(ctx, map[string]interface{}{
    "date": "2024-02-29T00:00:00Z", // Leap year date
})
// Returns: day_of_week, day_of_year, week_number, is_leap_year, quarter, etc.
```

### 7. Advanced Date Comparisons
```go
compareTool := tools.MustGetTool("datetime_compare")
// Sort dates and find min/max
result, _ := compareTool.Execute(ctx, map[string]interface{}{
    "operation": "sort",
    "dates":     dates,
    "sort_order": "desc",
})
// Also supports: min, max, is_between, compare operations
```

## Key Features Demonstrated

- **Multi-Timezone Support**: Work with dates across different timezones seamlessly
- **Natural Language Processing**: Parse human-friendly date descriptions
- **Business Logic**: Handle business days, weekends, and holidays
- **Localization**: Support for multiple languages (English, Spanish, French, German, Japanese)
- **Comprehensive Metadata**: Get detailed information about any date
- **Flexible Operations**: Combine multiple tools for complex workflows
- **Error Handling**: Graceful handling of invalid dates and edge cases

## Real-World Use Cases

1. **Meeting Scheduler**: Convert meeting times across participant timezones
2. **Deadline Calculator**: Add business days for project timelines
3. **Age Verification**: Calculate exact age for compliance
4. **Report Generator**: Format dates according to locale preferences
5. **Event Planning**: Check if dates fall on weekends or holidays
6. **Data Analysis**: Sort and compare historical dates
7. **Reminder System**: Parse natural language for scheduling

## Integration with Agents

These tools are designed to work seamlessly with LLM agents:

```go
// Create an agent with datetime capabilities
agent := workflow.NewAgent(provider).
    SetSystemPrompt("You are a scheduling assistant with datetime tools.").
    AddTool(tools.MustGetTool("datetime_now")).
    AddTool(tools.MustGetTool("datetime_parse")).
    AddTool(tools.MustGetTool("datetime_calculate")).
    AddTool(tools.MustGetTool("datetime_format")).
    AddTool(tools.MustGetTool("datetime_convert"))

// Use the agent
result, _ := agent.Run(ctx, "Schedule a meeting next Tuesday at 2pm EST and show me what time that is in Tokyo")
```

## Performance Considerations

- All tools are optimized for performance with minimal allocations
- Timezone data is cached for repeated operations
- Date parsing uses efficient algorithms for format detection
- Bulk operations (like sorting) are optimized for large datasets

## Important Notes

### Parameter Names
Each tool uses specific parameter names that must be matched exactly:
- **datetime_info**: Uses `"date"` (not `"date_time"`)
- **datetime_convert**: Uses `"to_timestamp"` and `"from_timestamp"` (not `"to_unix"` and `"from_unix"`)
- **datetime_convert**: Uses `"timestamp"` for from_timestamp operation (not `"unix_timestamp"`)
- **datetime_calculate**: Uses `"start_date"` and `"end_date"` for most operations

### Type Assertions
When handling tool outputs, use the correct struct types:
```go
// Correct type assertion for datetime_now
if output, ok := result.(*datetime.DateTimeNowOutput); ok {
    fmt.Printf("UTC: %s\n", output.UTC)
}

// NOT: result.(map[string]interface{})
```

## Next Steps

- Explore the [agent example](../agent/) to see datetime tools in workflows
- Check the [built-in components guide](../../../docs/user-guide/built-in-components.md) for more tools
- Review the [API documentation](../../../docs/api/agent.md) for advanced usage