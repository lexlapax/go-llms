# Code Analysis: Code Review Systems

> **[Project Root](/) / [Documentation](../..) / [User Guide](../../user-guide) / [Examples](../../user-guide/examples) / Code Analysis**

Build an AI-powered code review and analysis system that automates code quality assessment, security vulnerability detection, performance optimization suggestions, and documentation generation. This example demonstrates how to integrate Go-LLMs with software development workflows.

## System Overview

This code analysis system provides:

- **Automated Code Reviews** - Quality, style, and best practice analysis
- **Security Vulnerability Detection** - Common security issues and remediation
- **Performance Analysis** - Optimization opportunities and bottlenecks
- **Documentation Generation** - Automatic code documentation and comments
- **Technical Debt Assessment** - Code maintainability and refactoring suggestions
- **Git Integration** - Pull request analysis and commit validation
- **Multi-Language Support** - Go, Python, JavaScript, Java, and more
- **Team Analytics** - Code quality metrics and developer insights

## Architecture

![Code Analysis System Architecture](../../images/code-analysis-architecture.svg)

### Components
1. **Code Ingestion** - Repository scanning and file processing
2. **Static Analysis** - AST parsing and pattern detection
3. **AI Review Agent** - Intelligent code quality assessment
4. **Security Scanner** - Vulnerability detection and classification
5. **Performance Analyzer** - Optimization opportunity identification
6. **Documentation Generator** - Automatic comment and doc generation
7. **Report Engine** - Comprehensive analysis reporting
8. **Integration Layer** - Git, CI/CD, and IDE integration

---

## Complete Implementation

```go
package main

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "go/ast"
    "go/parser"
    "go/token"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "regexp"
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

// CodeAnalysisSystem is the main system orchestrator
type CodeAnalysisSystem struct {
    db                  *sqlx.DB
    reviewAgent         *core.LLMAgent
    securityAgent       *core.LLMAgent
    performanceAgent    *core.LLMAgent
    documentationAgent  *core.LLMAgent
    workflowAgent       *workflow.ParallelAgent
    analyzers          map[string]LanguageAnalyzer
    config             *AnalysisConfig
    metrics            *AnalysisMetrics
}

type AnalysisConfig struct {
    DatabaseURL         string                 `json:"database_url"`
    OpenAIKey          string                 `json:"openai_key"`
    SupportedLanguages []string               `json:"supported_languages"`
    MaxFileSize        int64                  `json:"max_file_size"`
    SecurityRules      []SecurityRule         `json:"security_rules"`
    QualityRules       []QualityRule          `json:"quality_rules"`
    PerformanceRules   []PerformanceRule      `json:"performance_rules"`
    ExcludePatterns    []string               `json:"exclude_patterns"`
    IncludeTests       bool                   `json:"include_tests"`
    GenerateDocs       bool                   `json:"generate_docs"`
}

// Analysis models
type CodeAnalysis struct {
    ID              int                    `json:"id" db:"id"`
    RepositoryURL   string                 `json:"repository_url" db:"repository_url"`
    CommitHash      string                 `json:"commit_hash" db:"commit_hash"`
    Branch          string                 `json:"branch" db:"branch"`
    RequestedBy     string                 `json:"requested_by" db:"requested_by"`
    Status          string                 `json:"status" db:"status"` // pending, analyzing, completed, failed
    TotalFiles      int                    `json:"total_files" db:"total_files"`
    ProcessedFiles  int                    `json:"processed_files" db:"processed_files"`
    OverallScore    float64                `json:"overall_score" db:"overall_score"`
    QualityScore    float64                `json:"quality_score" db:"quality_score"`
    SecurityScore   float64                `json:"security_score" db:"security_score"`
    PerformanceScore float64               `json:"performance_score" db:"performance_score"`
    IssueCount      int                    `json:"issue_count" db:"issue_count"`
    CriticalIssues  int                    `json:"critical_issues" db:"critical_issues"`
    WarningCount    int                    `json:"warning_count" db:"warning_count"`
    StartedAt       time.Time              `json:"started_at" db:"started_at"`
    CompletedAt     *time.Time             `json:"completed_at" db:"completed_at"`
    Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
}

type FileAnalysis struct {
    ID                int                    `json:"id" db:"id"`
    AnalysisID        int                    `json:"analysis_id" db:"analysis_id"`
    FilePath          string                 `json:"file_path" db:"file_path"`
    Language          string                 `json:"language" db:"language"`
    LineCount         int                    `json:"line_count" db:"line_count"`
    ComplexityScore   float64                `json:"complexity_score" db:"complexity_score"`
    QualityScore      float64                `json:"quality_score" db:"quality_score"`
    SecurityScore     float64                `json:"security_score" db:"security_score"`
    PerformanceScore  float64                `json:"performance_score" db:"performance_score"`
    Issues            []Issue                `json:"issues" db:"issues"`
    Suggestions       []Suggestion           `json:"suggestions" db:"suggestions"`
    GeneratedDocs     string                 `json:"generated_docs" db:"generated_docs"`
    ProcessedAt       time.Time              `json:"processed_at" db:"processed_at"`
    Metadata          map[string]interface{} `json:"metadata" db:"metadata"`
}

type Issue struct {
    Type        string  `json:"type"`         // quality, security, performance, style
    Severity    string  `json:"severity"`     // critical, high, medium, low, info
    Line        int     `json:"line"`
    Column      int     `json:"column"`
    Message     string  `json:"message"`
    Rule        string  `json:"rule"`
    Category    string  `json:"category"`
    Confidence  float64 `json:"confidence"`
    Suggestion  string  `json:"suggestion"`
    CodeSnippet string  `json:"code_snippet"`
}

type Suggestion struct {
    Type        string  `json:"type"`         // refactor, optimize, document, test
    Priority    string  `json:"priority"`     // high, medium, low
    Line        int     `json:"line"`
    Message     string  `json:"message"`
    Rationale   string  `json:"rationale"`
    Example     string  `json:"example"`
    Confidence  float64 `json:"confidence"`
}

type SecurityRule struct {
    ID          string   `json:"id"`
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Pattern     string   `json:"pattern"`
    Languages   []string `json:"languages"`
    Severity    string   `json:"severity"`
    Category    string   `json:"category"`
}

type QualityRule struct {
    ID          string   `json:"id"`
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Languages   []string `json:"languages"`
    Threshold   float64  `json:"threshold"`
    Metric      string   `json:"metric"`
}

type PerformanceRule struct {
    ID          string   `json:"id"`
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Pattern     string   `json:"pattern"`
    Languages   []string `json:"languages"`
    Impact      string   `json:"impact"`
}

// Language-specific analyzers
type LanguageAnalyzer interface {
    Name() string
    SupportedExtensions() []string
    ParseFile(filePath string, content []byte) (*ParsedFile, error)
    AnalyzeComplexity(parsed *ParsedFile) float64
    DetectPatterns(parsed *ParsedFile, rules []SecurityRule) []Issue
    GenerateDocumentation(parsed *ParsedFile) string
}

type ParsedFile struct {
    Path        string                 `json:"path"`
    Language    string                 `json:"language"`
    Functions   []Function             `json:"functions"`
    Classes     []Class                `json:"classes"`
    Imports     []string               `json:"imports"`
    Comments    []Comment              `json:"comments"`
    Metrics     FileMetrics            `json:"metrics"`
    AST         interface{}            `json:"-"` // Language-specific AST
    Metadata    map[string]interface{} `json:"metadata"`
}

type Function struct {
    Name        string   `json:"name"`
    Line        int      `json:"line"`
    Parameters  []string `json:"parameters"`
    ReturnType  string   `json:"return_type"`
    Complexity  int      `json:"complexity"`
    LineCount   int      `json:"line_count"`
    Comments    []string `json:"comments"`
    IsPublic    bool     `json:"is_public"`
    IsTest      bool     `json:"is_test"`
}

type Class struct {
    Name      string     `json:"name"`
    Line      int        `json:"line"`
    Methods   []Function `json:"methods"`
    Fields    []Field    `json:"fields"`
    IsPublic  bool       `json:"is_public"`
    Comments  []string   `json:"comments"`
}

type Field struct {
    Name     string `json:"name"`
    Type     string `json:"type"`
    Line     int    `json:"line"`
    IsPublic bool   `json:"is_public"`
}

type Comment struct {
    Line    int    `json:"line"`
    Content string `json:"content"`
    Type    string `json:"type"` // single, block, doc
}

type FileMetrics struct {
    TotalLines      int     `json:"total_lines"`
    CodeLines       int     `json:"code_lines"`
    CommentLines    int     `json:"comment_lines"`
    BlankLines      int     `json:"blank_lines"`
    Functions       int     `json:"functions"`
    Classes         int     `json:"classes"`
    Complexity      int     `json:"complexity"`
    Maintainability float64 `json:"maintainability"`
    Duplication     float64 `json:"duplication"`
}

type AnalysisMetrics struct {
    AnalysesStarted    prometheus.Counter
    AnalysesCompleted  prometheus.Counter
    FilesProcessed     prometheus.Counter
    IssuesDetected     *prometheus.CounterVec
    ProcessingTime     prometheus.Histogram
    QualityScores      prometheus.Histogram
    SecurityScores     prometheus.Histogram
}

func NewCodeAnalysisSystem(config *AnalysisConfig) (*CodeAnalysisSystem, error) {
    // Initialize database
    db, err := sqlx.Connect("postgres", config.DatabaseURL)
    if err != nil {
        return nil, fmt.Errorf("database connection failed: %w", err)
    }

    // Create LLM provider
    llm, err := provider.NewOpenAIProvider(
        provider.WithModel("gpt-4"),
        provider.WithMaxTokens(3000),
    )
    if err != nil {
        return nil, fmt.Errorf("LLM provider creation failed: %w", err)
    }

    // Create specialized agents
    reviewAgent := core.NewLLMAgent("code-reviewer", llm)
    securityAgent := core.NewLLMAgent("security-analyzer", llm)
    performanceAgent := core.NewLLMAgent("performance-analyzer", llm)
    documentationAgent := core.NewLLMAgent("documentation-generator", llm)

    // Create parallel workflow for concurrent analysis
    workflowAgent := workflow.NewParallelAgent("analysis-workflow").
        WithMaxConcurrency(4).
        AddAgent(reviewAgent).
        AddAgent(securityAgent).
        AddAgent(performanceAgent).
        AddAgent(documentationAgent)

    // Initialize language analyzers
    analyzers := make(map[string]LanguageAnalyzer)
    analyzers["go"] = NewGoAnalyzer()
    analyzers["python"] = NewPythonAnalyzer()
    analyzers["javascript"] = NewJavaScriptAnalyzer()
    analyzers["java"] = NewJavaAnalyzer()

    system := &CodeAnalysisSystem{
        db:                 db,
        reviewAgent:        reviewAgent,
        securityAgent:      securityAgent,
        performanceAgent:   performanceAgent,
        documentationAgent: documentationAgent,
        workflowAgent:      workflowAgent,
        analyzers:         analyzers,
        config:            config,
        metrics:           initializeAnalysisMetrics(),
    }

    // Initialize database schema
    if err := system.initializeSchema(); err != nil {
        return nil, fmt.Errorf("schema initialization failed: %w", err)
    }

    return system, nil
}

func (cas *CodeAnalysisSystem) initializeSchema() error {
    schema := `
    CREATE TABLE IF NOT EXISTS code_analyses (
        id SERIAL PRIMARY KEY,
        repository_url TEXT NOT NULL,
        commit_hash VARCHAR(40),
        branch VARCHAR(255),
        requested_by VARCHAR(255),
        status VARCHAR(50) DEFAULT 'pending',
        total_files INTEGER DEFAULT 0,
        processed_files INTEGER DEFAULT 0,
        overall_score DECIMAL(3,2) DEFAULT 0.0,
        quality_score DECIMAL(3,2) DEFAULT 0.0,
        security_score DECIMAL(3,2) DEFAULT 0.0,
        performance_score DECIMAL(3,2) DEFAULT 0.0,
        issue_count INTEGER DEFAULT 0,
        critical_issues INTEGER DEFAULT 0,
        warning_count INTEGER DEFAULT 0,
        started_at TIMESTAMP DEFAULT NOW(),
        completed_at TIMESTAMP,
        metadata JSONB
    );

    CREATE TABLE IF NOT EXISTS file_analyses (
        id SERIAL PRIMARY KEY,
        analysis_id INTEGER REFERENCES code_analyses(id) ON DELETE CASCADE,
        file_path TEXT NOT NULL,
        language VARCHAR(50),
        line_count INTEGER DEFAULT 0,
        complexity_score DECIMAL(3,2) DEFAULT 0.0,
        quality_score DECIMAL(3,2) DEFAULT 0.0,
        security_score DECIMAL(3,2) DEFAULT 0.0,
        performance_score DECIMAL(3,2) DEFAULT 0.0,
        issues JSONB,
        suggestions JSONB,
        generated_docs TEXT,
        processed_at TIMESTAMP DEFAULT NOW(),
        metadata JSONB
    );

    CREATE INDEX IF NOT EXISTS idx_analyses_status ON code_analyses(status);
    CREATE INDEX IF NOT EXISTS idx_analyses_requested_by ON code_analyses(requested_by);
    CREATE INDEX IF NOT EXISTS idx_file_analyses_analysis_id ON file_analyses(analysis_id);
    CREATE INDEX IF NOT EXISTS idx_file_analyses_language ON file_analyses(language);
    `

    _, err := cas.db.Exec(schema)
    return err
}

// Main analysis workflow
func (cas *CodeAnalysisSystem) AnalyzeRepository(ctx context.Context, request AnalysisRequest) (*CodeAnalysis, error) {
    start := time.Now()
    cas.metrics.AnalysesStarted.Inc()

    // Create analysis record
    analysis := &CodeAnalysis{
        RepositoryURL: request.RepositoryURL,
        CommitHash:   request.CommitHash,
        Branch:       request.Branch,
        RequestedBy:  request.RequestedBy,
        Status:       "analyzing",
        StartedAt:    time.Now(),
        Metadata:     request.Metadata,
    }

    // Store initial analysis
    if err := cas.storeAnalysis(ctx, analysis); err != nil {
        return nil, fmt.Errorf("failed to store analysis: %w", err)
    }

    // Clone or access repository
    repoPath, err := cas.prepareRepository(ctx, request)
    if err != nil {
        analysis.Status = "failed"
        cas.updateAnalysis(ctx, analysis)
        return nil, fmt.Errorf("repository preparation failed: %w", err)
    }

    // Discover files to analyze
    files, err := cas.discoverFiles(repoPath)
    if err != nil {
        analysis.Status = "failed"
        cas.updateAnalysis(ctx, analysis)
        return nil, fmt.Errorf("file discovery failed: %w", err)
    }

    analysis.TotalFiles = len(files)
    cas.updateAnalysis(ctx, analysis)

    // Analyze files
    var totalQuality, totalSecurity, totalPerformance float64
    var totalIssues, criticalIssues, warnings int

    for _, file := range files {
        fileAnalysis, err := cas.analyzeFile(ctx, analysis.ID, file)
        if err != nil {
            log.Printf("Failed to analyze file %s: %v", file, err)
            continue
        }

        cas.metrics.FilesProcessed.Inc()
        analysis.ProcessedFiles++

        // Aggregate scores
        totalQuality += fileAnalysis.QualityScore
        totalSecurity += fileAnalysis.SecurityScore
        totalPerformance += fileAnalysis.PerformanceScore

        // Count issues
        for _, issue := range fileAnalysis.Issues {
            totalIssues++
            if issue.Severity == "critical" {
                criticalIssues++
                cas.metrics.IssuesDetected.WithLabelValues("critical", issue.Type).Inc()
            } else if issue.Severity == "high" || issue.Severity == "medium" {
                warnings++
                cas.metrics.IssuesDetected.WithLabelValues(issue.Severity, issue.Type).Inc()
            }
        }

        // Update progress
        if analysis.ProcessedFiles%10 == 0 {
            cas.updateAnalysis(ctx, analysis)
        }
    }

    // Calculate final scores
    if analysis.ProcessedFiles > 0 {
        analysis.QualityScore = totalQuality / float64(analysis.ProcessedFiles)
        analysis.SecurityScore = totalSecurity / float64(analysis.ProcessedFiles)
        analysis.PerformanceScore = totalPerformance / float64(analysis.ProcessedFiles)
        analysis.OverallScore = (analysis.QualityScore + analysis.SecurityScore + analysis.PerformanceScore) / 3.0
    }

    analysis.IssueCount = totalIssues
    analysis.CriticalIssues = criticalIssues
    analysis.WarningCount = warnings
    analysis.Status = "completed"
    now := time.Now()
    analysis.CompletedAt = &now

    // Final update
    if err := cas.updateAnalysis(ctx, analysis); err != nil {
        log.Printf("Failed to update final analysis: %v", err)
    }

    // Record metrics
    cas.metrics.AnalysesCompleted.Inc()
    cas.metrics.ProcessingTime.Observe(time.Since(start).Seconds())
    cas.metrics.QualityScores.Observe(analysis.QualityScore)
    cas.metrics.SecurityScores.Observe(analysis.SecurityScore)

    return analysis, nil
}

type AnalysisRequest struct {
    RepositoryURL string                 `json:"repository_url" binding:"required"`
    CommitHash    string                 `json:"commit_hash"`
    Branch        string                 `json:"branch"`
    RequestedBy   string                 `json:"requested_by"`
    Options       AnalysisOptions        `json:"options"`
    Metadata      map[string]interface{} `json:"metadata"`
}

type AnalysisOptions struct {
    IncludeTests        bool     `json:"include_tests"`
    GenerateDocs        bool     `json:"generate_docs"`
    SecurityScanOnly    bool     `json:"security_scan_only"`
    Languages          []string  `json:"languages"`
    ExcludePaths       []string  `json:"exclude_paths"`
    MaxFiles           int       `json:"max_files"`
}

func (cas *CodeAnalysisSystem) analyzeFile(ctx context.Context, analysisID int, filePath string) (*FileAnalysis, error) {
    // Read file content
    content, err := os.ReadFile(filePath)
    if err != nil {
        return nil, err
    }

    // Detect language
    language := cas.detectLanguage(filePath)
    analyzer, exists := cas.analyzers[language]
    if !exists {
        return nil, fmt.Errorf("unsupported language: %s", language)
    }

    // Parse file
    parsed, err := analyzer.ParseFile(filePath, content)
    if err != nil {
        return nil, fmt.Errorf("parsing failed: %w", err)
    }

    // Static analysis
    complexityScore := analyzer.AnalyzeComplexity(parsed)
    securityIssues := analyzer.DetectPatterns(parsed, cas.config.SecurityRules)

    // AI-powered analysis
    qualityAnalysis, err := cas.analyzeQuality(ctx, parsed, string(content))
    if err != nil {
        log.Printf("Quality analysis failed: %v", err)
        qualityAnalysis = &QualityAnalysisResult{Score: 0.5}
    }

    performanceAnalysis, err := cas.analyzePerformance(ctx, parsed, string(content))
    if err != nil {
        log.Printf("Performance analysis failed: %v", err)
        performanceAnalysis = &PerformanceAnalysisResult{Score: 0.5}
    }

    // Generate documentation if requested
    var generatedDocs string
    if cas.config.GenerateDocs {
        docs, err := cas.generateDocumentation(ctx, parsed, string(content))
        if err != nil {
            log.Printf("Documentation generation failed: %v", err)
        } else {
            generatedDocs = docs
        }
    }

    // Combine all issues
    allIssues := append(securityIssues, qualityAnalysis.Issues...)
    allIssues = append(allIssues, performanceAnalysis.Issues...)

    // Create file analysis
    fileAnalysis := &FileAnalysis{
        AnalysisID:       analysisID,
        FilePath:         filePath,
        Language:         language,
        LineCount:        parsed.Metrics.TotalLines,
        ComplexityScore:  complexityScore,
        QualityScore:     qualityAnalysis.Score,
        SecurityScore:    cas.calculateSecurityScore(securityIssues),
        PerformanceScore: performanceAnalysis.Score,
        Issues:           allIssues,
        Suggestions:      performanceAnalysis.Suggestions,
        GeneratedDocs:    generatedDocs,
        ProcessedAt:      time.Now(),
        Metadata: map[string]interface{}{
            "functions": len(parsed.Functions),
            "classes":   len(parsed.Classes),
            "imports":   len(parsed.Imports),
        },
    }

    // Store file analysis
    if err := cas.storeFileAnalysis(ctx, fileAnalysis); err != nil {
        log.Printf("Failed to store file analysis: %v", err)
    }

    return fileAnalysis, nil
}

type QualityAnalysisResult struct {
    Score       float64      `json:"score"`
    Issues      []Issue      `json:"issues"`
    Suggestions []Suggestion `json:"suggestions"`
    Metrics     QualityMetrics `json:"metrics"`
}

type PerformanceAnalysisResult struct {
    Score       float64      `json:"score"`
    Issues      []Issue      `json:"issues"`
    Suggestions []Suggestion `json:"suggestions"`
    Hotspots    []PerformanceHotspot `json:"hotspots"`
}

type QualityMetrics struct {
    Maintainability float64 `json:"maintainability"`
    Readability     float64 `json:"readability"`
    Testability     float64 `json:"testability"`
    Reusability     float64 `json:"reusability"`
}

type PerformanceHotspot struct {
    Function    string  `json:"function"`
    Line        int     `json:"line"`
    Type        string  `json:"type"`
    Impact      string  `json:"impact"`
    Confidence  float64 `json:"confidence"`
    Suggestion  string  `json:"suggestion"`
}

func (cas *CodeAnalysisSystem) analyzeQuality(ctx context.Context, parsed *ParsedFile, content string) (*QualityAnalysisResult, error) {
    prompt := fmt.Sprintf(`Analyze the code quality of this %s file and provide detailed feedback:

File: %s
Functions: %d
Classes: %d
Lines: %d

Code:
%s

Evaluate:
1. Code structure and organization
2. Naming conventions and clarity
3. Function/method complexity
4. Error handling practices
5. Code duplication
6. Documentation quality
7. Best practices adherence
8. Maintainability factors

Provide specific line-by-line issues and suggestions for improvement.

Return analysis as JSON:
{
    "score": 0.0-1.0,
    "issues": [
        {
            "type": "quality",
            "severity": "high|medium|low",
            "line": 0,
            "message": "Issue description",
            "rule": "rule_name",
            "category": "naming|structure|complexity|documentation",
            "confidence": 0.0-1.0,
            "suggestion": "How to fix"
        }
    ],
    "suggestions": [
        {
            "type": "refactor",
            "priority": "high|medium|low",
            "line": 0,
            "message": "Suggestion",
            "rationale": "Why this helps",
            "example": "Code example",
            "confidence": 0.0-1.0
        }
    ],
    "metrics": {
        "maintainability": 0.0-1.0,
        "readability": 0.0-1.0,
        "testability": 0.0-1.0,
        "reusability": 0.0-1.0
    }
}`,
        parsed.Language, parsed.Path, len(parsed.Functions), len(parsed.Classes), parsed.Metrics.TotalLines, content)

    state := domain.NewState()
    state.AddMessage(domain.NewTextMessage(domain.RoleUser, prompt))

    result, err := cas.reviewAgent.Run(ctx, state)
    if err != nil {
        return nil, err
    }

    var analysis QualityAnalysisResult
    if len(result.Messages) > 0 {
        response := result.Messages[len(result.Messages)-1].TextContent()
        if err := json.Unmarshal([]byte(response), &analysis); err != nil {
            return nil, fmt.Errorf("failed to parse quality analysis: %w", err)
        }
    }

    return &analysis, nil
}

func (cas *CodeAnalysisSystem) analyzePerformance(ctx context.Context, parsed *ParsedFile, content string) (*PerformanceAnalysisResult, error) {
    prompt := fmt.Sprintf(`Analyze the performance characteristics of this %s code:

File: %s
Code:
%s

Focus on:
1. Algorithm efficiency and complexity
2. Memory usage patterns
3. I/O operations and blocking calls
4. Loop optimization opportunities
5. Data structure choices
6. Caching opportunities
7. Concurrency patterns
8. Resource management

Identify performance hotspots and optimization opportunities.

Return analysis as JSON:
{
    "score": 0.0-1.0,
    "issues": [
        {
            "type": "performance",
            "severity": "high|medium|low",
            "line": 0,
            "message": "Performance issue",
            "rule": "rule_name",
            "category": "algorithm|memory|io|concurrency",
            "confidence": 0.0-1.0,
            "suggestion": "Optimization approach"
        }
    ],
    "suggestions": [
        {
            "type": "optimize",
            "priority": "high|medium|low",
            "line": 0,
            "message": "Optimization opportunity",
            "rationale": "Performance benefit",
            "example": "Optimized code",
            "confidence": 0.0-1.0
        }
    ],
    "hotspots": [
        {
            "function": "function_name",
            "line": 0,
            "type": "cpu|memory|io",
            "impact": "high|medium|low",
            "confidence": 0.0-1.0,
            "suggestion": "Optimization approach"
        }
    ]
}`,
        parsed.Language, parsed.Path, content)

    state := domain.NewState()
    state.AddMessage(domain.NewTextMessage(domain.RoleUser, prompt))

    result, err := cas.performanceAgent.Run(ctx, state)
    if err != nil {
        return nil, err
    }

    var analysis PerformanceAnalysisResult
    if len(result.Messages) > 0 {
        response := result.Messages[len(result.Messages)-1].TextContent()
        if err := json.Unmarshal([]byte(response), &analysis); err != nil {
            return nil, fmt.Errorf("failed to parse performance analysis: %w", err)
        }
    }

    return &analysis, nil
}

func (cas *CodeAnalysisSystem) generateDocumentation(ctx context.Context, parsed *ParsedFile, content string) (string, error) {
    prompt := fmt.Sprintf(`Generate comprehensive documentation for this %s code:

File: %s
Functions: %s
Classes: %s

Code:
%s

Generate:
1. File-level documentation describing purpose and overview
2. Function/method documentation with parameters and return values
3. Class documentation with purpose and usage
4. Usage examples where appropriate
5. Code comments for complex logic

Follow the language's documentation conventions (%s style).

Return the documentation as markdown that can be used in README files or code comments.`,
        parsed.Language, parsed.Path,
        formatFunctions(parsed.Functions), formatClasses(parsed.Classes),
        content, parsed.Language)

    state := domain.NewState()
    state.AddMessage(domain.NewTextMessage(domain.RoleUser, prompt))

    result, err := cas.documentationAgent.Run(ctx, state)
    if err != nil {
        return "", err
    }

    if len(result.Messages) == 0 {
        return "", fmt.Errorf("no documentation generated")
    }

    return result.Messages[len(result.Messages)-1].TextContent(), nil
}

// Go language analyzer implementation
type GoAnalyzer struct{}

func NewGoAnalyzer() *GoAnalyzer {
    return &GoAnalyzer{}
}

func (ga *GoAnalyzer) Name() string {
    return "go"
}

func (ga *GoAnalyzer) SupportedExtensions() []string {
    return []string{".go"}
}

func (ga *GoAnalyzer) ParseFile(filePath string, content []byte) (*ParsedFile, error) {
    fset := token.NewFileSet()
    node, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
    if err != nil {
        return nil, err
    }

    parsed := &ParsedFile{
        Path:      filePath,
        Language:  "go",
        Functions: []Function{},
        Classes:   []Class{}, // Go doesn't have classes, but we'll use for structs
        Imports:   []string{},
        Comments:  []Comment{},
        AST:       node,
        Metadata:  make(map[string]interface{}),
    }

    // Extract imports
    for _, imp := range node.Imports {
        path := strings.Trim(imp.Path.Value, "\"")
        parsed.Imports = append(parsed.Imports, path)
    }

    // Extract functions and methods
    ast.Inspect(node, func(n ast.Node) bool {
        switch x := n.(type) {
        case *ast.FuncDecl:
            if x.Name.IsExported() || strings.HasSuffix(filePath, "_test.go") {
                function := Function{
                    Name:       x.Name.Name,
                    Line:       fset.Position(x.Pos()).Line,
                    Parameters: []string{},
                    IsPublic:   x.Name.IsExported(),
                    IsTest:     strings.HasPrefix(x.Name.Name, "Test"),
                    Complexity: calculateCyclomaticComplexity(x),
                }

                // Extract parameters
                if x.Type.Params != nil {
                    for _, param := range x.Type.Params.List {
                        for _, name := range param.Names {
                            function.Parameters = append(function.Parameters, name.Name)
                        }
                    }
                }

                parsed.Functions = append(parsed.Functions, function)
            }

        case *ast.GenDecl:
            // Handle struct declarations (treated as classes)
            for _, spec := range x.Specs {
                if typeSpec, ok := spec.(*ast.TypeSpec); ok {
                    if structType, ok := typeSpec.Type.(*ast.StructType); ok {
                        class := Class{
                            Name:     typeSpec.Name.Name,
                            Line:     fset.Position(typeSpec.Pos()).Line,
                            Methods:  []Function{},
                            Fields:   []Field{},
                            IsPublic: typeSpec.Name.IsExported(),
                        }

                        // Extract fields
                        for _, field := range structType.Fields.List {
                            for _, name := range field.Names {
                                class.Fields = append(class.Fields, Field{
                                    Name:     name.Name,
                                    Line:     fset.Position(name.Pos()).Line,
                                    IsPublic: name.IsExported(),
}
                            }
                        }

                        parsed.Classes = append(parsed.Classes, class)
                    }
                }
            }
        }
        return true
}

    // Calculate metrics
    lines := strings.Split(string(content), "\n")
    parsed.Metrics = FileMetrics{
        TotalLines:   len(lines),
        CodeLines:    countCodeLines(lines),
        CommentLines: countCommentLines(lines),
        BlankLines:   countBlankLines(lines),
        Functions:    len(parsed.Functions),
        Classes:      len(parsed.Classes),
        Complexity:   calculateTotalComplexity(parsed.Functions),
    }

    return parsed, nil
}

func (ga *GoAnalyzer) AnalyzeComplexity(parsed *ParsedFile) float64 {
    if parsed.Metrics.Functions == 0 {
        return 1.0
    }

    avgComplexity := float64(parsed.Metrics.Complexity) / float64(parsed.Metrics.Functions)
    
    // Normalize to 0-1 scale (higher is better)
    if avgComplexity <= 5 {
        return 1.0
    } else if avgComplexity <= 10 {
        return 0.8
    } else if avgComplexity <= 15 {
        return 0.6
    } else if avgComplexity <= 20 {
        return 0.4
    } else {
        return 0.2
    }
}

func (ga *GoAnalyzer) DetectPatterns(parsed *ParsedFile, rules []SecurityRule) []Issue {
    var issues []Issue
    
    // This is a simplified pattern detection - in production, use more sophisticated AST analysis
    content := fmt.Sprintf("%v", parsed.AST)
    
    for _, rule := range rules {
        if contains(rule.Languages, "go") {
            if matched, _ := regexp.MatchString(rule.Pattern, content); matched {
                issues = append(issues, Issue{
                    Type:     "security",
                    Severity: rule.Severity,
                    Message:  rule.Description,
                    Rule:     rule.ID,
                    Category: rule.Category,
                    Confidence: 0.8,
}
            }
        }
    }
    
    return issues
}

func (ga *GoAnalyzer) GenerateDocumentation(parsed *ParsedFile) string {
    var doc strings.Builder
    
    doc.WriteString(fmt.Sprintf("# %s\n\n", filepath.Base(parsed.Path)))
    doc.WriteString("## Functions\n\n")
    
    for _, fn := range parsed.Functions {
        if fn.IsPublic {
            doc.WriteString(fmt.Sprintf("### %s\n", fn.Name))
            doc.WriteString(fmt.Sprintf("Parameters: %s\n", strings.Join(fn.Parameters, ", ")))
            doc.WriteString(fmt.Sprintf("Complexity: %d\n\n", fn.Complexity))
        }
    }
    
    return doc.String()
}

// Utility functions for Go analysis
func calculateCyclomaticComplexity(fn *ast.FuncDecl) int {
    complexity := 1 // Base complexity
    
    ast.Inspect(fn, func(n ast.Node) bool {
        switch n.(type) {
        case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt, *ast.TypeSwitchStmt:
            complexity++
        case *ast.CaseClause:
            complexity++
        }
        return true
}
    
    return complexity
}

func calculateTotalComplexity(functions []Function) int {
    total := 0
    for _, fn := range functions {
        total += fn.Complexity
    }
    return total
}

func countCodeLines(lines []string) int {
    count := 0
    for _, line := range lines {
        trimmed := strings.TrimSpace(line)
        if trimmed != "" && !strings.HasPrefix(trimmed, "//") && !strings.HasPrefix(trimmed, "/*") {
            count++
        }
    }
    return count
}

func countCommentLines(lines []string) int {
    count := 0
    inBlockComment := false
    
    for _, line := range lines {
        trimmed := strings.TrimSpace(line)
        if strings.HasPrefix(trimmed, "//") {
            count++
        } else if strings.HasPrefix(trimmed, "/*") {
            count++
            inBlockComment = true
        } else if inBlockComment {
            count++
            if strings.Contains(trimmed, "*/") {
                inBlockComment = false
            }
        }
    }
    return count
}

func countBlankLines(lines []string) int {
    count := 0
    for _, line := range lines {
        if strings.TrimSpace(line) == "" {
            count++
        }
    }
    return count
}

// Placeholder implementations for other language analyzers
type PythonAnalyzer struct{}
func NewPythonAnalyzer() *PythonAnalyzer { return &PythonAnalyzer{} }
func (pa *PythonAnalyzer) Name() string { return "python" }
func (pa *PythonAnalyzer) SupportedExtensions() []string { return []string{".py"} }
func (pa *PythonAnalyzer) ParseFile(filePath string, content []byte) (*ParsedFile, error) {
    // Placeholder - would implement Python AST parsing
    return &ParsedFile{Path: filePath, Language: "python"}, nil
}
func (pa *PythonAnalyzer) AnalyzeComplexity(parsed *ParsedFile) float64 { return 0.8 }
func (pa *PythonAnalyzer) DetectPatterns(parsed *ParsedFile, rules []SecurityRule) []Issue { return []Issue{} }
func (pa *PythonAnalyzer) GenerateDocumentation(parsed *ParsedFile) string { return "" }

type JavaScriptAnalyzer struct{}
func NewJavaScriptAnalyzer() *JavaScriptAnalyzer { return &JavaScriptAnalyzer{} }
func (ja *JavaScriptAnalyzer) Name() string { return "javascript" }
func (ja *JavaScriptAnalyzer) SupportedExtensions() []string { return []string{".js", ".jsx", ".ts", ".tsx"} }
func (ja *JavaScriptAnalyzer) ParseFile(filePath string, content []byte) (*ParsedFile, error) {
    return &ParsedFile{Path: filePath, Language: "javascript"}, nil
}
func (ja *JavaScriptAnalyzer) AnalyzeComplexity(parsed *ParsedFile) float64 { return 0.8 }
func (ja *JavaScriptAnalyzer) DetectPatterns(parsed *ParsedFile, rules []SecurityRule) []Issue { return []Issue{} }
func (ja *JavaScriptAnalyzer) GenerateDocumentation(parsed *ParsedFile) string { return "" }

type JavaAnalyzer struct{}
func NewJavaAnalyzer() *JavaAnalyzer { return &JavaAnalyzer{} }
func (ja *JavaAnalyzer) Name() string { return "java" }
func (ja *JavaAnalyzer) SupportedExtensions() []string { return []string{".java"} }
func (ja *JavaAnalyzer) ParseFile(filePath string, content []byte) (*ParsedFile, error) {
    return &ParsedFile{Path: filePath, Language: "java"}, nil
}
func (ja *JavaAnalyzer) AnalyzeComplexity(parsed *ParsedFile) float64 { return 0.8 }
func (ja *JavaAnalyzer) DetectPatterns(parsed *ParsedFile, rules []SecurityRule) []Issue { return []Issue{} }
func (ja *JavaAnalyzer) GenerateDocumentation(parsed *ParsedFile) string { return "" }

// Database operations
func (cas *CodeAnalysisSystem) storeAnalysis(ctx context.Context, analysis *CodeAnalysis) error {
    query := `INSERT INTO code_analyses 
              (repository_url, commit_hash, branch, requested_by, status, total_files, 
               overall_score, quality_score, security_score, performance_score, 
               issue_count, critical_issues, warning_count, started_at, metadata)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
              RETURNING id`

    metadataJSON, _ := json.Marshal(analysis.Metadata)
    
    err := cas.db.QueryRowContext(ctx, query,
        analysis.RepositoryURL, analysis.CommitHash, analysis.Branch, analysis.RequestedBy,
        analysis.Status, analysis.TotalFiles, analysis.OverallScore, analysis.QualityScore,
        analysis.SecurityScore, analysis.PerformanceScore, analysis.IssueCount,
        analysis.CriticalIssues, analysis.WarningCount, analysis.StartedAt, metadataJSON,
    ).Scan(&analysis.ID)

    return err
}

func (cas *CodeAnalysisSystem) updateAnalysis(ctx context.Context, analysis *CodeAnalysis) error {
    query := `UPDATE code_analyses 
              SET status = $1, processed_files = $2, overall_score = $3, quality_score = $4,
                  security_score = $5, performance_score = $6, issue_count = $7, 
                  critical_issues = $8, warning_count = $9, completed_at = $10
              WHERE id = $11`

    _, err := cas.db.ExecContext(ctx, query,
        analysis.Status, analysis.ProcessedFiles, analysis.OverallScore, analysis.QualityScore,
        analysis.SecurityScore, analysis.PerformanceScore, analysis.IssueCount,
        analysis.CriticalIssues, analysis.WarningCount, analysis.CompletedAt, analysis.ID)

    return err
}

func (cas *CodeAnalysisSystem) storeFileAnalysis(ctx context.Context, fileAnalysis *FileAnalysis) error {
    query := `INSERT INTO file_analyses 
              (analysis_id, file_path, language, line_count, complexity_score, quality_score,
               security_score, performance_score, issues, suggestions, generated_docs, metadata)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
              RETURNING id`

    issuesJSON, _ := json.Marshal(fileAnalysis.Issues)
    suggestionsJSON, _ := json.Marshal(fileAnalysis.Suggestions)
    metadataJSON, _ := json.Marshal(fileAnalysis.Metadata)
    
    err := cas.db.QueryRowContext(ctx, query,
        fileAnalysis.AnalysisID, fileAnalysis.FilePath, fileAnalysis.Language,
        fileAnalysis.LineCount, fileAnalysis.ComplexityScore, fileAnalysis.QualityScore,
        fileAnalysis.SecurityScore, fileAnalysis.PerformanceScore, issuesJSON,
        suggestionsJSON, fileAnalysis.GeneratedDocs, metadataJSON,
    ).Scan(&fileAnalysis.ID)

    return err
}

// HTTP API handlers
func (cas *CodeAnalysisSystem) StartAnalysisHandler(c *gin.Context) {
    var request AnalysisRequest
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Start analysis asynchronously
    go func() {
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
        defer cancel()

        analysis, err := cas.AnalyzeRepository(ctx, request)
        if err != nil {
            log.Printf("Analysis failed: %v", err)
        } else {
            log.Printf("Analysis completed: %d", analysis.ID)
        }
    }()

    c.JSON(http.StatusAccepted, gin.H{
        "message": "Analysis started",
        "status":  "analyzing",
}
}

func (cas *CodeAnalysisSystem) GetAnalysisHandler(c *gin.Context) {
    analysisID, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid analysis ID"})
        return
    }

    var analysis CodeAnalysis
    query := `SELECT * FROM code_analyses WHERE id = $1`
    
    err = cas.db.GetContext(c.Request.Context(), &analysis, query, analysisID)
    if err != nil {
        if err == sql.ErrNoRows {
            c.JSON(http.StatusNotFound, gin.H{"error": "Analysis not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"analysis": analysis})
}

func (cas *CodeAnalysisSystem) GetFileAnalysesHandler(c *gin.Context) {
    analysisID, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid analysis ID"})
        return
    }

    var fileAnalyses []FileAnalysis
    query := `SELECT * FROM file_analyses WHERE analysis_id = $1 ORDER BY file_path`
    
    err = cas.db.SelectContext(c.Request.Context(), &fileAnalyses, query, analysisID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"file_analyses": fileAnalyses})
}

// Utility functions
func (cas *CodeAnalysisSystem) prepareRepository(ctx context.Context, request AnalysisRequest) (string, error) {
    // In a real implementation, this would clone the repository
    // For this example, we'll assume the repository is already available locally
    return "/tmp/repo", nil
}

func (cas *CodeAnalysisSystem) discoverFiles(repoPath string) ([]string, error) {
    var files []string
    
    err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        
        if info.IsDir() {
            return nil
        }
        
        // Check if file should be analyzed
        if cas.shouldAnalyzeFile(path) {
            files = append(files, path)
        }
        
        return nil
}
    
    return files, err
}

func (cas *CodeAnalysisSystem) shouldAnalyzeFile(filePath string) bool {
    ext := filepath.Ext(filePath)
    
    // Check supported extensions
    for _, analyzer := range cas.analyzers {
        for _, supportedExt := range analyzer.SupportedExtensions() {
            if ext == supportedExt {
                return true
            }
        }
    }
    
    return false
}

func (cas *CodeAnalysisSystem) detectLanguage(filePath string) string {
    ext := filepath.Ext(filePath)
    
    languageMap := map[string]string{
        ".go":   "go",
        ".py":   "python",
        ".js":   "javascript",
        ".jsx":  "javascript",
        ".ts":   "javascript",
        ".tsx":  "javascript",
        ".java": "java",
    }
    
    if lang, exists := languageMap[ext]; exists {
        return lang
    }
    
    return "unknown"
}

func (cas *CodeAnalysisSystem) calculateSecurityScore(issues []Issue) float64 {
    if len(issues) == 0 {
        return 1.0
    }
    
    criticalCount := 0
    highCount := 0
    
    for _, issue := range issues {
        switch issue.Severity {
        case "critical":
            criticalCount++
        case "high":
            highCount++
        }
    }
    
    // Simple scoring algorithm
    score := 1.0 - (float64(criticalCount)*0.3 + float64(highCount)*0.2)
    if score < 0 {
        score = 0
    }
    
    return score
}

func formatFunctions(functions []Function) string {
    var names []string
    for _, fn := range functions {
        names = append(names, fn.Name)
    }
    return strings.Join(names, ", ")
}

func formatClasses(classes []Class) string {
    var names []string
    for _, class := range classes {
        names = append(names, class.Name)
    }
    return strings.Join(names, ", ")
}

func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}

func initializeAnalysisMetrics() *AnalysisMetrics {
    metrics := &AnalysisMetrics{
        AnalysesStarted: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "code_analyses_started_total",
            Help: "Total number of code analyses started",
        }),
        AnalysesCompleted: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "code_analyses_completed_total",
            Help: "Total number of code analyses completed",
        }),
        FilesProcessed: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "code_files_processed_total",
            Help: "Total number of code files processed",
        }),
        IssuesDetected: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "code_issues_detected_total",
                Help: "Total number of code issues detected",
            },
            []string{"severity", "type"},
        ),
        ProcessingTime: prometheus.NewHistogram(prometheus.HistogramOpts{
            Name: "code_analysis_processing_time_seconds",
            Help: "Time taken to complete code analysis",
        }),
        QualityScores: prometheus.NewHistogram(prometheus.HistogramOpts{
            Name: "code_quality_scores",
            Help: "Distribution of code quality scores",
        }),
        SecurityScores: prometheus.NewHistogram(prometheus.HistogramOpts{
            Name: "code_security_scores",
            Help: "Distribution of code security scores",
        }),
    }

    prometheus.MustRegister(
        metrics.AnalysesStarted,
        metrics.AnalysesCompleted,
        metrics.FilesProcessed,
        metrics.IssuesDetected,
        metrics.ProcessingTime,
        metrics.QualityScores,
        metrics.SecurityScores,
    )

    return metrics
}

func main() {
    config := &AnalysisConfig{
        DatabaseURL:        "postgres://user:pass@localhost/code_analysis_db?sslmode=disable",
        SupportedLanguages: []string{"go", "python", "javascript", "java"},
        MaxFileSize:        1024 * 1024, // 1MB
        SecurityRules: []SecurityRule{
            {
                ID:          "sql_injection",
                Name:        "SQL Injection",
                Description: "Potential SQL injection vulnerability",
                Pattern:     `"SELECT.*\+.*"`,
                Languages:   []string{"go", "python", "java"},
                Severity:    "critical",
                Category:    "injection",
            },
        },
        QualityRules: []QualityRule{
            {
                ID:          "function_complexity",
                Name:        "Function Complexity",
                Description: "Function complexity too high",
                Languages:   []string{"go", "python", "javascript", "java"},
                Threshold:   10.0,
                Metric:      "cyclomatic_complexity",
            },
        },
        ExcludePatterns: []string{"*.test.go", "vendor/*", "node_modules/*"},
        IncludeTests:    true,
        GenerateDocs:    true,
    }

    system, err := NewCodeAnalysisSystem(config)
    if err != nil {
        log.Fatal("Failed to create code analysis system:", err)
    }

    // Setup HTTP server
    r := gin.Default()

    // API routes
    api := r.Group("/api/v1")
    {
        api.POST("/analyze", system.StartAnalysisHandler)
        api.GET("/analyses/:id", system.GetAnalysisHandler)
        api.GET("/analyses/:id/files", system.GetFileAnalysesHandler)
        api.GET("/metrics", gin.WrapH(promhttp.Handler()))
    }

    // Health check
    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "status": "healthy",
            "supported_languages": config.SupportedLanguages,
}
}

    log.Println("Code Analysis System starting on :8080")
    log.Fatal(r.Run(":8080"))
}
```

## Usage Examples

### Start Code Analysis

```bash
curl -X POST http://localhost:8080/api/v1/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "repository_url": "https://github.com/example/repo.git",
    "commit_hash": "abc123",
    "branch": "main",
    "requested_by": "developer@company.com",
    "options": {
      "include_tests": true,
      "generate_docs": true,
      "languages": ["go", "python"],
      "max_files": 1000
    }
  }'
```

### Get Analysis Results

```bash
curl http://localhost:8080/api/v1/analyses/123
```

### Get File-Level Analysis

```bash
curl http://localhost:8080/api/v1/analyses/123/files
```

## Key Features Demonstrated

1. **Multi-Language Support** - Go, Python, JavaScript, Java analysis
2. **AST-Based Analysis** - Deep code structure understanding
3. **AI-Powered Reviews** - Intelligent quality and performance analysis
4. **Security Scanning** - Vulnerability detection and classification
5. **Documentation Generation** - Automatic code documentation
6. **Metrics and Reporting** - Comprehensive analysis reporting
7. **Git Integration** - Repository and commit-based analysis
8. **Performance Monitoring** - System performance metrics

## Integration Opportunities

- **CI/CD Pipelines** - Automated code review in build processes
- **IDE Extensions** - Real-time code analysis in development environments
- **Pull Request Automation** - Automatic PR review and feedback
- **Quality Gates** - Code quality thresholds in deployment pipelines
- **Team Dashboards** - Code quality metrics and trends

This code analysis system demonstrates how to build a comprehensive AI-powered code review platform using Go-LLMs, showcasing static analysis, AI-enhanced review capabilities, and production-ready architecture patterns.