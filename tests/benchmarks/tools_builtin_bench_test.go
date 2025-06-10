package benchmarks

// ABOUTME: Benchmarks for built-in tools performance and registry operations
// ABOUTME: Measures performance of built-in tools and discovery operations

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/domain"

	// Import built-in tools
	builtinTools "github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/file"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/system"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
)

// BenchmarkBuiltinFileRead benchmarks the built-in file read tool
func BenchmarkBuiltinFileRead(b *testing.B) {
	toolCtx := &domain.ToolContext{
		Context: context.Background(),
		RunID:   "bench-builtin",
	}

	// Create a test file
	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "benchmark_test.txt")
	testContent := "This is a benchmark test file.\nIt contains multiple lines.\nFor testing file read performance."
	if err := os.WriteFile(testFile, []byte(testContent), 0600); err != nil {
		b.Fatal(err)
	}

	tool, ok := builtinTools.GetTool("file_read")
	if !ok {
		b.Fatal("file_read tool not found in registry")
	}

	params := map[string]interface{}{
		"path": testFile,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := tool.Execute(toolCtx, params)
		if err != nil {
			b.Fatal(err)
		}
		_ = result
	}
}

// BenchmarkBuiltinFileWrite benchmarks the built-in file write tool
func BenchmarkBuiltinFileWrite(b *testing.B) {
	toolCtx := &domain.ToolContext{
		Context: context.Background(),
		RunID:   "bench-builtin",
	}
	tempDir := b.TempDir()
	testContent := "This is benchmark test content for write operations."

	tool, ok := builtinTools.GetTool("file_write")
	if !ok {
		b.Fatal("file_write tool not found in registry")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testFile := filepath.Join(tempDir, "write_test.txt")
		params := map[string]interface{}{
			"path":    testFile,
			"content": testContent,
		}

		result, err := tool.Execute(toolCtx, params)
		if err != nil {
			b.Fatal(err)
		}
		_ = result
		_ = os.Remove(testFile) // Clean up for next iteration
	}
}

// BenchmarkBuiltinLargeFileHandling benchmarks large file handling
func BenchmarkBuiltinLargeFileHandling(b *testing.B) {
	toolCtx := &domain.ToolContext{
		Context: context.Background(),
		RunID:   "bench-builtin",
	}

	// Create a larger test file (1MB)
	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "large_test.txt")

	// Generate 1MB of content
	var content []byte
	line := []byte("This is a line of test content that will be repeated many times.\n")
	for len(content) < 1024*1024 {
		content = append(content, line...)
	}

	if err := os.WriteFile(testFile, content, 0600); err != nil {
		b.Fatal(err)
	}

	tool, ok := builtinTools.GetTool("file_read")
	if !ok {
		b.Fatal("file_read tool not found in registry")
	}

	// Built-in tool can handle large files with streaming
	params := map[string]interface{}{
		"path": testFile,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := tool.Execute(toolCtx, params)
		if err != nil {
			b.Fatal(err)
		}
		_ = result
	}
}

// BenchmarkBuiltinExecuteCommand benchmarks the built-in execute command tool
func BenchmarkBuiltinExecuteCommand(b *testing.B) {
	toolCtx := &domain.ToolContext{
		Context: context.Background(),
		RunID:   "bench-builtin",
	}

	tool, ok := builtinTools.GetTool("execute_command")
	if !ok {
		b.Fatal("execute_command tool not found in registry")
	}

	params := map[string]interface{}{
		"command": "echo 'hello world'",
		"timeout": 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := tool.Execute(toolCtx, params)
		if err != nil {
			b.Fatal(err)
		}
		_ = result
	}
}

// BenchmarkToolRegistryLookup benchmarks registry lookup performance
func BenchmarkToolRegistryLookup(b *testing.B) {
	toolNames := []string{"file_read", "file_write", "web_fetch", "web_search", "execute_command"}

	b.Run("GetTool", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, name := range toolNames {
				tool, _ := builtinTools.GetTool(name)
				_ = tool
			}
		}
	})

	b.Run("MustGetTool", func(b *testing.B) {
		// Ensure tools exist first
		for _, name := range toolNames {
			if _, ok := builtinTools.GetTool(name); !ok {
				b.Skipf("Tool %s not found", name)
			}
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, name := range toolNames {
				tool := builtinTools.MustGetTool(name)
				_ = tool
			}
		}
	})
}

// BenchmarkToolDiscovery benchmarks tool discovery operations
func BenchmarkToolDiscovery(b *testing.B) {
	b.Run("ListAll", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tools := builtinTools.Tools.List()
			_ = tools
		}
	})

	b.Run("ListByCategory", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			webTools := builtinTools.Tools.ListByCategory("web")
			fileTools := builtinTools.Tools.ListByCategory("file")
			systemTools := builtinTools.Tools.ListByCategory("system")
			_ = webTools
			_ = fileTools
			_ = systemTools
		}
	})

	b.Run("Search", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			fileResults := builtinTools.Tools.Search("file")
			webResults := builtinTools.Tools.Search("web")
			commandResults := builtinTools.Tools.Search("command")
			_ = fileResults
			_ = webResults
			_ = commandResults
		}
	})
}

// BenchmarkAtomicFileWrite benchmarks atomic file writing with backups
func BenchmarkAtomicFileWrite(b *testing.B) {
	toolCtx := &domain.ToolContext{
		Context: context.Background(),
		RunID:   "bench-builtin",
	}
	tempDir := b.TempDir()

	tool, ok := builtinTools.GetTool("file_write")
	if !ok {
		b.Fatal("file_write tool not found in registry")
	}

	// Create initial file
	testFile := filepath.Join(tempDir, "atomic_test.txt")
	initialContent := "Initial content"
	if err := os.WriteFile(testFile, []byte(initialContent), 0600); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		params := map[string]interface{}{
			"path":          testFile,
			"content":       "Updated content " + string(rune(i)),
			"atomic":        true,
			"create_backup": true,
		}

		result, err := tool.Execute(toolCtx, params)
		if err != nil {
			b.Fatal(err)
		}
		_ = result
	}
}
