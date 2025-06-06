package integration

// ABOUTME: Integration tests for agent handoff functionality
// ABOUTME: Tests delegation patterns and state passing between agents

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
)

// TestBasicHandoff tests basic handoff functionality between agents
func TestBasicHandoff(t *testing.T) {
	// Clean up registry after test
	defer core.Clear()
	// Create mock providers
	coordinatorProvider := provider.NewMockProvider()
	techSupportProvider := provider.NewMockProvider()
	billingSupportProvider := provider.NewMockProvider()

	// Tech support handler
	techSupportProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		// Look for technical issues
		for _, msg := range messages {
			if msg.Role == ldomain.RoleUser {
				for _, part := range msg.Content {
					if strings.Contains(part.Text, "internet") {
						return ldomain.Response{
							Content: "I've diagnosed the internet connectivity issue. Please restart your modem and check cable connections.",
						}, nil
					}
				}
			}
		}
		return ldomain.Response{
			Content: "I can help with technical issues. What problem are you experiencing?",
		}, nil
	})

	// Billing support handler
	billingSupportProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		// Look for billing issues
		for _, msg := range messages {
			if msg.Role == ldomain.RoleUser {
				for _, part := range msg.Content {
					if strings.Contains(part.Text, "refund") {
						return ldomain.Response{
							Content: "I've processed your refund request. You should see the credit in 3-5 business days.",
						}, nil
					}
				}
			}
		}
		return ldomain.Response{
			Content: "I can help with billing and payment issues.",
		}, nil
	})

	// Create agents
	techSupport := core.NewLLMAgent("techSupport", "Handles technical issues", core.LLMDeps{Provider: techSupportProvider})
	techSupport.SetSystemPrompt("You are a technical support specialist")

	billingSupport := core.NewLLMAgent("billingSupport", "Handles billing issues", core.LLMDeps{Provider: billingSupportProvider})
	billingSupport.SetSystemPrompt("You are a billing support specialist")

	// Register agents globally (required for handoff execution)
	err := core.Register(techSupport)
	if err != nil {
		t.Fatalf("Failed to register tech support: %v", err)
	}

	err = core.Register(billingSupport)
	if err != nil {
		t.Fatalf("Failed to register billing support: %v", err)
	}

	// Create coordinator with handoffs
	coordinator := core.NewLLMAgent("coordinator", "Routes requests", core.LLMDeps{Provider: coordinatorProvider})
	coordinator.SetSystemPrompt("You route customer requests to the appropriate support team")

	// Add as sub-agents which automatically create handoff capabilities
	err = coordinator.AddSubAgent(techSupport)
	if err != nil {
		t.Fatalf("Failed to add tech support: %v", err)
	}

	err = coordinator.AddSubAgent(billingSupport)
	if err != nil {
		t.Fatalf("Failed to add billing support: %v", err)
	}

	// Test 1: Technical issue handoff
	t.Run("Technical issue handoff", func(t *testing.T) {
		ctx := context.Background()
		state := domain.NewState()
		state.Set("issue", "My internet connection is not working")
		state.Set("customer_id", "CUST-12345")

		// Create input state with user_input
		inputState := map[string]interface{}{
			"user_input":  state.Values()["issue"],
			"customer_id": state.Values()["customer_id"],
		}

		// Use TransferTo method directly
		result, err := coordinator.TransferTo(ctx, "techSupport", "Customer needs technical help", inputState)
		if err != nil {
			t.Fatalf("Handoff failed: %v", err)
		}

		outputVal, _ := result.Get("output")
		output := fmt.Sprint(outputVal)

		if !strings.Contains(output, "restart your modem") {
			t.Errorf("Expected technical support response, got: %s", output)
		}
	})

	// Test 2: Billing issue handoff
	t.Run("Billing issue handoff", func(t *testing.T) {
		ctx := context.Background()
		state := domain.NewState()
		state.Set("issue", "I need a refund for my last payment")
		state.Set("customer_id", "CUST-67890")

		// Create input state with user_input
		inputState := map[string]interface{}{
			"user_input":  state.Values()["issue"],
			"customer_id": state.Values()["customer_id"],
		}

		// Use TransferTo method
		result, err := coordinator.TransferTo(ctx, "billingSupport", "Customer needs billing help", inputState)
		if err != nil {
			t.Fatalf("Handoff failed: %v", err)
		}

		outputVal, _ := result.Get("output")
		output := fmt.Sprint(outputVal)

		if !strings.Contains(output, "3-5 business days") {
			t.Errorf("Expected billing support response, got: %s", output)
		}
	})

	// Test 3: Invalid handoff target
	t.Run("Invalid handoff target", func(t *testing.T) {
		ctx := context.Background()
		state := domain.NewState()
		state.Set("issue", "General inquiry")

		// Try to transfer to non-existent agent
		inputState := map[string]interface{}{
			"user_input": state.Values()["issue"],
		}
		_, err := coordinator.TransferTo(ctx, "salesSupport", "Invalid transfer", inputState)
		if err == nil {
			t.Error("Expected error for invalid handoff target")
		}
		if !strings.Contains(err.Error(), "sub-agent") && !strings.Contains(err.Error(), "not found") {
			t.Errorf("Unexpected error message: %v", err)
		}
	})
}

// TestHandoffWithStatePreservation tests that state is preserved during handoffs
func TestHandoffWithStatePreservation(t *testing.T) {
	// Clean up registry after test
	defer core.Clear()
	// Create mock providers
	analyzerProvider := provider.NewMockProvider()
	processorProvider := provider.NewMockProvider()

	// Analyzer checks for specific state values
	analyzerProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		// Check if transaction info is in the messages
		var hasTransactionInfo bool
		for _, msg := range messages {
			if msg.Role == ldomain.RoleUser {
				for _, part := range msg.Content {
					if strings.Contains(part.Text, "TXN-99999") || strings.Contains(part.Text, "transaction") {
						hasTransactionInfo = true
						break
					}
				}
			}
		}

		if hasTransactionInfo {
			// Return analysis with transaction context
			return ldomain.Response{
				Content: "Analysis complete for TXN-99999. Risk level: HIGH",
			}, nil
		}

		// Default response
		return ldomain.Response{
			Content: "Analysis complete. Risk level: HIGH",
		}, nil
	})

	// Processor uses analysis results
	processorProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		// Look for risk level in messages
		for _, msg := range messages {
			if msg.Role == ldomain.RoleUser {
				for _, part := range msg.Content {
					if strings.Contains(part.Text, "Risk level: HIGH") {
						return ldomain.Response{
							Content: "High risk detected. Applying enhanced security measures.",
						}, nil
					}
				}
			}
		}
		return ldomain.Response{
			Content: "Processing completed with standard measures.",
		}, nil
	})

	// Create agents
	analyzer := core.NewLLMAgent("analyzer", "Analyzes data", core.LLMDeps{Provider: analyzerProvider})
	processor := core.NewLLMAgent("processor", "Processes results", core.LLMDeps{Provider: processorProvider})

	// Create coordinator
	coordinatorProvider := provider.NewMockProvider()
	coordinator := core.NewLLMAgent("coordinator", "Coordinates workflow", core.LLMDeps{Provider: coordinatorProvider})

	// Register agents globally (required for handoff execution)
	err := core.Register(analyzer)
	if err != nil {
		t.Fatalf("Failed to register analyzer: %v", err)
	}

	err = core.Register(processor)
	if err != nil {
		t.Fatalf("Failed to register processor: %v", err)
	}

	// Also add as sub-agents to coordinator
	err = coordinator.AddSubAgent(analyzer)
	if err != nil {
		t.Fatalf("Failed to add analyzer: %v", err)
	}

	err = coordinator.AddSubAgent(processor)
	if err != nil {
		t.Fatalf("Failed to add processor: %v", err)
	}

	// Test state preservation through handoff chain
	ctx := context.Background()
	state := domain.NewState()
	state.Set("transaction_id", "TXN-99999")
	state.Set("amount", 10000.00)
	state.Set("user_data", "Sensitive transaction data")

	// First handoff to analyzer
	inputState := map[string]interface{}{
		"user_input":     "Analyze this transaction",
		"transaction_id": state.Values()["transaction_id"],
		"amount":         state.Values()["amount"],
		"user_data":      state.Values()["user_data"],
	}
	analysisResult, err := coordinator.TransferTo(ctx, "analyzer", "Analyze this transaction", inputState)
	if err != nil {
		t.Fatalf("Analysis handoff failed: %v", err)
	}

	// Verify output was returned
	if output, _ := analysisResult.Get("output"); output == nil {
		t.Error("Expected output from analysis")
	}

	// Add analysis results to state
	for k, v := range analysisResult.Values() {
		state.Set(k, v)
	}

	// Second handoff to processor with accumulated state
	processState := map[string]interface{}{
		"user_input":      fmt.Sprintf("Process transaction with analysis: %v", analysisResult.Values()["output"]),
		"transaction_id":  state.Values()["transaction_id"],
		"amount":          state.Values()["amount"],
		"analysis_result": analysisResult.Values()["output"],
	}
	processingResult, err := coordinator.TransferTo(ctx, "processor",
		"Process based on analysis",
		processState)
	if err != nil {
		t.Fatalf("Processing handoff failed: %v", err)
	}

	// Verify final output includes high risk response
	outputVal, _ := processingResult.Get("output")
	output := fmt.Sprint(outputVal)
	if !strings.Contains(output, "enhanced security measures") {
		t.Errorf("Expected high risk processing, got: %s", output)
	}

	// Verify processing was based on analysis
	// The state values are not automatically preserved through handoffs,
	// but we explicitly passed them in the processState map
}

// TestHandoffWithConditionalRouting tests conditional handoff routing
func TestHandoffWithConditionalRouting(t *testing.T) {
	// Clean up registry after test
	defer core.Clear()
	// Create specialized agents
	urgentProvider := provider.NewMockProvider()
	urgentProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		return ldomain.Response{
			Content: "URGENT: Immediate action taken. Issue resolved with priority handling.",
		}, nil
	})

	normalProvider := provider.NewMockProvider()
	normalProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		return ldomain.Response{
			Content: "Request processed through normal channels. Resolution time: 24-48 hours.",
		}, nil
	})

	urgentHandler := core.NewLLMAgent("urgentHandler", "Handles urgent requests", core.LLMDeps{Provider: urgentProvider})
	normalHandler := core.NewLLMAgent("normalHandler", "Handles normal requests", core.LLMDeps{Provider: normalProvider})

	// Create router
	routerProvider := provider.NewMockProvider()
	router := core.NewLLMAgent("router", "Routes based on priority", core.LLMDeps{Provider: routerProvider})

	// Register handlers globally (required for handoff execution)
	err := core.Register(urgentHandler)
	if err != nil {
		t.Fatalf("Failed to register urgent handler: %v", err)
	}

	err = core.Register(normalHandler)
	if err != nil {
		t.Fatalf("Failed to register normal handler: %v", err)
	}

	// Also add as sub-agents to router
	err = router.AddSubAgent(urgentHandler)
	if err != nil {
		t.Fatalf("Failed to add urgent handler: %v", err)
	}

	err = router.AddSubAgent(normalHandler)
	if err != nil {
		t.Fatalf("Failed to add normal handler: %v", err)
	}

	// Test urgent routing
	t.Run("Urgent routing", func(t *testing.T) {
		ctx := context.Background()
		state := domain.NewState()
		state.Set("priority", "urgent")
		state.Set("issue", "System down")

		// Router should automatically choose urgent handler based on condition
		// For this test, we'll use TransferTo directly
		inputState := map[string]interface{}{
			"user_input": "System down - urgent help needed",
			"priority":   state.Values()["priority"],
			"issue":      state.Values()["issue"],
		}
		result, err := router.TransferTo(ctx, "urgentHandler", "Urgent issue", inputState)
		if err != nil {
			t.Fatalf("Urgent handoff failed: %v", err)
		}

		outputVal, _ := result.Get("output")
		output := fmt.Sprint(outputVal)
		if !strings.Contains(output, "URGENT") {
			t.Errorf("Expected urgent response, got: %s", output)
		}
	})

	// Test normal routing
	t.Run("Normal routing", func(t *testing.T) {
		ctx := context.Background()
		state := domain.NewState()
		state.Set("priority", "normal")
		state.Set("issue", "Feature request")

		inputState := map[string]interface{}{
			"user_input": "Feature request - add new report",
			"priority":   state.Values()["priority"],
			"issue":      state.Values()["issue"],
		}
		result, err := router.TransferTo(ctx, "normalHandler", "Normal request", inputState)
		if err != nil {
			t.Fatalf("Normal handoff failed: %v", err)
		}

		outputVal, _ := result.Get("output")
		output := fmt.Sprint(outputVal)
		if !strings.Contains(output, "24-48 hours") {
			t.Errorf("Expected normal response, got: %s", output)
		}
	})
}
