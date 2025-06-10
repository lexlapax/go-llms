// ABOUTME: Example demonstrating the use of built-in web tools
// ABOUTME: Shows web fetching, searching, scraping, and HTTP requests

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
	agentDomain "github.com/lexlapax/go-llms/pkg/agent/domain"
)

// Helper types for creating a minimal ToolContext for standalone tool execution

// minimalStateReader implements StateReader interface with empty state
type minimalStateReader struct {
	state *agentDomain.State
}

func (m *minimalStateReader) Get(key string) (interface{}, bool) {
	return m.state.Get(key)
}

func (m *minimalStateReader) Values() map[string]interface{} {
	return m.state.Values()
}

func (m *minimalStateReader) GetArtifact(id string) (*agentDomain.Artifact, bool) {
	return m.state.GetArtifact(id)
}

func (m *minimalStateReader) Artifacts() map[string]*agentDomain.Artifact {
	return m.state.Artifacts()
}

func (m *minimalStateReader) Messages() []agentDomain.Message {
	return m.state.Messages()
}

func (m *minimalStateReader) GetMetadata(key string) (interface{}, bool) {
	return m.state.GetMetadata(key)
}

func (m *minimalStateReader) Has(key string) bool {
	return m.state.Has(key)
}

func (m *minimalStateReader) Keys() []string {
	values := m.state.Values()
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	return keys
}

// minimalEventEmitter implements EventEmitter interface with no-op methods
type minimalEventEmitter struct{}

func (m *minimalEventEmitter) Emit(eventType agentDomain.EventType, data interface{}) {}
func (m *minimalEventEmitter) EmitProgress(current, total int, message string)        {}
func (m *minimalEventEmitter) EmitMessage(message string)                             {}
func (m *minimalEventEmitter) EmitError(err error)                                    {}
func (m *minimalEventEmitter) EmitCustom(eventName string, data interface{})          {}

// createToolContext creates a minimal ToolContext for standalone tool execution
func createToolContext(ctx context.Context) *agentDomain.ToolContext {
	state := agentDomain.NewState()
	stateReader := &minimalStateReader{state: state}

	toolCtx := &agentDomain.ToolContext{
		Context:   ctx,
		State:     stateReader,
		RunID:     "standalone-execution",
		Retry:     0,
		StartTime: time.Now(),
		Events:    &minimalEventEmitter{},
		Agent: agentDomain.AgentInfo{
			ID:          "standalone",
			Name:        "standalone-tool-executor",
			Description: "Minimal agent for standalone tool execution",
			Type:        agentDomain.AgentTypeLLM,
			Metadata:    make(map[string]interface{}),
		},
	}

	return toolCtx
}

func main() {
	ctx := context.Background()
	toolCtx := createToolContext(ctx)

	// List all web tools
	fmt.Println("=== Available Web Tools ===")
	webTools := tools.Tools.ListByCategory("web")
	for _, entry := range webTools {
		fmt.Printf("- %s: %s\n", entry.Metadata.Name, entry.Metadata.Description)
	}
	fmt.Println()

	// Example 1: Web Fetch
	fmt.Println("=== Example 1: Web Fetch ===")
	fetchTool := tools.MustGetTool("web_fetch")
	fetchResult, err := fetchTool.Execute(toolCtx, map[string]interface{}{
		"url":     "https://api.github.com/users/github",
		"timeout": 10,
		"headers": map[string]string{
			"Accept": "application/json",
		},
	})
	if err != nil {
		log.Printf("Failed to fetch URL: %v", err)
	} else {
		// Pretty print the result
		if jsonBytes, err := json.MarshalIndent(fetchResult, "", "  "); err == nil {
			fmt.Printf("Fetch result:\n%s\n\n", string(jsonBytes))
		}
	}

	// Example 2: Web Search
	fmt.Println("=== Example 2: Web Search ===")
	searchTool := tools.MustGetTool("web_search")
	searchResult, err := searchTool.Execute(toolCtx, map[string]interface{}{
		"query":       "golang generics tutorial",
		"max_results": 3,
		"safe_search": true,
	})
	if err != nil {
		log.Printf("Failed to search web: %v", err)
	} else {
		fmt.Printf("Search results: %+v\n\n", searchResult)
	}

	// Example 3: Web Scrape
	fmt.Println("=== Example 3: Web Scrape ===")
	scrapeTool := tools.MustGetTool("web_scrape")
	scrapeResult, err := scrapeTool.Execute(toolCtx, map[string]interface{}{
		"url":           "https://example.com",
		"extract_text":  true,
		"extract_links": true,
		"extract_meta":  true,
		"selectors": []string{
			"h1", // All h1 tags
			"p",  // All p tags
		},
	})
	if err != nil {
		log.Printf("Failed to scrape web page: %v", err)
	} else {
		// Pretty print the result
		if jsonBytes, err := json.MarshalIndent(scrapeResult, "", "  "); err == nil {
			fmt.Printf("Scrape result:\n%s\n\n", string(jsonBytes))
		}
	}

	// Example 4: HTTP Request (POST)
	fmt.Println("=== Example 4: HTTP Request (POST) ===")
	httpTool := tools.MustGetTool("http_request")

	// Example POST request to httpbin.org
	postBody := `{"name":"John Doe","email":"john@example.com","message":"Hello from go-llms!"}`
	postResult, err := httpTool.Execute(toolCtx, map[string]interface{}{
		"url":    "https://httpbin.org/post",
		"method": "POST",
		"headers": map[string]string{
			"User-Agent": "go-llms/1.0",
		},
		"body":      postBody,
		"body_type": "json",
		"timeout":   10,
	})
	if err != nil {
		log.Printf("Failed to make POST request: %v", err)
	} else {
		fmt.Printf("POST response: %+v\n\n", postResult)
	}

	// Example 5: HTTP Request with Authentication
	fmt.Println("=== Example 5: HTTP Request with Auth ===")
	authResult, err := httpTool.Execute(toolCtx, map[string]interface{}{
		"url":        "https://httpbin.org/bearer",
		"method":     "GET",
		"auth_type":  "bearer",
		"auth_token": "example-token-123",
		"timeout":    10,
	})
	if err != nil {
		log.Printf("Failed to make authenticated request: %v", err)
	} else {
		fmt.Printf("Auth response: %+v\n\n", authResult)
	}

	// Example 6: Advanced HTTP Request with Query Parameters
	fmt.Println("=== Example 6: HTTP Request with Query Params ===")
	queryResult, err := httpTool.Execute(toolCtx, map[string]interface{}{
		"url":    "https://httpbin.org/get",
		"method": "GET",
		"query_params": map[string]string{
			"page":     "1",
			"per_page": "10",
			"sort":     "created",
		},
		"follow_redirects": true,
		"timeout":          10,
	})
	if err != nil {
		log.Printf("Failed to make request with query params: %v", err)
	} else {
		fmt.Printf("Query params response: %+v\n", queryResult)
	}

	// Example 7: Web Fetch with Authentication
	fmt.Println("\n=== Example 7: Web Fetch with Auth ===")
	authFetchResult, err := fetchTool.Execute(toolCtx, map[string]interface{}{
		"url":        "https://httpbin.org/bearer",
		"timeout":    10,
		"auth_type":  "bearer",
		"auth_token": "example-bearer-token-123",
	})
	if err != nil {
		log.Printf("Failed to fetch with auth: %v", err)
	} else {
		fmt.Printf("Authenticated fetch result: %+v\n", authFetchResult)
	}

	// Example 8: Web Scrape with Authentication
	fmt.Println("\n=== Example 8: Web Scrape with Auth ===")
	authScrapeResult, err := scrapeTool.Execute(toolCtx, map[string]interface{}{
		"url":           "https://httpbin.org/basic-auth/user/pass",
		"auth_type":     "basic",
		"auth_username": "user",
		"auth_password": "pass",
		"extract_text":  true,
		"extract_meta":  true,
	})
	if err != nil {
		log.Printf("Failed to scrape with auth: %v", err)
	} else {
		fmt.Printf("Authenticated scrape result: %+v\n", authScrapeResult)
	}
}
