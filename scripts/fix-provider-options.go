package main

// ABOUTME: Script to fix remaining provider option patterns in documentation
// ABOUTME: Converts inline options to use domain option constructors

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	log.Println("Go-LLMs Provider Options Fixer")
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
	
	// Fix provider option patterns
	updatedContent = fixProviderOptions(updatedContent, &fixCount)
	
	// Only write if changes were made
	if fixCount > 0 {
		err = os.WriteFile(filePath, []byte(updatedContent), 0644)
		if err != nil {
			return 0, err
		}
	}
	
	return fixCount, nil
}

func fixProviderOptions(content string, fixCount *int) string {
	lines := strings.Split(content, "\n")
	var updatedLines []string
	inCodeBlock := false
	codeBlockLang := ""
	var codeBlockLines []string
	codeBlockStart := 0
	
	for i, line := range lines {
		// Track code blocks
		if strings.HasPrefix(line, "```") {
			if !inCodeBlock {
				inCodeBlock = true
				codeBlockStart = i
				codeBlockLines = []string{}
				if len(line) > 3 {
					codeBlockLang = strings.TrimSpace(line[3:])
				}
			} else {
				// End of code block - process it
				if codeBlockLang == "go" {
					processedBlock := processGoCodeBlock(codeBlockLines, fixCount)
					updatedLines = append(updatedLines, lines[codeBlockStart])
					updatedLines = append(updatedLines, processedBlock...)
				} else {
					// Not a Go block, keep original
					updatedLines = append(updatedLines, lines[codeBlockStart])
					updatedLines = append(updatedLines, codeBlockLines...)
				}
				inCodeBlock = false
				codeBlockLang = ""
				codeBlockLines = nil
			}
		} else if inCodeBlock {
			codeBlockLines = append(codeBlockLines, line)
			continue
		}
		
		if !inCodeBlock {
			updatedLines = append(updatedLines, line)
		}
	}
	
	// Handle case where file ends while still in code block
	if inCodeBlock && len(codeBlockLines) > 0 {
		updatedLines = append(updatedLines, lines[codeBlockStart])
		updatedLines = append(updatedLines, codeBlockLines...)
	}
	
	return strings.Join(updatedLines, "\n")
}

func processGoCodeBlock(lines []string, fixCount *int) []string {
	var result []string
	providerCallRe := regexp.MustCompile(`provider\.New(\w+)Provider\(`)
	
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		
		// Check if this line starts a provider initialization
		if providerCallRe.MatchString(line) {
			// Find the complete provider initialization block
			blockLines, endIdx := extractProviderBlock(lines, i)
			if endIdx > i {
				// Process the block
				fixedBlock := fixProviderBlock(blockLines, fixCount)
				result = append(result, fixedBlock...)
				i = endIdx // Skip processed lines
				continue
			}
		}
		
		result = append(result, line)
	}
	
	return result
}

func extractProviderBlock(lines []string, startIdx int) ([]string, int) {
	var block []string
	openParens := 0
	inBlock := false
	
	for i := startIdx; i < len(lines); i++ {
		line := lines[i]
		block = append(block, line)
		
		// Count parentheses
		for _, ch := range line {
			if ch == '(' {
				openParens++
				inBlock = true
			} else if ch == ')' {
				openParens--
			}
		}
		
		// Check if we've closed all parentheses
		if inBlock && openParens == 0 {
			// Check if next line is just a closing brace or paren
			if i+1 < len(lines) && (strings.TrimSpace(lines[i+1]) == "}" || strings.TrimSpace(lines[i+1]) == ")") {
				block = append(block, lines[i+1])
				return block, i + 1
			}
			return block, i
		}
	}
	
	return block, len(lines) - 1
}

func fixProviderBlock(lines []string, fixCount *int) []string {
	// Check if this block has inline options
	hasInlineOptions := false
	for _, line := range lines {
		if regexp.MustCompile(`^\s*(Organization|BaseURL|Timeout|SystemPrompt|Temperature|APIVersion|SiteURL|SiteName):\s*`).MatchString(line) {
			hasInlineOptions = true
			break
		}
	}
	
	if !hasInlineOptions {
		return lines
	}
	
	// Extract provider type and base parameters
	firstLine := lines[0]
	providerMatch := regexp.MustCompile(`provider\.New(\w+)Provider\(`).FindStringSubmatch(firstLine)
	if len(providerMatch) < 2 {
		return lines
	}
	
	providerType := providerMatch[1]
	
	// Parse options
	var baseParams []string
	var options []string
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		
		// Check for inline options
		optionMatch := regexp.MustCompile(`^(Organization|BaseURL|Timeout|SystemPrompt|Temperature|APIVersion|SiteURL|SiteName):\s*(.+),?\s*$`).FindStringSubmatch(line)
		if len(optionMatch) >= 3 {
			optName := optionMatch[1]
			optValue := strings.TrimSuffix(strings.TrimSpace(optionMatch[2]), ",")
			
			// Map to correct option constructor
			optionConstructor := getOptionConstructor(providerType, optName)
			if optionConstructor != "" {
				options = append(options, fmt.Sprintf("    %s(%s),", optionConstructor, optValue))
			}
			*fixCount++
		} else if strings.Contains(line, "os.Getenv") && !strings.Contains(line, ":") {
			// This is likely a base parameter
			baseParams = append(baseParams, strings.TrimSuffix(strings.TrimSpace(line), ","))
		}
	}
	
	// Reconstruct the provider initialization
	var result []string
	
	// First line with provider call and base params
	if len(baseParams) > 0 {
		result = append(result, fmt.Sprintf("provider := provider.New%sProvider(%s,", providerType, strings.Join(baseParams, ", ")))
	} else {
		result = append(result, fmt.Sprintf("provider := provider.New%sProvider(", providerType))
	}
	
	// Add options
	for _, opt := range options {
		result = append(result, opt)
	}
	
	// Close the call
	result = append(result, ")")
	
	return result
}

func getOptionConstructor(providerType, optionName string) string {
	// Map option names to their constructors
	optionMap := map[string]map[string]string{
		"OpenAI": {
			"Organization": "domain.NewOpenAIOrganizationOption",
			"BaseURL":      "domain.NewBaseURLOption",
		},
		"Anthropic": {
			"SystemPrompt": "domain.NewAnthropicSystemPromptOption",
			"BaseURL":      "domain.NewBaseURLOption",
			"APIVersion":   "domain.NewAnthropicAPIVersionOption",
		},
		"Gemini": {
			"BaseURL": "domain.NewBaseURLOption",
		},
		"Ollama": {
			"BaseURL": "domain.NewBaseURLOption",
			"Timeout": "domain.NewTimeoutOption",
		},
		"OpenRouter": {
			"BaseURL":  "domain.NewBaseURLOption",
			"SiteURL":  "domain.NewOpenRouterSiteURLOption",
			"SiteName": "domain.NewOpenRouterSiteNameOption",
		},
		"VertexAI": {
			"BaseURL": "domain.NewBaseURLOption",
		},
	}
	
	if providerOpts, ok := optionMap[providerType]; ok {
		if constructor, ok := providerOpts[optionName]; ok {
			return constructor
		}
	}
	
	// Default constructors for common options
	switch optionName {
	case "BaseURL":
		return "domain.NewBaseURLOption"
	case "Timeout":
		return "domain.NewTimeoutOption"
	case "Temperature":
		return "domain.NewTemperatureOption"
	}
	
	return ""
}