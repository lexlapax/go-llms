# Go-LLMs Tool Usage Examples

This document provides practical examples and patterns for using the built-in tools in Go-LLMs.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Common Patterns](#common-patterns)
3. [Real-World Scenarios](#real-world-scenarios)
4. [Advanced Integration](#advanced-integration)
5. [Performance Optimization](#performance-optimization)
6. [Troubleshooting](#troubleshooting)

## Quick Start

### Basic Tool Execution

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/file"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
)

func main() {
    // Create a tool
    readTool := file.ReadFile()
    
    // Create context
    ctx := &domain.ToolContext{
        Context: context.Background(),
        State:   domain.NewState(),
        Events:  domain.NewEventEmitter(),
    }
    
    // Execute tool
    result, err := readTool.Execute(ctx, map[string]interface{}{
        Path: "/etc/hosts",
        IncludeMeta: true,
}
    
    if err != nil {
        log.Fatal(err)
    }
    
    // Type assert result
    fileResult := result.(*file.ReadFileResult)
    fmt.Printf("Content: %s\n", fileResult.Content)
    if fileResult.Metadata != nil {
        fmt.Printf("File size: %d bytes\n", fileResult.Metadata.Size)
    }
}
```

### Using Multiple Tools

```go
func analyzeWebContent() error {
    ctx := &domain.ToolContext{
        Context: context.Background(),
        State:   domain.NewState(),
        Events:  domain.NewEventEmitter(),
    }
    
    // Fetch web content
    fetchTool := web.WebFetch()
    fetchResult, err := fetchTool.Execute(ctx, web.WebFetchParams{
        URL: "https://example.com/data.json",
        Selector: "body",
}
    if err != nil {
        return err
    }
    
    // Process JSON
    jsonTool := data.JSONProcess()
    jsonResult, err := jsonTool.Execute(ctx, data.JSONProcessInput{
        Data: fetchResult.(*web.WebFetchResult).Content,
        Operation: "query",
        JSONPath: "$.users[*].email",
}
    if err != nil {
        return err
    }
    
    // Save results
    writeTool := file.WriteFile()
    _, err = writeTool.Execute(ctx, map[string]interface{}{
        Path: "/tmp/emails.json",
        Content: fmt.Sprintf("%v", jsonResult.(*data.JSONProcessOutput).Result),
}
    
    return err
}
```

## Common Patterns

### Pattern 1: File Processing Pipeline

```go
func processLargeLogFile(logPath string) error {
    ctx := createContext()
    
    // Step 1: Search for error lines
    searchTool := file.FileSearch()
    searchResult, err := searchTool.Execute(ctx, file.FileSearchParams{
        Path: filepath.Dir(logPath),
        Pattern: filepath.Base(logPath),
        Content: "ERROR",
}
    if err != nil {
        return err
    }
    
    // Step 2: Read specific portions
    readTool := file.ReadFile()
    for _, match := range searchResult.(*file.FileSearchResult).Matches {
        result, err := readTool.Execute(ctx, map[string]interface{}{
            Path: match.Path,
            LineStart: match.LineNumber - 5,
            LineEnd: match.LineNumber + 5,
}
        if err != nil {
            continue
        }
        
        // Process error context
        processErrorContext(result.(*file.ReadFileResult).Content)
    }
    
    return nil
}
```

### Pattern 2: API Data Collection

```go
func collectAPIData(apiURL string, endpoints []string) (map[string]interface{}, error) {
    ctx := createContext()
    httpTool := web.HTTPRequest()
    results := make(map[string]interface{})
    
    for _, endpoint := range endpoints {
        result, err := httpTool.Execute(ctx, map[string]interface{}{
            URL: apiURL + endpoint,
            Method: "GET",
            Headers: map[string]string{
                "Accept": "application/json",
            },
            AuthType: "bearer",
            AuthToken: os.Getenv("API_TOKEN"),
            Timeout: 30,
}
        
        if err != nil {
            log.Printf("Failed to fetch %s: %v", endpoint, err)
            continue
        }
        
        httpResult := result.(*web.HTTPRequestResult)
        if httpResult.StatusCode == 200 {
            var data interface{}
            json.Unmarshal([]byte(httpResult.Body), &data)
            results[endpoint] = data
        }
    }
    
    return results, nil
}
```

### Pattern 3: System Monitoring

```go
func monitorSystem(interval time.Duration) {
    ctx := createContext()
    sysInfoTool := system.GetSystemInfo()
    procListTool := system.ProcessList()
    
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    
    for range ticker.C {
        // Get system info
        sysInfo, err := sysInfoTool.Execute(ctx, system.GetSystemInfoParams{
            IncludeMemory: true,
            IncludeRuntime: true,
}
        if err == nil {
            info := sysInfo.(*system.SystemInfo)
            log.Printf("Memory: %d MB used", info.Memory.Alloc/1024/1024)
        }
        
        // Check specific processes
        processes, err := procListTool.Execute(ctx, system.ProcessListParams{
            Filter: "myapp",
            SortBy: "cpu",
}
        if err == nil {
            procResult := processes.(*system.ProcessListResult)
            for _, proc := range procResult.Processes {
                if proc.CPUPercent > 80 {
                    log.Printf("High CPU usage: %s (%.1f%%)", 
                        proc.Name, proc.CPUPercent)
                }
            }
        }
    }
}
```

## Real-World Scenarios

### Scenario 1: Data Migration

```go
func migrateData(sourceDir, destURL string) error {
    ctx := createContext()
    
    // List all CSV files
    listTool := file.FileList()
    files, err := listTool.Execute(ctx, file.FileListParams{
        Path: sourceDir,
        Pattern: "*.csv",
        Recursive: true,
}
    if err != nil {
        return err
    }
    
    fileList := files.(*file.FileListResult)
    readTool := file.ReadFile()
    csvTool := data.CSVProcess()
    httpTool := web.HTTPRequest()
    
    for _, fileInfo := range fileList.Files {
        // Read CSV file
        content, err := readTool.Execute(ctx, map[string]interface{}{
            Path: fileInfo.Path,
}
        if err != nil {
            continue
        }
        
        // Parse CSV
        parsed, err := csvTool.Execute(ctx, data.CSVProcessInput{
            Data: content.(*file.ReadFileResult).Content,
            Operation: "parse",
            Headers: true,
}
        if err != nil {
            continue
        }
        
        // Upload to API
        jsonData, _ := json.Marshal(parsed.(*data.CSVProcessOutput).Result)
        _, err = httpTool.Execute(ctx, map[string]interface{}{
            URL: destURL + "/import",
            Method: "POST",
            Body: string(jsonData),
            BodyType: "json",
            Headers: map[string]string{
                "X-Source-File": fileInfo.Name,
            },
}
        
        if err != nil {
            log.Printf("Failed to upload %s: %v", fileInfo.Name, err)
        }
    }
    
    return nil
}
```

### Scenario 2: Web Scraping and Analysis

```go
func analyzeCompetitorPrices(urls []string) (map[string]float64, error) {
    ctx := createContext()
    scrapeTool := web.WebScrape()
    prices := make(map[string]float64)
    
    rules := map[string]string{
        "price": ".product-price",
        "name": ".product-name",
        "currency": ".price-currency",
    }
    
    for _, url := range urls {
        result, err := scrapeTool.Execute(ctx, web.WebScrapeParams{
            URL: url,
            Rules: rules,
            JavaScript: true, // Handle dynamic content
            WaitFor: ".product-price",
}
        
        if err != nil {
            log.Printf("Failed to scrape %s: %v", url, err)
            continue
        }
        
        scraped := result.(*web.WebScrapeResult)
        if priceStr, ok := scraped.Data["price"].(string); ok {
            // Parse price
            price := parsePrice(priceStr)
            productName := scraped.Data["name"].(string)
            prices[productName] = price
        }
    }
    
    return prices, nil
}
```

### Scenario 3: Log Analysis and Alerting

```go
func analyzeLogsAndAlert(logDir string, alertWebhook string) error {
    ctx := createContext()
    
    // Get current time for date calculations
    nowTool := datetime.DateTimeNow()
    now, _ := nowTool.Execute(ctx, datetime.DateTimeNowInput{})
    currentTime := now.(*datetime.DateTimeNowOutput).UTC
    
    // Calculate time range (last hour)
    calcTool := datetime.DateTimeCalculate()
    oneHourAgo, _ := calcTool.Execute(ctx, datetime.DateTimeCalculateInput{
        Base: currentTime,
        Operation: "subtract",
        Duration: "1h",
}
    
    // Search recent log files
    searchTool := file.FileSearch()
    results, err := searchTool.Execute(ctx, file.FileSearchParams{
        Path: logDir,
        Pattern: "*.log",
        ModifiedAfter: oneHourAgo.(*datetime.DateTimeCalculateOutput).Result,
}
    if err != nil {
        return err
    }
    
    // Analyze each file
    readTool := file.ReadFile()
    httpTool := web.HTTPRequest()
    
    errorCount := 0
    searchResults := results.(*file.FileSearchResult)
    
    for _, match := range searchResults.Matches {
        content, err := readTool.Execute(ctx, map[string]interface{}{
            Path: match.Path,
}
        if err != nil {
            continue
        }
        
        // Count errors
        lines := strings.Split(content.(*file.ReadFileResult).Content, "\n")
        for _, line := range lines {
            if strings.Contains(line, "ERROR") || strings.Contains(line, "FATAL") {
                errorCount++
            }
        }
    }
    
    // Send alert if threshold exceeded
    if errorCount > 100 {
        alert := map[string]interface{}{
            "level": "critical",
            "message": fmt.Sprintf("High error rate detected: %d errors in last hour", errorCount),
            "timestamp": currentTime,
            "source": "log-analyzer",
        }
        
        alertJSON, _ := json.Marshal(alert)
        _, err = httpTool.Execute(ctx, map[string]interface{}{
            URL: alertWebhook,
            Method: "POST",
            Body: string(alertJSON),
            BodyType: "json",
}
    }
    
    return nil
}
```

## Advanced Integration

### Custom Tool Wrapper

```go
// Create a wrapper that adds retry logic
type RetryableToolWrapper struct {
    tool       domain.Tool
    maxRetries int
    delay      time.Duration
}

func (r *RetryableToolWrapper) Execute(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
    var lastErr error
    
    for i := 0; i <= r.maxRetries; i++ {
        result, err := r.tool.Execute(ctx, params)
        if err == nil {
            return result, nil
        }
        
        lastErr = err
        if i < r.maxRetries {
            time.Sleep(r.delay * time.Duration(i+1))
        }
    }
    
    return nil, fmt.Errorf("failed after %d retries: %w", r.maxRetries, lastErr)
}

// Usage
retryableFetch := &RetryableToolWrapper{
    tool:       web.WebFetch(),
    maxRetries: 3,
    delay:      time.Second,
}
```

### Tool Result Caching

```go
type CachedToolExecutor struct {
    cache sync.Map
    ttl   time.Duration
}

type cacheEntry struct {
    result    interface{}
    timestamp time.Time
}

func (c *CachedToolExecutor) Execute(tool domain.Tool, ctx *domain.ToolContext, params interface{}) (interface{}, error) {
    // Generate cache key
    key := fmt.Sprintf("%s:%v", tool.Name(), params)
    
    // Check cache
    if cached, ok := c.cache.Load(key); ok {
        entry := cached.(cacheEntry)
        if time.Since(entry.timestamp) < c.ttl {
            return entry.result, nil
        }
    }
    
    // Execute tool
    result, err := tool.Execute(ctx, params)
    if err != nil {
        return nil, err
    }
    
    // Cache result
    c.cache.Store(key, cacheEntry{
        result:    result,
        timestamp: time.Now(),
}
    
    return result, nil
}
```

### Event-Driven Tool Execution

```go
func setupEventDrivenExecution() {
    ctx := &domain.ToolContext{
        Context: context.Background(),
        State:   domain.NewState(),
        Events:  domain.NewEventEmitter(),
    }
    
    // Set up event handlers
    ctx.Events.On("file_modified", func(event domain.Event) {
        data := event.Data.(map[string]interface{})
        filePath := data["path"].(string)
        
        // Automatically process modified files
        go processModifiedFile(ctx, filePath)
}
    
    ctx.Events.On("error", func(event domain.Event) {
        err := event.Data.(error)
        log.Printf("Tool error: %v", err)
        
        // Send notification
        notifyError(err)
}
    
    ctx.Events.On("progress", func(event domain.Event) {
        progress := event.Data.(domain.ProgressData)
        updateProgressBar(progress.Current, progress.Total, progress.Message)
}
}
```

## Performance Optimization

### Parallel Tool Execution

```go
func fetchMultipleURLsConcurrently(urls []string) ([]interface{}, error) {
    ctx := createContext()
    fetchTool := web.WebFetch()
    
    // Create result channel
    type result struct {
        url    string
        data   interface{}
        err    error
    }
    resultChan := make(chan result, len(urls))
    
    // Launch goroutines
    var wg sync.WaitGroup
    semaphore := make(chan struct{}, 10) // Limit concurrency
    
    for _, url := range urls {
        wg.Add(1)
        go func(u string) {
            defer wg.Done()
            
            semaphore <- struct{}{} // Acquire
            defer func() { <-semaphore }() // Release
            
            data, err := fetchTool.Execute(ctx, web.WebFetchParams{
                URL: u,
                Timeout: 10,
}
            
            resultChan <- result{url: u, data: data, err: err}
        }(url)
    }
    
    // Wait and close channel
    go func() {
        wg.Wait()
        close(resultChan)
    }()
    
    // Collect results
    var results []interface{}
    for r := range resultChan {
        if r.err != nil {
            log.Printf("Failed to fetch %s: %v", r.url, r.err)
            continue
        }
        results = append(results, r.data)
    }
    
    return results, nil
}
```

### Streaming Large Files

```go
func processLargeFileInChunks(filePath string, chunkSize int) error {
    ctx := createContext()
    
    // Get file info
    listTool := file.FileList()
    info, err := listTool.Execute(ctx, file.FileListParams{
        Path: filepath.Dir(filePath),
        Pattern: filepath.Base(filePath),
}
    if err != nil {
        return err
    }
    
    fileInfo := info.(*file.FileListResult).Files[0]
    totalLines := estimateLineCount(fileInfo.Size)
    
    readTool := file.ReadFile()
    processTool := data.DataTransform()
    
    // Process in chunks
    for start := 1; start <= totalLines; start += chunkSize {
        end := start + chunkSize - 1
        if end > totalLines {
            end = totalLines
        }
        
        // Read chunk
        chunk, err := readTool.Execute(ctx, map[string]interface{}{
            Path: filePath,
            LineStart: start,
            LineEnd: end,
}
        if err != nil {
            continue
        }
        
        // Process chunk
        _, err = processTool.Execute(ctx, data.DataTransformInput{
            Data: chunk.(*file.ReadFileResult).Content,
            Transform: "custom",
            Options: map[string]interface{}{
                "operation": "aggregate",
            },
}
    }
    
    return nil
}
```

## Troubleshooting

### Common Issues and Solutions

```go
// Issue: Permission denied
func handlePermissionDenied(tool domain.Tool, err error) error {
    if strings.Contains(err.Error(), "permission denied") {
        guidance := tool.ErrorGuidance()["permission denied"]
        
        // Try alternative approach
        ctx := createContext()
        ctx.State.Set("use_sudo", true)
        
        // Or try different location
        tempPath := filepath.Join(os.TempDir(), "fallback")
        // Retry with temp path...
    }
    return err
}

// Issue: Network timeouts
func handleNetworkTimeout(ctx *domain.ToolContext) {
    // Increase timeout in state
    ctx.State.Set("default_timeout", 60)
    
    // Enable retry mechanism
    ctx.State.Set("enable_retry", true)
    ctx.State.Set("max_retries", 3)
}

// Issue: Memory constraints
func handleMemoryConstraints(ctx *domain.ToolContext) {
    // Set lower limits
    ctx.State.Set("file_read_max_size", int64(10*1024*1024)) // 10MB
    ctx.State.Set("process_chunk_size", 1000)
    
    // Enable streaming mode
    ctx.State.Set("enable_streaming", true)
}
```

### Debugging Tool Execution

```go
func debugToolExecution(tool domain.Tool, ctx *domain.ToolContext, params interface{}) {
    // Enable verbose logging
    ctx.Events.On("*", func(event domain.Event) {
        log.Printf("[%s] %s: %v", event.Type, event.Source, event.Data)
}
    
    // Log tool metadata
    log.Printf("Executing tool: %s (v%s)", tool.Name(), tool.Version())
    log.Printf("Category: %s, Tags: %v", tool.Category(), tool.Tags())
    log.Printf("Parameters: %+v", params)
    
    // Time execution
    start := time.Now()
    result, err := tool.Execute(ctx, params)
    duration := time.Since(start)
    
    if err != nil {
        log.Printf("Execution failed after %v: %v", duration, err)
        
        // Check error guidance
        for errType, guidance := range tool.ErrorGuidance() {
            if strings.Contains(err.Error(), errType) {
                log.Printf("Error guidance: %s", guidance)
            }
        }
    } else {
        log.Printf("Execution succeeded in %v", duration)
        log.Printf("Result type: %T", result)
    }
}
```

## Best Practices Summary

1. **Always handle errors**: Check error guidance for recovery strategies
2. **Use appropriate contexts**: Set timeouts and cancellation
3. **Configure via state**: Use state for runtime configuration
4. **Monitor with events**: Set up event listeners for visibility
5. **Type assert safely**: Always check type assertions
6. **Batch operations**: Process multiple items efficiently
7. **Cache when possible**: Reduce redundant operations
8. **Respect rate limits**: Add delays for external services
9. **Clean up resources**: Use defer for cleanup
10. **Test edge cases**: Handle empty results and errors gracefully

## Helper Functions

```go
// Common helper functions used in examples

func createContext() *domain.ToolContext {
    return &domain.ToolContext{
        Context: context.Background(),
        State:   domain.NewState(),
        Events:  domain.NewEventEmitter(),
    }
}

func parsePrice(priceStr string) float64 {
    // Remove currency symbols and parse
    cleaned := strings.TrimSpace(priceStr)
    cleaned = strings.ReplaceAll(cleaned, "$", "")
    cleaned = strings.ReplaceAll(cleaned, ",", "")
    
    price, _ := strconv.ParseFloat(cleaned, 64)
    return price
}

func estimateLineCount(fileSize int64) int {
    // Rough estimate: 50 bytes per line average
    return int(fileSize / 50)
}

func processErrorContext(context string) {
    // Custom error processing logic
    lines := strings.Split(context, "\n")
    for _, line := range lines {
        if strings.Contains(line, "ERROR") {
            log.Printf("Error context: %s", line)
        }
    }
}

func notifyError(err error) {
    // Send error notification (email, slack, etc.)
    log.Printf("NOTIFICATION: %v", err)
}

func updateProgressBar(current, total int, message string) {
    percent := float64(current) / float64(total) * 100
    fmt.Printf("\r[%.0f%%] %s", percent, message)
}
```