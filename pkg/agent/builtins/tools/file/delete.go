// ABOUTME: File deletion tool with safety checks and confirmation
// ABOUTME: Built-in tool supporting safe file and directory removal with options

package file

import (
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

// fileDeleteOutputSchema defines the output schema for the FileDelete tool
var fileDeleteOutputSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"path": {
			Type:        "string",
			Description: "Absolute path that was processed",
		},
		"deleted": {
			Type:        "boolean",
			Description: "Whether the deletion was successful",
		},
		"was_directory": {
			Type:        "boolean",
			Description: "Whether the deleted item was a directory",
		},
		"message": {
			Type:        "string",
			Description: "Status message or reason for failure",
		},
	},
	Required: []string{"path", "deleted", "was_directory"},
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

// fileDeleteMain is the main function for the tool
func fileDeleteMain(ctx *domain.ToolContext, params FileDeleteParams) (*FileDeleteResult, error) {
	// Emit start event
	if ctx.Events != nil {
		ctx.Events.EmitMessage(fmt.Sprintf("Starting file deletion for %s", params.Path))
	}

	// Clean and resolve the path
	targetPath := filepath.Clean(params.Path)
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		if ctx.Events != nil {
			ctx.Events.EmitError(fmt.Errorf("invalid path: %w", err))
		}
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	// Check file access restrictions from state
	if ctx.State != nil {
		// Check restricted paths
		if restrictedPaths, exists := ctx.State.Get("file_restricted_paths"); exists {
			if paths, ok := restrictedPaths.([]string); ok {
				for _, restricted := range paths {
					if strings.HasPrefix(absPath, restricted) {
						errMsg := fmt.Errorf("access denied: path %s is restricted", absPath)
						if ctx.Events != nil {
							ctx.Events.EmitError(errMsg)
						}
						return nil, errMsg
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
					errMsg := fmt.Errorf("access denied: path %s is not in allowed paths", absPath)
					if ctx.Events != nil {
						ctx.Events.EmitError(errMsg)
					}
					return nil, errMsg
				}
			}
		}

		// Check if trash/recycle bin is preferred over permanent deletion
		if useTrash, exists := ctx.State.Get("file_use_trash"); exists {
			if trash, ok := useTrash.(bool); ok && trash && !params.Force {
				// Note: Actual trash implementation would require OS-specific code
				// This is a placeholder for the concept
				if ctx.Events != nil {
					ctx.Events.EmitMessage("Note: Trash/recycle bin requested but performing permanent deletion")
				}
			}
		}
	}

	// Emit checking file event
	if ctx.Events != nil {
		ctx.Events.EmitProgress(1, 4, "Checking file existence and permissions")
	}

	// Safety check: prevent deletion of critical system directories
	if isCriticalPath(absPath) && !params.Force {
		result := &FileDeleteResult{
			Path:    absPath,
			Deleted: false,
			Message: "Cannot delete critical system directory. Use force=true to override (dangerous!)",
		}
		if ctx.Events != nil {
			ctx.Events.EmitCustom("deletion_blocked", map[string]interface{}{
				"path":   absPath,
				"reason": "critical_system_path",
			})
		}
		return result, nil
	}

	// Check if path exists
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			result := &FileDeleteResult{
				Path:    absPath,
				Deleted: false,
				Message: "Path does not exist",
			}
			if ctx.Events != nil {
				ctx.Events.EmitCustom("deletion_skipped", map[string]interface{}{
					"path":   absPath,
					"reason": "not_found",
				})
			}
			return result, nil
		}
		if ctx.Events != nil {
			ctx.Events.EmitError(fmt.Errorf("error accessing path: %w", err))
		}
		return nil, fmt.Errorf("error accessing path: %w", err)
	}

	isDir := info.IsDir()

	// Emit progress for confirmation check
	if ctx.Events != nil {
		ctx.Events.EmitProgress(2, 4, "Validating deletion requirements")
	}

	// Check for confirmation requirement from state
	requireConfirmFromState := false
	if ctx.State != nil {
		if reqConfirm, exists := ctx.State.Get("file_require_delete_confirmation"); exists {
			if confirm, ok := reqConfirm.(bool); ok {
				requireConfirmFromState = confirm
			}
		}
	}

	// Confirmation check
	if params.RequireConfirm != "" || requireConfirmFromState {
		if params.RequireConfirm == "" && requireConfirmFromState {
			result := &FileDeleteResult{
				Path:         absPath,
				Deleted:      false,
				WasDirectory: isDir,
				Message:      "Deletion confirmation required by configuration but not provided",
			}
			if ctx.Events != nil {
				ctx.Events.EmitCustom("deletion_blocked", map[string]interface{}{
					"path":   absPath,
					"reason": "confirmation_required",
				})
			}
			return result, nil
		}

		// The confirmation must match either the full path or just the base name
		if params.RequireConfirm != "" && params.RequireConfirm != absPath && params.RequireConfirm != filepath.Base(absPath) {
			result := &FileDeleteResult{
				Path:         absPath,
				Deleted:      false,
				WasDirectory: isDir,
				Message:      fmt.Sprintf("Confirmation mismatch: expected '%s' or '%s', got '%s'", absPath, filepath.Base(absPath), params.RequireConfirm),
			}
			if ctx.Events != nil {
				ctx.Events.EmitCustom("deletion_blocked", map[string]interface{}{
					"path":   absPath,
					"reason": "confirmation_mismatch",
				})
			}
			return result, nil
		}
	}

	// Emit progress for deletion
	if ctx.Events != nil {
		ctx.Events.EmitProgress(3, 4, fmt.Sprintf("Deleting %s", absPath))
	}

	// Handle directory deletion
	if isDir {
		// Check if directory is empty
		entries, err := os.ReadDir(absPath)
		if err != nil {
			if ctx.Events != nil {
				ctx.Events.EmitError(fmt.Errorf("error reading directory: %w", err))
			}
			return nil, fmt.Errorf("error reading directory: %w", err)
		}

		if len(entries) > 0 && !params.Recursive {
			result := &FileDeleteResult{
				Path:         absPath,
				Deleted:      false,
				WasDirectory: true,
				Message:      fmt.Sprintf("Directory is not empty (%d items). Use recursive=true to delete contents", len(entries)),
			}
			if ctx.Events != nil {
				ctx.Events.EmitCustom("deletion_blocked", map[string]interface{}{
					"path":       absPath,
					"reason":     "directory_not_empty",
					"item_count": len(entries),
				})
			}
			return result, nil
		}

		// Additional safety for non-empty directories
		if len(entries) > 0 && !params.Force && params.RequireConfirm == "" {
			result := &FileDeleteResult{
				Path:         absPath,
				Deleted:      false,
				WasDirectory: true,
				Message:      "Non-empty directory deletion requires either force=true or require_confirm parameter",
			}
			if ctx.Events != nil {
				ctx.Events.EmitCustom("deletion_blocked", map[string]interface{}{
					"path":       absPath,
					"reason":     "safety_check_failed",
					"item_count": len(entries),
				})
			}
			return result, nil
		}

		// Delete directory
		if params.Recursive && len(entries) > 0 {
			err = os.RemoveAll(absPath)
		} else {
			err = os.Remove(absPath)
		}

		if err != nil {
			if ctx.Events != nil {
				ctx.Events.EmitError(fmt.Errorf("error deleting directory: %w", err))
			}
			return nil, fmt.Errorf("error deleting directory: %w", err)
		}

		result := &FileDeleteResult{
			Path:         absPath,
			Deleted:      true,
			WasDirectory: true,
			Message:      fmt.Sprintf("Directory deleted successfully (%d items removed)", len(entries)),
		}

		// Emit completion event with details
		if ctx.Events != nil {
			ctx.Events.EmitProgress(4, 4, "Deletion complete")
			ctx.Events.EmitCustom("deletion_complete", map[string]interface{}{
				"path":          absPath,
				"type":          "directory",
				"items_removed": len(entries),
				"recursive":     params.Recursive,
				"elapsed_ms":    ctx.ElapsedTime().Milliseconds(),
			})
		}

		return result, nil
	}

	// Get file size before deletion for events
	fileSize := info.Size()

	// Delete file
	err = os.Remove(absPath)
	if err != nil {
		if ctx.Events != nil {
			ctx.Events.EmitError(fmt.Errorf("error deleting file: %w", err))
		}
		return nil, fmt.Errorf("error deleting file: %w", err)
	}

	result := &FileDeleteResult{
		Path:         absPath,
		Deleted:      true,
		WasDirectory: false,
		Message:      "File deleted successfully",
	}

	// Emit completion event with details
	if ctx.Events != nil {
		ctx.Events.EmitProgress(4, 4, "Deletion complete")
		ctx.Events.EmitCustom("deletion_complete", map[string]interface{}{
			"path":       absPath,
			"type":       "file",
			"size_bytes": fileSize,
			"elapsed_ms": ctx.ElapsedTime().Milliseconds(),
		})
	}

	return result, nil
}

// FileDelete creates a tool for safely deleting files and directories with multiple safety mechanisms.
// It supports confirmation requirements, recursive directory deletion, and protection of critical system paths.
// The tool provides detailed events and progress tracking, making it suitable for both interactive and automated use.
// Safety features include path restrictions, confirmation strings, and protection against accidental deletion of important files.
func FileDelete() domain.Tool {
	builder := atools.NewToolBuilder("file_delete", "Safely deletes files and directories with confirmation options").
		WithFunction(fileDeleteMain).
		WithParameterSchema(fileDeleteParamSchema).
		WithOutputSchema(fileDeleteOutputSchema).
		WithUsageInstructions(`Use this tool to safely delete files and directories with multiple confirmation options.

IMPORTANT: This is a DESTRUCTIVE operation that permanently removes data.

Features:
- Multiple safety mechanisms to prevent accidental deletion
- Support for both files and directories
- Recursive directory deletion with safeguards
- Critical system path protection
- Confirmation requirements for dangerous operations
- Path access control via state configuration

Parameters:
- path: File or directory to delete (required)
- force: Skip safety checks (dangerous!)
- recursive: Delete directories and all contents
- require_confirm: Must match the path for deletion to proceed

Safety Mechanisms:
1. Critical system paths are protected by default
2. Non-empty directories require recursive=true
3. Confirmation can be required via require_confirm
4. State configuration can enforce confirmation
5. Path restrictions via allowed/restricted lists

Confirmation Requirements:
- Set require_confirm to the exact path or filename
- State can enforce confirmation via file_require_delete_confirmation
- Non-empty directories need force=true or confirmation

Critical Paths Protected:
- Root directories (/, C:\)
- System directories (/bin, /etc, C:\Windows)
- User home directory
- Program directories

State Configuration:
- file_restricted_paths: Array of paths that cannot be deleted
- file_allowed_paths: Array of allowed path prefixes
- file_require_delete_confirmation: Always require confirmation
- file_use_trash: Prefer trash/recycle bin (note: not implemented)

Best Practices:
- Always use require_confirm for important deletions
- Test with a dry run first (check path without deleting)
- Use recursive=true only when necessary
- Avoid force=true unless absolutely certain
- Consider backups before deletion

Error Handling:
- Non-existent paths return success with deleted=false
- Permission errors are reported clearly
- Directory not empty errors suggest recursive=true`).
		WithExamples([]domain.ToolExample{
			{
				Name:        "Delete a simple file",
				Description: "Remove a temporary file",
				Scenario:    "When cleaning up temporary files",
				Input: map[string]interface{}{
					"path": "/tmp/temp_file.txt",
				},
				Output: map[string]interface{}{
					"path":          "/tmp/temp_file.txt",
					"deleted":       true,
					"was_directory": false,
					"message":       "File deleted successfully",
				},
				Explanation: "Simple file deletion without confirmation",
			},
			{
				Name:        "Delete with confirmation",
				Description: "Delete important file with confirmation",
				Scenario:    "When deleting important files",
				Input: map[string]interface{}{
					"path":            "/home/user/important.db",
					"require_confirm": "important.db",
				},
				Output: map[string]interface{}{
					"path":          "/home/user/important.db",
					"deleted":       true,
					"was_directory": false,
					"message":       "File deleted successfully",
				},
				Explanation: "Confirmation string matches filename, deletion proceeds",
			},
			{
				Name:        "Delete empty directory",
				Description: "Remove an empty directory",
				Scenario:    "When cleaning up empty folders",
				Input: map[string]interface{}{
					"path": "/tmp/empty_dir",
				},
				Output: map[string]interface{}{
					"path":          "/tmp/empty_dir",
					"deleted":       true,
					"was_directory": true,
					"message":       "Directory deleted successfully (0 items removed)",
				},
				Explanation: "Empty directories can be deleted without recursive flag",
			},
			{
				Name:        "Delete directory with contents",
				Description: "Recursively delete directory and contents",
				Scenario:    "When removing project directories",
				Input: map[string]interface{}{
					"path":            "/tmp/old_project",
					"recursive":       true,
					"require_confirm": "old_project",
				},
				Output: map[string]interface{}{
					"path":          "/tmp/old_project",
					"deleted":       true,
					"was_directory": true,
					"message":       "Directory deleted successfully (42 items removed)",
				},
				Explanation: "Recursive deletion with confirmation for safety",
			},
			{
				Name:        "Failed confirmation",
				Description: "Deletion blocked by wrong confirmation",
				Scenario:    "When confirmation doesn't match",
				Input: map[string]interface{}{
					"path":            "/home/user/data.db",
					"require_confirm": "wrong_name",
				},
				Output: map[string]interface{}{
					"path":          "/home/user/data.db",
					"deleted":       false,
					"was_directory": false,
					"message":       "Confirmation mismatch: expected '/home/user/data.db' or 'data.db', got 'wrong_name'",
				},
				Explanation: "Protects against typos in confirmation",
			},
			{
				Name:        "Non-empty directory without recursive",
				Description: "Attempt to delete non-empty directory",
				Scenario:    "When forgetting recursive flag",
				Input: map[string]interface{}{
					"path": "/tmp/full_dir",
				},
				Output: map[string]interface{}{
					"path":          "/tmp/full_dir",
					"deleted":       false,
					"was_directory": true,
					"message":       "Directory is not empty (15 items). Use recursive=true to delete contents",
				},
				Explanation: "Prevents accidental deletion of directory contents",
			},
			{
				Name:        "Critical path protection",
				Description: "Attempt to delete system directory",
				Scenario:    "When trying to delete protected paths",
				Input: map[string]interface{}{
					"path": "/etc",
				},
				Output: map[string]interface{}{
					"path":          "/etc",
					"deleted":       false,
					"was_directory": true,
					"message":       "Cannot delete critical system directory. Use force=true to override (dangerous!)",
				},
				Explanation: "System directories are protected from accidental deletion",
			},
		}).
		WithConstraints([]string{
			"Deletion is permanent - files cannot be recovered",
			"Critical system paths are protected by default",
			"Non-empty directories require recursive=true",
			"Confirmation string must match path or filename exactly",
			"Force flag bypasses most safety checks - use with extreme caution",
			"Symlinks are not followed - only the link is deleted",
			"No undo functionality - consider backups first",
			"Cross-platform path handling may vary",
			"Concurrent deletion operations should be avoided",
			"Large recursive deletions may take significant time",
		}).
		WithErrorGuidance(map[string]string{
			"permission denied":       "Check file permissions and ownership. May need elevated privileges",
			"access denied":           "Path is restricted by security policy. Check allowed paths",
			"directory not empty":     "Use recursive=true to delete directory contents",
			"confirmation required":   "Set require_confirm parameter to match the path",
			"confirmation mismatch":   "The require_confirm value must exactly match the path or filename",
			"file in use":             "Close any programs using the file before deletion",
			"read-only file system":   "Cannot delete files on read-only media",
			"invalid path":            "Check path syntax and special characters",
			"device or resource busy": "File or directory is being used by another process",
			"operation not permitted": "File may have special attributes preventing deletion",
		}).
		WithCategory("file").
		WithTags([]string{"filesystem", "delete", "remove", "cleanup"}).
		WithVersion("2.0.0").
		WithBehavior(
			true,   // Deterministic - same path always deletes same file
			true,   // Destructive - permanently removes data
			true,   // Requires confirmation for safety
			"fast", // Usually fast, can be slow for large directories
		)

	return builder.Build()
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
