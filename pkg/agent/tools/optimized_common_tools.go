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
)

// OptimizedWebFetch creates an optimized tool for fetching web content
func OptimizedWebFetch() domain.Tool {
	return NewOptimizedTool(
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

// OptimizedExecuteCommand creates an optimized tool for executing shell commands
func OptimizedExecuteCommand() domain.Tool {
	return NewOptimizedTool(
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

// OptimizedReadFile creates an optimized tool for reading files
func OptimizedReadFile() domain.Tool {
	return NewOptimizedTool(
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

// OptimizedWriteFile creates an optimized tool for writing files
func OptimizedWriteFile() domain.Tool {
	return NewOptimizedTool(
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