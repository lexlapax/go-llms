# Feed Processing Tools Implementation Plan

## Overview

This document outlines the plan for implementing feed processing tools as part of the go-llms built-in components. These tools will enable agents to fetch, parse, process, and extract information from various feed formats commonly used on the web.

## 1. Research on Common Feed Formats

### 1.1 RSS 2.0 (Really Simple Syndication)
- **Structure**: XML-based format
- **Key Elements**: 
  - `<channel>`: Container for feed metadata
  - `<item>`: Individual feed entries
  - `<title>`, `<link>`, `<description>`: Core item fields
  - `<pubDate>`, `<guid>`, `<author>`: Additional metadata
- **Usage**: Most common feed format, widely supported
- **Example**:
  ```xml
  <rss version="2.0">
    <channel>
      <title>Example Feed</title>
      <link>https://example.com</link>
      <item>
        <title>Article Title</title>
        <link>https://example.com/article</link>
        <description>Article summary</description>
        <pubDate>Mon, 06 Jan 2025 12:00:00 GMT</pubDate>
      </item>
    </channel>
  </rss>
  ```

### 1.2 Atom 1.0
- **Structure**: XML-based format with namespace
- **Key Elements**:
  - `<feed>`: Root element
  - `<entry>`: Individual entries
  - `<title>`, `<link>`, `<content>`: Core fields
  - `<updated>`, `<id>`, `<author>`: Metadata
- **Usage**: IETF standard, common in blogs and technical sites
- **Example**:
  ```xml
  <feed xmlns="http://www.w3.org/2005/Atom">
    <title>Example Feed</title>
    <link href="https://example.com"/>
    <entry>
      <title>Article Title</title>
      <link href="https://example.com/article"/>
      <content>Article content</content>
      <updated>2025-01-06T12:00:00Z</updated>
    </entry>
  </feed>
  ```

### 1.3 JSON Feed 1.1
- **Structure**: JSON-based format
- **Key Elements**:
  - `version`: Format version
  - `title`, `home_page_url`: Feed metadata
  - `items`: Array of feed entries
  - Entry fields: `id`, `url`, `title`, `content_text`/`content_html`
- **Usage**: Modern format, easier to parse, growing adoption
- **Example**:
  ```json
  {
    "version": "https://jsonfeed.org/version/1.1",
    "title": "Example Feed",
    "home_page_url": "https://example.com",
    "items": [
      {
        "id": "1",
        "url": "https://example.com/article",
        "title": "Article Title",
        "content_text": "Article content",
        "date_published": "2025-01-06T12:00:00Z"
      }
    ]
  }
  ```

### 1.4 RDF/RSS 1.0
- **Structure**: RDF-based XML format
- **Key Elements**: Similar to RSS 2.0 but with RDF namespace
- **Usage**: Less common, mostly legacy systems
- **Note**: Can be supported through RSS 2.0 parser with minor adjustments

## 2. Common Feed Processing Operations

### 2.1 Feed Discovery
- Auto-discovery of feed URLs from HTML pages
- Detection of feed type from content
- Validation of feed URLs

### 2.2 Feed Fetching
- HTTP/HTTPS retrieval with proper headers
- Handling redirects and authentication
- Respecting cache headers and ETags
- Rate limiting and politeness

### 2.3 Feed Parsing
- Parse different feed formats into unified structure
- Handle encoding issues (UTF-8, ISO-8859-1, etc.)
- Extract metadata and entries
- Handle malformed feeds gracefully

### 2.4 Feed Filtering
- Filter entries by date range
- Filter by keywords in title/content
- Filter by author or category
- Limit number of entries

### 2.5 Feed Aggregation
- Combine multiple feeds
- Sort entries by date
- Remove duplicates
- Merge similar content

### 2.6 Feed Transformation
- Convert between feed formats
- Extract specific fields
- Generate summaries
- Create custom output formats

### 2.7 Feed Monitoring
- Check for new entries
- Track changes in existing entries
- Detect feed updates
- Generate notifications

## 3. Proposed Tool Designs

### 3.1 FeedFetch Tool
```go
// FeedFetch retrieves and parses a feed from a URL
type FeedFetch struct {
    BaseTool
}

// Input parameters
type FeedFetchInput struct {
    URL        string            `json:"url" description:"Feed URL to fetch"`
    Headers    map[string]string `json:"headers,omitempty" description:"Optional HTTP headers"`
    Timeout    int               `json:"timeout,omitempty" description:"Timeout in seconds (default: 30)"`
    MaxItems   int               `json:"max_items,omitempty" description:"Maximum number of items to return"`
}

// Output structure
type FeedFetchOutput struct {
    Feed Feed `json:"feed" description:"Parsed feed data"`
}

type Feed struct {
    Type        string      `json:"type" description:"Feed type: rss, atom, json"`
    Title       string      `json:"title" description:"Feed title"`
    Description string      `json:"description,omitempty" description:"Feed description"`
    Link        string      `json:"link,omitempty" description:"Feed website URL"`
    Updated     time.Time   `json:"updated,omitempty" description:"Last update time"`
    Items       []FeedItem  `json:"items" description:"Feed entries"`
}

type FeedItem struct {
    ID          string    `json:"id" description:"Unique identifier"`
    Title       string    `json:"title" description:"Item title"`
    Link        string    `json:"link,omitempty" description:"Item URL"`
    Description string    `json:"description,omitempty" description:"Item summary"`
    Content     string    `json:"content,omitempty" description:"Full content"`
    Published   time.Time `json:"published,omitempty" description:"Publication date"`
    Updated     time.Time `json:"updated,omitempty" description:"Update date"`
    Author      string    `json:"author,omitempty" description:"Author name"`
    Categories  []string  `json:"categories,omitempty" description:"Categories/tags"`
}
```

### 3.2 FeedDiscover Tool
```go
// FeedDiscover finds feed URLs from a webpage
type FeedDiscover struct {
    BaseTool
}

type FeedDiscoverInput struct {
    URL     string `json:"url" description:"Webpage URL to search for feeds"`
    Timeout int    `json:"timeout,omitempty" description:"Timeout in seconds (default: 30)"`
}

type FeedDiscoverOutput struct {
    Feeds []DiscoveredFeed `json:"feeds" description:"Discovered feed URLs"`
}

type DiscoveredFeed struct {
    URL   string `json:"url" description:"Feed URL"`
    Type  string `json:"type" description:"Feed type: rss, atom, json"`
    Title string `json:"title,omitempty" description:"Feed title if available"`
}
```

### 3.3 FeedFilter Tool
```go
// FeedFilter filters feed items based on criteria
type FeedFilter struct {
    BaseTool
}

type FeedFilterInput struct {
    Feed       Feed      `json:"feed" description:"Feed to filter"`
    Keywords   []string  `json:"keywords,omitempty" description:"Keywords to match in title/content"`
    Authors    []string  `json:"authors,omitempty" description:"Filter by authors"`
    Categories []string  `json:"categories,omitempty" description:"Filter by categories"`
    After      time.Time `json:"after,omitempty" description:"Only items after this date"`
    Before     time.Time `json:"before,omitempty" description:"Only items before this date"`
    MaxItems   int       `json:"max_items,omitempty" description:"Maximum items to return"`
}

type FeedFilterOutput struct {
    Items        []FeedItem `json:"items" description:"Filtered feed items"`
    TotalItems   int        `json:"total_items" description:"Total items before filtering"`
    FilteredOut  int        `json:"filtered_out" description:"Number of items filtered out"`
}
```

### 3.4 FeedAggregate Tool
```go
// FeedAggregate combines multiple feeds into one
type FeedAggregate struct {
    BaseTool
}

type FeedAggregateInput struct {
    Feeds          []Feed `json:"feeds" description:"Feeds to aggregate"`
    SortBy         string `json:"sort_by,omitempty" description:"Sort field: date, title (default: date)"`
    SortDescending bool   `json:"sort_descending,omitempty" description:"Sort in descending order"`
    RemoveDupes    bool   `json:"remove_dupes,omitempty" description:"Remove duplicate items"`
    MaxItems       int    `json:"max_items,omitempty" description:"Maximum items in result"`
}

type FeedAggregateOutput struct {
    Feed         Feed `json:"feed" description:"Aggregated feed"`
    SourceCount  int  `json:"source_count" description:"Number of source feeds"`
    TotalItems   int  `json:"total_items" description:"Total items before limits"`
}
```

### 3.5 FeedConvert Tool
```go
// FeedConvert converts between feed formats
type FeedConvert struct {
    BaseTool
}

type FeedConvertInput struct {
    Feed       Feed   `json:"feed" description:"Feed to convert"`
    TargetType string `json:"target_type" description:"Target format: rss, atom, json"`
    Pretty     bool   `json:"pretty,omitempty" description:"Pretty-print output"`
}

type FeedConvertOutput struct {
    Content     string `json:"content" description:"Converted feed content"`
    ContentType string `json:"content_type" description:"MIME type of output"`
}
```

### 3.6 FeedExtract Tool
```go
// FeedExtract extracts specific data from feeds
type FeedExtract struct {
    BaseTool
}

type FeedExtractInput struct {
    Feed   Feed     `json:"feed" description:"Feed to extract from"`
    Fields []string `json:"fields" description:"Fields to extract: title, link, description, etc."`
    Format string   `json:"format,omitempty" description:"Output format: json, csv, text (default: json)"`
}

type FeedExtractOutput struct {
    Data   interface{} `json:"data" description:"Extracted data in requested format"`
    Format string      `json:"format" description:"Output format used"`
}
```

## 4. Implementation Considerations

### 4.1 Parsing Without External Dependencies

#### XML Parsing (RSS/Atom)
- Use Go's built-in `encoding/xml` package
- Handle namespaces properly for Atom feeds
- Create flexible structs with xml tags
- Handle CDATA sections in descriptions
- Example approach:
  ```go
  type flexibleXML struct {
      XMLName xml.Name
      Attrs   []xml.Attr `xml:",any,attr"`
      Content string     `xml:",chardata"`
      Children []flexibleXML `xml:",any"`
  }
  ```

#### JSON Parsing (JSON Feed)
- Use Go's built-in `encoding/json` package
- Use `json.RawMessage` for flexible parsing
- Handle optional fields with pointers
- Validate against JSON Feed spec

#### Content Type Detection
- Check HTTP Content-Type header
- Inspect first bytes of content
- Look for format-specific markers:
  - RSS: `<rss` or `<?xml` with `<channel>`
  - Atom: `<feed` with Atom namespace
  - JSON: `{` with `"version"` field

### 4.2 Error Handling
- Graceful degradation for malformed feeds
- Partial parsing when possible
- Clear error messages with context
- Validation of required fields
- Handling of encoding errors

### 4.3 Performance Considerations
- Stream parsing for large feeds
- Concurrent feed fetching
- Caching parsed results
- Memory-efficient data structures
- Connection pooling for HTTP requests

### 4.4 Security Considerations
- Validate URLs before fetching
- Set reasonable size limits
- Timeout handling
- Prevent XML external entity (XXE) attacks
- Sanitize HTML content in descriptions

### 4.5 Testing Strategy
- Unit tests for each parser
- Integration tests with real feeds
- Mock HTTP responses for testing
- Edge cases: empty feeds, huge feeds, malformed data
- Benchmark tests for performance

## 5. Example Use Cases

### 5.1 News Monitoring Agent
```go
// Agent that monitors news feeds and alerts on keywords
agent := NewAgent(
    WithTools(
        NewFeedFetch(),
        NewFeedFilter(),
    ),
)

// Fetch and filter tech news
response := agent.Run(ctx, "Fetch the latest from https://news.ycombinator.com/rss and filter for AI-related articles")
```

### 5.2 Content Aggregator
```go
// Agent that aggregates multiple feeds
agent := NewAgent(
    WithTools(
        NewFeedFetch(),
        NewFeedAggregate(),
        NewFeedConvert(),
    ),
)

// Aggregate tech blogs
response := agent.Run(ctx, "Fetch feeds from techcrunch.com, theverge.com, and arstechnica.com, aggregate them, and convert to JSON Feed format")
```

### 5.3 Feed Discovery Bot
```go
// Agent that discovers and analyzes feeds
agent := NewAgent(
    WithTools(
        NewFeedDiscover(),
        NewFeedFetch(),
        NewFeedExtract(),
    ),
)

// Discover and analyze
response := agent.Run(ctx, "Find all feeds on example.com and extract the titles and publication dates of recent posts")
```

### 5.4 Content Summarizer
```go
// Agent that summarizes feed content
agent := NewAgent(
    WithTools(
        NewFeedFetch(),
        NewFeedFilter(),
    ),
    WithLLM(llm),
)

// Summarize recent posts
response := agent.Run(ctx, "Fetch the feed from blog.example.com, get posts from the last week, and summarize the main topics")
```

### 5.5 Feed Format Converter Service
```go
// Convert between feed formats
agent := NewAgent(
    WithTools(
        NewFeedFetch(),
        NewFeedConvert(),
    ),
)

// Convert RSS to JSON Feed
response := agent.Run(ctx, "Fetch the RSS feed from podcast.example.com and convert it to JSON Feed format")
```

## Implementation Timeline

### Phase 1: Core Infrastructure (Week 1)
- Create feed data structures
- Implement basic XML/JSON parsing
- Set up test infrastructure

### Phase 2: Basic Tools (Week 2)
- Implement FeedFetch tool
- Implement FeedDiscover tool
- Add comprehensive tests

### Phase 3: Processing Tools (Week 3)
- Implement FeedFilter tool
- Implement FeedAggregate tool
- Add performance optimizations

### Phase 4: Advanced Tools (Week 4)
- Implement FeedConvert tool
- Implement FeedExtract tool
- Add documentation and examples

### Phase 5: Polish and Integration (Week 5)
- Integration with agent system
- Performance benchmarks
- Example applications
- Documentation completion

## Success Criteria

1. **Functionality**: All tools work correctly with major feed formats
2. **Performance**: Can process feeds with 1000+ items efficiently
3. **Reliability**: Handles malformed feeds gracefully
4. **Usability**: Clear API with good documentation
5. **Testing**: >90% test coverage with real-world feed tests
6. **Integration**: Seamless integration with existing agent system

## Future Enhancements

1. **Feed Monitoring**: Persistent monitoring with change detection
2. **Feed Generation**: Create feeds from other data sources
3. **Advanced Filtering**: ML-based content classification
4. **Feed Analytics**: Statistics and insights from feed data
5. **Webhook Support**: Real-time notifications for feed updates
6. **Feed Validation**: Comprehensive feed format validation
7. **Content Enrichment**: Add metadata from external sources