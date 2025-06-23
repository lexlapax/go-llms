// ABOUTME: File move/rename tool with safety checks and overwrite options
// ABOUTME: Built-in tool supporting atomic moves, cross-device transfers, and directory operations

package file

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	atools "github.com/lexlapax/go-llms/pkg/agent/tools"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// FileMoveParams defines parameters for the FileMove tool
type FileMoveParams struct {
	Source        string `json:"source"`
	Destination   string `json:"destination"`
	Overwrite     bool   `json:"overwrite,omitempty"`      // Overwrite existing destination
	CreateDirs    bool   `json:"create_dirs,omitempty"`    // Create parent directories if needed
	PreserveAttrs bool   `json:"preserve_attrs,omitempty"` // Preserve file attributes (permissions, times)
}

// FileMoveResult defines the result of the FileMove tool
type FileMoveResult struct {
	Source         string `json:"source"`
	Destination    string `json:"destination"`
	Moved          bool   `json:"moved"`
	WasRename      bool   `json:"was_rename"`       // True if same directory (rename only)
	WasCrossDevice bool   `json:"was_cross_device"` // True if moved across filesystems
	Message        string `json:"message,omitempty"`
}

// fileMoveParamSchema defines parameters for the FileMove tool
var fileMoveParamSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"source": {
			Type:        "string",
			Description: "Source file or directory path",
		},
		"destination": {
			Type:        "string",
			Description: "Destination path (can be a new name or directory)",
		},
		"overwrite": {
			Type:        "boolean",
			Description: "Overwrite existing destination file",
		},
		"create_dirs": {
			Type:        "boolean",
			Description: "Create parent directories if they don't exist",
		},
		"preserve_attrs": {
			Type:        "boolean",
			Description: "Preserve file permissions and timestamps",
		},
	},
	Required: []string{"source", "destination"},
}

// fileMoveOutputSchema defines the output schema for the FileMove tool
var fileMoveOutputSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"source": {
			Type:        "string",
			Description: "Absolute source path that was moved",
		},
		"destination": {
			Type:        "string",
			Description: "Absolute destination path where file was moved",
		},
		"moved": {
			Type:        "boolean",
			Description: "Whether the move operation succeeded",
		},
		"was_rename": {
			Type:        "boolean",
			Description: "True if operation was a rename in same directory",
		},
		"was_cross_device": {
			Type:        "boolean",
			Description: "True if file was moved across filesystems",
		},
		"message": {
			Type:        "string",
			Description: "Status message or error description",
		},
	},
	Required: []string{"source", "destination", "moved", "was_rename", "was_cross_device"},
}

// init automatically registers the tool on package import
func init() {
	tools.MustRegisterTool("file_move", FileMove(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "file_move",
			Category:    "file",
			Tags:        []string{"filesystem", "move", "rename", "transfer"},
			Description: "Moves or renames files and directories",
			Version:     "1.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Rename a file",
					Description: "Rename a file in the same directory",
					Code:        `FileMove().Execute(ctx, FileMoveParams{Source: "old.txt", Destination: "new.txt"})`,
				},
				{
					Name:        "Move to directory",
					Description: "Move a file to a different directory",
					Code:        `FileMove().Execute(ctx, FileMoveParams{Source: "file.txt", Destination: "/backup/file.txt", CreateDirs: true})`,
				},
				{
					Name:        "Move with overwrite",
					Description: "Move and overwrite existing file",
					Code:        `FileMove().Execute(ctx, FileMoveParams{Source: "update.txt", Destination: "current.txt", Overwrite: true})`,
				},
			},
		},
		RequiredPermissions: []string{"filesystem:read", "filesystem:write"},
		ResourceUsage: tools.ResourceInfo{
			Memory:      "low",
			Network:     false,
			FileSystem:  true,
			Concurrency: false, // File operations should be sequential
		},
	})
}

// fileMoveMain is the main function for the tool
func fileMoveMain(ctx *domain.ToolContext, params FileMoveParams) (*FileMoveResult, error) {
	// Emit start event
	if ctx.Events != nil {
		ctx.Events.EmitCustom("file_move.start", map[string]interface{}{
			"source":      params.Source,
			"destination": params.Destination,
		})
	}

	// Clean and resolve paths
	srcPath := filepath.Clean(params.Source)
	dstPath := filepath.Clean(params.Destination)

	absSrc, err := filepath.Abs(srcPath)
	if err != nil {
		return nil, fmt.Errorf("invalid source path: %w", err)
	}

	absDst, err := filepath.Abs(dstPath)
	if err != nil {
		return nil, fmt.Errorf("invalid destination path: %w", err)
	}

	// Check for restricted paths from state
	if restrictedPathsVal, ok := ctx.State.Get("file.restricted_paths"); ok {
		if restrictedPaths, ok := restrictedPathsVal.([]interface{}); ok && len(restrictedPaths) > 0 {
			for _, restricted := range restrictedPaths {
				if restrictedStr, ok := restricted.(string); ok {
					// Check if source or destination is within restricted path
					if isSubPath(absSrc, restrictedStr) {
						return nil, fmt.Errorf("source path %s is within restricted path %s", absSrc, restrictedStr)
					}
					// Calculate final destination path if destination is a directory
					finalDstPath := absDst
					if dstInfo, err := os.Stat(absDst); err == nil && dstInfo.IsDir() {
						finalDstPath = filepath.Join(absDst, filepath.Base(absSrc))
					}
					if isSubPath(absDst, restrictedStr) || isSubPath(finalDstPath, restrictedStr) {
						return nil, fmt.Errorf("destination path %s is within restricted path %s", absDst, restrictedStr)
					}
				}
			}
		}
	}

	// Emit checking source event
	if ctx.Events != nil {
		ctx.Events.EmitCustom("file_move.checking_source", map[string]interface{}{
			"path": absSrc,
		})
	}

	// Check if source exists
	srcInfo, err := os.Stat(absSrc)
	if err != nil {
		if os.IsNotExist(err) {
			return &FileMoveResult{
				Source:      absSrc,
				Destination: absDst,
				Moved:       false,
				Message:     "Source file does not exist",
			}, nil
		}
		return nil, fmt.Errorf("error accessing source: %w", err)
	}

	// Determine if destination is a directory or file
	dstInfo, dstErr := os.Stat(absDst)
	isDstDir := dstErr == nil && dstInfo.IsDir()

	// If destination is a directory, append source filename
	finalDst := absDst
	if isDstDir {
		finalDst = filepath.Join(absDst, filepath.Base(absSrc))
	}

	// Check for same source and destination
	if absSrc == finalDst {
		return &FileMoveResult{
			Source:      absSrc,
			Destination: finalDst,
			Moved:       false,
			Message:     "Source and destination are the same",
		}, nil
	}

	// Check if it's just a rename (same directory)
	isRename := filepath.Dir(absSrc) == filepath.Dir(finalDst)

	// Emit checking destination event
	if ctx.Events != nil {
		ctx.Events.EmitCustom("file_move.checking_destination", map[string]interface{}{
			"path":      finalDst,
			"exists":    dstErr == nil,
			"is_rename": isRename,
		})
	}

	// Get overwrite permission from state if not explicitly set
	overwrite := params.Overwrite
	if !overwrite {
		if allowOverwrite, ok := ctx.State.Get("file.allow_overwrite"); ok {
			if boolVal, ok := allowOverwrite.(bool); ok && boolVal {
				overwrite = true
			}
		}
	}

	// Check if destination already exists
	if _, err := os.Stat(finalDst); err == nil && !overwrite {
		return &FileMoveResult{
			Source:      absSrc,
			Destination: finalDst,
			Moved:       false,
			WasRename:   isRename,
			Message:     "Destination already exists. Use overwrite=true to replace",
		}, nil
	}

	// Create parent directories if requested
	if params.CreateDirs {
		parentDir := filepath.Dir(finalDst)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create parent directories: %w", err)
		}
	}

	// Check if we should prefer copy-then-delete from state
	preferCopyDelete := false
	if preferCopy, ok := ctx.State.Get("file.prefer_copy_delete"); ok {
		if boolVal, ok := preferCopy.(bool); ok {
			preferCopyDelete = boolVal
		}
	}

	// Emit moving event
	if ctx.Events != nil {
		ctx.Events.EmitCustom("file_move.moving", map[string]interface{}{
			"method": "rename",
		})
	}

	// Try atomic rename first (unless prefer_copy_delete is set)
	var moveErr error
	wasCrossDevice := false

	if !preferCopyDelete {
		moveErr = os.Rename(absSrc, finalDst)
		if moveErr == nil {
			// Successful atomic move
			result := &FileMoveResult{
				Source:         absSrc,
				Destination:    finalDst,
				Moved:          true,
				WasRename:      isRename,
				WasCrossDevice: false,
				Message:        "Successfully moved",
			}

			// Emit completion event
			if ctx.Events != nil {
				ctx.Events.EmitCustom("file_move.completed", map[string]interface{}{
					"result":           result,
					"source":           absSrc,
					"destination":      finalDst,
					"was_rename":       isRename,
					"was_cross_device": false,
					"method":           "atomic_rename",
					"file_size":        srcInfo.Size(),
					"is_directory":     srcInfo.IsDir(),
				})
			}

			return result, nil
		}
	}

	// If rename failed or prefer_copy_delete is set, might need cross-device move
	// Only attempt cross-device move for files, not directories
	if srcInfo.IsDir() {
		return nil, fmt.Errorf("cannot move directory across devices: %w", moveErr)
	}

	// Emit copying event for cross-device move
	if ctx.Events != nil {
		ctx.Events.EmitCustom("file_move.copying", map[string]interface{}{
			"reason": "cross_device_or_preferred",
		})
	}

	// Perform cross-device file move (copy then delete)
	if err := crossDeviceMove(ctx.Context, absSrc, finalDst, srcInfo, params.PreserveAttrs); err != nil {
		return nil, fmt.Errorf("cross-device move failed: %w", err)
	}

	wasCrossDevice = true

	result := &FileMoveResult{
		Source:         absSrc,
		Destination:    finalDst,
		Moved:          true,
		WasRename:      false,
		WasCrossDevice: wasCrossDevice,
		Message:        "Successfully moved (cross-device)",
	}

	// Emit completion event
	if ctx.Events != nil {
		ctx.Events.EmitCustom("file_move.completed", map[string]interface{}{
			"result":           result,
			"source":           absSrc,
			"destination":      finalDst,
			"was_rename":       false,
			"was_cross_device": wasCrossDevice,
			"method":           "copy_then_delete",
			"file_size":        srcInfo.Size(),
			"is_directory":     srcInfo.IsDir(),
		})
	}

	return result, nil
}

// FileMove creates a tool for moving or renaming files and directories with atomic operations and cross-device support.
// It handles both simple renames within the same directory and complex moves across different filesystems.
// The tool provides safety features like overwrite protection, parent directory creation, and attribute preservation.
// Cross-device moves are automatically detected and handled using a copy-then-delete approach for files.
func FileMove() domain.Tool {
	builder := atools.NewToolBuilder("file_move", "Moves or renames files and directories").
		WithFunction(fileMoveMain).
		WithParameterSchema(fileMoveParamSchema).
		WithOutputSchema(fileMoveOutputSchema).
		WithUsageInstructions(`Use this tool to move or rename files and directories safely.

Features:
- Atomic moves within same filesystem (instant)
- Cross-device transfers (copy-then-delete)
- Directory and file support
- Overwrite protection with explicit flag
- Parent directory creation
- Attribute preservation
- Event tracking for operation progress

Parameters:
- source: Source file or directory path (required)
- destination: Target path or directory (required)
- overwrite: Allow overwriting existing files (optional, default false)
- create_dirs: Create parent directories if needed (optional)
- preserve_attrs: Keep file permissions and timestamps (optional)

Move Operations:
1. Rename: Same directory, different name
2. Move: Different directory, same or different name
3. Cross-device: Automatic copy-then-delete for different filesystems

Destination Behavior:
- If destination is a directory: moves source into it
- If destination is a file path: renames/moves to that path
- Parent directories must exist unless create_dirs is true

Safety Features:
- Won't overwrite without explicit permission
- Checks for same source and destination
- Validates paths before operation
- Restricted path checking via state

State Configuration:
- file.restricted_paths: Array of paths to block
- file.allow_overwrite: Default overwrite permission
- file.prefer_copy_delete: Force copy-delete method

Cross-Device Moves:
- Automatically detected when rename fails
- Only supported for files, not directories
- Preserves content, optionally preserves attributes
- Atomic within filesystem limits

Events Emitted:
- file_move.start: Operation beginning
- file_move.checking_source: Validating source
- file_move.checking_destination: Validating destination
- file_move.moving: Starting move operation
- file_move.copying: Cross-device copy in progress
- file_move.completed: Operation finished

Best Practices:
- Always check if destination exists first
- Use overwrite=true carefully
- Enable preserve_attrs for important files
- Create parent directories when organizing files
- Consider using rename for same-directory operations`).
		WithExamples([]domain.ToolExample{
			{
				Name:        "Simple rename",
				Description: "Rename a file in the same directory",
				Scenario:    "When changing a filename",
				Input: map[string]interface{}{
					"source":      "document.txt",
					"destination": "report.txt",
				},
				Output: map[string]interface{}{
					"source":           "/home/user/document.txt",
					"destination":      "/home/user/report.txt",
					"moved":            true,
					"was_rename":       true,
					"was_cross_device": false,
					"message":          "Successfully moved",
				},
				Explanation: "Atomic rename operation within same directory",
			},
			{
				Name:        "Move to directory",
				Description: "Move file to another directory",
				Scenario:    "When organizing files",
				Input: map[string]interface{}{
					"source":      "/tmp/download.pdf",
					"destination": "/home/user/documents/",
				},
				Output: map[string]interface{}{
					"source":           "/tmp/download.pdf",
					"destination":      "/home/user/documents/download.pdf",
					"moved":            true,
					"was_rename":       false,
					"was_cross_device": true,
					"message":          "Successfully moved (cross-device)",
				},
				Explanation: "File moved across filesystems using copy-then-delete",
			},
			{
				Name:        "Move with new name",
				Description: "Move and rename in one operation",
				Scenario:    "When relocating and renaming",
				Input: map[string]interface{}{
					"source":      "temp/draft.txt",
					"destination": "projects/final-report.txt",
					"create_dirs": true,
				},
				Output: map[string]interface{}{
					"source":           "/home/user/temp/draft.txt",
					"destination":      "/home/user/projects/final-report.txt",
					"moved":            true,
					"was_rename":       false,
					"was_cross_device": false,
					"message":          "Successfully moved",
				},
				Explanation: "Creates parent directory and moves with new name",
			},
			{
				Name:        "Overwrite existing file",
				Description: "Replace existing file with move",
				Scenario:    "When updating files",
				Input: map[string]interface{}{
					"source":      "new-config.json",
					"destination": "config.json",
					"overwrite":   true,
				},
				Output: map[string]interface{}{
					"source":           "/app/new-config.json",
					"destination":      "/app/config.json",
					"moved":            true,
					"was_rename":       true,
					"was_cross_device": false,
					"message":          "Successfully moved",
				},
				Explanation: "Overwrites existing config.json with new version",
			},
			{
				Name:        "Move directory",
				Description: "Relocate entire directory",
				Scenario:    "When reorganizing project structure",
				Input: map[string]interface{}{
					"source":      "old-location/project",
					"destination": "workspace/active-projects/",
				},
				Output: map[string]interface{}{
					"source":           "/home/user/old-location/project",
					"destination":      "/home/user/workspace/active-projects/project",
					"moved":            true,
					"was_rename":       false,
					"was_cross_device": false,
					"message":          "Successfully moved",
				},
				Explanation: "Moves entire directory tree to new location",
			},
			{
				Name:        "Failed move - exists",
				Description: "Destination already exists",
				Scenario:    "When destination is occupied",
				Input: map[string]interface{}{
					"source":      "update.txt",
					"destination": "existing.txt",
				},
				Output: map[string]interface{}{
					"source":           "/home/user/update.txt",
					"destination":      "/home/user/existing.txt",
					"moved":            false,
					"was_rename":       true,
					"was_cross_device": false,
					"message":          "Destination already exists. Use overwrite=true to replace",
				},
				Explanation: "Move blocked to prevent accidental data loss",
			},
			{
				Name:        "Preserve attributes",
				Description: "Move with timestamp preservation",
				Scenario:    "When file metadata is important",
				Input: map[string]interface{}{
					"source":         "/mnt/backup/archive.tar",
					"destination":    "/home/user/backups/",
					"preserve_attrs": true,
				},
				Output: map[string]interface{}{
					"source":           "/mnt/backup/archive.tar",
					"destination":      "/home/user/backups/archive.tar",
					"moved":            true,
					"was_rename":       false,
					"was_cross_device": true,
					"message":          "Successfully moved (cross-device)",
				},
				Explanation: "Cross-device move preserving original timestamps",
			},
		}).
		WithConstraints([]string{
			"Cannot move directory across different filesystems",
			"Atomic rename only works within same filesystem",
			"Cross-device moves use copy-then-delete (not atomic)",
			"Destination directory must exist unless create_dirs is true",
			"Cannot move a parent directory into its own subdirectory",
			"Overwrite protection is enabled by default",
			"Symbolic links are moved as links, not their targets",
			"Moving system files may require elevated permissions",
			"Large file moves across devices can be slow",
			"Interrupted cross-device moves may leave partial copies",
		}).
		WithErrorGuidance(map[string]string{
			"permission denied":          "Check file permissions and ownership for source and destination",
			"no such file or directory":  "Source doesn't exist or destination parent is missing",
			"destination already exists": "Use overwrite=true or choose different destination",
			"cross-device link":          "Filesystem boundary detected, using copy-delete method",
			"directory not empty":        "Cannot overwrite non-empty directory with file",
			"invalid argument":           "Check for invalid characters in paths",
			"device or resource busy":    "File is in use, close applications using it",
			"read-only file system":      "Destination filesystem is mounted read-only",
			"no space left on device":    "Insufficient space for cross-device copy",
			"operation not permitted":    "May need elevated privileges or check file attributes",
		}).
		WithCategory("file").
		WithTags([]string{"filesystem", "move", "rename", "transfer"}).
		WithVersion("2.0.0").
		WithBehavior(
			true,   // Deterministic - same inputs produce same results
			true,   // Destructive - removes source file
			false,  // No confirmation by default (use overwrite for safety)
			"fast", // Usually fast, can be slow for large cross-device
		)

	return builder.Build()
}

// crossDeviceMove performs a copy-then-delete operation for cross-device moves
func crossDeviceMove(ctx context.Context, src, dst string, srcInfo os.FileInfo, preserveAttrs bool) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer func() {
		_ = srcFile.Close()
	}()

	// Create destination file with same permissions
	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}

	// Copy content
	_, copyErr := io.Copy(dstFile, srcFile)
	closeErr := dstFile.Close()

	if copyErr != nil {
		_ = os.Remove(dst) // Clean up partial file
		return fmt.Errorf("failed to copy content: %w", copyErr)
	}
	if closeErr != nil {
		_ = os.Remove(dst) // Clean up
		return fmt.Errorf("failed to close destination: %w", closeErr)
	}

	// Preserve attributes if requested
	if preserveAttrs {
		// Set modification time
		if err := os.Chtimes(dst, srcInfo.ModTime(), srcInfo.ModTime()); err != nil {
			// Non-fatal error
			fmt.Printf("Warning: failed to preserve timestamps: %v\n", err)
		}

		// Set permissions (already done during creation, but ensure)
		if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
			// Non-fatal error
			fmt.Printf("Warning: failed to preserve permissions: %v\n", err)
		}
	}

	// Check context before deleting source
	select {
	case <-ctx.Done():
		// Context canceled, don't delete source
		_ = os.Remove(dst) // Clean up destination
		return ctx.Err()
	default:
	}

	// Delete source file
	if err := os.Remove(src); err != nil {
		// Try to clean up destination since we couldn't complete the move
		_ = os.Remove(dst)
		return fmt.Errorf("failed to remove source after copy: %w", err)
	}

	return nil
}

// MustGetFileMove retrieves the registered FileMove tool or panics
// This is a convenience function for users who want to ensure the tool exists
func MustGetFileMove() domain.Tool {
	return tools.MustGetTool("file_move")
}

// isSubPath checks if a path is a subpath of a parent path
func isSubPath(path, parent string) bool {
	// Clean and resolve both paths
	cleanPath := filepath.Clean(path)
	cleanParent := filepath.Clean(parent)

	// Get absolute paths
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return false
	}

	absParent, err := filepath.Abs(cleanParent)
	if err != nil {
		return false
	}

	// Check if path starts with parent
	relPath, err := filepath.Rel(absParent, absPath)
	if err != nil {
		return false
	}

	// If the relative path starts with "..", it's not a subpath
	return !strings.HasPrefix(relPath, "..")
}
