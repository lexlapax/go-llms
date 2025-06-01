package tools

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// TestNewTool tests the NewTool function
func TestNewTool(t *testing.T) {
	schema := &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"name": {Type: "string"},
		},
		Required: []string{"name"},
	}

	tool := NewTool("test_tool", "A test tool", func(ctx context.Context, name string) (string, error) {
		return "Hello, " + name, nil
	}, schema)

	if tool.Name() != "test_tool" {
		t.Errorf("Expected tool name to be 'test_tool', got '%s'", tool.Name())
	}

	if tool.Description() != "A test tool" {
		t.Errorf("Expected tool description to be 'A test tool', got '%s'", tool.Description())
	}

	if !reflect.DeepEqual(tool.ParameterSchema(), schema) {
		t.Errorf("Expected parameter schema to match the provided schema")
	}
}

// TestBaseTool_Execute tests the Execute method of BaseTool
func TestBaseTool_Execute(t *testing.T) {
	t.Run("with no params", func(t *testing.T) {
		tool := NewTool("no_params", "A tool with no params", func() string {
			return "Success"
		}, &sdomain.Schema{Type: "object"})

		result, err := tool.Execute(context.Background(), nil)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result != "Success" {
			t.Errorf("Expected result to be 'Success', got '%v'", result)
		}
	})

	t.Run("with string param", func(t *testing.T) {
		tool := NewTool("string_param", "A tool with a string param", func(name string) string {
			return "Hello, " + name
		}, &sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				"name": {Type: "string"},
			},
		})

		result, err := tool.Execute(context.Background(), "World")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result != "Hello, World" {
			t.Errorf("Expected result to be 'Hello, World', got '%v'", result)
		}
	})

	t.Run("with struct param", func(t *testing.T) {
		type Person struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		tool := NewTool("struct_param", "A tool with a struct param", func(p Person) string {
			return fmt.Sprintf("Hello, %s aged %d", p.Name, p.Age)
		}, &sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				"name": {Type: "string"},
				"age":  {Type: "integer"},
			},
		})

		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"name": "Alice",
			"age":  30,
		})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		expected := "Hello, Alice aged 30"
		if result != expected {
			t.Errorf("Expected result to be '%s', got '%v'", expected, result)
		}
	})

	t.Run("with context param", func(t *testing.T) {
		tool := NewTool("context_param", "A tool with context param", func(ctx context.Context, name string) string {
			// Verify we have a valid context
			if ctx == nil {
				return "Context is nil"
			}
			return "Hello, " + name + " with context"
		}, &sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				"name": {Type: "string"},
			},
		})

		result, err := tool.Execute(context.Background(), "World")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result != "Hello, World with context" {
			t.Errorf("Expected result to be 'Hello, World with context', got '%v'", result)
		}
	})

	t.Run("with error return", func(t *testing.T) {
		tool := NewTool("error_return", "A tool that returns an error", func() (string, error) {
			return "", ErrTest
		}, &sdomain.Schema{Type: "object"})

		_, err := tool.Execute(context.Background(), nil)
		if err != ErrTest {
			t.Errorf("Expected ErrTest, got %v", err)
		}
	})
}

// ErrTest is a test error
var ErrTest = NewTestError("test error")

// TestError is a custom error type for testing
type TestError struct {
	msg string
}

// NewTestError creates a new test error
func NewTestError(msg string) error {
	return &TestError{msg: msg}
}

// Error implements the error interface
func (e *TestError) Error() string {
	return e.msg
}

// TestFileTools has been removed - file tools are now tested in pkg/agent/builtins/tools/file/
// The built-in file tools provide enhanced functionality and are the recommended approach

// TestToolRegistry has been removed - common tools are deprecated in favor of built-in tools
// The built-in tools registry is tested in pkg/agent/builtins/tools/
