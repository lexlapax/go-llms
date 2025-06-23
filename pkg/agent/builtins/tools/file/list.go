// ABOUTME: File listing tool for directory enumeration with filters
// ABOUTME: Built-in tool supporting pattern matching, size/date filters, and sorting

package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	atools "github.com/lexlapax/go-llms/pkg/agent/tools"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// FileListParams defines parameters for the FileList tool
type FileListParams struct {
	Path           string `json:"path"`
	Pattern        string `json:"pattern,omitempty"`         // glob pattern like "*.txt"
	Recursive      bool   `json:"recursive,omitempty"`       // search subdirectories
	IncludeDirs    bool   `json:"include_dirs,omitempty"`    // include directories in results
	IncludeFiles   bool   `json:"include_files,omitempty"`   // include files in results (default: true)
	MinSize        int64  `json:"min_size,omitempty"`        // minimum file size in bytes
	MaxSize        int64  `json:"max_size,omitempty"`        // maximum file size in bytes
	ModifiedAfter  string `json:"modified_after,omitempty"`  // RFC3339 timestamp
	ModifiedBefore string `json:"modified_before,omitempty"` // RFC3339 timestamp
	SortBy         string `json:"sort_by,omitempty"`         // name, size, modified
	SortReverse    bool   `json:"sort_reverse,omitempty"`    // reverse sort order
	MaxResults     int    `json:"max_results,omitempty"`     // limit number of results
}

// FileInfo represents information about a file or directory
type FileInfo struct {
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	IsDir        bool      `json:"is_dir"`
	Size         int64     `json:"size"`
	Mode         string    `json:"mode"`
	ModifiedTime time.Time `json:"modified_time"`
	Extension    string    `json:"extension,omitempty"`
}

// FileListResult defines the result of the FileList tool
type FileListResult struct {
	Files       []FileInfo `json:"files"`
	TotalCount  int        `json:"total_count"`
	FilteredOut int        `json:"filtered_out"`
	SearchPath  string     `json:"search_path"`
	Pattern     string     `json:"pattern,omitempty"`
}

// fileListParamSchema defines parameters for the FileList tool
var fileListParamSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"path": {
			Type:        "string",
			Description: "Directory path to list",
		},
		"pattern": {
			Type:        "string",
			Description: "Glob pattern to match files (e.g., '*.txt', 'test_*')",
		},
		"recursive": {
			Type:        "boolean",
			Description: "Search subdirectories recursively",
		},
		"include_dirs": {
			Type:        "boolean",
			Description: "Include directories in results",
		},
		"include_files": {
			Type:        "boolean",
			Description: "Include files in results (default: true)",
		},
		"min_size": {
			Type:        "number",
			Description: "Minimum file size in bytes",
		},
		"max_size": {
			Type:        "number",
			Description: "Maximum file size in bytes",
		},
		"modified_after": {
			Type:        "string",
			Format:      "date-time",
			Description: "Only files modified after this time (RFC3339)",
		},
		"modified_before": {
			Type:        "string",
			Format:      "date-time",
			Description: "Only files modified before this time (RFC3339)",
		},
		"sort_by": {
			Type:        "string",
			Description: "Sort results by: name, size, or modified",
		},
		"sort_reverse": {
			Type:        "boolean",
			Description: "Reverse sort order",
		},
		"max_results": {
			Type:        "number",
			Description: "Maximum number of results to return",
		},
	},
	Required: []string{"path"},
}

// fileListOutputSchema defines the output schema for the FileList tool
var fileListOutputSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"files": {
			Type:        "array",
			Description: "List of files and directories found",
			Items: &sdomain.Property{
				Type: "object",
				Properties: map[string]sdomain.Property{
					"name": {
						Type:        "string",
						Description: "File or directory name",
					},
					"path": {
						Type:        "string",
						Description: "Full path to the file or directory",
					},
					"is_dir": {
						Type:        "boolean",
						Description: "Whether this is a directory",
					},
					"size": {
						Type:        "number",
						Description: "File size in bytes (0 for directories)",
					},
					"mode": {
						Type:        "string",
						Description: "File permissions mode string",
					},
					"modified_time": {
						Type:        "string",
						Description: "Last modification time (RFC3339)",
					},
					"extension": {
						Type:        "string",
						Description: "File extension without dot (empty for directories)",
					},
				},
				Required: []string{"name", "path", "is_dir", "size", "mode", "modified_time"},
			},
		},
		"total_count": {
			Type:        "number",
			Description: "Total number of items scanned",
		},
		"filtered_out": {
			Type:        "number",
			Description: "Number of items filtered out",
		},
		"search_path": {
			Type:        "string",
			Description: "Absolute path that was searched",
		},
		"pattern": {
			Type:        "string",
			Description: "Pattern used for filtering (if any)",
		},
	},
	Required: []string{"files", "total_count", "filtered_out", "search_path"},
}

// init automatically registers the tool on package import
func init() {
	tools.MustRegisterTool("file_list", FileList(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "file_list",
			Category:    "file",
			Tags:        []string{"filesystem", "directory", "list", "search"},
			Description: "Lists files and directories with filtering options",
			Version:     "1.0.0",
			Examples: []builtins.Example{
				{
					Name:        "List all files",
					Description: "List all files in a directory",
					Code:        `FileList().Execute(ctx, FileListParams{Path: "/tmp"})`,
				},
				{
					Name:        "Find text files",
					Description: "Find all .txt files recursively",
					Code:        `FileList().Execute(ctx, FileListParams{Path: "/docs", Pattern: "*.txt", Recursive: true})`,
				},
				{
					Name:        "Recent large files",
					Description: "Find files over 1MB modified in last 24 hours",
					Code:        `FileList().Execute(ctx, FileListParams{Path: ".", MinSize: 1048576, ModifiedAfter: time.Now().Add(-24*time.Hour).Format(time.RFC3339)})`,
				},
			},
		},
		RequiredPermissions: []string{"filesystem:read"},
		ResourceUsage: tools.ResourceInfo{
			Memory:      "low",
			Network:     false,
			FileSystem:  true,
			Concurrency: true,
		},
	})
}

// fileListMain is the main function for the tool
func fileListMain(ctx *domain.ToolContext, params FileListParams) (*FileListResult, error) {
	// Emit start event
	if ctx.Events != nil {
		ctx.Events.EmitMessage(fmt.Sprintf("Starting directory listing for %s", params.Path))
	}

	// Get configuration from state
	showHidden := false
	if ctx.State != nil {
		// Check for show hidden files preference
		if val, exists := ctx.State.Get("file_list_show_hidden"); exists {
			if show, ok := val.(bool); ok {
				showHidden = show
			}
		}

		// Check for default sort preference
		if params.SortBy == "" {
			if sortPref, exists := ctx.State.Get("file_list_default_sort"); exists {
				if sort, ok := sortPref.(string); ok {
					params.SortBy = sort
				}
			}
		}

		// Check for default max results
		if params.MaxResults == 0 {
			if maxResults, exists := ctx.State.Get("file_list_max_results"); exists {
				if max, ok := maxResults.(int); ok {
					params.MaxResults = max
				}
			}
		}
	}

	// Set remaining defaults
	if params.SortBy == "" {
		params.SortBy = "name"
	}
	if !params.IncludeDirs && !params.IncludeFiles {
		params.IncludeFiles = true // default to including files
	}

	// Clean and resolve the path
	searchPath := filepath.Clean(params.Path)
	absPath, err := filepath.Abs(searchPath)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	// Check file access restrictions from state
	if ctx.State != nil {
		if restrictedPaths, exists := ctx.State.Get("file_restricted_paths"); exists {
			if paths, ok := restrictedPaths.([]string); ok {
				for _, restricted := range paths {
					if strings.HasPrefix(absPath, restricted) {
						return nil, fmt.Errorf("access denied: path %s is restricted", absPath)
					}
				}
			}
		}

		// Check allowed paths if specified
		if allowedPaths, exists := ctx.State.Get("file_allowed_paths"); exists {
			if paths, ok := allowedPaths.([]string); ok && len(paths) > 0 {
				allowed := false
				for _, allowedPath := range paths {
					if strings.HasPrefix(absPath, allowedPath) {
						allowed = true
						break
					}
				}
				if !allowed {
					return nil, fmt.Errorf("access denied: path %s is not in allowed paths", absPath)
				}
			}
		}
	}

	// Verify the directory exists
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("path not found: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", absPath)
	}

	// Parse time filters
	var modifiedAfter, modifiedBefore time.Time
	if params.ModifiedAfter != "" {
		modifiedAfter, err = time.Parse(time.RFC3339, params.ModifiedAfter)
		if err != nil {
			return nil, fmt.Errorf("invalid modified_after time: %w", err)
		}
	}
	if params.ModifiedBefore != "" {
		modifiedBefore, err = time.Parse(time.RFC3339, params.ModifiedBefore)
		if err != nil {
			return nil, fmt.Errorf("invalid modified_before time: %w", err)
		}
	}

	// Emit event for starting directory enumeration
	if ctx.Events != nil {
		ctx.Events.EmitProgress(0, 0, fmt.Sprintf("Enumerating directory: %s", absPath))
	}

	// Collect files
	var files []FileInfo
	var totalCount, filteredOut int
	var directoriesProcessed int

	walkFunc := func(path string, info os.FileInfo, err error) error {
		// Check context cancellation
		select {
		case <-ctx.Context.Done():
			return ctx.Context.Err()
		default:
		}

		if err != nil {
			// Skip files we can't access
			filteredOut++
			return nil
		}

		// Skip the root directory itself
		if path == absPath {
			return nil
		}

		// Check if we're in a subdirectory when not recursive
		if !params.Recursive {
			dir := filepath.Dir(path)
			if dir != absPath {
				// We're in a subdirectory, skip it
				return filepath.SkipDir
			}
		}

		// Check hidden files
		if !showHidden && strings.HasPrefix(info.Name(), ".") {
			filteredOut++
			if info.IsDir() && params.Recursive {
				return filepath.SkipDir // Skip hidden directories entirely
			}
			return nil
		}

		totalCount++

		// Emit progress periodically for large directories
		if totalCount%100 == 0 && ctx.Events != nil {
			ctx.Events.EmitProgress(totalCount, 0, fmt.Sprintf("Processed %d items", totalCount))
		}

		// Apply filters
		isDir := info.IsDir()

		// Track directories processed
		if isDir {
			directoriesProcessed++
		}

		// Type filter
		if isDir && !params.IncludeDirs {
			filteredOut++
			// Still need to decide whether to recurse into this directory
			if !params.Recursive {
				return filepath.SkipDir
			}
			return nil
		}
		if !isDir && !params.IncludeFiles {
			filteredOut++
			return nil
		}

		// Pattern filter
		if params.Pattern != "" {
			matched, err := filepath.Match(params.Pattern, info.Name())
			if err != nil || !matched {
				filteredOut++
				// For directories, we still might want to recurse
				if isDir && params.Recursive {
					return nil
				}
				return nil
			}
		}

		// Size filter (only for files)
		if !isDir {
			size := info.Size()
			if params.MinSize > 0 && size < params.MinSize {
				filteredOut++
				return nil
			}
			if params.MaxSize > 0 && size > params.MaxSize {
				filteredOut++
				return nil
			}
		}

		// Time filters
		modTime := info.ModTime()
		if !modifiedAfter.IsZero() && modTime.Before(modifiedAfter) {
			filteredOut++
			return nil
		}
		if !modifiedBefore.IsZero() && modTime.After(modifiedBefore) {
			filteredOut++
			return nil
		}

		// Create file info
		fileInfo := FileInfo{
			Name:         info.Name(),
			Path:         path,
			IsDir:        isDir,
			Size:         info.Size(),
			Mode:         info.Mode().String(),
			ModifiedTime: modTime,
		}

		// Add extension for files
		if !isDir {
			fileInfo.Extension = strings.TrimPrefix(filepath.Ext(info.Name()), ".")
		}

		files = append(files, fileInfo)

		// Stop recursion for directories if not recursive
		if isDir && !params.Recursive {
			return filepath.SkipDir
		}

		return nil
	}

	// Walk the directory
	err = filepath.Walk(absPath, walkFunc)
	if err != nil && err != context.Canceled {
		return nil, fmt.Errorf("error walking directory: %w", err)
	}

	// Emit filtering progress
	if ctx.Events != nil && len(files) > 0 {
		ctx.Events.EmitProgress(totalCount, totalCount, "Sorting and filtering results")
	}

	// Sort results
	sortFiles(files, params.SortBy, params.SortReverse)

	// Apply max results limit
	truncated := false
	if params.MaxResults > 0 && len(files) > params.MaxResults {
		files = files[:params.MaxResults]
		truncated = true
	}

	// Emit completion event with summary
	if ctx.Events != nil {
		summary := map[string]interface{}{
			"path":                  absPath,
			"pattern":               params.Pattern,
			"files_found":           len(files),
			"total_scanned":         totalCount,
			"filtered_out":          filteredOut,
			"directories_processed": directoriesProcessed,
			"recursive":             params.Recursive,
			"truncated":             truncated,
			"sort_by":               params.SortBy,
			"elapsed_time":          ctx.ElapsedTime().String(),
		}
		ctx.Events.EmitCustom("file_list_complete", summary)
	}

	return &FileListResult{
		Files:       files,
		TotalCount:  totalCount,
		FilteredOut: filteredOut,
		SearchPath:  absPath,
		Pattern:     params.Pattern,
	}, nil
}

// FileList creates a tool for listing files and directories with extensive filtering and sorting options.
// It supports recursive directory traversal, pattern matching, size/date filtering, and custom sorting.
// The tool efficiently handles large directory structures with progress reporting and context cancellation support.
// Results can be limited and filtered based on multiple criteria including file patterns, size ranges, and modification times.
func FileList() domain.Tool {
	builder := atools.NewToolBuilder("file_list", "Lists files and directories with filtering options").
		WithFunction(fileListMain).
		WithParameterSchema(fileListParamSchema).
		WithOutputSchema(fileListOutputSchema).
		WithUsageInstructions(`Use this tool to list files and directories with extensive filtering options.

Features:
- Fast directory enumeration
- Flexible pattern matching (glob patterns)
- Recursive directory traversal
- Size-based filtering (min/max)
- Date-based filtering (modified before/after)
- Multiple sort options
- Hidden file control via state

Parameters:
- path: Directory to list (required)
- pattern: Glob pattern (e.g., *.txt, test_*, *.{jpg,png})
- recursive: Search subdirectories (default: false)
- include_dirs: Include directories in results (default: false)
- include_files: Include files in results (default: true)
- min_size/max_size: Filter by file size in bytes
- modified_after/before: Filter by modification time (RFC3339)
- sort_by: Sort by name, size, or modified (default: name)
- sort_reverse: Reverse sort order
- max_results: Limit number of results

Pattern Matching:
- Supports standard glob patterns
- * matches any sequence of characters
- ? matches any single character
- [abc] matches any character in brackets
- [a-z] matches any character in range
- {jpg,png} matches any of the alternatives

Size Filtering:
- Sizes are in bytes
- min_size: 1048576 = 1MB
- max_size: 10485760 = 10MB
- Only applies to files, not directories

Date Filtering:
- Use RFC3339 format: 2024-01-15T10:30:00Z
- Times are compared in UTC
- modified_after: Include files modified after this time
- modified_before: Include files modified before this time

State Configuration:
- file_list_show_hidden: Show hidden files (starting with .)
- file_list_default_sort: Default sort field
- file_list_max_results: Default max results
- file_restricted_paths: Array of restricted paths
- file_allowed_paths: Array of allowed path prefixes

Sorting:
- name: Alphabetical by filename (case-insensitive)
- size: By file size (smallest first)
- modified: By modification time (oldest first)
- Use sort_reverse: true to reverse order

Performance:
- Non-recursive listing is very fast
- Recursive searches may take time for large trees
- Progress events emitted every 100 items
- Context cancellation supported`).
		WithExamples([]domain.ToolExample{
			{
				Name:        "List current directory",
				Description: "List all files in current directory",
				Scenario:    "When you need to see what files are available",
				Input: map[string]interface{}{
					"path": ".",
				},
				Output: map[string]interface{}{
					"files": []map[string]interface{}{
						{"name": "README.md", "path": "./README.md", "is_dir": false, "size": 1024, "mode": "-rw-r--r--", "modified_time": "2024-01-15T10:00:00Z", "extension": "md"},
						{"name": "main.go", "path": "./main.go", "is_dir": false, "size": 2048, "mode": "-rw-r--r--", "modified_time": "2024-01-15T11:00:00Z", "extension": "go"},
						{"name": "go.mod", "path": "./go.mod", "is_dir": false, "size": 256, "mode": "-rw-r--r--", "modified_time": "2024-01-15T09:00:00Z", "extension": "mod"},
					},
					"total_count":  3,
					"filtered_out": 0,
					"search_path":  "/home/user/project",
				},
				Explanation: "Lists only files (not directories) in the current directory",
			},
			{
				Name:        "Find Go files recursively",
				Description: "Search for all Go source files",
				Scenario:    "When you need to find all code files in a project",
				Input: map[string]interface{}{
					"path":      ".",
					"pattern":   "*.go",
					"recursive": true,
				},
				Output: map[string]interface{}{
					"files": []map[string]interface{}{
						{"name": "main.go", "path": "./main.go", "is_dir": false, "size": 2048},
						{"name": "utils.go", "path": "./pkg/utils.go", "is_dir": false, "size": 1024},
						{"name": "test.go", "path": "./test/test.go", "is_dir": false, "size": 512},
					},
					"total_count":  10,
					"filtered_out": 7,
					"search_path":  "/home/user/project",
					"pattern":      "*.go",
				},
				Explanation: "Recursively finds all .go files, filtering out non-matching files",
			},
			{
				Name:        "List large files",
				Description: "Find files larger than 10MB",
				Scenario:    "When cleaning up disk space",
				Input: map[string]interface{}{
					"path":         "/home/user/downloads",
					"min_size":     10485760,
					"recursive":    true,
					"sort_by":      "size",
					"sort_reverse": true,
				},
				Output: map[string]interface{}{
					"files": []map[string]interface{}{
						{"name": "video.mp4", "size": 104857600},
						{"name": "backup.zip", "size": 52428800},
						{"name": "dataset.csv", "size": 20971520},
					},
					"total_count":  50,
					"filtered_out": 47,
					"search_path":  "/home/user/downloads",
				},
				Explanation: "Finds files >= 10MB, sorted by size descending",
			},
			{
				Name:        "Recent files",
				Description: "Find files modified in last 24 hours",
				Scenario:    "When looking for recently changed files",
				Input: map[string]interface{}{
					"path":           ".",
					"modified_after": "2024-01-14T10:30:00Z",
					"recursive":      true,
					"sort_by":        "modified",
					"sort_reverse":   true,
				},
				Output: map[string]interface{}{
					"files": []map[string]interface{}{
						{"name": "report.pdf", "modified_time": "2024-01-15T15:30:00Z"},
						{"name": "data.json", "modified_time": "2024-01-15T14:00:00Z"},
						{"name": "notes.txt", "modified_time": "2024-01-15T12:00:00Z"},
					},
					"total_count":  100,
					"filtered_out": 97,
					"search_path":  "/home/user/project",
				},
				Explanation: "Shows files modified after the specified time, newest first",
			},
			{
				Name:        "List directories only",
				Description: "Show only subdirectories",
				Scenario:    "When exploring project structure",
				Input: map[string]interface{}{
					"path":          ".",
					"include_dirs":  true,
					"include_files": false,
				},
				Output: map[string]interface{}{
					"files": []map[string]interface{}{
						{"name": "src", "path": "./src", "is_dir": true, "size": 0},
						{"name": "test", "path": "./test", "is_dir": true, "size": 0},
						{"name": "docs", "path": "./docs", "is_dir": true, "size": 0},
					},
					"total_count":  5,
					"filtered_out": 2,
					"search_path":  "/home/user/project",
				},
				Explanation: "Shows only directories, not files",
			},
			{
				Name:        "Image files by extension",
				Description: "Find all image files",
				Scenario:    "When organizing media files",
				Input: map[string]interface{}{
					"path":        "/home/user/pictures",
					"pattern":     "*.{jpg,jpeg,png,gif}",
					"recursive":   true,
					"max_results": 100,
				},
				Output: map[string]interface{}{
					"files": []map[string]interface{}{
						{"name": "photo1.jpg", "extension": "jpg"},
						{"name": "screenshot.png", "extension": "png"},
						{"name": "animation.gif", "extension": "gif"},
					},
					"total_count":  500,
					"filtered_out": 100,
					"search_path":  "/home/user/pictures",
					"pattern":      "*.{jpg,jpeg,png,gif}",
				},
				Explanation: "Pattern matches multiple image extensions, limited to 100 results",
			},
			{
				Name:        "Complex filter",
				Description: "Recent large Python files",
				Scenario:    "When looking for specific files with multiple criteria",
				Input: map[string]interface{}{
					"path":           ".",
					"pattern":        "*.py",
					"min_size":       1024,
					"modified_after": "2024-01-01T00:00:00Z",
					"recursive":      true,
					"sort_by":        "size",
					"sort_reverse":   true,
				},
				Output: map[string]interface{}{
					"files": []map[string]interface{}{
						{"name": "main.py", "size": 8192, "modified_time": "2024-01-10T10:00:00Z"},
						{"name": "utils.py", "size": 4096, "modified_time": "2024-01-05T10:00:00Z"},
					},
					"total_count":  50,
					"filtered_out": 48,
				},
				Explanation: "Combines pattern, size, and date filters with sorting",
			},
		}).
		WithConstraints([]string{
			"Path must be a directory, not a file",
			"Pattern matching uses glob syntax, not regex",
			"Hidden files (starting with .) excluded by default",
			"Size filters only apply to files, not directories",
			"Recursive searches may be slow for large directory trees",
			"Max results limit is applied after sorting",
			"Symlinks are followed (be careful with circular links)",
			"Cannot access files without read permissions",
			"Times are compared in UTC",
			"Context cancellation stops enumeration immediately",
		}).
		WithErrorGuidance(map[string]string{
			"path not found":               "Check if the directory exists and path is correct",
			"path is not a directory":      "The path points to a file, not a directory",
			"permission denied":            "Insufficient permissions to read directory",
			"access denied":                "Path is restricted by security policy",
			"invalid modified_after time":  "Use RFC3339 format: 2024-01-15T10:30:00Z",
			"invalid modified_before time": "Use RFC3339 format: 2024-01-15T10:30:00Z",
			"invalid path":                 "Path contains invalid characters or is malformed",
			"too many open files":          "System file handle limit reached",
		}).
		WithCategory("file").
		WithTags([]string{"filesystem", "directory", "list", "search"}).
		WithVersion("2.0.0").
		WithBehavior(
			false,    // Not deterministic - directory contents can change
			false,    // Not destructive - only reads
			false,    // No confirmation needed
			"medium", // Can be slow for large/recursive searches
		)

	return builder.Build()
}

// sortFiles sorts the file list based on the specified criteria
func sortFiles(files []FileInfo, sortBy string, reverse bool) {
	var sortFunc func(i, j int) bool

	switch sortBy {
	case "size":
		sortFunc = func(i, j int) bool {
			if files[i].Size == files[j].Size {
				return files[i].Name < files[j].Name
			}
			return files[i].Size < files[j].Size
		}
	case "modified":
		sortFunc = func(i, j int) bool {
			if files[i].ModifiedTime.Equal(files[j].ModifiedTime) {
				return files[i].Name < files[j].Name
			}
			return files[i].ModifiedTime.Before(files[j].ModifiedTime)
		}
	default: // "name"
		sortFunc = func(i, j int) bool {
			return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
		}
	}

	if reverse {
		sort.Slice(files, func(i, j int) bool {
			return !sortFunc(i, j)
		})
	} else {
		sort.Slice(files, sortFunc)
	}
}

// MustGetFileList retrieves the registered FileList tool or panics
// This is a convenience function for users who want to ensure the tool exists
func MustGetFileList() domain.Tool {
	return tools.MustGetTool("file_list")
}
