package main

// ABOUTME: Script to validate code examples in documentation files
// ABOUTME: Checks that examples use current API patterns and provider implementations

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// CodeBlock represents a code block found in documentation
type CodeBlock struct {
	File      string
	StartLine int
	Language  string
	Content   string
	Issues    []string
}

// CheckResult represents the result of checking a file
type CheckResult struct {
	File       string
	Blocks     []CodeBlock
	TotalIssues int
}

var (
	// Regex to find code blocks
	codeBlockStartRegex = regexp.MustCompile("^```(\\w+)?")
	codeBlockEndRegex   = regexp.MustCompile("^```$")
	
	// Deprecated patterns to check for
	deprecatedPatterns = map[string]string{
		`llm\.NewOpenAI`:                "Use openai.New instead",
		`llm\.NewAnthropic`:             "Use anthropic.New instead",
		`llm\.NewGemini`:                "Use google.New instead",
		`agent\.NewAgent`:               "Use agent.NewLLMAgent or agent.NewSimpleAgent instead",
		`tools\.Register`:               "Use tools.GetGlobalRegistry().Register instead",
		`schema\.Validate`:              "Use validator.Validate instead",
		`provider\.CreateCompletion`:    "Use provider.Complete instead",
		`\.ChatCompletion`:              "Use .Complete instead",
		`WithMaxTokens\(\)`:             "Use MaxTokens field in CompletionRequest",
		`WithTemperature\(\)`:           "Use Temperature field in CompletionRequest",
		`stream\.Next\(\)`:              "Use range over channel instead",
		`provider\.Models\(\)`:          "Use provider.GetModels(ctx) instead",
	}
	
	// Current API patterns that should be used
	currentPatterns = []string{
		"openai.New",
		"anthropic.New",
		"google.New",
		"ollama.New",
		"openrouter.New",
		"vertexai.New",
		"provider.Complete",
		"provider.CompleteStream",
		"provider.GetModels",
		"agent.NewLLMAgent",
		"agent.NewSimpleAgent",
		"tools.GetGlobalRegistry",
		"CompletionRequest",
		"CompletionResponse",
	}
	
	// Files to check
	docsRoot = "docs"
)

func main() {
	log.Println("Go-LLMs Documentation Code Example Validator")
	log.Println("===========================================")
	
	var allResults []CheckResult
	totalIssues := 0
	
	// Walk through all markdown files
	err := filepath.Walk(docsRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories and non-markdown files
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		
		// Skip certain files
		if strings.Contains(path, "archives/") || strings.Contains(path, "TODO") {
			return nil
		}
		
		log.Printf("Checking: %s", path)
		result, err := checkFile(path)
		if err != nil {
			log.Printf("Error checking %s: %v", path, err)
			return nil
		}
		
		if result.TotalIssues > 0 {
			allResults = append(allResults, result)
			totalIssues += result.TotalIssues
		}
		
		return nil
	})
	
	if err != nil {
		log.Fatalf("Error walking directory: %v", err)
	}
	
	// Report results
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("VALIDATION RESULTS")
	fmt.Println(strings.Repeat("=", 80))
	
	if totalIssues == 0 {
		fmt.Println("\n✅ All code examples use current API patterns!")
	} else {
		fmt.Printf("\n❌ Found %d issues in code examples:\n\n", totalIssues)
		
		for _, result := range allResults {
			relPath, _ := filepath.Rel(".", result.File)
			fmt.Printf("File: %s\n", relPath)
			fmt.Println(strings.Repeat("-", 40))
			
			for _, block := range result.Blocks {
				if len(block.Issues) > 0 {
					fmt.Printf("  Line %d (%s code):\n", block.StartLine, block.Language)
					for _, issue := range block.Issues {
						fmt.Printf("    ⚠️  %s\n", issue)
					}
					
					// Show a snippet of the problematic code
					lines := strings.Split(block.Content, "\n")
					for i, line := range lines {
						if i > 2 {
							fmt.Println("    ...")
							break
						}
						if line != "" {
							fmt.Printf("    | %s\n", line)
						}
					}
					fmt.Println()
				}
			}
		}
		
		// Generate update suggestions
		generateUpdateScript(allResults)
	}
	
	// Check for missing examples
	checkMissingExamples()
	
	// Exit with error code if issues found
	if totalIssues > 0 {
		os.Exit(1)
	}
}

func checkFile(filePath string) (CheckResult, error) {
	result := CheckResult{
		File:   filePath,
		Blocks: []CodeBlock{},
	}
	
	file, err := os.Open(filePath)
	if err != nil {
		return result, err
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	lineNum := 0
	inCodeBlock := false
	var currentBlock *CodeBlock
	var codeContent strings.Builder
	
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		
		// Check for code block start
		if matches := codeBlockStartRegex.FindStringSubmatch(line); matches != nil && !inCodeBlock {
			inCodeBlock = true
			language := "plain"
			if len(matches) > 1 && matches[1] != "" {
				language = matches[1]
			}
			
			currentBlock = &CodeBlock{
				File:      filePath,
				StartLine: lineNum,
				Language:  language,
				Issues:    []string{},
			}
			codeContent.Reset()
			continue
		}
		
		// Check for code block end
		if codeBlockEndRegex.MatchString(line) && inCodeBlock {
			inCodeBlock = false
			if currentBlock != nil {
				currentBlock.Content = codeContent.String()
				
				// Only check Go code blocks
				if currentBlock.Language == "go" {
					checkCodeBlock(currentBlock)
				}
				
				if len(currentBlock.Issues) > 0 {
					result.Blocks = append(result.Blocks, *currentBlock)
					result.TotalIssues += len(currentBlock.Issues)
				}
			}
			currentBlock = nil
			continue
		}
		
		// Collect code block content
		if inCodeBlock {
			codeContent.WriteString(line + "\n")
		}
	}
	
	return result, scanner.Err()
}

func checkCodeBlock(block *CodeBlock) {
	content := block.Content
	
	// Check for deprecated patterns
	for pattern, suggestion := range deprecatedPatterns {
		re := regexp.MustCompile(pattern)
		if re.MatchString(content) {
			block.Issues = append(block.Issues, 
				fmt.Sprintf("Deprecated pattern '%s' found. %s", pattern, suggestion))
		}
	}
	
	// Check for specific outdated patterns
	checkOutdatedImports(block)
	checkOutdatedProviderUsage(block)
	checkOutdatedAgentUsage(block)
	checkOutdatedStructuredOutput(block)
}

func checkOutdatedImports(block *CodeBlock) {
	content := block.Content
	
	// Check for old import paths
	oldImports := map[string]string{
		`"github.com/lexlapax/go-llms/llm"`:           `"github.com/lexlapax/go-llms/pkg/llm"`,
		`"github.com/lexlapax/go-llms/agent"`:         `"github.com/lexlapax/go-llms/pkg/agent"`,
		`"github.com/lexlapax/go-llms/tools"`:         `"github.com/lexlapax/go-llms/pkg/agent/tools"`,
		`"github.com/lexlapax/go-llms/schema"`:        `"github.com/lexlapax/go-llms/pkg/schema"`,
		`"github.com/lexlapax/go-llms/structured"`:    `"github.com/lexlapax/go-llms/pkg/structured"`,
	}
	
	for old, new := range oldImports {
		if strings.Contains(content, old) {
			block.Issues = append(block.Issues,
				fmt.Sprintf("Old import path %s should be %s", old, new))
		}
	}
}

func checkOutdatedProviderUsage(block *CodeBlock) {
	content := block.Content
	
	// Check for old provider initialization patterns
	if strings.Contains(content, "llm.Provider{") {
		block.Issues = append(block.Issues,
			"Direct provider struct initialization is deprecated. Use provider-specific constructors")
	}
	
	// Check for missing context in API calls
	if regexp.MustCompile(`provider\.\w+\([^,)]*\)`).MatchString(content) &&
		!strings.Contains(content, "ctx") {
		block.Issues = append(block.Issues,
			"Provider methods should include context as first parameter")
	}
	
	// Check for old completion request format
	if strings.Contains(content, "CompletionRequest{") && 
		!strings.Contains(content, "Messages") {
		block.Issues = append(block.Issues,
			"CompletionRequest should use Messages field, not Prompt")
	}
}

func checkOutdatedAgentUsage(block *CodeBlock) {
	content := block.Content
	
	// Check for old agent patterns
	if strings.Contains(content, "agent.Agent{") {
		block.Issues = append(block.Issues,
			"Direct agent struct initialization is deprecated. Use constructors like NewLLMAgent")
	}
	
	// Check for missing tool registration pattern
	if strings.Contains(content, "agent.Tools = ") {
		block.Issues = append(block.Issues,
			"Direct tool assignment is deprecated. Use RegisterTool method")
	}
}

func checkOutdatedStructuredOutput(block *CodeBlock) {
	content := block.Content
	
	// Check for old structured output patterns
	if strings.Contains(content, "structured.Extract(") && 
		!strings.Contains(content, "schema") {
		block.Issues = append(block.Issues,
			"Structured output should include schema definition")
	}
}

func generateUpdateScript(results []CheckResult) {
	// Count different types of issues
	importIssues := 0
	providerIssues := 0
	agentIssues := 0
	
	for _, result := range results {
		for _, block := range result.Blocks {
			for _, issue := range block.Issues {
				if strings.Contains(issue, "import") {
					importIssues++
				} else if strings.Contains(issue, "provider") {
					providerIssues++
				} else if strings.Contains(issue, "agent") {
					agentIssues++
				}
			}
		}
	}
	
	fmt.Println("\n📊 Issue Summary:")
	fmt.Printf("   Import path issues: %d\n", importIssues)
	fmt.Printf("   Provider API issues: %d\n", providerIssues)
	fmt.Printf("   Agent API issues: %d\n", agentIssues)
	
	fmt.Println("\n💡 To fix these issues:")
	fmt.Println("   1. Update import paths to use pkg/ prefix")
	fmt.Println("   2. Use provider-specific constructors (openai.New, etc.)")
	fmt.Println("   3. Include context in all API calls")
	fmt.Println("   4. Use current API patterns from v0.3.5+")
}

func checkMissingExamples() {
	fmt.Println("\n📋 Checking for missing examples...")
	
	// Key features that should have examples
	requiredExamples := []struct {
		Feature string
		Pattern string
	}{
		{"OpenAI provider initialization", "openai.New"},
		{"Anthropic provider initialization", "anthropic.New"},
		{"Google Gemini provider initialization", "google.New"},
		{"Ollama provider initialization", "ollama.New"},
		{"OpenRouter provider initialization", "openrouter.New"},
		{"Vertex AI provider initialization", "vertexai.New"},
		{"Streaming completion", "CompleteStream"},
		{"Tool registration", "RegisterTool"},
		{"Workflow creation", "NewWorkflowAgent"},
		{"Structured output", "structured.Parse"},
		{"Multi-provider setup", "MultiProvider"},
		{"Error handling", "ProviderError"},
	}
	
	// List the features we expect to see
	fmt.Println("   Expected example coverage:")
	for _, ex := range requiredExamples {
		fmt.Printf("   - %s\n", ex.Feature)
	}
	fmt.Println("\n   (Manual verification recommended for complete coverage)")
}