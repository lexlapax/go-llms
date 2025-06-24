# Custom Tools: Advanced Tool Development

> **[Project Root](/) / [Documentation](../..) / [User Guide](../../user-guide) / [Advanced Topics](../../user-guide/advanced) / Custom Tools**

Learn how to create powerful custom tools for Go-LLMs agents, including implementing the tool interface, handling complex inputs/outputs, supporting async operations, and integrating with external services.

## Tool Architecture Overview

Tools in Go-LLMs follow a simple but powerful interface:

```go
// Core tool interface
type Tool interface {
    // Metadata
    Name() string
    Description() string
    
    // Schema definition
    InputSchema() interface{}
    OutputSchema() interface{}
    
    // Execution
    Execute(ctx context.Context, input interface{}) (interface{}, error)
}

// Extended interfaces for advanced features
type AsyncTool interface {
    Tool
    ExecuteAsync(ctx context.Context, input interface{}) (<-chan ToolResult, error)
}

type StreamingTool interface {
    Tool
    ExecuteStream(ctx context.Context, input interface{}) (<-chan StreamChunk, error)
}

type StatefulTool interface {
    Tool
    Initialize(ctx context.Context, config map[string]interface{}) error
    Cleanup(ctx context.Context) error
}
```

---

## Basic Tool Implementation

### Step 1: Define Tool Structure

```go
package customtools

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/schema"
)

// WeatherTool fetches weather information for a location
type WeatherTool struct {
    apiKey     string
    httpClient *http.Client
    cache      *Cache
    rateLimit  *RateLimiter
}

// NewWeatherTool creates a new weather tool instance
func NewWeatherTool(apiKey string, opts ...WeatherOption) *WeatherTool {
    tool := &WeatherTool{
        apiKey: apiKey,
        httpClient: &http.Client{
            Timeout: 10 * time.Second,
        },
        cache:     NewCache(15 * time.Minute),
        rateLimit: NewRateLimiter(60, time.Minute), // 60 requests per minute
    }
    
    // Apply options
    for _, opt := range opts {
        opt(tool)
    }
    
    return tool
}

// WeatherOption configures the weather tool
type WeatherOption func(*WeatherTool)

func WithHTTPClient(client *http.Client) WeatherOption {
    return func(t *WeatherTool) {
        t.httpClient = client
    }
}

func WithCacheDuration(duration time.Duration) WeatherOption {
    return func(t *WeatherTool) {
        t.cache = NewCache(duration)
    }
}
```

### Step 2: Implement Metadata Methods

```go
// Name returns the tool name
func (w *WeatherTool) Name() string {
    return "weather_get"
}

// Description returns a detailed description for the LLM
func (w *WeatherTool) Description() string {
    return `Fetches current weather information for a specified location. 
    Supports city names, coordinates, and zip codes. Returns temperature, 
    conditions, humidity, wind speed, and forecast.`
}
```

### Step 3: Define Input/Output Schemas

```go
// InputSchema defines the expected input structure
func (w *WeatherTool) InputSchema() interface{} {
    return schema.Object{
        Properties: map[string]schema.Schema{
            "location": schema.String{
                Description: "Location to get weather for (city name, coordinates, or zip code)",
                MinLength:   1,
                MaxLength:   100,
                Examples:    []string{"London", "40.7128,-74.0060", "10001"},
            },
            "units": schema.String{
                Description: "Temperature units",
                Enum:        []string{"celsius", "fahrenheit", "kelvin"},
                Default:     "celsius",
            },
            "include_forecast": schema.Boolean{
                Description: "Include 5-day forecast",
                Default:     false,
            },
        },
        Required: []string{"location"},
    }
}

// OutputSchema defines the output structure
func (w *WeatherTool) OutputSchema() interface{} {
    return schema.Object{
        Properties: map[string]schema.Schema{
            "location": schema.String{
                Description: "Resolved location name",
            },
            "current": schema.Object{
                Properties: map[string]schema.Schema{
                    "temperature": schema.Number{
                        Description: "Current temperature",
                    },
                    "feels_like": schema.Number{
                        Description: "Feels like temperature",
                    },
                    "conditions": schema.String{
                        Description: "Weather conditions",
                    },
                    "humidity": schema.Integer{
                        Description: "Humidity percentage",
                        Minimum:     0,
                        Maximum:     100,
                    },
                    "wind_speed": schema.Number{
                        Description: "Wind speed",
                    },
                    "wind_direction": schema.String{
                        Description: "Wind direction",
                    },
                },
            },
            "forecast": schema.Array{
                Items: schema.Object{
                    Properties: map[string]schema.Schema{
                        "date": schema.String{
                            Format: "date",
                        },
                        "high": schema.Number{},
                        "low": schema.Number{},
                        "conditions": schema.String{},
                        "precipitation_chance": schema.Integer{
                            Minimum: 0,
                            Maximum: 100,
                        },
                    },
                },
            },
            "units": schema.String{},
            "timestamp": schema.String{
                Format: "date-time",
            },
        },
    }
}
```

### Step 4: Implement Execute Method

```go
// Execute performs the weather lookup
func (w *WeatherTool) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // Parse input
    params, err := w.parseInput(input)
    if err != nil {
        return nil, fmt.Errorf("invalid input: %w", err)
    }
    
    // Check cache
    cacheKey := w.getCacheKey(params)
    if cached, found := w.cache.Get(cacheKey); found {
        return cached, nil
    }
    
    // Apply rate limiting
    if err := w.rateLimit.Wait(ctx); err != nil {
        return nil, fmt.Errorf("rate limit exceeded: %w", err)
    }
    
    // Fetch weather data
    weather, err := w.fetchWeather(ctx, params)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch weather: %w", err)
    }
    
    // Cache result
    w.cache.Set(cacheKey, weather)
    
    return weather, nil
}

type weatherParams struct {
    Location        string
    Units           string
    IncludeForecast bool
}

func (w *WeatherTool) parseInput(input interface{}) (*weatherParams, error) {
    // Handle different input types
    var params weatherParams
    
    switch v := input.(type) {
    case map[string]interface{}:
        // JSON object input
        if location, ok := v["location"].(string); ok {
            params.Location = location
        } else {
            return nil, fmt.Errorf("location is required")
        }
        
        if units, ok := v["units"].(string); ok {
            params.Units = units
        } else {
            params.Units = "celsius"
        }
        
        if forecast, ok := v["include_forecast"].(bool); ok {
            params.IncludeForecast = forecast
        }
        
    case string:
        // Simple string input
        params.Location = v
        params.Units = "celsius"
        
    default:
        // Try to unmarshal from JSON
        data, err := json.Marshal(input)
        if err != nil {
            return nil, fmt.Errorf("unsupported input type: %T", input)
        }
        
        if err := json.Unmarshal(data, &params); err != nil {
            return nil, err
        }
    }
    
    // Validate units
    switch params.Units {
    case "celsius", "fahrenheit", "kelvin":
        // Valid
    default:
        return nil, fmt.Errorf("invalid units: %s", params.Units)
    }
    
    return &params, nil
}

func (w *WeatherTool) fetchWeather(ctx context.Context, params *weatherParams) (map[string]interface{}, error) {
    // Build API request
    url := fmt.Sprintf("https://api.weather.com/v1/current?location=%s&units=%s&apikey=%s",
        url.QueryEscape(params.Location),
        params.Units,
        w.apiKey,
    )
    
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    // Make request
    resp, err := w.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API error: %s", resp.Status)
    }
    
    // Parse response
    var apiResp weatherAPIResponse
    if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
        return nil, err
    }
    
    // Build result
    result := map[string]interface{}{
        "location": apiResp.Location.Name,
        "current": map[string]interface{}{
            "temperature":    apiResp.Current.Temperature,
            "feels_like":     apiResp.Current.FeelsLike,
            "conditions":     apiResp.Current.Conditions,
            "humidity":       apiResp.Current.Humidity,
            "wind_speed":     apiResp.Current.WindSpeed,
            "wind_direction": apiResp.Current.WindDirection,
        },
        "units":     params.Units,
        "timestamp": time.Now().Format(time.RFC3339),
    }
    
    // Add forecast if requested
    if params.IncludeForecast {
        forecast, err := w.fetchForecast(ctx, params.Location, params.Units)
        if err != nil {
            // Log error but don't fail the whole request
            log.Printf("Failed to fetch forecast: %v", err)
        } else {
            result["forecast"] = forecast
        }
    }
    
    return result, nil
}
```

---

## Advanced Tool Features

### Async Tool Implementation

```go
// DatabaseQueryTool performs long-running database queries
type DatabaseQueryTool struct {
    db         *sql.DB
    maxWorkers int
    timeout    time.Duration
}

// ExecuteAsync performs queries asynchronously
func (d *DatabaseQueryTool) ExecuteAsync(ctx context.Context, input interface{}) (<-chan domain.ToolResult, error) {
    params, err := d.parseInput(input)
    if err != nil {
        return nil, err
    }
    
    // Create result channel
    results := make(chan domain.ToolResult, 10)
    
    // Start async execution
    go func() {
        defer close(results)
        
        // Send progress updates
        results <- domain.ToolResult{
            Progress: &domain.Progress{
                Status:  "starting",
                Percent: 0,
            },
        }
        
        // Execute query with timeout
        queryCtx, cancel := context.WithTimeout(ctx, d.timeout)
        defer cancel()
        
        rows, err := d.db.QueryContext(queryCtx, params.Query, params.Args...)
        if err != nil {
            results <- domain.ToolResult{
                Error: fmt.Errorf("query failed: %w", err),
            }
            return
        }
        defer rows.Close()
        
        // Process results
        var rowCount int
        for rows.Next() {
            rowCount++
            
            // Parse row
            row, err := d.scanRow(rows)
            if err != nil {
                results <- domain.ToolResult{
                    Error: fmt.Errorf("row scan failed: %w", err),
                }
                return
            }
            
            // Send result
            results <- domain.ToolResult{
                Data: row,
                Progress: &domain.Progress{
                    Status:  "processing",
                    Percent: min(rowCount*10, 90), // Estimate progress
                },
            }
            
            // Check context
            select {
            case <-ctx.Done():
                results <- domain.ToolResult{
                    Error: ctx.Err(),
                }
                return
            default:
            }
        }
        
        // Final result
        results <- domain.ToolResult{
            Data: map[string]interface{}{
                "row_count": rowCount,
                "completed": true,
            },
            Progress: &domain.Progress{
                Status:  "completed",
                Percent: 100,
            },
        }
    }()
    
    return results, nil
}
```

### Streaming Tool Implementation

```go
// LogStreamTool streams log files in real-time
type LogStreamTool struct {
    allowedPaths []string
    bufferSize   int
}

// ExecuteStream streams log content
func (l *LogStreamTool) ExecuteStream(ctx context.Context, input interface{}) (<-chan domain.StreamChunk, error) {
    params, err := l.parseInput(input)
    if err != nil {
        return nil, err
    }
    
    // Validate path
    if !l.isAllowedPath(params.FilePath) {
        return nil, fmt.Errorf("access denied: %s", params.FilePath)
    }
    
    // Open file
    file, err := os.Open(params.FilePath)
    if err != nil {
        return nil, err
    }
    
    // Create stream channel
    stream := make(chan domain.StreamChunk, l.bufferSize)
    
    // Start streaming
    go func() {
        defer close(stream)
        defer file.Close()
        
        // Seek to position if specified
        if params.Follow && params.StartPosition == "end" {
            file.Seek(0, io.SeekEnd)
        }
        
        reader := bufio.NewReader(file)
        
        for {
            select {
            case <-ctx.Done():
                stream <- domain.StreamChunk{
                    Error: ctx.Err(),
                }
                return
                
            default:
                line, err := reader.ReadString('\n')
                if err != nil {
                    if err == io.EOF {
                        if params.Follow {
                            // Wait for new data
                            time.Sleep(100 * time.Millisecond)
                            continue
                        }
                        return
                    }
                    
                    stream <- domain.StreamChunk{
                        Error: err,
                    }
                    return
                }
                
                // Apply filters
                if l.matchesFilter(line, params.Filter) {
                    stream <- domain.StreamChunk{
                        Data: map[string]interface{}{
                            "line":      line,
                            "timestamp": time.Now(),
                        },
                    }
                }
            }
        }
    }()
    
    return stream, nil
}
```

### Stateful Tool Implementation

```go
// MLModelTool maintains model state across calls
type MLModelTool struct {
    modelPath   string
    model       *MLModel
    session     *ModelSession
    initialized bool
    mu          sync.RWMutex
}

// Initialize loads the model
func (m *MLModelTool) Initialize(ctx context.Context, config map[string]interface{}) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    if m.initialized {
        return nil
    }
    
    // Load model configuration
    if path, ok := config["model_path"].(string); ok {
        m.modelPath = path
    } else {
        return fmt.Errorf("model_path is required")
    }
    
    // Load model
    model, err := LoadMLModel(m.modelPath)
    if err != nil {
        return fmt.Errorf("failed to load model: %w", err)
    }
    
    // Create session
    session, err := model.CreateSession()
    if err != nil {
        return fmt.Errorf("failed to create session: %w", err)
    }
    
    m.model = model
    m.session = session
    m.initialized = true
    
    return nil
}

// Execute runs inference
func (m *MLModelTool) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    m.mu.RLock()
    if !m.initialized {
        m.mu.RUnlock()
        return nil, fmt.Errorf("tool not initialized")
    }
    session := m.session
    m.mu.RUnlock()
    
    // Parse input
    params, err := m.parseInput(input)
    if err != nil {
        return nil, err
    }
    
    // Prepare input tensor
    tensor, err := m.prepareTensor(params.Data)
    if err != nil {
        return nil, fmt.Errorf("failed to prepare tensor: %w", err)
    }
    
    // Run inference
    output, err := session.Run(ctx, tensor)
    if err != nil {
        return nil, fmt.Errorf("inference failed: %w", err)
    }
    
    // Post-process output
    result := m.postProcess(output, params.OutputFormat)
    
    return result, nil
}

// Cleanup releases resources
func (m *MLModelTool) Cleanup(ctx context.Context) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    if !m.initialized {
        return nil
    }
    
    // Close session
    if m.session != nil {
        if err := m.session.Close(); err != nil {
            return err
        }
    }
    
    // Unload model
    if m.model != nil {
        if err := m.model.Unload(); err != nil {
            return err
        }
    }
    
    m.initialized = false
    
    return nil
}
```

---

## Complex Tool Patterns

### Multi-Step Tool

```go
// DataPipelineTool executes multi-step data processing
type DataPipelineTool struct {
    steps []PipelineStep
}

type PipelineStep interface {
    Name() string
    Execute(ctx context.Context, data interface{}) (interface{}, error)
    Validate(data interface{}) error
}

func (d *DataPipelineTool) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // Parse pipeline configuration
    config, err := d.parseConfig(input)
    if err != nil {
        return nil, err
    }
    
    // Execute pipeline
    data := config.InitialData
    results := make([]StepResult, 0, len(d.steps))
    
    for i, step := range d.steps {
        // Validate input
        if err := step.Validate(data); err != nil {
            return nil, fmt.Errorf("step %d (%s) validation failed: %w", i, step.Name(), err)
        }
        
        // Execute step
        startTime := time.Now()
        output, err := step.Execute(ctx, data)
        duration := time.Since(startTime)
        
        // Record result
        result := StepResult{
            StepName: step.Name(),
            Index:    i,
            Duration: duration,
            Success:  err == nil,
        }
        
        if err != nil {
            result.Error = err.Error()
            results = append(results, result)
            
            // Handle error based on config
            if config.StopOnError {
                return map[string]interface{}{
                    "partial_results": results,
                    "error":           err.Error(),
                    "failed_at_step":  i,
                }, nil
            }
            
            // Continue with previous data
            continue
        }
        
        result.OutputSample = d.sampleData(output)
        results = append(results, result)
        
        // Use output as input for next step
        data = output
    }
    
    return map[string]interface{}{
        "final_output": data,
        "pipeline_results": results,
        "total_duration": d.sumDurations(results),
    }, nil
}
```

### Tool with External Service Integration

```go
// NotificationTool sends notifications through multiple channels
type NotificationTool struct {
    providers map[string]NotificationProvider
    fallbacks map[string][]string
}

type NotificationProvider interface {
    Send(ctx context.Context, notification Notification) error
    Available() bool
}

func (n *NotificationTool) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // Parse notification request
    req, err := n.parseRequest(input)
    if err != nil {
        return nil, err
    }
    
    // Validate required fields
    if err := n.validateRequest(req); err != nil {
        return nil, err
    }
    
    // Build notification
    notification := Notification{
        ID:        generateID(),
        Channel:   req.Channel,
        Recipient: req.Recipient,
        Subject:   req.Subject,
        Body:      req.Body,
        Priority:  req.Priority,
        Metadata:  req.Metadata,
        CreatedAt: time.Now(),
    }
    
    // Send through primary channel
    provider, exists := n.providers[req.Channel]
    if !exists {
        return nil, fmt.Errorf("unknown channel: %s", req.Channel)
    }
    
    err = n.sendWithFallback(ctx, provider, notification, req.Channel)
    
    // Build response
    response := map[string]interface{}{
        "notification_id": notification.ID,
        "status":          "sent",
        "channel":         req.Channel,
        "timestamp":       notification.CreatedAt,
    }
    
    if err != nil {
        response["status"] = "failed"
        response["error"] = err.Error()
        
        // Try fallback channels
        if fallbacks, ok := n.fallbacks[req.Channel]; ok {
            for _, fallbackChannel := range fallbacks {
                if fallbackProvider, exists := n.providers[fallbackChannel]; exists {
                    if err := fallbackProvider.Send(ctx, notification); err == nil {
                        response["status"] = "sent_via_fallback"
                        response["fallback_channel"] = fallbackChannel
                        break
                    }
                }
            }
        }
    }
    
    return response, nil
}

func (n *NotificationTool) sendWithFallback(ctx context.Context, provider NotificationProvider, notification Notification, channel string) error {
    // Check availability
    if !provider.Available() {
        return fmt.Errorf("provider %s is not available", channel)
    }
    
    // Send with timeout
    sendCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()
    
    // Send notification
    errChan := make(chan error, 1)
    go func() {
        errChan <- provider.Send(sendCtx, notification)
    }()
    
    select {
    case err := <-errChan:
        return err
    case <-sendCtx.Done():
        return fmt.Errorf("send timeout: %w", sendCtx.Err())
    }
}

// Email provider implementation
type EmailProvider struct {
    smtp     *SMTPClient
    from     string
    template *template.Template
}

func (e *EmailProvider) Send(ctx context.Context, notification Notification) error {
    // Render email body
    var body bytes.Buffer
    if err := e.template.Execute(&body, notification); err != nil {
        return fmt.Errorf("template rendering failed: %w", err)
    }
    
    // Build email
    email := Email{
        From:    e.from,
        To:      []string{notification.Recipient},
        Subject: notification.Subject,
        Body:    body.String(),
        HTML:    notification.Metadata["html"] == "true",
    }
    
    // Add attachments if any
    if attachments, ok := notification.Metadata["attachments"].([]string); ok {
        email.Attachments = attachments
    }
    
    // Send email
    return e.smtp.Send(ctx, email)
}

func (e *EmailProvider) Available() bool {
    return e.smtp.Ping() == nil
}
```

### Tool with Caching and Optimization

```go
// GeocodingTool with intelligent caching
type GeocodingTool struct {
    client       *GeocodingClient
    cache        *GeocodeCache
    rateLimiter  *RateLimiter
    metrics      *ToolMetrics
}

type GeocodeCache struct {
    exact    *LRUCache      // Exact address matches
    fuzzy    *FuzzyCache    // Fuzzy matching for similar addresses
    spatial  *SpatialCache  // Nearby locations
    ttl      time.Duration
}

func (g *GeocodingTool) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    startTime := time.Now()
    defer func() {
        g.metrics.RecordExecution(time.Since(startTime))
    }()
    
    // Parse input
    req, err := g.parseRequest(input)
    if err != nil {
        g.metrics.RecordError("parse_error")
        return nil, err
    }
    
    // Try cache layers
    if result := g.checkCache(req); result != nil {
        g.metrics.RecordCacheHit()
        return result, nil
    }
    
    g.metrics.RecordCacheMiss()
    
    // Apply rate limiting
    if err := g.rateLimiter.Wait(ctx); err != nil {
        g.metrics.RecordError("rate_limit")
        return nil, err
    }
    
    // Make API request
    result, err := g.client.Geocode(ctx, req)
    if err != nil {
        g.metrics.RecordError("api_error")
        return nil, err
    }
    
    // Update all cache layers
    g.updateCache(req, result)
    
    return result, nil
}

func (g *GeocodingTool) checkCache(req *GeocodeRequest) *GeocodeResult {
    // Check exact cache
    if result := g.cache.exact.Get(req.Address); result != nil {
        return result.(*GeocodeResult)
    }
    
    // Check fuzzy cache for similar addresses
    if req.AllowFuzzy {
        similar := g.cache.fuzzy.FindSimilar(req.Address, 0.9)
        if similar != nil {
            return similar.(*GeocodeResult)
        }
    }
    
    // Check spatial cache for coordinates
    if req.Coordinates != nil {
        nearby := g.cache.spatial.FindNearby(
            req.Coordinates.Lat,
            req.Coordinates.Lon,
            req.Radius,
        )
        if nearby != nil {
            return nearby.(*GeocodeResult)
        }
    }
    
    return nil
}

func (g *GeocodingTool) updateCache(req *GeocodeRequest, result *GeocodeResult) {
    // Update exact cache
    g.cache.exact.Set(req.Address, result, g.cache.ttl)
    
    // Update fuzzy cache
    g.cache.fuzzy.Add(req.Address, result)
    
    // Update spatial cache
    if result.Coordinates != nil {
        g.cache.spatial.Add(
            result.Coordinates.Lat,
            result.Coordinates.Lon,
            result,
        )
    }
    
    // Pre-cache nearby locations if requested
    if req.PrecacheNearby {
        go g.precacheNearbyLocations(result.Coordinates)
    }
}
```

---

## Tool Testing and Validation

### Unit Testing Tools

```go
package customtools_test

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestWeatherTool_Execute(t *testing.T) {
    // Create mock HTTP client
    mockClient := NewMockHTTPClient()
    mockClient.SetResponse("https://api.weather.com/v1/current", MockResponse{
        Status: 200,
        Body: `{
            "location": {"name": "London"},
            "current": {
                "temperature": 20,
                "conditions": "Partly cloudy",
                "humidity": 65
            }
        }`,
}
    
    // Create tool with mock
    tool := NewWeatherTool("test-api-key",
        WithHTTPClient(mockClient),
        WithCacheDuration(1*time.Minute),
    )
    
    // Test execution
    ctx := context.Background()
    result, err := tool.Execute(ctx, map[string]interface{}{
        "location": "London",
        "units":    "celsius",
}
    
    require.NoError(t, err)
    
    // Verify result
    data, ok := result.(map[string]interface{})
    require.True(t, ok)
    
    assert.Equal(t, "London", data["location"])
    assert.Equal(t, "celsius", data["units"])
    
    current, ok := data["current"].(map[string]interface{})
    require.True(t, ok)
    assert.Equal(t, 20.0, current["temperature"])
    assert.Equal(t, "Partly cloudy", current["conditions"])
}

func TestWeatherTool_Caching(t *testing.T) {
    mockClient := NewMockHTTPClient()
    tool := NewWeatherTool("test-api-key",
        WithHTTPClient(mockClient),
        WithCacheDuration(1*time.Hour),
    )
    
    ctx := context.Background()
    input := map[string]interface{}{"location": "Paris"}
    
    // First call - should hit API
    mockClient.SetResponse("https://api.weather.com/v1/current", MockResponse{
        Status: 200,
        Body:   `{"location": {"name": "Paris"}, "current": {"temperature": 15}}`,
}
    
    result1, err := tool.Execute(ctx, input)
    require.NoError(t, err)
    assert.Equal(t, 1, mockClient.CallCount())
    
    // Second call - should use cache
    result2, err := tool.Execute(ctx, input)
    require.NoError(t, err)
    assert.Equal(t, 1, mockClient.CallCount()) // No additional calls
    
    // Results should be identical
    assert.Equal(t, result1, result2)
}

func TestWeatherTool_RateLimiting(t *testing.T) {
    mockClient := NewMockHTTPClient()
    tool := NewWeatherTool("test-api-key",
        WithHTTPClient(mockClient),
        WithRateLimit(2, time.Second), // 2 requests per second
    )
    
    ctx := context.Background()
    
    // Make 3 rapid requests
    start := time.Now()
    
    for i := 0; i < 3; i++ {
        _, err := tool.Execute(ctx, map[string]interface{}{
            "location": fmt.Sprintf("City%d", i),
}
        require.NoError(t, err)
    }
    
    elapsed := time.Since(start)
    
    // Third request should have been delayed
    assert.True(t, elapsed >= 500*time.Millisecond)
}
```

### Integration Testing

```go
func TestWeatherTool_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    // Use real API key from environment
    apiKey := os.Getenv("WEATHER_API_KEY")
    if apiKey == "" {
        t.Skip("WEATHER_API_KEY not set")
    }
    
    tool := NewWeatherTool(apiKey)
    ctx := context.Background()
    
    // Test various input formats
    testCases := []struct {
        name     string
        input    interface{}
        validate func(t *testing.T, result interface{})
    }{
        {
            name: "City name",
            input: map[string]interface{}{
                "location": "Tokyo",
                "units":    "celsius",
            },
            validate: func(t *testing.T, result interface{}) {
                data := result.(map[string]interface{})
                assert.Contains(t, data["location"], "Tokyo")
            },
        },
        {
            name: "Coordinates",
            input: map[string]interface{}{
                "location": "35.6762,139.6503",
                "include_forecast": true,
            },
            validate: func(t *testing.T, result interface{}) {
                data := result.(map[string]interface{})
                assert.NotNil(t, data["forecast"])
            },
        },
        {
            name: "Simple string",
            input: "New York",
            validate: func(t *testing.T, result interface{}) {
                data := result.(map[string]interface{})
                assert.Contains(t, data["location"], "New York")
            },
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result, err := tool.Execute(ctx, tc.input)
            require.NoError(t, err)
            tc.validate(t, result)
}
    }
}
```

### Schema Validation Testing

```go
func TestWeatherTool_SchemaValidation(t *testing.T) {
    tool := NewWeatherTool("test-key")
    
    // Get schemas
    inputSchema := tool.InputSchema()
    outputSchema := tool.OutputSchema()
    
    // Validate schemas are valid JSON Schema
    inputJSON, err := json.Marshal(inputSchema)
    require.NoError(t, err)
    
    var schemaDoc map[string]interface{}
    err = json.Unmarshal(inputJSON, &schemaDoc)
    require.NoError(t, err)
    
    // Test input validation
    validator := NewSchemaValidator(inputSchema)
    
    // Valid input
    err = validator.Validate(map[string]interface{}{
        "location": "London",
        "units":    "celsius",
}
    assert.NoError(t, err)
    
    // Invalid input - missing required field
    err = validator.Validate(map[string]interface{}{
        "units": "celsius",
}
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "location")
    
    // Invalid input - wrong type
    err = validator.Validate(map[string]interface{}{
        "location": 123,
        "units":    "celsius",
}
    assert.Error(t, err)
    
    // Invalid input - invalid enum value
    err = validator.Validate(map[string]interface{}{
        "location": "London",
        "units":    "invalid",
}
    assert.Error(t, err)
}
```

---

## Tool Discovery and Registration

### Static Registration

```go
// Register tool at initialization
func init() {
    domain.RegisterTool("weather_get", func(config map[string]interface{}) (domain.Tool, error) {
        apiKey, ok := config["api_key"].(string)
        if !ok {
            apiKey = os.Getenv("WEATHER_API_KEY")
        }
        
        if apiKey == "" {
            return nil, errors.New("weather API key required")
        }
        
        var opts []WeatherOption
        
        if cacheMinutes, ok := config["cache_duration_minutes"].(float64); ok {
            opts = append(opts, WithCacheDuration(time.Duration(cacheMinutes)*time.Minute))
        }
        
        if rateLimit, ok := config["rate_limit"].(float64); ok {
            opts = append(opts, WithRateLimit(int(rateLimit), time.Minute))
        }
        
        return NewWeatherTool(apiKey, opts...), nil
}
}
```

### Dynamic Discovery

```go
// DiscoverableTools provides tool discovery
type DiscoverableTools struct {
    tools    map[string]domain.Tool
    metadata map[string]ToolMetadata
    mu       sync.RWMutex
}

type ToolMetadata struct {
    Name         string
    Category     string
    Description  string
    Version      string
    Author       string
    Tags         []string
    RequiredAuth []string
    Examples     []ToolExample
}

type ToolExample struct {
    Description string
    Input       interface{}
    Output      interface{}
}

func (dt *DiscoverableTools) Register(tool domain.Tool, metadata ToolMetadata) error {
    dt.mu.Lock()
    defer dt.mu.Unlock()
    
    name := tool.Name()
    if _, exists := dt.tools[name]; exists {
        return fmt.Errorf("tool %s already registered", name)
    }
    
    dt.tools[name] = tool
    dt.metadata[name] = metadata
    
    return nil
}

func (dt *DiscoverableTools) Discover(filters ...FilterFunc) []ToolInfo {
    dt.mu.RLock()
    defer dt.mu.RUnlock()
    
    var results []ToolInfo
    
    for name, tool := range dt.tools {
        meta := dt.metadata[name]
        info := ToolInfo{
            Tool:     tool,
            Metadata: meta,
        }
        
        // Apply filters
        include := true
        for _, filter := range filters {
            if !filter(info) {
                include = false
                break
            }
        }
        
        if include {
            results = append(results, info)
        }
    }
    
    return results
}

// Filter functions
func WithCategory(category string) FilterFunc {
    return func(info ToolInfo) bool {
        return info.Metadata.Category == category
    }
}

func WithTags(tags ...string) FilterFunc {
    return func(info ToolInfo) bool {
        for _, tag := range tags {
            found := false
            for _, toolTag := range info.Metadata.Tags {
                if toolTag == tag {
                    found = true
                    break
                }
            }
            if !found {
                return false
            }
        }
        return true
    }
}

// Usage
registry := NewDiscoverableTools()

// Register tools
registry.Register(weatherTool, ToolMetadata{
    Name:        "weather_get",
    Category:    "external_api",
    Description: "Fetches weather information",
    Version:     "1.0.0",
    Tags:        []string{"weather", "api", "location"},
    RequiredAuth: []string{"WEATHER_API_KEY"},
    Examples: []ToolExample{
        {
            Description: "Get weather for a city",
            Input:       map[string]interface{}{"location": "London"},
            Output:      map[string]interface{}{"temperature": 20, "conditions": "Cloudy"},
        },
    },
}

// Discover tools
apiTools := registry.Discover(WithCategory("external_api"))
weatherTools := registry.Discover(WithTags("weather"))
```

---

## Tool Composition and Chaining

### Tool Pipeline

```go
// ToolPipeline chains multiple tools
type ToolPipeline struct {
    steps []PipelineStep
}

type PipelineStep struct {
    Tool      domain.Tool
    Transform func(interface{}) (interface{}, error)
    OnError   ErrorHandler
}

func (tp *ToolPipeline) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    data := input
    
    for i, step := range tp.steps {
        // Transform input if needed
        if step.Transform != nil {
            transformed, err := step.Transform(data)
            if err != nil {
                return nil, fmt.Errorf("transform failed at step %d: %w", i, err)
            }
            data = transformed
        }
        
        // Execute tool
        result, err := step.Tool.Execute(ctx, data)
        if err != nil {
            if step.OnError != nil {
                handled, handledErr := step.OnError(err, data)
                if handledErr != nil {
                    return nil, fmt.Errorf("error handler failed at step %d: %w", i, handledErr)
                }
                data = handled
                continue
            }
            return nil, fmt.Errorf("step %d failed: %w", i, err)
        }
        
        data = result
    }
    
    return data, nil
}

// Example: Weather analysis pipeline
weatherPipeline := &ToolPipeline{
    steps: []PipelineStep{
        {
            Tool: weatherTool,
            Transform: func(input interface{}) (interface{}, error) {
                // Ensure location format
                return map[string]interface{}{
                    "location": input,
                    "include_forecast": true,
                }, nil
            },
        },
        {
            Tool: sentimentTool,
            Transform: func(input interface{}) (interface{}, error) {
                // Extract weather description
                weather := input.(map[string]interface{})
                current := weather["current"].(map[string]interface{})
                return map[string]interface{}{
                    "text": fmt.Sprintf("The weather is %s with %v degrees",
                        current["conditions"], current["temperature"]),
                }, nil
            },
        },
        {
            Tool: notificationTool,
            Transform: func(input interface{}) (interface{}, error) {
                // Create notification based on sentiment
                sentiment := input.(map[string]interface{})
                mood := sentiment["sentiment"].(string)
                
                message := "Have a great day!"
                if mood == "negative" {
                    message = "Stay cozy indoors!"
                }
                
                return map[string]interface{}{
                    "channel":   "email",
                    "recipient": "user@example.com",
                    "subject":   "Weather Update",
                    "body":      message,
                }, nil
            },
        },
    },
}
```

### Conditional Tool Execution

```go
// ConditionalTool executes different tools based on conditions
type ConditionalTool struct {
    conditions []Condition
    fallback   domain.Tool
}

type Condition struct {
    Predicate func(interface{}) bool
    Tool      domain.Tool
}

func (ct *ConditionalTool) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // Check conditions in order
    for _, condition := range ct.conditions {
        if condition.Predicate(input) {
            return condition.Tool.Execute(ctx, input)
        }
    }
    
    // Use fallback if no conditions match
    if ct.fallback != nil {
        return ct.fallback.Execute(ctx, input)
    }
    
    return nil, errors.New("no matching condition")
}

// Example: Smart location tool
smartLocationTool := &ConditionalTool{
    conditions: []Condition{
        {
            Predicate: func(input interface{}) bool {
                // Check if input is coordinates
                str, ok := input.(string)
                return ok && regexp.MustCompile(`^-?\d+\.\d+,-?\d+\.\d+$`).MatchString(str)
            },
            Tool: reverseGeocodeTool,
        },
        {
            Predicate: func(input interface{}) bool {
                // Check if input is address
                str, ok := input.(string)
                return ok && strings.Contains(str, ",")
            },
            Tool: geocodeTool,
        },
        {
            Predicate: func(input interface{}) bool {
                // Check if input is IP address
                str, ok := input.(string)
                return ok && net.ParseIP(str) != nil
            },
            Tool: ipLocationTool,
        },
    },
    fallback: searchLocationTool,
}
```

---

## Best Practices

### 1. Tool Design
- Keep tools focused on a single responsibility
- Use descriptive names and comprehensive descriptions
- Define clear input/output schemas
- Handle errors gracefully with meaningful messages

### 2. Performance
- Implement caching for expensive operations
- Use connection pooling for external services
- Add appropriate timeouts and cancellation
- Consider async execution for long operations

### 3. Security
- Validate all inputs thoroughly
- Sanitize outputs to prevent injection
- Use secure connections (TLS) for external services
- Never expose sensitive data in responses

### 4. Reliability
- Implement retry logic with exponential backoff
- Add circuit breakers for external services
- Provide fallback mechanisms
- Monitor tool health and performance

### 5. Testing
- Write comprehensive unit tests
- Include integration tests for external services
- Test error scenarios and edge cases
- Validate schemas match implementation

### 6. Documentation
- Document all configuration options
- Provide clear usage examples
- Explain error codes and recovery
- Include troubleshooting guide

---

## Tool Configuration Template

```yaml
# tool-config.yaml
tools:
  weather_get:
    enabled: true
    api_key: ${WEATHER_API_KEY}
    cache_duration_minutes: 15
    rate_limit: 60
    timeout_seconds: 10
    retry_attempts: 3
    
  database_query:
    enabled: true
    connection_string: ${DATABASE_URL}
    max_workers: 10
    query_timeout_seconds: 30
    result_limit: 1000
    
  notification_send:
    enabled: true
    providers:
      email:
        smtp_host: smtp.gmail.com
        smtp_port: 587
        from_address: noreply@example.com
        
      slack:
        webhook_url: ${SLACK_WEBHOOK}
        default_channel: "#notifications"
        
      sms:
        api_key: ${TWILIO_API_KEY}
        from_number: "+1234567890"
    
    fallback_chain:
      email: [slack]
      slack: [email]
      sms: [email, slack]
```

---

## Next Steps

- **[Workflow Orchestration](workflow-orchestration.md)** - Combine tools in complex workflows
- **[Built-in Tools Reference](../../user-guide/reference/built-in-tools-reference.md)** - Explore existing tools
- **[Agent Tools Guide](../../user-guide/guides/agent-tools.md)** - Using tools with agents
- **[Tool Discovery](../../technical/tools/tool-discovery.md)** - Runtime tool registration
- **[API Reference](../../technical/api-reference/tools.md)** - Tool interface documentation