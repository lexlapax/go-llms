package validation

import (
	"strconv"
	"strings"
	"time"
)

// Coerce attempts to convert a value to the expected type
func (v *Validator) Coerce(targetType string, value interface{}, format ...string) (interface{}, bool) {
	// Get the optional format
	var formatStr string
	if len(format) > 0 {
		formatStr = format[0]
	}

	// If format is specified, try to coerce based on format
	if formatStr != "" {
		// Check if we're validating a string type
		isStringType := targetType == "string"

		switch formatStr {
		case "date", "date-time":
			date, ok := CoerceToDate(value)
			if !ok {
				return value, false
			}
			// For string validation, return ISO format string
			if isStringType {
				return date.Format(time.RFC3339), true
			}
			return date, true

		case "uuid":
			uuid, ok := CoerceToUUID(value)
			if !ok {
				return value, false
			}
			// For string validation, return UUID string
			if isStringType {
				return uuid.String(), true
			}
			return uuid, true

		case "email":
			email, ok := CoerceToEmail(value)
			if !ok {
				return value, false
			}
			return email, true

		case "uri", "url":
			url, ok := CoerceToURL(value)
			if !ok {
				return value, false
			}
			// For string validation, return URL string
			if isStringType {
				return url.String(), true
			}
			return url, true

		case "duration":
			duration, ok := CoerceToDuration(value)
			if !ok {
				return value, false
			}
			// For string validation, return duration string
			if isStringType {
				return duration.String(), true
			}
			return duration, true

		case "ipv4", "ipv6", "ip":
			ip, ok := CoerceToIP(value)
			if !ok {
				return value, false
			}
			// For string validation, return IP string
			if isStringType {
				return ip.String(), true
			}
			return ip, true
		}
	}

	// Coerce based on target type
	switch targetType {
	case "string":
		return coerceToString(value)
	case "integer":
		return coerceToInteger(value)
	case "number":
		return coerceToNumber(value)
	case "boolean":
		return coerceToBoolean(value)
	case "array":
		return CoerceToArray(value)
	case "object":
		return CoerceToObject(value)
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
