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

// File Operation Tool Fixtures

// FileReadMockTool creates a mock file read tool
func FileReadMockTool() *mocks.MockTool {
	tool := mocks.NewMockTool("file_read", "Reads content from a file")

	// Virtual file system
	filesystem := map[string]string{
		"/test/file.txt":   "This is test file content.",
		"/config/app.json": `{"name": "test-app", "version": "1.0.0"}`,
		"/data/sample.csv": "name,age,city\nJohn,30,NYC\nJane,25,LA",
		"/logs/app.log":    "[INFO] Application started\n[DEBUG] Processing request",
	}

	tool.OnExecute = func(ctx *domain.ToolContext, input map[string]interface{}) (map[string]interface{}, error) {
		path, ok := input["path"].(string)
		if !ok {
			return nil, errors.New("path parameter required")
		}

		content, exists := filesystem[path]
		if !exists {
			return nil, fmt.Errorf("file not found: %s", path)
		}

		return map[string]interface{}{
			"path":    path,
			"content": content,
			"size":    len(content),
			"type":    "file",
		}, nil
	}

	return tool
}

// FileWriteMockTool creates a mock file write tool
func FileWriteMockTool() *mocks.MockTool {
	tool := mocks.NewMockTool("file_write", "Writes content to a file")

	// Shared filesystem (in real use, this would be persistent)
	filesystem := make(map[string]string)

	tool.OnExecute = func(ctx *domain.ToolContext, input map[string]interface{}) (map[string]interface{}, error) {
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
			"path":   path,
			"size":   len(content),
			"status": "written",
			"type":   "file",
		}, nil
	}

	return tool
}

// FileListMockTool creates a mock file listing tool
func FileListMockTool() *mocks.MockTool {
	tool := mocks.NewMockTool("file_list", "Lists files in a directory")

	// Mock directory structure
	directories := map[string][]map[string]interface{}{
		"/": {
			{"name": "etc", "type": "directory", "size": 4096},
			{"name": "home", "type": "directory", "size": 4096},
			{"name": "tmp", "type": "directory", "size": 4096},
		},
		"/etc": {
			{"name": "config.txt", "type": "file", "size": 156},
			{"name": "hosts", "type": "file", "size": 2048},
		},
		"/home": {
			{"name": "user", "type": "directory", "size": 4096},
		},
		"/home/user": {
			{"name": "readme.md", "type": "file", "size": 1024},
			{"name": "documents", "type": "directory", "size": 4096},
		},
		"/tmp": {
			{"name": "data.json", "type": "file", "size": 512},
			{"name": "temp.log", "type": "file", "size": 256},
		},
	}

	tool.OnExecute = func(ctx *domain.ToolContext, input map[string]interface{}) (map[string]interface{}, error) {
		path, ok := input["path"].(string)
		if !ok {
			path = "/"
		}

		files, exists := directories[path]
		if !exists {
			return nil, fmt.Errorf("directory not found: %s", path)
		}

		return map[string]interface{}{
			"path":  path,
			"files": files,
			"count": len(files),
		}, nil
	}

	return tool
}

// FileDeleteMockTool creates a mock file deletion tool
func FileDeleteMockTool() *mocks.MockTool {
	tool := mocks.NewMockTool("file_delete", "Deletes a file")

	// Track deleted files
	deletedFiles := make(map[string]bool)

	tool.OnExecute = func(ctx *domain.ToolContext, input map[string]interface{}) (map[string]interface{}, error) {
		path, ok := input["path"].(string)
		if !ok {
			return nil, errors.New("path parameter required")
		}

		if deletedFiles[path] {
			return nil, fmt.Errorf("file already deleted: %s", path)
		}

		deletedFiles[path] = true

		return map[string]interface{}{
			"path":   path,
			"status": "deleted",
			"type":   "file",
		}, nil
	}

	return tool
}

// FileMoveMockTool creates a mock file move/rename tool
func FileMoveMockTool() *mocks.MockTool {
	tool := mocks.NewMockTool("file_move", "Moves or renames a file")

	// Track file movements
	movements := make([]map[string]string, 0)

	tool.OnExecute = func(ctx *domain.ToolContext, input map[string]interface{}) (map[string]interface{}, error) {
		sourcePath, ok := input["source"].(string)
		if !ok {
			return nil, errors.New("source parameter required")
		}

		destPath, ok := input["destination"].(string)
		if !ok {
			return nil, errors.New("destination parameter required")
		}

		movement := map[string]string{
			"source":      sourcePath,
			"destination": destPath,
		}
		movements = append(movements, movement)

		return map[string]interface{}{
			"source":      sourcePath,
			"destination": destPath,
			"status":      "moved",
			"type":        "file",
		}, nil
	}

	return tool
}

// Web Tool Fixtures

// WebScrapeMockTool creates a mock web scraping tool
func WebScrapeMockTool() *mocks.MockTool {
	tool := mocks.NewMockTool("web_scrape", "Scrapes content from web pages")

	tool.OnExecute = func(ctx *domain.ToolContext, input map[string]interface{}) (map[string]interface{}, error) {
		url, ok := input["url"].(string)
		if !ok {
			return nil, errors.New("url parameter required")
		}

		// Mock scraped content based on domain
		var content, title string
		if matched, _ := regexp.MatchString("(?i).*github.*", url); matched {
			title = "GitHub Repository"
			content = "# go-llms\nUnified Go interface for LLM providers\n\n## Features\n- Multiple provider support\n- Agent tooling\n- Streaming responses"
		} else if matched, _ := regexp.MatchString("(?i).*stackoverflow.*", url); matched {
			title = "Stack Overflow Question"
			content = "Question: How to implement interfaces in Go?\n\nAnswer: In Go, you implement interfaces implicitly..."
		} else {
			title = "Web Page"
			content = "Mock scraped content from " + url
		}

		return map[string]interface{}{
			"url":     url,
			"title":   title,
			"content": content,
			"status":  "scraped",
			"size":    len(content),
		}, nil
	}

	return tool
}

// WebFetchMockTool creates a mock web fetching tool
func WebFetchMockTool() *mocks.MockTool {
	tool := mocks.NewMockTool("web_fetch", "Fetches web page content")

	tool.OnExecute = func(ctx *domain.ToolContext, input map[string]interface{}) (map[string]interface{}, error) {
		url, ok := input["url"].(string)
		if !ok {
			return nil, errors.New("url parameter required")
		}

		// Mock response based on URL pattern
		var statusCode int
		var content string

		if matched, _ := regexp.MatchString("(?i).*api.*", url); matched {
			statusCode = 200
			content = `{"status": "success", "data": {"message": "Mock API response"}}`
		} else if matched, _ := regexp.MatchString("(?i).*error.*", url); matched {
			statusCode = 404
			content = "Not Found"
		} else {
			statusCode = 200
			content = "<html><body><h1>Mock Web Page</h1><p>This is mock content.</p></body></html>"
		}

		return map[string]interface{}{
			"url":         url,
			"status_code": statusCode,
			"content":     content,
			"size":        len(content),
			"headers": map[string]string{
				"Content-Type": "text/html",
				"Server":       "Mock-Server/1.0",
			},
		}, nil
	}

	return tool
}

// HTTPRequestMockTool creates a mock HTTP request tool
func HTTPRequestMockTool() *mocks.MockTool {
	tool := mocks.NewMockTool("http_request", "Makes HTTP requests")

	tool.OnExecute = func(ctx *domain.ToolContext, input map[string]interface{}) (map[string]interface{}, error) {
		url, ok := input["url"].(string)
		if !ok {
			return nil, errors.New("url parameter required")
		}

		method, ok := input["method"].(string)
		if !ok {
			method = "GET"
		}

		// Mock response based on method and URL
		var statusCode int
		var responseBody string

		switch method {
		case "GET":
			statusCode = 200
			responseBody = `{"method": "GET", "url": "` + url + `", "data": "mock response"}`
		case "POST":
			statusCode = 201
			responseBody = `{"method": "POST", "url": "` + url + `", "status": "created"}`
		case "PUT":
			statusCode = 200
			responseBody = `{"method": "PUT", "url": "` + url + `", "status": "updated"}`
		case "DELETE":
			statusCode = 204
			responseBody = ""
		default:
			statusCode = 405
			responseBody = `{"error": "Method not allowed"}`
		}

		result := map[string]interface{}{
			"url":         url,
			"method":      method,
			"status_code": statusCode,
			"body":        responseBody,
			"headers": map[string]string{
				"Content-Type": "application/json",
				"Server":       "Mock-Server/1.0",
			},
		}

		// Include request body if provided
		if body, exists := input["body"]; exists {
			result["request_body"] = body
		}

		return result, nil
	}

	return tool
}

// Data Processing Tool Fixtures

// JSONProcessMockTool creates a mock JSON processing tool
func JSONProcessMockTool() *mocks.MockTool {
	tool := mocks.NewMockTool("json_process", "Processes JSON data")

	tool.OnExecute = func(ctx *domain.ToolContext, input map[string]interface{}) (map[string]interface{}, error) {
		operation, ok := input["operation"].(string)
		if !ok {
			return nil, errors.New("operation parameter required")
		}

		data, ok := input["data"].(string)
		if !ok {
			return nil, errors.New("data parameter required")
		}

		switch operation {
		case "validate":
			var jsonData interface{}
			if err := json.Unmarshal([]byte(data), &jsonData); err != nil {
				return map[string]interface{}{
					"valid":  false,
					"error":  err.Error(),
					"status": "invalid",
				}, nil
			}
			return map[string]interface{}{
				"valid":  true,
				"status": "valid",
			}, nil

		case "format":
			var jsonData interface{}
			if err := json.Unmarshal([]byte(data), &jsonData); err != nil {
				return nil, fmt.Errorf("invalid JSON: %v", err)
			}
			formatted, _ := json.MarshalIndent(jsonData, "", "  ")
			return map[string]interface{}{
				"formatted": string(formatted),
				"status":    "formatted",
			}, nil

		case "minify":
			var jsonData interface{}
			if err := json.Unmarshal([]byte(data), &jsonData); err != nil {
				return nil, fmt.Errorf("invalid JSON: %v", err)
			}
			minified, _ := json.Marshal(jsonData)
			return map[string]interface{}{
				"minified": string(minified),
				"status":   "minified",
			}, nil

		default:
			return nil, fmt.Errorf("unsupported operation: %s", operation)
		}
	}

	return tool
}

// CSVProcessMockTool creates a mock CSV processing tool
func CSVProcessMockTool() *mocks.MockTool {
	tool := mocks.NewMockTool("csv_process", "Processes CSV data")

	tool.OnExecute = func(ctx *domain.ToolContext, input map[string]interface{}) (map[string]interface{}, error) {
		operation, ok := input["operation"].(string)
		if !ok {
			return nil, errors.New("operation parameter required")
		}

		data, ok := input["data"].(string)
		if !ok {
			return nil, errors.New("data parameter required")
		}

		switch operation {
		case "parse":
			// Simple CSV parsing mock
			lines := regexp.MustCompile(`\r?\n`).Split(data, -1)
			var rows [][]string
			for _, line := range lines {
				if line != "" {
					rows = append(rows, regexp.MustCompile(`,`).Split(line, -1))
				}
			}

			return map[string]interface{}{
				"rows":   rows,
				"count":  len(rows),
				"status": "parsed",
			}, nil

		case "validate":
			lines := regexp.MustCompile(`\r?\n`).Split(data, -1)
			var columnCount int
			valid := true

			for i, line := range lines {
				if line == "" {
					continue
				}
				columns := regexp.MustCompile(`,`).Split(line, -1)
				if i == 0 {
					columnCount = len(columns)
				} else if len(columns) != columnCount {
					valid = false
					break
				}
			}

			return map[string]interface{}{
				"valid":   valid,
				"columns": columnCount,
				"status":  "validated",
			}, nil

		default:
			return nil, fmt.Errorf("unsupported operation: %s", operation)
		}
	}

	return tool
}

// TextProcessMockTool creates a mock text processing tool
func TextProcessMockTool() *mocks.MockTool {
	tool := mocks.NewMockTool("text_process", "Processes text data")

	tool.OnExecute = func(ctx *domain.ToolContext, input map[string]interface{}) (map[string]interface{}, error) {
		operation, ok := input["operation"].(string)
		if !ok {
			return nil, errors.New("operation parameter required")
		}

		text, ok := input["text"].(string)
		if !ok {
			return nil, errors.New("text parameter required")
		}

		switch operation {
		case "word_count":
			words := regexp.MustCompile(`\s+`).Split(text, -1)
			return map[string]interface{}{
				"words":      len(words),
				"characters": len(text),
				"lines":      len(regexp.MustCompile(`\r?\n`).Split(text, -1)),
				"status":     "counted",
			}, nil

		case "uppercase":
			return map[string]interface{}{
				"result": regexp.MustCompile(`\b\w`).ReplaceAllStringFunc(text, func(s string) string {
					return regexp.MustCompile(`\w`).ReplaceAllStringFunc(s, func(c string) string {
						if c >= "a" && c <= "z" {
							return string(rune(c[0] - 'a' + 'A'))
						}
						return c
					})
				}),
				"status": "converted",
			}, nil

		case "extract_emails":
			emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
			emails := emailRegex.FindAllString(text, -1)
			return map[string]interface{}{
				"emails": emails,
				"count":  len(emails),
				"status": "extracted",
			}, nil

		default:
			return nil, fmt.Errorf("unsupported operation: %s", operation)
		}
	}

	return tool
}
