package domain

import "time"

// Metadata holds information about the inventory list itself.
type Metadata struct {
	// Version of the inventory list format or data structure.
	Version string `json:"version"`
	// LastUpdated indicates when the inventory data was last generated or updated,
	// typically in "YYYY-MM-DD" format.
	LastUpdated string `json:"last_updated"`
	// Description provides a brief overview of the inventory content.
	Description string `json:"description"`
	// SchemaVersion indicates the version of the schema this inventory conforms to.
	SchemaVersion string `json:"schema_version"`
}

// Pricing details for a model, typically per 1,000 tokens.
type Pricing struct {
	// InputPer1kTokens is the cost for 1,000 input tokens.
	InputPer1kTokens float64 `json:"input_per_1k_tokens"`
	// OutputPer1kTokens is the cost for 1,000 output tokens.
	OutputPer1kTokens float64 `json:"output_per_1k_tokens"`
}

// MediaTypeCapability defines whether a model can read (process as input)
// or write (generate as output) a specific media type.
type MediaTypeCapability struct {
	// Read indicates if the model can process this media type as input.
	Read bool `json:"read"`
	// Write indicates if the model can generate this media type as output.
	Write bool `json:"write"`
}

// Capabilities defines the various functional and media-handling capabilities of an LLM model.
type Capabilities struct {
	// Text capability for processing and generating plain text.
	Text MediaTypeCapability `json:"text"`
	// Image capability for processing and generating images.
	Image MediaTypeCapability `json:"image"`
	// Audio capability for processing and generating audio.
	Audio MediaTypeCapability `json:"audio"`
	// Video capability for processing and generating video.
	Video MediaTypeCapability `json:"video"`
	// File capability for processing arbitrary files (e.g., PDFs, documents) as input.
	// This usually implies reading content from files rather than generating whole files as output.
	File MediaTypeCapability `json:"file"`
	// FunctionCalling indicates if the model supports tool use or function calling.
	FunctionCalling bool `json:"function_calling"`
	// Streaming indicates if the model can stream its responses token by token.
	Streaming bool `json:"streaming"`
	// JSONMode indicates if the model supports a constrained output mode ensuring valid JSON.
	JSONMode bool `json:"json_mode"`
}

// Model represents a single LLM model's detailed information.
type Model struct {
	// Provider is the name of the LLM provider (e.g., "openai", "google", "anthropic").
	Provider string `json:"provider"`
	// Name is the unique identifier or API name of the model (e.g., "gpt-4-turbo", "gemini-1.5-pro-latest").
	Name string `json:"name"`
	// DisplayName is a human-readable name for the model (e.g., "GPT-4 Turbo", "Gemini 1.5 Pro").
	DisplayName string `json:"display_name"`
	// Description provides a brief overview of the model's features and use cases.
	Description string `json:"description"`
	// DocumentationURL links to the official documentation for the model.
	DocumentationURL string `json:"documentation_url"`
	// Capabilities lists the functional and media-handling abilities of the model.
	Capabilities Capabilities `json:"capabilities"`
	// ContextWindow is the maximum number of tokens the model can process in a single request (input + output).
	ContextWindow int `json:"context_window"`
	// MaxOutputTokens is the maximum number of tokens the model can generate in a single response.
	MaxOutputTokens int `json:"max_output_tokens"`
	// TrainingCutoff is the date up to which the model has been trained,
	// typically in "YYYY-MM" or "YYYY-MM-DD" format (e.g., "2023-09").
	TrainingCutoff string `json:"training_cutoff"`
	// ModelFamily groups related models (e.g., "gpt-4", "claude-3", "gemini").
	ModelFamily string `json:"model_family"`
	// Pricing details for using the model.
	Pricing Pricing `json:"pricing"`
	// LastUpdated indicates when this model's information was last updated or verified,
	// typically in "YYYY-MM-DD" format. This can also be the model's release or update date.
	LastUpdated string `json:"last_updated"`
}

// ModelInventory is the top-level structure for the aggregated model inventory.
type ModelInventory struct {
	// Metadata contains information about the inventory list itself.
	Metadata Metadata `json:"_metadata"`
	// Models is a list of all available models from the aggregated providers.
	Models []Model `json:"models"`
}

// CachedModelInventory wraps ModelInventory with a timestamp for caching purposes.
// This structure is used when saving or loading the inventory from a cache.
type CachedModelInventory struct {
	// Inventory holds the actual model inventory data.
	Inventory ModelInventory `json:"inventory"`
	// FetchedAt is the timestamp indicating when this inventory data was fetched and cached.
	FetchedAt time.Time `json:"fetched_at"`
}
