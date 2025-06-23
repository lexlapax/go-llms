// ABOUTME: Output parser interface and registry for handling structured LLM outputs
// ABOUTME: Provides parsing with recovery for malformed outputs and format detection

package outputs

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// Parser defines the interface for parsing LLM outputs.
// Implementations handle specific formats (JSON, XML, YAML) and provide
// recovery mechanisms for malformed outputs, schema-guided parsing,
// and format detection capabilities.
type Parser interface {
	// Name returns the parser name (e.g., "json", "xml", "yaml")
	Name() string

	// Parse attempts to parse the output into a structured format
	Parse(ctx context.Context, output string) (interface{}, error)

	// ParseWithRecovery attempts to parse with error recovery
	ParseWithRecovery(ctx context.Context, output string, opts *RecoveryOptions) (interface{}, error)

	// ParseWithSchema parses output guided by a schema
	ParseWithSchema(ctx context.Context, output string, schema *OutputSchema) (interface{}, error)

	// CanParse checks if the parser can handle the given output
	CanParse(output string) bool
}

// RecoveryOptions configures the recovery behavior.
// These options control how parsers attempt to recover from
// malformed or partially correct LLM outputs, enabling more
// robust parsing in real-world scenarios.
type RecoveryOptions struct {
	// ExtractFromMarkdown attempts to extract content from markdown code blocks
	ExtractFromMarkdown bool

	// FixCommonIssues attempts to fix common formatting issues
	FixCommonIssues bool

	// StrictMode disables all recovery attempts
	StrictMode bool

	// MaxAttempts sets the maximum number of recovery attempts
	MaxAttempts int

	// Schema provides schema guidance for recovery
	Schema *OutputSchema
}

// DefaultRecoveryOptions returns the default recovery options.
// Provides sensible defaults for recovery behavior including
// markdown extraction and common issue fixes.
//
// Returns a configured RecoveryOptions instance.
func DefaultRecoveryOptions() *RecoveryOptions {
	return &RecoveryOptions{
		ExtractFromMarkdown: true,
		FixCommonIssues:     true,
		StrictMode:          false,
		MaxAttempts:         3,
	}
}

// ParserRegistry manages available parsers
type ParserRegistry struct {
	mu      sync.RWMutex
	parsers map[string]Parser
}

// globalRegistry is the global parser registry
var globalRegistry = &ParserRegistry{
	parsers: make(map[string]Parser),
}

// Register adds a parser to the registry
func (r *ParserRegistry) Register(parser Parser) error {
	if parser == nil {
		return errors.New("parser cannot be nil")
	}

	name := parser.Name()
	if name == "" {
		return errors.New("parser name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.parsers[name]; exists {
		return fmt.Errorf("parser %q already registered", name)
	}

	r.parsers[name] = parser
	return nil
}

// Get retrieves a parser by name
func (r *ParserRegistry) Get(name string) (Parser, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	parser, exists := r.parsers[name]
	if !exists {
		return nil, fmt.Errorf("parser %q not found", name)
	}

	return parser, nil
}

// List returns all registered parser names
func (r *ParserRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.parsers))
	for name := range r.parsers {
		names = append(names, name)
	}
	return names
}

// AutoDetect attempts to detect the appropriate parser for the output
func (r *ParserRegistry) AutoDetect(output string) (Parser, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, parser := range r.parsers {
		if parser.CanParse(output) {
			return parser, nil
		}
	}

	return nil, errors.New("no suitable parser found for output")
}

// RegisterParser registers a parser in the global registry
func RegisterParser(parser Parser) error {
	return globalRegistry.Register(parser)
}

// GetParser retrieves a parser from the global registry
func GetParser(name string) (Parser, error) {
	return globalRegistry.Get(name)
}

// ListParsers returns all registered parser names
func ListParsers() []string {
	return globalRegistry.List()
}

// AutoDetectParser attempts to detect the appropriate parser
func AutoDetectParser(output string) (Parser, error) {
	return globalRegistry.AutoDetect(output)
}

// ParseResult represents the result of parsing
type ParseResult struct {
	// Data is the parsed data
	Data interface{}

	// Format is the detected or used format
	Format string

	// RecoveryAttempts is the number of recovery attempts made
	RecoveryAttempts int

	// Warnings contains any warnings during parsing
	Warnings []string
}

// ParseWithAutoDetection attempts to parse output with automatic format detection
func ParseWithAutoDetection(ctx context.Context, output string, opts *RecoveryOptions) (*ParseResult, error) {
	parser, err := AutoDetectParser(output)
	if err != nil {
		return nil, fmt.Errorf("auto-detection failed: %w", err)
	}

	if opts == nil {
		opts = DefaultRecoveryOptions()
	}

	data, err := parser.ParseWithRecovery(ctx, output, opts)
	if err != nil {
		return nil, err
	}

	return &ParseResult{
		Data:   data,
		Format: parser.Name(),
	}, nil
}
