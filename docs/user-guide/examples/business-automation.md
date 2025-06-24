# Business Automation: Process Automation

> **[Project Root](/) / [Documentation](../..) / [User Guide](../../user-guide) / [Examples](../../user-guide/examples) / Business Automation**

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
provider := provider.NewOpenAIProvider(
)
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
provider := provider.NewOpenAIProvider(
)
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
provider := provider.NewOpenAIProvider(
)
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
provider := provider.NewOpenAIProvider(
)
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
provider := provider.NewOpenAIProvider(
)
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