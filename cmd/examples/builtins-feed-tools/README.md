# Built-in Feed Tools Example

This example demonstrates the usage of built-in feed processing tools in the go-llms library. These tools provide comprehensive feed handling capabilities including fetching, discovering, filtering, aggregating, converting, and extracting data from RSS, Atom, and JSON Feed formats.

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

## Notes

- The example includes fallback mechanisms for when real feeds are unavailable
- Mock data is provided for demonstration purposes
- In production, consider implementing caching and rate limiting
- Some feed providers may require authentication or have CORS restrictions