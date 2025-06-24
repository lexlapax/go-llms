package main

// ABOUTME: Script to validate all internal links in documentation files
// ABOUTME: Checks that referenced files exist and reports broken links

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// LinkInfo represents a link found in a document
type LinkInfo struct {
	File       string
	Line       int
	LinkText   string
	Target     string
	IsRelative bool
	IsBroken   bool
}

// FileCheck represents the result of checking a file
type FileCheck struct {
	File        string
	TotalLinks  int
	BrokenLinks []LinkInfo
}

var (
	// Regex to find markdown links: [text](url)
	linkRegex = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	
	// Files to check
	filesToCheck = []string{
		"docs/README.md",
		"docs/technical/README.md",
		"docs/user-guide/README.md",
	}
	
	// Root directory (will be set based on script location)
	rootDir string
)

func main() {
	// Determine root directory - use current working directory
	var err error
	rootDir, err = os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}
	
	log.Println("Go-LLMs Documentation Link Validator")
	log.Println("====================================")
	log.Printf("Root directory: %s\n", rootDir)
	
	var allChecks []FileCheck
	totalBroken := 0
	
	// Check each file
	for _, file := range filesToCheck {
		fullPath := filepath.Join(rootDir, file)
		log.Printf("\nChecking: %s", file)
		
		check, err := checkFile(fullPath)
		if err != nil {
			log.Printf("Error checking %s: %v", file, err)
			continue
		}
		
		allChecks = append(allChecks, check)
		totalBroken += len(check.BrokenLinks)
		
		log.Printf("  Total links: %d", check.TotalLinks)
		log.Printf("  Broken links: %d", len(check.BrokenLinks))
	}
	
	// Report results
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("VALIDATION RESULTS")
	fmt.Println(strings.Repeat("=", 80))
	
	if totalBroken == 0 {
		fmt.Println("\n✅ All documentation links are valid!")
	} else {
		fmt.Printf("\n❌ Found %d broken links:\n\n", totalBroken)
		
		for _, check := range allChecks {
			if len(check.BrokenLinks) > 0 {
				fmt.Printf("File: %s\n", check.File)
				fmt.Println(strings.Repeat("-", 40))
				
				for _, link := range check.BrokenLinks {
					fmt.Printf("  Line %d: [%s](%s)\n", link.Line, link.LinkText, link.Target)
					
					// Suggest fix if possible
					if suggestion := suggestFix(link); suggestion != "" {
						fmt.Printf("    → Suggestion: %s\n", suggestion)
					}
				}
				fmt.Println()
			}
		}
	}
	
	// Generate fix script if there are broken links
	if totalBroken > 0 {
		generateFixScript(allChecks)
	}
	
	// Exit with error code if broken links found
	if totalBroken > 0 {
		os.Exit(1)
	}
}

func checkFile(filePath string) (FileCheck, error) {
	check := FileCheck{
		File:        filePath,
		BrokenLinks: []LinkInfo{},
	}
	
	file, err := os.Open(filePath)
	if err != nil {
		return check, err
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	lineNum := 0
	
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		
		// Find all links in the line
		matches := linkRegex.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) >= 3 {
				link := LinkInfo{
					File:     filePath,
					Line:     lineNum,
					LinkText: match[1],
					Target:   match[2],
				}
				
				// Skip external links and anchors
				if strings.HasPrefix(link.Target, "http://") || 
				   strings.HasPrefix(link.Target, "https://") ||
				   strings.HasPrefix(link.Target, "#") ||
				   strings.HasPrefix(link.Target, "mailto:") {
					continue
				}
				
				check.TotalLinks++
				link.IsRelative = true
				
				// Check if the target exists
				if !checkLinkTarget(filePath, link.Target) {
					link.IsBroken = true
					check.BrokenLinks = append(check.BrokenLinks, link)
				}
			}
		}
	}
	
	return check, scanner.Err()
}

func checkLinkTarget(sourceFile, target string) bool {
	// Get the directory of the source file
	sourceDir := filepath.Dir(sourceFile)
	
	// Handle absolute paths (from root)
	if strings.HasPrefix(target, "/") {
		targetPath := filepath.Join(rootDir, target[1:])
		return fileExists(targetPath)
	}
	
	// Handle relative paths
	targetPath := filepath.Join(sourceDir, target)
	targetPath = filepath.Clean(targetPath)
	
	// Check if it's a directory reference (ends with /)
	if strings.HasSuffix(target, "/") {
		return dirExists(targetPath)
	}
	
	// Check if file exists
	if fileExists(targetPath) {
		return true
	}
	
	// Check if it's a directory that exists
	if dirExists(targetPath) {
		return true
	}
	
	// Check if README.md exists in the directory
	readmePath := filepath.Join(targetPath, "README.md")
	if fileExists(readmePath) {
		return true
	}
	
	return false
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func suggestFix(link LinkInfo) string {
	target := link.Target
	
	// Common fixes
	fixes := map[string]string{
		// Tool references
		"BUILT-IN-TOOLS-REFERENCE.md": "user-guide/reference/built-in-tools-reference.md",
		"TOOL-USAGE-EXAMPLES.md": "user-guide/reference/tool-usage-examples.md",
		
		// Images directory
		"images/": "../images/",
		
		// API references that should point to api directory
		"api/llm.md": "../api/llm.md",
		"api/agent.md": "../api/agent.md",
		
		// Technical references
		"technical/api-reference/README.md": "api-reference/",
		
		// Archives
		"archives/README.md": "../archives/README.md",
	}
	
	// Check for direct matches
	if fix, ok := fixes[target]; ok {
		return fix
	}
	
	// Check for missing ../ prefix
	if strings.HasPrefix(target, "CONTRIBUTING") || 
	   strings.HasPrefix(target, "CHANGELOG") ||
	   strings.HasPrefix(target, "TODO") {
		return "../" + target
	}
	
	// Check for incorrect technical/ paths
	if strings.HasPrefix(target, "technical/") && strings.Contains(link.File, "/technical/") {
		// Remove redundant technical/ prefix
		return strings.TrimPrefix(target, "technical/")
	}
	
	return ""
}

func generateFixScript(checks []FileCheck) {
	scriptPath := filepath.Join(rootDir, "scripts", "fix-doc-links.sh")
	
	file, err := os.Create(scriptPath)
	if err != nil {
		log.Printf("Failed to create fix script: %v", err)
		return
	}
	defer file.Close()
	
	fmt.Fprintln(file, "#!/bin/bash")
	fmt.Fprintln(file, "# Auto-generated script to fix broken documentation links")
	fmt.Fprintln(file, "# Generated by validate-doc-links.go")
	fmt.Fprintln(file, "")
	fmt.Fprintln(file, "set -e")
	fmt.Fprintln(file, "")
	
	for _, check := range checks {
		if len(check.BrokenLinks) > 0 {
			relPath, _ := filepath.Rel(rootDir, check.File)
			fmt.Fprintf(file, "# Fixing %s\n", relPath)
			
			for _, link := range check.BrokenLinks {
				if suggestion := suggestFix(link); suggestion != "" {
					// Generate sed command to fix the link
					oldPattern := fmt.Sprintf(`\[%s\](%s)`, 
						escapeForSed(link.LinkText), 
						escapeForSed(link.Target))
					newPattern := fmt.Sprintf(`[%s](%s)`, 
						link.LinkText, 
						suggestion)
					
					fmt.Fprintf(file, "sed -i.bak 's|%s|%s|g' %s\n", 
						oldPattern, newPattern, relPath)
				}
			}
			fmt.Fprintln(file, "")
		}
	}
	
	fmt.Fprintln(file, "echo '✅ Link fixes applied!'")
	fmt.Fprintln(file, "echo 'Backup files created with .bak extension'")
	
	// Make script executable
	os.Chmod(scriptPath, 0755)
	
	log.Printf("\n💡 Fix script generated: %s", scriptPath)
	log.Println("   Run it to automatically fix the broken links")
}

func escapeForSed(s string) string {
	// Escape special characters for sed
	s = strings.ReplaceAll(s, "/", `\/`)
	s = strings.ReplaceAll(s, "[", `\[`)
	s = strings.ReplaceAll(s, "]", `\]`)
	s = strings.ReplaceAll(s, "(", `\(`)
	s = strings.ReplaceAll(s, ")", `\)`)
	s = strings.ReplaceAll(s, ".", `\.`)
	s = strings.ReplaceAll(s, "*", `\*`)
	return s
}