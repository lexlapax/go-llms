# Security Considerations: Security Best Practices

> **[Project Root](/) / [Documentation](/docs/) / [User Guide](/docs/user-guide/) / [Advanced Topics](/docs/user-guide/advanced/) / Security Considerations**

Comprehensive security guide for Go-LLMs applications covering API key management, data protection, input validation, access control, and compliance requirements.

## Security Overview

Security in LLM applications requires multiple layers:
- **API Key Security** - Protecting provider credentials
- **Data Protection** - Encryption, privacy, retention
- **Input Validation** - Preventing injection and abuse
- **Access Control** - Authentication and authorization
- **Compliance** - Meeting regulatory requirements

---

## API Key Management

### Secure Storage

```go
// Never hardcode API keys
// ❌ WRONG
const OPENAI_KEY = "sk-abc123..."

// ✅ CORRECT - Environment variables
apiKey := os.Getenv("OPENAI_API_KEY")
if apiKey == "" {
    return errors.New("OPENAI_API_KEY not set")
}

// ✅ BETTER - Secret management service
type SecretManager interface {
    GetSecret(ctx context.Context, name string) (string, error)
    RotateSecret(ctx context.Context, name string) error
}

// AWS Secrets Manager implementation
type AWSSecretManager struct {
    client *secretsmanager.Client
}

func (sm *AWSSecretManager) GetSecret(ctx context.Context, name string) (string, error) {
    input := &secretsmanager.GetSecretValueInput{
        SecretId: aws.String(name),
    }
    
    result, err := sm.client.GetSecretValue(ctx, input)
    if err != nil {
        return "", fmt.Errorf("failed to retrieve secret: %w", err)
    }
    
    return *result.SecretString, nil
}

// HashiCorp Vault implementation
type VaultSecretManager struct {
    client *vault.Client
    path   string
}

func (vm *VaultSecretManager) GetSecret(ctx context.Context, name string) (string, error) {
    secret, err := vm.client.Logical().ReadWithContext(ctx, 
        fmt.Sprintf("%s/%s", vm.path, name))
    if err != nil {
        return "", err
    }
    
    if secret == nil || secret.Data == nil {
        return "", errors.New("secret not found")
    }
    
    // Handle both v1 and v2 KV secrets engine
    var secretData map[string]interface{}
    if data, ok := secret.Data["data"].(map[string]interface{}); ok {
        secretData = data // v2
    } else {
        secretData = secret.Data // v1
    }
    
    value, ok := secretData["value"].(string)
    if !ok {
        return "", errors.New("secret value not found")
    }
    
    return value, nil
}
```

### Key Rotation

```go
// Automated API key rotation
type KeyRotationService struct {
    secretManager SecretManager
    providers     map[string]Provider
    schedule      time.Duration
    alerter       Alerter
}

func (krs *KeyRotationService) StartRotation(ctx context.Context) {
    ticker := time.NewTicker(krs.schedule)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            if err := krs.rotateKeys(ctx); err != nil {
                krs.alerter.Alert("Key rotation failed", err)
            }
            
        case <-ctx.Done():
            return
        }
    }
}

func (krs *KeyRotationService) rotateKeys(ctx context.Context) error {
    for provider, p := range krs.providers {
        // Generate new key with provider
        newKey, err := p.GenerateNewAPIKey(ctx)
        if err != nil {
            return fmt.Errorf("failed to generate key for %s: %w", provider, err)
        }
        
        // Test new key
        if err := p.TestAPIKey(ctx, newKey); err != nil {
            return fmt.Errorf("new key test failed for %s: %w", provider, err)
        }
        
        // Store new key
        secretName := fmt.Sprintf("%s_api_key", provider)
        if err := krs.secretManager.RotateSecret(ctx, secretName); err != nil {
            return fmt.Errorf("failed to store key for %s: %w", provider, err)
        }
        
        // Update application
        if err := krs.updateApplication(provider, newKey); err != nil {
            // Rollback
            krs.rollbackKey(ctx, provider)
            return fmt.Errorf("application update failed for %s: %w", provider, err)
        }
        
        // Revoke old key after grace period
        go krs.scheduleKeyRevocation(ctx, provider, 24*time.Hour)
    }
    
    return nil
}
```

### Environment Isolation

```go
// Separate keys per environment
type EnvironmentConfig struct {
    Environment string
    Providers   map[string]ProviderConfig
}

type ProviderConfig struct {
    APIKeySecret string
    RateLimit    int
    Models       []string
}

func LoadEnvironmentConfig() (*EnvironmentConfig, error) {
    env := os.Getenv("ENVIRONMENT")
    if env == "" {
        env = "development"
    }
    
    config := &EnvironmentConfig{
        Environment: env,
        Providers:   make(map[string]ProviderConfig),
    }
    
    switch env {
    case "production":
        config.Providers["openai"] = ProviderConfig{
            APIKeySecret: "prod/openai/api_key",
            RateLimit:    100,
            Models:       []string{"gpt-4o", "gpt-4o-mini"},
        }
        
    case "staging":
        config.Providers["openai"] = ProviderConfig{
            APIKeySecret: "staging/openai/api_key",
            RateLimit:    50,
            Models:       []string{"gpt-4o-mini"},
        }
        
    case "development":
        config.Providers["openai"] = ProviderConfig{
            APIKeySecret: "dev/openai/api_key",
            RateLimit:    10,
            Models:       []string{"gpt-3.5-turbo"},
        }
    }
    
    return config, nil
}
```

---

## Data Protection

### Encryption at Rest

```go
// AES-256 encryption for sensitive data
type DataEncryptor struct {
    key []byte
}

func NewDataEncryptor(keySecret string) (*DataEncryptor, error) {
    // Derive key from secret
    key := pbkdf2.Key([]byte(keySecret), []byte("gollms-salt"), 10000, 32, sha256.New)
    
    return &DataEncryptor{key: key}, nil
}

func (de *DataEncryptor) Encrypt(plaintext []byte) ([]byte, error) {
    block, err := aes.NewCipher(de.key)
    if err != nil {
        return nil, err
    }
    
    // GCM mode for authenticated encryption
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    
    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, err
    }
    
    ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
    return ciphertext, nil
}

func (de *DataEncryptor) Decrypt(ciphertext []byte) ([]byte, error) {
    block, err := aes.NewCipher(de.key)
    if err != nil {
        return nil, err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    
    nonceSize := gcm.NonceSize()
    if len(ciphertext) < nonceSize {
        return nil, errors.New("ciphertext too short")
    }
    
    nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return nil, err
    }
    
    return plaintext, nil
}

// Database encryption layer
type EncryptedDB struct {
    db        *sql.DB
    encryptor *DataEncryptor
}

func (edb *EncryptedDB) StoreConversation(userID string, messages []Message) error {
    // Serialize messages
    data, err := json.Marshal(messages)
    if err != nil {
        return err
    }
    
    // Encrypt data
    encrypted, err := edb.encryptor.Encrypt(data)
    if err != nil {
        return err
    }
    
    // Store encrypted data
    _, err = edb.db.Exec(`
        INSERT INTO conversations (user_id, encrypted_data, created_at)
        VALUES ($1, $2, NOW())
    `, userID, encrypted)
    
    return err
}
```

### Data Privacy

```go
// PII detection and redaction
type PIIRedactor struct {
    patterns map[string]*regexp.Regexp
}

func NewPIIRedactor() *PIIRedactor {
    return &PIIRedactor{
        patterns: map[string]*regexp.Regexp{
            "ssn":    regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
            "email":  regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`),
            "phone":  regexp.MustCompile(`\b\d{3}[-.]?\d{3}[-.]?\d{4}\b`),
            "credit": regexp.MustCompile(`\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b`),
        },
    }
}

func (pr *PIIRedactor) Redact(text string) string {
    redacted := text
    
    for piiType, pattern := range pr.patterns {
        redacted = pattern.ReplaceAllStringFunc(redacted, func(match string) string {
            return fmt.Sprintf("[%s_REDACTED]", strings.ToUpper(piiType))
        })
    }
    
    return redacted
}

// Data retention policies
type DataRetentionService struct {
    db           *sql.DB
    policies     map[string]RetentionPolicy
    encryptor    *DataEncryptor
}

type RetentionPolicy struct {
    DataType  string
    Retention time.Duration
    Action    RetentionAction
}

type RetentionAction string

const (
    RetentionDelete     RetentionAction = "delete"
    RetentionAnonymize  RetentionAction = "anonymize"
    RetentionArchive    RetentionAction = "archive"
)

func (drs *DataRetentionService) EnforceRetention(ctx context.Context) error {
    for dataType, policy := range drs.policies {
        cutoff := time.Now().Add(-policy.Retention)
        
        switch policy.Action {
        case RetentionDelete:
            _, err := drs.db.ExecContext(ctx, `
                DELETE FROM conversations 
                WHERE data_type = $1 AND created_at < $2
            `, dataType, cutoff)
            if err != nil {
                return err
            }
            
        case RetentionAnonymize:
            _, err := drs.db.ExecContext(ctx, `
                UPDATE conversations 
                SET user_id = 'anonymous', 
                    encrypted_data = $3
                WHERE data_type = $1 AND created_at < $2
            `, dataType, cutoff, []byte("anonymized"))
            if err != nil {
                return err
            }
            
        case RetentionArchive:
            // Move to cold storage
            if err := drs.archiveOldData(ctx, dataType, cutoff); err != nil {
                return err
            }
        }
    }
    
    return nil
}
```

### Secure Communication

```go
// TLS configuration
func CreateSecureTLSConfig() *tls.Config {
    return &tls.Config{
        MinVersion:               tls.VersionTLS13,
        PreferServerCipherSuites: true,
        CipherSuites: []uint16{
            tls.TLS_AES_256_GCM_SHA384,
            tls.TLS_AES_128_GCM_SHA256,
            tls.TLS_CHACHA20_POLY1305_SHA256,
        },
        CurvePreferences: []tls.CurveID{
            tls.X25519,
            tls.CurveP256,
        },
    }
}

// mTLS for service-to-service communication
func CreateMTLSConfig(certFile, keyFile, caFile string) (*tls.Config, error) {
    // Load client certificate
    cert, err := tls.LoadX509KeyPair(certFile, keyFile)
    if err != nil {
        return nil, err
    }
    
    // Load CA certificate
    caCert, err := os.ReadFile(caFile)
    if err != nil {
        return nil, err
    }
    
    caCertPool := x509.NewCertPool()
    caCertPool.AppendCertsFromPEM(caCert)
    
    return &tls.Config{
        Certificates: []tls.Certificate{cert},
        RootCAs:      caCertPool,
        ClientAuth:   tls.RequireAndVerifyClientCert,
        ClientCAs:    caCertPool,
        MinVersion:   tls.VersionTLS13,
    }, nil
}
```

---

## Input Validation and Sanitization

### Request Validation

```go
// Comprehensive input validation
type RequestValidator struct {
    maxLength       int
    allowedPatterns []*regexp.Regexp
    blockedPatterns []*regexp.Regexp
    contentFilter   ContentFilter
}

func (rv *RequestValidator) Validate(req *CompletionRequest) error {
    // Validate message length
    totalLength := 0
    for _, msg := range req.Messages {
        totalLength += len(msg.Content)
        if totalLength > rv.maxLength {
            return fmt.Errorf("request too large: %d > %d", totalLength, rv.maxLength)
        }
    }
    
    // Check for blocked patterns (injection attempts)
    for _, msg := range req.Messages {
        for _, pattern := range rv.blockedPatterns {
            if pattern.MatchString(msg.Content) {
                return errors.New("request contains prohibited content")
            }
        }
    }
    
    // Validate against allowed patterns
    for _, msg := range req.Messages {
        valid := false
        for _, pattern := range rv.allowedPatterns {
            if pattern.MatchString(msg.Content) {
                valid = true
                break
            }
        }
        if !valid && len(rv.allowedPatterns) > 0 {
            return errors.New("request does not match allowed patterns")
        }
    }
    
    // Content filtering
    if err := rv.contentFilter.Check(req); err != nil {
        return fmt.Errorf("content filter failed: %w", err)
    }
    
    return nil
}

// Injection prevention
type InjectionPrevention struct {
    patterns map[string]*regexp.Regexp
}

func NewInjectionPrevention() *InjectionPrevention {
    return &InjectionPrevention{
        patterns: map[string]*regexp.Regexp{
            "prompt_injection": regexp.MustCompile(`(?i)(ignore|disregard|forget).*(previous|above|prior).*(instructions|prompt)`),
            "sql_injection":    regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop)\s+(from|into|table|database)`),
            "command_injection": regexp.MustCompile(`[;&|]\s*(rm|del|format|shutdown|reboot)`),
            "path_traversal":   regexp.MustCompile(`\.\.\/|\.\.\\`),
        },
    }
}

func (ip *InjectionPrevention) Detect(text string) (bool, string) {
    for injectionType, pattern := range ip.patterns {
        if pattern.MatchString(text) {
            return true, injectionType
        }
    }
    return false, ""
}
```

### Output Sanitization

```go
// Response sanitization
type ResponseSanitizer struct {
    filters []SanitizationFilter
}

type SanitizationFilter interface {
    Sanitize(text string) string
}

// HTML sanitization filter
type HTMLSanitizer struct{}

func (hs *HTMLSanitizer) Sanitize(text string) string {
    // Escape HTML special characters
    text = html.EscapeString(text)
    return text
}

// JSON sanitization filter
type JSONSanitizer struct{}

func (js *JSONSanitizer) Sanitize(text string) string {
    // Ensure valid JSON escaping
    var buf bytes.Buffer
    encoder := json.NewEncoder(&buf)
    encoder.SetEscapeHTML(true)
    
    if err := encoder.Encode(text); err != nil {
        return ""
    }
    
    // Remove quotes added by encoder
    result := buf.String()
    return result[1 : len(result)-2]
}

// URL sanitization filter
type URLSanitizer struct {
    allowedSchemes []string
}

func (us *URLSanitizer) Sanitize(text string) string {
    // Extract and validate URLs
    urlRegex := regexp.MustCompile(`https?://[^\s]+`)
    
    return urlRegex.ReplaceAllStringFunc(text, func(urlStr string) string {
        u, err := url.Parse(urlStr)
        if err != nil {
            return "[INVALID_URL]"
        }
        
        // Check allowed schemes
        allowed := false
        for _, scheme := range us.allowedSchemes {
            if u.Scheme == scheme {
                allowed = true
                break
            }
        }
        
        if !allowed {
            return "[BLOCKED_URL]"
        }
        
        return u.String()
    })
}
```

---

## Access Control

### Authentication

```go
// JWT-based authentication
type JWTAuthenticator struct {
    secretKey []byte
    issuer    string
    audience  string
}

func (ja *JWTAuthenticator) GenerateToken(userID string, roles []string) (string, error) {
    claims := jwt.MapClaims{
        "sub":   userID,
        "iss":   ja.issuer,
        "aud":   ja.audience,
        "exp":   time.Now().Add(24 * time.Hour).Unix(),
        "iat":   time.Now().Unix(),
        "roles": roles,
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(ja.secretKey)
}

func (ja *JWTAuthenticator) ValidateToken(tokenString string) (*User, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return ja.secretKey, nil
    })
    
    if err != nil {
        return nil, err
    }
    
    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        // Validate issuer and audience
        if claims["iss"] != ja.issuer || claims["aud"] != ja.audience {
            return nil, errors.New("invalid token claims")
        }
        
        user := &User{
            ID:    claims["sub"].(string),
            Roles: claims["roles"].([]string),
        }
        
        return user, nil
    }
    
    return nil, errors.New("invalid token")
}

// OAuth2 integration
type OAuth2Authenticator struct {
    provider     *oauth2.Config
    userInfoURL  string
}

func (oa *OAuth2Authenticator) HandleCallback(code string) (*User, error) {
    // Exchange code for token
    token, err := oa.provider.Exchange(context.Background(), code)
    if err != nil {
        return nil, err
    }
    
    // Get user info
    client := oa.provider.Client(context.Background(), token)
    resp, err := client.Get(oa.userInfoURL)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var userInfo map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
        return nil, err
    }
    
    return &User{
        ID:    userInfo["sub"].(string),
        Email: userInfo["email"].(string),
        Name:  userInfo["name"].(string),
    }, nil
}
```

### Authorization

```go
// Role-based access control (RBAC)
type RBACAuthorizer struct {
    permissions map[string][]string
}

func NewRBACAuthorizer() *RBACAuthorizer {
    return &RBACAuthorizer{
        permissions: map[string][]string{
            "admin": {"*"},
            "user":  {"read", "write:own"},
            "guest": {"read:public"},
        },
    }
}

func (ra *RBACAuthorizer) Authorize(user *User, resource string, action string) bool {
    for _, role := range user.Roles {
        if perms, ok := ra.permissions[role]; ok {
            for _, perm := range perms {
                if perm == "*" || perm == action {
                    return true
                }
                
                // Handle ownership-based permissions
                if strings.HasSuffix(perm, ":own") && 
                   strings.HasPrefix(perm, action) &&
                   ra.isOwner(user, resource) {
                    return true
                }
            }
        }
    }
    
    return false
}

// Attribute-based access control (ABAC)
type ABACAuthorizer struct {
    policies []Policy
}

type Policy struct {
    Subject  PolicyMatcher
    Resource PolicyMatcher
    Action   string
    Effect   PolicyEffect
}

type PolicyEffect string

const (
    PolicyAllow PolicyEffect = "allow"
    PolicyDeny  PolicyEffect = "deny"
)

func (aa *ABACAuthorizer) Authorize(ctx context.Context, attrs map[string]interface{}) bool {
    for _, policy := range aa.policies {
        if policy.Subject.Matches(attrs["subject"]) &&
           policy.Resource.Matches(attrs["resource"]) &&
           policy.Action == attrs["action"].(string) {
            
            if policy.Effect == PolicyDeny {
                return false
            }
            
            if policy.Effect == PolicyAllow {
                return true
            }
        }
    }
    
    return false // Default deny
}
```

### API Rate Limiting

```go
// User-based rate limiting
type UserRateLimiter struct {
    limiters map[string]*rate.Limiter
    mu       sync.RWMutex
    config   RateLimitConfig
}

type RateLimitConfig struct {
    DefaultRate  rate.Limit
    DefaultBurst int
    UserLimits   map[string]UserLimit
}

type UserLimit struct {
    Rate  rate.Limit
    Burst int
}

func (url *UserRateLimiter) Allow(userID string) bool {
    url.mu.RLock()
    limiter, exists := url.limiters[userID]
    url.mu.RUnlock()
    
    if !exists {
        url.mu.Lock()
        // Check if user has custom limits
        userLimit, hasCustom := url.config.UserLimits[userID]
        if hasCustom {
            limiter = rate.NewLimiter(userLimit.Rate, userLimit.Burst)
        } else {
            limiter = rate.NewLimiter(url.config.DefaultRate, url.config.DefaultBurst)
        }
        url.limiters[userID] = limiter
        url.mu.Unlock()
    }
    
    return limiter.Allow()
}

// Middleware for rate limiting
func RateLimitMiddleware(limiter *UserRateLimiter) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("user_id")
        if userID == "" {
            c.JSON(401, gin.H{"error": "unauthorized"})
            c.Abort()
            return
        }
        
        if !limiter.Allow(userID) {
            c.JSON(429, gin.H{"error": "rate limit exceeded"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

---

## Audit and Compliance

### Audit Logging

```go
// Comprehensive audit logging
type AuditLogger struct {
    db        *sql.DB
    encryptor *DataEncryptor
}

type AuditEvent struct {
    ID          string
    Timestamp   time.Time
    UserID      string
    Action      string
    Resource    string
    Result      string
    IPAddress   string
    UserAgent   string
    RequestData map[string]interface{}
    ResponseData map[string]interface{}
}

func (al *AuditLogger) Log(event *AuditEvent) error {
    // Sanitize sensitive data
    event.RequestData = al.sanitizeData(event.RequestData)
    event.ResponseData = al.sanitizeData(event.ResponseData)
    
    // Serialize event
    data, err := json.Marshal(event)
    if err != nil {
        return err
    }
    
    // Encrypt audit log
    encrypted, err := al.encryptor.Encrypt(data)
    if err != nil {
        return err
    }
    
    // Store with integrity check
    hash := sha256.Sum256(data)
    
    _, err = al.db.Exec(`
        INSERT INTO audit_logs (id, timestamp, user_id, action, encrypted_data, integrity_hash)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, event.ID, event.Timestamp, event.UserID, event.Action, encrypted, hash[:])
    
    return err
}

func (al *AuditLogger) sanitizeData(data map[string]interface{}) map[string]interface{} {
    sanitized := make(map[string]interface{})
    
    for k, v := range data {
        if al.isSensitiveField(k) {
            sanitized[k] = "[REDACTED]"
        } else {
            sanitized[k] = v
        }
    }
    
    return sanitized
}

// Audit trail verification
func (al *AuditLogger) VerifyAuditTrail(startTime, endTime time.Time) error {
    rows, err := al.db.Query(`
        SELECT encrypted_data, integrity_hash 
        FROM audit_logs 
        WHERE timestamp BETWEEN $1 AND $2
        ORDER BY timestamp
    `, startTime, endTime)
    if err != nil {
        return err
    }
    defer rows.Close()
    
    for rows.Next() {
        var encryptedData []byte
        var storedHash []byte
        
        if err := rows.Scan(&encryptedData, &storedHash); err != nil {
            return err
        }
        
        // Decrypt and verify
        data, err := al.encryptor.Decrypt(encryptedData)
        if err != nil {
            return fmt.Errorf("decryption failed: %w", err)
        }
        
        // Verify integrity
        computedHash := sha256.Sum256(data)
        if !bytes.Equal(computedHash[:], storedHash) {
            return errors.New("audit log integrity check failed")
        }
    }
    
    return nil
}
```

### Compliance Frameworks

```go
// GDPR compliance
type GDPRCompliance struct {
    db           *sql.DB
    encryptor    *DataEncryptor
    consentStore ConsentStore
}

func (gc *GDPRCompliance) HandleDataRequest(userID string, requestType DataRequestType) error {
    // Verify user identity
    if !gc.verifyUserIdentity(userID) {
        return errors.New("identity verification failed")
    }
    
    switch requestType {
    case DataRequestAccess:
        return gc.provideDataAccess(userID)
        
    case DataRequestPortability:
        return gc.exportUserData(userID)
        
    case DataRequestDeletion:
        return gc.deleteUserData(userID)
        
    case DataRequestRectification:
        return gc.allowDataCorrection(userID)
    }
    
    return errors.New("unknown request type")
}

func (gc *GDPRCompliance) deleteUserData(userID string) error {
    // Start transaction
    tx, err := gc.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // Delete from all tables
    tables := []string{"conversations", "user_preferences", "audit_logs"}
    for _, table := range tables {
        _, err := tx.Exec(fmt.Sprintf("DELETE FROM %s WHERE user_id = $1", table), userID)
        if err != nil {
            return err
        }
    }
    
    // Log deletion
    _, err = tx.Exec(`
        INSERT INTO deletion_log (user_id, deleted_at, confirmed)
        VALUES ($1, NOW(), true)
    `, userID)
    if err != nil {
        return err
    }
    
    return tx.Commit()
}

// SOC 2 compliance
type SOC2Compliance struct {
    controls []SecurityControl
}

type SecurityControl struct {
    ID          string
    Category    string
    Description string
    Test        func() error
}

func (sc *SOC2Compliance) RunComplianceChecks() []ComplianceResult {
    results := make([]ComplianceResult, len(sc.controls))
    
    for i, control := range sc.controls {
        result := ComplianceResult{
            ControlID: control.ID,
            Category:  control.Category,
            Timestamp: time.Now(),
        }
        
        if err := control.Test(); err != nil {
            result.Status = "FAILED"
            result.Details = err.Error()
        } else {
            result.Status = "PASSED"
        }
        
        results[i] = result
    }
    
    return results
}
```

---

## Security Monitoring

### Threat Detection

```go
// Anomaly detection system
type ThreatDetector struct {
    baseline    UsageBaseline
    alerter     Alerter
    rateLimiter *UserRateLimiter
}

type UsageBaseline struct {
    AvgRequestsPerHour map[string]float64
    AvgTokensPerRequest map[string]float64
    NormalHours        map[string][]int
}

func (td *ThreatDetector) DetectAnomalies(userID string, usage UserUsage) {
    // Check request rate anomaly
    if usage.RequestsPerHour > td.baseline.AvgRequestsPerHour[userID]*3 {
        td.alerter.Alert("High request rate", map[string]interface{}{
            "user_id": userID,
            "rate":    usage.RequestsPerHour,
            "baseline": td.baseline.AvgRequestsPerHour[userID],
        })
        
        // Apply temporary rate limit
        td.rateLimiter.ReduceLimit(userID, 0.5)
    }
    
    // Check token usage anomaly
    if usage.AvgTokensPerRequest > td.baseline.AvgTokensPerRequest[userID]*5 {
        td.alerter.Alert("Abnormal token usage", map[string]interface{}{
            "user_id": userID,
            "tokens":  usage.AvgTokensPerRequest,
            "baseline": td.baseline.AvgTokensPerRequest[userID],
        })
    }
    
    // Check access time anomaly
    currentHour := time.Now().Hour()
    normalHours := td.baseline.NormalHours[userID]
    isNormalHour := false
    
    for _, hour := range normalHours {
        if hour == currentHour {
            isNormalHour = true
            break
        }
    }
    
    if !isNormalHour {
        td.alerter.Alert("Unusual access time", map[string]interface{}{
            "user_id": userID,
            "hour":    currentHour,
            "normal_hours": normalHours,
        })
    }
}

// Security event correlation
type SecurityEventCorrelator struct {
    events      []SecurityEvent
    rules       []CorrelationRule
    timeWindow  time.Duration
}

func (sec *SecurityEventCorrelator) Correlate() []SecurityIncident {
    incidents := []SecurityIncident{}
    
    for _, rule := range sec.rules {
        matches := sec.findMatchingEvents(rule)
        if len(matches) >= rule.Threshold {
            incident := SecurityIncident{
                ID:          generateID(),
                Type:        rule.IncidentType,
                Severity:    rule.Severity,
                Events:      matches,
                DetectedAt:  time.Now(),
                Description: rule.Description,
            }
            
            incidents = append(incidents, incident)
        }
    }
    
    return incidents
}
```

### Incident Response

```go
// Automated incident response
type IncidentResponder struct {
    playbooks map[string]Playbook
    notifier  Notifier
    executor  CommandExecutor
}

type Playbook struct {
    Name     string
    Triggers []string
    Steps    []PlaybookStep
}

type PlaybookStep struct {
    Action      string
    Parameters  map[string]interface{}
    Conditions  []Condition
    OnSuccess   string
    OnFailure   string
}

func (ir *IncidentResponder) HandleIncident(incident SecurityIncident) error {
    // Find matching playbook
    playbook, found := ir.playbooks[incident.Type]
    if !found {
        // Manual intervention required
        return ir.notifier.NotifySecurityTeam(incident)
    }
    
    // Execute playbook
    context := map[string]interface{}{
        "incident": incident,
        "start_time": time.Now(),
    }
    
    for _, step := range playbook.Steps {
        // Check conditions
        if !ir.evaluateConditions(step.Conditions, context) {
            continue
        }
        
        // Execute action
        result, err := ir.executeAction(step.Action, step.Parameters, context)
        if err != nil {
            context["error"] = err
            if step.OnFailure != "" {
                ir.executeAction(step.OnFailure, nil, context)
            }
            return err
        }
        
        context["last_result"] = result
        
        if step.OnSuccess != "" {
            ir.executeAction(step.OnSuccess, nil, context)
        }
    }
    
    return nil
}

func (ir *IncidentResponder) executeAction(action string, params map[string]interface{}, context map[string]interface{}) (interface{}, error) {
    switch action {
    case "block_user":
        userID := params["user_id"].(string)
        return nil, ir.blockUser(userID)
        
    case "revoke_api_key":
        provider := params["provider"].(string)
        return nil, ir.revokeAPIKey(provider)
        
    case "increase_monitoring":
        level := params["level"].(string)
        return nil, ir.increaseMonitoring(level)
        
    case "notify_team":
        message := params["message"].(string)
        return nil, ir.notifier.NotifySecurityTeam(message)
        
    default:
        return nil, fmt.Errorf("unknown action: %s", action)
    }
}
```

---

## Security Best Practices Checklist

### Development Phase
- [ ] **Code Security**
  - [ ] Regular dependency scanning
  - [ ] Static code analysis (SAST)
  - [ ] Secret scanning in code
  - [ ] Secure coding training
  - [ ] Code review process

- [ ] **Testing**
  - [ ] Security unit tests
  - [ ] Integration security tests
  - [ ] Penetration testing
  - [ ] Vulnerability scanning
  - [ ] Security regression tests

### Deployment Phase
- [ ] **Infrastructure**
  - [ ] Network segmentation
  - [ ] Firewall rules configured
  - [ ] IDS/IPS deployed
  - [ ] DDoS protection enabled
  - [ ] SSL/TLS properly configured

- [ ] **Access Control**
  - [ ] Least privilege principle
  - [ ] MFA enabled
  - [ ] Service accounts secured
  - [ ] API keys rotated
  - [ ] Admin access logged

### Operations Phase
- [ ] **Monitoring**
  - [ ] Security event logging
  - [ ] Anomaly detection active
  - [ ] Threat intelligence feeds
  - [ ] Incident response plan
  - [ ] Regular security audits

- [ ] **Compliance**
  - [ ] Data classification done
  - [ ] Privacy policies updated
  - [ ] Consent management active
  - [ ] Audit trails maintained
  - [ ] Compliance reports generated

### Incident Response
- [ ] **Preparation**
  - [ ] Response team identified
  - [ ] Communication plan ready
  - [ ] Playbooks documented
  - [ ] Tools configured
  - [ ] Regular drills conducted

- [ ] **Response**
  - [ ] Detection mechanisms active
  - [ ] Containment procedures ready
  - [ ] Eradication steps defined
  - [ ] Recovery plans tested
  - [ ] Lessons learned process

---

## Security Configuration Template

```yaml
# security-config.yaml
security:
  api_keys:
    storage: vault
    rotation_schedule: 30d
    environments:
      - production
      - staging
      - development
  
  encryption:
    algorithm: AES-256-GCM
    key_derivation: PBKDF2
    iterations: 100000
    
  authentication:
    type: jwt
    expiration: 24h
    refresh_enabled: true
    mfa_required: true
    
  authorization:
    model: rbac
    default_role: user
    admin_roles:
      - admin
      - superadmin
      
  rate_limiting:
    default_rate: 100/hour
    burst: 20
    by_endpoint:
      /api/complete: 50/hour
      /api/embed: 200/hour
      
  input_validation:
    max_length: 10000
    blocked_patterns:
      - "(?i)ignore.*previous.*instructions"
      - "(?i)reveal.*api.*key"
    
  audit:
    enabled: true
    retention: 90d
    encryption: true
    integrity_check: true
    
  monitoring:
    anomaly_detection: true
    threat_correlation: true
    alert_channels:
      - email
      - slack
      - pagerduty
```

---

## Next Steps

- **[Custom Providers](custom-providers.md)** - Create secure custom providers
- **[Production Deployment](production-deployment.md)** - Secure deployment practices
- **[Troubleshooting Guide](troubleshooting.md)** - Security issue resolution
- **[Best Practices Checklist](/docs/user-guide/reference/best-practices-checklist.md)** - Security checklist
- **[Configuration Reference](/docs/user-guide/reference/configuration-reference.md)** - Security settings