// ABOUTME: OAuth2 authentication flows and token management for API client
// ABOUTME: Supports client credentials, authorization code flows, and token refresh

package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// OAuth2Config represents OAuth2 configuration
type OAuth2Config struct {
	TokenURL     string `json:"token_url"`
	AuthURL      string `json:"auth_url,omitempty"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret,omitempty"`
	RedirectURI  string `json:"redirect_uri,omitempty"`
	Scope        string `json:"scope,omitempty"`
	Flow         string `json:"flow"` // "client_credentials", "authorization_code"
}

// OAuth2Token represents an OAuth2 token response
type OAuth2Token struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresIn    int       `json:"expires_in,omitempty"`
	Scope        string    `json:"scope,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
}

// IsExpired checks if the token is expired
func (t *OAuth2Token) IsExpired() bool {
	if t.ExpiresAt.IsZero() {
		return false // No expiry information
	}
	// Add 30 second buffer before actual expiry
	return time.Now().After(t.ExpiresAt.Add(-30 * time.Second))
}

// OAuth2TokenManager manages OAuth2 tokens with refresh capability
type OAuth2TokenManager struct {
	config *OAuth2Config
	client *http.Client
}

// NewOAuth2TokenManager creates a new token manager
func NewOAuth2TokenManager(config *OAuth2Config, client *http.Client) *OAuth2TokenManager {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return &OAuth2TokenManager{
		config: config,
		client: client,
	}
}

// GetToken obtains an OAuth2 token using the configured flow
func (m *OAuth2TokenManager) GetToken(ctx context.Context, params map[string]string) (*OAuth2Token, error) {
	switch m.config.Flow {
	case "client_credentials":
		return m.clientCredentialsFlow(ctx)
	case "authorization_code":
		code, ok := params["code"]
		if !ok {
			return nil, fmt.Errorf("authorization code flow requires 'code' parameter")
		}
		return m.authorizationCodeFlow(ctx, code)
	default:
		return nil, fmt.Errorf("unsupported OAuth2 flow: %s", m.config.Flow)
	}
}

// RefreshToken refreshes an OAuth2 token
func (m *OAuth2TokenManager) RefreshToken(ctx context.Context, refreshToken string) (*OAuth2Token, error) {
	if m.config.TokenURL == "" {
		return nil, fmt.Errorf("token URL is required for refresh")
	}

	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	}

	if m.config.ClientID != "" {
		data.Set("client_id", m.config.ClientID)
	}
	if m.config.ClientSecret != "" {
		data.Set("client_secret", m.config.ClientSecret)
	}

	return m.exchangeToken(ctx, data)
}

// clientCredentialsFlow implements OAuth2 client credentials flow
func (m *OAuth2TokenManager) clientCredentialsFlow(ctx context.Context) (*OAuth2Token, error) {
	if m.config.TokenURL == "" {
		return nil, fmt.Errorf("token URL is required for client credentials flow")
	}
	if m.config.ClientID == "" || m.config.ClientSecret == "" {
		return nil, fmt.Errorf("client ID and secret are required for client credentials flow")
	}

	data := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {m.config.ClientID},
		"client_secret": {m.config.ClientSecret},
	}

	if m.config.Scope != "" {
		data.Set("scope", m.config.Scope)
	}

	return m.exchangeToken(ctx, data)
}

// authorizationCodeFlow implements OAuth2 authorization code flow
func (m *OAuth2TokenManager) authorizationCodeFlow(ctx context.Context, code string) (*OAuth2Token, error) {
	if m.config.TokenURL == "" {
		return nil, fmt.Errorf("token URL is required for authorization code flow")
	}
	if m.config.ClientID == "" {
		return nil, fmt.Errorf("client ID is required for authorization code flow")
	}

	data := url.Values{
		"grant_type": {"authorization_code"},
		"code":       {code},
		"client_id":  {m.config.ClientID},
	}

	if m.config.ClientSecret != "" {
		data.Set("client_secret", m.config.ClientSecret)
	}
	if m.config.RedirectURI != "" {
		data.Set("redirect_uri", m.config.RedirectURI)
	}

	return m.exchangeToken(ctx, data)
}

// exchangeToken performs the actual token exchange request
func (m *OAuth2TokenManager) exchangeToken(ctx context.Context, data url.Values) (*OAuth2Token, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", m.config.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Try to parse error response
		var errorResp struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return nil, fmt.Errorf("OAuth2 error: %s - %s", errorResp.Error, errorResp.ErrorDescription)
		}
		return nil, fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	var token OAuth2Token
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	// Calculate expiry time if expires_in is provided
	if token.ExpiresIn > 0 {
		token.ExpiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	}

	return &token, nil
}

// BuildAuthorizationURL builds the authorization URL for the authorization code flow
func (m *OAuth2TokenManager) BuildAuthorizationURL(state string, additionalParams map[string]string) (string, error) {
	if m.config.AuthURL == "" {
		return "", fmt.Errorf("auth URL is required for authorization code flow")
	}
	if m.config.ClientID == "" {
		return "", fmt.Errorf("client ID is required for authorization code flow")
	}

	u, err := url.Parse(m.config.AuthURL)
	if err != nil {
		return "", fmt.Errorf("invalid auth URL: %w", err)
	}

	q := u.Query()
	q.Set("response_type", "code")
	q.Set("client_id", m.config.ClientID)

	if m.config.RedirectURI != "" {
		q.Set("redirect_uri", m.config.RedirectURI)
	}
	if m.config.Scope != "" {
		q.Set("scope", m.config.Scope)
	}
	if state != "" {
		q.Set("state", state)
	}

	// Add any additional parameters (like PKCE challenge)
	for k, v := range additionalParams {
		q.Set(k, v)
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}

// DetectOAuth2Config attempts to detect OAuth2 configuration from state
func DetectOAuth2Config(state StateReader, baseURL string) *OAuth2Config {
	// Check for OAuth2-specific configuration
	if oauth2Config, exists := state.Get("oauth2_config"); exists {
		if config, ok := oauth2Config.(map[string]interface{}); ok {
			return parseOAuth2Config(config)
		}
	}

	// Try to detect based on known patterns
	normalizedURL := strings.ToLower(strings.TrimRight(baseURL, "/"))

	// GitHub OAuth2
	if strings.Contains(normalizedURL, "github") {
		if clientID, exists := state.Get("github_client_id"); exists {
			if clientSecret, exists := state.Get("github_client_secret"); exists {
				return &OAuth2Config{
					TokenURL:     "https://github.com/login/oauth/access_token",
					AuthURL:      "https://github.com/login/oauth/authorize",
					ClientID:     clientID.(string),
					ClientSecret: clientSecret.(string),
					Flow:         "authorization_code",
				}
			}
		}
	}

	// Google OAuth2
	if strings.Contains(normalizedURL, "google") || strings.Contains(normalizedURL, "googleapis") {
		if clientID, exists := state.Get("google_client_id"); exists {
			if clientSecret, exists := state.Get("google_client_secret"); exists {
				return &OAuth2Config{
					TokenURL:     "https://oauth2.googleapis.com/token",
					AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
					ClientID:     clientID.(string),
					ClientSecret: clientSecret.(string),
					Flow:         "authorization_code",
				}
			}
		}
	}

	return nil
}

func parseOAuth2Config(config map[string]interface{}) *OAuth2Config {
	result := &OAuth2Config{}

	if v, ok := config["token_url"].(string); ok {
		result.TokenURL = v
	}
	if v, ok := config["auth_url"].(string); ok {
		result.AuthURL = v
	}
	if v, ok := config["client_id"].(string); ok {
		result.ClientID = v
	}
	if v, ok := config["client_secret"].(string); ok {
		result.ClientSecret = v
	}
	if v, ok := config["redirect_uri"].(string); ok {
		result.RedirectURI = v
	}
	if v, ok := config["scope"].(string); ok {
		result.Scope = v
	}
	if v, ok := config["flow"].(string); ok {
		result.Flow = v
	}

	return result
}

// JWTClaims represents basic JWT claims for token inspection
type JWTClaims struct {
	Exp int64  `json:"exp,omitempty"`
	Iat int64  `json:"iat,omitempty"`
	Sub string `json:"sub,omitempty"`
	Aud string `json:"aud,omitempty"`
	Iss string `json:"iss,omitempty"`
}

// ParseJWTClaims extracts basic claims from a JWT without verification
// This is useful for checking expiry without having the signing key
func ParseJWTClaims(token string) (*JWTClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format")
	}

	// Decode the payload (second part)
	payload, err := base64URLDecode(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT payload: %w", err)
	}

	var claims JWTClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse JWT claims: %w", err)
	}

	return &claims, nil
}

// IsJWTExpired checks if a JWT token is expired based on the exp claim
func IsJWTExpired(token string) bool {
	claims, err := ParseJWTClaims(token)
	if err != nil || claims.Exp == 0 {
		return false // Can't determine expiry
	}

	// Add 30 second buffer
	return time.Now().Unix() > (claims.Exp - 30)
}

// base64URLDecode decodes a base64url encoded string
func base64URLDecode(s string) ([]byte, error) {
	// Add padding if necessary
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}

	// Use RawURLEncoding which handles URL-safe characters
	return base64.RawURLEncoding.DecodeString(s)
}
