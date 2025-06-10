// ABOUTME: Demonstrates agent handoff patterns for delegating between specialized agents
// ABOUTME: Shows how to create a customer support system with automatic handoffs

package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// TechSupportAgent handles technical issues
type TechSupportAgent struct {
	*core.BaseAgentImpl
}

func NewTechSupportAgent() *TechSupportAgent {
	return &TechSupportAgent{
		BaseAgentImpl: core.NewBaseAgent("techSupport", "Handles technical issues and troubleshooting", domain.AgentTypeLLM),
	}
}

func (t *TechSupportAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	issue, _ := state.Get("issue")
	customerID, _ := state.Get("customer_id")

	log.Printf("[Tech Support] Handling issue for customer %v: %v\n", customerID, issue)

	output := domain.NewState()

	// Copy customer context
	if customerID != nil {
		output.Set("customer_id", customerID)
	}

	// Simulate technical troubleshooting
	issueStr := fmt.Sprint(issue)
	if strings.Contains(strings.ToLower(issueStr), "internet") {
		output.Set("output", "I've diagnosed an internet connectivity issue. Please try these steps:\n1. Restart your modem\n2. Check cable connections\n3. Run network diagnostics\nIf the issue persists, I can schedule a technician visit.")
		output.Set("resolution", "network_troubleshooting")
		output.Set("needs_escalation", false)
	} else if strings.Contains(strings.ToLower(issueStr), "email") {
		output.Set("output", "For email issues, please check:\n1. Your email settings\n2. Username and password\n3. Server configuration\nI can help you reconfigure your email client.")
		output.Set("resolution", "email_configuration")
		output.Set("needs_escalation", false)
	} else {
		output.Set("output", "This issue requires senior technical support. Transferring to a senior technician...")
		output.Set("needs_escalation", true)
		output.Set("escalation_reason", "Complex technical issue")
	}

	return output, nil
}

// BillingSupportAgent handles billing and payment issues
type BillingSupportAgent struct {
	*core.BaseAgentImpl
}

func NewBillingSupportAgent() *BillingSupportAgent {
	return &BillingSupportAgent{
		BaseAgentImpl: core.NewBaseAgent("billingSupport", "Handles billing, payments, and account issues", domain.AgentTypeLLM),
	}
}

func (b *BillingSupportAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	issue, _ := state.Get("issue")
	customerID, _ := state.Get("customer_id")

	log.Printf("[Billing Support] Handling billing issue for customer %v: %v\n", customerID, issue)

	output := domain.NewState()

	// Simulate billing support
	issueStr := fmt.Sprint(issue)
	if strings.Contains(strings.ToLower(issueStr), "refund") {
		output.Set("output", "I've processed your refund request. You should see the credit in 3-5 business days.")
		output.Set("resolution", "refund_processed")
		output.Set("amount", 49.99)
	} else if strings.Contains(strings.ToLower(issueStr), "charge") {
		output.Set("output", "I've reviewed the charge on your account. This was for your monthly subscription renewal.")
		output.Set("resolution", "charge_explained")
	} else {
		output.Set("output", "I've updated your billing information. Your next payment will use the new details.")
		output.Set("resolution", "billing_updated")
	}

	output.Set("customer_id", customerID)
	output.Set("needs_escalation", false)

	return output, nil
}

// SeniorSupportAgent handles escalated issues
type SeniorSupportAgent struct {
	*core.BaseAgentImpl
}

func NewSeniorSupportAgent() *SeniorSupportAgent {
	return &SeniorSupportAgent{
		BaseAgentImpl: core.NewBaseAgent("seniorSupport", "Handles escalated and complex issues", domain.AgentTypeLLM),
	}
}

func (s *SeniorSupportAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	issue, _ := state.Get("issue")
	customerID, _ := state.Get("customer_id")
	escalationReason, _ := state.Get("escalation_reason")

	log.Printf("[Senior Support] Handling escalated issue for customer %v: %v (Reason: %v)\n",
		customerID, issue, escalationReason)

	output := domain.NewState()
	output.Set("output", "As a senior support specialist, I've reviewed your case. I'm providing a custom solution and have added a service credit to your account for the inconvenience.")
	output.Set("resolution", "escalated_resolved")
	output.Set("service_credit", 25.00)
	output.Set("customer_id", customerID)

	return output, nil
}

func main() {
	log.Println("=== Agent Handoff Example - Customer Support System ===")
	log.Println()

	// Create specialized support agents
	techSupport := NewTechSupportAgent()
	billingSupport := NewBillingSupportAgent()
	seniorSupport := NewSeniorSupportAgent()

	// Create coordinator agent with sub-agents using new simplified API
	coordinator, err := core.NewLLMAgentWithSubAgentsFromString(
		"supportCoordinator",
		"mock", // In production: "openai/gpt-4"
		techSupport,
		billingSupport,
		seniorSupport,
	)
	if err != nil {
		log.Fatal("Failed to create coordinator:", err)
	}

	// Configure the coordinator
	coordinator.SetSystemPrompt(`You are a customer support coordinator. Route issues to the appropriate specialist:
- techSupport: For technical issues, connectivity problems, configuration
- billingSupport: For billing, payments, refunds, charges
- seniorSupport: For escalated or complex issues

Use the transfer_to_agent tool to delegate to the appropriate specialist.`)

	// Register agents for handoff
	_ = core.Register(techSupport)
	_ = core.Register(billingSupport)
	_ = core.Register(seniorSupport)

	// Display available tools (shows automatic sub-agent registration)
	fmt.Println("Coordinator's available tools:")
	for _, tool := range coordinator.ListTools() {
		if t, ok := coordinator.GetTool(tool); ok {
			fmt.Printf("- %s: %s\n", tool, t.Description())
		}
	}
	fmt.Println()

	ctx := context.Background()

	// Scenario 1: Technical Issue
	fmt.Println("=== Scenario 1: Technical Issue ===")
	result, err := coordinator.TransferTo(ctx, "techSupport", "Customer reporting internet issues",
		map[string]interface{}{
			"issue":       "My internet keeps disconnecting every few minutes",
			"customer_id": "CUST-12345",
		})
	if err != nil {
		log.Fatal("Handoff failed:", err)
	}
	if output, ok := result.Get("output"); ok {
		fmt.Printf("Response: %v\n", output)
	}
	fmt.Println()

	// Scenario 2: Billing Issue
	fmt.Println("=== Scenario 2: Billing Issue ===")
	result, err = coordinator.TransferTo(ctx, "billingSupport", "Customer requesting refund",
		map[string]interface{}{
			"issue":       "I was charged twice this month, need a refund",
			"customer_id": "CUST-67890",
		})
	if err != nil {
		log.Fatal("Handoff failed:", err)
	}
	if output, ok := result.Get("output"); ok {
		fmt.Printf("Response: %v\n", output)
	}
	if amount, ok := result.Get("amount"); ok {
		fmt.Printf("Refund amount: $%.2f\n", amount)
	}
	fmt.Println()

	// Scenario 3: Complex Issue requiring escalation
	fmt.Println("=== Scenario 3: Complex Technical Issue ===")

	// First, try tech support
	techResult, err := coordinator.TransferTo(ctx, "techSupport", "Complex technical issue",
		map[string]interface{}{
			"issue":       "Custom API integration failing with authentication",
			"customer_id": "CUST-99999",
		})
	if err != nil {
		log.Fatal("Handoff failed:", err)
	}

	// Check if escalation is needed
	if needsEscalation, ok := techResult.Get("needs_escalation"); ok && needsEscalation.(bool) {
		fmt.Println("Tech support requesting escalation...")

		// Prepare state for senior support with context
		escalationState := map[string]interface{}{
			"issue":             "Custom API integration failing with authentication",
			"customer_id":       "CUST-99999",
			"escalation_reason": techResult.Values()["escalation_reason"],
			"previous_agent":    "techSupport",
		}

		// Escalate to senior support
		seniorResult, err := coordinator.TransferTo(ctx, "seniorSupport",
			"Escalating complex issue", escalationState)
		if err != nil {
			log.Fatal("Escalation failed:", err)
		}

		if output, ok := seniorResult.Get("output"); ok {
			fmt.Printf("Senior Support Response: %v\n", output)
		}
		if credit, ok := seniorResult.Get("service_credit"); ok {
			fmt.Printf("Service credit applied: $%.2f\n", credit)
		}
	}
	fmt.Println()

	// Demonstrate handoff patterns
	fmt.Println("=== Handoff Patterns Demonstrated ===")
	fmt.Println("1. Automatic sub-agent registration as tools")
	fmt.Println("2. Direct handoff using TransferTo() method")
	fmt.Println("3. State passing between agents (customer_id maintained)")
	fmt.Println("4. Conditional escalation based on agent response")
	fmt.Println("5. Context preservation through handoff chain")
	fmt.Println()

	// Show simplified creation pattern
	fmt.Println("=== Simplified API Pattern ===")
	fmt.Println("Creating a multi-agent system is now as simple as:")
	fmt.Println(`
coordinator, err := core.NewLLMAgentWithSubAgentsFromString(
    "coordinator", 
    "openai/gpt-4",
    techSupport,
    billingSupport,
    seniorSupport,
)`)
}
