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

// ReadFile creates a tool for reading files with enhanced capabilities
func ReadFile() domain.Tool {
	return atools.NewTool(
		"file_read",
		"Reads file contents with support for large files, line ranges, and metadata",
		func(ctx *domain.ToolContext, params ReadFileParams) (*ReadFileResult, error) {
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
			defer file.Close()

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
		},
		readFileParamSchema,
	)
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
