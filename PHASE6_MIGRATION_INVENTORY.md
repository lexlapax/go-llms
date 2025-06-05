# Phase 6 Migration Inventory

## Overview
This document inventories all deprecated code, old patterns, and migration tasks identified during Phase 6 analysis.

## 1. Code to Remove/Clean Up

### Comments to Clean Up
These comments reference removed functionality and can be deleted:

1. **pkg/llm/provider/multi.go** (line 866)
   - Comment about removed "legacy_selectConsensusTextResult function"

2. **pkg/structured/processor/processor.go** (line 98)
   - Comment about "Legacy extraction functions" being replaced

### Backward Compatibility Code to Review
These appear to be intentional for API stability but should be reviewed:

1. **pkg/llm/provider/gemini.go** (line 174)
   - Legacy compatibility for old message format

2. **pkg/llm/provider/openai.go** (lines 191, 204, 219, 244)
   - Multiple legacy format compatibility sections

3. **pkg/llm/provider/consensus.go** (line 518)
   - getCacheKey alias for backward compatibility

4. **pkg/schema/validation/validator.go** (lines 59-60)
   - Features disabled for backward compatibility

5. **cmd/config.go** (line 115)
   - Standard API key environment variables backward compatibility

## 2. Test Files to Migrate

### Files with `workflow_migration` Build Tag (10 files)
These files are currently excluded from builds and need migration:

**Benchmarks (3 files):**
- `benchmarks/agent_bench_test.go`
- `benchmarks/tools_bench_test.go` 
- `benchmarks/tools_builtin_bench_test.go`

**Integration Tests (6 files):**
- `tests/integration/agent_edge_cases_test.go`
- `tests/integration/agent_errors_test.go`
- `tests/integration/agent_test.go`
- `tests/integration/anthropic_e2e_test.go`
- `tests/integration/e2e_test.go`
- `tests/integration/gemini_agent_e2e_test.go`

**Stress Tests (1 file):**
- `tests/stress/agent_stress_test.go`

## 3. Examples to Update

### Workflow Examples (6 examples)
All workflow examples are functional but could be verified/updated:
- `cmd/examples/workflow-sequential/`
- `cmd/examples/workflow-parallel/`
- `cmd/examples/workflow-conditional/`
- `cmd/examples/workflow-loop/`
- `cmd/examples/workflow-hooks/`
- `cmd/examples/agent-workflow-as-tool/` (already updated but verify)

### Provider Examples to Verify
- `cmd/examples/multi/` - mentioned in TODO as needing update
- `cmd/examples/consensus/` - mentioned in TODO as needing update

## 4. Build Configuration

### .golangci.yml
Currently excludes `workflow_migration` tagged files from linting:
```yaml
build-tags:
  - "!workflow_migration"
```
This line should be removed after migration is complete.

## 5. No Action Required

### Working As Intended
1. **Deprecated field in builtins registry** - This is a feature, not deprecated code
2. **Debug build tags** in `pkg/internal/debug/` - Part of debug infrastructure
3. **SetDefaultAgent in workflow package** - This is the new architecture, not old code

### Already Migrated
1. No references to `workflow.Agent`, `DefaultAgent`, or `UnoptimizedDefaultAgent` types found
2. The convenience example has been properly updated
3. All agent-tool examples are working with new architecture

## 6. Priority Order for Migration

1. **High Priority:**
   - Migrate the 10 test files with `workflow_migration` tag
   - Move benchmarks directory to tests/benchmarks/
   - Remove .golangci.yml exclusion

2. **Medium Priority:**
   - Clean up legacy comments that reference removed code
   - Verify and update workflow examples
   - Update multi and consensus examples

3. **Low Priority:**
   - Review backward compatibility code for potential removal
   - Document any intentional backward compatibility decisions

## Next Steps
1. Start with migrating test files one by one
2. Update benchmarks location and content
3. Verify all examples work correctly
4. Clean up comments and build configuration