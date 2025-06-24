package main

// ABOUTME: Script to fix outdated code examples in documentation
// ABOUTME: Updates provider initialization patterns to match current API

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Replacement patterns for updating code examples
var replacements = map[string]string{
	// Provider initialization patterns
	`provider.NewOpenAI(`:      `provider.NewOpenAIProvider(`,
	`provider.NewAnthropic(`:   `provider.NewAnthropicProvider(`,
	`provider.NewGemini(`:      `provider.NewGeminiProvider(`,
	`provider.NewOllama(`:      `provider.NewOllamaProvider(`,
	`provider.NewOpenRouter(`:  `provider.NewOpenRouterProvider(`,
	`provider.NewVertexAI(`:    `provider.NewVertexAIProvider(`,
	
	// Option patterns
	`provider.OpenAIOptions{`:      `domain.OpenAIOptions{`,
	`provider.AnthropicOptions{`:   `domain.AnthropicOptions{`,
	`provider.GeminiOptions{`:      `domain.GeminiOptions{`,
	`provider.OllamaOptions{`:      `domain.OllamaOptions{`,
	`provider.OpenRouterOptions{`:  `domain.OpenRouterOptions{`,
	`provider.VertexAIOptions{`:    `domain.VertexAIOptions{`,
	
	// Import updates
	`"github.com/lexlapax/go-llms/llm"`:        `"github.com/lexlapax/go-llms/pkg/llm"`,
	`"github.com/lexlapax/go-llms/agent"`:      `"github.com/lexlapax/go-llms/pkg/agent"`,
	`"github.com/lexlapax/go-llms/tools"`:      `"github.com/lexlapax/go-llms/pkg/agent/tools"`,
	`"github.com/lexlapax/go-llms/schema"`:     `"github.com/lexlapax/go-llms/pkg/schema"`,
	`"github.com/lexlapax/go-llms/structured"`: `"github.com/lexlapax/go-llms/pkg/structured"`,
}

func main() {
	log.Println("Go-LLMs Documentation Example Fixer")
	log.Println("===================================")
	
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
		
		fixed, err := fixFile(path)
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

func fixFile(filePath string) (int, error) {
	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0, err
	}
	
	originalContent := string(content)
	updatedContent := originalContent
	fixCount := 0
	
	// Apply replacements
	for old, new := range replacements {
		beforeCount := strings.Count(updatedContent, old)
		if beforeCount > 0 {
			updatedContent = strings.ReplaceAll(updatedContent, old, new)
			fixCount += beforeCount
		}
	}
	
	// Fix specific patterns that need more complex replacements
	updatedContent = fixProviderExamples(updatedContent, &fixCount)
	
	// Only write if changes were made
	if fixCount > 0 {
		err = os.WriteFile(filePath, []byte(updatedContent), 0644)
		if err != nil {
			return 0, err
		}
	}
	
	return fixCount, nil
}

func fixProviderExamples(content string, fixCount *int) string {
	// Fix incorrect provider initialization examples
	lines := strings.Split(content, "\n")
	var updatedLines []string
	inCodeBlock := false
	codeBlockLang := ""
	
	for _, line := range lines {
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
		}
		
		// Only process Go code blocks
		if inCodeBlock && codeBlockLang == "go" {
			// Fix openai.New pattern
			if strings.Contains(line, "openai.New(") && strings.Contains(line, "openai.Config{") {
				oldLine := line
				line = strings.Replace(line, "openai.New(", "provider.NewOpenAIProvider(", 1)
				line = strings.Replace(line, "openai.Config{", "", 1)
				line = strings.Replace(line, "APIKey:", `"api-key", "model-name", domain.WithOpenAIOption(`, 1)
				if line != oldLine {
					*fixCount++
				}
			}
			
			// Fix similar patterns for other providers
			providerPatterns := []struct {
				old    string
				new    string
				config string
			}{
				{"anthropic.New(", "provider.NewAnthropicProvider(", "anthropic.Config{"},
				{"google.New(", "provider.NewGeminiProvider(", "google.Config{"},
				{"ollama.New(", "provider.NewOllamaProvider(", "ollama.Config{"},
				{"vertexai.New(", "provider.NewVertexAIProvider(", "vertexai.Config{"},
				{"openrouter.New(", "provider.NewOpenRouterProvider(", "openrouter.Config{"},
			}
			
			for _, pattern := range providerPatterns {
				if strings.Contains(line, pattern.old) && strings.Contains(line, pattern.config) {
					oldLine := line
					line = strings.Replace(line, pattern.old, pattern.new, 1)
					line = strings.Replace(line, pattern.config, "", 1)
					if line != oldLine {
						*fixCount++
					}
				}
			}
		}
		
		updatedLines = append(updatedLines, line)
	}
	
	return strings.Join(updatedLines, "\n")
}