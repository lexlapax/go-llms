package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	agentDomain "github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/tools"
	"github.com/lexlapax/go-llms/pkg/agent/workflow"
	llmDomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// CustomHook implements a simple custom hook for tracking all agent events
type CustomHook struct {
	name      string
	startTime time.Time
	events    []string
}

// NewCustomHook creates a new CustomHook with the given name
func NewCustomHook(name string) *CustomHook {
	return &CustomHook{
		name:      name,
		startTime: time.Now(),
		events:    make([]string, 0),
	}
}

// BeforeGenerate is called before generating a response
func (h *CustomHook) BeforeGenerate(ctx context.Context, messages []llmDomain.Message) {
	h.events = append(h.events, fmt.Sprintf("[%s] BeforeGenerate: %d messages", h.name, len(messages)))
}

// AfterGenerate is called after generating a response
func (h *CustomHook) AfterGenerate(ctx context.Context, response llmDomain.Response, err error) {
	if err != nil {
		h.events = append(h.events, fmt.Sprintf("[%s] AfterGenerate Error: %v", h.name, err))
	} else {
		h.events = append(h.events, fmt.Sprintf("[%s] AfterGenerate: Response received", h.name))
	}
}

// BeforeToolCall is called before executing a tool
func (h *CustomHook) BeforeToolCall(ctx context.Context, toolName string, params map[string]interface{}) {
	paramJSON, _ := json.Marshal(params)
	h.events = append(h.events, fmt.Sprintf("[%s] BeforeToolCall: %s with params: %s", h.name, toolName, paramJSON))
}

// AfterToolCall is called after executing a tool
func (h *CustomHook) AfterToolCall(ctx context.Context, toolName string, result interface{}, err error) {
	if err != nil {
		h.events = append(h.events, fmt.Sprintf("[%s] AfterToolCall: %s error: %v", h.name, toolName, err))
	} else {
		resultJSON, _ := json.Marshal(result)
		h.events = append(h.events, fmt.Sprintf("[%s] AfterToolCall: %s result: %s", h.name, toolName, resultJSON))
	}
}

// GetEvents returns all collected events
func (h *CustomHook) GetEvents() []string {
	return h.events
}

// PrintSummary prints a summary of all events
func (h *CustomHook) PrintSummary() {
	fmt.Printf("\nCustom Hook (%s) Summary:\n", h.name)
	fmt.Printf("Total events: %d\n", len(h.events))
	fmt.Printf("Total duration: %v\n", time.Since(h.startTime))
}

func main() {
	// Set up logging
	console := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	// Create a log file
	logFile, err := os.Create("agent.log")
	if err != nil {
		fmt.Printf("Error creating log file: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()

	// Create file logger with debug level for full details
	fileHandler := slog.NewTextHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	// Create a multi-writer logger
	multiLogger := slog.New(slog.NewJSONHandler(io.MultiWriter(os.Stdout, logFile), &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	_ = multiLogger // Using multiLogger for more complex scenarios

	// Check for required API keys
	var providerName, modelName string
	var llmProvider llmDomain.Provider

	// Try different providers in order of preference
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		providerName = "OpenAI"
		modelName = "gpt-4o"
		llmProvider = provider.NewOpenAIProvider(apiKey, modelName)
	} else if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		providerName = "Anthropic"
		modelName = "claude-3-5-sonnet-latest"
		llmProvider = provider.NewAnthropicProvider(apiKey, modelName)
	} else {
		providerName = "Mock"
		modelName = "mock-model"
		llmProvider = provider.NewMockProvider()
		fmt.Println("No API keys found for OpenAI or Anthropic. Using mock provider.")
	}

	fmt.Printf("Using %s provider with model: %s\n", providerName, modelName)

	// Create advanced optimized agent (CachedAgent with all optimizations enabled)
	cachedAgent := workflow.NewCachedAgent(llmProvider)
	
	// Use the Agent interface for compatibility
	var agent agentDomain.Agent = cachedAgent

	// Configure the agent
	agent.SetSystemPrompt(`You are a helpful assistant with access to various tools.
Your goal is to provide accurate, helpful responses to user queries.
When a user asks you a question, think about whether you need to use a tool to answer it.
Only use tools when necessary - if you know the answer, you can just respond directly.
Always explain your reasoning and the steps you're taking.
When using tools, provide the exact information requested and format your response clearly.
For calculations, use the calculator tool. For current date or time, use the get_current_date tool.
For web content, use the web_fetch tool. For file operations, use the read_file and write_file tools.`)

	// Set the model
	agent.WithModel(modelName)

	// Add hooks for monitoring and logging
	// 1. Add a basic logging hook (minimal information)
	basicLogger := slog.New(console)
	agent.WithHook(workflow.NewLoggingHook(basicLogger, workflow.LogLevelBasic))

	// 2. Add a detailed logging hook (saved to file)
	debugLogger := slog.New(fileHandler)
	agent.WithHook(workflow.NewLoggingHook(debugLogger, workflow.LogLevelDetailed))

	// 3. Add a metrics hook
	metricsHook := workflow.NewMetricsHook()
	agent.WithHook(metricsHook)

	// 4. Add our custom hook
	customHook := NewCustomHook("AgentMonitor")
	agent.WithHook(customHook)

	// Add all available tools to the agent
	addTools(agent)

	// Create a metrics-enabled context
	ctx := workflow.WithMetrics(context.Background())

	fmt.Println("\n=== Example 1: Basic Tool Usage ===")
	fmt.Println("Running the agent with a basic query that needs tools...")

	// Basic example - needs calculator and date tools
	result, err := agent.Run(ctx, "What is the current year? Then calculate 25 * 42.")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("\nResult:")
		fmt.Println(result)
	}

	// Display metrics after the first run
	displayMetrics(metricsHook)
	
	// Display cache metrics
	displayCacheMetrics(cachedAgent)
	
	metricsHook.Reset() // Reset for next example

	// Example 2: Multiple tools in one response
	fmt.Println("\n=== Example 2: Multiple Tools in One Response ===")
	fmt.Println("Running the agent with a query that should trigger multiple tool calls...")
	
	// Complex example - should trigger multiple tool calls
	result, err = agent.Run(ctx, "What's the current date, what's 15 * 7, and can you fetch the title from https://example.com?")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("\nResult:")
		fmt.Println(result)
	}

	// Display metrics after the second run
	displayMetrics(metricsHook)
	
	// Display cache metrics
	displayCacheMetrics(cachedAgent)
	
	metricsHook.Reset()

	// Example 3: Parallel tool execution example
	fmt.Println("\n=== Example 3: Parallel Tool Execution ===")
	
	// Create multiple temporary files for testing parallel operations
	tempDir := os.TempDir()
	tempFiles := []string{
		filepath.Join(tempDir, "go-llms-agent-test1.txt"),
		filepath.Join(tempDir, "go-llms-agent-test2.txt"),
		filepath.Join(tempDir, "go-llms-agent-test3.txt"),
	}
	
	// Create the files with different content
	os.WriteFile(tempFiles[0], []byte("This is file 1.\nIt contains information about tools."), 0644)
	os.WriteFile(tempFiles[1], []byte("This is file 2.\nIt contains information about agents."), 0644)
	os.WriteFile(tempFiles[2], []byte("This is file 3.\nIt contains information about providers."), 0644)
	
	// Clean up files when done
	defer func() {
		for _, file := range tempFiles {
			os.Remove(file)
		}
	}()
	
	// Prepare a prompt that requires multiple operations in parallel
	prompt := fmt.Sprintf("Please perform these operations in parallel:\n"+
		"1. Read file %s and extract the key topic.\n"+
		"2. Read file %s and extract the key topic.\n"+
		"3. Read file %s and extract the key topic.\n"+
		"4. Calculate 25 * 42.\n"+
		"5. Get the current date.\n"+
		"Summarize all results together.", tempFiles[0], tempFiles[1], tempFiles[2])
	
	fmt.Println("Running parallel tool operations...")
	
	// Start timing for parallel execution
	startTime := time.Now()
	result, err = agent.Run(ctx, prompt)
	duration := time.Since(startTime)
	
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("\nResult:")
		fmt.Println(result)
		fmt.Printf("\nTotal execution time: %v\n", duration)
	}
	
	// Display metrics
	displayMetrics(metricsHook)
	
	// Display cache metrics
	displayCacheMetrics(cachedAgent)
	
	metricsHook.Reset()
	
	// Example 4: Demonstrate caching with repeated queries
	fmt.Println("\n=== Example 4: Caching with Repeated Queries ===")
	
	// Ensure caching is enabled
	cachedAgent.EnableCaching(true)
	fmt.Println("Response caching is enabled")
	
	// Define a query that will be repeated
	repeatedQuery := "What's 25 * 42 and what's the current year?"
	
	// First run - should hit the provider
	fmt.Println("\nFirst run of repeated query (should miss cache):")
	startTime1 := time.Now()
	result1, err := agent.Run(ctx, repeatedQuery)
	duration1 := time.Since(startTime1)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("\nResult (first run):")
		fmt.Println(result1)
		fmt.Printf("Duration: %v\n", duration1)
	}
	
	// Second run - should hit the cache
	fmt.Println("\nSecond run of identical query (should hit cache):")
	startTime2 := time.Now()
	result2, err := agent.Run(ctx, repeatedQuery)
	duration2 := time.Since(startTime2)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("\nResult (second run):")
		fmt.Println(result2)
		fmt.Printf("Duration: %v\n", duration2)
		
		// Display cache hit metrics
		if cachedAgent, ok := agent.(*workflow.CachedAgent); ok {
			stats := cachedAgent.GetCacheStats()
			fmt.Printf("\nCache hits: %d, Misses: %d\n", stats["hits"], stats["misses"])
			fmt.Printf("Time saved: approximately %v\n", duration1-duration2)
		}
	}
	
	// Example 5: Complex analysis with schema
	fmt.Println("\n=== Example 5: Structured Output with Schema ===")
	
	// Define a schema for analysis results
	analysisSchema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"analysis": {
				Type:        "string",
				Description: "The main analysis text",
			},
			"tools_used": {
				Type: "array",
				Items: &schemaDomain.Property{
					Type: "string",
				},
				Description: "List of tool names that were used",
			},
			"calculations": {
				Type: "array",
				Items: &schemaDomain.Property{
					Type: "object",
					Properties: map[string]schemaDomain.Property{
						"expression": {Type: "string"},
						"result":     {Type: "number"},
					},
				},
				Description: "Any calculations performed",
			},
			"current_date": {
				Type:        "string",
				Description: "The current date when the analysis was performed",
			},
		},
		Required: []string{"analysis", "tools_used"},
	}
	
	// Run a complex analysis with schema
	result, err = agent.RunWithSchema(
		ctx,
		"Please analyze the current date, calculate how many days are in this month, and determine how many more days until the end of the year.",
		analysisSchema,
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		// Format structured result
		resultJSON, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println("\nStructured Result:")
		fmt.Println(string(resultJSON))
	}
	
	// Display metrics and reset
	displayMetrics(metricsHook)
	
	// Display cache metrics (for CachedAgent)
	if cachedAgent, ok := agent.(*workflow.CachedAgent); ok {
		displayCacheMetrics(cachedAgent)
	}
	
	customHook.PrintSummary()
}

// displayMetrics prints the metrics collected by the metrics hook
func displayMetrics(hook *workflow.MetricsHook) {
	metrics := hook.GetMetrics()
	
	fmt.Println("\nAgent Metrics:")
	fmt.Println("===================")
	fmt.Printf("Total requests:       %d\n", metrics.Requests)
	fmt.Printf("Total tool calls:     %d\n", metrics.ToolCalls)
	fmt.Printf("Avg tool calls/req:   %.2f\n", float64(metrics.ToolCalls)/float64(metrics.Requests))
	fmt.Printf("Errors:               %d\n", metrics.ErrorCount)
	fmt.Printf("Total tokens:         %d\n", metrics.TotalTokens)
	fmt.Printf("Average gen time:     %.2f ms\n", metrics.AverageGenTimeMs)
	
	if len(metrics.ToolStats) > 0 {
		fmt.Println("\nTool Statistics:")
		fmt.Println("--------------------")
		fmt.Printf("%-20s | %-8s | %-12s | %-12s | %-12s\n", "Tool", "Calls", "Avg Time (ms)", "Fastest (ms)", "Slowest (ms)")
		fmt.Println(strings.Repeat("-", 78))
		
		for tool, stats := range metrics.ToolStats {
			fmt.Printf("%-20s | %-8d | %-12.2f | %-12.2f | %-12.2f\n",
				tool, stats.Calls, stats.AverageTimeMs, stats.FastestCallMs, stats.SlowestCallMs)
		}
	}
}

// displayCacheMetrics displays statistics from the response cache
func displayCacheMetrics(agent *workflow.CachedAgent) {
	stats := agent.GetCacheStats()
	
	fmt.Println("\nCache Metrics:")
	fmt.Println("===================")
	fmt.Printf("Cache hits:           %d\n", stats["hits"])
	fmt.Printf("Cache misses:         %d\n", stats["misses"])
	
	// Calculate and display hit ratio
	hitRatio, _ := stats["hit_ratio"].(float64)
	fmt.Printf("Hit ratio:            %.2f%%\n", hitRatio*100)
	
	fmt.Printf("Stored responses:     %d\n", stats["stored_responses"])
	fmt.Printf("Evicted responses:    %d\n", stats["evicted_responses"])
	fmt.Printf("Current cache size:   %d\n", stats["cache_size"])
	fmt.Printf("Cache capacity:       %d\n", stats["config"].(map[string]interface{})["capacity"])
	
	// Display average response time saved
	if avgSaved, ok := stats["avg_response_saving_ms"].(int64); ok && avgSaved > 0 {
		fmt.Printf("Avg time saved:       %d ms/request\n", avgSaved)
	}
}

// addTools adds all available tools to the agent
func addTools(agent agentDomain.Agent) {
	// 1. Date tool - provides current date information
	agent.AddTool(tools.NewTool(
		"get_current_date",
		"Get the current date and time information",
		func() map[string]string {
			now := time.Now()
			return map[string]string{
				"date":          now.Format("2006-01-02"),
				"time":          now.Format("15:04:05"),
				"year":          strconv.Itoa(now.Year()),
				"month":         now.Month().String(),
				"day":           strconv.Itoa(now.Day()),
				"weekday":       now.Weekday().String(),
				"timezone":      now.Location().String(),
				"unix_epoch":    strconv.FormatInt(now.Unix(), 10),
				"days_in_month": strconv.Itoa(daysInMonth(now.Year(), now.Month())),
				"days_in_year":  strconv.Itoa(daysInYear(now.Year())),
				"days_left_in_year": strconv.Itoa(daysLeftInYear(now)),
			}
		},
		&schemaDomain.Schema{
			Type:        "object",
			Description: "Returns the current date and time information",
		},
	))

	// 2. Enhanced calculator tool with more operations
	agent.AddTool(tools.NewTool(
		"calculator",
		"Perform mathematical calculations",
		func(params struct {
			Expression string `json:"expression"`
		}) (map[string]interface{}, error) {
			// Handle basic operations
			result, err := evaluateExpression(params.Expression)
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   err.Error(),
				}, nil
			}
			
			return map[string]interface{}{
				"success":    true,
				"expression": params.Expression,
				"result":     result,
			}, nil
		},
		&schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"expression": {
					Type:        "string",
					Description: "The mathematical expression to evaluate (supports +, -, *, /, sqrt, factorial)",
				},
			},
			Required: []string{"expression"},
		},
	))
	
	// 3. Web fetch tool
	agent.AddTool(tools.NewTool(
		"web_fetch",
		"Fetch content from a URL",
		func(params struct {
			URL string `json:"url"`
		}) (map[string]interface{}, error) {
			// Create a client with a timeout
			client := &http.Client{
				Timeout: 30 * time.Second,
			}
			
			// Basic validation of URL
			if !strings.HasPrefix(params.URL, "http://") && !strings.HasPrefix(params.URL, "https://") {
				return nil, fmt.Errorf("invalid URL: must begin with http:// or https://")
			}

			// Fetch the URL
			resp, err := client.Get(params.URL)
			if err != nil {
				return nil, fmt.Errorf("error fetching URL: %v", err)
			}
			defer resp.Body.Close()

			// Check status code
			if resp.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
			}

			// Read response body (limited to 100KB to prevent huge responses)
			maxBytes := int64(100 * 1024) // 100KB
			body, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes))
			if err != nil {
				return nil, fmt.Errorf("error reading response: %v", err)
			}

			// Extract title if it's HTML
			title := ""
			contentType := resp.Header.Get("Content-Type")
			if strings.Contains(contentType, "text/html") {
				// Very simple title extraction, not a full HTML parser
				bodyStr := string(body)
				titleStart := strings.Index(strings.ToLower(bodyStr), "<title>")
				if titleStart >= 0 {
					titleStart += 7 // length of <title>
					titleEnd := strings.Index(strings.ToLower(bodyStr[titleStart:]), "</title>")
					if titleEnd >= 0 {
						title = strings.TrimSpace(bodyStr[titleStart : titleStart+titleEnd])
					}
				}
			}

			return map[string]interface{}{
				"content":      string(body),
				"status_code":  resp.StatusCode,
				"content_type": contentType,
				"url":          params.URL,
				"title":        title,
				"headers":      resp.Header,
				"length":       len(body),
			}, nil
		},
		&schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"url": {
					Type:        "string",
					Description: "The URL to fetch content from",
					Format:      "uri",
				},
			},
			Required: []string{"url"},
		},
	))
	
	// 4. Read file tool
	agent.AddTool(tools.NewTool(
		"read_file",
		"Read content from a file",
		func(params struct {
			Path string `json:"path"`
		}) (map[string]interface{}, error) {
			// Security check: Prevent path traversal
			cleanPath := filepath.Clean(params.Path)
			if strings.Contains(cleanPath, "..") {
				return nil, fmt.Errorf("path traversal not allowed")
			}

			// Make path absolute if not already
			if !filepath.IsAbs(cleanPath) {
				workDir, err := os.Getwd()
				if err != nil {
					return nil, fmt.Errorf("error getting working directory: %v", err)
				}
				cleanPath = filepath.Join(workDir, cleanPath)
			}

			// Check if file exists
			fileInfo, err := os.Stat(cleanPath)
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("file not found: %s", cleanPath)
			}
			if err != nil {
				return nil, fmt.Errorf("error accessing file: %v", err)
			}
			
			// Don't read directories
			if fileInfo.IsDir() {
				return nil, fmt.Errorf("cannot read directory, specify a file path")
			}

			// Read file content
			content, err := os.ReadFile(cleanPath)
			if err != nil {
				return nil, fmt.Errorf("error reading file: %v", err)
			}

			return map[string]interface{}{
				"content":     string(content),
				"path":        cleanPath,
				"size":        fileInfo.Size(),
				"modified":    fileInfo.ModTime().Format(time.RFC3339),
				"is_dir":      fileInfo.IsDir(),
				"permissions": fileInfo.Mode().String(),
			}, nil
		},
		&schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"path": {
					Type:        "string",
					Description: "Path to the file to read",
				},
			},
			Required: []string{"path"},
		},
	))
	
	// 5. Write file tool
	agent.AddTool(tools.NewTool(
		"write_file",
		"Write content to a file",
		func(params struct {
			Path    string `json:"path"`
			Content string `json:"content"`
		}) (map[string]interface{}, error) {
			// Security check: Prevent path traversal
			cleanPath := filepath.Clean(params.Path)
			if strings.Contains(cleanPath, "..") {
				return nil, fmt.Errorf("path traversal not allowed")
			}

			// Make path absolute if not already
			if !filepath.IsAbs(cleanPath) {
				workDir, err := os.Getwd()
				if err != nil {
					return nil, fmt.Errorf("error getting working directory: %v", err)
				}
				cleanPath = filepath.Join(workDir, cleanPath)
			}

			// Ensure directory exists
			dir := filepath.Dir(cleanPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("error creating directory: %v", err)
			}

			// Write to the file
			err := os.WriteFile(cleanPath, []byte(params.Content), 0644)
			if err != nil {
				return nil, fmt.Errorf("error writing file: %v", err)
			}

			// Get file info after writing
			fileInfo, err := os.Stat(cleanPath)
			if err != nil {
				return map[string]interface{}{
					"success":     true,
					"path":        cleanPath,
					"bytes":       len(params.Content),
					"error_info":  "File written but could not get file info: " + err.Error(),
				}, nil
			}

			return map[string]interface{}{
				"success":     true,
				"path":        cleanPath,
				"bytes":       len(params.Content),
				"size":        fileInfo.Size(),
				"modified":    fileInfo.ModTime().Format(time.RFC3339),
				"permissions": fileInfo.Mode().String(),
			}, nil
		},
		&schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"path": {
					Type:        "string",
					Description: "Path to the file to write",
				},
				"content": {
					Type:        "string",
					Description: "Content to write to the file",
				},
			},
			Required: []string{"path", "content"},
		},
	))
	
	// 6. Execute command tool (very limited and secure)
	agent.AddTool(tools.NewTool(
		"execute_command",
		"Execute a safe system command. Limited to 'echo', 'pwd', and 'ls' commands only.",
		func(params struct {
			Command string `json:"command"`
		}) (map[string]interface{}, error) {
			// Extremely restricted command execution
			// Only allow specific commands
			cmd := strings.TrimSpace(params.Command)
			
			// Only allow certain commands (extreme caution for a real app!)
			allowedCommands := map[string]bool{
				"echo": true,
				"ls":   true,
				"pwd":  true,
			}
			
			// Extract the base command (everything before first space)
			baseCmd := cmd
			if idx := strings.Index(cmd, " "); idx > 0 {
				baseCmd = cmd[:idx]
			}
			
			// Check if command is allowed
			if !allowedCommands[baseCmd] {
				return nil, fmt.Errorf("command not allowed for security reasons: %s", baseCmd)
			}
			
			// Set a short timeout
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			
			// Create and execute the command
			var stdout, stderr bytes.Buffer
			execCmd := exec.CommandContext(ctx, "sh", "-c", cmd)
			execCmd.Stdout = &stdout
			execCmd.Stderr = &stderr
			
			err := execCmd.Run()
			
			result := map[string]interface{}{
				"command": cmd,
				"stdout":  stdout.String(),
				"stderr":  stderr.String(),
			}
			
			if ctx.Err() == context.DeadlineExceeded {
				return nil, fmt.Errorf("command timed out after 5 seconds")
			}
			
			if err != nil {
				result["success"] = false
				result["error"] = err.Error()
			} else {
				result["success"] = true
				result["exit_code"] = 0
			}
			
			return result, nil
		},
		&schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"command": {
					Type:        "string",
					Description: "The command to execute (only 'echo', 'pwd', and 'ls' allowed)",
				},
			},
			Required: []string{"command"},
		},
	))
}

// Helper functions for the calculator tool
func evaluateExpression(expr string) (float64, error) {
	expr = strings.TrimSpace(expr)
	
	// Check for sqrt operation
	if strings.HasPrefix(expr, "sqrt(") && strings.HasSuffix(expr, ")") {
		numStr := expr[5 : len(expr)-1]
		num, err := strconv.ParseFloat(strings.TrimSpace(numStr), 64)
		if err != nil {
			return 0, fmt.Errorf("invalid number for sqrt: %s", numStr)
		}
		if num < 0 {
			return 0, fmt.Errorf("cannot take square root of negative number")
		}
		return math.Sqrt(num), nil
	}
	
	// Check for factorial operation
	if strings.HasPrefix(expr, "factorial(") && strings.HasSuffix(expr, ")") {
		numStr := expr[10 : len(expr)-1]
		num, err := strconv.ParseFloat(strings.TrimSpace(numStr), 64)
		if err != nil {
			return 0, fmt.Errorf("invalid number for factorial: %s", numStr)
		}
		
		// Convert to int and check if it's a positive integer
		intNum := int(num)
		if float64(intNum) != num || intNum < 0 {
			return 0, fmt.Errorf("factorial only works on non-negative integers")
		}
		
		result := 1
		for i := 2; i <= intNum; i++ {
			result *= i
		}
		return float64(result), nil
	}
	
	// Handle basic arithmetic
	parts := strings.Split(expr, "*")
	if len(parts) == 2 {
		// Multiplication
		a, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		if err != nil {
			return 0, fmt.Errorf("invalid number: %s", parts[0])
		}
		b, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			return 0, fmt.Errorf("invalid number: %s", parts[1])
		}
		return a * b, nil
	}

	parts = strings.Split(expr, "/")
	if len(parts) == 2 {
		// Division
		a, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		if err != nil {
			return 0, fmt.Errorf("invalid number: %s", parts[0])
		}
		b, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			return 0, fmt.Errorf("invalid number: %s", parts[1])
		}
		if b == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		return a / b, nil
	}

	parts = strings.Split(expr, "+")
	if len(parts) == 2 {
		// Addition
		a, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		if err != nil {
			return 0, fmt.Errorf("invalid number: %s", parts[0])
		}
		b, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			return 0, fmt.Errorf("invalid number: %s", parts[1])
		}
		return a + b, nil
	}

	parts = strings.Split(expr, "-")
	if len(parts) == 2 {
		// Subtraction
		a, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		if err != nil {
			return 0, fmt.Errorf("invalid number: %s", parts[0])
		}
		b, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			return 0, fmt.Errorf("invalid number: %s", parts[1])
		}
		return a - b, nil
	}

	// If we got here, we couldn't evaluate the expression
	return 0, fmt.Errorf("unsupported or invalid expression: %s", expr)
}

// Helper function for date calculations
func daysInMonth(year int, month time.Month) int {
	// This returns the first day of the next month, then subtracts 1 day and gets the day
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// Helper function to calculate days in a year (handling leap years)
func daysInYear(year int) int {
	if year%400 == 0 || (year%4 == 0 && year%100 != 0) {
		return 366 // Leap year
	}
	return 365 // Non-leap year
}

// Helper function to calculate days left in the year
func daysLeftInYear(now time.Time) int {
	year := now.Year()
	endOfYear := time.Date(year, 12, 31, 23, 59, 59, 0, now.Location())
	return int(endOfYear.Sub(now).Hours() / 24) + 1
}