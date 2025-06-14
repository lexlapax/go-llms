// ABOUTME: Pre-configured mock tools for common testing scenarios
// ABOUTME: Provides ready-to-use tool fixtures with typical behavior patterns

package fixtures

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/testutils/mocks"
)

// CalculatorMockTool creates a mock tool that handles arithmetic operations
func CalculatorMockTool() *mocks.MockTool {
	tool := mocks.NewMockTool("calculator", "Performs arithmetic calculations")

	// Set up the execution logic
	tool.OnExecute = func(ctx *domain.ToolContext, input map[string]interface{}) (map[string]interface{}, error) {
		operation, ok := input["operation"].(string)
		if !ok {
			return map[string]interface{}{
				"error":   "Missing operation",
				"message": "Please specify an operation (add, subtract, multiply, divide)",
			}, nil
		}

		switch operation {
		case "add", "+":
			a, err := getNumber(input, "a")
			if err != nil {
				return nil, err
			}
			b, err := getNumber(input, "b")
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{
				"operation": "addition",
				"a":         a,
				"b":         b,
				"result":    a + b,
			}, nil

		case "subtract", "-":
			a, err := getNumber(input, "a")
			if err != nil {
				return nil, err
			}
			b, err := getNumber(input, "b")
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{
				"operation": "subtraction",
				"a":         a,
				"b":         b,
				"result":    a - b,
			}, nil

		case "multiply", "*":
			a, err := getNumber(input, "a")
			if err != nil {
				return nil, err
			}
			b, err := getNumber(input, "b")
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{
				"operation": "multiplication",
				"a":         a,
				"b":         b,
				"result":    a * b,
			}, nil

		case "divide", "/":
			a, err := getNumber(input, "a")
			if err != nil {
				return nil, err
			}
			b, err := getNumber(input, "b")
			if err != nil {
				return nil, err
			}
			if b == 0 {
				return nil, errors.New("division by zero")
			}
			return map[string]interface{}{
				"operation": "division",
				"a":         a,
				"b":         b,
				"result":    a / b,
			}, nil

		default:
			return map[string]interface{}{
				"error":   "Unsupported operation",
				"input":   input,
				"message": "Please specify a valid arithmetic operation (add, subtract, multiply, divide)",
			}, nil
		}
	}

	return tool
}

// WebSearchMockTool creates a mock tool that simulates web search results
func WebSearchMockTool() *mocks.MockTool {
	tool := mocks.NewMockTool("web_search", "Searches the web for information")

	tool.OnExecute = func(ctx *domain.ToolContext, input map[string]interface{}) (map[string]interface{}, error) {
		query := getQuery(input)

		// Check for weather queries
		if matched, _ := regexp.MatchString("(?i).*weather.*", query); matched {
			return map[string]interface{}{
				"query": query,
				"results": []map[string]interface{}{
					{
						"title":       "Current Weather Forecast",
						"url":         "https://weather.example.com",
						"description": "Today's weather is sunny with a high of 75°F and a low of 60°F. Clear skies expected throughout the day.",
						"timestamp":   "2024-01-01T12:00:00Z",
					},
					{
						"title":       "Weather.com - Local Weather",
						"url":         "https://weather.com/local",
						"description": "Get accurate weather forecasts for your location. Hourly and 10-day forecasts available.",
						"timestamp":   "2024-01-01T12:00:00Z",
					},
				},
				"total_results": 2,
			}, nil
		}

		// Check for programming queries
		if matched, _ := regexp.MatchString("(?i).*programming.*", query); matched {
			return map[string]interface{}{
				"query": query,
				"results": []map[string]interface{}{
					{
						"title":       "Learn Programming - Codecademy",
						"url":         "https://codecademy.com/learn",
						"description": "Interactive programming courses for beginners and experts. Learn Python, JavaScript, Java, and more.",
						"timestamp":   "2024-01-01T12:00:00Z",
					},
					{
						"title":       "Programming Best Practices",
						"url":         "https://blog.example.com/programming-best-practices",
						"description": "Essential programming principles every developer should know. Clean code, testing, and design patterns.",
						"timestamp":   "2024-01-01T12:00:00Z",
					},
				},
				"total_results": 2,
			}, nil
		}

		// Default response
		return map[string]interface{}{
			"query": query,
			"results": []map[string]interface{}{
				{
					"title":       fmt.Sprintf("Search Results for: %s", query),
					"url":         "https://example.com/search?q=" + query,
					"description": "Mock search result for testing purposes. This is a generic response.",
					"timestamp":   "2024-01-01T12:00:00Z",
				},
			},
			"total_results": 1,
		}, nil
	}

	return tool
}

// FileMockTool creates a mock tool with a virtual filesystem
func FileMockTool() *mocks.MockTool {
	tool := mocks.NewMockTool("file_manager", "Manages files in a virtual filesystem")

	// Virtual filesystem storage
	filesystem := make(map[string]string)

	// Initialize with some default files
	filesystem["/etc/config.txt"] = "# Configuration file\nkey=value\n"
	filesystem["/home/user/readme.md"] = "# Welcome\nThis is a test file.\n"
	filesystem["/tmp/data.json"] = `{"name": "test", "value": 42}`

	tool.OnExecute = func(ctx *domain.ToolContext, input map[string]interface{}) (map[string]interface{}, error) {
		operation, ok := input["operation"].(string)
		if !ok {
			return map[string]interface{}{
				"error":   "Missing operation",
				"message": "Please specify an operation (read, write, list)",
			}, nil
		}

		switch {
		case regexp.MustCompile("(?i)read|cat").MatchString(operation):
			path, ok := input["path"].(string)
			if !ok {
				return nil, errors.New("path parameter required")
			}

			content, exists := filesystem[path]
			if !exists {
				return nil, fmt.Errorf("file not found: %s", path)
			}

			return map[string]interface{}{
				"operation": "read",
				"path":      path,
				"content":   content,
				"size":      len(content),
			}, nil

		case regexp.MustCompile("(?i)write|save").MatchString(operation):
			path, ok := input["path"].(string)
			if !ok {
				return nil, errors.New("path parameter required")
			}

			content, ok := input["content"].(string)
			if !ok {
				return nil, errors.New("content parameter required")
			}

			filesystem[path] = content

			return map[string]interface{}{
				"operation": "write",
				"path":      path,
				"size":      len(content),
				"status":    "success",
			}, nil

		case regexp.MustCompile("(?i)list|ls").MatchString(operation):
			var files []map[string]interface{}
			for path, content := range filesystem {
				files = append(files, map[string]interface{}{
					"path": path,
					"size": len(content),
					"type": "file",
				})
			}

			return map[string]interface{}{
				"operation": "list",
				"files":     files,
				"count":     len(files),
			}, nil

		default:
			return map[string]interface{}{
				"error":   "Unsupported file operation",
				"input":   input,
				"message": "Please specify a valid file operation (read, write, list)",
			}, nil
		}
	}

	return tool
}

// ErrorMockTool creates a mock tool that randomly fails based on error rate
func ErrorMockTool(errorRate float64) *mocks.MockTool {
	tool := mocks.NewMockTool("error_tool", "A tool that simulates errors for testing")

	// Set up error injection
	tool.ErrorRate = errorRate

	// Set default response for when it doesn't error
	tool.DefaultOutput = map[string]interface{}{
		"message": "Tool executed successfully",
	}

	return tool
}

// Helper functions

func getNumber(input map[string]interface{}, key string) (float64, error) {
	val, ok := input[key]
	if !ok {
		return 0, fmt.Errorf("parameter '%s' is required", key)
	}

	switch v := val.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, fmt.Errorf("parameter '%s' must be a number", key)
		}
		return parsed, nil
	default:
		return 0, fmt.Errorf("parameter '%s' must be a number", key)
	}
}

func getQuery(input map[string]interface{}) string {
	if query, ok := input["query"].(string); ok {
		return query
	}
	if q, ok := input["q"].(string); ok {
		return q
	}
	if search, ok := input["search"].(string); ok {
		return search
	}

	// Try to extract from JSON string
	if jsonStr, ok := input["input"].(string); ok {
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &parsed); err == nil {
			if query, ok := parsed["query"].(string); ok {
				return query
			}
		}
	}

	// Fall back to converting input to string
	return fmt.Sprintf("%v", input)
}
