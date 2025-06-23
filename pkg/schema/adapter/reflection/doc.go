// ABOUTME: Reflection-based schema generation from Go types and structs.
// ABOUTME: Automatic JSON schema creation using struct tags and type analysis.
// Package reflection provides adapters for generating JSON schemas from
// Go types using reflection. It analyzes struct tags, field types, and
// validation constraints to automatically create accurate schemas.
//
// The reflection adapter supports:
//   - Struct field analysis with json tags
//   - Nested type resolution
//   - Validation constraint detection
//   - Custom type handlers
//   - Recursive schema generation
package reflection
