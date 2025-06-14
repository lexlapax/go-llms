package fixtures

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/go-llms/pkg/testutils/helpers"
)

func TestCalculatorMockTool(t *testing.T) {
	calc := CalculatorMockTool()
	assert.NotNil(t, calc)
	assert.Equal(t, "calculator", calc.Name())

	ctx := helpers.CreateTestToolContext()

	t.Run("addition", func(t *testing.T) {
		input := map[string]interface{}{
			"operation": "add",
			"a":         5.0,
			"b":         3.0,
		}

		result, err := calc.Execute(ctx, input)
		assert.NoError(t, err)

		response, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "addition", response["operation"])
		assert.Equal(t, 5.0, response["a"])
		assert.Equal(t, 3.0, response["b"])
		assert.Equal(t, 8.0, response["result"])
	})

	t.Run("subtraction", func(t *testing.T) {
		input := map[string]interface{}{
			"operation": "subtract",
			"a":         10.0,
			"b":         4.0,
		}

		result, err := calc.Execute(ctx, input)
		assert.NoError(t, err)

		response, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "subtraction", response["operation"])
		assert.Equal(t, 6.0, response["result"])
	})

	t.Run("multiplication", func(t *testing.T) {
		input := map[string]interface{}{
			"operation": "multiply",
			"a":         6.0,
			"b":         7.0,
		}

		result, err := calc.Execute(ctx, input)
		assert.NoError(t, err)

		response, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "multiplication", response["operation"])
		assert.Equal(t, 42.0, response["result"])
	})

	t.Run("division", func(t *testing.T) {
		input := map[string]interface{}{
			"operation": "divide",
			"a":         15.0,
			"b":         3.0,
		}

		result, err := calc.Execute(ctx, input)
		assert.NoError(t, err)

		response, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "division", response["operation"])
		assert.Equal(t, 5.0, response["result"])
	})

	t.Run("division by zero", func(t *testing.T) {
		input := map[string]interface{}{
			"operation": "divide",
			"a":         10.0,
			"b":         0.0,
		}

		_, err := calc.Execute(ctx, input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "division by zero")
	})

	t.Run("string numbers", func(t *testing.T) {
		input := map[string]interface{}{
			"operation": "add",
			"a":         "5",
			"b":         "3",
		}

		result, err := calc.Execute(ctx, input)
		assert.NoError(t, err)

		response, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, 8.0, response["result"])
	})

	t.Run("unsupported operation", func(t *testing.T) {
		input := map[string]interface{}{
			"operation": "square_root",
			"a":         9.0,
		}

		result, err := calc.Execute(ctx, input)
		assert.NoError(t, err)

		response, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, response["error"].(string), "Unsupported operation")
	})
}

func TestWebSearchMockTool(t *testing.T) {
	search := WebSearchMockTool()
	assert.NotNil(t, search)
	assert.Equal(t, "web_search", search.Name())

	ctx := helpers.CreateTestToolContext()

	t.Run("weather search", func(t *testing.T) {
		input := map[string]interface{}{
			"query": "weather forecast today",
		}

		result, err := search.Execute(ctx, input)
		assert.NoError(t, err)

		response, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "weather forecast today", response["query"])

		results, ok := response["results"].([]map[string]interface{})
		require.True(t, ok)
		assert.Len(t, results, 2)
		assert.Contains(t, results[0]["title"].(string), "Weather")
		assert.Equal(t, 2, response["total_results"])
	})

	t.Run("programming search", func(t *testing.T) {
		input := map[string]interface{}{
			"query": "programming best practices",
		}

		result, err := search.Execute(ctx, input)
		assert.NoError(t, err)

		response, ok := result.(map[string]interface{})
		require.True(t, ok)

		results, ok := response["results"].([]map[string]interface{})
		require.True(t, ok)
		assert.Len(t, results, 2)
		assert.Contains(t, results[0]["title"].(string), "Programming")
	})

	t.Run("generic search", func(t *testing.T) {
		input := map[string]interface{}{
			"query": "random topic",
		}

		result, err := search.Execute(ctx, input)
		assert.NoError(t, err)

		response, ok := result.(map[string]interface{})
		require.True(t, ok)

		results, ok := response["results"].([]map[string]interface{})
		require.True(t, ok)
		assert.Len(t, results, 1)
		assert.Contains(t, results[0]["title"].(string), "random topic")
		assert.Equal(t, 1, response["total_results"])
	})
}

func TestFileMockTool(t *testing.T) {
	file := FileMockTool()
	assert.NotNil(t, file)
	assert.Equal(t, "file_manager", file.Name())

	ctx := helpers.CreateTestToolContext()

	t.Run("read existing file", func(t *testing.T) {
		input := map[string]interface{}{
			"operation": "read",
			"path":      "/etc/config.txt",
		}

		result, err := file.Execute(ctx, input)
		assert.NoError(t, err)

		response, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "read", response["operation"])
		assert.Equal(t, "/etc/config.txt", response["path"])
		assert.Contains(t, response["content"].(string), "Configuration file")
		assert.Greater(t, response["size"].(int), 0)
	})

	t.Run("read non-existent file", func(t *testing.T) {
		input := map[string]interface{}{
			"operation": "read",
			"path":      "/non/existent/file.txt",
		}

		_, err := file.Execute(ctx, input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file not found")
	})

	t.Run("write new file", func(t *testing.T) {
		input := map[string]interface{}{
			"operation": "write",
			"path":      "/tmp/new_file.txt",
			"content":   "Hello, World!",
		}

		result, err := file.Execute(ctx, input)
		assert.NoError(t, err)

		response, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "write", response["operation"])
		assert.Equal(t, "/tmp/new_file.txt", response["path"])
		assert.Equal(t, 13, response["size"])
		assert.Equal(t, "success", response["status"])
	})

	t.Run("read written file", func(t *testing.T) {
		// First write
		writeInput := map[string]interface{}{
			"operation": "write",
			"path":      "/tmp/test.txt",
			"content":   "Test content",
		}
		_, err := file.Execute(ctx, writeInput)
		assert.NoError(t, err)

		// Then read
		readInput := map[string]interface{}{
			"operation": "read",
			"path":      "/tmp/test.txt",
		}

		result, err := file.Execute(ctx, readInput)
		assert.NoError(t, err)

		response, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Test content", response["content"])
	})

	t.Run("list files", func(t *testing.T) {
		input := map[string]interface{}{
			"operation": "list",
		}

		result, err := file.Execute(ctx, input)
		assert.NoError(t, err)

		response, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "list", response["operation"])

		files, ok := response["files"].([]map[string]interface{})
		require.True(t, ok)
		assert.GreaterOrEqual(t, len(files), 3) // At least the 3 default files
		assert.Equal(t, len(files), response["count"])
	})

	t.Run("unsupported operation", func(t *testing.T) {
		input := map[string]interface{}{
			"operation": "delete",
			"path":      "/tmp/file.txt",
		}

		result, err := file.Execute(ctx, input)
		assert.NoError(t, err)

		response, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, response["error"].(string), "Unsupported file operation")
	})
}

func TestErrorMockTool(t *testing.T) {
	t.Run("no errors (rate 0.0)", func(t *testing.T) {
		tool := ErrorMockTool(0.0)
		assert.NotNil(t, tool)
		assert.Equal(t, "error_tool", tool.Name())

		ctx := helpers.CreateTestToolContext()

		// Should never error with rate 0.0
		for i := 0; i < 10; i++ {
			input := map[string]interface{}{"test": i}
			result, err := tool.Execute(ctx, input)
			assert.NoError(t, err)

			response, ok := result.(map[string]interface{})
			require.True(t, ok)
			assert.Equal(t, "Tool executed successfully", response["message"])
		}
	})

	t.Run("always errors (rate 1.0)", func(t *testing.T) {
		tool := ErrorMockTool(1.0)
		ctx := helpers.CreateTestToolContext()

		// Should always error with rate 1.0
		for i := 0; i < 5; i++ {
			input := map[string]interface{}{"test": i}
			_, err := tool.Execute(ctx, input)
			assert.Error(t, err)
		}
	})

	t.Run("sometimes errors (rate 0.5)", func(t *testing.T) {
		tool := ErrorMockTool(0.5)
		ctx := helpers.CreateTestToolContext()

		// With rate 0.5, approximately half should error
		errorCount := 0
		totalRuns := 20

		for i := 0; i < totalRuns; i++ {
			input := map[string]interface{}{"test": i}
			_, err := tool.Execute(ctx, input)
			if err != nil {
				errorCount++
			}
		}

		// Should be approximately half, allow some variance
		assert.InDelta(t, float64(totalRuns)/2, float64(errorCount), float64(totalRuns)/4)
	})
}
