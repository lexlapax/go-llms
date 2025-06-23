package docs

// ABOUTME: Common types for documentation generation including OpenAPI and component types
// ABOUTME: All types are bridge-friendly with JSON serialization support

import (
	"encoding/json"
)

// OpenAPISpec represents an OpenAPI 3.0 specification.
// It provides a complete description of an API including paths,
// components, security schemes, and metadata. All fields are
// JSON-serializable for easy integration with external systems.
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

// Info contains API metadata.
// It provides essential information about the API including
// title, version, description, and contact details.
type Info struct {
	Title          string   `json:"title"`
	Description    string   `json:"description,omitempty"`
	TermsOfService string   `json:"termsOfService,omitempty"`
	Contact        *Contact `json:"contact,omitempty"`
	License        *License `json:"license,omitempty"`
	Version        string   `json:"version"`
}

// Contact information.
// Provides contact details for the API support or development team.
type Contact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// License information.
// Specifies the license under which the API is provided.
type License struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// Server represents a server.
// It defines a server URL along with optional variables
// that can be substituted into the URL template.
type Server struct {
	URL         string                    `json:"url"`
	Description string                    `json:"description,omitempty"`
	Variables   map[string]ServerVariable `json:"variables,omitempty"`
}

// ServerVariable represents a server variable.
// Variables can be used in server URL templates and support
// enumerated values with defaults.
type ServerVariable struct {
	Enum        []string `json:"enum,omitempty"`
	Default     string   `json:"default"`
	Description string   `json:"description,omitempty"`
}

// PathItem represents operations on a path.
// It can contain multiple operations (GET, POST, etc.) along with
// common parameters and descriptions that apply to all operations.
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

// Operation represents an API operation.
// It describes a single API operation on a path, including
// parameters, request body, responses, and security requirements.
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

// Parameter represents an operation parameter.
// Parameters can be located in the path, query, header, or cookie.
// They define the expected inputs for an operation.
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

// RequestBody represents a request body.
// It describes a single request body with content in various
// media types (e.g., application/json, multipart/form-data).
type RequestBody struct {
	Description string               `json:"description,omitempty"`
	Content     map[string]MediaType `json:"content"`
	Required    bool                 `json:"required,omitempty"`
}

// Response represents an operation response.
// It describes a single response from an API operation,
// including headers and content in various media types.
type Response struct {
	Description string               `json:"description"`
	Headers     map[string]Header    `json:"headers,omitempty"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

// Header represents a response header.
// It defines a single HTTP header that can be returned
// in the response, including its schema and examples.
type Header struct {
	Description string      `json:"description,omitempty"`
	Schema      *Schema     `json:"schema,omitempty"`
	Example     interface{} `json:"example,omitempty"`
}

// MediaType represents content for a specific media type.
// It describes the schema, examples, and encoding for content
// in a specific format (e.g., application/json).
type MediaType struct {
	Schema   *Schema             `json:"schema,omitempty"`
	Example  interface{}         `json:"example,omitempty"`
	Examples map[string]Example  `json:"examples,omitempty"`
	Encoding map[string]Encoding `json:"encoding,omitempty"`
}

// Encoding represents encoding information.
// It defines how a specific property should be encoded
// in multipart or application/x-www-form-urlencoded requests.
type Encoding struct {
	ContentType string            `json:"contentType,omitempty"`
	Headers     map[string]Header `json:"headers,omitempty"`
	Style       string            `json:"style,omitempty"`
	Explode     bool              `json:"explode,omitempty"`
}

// Components holds reusable objects.
// It provides a container for various reusable definitions
// that can be referenced throughout the OpenAPI specification.
type Components struct {
	Schemas         map[string]*Schema        `json:"schemas,omitempty"`
	Responses       map[string]*Response      `json:"responses,omitempty"`
	Parameters      map[string]*Parameter     `json:"parameters,omitempty"`
	Examples        map[string]*Example       `json:"examples,omitempty"`
	RequestBodies   map[string]*RequestBody   `json:"requestBodies,omitempty"`
	Headers         map[string]*Header        `json:"headers,omitempty"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
}

// SecurityScheme represents a security scheme.
// It defines a security mechanism that can be used across
// the API, such as API keys, HTTP authentication, or OAuth2.
type SecurityScheme struct {
	Type        string      `json:"type"`
	Description string      `json:"description,omitempty"`
	Name        string      `json:"name,omitempty"`
	In          string      `json:"in,omitempty"`
	Scheme      string      `json:"scheme,omitempty"`
	Flows       *OAuthFlows `json:"flows,omitempty"`
}

// OAuthFlows represents OAuth flows.
// It contains configuration for the supported OAuth 2.0 flows
// (implicit, password, client credentials, authorization code).
type OAuthFlows struct {
	Implicit          *OAuthFlow `json:"implicit,omitempty"`
	Password          *OAuthFlow `json:"password,omitempty"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty"`
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty"`
}

// OAuthFlow represents an OAuth flow.
// It contains the configuration details for a specific
// OAuth 2.0 flow, including URLs and available scopes.
type OAuthFlow struct {
	AuthorizationURL string            `json:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty"`
	RefreshURL       string            `json:"refreshUrl,omitempty"`
	Scopes           map[string]string `json:"scopes"`
}

// SecurityRequirement represents a security requirement.
// It maps security scheme names to the scopes required for execution.
// An empty array means the security scheme is applied without scopes.
type SecurityRequirement map[string][]string

// Tag represents a tag for grouping.
// Tags are used to group operations in the OpenAPI specification
// and can include descriptions and links to external documentation.
type Tag struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty"`
}

// ExternalDocumentation represents external documentation.
// It provides a reference to external documentation that
// supplements the API description.
type ExternalDocumentation struct {
	Description string `json:"description,omitempty"`
	URL         string `json:"url"`
}

// DocumentationType represents the type of documentation to generate.
// It defines the output format for documentation generation,
// supporting OpenAPI, Markdown, and JSON formats.
type DocumentationType string

const (
	// TypeOpenAPI generates OpenAPI 3.0 specification
	TypeOpenAPI DocumentationType = "openapi"

	// TypeMarkdown generates Markdown documentation
	TypeMarkdown DocumentationType = "markdown"

	// TypeJSON generates JSON documentation
	TypeJSON DocumentationType = "json"
)

// GeneratorConfig contains configuration for documentation generation.
// It provides options to customize the output format, grouping,
// and content inclusion for generated documentation.
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

// MarshalJSON ensures all types are JSON serializable.
// This custom marshaler handles the OpenAPISpec serialization,
// ensuring all nested structures are properly converted to JSON.
//
// Returns the JSON representation or an error.
func (o *OpenAPISpec) MarshalJSON() ([]byte, error) {
	type Alias OpenAPISpec
	return json.Marshal((*Alias)(o))
}

// UnmarshalJSON ensures OpenAPISpec can be deserialized.
// This custom unmarshaler handles the OpenAPISpec deserialization,
// properly reconstructing all nested structures from JSON.
//
// Parameters:
//   - data: The JSON data to unmarshal
//
// Returns an error if unmarshaling fails.
func (o *OpenAPISpec) UnmarshalJSON(data []byte) error {
	type Alias OpenAPISpec
	return json.Unmarshal(data, (*Alias)(o))
}
