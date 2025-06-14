// ABOUTME: Built-in type converters for common Go types and primitives
// ABOUTME: Provides fundamental conversion capabilities for strings, numbers, booleans, slices, and maps

package types

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// StringConverter handles conversions to/from string types
type StringConverter struct{}

func (c *StringConverter) Name() string  { return "StringConverter" }
func (c *StringConverter) Priority() int { return 100 }

func (c *StringConverter) CanConvert(fromType, toType reflect.Type) bool {
	return toType.Kind() == reflect.String || fromType.Kind() == reflect.String
}

func (c *StringConverter) CanReverse(fromType, toType reflect.Type) bool {
	return true // Most conversions to/from string are reversible
}

func (c *StringConverter) Convert(from any, toType reflect.Type) (any, error) {
	if toType.Kind() == reflect.String {
		return c.convertToString(from)
	}

	if fromType := reflect.TypeOf(from); fromType.Kind() == reflect.String {
		return c.convertFromString(from.(string), toType)
	}

	return nil, NewConversionError(reflect.TypeOf(from), toType, from, "not a string conversion", nil)
}

func (c *StringConverter) convertToString(from any) (string, error) {
	switch v := from.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case time.Time:
		return v.Format(time.RFC3339), nil
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v), nil
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v), nil
	case float32, float64:
		return fmt.Sprintf("%g", v), nil
	case bool:
		return strconv.FormatBool(v), nil
	case fmt.Stringer:
		return v.String(), nil
	default:
		// Fallback to JSON marshaling for complex types
		bytes, err := json.Marshal(v)
		if err != nil {
			return "", NewConversionError(reflect.TypeOf(from), reflect.TypeOf(""), from, "json marshal failed", err)
		}
		return string(bytes), nil
	}
}

func (c *StringConverter) convertFromString(from string, toType reflect.Type) (any, error) {
	switch toType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val, err := strconv.ParseInt(from, 10, int(toType.Size()*8))
		if err != nil {
			return nil, NewConversionError(reflect.TypeOf(from), toType, from, "invalid integer", err)
		}
		return reflect.ValueOf(val).Convert(toType).Interface(), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := strconv.ParseUint(from, 10, int(toType.Size()*8))
		if err != nil {
			return nil, NewConversionError(reflect.TypeOf(from), toType, from, "invalid unsigned integer", err)
		}
		return reflect.ValueOf(val).Convert(toType).Interface(), nil

	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(from, int(toType.Size()*8))
		if err != nil {
			return nil, NewConversionError(reflect.TypeOf(from), toType, from, "invalid float", err)
		}
		return reflect.ValueOf(val).Convert(toType).Interface(), nil

	case reflect.Bool:
		val, err := strconv.ParseBool(from)
		if err != nil {
			return nil, NewConversionError(reflect.TypeOf(from), toType, from, "invalid boolean", err)
		}
		return val, nil

	default:
		// Try JSON unmarshaling for complex types
		result := reflect.New(toType).Interface()
		err := json.Unmarshal([]byte(from), result)
		if err != nil {
			return nil, NewConversionError(reflect.TypeOf(from), toType, from, "json unmarshal failed", err)
		}
		return reflect.ValueOf(result).Elem().Interface(), nil
	}
}

// NumberConverter handles conversions between numeric types
type NumberConverter struct{}

func (c *NumberConverter) Name() string  { return "NumberConverter" }
func (c *NumberConverter) Priority() int { return 95 }

func (c *NumberConverter) CanConvert(fromType, toType reflect.Type) bool {
	return isNumeric(fromType) && isNumeric(toType)
}

func (c *NumberConverter) CanReverse(fromType, toType reflect.Type) bool {
	return true
}

func (c *NumberConverter) Convert(from any, toType reflect.Type) (any, error) {
	fromValue := reflect.ValueOf(from)

	if !fromValue.Type().ConvertibleTo(toType) {
		return nil, NewConversionError(fromValue.Type(), toType, from, "numeric types not convertible", nil)
	}

	converted := fromValue.Convert(toType)
	return converted.Interface(), nil
}

// SliceConverter handles conversions between slice types and to/from []any
type SliceConverter struct{}

func (c *SliceConverter) Name() string  { return "SliceConverter" }
func (c *SliceConverter) Priority() int { return 90 }

func (c *SliceConverter) CanConvert(fromType, toType reflect.Type) bool {
	return (fromType.Kind() == reflect.Slice) &&
		(toType.Kind() == reflect.Slice || toType == reflect.TypeOf([]any{}))
}

func (c *SliceConverter) CanReverse(fromType, toType reflect.Type) bool {
	return fromType.Kind() == reflect.Slice && toType.Kind() == reflect.Slice
}

func (c *SliceConverter) Convert(from any, toType reflect.Type) (any, error) {
	fromValue := reflect.ValueOf(from)

	if fromValue.Kind() != reflect.Slice {
		return nil, NewConversionError(fromValue.Type(), toType, from, "source is not a slice", nil)
	}

	if toType == reflect.TypeOf([]any{}) {
		// Convert to []any
		result := make([]any, fromValue.Len())
		for i := 0; i < fromValue.Len(); i++ {
			result[i] = fromValue.Index(i).Interface()
		}
		return result, nil
	}

	if toType.Kind() == reflect.Slice {
		// Convert between slice types
		elementType := toType.Elem()
		result := reflect.MakeSlice(toType, fromValue.Len(), fromValue.Len())

		for i := 0; i < fromValue.Len(); i++ {
			srcElement := fromValue.Index(i)
			if srcElement.Type().ConvertibleTo(elementType) {
				result.Index(i).Set(srcElement.Convert(elementType))
			} else {
				return nil, NewConversionError(fromValue.Type(), toType, from,
					fmt.Sprintf("cannot convert slice element from %v to %v", srcElement.Type(), elementType), nil)
			}
		}

		return result.Interface(), nil
	}

	return nil, NewConversionError(fromValue.Type(), toType, from, "unsupported slice conversion", nil)
}

// MapConverter handles conversions between map types and to/from map[string]any
type MapConverter struct{}

func (c *MapConverter) Name() string  { return "MapConverter" }
func (c *MapConverter) Priority() int { return 85 }

func (c *MapConverter) CanConvert(fromType, toType reflect.Type) bool {
	return (fromType.Kind() == reflect.Map) &&
		(toType.Kind() == reflect.Map || toType == reflect.TypeOf(map[string]any{}))
}

func (c *MapConverter) CanReverse(fromType, toType reflect.Type) bool {
	return fromType.Kind() == reflect.Map && toType.Kind() == reflect.Map
}

func (c *MapConverter) Convert(from any, toType reflect.Type) (any, error) {
	fromValue := reflect.ValueOf(from)

	if fromValue.Kind() != reflect.Map {
		return nil, NewConversionError(fromValue.Type(), toType, from, "source is not a map", nil)
	}

	if toType == reflect.TypeOf(map[string]any{}) {
		// Convert to map[string]any
		result := make(map[string]any)
		for _, key := range fromValue.MapKeys() {
			keyStr := fmt.Sprintf("%v", key.Interface())
			result[keyStr] = fromValue.MapIndex(key).Interface()
		}
		return result, nil
	}

	if toType.Kind() == reflect.Map {
		// Convert between map types
		keyType := toType.Key()
		valueType := toType.Elem()
		result := reflect.MakeMap(toType)

		for _, key := range fromValue.MapKeys() {
			srcKey := key
			srcValue := fromValue.MapIndex(key)

			// Convert key
			var newKey reflect.Value
			if srcKey.Type().ConvertibleTo(keyType) {
				newKey = srcKey.Convert(keyType)
			} else {
				return nil, NewConversionError(fromValue.Type(), toType, from,
					fmt.Sprintf("cannot convert map key from %v to %v", srcKey.Type(), keyType), nil)
			}

			// Convert value
			var newValue reflect.Value
			if srcValue.Type().ConvertibleTo(valueType) {
				newValue = srcValue.Convert(valueType)
			} else {
				return nil, NewConversionError(fromValue.Type(), toType, from,
					fmt.Sprintf("cannot convert map value from %v to %v", srcValue.Type(), valueType), nil)
			}

			result.SetMapIndex(newKey, newValue)
		}

		return result.Interface(), nil
	}

	return nil, NewConversionError(fromValue.Type(), toType, from, "unsupported map conversion", nil)
}

// JSONConverter handles conversions through JSON marshaling/unmarshaling
type JSONConverter struct{}

func (c *JSONConverter) Name() string  { return "JSONConverter" }
func (c *JSONConverter) Priority() int { return 50 } // Lower priority, fallback converter

func (c *JSONConverter) CanConvert(fromType, toType reflect.Type) bool {
	// Can convert between any types that support JSON marshaling/unmarshaling
	return true
}

func (c *JSONConverter) CanReverse(fromType, toType reflect.Type) bool {
	return true
}

func (c *JSONConverter) Convert(from any, toType reflect.Type) (any, error) {
	// Marshal to JSON
	jsonBytes, err := json.Marshal(from)
	if err != nil {
		return nil, NewConversionError(reflect.TypeOf(from), toType, from, "json marshal failed", err)
	}

	// Unmarshal to target type
	result := reflect.New(toType).Interface()
	err = json.Unmarshal(jsonBytes, result)
	if err != nil {
		return nil, NewConversionError(reflect.TypeOf(from), toType, from, "json unmarshal failed", err)
	}

	return reflect.ValueOf(result).Elem().Interface(), nil
}

// isNumeric checks if a type is numeric
func isNumeric(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}
