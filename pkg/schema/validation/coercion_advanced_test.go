package validation

import (
	"net"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCoerceToDate(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		want     time.Time
		wantBool bool
	}{
		{
			name:     "ISO 8601 date",
			input:    "2023-05-10",
			want:     time.Date(2023, 5, 10, 0, 0, 0, 0, time.UTC),
			wantBool: true,
		},
		{
			name:     "RFC3339",
			input:    "2023-05-10T14:30:00Z",
			want:     time.Date(2023, 5, 10, 14, 30, 0, 0, time.UTC),
			wantBool: true,
		},
		{
			name:     "Unix timestamp as string",
			input:    "1683731400",
			want:     time.Unix(1683731400, 0),
			wantBool: true,
		},
		{
			name:     "Unix timestamp as int",
			input:    1683731400,
			want:     time.Unix(1683731400, 0),
			wantBool: true,
		},
		{
			name:     "Already a time.Time",
			input:    time.Date(2023, 5, 10, 14, 30, 0, 0, time.UTC),
			want:     time.Date(2023, 5, 10, 14, 30, 0, 0, time.UTC),
			wantBool: true,
		},
		{
			name:     "Invalid string",
			input:    "not a date",
			want:     time.Time{},
			wantBool: false,
		},
		{
			name:     "Invalid type",
			input:    []string{"not", "a", "date"},
			want:     time.Time{},
			wantBool: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotBool := CoerceToDate(tt.input)
			if gotBool != tt.wantBool {
				t.Errorf("CoerceToDate() gotBool = %v, want %v", gotBool, tt.wantBool)
			}
			if gotBool && !got.Equal(tt.want) {
				t.Errorf("CoerceToDate() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCoerceToUUID(t *testing.T) {
	// Create a valid UUID for testing
	validUUID := uuid.MustParse("f47ac10b-58cc-0372-8567-0e02b2c3d479")

	tests := []struct {
		name     string
		input    interface{}
		want     uuid.UUID
		wantBool bool
	}{
		{
			name:     "Valid UUID string",
			input:    "f47ac10b-58cc-0372-8567-0e02b2c3d479",
			want:     validUUID,
			wantBool: true,
		},
		{
			name:     "Already a UUID",
			input:    validUUID,
			want:     validUUID,
			wantBool: true,
		},
		{
			name:     "Invalid UUID string",
			input:    "not-a-uuid",
			want:     uuid.UUID{},
			wantBool: false,
		},
		{
			name:     "Invalid type",
			input:    123,
			want:     uuid.UUID{},
			wantBool: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotBool := CoerceToUUID(tt.input)
			if gotBool != tt.wantBool {
				t.Errorf("CoerceToUUID() gotBool = %v, want %v", gotBool, tt.wantBool)
			}
			if gotBool && got != tt.want {
				t.Errorf("CoerceToUUID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCoerceToEmail(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		want     string
		wantBool bool
	}{
		{
			name:     "Valid email",
			input:    "test@example.com",
			want:     "test@example.com",
			wantBool: true,
		},
		{
			name:     "Email with display name",
			input:    "John Doe <john.doe@example.com>",
			want:     "john.doe@example.com",
			wantBool: true,
		},
		{
			name:     "Invalid email",
			input:    "not-an-email",
			want:     "",
			wantBool: false,
		},
		{
			name:     "Non-string convertible to string",
			input:    123,
			want:     "",
			wantBool: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotBool := CoerceToEmail(tt.input)
			if gotBool != tt.wantBool {
				t.Errorf("CoerceToEmail() gotBool = %v, want %v", gotBool, tt.wantBool)
			}
			if got != tt.want {
				t.Errorf("CoerceToEmail() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCoerceToURL(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		want     string
		wantBool bool
	}{
		{
			name:     "Valid URL with scheme",
			input:    "https://example.com",
			want:     "https://example.com",
			wantBool: true,
		},
		{
			name:     "URL without scheme",
			input:    "example.com",
			want:     "http://example.com",
			wantBool: true,
		},
		{
			name:     "URL with path and query",
			input:    "https://example.com/path?query=value",
			want:     "https://example.com/path?query=value",
			wantBool: true,
		},
		{
			name:     "Invalid URL",
			input:    "not a url",
			want:     "",
			wantBool: false,
		},
		{
			name:     "Number as URL",
			input:    123,
			want:     "http://123",
			wantBool: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotBool := CoerceToURL(tt.input)
			if gotBool != tt.wantBool {
				t.Errorf("CoerceToURL() gotBool = %v, want %v", gotBool, tt.wantBool)
			}
			if gotBool && got.String() != tt.want {
				t.Errorf("CoerceToURL() got = %v, want %v", got.String(), tt.want)
			}
		})
	}
}

func TestCoerceToDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		want     time.Duration
		wantBool bool
	}{
		{
			name:     "Go duration format",
			input:    "1h30m",
			want:     time.Hour + 30*time.Minute,
			wantBool: true,
		},
		{
			name:     "Hours only",
			input:    "2h",
			want:     2 * time.Hour,
			wantBool: true,
		},
		{
			name:     "Minutes only",
			input:    "45m",
			want:     45 * time.Minute,
			wantBool: true,
		},
		{
			name:     "Seconds only",
			input:    "30s",
			want:     30 * time.Second,
			wantBool: true,
		},
		{
			name:     "HH:MM format",
			input:    "1:30",
			want:     time.Hour + 30*time.Minute,
			wantBool: true,
		},
		{
			name:     "HH:MM:SS format",
			input:    "1:30:45",
			want:     time.Hour + 30*time.Minute + 45*time.Second,
			wantBool: true,
		},
		{
			name:     "Natural language - hours",
			input:    "2 hours",
			want:     2 * time.Hour,
			wantBool: true,
		},
		{
			name:     "Natural language - minutes",
			input:    "45 minutes",
			want:     45 * time.Minute,
			wantBool: true,
		},
		{
			name:     "Natural language - day",
			input:    "1 day",
			want:     24 * time.Hour,
			wantBool: true,
		},
		{
			name:     "Integer as milliseconds",
			input:    1500,
			want:     1500 * time.Millisecond,
			wantBool: true,
		},
		{
			name:     "Already a duration",
			input:    2*time.Hour + 30*time.Minute,
			want:     2*time.Hour + 30*time.Minute,
			wantBool: true,
		},
		{
			name:     "Invalid format",
			input:    "not a duration",
			want:     0,
			wantBool: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotBool := CoerceToDuration(tt.input)
			if gotBool != tt.wantBool {
				t.Errorf("CoerceToDuration() gotBool = %v, want %v", gotBool, tt.wantBool)
			}
			if got != tt.want {
				t.Errorf("CoerceToDuration() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCoerceToIP(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		want     string
		wantBool bool
	}{
		{
			name:     "Valid IPv4",
			input:    "192.168.1.1",
			want:     "192.168.1.1",
			wantBool: true,
		},
		{
			name:     "Valid IPv6",
			input:    "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			want:     "2001:db8:85a3::8a2e:370:7334",
			wantBool: true,
		},
		{
			name:     "Invalid IP",
			input:    "not an ip",
			want:     "",
			wantBool: false,
		},
		{
			name:     "Already an IP",
			input:    net.ParseIP("192.168.1.1"),
			want:     "192.168.1.1",
			wantBool: true,
		},
		{
			name:     "Non-string convertible to string",
			input:    123,
			want:     "",
			wantBool: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotBool := CoerceToIP(tt.input)
			if gotBool != tt.wantBool {
				t.Errorf("CoerceToIP() gotBool = %v, want %v", gotBool, tt.wantBool)
			}
			if gotBool && got.String() != tt.want {
				t.Errorf("CoerceToIP() got = %v, want %v", got.String(), tt.want)
			}
		})
	}
}

func TestCoerceToArray(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		want     []interface{}
		wantBool bool
	}{
		{
			name:     "Already an array",
			input:    []interface{}{"a", "b", "c"},
			want:     []interface{}{"a", "b", "c"},
			wantBool: true,
		},
		{
			name:     "JSON array string",
			input:    `["a","b","c"]`,
			want:     []interface{}{"a", "b", "c"},
			wantBool: true,
		},
		{
			name:     "Comma-separated string",
			input:    "a, b, c",
			want:     []interface{}{"a", "b", "c"},
			wantBool: true,
		},
		{
			name:     "Single value",
			input:    "value",
			want:     []interface{}{"value"},
			wantBool: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotBool := CoerceToArray(tt.input)
			if gotBool != tt.wantBool {
				t.Errorf("CoerceToArray() gotBool = %v, want %v", gotBool, tt.wantBool)
			}
			if gotBool {
				if len(got) != len(tt.want) {
					t.Errorf("CoerceToArray() got len = %v, want len = %v", len(got), len(tt.want))
					return
				}
				for i := range got {
					if got[i] != tt.want[i] {
						t.Errorf("CoerceToArray() at index %d got = %v, want %v", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}

func TestCoerceToObject(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		want     map[string]interface{}
		wantBool bool
	}{
		{
			name:     "Already an object",
			input:    map[string]interface{}{"key": "value"},
			want:     map[string]interface{}{"key": "value"},
			wantBool: true,
		},
		{
			name:     "JSON object string",
			input:    `{"key":"value"}`,
			want:     map[string]interface{}{"key": "value"},
			wantBool: true,
		},
		{
			name:     "Invalid JSON string",
			input:    "not json",
			want:     nil,
			wantBool: false,
		},
		{
			name:     "Non-object type",
			input:    123,
			want:     nil,
			wantBool: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotBool := CoerceToObject(tt.input)
			if gotBool != tt.wantBool {
				t.Errorf("CoerceToObject() gotBool = %v, want %v", gotBool, tt.wantBool)
			}
			if gotBool {
				if len(got) != len(tt.want) {
					t.Errorf("CoerceToObject() got len = %v, want len = %v", len(got), len(tt.want))
					return
				}
				// Check if all keys and values match
				for k, v := range tt.want {
					if gotV, ok := got[k]; !ok || gotV != v {
						t.Errorf("CoerceToObject() key %s got = %v, want %v", k, gotV, v)
					}
				}
			}
		})
	}
}