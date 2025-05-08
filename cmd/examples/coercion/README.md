# Advanced Type Coercion Example

This example demonstrates the comprehensive type coercion functionality available in the Go-LLMs library, which allows for flexible handling of different data types in your schema validations.

## Overview

The advanced type coercion system enables automatic conversion between different data types when validating JSON data against schemas. This makes it much easier to work with data from various sources that may not exactly match the expected schema types.

## Key Features

### Basic Type Coercion

- **String ↔ Number**: Convert strings to numbers and vice versa
- **String ↔ Boolean**: Convert "true"/"false" strings to boolean values
- **Number ↔ Boolean**: Convert 0/1 to false/true values
- **String ↔ Integer**: Convert string representations of integers to actual integers

### Advanced Type Coercion

- **Date/Time Coercion**: Convert strings in various date formats to time.Time objects
  - ISO8601, RFC3339, Unix timestamps, natural language dates, etc.
  
- **UUID Coercion**: Parse UUID strings into structured UUID objects

- **Email Validation and Coercion**: 
  - Normalize email addresses (extract from formats like "John Doe <john.doe@example.com>")
  - Validate email format

- **URL Coercion**:
  - Add missing schemes (e.g., "example.com" → "http://example.com")
  - Validate URL format

- **Duration Coercion**:
  - Convert between various duration formats
    - Go duration format ("1h30m")
    - Natural language ("2 days")
    - Clock format ("1:30:00")
    - Seconds, milliseconds, etc.

- **Array Coercion**:
  - Convert comma-separated strings to arrays
  - Parse JSON array strings

- **Object Coercion**:
  - Parse JSON object strings into Go maps
  - Convert to nested structures

## Example Usage

The example demonstrates:

1. Basic validation of well-formed data
2. Validation with automatic type coercion
3. Individual coercion functions for specific data types
4. Parsing data with type mismatches into Go structs

## Running the Example

```bash
# Build the coercion example
make example EXAMPLE=coercion

# Run the example
./bin/coercion
```

## Coercion in Action

```go
// Date coercion from various formats
dateStr := "2023-09-15"
dateVal, ok := validation.CoerceToDate(dateStr)

// Unix timestamp to date
timestamp := 1694883600
dateVal, ok := validation.CoerceToDate(timestamp)

// Email normalization
emailWithName := "John Doe <john.doe@example.com>"
emailVal, ok := validation.CoerceToEmail(emailWithName) // "john.doe@example.com"

// URL normalization
urlStr := "example.com"
urlVal, ok := validation.CoerceToURL(urlStr) // "http://example.com"

// Duration from various formats
durationStr := "1:30:00"
durationVal, ok := validation.CoerceToDuration(durationStr) // 1h30m0s

// Array from comma-separated string
arrayStr := "item1, item2, item3"
arrayVal, ok := validation.CoerceToArray(arrayStr) // []interface{}{"item1", "item2", "item3"}
```

## Schema Validation with Coercion

The schema validator has been enhanced to automatically apply appropriate coercion based on the target type and format:

```go
schema := &domain.Schema{
    Type: "object",
    Properties: map[string]domain.Property{
        "startDate": {
            Type:   "string",
            Format: "date-time",
        },
        "duration": {
            Type:   "string",
            Format: "duration",
        },
        "maxAttendees": {
            Type:    "integer",
            Minimum: float64Ptr(1),
        },
    },
}

// Even with type mismatches, validation succeeds due to coercion
eventWithTypeMismatch := `{
    "startDate": "2023-09-15",         // Not ISO8601 but still coerced
    "duration": "1:30:00",             // Clock format coerced to duration
    "maxAttendees": "500"              // String coerced to integer
}`

result, err := validator.Validate(schema, eventWithTypeMismatch)
// result.Valid will be true due to coercion
```

## Using with LLM Outputs

The type coercion system is particularly useful with LLM outputs, which may not always strictly adhere to the requested format but are semantically correct. With automatic coercion, many common variations in LLM output formats can be handled without error.