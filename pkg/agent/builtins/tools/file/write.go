// ABOUTME: File writing tool with atomic operations, append mode, and directory creation
// ABOUTME: Built-in tool that provides safe and enhanced file writing capabilities for agents

package file

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	atools "github.com/lexlapax/go-llms/pkg/agent/tools"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// WriteFileParams defines parameters for the WriteFile tool
type WriteFileParams struct {
	Path       string      `json:"path"`
	Content    string      `json:"content"`
	Mode       os.FileMode `json:"mode,omitempty"`        // File permissions (default: 0644)
	Append     bool        `json:"append,omitempty"`      // Append to file instead of overwrite
	CreateDirs bool        `json:"create_dirs,omitempty"` // Create parent directories if needed
	Atomic     bool        `json:"atomic,omitempty"`      // Use atomic write (write to temp, then rename)
	Backup     bool        `json:"backup,omitempty"`      // Create backup of existing file
}

// WriteFileResult defines the result of the WriteFile tool
type WriteFileResult struct {
	Success      bool      `json:"success"`
	BytesWritten int       `json:"bytes_written"`
	AbsolutePath string    `json:"absolute_path"`
	BackupPath   string    `json:"backup_path,omitempty"` // Path to backup file if created
	FileExisted  bool      `json:"file_existed"`
	ModTime      time.Time `json:"mod_time"`
}

// writeFileParamSchema defines parameters for the WriteFile tool
var writeFileParamSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"path": {
			Type:        "string",
			Description: "The path to the file to write",
		},
		"content": {
			Type:        "string",
			Description: "The content to write to the file",
		},
		"mode": {
			Type:        "number",
			Description: "File permissions in octal (default: 0644)",
		},
		"append": {
			Type:        "boolean",
			Description: "Append to existing file instead of overwriting",
		},
		"create_dirs": {
			Type:        "boolean",
			Description: "Create parent directories if they don't exist",
		},
		"atomic": {
			Type:        "boolean",
			Description: "Use atomic write operation (safer for important files)",
		},
		"backup": {
			Type:        "boolean",
			Description: "Create backup of existing file before writing",
		},
	},
	Required: []string{"path", "content"},
}

// init automatically registers the tool on package import
func init() {
	tools.MustRegisterTool("file_write", WriteFile(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "file_write",
			Category:    "file",
			Tags:        []string{"file", "write", "filesystem", "save", "create"},
			Description: "Writes content to files with atomic operations, append mode, and backup support",
			Version:     "1.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Basic file write",
					Description: "Write content to a file",
					Code:        `WriteFile().Execute(ctx, WriteFileParams{Path: "/path/to/file.txt", Content: "Hello, World!"})`,
				},
				{
					Name:        "Append to file",
					Description: "Append content to existing file",
					Code:        `WriteFile().Execute(ctx, WriteFileParams{Path: "log.txt", Content: "New log entry\n", Append: true})`,
				},
				{
					Name:        "Atomic write with backup",
					Description: "Safely write important file with backup",
					Code:        `WriteFile().Execute(ctx, WriteFileParams{Path: "config.json", Content: jsonContent, Atomic: true, Backup: true})`,
				},
				{
					Name:        "Create with directories",
					Description: "Create file and parent directories",
					Code:        `WriteFile().Execute(ctx, WriteFileParams{Path: "data/output/result.txt", Content: "Results", CreateDirs: true})`,
				},
			},
		},
		RequiredPermissions: []string{"file:write"},
		ResourceUsage: tools.ResourceInfo{
			Memory:      "low",
			Network:     false,
			FileSystem:  true,
			Concurrency: true, // Atomic operations are thread-safe
		},
	})
}

// WriteFile creates a tool for writing files with enhanced safety features
func WriteFile() domain.Tool {
	return atools.NewTool(
		"file_write",
		"Writes content to files with atomic operations, append mode, and backup support",
		func(ctx context.Context, params WriteFileParams) (*WriteFileResult, error) {
			result := &WriteFileResult{
				Success: false,
			}

			// Set default mode if not specified
			mode := params.Mode
			if mode == 0 {
				mode = 0644
			}

			// Check if file exists
			if _, err := os.Stat(params.Path); err == nil {
				result.FileExisted = true
			}

			// Create parent directories if requested
			if params.CreateDirs {
				dir := filepath.Dir(params.Path)
				if err := os.MkdirAll(dir, 0755); err != nil {
					return nil, fmt.Errorf("error creating directories: %w", err)
				}
			}

			// Create backup if requested and file exists
			if params.Backup && result.FileExisted {
				backupPath, err := createBackup(params.Path)
				if err != nil {
					return nil, fmt.Errorf("error creating backup: %w", err)
				}
				result.BackupPath = backupPath
			}

			// Write the file
			var bytesWritten int
			var err error

			if params.Atomic {
				bytesWritten, err = atomicWrite(ctx, params.Path, params.Content, mode, params.Append)
			} else if params.Append {
				bytesWritten, err = appendToFile(ctx, params.Path, params.Content, mode)
			} else {
				bytesWritten, err = writeFile(ctx, params.Path, params.Content, mode)
			}

			if err != nil {
				return nil, err
			}

			// Get absolute path
			absPath, _ := filepath.Abs(params.Path)

			// Get file info for mod time
			if info, err := os.Stat(params.Path); err == nil {
				result.ModTime = info.ModTime()
			}

			result.Success = true
			result.BytesWritten = bytesWritten
			result.AbsolutePath = absPath

			return result, nil
		},
		writeFileParamSchema,
	)
}

// writeFile performs a standard file write
func writeFile(ctx context.Context, path, content string, mode os.FileMode) (int, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
		data := []byte(content)
		if err := os.WriteFile(path, data, mode); err != nil {
			return 0, fmt.Errorf("error writing file: %w", err)
		}
		return len(data), nil
	}
}

// appendToFile appends content to an existing file
func appendToFile(ctx context.Context, path, content string, mode os.FileMode) (int, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
		// Open file with append flag, create if doesn't exist
		file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, mode)
		if err != nil {
			return 0, fmt.Errorf("error opening file for append: %w", err)
		}
		defer file.Close()

		n, err := file.WriteString(content)
		if err != nil {
			return 0, fmt.Errorf("error appending to file: %w", err)
		}

		return n, nil
	}
}

// atomicWrite performs an atomic file write operation
func atomicWrite(ctx context.Context, path, content string, mode os.FileMode, append bool) (int, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
		// Create temporary file in the same directory
		dir := filepath.Dir(path)
		tempFile, err := os.CreateTemp(dir, ".tmp-*")
		if err != nil {
			return 0, fmt.Errorf("error creating temp file: %w", err)
		}
		tempPath := tempFile.Name()

		// Ensure temp file is cleaned up
		defer func() {
			tempFile.Close()
			os.Remove(tempPath)
		}()

		// If appending, first copy existing content
		if append {
			if existingFile, err := os.Open(path); err == nil {
				defer existingFile.Close()
				if _, err := io.Copy(tempFile, existingFile); err != nil {
					return 0, fmt.Errorf("error copying existing content: %w", err)
				}
			}
		}

		// Write new content
		n, err := tempFile.WriteString(content)
		if err != nil {
			return 0, fmt.Errorf("error writing to temp file: %w", err)
		}

		// Sync to ensure data is written to disk
		if err := tempFile.Sync(); err != nil {
			return 0, fmt.Errorf("error syncing temp file: %w", err)
		}

		// Close temp file before rename
		tempFile.Close()

		// Set proper permissions
		if err := os.Chmod(tempPath, mode); err != nil {
			return 0, fmt.Errorf("error setting file permissions: %w", err)
		}

		// Atomic rename
		if err := os.Rename(tempPath, path); err != nil {
			return 0, fmt.Errorf("error performing atomic rename: %w", err)
		}

		return n, nil
	}
}

// createBackup creates a backup of the existing file
func createBackup(path string) (string, error) {
	// Generate backup filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	ext := filepath.Ext(path)
	base := path[:len(path)-len(ext)]
	backupPath := fmt.Sprintf("%s.backup-%s%s", base, timestamp, ext)

	// Copy file to backup location
	source, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("error opening source file: %w", err)
	}
	defer source.Close()

	dest, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("error creating backup file: %w", err)
	}
	defer dest.Close()

	if _, err := io.Copy(dest, source); err != nil {
		return "", fmt.Errorf("error copying to backup: %w", err)
	}

	// Copy file permissions
	if info, err := source.Stat(); err == nil {
		if err := dest.Chmod(info.Mode()); err != nil {
			// Log error but don't fail backup creation
			// Some filesystems may not support chmod
			_ = err
		}
	}

	return backupPath, nil
}

// MustGetWriteFile retrieves the registered WriteFile tool or panics
func MustGetWriteFile() domain.Tool {
	return tools.MustGetTool("file_write")
}
