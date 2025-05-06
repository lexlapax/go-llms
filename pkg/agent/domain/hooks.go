package domain

import (
	"context"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
)

// Hook provides callbacks for monitoring agent operations
type Hook interface {
	// BeforeGenerate is called before generating a response
	BeforeGenerate(ctx context.Context, messages []domain.Message)
	
	// AfterGenerate is called after generating a response
	AfterGenerate(ctx context.Context, response domain.Response, err error)
	
	// BeforeToolCall is called before executing a tool
	BeforeToolCall(ctx context.Context, tool string, params map[string]interface{})
	
	// AfterToolCall is called after executing a tool
	AfterToolCall(ctx context.Context, tool string, result interface{}, err error)
}