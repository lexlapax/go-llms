# Built-in Feed Tools Example

This comprehensive example demonstrates all 6 feed processing tools available in the go-llms library, showcasing their full capabilities with practical examples and proper type-safe result handling.

## Overview

The built-in feed tools provide comprehensive feed handling capabilities:
- **feed_fetch**: Fetch and parse RSS, Atom, and JSON Feed formats with HTTP optimizations
- **feed_discover**: Auto-discover feed URLs from websites using HTML analysis
- **feed_filter**: Filter feed items by keywords, dates, authors, and categories
- **feed_aggregate**: Combine multiple feeds with deduplication and sorting
- **feed_convert**: Convert between RSS, Atom, and JSON Feed formats
- **feed_extract**: Extract specific data fields for analysis or export

## Available Feed Tools

### 1. FeedFetch
Fetches and parses feeds from URLs, supporting RSS 2.0, Atom, and JSON Feed formats.

**Features:**
- Automatic format detection
- Conditional requests (ETag, If-Modified-Since)
- Custom user agents
- Item limiting
- Timeout control

**Example:**
```go
result, err := feedFetchTool.Execute(ctx, map[string]interface{}{
    "url":        "https://example.com/feed.xml",
    "max_items":  10,
    "timeout":    30,
    "user_agent": "MyApp/1.0",
    "etag":       "W/\"123456\"", // For conditional requests
})
```

### 2. FeedDiscover
Discovers feed URLs from websites by analyzing HTML for feed links.

**Features:**
- Auto-discovery from HTML link tags
- Deep link following
- Podcast feed detection
- Multiple format support

**Example:**
```go
result, err := feedDiscoverTool.Execute(ctx, map[string]interface{}{
    "url":              "https://example.com",
    "follow_links":     true,
    "max_depth":        2,
    "include_podcasts": true,
})
```

### 3. FeedFilter
Filters feed items based on various criteria.

**Features:**
- Keyword filtering (include/exclude)
- Date range filtering
- Category filtering
- Author filtering
- Custom sorting

**Example:**
```go
result, err := feedFilterTool.Execute(ctx, map[string]interface{}{
    "feed": feedData,
    "filters": map[string]interface{}{
        "keywords":         []string{"golang", "programming"},
        "exclude_keywords": []string{"spam"},
        "min_date":         "2024-01-01T00:00:00Z",
        "categories":       []string{"technology"},
    },
    "sort_by": "date",
    "limit":   20,
})
```

### 4. FeedAggregate
Combines multiple feeds into a single unified feed.

**Features:**
- Deduplication
- Custom sorting
- Metadata merging
- Item limiting

**Example:**
```go
result, err := feedAggregateTool.Execute(ctx, map[string]interface{}{
    "feeds": []interface{}{feed1, feed2, feed3},
    "merge_options": map[string]interface{}{
        "deduplicate":    true,
        "sort_by":        "date",
        "max_items":      50,
        "merge_metadata": true,
    },
})
```

### 5. FeedConvert
Converts feeds between different formats.

**Features:**
- RSS to Atom conversion
- Atom to JSON Feed conversion
- RSS to JSON Feed conversion
- Pretty printing options
- Version control

**Example:**
```go
result, err := feedConvertTool.Execute(ctx, map[string]interface{}{
    "feed":          feedData,
    "output_format": "json_feed",
    "options": map[string]interface{}{
        "pretty_print":    true,
        "include_content": true,
        "version":         "1.1",
    },
})
```

### 6. FeedExtract
Extracts specific data from feeds for analysis or export.

**Features:**
- Field selection
- CSV/JSON/XML export
- Custom transformations
- Date formatting

**Example:**
```go
result, err := feedExtractTool.Execute(ctx, map[string]interface{}{
    "feed": feedData,
    "extract_options": map[string]interface{}{
        "fields":         []string{"title", "link", "published"},
        "transform":      "csv",
        "include_header": true,
        "date_format":    "2006-01-02",
    },
})
```

## Running the Example

```bash
# From the project root
go run cmd/examples/builtins-feed-tools/main.go

# Or build and run
go build -o bin/builtins-feed-tools cmd/examples/builtins-feed-tools/main.go
./bin/builtins-feed-tools
```

## Use Cases

### 1. News Aggregation
Combine multiple news feeds, filter by topics, and export summaries:
```
Discover → Fetch → Filter → Aggregate → Extract
```

### 2. Podcast Management
Find podcast feeds, filter by date, convert to standard format:
```
Discover (podcasts) → Fetch → Filter (recent) → Convert
```

### 3. Content Monitoring
Monitor multiple feeds for specific keywords:
```
Fetch (multiple) → Filter (keywords) → Aggregate → Extract (alerts)
```

### 4. Feed Migration
Convert legacy RSS feeds to modern JSON Feed format:
```
Fetch → Convert (JSON Feed) → Validate
```

### 5. Data Analysis
Extract feed data for analysis in other tools:
```
Fetch → Filter → Extract (CSV) → Analyze
```

## Tool Examples in Registry

Each feed tool includes comprehensive examples in their metadata registration:

### feed_fetch (3 examples)
- **Basic RSS fetch**: Simple feed fetching
- **Fetch with limit**: Fetch only recent items
- **Conditional fetch**: Fetch only if modified using ETag

### feed_discover (3 examples)
- **Basic discovery**: Discover feeds from a blog homepage
- **Discovery with timeout**: Set custom timeout for slow sites
- **No redirects**: Discover feeds without following redirects

### feed_extract (3 examples)
- **Extract titles and links**: Get just titles and links from feed items
- **Extract with metadata**: Include feed metadata in extraction
- **Flatten nested fields**: Extract and flatten author information

### feed_filter (3 examples)
- **Filter by keywords**: Find items containing specific keywords
- **Filter by date range**: Get items from the last week
- **Complex filter**: Filter by multiple criteria with all conditions matching

### feed_aggregate (3 examples)
- **Combine news feeds**: Merge multiple news feeds into one
- **Sort by date descending**: Aggregate and sort by most recent first
- **Remove duplicates**: Combine feeds and remove duplicate articles

### feed_convert (3 examples)
- **Convert to RSS**: Convert any feed format to RSS 2.0
- **Convert to JSON Feed**: Convert RSS/Atom to modern JSON Feed format
- **Minimal Atom conversion**: Convert to Atom without full content

## Integration with Agents

These feed tools can be used with agents for automated feed processing:

```go
agent := workflow.NewAgent(
    "feed-processor",
    llmProvider,
    workflow.WithTools(
        tools.MustGetTool("feed_fetch"),
        tools.MustGetTool("feed_filter"),
        tools.MustGetTool("feed_extract"),
    ),
)

response, err := agent.Run(ctx, 
    "Find technology news about Go programming from the last week and create a summary",
    nil,
)
```

## Best Practices

1. **Rate Limiting**: Respect feed provider rate limits
2. **Caching**: Use ETags and If-Modified-Since headers
3. **Error Handling**: Handle network errors gracefully
4. **Timeout Settings**: Set appropriate timeouts for feed fetching
5. **Content Validation**: Validate feed content before processing

## Important Parameter Names

Each feed tool uses specific parameter names that must be matched exactly:

### feed_fetch Tool
- `url`: Feed URL to fetch (required)
- `max_items`: Maximum items to return
- `timeout`: Request timeout in seconds
- `user_agent`: Custom user agent string
- `etag`: ETag for conditional requests
- `if_modified`: If-Modified-Since header value

### feed_discover Tool  
- `url`: Website URL to discover feeds from (required)
- `follow_redirects`: Follow HTTP redirects (default: true)
- `timeout`: Request timeout in seconds
- `max_size`: Maximum response size in bytes

### feed_filter Tool
- `feed`: Feed data to filter (required)
- `keywords`: Keywords to match in title/content
- `authors`: Filter by author names
- `categories`: Filter by categories
- `after`: Only items published after this date (RFC3339)
- `before`: Only items published before this date (RFC3339)
- `max_items`: Maximum number of items to return
- `match_all`: If true, items must match ALL criteria

### feed_aggregate Tool
- `feeds`: Array of feed objects to combine (required)
- `deduplicate`: Remove duplicate items
- `sort_by`: Sort criteria ("date", "title")
- `max_items`: Maximum items in result

### feed_convert Tool
- `feed`: Feed data to convert (required)
- `target_type`: Target format ("json", "rss", "atom")
- `pretty_print`: Format output for readability
- `include_meta`: Include metadata in conversion

### feed_extract Tool
- `feed`: Feed data to extract from (required)
- `fields`: Array of field names to extract
- `max_items`: Maximum items to process
- `flatten`: Flatten nested data structures

## Type Assertions

When handling feed tool outputs, use the correct struct types:

```go
// feed_fetch results
if fetchRes, ok := result.(*feed.FeedFetchResult); ok {
    fmt.Printf("Format: %s\n", fetchRes.Format)
    fmt.Printf("Status: %d\n", fetchRes.Status)
    fmt.Printf("Title: %s\n", fetchRes.Feed.Title)
    fmt.Printf("Items: %d\n", len(fetchRes.Feed.Items))
    if len(fetchRes.Headers) > 0 {
        fmt.Printf("Headers: %d\n", len(fetchRes.Headers))
    }
}

// feed_discover results
if discoverRes, ok := result.(*feed.FeedDiscoverResult); ok {
    fmt.Printf("Discovered: %d feeds\n", len(discoverRes.Feeds))
    for _, feed := range discoverRes.Feeds {
        fmt.Printf("- %s (%s) from %s\n", feed.Title, feed.Type, feed.Source)
    }
    if discoverRes.Error != "" {
        fmt.Printf("Errors: %s\n", discoverRes.Error)
    }
}

// feed_filter results
if filterRes, ok := result.(*feed.FeedFilterResult); ok {
    fmt.Printf("Filtered: %d items\n", len(filterRes.Items))
    fmt.Printf("Total processed: %d\n", filterRes.TotalItems)
    fmt.Printf("Filtered out: %d\n", filterRes.FilteredOut)
}

// feed_aggregate results
if aggregateRes, ok := result.(*feed.FeedAggregateResult); ok {
    fmt.Printf("Sources: %d\n", aggregateRes.SourceCount)
    fmt.Printf("Total items: %d\n", aggregateRes.TotalItems)
    fmt.Printf("Duplicates removed: %d\n", aggregateRes.DupesRemoved)
    fmt.Printf("Final items: %d\n", len(aggregateRes.Feed.Items))
}

// feed_convert results
if convertRes, ok := result.(*feed.FeedConvertResult); ok {
    fmt.Printf("Format: %s\n", convertRes.Format)
    fmt.Printf("Content type: %s\n", convertRes.ContentType)
    fmt.Printf("Size: %d chars\n", len(convertRes.Content))
}

// feed_extract results
if extractRes, ok := result.(*feed.FeedExtractResult); ok {
    fmt.Printf("Fields: %v\n", extractRes.Fields)
    fmt.Printf("Items: %d\n", extractRes.Count)
    fmt.Printf("Data entries: %d\n", len(extractRes.Data))
    if len(extractRes.Metadata) > 0 {
        fmt.Printf("Metadata: %d fields\n", len(extractRes.Metadata))
    }
}
```

## Complete Working Examples

### Feed Fetching with Error Handling
```go
result, err := fetchTool.Execute(ctx, map[string]interface{}{
    "url":        "https://hnrss.org/frontpage",
    "max_items":  5,
    "timeout":    10,
    "user_agent": "MyApp/1.0",
})
if err != nil {
    log.Printf("Fetch failed: %v", err)
} else if fetchRes, ok := result.(*feed.FeedFetchResult); ok {
    fmt.Printf("Fetched %d items from %s\n", 
        len(fetchRes.Feed.Items), fetchRes.Feed.Title)
}
```

### Feed Discovery
```go
result, err := discoverTool.Execute(ctx, map[string]interface{}{
    "url":              "https://blog.golang.org",
    "follow_redirects": true,
    "timeout":          30,
})
if discoverRes, ok := result.(*feed.FeedDiscoverResult); ok {
    for _, feed := range discoverRes.Feeds {
        fmt.Printf("Found: %s (%s)\n", feed.Title, feed.URL)
    }
}
```

### Feed Filtering
```go
result, err := filterTool.Execute(ctx, map[string]interface{}{
    "feed":      feedData,
    "keywords":  []string{"golang", "programming"},
    "after":     "2024-01-01T00:00:00Z",
    "max_items": 10,
    "match_all": false,
})
if filterRes, ok := result.(*feed.FeedFilterResult); ok {
    fmt.Printf("Filtered to %d items\n", len(filterRes.Items))
}
```

## Performance Considerations

- **HTTP Caching**: Use ETags and If-Modified-Since for conditional requests
- **Rate Limiting**: Respect feed provider rate limits and robots.txt
- **Memory Usage**: Large feeds are processed in memory, consider max_items limits
- **Timeout Settings**: Set appropriate timeouts for network operations
- **Deduplication**: Use aggregation tool to remove duplicates across feeds

## Real-World Use Cases

1. **News Aggregation**: Combine multiple news feeds, filter by topics
2. **Podcast Management**: Discover and organize podcast feeds
3. **Content Monitoring**: Track specific keywords across feeds
4. **Feed Migration**: Convert between RSS, Atom, and JSON Feed formats
5. **Data Analysis**: Extract feed data for analysis in other tools
6. **Alert Systems**: Monitor feeds for specific content and trigger actions

## Integration with Agents

```go
agent := workflow.NewAgent(provider).
    SetSystemPrompt("You are a feed processing assistant.").
    AddTool(tools.MustGetTool("feed_fetch")).
    AddTool(tools.MustGetTool("feed_filter")).
    AddTool(tools.MustGetTool("feed_extract"))

result, _ := agent.Run(ctx, 
    "Fetch Hacker News RSS, filter for Go-related items, and extract titles and links")
```

## Notes

- The example includes fallback mechanisms for when real feeds are unavailable
- Mock data is provided for demonstration purposes with realistic feed structures
- In production, consider implementing caching and rate limiting
- Some feed providers may require authentication or have CORS restrictions
- All tools include comprehensive error handling and type-safe result processing