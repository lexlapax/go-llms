package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	llmDomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	"github.com/willabides/kongplete"
)

// CLI represents our command-line interface
type CLI struct {
	Config   string `kong:"type='path',help='Config file location'"`
	Provider string `kong:"default='openai',help='LLM provider to use (openai, anthropic, gemini, mock)'"`
	Model    string `kong:"help='Model to use with the provider'"`
	Verbose  bool   `kong:"short='v',help='Enable verbose output'"`
	Output   string `kong:"short='o',default='text',enum='text,json',help='Output format'"`

	Chat       ChatCmd       `kong:"cmd,help='Interactive chat with an LLM'"`
	Complete   CompleteCmd   `kong:"cmd,help='One-shot text completion with an LLM'"`
	Agent      AgentCmd      `kong:"cmd,help='Run an agent with tools'"`
	Structured StructuredCmd `kong:"cmd,help='Extract structured data from text'"`
	Completion CompletionCmd `kong:"cmd,help='Generate shell completions'"`

	// For shell completion
	InstallCompletions kongplete.InstallCompletions `kong:"cmd,help='Install shell completions'"`
}

// ChatCmd represents the chat command
type ChatCmd struct {
	System      string  `kong:"short='s',help='System prompt for the chat'"`
	Temperature float32 `kong:"short='t',default='0.7',help='Temperature for generation'"`
	MaxTokens   int     `kong:"default='1000',help='Maximum tokens to generate'"`
	Stream      bool    `kong:"help='Stream responses in real-time'"`
	NoStream    bool    `kong:"help='Disable streaming'"`
}

// CompleteCmd represents the complete command
type CompleteCmd struct {
	Temperature float32  `kong:"short='t',default='0.7',help='Temperature for generation'"`
	MaxTokens   int      `kong:"default='1000',help='Maximum tokens to generate'"`
	Stream      bool     `kong:"help='Stream responses'"`
	Prompt      []string `kong:"arg,required,help='Prompt for completion'"`
}

// AgentCmd represents the agent command
type AgentCmd struct {
	Tools       []string `kong:"name='tool',short='t',help='Enable specific tools'"`
	System      string   `kong:"short='s',help='System prompt for the agent'"`
	Schema      string   `kong:"help='Path to schema file for tool responses'"`
	Temperature float32  `kong:"default='0.7',help='Temperature for generation'"`
	Prompt      []string `kong:"arg,required,help='Initial prompt for the agent'"`
}

// StructuredCmd represents the structured command
type StructuredCmd struct {
	Schema      string  `kong:"name='schema',short='s',help='Path to JSON schema file',type='existingfile'"`
	OutputFile  string  `kong:"name='output-file',help='Output file for extracted data',type='path'"`
	Validate    bool    `kong:"name='validate',help='Validate against schema without LLM'"`
	Temperature float32 `kong:"short='t',default='0.7',help='Temperature for generation'"`
	MaxTokens   int     `kong:"default='1000',help='Maximum tokens to generate'"`
	Prompt      string  `kong:"arg,required,help='Prompt or file containing prompt'"`
}

// CompletionCmd represents the completion command
type CompletionCmd struct {
	Shell string `kong:"arg,required,enum='bash,zsh,fish,powershell',help='Shell to generate completions for'"`
}

// Context holds shared state for commands
type Context struct {
	Config *Config
	CLI    *CLI
}

// NewContext creates a new context
func NewContext(cli *CLI, config *Config) *Context {
	return &Context{
		CLI:    cli,
		Config: config,
	}
}

// GetProviderInfo returns the configured provider and model
func (ctx *Context) GetProviderInfo() (string, string, error) {
	// CLI flags take precedence over config
	provider := ctx.CLI.Provider
	if provider == "" {
		provider = k.String("provider")
	}

	model := ctx.CLI.Model
	if model == "" {
		model = k.String("model")
		if model == "" {
			// Try to get default model for the provider
			model = k.String(fmt.Sprintf("providers.%s.default_model", provider))
			if model == "" {
				return "", "", fmt.Errorf("no model specified for provider %s", provider)
			}
		}
	}

	return provider, model, nil
}

// CreateProvider creates an LLM provider based on configuration
func (ctx *Context) CreateProvider() (llmDomain.Provider, error) {
	providerType, modelName, err := ctx.GetProviderInfo()
	if err != nil {
		return nil, err
	}

	apiKey, err := GetAPIKey(providerType)
	if err != nil && providerType != "mock" {
		return nil, err
	}

	switch providerType {
	case "openai":
		return provider.NewOpenAIProvider(apiKey, modelName), nil
	case "anthropic":
		return provider.NewAnthropicProvider(apiKey, modelName), nil
	case "gemini":
		return provider.NewGeminiProvider(apiKey, modelName), nil
	case "mock":
		return provider.NewMockProvider(), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerType)
	}
}

// Run implements the chat command
func (c *ChatCmd) Run(ctx *Context) error {
	if c.Stream && c.NoStream {
		return fmt.Errorf("cannot specify both --stream and --no-stream")
	}

	// Determine if we should stream
	shouldStream := c.Stream
	if !c.NoStream && !c.Stream {
		// Auto-enable streaming if we're in an interactive terminal
		shouldStream = isInteractiveTerminal()
	}

	// Create the LLM provider
	llmProvider, err := ctx.CreateProvider()
	if err != nil {
		return err
	}

	kctx := context.Background()

	// Set up generation options
	options := []llmDomain.Option{
		llmDomain.WithTemperature(float64(c.Temperature)),
		llmDomain.WithMaxTokens(c.MaxTokens),
	}

	// Initialize messages with system prompt if provided
	var messages []llmDomain.Message
	if c.System != "" {
		messages = append(messages, llmDomain.NewTextMessage(llmDomain.RoleSystem, c.System))
	}

	// If we have piped input, read it as the initial prompt
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// We have piped input
		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("error reading piped input: %w", err)
		}
		prompt := strings.TrimSpace(string(input))
		if prompt != "" {
			messages = append(messages, llmDomain.NewTextMessage(llmDomain.RoleUser, prompt))

			// Get response and exit (non-interactive mode)
			if shouldStream {
				return c.streamResponse(kctx, llmProvider, messages, options)
			}
			return c.generateResponse(kctx, llmProvider, messages, options)
		}
	}

	// Interactive chat loop
	fmt.Println("Chat with", k.String("provider"), "model:", k.String("model"))
	fmt.Println("Type 'exit' or 'quit' to end the conversation.")
	fmt.Println()

	// Interactive chat implementation would go here
	// (Similar to the original implementation)

	return nil
}

// Helper methods for chat command
func (c *ChatCmd) streamResponse(ctx context.Context, provider llmDomain.Provider, messages []llmDomain.Message, options []llmDomain.Option) error {
	// The Stream API is for single prompts, not message arrays
	// Let's convert the last user message to a prompt
	var prompt string
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == llmDomain.RoleUser {
			// Get the text content
			for _, part := range messages[i].Content {
				if part.Type == llmDomain.ContentTypeText {
					prompt = part.Text
					break
				}
			}
			break
		}
	}
	
	if prompt == "" {
		return fmt.Errorf("no user message found")
	}
	
	tokenStream, err := provider.Stream(ctx, prompt, options...)
	if err != nil {
		return fmt.Errorf("streaming error: %w", err)
	}

	for token := range tokenStream {
		fmt.Print(token.Text)
		if token.Finished {
			break
		}
	}
	fmt.Println()
	return nil
}

func (c *ChatCmd) generateResponse(ctx context.Context, provider llmDomain.Provider, messages []llmDomain.Message, options []llmDomain.Option) error {
	response, err := provider.GenerateMessage(ctx, messages, options...)
	if err != nil {
		return fmt.Errorf("generation error: %w", err)
	}

	// The response has Content as a string
	fmt.Println(response.Content)
	return nil
}

// Run implements the complete command
func (c *CompleteCmd) Run(ctx *Context) error {
	// Join all prompt arguments
	prompt := strings.Join(c.Prompt, " ")

	// Create the LLM provider
	llmProvider, err := ctx.CreateProvider()
	if err != nil {
		return err
	}

	kctx := context.Background()

	// Set up generation options
	options := []llmDomain.Option{
		llmDomain.WithTemperature(float64(c.Temperature)),
		llmDomain.WithMaxTokens(c.MaxTokens),
	}

	// Perform the generation
	if c.Stream {
		tokenStream, err := llmProvider.Stream(kctx, prompt, options...)
		if err != nil {
			return fmt.Errorf("streaming error: %w", err)
		}

		for token := range tokenStream {
			fmt.Print(token.Text)
			if token.Finished {
				fmt.Println()
			}
		}
	} else {
		response, err := llmProvider.Generate(kctx, prompt, options...)
		if err != nil {
			return fmt.Errorf("generation error: %w", err)
		}

		// The response is already a string
		fmt.Print(response)
		if ctx.CLI.Output == "json" {
			// Format as JSON if requested
			fmt.Println()
		}
	}

	return nil
}

// Run implements the completion command
func (c *CompletionCmd) Run(ctx *Context) error {
	// Generate shell completion script
	switch c.Shell {
	case "bash":
		return ctx.CLI.GenerateBashCompletion(os.Stdout)
	case "zsh":
		return ctx.CLI.GenerateZshCompletion(os.Stdout)
	case "fish":
		return ctx.CLI.GenerateFishCompletion(os.Stdout)
	case "powershell":
		return ctx.CLI.GeneratePowerShellCompletion(os.Stdout)
	default:
		return fmt.Errorf("unsupported shell: %s", c.Shell)
	}
}

// Helper functions
func isInteractiveTerminal() bool {
	stat, _ := os.Stdout.Stat()
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// GenerateBashCompletion generates bash completion script
func (cli *CLI) GenerateBashCompletion(w io.Writer) error {
	fmt.Fprintln(w, "# Bash completion for go-llms")
	fmt.Fprintln(w, "# Add this to your .bashrc or .bash_profile")
	fmt.Fprintln(w, "_go_llms_completions() {")
	fmt.Fprintln(w, "    local cur=${COMP_WORDS[COMP_CWORD]}")
	fmt.Fprintln(w, "    COMPREPLY=( $(compgen -W \"chat complete agent structured completion install-completions --help --version\" -- $cur) )")
	fmt.Fprintln(w, "}")
	fmt.Fprintln(w, "complete -F _go_llms_completions go-llms")
	return nil
}

// GenerateZshCompletion generates zsh completion script
func (cli *CLI) GenerateZshCompletion(w io.Writer) error {
	fmt.Fprintln(w, "#compdef go-llms")
	fmt.Fprintln(w, "# Zsh completion for go-llms")
	fmt.Fprintln(w, "_go_llms() {")
	fmt.Fprintln(w, "    local -a commands")
	fmt.Fprintln(w, "    commands=(")
	fmt.Fprintln(w, "        'chat:Interactive chat with an LLM'")
	fmt.Fprintln(w, "        'complete:One-shot text completion with an LLM'")
	fmt.Fprintln(w, "        'agent:Run an agent with tools'")
	fmt.Fprintln(w, "        'structured:Extract structured data from text'")
	fmt.Fprintln(w, "        'completion:Generate shell completions'")
	fmt.Fprintln(w, "        'install-completions:Install shell completions'")
	fmt.Fprintln(w, "    )")
	fmt.Fprintln(w, "    _describe 'command' commands")
	fmt.Fprintln(w, "}")
	fmt.Fprintln(w, "_go_llms")
	return nil
}

// GenerateFishCompletion generates fish completion script
func (cli *CLI) GenerateFishCompletion(w io.Writer) error {
	fmt.Fprintln(w, "# Fish completion for go-llms")
	fmt.Fprintln(w, "complete -c go-llms -f")
	fmt.Fprintln(w, "complete -c go-llms -n \"__fish_use_subcommand\" -a chat -d \"Interactive chat with an LLM\"")
	fmt.Fprintln(w, "complete -c go-llms -n \"__fish_use_subcommand\" -a complete -d \"One-shot text completion with an LLM\"")
	fmt.Fprintln(w, "complete -c go-llms -n \"__fish_use_subcommand\" -a agent -d \"Run an agent with tools\"")
	fmt.Fprintln(w, "complete -c go-llms -n \"__fish_use_subcommand\" -a structured -d \"Extract structured data from text\"")
	fmt.Fprintln(w, "complete -c go-llms -n \"__fish_use_subcommand\" -a completion -d \"Generate shell completions\"")
	fmt.Fprintln(w, "complete -c go-llms -n \"__fish_use_subcommand\" -a install-completions -d \"Install shell completions\"")
	return nil
}

// GeneratePowerShellCompletion generates PowerShell completion script
func (cli *CLI) GeneratePowerShellCompletion(w io.Writer) error {
	fmt.Fprintln(w, "# PowerShell completion for go-llms")
	fmt.Fprintln(w, "Register-ArgumentCompleter -Native -CommandName go-llms -ScriptBlock {")
	fmt.Fprintln(w, "    param($wordToComplete, $commandAst, $cursorPosition)")
	fmt.Fprintln(w, "    $commands = @(")
	fmt.Fprintln(w, "        'chat'")
	fmt.Fprintln(w, "        'complete'")
	fmt.Fprintln(w, "        'agent'")
	fmt.Fprintln(w, "        'structured'")
	fmt.Fprintln(w, "        'completion'")
	fmt.Fprintln(w, "        'install-completions'")
	fmt.Fprintln(w, "    )")
	fmt.Fprintln(w, "    $commands | Where-Object { $_ -like \"$wordToComplete*\" }")
	fmt.Fprintln(w, "}")
	return nil
}