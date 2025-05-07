// Package tools provides implementations of agent tools.
package tools

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"sync"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// OptimizedTool provides an optimized implementation of tools with reduced allocations
type OptimizedTool struct {
	name        string
	description string
	fn          interface{}
	paramSchema *sdomain.Schema

	// Pre-computed type information
	fnType        reflect.Type
	fnValue       reflect.Value
	numArgs       int
	hasContext    bool
	nonContextArg int
	
	// Cache for commonly used values to reduce allocations
	argsPool      sync.Pool
}

// NewOptimizedTool creates a new optimized tool from a function
func NewOptimizedTool(name, description string, fn interface{}, paramSchema *sdomain.Schema) domain.Tool {
	fnValue := reflect.ValueOf(fn)
	if fnValue.Kind() != reflect.Func {
		panic("tool function must be a function")
	}

	fnType := fnValue.Type()
	numArgs := fnType.NumIn()
	
	// Determine if the function accepts context as first argument
	hasContext := numArgs > 0 && fnType.In(0).Implements(reflect.TypeOf((*context.Context)(nil)).Elem())
	
	// Calculate index of the first non-context argument
	nonContextArg := 0
	if hasContext {
		nonContextArg = 1
	}

	tool := &OptimizedTool{
		name:          name,
		description:   description,
		fn:            fn,
		paramSchema:   paramSchema,
		fnType:        fnType,
		fnValue:       fnValue,
		numArgs:       numArgs,
		hasContext:    hasContext,
		nonContextArg: nonContextArg,
	}

	// Initialize argument pool
	tool.argsPool = sync.Pool{
		New: func() interface{} {
			return make([]reflect.Value, numArgs)
		},
	}

	return tool
}

// Name returns the tool's name
func (t *OptimizedTool) Name() string {
	return t.name
}

// Description provides information about the tool
func (t *OptimizedTool) Description() string {
	return t.description
}

// ParameterSchema returns the schema for the tool parameters
func (t *OptimizedTool) ParameterSchema() *sdomain.Schema {
	return t.paramSchema
}

// Execute runs the tool with parameters
func (t *OptimizedTool) Execute(ctx context.Context, params interface{}) (interface{}, error) {
	// Get an arguments slice from the pool
	args := t.argsPool.Get().([]reflect.Value)
	defer t.argsPool.Put(args)

	// Debug information (commented out for performance)
	/*
	fmt.Printf("DEBUG: Executing tool %s with params: %+v\n", t.name, params)
	fmt.Printf("DEBUG: Function type: %v, hasContext: %v, nonContextArg: %d, numArgs: %d\n", 
		t.fnType, t.hasContext, t.nonContextArg, t.numArgs)
	*/

	// If function expects a context, set it as the first argument
	if t.hasContext {
		args[0] = reflect.ValueOf(ctx)
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
	err := t.prepareArguments(ctx, params, args)
	if err != nil {
		return nil, fmt.Errorf("error preparing arguments: %w", err)
	}

	// Debug code (commented out for performance)
	/*
	fmt.Println("DEBUG: Prepared arguments:")
	for i, arg := range args {
		fmt.Printf("  Arg[%d]: %v (type: %v)\n", i, arg, arg.Type())
	}
	*/

	// Call the function
	result, err := t.callFunction(args)
	/*
	fmt.Printf("DEBUG: Call result: %v (err: %v)\n", result, err)
	*/
	return result, err
}

// prepareArguments converts the params to the appropriate argument types for the function
func (t *OptimizedTool) prepareArguments(ctx context.Context, params interface{}, args []reflect.Value) error {
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
		
		// Debug code (commented out for performance)
		/*
		fmt.Println("DEBUG: Map keys:")
		for _, key := range paramsValue.MapKeys() {
			fmt.Printf("  Key: %v (type: %v)\n", key, key.Type())
			mapVal := paramsValue.MapIndex(key)
			fmt.Printf("    Value: %v (type: %v)\n", mapVal, mapVal.Type())
		}
		*/
		
		for _, field := range fields {
			if !field.isExported {
				// fmt.Printf("DEBUG: Skipping unexported field: %s\n", field.name)
				continue // Skip unexported fields
			}
			
			// Try to find the field in the map by both the struct field name and JSON name
			var mapFieldValue reflect.Value
			
			jsonKeyValue := reflect.ValueOf(field.jsonName)
			mapFieldValue = paramsValue.MapIndex(jsonKeyValue)
			// fmt.Printf("DEBUG: Looking for JSON key '%s': %v\n", field.jsonName, mapFieldValue.IsValid())
			
			if !mapFieldValue.IsValid() {
				nameKeyValue := reflect.ValueOf(field.name)
				mapFieldValue = paramsValue.MapIndex(nameKeyValue)
				// fmt.Printf("DEBUG: Looking for field name '%s': %v\n", field.name, mapFieldValue.IsValid())
			}
			
			if !mapFieldValue.IsValid() {
				// fmt.Printf("DEBUG: Field not found in map: %s\n", field.name)
				continue // Skip fields not found in the map
			}
			
			// Get the field value and ensure we can set it
			fieldValue := structVal.Field(field.index)
			if !fieldValue.CanSet() {
				// fmt.Printf("DEBUG: Field can't be set: %s\n", field.name)
				continue
			}
			
			/* fmt.Printf("DEBUG: Converting field %s from %v (%v) to %v\n", 
				field.name, mapFieldValue, mapFieldValue.Type(), field.fieldType) */
			
			// Try to convert and set the value (using optimized conversion)
			convertedValue, ok := optimizedConvertValue(mapFieldValue, field.fieldType)
			if ok {
				// fmt.Printf("DEBUG: Setting field %s to %v\n", field.name, convertedValue)
				fieldValue.Set(convertedValue)
			} /* else {
				fmt.Printf("DEBUG: FAILED to convert value for field %s\n", field.name)
			} */
		}
		
		// fmt.Printf("DEBUG: Final struct value: %+v\n", structVal.Interface())
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
func optimizedConvertValue(value reflect.Value, targetType reflect.Type) (reflect.Value, bool) {
	// Special handling for interface{} type
	if value.Type().Kind() == reflect.Interface && !value.IsNil() {
		// Extract the actual value from the interface
		// fmt.Printf("DEBUG: Extracting value from interface: %v (type: %v)\n", value.Interface(), value.Elem().Type())
		return optimizedConvertValue(value.Elem(), targetType)
	}

	// Fast path: if directly assignable, return as is
	if value.Type().AssignableTo(targetType) {
		return value, true
	}
	
	// Check if conversion is possible (using cache)
	if !globalParamCache.canConvert(value.Type(), targetType) {
		// Skip cache check for complex types like slices, as the cache may not have them
		if targetType.Kind() != reflect.Slice && targetType.Kind() != reflect.Array {
			// But don't return false yet - try conversion anyway
		}
	}

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

	// Use the regular conversion as fallback for complex cases
	// This part is less optimized but handles the less common cases
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

// callFunction calls the function with the provided arguments
func (t *OptimizedTool) callFunction(args []reflect.Value) (interface{}, error) {
	// Debug code (commented out for performance)
	// fmt.Printf("DEBUG: Calling function with %d args\n", len(args))
	
	// Call the function
	results := t.fnValue.Call(args)

	// Debug the results
	/*
	fmt.Printf("DEBUG: Function returned %d results\n", len(results))
	for i, res := range results {
		fmt.Printf("DEBUG: Result[%d] = %v (valid: %v, type: %v)\n", 
			i, res, res.IsValid(), res.Type())
	}
	*/

	// Check the results
	if len(results) == 0 {
		return nil, nil
	}

	// Get the actual result
	var result interface{}
	if len(results) > 0 && results[0].IsValid() {
		result = results[0].Interface()
		// fmt.Printf("DEBUG: Final result after Interface(): %v (type: %T)\n", result, result)
	}

	// Check for an error
	var err error
	if len(results) > 1 && results[1].IsValid() && !results[1].IsNil() {
		err = results[1].Interface().(error)
	}

	return result, err
}