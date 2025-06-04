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
		func(ctx *domain.ToolContext, params FileDeleteParams) (*FileDeleteResult, error) {
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
