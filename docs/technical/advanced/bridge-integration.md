# Bridge Integration: Scripting Engine Integration

> **[Project Root](/) / [Documentation](/docs/) / [Technical Documentation](/docs/technical/) / [Advanced Topics](/docs/technical/advanced/) / Bridge Integration**

Comprehensive guide to scripting engine integration in Go-LLMs, covering bridge architecture, script execution environments, language-specific bridges, security isolation, script lifecycle management, and advanced scripting patterns for extending LLM applications with dynamic code execution.

## Bridge Architecture

### 1. Core Bridge Interfaces

```go
// ScriptBridge provides the interface for script execution environments
type ScriptBridge interface {
    // Execution
    Execute(ctx context.Context, script Script) (*ScriptResult, error)
    ExecuteWithData(ctx context.Context, script Script, data map[string]interface{}) (*ScriptResult, error)
    
    // Script management
    Compile(source string, options CompileOptions) (CompiledScript, error)
    Validate(script Script) error
    
    // Environment management
    CreateEnvironment(config EnvironmentConfig) (ScriptEnvironment, error)
    DestroyEnvironment(envID string) error
    
    // Capabilities
    GetSupportedLanguages() []string
    GetCapabilities() BridgeCapabilities
    
    // Lifecycle
    Initialize(ctx context.Context, config BridgeConfig) error
    Shutdown(ctx context.Context) error
    
    // Security
    SetSecurityPolicy(policy SecurityPolicy) error
    GetSecurityPolicy() SecurityPolicy
}

// ScriptEnvironment represents an isolated execution environment
type ScriptEnvironment interface {
    // Execution
    Run(ctx context.Context, script CompiledScript, input map[string]interface{}) (*ScriptResult, error)
    RunWithTimeout(ctx context.Context, script CompiledScript, input map[string]interface{}, timeout time.Duration) (*ScriptResult, error)
    
    // State management
    SetGlobal(name string, value interface{}) error
    GetGlobal(name string) (interface{}, error)
    ClearGlobals() error
    
    // Module system
    ImportModule(name string, module ScriptModule) error
    LoadLibrary(name string, library ScriptLibrary) error
    
    // Resource management
    GetMemoryUsage() int64
    GetExecutionTime() time.Duration
    SetResourceLimits(limits ResourceLimits) error
    
    // Cleanup
    Reset() error
    Close() error
}

type Script struct {
    ID          string                 `json:"id"`
    Language    string                 `json:"language"`
    Source      string                 `json:"source"`
    Metadata    ScriptMetadata         `json:"metadata"`
    Parameters  map[string]interface{} `json:"parameters,omitempty"`
    Environment string                 `json:"environment,omitempty"`
}

type ScriptMetadata struct {
    Name        string            `json:"name"`
    Version     string            `json:"version"`
    Description string            `json:"description"`
    Author      string            `json:"author,omitempty"`
    Tags        []string          `json:"tags,omitempty"`
    Dependencies []string         `json:"dependencies,omitempty"`
    Permissions  []string         `json:"permissions,omitempty"`
    Schema      *ScriptSchema     `json:"schema,omitempty"`
}

type ScriptSchema struct {
    Input  *JSONSchema `json:"input,omitempty"`
    Output *JSONSchema `json:"output,omitempty"`
}

type ScriptResult struct {
    Success     bool                   `json:"success"`
    Output      interface{}            `json:"output,omitempty"`
    Error       error                  `json:"error,omitempty"`
    Logs        []LogEntry             `json:"logs,omitempty"`
    Metrics     ExecutionMetrics       `json:"metrics"`
    Environment string                 `json:"environment"`
}

type ExecutionMetrics struct {
    Duration     time.Duration `json:"duration"`
    MemoryUsed   int64         `json:"memory_used"`
    CPUTime      time.Duration `json:"cpu_time"`
    Instructions int64         `json:"instructions,omitempty"`
    ExitCode     int           `json:"exit_code"`
}

type BridgeCapabilities struct {
    Languages       []string          `json:"languages"`
    Sandboxing      bool              `json:"sandboxing"`
    Streaming       bool              `json:"streaming"`
    AsyncExecution  bool              `json:"async_execution"`
    StateManagement bool              `json:"state_management"`
    ModuleSystem    bool              `json:"module_system"`
    Debugging       bool              `json:"debugging"`
    Profiling       bool              `json:"profiling"`
    Features        map[string]bool   `json:"features"`
}

type EnvironmentConfig struct {
    Language      string                 `yaml:"language" json:"language"`
    Sandbox       bool                   `yaml:"sandbox" json:"sandbox"`
    ResourceLimits ResourceLimits        `yaml:"resource_limits" json:"resource_limits"`
    Permissions   []string               `yaml:"permissions" json:"permissions"`
    Modules       []string               `yaml:"modules" json:"modules"`
    Variables     map[string]interface{} `yaml:"variables" json:"variables"`
    WorkingDir    string                 `yaml:"working_dir" json:"working_dir"`
}

type ResourceLimits struct {
    MaxMemory      int64         `yaml:"max_memory" json:"max_memory"`
    MaxCPUTime     time.Duration `yaml:"max_cpu_time" json:"max_cpu_time"`
    MaxDuration    time.Duration `yaml:"max_duration" json:"max_duration"`
    MaxInstructions int64        `yaml:"max_instructions" json:"max_instructions"`
    MaxOutputSize  int64         `yaml:"max_output_size" json:"max_output_size"`
    MaxFileSize    int64         `yaml:"max_file_size" json:"max_file_size"`
}
```

### 2. Bridge Manager

```go
// BridgeManager orchestrates multiple script bridges
type BridgeManager struct {
    bridges     map[string]ScriptBridge
    environments map[string]ScriptEnvironment
    policies    map[string]SecurityPolicy
    config      BridgeManagerConfig
    mu          sync.RWMutex
}

type BridgeManagerConfig struct {
    DefaultLanguage    string                    `yaml:"default_language" json:"default_language"`
    SecurityMode       string                    `yaml:"security_mode" json:"security_mode"`
    GlobalLimits       ResourceLimits            `yaml:"global_limits" json:"global_limits"`
    EnabledLanguages   []string                  `yaml:"enabled_languages" json:"enabled_languages"`
    BridgeConfigs      map[string]BridgeConfig   `yaml:"bridge_configs" json:"bridge_configs"`
}

// NewBridgeManager creates a new bridge manager
func NewBridgeManager(config BridgeManagerConfig) *BridgeManager {
    return &BridgeManager{
        bridges:     make(map[string]ScriptBridge),
        environments: make(map[string]ScriptEnvironment),
        policies:    make(map[string]SecurityPolicy),
        config:      config,
    }
}

// RegisterBridge registers a script bridge for a language
func (bm *BridgeManager) RegisterBridge(language string, bridge ScriptBridge) error {
    bm.mu.Lock()
    defer bm.mu.Unlock()
    
    if _, exists := bm.bridges[language]; exists {
        return fmt.Errorf("bridge for language %s already registered", language)
    }
    
    bm.bridges[language] = bridge
    return nil
}

// ExecuteScript executes a script using the appropriate bridge
func (bm *BridgeManager) ExecuteScript(ctx context.Context, script Script, data map[string]interface{}) (*ScriptResult, error) {
    bridge, err := bm.getBridge(script.Language)
    if err != nil {
        return nil, err
    }
    
    // Apply security policy
    if err := bm.applySecurityPolicy(script); err != nil {
        return nil, fmt.Errorf("security policy violation: %w", err)
    }
    
    // Create or get environment
    env, err := bm.getOrCreateEnvironment(script)
    if err != nil {
        return nil, fmt.Errorf("failed to get environment: %w", err)
    }
    
    // Compile script if needed
    compiled, err := bridge.Compile(script.Source, CompileOptions{
        Optimize: true,
        Debug:    false,
    })
    if err != nil {
        return nil, fmt.Errorf("script compilation failed: %w", err)
    }
    
    // Execute with timeout and resource limits
    result, err := env.RunWithTimeout(ctx, compiled, data, bm.config.GlobalLimits.MaxDuration)
    if err != nil {
        return nil, fmt.Errorf("script execution failed: %w", err)
    }
    
    return result, nil
}

// getBridge retrieves the bridge for a language
func (bm *BridgeManager) getBridge(language string) (ScriptBridge, error) {
    bm.mu.RLock()
    defer bm.mu.RUnlock()
    
    bridge, exists := bm.bridges[language]
    if !exists {
        return nil, fmt.Errorf("no bridge registered for language: %s", language)
    }
    
    return bridge, nil
}

// getOrCreateEnvironment gets or creates a script environment
func (bm *BridgeManager) getOrCreateEnvironment(script Script) (ScriptEnvironment, error) {
    envKey := fmt.Sprintf("%s_%s", script.Language, script.Environment)
    
    bm.mu.RLock()
    if env, exists := bm.environments[envKey]; exists {
        bm.mu.RUnlock()
        return env, nil
    }
    bm.mu.RUnlock()
    
    bm.mu.Lock()
    defer bm.mu.Unlock()
    
    // Double-check after acquiring write lock
    if env, exists := bm.environments[envKey]; exists {
        return env, nil
    }
    
    // Create new environment
    bridge, err := bm.getBridge(script.Language)
    if err != nil {
        return nil, err
    }
    
    envConfig := EnvironmentConfig{
        Language:       script.Language,
        Sandbox:        true,
        ResourceLimits: bm.config.GlobalLimits,
        Permissions:    script.Metadata.Permissions,
    }
    
    env, err := bridge.CreateEnvironment(envConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to create environment: %w", err)
    }
    
    bm.environments[envKey] = env
    return env, nil
}
```

## Language-Specific Bridges

### 1. Lua Bridge Implementation

```go
// LuaBridge implements script execution for Lua
type LuaBridge struct {
    config     LuaBridgeConfig
    vmPool     *sync.Pool
    libraries  map[string]LuaLibrary
    modules    map[string]LuaModule
    security   SecurityPolicy
}

type LuaBridgeConfig struct {
    MaxMemory       int64         `yaml:"max_memory" json:"max_memory"`
    MaxInstructions int64         `yaml:"max_instructions" json:"max_instructions"`
    EnableDebug     bool          `yaml:"enable_debug" json:"enable_debug"`
    AllowedModules  []string      `yaml:"allowed_modules" json:"allowed_modules"`
    Preload         []string      `yaml:"preload" json:"preload"`
}

// NewLuaBridge creates a new Lua bridge
func NewLuaBridge(config LuaBridgeConfig) *LuaBridge {
    bridge := &LuaBridge{
        config:    config,
        libraries: make(map[string]LuaLibrary),
        modules:   make(map[string]LuaModule),
    }
    
    // Initialize VM pool
    bridge.vmPool = &sync.Pool{
        New: func() interface{} {
            return bridge.createLuaVM()
        },
    }
    
    // Register built-in libraries
    bridge.registerBuiltinLibraries()
    
    return bridge
}

// Execute runs a Lua script
func (lb *LuaBridge) Execute(ctx context.Context, script Script) (*ScriptResult, error) {
    return lb.ExecuteWithData(ctx, script, nil)
}

// ExecuteWithData runs a Lua script with input data
func (lb *LuaBridge) ExecuteWithData(ctx context.Context, script Script, data map[string]interface{}) (*ScriptResult, error) {
    // Get VM from pool
    vm := lb.vmPool.Get().(*lua.LState)
    defer lb.vmPool.Put(vm)
    
    // Reset VM state
    vm.Close()
    vm = lb.createLuaVM()
    
    // Set up execution context
    if err := lb.setupExecutionContext(vm, script, data); err != nil {
        return nil, fmt.Errorf("failed to setup execution context: %w", err)
    }
    
    // Set up cancellation
    done := make(chan struct{})
    go func() {
        select {
        case <-ctx.Done():
            vm.Close() // Force close to interrupt execution
        case <-done:
        }
    }()
    defer close(done)
    
    // Execute script
    start := time.Now()
    result := &ScriptResult{
        Environment: "lua",
        Metrics: ExecutionMetrics{
            Duration: 0,
        },
    }
    
    err := vm.DoString(script.Source)
    result.Metrics.Duration = time.Since(start)
    
    if err != nil {
        result.Success = false
        result.Error = err
        return result, nil
    }
    
    // Extract output
    output, err := lb.extractOutput(vm)
    if err != nil {
        result.Success = false
        result.Error = err
        return result, nil
    }
    
    result.Success = true
    result.Output = output
    
    // Extract logs
    result.Logs = lb.extractLogs(vm)
    
    return result, nil
}

// createLuaVM creates a new Lua virtual machine
func (lb *LuaBridge) createLuaVM() *lua.LState {
    vm := lua.NewState(lua.Options{
        CallStackSize:       1024,
        RegistrySize:        1024,
        SkipOpenLibs:        true, // We'll open libraries selectively
        IncludeGoStackTrace: false,
    })
    
    // Open safe libraries only
    lb.openSafeLibraries(vm)
    
    // Install security hooks
    lb.installSecurityHooks(vm)
    
    return vm
}

// openSafeLibraries opens only safe Lua libraries
func (lb *LuaBridge) openSafeLibraries(vm *lua.LState) {
    // Base library (safe functions only)
    vm.Push(vm.NewFunction(lb.safeBaseLib))
    vm.Push(lua.LString("base"))
    vm.Call(1, 0)
    
    // String library
    luaopen_string(vm)
    
    // Table library
    luaopen_table(vm)
    
    // Math library
    luaopen_math(vm)
    
    // JSON library
    vm.Push(vm.NewFunction(lb.jsonLib))
    vm.Push(lua.LString("json"))
    vm.Call(1, 0)
    
    // HTTP library (with restrictions)
    if lb.security.AllowNetworkAccess {
        vm.Push(vm.NewFunction(lb.httpLib))
        vm.Push(lua.LString("http"))
        vm.Call(1, 0)
    }
}

// setupExecutionContext sets up the execution environment
func (lb *LuaBridge) setupExecutionContext(vm *lua.LState, script Script, data map[string]interface{}) error {
    // Set input data
    if data != nil {
        inputTable := lb.goToLua(vm, data)
        vm.SetGlobal("input", inputTable)
    }
    
    // Set script metadata
    metaTable := vm.NewTable()
    vm.SetField(metaTable, "name", lua.LString(script.Metadata.Name))
    vm.SetField(metaTable, "version", lua.LString(script.Metadata.Version))
    vm.SetGlobal("script", metaTable)
    
    // Set up output capture
    vm.SetGlobal("output", lua.LNil)
    
    // Set up logging
    vm.Push(vm.NewFunction(lb.logFunction))
    vm.SetGlobal("log")
    
    return nil
}

// JavaScript Bridge Implementation
type JavaScriptBridge struct {
    config   JSBridgeConfig
    runtime  *goja.Runtime
    pool     *sync.Pool
    modules  map[string]JSModule
    security SecurityPolicy
}

type JSBridgeConfig struct {
    MaxMemory      int64    `yaml:"max_memory" json:"max_memory"`
    Strict         bool     `yaml:"strict" json:"strict"`
    ES6            bool     `yaml:"es6" json:"es6"`
    AllowedModules []string `yaml:"allowed_modules" json:"allowed_modules"`
}

// NewJavaScriptBridge creates a new JavaScript bridge
func NewJavaScriptBridge(config JSBridgeConfig) *JavaScriptBridge {
    bridge := &JavaScriptBridge{
        config:  config,
        modules: make(map[string]JSModule),
    }
    
    bridge.pool = &sync.Pool{
        New: func() interface{} {
            return bridge.createJSRuntime()
        },
    }
    
    return bridge
}

// Execute runs a JavaScript script
func (jsb *JavaScriptBridge) ExecuteWithData(ctx context.Context, script Script, data map[string]interface{}) (*ScriptResult, error) {
    runtime := jsb.pool.Get().(*goja.Runtime)
    defer jsb.pool.Put(runtime)
    
    // Reset runtime
    runtime = jsb.createJSRuntime()
    
    // Set up context
    if err := jsb.setupJSContext(runtime, script, data); err != nil {
        return nil, fmt.Errorf("failed to setup JS context: %w", err)
    }
    
    // Execute with timeout
    done := make(chan struct{})
    var execErr error
    var execResult goja.Value
    
    go func() {
        defer close(done)
        defer func() {
            if r := recover(); r != nil {
                execErr = fmt.Errorf("script panicked: %v", r)
            }
        }()
        
        execResult, execErr = runtime.RunString(script.Source)
    }()
    
    // Wait for completion or timeout
    select {
    case <-done:
        // Execution completed
    case <-ctx.Done():
        return nil, ctx.Err()
    }
    
    result := &ScriptResult{
        Environment: "javascript",
        Metrics: ExecutionMetrics{
            Duration: time.Since(time.Now()),
        },
    }
    
    if execErr != nil {
        result.Success = false
        result.Error = execErr
        return result, nil
    }
    
    // Convert result
    output, err := jsb.gojaToGo(execResult)
    if err != nil {
        result.Success = false
        result.Error = err
        return result, nil
    }
    
    result.Success = true
    result.Output = output
    
    return result, nil
}

func (jsb *JavaScriptBridge) createJSRuntime() *goja.Runtime {
    runtime := goja.New()
    
    // Set strict mode if enabled
    if jsb.config.Strict {
        runtime.RunString("'use strict';")
    }
    
    // Install security restrictions
    jsb.installJSSecurityHooks(runtime)
    
    // Add safe built-ins
    jsb.addSafeBuiltins(runtime)
    
    return runtime
}
```

### 3. Python Bridge Implementation

```go
// PythonBridge implements script execution for Python
type PythonBridge struct {
    config      PythonBridgeConfig
    interpreter *python.Interpreter
    modules     map[string]PythonModule
    security    SecurityPolicy
    mu          sync.Mutex
}

type PythonBridgeConfig struct {
    PythonPath      []string `yaml:"python_path" json:"python_path"`
    AllowedModules  []string `yaml:"allowed_modules" json:"allowed_modules"`
    VirtualEnv      string   `yaml:"virtual_env" json:"virtual_env"`
    MaxMemory       int64    `yaml:"max_memory" json:"max_memory"`
    Timeout         time.Duration `yaml:"timeout" json:"timeout"`
}

// NewPythonBridge creates a new Python bridge
func NewPythonBridge(config PythonBridgeConfig) *PythonBridge {
    bridge := &PythonBridge{
        config:  config,
        modules: make(map[string]PythonModule),
    }
    
    // Initialize Python interpreter
    bridge.initializePython()
    
    return bridge
}

// ExecuteWithData runs a Python script
func (pb *PythonBridge) ExecuteWithData(ctx context.Context, script Script, data map[string]interface{}) (*ScriptResult, error) {
    pb.mu.Lock()
    defer pb.mu.Unlock()
    
    // Create isolated execution context
    namespace := pb.createNamespace(script, data)
    
    // Set up execution monitoring
    result := &ScriptResult{
        Environment: "python",
        Metrics:     ExecutionMetrics{},
    }
    
    start := time.Now()
    
    // Execute script with timeout
    execChan := make(chan error, 1)
    go func() {
        execChan <- pb.executeInNamespace(script.Source, namespace)
    }()
    
    select {
    case err := <-execChan:
        result.Metrics.Duration = time.Since(start)
        
        if err != nil {
            result.Success = false
            result.Error = err
            return result, nil
        }
        
        // Extract output
        output, err := pb.extractOutput(namespace)
        if err != nil {
            result.Success = false
            result.Error = err
            return result, nil
        }
        
        result.Success = true
        result.Output = output
        
    case <-ctx.Done():
        return nil, ctx.Err()
    }
    
    return result, nil
}

// initializePython sets up the Python interpreter
func (pb *PythonBridge) initializePython() error {
    // Initialize Python
    if !python.Py_IsInitialized() {
        python.Py_Initialize()
    }
    
    // Set up virtual environment if specified
    if pb.config.VirtualEnv != "" {
        if err := pb.setupVirtualEnv(); err != nil {
            return fmt.Errorf("failed to setup virtual environment: %w", err)
        }
    }
    
    // Configure Python path
    for _, path := range pb.config.PythonPath {
        pb.addToSysPath(path)
    }
    
    // Install security restrictions
    pb.installPythonSecurity()
    
    return nil
}

// createNamespace creates an isolated execution namespace
func (pb *PythonBridge) createNamespace(script Script, data map[string]interface{}) map[string]interface{} {
    namespace := make(map[string]interface{})
    
    // Add safe built-ins
    namespace["__builtins__"] = pb.getSafeBuiltins()
    
    // Add input data
    if data != nil {
        namespace["input"] = data
    }
    
    // Add script metadata
    namespace["script"] = map[string]interface{}{
        "name":    script.Metadata.Name,
        "version": script.Metadata.Version,
    }
    
    // Add output placeholder
    namespace["output"] = nil
    
    // Add logging function
    namespace["log"] = pb.createLogFunction()
    
    return namespace
}
```

## Security and Sandboxing

### 1. Security Framework

```go
// SecurityPolicy defines security restrictions for script execution
type SecurityPolicy struct {
    // Access controls
    AllowFileAccess    bool     `yaml:"allow_file_access" json:"allow_file_access"`
    AllowNetworkAccess bool     `yaml:"allow_network_access" json:"allow_network_access"`
    AllowSystemCalls   bool     `yaml:"allow_system_calls" json:"allow_system_calls"`
    AllowEnvironment   bool     `yaml:"allow_environment" json:"allow_environment"`
    
    // File system restrictions
    AllowedPaths       []string `yaml:"allowed_paths" json:"allowed_paths"`
    BlockedPaths       []string `yaml:"blocked_paths" json:"blocked_paths"`
    ReadOnlyPaths      []string `yaml:"read_only_paths" json:"read_only_paths"`
    MaxFileSize        int64    `yaml:"max_file_size" json:"max_file_size"`
    
    // Network restrictions
    AllowedHosts       []string `yaml:"allowed_hosts" json:"allowed_hosts"`
    BlockedHosts       []string `yaml:"blocked_hosts" json:"blocked_hosts"`
    AllowedPorts       []int    `yaml:"allowed_ports" json:"allowed_ports"`
    BlockedPorts       []int    `yaml:"blocked_ports" json:"blocked_ports"`
    
    // Module restrictions
    AllowedModules     []string `yaml:"allowed_modules" json:"allowed_modules"`
    BlockedModules     []string `yaml:"blocked_modules" json:"blocked_modules"`
    AllowedFunctions   []string `yaml:"allowed_functions" json:"allowed_functions"`
    BlockedFunctions   []string `yaml:"blocked_functions" json:"blocked_functions"`
    
    // Resource limits
    ResourceLimits     ResourceLimits `yaml:"resource_limits" json:"resource_limits"`
    
    // Execution restrictions
    AllowInfiniteLoops bool     `yaml:"allow_infinite_loops" json:"allow_infinite_loops"`
    AllowEval          bool     `yaml:"allow_eval" json:"allow_eval"`
    AllowImport        bool     `yaml:"allow_import" json:"allow_import"`
    
    // Custom restrictions
    CustomRules        []SecurityRule `yaml:"custom_rules" json:"custom_rules"`
}

type SecurityRule struct {
    Name        string      `yaml:"name" json:"name"`
    Type        string      `yaml:"type" json:"type"`
    Pattern     string      `yaml:"pattern" json:"pattern"`
    Action      string      `yaml:"action" json:"action"`
    Parameters  interface{} `yaml:"parameters" json:"parameters"`
}

// SecurityEnforcer enforces security policies
type SecurityEnforcer struct {
    policy    SecurityPolicy
    monitor   *SecurityMonitor
    violations chan SecurityViolation
}

type SecurityViolation struct {
    Type        string                 `json:"type"`
    Rule        string                 `json:"rule"`
    Description string                 `json:"description"`
    Context     map[string]interface{} `json:"context"`
    Timestamp   time.Time              `json:"timestamp"`
    Severity    string                 `json:"severity"`
}

// EnforcePolicy applies security policy to script execution
func (se *SecurityEnforcer) EnforcePolicy(script Script) error {
    // Validate script against policy
    if err := se.validateScript(script); err != nil {
        return fmt.Errorf("script validation failed: %w", err)
    }
    
    // Check for dangerous patterns
    if err := se.scanForThreats(script); err != nil {
        return fmt.Errorf("threat scan failed: %w", err)
    }
    
    // Validate module usage
    if err := se.validateModules(script); err != nil {
        return fmt.Errorf("module validation failed: %w", err)
    }
    
    return nil
}

// validateScript performs basic script validation
func (se *SecurityEnforcer) validateScript(script Script) error {
    // Check file access patterns
    if !se.policy.AllowFileAccess {
        patterns := []string{
            `open\s*\(`,
            `file\s*\(`,
            `readfile`,
            `writefile`,
            `os\.open`,
            `fs\.`,
        }
        
        for _, pattern := range patterns {
            if matched, _ := regexp.MatchString(pattern, script.Source); matched {
                return fmt.Errorf("file access not allowed")
            }
        }
    }
    
    // Check network access patterns
    if !se.policy.AllowNetworkAccess {
        patterns := []string{
            `http\.`,
            `fetch\s*\(`,
            `request\s*\(`,
            `socket\.`,
            `urllib`,
            `requests\.`,
        }
        
        for _, pattern := range patterns {
            if matched, _ := regexp.MatchString(pattern, script.Source); matched {
                return fmt.Errorf("network access not allowed")
            }
        }
    }
    
    // Check system call patterns
    if !se.policy.AllowSystemCalls {
        patterns := []string{
            `system\s*\(`,
            `exec\s*\(`,
            `spawn\s*\(`,
            `os\.system`,
            `subprocess\.`,
            `eval\s*\(`,
        }
        
        for _, pattern := range patterns {
            if matched, _ := regexp.MatchString(pattern, script.Source); matched {
                return fmt.Errorf("system calls not allowed")
            }
        }
    }
    
    return nil
}

// Sandbox provides isolated execution environment
type Sandbox struct {
    config     SandboxConfig
    containers map[string]*Container
    monitor    *ResourceMonitor
    mu         sync.RWMutex
}

type SandboxConfig struct {
    Type           string         `yaml:"type" json:"type"` // docker, vm, process
    BaseImage      string         `yaml:"base_image" json:"base_image"`
    ResourceLimits ResourceLimits `yaml:"resource_limits" json:"resource_limits"`
    NetworkMode    string         `yaml:"network_mode" json:"network_mode"`
    ReadOnlyFS     bool           `yaml:"read_only_fs" json:"read_only_fs"`
    TempFS         []string       `yaml:"temp_fs" json:"temp_fs"`
    Capabilities   []string       `yaml:"capabilities" json:"capabilities"`
}

type Container struct {
    ID          string                 `json:"id"`
    Status      string                 `json:"status"`
    CreatedAt   time.Time              `json:"created_at"`
    Environment map[string]interface{} `json:"environment"`
    Limits      ResourceLimits         `json:"limits"`
    Usage       ResourceUsage          `json:"usage"`
}

type ResourceUsage struct {
    CPU        float64 `json:"cpu"`
    Memory     int64   `json:"memory"`
    Disk       int64   `json:"disk"`
    Network    int64   `json:"network"`
    FileOps    int64   `json:"file_ops"`
    NetworkOps int64   `json:"network_ops"`
}

// CreateContainer creates a new sandboxed container
func (s *Sandbox) CreateContainer(config EnvironmentConfig) (*Container, error) {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    containerID := generateContainerID()
    
    container := &Container{
        ID:          containerID,
        Status:      "creating",
        CreatedAt:   time.Now(),
        Environment: config.Variables,
        Limits:      config.ResourceLimits,
    }
    
    // Create container based on sandbox type
    switch s.config.Type {
    case "docker":
        if err := s.createDockerContainer(container, config); err != nil {
            return nil, fmt.Errorf("failed to create docker container: %w", err)
        }
        
    case "process":
        if err := s.createProcessContainer(container, config); err != nil {
            return nil, fmt.Errorf("failed to create process container: %w", err)
        }
        
    default:
        return nil, fmt.Errorf("unsupported sandbox type: %s", s.config.Type)
    }
    
    container.Status = "running"
    s.containers[containerID] = container
    
    return container, nil
}

// ExecuteInContainer runs a script in a sandboxed container
func (s *Sandbox) ExecuteInContainer(ctx context.Context, containerID string, script CompiledScript, input map[string]interface{}) (*ScriptResult, error) {
    s.mu.RLock()
    container, exists := s.containers[containerID]
    s.mu.RUnlock()
    
    if !exists {
        return nil, fmt.Errorf("container %s not found", containerID)
    }
    
    if container.Status != "running" {
        return nil, fmt.Errorf("container %s is not running", containerID)
    }
    
    // Execute script with monitoring
    monitor := s.monitor.StartMonitoring(containerID)
    defer monitor.Stop()
    
    result, err := s.executeInSandbox(ctx, container, script, input)
    
    // Update resource usage
    container.Usage = monitor.GetUsage()
    
    return result, err
}
```

### 2. Advanced Execution Patterns

```go
// ScriptOrchestrator manages complex script execution workflows
type ScriptOrchestrator struct {
    bridgeManager *BridgeManager
    scheduler     *ExecutionScheduler
    dependencies  *DependencyResolver
    cache         *ScriptCache
}

// ExecuteWorkflow executes a workflow of related scripts
func (so *ScriptOrchestrator) ExecuteWorkflow(ctx context.Context, workflow ScriptWorkflow) (*WorkflowResult, error) {
    result := &WorkflowResult{
        WorkflowID: workflow.ID,
        StartTime:  time.Now(),
        Steps:      make([]StepResult, len(workflow.Steps)),
    }
    
    // Resolve dependencies
    executionOrder, err := so.dependencies.ResolveOrder(workflow.Steps)
    if err != nil {
        return nil, fmt.Errorf("dependency resolution failed: %w", err)
    }
    
    // Execute steps in order
    context := make(map[string]interface{})
    
    for i, stepIndex := range executionOrder {
        step := workflow.Steps[stepIndex]
        
        // Prepare step input
        stepInput, err := so.prepareStepInput(step, context)
        if err != nil {
            return nil, fmt.Errorf("failed to prepare input for step %s: %w", step.Name, err)
        }
        
        // Execute step
        stepResult, err := so.executeStep(ctx, step, stepInput)
        if err != nil {
            return nil, fmt.Errorf("step %s failed: %w", step.Name, err)
        }
        
        result.Steps[i] = *stepResult
        
        // Update context with step output
        if step.OutputName != "" {
            context[step.OutputName] = stepResult.Output
        }
        
        // Check for early termination conditions
        if stepResult.ShouldTerminate {
            break
        }
    }
    
    result.EndTime = time.Now()
    result.Duration = result.EndTime.Sub(result.StartTime)
    result.Success = so.evaluateWorkflowSuccess(result.Steps)
    
    return result, nil
}

type ScriptWorkflow struct {
    ID          string         `json:"id"`
    Name        string         `json:"name"`
    Description string         `json:"description"`
    Steps       []WorkflowStep `json:"steps"`
    Metadata    map[string]interface{} `json:"metadata"`
}

type WorkflowStep struct {
    Name         string                 `json:"name"`
    Script       Script                 `json:"script"`
    InputMapping map[string]string      `json:"input_mapping"`
    OutputName   string                 `json:"output_name"`
    Dependencies []string               `json:"dependencies"`
    Condition    string                 `json:"condition,omitempty"`
    OnError      string                 `json:"on_error,omitempty"`
    Timeout      time.Duration          `json:"timeout,omitempty"`
    Parallel     bool                   `json:"parallel,omitempty"`
}

type WorkflowResult struct {
    WorkflowID string        `json:"workflow_id"`
    Success    bool          `json:"success"`
    StartTime  time.Time     `json:"start_time"`
    EndTime    time.Time     `json:"end_time"`
    Duration   time.Duration `json:"duration"`
    Steps      []StepResult  `json:"steps"`
    Context    map[string]interface{} `json:"context"`
    Error      error         `json:"error,omitempty"`
}

type StepResult struct {
    StepName        string          `json:"step_name"`
    Success         bool            `json:"success"`
    Output          interface{}     `json:"output"`
    Error           error           `json:"error,omitempty"`
    Duration        time.Duration   `json:"duration"`
    ShouldTerminate bool            `json:"should_terminate"`
    Metadata        map[string]interface{} `json:"metadata"`
}

// StreamingExecutor provides streaming script execution
type StreamingExecutor struct {
    bridge     ScriptBridge
    bufferSize int
}

// ExecuteStream executes a script with streaming output
func (se *StreamingExecutor) ExecuteStream(ctx context.Context, script Script, input map[string]interface{}) (<-chan StreamChunk, error) {
    outputChan := make(chan StreamChunk, se.bufferSize)
    
    go func() {
        defer close(outputChan)
        
        // Create streaming environment
        env, err := se.bridge.CreateEnvironment(EnvironmentConfig{
            Language: script.Language,
            Sandbox:  true,
        })
        if err != nil {
            outputChan <- StreamChunk{
                Type:  ChunkTypeError,
                Error: err,
            }
            return
        }
        defer env.Close()
        
        // Set up streaming handlers
        se.setupStreamingHandlers(env, outputChan)
        
        // Compile and execute
        compiled, err := se.bridge.Compile(script.Source, CompileOptions{})
        if err != nil {
            outputChan <- StreamChunk{
                Type:  ChunkTypeError,
                Error: err,
            }
            return
        }
        
        result, err := env.Run(ctx, compiled, input)
        if err != nil {
            outputChan <- StreamChunk{
                Type:  ChunkTypeError,
                Error: err,
            }
            return
        }
        
        // Send final result
        outputChan <- StreamChunk{
            Type:   ChunkTypeResult,
            Data:   result.Output,
            Final:  true,
        }
    }()
    
    return outputChan, nil
}

type StreamChunk struct {
    Type      ChunkType   `json:"type"`
    Data      interface{} `json:"data,omitempty"`
    Error     error       `json:"error,omitempty"`
    Timestamp time.Time   `json:"timestamp"`
    Final     bool        `json:"final"`
}

type ChunkType string

const (
    ChunkTypeOutput ChunkType = "output"
    ChunkTypeLog    ChunkType = "log"
    ChunkTypeError  ChunkType = "error"
    ChunkTypeResult ChunkType = "result"
    ChunkTypeMetric ChunkType = "metric"
)
```

This comprehensive bridge integration guide provides the foundation for building powerful, secure, and flexible scripting capabilities within Go-LLMs applications, enabling dynamic code execution while maintaining safety and performance.