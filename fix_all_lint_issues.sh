#!/bin/bash

echo "Fixing all remaining lint issues..."

# Fix remaining errcheck issues
echo "Fixing errcheck issues..."

# agent-multi-coordination
sed -i 's/core\.Register(coordinator)/_ = core.Register(coordinator)/' cmd/examples/agent-multi-coordination/main.go
sed -i 's/core\.Register(analyst)/_ = core.Register(analyst)/' cmd/examples/agent-multi-coordination/main.go

# web tools - Body.Close issues
sed -i 's/defer resp\.Body\.Close()/defer func() { _ = resp.Body.Close() }()/' pkg/agent/builtins/tools/web/fetch.go
sed -i 's/defer resp\.Body\.Close()/defer func() { _ = resp.Body.Close() }()/' pkg/agent/builtins/tools/web/graphql.go
sed -i 's/defer resp\.Body\.Close()/defer func() { _ = resp.Body.Close() }()/' pkg/agent/builtins/tools/web/http_request.go

# search_extended_test.go
sed -i 's/os\.Unsetenv("BRAVE_API_KEY")/_ = os.Unsetenv("BRAVE_API_KEY")/' pkg/agent/builtins/tools/web/search_extended_test.go
sed -i 's/os\.Unsetenv("TAVILY_API_KEY")/_ = os.Unsetenv("TAVILY_API_KEY")/' pkg/agent/builtins/tools/web/search_extended_test.go
sed -i 's/os\.Unsetenv("SERPERDEV_API_KEY")/_ = os.Unsetenv("SERPERDEV_API_KEY")/' pkg/agent/builtins/tools/web/search_extended_test.go

# google_fetcher_test.go
sed -i 's/os\.Setenv("GEMINI_API_KEY",/_ = os.Setenv("GEMINI_API_KEY",/' pkg/util/llmutil/modelinfo/fetchers/google_fetcher_test.go
sed -i 's/defer os\.Setenv("GEMINI_API_KEY", origApiKey)/defer func() { _ = os.Setenv("GEMINI_API_KEY", origApiKey) }()/' pkg/util/llmutil/modelinfo/fetchers/google_fetcher_test.go

# api_client_bench_test.go
sed -i 's/json\.NewEncoder.*\.Encode/_ = &/' tests/benchmarks/api_client_bench_test.go
sed -i 's/json\.NewDecoder.*\.Decode/_ = &/' tests/benchmarks/api_client_bench_test.go

# tools_builtin_bench_test.go
sed -i 's/os\.Remove(csvFile)/_ = os.Remove(csvFile)/' tests/benchmarks/tools_builtin_bench_test.go

echo "Fixed errcheck issues"

# Fix ineffassign issues
echo "Fixing ineffassign issues..."

# provider-convenience
sed -i 's/err = loadNonExistentFile()/_ = loadNonExistentFile()/' cmd/examples/provider-convenience/main_test.go

# utils-profiling
sed -i 's/err = provider\.Complete(ctx, messages)/_ = provider.Complete(ctx, messages)/' cmd/examples/utils-profiling/main.go

echo "Fixed ineffassign issues"

# Fix staticcheck issues
echo "Fixing staticcheck issues..."

# Remove embedded field references (QF1008)
sed -i 's/a\.BaseAgentImpl\.Name()/a.Name()/' cmd/examples/agent-advanced-toolcontext/main.go
sed -i 's/c\.BaseAgentImpl\.Name()/c.Name()/' cmd/examples/agent-advanced-toolcontext/main.go
sed -i 's/r\.Registry\.List()/r.List()/' pkg/agent/builtins/agents/registry.go
sed -i 's/r\.Registry\.List()/r.List()/' pkg/agent/builtins/tools/registry.go
sed -i 's/t\.BaseAgentImpl\.Name()/t.Name()/' pkg/agent/tools/tool_agent.go
sed -i 's/t\.BaseAgentImpl\.Description()/t.Description()/' pkg/agent/tools/tool_agent.go

# Fix unnecessary fmt.Sprintf (S1039)
sed -i 's/supportType := fmt\.Sprintf("general")/supportType := "general"/' cmd/examples/agent-handoff/main.go

# Fix nil context (SA1012) 
sed -i 's/agent\.Run(nil, state)/agent.Run(context.TODO(), state)/' cmd/examples/simple/main_test.go

# Fix capitalized error strings (ST1005)
sed -i 's/"API key or token must be provided"/"api key or token must be provided"/' pkg/agent/builtins/tools/web/search.go
sed -i 's/"API response was empty or malformed"/"api response was empty or malformed"/' pkg/agent/builtins/tools/web/search.go  
sed -i 's/"API request failed"/"api request failed"/' pkg/agent/builtins/tools/web/search.go

echo "Fixed staticcheck issues"

# Fix unused functions - either remove underscore if they're used or keep underscore if truly unused
echo "Checking unused functions..."

# These functions are actually used, so remove underscore
sed -i 's/func _matchesResourceCriteria/func matchesResourceCriteria/' pkg/agent/builtins/tools/registry.go

echo "Fixed unused function issues"

echo "All lint issues fixed!"