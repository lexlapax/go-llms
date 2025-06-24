# Business Automation: Process Automation

> **[Project Root](/) / [Documentation](/docs/) / [User Guide](/docs/user-guide/) / [Examples](/docs/user-guide/examples/) / Business Automation**

Build comprehensive business process automation systems using Go-LLMs. These examples demonstrate how to automate complex business workflows, integrate with existing systems, and create intelligent automation that adapts to business needs.

## Overview

Business automation with Go-LLMs enables:
- **Intelligent Process Automation** - Beyond simple rule-based automation
- **Natural Language Interfaces** - Business users can interact naturally
- **Adaptive Workflows** - Systems that learn and improve over time
- **Seamless Integration** - Connect with existing business systems
- **Compliance and Governance** - Built-in audit trails and controls

---

## Invoice Processing Automation

Automate the entire invoice processing workflow from receipt to payment, including validation, approval routing, and exception handling.

### Features
- Multi-format invoice extraction (PDF, images, emails)
- Intelligent data validation and error correction
- Dynamic approval routing based on business rules
- Vendor verification and fraud detection
- Integration with accounting systems

### Implementation

```go
package main

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq"
    
    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/agent/workflow"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
    "github.com/lexlapax/go-llms/pkg/structured/processor"
)

type InvoiceProcessingSystem struct {
    db                  *sqlx.DB
    extractionAgent     *core.LLMAgent
    validationAgent     *core.LLMAgent
    approvalAgent       *core.LLMAgent
    fraudDetectionAgent *core.LLMAgent
    workflowEngine      *workflow.ConditionalAgent
    accountingIntegration *AccountingIntegration
    config             *ProcessingConfig
}

type Invoice struct {
    ID              string          `json:"id" db:"id"`
    VendorName      string          `json:"vendor_name" db:"vendor_name"`
    VendorID        string          `json:"vendor_id" db:"vendor_id"`
    InvoiceNumber   string          `json:"invoice_number" db:"invoice_number"`
    InvoiceDate     time.Time       `json:"invoice_date" db:"invoice_date"`
    DueDate         time.Time       `json:"due_date" db:"due_date"`
    TotalAmount     float64         `json:"total_amount" db:"total_amount"`
    Currency        string          `json:"currency" db:"currency"`
    LineItems       []LineItem      `json:"line_items"`
    Status          InvoiceStatus   `json:"status" db:"status"`
    ApprovalChain   []ApprovalStep  `json:"approval_chain"`
    ValidationFlags []ValidationFlag `json:"validation_flags"`
    ProcessingTime  time.Duration   `json:"processing_time"`
    CreatedAt       time.Time       `json:"created_at" db:"created_at"`
}

type InvoiceStatus string

const (
    StatusPending     InvoiceStatus = "pending"
    StatusValidating  InvoiceStatus = "validating"
    StatusApproving   InvoiceStatus = "approving"
    StatusApproved    InvoiceStatus = "approved"
    StatusRejected    InvoiceStatus = "rejected"
    StatusProcessed   InvoiceStatus = "processed"
)

func NewInvoiceProcessingSystem(config *ProcessingConfig) (*InvoiceProcessingSystem, error) {
    // Initialize database
    db, err := sqlx.Connect("postgres", config.DatabaseURL)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }

    // Initialize LLM provider
    openaiProvider, err := provider.NewOpenAI(provider.OpenAIOptions{
        APIKey: config.OpenAIKey,
        Model:  "gpt-4",
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create OpenAI provider: %w", err)
    }

    // Create specialized agents
    extractionAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "invoice-extractor",
        SystemPrompt: invoiceExtractionPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create extraction agent: %w", err)
    }

    validationAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "invoice-validator",
        SystemPrompt: invoiceValidationPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create validation agent: %w", err)
    }

    approvalAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "approval-router",
        SystemPrompt: approvalRoutingPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create approval agent: %w", err)
    }

    fraudDetectionAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "fraud-detector",
        SystemPrompt: fraudDetectionPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create fraud detection agent: %w", err)
    }

    // Create conditional workflow
    workflowEngine := workflow.NewConditionalAgent(workflow.ConditionalAgentOptions{
        Name: "invoice-processing-workflow",
        Conditions: []workflow.Condition{
            {
                Name:      "extraction-needed",
                Predicate: func(ctx *domain.AgentContext) bool {
                    return ctx.GetInputProperty("needs_extraction").(bool)
                },
                Agent: extractionAgent,
            },
            {
                Name:      "validation-needed",
                Predicate: func(ctx *domain.AgentContext) bool {
                    return ctx.GetInputProperty("needs_validation").(bool)
                },
                Agent: validationAgent,
            },
            {
                Name:      "approval-needed",
                Predicate: func(ctx *domain.AgentContext) bool {
                    return ctx.GetInputProperty("needs_approval").(bool)
                },
                Agent: approvalAgent,
            },
        },
    })

    return &InvoiceProcessingSystem{
        db:                   db,
        extractionAgent:      extractionAgent,
        validationAgent:      validationAgent,
        approvalAgent:        approvalAgent,
        fraudDetectionAgent:  fraudDetectionAgent,
        workflowEngine:       workflowEngine,
        accountingIntegration: NewAccountingIntegration(config.AccountingSystem),
        config:              config,
    }, nil
}

func (ips *InvoiceProcessingSystem) ProcessInvoice(ctx context.Context, document []byte, format string) (*Invoice, error) {
    startTime := time.Now()

    // Step 1: Extract invoice data
    extractionCtx := domain.NewAgentContext().
        SetInputProperty("document", document).
        SetInputProperty("format", format).
        SetInputProperty("needs_extraction", true)

    extractionResult, err := ips.workflowEngine.Execute(ctx, extractionCtx)
    if err != nil {
        return nil, fmt.Errorf("extraction failed: %w", err)
    }

    // Parse extracted data
    var invoice Invoice
    if invoiceData := extractionResult.GetOutputProperty("invoice_data"); invoiceData != nil {
        invoiceJSON, _ := json.Marshal(invoiceData)
        json.Unmarshal(invoiceJSON, &invoice)
    }

    // Step 2: Validate invoice
    validationCtx := domain.NewAgentContext().
        SetInputProperty("invoice", invoice).
        SetInputProperty("vendor_history", ips.getVendorHistory(invoice.VendorID)).
        SetInputProperty("needs_validation", true)

    validationResult, err := ips.workflowEngine.Execute(ctx, validationCtx)
    if err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }

    // Step 3: Fraud detection
    fraudCtx := domain.NewAgentContext().
        SetInputProperty("invoice", invoice).
        SetInputProperty("vendor_profile", ips.getVendorProfile(invoice.VendorID)).
        SetInputProperty("historical_patterns", ips.getHistoricalPatterns())

    fraudResult, err := ips.fraudDetectionAgent.Execute(ctx, fraudCtx)
    if err != nil {
        return nil, fmt.Errorf("fraud detection failed: %w", err)
    }

    if fraudRisk := fraudResult.GetOutputProperty("fraud_risk"); fraudRisk.(float64) > 0.7 {
        invoice.Status = StatusRejected
        invoice.ValidationFlags = append(invoice.ValidationFlags, ValidationFlag{
            Type:     "fraud_risk",
            Severity: "high",
            Message:  fraudResult.GetOutputProperty("fraud_reason").(string),
        })
        return &invoice, nil
    }

    // Step 4: Determine approval routing
    approvalCtx := domain.NewAgentContext().
        SetInputProperty("invoice", invoice).
        SetInputProperty("business_rules", ips.config.ApprovalRules).
        SetInputProperty("needs_approval", true)

    approvalResult, err := ips.workflowEngine.Execute(ctx, approvalCtx)
    if err != nil {
        return nil, fmt.Errorf("approval routing failed: %w", err)
    }

    // Set approval chain
    if approvalChain := approvalResult.GetOutputProperty("approval_chain"); approvalChain != nil {
        chainJSON, _ := json.Marshal(approvalChain)
        json.Unmarshal(chainJSON, &invoice.ApprovalChain)
    }

    invoice.Status = StatusApproving
    invoice.ProcessingTime = time.Since(startTime)

    // Save to database
    if err := ips.saveInvoice(&invoice); err != nil {
        return nil, fmt.Errorf("failed to save invoice: %w", err)
    }

    // Trigger approval workflow
    go ips.startApprovalWorkflow(ctx, &invoice)

    return &invoice, nil
}

func (ips *InvoiceProcessingSystem) startApprovalWorkflow(ctx context.Context, invoice *Invoice) {
    for _, step := range invoice.ApprovalChain {
        // Send approval request
        if err := ips.sendApprovalRequest(step.ApproverID, invoice); err != nil {
            log.Printf("Failed to send approval request: %v", err)
            continue
        }

        // Wait for approval or timeout
        approved, err := ips.waitForApproval(ctx, invoice.ID, step.ApproverID, step.Timeout)
        if err != nil || !approved {
            invoice.Status = StatusRejected
            ips.updateInvoiceStatus(invoice)
            return
        }
    }

    // All approvals received
    invoice.Status = StatusApproved
    ips.updateInvoiceStatus(invoice)

    // Process payment
    if err := ips.processPayment(ctx, invoice); err != nil {
        log.Printf("Payment processing failed: %v", err)
    }
}

const invoiceExtractionPrompt = `You are an invoice data extraction specialist. Extract all relevant information from invoices including:
- Vendor information (name, ID, address, contact)
- Invoice details (number, date, due date)
- Line items with descriptions, quantities, and amounts
- Total amount and currency
- Payment terms and conditions
- Tax information

Handle various invoice formats and layouts. Return structured JSON data.`

const invoiceValidationPrompt = `You are an invoice validation expert. Validate invoices by:
- Checking mathematical accuracy of totals
- Verifying vendor information against database
- Validating tax calculations
- Checking for duplicate invoices
- Ensuring compliance with business rules
- Identifying missing or suspicious information

Return validation results with specific issues and recommendations.`

const approvalRoutingPrompt = `You are an approval routing specialist. Determine approval chains based on:
- Invoice amount and thresholds
- Vendor category and risk level
- Department and cost center rules
- Special approval requirements
- Delegation of authority matrix

Create optimal approval chains that balance control with efficiency.`

const fraudDetectionPrompt = `You are a fraud detection specialist. Analyze invoices for:
- Unusual patterns or anomalies
- Vendor legitimacy indicators
- Invoice authenticity markers
- Historical fraud patterns
- Risk indicators and red flags

Provide risk assessment with specific concerns and confidence levels.`

// API endpoints
func (ips *InvoiceProcessingSystem) SetupAPI() *gin.Engine {
    r := gin.Default()

    r.POST("/api/invoices/upload", ips.UploadInvoice)
    r.GET("/api/invoices/:id", ips.GetInvoice)
    r.GET("/api/invoices", ips.ListInvoices)
    r.POST("/api/invoices/:id/approve", ips.ApproveInvoice)
    r.POST("/api/invoices/:id/reject", ips.RejectInvoice)
    r.GET("/api/dashboard", ips.GetDashboard)

    return r
}

func main() {
    config := &ProcessingConfig{
        DatabaseURL: "postgres://user:pass@localhost/invoices",
        OpenAIKey:   "your-openai-key",
        ApprovalRules: ApprovalRules{
            Thresholds: []ThresholdRule{
                {MaxAmount: 1000, Approvers: []string{"manager"}},
                {MaxAmount: 10000, Approvers: []string{"manager", "director"}},
                {MaxAmount: 100000, Approvers: []string{"manager", "director", "vp"}},
            },
        },
        AccountingSystem: "quickbooks",
    }

    system, err := NewInvoiceProcessingSystem(config)
    if err != nil {
        log.Fatalf("Failed to initialize system: %v", err)
    }

    r := system.SetupAPI()
    fmt.Println("Invoice Processing System running on :8080")
    log.Fatal(r.Run(":8080"))
}
```

---

## HR Onboarding Automation

Streamline the employee onboarding process from offer acceptance to first day readiness.

### Features
- Document collection and verification
- System access provisioning
- Training schedule creation
- Equipment requisition
- Compliance tracking

### Implementation

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/agent/workflow"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
)

type HROnboardingSystem struct {
    documentAgent      *core.LLMAgent
    provisioningAgent  *core.LLMAgent
    trainingAgent      *core.LLMAgent
    complianceAgent    *core.LLMAgent
    coordinatorAgent   *core.LLMAgent
    workflowOrchestrator *workflow.SequentialAgent
    systems           map[string]SystemIntegration
    config            *OnboardingConfig
}

type NewEmployee struct {
    ID              string           `json:"id"`
    Name            string           `json:"name"`
    Email           string           `json:"email"`
    Department      string           `json:"department"`
    Role            string           `json:"role"`
    Manager         string           `json:"manager"`
    StartDate       time.Time        `json:"start_date"`
    Location        string           `json:"location"`
    EmploymentType  string           `json:"employment_type"`
    OnboardingSteps []OnboardingStep `json:"onboarding_steps"`
}

type OnboardingStep struct {
    ID              string        `json:"id"`
    Name            string        `json:"name"`
    Category        string        `json:"category"`
    Description     string        `json:"description"`
    Assignee        string        `json:"assignee"`
    DueDate         time.Time     `json:"due_date"`
    Status          StepStatus    `json:"status"`
    CompletedAt     *time.Time    `json:"completed_at,omitempty"`
    Documents       []Document    `json:"documents,omitempty"`
    SystemAccess    []SystemAccess `json:"system_access,omitempty"`
    Dependencies    []string      `json:"dependencies"`
}

func NewHROnboardingSystem(config *OnboardingConfig) (*HROnboardingSystem, error) {
    openaiProvider, err := provider.NewOpenAI(provider.OpenAIOptions{
        APIKey: config.OpenAIKey,
        Model:  "gpt-4",
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create OpenAI provider: %w", err)
    }

    // Create specialized onboarding agents
    documentAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "document-processor",
        SystemPrompt: documentProcessingPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create document agent: %w", err)
    }

    provisioningAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "system-provisioner",
        SystemPrompt: systemProvisioningPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create provisioning agent: %w", err)
    }

    trainingAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "training-coordinator",
        SystemPrompt: trainingCoordinationPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create training agent: %w", err)
    }

    complianceAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "compliance-checker",
        SystemPrompt: complianceCheckingPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create compliance agent: %w", err)
    }

    coordinatorAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "onboarding-coordinator",
        SystemPrompt: onboardingCoordinationPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create coordinator agent: %w", err)
    }

    // Create onboarding workflow
    workflowOrchestrator := workflow.NewSequentialAgent(workflow.SequentialAgentOptions{
        Name: "onboarding-workflow",
        Steps: []domain.Agent{
            documentAgent,
            provisioningAgent,
            trainingAgent,
            complianceAgent,
            coordinatorAgent,
        },
    })

    return &HROnboardingSystem{
        documentAgent:        documentAgent,
        provisioningAgent:    provisioningAgent,
        trainingAgent:        trainingAgent,
        complianceAgent:      complianceAgent,
        coordinatorAgent:     coordinatorAgent,
        workflowOrchestrator: workflowOrchestrator,
        systems:             initializeSystemIntegrations(config),
        config:              config,
    }, nil
}

func (hos *HROnboardingSystem) StartOnboarding(ctx context.Context, employee *NewEmployee) (*OnboardingPlan, error) {
    // Create personalized onboarding plan
    planCtx := domain.NewAgentContext().
        SetInputProperty("employee", employee).
        SetInputProperty("role_requirements", hos.getRoleRequirements(employee.Role)).
        SetInputProperty("compliance_requirements", hos.getComplianceRequirements(employee)).
        SetInputProperty("system_access_matrix", hos.getSystemAccessMatrix())

    result, err := hos.workflowOrchestrator.Execute(ctx, planCtx)
    if err != nil {
        return nil, fmt.Errorf("failed to create onboarding plan: %w", err)
    }

    var plan OnboardingPlan
    if planData := result.GetOutputProperty("onboarding_plan"); planData != nil {
        // Parse plan data
    }

    // Start automated onboarding tasks
    go hos.executeOnboardingPlan(ctx, &plan, employee)

    return &plan, nil
}

func (hos *HROnboardingSystem) executeOnboardingPlan(ctx context.Context, plan *OnboardingPlan, employee *NewEmployee) {
    for _, phase := range plan.Phases {
        log.Printf("Starting onboarding phase: %s", phase.Name)

        switch phase.Type {
        case PhaseTypeDocuments:
            hos.processDocuments(ctx, phase, employee)
        case PhaseTypeProvisioning:
            hos.provisionAccess(ctx, phase, employee)
        case PhaseTypeTraining:
            hos.scheduleTraining(ctx, phase, employee)
        case PhaseTypeEquipment:
            hos.requestEquipment(ctx, phase, employee)
        }

        // Update progress
        hos.updateOnboardingProgress(employee.ID, phase.ID, PhaseCompleted)
    }

    // Final compliance check
    if err := hos.performFinalComplianceCheck(ctx, employee); err != nil {
        log.Printf("Compliance check failed: %v", err)
    }

    // Send completion notification
    hos.notifyOnboardingComplete(employee)
}

const documentProcessingPrompt = `You are a document processing specialist for HR onboarding. Your responsibilities:
- Identify required documents based on role and location
- Verify document completeness and validity
- Extract relevant information from documents
- Flag missing or incorrect documentation
- Ensure compliance with legal requirements

Create comprehensive document checklists and validation rules.`

const systemProvisioningPrompt = `You are a system access provisioning specialist. Your tasks:
- Determine required system access based on role
- Create provisioning requests for each system
- Set appropriate permission levels
- Schedule access reviews
- Document access grants for audit purposes

Ensure least-privilege access while enabling productivity.`

const trainingCoordinationPrompt = `You are a training coordination specialist. Your responsibilities:
- Identify required training based on role and compliance
- Schedule training sessions optimally
- Create personalized learning paths
- Track training completion
- Coordinate with trainers and departments

Design effective onboarding training that accelerates time-to-productivity.`

const complianceCheckingPrompt = `You are a compliance checking specialist for onboarding. Ensure:
- All legal requirements are met
- Required certifications are obtained
- Background checks are completed
- Policy acknowledgments are collected
- Regulatory requirements are satisfied

Maintain detailed compliance records and flag any issues.`

const onboardingCoordinationPrompt = `You are an onboarding coordination specialist. Your role:
- Create comprehensive onboarding plans
- Coordinate between departments and stakeholders
- Track progress and resolve bottlenecks
- Communicate with new employees and managers
- Ensure smooth first-day experience

Optimize the onboarding experience for engagement and efficiency.`

func main() {
    config := &OnboardingConfig{
        OpenAIKey: "your-openai-key",
        Systems: SystemsConfig{
            HRIS:     "workday",
            Identity: "okta",
            Email:    "google",
            Slack:    true,
        },
        ComplianceRules: []ComplianceRule{
            {Type: "i9_verification", Required: true, Deadline: 3},
            {Type: "background_check", Required: true, Deadline: -7}, // Before start
            {Type: "confidentiality_agreement", Required: true, Deadline: 0},
        },
    }

    system, err := NewHROnboardingSystem(config)
    if err != nil {
        log.Fatalf("Failed to initialize HR onboarding system: %v", err)
    }

    // Example new employee
    newEmployee := &NewEmployee{
        ID:            "emp-12345",
        Name:          "Jane Smith",
        Email:         "jane.smith@company.com",
        Department:    "Engineering",
        Role:          "Senior Software Engineer",
        Manager:       "john.doe@company.com",
        StartDate:     time.Now().Add(14 * 24 * time.Hour),
        Location:      "San Francisco",
        EmploymentType: "full-time",
    }

    ctx := context.Background()
    plan, err := system.StartOnboarding(ctx, newEmployee)
    if err != nil {
        log.Printf("Failed to start onboarding: %v", err)
    } else {
        fmt.Printf("Onboarding plan created for %s\n", newEmployee.Name)
        fmt.Printf("Total steps: %d\n", plan.TotalSteps)
        fmt.Printf("Estimated completion: %s\n", plan.EstimatedCompletion)
    }
}
```

---

## Contract Management Automation

Automate contract lifecycle management from creation to renewal, including review, negotiation, and compliance monitoring.

### Features
- Intelligent contract drafting
- Automated clause analysis
- Risk assessment and flagging
- Renewal tracking and alerts
- Compliance monitoring

### Implementation

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/agent/workflow"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
)

type ContractManagementSystem struct {
    draftingAgent      *core.LLMAgent
    analysisAgent      *core.LLMAgent
    riskAgent          *core.LLMAgent
    negotiationAgent   *core.LLMAgent
    complianceAgent    *core.LLMAgent
    workflowEngine     *workflow.ParallelAgent
    templateLibrary    *TemplateLibrary
    clauseDatabase     *ClauseDatabase
    config            *ContractConfig
}

type Contract struct {
    ID                string              `json:"id"`
    Type              string              `json:"type"`
    Title             string              `json:"title"`
    Parties           []Party             `json:"parties"`
    EffectiveDate     time.Time           `json:"effective_date"`
    ExpirationDate    time.Time           `json:"expiration_date"`
    Value             ContractValue       `json:"value"`
    Terms             []ContractTerm      `json:"terms"`
    Clauses           []Clause            `json:"clauses"`
    RiskAssessment    RiskAssessment      `json:"risk_assessment"`
    ComplianceStatus  ComplianceStatus    `json:"compliance_status"`
    Status            ContractStatus      `json:"status"`
    Version           int                 `json:"version"`
    Amendments        []Amendment         `json:"amendments"`
    RenewalTerms      *RenewalTerms       `json:"renewal_terms,omitempty"`
}

type RiskAssessment struct {
    OverallRisk       RiskLevel           `json:"overall_risk"`
    RiskFactors       []RiskFactor        `json:"risk_factors"`
    Recommendations   []string            `json:"recommendations"`
    ConfidenceScore   float64             `json:"confidence_score"`
}

func NewContractManagementSystem(config *ContractConfig) (*ContractManagementSystem, error) {
    // Initialize providers
    openaiProvider, err := provider.NewOpenAI(provider.OpenAIOptions{
        APIKey: config.OpenAIKey,
        Model:  "gpt-4",
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create OpenAI provider: %w", err)
    }

    anthropicProvider, err := provider.NewAnthropic(provider.AnthropicOptions{
        APIKey: config.AnthropicKey,
        Model:  "claude-3-sonnet-20240229",
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create Anthropic provider: %w", err)
    }

    // Create specialized contract agents
    draftingAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "contract-drafter",
        SystemPrompt: contractDraftingPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create drafting agent: %w", err)
    }

    analysisAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "contract-analyzer",
        SystemPrompt: contractAnalysisPrompt,
        Provider:     anthropicProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create analysis agent: %w", err)
    }

    riskAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "risk-assessor",
        SystemPrompt: riskAssessmentPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create risk agent: %w", err)
    }

    negotiationAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "negotiation-assistant",
        SystemPrompt: negotiationAssistancePrompt,
        Provider:     anthropicProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create negotiation agent: %w", err)
    }

    // Create parallel workflow for contract analysis
    workflowEngine := workflow.NewParallelAgent(workflow.ParallelAgentOptions{
        Name: "contract-analysis-workflow",
        Agents: []domain.Agent{
            analysisAgent,
            riskAgent,
            complianceAgent,
        },
        MergeStrategy: workflow.MergeAll,
    })

    return &ContractManagementSystem{
        draftingAgent:    draftingAgent,
        analysisAgent:    analysisAgent,
        riskAgent:        riskAgent,
        negotiationAgent: negotiationAgent,
        workflowEngine:   workflowEngine,
        templateLibrary:  NewTemplateLibrary(),
        clauseDatabase:   NewClauseDatabase(),
        config:          config,
    }, nil
}

func (cms *ContractManagementSystem) DraftContract(ctx context.Context, requirements *ContractRequirements) (*Contract, error) {
    // Get appropriate template
    template := cms.templateLibrary.GetTemplate(requirements.Type)
    
    // Get standard clauses
    standardClauses := cms.clauseDatabase.GetStandardClauses(requirements.Type)

    // Create drafting context
    draftCtx := domain.NewAgentContext().
        SetInputProperty("requirements", requirements).
        SetInputProperty("template", template).
        SetInputProperty("standard_clauses", standardClauses).
        SetInputProperty("jurisdiction", requirements.Jurisdiction)

    result, err := cms.draftingAgent.Execute(ctx, draftCtx)
    if err != nil {
        return nil, fmt.Errorf("failed to draft contract: %w", err)
    }

    var contract Contract
    // Parse drafted contract
    
    // Perform initial risk assessment
    riskAssessment, err := cms.AssessContractRisk(ctx, &contract)
    if err != nil {
        log.Printf("Risk assessment failed: %v", err)
    } else {
        contract.RiskAssessment = *riskAssessment
    }

    return &contract, nil
}

func (cms *ContractManagementSystem) AnalyzeContract(ctx context.Context, contractText string) (*ContractAnalysis, error) {
    analysisCtx := domain.NewAgentContext().
        SetInputProperty("contract_text", contractText).
        SetInputProperty("analysis_parameters", cms.config.AnalysisParameters).
        SetInputProperty("clause_library", cms.clauseDatabase.GetAllClauses())

    result, err := cms.workflowEngine.Execute(ctx, analysisCtx)
    if err != nil {
        return nil, fmt.Errorf("failed to analyze contract: %w", err)
    }

    analysis := &ContractAnalysis{
        Timestamp: time.Now(),
    }

    // Extract analysis results from parallel execution
    if clauseAnalysis := result.GetOutputProperty("clause_analysis"); clauseAnalysis != nil {
        // Parse clause analysis
    }
    if riskFactors := result.GetOutputProperty("risk_factors"); riskFactors != nil {
        // Parse risk factors
    }
    if compliance := result.GetOutputProperty("compliance_issues"); compliance != nil {
        // Parse compliance issues
    }

    return analysis, nil
}

func (cms *ContractManagementSystem) NegotiateTerms(ctx context.Context, contract *Contract, counterpartyPosition *NegotiationPosition) (*NegotiationStrategy, error) {
    negotiationCtx := domain.NewAgentContext().
        SetInputProperty("current_contract", contract).
        SetInputProperty("counterparty_position", counterpartyPosition).
        SetInputProperty("negotiation_limits", cms.config.NegotiationLimits).
        SetInputProperty("historical_negotiations", cms.getHistoricalNegotiations(contract.Type))

    result, err := cms.negotiationAgent.Execute(ctx, negotiationCtx)
    if err != nil {
        return nil, fmt.Errorf("failed to develop negotiation strategy: %w", err)
    }

    var strategy NegotiationStrategy
    // Parse negotiation strategy

    return &strategy, nil
}

func (cms *ContractManagementSystem) MonitorCompliance(ctx context.Context) {
    ticker := time.NewTicker(24 * time.Hour)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            activeContracts := cms.getActiveContracts()
            
            for _, contract := range activeContracts {
                // Check compliance status
                complianceStatus, err := cms.checkContractCompliance(ctx, contract)
                if err != nil {
                    log.Printf("Compliance check failed for contract %s: %v", contract.ID, err)
                    continue
                }

                // Update status and send alerts if needed
                if complianceStatus.HasIssues() {
                    cms.sendComplianceAlert(contract, complianceStatus)
                }

                // Check for upcoming renewals
                if cms.shouldPrepareRenewal(contract) {
                    cms.initiateRenewalProcess(ctx, contract)
                }
            }
        }
    }
}

const contractDraftingPrompt = `You are a contract drafting specialist. Your responsibilities:
- Draft clear, comprehensive contracts based on requirements
- Use appropriate legal language and structure
- Include all necessary clauses and provisions
- Ensure enforceability and clarity
- Protect client interests while maintaining fairness

Create professional contracts that minimize ambiguity and risk.`

const contractAnalysisPrompt = `You are a contract analysis expert. Analyze contracts for:
- Key terms and obligations
- Potential risks and liabilities
- Missing or problematic clauses
- Compliance with regulations
- Opportunities for improvement

Provide thorough analysis with actionable insights.`

const riskAssessmentPrompt = `You are a contract risk assessment specialist. Evaluate:
- Legal and financial risks
- Performance and delivery risks
- Compliance and regulatory risks
- Termination and dispute risks
- Force majeure and contingencies

Quantify risks and provide mitigation strategies.`

const negotiationAssistancePrompt = `You are a contract negotiation specialist. Your role:
- Develop negotiation strategies
- Identify win-win opportunities
- Suggest alternative terms
- Prioritize negotiation points
- Maintain relationship focus

Create strategies that achieve objectives while preserving relationships.`

func main() {
    config := &ContractConfig{
        OpenAIKey:     "your-openai-key",
        AnthropicKey:  "your-anthropic-key",
        DatabaseURL:   "postgres://user:pass@localhost/contracts",
        AnalysisParameters: AnalysisParameters{
            RiskThreshold:     0.7,
            ComplianceStrict:  true,
            JurisdictionFocus: []string{"US", "UK", "EU"},
        },
        NegotiationLimits: NegotiationLimits{
            MaxDiscountPercent: 15,
            MinContractValue:   10000,
            RequiredClauses:    []string{"liability_cap", "termination", "ip_ownership"},
        },
    }

    system, err := NewContractManagementSystem(config)
    if err != nil {
        log.Fatalf("Failed to initialize contract management system: %v", err)
    }

    // Example: Draft a new contract
    requirements := &ContractRequirements{
        Type:         "software_license",
        Parties:      []string{"Acme Corp", "Tech Solutions Inc"},
        Value:        50000,
        Duration:     12, // months
        Jurisdiction: "California",
        SpecialTerms: []string{"source_code_escrow", "performance_sla"},
    }

    ctx := context.Background()
    contract, err := system.DraftContract(ctx, requirements)
    if err != nil {
        log.Printf("Failed to draft contract: %v", err)
    } else {
        fmt.Printf("Contract drafted: %s\n", contract.Title)
        fmt.Printf("Risk level: %s\n", contract.RiskAssessment.OverallRisk)
        fmt.Printf("Clauses: %d\n", len(contract.Clauses))
    }

    // Start compliance monitoring
    go system.MonitorCompliance(ctx)

    select {}
}
```

---

## Supply Chain Optimization

Optimize supply chain operations with intelligent demand forecasting, inventory management, and logistics coordination.

### Features
- Demand prediction and planning
- Inventory optimization
- Supplier performance monitoring
- Logistics route optimization
- Risk mitigation

### Implementation

```go
package main

import (
    "context"
    "fmt"
    "log"
    "math"
    "time"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/agent/workflow"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
)

type SupplyChainOptimizer struct {
    forecastingAgent    *core.LLMAgent
    inventoryAgent      *core.LLMAgent
    logisticsAgent      *core.LLMAgent
    supplierAgent       *core.LLMAgent
    riskAgent           *core.LLMAgent
    optimizationWorkflow *workflow.ParallelAgent
    dataWarehouse       *DataWarehouse
    config             *SupplyChainConfig
}

type SupplyChainState struct {
    Timestamp           time.Time                  `json:"timestamp"`
    Inventory           map[string]InventoryLevel  `json:"inventory"`
    DemandForecast      map[string]DemandForecast  `json:"demand_forecast"`
    SupplierStatus      map[string]SupplierMetrics `json:"supplier_status"`
    LogisticsNetwork    LogisticsState             `json:"logistics_network"`
    RiskAssessment      SupplyChainRisks           `json:"risk_assessment"`
    OptimizationMetrics OptimizationMetrics        `json:"optimization_metrics"`
}

type DemandForecast struct {
    ProductID          string                     `json:"product_id"`
    ForecastPeriod     time.Duration              `json:"forecast_period"`
    PredictedDemand    float64                    `json:"predicted_demand"`
    ConfidenceInterval ConfidenceInterval         `json:"confidence_interval"`
    Seasonality        SeasonalityFactors         `json:"seasonality"`
    TrendAnalysis      TrendData                  `json:"trend_analysis"`
}

type OptimizationRecommendation struct {
    Type               RecommendationType         `json:"type"`
    Priority           Priority                   `json:"priority"`
    Description        string                     `json:"description"`
    ExpectedImpact     Impact                     `json:"expected_impact"`
    Implementation     ImplementationPlan         `json:"implementation"`
    ROI                float64                    `json:"roi"`
}

func NewSupplyChainOptimizer(config *SupplyChainConfig) (*SupplyChainOptimizer, error) {
    openaiProvider, err := provider.NewOpenAI(provider.OpenAIOptions{
        APIKey: config.OpenAIKey,
        Model:  "gpt-4",
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create OpenAI provider: %w", err)
    }

    // Create specialized supply chain agents
    forecastingAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "demand-forecaster",
        SystemPrompt: demandForecastingPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create forecasting agent: %w", err)
    }

    inventoryAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "inventory-optimizer",
        SystemPrompt: inventoryOptimizationPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create inventory agent: %w", err)
    }

    logisticsAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "logistics-coordinator",
        SystemPrompt: logisticsOptimizationPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create logistics agent: %w", err)
    }

    supplierAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "supplier-analyzer",
        SystemPrompt: supplierAnalysisPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create supplier agent: %w", err)
    }

    riskAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "risk-analyzer",
        SystemPrompt: supplyChainRiskPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create risk agent: %w", err)
    }

    // Create optimization workflow
    optimizationWorkflow := workflow.NewParallelAgent(workflow.ParallelAgentOptions{
        Name: "supply-chain-optimization",
        Agents: []domain.Agent{
            forecastingAgent,
            inventoryAgent,
            logisticsAgent,
            supplierAgent,
            riskAgent,
        },
        MergeStrategy: workflow.MergeWeighted,
    })

    return &SupplyChainOptimizer{
        forecastingAgent:     forecastingAgent,
        inventoryAgent:       inventoryAgent,
        logisticsAgent:       logisticsAgent,
        supplierAgent:        supplierAgent,
        riskAgent:           riskAgent,
        optimizationWorkflow: optimizationWorkflow,
        dataWarehouse:       NewDataWarehouse(config.DatabaseURL),
        config:              config,
    }, nil
}

func (sco *SupplyChainOptimizer) OptimizeSupplyChain(ctx context.Context) (*OptimizationPlan, error) {
    // Get current state
    currentState, err := sco.getCurrentState()
    if err != nil {
        return nil, fmt.Errorf("failed to get current state: %w", err)
    }

    // Create optimization context
    optCtx := domain.NewAgentContext().
        SetInputProperty("current_state", currentState).
        SetInputProperty("historical_data", sco.getHistoricalData()).
        SetInputProperty("market_conditions", sco.getMarketConditions()).
        SetInputProperty("constraints", sco.config.Constraints)

    // Run parallel optimization
    result, err := sco.optimizationWorkflow.Execute(ctx, optCtx)
    if err != nil {
        return nil, fmt.Errorf("optimization failed: %w", err)
    }

    // Synthesize optimization plan
    plan := sco.synthesizeOptimizationPlan(result)

    // Validate and prioritize recommendations
    plan.Recommendations = sco.prioritizeRecommendations(plan.Recommendations)

    return plan, nil
}

func (sco *SupplyChainOptimizer) ForecastDemand(ctx context.Context, productID string, horizon time.Duration) (*DemandForecast, error) {
    forecastCtx := domain.NewAgentContext().
        SetInputProperty("product_id", productID).
        SetInputProperty("forecast_horizon", horizon).
        SetInputProperty("historical_sales", sco.getHistoricalSales(productID)).
        SetInputProperty("external_factors", sco.getExternalFactors()).
        SetInputProperty("seasonality_data", sco.getSeasonalityData(productID))

    result, err := sco.forecastingAgent.Execute(ctx, forecastCtx)
    if err != nil {
        return nil, fmt.Errorf("demand forecasting failed: %w", err)
    }

    forecast := &DemandForecast{
        ProductID:       productID,
        ForecastPeriod:  horizon,
        PredictedDemand: result.GetOutputProperty("predicted_demand").(float64),
    }

    // Add confidence intervals and analysis
    if ci := result.GetOutputProperty("confidence_interval"); ci != nil {
        // Parse confidence interval
    }

    return forecast, nil
}

func (sco *SupplyChainOptimizer) OptimizeInventory(ctx context.Context) ([]InventoryRecommendation, error) {
    inventoryCtx := domain.NewAgentContext().
        SetInputProperty("current_inventory", sco.getCurrentInventory()).
        SetInputProperty("demand_forecasts", sco.getAllDemandForecasts()).
        SetInputProperty("lead_times", sco.getSupplierLeadTimes()).
        SetInputProperty("holding_costs", sco.config.HoldingCosts).
        SetInputProperty("stockout_costs", sco.config.StockoutCosts)

    result, err := sco.inventoryAgent.Execute(ctx, inventoryCtx)
    if err != nil {
        return nil, fmt.Errorf("inventory optimization failed: %w", err)
    }

    var recommendations []InventoryRecommendation
    // Parse inventory recommendations

    return recommendations, nil
}

func (sco *SupplyChainOptimizer) ContinuousOptimization(ctx context.Context) {
    ticker := time.NewTicker(sco.config.OptimizationInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            log.Println("Running supply chain optimization cycle...")

            // Run optimization
            plan, err := sco.OptimizeSupplyChain(ctx)
            if err != nil {
                log.Printf("Optimization error: %v", err)
                continue
            }

            // Execute high-priority recommendations
            for _, rec := range plan.Recommendations {
                if rec.Priority == PriorityHigh && rec.AutoExecutable {
                    if err := sco.executeRecommendation(ctx, rec); err != nil {
                        log.Printf("Failed to execute recommendation: %v", err)
                    }
                }
            }

            // Monitor execution results
            sco.monitorOptimizationResults(plan)
        }
    }
}

const demandForecastingPrompt = `You are a demand forecasting specialist. Your responsibilities:
- Analyze historical sales data and trends
- Consider seasonal patterns and cyclicality
- Factor in external events and market conditions
- Use multiple forecasting methods
- Provide confidence intervals and risk assessments

Generate accurate demand forecasts to optimize inventory and planning.`

const inventoryOptimizationPrompt = `You are an inventory optimization expert. Your tasks:
- Calculate optimal inventory levels
- Balance holding costs vs stockout risks
- Implement just-in-time principles where appropriate
- Consider lead times and variability
- Optimize safety stock levels

Minimize total inventory costs while maintaining service levels.`

const logisticsOptimizationPrompt = `You are a logistics optimization specialist. Focus on:
- Route optimization for deliveries
- Warehouse location and capacity planning
- Transportation mode selection
- Last-mile delivery efficiency
- Cross-docking opportunities

Optimize the logistics network for cost and service level.`

const supplierAnalysisPrompt = `You are a supplier analysis expert. Evaluate:
- Supplier performance metrics
- Quality and reliability scores
- Cost competitiveness
- Risk factors and dependencies
- Alternative sourcing options

Optimize supplier relationships and mitigate supply risks.`

const supplyChainRiskPrompt = `You are a supply chain risk analyst. Identify and assess:
- Supply disruption risks
- Demand volatility risks
- Geopolitical and regulatory risks
- Quality and compliance risks
- Financial and operational risks

Provide risk mitigation strategies and contingency plans.`

// Optimization algorithms
func (sco *SupplyChainOptimizer) calculateOptimalOrderQuantity(product string, demandRate, orderingCost, holdingCost float64) float64 {
    // Economic Order Quantity (EOQ) formula
    return math.Sqrt((2 * demandRate * orderingCost) / holdingCost)
}

func (sco *SupplyChainOptimizer) calculateSafetyStock(avgDemand, demandStdDev, leadTime, leadTimeStdDev, serviceLevel float64) float64 {
    // Safety stock calculation with demand and lead time variability
    zScore := sco.getZScore(serviceLevel)
    demandVariability := math.Sqrt(leadTime)*demandStdDev
    leadTimeVariability := avgDemand*leadTimeStdDev
    
    return zScore * math.Sqrt(math.Pow(demandVariability, 2) + math.Pow(leadTimeVariability, 2))
}

func main() {
    config := &SupplyChainConfig{
        OpenAIKey:    "your-openai-key",
        DatabaseURL:  "postgres://user:pass@localhost/supply_chain",
        OptimizationInterval: 6 * time.Hour,
        Constraints: OptimizationConstraints{
            MaxInventoryValue:    1000000,
            MinServiceLevel:      0.95,
            MaxTransportCost:     50000,
            PreferredSuppliers:   []string{"supplier_a", "supplier_b"},
        },
        HoldingCosts: map[string]float64{
            "default": 0.2, // 20% of product value per year
        },
        StockoutCosts: map[string]float64{
            "default": 50.0, // per unit
        },
    }

    optimizer, err := NewSupplyChainOptimizer(config)
    if err != nil {
        log.Fatalf("Failed to initialize supply chain optimizer: %v", err)
    }

    ctx := context.Background()

    // Example: Forecast demand for a product
    forecast, err := optimizer.ForecastDemand(ctx, "PROD-12345", 30*24*time.Hour)
    if err != nil {
        log.Printf("Forecasting error: %v", err)
    } else {
        fmt.Printf("Demand forecast for PROD-12345: %.2f units\n", forecast.PredictedDemand)
    }

    // Run optimization
    plan, err := optimizer.OptimizeSupplyChain(ctx)
    if err != nil {
        log.Printf("Optimization error: %v", err)
    } else {
        fmt.Printf("Optimization plan generated with %d recommendations\n", len(plan.Recommendations))
        fmt.Printf("Expected cost savings: $%.2f\n", plan.ExpectedSavings)
    }

    // Start continuous optimization
    go optimizer.ContinuousOptimization(ctx)

    select {}
}
```

---

## Customer Service Automation

Create an intelligent customer service system that handles inquiries across multiple channels with personalized responses.

### Features
- Multi-channel support (email, chat, voice)
- Intent recognition and routing
- Personalized response generation
- Escalation management
- Knowledge base integration

### Implementation

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/agent/workflow"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
)

type CustomerServiceAutomation struct {
    intentAgent         *core.LLMAgent
    responseAgent       *core.LLMAgent
    sentimentAgent      *core.LLMAgent
    escalationAgent     *core.LLMAgent
    resolutionAgent     *core.LLMAgent
    workflowOrchestrator *workflow.ConditionalAgent
    knowledgeBase       *KnowledgeBase
    customerDB          *CustomerDatabase
    config             *ServiceConfig
}

type CustomerInquiry struct {
    ID              string             `json:"id"`
    CustomerID      string             `json:"customer_id"`
    Channel         CommunicationChannel `json:"channel"`
    Message         string             `json:"message"`
    Timestamp       time.Time          `json:"timestamp"`
    Intent          CustomerIntent     `json:"intent"`
    Sentiment       SentimentAnalysis  `json:"sentiment"`
    Priority        Priority           `json:"priority"`
    Context         ConversationContext `json:"context"`
    Resolution      *Resolution        `json:"resolution,omitempty"`
}

type CustomerIntent struct {
    Primary         string             `json:"primary"`
    Secondary       []string           `json:"secondary"`
    Confidence      float64            `json:"confidence"`
    Entities        []Entity           `json:"entities"`
    RequiresAction  []RequiredAction   `json:"requires_action"`
}

type ServiceResponse struct {
    ResponseID      string             `json:"response_id"`
    Content         string             `json:"content"`
    Tone            ResponseTone       `json:"tone"`
    Personalization PersonalizationData `json:"personalization"`
    Actions         []AutomatedAction  `json:"actions"`
    NextSteps       []NextStep         `json:"next_steps"`
    SatisfactionPrediction float64     `json:"satisfaction_prediction"`
}

func NewCustomerServiceAutomation(config *ServiceConfig) (*CustomerServiceAutomation, error) {
    openaiProvider, err := provider.NewOpenAI(provider.OpenAIOptions{
        APIKey: config.OpenAIKey,
        Model:  "gpt-4",
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create OpenAI provider: %w", err)
    }

    // Create specialized service agents
    intentAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "intent-classifier",
        SystemPrompt: intentClassificationPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create intent agent: %w", err)
    }

    responseAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "response-generator",
        SystemPrompt: responseGenerationPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create response agent: %w", err)
    }

    sentimentAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "sentiment-analyzer",
        SystemPrompt: sentimentAnalysisPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create sentiment agent: %w", err)
    }

    escalationAgent, err := core.NewLLMAgent(core.LLMAgentOptions{
        Name:         "escalation-manager",
        SystemPrompt: escalationManagementPrompt,
        Provider:     openaiProvider,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create escalation agent: %w", err)
    }

    // Create workflow orchestrator
    workflowOrchestrator := workflow.NewConditionalAgent(workflow.ConditionalAgentOptions{
        Name: "customer-service-workflow",
        Conditions: []workflow.Condition{
            {
                Name: "intent-classification",
                Predicate: func(ctx *domain.AgentContext) bool {
                    return ctx.GetInputProperty("stage") == "intent"
                },
                Agent: intentAgent,
            },
            {
                Name: "sentiment-analysis",
                Predicate: func(ctx *domain.AgentContext) bool {
                    return ctx.GetInputProperty("stage") == "sentiment"
                },
                Agent: sentimentAgent,
            },
            {
                Name: "response-generation",
                Predicate: func(ctx *domain.AgentContext) bool {
                    return ctx.GetInputProperty("stage") == "response"
                },
                Agent: responseAgent,
            },
            {
                Name: "escalation-check",
                Predicate: func(ctx *domain.AgentContext) bool {
                    return ctx.GetInputProperty("needs_escalation").(bool)
                },
                Agent: escalationAgent,
            },
        },
    })

    return &CustomerServiceAutomation{
        intentAgent:          intentAgent,
        responseAgent:        responseAgent,
        sentimentAgent:       sentimentAgent,
        escalationAgent:      escalationAgent,
        workflowOrchestrator: workflowOrchestrator,
        knowledgeBase:        NewKnowledgeBase(),
        customerDB:           NewCustomerDatabase(config.DatabaseURL),
        config:              config,
    }, nil
}

func (csa *CustomerServiceAutomation) HandleInquiry(ctx context.Context, inquiry *CustomerInquiry) (*ServiceResponse, error) {
    // Get customer profile
    customer, err := csa.customerDB.GetCustomer(inquiry.CustomerID)
    if err != nil {
        return nil, fmt.Errorf("failed to get customer profile: %w", err)
    }

    // Step 1: Classify intent
    intentCtx := domain.NewAgentContext().
        SetInputProperty("stage", "intent").
        SetInputProperty("message", inquiry.Message).
        SetInputProperty("customer_history", customer.InteractionHistory).
        SetInputProperty("context", inquiry.Context)

    intentResult, err := csa.workflowOrchestrator.Execute(ctx, intentCtx)
    if err != nil {
        return nil, fmt.Errorf("intent classification failed: %w", err)
    }

    // Parse intent
    var intent CustomerIntent
    if intentData := intentResult.GetOutputProperty("intent"); intentData != nil {
        // Parse intent data
    }
    inquiry.Intent = intent

    // Step 2: Analyze sentiment
    sentimentCtx := domain.NewAgentContext().
        SetInputProperty("stage", "sentiment").
        SetInputProperty("message", inquiry.Message).
        SetInputProperty("customer_profile", customer)

    sentimentResult, err := csa.workflowOrchestrator.Execute(ctx, sentimentCtx)
    if err != nil {
        log.Printf("Sentiment analysis failed: %v", err)
    }

    // Step 3: Check for escalation needs
    needsEscalation := csa.checkEscalationCriteria(inquiry, customer)
    
    if needsEscalation {
        escalationCtx := domain.NewAgentContext().
            SetInputProperty("needs_escalation", true).
            SetInputProperty("inquiry", inquiry).
            SetInputProperty("escalation_reason", csa.getEscalationReason(inquiry))

        escalationResult, err := csa.workflowOrchestrator.Execute(ctx, escalationCtx)
        if err != nil {
            log.Printf("Escalation failed: %v", err)
        }
        // Handle escalation
        return csa.createEscalationResponse(escalationResult), nil
    }

    // Step 4: Generate response
    responseCtx := domain.NewAgentContext().
        SetInputProperty("stage", "response").
        SetInputProperty("inquiry", inquiry).
        SetInputProperty("customer", customer).
        SetInputProperty("knowledge_base", csa.knowledgeBase.SearchRelevant(inquiry.Intent.Primary)).
        SetInputProperty("tone_preference", customer.CommunicationPreference)

    responseResult, err := csa.workflowOrchestrator.Execute(ctx, responseCtx)
    if err != nil {
        return nil, fmt.Errorf("response generation failed: %w", err)
    }

    // Create service response
    response := &ServiceResponse{
        ResponseID: generateResponseID(),
        Content:    responseResult.GetOutputProperty("response_content").(string),
    }

    // Add personalization
    if personalization := responseResult.GetOutputProperty("personalization"); personalization != nil {
        // Parse personalization data
    }

    // Execute automated actions
    if actions := responseResult.GetOutputProperty("automated_actions"); actions != nil {
        go csa.executeAutomatedActions(ctx, actions)
    }

    // Update interaction history
    csa.customerDB.RecordInteraction(inquiry, response)

    return response, nil
}

func (csa *CustomerServiceAutomation) checkEscalationCriteria(inquiry *CustomerInquiry, customer *Customer) bool {
    // Check various escalation criteria
    criteria := []func() bool{
        func() bool { return inquiry.Sentiment.Score < -0.7 }, // Very negative sentiment
        func() bool { return customer.ChurnRisk > 0.8 },       // High churn risk
        func() bool { return inquiry.Priority == PriorityUrgent },
        func() bool { return containsEscalationKeywords(inquiry.Message) },
        func() bool { return customer.LifetimeValue > 10000 }, // High-value customer
    }

    for _, check := range criteria {
        if check() {
            return true
        }
    }

    return false
}

const intentClassificationPrompt = `You are a customer intent classification specialist. Your role:
- Identify primary and secondary intents from customer messages
- Extract relevant entities (products, services, issues)
- Determine required actions to resolve the inquiry
- Consider conversation context and history
- Provide confidence scores for classifications

Accurately classify intents to enable appropriate responses and actions.`

const responseGenerationPrompt = `You are a customer service response specialist. Your responsibilities:
- Generate helpful, empathetic responses to customer inquiries
- Personalize responses based on customer profile and history
- Provide clear solutions and next steps
- Maintain appropriate tone and professionalism
- Include relevant information from knowledge base

Create responses that resolve issues effectively while building customer satisfaction.`

const sentimentAnalysisPrompt = `You are a customer sentiment analysis expert. Analyze:
- Emotional tone and intensity in customer messages
- Frustration, satisfaction, or urgency indicators
- Changes in sentiment throughout conversation
- Risk indicators for churn or escalation
- Opportunities for positive engagement

Provide nuanced sentiment analysis to guide appropriate responses.`

const escalationManagementPrompt = `You are an escalation management specialist. Your tasks:
- Determine appropriate escalation paths
- Prioritize based on urgency and impact
- Route to specialized teams or agents
- Provide context for human agents
- Suggest de-escalation strategies

Manage escalations to ensure critical issues receive appropriate attention.`

func main() {
    config := &ServiceConfig{
        OpenAIKey:   "your-openai-key",
        DatabaseURL: "postgres://user:pass@localhost/customer_service",
        Channels: []CommunicationChannel{
            ChannelEmail,
            ChannelChat,
            ChannelVoice,
            ChannelSocial,
        },
        ResponseTimeTargets: map[CommunicationChannel]time.Duration{
            ChannelChat:   30 * time.Second,
            ChannelEmail:  2 * time.Hour,
            ChannelVoice:  0, // Real-time
            ChannelSocial: 1 * time.Hour,
        },
    }

    automation, err := NewCustomerServiceAutomation(config)
    if err != nil {
        log.Fatalf("Failed to initialize customer service automation: %v", err)
    }

    // Example inquiry
    inquiry := &CustomerInquiry{
        ID:         "INQ-12345",
        CustomerID: "CUST-67890",
        Channel:    ChannelChat,
        Message:    "I've been trying to reset my password for the last hour and nothing is working! This is ridiculous!",
        Timestamp:  time.Now(),
    }

    ctx := context.Background()
    response, err := automation.HandleInquiry(ctx, inquiry)
    if err != nil {
        log.Printf("Failed to handle inquiry: %v", err)
    } else {
        fmt.Printf("Generated response: %s\n", response.Content)
        fmt.Printf("Satisfaction prediction: %.2f\n", response.SatisfactionPrediction)
    }
}
```

---

## Summary

These business automation examples demonstrate how Go-LLMs can transform various business processes:

1. **Invoice Processing** - End-to-end automation with validation and fraud detection
2. **HR Onboarding** - Comprehensive employee onboarding coordination
3. **Contract Management** - Intelligent contract lifecycle management
4. **Supply Chain Optimization** - Demand forecasting and inventory optimization
5. **Customer Service** - Multi-channel support with intelligent routing

Each implementation showcases:
- **Process Intelligence** - Beyond simple rule-based automation
- **Integration Capabilities** - Seamless connection with existing systems
- **Adaptive Behavior** - Systems that learn and improve
- **Business Value** - Direct impact on efficiency and cost savings
- **Compliance & Governance** - Built-in controls and audit trails

These examples provide templates for building sophisticated business automation solutions that combine AI intelligence with practical business requirements.

> **Next:** [Education Tools](education-tools.md) - Educational applications and learning systems