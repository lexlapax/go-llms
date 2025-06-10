// ABOUTME: Unified authentication middleware for HTTP requests supporting multiple auth schemes
// ABOUTME: Provides auth detection from state and application to requests for REST, OpenAPI, and GraphQL

package auth

import (
	"fmt"
	"net/http"
)

// AuthConfig represents authentication configuration
type AuthConfig struct {
	Type string                 `json:"type"` // "api_key", "bearer", "basic", "oauth2"
	Data map[string]interface{} `json:"data"` // Auth-specific data
}

// AuthScheme represents an authentication scheme definition (e.g., from OpenAPI)
type AuthScheme struct {
	Type        string `json:"type"`        // "apiKey", "http", "oauth2", "openIdConnect"
	Scheme      string `json:"scheme"`      // For http type: "basic", "bearer", etc.
	Name        string `json:"name"`        // Parameter name (for apiKey)
	In          string `json:"in"`          // Location: "header", "query", "cookie"
	Description string `json:"description"` // Human-readable description
}

// StateReader interface for reading values from agent state
type StateReader interface {
	Get(key string) (interface{}, bool)
}

// ApplyAuth applies authentication configuration to an HTTP request
func ApplyAuth(req *http.Request, auth map[string]interface{}) error {
	authType, ok := auth["type"].(string)
	if !ok {
		return fmt.Errorf("auth type is required")
	}

	switch authType {
	case "api_key":
		return applyAPIKeyAuth(req, auth)
	case "bearer":
		return applyBearerAuth(req, auth)
	case "basic":
		return applyBasicAuth(req, auth)
	case "oauth2":
		return applyOAuth2Auth(req, auth)
	case "custom":
		return applyCustomAuth(req, auth)
	default:
		return fmt.Errorf("unsupported auth type: %s", authType)
	}
}

// DetectAuthFromState attempts to detect authentication based on URL and state
func DetectAuthFromState(state StateReader, baseURL string, schemes map[string]AuthScheme) *AuthConfig {
	// Try generic auth detection first, which includes provider-specific tokens
	if auth := detectGenericAuthWithProviderTokens(state); auth != nil {
		return auth
	}

	// If schemes are provided, try to match against them
	if len(schemes) > 0 {
		return detectFromSchemes(state, schemes)
	}

	// Fall back to basic generic auth detection
	return detectGenericAuth(state)
}

// Private helper functions

func applyAPIKeyAuth(req *http.Request, auth map[string]interface{}) error {
	apiKey, ok := auth["api_key"].(string)
	if !ok || apiKey == "" {
		return fmt.Errorf("api_key is required for api_key auth")
	}

	keyLocation := "header"
	if loc, ok := auth["key_location"].(string); ok {
		keyLocation = loc
	}

	keyName := "X-API-Key"
	if name, ok := auth["key_name"].(string); ok {
		keyName = name
	}

	switch keyLocation {
	case "header":
		req.Header.Set(keyName, apiKey)
	case "query":
		q := req.URL.Query()
		q.Set(keyName, apiKey)
		req.URL.RawQuery = q.Encode()
	case "cookie":
		req.AddCookie(&http.Cookie{
			Name:  keyName,
			Value: apiKey,
		})
	default:
		return fmt.Errorf("invalid key_location: %s", keyLocation)
	}

	return nil
}

func applyBearerAuth(req *http.Request, auth map[string]interface{}) error {
	token, ok := auth["token"].(string)
	if !ok || token == "" {
		return fmt.Errorf("token is required for bearer auth")
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func applyBasicAuth(req *http.Request, auth map[string]interface{}) error {
	username, ok1 := auth["username"].(string)
	password, ok2 := auth["password"].(string)
	if !ok1 || !ok2 || username == "" {
		return fmt.Errorf("username and password are required for basic auth")
	}
	req.SetBasicAuth(username, password)
	return nil
}

// detectFromSchemes attempts to find auth config based on defined schemes
func detectFromSchemes(state StateReader, schemes map[string]AuthScheme) *AuthConfig {
	// Try each scheme in order
	for schemeName, scheme := range schemes {
		switch scheme.Type {
		case "apiKey":
			if auth := detectAPIKeyFromState(state, schemeName, scheme); auth != nil {
				return auth
			}
		case "http":
			if auth := detectHTTPAuthFromState(state, schemeName, scheme); auth != nil {
				return auth
			}
		}
	}
	return nil
}

func detectAPIKeyFromState(state StateReader, schemeName string, scheme AuthScheme) *AuthConfig {
	// Try various key naming patterns
	keyNames := []string{
		fmt.Sprintf("%s_api_key", schemeName),
		fmt.Sprintf("%s_key", schemeName),
		schemeName,
		"api_key",
		"apiKey",
		scheme.Name, // The actual parameter name
	}

	for _, keyName := range keyNames {
		if value, exists := state.Get(keyName); exists {
			if apiKey, ok := value.(string); ok && apiKey != "" {
				return &AuthConfig{
					Type: "api_key",
					Data: map[string]interface{}{
						"api_key":      apiKey,
						"key_location": scheme.In,
						"key_name":     scheme.Name,
					},
				}
			}
		}
	}

	return nil
}

func detectHTTPAuthFromState(state StateReader, schemeName string, scheme AuthScheme) *AuthConfig {
	switch scheme.Scheme {
	case "bearer":
		tokenKeys := []string{
			fmt.Sprintf("%s_token", schemeName),
			fmt.Sprintf("%s_bearer", schemeName),
			schemeName,
			"bearer_token",
			"access_token",
			"token",
		}

		for _, key := range tokenKeys {
			if value, exists := state.Get(key); exists {
				if token, ok := value.(string); ok && token != "" {
					return &AuthConfig{
						Type: "bearer",
						Data: map[string]interface{}{
							"token": token,
						},
					}
				}
			}
		}

	case "basic":
		// Look for username/password pairs
		var username, password string

		usernameKeys := []string{
			fmt.Sprintf("%s_username", schemeName),
			"api_username",
			"username",
		}
		passwordKeys := []string{
			fmt.Sprintf("%s_password", schemeName),
			"api_password",
			"password",
		}

		for _, key := range usernameKeys {
			if value, exists := state.Get(key); exists {
				if u, ok := value.(string); ok && u != "" {
					username = u
					break
				}
			}
		}

		for _, key := range passwordKeys {
			if value, exists := state.Get(key); exists {
				if p, ok := value.(string); ok && p != "" {
					password = p
					break
				}
			}
		}

		if username != "" && password != "" {
			return &AuthConfig{
				Type: "basic",
				Data: map[string]interface{}{
					"username": username,
					"password": password,
				},
			}
		}
	}

	return nil
}

// detectGenericAuthWithProviderTokens tries both generic and provider-specific token patterns
func detectGenericAuthWithProviderTokens(state StateReader) *AuthConfig {
	// Combined list of all possible token keys, including provider-specific ones
	tokenKeys := []string{
		// GitHub tokens
		"github_token",
		"github_api_key",
		"GITHUB_TOKEN",
		"GITHUB_API_KEY",
		"github_personal_access_token",
		"github_pat",
		"gh_token",

		// GitLab tokens
		"gitlab_token",
		"gitlab_api_key",
		"GITLAB_TOKEN",
		"GITLAB_API_KEY",
		"gitlab_personal_access_token",
		"gitlab_pat",

		// Generic tokens
		"api_token",
		"access_token",
		"bearer_token",
		"auth_token",
		"token",
		"API_TOKEN",
		"ACCESS_TOKEN",
		"BEARER_TOKEN",
		"AUTH_TOKEN",
		"TOKEN",
	}

	// Try each token key
	for _, key := range tokenKeys {
		if value, exists := state.Get(key); exists {
			if token, ok := value.(string); ok && token != "" {
				return &AuthConfig{
					Type: "bearer",
					Data: map[string]interface{}{
						"token": token,
					},
				}
			}
		}
	}

	return nil
}

// detectGenericAuth tries common auth patterns
func detectGenericAuth(state StateReader) *AuthConfig {
	// Try API key patterns first (more specific)
	apiKeyNames := []string{
		"api_key",
		"apikey",
		"x_api_key",
		"X_API_KEY",
	}

	for _, key := range apiKeyNames {
		if value, exists := state.Get(key); exists {
			if apiKey, ok := value.(string); ok && apiKey != "" {
				return &AuthConfig{
					Type: "api_key",
					Data: map[string]interface{}{
						"api_key":      apiKey,
						"key_location": "header",
						"key_name":     "X-API-Key",
					},
				}
			}
		}
	}

	// Try bearer token patterns (excluding api_key which is handled above)
	tokenKeys := []string{
		"api_token",
		"access_token",
		"bearer_token",
		"auth_token",
		"token",
	}

	for _, key := range tokenKeys {
		if value, exists := state.Get(key); exists {
			if token, ok := value.(string); ok && token != "" {
				return &AuthConfig{
					Type: "bearer",
					Data: map[string]interface{}{
						"token": token,
					},
				}
			}
		}
	}

	return nil
}

// ConvertAuthConfigToMap converts AuthConfig to map for compatibility
func ConvertAuthConfigToMap(config *AuthConfig) map[string]interface{} {
	if config == nil {
		return nil
	}

	result := map[string]interface{}{
		"type": config.Type,
	}

	// Merge data fields into result
	for k, v := range config.Data {
		result[k] = v
	}

	return result
}

func applyOAuth2Auth(req *http.Request, auth map[string]interface{}) error {
	// Check for access token first (simplest case)
	if accessToken, ok := auth["access_token"].(string); ok && accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
		return nil
	}

	// Handle different OAuth2 flows
	flowType, _ := auth["flow"].(string)
	switch flowType {
	case "client_credentials":
		// Client credentials should have already obtained the token
		// Just use the access_token field
		return fmt.Errorf("client credentials flow requires access_token to be set")
	case "authorization_code":
		// Authorization code flow should have already exchanged code for token
		return fmt.Errorf("authorization code flow requires access_token to be set")
	default:
		// Default to bearer token if available
		if token, ok := auth["token"].(string); ok && token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
			return nil
		}
		return fmt.Errorf("OAuth2 requires access_token or token field")
	}
}

func applyCustomAuth(req *http.Request, auth map[string]interface{}) error {
	// Support custom header authentication
	headerName, ok := auth["header_name"].(string)
	if !ok || headerName == "" {
		return fmt.Errorf("custom auth requires header_name")
	}

	headerValue, ok := auth["header_value"].(string)
	if !ok || headerValue == "" {
		return fmt.Errorf("custom auth requires header_value")
	}

	// Optional prefix (like "Bearer", "Token", etc.)
	if prefix, ok := auth["prefix"].(string); ok && prefix != "" {
		headerValue = prefix + " " + headerValue
	}

	req.Header.Set(headerName, headerValue)
	return nil
}
