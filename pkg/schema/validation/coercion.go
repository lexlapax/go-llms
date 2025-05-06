package validation

import (
	"strconv"
	"strings"
)

// Coerce attempts to convert a value to the expected type
func (v *DefaultValidator) Coerce(targetType string, value interface{}) (interface{}, bool) {
	switch targetType {
	case "string":
		return coerceToString(value)
	case "integer":
		return coerceToInteger(value)
	case "number":
		return coerceToNumber(value)
	case "boolean":
		return coerceToBoolean(value)
	default:
		return value, false
	}
}

// coerceToString attempts to convert a value to a string
func coerceToString(value interface{}) (interface{}, bool) {
	switch v := value.(type) {
	case string:
		return v, true
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), true
	case int:
		return strconv.Itoa(v), true
	case int64:
		return strconv.FormatInt(v, 10), true
	case bool:
		return strconv.FormatBool(v), true
	default:
		return nil, false
	}
}

// coerceToInteger attempts to convert a value to an integer
func coerceToInteger(value interface{}) (interface{}, bool) {
	switch v := value.(type) {
	case int:
		return int64(v), true
	case int64:
		return v, true
	case float64:
		return int64(v), true
	case string:
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i, true
		}
		return nil, false
	case bool:
		if v {
			return int64(1), true
		}
		return int64(0), true
	default:
		return nil, false
	}
}

// coerceToNumber attempts to convert a value to a number
func coerceToNumber(value interface{}) (interface{}, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case float64:
		return v, true
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, true
		}
		return nil, false
	case bool:
		if v {
			return float64(1), true
		}
		return float64(0), true
	default:
		return nil, false
	}
}

// coerceToBoolean attempts to convert a value to a boolean
func coerceToBoolean(value interface{}) (interface{}, bool) {
	switch v := value.(type) {
	case bool:
		return v, true
	case string:
		// Handle case insensitive true/false strings
		lowered := strings.ToLower(v)
		if lowered == "true" || lowered == "yes" || lowered == "1" {
			return true, true
		}
		if lowered == "false" || lowered == "no" || lowered == "0" {
			return false, true
		}
		return nil, false
	case int:
		return v != 0, true
	case int64:
		return v != 0, true
	case float64:
		return v != 0, true
	default:
		return nil, false
	}
}
