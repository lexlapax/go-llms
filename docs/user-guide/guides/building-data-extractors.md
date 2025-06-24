# Building Data Extractors: Reliable Data Processing Workflows

> **[Project Root](/) / [Documentation](../..) / [User Guide](../../user-guide) / [Guides](../../user-guide/guides) / Building Data Extractors**

Transform unstructured text into reliable, validated data with Go-LLMs' powerful extraction capabilities. Learn to build production-ready data processing systems that handle emails, documents, web content, and more.

## Why Go-LLMs for Data Extraction?

- **Schema Validation** - Guaranteed output format with JSON Schema
- **Error Recovery** - Automatic retry and validation mechanisms  
- **Type Safety** - Direct conversion to Go structs
- **Multiple Methods** - Agent-based, processor-based, and streaming extraction
- **Production Ready** - Built-in error handling and monitoring

## Learning Path

![Data Extraction Workflow](../../images/data-flow.svg)

1. **Basic Extraction** - Simple structured data from text (10 min)
2. **Schema-Based Extraction** - Validated data with custom schemas (15 min)
3. **Batch Processing** - Handle multiple documents efficiently (15 min)
4. **Error Recovery** - Handle malformed or incomplete data (10 min)
5. **Production Pipeline** - End-to-end data processing system (20 min)

## Prerequisites

- [First Steps completed](../getting-started/first-steps.md) ✅
- Basic understanding of JSON Schema ✅
- Go struct experience ✅

---

## Example 1: Customer Lead Extractor
*Extract structured customer information from sales emails*

### The Problem
Sales teams receive hundreds of emails daily. Manually extracting customer details is time-consuming and error-prone.

### The Solution
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

// Lead represents a potential customer
type Lead struct {
    Name        string   `json:"name" validate:"required"`
    Email       string   `json:"email" validate:"required,email"`
    Phone       string   `json:"phone"`
    Company     string   `json:"company"`
    Industry    string   `json:"industry"`
    Budget      string   `json:"budget"`
    Timeline    string   `json:"timeline"`
    Pain_points []string `json:"pain_points"`
    Priority    string   `json:"priority" validate:"oneof=low medium high urgent"`
    NextAction  string   `json:"next_action"`
}

func main() {
    // Create extraction agent
    extractor, err := core.NewAgentFromString("lead-extractor", "openai/gpt-4o-mini")
    if err != nil {
        log.Fatalf("Failed to create extractor: %v", err)
    }

    extractor.SetSystemPrompt(`You are a sales lead extraction specialist.
Extract complete customer information from sales communications.
Focus on identifying budget, timeline, and pain points.
Assign priority based on urgency and budget indicators.`)

    // Define schema for lead data
    leadSchema := &schemaDomain.Schema{
        Type: "object",
        Properties: map[string]schemaDomain.Property{
            "name":        {Type: "string", Description: "Contact's full name"},
            "email":       {Type: "string", Format: "email", Description: "Email address"},
            "phone":       {Type: "string", Description: "Phone number"},
            "company":     {Type: "string", Description: "Company name"},
            "industry":    {Type: "string", Description: "Industry sector"},
            "budget":      {Type: "string", Description: "Budget information or range"},
            "timeline":    {Type: "string", Description: "Project timeline or urgency"},
            "pain_points": {Type: "array", Items: &schemaDomain.Property{Type: "string"}, Description: "Business problems to solve"},
            "priority":    {Type: "string", Enum: []interface{}{"low", "medium", "high", "urgent"}, Description: "Lead priority"},
            "next_action": {Type: "string", Description: "Recommended next step"},
        },
        Required: []string{"name", "email", "priority"},
    }

    // Set schema on agent
    extractor.SetSchema(leadSchema)

    // Sample sales email
    salesEmail := `Subject: Urgent: Need CRM Solution ASAP

Hi there,

My name is Sarah Johnson from TechStart Inc. We're a fintech startup with 50 employees.
Our current CRM is falling apart and we need a replacement within 6 weeks.

We've allocated $25,000 for this project and need something that can:
- Handle complex sales pipelines
- Integrate with our existing tools
- Scale as we grow rapidly

Our biggest issues right now:
- Lost leads due to poor follow-up tracking
- No integration with our marketing automation
- Reports are manually created and often wrong
- Team can't collaborate effectively

Can you help? We need to make a decision by end of month.

Best regards,
Sarah Johnson
Head of Sales
sarah.johnson@techstart.com
+1-555-0123`

    // Extract lead data
    state := domain.NewState()
    state.Set("user_input", fmt.Sprintf("Extract lead information from this email: %s", salesEmail))

    result, err := extractor.Run(context.Background(), state)
    if err != nil {
        log.Fatalf("Extraction failed: %v", err)
    }

    // Get structured output
    if structured, exists := result.Get("structured_output"); exists {
        lead := &Lead{}
        if data, ok := structured.(map[string]interface{}); ok {
            jsonData, _ := json.Marshal(data)
            json.Unmarshal(jsonData, lead)
            
            fmt.Println("🎯 Extracted Lead Information:")
            fmt.Printf("Name: %s\n", lead.Name)
            fmt.Printf("Company: %s (%s)\n", lead.Company, lead.Industry)
            fmt.Printf("Contact: %s | %s\n", lead.Email, lead.Phone)
            fmt.Printf("Budget: %s\n", lead.Budget)
            fmt.Printf("Timeline: %s\n", lead.Timeline)
            fmt.Printf("Priority: %s\n", lead.Priority)
            fmt.Printf("Pain Points: %v\n", lead.Pain_points)
            fmt.Printf("Next Action: %s\n", lead.NextAction)
        }
    }
}
```

### Key Concepts

✅ **Schema Definition** - Define exact data structure expected  
✅ **Required Fields** - Specify which fields must be present  
✅ **Validation Rules** - Email format, enum values, etc.  
✅ **System Prompts** - Guide the extraction with specific instructions  
✅ **Type Safety** - Convert to Go structs automatically  

---

## Example 2: Document Analysis Pipeline
*Process multiple documents with batch validation*

### The Problem
Analyze hundreds of product reviews, support tickets, or research papers efficiently while maintaining data quality.

### The Solution
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

// ReviewAnalysis represents analyzed review data
type ReviewAnalysis struct {
    ProductName   string   `json:"product_name"`
    Rating        int      `json:"rating" validate:"min=1,max=5"`
    Sentiment     string   `json:"sentiment" validate:"oneof=positive negative neutral"`
    Themes        []string `json:"themes"`
    Issues        []string `json:"issues"`
    Suggestions   []string `json:"suggestions"`
    CustomerType  string   `json:"customer_type"`
    Verified      bool     `json:"verified"`
    TrustScore    float64  `json:"trust_score" validate:"min=0,max=1"`
    Summary       string   `json:"summary"`
}

// DocumentProcessor handles batch processing of documents
type DocumentProcessor struct {
    agent           domain.BaseAgent
    schema          *schemaDomain.Schema
    validator       *validation.Validator
    processor       *processor.StructuredProcessor
    maxConcurrency  int
}

func NewDocumentProcessor(maxConcurrency int) (*DocumentProcessor, error) {
    // Create analysis agent
    agent, err := core.NewAgentFromString("review-analyzer", "anthropic/claude-3-5-haiku")
    if err != nil {
        return nil, fmt.Errorf("failed to create agent: %w", err)
    }

    agent.SetSystemPrompt(`You are an expert product review analyzer.
Extract insights from customer reviews focusing on:
- Product strengths and weaknesses
- Customer sentiment and satisfaction
- Specific issues and suggestions
- Trust indicators (verified purchase, detail level, etc.)

Rate trust score from 0-1 based on review authenticity and helpfulness.`)

    // Define analysis schema
    schema := &schemaDomain.Schema{
        Type: "object",
        Properties: map[string]schemaDomain.Property{
            "product_name":  {Type: "string", Description: "Product being reviewed"},
            "rating":        {Type: "integer", Minimum: float64Ptr(1), Maximum: float64Ptr(5), Description: "Rating 1-5"},
            "sentiment":     {Type: "string", Enum: []interface{}{"positive", "negative", "neutral"}, Description: "Overall sentiment"},
            "themes":        {Type: "array", Items: &schemaDomain.Property{Type: "string"}, Description: "Main themes discussed"},
            "issues":        {Type: "array", Items: &schemaDomain.Property{Type: "string"}, Description: "Problems mentioned"},
            "suggestions":   {Type: "array", Items: &schemaDomain.Property{Type: "string"}, Description: "Improvement suggestions"},
            "customer_type": {Type: "string", Description: "Type of customer (professional, casual, expert, etc.)"},
            "verified":      {Type: "boolean", Description: "Appears to be verified purchase"},
            "trust_score":   {Type: "number", Minimum: float64Ptr(0), Maximum: float64Ptr(1), Description: "Review trustworthiness 0-1"},
            "summary":       {Type: "string", Description: "Brief summary of review"},
        },
        Required: []string{"product_name", "rating", "sentiment", "summary"},
    }

    agent.SetSchema(schema)

    // Create validator and processor
    validator := validation.NewValidator()
    structProcessor := processor.NewStructuredProcessor(validator)

    return &DocumentProcessor{
        agent:           agent,
        schema:          schema,
        validator:       validator,
        processor:       structProcessor,
        maxConcurrency:  maxConcurrency,
    }, nil
}

// ProcessDocuments processes multiple documents concurrently
func (dp *DocumentProcessor) ProcessDocuments(ctx context.Context, documents []string) ([]ReviewAnalysis, []error) {
    // Create semaphore for concurrency control
    semaphore := make(chan struct{}, dp.maxConcurrency)
    
    var wg sync.WaitGroup
    results := make([]ReviewAnalysis, len(documents))
    errors := make([]error, len(documents))

    // Process documents concurrently
    for i, doc := range documents {
        wg.Add(1)
        go func(index int, document string) {
            defer wg.Done()
            
            // Acquire semaphore
            semaphore <- struct{}{}
            defer func() { <-semaphore }()

            // Process single document
            analysis, err := dp.processDocument(ctx, document)
            if err != nil {
                errors[index] = err
                return
            }
            
            results[index] = analysis
        }(i, doc)
    }

    wg.Wait()
    
    // Filter out failed results
    validResults := make([]ReviewAnalysis, 0)
    validErrors := make([]error, 0)
    
    for i, result := range results {
        if errors[i] != nil {
            validErrors = append(validErrors, errors[i])
        } else {
            validResults = append(validResults, result)
        }
    }

    return validResults, validErrors
}

// processDocument processes a single document
func (dp *DocumentProcessor) processDocument(ctx context.Context, document string) (ReviewAnalysis, error) {
    state := domain.NewState()
    state.Set("user_input", fmt.Sprintf("Analyze this product review: %s", document))

    result, err := dp.agent.Run(ctx, state)
    if err != nil {
        return ReviewAnalysis{}, fmt.Errorf("agent processing failed: %w", err)
    }

    if structured, exists := result.Get("structured_output"); exists {
        analysis := ReviewAnalysis{}
        if data, ok := structured.(map[string]interface{}); ok {
            jsonData, _ := json.Marshal(data)
            if err := json.Unmarshal(jsonData, &analysis); err != nil {
                return ReviewAnalysis{}, fmt.Errorf("failed to parse result: %w", err)
            }
            return analysis, nil
        }
    }

    return ReviewAnalysis{}, fmt.Errorf("no structured output found")
}

func main() {
    // Sample product reviews
    reviews := []string{
        `⭐⭐⭐⭐⭐ Amazing wireless headphones! The noise cancellation is incredible and battery life lasts all day. Perfect for my daily commute. The build quality feels premium and the sound is crystal clear. Verified purchase.`,
        
        `⭐⭐ Disappointed with this laptop. Screen brightness is poor and the keyboard feels cheap. Customer service was unhelpful when I tried to return it. For this price, expected much better quality.`,
        
        `⭐⭐⭐⭐ Good smartphone overall. Camera quality is excellent, especially in low light. Battery life could be better - barely lasts a full day with heavy use. The interface is intuitive and responsive. Would recommend for photography enthusiasts.`,
        
        `⭐ DO NOT BUY! This coffee maker broke after 2 weeks. Water leaks everywhere and customer support ignores emails. Complete waste of money. Looking for alternatives now.`,
        
        `⭐⭐⭐⭐⭐ Best running shoes I've ever owned! Comfortable for long distances, great support, and they've held up well after 6 months of daily use. The cushioning is perfect for my heel strike pattern. Highly recommend for serious runners.`,
    }

    fmt.Println("🔄 Processing Product Reviews...")
    fmt.Printf("Documents to process: %d\n", len(reviews))
    fmt.Println("---")

    // Create document processor
    processor, err := NewDocumentProcessor(3) // Max 3 concurrent processes
    if err != nil {
        log.Fatalf("Failed to create processor: %v", err)
    }

    // Process all reviews
    startTime := time.Now()
    results, errors := processor.ProcessDocuments(context.Background(), reviews)
    processingTime := time.Since(startTime)

    // Display results
    fmt.Printf("✅ Processing complete in %v\n", processingTime)
    fmt.Printf("📊 Successfully processed: %d/%d documents\n", len(results), len(reviews))
    
    if len(errors) > 0 {
        fmt.Printf("❌ Failed: %d documents\n", len(errors))
        for i, err := range errors {
            fmt.Printf("   Error %d: %v\n", i+1, err)
        }
    }

    // Analyze results
    fmt.Println("\n📈 Analysis Summary:")
    
    positiveCount := 0
    negativeCount := 0
    neutralCount := 0
    totalTrust := 0.0
    verifiedCount := 0

    for i, result := range results {
        fmt.Printf("\n--- Review %d ---\n", i+1)
        fmt.Printf("Product: %s\n", result.ProductName)
        fmt.Printf("Rating: %d/5 | Sentiment: %s\n", result.Rating, result.Sentiment)
        fmt.Printf("Trust Score: %.2f | Verified: %t\n", result.TrustScore, result.Verified)
        fmt.Printf("Summary: %s\n", result.Summary)
        
        if len(result.Issues) > 0 {
            fmt.Printf("Issues: %v\n", result.Issues)
        }
        if len(result.Suggestions) > 0 {
            fmt.Printf("Suggestions: %v\n", result.Suggestions)
        }

        // Aggregate stats
        switch result.Sentiment {
        case "positive":
            positiveCount++
        case "negative":
            negativeCount++
        case "neutral":
            neutralCount++
        }
        
        totalTrust += result.TrustScore
        if result.Verified {
            verifiedCount++
        }
    }

    // Summary statistics
    if len(results) > 0 {
        fmt.Println("\n📊 Overall Statistics:")
        fmt.Printf("Positive: %d | Negative: %d | Neutral: %d\n", positiveCount, negativeCount, neutralCount)
        fmt.Printf("Average Trust Score: %.2f\n", totalTrust/float64(len(results)))
        fmt.Printf("Verified Reviews: %d/%d (%.1f%%)\n", verifiedCount, len(results), float64(verifiedCount)/float64(len(results))*100)
    }
}

func float64Ptr(v float64) *float64 {
    return &v
}
```

### Advanced Features

✅ **Concurrent Processing** - Handle multiple documents simultaneously  
✅ **Error Recovery** - Continue processing even if some documents fail  
✅ **Batch Validation** - Validate entire batches with comprehensive error reporting  
✅ **Progress Tracking** - Monitor processing status and performance  
✅ **Aggregation** - Combine results for analysis and reporting  

---

## Example 3: Error Recovery & Validation Pipeline
*Handle malformed data with automatic retry and fallback strategies*

### The Problem
Real-world data is messy. LLMs sometimes produce invalid JSON, miss required fields, or format data incorrectly.

### The Solution
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
    "github.com/lexlapax/go-llms/pkg/structured/processor"
)

// Contact represents extracted contact information
type Contact struct {
    Name     string `json:"name" validate:"required"`
    Email    string `json:"email" validate:"required,email"`
    Phone    string `json:"phone"`
    Company  string `json:"company"`
    Role     string `json:"role"`
    Priority string `json:"priority" validate:"oneof=low medium high"`
}

// ExtractionResult represents the result of an extraction attempt
type ExtractionResult struct {
    Contact   *Contact
    Success   bool
    Attempts  int
    Errors    []string
    Duration  time.Duration
}

// RobustExtractor handles extraction with error recovery
type RobustExtractor struct {
    primaryAgent   domain.BaseAgent
    fallbackAgent  domain.BaseAgent
    validator      *validation.Validator
    processor      *processor.StructuredProcessor
    schema         *schemaDomain.Schema
    maxRetries     int
}

func NewRobustExtractor() (*RobustExtractor, error) {
    // Primary agent (more capable but potentially more error-prone)
    primaryAgent, err := core.NewAgentFromString("primary-extractor", "openai/gpt-4o")
    if err != nil {
        return nil, fmt.Errorf("failed to create primary agent: %w", err)
    }
    primaryAgent.SetSystemPrompt(`You are a precise contact extraction specialist.
Extract contact information from text and return valid JSON.
Always include required fields: name, email, priority.
Use priority: "high" for executives, "medium" for managers, "low" for others.`)

    // Fallback agent (more reliable but potentially less capable)
    fallbackAgent, err := core.NewAgentFromString("fallback-extractor", "anthropic/claude-3-5-haiku")
    if err != nil {
        return nil, fmt.Errorf("failed to create fallback agent: %w", err)
    }
    fallbackAgent.SetSystemPrompt(`You are a reliable contact extractor.
Focus on accuracy over completeness. Extract only clearly identifiable information.
Always provide: name, email (if found), and priority (default to "medium").`)

    // Create schema
    schema := &schemaDomain.Schema{
        Type: "object",
        Properties: map[string]schemaDomain.Property{
            "name":     {Type: "string", Description: "Full name"},
            "email":    {Type: "string", Format: "email", Description: "Email address"},
            "phone":    {Type: "string", Description: "Phone number"},
            "company":  {Type: "string", Description: "Company name"},
            "role":     {Type: "string", Description: "Job title or role"},
            "priority": {Type: "string", Enum: []interface{}{"low", "medium", "high"}, Description: "Contact priority"},
        },
        Required: []string{"name", "email", "priority"},
    }

    primaryAgent.SetSchema(schema)
    fallbackAgent.SetSchema(schema)

    // Create validator and processor
    validator := validation.NewValidator()
    structProcessor := processor.NewStructuredProcessor(validator)

    return &RobustExtractor{
        primaryAgent:   primaryAgent,
        fallbackAgent:  fallbackAgent,
        validator:      validator,
        processor:      structProcessor,
        schema:         schema,
        maxRetries:     3,
    }, nil
}

// ExtractContact extracts contact information with error recovery
func (re *RobustExtractor) ExtractContact(ctx context.Context, text string) *ExtractionResult {
    startTime := time.Now()
    result := &ExtractionResult{
        Errors: make([]string, 0),
    }

    // Try primary agent first
    contact, err := re.tryExtraction(ctx, re.primaryAgent, text, "primary")
    if err == nil && contact != nil {
        result.Contact = contact
        result.Success = true
        result.Attempts = 1
        result.Duration = time.Since(startTime)
        return result
    }
    
    result.Errors = append(result.Errors, fmt.Sprintf("Primary agent failed: %v", err))
    result.Attempts = 1

    // Retry with primary agent
    for i := 0; i < re.maxRetries-1; i++ {
        result.Attempts++
        contact, err = re.tryExtraction(ctx, re.primaryAgent, text, fmt.Sprintf("primary-retry-%d", i+1))
        if err == nil && contact != nil {
            result.Contact = contact
            result.Success = true
            result.Duration = time.Since(startTime)
            return result
        }
        result.Errors = append(result.Errors, fmt.Sprintf("Primary retry %d failed: %v", i+1, err))
    }

    // Try fallback agent
    result.Attempts++
    contact, err = re.tryExtraction(ctx, re.fallbackAgent, text, "fallback")
    if err == nil && contact != nil {
        result.Contact = contact
        result.Success = true
        result.Duration = time.Since(startTime)
        return result
    }
    
    result.Errors = append(result.Errors, fmt.Sprintf("Fallback agent failed: %v", err))
    result.Duration = time.Since(startTime)
    return result
}

// tryExtraction attempts extraction with a specific agent
func (re *RobustExtractor) tryExtraction(ctx context.Context, agent domain.BaseAgent, text, method string) (*Contact, error) {
    state := domain.NewState()
    state.Set("user_input", fmt.Sprintf("Extract contact information from: %s", text))

    result, err := agent.Run(ctx, state)
    if err != nil {
        return nil, fmt.Errorf("agent run failed: %w", err)
    }

    if structured, exists := result.Get("structured_output"); exists {
        contact := &Contact{}
        if data, ok := structured.(map[string]interface{}); ok {
            jsonData, _ := json.Marshal(data)
            if err := json.Unmarshal(jsonData, contact); err != nil {
                return nil, fmt.Errorf("JSON unmarshal failed: %w", err)
            }

            // Validate the result
            if err := re.validateContact(contact); err != nil {
                return nil, fmt.Errorf("validation failed: %w", err)
            }

            return contact, nil
        }
    }

    return nil, fmt.Errorf("no structured output found")
}

// validateContact performs additional validation beyond schema
func (re *RobustExtractor) validateContact(contact *Contact) error {
    if contact.Name == "" {
        return fmt.Errorf("name cannot be empty")
    }
    
    if contact.Email == "" {
        return fmt.Errorf("email cannot be empty")
    }

    // Additional business logic validation
    if len(contact.Name) < 2 {
        return fmt.Errorf("name too short: %s", contact.Name)
    }

    // Email format validation (beyond schema)
    if !contains(contact.Email, "@") || !contains(contact.Email, ".") {
        return fmt.Errorf("invalid email format: %s", contact.Email)
    }

    return nil
}

func contains(str, substr string) bool {
    return len(str) > 0 && len(substr) > 0 && 
           len(str) >= len(substr) && 
           str != substr && 
           (str == substr || (len(str) > len(substr) && (
               str[:len(substr)] == substr || 
               str[len(str)-len(substr):] == substr || 
               (len(str) > len(substr)*2 && str[len(str)/2-len(substr)/2:len(str)/2+len(substr)/2+len(substr)%2] == substr))))
}

func main() {
    fmt.Println("🛡️ Robust Data Extraction with Error Recovery")
    fmt.Println("=============================================")

    // Create robust extractor
    extractor, err := NewRobustExtractor()
    if err != nil {
        log.Fatalf("Failed to create extractor: %v", err)
    }

    // Test cases with varying data quality
    testCases := []struct {
        name string
        text string
    }{
        {
            name: "Clean data",
            text: "Contact: John Smith, CEO at TechCorp Inc. Email: john.smith@techcorp.com Phone: +1-555-0123",
        },
        {
            name: "Messy data",
            text: "Hi there! My name's Sarah (sarah@startup.co) and I work for StartupCo as a product manager. Call me at 555.0456 if you need anything!",
        },
        {
            name: "Incomplete data",
            text: "Dr. Mike Johnson from University Research Lab. Very interested in our project.",
        },
        {
            name: "Complex format",
            text: `
            Business Card:
            ═══════════════════════════════════
            │ Amanda Rodriguez                  │
            │ Chief Technology Officer          │
            │ InnovateTech Solutions            │
            │                                   │
            │ 📧 a.rodriguez@innovatetech.com   │
            │ 📱 +1 (555) 789-0123              │
            │ 🌐 www.innovatetech.com           │
            ═══════════════════════════════════
            `,
        },
        {
            name: "Challenging extraction",
            text: "Email thread: From: Bob CEO <ceo@company.com> - needs urgent meeting about partnership",
        },
    }

    successCount := 0
    totalAttempts := 0
    totalDuration := time.Duration(0)

    for i, test := range testCases {
        fmt.Printf("\n--- Test %d: %s ---\n", i+1, test.name)
        fmt.Printf("Input: %s\n", truncateText(test.text, 100))

        result := extractor.ExtractContact(context.Background(), test.text)
        totalAttempts += result.Attempts
        totalDuration += result.Duration

        if result.Success {
            successCount++
            fmt.Printf("✅ Success (attempts: %d, duration: %v)\n", result.Attempts, result.Duration)
            fmt.Printf("   Name: %s\n", result.Contact.Name)
            fmt.Printf("   Email: %s\n", result.Contact.Email)
            fmt.Printf("   Company: %s\n", result.Contact.Company)
            fmt.Printf("   Role: %s\n", result.Contact.Role)
            fmt.Printf("   Priority: %s\n", result.Contact.Priority)
            if result.Contact.Phone != "" {
                fmt.Printf("   Phone: %s\n", result.Contact.Phone)
            }
        } else {
            fmt.Printf("❌ Failed (attempts: %d, duration: %v)\n", result.Attempts, result.Duration)
            fmt.Printf("   Errors:\n")
            for j, errMsg := range result.Errors {
                fmt.Printf("     %d. %s\n", j+1, errMsg)
            }
        }
    }

    // Summary statistics
    fmt.Println("\n📊 Extraction Summary:")
    fmt.Printf("Success Rate: %d/%d (%.1f%%)\n", successCount, len(testCases), float64(successCount)/float64(len(testCases))*100)
    fmt.Printf("Average Attempts: %.1f\n", float64(totalAttempts)/float64(len(testCases)))
    fmt.Printf("Average Duration: %v\n", totalDuration/time.Duration(len(testCases)))
}

func truncateText(text string, maxLen int) string {
    if len(text) <= maxLen {
        return text
    }
    return text[:maxLen] + "..."
}
```

### Resilience Features

✅ **Multiple Agents** - Primary and fallback agents for reliability  
✅ **Retry Logic** - Automatic retries with configurable limits  
✅ **Validation Layers** - Schema validation plus custom business logic  
✅ **Error Tracking** - Detailed error reporting and diagnostics  
✅ **Performance Monitoring** - Track attempts, duration, and success rates  

---

## Production Considerations

### Performance Optimization

```go
// Use connection pooling for high-throughput applications
type ProductionExtractor struct {
    agentPool    chan domain.BaseAgent
    poolSize     int
    maxRetries   int
    timeout      time.Duration
}

func NewProductionExtractor(poolSize int) *ProductionExtractor {
    pool := make(chan domain.BaseAgent, poolSize)
    
    // Pre-create agents for the pool
    for i := 0; i < poolSize; i++ {
        agent, _ := core.NewAgentFromString(fmt.Sprintf("extractor-%d", i), "openai/gpt-4o-mini")
        // Configure agent...
        pool <- agent
    }
    
    return &ProductionExtractor{
        agentPool:   pool,
        poolSize:    poolSize,
        maxRetries:  3,
        timeout:     30 * time.Second,
    }
}
```

### Error Monitoring

```go
// Track extraction metrics for monitoring
type ExtractionMetrics struct {
    TotalProcessed   int64
    SuccessCount     int64
    FailureCount     int64
    AverageLatency   time.Duration
    ErrorTypes       map[string]int64
}

func (em *ExtractionMetrics) RecordSuccess(duration time.Duration) {
    atomic.AddInt64(&em.TotalProcessed, 1)
    atomic.AddInt64(&em.SuccessCount, 1)
    // Update average latency...
}

func (em *ExtractionMetrics) RecordFailure(errorType string) {
    atomic.AddInt64(&em.TotalProcessed, 1)
    atomic.AddInt64(&em.FailureCount, 1)
    // Update error counts...
}
```

### Schema Evolution

```go
// Version your schemas for backward compatibility
type SchemaManager struct {
    schemas map[string]map[int]*schemaDomain.Schema // name -> version -> schema
}

func (sm *SchemaManager) GetSchema(name string, version int) *schemaDomain.Schema {
    if versions, exists := sm.schemas[name]; exists {
        if schema, exists := versions[version]; exists {
            return schema
        }
    }
    return nil
}

// Migration support for schema changes
func (sm *SchemaManager) MigrateData(data interface{}, fromVersion, toVersion int) (interface{}, error) {
    // Implement migration logic...
    return data, nil
}
```

## Best Practices

### 1. Schema Design
- **Start Simple** - Begin with basic required fields, add complexity gradually
- **Use Validation** - Leverage JSON Schema validation for data quality
- **Document Fields** - Provide clear descriptions for better extraction
- **Plan for Growth** - Design schemas that can evolve

### 2. Error Handling
- **Multiple Strategies** - Use primary/fallback agents for reliability
- **Retry Logic** - Implement intelligent retry with backoff
- **Validation Layers** - Schema + business logic validation
- **Error Categorization** - Track different error types for debugging

### 3. Performance
- **Batch Processing** - Process multiple documents efficiently
- **Concurrency Control** - Use semaphores to limit concurrent requests
- **Agent Pooling** - Pre-create agents for high-throughput scenarios
- **Caching** - Cache common extractions and schema validations

### 4. Monitoring
- **Success Rates** - Track extraction success by data type
- **Performance Metrics** - Monitor latency and throughput
- **Error Analytics** - Analyze failure patterns
- **Data Quality** - Monitor extracted data quality over time

## Troubleshooting Guide

### Common Issues

**Low Success Rates**
- Improve system prompts with specific instructions
- Add more validation rules to catch edge cases
- Use multiple provider strategies
- Implement better error recovery

**Slow Performance**
- Implement agent pooling
- Use faster models for simple extractions
- Add concurrent processing
- Cache frequently extracted data

**Data Quality Issues**
- Strengthen schema validation rules
- Add business logic validation
- Improve training examples in prompts
- Use multiple agents for cross-validation

**Schema Evolution**
- Version your schemas properly
- Implement migration strategies
- Test with real data before deployment
- Maintain backward compatibility

## Next Steps

🚀 **Ready to extract data like a pro?** Explore these advanced topics:

- **[Structured Data Guide](structured-data.md)** - Deep dive into schemas and validation
- **[Data Validation](data-validation.md)** - Advanced validation techniques
- **[Data Pipelines](data-pipelines.md)** - End-to-end processing workflows
- **[Agent Communication](agent-communication.md)** - Multi-agent coordination

### Related Examples

- **[Agent Structured Output](../../cmd/examples/agent-structured-output/)** - Complete structured output examples
- **[Structured Schema](../../cmd/examples/structured-schema/)** - Schema generation and usage
- **[Structured Coercion](../../cmd/examples/structured-coercion/)** - Type coercion and validation

---

**Need help?** Check our [troubleshooting guide](../advanced/troubleshooting.md) or join the discussion on [GitHub](https://github.com/lexlapax/go-llms/discussions).