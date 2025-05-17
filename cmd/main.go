package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	
	llmDomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
)

// Version information (set during build)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// Command line flags
var (
	configFile   = flag.String("c", "", "Config file location")
	providerFlag = flag.String("p", "", "LLM provider to use")
	modelFlag    = flag.String("m", "", "Model to use (overrides provider default)")
	verbose      = flag.Bool("v", false, "Enable verbose output")
	output       = flag.String("o", "text", "Output format (text or json)")
	help         = flag.Bool("h", false, "Show help")
	versionFlag  = flag.Bool("version", false, "Show version information")
)

func main() {
	flag.Usage = func() {
		fmt.Printf(`Go-LLMs - CLI for interacting with various LLM providers

Usage:
  %s [OPTIONS] COMMAND [ARGS...]

Commands:
  chat        Interactive chat with an LLM
  complete    One-shot text completion
  version     Show version information

Options:
`, filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	
	flag.Parse()
	
	// Handle help and version flags
	if *help {
		flag.Usage()
		os.Exit(0)
	}
	
	if *versionFlag {
		showVersion()
		os.Exit(0)
	}
	
	// Check for command
	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Error: No command specified\n\n")
		flag.Usage()
		os.Exit(1)
	}
	
	command := flag.Arg(0)
	
	// Initialize configuration
	if err := InitOptimizedConfig(*configFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
	
	// Override config with command line flags
	if *providerFlag != "" {
		config.Provider = *providerFlag
	}
	if *modelFlag != "" {
		config.Model = *modelFlag
	}
	if *verbose {
		config.Verbose = *verbose
	}
	if *output != "" {
		config.Output = *output
	}
	
	// Execute command
	ctx := context.Background()
	
	switch command {
	case "version":
		showVersion()
	case "chat":
		runMinimalChat(ctx)
	case "complete":
		runMinimalComplete(ctx)
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown command '%s'\n\n", command)
		flag.Usage()
		os.Exit(1)
	}
}

func showVersion() {
	fmt.Printf("go-llms version %s\n", version)
	fmt.Printf("commit: %s\n", commit)
	fmt.Printf("built at: %s\n", date)
}

func createMinimalProvider() (llmDomain.Provider, error) {
	providerName, modelName, err := GetOptimizedProvider()
	if err != nil {
		return nil, err
	}
	
	apiKey, err := GetOptimizedAPIKey(providerName)
	if err != nil {
		return nil, err
	}
	
	switch providerName {
	case "openai":
		return provider.NewOpenAIProvider(apiKey, modelName), nil
	case "anthropic":
		return provider.NewAnthropicProvider(apiKey, modelName), nil
	case "gemini":
		return provider.NewGeminiProvider(apiKey, modelName), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerName)
	}
}

func runMinimalChat(ctx context.Context) {
	// Parse chat-specific flags
	chatFlags := flag.NewFlagSet("chat", flag.ExitOnError)
	system := chatFlags.String("s", "", "System prompt to set context")
	
	chatFlags.Parse(flag.Args()[1:])
	
	provider, err := createMinimalProvider()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating provider: %v\n", err)
		os.Exit(1)
	}
	
	prompt := ""
	if *system != "" {
		prompt = *system + "\n\n"
	}
	
	fmt.Println("Chat mode - Type 'exit' to quit")
	
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		
		input := scanner.Text()
		if input == "exit" {
			return
		}
		if input == "" {
			continue
		}
		
		response, err := provider.Generate(ctx, prompt + input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}
		
		fmt.Println(response)
	}
	
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}
}

func runMinimalComplete(ctx context.Context) {
	// Parse complete-specific flags
	completeFlags := flag.NewFlagSet("complete", flag.ExitOnError)
	system := completeFlags.String("s", "", "System prompt to set context")
	
	completeFlags.Parse(flag.Args()[1:])
	
	if completeFlags.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: No prompt provided")
		fmt.Fprintln(os.Stderr, "Usage: go-llms complete [OPTIONS] PROMPT")
		os.Exit(1)
	}
	
	userPrompt := strings.Join(completeFlags.Args(), " ")
	prompt := userPrompt
	if *system != "" {
		prompt = *system + "\n\n" + userPrompt
	}
	
	provider, err := createMinimalProvider()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating provider: %v\n", err)
		os.Exit(1)
	}
	
	response, err := provider.Generate(ctx, prompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println(response)
}