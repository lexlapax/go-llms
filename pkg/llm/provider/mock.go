package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// MockProvider implements the Provider interface for testing
type MockProvider struct {
	generateFunc           func(ctx context.Context, prompt string, options ...domain.Option) (string, error)
	generateMessageFunc    func(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.Response, error)
	streamFunc             func(ctx context.Context, prompt string, options ...domain.Option) (domain.ResponseStream, error)
	streamMessageFunc      func(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.ResponseStream, error)
	generateWithSchemaFunc func(ctx context.Context, prompt string, schema *schemaDomain.Schema, options ...domain.Option) (interface{}, error)
	// Predefined responses for testing
	predefinedResponses map[string]string
	// Options storage for testing
	httpClient     *http.Client
	baseURL        string
	headers        map[string]string
	safetySettings []map[string]interface{}
	customSettings map[string]interface{}
}

// NewMockProvider creates a new mock provider with default implementations and options
func NewMockProvider(options ...domain.ProviderOption) *MockProvider {
	provider := &MockProvider{
		generateFunc: func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
			return `{"result": "This is a mock response"}`, nil
		},
		generateMessageFunc: func(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.Response, error) {
			return domain.Response{Content: "This is a mock message response"}, nil
		},
		streamFunc: func(ctx context.Context, prompt string, options ...domain.Option) (domain.ResponseStream, error) {
			ch := make(chan domain.Token)
			go func() {
				defer close(ch)
				words := strings.Split("This is a mock streamed response", " ")
				for i, word := range words {
					select {
					case <-ctx.Done():
						return
					case ch <- domain.Token{
						Text:     word,
						Finished: i == len(words)-1,
					}:
						time.Sleep(50 * time.Millisecond) // Simulate delay
					}
				}
			}()
			return ch, nil
		},
		streamMessageFunc: func(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.ResponseStream, error) {
			ch := make(chan domain.Token)
			go func() {
				defer close(ch)
				words := strings.Split("This is a mock streamed message response", " ")
				for i, word := range words {
					select {
					case <-ctx.Done():
						return
					case ch <- domain.Token{
						Text:     word,
						Finished: i == len(words)-1,
					}:
						time.Sleep(50 * time.Millisecond) // Simulate delay
					}
				}
			}()
			return ch, nil
		},
		generateWithSchemaFunc: func(ctx context.Context, prompt string, schema *schemaDomain.Schema, options ...domain.Option) (interface{}, error) {
			// Return a mock response based on schema
			if schema.Type == "object" {
				// Create a response object with mock data for each property
				result := make(map[string]interface{})
				for propName, prop := range schema.Properties {
					switch prop.Type {
					case "string":
						result[propName] = fmt.Sprintf("mock_%s", propName)
					case "integer":
						result[propName] = 42
					case "number":
						result[propName] = 42.5
					case "boolean":
						result[propName] = true
					case "array":
						// If items are defined, create a mock array with 2 items
						if prop.Items != nil {
							switch prop.Items.Type {
							case "string":
								result[propName] = []string{"item1", "item2"}
							case "integer":
								result[propName] = []int{1, 2}
							case "number":
								result[propName] = []float64{1.1, 2.2}
							case "boolean":
								result[propName] = []bool{true, false}
							default:
								result[propName] = []string{"item1", "item2"}
							}
						} else {
							result[propName] = []string{"item1", "item2"}
						}
					case "object":
						// Create nested object with mock data
						nested := make(map[string]interface{})
						for nestedPropName, nestedProp := range prop.Properties {
							switch nestedProp.Type {
							case "string":
								nested[nestedPropName] = fmt.Sprintf("nested_%s", nestedPropName)
							case "integer":
								nested[nestedPropName] = 42
							case "number":
								nested[nestedPropName] = 42.5
							case "boolean":
								nested[nestedPropName] = true
							default:
								nested[nestedPropName] = fmt.Sprintf("nested_%s", nestedPropName)
							}
						}
						result[propName] = nested
					default:
						result[propName] = fmt.Sprintf("mock_%s", propName)
					}
				}
				return result, nil
			}

			// Default for non-object schemas
			return map[string]interface{}{"result": "mock response"}, nil
		},
		// Initialize the headers map
		headers: make(map[string]string),
		// Initialize the custom settings map
		customSettings: make(map[string]interface{}),
	}

	// Apply provider options
	for _, option := range options {
		if mockOption, ok := option.(domain.MockOption); ok {
			mockOption.ApplyToMock(provider)
		}
	}

	return provider
}

// Generate produces text from a prompt
func (p *MockProvider) Generate(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
	return p.generateFunc(ctx, prompt, options...)
}

// GenerateMessage produces text from a list of messages
func (p *MockProvider) GenerateMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.Response, error) {
	return p.generateMessageFunc(ctx, messages, options...)
}

// GenerateWithSchema produces structured output conforming to a schema
func (p *MockProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *schemaDomain.Schema, options ...domain.Option) (interface{}, error) {
	return p.generateWithSchemaFunc(ctx, prompt, schema, options...)
}

// Stream streams responses token by token
func (p *MockProvider) Stream(ctx context.Context, prompt string, options ...domain.Option) (domain.ResponseStream, error) {
	return p.streamFunc(ctx, prompt, options...)
}

// StreamMessage streams responses from a list of messages
func (p *MockProvider) StreamMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.ResponseStream, error) {
	return p.streamMessageFunc(ctx, messages, options...)
}

// WithGenerateFunc sets a custom generate function
func (p *MockProvider) WithGenerateFunc(f func(ctx context.Context, prompt string, options ...domain.Option) (string, error)) *MockProvider {
	p.generateFunc = f
	return p
}

// WithGenerateMessageFunc sets a custom generate message function
func (p *MockProvider) WithGenerateMessageFunc(f func(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.Response, error)) *MockProvider {
	p.generateMessageFunc = f
	return p
}

// WithGenerateWithSchemaFunc sets a custom generate with schema function
func (p *MockProvider) WithGenerateWithSchemaFunc(f func(ctx context.Context, prompt string, schema *schemaDomain.Schema, options ...domain.Option) (interface{}, error)) *MockProvider {
	p.generateWithSchemaFunc = f
	return p
}

// WithStreamFunc sets a custom stream function
func (p *MockProvider) WithStreamFunc(f func(ctx context.Context, prompt string, options ...domain.Option) (domain.ResponseStream, error)) *MockProvider {
	p.streamFunc = f
	return p
}

// WithStreamMessageFunc sets a custom stream message function
func (p *MockProvider) WithStreamMessageFunc(f func(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.ResponseStream, error)) *MockProvider {
	p.streamMessageFunc = f
	return p
}

// WithPredefinedResponses sets predefined responses for specific prompts
func (p *MockProvider) WithPredefinedResponses(responses map[string]string) *MockProvider {
	// Initialize the map if it's nil
	if p.predefinedResponses == nil {
		p.predefinedResponses = make(map[string]string)
	}

	// Copy the responses
	for prompt, response := range responses {
		p.predefinedResponses[prompt] = response
	}

	// Override the generateFunc to use predefined responses
	p.generateFunc = func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
		if response, ok := p.predefinedResponses[prompt]; ok {
			return response, nil
		}
		// Fallback to default response
		return "This is a default mock response", nil
	}

	return p
}

// Getter and setter methods for MockProvider options

// SetHTTPClient sets the HTTP client for the provider
func (p *MockProvider) SetHTTPClient(client *http.Client) {
	p.httpClient = client
}

// SetBaseURL sets the base URL for the provider
func (p *MockProvider) SetBaseURL(url string) {
	p.baseURL = url
}

// SetHeaders sets custom HTTP headers for the provider
func (p *MockProvider) SetHeaders(headers map[string]string) {
	p.headers = headers
}

// SetSafetySettings sets safety settings for the provider
func (p *MockProvider) SetSafetySettings(settings []map[string]interface{}) {
	p.safetySettings = settings
}

// SetCustomSetting sets a custom setting for the provider
func (p *MockProvider) SetCustomSetting(key string, value interface{}) {
	p.customSettings[key] = value
}

// GenerateJSONResponse generates a mock JSON response for testing
func GenerateJSONResponse(schema *schemaDomain.Schema) (string, error) {
	if schema == nil {
		return `{"result": "This is a mock response"}`, nil
	}

	// Create mock data based on the schema
	mockData, err := generateMockData(schema)
	if err != nil {
		return "", err
	}

	// Convert to JSON
	jsonData, err := json.Marshal(mockData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal mock data: %w", err)
	}

	return string(jsonData), nil
}

// generateMockData creates mock data based on a schema
func generateMockData(schema *schemaDomain.Schema) (interface{}, error) {
	switch schema.Type {
	case "object":
		obj := make(map[string]interface{})
		for name, prop := range schema.Properties {
			value, err := generateMockValue(&prop)
			if err != nil {
				return nil, err
			}
			obj[name] = value
		}
		return obj, nil

	case "array":
		if schema.Properties != nil && schema.Properties[""].Items != nil {
			// Create a mock array with items of the specified type
			itemSchema := schema.Properties[""].Items
			itemValue, err := generateMockValue(itemSchema)
			if err != nil {
				return nil, err
			}
			// Return an array with 2 items
			return []interface{}{itemValue, itemValue}, nil
		}
		return []interface{}{}, nil

	case "string":
		return "mock_string", nil

	case "integer":
		return 42, nil

	case "number":
		return 42.5, nil

	case "boolean":
		return true, nil

	default:
		return "mock_default", nil
	}
}

// generateMockValue creates a mock value based on a property definition
func generateMockValue(prop *schemaDomain.Property) (interface{}, error) {
	switch prop.Type {
	case "object":
		obj := make(map[string]interface{})
		for name, nestedProp := range prop.Properties {
			value, err := generateMockValue(&nestedProp)
			if err != nil {
				return nil, err
			}
			obj[name] = value
		}
		return obj, nil

	case "array":
		if prop.Items != nil {
			// Create a mock array with items of the specified type
			itemValue, err := generateMockValue(prop.Items)
			if err != nil {
				return nil, err
			}
			// Return an array with 2 items
			return []interface{}{itemValue, itemValue}, nil
		}
		return []interface{}{}, nil

	case "string":
		if prop.Format == "email" {
			return "user@example.com", nil
		}
		if prop.Format == "date-time" {
			return time.Now().Format(time.RFC3339), nil
		}
		if prop.Format == "uri" {
			return "https://example.com", nil
		}
		if len(prop.Enum) > 0 {
			// Return the first enum value
			return prop.Enum[0], nil
		}
		return "mock_string", nil

	case "integer":
		if prop.Minimum != nil && prop.Maximum != nil {
			// Return a value between min and max
			min := int(*prop.Minimum)
			max := int(*prop.Maximum)
			return min + (max-min)/2, nil
		}
		if prop.Minimum != nil {
			// Return a value greater than min
			return int(*prop.Minimum) + 1, nil
		}
		if prop.Maximum != nil {
			// Return a value less than max
			return int(*prop.Maximum) - 1, nil
		}
		return 42, nil

	case "number":
		if prop.Minimum != nil && prop.Maximum != nil {
			// Return a value between min and max
			min := *prop.Minimum
			max := *prop.Maximum
			return min + (max-min)/2, nil
		}
		if prop.Minimum != nil {
			// Return a value greater than min
			return *prop.Minimum + 1.0, nil
		}
		if prop.Maximum != nil {
			// Return a value less than max
			return *prop.Maximum - 1.0, nil
		}
		return 42.5, nil

	case "boolean":
		return true, nil

	default:
		return "mock_default", nil
	}
}
