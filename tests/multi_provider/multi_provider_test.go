package multi_provider

import (
	"testing"
)

// TestPrimaryStrategy tests just the primary provider strategy in isolation
func TestPrimaryStrategy(t *testing.T) {
	// The test should no longer be flaky after our sequential implementation fix
	// that makes primary provider behavior deterministic, but we've moved the tests
	// to the pkg/llm/provider directory for better organization
	t.Skip("This test has been replaced by deterministic tests in pkg/llm/provider/multi_primary_test.go and multi_deterministic_test.go")
}
