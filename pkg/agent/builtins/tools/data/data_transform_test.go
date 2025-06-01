package data

import (
	"context"
	"testing"
)

func TestDataTransform_Filter(t *testing.T) {
	tool := DataTransform()
	ctx := context.Background()

	testData := `[
		{"name": "John", "age": 30, "city": "New York"},
		{"name": "Jane", "age": 25, "city": "Boston"},
		{"name": "Bob", "age": 30, "city": "Chicago"}
	]`

	tests := []struct {
		name      string
		input     DataTransformInput
		wantError bool
		checkFunc func(t *testing.T, output *DataTransformOutput)
	}{
		{
			name: "filter by equality",
			input: DataTransformInput{
				Data:      testData,
				Operation: "filter",
				Field:     "age",
				Condition: "eq:30",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *DataTransformOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				if output.ItemCount != 2 {
					t.Errorf("expected 2 items, got %d", output.ItemCount)
				}
			},
		},
		{
			name: "filter by contains",
			input: DataTransformInput{
				Data:      testData,
				Operation: "filter",
				Field:     "city",
				Condition: "contains:New",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *DataTransformOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				if output.ItemCount != 1 {
					t.Errorf("expected 1 item, got %d", output.ItemCount)
				}
			},
		},
		{
			name: "filter by numeric comparison",
			input: DataTransformInput{
				Data:      testData,
				Operation: "filter",
				Field:     "age",
				Condition: "gt:25",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *DataTransformOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				if output.ItemCount != 2 {
					t.Errorf("expected 2 items, got %d", output.ItemCount)
				}
			},
		},
		{
			name: "filter by field exists",
			input: DataTransformInput{
				Data:      testData,
				Operation: "filter",
				Field:     "city",
				Condition: "exists:true",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *DataTransformOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				if output.ItemCount != 3 {
					t.Errorf("expected 3 items, got %d", output.ItemCount)
				}
			},
		},
		{
			name: "filter simple array",
			input: DataTransformInput{
				Data:      `[1, 2, 3, 4, 5]`,
				Operation: "filter",
				Field:     "",
				Condition: "gt:3",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *DataTransformOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				if output.ItemCount != 2 {
					t.Errorf("expected 2 items, got %d", output.ItemCount)
				}
			},
		},
		{
			name: "missing condition",
			input: DataTransformInput{
				Data:      testData,
				Operation: "filter",
				Field:     "age",
			},
			wantError: true,
		},
		{
			name: "invalid condition format",
			input: DataTransformInput{
				Data:      testData,
				Operation: "filter",
				Field:     "age",
				Condition: "invalid",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *DataTransformOutput) {
				if output.Error == "" {
					t.Error("expected error for invalid condition")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := tool.Execute(ctx, tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("Execute() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if tt.checkFunc != nil && output != nil {
				dtOutput, ok := output.(*DataTransformOutput)
				if !ok {
					t.Fatalf("Expected *DataTransformOutput, got %T", output)
				}
				tt.checkFunc(t, dtOutput)
			}
		})
	}
}

func TestDataTransform_Map(t *testing.T) {
	tool := DataTransform()
	ctx := context.Background()

	testData := `[
		{"name": "John", "age": 30},
		{"name": "Jane", "age": 25}
	]`

	tests := []struct {
		name      string
		input     DataTransformInput
		wantError bool
		checkFunc func(t *testing.T, output *DataTransformOutput)
	}{
		{
			name: "extract field",
			input: DataTransformInput{
				Data:      testData,
				Operation: "map",
				Field:     "name",
				MapType:   "extract_field",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *DataTransformOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				result, ok := output.Result.([]interface{})
				if !ok {
					t.Error("expected array result")
					return
				}
				if len(result) != 2 {
					t.Errorf("expected 2 items, got %d", len(result))
				}
				if result[0] != "John" || result[1] != "Jane" {
					t.Errorf("unexpected values: %v", result)
				}
			},
		},
		{
			name: "to upper case",
			input: DataTransformInput{
				Data:      `["hello", "world"]`,
				Operation: "map",
				MapType:   "to_upper",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *DataTransformOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				result, ok := output.Result.([]interface{})
				if !ok {
					t.Error("expected array result")
					return
				}
				if result[0] != "HELLO" || result[1] != "WORLD" {
					t.Errorf("expected uppercase values, got %v", result)
				}
			},
		},
		{
			name: "to number",
			input: DataTransformInput{
				Data:      `["123", "456", "abc"]`,
				Operation: "map",
				MapType:   "to_number",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *DataTransformOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				result, ok := output.Result.([]interface{})
				if !ok {
					t.Error("expected array result")
					return
				}
				if result[0] != float64(123) || result[1] != float64(456) {
					t.Errorf("expected numeric values, got %v", result[0:2])
				}
				if result[2] != float64(0) {
					t.Errorf("expected 0 for non-numeric string, got %v", result[2])
				}
			},
		},
		{
			name: "missing map type",
			input: DataTransformInput{
				Data:      testData,
				Operation: "map",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := tool.Execute(ctx, tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("Execute() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if tt.checkFunc != nil && output != nil {
				dtOutput, ok := output.(*DataTransformOutput)
				if !ok {
					t.Fatalf("Expected *DataTransformOutput, got %T", output)
				}
				tt.checkFunc(t, dtOutput)
			}
		})
	}
}

func TestDataTransform_Reduce(t *testing.T) {
	tool := DataTransform()
	ctx := context.Background()

	tests := []struct {
		name      string
		input     DataTransformInput
		wantError bool
		checkFunc func(t *testing.T, output *DataTransformOutput)
	}{
		{
			name: "sum numbers",
			input: DataTransformInput{
				Data:       `[1, 2, 3, 4, 5]`,
				Operation:  "reduce",
				ReduceType: "sum",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *DataTransformOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				if sum, ok := output.Result.(float64); !ok || sum != 15 {
					t.Errorf("expected sum 15, got %v", output.Result)
				}
			},
		},
		{
			name: "count items",
			input: DataTransformInput{
				Data:       `[{"a": 1}, {"b": 2}, {"c": 3}]`,
				Operation:  "reduce",
				ReduceType: "count",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *DataTransformOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				if count, ok := output.Result.(int); !ok || count != 3 {
					t.Errorf("expected count 3, got %v", output.Result)
				}
			},
		},
		{
			name: "find min",
			input: DataTransformInput{
				Data:       `[5, 2, 8, 1, 9]`,
				Operation:  "reduce",
				ReduceType: "min",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *DataTransformOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				if min, ok := output.Result.(float64); !ok || min != 1 {
					t.Errorf("expected min 1, got %v", output.Result)
				}
			},
		},
		{
			name: "find max",
			input: DataTransformInput{
				Data:       `[5, 2, 8, 1, 9]`,
				Operation:  "reduce",
				ReduceType: "max",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *DataTransformOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				if max, ok := output.Result.(float64); !ok || max != 9 {
					t.Errorf("expected max 9, got %v", output.Result)
				}
			},
		},
		{
			name: "calculate average",
			input: DataTransformInput{
				Data:       `[10, 20, 30]`,
				Operation:  "reduce",
				ReduceType: "average",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *DataTransformOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				if avg, ok := output.Result.(float64); !ok || avg != 20 {
					t.Errorf("expected average 20, got %v", output.Result)
				}
			},
		},
		{
			name: "concatenate strings",
			input: DataTransformInput{
				Data:       `["Hello", "World", "!"]`,
				Operation:  "reduce",
				ReduceType: "concat",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *DataTransformOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				if concat, ok := output.Result.(string); !ok || concat != "Hello, World, !" {
					t.Errorf("expected 'Hello, World, !', got %v", output.Result)
				}
			},
		},
		{
			name: "missing reduce type",
			input: DataTransformInput{
				Data:      `[1, 2, 3]`,
				Operation: "reduce",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := tool.Execute(ctx, tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("Execute() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if tt.checkFunc != nil && output != nil {
				dtOutput, ok := output.(*DataTransformOutput)
				if !ok {
					t.Fatalf("Expected *DataTransformOutput, got %T", output)
				}
				tt.checkFunc(t, dtOutput)
			}
		})
	}
}

func TestDataTransform_Sort(t *testing.T) {
	tool := DataTransform()
	ctx := context.Background()

	tests := []struct {
		name      string
		input     DataTransformInput
		wantError bool
		checkFunc func(t *testing.T, output *DataTransformOutput)
	}{
		{
			name: "sort numbers ascending",
			input: DataTransformInput{
				Data:      `[3, 1, 4, 1, 5, 9, 2, 6]`,
				Operation: "sort",
				SortOrder: "asc",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *DataTransformOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				result, ok := output.Result.([]interface{})
				if !ok {
					t.Error("expected array result")
					return
				}
				// Check first and last elements
				if result[0] != float64(1) || result[len(result)-1] != float64(9) {
					t.Errorf("unexpected sort order: %v", result)
				}
			},
		},
		{
			name: "sort objects by field",
			input: DataTransformInput{
				Data: `[
					{"name": "Charlie", "age": 35},
					{"name": "Alice", "age": 25},
					{"name": "Bob", "age": 30}
				]`,
				Operation: "sort",
				Field:     "name",
				SortOrder: "asc",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *DataTransformOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				result, ok := output.Result.([]interface{})
				if !ok {
					t.Error("expected array result")
					return
				}
				// Check first name
				if firstItem, ok := result[0].(map[string]interface{}); ok {
					if firstItem["name"] != "Alice" {
						t.Errorf("expected first name to be Alice, got %v", firstItem["name"])
					}
				}
			},
		},
		{
			name: "sort descending",
			input: DataTransformInput{
				Data:      `[1, 2, 3, 4, 5]`,
				Operation: "sort",
				SortOrder: "desc",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *DataTransformOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				result, ok := output.Result.([]interface{})
				if !ok {
					t.Error("expected array result")
					return
				}
				if result[0] != float64(5) || result[len(result)-1] != float64(1) {
					t.Errorf("unexpected sort order: %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := tool.Execute(ctx, tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("Execute() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if tt.checkFunc != nil && output != nil {
				dtOutput, ok := output.(*DataTransformOutput)
				if !ok {
					t.Fatalf("Expected *DataTransformOutput, got %T", output)
				}
				tt.checkFunc(t, dtOutput)
			}
		})
	}
}

func TestDataTransform_GroupBy(t *testing.T) {
	tool := DataTransform()
	ctx := context.Background()

	input := DataTransformInput{
		Data: `[
			{"category": "fruit", "name": "apple"},
			{"category": "fruit", "name": "banana"},
			{"category": "vegetable", "name": "carrot"}
		]`,
		Operation: "group_by",
		Field:     "category",
	}

	outputRaw, err := tool.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output, ok := outputRaw.(*DataTransformOutput)
	if !ok {
		t.Fatalf("Expected *DataTransformOutput, got %T", outputRaw)
	}

	if output.Error != "" {
		t.Errorf("unexpected error: %s", output.Error)
	}

	groups, ok := output.Result.(map[string]interface{})
	if !ok {
		t.Error("expected map result")
		return
	}

	if len(groups) != 2 {
		t.Errorf("expected 2 groups, got %d", len(groups))
	}

	fruitGroup, ok := groups["fruit"].([]interface{})
	if !ok || len(fruitGroup) != 2 {
		t.Error("expected fruit group with 2 items")
	}

	vegGroup, ok := groups["vegetable"].([]interface{})
	if !ok || len(vegGroup) != 1 {
		t.Error("expected vegetable group with 1 item")
	}
}

func TestDataTransform_Unique(t *testing.T) {
	tool := DataTransform()
	ctx := context.Background()

	tests := []struct {
		name      string
		input     DataTransformInput
		wantError bool
		checkFunc func(t *testing.T, output *DataTransformOutput)
	}{
		{
			name: "unique simple values",
			input: DataTransformInput{
				Data:      `[1, 2, 2, 3, 3, 3, 4]`,
				Operation: "unique",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *DataTransformOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				if output.ItemCount != 4 {
					t.Errorf("expected 4 unique items, got %d", output.ItemCount)
				}
			},
		},
		{
			name: "unique field values",
			input: DataTransformInput{
				Data: `[
					{"id": 1, "name": "John"},
					{"id": 2, "name": "Jane"},
					{"id": 3, "name": "John"}
				]`,
				Operation: "unique",
				Field:     "name",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *DataTransformOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				if output.ItemCount != 2 {
					t.Errorf("expected 2 unique names, got %d", output.ItemCount)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := tool.Execute(ctx, tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("Execute() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if tt.checkFunc != nil && output != nil {
				dtOutput, ok := output.(*DataTransformOutput)
				if !ok {
					t.Fatalf("Expected *DataTransformOutput, got %T", output)
				}
				tt.checkFunc(t, dtOutput)
			}
		})
	}
}

func TestDataTransform_Reverse(t *testing.T) {
	tool := DataTransform()
	ctx := context.Background()

	input := DataTransformInput{
		Data:      `[1, 2, 3, 4, 5]`,
		Operation: "reverse",
	}

	outputRaw, err := tool.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output, ok := outputRaw.(*DataTransformOutput)
	if !ok {
		t.Fatalf("Expected *DataTransformOutput, got %T", outputRaw)
	}

	if output.Error != "" {
		t.Errorf("unexpected error: %s", output.Error)
	}

	result, ok := output.Result.([]interface{})
	if !ok {
		t.Error("expected array result")
		return
	}

	if len(result) != 5 {
		t.Errorf("expected 5 items, got %d", len(result))
	}

	if result[0] != float64(5) || result[4] != float64(1) {
		t.Errorf("expected reversed array [5,4,3,2,1], got %v", result)
	}
}

func TestDataTransform_InvalidData(t *testing.T) {
	tool := DataTransform()
	ctx := context.Background()

	input := DataTransformInput{
		Data:      `invalid json`,
		Operation: "filter",
		Field:     "test",
		Condition: "eq:test",
	}

	outputRaw, err := tool.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output, ok := outputRaw.(*DataTransformOutput)
	if !ok {
		t.Fatalf("Expected *DataTransformOutput, got %T", outputRaw)
	}

	if output.Error == "" {
		t.Error("expected error for invalid JSON data")
	}
}

func TestDataTransform_InvalidOperation(t *testing.T) {
	tool := DataTransform()
	ctx := context.Background()

	input := DataTransformInput{
		Data:      `[1, 2, 3]`,
		Operation: "invalid",
	}

	_, err := tool.Execute(ctx, input)
	if err == nil {
		t.Error("expected error for invalid operation")
	}
}

func TestDataTransform_NestedFieldAccess(t *testing.T) {
	tool := DataTransform()
	ctx := context.Background()

	input := DataTransformInput{
		Data: `[
			{"user": {"name": "John", "address": {"city": "New York"}}},
			{"user": {"name": "Jane", "address": {"city": "Boston"}}}
		]`,
		Operation: "map",
		Field:     "user.address.city",
		MapType:   "extract_field",
	}

	outputRaw, err := tool.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output, ok := outputRaw.(*DataTransformOutput)
	if !ok {
		t.Fatalf("Expected *DataTransformOutput, got %T", outputRaw)
	}

	if output.Error != "" {
		t.Errorf("unexpected error: %s", output.Error)
	}

	result, ok := output.Result.([]interface{})
	if !ok {
		t.Error("expected array result")
		return
	}

	if len(result) != 2 {
		t.Errorf("expected 2 items, got %d", len(result))
	}

	if result[0] != "New York" || result[1] != "Boston" {
		t.Errorf("expected cities [New York, Boston], got %v", result)
	}
}

func TestDataTransform_SingleItemToArray(t *testing.T) {
	tool := DataTransform()
	ctx := context.Background()

	input := DataTransformInput{
		Data:      `{"name": "John", "age": 30}`,
		Operation: "filter",
		Field:     "age",
		Condition: "eq:30",
	}

	outputRaw, err := tool.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output, ok := outputRaw.(*DataTransformOutput)
	if !ok {
		t.Fatalf("Expected *DataTransformOutput, got %T", outputRaw)
	}

	if output.Error != "" {
		t.Errorf("unexpected error: %s", output.Error)
	}

	// Single item should be converted to array and filtered
	if output.ItemCount != 1 {
		t.Errorf("expected 1 item, got %d", output.ItemCount)
	}
}
