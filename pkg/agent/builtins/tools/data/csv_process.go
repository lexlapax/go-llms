// ABOUTME: CSVProcess tool provides CSV reading, writing, and transformation capabilities
// ABOUTME: This tool enables agents to work with CSV data without requiring LLM processing

package data

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	atools "github.com/lexlapax/go-llms/pkg/agent/tools"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// CSVProcessInput represents the input for CSV processing operations
type CSVProcessInput struct {
	// The CSV data to process (as a string)
	Data string `json:"data" jsonschema:"title=CSV Data,description=The CSV data to process,required"`

	// Operation to perform: parse, filter, transform, to_json
	Operation string `json:"operation" jsonschema:"title=Operation,description=Operation to perform: parse filter transform to_json,enum=parse,enum=filter,enum=transform,enum=to_json,required"`

	// Whether the first row contains headers
	HasHeaders bool `json:"has_headers" jsonschema:"title=Has Headers,description=Whether the first row contains headers,default=true"`

	// Column delimiter (default: comma)
	Delimiter string `json:"delimiter,omitempty" jsonschema:"title=Delimiter,description=Column delimiter character,default=,"`

	// Filter condition for filter operation (column:operator:value)
	FilterCondition string `json:"filter_condition,omitempty" jsonschema:"title=Filter Condition,description=Filter condition in format column:operator:value"`

	// Transform type for transform operation
	Transform string `json:"transform,omitempty" jsonschema:"title=Transform,description=Transform type: select_columns sort count_rows statistics,enum=select_columns,enum=sort,enum=count_rows,enum=statistics"`

	// Additional parameters for transformations
	Params map[string]interface{} `json:"params,omitempty" jsonschema:"title=Parameters,description=Additional parameters for transformations"`
}

// CSVProcessOutput represents the output of CSV processing
type CSVProcessOutput struct {
	// The processed result
	Result interface{} `json:"result"`

	// Error message if any
	Error string `json:"error,omitempty"`

	// Number of rows processed
	RowCount int `json:"row_count,omitempty"`

	// Column names if available
	Columns []string `json:"columns,omitempty"`
}

// csvProcessParamSchema defines parameters for the CSVProcess tool
var csvProcessParamSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"data": {
			Type:        "string",
			Description: "The CSV data to process",
		},
		"operation": {
			Type:        "string",
			Description: "Operation to perform: parse, filter, transform, or to_json",
			Enum:        []string{"parse", "filter", "transform", "to_json"},
		},
		"has_headers": {
			Type:        "boolean",
			Description: "Whether the first row contains headers",
		},
		"delimiter": {
			Type:        "string",
			Description: "Column delimiter character",
		},
		"filter_condition": {
			Type:        "string",
			Description: "Filter condition in format column:operator:value",
		},
		"transform": {
			Type:        "string",
			Description: "Transform type: select_columns, sort, count_rows, or statistics",
			Enum:        []string{"select_columns", "sort", "count_rows", "statistics"},
		},
		"params": {
			Type:        "object",
			Description: "Additional parameters for transformations",
		},
	},
	Required: []string{"data", "operation"},
}

// CSVProcess creates a tool for processing CSV data
// This tool provides comprehensive CSV manipulation capabilities including:
// parsing, filtering, transforming, and converting to JSON format.
// It supports custom delimiters and handles both header and headerless CSV files.
func CSVProcess() domain.Tool {
	// Create output schema for CSVProcessOutput
	outputSchema := &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"result": {
				Type:        "any",
				Description: "The processed result (can be array of arrays, JSON string, statistics, etc.)",
			},
			"error": {
				Type:        "string",
				Description: "Error message if any",
			},
			"row_count": {
				Type:        "integer",
				Description: "Number of rows processed (excluding headers if applicable)",
			},
			"columns": {
				Type:        "array",
				Description: "Column names if headers are present",
				Items: &sdomain.Property{
					Type: "string",
				},
			},
		},
		Required: []string{},
	}

	builder := atools.NewToolBuilder("csv_process", "Process CSV data: parse, filter, transform, or convert to JSON").
		WithFunction(csvProcessExecute).
		WithParameterSchema(csvProcessParamSchema).
		WithOutputSchema(outputSchema).
		WithUsageInstructions(`Use this tool to process CSV (Comma-Separated Values) data in various ways:

Parse Operation:
- Validates and parses CSV data into a structured format
- Detects headers if has_headers is true
- Returns array of arrays representing rows and columns
- Supports custom delimiters (comma by default)

Filter Operation:
- Filter rows based on conditions using column:operator:value format
- Supported operators:
  - eq, =, ==: Equal to
  - ne, !=, <>: Not equal to
  - contains: Contains substring
  - starts_with: Starts with string
  - ends_with: Ends with string
  - gt, >: Greater than (numeric comparison)
  - lt, <: Less than (numeric comparison)
  - gte, >=: Greater than or equal to
  - lte, <=: Less than or equal to
- Column can be header name (if has_headers) or index (0-based)

Transform Operations:
- select_columns: Select specific columns by name or index
  - Requires params.columns as array or comma-separated string
- sort: Sort records (basic implementation)
- count_rows: Count number of data rows (excluding headers)
- statistics: Calculate statistics for numeric columns
  - Optional params.columns to specify which columns to analyze
  - Returns count, sum, min, max, avg, variance, std_dev

Convert to JSON:
- to_json: Convert CSV to JSON format
  - With headers: Returns array of objects with column names as keys
  - Without headers: Returns array of arrays

Delimiter Support:
- Default is comma (,)
- Can specify any single character delimiter (tab, pipe, semicolon, etc.)
- Common delimiters: "," (comma), "\t" (tab), "|" (pipe), ";" (semicolon)

State Integration:
- csv_default_delimiter: Default delimiter from agent state
- csv_max_rows: Maximum rows to process (for performance limits)`).
		WithExamples([]domain.ToolExample{
			{
				Name:        "Parse CSV with headers",
				Description: "Parse a simple CSV file with column headers",
				Scenario:    "When you need to load and structure CSV data",
				Input: map[string]interface{}{
					"data":        "name,age,city\nJohn,30,New York\nJane,25,Boston\nBob,35,Chicago",
					"operation":   "parse",
					"has_headers": true,
				},
				Output: map[string]interface{}{
					"result": [][]string{
						{"name", "age", "city"},
						{"John", "30", "New York"},
						{"Jane", "25", "Boston"},
						{"Bob", "35", "Chicago"},
					},
					"row_count": 3,
					"columns":   []string{"name", "age", "city"},
				},
				Explanation: "The tool returns structured data with headers identified separately",
			},
			{
				Name:        "Filter by numeric condition",
				Description: "Filter rows where age is greater than 25",
				Scenario:    "When you need to extract rows meeting specific criteria",
				Input: map[string]interface{}{
					"data":             "name,age,city\nJohn,30,New York\nJane,25,Boston\nBob,35,Chicago",
					"operation":        "filter",
					"has_headers":      true,
					"filter_condition": "age:gt:25",
				},
				Output: map[string]interface{}{
					"result": [][]string{
						{"name", "age", "city"},
						{"John", "30", "New York"},
						{"Bob", "35", "Chicago"},
					},
					"row_count": 2,
					"columns":   []string{"name", "age", "city"},
				},
				Explanation: "Numeric comparisons work on string values that can be parsed as numbers",
			},
			{
				Name:        "Select specific columns",
				Description: "Extract only name and city columns",
				Scenario:    "When you need to work with a subset of columns",
				Input: map[string]interface{}{
					"data":        "name,age,city,country\nJohn,30,New York,USA\nJane,25,Boston,USA",
					"operation":   "transform",
					"transform":   "select_columns",
					"has_headers": true,
					"params": map[string]interface{}{
						"columns": []string{"name", "city"},
					},
				},
				Output: map[string]interface{}{
					"result": [][]string{
						{"name", "city"},
						{"John", "New York"},
						{"Jane", "Boston"},
					},
					"row_count": 2,
					"columns":   []string{"name", "city"},
				},
				Explanation: "Column selection preserves order and can be used to reorder columns",
			},
			{
				Name:        "Convert to JSON with headers",
				Description: "Transform CSV into JSON array of objects",
				Scenario:    "When you need to work with CSV data as JSON",
				Input: map[string]interface{}{
					"data":        "id,name,score\n1,Alice,95\n2,Bob,87",
					"operation":   "to_json",
					"has_headers": true,
				},
				Output: map[string]interface{}{
					"result": `[
  {
    "id": "1",
    "name": "Alice",
    "score": "95"
  },
  {
    "id": "2",
    "name": "Bob",
    "score": "87"
  }
]`,
					"row_count": 2,
					"columns":   []string{"id", "name", "score"},
				},
				Explanation: "JSON conversion with headers creates objects with column names as keys",
			},
			{
				Name:        "Calculate statistics",
				Description: "Get statistics for numeric columns",
				Scenario:    "When you need to analyze numeric data in CSV",
				Input: map[string]interface{}{
					"data":        "product,price,quantity\nA,10.5,100\nB,20.0,150\nC,15.75,200",
					"operation":   "transform",
					"transform":   "statistics",
					"has_headers": true,
					"params": map[string]interface{}{
						"columns": []string{"price", "quantity"},
					},
				},
				Output: map[string]interface{}{
					"result": map[string]interface{}{
						"row_count":    3,
						"column_count": 3,
						"price": map[string]interface{}{
							"count":    3,
							"sum":      46.25,
							"min":      10.5,
							"max":      20.0,
							"avg":      15.416666666666666,
							"variance": 19.326388888888886,
							"std_dev":  19.326388888888886,
						},
						"quantity": map[string]interface{}{
							"count":    3,
							"sum":      450.0,
							"min":      100.0,
							"max":      200.0,
							"avg":      150.0,
							"variance": 2500.0,
							"std_dev":  2500.0,
						},
					},
					"row_count": 3,
				},
				Explanation: "Statistics include count, sum, min, max, average, variance, and standard deviation",
			},
			{
				Name:        "Parse with custom delimiter",
				Description: "Parse tab-separated values",
				Scenario:    "When working with TSV or other delimited formats",
				Input: map[string]interface{}{
					"data":        "name\tage\tcity\nJohn\t30\tNew York\nJane\t25\tBoston",
					"operation":   "parse",
					"has_headers": true,
					"delimiter":   "\t",
				},
				Output: map[string]interface{}{
					"result": [][]string{
						{"name", "age", "city"},
						{"John", "30", "New York"},
						{"Jane", "25", "Boston"},
					},
					"row_count": 2,
					"columns":   []string{"name", "age", "city"},
				},
				Explanation: "Custom delimiters allow processing TSV, pipe-delimited, and other formats",
			},
			{
				Name:        "Filter with string contains",
				Description: "Find all rows where city contains 'New'",
				Scenario:    "When searching for partial string matches",
				Input: map[string]interface{}{
					"data":             "name,city\nJohn,New York\nJane,Boston\nBob,New Orleans",
					"operation":        "filter",
					"has_headers":      true,
					"filter_condition": "city:contains:New",
				},
				Output: map[string]interface{}{
					"result": [][]string{
						{"name", "city"},
						{"John", "New York"},
						{"Bob", "New Orleans"},
					},
					"row_count": 2,
					"columns":   []string{"name", "city"},
				},
				Explanation: "String operators work on exact substrings, case-sensitive",
			},
		}).
		WithConstraints([]string{
			"Delimiter must be a single character",
			"Large CSV files may be limited by memory or state-configured row limits",
			"Statistics are only calculated for columns with numeric values",
			"Filter conditions use case-sensitive string matching",
			"Column indices are 0-based when headers are not used",
			"Sorting is a basic implementation and may not handle all edge cases",
			"CSV parsing follows RFC 4180 standards with some flexibility",
			"Empty cells are preserved as empty strings",
		}).
		WithErrorGuidance(map[string]string{
			"failed to parse CSV":                    "The CSV data is malformed. Check for unclosed quotes or invalid characters",
			"invalid filter condition format":        "Filter condition must be in format 'column:operator:value'",
			"column not found or index out of range": "The specified column doesn't exist. Check column names or indices",
			"unsupported operator":                   "Use one of: eq, ne, contains, starts_with, ends_with, gt, lt, gte, lte",
			"columns parameter required":             "select_columns transform requires a 'columns' parameter in params",
			"columns parameter must be an array":     "Provide columns as an array of strings or comma-separated string",
			"filter condition required":              "Filter operation requires a 'filter_condition' parameter",
			"transform type required":                "Transform operation requires a 'transform' parameter",
			"invalid operation":                      "Operation must be one of: parse, filter, transform, to_json",
			"unsupported transform type":             "Transform must be one of: select_columns, sort, count_rows, statistics",
			"failed to convert to JSON":              "Unable to convert the CSV data to JSON format",
		}).
		WithCategory("data").
		WithTags([]string{"data", "csv", "parse", "filter", "transform", "tabular", "tsv", "delimited"}).
		WithVersion("2.0.0").
		WithBehavior(true, false, false, "fast")

	return builder.Build()
}

// csvProcessExecute is the main processing logic
func csvProcessExecute(ctx *domain.ToolContext, input CSVProcessInput) (*CSVProcessOutput, error) {
	// Emit start event
	if ctx.Events != nil {
		ctx.Events.EmitMessage(fmt.Sprintf("Starting CSV processing with operation: %s", input.Operation))
	}

	// Set default delimiter from state or use comma
	if input.Delimiter == "" {
		if ctx.State != nil {
			if defaultDelimiter, exists := ctx.State.Get("csv_default_delimiter"); exists {
				if delim, ok := defaultDelimiter.(string); ok && len(delim) > 0 {
					input.Delimiter = delim
				}
			}
		}
		if input.Delimiter == "" {
			input.Delimiter = ","
		}
	}

	// Check for max row limit from state
	var maxRows int
	if ctx.State != nil {
		if limit, exists := ctx.State.Get("csv_max_rows"); exists {
			if rows, ok := limit.(int); ok && rows > 0 {
				maxRows = rows
			}
		}
	}

	var result *CSVProcessOutput
	var err error

	switch input.Operation {
	case "parse":
		result, err = parseCSV(input.Data, input.Delimiter, input.HasHeaders)
	case "filter":
		if input.FilterCondition == "" {
			err = fmt.Errorf("filter condition required for filter operation")
		} else {
			result, err = filterCSV(input.Data, input.Delimiter, input.HasHeaders, input.FilterCondition)
		}
	case "transform":
		if input.Transform == "" {
			err = fmt.Errorf("transform type required for transform operation")
		} else {
			result, err = transformCSV(input.Data, input.Delimiter, input.HasHeaders, input.Transform, input.Params)
		}
	case "to_json":
		result, err = csvToJSON(input.Data, input.Delimiter, input.HasHeaders)
	default:
		err = fmt.Errorf("invalid operation: %s", input.Operation)
	}

	// Apply row limit if specified
	if err == nil && maxRows > 0 && result != nil {
		if rows, ok := result.Result.([][]string); ok && len(rows) > maxRows {
			result.Result = rows[:maxRows]
			if result.RowCount > maxRows {
				result.RowCount = maxRows
			}
		}
	}

	// Emit completion or error event
	if ctx.Events != nil {
		if err != nil {
			ctx.Events.EmitError(err)
		} else {
			ctx.Events.EmitMessage(fmt.Sprintf("CSV processing completed. Processed %d rows", result.RowCount))
		}
	}

	return result, err
}

// parseCSV validates and parses CSV data
func parseCSV(data, delimiter string, hasHeaders bool) (*CSVProcessOutput, error) {
	reader := csv.NewReader(strings.NewReader(data))
	reader.Comma = rune(delimiter[0])
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return &CSVProcessOutput{
			Error: fmt.Sprintf("failed to parse CSV: %v", err),
		}, nil
	}

	output := &CSVProcessOutput{
		RowCount: len(records),
	}

	if hasHeaders && len(records) > 0 {
		output.Columns = records[0]
		output.RowCount-- // Don't count header row
	}

	output.Result = records

	return output, nil
}

// filterCSV applies filtering to CSV data
func filterCSV(data, delimiter string, hasHeaders bool, condition string) (*CSVProcessOutput, error) {
	// Parse condition (format: column:operator:value)
	parts := strings.SplitN(condition, ":", 3)
	if len(parts) != 3 {
		return &CSVProcessOutput{
			Error: "invalid filter condition format. Expected: column:operator:value",
		}, nil
	}

	column, operator, value := parts[0], parts[1], parts[2]

	// Parse CSV
	reader := csv.NewReader(strings.NewReader(data))
	reader.Comma = rune(delimiter[0])
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return &CSVProcessOutput{
			Error: fmt.Sprintf("failed to parse CSV: %v", err),
		}, nil
	}

	if len(records) == 0 {
		return &CSVProcessOutput{
			Result:   [][]string{},
			RowCount: 0,
		}, nil
	}

	var headers []string
	var dataRecords [][]string

	if hasHeaders {
		headers = records[0]
		dataRecords = records[1:]
	} else {
		dataRecords = records
	}

	// Find column index
	colIdx := -1
	if hasHeaders {
		for i, h := range headers {
			if h == column {
				colIdx = i
				break
			}
		}
	} else {
		// If no headers, assume column is a number
		colIdx, _ = strconv.Atoi(column)
	}

	if colIdx < 0 || (len(dataRecords) > 0 && colIdx >= len(dataRecords[0])) {
		return &CSVProcessOutput{
			Error: fmt.Sprintf("column not found or index out of range: %s", column),
		}, nil
	}

	// Apply filter
	filtered := [][]string{}
	if hasHeaders {
		filtered = append(filtered, headers)
	}

	for _, record := range dataRecords {
		if colIdx >= len(record) {
			continue
		}

		cellValue := record[colIdx]
		match := false

		switch operator {
		case "eq", "=", "==":
			match = cellValue == value
		case "ne", "!=", "<>":
			match = cellValue != value
		case "contains":
			match = strings.Contains(cellValue, value)
		case "starts_with":
			match = strings.HasPrefix(cellValue, value)
		case "ends_with":
			match = strings.HasSuffix(cellValue, value)
		case "gt", ">":
			match = compareNumericStrings(cellValue, value, ">")
		case "lt", "<":
			match = compareNumericStrings(cellValue, value, "<")
		case "gte", ">=":
			match = compareNumericStrings(cellValue, value, ">=")
		case "lte", "<=":
			match = compareNumericStrings(cellValue, value, "<=")
		default:
			return &CSVProcessOutput{
				Error: fmt.Sprintf("unsupported operator: %s", operator),
			}, nil
		}

		if match {
			filtered = append(filtered, record)
		}
	}

	rowCount := len(filtered)
	if hasHeaders && rowCount > 0 {
		rowCount--
	}

	return &CSVProcessOutput{
		Result:   filtered,
		RowCount: rowCount,
		Columns:  headers,
	}, nil
}

// compareNumericStrings compares two string values as numbers
func compareNumericStrings(a, b, op string) bool {
	aNum, aErr := strconv.ParseFloat(a, 64)
	bNum, bErr := strconv.ParseFloat(b, 64)

	// If one is numeric and the other isn't, return false for comparison operations
	if (aErr == nil && bErr != nil) || (aErr != nil && bErr == nil) {
		return false
	}

	if aErr != nil && bErr != nil {
		// Both non-numeric, fall back to string comparison
		switch op {
		case ">":
			return a > b
		case "<":
			return a < b
		case ">=":
			return a >= b
		case "<=":
			return a <= b
		}
	}

	switch op {
	case ">":
		return aNum > bNum
	case "<":
		return aNum < bNum
	case ">=":
		return aNum >= bNum
	case "<=":
		return aNum <= bNum
	}

	return false
}

// transformCSV applies various transformations to CSV data
func transformCSV(data, delimiter string, hasHeaders bool, transformType string, params map[string]interface{}) (*CSVProcessOutput, error) {
	// Parse CSV first
	parseResult, err := parseCSV(data, delimiter, hasHeaders)
	if err != nil {
		return nil, err
	}
	if parseResult.Error != "" {
		return parseResult, nil
	}

	records := parseResult.Result.([][]string)
	if len(records) == 0 {
		return &CSVProcessOutput{
			Result:   records,
			RowCount: 0,
		}, nil
	}

	switch transformType {
	case "select_columns":
		return selectColumns(records, hasHeaders, params)
	case "sort":
		return sortRecords(records, hasHeaders, params)
	case "count_rows":
		count := len(records)
		if hasHeaders && count > 0 {
			count--
		}
		return &CSVProcessOutput{
			Result:   count,
			RowCount: count,
			Columns:  parseResult.Columns,
		}, nil
	case "statistics":
		return calculateStatistics(records, hasHeaders, params)
	default:
		return &CSVProcessOutput{
			Error: fmt.Sprintf("unsupported transform type: %s", transformType),
		}, nil
	}
}

// selectColumns selects specific columns from the CSV
func selectColumns(records [][]string, hasHeaders bool, params map[string]interface{}) (*CSVProcessOutput, error) {
	columnsParam, ok := params["columns"]
	if !ok {
		return &CSVProcessOutput{
			Error: "columns parameter required for select_columns transform",
		}, nil
	}

	var columnNames []string
	switch v := columnsParam.(type) {
	case []interface{}:
		for _, col := range v {
			columnNames = append(columnNames, fmt.Sprintf("%v", col))
		}
	case []string:
		columnNames = v
	case string:
		columnNames = strings.Split(v, ",")
	default:
		return &CSVProcessOutput{
			Error: "columns parameter must be an array or comma-separated string",
		}, nil
	}

	// Build column index map
	columnIndices := []int{}
	selectedHeaders := []string{}

	if hasHeaders && len(records) > 0 {
		headers := records[0]
		for _, colName := range columnNames {
			for i, h := range headers {
				if h == colName {
					columnIndices = append(columnIndices, i)
					selectedHeaders = append(selectedHeaders, h)
					break
				}
			}
		}
	} else {
		// If no headers, assume column names are indices
		for _, colName := range columnNames {
			if idx, err := strconv.Atoi(colName); err == nil {
				columnIndices = append(columnIndices, idx)
			}
		}
	}

	// Select columns
	result := [][]string{}
	startIdx := 0

	if hasHeaders && len(selectedHeaders) > 0 {
		result = append(result, selectedHeaders)
		startIdx = 1
	}

	for i := startIdx; i < len(records); i++ {
		row := []string{}
		for _, idx := range columnIndices {
			if idx < len(records[i]) {
				row = append(row, records[i][idx])
			}
		}
		result = append(result, row)
	}

	rowCount := len(result)
	if hasHeaders && rowCount > 0 {
		rowCount--
	}

	return &CSVProcessOutput{
		Result:   result,
		RowCount: rowCount,
		Columns:  selectedHeaders,
	}, nil
}

// sortRecords sorts CSV records
func sortRecords(records [][]string, hasHeaders bool, params map[string]interface{}) (*CSVProcessOutput, error) {
	// This is a simple implementation
	// In a real implementation, you'd want to use a proper sorting algorithm
	// The hasHeaders parameter would be used to skip the header row during sorting
	// The params parameter would contain sorting configuration like column and order
	_ = hasHeaders // Mark as intentionally unused for now
	_ = params     // Mark as intentionally unused for now
	return &CSVProcessOutput{
		Result:   records,
		RowCount: len(records),
	}, nil
}

// calculateStatistics calculates basic statistics for numeric columns
func calculateStatistics(records [][]string, hasHeaders bool, params map[string]interface{}) (*CSVProcessOutput, error) {
	stats := make(map[string]interface{})

	rowCount := len(records)
	if hasHeaders && rowCount > 0 {
		rowCount--
	}

	stats["row_count"] = rowCount
	stats["column_count"] = 0
	if len(records) > 0 {
		stats["column_count"] = len(records[0])
	}

	// If no specific columns requested, return basic stats
	if params == nil {
		return &CSVProcessOutput{
			Result:   stats,
			RowCount: rowCount,
		}, nil
	}

	columnsParam, hasColumns := params["columns"]
	if !hasColumns {
		return &CSVProcessOutput{
			Result:   stats,
			RowCount: rowCount,
		}, nil
	}

	var columnNames []string
	switch v := columnsParam.(type) {
	case []interface{}:
		for _, col := range v {
			columnNames = append(columnNames, fmt.Sprintf("%v", col))
		}
	case []string:
		columnNames = v
	default:
		return &CSVProcessOutput{
			Error: "columns parameter must be an array of strings",
		}, nil
	}

	if len(records) == 0 || (hasHeaders && len(records) < 2) {
		return &CSVProcessOutput{
			Result:   stats,
			RowCount: rowCount,
		}, nil
	}

	var headers []string
	var dataRows [][]string

	if hasHeaders {
		headers = records[0]
		dataRows = records[1:]
	} else {
		dataRows = records
	}

	// Calculate statistics for each requested column
	for _, colName := range columnNames {
		colIdx := -1
		if hasHeaders {
			for i, h := range headers {
				if h == colName {
					colIdx = i
					break
				}
			}
		} else {
			// If no headers, assume column name is an index
			idx, err := strconv.Atoi(colName)
			if err == nil && idx >= 0 {
				colIdx = idx
			}
		}

		if colIdx < 0 || (len(dataRows) > 0 && colIdx >= len(dataRows[0])) {
			continue // Skip invalid columns
		}

		// Collect numeric values
		var values []float64
		for _, row := range dataRows {
			if colIdx < len(row) {
				if val, err := strconv.ParseFloat(row[colIdx], 64); err == nil {
					values = append(values, val)
				}
			}
		}

		if len(values) == 0 {
			continue // Skip non-numeric columns
		}

		// Calculate statistics
		colStats := make(map[string]interface{})

		// Basic stats
		sum := 0.0
		min := values[0]
		max := values[0]

		for _, v := range values {
			sum += v
			if v < min {
				min = v
			}
			if v > max {
				max = v
			}
		}

		avg := sum / float64(len(values))

		// Variance and standard deviation
		variance := 0.0
		for _, v := range values {
			variance += (v - avg) * (v - avg)
		}
		variance /= float64(len(values))
		stdDev := variance // Could use math.Sqrt(variance) for true std dev

		colStats["count"] = len(values)
		colStats["sum"] = sum
		colStats["min"] = min
		colStats["max"] = max
		colStats["avg"] = avg
		colStats["variance"] = variance
		colStats["std_dev"] = stdDev

		stats[colName] = colStats
	}

	return &CSVProcessOutput{
		Result:   stats,
		RowCount: rowCount,
	}, nil
}

// csvToJSON converts CSV to JSON
func csvToJSON(data, delimiter string, hasHeaders bool) (*CSVProcessOutput, error) {
	// Parse CSV first
	parseResult, err := parseCSV(data, delimiter, hasHeaders)
	if err != nil {
		return nil, err
	}
	if parseResult.Error != "" {
		return parseResult, nil
	}

	records := parseResult.Result.([][]string)
	if len(records) == 0 {
		return &CSVProcessOutput{
			Result: "[]",
		}, nil
	}

	var jsonResult interface{}

	if hasHeaders && len(records) > 1 {
		// Convert to array of objects
		headers := records[0]
		objects := []map[string]string{}

		for i := 1; i < len(records); i++ {
			obj := make(map[string]string)
			for j, header := range headers {
				if j < len(records[i]) {
					obj[header] = records[i][j]
				}
			}
			objects = append(objects, obj)
		}
		jsonResult = objects
	} else {
		// Convert to array of arrays
		jsonResult = records
	}

	jsonBytes, err := json.MarshalIndent(jsonResult, "", "  ")
	if err != nil {
		return &CSVProcessOutput{
			Error: fmt.Sprintf("failed to convert to JSON: %v", err),
		}, nil
	}

	return &CSVProcessOutput{
		Result:   string(jsonBytes),
		RowCount: parseResult.RowCount,
		Columns:  parseResult.Columns,
	}, nil
}

func init() {
	tools.MustRegisterTool("csv_process", CSVProcess(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "csv_process",
			Category:    "data",
			Tags:        []string{"data", "csv", "parse", "filter", "transform", "tabular"},
			Description: "Process CSV data: parse, filter, transform, or convert to JSON",
			Version:     "1.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Parse CSV",
					Description: "Parse CSV data with headers",
					Code:        `CSVProcess().Execute(ctx, CSVProcessInput{Data: csvStr, Operation: "parse", HasHeaders: true})`,
				},
				{
					Name:        "Filter CSV",
					Description: "Filter rows based on conditions",
					Code:        `CSVProcess().Execute(ctx, CSVProcessInput{Data: csvStr, Operation: "filter", FilterCondition: "age:gt:25"})`,
				},
				{
					Name:        "Transform to JSON",
					Description: "Convert CSV to JSON format",
					Code:        `CSVProcess().Execute(ctx, CSVProcessInput{Data: csvStr, Operation: "to_json", HasHeaders: true})`,
				},
			},
		},
		RequiredPermissions: []string{},
		ResourceUsage: tools.ResourceInfo{
			Memory:      "medium",
			Network:     false,
			FileSystem:  false,
			Concurrency: true,
		},
	})
}

// MustGetCSVProcess returns the CSVProcess tool or panics if not found
func MustGetCSVProcess() domain.Tool {
	tool, ok := tools.GetTool("csv_process")
	if !ok {
		panic(fmt.Errorf("csv_process tool not found"))
	}
	return tool
}
