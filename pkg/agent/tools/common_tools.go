package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// Common tool parameter schemas
var (
	// WebFetchParamSchema defines parameters for the WebFetch tool
	WebFetchParamSchema = &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"url": {
				Type:        "string",
				Format:      "uri",
				Description: "The URL to fetch content from",
			},
		},
		Required: []string{"url"},
	}

	// SearchParamSchema defines parameters for the Search tool
	SearchParamSchema = &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"query": {
				Type:        "string",
				Description: "The search query",
			},
		},
		Required: []string{"query"},
	}

	// ExecuteCommandParamSchema defines parameters for the ExecuteCommand tool
	ExecuteCommandParamSchema = &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"command": {
				Type:        "string",
				Description: "The command to execute",
			},
			"timeout": {
				Type:        "number",
				Description: "Timeout in seconds (default: 30)",
			},
		},
		Required: []string{"command"},
	}

	// ReadFileParamSchema defines parameters for the ReadFile tool
	ReadFileParamSchema = &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"path": {
				Type:        "string",
				Description: "The path to the file",
			},
		},
		Required: []string{"path"},
	}

	// WriteFileParamSchema defines parameters for the WriteFile tool
	WriteFileParamSchema = &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"path": {
				Type:        "string",
				Description: "The path to the file",
			},
			"content": {
				Type:        "string",
				Description: "The content to write",
			},
		},
		Required: []string{"path", "content"},
	}
)

// WebFetchParams defines parameters for the WebFetch tool
type WebFetchParams struct {
	URL string `json:"url"`
}

// WebFetchResult defines the result of the WebFetch tool
type WebFetchResult struct {
	Content string `json:"content"`
	Status  int    `json:"status"`
}

// WebFetch creates a tool for fetching web content
// This function is optimized for better performance through:
// - Using pre-computed type information
// - Object pooling for arguments
// - Efficient parameter conversion
func WebFetch() domain.Tool {
	return NewTool(
		"web_fetch",
		"Fetches content from a URL",
		func(ctx context.Context, params WebFetchParams) (*WebFetchResult, error) {
			client := &http.Client{
				Timeout: 30 * time.Second,
			}

			req, err := http.NewRequestWithContext(ctx, "GET", params.URL, nil)
			if err != nil {
				return nil, fmt.Errorf("error creating request: %w", err)
			}

			resp, err := client.Do(req)
			if err != nil {
				return nil, fmt.Errorf("error fetching URL: %w", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("error reading response: %w", err)
			}

			return &WebFetchResult{
				Content: string(body),
				Status:  resp.StatusCode,
			}, nil
		},
		WebFetchParamSchema,
	)
}

// ExecuteCommandParams defines parameters for the ExecuteCommand tool
type ExecuteCommandParams struct {
	Command string  `json:"command"`
	Timeout float64 `json:"timeout"`
}

// ExecuteCommandResult defines the result of the ExecuteCommand tool
type ExecuteCommandResult struct {
	Output   string `json:"output"`
	ExitCode int    `json:"exit_code"`
}

// ExecuteCommand creates a tool for executing shell commands
// This function is optimized for better performance through:
// - Using pre-computed type information
// - Object pooling for arguments
// - Efficient parameter conversion
func ExecuteCommand() domain.Tool {
	return NewTool(
		"execute_command",
		"Executes a shell command",
		func(ctx context.Context, params ExecuteCommandParams) (*ExecuteCommandResult, error) {
			timeout := 30 * time.Second
			if params.Timeout > 0 {
				timeout = time.Duration(params.Timeout) * time.Second
			}

			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			cmd := exec.CommandContext(ctx, "sh", "-c", params.Command)
			output, err := cmd.CombinedOutput()

			exitCode := 0
			if err != nil {
				// Try to get the exit code
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitCode = exitErr.ExitCode()
				} else {
					return nil, fmt.Errorf("error executing command: %w", err)
				}
			}

			return &ExecuteCommandResult{
				Output:   string(output),
				ExitCode: exitCode,
			}, nil
		},
		ExecuteCommandParamSchema,
	)
}

// ReadFileParams defines parameters for the ReadFile tool
type ReadFileParams struct {
	Path string `json:"path"`
}

// ReadFileResult defines the result of the ReadFile tool
type ReadFileResult struct {
	Content string `json:"content"`
}

// ReadFile creates a tool for reading files
// This function is optimized for better performance through:
// - Using pre-computed type information
// - Object pooling for arguments
// - Efficient parameter conversion
func ReadFile() domain.Tool {
	return NewTool(
		"read_file",
		"Reads the content of a file",
		func(ctx context.Context, params ReadFileParams) (*ReadFileResult, error) {
			content, err := os.ReadFile(params.Path)
			if err != nil {
				return nil, fmt.Errorf("error reading file: %w", err)
			}

			return &ReadFileResult{
				Content: string(content),
			}, nil
		},
		ReadFileParamSchema,
	)
}

// WriteFileParams defines parameters for the WriteFile tool
type WriteFileParams struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// WriteFileResult defines the result of the WriteFile tool
type WriteFileResult struct {
	Success bool `json:"success"`
}

// WriteFile creates a tool for writing files
// This function is optimized for better performance through:
// - Using pre-computed type information
// - Object pooling for arguments
// - Efficient parameter conversion
func WriteFile() domain.Tool {
	return NewTool(
		"write_file",
		"Writes content to a file",
		func(ctx context.Context, params WriteFileParams) (*WriteFileResult, error) {
			err := os.WriteFile(params.Path, []byte(params.Content), 0644)
			if err != nil {
				return nil, fmt.Errorf("error writing file: %w", err)
			}

			return &WriteFileResult{
				Success: true,
			}, nil
		},
		WriteFileParamSchema,
	)
}
