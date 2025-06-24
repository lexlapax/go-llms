# Documentation-Only TODO

This file tracks ONLY documentation improvements. NO CODE CHANGES during this phase.

## Philosophy
We're in a documentation-only phase. The code works - now we need to document it properly. This includes:
1. In-line documentation (comments in code files)
2. Package documentation (doc.go files)  
3. Project documentation (markdown files)
4. Example documentation (example_test.go files)

## Priority 1: ABOUTME Comments (109 files)

Every Go file needs 2-line ABOUTME comments at the top. Currently missing in:

### Critical Provider Files
- [ ] pkg/llm/provider/anthropic.go
- [ ] pkg/llm/provider/openai.go
- [ ] pkg/llm/provider/gemini.go
- [ ] pkg/llm/provider/vertexai.go
- [ ] pkg/llm/provider/ollama.go (has one but needs second line)
- [ ] pkg/llm/provider/openrouter.go (has one but needs second line)
- [ ] pkg/llm/provider/mock.go
- [ ] pkg/llm/provider/consensus.go
- [ ] pkg/llm/provider/multi.go
- [ ] pkg/llm/provider/factories.go

### Core Domain Files
- [ ] pkg/llm/domain/interfaces.go
- [ ] pkg/llm/domain/types.go
- [ ] pkg/llm/domain/errors.go
- [ ] pkg/llm/domain/pool.go
- [ ] pkg/llm/domain/message_utils.go
- [ ] pkg/agent/domain/hooks.go
- [ ] pkg/agent/domain/interfaces.go
- [ ] pkg/agent/domain/types.go
- [ ] pkg/agent/domain/discovery.go
- [ ] pkg/agent/domain/events.go

### Schema & Validation Files
- [ ] pkg/schema/domain/types.go
- [ ] pkg/schema/validation/validator.go
- [ ] pkg/schema/validation/coercion.go
- [ ] pkg/schema/validation/cache.go
- [ ] pkg/schema/adapter/reflection/schema_generator.go
- [ ] pkg/schema/repository/file_repository.go

### Utility Files
- [ ] pkg/util/coalesce/coalesce.go
- [ ] pkg/util/metrics/metrics.go
- [ ] pkg/util/typeutil/type_utils.go
- [ ] pkg/util/llmutil/env.go
- [ ] pkg/util/llmutil/errors.go

(Plus 74 more files - see full list from audit)

## Priority 2: Package Documentation (doc.go files)

Critical packages needing doc.go files:

### High Priority Packages
- [ ] pkg/agent/tools - Explain tool system architecture, registration, discovery
- [ ] pkg/llm/domain - Explain core domain models, message types, responses
- [ ] pkg/llm/provider - Explain provider architecture, how to implement new providers
- [ ] pkg/schema/domain - Explain schema concepts, JSON schema integration
- [ ] pkg/structured/domain - Explain structured output concepts
- [ ] pkg/testutils/fixtures - Explain test fixture system

### Medium Priority Packages
- [ ] pkg/agent/scripting/domain - Explain scripting integration
- [ ] pkg/agent/scripting/registry - Explain script registry system
- [ ] pkg/util/auth - Explain authentication utilities
- [ ] pkg/util/coalesce - Explain coalesce pattern usage
- [ ] pkg/util/retry - Explain retry mechanisms

## Priority 3: Complex Function Documentation

Functions over 50 lines needing internal step-by-step comments:

### Critical Provider Methods
- [ ] GeminiProvider.StreamMessage (173 lines) - pkg/llm/provider/gemini.go
- [ ] MultiProvider.selectStructuredResult (138 lines) - pkg/llm/provider/multi.go
- [ ] GeminiProvider.GenerateMessage (119 lines) - pkg/llm/provider/gemini.go
- [ ] GeminiProvider.ConvertMessagesToGeminiFormat (104 lines) - pkg/llm/provider/gemini.go
- [ ] OpenAIFactory.GetTemplate (83 lines) - pkg/llm/provider/factories.go
- [ ] AnthropicProvider.ConvertMessagesToAnthropicFormat (82 lines) - pkg/llm/provider/anthropic.go
- [ ] VertexAIProvider.GenerateMessage (80 lines) - pkg/llm/provider/vertexai.go
- [ ] ConsensusProvider.StreamMessage (74 lines) - pkg/llm/provider/consensus.go
- [ ] AnthropicProvider.StreamMessage (65 lines) - pkg/llm/provider/anthropic.go
- [ ] mapGeminiErrorToStandard (60 lines) - pkg/llm/provider/gemini.go

### Complex Tool Functions
- [ ] validateJSONNodes (93 lines) - pkg/agent/builtins/tools/data/json_validator.go
- [ ] detectFileFormat (93 lines) - pkg/agent/builtins/tools/data/file_analyzer.go
- [ ] processJSONNodes (72 lines) - pkg/agent/builtins/tools/data/json_processor.go
- [ ] transformData (60 lines) - pkg/agent/builtins/tools/data/data_transform.go

### Complex Validation Logic
- [ ] coerceToType (152 lines) - pkg/schema/validation/coercion.go
- [ ] getFieldByNameOrTag (109 lines) - pkg/schema/validation/type_mapping.go
- [ ] stringToType (56 lines) - pkg/schema/validation/coercion.go

## Priority 4: Example Tests

Major packages needing example_test.go files:

### Provider Examples
- [ ] pkg/llm/provider/example_test.go - Show how to use each provider
- [ ] pkg/agent/core/example_test.go - Show agent creation and usage
- [ ] pkg/agent/tools/example_test.go - Show tool creation and registration
- [ ] pkg/schema/validation/example_test.go - Show schema validation
- [ ] pkg/structured/processor/example_test.go - Show structured output processing
- [ ] pkg/agent/workflows/example_test.go - Show workflow creation

## Priority 5: Project Documentation

### Architecture & Design Docs
- [ ] docs/architecture.md - System design, component relationships
- [ ] docs/provider-comparison.md - Provider capabilities, limitations, pricing
- [ ] docs/tool-development.md - How to create custom tools
- [ ] docs/schema-guide.md - JSON schema usage and validation
- [ ] docs/workflow-guide.md - Creating and running workflows

### API Reference
- [ ] docs/api/providers.md - Provider API reference
- [ ] docs/api/agents.md - Agent API reference
- [ ] docs/api/tools.md - Tool API reference
- [ ] docs/api/schemas.md - Schema API reference

### Guides
- [ ] docs/guides/getting-started.md - Beyond basic README
- [ ] docs/guides/migration.md - Upgrading between versions
- [ ] docs/guides/troubleshooting.md - Common issues and solutions
- [ ] docs/guides/performance.md - Performance tuning and optimization

## Priority 6: Comment Quality Improvements

### Files with Poor Comment-to-Code Ratio
Large files (100+ lines) with <5 comments needing improvement:
- [ ] pkg/agent/builtins/tools/data/data_transform_test.go (806 lines, 3 comments)
- [ ] pkg/agent/builtins/tools/data/csv_process_test.go (506 lines, 2 comments)
- [ ] pkg/schema/validation/coercion_advanced_test.go (475 lines, 2 comments)
- [ ] pkg/agent/builtins/tools/data/xml_process_test.go (446 lines, 2 comments)
- [ ] pkg/testutils/helpers/state_test.go (424 lines, 4 comments)

## Measurement Criteria

For each file/package, documentation is complete when:
1. **ABOUTME**: Has 2 descriptive lines at file top
2. **Godoc**: Every exported item has proper godoc comment
3. **Complex functions**: Logic is explained with inline comments
4. **Package doc**: doc.go explains package purpose and usage
5. **Examples**: At least one example_test.go demonstrating usage
6. **No code changes**: Only comments and documentation files added

## Progress Tracking

- Total ABOUTME needed: 109
- Total doc.go needed: 11+ 
- Total complex functions: 17+
- Total example tests needed: 20+
- Total project docs needed: 12+

Start with Priority 1 and work systematically through each priority level.