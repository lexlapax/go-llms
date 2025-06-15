// ABOUTME: Integration tests for dynamic tool discovery and registration
// ABOUTME: Tests the enhanced tool discovery system with script-based tools and runtime registration

package tools

import (
	"bytes"
	"context"
	"fmt"
	"slices"
	"sync"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/testutils/mocks"
)

// mockScriptHandler implements ScriptHandler for testing
type mockScriptHandler struct {
	engine  ScriptEngine
	results map[string]any
	errors  map[string]error
}

func newMockScriptHandler(engine ScriptEngine) *mockScriptHandler {
	return &mockScriptHandler{
		engine:  engine,
		results: make(map[string]any),
		errors:  make(map[string]error),
	}
}

func (h *mockScriptHandler) Execute(ctx context.Context, script string, toolCtx *domain.ToolContext, params any) (any, error) {
	if err, hasError := h.errors[script]; hasError {
		return nil, err
	}

	if result, hasResult := h.results[script]; hasResult {
		return result, nil
	}

	// Default behavior - echo the parameters
	return map[string]any{
		"script":     script,
		"params":     params,
		"engine":     string(h.engine),
		"context_id": "test-context",
	}, nil
}

func (h *mockScriptHandler) Validate(script string) error {
	if err, hasError := h.errors[script]; hasError {
		return err
	}
	return nil
}

func (h *mockScriptHandler) Engine() ScriptEngine {
	return h.engine
}

func (h *mockScriptHandler) SupportsFeature(feature string) bool {
	return feature == "basic_execution"
}

func (h *mockScriptHandler) SetResult(script string, result any) {
	h.results[script] = result
}

func (h *mockScriptHandler) SetError(script string, err error) {
	h.errors[script] = err
}

func TestDynamicToolRegistration(t *testing.T) {
	// Create a new discovery instance for testing
	discovery := &toolDiscovery{
		namespaces:       make(map[string]*NamespaceRegistry),
		currentNamespace: "test",
	}
	_ = discovery.CreateNamespace("test")

	// Test basic tool registration
	toolInfo := ToolInfo{
		Name:        "test_tool",
		Description: "A test tool",
		Category:    "testing",
		Tags:        []string{"test"},
		Version:     "1.0.0",
	}

	factory := func() (domain.Tool, error) {
		return mocks.NewMockTool("test_tool", "test tool"), nil
	}

	// Register the tool
	err := discovery.RegisterTool(toolInfo, factory)
	if err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}

	// Verify tool was registered
	tools := discovery.ListTools()
	if len(tools) != 1 {
		t.Fatalf("Expected 1 tool, got %d", len(tools))
	}

	if tools[0].Name != "test_tool" {
		t.Errorf("Expected tool name 'test_tool', got '%s'", tools[0].Name)
	}

	// Test tool creation
	tool, err := discovery.CreateTool("test_tool")
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}

	if tool.Name() != "test_tool" {
		t.Errorf("Expected tool name 'test_tool', got '%s'", tool.Name())
	}

	// Test unregistration
	err = discovery.UnregisterTool("test_tool")
	if err != nil {
		t.Fatalf("Failed to unregister tool: %v", err)
	}

	// Verify tool was unregistered
	tools = discovery.ListTools()
	if len(tools) != 0 {
		t.Fatalf("Expected 0 tools after unregistration, got %d", len(tools))
	}
}

func TestScriptToolFactory(t *testing.T) {
	factory := NewScriptToolFactory()

	// Register mock handlers
	jsHandler := newMockScriptHandler(ScriptEngineJavaScript)
	luaHandler := newMockScriptHandler(ScriptEngineLua)

	err := factory.RegisterScriptHandler(jsHandler)
	if err != nil {
		t.Fatalf("Failed to register JavaScript handler: %v", err)
	}

	err = factory.RegisterScriptHandler(luaHandler)
	if err != nil {
		t.Fatalf("Failed to register Lua handler: %v", err)
	}

	// Test script tool creation
	def := ScriptToolDefinition{
		Name:        "js_calculator",
		Description: "A JavaScript calculator",
		Category:    "math",
		Tags:        []string{"calculator", "math"},
		Version:     "1.0.0",
		Engine:      ScriptEngineJavaScript,
		Script:      "return a + b;",
		ParameterSchema: &schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"a": {Type: "number"},
				"b": {Type: "number"},
			},
		},
		Examples: []domain.ToolExample{
			{
				Name:        "add_numbers",
				Description: "Add two numbers",
				Input:       map[string]any{"a": 5, "b": 3},
				Output:      8,
			},
		},
	}

	tool, err := factory.CreateTool(def)
	if err != nil {
		t.Fatalf("Failed to create script tool: %v", err)
	}

	// Test tool properties
	if tool.Name() != "js_calculator" {
		t.Errorf("Expected tool name 'js_calculator', got '%s'", tool.Name())
	}

	if tool.Category() != "math" {
		t.Errorf("Expected category 'math', got '%s'", tool.Category())
	}

	tags := tool.Tags()
	expectedTags := []string{"calculator", "math", "javascript"}
	if len(tags) != len(expectedTags) {
		t.Errorf("Expected %d tags, got %d", len(expectedTags), len(tags))
	}

	// Test script execution (mock)
	ctx := &domain.ToolContext{}
	params := map[string]any{"a": 5, "b": 3}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Failed to execute script tool: %v", err)
	}

	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("Expected map result, got %T", result)
	}

	if resultMap["engine"] != "javascript" {
		t.Errorf("Expected engine 'javascript', got '%v'", resultMap["engine"])
	}
}

func TestToolVersioning(t *testing.T) {
	discovery := &toolDiscovery{
		namespaces:       make(map[string]*NamespaceRegistry),
		currentNamespace: "test",
	}
	_ = discovery.CreateNamespace("test")

	// Register version 1.0.0
	toolInfo1 := ToolInfo{
		Name:        "versioned_tool",
		Description: "Version 1.0.0",
		Version:     "1.0.0",
	}

	factory1 := func() (domain.Tool, error) {
		return &versionedMockTool{
			MockTool: mocks.NewMockTool("versioned_tool", "versioned tool"),
			version:  "1.0.0",
		}, nil
	}

	err := discovery.RegisterToolVersion(toolInfo1, factory1, "1.0.0")
	if err != nil {
		t.Fatalf("Failed to register tool version 1.0.0: %v", err)
	}

	// Register version 2.0.0
	toolInfo2 := ToolInfo{
		Name:        "versioned_tool",
		Description: "Version 2.0.0",
		Version:     "2.0.0",
	}

	factory2 := func() (domain.Tool, error) {
		return &versionedMockTool{
			MockTool: mocks.NewMockTool("versioned_tool", "versioned tool"),
			version:  "2.0.0",
		}, nil
	}

	err = discovery.RegisterToolVersion(toolInfo2, factory2, "2.0.0")
	if err != nil {
		t.Fatalf("Failed to register tool version 2.0.0: %v", err)
	}

	// Test version listing
	versions := discovery.GetToolVersions("versioned_tool")
	if len(versions) != 2 {
		t.Fatalf("Expected 2 versions, got %d", len(versions))
	}

	// Test creating specific versions
	tool1, err := discovery.CreateToolVersion("versioned_tool", "1.0.0")
	if err != nil {
		t.Fatalf("Failed to create tool version 1.0.0: %v", err)
	}

	if tool1.Version() != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", tool1.Version())
	}

	tool2, err := discovery.CreateToolVersion("versioned_tool", "2.0.0")
	if err != nil {
		t.Fatalf("Failed to create tool version 2.0.0: %v", err)
	}

	if tool2.Version() != "2.0.0" {
		t.Errorf("Expected version '2.0.0', got '%s'", tool2.Version())
	}
}

func TestNamespaceIsolation(t *testing.T) {
	discovery := &toolDiscovery{
		namespaces:       make(map[string]*NamespaceRegistry),
		currentNamespace: "default",
	}
	_ = discovery.CreateNamespace("default")

	// Create namespaces
	err := discovery.CreateNamespace("tenant1")
	if err != nil {
		t.Fatalf("Failed to create namespace 'tenant1': %v", err)
	}

	err = discovery.CreateNamespace("tenant2")
	if err != nil {
		t.Fatalf("Failed to create namespace 'tenant2': %v", err)
	}

	// Register tool in tenant1
	err = discovery.SwitchNamespace("tenant1")
	if err != nil {
		t.Fatalf("Failed to switch to namespace 'tenant1': %v", err)
	}

	toolInfo1 := ToolInfo{
		Name:        "tenant_tool",
		Description: "Tool for tenant 1",
	}

	factory1 := func() (domain.Tool, error) {
		return mocks.NewMockTool("tenant_tool", "tenant tool"), nil
	}

	err = discovery.RegisterTool(toolInfo1, factory1)
	if err != nil {
		t.Fatalf("Failed to register tool in tenant1: %v", err)
	}

	// Register different tool in tenant2
	err = discovery.SwitchNamespace("tenant2")
	if err != nil {
		t.Fatalf("Failed to switch to namespace 'tenant2': %v", err)
	}

	toolInfo2 := ToolInfo{
		Name:        "other_tool",
		Description: "Tool for tenant 2",
	}

	factory2 := func() (domain.Tool, error) {
		return mocks.NewMockTool("other_tool", "other tool"), nil
	}

	err = discovery.RegisterTool(toolInfo2, factory2)
	if err != nil {
		t.Fatalf("Failed to register tool in tenant2: %v", err)
	}

	// Verify namespace isolation
	tools := discovery.ListTools()
	if len(tools) != 1 {
		t.Fatalf("Expected 1 tool in tenant2, got %d", len(tools))
	}

	if tools[0].Name != "other_tool" {
		t.Errorf("Expected 'other_tool' in tenant2, got '%s'", tools[0].Name)
	}

	// Switch back to tenant1 and verify
	err = discovery.SwitchNamespace("tenant1")
	if err != nil {
		t.Fatalf("Failed to switch back to namespace 'tenant1': %v", err)
	}

	tools = discovery.ListTools()
	if len(tools) != 1 {
		t.Fatalf("Expected 1 tool in tenant1, got %d", len(tools))
	}

	if tools[0].Name != "tenant_tool" {
		t.Errorf("Expected 'tenant_tool' in tenant1, got '%s'", tools[0].Name)
	}
}

func TestRegistryPersistence(t *testing.T) {
	discovery := &toolDiscovery{
		namespaces:       make(map[string]*NamespaceRegistry),
		currentNamespace: "default",
	}
	_ = discovery.CreateNamespace("default")

	// Register some tools
	toolInfo := ToolInfo{
		Name:        "persistent_tool",
		Description: "A tool for testing persistence",
		Category:    "testing",
	}

	factory := func() (domain.Tool, error) {
		return mocks.NewMockTool("persistent_tool", "persistent tool"), nil
	}

	err := discovery.RegisterTool(toolInfo, factory)
	if err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}

	// Save registry
	var buf bytes.Buffer
	err = discovery.SaveRegistry(&buf)
	if err != nil {
		t.Fatalf("Failed to save registry: %v", err)
	}

	// Create new discovery instance
	newDiscovery := &toolDiscovery{
		namespaces:       make(map[string]*NamespaceRegistry),
		currentNamespace: "default",
	}

	// Load registry
	err = newDiscovery.LoadRegistry(&buf)
	if err != nil {
		t.Fatalf("Failed to load registry: %v", err)
	}

	// Verify tool metadata was restored
	tools := newDiscovery.ListTools()
	if len(tools) != 1 {
		t.Fatalf("Expected 1 tool after loading, got %d", len(tools))
	}

	if tools[0].Name != "persistent_tool" {
		t.Errorf("Expected tool name 'persistent_tool', got '%s'", tools[0].Name)
	}

	// Note: Factories are not persisted, so tool creation will fail
	// This is expected behavior - factories need to be re-registered
	_, err = newDiscovery.CreateTool("persistent_tool")
	if err == nil {
		t.Error("Expected error when creating tool without factory, but got none")
	}
}

func TestConcurrentToolRegistration(t *testing.T) {
	discovery := &toolDiscovery{
		namespaces:       make(map[string]*NamespaceRegistry),
		currentNamespace: "default",
	}
	_ = discovery.CreateNamespace("default")

	const numGoroutines = 10
	const toolsPerGoroutine = 5

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Register tools concurrently
	for i := range numGoroutines {
		go func(routineID int) {
			defer wg.Done()

			for j := range toolsPerGoroutine {
				toolName := fmt.Sprintf("tool_%d_%d", routineID, j)
				toolInfo := ToolInfo{
					Name:        toolName,
					Description: fmt.Sprintf("Tool %s", toolName),
				}

				factory := func() (domain.Tool, error) {
					return mocks.NewMockTool(toolName, toolName+" tool"), nil
				}

				err := discovery.RegisterTool(toolInfo, factory)
				if err != nil {
					t.Errorf("Failed to register tool %s: %v", toolName, err)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify all tools were registered
	tools := discovery.ListTools()
	expectedCount := numGoroutines * toolsPerGoroutine
	if len(tools) != expectedCount {
		t.Errorf("Expected %d tools, got %d", expectedCount, len(tools))
	}
}

func TestScriptToolIntegration(t *testing.T) {
	// Register mock script handler globally
	handler := newMockScriptHandler(ScriptEngineJavaScript)
	handler.SetResult("return 'Hello from script!';", "Hello from script!")

	err := RegisterScriptHandler(handler)
	if err != nil {
		t.Fatalf("Failed to register script handler: %v", err)
	}

	// Create script tool definition
	def := ScriptToolDefinition{
		Name:        "greeting_tool",
		Description: "A tool that greets users",
		Category:    "utility",
		Tags:        []string{"greeting"},
		Version:     "1.0.0",
		Engine:      ScriptEngineJavaScript,
		Script:      "return 'Hello from script!';",
	}

	// Register script tool with global discovery
	err = RegisterScriptToolWithDiscovery(def)
	if err != nil {
		t.Fatalf("Failed to register script tool with discovery: %v", err)
	}

	// Use global discovery instance to check registration
	discovery := NewDiscovery()

	// Verify tool was registered
	tools := discovery.ListTools()
	found := false
	for _, tool := range tools {
		if tool.Name == "greeting_tool" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Script tool was not found in discovery")
	}

	// Create and test the tool
	tool, err := discovery.CreateTool("greeting_tool")
	if err != nil {
		t.Fatalf("Failed to create script tool: %v", err)
	}

	// Verify tool properties
	if tool.Name() != "greeting_tool" {
		t.Errorf("Expected tool name 'greeting_tool', got '%s'", tool.Name())
	}

	if tool.Category() != "utility" {
		t.Errorf("Expected category 'utility', got '%s'", tool.Category())
	}

	// Verify tags include engine
	tags := tool.Tags()
	if !slices.Contains(tags, "javascript") {
		t.Error("Expected tool tags to include engine tag 'javascript'")
	}
}

// versionedMockTool wraps MockTool to override version
type versionedMockTool struct {
	*mocks.MockTool
	version string
}

func (v *versionedMockTool) Version() string {
	if v.version != "" {
		return v.version
	}
	return v.MockTool.Version()
}
