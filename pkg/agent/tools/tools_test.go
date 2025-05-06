package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
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

// TestFileTools tests the file-related tools
func TestFileTools(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "test-tools-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("WriteFile and ReadFile", func(t *testing.T) {
		// Get the tools
		writeFileTool := WriteFile()
		readFileTool := ReadFile()

		// Create a test file path
		testFilePath := filepath.Join(tempDir, "test.txt")
		testContent := "Hello, world!"

		// Test WriteFile
		writeResult, err := writeFileTool.Execute(context.Background(), WriteFileParams{
			Path:    testFilePath,
			Content: testContent,
		})
		if err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}

		writeFileResult, ok := writeResult.(*WriteFileResult)
		if !ok {
			t.Fatalf("Expected WriteFileResult, got %T", writeResult)
		}

		if !writeFileResult.Success {
			t.Errorf("Expected WriteFile to succeed")
		}

		// Verify the file was written
		content, err := os.ReadFile(testFilePath)
		if err != nil {
			t.Fatalf("Failed to read test file: %v", err)
		}

		if string(content) != testContent {
			t.Errorf("Expected file content '%s', got '%s'", testContent, string(content))
		}

		// Test ReadFile
		readResult, err := readFileTool.Execute(context.Background(), ReadFileParams{
			Path: testFilePath,
		})
		if err != nil {
			t.Fatalf("ReadFile failed: %v", err)
		}

		readFileResult, ok := readResult.(*ReadFileResult)
		if !ok {
			t.Fatalf("Expected ReadFileResult, got %T", readResult)
		}

		if readFileResult.Content != testContent {
			t.Errorf("Expected ReadFile content '%s', got '%s'", testContent, readFileResult.Content)
		}
	})
}

// TestToolRegistry tests registering and retrieving tools from a registry
func TestToolRegistry(t *testing.T) {
	// Create a set of tools
	tools := []domain.Tool{
		WebFetch(),
		ReadFile(),
		WriteFile(),
		ExecuteCommand(),
	}

	// Verify each tool has a unique name
	names := make(map[string]bool)
	for _, tool := range tools {
		name := tool.Name()
		if names[name] {
			t.Fatalf("Duplicate tool name: %s", name)
		}
		names[name] = true

		// Verify each tool has a non-empty description
		if tool.Description() == "" {
			t.Errorf("Empty tool description for %s", name)
		}

		// Verify each tool has a parameter schema
		if tool.ParameterSchema() == nil {
			t.Errorf("Nil parameter schema for %s", name)
		}
	}
}
