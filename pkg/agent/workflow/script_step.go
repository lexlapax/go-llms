// ABOUTME: Script-based workflow steps for embedding scripts in workflows
// ABOUTME: Provides support for multiple scripting languages in workflow execution

package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ScriptStep represents a script-based workflow step.
// It executes scripts in various languages as part of a workflow,
// with support for environment variables, timeouts, and metadata.
type ScriptStep struct {
	name        string
	description string
	language    string
	script      string
	handler     ScriptHandler
	environment map[string]interface{}
	timeout     time.Duration
	metadata    map[string]interface{}
}

// ScriptHandler executes scripts in a specific language.
// Implementations provide language-specific script execution,
// validation, and environment integration.
type ScriptHandler interface {
	// Execute runs the script with the given context and state
	Execute(ctx context.Context, state *WorkflowState, script string, env map[string]interface{}) (*WorkflowState, error)
	// Language returns the language this handler supports
	Language() string
	// Validate checks if a script is valid
	Validate(script string) error
}

// ScriptHandlerFunc is a function adapter for ScriptHandler.
// It allows using functions as ScriptHandler implementations,
// useful for creating lightweight handlers without full structs.
type ScriptHandlerFunc struct {
	ExecuteFn  func(ctx context.Context, state *WorkflowState, script string, env map[string]interface{}) (*WorkflowState, error)
	LanguageFn func() string
	ValidateFn func(script string) error
}

// Execute implements ScriptHandler.
// It delegates to the ExecuteFn if provided.
func (f *ScriptHandlerFunc) Execute(ctx context.Context, state *WorkflowState, script string, env map[string]interface{}) (*WorkflowState, error) {
	if f.ExecuteFn != nil {
		return f.ExecuteFn(ctx, state, script, env)
	}
	return nil, fmt.Errorf("execute function not implemented")
}

// Language implements ScriptHandler.
// It delegates to the LanguageFn if provided.
func (f *ScriptHandlerFunc) Language() string {
	if f.LanguageFn != nil {
		return f.LanguageFn()
	}
	return ""
}

// Validate implements ScriptHandler.
// It delegates to the ValidateFn if provided.
func (f *ScriptHandlerFunc) Validate(script string) error {
	if f.ValidateFn != nil {
		return f.ValidateFn(script)
	}
	return nil
}

// scriptHandlerRegistry manages script handlers
type scriptHandlerRegistry struct {
	mu       sync.RWMutex
	handlers map[string]ScriptHandler
}

// globalScriptHandlers is the global script handler registry
var globalScriptHandlers = &scriptHandlerRegistry{
	handlers: make(map[string]ScriptHandler),
}

// RegisterScriptHandler registers a handler for a scripting language.
// Only one handler per language is allowed; registering a new handler
// for an existing language will replace the previous one.
//
// Parameters:
//   - language: The language identifier (e.g., "javascript", "python")
//   - handler: The handler implementation for the language
//
// Returns an error if language is empty or handler is nil.
func RegisterScriptHandler(language string, handler ScriptHandler) error {
	globalScriptHandlers.mu.Lock()
	defer globalScriptHandlers.mu.Unlock()

	if language == "" {
		return fmt.Errorf("language cannot be empty")
	}
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	globalScriptHandlers.handlers[language] = handler
	return nil
}

// UnregisterScriptHandler removes a script handler.
// If the language doesn't exist, this is a no-op.
//
// Parameters:
//   - language: The language identifier to unregister
func UnregisterScriptHandler(language string) {
	globalScriptHandlers.mu.Lock()
	defer globalScriptHandlers.mu.Unlock()
	delete(globalScriptHandlers.handlers, language)
}

// GetScriptHandler retrieves a handler for a language.
//
// Parameters:
//   - language: The language identifier to look up
//
// Returns the handler and true if found, nil and false otherwise.
func GetScriptHandler(language string) (ScriptHandler, bool) {
	globalScriptHandlers.mu.RLock()
	defer globalScriptHandlers.mu.RUnlock()
	handler, exists := globalScriptHandlers.handlers[language]
	return handler, exists
}

// ListScriptLanguages returns all registered script languages.
// The returned slice contains language identifiers in no particular order.
//
// Returns a slice of registered language identifiers.
func ListScriptLanguages() []string {
	globalScriptHandlers.mu.RLock()
	defer globalScriptHandlers.mu.RUnlock()

	languages := make([]string, 0, len(globalScriptHandlers.handlers))
	for lang := range globalScriptHandlers.handlers {
		languages = append(languages, lang)
	}
	return languages
}

// NewScriptStep creates a new script-based workflow step.
// The step is initialized with a default timeout of 30 seconds.
//
// Parameters:
//   - name: The name of the step
//   - language: The scripting language to use
//   - script: The script source code
//
// Returns a new ScriptStep or an error if no handler exists for the language.
func NewScriptStep(name, language, script string) (*ScriptStep, error) {
	handler, exists := GetScriptHandler(language)
	if !exists {
		return nil, fmt.Errorf("no handler registered for language: %s", language)
	}

	return &ScriptStep{
		name:        name,
		language:    language,
		script:      script,
		handler:     handler,
		environment: make(map[string]interface{}),
		timeout:     30 * time.Second, // Default timeout
		metadata:    make(map[string]interface{}),
	}, nil
}

// ScriptStepBuilder provides a fluent interface for building script steps.
// It accumulates configuration and validates the final step before creation.
type ScriptStepBuilder struct {
	step *ScriptStep
	err  error
}

// NewScriptStepBuilder creates a new builder.
// The builder starts with a default timeout of 30 seconds.
//
// Parameters:
//   - name: The name of the script step
//
// Returns a new ScriptStepBuilder instance.
func NewScriptStepBuilder(name string) *ScriptStepBuilder {
	return &ScriptStepBuilder{
		step: &ScriptStep{
			name:        name,
			environment: make(map[string]interface{}),
			metadata:    make(map[string]interface{}),
			timeout:     30 * time.Second,
		},
	}
}

// WithLanguage sets the script language.
// It also retrieves and sets the appropriate handler.
//
// Parameters:
//   - language: The scripting language identifier
//
// Returns the builder for method chaining.
func (b *ScriptStepBuilder) WithLanguage(language string) *ScriptStepBuilder {
	if b.err != nil {
		return b
	}

	handler, exists := GetScriptHandler(language)
	if !exists {
		b.err = fmt.Errorf("no handler registered for language: %s", language)
		return b
	}

	b.step.language = language
	b.step.handler = handler
	return b
}

// WithScript sets the script source.
//
// Parameters:
//   - script: The script source code
//
// Returns the builder for method chaining.
func (b *ScriptStepBuilder) WithScript(script string) *ScriptStepBuilder {
	if b.err != nil {
		return b
	}
	b.step.script = script
	return b
}

// WithDescription sets the step description.
//
// Parameters:
//   - description: Human-readable description of the step
//
// Returns the builder for method chaining.
func (b *ScriptStepBuilder) WithDescription(description string) *ScriptStepBuilder {
	if b.err != nil {
		return b
	}
	b.step.description = description
	return b
}

// WithEnvironment adds environment variables.
// Variables are made available to the script during execution.
//
// Parameters:
//   - key: The environment variable name
//   - value: The environment variable value
//
// Returns the builder for method chaining.
func (b *ScriptStepBuilder) WithEnvironment(key string, value interface{}) *ScriptStepBuilder {
	if b.err != nil {
		return b
	}
	b.step.environment[key] = value
	return b
}

// WithTimeout sets the execution timeout.
// Scripts that exceed this timeout will be canceled.
//
// Parameters:
//   - timeout: Maximum execution duration
//
// Returns the builder for method chaining.
func (b *ScriptStepBuilder) WithTimeout(timeout time.Duration) *ScriptStepBuilder {
	if b.err != nil {
		return b
	}
	b.step.timeout = timeout
	return b
}

// WithMetadata adds metadata.
// Metadata is preserved across the step execution for tracking purposes.
//
// Parameters:
//   - key: The metadata key
//   - value: The metadata value
//
// Returns the builder for method chaining.
func (b *ScriptStepBuilder) WithMetadata(key string, value interface{}) *ScriptStepBuilder {
	if b.err != nil {
		return b
	}
	b.step.metadata[key] = value
	return b
}

// Build creates the script step.
// It validates that language and script are provided, and that
// the script is valid according to the language handler.
//
// Returns the configured ScriptStep or an error if validation fails.
func (b *ScriptStepBuilder) Build() (*ScriptStep, error) {
	if b.err != nil {
		return nil, b.err
	}

	if b.step.language == "" {
		return nil, fmt.Errorf("language is required")
	}
	if b.step.script == "" {
		return nil, fmt.Errorf("script is required")
	}
	if b.step.handler == nil {
		return nil, fmt.Errorf("no handler available for language: %s", b.step.language)
	}

	// Validate the script
	if err := b.step.handler.Validate(b.step.script); err != nil {
		return nil, fmt.Errorf("script validation failed: %w", err)
	}

	return b.step, nil
}

// Name implements WorkflowStep.
// Returns the name of the script step.
func (s *ScriptStep) Name() string {
	return s.name
}

// Execute implements WorkflowStep.
// It runs the script with the configured handler, environment, and timeout.
// The current state values and metadata are made available to the script.
//
// Parameters:
//   - ctx: The execution context
//   - state: The current workflow state
//
// Returns the new workflow state or an error.
func (s *ScriptStep) Execute(ctx context.Context, state *WorkflowState) (*WorkflowState, error) {
	// Create a context with timeout
	if s.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.timeout)
		defer cancel()
	}

	// Prepare environment with state data
	env := make(map[string]interface{})
	for k, v := range s.environment {
		env[k] = v
	}

	// Add state data to environment
	if state.State != nil {
		env["state"] = state.Values()
	}
	env["metadata"] = state.Metadata

	// Execute the script
	result, err := s.handler.Execute(ctx, state, s.script, env)
	if err != nil {
		return nil, fmt.Errorf("script execution failed: %w", err)
	}

	// Add execution metadata
	if result.Metadata == nil {
		result.Metadata = make(map[string]interface{})
	}
	result.Metadata[fmt.Sprintf("script_%s_executed", s.name)] = time.Now()
	result.Metadata[fmt.Sprintf("script_%s_language", s.name)] = s.language

	return result, nil
}

// Validate implements WorkflowStep.
// It checks that all required fields are set and validates
// the script with the language handler.
//
// Returns an error if validation fails.
func (s *ScriptStep) Validate() error {
	if s.name == "" {
		return fmt.Errorf("step name cannot be empty")
	}
	if s.language == "" {
		return fmt.Errorf("script language cannot be empty")
	}
	if s.script == "" {
		return fmt.Errorf("script cannot be empty")
	}
	if s.handler == nil {
		return fmt.Errorf("no handler available for language: %s", s.language)
	}

	// Validate the script with the handler
	return s.handler.Validate(s.script)
}

// Language returns the script language.
// This identifies which handler will execute the script.
func (s *ScriptStep) Language() string {
	return s.language
}

// Script returns the script source.
// This is the actual code that will be executed.
func (s *ScriptStep) Script() string {
	return s.script
}

// Description returns the step description.
// This provides human-readable information about what the script does.
func (s *ScriptStep) Description() string {
	return s.description
}

// Environment returns the environment variables.
// Returns a copy to prevent external modifications.
func (s *ScriptStep) Environment() map[string]interface{} {
	env := make(map[string]interface{})
	for k, v := range s.environment {
		env[k] = v
	}
	return env
}

// Timeout returns the execution timeout.
// Zero means no timeout is enforced.
func (s *ScriptStep) Timeout() time.Duration {
	return s.timeout
}

// Metadata returns the step metadata.
// Returns a copy to prevent external modifications.
func (s *ScriptStep) Metadata() map[string]interface{} {
	meta := make(map[string]interface{})
	for k, v := range s.metadata {
		meta[k] = v
	}
	return meta
}
