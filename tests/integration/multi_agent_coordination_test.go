// ABOUTME: Integration tests for multi-agent coordination and collaboration patterns
// ABOUTME: Tests hierarchical agent systems, delegation, and result aggregation

package integration

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/workflow"
	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	"github.com/lexlapax/go-llms/pkg/testutils/fixtures"
)

// Note: Using centralized agent fixtures from pkg/testutils/fixtures
// MockCoordinatorAgent and MockSpecialistAgent have been replaced with fixtures.CoordinatorMockAgent and fixtures.SpecialistMockAgent

// TestBasicMultiAgentCoordination tests basic coordinator-specialist pattern
func TestBasicMultiAgentCoordination(t *testing.T) {
	// Create mock provider
	mockProvider := provider.NewMockProvider()

	// Create coordinator agent
	deps := core.LLMDeps{
		Provider: mockProvider,
	}
	coordinator := core.NewLLMAgent("coordinator", "Main coordinator", deps)
	coordinator.SetSystemPrompt("You coordinate tasks between specialists.")

	// Create specialist sub-agents using fixtures
	dataSpecialist := fixtures.SpecialistMockAgent("data_specialist", "data_analysis", 0)
	mlSpecialist := fixtures.SpecialistMockAgent("ml_specialist", "machine learning", 0)

	// Add specialists as sub-agents
	err := coordinator.AddSubAgent(dataSpecialist)
	if err != nil {
		t.Fatalf("Failed to add data specialist: %v", err)
	}

	err = coordinator.AddSubAgent(mlSpecialist)
	if err != nil {
		t.Fatalf("Failed to add ML specialist: %v", err)
	}

	// Verify sub-agents were added
	subAgents := coordinator.SubAgents()
	if len(subAgents) != 2 {
		t.Fatalf("Expected 2 sub-agents, got %d", len(subAgents))
	}

	// Verify transfer_to_agent tool was added automatically
	tools := coordinator.ListTools()
	hasTransferTool := false
	for _, tool := range tools {
		if tool == "transfer_to_agent" {
			hasTransferTool = true
			break
		}
	}
	if !hasTransferTool {
		t.Error("transfer_to_agent tool not found")
	}

	// Set up mock provider to simulate delegation
	responses := []ldomain.Response{
		{Content: `{"tool": "data_specialist", "params": {"task": "analyze customer data"}}`},
		{Content: "Analysis complete. The data shows positive trends."},
	}
	responseIndex := 0

	mockProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		if responseIndex >= len(responses) {
			return ldomain.Response{}, fmt.Errorf("no more responses")
		}
		resp := responses[responseIndex]
		responseIndex++
		return resp, nil
	})

	// Execute coordinator
	ctx := context.Background()
	state := domain.NewState()
	state.Set("user_input", "Analyze our customer data for ML insights")

	result, err := coordinator.Run(ctx, state)
	if err != nil {
		t.Fatalf("Coordinator execution failed: %v", err)
	}

	// Verify result
	output, exists := result.Get("output")
	if !exists {
		t.Fatal("No output in result")
	}

	outputStr, ok := output.(string)
	if !ok {
		t.Fatalf("Output is not string: %T", output)
	}

	if !strings.Contains(outputStr, "Analysis complete") {
		t.Errorf("Unexpected output: %s", outputStr)
	}
}

// TestHierarchicalMultiAgentSystem tests multi-level agent hierarchy
func TestHierarchicalMultiAgentSystem(t *testing.T) {
	// Create mock provider
	mockProvider := provider.NewMockProvider()
	deps := core.LLMDeps{
		Provider: mockProvider,
	}

	// Level 1: CEO Agent
	ceo := core.NewLLMAgent("ceo", "Chief Executive Officer", deps)
	ceo.SetSystemPrompt("You are the CEO who delegates to department heads.")

	// Level 2: Department Heads
	cto := core.NewLLMAgent("cto", "Chief Technology Officer", deps)
	cto.SetSystemPrompt("You are the CTO who manages tech teams.")

	cmo := core.NewLLMAgent("cmo", "Chief Marketing Officer", deps)
	cmo.SetSystemPrompt("You are the CMO who manages marketing teams.")

	// Level 3: Team Leads using fixtures
	devLead := fixtures.SpecialistMockAgent("dev_lead", "development", 0)
	qaLead := fixtures.SpecialistMockAgent("qa_lead", "quality assurance", 0)
	marketingLead := fixtures.SpecialistMockAgent("marketing_lead", "marketing campaigns", 0)

	// Build hierarchy
	err := cto.AddSubAgent(devLead)
	if err != nil {
		t.Fatalf("Failed to add dev lead: %v", err)
	}

	err = cto.AddSubAgent(qaLead)
	if err != nil {
		t.Fatalf("Failed to add QA lead: %v", err)
	}

	err = cmo.AddSubAgent(marketingLead)
	if err != nil {
		t.Fatalf("Failed to add marketing lead: %v", err)
	}

	err = ceo.AddSubAgent(cto)
	if err != nil {
		t.Fatalf("Failed to add CTO: %v", err)
	}

	err = ceo.AddSubAgent(cmo)
	if err != nil {
		t.Fatalf("Failed to add CMO: %v", err)
	}

	// Verify hierarchy
	if len(ceo.SubAgents()) != 2 {
		t.Errorf("CEO should have 2 sub-agents, got %d", len(ceo.SubAgents()))
	}

	if len(cto.SubAgents()) != 2 {
		t.Errorf("CTO should have 2 sub-agents, got %d", len(cto.SubAgents()))
	}

	if len(cmo.SubAgents()) != 1 {
		t.Errorf("CMO should have 1 sub-agent, got %d", len(cmo.SubAgents()))
	}

	// Test agent lookup
	foundCTO := ceo.FindSubAgent("cto")
	if foundCTO == nil {
		t.Error("Could not find CTO sub-agent")
	}

	// Test nested lookup
	foundDev := ceo.FindAgent("dev_lead")
	if foundDev == nil {
		t.Error("Could not find dev_lead through CEO")
	}
}

// TestMultiAgentWorkflowIntegration tests agents in workflow patterns
func TestMultiAgentWorkflowIntegration(t *testing.T) {
	// Create research team using fixtures
	webResearcher := fixtures.SpecialistMockAgent("web_researcher", "research", 50*time.Millisecond)
	academicResearcher := fixtures.SpecialistMockAgent("academic_researcher", "research", 50*time.Millisecond)

	// Create research team workflow (parallel)
	researchTeam := workflow.NewParallelAgent("research_team")
	researchTeam.AddAgent(webResearcher)
	researchTeam.AddAgent(academicResearcher)

	// Create analysis team using fixtures
	dataAnalyst := fixtures.SpecialistMockAgent("data_analyst", "data_analysis", 30*time.Millisecond)
	trendAnalyst := fixtures.SpecialistMockAgent("trend_analyst", "data_analysis", 30*time.Millisecond)

	// Create analysis team workflow (parallel)
	analysisTeam := workflow.NewParallelAgent("analysis_team")
	analysisTeam.AddAgent(dataAnalyst)
	analysisTeam.AddAgent(trendAnalyst)

	// Create report writer using fixture
	reportWriter := fixtures.SpecialistMockAgent("report_writer", "report writing", 20*time.Millisecond)

	// Create main workflow (sequential: research -> analysis -> report)
	mainWorkflow := workflow.NewSequentialAgent("main_workflow")
	mainWorkflow.AddAgent(researchTeam)
	mainWorkflow.AddAgent(analysisTeam)
	mainWorkflow.AddAgent(reportWriter)

	// Execute workflow
	ctx := context.Background()
	state := domain.NewState()
	state.Set("task", "Research AI trends for 2024")

	startTime := time.Now()
	result, err := mainWorkflow.Run(ctx, state)
	duration := time.Since(startTime)

	if err != nil {
		t.Fatalf("Workflow execution failed: %v", err)
	}

	// Verify result contains output
	output, exists := result.Get("output")
	if !exists {
		t.Fatal("No output in result")
	}

	// Verify it's from the report writer (last in sequence)
	outputStr := fmt.Sprintf("%v", output)
	if !strings.Contains(outputStr, "report writing") {
		t.Errorf("Expected output from report writer, got: %s", outputStr)
	}

	// Verify parallel execution saved time
	// Sequential would take 50+50+30+30+20 = 180ms minimum
	// Parallel should take roughly 50+30+20 = 100ms
	if duration > 150*time.Millisecond {
		t.Logf("Warning: Workflow took longer than expected: %v", duration)
	}
}

// TestMultiAgentCommunication tests agent-to-agent communication
func TestMultiAgentCommunication(t *testing.T) {
	// Create mock provider
	mockProvider := provider.NewMockProvider()
	deps := core.LLMDeps{
		Provider: mockProvider,
	}

	// Create negotiator agents
	buyer := core.NewLLMAgent("buyer", "Buyer agent", deps)
	buyer.SetSystemPrompt("You are a buyer negotiating prices.")

	seller := core.NewLLMAgent("seller", "Seller agent", deps)
	seller.SetSystemPrompt("You are a seller negotiating prices.")

	// Create mediator that has both as sub-agents
	mediator := core.NewLLMAgent("mediator", "Mediator agent", deps)
	mediator.SetSystemPrompt("You mediate between buyer and seller.")

	err := mediator.AddSubAgent(buyer)
	if err != nil {
		t.Fatalf("Failed to add buyer: %v", err)
	}

	err = mediator.AddSubAgent(seller)
	if err != nil {
		t.Fatalf("Failed to add seller: %v", err)
	}

	// Test TransferTo convenience method
	ctx := context.Background()

	// Set up mock responses for negotiation
	mockProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		return ldomain.Response{Content: "I'll start with an offer of $100"}, nil
	})

	// Test finding sub-agent
	foundBuyer := mediator.FindSubAgent("buyer")
	if foundBuyer == nil {
		t.Fatal("Could not find buyer sub-agent")
	}

	// Test executing the sub-agent directly
	buyerState := domain.NewState()
	buyerState.Set("user_input", "Make an offer for a laptop")
	buyerState.Set("item", "laptop")
	buyerState.Set("budget", 1000)

	buyerResult, err := foundBuyer.Run(ctx, buyerState)
	if err != nil {
		t.Fatalf("Buyer execution failed: %v", err)
	}

	if buyerResult == nil {
		t.Fatal("Buyer result is nil")
	}

	// Verify buyer produced output
	output, exists := buyerResult.Get("output")
	if !exists {
		t.Fatal("No output from buyer")
	}

	if !strings.Contains(fmt.Sprintf("%v", output), "offer") {
		t.Errorf("Unexpected buyer output: %v", output)
	}
}

// Note: FailingSpecialistAgent replaced with fixtures.ErrorSimulationMockAgent

// TestMultiAgentErrorHandling tests error handling in multi-agent systems
func TestMultiAgentErrorHandling(t *testing.T) {
	// Create workflow with failing agent
	workflow := workflow.NewSequentialAgent("workflow")

	successAgent := fixtures.SpecialistMockAgent("success_agent", "success", 0)
	failingAgent := fixtures.ErrorSimulationMockAgent("failing_agent", "always_fail", 1)

	workflow.AddAgent(successAgent)
	workflow.AddAgent(failingAgent)

	// Execute workflow
	ctx := context.Background()
	state := domain.NewState()
	state.Set("task", "This will partially succeed")

	result, err := workflow.Run(ctx, state)

	// Sequential workflow should fail if any agent fails
	if err == nil {
		t.Error("Expected error from failing agent")
	}

	if !strings.Contains(err.Error(), "simulated error") {
		t.Errorf("Unexpected error: %v", err)
	}

	// Result should be nil on error
	if result != nil {
		t.Error("Expected nil result on error")
	}
}

// TestMultiAgentScalability tests handling many agents
func TestMultiAgentScalability(t *testing.T) {
	// Create a coordinator using fixture
	coordinator := fixtures.CoordinatorMockAgent("main_coordinator")

	// Create many specialist agents
	numSpecialists := 20
	specialists := make([]domain.BaseAgent, numSpecialists)

	for i := 0; i < numSpecialists; i++ {
		specialist := fixtures.SpecialistMockAgent(
			fmt.Sprintf("specialist_%d", i),
			fmt.Sprintf("domain_%d", i),
			time.Duration(i)*time.Millisecond,
		)
		specialists[i] = specialist

		err := coordinator.AddSubAgent(specialist)
		if err != nil {
			t.Fatalf("Failed to add specialist %d: %v", i, err)
		}
	}

	// Verify all agents were added
	if len(coordinator.SubAgents()) != numSpecialists {
		t.Errorf("Expected %d sub-agents, got %d", numSpecialists, len(coordinator.SubAgents()))
	}

	// Create parallel workflow with all specialists
	parallelWorkflow := workflow.NewParallelAgent("parallel_specialists")
	parallelWorkflow.WithMaxConcurrency(10) // Limit concurrency

	for _, specialist := range specialists {
		parallelWorkflow.AddAgent(specialist)
	}

	// Execute parallel workflow
	ctx := context.Background()
	state := domain.NewState()
	state.Set("task", "Process large dataset")

	startTime := time.Now()
	result, err := parallelWorkflow.Run(ctx, state)
	duration := time.Since(startTime)

	if err != nil {
		t.Fatalf("Parallel workflow failed: %v", err)
	}

	// Verify result
	if result == nil {
		t.Fatal("Result is nil")
	}

	// With max concurrency of 10 and max delay of 19ms, should complete in ~40ms
	if duration > 100*time.Millisecond {
		t.Logf("Warning: Parallel execution took longer than expected: %v", duration)
	}
}

// Note: StateModifierAgent replaced with fixtures.StateBuilderMockAgent

// TestMultiAgentStateSharing tests shared state between agents
func TestMultiAgentStateSharing(t *testing.T) {

	// Create sequential workflow
	workflow := workflow.NewSequentialAgent("state_sharing")

	// Create agents that build shared data
	agent1 := fixtures.SharedDataBuilderMockAgent("agent1", "step1", "data1")
	agent2 := fixtures.SharedDataBuilderMockAgent("agent2", "step2", "data2")
	agent3 := fixtures.SharedDataBuilderMockAgent("agent3", "step3", "data3")

	workflow.AddAgent(agent1)
	workflow.AddAgent(agent2)
	workflow.AddAgent(agent3)

	// Execute workflow
	ctx := context.Background()
	state := domain.NewState()
	state.Set("task", "Build shared dataset")

	result, err := workflow.Run(ctx, state)
	if err != nil {
		t.Fatalf("Workflow failed: %v", err)
	}

	// Verify shared data contains all entries
	sharedData, exists := result.Get("shared_data")
	if !exists {
		t.Fatal("No shared_data in result")
	}

	dataMap, ok := sharedData.(map[string]interface{})
	if !ok {
		t.Fatalf("shared_data is not a map: %T", sharedData)
	}

	// Verify all agents contributed
	expectedKeys := []string{"step1", "step2", "step3"}
	for _, key := range expectedKeys {
		if _, exists := dataMap[key]; !exists {
			t.Errorf("Missing key %s in shared data", key)
		}
	}

	if len(dataMap) != 3 {
		t.Errorf("Expected 3 entries in shared data, got %d", len(dataMap))
	}
}

// TestGetSubAgentByName tests the convenience method for agent retrieval
func TestGetSubAgentByName(t *testing.T) {
	// Create mock provider
	mockProvider := provider.NewMockProvider()
	deps := core.LLMDeps{
		Provider: mockProvider,
	}

	// Create main agent
	mainAgent := core.NewLLMAgent("main", "Main agent", deps)

	// Create sub-agents using fixtures
	subAgent1 := fixtures.SpecialistMockAgent("sub1", "specialty1", 0)
	subAgent2 := fixtures.SpecialistMockAgent("sub2", "specialty2", 0)

	// Add sub-agents
	err := mainAgent.AddSubAgent(subAgent1)
	if err != nil {
		t.Fatalf("Failed to add subAgent1: %v", err)
	}
	err = mainAgent.AddSubAgent(subAgent2)
	if err != nil {
		t.Fatalf("Failed to add subAgent2: %v", err)
	}

	// Test GetSubAgentByName (alias for FindSubAgent)
	found1 := mainAgent.GetSubAgentByName("sub1")
	if found1 == nil {
		t.Error("Could not find sub1")
	}
	if found1.Name() != "sub1" {
		t.Errorf("Found wrong agent: %s", found1.Name())
	}

	found2 := mainAgent.GetSubAgentByName("sub2")
	if found2 == nil {
		t.Error("Could not find sub2")
	}

	// Test non-existent agent
	notFound := mainAgent.GetSubAgentByName("nonexistent")
	if notFound != nil {
		t.Error("Found non-existent agent")
	}
}
