// Package tools provides implementations of agent tools.
package tools

// ABOUTME: Provides base implementation for agent tools with reflection-based parameter handling
// ABOUTME: Handles automatic parameter conversion and validation for tool functions

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"sync"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// Tool provides an optimized implementation of tools with reduced allocations
type Tool struct {
	name         string
	description  string
	fn           interface{}
	paramSchema  *sdomain.Schema
	outputSchema *sdomain.Schema

	// LLM guidance fields
	usageInstructions string
	examples          []domain.ToolExample
	constraints       []string
	errorGuidance     map[string]string

	// Metadata fields
	category string
	tags     []string
	version  string

	// Behavioral fields
	isDeterministic      bool
	isDestructive        bool
	requiresConfirmation bool
	estimatedLatency     string

	// Pre-computed type information
	fnType         reflect.Type
	fnValue        reflect.Value
	numArgs        int
	hasContext     bool
	hasToolContext bool
	nonContextArg  int

	// Cache for commonly used values to reduce allocations
	argsPool sync.Pool
}

// NewTool creates a new tool from a function
// This implementation includes optimizations for better performance:
// - Pre-computes type information at creation time
// - Uses object pooling to reduce GC pressure
// - Implements fast paths for common type conversions
// - Caches struct field information for improved parameter mapping
func NewTool(name, description string, fn interface{}, paramSchema *sdomain.Schema) domain.Tool {
	fnValue := reflect.ValueOf(fn)
	if fnValue.Kind() != reflect.Func {
		panic("tool function must be a function")
	}

	fnType := fnValue.Type()
	numArgs := fnType.NumIn()

	// Determine if the function accepts context as first argument
	// Check for both context.Context and *domain.ToolContext
	hasContext := false
	hasToolContext := false
	if numArgs > 0 {
		firstArgType := fnType.In(0)
		hasContext = firstArgType.Implements(reflect.TypeOf((*context.Context)(nil)).Elem())
		hasToolContext = firstArgType == reflect.TypeOf((*domain.ToolContext)(nil))
	}

	// Calculate index of the first non-context argument
	nonContextArg := 0
	if hasContext || hasToolContext {
		nonContextArg = 1
	}

	tool := &Tool{
		name:           name,
		description:    description,
		fn:             fn,
		paramSchema:    paramSchema,
		fnType:         fnType,
		fnValue:        fnValue,
		numArgs:        numArgs,
		hasContext:     hasContext,
		hasToolContext: hasToolContext,
		nonContextArg:  nonContextArg,
		// Set default values for new fields
		version:              "1.0.0",
		isDeterministic:      true,
		isDestructive:        false,
		requiresConfirmation: false,
		estimatedLatency:     "medium",
	}

	// Initialize argument pool with pointer to slice for efficient pooling
	tool.argsPool = sync.Pool{
		New: func() interface{} {
			slice := make([]reflect.Value, numArgs)
			return &slice // Return pointer to avoid allocations on pool.Put
		},
	}

	return tool
}

// ToolBuilder provides a fluent interface for building tools with comprehensive metadata
type ToolBuilder struct {
	tool *Tool
}

// NewToolBuilder creates a new tool builder
func NewToolBuilder(name, description string) *ToolBuilder {
	return &ToolBuilder{
		tool: &Tool{
			name:                 name,
			description:          description,
			version:              "1.0.0",
			isDeterministic:      true,
			isDestructive:        false,
			requiresConfirmation: false,
			estimatedLatency:     "medium",
		},
	}
}

// WithFunction sets the tool function
func (b *ToolBuilder) WithFunction(fn interface{}) *ToolBuilder {
	fnValue := reflect.ValueOf(fn)
	if fnValue.Kind() != reflect.Func {
		panic("tool function must be a function")
	}

	b.tool.fn = fn
	b.tool.fnValue = fnValue
	b.tool.fnType = fnValue.Type()
	b.tool.numArgs = b.tool.fnType.NumIn()

	// Determine if the function accepts context as first argument
	if b.tool.numArgs > 0 {
		firstArgType := b.tool.fnType.In(0)
		b.tool.hasContext = firstArgType.Implements(reflect.TypeOf((*context.Context)(nil)).Elem())
		b.tool.hasToolContext = firstArgType == reflect.TypeOf((*domain.ToolContext)(nil))
	}

	// Calculate index of the first non-context argument
	b.tool.nonContextArg = 0
	if b.tool.hasContext || b.tool.hasToolContext {
		b.tool.nonContextArg = 1
	}

	return b
}

// WithParameterSchema sets the parameter schema
func (b *ToolBuilder) WithParameterSchema(schema *sdomain.Schema) *ToolBuilder {
	b.tool.paramSchema = schema
	return b
}

// WithOutputSchema sets the output schema
func (b *ToolBuilder) WithOutputSchema(schema *sdomain.Schema) *ToolBuilder {
	b.tool.outputSchema = schema
	return b
}

// WithUsageInstructions sets the usage instructions
func (b *ToolBuilder) WithUsageInstructions(instructions string) *ToolBuilder {
	b.tool.usageInstructions = instructions
	return b
}

// WithExamples sets the examples
func (b *ToolBuilder) WithExamples(examples []domain.ToolExample) *ToolBuilder {
	b.tool.examples = examples
	return b
}

// WithConstraints sets the constraints
func (b *ToolBuilder) WithConstraints(constraints []string) *ToolBuilder {
	b.tool.constraints = constraints
	return b
}

// WithErrorGuidance sets the error guidance
func (b *ToolBuilder) WithErrorGuidance(guidance map[string]string) *ToolBuilder {
	b.tool.errorGuidance = guidance
	return b
}

// WithCategory sets the category
func (b *ToolBuilder) WithCategory(category string) *ToolBuilder {
	b.tool.category = category
	return b
}

// WithTags sets the tags
func (b *ToolBuilder) WithTags(tags []string) *ToolBuilder {
	b.tool.tags = tags
	return b
}

// WithVersion sets the version
func (b *ToolBuilder) WithVersion(version string) *ToolBuilder {
	b.tool.version = version
	return b
}

// WithBehavior sets the behavioral metadata
func (b *ToolBuilder) WithBehavior(deterministic, destructive, requiresConfirmation bool, latency string) *ToolBuilder {
	b.tool.isDeterministic = deterministic
	b.tool.isDestructive = destructive
	b.tool.requiresConfirmation = requiresConfirmation
	b.tool.estimatedLatency = latency
	return b
}

// Build creates the final tool
func (b *ToolBuilder) Build() domain.Tool {
	if b.tool.fn == nil {
		panic("tool function is required")
	}

	// Initialize argument pool
	b.tool.argsPool = sync.Pool{
		New: func() interface{} {
			slice := make([]reflect.Value, b.tool.numArgs)
			return &slice
		},
	}

	return b.tool
}

// Name returns the tool's name
func (t *Tool) Name() string {
	return t.name
}

// Description provides information about the tool
func (t *Tool) Description() string {
	return t.description
}

// ParameterSchema returns the schema for the tool parameters
func (t *Tool) ParameterSchema() *sdomain.Schema {
	return t.paramSchema
}

// Execute runs the tool with parameters
func (t *Tool) Execute(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
	// Get an arguments slice from the pool
	argsPtr := t.argsPool.Get().(*[]reflect.Value)
	args := *argsPtr
	defer func() {
		// Clear the slice before returning it to the pool
		for i := range args {
			args[i] = reflect.Value{}
		}
		t.argsPool.Put(argsPtr)
	}()

	// If function expects a context, set it as the first argument
	if t.hasToolContext {
		args[0] = reflect.ValueOf(ctx)
	} else if t.hasContext {
		args[0] = reflect.ValueOf(ctx.Context)
	}

	// Check if we need parameters
	if params == nil {
		// If the function takes no arguments (besides potentially context), call it directly
		if t.nonContextArg >= t.numArgs {
			return t.callFunction(args)
		}
		return nil, fmt.Errorf("function requires parameters but none provided")
	}

	// Handle parameter preparation with optimized path
	err := t.prepareArguments(ctx.Context, params, args)
	if err != nil {
		return nil, fmt.Errorf("error preparing arguments: %w", err)
	}

	// Call the function
	return t.callFunction(args)
}

// prepareArguments converts the params to the appropriate argument types for the function
// nolint:gocyclo // This function handles many parameter conversion cases
func (t *Tool) prepareArguments(ctx context.Context, params interface{}, args []reflect.Value) error {
	// If no more arguments needed besides context, we're done
	if t.nonContextArg >= t.numArgs {
		return nil
	}

	// Handle the params based on what was provided
	paramsValue := reflect.ValueOf(params)

	// Handle slice parameters specially for functions taking multiple arguments
	if paramsValue.Kind() == reflect.Slice && t.numArgs-t.nonContextArg == paramsValue.Len() {
		// Directly assign each slice element to each function argument
		for i := 0; i < paramsValue.Len(); i++ {
			argIndex := t.nonContextArg + i
			argValue := paramsValue.Index(i)

			// Try to convert if needed
			if argValue.Type().AssignableTo(t.fnType.In(argIndex)) {
				args[argIndex] = argValue
			} else if convertedValue, ok := optimizedConvertValue(argValue, t.fnType.In(argIndex)); ok {
				args[argIndex] = convertedValue
			} else {
				return fmt.Errorf("unable to convert slice parameter at index %d to function argument type", i)
			}
		}
		return nil
	}

	// If params is a map and function expects a struct, try to map fields
	if paramsValue.Kind() == reflect.Map && t.fnType.In(t.nonContextArg).Kind() == reflect.Struct {
		targetType := t.fnType.In(t.nonContextArg)
		structVal := reflect.New(targetType).Elem()

		// Use the cached field info to map values
		fields := globalParamCache.getStructFields(targetType)

		for _, field := range fields {
			if !field.isExported {
				continue // Skip unexported fields
			}

			// Try to find the field in the map by both the struct field name and JSON name
			var mapFieldValue reflect.Value

			jsonKeyValue := reflect.ValueOf(field.jsonName)
			mapFieldValue = paramsValue.MapIndex(jsonKeyValue)

			if !mapFieldValue.IsValid() {
				nameKeyValue := reflect.ValueOf(field.name)
				mapFieldValue = paramsValue.MapIndex(nameKeyValue)
			}

			if !mapFieldValue.IsValid() {
				continue // Skip fields not found in the map
			}

			// Get the field value and ensure we can set it
			fieldValue := structVal.Field(field.index)
			if !fieldValue.CanSet() {
				continue
			}

			// Try to convert and set the value (using optimized conversion)
			convertedValue, ok := optimizedConvertValue(mapFieldValue, field.fieldType)
			if ok {
				fieldValue.Set(convertedValue)
			}
		}

		args[t.nonContextArg] = structVal
		return nil
	}

	// If params can be directly assigned to the function's argument type
	if t.nonContextArg < t.numArgs && paramsValue.Type().AssignableTo(t.fnType.In(t.nonContextArg)) {
		args[t.nonContextArg] = paramsValue
		return nil
	}

	// Try to convert the value
	if t.nonContextArg < t.numArgs {
		if convertedValue, ok := optimizedConvertValue(paramsValue, t.fnType.In(t.nonContextArg)); ok {
			args[t.nonContextArg] = convertedValue
			return nil
		}
	}

	return fmt.Errorf("unable to convert parameters to function argument types")
}

// optimizedConvertValue attempts to convert a value to the target type
// This version is optimized to reduce allocations
// nolint:gocyclo // This function handles many type conversion cases
func optimizedConvertValue(value reflect.Value, targetType reflect.Type) (reflect.Value, bool) {
	// Special handling for interface{} type
	if value.Type().Kind() == reflect.Interface && !value.IsNil() {
		// Extract the actual value from the interface
		return optimizedConvertValue(value.Elem(), targetType)
	}

	// Fast path: if directly assignable, return as is
	if value.Type().AssignableTo(targetType) {
		return value, true
	}

	// Check if conversion is possible (using cache)
	// Skip this check for complex types as the cache may be incomplete
	canConvert := globalParamCache.canConvert(value.Type(), targetType)
	_ = canConvert // We use this later for cache updates

	// Note: We continue with conversion attempts regardless of canConvert result,
	// as the conversion logic below might handle cases not in the cache

	// Handle basic type conversions with optimized paths
	switch targetType.Kind() {
	case reflect.String:
		// Fast path for string conversion
		switch value.Kind() {
		case reflect.String:
			return value, true
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return reflect.ValueOf(strconv.FormatInt(value.Int(), 10)), true
		case reflect.Float32, reflect.Float64:
			return reflect.ValueOf(strconv.FormatFloat(value.Float(), 'f', -1, 64)), true
		case reflect.Bool:
			return reflect.ValueOf(strconv.FormatBool(value.Bool())), true
		default:
			return reflect.ValueOf(fmt.Sprintf("%v", value.Interface())), true
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Fast path for int conversion
		switch value.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return reflect.ValueOf(value.Int()).Convert(targetType), true
		case reflect.Float32, reflect.Float64:
			return reflect.ValueOf(int64(value.Float())).Convert(targetType), true
		case reflect.String:
			if i, err := strconv.ParseInt(value.String(), 10, 64); err == nil {
				return reflect.ValueOf(i).Convert(targetType), true
			}
		case reflect.Bool:
			if value.Bool() {
				return reflect.ValueOf(int64(1)).Convert(targetType), true
			}
			return reflect.ValueOf(int64(0)).Convert(targetType), true
		}

	case reflect.Float32, reflect.Float64:
		// Fast path for float conversion
		switch value.Kind() {
		case reflect.Float32, reflect.Float64:
			return reflect.ValueOf(value.Float()).Convert(targetType), true
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return reflect.ValueOf(float64(value.Int())).Convert(targetType), true
		case reflect.String:
			if f, err := strconv.ParseFloat(value.String(), 64); err == nil {
				return reflect.ValueOf(f).Convert(targetType), true
			}
		case reflect.Bool:
			if value.Bool() {
				return reflect.ValueOf(float64(1.0)).Convert(targetType), true
			}
			return reflect.ValueOf(float64(0.0)).Convert(targetType), true
		}

	case reflect.Bool:
		// Fast path for bool conversion
		switch value.Kind() {
		case reflect.Bool:
			return value, true
		case reflect.String:
			if b, err := strconv.ParseBool(value.String()); err == nil {
				return reflect.ValueOf(b), true
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return reflect.ValueOf(value.Int() != 0), true
		case reflect.Float32, reflect.Float64:
			return reflect.ValueOf(value.Float() != 0), true
		}

	// Handle slice conversions
	case reflect.Slice, reflect.Array:
		if value.Kind() == reflect.Slice || value.Kind() == reflect.Array {
			elemType := targetType.Elem()
			length := value.Len()
			result := reflect.MakeSlice(targetType, length, length)

			for i := 0; i < length; i++ {
				elemValue := value.Index(i)
				convertedElem, ok := optimizedConvertValue(elemValue, elemType)
				if !ok {
					return reflect.Value{}, false
				}
				result.Index(i).Set(convertedElem)
			}
			return result, true
		}
	}

	// Try direct conversion for numeric types
	if isNumericType(targetType) && isNumericType(value.Type()) {
		// Try to convert using reflection
		if value.Type().ConvertibleTo(targetType) {
			return value.Convert(targetType), true
		}
	}

	// Use string as intermediate conversion
	if value.Type().Kind() != reflect.String && value.Type().ConvertibleTo(reflect.TypeOf("")) {
		// Convert to string first
		strVal := value.Convert(reflect.TypeOf("")).Interface().(string)

		// Then try to convert from string to target type
		switch targetType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if i, err := strconv.ParseInt(strVal, 10, 64); err == nil {
				return reflect.ValueOf(i).Convert(targetType), true
			}
		case reflect.Float32, reflect.Float64:
			if f, err := strconv.ParseFloat(strVal, 64); err == nil {
				return reflect.ValueOf(f).Convert(targetType), true
			}
		case reflect.Bool:
			if b, err := strconv.ParseBool(strVal); err == nil {
				return reflect.ValueOf(b), true
			}
		}
	}

	// Handle map conversions
	if targetType.Kind() == reflect.Map && value.Kind() == reflect.Map {
		keyType := targetType.Key()
		elemType := targetType.Elem()
		result := reflect.MakeMap(targetType)

		for _, key := range value.MapKeys() {
			if convertedKey, ok := optimizedConvertValue(key, keyType); ok {
				if convertedValue, ok := optimizedConvertValue(value.MapIndex(key), elemType); ok {
					result.SetMapIndex(convertedKey, convertedValue)
				}
			}
		}
		return result, true
	}

	// Fall back to original method for compatibility
	return convertValue(value, targetType)
}

// isNumericType checks if the type is a numeric type
func isNumericType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

// convertValue attempts to convert a value to the target type (fallback method)
// nolint:gocyclo // This function handles many type conversion cases
func convertValue(value reflect.Value, targetType reflect.Type) (reflect.Value, bool) {
	// If directly assignable, return as is
	if value.Type().AssignableTo(targetType) {
		return value, true
	}

	// Handle basic type conversions
	switch targetType.Kind() {
	case reflect.String:
		return reflect.ValueOf(fmt.Sprintf("%v", value.Interface())), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if value.Kind() == reflect.Float64 {
			// Convert float64 to int
			intVal := int64(value.Float())
			return reflect.ValueOf(intVal).Convert(targetType), true
		} else if value.Kind() == reflect.String {
			if i, err := strconv.ParseInt(value.String(), 10, 64); err == nil {
				return reflect.ValueOf(i).Convert(targetType), true
			}
		} else if value.CanInt() {
			// Use CanInt for any numeric type that can be represented as an int
			return reflect.ValueOf(value.Int()).Convert(targetType), true
		} else if value.CanFloat() {
			// Convert any float value to int
			return reflect.ValueOf(int64(value.Float())).Convert(targetType), true
		} else {
			// Try a last-ditch effort by going through string representation
			str := fmt.Sprintf("%v", value.Interface())
			if i, err := strconv.ParseInt(str, 10, 64); err == nil {
				return reflect.ValueOf(i).Convert(targetType), true
			}
		}
	case reflect.Float32, reflect.Float64:
		if value.Kind() == reflect.Int || value.Kind() == reflect.Int64 {
			return reflect.ValueOf(float64(value.Int())).Convert(targetType), true
		} else if value.Kind() == reflect.String {
			if f, err := strconv.ParseFloat(value.String(), 64); err == nil {
				return reflect.ValueOf(f).Convert(targetType), true
			}
		} else if value.CanFloat() {
			return reflect.ValueOf(value.Float()).Convert(targetType), true
		}
	case reflect.Bool:
		if value.Kind() == reflect.String {
			if b, err := strconv.ParseBool(value.String()); err == nil {
				return reflect.ValueOf(b), true
			}
		} else if value.Kind() == reflect.Float64 {
			// Non-zero is true, zero is false
			return reflect.ValueOf(value.Float() != 0), true
		} else if value.Kind() == reflect.Int64 {
			// Non-zero is true, zero is false
			return reflect.ValueOf(value.Int() != 0), true
		}
	}

	// Handle slice conversions
	if targetType.Kind() == reflect.Slice && value.Kind() == reflect.Slice {
		elemType := targetType.Elem()
		length := value.Len()
		result := reflect.MakeSlice(targetType, length, length)

		for i := 0; i < length; i++ {
			if convertedItem, ok := convertValue(value.Index(i), elemType); ok {
				result.Index(i).Set(convertedItem)
			} else {
				return reflect.Value{}, false
			}
		}
		return result, true
	}

	// Handle map conversions
	if targetType.Kind() == reflect.Map && value.Kind() == reflect.Map {
		keyType := targetType.Key()
		elemType := targetType.Elem()
		result := reflect.MakeMap(targetType)

		for _, key := range value.MapKeys() {
			if convertedKey, ok := convertValue(key, keyType); ok {
				if convertedValue, ok := convertValue(value.MapIndex(key), elemType); ok {
					result.SetMapIndex(convertedKey, convertedValue)
				}
			}
		}
		return result, true
	}

	return reflect.Value{}, false
}

// callFunction calls the function with the provided arguments
func (t *Tool) callFunction(args []reflect.Value) (interface{}, error) {
	// Call the function
	results := t.fnValue.Call(args)

	// Check the results
	if len(results) == 0 {
		return nil, nil
	}

	// Get the actual result
	var result interface{}
	if len(results) > 0 && results[0].IsValid() {
		result = results[0].Interface()
	}

	// Check for an error
	var err error
	if len(results) > 1 && results[1].IsValid() && !results[1].IsNil() {
		err = results[1].Interface().(error)
	}

	return result, err
}

// OutputSchema returns the schema for the tool output
func (t *Tool) OutputSchema() *sdomain.Schema {
	return t.outputSchema
}

// UsageInstructions returns detailed instructions on when and how to use the tool
func (t *Tool) UsageInstructions() string {
	return t.usageInstructions
}

// Examples returns concrete examples showing tool usage
func (t *Tool) Examples() []domain.ToolExample {
	return t.examples
}

// Constraints returns limitations and constraints of the tool
func (t *Tool) Constraints() []string {
	return t.constraints
}

// ErrorGuidance returns a map of error types to helpful guidance
func (t *Tool) ErrorGuidance() map[string]string {
	return t.errorGuidance
}

// Category returns the category for grouping
func (t *Tool) Category() string {
	return t.category
}

// Tags returns tags for discovery and filtering
func (t *Tool) Tags() []string {
	return t.tags
}

// Version returns the tool version for compatibility tracking
func (t *Tool) Version() string {
	return t.version
}

// IsDeterministic returns whether the same input always produces same output
func (t *Tool) IsDeterministic() bool {
	return t.isDeterministic
}

// IsDestructive returns whether the tool modifies state or has side effects
func (t *Tool) IsDestructive() bool {
	return t.isDestructive
}

// RequiresConfirmation returns whether user confirmation is needed before execution
func (t *Tool) RequiresConfirmation() bool {
	return t.requiresConfirmation
}

// EstimatedLatency returns the expected execution time
func (t *Tool) EstimatedLatency() string {
	return t.estimatedLatency
}

// ToMCPDefinition exports the tool definition in MCP format
func (t *Tool) ToMCPDefinition() domain.MCPToolDefinition {
	annotations := make(map[string]interface{})

	// Add behavioral metadata
	annotations["deterministic"] = t.isDeterministic
	annotations["destructive"] = t.isDestructive
	annotations["requires_confirmation"] = t.requiresConfirmation
	annotations["estimated_latency"] = t.estimatedLatency

	// Add metadata
	if t.category != "" {
		annotations["category"] = t.category
	}
	if len(t.tags) > 0 {
		annotations["tags"] = t.tags
	}
	if t.version != "" {
		annotations["version"] = t.version
	}

	// Add guidance if present
	if t.usageInstructions != "" {
		annotations["usage_instructions"] = t.usageInstructions
	}
	if len(t.examples) > 0 {
		annotations["examples"] = t.examples
	}
	if len(t.constraints) > 0 {
		annotations["constraints"] = t.constraints
	}
	if len(t.errorGuidance) > 0 {
		annotations["error_guidance"] = t.errorGuidance
	}

	return domain.MCPToolDefinition{
		Name:         t.name,
		Description:  t.description,
		InputSchema:  t.paramSchema,
		OutputSchema: t.outputSchema,
		Annotations:  annotations,
	}
}
