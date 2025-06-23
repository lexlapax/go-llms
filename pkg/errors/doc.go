// ABOUTME: Enhanced error handling with serialization and recovery strategies.
// ABOUTME: Context-aware errors with aggregation and structured error types.
// Package errors provides enhanced error handling capabilities beyond
// standard Go errors. It includes serializable errors for API responses,
// error aggregation for multiple failures, context tracking for debugging,
// and recovery strategies for error handling patterns.
//
// Features:
//   - Serializable errors with JSON support
//   - Error aggregation for batch operations
//   - Context propagation for debugging
//   - Recovery strategies and patterns
//   - Structured error types with metadata
package errors
