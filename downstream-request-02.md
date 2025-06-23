# go-llms Upstream Feature Requests

This document tracks features that are missing from go-llms but needed for go-llmspell's bridge architecture. These should be implemented in go-llms first, then bridged in go-llmspell.

## Current Date: 2025-06-16

## Bridge-First Principle Compliance

**Rule**: If it's not in go-llms, we don't implement it in go-llmspell.

When we identify missing functionality during bridge implementation, we document it here for upstream contribution to go-llms.

---

## 1. Model Performance Analytics System

**Context**: Phase 1.4.6.1 - Add Model Performance Analytics  
**Status**: Missing from go-llms  
**Priority**: High  

### 1.1 Model Performance Tracking
- [ ] Implement model-specific performance metrics collection
- [ ] Add latency tracking per model/provider
- [ ] Implement token usage analytics per model
- [ ] Add cost tracking and calculation per model
- [ ] Create performance trend analysis over time
- [ ] Add anomaly detection for model performance

### 1.2 Suggested Implementation Location
- **Package**: `pkg/util/llmutil/analytics/`
- **Files**: 
  - `performance_tracker.go` - Core performance tracking
  - `model_metrics.go` - Model-specific metrics
  - `cost_calculator.go` - Token cost calculations
  - `trend_analyzer.go` - Performance trend analysis
  - `anomaly_detector.go` - Performance anomaly detection

### 1.3 Interface Design
```go
type ModelPerformanceTracker interface {
    TrackRequest(modelID string, latency time.Duration, inputTokens, outputTokens int, cost float64)
    GetModelPerformance(modelID string) (*ModelPerformanceReport, error)
    GetPerformanceTrends(modelID string, timeRange time.Duration) (*PerformanceTrends, error)
    DetectAnomalies(modelID string) ([]PerformanceAnomaly, error)
}
```

---

## 2. Model Recommendation Engine

**Context**: Phase 1.4.6.2 - Add Model Recommendation Engine  
**Status**: Missing from go-llms  
**Priority**: High  

### 2.1 Model Selection Algorithms
- [ ] Implement capability-based model matching
- [ ] Add task-specific model recommendations
- [ ] Create cost/performance optimization algorithms
- [ ] Implement multi-criteria decision making (MCDM)
- [ ] Add recommendation explanations and reasoning
- [ ] Support A/B testing for model selection

### 2.2 Suggested Implementation Location
- **Package**: `pkg/util/llmutil/recommendation/`
- **Files**:
  - `recommender.go` - Core recommendation engine
  - `capability_matcher.go` - Capability-based matching
  - `cost_optimizer.go` - Cost/performance optimization
  - `task_classifier.go` - Task-specific recommendations
  - `explanation_generator.go` - Recommendation explanations

### 2.3 Interface Design
```go
type ModelRecommender interface {
    FindModelsWithCapabilities(capabilities []string) ([]ModelRecommendation, error)
    RecommendForTask(taskType string, constraints ModelConstraints) ([]ModelRecommendation, error)
    OptimizeForCostPerformance(requirements PerformanceRequirements) ([]ModelRecommendation, error)
    ExplainRecommendation(recommendation ModelRecommendation) (string, error)
}
```

---

## 3. Model Catalog Export System

**Context**: Phase 1.4.6.3 - Add Model Catalog Export  
**Status**: Missing from go-llms  
**Priority**: Medium  

### 3.1 Documentation Generation
- [ ] Implement OpenAPI specification export for model catalog
- [ ] Add interactive documentation generation
- [ ] Include pricing information in exports
- [ ] Generate capability matrices and comparison charts
- [ ] Support multiple export formats (JSON, YAML, Markdown)
- [ ] Add custom export format support

### 3.2 Suggested Implementation Location
- **Package**: `pkg/util/llmutil/catalog/`
- **Files**:
  - `exporter.go` - Core catalog export functionality
  - `openapi_generator.go` - OpenAPI specification generation
  - `docs_generator.go` - Interactive documentation
  - `comparison_generator.go` - Capability comparison charts
  - `format_converter.go` - Multi-format export support

### 3.3 Interface Design
```go
type CatalogExporter interface {
    ExportToOpenAPI(models []Model, version string) (*OpenAPISpec, error)
    GenerateInteractiveDocs(models []Model) (*InteractiveDocs, error)
    ExportComparisonMatrix(models []Model) (*ComparisonMatrix, error)
    ExportToFormat(models []Model, format ExportFormat) ([]byte, error)
}
```

---

## 4. Enhanced Metrics Integration

**Context**: Extend existing metrics system for model-specific tracking  
**Status**: Partially exists in go-llms  
**Priority**: Medium  

### 4.1 Model-Specific Metrics
- [ ] Extend existing `pkg/util/metrics/` with model-specific counters
- [ ] Add model performance histograms
- [ ] Implement provider-specific metric aggregation
- [ ] Add real-time metric streaming capabilities

### 4.2 Suggested Enhancement Location
- **Package**: `pkg/util/metrics/` (extend existing)
- **Files**:
  - `model_metrics.go` - Model-specific metric types
  - `provider_aggregator.go` - Provider-level aggregation
  - `streaming_metrics.go` - Real-time metric streaming

---

## 5. Integration with Existing Systems

**Context**: Ensure new features integrate with existing go-llms architecture  
**Status**: Design consideration  
**Priority**: High  

### 5.1 Integration Points
- [ ] Integrate with existing `ModelInfoService`
- [ ] Leverage existing metrics registry
- [ ] Extend model inventory with performance data
- [ ] Integrate with provider metadata system
- [ ] Use existing event system for metric collection

### 5.2 Backward Compatibility
- [ ] Ensure all new features are optional and don't break existing APIs
- [ ] Provide configuration options to enable/disable analytics
- [ ] Maintain existing model info interfaces

---

## 6. Testing Requirements

**Context**: Comprehensive testing for new features  
**Status**: Required for all new features  
**Priority**: High  

### 6.1 Test Coverage Requirements
- [ ] Unit tests for all new interfaces and implementations
- [ ] Integration tests with existing model info system
- [ ] Performance benchmarks for analytics overhead
- [ ] Mock implementations for testing
- [ ] Example usage in `cmd/examples/`

### 6.2 Test Location
- **Package**: Follow existing go-llms testing patterns
- **Files**: `*_test.go` files alongside implementations
- **Benchmarks**: `tests/benchmarks/` directory
- **Integration**: `tests/integration/` directory

---

## 7. Implementation Priority

**Recommended Order**:

1. **Model Performance Analytics** (Phase 1.4.6.1)
   - Foundation for other features
   - Extends existing metrics system
   - High value for users

2. **Model Recommendation Engine** (Phase 1.4.6.2)
   - Builds on performance analytics
   - Complex algorithms requiring careful design
   - High impact on user experience

3. **Model Catalog Export** (Phase 1.4.6.3)
   - Documentation and tooling feature
   - Lower complexity
   - Can be implemented independently

4. **Enhanced Metrics Integration** (Phase 4)
   - Supports all other features
   - Extends existing system
   - Continuous improvement

---

## 8. Notes for go-llms Contributors

### 8.1 Design Principles
- Follow existing go-llms patterns and conventions
- Maintain backward compatibility
- Use dependency injection where appropriate
- Leverage existing interfaces and abstractions

### 8.2 Dependencies
- Build on existing `pkg/util/metrics/` system
- Integrate with `pkg/util/llmutil/modelinfo/` 
- Use existing provider metadata where available
- Follow existing error handling patterns

### 8.3 Documentation
- Add comprehensive godoc comments
- Include usage examples
- Update main README.md with new features
- Add to `docs/` if complex features require detailed documentation

---

## 9. go-llmspell Bridge Implementation Plan

**After go-llms Implementation**:

Once these features are available in go-llms, go-llmspell will implement corresponding bridges:

- `ModelPerformanceBridge` - Bridge performance analytics
- `ModelRecommendationBridge` - Bridge recommendation engine  
- `ModelCatalogBridge` - Bridge catalog export functionality

**Bridge Location**: `pkg/bridge/modelinfo/` (extend existing)

---

## 10. Review and Approval Process

### 10.1 go-llms Review
- [ ] Feature design review with go-llms maintainers
- [ ] API design approval
- [ ] Implementation review
- [ ] Testing and documentation review

### 10.2 go-llmspell Integration
- [ ] Bridge implementation after go-llms release
- [ ] Integration testing
- [ ] Documentation updates
- [ ] Phase 1.4.6 completion

---

## 11. Script Documentation Generation Extensions

**Context**: Task 2.4.3.3 - Documentation Generator capabilities  
**Status**: Partially exists in go-llms (tools only)  
**Priority**: High  
**Date Added**: 2025-06-21

### 11.1 Script-Aware Documentation Support

go-llms already has excellent documentation generation for tools, but needs extensions for script-based systems:

- [ ] Extend `Documentable` interface to support script metadata
- [ ] Add language-specific documentation extraction
- [ ] Support for script-specific schemas (input/output parameters)
- [ ] Script example extraction and validation
- [ ] Multi-language script support (Lua, JavaScript, Tengo)

### 11.2 Suggested Implementation Location
- **Package**: `pkg/docs/` (extend existing)
- **Files**:
  - `script_documentable.go` - Script-aware Documentable implementation
  - `script_extractor.go` - Extract documentation from script files
  - `language_analyzer.go` - Language-specific analysis
  - `example_validator.go` - Validate script examples

### 11.3 Interface Design
```go
// ScriptDocumentable extends Documentable for script systems
type ScriptDocumentable interface {
    Documentable
    GetScriptLanguage() string
    GetScriptSource() string
    GetScriptParameters() []Parameter
    GetScriptExamples() []ScriptExample
}

// ScriptDocumentationExtractor extracts docs from scripts
type ScriptDocumentationExtractor interface {
    ExtractFromScript(path string, language string) (Documentation, error)
    ExtractFromSource(source string, language string) (Documentation, error)
    ValidateExamples(doc Documentation) ([]ValidationResult, error)
}
```

### 11.4 go-llmspell Usage

Once implemented in go-llms, go-llmspell would:

1. **Bridge the ScriptDocumentable interface**
   ```go
   // In pkg/bridge/docs/
   type ScriptDocumentableBridge struct {
       script *runner.Script
       engine engine.Engine
   }
   
   func (s *ScriptDocumentableBridge) GetDocumentation() docs.Documentation {
       // Convert script to go-llms Documentation format
   }
   ```

2. **Use go-llms generators directly**
   ```go
   // Generate documentation using go-llms infrastructure
   generator := docs.NewMarkdownGenerator(config)
   markdown, _ := generator.GenerateMarkdown(ctx, documentables)
   ```

3. **Language-specific extractors remain in go-llmspell**
   - Keep `gendocs_lua.go`, `gendocs_javascript.go`, etc.
   - These implement the extraction logic for each language
   - Feed extracted data into go-llms documentation system

---

## 12. Man Page Generation System

**Context**: Task 3.3 - Shell completion and man page generation  
**Status**: Not in go-llms  
**Priority**: Medium  
**Date Added**: 2025-06-21

### 12.1 Unix Man Page Generation

The man page generation system in go-llmspell's `pkg/docs/manpage.go` is generic and could benefit other go-llms tools:

- [ ] Implement structured man page data model
- [ ] Add troff format generation
- [ ] Support man page sections (1-8)
- [ ] Generate man pages from command metadata
- [ ] Support sub-command documentation
- [ ] Include examples and cross-references

### 12.2 Suggested Implementation Location
- **Package**: `pkg/util/docs/manpage/`
- **Files**:
  - `manpage.go` - Core man page types and generator
  - `troff.go` - Troff format generation
  - `command_extractor.go` - Extract from CLI commands
  - `formatter.go` - Format conversions (HTML, text)

### 12.3 Interface Design
```go
// ManPage represents a complete man page
type ManPage struct {
    Name        string
    Section     int
    Version     string
    Date        string
    Description string
    Synopsis    string
    Options     []Option
    Commands    []Command
    Examples    []Example
    Files       []string
    SeeAlso     []string
    Authors     []string
    Bugs        string
}

// ManPageGenerator generates man pages
type ManPageGenerator interface {
    GenerateManPage(cmd Command) (*ManPage, error)
    GenerateTroff(man *ManPage) string
    GenerateHTML(man *ManPage) string
    GenerateText(man *ManPage) string
}
```

### 12.4 Integration with CLI Tools

This would integrate with Kong-based CLIs (like go-llms tools):

```go
// Extract from Kong CLI structure
type KongManPageExtractor interface {
    ExtractFromKongApp(app *kong.Application) ([]*ManPage, error)
    ExtractFromKongCommand(cmd *kong.Command) (*ManPage, error)
}
```

### 12.5 go-llmspell Usage

Once upstreamed, go-llmspell would:

1. **Remove local manpage.go implementation**
2. **Bridge go-llms man page generator**
3. **Use for all man page generation needs**

---

## 13. Enhanced Documentation Integration

**Context**: Unified documentation system for go-llms ecosystem  
**Status**: Design consideration  
**Priority**: High  
**Date Added**: 2025-06-21

### 13.1 Unified Documentation Pipeline

Create a comprehensive documentation pipeline that supports:

- [ ] Tools (existing)
- [ ] Scripts (new - Section 11)
- [ ] CLI commands (via man pages - Section 12)
- [ ] APIs and bridges
- [ ] Examples and tutorials

### 13.2 Benefits of Upstream Implementation

1. **Consistency**: All go-llms-based tools use same documentation format
2. **Reusability**: Man page generation benefits all CLI tools
3. **Maintenance**: Single implementation to maintain
4. **Integration**: Documentation can cross-reference between tools/scripts
5. **Export**: Unified export to various formats

### 13.3 Migration Path for go-llmspell

1. **Phase 1**: Upstream man page generation
2. **Phase 2**: Extend Documentable for scripts
3. **Phase 3**: Migrate go-llmspell to use go-llms docs
4. **Phase 4**: Remove duplicate implementation

---

## 14. Implementation Priority for Documentation Features

**Recommended Order**:

1. **Man Page Generation** (Section 12)
   - Self-contained feature
   - Immediately useful for go-llms CLI tools
   - No breaking changes

2. **Script Documentation Extensions** (Section 11)
   - Builds on existing documentation system
   - Extends interfaces without breaking changes
   - Enables go-llmspell migration

3. **Enhanced Integration** (Section 13)
   - Long-term vision
   - Requires both previous features
   - Provides unified documentation experience