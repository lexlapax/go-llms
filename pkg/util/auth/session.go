// ABOUTME: Session and cookie management for maintaining authentication state
// ABOUTME: Provides cookie jar functionality and session persistence across API calls

package auth

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"
)

// SessionManager manages cookies and session state for API requests
type SessionManager struct {
	jar      *cookiejar.Jar
	sessions map[string]*SessionData
	mu       sync.RWMutex
}

// SessionData represents session information for a domain
type SessionData struct {
	Domain       string                 `json:"domain"`
	SessionID    string                 `json:"session_id,omitempty"`
	Cookies      []*http.Cookie         `json:"cookies,omitempty"`
	AuthData     map[string]interface{} `json:"auth_data,omitempty"`
	LastAccessed time.Time              `json:"last_accessed"`
	ExpiresAt    time.Time              `json:"expires_at,omitempty"`
}

// NewSessionManager creates a new session manager with cookie jar
func NewSessionManager() (*SessionManager, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	return &SessionManager{
		jar:      jar,
		sessions: make(map[string]*SessionData),
	}, nil
}

// ApplySession applies session cookies to a request
func (sm *SessionManager) ApplySession(req *http.Request) {
	// The cookie jar automatically handles cookies
	// No additional action needed as http.Client uses the jar
}

// SaveSession saves session data for a domain
func (sm *SessionManager) SaveSession(domain string, resp *http.Response, authData map[string]interface{}) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	u, err := url.Parse(domain)
	if err != nil {
		return
	}

	// Get cookies from the jar for this domain
	cookies := sm.jar.Cookies(u)

	session := &SessionData{
		Domain:       domain,
		Cookies:      cookies,
		AuthData:     authData,
		LastAccessed: time.Now(),
	}

	// Look for session ID in cookies
	for _, cookie := range cookies {
		if cookie.Name == "session_id" || cookie.Name == "sessionid" || cookie.Name == "PHPSESSID" {
			session.SessionID = cookie.Value
			if !cookie.Expires.IsZero() {
				session.ExpiresAt = cookie.Expires
			}
			break
		}
	}

	sm.sessions[domain] = session
}

// GetSession retrieves session data for a domain
func (sm *SessionManager) GetSession(domain string) (*SessionData, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[domain]
	if !exists {
		return nil, false
	}

	// Check if session is expired
	if !session.ExpiresAt.IsZero() && time.Now().After(session.ExpiresAt) {
		return nil, false
	}

	return session, true
}

// ClearSession removes session data for a domain
func (sm *SessionManager) ClearSession(domain string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.sessions, domain)

	// Clear cookies from jar
	u, err := url.Parse(domain)
	if err != nil {
		return
	}

	// Clear cookies by setting them as expired
	cookies := sm.jar.Cookies(u)
	for _, cookie := range cookies {
		cookie.MaxAge = -1
		cookie.Expires = time.Now().Add(-24 * time.Hour)
	}
	sm.jar.SetCookies(u, cookies)
}

// GetCookieJar returns the underlying cookie jar for use with http.Client
func (sm *SessionManager) GetCookieJar() *cookiejar.Jar {
	return sm.jar
}

// SerializeSessions returns all sessions for persistence
func (sm *SessionManager) SerializeSessions() map[string]*SessionData {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make(map[string]*SessionData)
	for k, v := range sm.sessions {
		result[k] = v
	}
	return result
}

// RestoreSessions restores sessions from serialized data
func (sm *SessionManager) RestoreSessions(sessions map[string]*SessionData) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for domain, session := range sessions {
		// Skip expired sessions
		if !session.ExpiresAt.IsZero() && time.Now().After(session.ExpiresAt) {
			continue
		}

		sm.sessions[domain] = session

		// Restore cookies to jar
		if len(session.Cookies) > 0 {
			u, err := url.Parse(domain)
			if err != nil {
				continue
			}
			sm.jar.SetCookies(u, session.Cookies)
		}
	}

	return nil
}

// ExtractSetCookieHeaders extracts Set-Cookie headers from response
// This is useful for manual cookie management
func ExtractSetCookieHeaders(resp *http.Response) []*http.Cookie {
	return resp.Cookies()
}

// BuildCookieHeader builds a Cookie header value from a slice of cookies
func BuildCookieHeader(cookies []*http.Cookie) string {
	if len(cookies) == 0 {
		return ""
	}

	var cookieStrs []string
	for _, cookie := range cookies {
		cookieStrs = append(cookieStrs, cookie.String())
	}

	return strings.Join(cookieStrs, "; ")
}

// ParseCookies parses cookies from a Cookie header value
func ParseCookies(header string) []*http.Cookie {
	var cookies []*http.Cookie

	pairs := strings.Split(header, ";")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			continue
		}

		cookie := &http.Cookie{
			Name:  strings.TrimSpace(parts[0]),
			Value: strings.TrimSpace(parts[1]),
		}
		cookies = append(cookies, cookie)
	}

	return cookies
}
