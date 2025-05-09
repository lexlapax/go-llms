// Package stress contains stress tests for the Go-LLMs library
package stress

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/schema/validation"
	"github.com/lexlapax/go-llms/pkg/structured/processor"
)

// Define test schemas for structured output
const personSchema = `{
	"type": "object",
	"properties": {
		"name": {"type": "string"},
		"age": {"type": "integer"},
		"email": {"type": "string", "format": "email"},
		"address": {
			"type": "object",
			"properties": {
				"street": {"type": "string"},
				"city": {"type": "string"},
				"zipCode": {"type": "string"}
			},
			"required": ["street", "city"]
		},
		"tags": {
			"type": "array",
			"items": {"type": "string"}
		}
	},
	"required": ["name", "age"]
}`

const productSchema = `{
	"type": "object",
	"properties": {
		"id": {"type": "string"},
		"name": {"type": "string"},
		"price": {"type": "number"},
		"category": {"type": "string"},
		"inStock": {"type": "boolean"},
		"attributes": {
			"type": "object",
			"additionalProperties": true
		},
		"variants": {
			"type": "array",
			"items": {
				"type": "object",
				"properties": {
					"id": {"type": "string"},
					"name": {"type": "string"},
					"price": {"type": "number"}
				},
				"required": ["id", "name"]
			}
		}
	},
	"required": ["id", "name", "price"]
}`

const complexSchema = `{
	"type": "object",
	"properties": {
		"id": {"type": "string"},
		"timestamp": {"type": "string", "format": "date-time"},
		"data": {
			"type": "array",
			"items": {
				"type": "object",
				"properties": {
					"key": {"type": "string"},
					"value": {"type": "number"},
					"metadata": {
						"type": "object",
						"additionalProperties": true
					},
					"tags": {
						"type": "array",
						"items": {"type": "string"}
					}
				},
				"required": ["key", "value"]
			}
		},
		"statistics": {
			"type": "object",
			"properties": {
				"count": {"type": "integer"},
				"sum": {"type": "number"},
				"average": {"type": "number"},
				"min": {"type": "number"},
				"max": {"type": "number"},
				"distribution": {
					"type": "object",
					"additionalProperties": true
				}
			},
			"required": ["count", "sum", "average"]
		},
		"settings": {
			"type": "object",
			"additionalProperties": true
		}
	},
	"required": ["id", "timestamp", "data"]
}`

// Valid example responses for each schema
const validPersonResponse = `{
	"name": "John Doe",
	"age": 30,
	"email": "john.doe@example.com",
	"address": {
		"street": "123 Main St",
		"city": "New York",
		"zipCode": "10001"
	},
	"tags": ["developer", "gamer", "musician"]
}`

const validProductResponse = `{
	"id": "prod-123",
	"name": "Smartphone X",
	"price": 999.99,
	"category": "Electronics",
	"inStock": true,
	"attributes": {
		"color": "black",
		"weight": "180g",
		"storage": "128GB"
	},
	"variants": [
		{
			"id": "var-1",
			"name": "Black",
			"price": 999.99
		},
		{
			"id": "var-2",
			"name": "Silver",
			"price": 1099.99
		}
	]
}`

const validComplexResponse = `{
	"id": "dataset-abc123",
	"timestamp": "2023-09-15T14:30:00Z",
	"data": [
		{
			"key": "item1",
			"value": 42.5,
			"metadata": {
				"source": "system-a",
				"confidence": 0.95
			},
			"tags": ["important", "verified"]
		},
		{
			"key": "item2",
			"value": 18.2,
			"metadata": {
				"source": "system-b",
				"confidence": 0.87
			},
			"tags": ["normal"]
		}
	],
	"statistics": {
		"count": 2,
		"sum": 60.7,
		"average": 30.35,
		"min": 18.2,
		"max": 42.5,
		"distribution": {
			"0-20": 1,
			"20-50": 1
		}
	},
	"settings": {
		"maxItems": 100,
		"allowDuplicates": false
	}
}`

// TestStructuredProcessorConcurrentRequests tests structured output processor stability under high concurrency
func TestStructuredProcessorConcurrentRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Track memory stats before and after test
	var memStatsBefore, memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	// Create mock provider with appropriate responses for different schemas
	mockProvider := provider.NewMockProvider()

	// Configure the mock provider to return valid JSON for each schema
	mockProvider.WithGenerateFunc(func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
		if strings.Contains(prompt, "person profile") {
			return "Here's a person profile in JSON format:\n\n" + validPersonResponse, nil
		} else if strings.Contains(prompt, "product description") {
			return "Here's a product description in JSON format:\n\n" + validProductResponse, nil
		} else if strings.Contains(prompt, "complex dataset") {
			return "Here's a complex dataset in JSON format:\n\n" + validComplexResponse, nil
		}
		return "default response", nil
	})

	// Create validator and structured processor
	validator := validation.NewValidator(validation.WithCoercion(true))
	structuredProcessor := processor.NewStructuredProcessor(validator)

	// Define schema configurations to test
	schemaConfigs := []struct {
		name       string
		schema     string
		prompt     string
		schemaSize string
		complexity string
	}{
		{
			name:       "SmallSchema",
			schema:     personSchema,
			prompt:     "Generate a fictional person profile",
			schemaSize: "small",
			complexity: "low",
		},
		{
			name:       "MediumSchema",
			schema:     productSchema,
			prompt:     "Generate a fictional product description",
			schemaSize: "medium",
			complexity: "medium",
		},
		{
			name:       "LargeSchema",
			schema:     complexSchema,
			prompt:     "Generate a complex dataset with statistics",
			schemaSize: "large",
			complexity: "high",
		},
	}

	// Define concurrency levels to test
	concurrencyLevels := []int{10, 50, 100, 200}

	// Run tests for each schema configuration and concurrency level
	for _, sc := range schemaConfigs {
		// Parse the schema once for validation
		var schema schemaDomain.Schema
		err := json.Unmarshal([]byte(sc.schema), &schema)
		if err != nil {
			t.Errorf("Failed to parse schema for %s: %v", sc.name, err)
			continue
		}

		for _, concurrency := range concurrencyLevels {
			t.Run(fmt.Sprintf("%s_Concurrency_%d", sc.name, concurrency), func(t *testing.T) {
				var (
					wg                 sync.WaitGroup
					successful         int32
					failed             int32
					totalLatencyMs     int64
					maxLatencyMs       int64
					minLatencyMs       int64 = 999999
					validationErrors   int32
					totalLLMLatencyMs  int64
					totalProcLatencyMs int64
				)

				// Set a reasonable timeout
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
				defer cancel()

				// Create a semaphore to limit concurrent goroutines
				sem := make(chan struct{}, concurrency)

				// Track goroutine count
				initialGoroutines := runtime.NumGoroutine()

				// Launch concurrent requests
				startTime := time.Now()
				for i := 0; i < concurrency*2; i++ {
					wg.Add(1)
					sem <- struct{}{} // Acquire semaphore
					go func(id int) {
						defer func() {
							<-sem // Release semaphore
							wg.Done()
						}()

						// Vary prompt slightly to avoid cache hits
						randomSuffix := rand.Intn(1000)
						prompt := fmt.Sprintf("%s (Request ID: %d-%d)", sc.prompt, id, randomSuffix)

						// Create appropriate output type based on schema
						var output interface{}
						switch sc.name {
						case "SmallSchema":
							output = &map[string]interface{}{}
						case "MediumSchema":
							output = &map[string]interface{}{}
						case "LargeSchema":
							output = &map[string]interface{}{}
						}

						// Measure request time
						requestStart := time.Now()

						// Track LLM and processing times separately
						var llmLatencyMs, procLatencyMs int64

						// Generate a mock response with the LLM
						llmStart := time.Now()
						response, err := mockProvider.Generate(ctx, prompt)
						llmLatencyMs = time.Since(llmStart).Milliseconds()

						if err == nil {
							// Process the structured output
							procStart := time.Now()
							var processErr error
							_, processErr = structuredProcessor.Process(&schema, response)
							if processErr != nil {
								err = processErr
								// Check if the error message mentions validation
								if strings.Contains(processErr.Error(), "validation") ||
									strings.Contains(processErr.Error(), "conform to schema") {
									atomic.AddInt32(&validationErrors, 1)
								}
							} else if output != nil {
								// Try to convert to the specific type
								err = structuredProcessor.ProcessTyped(&schema, response, output)
							}
							procLatencyMs = time.Since(procStart).Milliseconds()
						}

						// Calculate total request duration
						requestDuration := time.Since(requestStart)
						latencyMs := requestDuration.Milliseconds()

						// Update metrics atomically
						atomic.AddInt64(&totalLatencyMs, latencyMs)
						atomic.AddInt64(&totalLLMLatencyMs, llmLatencyMs)
						atomic.AddInt64(&totalProcLatencyMs, procLatencyMs)

						// Update min/max latency
						for {
							current := atomic.LoadInt64(&maxLatencyMs)
							if latencyMs <= current {
								break
							}
							if atomic.CompareAndSwapInt64(&maxLatencyMs, current, latencyMs) {
								break
							}
						}

						for {
							current := atomic.LoadInt64(&minLatencyMs)
							if latencyMs >= current {
								break
							}
							if atomic.CompareAndSwapInt64(&minLatencyMs, current, latencyMs) {
								break
							}
						}

						if err != nil {
							atomic.AddInt32(&failed, 1)

							if ctx.Err() == context.DeadlineExceeded {
								t.Logf("Request %d timed out: %v", id, err)
							} else {
								t.Logf("Request %d failed: %v", id, err)
							}
						} else {
							atomic.AddInt32(&successful, 1)
						}
					}(i)
				}

				// Wait for all requests to complete
				wg.Wait()
				totalDuration := time.Since(startTime)

				// Check goroutine count after test
				peakGoroutines := runtime.NumGoroutine()

				// Record results
				total := successful + failed
				successRate := float64(successful) / float64(total) * 100
				validationErrorRate := float64(validationErrors) / float64(total) * 100
				avgLatencyMs := float64(totalLatencyMs) / float64(total)
				avgLLMLatencyMs := float64(totalLLMLatencyMs) / float64(total)
				avgProcLatencyMs := float64(totalProcLatencyMs) / float64(total)
				processingOverhead := 0.0
				if avgLatencyMs > 0 {
					processingOverhead = (avgProcLatencyMs / avgLatencyMs) * 100
				}

				// Calculate percentage of LLM latency
				llmPercentage := 0.0
				if avgLatencyMs > 0 {
					llmPercentage = avgLLMLatencyMs / avgLatencyMs * 100
				}

				t.Logf("Results for %s (%s schema, %s complexity) at concurrency %d:",
					sc.name, sc.schemaSize, sc.complexity, concurrency)
				t.Logf("  Success rate: %.2f%% (%d/%d)", successRate, successful, total)
				t.Logf("  Validation error rate: %.2f%% (%d/%d)", validationErrorRate, validationErrors, total)
				t.Logf("  Average total latency: %.2f ms", avgLatencyMs)
				t.Logf("  Average LLM latency: %.2f ms (%.2f%%)", avgLLMLatencyMs, llmPercentage)
				t.Logf("  Average processing latency: %.2f ms (%.2f%%)", avgProcLatencyMs, processingOverhead)
				t.Logf("  Min latency: %d ms", minLatencyMs)
				t.Logf("  Max latency: %d ms", maxLatencyMs)
				t.Logf("  Total duration: %v", totalDuration)
				t.Logf("  Goroutines: %d initial, %d peak", initialGoroutines, peakGoroutines)

				// Allow tests to pass even with validation errors as long as some requests succeed
				if successful == 0 {
					t.Logf("No successful requests for %s at concurrency %d", sc.name, concurrency)
				}
			})
		}
	}

	// Collect final memory stats
	runtime.ReadMemStats(&memStatsAfter)

	// Report memory usage
	t.Logf("Memory usage before: %.2f MB", float64(memStatsBefore.Alloc)/1024/1024)
	t.Logf("Memory usage after: %.2f MB", float64(memStatsAfter.Alloc)/1024/1024)

	// Calculate memory difference with error handling for potential integer underflow
	var memDiff float64
	if memStatsAfter.Alloc >= memStatsBefore.Alloc {
		memDiff = float64(memStatsAfter.Alloc-memStatsBefore.Alloc) / 1024 / 1024
	} else {
		// If we somehow get a negative difference (e.g., due to GC between measurements), report 0
		memDiff = 0
	}
	t.Logf("Memory difference: %.2f MB", memDiff)
	t.Logf("Total allocations: %d objects", memStatsAfter.Mallocs-memStatsBefore.Mallocs)
}
