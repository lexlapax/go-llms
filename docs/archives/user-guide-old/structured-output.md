# Structured Output Guide

Learn how to reliably extract structured data from LLMs using go-llms.

## Overview

Getting structured data from LLMs can be challenging. Go-llms provides powerful tools to ensure you get exactly the data structure you need, with validation and type safety.

## Quick Start

### Basic Structured Output

```go
// Define what you want
type ProductInfo struct {
    Name        string   `json:"name"`
    Price       float64  `json:"price"`
    InStock     bool     `json:"inStock"`
    Categories  []string `json:"categories"`
}

// Ask for it
var product ProductInfo
err := provider.GenerateWithSchema(
    context.Background(),
    "Tell me about the iPhone 15 Pro",
    &product,
)

// Use it with confidence
fmt.Printf("%s costs $%.2f\n", product.Name, product.Price)
```

### With Validation

```go
// Define schema with constraints
schema := &domain.Schema{
    Type: "object",
    Properties: map[string]*domain.Schema{
        "age": {
            Type:        "integer",
            Minimum:     float64Ptr(0),
            Maximum:     float64Ptr(150),
            Description: "Person's age in years",
        },
        "email": {
            Type:        "string",
            Format:      "email",
            Description: "Valid email address",
        },
        "score": {
            Type:        "number",
            Minimum:     float64Ptr(0.0),
            Maximum:     float64Ptr(100.0),
            Description: "Score as percentage",
        },
    },
    Required: []string{"age", "email"},
}

// Generate with validation
result, err := structured.ProcessWithSchema(
    provider,
    "Extract user data from: John is 25, email john@example.com, scored 95%",
    schema,
)
```

## Schema Definition

### Using Go Structs

The easiest way is using Go structs with JSON tags:

```go
type Address struct {
    Street     string `json:"street"`
    City       string `json:"city"`
    State      string `json:"state,omitempty"`
    PostalCode string `json:"postalCode"`
    Country    string `json:"country"`
}

type Person struct {
    Name      string    `json:"name"`
    Age       int       `json:"age"`
    Email     string    `json:"email"`
    Address   Address   `json:"address"`
    Hobbies   []string  `json:"hobbies"`
    IsPremium bool      `json:"isPremium"`
    JoinedAt  time.Time `json:"joinedAt"`
}

// Use directly
var person Person
err := provider.GenerateWithSchema(ctx, prompt, &person)
```

### Using Schema Objects

For more control, use schema objects:

```go
personSchema := &domain.Schema{
    Type: "object",
    Properties: map[string]*domain.Schema{
        "name": {
            Type:        "string",
            MinLength:   intPtr(1),
            MaxLength:   intPtr(100),
            Description: "Person's full name",
        },
        "age": {
            Type:        "integer",
            Minimum:     float64Ptr(0),
            Maximum:     float64Ptr(150),
            Description: "Age in years",
        },
        "email": {
            Type:        "string",
            Format:      "email",
            Pattern:     `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
            Description: "Valid email address",
        },
        "hobbies": {
            Type: "array",
            Items: &domain.Schema{
                Type: "string",
            },
            MinItems:    intPtr(0),
            MaxItems:    intPtr(10),
            UniqueItems: true,
            Description: "List of hobbies",
        },
        "address": {
            Type: "object",
            Properties: map[string]*domain.Schema{
                "street":     {Type: "string"},
                "city":       {Type: "string"},
                "postalCode": {Type: "string", Pattern: `^\d{5}(-\d{4})?$`},
            },
            Required: []string{"street", "city"},
        },
    },
    Required:             []string{"name", "email"},
    AdditionalProperties: false, // No extra fields allowed
}
```

## Advanced Validation

### Built-in Validators

```go
// String validation
{
    Type:      "string",
    MinLength: intPtr(3),
    MaxLength: intPtr(50),
    Pattern:   `^[A-Z][a-z]+$`, // Must start with capital
    Enum:      []interface{}{"red", "green", "blue"}, // Allowed values
}

// Number validation
{
    Type:             "number",
    Minimum:          float64Ptr(0),
    Maximum:          float64Ptr(100),
    ExclusiveMinimum: true,  // > 0, not >= 0
    MultipleOf:       float64Ptr(0.25), // Must be multiple of 0.25
}

// Array validation
{
    Type:     "array",
    MinItems: intPtr(1),
    MaxItems: intPtr(100),
    UniqueItems: true, // No duplicates
    Items: &domain.Schema{
        Type: "string",
        Enum: []interface{}{"option1", "option2", "option3"},
    },
}
```

### Custom Validators

```go
import "github.com/lexlapax/go-llms/pkg/schema/validation"

// Create custom validator
ageValidator := validation.NewCustomValidator(
    "age_validator",
    func(value interface{}) error {
        age, ok := value.(float64)
        if !ok {
            return fmt.Errorf("age must be a number")
        }
        
        if age < 0 || age > 150 {
            return fmt.Errorf("age must be between 0 and 150")
        }
        
        if age < 18 {
            return fmt.Errorf("must be 18 or older")
        }
        
        return nil
    },
)

// Add to schema
schema.Properties["age"].CustomValidator = ageValidator
```

### Complex Validation

```go
// Cross-field validation
type DateRange struct {
    StartDate string `json:"startDate"`
    EndDate   string `json:"endDate"`
}

validator := validation.NewCustomValidator(
    "date_range_validator",
    func(value interface{}) error {
        dr, ok := value.(map[string]interface{})
        if !ok {
            return fmt.Errorf("invalid date range")
        }
        
        start, _ := time.Parse("2006-01-02", dr["startDate"].(string))
        end, _ := time.Parse("2006-01-02", dr["endDate"].(string))
        
        if end.Before(start) {
            return fmt.Errorf("end date must be after start date")
        }
        
        return nil
    },
)
```

## Processing Strategies

### Direct Generation

Best for simple, well-defined structures:

```go
type SimpleResponse struct {
    Answer     string `json:"answer"`
    Confidence float64 `json:"confidence"`
}

var response SimpleResponse
err := provider.GenerateWithSchema(ctx, 
    "What is the capital of France?", 
    &response,
)
```

### With Retry Logic

For complex structures that might need refinement:

```go
processor := processor.NewJsonProcessor()

var result interface{}
maxRetries := 3

for i := 0; i < maxRetries; i++ {
    // Generate response
    response, err := provider.Generate(ctx, enhancedPrompt)
    if err != nil {
        return err
    }
    
    // Try to extract and validate
    result, err = processor.Process(schema, response)
    if err == nil {
        break // Success!
    }
    
    // Enhance prompt with error feedback
    enhancedPrompt = fmt.Sprintf(
        "%s\n\nPlease fix this error: %v",
        originalPrompt,
        err,
    )
}
```

### Streaming with Validation

For real-time applications:

```go
stream, err := provider.Stream(ctx, prompt)
if err != nil {
    return err
}

var buffer strings.Builder
for chunk := range stream {
    buffer.WriteString(chunk)
    
    // Try to parse partial JSON
    var partial interface{}
    if err := json.Unmarshal([]byte(buffer.String()), &partial); err == nil {
        // Validate what we have so far
        if err := validator.ValidatePartial(schema, partial); err == nil {
            // Update UI with valid partial data
            updateUI(partial)
        }
    }
}
```

## Type Coercion

Go-llms automatically handles common type mismatches:

```go
// LLM returns "25" (string), but schema expects integer
// Coercion handles this automatically

schema := &domain.Schema{
    Type: "object",
    Properties: map[string]*domain.Schema{
        "age": {Type: "integer"}, // Expects number
        "active": {Type: "boolean"}, // Expects bool
        "score": {Type: "number"}, // Expects float
    },
}

// These all work:
// {"age": "25", "active": "true", "score": "98.5"}
// {"age": 25, "active": 1, "score": "98.5"}
// {"age": "25", "active": "yes", "score": 98.5}
```

### Coercion Rules

```go
// String to Number
"123" → 123
"45.67" → 45.67
"" → 0 (if not required)

// String to Boolean  
"true", "yes", "1", "on" → true
"false", "no", "0", "off" → false

// Number to String
123 → "123"
45.67 → "45.67"

// Number to Boolean
0 → false
any other → true
```

## Error Handling

### Validation Errors

```go
result, err := processor.Process(schema, llmOutput)
if err != nil {
    if validationErr, ok := err.(*validation.ValidationError); ok {
        // Detailed validation errors
        for _, detail := range validationErr.Details {
            fmt.Printf("Field %s: %s\n", detail.Path, detail.Message)
        }
        
        // Generate helpful error message
        suggestion := generateFixSuggestion(validationErr)
        fmt.Printf("Suggestion: %s\n", suggestion)
    }
}
```

### Recovery Strategies

```go
// Strategy 1: Fallback to simpler schema
complexResult, err := tryComplexSchema(provider, prompt)
if err != nil {
    // Fall back to simpler version
    simpleResult, err := trySimpleSchema(provider, prompt)
}

// Strategy 2: Progressive enhancement
baseResult, _ := generateBase(provider, prompt)
enhancedResult, _ := enhanceWithDetails(provider, baseResult)
finalResult, _ := validateAndClean(enhancedResult)

// Strategy 3: Multi-provider consensus
results := []interface{}{}
for _, provider := range providers {
    if result, err := provider.GenerateWithSchema(ctx, prompt, schema); err == nil {
        results = append(results, result)
    }
}
consensusResult := findConsensus(results)
```

## Real-World Examples

### E-commerce Product Extraction

```go
type Product struct {
    Name        string      `json:"name"`
    Brand       string      `json:"brand"`
    Price       PriceInfo   `json:"price"`
    Availability string     `json:"availability"`
    Features    []string    `json:"features"`
    Ratings     RatingInfo  `json:"ratings"`
    Images      []ImageInfo `json:"images"`
}

type PriceInfo struct {
    Currency string  `json:"currency"`
    Amount   float64 `json:"amount"`
    Discount float64 `json:"discount,omitempty"`
}

type RatingInfo struct {
    Average float64 `json:"average"`
    Count   int     `json:"count"`
    Distribution map[string]int `json:"distribution"`
}

// Extract product info from description
prompt := fmt.Sprintf("Extract product information from this listing: %s", productHTML)
var product Product
err := provider.GenerateWithSchema(ctx, prompt, &product)
```

### Form Data Extraction

```go
type FormSubmission struct {
    PersonalInfo struct {
        FirstName   string `json:"firstName"`
        LastName    string `json:"lastName"`
        DateOfBirth string `json:"dateOfBirth"`
        SSN         string `json:"ssn,omitempty"`
    } `json:"personalInfo"`
    
    ContactInfo struct {
        Email       string `json:"email"`
        Phone       string `json:"phone"`
        Address     Address `json:"address"`
        Preferred   string `json:"preferredContact"`
    } `json:"contactInfo"`
    
    Preferences struct {
        Newsletter  bool     `json:"newsletter"`
        Interests   []string `json:"interests"`
        Language    string   `json:"language"`
    } `json:"preferences"`
}

// Extract from unstructured text
var form FormSubmission
err := provider.GenerateWithSchema(ctx, 
    "Extract form data from: " + userInput,
    &form,
)
```

### Data Analysis Results

```go
type AnalysisResult struct {
    Summary    string                 `json:"summary"`
    Metrics    map[string]float64     `json:"metrics"`
    Trends     []TrendInfo            `json:"trends"`
    Anomalies  []AnomalyInfo          `json:"anomalies"`
    Recommendations []string          `json:"recommendations"`
    Confidence float64                `json:"confidence"`
    Metadata   map[string]interface{} `json:"metadata"`
}

type TrendInfo struct {
    Metric    string  `json:"metric"`
    Direction string  `json:"direction"` // "up", "down", "stable"
    Magnitude float64 `json:"magnitude"`
    Period    string  `json:"period"`
}

// Analyze data and get structured insights
var analysis AnalysisResult
err := provider.GenerateWithSchema(ctx,
    fmt.Sprintf("Analyze this dataset and provide insights: %v", dataset),
    &analysis,
)
```

## Best Practices

### 1. Schema Design
- Start simple and add complexity gradually
- Use clear, descriptive field names
- Provide descriptions for each field
- Set reasonable constraints

### 2. Prompt Engineering
```go
// Include schema hints in prompt
prompt := fmt.Sprintf(`
Extract product information in this exact format:
- name: product name (string)
- price: price in USD (number)
- inStock: availability (boolean)
- categories: list of categories (array of strings)

Product description: %s
`, description)
```

### 3. Error Recovery
- Always validate output
- Provide clear error messages
- Have fallback strategies
- Log failures for analysis

### 4. Performance
- Cache validated schemas
- Reuse processors
- Batch similar requests
- Monitor token usage

### 5. Testing
```go
func TestStructuredOutput(t *testing.T) {
    testCases := []struct {
        name     string
        input    string
        expected Product
    }{
        {
            name:  "simple product",
            input: "iPhone 15 Pro, $999, in stock",
            expected: Product{
                Name:    "iPhone 15 Pro",
                Price:   999.0,
                InStock: true,
            },
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            var result Product
            err := provider.GenerateWithSchema(ctx, tc.input, &result)
            assert.NoError(t, err)
            assert.Equal(t, tc.expected, result)
        })
    }
}
```

## Integration with Agents

Structured output works seamlessly with agents:

```go
// Agent that extracts structured data
agent := core.NewLLMAgent("data-extractor", provider)
agent.SetSystemPrompt("Extract structured data according to schemas")

// Add extraction tool
extractTool := tools.NewToolBuilder("extract", "Extract structured data").
    WithFunction(func(ctx *domain.ToolContext, params map[string]interface{}) (interface{}, error) {
        text := params["text"].(string)
        schemaType := params["schema"].(string)
        
        switch schemaType {
        case "product":
            var product Product
            err := provider.GenerateWithSchema(ctx.Context, text, &product)
            return product, err
        case "person":
            var person Person
            err := provider.GenerateWithSchema(ctx.Context, text, &person)
            return person, err
        }
        
        return nil, fmt.Errorf("unknown schema type: %s", schemaType)
    }).
    Build()

agent.AddTool(extractTool)
```

## Next Steps

Now that you understand structured output:
- Explore [Schema Validation](../api/schema.md) for advanced validation
- Learn about [Custom Validators](advanced-validation.md) for complex rules
- See [Examples Gallery](examples-gallery.md) for more patterns

Ready to get reliable, structured data from LLMs? Let's structure! 📊