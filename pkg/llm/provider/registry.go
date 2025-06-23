// ABOUTME: Dynamic provider registry for runtime provider registration
// ABOUTME: Supports provider factories, templates, and hot-reload capabilities

package provider

import (
	"fmt"
	"sync"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
)

// ProviderFactory creates provider instances from configuration
type ProviderFactory interface {
	// CreateProvider creates a new provider instance
	CreateProvider(config map[string]interface{}) (domain.Provider, error)
	// ValidateConfig validates the configuration
	ValidateConfig(config map[string]interface{}) error
	// GetTemplate returns a configuration template
	GetTemplate() ProviderTemplate
}

// ProviderTemplate defines a template for provider configuration
type ProviderTemplate struct {
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Schema      ConfigSchema           `json:"schema"`
	Defaults    map[string]interface{} `json:"defaults"`
	Examples    []ConfigExample        `json:"examples"`
}

// ProviderRegistration contains provider registration information
type ProviderRegistration struct {
	Type     string                 `json:"type"`
	Provider domain.Provider        `json:"-"`
	Factory  ProviderFactory        `json:"-"`
	Metadata ProviderMetadata       `json:"-"`
	Config   map[string]interface{} `json:"config,omitempty"`
	Active   bool                   `json:"active"`
}

// DynamicRegistry extends ModelRegistry with dynamic provider capabilities
type DynamicRegistry struct {
	mu            sync.RWMutex
	providers     map[string]*ProviderRegistration
	factories     map[string]ProviderFactory
	listeners     []RegistryListener
	defaultModels map[string]string // Maps model names to provider types
}

// RegistryListener defines callbacks for monitoring registry changes.
// Implementations can react to provider registration, updates, and removal events.
type RegistryListener interface {
	// OnProviderRegistered is called when a provider is registered
	OnProviderRegistered(name string, provider domain.Provider)
	// OnProviderUnregistered is called when a provider is unregistered
	OnProviderUnregistered(name string)
	// OnProviderUpdated is called when a provider is updated
	OnProviderUpdated(name string, provider domain.Provider)
}

// NewDynamicRegistry creates a new dynamic registry for managing LLM providers.
// The registry starts empty and providers can be registered using RegisterProvider
// or created dynamically using RegisterFactory.
func NewDynamicRegistry() *DynamicRegistry {
	return &DynamicRegistry{
		providers:     make(map[string]*ProviderRegistration),
		factories:     make(map[string]ProviderFactory),
		defaultModels: make(map[string]string),
		listeners:     make([]RegistryListener, 0),
	}
}

// RegisterFactory registers a provider factory with the specified type name.
// The factory will be used to create provider instances from configuration.
// Returns an error if the type is empty or factory is nil.
func (r *DynamicRegistry) RegisterFactory(providerType string, factory ProviderFactory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if providerType == "" {
		return fmt.Errorf("provider type cannot be empty")
	}
	if factory == nil {
		return fmt.Errorf("factory cannot be nil")
	}

	r.factories[providerType] = factory
	return nil
}

// RegisterProvider registers a provider instance with the given name.
// The provider is immediately available for use. Metadata is optional but recommended
// for capability discovery. Returns an error if name is empty or provider is nil.
func (r *DynamicRegistry) RegisterProvider(name string, provider domain.Provider, metadata ProviderMetadata) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	registration := &ProviderRegistration{
		Type:     name,
		Provider: provider,
		Metadata: metadata,
		Active:   true,
	}

	r.providers[name] = registration

	// Notify listeners
	for _, listener := range r.listeners {
		listener.OnProviderRegistered(name, provider)
	}

	return nil
}

// CreateProviderFromTemplate creates and registers a new provider using a factory.
// The providerType must match a registered factory, and the config is validated
// before provider creation. The created provider is registered under the given name.
func (r *DynamicRegistry) CreateProviderFromTemplate(providerType string, name string, config map[string]interface{}) error {
	r.mu.RLock()
	factory, exists := r.factories[providerType]
	r.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no factory registered for provider type: %s", providerType)
	}

	// Validate configuration
	if err := factory.ValidateConfig(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Create provider
	provider, err := factory.CreateProvider(config)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	// Get metadata if provider supports it
	var metadata ProviderMetadata
	if metaProvider, ok := provider.(MetadataProvider); ok {
		metadata = metaProvider.GetMetadata()
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	registration := &ProviderRegistration{
		Type:     providerType,
		Provider: provider,
		Factory:  factory,
		Metadata: metadata,
		Config:   config,
		Active:   true,
	}

	r.providers[name] = registration

	// Notify listeners
	for _, listener := range r.listeners {
		listener.OnProviderRegistered(name, provider)
	}

	return nil
}

// UnregisterProvider removes a provider from the registry.
// Returns an error if the provider doesn't exist. Notifies all registered
// listeners about the removal.
func (r *DynamicRegistry) UnregisterProvider(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.providers[name]; !exists {
		return fmt.Errorf("provider not found: %s", name)
	}

	delete(r.providers, name)

	// Notify listeners
	for _, listener := range r.listeners {
		listener.OnProviderUnregistered(name)
	}

	return nil
}

// GetProvider retrieves a provider by name.
// Returns an error if the provider doesn't exist or is marked as inactive.
func (r *DynamicRegistry) GetProvider(name string) (domain.Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	registration, exists := r.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider not found: %s", name)
	}

	if !registration.Active {
		return nil, fmt.Errorf("provider is inactive: %s", name)
	}

	return registration.Provider, nil
}

// GetMetadata retrieves provider metadata for the specified provider.
// Returns an error if the provider doesn't exist or has no metadata associated.
func (r *DynamicRegistry) GetMetadata(name string) (ProviderMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	registration, exists := r.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider not found: %s", name)
	}

	if registration.Metadata == nil {
		return nil, fmt.Errorf("provider has no metadata: %s", name)
	}

	return registration.Metadata, nil
}

// ListProviders returns all registered provider names.
// The returned slice contains names of both active and inactive providers.
func (r *DynamicRegistry) ListProviders() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// ListProvidersByCapability returns providers that support a specific capability.
// Only providers with metadata that includes the specified capability are returned.
func (r *DynamicRegistry) ListProvidersByCapability(capability Capability) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var names []string
	for name, reg := range r.providers {
		if reg.Metadata != nil && HasCapability(reg.Metadata, capability) {
			names = append(names, name)
		}
	}
	return names
}

// GetTemplate returns the configuration template for a provider type.
// The template includes schema, defaults, and examples for provider configuration.
func (r *DynamicRegistry) GetTemplate(providerType string) (ProviderTemplate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	factory, exists := r.factories[providerType]
	if !exists {
		return ProviderTemplate{}, fmt.Errorf("no factory registered for provider type: %s", providerType)
	}

	return factory.GetTemplate(), nil
}

// ListTemplates returns all available provider templates from registered factories.
// Templates provide configuration schemas and examples for each provider type.
func (r *DynamicRegistry) ListTemplates() []ProviderTemplate {
	r.mu.RLock()
	defer r.mu.RUnlock()

	templates := make([]ProviderTemplate, 0, len(r.factories))
	for _, factory := range r.factories {
		templates = append(templates, factory.GetTemplate())
	}
	return templates
}

// AddListener adds a registry listener to receive notifications about provider changes.
// Listeners are notified when providers are registered, updated, or removed.
func (r *DynamicRegistry) AddListener(listener RegistryListener) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.listeners = append(r.listeners, listener)
}

// RemoveListener removes a previously added registry listener.
// The listener will no longer receive notifications about registry changes.
func (r *DynamicRegistry) RemoveListener(listener RegistryListener) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, l := range r.listeners {
		if l == listener {
			r.listeners = append(r.listeners[:i], r.listeners[i+1:]...)
			break
		}
	}
}

// UpdateProvider updates a provider configuration and recreates it using its factory.
// This only works for providers created through factories. The new configuration
// is validated before the provider is recreated. Listeners are notified of the update.
func (r *DynamicRegistry) UpdateProvider(name string, config map[string]interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	registration, exists := r.providers[name]
	if !exists {
		return fmt.Errorf("provider not found: %s", name)
	}

	if registration.Factory == nil {
		return fmt.Errorf("provider was not created from a factory: %s", name)
	}

	// Validate new configuration
	if err := registration.Factory.ValidateConfig(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Create new provider instance
	newProvider, err := registration.Factory.CreateProvider(config)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	// Update metadata if supported
	var metadata ProviderMetadata
	if metaProvider, ok := newProvider.(MetadataProvider); ok {
		metadata = metaProvider.GetMetadata()
	}

	// Update registration
	registration.Provider = newProvider
	registration.Metadata = metadata
	registration.Config = config

	// Notify listeners
	for _, listener := range r.listeners {
		listener.OnProviderUpdated(name, newProvider)
	}

	return nil
}

// ModelRegistry implementation

// RegisterModel implements domain.ModelRegistry.
// This method provides compatibility with the ModelRegistry interface
// by delegating to RegisterProvider without metadata.
func (r *DynamicRegistry) RegisterModel(name string, provider domain.Provider) error {
	// For compatibility, register as a provider without metadata
	return r.RegisterProvider(name, provider, nil)
}

// GetModel implements domain.ModelRegistry.
// This method provides compatibility with the ModelRegistry interface
// by delegating to GetProvider.
func (r *DynamicRegistry) GetModel(name string) (domain.Provider, error) {
	return r.GetProvider(name)
}

// ListModels implements domain.ModelRegistry.
// This method provides compatibility with the ModelRegistry interface
// by delegating to ListProviders.
func (r *DynamicRegistry) ListModels() []string {
	return r.ListProviders()
}

// ExportConfig exports the registry configuration to a map.
// The exported configuration includes all providers that were created
// from templates with their configuration and active status.
func (r *DynamicRegistry) ExportConfig() (map[string]interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	config := make(map[string]interface{})
	providers := make(map[string]interface{})

	for name, reg := range r.providers {
		if reg.Config != nil {
			providers[name] = map[string]interface{}{
				"type":   reg.Type,
				"config": reg.Config,
				"active": reg.Active,
			}
		}
	}

	config["providers"] = providers
	config["version"] = "1.0"

	return config, nil
}

// ImportConfig imports a registry configuration from a map.
// The configuration should contain a "providers" key with provider
// definitions. Invalid providers are skipped without failing the import.
func (r *DynamicRegistry) ImportConfig(config map[string]interface{}) error {
	providers, ok := config["providers"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid config format: missing providers")
	}

	for name, providerConfig := range providers {
		cfg, ok := providerConfig.(map[string]interface{})
		if !ok {
			continue
		}

		providerType, _ := cfg["type"].(string)
		providerCfg, _ := cfg["config"].(map[string]interface{})

		if providerType != "" && providerCfg != nil {
			if err := r.CreateProviderFromTemplate(providerType, name, providerCfg); err != nil {
				// Log error but continue with other providers
				continue
			}
		}
	}

	return nil
}

// globalRegistry is the singleton instance used throughout the application.
// It is initialized with an empty registry and can be populated using
// RegisterDefaultFactories or custom factory registrations.
var globalRegistry = NewDynamicRegistry()

// GetGlobalRegistry returns the singleton global provider registry.
// This registry is used by default throughout the application for provider management.
func GetGlobalRegistry() *DynamicRegistry {
	return globalRegistry
}
