// ABOUTME: System information tool for retrieving OS, architecture, and resource details
// ABOUTME: Built-in tool providing comprehensive system metadata and runtime information

package system

import (
	"os"
	"runtime"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	atools "github.com/lexlapax/go-llms/pkg/agent/tools"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// GetSystemInfoParams defines parameters for the GetSystemInfo tool
type GetSystemInfoParams struct {
	IncludeEnvironment bool `json:"include_environment,omitempty"` // Include environment summary
	IncludeMemory      bool `json:"include_memory,omitempty"`      // Include memory statistics
	IncludeRuntime     bool `json:"include_runtime,omitempty"`     // Include Go runtime info
}

// SystemInfo represents comprehensive system information
type SystemInfo struct {
	OS           OSInfo           `json:"os"`
	Architecture string           `json:"architecture"`
	CPUs         int              `json:"cpus"`
	Hostname     string           `json:"hostname,omitempty"`
	Memory       *MemoryInfo      `json:"memory,omitempty"`
	Runtime      *RuntimeInfo     `json:"runtime,omitempty"`
	Environment  *EnvironmentInfo `json:"environment,omitempty"`
	Timestamp    string           `json:"timestamp"`
}

// OSInfo contains operating system details
type OSInfo struct {
	Name     string `json:"name"`     // OS name (e.g., "darwin", "linux", "windows")
	Platform string `json:"platform"` // Human-readable name
	Version  string `json:"version,omitempty"`
}

// MemoryInfo contains memory statistics
type MemoryInfo struct {
	Alloc      uint64 `json:"alloc"`       // Bytes allocated and in use
	TotalAlloc uint64 `json:"total_alloc"` // Total bytes allocated (even if freed)
	Sys        uint64 `json:"sys"`         // Bytes obtained from system
	NumGC      uint32 `json:"num_gc"`      // Number of completed GC cycles
}

// RuntimeInfo contains Go runtime information
type RuntimeInfo struct {
	Version      string `json:"version"`
	Compiler     string `json:"compiler"`
	NumCPU       int    `json:"num_cpu"`
	NumGoroutine int    `json:"num_goroutine"`
	GOMAXPROCS   int    `json:"gomaxprocs"`
}

// EnvironmentInfo contains environment summary
type EnvironmentInfo struct {
	User         string `json:"user,omitempty"`
	Home         string `json:"home,omitempty"`
	WorkingDir   string `json:"working_dir,omitempty"`
	TempDir      string `json:"temp_dir"`
	PathDirs     int    `json:"path_dirs"`      // Number of directories in PATH
	TotalEnvVars int    `json:"total_env_vars"` // Total environment variables
	GoPath       string `json:"gopath,omitempty"`
	GoRoot       string `json:"goroot,omitempty"`
}

// getSystemInfoParamSchema defines parameters for the GetSystemInfo tool
var getSystemInfoParamSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"include_environment": {
			Type:        "boolean",
			Description: "Include environment summary information",
		},
		"include_memory": {
			Type:        "boolean",
			Description: "Include memory statistics",
		},
		"include_runtime": {
			Type:        "boolean",
			Description: "Include Go runtime information",
		},
	},
}

// Platform name mappings
var platformNames = map[string]string{
	"darwin":  "macOS",
	"linux":   "Linux",
	"windows": "Windows",
	"freebsd": "FreeBSD",
	"openbsd": "OpenBSD",
	"netbsd":  "NetBSD",
	"android": "Android",
	"ios":     "iOS",
}

// init automatically registers the tool on package import
func init() {
	tools.MustRegisterTool("get_system_info", GetSystemInfo(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "get_system_info",
			Category:    "system",
			Tags:        []string{"system", "info", "os", "architecture", "resources"},
			Description: "Retrieves comprehensive system information",
			Version:     "1.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Basic system info",
					Description: "Get basic OS and architecture information",
					Code:        `GetSystemInfo().Execute(ctx, GetSystemInfoParams{})`,
				},
				{
					Name:        "Full system info",
					Description: "Get all available system information",
					Code:        `GetSystemInfo().Execute(ctx, GetSystemInfoParams{IncludeEnvironment: true, IncludeMemory: true, IncludeRuntime: true})`,
				},
				{
					Name:        "Memory statistics",
					Description: "Get system info with memory statistics",
					Code:        `GetSystemInfo().Execute(ctx, GetSystemInfoParams{IncludeMemory: true})`,
				},
			},
		},
		RequiredPermissions: []string{"system:read"},
		ResourceUsage: tools.ResourceInfo{
			Memory:      "low",
			Network:     false,
			FileSystem:  false,
			Concurrency: true,
		},
	})
}

// GetSystemInfo creates a tool for retrieving system information
// This is a built-in tool optimized for:
// - Quick system identification
// - Resource monitoring
// - Environment discovery
// - Cross-platform compatibility
func GetSystemInfo() domain.Tool {
	return atools.NewTool(
		"get_system_info",
		"Retrieves comprehensive system information",
		func(ctx *domain.ToolContext, params GetSystemInfoParams) (*SystemInfo, error) {
			// Emit start event
			if ctx.Events != nil {
				ctx.Events.Emit(domain.EventToolCall, domain.ToolCallEventData{
					ToolName:   "get_system_info",
					Parameters: params,
					RequestID:  ctx.RunID,
				})
			}
			info := &SystemInfo{
				OS: OSInfo{
					Name:     runtime.GOOS,
					Platform: getPlatformName(runtime.GOOS),
				},
				Architecture: runtime.GOARCH,
				CPUs:         runtime.NumCPU(),
				Timestamp:    time.Now().UTC().Format(time.RFC3339),
			}

			// Get hostname
			if hostname, err := os.Hostname(); err == nil {
				info.Hostname = hostname
			}

			// Get OS version if available
			if version := getOSVersion(); version != "" {
				info.OS.Version = version
			}

			// Check state for default includes if not explicitly set
			if ctx.State != nil {
				if !params.IncludeMemory {
					if val, ok := ctx.State.Get("system_info_include_memory"); ok {
						if b, ok := val.(bool); ok {
							params.IncludeMemory = b
						}
					}
				}
				if !params.IncludeRuntime {
					if val, ok := ctx.State.Get("system_info_include_runtime"); ok {
						if b, ok := val.(bool); ok {
							params.IncludeRuntime = b
						}
					}
				}
				if !params.IncludeEnvironment {
					if val, ok := ctx.State.Get("system_info_include_environment"); ok {
						if b, ok := val.(bool); ok {
							params.IncludeEnvironment = b
						}
					}
				}
			}
			
			// Include memory statistics if requested
			if params.IncludeMemory {
				info.Memory = getMemoryInfo()
			}

			// Include runtime information if requested
			if params.IncludeRuntime {
				info.Runtime = getRuntimeInfo()
			}

			// Include environment summary if requested
			if params.IncludeEnvironment {
				info.Environment = getEnvironmentInfo()
			}
			
			// Emit result event
			if ctx.Events != nil {
				ctx.Events.Emit(domain.EventToolResult, domain.ToolResultEventData{
					ToolName:  "get_system_info",
					Result:    info,
					RequestID: ctx.RunID,
				})
			}

			return info, nil
		},
		getSystemInfoParamSchema,
	)
}

// getPlatformName returns a human-readable platform name
func getPlatformName(goos string) string {
	if name, ok := platformNames[goos]; ok {
		return name
	}
	return goos
}

// getOSVersion attempts to get OS version information
func getOSVersion() string {
	// This is platform-specific and would need different implementations
	// For now, we'll use a simple approach
	switch runtime.GOOS {
	case "darwin":
		// On macOS, we could read from /System/Library/CoreServices/SystemVersion.plist
		// but that requires parsing plist files
		return ""
	case "linux":
		// On Linux, we could read from /etc/os-release
		if data, err := os.ReadFile("/etc/os-release"); err == nil {
			// Simple extraction of VERSION or PRETTY_NAME
			lines := string(data)
			for _, line := range splitLines(lines) {
				if hasPrefix(line, "PRETTY_NAME=") {
					return trimQuotes(line[12:])
				}
			}
		}
		return ""
	case "windows":
		// On Windows, this would require Windows API calls
		return ""
	default:
		return ""
	}
}

// getMemoryInfo returns current memory statistics
func getMemoryInfo() *MemoryInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &MemoryInfo{
		Alloc:      m.Alloc,
		TotalAlloc: m.TotalAlloc,
		Sys:        m.Sys,
		NumGC:      m.NumGC,
	}
}

// getRuntimeInfo returns Go runtime information
func getRuntimeInfo() *RuntimeInfo {
	return &RuntimeInfo{
		Version:      runtime.Version(),
		Compiler:     runtime.Compiler,
		NumCPU:       runtime.NumCPU(),
		NumGoroutine: runtime.NumGoroutine(),
		GOMAXPROCS:   runtime.GOMAXPROCS(0),
	}
}

// getEnvironmentInfo returns environment summary
func getEnvironmentInfo() *EnvironmentInfo {
	info := &EnvironmentInfo{
		User:         os.Getenv("USER"),
		Home:         os.Getenv("HOME"),
		TempDir:      os.TempDir(),
		TotalEnvVars: len(os.Environ()),
		GoPath:       os.Getenv("GOPATH"),
		GoRoot:       os.Getenv("GOROOT"),
	}

	// Get working directory
	if wd, err := os.Getwd(); err == nil {
		info.WorkingDir = wd
	}

	// Count PATH directories
	if path := os.Getenv("PATH"); path != "" {
		info.PathDirs = countPathDirs(path)
	}

	return info
}

// countPathDirs counts the number of directories in PATH
func countPathDirs(path string) int {
	if path == "" {
		return 0
	}

	separator := ":"
	if runtime.GOOS == "windows" {
		separator = ";"
	}

	count := 1
	for i := 0; i < len(path); i++ {
		if path[i] == separator[0] {
			count++
		}
	}
	return count
}

// Helper functions to avoid importing strings package
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func trimQuotes(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

// MustGetSystemInfo retrieves the registered GetSystemInfo tool or panics
// This is a convenience function for users who want to ensure the tool exists
func MustGetSystemInfo() domain.Tool {
	return tools.MustGetTool("get_system_info")
}
