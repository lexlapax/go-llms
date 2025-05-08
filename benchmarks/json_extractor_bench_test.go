package benchmarks

import (
	"strings"
	"testing"

	structProcessor "github.com/lexlapax/go-llms/pkg/structured/processor"
)

// Helper for the original simple extraction method
func originalExtractJSON(s string) string {
	// Look for JSON object between curly braces
	startIdx := strings.Index(s, "{")
	endIdx := strings.LastIndex(s, "}")

	if startIdx >= 0 && endIdx > startIdx {
		return s[startIdx : endIdx+1]
	}

	// Look for JSON array between square brackets
	startIdx = strings.Index(s, "[")
	endIdx = strings.LastIndex(s, "]")

	if startIdx >= 0 && endIdx > startIdx {
		return s[startIdx : endIdx+1]
	}

	// No JSON found
	return ""
}

// BenchmarkJSONExtraction_Simple tests extraction from simple text containing JSON
func BenchmarkJSONExtraction_Simple(b *testing.B) {
	// Simple text with JSON
	simpleText := `Here's the information you requested:
{
  "name": "John Doe",
  "age": 30,
  "email": "john@example.com"
}`

	// Run benchmark for original method
	b.Run("Original", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			result := originalExtractJSON(simpleText)
			if result == "" {
				b.Fatal("Failed to extract JSON")
			}
		}
	})

	// Run benchmark for optimized method
	b.Run("Optimized", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			result := structProcessor.ExtractJSON(simpleText)
			if result == "" {
				b.Fatal("Failed to extract JSON")
			}
		}
	})
}

// BenchmarkJSONExtraction_Complex tests extraction from markdown with JSON
func BenchmarkJSONExtraction_Complex(b *testing.B) {
	// Text with JSON in markdown code block
	markdownText := `# User Information

Here's the user information you requested:

` + "```json" + `
{
  "name": "Jane Smith",
  "age": 28,
  "email": "jane@example.com",
  "address": {
    "street": "123 Main St",
    "city": "Anytown",
    "state": "CA",
    "zip": "12345"
  },
  "preferences": {
    "theme": "dark",
    "notifications": true,
    "language": "en-US"
  },
  "tags": ["premium", "verified", "active"]
}
` + "```" + `

I hope this helps!`

	// Run benchmark for original method
	b.Run("Original", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			result := originalExtractJSON(markdownText)
			if result == "" {
				b.Fatal("Failed to extract JSON")
			}
		}
	})

	// Run benchmark for optimized method
	b.Run("Optimized", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			result := structProcessor.ExtractJSON(markdownText)
			if result == "" {
				b.Fatal("Failed to extract JSON")
			}
		}
	})
}

// BenchmarkJSONExtraction_Nested tests extraction of deeply nested JSON
func BenchmarkJSONExtraction_Nested(b *testing.B) {
	// Text with nested JSON object
	nestedText := `The API returned the following response:
{
  "data": {
    "user": {
      "id": 12345,
      "profile": {
        "name": "John Doe",
        "email": "john@example.com",
        "addresses": [
          {
            "type": "home",
            "street": "123 Main St",
            "city": "Anytown",
            "state": "CA"
          },
          {
            "type": "work",
            "street": "456 Market St",
            "city": "Bigcity",
            "state": "NY"
          }
        ]
      },
      "settings": {
        "theme": {
          "mode": "dark",
          "colors": {
            "primary": "#336699",
            "secondary": "#993366"
          }
        },
        "notifications": {
          "email": true,
          "push": false,
          "sms": true
        }
      }
    }
  },
  "meta": {
    "status": "success",
    "timestamp": "2025-05-06T12:30:00Z"
  }
}`

	// Run benchmark for original method
	b.Run("Original", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			result := originalExtractJSON(nestedText)
			if result == "" {
				b.Fatal("Failed to extract JSON")
			}
		}
	})

	// Run benchmark for optimized method
	b.Run("Optimized", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			result := structProcessor.ExtractJSON(nestedText)
			if result == "" {
				b.Fatal("Failed to extract JSON")
			}
		}
	})
}
