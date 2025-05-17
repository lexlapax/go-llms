package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	
	"gopkg.in/yaml.v3"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	"github.com/lexlapax/go-llms/pkg/llm/domain"
)

// Simple config struct
type Config struct {
	Provider string `yaml:"provider"`
	Model    string `yaml:"model"`
	Providers struct {
		OpenAI struct {
			APIKey       string `yaml:"api_key"`
			DefaultModel string `yaml:"default_model"`
		} `yaml:"openai"`
		Anthropic struct {
			APIKey       string `yaml:"api_key"`
			DefaultModel string `yaml:"default_model"`
		} `yaml:"anthropic"`
	} `yaml:"providers"`
}

var (
	configFile = flag.String("config", "", "Config file path")
	providerFlag   = flag.String("provider", "openai", "LLM provider")
	modelFlag      = flag.String("model", "", "Model to use")
	verbose    = flag.Bool("verbose", false, "Verbose output")
)

func main() {
	flag.Parse()
	
	config := loadConfig(*configFile)
	
	// Override with flags
	if *providerFlag != "" {
		config.Provider = *providerFlag
	}
	if *modelFlag != "" {
		config.Model = *modelFlag
	}
	
	if flag.NArg() < 1 {
		fmt.Println("Usage: go-llms [chat|complete] [args...]")
		os.Exit(1)
	}
	
	command := flag.Arg(0)
	
	// Get provider and API key
	apiKey := getAPIKey(config)
	modelName := getModel(config)
	
	var llmProvider domain.Provider
	
	switch config.Provider {
	case "openai":
		llmProvider = provider.NewOpenAIProvider(apiKey, modelName)
	case "anthropic":
		llmProvider = provider.NewAnthropicProvider(apiKey, modelName)
	default:
		fmt.Fprintf(os.Stderr, "Unknown provider: %s\n", config.Provider)
		os.Exit(1)
	}
	
	ctx := context.Background()
	
	switch command {
	case "chat":
		runChat(ctx, llmProvider)
	case "complete":
		runComplete(ctx, llmProvider)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func loadConfig(configFile string) Config {
	config := Config{
		Provider: "openai",
	}
	config.Providers.OpenAI.DefaultModel = "gpt-4o"
	config.Providers.Anthropic.DefaultModel = "claude-3-5-sonnet-latest"
	
	// Try to load config file
	if configFile != "" {
		if data, err := os.ReadFile(configFile); err == nil {
			yaml.Unmarshal(data, &config)
		}
	} else {
		// Try standard locations
		home, _ := os.UserHomeDir()
		paths := []string{
			home + "/.go-llms.yaml",
			".go-llms.yaml",
		}
		for _, path := range paths {
			if data, err := os.ReadFile(path); err == nil {
				yaml.Unmarshal(data, &config)
				break
			}
		}
	}
	
	// Environment overrides
	if val := os.Getenv("GO_LLMS_PROVIDER"); val != "" {
		config.Provider = val
	}
	
	return config
}

func getAPIKey(config Config) string {
	var key string
	
	switch config.Provider {
	case "openai":
		key = config.Providers.OpenAI.APIKey
		if key == "" {
			key = os.Getenv("OPENAI_API_KEY")
		}
	case "anthropic":
		key = config.Providers.Anthropic.APIKey
		if key == "" {
			key = os.Getenv("ANTHROPIC_API_KEY")
		}
	}
	
	if key == "" {
		fmt.Fprintf(os.Stderr, "No API key for provider %s\n", config.Provider)
		os.Exit(1)
	}
	
	return key
}

func getModel(config Config) string {
	if config.Model != "" {
		return config.Model
	}
	
	switch config.Provider {
	case "openai":
		return config.Providers.OpenAI.DefaultModel
	case "anthropic":
		return config.Providers.Anthropic.DefaultModel
	default:
		return ""
	}
}

func runChat(ctx context.Context, llmProvider domain.Provider) {
	fmt.Println("Chat mode. Type 'exit' to quit.")
	
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		
		input := scanner.Text()
		if input == "exit" {
			break
		}
		
		resp, err := llmProvider.Generate(ctx, input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}
		
		fmt.Println(resp)
	}
}

func runComplete(ctx context.Context, llmProvider domain.Provider) {
	if flag.NArg() < 2 {
		fmt.Println("Usage: go-llms complete <prompt>")
		os.Exit(1)
	}
	
	prompt := strings.Join(flag.Args()[1:], " ")
	resp, err := llmProvider.Generate(ctx, prompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println(resp)
}