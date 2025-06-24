# Data Analysis: Data Insights Generation

> **[Project Root](/) / [Documentation](/docs/) / [User Guide](/docs/user-guide/) / [Examples](/docs/user-guide/examples/) / Data Analysis**

Build an AI-powered data analysis and insights generation system that automates data processing, statistical analysis, visualization generation, and predictive modeling. This example demonstrates how to combine AI agents with data science workflows for comprehensive business intelligence.

## System Overview

This data analysis system provides:

- **Automated Data Processing** - ETL pipelines with intelligent data cleaning
- **Statistical Analysis** - Descriptive and inferential statistics with AI interpretation
- **Pattern Recognition** - Anomaly detection and trend identification
- **Predictive Modeling** - Forecast generation and scenario analysis
- **Natural Language Insights** - AI-generated explanations of findings
- **Interactive Visualizations** - Dynamic charts and dashboards
- **Report Generation** - Executive summaries and technical reports
- **Real-time Monitoring** - Live data streaming and alerts

## Architecture

![Data Analysis System Architecture](../../images/data-analysis-architecture.svg)

### Components
1. **Data Ingestion Engine** - Multi-source data collection and processing
2. **Analysis Orchestrator** - Coordinates analysis workflows
3. **Statistical Analyzer** - Performs statistical computations
4. **Pattern Detector** - Identifies trends, anomalies, and relationships
5. **Insight Generator** - Creates natural language explanations
6. **Visualization Engine** - Generates charts and interactive dashboards
7. **Predictive Modeler** - Builds and applies forecasting models
8. **Report Synthesizer** - Combines findings into comprehensive reports

---

## Complete Implementation

```go
package main

import (
    "context"
    "database/sql"
    "encoding/csv"
    "encoding/json"
    "fmt"
    "log"
    "math"
    "net/http"
    "os"
    "sort"
    "strconv"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    
    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/agent/workflow"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
)

// DataAnalysisSystem is the main system orchestrator
type DataAnalysisSystem struct {
    db                  *sqlx.DB
    ingestionAgent      *core.LLMAgent
    statisticalAgent    *core.LLMAgent
    patternAgent        *core.LLMAgent
    insightAgent        *core.LLMAgent
    visualizationAgent  *core.LLMAgent
    predictiveAgent     *core.LLMAgent
    reportAgent         *core.LLMAgent
    workflowAgent       *workflow.SequentialAgent
    dataProcessor       *DataProcessor
    statisticalEngine   *StatisticalEngine
    visualizationEngine *VisualizationEngine
    config             *AnalysisConfig
    metrics            *AnalysisMetrics
}

type AnalysisConfig struct {
    DatabaseURL          string                 `json:"database_url"`
    OpenAIKey           string                 `json:"openai_key"`
    MaxDataPoints       int                    `json:"max_data_points"`
    StatisticalMethods  []string               `json:"statistical_methods"`
    VisualizationTypes  []string               `json:"visualization_types"`
    PredictiveModels    []string               `json:"predictive_models"`
    ConfidenceThreshold float64                `json:"confidence_threshold"`
    AnomalyThreshold    float64                `json:"anomaly_threshold"`
    TrendThreshold      float64                `json:"trend_threshold"`
    ReportFormats       []string               `json:"report_formats"`
    RealTimeEnabled     bool                   `json:"real_time_enabled"`
    CacheEnabled        bool                   `json:"cache_enabled"`
}

// Data models
type AnalysisProject struct {
    ID               int                    `json:"id" db:"id"`
    Name             string                 `json:"name" db:"name"`
    Description      string                 `json:"description" db:"description"`
    DataSources      []string               `json:"data_sources" db:"data_sources"`
    AnalysisType     string                 `json:"analysis_type" db:"analysis_type"` // descriptive, diagnostic, predictive, prescriptive
    Objectives       []string               `json:"objectives" db:"objectives"`
    Status           string                 `json:"status" db:"status"` // queued, processing, completed, failed
    RequestedBy      string                 `json:"requested_by" db:"requested_by"`
    Priority         string                 `json:"priority" db:"priority"` // low, medium, high, urgent
    ScheduledFor     *time.Time             `json:"scheduled_for" db:"scheduled_for"`
    StartedAt        *time.Time             `json:"started_at" db:"started_at"`
    CompletedAt      *time.Time             `json:"completed_at" db:"completed_at"`
    DataPointCount   int                    `json:"data_point_count" db:"data_point_count"`
    InsightCount     int                    `json:"insight_count" db:"insight_count"`
    AnomalyCount     int                    `json:"anomaly_count" db:"anomaly_count"`
    OverallScore     float64                `json:"overall_score" db:"overall_score"`
    CreatedAt        time.Time              `json:"created_at" db:"created_at"`
    UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`
    Metadata         map[string]interface{} `json:"metadata" db:"metadata"`
}

type Dataset struct {
    ID            int                    `json:"id" db:"id"`
    ProjectID     int                    `json:"project_id" db:"project_id"`
    Name          string                 `json:"name" db:"name"`
    Source        string                 `json:"source" db:"source"`
    Type          string                 `json:"type" db:"type"` // csv, json, sql, api
    Schema        DataSchema             `json:"schema" db:"schema"`
    RowCount      int                    `json:"row_count" db:"row_count"`
    ColumnCount   int                    `json:"column_count" db:"column_count"`
    Quality       DataQuality            `json:"quality" db:"quality"`
    ProcessingStatus string              `json:"processing_status" db:"processing_status"`
    LoadedAt      time.Time              `json:"loaded_at" db:"loaded_at"`
    Metadata      map[string]interface{} `json:"metadata" db:"metadata"`
}

type DataSchema struct {
    Columns []ColumnInfo `json:"columns"`
}

type ColumnInfo struct {
    Name        string      `json:"name"`
    Type        string      `json:"type"` // numeric, categorical, datetime, text
    DataType    string      `json:"data_type"` // int, float, string, timestamp
    NullCount   int         `json:"null_count"`
    UniqueCount int         `json:"unique_count"`
    Min         interface{} `json:"min,omitempty"`
    Max         interface{} `json:"max,omitempty"`
    Mean        *float64    `json:"mean,omitempty"`
    StdDev      *float64    `json:"std_dev,omitempty"`
    Sample      []interface{} `json:"sample"`
}

type DataQuality struct {
    OverallScore     float64            `json:"overall_score"`
    Completeness     float64            `json:"completeness"`
    Accuracy         float64            `json:"accuracy"`
    Consistency      float64            `json:"consistency"`
    Validity         float64            `json:"validity"`
    Issues           []DataQualityIssue `json:"issues"`
    Recommendations  []string           `json:"recommendations"`
}

type DataQualityIssue struct {
    Type        string  `json:"type"` // missing_values, outliers, inconsistency, format_error
    Column      string  `json:"column"`
    Severity    string  `json:"severity"` // low, medium, high, critical
    Count       int     `json:"count"`
    Description string  `json:"description"`
    Suggestion  string  `json:"suggestion"`
}

type StatisticalAnalysis struct {
    ID               int                    `json:"id" db:"id"`
    ProjectID        int                    `json:"project_id" db:"project_id"`
    DatasetID        int                    `json:"dataset_id" db:"dataset_id"`
    Type             string                 `json:"type" db:"type"` // descriptive, correlation, regression, test
    Method           string                 `json:"method" db:"method"`
    Variables        []string               `json:"variables" db:"variables"`
    Results          StatisticalResults     `json:"results" db:"results"`
    Interpretation   string                 `json:"interpretation" db:"interpretation"`
    Confidence       float64                `json:"confidence" db:"confidence"`
    PValue           *float64               `json:"p_value" db:"p_value"`
    EffectSize       *float64               `json:"effect_size" db:"effect_size"`
    CreatedAt        time.Time              `json:"created_at" db:"created_at"`
    Metadata         map[string]interface{} `json:"metadata" db:"metadata"`
}

type StatisticalResults struct {
    Summary      map[string]interface{} `json:"summary"`
    Distribution map[string]interface{} `json:"distribution"`
    Correlation  map[string]interface{} `json:"correlation"`
    Regression   map[string]interface{} `json:"regression"`
    Tests        map[string]interface{} `json:"tests"`
    Confidence   map[string]interface{} `json:"confidence"`
}

type Pattern struct {
    ID            int                    `json:"id" db:"id"`
    ProjectID     int                    `json:"project_id" db:"project_id"`
    Type          string                 `json:"type" db:"type"` // trend, seasonality, anomaly, correlation
    Name          string                 `json:"name" db:"name"`
    Description   string                 `json:"description" db:"description"`
    Variables     []string               `json:"variables" db:"variables"`
    Strength      float64                `json:"strength" db:"strength"` // 0.0 to 1.0
    Confidence    float64                `json:"confidence" db:"confidence"`
    StartPeriod   *time.Time             `json:"start_period" db:"start_period"`
    EndPeriod     *time.Time             `json:"end_period" db:"end_period"`
    Parameters    map[string]interface{} `json:"parameters" db:"parameters"`
    Implications  []string               `json:"implications" db:"implications"`
    Recommendations []string             `json:"recommendations" db:"recommendations"`
    DetectedAt    time.Time              `json:"detected_at" db:"detected_at"`
    Metadata      map[string]interface{} `json:"metadata" db:"metadata"`
}

type Insight struct {
    ID              int                    `json:"id" db:"id"`
    ProjectID       int                    `json:"project_id" db:"project_id"`
    Type            string                 `json:"type" db:"type"` // observation, hypothesis, recommendation, alert
    Title           string                 `json:"title" db:"title"`
    Summary         string                 `json:"summary" db:"summary"`
    Details         string                 `json:"details" db:"details"`
    Category        string                 `json:"category" db:"category"`
    Priority        string                 `json:"priority" db:"priority"` // low, medium, high, critical
    Confidence      float64                `json:"confidence" db:"confidence"`
    Impact          string                 `json:"impact" db:"impact"` // low, medium, high
    ActionRequired  bool                   `json:"action_required" db:"action_required"`
    SupportingData  []EvidenceItem         `json:"supporting_data" db:"supporting_data"`
    RelatedPatterns []int                  `json:"related_patterns" db:"related_patterns"`
    Visualizations  []string               `json:"visualizations" db:"visualizations"`
    GeneratedAt     time.Time              `json:"generated_at" db:"generated_at"`
    Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
}

type EvidenceItem struct {
    Type        string      `json:"type"` // statistic, pattern, correlation, trend
    Description string      `json:"description"`
    Value       interface{} `json:"value"`
    Confidence  float64     `json:"confidence"`
    Source      string      `json:"source"`
}

type Visualization struct {
    ID          int                    `json:"id" db:"id"`
    ProjectID   int                    `json:"project_id" db:"project_id"`
    Type        string                 `json:"type" db:"type"` // chart, graph, heatmap, dashboard
    ChartType   string                 `json:"chart_type" db:"chart_type"` // bar, line, scatter, pie, histogram
    Title       string                 `json:"title" db:"title"`
    Description string                 `json:"description" db:"description"`
    Data        interface{}            `json:"data" db:"data"`
    Config      VisualizationConfig    `json:"config" db:"config"`
    Format      string                 `json:"format" db:"format"` // svg, png, html, json
    URL         string                 `json:"url" db:"url"`
    CreatedAt   time.Time              `json:"created_at" db:"created_at"`
    Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
}

type VisualizationConfig struct {
    Width      int                    `json:"width"`
    Height     int                    `json:"height"`
    Theme      string                 `json:"theme"`
    Colors     []string               `json:"colors"`
    Axes       map[string]interface{} `json:"axes"`
    Legend     map[string]interface{} `json:"legend"`
    Animations bool                   `json:"animations"`
    Interactive bool                  `json:"interactive"`
}

type Prediction struct {
    ID              int                    `json:"id" db:"id"`
    ProjectID       int                    `json:"project_id" db:"project_id"`
    Model           string                 `json:"model" db:"model"` // linear_regression, arima, prophet, neural_network
    Variables       []string               `json:"variables" db:"variables"`
    TargetVariable  string                 `json:"target_variable" db:"target_variable"`
    Horizon         int                    `json:"horizon" db:"horizon"` // prediction periods
    Predictions     []PredictionPoint      `json:"predictions" db:"predictions"`
    Accuracy        ModelAccuracy          `json:"accuracy" db:"accuracy"`
    Confidence      float64                `json:"confidence" db:"confidence"`
    Assumptions     []string               `json:"assumptions" db:"assumptions"`
    Limitations     []string               `json:"limitations" db:"limitations"`
    CreatedAt       time.Time              `json:"created_at" db:"created_at"`
    Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
}

type PredictionPoint struct {
    Period     time.Time `json:"period"`
    Value      float64   `json:"value"`
    LowerBound float64   `json:"lower_bound"`
    UpperBound float64   `json:"upper_bound"`
    Confidence float64   `json:"confidence"`
}

type ModelAccuracy struct {
    MAE         float64 `json:"mae"`          // Mean Absolute Error
    RMSE        float64 `json:"rmse"`         // Root Mean Square Error
    MAPE        float64 `json:"mape"`         // Mean Absolute Percentage Error
    R2          float64 `json:"r2"`           // R-squared
    AdjustedR2  float64 `json:"adjusted_r2"` // Adjusted R-squared
    AIC         float64 `json:"aic"`          // Akaike Information Criterion
}

type AnalysisReport struct {
    ID                 int                    `json:"id" db:"id"`
    ProjectID          int                    `json:"project_id" db:"project_id"`
    Title              string                 `json:"title" db:"title"`
    ExecutiveSummary   string                 `json:"executive_summary" db:"executive_summary"`
    KeyFindings        []string               `json:"key_findings" db:"key_findings"`
    Recommendations    []string               `json:"recommendations" db:"recommendations"`
    Methodology        string                 `json:"methodology" db:"methodology"`
    DataSummary        string                 `json:"data_summary" db:"data_summary"`
    StatisticalSummary string                 `json:"statistical_summary" db:"statistical_summary"`
    VisualizationSummary string               `json:"visualization_summary" db:"visualization_summary"`
    Conclusions        string                 `json:"conclusions" db:"conclusions"`
    Limitations        []string               `json:"limitations" db:"limitations"`
    NextSteps          []string               `json:"next_steps" db:"next_steps"`
    Format             string                 `json:"format" db:"format"` // markdown, html, pdf
    Content            string                 `json:"content" db:"content"`
    GeneratedAt        time.Time              `json:"generated_at" db:"generated_at"`
    Metadata           map[string]interface{} `json:"metadata" db:"metadata"`
}

type AnalysisMetrics struct {
    ProjectsProcessed   prometheus.Counter
    DataPointsAnalyzed  prometheus.Counter
    InsightsGenerated   prometheus.Counter
    PatternsDetected    prometheus.Counter
    PredictionsCreated  prometheus.Counter
    ProcessingTime      prometheus.Histogram
    AccuracyScores      prometheus.Histogram
    DataQualityScores   prometheus.Histogram
}

func NewDataAnalysisSystem(config *AnalysisConfig) (*DataAnalysisSystem, error) {
    // Initialize database
    db, err := sqlx.Connect("postgres", config.DatabaseURL)
    if err != nil {
        return nil, fmt.Errorf("database connection failed: %w", err)
    }

    // Create LLM provider
    llm, err := provider.NewOpenAI(
        provider.WithModel("gpt-4"),
        provider.WithMaxTokens(4000),
    )
    if err != nil {
        return nil, fmt.Errorf("LLM provider creation failed: %w", err)
    }

    // Create specialized agents
    ingestionAgent := core.NewLLMAgent("data-ingestion", llm)
    statisticalAgent := core.NewLLMAgent("statistical-analyzer", llm)
    patternAgent := core.NewLLMAgent("pattern-detector", llm)
    insightAgent := core.NewLLMAgent("insight-generator", llm)
    visualizationAgent := core.NewLLMAgent("visualization-engine", llm)
    predictiveAgent := core.NewLLMAgent("predictive-modeler", llm)
    reportAgent := core.NewLLMAgent("report-synthesizer", llm)

    // Create workflow
    workflowAgent := workflow.NewSequentialAgent("analysis-workflow").
        AddAgent(ingestionAgent).
        AddAgent(statisticalAgent).
        AddAgent(patternAgent).
        AddAgent(insightAgent).
        AddAgent(visualizationAgent).
        AddAgent(predictiveAgent).
        AddAgent(reportAgent)

    // Initialize components
    dataProcessor := NewDataProcessor()
    statisticalEngine := NewStatisticalEngine()
    visualizationEngine := NewVisualizationEngine()

    system := &DataAnalysisSystem{
        db:                  db,
        ingestionAgent:      ingestionAgent,
        statisticalAgent:    statisticalAgent,
        patternAgent:        patternAgent,
        insightAgent:        insightAgent,
        visualizationAgent:  visualizationAgent,
        predictiveAgent:     predictiveAgent,
        reportAgent:         reportAgent,
        workflowAgent:       workflowAgent,
        dataProcessor:       dataProcessor,
        statisticalEngine:   statisticalEngine,
        visualizationEngine: visualizationEngine,
        config:             config,
        metrics:            initializeAnalysisMetrics(),
    }

    // Initialize database schema
    if err := system.initializeSchema(); err != nil {
        return nil, fmt.Errorf("schema initialization failed: %w", err)
    }

    return system, nil
}

func (das *DataAnalysisSystem) initializeSchema() error {
    schema := `
    CREATE TABLE IF NOT EXISTS analysis_projects (
        id SERIAL PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        description TEXT,
        data_sources TEXT[],
        analysis_type VARCHAR(50),
        objectives TEXT[],
        status VARCHAR(50) DEFAULT 'queued',
        requested_by VARCHAR(255),
        priority VARCHAR(20) DEFAULT 'medium',
        scheduled_for TIMESTAMP,
        started_at TIMESTAMP,
        completed_at TIMESTAMP,
        data_point_count INTEGER DEFAULT 0,
        insight_count INTEGER DEFAULT 0,
        anomaly_count INTEGER DEFAULT 0,
        overall_score DECIMAL(3,2) DEFAULT 0.0,
        created_at TIMESTAMP DEFAULT NOW(),
        updated_at TIMESTAMP DEFAULT NOW(),
        metadata JSONB
    );

    CREATE TABLE IF NOT EXISTS datasets (
        id SERIAL PRIMARY KEY,
        project_id INTEGER REFERENCES analysis_projects(id) ON DELETE CASCADE,
        name VARCHAR(255) NOT NULL,
        source TEXT,
        type VARCHAR(50),
        schema JSONB,
        row_count INTEGER DEFAULT 0,
        column_count INTEGER DEFAULT 0,
        quality JSONB,
        processing_status VARCHAR(50) DEFAULT 'pending',
        loaded_at TIMESTAMP DEFAULT NOW(),
        metadata JSONB
    );

    CREATE TABLE IF NOT EXISTS statistical_analyses (
        id SERIAL PRIMARY KEY,
        project_id INTEGER REFERENCES analysis_projects(id) ON DELETE CASCADE,
        dataset_id INTEGER REFERENCES datasets(id) ON DELETE CASCADE,
        type VARCHAR(50),
        method VARCHAR(100),
        variables TEXT[],
        results JSONB,
        interpretation TEXT,
        confidence DECIMAL(3,2),
        p_value DECIMAL(10,8),
        effect_size DECIMAL(6,4),
        created_at TIMESTAMP DEFAULT NOW(),
        metadata JSONB
    );

    CREATE TABLE IF NOT EXISTS patterns (
        id SERIAL PRIMARY KEY,
        project_id INTEGER REFERENCES analysis_projects(id) ON DELETE CASCADE,
        type VARCHAR(50),
        name VARCHAR(255),
        description TEXT,
        variables TEXT[],
        strength DECIMAL(3,2),
        confidence DECIMAL(3,2),
        start_period TIMESTAMP,
        end_period TIMESTAMP,
        parameters JSONB,
        implications TEXT[],
        recommendations TEXT[],
        detected_at TIMESTAMP DEFAULT NOW(),
        metadata JSONB
    );

    CREATE TABLE IF NOT EXISTS insights (
        id SERIAL PRIMARY KEY,
        project_id INTEGER REFERENCES analysis_projects(id) ON DELETE CASCADE,
        type VARCHAR(50),
        title TEXT,
        summary TEXT,
        details TEXT,
        category VARCHAR(100),
        priority VARCHAR(20),
        confidence DECIMAL(3,2),
        impact VARCHAR(20),
        action_required BOOLEAN DEFAULT FALSE,
        supporting_data JSONB,
        related_patterns INTEGER[],
        visualizations TEXT[],
        generated_at TIMESTAMP DEFAULT NOW(),
        metadata JSONB
    );

    CREATE TABLE IF NOT EXISTS visualizations (
        id SERIAL PRIMARY KEY,
        project_id INTEGER REFERENCES analysis_projects(id) ON DELETE CASCADE,
        type VARCHAR(50),
        chart_type VARCHAR(50),
        title TEXT,
        description TEXT,
        data JSONB,
        config JSONB,
        format VARCHAR(20),
        url TEXT,
        created_at TIMESTAMP DEFAULT NOW(),
        metadata JSONB
    );

    CREATE TABLE IF NOT EXISTS predictions (
        id SERIAL PRIMARY KEY,
        project_id INTEGER REFERENCES analysis_projects(id) ON DELETE CASCADE,
        model VARCHAR(100),
        variables TEXT[],
        target_variable VARCHAR(255),
        horizon INTEGER,
        predictions JSONB,
        accuracy JSONB,
        confidence DECIMAL(3,2),
        assumptions TEXT[],
        limitations TEXT[],
        created_at TIMESTAMP DEFAULT NOW(),
        metadata JSONB
    );

    CREATE TABLE IF NOT EXISTS analysis_reports (
        id SERIAL PRIMARY KEY,
        project_id INTEGER REFERENCES analysis_projects(id) ON DELETE CASCADE,
        title TEXT,
        executive_summary TEXT,
        key_findings TEXT[],
        recommendations TEXT[],
        methodology TEXT,
        data_summary TEXT,
        statistical_summary TEXT,
        visualization_summary TEXT,
        conclusions TEXT,
        limitations TEXT[],
        next_steps TEXT[],
        format VARCHAR(20) DEFAULT 'markdown',
        content TEXT,
        generated_at TIMESTAMP DEFAULT NOW(),
        metadata JSONB
    );

    CREATE INDEX IF NOT EXISTS idx_projects_status ON analysis_projects(status);
    CREATE INDEX IF NOT EXISTS idx_datasets_project ON datasets(project_id);
    CREATE INDEX IF NOT EXISTS idx_analyses_project ON statistical_analyses(project_id);
    CREATE INDEX IF NOT EXISTS idx_patterns_project ON patterns(project_id);
    CREATE INDEX IF NOT EXISTS idx_insights_project ON insights(project_id);
    `

    _, err := das.db.Exec(schema)
    return err
}

// Main analysis workflow
func (das *DataAnalysisSystem) AnalyzeData(ctx context.Context, request AnalysisRequest) (*AnalysisProject, error) {
    start := time.Now()
    das.metrics.ProjectsProcessed.Inc()

    // Create analysis project
    project := &AnalysisProject{
        Name:        request.Name,
        Description: request.Description,
        DataSources: request.DataSources,
        AnalysisType: request.AnalysisType,
        Objectives:  request.Objectives,
        Status:      "processing",
        RequestedBy: request.RequestedBy,
        Priority:    request.Priority,
        StartedAt:   &start,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
        Metadata:    request.Metadata,
    }

    // Store project
    if err := das.storeProject(ctx, project); err != nil {
        return nil, fmt.Errorf("failed to store project: %w", err)
    }

    // Phase 1: Data ingestion and quality assessment
    datasets, err := das.ingestData(ctx, project, request.DataSources)
    if err != nil {
        log.Printf("Data ingestion failed: %v", err)
        project.Status = "failed"
        das.updateProject(ctx, project)
        return nil, err
    }

    totalDataPoints := 0
    for _, dataset := range datasets {
        totalDataPoints += dataset.RowCount
    }
    project.DataPointCount = totalDataPoints
    das.updateProject(ctx, project)

    // Phase 2: Statistical analysis
    analyses, err := das.performStatisticalAnalysis(ctx, project.ID, datasets)
    if err != nil {
        log.Printf("Statistical analysis failed: %v", err)
        analyses = []StatisticalAnalysis{}
    }

    // Phase 3: Pattern detection
    patterns, err := das.detectPatterns(ctx, project.ID, datasets, analyses)
    if err != nil {
        log.Printf("Pattern detection failed: %v", err)
        patterns = []Pattern{}
    }

    // Phase 4: Insight generation
    insights, err := das.generateInsights(ctx, project, datasets, analyses, patterns)
    if err != nil {
        log.Printf("Insight generation failed: %v", err)
        insights = []Insight{}
    }

    project.InsightCount = len(insights)
    project.AnomalyCount = das.countAnomalies(patterns)

    // Phase 5: Create visualizations
    visualizations, err := das.createVisualizations(ctx, project.ID, datasets, analyses, patterns)
    if err != nil {
        log.Printf("Visualization creation failed: %v", err)
        visualizations = []Visualization{}
    }

    // Phase 6: Predictive modeling (if requested)
    var predictions []Prediction
    if request.IncludePredictions {
        predictions, err = das.createPredictions(ctx, project.ID, datasets)
        if err != nil {
            log.Printf("Prediction creation failed: %v", err)
            predictions = []Prediction{}
        }
    }

    // Phase 7: Generate comprehensive report
    report, err := das.generateReport(ctx, project, datasets, analyses, patterns, insights, visualizations, predictions)
    if err != nil {
        log.Printf("Report generation failed: %v", err)
    }

    // Calculate overall project score
    project.OverallScore = das.calculateOverallScore(datasets, analyses, insights, patterns)

    // Complete project
    project.Status = "completed"
    now := time.Now()
    project.CompletedAt = &now
    das.updateProject(ctx, project)

    // Record metrics
    das.metrics.ProcessingTime.Observe(time.Since(start).Seconds())
    das.metrics.DataPointsAnalyzed.Add(float64(totalDataPoints))
    das.metrics.InsightsGenerated.Add(float64(len(insights)))
    das.metrics.PatternsDetected.Add(float64(len(patterns)))

    return project, nil
}

type AnalysisRequest struct {
    Name                string                 `json:"name" binding:"required"`
    Description         string                 `json:"description"`
    DataSources         []string               `json:"data_sources" binding:"required"`
    AnalysisType        string                 `json:"analysis_type" binding:"required"`
    Objectives          []string               `json:"objectives"`
    RequestedBy         string                 `json:"requested_by"`
    Priority            string                 `json:"priority"`
    IncludePredictions  bool                   `json:"include_predictions"`
    VisualizationTypes  []string               `json:"visualization_types"`
    StatisticalMethods  []string               `json:"statistical_methods"`
    Options             AnalysisOptions        `json:"options"`
    Metadata            map[string]interface{} `json:"metadata"`
}

type AnalysisOptions struct {
    CleanData           bool     `json:"clean_data"`
    DetectAnomalies     bool     `json:"detect_anomalies"`
    IncludeCorrelations bool     `json:"include_correlations"`
    GenerateInsights    bool     `json:"generate_insights"`
    CreateDashboard     bool     `json:"create_dashboard"`
    AutoVisualize       bool     `json:"auto_visualize"`
    ConfidenceLevel     float64  `json:"confidence_level"`
    SignificanceLevel   float64  `json:"significance_level"`
}

func (das *DataAnalysisSystem) ingestData(ctx context.Context, project *AnalysisProject, sources []string) ([]Dataset, error) {
    var datasets []Dataset

    for _, source := range sources {
        dataset, err := das.loadDataSource(ctx, project.ID, source)
        if err != nil {
            log.Printf("Failed to load data source %s: %v", source, err)
            continue
        }

        // Assess data quality
        quality, err := das.assessDataQuality(ctx, dataset)
        if err != nil {
            log.Printf("Data quality assessment failed: %v", err)
            quality = &DataQuality{OverallScore: 0.5}
        }
        dataset.Quality = *quality

        // Store dataset
        if err := das.storeDataset(ctx, &dataset); err != nil {
            log.Printf("Failed to store dataset: %v", err)
            continue
        }

        datasets = append(datasets, dataset)
    }

    return datasets, nil
}

func (das *DataAnalysisSystem) loadDataSource(ctx context.Context, projectID int, source string) (Dataset, error) {
    // For demonstration, we'll simulate loading data
    // In production, this would handle CSV, JSON, SQL, API sources
    
    dataset := Dataset{
        ProjectID:        projectID,
        Name:            fmt.Sprintf("Dataset_%d", projectID),
        Source:          source,
        Type:            "csv",
        RowCount:        1000,
        ColumnCount:     10,
        ProcessingStatus: "processed",
        LoadedAt:        time.Now(),
        Schema: DataSchema{
            Columns: []ColumnInfo{
                {
                    Name:        "revenue",
                    Type:        "numeric",
                    DataType:    "float",
                    NullCount:   0,
                    UniqueCount: 800,
                    Min:         1000.0,
                    Max:         50000.0,
                    Mean:        func() *float64 { v := 25000.0; return &v }(),
                    StdDev:      func() *float64 { v := 8500.0; return &v }(),
                },
                {
                    Name:        "category",
                    Type:        "categorical",
                    DataType:    "string",
                    NullCount:   0,
                    UniqueCount: 5,
                    Sample:      []interface{}{"A", "B", "C", "D", "E"},
                },
                {
                    Name:        "date",
                    Type:        "datetime",
                    DataType:    "timestamp",
                    NullCount:   0,
                    UniqueCount: 365,
                },
            },
        },
    }

    return dataset, nil
}

func (das *DataAnalysisSystem) assessDataQuality(ctx context.Context, dataset Dataset) (*DataQuality, error) {
    prompt := fmt.Sprintf(`Assess the data quality of this dataset:

Dataset: %s
Source: %s
Rows: %d
Columns: %d

Schema:
%s

Evaluate:
1. Completeness (missing values, null counts)
2. Accuracy (data validity, format consistency)
3. Consistency (standardization across records)
4. Validity (adherence to business rules)

Identify specific issues and provide recommendations for improvement.

Return assessment as JSON:
{
    "overall_score": 0.0-1.0,
    "completeness": 0.0-1.0,
    "accuracy": 0.0-1.0,
    "consistency": 0.0-1.0,
    "validity": 0.0-1.0,
    "issues": [
        {
            "type": "missing_values|outliers|inconsistency|format_error",
            "column": "column_name",
            "severity": "low|medium|high|critical",
            "count": 0,
            "description": "issue description",
            "suggestion": "how to fix"
        }
    ],
    "recommendations": ["rec1", "rec2"]
}`,
        dataset.Name, dataset.Source, dataset.RowCount, dataset.ColumnCount,
        formatSchema(dataset.Schema))

    state := domain.NewState()
    state.AddMessage(domain.NewTextMessage(domain.RoleUser, prompt))

    result, err := das.ingestionAgent.Run(ctx, state)
    if err != nil {
        return nil, err
    }

    var quality DataQuality
    if len(result.Messages) > 0 {
        response := result.Messages[len(result.Messages)-1].TextContent()
        if err := json.Unmarshal([]byte(response), &quality); err != nil {
            return nil, fmt.Errorf("failed to parse quality assessment: %w", err)
        }
    }

    return &quality, nil
}

func (das *DataAnalysisSystem) performStatisticalAnalysis(ctx context.Context, projectID int, datasets []Dataset) ([]StatisticalAnalysis, error) {
    var analyses []StatisticalAnalysis

    for _, dataset := range datasets {
        // Generate analysis plan based on data types
        plan, err := das.generateAnalysisPlan(ctx, dataset)
        if err != nil {
            log.Printf("Failed to generate analysis plan: %v", err)
            continue
        }

        // Execute each analysis in the plan
        for _, method := range plan.Methods {
            analysis, err := das.executeStatisticalMethod(ctx, projectID, dataset.ID, method)
            if err != nil {
                log.Printf("Statistical method %s failed: %v", method, err)
                continue
            }

            // Store analysis
            if err := das.storeStatisticalAnalysis(ctx, &analysis); err != nil {
                log.Printf("Failed to store analysis: %v", err)
                continue
            }

            analyses = append(analyses, analysis)
        }
    }

    return analyses, nil
}

type AnalysisPlan struct {
    Methods   []string `json:"methods"`
    Variables []string `json:"variables"`
    Tests     []string `json:"tests"`
}

func (das *DataAnalysisSystem) generateAnalysisPlan(ctx context.Context, dataset Dataset) (*AnalysisPlan, error) {
    prompt := fmt.Sprintf(`Generate a comprehensive statistical analysis plan for this dataset:

Dataset: %s
Schema: %s

Based on the data types and structure, recommend:
1. Descriptive statistics methods
2. Correlation analyses
3. Appropriate statistical tests
4. Variables to analyze together
5. Potential relationships to explore

Return plan as JSON:
{
    "methods": ["descriptive", "correlation", "regression", "anova"],
    "variables": ["var1", "var2", "var3"],
    "tests": ["t_test", "chi_square", "normality_test"]
}`,
        dataset.Name, formatSchema(dataset.Schema))

    state := domain.NewState()
    state.AddMessage(domain.NewTextMessage(domain.RoleUser, prompt))

    result, err := das.statisticalAgent.Run(ctx, state)
    if err != nil {
        return nil, err
    }

    var plan AnalysisPlan
    if len(result.Messages) > 0 {
        response := result.Messages[len(result.Messages)-1].TextContent()
        if err := json.Unmarshal([]byte(response), &plan); err != nil {
            return nil, fmt.Errorf("failed to parse analysis plan: %w", err)
        }
    }

    return &plan, nil
}

func (das *DataAnalysisSystem) executeStatisticalMethod(ctx context.Context, projectID, datasetID int, method string) (StatisticalAnalysis, error) {
    // This would execute actual statistical computations
    // For demonstration, we'll create mock results
    
    results := StatisticalResults{
        Summary: map[string]interface{}{
            "mean":   25000.0,
            "median": 24500.0,
            "std":    8500.0,
            "min":    1000.0,
            "max":    50000.0,
        },
        Distribution: map[string]interface{}{
            "skewness": 0.2,
            "kurtosis": -0.5,
            "normality_test": map[string]interface{}{
                "statistic": 0.95,
                "p_value":   0.12,
                "normal":    true,
            },
        },
    }

    interpretation, err := das.interpretStatisticalResults(ctx, method, results)
    if err != nil {
        log.Printf("Result interpretation failed: %v", err)
        interpretation = fmt.Sprintf("Statistical analysis using %s method completed", method)
    }

    analysis := StatisticalAnalysis{
        ProjectID:      projectID,
        DatasetID:      datasetID,
        Type:           "descriptive",
        Method:         method,
        Variables:      []string{"revenue", "category"},
        Results:        results,
        Interpretation: interpretation,
        Confidence:     0.95,
        CreatedAt:      time.Now(),
    }

    return analysis, nil
}

func (das *DataAnalysisSystem) interpretStatisticalResults(ctx context.Context, method string, results StatisticalResults) (string, error) {
    prompt := fmt.Sprintf(`Interpret these statistical analysis results:

Method: %s
Results: %s

Provide a clear, business-friendly interpretation that explains:
1. What the results mean
2. Key insights and implications
3. Statistical significance
4. Practical significance
5. Limitations and caveats

Write for a business audience, avoiding excessive technical jargon.`,
        method, formatResults(results))

    state := domain.NewState()
    state.AddMessage(domain.NewTextMessage(domain.RoleUser, prompt))

    result, err := das.statisticalAgent.Run(ctx, state)
    if err != nil {
        return "", err
    }

    if len(result.Messages) == 0 {
        return "", fmt.Errorf("no interpretation generated")
    }

    return result.Messages[len(result.Messages)-1].TextContent(), nil
}

func (das *DataAnalysisSystem) detectPatterns(ctx context.Context, projectID int, datasets []Dataset, analyses []StatisticalAnalysis) ([]Pattern, error) {
    var patterns []Pattern

    // Combine data from all sources for pattern detection
    combinedData := das.combineDatasets(datasets)
    
    // Detect different types of patterns
    patternTypes := []string{"trend", "seasonality", "anomaly", "correlation"}
    
    for _, patternType := range patternTypes {
        detectedPatterns, err := das.detectPatternType(ctx, projectID, combinedData, analyses, patternType)
        if err != nil {
            log.Printf("Pattern detection failed for type %s: %v", patternType, err)
            continue
        }
        
        patterns = append(patterns, detectedPatterns...)
    }

    // Store patterns
    for i := range patterns {
        if err := das.storePattern(ctx, &patterns[i]); err != nil {
            log.Printf("Failed to store pattern: %v", err)
        }
    }

    return patterns, nil
}

func (das *DataAnalysisSystem) detectPatternType(ctx context.Context, projectID int, data interface{}, analyses []StatisticalAnalysis, patternType string) ([]Pattern, error) {
    prompt := fmt.Sprintf(`Analyze this data for %s patterns:

Data Summary: %v
Statistical Analyses: %s

For %s patterns, identify:
1. Pattern strength and confidence
2. Variables involved
3. Time periods (if applicable)
4. Statistical evidence
5. Business implications
6. Actionable recommendations

Return patterns as JSON array:
[
    {
        "type": "%s",
        "name": "pattern name",
        "description": "detailed description",
        "variables": ["var1", "var2"],
        "strength": 0.0-1.0,
        "confidence": 0.0-1.0,
        "start_period": "2024-01-01T00:00:00Z",
        "end_period": "2024-12-31T23:59:59Z",
        "parameters": {"param1": "value1"},
        "implications": ["implication1", "implication2"],
        "recommendations": ["rec1", "rec2"]
    }
]`,
        patternType, data, formatAnalyses(analyses), patternType, patternType)

    state := domain.NewState()
    state.AddMessage(domain.NewTextMessage(domain.RoleUser, prompt))

    result, err := das.patternAgent.Run(ctx, state)
    if err != nil {
        return nil, err
    }

    var patterns []Pattern
    if len(result.Messages) > 0 {
        response := result.Messages[len(result.Messages)-1].TextContent()
        
        var rawPatterns []map[string]interface{}
        if err := json.Unmarshal([]byte(response), &rawPatterns); err != nil {
            return nil, fmt.Errorf("failed to parse patterns: %w", err)
        }

        // Convert to structured patterns
        for _, rawPattern := range rawPatterns {
            pattern := Pattern{
                ProjectID:   projectID,
                Type:        getString(rawPattern, "type"),
                Name:        getString(rawPattern, "name"),
                Description: getString(rawPattern, "description"),
                Strength:    getFloat64(rawPattern, "strength"),
                Confidence:  getFloat64(rawPattern, "confidence"),
                DetectedAt:  time.Now(),
            }

            // Parse variables
            if varsData, ok := rawPattern["variables"].([]interface{}); ok {
                for _, varData := range varsData {
                    if varStr, ok := varData.(string); ok {
                        pattern.Variables = append(pattern.Variables, varStr)
                    }
                }
            }

            // Parse implications and recommendations
            if implData, ok := rawPattern["implications"].([]interface{}); ok {
                for _, impl := range implData {
                    if implStr, ok := impl.(string); ok {
                        pattern.Implications = append(pattern.Implications, implStr)
                    }
                }
            }

            if recData, ok := rawPattern["recommendations"].([]interface{}); ok {
                for _, rec := range recData {
                    if recStr, ok := rec.(string); ok {
                        pattern.Recommendations = append(pattern.Recommendations, recStr)
                    }
                }
            }

            patterns = append(patterns, pattern)
        }
    }

    return patterns, nil
}

func (das *DataAnalysisSystem) generateInsights(ctx context.Context, project *AnalysisProject, datasets []Dataset, analyses []StatisticalAnalysis, patterns []Pattern) ([]Insight, error) {
    prompt := fmt.Sprintf(`Generate actionable business insights from this analysis:

Project: %s
Objectives: %s
Analysis Type: %s

Data Summary:
%s

Statistical Results:
%s

Detected Patterns:
%s

Generate insights that:
1. Answer the project objectives
2. Provide actionable recommendations
3. Identify critical findings
4. Highlight opportunities and risks
5. Suggest next steps

Return insights as JSON array:
[
    {
        "type": "observation|hypothesis|recommendation|alert",
        "title": "insight title",
        "summary": "brief summary",
        "details": "detailed explanation",
        "category": "performance|risk|opportunity|efficiency",
        "priority": "low|medium|high|critical",
        "confidence": 0.0-1.0,
        "impact": "low|medium|high",
        "action_required": true/false,
        "supporting_data": [
            {
                "type": "statistic|pattern|correlation|trend",
                "description": "evidence description",
                "value": "evidence value",
                "confidence": 0.0-1.0,
                "source": "data source"
            }
        ],
        "related_patterns": [0, 1],
        "visualizations": ["chart_type1", "chart_type2"]
    }
]`,
        project.Name, strings.Join(project.Objectives, ", "), project.AnalysisType,
        formatDatasets(datasets), formatAnalyses(analyses), formatPatterns(patterns))

    state := domain.NewState()
    state.AddMessage(domain.NewTextMessage(domain.RoleUser, prompt))

    result, err := das.insightAgent.Run(ctx, state)
    if err != nil {
        return nil, err
    }

    var insights []Insight
    if len(result.Messages) > 0 {
        response := result.Messages[len(result.Messages)-1].TextContent()
        
        var rawInsights []map[string]interface{}
        if err := json.Unmarshal([]byte(response), &rawInsights); err != nil {
            return nil, fmt.Errorf("failed to parse insights: %w", err)
        }

        // Convert to structured insights
        for _, rawInsight := range rawInsights {
            insight := Insight{
                ProjectID:      project.ID,
                Type:           getString(rawInsight, "type"),
                Title:          getString(rawInsight, "title"),
                Summary:        getString(rawInsight, "summary"),
                Details:        getString(rawInsight, "details"),
                Category:       getString(rawInsight, "category"),
                Priority:       getString(rawInsight, "priority"),
                Confidence:     getFloat64(rawInsight, "confidence"),
                Impact:         getString(rawInsight, "impact"),
                ActionRequired: getBool(rawInsight, "action_required"),
                GeneratedAt:    time.Now(),
            }

            // Parse supporting data
            if supportData, ok := rawInsight["supporting_data"].([]interface{}); ok {
                for _, suppData := range supportData {
                    if suppMap, ok := suppData.(map[string]interface{}); ok {
                        evidence := EvidenceItem{
                            Type:        getString(suppMap, "type"),
                            Description: getString(suppMap, "description"),
                            Value:       suppMap["value"],
                            Confidence:  getFloat64(suppMap, "confidence"),
                            Source:      getString(suppMap, "source"),
                        }
                        insight.SupportingData = append(insight.SupportingData, evidence)
                    }
                }
            }

            insights = append(insights, insight)
        }
    }

    // Store insights
    for i := range insights {
        if err := das.storeInsight(ctx, &insights[i]); err != nil {
            log.Printf("Failed to store insight: %v", err)
        }
    }

    return insights, nil
}

func (das *DataAnalysisSystem) createVisualizations(ctx context.Context, projectID int, datasets []Dataset, analyses []StatisticalAnalysis, patterns []Pattern) ([]Visualization, error) {
    var visualizations []Visualization

    // Generate visualization plan
    vizPlan, err := das.generateVisualizationPlan(ctx, datasets, analyses, patterns)
    if err != nil {
        log.Printf("Visualization planning failed: %v", err)
        return visualizations, nil
    }

    // Create each visualization
    for _, vizType := range vizPlan.ChartTypes {
        viz, err := das.createVisualization(ctx, projectID, vizType, datasets, analyses)
        if err != nil {
            log.Printf("Visualization creation failed for %s: %v", vizType, err)
            continue
        }

        // Store visualization
        if err := das.storeVisualization(ctx, &viz); err != nil {
            log.Printf("Failed to store visualization: %v", err)
            continue
        }

        visualizations = append(visualizations, viz)
    }

    return visualizations, nil
}

type VisualizationPlan struct {
    ChartTypes  []string `json:"chart_types"`
    Variables   []string `json:"variables"`
    Layouts     []string `json:"layouts"`
}

func (das *DataAnalysisSystem) generateVisualizationPlan(ctx context.Context, datasets []Dataset, analyses []StatisticalAnalysis, patterns []Pattern) (*VisualizationPlan, error) {
    prompt := fmt.Sprintf(`Generate a visualization plan for this data analysis:

Datasets: %s
Statistical Analyses: %s
Patterns: %s

Recommend visualizations that:
1. Show key statistical findings
2. Highlight detected patterns
3. Support business decision-making
4. Are appropriate for the data types
5. Tell a coherent story

Return plan as JSON:
{
    "chart_types": ["bar", "line", "scatter", "heatmap", "histogram"],
    "variables": ["var1", "var2", "var3"],
    "layouts": ["single", "dashboard", "comparison"]
}`,
        formatDatasets(datasets), formatAnalyses(analyses), formatPatterns(patterns))

    state := domain.NewState()
    state.AddMessage(domain.NewTextMessage(domain.RoleUser, prompt))

    result, err := das.visualizationAgent.Run(ctx, state)
    if err != nil {
        return nil, err
    }

    var plan VisualizationPlan
    if len(result.Messages) > 0 {
        response := result.Messages[len(result.Messages)-1].TextContent()
        if err := json.Unmarshal([]byte(response), &plan); err != nil {
            return nil, fmt.Errorf("failed to parse visualization plan: %w", err)
        }
    }

    return &plan, nil
}

func (das *DataAnalysisSystem) createVisualization(ctx context.Context, projectID int, chartType string, datasets []Dataset, analyses []StatisticalAnalysis) (Visualization, error) {
    // Generate mock visualization data
    vizData := map[string]interface{}{
        "data": []map[string]interface{}{
            {"x": "Category A", "y": 120},
            {"x": "Category B", "y": 190},
            {"x": "Category C", "y": 300},
            {"x": "Category D", "y": 500},
            {"x": "Category E", "y": 200},
        },
        "title": fmt.Sprintf("%s Chart", strings.Title(chartType)),
        "xAxis": "Categories",
        "yAxis": "Values",
    }

    visualization := Visualization{
        ProjectID:   projectID,
        Type:        "chart",
        ChartType:   chartType,
        Title:       fmt.Sprintf("%s Analysis Chart", strings.Title(chartType)),
        Description: fmt.Sprintf("Analysis visualization showing %s data", chartType),
        Data:        vizData,
        Config: VisualizationConfig{
            Width:       800,
            Height:      600,
            Theme:       "default",
            Colors:      []string{"#1f77b4", "#ff7f0e", "#2ca02c", "#d62728", "#9467bd"},
            Interactive: true,
            Animations:  true,
        },
        Format:    "json",
        CreatedAt: time.Now(),
    }

    return visualization, nil
}

func (das *DataAnalysisSystem) generateReport(ctx context.Context, project *AnalysisProject, datasets []Dataset, analyses []StatisticalAnalysis, patterns []Pattern, insights []Insight, visualizations []Visualization, predictions []Prediction) (*AnalysisReport, error) {
    prompt := fmt.Sprintf(`Generate a comprehensive data analysis report:

Project: %s
Objectives: %s
Analysis Type: %s

Data Summary: %s
Statistical Results: %s
Patterns Found: %s
Key Insights: %s
Visualizations: %d created
Predictions: %d models

Create a professional report including:
1. Executive Summary
2. Data Overview and Quality Assessment
3. Statistical Analysis Results
4. Pattern Analysis and Findings
5. Key Insights and Recommendations
6. Visualizations Summary
7. Predictive Analysis (if applicable)
8. Conclusions and Next Steps
9. Limitations and Assumptions

Write for business stakeholders with clear, actionable insights.`,
        project.Name, strings.Join(project.Objectives, ", "), project.AnalysisType,
        formatDatasets(datasets), formatAnalyses(analyses), formatPatterns(patterns),
        formatInsights(insights), len(visualizations), len(predictions))

    state := domain.NewState()
    state.AddMessage(domain.NewTextMessage(domain.RoleUser, prompt))

    result, err := das.reportAgent.Run(ctx, state)
    if err != nil {
        return nil, err
    }

    var reportContent string
    if len(result.Messages) > 0 {
        reportContent = result.Messages[len(result.Messages)-1].TextContent()
    }

    // Extract key findings and recommendations
    keyFindings := das.extractKeyFindings(insights)
    recommendations := das.extractRecommendations(insights, patterns)

    report := &AnalysisReport{
        ProjectID:        project.ID,
        Title:            fmt.Sprintf("Data Analysis Report: %s", project.Name),
        ExecutiveSummary: das.extractExecutiveSummary(reportContent),
        KeyFindings:      keyFindings,
        Recommendations:  recommendations,
        Content:          reportContent,
        Format:           "markdown",
        GeneratedAt:      time.Now(),
    }

    // Store report
    if err := das.storeReport(ctx, report); err != nil {
        log.Printf("Failed to store report: %v", err)
    }

    return report, nil
}

// Supporting components and data processors
type DataProcessor struct{}

func NewDataProcessor() *DataProcessor {
    return &DataProcessor{}
}

type StatisticalEngine struct{}

func NewStatisticalEngine() *StatisticalEngine {
    return &StatisticalEngine{}
}

type VisualizationEngine struct{}

func NewVisualizationEngine() *VisualizationEngine {
    return &VisualizationEngine{}
}

// Database operations
func (das *DataAnalysisSystem) storeProject(ctx context.Context, project *AnalysisProject) error {
    query := `INSERT INTO analysis_projects 
              (name, description, data_sources, analysis_type, objectives, status, 
               requested_by, priority, started_at, data_point_count, metadata)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
              RETURNING id, created_at, updated_at`

    metadataJSON, _ := json.Marshal(project.Metadata)
    
    err := das.db.QueryRowContext(ctx, query,
        project.Name, project.Description, pq.Array(project.DataSources),
        project.AnalysisType, pq.Array(project.Objectives), project.Status,
        project.RequestedBy, project.Priority, project.StartedAt,
        project.DataPointCount, metadataJSON,
    ).Scan(&project.ID, &project.CreatedAt, &project.UpdatedAt)

    return err
}

func (das *DataAnalysisSystem) updateProject(ctx context.Context, project *AnalysisProject) error {
    query := `UPDATE analysis_projects 
              SET status = $1, data_point_count = $2, insight_count = $3, 
                  anomaly_count = $4, overall_score = $5, completed_at = $6, updated_at = $7
              WHERE id = $8`

    _, err := das.db.ExecContext(ctx, query,
        project.Status, project.DataPointCount, project.InsightCount,
        project.AnomalyCount, project.OverallScore, project.CompletedAt,
        time.Now(), project.ID)

    return err
}

func (das *DataAnalysisSystem) storeDataset(ctx context.Context, dataset *Dataset) error {
    query := `INSERT INTO datasets 
              (project_id, name, source, type, schema, row_count, column_count, 
               quality, processing_status, metadata)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
              RETURNING id, loaded_at`

    schemaJSON, _ := json.Marshal(dataset.Schema)
    qualityJSON, _ := json.Marshal(dataset.Quality)
    metadataJSON, _ := json.Marshal(dataset.Metadata)
    
    err := das.db.QueryRowContext(ctx, query,
        dataset.ProjectID, dataset.Name, dataset.Source, dataset.Type,
        schemaJSON, dataset.RowCount, dataset.ColumnCount, qualityJSON,
        dataset.ProcessingStatus, metadataJSON,
    ).Scan(&dataset.ID, &dataset.LoadedAt)

    return err
}

func (das *DataAnalysisSystem) storeStatisticalAnalysis(ctx context.Context, analysis *StatisticalAnalysis) error {
    query := `INSERT INTO statistical_analyses 
              (project_id, dataset_id, type, method, variables, results, 
               interpretation, confidence, p_value, effect_size, metadata)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
              RETURNING id, created_at`

    resultsJSON, _ := json.Marshal(analysis.Results)
    metadataJSON, _ := json.Marshal(analysis.Metadata)
    
    err := das.db.QueryRowContext(ctx, query,
        analysis.ProjectID, analysis.DatasetID, analysis.Type, analysis.Method,
        pq.Array(analysis.Variables), resultsJSON, analysis.Interpretation,
        analysis.Confidence, analysis.PValue, analysis.EffectSize, metadataJSON,
    ).Scan(&analysis.ID, &analysis.CreatedAt)

    return err
}

func (das *DataAnalysisSystem) storePattern(ctx context.Context, pattern *Pattern) error {
    query := `INSERT INTO patterns 
              (project_id, type, name, description, variables, strength, confidence,
               start_period, end_period, parameters, implications, recommendations, metadata)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
              RETURNING id, detected_at`

    parametersJSON, _ := json.Marshal(pattern.Parameters)
    metadataJSON, _ := json.Marshal(pattern.Metadata)
    
    err := das.db.QueryRowContext(ctx, query,
        pattern.ProjectID, pattern.Type, pattern.Name, pattern.Description,
        pq.Array(pattern.Variables), pattern.Strength, pattern.Confidence,
        pattern.StartPeriod, pattern.EndPeriod, parametersJSON,
        pq.Array(pattern.Implications), pq.Array(pattern.Recommendations), metadataJSON,
    ).Scan(&pattern.ID, &pattern.DetectedAt)

    return err
}

func (das *DataAnalysisSystem) storeInsight(ctx context.Context, insight *Insight) error {
    query := `INSERT INTO insights 
              (project_id, type, title, summary, details, category, priority, 
               confidence, impact, action_required, supporting_data, related_patterns,
               visualizations, metadata)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
              RETURNING id, generated_at`

    supportingDataJSON, _ := json.Marshal(insight.SupportingData)
    metadataJSON, _ := json.Marshal(insight.Metadata)
    
    err := das.db.QueryRowContext(ctx, query,
        insight.ProjectID, insight.Type, insight.Title, insight.Summary,
        insight.Details, insight.Category, insight.Priority, insight.Confidence,
        insight.Impact, insight.ActionRequired, supportingDataJSON,
        pq.Array(insight.RelatedPatterns), pq.Array(insight.Visualizations), metadataJSON,
    ).Scan(&insight.ID, &insight.GeneratedAt)

    return err
}

func (das *DataAnalysisSystem) storeVisualization(ctx context.Context, viz *Visualization) error {
    query := `INSERT INTO visualizations 
              (project_id, type, chart_type, title, description, data, config, 
               format, url, metadata)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
              RETURNING id, created_at`

    dataJSON, _ := json.Marshal(viz.Data)
    configJSON, _ := json.Marshal(viz.Config)
    metadataJSON, _ := json.Marshal(viz.Metadata)
    
    err := das.db.QueryRowContext(ctx, query,
        viz.ProjectID, viz.Type, viz.ChartType, viz.Title,
        viz.Description, dataJSON, configJSON, viz.Format,
        viz.URL, metadataJSON,
    ).Scan(&viz.ID, &viz.CreatedAt)

    return err
}

func (das *DataAnalysisSystem) storeReport(ctx context.Context, report *AnalysisReport) error {
    query := `INSERT INTO analysis_reports 
              (project_id, title, executive_summary, key_findings, recommendations,
               methodology, data_summary, statistical_summary, visualization_summary,
               conclusions, limitations, next_steps, format, content, metadata)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
              RETURNING id, generated_at`

    metadataJSON, _ := json.Marshal(report.Metadata)
    
    err := das.db.QueryRowContext(ctx, query,
        report.ProjectID, report.Title, report.ExecutiveSummary,
        pq.Array(report.KeyFindings), pq.Array(report.Recommendations),
        report.Methodology, report.DataSummary, report.StatisticalSummary,
        report.VisualizationSummary, report.Conclusions,
        pq.Array(report.Limitations), pq.Array(report.NextSteps),
        report.Format, report.Content, metadataJSON,
    ).Scan(&report.ID, &report.GeneratedAt)

    return err
}

// HTTP API handlers
func (das *DataAnalysisSystem) StartAnalysisHandler(c *gin.Context) {
    var request AnalysisRequest
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Start analysis asynchronously
    go func() {
        ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
        defer cancel()

        project, err := das.AnalyzeData(ctx, request)
        if err != nil {
            log.Printf("Analysis failed: %v", err)
        } else {
            log.Printf("Analysis completed: %d", project.ID)
        }
    }()

    c.JSON(http.StatusAccepted, gin.H{
        "message": "Analysis started",
        "status":  "processing",
    })
}

func (das *DataAnalysisSystem) GetProjectHandler(c *gin.Context) {
    projectID, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
        return
    }

    var project AnalysisProject
    query := `SELECT * FROM analysis_projects WHERE id = $1`
    
    err = das.db.GetContext(c.Request.Context(), &project, query, projectID)
    if err != nil {
        if err == sql.ErrNoRows {
            c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"project": project})
}

func (das *DataAnalysisSystem) GetInsightsHandler(c *gin.Context) {
    projectID, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
        return
    }

    var insights []Insight
    query := `SELECT * FROM insights WHERE project_id = $1 ORDER BY priority DESC, confidence DESC`
    
    err = das.db.SelectContext(c.Request.Context(), &insights, query, projectID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"insights": insights})
}

func (das *DataAnalysisSystem) GetReportHandler(c *gin.Context) {
    projectID, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
        return
    }

    var report AnalysisReport
    query := `SELECT * FROM analysis_reports WHERE project_id = $1 ORDER BY generated_at DESC LIMIT 1`
    
    err = das.db.GetContext(c.Request.Context(), &report, query, projectID)
    if err != nil {
        if err == sql.ErrNoRows {
            c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"report": report})
}

// Utility functions
func (das *DataAnalysisSystem) combineDatasets(datasets []Dataset) interface{} {
    // Simplified combination for demonstration
    return map[string]interface{}{
        "total_rows":    das.calculateTotalRows(datasets),
        "total_columns": das.calculateTotalColumns(datasets),
        "data_types":    das.extractDataTypes(datasets),
    }
}

func (das *DataAnalysisSystem) calculateTotalRows(datasets []Dataset) int {
    total := 0
    for _, dataset := range datasets {
        total += dataset.RowCount
    }
    return total
}

func (das *DataAnalysisSystem) calculateTotalColumns(datasets []Dataset) int {
    total := 0
    for _, dataset := range datasets {
        total += dataset.ColumnCount
    }
    return total
}

func (das *DataAnalysisSystem) extractDataTypes(datasets []Dataset) []string {
    typeSet := make(map[string]bool)
    for _, dataset := range datasets {
        for _, column := range dataset.Schema.Columns {
            typeSet[column.Type] = true
        }
    }
    
    var types []string
    for t := range typeSet {
        types = append(types, t)
    }
    return types
}

func (das *DataAnalysisSystem) countAnomalies(patterns []Pattern) int {
    count := 0
    for _, pattern := range patterns {
        if pattern.Type == "anomaly" {
            count++
        }
    }
    return count
}

func (das *DataAnalysisSystem) calculateOverallScore(datasets []Dataset, analyses []StatisticalAnalysis, insights []Insight, patterns []Pattern) float64 {
    // Simplified scoring algorithm
    dataQualityScore := das.calculateAverageDataQuality(datasets)
    insightQualityScore := das.calculateAverageInsightConfidence(insights)
    patternStrengthScore := das.calculateAveragePatternStrength(patterns)
    
    return (dataQualityScore + insightQualityScore + patternStrengthScore) / 3.0
}

func (das *DataAnalysisSystem) calculateAverageDataQuality(datasets []Dataset) float64 {
    if len(datasets) == 0 {
        return 0.0
    }
    
    total := 0.0
    for _, dataset := range datasets {
        total += dataset.Quality.OverallScore
    }
    return total / float64(len(datasets))
}

func (das *DataAnalysisSystem) calculateAverageInsightConfidence(insights []Insight) float64 {
    if len(insights) == 0 {
        return 0.0
    }
    
    total := 0.0
    for _, insight := range insights {
        total += insight.Confidence
    }
    return total / float64(len(insights))
}

func (das *DataAnalysisSystem) calculateAveragePatternStrength(patterns []Pattern) float64 {
    if len(patterns) == 0 {
        return 0.0
    }
    
    total := 0.0
    for _, pattern := range patterns {
        total += pattern.Strength
    }
    return total / float64(len(patterns))
}

func (das *DataAnalysisSystem) createPredictions(ctx context.Context, projectID int, datasets []Dataset) ([]Prediction, error) {
    // Simplified prediction creation for demonstration
    var predictions []Prediction
    
    for _, dataset := range datasets {
        // Create mock prediction
        prediction := Prediction{
            ProjectID:      projectID,
            Model:          "linear_regression",
            Variables:      []string{"time", "value"},
            TargetVariable: "revenue",
            Horizon:        12,
            Confidence:     0.85,
            CreatedAt:      time.Now(),
        }
        
        // Generate mock prediction points
        for i := 1; i <= 12; i++ {
            point := PredictionPoint{
                Period:     time.Now().AddDate(0, i, 0),
                Value:      25000.0 + float64(i)*1000.0,
                LowerBound: 20000.0 + float64(i)*800.0,
                UpperBound: 30000.0 + float64(i)*1200.0,
                Confidence: 0.85,
            }
            prediction.Predictions = append(prediction.Predictions, point)
        }
        
        predictions = append(predictions, prediction)
    }
    
    return predictions, nil
}

func (das *DataAnalysisSystem) extractKeyFindings(insights []Insight) []string {
    var findings []string
    for _, insight := range insights {
        if insight.Priority == "high" || insight.Priority == "critical" {
            findings = append(findings, insight.Summary)
        }
    }
    return findings
}

func (das *DataAnalysisSystem) extractRecommendations(insights []Insight, patterns []Pattern) []string {
    var recommendations []string
    
    // Extract from insights
    for _, insight := range insights {
        if insight.ActionRequired {
            recommendations = append(recommendations, insight.Details)
        }
    }
    
    // Extract from patterns
    for _, pattern := range patterns {
        recommendations = append(recommendations, pattern.Recommendations...)
    }
    
    return recommendations
}

func (das *DataAnalysisSystem) extractExecutiveSummary(content string) string {
    lines := strings.Split(content, "\n")
    for i, line := range lines {
        if strings.Contains(strings.ToLower(line), "executive summary") && i+1 < len(lines) {
            end := i + 10
            if end > len(lines) {
                end = len(lines)
            }
            return strings.Join(lines[i+1:end], "\n")
        }
    }
    return "Executive summary not available"
}

// Formatting functions for prompts
func formatSchema(schema DataSchema) string {
    var formatted strings.Builder
    for _, col := range schema.Columns {
        formatted.WriteString(fmt.Sprintf("- %s (%s): nulls=%d, unique=%d\n", 
            col.Name, col.Type, col.NullCount, col.UniqueCount))
    }
    return formatted.String()
}

func formatResults(results StatisticalResults) string {
    resultsJSON, _ := json.Marshal(results)
    return string(resultsJSON)
}

func formatAnalyses(analyses []StatisticalAnalysis) string {
    var formatted strings.Builder
    for _, analysis := range analyses {
        formatted.WriteString(fmt.Sprintf("- %s (%s): confidence=%.2f\n", 
            analysis.Method, analysis.Type, analysis.Confidence))
    }
    return formatted.String()
}

func formatPatterns(patterns []Pattern) string {
    var formatted strings.Builder
    for _, pattern := range patterns {
        formatted.WriteString(fmt.Sprintf("- %s (%s): strength=%.2f, confidence=%.2f\n", 
            pattern.Name, pattern.Type, pattern.Strength, pattern.Confidence))
    }
    return formatted.String()
}

func formatDatasets(datasets []Dataset) string {
    var formatted strings.Builder
    for _, dataset := range datasets {
        formatted.WriteString(fmt.Sprintf("- %s: %d rows, %d columns, quality=%.2f\n", 
            dataset.Name, dataset.RowCount, dataset.ColumnCount, dataset.Quality.OverallScore))
    }
    return formatted.String()
}

func formatInsights(insights []Insight) string {
    var formatted strings.Builder
    for _, insight := range insights {
        formatted.WriteString(fmt.Sprintf("- %s (%s): %s\n", 
            insight.Title, insight.Priority, insight.Summary))
    }
    return formatted.String()
}

// Utility functions for JSON parsing
func getString(m map[string]interface{}, key string) string {
    if val, ok := m[key].(string); ok {
        return val
    }
    return ""
}

func getFloat64(m map[string]interface{}, key string) float64 {
    if val, ok := m[key].(float64); ok {
        return val
    }
    return 0.0
}

func getBool(m map[string]interface{}, key string) bool {
    if val, ok := m[key].(bool); ok {
        return val
    }
    return false
}

func initializeAnalysisMetrics() *AnalysisMetrics {
    metrics := &AnalysisMetrics{
        ProjectsProcessed: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "analysis_projects_processed_total",
            Help: "Total number of analysis projects processed",
        }),
        DataPointsAnalyzed: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "analysis_data_points_analyzed_total",
            Help: "Total number of data points analyzed",
        }),
        InsightsGenerated: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "analysis_insights_generated_total",
            Help: "Total number of insights generated",
        }),
        PatternsDetected: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "analysis_patterns_detected_total",
            Help: "Total number of patterns detected",
        }),
        PredictionsCreated: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "analysis_predictions_created_total",
            Help: "Total number of predictions created",
        }),
        ProcessingTime: prometheus.NewHistogram(prometheus.HistogramOpts{
            Name: "analysis_processing_time_seconds",
            Help: "Time taken to complete analysis projects",
        }),
        AccuracyScores: prometheus.NewHistogram(prometheus.HistogramOpts{
            Name: "analysis_accuracy_scores",
            Help: "Distribution of analysis accuracy scores",
        }),
        DataQualityScores: prometheus.NewHistogram(prometheus.HistogramOpts{
            Name: "analysis_data_quality_scores",
            Help: "Distribution of data quality scores",
        }),
    }

    prometheus.MustRegister(
        metrics.ProjectsProcessed,
        metrics.DataPointsAnalyzed,
        metrics.InsightsGenerated,
        metrics.PatternsDetected,
        metrics.PredictionsCreated,
        metrics.ProcessingTime,
        metrics.AccuracyScores,
        metrics.DataQualityScores,
    )

    return metrics
}

func main() {
    config := &AnalysisConfig{
        DatabaseURL:         "postgres://user:pass@localhost/analysis_db?sslmode=disable",
        MaxDataPoints:       1000000,
        StatisticalMethods:  []string{"descriptive", "correlation", "regression", "anova"},
        VisualizationTypes:  []string{"bar", "line", "scatter", "heatmap", "histogram"},
        PredictiveModels:    []string{"linear_regression", "arima", "prophet"},
        ConfidenceThreshold: 0.8,
        AnomalyThreshold:    0.05,
        TrendThreshold:      0.7,
        ReportFormats:       []string{"markdown", "html", "pdf"},
        RealTimeEnabled:     true,
        CacheEnabled:        true,
    }

    system, err := NewDataAnalysisSystem(config)
    if err != nil {
        log.Fatal("Failed to create analysis system:", err)
    }

    // Setup HTTP server
    r := gin.Default()

    // API routes
    api := r.Group("/api/v1")
    {
        api.POST("/analyze", system.StartAnalysisHandler)
        api.GET("/projects/:id", system.GetProjectHandler)
        api.GET("/projects/:id/insights", system.GetInsightsHandler)
        api.GET("/projects/:id/report", system.GetReportHandler)
        api.GET("/metrics", gin.WrapH(promhttp.Handler()))
    }

    // Health check
    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "status": "healthy",
            "max_data_points": config.MaxDataPoints,
            "methods_available": len(config.StatisticalMethods),
        })
    })

    log.Println("Data Analysis System starting on :8080")
    log.Fatal(r.Run(":8080"))
}
```

## Usage Examples

### Start Data Analysis

```bash
curl -X POST http://localhost:8080/api/v1/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Sales Performance Analysis",
    "description": "Comprehensive analysis of sales data to identify trends and opportunities",
    "data_sources": ["sales_data.csv", "customer_data.json"],
    "analysis_type": "descriptive",
    "objectives": [
      "Identify sales trends and patterns",
      "Analyze customer behavior",
      "Detect performance anomalies",
      "Generate actionable insights"
    ],
    "requested_by": "analyst@company.com",
    "priority": "high",
    "include_predictions": true,
    "visualization_types": ["bar", "line", "heatmap"],
    "statistical_methods": ["descriptive", "correlation", "regression"],
    "options": {
      "clean_data": true,
      "detect_anomalies": true,
      "include_correlations": true,
      "generate_insights": true,
      "create_dashboard": true,
      "confidence_level": 0.95
    }
  }'
```

### Get Analysis Results

```bash
curl http://localhost:8080/api/v1/projects/123
```

### Get Generated Insights

```bash
curl http://localhost:8080/api/v1/projects/123/insights
```

### Get Analysis Report

```bash
curl http://localhost:8080/api/v1/projects/123/report
```

## Key Features Demonstrated

1. **Automated Data Processing** - Quality assessment and cleaning
2. **Statistical Analysis** - Comprehensive statistical computations with AI interpretation
3. **Pattern Recognition** - Trend, anomaly, and correlation detection
4. **Natural Language Insights** - AI-generated explanations of findings
5. **Visualization Generation** - Automatic chart and dashboard creation
6. **Predictive Modeling** - Forecasting and scenario analysis
7. **Comprehensive Reporting** - Executive summaries and technical reports
8. **Performance Monitoring** - System metrics and processing analytics

## Production Considerations

- **Data Security** - Encryption and access controls for sensitive data
- **Scalability** - Distributed processing for large datasets
- **Real-time Processing** - Streaming data analysis capabilities
- **Model Management** - Version control for statistical models
- **Quality Assurance** - Validation of analysis results
- **Integration** - APIs for business intelligence platforms

This data analysis system demonstrates how to build a sophisticated AI-powered analytics platform using Go-LLMs, showcasing the integration of statistical computing, pattern recognition, and natural language generation for comprehensive business intelligence.