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

// FileMove creates a tool for moving or renaming files and directories
// This is a built-in tool optimized for:
// - Atomic moves within same filesystem
// - Safe cross-device transfers
// - Directory structure preservation
// - Conflict resolution options
func FileMove() domain.Tool {
	return atools.NewTool(
		"file_move",
		"Moves or renames files and directories",
		func(ctx *domain.ToolContext, params FileMoveParams) (*FileMoveResult, error) {
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
		},
		fileMoveParamSchema,
	)
}

// crossDeviceMove performs a copy-then-delete operation for cross-device moves
func crossDeviceMove(ctx context.Context, src, dst string, srcInfo os.FileInfo, preserveAttrs bool) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer srcFile.Close()

	// Create destination file with same permissions
	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}

	// Copy content
	_, copyErr := io.Copy(dstFile, srcFile)
	closeErr := dstFile.Close()

	if copyErr != nil {
		os.Remove(dst) // Clean up partial file
		return fmt.Errorf("failed to copy content: %w", copyErr)
	}
	if closeErr != nil {
		os.Remove(dst) // Clean up
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
		// Context cancelled, don't delete source
		os.Remove(dst) // Clean up destination
		return ctx.Err()
	default:
	}

	// Delete source file
	if err := os.Remove(src); err != nil {
		// Try to clean up destination since we couldn't complete the move
		os.Remove(dst)
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
