# Tool Discovery: Runtime Registration and Metadata

> **[Project Root](/) / [Documentation](../..) / [Technical Documentation](../../technical) / [Tools](../../technical/tools) / Tool Discovery**

Comprehensive guide to the Go-LLMs tool discovery system, covering runtime registration mechanisms, metadata management, dynamic loading, search and filtering capabilities, plugin architectures, and advanced discovery patterns for building extensible tool ecosystems.

## Discovery System Architecture

### Core Discovery Interfaces

```go
// ToolDiscovery manages tool discovery and registration
type ToolDiscovery interface {
    // Discovery methods
    DiscoverTools(ctx context.Context, sources []DiscoverySource) ([]DiscoveredTool, error)
    DiscoverFromPath(ctx context.Context, path string, options DiscoveryOptions) ([]DiscoveredTool, error)
    DiscoverFromRegistry(ctx context.Context, registry string) ([]DiscoveredTool, error)
    
    // Registration
    RegisterDiscoveredTool(tool DiscoveredTool) error
    RegisterDiscoverySource(source DiscoverySource) error
    UnregisterDiscoverySource(sourceID string) error
    
    // Search and filtering
    SearchTools(ctx context.Context, query SearchQuery) ([]ToolInfo, error)
    FilterTools(tools []ToolInfo, filters []Filter) []ToolInfo
    
    // Metadata management
    GetToolMetadata(toolName string) (*ToolMetadata, error)
    UpdateToolMetadata(toolName string, metadata *ToolMetadata) error
    
    // Lifecycle
    StartDiscovery(ctx context.Context) error
    StopDiscovery(ctx context.Context) error
    RefreshDiscovery(ctx context.Context) error
    
    // Events
    Subscribe(eventType DiscoveryEventType) (<-chan DiscoveryEvent, error)
    Unsubscribe(eventType DiscoveryEventType) error
}

// DiscoverySource represents a source of tools
type DiscoverySource interface {
    // Source identification
    ID() string
    Name() string
    Type() SourceType
    
    // Discovery
    Discover(ctx context.Context, options DiscoveryOptions) ([]DiscoveredTool, error)
    
    // Validation
    Validate() error
    IsAvailable() bool
    
    // Configuration
    GetConfig() SourceConfig
    SetConfig(config SourceConfig) error
    
    // Monitoring
    GetStats() SourceStats
    GetLastError() error
}

// ToolMetadata contains comprehensive tool information
type ToolMetadata struct {
    // Basic information
    Name        string    `json:"name"`
    Version     string    `json:"version"`
    Description string    `json:"description"`
    Category    string    `json:"category"`
    Author      string    `json:"author,omitempty"`
    License     string    `json:"license,omitempty"`
    
    // Discovery information
    Source      string    `json:"source"`
    SourceType  string    `json:"source_type"`
    DiscoveredAt time.Time `json:"discovered_at"`
    
    // Capabilities and requirements
    Capabilities   ToolCapabilities      `json:"capabilities"`
    Requirements   ToolRequirements      `json:"requirements"`
    Dependencies   []string              `json:"dependencies,omitempty"`
    
    // Schema information
    InputSchema    *jsonschema.Schema    `json:"input_schema,omitempty"`
    OutputSchema   *jsonschema.Schema    `json:"output_schema,omitempty"`
    
    // Documentation
    Documentation  *ToolDocumentation    `json:"documentation,omitempty"`
    Examples       []ToolExample         `json:"examples,omitempty"`
    
    // Usage statistics
    Usage          *UsageStats           `json:"usage,omitempty"`
    
    // Tags and classification
    Tags           []string              `json:"tags,omitempty"`
    Keywords       []string              `json:"keywords,omitempty"`
    Classification map[string]string     `json:"classification,omitempty"`
    
    // Validation and quality
    Validated      bool                  `json:"validated"`
    Quality        QualityMetrics        `json:"quality"`
    
    // Custom metadata
    Custom         map[string]interface{} `json:"custom,omitempty"`
}

type DiscoveredTool struct {
    Tool     Tool         `json:"-"`
    Metadata ToolMetadata `json:"metadata"`
    Source   string       `json:"source"`
    LoadPath string       `json:"load_path,omitempty"`
    Config   ToolConfig   `json:"config,omitempty"`
}

type SourceType string

const (
    SourceTypeFileSystem SourceType = "filesystem"
    SourceTypeRegistry   SourceType = "registry"
    SourceTypePlugin     SourceType = "plugin"
    SourceTypeRemote     SourceType = "remote"
    SourceTypeGit        SourceType = "git"
    SourceTypeDocker     SourceType = "docker"
    SourceTypeEmbedded   SourceType = "embedded"
)

type DiscoveryOptions struct {
    Recursive    bool              `json:"recursive"`
    IncludeTests bool              `json:"include_tests"`
    IncludeDocs  bool              `json:"include_docs"`
    Filters      []Filter          `json:"filters,omitempty"`
    MaxDepth     int               `json:"max_depth,omitempty"`
    Timeout      time.Duration     `json:"timeout,omitempty"`
    Parallel     bool              `json:"parallel"`
    CacheResults bool              `json:"cache_results"`
    Validate     bool              `json:"validate"`
}
```

### Default Discovery Implementation

```go
// DefaultToolDiscovery implements comprehensive tool discovery
type DefaultToolDiscovery struct {
    sources      map[string]DiscoverySource
    cache        DiscoveryCache
    registry     ToolRegistry
    searcher     *ToolSearcher
    validator    *ToolValidator
    eventBus     *EventBus
    config       DiscoveryConfig
    mu           sync.RWMutex
}

type DiscoveryConfig struct {
    AutoRefresh       bool          `yaml:"auto_refresh" json:"auto_refresh"`
    RefreshInterval   time.Duration `yaml:"refresh_interval" json:"refresh_interval"`
    ParallelDiscovery bool          `yaml:"parallel_discovery" json:"parallel_discovery"`
    MaxWorkers        int           `yaml:"max_workers" json:"max_workers"`
    CacheEnabled      bool          `yaml:"cache_enabled" json:"cache_enabled"`
    CacheTTL          time.Duration `yaml:"cache_ttl" json:"cache_ttl"`
    ValidateOnDiscovery bool        `yaml:"validate_on_discovery" json:"validate_on_discovery"`
}

// NewDefaultToolDiscovery creates a new discovery service
func NewDefaultToolDiscovery(config DiscoveryConfig) *DefaultToolDiscovery {
    return &DefaultToolDiscovery{
        sources:   make(map[string]DiscoverySource),
        cache:     NewDiscoveryCache(config.CacheTTL),
        registry:  NewToolRegistry(),
        searcher:  NewToolSearcher(),
        validator: NewToolValidator(),
        eventBus:  NewEventBus(),
        config:    config,
    }
}

// DiscoverTools discovers tools from multiple sources
func (d *DefaultToolDiscovery) DiscoverTools(ctx context.Context, sources []DiscoverySource) ([]DiscoveredTool, error) {
    var allTools []DiscoveredTool
    var mu sync.Mutex
    var wg sync.WaitGroup
    
    // Channel for collecting errors
    errorChan := make(chan error, len(sources))
    
    for _, source := range sources {
        wg.Add(1)
        
        if d.config.ParallelDiscovery {
            go func(src DiscoverySource) {
                defer wg.Done()
                
                tools, err := d.discoverFromSource(ctx, src)
                if err != nil {
                    errorChan <- fmt.Errorf("discovery failed for source %s: %w", src.ID(), err)
                    return
                }
                
                mu.Lock()
                allTools = append(allTools, tools...)
                mu.Unlock()
            }(source)
        } else {
            go func(src DiscoverySource) {
                defer wg.Done()
                
                tools, err := d.discoverFromSource(ctx, src)
                if err != nil {
                    errorChan <- fmt.Errorf("discovery failed for source %s: %w", src.ID(), err)
                    return
                }
                
                allTools = append(allTools, tools...)
            }(source)
        }
    }
    
    wg.Wait()
    close(errorChan)
    
    // Collect errors
    var errors []error
    for err := range errorChan {
        errors = append(errors, err)
    }
    
    if len(errors) > 0 {
        return allTools, fmt.Errorf("discovery completed with errors: %v", errors)
    }
    
    // Deduplicate tools
    deduplicatedTools := d.deduplicateTools(allTools)
    
    // Emit discovery event
    d.eventBus.Emit(DiscoveryEvent{
        Type:      EventTypeDiscoveryCompleted,
        Timestamp: time.Now(),
        Data: map[string]interface{}{
            "tool_count":    len(deduplicatedTools),
            "source_count":  len(sources),
            "errors":        errors,
        },
}
    
    return deduplicatedTools, nil
}

// discoverFromSource discovers tools from a single source
func (d *DefaultToolDiscovery) discoverFromSource(ctx context.Context, source DiscoverySource) ([]DiscoveredTool, error) {
    // Check cache first
    if d.config.CacheEnabled {
        if cached, found := d.cache.Get(source.ID()); found {
            return cached, nil
        }
    }
    
    // Validate source availability
    if !source.IsAvailable() {
        return nil, fmt.Errorf("source %s is not available", source.ID())
    }
    
    // Perform discovery
    options := DiscoveryOptions{
        Recursive:    true,
        IncludeTests: false,
        IncludeDocs:  true,
        Validate:     d.config.ValidateOnDiscovery,
        Parallel:     d.config.ParallelDiscovery,
        CacheResults: d.config.CacheEnabled,
    }
    
    tools, err := source.Discover(ctx, options)
    if err != nil {
        return nil, fmt.Errorf("discovery failed: %w", err)
    }
    
    // Validate discovered tools
    if d.config.ValidateOnDiscovery {
        validatedTools := make([]DiscoveredTool, 0, len(tools))
        for _, tool := range tools {
            if err := d.validator.ValidateTool(tool.Tool); err != nil {
                d.eventBus.Emit(DiscoveryEvent{
                    Type: EventTypeValidationFailed,
                    Data: map[string]interface{}{
                        "tool_name": tool.Metadata.Name,
                        "error":     err.Error(),
                    },
}
                continue
            }
            
            tool.Metadata.Validated = true
            validatedTools = append(validatedTools, tool)
        }
        tools = validatedTools
    }
    
    // Cache results
    if d.config.CacheEnabled {
        d.cache.Set(source.ID(), tools)
    }
    
    return tools, nil
}

// SearchTools searches for tools matching the query
func (d *DefaultToolDiscovery) SearchTools(ctx context.Context, query SearchQuery) ([]ToolInfo, error) {
    // Get all tools from registry
    allTools := d.registry.List()
    
    // Convert to ToolInfo for searching
    toolInfos := make([]ToolInfo, len(allTools))
    for i, tool := range allTools {
        metadata, _ := d.GetToolMetadata(tool.Name())
        toolInfos[i] = ToolInfo{
            Name:        tool.Name(),
            Description: tool.Description(),
            Version:     tool.Version(),
            Category:    metadata.Category,
            Tags:        metadata.Tags,
            Metadata:    metadata,
        }
    }
    
    // Perform search
    return d.searcher.Search(toolInfos, query)
}

type SearchQuery struct {
    Text         string            `json:"text,omitempty"`
    Category     string            `json:"category,omitempty"`
    Tags         []string          `json:"tags,omitempty"`
    Capabilities []string          `json:"capabilities,omitempty"`
    Author       string            `json:"author,omitempty"`
    Version      string            `json:"version,omitempty"`
    Filters      map[string]string `json:"filters,omitempty"`
    Limit        int               `json:"limit,omitempty"`
    Offset       int               `json:"offset,omitempty"`
    SortBy       string            `json:"sort_by,omitempty"`
    SortOrder    string            `json:"sort_order,omitempty"`
}

type ToolInfo struct {
    Name        string       `json:"name"`
    Description string       `json:"description"`
    Version     string       `json:"version"`
    Category    string       `json:"category"`
    Tags        []string     `json:"tags"`
    Metadata    *ToolMetadata `json:"metadata,omitempty"`
}
```

## Discovery Sources

### File System Discovery

```go
// FileSystemDiscoverySource discovers tools from the file system
type FileSystemDiscoverySource struct {
    id        string
    name      string
    basePath  string
    patterns  []string
    loader    *ToolLoader
    config    FileSystemSourceConfig
}

type FileSystemSourceConfig struct {
    BasePath         string        `yaml:"base_path" json:"base_path"`
    SearchPatterns   []string      `yaml:"search_patterns" json:"search_patterns"`
    ExcludePatterns  []string      `yaml:"exclude_patterns" json:"exclude_patterns"`
    MaxDepth         int           `yaml:"max_depth" json:"max_depth"`
    FollowSymlinks   bool          `yaml:"follow_symlinks" json:"follow_symlinks"`
    WatchForChanges  bool          `yaml:"watch_for_changes" json:"watch_for_changes"`
    CacheResults     bool          `yaml:"cache_results" json:"cache_results"`
    ValidateTools    bool          `yaml:"validate_tools" json:"validate_tools"`
}

// NewFileSystemDiscoverySource creates a filesystem discovery source
func NewFileSystemDiscoverySource(basePath string, config FileSystemSourceConfig) *FileSystemDiscoverySource {
    return &FileSystemDiscoverySource{
        id:       fmt.Sprintf("fs_%s", filepath.Base(basePath)),
        name:     fmt.Sprintf("FileSystem: %s", basePath),
        basePath: basePath,
        patterns: config.SearchPatterns,
        loader:   NewToolLoader(),
        config:   config,
    }
}

// Discover finds tools in the filesystem
func (fs *FileSystemDiscoverySource) Discover(ctx context.Context, options DiscoveryOptions) ([]DiscoveredTool, error) {
    var discoveredTools []DiscoveredTool
    
    // Walk the directory tree
    err := filepath.Walk(fs.basePath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        
        // Check context cancellation
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }
        
        // Skip if depth exceeded
        if options.MaxDepth > 0 {
            depth := strings.Count(strings.TrimPrefix(path, fs.basePath), string(os.PathSeparator))
            if depth > options.MaxDepth {
                return filepath.SkipDir
            }
        }
        
        // Skip directories
        if info.IsDir() {
            return nil
        }
        
        // Check if file matches patterns
        if !fs.matchesPatterns(path) {
            return nil
        }
        
        // Attempt to load tool
        tool, err := fs.loader.LoadFromFile(path)
        if err != nil {
            // Log error but continue discovery
            logError("failed to load tool from %s: %v", path, err)
            return nil
        }
        
        // Create discovered tool
        discoveredTool := DiscoveredTool{
            Tool:     tool,
            Source:   fs.ID(),
            LoadPath: path,
            Metadata: fs.extractMetadata(tool, path, info),
        }
        
        discoveredTools = append(discoveredTools, discoveredTool)
        return nil
}
    
    if err != nil {
        return nil, fmt.Errorf("filesystem discovery failed: %w", err)
    }
    
    return discoveredTools, nil
}

// matchesPatterns checks if a file path matches discovery patterns
func (fs *FileSystemDiscoverySource) matchesPatterns(path string) bool {
    // Check include patterns
    matched := false
    for _, pattern := range fs.patterns {
        if match, _ := filepath.Match(pattern, filepath.Base(path)); match {
            matched = true
            break
        }
    }
    
    if !matched {
        return false
    }
    
    // Check exclude patterns
    for _, pattern := range fs.config.ExcludePatterns {
        if match, _ := filepath.Match(pattern, filepath.Base(path)); match {
            return false
        }
    }
    
    return true
}

// extractMetadata extracts metadata from a file-based tool
func (fs *FileSystemDiscoverySource) extractMetadata(tool Tool, path string, info os.FileInfo) ToolMetadata {
    metadata := ToolMetadata{
        Name:         tool.Name(),
        Version:      tool.Version(),
        Description:  tool.Description(),
        Source:       fs.ID(),
        SourceType:   string(SourceTypeFileSystem),
        DiscoveredAt: time.Now(),
        Capabilities: tool.GetCapabilities(),
    }
    
    // Extract file-specific metadata
    metadata.Custom = map[string]interface{}{
        "file_path":    path,
        "file_size":    info.Size(),
        "file_mode":    info.Mode().String(),
        "modified_at":  info.ModTime(),
    }
    
    // Try to extract additional metadata from file
    if fileMetadata, err := fs.extractFileMetadata(path); err == nil {
        metadata.Author = fileMetadata.Author
        metadata.License = fileMetadata.License
        metadata.Tags = fileMetadata.Tags
        metadata.Keywords = fileMetadata.Keywords
    }
    
    return metadata
}
```

### Plugin Discovery

```go
// PluginDiscoverySource discovers tools from plugin files
type PluginDiscoverySource struct {
    id         string
    name       string
    pluginDir  string
    loader     *PluginLoader
    registry   *PluginRegistry
    config     PluginSourceConfig
}

type PluginSourceConfig struct {
    PluginDir       string   `yaml:"plugin_dir" json:"plugin_dir"`
    AllowedFormats  []string `yaml:"allowed_formats" json:"allowed_formats"`
    SecurityPolicy  string   `yaml:"security_policy" json:"security_policy"`
    Sandbox         bool     `yaml:"sandbox" json:"sandbox"`
    MaxPluginSize   int64    `yaml:"max_plugin_size" json:"max_plugin_size"`
    LoadTimeout     time.Duration `yaml:"load_timeout" json:"load_timeout"`
}

// NewPluginDiscoverySource creates a plugin discovery source
func NewPluginDiscoverySource(pluginDir string, config PluginSourceConfig) *PluginDiscoverySource {
    return &PluginDiscoverySource{
        id:        fmt.Sprintf("plugin_%s", filepath.Base(pluginDir)),
        name:      fmt.Sprintf("Plugins: %s", pluginDir),
        pluginDir: pluginDir,
        loader:    NewPluginLoader(),
        registry:  NewPluginRegistry(),
        config:    config,
    }
}

// Discover finds tools in plugin files
func (ps *PluginDiscoverySource) Discover(ctx context.Context, options DiscoveryOptions) ([]DiscoveredTool, error) {
    var discoveredTools []DiscoveredTool
    
    // Scan plugin directory
    pluginFiles, err := ps.scanPluginFiles()
    if err != nil {
        return nil, fmt.Errorf("failed to scan plugin files: %w", err)
    }
    
    // Load plugins in parallel or sequential based on options
    if options.Parallel {
        discoveredTools, err = ps.loadPluginsParallel(ctx, pluginFiles, options)
    } else {
        discoveredTools, err = ps.loadPluginsSequential(ctx, pluginFiles, options)
    }
    
    if err != nil {
        return nil, fmt.Errorf("plugin loading failed: %w", err)
    }
    
    return discoveredTools, nil
}

// loadPluginsParallel loads plugins concurrently
func (ps *PluginDiscoverySource) loadPluginsParallel(ctx context.Context, pluginFiles []string, options DiscoveryOptions) ([]DiscoveredTool, error) {
    var discoveredTools []DiscoveredTool
    var mu sync.Mutex
    var wg sync.WaitGroup
    
    // Limit concurrency
    semaphore := make(chan struct{}, 10)
    errorChan := make(chan error, len(pluginFiles))
    
    for _, pluginFile := range pluginFiles {
        wg.Add(1)
        go func(file string) {
            defer wg.Done()
            
            // Acquire semaphore
            semaphore <- struct{}{}
            defer func() { <-semaphore }()
            
            // Load plugin with timeout
            pluginCtx, cancel := context.WithTimeout(ctx, ps.config.LoadTimeout)
            defer cancel()
            
            tools, err := ps.loadPluginFile(pluginCtx, file, options)
            if err != nil {
                errorChan <- fmt.Errorf("failed to load plugin %s: %w", file, err)
                return
            }
            
            mu.Lock()
            discoveredTools = append(discoveredTools, tools...)
            mu.Unlock()
        }(pluginFile)
    }
    
    wg.Wait()
    close(errorChan)
    
    // Collect errors
    var errors []error
    for err := range errorChan {
        errors = append(errors, err)
    }
    
    if len(errors) > 0 {
        return discoveredTools, fmt.Errorf("plugin loading completed with errors: %v", errors)
    }
    
    return discoveredTools, nil
}

// loadPluginFile loads tools from a single plugin file
func (ps *PluginDiscoverySource) loadPluginFile(ctx context.Context, pluginFile string, options DiscoveryOptions) ([]DiscoveredTool, error) {
    // Validate plugin file
    if err := ps.validatePluginFile(pluginFile); err != nil {
        return nil, fmt.Errorf("plugin validation failed: %w", err)
    }
    
    // Load plugin
    plugin, err := ps.loader.LoadPlugin(ctx, pluginFile, PluginLoadOptions{
        Sandbox:        ps.config.Sandbox,
        SecurityPolicy: ps.config.SecurityPolicy,
        Timeout:        ps.config.LoadTimeout,
}
    if err != nil {
        return nil, fmt.Errorf("failed to load plugin: %w", err)
    }
    
    // Extract tools from plugin
    tools, err := plugin.GetTools()
    if err != nil {
        return nil, fmt.Errorf("failed to extract tools from plugin: %w", err)
    }
    
    // Create discovered tools
    var discoveredTools []DiscoveredTool
    for _, tool := range tools {
        discoveredTool := DiscoveredTool{
            Tool:     tool,
            Source:   ps.ID(),
            LoadPath: pluginFile,
            Metadata: ps.extractPluginMetadata(tool, plugin, pluginFile),
        }
        
        discoveredTools = append(discoveredTools, discoveredTool)
    }
    
    return discoveredTools, nil
}

// validatePluginFile validates a plugin file before loading
func (ps *PluginDiscoverySource) validatePluginFile(pluginFile string) error {
    // Check file size
    info, err := os.Stat(pluginFile)
    if err != nil {
        return fmt.Errorf("failed to stat plugin file: %w", err)
    }
    
    if ps.config.MaxPluginSize > 0 && info.Size() > ps.config.MaxPluginSize {
        return fmt.Errorf("plugin file size %d exceeds maximum %d", info.Size(), ps.config.MaxPluginSize)
    }
    
    // Check file format
    ext := filepath.Ext(pluginFile)
    if len(ps.config.AllowedFormats) > 0 {
        allowed := false
        for _, format := range ps.config.AllowedFormats {
            if ext == format {
                allowed = true
                break
            }
        }
        if !allowed {
            return fmt.Errorf("plugin format %s not allowed", ext)
        }
    }
    
    return nil
}
```

### Remote Registry Discovery

```go
// RemoteRegistryDiscoverySource discovers tools from remote registries
type RemoteRegistryDiscoverySource struct {
    id       string
    name     string
    baseURL  string
    client   *RegistryClient
    cache    *RegistryCache
    config   RemoteRegistryConfig
}

type RemoteRegistryConfig struct {
    BaseURL         string        `yaml:"base_url" json:"base_url"`
    APIKey          string        `yaml:"api_key,omitempty" json:"api_key,omitempty"`
    Timeout         time.Duration `yaml:"timeout" json:"timeout"`
    RetryAttempts   int           `yaml:"retry_attempts" json:"retry_attempts"`
    CacheEnabled    bool          `yaml:"cache_enabled" json:"cache_enabled"`
    CacheTTL        time.Duration `yaml:"cache_ttl" json:"cache_ttl"`
    VerifySignature bool          `yaml:"verify_signature" json:"verify_signature"`
    TrustLevel      string        `yaml:"trust_level" json:"trust_level"`
}

// NewRemoteRegistryDiscoverySource creates a remote registry discovery source
func NewRemoteRegistryDiscoverySource(baseURL string, config RemoteRegistryConfig) *RemoteRegistryDiscoverySource {
    return &RemoteRegistryDiscoverySource{
        id:      fmt.Sprintf("registry_%s", hashURL(baseURL)),
        name:    fmt.Sprintf("Registry: %s", baseURL),
        baseURL: baseURL,
        client:  NewRegistryClient(baseURL, config),
        cache:   NewRegistryCache(config.CacheTTL),
        config:  config,
    }
}

// Discover fetches tools from remote registry
func (rs *RemoteRegistryDiscoverySource) Discover(ctx context.Context, options DiscoveryOptions) ([]DiscoveredTool, error) {
    // Check cache first
    if rs.config.CacheEnabled && options.CacheResults {
        if cached, found := rs.cache.GetToolList(); found {
            return rs.convertToDiscoveredTools(cached), nil
        }
    }
    
    // Fetch tool list from registry
    toolList, err := rs.client.ListTools(ctx, RegistryListOptions{
        Category:    options.Filters,
        IncludeDocs: options.IncludeDocs,
        Page:        1,
        Limit:       1000, // Get all tools
}
    if err != nil {
        return nil, fmt.Errorf("failed to fetch tool list: %w", err)
    }
    
    // Download and validate tools
    discoveredTools := make([]DiscoveredTool, 0, len(toolList.Tools))
    
    for _, toolInfo := range toolList.Tools {
        // Check if tool meets requirements
        if !rs.meetsRequirements(toolInfo, options) {
            continue
        }
        
        // Download tool
        tool, err := rs.downloadTool(ctx, toolInfo)
        if err != nil {
            logError("failed to download tool %s: %v", toolInfo.Name, err)
            continue
        }
        
        // Validate tool if required
        if options.Validate {
            if err := rs.validateDownloadedTool(tool, toolInfo); err != nil {
                logError("tool validation failed for %s: %v", toolInfo.Name, err)
                continue
            }
        }
        
        // Create discovered tool
        discoveredTool := DiscoveredTool{
            Tool:     tool,
            Source:   rs.ID(),
            LoadPath: toolInfo.DownloadURL,
            Metadata: rs.convertRegistryMetadata(toolInfo),
        }
        
        discoveredTools = append(discoveredTools, discoveredTool)
    }
    
    // Cache results
    if rs.config.CacheEnabled {
        rs.cache.SetToolList(toolList)
    }
    
    return discoveredTools, nil
}

// downloadTool downloads a tool from the registry
func (rs *RemoteRegistryDiscoverySource) downloadTool(ctx context.Context, toolInfo RegistryToolInfo) (Tool, error) {
    // Download tool package
    packageData, err := rs.client.DownloadTool(ctx, toolInfo.Name, toolInfo.Version)
    if err != nil {
        return nil, fmt.Errorf("failed to download tool package: %w", err)
    }
    
    // Verify signature if required
    if rs.config.VerifySignature && toolInfo.Signature != "" {
        if err := rs.verifySignature(packageData, toolInfo.Signature); err != nil {
            return nil, fmt.Errorf("signature verification failed: %w", err)
        }
    }
    
    // Load tool from package
    loader := NewPackageLoader()
    tool, err := loader.LoadFromData(packageData, LoadOptions{
        Format:         toolInfo.Format,
        TrustLevel:     rs.config.TrustLevel,
        ValidateSchema: true,
}
    if err != nil {
        return nil, fmt.Errorf("failed to load tool from package: %w", err)
    }
    
    return tool, nil
}

type RegistryToolInfo struct {
    Name        string            `json:"name"`
    Version     string            `json:"version"`
    Description string            `json:"description"`
    Category    string            `json:"category"`
    Author      string            `json:"author"`
    License     string            `json:"license"`
    DownloadURL string            `json:"download_url"`
    Signature   string            `json:"signature,omitempty"`
    Format      string            `json:"format"`
    Size        int64             `json:"size"`
    Checksum    string            `json:"checksum"`
    Tags        []string          `json:"tags"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
}
```

## Advanced Search and Filtering

### Tool Search Engine

```go
// ToolSearcher implements advanced tool search capabilities
type ToolSearcher struct {
    indexer    *SearchIndexer
    scorer     *RelevanceScorer
    filters    *FilterEngine
    ranker     *ResultRanker
}

// NewToolSearcher creates a new tool searcher
func NewToolSearcher() *ToolSearcher {
    return &ToolSearcher{
        indexer: NewSearchIndexer(),
        scorer:  NewRelevanceScorer(),
        filters: NewFilterEngine(),
        ranker:  NewResultRanker(),
    }
}

// Search performs a comprehensive search across tools
func (ts *ToolSearcher) Search(tools []ToolInfo, query SearchQuery) ([]ToolInfo, error) {
    // Build search index if not already built
    if !ts.indexer.IsIndexed() {
        if err := ts.indexer.BuildIndex(tools); err != nil {
            return nil, fmt.Errorf("failed to build search index: %w", err)
        }
    }
    
    // Perform text search if query has text
    var candidates []ToolInfo
    if query.Text != "" {
        textResults, err := ts.performTextSearch(query.Text, tools)
        if err != nil {
            return nil, fmt.Errorf("text search failed: %w", err)
        }
        candidates = textResults
    } else {
        candidates = tools
    }
    
    // Apply filters
    filtered := ts.filters.ApplyFilters(candidates, ts.buildFilters(query))
    
    // Score and rank results
    scored := ts.scorer.ScoreResults(filtered, query)
    ranked := ts.ranker.RankResults(scored, query)
    
    // Apply pagination
    return ts.applyPagination(ranked, query), nil
}

// performTextSearch performs full-text search on tools
func (ts *ToolSearcher) performTextSearch(text string, tools []ToolInfo) ([]ToolInfo, error) {
    // Tokenize search text
    tokens := ts.tokenizeText(text)
    
    var results []ScoredResult
    
    for _, tool := range tools {
        score := ts.calculateTextScore(tool, tokens)
        if score > 0 {
            results = append(results, ScoredResult{
                Tool:  tool,
                Score: score,
}
        }
    }
    
    // Sort by score descending
    sort.Slice(results, func(i, j int) bool {
        return results[i].Score > results[j].Score
}
    
    // Extract tools
    var searchResults []ToolInfo
    for _, result := range results {
        searchResults = append(searchResults, result.Tool)
    }
    
    return searchResults, nil
}

// calculateTextScore calculates relevance score for text search
func (ts *ToolSearcher) calculateTextScore(tool ToolInfo, tokens []string) float64 {
    score := 0.0
    
    // Score name matches (highest weight)
    nameScore := ts.calculateFieldScore(tool.Name, tokens) * 3.0
    score += nameScore
    
    // Score description matches
    descScore := ts.calculateFieldScore(tool.Description, tokens) * 2.0
    score += descScore
    
    // Score tag matches
    for _, tag := range tool.Tags {
        tagScore := ts.calculateFieldScore(tag, tokens) * 1.5
        score += tagScore
    }
    
    // Score category matches
    catScore := ts.calculateFieldScore(tool.Category, tokens) * 1.0
    score += catScore
    
    return score
}

type ScoredResult struct {
    Tool  ToolInfo `json:"tool"`
    Score float64  `json:"score"`
}

// FilterEngine provides advanced filtering capabilities
type FilterEngine struct {
    predicates map[string]FilterPredicate
}

type FilterPredicate func(tool ToolInfo, value interface{}) bool

// NewFilterEngine creates a new filter engine
func NewFilterEngine() *FilterEngine {
    engine := &FilterEngine{
        predicates: make(map[string]FilterPredicate),
    }
    
    // Register built-in predicates
    engine.registerBuiltinPredicates()
    
    return engine
}

// registerBuiltinPredicates registers standard filter predicates
func (fe *FilterEngine) registerBuiltinPredicates() {
    fe.predicates["category"] = func(tool ToolInfo, value interface{}) bool {
        if category, ok := value.(string); ok {
            return tool.Category == category
        }
        return false
    }
    
    fe.predicates["tag"] = func(tool ToolInfo, value interface{}) bool {
        if tag, ok := value.(string); ok {
            for _, toolTag := range tool.Tags {
                if toolTag == tag {
                    return true
                }
            }
        }
        return false
    }
    
    fe.predicates["version"] = func(tool ToolInfo, value interface{}) bool {
        if version, ok := value.(string); ok {
            return tool.Version == version
        }
        return false
    }
    
    fe.predicates["has_capability"] = func(tool ToolInfo, value interface{}) bool {
        if capability, ok := value.(string); ok {
            if tool.Metadata != nil {
                caps := tool.Metadata.Capabilities
                switch capability {
                case "async":
                    return caps.Async
                case "streaming":
                    return caps.Streaming
                case "batching":
                    return caps.Batching
                case "cancellable":
                    return caps.Cancellable
                case "stateful":
                    return caps.Stateful
                }
            }
        }
        return false
    }
    
    fe.predicates["author"] = func(tool ToolInfo, value interface{}) bool {
        if author, ok := value.(string); ok {
            if tool.Metadata != nil {
                return tool.Metadata.Author == author
            }
        }
        return false
    }
}

// ApplyFilters applies multiple filters to tools
func (fe *FilterEngine) ApplyFilters(tools []ToolInfo, filters []Filter) []ToolInfo {
    if len(filters) == 0 {
        return tools
    }
    
    var filtered []ToolInfo
    
    for _, tool := range tools {
        matches := true
        
        for _, filter := range filters {
            predicate, exists := fe.predicates[filter.Field]
            if !exists {
                continue // Skip unknown predicates
            }
            
            if !predicate(tool, filter.Value) {
                matches = false
                break
            }
        }
        
        if matches {
            filtered = append(filtered, tool)
        }
    }
    
    return filtered
}

type Filter struct {
    Field    string      `json:"field"`
    Operator string      `json:"operator"`
    Value    interface{} `json:"value"`
}
```

## Dynamic Loading and Hot Reloading

### Dynamic Tool Loader

```go
// DynamicToolLoader handles runtime loading and reloading of tools
type DynamicToolLoader struct {
    registry    ToolRegistry
    discovery   ToolDiscovery
    watcher     *FileWatcher
    reloader    *ToolReloader
    config      LoaderConfig
    mu          sync.RWMutex
}

type LoaderConfig struct {
    WatchPaths      []string      `yaml:"watch_paths" json:"watch_paths"`
    HotReload       bool          `yaml:"hot_reload" json:"hot_reload"`
    ReloadDelay     time.Duration `yaml:"reload_delay" json:"reload_delay"`
    ValidateOnLoad  bool          `yaml:"validate_on_load" json:"validate_on_load"`
    BackupOnReload  bool          `yaml:"backup_on_reload" json:"backup_on_reload"`
    MaxRetries      int           `yaml:"max_retries" json:"max_retries"`
}

// NewDynamicToolLoader creates a new dynamic tool loader
func NewDynamicToolLoader(registry ToolRegistry, discovery ToolDiscovery, config LoaderConfig) *DynamicToolLoader {
    return &DynamicToolLoader{
        registry:  registry,
        discovery: discovery,
        watcher:   NewFileWatcher(),
        reloader:  NewToolReloader(),
        config:    config,
    }
}

// StartWatching begins watching for tool changes
func (dl *DynamicToolLoader) StartWatching(ctx context.Context) error {
    if !dl.config.HotReload {
        return nil
    }
    
    // Start file watcher
    for _, path := range dl.config.WatchPaths {
        if err := dl.watcher.Watch(path, dl.handleFileChange); err != nil {
            return fmt.Errorf("failed to watch path %s: %w", path, err)
        }
    }
    
    return dl.watcher.Start(ctx)
}

// handleFileChange processes file system changes
func (dl *DynamicToolLoader) handleFileChange(event FileEvent) {
    switch event.Type {
    case FileEventModified, FileEventCreated:
        dl.handleToolUpdate(event.Path)
    case FileEventDeleted:
        dl.handleToolRemoval(event.Path)
    }
}

// handleToolUpdate handles tool updates
func (dl *DynamicToolLoader) handleToolUpdate(path string) {
    // Debounce rapid changes
    time.Sleep(dl.config.ReloadDelay)
    
    // Rediscover tools from the updated path
    source := NewFileSystemDiscoverySource(filepath.Dir(path), FileSystemSourceConfig{
        BasePath: filepath.Dir(path),
        SearchPatterns: []string{"*.so", "*.dll", "*.dylib", "*.tool"},
}
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    discoveredTools, err := source.Discover(ctx, DiscoveryOptions{
        Validate: dl.config.ValidateOnLoad,
}
    if err != nil {
        logError("tool discovery failed for path %s: %v", path, err)
        return
    }
    
    // Reload tools
    for _, discoveredTool := range discoveredTools {
        if err := dl.reloadTool(discoveredTool); err != nil {
            logError("tool reload failed for %s: %v", discoveredTool.Metadata.Name, err)
        }
    }
}

// reloadTool reloads a specific tool
func (dl *DynamicToolLoader) reloadTool(discoveredTool DiscoveredTool) error {
    dl.mu.Lock()
    defer dl.mu.Unlock()
    
    toolName := discoveredTool.Metadata.Name
    
    // Backup existing tool if configured
    if dl.config.BackupOnReload {
        if existing, err := dl.registry.Get(toolName); err == nil {
            if err := dl.reloader.BackupTool(existing); err != nil {
                logWarning("failed to backup tool %s: %v", toolName, err)
            }
        }
    }
    
    // Unregister existing tool
    if err := dl.registry.Unregister(toolName); err != nil {
        logWarning("failed to unregister existing tool %s: %v", toolName, err)
    }
    
    // Register new tool
    if err := dl.registry.Register(discoveredTool.Tool); err != nil {
        // Try to restore from backup
        if dl.config.BackupOnReload {
            if backup, restoreErr := dl.reloader.RestoreTool(toolName); restoreErr == nil {
                dl.registry.Register(backup)
            }
        }
        return fmt.Errorf("failed to register reloaded tool: %w", err)
    }
    
    logInfo("successfully reloaded tool %s", toolName)
    return nil
}

type FileEvent struct {
    Type FileEventType `json:"type"`
    Path string        `json:"path"`
    Time time.Time     `json:"time"`
}

type FileEventType string

const (
    FileEventCreated  FileEventType = "created"
    FileEventModified FileEventType = "modified"
    FileEventDeleted  FileEventType = "deleted"
)
```

This comprehensive tool discovery documentation covers all aspects of the Go-LLMs discovery system, from basic source implementations to advanced search capabilities and dynamic loading patterns. The system provides flexible, extensible mechanisms for finding, registering, and managing tools across diverse environments and deployment scenarios.