// Package domain provides core types and interfaces for the agent framework.
// This file defines hook interfaces for monitoring and intercepting agent
// operations, enabling logging, metrics collection, and debugging of workflows.
package domain

// ABOUTME: Defines hook interfaces for monitoring and intercepting agent operations
// ABOUTME: Enables logging, metrics collection, and debugging of agent workflows

import (
	"context"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
)

// Hook provides callbacks for monitoring agent operations.
// Hooks enable logging, metrics collection, debugging, and other
// cross-cutting concerns during agent execution and tool invocation.
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
