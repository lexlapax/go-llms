// ABOUTME: File move/rename tool with safety checks and overwrite options
// ABOUTME: Built-in tool supporting atomic moves, cross-device transfers, and directory operations

package file

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	atools "github.com/lexlapax/go-llms/pkg/agent/tools"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// FileMoveParams defines parameters for the FileMove tool
type FileMoveParams struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Overwrite   bool   `json:"overwrite,omitempty"`    // Overwrite existing destination
	CreateDirs  bool   `json:"create_dirs,omitempty"`  // Create parent directories if needed
	PreserveAttrs bool `json:"preserve_attrs,omitempty"` // Preserve file attributes (permissions, times)
}

// FileMoveResult defines the result of the FileMove tool
type FileMoveResult struct {
	Source        string `json:"source"`
	Destination   string `json:"destination"`
	Moved         bool   `json:"moved"`
	WasRename     bool   `json:"was_rename"`      // True if same directory (rename only)
	WasCrossDevice bool  `json:"was_cross_device"` // True if moved across filesystems
	Message       string `json:"message,omitempty"`
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
		func(ctx context.Context, params FileMoveParams) (*FileMoveResult, error) {
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

			// Check if destination already exists
			if _, err := os.Stat(finalDst); err == nil && !params.Overwrite {
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

			// Try atomic rename first (works only on same filesystem)
			err = os.Rename(absSrc, finalDst)
			if err == nil {
				// Successful atomic move
				return &FileMoveResult{
					Source:        absSrc,
					Destination:   finalDst,
					Moved:         true,
					WasRename:     isRename,
					WasCrossDevice: false,
					Message:       "Successfully moved",
				}, nil
			}

			// If rename failed, might be cross-device
			// Only attempt cross-device move for files, not directories
			if srcInfo.IsDir() {
				return nil, fmt.Errorf("cannot move directory across devices: %w", err)
			}

			// Perform cross-device file move (copy then delete)
			if err := crossDeviceMove(ctx, absSrc, finalDst, srcInfo, params.PreserveAttrs); err != nil {
				return nil, fmt.Errorf("cross-device move failed: %w", err)
			}

			return &FileMoveResult{
				Source:        absSrc,
				Destination:   finalDst,
				Moved:         true,
				WasRename:     false,
				WasCrossDevice: true,
				Message:       "Successfully moved (cross-device)",
			}, nil
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