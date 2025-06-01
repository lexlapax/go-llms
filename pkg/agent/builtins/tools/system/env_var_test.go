// ABOUTME: Tests for the GetEnvironmentVariable built-in tool
// ABOUTME: Validates environment variable reading, pattern matching, and security features

package system

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
)

func TestGetEnvironmentVariableRegistration(t *testing.T) {
	// Test that the tool is registered
	tool, ok := tools.GetTool("get_environment_variable")
	if !ok {
		t.Fatal("GetEnvironmentVariable tool not registered")
	}
	if tool == nil {
		t.Fatal("GetEnvironmentVariable tool is nil")
	}

	// Test tool name
	if tool.Name() != "get_environment_variable" {
		t.Errorf("Expected tool name 'get_environment_variable', got '%s'", tool.Name())
	}

	// Test metadata
	entries := tools.Tools.Search("get_environment_variable")
	if len(entries) == 0 {
		t.Fatal("GetEnvironmentVariable tool not found in registry")
	}
	
	meta := entries[0].Metadata
	if meta.Category != "system" {
		t.Errorf("Expected category 'system', got '%s'", meta.Category)
	}
}

func TestGetEnvironmentVariableBasic(t *testing.T) {
	tool := GetEnvironmentVariable()
	ctx := context.Background()

	// Set up test environment variables
	testVars := map[string]string{
		"TEST_VAR_1": "value1",
		"TEST_VAR_2": "value2",
		"OTHER_VAR":  "other_value",
	}

	for name, value := range testVars {
		os.Setenv(name, value)
		defer os.Unsetenv(name)
	}

	// Test 1: Get specific variable
	result, err := tool.Execute(ctx, map[string]interface{}{
		"name": "TEST_VAR_1",
	})
	if err != nil {
		t.Fatalf("Failed to get environment variable: %v", err)
	}

	envResult := result.(*GetEnvironmentVariableResult)
	if envResult.Count != 1 {
		t.Errorf("Expected 1 variable, got %d", envResult.Count)
	}
	if len(envResult.Variables) > 0 {
		if envResult.Variables[0].Name != "TEST_VAR_1" {
			t.Errorf("Expected variable name 'TEST_VAR_1', got '%s'", envResult.Variables[0].Name)
		}
		if envResult.Variables[0].Value != "value1" {
			t.Errorf("Expected value 'value1', got '%s'", envResult.Variables[0].Value)
		}
	}

	// Test 2: Non-existent variable
	result, err = tool.Execute(ctx, map[string]interface{}{
		"name": "NON_EXISTENT_VAR",
	})
	if err != nil {
		t.Fatalf("Failed to get non-existent variable: %v", err)
	}

	envResult = result.(*GetEnvironmentVariableResult)
	if envResult.Count != 0 {
		t.Errorf("Expected 0 variables for non-existent var, got %d", envResult.Count)
	}

	// Test 3: Get without value
	result, err = tool.Execute(ctx, map[string]interface{}{
		"name":      "TEST_VAR_1",
		"no_values": true,
	})
	if err != nil {
		t.Fatalf("Failed to get variable without value: %v", err)
	}

	envResult = result.(*GetEnvironmentVariableResult)
	if len(envResult.Variables) > 0 {
		if envResult.Variables[0].Value != "" {
			t.Error("Expected empty value when no_values is true")
		}
	}
}

func TestGetEnvironmentVariablePattern(t *testing.T) {
	tool := GetEnvironmentVariable()
	ctx := context.Background()

	// Set up test environment variables
	testVars := map[string]string{
		"TEST_VAR_1":    "value1",
		"TEST_VAR_2":    "value2",
		"ANOTHER_TEST":  "value3",
		"OTHER_VAR":     "other",
		"VAR_TEST_END":  "end",
	}

	for name, value := range testVars {
		os.Setenv(name, value)
		defer os.Unsetenv(name)
	}

	// Test 1: Prefix pattern
	result, err := tool.Execute(ctx, map[string]interface{}{
		"pattern": "TEST_*",
	})
	if err != nil {
		t.Fatalf("Failed with prefix pattern: %v", err)
	}

	envResult := result.(*GetEnvironmentVariableResult)
	if envResult.Count != 2 { // TEST_VAR_1 and TEST_VAR_2
		t.Errorf("Expected 2 variables with TEST_* pattern, got %d", envResult.Count)
	}

	// Test 2: Suffix pattern
	result, err = tool.Execute(ctx, map[string]interface{}{
		"pattern": "*_TEST",
	})
	if err != nil {
		t.Fatalf("Failed with suffix pattern: %v", err)
	}

	envResult = result.(*GetEnvironmentVariableResult)
	if envResult.Count != 1 { // ANOTHER_TEST
		t.Errorf("Expected 1 variable with *_TEST pattern, got %d", envResult.Count)
	}

	// Test 3: Contains pattern
	result, err = tool.Execute(ctx, map[string]interface{}{
		"pattern": "*TEST*",
	})
	if err != nil {
		t.Fatalf("Failed with contains pattern: %v", err)
	}

	envResult = result.(*GetEnvironmentVariableResult)
	if envResult.Count != 4 { // All vars with TEST in name (excluding OTHER_VAR)
		t.Errorf("Expected 4 variables with *TEST* pattern, got %d", envResult.Count)
		for _, v := range envResult.Variables {
			t.Logf("  Found: %s", v.Name)
		}
	}
}

func TestGetEnvironmentVariableSensitive(t *testing.T) {
	tool := GetEnvironmentVariable()
	ctx := context.Background()

	// Set up sensitive environment variables
	sensitiveVars := map[string]string{
		"API_KEY":         "secret123456789",
		"DB_PASSWORD":     "mypassword123",
		"AUTH_TOKEN":      "token987654321",
		"NORMAL_VAR":      "not_sensitive",
		"SECRET_DATA":     "verysecret",
	}

	for name, value := range sensitiveVars {
		os.Setenv(name, value)
		defer os.Unsetenv(name)
	}

	// Test 1: Get sensitive variable without sensitive flag (should be masked)
	result, err := tool.Execute(ctx, map[string]interface{}{
		"name": "API_KEY",
	})
	if err != nil {
		t.Fatalf("Failed to get sensitive variable: %v", err)
	}

	envResult := result.(*GetEnvironmentVariableResult)
	if len(envResult.Variables) > 0 {
		v := envResult.Variables[0]
		if !v.Masked {
			t.Error("Expected sensitive variable to be masked")
		}
		if v.Value == "secret123456789" {
			t.Error("Sensitive value should not be exposed")
		}
		if !strings.Contains(v.Value, "...") {
			t.Errorf("Expected masked value to contain '...', got '%s'", v.Value)
		}
	}

	// Test 2: Get sensitive variable with sensitive flag (should not be masked)
	result, err = tool.Execute(ctx, map[string]interface{}{
		"name":      "API_KEY",
		"sensitive": true,
	})
	if err != nil {
		t.Fatalf("Failed to get sensitive variable with flag: %v", err)
	}

	envResult = result.(*GetEnvironmentVariableResult)
	if len(envResult.Variables) > 0 {
		v := envResult.Variables[0]
		if v.Masked {
			t.Error("Variable should not be masked with sensitive flag")
		}
		if v.Value != "secret123456789" {
			t.Errorf("Expected full value 'secret123456789', got '%s'", v.Value)
		}
	}

	// Test 3: Pattern with sensitive variables
	result, err = tool.Execute(ctx, map[string]interface{}{
		"pattern": "*",
	})
	if err != nil {
		t.Fatalf("Failed to get all variables: %v", err)
	}

	envResult = result.(*GetEnvironmentVariableResult)
	
	// Check that sensitive variables are masked
	for _, v := range envResult.Variables {
		if isSensitiveVariable(v.Name) && !v.Masked {
			t.Errorf("Variable %s should be masked", v.Name)
		}
		if v.Name == "NORMAL_VAR" && v.Masked {
			t.Error("Normal variable should not be masked")
		}
	}
}

func TestGetEnvironmentVariableAllVariables(t *testing.T) {
	tool := GetEnvironmentVariable()
	ctx := context.Background()

	// Get all environment variables (without values for security)
	result, err := tool.Execute(ctx, map[string]interface{}{
		"pattern":   "*",
		"no_values": true,
	})
	if err != nil {
		t.Fatalf("Failed to get all variables: %v", err)
	}

	envResult := result.(*GetEnvironmentVariableResult)
	
	// Should have at least some system variables
	if envResult.Count == 0 {
		t.Error("Expected at least some environment variables")
	}

	// All values should be empty
	for _, v := range envResult.Variables {
		if v.Value != "" {
			t.Errorf("Expected empty value for %s when no_values is true", v.Name)
		}
	}

	// Variables should be sorted
	for i := 1; i < len(envResult.Variables); i++ {
		if envResult.Variables[i-1].Name > envResult.Variables[i].Name {
			t.Error("Variables are not sorted alphabetically")
			break
		}
	}
}

func TestPatternMatching(t *testing.T) {
	testCases := []struct {
		name     string
		pattern  string
		expected bool
	}{
		{"TEST_VAR", "TEST_*", true},
		{"TEST_VAR", "*_VAR", true},
		{"TEST_VAR", "*EST*", true},
		{"TEST_VAR", "TEST_VAR", true},
		{"TEST_VAR", "*", true},
		{"TEST_VAR", "OTHER_*", false},
		{"TEST_VAR", "*_OTHER", false},
		{"TEST_VAR", "*OTHER*", false},
	}

	for _, tc := range testCases {
		matched, err := matchPattern(tc.name, tc.pattern)
		if err != nil {
			t.Errorf("Pattern matching error for %s with pattern %s: %v", tc.name, tc.pattern, err)
		}
		if matched != tc.expected {
			t.Errorf("Pattern %s with name %s: expected %v, got %v", tc.pattern, tc.name, tc.expected, matched)
		}
	}
}

func TestSensitiveVariableDetection(t *testing.T) {
	testCases := []struct {
		name      string
		sensitive bool
	}{
		{"API_KEY", true},
		{"SECRET_VALUE", true},
		{"DB_PASSWORD", true},
		{"AUTH_TOKEN", true},
		{"PRIVATE_KEY", true},
		{"CREDENTIAL_FILE", true},
		{"NORMAL_VAR", false},
		{"HOME", false},
		{"PATH", false},
		{"USER", false},
	}

	for _, tc := range testCases {
		result := isSensitiveVariable(tc.name)
		if result != tc.sensitive {
			t.Errorf("Variable %s: expected sensitive=%v, got %v", tc.name, tc.sensitive, result)
		}
	}
}

func TestMaskValue(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"short", "***"},
		{"12345678", "***"},
		{"123456789", "123...789"},
		{"verylongsecretvalue", "ver...lue"},
		{"", "***"},
	}

	for _, tc := range testCases {
		result := maskValue(tc.input)
		if result != tc.expected {
			t.Errorf("Mask value %s: expected %s, got %s", tc.input, tc.expected, result)
		}
	}
}