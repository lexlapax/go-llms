package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
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
	"github.com/lexlapax/go-llms/pkg/schema/validation"
	structuredProcessor "github.com/lexlapax/go-llms/pkg/structured/processor"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "go-llms",
		Short: "A Go library for LLM-powered applications with structured outputs",
		Long: `Go-LLMs is a Go library for creating LLM-powered applications with 
structured outputs and type safety. It provides interfaces for LLM providers, 
schema validation, structured output processing, and agent workflows with tools.`,
	}
)

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.go-llms.yaml)")
	rootCmd.PersistentFlags().String("provider", "openai", "LLM provider to use (openai, anthropic, mock)")
	rootCmd.PersistentFlags().String("model", "", "Model to use with the provider")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringP("output", "o", "text", "Output format (text, json)")

	if err := viper.BindPFlag("provider", rootCmd.PersistentFlags().Lookup("provider")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding provider flag: %v\n", err)
	}
	if err := viper.BindPFlag("model", rootCmd.PersistentFlags().Lookup("model")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding model flag: %v\n", err)
	}
	if err := viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding verbose flag: %v\n", err)
	}
	if err := viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding output flag: %v\n", err)
	}

	// Add subcommands
	rootCmd.AddCommand(newChatCmd())
	rootCmd.AddCommand(newCompleteCmd())
	rootCmd.AddCommand(newAgentCmd())
	rootCmd.AddCommand(newStructuredCmd())
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".go-llms")

		// Also look for config in .config/go-llms
		configDir := filepath.Join(home, ".config", "go-llms")
		viper.AddConfigPath(configDir)
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	// Set default provider models
	viper.SetDefault("providers.openai.default_model", "gpt-4o")
	viper.SetDefault("providers.anthropic.default_model", "claude-3-5-sonnet-latest")

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Helper function to get a provider from config
func getProvider() (string, string, error) {
	provider := viper.GetString("provider")
	var model string

	// Get the model, either from flag or from provider default
	model = viper.GetString("model")
	if model == "" {
		model = viper.GetString(fmt.Sprintf("providers.%s.default_model", provider))
		if model == "" {
			return "", "", fmt.Errorf("no model specified and no default model configured for provider %s", provider)
		}
	}

	return provider, model, nil
}

// Helper to get API key for a provider
func getAPIKey(provider string) (string, error) {
	key := viper.GetString(fmt.Sprintf("providers.%s.api_key", provider))
	if key == "" {
		// Try environment variable as fallback with uppercase format
		envVar := fmt.Sprintf("%s_API_KEY", strings.ToUpper(provider))
		key = os.Getenv(envVar)
		if key == "" {
			return "", fmt.Errorf("no API key configured for provider %s. Set it in config file or %s environment variable", provider, envVar)
		}
	}
	return key, nil
}

// addRequestedTools adds the requested tools to the agent
func addRequestedTools(agent agentDomain.Agent, toolNames []string) {
	for _, name := range toolNames {
		switch strings.ToLower(name) {
		case "calculator":
			agent.AddTool(tools.NewTool(
				"calculator",
				"Perform mathematical calculations",
				func(params struct {
					Expression string `json:"expression"`
				}) (float64, error) {
					// Simple calculator that handles basic operations (not a full parser)
					// Split by operators and handle each case
					parts := strings.Split(params.Expression, "*")
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

					parts = strings.Split(params.Expression, "/")
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

					parts = strings.Split(params.Expression, "+")
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

					parts = strings.Split(params.Expression, "-")
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

					return 0, fmt.Errorf("unsupported operation in expression: %s", params.Expression)
				},
				&schemaDomain.Schema{
					Type: "object",
					Properties: map[string]schemaDomain.Property{
						"expression": {
							Type:        "string",
							Description: "The mathematical expression to evaluate (supports +, -, *, /)",
						},
					},
					Required: []string{"expression"},
				},
			))

		case "date":
			agent.AddTool(tools.NewTool(
				"get_current_date",
				"Get the current date and time",
				func() map[string]string {
					now := time.Now()
					return map[string]string{
						"date":       now.Format("2006-01-02"),
						"time":       now.Format("15:04:05"),
						"year":       strconv.Itoa(now.Year()),
						"month":      now.Month().String(),
						"day":        strconv.Itoa(now.Day()),
						"weekday":    now.Weekday().String(),
						"timezone":   now.Location().String(),
						"unix_epoch": strconv.FormatInt(now.Unix(), 10),
					}
				},
				&schemaDomain.Schema{
					Type:        "object",
					Description: "Returns the current date and time information",
				},
			))

		case "web", "web_fetch":
			agent.AddTool(tools.NewTool(
				"web_fetch",
				"Fetch content from a URL",
				func(params struct {
					URL string `json:"url"`
				}) (map[string]string, error) {
					// Create a client with a timeout
					client := &http.Client{
						Timeout: 30 * time.Second,
					}

					// Validate URL (basic check)
					if !strings.HasPrefix(params.URL, "http://") && !strings.HasPrefix(params.URL, "https://") {
						return nil, fmt.Errorf("invalid URL: must begin with http:// or https://")
					}

					// Make the request
					resp, err := client.Get(params.URL)
					if err != nil {
						return nil, fmt.Errorf("error fetching URL: %v", err)
					}
					defer resp.Body.Close()

					// Check response status
					if resp.StatusCode != http.StatusOK {
						return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
					}

					// Read response body (limit to 100KB)
					maxBytes := int64(100 * 1024) // 100KB limit
					body, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes))
					if err != nil {
						return nil, fmt.Errorf("error reading response: %v", err)
					}

					// Return the content and metadata
					return map[string]string{
						"content":      string(body),
						"status_code":  strconv.Itoa(resp.StatusCode),
						"content_type": resp.Header.Get("Content-Type"),
						"url":          params.URL,
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

		case "read_file":
			agent.AddTool(tools.NewTool(
				"read_file",
				"Read content from a file",
				func(params struct {
					Path string `json:"path"`
				}) (string, error) {
					// Validate path (basic security check)
					cleanPath := filepath.Clean(params.Path)
					if strings.Contains(cleanPath, "..") {
						return "", fmt.Errorf("path traversal not allowed")
					}

					// Make path absolute if not already
					if !filepath.IsAbs(cleanPath) {
						workDir, err := os.Getwd()
						if err != nil {
							return "", fmt.Errorf("error getting working directory: %v", err)
						}
						cleanPath = filepath.Join(workDir, cleanPath)
					}

					// Check if file exists
					if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
						return "", fmt.Errorf("file not found: %s", cleanPath)
					}

					// Read file content
					content, err := os.ReadFile(cleanPath)
					if err != nil {
						return "", fmt.Errorf("error reading file: %v", err)
					}

					return string(content), nil
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

		case "write_file":
			agent.AddTool(tools.NewTool(
				"write_file",
				"Write content to a file",
				func(params struct {
					Path    string `json:"path"`
					Content string `json:"content"`
				}) (map[string]string, error) {
					// Validate path (basic security check)
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

					// Write content to file
					err := os.WriteFile(cleanPath, []byte(params.Content), 0644)
					if err != nil {
						return nil, fmt.Errorf("error writing to file: %v", err)
					}

					return map[string]string{
						"success": "true",
						"path":    cleanPath,
						"bytes":   strconv.Itoa(len(params.Content)),
						"message": "File written successfully",
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

		case "execute", "execute_command":
			agent.AddTool(tools.NewTool(
				"execute_command",
				"Execute a shell command",
				func(params struct {
					Command string `json:"command"`
					Timeout int    `json:"timeout,omitempty"`
				}) (map[string]string, error) {
					// Set a reasonable default timeout
					timeout := 10 * time.Second
					if params.Timeout > 0 {
						timeout = time.Duration(params.Timeout) * time.Second
					}

					// Create a context with timeout
					ctx, cancel := context.WithTimeout(context.Background(), timeout)
					defer cancel()

					// Create command
					cmd := exec.CommandContext(ctx, "sh", "-c", params.Command)

					// Capture stdout and stderr
					var stdout, stderr bytes.Buffer
					cmd.Stdout = &stdout
					cmd.Stderr = &stderr

					// Run the command
					err := cmd.Run()

					// Build the response
					result := map[string]string{
						"stdout":  stdout.String(),
						"stderr":  stderr.String(),
						"command": params.Command,
					}

					if err != nil {
						if ctx.Err() == context.DeadlineExceeded {
							return nil, fmt.Errorf("command timed out after %d seconds", params.Timeout)
						}
						result["error"] = err.Error()
						result["success"] = "false"
					} else {
						result["success"] = "true"
						result["exit_code"] = "0"
					}

					return result, nil
				},
				&schemaDomain.Schema{
					Type: "object",
					Properties: map[string]schemaDomain.Property{
						"command": {
							Type:        "string",
							Description: "The command to execute",
						},
						"timeout": {
							Type:        "integer",
							Description: "Timeout in seconds (default 10)",
							Minimum:     float64Ptr(1),
							Maximum:     float64Ptr(60),
						},
					},
					Required: []string{"command"},
				},
			))
		}
	}
}

// Helper function for creating float pointers
func float64Ptr(v float64) *float64 {
	return &v
}

// Initialize each command (stub implementations for now)

func newChatCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chat",
		Short: "Start an interactive chat with an LLM",
		Run: func(cmd *cobra.Command, args []string) {
			// Get provider type and model
			providerType, modelName, err := getProvider()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			// Get API key
			apiKey, err := getAPIKey(providerType)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			// Get settings from flags
			systemPrompt, _ := cmd.Flags().GetString("system")
			temperature, _ := cmd.Flags().GetFloat32("temperature")
			maxTokens, _ := cmd.Flags().GetInt("max-tokens")

			// Create the LLM provider based on provider type
			var llmProvider llmDomain.Provider
			switch providerType {
			case "openai":
				llmProvider = provider.NewOpenAIProvider(apiKey, modelName)
			case "anthropic":
				llmProvider = provider.NewAnthropicProvider(apiKey, modelName)
			case "mock":
				llmProvider = provider.NewMockProvider()
			default:
				fmt.Fprintf(os.Stderr, "Unsupported provider: %s\n", providerType)
				os.Exit(1)
			}

			// Initialize chat with system message if provided
			messages := make([]llmDomain.Message, 0)
			if systemPrompt != "" {
				messages = append(messages, llmDomain.Message{
					Role:    llmDomain.RoleSystem,
					Content: systemPrompt,
				})
			}

			fmt.Printf("Chat session started with %s using %s\n", providerType, modelName)
			fmt.Println("Type 'exit' or 'quit' to end the session")
			fmt.Println("-----------------------------------------")

			// Set up generation options
			options := []llmDomain.Option{
				llmDomain.WithTemperature(float64(temperature)),
				llmDomain.WithMaxTokens(maxTokens),
			}

			// Main chat loop
			ctx := context.Background()
			scanner := bufio.NewScanner(os.Stdin)
			for {
				fmt.Print("\nUser: ")
				if !scanner.Scan() {
					break
				}

				userInput := scanner.Text()
				if userInput == "exit" || userInput == "quit" {
					fmt.Println("Ending chat session")
					break
				}

				// Add user message to context
				messages = append(messages, llmDomain.Message{
					Role:    llmDomain.RoleUser,
					Content: userInput,
				})

				// Generate assistant response
				response, err := llmProvider.GenerateMessage(ctx, messages, options...)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					continue
				}

				// Print response
				fmt.Print("Assistant: ")
				fmt.Println(response.Content)

				// Add assistant response to context for next round
				messages = append(messages, llmDomain.Message{
					Role:    llmDomain.RoleAssistant,
					Content: response.Content,
				})
			}
		},
	}
	cmd.Flags().StringP("system", "s", "", "System prompt for the chat")
	cmd.Flags().Float32P("temperature", "t", 0.7, "Temperature for generation")
	cmd.Flags().Int("max-tokens", 1000, "Maximum tokens to generate")
	return cmd
}

func newCompleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "complete [prompt]",
		Short: "One-shot text completion with an LLM",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get provider type and model
			providerType, modelName, err := getProvider()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			// Get API key
			apiKey, err := getAPIKey(providerType)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			// Get temperature and max tokens
			temperature, _ := cmd.Flags().GetFloat32("temperature")
			maxTokens, _ := cmd.Flags().GetInt("max-tokens")
			stream, _ := cmd.Flags().GetBool("stream")

			// Get the prompt - combine all args with spaces
			prompt := strings.Join(args, " ")

			// Create the LLM provider based on provider type
			var llmProvider llmDomain.Provider
			switch providerType {
			case "openai":
				llmProvider = provider.NewOpenAIProvider(apiKey, modelName)
			case "anthropic":
				llmProvider = provider.NewAnthropicProvider(apiKey, modelName)
			case "mock":
				llmProvider = provider.NewMockProvider()
			default:
				fmt.Fprintf(os.Stderr, "Unsupported provider: %s\n", providerType)
				os.Exit(1)
			}

			ctx := context.Background()

			// Set up generation options
			options := []llmDomain.Option{
				llmDomain.WithTemperature(float64(temperature)),
				llmDomain.WithMaxTokens(maxTokens),
			}

			// Perform the generation
			if stream {
				// Streaming mode
				fmt.Println("Streaming response:")
				tokenStream, err := llmProvider.Stream(ctx, prompt, options...)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Generation error: %v\n", err)
					os.Exit(1)
				}

				for token := range tokenStream {
					fmt.Print(token.Text)
					if token.Finished {
						fmt.Println()
					}
				}
			} else {
				// Standard completion
				response, err := llmProvider.Generate(ctx, prompt, options...)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Generation error: %v\n", err)
					os.Exit(1)
				}

				fmt.Println(response)
			}
		},
	}
	cmd.Flags().Float32P("temperature", "t", 0.7, "Temperature for generation")
	cmd.Flags().Int("max-tokens", 1000, "Maximum tokens to generate")
	cmd.Flags().Bool("stream", false, "Stream the response token by token")
	return cmd
}

func newAgentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent [input]",
		Short: "Execute agent workflows with tool integration",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get provider type and model
			providerType, modelName, err := getProvider()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			// Get API key
			apiKey, err := getAPIKey(providerType)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			// Get tool list and system prompt
			toolNames, _ := cmd.Flags().GetStringSlice("tools")
			systemPrompt, _ := cmd.Flags().GetString("system")
			schemaPath, _ := cmd.Flags().GetString("schema")
			verbose, _ := cmd.Flags().GetBool("verbose")
			// temperature is defined but not used since agent interface doesn't support setting it
			_, _ = cmd.Flags().GetFloat32("temperature")

			// Get the prompt - combine all args with spaces
			prompt := strings.Join(args, " ")

			// Create the LLM provider based on provider type
			var llmProvider llmDomain.Provider
			switch providerType {
			case "openai":
				llmProvider = provider.NewOpenAIProvider(apiKey, modelName)
			case "anthropic":
				llmProvider = provider.NewAnthropicProvider(apiKey, modelName)
			case "mock":
				llmProvider = provider.NewMockProvider()
			default:
				fmt.Fprintf(os.Stderr, "Unsupported provider: %s\n", providerType)
				os.Exit(1)
			}

			// Create an agent
			agent := workflow.NewAgent(llmProvider)

			// Set the model
			agent.WithModel(modelName)

			// Set system prompt if provided
			if systemPrompt != "" {
				agent.SetSystemPrompt(systemPrompt)
			} else {
				agent.SetSystemPrompt("You are a helpful assistant that can answer questions and use tools when necessary.")
			}

			// Set options (if we want to configure temperature, we'll have to add that method to the Agent interface)

			// Add logging hook if verbose
			var logger *slog.Logger
			if verbose {
				handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
					Level: slog.LevelDebug,
				})
				logger = slog.New(handler)
				agent.WithHook(workflow.NewLoggingHook(logger, workflow.LogLevelDetailed))
			}

			// Add metrics hook
			metricsHook := workflow.NewMetricsHook()
			agent.WithHook(metricsHook)

			// Add the requested tools
			addRequestedTools(agent, toolNames)

			// Create context with metrics
			ctx := workflow.WithMetrics(context.Background())

			var result interface{}
			var runErr error

			// Run with schema if provided
			if schemaPath != "" {
				// Read schema file
				schemaData, err := os.ReadFile(schemaPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error reading schema file: %v\n", err)
					os.Exit(1)
				}

				// Parse schema
				var schema schemaDomain.Schema
				err = json.Unmarshal(schemaData, &schema)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error parsing schema: %v\n", err)
					os.Exit(1)
				}

				// Run with schema
				result, runErr = agent.RunWithSchema(ctx, prompt, &schema)
			} else {
				// Run without schema
				result, runErr = agent.Run(ctx, prompt)
			}

			// Handle errors
			if runErr != nil {
				fmt.Fprintf(os.Stderr, "Agent execution error: %v\n", runErr)
				os.Exit(1)
			}

			// Display the result
			fmt.Println("\nAgent Result:")
			fmt.Println("--------------")

			// Format the result based on type
			switch v := result.(type) {
			case string:
				fmt.Println(v)
			case map[string]interface{}:
				resultJSON, _ := json.MarshalIndent(v, "", "  ")
				fmt.Println(string(resultJSON))
			default:
				resultJSON, _ := json.MarshalIndent(result, "", "  ")
				fmt.Println(string(resultJSON))
			}

			// Display metrics if verbose
			if verbose {
				fmt.Println("\nMetrics:")
				fmt.Println("--------------")
				metrics := metricsHook.GetMetrics()
				fmt.Printf("Total requests: %d\n", metrics.Requests)
				fmt.Printf("Tool calls: %d\n", metrics.ToolCalls)
				fmt.Printf("Errors: %d\n", metrics.ErrorCount)
				fmt.Printf("Total tokens: %d\n", metrics.TotalTokens)
				fmt.Printf("Average generation time: %.2f ms\n", metrics.AverageGenTimeMs)

				if len(metrics.ToolStats) > 0 {
					fmt.Println("\nTool Usage:")
					for tool, stats := range metrics.ToolStats {
						fmt.Printf("- %s: %d calls, avg: %.2f ms, fastest: %.2f ms, slowest: %.2f ms\n",
							tool, stats.Calls, stats.AverageTimeMs, stats.FastestCallMs, stats.SlowestCallMs)
					}
				}
			}
		},
	}

	cmd.Flags().StringSliceP("tools", "t", []string{"calculator", "date"}, "Tools to enable (comma-separated, available: calculator, date, web, read_file, write_file, execute_command)")
	cmd.Flags().StringP("system", "s", "", "System prompt for the agent")
	cmd.Flags().StringP("schema", "S", "", "Path to output schema file")
	cmd.Flags().Float32P("temperature", "T", 0.7, "Temperature for generation")
	return cmd
}

func newStructuredCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "structured [prompt]",
		Short: "Generate structured outputs using schemas",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get provider type and model
			providerType, modelName, err := getProvider()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			// Get API key
			apiKey, err := getAPIKey(providerType)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			// Get schema file path
			schemaPath, _ := cmd.Flags().GetString("schema")
			if schemaPath == "" {
				fmt.Fprintf(os.Stderr, "Error: schema file path is required\n")
				os.Exit(1)
			}

			// Read schema file
			schemaData, err := os.ReadFile(schemaPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading schema file: %v\n", err)
				os.Exit(1)
			}

			// Parse schema
			var schema schemaDomain.Schema
			err = json.Unmarshal(schemaData, &schema)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing schema: %v\n", err)
				os.Exit(1)
			}

			// Get settings from flags
			temperature, _ := cmd.Flags().GetFloat32("temperature")
			maxTokens, _ := cmd.Flags().GetInt("max-tokens")
			outputFile, _ := cmd.Flags().GetString("output-file")
			shouldValidate, _ := cmd.Flags().GetBool("validate")

			// Get the prompt - combine all args with spaces
			prompt := strings.Join(args, " ")

			// Create validator and processor
			validator := validation.NewValidator()
			processor := structuredProcessor.NewStructuredProcessor(validator)
			promptEnhancer := structuredProcessor.NewPromptEnhancer()

			// Enhance prompt with schema
			enhancedPrompt, err := promptEnhancer.Enhance(prompt, &schema)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error enhancing prompt: %v\n", err)
				os.Exit(1)
			}

			// Create the LLM provider based on provider type
			var llmProvider llmDomain.Provider
			switch providerType {
			case "openai":
				llmProvider = provider.NewOpenAIProvider(apiKey, modelName)
			case "anthropic":
				llmProvider = provider.NewAnthropicProvider(apiKey, modelName)
			case "mock":
				llmProvider = provider.NewMockProvider()
			default:
				fmt.Fprintf(os.Stderr, "Unsupported provider: %s\n", providerType)
				os.Exit(1)
			}

			ctx := context.Background()

			// Set up generation options
			options := []llmDomain.Option{
				llmDomain.WithTemperature(float64(temperature)),
				llmDomain.WithMaxTokens(maxTokens),
			}

			// Generate the structured output
			fmt.Println("Generating structured output...")
			response, err := llmProvider.Generate(ctx, enhancedPrompt, options...)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Generation error: %v\n", err)
				os.Exit(1)
			}

			// Process the raw response to extract and validate structured data
			var result interface{}
			if shouldValidate {
				result, err = processor.Process(&schema, response)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Processing error: %v\n", err)
					fmt.Println("\nRaw response:")
					fmt.Println(response)
					os.Exit(1)
				}
			} else {
				// Still try to extract JSON but don't validate
				err = json.Unmarshal([]byte(response), &result)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error extracting JSON: %v\n", err)
					fmt.Println("\nRaw response:")
					fmt.Println(response)
					os.Exit(1)
				}
			}

			// Format the result
			resultJSON, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error formatting result: %v\n", err)
				os.Exit(1)
			}

			// Output the result
			if outputFile != "" {
				err = os.WriteFile(outputFile, resultJSON, 0644)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error writing to file: %v\n", err)
					os.Exit(1)
				}
				fmt.Printf("Structured output written to %s\n", outputFile)
			} else {
				fmt.Println("\nStructured Output:")
				fmt.Println(string(resultJSON))
			}
		},
	}
	cmd.Flags().StringP("schema", "s", "", "Path to schema file")
	cmd.Flags().StringP("output-file", "O", "", "File to write result to")
	cmd.Flags().Bool("validate", true, "Validate output against schema")
	cmd.Flags().Float32P("temperature", "t", 0.7, "Temperature for generation")
	cmd.Flags().Int("max-tokens", 1000, "Maximum tokens to generate")
	return cmd
}
