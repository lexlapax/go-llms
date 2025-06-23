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

// getSystemInfoOutputSchema defines the output schema for the GetSystemInfo tool
var getSystemInfoOutputSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"os": {
			Type:        "object",
			Description: "Operating system details",
			Properties: map[string]sdomain.Property{
				"name": {
					Type:        "string",
					Description: "OS name (e.g., darwin, linux, windows)",
				},
				"platform": {
					Type:        "string",
					Description: "Human-readable platform name",
				},
				"version": {
					Type:        "string",
					Description: "OS version if available",
				},
			},
			Required: []string{"name", "platform"},
		},
		"architecture": {
			Type:        "string",
			Description: "System architecture (e.g., amd64, arm64)",
		},
		"cpus": {
			Type:        "number",
			Description: "Number of CPU cores",
		},
		"hostname": {
			Type:        "string",
			Description: "System hostname",
		},
		"memory": {
			Type:        "object",
			Description: "Memory statistics (if include_memory is true)",
			Properties: map[string]sdomain.Property{
				"alloc": {
					Type:        "number",
					Description: "Bytes allocated and in use",
				},
				"total_alloc": {
					Type:        "number",
					Description: "Total bytes allocated (even if freed)",
				},
				"sys": {
					Type:        "number",
					Description: "Bytes obtained from system",
				},
				"num_gc": {
					Type:        "number",
					Description: "Number of completed GC cycles",
				},
			},
		},
		"runtime": {
			Type:        "object",
			Description: "Go runtime information (if include_runtime is true)",
			Properties: map[string]sdomain.Property{
				"version": {
					Type:        "string",
					Description: "Go version",
				},
				"compiler": {
					Type:        "string",
					Description: "Go compiler",
				},
				"num_cpu": {
					Type:        "number",
					Description: "Number of logical CPUs",
				},
				"num_goroutine": {
					Type:        "number",
					Description: "Current number of goroutines",
				},
				"gomaxprocs": {
					Type:        "number",
					Description: "GOMAXPROCS setting",
				},
			},
		},
		"environment": {
			Type:        "object",
			Description: "Environment summary (if include_environment is true)",
			Properties: map[string]sdomain.Property{
				"user": {
					Type:        "string",
					Description: "Current user",
				},
				"home": {
					Type:        "string",
					Description: "User home directory",
				},
				"working_dir": {
					Type:        "string",
					Description: "Current working directory",
				},
				"temp_dir": {
					Type:        "string",
					Description: "System temporary directory",
				},
				"path_dirs": {
					Type:        "number",
					Description: "Number of directories in PATH",
				},
				"total_env_vars": {
					Type:        "number",
					Description: "Total number of environment variables",
				},
				"gopath": {
					Type:        "string",
					Description: "GOPATH environment variable",
				},
				"goroot": {
					Type:        "string",
					Description: "GOROOT environment variable",
				},
			},
		},
		"timestamp": {
			Type:        "string",
			Description: "Timestamp when information was collected (RFC3339)",
		},
	},
	Required: []string{"os", "architecture", "cpus", "timestamp"},
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

// getSystemInfo is the main function for the tool
func getSystemInfo(ctx *domain.ToolContext, params GetSystemInfoParams) (*SystemInfo, error) {
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
}

// GetSystemInfo creates a tool for retrieving comprehensive system information including OS details,
// architecture, CPU count, hostname, and optionally memory statistics, Go runtime information, and
// environment summary. The tool provides cross-platform compatibility and returns consistent structured
// data suitable for system identification, resource monitoring, and environment discovery.
func GetSystemInfo() domain.Tool {
	builder := atools.NewToolBuilder("get_system_info", "Retrieves comprehensive system information").
		WithFunction(getSystemInfo).
		WithParameterSchema(getSystemInfoParamSchema).
		WithOutputSchema(getSystemInfoOutputSchema).
		WithUsageInstructions(`Use this tool to retrieve comprehensive information about the system.

By default, returns basic system information:
- Operating system (name, platform, version if available)
- Architecture (e.g., amd64, arm64)
- Number of CPUs
- Hostname
- Timestamp

Optional information can be included:
- include_memory: Memory statistics (allocated, system, GC stats)
- include_runtime: Go runtime information (version, goroutines, GOMAXPROCS)
- include_environment: Environment summary (user, paths, env var count)

Parameters:
- include_environment: Add environment summary (optional, default false)
- include_memory: Add memory statistics (optional, default false)
- include_runtime: Add Go runtime info (optional, default false)

Memory statistics include:
- alloc: Current memory allocated and in use
- total_alloc: Total memory allocated since program start
- sys: Memory obtained from the OS
- num_gc: Number of garbage collection cycles

Runtime information includes:
- Go version and compiler
- Number of logical CPUs
- Current goroutine count
- GOMAXPROCS setting

Environment summary includes:
- Current user and home directory
- Working directory and temp directory
- PATH directory count
- Total environment variables
- Go-specific paths (GOPATH, GOROOT)

Note: You can set defaults via state:
- state.Set("system_info_include_memory", true)
- state.Set("system_info_include_runtime", true)
- state.Set("system_info_include_environment", true)`).
		WithExamples([]domain.ToolExample{
			{
				Name:        "Basic system information",
				Description: "Get core system details",
				Scenario:    "When you need to identify the system",
				Input:       map[string]interface{}{},
				Output: map[string]interface{}{
					"os": map[string]interface{}{
						"name":     "linux",
						"platform": "Linux",
					},
					"architecture": "amd64",
					"cpus":         8,
					"hostname":     "dev-machine",
					"timestamp":    "2024-01-15T10:30:00Z",
				},
				Explanation: "Returns essential system identification without optional data",
			},
			{
				Name:        "System with memory statistics",
				Description: "Include current memory usage",
				Scenario:    "When monitoring resource usage",
				Input: map[string]interface{}{
					"include_memory": true,
				},
				Output: map[string]interface{}{
					"os": map[string]interface{}{
						"name":     "darwin",
						"platform": "macOS",
					},
					"architecture": "arm64",
					"cpus":         10,
					"memory": map[string]interface{}{
						"alloc":       52428800,
						"total_alloc": 104857600,
						"sys":         75497472,
						"num_gc":      5,
					},
					"timestamp": "2024-01-15T10:30:00Z",
				},
				Explanation: "Memory stats show current allocation and GC information",
			},
			{
				Name:        "Full system information",
				Description: "Get all available system details",
				Scenario:    "When you need comprehensive system analysis",
				Input: map[string]interface{}{
					"include_memory":      true,
					"include_runtime":     true,
					"include_environment": true,
				},
				Output: map[string]interface{}{
					"os": map[string]interface{}{
						"name":     "linux",
						"platform": "Linux",
						"version":  "Ubuntu 22.04.3 LTS",
					},
					"architecture": "amd64",
					"cpus":         16,
					"hostname":     "prod-server",
					"memory": map[string]interface{}{
						"alloc":       104857600,
						"total_alloc": 209715200,
						"sys":         150994944,
						"num_gc":      10,
					},
					"runtime": map[string]interface{}{
						"version":       "go1.21.5",
						"compiler":      "gc",
						"num_cpu":       16,
						"num_goroutine": 42,
						"gomaxprocs":    16,
					},
					"environment": map[string]interface{}{
						"user":           "appuser",
						"home":           "/home/appuser",
						"working_dir":    "/app",
						"temp_dir":       "/tmp",
						"path_dirs":      12,
						"total_env_vars": 35,
						"gopath":         "/home/appuser/go",
						"goroot":         "/usr/local/go",
					},
					"timestamp": "2024-01-15T10:30:00Z",
				},
				Explanation: "Complete system snapshot with all optional data included",
			},
			{
				Name:        "Runtime monitoring",
				Description: "Check Go runtime statistics",
				Scenario:    "When debugging goroutine leaks or performance",
				Input: map[string]interface{}{
					"include_runtime": true,
				},
				Output: map[string]interface{}{
					"os": map[string]interface{}{
						"name":     "windows",
						"platform": "Windows",
					},
					"architecture": "amd64",
					"cpus":         12,
					"hostname":     "WIN-DEV",
					"runtime": map[string]interface{}{
						"version":       "go1.21.5",
						"compiler":      "gc",
						"num_cpu":       12,
						"num_goroutine": 156,
						"gomaxprocs":    12,
					},
					"timestamp": "2024-01-15T10:30:00Z",
				},
				Explanation: "High goroutine count might indicate a leak",
			},
			{
				Name:        "Environment check",
				Description: "Verify environment configuration",
				Scenario:    "When troubleshooting path or environment issues",
				Input: map[string]interface{}{
					"include_environment": true,
				},
				Output: map[string]interface{}{
					"os": map[string]interface{}{
						"name":     "darwin",
						"platform": "macOS",
					},
					"architecture": "arm64",
					"cpus":         8,
					"environment": map[string]interface{}{
						"user":           "developer",
						"home":           "/Users/developer",
						"working_dir":    "/Users/developer/projects",
						"temp_dir":       "/var/folders/xx/yyyyyy/T",
						"path_dirs":      15,
						"total_env_vars": 52,
						"gopath":         "/Users/developer/go",
						"goroot":         "/opt/homebrew/opt/go/libexec",
					},
					"timestamp": "2024-01-15T10:30:00Z",
				},
				Explanation: "Shows environment paths and configuration",
			},
		}).
		WithConstraints([]string{
			"Cannot modify system information, only read it",
			"Memory values are in bytes",
			"OS version detection is platform-specific and may not always be available",
			"Runtime statistics reflect the current process, not the entire system",
			"Environment variables shown are from the process environment",
			"Timestamp is always in UTC with RFC3339 format",
			"GOMAXPROCS reflects the current setting, not necessarily CPU count",
			"Some fields may be empty on certain platforms",
		}).
		WithErrorGuidance(map[string]string{
			"permission denied": "Some system information may require elevated permissions",
			"not supported":     "Some features may not be available on all platforms",
		}).
		WithCategory("system").
		WithTags([]string{"system", "info", "os", "architecture", "resources"}).
		WithVersion("2.0.0").
		WithBehavior(
			true,   // Deterministic at a point in time
			false,  // Not destructive - only reads
			false,  // No confirmation needed
			"fast", // Very fast operation
		)

	return builder.Build()
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
