// ABOUTME: File deletion tool with safety checks and confirmation
// ABOUTME: Built-in tool supporting safe file and directory removal with options

package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	atools "github.com/lexlapax/go-llms/pkg/agent/tools"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// FileDeleteParams defines parameters for the FileDelete tool
type FileDeleteParams struct {
	Path           string `json:"path"`
	Force          bool   `json:"force,omitempty"`           // Skip confirmation for non-empty directories
	Recursive      bool   `json:"recursive,omitempty"`       // Delete directories and their contents
	RequireConfirm string `json:"require_confirm,omitempty"` // Confirmation string that must match path
}

// FileDeleteResult defines the result of the FileDelete tool
type FileDeleteResult struct {
	Path         string `json:"path"`
	Deleted      bool   `json:"deleted"`
	WasDirectory bool   `json:"was_directory"`
	Message      string `json:"message,omitempty"`
}

// fileDeleteParamSchema defines parameters for the FileDelete tool
var fileDeleteParamSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"path": {
			Type:        "string",
			Description: "Path to the file or directory to delete",
		},
		"force": {
			Type:        "boolean",
			Description: "Force deletion without safety checks (use with caution)",
		},
		"recursive": {
			Type:        "boolean",
			Description: "Delete directories and all their contents",
		},
		"require_confirm": {
			Type:        "string",
			Description: "Safety confirmation - must match the path being deleted",
		},
	},
	Required: []string{"path"},
}

// init automatically registers the tool on package import
func init() {
	tools.MustRegisterTool("file_delete", FileDelete(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "file_delete",
			Category:    "file",
			Tags:        []string{"filesystem", "delete", "remove", "cleanup"},
			Description: "Safely deletes files and directories with confirmation options",
			Version:     "1.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Delete a file",
					Description: "Delete a single file",
					Code:        `FileDelete().Execute(ctx, FileDeleteParams{Path: "/tmp/old.txt"})`,
				},
				{
					Name:        "Delete empty directory",
					Description: "Delete an empty directory",
					Code:        `FileDelete().Execute(ctx, FileDeleteParams{Path: "/tmp/empty_dir"})`,
				},
				{
					Name:        "Delete directory with contents",
					Description: "Recursively delete a directory and all contents",
					Code:        `FileDelete().Execute(ctx, FileDeleteParams{Path: "/tmp/full_dir", Recursive: true, RequireConfirm: "/tmp/full_dir"})`,
				},
			},
		},
		RequiredPermissions: []string{"filesystem:write", "filesystem:delete"},
		ResourceUsage: tools.ResourceInfo{
			Memory:      "low",
			Network:     false,
			FileSystem:  true,
			Concurrency: false, // File deletion should not be concurrent
		},
	})
}

// FileDelete creates a tool for safely deleting files and directories
// This is a built-in tool optimized for:
// - Safe deletion with multiple confirmation options
// - Protection against accidental deletions
// - Clear feedback about what was deleted
// - Support for both files and directories
func FileDelete() domain.Tool {
	return atools.NewTool(
		"file_delete",
		"Safely deletes files and directories with confirmation options",
		func(ctx context.Context, params FileDeleteParams) (*FileDeleteResult, error) {
			// Clean and resolve the path
			targetPath := filepath.Clean(params.Path)
			absPath, err := filepath.Abs(targetPath)
			if err != nil {
				return nil, fmt.Errorf("invalid path: %w", err)
			}

			// Safety check: prevent deletion of critical system directories
			if isCriticalPath(absPath) && !params.Force {
				return &FileDeleteResult{
					Path:    absPath,
					Deleted: false,
					Message: "Cannot delete critical system directory. Use force=true to override (dangerous!)",
				}, nil
			}

			// Check if path exists
			info, err := os.Stat(absPath)
			if err != nil {
				if os.IsNotExist(err) {
					return &FileDeleteResult{
						Path:    absPath,
						Deleted: false,
						Message: "Path does not exist",
					}, nil
				}
				return nil, fmt.Errorf("error accessing path: %w", err)
			}

			isDir := info.IsDir()

			// Confirmation check
			if params.RequireConfirm != "" {
				// The confirmation must match either the full path or just the base name
				if params.RequireConfirm != absPath && params.RequireConfirm != filepath.Base(absPath) {
					return &FileDeleteResult{
						Path:         absPath,
						Deleted:      false,
						WasDirectory: isDir,
						Message:      fmt.Sprintf("Confirmation mismatch: expected '%s' or '%s', got '%s'", absPath, filepath.Base(absPath), params.RequireConfirm),
					}, nil
				}
			}

			// Handle directory deletion
			if isDir {
				// Check if directory is empty
				entries, err := os.ReadDir(absPath)
				if err != nil {
					return nil, fmt.Errorf("error reading directory: %w", err)
				}

				if len(entries) > 0 && !params.Recursive {
					return &FileDeleteResult{
						Path:         absPath,
						Deleted:      false,
						WasDirectory: true,
						Message:      fmt.Sprintf("Directory is not empty (%d items). Use recursive=true to delete contents", len(entries)),
					}, nil
				}

				// Additional safety for non-empty directories
				if len(entries) > 0 && !params.Force && params.RequireConfirm == "" {
					return &FileDeleteResult{
						Path:         absPath,
						Deleted:      false,
						WasDirectory: true,
						Message:      "Non-empty directory deletion requires either force=true or require_confirm parameter",
					}, nil
				}

				// Delete directory
				if params.Recursive && len(entries) > 0 {
					err = os.RemoveAll(absPath)
				} else {
					err = os.Remove(absPath)
				}

				if err != nil {
					return nil, fmt.Errorf("error deleting directory: %w", err)
				}

				return &FileDeleteResult{
					Path:         absPath,
					Deleted:      true,
					WasDirectory: true,
					Message:      fmt.Sprintf("Directory deleted successfully (%d items removed)", len(entries)),
				}, nil
			}

			// Delete file
			err = os.Remove(absPath)
			if err != nil {
				return nil, fmt.Errorf("error deleting file: %w", err)
			}

			return &FileDeleteResult{
				Path:         absPath,
				Deleted:      true,
				WasDirectory: false,
				Message:      "File deleted successfully",
			}, nil
		},
		fileDeleteParamSchema,
	)
}

// isCriticalPath checks if a path is a critical system directory
func isCriticalPath(path string) bool {
	// Normalize path for comparison
	path = filepath.Clean(path)

	// List of critical paths that should not be deleted
	criticalPaths := []string{
		"/",
		"/bin",
		"/boot",
		"/dev",
		"/etc",
		"/home",
		"/lib",
		"/lib64",
		"/opt",
		"/proc",
		"/root",
		"/sbin",
		"/sys",
		"/usr",
		"/var",
		// Windows critical paths
		"C:\\",
		"C:\\Windows",
		"C:\\Program Files",
		"C:\\Program Files (x86)",
		"C:\\ProgramData",
		"C:\\Users",
	}

	// Also protect user home directory
	if home, err := os.UserHomeDir(); err == nil {
		criticalPaths = append(criticalPaths, home)
	}

	for _, critical := range criticalPaths {
		if strings.EqualFold(path, critical) {
			return true
		}
	}

	return false
}

// MustGetFileDelete retrieves the registered FileDelete tool or panics
// This is a convenience function for users who want to ensure the tool exists
func MustGetFileDelete() domain.Tool {
	return tools.MustGetTool("file_delete")
}
