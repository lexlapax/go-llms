// ABOUTME: Script-based tool factory for dynamic tool creation from scripting engines
// ABOUTME: Enables go-llmspell and other bridges to register tools written in JavaScript, Lua, Tengo, etc.

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"sync"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// ScriptEngine represents a scripting engine type.
// It identifies the language or runtime used to execute script-based tools.
type ScriptEngine string

const (
	ScriptEngineJavaScript ScriptEngine = "javascript"
	ScriptEngineLua        ScriptEngine = "lua"
	ScriptEngineTengo      ScriptEngine = "tengo"
	ScriptEngineExpr       ScriptEngine = "expr"
	ScriptEnginePython     ScriptEngine = "python"
)

// ScriptToolDefinition defines a tool implemented in a scripting language.
// It contains all necessary information to create and execute a script-based tool,
// including the script code, schemas, and metadata.
type ScriptToolDefinition struct {
	Name            string               `json:"name"`
	Description     string               `json:"description"`
	Category        string               `json:"category"`
	Tags            []string             `json:"tags"`
	Version         string               `json:"version"`
	Engine          ScriptEngine         `json:"engine"`
	Script          string               `json:"script"`
	ParameterSchema *schemaDomain.Schema `json:"parameter_schema,omitempty"`
	OutputSchema    *schemaDomain.Schema `json:"output_schema,omitempty"`
	Examples        []domain.ToolExample `json:"examples,omitempty"`
	Constraints     []string             `json:"constraints,omitempty"`
	ErrorGuidance   map[string]string    `json:"error_guidance,omitempty"`
	Context         map[string]any       `json:"context,omitempty"` // Additional context for script execution
}

// ScriptHandler defines the interface for executing scripts in different languages.
// Implementations provide language-specific script execution and validation.
type ScriptHandler interface {
	// Execute runs the script with given context, state, and parameters
	Execute(ctx context.Context, script string, toolCtx *domain.ToolContext, params any) (any, error)

	// Validate checks if the script is valid for this engine
	Validate(script string) error

	// Engine returns the script engine type this handler supports
	Engine() ScriptEngine

	// SupportsFeature checks if the handler supports a specific feature
	SupportsFeature(feature string) bool
}

// ScriptToolFactory creates tools from script definitions.
// It manages script handlers for different engines and provides
// a unified interface for creating script-based tools.
type ScriptToolFactory struct {
	handlers map[ScriptEngine]ScriptHandler
	mu       sync.RWMutex
}

// NewScriptToolFactory creates a new script tool factory.
// The factory starts with no handlers; they must be registered separately.
//
// Returns a new ScriptToolFactory instance.
func NewScriptToolFactory() *ScriptToolFactory {
	return &ScriptToolFactory{
		handlers: make(map[ScriptEngine]ScriptHandler),
	}
}

// RegisterScriptHandler registers a handler for a specific scripting engine.
// Only one handler per engine type is allowed.
//
// Parameters:
//   - handler: The script handler to register
//
// Returns an error if handler is nil or engine already has a handler.
func (f *ScriptToolFactory) RegisterScriptHandler(handler ScriptHandler) error {
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	engine := handler.Engine()
	if _, exists := f.handlers[engine]; exists {
		return fmt.Errorf("handler for engine %s already registered", engine)
	}

	f.handlers[engine] = handler
	return nil
}

// GetHandler returns the handler for a specific engine.
//
// Parameters:
//   - engine: The script engine type
//
// Returns the handler and true if found, or nil and false if not.
func (f *ScriptToolFactory) GetHandler(engine ScriptEngine) (ScriptHandler, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	handler, exists := f.handlers[engine]
	return handler, exists
}

// CreateTool creates a domain.Tool from a script definition.
// It validates the script and creates a tool instance that executes
// the script using the appropriate handler.
//
// Parameters:
//   - def: The script tool definition
//
// Returns the created tool or an error if creation fails.
func (f *ScriptToolFactory) CreateTool(def ScriptToolDefinition) (domain.Tool, error) {
	f.mu.RLock()
	handler, exists := f.handlers[def.Engine]
	f.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no handler registered for engine %s", def.Engine)
	}

	// Validate the script
	if err := handler.Validate(def.Script); err != nil {
		return nil, fmt.Errorf("script validation failed: %w", err)
	}

	// Create the tool instance
	tool := &scriptTool{
		definition: def,
		handler:    handler,
	}

	return tool, nil
}

// CreateToolFactory creates a ToolFactory function from a script definition.
// This enables lazy tool creation for the discovery system.
//
// Parameters:
//   - def: The script tool definition
//
// Returns a ToolFactory that creates the script tool on demand.
func (f *ScriptToolFactory) CreateToolFactory(def ScriptToolDefinition) ToolFactory {
	return func() (domain.Tool, error) {
		return f.CreateTool(def)
	}
}

// SupportedEngines returns all supported script engines.
// These are the engines that have registered handlers.
//
// Returns a slice of supported engine types.
func (f *ScriptToolFactory) SupportedEngines() []ScriptEngine {
	f.mu.RLock()
	defer f.mu.RUnlock()

	engines := make([]ScriptEngine, 0, len(f.handlers))
	for engine := range f.handlers {
		engines = append(engines, engine)
	}
	return engines
}

// scriptTool implements domain.Tool for script-based tools.
// It delegates execution to the appropriate script handler.
type scriptTool struct {
	definition ScriptToolDefinition
	handler    ScriptHandler
}

// Core Tool interface implementation
func (t *scriptTool) Name() string {
	return t.definition.Name
}

func (t *scriptTool) Description() string {
	return t.definition.Description
}

func (t *scriptTool) Execute(ctx *domain.ToolContext, params any) (any, error) {
	// Execute the script using the registered handler
	return t.handler.Execute(ctx.Context, t.definition.Script, ctx, params)
}

// Schema definitions
func (t *scriptTool) ParameterSchema() *schemaDomain.Schema {
	return t.definition.ParameterSchema
}

func (t *scriptTool) OutputSchema() *schemaDomain.Schema {
	return t.definition.OutputSchema
}

// LLM guidance
func (t *scriptTool) UsageInstructions() string {
	if t.definition.Description != "" {
		return t.definition.Description
	}
	return "A script-based tool. Check examples for usage patterns."
}

func (t *scriptTool) Examples() []domain.ToolExample {
	return t.definition.Examples
}

func (t *scriptTool) Constraints() []string {
	constraints := make([]string, len(t.definition.Constraints))
	copy(constraints, t.definition.Constraints)

	// Add script engine constraint
	constraints = append(constraints, fmt.Sprintf("Implemented in %s", t.definition.Engine))

	return constraints
}

func (t *scriptTool) ErrorGuidance() map[string]string {
	guidance := make(map[string]string)
	maps.Copy(guidance, t.definition.ErrorGuidance)

	// Add default script error guidance
	if guidance["script_error"] == "" {
		guidance["script_error"] = fmt.Sprintf("Script execution failed. Check %s syntax and runtime environment.", t.definition.Engine)
	}

	return guidance
}

// Metadata
func (t *scriptTool) Category() string {
	if t.definition.Category != "" {
		return t.definition.Category
	}
	return "script"
}

func (t *scriptTool) Tags() []string {
	tags := make([]string, len(t.definition.Tags))
	copy(tags, t.definition.Tags)

	// Add engine tag
	tags = append(tags, string(t.definition.Engine))

	return tags
}

func (t *scriptTool) Version() string {
	if t.definition.Version != "" {
		return t.definition.Version
	}
	return "1.0.0"
}

// Behavioral hints
func (t *scriptTool) IsDeterministic() bool {
	// Script tools are generally non-deterministic unless explicitly marked
	return false
}

func (t *scriptTool) IsDestructive() bool {
	// Script tools could be destructive - be conservative
	return true
}

func (t *scriptTool) RequiresConfirmation() bool {
	// Script tools should require confirmation by default for safety
	return true
}

func (t *scriptTool) EstimatedLatency() string {
	// Script execution is typically medium latency
	return "medium"
}

// MCP compatibility
func (t *scriptTool) ToMCPDefinition() domain.MCPToolDefinition {
	annotations := make(map[string]any)
	annotations["engine"] = string(t.definition.Engine)
	annotations["script_length"] = len(t.definition.Script)

	// Add context if available
	for k, v := range t.definition.Context {
		annotations[fmt.Sprintf("context_%s", k)] = v
	}

	return domain.MCPToolDefinition{
		Name:         t.definition.Name,
		Description:  t.definition.Description,
		InputSchema:  t.definition.ParameterSchema,
		OutputSchema: t.definition.OutputSchema,
		Annotations:  annotations,
	}
}

// Global script factory instance
var (
	globalScriptFactory *ScriptToolFactory
	scriptFactoryOnce   sync.Once
)

// GetScriptFactory returns the global script tool factory instance.
// It uses a singleton pattern to ensure consistent handler registration.
//
// Returns the global ScriptToolFactory.
func GetScriptFactory() *ScriptToolFactory {
	scriptFactoryOnce.Do(func() {
		globalScriptFactory = NewScriptToolFactory()
	})
	return globalScriptFactory
}

// RegisterScriptHandler registers a script handler globally.
// This is required for downstream scripting engine integration.
//
// Parameters:
//   - handler: The script handler to register
//
// Returns an error if registration fails.
func RegisterScriptHandler(handler ScriptHandler) error {
	return GetScriptFactory().RegisterScriptHandler(handler)
}

// CreateScriptTool creates a tool from a script definition using the global factory.
// This is a convenience function for creating script tools.
//
// Parameters:
//   - def: The script tool definition
//
// Returns the created tool or an error.
func CreateScriptTool(def ScriptToolDefinition) (domain.Tool, error) {
	return GetScriptFactory().CreateTool(def)
}

// RegisterScriptToolWithDiscovery creates and registers a script tool with the discovery system.
// This enables script tools to be discovered and created dynamically.
//
// Parameters:
//   - def: The script tool definition
//
// Returns an error if registration fails.
func RegisterScriptToolWithDiscovery(def ScriptToolDefinition) error {
	// Create tool factory
	factory := GetScriptFactory().CreateToolFactory(def)

	// Create ToolInfo for discovery
	toolInfo := ToolInfo{
		Name:        def.Name,
		Description: def.Description,
		Category:    def.Category,
		Tags:        append(def.Tags, string(def.Engine)),
		Version:     def.Version,
		UsageHint:   fmt.Sprintf("Script-based tool implemented in %s", def.Engine),
		Package:     "script",
	}

	// Marshal schemas for ToolInfo
	if def.ParameterSchema != nil {
		if paramBytes, err := json.Marshal(def.ParameterSchema); err == nil {
			toolInfo.ParameterSchema = paramBytes
		}
	}

	if def.OutputSchema != nil {
		if outputBytes, err := json.Marshal(def.OutputSchema); err == nil {
			toolInfo.OutputSchema = outputBytes
		}
	}

	// Convert examples
	for _, ex := range def.Examples {
		example := Example{
			Name:        ex.Name,
			Description: ex.Description,
		}

		if inputBytes, err := json.Marshal(ex.Input); err == nil {
			example.Input = inputBytes
		}

		if outputBytes, err := json.Marshal(ex.Output); err == nil {
			example.Output = outputBytes
		}

		toolInfo.Examples = append(toolInfo.Examples, example)
	}

	// Register with discovery system
	discovery := NewDiscovery()
	return discovery.RegisterTool(toolInfo, factory)
}
