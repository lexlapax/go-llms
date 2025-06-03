// ABOUTME: Common test helper functions for domain package tests
// ABOUTME: Provides shared utilities to avoid duplication across test files

package domain_test

import "fmt"

// sprintf is a helper function for string formatting in tests
func sprintf(format string, a ...interface{}) string {
	return fmt.Sprintf(format, a...)
}
