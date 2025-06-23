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

// getEnvironmentVariableOutputSchema defines the output schema for the GetEnvironmentVariable tool
var getEnvironmentVariableOutputSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"variables": {
			Type:        "array",
			Description: "List of environment variables matching the query",
			Items: &sdomain.Property{
				Type: "object",
				Properties: map[string]sdomain.Property{
					"name": {
						Type:        "string",
						Description: "Environment variable name",
					},
					"value": {
						Type:        "string",
						Description: "Environment variable value (if no_values is false)",
					},
					"masked": {
						Type:        "boolean",
						Description: "Whether the value was masked for security",
					},
				},
				Required: []string{"name"},
			},
		},
		"count": {
			Type:        "number",
			Description: "Number of variables found",
		},
		"query": {
			Type:        "string",
			Description: "The name or pattern that was searched",
		},
	},
	Required: []string{"variables", "count"},
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

// getEnvironmentVariable is the main function for the tool
func getEnvironmentVariable(ctx *domain.ToolContext, params GetEnvironmentVariableParams) (*GetEnvironmentVariableResult, error) {
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
}

// GetEnvironmentVariable creates a tool for safely reading environment variables with built-in security
// features including automatic masking of sensitive values (API keys, secrets, tokens) and pattern-based
// filtering. The tool supports retrieving specific variables by name, searching with wildcards (*PATH, GO*),
// and provides options to exclude values or unmask sensitive data when explicitly needed.
func GetEnvironmentVariable() domain.Tool {
	builder := atools.NewToolBuilder("get_environment_variable", "Retrieves environment variables safely").
		WithFunction(getEnvironmentVariable).
		WithParameterSchema(getEnvironmentVariableParamSchema).
		WithOutputSchema(getEnvironmentVariableOutputSchema).
		WithUsageInstructions(`Use this tool to safely read environment variables from the system.

Security Features:
- Sensitive variables (containing KEY, SECRET, TOKEN, PASSWORD, etc.) are masked by default
- Use 'sensitive: true' to see unmasked values when necessary
- Add custom sensitive patterns via state: state.Set("sensitive_env_patterns", []string{"*PRIVATE*"})

Parameters:
- name: Retrieve a specific environment variable by exact name (optional)
- pattern: Search for variables matching a pattern (optional)
- no_values: Return only variable names without values (optional, default false)
- sensitive: Allow unmasked retrieval of sensitive variables (optional, default false)

Pattern Matching:
- Use * as wildcard: "GO*" matches all variables starting with GO
- "*_PATH" matches all variables ending with _PATH
- "*API*" matches all variables containing API
- "*" or empty pattern returns all variables

Output:
- variables: Array of found environment variables
- count: Number of variables found
- query: The search term used (name or pattern)

Security Masking:
Sensitive values show only first 3 and last 3 characters:
- Full value: "sk-abc123def456ghi789"
- Masked: "sk-...789"`).
		WithExamples([]domain.ToolExample{
			{
				Name:        "Get specific variable",
				Description: "Retrieve the HOME environment variable",
				Scenario:    "When you need the value of a specific environment variable",
				Input: map[string]interface{}{
					"name": "HOME",
				},
				Output: map[string]interface{}{
					"variables": []map[string]interface{}{
						{
							"name":  "HOME",
							"value": "/home/user",
						},
					},
					"count": 1,
					"query": "HOME",
				},
				Explanation: "Returns the exact variable if it exists",
			},
			{
				Name:        "Find Go-related variables",
				Description: "List all environment variables starting with GO",
				Scenario:    "When you need to check Go language configuration",
				Input: map[string]interface{}{
					"pattern": "GO*",
				},
				Output: map[string]interface{}{
					"variables": []map[string]interface{}{
						{"name": "GOPATH", "value": "/home/user/go"},
						{"name": "GOROOT", "value": "/usr/local/go"},
						{"name": "GO111MODULE", "value": "on"},
					},
					"count": 3,
					"query": "GO*",
				},
				Explanation: "Pattern matching returns all variables starting with GO",
			},
			{
				Name:        "List PATH variables",
				Description: "Find all variables ending with PATH",
				Scenario:    "When you need to check system paths",
				Input: map[string]interface{}{
					"pattern": "*PATH",
				},
				Output: map[string]interface{}{
					"variables": []map[string]interface{}{
						{"name": "PATH", "value": "/usr/bin:/bin:/usr/local/bin"},
						{"name": "GOPATH", "value": "/home/user/go"},
						{"name": "PYTHONPATH", "value": "/usr/lib/python3"},
					},
					"count": 3,
					"query": "*PATH",
				},
				Explanation: "Suffix pattern matching finds all *PATH variables",
			},
			{
				Name:        "List variable names only",
				Description: "Get all variable names without values",
				Scenario:    "When you need to discover available variables without exposing values",
				Input: map[string]interface{}{
					"pattern":   "*",
					"no_values": true,
				},
				Output: map[string]interface{}{
					"variables": []map[string]interface{}{
						{"name": "HOME"},
						{"name": "PATH"},
						{"name": "USER"},
					},
					"count": 3,
					"query": "*",
				},
				Explanation: "no_values: true returns only variable names",
			},
			{
				Name:        "Handle sensitive variables",
				Description: "Retrieve API key with masked value",
				Scenario:    "When checking if sensitive variables are set",
				Input: map[string]interface{}{
					"name": "OPENAI_API_KEY",
				},
				Output: map[string]interface{}{
					"variables": []map[string]interface{}{
						{
							"name":   "OPENAI_API_KEY",
							"value":  "sk-...789",
							"masked": true,
						},
					},
					"count": 1,
					"query": "OPENAI_API_KEY",
				},
				Explanation: "Sensitive variables are automatically masked for security",
			},
			{
				Name:        "Retrieve sensitive unmasked",
				Description: "Get API key value unmasked when needed",
				Scenario:    "When you explicitly need the full sensitive value",
				Input: map[string]interface{}{
					"name":      "OPENAI_API_KEY",
					"sensitive": true,
				},
				Output: map[string]interface{}{
					"variables": []map[string]interface{}{
						{
							"name":  "OPENAI_API_KEY",
							"value": "sk-abc123def456ghi789",
						},
					},
					"count": 1,
					"query": "OPENAI_API_KEY",
				},
				Explanation: "sensitive: true allows full value retrieval",
			},
			{
				Name:        "Non-existent variable",
				Description: "Request a variable that doesn't exist",
				Scenario:    "When checking if a variable is set",
				Input: map[string]interface{}{
					"name": "NONEXISTENT_VAR",
				},
				Output: map[string]interface{}{
					"variables": []interface{}{},
					"count":     0,
					"query":     "NONEXISTENT_VAR",
				},
				Explanation: "Returns empty array when variable not found",
			},
		}).
		WithConstraints([]string{
			"Cannot modify environment variables, only read them",
			"Pattern matching is case-insensitive",
			"Sensitive variables are masked by default for security",
			"Only simple glob patterns supported: *, prefix*, *suffix, *middle*",
			"Variables are sorted alphabetically in results",
			"Empty values are included in results",
			"Masking shows first 3 and last 3 characters for values longer than 8 chars",
			"Custom sensitive patterns can be added via state configuration",
		}).
		WithErrorGuidance(map[string]string{
			"invalid pattern":    "Use simple patterns: *, prefix*, *suffix, or *contains*",
			"no variables found": "Check the variable name or pattern spelling",
			"access denied":      "Some systems may restrict environment variable access",
		}).
		WithCategory("system").
		WithTags([]string{"environment", "config", "system", "variables"}).
		WithVersion("2.0.0").
		WithBehavior(
			true,   // Deterministic - same query returns same variables
			false,  // Not destructive - only reads
			false,  // No confirmation needed
			"fast", // Very fast operation
		)

	return builder.Build()
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
