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
			Enum:        []string{"pid", "name", "cpu", "memory"},
		},
		"limit": {
			Type:        "integer",
			Description: "Maximum number of processes to return",
			Minimum:     floatPtr(1),
			Maximum:     floatPtr(1000),
		},
	},
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

// ProcessList creates a tool for listing running processes
// This is a built-in tool optimized for:
// - Cross-platform process enumeration
// - Process filtering and sorting
// - Resource usage monitoring
// - System diagnostics
func ProcessList() domain.Tool {
	return atools.NewTool(
		"process_list",
		"Lists running processes with filtering and sorting",
		func(ctx *domain.ToolContext, params ProcessListParams) (*ProcessListResult, error) {
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
		},
		processListParamSchema,
	)
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
