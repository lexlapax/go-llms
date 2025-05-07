package processor

import (
	"sync"

	schemaDomain "github.com/lexlapax/go-llms/pkg/structured/domain"
)

var (
	// defaultEnhancer is a global singleton instance of the PromptEnhancer
	defaultEnhancer schemaDomain.PromptEnhancer
	// enhancerMutex protects the initialization of the defaultEnhancer
	enhancerMutex sync.Once
)

// GetDefaultPromptEnhancer returns the singleton instance of the PromptEnhancer
// This avoids repeatedly creating new enhancer instances
func GetDefaultPromptEnhancer() schemaDomain.PromptEnhancer {
	enhancerMutex.Do(func() {
		defaultEnhancer = NewPromptEnhancer()
	})
	return defaultEnhancer
}