# Structured Data: Reliable Data Extraction with Schemas

> **[Project Root](/) / [Documentation](../..) / [User Guide](../../user-guide) / [Guides](../../user-guide/guides) / Structured Data**

Master reliable data extraction from LLMs using JSON Schema validation, type coercion, and error recovery. Build production systems that guarantee consistent, validated output from any LLM provider.

## Why Structured Data Matters

- **Reliability** - Guaranteed output format instead of unpredictable text
- **Validation** - Automatic validation against JSON Schema
- **Type Safety** - Direct conversion to Go structs with validation
- **Error Recovery** - Automatic retry and correction mechanisms
- **Production Ready** - Handle edge cases and malformed responses

## Core Concepts

![Schema Validation Flow](../../images/schema-validation.svg)

### Schema Definition
JSON Schema that defines the expected structure, types, and validation rules for your data.

### Type Coercion
Automatic conversion between types (string ↔ number, arrays, etc.) to handle LLM inconsistencies.

### Validation Pipeline
Multi-stage validation: schema → business rules → type safety → final output.

### Error Recovery
Automatic retry with improved prompts when validation fails.

## Prerequisites

- [Creating Agents guide completed](creating-agents.md) ✅
- [Building Data Extractors](building-data-extractors.md) helpful ✅
- Basic JSON Schema knowledge ✅

---

## Level 1: Basic Schema Usage
*Get started with simple schema validation*

### Simple Product Schema
```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/schema/domain" as schemaDomain
    "github.com/lexlapax/go-llms/pkg/schema/validation"
    "github.com/lexlapax/go-llms/pkg/structured/processor"
)

// Product represents a product with validation
type Product struct {
    Name        string  `json:"name" validate:"required,min=2"`
    Price       float64 `json:"price" validate:"required,gt=0"`
    Category    string  `json:"category" validate:"required"`
    Description string  `json:"description"`
    InStock     bool    `json:"in_stock"`
    Rating      float64 `json:"rating" validate:"gte=0,lte=5"`
}

func main() {
    fmt.Println("📋 Basic Schema Usage")
    fmt.Println("====================")

    // Create agent
    agent, err := core.NewAgentFromString("product-extractor", "openai/gpt-4o-mini")
    if err != nil {
        log.Fatalf("Failed to create agent: %v", err)
    }

    agent.SetSystemPrompt(`You are a product information extractor.
    Extract product details from descriptions and return valid JSON.
    
    Requirements:
    - name: Product name (required, at least 2 characters)
    - price: Price in USD (required, must be positive)
    - category: Product category (required)
    - description: Detailed description
    - in_stock: Availability status (boolean)
    - rating: Average rating from 0-5`)

    // Define product schema
    productSchema := &schemaDomain.Schema{
        Type: "object",
        Properties: map[string]schemaDomain.Property{
            "name": {
                Type:        "string",
                Description: "Product name",
                MinLength:   intPtr(2),
            },
            "price": {
                Type:        "number",
                Description: "Price in USD",
                Minimum:     float64Ptr(0.01),
            },
            "category": {
                Type:        "string",
                Description: "Product category",
                Enum:        []interface{}{"electronics", "clothing", "books", "home", "sports", "other"},
            },
            "description": {
                Type:        "string",
                Description: "Product description",
            },
            "in_stock": {
                Type:        "boolean",
                Description: "Whether product is in stock",
            },
            "rating": {
                Type:        "number",
                Description: "Average rating 0-5",
                Minimum:     float64Ptr(0),
                Maximum:     float64Ptr(5),
            },
        },
        Required: []string{"name", "price", "category"},
    }

    // Set schema on agent
    agent.SetSchema(productSchema)

    // Product descriptions to extract from
    descriptions := []string{
        "The iPhone 15 Pro is available for $999. It's an amazing smartphone with a 4.8-star rating. Great for photography and gaming. Currently in stock.",
        
        "Vintage leather jacket, $250, perfect condition. Fashion category. Classic style that never goes out of fashion. Limited stock available. 4.2 out of 5 stars.",
        
        "Programming Go book by expert authors. $45 price point. Educational book about Go programming language. Highly rated at 4.9 stars. Available now.",
        
        "Wireless Bluetooth headphones - $79.99. Electronics category. Noise-canceling with 20-hour battery. Customer rating: 4.1/5. Out of stock.",
    }

    state := domain.NewState()

    for i, description := range descriptions {
        fmt.Printf("\n--- Product %d ---\n", i+1)
        fmt.Printf("Description: %s\n", description)

        state.Set("user_input", fmt.Sprintf("Extract product information: %s", description))

        result, err := agent.Run(context.Background(), state)
        if err != nil {
            fmt.Printf("❌ Extraction failed: %v\n", err)
            continue
        }

        if structured, exists := result.Get("structured_output"); exists {
            // Convert to Product struct
            var product Product
            if data, ok := structured.(map[string]interface{}); ok {
                jsonData, _ := json.Marshal(data)
                if err := json.Unmarshal(jsonData, &product); err != nil {
                    fmt.Printf("❌ Type conversion failed: %v\n", err)
                    continue
                }

                fmt.Printf("✅ Extracted Product:\n")
                fmt.Printf("   Name: %s\n", product.Name)
                fmt.Printf("   Price: $%.2f\n", product.Price)
                fmt.Printf("   Category: %s\n", product.Category)
                fmt.Printf("   In Stock: %t\n", product.InStock)
                fmt.Printf("   Rating: %.1f/5\n", product.Rating)
                if product.Description != "" {
                    fmt.Printf("   Description: %s\n", truncateString(product.Description, 60))
                }
            }
        }
    }
}

// Helper functions
func float64Ptr(v float64) *float64 {
    return &v
}

func intPtr(v int) *int {
    return &v
}

func truncateString(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    return s[:maxLen] + "..."
}
```

---

## Level 2: Advanced Schema Features
*Complex validation rules and nested structures*

### Contact Management System
```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/schema/domain" as schemaDomain
    "github.com/lexlapax/go-llms/pkg/schema/validation"
)

// Address represents a structured address
type Address struct {
    Street   string `json:"street" validate:"required"`
    City     string `json:"city" validate:"required"`
    State    string `json:"state" validate:"required,len=2"`
    ZipCode  string `json:"zip_code" validate:"required,len=5"`
    Country  string `json:"country" validate:"required"`
}

// ContactInfo represents contact information
type ContactInfo struct {
    Email     string   `json:"email" validate:"required,email"`
    Phone     string   `json:"phone" validate:"required"`
    LinkedIn  string   `json:"linkedin,omitempty" validate:"omitempty,url"`
    Languages []string `json:"languages" validate:"required,min=1"`
}

// Company represents company information
type Company struct {
    Name     string `json:"name" validate:"required"`
    Industry string `json:"industry" validate:"required"`
    Size     string `json:"size" validate:"required,oneof=startup small medium large enterprise"`
    Website  string `json:"website,omitempty" validate:"omitempty,url"`
}

// FullContact represents a complete contact with nested structures
type FullContact struct {
    ID          string      `json:"id" validate:"required,uuid4"`
    FirstName   string      `json:"first_name" validate:"required,min=1"`
    LastName    string      `json:"last_name" validate:"required,min=1"`
    Title       string      `json:"title" validate:"required"`
    Contact     ContactInfo `json:"contact" validate:"required"`
    Address     Address     `json:"address" validate:"required"`
    Company     Company     `json:"company" validate:"required"`
    Skills      []string    `json:"skills" validate:"required,min=1"`
    Experience  int         `json:"experience_years" validate:"gte=0,lte=50"`
    Salary      *float64    `json:"salary,omitempty" validate:"omitempty,gt=0"`
    Notes       string      `json:"notes"`
    CreatedAt   time.Time   `json:"created_at"`
    UpdatedAt   time.Time   `json:"updated_at"`
}

func main() {
    fmt.Println("🏢 Advanced Schema Features - Nested Structures")
    fmt.Println("===============================================")

    // Create advanced contact extractor
    agent, err := core.NewAgentFromString("contact-extractor", "anthropic/claude-3-5-sonnet")
    if err != nil {
        log.Fatalf("Failed to create agent: %v", err)
    }

    agent.SetSystemPrompt(`You are an advanced contact information extractor.
    Extract comprehensive contact details and return structured JSON.
    
    Requirements:
    - Generate a valid UUID for ID
    - Extract all available information from the text
    - Use proper formatting for addresses (2-letter state codes, 5-digit zip)
    - Validate email formats and URLs
    - Categorize company size: startup, small, medium, large, enterprise
    - Set created_at and updated_at to current timestamp
    - Include relevant skills based on title/industry`)

    // Define complex nested schema
    contactSchema := &schemaDomain.Schema{
        Type: "object",
        Properties: map[string]schemaDomain.Property{
            "id": {
                Type:        "string",
                Description: "Unique UUID identifier",
                Format:      "uuid",
            },
            "first_name": {
                Type:        "string",
                Description: "First name",
                MinLength:   intPtr(1),
            },
            "last_name": {
                Type:        "string",
                Description: "Last name", 
                MinLength:   intPtr(1),
            },
            "title": {
                Type:        "string",
                Description: "Job title",
            },
            "contact": {
                Type:        "object",
                Description: "Contact information",
                Properties: map[string]schemaDomain.Property{
                    "email": {
                        Type:        "string",
                        Format:      "email",
                        Description: "Email address",
                    },
                    "phone": {
                        Type:        "string",
                        Description: "Phone number",
                    },
                    "linkedin": {
                        Type:        "string",
                        Format:      "uri",
                        Description: "LinkedIn profile URL",
                    },
                    "languages": {
                        Type:        "array",
                        Items:       &schemaDomain.Property{Type: "string"},
                        Description: "Languages spoken",
                        MinItems:    intPtr(1),
                    },
                },
                Required: []string{"email", "phone", "languages"},
            },
            "address": {
                Type:        "object",
                Description: "Physical address",
                Properties: map[string]schemaDomain.Property{
                    "street": {
                        Type:        "string",
                        Description: "Street address",
                    },
                    "city": {
                        Type:        "string",
                        Description: "City",
                    },
                    "state": {
                        Type:        "string",
                        Description: "State (2-letter code)",
                        MinLength:   intPtr(2),
                        MaxLength:   intPtr(2),
                    },
                    "zip_code": {
                        Type:        "string",
                        Description: "ZIP code (5 digits)",
                        Pattern:     "^[0-9]{5}$",
                    },
                    "country": {
                        Type:        "string",
                        Description: "Country",
                    },
                },
                Required: []string{"street", "city", "state", "zip_code", "country"},
            },
            "company": {
                Type:        "object",
                Description: "Company information",
                Properties: map[string]schemaDomain.Property{
                    "name": {
                        Type:        "string",
                        Description: "Company name",
                    },
                    "industry": {
                        Type:        "string",
                        Description: "Industry sector",
                    },
                    "size": {
                        Type:        "string",
                        Description: "Company size category",
                        Enum:        []interface{}{"startup", "small", "medium", "large", "enterprise"},
                    },
                    "website": {
                        Type:        "string",
                        Format:      "uri",
                        Description: "Company website",
                    },
                },
                Required: []string{"name", "industry", "size"},
            },
            "skills": {
                Type:        "array",
                Items:       &schemaDomain.Property{Type: "string"},
                Description: "Professional skills",
                MinItems:    intPtr(1),
            },
            "experience_years": {
                Type:        "integer",
                Description: "Years of experience",
                Minimum:     float64Ptr(0),
                Maximum:     float64Ptr(50),
            },
            "salary": {
                Type:        "number",
                Description: "Annual salary (optional)",
                Minimum:     float64Ptr(0),
            },
            "notes": {
                Type:        "string",
                Description: "Additional notes",
            },
            "created_at": {
                Type:        "string",
                Format:      "date-time",
                Description: "Creation timestamp",
            },
            "updated_at": {
                Type:        "string",
                Format:      "date-time",
                Description: "Update timestamp",
            },
        },
        Required: []string{"id", "first_name", "last_name", "title", "contact", "address", "company", "skills", "experience_years", "created_at", "updated_at"},
    }

    // Set schema on agent
    agent.SetSchema(contactSchema)

    // Complex contact descriptions
    descriptions := []string{
        `Sarah Johnson is the Chief Technology Officer at TechStart Inc., a medium-sized fintech startup. 
         Her email is sarah.johnson@techstart.com and phone is +1-555-0123. 
         She lives at 123 Main Street, San Francisco, CA 94105, USA.
         LinkedIn: https://linkedin.com/in/sarahjohnson
         She has 12 years of experience in software engineering, specializing in Go, Python, and cloud architecture.
         Languages: English, Spanish. Estimated salary around $180,000.
         Company website: https://techstart.com`,

        `Dr. Michael Chen, Senior Data Scientist at GlobalCorp (large enterprise, healthcare industry).
         Contact: michael.chen@globalcorp.com, 555-987-6543
         Address: 456 Oak Avenue, Boston, MA 02101, United States
         15 years experience in machine learning, statistics, and data analysis.
         Skills: Python, R, SQL, TensorFlow, PyTorch
         Speaks English, Mandarin, Japanese
         Company: www.globalcorp.com`,

        `Emily Rodriguez, Marketing Manager at LocalBiz (small business, retail industry)
         Email: emily@localbiz.com, Phone: 555-555-0199
         Lives in Austin, TX 78701 at 789 Pine Street, USA
         5 years marketing experience
         Skills: Digital marketing, SEO, social media, content creation
         Languages: English, French
         No LinkedIn profile available`,
    }

    state := domain.NewState()

    for i, description := range descriptions {
        fmt.Printf("\n--- Contact %d ---\n", i+1)
        fmt.Printf("Source: %s\n", truncateString(description, 100))

        state.Set("user_input", fmt.Sprintf("Extract complete contact information: %s", description))

        result, err := agent.Run(context.Background(), state)
        if err != nil {
            fmt.Printf("❌ Extraction failed: %v\n", err)
            continue
        }

        if structured, exists := result.Get("structured_output"); exists {
            // Convert to FullContact struct
            var contact FullContact
            if data, ok := structured.(map[string]interface{}); ok {
                jsonData, _ := json.Marshal(data)
                if err := json.Unmarshal(jsonData, &contact); err != nil {
                    fmt.Printf("❌ Type conversion failed: %v\n", err)
                    continue
                }

                fmt.Printf("✅ Extracted Contact:\n")
                fmt.Printf("   ID: %s\n", contact.ID)
                fmt.Printf("   Name: %s %s (%s)\n", contact.FirstName, contact.LastName, contact.Title)
                fmt.Printf("   Email: %s | Phone: %s\n", contact.Contact.Email, contact.Contact.Phone)
                fmt.Printf("   Company: %s (%s, %s)\n", contact.Company.Name, contact.Company.Industry, contact.Company.Size)
                fmt.Printf("   Location: %s, %s %s\n", contact.Address.City, contact.Address.State, contact.Address.ZipCode)
                fmt.Printf("   Experience: %d years\n", contact.Experience)
                fmt.Printf("   Skills: %v\n", contact.Skills)
                fmt.Printf("   Languages: %v\n", contact.Contact.Languages)
                if contact.Salary != nil {
                    fmt.Printf("   Salary: $%.0f\n", *contact.Salary)
                }
                if contact.Contact.LinkedIn != "" {
                    fmt.Printf("   LinkedIn: %s\n", contact.Contact.LinkedIn)
                }
            }
        }
    }
}
```

---

## Level 3: Type Coercion and Error Recovery
*Handle inconsistent LLM outputs with automatic correction*

### Type Coercion Engine
```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "strconv"
    "strings"
    "time"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/schema/domain" as schemaDomain
    "github.com/lexlapax/go-llms/pkg/schema/validation"
    "github.com/lexlapax/go-llms/pkg/structured/processor"
)

// SalesData represents sales information with various data types
type SalesData struct {
    OrderID     string    `json:"order_id"`
    CustomerID  int       `json:"customer_id"`
    ProductName string    `json:"product_name"`
    Quantity    int       `json:"quantity"`
    UnitPrice   float64   `json:"unit_price"`
    TotalAmount float64   `json:"total_amount"`
    OrderDate   time.Time `json:"order_date"`
    IsRush      bool      `json:"is_rush_order"`
    Region      string    `json:"region"`
    SalesRep    string    `json:"sales_rep"`
    Commission  *float64  `json:"commission,omitempty"`
    Tags        []string  `json:"tags"`
}

// CoercionEngine handles type coercion and validation
type CoercionEngine struct {
    validator *validation.Validator
    processor *processor.StructuredProcessor
}

// NewCoercionEngine creates a new coercion engine
func NewCoercionEngine() *CoercionEngine {
    validator := validation.NewValidator()
    structProcessor := processor.NewStructuredProcessor(validator)
    
    return &CoercionEngine{
        validator: validator,
        processor: structProcessor,
    }
}

// CoerceAndValidate performs type coercion and validation
func (ce *CoercionEngine) CoerceAndValidate(schema *schemaDomain.Schema, rawData map[string]interface{}) (map[string]interface{}, error) {
    coercedData := make(map[string]interface{})
    
    for key, value := range rawData {
        property, exists := schema.Properties[key]
        if !exists {
            continue // Skip unknown properties
        }
        
        coercedValue, err := ce.coerceValue(value, property)
        if err != nil {
            return nil, fmt.Errorf("coercion failed for %s: %w", key, err)
        }
        
        coercedData[key] = coercedValue
    }
    
    // Validate coerced data
    err := ce.processor.ValidateData(schema, coercedData)
    if err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    return coercedData, nil
}

// coerceValue coerces a value to the expected type
func (ce *CoercionEngine) coerceValue(value interface{}, property schemaDomain.Property) (interface{}, error) {
    switch property.Type {
    case "string":
        return ce.coerceToString(value), nil
    case "integer":
        return ce.coerceToInt(value)
    case "number":
        return ce.coerceToFloat(value)
    case "boolean":
        return ce.coerceToBool(value)
    case "array":
        return ce.coerceToArray(value, property)
    default:
        return value, nil
    }
}

// coerceToString converts various types to string
func (ce *CoercionEngine) coerceToString(value interface{}) string {
    switch v := value.(type) {
    case string:
        return strings.TrimSpace(v)
    case int:
        return strconv.Itoa(v)
    case float64:
        return strconv.FormatFloat(v, 'f', -1, 64)
    case bool:
        return strconv.FormatBool(v)
    default:
        return fmt.Sprintf("%v", v)
    }
}

// coerceToInt converts various types to integer
func (ce *CoercionEngine) coerceToInt(value interface{}) (int, error) {
    switch v := value.(type) {
    case int:
        return v, nil
    case float64:
        return int(v), nil
    case string:
        // Try to parse as integer
        if i, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
            return i, nil
        }
        // Try to parse as float then convert
        if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
            return int(f), nil
        }
        return 0, fmt.Errorf("cannot convert string '%s' to integer", v)
    case bool:
        if v {
            return 1, nil
        }
        return 0, nil
    default:
        return 0, fmt.Errorf("cannot convert %T to integer", value)
    }
}

// coerceToFloat converts various types to float64
func (ce *CoercionEngine) coerceToFloat(value interface{}) (float64, error) {
    switch v := value.(type) {
    case float64:
        return v, nil
    case int:
        return float64(v), nil
    case string:
        // Remove common currency symbols and formatting
        cleaned := strings.TrimSpace(v)
        cleaned = strings.ReplaceAll(cleaned, "$", "")
        cleaned = strings.ReplaceAll(cleaned, ",", "")
        
        if f, err := strconv.ParseFloat(cleaned, 64); err == nil {
            return f, nil
        }
        return 0, fmt.Errorf("cannot convert string '%s' to float", v)
    case bool:
        if v {
            return 1.0, nil
        }
        return 0.0, nil
    default:
        return 0, fmt.Errorf("cannot convert %T to float", value)
    }
}

// coerceToBool converts various types to boolean
func (ce *CoercionEngine) coerceToBool(value interface{}) (bool, error) {
    switch v := value.(type) {
    case bool:
        return v, nil
    case int:
        return v != 0, nil
    case float64:
        return v != 0, nil
    case string:
        cleaned := strings.ToLower(strings.TrimSpace(v))
        switch cleaned {
        case "true", "yes", "y", "1", "on", "rush", "urgent", "priority":
            return true, nil
        case "false", "no", "n", "0", "off", "normal", "standard":
            return false, nil
        default:
            return false, fmt.Errorf("cannot convert string '%s' to boolean", v)
        }
    default:
        return false, fmt.Errorf("cannot convert %T to boolean", value)
    }
}

// coerceToArray converts various types to array
func (ce *CoercionEngine) coerceToArray(value interface{}, property schemaDomain.Property) ([]interface{}, error) {
    switch v := value.(type) {
    case []interface{}:
        return v, nil
    case string:
        // Split string by common delimiters
        parts := strings.FieldsFunc(v, func(r rune) bool {
            return r == ',' || r == ';' || r == '|'
}
        
        result := make([]interface{}, len(parts))
        for i, part := range parts {
            result[i] = strings.TrimSpace(part)
        }
        return result, nil
    default:
        // Single value becomes array with one element
        return []interface{}{value}, nil
    }
}

func main() {
    fmt.Println("🔄 Type Coercion and Error Recovery")
    fmt.Println("===================================")

    // Create sales data extractor
    agent, err := core.NewAgentFromString("sales-extractor", "openai/gpt-4o")
    if err != nil {
        log.Fatalf("Failed to create agent: %v", err)
    }

    agent.SetSystemPrompt(`You are a sales data extractor.
    Extract sales information from text and return JSON.
    Be flexible with data types - the system will handle coercion.
    
    Fields to extract:
    - order_id: Order identifier (string)
    - customer_id: Customer number (integer)
    - product_name: Product name (string)
    - quantity: Number of items (integer)
    - unit_price: Price per item (number)
    - total_amount: Total order value (number)
    - order_date: Order date (ISO date string)
    - is_rush_order: Rush order flag (boolean)
    - region: Sales region (string)
    - sales_rep: Sales representative name (string)
    - commission: Commission amount if mentioned (number, optional)
    - tags: Tags or categories (array of strings)`)

    // Define sales schema
    salesSchema := &schemaDomain.Schema{
        Type: "object",
        Properties: map[string]schemaDomain.Property{
            "order_id": {
                Type:        "string",
                Description: "Order identifier",
            },
            "customer_id": {
                Type:        "integer",
                Description: "Customer ID number",
            },
            "product_name": {
                Type:        "string",
                Description: "Product name",
            },
            "quantity": {
                Type:        "integer",
                Description: "Quantity ordered",
                Minimum:     float64Ptr(1),
            },
            "unit_price": {
                Type:        "number",
                Description: "Price per unit",
                Minimum:     float64Ptr(0),
            },
            "total_amount": {
                Type:        "number",
                Description: "Total order amount",
                Minimum:     float64Ptr(0),
            },
            "order_date": {
                Type:        "string",
                Format:      "date-time",
                Description: "Order date",
            },
            "is_rush_order": {
                Type:        "boolean",
                Description: "Rush order flag",
            },
            "region": {
                Type:        "string",
                Description: "Sales region",
                Enum:        []interface{}{"North", "South", "East", "West", "Central"},
            },
            "sales_rep": {
                Type:        "string",
                Description: "Sales representative",
            },
            "commission": {
                Type:        "number",
                Description: "Commission amount",
                Minimum:     float64Ptr(0),
            },
            "tags": {
                Type:        "array",
                Items:       &schemaDomain.Property{Type: "string"},
                Description: "Order tags",
            },
        },
        Required: []string{"order_id", "customer_id", "product_name", "quantity", "unit_price", "total_amount", "order_date", "is_rush_order", "region", "sales_rep", "tags"},
    }

    // Set schema on agent
    agent.SetSchema(salesSchema)

    // Create coercion engine
    coercionEngine := NewCoercionEngine()

    // Sales order descriptions with various data format issues
    descriptions := []string{
        `Order ORD-12345 from customer 1001 for "Wireless Headphones".
         Quantity: 2, Unit price: $79.99, Total: $159.98
         Order date: 2024-01-15, Rush order: YES
         Region: North, Sales rep: John Smith
         Commission: $15.99, Tags: electronics, audio, premium`,

        `Customer #2005 ordered 5 units of "Gaming Mouse" at 45.50 each = 227.50 total
         Order ID: ORD-12346, Date: January 20, 2024
         Not rush, West region, Sarah Johnson handling
         Tags: gaming|computer|accessories`,

        `ORD-12347: Customer 3010, "Office Chair" x 1 @ $299
         Total amount: $299.00, Date: 2024-01-25T10:30:00Z
         Standard priority (not rush), Central region
         Rep: Mike Davis, Commission not specified
         Categories: furniture; office; ergonomic`,

        `Order ORD-12348, Cust: 4020
         Product: "Smartphone Case", Qty: 3, Price: 19.99 per unit
         Total: 59.97, Date: 01/30/2024, URGENT delivery
         East region, handled by Lisa Wong
         Tags: mobile,accessories,protection`,
    }

    state := domain.NewState()

    for i, description := range descriptions {
        fmt.Printf("\n--- Sales Order %d ---\n", i+1)
        fmt.Printf("Description: %s\n", truncateString(description, 100))

        state.Set("user_input", fmt.Sprintf("Extract sales data: %s", description))

        result, err := agent.Run(context.Background(), state)
        if err != nil {
            fmt.Printf("❌ Extraction failed: %v\n", err)
            continue
        }

        if structured, exists := result.Get("structured_output"); exists {
            if rawData, ok := structured.(map[string]interface{}); ok {
                fmt.Printf("📥 Raw extracted data:\n")
                for key, value := range rawData {
                    fmt.Printf("   %s: %v (%T)\n", key, value, value)
                }

                // Apply type coercion
                coercedData, err := coercionEngine.CoerceAndValidate(salesSchema, rawData)
                if err != nil {
                    fmt.Printf("❌ Coercion failed: %v\n", err)
                    continue
                }

                // Convert to SalesData struct
                var salesData SalesData
                jsonData, _ := json.Marshal(coercedData)
                if err := json.Unmarshal(jsonData, &salesData); err != nil {
                    fmt.Printf("❌ Struct conversion failed: %v\n", err)
                    continue
                }

                fmt.Printf("✅ Coerced and validated data:\n")
                fmt.Printf("   Order ID: %s\n", salesData.OrderID)
                fmt.Printf("   Customer: %d\n", salesData.CustomerID)
                fmt.Printf("   Product: %s\n", salesData.ProductName)
                fmt.Printf("   Quantity: %d × $%.2f = $%.2f\n", salesData.Quantity, salesData.UnitPrice, salesData.TotalAmount)
                fmt.Printf("   Date: %s\n", salesData.OrderDate.Format("2006-01-02 15:04"))
                fmt.Printf("   Rush Order: %t\n", salesData.IsRush)
                fmt.Printf("   Region: %s | Rep: %s\n", salesData.Region, salesData.SalesRep)
                if salesData.Commission != nil {
                    fmt.Printf("   Commission: $%.2f\n", *salesData.Commission)
                }
                fmt.Printf("   Tags: %v\n", salesData.Tags)
            }
        }
    }
}
```

---

## Level 4: Production Schema Management
*Version control, evolution, and enterprise patterns*

### Schema Registry and Versioning
```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/schema/domain" as schemaDomain
    "github.com/lexlapax/go-llms/pkg/schema/validation"
    "github.com/lexlapax/go-llms/pkg/structured/processor"
)

// SchemaVersion represents a versioned schema
type SchemaVersion struct {
    Name        string                 `json:"name"`
    Version     string                 `json:"version"`
    Schema      *schemaDomain.Schema   `json:"schema"`
    Description string                 `json:"description"`
    CreatedAt   time.Time              `json:"created_at"`
    Deprecated  bool                   `json:"deprecated"`
    Migration   map[string]interface{} `json:"migration,omitempty"`
}

// SchemaRegistry manages schema versions and evolution
type SchemaRegistry struct {
    schemas map[string]map[string]*SchemaVersion // name -> version -> schema
    latest  map[string]string                    // name -> latest version
    mu      sync.RWMutex
}

// NewSchemaRegistry creates a new schema registry
func NewSchemaRegistry() *SchemaRegistry {
    return &SchemaRegistry{
        schemas: make(map[string]map[string]*SchemaVersion),
        latest:  make(map[string]string),
    }
}

// RegisterSchema registers a new schema version
func (sr *SchemaRegistry) RegisterSchema(name, version, description string, schema *schemaDomain.Schema) error {
    sr.mu.Lock()
    defer sr.mu.Unlock()

    if sr.schemas[name] == nil {
        sr.schemas[name] = make(map[string]*SchemaVersion)
    }

    if _, exists := sr.schemas[name][version]; exists {
        return fmt.Errorf("schema %s version %s already exists", name, version)
    }

    schemaVersion := &SchemaVersion{
        Name:        name,
        Version:     version,
        Schema:      schema,
        Description: description,
        CreatedAt:   time.Now(),
        Deprecated:  false,
    }

    sr.schemas[name][version] = schemaVersion
    sr.latest[name] = version

    return nil
}

// GetSchema retrieves a specific schema version
func (sr *SchemaRegistry) GetSchema(name, version string) (*SchemaVersion, error) {
    sr.mu.RLock()
    defer sr.mu.RUnlock()

    if version == "latest" {
        if latestVersion, exists := sr.latest[name]; exists {
            version = latestVersion
        } else {
            return nil, fmt.Errorf("no schemas found for %s", name)
        }
    }

    if versions, exists := sr.schemas[name]; exists {
        if schema, exists := versions[version]; exists {
            return schema, nil
        }
    }

    return nil, fmt.Errorf("schema %s version %s not found", name, version)
}

// ListSchemas lists all registered schemas
func (sr *SchemaRegistry) ListSchemas() map[string][]string {
    sr.mu.RLock()
    defer sr.mu.RUnlock()

    result := make(map[string][]string)
    for name, versions := range sr.schemas {
        versionList := make([]string, 0, len(versions))
        for version := range versions {
            versionList = append(versionList, version)
        }
        result[name] = versionList
    }

    return result
}

// MigrateData migrates data from one schema version to another
func (sr *SchemaRegistry) MigrateData(name, fromVersion, toVersion string, data map[string]interface{}) (map[string]interface{}, error) {
    // Simplified migration logic - in production, implement proper migration rules
    return data, nil
}

// ProductionExtractor handles schema-driven extraction with versioning
type ProductionExtractor struct {
    registry *SchemaRegistry
    agent    domain.BaseAgent
    processor *processor.StructuredProcessor
}

// NewProductionExtractor creates a production-ready extractor
func NewProductionExtractor(registry *SchemaRegistry) (*ProductionExtractor, error) {
    agent, err := core.NewAgentFromString("production-extractor", "anthropic/claude-3-5-sonnet")
    if err != nil {
        return nil, fmt.Errorf("failed to create agent: %w", err)
    }

    agent.SetSystemPrompt(`You are a production data extraction system.
    You extract structured data according to specific schemas.
    Always return valid JSON that matches the provided schema exactly.
    Handle edge cases gracefully and provide reasonable defaults when needed.`)

    validator := validation.NewValidator()
    structProcessor := processor.NewStructuredProcessor(validator)

    return &ProductionExtractor{
        registry:  registry,
        agent:     agent,
        processor: structProcessor,
    }, nil
}

// Extract performs schema-driven extraction
func (pe *ProductionExtractor) Extract(ctx context.Context, schemaName, schemaVersion, text string) (map[string]interface{}, error) {
    // Get schema
    schemaVersion_obj, err := pe.registry.GetSchema(schemaName, schemaVersion)
    if err != nil {
        return nil, fmt.Errorf("schema not found: %w", err)
    }

    if schemaVersion_obj.Deprecated {
        log.Printf("Warning: Using deprecated schema %s v%s", schemaName, schemaVersion)
    }

    // Set schema on agent
    pe.agent.SetSchema(schemaVersion_obj.Schema)

    // Extract data
    state := domain.NewState()
    state.Set("user_input", fmt.Sprintf("Extract data according to %s schema v%s: %s", schemaName, schemaVersion, text))

    result, err := pe.agent.Run(ctx, state)
    if err != nil {
        return nil, fmt.Errorf("extraction failed: %w", err)
    }

    if structured, exists := result.Get("structured_output"); exists {
        if data, ok := structured.(map[string]interface{}); ok {
            return data, nil
        }
    }

    return nil, fmt.Errorf("no structured output returned")
}

func main() {
    fmt.Println("🏭 Production Schema Management")
    fmt.Println("==============================")

    // Create schema registry
    registry := NewSchemaRegistry()

    // Register Customer schema v1.0
    customerSchemaV1 := &schemaDomain.Schema{
        Type: "object",
        Properties: map[string]schemaDomain.Property{
            "name": {
                Type:        "string",
                Description: "Customer name",
            },
            "email": {
                Type:        "string",
                Format:      "email",
                Description: "Email address",
            },
            "phone": {
                Type:        "string",
                Description: "Phone number",
            },
            "company": {
                Type:        "string",
                Description: "Company name",
            },
        },
        Required: []string{"name", "email"},
    }

    err := registry.RegisterSchema("customer", "1.0", "Initial customer schema", customerSchemaV1)
    if err != nil {
        log.Fatalf("Failed to register schema: %v", err)
    }

    // Register Customer schema v1.1 (added address)
    customerSchemaV11 := &schemaDomain.Schema{
        Type: "object",
        Properties: map[string]schemaDomain.Property{
            "name": {
                Type:        "string",
                Description: "Customer name",
            },
            "email": {
                Type:        "string",
                Format:      "email",
                Description: "Email address",
            },
            "phone": {
                Type:        "string",
                Description: "Phone number",
            },
            "company": {
                Type:        "string",
                Description: "Company name",
            },
            "address": {
                Type:        "string",
                Description: "Physical address",
            },
            "priority": {
                Type:        "string",
                Description: "Customer priority level",
                Enum:        []interface{}{"low", "medium", "high"},
                Default:     "medium",
            },
        },
        Required: []string{"name", "email"},
    }

    err = registry.RegisterSchema("customer", "1.1", "Added address and priority fields", customerSchemaV11)
    if err != nil {
        log.Fatalf("Failed to register schema v1.1: %v", err)
    }

    // Register Product schema v2.0
    productSchemaV2 := &schemaDomain.Schema{
        Type: "object",
        Properties: map[string]schemaDomain.Property{
            "id": {
                Type:        "string",
                Description: "Product ID",
            },
            "name": {
                Type:        "string",
                Description: "Product name",
            },
            "price": {
                Type:        "number",
                Description: "Product price",
                Minimum:     float64Ptr(0),
            },
            "category": {
                Type:        "string",
                Description: "Product category",
            },
            "in_stock": {
                Type:        "boolean",
                Description: "Stock availability",
            },
            "tags": {
                Type:        "array",
                Items:       &schemaDomain.Property{Type: "string"},
                Description: "Product tags",
            },
        },
        Required: []string{"id", "name", "price", "category", "in_stock"},
    }

    err = registry.RegisterSchema("product", "2.0", "Enhanced product schema with tags", productSchemaV2)
    if err != nil {
        log.Fatalf("Failed to register product schema: %v", err)
    }

    // Create production extractor
    extractor, err := NewProductionExtractor(registry)
    if err != nil {
        log.Fatalf("Failed to create extractor: %v", err)
    }

    // Display registered schemas
    fmt.Println("📋 Registered Schemas:")
    schemas := registry.ListSchemas()
    for name, versions := range schemas {
        fmt.Printf("  %s: %v\n", name, versions)
    }

    // Test extraction with different schema versions
    testCases := []struct {
        schemaName    string
        schemaVersion string
        text          string
    }{
        {
            schemaName:    "customer",
            schemaVersion: "1.0",
            text:          "John Doe, email: john@example.com, phone: 555-0123, works at TechCorp",
        },
        {
            schemaName:    "customer",
            schemaVersion: "1.1",
            text:          "Sarah Smith, sarah@startup.io, 555-0456, StartupCorp, 123 Main St, high priority customer",
        },
        {
            schemaName:    "customer",
            schemaVersion: "latest",
            text:          "Mike Johnson, mike@company.com, 555-0789, BigCorp, medium priority",
        },
        {
            schemaName:    "product",
            schemaVersion: "2.0",
            text:          "Product ABC-123: Wireless Headphones, $79.99, Electronics category, in stock, tags: audio, wireless, premium",
        },
    }

    for i, testCase := range testCases {
        fmt.Printf("\n--- Extraction Test %d ---\n", i+1)
        fmt.Printf("Schema: %s v%s\n", testCase.schemaName, testCase.schemaVersion)
        fmt.Printf("Text: %s\n", testCase.text)

        result, err := extractor.Extract(context.Background(), testCase.schemaName, testCase.schemaVersion, testCase.text)
        if err != nil {
            fmt.Printf("❌ Extraction failed: %v\n", err)
            continue
        }

        fmt.Printf("✅ Extracted data:\n")
        resultJSON, _ := json.MarshalIndent(result, "", "  ")
        fmt.Println(string(resultJSON))
    }

    // Demonstrate schema information retrieval
    fmt.Printf("\n📊 Schema Details:\n")
    customerV11, _ := registry.GetSchema("customer", "1.1")
    fmt.Printf("Customer v1.1: %s (created: %s)\n", 
        customerV11.Description, 
        customerV11.CreatedAt.Format("2006-01-02"))
}
```

## Schema Best Practices

### 1. Schema Design
- **Start Simple** - Begin with minimal required fields
- **Evolve Gradually** - Add fields incrementally with versioning
- **Document Everything** - Clear descriptions for all properties
- **Use Validation** - Leverage JSON Schema validation features

### 2. Type Strategy
- **Be Flexible** - Use type coercion for LLM inconsistencies
- **Validate Strictly** - Enforce business rules after coercion
- **Handle Nulls** - Plan for optional and missing data
- **Array Handling** - Support various array input formats

### 3. Error Recovery
- **Multiple Attempts** - Retry with improved prompts
- **Fallback Schemas** - Simpler schemas for difficult extractions
- **Partial Success** - Accept partial data when appropriate
- **Logging** - Comprehensive error logging for debugging

### 4. Production Considerations
- **Schema Versioning** - Version control for schema evolution
- **Backward Compatibility** - Support multiple schema versions
- **Performance** - Cache schemas and validation rules
- **Monitoring** - Track extraction success rates and errors

## Troubleshooting

### Common Issues

**Schema Validation Errors**
- Check property types and formats
- Verify required fields are present
- Review enum values and constraints
- Test with simplified schemas first

**Type Coercion Failures**
- Improve LLM prompts for specific types
- Add more coercion rules for edge cases
- Use more flexible schemas initially
- Implement custom coercion logic

**Inconsistent Extraction**
- Enhance system prompts with examples
- Use more structured prompts
- Implement retry logic with variations
- Consider multiple extraction attempts

**Performance Issues**
- Cache compiled schemas
- Use simpler schemas for high-volume extraction
- Implement extraction result caching
- Profile and optimize validation steps

## Next Steps

📋 **Schema mastery achieved!** Continue with:

- **[Data Validation](data-validation.md)** - Advanced validation techniques
- **[Building Data Extractors](building-data-extractors.md)** - Complete extraction systems
- **[Data Pipelines](data-pipelines.md)** - End-to-end processing workflows
- **[Agent Tools](agent-tools.md)** - Tool-based data processing

### Related Examples

- **[Structured Schema](../../cmd/examples/structured-schema/)** - Schema generation examples
- **[Structured Coercion](../../cmd/examples/structured-coercion/)** - Type coercion patterns
- **[Agent Structured Output](../../cmd/examples/agent-structured-output/)** - Agent integration

---

**Need help?** Check our [troubleshooting guide](../advanced/troubleshooting.md) or join the discussion on [GitHub](https://github.com/lexlapax/go-llms/discussions).