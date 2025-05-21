package fetchers

import (
	"github.com/lexlapax/go-llms/pkg/modelinfo/domain"
)

// AnthropicFetcher fetches model information for Anthropic models.
// Currently, it returns a hardcoded list.
type AnthropicFetcher struct {
	// No fields needed for now, but can hold client or config later.
}

// FetchModels retrieves a hardcoded list of Anthropic models.
func (f *AnthropicFetcher) FetchModels() ([]domain.Model, error) {
	models := []domain.Model{
		{
			Provider:         "anthropic",
			Name:             "claude-3-opus-20240229",
			DisplayName:      "Claude 3 Opus",
			Description:      "Anthropic's most powerful model, delivering state-of-the-art performance on highly complex tasks and demonstrating fluency and human-like understanding.",
			DocumentationURL: "https://docs.anthropic.com/claude/docs/models-overview",
			Capabilities: domain.Capabilities{
				Text:            domain.MediaTypeCapability{Read: true, Write: true},
				Image:           domain.MediaTypeCapability{Read: true, Write: false}, // Can process images
				Audio:           domain.MediaTypeCapability{Read: false, Write: false},
				Video:           domain.MediaTypeCapability{Read: false, Write: false},
				File:            domain.MediaTypeCapability{Read: false, Write: false}, // General file processing not a standard capability
				FunctionCalling: true,
				Streaming:       true,
				JSONMode:        true, // Supported via specific prompting techniques
			},
			ContextWindow:    200000,
			MaxOutputTokens:  4096,
			TrainingCutoff:   "2023-08", // Approximate, actual data may vary
			ModelFamily:      "claude-3",
			Pricing: domain.Pricing{ // Example pricing, verify from official source
				InputPer1kTokens:  0.015,
				OutputPer1kTokens: 0.075,
			},
			LastUpdated: "2024-02-29", // Model release date as proxy
		},
		{
			Provider:         "anthropic",
			Name:             "claude-3-sonnet-20240229",
			DisplayName:      "Claude 3 Sonnet",
			Description:      "Anthropic's model balancing intelligence and speed, ideal for enterprise workloads, RAG, and scaled deployments.",
			DocumentationURL: "https://docs.anthropic.com/claude/docs/models-overview",
			Capabilities: domain.Capabilities{
				Text:            domain.MediaTypeCapability{Read: true, Write: true},
				Image:           domain.MediaTypeCapability{Read: true, Write: false}, // Can process images
				Audio:           domain.MediaTypeCapability{Read: false, Write: false},
				Video:           domain.MediaTypeCapability{Read: false, Write: false},
				File:            domain.MediaTypeCapability{Read: false, Write: false},
				FunctionCalling: true,
				Streaming:       true,
				JSONMode:        true, // Supported via specific prompting techniques
			},
			ContextWindow:    200000,
			MaxOutputTokens:  4096, // Can be up to 8192 for specific use cases via API
			TrainingCutoff:   "2023-08", // Approximate
			ModelFamily:      "claude-3",
			Pricing: domain.Pricing{ // Example pricing, verify from official source
				InputPer1kTokens:  0.003,
				OutputPer1kTokens: 0.015,
			},
			LastUpdated: "2024-02-29",
		},
		{
			Provider:         "anthropic",
			Name:             "claude-3-haiku-20240307",
			DisplayName:      "Claude 3 Haiku",
			Description:      "Anthropic's fastest and most compact model for near-instant responsiveness, ideal for customer interactions and content moderation.",
			DocumentationURL: "https://docs.anthropic.com/claude/docs/models-overview",
			Capabilities: domain.Capabilities{
				Text:            domain.MediaTypeCapability{Read: true, Write: true},
				Image:           domain.MediaTypeCapability{Read: true, Write: false}, // Can process images
				Audio:           domain.MediaTypeCapability{Read: false, Write: false},
				Video:           domain.MediaTypeCapability{Read: false, Write: false},
				File:            domain.MediaTypeCapability{Read: false, Write: false},
				FunctionCalling: true,
				Streaming:       true,
				JSONMode:        true, // Supported via specific prompting techniques
			},
			ContextWindow:    200000,
			MaxOutputTokens:  4096,
			TrainingCutoff:   "2023-08", // Approximate
			ModelFamily:      "claude-3",
			Pricing: domain.Pricing{ // Example pricing, verify from official source
				InputPer1kTokens:  0.00025,
				OutputPer1kTokens: 0.00125,
			},
			LastUpdated: "2024-03-07",
		},
		{
			Provider:         "anthropic",
			Name:             "claude-2.1",
			DisplayName:      "Claude 2.1",
			Description:      "Previous generation model with a 200K context window and reduced rates of hallucination.",
			DocumentationURL: "https://docs.anthropic.com/claude/docs/models-overview",
			Capabilities: domain.Capabilities{
				Text:            domain.MediaTypeCapability{Read: true, Write: true},
				Image:           domain.MediaTypeCapability{Read: false, Write: false},
				Audio:           domain.MediaTypeCapability{Read: false, Write: false},
				Video:           domain.MediaTypeCapability{Read: false, Write: false},
				File:            domain.MediaTypeCapability{Read: true, Write: false}, // Supported document Q&A
				FunctionCalling: false, // Not officially supported in the same way as Claude 3
				Streaming:       true,
				JSONMode:        false,
			},
			ContextWindow:    200000,
			MaxOutputTokens:  4096,
			TrainingCutoff:   "Early 2023", // Approximate
			ModelFamily:      "claude-2",
			Pricing: domain.Pricing{
				InputPer1kTokens:  0.008,
				OutputPer1kTokens: 0.024,
			},
			LastUpdated: "2023-11-21", // Release date
		},
	}
	return models, nil
}
