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
	Path         string `json:"path"`
	Pattern      string `json:"pattern,omitempty"`      // glob pattern like "*.txt"
	Recursive    bool   `json:"recursive,omitempty"`    // search subdirectories
	IncludeDirs  bool   `json:"include_dirs,omitempty"` // include directories in results
	IncludeFiles bool   `json:"include_files,omitempty"` // include files in results (default: true)
	MinSize      int64  `json:"min_size,omitempty"`      // minimum file size in bytes
	MaxSize      int64  `json:"max_size,omitempty"`      // maximum file size in bytes
	ModifiedAfter  string `json:"modified_after,omitempty"`  // RFC3339 timestamp
	ModifiedBefore string `json:"modified_before,omitempty"` // RFC3339 timestamp
	SortBy       string `json:"sort_by,omitempty"`       // name, size, modified
	SortReverse  bool   `json:"sort_reverse,omitempty"`  // reverse sort order
	MaxResults   int    `json:"max_results,omitempty"`   // limit number of results
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

// FileList creates a tool for listing files and directories
// This is a built-in tool optimized for:
// - Fast directory enumeration
// - Flexible filtering options
// - Pattern matching support
// - Size and date filtering
// - Sorting capabilities
func FileList() domain.Tool {
	return atools.NewTool(
		"file_list",
		"Lists files and directories with filtering options",
		func(ctx context.Context, params FileListParams) (*FileListResult, error) {
			// Set defaults
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

			// Collect files
			var files []FileInfo
			var totalCount, filteredOut int

			walkFunc := func(path string, info os.FileInfo, err error) error {
				// Check context cancellation
				select {
				case <-ctx.Done():
					return ctx.Err()
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

				totalCount++

				// Apply filters
				isDir := info.IsDir()
				
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

			// Sort results
			sortFiles(files, params.SortBy, params.SortReverse)

			// Apply max results limit
			if params.MaxResults > 0 && len(files) > params.MaxResults {
				files = files[:params.MaxResults]
			}

			return &FileListResult{
				Files:       files,
				TotalCount:  totalCount,
				FilteredOut: filteredOut,
				SearchPath:  absPath,
				Pattern:     params.Pattern,
			}, nil
		},
		fileListParamSchema,
	)
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