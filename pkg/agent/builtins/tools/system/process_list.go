// ABOUTME: Process list tool for retrieving information about running processes
// ABOUTME: Built-in tool providing cross-platform process enumeration and filtering

package system

import (
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	atools "github.com/lexlapax/go-llms/pkg/agent/tools"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// ProcessListParams defines parameters for the ProcessList tool
type ProcessListParams struct {
	Filter      string `json:"filter,omitempty"`       // Filter by process name (contains)
	IncludeSelf bool   `json:"include_self,omitempty"` // Include current process
	SortBy      string `json:"sort_by,omitempty"`      // Sort by: pid, name, cpu, memory
	Limit       int    `json:"limit,omitempty"`        // Limit number of results
}

// ProcessListResult represents the result of process listing
type ProcessListResult struct {
	Processes []ProcessInfo `json:"processes"`
	Count     int           `json:"count"`
	Platform  string        `json:"platform"`
	Timestamp string        `json:"timestamp"`
}

// ProcessInfo represents information about a single process
type ProcessInfo struct {
	PID         int     `json:"pid"`
	Name        string  `json:"name"`
	Command     string  `json:"command,omitempty"`
	CPUPercent  float64 `json:"cpu_percent,omitempty"`
	MemoryUsage int64   `json:"memory_kb,omitempty"`
	User        string  `json:"user,omitempty"`
	StartTime   string  `json:"start_time,omitempty"`
}

// processListParamSchema defines parameters for the ProcessList tool
var processListParamSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"filter": {
			Type:        "string",
			Description: "Filter processes by name (case-insensitive contains)",
		},
		"include_self": {
			Type:        "boolean",
			Description: "Include the current process in results",
		},
		"sort_by": {
			Type:        "string",
			Description: "Sort results by: pid, name, cpu, memory",
		},
		"limit": {
			Type:        "number",
			Description: "Maximum number of processes to return",
			Minimum:     floatPtr(1),
			Maximum:     floatPtr(1000),
		},
	},
}

// processListOutputSchema defines the output schema for the ProcessList tool
var processListOutputSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"processes": {
			Type:        "array",
			Description: "List of running processes",
			Items: &sdomain.Property{
				Type: "object",
				Properties: map[string]sdomain.Property{
					"pid": {
						Type:        "number",
						Description: "Process ID",
					},
					"name": {
						Type:        "string",
						Description: "Process name",
					},
					"command": {
						Type:        "string",
						Description: "Full command line",
					},
					"cpu_percent": {
						Type:        "number",
						Description: "CPU usage percentage",
					},
					"memory_kb": {
						Type:        "number",
						Description: "Memory usage in kilobytes",
					},
					"user": {
						Type:        "string",
						Description: "Process owner",
					},
					"start_time": {
						Type:        "string",
						Description: "Process start time",
					},
				},
				Required: []string{"pid", "name"},
			},
		},
		"count": {
			Type:        "number",
			Description: "Number of processes returned",
		},
		"platform": {
			Type:        "string",
			Description: "Operating system platform",
		},
		"timestamp": {
			Type:        "string",
			Description: "Timestamp when the list was generated (RFC3339)",
		},
	},
	Required: []string{"processes", "count", "platform", "timestamp"},
}

// init automatically registers the tool on package import
func init() {
	tools.MustRegisterTool("process_list", ProcessList(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "process_list",
			Category:    "system",
			Tags:        []string{"process", "system", "monitoring", "ps"},
			Description: "Lists running processes with filtering and sorting",
			Version:     "1.0.0",
			Examples: []builtins.Example{
				{
					Name:        "List all processes",
					Description: "Get a list of all running processes",
					Code:        `ProcessList().Execute(ctx, ProcessListParams{})`,
				},
				{
					Name:        "Filter by name",
					Description: "Find processes containing 'chrome' in the name",
					Code:        `ProcessList().Execute(ctx, ProcessListParams{Filter: "chrome"})`,
				},
				{
					Name:        "Top CPU processes",
					Description: "Get top 10 processes by CPU usage",
					Code:        `ProcessList().Execute(ctx, ProcessListParams{SortBy: "cpu", Limit: 10})`,
				},
			},
		},
		RequiredPermissions: []string{"system:read"},
		ResourceUsage: tools.ResourceInfo{
			Memory:      "medium",
			Network:     false,
			FileSystem:  false,
			Concurrency: true,
		},
	})
}

// processList is the main function for the tool
func processList(ctx *domain.ToolContext, params ProcessListParams) (*ProcessListResult, error) {
	// Emit start event
	if ctx.Events != nil {
		ctx.Events.Emit(domain.EventToolCall, domain.ToolCallEventData{
			ToolName:   "process_list",
			Parameters: params,
			RequestID:  ctx.RunID,
		})
	}
	// Get current PID for self-filtering
	currentPID := os.Getpid()

	// Check state for default limit if not provided
	if params.Limit == 0 && ctx.State != nil {
		if limit, ok := ctx.State.Get("process_list_default_limit"); ok {
			if l, ok := limit.(int); ok && l > 0 {
				params.Limit = l
			}
		}
	}

	// Get process list based on platform
	var processes []ProcessInfo
	var err error

	switch runtime.GOOS {
	case "darwin", "linux", "freebsd", "openbsd", "netbsd":
		processes, err = getUnixProcesses()
	case "windows":
		processes, err = getWindowsProcesses()
	default:
		processes, err = getFallbackProcesses()
	}

	if err != nil {
		return nil, err
	}

	// Filter processes
	if params.Filter != "" || !params.IncludeSelf {
		filtered := make([]ProcessInfo, 0, len(processes))
		filterUpper := strings.ToUpper(params.Filter)

		for _, p := range processes {
			// Skip self if requested
			if !params.IncludeSelf && p.PID == currentPID {
				continue
			}

			// Apply name filter
			if params.Filter != "" {
				nameUpper := strings.ToUpper(p.Name)
				commandUpper := strings.ToUpper(p.Command)
				if !strings.Contains(nameUpper, filterUpper) &&
					!strings.Contains(commandUpper, filterUpper) {
					continue
				}
			}

			filtered = append(filtered, p)
		}
		processes = filtered
	}

	// Sort if requested
	if params.SortBy != "" {
		sortProcesses(processes, params.SortBy)
	}

	// Apply limit
	if params.Limit > 0 && len(processes) > params.Limit {
		processes = processes[:params.Limit]
	}

	result := &ProcessListResult{
		Processes: processes,
		Count:     len(processes),
		Platform:  runtime.GOOS,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// Emit result event
	if ctx.Events != nil {
		ctx.Events.Emit(domain.EventToolResult, domain.ToolResultEventData{
			ToolName:  "process_list",
			Result:    result,
			RequestID: ctx.RunID,
		})
	}

	return result, nil
}

// ProcessList creates a tool for listing and analyzing running system processes across different platforms
// (Unix/Linux/macOS/Windows) with support for filtering by name, sorting by various metrics (PID, CPU, memory),
// and retrieving detailed process information including resource usage. The tool provides consistent output
// across platforms while handling platform-specific limitations gracefully.
func ProcessList() domain.Tool {
	builder := atools.NewToolBuilder("process_list", "Lists running processes with filtering and sorting").
		WithFunction(processList).
		WithParameterSchema(processListParamSchema).
		WithOutputSchema(processListOutputSchema).
		WithUsageInstructions(`Use this tool to list and analyze running processes on the system.

Cross-Platform Support:
- Unix/Linux/macOS: Uses 'ps aux' command for detailed process information
- Windows: Uses 'tasklist' command (limited CPU info)
- Other platforms: Returns minimal process information

Parameters:
- filter: Search for processes by name (case-insensitive, partial match)
- include_self: Include the current process (default: false)
- sort_by: Order results by pid, name, cpu, or memory
- limit: Maximum processes to return (1-1000)

Output Information:
- pid: Process identifier
- name: Process executable name
- command: Full command line (Unix only)
- cpu_percent: CPU usage percentage (Unix only)
- memory_kb: Memory usage in kilobytes
- user: Process owner
- start_time: When process started

Filtering:
- Searches both process name and command line
- Case-insensitive partial matching
- Example: filter "chrome" matches "Google Chrome Helper"

Sorting:
- pid: Ascending by process ID
- name: Alphabetical by process name
- cpu: Descending by CPU usage (highest first)
- memory: Descending by memory usage (highest first)

State Configuration:
Set default limit via state:
state.Set("process_list_default_limit", 50)

Platform Notes:
- CPU percentage may be 0 on Windows
- Command field may be empty on some platforms
- Memory values are estimates on some systems`).
		WithExamples([]domain.ToolExample{
			{
				Name:        "List all processes",
				Description: "Get a complete process list",
				Scenario:    "When you need to see all running processes",
				Input:       map[string]interface{}{},
				Output: map[string]interface{}{
					"processes": []map[string]interface{}{
						{
							"pid":         1234,
							"name":        "chrome",
							"command":     "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
							"cpu_percent": 5.2,
							"memory_kb":   524288,
							"user":        "john",
							"start_time":  "10:30AM",
						},
						{
							"pid":         5678,
							"name":        "code",
							"command":     "/usr/local/bin/code",
							"cpu_percent": 2.1,
							"memory_kb":   262144,
							"user":        "john",
							"start_time":  "09:15AM",
						},
					},
					"count":     2,
					"platform":  "darwin",
					"timestamp": "2024-01-15T10:30:00Z",
				},
				Explanation: "Returns all processes with available information",
			},
			{
				Name:        "Find specific processes",
				Description: "Search for Chrome processes",
				Scenario:    "When troubleshooting a specific application",
				Input: map[string]interface{}{
					"filter": "chrome",
				},
				Output: map[string]interface{}{
					"processes": []map[string]interface{}{
						{
							"pid":       1234,
							"name":      "chrome",
							"memory_kb": 524288,
						},
						{
							"pid":       1235,
							"name":      "chrome_crashpad",
							"memory_kb": 8192,
						},
					},
					"count":     2,
					"platform":  "linux",
					"timestamp": "2024-01-15T10:30:00Z",
				},
				Explanation: "Filter matches any process containing 'chrome' in name or command",
			},
			{
				Name:        "Top CPU consumers",
				Description: "Find processes using most CPU",
				Scenario:    "When system is running slow",
				Input: map[string]interface{}{
					"sort_by": "cpu",
					"limit":   5,
				},
				Output: map[string]interface{}{
					"processes": []map[string]interface{}{
						{"pid": 9999, "name": "video_encoder", "cpu_percent": 95.5},
						{"pid": 8888, "name": "chrome", "cpu_percent": 45.2},
						{"pid": 7777, "name": "spotlight", "cpu_percent": 25.0},
						{"pid": 6666, "name": "docker", "cpu_percent": 15.3},
						{"pid": 5555, "name": "vscode", "cpu_percent": 10.1},
					},
					"count":     5,
					"platform":  "darwin",
					"timestamp": "2024-01-15T10:30:00Z",
				},
				Explanation: "Sorted by CPU usage descending, limited to top 5",
			},
			{
				Name:        "Memory usage analysis",
				Description: "Find memory-hungry processes",
				Scenario:    "When investigating high memory usage",
				Input: map[string]interface{}{
					"sort_by": "memory",
					"limit":   10,
				},
				Output: map[string]interface{}{
					"processes": []map[string]interface{}{
						{"pid": 1111, "name": "docker", "memory_kb": 2097152},
						{"pid": 2222, "name": "chrome", "memory_kb": 1048576},
						{"pid": 3333, "name": "slack", "memory_kb": 524288},
					},
					"count":     3,
					"platform":  "linux",
					"timestamp": "2024-01-15T10:30:00Z",
				},
				Explanation: "Shows top memory consumers in descending order",
			},
			{
				Name:        "Include current process",
				Description: "List processes including self",
				Scenario:    "When debugging the current application",
				Input: map[string]interface{}{
					"include_self": true,
					"filter":       "go",
				},
				Output: map[string]interface{}{
					"processes": []map[string]interface{}{
						{"pid": 12345, "name": "go", "command": "go run main.go"},
						{"pid": 12346, "name": "gopls", "command": "gopls serve"},
						{"pid": os.Getpid(), "name": "go-llms", "command": "./go-llms"},
					},
					"count":     3,
					"platform":  "linux",
					"timestamp": "2024-01-15T10:30:00Z",
				},
				Explanation: "include_self: true includes the current process in results",
			},
			{
				Name:        "Windows process list",
				Description: "List processes on Windows",
				Scenario:    "When running on Windows platform",
				Input: map[string]interface{}{
					"limit": 3,
				},
				Output: map[string]interface{}{
					"processes": []map[string]interface{}{
						{"pid": 1000, "name": "chrome.exe", "memory_kb": 512000, "user": "SYSTEM"},
						{"pid": 2000, "name": "svchost.exe", "memory_kb": 64000, "user": "SYSTEM"},
						{"pid": 3000, "name": "explorer.exe", "memory_kb": 128000, "user": "User"},
					},
					"count":     3,
					"platform":  "windows",
					"timestamp": "2024-01-15T10:30:00Z",
				},
				Explanation: "Windows may have limited CPU info but includes memory and user",
			},
		}).
		WithConstraints([]string{
			"Requires appropriate permissions to see all processes",
			"CPU percentages may be instantaneous snapshots",
			"Memory values are estimates and may vary by platform",
			"Some fields may be empty depending on OS and permissions",
			"Process information is a point-in-time snapshot",
			"Filtering is case-insensitive partial string matching",
			"Sort by CPU/memory may not work on all platforms",
			"Maximum limit is 1000 processes",
			"Windows has limited CPU usage information",
			"Start time format varies by platform",
		}).
		WithErrorGuidance(map[string]string{
			"command not found":      "ps or tasklist command not available on this system",
			"permission denied":      "Insufficient permissions to list processes. May need elevated privileges",
			"invalid sort field":     "Use one of: pid, name, cpu, memory",
			"limit out of range":     "Limit must be between 1 and 1000",
			"no processes found":     "No processes match the filter criteria",
			"platform not supported": "Process listing not available on this platform",
		}).
		WithCategory("system").
		WithTags([]string{"process", "system", "monitoring", "ps"}).
		WithVersion("2.0.0").
		WithBehavior(
			false,    // Not deterministic - process list changes
			false,    // Not destructive - only reads
			false,    // No confirmation needed
			"medium", // Medium latency - depends on process count
		)

	return builder.Build()
}

// getUnixProcesses gets process list on Unix-like systems
func getUnixProcesses() ([]ProcessInfo, error) {
	// Use ps command with specific format
	// Format: PID %CPU %MEM USER STARTED COMMAND
	cmd := exec.Command("ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	processes := make([]ProcessInfo, 0, len(lines))

	// Skip header line
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		// Parse ps output
		fields := strings.Fields(line)
		if len(fields) < 11 {
			continue
		}

		pid, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}

		cpu, _ := strconv.ParseFloat(fields[2], 64)
		mem, _ := strconv.ParseFloat(fields[3], 64)

		// Memory is in percentage, convert to approximate KB
		// This is a rough estimate based on system memory
		memKB := int64(mem * 1024) // Simplified conversion

		// Command is everything from field 10 onwards
		command := strings.Join(fields[10:], " ")

		// Extract process name from command
		name := extractProcessName(command)

		process := ProcessInfo{
			PID:         pid,
			Name:        name,
			Command:     command,
			CPUPercent:  cpu,
			MemoryUsage: memKB,
			User:        fields[0],
			StartTime:   fields[8], // This is the START column
		}

		processes = append(processes, process)
	}

	return processes, nil
}

// getWindowsProcesses gets process list on Windows
func getWindowsProcesses() ([]ProcessInfo, error) {
	// Use tasklist command
	cmd := exec.Command("tasklist", "/fo", "csv", "/v")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	processes := make([]ProcessInfo, 0, len(lines))

	// Skip header line
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		// Parse CSV output
		fields := parseCSVLine(line)
		if len(fields) < 9 {
			continue
		}

		// Extract PID from second field
		pidStr := strings.Trim(fields[1], "\"")
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}

		// Extract memory usage (in K)
		memStr := strings.Trim(fields[5], "\"")
		memStr = strings.ReplaceAll(memStr, ",", "")
		memStr = strings.ReplaceAll(memStr, " K", "")
		memKB, _ := strconv.ParseInt(memStr, 10, 64)

		process := ProcessInfo{
			PID:         pid,
			Name:        strings.Trim(fields[0], "\""),
			MemoryUsage: memKB,
			User:        strings.Trim(fields[6], "\""),
		}

		processes = append(processes, process)
	}

	return processes, nil
}

// getFallbackProcesses provides a minimal process list
func getFallbackProcesses() ([]ProcessInfo, error) {
	// At minimum, return the current process
	currentPID := os.Getpid()

	processes := []ProcessInfo{
		{
			PID:  currentPID,
			Name: "go-llms",
		},
	}

	return processes, nil
}

// extractProcessName extracts the process name from a command line
func extractProcessName(command string) string {
	if command == "" {
		return "unknown"
	}

	// Remove leading/trailing spaces
	command = strings.TrimSpace(command)

	// Split by space to get the executable
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "unknown"
	}

	// Get the base name (last component of path)
	executable := parts[0]

	// Handle both Unix and Windows path separators
	lastSlash := strings.LastIndex(executable, "/")
	lastBackslash := strings.LastIndex(executable, "\\")

	// Use whichever separator appears last
	lastSep := lastSlash
	if lastBackslash > lastSep {
		lastSep = lastBackslash
	}

	if lastSep >= 0 {
		executable = executable[lastSep+1:]
	}

	// Remove common suffixes
	executable = strings.TrimSuffix(executable, ".exe")

	return executable
}

// parseCSVLine does simple CSV parsing (handles quoted fields)
func parseCSVLine(line string) []string {
	var fields []string
	var current strings.Builder
	inQuotes := false

	for i := 0; i < len(line); i++ {
		ch := line[i]

		if ch == '"' {
			inQuotes = !inQuotes
		} else if ch == ',' && !inQuotes {
			fields = append(fields, current.String())
			current.Reset()
		} else {
			current.WriteByte(ch)
		}
	}

	// Add the last field
	if current.Len() > 0 {
		fields = append(fields, current.String())
	}

	return fields
}

// sortProcesses sorts the process list by the specified field
func sortProcesses(processes []ProcessInfo, sortBy string) {
	switch sortBy {
	case "pid":
		sortByPID(processes)
	case "name":
		sortByName(processes)
	case "cpu":
		sortByCPU(processes)
	case "memory":
		sortByMemory(processes)
	}
}

// Sort functions
func sortByPID(processes []ProcessInfo) {
	for i := 0; i < len(processes)-1; i++ {
		for j := i + 1; j < len(processes); j++ {
			if processes[i].PID > processes[j].PID {
				processes[i], processes[j] = processes[j], processes[i]
			}
		}
	}
}

func sortByName(processes []ProcessInfo) {
	for i := 0; i < len(processes)-1; i++ {
		for j := i + 1; j < len(processes); j++ {
			if strings.ToLower(processes[i].Name) > strings.ToLower(processes[j].Name) {
				processes[i], processes[j] = processes[j], processes[i]
			}
		}
	}
}

func sortByCPU(processes []ProcessInfo) {
	for i := 0; i < len(processes)-1; i++ {
		for j := i + 1; j < len(processes); j++ {
			if processes[i].CPUPercent < processes[j].CPUPercent {
				processes[i], processes[j] = processes[j], processes[i]
			}
		}
	}
}

func sortByMemory(processes []ProcessInfo) {
	for i := 0; i < len(processes)-1; i++ {
		for j := i + 1; j < len(processes); j++ {
			if processes[i].MemoryUsage < processes[j].MemoryUsage {
				processes[i], processes[j] = processes[j], processes[i]
			}
		}
	}
}

// floatPtr is a helper to create float64 pointers for schema
func floatPtr(f float64) *float64 {
	return &f
}

// MustProcessList retrieves the registered ProcessList tool or panics
// This is a convenience function for users who want to ensure the tool exists
func MustProcessList() domain.Tool {
	return tools.MustGetTool("process_list")
}
