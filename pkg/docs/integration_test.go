package docs

import (
	"context"
	"strings"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/tools"
)

// mockDiscovery provides a mock implementation of ToolDiscovery for testing
type mockDiscovery struct {
	tools []tools.ToolInfo
}

func (m *mockDiscovery) ListTools() []tools.ToolInfo {
	return m.tools
}

func (m *mockDiscovery) SearchTools(query string) []tools.ToolInfo {
	var results []tools.ToolInfo
	query = strings.ToLower(query)
	for _, tool := range m.tools {
		if strings.Contains(strings.ToLower(tool.Name), query) ||
			strings.Contains(strings.ToLower(tool.Description), query) {
			results = append(results, tool)
		}
	}
	return results
}

func (m *mockDiscovery) ListByCategory(category string) []tools.ToolInfo {
	var results []tools.ToolInfo
	for _, tool := range m.tools {
		if tool.Category == category {
			results = append(results, tool)
		}
	}
	return results
}

func (m *mockDiscovery) GetToolSchema(name string) (*tools.ToolSchema, error) {
	return nil, nil
}

func (m *mockDiscovery) GetToolExamples(name string) ([]domain.ToolExample, error) {
	return nil, nil
}

func (m *mockDiscovery) CreateTool(name string) (domain.Tool, error) {
	return nil, nil
}

func (m *mockDiscovery) CreateTools(names ...string) (map[string]domain.Tool, error) {
	return nil, nil
}

func (m *mockDiscovery) GetToolHelp(name string) (string, error) {
	return "Basic help for " + name, nil
}

func (m *mockDiscovery) RegisterTool(info tools.ToolInfo, factory tools.ToolFactory) error {
	return nil
}

func (m *mockDiscovery) UnregisterTool(name string) error {
	return nil
}

func (m *mockDiscovery) GetRegisteredTools() []tools.ToolInfo {
	return m.tools
}

func (m *mockDiscovery) SaveRegistry(writer interface{}) error {
	return nil
}

func (m *mockDiscovery) LoadRegistry(reader interface{}) error {
	return nil
}

func (m *mockDiscovery) RegisterToolVersion(info tools.ToolInfo, factory tools.ToolFactory, version string) error {
	return nil
}

func (m *mockDiscovery) GetToolVersions(name string) []string {
	return []string{"1.0.0"}
}

func (m *mockDiscovery) CreateToolVersion(name, version string) (domain.Tool, error) {
	return nil, nil
}

func (m *mockDiscovery) CreateNamespace(namespace string) error {
	return nil
}

func (m *mockDiscovery) ListNamespaces() []string {
	return []string{"default"}
}

func (m *mockDiscovery) SwitchNamespace(namespace string) error {
	return nil
}

func (m *mockDiscovery) GetCurrentNamespace() string {
	return "default"
}

func createMockDiscovery() *mockDiscovery {
	return &mockDiscovery{
		tools: []tools.ToolInfo{
			{
				Name:        "file_read",
				Description: "Read files from disk",
				Category:    "file",
				Tags:        []string{"file", "io"},
				Version:     "1.0.0",
			},
			{
				Name:        "file_write",
				Description: "Write files to disk",
				Category:    "file",
				Tags:        []string{"file", "io"},
				Version:     "1.0.0",
			},
			{
				Name:        "web_fetch",
				Description: "Fetch web pages",
				Category:    "web",
				Tags:        []string{"web", "http"},
				Version:     "1.0.0",
			},
		},
	}
}

func TestToolDocumentationIntegrator_GenerateDocsForAllTools(t *testing.T) {
	ctx := context.Background()
	discovery := createMockDiscovery()
	config := GeneratorConfig{
		Title:   "Test Tools",
		Version: "1.0.0",
	}

	integrator := NewToolDocumentationIntegrator(discovery, config)

	docs, err := integrator.GenerateDocsForAllTools(ctx)
	if err != nil {
		t.Fatalf("GenerateDocsForAllTools failed: %v", err)
	}

	if len(docs) != 3 {
		t.Errorf("Expected 3 docs, got %d", len(docs))
	}

	// Verify first doc
	if docs[0].Name != "file_read" {
		t.Errorf("Expected first doc name 'file_read', got '%s'", docs[0].Name)
	}
}

func TestToolDocumentationIntegrator_GenerateOpenAPIForAllTools(t *testing.T) {
	ctx := context.Background()
	discovery := createMockDiscovery()
	config := GeneratorConfig{
		Title:   "Test API",
		Version: "1.0.0",
		BaseURL: "https://api.test.com",
	}

	integrator := NewToolDocumentationIntegrator(discovery, config)

	spec, err := integrator.GenerateOpenAPIForAllTools(ctx)
	if err != nil {
		t.Fatalf("GenerateOpenAPIForAllTools failed: %v", err)
	}

	if spec.Info.Title != "Test API" {
		t.Errorf("Expected title 'Test API', got '%s'", spec.Info.Title)
	}

	if len(spec.Paths) != 3 {
		t.Errorf("Expected 3 paths, got %d", len(spec.Paths))
	}
}

func TestToolDocumentationIntegrator_GenerateDocsForCategory(t *testing.T) {
	ctx := context.Background()
	discovery := createMockDiscovery()
	config := GeneratorConfig{Title: "Test"}

	integrator := NewToolDocumentationIntegrator(discovery, config)

	docs, err := integrator.GenerateDocsForCategory(ctx, "file")
	if err != nil {
		t.Fatalf("GenerateDocsForCategory failed: %v", err)
	}

	if len(docs) != 2 {
		t.Errorf("Expected 2 docs for 'file' category, got %d", len(docs))
	}

	for _, doc := range docs {
		if doc.Category != "file" {
			t.Errorf("Expected category 'file', got '%s'", doc.Category)
		}
	}
}

func TestToolDocumentationIntegrator_GenerateDocsForSearchQuery(t *testing.T) {
	ctx := context.Background()
	discovery := createMockDiscovery()
	config := GeneratorConfig{Title: "Test"}

	integrator := NewToolDocumentationIntegrator(discovery, config)

	docs, err := integrator.GenerateDocsForSearchQuery(ctx, "file")
	if err != nil {
		t.Fatalf("GenerateDocsForSearchQuery failed: %v", err)
	}

	// Should find both file tools
	if len(docs) != 2 {
		t.Errorf("Expected 2 docs for 'file' search, got %d", len(docs))
	}

	// Test search by description
	webDocs, err := integrator.GenerateDocsForSearchQuery(ctx, "web")
	if err != nil {
		t.Fatalf("GenerateDocsForSearchQuery failed: %v", err)
	}

	if len(webDocs) != 1 {
		t.Errorf("Expected 1 doc for 'web' search, got %d", len(webDocs))
	}
}

func TestToolDocumentationIntegrator_IntegrateWithToolHelp(t *testing.T) {
	ctx := context.Background()
	discovery := createMockDiscovery()
	config := GeneratorConfig{Title: "Test"}

	integrator := NewToolDocumentationIntegrator(discovery, config)

	help, err := integrator.IntegrateWithToolHelp(ctx, "file_read")
	if err != nil {
		t.Fatalf("IntegrateWithToolHelp failed: %v", err)
	}

	// Should contain enhanced help
	if !strings.Contains(help, "# file_read") {
		t.Error("Expected enhanced help to contain tool name as header")
	}

	if !strings.Contains(help, "**Description:**") {
		t.Error("Expected enhanced help to contain description section")
	}
}

func TestToolDocumentationIntegrator_GetToolCategories(t *testing.T) {
	discovery := createMockDiscovery()
	config := GeneratorConfig{Title: "Test"}

	integrator := NewToolDocumentationIntegrator(discovery, config)

	categories := integrator.GetToolCategories()

	expectedCategories := map[string]bool{
		"file": true,
		"web":  true,
	}

	if len(categories) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(categories))
	}

	for _, category := range categories {
		if !expectedCategories[category] {
			t.Errorf("Unexpected category: %s", category)
		}
	}
}

func TestToolDocumentationIntegrator_GetToolTags(t *testing.T) {
	discovery := createMockDiscovery()
	config := GeneratorConfig{Title: "Test"}

	integrator := NewToolDocumentationIntegrator(discovery, config)

	tags := integrator.GetToolTags()

	expectedTags := map[string]bool{
		"file": true,
		"io":   true,
		"web":  true,
		"http": true,
	}

	if len(tags) != 4 {
		t.Errorf("Expected 4 tags, got %d", len(tags))
	}

	for _, tag := range tags {
		if !expectedTags[tag] {
			t.Errorf("Unexpected tag: %s", tag)
		}
	}
}

func TestToolDocumentationIntegrator_BatchGenerate(t *testing.T) {
	ctx := context.Background()
	discovery := createMockDiscovery()
	config := GeneratorConfig{Title: "Test"}

	integrator := NewToolDocumentationIntegrator(discovery, config)

	// Test JSON output
	options := BatchGenerationOptions{
		Categories:      []string{"file"},
		IncludeExamples: true,
		IncludeSchemas:  true,
		OutputFormat:    "json",
	}

	result, err := integrator.BatchGenerate(ctx, options)
	if err != nil {
		t.Fatalf("BatchGenerate failed: %v", err)
	}

	docs, ok := result.([]Documentation)
	if !ok {
		t.Fatal("Expected result to be []Documentation")
	}

	if len(docs) != 2 {
		t.Errorf("Expected 2 docs for file category, got %d", len(docs))
	}

	// Test OpenAPI output
	options.OutputFormat = "openapi"
	result, err = integrator.BatchGenerate(ctx, options)
	if err != nil {
		t.Fatalf("BatchGenerate OpenAPI failed: %v", err)
	}

	spec, ok := result.(*OpenAPISpec)
	if !ok {
		t.Fatal("Expected result to be *OpenAPISpec")
	}

	if len(spec.Paths) != 2 {
		t.Errorf("Expected 2 paths in OpenAPI spec, got %d", len(spec.Paths))
	}
}

func TestConvenienceFunctions(t *testing.T) {
	ctx := context.Background()
	config := GeneratorConfig{
		Title:   "Convenience Test",
		Version: "1.0.0",
	}

	// Test GenerateToolsOpenAPI convenience function
	_, err := GenerateToolsOpenAPI(ctx, config)
	if err != nil {
		t.Fatalf("GenerateToolsOpenAPI convenience function failed: %v", err)
	}

	// Test GenerateToolsMarkdown convenience function
	_, err = GenerateToolsMarkdown(ctx, config)
	if err != nil {
		t.Fatalf("GenerateToolsMarkdown convenience function failed: %v", err)
	}

	// Test GenerateToolsJSON convenience function
	_, err = GenerateToolsJSON(ctx, config)
	if err != nil {
		t.Fatalf("GenerateToolsJSON convenience function failed: %v", err)
	}
}
