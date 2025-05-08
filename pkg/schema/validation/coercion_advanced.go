package validation

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Common date formats for parsing
var dateFormats = []string{
	time.RFC3339,
	time.RFC3339Nano,
	time.RFC1123,
	time.RFC1123Z,
	time.RFC822,
	time.RFC822Z,
	time.ANSIC,
	time.UnixDate,
	time.RubyDate,
	"2006-01-02",                         // ISO 8601 date
	"2006-01-02T15:04:05",                // ISO 8601 with time
	"2006-01-02 15:04:05",                // Common SQL datetime format
	"01/02/2006",                         // MM/DD/YYYY
	"02/01/2006",                         // DD/MM/YYYY
	"Jan 2, 2006",                        // Month Day, Year
	"January 2, 2006",                    // Full Month Day, Year
	"2006-01-02T15:04:05-07:00",          // ISO 8601 with timezone
	"2006-01-02T15:04:05.999999999-07:00", // ISO 8601 with fractional seconds
}

// Common duration formats for parsing
var durationRegexps = []*regexp.Regexp{
	regexp.MustCompile(`^(\d+)h$`),               // Hours
	regexp.MustCompile(`^(\d+)m$`),               // Minutes
	regexp.MustCompile(`^(\d+)s$`),               // Seconds
	regexp.MustCompile(`^(\d+)ms$`),              // Milliseconds
	regexp.MustCompile(`^(\d+)h(\d+)m$`),         // Hours and minutes
	regexp.MustCompile(`^(\d+)h(\d+)m(\d+)s$`),   // Hours, minutes, and seconds
	regexp.MustCompile(`^(\d+)m(\d+)s$`),         // Minutes and seconds
	regexp.MustCompile(`^(\d+):(\d+)$`),          // HH:MM
	regexp.MustCompile(`^(\d+):(\d+):(\d+)$`),    // HH:MM:SS
	regexp.MustCompile(`^(\d+)d(\d+)h$`),         // Days and hours
	regexp.MustCompile(`^(\d+) days?$`),          // Days (natural language)
	regexp.MustCompile(`^(\d+) hours?$`),         // Hours (natural language)
	regexp.MustCompile(`^(\d+) minutes?$`),       // Minutes (natural language)
	regexp.MustCompile(`^(\d+) seconds?$`),       // Seconds (natural language)
}

// CoerceToDate attempts to convert a string to a time.Time object
func CoerceToDate(value interface{}) (time.Time, bool) {
	switch v := value.(type) {
	case time.Time:
		return v, true
	case string:
		// Try to parse the string using various date formats
		for _, format := range dateFormats {
			if t, err := time.Parse(format, v); err == nil {
				return t, true
			}
		}
		
		// Try to parse Unix timestamp if it's numeric
		if timestamp, err := CoerceToInt64(v); err == nil {
			return time.Unix(timestamp, 0), true
		}
		
		return time.Time{}, false
	case float64:
		// Assume Unix timestamp
		return time.Unix(int64(v), 0), true
	case int64:
		// Assume Unix timestamp
		return time.Unix(v, 0), true
	case int:
		// Assume Unix timestamp
		return time.Unix(int64(v), 0), true
	default:
		return time.Time{}, false
	}
}

// CoerceToUUID attempts to convert a string to a UUID
func CoerceToUUID(value interface{}) (uuid.UUID, bool) {
	switch v := value.(type) {
	case uuid.UUID:
		return v, true
	case string:
		// Try to parse the string as a UUID
		if u, err := uuid.Parse(v); err == nil {
			return u, true
		}
		return uuid.UUID{}, false
	default:
		return uuid.UUID{}, false
	}
}

// CoerceToEmail validates and normalizes an email address
func CoerceToEmail(value interface{}) (string, bool) {
	switch v := value.(type) {
	case string:
		// Parse the email address
		addr, err := mail.ParseAddress(v)
		if err != nil {
			return "", false
		}
		// Return the normalized email address
		return addr.Address, true
	default:
		// Try to convert to string first
		if str, ok := CoerceToString(value); ok {
			return CoerceToEmail(str)
		}
		return "", false
	}
}

// CoerceToURL validates and normalizes a URL
func CoerceToURL(value interface{}) (*url.URL, bool) {
	switch v := value.(type) {
	case *url.URL:
		return v, true
	case url.URL:
		return &v, true
	case string:
		// Check if the URL has a scheme
		if !strings.Contains(v, "://") {
			// If not, assume http://
			v = "http://" + v
		}
		
		// Parse the URL
		if u, err := url.Parse(v); err == nil && u.Host != "" {
			return u, true
		}
		return nil, false
	default:
		// Only attempt string coercion for numbers
		switch v.(type) {
		case int, int64, float64:
			if str, ok := CoerceToString(value); ok {
				return CoerceToURL(str)
			}
		}
		return nil, false
	}
}

// CoerceToDuration attempts to convert a value to a time.Duration
func CoerceToDuration(value interface{}) (time.Duration, bool) {
	switch v := value.(type) {
	case time.Duration:
		return v, true
	case int:
		// Assume milliseconds
		return time.Duration(v) * time.Millisecond, true
	case int64:
		// Assume milliseconds
		return time.Duration(v) * time.Millisecond, true
	case float64:
		// Assume milliseconds
		return time.Duration(v) * time.Millisecond, true
	case string:
		// First try standard Go duration format
		if d, err := time.ParseDuration(v); err == nil {
			return d, true
		}
		
		// Try to match custom duration formats
		for _, re := range durationRegexps {
			matches := re.FindStringSubmatch(v)
			if len(matches) > 0 {
				// Different handling based on the regex pattern
				switch len(matches) {
				case 2: // Simple formats like "10h", "20m"
					val := matches[1]
					unit := v[len(val):]
					i, err := CoerceToInt64(val)
					if err != nil {
						continue
					}
					
					switch unit {
					case "h":
						return time.Duration(i) * time.Hour, true
					case "m":
						return time.Duration(i) * time.Minute, true
					case "s":
						return time.Duration(i) * time.Second, true
					case "ms":
						return time.Duration(i) * time.Millisecond, true
					case " day", " days":
						return time.Duration(i) * 24 * time.Hour, true
					case " hour", " hours":
						return time.Duration(i) * time.Hour, true
					case " minute", " minutes":
						return time.Duration(i) * time.Minute, true
					case " second", " seconds":
						return time.Duration(i) * time.Second, true
					}
				case 3: // Formats like "1h30m", "10:30"
					v1, err1 := CoerceToInt64(matches[1])
					v2, err2 := CoerceToInt64(matches[2])
					if err1 != nil || err2 != nil {
						continue
					}
					
					if strings.Contains(v, ":") {
						// HH:MM format
						return time.Duration(v1)*time.Hour + time.Duration(v2)*time.Minute, true
					}
					
					// Formats like "1h30m", "5m30s"
					if strings.Contains(v, "h") && strings.Contains(v, "m") {
						return time.Duration(v1)*time.Hour + time.Duration(v2)*time.Minute, true
					}
					if strings.Contains(v, "m") && strings.Contains(v, "s") {
						return time.Duration(v1)*time.Minute + time.Duration(v2)*time.Second, true
					}
					if strings.Contains(v, "d") && strings.Contains(v, "h") {
						return time.Duration(v1)*24*time.Hour + time.Duration(v2)*time.Hour, true
					}
				case 4: // Formats like "1h30m45s", "10:30:45"
					v1, err1 := CoerceToInt64(matches[1])
					v2, err2 := CoerceToInt64(matches[2])
					v3, err3 := CoerceToInt64(matches[3])
					if err1 != nil || err2 != nil || err3 != nil {
						continue
					}
					
					if strings.Contains(v, ":") {
						// HH:MM:SS format
						return time.Duration(v1)*time.Hour + time.Duration(v2)*time.Minute + time.Duration(v3)*time.Second, true
					}
					
					// Format like "1h30m45s"
					return time.Duration(v1)*time.Hour + time.Duration(v2)*time.Minute + time.Duration(v3)*time.Second, true
				}
			}
		}
		
		return 0, false
	default:
		return 0, false
	}
}

// CoerceToIP attempts to convert a string to a net.IP
func CoerceToIP(value interface{}) (net.IP, bool) {
	switch v := value.(type) {
	case net.IP:
		return v, true
	case string:
		ip := net.ParseIP(v)
		if ip == nil {
			return nil, false
		}
		return ip, true
	default:
		// Try to convert to string first
		if str, ok := CoerceToString(value); ok {
			return CoerceToIP(str)
		}
		return nil, false
	}
}

// CoerceToBase64 validates and decodes a base64 string
func CoerceToBase64(value interface{}) ([]byte, bool) {
	switch v := value.(type) {
	case []byte:
		return v, true
	case string:
		// Try to decode as base64
		decoded, err := base64.StdEncoding.DecodeString(v)
		if err == nil {
			return decoded, true
		}
		// Try URL-safe base64
		decoded, err = base64.URLEncoding.DecodeString(v)
		if err == nil {
			return decoded, true
		}
		// Try URL-safe base64 without padding
		decoded, err = base64.RawURLEncoding.DecodeString(v)
		if err == nil {
			return decoded, true
		}
		return nil, false
	default:
		// Try to convert to string first
		if str, ok := CoerceToString(value); ok {
			return CoerceToBase64(str)
		}
		return nil, false
	}
}

// CoerceToHostname validates a hostname according to RFC 1123
func CoerceToHostname(value interface{}) (string, bool) {
	switch v := value.(type) {
	case string:
		// Simple hostname validation based on RFC 1123
		// Hostname must be at least 1 character and at most 253 characters
		if len(v) < 1 || len(v) > 253 {
			return "", false
		}
		
		// Each label must be between 1 and 63 characters and consist of letters, numbers, and hyphens
		// Labels cannot start or end with hyphens
		labels := strings.Split(v, ".")
		for _, label := range labels {
			if len(label) < 1 || len(label) > 63 {
				return "", false
			}
			if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
				return "", false
			}
			for _, r := range label {
				if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-') {
					return "", false
				}
			}
		}
		
		return v, true
	default:
		// Try to convert to string first
		if str, ok := CoerceToString(value); ok {
			return CoerceToHostname(str)
		}
		return "", false
	}
}

// CoerceToIPv4 validates that a string is a valid IPv4 address
func CoerceToIPv4(value interface{}) (net.IP, bool) {
	switch v := value.(type) {
	case net.IP:
		// Check if this is an IPv4 address
		if v.To4() != nil {
			return v, true
		}
		return nil, false
	case string:
		ip := net.ParseIP(v)
		if ip == nil || ip.To4() == nil {
			return nil, false
		}
		return ip, true
	default:
		// Try to convert to string first
		if str, ok := CoerceToString(value); ok {
			return CoerceToIPv4(str)
		}
		return nil, false
	}
}

// CoerceToIPv6 validates that a string is a valid IPv6 address
func CoerceToIPv6(value interface{}) (net.IP, bool) {
	switch v := value.(type) {
	case net.IP:
		// Check if this is an IPv6 address (but not IPv4-mapped)
		if v.To4() == nil && len(v) == net.IPv6len {
			return v, true
		}
		return nil, false
	case string:
		ip := net.ParseIP(v)
		if ip == nil || ip.To4() != nil {
			return nil, false
		}
		return ip, true
	default:
		// Try to convert to string first
		if str, ok := CoerceToString(value); ok {
			return CoerceToIPv6(str)
		}
		return nil, false
	}
}

// CoerceToJSON validates that a string is valid JSON
func CoerceToJSON(value interface{}) (interface{}, bool) {
	switch v := value.(type) {
	case string:
		var result interface{}
		err := json.Unmarshal([]byte(v), &result)
		if err != nil {
			return nil, false
		}
		return result, true
	default:
		// If not a string, check if it's already a valid JSON type
		switch v.(type) {
		case map[string]interface{}, []interface{}, string, float64, bool, nil:
			return v, true
		default:
			// Try to convert to string first
			if str, ok := CoerceToString(value); ok {
				return CoerceToJSON(str)
			}
			return nil, false
		}
	}
}

// CoerceToArray attempts to convert a value to an array
func CoerceToArray(value interface{}) ([]interface{}, bool) {
	switch v := value.(type) {
	case []interface{}:
		return v, true
	case []string:
		// Convert []string to []interface{}
		result := make([]interface{}, len(v))
		for i, s := range v {
			result[i] = s
		}
		return result, true
	case []int:
		// Convert []int to []interface{}
		result := make([]interface{}, len(v))
		for i, n := range v {
			result[i] = n
		}
		return result, true
	case string:
		// Try to parse as JSON array
		var result []interface{}
		err := json.Unmarshal([]byte(v), &result)
		if err != nil {
			// If not JSON, split by commas as a fallback
			if strings.Contains(v, ",") {
				parts := strings.Split(v, ",")
				result = make([]interface{}, len(parts))
				for i, part := range parts {
					result[i] = strings.TrimSpace(part)
				}
				return result, true
			}
			// For a single string value, wrap it in an array
			return []interface{}{v}, true
		}
		return result, true
	default:
		// Try to wrap a single value in an array
		return []interface{}{value}, true
	}
}

// CoerceToObject attempts to convert a value to a map
func CoerceToObject(value interface{}) (map[string]interface{}, bool) {
	switch v := value.(type) {
	case map[string]interface{}:
		return v, true
	case string:
		// Try to parse as JSON object
		var result map[string]interface{}
		err := json.Unmarshal([]byte(v), &result)
		if err != nil {
			return nil, false
		}
		return result, true
	default:
		return nil, false
	}
}

// CoerceToInt64 is a helper function that converts a value to int64
func CoerceToInt64(value interface{}) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", value)
	}
}

// CoerceToString is a helper function that converts a value to string
func CoerceToString(value interface{}) (string, bool) {
	result, ok := coerceToString(value)
	if !ok {
		return "", false
	}
	return result.(string), true
}