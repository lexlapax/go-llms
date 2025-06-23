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

// fileSearchOutputSchema defines the output schema for the FileSearch tool
var fileSearchOutputSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"matches": {
			Type:        "array",
			Description: "List of matches found",
			Items: &sdomain.Property{
				Type: "object",
				Properties: map[string]sdomain.Property{
					"file": {
						Type:        "string",
						Description: "File path where match was found",
					},
					"line_number": {
						Type:        "number",
						Description: "Line number of the match",
					},
					"line": {
						Type:        "string",
						Description: "Full line containing the match",
					},
					"match_start": {
						Type:        "number",
						Description: "Character position where match starts",
					},
					"match_end": {
						Type:        "number",
						Description: "Character position where match ends",
					},
					"context_before": {
						Type:        "array",
						Description: "Lines before the match (if context_lines > 0)",
						Items: &sdomain.Property{
							Type: "string",
						},
					},
					"context_after": {
						Type:        "array",
						Description: "Lines after the match (if context_lines > 0)",
						Items: &sdomain.Property{
							Type: "string",
						},
					},
				},
				Required: []string{"file", "line_number", "line", "match_start", "match_end"},
			},
		},
		"total_matches": {
			Type:        "number",
			Description: "Total number of matches found",
		},
		"files_searched": {
			Type:        "number",
			Description: "Number of files searched",
		},
		"pattern": {
			Type:        "string",
			Description: "The search pattern used",
		},
		"search_path": {
			Type:        "string",
			Description: "The path that was searched",
		},
	},
	Required: []string{"matches", "total_matches", "files_searched", "pattern", "search_path"},
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

// fileSearchMain is the main function for the tool
func fileSearchMain(ctx *domain.ToolContext, params FileSearchParams) (*FileSearchResult, error) {
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
}

// FileSearch creates a tool for searching patterns in file contents with grep-like functionality.
// It supports both plain text and regular expression patterns, case-sensitive/insensitive matching, and context lines.
// The tool can search single files or recursively traverse directories with file pattern filtering and binary file detection.
// Results include exact match positions and optional surrounding context for better understanding of matches.
func FileSearch() domain.Tool {
	builder := atools.NewToolBuilder("file_search", "Searches for patterns in file contents").
		WithFunction(fileSearchMain).
		WithParameterSchema(fileSearchParamSchema).
		WithOutputSchema(fileSearchOutputSchema).
		WithUsageInstructions(`Use this tool to search for text patterns within files, similar to grep.

Features:
- Plain text or regex pattern matching
- Case-sensitive or case-insensitive search
- File filtering by name patterns (glob)
- Recursive directory searching
- Context lines before/after matches
- Binary file detection and skipping
- Progress tracking for large searches

Parameters:
- path: File or directory to search (required)
- pattern: Search pattern (required)
- file_pattern: Filter files by name (e.g., *.txt, *.go)
- is_regex: Treat pattern as regular expression
- case_sensitive: Enable case-sensitive matching
- recursive: Search subdirectories
- max_results: Limit number of matches (default: 1000)
- context_lines: Show N lines before/after matches
- include_line_numbers: Show line numbers (default: true)

Pattern Matching:
- Plain text: Exact substring matching
- Regex: Full regular expression support
- Case-insensitive: Controlled by case_sensitive flag
- Special regex chars: . * + ? ^ $ [] {} () | \

File Filtering:
- Use file_pattern for glob matching
- Examples: *.txt, test_*, *.{js,ts}
- Applied to filename only, not path

Context Lines:
- Shows surrounding lines for better understanding
- context_before: Lines preceding the match
- context_after: Lines following the match
- Useful for understanding code context

State Configuration:
- file_access_restrictions: Restricted paths
- file_search_max_results: Default result limit
- file_search_case_sensitive: Default case sensitivity
- file_search_encoding: Default file encoding

Performance:
- Streams files to handle large files efficiently
- Binary files automatically skipped
- Progress reporting for directory searches
- Cancellable via context

Events Emitted:
- Tool call/result events
- Progress events during search
- file_search_complete with statistics
- Error events for invalid patterns

Best Practices:
- Use file_pattern to narrow search scope
- Enable recursive for project-wide searches
- Use context_lines for code searches
- Set reasonable max_results to avoid overload
- Use regex for complex pattern matching`).
		WithExamples([]domain.ToolExample{
			{
				Name:        "Simple text search",
				Description: "Find occurrences of a word",
				Scenario:    "When looking for specific text",
				Input: map[string]interface{}{
					"path":    "document.txt",
					"pattern": "TODO",
				},
				Output: map[string]interface{}{
					"matches": []map[string]interface{}{
						{
							"file":        "/home/user/document.txt",
							"line_number": 15,
							"line":        "// TODO: Implement error handling",
							"match_start": 3,
							"match_end":   7,
						},
					},
					"total_matches":  1,
					"files_searched": 1,
					"pattern":        "TODO",
					"search_path":    "/home/user/document.txt",
				},
				Explanation: "Found TODO comment on line 15",
			},
			{
				Name:        "Recursive code search",
				Description: "Find function definitions in Go files",
				Scenario:    "When searching across a codebase",
				Input: map[string]interface{}{
					"path":         "src/",
					"pattern":      "func main",
					"file_pattern": "*.go",
					"recursive":    true,
				},
				Output: map[string]interface{}{
					"matches": []map[string]interface{}{
						{
							"file":        "/project/src/cmd/app/main.go",
							"line_number": 10,
							"line":        "func main() {",
							"match_start": 0,
							"match_end":   9,
						},
						{
							"file":        "/project/src/examples/demo.go",
							"line_number": 8,
							"line":        "func main() {",
							"match_start": 0,
							"match_end":   9,
						},
					},
					"total_matches":  2,
					"files_searched": 45,
					"pattern":        "func main",
					"search_path":    "/project/src",
				},
				Explanation: "Found main functions in 2 Go files",
			},
			{
				Name:        "Regex with context",
				Description: "Find error patterns with surrounding context",
				Scenario:    "When analyzing log files",
				Input: map[string]interface{}{
					"path":          "app.log",
					"pattern":       "ERROR.*database",
					"is_regex":      true,
					"context_lines": 2,
				},
				Output: map[string]interface{}{
					"matches": []map[string]interface{}{
						{
							"file":           "/var/log/app.log",
							"line_number":    156,
							"line":           "[2024-01-15 10:30:15] ERROR: database connection failed",
							"match_start":    23,
							"match_end":      38,
							"context_before": []string{"[2024-01-15 10:30:14] INFO: Attempting database connection", "[2024-01-15 10:30:14] DEBUG: Using connection string: ..."},
							"context_after":  []string{"[2024-01-15 10:30:15] WARN: Retrying connection in 5s", "[2024-01-15 10:30:20] INFO: Connection retry attempt 1"},
						},
					},
					"total_matches":  1,
					"files_searched": 1,
					"pattern":        "ERROR.*database",
					"search_path":    "/var/log/app.log",
				},
				Explanation: "Database error with context showing retry attempts",
			},
			{
				Name:        "Case-insensitive search",
				Description: "Find variables regardless of case",
				Scenario:    "When variable naming is inconsistent",
				Input: map[string]interface{}{
					"path":           "config.ini",
					"pattern":        "api_key",
					"case_sensitive": false,
				},
				Output: map[string]interface{}{
					"matches": []map[string]interface{}{
						{
							"file":        "/app/config.ini",
							"line_number": 5,
							"line":        "API_KEY=secret123",
							"match_start": 0,
							"match_end":   7,
						},
						{
							"file":        "/app/config.ini",
							"line_number": 12,
							"line":        "backup_api_key=secret456",
							"match_start": 7,
							"match_end":   14,
						},
					},
					"total_matches":  2,
					"files_searched": 1,
					"pattern":        "api_key",
					"search_path":    "/app/config.ini",
				},
				Explanation: "Found both uppercase and lowercase variants",
			},
			{
				Name:        "Limited results",
				Description: "Search with result limit",
				Scenario:    "When there are many matches",
				Input: map[string]interface{}{
					"path":        "logs/",
					"pattern":     "INFO",
					"recursive":   true,
					"max_results": 10,
				},
				Output: map[string]interface{}{
					"matches":        []map[string]interface{}{}, // 10 matches
					"total_matches":  10,
					"files_searched": 3,
					"pattern":        "INFO",
					"search_path":    "/app/logs",
				},
				Explanation: "Stopped at 10 matches though more exist",
			},
			{
				Name:        "Multiple file types",
				Description: "Search specific file extensions",
				Scenario:    "When searching documentation",
				Input: map[string]interface{}{
					"path":         "docs/",
					"pattern":      "deprecated",
					"file_pattern": "*.{md,txt,rst}",
					"recursive":    true,
				},
				Output: map[string]interface{}{
					"matches": []map[string]interface{}{
						{
							"file":        "/project/docs/api.md",
							"line_number": 45,
							"line":        "**Deprecated**: This endpoint will be removed in v2.0",
						},
						{
							"file":        "/project/docs/changelog.txt",
							"line_number": 23,
							"line":        "- Deprecated old authentication method",
						},
					},
					"total_matches":  2,
					"files_searched": 15,
					"pattern":        "deprecated",
					"search_path":    "/project/docs",
				},
				Explanation: "Found deprecation notices in documentation files",
			},
			{
				Name:        "Complex regex pattern",
				Description: "Extract email addresses",
				Scenario:    "When finding contact information",
				Input: map[string]interface{}{
					"path":     "contacts.csv",
					"pattern":  "[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}",
					"is_regex": true,
				},
				Output: map[string]interface{}{
					"matches": []map[string]interface{}{
						{
							"file":        "/data/contacts.csv",
							"line_number": 3,
							"line":        "John Doe,john.doe@example.com,Marketing",
							"match_start": 9,
							"match_end":   29,
						},
					},
					"total_matches":  1,
					"files_searched": 1,
					"pattern":        "[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}",
					"search_path":    "/data/contacts.csv",
				},
				Explanation: "Extracted email address using regex pattern",
			},
		}).
		WithConstraints([]string{
			"Binary files are automatically skipped",
			"Large files are streamed, not loaded entirely into memory",
			"Line length limited to 1MB to prevent memory issues",
			"Regex patterns must be valid Go regexp syntax",
			"File patterns use filepath.Match glob syntax, not regex",
			"Context lines may be truncated at file boundaries",
			"Hidden files (starting with .) are included in searches",
			"Symbolic links are followed during recursive search",
			"Search results limited by max_results parameter",
			"Case-insensitive regex uses (?i) flag prefix",
		}).
		WithErrorGuidance(map[string]string{
			"invalid regex pattern":     "Check regex syntax, escape special characters properly",
			"path not found":            "Verify the file or directory exists and path is correct",
			"access denied":             "Check file permissions or restricted paths",
			"pattern too complex":       "Simplify regex pattern or increase system resources",
			"context deadline exceeded": "Search taking too long, try narrowing scope",
			"invalid path":              "Path contains invalid characters or is malformed",
			"file too large":            "File exceeds reasonable size for text search",
			"permission denied":         "Insufficient permissions to read file or directory",
		}).
		WithCategory("file").
		WithTags([]string{"filesystem", "search", "grep", "find", "pattern"}).
		WithVersion("2.0.0").
		WithBehavior(
			false,    // Not deterministic - file contents can change
			false,    // Not destructive - only reads files
			false,    // No confirmation needed
			"medium", // Can be slow for large directories
		)

	return builder.Build()
}

// searchFile searches for pattern in a single file
func searchFile(ctx context.Context, filePath string, pattern string, regex *regexp.Regexp, params FileSearchParams, encoding string) ([]FileMatch, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

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
	defer func() {
		_ = file.Close()
	}()

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
