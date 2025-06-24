package main

// ABOUTME: Script to check documentation completeness across go-llms project
// ABOUTME: Validates ABOUTME comments, learning paths, and prerequisites

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type DocIssue struct {
	File    string
	Line    int
	Type    string
	Message string
}

type DocStats struct {
	TotalMarkdownFiles   int
	TotalGoFiles         int
	FilesWithIssues      int
	TotalIssues          int
	MissingABOUTME       int
	InvalidABOUTME       int
	MissingPrerequisites int
	BrokenLinks          int
	IncompleteExamples   int
}

var (
	aboutMeRe        = regexp.MustCompile(`^// ABOUTME: .+$`)
	prerequisiteRe   = regexp.MustCompile(`(?i)(prerequisite|requirement|before|prior|knowledge|expected)`)
	linkRe           = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	codeBlockRe      = regexp.MustCompile("```[a-zA-Z]*")
	learningPathRe   = regexp.MustCompile(`(?i)(next|previous|continue|proceed|step|guide|tutorial)`)
	advancedTopicRe  = regexp.MustCompile(`(?i)(advanced|complex|expert|deep.dive|in.depth)`)
)

func main() {
	log.Println("Go-LLMs Documentation Completeness Check")
	log.Println("=======================================")
	
	stats := &DocStats{}
	issues := []DocIssue{}
	
	// Check Go files for ABOUTME comments
	log.Println("\nChecking Go files for ABOUTME comments...")
	goIssues := checkGoFiles(stats)
	issues = append(issues, goIssues...)
	
	// Check markdown documentation
	log.Println("\nChecking markdown documentation...")
	mdIssues := checkMarkdownDocs(stats)
	issues = append(issues, mdIssues...)
	
	// Print summary
	printSummary(stats, issues)
	
	// Generate report
	generateReport(stats, issues)
}

func checkGoFiles(stats *DocStats) []DocIssue {
	var issues []DocIssue
	
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip non-Go files
		if !strings.HasSuffix(path, ".go") || info.IsDir() {
			return nil
		}
		
		// Skip vendor, test files, and generated files
		if strings.Contains(path, "vendor/") || 
		   strings.Contains(path, ".pb.go") ||
		   strings.Contains(path, "_test.go") ||
		   strings.Contains(path, "scripts/") {
			return nil
		}
		
		// Check only pkg/ files for ABOUTME
		if !strings.Contains(path, "pkg/") {
			return nil
		}
		
		stats.TotalGoFiles++
		fileIssues := checkGoFile(path, stats)
		if len(fileIssues) > 0 {
			stats.FilesWithIssues++
			issues = append(issues, fileIssues...)
		}
		
		return nil
	})
	
	if err != nil {
		log.Printf("Error walking Go files: %v", err)
	}
	
	return issues
}

func checkGoFile(filePath string, stats *DocStats) []DocIssue {
	var issues []DocIssue
	
	content, err := os.ReadFile(filePath)
	if err != nil {
		return issues
	}
	
	lines := strings.Split(string(content), "\n")
	
	// Check for ABOUTME comments
	aboutMeCount := 0
	aboutMeStartLine := -1
	validABOUTME := true
	
	for i, line := range lines {
		if aboutMeRe.MatchString(line) {
			if aboutMeCount == 0 {
				aboutMeStartLine = i + 1
			}
			aboutMeCount++
			
			// Check line length
			if len(line) > 80 {
				validABOUTME = false
				issues = append(issues, DocIssue{
					File:    filePath,
					Line:    i + 1,
					Type:    "ABOUTME_TOO_LONG",
					Message: fmt.Sprintf("ABOUTME line exceeds 80 characters (%d chars)", len(line)),
				})
			}
			
			// Check content quality
			content := strings.TrimPrefix(line, "// ABOUTME: ")
			if len(content) < 20 {
				validABOUTME = false
				issues = append(issues, DocIssue{
					File:    filePath,
					Line:    i + 1,
					Type:    "ABOUTME_TOO_SHORT",
					Message: "ABOUTME content is too brief (less than 20 chars)",
				})
			}
		}
		
		// Break after package declaration area
		if strings.HasPrefix(line, "package ") && i > 10 {
			break
		}
	}
	
	// Validate ABOUTME presence and format
	if aboutMeCount == 0 {
		stats.MissingABOUTME++
		issues = append(issues, DocIssue{
			File:    filePath,
			Line:    1,
			Type:    "MISSING_ABOUTME",
			Message: "File missing required ABOUTME comments",
		})
	} else if aboutMeCount != 2 {
		stats.InvalidABOUTME++
		issues = append(issues, DocIssue{
			File:    filePath,
			Line:    aboutMeStartLine,
			Type:    "INVALID_ABOUTME",
			Message: fmt.Sprintf("ABOUTME must have exactly 2 lines, found %d", aboutMeCount),
		})
	} else if !validABOUTME {
		stats.InvalidABOUTME++
	}
	
	stats.TotalIssues += len(issues)
	return issues
}

func checkMarkdownDocs(stats *DocStats) []DocIssue {
	var issues []DocIssue
	
	// Define documentation paths
	docPaths := []string{
		"docs/user-guide",
		"docs/technical",
		"docs/api",
	}
	
	for _, basePath := range docPaths {
		err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			
			if !strings.HasSuffix(path, ".md") || info.IsDir() {
				return nil
			}
			
			// Skip archives
			if strings.Contains(path, "archives/") {
				return nil
			}
			
			stats.TotalMarkdownFiles++
			fileIssues := checkMarkdownFile(path, stats)
			if len(fileIssues) > 0 {
				stats.FilesWithIssues++
				issues = append(issues, fileIssues...)
			}
			
			return nil
		})
		
		if err != nil {
			log.Printf("Error walking %s: %v", basePath, err)
		}
	}
	
	return issues
}

func checkMarkdownFile(filePath string, stats *DocStats) []DocIssue {
	var issues []DocIssue
	
	content, err := os.ReadFile(filePath)
	if err != nil {
		return issues
	}
	
	lines := strings.Split(string(content), "\n")
	
	// Check for advanced topics without prerequisites
	if strings.Contains(filePath, "advanced/") || advancedTopicRe.MatchString(filePath) {
		hasPrerequisites := false
		for _, line := range lines {
			if prerequisiteRe.MatchString(line) {
				hasPrerequisites = true
				break
			}
		}
		
		if !hasPrerequisites {
			stats.MissingPrerequisites++
			issues = append(issues, DocIssue{
				File:    filePath,
				Line:    1,
				Type:    "MISSING_PREREQUISITES",
				Message: "Advanced topic missing prerequisites section",
			})
		}
	}
	
	// Check learning paths navigation
	if strings.Contains(filePath, "getting-started/") || 
	   strings.Contains(filePath, "guides/") || 
	   strings.Contains(filePath, "examples/") {
		hasNavigation := false
		for _, line := range lines {
			if learningPathRe.MatchString(line) || linkRe.MatchString(line) {
				hasNavigation = true
				break
			}
		}
		
		if !hasNavigation && !strings.Contains(filePath, "README.md") {
			issues = append(issues, DocIssue{
				File:    filePath,
				Line:    1,
				Type:    "MISSING_NAVIGATION",
				Message: "Learning path document missing navigation links",
			})
		}
	}
	
	// Check for broken internal links
	for i, line := range lines {
		matches := linkRe.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) >= 3 {
				// linkText := match[1] // Currently unused
				linkPath := match[2]
				
				// Check internal links
				if !strings.HasPrefix(linkPath, "http") && !strings.HasPrefix(linkPath, "#") {
					// Resolve relative path
					fullPath := filepath.Join(filepath.Dir(filePath), linkPath)
					fullPath = strings.TrimSuffix(fullPath, "#.*") // Remove anchors
					
					if _, err := os.Stat(fullPath); os.IsNotExist(err) {
						stats.BrokenLinks++
						issues = append(issues, DocIssue{
							File:    filePath,
							Line:    i + 1,
							Type:    "BROKEN_LINK",
							Message: fmt.Sprintf("Broken link to '%s'", linkPath),
						})
					}
				}
			}
		}
	}
	
	// Check code examples completeness
	inCodeBlock := false
	codeBlockStart := 0
	codeBlockContent := []string{}
	
	for i, line := range lines {
		if codeBlockRe.MatchString(line) {
			if !inCodeBlock {
				inCodeBlock = true
				codeBlockStart = i + 1
				codeBlockContent = []string{}
			} else {
				// End of code block - check completeness
				if len(codeBlockContent) > 0 {
					checkCodeExample(filePath, codeBlockStart, codeBlockContent, &issues, stats)
				}
				inCodeBlock = false
			}
		} else if inCodeBlock {
			codeBlockContent = append(codeBlockContent, line)
		}
	}
	
	stats.TotalIssues += len(issues)
	return issues
}

func checkCodeExample(filePath string, startLine int, content []string, issues *[]DocIssue, stats *DocStats) {
	// Join content for analysis
	code := strings.Join(content, "\n")
	
	// Check for incomplete Go examples
	if strings.Contains(code, "// ...") || strings.Contains(code, "/* ... */") {
		stats.IncompleteExamples++
		*issues = append(*issues, DocIssue{
			File:    filePath,
			Line:    startLine,
			Type:    "INCOMPLETE_EXAMPLE",
			Message: "Code example contains placeholder (...) indicating incomplete code",
		})
	}
	
	// Check for error handling in examples
	if strings.Contains(code, "err :=") && !strings.Contains(code, "if err") {
		*issues = append(*issues, DocIssue{
			File:    filePath,
			Line:    startLine,
			Type:    "MISSING_ERROR_HANDLING",
			Message: "Code example assigns error but doesn't check it",
		})
	}
}

func printSummary(stats *DocStats, issues []DocIssue) {
	fmt.Println("\n=== Documentation Completeness Summary ===")
	fmt.Printf("Total Go files checked: %d\n", stats.TotalGoFiles)
	fmt.Printf("Total Markdown files checked: %d\n", stats.TotalMarkdownFiles)
	fmt.Printf("Files with issues: %d\n", stats.FilesWithIssues)
	fmt.Printf("Total issues found: %d\n", stats.TotalIssues)
	fmt.Println("\nIssue Breakdown:")
	fmt.Printf("- Missing ABOUTME comments: %d\n", stats.MissingABOUTME)
	fmt.Printf("- Invalid ABOUTME format: %d\n", stats.InvalidABOUTME)
	fmt.Printf("- Missing prerequisites: %d\n", stats.MissingPrerequisites)
	fmt.Printf("- Broken internal links: %d\n", stats.BrokenLinks)
	fmt.Printf("- Incomplete code examples: %d\n", stats.IncompleteExamples)
	
	if len(issues) > 0 {
		fmt.Println("\nTop Issues (first 10):")
		for i, issue := range issues {
			if i >= 10 {
				break
			}
			fmt.Printf("%s:%d [%s] %s\n", issue.File, issue.Line, issue.Type, issue.Message)
		}
	}
}

func generateReport(stats *DocStats, issues []DocIssue) {
	report := fmt.Sprintf(`# Documentation Completeness Check Report

Generated: %s

## Summary

- **Total Go Files Checked**: %d
- **Total Markdown Files Checked**: %d
- **Files with Issues**: %d
- **Total Issues Found**: %d

## Issue Categories

| Category | Count |
|----------|-------|
| Missing ABOUTME Comments | %d |
| Invalid ABOUTME Format | %d |
| Missing Prerequisites | %d |
| Broken Internal Links | %d |
| Incomplete Code Examples | %d |

## Detailed Issues

`,
		"January 23, 2025",
		stats.TotalGoFiles,
		stats.TotalMarkdownFiles,
		stats.FilesWithIssues,
		stats.TotalIssues,
		stats.MissingABOUTME,
		stats.InvalidABOUTME,
		stats.MissingPrerequisites,
		stats.BrokenLinks,
		stats.IncompleteExamples,
	)
	
	// Group issues by type
	issuesByType := make(map[string][]DocIssue)
	for _, issue := range issues {
		issuesByType[issue.Type] = append(issuesByType[issue.Type], issue)
	}
	
	// Add issues to report
	for issueType, typeIssues := range issuesByType {
		report += fmt.Sprintf("### %s\n\n", issueType)
		for _, issue := range typeIssues {
			report += fmt.Sprintf("- `%s:%d` - %s\n", issue.File, issue.Line, issue.Message)
		}
		report += "\n"
	}
	
	// Write report
	err := os.WriteFile("docs/0.3.6.7.15-COMPLETENESS-REPORT.md", []byte(report), 0644)
	if err != nil {
		log.Printf("Error writing report: %v", err)
	} else {
		log.Println("\nReport written to docs/0.3.6.7.15-COMPLETENESS-REPORT.md")
	}
}