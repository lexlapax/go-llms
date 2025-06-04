// ABOUTME: Example demonstrating LLM agents with structured output using schemas
// ABOUTME: Shows how to get structured responses from LLMs with validation and type safety

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	agentDomain "github.com/lexlapax/go-llms/pkg/agent/domain"
	llmDomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	"github.com/lexlapax/go-llms/pkg/schema/adapter/reflection"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

// Priority represents task priority levels
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

// Task represents a project task
type Task struct {
	ID             string     `json:"id" validate:"required" description:"Unique task identifier"`
	Title          string     `json:"title" validate:"required" description:"Task title"`
	Description    string     `json:"description" description:"Detailed task description"`
	Status         TaskStatus `json:"status" validate:"required,oneof=pending in_progress completed cancelled" description:"Current task status"`
	Priority       Priority   `json:"priority" validate:"required,oneof=low medium high urgent" description:"Task priority level"`
	DueDate        *time.Time `json:"due_date,omitempty" description:"Task due date (optional)"`
	EstimatedHours float64    `json:"estimated_hours" validate:"min=0" description:"Estimated hours to complete"`
	Tags           []string   `json:"tags" description:"Task tags for categorization"`
	CreatedAt      time.Time  `json:"created_at" validate:"required" description:"When the task was created"`
}

// ProjectAnalysis represents analysis of a project
type ProjectAnalysis struct {
	ProjectName       string   `json:"project_name" validate:"required" description:"Name of the project"`
	TotalTasks        int      `json:"total_tasks" validate:"min=0" description:"Total number of tasks"`
	CompletedTasks    int      `json:"completed_tasks" validate:"min=0" description:"Number of completed tasks"`
	PendingTasks      int      `json:"pending_tasks" validate:"min=0" description:"Number of pending tasks"`
	InProgressTasks   int      `json:"in_progress_tasks" validate:"min=0" description:"Number of tasks in progress"`
	HighPriorityTasks int      `json:"high_priority_tasks" validate:"min=0" description:"Number of high priority tasks"`
	OverdueTasks      int      `json:"overdue_tasks" validate:"min=0" description:"Number of overdue tasks"`
	EstimatedHours    float64  `json:"estimated_hours" validate:"min=0" description:"Total estimated hours"`
	CompletionRate    float64  `json:"completion_rate" validate:"min=0,max=100" description:"Completion percentage"`
	Recommendations   []string `json:"recommendations" description:"Recommendations for project improvement"`
	Risks             []string `json:"risks" description:"Identified project risks"`
	NextActions       []string `json:"next_actions" description:"Suggested next actions"`
}

// MeetingNotes represents structured meeting notes
type MeetingNotes struct {
	MeetingTitle   string       `json:"meeting_title" validate:"required" description:"Title of the meeting"`
	Date           time.Time    `json:"date" validate:"required" description:"Meeting date"`
	Attendees      []string     `json:"attendees" validate:"required,min=1" description:"List of meeting attendees"`
	Duration       int          `json:"duration" validate:"min=1" description:"Meeting duration in minutes"`
	KeyDiscussions []string     `json:"key_discussions" description:"Main discussion points"`
	Decisions      []string     `json:"decisions" description:"Decisions made during the meeting"`
	ActionItems    []ActionItem `json:"action_items" description:"Action items from the meeting"`
	NextMeeting    *time.Time   `json:"next_meeting,omitempty" description:"Next meeting date if scheduled"`
}

// ActionItem represents an action item from a meeting
type ActionItem struct {
	Description string     `json:"description" validate:"required" description:"Description of the action item"`
	AssignedTo  string     `json:"assigned_to" validate:"required" description:"Person assigned to the action item"`
	DueDate     *time.Time `json:"due_date,omitempty" description:"Due date for the action item"`
	Priority    Priority   `json:"priority" validate:"required,oneof=low medium high urgent" description:"Priority level"`
	Status      TaskStatus `json:"status" validate:"required,oneof=pending in_progress completed cancelled" description:"Current status"`
}

func main() {
	// Check for API key and create provider
	var llmProvider llmDomain.Provider
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		llmProvider = provider.NewOpenAIProvider(apiKey, "gpt-4o")
		fmt.Println("Using OpenAI provider with GPT-4o")
	} else if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		llmProvider = provider.NewAnthropicProvider(apiKey, "claude-3-5-sonnet-latest")
		fmt.Println("Using Anthropic provider with Claude")
	} else if apiKey := os.Getenv("GEMINI_API_KEY"); apiKey != "" {
		llmProvider = provider.NewGeminiProvider(apiKey, "gemini-2.0-flash")
		fmt.Println("Using Gemini provider")
	} else {
		llmProvider = provider.NewMockProvider()
		fmt.Println("No API keys found. Using mock provider for demonstration.")
	}

	// Create agent
	agent, err := core.NewAgentFromString("structured-agent", "openai/gpt-4o")
	if err != nil {
		// Fallback to mock if no API key
		agent = core.NewAgent("structured-agent", llmProvider)
	}

	// Set system prompt for structured output
	agent.SetSystemPrompt(`You are a helpful assistant that provides structured responses. 
When asked to analyze data or create structured content, respond with valid JSON that matches the requested schema.
Be accurate, thorough, and follow the exact structure provided.
Use realistic data and provide meaningful insights in your responses.`)

	ctx := context.Background()

	// Example 1: Generate structured task data
	fmt.Println("=== Example 1: Structured Task Generation ===")

	taskSchema, err := reflection.GenerateSchema(Task{})
	if err != nil {
		log.Fatalf("Failed to generate task schema: %v", err)
	}

	// Generate a task using structured output
	taskSchemaJSON, _ := json.MarshalIndent(taskSchema, "", "  ")
	prompt := fmt.Sprintf(`Create a realistic software development task for implementing a user authentication system. 
The task should be high priority and include estimated hours and relevant tags.

Please respond with a JSON object that matches this schema:
%s

Respond only with valid JSON, no additional text.`, taskSchemaJSON)

	demonstrateTaskGeneration(agent, ctx, prompt)

	// Example 2: Project Analysis with Complex Schema
	fmt.Println("\n=== Example 2: Project Analysis ===")

	analysisSchema, err := reflection.GenerateSchema(ProjectAnalysis{})
	if err != nil {
		log.Fatalf("Failed to generate analysis schema: %v", err)
	}

	projectData := `
	Project: E-commerce Platform Redesign
	Tasks:
	1. Update homepage design (completed, 8 hours)
	2. Implement shopping cart (in_progress, 12 hours)
	3. Add payment integration (pending, 16 hours)
	4. Setup user accounts (completed, 6 hours)
	5. Mobile optimization (pending, 20 hours)
	6. Security audit (urgent, pending, 8 hours)
	7. Performance testing (high priority, pending, 4 hours)
	8. Database migration (in_progress, 10 hours)
	`

	analysisSchemaJSON, _ := json.MarshalIndent(analysisSchema, "", "  ")
	analysisPrompt := fmt.Sprintf(`Analyze this project data and provide a comprehensive analysis:
%s

Calculate completion rates, identify risks, and provide actionable recommendations.

Please respond with a JSON object that matches this schema:
%s

Respond only with valid JSON, no additional text.`, projectData, analysisSchemaJSON)

	demonstrateProjectAnalysis(agent, ctx, analysisPrompt)

	// Example 3: Meeting Notes Structure
	fmt.Println("\n=== Example 3: Structured Meeting Notes ===")

	meetingSchema, err := reflection.GenerateSchema(MeetingNotes{})
	if err != nil {
		log.Fatalf("Failed to generate meeting notes schema: %v", err)
	}

	meetingTranscript := `
	Meeting: Sprint Planning Session
	Attendees: Sarah (Product Manager), Mike (Tech Lead), Jenny (Developer), Alex (Designer)
	Duration: 90 minutes
	Date: Today
	
	Discussion:
	- Reviewed last sprint performance (85% completion rate)
	- Prioritized user authentication features for next sprint
	- Discussed technical debt in payment module
	- Alex presented new UI mockups for dashboard
	- Mike raised concerns about database performance
	
	Decisions:
	- Authentication system is top priority for next sprint
	- Will allocate 20% of sprint capacity to technical debt
	- UI redesign approved for implementation
	- Database optimization to be scheduled for following sprint
	
	Action Items:
	- Sarah: Update user stories for authentication (by end of week)
	- Mike: Create technical architecture document (by Tuesday)
	- Jenny: Set up development environment for new features (by Thursday)
	- Alex: Finalize UI components library (by Monday)
	- Team: Schedule database performance review meeting
	`

	meetingSchemaJSON, _ := json.MarshalIndent(meetingSchema, "", "  ")
	meetingPrompt := fmt.Sprintf(`Convert this meeting transcript into structured meeting notes:
%s

Extract key discussions, decisions, and create proper action items with assignments.

Please respond with a JSON object that matches this schema:
%s

Respond only with valid JSON, no additional text.`, meetingTranscript, meetingSchemaJSON)

	demonstrateMeetingNotes(agent, ctx, meetingPrompt)

	// Example 4: Schema Validation Demo
	fmt.Println("\n=== Example 4: Schema Validation ===")
	demonstrateSchemaValidation()

	// Example 5: Advanced Schema Features
	fmt.Println("\n=== Example 5: Advanced Schema Features ===")
	demonstrateAdvancedSchemas()
}

func demonstrateTaskGeneration(agent *core.LLMAgent, ctx context.Context, prompt string) {
	fmt.Println("Generating structured task...")

	state := agentDomain.NewState()
	state.Set("prompt", prompt)

	result, err := agent.Run(ctx, state)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	if response, exists := result.Get("result"); exists {
		fmt.Println("Generated Task:")
		fmt.Println(response)
	}
}

func demonstrateProjectAnalysis(agent *core.LLMAgent, ctx context.Context, prompt string) {
	fmt.Println("Generating structured project analysis...")

	state := agentDomain.NewState()
	state.Set("prompt", prompt)

	result, err := agent.Run(ctx, state)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	if response, exists := result.Get("result"); exists {
		fmt.Println("Project Analysis:")
		fmt.Println(response)
	}
}

func demonstrateMeetingNotes(agent *core.LLMAgent, ctx context.Context, prompt string) {
	fmt.Println("Generating structured meeting notes...")

	state := agentDomain.NewState()
	state.Set("prompt", prompt)

	result, err := agent.Run(ctx, state)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	if response, exists := result.Get("result"); exists {
		fmt.Println("Meeting Notes:")
		fmt.Println(response)
	}
}

func demonstrateSchemaValidation() {
	fmt.Println("Demonstrating schema validation...")

	// Create a task schema
	taskSchema, err := reflection.GenerateSchema(Task{})
	if err != nil {
		log.Printf("Error generating schema: %v", err)
		return
	}

	// Valid task data
	validTask := Task{
		ID:             "task-001",
		Title:          "Implement OAuth",
		Description:    "Add OAuth authentication to the application",
		Status:         TaskStatusPending,
		Priority:       PriorityHigh,
		EstimatedHours: 8.5,
		Tags:           []string{"authentication", "security", "backend"},
		CreatedAt:      time.Now(),
	}

	taskJSON, _ := json.MarshalIndent(validTask, "", "  ")
	fmt.Printf("Valid Task Schema Example:\n%s\n", taskJSON)

	// Display the generated schema
	schemaJSON, _ := json.MarshalIndent(taskSchema, "", "  ")
	fmt.Printf("Generated Schema:\n%s\n", schemaJSON)
}

func demonstrateAdvancedSchemas() {
	fmt.Println("Demonstrating advanced schema features...")

	// Show nested objects
	meetingSchema, err := reflection.GenerateSchema(MeetingNotes{})
	if err != nil {
		log.Printf("Error generating meeting schema: %v", err)
		return
	}

	fmt.Println("Advanced Features Demonstrated:")
	fmt.Println("✓ Nested objects (ActionItem within MeetingNotes)")
	fmt.Println("✓ Arrays of complex objects")
	fmt.Println("✓ Optional fields with pointers")
	fmt.Println("✓ Enum validation with oneof constraints")
	fmt.Println("✓ Time.Time handling")
	fmt.Println("✓ Validation constraints (min, max, required)")
	fmt.Println("✓ Pattern validation for strings")

	// Show partial schema for ActionItem
	actionItemSchema, _ := reflection.GenerateSchema(ActionItem{})
	actionItemJSON, _ := json.MarshalIndent(actionItemSchema, "", "  ")
	fmt.Printf("\nActionItem Schema (nested object):\n%s\n", actionItemJSON)

	// Show that meetingSchema contains the action items
	fmt.Printf("\nMeeting schema has %d properties\n", len(meetingSchema.Properties))

	// Example of advanced schema usage
	fmt.Println("\nKey Benefits:")
	fmt.Println("• Type-safe LLM interactions")
	fmt.Println("• Automatic validation of LLM responses")
	fmt.Println("• Clear structure definition for complex data")
	fmt.Println("• Integration with Go's type system")
	fmt.Println("• Reduced parsing errors and data inconsistencies")
}
