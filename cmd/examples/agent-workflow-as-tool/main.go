// ABOUTME: Example demonstrating a multi-stage research pipeline with workflow agents wrapped as tools
// ABOUTME: Shows how a main LLM agent orchestrates complex research using sequential and parallel workflow agents

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	agenttools "github.com/lexlapax/go-llms/pkg/agent/tools"
	"github.com/lexlapax/go-llms/pkg/agent/workflow"
	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/util/llmutil"
)

func main() {
	ctx := context.Background()

	// Get provider from environment
	llmProvider, _, _, err := llmutil.ProviderFromEnv()
	if err != nil {
		log.Fatalf("Failed to get LLM provider: %v", err)
	}

	// Create specialized agents for analysis pipeline
	contentAnalyzer := createContentAnalyzer(llmProvider)
	factChecker := createFactChecker(llmProvider)
	summaryGenerator := createSummaryGenerator(llmProvider)

	// Create Sequential Workflow (Analysis Pipeline)
	analysisPipeline := workflow.NewSequentialAgent("analysis-pipeline").
		AddAgent(contentAnalyzer).
		AddAgent(factChecker).
		AddAgent(summaryGenerator)

	// Create specialized agents for comparison
	sourceAnalyzerA := createSourceAnalyzer("A", llmProvider)
	sourceAnalyzerB := createSourceAnalyzer("B", llmProvider)

	// Create Parallel Workflow (Comparison)
	comparisonAgent := workflow.NewParallelAgent("comparison-agent").
		AddAgent(sourceAnalyzerA).
		AddAgent(sourceAnalyzerB).
		WithMergeFunc(createComparisonMergeFunc()).
		WithMaxConcurrency(2)

	// Wrap workflow agents as tools
	analysisTool := agenttools.NewAgentTool(analysisPipeline)
	comparisonTool := agenttools.NewAgentTool(comparisonAgent)

	// Create main Research Coordinator
	coordinator := core.NewAgent("research-coordinator", llmProvider)

	// Add tools to coordinator
	coordinator.AddTool(tools.MustGetTool("web_search"))
	coordinator.AddTool(tools.MustGetTool("web_fetch"))
	coordinator.AddTool(analysisTool)
	coordinator.AddTool(comparisonTool)
	coordinator.AddTool(tools.MustGetTool("file_write"))

	// Set system prompt
	coordinator.SetSystemPrompt(`You are a research coordinator with advanced tools for comprehensive research and analysis.

Your tools include:
1. web_search: Search the web for information
2. web_fetch: Fetch content from specific URLs
3. analysis-pipeline: A sophisticated analysis pipeline that:
   - Analyzes content to extract key points, entities, and topics
   - Fact-checks claims against reliable sources
   - Generates a verified summary
4. comparison-agent: Compares two sources simultaneously to identify similarities and differences
5. file_write: Save your findings to a file

For research tasks:
1. First search for relevant sources using web_search
2. Fetch content from the most promising sources using web_fetch
3. If comparing multiple perspectives, use the comparison-agent with two sources
4. Run the analysis-pipeline on the content or comparison results
5. Save the final report using file_write

Always provide thorough, fact-checked research with clear citations.`)

	// Add logging hook to see what's happening
	coordinator.WithHook(&core.LoggingHook{})

	// Execute research task
	userQuery := "Research and compare perspectives on AI safety from two different authoritative sources"
	if len(os.Args) > 1 {
		userQuery = strings.Join(os.Args[1:], " ")
	}

	fmt.Printf("Research Query: %s\n", userQuery)
	fmt.Println("Starting research pipeline...")
	fmt.Println()

	// Create initial state
	initialState := domain.NewState()
	initialState.AddMessage(domain.Message{
		Role:    domain.RoleUser,
		Content: userQuery,
	})

	// Run the coordinator
	result, err := coordinator.Run(ctx, initialState)
	if err != nil {
		log.Fatalf("Research failed: %v", err)
	}

	// Display results
	fmt.Println("\n=== Research Complete ===")
	messages := result.Messages()
	if len(messages) > 0 {
		lastMessage := messages[len(messages)-1]
		fmt.Printf("Final Report:\n%s\n", lastMessage.Content)
	}

	// Check if report was saved
	if artifacts, ok := result.GetMetadata("artifacts"); ok && artifacts != nil {
		fmt.Println("\n=== Report saved to file ===")
	}
}

// createContentAnalyzer creates an LLM agent that analyzes content
func createContentAnalyzer(llmProvider ldomain.Provider) domain.BaseAgent {
	agent := core.NewAgent("content-analyzer", llmProvider)

	agent.SetSystemPrompt(`You are a content analyzer. Extract and structure the following from the provided text:
1. Key Points: Main arguments or claims
2. Entities: Important people, organizations, or concepts mentioned
3. Topics: Primary subjects discussed
4. Tone: Overall tone and perspective

Format your response as a JSON object with these fields.`)

	return agent
}

// createFactChecker creates an LLM agent that fact-checks claims
func createFactChecker(llmProvider ldomain.Provider) domain.BaseAgent {
	agent := core.NewAgent("fact-checker", llmProvider)

	agent.SetSystemPrompt(`You are a fact checker. Review the analysis provided and:
1. Identify specific claims that can be verified
2. Note which claims are opinions vs facts
3. Flag any potentially misleading or unsupported statements
4. Rate the overall credibility (1-10)

Add your fact-checking results to the existing analysis.`)

	return agent
}

// createSummaryGenerator creates an LLM agent that generates summaries
func createSummaryGenerator(llmProvider ldomain.Provider) domain.BaseAgent {
	agent := core.NewAgent("summary-generator", llmProvider)

	agent.SetSystemPrompt(`You are a summary generator. Based on the analyzed and fact-checked content:
1. Create a concise executive summary (2-3 paragraphs)
2. List the most important verified facts
3. Note any areas of uncertainty or conflicting information
4. Provide an overall assessment

Format as a clear, readable summary suitable for decision-makers.`)

	return agent
}

// createSourceAnalyzer creates an LLM agent that analyzes a specific source
func createSourceAnalyzer(label string, llmProvider ldomain.Provider) domain.BaseAgent {
	agent := core.NewAgent(fmt.Sprintf("source-analyzer-%s", label), llmProvider)

	agent.SetSystemPrompt(fmt.Sprintf(`You are Source Analyzer %s. Analyze the provided content and extract:
1. Main thesis or position
2. Supporting arguments
3. Evidence cited
4. Potential biases or limitations
5. Unique perspectives offered

Format your analysis as structured JSON for easy comparison.`, label))

	return agent
}

// createComparisonMergeFunc creates a custom merge function for parallel comparison
func createComparisonMergeFunc() func(results map[string]*domain.State) *domain.State {
	return func(results map[string]*domain.State) *domain.State {
		merged := domain.NewState()

		// Extract analyzes from both sources
		var analysisA, analysisB map[string]interface{}
		i := 0
		for _, result := range results {
			messages := result.Messages()
			if len(messages) > 0 {
				lastMsg := messages[len(messages)-1]
				var analysis map[string]interface{}
				if err := json.Unmarshal([]byte(lastMsg.Content), &analysis); err == nil {
					if i == 0 {
						analysisA = analysis
					} else {
						analysisB = analysis
					}
				}
			}
			i++
		}

		// Create comparison summary
		comparison := map[string]interface{}{
			"source_a_analysis": analysisA,
			"source_b_analysis": analysisB,
			"comparison": map[string]interface{}{
				"similarities": findSimilarities(analysisA, analysisB),
				"differences":  findDifferences(analysisA, analysisB),
				"timestamp":    time.Now().Format(time.RFC3339),
			},
		}

		comparisonJSON, _ := json.MarshalIndent(comparison, "", "  ")

		merged.AddMessage(domain.Message{
			Role:    domain.RoleAssistant,
			Content: string(comparisonJSON),
		})

		merged.SetMetadata("comparison_complete", true)
		merged.SetMetadata("sources_analyzed", 2)

		return merged
	}
}

// findSimilarities identifies common elements between analyzes
func findSimilarities(a, b map[string]interface{}) []string {
	similarities := []string{}

	// This is a simplified comparison - in production, you'd want more sophisticated analysis
	if a != nil && b != nil {
		// Check if both mention similar topics
		if topicsA, okA := a["topics"]; okA && topicsA != nil {
			if topicsB, okB := b["topics"]; okB && topicsB != nil {
				similarities = append(similarities, "Both sources discuss similar topics")
			}
		}

		// Check if both have similar tone
		if toneA, okA := a["tone"]; okA {
			if toneB, okB := b["tone"]; okB {
				if toneA == toneB {
					similarities = append(similarities, fmt.Sprintf("Both sources have a %v tone", toneA))
				}
			}
		}
	}

	if len(similarities) == 0 {
		similarities = append(similarities, "Limited similarities found between sources")
	}

	return similarities
}

// findDifferences identifies differing elements between analyzes
func findDifferences(a, b map[string]interface{}) []string {
	differences := []string{}

	// This is a simplified comparison - in production, you'd want more sophisticated analysis
	if a != nil && b != nil {
		// Check for different main thesis
		if thesisA, okA := a["main_thesis"]; okA {
			if thesisB, okB := b["main_thesis"]; okB {
				if thesisA != thesisB {
					differences = append(differences, "Sources present different main arguments")
				}
			}
		}

		// Check for different perspectives
		differences = append(differences, "Each source offers unique perspectives and evidence")
	}

	if len(differences) == 0 {
		differences = append(differences, "No significant differences identified")
	}

	return differences
}
