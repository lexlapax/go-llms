# Google Vertex AI Provider Design Document

## Overview

This document outlines the design for implementing Google Vertex AI provider support in go-llms. Unlike the standard Gemini provider which uses API keys, Vertex AI requires Google Cloud authentication and has different API endpoints and requirements.

## Key Differences from Standard Gemini Provider

### 1. Authentication
- **Vertex AI**: Uses Google Cloud IAM authentication (service accounts, ADC)
- **Standard Gemini**: Uses simple API keys
- **Environment**: `GOOGLE_APPLICATION_CREDENTIALS` vs `GEMINI_API_KEY`

### 2. Required Parameters
- **Project ID**: Required for all Vertex AI calls
- **Region/Location**: Must specify deployment region
- **Model Path**: Different model naming convention

### 3. API Structure
- **Base URL**: `https://{REGION}-aiplatform.googleapis.com/v1/`
- **Path Format**: `projects/{PROJECT}/locations/{LOCATION}/publishers/google/models/{MODEL}:{METHOD}`
- **Methods**: `generateContent`, `streamGenerateContent`

## Implementation Design

### Provider Structure

```go
type VertexAIProvider struct {
    projectID    string
    location     string
    model        string
    httpClient   *http.Client
    tokenSource  oauth2.TokenSource
    // Optimization: cache for converted messages
    messageCache *MessageCache
}
```

### Configuration Options

```go
// Required configuration
type VertexAIConfig struct {
    ProjectID string
    Location  string
    Model     string
}

// Optional configuration via provider options
- WithServiceAccountFile(path string)
- WithScopes(scopes []string)
- WithEndpoint(endpoint string) // For private/custom endpoints
```

### Authentication Flow

1. **Service Account JSON**: 
   - Check `GOOGLE_APPLICATION_CREDENTIALS` environment variable
   - Load credentials from JSON file
   - Create OAuth2 token source

2. **Application Default Credentials (ADC)**:
   - Use Google's ADC when no explicit credentials provided
   - Works automatically on GCP services

3. **Token Management**:
   - Automatic token refresh (tokens expire in 1 hour)
   - Use `golang.org/x/oauth2/google` package

### API Endpoints

#### Model Endpoints by Region
```
Base: https://{REGION}-aiplatform.googleapis.com/v1/

Regions:
- us-central1
- us-east1
- us-east4
- us-west1
- us-west4
- europe-west1
- europe-west2
- europe-west3
- europe-west4
- asia-east1
- asia-northeast1
- asia-southeast1
- australia-southeast1
```

#### Request Format
```json
{
  "contents": [
    {
      "role": "user",
      "parts": [
        {
          "text": "Hello"
        }
      ]
    }
  ],
  "generationConfig": {
    "temperature": 0.7,
    "maxOutputTokens": 1024,
    "topP": 0.95,
    "topK": 40
  }
}
```

### Model Support

#### Available Models
- `gemini-2.0-flash-preview-04-15`
- `gemini-1.5-pro-001`
- `gemini-1.5-flash-001`
- `claude-3-opus@20240229` (partner model)
- `claude-3-7-sonnet@20241022` (partner model)
- `claude-3-5-sonnet@20240620` (partner model)
- `claude-3-5-haiku@20241022` (partner model)

#### Model Naming Convention
- Google models: `{MODEL_NAME}`
- Partner models: `{MODEL_NAME}@{VERSION}`

### Error Handling

#### Common Errors
1. **Authentication Errors**:
   - Missing credentials
   - Invalid service account
   - Insufficient IAM permissions

2. **Configuration Errors**:
   - Invalid project ID
   - Unsupported region
   - Model not available in region

3. **Quota Errors**:
   - Rate limit exceeded
   - Token quota exceeded

### Testing Strategy

1. **Unit Tests**:
   - Mock OAuth2 token source
   - Mock HTTP responses
   - Test message conversion

2. **Integration Tests**:
   - Use test project with limited quota
   - Test with emulator if available
   - Environment variable: `VERTEX_AI_TEST_PROJECT`

### Environment Variables

```bash
# Required
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
VERTEX_AI_PROJECT_ID=my-gcp-project
VERTEX_AI_LOCATION=us-central1

# Optional
VERTEX_AI_MODEL=gemini-2.0-flash-preview-04-15
VERTEX_AI_ENDPOINT=https://custom-endpoint.googleapis.com
```

### Usage Example

```go
// Create provider
provider := provider.NewVertexAIProvider(
    "my-project-id",
    "us-central1",
    "gemini-2.0-flash-preview-04-15",
    provider.WithServiceAccountFile("/path/to/credentials.json"),
)

// Or with environment variables
provider := provider.NewVertexAIProvider(
    os.Getenv("VERTEX_AI_PROJECT_ID"),
    os.Getenv("VERTEX_AI_LOCATION"),
    os.Getenv("VERTEX_AI_MODEL"),
)

// Generate response
response, err := provider.Generate(ctx, "Hello, Vertex AI!")
```

### Integration with Model Discovery

```go
// Vertex AI model fetcher
type VertexAIFetcher struct {
    projectID   string
    location    string
    tokenSource oauth2.TokenSource
}

// List available models in a region
models, err := fetcher.FetchModels()
```

### Considerations

1. **Regional Deployment**: Users must choose appropriate region for latency/compliance
2. **Cost Management**: Different pricing than standard Gemini API
3. **Enterprise Features**: Access to Model Optimizer, request logging
4. **Partner Models**: Claude models available through Vertex AI
5. **Migration Path**: Help users migrate from standard Gemini to Vertex AI

## REST API Implementation Approach

Based on research, we will implement Vertex AI using REST APIs instead of the Google SDK to minimize dependencies and maintain consistency with other providers.

### Authentication Implementation

We can leverage our existing `pkg/util/auth` package and `golang.org/x/oauth2` for authentication:

```go
// Using service account with our auth utilities
func (p *VertexAIProvider) authenticate() error {
    // Option 1: Service Account JSON
    if p.serviceAccountPath != "" {
        keyData, err := os.ReadFile(p.serviceAccountPath)
        if err != nil {
            return err
        }
        
        config, err := google.JWTConfigFromJSON(keyData, 
            "https://www.googleapis.com/auth/cloud-platform")
        if err != nil {
            return err
        }
        
        p.tokenSource = config.TokenSource(context.Background())
        return nil
    }
    
    // Option 2: Application Default Credentials
    credentials, err := google.FindDefaultCredentials(context.Background(),
        "https://www.googleapis.com/auth/cloud-platform")
    if err != nil {
        return err
    }
    
    p.tokenSource = credentials.TokenSource
    return nil
}

// Apply auth to requests
func (p *VertexAIProvider) applyAuth(req *http.Request) error {
    token, err := p.tokenSource.Token()
    if err != nil {
        return fmt.Errorf("failed to get token: %w", err)
    }
    
    req.Header.Set("Authorization", "Bearer " + token.AccessToken)
    return nil
}
```

### REST API Endpoints

```go
// Generate content endpoint
func (p *VertexAIProvider) buildGenerateURL() string {
    return fmt.Sprintf(
        "https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:generateContent",
        p.location, p.projectID, p.location, p.model,
    )
}

// Stream generate content endpoint
func (p *VertexAIProvider) buildStreamURL() string {
    return fmt.Sprintf(
        "https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:streamGenerateContent",
        p.location, p.projectID, p.location, p.model,
    )
}
```

### Request/Response Handling

```go
// Convert domain messages to Vertex AI format
func (p *VertexAIProvider) convertToVertexFormat(messages []domain.Message) map[string]interface{} {
    contents := make([]map[string]interface{}, 0, len(messages))
    
    for _, msg := range messages {
        content := map[string]interface{}{
            "role": p.mapRole(msg.Role),
            "parts": p.convertParts(msg),
        }
        contents = append(contents, content)
    }
    
    return map[string]interface{}{
        "contents": contents,
        "generation_config": p.buildGenerationConfig(),
    }
}

// Map domain roles to Vertex AI roles
func (p *VertexAIProvider) mapRole(role domain.Role) string {
    switch role {
    case domain.RoleUser:
        return "USER"
    case domain.RoleAssistant:
        return "MODEL"
    case domain.RoleSystem:
        return "USER" // Vertex AI doesn't have system role
    default:
        return "USER"
    }
}
```

### SSE Streaming Implementation

```go
// Parse Server-Sent Events for streaming
func (p *VertexAIProvider) parseSSEStream(reader io.Reader, tokenCh chan<- domain.Token) {
    scanner := bufio.NewScanner(reader)
    var dataBuffer strings.Builder
    
    for scanner.Scan() {
        line := scanner.Text()
        
        if strings.HasPrefix(line, "data: ") {
            data := strings.TrimPrefix(line, "data: ")
            if data == "[DONE]" {
                return
            }
            dataBuffer.WriteString(data)
        } else if line == "" && dataBuffer.Len() > 0 {
            // Process accumulated data
            var response struct {
                Candidates []struct {
                    Content struct {
                        Parts []struct {
                            Text string `json:"text"`
                        } `json:"parts"`
                    } `json:"content"`
                } `json:"candidates"`
            }
            
            if err := json.Unmarshal([]byte(dataBuffer.String()), &response); err == nil {
                if len(response.Candidates) > 0 && len(response.Candidates[0].Content.Parts) > 0 {
                    text := response.Candidates[0].Content.Parts[0].Text
                    tokenCh <- domain.NewToken(text, false)
                }
            }
            
            dataBuffer.Reset()
        }
    }
}
```

### Integration with Existing Auth Package

```go
// Can use our existing OAuth2 utilities
func (p *VertexAIProvider) setupOAuth2() error {
    oauth2Config := &auth.OAuth2Config{
        TokenURL: "https://oauth2.googleapis.com/token",
        Scope:    "https://www.googleapis.com/auth/cloud-platform",
        Flow:     "service_account",
    }
    
    p.tokenManager = auth.NewOAuth2TokenManager(oauth2Config, p.httpClient)
    return nil
}
```

### Dependencies (Minimal)

```go
import (
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    // Our internal packages
    "github.com/lexlapax/go-llms/pkg/util/auth"
    "github.com/lexlapax/go-llms/pkg/llm/domain"
)
```

### Advantages of REST API Approach

1. **Reduced Dependencies**: No need for `cloud.google.com/go/vertexai/genai`
2. **Consistency**: Similar implementation pattern to other providers
3. **Control**: Direct control over HTTP requests and responses
4. **Flexibility**: Can add custom retry logic, caching, etc.
5. **Size**: Smaller binary size without Google SDK

### Implementation Considerations

1. **Token Management**: Use golang.org/x/oauth2 for automatic token refresh
2. **Error Handling**: Parse Vertex AI specific error responses
3. **Streaming**: Implement SSE parsing for streaming responses
4. **Multimodal**: Handle image/file uploads in request format
5. **Regional Endpoints**: Support all Vertex AI regions

### Next Steps

1. Implement core provider with REST API calls
2. Add OAuth2 authentication using existing utilities
3. Implement SSE streaming parser
4. Add comprehensive error handling
5. Create integration tests with real API
6. Add examples and documentation