package data

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestXMLProcess_Parse(t *testing.T) {
	tool := XMLProcess()
	ctx := context.Background()

	tests := []struct {
		name      string
		input     XMLProcessInput
		wantError bool
		checkFunc func(t *testing.T, output *XMLProcessOutput)
	}{
		{
			name: "parse simple XML",
			input: XMLProcessInput{
				Data:              `<person><name>John</name><age>30</age></person>`,
				Operation:         "parse",
				IncludeAttributes: true,
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *XMLProcessOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				if output.RootElement != "person" {
					t.Errorf("expected root element 'person', got %s", output.RootElement)
				}
				result, ok := output.Result.(map[string]interface{})
				if !ok {
					t.Error("expected map result")
					return
				}
				if result["_name"] != "person" {
					t.Errorf("expected _name 'person', got %v", result["_name"])
				}
			},
		},
		{
			name: "parse XML with attributes",
			input: XMLProcessInput{
				Data:              `<person id="123" active="true"><name>John</name></person>`,
				Operation:         "parse",
				IncludeAttributes: true,
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *XMLProcessOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				result, ok := output.Result.(map[string]interface{})
				if !ok {
					t.Error("expected map result")
					return
				}
				attrs, ok := result["_attributes"].(map[string]string)
				if !ok {
					t.Error("expected attributes map")
					return
				}
				if attrs["id"] != "123" {
					t.Errorf("expected id attribute '123', got %s", attrs["id"])
				}
				if attrs["active"] != "true" {
					t.Errorf("expected active attribute 'true', got %s", attrs["active"])
				}
			},
		},
		{
			name: "parse without attributes",
			input: XMLProcessInput{
				Data:              `<person id="123"><name>John</name></person>`,
				Operation:         "parse",
				IncludeAttributes: false,
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *XMLProcessOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				result, ok := output.Result.(map[string]interface{})
				if !ok {
					t.Error("expected map result")
					return
				}
				if _, hasAttrs := result["_attributes"]; hasAttrs {
					t.Error("expected no attributes in result")
				}
			},
		},
		{
			name: "parse nested XML",
			input: XMLProcessInput{
				Data: `<company>
					<name>Tech Corp</name>
					<employees>
						<employee><name>John</name><role>Developer</role></employee>
						<employee><name>Jane</name><role>Manager</role></employee>
					</employees>
				</company>`,
				Operation:         "parse",
				IncludeAttributes: true,
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *XMLProcessOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				result, ok := output.Result.(map[string]interface{})
				if !ok {
					t.Error("expected map result")
					return
				}
				employees, ok := result["employees"].(map[string]interface{})
				if !ok {
					t.Error("expected employees to be a map")
					return
				}
				employeeList, ok := employees["employee"].([]interface{})
				if !ok || len(employeeList) != 2 {
					t.Error("expected employee to be an array of 2 elements")
				}
			},
		},
		{
			name: "parse invalid XML",
			input: XMLProcessInput{
				Data:              `<person><name>John</age>`,
				Operation:         "parse",
				IncludeAttributes: true,
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *XMLProcessOutput) {
				if output.Error == "" {
					t.Error("expected error for invalid XML")
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
				xmlOutput, ok := output.(*XMLProcessOutput)
				if !ok {
					t.Fatalf("Expected *XMLProcessOutput, got %T", output)
				}
				tt.checkFunc(t, xmlOutput)
			}
		})
	}
}

func TestXMLProcess_Query(t *testing.T) {
	tool := XMLProcess()
	ctx := context.Background()

	testData := `<library>
		<book id="1">
			<title>Go Programming</title>
			<author>John Doe</author>
			<year>2023</year>
		</book>
		<book id="2">
			<title>XML Processing</title>
			<author>Jane Smith</author>
			<year>2024</year>
		</book>
	</library>`

	tests := []struct {
		name      string
		input     XMLProcessInput
		wantError bool
		checkFunc func(t *testing.T, output *XMLProcessOutput)
	}{
		{
			name: "query root element",
			input: XMLProcessInput{
				Data:              testData,
				Operation:         "query",
				XPath:             "/library",
				IncludeAttributes: true,
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *XMLProcessOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				result, ok := output.Result.(map[string]interface{})
				if !ok {
					t.Error("expected map result")
					return
				}
				if result["_name"] != "library" {
					t.Errorf("expected _name 'library', got %v", result["_name"])
				}
			},
		},
		{
			name: "query specific element",
			input: XMLProcessInput{
				Data:              testData,
				Operation:         "query",
				XPath:             "book",
				IncludeAttributes: true,
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *XMLProcessOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				// Should return array of books
				books, ok := output.Result.([]interface{})
				if !ok {
					t.Error("expected array result")
					return
				}
				if len(books) != 2 {
					t.Errorf("expected 2 books, got %d", len(books))
				}
			},
		},
		{
			name: "query nested element",
			input: XMLProcessInput{
				Data:              testData,
				Operation:         "query",
				XPath:             "book/title",
				IncludeAttributes: false,
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *XMLProcessOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				titles, ok := output.Result.([]interface{})
				if !ok {
					t.Error("expected array result")
					return
				}
				if len(titles) != 2 {
					t.Errorf("expected 2 titles, got %d", len(titles))
				}
			},
		},
		{
			name: "query attribute",
			input: XMLProcessInput{
				Data:              `<book id="123" isbn="978-0-123456-78-9"><title>Test</title></book>`,
				Operation:         "query",
				XPath:             "book/@id",
				IncludeAttributes: true,
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *XMLProcessOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				ids, ok := output.Result.([]string)
				if !ok {
					t.Error("expected string array result")
					return
				}
				if len(ids) != 1 || ids[0] != "123" {
					t.Errorf("expected id '123', got %v", ids)
				}
			},
		},
		{
			name: "query non-existent element",
			input: XMLProcessInput{
				Data:              testData,
				Operation:         "query",
				XPath:             "nonexistent",
				IncludeAttributes: true,
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *XMLProcessOutput) {
				if output.Error == "" {
					t.Error("expected error for non-existent element")
				}
			},
		},
		{
			name: "missing XPath",
			input: XMLProcessInput{
				Data:              testData,
				Operation:         "query",
				IncludeAttributes: true,
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
				xmlOutput, ok := output.(*XMLProcessOutput)
				if !ok {
					t.Fatalf("Expected *XMLProcessOutput, got %T", output)
				}
				tt.checkFunc(t, xmlOutput)
			}
		})
	}
}

func TestXMLProcess_ToJSON(t *testing.T) {
	tool := XMLProcess()
	ctx := context.Background()

	tests := []struct {
		name      string
		input     XMLProcessInput
		wantError bool
		checkFunc func(t *testing.T, output *XMLProcessOutput)
	}{
		{
			name: "simple XML to JSON",
			input: XMLProcessInput{
				Data:              `<person><name>John</name><age>30</age></person>`,
				Operation:         "to_json",
				IncludeAttributes: true,
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *XMLProcessOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				jsonStr, ok := output.Result.(string)
				if !ok {
					t.Error("expected string result")
					return
				}
				// Verify it's valid JSON
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
					t.Errorf("invalid JSON output: %v", err)
				}
				if !strings.Contains(jsonStr, "\"_name\"") {
					t.Error("expected JSON to contain _name field")
				}
			},
		},
		{
			name: "XML with attributes to JSON",
			input: XMLProcessInput{
				Data:              `<person id="123"><name>John</name></person>`,
				Operation:         "to_json",
				IncludeAttributes: true,
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *XMLProcessOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				jsonStr, ok := output.Result.(string)
				if !ok {
					t.Error("expected string result")
					return
				}
				if !strings.Contains(jsonStr, "\"_attributes\"") {
					t.Error("expected JSON to contain _attributes field")
				}
				if !strings.Contains(jsonStr, "\"id\"") {
					t.Error("expected JSON to contain id attribute")
				}
			},
		},
		{
			name: "XML to JSON without attributes",
			input: XMLProcessInput{
				Data:              `<person id="123"><name>John</name></person>`,
				Operation:         "to_json",
				IncludeAttributes: false,
			},
			wantError: false,
			checkFunc: func(t *testing.T, output *XMLProcessOutput) {
				if output.Error != "" {
					t.Errorf("unexpected error: %s", output.Error)
				}
				jsonStr, ok := output.Result.(string)
				if !ok {
					t.Error("expected string result")
					return
				}
				if strings.Contains(jsonStr, "\"_attributes\"") {
					t.Error("expected JSON to not contain _attributes field")
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
				xmlOutput, ok := output.(*XMLProcessOutput)
				if !ok {
					t.Fatalf("Expected *XMLProcessOutput, got %T", output)
				}
				tt.checkFunc(t, xmlOutput)
			}
		})
	}
}

func TestXMLProcess_InvalidOperation(t *testing.T) {
	tool := XMLProcess()
	ctx := context.Background()

	input := XMLProcessInput{
		Data:      `<test>data</test>`,
		Operation: "invalid",
	}

	_, err := tool.Execute(ctx, input)
	if err == nil {
		t.Error("expected error for invalid operation")
	}
}
