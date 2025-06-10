// ABOUTME: Demonstrates LLM agents using all categories of built-in tools
// ABOUTME: Shows proper system prompts and tool integration patterns for each tool category

package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	"github.com/lexlapax/go-llms/pkg/util/llmutil"

	// Import all tool categories to register them
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/data"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/datetime"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/feed"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/file"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/system"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: agent-llm-builtin-tools <scenario> [prompt]")
		fmt.Println("\nAvailable scenarios:")
		fmt.Println("  research      - Research assistant with web tools")
		fmt.Println("  files         - File manager with file tools")
		fmt.Println("  system        - System admin with system tools")
		fmt.Println("  data          - Data analyst with data tools")
		fmt.Println("  datetime      - Scheduler with datetime tools")
		fmt.Println("  feeds         - News curator with feed tools")
		fmt.Println("  all-tools     - Demo agent with ALL tools")
		fmt.Println("\nExample:")
		fmt.Println("  agent-llm-builtin-tools research \"Find information about Go generics\"")
		os.Exit(1)
	}

	scenario := os.Args[1]

	// Get prompt from args or use default
	var prompt string
	if len(os.Args) > 2 {
		prompt = strings.Join(os.Args[2:], " ")
	}

	// Create LLM provider from environment
	llmProvider, providerName, modelName, err := createProvider()
	if err != nil {
		fmt.Println("Note: No LLM API keys found. Using mock provider for demonstration.")
		fmt.Println("Set ANTHROPIC_API_KEY, OPENAI_API_KEY, or GEMINI_API_KEY for real LLM usage.")
		fmt.Println("Tip: Set DEBUG=1 to see detailed logging of agent execution.")
		fmt.Println()
		llmProvider = createMockProvider(scenario)
		providerName = "mock"
		modelName = "mock-model"
	}

	fmt.Printf("Using %s provider with model %s\n\n", providerName, modelName)

	// Run the selected scenario
	ctx := context.Background()

	switch scenario {
	case "research":
		runResearchAssistant(ctx, llmProvider, prompt)
	case "files":
		runFileManager(ctx, llmProvider, prompt)
	case "system":
		runSystemAdmin(ctx, llmProvider, prompt)
	case "data":
		runDataAnalyst(ctx, llmProvider, prompt)
	case "datetime":
		runScheduler(ctx, llmProvider, prompt)
	case "feeds":
		runNewsCurator(ctx, llmProvider, prompt)
	case "all-tools":
		runAllToolsDemo(ctx, llmProvider, prompt)
	default:
		fmt.Printf("Unknown scenario: %s\n", scenario)
		os.Exit(1)
	}
}

func createProvider() (ldomain.Provider, string, string, error) {
	// Try to create provider from environment
	if os.Getenv("OPENAI_API_KEY") != "" {
		return provider.NewOpenAIProvider(
			os.Getenv("OPENAI_API_KEY"),
			"gpt-4o",
		), "openai", "gpt-4o", nil
	}

	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		return provider.NewAnthropicProvider(
			os.Getenv("ANTHROPIC_API_KEY"),
			"claude-3-7-sonnet-latest",
		), "anthropic", "claude-3-7-sonnet-latest", nil
	}

	if os.Getenv("GEMINI_API_KEY") != "" {
		return provider.NewGeminiProvider(
			os.Getenv("GEMINI_API_KEY"),
			"gemini-2.0-flash",
		), "gemini", "gemini-2.0-flash", nil
	}

	// Try to create from GO_LLMS environment variables
	llmProvider, providerName, modelName, err := llmutil.ProviderFromEnv()
	if err == nil {
		return llmProvider, providerName, modelName, nil
	}

	return nil, "", "", fmt.Errorf("no LLM provider configured. Set OPENAI_API_KEY, ANTHROPIC_API_KEY, or GEMINI_API_KEY")
}

// Research Assistant with Web Tools
func runResearchAssistant(ctx context.Context, llmProvider ldomain.Provider, customPrompt string) {
	fmt.Println("=== Research Assistant Demo ===")

	deps := core.LLMDeps{
		Provider: llmProvider,
	}
	agent := core.NewLLMAgent("research-assistant", "Research Assistant", deps)

	// Add web tools
	agent.AddTool(tools.MustGetTool("web_search"))
	agent.AddTool(tools.MustGetTool("web_fetch"))
	agent.AddTool(tools.MustGetTool("web_scrape"))
	agent.AddTool(tools.MustGetTool("file_write"))

	agent.SetSystemPrompt(`You are a research assistant with access to web tools.

Available tools:
- web_search: Search the web for information on any topic. Returns a list of results with titles, URLs, and snippets.
- web_fetch: Retrieve the full content of a web page given its URL.
- web_scrape: Extract structured data from HTML pages using CSS selectors.
- file_write: Save your research findings to a file.

When researching a topic:
1. Start with web_search to find relevant sources
2. Use web_fetch to get detailed content from the most promising URLs
3. Use web_scrape if you need to extract specific structured data
4. Compile your findings and save a summary using file_write

Always cite your sources with URLs and verify information from multiple sources when possible.`)

	// Default prompt if none provided
	if customPrompt == "" {
		customPrompt = "Research the latest developments in Go programming language and create a summary report"
	}

	runAgent(ctx, agent, customPrompt)
}

// File Manager with File Tools
func runFileManager(ctx context.Context, llmProvider ldomain.Provider, customPrompt string) {
	fmt.Println("=== File Manager Demo ===")

	deps := core.LLMDeps{
		Provider: llmProvider,
	}
	agent := core.NewLLMAgent("file-manager", "File Manager", deps)

	// Add file tools
	agent.AddTool(tools.MustGetTool("file_list"))
	agent.AddTool(tools.MustGetTool("file_read"))
	agent.AddTool(tools.MustGetTool("file_write"))
	agent.AddTool(tools.MustGetTool("file_move"))
	agent.AddTool(tools.MustGetTool("file_search"))

	agent.SetSystemPrompt(`You are a file management assistant.

Available tools:
- file_list: List files in a directory with optional filtering by pattern, size, or date
- file_read: Read the contents of a file (supports large files and specific line ranges)
- file_write: Write or append content to a file with optional backup
- file_move: Move or rename files and directories
- file_search: Search for content within files using regex patterns

Safety guidelines:
- Always use relative paths when possible
- Ask for confirmation before overwriting files
- Be careful with file_move operations
- Create backups when modifying important files
- Explain what each operation will do before executing`)

	if customPrompt == "" {
		customPrompt = "List all .go files in the current directory and show me the contents of the README if it exists"
	}

	runAgent(ctx, agent, customPrompt)
}

// System Admin with System Tools
func runSystemAdmin(ctx context.Context, llmProvider ldomain.Provider, customPrompt string) {
	fmt.Println("=== System Admin Demo ===")

	deps := core.LLMDeps{
		Provider: llmProvider,
	}
	agent := core.NewLLMAgent("system-admin", "System Administrator", deps)

	// Add system tools
	agent.AddTool(tools.MustGetTool("get_system_info"))
	agent.AddTool(tools.MustGetTool("process_list"))
	agent.AddTool(tools.MustGetTool("get_environment_variable"))
	agent.AddTool(tools.MustGetTool("execute_command"))

	agent.SetSystemPrompt(`You are a system administration assistant.

Available tools:
- get_system_info: Get comprehensive system information including OS, memory, and runtime details
- process_list: List running processes with filtering and sorting options
- get_environment_variable: Read environment variables (with pattern matching)
- execute_command: Execute system commands (use with extreme caution)

Safety rules:
- Only run read-only commands unless explicitly authorized
- Always explain what a command will do before executing
- Never run commands that could harm the system
- Check system resources before running intensive operations
- For execute_command, prefer safe commands like 'ls', 'pwd', 'echo', 'date'
- Avoid commands that modify system state or require sudo`)

	if customPrompt == "" {
		customPrompt = "Show me system information and list the top 5 processes by memory usage"
	}

	runAgent(ctx, agent, customPrompt)
}

// Data Analyst with Data Tools
func runDataAnalyst(ctx context.Context, llmProvider ldomain.Provider, customPrompt string) {
	fmt.Println("=== Data Analyst Demo ===")

	deps := core.LLMDeps{
		Provider: llmProvider,
	}
	agent := core.NewLLMAgent("data-analyst", "Data Analyst", deps)

	// Add data tools
	agent.AddTool(tools.MustGetTool("json_process"))
	agent.AddTool(tools.MustGetTool("csv_process"))
	agent.AddTool(tools.MustGetTool("xml_process"))
	agent.AddTool(tools.MustGetTool("data_transform"))

	agent.SetSystemPrompt(`You are a data analyst with tools for processing various data formats.

Available tools:
- json_process: Parse, query (JSONPath), and transform JSON data
- csv_process: Parse, filter, and analyze CSV data with statistics
- xml_process: Parse and query XML data with XPath
- data_transform: Apply transformations like filter, map, reduce, sort, and group_by

For each tool:
- json_process operations: parse, query, flatten, prettify, minify, extract_keys, extract_values
- csv_process operations: parse, filter, transform, stats
- xml_process operations: parse, query, to_json
- data_transform operations: filter, map, reduce, sort, group_by, unique, reverse

When analyzing data:
1. First identify the data format
2. Parse and validate the data
3. Apply appropriate queries or transformations
4. Generate insights or statistics
5. Present findings clearly`)

	if customPrompt == "" {
		customPrompt = `Analyze this JSON data and tell me the average age: 
{"users": [{"name": "Alice", "age": 30}, {"name": "Bob", "age": 25}, {"name": "Charlie", "age": 35}]}`
	}

	runAgent(ctx, agent, customPrompt)
}

// Scheduler with DateTime Tools
func runScheduler(ctx context.Context, llmProvider ldomain.Provider, customPrompt string) {
	fmt.Println("=== Scheduler Demo ===")

	deps := core.LLMDeps{
		Provider: llmProvider,
	}
	agent := core.NewLLMAgent("scheduler", "Scheduler", deps)

	// Add datetime tools
	agent.AddTool(tools.MustGetTool("datetime_now"))
	agent.AddTool(tools.MustGetTool("datetime_parse"))
	agent.AddTool(tools.MustGetTool("datetime_calculate"))
	agent.AddTool(tools.MustGetTool("datetime_format"))
	agent.AddTool(tools.MustGetTool("datetime_compare"))
	agent.AddTool(tools.MustGetTool("datetime_info"))
	agent.AddTool(tools.MustGetTool("datetime_convert"))

	agent.SetSystemPrompt(`You are a scheduling assistant with comprehensive date/time tools.

Available tools:
- datetime_now: Get current date/time in any timezone
- datetime_parse: Parse dates from various formats including relative dates
- datetime_calculate: Add/subtract time, find business days, calculate durations
- datetime_format: Format dates in standard or custom formats, with localization
- datetime_compare: Compare dates, check ranges, find earliest/latest
- datetime_info: Get date properties like day of week, week number, days in month
- datetime_convert: Convert between timezones

When scheduling:
1. Always clarify the timezone if not specified
2. Consider business days vs calendar days
3. Account for holidays when mentioned
4. Provide times in user's local timezone when possible
5. Use relative descriptions ("next Monday", "in 3 days") when appropriate`)

	if customPrompt == "" {
		customPrompt = "What day of the week will it be 30 days from now? Also, calculate how many business days are between now and then."
	}

	runAgent(ctx, agent, customPrompt)
}

// News Curator with Feed Tools
func runNewsCurator(ctx context.Context, llmProvider ldomain.Provider, customPrompt string) {
	fmt.Println("=== News Curator Demo ===")

	deps := core.LLMDeps{
		Provider: llmProvider,
	}
	agent := core.NewLLMAgent("news-curator", "News Curator", deps)

	// Add feed tools
	agent.AddTool(tools.MustGetTool("feed_discover"))
	agent.AddTool(tools.MustGetTool("feed_fetch"))
	agent.AddTool(tools.MustGetTool("feed_filter"))
	agent.AddTool(tools.MustGetTool("feed_aggregate"))
	agent.AddTool(tools.MustGetTool("feed_extract"))

	agent.SetSystemPrompt(`You are a news curator with tools for processing RSS/Atom feeds.

Available tools:
- feed_discover: Auto-discover feed URLs from websites
- feed_fetch: Retrieve and parse RSS/Atom/JSON feeds
- feed_filter: Filter feed items by date, keywords, author, or tags
- feed_aggregate: Combine multiple feeds and remove duplicates
- feed_extract: Extract specific fields from feed items

When curating news:
1. Use feed_discover to find feeds if only given a website URL
2. Use feed_fetch to retrieve feed content
3. Apply feed_filter to find relevant items
4. Use feed_aggregate when working with multiple sources
5. Use feed_extract to get specific data like titles or summaries

Always include publication dates and source attribution.`)

	if customPrompt == "" {
		customPrompt = "Discover feeds from https://news.ycombinator.com and show me the latest 5 items"
	}

	runAgent(ctx, agent, customPrompt)
}

// Demo with ALL tools
func runAllToolsDemo(ctx context.Context, llmProvider ldomain.Provider, customPrompt string) {
	fmt.Println("=== All Tools Demo ===")

	deps := core.LLMDeps{
		Provider: llmProvider,
	}
	agent := core.NewLLMAgent("universal-assistant", "Universal Assistant", deps)

	// Add ALL tools from all categories
	allTools := tools.Tools.List()
	toolCount := 0
	for _, entry := range allTools {
		agent.AddTool(entry.Component)
		toolCount++
	}

	fmt.Printf("Loaded %d tools across all categories\n", toolCount)

	agent.SetSystemPrompt(`You are a universal assistant with access to ALL available tools.

Tool Categories:
- Web Tools: web_search, web_fetch, web_scrape, http_request
- File Tools: file_list, file_read, file_write, file_delete, file_move, file_search  
- System Tools: get_system_info, process_list, get_environment_variable, execute_command
- Data Tools: json_process, csv_process, xml_process, data_transform
- DateTime Tools: datetime_now, datetime_parse, datetime_calculate, datetime_format, datetime_compare, datetime_info, datetime_convert
- Feed Tools: feed_discover, feed_fetch, feed_filter, feed_aggregate, feed_convert, feed_extract

Choose the appropriate tools based on the task. Be careful with destructive operations (file_delete, execute_command).
Always explain your tool choices and what you're doing.`)

	if customPrompt == "" {
		customPrompt = "Show me what tools you have available and demonstrate using one from each category"
	}

	runAgent(ctx, agent, customPrompt)
}

// Helper function to run an agent
func runAgent(ctx context.Context, agent *core.LLMAgent, prompt string) {
	fmt.Printf("\nPrompt: %s\n", prompt)
	fmt.Println("\nProcessing...")

	// Create initial state
	state := domain.NewState()
	state.Set("user_input", prompt)

	// Add hook to see tool calls
	agent.WithHook(&toolCallLogger{})

	// Add debug logging if DEBUG is enabled
	if os.Getenv("DEBUG") == "1" {
		opts := &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}
		logger := slog.New(slog.NewTextHandler(os.Stderr, opts))
		loggingHook := core.NewLoggingHook(logger, core.LogLevelDebug)
		agent.WithHook(loggingHook)
		log.Println("Debug logging enabled")
	}

	// Run the agent
	result, err := agent.Run(ctx, state)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	// Display response
	if response, ok := result.Get("response"); ok {
		fmt.Printf("\nResponse:\n%v\n", response)
	} else if output, ok := result.Get("output"); ok {
		fmt.Printf("\nOutput:\n%v\n", output)
	} else {
		fmt.Printf("\nFull Result State:\n")
		for k, v := range result.Values() {
			fmt.Printf("  %s: %v\n", k, v)
		}
	}
}

// toolCallLogger logs tool calls for visibility
type toolCallLogger struct{}

// BeforeGenerate is called before generating a response
func (h *toolCallLogger) BeforeGenerate(ctx context.Context, messages []ldomain.Message) {
	// We don't log message generation in this example
}

// AfterGenerate is called after generating a response
func (h *toolCallLogger) AfterGenerate(ctx context.Context, response ldomain.Response, err error) {
	// We don't log message generation in this example
}

// BeforeToolCall is called before executing a tool
func (h *toolCallLogger) BeforeToolCall(ctx context.Context, tool string, params map[string]interface{}) {
	fmt.Printf("[Tool Call] %s with params: %v\n", tool, params)
}

// AfterToolCall is called after executing a tool
func (h *toolCallLogger) AfterToolCall(ctx context.Context, tool string, result interface{}, err error) {
	if err != nil {
		fmt.Printf("[Tool Error] %s: %v\n", tool, err)
	} else {
		// Truncate large results for display
		resultStr := fmt.Sprintf("%v", result)
		if len(resultStr) > 200 {
			resultStr = resultStr[:200] + "..."
		}
		fmt.Printf("[Tool Result] %s: %s\n", tool, resultStr)
	}
}

// createMockProvider creates a mock provider for the given scenario
func createMockProvider(scenario string) ldomain.Provider {
	mockProvider := provider.NewMockProvider()

	mockProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		// Extract the last user message
		var lastUserMsg string
		for i := len(messages) - 1; i >= 0; i-- {
			if messages[i].Role == "user" {
				for _, part := range messages[i].Content {
					if part.Type == ldomain.ContentTypeText {
						lastUserMsg = part.Text
						break
					}
				}
				if lastUserMsg != "" {
					break
				}
			}
		}

		// Check if this is a tool result
		if strings.Contains(lastUserMsg, "Tool results:") {
			// Return appropriate response based on scenario
			switch scenario {
			case "research":
				return ldomain.Response{
					Content: "Based on my web search, I found that Go generics were introduced in Go 1.18. The key features include type parameters, type constraints, and type inference. This is a major addition to the language that enables writing more reusable code.",
				}, nil
			case "files":
				return ldomain.Response{
					Content: "I've listed the Go files in the current directory. I found main.go and several test files. The README.md contains project documentation about the go-llms library.",
				}, nil
			case "system":
				return ldomain.Response{
					Content: "The system is running Linux with 16GB of memory. The top 5 processes by memory usage are: Chrome (2.5GB), VS Code (1.8GB), Docker (1.2GB), Go compiler (800MB), and System processes (600MB).",
				}, nil
			case "data":
				return ldomain.Response{
					Content: "I analyzed the JSON data. The average age of the users is 30 years. Alice is 30, Bob is 25, and Charlie is 35.",
				}, nil
			case "datetime":
				return ldomain.Response{
					Content: "30 days from now will be a Tuesday. There are 22 business days between now and then (excluding weekends).",
				}, nil
			case "feeds":
				return ldomain.Response{
					Content: "I discovered the Hacker News RSS feed and fetched the latest 5 items. The top stories include discussions about AI, programming languages, and startup news.",
				}, nil
			default:
				return ldomain.Response{
					Content: "I've demonstrated using tools from each category. The system is working properly.",
				}, nil
			}
		}

		// Initial tool call based on scenario
		switch scenario {
		case "research":
			return ldomain.Response{
				Content: `{"tool": "web_search", "params": {"query": "Go generics introduction features"}}`,
			}, nil
		case "files":
			return ldomain.Response{
				Content: `{"tool": "file_list", "params": {"path": ".", "pattern": "*.go"}}`,
			}, nil
		case "system":
			return ldomain.Response{
				Content: `{"tool": "get_system_info", "params": {"include_memory": true}}`,
			}, nil
		case "data":
			return ldomain.Response{
				Content: `{"tool": "json_process", "params": {"data": "{\"users\": [{\"name\": \"Alice\", \"age\": 30}, {\"name\": \"Bob\", \"age\": 25}, {\"name\": \"Charlie\", \"age\": 35}]}", "operation": "query", "query": "$.users[*].age"}}`,
			}, nil
		case "datetime":
			return ldomain.Response{
				Content: `{"tool": "datetime_calculate", "params": {"operation": "add", "value": 30, "unit": "days"}}`,
			}, nil
		case "feeds":
			return ldomain.Response{
				Content: `{"tool": "feed_discover", "params": {"url": "https://news.ycombinator.com"}}`,
			}, nil
		default:
			return ldomain.Response{
				Content: `{"tool": "get_system_info", "params": {}}`,
			}, nil
		}
	})

	return mockProvider
}
