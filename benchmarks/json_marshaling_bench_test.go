package benchmarks

import (
	"bytes"
	"encoding/json"
	"testing"

	optimizedJson "github.com/lexlapax/go-llms/pkg/util/json"
)

// BenchmarkJSONMarshal benchmarks the performance of JSON marshaling
func BenchmarkJSONMarshal(b *testing.B) {
	// Simple map
	simpleMap := map[string]interface{}{
		"name":  "John Doe",
		"age":   30,
		"email": "john@example.com",
	}

	// Complex struct with nested fields - similar to request bodies in providers
	complexStruct := struct {
		Model       string                   `json:"model"`
		Messages    []map[string]interface{} `json:"messages"`
		Temperature float64                  `json:"temperature,omitempty"`
		MaxTokens   int                      `json:"max_tokens,omitempty"`
		Stream      bool                     `json:"stream,omitempty"`
		StopSequences []string              `json:"stop,omitempty"`
		Tools        []map[string]interface{} `json:"tools,omitempty"`
	}{
		Model: "gpt-4-turbo",
		Messages: []map[string]interface{}{
			{"role": "system", "content": "You are a helpful assistant."},
			{"role": "user", "content": "Tell me about JSON performance in Go."},
		},
		Temperature: 0.7,
		MaxTokens:   1024,
	}

	// Very large nested structure with arrays - similar to LLM responses
	var largeStruct = make(map[string]interface{})
	largeStruct["choices"] = []map[string]interface{}{
		{
			"index": 0,
			"message": map[string]interface{}{
				"role": "assistant",
				"content": "JSON (JavaScript Object Notation) is a lightweight data interchange format that is easy for humans to read and write and easy for machines to parse and generate. Here's a comprehensive overview of JSON performance in Go...",
				"function_call": nil,
				"tool_calls": []map[string]interface{}{
					{
						"id": "call_123",
						"type": "function",
						"function": map[string]interface{}{
							"name": "get_weather",
							"arguments": `{"location": "San Francisco", "unit": "celsius"}`,
						},
					},
				},
			},
			"finish_reason": "stop",
		},
	}
	largeStruct["id"] = "chatcmpl-123456789"
	largeStruct["object"] = "chat.completion"
	largeStruct["created"] = 1677858242
	largeStruct["model"] = "gpt-4-turbo"
	largeStruct["usage"] = map[string]interface{}{
		"prompt_tokens":     56,
		"completion_tokens": 378,
		"total_tokens":      434,
	}

	// Benchmark standard JSON marshaling for simple map
	b.Run("StandardJSON_SimpleMap", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := json.Marshal(simpleMap)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// Benchmark optimized JSON marshaling for simple map
	b.Run("OptimizedJSON_SimpleMap", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := optimizedJson.Marshal(simpleMap)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// Benchmark standard JSON marshaling for complex struct
	b.Run("StandardJSON_ComplexStruct", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := json.Marshal(complexStruct)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// Benchmark optimized JSON marshaling for complex struct
	b.Run("OptimizedJSON_ComplexStruct", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := optimizedJson.Marshal(complexStruct)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// Benchmark standard JSON marshaling for large struct
	b.Run("StandardJSON_LargeStruct", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := json.Marshal(largeStruct)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// Benchmark optimized JSON marshaling for large struct
	b.Run("OptimizedJSON_LargeStruct", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := optimizedJson.Marshal(largeStruct)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	
	// Benchmark optimized JSON marshaling with buffer reuse
	b.Run("OptimizedJSON_WithBuffer_LargeStruct", func(b *testing.B) {
		buf := &bytes.Buffer{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := optimizedJson.MarshalWithBuffer(largeStruct, buf)
			if err != nil {
				b.Fatal(err)
			}
			// Don't do anything with the buffer - it will be reused
		}
	})
	
	// Benchmark string-optimized marshaling
	b.Run("OptimizedJSON_ToString_LargeStruct", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := optimizedJson.MarshalToString(largeStruct)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkJSONUnmarshal benchmarks the performance of JSON unmarshaling
func BenchmarkJSONUnmarshal(b *testing.B) {
	// Simple JSON
	simpleJSON := []byte(`{"name":"John Doe","age":30,"email":"john@example.com"}`)
	
	// Medium JSON - typical provider request
	mediumJSON := []byte(`{
		"model": "gpt-4-turbo",
		"messages": [
			{"role": "system", "content": "You are a helpful assistant."},
			{"role": "user", "content": "Tell me about JSON performance in Go."}
		],
		"temperature": 0.7,
		"max_tokens": 1024
	}`)
	
	// Large JSON - typical provider response
	largeJSON := []byte(`{
		"id": "chatcmpl-123456789",
		"object": "chat.completion",
		"created": 1677858242,
		"model": "gpt-4-turbo",
		"choices": [
			{
				"index": 0,
				"message": {
					"role": "assistant",
					"content": "JSON (JavaScript Object Notation) is a lightweight data interchange format that is easy for humans to read and write and easy for machines to parse and generate. Here's a comprehensive overview of JSON performance in Go...",
					"tool_calls": [
						{
							"id": "call_123",
							"type": "function",
							"function": {
								"name": "get_weather",
								"arguments": "{\"location\": \"San Francisco\", \"unit\": \"celsius\"}"
							}
						}
					]
				},
				"finish_reason": "stop"
			}
		],
		"usage": {
			"prompt_tokens": 56,
			"completion_tokens": 378,
			"total_tokens": 434
		}
	}`)
	
	// Benchmark standard JSON unmarshaling for simple JSON
	b.Run("StandardJSON_SimpleUnmarshal", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var data map[string]interface{}
			err := json.Unmarshal(simpleJSON, &data)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	
	// Benchmark optimized JSON unmarshaling for simple JSON
	b.Run("OptimizedJSON_SimpleUnmarshal", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var data map[string]interface{}
			err := optimizedJson.Unmarshal(simpleJSON, &data)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	
	// Benchmark standard JSON unmarshaling for medium JSON
	b.Run("StandardJSON_MediumUnmarshal", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var data map[string]interface{}
			err := json.Unmarshal(mediumJSON, &data)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	
	// Benchmark optimized JSON unmarshaling for medium JSON
	b.Run("OptimizedJSON_MediumUnmarshal", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var data map[string]interface{}
			err := optimizedJson.Unmarshal(mediumJSON, &data)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	
	// Benchmark standard JSON unmarshaling for large JSON
	b.Run("StandardJSON_LargeUnmarshal", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var data map[string]interface{}
			err := json.Unmarshal(largeJSON, &data)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	
	// Benchmark optimized JSON unmarshaling for large JSON
	b.Run("OptimizedJSON_LargeUnmarshal", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var data map[string]interface{}
			err := optimizedJson.Unmarshal(largeJSON, &data)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	
	// Benchmark string-optimized JSON unmarshaling
	b.Run("OptimizedJSON_FromString_LargeUnmarshal", func(b *testing.B) {
		jsonString := string(largeJSON)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var data map[string]interface{}
			err := optimizedJson.UnmarshalFromString(jsonString, &data)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	
	// Benchmark with a specific struct type for provider requests
	type ProviderRequest struct {
		Model       string                   `json:"model"`
		Messages    []map[string]interface{} `json:"messages"`
		Temperature float64                  `json:"temperature,omitempty"`
		MaxTokens   int                      `json:"max_tokens,omitempty"`
		Stream      bool                     `json:"stream,omitempty"`
	}
	
	b.Run("StandardJSON_StructUnmarshal", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var req ProviderRequest
			err := json.Unmarshal(mediumJSON, &req)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	
	b.Run("OptimizedJSON_StructUnmarshal", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var req ProviderRequest
			err := optimizedJson.Unmarshal(mediumJSON, &req)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}