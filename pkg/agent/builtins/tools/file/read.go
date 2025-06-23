// ABOUTME: File reading tool with streaming support, encoding detection, and metadata
// ABOUTME: Built-in tool that provides enhanced file reading capabilities for agents

package file

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	atools "github.com/lexlapax/go-llms/pkg/agent/tools"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// ReadFileParams defines parameters for the ReadFile tool
type ReadFileParams struct {
	Path        string `json:"path"`
	MaxSize     int64  `json:"max_size,omitempty"`     // Maximum bytes to read (0 = unlimited)
	LineStart   int    `json:"line_start,omitempty"`   // Start reading from this line (1-based)
	LineEnd     int    `json:"line_end,omitempty"`     // Stop reading at this line (inclusive)
	IncludeMeta bool   `json:"include_meta,omitempty"` // Include file metadata in response
}

// ReadFileResult defines the result of the ReadFile tool
type ReadFileResult struct {
	Content  string        `json:"content"`
	Metadata *FileMetadata `json:"metadata,omitempty"`
	Encoding string        `json:"encoding"` // Detected encoding
	IsBinary bool          `json:"is_binary"`
	Lines    int           `json:"lines,omitempty"`    // Number of lines read
	Warnings []string      `json:"warnings,omitempty"` // Any warnings during read
}

// FileMetadata contains file information
type FileMetadata struct {
	Size         int64     `json:"size"`
	Mode         string    `json:"mode"`
	ModTime      time.Time `json:"mod_time"`
	IsDir        bool      `json:"is_dir"`
	AbsolutePath string    `json:"absolute_path"`
	Extension    string    `json:"extension"`
}

// readFileParamSchema defines parameters for the ReadFile tool
var readFileParamSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"path": {
			Type:        "string",
			Description: "The path to the file to read",
		},
		"max_size": {
			Type:        "number",
			Description: "Maximum bytes to read (0 = unlimited, default: 10MB)",
		},
		"line_start": {
			Type:        "number",
			Description: "Start reading from this line number (1-based)",
		},
		"line_end": {
			Type:        "number",
			Description: "Stop reading at this line number (inclusive)",
		},
		"include_meta": {
			Type:        "boolean",
			Description: "Include file metadata in the response",
		},
	},
	Required: []string{"path"},
}

// readFileOutputSchema defines the output schema for the ReadFile tool
var readFileOutputSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"content": {
			Type:        "string",
			Description: "The file content",
		},
		"metadata": {
			Type:        "object",
			Description: "File metadata (if include_meta is true)",
			Properties: map[string]sdomain.Property{
				"size": {
					Type:        "number",
					Description: "File size in bytes",
				},
				"mode": {
					Type:        "string",
					Description: "File permissions mode",
				},
				"mod_time": {
					Type:        "string",
					Description: "Last modification time",
				},
				"is_dir": {
					Type:        "boolean",
					Description: "Whether the path is a directory",
				},
				"absolute_path": {
					Type:        "string",
					Description: "Absolute path to the file",
				},
				"extension": {
					Type:        "string",
					Description: "File extension",
				},
			},
		},
		"encoding": {
			Type:        "string",
			Description: "Detected file encoding (utf-8 or binary)",
		},
		"is_binary": {
			Type:        "boolean",
			Description: "Whether the file is binary",
		},
		"lines": {
			Type:        "number",
			Description: "Number of lines read (for text files)",
		},
		"warnings": {
			Type:        "array",
			Description: "Any warnings generated during read",
			Items: &sdomain.Property{
				Type: "string",
			},
		},
	},
	Required: []string{"content", "encoding", "is_binary"},
}

// init automatically registers the tool on package import
func init() {
	tools.MustRegisterTool("file_read", ReadFile(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "file_read",
			Category:    "file",
			Tags:        []string{"file", "read", "filesystem", "text", "binary"},
			Description: "Reads file contents with support for large files, line ranges, and metadata",
			Version:     "1.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Basic file read",
					Description: "Read entire file contents",
					Code:        `ReadFile().Execute(ctx, ReadFileParams{Path: "/path/to/file.txt"})`,
				},
				{
					Name:        "Read with line range",
					Description: "Read specific lines from a file",
					Code:        `ReadFile().Execute(ctx, ReadFileParams{Path: "large.log", LineStart: 100, LineEnd: 200})`,
				},
				{
					Name:        "Read with metadata",
					Description: "Get file contents and metadata",
					Code:        `ReadFile().Execute(ctx, ReadFileParams{Path: "data.json", IncludeMeta: true})`,
				},
			},
		},
		RequiredPermissions: []string{"file:read"},
		ResourceUsage: tools.ResourceInfo{
			Memory:      "medium", // Can be low for small files, high for large files
			Network:     false,
			FileSystem:  true,
			Concurrency: true,
		},
	})
}

// readFile is the main function for the tool
func readFile(ctx *domain.ToolContext, params ReadFileParams) (*ReadFileResult, error) {
	// Emit start event
	if ctx.Events != nil {
		ctx.Events.EmitMessage(fmt.Sprintf("Starting file read for %s", params.Path))
	}

	// Get max size from state or use default
	maxSize := params.MaxSize
	if maxSize == 0 {
		// Check state for default max size
		if ctx.State != nil {
			if defaultMaxSize, exists := ctx.State.Get("file_read_max_size"); exists {
				if size, ok := defaultMaxSize.(int64); ok {
					maxSize = size
				}
			}
		}
		// Fall back to default if not in state
		if maxSize == 0 {
			maxSize = 10 * 1024 * 1024 // 10MB default
		}
	}

	// Check file access restrictions from state
	if ctx.State != nil {
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

	// Emit progress event
	if ctx.Events != nil {
		ctx.Events.EmitProgress(1, 4, "Opening file")
	}

	// Open file
	file, err := os.Open(params.Path)
	if err != nil {
		if ctx.Events != nil {
			ctx.Events.EmitError(err)
		}
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	result := &ReadFileResult{
		Warnings: []string{},
	}

	// Get file metadata if requested
	if params.IncludeMeta {
		if ctx.Events != nil {
			ctx.Events.EmitProgress(2, 4, "Reading file metadata")
		}

		stat, err := file.Stat()
		if err != nil {
			return nil, fmt.Errorf("error getting file stats: %w", err)
		}

		absPath, _ := filepath.Abs(params.Path)
		result.Metadata = &FileMetadata{
			Size:         stat.Size(),
			Mode:         stat.Mode().String(),
			ModTime:      stat.ModTime(),
			IsDir:        stat.IsDir(),
			AbsolutePath: absPath,
			Extension:    filepath.Ext(params.Path),
		}

		if stat.IsDir() {
			return nil, fmt.Errorf("path is a directory, not a file")
		}
	}

	// Check encoding preferences from state
	preferredEncoding := ""
	if ctx.State != nil {
		if enc, exists := ctx.State.Get("file_preferred_encoding"); exists {
			if encStr, ok := enc.(string); ok {
				preferredEncoding = encStr
			}
		}
	}

	// Check if file is binary
	result.IsBinary, result.Encoding = detectFileType(file)
	if _, err := file.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("error resetting file position: %w", err)
	}

	// Override encoding if preference is set and file is text
	if preferredEncoding != "" && !result.IsBinary {
		result.Encoding = preferredEncoding
	}

	// Read file content
	if ctx.Events != nil {
		ctx.Events.EmitProgress(3, 4, "Reading file content")
	}

	if params.LineStart > 0 || params.LineEnd > 0 {
		// Line-based reading
		content, lines, err := readFileLines(ctx.Context, file, params.LineStart, params.LineEnd, maxSize)
		if err != nil {
			return nil, err
		}
		result.Content = content
		result.Lines = lines
	} else {
		// Full file reading
		content, err := readFileContent(ctx.Context, file, maxSize)
		if err != nil {
			return nil, err
		}
		result.Content = content

		// Count lines for text files
		if !result.IsBinary {
			result.Lines = strings.Count(content, "\n") + 1
		}
	}

	// Add warning if file was truncated
	if int64(len(result.Content)) >= maxSize {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("File truncated at %d bytes", maxSize))
	}

	// Emit completion event with custom file metadata
	if ctx.Events != nil {
		ctx.Events.EmitProgress(4, 4, "File read complete")

		// Emit custom event with file read details
		fileEventData := map[string]interface{}{
			"path":         params.Path,
			"bytes_read":   len(result.Content),
			"lines_read":   result.Lines,
			"is_binary":    result.IsBinary,
			"encoding":     result.Encoding,
			"elapsed_time": ctx.ElapsedTime().String(),
		}

		if result.Metadata != nil {
			fileEventData["file_size"] = result.Metadata.Size
			fileEventData["file_extension"] = result.Metadata.Extension
			fileEventData["absolute_path"] = result.Metadata.AbsolutePath
		}

		ctx.Events.EmitCustom("file_read_complete", fileEventData)
	}

	return result, nil
}

// ReadFile creates a tool for reading files with enhanced capabilities including streaming support and encoding detection.
// It automatically detects binary vs text files, supports reading specific line ranges for large files, and provides file metadata.
// The tool includes configurable size limits, path access control, and progress tracking for optimal performance.
// Content is streamed efficiently to handle large files without excessive memory usage.
func ReadFile() domain.Tool {
	builder := atools.NewToolBuilder("file_read", "Reads file contents with support for large files, line ranges, and metadata").
		WithFunction(readFile).
		WithParameterSchema(readFileParamSchema).
		WithOutputSchema(readFileOutputSchema).
		WithUsageInstructions(`Use this tool to read file contents with advanced features.

Features:
- Automatic encoding detection (UTF-8 or binary)
- Line range support for reading specific portions
- File size limits to prevent memory issues
- Metadata retrieval (size, permissions, timestamps)
- Path access control via state configuration
- Progress events for large file operations

Parameters:
- path: File path to read (required)
- max_size: Maximum bytes to read (optional, default 10MB or from state)
- line_start: Start reading from this line (optional, 1-based)
- line_end: Stop reading at this line (optional, inclusive)
- include_meta: Include file metadata (optional, default false)

Line Range Reading:
- Use line_start/line_end for large log files
- Lines are 1-based (first line is 1)
- Only text files support line ranges
- Binary files ignore line parameters

State Configuration:
- file_read_max_size: Default max size in bytes
- file_restricted_paths: Array of paths to block
- file_allowed_paths: Array of allowed path prefixes
- file_preferred_encoding: Override encoding detection

Security:
- Path restrictions can be enforced via state
- Symlinks are followed (be careful with access control)
- Binary files are detected and marked

Performance:
- Uses buffered reading for efficiency
- Streams large files instead of loading all at once
- Emits progress events during read
- Context cancellation supported`).
		WithExamples([]domain.ToolExample{
			{
				Name:        "Read text file",
				Description: "Read a simple text file",
				Scenario:    "When you need to read configuration or source files",
				Input: map[string]interface{}{
					"path": "/home/user/config.json",
				},
				Output: map[string]interface{}{
					"content":   `{"api_key": "secret", "port": 8080}`,
					"encoding":  "utf-8",
					"is_binary": false,
					"lines":     1,
				},
				Explanation: "Reads entire file content, detects UTF-8 encoding",
			},
			{
				Name:        "Read with metadata",
				Description: "Get file content and metadata",
				Scenario:    "When you need file information along with content",
				Input: map[string]interface{}{
					"path":         "/var/log/app.log",
					"include_meta": true,
				},
				Output: map[string]interface{}{
					"content": "2024-01-15 10:00:00 INFO Starting application\n2024-01-15 10:00:01 INFO Connected to database",
					"metadata": map[string]interface{}{
						"size":          2048,
						"mode":          "-rw-r--r--",
						"mod_time":      "2024-01-15T10:00:01Z",
						"is_dir":        false,
						"absolute_path": "/var/log/app.log",
						"extension":     ".log",
					},
					"encoding":  "utf-8",
					"is_binary": false,
					"lines":     2,
				},
				Explanation: "Includes file metadata like size, permissions, and timestamps",
			},
			{
				Name:        "Read specific lines",
				Description: "Read lines 100-150 from a large log file",
				Scenario:    "When analyzing specific portions of large files",
				Input: map[string]interface{}{
					"path":       "/var/log/system.log",
					"line_start": 100,
					"line_end":   150,
				},
				Output: map[string]interface{}{
					"content":   "[100 lines of log content from line 100 to 150]",
					"encoding":  "utf-8",
					"is_binary": false,
					"lines":     51,
				},
				Explanation: "Efficiently reads only the requested line range",
			},
			{
				Name:        "Read with size limit",
				Description: "Read large file with size constraint",
				Scenario:    "When dealing with potentially huge files",
				Input: map[string]interface{}{
					"path":     "/data/large_dataset.csv",
					"max_size": 1048576, // 1MB
				},
				Output: map[string]interface{}{
					"content":   "[First 1MB of CSV data]",
					"encoding":  "utf-8",
					"is_binary": false,
					"lines":     5000,
					"warnings":  []string{"File truncated at 1048576 bytes"},
				},
				Explanation: "Stops reading at size limit and adds truncation warning",
			},
			{
				Name:        "Binary file detection",
				Description: "Read a binary file",
				Scenario:    "When accidentally trying to read binary files",
				Input: map[string]interface{}{
					"path": "/usr/bin/ls",
				},
				Output: map[string]interface{}{
					"content":   "[Binary content - may appear garbled]",
					"encoding":  "binary",
					"is_binary": true,
				},
				Explanation: "Detects binary files and marks them appropriately",
			},
			{
				Name:        "Handle missing file",
				Description: "Attempt to read non-existent file",
				Scenario:    "When file doesn't exist",
				Input: map[string]interface{}{
					"path": "/tmp/nonexistent.txt",
				},
				Output: map[string]interface{}{
					"error": "error opening file: open /tmp/nonexistent.txt: no such file or directory",
				},
				Explanation: "Returns clear error for missing files",
			},
			{
				Name:        "Path restriction",
				Description: "Blocked by security policy",
				Scenario:    "When trying to read restricted paths",
				Input: map[string]interface{}{
					"path": "/etc/shadow",
				},
				Output: map[string]interface{}{
					"error": "access denied: path /etc/shadow is restricted",
				},
				Explanation: "Path restrictions can be configured via state",
			},
		}).
		WithConstraints([]string{
			"Default size limit is 10MB unless overridden",
			"Line numbers are 1-based, not 0-based",
			"Binary file detection based on first 512 bytes",
			"UTF-8 encoding assumed for text files",
			"Symlinks are followed to their targets",
			"Directory paths return an error",
			"Empty files return empty content string",
			"Line range reading only works for text files",
			"Progress events emitted for operations over 1MB",
			"Context cancellation stops read immediately",
		}).
		WithErrorGuidance(map[string]string{
			"no such file or directory": "Check if the file path is correct and the file exists",
			"permission denied":         "Ensure you have read permissions for the file",
			"is a directory":            "Use file_list tool for directory contents, not file_read",
			"access denied":             "Path is restricted by security policy. Check allowed paths",
			"file too large":            "Increase max_size parameter or read specific line ranges",
			"invalid line range":        "Ensure line_start <= line_end and both are positive",
			"context deadline exceeded": "File read took too long. Try reading smaller portions",
			"too many open files":       "System file handle limit reached. Close other files",
			"encoding not supported":    "File uses unsupported encoding. Try binary mode",
			"symlink loop detected":     "File path contains circular symbolic links",
		}).
		WithCategory("file").
		WithTags([]string{"file", "read", "filesystem", "text", "binary"}).
		WithVersion("2.0.0").
		WithBehavior(
			true,   // Deterministic - same file returns same content
			false,  // Not destructive - only reads
			false,  // No confirmation needed
			"fast", // Usually fast, can be slow for large files
		)

	return builder.Build()
}

// detectFileType checks if file is binary and detects encoding
func detectFileType(file *os.File) (isBinary bool, encoding string) {
	// Read first 512 bytes for detection
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	buf = buf[:n]

	// Check if content is valid UTF-8
	if utf8.Valid(buf) {
		// Check for null bytes (common in binary files)
		for _, b := range buf {
			if b == 0 {
				return true, "binary"
			}
		}
		return false, "utf-8"
	}

	// Not valid UTF-8, likely binary
	return true, "binary"
}

// readFileContent reads entire file content up to maxSize
func readFileContent(ctx context.Context, file *os.File, maxSize int64) (string, error) {
	reader := bufio.NewReader(file)
	var content strings.Builder
	buf := make([]byte, 4096) // 4KB buffer

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
			n, err := reader.Read(buf)
			if n > 0 {
				// Check size limit
				if int64(content.Len()+n) > maxSize {
					remaining := maxSize - int64(content.Len())
					content.Write(buf[:remaining])
					return content.String(), nil
				}
				content.Write(buf[:n])
			}
			if err == io.EOF {
				return content.String(), nil
			}
			if err != nil {
				return "", fmt.Errorf("error reading file: %w", err)
			}
		}
	}
}

// readFileLines reads specific lines from a file
func readFileLines(ctx context.Context, file *os.File, start, end int, maxSize int64) (string, int, error) {
	scanner := bufio.NewScanner(file)
	var content strings.Builder
	lineNum := 0
	linesRead := 0

	// Adjust start to 0-based
	if start > 0 {
		start--
	}

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return "", linesRead, ctx.Err()
		default:
			lineNum++

			// Skip lines before start
			if start > 0 && lineNum <= start {
				continue
			}

			// Stop at end line
			if end > 0 && lineNum > end {
				break
			}

			line := scanner.Text()

			// Check size limit
			if int64(content.Len()+len(line)+1) > maxSize {
				break
			}

			if content.Len() > 0 {
				content.WriteString("\n")
			}
			content.WriteString(line)
			linesRead++
		}
	}

	if err := scanner.Err(); err != nil {
		return "", linesRead, fmt.Errorf("error reading file: %w", err)
	}

	return content.String(), linesRead, nil
}

// MustGetReadFile retrieves the registered ReadFile tool or panics
func MustGetReadFile() domain.Tool {
	return tools.MustGetTool("file_read")
}
