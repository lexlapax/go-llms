// ABOUTME: Example demonstrating the use of built-in web tools
// ABOUTME: Shows web fetching, searching, scraping, and HTTP requests

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
)

func main() {
	ctx := context.Background()

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
	fetchResult, err := fetchTool.Execute(ctx, map[string]interface{}{
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
	searchResult, err := searchTool.Execute(ctx, map[string]interface{}{
		"query":       "golang generics tutorial",
		"max_results": 3,
		"safe_search": "moderate",
	})
	if err != nil {
		log.Printf("Failed to search web: %v", err)
	} else {
		fmt.Printf("Search results: %+v\n\n", searchResult)
	}

	// Example 3: Web Scrape
	fmt.Println("=== Example 3: Web Scrape ===")
	scrapeTool := tools.MustGetTool("web_scrape")
	scrapeResult, err := scrapeTool.Execute(ctx, map[string]interface{}{
		"url": "https://example.com",
		"extract_options": map[string]interface{}{
			"include_text":     true,
			"include_links":    true,
			"include_metadata": true,
			"selectors": []string{
				"h1",       // All h1 tags
				".content", // Elements with class "content"
			},
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
	postResult, err := httpTool.Execute(ctx, map[string]interface{}{
		"url":    "https://httpbin.org/post",
		"method": "POST",
		"headers": map[string]string{
			"Content-Type": "application/json",
			"User-Agent":   "go-llms/1.0",
		},
		"body": map[string]interface{}{
			"name":    "John Doe",
			"email":   "john@example.com",
			"message": "Hello from go-llms!",
		},
		"timeout": 10,
	})
	if err != nil {
		log.Printf("Failed to make POST request: %v", err)
	} else {
		fmt.Printf("POST response: %+v\n\n", postResult)
	}

	// Example 5: HTTP Request with Authentication
	fmt.Println("=== Example 5: HTTP Request with Auth ===")
	authResult, err := httpTool.Execute(ctx, map[string]interface{}{
		"url":               "https://httpbin.org/bearer",
		"method":            "GET",
		"auth_type":         "bearer",
		"auth_bearer_token": "example-token-123",
		"timeout":           10,
	})
	if err != nil {
		log.Printf("Failed to make authenticated request: %v", err)
	} else {
		fmt.Printf("Auth response: %+v\n\n", authResult)
	}

	// Example 6: Advanced HTTP Request with Query Parameters
	fmt.Println("=== Example 6: HTTP Request with Query Params ===")
	queryResult, err := httpTool.Execute(ctx, map[string]interface{}{
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
}
