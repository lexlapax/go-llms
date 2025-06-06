// ABOUTME: Example demonstrating agent-level guardrails for input/output validation and safety
// ABOUTME: Shows how to implement content filtering, validation rules, and safety checks for agents

package main

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// ContentFilterGuardrail implements content validation with patterns and keywords
type ContentFilterGuardrail struct {
	name            string
	blockedPatterns []*regexp.Regexp
	sensitiveWords  []string
	maxLength       int
}

func NewContentFilterGuardrail(name string, patterns []string, keywords []string, maxLen int) (*ContentFilterGuardrail, error) {
	// Compile regex patterns
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern %s: %w", pattern, err)
		}
		compiled = append(compiled, re)
	}

	return &ContentFilterGuardrail{
		name:            name,
		blockedPatterns: compiled,
		sensitiveWords:  keywords,
		maxLength:       maxLen,
	}, nil
}

func (g *ContentFilterGuardrail) Name() string {
	return g.name
}

func (g *ContentFilterGuardrail) Type() domain.GuardrailType {
	return domain.GuardrailTypeBoth
}

func (g *ContentFilterGuardrail) Validate(ctx context.Context, state *domain.State) error {
	// Check user input
	if input, exists := state.Get("user_input"); exists {
		content := fmt.Sprintf("%v", input)

		// Check length
		if g.maxLength > 0 && len(content) > g.maxLength {
			return fmt.Errorf("content exceeds maximum length of %d characters", g.maxLength)
		}

		// Check blocked patterns
		for _, pattern := range g.blockedPatterns {
			if pattern.MatchString(content) {
				return fmt.Errorf("content contains blocked pattern")
			}
		}

		// Check sensitive keywords
		lowerContent := strings.ToLower(content)
		for _, keyword := range g.sensitiveWords {
			if strings.Contains(lowerContent, strings.ToLower(keyword)) {
				log.Printf("Warning: Content contains sensitive keyword: %s", keyword)
			}
		}
	}

	return nil
}

func (g *ContentFilterGuardrail) ValidateAsync(ctx context.Context, state *domain.State, timeout time.Duration) <-chan error {
	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)

		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		done := make(chan error, 1)
		go func() {
			done <- g.Validate(timeoutCtx, state)
		}()

		select {
		case err := <-done:
			errCh <- err
		case <-timeoutCtx.Done():
			errCh <- fmt.Errorf("validation timeout after %v", timeout)
		}
	}()

	return errCh
}

// PIIDetectionGuardrail checks for personally identifiable information
type PIIDetectionGuardrail struct {
	patterns map[string]*regexp.Regexp
}

func NewPIIDetectionGuardrail() *PIIDetectionGuardrail {
	return &PIIDetectionGuardrail{
		patterns: map[string]*regexp.Regexp{
			"ssn":         regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
			"credit_card": regexp.MustCompile(`\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b`),
			"email":       regexp.MustCompile(`[\w\.-]+@[\w\.-]+\.\w+`),
			"phone":       regexp.MustCompile(`\b\d{3}[-.]?\d{3}[-.]?\d{4}\b`),
		},
	}
}

func (g *PIIDetectionGuardrail) Name() string {
	return "pii-detection"
}

func (g *PIIDetectionGuardrail) Type() domain.GuardrailType {
	return domain.GuardrailTypeOutput
}

func (g *PIIDetectionGuardrail) Validate(ctx context.Context, state *domain.State) error {
	// Check output for PII
	if output, exists := state.Get("output"); exists {
		content := fmt.Sprintf("%v", output)

		for piiType, pattern := range g.patterns {
			if pattern.MatchString(content) {
				return fmt.Errorf("output may contain PII: %s", piiType)
			}
		}
	}

	return nil
}

func (g *PIIDetectionGuardrail) ValidateAsync(ctx context.Context, state *domain.State, timeout time.Duration) <-chan error {
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		errCh <- g.Validate(ctx, state)
	}()
	return errCh
}

// TopicRestrictionGuardrail ensures conversations stay on allowed topics
func TopicRestrictionGuardrail(allowed, forbidden []string) domain.Guardrail {
	return domain.NewGuardrailFunc(
		"topic-restriction",
		domain.GuardrailTypeInput,
		func(ctx context.Context, state *domain.State) error {
			input, _ := state.Get("user_input")
			content := strings.ToLower(fmt.Sprintf("%v", input))

			// Check forbidden topics
			for _, topic := range forbidden {
				if strings.Contains(content, strings.ToLower(topic)) {
					return fmt.Errorf("topic '%s' is not allowed", topic)
				}
			}

			// If allowed topics specified, ensure input matches one
			if len(allowed) > 0 {
				found := false
				for _, topic := range allowed {
					if strings.Contains(content, strings.ToLower(topic)) {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("input must be about allowed topics: %v", allowed)
				}
			}

			return nil
		},
	)
}

// RateLimitGuardrail tracks and limits request frequency
type RateLimitGuardrail struct {
	requestCount int
	maxRequests  int
	resetTime    time.Time
	window       time.Duration
}

func NewRateLimitGuardrail(maxRequests int, window time.Duration) *RateLimitGuardrail {
	return &RateLimitGuardrail{
		maxRequests: maxRequests,
		window:      window,
		resetTime:   time.Now().Add(window),
	}
}

func (r *RateLimitGuardrail) Name() string {
	return "rate-limit"
}

func (r *RateLimitGuardrail) Type() domain.GuardrailType {
	return domain.GuardrailTypeInput
}

func (r *RateLimitGuardrail) Validate(ctx context.Context, state *domain.State) error {
	// Reset counter if window expired
	if time.Now().After(r.resetTime) {
		r.requestCount = 0
		r.resetTime = time.Now().Add(r.window)
	}

	r.requestCount++
	if r.requestCount > r.maxRequests {
		return fmt.Errorf("rate limit exceeded: %d requests in %v (max: %d)",
			r.requestCount, r.window, r.maxRequests)
	}

	return nil
}

func (r *RateLimitGuardrail) ValidateAsync(ctx context.Context, state *domain.State, timeout time.Duration) <-chan error {
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		errCh <- r.Validate(ctx, state)
	}()
	return errCh
}

// GuardedAgent wraps an agent with guardrails
type GuardedAgent struct {
	*core.BaseAgentImpl
	innerAgent   domain.BaseAgent
	inputGuards  domain.GuardrailChain
	outputGuards domain.GuardrailChain
}

func NewGuardedAgent(name string, inner domain.BaseAgent) *GuardedAgent {
	return &GuardedAgent{
		BaseAgentImpl: core.NewBaseAgent(name, "Agent with guardrails", domain.AgentTypeCustom),
		innerAgent:    inner,
		inputGuards:   *domain.NewGuardrailChain("input-guards", domain.GuardrailTypeInput, true),
		outputGuards:  *domain.NewGuardrailChain("output-guards", domain.GuardrailTypeOutput, true),
	}
}

func (g *GuardedAgent) AddInputGuard(guard domain.Guardrail) *GuardedAgent {
	g.inputGuards.Add(guard)
	return g
}

func (g *GuardedAgent) AddOutputGuard(guard domain.Guardrail) *GuardedAgent {
	g.outputGuards.Add(guard)
	return g
}

func (g *GuardedAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	// Validate input
	if err := g.inputGuards.Validate(ctx, state); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	// Run inner agent
	result, err := g.innerAgent.Run(ctx, state)
	if err != nil {
		return nil, err
	}

	// Validate output
	if err := g.outputGuards.Validate(ctx, result); err != nil {
		return nil, fmt.Errorf("output validation failed: %w", err)
	}

	return result, nil
}

func main() {
	fmt.Println("=== Guardrails Example ===")

	// Create base agent
	baseAgent, err := core.NewAgentFromString("base-agent", "mock")
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Example 1: Content filtering
	fmt.Println("Example 1: Content Filtering")
	fmt.Println(strings.Repeat("-", 40))

	contentFilter, err := NewContentFilterGuardrail(
		"content-filter",
		[]string{
			`(?i)password`,
			`(?i)secret`,
			`\b\d{16}\b`, // Credit card pattern
		},
		[]string{"medical", "financial", "personal"},
		1000,
	)
	if err != nil {
		log.Fatalf("Failed to create content filter: %v", err)
	}

	// Test valid input
	state1 := domain.NewState()
	state1.Set("user_input", "Tell me about programming best practices")

	err = contentFilter.Validate(context.Background(), state1)
	if err != nil {
		fmt.Printf("❌ Input rejected: %v\n", err)
	} else {
		fmt.Println("✅ Input accepted")
	}

	// Test blocked input
	state2 := domain.NewState()
	state2.Set("user_input", "My password is abc123")

	err = contentFilter.Validate(context.Background(), state2)
	if err != nil {
		fmt.Printf("❌ Input rejected: %v\n", err)
	} else {
		fmt.Println("✅ Input accepted")
	}

	// Example 2: Using built-in guardrails
	fmt.Println("\nExample 2: Built-in Guardrails")
	fmt.Println(strings.Repeat("-", 40))

	// Test required keys guardrail
	requiredGuard := domain.RequiredKeysGuardrail("required-fields", "user_input", "context")

	state3 := domain.NewState()
	state3.Set("user_input", "Hello")
	// Missing "context" key

	err = requiredGuard.Validate(context.Background(), state3)
	if err != nil {
		fmt.Printf("❌ Validation failed: %v\n", err)
	}

	state3.Set("context", "conversation")
	err = requiredGuard.Validate(context.Background(), state3)
	if err == nil {
		fmt.Println("✅ All required fields present")
	}

	// Example 3: Guardrail chains
	fmt.Println("\nExample 3: Guardrail Chains")
	fmt.Println(strings.Repeat("-", 40))

	chain := domain.NewGuardrailChain("safety-chain", domain.GuardrailTypeInput, true)
	chain.Add(domain.RequiredKeysGuardrail("required", "user_input"))
	chain.Add(domain.MaxStateSizeGuardrail("size-limit", 10*1024)) // 10KB limit
	chain.Add(domain.ContentModerationGuardrail("moderation", []string{"spam", "abuse"}))

	state4 := domain.NewState()
	state4.Set("user_input", "This is a normal message")

	err = chain.Validate(context.Background(), state4)
	if err != nil {
		fmt.Printf("❌ Chain validation failed: %v\n", err)
	} else {
		fmt.Println("✅ Chain validation passed")
	}

	// Example 4: Guarded agent with multiple guardrails
	fmt.Println("\nExample 4: Agent with Multiple Guardrails")
	fmt.Println(strings.Repeat("-", 40))

	guardedAgent := NewGuardedAgent("safe-agent", baseAgent)
	guardedAgent.AddInputGuard(contentFilter)
	guardedAgent.AddInputGuard(TopicRestrictionGuardrail(
		[]string{"technology", "programming", "software"},
		[]string{"politics", "religion"},
	))
	guardedAgent.AddInputGuard(NewRateLimitGuardrail(3, 1*time.Minute))
	guardedAgent.AddOutputGuard(NewPIIDetectionGuardrail())

	// Test multiple requests
	testInputs := []string{
		"Tell me about software testing",
		"How do I improve code quality?",
		"What are design patterns?",
		"One more question about coding",
	}

	for i, input := range testInputs {
		fmt.Printf("\nRequest %d: %s\n", i+1, input)

		state := domain.NewState()
		state.Set("user_input", input)

		result, err := guardedAgent.Run(context.Background(), state)
		if err != nil {
			fmt.Printf("❌ Request failed: %v\n", err)
		} else {
			output, _ := result.Get("output")
			fmt.Printf("✅ Response: %v\n", output)
		}
	}

	// Example 5: Async validation
	fmt.Println("\nExample 5: Async Validation with Timeout")
	fmt.Println(strings.Repeat("-", 40))

	// Create a slow guardrail
	slowGuard := domain.NewGuardrailFunc(
		"slow-check",
		domain.GuardrailTypeInput,
		func(ctx context.Context, state *domain.State) error {
			// Simulate slow validation
			select {
			case <-time.After(2 * time.Second):
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
	)

	state5 := domain.NewState()
	state5.Set("user_input", "Test async validation")

	// Try with short timeout
	fmt.Println("Testing with 1s timeout...")
	errCh := slowGuard.ValidateAsync(context.Background(), state5, 1*time.Second)
	if err := <-errCh; err != nil {
		fmt.Printf("❌ Validation timed out: %v\n", err)
	}

	// Try with longer timeout
	fmt.Println("Testing with 3s timeout...")
	errCh = slowGuard.ValidateAsync(context.Background(), state5, 3*time.Second)
	if err := <-errCh; err != nil {
		fmt.Printf("❌ Validation failed: %v\n", err)
	} else {
		fmt.Println("✅ Validation completed")
	}

	fmt.Println("\n=== Guardrails Example Complete ===")
}
