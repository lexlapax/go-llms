// ABOUTME: Script-based workflow steps for embedding scripts in workflows
// ABOUTME: Provides support for multiple scripting languages in workflow execution

package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ScriptStep represents a script-based workflow step
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

// ScriptHandler executes scripts in a specific language
type ScriptHandler interface {
	// Execute runs the script with the given context and state
	Execute(ctx context.Context, state *WorkflowState, script string, env map[string]interface{}) (*WorkflowState, error)
	// Language returns the language this handler supports
	Language() string
	// Validate checks if a script is valid
	Validate(script string) error
}

// ScriptHandlerFunc is a function adapter for ScriptHandler
type ScriptHandlerFunc struct {
	ExecuteFn  func(ctx context.Context, state *WorkflowState, script string, env map[string]interface{}) (*WorkflowState, error)
	LanguageFn func() string
	ValidateFn func(script string) error
}

// Execute implements ScriptHandler
func (f *ScriptHandlerFunc) Execute(ctx context.Context, state *WorkflowState, script string, env map[string]interface{}) (*WorkflowState, error) {
	if f.ExecuteFn != nil {
		return f.ExecuteFn(ctx, state, script, env)
	}
	return nil, fmt.Errorf("execute function not implemented")
}

// Language implements ScriptHandler
func (f *ScriptHandlerFunc) Language() string {
	if f.LanguageFn != nil {
		return f.LanguageFn()
	}
	return ""
}

// Validate implements ScriptHandler
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

// RegisterScriptHandler registers a handler for a scripting language
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

// UnregisterScriptHandler removes a script handler
func UnregisterScriptHandler(language string) {
	globalScriptHandlers.mu.Lock()
	defer globalScriptHandlers.mu.Unlock()
	delete(globalScriptHandlers.handlers, language)
}

// GetScriptHandler retrieves a handler for a language
func GetScriptHandler(language string) (ScriptHandler, bool) {
	globalScriptHandlers.mu.RLock()
	defer globalScriptHandlers.mu.RUnlock()
	handler, exists := globalScriptHandlers.handlers[language]
	return handler, exists
}

// ListScriptLanguages returns all registered script languages
func ListScriptLanguages() []string {
	globalScriptHandlers.mu.RLock()
	defer globalScriptHandlers.mu.RUnlock()

	languages := make([]string, 0, len(globalScriptHandlers.handlers))
	for lang := range globalScriptHandlers.handlers {
		languages = append(languages, lang)
	}
	return languages
}

// NewScriptStep creates a new script-based workflow step
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

// ScriptStepBuilder provides a fluent interface for building script steps
type ScriptStepBuilder struct {
	step *ScriptStep
	err  error
}

// NewScriptStepBuilder creates a new builder
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

// WithLanguage sets the script language
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

// WithScript sets the script source
func (b *ScriptStepBuilder) WithScript(script string) *ScriptStepBuilder {
	if b.err != nil {
		return b
	}
	b.step.script = script
	return b
}

// WithDescription sets the step description
func (b *ScriptStepBuilder) WithDescription(description string) *ScriptStepBuilder {
	if b.err != nil {
		return b
	}
	b.step.description = description
	return b
}

// WithEnvironment adds environment variables
func (b *ScriptStepBuilder) WithEnvironment(key string, value interface{}) *ScriptStepBuilder {
	if b.err != nil {
		return b
	}
	b.step.environment[key] = value
	return b
}

// WithTimeout sets the execution timeout
func (b *ScriptStepBuilder) WithTimeout(timeout time.Duration) *ScriptStepBuilder {
	if b.err != nil {
		return b
	}
	b.step.timeout = timeout
	return b
}

// WithMetadata adds metadata
func (b *ScriptStepBuilder) WithMetadata(key string, value interface{}) *ScriptStepBuilder {
	if b.err != nil {
		return b
	}
	b.step.metadata[key] = value
	return b
}

// Build creates the script step
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

// Name implements WorkflowStep
func (s *ScriptStep) Name() string {
	return s.name
}

// Execute implements WorkflowStep
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

// Validate implements WorkflowStep
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

// Language returns the script language
func (s *ScriptStep) Language() string {
	return s.language
}

// Script returns the script source
func (s *ScriptStep) Script() string {
	return s.script
}

// Description returns the step description
func (s *ScriptStep) Description() string {
	return s.description
}

// Environment returns the environment variables
func (s *ScriptStep) Environment() map[string]interface{} {
	env := make(map[string]interface{})
	for k, v := range s.environment {
		env[k] = v
	}
	return env
}

// Timeout returns the execution timeout
func (s *ScriptStep) Timeout() time.Duration {
	return s.timeout
}

// Metadata returns the step metadata
func (s *ScriptStep) Metadata() map[string]interface{} {
	meta := make(map[string]interface{})
	for k, v := range s.metadata {
		meta[k] = v
	}
	return meta
}
