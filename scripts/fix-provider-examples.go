package main

// ABOUTME: Script to fix provider initialization examples to match current API
// ABOUTME: Updates to use correct constructor signatures with apiKey, model parameters

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	log.Println("Go-LLMs Provider Example Fixer")
	log.Println("==============================")
	
	totalFixed := 0
	filesProcessed := 0
	
	// Walk through all markdown files
	err := filepath.Walk("docs", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories and non-markdown files
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		
		// Skip archive files
		if strings.Contains(path, "archives/") {
			return nil
		}
		
		fixed, err := fixProviderExamples(path)
		if err != nil {
			log.Printf("Error fixing %s: %v", path, err)
			return nil
		}
		
		if fixed > 0 {
			log.Printf("Fixed %d patterns in %s", fixed, path)
			totalFixed += fixed
			filesProcessed++
		}
		
		return nil
	})
	
	if err != nil {
		log.Fatalf("Error walking directory: %v", err)
	}
	
	log.Printf("\n✅ Fixed %d patterns across %d files", totalFixed, filesProcessed)
}

func fixProviderExamples(filePath string) (int, error) {
	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0, err
	}
	
	originalContent := string(content)
	fixCount := 0
	
	// Process the content line by line to handle code blocks properly
	lines := strings.Split(originalContent, "\n")
	var updatedLines []string
	inCodeBlock := false
	codeBlockLang := ""
	
	for i, line := range lines {
		// Track code blocks
		if strings.HasPrefix(line, "```") {
			if !inCodeBlock {
				inCodeBlock = true
				if len(line) > 3 {
					codeBlockLang = strings.TrimSpace(line[3:])
				}
			} else {
				inCodeBlock = false
				codeBlockLang = ""
			}
			updatedLines = append(updatedLines, line)
			continue
		}
		
		// Only process Go code blocks
		if inCodeBlock && codeBlockLang == "go" {
			// Fix provider initialization patterns
			fixedLine, fixes := fixProviderLine(line, i < len(lines)-1 && i < len(lines)-2)
			fixCount += fixes
			updatedLines = append(updatedLines, fixedLine)
		} else {
			updatedLines = append(updatedLines, line)
		}
	}
	
	// Only write if changes were made
	if fixCount > 0 {
		updatedContent := strings.Join(updatedLines, "\n")
		err = os.WriteFile(filePath, []byte(updatedContent), 0644)
		if err != nil {
			return 0, err
		}
	}
	
	return fixCount, nil
}

func fixProviderLine(line string, hasMoreLines bool) (string, int) {
	fixes := 0
	
	// Pattern 1: Fix provider.NewXXXProvider(domain.XXXOptions{...})
	providerPatterns := []struct {
		pattern string
		replace string
	}{
		{
			pattern: `provider\.NewOpenAIProvider\(domain\.OpenAIOptions\{`,
			replace: `provider.NewOpenAIProvider(os.Getenv("OPENAI_API_KEY"), "gpt-4"`,
		},
		{
			pattern: `provider\.NewAnthropicProvider\(domain\.AnthropicOptions\{`,
			replace: `provider.NewAnthropicProvider(os.Getenv("ANTHROPIC_API_KEY"), "claude-3-opus-20240229"`,
		},
		{
			pattern: `provider\.NewGeminiProvider\(domain\.GeminiOptions\{`,
			replace: `provider.NewGeminiProvider(os.Getenv("GOOGLE_API_KEY"), "gemini-pro"`,
		},
		{
			pattern: `provider\.NewOllamaProvider\(domain\.OllamaOptions\{`,
			replace: `provider.NewOllamaProvider("", "llama2"`,
		},
		{
			pattern: `provider\.NewVertexAIProvider\(domain\.VertexAIOptions\{`,
			replace: `provider.NewVertexAIProvider("my-project", "us-central1", "gemini-pro"`,
		},
		{
			pattern: `provider\.NewOpenRouterProvider\(domain\.OpenRouterOptions\{`,
			replace: `provider.NewOpenRouterProvider(os.Getenv("OPENROUTER_API_KEY"), "openai/gpt-3.5-turbo"`,
		},
	}
	
	for _, p := range providerPatterns {
		re := regexp.MustCompile(p.pattern)
		if re.MatchString(line) {
			// Check if this is a multi-line initialization
			if strings.TrimSpace(line) == strings.TrimSpace(re.FindString(line)) {
				// This line only contains the start of initialization
				line = p.replace + ", // Options can be added here"
			} else {
				// Replace just the pattern part
				line = re.ReplaceAllString(line, p.replace+", ")
			}
			fixes++
		}
	}
	
	// Pattern 2: Fix incorrect option structs on their own
	if strings.Contains(line, "APIKey:") && strings.Contains(line, "os.Getenv") {
		// This is likely an options line that needs to be commented or removed
		if !strings.Contains(line, "//") {
			line = "    // " + strings.TrimSpace(line) + " // Moved to constructor parameters"
			fixes++
		}
	}
	
	// Pattern 3: Fix closing braces for old option structs
	if strings.TrimSpace(line) == "})" && hasMoreLines {
		line = ") // Updated initialization"
		fixes++
	}
	
	return line, fixes
}

func cleanupExamples() {
	// Additional cleanup patterns that might be needed
	patterns := []struct {
		desc string
		old  string
		new  string
	}{
		{
			desc: "Import statements",
			old:  `"github.com/lexlapax/go-llms/pkg/llm/provider/openai"`,
			new:  `"github.com/lexlapax/go-llms/pkg/llm/provider"`,
		},
		{
			desc: "Domain imports",
			old:  `domain "github.com/lexlapax/go-llms/pkg/llm/domain"`,
			new:  `"github.com/lexlapax/go-llms/pkg/llm/domain"`,
		},
	}
	
	fmt.Println("\nAdditional cleanup patterns to consider:")
	for _, p := range patterns {
		fmt.Printf("- %s: Replace '%s' with '%s'\n", p.desc, p.old, p.new)
	}
}