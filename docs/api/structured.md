# Structured Package

The `structured` package provides functionality for generating and processing structured outputs from language models according to predefined schemas. It handles prompt enhancement for schema-guided generation and processing of raw LLM outputs.

## Core Components

### Domain

#### Processor Interface

```go
type Processor interface {
    // Process processes a raw output string against a schema
    Process(schema *schemaDomain.Schema, output string) (interface{}, error)

    // ProcessTyped processes a raw output string against a schema and maps it to a specific type
    ProcessTyped(schema *schemaDomain.Schema, output string, target interface{}) error

    // ToJSON converts an object to a JSON string
    ToJSON(obj interface{}) (string, error)
}
```

The `Processor` interface defines methods for validating and processing raw LLM outputs against schemas, with support for mapping to specific Go types.

#### PromptEnhancer Interface

```go
type PromptEnhancer interface {
    // Enhance adds schema information to a prompt
    Enhance(prompt string, schema *schemaDomain.Schema) (string, error)

    // EnhanceWithOptions adds schema information to a prompt with additional options
    EnhanceWithOptions(prompt string, schema *schemaDomain.Schema, options map[string]interface{}) (string, error)
}
```

The `PromptEnhancer` interface defines methods for adding schema information to prompts to guide language models in generating structured outputs.

## Processor Package

The processor package implements the `Processor` interface and provides functionality for processing structured outputs.

### JsonProcessor

```go
// Create a new processor
processor := processor.NewJsonProcessor()

// Process raw output against a schema
result, err := processor.Process(schema, rawOutput)
if err != nil {
    // Handle error
}

// Use the processed result
// result is a map[string]interface{} for objects or []interface{} for arrays
```

The `JsonProcessor` handles JSON extraction, validation, and conversion to Go types:

```go
// Define a target struct
type Person struct {
    Name  string `json:"name"`
    Age   int    `json:"age"`
    Email string `json:"email"`
}

// Process raw output into a specific type
var person Person
err := processor.ProcessTyped(schema, rawOutput, &person)
if err != nil {
    // Handle error
}

// Use the typed result
fmt.Printf("Name: %s, Age: %d, Email: %s\n", person.Name, person.Age, person.Email)
```

### JsonExtractor

The `JsonExtractor` is used internally by the processor to extract JSON from raw LLM outputs, handling common formatting issues:

```go
// Create a new extractor
extractor := processor.NewJsonExtractor()

// Extract JSON from raw output
json, err := extractor.Extract(rawOutput)
if err != nil {
    // Handle error
}

// Use the extracted JSON
// ...
```

The extractor handles various edge cases such as:

- JSON embedded in markdown code blocks
- Text before or after the JSON
- Line breaks and indentation
- Incomplete or malformed JSON in some cases

## PromptEnhancer Implementation

The `PromptEnhancer` implementation adds schema information to prompts to guide language models in generating structured outputs:

```go
// Create a new prompt enhancer
enhancer := processor.NewPromptEnhancer()

// Enhance a prompt with schema information
enhancedPrompt, err := enhancer.Enhance(prompt, schema)
if err != nil {
    // Handle error
}

// Use the enhanced prompt with an LLM provider
// ...
```

The enhancer adds detailed instructions based on the schema:

```go
// Enhanced prompt example
/*
What is the capital of France?

Please provide your response as a valid JSON object that conforms to the following JSON schema:

```json
{
  "type": "object",
  "properties": {
    "capital": {
      "type": "string",
      "description": "The capital city name"
    },
    "country": {
      "type": "string",
      "description": "The country name"
    },
    "population": {
      "type": "integer",
      "description": "The population of the capital"
    }
  },
  "required": ["capital", "country"]
}
```

Your response must be valid JSON only, following these guidelines:
1. Do not include any explanations, markdown code blocks, or additional text before or after the JSON.
2. Ensure all required fields are included.
3. The required fields are: capital, country.
4. Field descriptions:
   - capital: The capital city name
   - country: The country name
   - population: The population of the capital
*/
```

### Enhanced Options

The enhancer also supports additional options for more specific guidance:

```go
// Create options
options := map[string]interface{}{
    "instructions": "Be concise and accurate with population estimates.",
    "format": "a complete JSON object with all fields.",
    "examples": []map[string]interface{}{
        {
            "capital": "Berlin",
            "country": "Germany",
            "population": 3700000,
        },
    },
}

// Enhance prompt with options
enhancedPrompt, err := enhancer.EnhanceWithOptions(prompt, schema, options)
```

## Schema Cache

The `SchemaCache` provides caching for schema JSON to avoid repeated marshaling:

```go
// The cache is used internally by the PromptEnhancer
// You generally don't need to interact with it directly

// If needed, you can clear the cache
cache := processor.GetSchemaCache()
cache.Clear()
```

## Convenience Functions

The package provides several convenience functions:

```go
// Enhance a prompt with schema information (uses singleton enhancer)
enhancedPrompt, err := processor.EnhancePromptWithSchema(prompt, schema)

// Get the default prompt enhancer singleton
enhancer := processor.GetDefaultPromptEnhancer()
```

## Example Usage

### Complete Structured Output Flow

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/lexlapax/go-llms/pkg/llm/domain"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
    "github.com/lexlapax/go-llms/pkg/schema/domain"
    "github.com/lexlapax/go-llms/pkg/structured/processor"
)

func main() {
    // Define a schema
    schema := &domain.Schema{
        Type: "object",
        Properties: map[string]domain.Property{
            "capital": {Type: "string", Description: "The capital city name"},
            "country": {Type: "string", Description: "The country name"},
            "population": {Type: "integer", Description: "The population of the capital"},
        },
        Required: []string{"capital", "country"},
    }
    
    // Create a prompt
    prompt := "What is the capital of France?"
    
    // Enhance the prompt with schema information
    enhancedPrompt, err := processor.EnhancePromptWithSchema(prompt, schema)
    if err != nil {
        fmt.Printf("Error enhancing prompt: %v\n", err)
        return
    }
    
    // Create an LLM provider
    llmProvider := provider.NewOpenAIProvider("your-api-key", "gpt-4o")
    
    // Generate a response using the enhanced prompt
    rawOutput, err := llmProvider.Generate(context.Background(), enhancedPrompt)
    if err != nil {
        fmt.Printf("Error generating response: %v\n", err)
        return
    }
    
    // Create a processor
    proc := processor.NewJsonProcessor()
    
    // Process the raw output
    result, err := proc.Process(schema, rawOutput)
    if err != nil {
        fmt.Printf("Error processing output: %v\n", err)
        return
    }
    
    // Use the result
    data := result.(map[string]interface{})
    fmt.Printf("Capital: %s\n", data["capital"])
    fmt.Printf("Country: %s\n", data["country"])
    if population, ok := data["population"].(float64); ok {
        fmt.Printf("Population: %.0f\n", population)
    }
}
```

### Using ProcessTyped for Direct Type Mapping

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/lexlapax/go-llms/pkg/llm/provider"
    "github.com/lexlapax/go-llms/pkg/schema/domain"
    "github.com/lexlapax/go-llms/pkg/structured/processor"
)

// Define a struct matching the schema
type Capital struct {
    Capital    string `json:"capital"`
    Country    string `json:"country"`
    Population int    `json:"population,omitempty"`
}

func main() {
    // Define a schema
    schema := &domain.Schema{
        Type: "object",
        Properties: map[string]domain.Property{
            "capital": {Type: "string", Description: "The capital city name"},
            "country": {Type: "string", Description: "The country name"},
            "population": {Type: "integer", Description: "The population of the capital"},
        },
        Required: []string{"capital", "country"},
    }
    
    // Create a prompt
    prompt := "What is the capital of Japan?"
    
    // Enhance the prompt with schema information
    enhancedPrompt, err := processor.EnhancePromptWithSchema(prompt, schema)
    if err != nil {
        fmt.Printf("Error enhancing prompt: %v\n", err)
        return
    }
    
    // Create an LLM provider
    llmProvider := provider.NewOpenAIProvider("your-api-key", "gpt-4o")
    
    // Generate a response
    rawOutput, err := llmProvider.Generate(context.Background(), enhancedPrompt)
    if err != nil {
        fmt.Printf("Error generating response: %v\n", err)
        return
    }
    
    // Create a processor
    proc := processor.NewJsonProcessor()
    
    // Process the raw output directly into a struct
    var capital Capital
    err = proc.ProcessTyped(schema, rawOutput, &capital)
    if err != nil {
        fmt.Printf("Error processing output: %v\n", err)
        return
    }
    
    // Use the typed result
    fmt.Printf("Capital: %s\n", capital.Capital)
    fmt.Printf("Country: %s\n", capital.Country)
    fmt.Printf("Population: %d\n", capital.Population)
}
```
