# Data Validation: Validation and Error Recovery

> **[Project Root](/) / [Documentation](../..) / [User Guide](../../user-guide) / [Guides](../../user-guide/guides) / Data Validation**

Master comprehensive data validation and error recovery strategies for LLM outputs. Build robust systems that handle edge cases, validate complex data structures, and recover gracefully from errors.

## Why Data Validation Matters

- **Reliability** - Ensure data meets business requirements before use
- **Error Prevention** - Catch issues early in the processing pipeline
- **Data Quality** - Maintain high standards for downstream systems
- **User Trust** - Provide consistent, predictable behavior
- **Compliance** - Meet regulatory and security requirements

## Validation Architecture

![Data Validation Pipeline](../../images/validation-architecture.svg)

### Validation Layers
1. **Schema Validation** - Structure and type checking
2. **Business Rules** - Domain-specific constraints
3. **Semantic Validation** - Meaning and consistency
4. **Security Validation** - Sanitization and safety
5. **Cross-Field Validation** - Relationships between fields

### Error Recovery Strategies
| Strategy | Use Case | Recovery Method |
|----------|----------|-----------------|
| **Retry with Guidance** | Schema violations | Enhanced prompts |
| **Partial Acceptance** | Minor issues | Accept valid portions |
| **Fallback Values** | Missing data | Use defaults |
| **Alternative Parsing** | Format issues | Try different parsers |
| **Human Escalation** | Critical failures | Manual review |

## Prerequisites

- [Structured Data completed](structured-data.md) ✅
- [Schema validation understanding](structured-data.md#schema-validation) ✅
- Basic knowledge of validation patterns ✅

---

## Level 1: Basic Validation Patterns
*Implement fundamental validation and error handling*

### Comprehensive Validation System
```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "regexp"
    "strings"
    "time"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
    "github.com/lexlapax/go-llms/pkg/schema/validation"
)

// ValidationSystem provides multi-layer validation
type ValidationSystem struct {
    schemaValidator   *validation.Validator
    businessValidator *BusinessRuleValidator
    securityValidator *SecurityValidator
    semanticValidator *SemanticValidator
}

// BusinessRuleValidator checks domain-specific rules
type BusinessRuleValidator struct {
    rules []BusinessRule
}

type BusinessRule struct {
    Name        string
    Field       string
    Validator   func(value interface{}, data map[string]interface{}) error
    ErrorMessage string
}

// SecurityValidator checks for security issues
type SecurityValidator struct {
    sqlInjectionPattern    *regexp.Regexp
    scriptInjectionPattern *regexp.Regexp
    maxStringLength        int
    allowedDomains         []string
}

// SemanticValidator checks data meaning and consistency
type SemanticValidator struct {
    dateRangeValidator  func(start, end time.Time) error
    referenceValidator  func(field, reference string, data map[string]interface{}) error
}

// ValidationResult contains validation outcome
type ValidationResult struct {
    Valid       bool                   `json:"valid"`
    Errors      []ValidationError      `json:"errors,omitempty"`
    Warnings    []ValidationWarning    `json:"warnings,omitempty"`
    Suggestions []string               `json:"suggestions,omitempty"`
    CleanedData map[string]interface{} `json:"cleaned_data,omitempty"`
}

type ValidationError struct {
    Field    string      `json:"field"`
    Rule     string      `json:"rule"`
    Message  string      `json:"message"`
    Value    interface{} `json:"value,omitempty"`
    Severity string      `json:"severity"` // "error", "critical"
}

type ValidationWarning struct {
    Field   string `json:"field"`
    Message string `json:"message"`
    Type    string `json:"type"`
}

func NewValidationSystem() *ValidationSystem {
    return &ValidationSystem{
        schemaValidator:   validation.NewValidator(),
        businessValidator: NewBusinessRuleValidator(),
        securityValidator: NewSecurityValidator(),
        semanticValidator: NewSemanticValidator(),
    }
}

func NewBusinessRuleValidator() *BusinessRuleValidator {
    return &BusinessRuleValidator{
        rules: []BusinessRule{
            {
                Name:  "price_range",
                Field: "price",
                Validator: func(value interface{}, data map[string]interface{}) error {
                    if price, ok := value.(float64); ok {
                        if price < 0 || price > 1000000 {
                            return fmt.Errorf("price must be between 0 and 1,000,000")
                        }
                    }
                    return nil
                },
                ErrorMessage: "Price is outside acceptable range",
            },
            {
                Name:  "email_domain",
                Field: "email",
                Validator: func(value interface{}, data map[string]interface{}) error {
                    if email, ok := value.(string); ok {
                        // Check for disposable email domains
                        disposableDomains := []string{"tempmail.com", "throwaway.email", "guerrillamail.com"}
                        for _, domain := range disposableDomains {
                            if strings.Contains(email, domain) {
                                return fmt.Errorf("disposable email addresses not allowed")
                            }
                        }
                    }
                    return nil
                },
                ErrorMessage: "Invalid email domain",
            },
            {
                Name:  "phone_format",
                Field: "phone",
                Validator: func(value interface{}, data map[string]interface{}) error {
                    if phone, ok := value.(string); ok {
                        // Basic phone validation
                        phoneRegex := regexp.MustCompile(`^\+?[\d\s\-\(\)]+$`)
                        if !phoneRegex.MatchString(phone) {
                            return fmt.Errorf("invalid phone format")
                        }
                        // Check minimum length
                        digits := regexp.MustCompile(`\d`).FindAllString(phone, -1)
                        if len(digits) < 10 {
                            return fmt.Errorf("phone number must have at least 10 digits")
                        }
                    }
                    return nil
                },
                ErrorMessage: "Invalid phone number format",
            },
        },
    }
}

func NewSecurityValidator() *SecurityValidator {
    return &SecurityValidator{
        sqlInjectionPattern:    regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|script|javascript|onload|onerror)`),
        scriptInjectionPattern: regexp.MustCompile(`(?i)(<script|javascript:|onerror=|onload=|<iframe|<object|<embed)`),
        maxStringLength:        10000,
        allowedDomains:         []string{"example.com", "trusted.org"},
    }
}

func NewSemanticValidator() *SemanticValidator {
    return &SemanticValidator{
        dateRangeValidator: func(start, end time.Time) error {
            if end.Before(start) {
                return fmt.Errorf("end date must be after start date")
            }
            if end.Sub(start) > 365*24*time.Hour {
                return fmt.Errorf("date range cannot exceed one year")
            }
            return nil
        },
        referenceValidator: func(field, reference string, data map[string]interface{}) error {
            // Validate that referenced items exist
            if refValue, exists := data[reference]; !exists || refValue == nil {
                return fmt.Errorf("%s references non-existent %s", field, reference)
            }
            return nil
        },
    }
}

func (vs *ValidationSystem) ValidateData(schema *schemaDomain.Schema, data map[string]interface{}) *ValidationResult {
    result := &ValidationResult{
        Valid:       true,
        Errors:      []ValidationError{},
        Warnings:    []ValidationWarning{},
        Suggestions: []string{},
        CleanedData: make(map[string]interface{}),
    }

    // Layer 1: Schema validation
    if err := vs.schemaValidator.Validate(schema, data); err != nil {
        result.Valid = false
        result.Errors = append(result.Errors, ValidationError{
            Field:    "schema",
            Rule:     "json_schema",
            Message:  err.Error(),
            Severity: "error",
}
    }

    // Layer 2: Business rules validation
    for _, rule := range vs.businessValidator.rules {
        if value, exists := data[rule.Field]; exists {
            if err := rule.Validator(value, data); err != nil {
                result.Valid = false
                result.Errors = append(result.Errors, ValidationError{
                    Field:    rule.Field,
                    Rule:     rule.Name,
                    Message:  err.Error(),
                    Value:    value,
                    Severity: "error",
}
            }
        }
    }

    // Layer 3: Security validation
    vs.validateSecurity(data, result)

    // Layer 4: Semantic validation
    vs.validateSemantics(data, result)

    // Clean and sanitize data
    vs.cleanData(data, result)

    // Add suggestions if validation failed
    if !result.Valid {
        result.Suggestions = vs.generateSuggestions(result.Errors)
    }

    return result
}

func (vs *ValidationSystem) validateSecurity(data map[string]interface{}, result *ValidationResult) {
    for field, value := range data {
        if strValue, ok := value.(string); ok {
            // Check string length
            if len(strValue) > vs.securityValidator.maxStringLength {
                result.Warnings = append(result.Warnings, ValidationWarning{
                    Field:   field,
                    Message: fmt.Sprintf("String exceeds maximum length of %d characters", vs.securityValidator.maxStringLength),
                    Type:    "length",
}
            }

            // Check for SQL injection patterns
            if vs.securityValidator.sqlInjectionPattern.MatchString(strValue) {
                result.Valid = false
                result.Errors = append(result.Errors, ValidationError{
                    Field:    field,
                    Rule:     "sql_injection",
                    Message:  "Potential SQL injection detected",
                    Severity: "critical",
}
            }

            // Check for script injection
            if vs.securityValidator.scriptInjectionPattern.MatchString(strValue) {
                result.Valid = false
                result.Errors = append(result.Errors, ValidationError{
                    Field:    field,
                    Rule:     "script_injection",
                    Message:  "Potential script injection detected",
                    Severity: "critical",
}
            }

            // Validate URLs against allowed domains
            if strings.HasPrefix(strValue, "http://") || strings.HasPrefix(strValue, "https://") {
                allowed := false
                for _, domain := range vs.securityValidator.allowedDomains {
                    if strings.Contains(strValue, domain) {
                        allowed = true
                        break
                    }
                }
                if !allowed {
                    result.Warnings = append(result.Warnings, ValidationWarning{
                        Field:   field,
                        Message: "URL points to untrusted domain",
                        Type:    "security",
}
                }
            }
        }
    }
}

func (vs *ValidationSystem) validateSemantics(data map[string]interface{}, result *ValidationResult) {
    // Check date relationships
    if startDate, hasStart := data["start_date"]; hasStart {
        if endDate, hasEnd := data["end_date"]; hasEnd {
            // Convert to time.Time (simplified)
            if startStr, ok1 := startDate.(string); ok1 {
                if endStr, ok2 := endDate.(string); ok2 {
                    start, err1 := time.Parse("2006-01-02", startStr)
                    end, err2 := time.Parse("2006-01-02", endStr)
                    
                    if err1 == nil && err2 == nil {
                        if err := vs.semanticValidator.dateRangeValidator(start, end); err != nil {
                            result.Valid = false
                            result.Errors = append(result.Errors, ValidationError{
                                Field:    "date_range",
                                Rule:     "semantic_dates",
                                Message:  err.Error(),
                                Severity: "error",
}
                        }
                    }
                }
            }
        }
    }

    // Cross-field validation example
    if category, hasCategory := data["category"]; hasCategory {
        if category == "premium" {
            if price, hasPrice := data["price"]; hasPrice {
                if priceVal, ok := price.(float64); ok && priceVal < 100 {
                    result.Warnings = append(result.Warnings, ValidationWarning{
                        Field:   "price",
                        Message: "Premium products typically have prices above $100",
                        Type:    "semantic",
}
                }
            }
        }
    }
}

func (vs *ValidationSystem) cleanData(data map[string]interface{}, result *ValidationResult) {
    for field, value := range data {
        cleanedValue := value

        // Clean strings
        if strValue, ok := value.(string); ok {
            // Trim whitespace
            cleanedValue = strings.TrimSpace(strValue)
            
            // Remove dangerous characters
            cleanedValue = vs.sanitizeString(strValue)
        }

        result.CleanedData[field] = cleanedValue
    }
}

func (vs *ValidationSystem) sanitizeString(input string) string {
    // Basic sanitization
    sanitized := input
    
    // Remove null bytes
    sanitized = strings.ReplaceAll(sanitized, "\x00", "")
    
    // Escape HTML entities
    replacements := map[string]string{
        "<": "&lt;",
        ">": "&gt;",
        "&": "&amp;",
        "\"": "&quot;",
        "'": "&#39;",
    }
    
    for old, new := range replacements {
        sanitized = strings.ReplaceAll(sanitized, old, new)
    }
    
    return sanitized
}

func (vs *ValidationSystem) generateSuggestions(errors []ValidationError) []string {
    suggestions := []string{}
    
    for _, err := range errors {
        switch err.Rule {
        case "json_schema":
            suggestions = append(suggestions, "Review the schema requirements and ensure all required fields are present")
        case "sql_injection":
            suggestions = append(suggestions, fmt.Sprintf("Remove SQL keywords from field '%s'", err.Field))
        case "script_injection":
            suggestions = append(suggestions, fmt.Sprintf("Remove script tags or JavaScript from field '%s'", err.Field))
        case "price_range":
            suggestions = append(suggestions, "Ensure price is within acceptable range (0-1,000,000)")
        case "email_domain":
            suggestions = append(suggestions, "Use a professional email address from a non-disposable domain")
        case "semantic_dates":
            suggestions = append(suggestions, "Verify date ranges are logical and within acceptable limits")
        }
    }
    
    return suggestions
}

// Example usage with LLM integration
func main() {
    fmt.Println("🛡️ Data Validation - Comprehensive Validation")
    fmt.Println("============================================")

    // Create validation system
    validationSystem := NewValidationSystem()

    // Define schema for customer data
    customerSchema := &schemaDomain.Schema{
        Type: "object",
        Properties: map[string]schemaDomain.Property{
            "name": {
                Type:      "string",
                MinLength: intPtr(2),
                MaxLength: intPtr(100),
            },
            "email": {
                Type:   "string",
                Format: "email",
            },
            "phone": {
                Type: "string",
            },
            "age": {
                Type:    "integer",
                Minimum: float64Ptr(18),
                Maximum: float64Ptr(120),
            },
            "category": {
                Type: "string",
                Enum: []interface{}{"standard", "premium", "vip"},
            },
            "price": {
                Type:    "number",
                Minimum: float64Ptr(0),
            },
            "start_date": {
                Type:   "string",
                Format: "date",
            },
            "end_date": {
                Type:   "string",
                Format: "date",
            },
            "notes": {
                Type: "string",
            },
        },
        Required: []string{"name", "email"},
    }

    // Test cases with various validation issues
    testCases := []struct {
        name string
        data map[string]interface{}
    }{
        {
            name: "Valid data",
            data: map[string]interface{}{
                "name":       "John Doe",
                "email":      "john.doe@example.com",
                "phone":      "+1-555-123-4567",
                "age":        30,
                "category":   "premium",
                "price":      299.99,
                "start_date": "2024-01-01",
                "end_date":   "2024-12-31",
                "notes":      "Good customer",
            },
        },
        {
            name: "Security issues",
            data: map[string]interface{}{
                "name":  "Robert'; DROP TABLE users;--",
                "email": "hacker@tempmail.com",
                "notes": "<script>alert('XSS')</script>",
            },
        },
        {
            name: "Business rule violations",
            data: map[string]interface{}{
                "name":     "Jane Smith",
                "email":    "jane@example.com",
                "phone":    "123", // Too short
                "price":    2000000, // Too high
                "category": "premium",
            },
        },
        {
            name: "Semantic issues",
            data: map[string]interface{}{
                "name":       "Alice Johnson",
                "email":      "alice@example.com",
                "category":   "premium",
                "price":      50.00, // Too low for premium
                "start_date": "2024-12-31",
                "end_date":   "2024-01-01", // End before start
            },
        },
    }

    // Validate each test case
    for _, testCase := range testCases {
        fmt.Printf("\n📋 Test Case: %s\n", testCase.name)
        fmt.Println(strings.Repeat("-", 50))

        result := validationSystem.ValidateData(customerSchema, testCase.data)

        fmt.Printf("Valid: %v\n", result.Valid)

        if len(result.Errors) > 0 {
            fmt.Printf("\n❌ Errors (%d):\n", len(result.Errors))
            for _, err := range result.Errors {
                fmt.Printf("  - [%s] %s: %s (Field: %s)\n", 
                    err.Severity, err.Rule, err.Message, err.Field)
            }
        }

        if len(result.Warnings) > 0 {
            fmt.Printf("\n⚠️  Warnings (%d):\n", len(result.Warnings))
            for _, warn := range result.Warnings {
                fmt.Printf("  - [%s] %s: %s\n", warn.Type, warn.Field, warn.Message)
            }
        }

        if len(result.Suggestions) > 0 {
            fmt.Printf("\n💡 Suggestions:\n")
            for _, suggestion := range result.Suggestions {
                fmt.Printf("  - %s\n", suggestion)
            }
        }

        if result.Valid && len(result.CleanedData) > 0 {
            fmt.Printf("\n✅ Cleaned Data:\n")
            cleanedJSON, _ := json.MarshalIndent(result.CleanedData, "  ", "  ")
            fmt.Println(string(cleanedJSON))
        }
    }
}

// Helper functions
func intPtr(i int) *int          { return &i }
func float64Ptr(f float64) *float64 { return &f }
```

---

## Level 2: Advanced Error Recovery
*Implement sophisticated recovery strategies*

### Intelligent Error Recovery System
```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "strings"
    "time"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/errors/recovery"
    schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
    "github.com/lexlapax/go-llms/pkg/schema/validation"
    "github.com/lexlapax/go-llms/pkg/structured/processor"
)

// ErrorRecoverySystem handles validation failures with intelligence
type ErrorRecoverySystem struct {
    agent            domain.BaseAgent
    validator        *validation.Validator
    processor        *processor.StructuredProcessor
    recoveryStrategy recovery.Strategy
    maxAttempts      int
}

// RecoveryContext provides information for recovery attempts
type RecoveryContext struct {
    Schema           *schemaDomain.Schema
    OriginalData     map[string]interface{}
    ValidationErrors []ValidationError
    AttemptNumber    int
    History          []RecoveryAttempt
}

type RecoveryAttempt struct {
    Timestamp time.Time
    Strategy  string
    Success   bool
    Result    map[string]interface{}
    Error     error
}

// RecoveryResult contains the outcome of recovery attempts
type RecoveryResult struct {
    Success        bool                   `json:"success"`
    RecoveredData  map[string]interface{} `json:"recovered_data,omitempty"`
    PartialData    map[string]interface{} `json:"partial_data,omitempty"`
    UnrecoverableErrors []string          `json:"unrecoverable_errors,omitempty"`
    RecoverySteps  []string              `json:"recovery_steps"`
    Attempts       int                   `json:"attempts"`
}

func NewErrorRecoverySystem(agent domain.BaseAgent) *ErrorRecoverySystem {
    validator := validation.NewValidator()
    structProcessor := processor.NewStructuredProcessor(validator)

    // Create composite recovery strategy
    recoveryStrat := recovery.NewCompositeStrategy(
        recovery.NewExponentialBackoffStrategy(3, time.Second),
        recovery.NewFallbackStrategy(defaultValueProvider),
    )

    return &ErrorRecoverySystem{
        agent:            agent,
        validator:        validator,
        processor:        structProcessor,
        recoveryStrategy: recoveryStrat,
        maxAttempts:      3,
    }
}

func (ers *ErrorRecoverySystem) RecoverFromErrors(ctx context.Context, context RecoveryContext) (*RecoveryResult, error) {
    result := &RecoveryResult{
        RecoverySteps:       []string{},
        UnrecoverableErrors: []string{},
    }

    // Try different recovery strategies
    strategies := []struct {
        name     string
        recoverer func(context.Context, RecoveryContext) (map[string]interface{}, error)
    }{
        {"guided_retry", ers.guidedRetryRecovery},
        {"partial_acceptance", ers.partialAcceptanceRecovery},
        {"type_coercion", ers.typeCoercionRecovery},
        {"llm_correction", ers.llmCorrectionRecovery},
        {"fallback_defaults", ers.fallbackDefaultsRecovery},
    }

    currentData := context.OriginalData
    
    for attempt := 1; attempt <= ers.maxAttempts; attempt++ {
        context.AttemptNumber = attempt
        
        for _, strategy := range strategies {
            result.RecoverySteps = append(result.RecoverySteps, 
                fmt.Sprintf("Attempt %d: Trying %s strategy", attempt, strategy.name))

            recoveredData, err := strategy.recoverer(ctx, context)
            if err != nil {
                continue
            }

            // Validate recovered data
            validationErr := ers.validator.Validate(context.Schema, recoveredData)
            if validationErr == nil {
                result.Success = true
                result.RecoveredData = recoveredData
                result.Attempts = attempt
                
                context.History = append(context.History, RecoveryAttempt{
                    Timestamp: time.Now(),
                    Strategy:  strategy.name,
                    Success:   true,
                    Result:    recoveredData,
}
                
                return result, nil
            }

            // Update context with new data for next attempt
            currentData = recoveredData
        }
    }

    // If all recovery attempts failed, return partial data
    result.PartialData = ers.extractValidFields(context.Schema, currentData)
    result.Attempts = ers.maxAttempts
    
    for _, err := range context.ValidationErrors {
        result.UnrecoverableErrors = append(result.UnrecoverableErrors, 
            fmt.Sprintf("%s: %s", err.Field, err.Message))
    }

    return result, fmt.Errorf("recovery failed after %d attempts", ers.maxAttempts)
}

func (ers *ErrorRecoverySystem) guidedRetryRecovery(ctx context.Context, recContext RecoveryContext) (map[string]interface{}, error) {
    // Build enhanced prompt with specific error guidance
    prompt := ers.buildGuidedPrompt(recContext)

    state := domain.NewState()
    state.Set("user_input", prompt)
    state.Set("recovery_attempt", recContext.AttemptNumber)

    result, err := ers.agent.Run(ctx, state)
    if err != nil {
        return nil, err
    }

    if response, exists := result.Get("response"); exists {
        // Try to extract JSON from response
        var recoveredData map[string]interface{}
        if err := json.Unmarshal([]byte(response.(string)), &recoveredData); err != nil {
            return nil, fmt.Errorf("failed to parse guided retry response: %w", err)
        }
        
        return recoveredData, nil
    }

    return nil, fmt.Errorf("no response from guided retry")
}

func (ers *ErrorRecoverySystem) partialAcceptanceRecovery(ctx context.Context, recContext RecoveryContext) (map[string]interface{}, error) {
    // Accept valid fields and fix only problematic ones
    validData := make(map[string]interface{})
    problematicFields := make(map[string]ValidationError)

    // Identify problematic fields
    for _, err := range recContext.ValidationErrors {
        problematicFields[err.Field] = err
    }

    // Copy valid fields
    for field, value := range recContext.OriginalData {
        if _, isProblematic := problematicFields[field]; !isProblematic {
            validData[field] = value
        }
    }

    // Try to fix problematic fields individually
    for field, err := range problematicFields {
        fixedValue := ers.fixField(field, recContext.OriginalData[field], err, recContext.Schema)
        if fixedValue != nil {
            validData[field] = fixedValue
        }
    }

    return validData, nil
}

func (ers *ErrorRecoverySystem) typeCoercionRecovery(ctx context.Context, recContext RecoveryContext) (map[string]interface{}, error) {
    // Apply aggressive type coercion
    coercedData := make(map[string]interface{})

    for field, value := range recContext.OriginalData {
        if prop, exists := recContext.Schema.Properties[field]; exists {
            coercedValue := ers.aggressiveCoerce(value, prop)
            coercedData[field] = coercedValue
        } else {
            coercedData[field] = value
        }
    }

    return coercedData, nil
}

func (ers *ErrorRecoverySystem) llmCorrectionRecovery(ctx context.Context, recContext RecoveryContext) (map[string]interface{}, error) {
    // Use LLM to intelligently correct the data
    correctionPrompt := fmt.Sprintf(`Fix the following JSON data to match the schema requirements:

Original Data:
%s

Validation Errors:
%s

Schema Requirements:
%s

Return only the corrected JSON with all errors fixed.`,
        mustMarshalJSON(recContext.OriginalData),
        ers.formatValidationErrors(recContext.ValidationErrors),
        ers.summarizeSchema(recContext.Schema))

    state := domain.NewState()
    state.Set("user_input", correctionPrompt)

    result, err := ers.agent.Run(ctx, state)
    if err != nil {
        return nil, err
    }

    if response, exists := result.Get("response"); exists {
        var correctedData map[string]interface{}
        if err := json.Unmarshal([]byte(response.(string)), &correctedData); err != nil {
            return nil, fmt.Errorf("failed to parse LLM correction: %w", err)
        }
        
        return correctedData, nil
    }

    return nil, fmt.Errorf("no response from LLM correction")
}

func (ers *ErrorRecoverySystem) fallbackDefaultsRecovery(ctx context.Context, recContext RecoveryContext) (map[string]interface{}, error) {
    // Use schema defaults and smart defaults
    defaultedData := make(map[string]interface{})

    // Copy existing valid data
    for field, value := range recContext.OriginalData {
        defaultedData[field] = value
    }

    // Fill in missing required fields with defaults
    for _, requiredField := range recContext.Schema.Required {
        if _, exists := defaultedData[requiredField]; !exists {
            if prop, hasProp := recContext.Schema.Properties[requiredField]; hasProp {
                defaultValue := ers.getSmartDefault(requiredField, prop)
                if defaultValue != nil {
                    defaultedData[requiredField] = defaultValue
                }
            }
        }
    }

    return defaultedData, nil
}

func (ers *ErrorRecoverySystem) buildGuidedPrompt(context RecoveryContext) string {
    errorSummary := strings.Builder{}
    for _, err := range context.ValidationErrors {
        errorSummary.WriteString(fmt.Sprintf("- %s: %s\n", err.Field, err.Message))
    }

    return fmt.Sprintf(`The previous data extraction had validation errors. Please provide corrected data.

Validation Errors:
%s

Requirements:
1. Fix all validation errors listed above
2. Maintain all valid data from the original extraction
3. Follow the exact schema requirements
4. Return only valid JSON

Original Data for Reference:
%s

Please provide the corrected JSON:`, 
        errorSummary.String(), 
        mustMarshalJSON(context.OriginalData))
}

func (ers *ErrorRecoverySystem) fixField(field string, value interface{}, err ValidationError, schema *schemaDomain.Schema) interface{} {
    prop, exists := schema.Properties[field]
    if !exists {
        return nil
    }

    switch err.Rule {
    case "type":
        // Try type conversion
        return ers.convertType(value, prop.Type)
    
    case "format":
        // Try format fixing
        return ers.fixFormat(value, prop.Format)
    
    case "minimum", "maximum":
        // Clamp to valid range
        return ers.clampValue(value, prop)
    
    case "pattern":
        // Try to fix pattern
        return ers.fixPattern(value, prop.Pattern)
    
    default:
        return nil
    }
}

func (ers *ErrorRecoverySystem) aggressiveCoerce(value interface{}, prop schemaDomain.Property) interface{} {
    switch prop.Type {
    case "string":
        return fmt.Sprintf("%v", value)
    
    case "number", "integer":
        // Try various number conversions
        switch v := value.(type) {
        case string:
            // Remove common formatting
            cleaned := strings.ReplaceAll(v, ",", "")
            cleaned = strings.ReplaceAll(cleaned, "$", "")
            cleaned = strings.TrimSpace(cleaned)
            
            var num float64
            fmt.Sscanf(cleaned, "%f", &num)
            
            if prop.Type == "integer" {
                return int(num)
            }
            return num
        case bool:
            if v {
                return 1
            }
            return 0
        default:
            return value
        }
    
    case "boolean":
        switch v := value.(type) {
        case string:
            lower := strings.ToLower(strings.TrimSpace(v))
            return lower == "true" || lower == "yes" || lower == "1" || lower == "on"
        case int, float64:
            return v != 0
        default:
            return false
        }
    
    case "array":
        // Convert single values to arrays
        switch v := value.(type) {
        case []interface{}:
            return v
        case string:
            // Try to split string into array
            if strings.Contains(v, ",") {
                parts := strings.Split(v, ",")
                result := make([]interface{}, len(parts))
                for i, part := range parts {
                    result[i] = strings.TrimSpace(part)
                }
                return result
            }
            return []interface{}{v}
        default:
            return []interface{}{v}
        }
    
    default:
        return value
    }
}

func (ers *ErrorRecoverySystem) getSmartDefault(field string, prop schemaDomain.Property) interface{} {
    // First check if schema has default
    if prop.Default != nil {
        return prop.Default
    }

    // Smart defaults based on field name and type
    fieldLower := strings.ToLower(field)
    
    switch prop.Type {
    case "string":
        if strings.Contains(fieldLower, "email") {
            return "default@example.com"
        }
        if strings.Contains(fieldLower, "name") {
            return "Unknown"
        }
        if strings.Contains(fieldLower, "id") {
            return fmt.Sprintf("auto_%d", time.Now().Unix())
        }
        if prop.Format == "date" {
            return time.Now().Format("2006-01-02")
        }
        if prop.Format == "date-time" {
            return time.Now().Format(time.RFC3339)
        }
        return ""
    
    case "number", "integer":
        if strings.Contains(fieldLower, "price") || strings.Contains(fieldLower, "cost") {
            return 0.0
        }
        if strings.Contains(fieldLower, "quantity") || strings.Contains(fieldLower, "count") {
            return 1
        }
        if prop.Minimum != nil {
            return *prop.Minimum
        }
        return 0
    
    case "boolean":
        if strings.Contains(fieldLower, "active") || strings.Contains(fieldLower, "enabled") {
            return true
        }
        return false
    
    case "array":
        return []interface{}{}
    
    case "object":
        return map[string]interface{}{}
    
    default:
        return nil
    }
}

func (ers *ErrorRecoverySystem) extractValidFields(schema *schemaDomain.Schema, data map[string]interface{}) map[string]interface{} {
    validData := make(map[string]interface{})

    for field, value := range data {
        if prop, exists := schema.Properties[field]; exists {
            // Try to validate just this field
            singleFieldData := map[string]interface{}{field: value}
            singleFieldSchema := &schemaDomain.Schema{
                Type: "object",
                Properties: map[string]schemaDomain.Property{
                    field: prop,
                },
            }
            
            if err := ers.validator.Validate(singleFieldSchema, singleFieldData); err == nil {
                validData[field] = value
            }
        }
    }

    return validData
}

// Helper functions
func (ers *ErrorRecoverySystem) formatValidationErrors(errors []ValidationError) string {
    var sb strings.Builder
    for _, err := range errors {
        sb.WriteString(fmt.Sprintf("- Field '%s': %s (Rule: %s)\n", err.Field, err.Message, err.Rule))
    }
    return sb.String()
}

func (ers *ErrorRecoverySystem) summarizeSchema(schema *schemaDomain.Schema) string {
    summary, _ := json.MarshalIndent(schema, "", "  ")
    return string(summary)
}

func (ers *ErrorRecoverySystem) convertType(value interface{}, targetType string) interface{} {
    // Implementation of type conversion logic
    return value
}

func (ers *ErrorRecoverySystem) fixFormat(value interface{}, format string) interface{} {
    // Implementation of format fixing logic
    return value
}

func (ers *ErrorRecoverySystem) clampValue(value interface{}, prop schemaDomain.Property) interface{} {
    // Implementation of value clamping logic
    return value
}

func (ers *ErrorRecoverySystem) fixPattern(value interface{}, pattern string) interface{} {
    // Implementation of pattern fixing logic
    return value
}

func mustMarshalJSON(v interface{}) string {
    data, _ := json.MarshalIndent(v, "", "  ")
    return string(data)
}

func defaultValueProvider() error {
    return nil
}

// Example usage
func main() {
    fmt.Println("🔧 Data Validation - Error Recovery")
    fmt.Println("===================================")

    // Create agent
    agent, err := core.NewAgentFromString("recovery-agent", "openai/gpt-4o-mini")
    if err != nil {
        log.Fatalf("Failed to create agent: %v", err)
    }

    // Create recovery system
    recoverySystem := NewErrorRecoverySystem(agent)

    // Define schema
    orderSchema := &schemaDomain.Schema{
        Type: "object",
        Properties: map[string]schemaDomain.Property{
            "order_id": {
                Type:    "string",
                Pattern: `^ORD-\d{6}$`,
            },
            "customer_email": {
                Type:   "string",
                Format: "email",
            },
            "items": {
                Type: "array",
                Items: &schemaDomain.Schema{
                    Type: "object",
                    Properties: map[string]schemaDomain.Property{
                        "product": {Type: "string"},
                        "quantity": {Type: "integer", Minimum: float64Ptr(1)},
                        "price": {Type: "number", Minimum: float64Ptr(0)},
                    },
                    Required: []string{"product", "quantity", "price"},
                },
                MinItems: intPtr(1),
            },
            "total": {
                Type:    "number",
                Minimum: float64Ptr(0),
            },
            "status": {
                Type: "string",
                Enum: []interface{}{"pending", "processing", "shipped", "delivered"},
            },
        },
        Required: []string{"order_id", "customer_email", "items", "total", "status"},
    }

    // Test cases with various errors
    errorCases := []struct {
        name string
        data map[string]interface{}
        errors []ValidationError
    }{
        {
            name: "Invalid types and formats",
            data: map[string]interface{}{
                "order_id":       "12345", // Missing ORD- prefix
                "customer_email": "not-an-email",
                "items": []interface{}{
                    map[string]interface{}{
                        "product":  "Widget",
                        "quantity": "two", // String instead of integer
                        "price":    "$19.99", // String with currency symbol
                    },
                },
                "total":  "39.98 USD", // String instead of number
                "status": "in transit", // Not in enum
            },
            errors: []ValidationError{
                {Field: "order_id", Rule: "pattern", Message: "Does not match pattern ^ORD-\\d{6}$"},
                {Field: "customer_email", Rule: "format", Message: "Invalid email format"},
                {Field: "items[0].quantity", Rule: "type", Message: "Expected integer, got string"},
                {Field: "items[0].price", Rule: "type", Message: "Expected number, got string"},
                {Field: "total", Rule: "type", Message: "Expected number, got string"},
                {Field: "status", Rule: "enum", Message: "Value not in allowed list"},
            },
        },
    }

    ctx := context.Background()

    for _, testCase := range errorCases {
        fmt.Printf("\n📋 Recovery Test: %s\n", testCase.name)
        fmt.Println(strings.Repeat("-", 50))

        recoveryContext := RecoveryContext{
            Schema:           orderSchema,
            OriginalData:     testCase.data,
            ValidationErrors: testCase.errors,
            History:          []RecoveryAttempt{},
        }

        fmt.Println("Original Data:")
        fmt.Println(mustMarshalJSON(testCase.data))

        fmt.Printf("\nValidation Errors: %d\n", len(testCase.errors))
        for _, err := range testCase.errors {
            fmt.Printf("  - %s: %s\n", err.Field, err.Message)
        }

        // Attempt recovery
        result, err := recoverySystem.RecoverFromErrors(ctx, recoveryContext)
        
        fmt.Printf("\n🔧 Recovery Result:\n")
        fmt.Printf("Success: %v\n", result.Success)
        fmt.Printf("Attempts: %d\n", result.Attempts)

        if result.Success {
            fmt.Println("\n✅ Recovered Data:")
            fmt.Println(mustMarshalJSON(result.RecoveredData))
        } else {
            if len(result.PartialData) > 0 {
                fmt.Println("\n⚠️  Partial Data:")
                fmt.Println(mustMarshalJSON(result.PartialData))
            }
            
            if len(result.UnrecoverableErrors) > 0 {
                fmt.Println("\n❌ Unrecoverable Errors:")
                for _, err := range result.UnrecoverableErrors {
                    fmt.Printf("  - %s\n", err)
                }
            }
        }

        if len(result.RecoverySteps) > 0 {
            fmt.Println("\n📝 Recovery Steps:")
            for _, step := range result.RecoverySteps {
                fmt.Printf("  - %s\n", step)
            }
        }
    }
}
```

---

## Level 3: Production Validation Framework
*Enterprise-grade validation with monitoring*

### Comprehensive Validation Platform
```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/google/uuid"
    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
    "github.com/lexlapax/go-llms/pkg/schema/validation"
)

// ValidationPlatform provides enterprise validation capabilities
type ValidationPlatform struct {
    validators       map[string]Validator
    rules            *RuleEngine
    monitor          *ValidationMonitor
    auditLog         *AuditLogger
    cache            *ValidationCache
    circuitBreaker   *CircuitBreaker
}

// Validator interface for different validation types
type Validator interface {
    Name() string
    Validate(ctx context.Context, data interface{}) (*ValidationResult, error)
    Priority() int
}

// RuleEngine manages business rules
type RuleEngine struct {
    rules    map[string][]Rule
    mu       sync.RWMutex
    executor *RuleExecutor
}

type Rule struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Category    string                 `json:"category"`
    Condition   string                 `json:"condition"`
    Action      string                 `json:"action"`
    Priority    int                    `json:"priority"`
    Enabled     bool                   `json:"enabled"`
    Metadata    map[string]interface{} `json:"metadata"`
}

// ValidationMonitor tracks validation metrics
type ValidationMonitor struct {
    mu              sync.RWMutex
    metrics         map[string]*ValidationMetrics
    alerts          chan ValidationAlert
    thresholds      map[string]float64
}

type ValidationMetrics struct {
    TotalValidations   int64
    SuccessCount       int64
    FailureCount       int64
    AvgValidationTime  time.Duration
    ErrorsByType       map[string]int64
    LastValidation     time.Time
}

type ValidationAlert struct {
    ID          string    `json:"id"`
    Type        string    `json:"type"`
    Severity    string    `json:"severity"`
    Message     string    `json:"message"`
    Metric      string    `json:"metric"`
    Value       float64   `json:"value"`
    Threshold   float64   `json:"threshold"`
    Timestamp   time.Time `json:"timestamp"`
}

// AuditLogger provides compliance logging
type AuditLogger struct {
    entries  []AuditEntry
    mu       sync.Mutex
    storage  AuditStorage
}

type AuditEntry struct {
    ID            string                 `json:"id"`
    Timestamp     time.Time              `json:"timestamp"`
    Operation     string                 `json:"operation"`
    UserID        string                 `json:"user_id"`
    DataType      string                 `json:"data_type"`
    Result        string                 `json:"result"`
    ErrorDetails  []string               `json:"error_details,omitempty"`
    DataHash      string                 `json:"data_hash"`
    Metadata      map[string]interface{} `json:"metadata"`
}

type AuditStorage interface {
    Store(entry AuditEntry) error
    Query(filter AuditFilter) ([]AuditEntry, error)
}

// ValidationCache for performance
type ValidationCache struct {
    cache    map[string]*CachedValidation
    mu       sync.RWMutex
    ttl      time.Duration
    maxSize  int
}

type CachedValidation struct {
    Result    *ValidationResult
    Timestamp time.Time
    Hash      string
}

// CircuitBreaker for fault tolerance
type CircuitBreaker struct {
    failureThreshold int
    resetTimeout     time.Duration
    failures         int
    lastFailure      time.Time
    state            string // "closed", "open", "half-open"
    mu               sync.Mutex
}

// ValidationRequest for batch processing
type ValidationRequest struct {
    ID           string                 `json:"id"`
    Type         string                 `json:"type"`
    Data         interface{}            `json:"data"`
    Schema       *schemaDomain.Schema   `json:"schema,omitempty"`
    Rules        []string               `json:"rules,omitempty"`
    Options      ValidationOptions      `json:"options"`
    Context      map[string]interface{} `json:"context"`
}

type ValidationOptions struct {
    StrictMode       bool          `json:"strict_mode"`
    PartialAccept    bool          `json:"partial_accept"`
    EnableRecovery   bool          `json:"enable_recovery"`
    MaxRecoveryTime  time.Duration `json:"max_recovery_time"`
    CacheResults     bool          `json:"cache_results"`
    AuditRequired    bool          `json:"audit_required"`
}

// BatchValidationResult for multiple validations
type BatchValidationResult struct {
    ID              string                    `json:"id"`
    TotalRequests   int                      `json:"total_requests"`
    SuccessCount    int                      `json:"success_count"`
    FailureCount    int                      `json:"failure_count"`
    Results         map[string]*ValidationResult `json:"results"`
    ProcessingTime  time.Duration            `json:"processing_time"`
    Timestamp       time.Time                `json:"timestamp"`
}

func NewValidationPlatform() *ValidationPlatform {
    return &ValidationPlatform{
        validators:     make(map[string]Validator),
        rules:          NewRuleEngine(),
        monitor:        NewValidationMonitor(),
        auditLog:       NewAuditLogger(),
        cache:          NewValidationCache(15*time.Minute, 1000),
        circuitBreaker: NewCircuitBreaker(5, 30*time.Second),
    }
}

func NewRuleEngine() *RuleEngine {
    return &RuleEngine{
        rules:    make(map[string][]Rule),
        executor: NewRuleExecutor(),
    }
}

func NewValidationMonitor() *ValidationMonitor {
    return &ValidationMonitor{
        metrics:    make(map[string]*ValidationMetrics),
        alerts:     make(chan ValidationAlert, 100),
        thresholds: map[string]float64{
            "failure_rate":     0.1, // 10% failure rate
            "avg_latency_ms":   1000, // 1 second
            "error_spike":      5.0, // 5x normal error rate
        },
    }
}

func NewAuditLogger() *AuditLogger {
    return &AuditLogger{
        entries: []AuditEntry{},
        storage: NewMemoryAuditStorage(), // In production, use persistent storage
    }
}

func NewValidationCache(ttl time.Duration, maxSize int) *ValidationCache {
    return &ValidationCache{
        cache:   make(map[string]*CachedValidation),
        ttl:     ttl,
        maxSize: maxSize,
    }
}

func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        failureThreshold: threshold,
        resetTimeout:     timeout,
        state:           "closed",
    }
}

func (vp *ValidationPlatform) RegisterValidator(validator Validator) {
    vp.validators[validator.Name()] = validator
    log.Printf("✅ Registered validator: %s", validator.Name())
}

func (vp *ValidationPlatform) ValidateData(ctx context.Context, request ValidationRequest) (*ValidationResult, error) {
    startTime := time.Now()

    // Check circuit breaker
    if err := vp.circuitBreaker.Call(func() error {
        return nil // Check if we should proceed
    }); err != nil {
        return nil, fmt.Errorf("circuit breaker open: %w", err)
    }

    // Check cache if enabled
    if request.Options.CacheResults {
        if cached := vp.cache.Get(request.ID); cached != nil {
            vp.monitor.RecordValidation("cache_hit", true, time.Since(startTime))
            return cached, nil
        }
    }

    // Initialize result
    result := &ValidationResult{
        Valid:       true,
        Errors:      []ValidationError{},
        Warnings:    []ValidationWarning{},
        Suggestions: []string{},
    }

    // Run validators in priority order
    for _, validator := range vp.getSortedValidators() {
        validatorResult, err := validator.Validate(ctx, request.Data)
        if err != nil {
            vp.circuitBreaker.RecordFailure()
            return nil, err
        }

        // Merge results
        result = vp.mergeResults(result, validatorResult)
        
        // Stop on critical errors if in strict mode
        if request.Options.StrictMode && !result.Valid {
            break
        }
    }

    // Apply business rules
    if len(request.Rules) > 0 {
        ruleResult := vp.rules.Execute(ctx, request.Data, request.Rules)
        result = vp.mergeResults(result, ruleResult)
    }

    // Attempt recovery if enabled and validation failed
    if !result.Valid && request.Options.EnableRecovery {
        recoveryCtx, cancel := context.WithTimeout(ctx, request.Options.MaxRecoveryTime)
        defer cancel()
        
        if recovered := vp.attemptRecovery(recoveryCtx, request, result); recovered != nil {
            result = recovered
        }
    }

    // Record metrics
    vp.monitor.RecordValidation(request.Type, result.Valid, time.Since(startTime))

    // Audit if required
    if request.Options.AuditRequired {
        vp.auditLog.LogValidation(request, result)
    }

    // Cache result if enabled
    if request.Options.CacheResults && result.Valid {
        vp.cache.Set(request.ID, result)
    }

    // Check for alerts
    vp.monitor.CheckAlerts()

    return result, nil
}

func (vp *ValidationPlatform) ValidateBatch(ctx context.Context, requests []ValidationRequest) (*BatchValidationResult, error) {
    startTime := time.Now()
    
    batchResult := &BatchValidationResult{
        ID:        uuid.New().String(),
        Results:   make(map[string]*ValidationResult),
        Timestamp: time.Now(),
    }

    // Process requests concurrently with limit
    semaphore := make(chan struct{}, 10) // Max 10 concurrent validations
    resultsChan := make(chan struct {
        ID     string
        Result *ValidationResult
        Error  error
    }, len(requests))

    var wg sync.WaitGroup
    
    for _, request := range requests {
        wg.Add(1)
        go func(req ValidationRequest) {
            defer wg.Done()
            
            semaphore <- struct{}{}
            defer func() { <-semaphore }()
            
            result, err := vp.ValidateData(ctx, req)
            resultsChan <- struct {
                ID     string
                Result *ValidationResult
                Error  error
            }{
                ID:     req.ID,
                Result: result,
                Error:  err,
            }
        }(request)
    }

    // Wait for all validations to complete
    go func() {
        wg.Wait()
        close(resultsChan)
    }()

    // Collect results
    for res := range resultsChan {
        if res.Error == nil && res.Result != nil {
            batchResult.Results[res.ID] = res.Result
            if res.Result.Valid {
                batchResult.SuccessCount++
            } else {
                batchResult.FailureCount++
            }
        } else {
            batchResult.FailureCount++
        }
    }

    batchResult.TotalRequests = len(requests)
    batchResult.ProcessingTime = time.Since(startTime)

    return batchResult, nil
}

func (vp *ValidationPlatform) getSortedValidators() []Validator {
    // Sort validators by priority
    validators := make([]Validator, 0, len(vp.validators))
    for _, v := range vp.validators {
        validators = append(validators, v)
    }
    
    // Simple sort by priority (in practice, use sort.Slice)
    for i := 0; i < len(validators); i++ {
        for j := i + 1; j < len(validators); j++ {
            if validators[i].Priority() > validators[j].Priority() {
                validators[i], validators[j] = validators[j], validators[i]
            }
        }
    }
    
    return validators
}

func (vp *ValidationPlatform) mergeResults(result1, result2 *ValidationResult) *ValidationResult {
    merged := &ValidationResult{
        Valid:       result1.Valid && result2.Valid,
        Errors:      append(result1.Errors, result2.Errors...),
        Warnings:    append(result1.Warnings, result2.Warnings...),
        Suggestions: append(result1.Suggestions, result2.Suggestions...),
    }
    
    // Merge cleaned data
    if result1.CleanedData != nil && result2.CleanedData != nil {
        merged.CleanedData = make(map[string]interface{})
        for k, v := range result1.CleanedData {
            merged.CleanedData[k] = v
        }
        for k, v := range result2.CleanedData {
            merged.CleanedData[k] = v
        }
    }
    
    return merged
}

func (vp *ValidationPlatform) attemptRecovery(ctx context.Context, request ValidationRequest, result *ValidationResult) *ValidationResult {
    // Implement recovery logic
    return nil
}

// Monitor methods
func (vm *ValidationMonitor) RecordValidation(validationType string, success bool, duration time.Duration) {
    vm.mu.Lock()
    defer vm.mu.Unlock()

    if _, exists := vm.metrics[validationType]; !exists {
        vm.metrics[validationType] = &ValidationMetrics{
            ErrorsByType: make(map[string]int64),
        }
    }

    metrics := vm.metrics[validationType]
    metrics.TotalValidations++
    
    if success {
        metrics.SuccessCount++
    } else {
        metrics.FailureCount++
    }
    
    // Update average time
    if metrics.AvgValidationTime == 0 {
        metrics.AvgValidationTime = duration
    } else {
        metrics.AvgValidationTime = (metrics.AvgValidationTime + duration) / 2
    }
    
    metrics.LastValidation = time.Now()
}

func (vm *ValidationMonitor) CheckAlerts() {
    vm.mu.RLock()
    defer vm.mu.RUnlock()

    for validationType, metrics := range vm.metrics {
        // Check failure rate
        if metrics.TotalValidations > 0 {
            failureRate := float64(metrics.FailureCount) / float64(metrics.TotalValidations)
            if threshold, exists := vm.thresholds["failure_rate"]; exists && failureRate > threshold {
                alert := ValidationAlert{
                    ID:        uuid.New().String(),
                    Type:      "failure_rate",
                    Severity:  "high",
                    Message:   fmt.Sprintf("High failure rate for %s validations", validationType),
                    Metric:    "failure_rate",
                    Value:     failureRate,
                    Threshold: threshold,
                    Timestamp: time.Now(),
                }
                
                select {
                case vm.alerts <- alert:
                default:
                    // Alert channel full, log instead
                    log.Printf("Alert: %s", alert.Message)
                }
            }
        }
        
        // Check latency
        avgLatencyMs := metrics.AvgValidationTime.Milliseconds()
        if threshold, exists := vm.thresholds["avg_latency_ms"]; exists && float64(avgLatencyMs) > threshold {
            alert := ValidationAlert{
                ID:        uuid.New().String(),
                Type:      "latency",
                Severity:  "medium",
                Message:   fmt.Sprintf("High latency for %s validations", validationType),
                Metric:    "avg_latency_ms",
                Value:     float64(avgLatencyMs),
                Threshold: threshold,
                Timestamp: time.Now(),
            }
            
            select {
            case vm.alerts <- alert:
            default:
                log.Printf("Alert: %s", alert.Message)
            }
        }
    }
}

func (vm *ValidationMonitor) GetMetrics() map[string]*ValidationMetrics {
    vm.mu.RLock()
    defer vm.mu.RUnlock()
    
    // Return copy of metrics
    metricsCopy := make(map[string]*ValidationMetrics)
    for k, v := range vm.metrics {
        metricsCopy[k] = v
    }
    
    return metricsCopy
}

// Circuit breaker methods
func (cb *CircuitBreaker) Call(fn func() error) error {
    cb.mu.Lock()
    defer cb.mu.Unlock()

    switch cb.state {
    case "open":
        if time.Since(cb.lastFailure) > cb.resetTimeout {
            cb.state = "half-open"
            cb.failures = 0
        } else {
            return fmt.Errorf("circuit breaker is open")
        }
    }

    err := fn()
    if err != nil {
        cb.RecordFailure()
        return err
    }

    if cb.state == "half-open" {
        cb.state = "closed"
        cb.failures = 0
    }

    return nil
}

func (cb *CircuitBreaker) RecordFailure() {
    cb.failures++
    cb.lastFailure = time.Now()
    
    if cb.failures >= cb.failureThreshold {
        cb.state = "open"
    }
}

// Cache methods
func (vc *ValidationCache) Get(key string) *ValidationResult {
    vc.mu.RLock()
    defer vc.mu.RUnlock()

    if cached, exists := vc.cache[key]; exists {
        if time.Since(cached.Timestamp) < vc.ttl {
            return cached.Result
        }
        // Expired, remove it
        delete(vc.cache, key)
    }
    
    return nil
}

func (vc *ValidationCache) Set(key string, result *ValidationResult) {
    vc.mu.Lock()
    defer vc.mu.Unlock()

    // Implement LRU if cache is full
    if len(vc.cache) >= vc.maxSize {
        // Remove oldest entry (simplified)
        for k := range vc.cache {
            delete(vc.cache, k)
            break
        }
    }

    vc.cache[key] = &CachedValidation{
        Result:    result,
        Timestamp: time.Now(),
        Hash:      key,
    }
}

// Audit methods
func (al *AuditLogger) LogValidation(request ValidationRequest, result *ValidationResult) {
    al.mu.Lock()
    defer al.mu.Unlock()

    entry := AuditEntry{
        ID:        uuid.New().String(),
        Timestamp: time.Now(),
        Operation: "validation",
        UserID:    request.Context["user_id"].(string),
        DataType:  request.Type,
        Result:    fmt.Sprintf("valid=%v", result.Valid),
        Metadata: map[string]interface{}{
            "request_id":   request.ID,
            "error_count":  len(result.Errors),
            "warning_count": len(result.Warnings),
        },
    }

    if !result.Valid {
        entry.ErrorDetails = make([]string, len(result.Errors))
        for i, err := range result.Errors {
            entry.ErrorDetails[i] = fmt.Sprintf("%s: %s", err.Field, err.Message)
        }
    }

    al.entries = append(al.entries, entry)
    
    // Store to persistent storage
    if al.storage != nil {
        al.storage.Store(entry)
    }
}

// Memory audit storage (for demo)
type MemoryAuditStorage struct {
    entries []AuditEntry
    mu      sync.Mutex
}

func NewMemoryAuditStorage() *MemoryAuditStorage {
    return &MemoryAuditStorage{
        entries: []AuditEntry{},
    }
}

func (mas *MemoryAuditStorage) Store(entry AuditEntry) error {
    mas.mu.Lock()
    defer mas.mu.Unlock()
    mas.entries = append(mas.entries, entry)
    return nil
}

func (mas *MemoryAuditStorage) Query(filter AuditFilter) ([]AuditEntry, error) {
    mas.mu.Lock()
    defer mas.mu.Unlock()
    
    // Simple filtering implementation
    var results []AuditEntry
    for _, entry := range mas.entries {
        if filter.Matches(entry) {
            results = append(results, entry)
        }
    }
    
    return results, nil
}

type AuditFilter struct {
    StartTime  *time.Time
    EndTime    *time.Time
    UserID     string
    Operation  string
    ResultType string
}

func (af AuditFilter) Matches(entry AuditEntry) bool {
    if af.StartTime != nil && entry.Timestamp.Before(*af.StartTime) {
        return false
    }
    if af.EndTime != nil && entry.Timestamp.After(*af.EndTime) {
        return false
    }
    if af.UserID != "" && entry.UserID != af.UserID {
        return false
    }
    if af.Operation != "" && entry.Operation != af.Operation {
        return false
    }
    return true
}

// Example validators
type SchemaValidator struct {
    validator *validation.Validator
}

func NewSchemaValidator() *SchemaValidator {
    return &SchemaValidator{
        validator: validation.NewValidator(),
    }
}

func (sv *SchemaValidator) Name() string { return "schema" }
func (sv *SchemaValidator) Priority() int { return 1 }

func (sv *SchemaValidator) Validate(ctx context.Context, data interface{}) (*ValidationResult, error) {
    // Implementation of schema validation
    return &ValidationResult{Valid: true}, nil
}

// Rule executor
type RuleExecutor struct{}

func NewRuleExecutor() *RuleExecutor {
    return &RuleExecutor{}
}

func (re *RuleEngine) Execute(ctx context.Context, data interface{}, ruleIDs []string) *ValidationResult {
    // Implementation of rule execution
    return &ValidationResult{Valid: true}
}

// Example usage
func main() {
    fmt.Println("🏭 Enterprise Validation Platform")
    fmt.Println("================================")

    // Create platform
    platform := NewValidationPlatform()

    // Register validators
    platform.RegisterValidator(NewSchemaValidator())

    // Define test requests
    requests := []ValidationRequest{
        {
            ID:   "req-001",
            Type: "customer",
            Data: map[string]interface{}{
                "name":  "John Doe",
                "email": "john@example.com",
                "age":   30,
            },
            Options: ValidationOptions{
                StrictMode:    true,
                AuditRequired: true,
                CacheResults:  true,
            },
            Context: map[string]interface{}{
                "user_id": "user-123",
            },
        },
        {
            ID:   "req-002",
            Type: "order",
            Data: map[string]interface{}{
                "order_id": "ORD-123456",
                "total":    99.99,
                "items":    []interface{}{"item1", "item2"},
            },
            Options: ValidationOptions{
                EnableRecovery:  true,
                MaxRecoveryTime: 5 * time.Second,
            },
            Context: map[string]interface{}{
                "user_id": "user-456",
            },
        },
    }

    ctx := context.Background()

    // Validate batch
    fmt.Println("🔄 Processing batch validation...")
    batchResult, err := platform.ValidateBatch(ctx, requests)
    if err != nil {
        log.Fatalf("Batch validation failed: %v", err)
    }

    fmt.Printf("\n📊 Batch Results:\n")
    fmt.Printf("Total: %d | Success: %d | Failed: %d\n",
        batchResult.TotalRequests,
        batchResult.SuccessCount,
        batchResult.FailureCount)
    fmt.Printf("Processing Time: %v\n", batchResult.ProcessingTime)

    // Show individual results
    for id, result := range batchResult.Results {
        fmt.Printf("\n  Request %s: Valid=%v\n", id, result.Valid)
        if !result.Valid && len(result.Errors) > 0 {
            fmt.Printf("    Errors: %d\n", len(result.Errors))
        }
    }

    // Display metrics
    fmt.Printf("\n📈 Validation Metrics:\n")
    metrics := platform.monitor.GetMetrics()
    for validationType, m := range metrics {
        fmt.Printf("\n  %s:\n", validationType)
        fmt.Printf("    Total: %d\n", m.TotalValidations)
        fmt.Printf("    Success Rate: %.2f%%\n", 
            float64(m.SuccessCount)/float64(m.TotalValidations)*100)
        fmt.Printf("    Avg Time: %v\n", m.AvgValidationTime)
    }

    // Monitor alerts
    fmt.Printf("\n🚨 Monitoring Alerts:\n")
    go func() {
        for alert := range platform.monitor.alerts {
            fmt.Printf("  [%s] %s: %s (%.2f > %.2f)\n",
                alert.Severity,
                alert.Type,
                alert.Message,
                alert.Value,
                alert.Threshold)
        }
    }()

    // Give time for alerts to process
    time.Sleep(100 * time.Millisecond)
}
```

## Validation Best Practices

### 1. Schema Design
- **Start strict** - Begin with tight constraints
- **Document rules** - Clear descriptions for all validations
- **Version schemas** - Track schema evolution
- **Test extensively** - Validate edge cases

### 2. Error Recovery
- **Graduated approach** - Try simple fixes first
- **Preserve valid data** - Don't discard good portions
- **Learn from failures** - Improve prompts based on errors
- **Set limits** - Don't retry indefinitely

### 3. Security Validation
- **Input sanitization** - Clean all user inputs
- **Injection prevention** - Check for SQL/script injection
- **Size limits** - Prevent resource exhaustion
- **Domain validation** - Verify external references

### 4. Performance Optimization
- **Cache validations** - Reuse results when possible
- **Batch processing** - Validate multiple items together
- **Async validation** - Non-blocking for large datasets
- **Circuit breakers** - Prevent cascade failures

## Common Validation Patterns

### Field-Level Validation
- **Type checking** - Ensure correct data types
- **Format validation** - Email, phone, dates
- **Range validation** - Min/max for numbers
- **Pattern matching** - Regex for complex formats

### Cross-Field Validation
- **Dependent fields** - If A then B required
- **Mutual exclusion** - Either A or B, not both
- **Calculated fields** - Verify computed values
- **Temporal logic** - Date relationships

### Business Rule Validation
- **Domain constraints** - Industry-specific rules
- **Workflow validation** - State transitions
- **Authorization** - Permission-based validation
- **Compliance** - Regulatory requirements

### Semantic Validation
- **Logical consistency** - Data makes sense
- **Reference integrity** - Related data exists
- **Contextual validation** - Based on use case
- **AI-assisted** - LLM for complex validation

## Troubleshooting

### Common Issues

**"Schema too restrictive" errors**
- Relax constraints gradually
- Add format flexibility
- Implement type coercion
- Provide clear examples

**Poor recovery success**
- Enhance error messages
- Add specific examples
- Implement fallback strategies
- Consider partial acceptance

**Performance problems**
- Enable caching
- Optimize validation order
- Batch similar validations
- Profile bottlenecks

**Security false positives**
- Tune detection patterns
- Whitelist safe patterns
- Context-aware validation
- Manual review process

## Next Steps

🛡️ **Validation mastery achieved!** Continue with:

- **[Data Pipelines](data-pipelines.md)** - Build validation into pipelines
- **[Error Codes Reference](../reference/error-codes-reference.md)** - Comprehensive error guide
- **[Production Deployment](../advanced/production-deployment.md)** - Deploy validation at scale
- **[Troubleshooting](../advanced/troubleshooting.md)** - Advanced debugging

### Quick Reference

- **[Best Practices Checklist](../reference/best-practices-checklist.md)** - Validation checklist
- **[Configuration Reference](../reference/configuration-reference.md)** - Validation settings
- **[Schema Examples](../examples/schema-examples.md)** - Common patterns

---

**Need help with validation?** Check our [validation examples](../examples/validation-patterns.md) or join the discussion on [GitHub](https://github.com/lexlapax/go-llms/discussions).