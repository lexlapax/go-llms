package processor

import (
	"reflect"
	"testing"

	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/schema/validation"
)

func float64Ptr(v float64) *float64 {
	return &v
}

func intPtr(v int) *int {
	return &v
}

func TestProcessStructuredOutput(t *testing.T) {
	validator := validation.NewValidator()
	processor := NewStructuredProcessor(validator)

	t.Run("valid JSON object", func(t *testing.T) {
		schema := &schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"name": {Type: "string"},
				"age":  {Type: "integer", Minimum: float64Ptr(0)},
			},
			Required: []string{"name"},
		}

		output := `{"name": "John Doe", "age": 30}`
		result, err := processor.Process(schema, output)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map result, got %T", result)
		}

		if resultMap["name"] != "John Doe" {
			t.Errorf("Expected name 'John Doe', got '%v'", resultMap["name"])
		}

		if resultMap["age"] != float64(30) {
			t.Errorf("Expected age 30, got '%v'", resultMap["age"])
		}
	})

	t.Run("valid JSON with extra text", func(t *testing.T) {
		schema := &schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"name": {Type: "string"},
				"age":  {Type: "integer", Minimum: float64Ptr(0)},
			},
			Required: []string{"name"},
		}

		output := `Here's the information about the person:
		
		{"name": "John Doe", "age": 30}
		
		Hope this helps!`
		result, err := processor.Process(schema, output)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map result, got %T", result)
		}

		if resultMap["name"] != "John Doe" {
			t.Errorf("Expected name 'John Doe', got '%v'", resultMap["name"])
		}

		if resultMap["age"] != float64(30) {
			t.Errorf("Expected age 30, got '%v'", resultMap["age"])
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		schema := &schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"name": {Type: "string"},
				"age":  {Type: "integer", Minimum: float64Ptr(0)},
			},
			Required: []string{"name"},
		}

		output := `This is not valid JSON`
		_, err := processor.Process(schema, output)

		if err == nil {
			t.Fatal("Expected error for invalid JSON, got nil")
		}
	})

	t.Run("valid JSON array", func(t *testing.T) {
		schema := &schemaDomain.Schema{
			Type: "array",
			Properties: map[string]schemaDomain.Property{
				"": {
					Type: "array",
					Items: &schemaDomain.Property{
						Type: "string",
					},
				},
			},
		}

		output := `["item1", "item2", "item3"]`
		result, err := processor.Process(schema, output)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		resultArray, ok := result.([]interface{})
		if !ok {
			t.Fatalf("Expected array result, got %T", result)
		}

		if len(resultArray) != 3 {
			t.Errorf("Expected 3 items, got %d", len(resultArray))
		}

		if resultArray[0] != "item1" {
			t.Errorf("Expected first item 'item1', got '%v'", resultArray[0])
		}
	})

	t.Run("JSON fails validation", func(t *testing.T) {
		schema := &schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"name": {Type: "string", MinLength: intPtr(5)},
				"age":  {Type: "integer", Minimum: float64Ptr(18)},
			},
			Required: []string{"name", "age"},
		}

		output := `{"name": "John", "age": 15}`
		_, err := processor.Process(schema, output)

		if err == nil {
			t.Fatal("Expected validation error, got nil")
		}
	})

	t.Run("JSON with code block", func(t *testing.T) {
		schema := &schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"name": {Type: "string"},
				"age":  {Type: "integer", Minimum: float64Ptr(0)},
			},
			Required: []string{"name"},
		}

		output := "```json\n{\"name\": \"John Doe\", \"age\": 30}\n```"
		result, err := processor.Process(schema, output)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map result, got %T", result)
		}

		if resultMap["name"] != "John Doe" {
			t.Errorf("Expected name 'John Doe', got '%v'", resultMap["name"])
		}
	})
}

// Define Person struct for testing ProcessTyped
type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestProcessTyped(t *testing.T) {
	validator := validation.NewValidator()
	processor := NewStructuredProcessor(validator)

	t.Run("process to struct", func(t *testing.T) {
		schema := &schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"name": {Type: "string"},
				"age":  {Type: "integer", Minimum: float64Ptr(0)},
			},
			Required: []string{"name"},
		}

		output := `{"name": "John Doe", "age": 30}`
		var person Person
		err := processor.ProcessTyped(schema, output, &person)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if person.Name != "John Doe" {
			t.Errorf("Expected name 'John Doe', got '%s'", person.Name)
		}

		if person.Age != 30 {
			t.Errorf("Expected age 30, got %d", person.Age)
		}
	})

	t.Run("process to slice", func(t *testing.T) {
		schema := &schemaDomain.Schema{
			Type: "array",
			Properties: map[string]schemaDomain.Property{
				"": {
					Type: "array",
					Items: &schemaDomain.Property{
						Type: "string",
					},
				},
			},
		}

		output := `["item1", "item2", "item3"]`
		var items []string
		err := processor.ProcessTyped(schema, output, &items)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		expected := []string{"item1", "item2", "item3"}
		if !reflect.DeepEqual(items, expected) {
			t.Errorf("Expected %v, got %v", expected, items)
		}
	})

	t.Run("invalid target", func(t *testing.T) {
		schema := &schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"name": {Type: "string"},
				"age":  {Type: "integer", Minimum: float64Ptr(0)},
			},
			Required: []string{"name"},
		}

		output := `{"name": "John Doe", "age": 30}`
		var person string // Not a pointer
		err := processor.ProcessTyped(schema, output, person)

		if err == nil {
			t.Fatal("Expected error for non-pointer target, got nil")
		}
	})
}
