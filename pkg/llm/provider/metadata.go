// ABOUTME: Provider metadata system for capability discovery and configuration
// ABOUTME: Provides interfaces and types for LLM provider metadata and registration

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/lexlapax/go-llms/pkg/util/llmutil/modelinfo"
	"github.com/lexlapax/go-llms/pkg/util/llmutil/modelinfo/domain"
)

// Capability represents a specific feature or functionality that an LLM provider supports.
// Capabilities are used for runtime feature discovery and provider selection.
type Capability string

const (
	// CapabilityStreaming indicates the provider supports streaming responses
	CapabilityStreaming Capability = "streaming"
	// CapabilityFunctionCalling indicates the provider supports function/tool calling
	CapabilityFunctionCalling Capability = "function_calling"
	// CapabilityVision indicates the provider supports vision/image inputs
	CapabilityVision Capability = "vision"
	// CapabilityEmbeddings indicates the provider supports generating embeddings
	CapabilityEmbeddings Capability = "embeddings"
	// CapabilityAudio indicates the provider supports audio inputs
	CapabilityAudio Capability = "audio"
	// CapabilityVideo indicates the provider supports video inputs
	CapabilityVideo Capability = "video"
	// CapabilityStructuredOutput indicates the provider supports structured output with schemas
	CapabilityStructuredOutput Capability = "structured_output"
)

// ModelInfo describes a specific LLM model's characteristics and capabilities.
// It includes pricing, context limits, supported features, and deprecation status.
type ModelInfo struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description,omitempty"`
	Capabilities    []Capability           `json:"capabilities"`
	ContextWindow   int                    `json:"context_window"`
	MaxTokens       int                    `json:"max_tokens,omitempty"`
	InputPricing    *PricingInfo           `json:"input_pricing,omitempty"`
	OutputPricing   *PricingInfo           `json:"output_pricing,omitempty"`
	Deprecated      bool                   `json:"deprecated"`
	DeprecationDate *time.Time             `json:"deprecation_date,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// PricingInfo represents the cost structure for using an LLM model.
// Prices are specified per token count (e.g., per 1K or 1M tokens) in the given currency.
type PricingInfo struct {
	Currency  string  `json:"currency"`
	PerTokens int     `json:"per_tokens"` // Price per N tokens (usually 1K or 1M)
	Price     float64 `json:"price"`
}

// RateLimit specifies the usage limits imposed by an LLM provider.
// Limits can be defined for requests and tokens over different time periods.
type RateLimit struct {
	RequestsPerMinute int `json:"requests_per_minute,omitempty"`
	RequestsPerHour   int `json:"requests_per_hour,omitempty"`
	RequestsPerDay    int `json:"requests_per_day,omitempty"`
	TokensPerMinute   int `json:"tokens_per_minute,omitempty"`
	TokensPerHour     int `json:"tokens_per_hour,omitempty"`
	TokensPerDay      int `json:"tokens_per_day,omitempty"`
}

// Constraints describes operational limits and requirements for a provider.
// This includes rate limits, concurrency restrictions, and regional availability.
type Constraints struct {
	MaxBatchSize    int                    `json:"max_batch_size,omitempty"`
	MaxConcurrency  int                    `json:"max_concurrency,omitempty"`
	RateLimit       *RateLimit             `json:"rate_limit,omitempty"`
	RequiredHeaders []string               `json:"required_headers,omitempty"`
	AllowedRegions  []string               `json:"allowed_regions,omitempty"`
	MinRequestDelay time.Duration          `json:"min_request_delay,omitempty"`
	MaxRetries      int                    `json:"max_retries,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// ConfigField describes a single configuration parameter for a provider.
// It includes type information, validation rules, and default values.
type ConfigField struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"` // string, number, boolean, object, array
	Description string                 `json:"description"`
	Required    bool                   `json:"required"`
	Default     interface{}            `json:"default,omitempty"`
	Validation  string                 `json:"validation,omitempty"` // Simple validation expression
	Options     []interface{}          `json:"options,omitempty"`    // For enum-like fields
	Secret      bool                   `json:"secret"`               // For API keys and passwords
	EnvVar      string                 `json:"env_var,omitempty"`    // Environment variable mapping
	Properties  map[string]ConfigField `json:"properties,omitempty"` // For object types
}

// ConfigSchema defines the complete configuration structure for a provider.
// It includes field definitions, validation rules, and usage examples.
type ConfigSchema struct {
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Fields      map[string]ConfigField `json:"fields"`
	Examples    []ConfigExample        `json:"examples,omitempty"`
}

// ConfigExample provides a concrete example of provider configuration.
// Examples help users understand how to configure providers correctly.
type ConfigExample struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config"`
}

// ProviderMetadata defines the interface for accessing provider capabilities and configuration.
// Implementations provide runtime information about models, features, and constraints.
type ProviderMetadata interface {
	// Name returns the provider name
	Name() string

	// Description returns a human-readable description
	Description() string

	// GetCapabilities returns the provider's capabilities
	GetCapabilities() []Capability

	// GetModels returns available models dynamically
	GetModels(ctx context.Context) ([]ModelInfo, error)

	// GetConstraints returns provider constraints
	GetConstraints() Constraints

	// GetConfigSchema returns the configuration schema
	GetConfigSchema() ConfigSchema
}

// MetadataProvider indicates that a provider can supply metadata about its capabilities.
// Providers implementing this interface enable dynamic feature discovery.
type MetadataProvider interface {
	// GetMetadata returns the provider's metadata
	GetMetadata() ProviderMetadata
}

// BaseProviderMetadata provides a reusable implementation of the ProviderMetadata interface.
// It includes caching for model information and can be embedded in provider implementations.
type BaseProviderMetadata struct {
	ProviderName        string       `json:"name"`
	ProviderDescription string       `json:"description"`
	Capabilities        []Capability `json:"capabilities"`
	ProviderConstraints Constraints  `json:"constraints"`
	Schema              ConfigSchema `json:"config_schema"`

	// For dynamic model loading
	modelService  *modelinfo.ModelInfoService
	providerType  string // "openai", "anthropic", etc. for filtering
	cachedModels  []ModelInfo
	cacheMutex    sync.RWMutex
	cacheExpiry   time.Time
	cacheDuration time.Duration
}

// Name implements ProviderMetadata
func (m *BaseProviderMetadata) Name() string {
	return m.ProviderName
}

// Description implements ProviderMetadata
func (m *BaseProviderMetadata) Description() string {
	return m.ProviderDescription
}

// GetCapabilities implements ProviderMetadata
func (m *BaseProviderMetadata) GetCapabilities() []Capability {
	return m.Capabilities
}

// GetModels implements ProviderMetadata with dynamic loading
func (m *BaseProviderMetadata) GetModels(ctx context.Context) ([]ModelInfo, error) {
	// Check cache first
	m.cacheMutex.RLock()
	if time.Now().Before(m.cacheExpiry) && len(m.cachedModels) > 0 {
		models := m.cachedModels
		m.cacheMutex.RUnlock()
		return models, nil
	}
	m.cacheMutex.RUnlock()

	// Need to fetch fresh data
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	// Double-check after acquiring write lock
	if time.Now().Before(m.cacheExpiry) && len(m.cachedModels) > 0 {
		return m.cachedModels, nil
	}

	// If no model service configured, return empty
	if m.modelService == nil {
		return []ModelInfo{}, nil
	}

	// Fetch from modelinfo service
	inventory, err := m.modelService.AggregateModels()
	if err != nil {
		// Return cached data if available, even if expired
		if len(m.cachedModels) > 0 {
			return m.cachedModels, fmt.Errorf("failed to fetch fresh models, returning cached: %w", err)
		}
		return nil, fmt.Errorf("failed to fetch models: %w", err)
	}

	// Filter and convert models for this provider
	var models []ModelInfo
	for _, model := range inventory.Models {
		if model.Provider == m.providerType {
			models = append(models, convertDomainModelToModelInfo(model))
		}
	}

	// Update cache
	m.cachedModels = models
	m.cacheExpiry = time.Now().Add(m.cacheDuration)

	return models, nil
}

// GetConstraints implements ProviderMetadata
func (m *BaseProviderMetadata) GetConstraints() Constraints {
	return m.ProviderConstraints
}

// GetConfigSchema implements ProviderMetadata
func (m *BaseProviderMetadata) GetConfigSchema() ConfigSchema {
	return m.Schema
}

// MarshalJSON implements json.Marshaler
func (m *BaseProviderMetadata) MarshalJSON() ([]byte, error) {
	// For JSON serialization, we'll include cached models if available
	return json.Marshal(struct {
		Name         string       `json:"name"`
		Description  string       `json:"description"`
		Capabilities []Capability `json:"capabilities"`
		Models       []ModelInfo  `json:"models,omitempty"`
		Constraints  Constraints  `json:"constraints"`
		ConfigSchema ConfigSchema `json:"config_schema"`
	}{
		Name:         m.ProviderName,
		Description:  m.ProviderDescription,
		Capabilities: m.Capabilities,
		Models:       m.cachedModels, // Use cached models for serialization
		Constraints:  m.ProviderConstraints,
		ConfigSchema: m.Schema,
	})
}

// HasCapability checks if a provider supports a specific capability.
// It returns true if the capability is found in the provider's capability list.
func HasCapability(metadata ProviderMetadata, capability Capability) bool {
	for _, c := range metadata.GetCapabilities() {
		if c == capability {
			return true
		}
	}
	return false
}

// FindModelByID searches for a model with the given ID in the provider's model list.
// It returns the model info, a boolean indicating if found, and any error that occurred.
func FindModelByID(metadata ProviderMetadata, modelID string, ctx context.Context) (*ModelInfo, bool, error) {
	models, err := metadata.GetModels(ctx)
	if err != nil {
		return nil, false, err
	}

	for _, model := range models {
		if model.ID == modelID {
			return &model, true, nil
		}
	}
	return nil, false, nil
}

// convertDomainModelToModelInfo converts modelinfo domain model to our ModelInfo
func convertDomainModelToModelInfo(dm domain.Model) ModelInfo {
	// Convert capabilities
	var caps []Capability
	if dm.Capabilities.Streaming {
		caps = append(caps, CapabilityStreaming)
	}
	if dm.Capabilities.FunctionCalling {
		caps = append(caps, CapabilityFunctionCalling)
	}
	if dm.Capabilities.Image.Read {
		caps = append(caps, CapabilityVision)
	}
	if dm.Capabilities.JSONMode {
		caps = append(caps, CapabilityStructuredOutput)
	}

	return ModelInfo{
		ID:            dm.Name,
		Name:          dm.DisplayName,
		Description:   dm.Description,
		Capabilities:  caps,
		ContextWindow: dm.ContextWindow,
		MaxTokens:     dm.MaxOutputTokens,
		InputPricing: &PricingInfo{
			Currency:  "USD",
			PerTokens: 1000, // modelinfo uses per 1k tokens
			Price:     dm.Pricing.InputPer1kTokens,
		},
		OutputPricing: &PricingInfo{
			Currency:  "USD",
			PerTokens: 1000,
			Price:     dm.Pricing.OutputPer1kTokens,
		},
		// Could parse dm.LastUpdated to check if deprecated
		Deprecated: false,
	}
}

// NewBaseProviderMetadata creates a new BaseProviderMetadata instance with model caching support.
// The providerType parameter determines which models are fetched from the model service.
func NewBaseProviderMetadata(
	name, description, providerType string,
	capabilities []Capability,
	constraints Constraints,
	schema ConfigSchema,
	modelService *modelinfo.ModelInfoService,
) *BaseProviderMetadata {
	return &BaseProviderMetadata{
		ProviderName:        name,
		ProviderDescription: description,
		Capabilities:        capabilities,
		ProviderConstraints: constraints,
		Schema:              schema,
		modelService:        modelService,
		providerType:        providerType,
		cacheDuration:       5 * time.Minute, // Default cache duration
	}
}
