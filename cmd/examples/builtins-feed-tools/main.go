// ABOUTME: Example demonstrating the use of built-in feed processing tools
// ABOUTME: Shows fetching, discovering, filtering, aggregating, converting, and extracting feed data

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/feed"
)

func main() {
	ctx := context.Background()

	// List all feed tools
	fmt.Println("=== Available Feed Tools ===")
	feedTools := tools.Tools.ListByCategory("feed")
	for _, entry := range feedTools {
		fmt.Printf("- %s: %s\n", entry.Metadata.Name, entry.Metadata.Description)
	}
	fmt.Println()

	// Example 1: Fetch a feed
	fmt.Println("=== Example 1: Fetch Feed ===")
	fetchTool := tools.MustGetTool("feed_fetch")
	
	// Note: Using a real feed URL for demonstration
	// In practice, you might want to use a test feed or mock server
	fetchResult, err := fetchTool.Execute(ctx, map[string]interface{}{
		"url":        "https://hnrss.org/frontpage",
		"max_items":  5,
		"timeout":    10,
		"user_agent": "go-llms-feed-example/1.0",
	})
	if err != nil {
		// Try a backup feed if the first one fails
		fmt.Printf("Primary feed failed: %v, trying backup...\n", err)
		fetchResult, err = fetchTool.Execute(ctx, map[string]interface{}{
			"url":       "https://www.reddit.com/r/golang/.rss",
			"max_items": 5,
			"timeout":   10,
		})
		if err != nil {
			log.Printf("Warning: Feed fetch failed: %v", err)
			// Continue with mock data for other examples
			fetchResult = createMockFeedResult()
		}
	}
	
	if result, ok := fetchResult.(map[string]interface{}); ok {
		fmt.Printf("Fetched feed format: %s\n", result["format"])
		fmt.Printf("Status: %d\n", result["status"])
		if feed, ok := result["feed"].(map[string]interface{}); ok {
			fmt.Printf("Feed title: %s\n", feed["title"])
			if items, ok := feed["items"].([]interface{}); ok {
				fmt.Printf("Number of items: %d\n", len(items))
			}
		}
	}
	fmt.Println()

	// Example 2: Discover feeds from a website
	fmt.Println("=== Example 2: Discover Feeds ===")
	discoverTool := tools.MustGetTool("feed_discover")
	discoverResult, err := discoverTool.Execute(ctx, map[string]interface{}{
		"url":            "https://blog.golang.org",
		"follow_links":   true,
		"max_depth":      2,
		"include_podcasts": true,
	})
	if err != nil {
		log.Printf("Warning: Feed discovery failed: %v", err)
		// Continue with mock data
		discoverResult = createMockDiscoverResult()
	}
	
	if result, ok := discoverResult.(map[string]interface{}); ok {
		if feeds, ok := result["feeds"].([]interface{}); ok {
			fmt.Printf("Discovered %d feeds:\n", len(feeds))
			for i, feed := range feeds {
				if f, ok := feed.(map[string]interface{}); ok {
					fmt.Printf("  %d. %s (%s)\n", i+1, f["title"], f["type"])
				}
			}
		}
	}
	fmt.Println()

	// Example 3: Filter feed items
	fmt.Println("=== Example 3: Filter Feed Items ===")
	filterTool := tools.MustGetTool("feed_filter")
	
	// Use the fetched feed or mock data
	_, err = filterTool.Execute(ctx, map[string]interface{}{
		"feed": fetchResult,
		"filters": map[string]interface{}{
			"keywords":        []string{"go", "golang", "programming"},
			"exclude_keywords": []string{"spam", "advertisement"},
			"min_date":        time.Now().AddDate(0, 0, -7).Format(time.RFC3339), // Last week
			"categories":      []string{"technology", "programming"},
		},
		"sort_by": "date",
		"limit":   10,
	})
	if err != nil {
		log.Printf("Warning: Feed filtering failed: %v", err)
	}
	
	fmt.Printf("Filtered feed items based on criteria\n\n")

	// Example 4: Aggregate multiple feeds
	fmt.Println("=== Example 4: Aggregate Feeds ===")
	aggregateTool := tools.MustGetTool("feed_aggregate")
	
	// Create mock feeds for aggregation
	feed1 := createMockFeedResult()
	feed2 := createMockFeedResult()
	
	_, err = aggregateTool.Execute(ctx, map[string]interface{}{
		"feeds": []interface{}{feed1, feed2},
		"merge_options": map[string]interface{}{
			"deduplicate":     true,
			"sort_by":         "date",
			"max_items":       20,
			"merge_metadata":  true,
		},
	})
	if err != nil {
		log.Printf("Warning: Feed aggregation failed: %v", err)
	}
	
	fmt.Printf("Aggregated multiple feeds into one\n\n")

	// Example 5: Convert feed format
	fmt.Println("=== Example 5: Convert Feed Format ===")
	convertTool := tools.MustGetTool("feed_convert")
	
	convertResult, err := convertTool.Execute(ctx, map[string]interface{}{
		"feed":            createMockFeedResult()["feed"],
		"target_type":     "json",
		"pretty":          true,
		"include_content": true,
	})
	if err != nil {
		log.Printf("Warning: Feed conversion failed: %v", err)
	} else {
		if result, ok := convertResult.(map[string]interface{}); ok {
			fmt.Printf("Converted feed to format: %s\n", result["format"])
			// In real usage, you might write this to a file
		}
	}
	fmt.Println()

	// Example 6: Extract specific data from feeds
	fmt.Println("=== Example 6: Extract Feed Data ===")
	extractTool := tools.MustGetTool("feed_extract")
	
	extractResult, err := extractTool.Execute(ctx, map[string]interface{}{
		"feed":             createMockFeedResult()["feed"],
		"fields":           []string{"title", "link", "author.name", "published"},
		"flatten":          true,
		"include_metadata": true,
		"max_items":        5,
	})
	if err != nil {
		log.Printf("Warning: Feed extraction failed: %v", err)
	} else {
		if result, ok := extractResult.(map[string]interface{}); ok {
			fmt.Printf("Extracted data format: %s\n", result["format"])
			// Show first few lines of extracted data
			if data, ok := result["data"].(string); ok && len(data) > 0 {
				fmt.Println("Sample extracted data:")
				lines := []byte(data)
				if len(lines) > 200 {
					fmt.Printf("%s...\n", string(lines[:200]))
				} else {
					fmt.Println(string(lines))
				}
			}
		}
	}
	fmt.Println()

	// Example 7: Combined workflow
	fmt.Println("=== Example 7: Combined Workflow ===")
	fmt.Println("Demonstrating a complete feed processing pipeline:")
	fmt.Println("1. Discover feeds from multiple sources")
	fmt.Println("2. Fetch the discovered feeds")
	fmt.Println("3. Filter items by criteria")
	fmt.Println("4. Aggregate filtered results")
	fmt.Println("5. Convert to desired format")
	fmt.Println("6. Extract specific fields for analysis")
	
	// This demonstrates how the tools can work together
	// In a real application, you might use an agent to orchestrate this workflow
}

// createMockFeedResult creates a mock feed result for demonstration
func createMockFeedResult() map[string]interface{} {
	now := time.Now()
	return map[string]interface{}{
		"feed": map[string]interface{}{
			"title":       "Mock Technology Feed",
			"description": "A mock feed for demonstration purposes",
			"link":        "https://example.com/feed",
			"updated":     now.Format(time.RFC3339),
			"items": []interface{}{
				map[string]interface{}{
					"id":          "item1",
					"title":       "Go 1.22 Released with New Features",
					"description": "The latest version of Go includes exciting new features...",
					"link":        "https://example.com/go-1-22",
					"published":   now.Add(-24 * time.Hour).Format(time.RFC3339),
					"author":      map[string]interface{}{"name": "Go Team"},
					"categories":  []string{"golang", "programming", "technology"},
				},
				map[string]interface{}{
					"id":          "item2",
					"title":       "Building Microservices with Go",
					"description": "Learn how to build scalable microservices using Go...",
					"link":        "https://example.com/microservices",
					"published":   now.Add(-48 * time.Hour).Format(time.RFC3339),
					"author":      map[string]interface{}{"name": "Tech Writer"},
					"categories":  []string{"golang", "microservices", "architecture"},
				},
				map[string]interface{}{
					"id":          "item3",
					"title":       "Understanding Go Concurrency",
					"description": "Deep dive into goroutines and channels...",
					"link":        "https://example.com/concurrency",
					"published":   now.Add(-72 * time.Hour).Format(time.RFC3339),
					"author":      map[string]interface{}{"name": "Go Expert"},
					"categories":  []string{"golang", "concurrency", "programming"},
				},
			},
		},
		"status": 200,
		"format": "RSS2",
		"headers": map[string]string{
			"Content-Type": "application/rss+xml",
			"Last-Modified": now.Format(time.RFC1123),
		},
	}
}

// createMockDiscoverResult creates a mock discover result
func createMockDiscoverResult() map[string]interface{} {
	return map[string]interface{}{
		"feeds": []interface{}{
			map[string]interface{}{
				"url":         "https://example.com/blog/feed.xml",
				"type":        "RSS",
				"title":       "Example Blog RSS Feed",
				"auto_discovered": true,
			},
			map[string]interface{}{
				"url":         "https://example.com/blog/atom.xml",
				"type":        "Atom",
				"title":       "Example Blog Atom Feed",
				"auto_discovered": true,
			},
			map[string]interface{}{
				"url":         "https://example.com/podcast/feed.xml",
				"type":        "Podcast",
				"title":       "Example Podcast Feed",
				"media_type":  "audio/mpeg",
			},
		},
		"total_discovered": 3,
		"sources_checked": 5,
	}
}