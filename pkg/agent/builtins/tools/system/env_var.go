// ABOUTME: Environment variable tool for safe reading of system environment
// ABOUTME: Built-in tool supporting single var lookup, pattern matching, and filtering

package system

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	atools "github.com/lexlapax/go-llms/pkg/agent/tools"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// GetEnvironmentVariableParams defines parameters for the GetEnvironmentVariable tool
type GetEnvironmentVariableParams struct {
	Name      string `json:"name,omitempty"`      // Specific variable name
	Pattern   string `json:"pattern,omitempty"`   // Pattern to match variable names
	NoValues  bool   `json:"no_values,omitempty"` // Exclude values from results (default: false)
	Sensitive bool   `json:"sensitive,omitempty"` // Allow sensitive variables (PATH, API keys, etc.)
}

// EnvironmentVariable represents a single environment variable
type EnvironmentVariable struct {
	Name   string `json:"name"`
	Value  string `json:"value,omitempty"`
	Masked bool   `json:"masked,omitempty"` // True if value was masked for security
}

// GetEnvironmentVariableResult defines the result of the GetEnvironmentVariable tool
type GetEnvironmentVariableResult struct {
	Variables []EnvironmentVariable `json:"variables"`
	Count     int                   `json:"count"`
	Query     string                `json:"query,omitempty"` // The name or pattern searched
}

// getEnvironmentVariableParamSchema defines parameters for the GetEnvironmentVariable tool
var getEnvironmentVariableParamSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"name": {
			Type:        "string",
			Description: "Specific environment variable name to retrieve",
		},
		"pattern": {
			Type:        "string",
			Description: "Pattern to match variable names (e.g., 'GO_*', '*_PATH')",
		},
		"no_values": {
			Type:        "boolean",
			Description: "Exclude values from results (default: false, meaning values are included)",
		},
		"sensitive": {
			Type:        "boolean",
			Description: "Allow retrieval of potentially sensitive variables",
		},
	},
}

// Sensitive variable patterns that should be masked by default
var sensitivePatterns = []string{
	"*KEY*", "*SECRET*", "*TOKEN*", "*PASSWORD*", "*PASS*",
	"*CREDENTIAL*", "*AUTH*", "*API*", "*PRIVATE*",
}

// init automatically registers the tool on package import
func init() {
	tools.MustRegisterTool("get_environment_variable", GetEnvironmentVariable(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "get_environment_variable",
			Category:    "system",
			Tags:        []string{"environment", "config", "system", "variables"},
			Description: "Retrieves environment variables safely",
			Version:     "1.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Get specific variable",
					Description: "Retrieve a specific environment variable",
					Code:        `GetEnvironmentVariable().Execute(ctx, GetEnvironmentVariableParams{Name: "HOME"})`,
				},
				{
					Name:        "List Go variables",
					Description: "Find all Go-related environment variables",
					Code:        `GetEnvironmentVariable().Execute(ctx, GetEnvironmentVariableParams{Pattern: "GO*"})`,
				},
				{
					Name:        "List variables without values",
					Description: "List variable names only",
					Code:        `GetEnvironmentVariable().Execute(ctx, GetEnvironmentVariableParams{Pattern: "*", NoValues: true})`,
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

// GetEnvironmentVariable creates a tool for safely reading environment variables
// This is a built-in tool optimized for:
// - Safe access to environment configuration
// - Pattern-based variable discovery
// - Security through sensitive variable masking
// - Filtered results for specific use cases
func GetEnvironmentVariable() domain.Tool {
	return atools.NewTool(
		"get_environment_variable",
		"Retrieves environment variables safely",
		func(ctx *domain.ToolContext, params GetEnvironmentVariableParams) (*GetEnvironmentVariableResult, error) {
			// Emit start event
			if ctx.Events != nil {
				ctx.Events.Emit(domain.EventToolCall, domain.ToolCallEventData{
					ToolName:   "get_environment_variable",
					Parameters: params,
					RequestID:  ctx.RunID,
				})
			}
			var variables []EnvironmentVariable
			query := params.Name
			if query == "" {
				query = params.Pattern
			}

			// By default, include values unless NoValues is true
			includeValue := !params.NoValues

			if params.Name != "" {
				// Get specific variable
				value := os.Getenv(params.Name)
				if value == "" {
					return &GetEnvironmentVariableResult{
						Variables: []EnvironmentVariable{},
						Count:     0,
						Query:     params.Name,
					}, nil
				}

				envVar := EnvironmentVariable{
					Name: params.Name,
				}

				if includeValue {
					if !params.Sensitive && isSensitiveVariable(params.Name, ctx) {
						envVar.Value = maskValue(value)
						envVar.Masked = true
					} else {
						envVar.Value = value
					}
				}

				variables = append(variables, envVar)
			} else {
				// Get all or pattern-matched variables
				environ := os.Environ()
				pattern := params.Pattern
				if pattern == "" {
					pattern = "*"
				}

				for _, env := range environ {
					// Split on first = to handle values with = in them
					parts := strings.SplitN(env, "=", 2)
					if len(parts) != 2 {
						continue
					}

					name := parts[0]
					value := parts[1]

					// Check pattern match
					if pattern != "*" {
						matched, err := matchPattern(name, pattern)
						if err != nil || !matched {
							continue
						}
					}

					envVar := EnvironmentVariable{
						Name: name,
					}

					if includeValue {
						if !params.Sensitive && isSensitiveVariable(name, ctx) {
							envVar.Value = maskValue(value)
							envVar.Masked = true
						} else {
							envVar.Value = value
						}
					}

					variables = append(variables, envVar)
				}

				// Sort variables by name
				sort.Slice(variables, func(i, j int) bool {
					return variables[i].Name < variables[j].Name
				})
			}

			result := &GetEnvironmentVariableResult{
				Variables: variables,
				Count:     len(variables),
				Query:     query,
			}

			// Emit result event
			if ctx.Events != nil {
				ctx.Events.Emit(domain.EventToolResult, domain.ToolResultEventData{
					ToolName:  "get_environment_variable",
					Result:    result,
					RequestID: ctx.RunID,
				})
			}

			return result, nil
		},
		getEnvironmentVariableParamSchema,
	)
}

// matchPattern implements simple glob pattern matching
func matchPattern(name, pattern string) (bool, error) {
	// Convert pattern to case-insensitive matching
	name = strings.ToUpper(name)
	pattern = strings.ToUpper(pattern)

	// Simple glob patterns: *, prefix*, *suffix, *middle*
	if pattern == "*" {
		return true, nil
	}

	if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
		// *middle*
		middle := pattern[1 : len(pattern)-1]
		return strings.Contains(name, middle), nil
	} else if strings.HasPrefix(pattern, "*") {
		// *suffix
		suffix := pattern[1:]
		return strings.HasSuffix(name, suffix), nil
	} else if strings.HasSuffix(pattern, "*") {
		// prefix*
		prefix := pattern[:len(pattern)-1]
		return strings.HasPrefix(name, prefix), nil
	} else {
		// exact match
		return name == pattern, nil
	}
}

// isSensitiveVariable checks if a variable name matches sensitive patterns
func isSensitiveVariable(name string, ctx *domain.ToolContext) bool {
	upperName := strings.ToUpper(name)

	// Check default sensitive patterns
	for _, pattern := range sensitivePatterns {
		if matched, _ := matchPattern(upperName, pattern); matched {
			return true
		}
	}

	// Check state for additional sensitive patterns
	if ctx != nil && ctx.State != nil {
		if patterns, ok := ctx.State.Get("sensitive_env_patterns"); ok {
			if patternList, ok := patterns.([]string); ok {
				for _, pattern := range patternList {
					if matched, _ := matchPattern(upperName, pattern); matched {
						return true
					}
				}
			}
		}
	}

	return false
}

// maskValue masks sensitive values, showing only first and last few characters
func maskValue(value string) string {
	if len(value) <= 8 {
		return "***"
	}

	// Show first 3 and last 3 characters
	return fmt.Sprintf("%s...%s", value[:3], value[len(value)-3:])
}

// MustGetEnvironmentVariable retrieves the registered GetEnvironmentVariable tool or panics
// This is a convenience function for users who want to ensure the tool exists
func MustGetEnvironmentVariable() domain.Tool {
	return tools.MustGetTool("get_environment_variable")
}
