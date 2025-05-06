// Package tools provides implementations of agent tools.
package tools

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// BaseTool provides a foundation for tool implementations
type BaseTool struct {
	name        string
	description string
	fn          interface{}
	paramSchema *sdomain.Schema
}

// NewTool creates a new tool from a function
func NewTool(name, description string, fn interface{}, paramSchema *sdomain.Schema) domain.Tool {
	return &BaseTool{
		name:        name,
		description: description,
		fn:          fn,
		paramSchema: paramSchema,
	}
}

// Name returns the tool's name
func (t *BaseTool) Name() string {
	return t.name
}

// Description provides information about the tool
func (t *BaseTool) Description() string {
	return t.description
}

// ParameterSchema returns the schema for the tool parameters
func (t *BaseTool) ParameterSchema() *sdomain.Schema {
	return t.paramSchema
}

// Execute runs the tool with parameters
func (t *BaseTool) Execute(ctx context.Context, params interface{}) (interface{}, error) {
	// Get the function value
	fnValue := reflect.ValueOf(t.fn)
	if fnValue.Kind() != reflect.Func {
		return nil, fmt.Errorf("not a function")
	}

	// Check if we have params
	if params == nil {
		// If the function takes no arguments, call it directly
		if fnValue.Type().NumIn() == 0 {
			return callFunction(fnValue, nil)
		}
		return nil, fmt.Errorf("function requires parameters but none provided")
	}

	// Convert params to appropriate argument types
	args, err := prepareArguments(ctx, fnValue, params)
	if err != nil {
		return nil, fmt.Errorf("error preparing arguments: %w", err)
	}

	// Call the function
	return callFunction(fnValue, args)
}

// prepareArguments converts the params to the appropriate argument types for the function
func prepareArguments(ctx context.Context, fnValue reflect.Value, params interface{}) ([]reflect.Value, error) {
	fnType := fnValue.Type()
	numArgs := fnType.NumIn()
	args := make([]reflect.Value, numArgs)

	// Check if the function expects a context
	startIdx := 0

	if numArgs > 0 && fnType.In(0).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
		args[0] = reflect.ValueOf(ctx)
		startIdx = 1
	}

	// If no more arguments needed besides context, we're done
	if startIdx >= numArgs {
		return args, nil
	}

	// Handle the params based on what was provided
	paramsValue := reflect.ValueOf(params)

	// If params is a map and function expects a struct, try to map fields
	if paramsValue.Kind() == reflect.Map && fnType.In(startIdx).Kind() == reflect.Struct {
		structVal := reflect.New(fnType.In(startIdx)).Elem()
		if err := mapToStruct(paramsValue, structVal); err != nil {
			return nil, err
		}
		args[startIdx] = structVal
		return args, nil
	}

	// If params can be directly assigned to the function's argument type
	if startIdx < numArgs && paramsValue.Type().AssignableTo(fnType.In(startIdx)) {
		args[startIdx] = paramsValue
		return args, nil
	}

	// Try to convert the value
	if startIdx < numArgs {
		if convertedValue, ok := convertValue(paramsValue, fnType.In(startIdx)); ok {
			args[startIdx] = convertedValue
			return args, nil
		}
	}

	return nil, fmt.Errorf("unable to convert parameters to function argument types")
}

// mapToStruct maps values from a map to a struct
func mapToStruct(mapValue, structValue reflect.Value) error {
	if mapValue.Kind() != reflect.Map || structValue.Kind() != reflect.Struct {
		return fmt.Errorf("expected map and struct, got %s and %s", mapValue.Kind(), structValue.Kind())
	}

	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Type().Field(i)
		fieldName := field.Name

		// Check for json tag
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			// Split the tag by comma to handle options like omitempty
			parts := reflect.StructTag(jsonTag).Get("json")
			if parts != "" {
				fieldName = parts
			}
		}

		// Look up the field in the map
		mapFieldValue := mapValue.MapIndex(reflect.ValueOf(fieldName))
		if !mapFieldValue.IsValid() {
			// Try case-insensitive lookup
			for _, key := range mapValue.MapKeys() {
				if key.Kind() == reflect.String &&
					strings.EqualFold(key.String(), fieldName) {
					mapFieldValue = mapValue.MapIndex(key)
					break
				}
			}
		}

		if !mapFieldValue.IsValid() {
			continue // Skip fields not found in the map
		}

		fieldValue := structValue.Field(i)
		if !fieldValue.CanSet() {
			continue // Skip unexported fields
		}

		// Try to convert and set the value
		if convertedValue, ok := convertValue(mapFieldValue, fieldValue.Type()); ok {
			fieldValue.Set(convertedValue)
		}
	}

	return nil
}

// convertValue attempts to convert a value to the target type
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
func callFunction(fnValue reflect.Value, args []reflect.Value) (interface{}, error) {
	// Call the function
	results := fnValue.Call(args)

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
