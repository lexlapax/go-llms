// ABOUTME: Comprehensive example demonstrating all built-in feed processing tools
// ABOUTME: Shows fetching, discovering, filtering, aggregating, converting, and extracting feed data

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	feedtools "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/feed"
	"github.com/lexlapax/go-llms/pkg/agent/core"
	agentDomain "github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
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

	// Demonstrate tool discovery
	fmt.Println("=== Available Feed Tools ===")
	fmt.Println()
	feedTools := tools.Tools.ListByCategory("feed")
	fmt.Printf("Total feed tools: %d\n", len(feedTools))
	for _, entry := range feedTools {
		fmt.Printf("• %s: %s\n", entry.Metadata.Name, entry.Metadata.Description)
		fmt.Printf("  Version: %s\n", entry.Metadata.Version)
		fmt.Printf("  Tags: %v\n", entry.Metadata.Tags)
		fmt.Println()
	}

	// Get all the feed tools
	fetchTool := tools.MustGetTool("feed_fetch")
	discoverTool := tools.MustGetTool("feed_discover")
	filterTool := tools.MustGetTool("feed_filter")
	aggregateTool := tools.MustGetTool("feed_aggregate")
	convertTool := tools.MustGetTool("feed_convert")
	extractTool := tools.MustGetTool("feed_extract")

	// Example 1: Fetch a feed
	fmt.Println("=== Example 1: Feed Fetch ===")
	fmt.Println("Attempting to fetch a real feed (with fallback to mock data)...")

	var fetchResult interface{}
	result, err := fetchTool.Execute(toolCtx, map[string]interface{}{
		"url":        "https://hnrss.org/frontpage",
		"max_items":  5,
		"timeout":    10,
		"user_agent": "go-llms-feed-example/1.0",
	})
	if err != nil {
		fmt.Printf("Primary feed failed: %v, trying backup...\n", err)
		result, err = fetchTool.Execute(toolCtx, map[string]interface{}{
			"url":       "https://www.reddit.com/r/golang/.rss",
			"max_items": 5,
			"timeout":   10,
		})
		if err != nil {
			log.Printf("Warning: Feed fetch failed: %v", err)
			// Continue with mock data for other examples
			fetchResult = createMockFeedFetchResult()
			fmt.Println("Using mock data for demonstration")
		} else {
			fetchResult = result
		}
	} else {
		fetchResult = result
	}

	if fetchRes, ok := fetchResult.(*feedtools.FeedFetchResult); ok {
		fmt.Printf("✓ Fetched feed successfully\n")
		fmt.Printf("  Format: %s\n", fetchRes.Format)
		fmt.Printf("  Status: %d\n", fetchRes.Status)
		fmt.Printf("  Feed title: %s\n", fetchRes.Feed.Title)
		fmt.Printf("  Feed description: %s\n", fetchRes.Feed.Description)
		fmt.Printf("  Number of items: %d\n", len(fetchRes.Feed.Items))
		if len(fetchRes.Feed.Items) > 0 {
			fmt.Printf("  Latest item: %s\n", fetchRes.Feed.Items[0].Title)
			if fetchRes.Feed.Items[0].Published != nil {
				fmt.Printf("  Published: %s\n", fetchRes.Feed.Items[0].Published.Format("2006-01-02 15:04"))
			}
		}
		if len(fetchRes.Headers) > 0 {
			fmt.Printf("  Response headers: %d\n", len(fetchRes.Headers))
		}
	} else {
		fmt.Printf("Unexpected result type for feed fetch operation\n")
		// Fallback to mock data
		fetchResult = createMockFeedFetchResult()
	}
	fmt.Println()

	// Example 2: Discover feeds from a website
	fmt.Println("=== Example 2: Feed Discovery ===")
	fmt.Println("Attempting to discover feeds from a website...")

	discoverResult, err := discoverTool.Execute(toolCtx, map[string]interface{}{
		"url":              "https://blog.golang.org",
		"follow_links":     true,
		"max_depth":        2,
		"include_podcasts": true,
	})
	if err != nil {
		log.Printf("Warning: Feed discovery failed: %v", err)
		// Continue with mock data
		discoverResult = createMockDiscoverResult()
		fmt.Println("Using mock discovery data for demonstration")
	}

	if discoverRes, ok := discoverResult.(*feedtools.FeedDiscoverResult); ok {
		fmt.Printf("✓ Feed discovery completed\n")
		fmt.Printf("  Discovered %d feeds:\n", len(discoverRes.Feeds))
		for i, discoveredFeed := range discoverRes.Feeds {
			if i < 3 { // Show first 3 feeds
				fmt.Printf("    %d. %s (%s)\n", i+1, discoveredFeed.Title, discoveredFeed.Type)
				fmt.Printf("       URL: %s\n", discoveredFeed.URL)
				fmt.Printf("       Source: %s\n", discoveredFeed.Source)
			}
		}
		if len(discoverRes.Feeds) > 3 {
			fmt.Printf("    ... and %d more feeds\n", len(discoverRes.Feeds)-3)
		}
		if discoverRes.Error != "" {
			fmt.Printf("  Discovery warnings: %s\n", discoverRes.Error)
		}
	} else {
		fmt.Printf("Unexpected result type for feed discovery operation\n")
	}
	fmt.Println()

	// Example 3: Filter feed items
	fmt.Println("=== Example 3: Feed Filtering ===")
	fmt.Println("Filtering feed items by keywords and date range...")

	// Get the fetch result for filtering
	var feedForFilter interface{}
	if fetchRes, ok := fetchResult.(*feedtools.FeedFetchResult); ok {
		feedForFilter = fetchRes.Feed
	} else {
		feedForFilter = createMockFeedFetchResult().Feed
	}

	filterResult, err := filterTool.Execute(toolCtx, map[string]interface{}{
		"feed":      feedForFilter,
		"keywords":  []string{"go", "golang", "programming"},
		"after":     time.Now().AddDate(0, 0, -30).Format(time.RFC3339), // Last 30 days
		"max_items": 10,
		"match_all": false, // Match ANY keyword, not all
	})
	if err != nil {
		log.Printf("Warning: Feed filtering failed: %v", err)
	} else {
		if filterRes, ok := filterResult.(*feedtools.FeedFilterResult); ok {
			fmt.Printf("✓ Feed filtering completed\n")
			fmt.Printf("  Filtered items: %d\n", len(filterRes.Items))
			fmt.Printf("  Total items processed: %d\n", filterRes.TotalItems)
			fmt.Printf("  Items filtered out: %d\n", filterRes.FilteredOut)

			for i, item := range filterRes.Items {
				if i < 3 { // Show first 3 filtered items
					fmt.Printf("    %d. %s\n", i+1, item.Title)
					if item.Published != nil {
						fmt.Printf("       Published: %s\n", item.Published.Format("2006-01-02"))
					}
					if len(item.Categories) > 0 {
						fmt.Printf("       Categories: %v\n", item.Categories)
					}
				}
			}
			if len(filterRes.Items) > 3 {
				fmt.Printf("    ... and %d more items\n", len(filterRes.Items)-3)
			}
		} else {
			fmt.Printf("Unexpected result type for feed filtering operation\n")
		}
	}
	fmt.Println()

	// Example 4: Aggregate multiple feeds
	fmt.Println("=== Example 4: Feed Aggregation ===")
	fmt.Println("Aggregating multiple feeds into one unified feed..")

	// Create multiple feeds for aggregation
	feed1 := createMockFeedFetchResult().Feed
	feed2 := createMockFeedFetchResult().Feed

	// Modify feed2 to be different
	feed2.Title = "Secondary Tech Feed"
	feed2.Description = "Another technology news source"
	if len(feed2.Items) > 0 {
		feed2.Items[0].Title = "Advanced Go Patterns and Best Practices"
		feed2.Items[0].ID = "secondary-item-1"
	}

	aggregateResult, err := aggregateTool.Execute(toolCtx, map[string]interface{}{
		"feeds":       []interface{}{feed1, feed2},
		"deduplicate": true,
		"sort_by":     "date",
		"max_items":   20,
	})
	if err != nil {
		log.Printf("Warning: Feed aggregation failed: %v", err)
	} else {
		if aggregateRes, ok := aggregateResult.(*feedtools.FeedAggregateResult); ok {
			fmt.Printf("✓ Feed aggregation completed\n")
			fmt.Printf("  Source feeds: %d\n", aggregateRes.SourceCount)
			fmt.Printf("  Total items before deduplication: %d\n", aggregateRes.TotalItems)
			fmt.Printf("  Duplicates removed: %d\n", aggregateRes.DupesRemoved)
			fmt.Printf("  Final aggregated items: %d\n", len(aggregateRes.Feed.Items))
			fmt.Printf("  Aggregated feed title: %s\n", aggregateRes.Feed.Title)

			if len(aggregateRes.Feed.Items) > 0 {
				fmt.Printf("  Sample aggregated items:\n")
				for i, item := range aggregateRes.Feed.Items {
					if i < 3 {
						fmt.Printf("    %d. %s\n", i+1, item.Title)
						if item.Published != nil {
							fmt.Printf("       Date: %s\n", item.Published.Format("2006-01-02 15:04"))
						}
					}
				}
			}
		} else {
			fmt.Printf("Unexpected result type for feed aggregation operation\n")
		}
	}
	fmt.Println()

	// Example 5: Convert feed format
	fmt.Println("=== Example 5: Feed Format Conversion ===")
	fmt.Println("Converting feed to different formats...")

	convertResult, err := convertTool.Execute(toolCtx, map[string]interface{}{
		"feed":         feedForFilter,
		"target_type":  "json",
		"pretty_print": true,
		"include_meta": true,
	})
	if err != nil {
		log.Printf("Warning: Feed conversion failed: %v", err)
	} else {
		if convertRes, ok := convertResult.(*feedtools.FeedConvertResult); ok {
			fmt.Printf("✓ Feed conversion completed\n")
			fmt.Printf("  Target format: %s\n", convertRes.Format)
			fmt.Printf("  Content type: %s\n", convertRes.ContentType)
			fmt.Printf("  Content size: %d characters\n", len(convertRes.Content))

			// Show a preview of the converted content
			if len(convertRes.Content) > 0 {
				fmt.Printf("  Content preview:\n")
				preview := convertRes.Content
				if len(preview) > 200 {
					preview = preview[:200] + "..."
				}
				// Indent each line for better readability
				lines := []rune(preview)
				fmt.Printf("    %s\n", string(lines))
			}
		} else {
			fmt.Printf("Unexpected result type for feed conversion operation\n")
		}
	}
	fmt.Println()

	// Example 6: Extract specific data from feeds
	fmt.Println("=== Example 6: Feed Data Extraction ===")
	fmt.Println("Extracting specific fields from feed items...")

	extractResult, err := extractTool.Execute(toolCtx, map[string]interface{}{
		"feed":      feedForFilter,
		"fields":    []string{"title", "link", "published", "author"},
		"max_items": 5,
		"flatten":   true,
	})
	if err != nil {
		log.Printf("Warning: Feed extraction failed: %v", err)
	} else {
		if extractRes, ok := extractResult.(*feedtools.FeedExtractResult); ok {
			fmt.Printf("✓ Feed data extraction completed\n")
			fmt.Printf("  Extracted fields: %v\n", extractRes.Fields)
			fmt.Printf("  Items extracted: %d\n", extractRes.Count)

			if len(extractRes.Data) > 0 {
				fmt.Printf("  Sample extracted data:\n")
				for i, item := range extractRes.Data {
					if i < 3 {
						fmt.Printf("    Item %d:\n", i+1)
						for key, value := range item {
							fmt.Printf("      %s: %v\n", key, value)
						}
						fmt.Println()
					}
				}
				if len(extractRes.Data) > 3 {
					fmt.Printf("    ... and %d more items\n", len(extractRes.Data)-3)
				}
			}

			if len(extractRes.Metadata) > 0 {
				fmt.Printf("  Extraction metadata: %d fields\n", len(extractRes.Metadata))
			}
		} else {
			fmt.Printf("Unexpected result type for feed extraction operation\n")
		}
	}
	fmt.Println()

	// Example 7: Using feed tools with an agent
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey != "" {
		fmt.Println("=== Example 7: Agent with Feed Tools ===")
		fmt.Println("Demonstrating agent-based feed processing...")

		// Create provider and agent
		p := provider.NewOpenAIProvider(apiKey, "gpt-4o-mini")
		agent := core.NewAgent("feed-processing-agent", p)
		agent.SetSystemPrompt("You are a feed processing assistant that can fetch, filter, and analyze RSS/Atom feeds.")
		agent.AddTool(fetchTool)
		agent.AddTool(filterTool)
		agent.AddTool(extractTool)

		// Use the agent for feed analysis
		state := agentDomain.NewState()
		state.Set("prompt", "Fetch the Hacker News RSS feed, filter for items containing 'Go' or 'golang', and extract the titles and links of the top 3 items.")
		resultState, err := agent.Run(ctx, state)
		if err != nil {
			log.Printf("Error running feed agent: %v", err)
		} else {
			if result, exists := resultState.Get("result"); exists {
				fmt.Printf("Agent response: %v\n", result)
			}
		}
	} else {
		fmt.Println("=== Agent Example Skipped (no API key) ===")
		fmt.Println("Set OPENAI_API_KEY to see agent-based feed processing")
	}
	fmt.Println()

	// Example 8: Combined workflow demonstration
	fmt.Println("=== Example 8: Complete Feed Processing Workflow ===")
	fmt.Println("Demonstrating a complete feed processing pipeline:")
	fmt.Println("1. ✓ Discover feeds from websites")
	fmt.Println("2. ✓ Fetch discovered feeds")
	fmt.Println("3. ✓ Filter items by criteria (keywords, dates)")
	fmt.Println("4. ✓ Aggregate multiple filtered feeds")
	fmt.Println("5. ✓ Convert to desired output format")
	fmt.Println("6. ✓ Extract specific data fields")
	fmt.Println()

	// Show tool examples from metadata
	fmt.Println("=== Tool Usage Examples ===")
	for _, entry := range feedTools {
		fmt.Printf("\n%s examples:\n", entry.Metadata.Name)
		for _, example := range entry.Metadata.Examples {
			fmt.Printf("  %s: %s\n", example.Name, example.Description)
			fmt.Printf("    %s\n", example.Code)
		}
	}

	fmt.Println("\n=== Summary ===")
	fmt.Println("This example demonstrated all 6 feed tools:")
	fmt.Println("• feed_fetch: Fetch and parse RSS/Atom/JSON feeds with HTTP optimizations")
	fmt.Println("• feed_discover: Auto-discover feed URLs from websites")
	fmt.Println("• feed_filter: Filter feed items by keywords, dates, authors, categories")
	fmt.Println("• feed_aggregate: Combine multiple feeds with deduplication and sorting")
	fmt.Println("• feed_convert: Convert between RSS, Atom, and JSON Feed formats")
	fmt.Println("• feed_extract: Extract specific data fields for analysis or export")
	fmt.Println("\nAll operations include comprehensive error handling and type-safe result processing.")
}

// createMockFeedFetchResult creates a mock feed fetch result for demonstration
func createMockFeedFetchResult() *feedtools.FeedFetchResult {
	now := time.Now()
	pub1 := now.Add(-24 * time.Hour)
	pub2 := now.Add(-48 * time.Hour)
	pub3 := now.Add(-72 * time.Hour)

	return &feedtools.FeedFetchResult{
		Feed: feedtools.UnifiedFeed{
			Title:       "Mock Technology Feed",
			Description: "A mock feed for demonstration purposes",
			Link:        "https://example.com/feed",
			Updated:     &now,
			Language:    "en",
			Author: &feedtools.Author{
				Name:  "Tech Team",
				Email: "tech@example.com",
			},
			Items: []feedtools.FeedItem{
				{
					ID:          "item1",
					Title:       "Go 1.22 Released with New Features",
					Description: "The latest version of Go includes exciting new features and improvements...",
					Content:     "Full article content about Go 1.22 release with detailed feature descriptions.",
					Link:        "https://example.com/go-1-22",
					Published:   &pub1,
					Updated:     &pub1,
					Author: &feedtools.Author{
						Name: "Go Team",
					},
					Categories: []string{"golang", "programming", "technology", "release"},
				},
				{
					ID:          "item2",
					Title:       "Building Microservices with Go",
					Description: "Learn how to build scalable microservices using Go and modern patterns...",
					Content:     "Comprehensive guide to microservices architecture with Go examples.",
					Link:        "https://example.com/microservices",
					Published:   &pub2,
					Updated:     &pub2,
					Author: &feedtools.Author{
						Name: "Tech Writer",
					},
					Categories: []string{"golang", "microservices", "architecture", "programming"},
				},
				{
					ID:          "item3",
					Title:       "Understanding Go Concurrency Patterns",
					Description: "Deep dive into goroutines, channels, and advanced concurrency patterns...",
					Content:     "Advanced tutorial on Go concurrency with practical examples.",
					Link:        "https://example.com/concurrency",
					Published:   &pub3,
					Updated:     &pub3,
					Author: &feedtools.Author{
						Name:  "Go Expert",
						Email: "expert@example.com",
					},
					Categories: []string{"golang", "concurrency", "programming", "tutorial"},
				},
			},
		},
		Status: 200,
		Format: "RSS2",
		Headers: map[string]string{
			"Content-Type":  "application/rss+xml",
			"Last-Modified": now.Format(time.RFC1123),
			"ETag":          `"mock-etag-123456"`,
		},
		NotModified: false,
	}
}

// createMockDiscoverResult creates a mock discover result
func createMockDiscoverResult() *feedtools.FeedDiscoverResult {
	return &feedtools.FeedDiscoverResult{
		Feeds: []feedtools.DiscoveredFeed{
			{
				URL:    "https://example.com/blog/feedtools.xml",
				Type:   "RSS",
				Title:  "Example Blog RSS Feed",
				Source: "link_tag",
			},
			{
				URL:    "https://example.com/blog/atom.xml",
				Type:   "Atom",
				Title:  "Example Blog Atom Feed",
				Source: "auto_discovery",
			},
			{
				URL:    "https://example.com/podcast/feedtools.xml",
				Type:   "Podcast",
				Title:  "Example Podcast Feed",
				Source: "common_path",
			},
		},
		Error: "",
	}
}
