// ABOUTME: File writing tool with atomic operations, append mode, and directory creation
// ABOUTME: Built-in tool that provides safe and enhanced file writing capabilities for agents

package file

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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

// writeFileOutputSchema defines the output schema for the WriteFile tool
var writeFileOutputSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"success": {
			Type:        "boolean",
			Description: "Whether the write operation succeeded",
		},
		"bytes_written": {
			Type:        "number",
			Description: "Number of bytes written",
		},
		"absolute_path": {
			Type:        "string",
			Description: "Absolute path to the written file",
		},
		"backup_path": {
			Type:        "string",
			Description: "Path to backup file if created",
		},
		"file_existed": {
			Type:        "boolean",
			Description: "Whether the file existed before writing",
		},
		"mod_time": {
			Type:        "string",
			Description: "Modification time after write",
		},
	},
	Required: []string{"success", "bytes_written", "absolute_path", "file_existed", "mod_time"},
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

// WriteFile creates a tool for writing files with enhanced capabilities including atomic operations and backup support.
// It provides safe file writing with options for atomic writes (write to temp, then rename), append mode, and automatic backups.
// The tool supports creating parent directories, custom file permissions, and path access control via state configuration.
// Atomic operations ensure data integrity by preventing partial writes in case of failures.
func WriteFile() domain.Tool {
	builder := atools.NewToolBuilder("file_write", "Writes content to files with atomic operations, append mode, and backup support").
		WithFunction(writeFileMain).
		WithParameterSchema(writeFileParamSchema).
		WithOutputSchema(writeFileOutputSchema).
		WithUsageInstructions(`Use this tool to write content to files with advanced features.

Features:
- Atomic write operations for data integrity
- Append mode for adding to existing files
- Automatic parent directory creation
- File backup before overwriting
- Path access control via state configuration
- Progress events for operation tracking

Parameters:
- path: File path to write (required)
- content: Content to write (required)
- mode: File permissions in octal (optional, default 0644)
- append: Append to file instead of overwrite (optional)
- create_dirs: Create parent directories if needed (optional)
- atomic: Use atomic write operation (optional)
- backup: Create backup of existing file (optional)

Atomic Write:
- Writes to temporary file first
- Renames to target path on success
- Prevents partial writes on failure
- Recommended for critical files

Backup Feature:
- Creates timestamped backup before overwrite
- Format: filename.backup-YYYYMMDD-HHMMSS.ext
- Only backs up existing files
- Can be auto-enabled via state

State Configuration:
- file_restricted_paths: Array of paths to block
- file_allowed_paths: Array of allowed path prefixes
- file_default_permissions: Default file mode
- file_auto_backup: Enable automatic backups

Security:
- Path restrictions enforced via state
- Parent directory creation requires explicit flag
- Atomic writes prevent corruption
- Proper permission setting

Performance:
- Atomic writes may be slower for large files
- Direct writes are fastest
- Progress events emitted for operations
- Context cancellation supported`).
		WithExamples([]domain.ToolExample{
			{
				Name:        "Simple file write",
				Description: "Write text to a file",
				Scenario:    "When creating or overwriting a file",
				Input: map[string]interface{}{
					"path":    "/home/user/notes.txt",
					"content": "Meeting notes:\n- Discuss project timeline\n- Review budget",
				},
				Output: map[string]interface{}{
					"success":       true,
					"bytes_written": 48,
					"absolute_path": "/home/user/notes.txt",
					"file_existed":  false,
					"mod_time":      "2024-01-15T10:30:00Z",
				},
				Explanation: "Creates new file with content, default permissions 0644",
			},
			{
				Name:        "Append to log file",
				Description: "Add entry to existing log",
				Scenario:    "When logging events or data",
				Input: map[string]interface{}{
					"path":    "app.log",
					"content": "[2024-01-15 10:30:00] User login successful\n",
					"append":  true,
				},
				Output: map[string]interface{}{
					"success":       true,
					"bytes_written": 46,
					"absolute_path": "/current/dir/app.log",
					"file_existed":  true,
					"mod_time":      "2024-01-15T10:30:00Z",
				},
				Explanation: "Appends to existing file without overwriting previous content",
			},
			{
				Name:        "Atomic write with backup",
				Description: "Safely update configuration file",
				Scenario:    "When updating critical configuration",
				Input: map[string]interface{}{
					"path":    "config.json",
					"content": `{"version": "2.0", "port": 8080, "debug": false}`,
					"atomic":  true,
					"backup":  true,
				},
				Output: map[string]interface{}{
					"success":       true,
					"bytes_written": 48,
					"absolute_path": "/app/config.json",
					"backup_path":   "/app/config.backup-20240115-103000.json",
					"file_existed":  true,
					"mod_time":      "2024-01-15T10:30:00Z",
				},
				Explanation: "Creates backup, writes atomically to prevent corruption",
			},
			{
				Name:        "Create file with directories",
				Description: "Write file in non-existent directory",
				Scenario:    "When output directory doesn't exist",
				Input: map[string]interface{}{
					"path":        "output/reports/2024/january/sales.csv",
					"content":     "Date,Product,Amount\n2024-01-15,Widget,100.00\n",
					"create_dirs": true,
				},
				Output: map[string]interface{}{
					"success":       true,
					"bytes_written": 48,
					"absolute_path": "/home/user/output/reports/2024/january/sales.csv",
					"file_existed":  false,
					"mod_time":      "2024-01-15T10:30:00Z",
				},
				Explanation: "Creates all parent directories before writing file",
			},
			{
				Name:        "Write with custom permissions",
				Description: "Create executable script",
				Scenario:    "When creating scripts or executables",
				Input: map[string]interface{}{
					"path":    "deploy.sh",
					"content": "#!/bin/bash\necho 'Deploying application...'\n",
					"mode":    0755,
				},
				Output: map[string]interface{}{
					"success":       true,
					"bytes_written": 44,
					"absolute_path": "/home/user/deploy.sh",
					"file_existed":  false,
					"mod_time":      "2024-01-15T10:30:00Z",
				},
				Explanation: "Creates file with executable permissions (755)",
			},
			{
				Name:        "Handle write errors",
				Description: "Attempt to write to read-only location",
				Scenario:    "When lacking write permissions",
				Input: map[string]interface{}{
					"path":    "/etc/system.conf",
					"content": "system config",
				},
				Output: map[string]interface{}{
					"error": "error writing file: open /etc/system.conf: permission denied",
				},
				Explanation: "Returns clear error when write fails",
			},
			{
				Name:        "Path restriction",
				Description: "Blocked by security policy",
				Scenario:    "When trying to write to restricted paths",
				Input: map[string]interface{}{
					"path":    "/etc/passwd",
					"content": "malicious content",
				},
				Output: map[string]interface{}{
					"error": "access denied: path /etc/passwd is restricted",
				},
				Explanation: "Path restrictions prevent writing to sensitive locations",
			},
		}).
		WithConstraints([]string{
			"Overwrites existing files unless append mode is used",
			"Atomic writes use temporary files requiring extra space",
			"Backup files are not automatically cleaned up",
			"Parent directory creation requires explicit flag",
			"File permissions default to 0644 (readable, not executable)",
			"Large files may take time to write atomically",
			"Path restrictions are enforced via state configuration",
			"Binary content should be base64 encoded in content field",
			"Line endings are preserved as provided",
			"Context cancellation stops write immediately",
		}).
		WithErrorGuidance(map[string]string{
			"permission denied":         "Check file and directory permissions, may need elevated privileges",
			"no such file or directory": "Parent directory doesn't exist, use create_dirs: true",
			"access denied":             "Path is restricted by security policy. Check allowed paths",
			"disk full":                 "Insufficient disk space. Free up space or write elsewhere",
			"file exists":               "File already exists and exclusive mode was requested",
			"invalid argument":          "Check file path for invalid characters",
			"too many open files":       "System file handle limit reached. Close other files",
			"operation not permitted":   "File may be immutable or system-protected",
			"cross-device link":         "Atomic rename failed across filesystems",
			"context deadline exceeded": "Write operation took too long. Try smaller content",
		}).
		WithCategory("file").
		WithTags([]string{"file", "write", "filesystem", "save", "create"}).
		WithVersion("2.0.0").
		WithBehavior(
			true,   // Deterministic - same input produces same file
			true,   // Destructive - overwrites existing files
			true,   // Requires confirmation for overwrites
			"fast", // Usually fast, can be slow for large files
		)

	return builder.Build()
}

// writeFileMain is the main function for the tool
func writeFileMain(ctx *domain.ToolContext, params WriteFileParams) (*WriteFileResult, error) {
	// Emit start event
	if ctx.Events != nil {
		ctx.Events.EmitMessage(fmt.Sprintf("Starting file write to %s", params.Path))
	}

	result := &WriteFileResult{
		Success: false,
	}

	// Check file access restrictions from state
	if ctx.State != nil {
		// Check restricted paths
		if restrictedPaths, exists := ctx.State.Get("file_restricted_paths"); exists {
			if paths, ok := restrictedPaths.([]string); ok {
				for _, restricted := range paths {
					if strings.HasPrefix(params.Path, restricted) {
						return nil, fmt.Errorf("access denied: path %s is restricted", params.Path)
					}
				}
			}
		}

		// Check allowed paths if specified
		if allowedPaths, exists := ctx.State.Get("file_allowed_paths"); exists {
			if paths, ok := allowedPaths.([]string); ok && len(paths) > 0 {
				allowed := false
				for _, allowedPath := range paths {
					if strings.HasPrefix(params.Path, allowedPath) {
						allowed = true
						break
					}
				}
				if !allowed {
					return nil, fmt.Errorf("access denied: path %s is not in allowed paths", params.Path)
				}
			}
		}
	}

	// Set default mode if not specified
	mode := params.Mode
	if mode == 0 {
		// Check state for default permissions
		if ctx.State != nil {
			if defaultMode, exists := ctx.State.Get("file_default_permissions"); exists {
				if m, ok := defaultMode.(os.FileMode); ok {
					mode = m
				}
			}
		}
		// Fall back to default if not in state
		if mode == 0 {
			mode = 0644
		}
	}

	// Check if file exists
	if _, err := os.Stat(params.Path); err == nil {
		result.FileExisted = true
	}

	// Check backup preferences from state
	shouldBackup := params.Backup
	if !shouldBackup && result.FileExisted && ctx.State != nil {
		// Check if auto-backup is enabled in state
		if autoBackup, exists := ctx.State.Get("file_auto_backup"); exists {
			if backup, ok := autoBackup.(bool); ok && backup {
				shouldBackup = true
			}
		}
	}

	// Emit progress: Creating directories
	if params.CreateDirs {
		if ctx.Events != nil {
			ctx.Events.EmitProgress(1, 4, "Creating parent directories")
		}
		dir := filepath.Dir(params.Path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			if ctx.Events != nil {
				ctx.Events.EmitError(err)
			}
			return nil, fmt.Errorf("error creating directories: %w", err)
		}
	}

	// Emit progress: Creating backup
	if shouldBackup && result.FileExisted {
		if ctx.Events != nil {
			ctx.Events.EmitProgress(2, 4, "Creating backup")
		}
		backupPath, err := createBackup(params.Path)
		if err != nil {
			if ctx.Events != nil {
				ctx.Events.EmitError(err)
			}
			return nil, fmt.Errorf("error creating backup: %w", err)
		}
		result.BackupPath = backupPath
	}

	// Emit progress: Writing file
	if ctx.Events != nil {
		ctx.Events.EmitProgress(3, 4, "Writing file content")
	}

	// Write the file
	var bytesWritten int
	var err error

	if params.Atomic {
		bytesWritten, err = atomicWrite(ctx.Context, params.Path, params.Content, mode, params.Append)
	} else if params.Append {
		bytesWritten, err = appendToFile(ctx.Context, params.Path, params.Content, mode)
	} else {
		bytesWritten, err = writeFile(ctx.Context, params.Path, params.Content, mode)
	}

	if err != nil {
		if ctx.Events != nil {
			ctx.Events.EmitError(err)
		}
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

	// Emit completion event with write operation details
	if ctx.Events != nil {
		ctx.Events.EmitProgress(4, 4, "File write complete")
		ctx.Events.EmitCustom("file_write_complete", map[string]interface{}{
			"path":          params.Path,
			"absolute_path": absPath,
			"bytes_written": bytesWritten,
			"file_existed":  result.FileExisted,
			"backup_path":   result.BackupPath,
			"mode":          mode,
			"atomic":        params.Atomic,
			"append":        params.Append,
		})
	}

	return result, nil
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
		defer func() {
			_ = file.Close()
		}()

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
			_ = tempFile.Close()
			_ = os.Remove(tempPath)
		}()

		// If appending, first copy existing content
		if append {
			if existingFile, err := os.Open(path); err == nil {
				defer func() {
					_ = existingFile.Close()
				}()
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
		_ = tempFile.Close()

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
	defer func() {
		_ = source.Close()
	}()

	dest, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("error creating backup file: %w", err)
	}
	defer func() {
		_ = dest.Close()
	}()

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
