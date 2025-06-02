# Built-in Data Tools Example

This example demonstrates the data processing tools available in the go-llms library.

## Overview

The built-in data tools provide functionality for:
- Processing JSON data with JSONPath queries
- Handling CSV data with filtering and statistics
- Parsing XML and converting to JSON
- Common data transformations (filter, map, reduce, sort, group)

## Running the Example

```bash
go run main.go
```

## Available Data Tools

1. **json_process** - Process JSON data
   - Parse and validate JSON
   - Query with JSONPath expressions
   - Transform operations: extract_keys, extract_values, flatten, prettify, minify

2. **csv_process** - Handle CSV data
   - Parse with configurable delimiters
   - Filter with multiple operators
   - Transform: select columns, sort, statistics
   - Convert to JSON

3. **xml_process** - Process XML data
   - Parse with attribute support
   - Query with simplified XPath
   - Convert to JSON with configurable options

4. **data_transform** - Common transformations
   - Filter with complex conditions
   - Map operations (extract field, case conversion, type conversion)
   - Reduce operations (sum, count, min, max, average, concat)
   - Additional: sort, group_by, unique, reverse

## Example Usage

### JSON Processing
```go
jsonTool := tools.MustGetTool("json_process")

// Query with JSONPath
result, _ := jsonTool.Execute(ctx, map[string]interface{}{
    "operation": "query",
    "data":      jsonString,
    "jsonpath":  "$.users[?(@.age > 25)].name",
})

// Transform JSON
result, _ := jsonTool.Execute(ctx, map[string]interface{}{
    "operation": "transform",
    "data":      jsonString,
    "transform": "flatten",
})
```

### CSV Processing
```go
csvTool := tools.MustGetTool("csv_process")

// Filter CSV data
result, _ := csvTool.Execute(ctx, map[string]interface{}{
    "operation": "filter",
    "data":      csvString,
    "filters": []map[string]interface{}{
        {
            "column":   "age",
            "operator": "gt",
            "value":    25,
        },
    },
    "has_header": true,
})

// Get statistics
result, _ := csvTool.Execute(ctx, map[string]interface{}{
    "operation": "transform",
    "data":      csvString,
    "transform": "statistics",
    "columns":   []string{"salary"},
})
```

### XML Processing
```go
xmlTool := tools.MustGetTool("xml_process")

// Query with XPath
result, _ := xmlTool.Execute(ctx, map[string]interface{}{
    "operation": "query",
    "data":      xmlString,
    "xpath":     "//book[@id='1']/title",
})

// Convert to JSON
result, _ := xmlTool.Execute(ctx, map[string]interface{}{
    "operation": "to_json",
    "data":      xmlString,
    "include_attributes": true,
})
```

### Data Transformations
```go
transformTool := tools.MustGetTool("data_transform")

// Filter data
result, _ := transformTool.Execute(ctx, map[string]interface{}{
    "operation": "filter",
    "data":      dataArray,
    "condition": map[string]interface{}{
        "field":    "score",
        "operator": "gte",
        "value":    80,
    },
})

// Map to extract field
result, _ := transformTool.Execute(ctx, map[string]interface{}{
    "operation": "map",
    "data":      dataArray,
    "transform": "extract_field",
    "field":     "name",
})

// Reduce to sum
result, _ := transformTool.Execute(ctx, map[string]interface{}{
    "operation": "reduce",
    "data":      dataArray,
    "reducer":   "sum",
    "field":     "amount",
})
```

## Key Features

### JSONPath Support
- Object navigation: `$.store.book`
- Array indexing: `$.store.book[0]`
- Wildcards: `$.store.book[*].author`
- Filters: `$.store.book[?(@.price < 10)]`
- Recursive descent: `$..author`

### CSV Operations
- **Filtering Operators**: eq, ne, contains, starts_with, ends_with, gt, lt, gte, lte
- **Statistics**: count, sum, average, min, max, standard deviation
- **Transformations**: column selection, sorting, JSON conversion

### XML Features
- Element selection by tag name
- Attribute queries with @ prefix
- Nested element navigation
- Preserve or ignore attributes in JSON conversion

### Transform Operations
- **Filter**: Complex conditions with AND/OR logic
- **Map**: Field extraction, case conversion, type conversion
- **Reduce**: Aggregation operations
- **Group**: Group by field values
- **Sort**: Ascending/descending by field
- **Unique**: Remove duplicates

## Integration with Agents

Data tools can be used with agents for data processing workflows:

```go
agent := workflow.NewAgent(
    "data-processor",
    provider,
    workflow.WithTools(
        tools.MustGetTool("json_process"),
        tools.MustGetTool("csv_process"),
        tools.MustGetTool("data_transform"),
    ),
)

// Agent can now process and analyze data
response, _ := agent.Run(ctx, workflow.UserMessage(
    "Load the CSV file, filter for high-value customers, and calculate statistics",
))
```

## Performance Considerations

- These tools process data in-memory
- For large datasets, consider streaming or chunking
- JSONPath queries are optimized for common patterns
- CSV parsing handles quoted fields and escape characters

## Common Use Cases

1. **Data Analysis**: Filter, aggregate, and analyze structured data
2. **ETL Operations**: Extract, transform, and load data between formats
3. **API Response Processing**: Parse and extract data from JSON APIs
4. **Report Generation**: Transform raw data into summary statistics
5. **Configuration Management**: Process XML/JSON config files

## Error Handling

All tools provide detailed error messages for:
- Invalid data format
- Malformed queries (JSONPath, XPath)
- Type mismatches in operations
- Missing required fields
- Invalid operator usage