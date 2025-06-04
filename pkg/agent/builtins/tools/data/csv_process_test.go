package data

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestCSVProcess_Parse(t *testing.T) {
	tool := CSVProcess()
	ctx := context.Background()
	toolCtx := createTestToolContext(ctx)

	tests := []struct {
		name      string
		input     CSVProcessInput
		wantError bool
		checkFunc func(t *testing.T, output *CSVProcessOutput)
	}{
		{
			name: "parse with headers",
			input: CSVProcessInput{
				Data:       "name,age,city\nJohn,30,New York\nJane,25,Boston",
				Operation:  "parse",
				HasHeaders: true,
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *CSVProcessOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				if output.RowCount != 2 {
					t.Errorf("expected 2 rows, got %d", output.RowCount)
				}
				if len(output.Columns) != 3 {
					t.Errorf("expected 3 columns, got %d", len(output.Columns))
				}
				if output.Columns[0] != "name" {
					t.Errorf("expected first column 'name', got %s", output.Columns[0])
				}
			},
		},
		{
			name: "parse without headers",
			input: CSVProcessInput{
				Data:       "John,30,New York\nJane,25,Boston",
				Operation:  "parse",
				HasHeaders: false,
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *CSVProcessOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				if output.RowCount != 2 {
					t.Errorf("expected 2 rows, got %d", output.RowCount)
				}
				if len(output.Columns) != 0 {
					t.Errorf("expected no columns, got %d", len(output.Columns))
				}
			},
		},
		{
			name: "parse with custom delimiter",
			input: CSVProcessInput{
				Data:       "name|age|city\nJohn|30|New York",
				Operation:  "parse",
				HasHeaders: true,
				Delimiter:  "|",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *CSVProcessOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				if output.RowCount != 1 {
					t.Errorf("expected 1 row, got %d", output.RowCount)
				}
				records := output.Result.([][]string)
				if len(records) != 2 { // header + 1 data row
					t.Errorf("expected 2 records, got %d", len(records))
				}
			},
		},
		{
			name: "parse empty CSV",
			input: CSVProcessInput{
				Data:       "",
				Operation:  "parse",
				HasHeaders: false,
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *CSVProcessOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				if output.RowCount != 0 {
					t.Errorf("expected 0 rows, got %d", output.RowCount)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := tool.Execute(toolCtx, tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("Execute() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if tt.checkFunc != nil && output != nil {
				csvOutput, ok := output.(*CSVProcessOutput)
				if !ok {
					t.Fatalf("Expected *CSVProcessOutput, got %T", output)
				}
				tt.checkFunc(t, csvOutput)
			}
		})
	}
}

func TestCSVProcess_Filter(t *testing.T) {
	tool := CSVProcess()
	ctx := context.Background()
	toolCtx := createTestToolContext(ctx)

	testData := "name,age,city\nJohn,30,New York\nJane,25,Boston\nBob,30,Chicago"

	tests := []struct {
		name      string
		input     CSVProcessInput
		wantError bool
		checkFunc func(t *testing.T, output *CSVProcessOutput)
	}{
		{
			name: "filter by equality",
			input: CSVProcessInput{
				Data:            testData,
				Operation:       "filter",
				HasHeaders:      true,
				FilterCondition: "age:eq:30",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *CSVProcessOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				if output.RowCount != 2 {
					t.Errorf("expected 2 rows with age 30, got %d", output.RowCount)
				}
			},
		},
		{
			name: "filter by contains",
			input: CSVProcessInput{
				Data:            testData,
				Operation:       "filter",
				HasHeaders:      true,
				FilterCondition: "city:contains:New",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *CSVProcessOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				if output.RowCount != 1 {
					t.Errorf("expected 1 row with city containing 'New', got %d", output.RowCount)
				}
			},
		},
		{
			name: "filter by numeric comparison",
			input: CSVProcessInput{
				Data:            testData,
				Operation:       "filter",
				HasHeaders:      true,
				FilterCondition: "age:gt:25",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *CSVProcessOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				if output.RowCount != 2 {
					t.Errorf("expected 2 rows with age > 25, got %d", output.RowCount)
				}
			},
		},
		{
			name: "filter without headers using column index",
			input: CSVProcessInput{
				Data:            "John,30,New York\nJane,25,Boston",
				Operation:       "filter",
				HasHeaders:      false,
				FilterCondition: "1:eq:30",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *CSVProcessOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				if output.RowCount != 1 {
					t.Errorf("expected 1 row, got %d", output.RowCount)
				}
			},
		},
		{
			name: "missing filter condition",
			input: CSVProcessInput{
				Data:       testData,
				Operation:  "filter",
				HasHeaders: true,
			},
			wantError: true,
		},
		{
			name: "invalid filter condition format",
			input: CSVProcessInput{
				Data:            testData,
				Operation:       "filter",
				HasHeaders:      true,
				FilterCondition: "invalid",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *CSVProcessOutput) {
				if output.Error == "" {
					t.Error("expected error for invalid filter condition")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := tool.Execute(toolCtx, tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("Execute() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if tt.checkFunc != nil && output != nil {
				csvOutput, ok := output.(*CSVProcessOutput)
				if !ok {
					t.Fatalf("Expected *CSVProcessOutput, got %T", output)
				}
				tt.checkFunc(t, csvOutput)
			}
		})
	}
}

func TestCSVProcess_Transform(t *testing.T) {
	tool := CSVProcess()
	ctx := context.Background()
	toolCtx := createTestToolContext(ctx)

	testData := "name,age,city,score\nJohn,30,New York,85\nJane,25,Boston,92\nBob,30,Chicago,78"

	tests := []struct {
		name      string
		input     CSVProcessInput
		wantError bool
		checkFunc func(t *testing.T, output *CSVProcessOutput)
	}{
		{
			name: "select columns",
			input: CSVProcessInput{
				Data:       testData,
				Operation:  "transform",
				HasHeaders: true,
				Transform:  "select_columns",
				Params: map[string]interface{}{
					"columns": []string{"name", "score"},
				},
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *CSVProcessOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				records := output.Result.([][]string)
				if len(records[0]) != 2 {
					t.Errorf("expected 2 columns, got %d", len(records[0]))
				}
				if records[0][0] != "name" || records[0][1] != "score" {
					t.Errorf("unexpected column headers: %v", records[0])
				}
			},
		},
		{
			name: "count rows",
			input: CSVProcessInput{
				Data:       testData,
				Operation:  "transform",
				HasHeaders: true,
				Transform:  "count_rows",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *CSVProcessOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				if count, ok := output.Result.(int); !ok || count != 3 {
					t.Errorf("expected count 3, got %v", output.Result)
				}
			},
		},
		{
			name: "statistics",
			input: CSVProcessInput{
				Data:       testData,
				Operation:  "transform",
				HasHeaders: true,
				Transform:  "statistics",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *CSVProcessOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				stats, ok := output.Result.(map[string]interface{})
				if !ok {
					t.Error("expected map result")
					return
				}
				if stats["row_count"] != 3 {
					t.Errorf("expected row_count 3, got %v", stats["row_count"])
				}
				if stats["column_count"] != 4 {
					t.Errorf("expected column_count 4, got %v", stats["column_count"])
				}
			},
		},
		{
			name: "missing transform type",
			input: CSVProcessInput{
				Data:       testData,
				Operation:  "transform",
				HasHeaders: true,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := tool.Execute(toolCtx, tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("Execute() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if tt.checkFunc != nil && output != nil {
				csvOutput, ok := output.(*CSVProcessOutput)
				if !ok {
					t.Fatalf("Expected *CSVProcessOutput, got %T", output)
				}
				tt.checkFunc(t, csvOutput)
			}
		})
	}
}

func TestCSVProcess_ToJSON(t *testing.T) {
	tool := CSVProcess()
	ctx := context.Background()
	toolCtx := createTestToolContext(ctx)

	tests := []struct {
		name      string
		input     CSVProcessInput
		wantError bool
		checkFunc func(t *testing.T, output *CSVProcessOutput)
	}{
		{
			name: "CSV to JSON with headers",
			input: CSVProcessInput{
				Data:       "name,age,city\nJohn,30,New York\nJane,25,Boston",
				Operation:  "to_json",
				HasHeaders: true,
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *CSVProcessOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				jsonStr, ok := output.Result.(string)
				if !ok {
					t.Error("expected string result")
					return
				}
				// Verify it's valid JSON
				var result []map[string]string
				if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
					t.Errorf("invalid JSON output: %v", err)
				}
				if len(result) != 2 {
					t.Errorf("expected 2 objects, got %d", len(result))
				}
				if result[0]["name"] != "John" {
					t.Errorf("expected first name to be John, got %s", result[0]["name"])
				}
			},
		},
		{
			name: "CSV to JSON without headers",
			input: CSVProcessInput{
				Data:       "John,30,New York\nJane,25,Boston",
				Operation:  "to_json",
				HasHeaders: false,
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *CSVProcessOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				jsonStr, ok := output.Result.(string)
				if !ok {
					t.Error("expected string result")
					return
				}
				// Verify it's valid JSON
				var result [][]string
				if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
					t.Errorf("invalid JSON output: %v", err)
				}
				if len(result) != 2 {
					t.Errorf("expected 2 arrays, got %d", len(result))
				}
			},
		},
		{
			name: "empty CSV to JSON",
			input: CSVProcessInput{
				Data:       "",
				Operation:  "to_json",
				HasHeaders: false,
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *CSVProcessOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				if output.Result != "[]" {
					t.Errorf("expected empty array '[]', got %v", output.Result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := tool.Execute(toolCtx, tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("Execute() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if tt.checkFunc != nil && output != nil {
				csvOutput, ok := output.(*CSVProcessOutput)
				if !ok {
					t.Fatalf("Expected *CSVProcessOutput, got %T", output)
				}
				tt.checkFunc(t, csvOutput)
			}
		})
	}
}

func TestCSVProcess_InvalidOperation(t *testing.T) {
	tool := CSVProcess()
	ctx := context.Background()
	toolCtx := createTestToolContext(ctx)

	input := CSVProcessInput{
		Data:      "test,data",
		Operation: "invalid",
	}

	_, err := tool.Execute(toolCtx, input)
	if err == nil {
		t.Error("expected error for invalid operation")
	}
}

func TestCompareNumeric(t *testing.T) {
	tests := []struct {
		a, b, op string
		want     bool
	}{
		{"10", "5", ">", true},
		{"5", "10", ">", false},
		{"10", "10", ">=", true},
		{"5", "10", "<", true},
		{"10", "5", "<", false},
		{"10", "10", "<=", true},
		{"abc", "5", ">", false}, // non-numeric comparison
	}

	for _, tt := range tests {
		t.Run(strings.Join([]string{tt.a, tt.op, tt.b}, " "), func(t *testing.T) {
			got := compareNumericStrings(tt.a, tt.b, tt.op)
			if got != tt.want {
				t.Errorf("compareNumericStrings(%s, %s, %s) = %v, want %v", tt.a, tt.b, tt.op, got, tt.want)
			}
		})
	}
}
