// ABOUTME: Tests for unified authentication middleware
// ABOUTME: Validates auth detection, application, and various auth schemes

package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockStateReader implements StateReader for testing
type mockStateReader struct {
	values map[string]interface{}
}

func (m *mockStateReader) Get(key string) (interface{}, bool) {
	val, exists := m.values[key]
	return val, exists
}

func TestApplyAuth(t *testing.T) {
	tests := []struct {
		name          string
		auth          map[string]interface{}
		expectedError bool
		checkFunc     func(*http.Request) error
	}{
		{
			name: "API Key in header",
			auth: map[string]interface{}{
				"type":         "api_key",
				"api_key":      "test-key-123",
				"key_location": "header",
				"key_name":     "X-API-Key",
			},
			expectedError: false,
			checkFunc: func(req *http.Request) error {
				if req.Header.Get("X-API-Key") != "test-key-123" {
					return fmt.Errorf("expected X-API-Key header to be 'test-key-123', got '%s'", req.Header.Get("X-API-Key"))
				}
				return nil
			},
		},
		{
			name: "API Key in query",
			auth: map[string]interface{}{
				"type":         "api_key",
				"api_key":      "test-key-456",
				"key_location": "query",
				"key_name":     "api_key",
			},
			expectedError: false,
			checkFunc: func(req *http.Request) error {
				if req.URL.Query().Get("api_key") != "test-key-456" {
					return fmt.Errorf("expected api_key query param to be 'test-key-456', got '%s'", req.URL.Query().Get("api_key"))
				}
				return nil
			},
		},
		{
			name: "Bearer token",
			auth: map[string]interface{}{
				"type":  "bearer",
				"token": "bearer-token-789",
			},
			expectedError: false,
			checkFunc: func(req *http.Request) error {
				expected := "Bearer bearer-token-789"
				if req.Header.Get("Authorization") != expected {
					return fmt.Errorf("expected Authorization header to be '%s', got '%s'", expected, req.Header.Get("Authorization"))
				}
				return nil
			},
		},
		{
			name: "Basic auth",
			auth: map[string]interface{}{
				"type":     "basic",
				"username": "testuser",
				"password": "testpass",
			},
			expectedError: false,
			checkFunc: func(req *http.Request) error {
				username, password, ok := req.BasicAuth()
				if !ok {
					return fmt.Errorf("expected basic auth to be set")
				}
				if username != "testuser" || password != "testpass" {
					return fmt.Errorf("expected basic auth testuser:testpass, got %s:%s", username, password)
				}
				return nil
			},
		},
		{
			name: "Missing auth type",
			auth: map[string]interface{}{
				"api_key": "test",
			},
			expectedError: true,
		},
		{
			name: "Invalid auth type",
			auth: map[string]interface{}{
				"type": "invalid",
			},
			expectedError: true,
		},
		{
			name: "API key missing key",
			auth: map[string]interface{}{
				"type": "api_key",
			},
			expectedError: true,
		},
		{
			name: "Bearer missing token",
			auth: map[string]interface{}{
				"type": "bearer",
			},
			expectedError: true,
		},
		{
			name: "Basic missing username",
			auth: map[string]interface{}{
				"type":     "basic",
				"password": "test",
			},
			expectedError: true,
		},
		{
			name: "OAuth2 with access token",
			auth: map[string]interface{}{
				"type":         "oauth2",
				"access_token": "oauth-token-123",
			},
			expectedError: false,
			checkFunc: func(req *http.Request) error {
				expected := "Bearer oauth-token-123"
				if req.Header.Get("Authorization") != expected {
					return fmt.Errorf("expected Authorization header to be '%s', got '%s'", expected, req.Header.Get("Authorization"))
				}
				return nil
			},
		},
		{
			name: "OAuth2 missing access token",
			auth: map[string]interface{}{
				"type": "oauth2",
			},
			expectedError: true,
		},
		{
			name: "Custom header auth",
			auth: map[string]interface{}{
				"type":         "custom",
				"header_name":  "X-Custom-Auth",
				"header_value": "custom-secret",
			},
			expectedError: false,
			checkFunc: func(req *http.Request) error {
				if req.Header.Get("X-Custom-Auth") != "custom-secret" {
					return fmt.Errorf("expected X-Custom-Auth header to be 'custom-secret', got '%s'", req.Header.Get("X-Custom-Auth"))
				}
				return nil
			},
		},
		{
			name: "Custom header auth with prefix",
			auth: map[string]interface{}{
				"type":         "custom",
				"header_name":  "Authorization",
				"header_value": "12345",
				"prefix":       "Token",
			},
			expectedError: false,
			checkFunc: func(req *http.Request) error {
				expected := "Token 12345"
				if req.Header.Get("Authorization") != expected {
					return fmt.Errorf("expected Authorization header to be '%s', got '%s'", expected, req.Header.Get("Authorization"))
				}
				return nil
			},
		},
		{
			name: "Custom header missing header_name",
			auth: map[string]interface{}{
				"type":         "custom",
				"header_value": "test",
			},
			expectedError: true,
		},
		{
			name: "API key in cookie",
			auth: map[string]interface{}{
				"type":         "api_key",
				"api_key":      "cookie-key-123",
				"key_location": "cookie",
				"key_name":     "session_id",
			},
			expectedError: false,
			checkFunc: func(req *http.Request) error {
				cookie, err := req.Cookie("session_id")
				if err != nil {
					return fmt.Errorf("expected cookie 'session_id' to be set: %v", err)
				}
				if cookie.Value != "cookie-key-123" {
					return fmt.Errorf("expected cookie value to be 'cookie-key-123', got '%s'", cookie.Value)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "https://api.example.com/test", nil)
			err := ApplyAuth(req, tt.auth)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				} else if tt.checkFunc != nil {
					if err := tt.checkFunc(req); err != nil {
						t.Error(err)
					}
				}
			}
		})
	}
}

func TestDetectAuthFromState(t *testing.T) {
	tests := []struct {
		name         string
		stateValues  map[string]interface{}
		baseURL      string
		schemes      map[string]AuthScheme
		expectedAuth *AuthConfig
	}{
		{
			name: "GitHub bearer token detection",
			stateValues: map[string]interface{}{
				"github_token": "ghp_test123",
			},
			baseURL: "https://api.github.com",
			expectedAuth: &AuthConfig{
				Type: "bearer",
				Data: map[string]interface{}{
					"token": "ghp_test123",
				},
			},
		},
		{
			name: "GitHub API key detection",
			stateValues: map[string]interface{}{
				"github_api_key": "ghp_test456",
			},
			baseURL: "https://api.github.com",
			expectedAuth: &AuthConfig{
				Type: "bearer",
				Data: map[string]interface{}{
					"token": "ghp_test456",
				},
			},
		},
		{
			name: "GitLab token detection",
			stateValues: map[string]interface{}{
				"gitlab_token": "glpat_test789",
			},
			baseURL: "https://gitlab.com",
			expectedAuth: &AuthConfig{
				Type: "bearer",
				Data: map[string]interface{}{
					"token": "glpat_test789",
				},
			},
		},
		{
			name: "Generic bearer token",
			stateValues: map[string]interface{}{
				"api_token": "generic_token_123",
			},
			baseURL: "https://example.com/api",
			expectedAuth: &AuthConfig{
				Type: "bearer",
				Data: map[string]interface{}{
					"token": "generic_token_123",
				},
			},
		},
		{
			name: "Generic API key",
			stateValues: map[string]interface{}{
				"api_key": "key_123",
			},
			baseURL: "https://example.com/api",
			expectedAuth: &AuthConfig{
				Type: "api_key",
				Data: map[string]interface{}{
					"api_key":      "key_123",
					"key_location": "header",
					"key_name":     "X-API-Key",
				},
			},
		},
		{
			name: "Scheme-based API key detection",
			stateValues: map[string]interface{}{
				"petstore_api_key": "pet_key_123",
			},
			baseURL: "https://petstore.example.com",
			schemes: map[string]AuthScheme{
				"petstore": {
					Type: "apiKey",
					Name: "api_key",
					In:   "header",
				},
			},
			expectedAuth: &AuthConfig{
				Type: "api_key",
				Data: map[string]interface{}{
					"api_key":      "pet_key_123",
					"key_location": "header",
					"key_name":     "api_key",
				},
			},
		},
		{
			name: "Scheme-based bearer detection",
			stateValues: map[string]interface{}{
				"myapi_token": "bearer_123",
			},
			baseURL: "https://myapi.example.com",
			schemes: map[string]AuthScheme{
				"myapi": {
					Type:   "http",
					Scheme: "bearer",
				},
			},
			expectedAuth: &AuthConfig{
				Type: "bearer",
				Data: map[string]interface{}{
					"token": "bearer_123",
				},
			},
		},
		{
			name: "Basic auth detection",
			stateValues: map[string]interface{}{
				"api_username": "user123",
				"api_password": "pass456",
			},
			baseURL: "https://example.com",
			schemes: map[string]AuthScheme{
				"basic": {
					Type:   "http",
					Scheme: "basic",
				},
			},
			expectedAuth: &AuthConfig{
				Type: "basic",
				Data: map[string]interface{}{
					"username": "user123",
					"password": "pass456",
				},
			},
		},
		{
			name:         "No auth found",
			stateValues:  map[string]interface{}{},
			baseURL:      "https://example.com",
			expectedAuth: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &mockStateReader{values: tt.stateValues}
			result := DetectAuthFromState(state, tt.baseURL, tt.schemes)

			if tt.expectedAuth == nil {
				if result != nil {
					t.Errorf("expected nil auth, got %+v", result)
				}
			} else {
				if result == nil {
					t.Error("expected auth config, got nil")
				} else {
					// Check type
					if result.Type != tt.expectedAuth.Type {
						t.Errorf("expected auth type %s, got %s", tt.expectedAuth.Type, result.Type)
					}

					// Check data fields
					for k, v := range tt.expectedAuth.Data {
						if result.Data[k] != v {
							t.Errorf("expected data[%s] = %v, got %v", k, v, result.Data[k])
						}
					}
				}
			}
		})
	}
}

func TestConvertAuthConfigToMap(t *testing.T) {
	tests := []struct {
		name     string
		config   *AuthConfig
		expected map[string]interface{}
	}{
		{
			name: "Bearer token",
			config: &AuthConfig{
				Type: "bearer",
				Data: map[string]interface{}{
					"token": "test123",
				},
			},
			expected: map[string]interface{}{
				"type":  "bearer",
				"token": "test123",
			},
		},
		{
			name: "API key",
			config: &AuthConfig{
				Type: "api_key",
				Data: map[string]interface{}{
					"api_key":      "key123",
					"key_location": "header",
					"key_name":     "X-API-Key",
				},
			},
			expected: map[string]interface{}{
				"type":         "api_key",
				"api_key":      "key123",
				"key_location": "header",
				"key_name":     "X-API-Key",
			},
		},
		{
			name:     "Nil config",
			config:   nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertAuthConfigToMap(tt.config)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %+v", result)
				}
			} else {
				for k, v := range tt.expected {
					if result[k] != v {
						t.Errorf("expected result[%s] = %v, got %v", k, v, result[k])
					}
				}
			}
		})
	}
}
