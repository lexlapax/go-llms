package main

// ABOUTME: Enhanced script to fix outdated code examples in documentation
// ABOUTME: Updates provider, tool, and schema patterns to match v0.3.5+ API

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
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
	
	// Tool patterns
	`tools.NewFileReadTool()`:      `tools.Tools.Get("file_read")`,
	`tools.NewFileWriteTool()`:     `tools.Tools.Get("file_write")`,
	`tools.NewHTTPTool()`:          `tools.Tools.Get("http_request")`,
	`tools.NewShellExecTool()`:     `tools.Tools.Get("shell_exec")`,
	`tools.NewDataProcessTool()`:   `tools.Tools.Get("data_process")`,
	`tools.NewSystemInfoTool()`:    `tools.Tools.Get("system_info")`,
	
	// Schema type references
	`*domain.Schema`:  `*sdomain.Schema`,
	`domain.Schema`:   `sdomain.Schema`,
	
	// Tool registry access
	`tools.Registry`:  `tools.Tools`,
	`Registry.Get`:    `Tools.Get`,
	
	// Agent creation pattern fixes
	`) // Updated initialization`: `}`,
}

// Regular expressions for more complex patterns
var (
	// Tool parameter struct pattern: file.ReadFileParams{...} -> map[string]interface{}{...}
	toolParamStructRe = regexp.MustCompile(`(\w+)\.(\w+Params)\{`)
	
	// Provider option inline pattern
	providerOptionRe = regexp.MustCompile(`(Organization|BaseURL|APIKey|SystemPrompt|Temperature):\s*("[^"]*"|[^,}]+)`)
	
	// Schema interface pattern
	schemaInterfaceRe = regexp.MustCompile(`(ParameterSchema|OutputSchema)\(\)\s*\*domain\.Schema`)
)

// Tool parameter type mappings
var toolParamTypes = map[string]bool{
	"ReadFileParams":     true,
	"WriteFileParams":    true,
	"HTTPRequestParams":  true,
	"ShellExecParams":    true,
	"DataProcessParams":  true,
	"SystemInfoParams":   true,
}

func main() {
	log.Println("Go-LLMs Documentation Example Fixer (Enhanced)")
	log.Println("=============================================")
	
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
	
	// Apply simple replacements
	for old, new := range replacements {
		beforeCount := strings.Count(updatedContent, old)
		if beforeCount > 0 {
			updatedContent = strings.ReplaceAll(updatedContent, old, new)
			fixCount += beforeCount
		}
	}
	
	// Apply complex pattern fixes
	updatedContent = fixComplexPatterns(updatedContent, &fixCount)
	
	// Fix specific patterns that need more complex replacements
	updatedContent = fixProviderExamples(updatedContent, &fixCount)
	
	// Fix tool examples
	updatedContent = fixToolExamples(updatedContent, &fixCount)
	
	// Fix agent initialization syntax errors
	updatedContent = fixAgentSyntaxErrors(updatedContent, &fixCount)
	
	// Only write if changes were made
	if fixCount > 0 {
		err = os.WriteFile(filePath, []byte(updatedContent), 0644)
		if err != nil {
			return 0, err
		}
	}
	
	return fixCount, nil
}

func fixComplexPatterns(content string, fixCount *int) string {
	// Fix tool parameter struct patterns
	matches := toolParamStructRe.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) >= 3 && toolParamTypes[match[2]] {
			old := match[0]
			new := "map[string]interface{}{"
			content = strings.Replace(content, old, new, -1)
			*fixCount++
		}
	}
	
	// Fix schema interface patterns
	content = schemaInterfaceRe.ReplaceAllString(content, "${1}() *sdomain.Schema")
	
	return content
}

func fixProviderExamples(content string, fixCount *int) string {
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
			// Fix provider option patterns
			line = fixProviderOptions(line, fixCount)
			
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

func fixProviderOptions(line string, fixCount *int) string {
	// Check if this line has provider initialization with inline options
	if !strings.Contains(line, "Provider(") || !strings.Contains(line, ":") {
		return line
	}
	
	// Map of option names to their constructor functions
	optionConstructors := map[string]string{
		"Organization": "domain.NewOpenAIOrganizationOption",
		"BaseURL":      "domain.NewBaseURLOption",
		"SystemPrompt": "domain.NewAnthropicSystemPromptOption",
		"Temperature":  "domain.NewTemperatureOption",
	}
	
	// Find and replace inline options
	for optName, constructor := range optionConstructors {
		pattern := fmt.Sprintf(`%s:\s*("[^"]*"|[^,}]+)`, optName)
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(line); len(matches) > 1 {
			value := strings.TrimSpace(matches[1])
			replacement := fmt.Sprintf("%s(%s)", constructor, value)
			line = re.ReplaceAllString(line, replacement)
			*fixCount++
		}
	}
	
	return line
}

func fixToolExamples(content string, fixCount *int) string {
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
			// Fix tool instantiation patterns
			if strings.Contains(line, "tool := tools.New") && strings.Contains(line, "Tool()") {
				// Extract tool name
				re := regexp.MustCompile(`tools\.New(\w+)Tool\(\)`)
				if matches := re.FindStringSubmatch(line); len(matches) > 1 {
					toolName := strings.ToLower(matches[1])
					// Map to actual tool names in registry
					toolNameMap := map[string]string{
						"fileread":    "file_read",
						"filewrite":   "file_write",
						"http":        "http_request",
						"shellexec":   "shell_exec",
						"dataprocess": "data_process",
						"systeminfo":  "system_info",
					}
					
					if registryName, ok := toolNameMap[toolName]; ok {
						newLine := strings.Replace(line, matches[0], fmt.Sprintf(`tools.Tools.Get("%s")`, registryName), 1)
						if strings.Contains(line, ":=") {
							// Add error handling
							newLine = strings.Replace(newLine, "tool :=", "tool, err :=", 1)
						}
						line = newLine
						*fixCount++
					}
				}
			}
			
			// Fix MustRegisterTool patterns
			if strings.Contains(line, "tools.RegisterTool(") {
				line = strings.Replace(line, "tools.RegisterTool(", "tools.MustRegisterTool(", 1)
				*fixCount++
			}
		}
		
		updatedLines = append(updatedLines, line)
	}
	
	return strings.Join(updatedLines, "\n")
}

func fixAgentSyntaxErrors(content string, fixCount *int) string {
	lines := strings.Split(content, "\n")
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
		}
		
		// Fix agent initialization syntax errors
		if inCodeBlock && codeBlockLang == "go" {
			// Remove extra closing parenthesis with comment
			if strings.Contains(line, ") // Updated initialization") {
				line = strings.Replace(line, ") // Updated initialization", "}", 1)
				*fixCount++
			}
			
			// Fix double closing parentheses
			if strings.Contains(line, "))") && strings.Contains(line, "Provider(") {
				// Check if next line is just a single )
				if i+1 < len(lines) && strings.TrimSpace(lines[i+1]) == ")" {
					line = strings.Replace(line, "))", ")", 1)
					*fixCount++
				}
			}
		}
		
		updatedLines = append(updatedLines, line)
	}
	
	return strings.Join(updatedLines, "\n")
}