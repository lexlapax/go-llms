# Agent Guardrails Example

This example demonstrates how to implement agent-level guardrails for input/output validation and safety in the agent framework. Guardrails help ensure that agents operate within defined boundaries and don't process or generate harmful content.

## Features

- **Content Filtering**: Block patterns and sensitive keywords
- **PII Detection**: Identify and prevent exposure of personal information
- **Topic Restrictions**: Keep conversations within allowed topics
- **Rate Limiting**: Prevent abuse through request throttling
- **Built-in Guardrails**: Use pre-built validation rules
- **Guardrail Chains**: Combine multiple guardrails
- **Async Validation**: Support for timeout-based validation

## Running the Example

```bash
go run main.go
```

## Key Concepts

### 1. Guardrail Interface

All guardrails implement the `domain.Guardrail` interface:

```go
type Guardrail interface {
    Name() string
    Type() GuardrailType
    Validate(ctx context.Context, state *State) error
    ValidateAsync(ctx context.Context, state *State, timeout time.Duration) <-chan error
}
```

### 2. Guardrail Types

- **Input**: Validates user input before processing
- **Output**: Validates agent output before returning
- **Both**: Validates both input and output

### 3. Content Filtering

The example shows custom content filtering:

```go
contentFilter := NewContentFilterGuardrail(
    "content-filter",
    []string{`(?i)password`, `(?i)secret`}, // Blocked patterns
    []string{"medical", "financial"},        // Sensitive keywords
    1000,                                    // Max length
)
```

### 4. Built-in Guardrails

Go-LLMs provides several built-in guardrails:

- `RequiredKeysGuardrail`: Ensures required state keys exist
- `MaxStateSizeGuardrail`: Limits state size
- `MessageCountGuardrail`: Limits conversation length
- `ContentModerationGuardrail`: Blocks prohibited words

### 5. Guardrail Chains

Combine multiple guardrails with configurable behavior:

```go
chain := domain.NewGuardrailChain("safety", GuardrailTypeInput, true)
chain.Add(requiredKeys).Add(sizeLimit).Add(contentFilter)
```

### 6. Guarded Agents

Wrap any agent with guardrails:

```go
guardedAgent := NewGuardedAgent("safe-agent", baseAgent)
guardedAgent.AddInputGuard(contentFilter)
guardedAgent.AddOutputGuard(piiDetector)
```

## Implementation Examples

### PII Detection

```go
type PIIDetectionGuardrail struct {
    patterns map[string]*regexp.Regexp
}

// Detects SSN, credit cards, emails, phone numbers
patterns := map[string]*regexp.Regexp{
    "ssn": regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
    "credit_card": regexp.MustCompile(`\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b`),
}
```

### Topic Restrictions

```go
topicGuard := TopicRestrictionGuardrail(
    []string{"technology", "programming"}, // Allowed
    []string{"politics", "religion"},      // Forbidden
)
```

### Rate Limiting

```go
rateGuard := NewRateLimitGuardrail(
    10,              // Max requests
    1*time.Minute,   // Time window
)
```

## Best Practices

1. **Layer Guards**: Use multiple guardrails for defense in depth
2. **Fail Fast**: Stop processing early when violations occur
3. **Log Warnings**: Track sensitive content without blocking
4. **Async Validation**: Use timeouts for expensive checks
5. **Custom Guards**: Create domain-specific validation rules

## Use Cases

- **Content Moderation**: Filter inappropriate content
- **Data Protection**: Prevent PII leakage
- **Compliance**: Enforce regulatory requirements
- **Safety**: Prevent harmful outputs
- **Resource Protection**: Limit usage and prevent abuse

## Extensions

You could extend this example to:

- Add machine learning-based content classification
- Implement dynamic rule updates
- Create audit logs for all validations
- Add metrics for guardrail performance
- Integrate with external validation services