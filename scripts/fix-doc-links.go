package main

// ABOUTME: Script to fix broken documentation links in markdown files
// ABOUTME: Converts absolute paths to relative and fixes navigation patterns

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// Match markdown links
	linkRe = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	
	// Breadcrumb pattern
	breadcrumbRe = regexp.MustCompile(`\[([^\]]+)\]\(/docs/[^)]+\)`)
)

func main() {
	log.Println("Go-LLMs Documentation Link Fixer")
	log.Println("================================")
	
	totalFixed := 0
	filesProcessed := 0
	
	// Walk through all markdown files in docs
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
			log.Printf("Fixed %d links in %s", fixed, path)
			totalFixed += fixed
			filesProcessed++
		}
		
		return nil
	})
	
	if err != nil {
		log.Fatalf("Error walking directory: %v", err)
	}
	
	log.Printf("\n✅ Fixed %d links across %d files", totalFixed, filesProcessed)
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
	
	// Fix breadcrumb navigation
	updatedContent = fixBreadcrumbs(filePath, updatedContent, &fixCount)
	
	// Fix other links
	updatedContent = fixLinks(filePath, updatedContent, &fixCount)
	
	// Only write if changes were made
	if fixCount > 0 {
		err = os.WriteFile(filePath, []byte(updatedContent), 0644)
		if err != nil {
			return 0, err
		}
	}
	
	return fixCount, nil
}

func fixBreadcrumbs(filePath, content string, fixCount *int) string {
	// Get the current file's depth in the docs structure
	relPath, _ := filepath.Rel("docs", filePath)
	depth := strings.Count(relPath, string(os.PathSeparator))
	
	// Build the relative path prefix
	prefix := strings.Repeat("../", depth)
	
	// Replace absolute breadcrumb links with relative ones
	return breadcrumbRe.ReplaceAllStringFunc(content, func(match string) string {
		// Extract link text and path
		parts := linkRe.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}
		
		linkText := parts[1]
		linkPath := parts[2]
		
		// Convert absolute /docs/ path to relative
		if strings.HasPrefix(linkPath, "/docs/") {
			newPath := strings.TrimPrefix(linkPath, "/docs/")
			newPath = prefix + newPath
			// Clean up the path
			newPath = filepath.Clean(newPath)
			newPath = strings.ReplaceAll(newPath, string(os.PathSeparator), "/")
			
			*fixCount++
			return fmt.Sprintf("[%s](%s)", linkText, newPath)
		}
		
		return match
	})
}

func fixLinks(filePath, content string, fixCount *int) string {
	// Get the directory of the current file
	fileDir := filepath.Dir(filePath)
	
	lines := strings.Split(content, "\n")
	var updatedLines []string
	
	for _, line := range lines {
		// Find all links in the line
		matches := linkRe.FindAllStringSubmatch(line, -1)
		updatedLine := line
		
		for _, match := range matches {
			if len(match) < 3 {
				continue
			}
			
			linkText := match[1]
			linkPath := match[2]
			
			// Skip external links and anchors
			if strings.HasPrefix(linkPath, "http") || strings.HasPrefix(linkPath, "#") {
				continue
			}
			
			// Fix absolute paths starting with /docs/
			if strings.HasPrefix(linkPath, "/docs/") {
				// Convert to relative path from current file
				absPath := strings.TrimPrefix(linkPath, "/docs/")
				relPath, err := filepath.Rel(fileDir, filepath.Join("docs", absPath))
				if err == nil {
					relPath = strings.ReplaceAll(relPath, string(os.PathSeparator), "/")
					newLink := fmt.Sprintf("[%s](%s)", linkText, relPath)
					updatedLine = strings.Replace(updatedLine, match[0], newLink, 1)
					*fixCount++
				}
			}
			
			// Fix image paths
			if strings.Contains(linkPath, "/images/") && !strings.HasPrefix(linkPath, "../") {
				// Assume images are in docs/images/
				relPath, err := filepath.Rel(fileDir, filepath.Join("docs/images", filepath.Base(linkPath)))
				if err == nil {
					relPath = strings.ReplaceAll(relPath, string(os.PathSeparator), "/")
					newLink := fmt.Sprintf("[%s](%s)", linkText, relPath)
					updatedLine = strings.Replace(updatedLine, match[0], newLink, 1)
					*fixCount++
				}
			}
		}
		
		updatedLines = append(updatedLines, updatedLine)
	}
	
	return strings.Join(updatedLines, "\n")
}