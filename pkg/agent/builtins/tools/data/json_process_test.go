package data

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestJSONProcess_Parse(t *testing.T) {
	tool := JSONProcess()
	ctx := context.Background()
	toolCtx := createTestToolContext(ctx)

	tests := []struct {
		name      string
		input     JSONProcessInput
		wantError bool
		checkFunc func(t *testing.T, output *JSONProcessOutput)
	}{
		{
			name: "valid JSON object",
			input: JSONProcessInput{
				Data:      `{"name": "John", "age": 30, "city": "New York"}`,
				Operation: "parse",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *JSONProcessOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				if output.ResultType != "map[string]interface {}" {
					t.Errorf("expected map type, got %s", output.ResultType)
				}
			},
		},
		{
			name: "valid JSON array",
			input: JSONProcessInput{
				Data:      `[1, 2, 3, 4, 5]`,
				Operation: "parse",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *JSONProcessOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				if output.ResultType != "[]interface {}" {
					t.Errorf("expected array type, got %s", output.ResultType)
				}
			},
		},
		{
			name: "invalid JSON",
			input: JSONProcessInput{
				Data:      `{"name": "John", "age": }`,
				Operation: "parse",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *JSONProcessOutput) {
				if output.Error == "" {
					t.Error("expected error for invalid JSON")
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
				jsonOutput, ok := output.(*JSONProcessOutput)
				if !ok {
					t.Fatalf("Expected *JSONProcessOutput, got %T", output)
				}
				tt.checkFunc(t, jsonOutput)
			}
		})
	}
}

func TestJSONProcess_Query(t *testing.T) {
	tool := JSONProcess()
	ctx := context.Background()
	toolCtx := createTestToolContext(ctx)

	testData := `{
		"users": [
			{"name": "John", "age": 30, "address": {"city": "New York"}},
			{"name": "Jane", "age": 25, "address": {"city": "Boston"}}
		],
		"count": 2
	}`

	tests := []struct {
		name      string
		input     JSONProcessInput
		wantError bool
		checkFunc func(t *testing.T, output *JSONProcessOutput)
	}{
		{
			name: "query root field",
			input: JSONProcessInput{
				Data:      testData,
				Operation: "query",
				JSONPath:  "$.count",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *JSONProcessOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				if count, ok := output.Result.(float64); !ok || count != 2 {
					t.Errorf("expected count 2, got %v", output.Result)
				}
			},
		},
		{
			name: "query array element",
			input: JSONProcessInput{
				Data:      testData,
				Operation: "query",
				JSONPath:  "users[0]",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *JSONProcessOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				if user, ok := output.Result.(map[string]interface{}); ok {
					if user["name"] != "John" {
						t.Errorf("expected name John, got %v", user["name"])
					}
				} else {
					t.Error("expected map result")
				}
			},
		},
		{
			name: "query nested field",
			input: JSONProcessInput{
				Data:      testData,
				Operation: "query",
				JSONPath:  "users[0].address.city",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *JSONProcessOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				if city, ok := output.Result.(string); !ok || city != "New York" {
					t.Errorf("expected city New York, got %v", output.Result)
				}
			},
		},
		{
			name: "query non-existent field",
			input: JSONProcessInput{
				Data:      testData,
				Operation: "query",
				JSONPath:  "nonexistent",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *JSONProcessOutput) {
				if output.Error == "" {
					t.Error("expected error for non-existent field")
				}
			},
		},
		{
			name: "missing JSONPath",
			input: JSONProcessInput{
				Data:      testData,
				Operation: "query",
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
				jsonOutput, ok := output.(*JSONProcessOutput)
				if !ok {
					t.Fatalf("Expected *JSONProcessOutput, got %T", output)
				}
				tt.checkFunc(t, jsonOutput)
			}
		})
	}
}

func TestJSONProcess_Transform(t *testing.T) {
	tool := JSONProcess()
	ctx := context.Background()
	toolCtx := createTestToolContext(ctx)

	testData := `{
		"user": {
			"name": "John Doe",
			"details": {
				"age": 30,
				"city": "New York"
			}
		},
		"scores": [85, 90, 78]
	}`

	tests := []struct {
		name      string
		input     JSONProcessInput
		wantError bool
		checkFunc func(t *testing.T, output *JSONProcessOutput)
	}{
		{
			name: "extract keys",
			input: JSONProcessInput{
				Data:      testData,
				Operation: "transform",
				Transform: "extract_keys",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *JSONProcessOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				keys, ok := output.Result.([]string)
				if !ok {
					t.Error("expected string array result")
					return
				}
				// Check for some expected keys
				hasUser := false
				hasScores := false
				for _, key := range keys {
					if key == "user" {
						hasUser = true
					}
					if key == "scores" {
						hasScores = true
					}
				}
				if !hasUser || !hasScores {
					t.Errorf("missing expected keys in %v", keys)
				}
			},
		},
		{
			name: "extract values",
			input: JSONProcessInput{
				Data:      `{"a": 1, "b": 2, "c": 3}`,
				Operation: "transform",
				Transform: "extract_values",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *JSONProcessOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				values, ok := output.Result.([]interface{})
				if !ok {
					t.Error("expected array result")
					return
				}
				if len(values) != 3 {
					t.Errorf("expected 3 values, got %d", len(values))
				}
			},
		},
		{
			name: "flatten",
			input: JSONProcessInput{
				Data:      `{"user": {"name": "John", "age": 30}, "active": true}`,
				Operation: "transform",
				Transform: "flatten",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *JSONProcessOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				flattened, ok := output.Result.(map[string]interface{})
				if !ok {
					t.Error("expected map result")
					return
				}
				if flattened["user.name"] != "John" {
					t.Errorf("expected user.name = John, got %v", flattened["user.name"])
				}
				if flattened["user.age"] != float64(30) {
					t.Errorf("expected user.age = 30, got %v", flattened["user.age"])
				}
			},
		},
		{
			name: "prettify",
			input: JSONProcessInput{
				Data:      `{"name":"John","age":30}`,
				Operation: "transform",
				Transform: "prettify",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *JSONProcessOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				pretty, ok := output.Result.(string)
				if !ok {
					t.Error("expected string result")
					return
				}
				if !strings.Contains(pretty, "\n") || !strings.Contains(pretty, "  ") {
					t.Error("expected prettified JSON with newlines and indentation")
				}
			},
		},
		{
			name: "minify",
			input: JSONProcessInput{
				Data: `{
					"name": "John",
					"age": 30
				}`,
				Operation: "transform",
				Transform: "minify",
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *JSONProcessOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				minified, ok := output.Result.(string)
				if !ok {
					t.Error("expected string result")
					return
				}
				if strings.Contains(minified, "\n") || strings.Contains(minified, "  ") {
					t.Error("expected minified JSON without newlines or extra spaces")
				}
				// Verify it's still valid JSON
				var test map[string]interface{}
				if err := json.Unmarshal([]byte(minified), &test); err != nil {
					t.Errorf("minified result is not valid JSON: %v", err)
				}
			},
		},
		{
			name: "missing transform type",
			input: JSONProcessInput{
				Data:      testData,
				Operation: "transform",
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
				jsonOutput, ok := output.(*JSONProcessOutput)
				if !ok {
					t.Fatalf("Expected *JSONProcessOutput, got %T", output)
				}
				tt.checkFunc(t, jsonOutput)
			}
		})
	}
}

func TestJSONProcess_InvalidOperation(t *testing.T) {
	tool := JSONProcess()
	ctx := context.Background()
	toolCtx := createTestToolContext(ctx)

	input := JSONProcessInput{
		Data:      `{"test": "data"}`,
		Operation: "invalid",
	}

	_, err := tool.Execute(toolCtx, input)
	if err == nil {
		t.Error("expected error for invalid operation")
	}
}
