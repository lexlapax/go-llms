package docs

// ABOUTME: Common types for documentation generation including OpenAPI and component types
// ABOUTME: All types are bridge-friendly with JSON serialization support

import (
	"encoding/json"
)

// OpenAPISpec represents an OpenAPI 3.0 specification
type OpenAPISpec struct {
	OpenAPI      string                 `json:"openapi"`
	Info         *Info                  `json:"info"`
	Servers      []Server               `json:"servers,omitempty"`
	Paths        map[string]*PathItem   `json:"paths,omitempty"`
	Components   *Components            `json:"components,omitempty"`
	Security     []SecurityRequirement  `json:"security,omitempty"`
	Tags         []Tag                  `json:"tags,omitempty"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty"`
}

// Info contains API metadata
type Info struct {
	Title          string   `json:"title"`
	Description    string   `json:"description,omitempty"`
	TermsOfService string   `json:"termsOfService,omitempty"`
	Contact        *Contact `json:"contact,omitempty"`
	License        *License `json:"license,omitempty"`
	Version        string   `json:"version"`
}

// Contact information
type Contact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// License information
type License struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// Server represents a server
type Server struct {
	URL         string                    `json:"url"`
	Description string                    `json:"description,omitempty"`
	Variables   map[string]ServerVariable `json:"variables,omitempty"`
}

// ServerVariable represents a server variable
type ServerVariable struct {
	Enum        []string `json:"enum,omitempty"`
	Default     string   `json:"default"`
	Description string   `json:"description,omitempty"`
}

// PathItem represents operations on a path
type PathItem struct {
	Summary     string     `json:"summary,omitempty"`
	Description string     `json:"description,omitempty"`
	Get         *Operation `json:"get,omitempty"`
	Put         *Operation `json:"put,omitempty"`
	Post        *Operation `json:"post,omitempty"`
	Delete      *Operation `json:"delete,omitempty"`
	Options     *Operation `json:"options,omitempty"`
	Head        *Operation `json:"head,omitempty"`
	Patch       *Operation `json:"patch,omitempty"`
	Trace       *Operation `json:"trace,omitempty"`
}

// Operation represents an API operation
type Operation struct {
	Tags         []string               `json:"tags,omitempty"`
	Summary      string                 `json:"summary,omitempty"`
	Description  string                 `json:"description,omitempty"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty"`
	OperationID  string                 `json:"operationId,omitempty"`
	Parameters   []Parameter            `json:"parameters,omitempty"`
	RequestBody  *RequestBody           `json:"requestBody,omitempty"`
	Responses    map[string]*Response   `json:"responses"`
	Deprecated   bool                   `json:"deprecated,omitempty"`
	Security     []SecurityRequirement  `json:"security,omitempty"`
}

// Parameter represents an operation parameter
type Parameter struct {
	Name            string      `json:"name"`
	In              string      `json:"in"`
	Description     string      `json:"description,omitempty"`
	Required        bool        `json:"required,omitempty"`
	Deprecated      bool        `json:"deprecated,omitempty"`
	AllowEmptyValue bool        `json:"allowEmptyValue,omitempty"`
	Schema          *Schema     `json:"schema,omitempty"`
	Example         interface{} `json:"example,omitempty"`
}

// RequestBody represents a request body
type RequestBody struct {
	Description string               `json:"description,omitempty"`
	Content     map[string]MediaType `json:"content"`
	Required    bool                 `json:"required,omitempty"`
}

// Response represents an operation response
type Response struct {
	Description string               `json:"description"`
	Headers     map[string]Header    `json:"headers,omitempty"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

// Header represents a response header
type Header struct {
	Description string      `json:"description,omitempty"`
	Schema      *Schema     `json:"schema,omitempty"`
	Example     interface{} `json:"example,omitempty"`
}

// MediaType represents content for a specific media type
type MediaType struct {
	Schema   *Schema             `json:"schema,omitempty"`
	Example  interface{}         `json:"example,omitempty"`
	Examples map[string]Example  `json:"examples,omitempty"`
	Encoding map[string]Encoding `json:"encoding,omitempty"`
}

// Encoding represents encoding information
type Encoding struct {
	ContentType string            `json:"contentType,omitempty"`
	Headers     map[string]Header `json:"headers,omitempty"`
	Style       string            `json:"style,omitempty"`
	Explode     bool              `json:"explode,omitempty"`
}

// Components holds reusable objects
type Components struct {
	Schemas         map[string]*Schema        `json:"schemas,omitempty"`
	Responses       map[string]*Response      `json:"responses,omitempty"`
	Parameters      map[string]*Parameter     `json:"parameters,omitempty"`
	Examples        map[string]*Example       `json:"examples,omitempty"`
	RequestBodies   map[string]*RequestBody   `json:"requestBodies,omitempty"`
	Headers         map[string]*Header        `json:"headers,omitempty"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
}

// SecurityScheme represents a security scheme
type SecurityScheme struct {
	Type        string      `json:"type"`
	Description string      `json:"description,omitempty"`
	Name        string      `json:"name,omitempty"`
	In          string      `json:"in,omitempty"`
	Scheme      string      `json:"scheme,omitempty"`
	Flows       *OAuthFlows `json:"flows,omitempty"`
}

// OAuthFlows represents OAuth flows
type OAuthFlows struct {
	Implicit          *OAuthFlow `json:"implicit,omitempty"`
	Password          *OAuthFlow `json:"password,omitempty"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty"`
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty"`
}

// OAuthFlow represents an OAuth flow
type OAuthFlow struct {
	AuthorizationURL string            `json:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty"`
	RefreshURL       string            `json:"refreshUrl,omitempty"`
	Scopes           map[string]string `json:"scopes"`
}

// SecurityRequirement represents a security requirement
type SecurityRequirement map[string][]string

// Tag represents a tag for grouping
type Tag struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty"`
}

// ExternalDocumentation represents external documentation
type ExternalDocumentation struct {
	Description string `json:"description,omitempty"`
	URL         string `json:"url"`
}

// DocumentationType represents the type of documentation to generate
type DocumentationType string

const (
	// TypeOpenAPI generates OpenAPI 3.0 specification
	TypeOpenAPI DocumentationType = "openapi"

	// TypeMarkdown generates Markdown documentation
	TypeMarkdown DocumentationType = "markdown"

	// TypeJSON generates JSON documentation
	TypeJSON DocumentationType = "json"
)

// GeneratorConfig contains configuration for documentation generation
type GeneratorConfig struct {
	// Title for the documentation
	Title string `json:"title"`

	// Description for the documentation
	Description string `json:"description"`

	// Version of the API/component
	Version string `json:"version"`

	// BaseURL for API endpoints (OpenAPI)
	BaseURL string `json:"baseUrl,omitempty"`

	// GroupBy specifies how to group items (e.g., "category", "type")
	GroupBy string `json:"groupBy,omitempty"`

	// IncludeExamples whether to include examples
	IncludeExamples bool `json:"includeExamples,omitempty"`

	// IncludeSchemas whether to include schemas
	IncludeSchemas bool `json:"includeSchemas,omitempty"`

	// CustomMetadata additional metadata
	CustomMetadata map[string]interface{} `json:"customMetadata,omitempty"`
}

// MarshalJSON ensures all types are JSON serializable
func (o *OpenAPISpec) MarshalJSON() ([]byte, error) {
	type Alias OpenAPISpec
	return json.Marshal((*Alias)(o))
}

// UnmarshalJSON ensures OpenAPISpec can be deserialized
func (o *OpenAPISpec) UnmarshalJSON(data []byte) error {
	type Alias OpenAPISpec
	return json.Unmarshal(data, (*Alias)(o))
}
