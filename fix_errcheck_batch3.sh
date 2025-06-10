#!/bin/bash

# Fix agent-multi-coordination example
sed -i 's/core.Register(webResearcher)/_ = core.Register(webResearcher)/' cmd/examples/agent-multi-coordination/main.go
sed -i 's/core.Register(academicResearcher)/_ = core.Register(academicResearcher)/' cmd/examples/agent-multi-coordination/main.go
sed -i 's/core.Register(dataAnalyst)/_ = core.Register(dataAnalyst)/' cmd/examples/agent-multi-coordination/main.go

# Fix provider-options example
sed -i 's/os.Setenv("OPENAI_USE_CASE",/_ = os.Setenv("OPENAI_USE_CASE",/' cmd/examples/provider-options/main.go
sed -i 's/defer os.Setenv("OPENAI_USE_CASE", origUseCase)/defer func() { _ = os.Setenv("OPENAI_USE_CASE", origUseCase) }()/' cmd/examples/provider-options/main.go

# Fix provider-options test
sed -i 's/os.Setenv("LLM_HTTP_TIMEOUT", origTimeout)/_ = os.Setenv("LLM_HTTP_TIMEOUT", origTimeout)/' cmd/examples/provider-options/main_test.go

# Fix any remaining fmt.Fprintln issues
sed -i 's/fmt.Fprintln(w,/_, _ = fmt.Fprintln(w,/' pkg/llm/provider/openai_test.go

echo "Fixed batch 3 of errcheck issues"