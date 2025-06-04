// ABOUTME: File content search tool with grep-like functionality
// ABOUTME: Built-in tool supporting pattern matching, regex, and context lines

package file

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	atools "github.com/lexlapax/go-llms/pkg/agent/tools"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// FileSearchParams defines parameters for the FileSearch tool
type FileSearchParams struct {
	Path               string `json:"path"`
	Pattern            string `json:"pattern"`                        // Search pattern (string or regex)
	FilePattern        string `json:"file_pattern,omitempty"`         // File name pattern (e.g., "*.txt")
	IsRegex            bool   `json:"is_regex,omitempty"`             // Treat pattern as regex
	CaseSensitive      bool   `json:"case_sensitive,omitempty"`       // Case-sensitive search
	Recursive          bool   `json:"recursive,omitempty"`            // Search subdirectories
	MaxResults         int    `json:"max_results,omitempty"`          // Limit total matches
	ContextLines       int    `json:"context_lines,omitempty"`        // Lines before/after match
	IncludeLineNumbers bool   `json:"include_line_numbers,omitempty"` // Include line numbers
}

// FileMatch represents a single match in a file
type FileMatch struct {
	File          string   `json:"file"`
	LineNumber    int      `json:"line_number"`
	Line          string   `json:"line"`
	MatchStart    int      `json:"match_start"` // Character position of match start
	MatchEnd      int      `json:"match_end"`   // Character position of match end
	ContextBefore []string `json:"context_before,omitempty"`
	ContextAfter  []string `json:"context_after,omitempty"`
}

// FileSearchResult defines the result of the FileSearch tool
type FileSearchResult struct {
	Matches       []FileMatch `json:"matches"`
	TotalMatches  int         `json:"total_matches"`
	FilesSearched int         `json:"files_searched"`
	Pattern       string      `json:"pattern"`
	SearchPath    string      `json:"search_path"`
}

// fileSearchParamSchema defines parameters for the FileSearch tool
var fileSearchParamSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"path": {
			Type:        "string",
			Description: "File or directory path to search",
		},
		"pattern": {
			Type:        "string",
			Description: "Search pattern (plain text or regex)",
		},
		"file_pattern": {
			Type:        "string",
			Description: "File name pattern to filter (e.g., '*.txt')",
		},
		"is_regex": {
			Type:        "boolean",
			Description: "Treat pattern as regular expression",
		},
		"case_sensitive": {
			Type:        "boolean",
			Description: "Perform case-sensitive search",
		},
		"recursive": {
			Type:        "boolean",
			Description: "Search subdirectories recursively",
		},
		"max_results": {
			Type:        "number",
			Description: "Maximum number of matches to return",
		},
		"context_lines": {
			Type:        "number",
			Description: "Number of context lines before/after matches",
		},
		"include_line_numbers": {
			Type:        "boolean",
			Description: "Include line numbers in results",
		},
	},
	Required: []string{"path", "pattern"},
}

// init automatically registers the tool on package import
func init() {
	tools.MustRegisterTool("file_search", FileSearch(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "file_search",
			Category:    "file",
			Tags:        []string{"filesystem", "search", "grep", "find", "pattern"},
			Description: "Searches for patterns in file contents",
			Version:     "1.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Simple text search",
					Description: "Search for text in a file",
					Code:        `FileSearch().Execute(ctx, FileSearchParams{Path: "file.txt", Pattern: "TODO"})`,
				},
				{
					Name:        "Recursive search",
					Description: "Search all .go files for a pattern",
					Code:        `FileSearch().Execute(ctx, FileSearchParams{Path: ".", Pattern: "func main", FilePattern: "*.go", Recursive: true})`,
				},
				{
					Name:        "Regex with context",
					Description: "Search with regex and show context",
					Code:        `FileSearch().Execute(ctx, FileSearchParams{Path: "logs/", Pattern: "ERROR.*failed", IsRegex: true, ContextLines: 2})`,
				},
			},
		},
		RequiredPermissions: []string{"filesystem:read"},
		ResourceUsage: tools.ResourceInfo{
			Memory:      "medium", // Can use memory for large files
			Network:     false,
			FileSystem:  true,
			Concurrency: true,
		},
	})
}

// FileSearch creates a tool for searching patterns in file contents
// This is a built-in tool optimized for:
// - Fast pattern matching with string search or regex
// - Flexible file filtering
// - Context line support for better understanding
// - Memory-efficient streaming for large files
func FileSearch() domain.Tool {
	return atools.NewTool(
		"file_search",
		"Searches for patterns in file contents",
		func(ctx *domain.ToolContext, params FileSearchParams) (*FileSearchResult, error) {
			// Emit start event
			if ctx.Events != nil {
				ctx.Events.Emit(domain.EventToolCall, domain.ToolCallEventData{
					ToolName:   "file_search",
					Parameters: params,
					RequestID:  ctx.RunID,
				})
			}

			// Clean and resolve the path
			searchPath := filepath.Clean(params.Path)
			absPath, err := filepath.Abs(searchPath)
			if err != nil {
				return nil, fmt.Errorf("invalid path: %w", err)
			}

			// Check file access restrictions from state
			if ctx.State != nil {
				if restrictedPaths, ok := ctx.State.Get("file_access_restrictions"); ok {
					if paths, ok := restrictedPaths.([]string); ok {
						for _, restricted := range paths {
							if strings.HasPrefix(absPath, restricted) {
								return nil, fmt.Errorf("access denied: path is restricted")
							}
						}
					}
				}
			}

			// Get configuration from state
			maxResults := params.MaxResults
			if maxResults == 0 {
				// Check state for default
				if ctx.State != nil {
					if val, ok := ctx.State.Get("file_search_max_results"); ok {
						if limit, ok := val.(int); ok && limit > 0 {
							maxResults = limit
						}
					}
				}
				if maxResults == 0 {
					maxResults = 1000 // Default limit
				}
			}

			// Get case sensitivity preference from state if not specified
			caseSensitive := params.CaseSensitive
			if ctx.State != nil && !params.CaseSensitive {
				if val, ok := ctx.State.Get("file_search_case_sensitive"); ok {
					if sensitive, ok := val.(bool); ok {
						caseSensitive = sensitive
					}
				}
			}

			// Get default encoding preference from state
			var encoding string
			if ctx.State != nil {
				if val, ok := ctx.State.Get("file_search_encoding"); ok {
					if enc, ok := val.(string); ok {
						encoding = enc
					}
				}
			}

			// Update params with calculated maxResults
			params.MaxResults = maxResults

			// Set include line numbers default
			if !params.IncludeLineNumbers {
				params.IncludeLineNumbers = true // Default to showing line numbers
			}

			// Compile regex if needed
			var searchRegex *regexp.Regexp
			searchPattern := params.Pattern
			if params.IsRegex {
				flags := ""
				if !caseSensitive {
					flags = "(?i)"
				}
				searchRegex, err = regexp.Compile(flags + searchPattern)
				if err != nil {
					if ctx.Events != nil {
						ctx.Events.EmitError(fmt.Errorf("invalid regex pattern: %w", err))
					}
					return nil, fmt.Errorf("invalid regex pattern: %w", err)
				}
			} else if !caseSensitive {
				searchPattern = strings.ToLower(searchPattern)
			}

			// Check if path exists
			info, err := os.Stat(absPath)
			if err != nil {
				if ctx.Events != nil {
					ctx.Events.EmitError(fmt.Errorf("path not found: %w", err))
				}
				return nil, fmt.Errorf("path not found: %w", err)
			}

			var matches []FileMatch
			filesSearched := 0

			// Search single file or directory
			if !info.IsDir() {
				// Single file search
				if ctx.Events != nil {
					ctx.Events.EmitProgress(0, 1, fmt.Sprintf("Searching file: %s", absPath))
				}

				fileMatches, err := searchFile(ctx.Context, absPath, searchPattern, searchRegex, params, encoding)
				if err != nil {
					return nil, err
				}
				matches = fileMatches
				filesSearched = 1

				if ctx.Events != nil {
					ctx.Events.EmitProgress(1, 1, "File search complete")
				}
			} else {
				// Directory search
				if ctx.Events != nil {
					ctx.Events.EmitMessage(fmt.Sprintf("Searching directory: %s", absPath))
				}

				// Count total files first for progress reporting
				totalFiles := 0
				processedFiles := 0
				if ctx.Events != nil {
					_ = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
						if err != nil || info.IsDir() {
							return nil
						}
						if params.FilePattern != "" {
							if matched, _ := filepath.Match(params.FilePattern, info.Name()); !matched {
								return nil
							}
						}
						totalFiles++
						return nil
					})
				}

				err = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
					// Check context cancellation
					select {
					case <-ctx.Done():
						return ctx.Err()
					default:
					}

					if err != nil {
						return nil // Skip files we can't access
					}

					// Skip directories
					if info.IsDir() {
						if !params.Recursive && path != absPath {
							return filepath.SkipDir
						}
						return nil
					}

					// Check file pattern
					if params.FilePattern != "" {
						matched, err := filepath.Match(params.FilePattern, info.Name())
						if err != nil || !matched {
							return nil
						}
					}

					// Skip binary files (simple heuristic)
					if isBinaryFile(path) {
						return nil
					}

					// Emit progress
					if ctx.Events != nil && totalFiles > 0 {
						processedFiles++
						ctx.Events.EmitProgress(processedFiles, totalFiles, fmt.Sprintf("Searching: %s", filepath.Base(path)))
					}

					// Search the file
					fileMatches, err := searchFile(ctx.Context, path, searchPattern, searchRegex, params, encoding)
					if err != nil {
						return nil // Skip files with errors
					}

					filesSearched++
					matches = append(matches, fileMatches...)

					// Check max results limit
					if len(matches) >= maxResults {
						if ctx.Events != nil {
							ctx.Events.EmitMessage(fmt.Sprintf("Reached maximum results limit: %d", maxResults))
						}
						return filepath.SkipAll
					}

					return nil
				})

				if err != nil && err != filepath.SkipAll && err != context.Canceled {
					return nil, fmt.Errorf("error searching directory: %w", err)
				}
			}

			// Trim to max results
			if len(matches) > maxResults {
				matches = matches[:maxResults]
			}

			result := &FileSearchResult{
				Matches:       matches,
				TotalMatches:  len(matches),
				FilesSearched: filesSearched,
				Pattern:       params.Pattern,
				SearchPath:    absPath,
			}

			// Emit completion event with details
			if ctx.Events != nil {
				ctx.Events.EmitCustom("file_search_complete", map[string]interface{}{
					"total_matches":  len(matches),
					"files_searched": filesSearched,
					"pattern":        params.Pattern,
					"search_path":    absPath,
					"is_regex":       params.IsRegex,
					"case_sensitive": caseSensitive,
					"recursive":      params.Recursive,
					"file_pattern":   params.FilePattern,
					"elapsed_time":   ctx.ElapsedTime().String(),
				})

				ctx.Events.Emit(domain.EventToolResult, domain.ToolResultEventData{
					ToolName:  "file_search",
					RequestID: ctx.RunID,
					Result:    result,
				})
			}

			return result, nil
		},
		fileSearchParamSchema,
	)
}

// searchFile searches for pattern in a single file
func searchFile(ctx context.Context, filePath string, pattern string, regex *regexp.Regexp, params FileSearchParams, encoding string) ([]FileMatch, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var matches []FileMatch
	scanner := bufio.NewScanner(file)

	// Limit line length to prevent memory issues
	const maxLineLength = 1024 * 1024 // 1MB
	scanner.Buffer(make([]byte, 0, 64*1024), maxLineLength)

	lineNum := 0
	var contextBuffer []string

	for scanner.Scan() {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return matches, ctx.Err()
		default:
		}

		lineNum++
		line := scanner.Text()

		// Search for pattern
		var matchStart, matchEnd int
		found := false

		if regex != nil {
			// Regex search
			loc := regex.FindStringIndex(line)
			if loc != nil {
				found = true
				matchStart = loc[0]
				matchEnd = loc[1]
			}
		} else {
			// Plain text search
			searchLine := line
			if !params.CaseSensitive {
				searchLine = strings.ToLower(line)
			}
			index := strings.Index(searchLine, pattern)
			if index >= 0 {
				found = true
				matchStart = index
				matchEnd = index + len(pattern)
			}
		}

		if found {
			match := FileMatch{
				File:       filePath,
				LineNumber: lineNum,
				Line:       line,
				MatchStart: matchStart,
				MatchEnd:   matchEnd,
			}

			// Add context lines if requested
			if params.ContextLines > 0 {
				// Get context before (from buffer)
				contextStart := len(contextBuffer) - params.ContextLines
				if contextStart < 0 {
					contextStart = 0
				}
				if contextStart < len(contextBuffer) {
					match.ContextBefore = make([]string, len(contextBuffer)-contextStart)
					copy(match.ContextBefore, contextBuffer[contextStart:])
				}

				// Read context after
				match.ContextAfter = readContextAfter(scanner, params.ContextLines, &lineNum)
			}

			matches = append(matches, match)

			// Check max results
			if len(matches) >= params.MaxResults {
				break
			}
		}

		// Maintain context buffer
		if params.ContextLines > 0 {
			contextBuffer = append(contextBuffer, line)
			if len(contextBuffer) > params.ContextLines {
				contextBuffer = contextBuffer[1:]
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return matches, fmt.Errorf("error reading file: %w", err)
	}

	return matches, nil
}

// readContextAfter reads n lines after the current position
func readContextAfter(scanner *bufio.Scanner, n int, lineNum *int) []string {
	var lines []string
	for i := 0; i < n && scanner.Scan(); i++ {
		*lineNum++
		lines = append(lines, scanner.Text())
	}
	return lines
}

// isBinaryFile performs a simple check to detect binary files
func isBinaryFile(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return true // Assume binary if can't open
	}
	defer file.Close()

	// Read first 512 bytes
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return true
	}

	// Check for null bytes (common in binary files)
	for i := 0; i < n; i++ {
		if buffer[i] == 0 {
			return true
		}
	}

	return false
}

// MustGetFileSearch retrieves the registered FileSearch tool or panics
// This is a convenience function for users who want to ensure the tool exists
func MustGetFileSearch() domain.Tool {
	return tools.MustGetTool("file_search")
}
