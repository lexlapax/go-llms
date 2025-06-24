# Built-in Tools: Available Tools and Examples

> **[Project Root](/) / [Documentation](../..) / [Technical Documentation](../../technical) / [Tools](../../technical/tools) / Built-in Tools**

Comprehensive reference for all built-in tools available in Go-LLMs, including detailed documentation, usage examples, configuration options, and integration patterns for each tool category in the standard tool library.

## Tool Categories Overview

Go-LLMs includes 30+ built-in tools organized into logical categories:

| Category | Tools | Description |
|----------|-------|-------------|
| [File System](#file-system-tools) | 10 tools | File operations, directory management, search |
| [Web & HTTP](#web-http-tools) | 8 tools | HTTP requests, web scraping, API clients |
| [System](#system-tools) | 7 tools | Process management, environment, system info |
| [Data Processing](#data-processing-tools) | 10 tools | JSON, XML, CSV, template rendering, validation |
| [Math & Computation](#math-computation-tools) | 5 tools | Mathematical operations, statistical analysis |
| [Date & Time](#date-time-tools) | 4 tools | Date parsing, formatting, timezone operations |
| [Text Processing](#text-processing-tools) | 6 tools | String manipulation, regex, encoding |

## File System Tools

### ReadFile Tool

Reads file contents with encoding support and size limits.

```go
// ReadFileInput defines the input schema
type ReadFileInput struct {
    Path     string `json:"path" validate:"required"`
    Encoding string `json:"encoding,omitempty" default:"utf-8"`
    MaxSize  int64  `json:"max_size,omitempty" default:"10485760"` // 10MB
    Offset   int64  `json:"offset,omitempty"`
    Length   int64  `json:"length,omitempty"`
}

// ReadFileOutput defines the output schema
type ReadFileOutput struct {
    Content   string            `json:"content"`
    Size      int64             `json:"size"`
    Encoding  string            `json:"encoding"`
    MimeType  string            `json:"mime_type,omitempty"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Usage example
func ExampleReadFileTool() {
    tool, _ := tools.GetTool("read_file")
    
    result, err := tool.Execute(context.Background(), ReadFileInput{
        Path:     "/path/to/file.txt",
        Encoding: "utf-8",
        MaxSize:  1048576, // 1MB limit
}
    
    if err != nil {
        log.Fatal(err)
    }
    
    output := result.(*ReadFileOutput)
    fmt.Printf("File content: %s\n", output.Content)
    fmt.Printf("File size: %d bytes\n", output.Size)
}
```

**Configuration Options:**
```yaml
read_file:
  max_size: 10485760  # 10MB default limit
  allowed_paths:
    - "/home/user/documents"
    - "/tmp"
  blocked_paths:
    - "/etc"
    - "/proc"
  default_encoding: "utf-8"
  auto_detect_encoding: true
```

### WriteFile Tool

Writes content to files with atomic operations and backup support.

```go
type WriteFileInput struct {
    Path      string `json:"path" validate:"required"`
    Content   string `json:"content" validate:"required"`
    Encoding  string `json:"encoding,omitempty" default:"utf-8"`
    Mode      string `json:"mode,omitempty" default:"0644"`
    Atomic    bool   `json:"atomic,omitempty" default:"true"`
    Backup    bool   `json:"backup,omitempty" default:"false"`
    CreateDir bool   `json:"create_dir,omitempty" default:"false"`
}

type WriteFileOutput struct {
    BytesWritten int64  `json:"bytes_written"`
    Path         string `json:"path"`
    BackupPath   string `json:"backup_path,omitempty"`
    Checksum     string `json:"checksum"`
}

// Usage example
func ExampleWriteFileTool() {
    tool, _ := tools.GetTool("write_file")
    
    result, err := tool.Execute(context.Background(), WriteFileInput{
        Path:      "/tmp/output.txt",
        Content:   "Hello, World!",
        Atomic:    true,
        Backup:    true,
        CreateDir: true,
}
    
    if err != nil {
        log.Fatal(err)
    }
    
    output := result.(*WriteFileOutput)
    fmt.Printf("Wrote %d bytes to %s\n", output.BytesWritten, output.Path)
}
```

### ListDirectory Tool

Lists directory contents with filtering and metadata.

```go
type ListDirectoryInput struct {
    Path      string   `json:"path" validate:"required"`
    Recursive bool     `json:"recursive,omitempty"`
    Pattern   string   `json:"pattern,omitempty"`
    ShowHidden bool    `json:"show_hidden,omitempty"`
    SortBy    string   `json:"sort_by,omitempty" default:"name"`
    SortOrder string   `json:"sort_order,omitempty" default:"asc"`
    Include   []string `json:"include,omitempty"`
    Exclude   []string `json:"exclude,omitempty"`
}

type FileInfo struct {
    Name     string    `json:"name"`
    Path     string    `json:"path"`
    Size     int64     `json:"size"`
    Mode     string    `json:"mode"`
    ModTime  time.Time `json:"mod_time"`
    IsDir    bool      `json:"is_dir"`
    MimeType string    `json:"mime_type,omitempty"`
}

type ListDirectoryOutput struct {
    Files     []FileInfo `json:"files"`
    Dirs      []FileInfo `json:"dirs"`
    Total     int        `json:"total"`
    TotalSize int64      `json:"total_size"`
}

// Usage example
func ExampleListDirectoryTool() {
    tool, _ := tools.GetTool("list_directory")
    
    result, err := tool.Execute(context.Background(), ListDirectoryInput{
        Path:      "/home/user/documents",
        Recursive: true,
        Pattern:   "*.go",
        SortBy:    "mod_time",
        SortOrder: "desc",
}
    
    if err != nil {
        log.Fatal(err)
    }
    
    output := result.(*ListDirectoryOutput)
    fmt.Printf("Found %d Go files\n", len(output.Files))
    for _, file := range output.Files {
        fmt.Printf("- %s (%d bytes)\n", file.Name, file.Size)
    }
}
```

### SearchFiles Tool

Advanced file search with content and metadata matching.

```go
type SearchFilesInput struct {
    Path        string            `json:"path" validate:"required"`
    Query       string            `json:"query,omitempty"`
    Pattern     string            `json:"pattern,omitempty"`
    ContentMatch string           `json:"content_match,omitempty"`
    MaxDepth    int               `json:"max_depth,omitempty"`
    MaxResults  int               `json:"max_results,omitempty"`
    IncludeContent bool           `json:"include_content,omitempty"`
    CaseSensitive bool            `json:"case_sensitive,omitempty"`
    Filters     map[string]interface{} `json:"filters,omitempty"`
}

type SearchResult struct {
    File     FileInfo `json:"file"`
    Score    float64  `json:"score"`
    Matches  []Match  `json:"matches,omitempty"`
    Content  string   `json:"content,omitempty"`
}

type Match struct {
    Line   int    `json:"line"`
    Column int    `json:"column"`
    Text   string `json:"text"`
    Before string `json:"before,omitempty"`
    After  string `json:"after,omitempty"`
}

// Usage example
func ExampleSearchFilesTool() {
    tool, _ := tools.GetTool("search_files")
    
    result, err := tool.Execute(context.Background(), SearchFilesInput{
        Path:         "/project/src",
        Pattern:      "*.go",
        ContentMatch: "func.*Error",
        MaxResults:   20,
        IncludeContent: true,
        CaseSensitive: false,
}
    
    output := result.(*SearchFilesOutput)
    fmt.Printf("Found %d matching files\n", len(output.Results))
}
```

## Web & HTTP Tools

### HTTPRequest Tool

Comprehensive HTTP client with authentication and advanced features.

```go
type HTTPRequestInput struct {
    URL     string            `json:"url" validate:"required,url"`
    Method  string            `json:"method,omitempty" default:"GET"`
    Headers map[string]string `json:"headers,omitempty"`
    Body    string            `json:"body,omitempty"`
    Params  map[string]string `json:"params,omitempty"`
    Auth    *AuthConfig       `json:"auth,omitempty"`
    Timeout int               `json:"timeout,omitempty" default:"30"`
    Follow  bool              `json:"follow,omitempty" default:"true"`
    Verify  bool              `json:"verify,omitempty" default:"true"`
}

type AuthConfig struct {
    Type     string `json:"type"` // basic, bearer, api_key, oauth
    Username string `json:"username,omitempty"`
    Password string `json:"password,omitempty"`
    Token    string `json:"token,omitempty"`
    Key      string `json:"key,omitempty"`
    Value    string `json:"value,omitempty"`
}

type HTTPRequestOutput struct {
    StatusCode int               `json:"status_code"`
    Headers    map[string]string `json:"headers"`
    Body       string            `json:"body"`
    Size       int64             `json:"size"`
    Duration   time.Duration     `json:"duration"`
    URL        string            `json:"url"`
    Redirects  []string          `json:"redirects,omitempty"`
}

// Usage examples
func ExampleHTTPRequestTool() {
    tool, _ := tools.GetTool("http_request")
    
    // Simple GET request
    result, err := tool.Execute(context.Background(), HTTPRequestInput{
        URL:    "https://api.github.com/user",
        Method: "GET",
        Headers: map[string]string{
            "Accept": "application/vnd.github.v3+json",
        },
        Auth: &AuthConfig{
            Type:  "bearer",
            Token: "github_token_here",
        },
}
    
    // POST request with JSON body
    postResult, err := tool.Execute(context.Background(), HTTPRequestInput{
        URL:    "https://api.example.com/data",
        Method: "POST",
        Headers: map[string]string{
            "Content-Type": "application/json",
        },
        Body: `{"name": "John", "email": "john@example.com"}`,
        Auth: &AuthConfig{
            Type: "api_key",
            Key:  "X-API-Key",
            Value: "your_api_key",
        },
}
    
    output := result.(*HTTPRequestOutput)
    fmt.Printf("Status: %d, Size: %d bytes\n", output.StatusCode, output.Size)
}
```

### WebScrape Tool

Web scraping with CSS selectors and content extraction.

```go
type WebScrapeInput struct {
    URL         string            `json:"url" validate:"required,url"`
    Selectors   map[string]string `json:"selectors,omitempty"`
    WaitFor     string            `json:"wait_for,omitempty"`
    JavaScript  bool              `json:"javascript,omitempty"`
    Screenshots bool              `json:"screenshots,omitempty"`
    UserAgent   string            `json:"user_agent,omitempty"`
    Headers     map[string]string `json:"headers,omitempty"`
    Timeout     int               `json:"timeout,omitempty" default:"30"`
}

type WebScrapeOutput struct {
    Title       string                    `json:"title"`
    URL         string                    `json:"url"`
    Content     string                    `json:"content"`
    Extracted   map[string]interface{}    `json:"extracted"`
    Links       []Link                    `json:"links"`
    Images      []Image                   `json:"images"`
    Metadata    map[string]string         `json:"metadata"`
    Screenshot  string                    `json:"screenshot,omitempty"`
}

type Link struct {
    Text string `json:"text"`
    URL  string `json:"url"`
    Type string `json:"type"`
}

type Image struct {
    Src    string `json:"src"`
    Alt    string `json:"alt"`
    Width  int    `json:"width,omitempty"`
    Height int    `json:"height,omitempty"`
}

// Usage example
func ExampleWebScrapeTool() {
    tool, _ := tools.GetTool("web_scrape")
    
    result, err := tool.Execute(context.Background(), WebScrapeInput{
        URL: "https://news.ycombinator.com",
        Selectors: map[string]string{
            "headlines": ".titleline > a",
            "scores":    ".score",
            "comments":  ".subtext",
        },
        JavaScript: false,
        Timeout:    30,
}
    
    output := result.(*WebScrapeOutput)
    fmt.Printf("Scraped: %s\n", output.Title)
    fmt.Printf("Found %d links\n", len(output.Links))
}
```

### APIClient Tool

REST API client with automatic retry and rate limiting.

```go
type APIClientInput struct {
    BaseURL     string            `json:"base_url" validate:"required,url"`
    Endpoint    string            `json:"endpoint" validate:"required"`
    Method      string            `json:"method,omitempty" default:"GET"`
    Data        interface{}       `json:"data,omitempty"`
    Params      map[string]string `json:"params,omitempty"`
    Headers     map[string]string `json:"headers,omitempty"`
    Auth        *AuthConfig       `json:"auth,omitempty"`
    RetryCount  int               `json:"retry_count,omitempty" default:"3"`
    RetryDelay  int               `json:"retry_delay,omitempty" default:"1"`
    RateLimit   *RateLimitConfig  `json:"rate_limit,omitempty"`
    Cache       bool              `json:"cache,omitempty"`
    CacheTTL    int               `json:"cache_ttl,omitempty" default:"300"`
}

type RateLimitConfig struct {
    Requests int `json:"requests"`
    Period   int `json:"period"` // seconds
}

// Usage example
func ExampleAPIClientTool() {
    tool, _ := tools.GetTool("api_client")
    
    result, err := tool.Execute(context.Background(), APIClientInput{
        BaseURL:  "https://api.openweathermap.org/data/2.5",
        Endpoint: "/weather",
        Params: map[string]string{
            "q":     "London",
            "appid": "your_api_key",
            "units": "metric",
        },
        RetryCount: 3,
        Cache:      true,
        CacheTTL:   600,
}
    
    output := result.(*APIClientOutput)
    fmt.Printf("API Response: %v\n", output.Data)
}
```

## System Tools

### ExecuteCommand Tool

Secure command execution with sandboxing and resource limits.

```go
type ExecuteCommandInput struct {
    Command   string            `json:"command" validate:"required"`
    Args      []string          `json:"args,omitempty"`
    WorkDir   string            `json:"work_dir,omitempty"`
    Env       map[string]string `json:"env,omitempty"`
    Timeout   int               `json:"timeout,omitempty" default:"30"`
    User      string            `json:"user,omitempty"`
    Shell     bool              `json:"shell,omitempty"`
    Capture   bool              `json:"capture,omitempty" default:"true"`
    StreamOutput bool          `json:"stream_output,omitempty"`
}

type ExecuteCommandOutput struct {
    ExitCode int    `json:"exit_code"`
    Stdout   string `json:"stdout"`
    Stderr   string `json:"stderr"`
    Duration time.Duration `json:"duration"`
    PID      int    `json:"pid"`
}

// Usage example
func ExampleExecuteCommandTool() {
    tool, _ := tools.GetTool("execute_command")
    
    result, err := tool.Execute(context.Background(), ExecuteCommandInput{
        Command: "ls",
        Args:    []string{"-la", "/tmp"},
        WorkDir: "/home/user",
        Timeout: 10,
        Capture: true,
}
    
    output := result.(*ExecuteCommandOutput)
    fmt.Printf("Exit code: %d\n", output.ExitCode)
    fmt.Printf("Output:\n%s\n", output.Stdout)
}
```

**Security Configuration:**
```yaml
execute_command:
  allowed_commands:
    - "ls"
    - "cat"
    - "grep"
    - "find"
  blocked_commands:
    - "rm"
    - "sudo"
    - "passwd"
  max_execution_time: 30
  sandbox_enabled: true
  resource_limits:
    max_memory: "128MB"
    max_cpu: 0.5
```

### SystemInfo Tool

Comprehensive system information gathering.

```go
type SystemInfoInput struct {
    Include []string `json:"include,omitempty"` // cpu, memory, disk, network, os, processes
    Detailed bool    `json:"detailed,omitempty"`
}

type SystemInfoOutput struct {
    OS        OSInfo        `json:"os"`
    CPU       CPUInfo       `json:"cpu"`
    Memory    MemoryInfo    `json:"memory"`
    Disk      []DiskInfo    `json:"disk"`
    Network   []NetworkInfo `json:"network"`
    Processes []ProcessInfo `json:"processes,omitempty"`
    Load      LoadInfo      `json:"load"`
    Uptime    time.Duration `json:"uptime"`
}

type OSInfo struct {
    Name         string `json:"name"`
    Version      string `json:"version"`
    Architecture string `json:"architecture"`
    Hostname     string `json:"hostname"`
    Kernel       string `json:"kernel"`
}

type CPUInfo struct {
    Model     string  `json:"model"`
    Cores     int     `json:"cores"`
    Threads   int     `json:"threads"`
    Frequency float64 `json:"frequency"`
    Usage     float64 `json:"usage"`
}

// Usage example
func ExampleSystemInfoTool() {
    tool, _ := tools.GetTool("system_info")
    
    result, err := tool.Execute(context.Background(), SystemInfoInput{
        Include:  []string{"cpu", "memory", "disk"},
        Detailed: true,
}
    
    output := result.(*SystemInfoOutput)
    fmt.Printf("OS: %s %s\n", output.OS.Name, output.OS.Version)
    fmt.Printf("CPU: %s (%d cores)\n", output.CPU.Model, output.CPU.Cores)
    fmt.Printf("Memory: %.1f%% used\n", output.Memory.UsedPercent)
}
```

## Data Processing Tools

### JSONProcessor Tool

Advanced JSON processing with JSONPath and transformation.

```go
type JSONProcessorInput struct {
    Data        interface{} `json:"data" validate:"required"`
    Operation   string      `json:"operation" validate:"required"` // parse, stringify, query, transform, validate
    Query       string      `json:"query,omitempty"`        // JSONPath expression
    Transform   string      `json:"transform,omitempty"`    // JQ-style transformation
    Schema      interface{} `json:"schema,omitempty"`       // JSON Schema for validation
    PrettyPrint bool        `json:"pretty_print,omitempty"`
    Options     JSONProcessorOptions `json:"options,omitempty"`
}

type JSONProcessorOptions struct {
    AllowComments    bool `json:"allow_comments,omitempty"`
    StrictMode      bool `json:"strict_mode,omitempty"`
    MaxDepth        int  `json:"max_depth,omitempty"`
    MaxSize         int  `json:"max_size,omitempty"`
    PreserveOrder   bool `json:"preserve_order,omitempty"`
}

type JSONProcessorOutput struct {
    Result   interface{} `json:"result"`
    Valid    bool        `json:"valid"`
    Errors   []string    `json:"errors,omitempty"`
    Warnings []string    `json:"warnings,omitempty"`
    Stats    JSONStats   `json:"stats"`
}

type JSONStats struct {
    Size     int `json:"size"`
    Depth    int `json:"depth"`
    Objects  int `json:"objects"`
    Arrays   int `json:"arrays"`
    Strings  int `json:"strings"`
    Numbers  int `json:"numbers"`
    Booleans int `json:"booleans"`
    Nulls    int `json:"nulls"`
}

// Usage examples
func ExampleJSONProcessorTool() {
    tool, _ := tools.GetTool("json_processor")
    
    // Parse and query JSON
    result, err := tool.Execute(context.Background(), JSONProcessorInput{
        Data:      `{"users": [{"name": "John", "age": 30}, {"name": "Jane", "age": 25}]}`,
        Operation: "query",
        Query:     "$.users[?(@.age > 25)].name",
}
    
    // Transform JSON structure
    transformResult, err := tool.Execute(context.Background(), JSONProcessorInput{
        Data:      map[string]interface{}{"a": 1, "b": 2},
        Operation: "transform",
        Transform: `{sum: (.a + .b), product: (.a * .b)}`,
}
    
    // Validate against schema
    schemaResult, err := tool.Execute(context.Background(), JSONProcessorInput{
        Data:      `{"name": "John", "age": 30}`,
        Operation: "validate",
        Schema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "name": map[string]interface{}{"type": "string"},
                "age":  map[string]interface{}{"type": "integer", "minimum": 0},
            },
            "required": []string{"name", "age"},
        },
}
    
    output := result.(*JSONProcessorOutput)
    fmt.Printf("Query result: %v\n", output.Result)
}
```

### TemplateRender Tool

Template rendering with multiple engines and context injection.

```go
type TemplateRenderInput struct {
    Template string                 `json:"template" validate:"required"`
    Data     map[string]interface{} `json:"data"`
    Engine   string                 `json:"engine,omitempty" default:"go"` // go, mustache, handlebars
    Options  TemplateOptions        `json:"options,omitempty"`
    Partials map[string]string      `json:"partials,omitempty"`
    Helpers  map[string]string      `json:"helpers,omitempty"`
}

type TemplateOptions struct {
    StrictMode    bool   `json:"strict_mode,omitempty"`
    MissingKey    string `json:"missing_key,omitempty"` // error, default, zero
    Delimiters    string `json:"delimiters,omitempty"`  // custom delimiters
    AutoEscape    bool   `json:"auto_escape,omitempty"`
    AllowUnsafe   bool   `json:"allow_unsafe,omitempty"`
}

type TemplateRenderOutput struct {
    Rendered string   `json:"rendered"`
    Errors   []string `json:"errors,omitempty"`
    Warnings []string `json:"warnings,omitempty"`
}

// Usage example
func ExampleTemplateRenderTool() {
    tool, _ := tools.GetTool("template_render")
    
    result, err := tool.Execute(context.Background(), TemplateRenderInput{
        Template: `Hello {{.Name}}! You have {{.Count}} new {{if eq .Count 1}}message{{else}}messages{{end}}.`,
        Data: map[string]interface{}{
            "Name":  "John",
            "Count": 3,
        },
        Engine: "go",
        Options: TemplateOptions{
            StrictMode: true,
            MissingKey: "error",
        },
}
    
    output := result.(*TemplateRenderOutput)
    fmt.Printf("Rendered: %s\n", output.Rendered)
    // Output: "Hello John! You have 3 new messages."
}
```

## Math & Computation Tools

### Calculator Tool

Advanced mathematical expressions and scientific computing.

```go
type CalculatorInput struct {
    Expression string                 `json:"expression" validate:"required"`
    Variables  map[string]float64     `json:"variables,omitempty"`
    Functions  map[string]string      `json:"functions,omitempty"`
    Precision  int                    `json:"precision,omitempty" default:"15"`
    Format     string                 `json:"format,omitempty" default:"decimal"` // decimal, scientific, engineering
    Mode       string                 `json:"mode,omitempty" default:"float"`     // float, rational, complex
}

type CalculatorOutput struct {
    Result     interface{} `json:"result"`
    Expression string      `json:"expression"`
    Steps      []string    `json:"steps,omitempty"`
    Variables  map[string]float64 `json:"variables,omitempty"`
    Errors     []string    `json:"errors,omitempty"`
}

// Usage examples
func ExampleCalculatorTool() {
    tool, _ := tools.GetTool("calculator")
    
    // Basic calculation
    result, err := tool.Execute(context.Background(), CalculatorInput{
        Expression: "2 + 3 * 4",
}
    
    // With variables
    varResult, err := tool.Execute(context.Background(), CalculatorInput{
        Expression: "sqrt(a^2 + b^2)",
        Variables: map[string]float64{
            "a": 3,
            "b": 4,
        },
        Precision: 10,
}
    
    // Complex numbers
    complexResult, err := tool.Execute(context.Background(), CalculatorInput{
        Expression: "(3+4i) * (2-i)",
        Mode:       "complex",
}
    
    output := result.(*CalculatorOutput)
    fmt.Printf("Result: %v\n", output.Result)
}
```

### Statistics Tool

Statistical analysis and data processing.

```go
type StatisticsInput struct {
    Data       []float64 `json:"data" validate:"required"`
    Operations []string  `json:"operations,omitempty"` // mean, median, mode, std, var, etc.
    Bins       int       `json:"bins,omitempty"`       // for histogram
    Confidence float64   `json:"confidence,omitempty"` // for confidence intervals
}

type StatisticsOutput struct {
    Count      int                    `json:"count"`
    Mean       float64                `json:"mean"`
    Median     float64                `json:"median"`
    Mode       []float64              `json:"mode"`
    StdDev     float64                `json:"std_dev"`
    Variance   float64                `json:"variance"`
    Min        float64                `json:"min"`
    Max        float64                `json:"max"`
    Range      float64                `json:"range"`
    Q1         float64                `json:"q1"`
    Q3         float64                `json:"q3"`
    IQR        float64                `json:"iqr"`
    Skewness   float64                `json:"skewness"`
    Kurtosis   float64                `json:"kurtosis"`
    Histogram  map[string]int         `json:"histogram,omitempty"`
    Percentiles map[string]float64    `json:"percentiles,omitempty"`
}

// Usage example
func ExampleStatisticsTool() {
    tool, _ := tools.GetTool("statistics")
    
    data := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
    
    result, err := tool.Execute(context.Background(), StatisticsInput{
        Data: data,
        Operations: []string{"mean", "median", "std", "histogram"},
        Bins: 5,
        Confidence: 0.95,
}
    
    output := result.(*StatisticsOutput)
    fmt.Printf("Mean: %.2f, Median: %.2f, StdDev: %.2f\n", 
        output.Mean, output.Median, output.StdDev)
}
```

## Date & Time Tools

### DateParser Tool

Flexible date parsing and formatting with timezone support.

```go
type DateParserInput struct {
    Date       string   `json:"date" validate:"required"`
    Format     string   `json:"format,omitempty"`     // Go time format or auto-detect
    Timezone   string   `json:"timezone,omitempty"`   // target timezone
    OutputFormat string `json:"output_format,omitempty"` // desired output format
    Locale     string   `json:"locale,omitempty"`     // for localized parsing
    Strict     bool     `json:"strict,omitempty"`     // strict parsing mode
}

type DateParserOutput struct {
    Parsed     time.Time              `json:"parsed"`
    Formatted  string                 `json:"formatted"`
    Timezone   string                 `json:"timezone"`
    Unix       int64                  `json:"unix"`
    ISO8601    string                 `json:"iso8601"`
    Components map[string]interface{} `json:"components"`
    Valid      bool                   `json:"valid"`
    Errors     []string               `json:"errors,omitempty"`
}

// Usage examples
func ExampleDateParserTool() {
    tool, _ := tools.GetTool("date_parser")
    
    // Parse flexible date format
    result, err := tool.Execute(context.Background(), DateParserInput{
        Date:         "March 15, 2024 at 3:30 PM",
        Timezone:     "America/New_York",
        OutputFormat: "2006-01-02 15:04:05 MST",
}
    
    // Parse with specific format
    specificResult, err := tool.Execute(context.Background(), DateParserInput{
        Date:         "2024-03-15T15:30:00Z",
        Format:       time.RFC3339,
        Timezone:     "UTC",
        OutputFormat: "January 2, 2006 at 3:04 PM",
}
    
    output := result.(*DateParserOutput)
    fmt.Printf("Parsed: %s\n", output.Formatted)
    fmt.Printf("Unix: %d\n", output.Unix)
}
```

## Text Processing Tools

### RegexTool

Regular expression matching and replacement with multiple patterns.

```go
type RegexInput struct {
    Text     string            `json:"text" validate:"required"`
    Pattern  string            `json:"pattern" validate:"required"`
    Operation string           `json:"operation" validate:"required"` // match, replace, split, extract
    Replace  string            `json:"replace,omitempty"`
    Flags    []string          `json:"flags,omitempty"` // i, m, s, x
    Global   bool              `json:"global,omitempty"`
    Groups   bool              `json:"groups,omitempty"`
}

type RegexOutput struct {
    Matches    []RegexMatch `json:"matches"`
    Result     string       `json:"result"`
    Groups     [][]string   `json:"groups,omitempty"`
    Success    bool         `json:"success"`
    Count      int          `json:"count"`
}

type RegexMatch struct {
    Text   string `json:"text"`
    Start  int    `json:"start"`
    End    int    `json:"end"`
    Groups []string `json:"groups,omitempty"`
}

// Usage examples
func ExampleRegexTool() {
    tool, _ := tools.GetTool("regex")
    
    // Extract email addresses
    result, err := tool.Execute(context.Background(), RegexInput{
        Text:      "Contact us at hello@example.com or support@company.org",
        Pattern:   `\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`,
        Operation: "match",
        Global:    true,
}
    
    // Replace with groups
    replaceResult, err := tool.Execute(context.Background(), RegexInput{
        Text:      "Today is 2024-03-15",
        Pattern:   `(\d{4})-(\d{2})-(\d{2})`,
        Operation: "replace",
        Replace:   "$2/$3/$1",
        Global:    true,
}
    
    output := result.(*RegexOutput)
    fmt.Printf("Found %d matches\n", output.Count)
    for _, match := range output.Matches {
        fmt.Printf("- %s\n", match.Text)
    }
}
```

## Tool Configuration and Management

### Global Tool Configuration

```yaml
# config/tools.yaml
tools:
  # Global settings
  global:
    timeout: 30
    max_memory: "128MB"
    temp_dir: "/tmp/go-llms-tools"
    log_level: "info"
    
  # Security settings
  security:
    sandbox_enabled: true
    allowed_networks:
      - "0.0.0.0/0"
    blocked_hosts:
      - "localhost"
      - "127.0.0.1"
    max_file_size: "10MB"
    
  # Tool-specific configurations
  read_file:
    max_size: 10485760
    allowed_paths:
      - "/home/user"
      - "/tmp"
    encoding: "utf-8"
    
  http_request:
    timeout: 30
    max_redirects: 5
    user_agent: "Go-LLMs/0.3.5"
    verify_ssl: true
    
  execute_command:
    sandbox: true
    allowed_commands:
      - "ls"
      - "cat"
      - "grep"
    timeout: 30
```

### Tool Usage Analytics

```go
// ToolAnalytics tracks tool usage and performance
type ToolAnalytics struct {
    registry ToolRegistry
    metrics  *MetricsCollector
    storage  AnalyticsStorage
}

type ToolUsageStats struct {
    ToolName      string        `json:"tool_name"`
    CallCount     int64         `json:"call_count"`
    SuccessCount  int64         `json:"success_count"`
    ErrorCount    int64         `json:"error_count"`
    AvgDuration   time.Duration `json:"avg_duration"`
    LastUsed      time.Time     `json:"last_used"`
    TotalDataIn   int64         `json:"total_data_in"`
    TotalDataOut  int64         `json:"total_data_out"`
}

// Example usage tracking
func ExampleToolAnalytics() {
    analytics := NewToolAnalytics()
    
    // Get usage statistics
    stats, err := analytics.GetToolStats("http_request", time.Hour*24)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("HTTP Request Tool Usage (24h):\n")
    fmt.Printf("- Calls: %d\n", stats.CallCount)
    fmt.Printf("- Success Rate: %.1f%%\n", 
        float64(stats.SuccessCount)/float64(stats.CallCount)*100)
    fmt.Printf("- Avg Duration: %v\n", stats.AvgDuration)
    
    // Get most used tools
    topTools, err := analytics.GetTopTools(10, time.Hour*24*7)
    for i, tool := range topTools {
        fmt.Printf("%d. %s (%d calls)\n", i+1, tool.ToolName, tool.CallCount)
    }
}
```

### Tool Performance Monitoring

```go
// PerformanceMonitor tracks tool execution metrics
type PerformanceMonitor struct {
    metrics map[string]*ToolMetrics
    mu      sync.RWMutex
}

type ToolMetrics struct {
    ExecutionTimes []time.Duration
    MemoryUsage    []int64
    ErrorRates     []float64
    LastUpdated    time.Time
}

// Example monitoring setup
func ExamplePerformanceMonitoring() {
    monitor := NewPerformanceMonitor()
    
    // Wrap tool execution with monitoring
    executor := NewMonitoredExecutor(monitor)
    
    tool, _ := tools.GetTool("json_processor")
    result, err := executor.Execute(context.Background(), tool, input)
    
    // Get performance report
    report := monitor.GenerateReport("json_processor", time.Hour*24)
    fmt.Printf("Performance Report for JSON Processor:\n")
    fmt.Printf("- P50 Latency: %v\n", report.P50Latency)
    fmt.Printf("- P95 Latency: %v\n", report.P95Latency)
    fmt.Printf("- Error Rate: %.2f%%\n", report.ErrorRate*100)
    fmt.Printf("- Avg Memory: %s\n", formatBytes(report.AvgMemory))
}
```

This comprehensive reference provides detailed documentation for all built-in tools in Go-LLMs, including complete input/output schemas, usage examples, configuration options, and integration patterns. The tools are designed to be composable, secure, and performant, providing a solid foundation for building sophisticated agent-based applications.