# Built-in Data Tools Example

This comprehensive example demonstrates all 4 data processing tools available in the go-llms library, showcasing their full capabilities with practical examples.

## Overview

The built-in data tools provide powerful, LLM-free data processing:
- **JSON Processing**: Parse, query with JSONPath, and transform JSON data
- **CSV Processing**: Parse, filter, calculate statistics, and convert formats
- **XML Processing**: Parse, query with XPath, and convert to JSON
- **Data Transformations**: Filter, map, reduce, sort, group, and more

## Running the Example

```bash
go run main.go
```

## Available Data Tools

### 1. json_process - JSON Data Processing
- Parse and validate JSON
- Query with JSONPath expressions (simple subset)
- Transform operations: flatten, extract, prettify, minify

### 2. csv_process - CSV Data Handling
- Parse CSV with headers and custom delimiters
- Filter rows with conditions
- Calculate statistics on numeric columns
- Convert to JSON format

### 3. xml_process - XML Data Processing
- Parse and validate XML
- Query with simplified XPath
- Convert to JSON with attribute preservation

### 4. data_transform - General Data Transformations
- Filter with conditions
- Map operations (field extraction, case conversion)
- Reduce operations (sum, average, min, max, count)
- Additional: sort, group_by, unique, reverse

## Important Parameter Names

Each tool uses specific parameter names that must be matched exactly:

### JSON Process Tool
```go
// Query operation
"operation": "query"
"data": jsonString
"jsonpath": "$.users[0].name"

// Transform operation
"operation": "transform"
"data": jsonString
"transform": "flatten"  // or "extract", "prettify", "minify"
```

### CSV Process Tool
```go
// Filter operation
"operation": "filter"
"data": csvString
"filter_condition": "column:operator:value"  // NOT "condition"
"has_headers": true  // MUST be plural

// Convert to JSON operation (direct)
"operation": "to_json"  // NOT "transform" with "to_json"
"data": csvString
"has_headers": true

// Statistics operation
"operation": "transform"
"data": csvString
"transform": "statistics"
"has_headers": true
"params": map[string]interface{}{
    "columns": []string{"col1", "col2"},  // wrap in params
}
```

### XML Process Tool
```go
// Query operation
"operation": "query"
"data": xmlString
"xpath": "metadata/version"  // Use XPath syntax with /, not dot notation

// Convert operation
"operation": "to_json"
"data": xmlString
"include_attributes": true
```

### Data Transform Tool
```go
// All operations expect data as JSON string
"data": string(jsonBytes)  // NOT raw object/array

// Filter operation
"operation": "filter"
"field": "score"
"condition": "gt:85"  // format: operator:value

// Map operation
"operation": "map"
"map_type": "extract_field"  // NOT "transform"
"field": "name"

// Reduce operation
"operation": "reduce"
"reduce_type": "average"  // NOT "reducer"
"field": "score"

// Sort operation
"operation": "sort"
"field": "score"
"sort_order": "desc"  // NOT "order"
```

## Limitations and Workarounds

### JSONPath Limitations
The JSONPath implementation supports only basic queries:
- ✅ Simple paths: `$.users[0].name`
- ✅ Array indexing: `$.items[2]`
- ❌ Wildcards: `$.users[*].name`
- ❌ Filters: `$.users[?(@.age > 25)]`
- ❌ Recursive descent: `$..name`

**Workaround**: Use data_transform tool for complex filtering after extracting arrays.

### XPath Limitations
The XPath implementation is simplified:
- ✅ Direct paths: `metadata/version`, `book/title`
- ❌ Double slash: `//book/title`
- ❌ Attributes: `//book[@id='1']`
- ❌ Array indexing: `book[0]/title`
- ❌ Complex predicates

**Workaround**: Convert to JSON first, then use json_process for complex queries.

### CSV Filter Limitations
CSV filtering uses single conditions only:
- ✅ Single condition: `department:eq:Engineering`
- ❌ Multiple conditions with AND/OR logic

**Workaround**: Apply filters sequentially or convert to JSON for complex filtering.

## Complete Working Examples

### JSON Processing Example
```go
jsonData := `{"users": [{"name": "Alice", "age": 30}, {"name": "Bob", "age": 25}]}`

// Parse JSON
result, _ := jsonTool.Execute(ctx, map[string]interface{}{
    "operation": "parse",
    "data":      jsonData,
})

// Query specific user
result, _ := jsonTool.Execute(ctx, map[string]interface{}{
    "operation": "query",
    "data":      jsonData,
    "jsonpath":  "$.users[0]",
})

// Flatten structure
result, _ := jsonTool.Execute(ctx, map[string]interface{}{
    "operation": "transform",
    "data":      jsonData,
    "transform": "flatten",
})
```

### CSV Processing Example
```go
csvData := `name,age,department,salary
Alice,30,Engineering,75000
Bob,25,Marketing,55000`

// Filter by department
result, _ := csvTool.Execute(ctx, map[string]interface{}{
    "operation":        "filter",
    "data":             csvData,
    "filter_condition": "department:eq:Engineering",
    "has_headers":      true,  // Fixed: plural
})

// Convert to JSON
result, _ := csvTool.Execute(ctx, map[string]interface{}{
    "operation":   "to_json",  // Fixed: direct operation
    "data":        csvData,
    "has_headers": true,
})

// Get statistics
result, _ := csvTool.Execute(ctx, map[string]interface{}{
    "operation":   "transform",
    "data":        csvData,
    "transform":   "statistics",
    "has_headers": true,
    "params": map[string]interface{}{  // Fixed: wrap in params
        "columns": []string{"salary"},
    },
})
```

### Data Transformation Example
```go
// Convert data to JSON string first
data := []map[string]interface{}{
    {"name": "Alice", "score": 85},
    {"name": "Bob", "score": 92},
}
jsonData, _ := json.Marshal(data)

// Filter high scores
result, _ := transformTool.Execute(ctx, map[string]interface{}{
    "operation": "filter",
    "data":      string(jsonData),
    "field":     "score",
    "condition": "gt:80",
})

// Extract names
result, _ := transformTool.Execute(ctx, map[string]interface{}{
    "operation": "map",
    "data":      string(jsonData),
    "map_type":  "extract_field",
    "field":     "name",
})

// Calculate average
result, _ := transformTool.Execute(ctx, map[string]interface{}{
    "operation":   "reduce",
    "data":        string(jsonData),
    "reduce_type": "average",
    "field":       "score",
})
```

## Type Assertions

When handling tool outputs, use the correct struct types:
```go
// JSON Process
if output, ok := result.(*data.JSONProcessOutput); ok {
    fmt.Printf("Result: %v\n", output.Result)
}

// CSV Process
if output, ok := result.(*data.CSVProcessOutput); ok {
    fmt.Printf("Rows: %d, Columns: %v\n", output.RowCount, output.Columns)
}

// XML Process
if output, ok := result.(*data.XMLProcessOutput); ok {
    fmt.Printf("Root: %s, Result: %v\n", output.RootElement, output.Result)
}

// Data Transform
if output, ok := result.(*data.DataTransformOutput); ok {
    fmt.Printf("Items: %d, Result: %v\n", output.ItemCount, output.Result)
}
```

## Enhanced Features

### CSV Statistics Output
The statistics operation now provides comprehensive numeric analysis:
```go
{
  "row_count": 7,
  "column_count": 6,
  "salary": {
    "count": 7,
    "sum": 472000,
    "min": 55000,
    "max": 85000,
    "avg": 67428.57,
    "variance": 88244897.96,
    "std_dev": 88244897.96
  },
  "performance_rating": {
    "count": 7,
    "sum": 29.4,
    "min": 3.8,
    "max": 4.7,
    "avg": 4.2,
    "variance": 0.091,
    "std_dev": 0.091
  }
}
```

## Real-World Use Cases

1. **API Response Processing**: Extract and transform data from JSON APIs
2. **CSV Report Analysis**: Filter and calculate statistics on CSV exports  
3. **Configuration Processing**: Parse and query XML/JSON config files
4. **Data Pipeline**: Chain tools for ETL operations
5. **Log Analysis**: Process structured log data

## Performance Considerations

- All tools process data in-memory (no streaming)
- Large datasets may require chunking
- Type conversions are handled automatically where possible
- Error handling adds minimal overhead

## Integration with Agents

```go
agent := workflow.NewAgent(provider).
    SetSystemPrompt("You are a data analyst assistant.").
    AddTool(tools.MustGetTool("json_process")).
    AddTool(tools.MustGetTool("csv_process")).
    AddTool(tools.MustGetTool("data_transform"))

result, _ := agent.Run(ctx, "Analyze the sales CSV and find top performers")
```

## Next Steps

- Explore the [agent example](../agent/) to see data tools in workflows
- Check the [built-in components guide](../../../docs/user-guide/built-in-components.md) for all tools
- Review tool source code for advanced usage patterns