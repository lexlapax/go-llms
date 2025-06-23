// ABOUTME: OpenAPI specification parsing and operation discovery for the API Client Tool
// ABOUTME: Supports OpenAPI 3.0/3.1 specs with automatic endpoint discovery and validation

package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/schema/validation"
)

// OpenAPISpec represents the root OpenAPI specification document
type OpenAPISpec struct {
	OpenAPI    string                `json:"openapi" yaml:"openapi"`
	Info       InfoObject            `json:"info" yaml:"info"`
	Servers    []ServerObject        `json:"servers,omitempty" yaml:"servers,omitempty"`
	Paths      map[string]PathItem   `json:"paths" yaml:"paths"`
	Components *ComponentsObject     `json:"components,omitempty" yaml:"components,omitempty"`
	Security   []SecurityRequirement `json:"security,omitempty" yaml:"security,omitempty"`
	Tags       []TagObject           `json:"tags,omitempty" yaml:"tags,omitempty"`
	Webhooks   map[string]PathItem   `json:"webhooks,omitempty" yaml:"webhooks,omitempty"` // OpenAPI 3.1+
}

// InfoObject provides metadata about the API
type InfoObject struct {
	Title          string         `json:"title" yaml:"title"`
	Description    string         `json:"description,omitempty" yaml:"description,omitempty"`
	TermsOfService string         `json:"termsOfService,omitempty" yaml:"termsOfService,omitempty"`
	Contact        *ContactObject `json:"contact,omitempty" yaml:"contact,omitempty"`
	License        *LicenseObject `json:"license,omitempty" yaml:"license,omitempty"`
	Version        string         `json:"version" yaml:"version"`
}

// ContactObject represents contact information
type ContactObject struct {
	Name  string `json:"name,omitempty" yaml:"name,omitempty"`
	URL   string `json:"url,omitempty" yaml:"url,omitempty"`
	Email string `json:"email,omitempty" yaml:"email,omitempty"`
}

// LicenseObject represents license information
type LicenseObject struct {
	Name string `json:"name" yaml:"name"`
	URL  string `json:"url,omitempty" yaml:"url,omitempty"`
}

// ServerObject represents a server
type ServerObject struct {
	URL         string                    `json:"url" yaml:"url"`
	Description string                    `json:"description,omitempty" yaml:"description,omitempty"`
	Variables   map[string]ServerVariable `json:"variables,omitempty" yaml:"variables,omitempty"`
}

// ServerVariable represents a server URL template variable
type ServerVariable struct {
	Enum        []string `json:"enum,omitempty" yaml:"enum,omitempty"`
	Default     string   `json:"default" yaml:"default"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
}

// PathItem describes operations available on a single path
type PathItem struct {
	Ref         string         `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Summary     string         `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string         `json:"description,omitempty" yaml:"description,omitempty"`
	Get         *Operation     `json:"get,omitempty" yaml:"get,omitempty"`
	Put         *Operation     `json:"put,omitempty" yaml:"put,omitempty"`
	Post        *Operation     `json:"post,omitempty" yaml:"post,omitempty"`
	Delete      *Operation     `json:"delete,omitempty" yaml:"delete,omitempty"`
	Options     *Operation     `json:"options,omitempty" yaml:"options,omitempty"`
	Head        *Operation     `json:"head,omitempty" yaml:"head,omitempty"`
	Patch       *Operation     `json:"patch,omitempty" yaml:"patch,omitempty"`
	Trace       *Operation     `json:"trace,omitempty" yaml:"trace,omitempty"`
	Servers     []ServerObject `json:"servers,omitempty" yaml:"servers,omitempty"`
	Parameters  []Parameter    `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

// Operation describes a single API operation
type Operation struct {
	Tags        []string              `json:"tags,omitempty" yaml:"tags,omitempty"`
	Summary     string                `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string                `json:"description,omitempty" yaml:"description,omitempty"`
	OperationID string                `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Parameters  []Parameter           `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses   map[string]Response   `json:"responses" yaml:"responses"`
	Callbacks   map[string]Callback   `json:"callbacks,omitempty" yaml:"callbacks,omitempty"`
	Deprecated  bool                  `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	Security    []SecurityRequirement `json:"security,omitempty" yaml:"security,omitempty"`
	Servers     []ServerObject        `json:"servers,omitempty" yaml:"servers,omitempty"`
}

// Parameter describes a single operation parameter
type Parameter struct {
	Name            string             `json:"name" yaml:"name"`
	In              string             `json:"in" yaml:"in"` // "query", "header", "path", "cookie"
	Description     string             `json:"description,omitempty" yaml:"description,omitempty"`
	Required        bool               `json:"required,omitempty" yaml:"required,omitempty"`
	Deprecated      bool               `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	AllowEmptyValue bool               `json:"allowEmptyValue,omitempty" yaml:"allowEmptyValue,omitempty"`
	Style           string             `json:"style,omitempty" yaml:"style,omitempty"`
	Explode         bool               `json:"explode,omitempty" yaml:"explode,omitempty"`
	AllowReserved   bool               `json:"allowReserved,omitempty" yaml:"allowReserved,omitempty"`
	Schema          *Schema            `json:"schema,omitempty" yaml:"schema,omitempty"`
	Example         interface{}        `json:"example,omitempty" yaml:"example,omitempty"`
	Examples        map[string]Example `json:"examples,omitempty" yaml:"examples,omitempty"`
}

// RequestBody describes a request body
type RequestBody struct {
	Description string               `json:"description,omitempty" yaml:"description,omitempty"`
	Content     map[string]MediaType `json:"content" yaml:"content"`
	Required    bool                 `json:"required,omitempty" yaml:"required,omitempty"`
}

// Response describes a single response from an API operation
type Response struct {
	Description string               `json:"description" yaml:"description"`
	Headers     map[string]Header    `json:"headers,omitempty" yaml:"headers,omitempty"`
	Content     map[string]MediaType `json:"content,omitempty" yaml:"content,omitempty"`
	Links       map[string]Link      `json:"links,omitempty" yaml:"links,omitempty"`
}

// MediaType provides schema and examples for a media type
type MediaType struct {
	Schema   *Schema             `json:"schema,omitempty" yaml:"schema,omitempty"`
	Example  interface{}         `json:"example,omitempty" yaml:"example,omitempty"`
	Examples map[string]Example  `json:"examples,omitempty" yaml:"examples,omitempty"`
	Encoding map[string]Encoding `json:"encoding,omitempty" yaml:"encoding,omitempty"`
}

// Schema represents a JSON Schema (OpenAPI 3.0 uses a subset of JSON Schema Draft 4)
type Schema struct {
	Type                 string             `json:"type,omitempty" yaml:"type,omitempty"`
	AllOf                []*Schema          `json:"allOf,omitempty" yaml:"allOf,omitempty"`
	OneOf                []*Schema          `json:"oneOf,omitempty" yaml:"oneOf,omitempty"`
	AnyOf                []*Schema          `json:"anyOf,omitempty" yaml:"anyOf,omitempty"`
	Not                  *Schema            `json:"not,omitempty" yaml:"not,omitempty"`
	Items                *Schema            `json:"items,omitempty" yaml:"items,omitempty"`
	Properties           map[string]*Schema `json:"properties,omitempty" yaml:"properties,omitempty"`
	AdditionalProperties interface{}        `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
	Description          string             `json:"description,omitempty" yaml:"description,omitempty"`
	Format               string             `json:"format,omitempty" yaml:"format,omitempty"`
	Default              interface{}        `json:"default,omitempty" yaml:"default,omitempty"`
	Title                string             `json:"title,omitempty" yaml:"title,omitempty"`
	MultipleOf           *float64           `json:"multipleOf,omitempty" yaml:"multipleOf,omitempty"`
	Maximum              *float64           `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	ExclusiveMaximum     *float64           `json:"exclusiveMaximum,omitempty" yaml:"exclusiveMaximum,omitempty"`
	Minimum              *float64           `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	ExclusiveMinimum     *float64           `json:"exclusiveMinimum,omitempty" yaml:"exclusiveMinimum,omitempty"`
	MaxLength            *int               `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	MinLength            *int               `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	Pattern              string             `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	MaxItems             *int               `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
	MinItems             *int               `json:"minItems,omitempty" yaml:"minItems,omitempty"`
	UniqueItems          bool               `json:"uniqueItems,omitempty" yaml:"uniqueItems,omitempty"`
	MaxProperties        *int               `json:"maxProperties,omitempty" yaml:"maxProperties,omitempty"`
	MinProperties        *int               `json:"minProperties,omitempty" yaml:"minProperties,omitempty"`
	Required             []string           `json:"required,omitempty" yaml:"required,omitempty"`
	Enum                 []interface{}      `json:"enum,omitempty" yaml:"enum,omitempty"`
	Example              interface{}        `json:"example,omitempty" yaml:"example,omitempty"`
	Nullable             bool               `json:"nullable,omitempty" yaml:"nullable,omitempty"`
	Discriminator        *Discriminator     `json:"discriminator,omitempty" yaml:"discriminator,omitempty"`
	ReadOnly             bool               `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
	WriteOnly            bool               `json:"writeOnly,omitempty" yaml:"writeOnly,omitempty"`
	XML                  *XML               `json:"xml,omitempty" yaml:"xml,omitempty"`
	ExternalDocs         *ExternalDocs      `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	Deprecated           bool               `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	Ref                  string             `json:"$ref,omitempty" yaml:"$ref,omitempty"`
}

// ComponentsObject holds a set of reusable objects for different aspects of the OAS
type ComponentsObject struct {
	Schemas         map[string]*Schema        `json:"schemas,omitempty" yaml:"schemas,omitempty"`
	Responses       map[string]Response       `json:"responses,omitempty" yaml:"responses,omitempty"`
	Parameters      map[string]Parameter      `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Examples        map[string]Example        `json:"examples,omitempty" yaml:"examples,omitempty"`
	RequestBodies   map[string]RequestBody    `json:"requestBodies,omitempty" yaml:"requestBodies,omitempty"`
	Headers         map[string]Header         `json:"headers,omitempty" yaml:"headers,omitempty"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty" yaml:"securitySchemes,omitempty"`
	Links           map[string]Link           `json:"links,omitempty" yaml:"links,omitempty"`
	Callbacks       map[string]Callback       `json:"callbacks,omitempty" yaml:"callbacks,omitempty"`
}

// SecurityScheme defines a security scheme
type SecurityScheme struct {
	Type             string      `json:"type" yaml:"type"`
	Description      string      `json:"description,omitempty" yaml:"description,omitempty"`
	Name             string      `json:"name,omitempty" yaml:"name,omitempty"`
	In               string      `json:"in,omitempty" yaml:"in,omitempty"`
	Scheme           string      `json:"scheme,omitempty" yaml:"scheme,omitempty"`
	BearerFormat     string      `json:"bearerFormat,omitempty" yaml:"bearerFormat,omitempty"`
	Flows            *OAuthFlows `json:"flows,omitempty" yaml:"flows,omitempty"`
	OpenIDConnectURL string      `json:"openIdConnectUrl,omitempty" yaml:"openIdConnectUrl,omitempty"`
}

// SecurityRequirement lists the required security schemes
type SecurityRequirement map[string][]string

// Supporting types (simplified for brevity)
// TagObject describes tags for API documentation and grouping.
type TagObject struct {
	Name         string        `json:"name" yaml:"name"`
	Description  string        `json:"description,omitempty" yaml:"description,omitempty"`
	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
}

// Example represents an example of a schema, parameter, or response.
type Example struct {
	Summary       string      `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description   string      `json:"description,omitempty" yaml:"description,omitempty"`
	Value         interface{} `json:"value,omitempty" yaml:"value,omitempty"`
	ExternalValue string      `json:"externalValue,omitempty" yaml:"externalValue,omitempty"`
}

// Header represents a header parameter in an HTTP response.
type Header struct {
	Description     string             `json:"description,omitempty" yaml:"description,omitempty"`
	Required        bool               `json:"required,omitempty" yaml:"required,omitempty"`
	Deprecated      bool               `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	AllowEmptyValue bool               `json:"allowEmptyValue,omitempty" yaml:"allowEmptyValue,omitempty"`
	Style           string             `json:"style,omitempty" yaml:"style,omitempty"`
	Explode         bool               `json:"explode,omitempty" yaml:"explode,omitempty"`
	AllowReserved   bool               `json:"allowReserved,omitempty" yaml:"allowReserved,omitempty"`
	Schema          *Schema            `json:"schema,omitempty" yaml:"schema,omitempty"`
	Example         interface{}        `json:"example,omitempty" yaml:"example,omitempty"`
	Examples        map[string]Example `json:"examples,omitempty" yaml:"examples,omitempty"`
}

// Link represents a design-time link for a response.
type Link struct {
	OperationRef string                 `json:"operationRef,omitempty" yaml:"operationRef,omitempty"`
	OperationID  string                 `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Parameters   map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody  interface{}            `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Description  string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Server       *ServerObject          `json:"server,omitempty" yaml:"server,omitempty"`
}

// Callback represents a map of possible out-of band callbacks related to the parent operation.
type Callback map[string]PathItem

// Encoding represents encoding information for a single schema property.
type Encoding struct {
	ContentType   string            `json:"contentType,omitempty" yaml:"contentType,omitempty"`
	Headers       map[string]Header `json:"headers,omitempty" yaml:"headers,omitempty"`
	Style         string            `json:"style,omitempty" yaml:"style,omitempty"`
	Explode       bool              `json:"explode,omitempty" yaml:"explode,omitempty"`
	AllowReserved bool              `json:"allowReserved,omitempty" yaml:"allowReserved,omitempty"`
}

// Discriminator represents a discriminator object for polymorphism support.
type Discriminator struct {
	PropertyName string            `json:"propertyName" yaml:"propertyName"`
	Mapping      map[string]string `json:"mapping,omitempty" yaml:"mapping,omitempty"`
}

// XML represents metadata about the XML representation of a schema.
type XML struct {
	Name      string `json:"name,omitempty" yaml:"name,omitempty"`
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Prefix    string `json:"prefix,omitempty" yaml:"prefix,omitempty"`
	Attribute bool   `json:"attribute,omitempty" yaml:"attribute,omitempty"`
	Wrapped   bool   `json:"wrapped,omitempty" yaml:"wrapped,omitempty"`
}

// ExternalDocs represents a reference to external documentation.
type ExternalDocs struct {
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	URL         string `json:"url" yaml:"url"`
}

// OAuthFlows represents OAuth flow configuration objects.
type OAuthFlows struct {
	Implicit          *OAuthFlow `json:"implicit,omitempty" yaml:"implicit,omitempty"`
	Password          *OAuthFlow `json:"password,omitempty" yaml:"password,omitempty"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty" yaml:"clientCredentials,omitempty"`
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty" yaml:"authorizationCode,omitempty"`
}

// OAuthFlow represents configuration details for a supported OAuth flow.
type OAuthFlow struct {
	AuthorizationURL string            `json:"authorizationUrl,omitempty" yaml:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty" yaml:"tokenUrl,omitempty"`
	RefreshURL       string            `json:"refreshUrl,omitempty" yaml:"refreshUrl,omitempty"`
	Scopes           map[string]string `json:"scopes" yaml:"scopes"`
}

// OpenAPIParser handles fetching and parsing OpenAPI specifications
type OpenAPIParser struct {
	client  *http.Client
	timeout time.Duration
}

// NewOpenAPIParser creates a new OpenAPI specification parser for API discovery and validation.
// It handles OpenAPI 3.0/3.1 specifications in both JSON and YAML formats, providing
// automatic endpoint discovery, operation enumeration, parameter validation, and
// LLM-friendly metadata extraction for seamless API exploration and integration.
func NewOpenAPIParser() *OpenAPIParser {
	return &OpenAPIParser{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		timeout: 30 * time.Second,
	}
}

// FetchSpec fetches an OpenAPI specification from a URL with automatic caching.
// It retrieves OpenAPI/Swagger specifications via HTTP, validates the content,
// parses JSON or YAML formats, and caches the results for improved performance
// while creating operation discovery instances for efficient API exploration.
func (p *OpenAPIParser) FetchSpec(specURL string) (*OpenAPISpec, error) {
	// Check cache first
	cache := GetOpenAPICache()
	if spec, _, found := cache.Get(specURL); found {
		return spec, nil
	}

	// Fetch from network
	resp, err := p.client.Get(specURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch OpenAPI spec from %s: %w", specURL, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch OpenAPI spec: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	spec, err := p.ParseSpec(body, specURL)
	if err != nil {
		return nil, err
	}

	// Store in cache
	discovery := NewOperationDiscovery(spec)
	cache.Set(specURL, spec, discovery)

	return spec, nil
}

// ParseSpec parses an OpenAPI specification from raw bytes in JSON or YAML format.
// It automatically detects the format, validates required fields and version compatibility,
// and returns a structured representation of the API specification suitable for
// operation discovery, validation, and LLM-friendly API exploration.
func (p *OpenAPIParser) ParseSpec(data []byte, source string) (*OpenAPISpec, error) {
	var spec OpenAPISpec

	// Determine if it's JSON or YAML by trying JSON first
	if err := json.Unmarshal(data, &spec); err != nil {
		// Try YAML if JSON parsing fails
		if yamlErr := yaml.Unmarshal(data, &spec); yamlErr != nil {
			return nil, fmt.Errorf("failed to parse spec as JSON or YAML: JSON error: %v, YAML error: %v", err, yamlErr)
		}
	}

	// Basic validation
	if err := p.validateSpec(&spec); err != nil {
		return nil, fmt.Errorf("invalid OpenAPI spec from %s: %w", source, err)
	}

	return &spec, nil
}

// validateSpec performs basic validation of the OpenAPI specification
func (p *OpenAPIParser) validateSpec(spec *OpenAPISpec) error {
	if spec.OpenAPI == "" {
		return fmt.Errorf("missing required field: openapi")
	}

	// Check if it's a supported version
	if !strings.HasPrefix(spec.OpenAPI, "3.0") && !strings.HasPrefix(spec.OpenAPI, "3.1") {
		return fmt.Errorf("unsupported OpenAPI version: %s (only 3.0.x and 3.1.x are supported)", spec.OpenAPI)
	}

	if spec.Info.Title == "" {
		return fmt.Errorf("missing required field: info.title")
	}

	if spec.Info.Version == "" {
		return fmt.Errorf("missing required field: info.version")
	}

	if spec.Paths == nil && spec.Components == nil && spec.Webhooks == nil {
		return fmt.Errorf("specification must contain at least one of: paths, components, or webhooks")
	}

	return nil
}

// GetOperations extracts all operations from the OpenAPI spec
func (spec *OpenAPISpec) GetOperations() []OperationInfo {
	var operations []OperationInfo

	for path, pathItem := range spec.Paths {
		// Handle operations at the path level
		pathOps := map[string]*Operation{
			"get":     pathItem.Get,
			"post":    pathItem.Post,
			"put":     pathItem.Put,
			"delete":  pathItem.Delete,
			"options": pathItem.Options,
			"head":    pathItem.Head,
			"patch":   pathItem.Patch,
			"trace":   pathItem.Trace,
		}

		for method, operation := range pathOps {
			if operation != nil {
				operations = append(operations, OperationInfo{
					Path:        path,
					Method:      strings.ToUpper(method),
					OperationID: operation.OperationID,
					Summary:     operation.Summary,
					Description: operation.Description,
					Tags:        operation.Tags,
					Parameters:  operation.Parameters,
					RequestBody: operation.RequestBody,
					Responses:   operation.Responses,
					Deprecated:  operation.Deprecated,
					Security:    operation.Security,
				})
			}
		}
	}

	return operations
}

// OperationInfo provides a simplified view of an operation for LLM usage
type OperationInfo struct {
	Path        string                `json:"path"`
	Method      string                `json:"method"`
	OperationID string                `json:"operationId,omitempty"`
	Summary     string                `json:"summary,omitempty"`
	Description string                `json:"description,omitempty"`
	Tags        []string              `json:"tags,omitempty"`
	Parameters  []Parameter           `json:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty"`
	Responses   map[string]Response   `json:"responses,omitempty"`
	Deprecated  bool                  `json:"deprecated,omitempty"`
	Security    []SecurityRequirement `json:"security,omitempty"`
}

// GetBaseURL returns the first server URL or empty string if none
func (spec *OpenAPISpec) GetBaseURL() string {
	if len(spec.Servers) > 0 {
		return spec.Servers[0].URL
	}
	return ""
}

// GetSecuritySchemes returns all security schemes defined in the spec
func (spec *OpenAPISpec) GetSecuritySchemes() map[string]SecurityScheme {
	if spec.Components != nil && spec.Components.SecuritySchemes != nil {
		return spec.Components.SecuritySchemes
	}
	return make(map[string]SecurityScheme)
}

// OperationDiscovery provides advanced operation discovery and metadata extraction
type OperationDiscovery struct {
	spec       *OpenAPISpec
	validator  *validation.Validator
	index      *OperationIndex
	operations []EnhancedOperationInfo // Cache enumerated operations
}

// NewOperationDiscovery creates a new operation discovery instance for advanced API exploration.
// It provides comprehensive operation enumeration, parameter extraction, schema validation,
// efficient operation lookup via indexing, and LLM-specific guidance generation
// to facilitate intelligent API interaction and automated request construction.
func NewOperationDiscovery(spec *OpenAPISpec) *OperationDiscovery {
	// Create validator without coercion to ensure proper constraint validation
	// Note: Coercion can interfere with constraint validation for numbers/integers
	// as it may transform values in ways that bypass min/max checks
	validator := validation.NewValidator()

	return &OperationDiscovery{
		spec:      spec,
		validator: validator,
	}
}

// EnumerateOperations returns all operations with comprehensive metadata for LLM consumption.
// It extracts and enhances operation information including parameters, request/response schemas,
// authentication requirements, and generates LLM-specific guidance, caching results
// for efficient repeated access while building operation indexes for fast lookups.
func (od *OperationDiscovery) EnumerateOperations() []EnhancedOperationInfo {
	// Return cached operations if already enumerated
	if od.operations != nil {
		return od.operations
	}

	var operations []EnhancedOperationInfo

	for path, pathItem := range od.spec.Paths {
		// Handle operations at the path level
		pathOps := map[string]*Operation{
			"get":     pathItem.Get,
			"post":    pathItem.Post,
			"put":     pathItem.Put,
			"delete":  pathItem.Delete,
			"options": pathItem.Options,
			"head":    pathItem.Head,
			"patch":   pathItem.Patch,
			"trace":   pathItem.Trace,
		}

		for method, operation := range pathOps {
			if operation != nil {
				enhanced := od.extractOperationMetadata(path, strings.ToUpper(method), operation, pathItem)
				operations = append(operations, enhanced)
			}
		}
	}

	// Cache operations and build index
	od.operations = operations
	od.index = NewOperationIndex(operations)

	return operations
}

// FindOperation efficiently finds an operation by method and path
func (od *OperationDiscovery) FindOperation(method, path string) (*EnhancedOperationInfo, bool) {
	// Ensure operations are enumerated and index is built
	if od.index == nil {
		od.EnumerateOperations()
	}

	return od.index.FindOperation(method, path)
}

// GetOperationsByTag returns operations grouped by tag
func (od *OperationDiscovery) GetOperationsByTag(tag string) []*EnhancedOperationInfo {
	// Ensure operations are enumerated and index is built
	if od.index == nil {
		od.EnumerateOperations()
	}

	return od.index.GetOperationsByTag(tag)
}

// EnhancedOperationInfo provides comprehensive operation metadata for LLM consumption
type EnhancedOperationInfo struct {
	// Basic Info
	Path        string   `json:"path"`
	Method      string   `json:"method"`
	OperationID string   `json:"operationId,omitempty"`
	Summary     string   `json:"summary,omitempty"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Deprecated  bool     `json:"deprecated,omitempty"`

	// Parameters
	PathParameters   []ParameterInfo `json:"pathParameters,omitempty"`
	QueryParameters  []ParameterInfo `json:"queryParameters,omitempty"`
	HeaderParameters []ParameterInfo `json:"headerParameters,omitempty"`
	CookieParameters []ParameterInfo `json:"cookieParameters,omitempty"`

	// Request Body
	RequestBodyInfo *RequestBodyInfo `json:"requestBody,omitempty"`

	// Responses
	ResponseInfo map[string]ResponseInfo `json:"responses,omitempty"`

	// Security
	SecurityRequirements  []SecurityRequirement `json:"security,omitempty"`
	AuthenticationMethods []string              `json:"authenticationMethods,omitempty"`

	// Server Info
	ServerURLs []string `json:"serverUrls,omitempty"`

	// LLM Guidance
	LLMGuidance OperationGuidance `json:"llmGuidance"`
}

// ParameterInfo provides detailed parameter information
type ParameterInfo struct {
	Name        string                 `json:"name"`
	In          string                 `json:"in"`
	Description string                 `json:"description,omitempty"`
	Required    bool                   `json:"required"`
	Deprecated  bool                   `json:"deprecated,omitempty"`
	Schema      SchemaInfo             `json:"schema,omitempty"`
	Example     interface{}            `json:"example,omitempty"`
	Examples    map[string]ExampleInfo `json:"examples,omitempty"`
}

// SchemaInfo provides simplified schema information for LLMs
type SchemaInfo struct {
	Type        string                `json:"type,omitempty"`
	Format      string                `json:"format,omitempty"`
	Description string                `json:"description,omitempty"`
	Enum        []interface{}         `json:"enum,omitempty"`
	Default     interface{}           `json:"default,omitempty"`
	Example     interface{}           `json:"example,omitempty"`
	Minimum     *float64              `json:"minimum,omitempty"`
	Maximum     *float64              `json:"maximum,omitempty"`
	MinLength   *int                  `json:"minLength,omitempty"`
	MaxLength   *int                  `json:"maxLength,omitempty"`
	Pattern     string                `json:"pattern,omitempty"`
	Properties  map[string]SchemaInfo `json:"properties,omitempty"`
	Items       *SchemaInfo           `json:"items,omitempty"`
	Required    []string              `json:"required,omitempty"`
}

// RequestBodyInfo provides request body metadata
type RequestBodyInfo struct {
	Description  string                 `json:"description,omitempty"`
	Required     bool                   `json:"required"`
	ContentTypes []string               `json:"contentTypes"`
	Schema       SchemaInfo             `json:"schema,omitempty"`
	Examples     map[string]ExampleInfo `json:"examples,omitempty"`
}

// ResponseInfo provides response metadata
type ResponseInfo struct {
	Description  string                   `json:"description"`
	ContentTypes []string                 `json:"contentTypes,omitempty"`
	Schema       SchemaInfo               `json:"schema,omitempty"`
	Headers      map[string]ParameterInfo `json:"headers,omitempty"`
	Examples     map[string]ExampleInfo   `json:"examples,omitempty"`
}

// ExampleInfo provides example metadata
type ExampleInfo struct {
	Summary     string      `json:"summary,omitempty"`
	Description string      `json:"description,omitempty"`
	Value       interface{} `json:"value,omitempty"`
}

// OperationGuidance provides LLM-specific guidance for using the operation
type OperationGuidance struct {
	UsageInstructions string             `json:"usageInstructions"`
	ParameterGuidance map[string]string  `json:"parameterGuidance,omitempty"`
	ErrorGuidance     map[string]string  `json:"errorGuidance,omitempty"`
	Examples          []OperationExample `json:"examples,omitempty"`
	Constraints       []string           `json:"constraints,omitempty"`
}

// OperationExample provides usage examples for LLMs
type OperationExample struct {
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	Parameters       map[string]interface{} `json:"parameters,omitempty"`
	RequestBody      interface{}            `json:"requestBody,omitempty"`
	ExpectedResponse string                 `json:"expectedResponse,omitempty"`
}

// extractOperationMetadata extracts comprehensive metadata from an operation
func (od *OperationDiscovery) extractOperationMetadata(path, method string, operation *Operation, pathItem PathItem) EnhancedOperationInfo {
	info := EnhancedOperationInfo{
		Path:                 path,
		Method:               method,
		OperationID:          operation.OperationID,
		Summary:              operation.Summary,
		Description:          operation.Description,
		Tags:                 operation.Tags,
		Deprecated:           operation.Deprecated,
		SecurityRequirements: operation.Security,
		ResponseInfo:         make(map[string]ResponseInfo),
	}

	// Extract parameters by type
	allParams := append(pathItem.Parameters, operation.Parameters...)
	for _, param := range allParams {
		paramInfo := od.convertParameter(param)
		switch param.In {
		case "path":
			info.PathParameters = append(info.PathParameters, paramInfo)
		case "query":
			info.QueryParameters = append(info.QueryParameters, paramInfo)
		case "header":
			info.HeaderParameters = append(info.HeaderParameters, paramInfo)
		case "cookie":
			info.CookieParameters = append(info.CookieParameters, paramInfo)
		}
	}

	// Extract request body information
	if operation.RequestBody != nil {
		info.RequestBodyInfo = od.convertRequestBody(*operation.RequestBody)
	}

	// Extract response information
	for statusCode, response := range operation.Responses {
		info.ResponseInfo[statusCode] = od.convertResponse(response)
	}

	// Extract server URLs
	info.ServerURLs = od.getOperationServerURLs(operation, pathItem)

	// Extract authentication methods
	info.AuthenticationMethods = od.getAuthenticationMethods(operation.Security)

	// Generate LLM guidance
	info.LLMGuidance = od.generateOperationGuidance(info, operation)

	return info
}

// convertParameter converts OpenAPI parameter to ParameterInfo
func (od *OperationDiscovery) convertParameter(param Parameter) ParameterInfo {
	paramInfo := ParameterInfo{
		Name:        param.Name,
		In:          param.In,
		Description: param.Description,
		Required:    param.Required,
		Deprecated:  param.Deprecated,
		Example:     param.Example,
	}

	if param.Schema != nil {
		paramInfo.Schema = od.convertSchema(*param.Schema)
	}

	// Convert examples
	paramInfo.Examples = make(map[string]ExampleInfo)
	for name, example := range param.Examples {
		paramInfo.Examples[name] = ExampleInfo{
			Summary:     example.Summary,
			Description: example.Description,
			Value:       example.Value,
		}
	}

	return paramInfo
}

// convertSchema converts OpenAPI schema to SchemaInfo
func (od *OperationDiscovery) convertSchema(schema Schema) SchemaInfo {
	schemaInfo := SchemaInfo{
		Type:        schema.Type,
		Format:      schema.Format,
		Description: schema.Description,
		Enum:        schema.Enum,
		Default:     schema.Default,
		Example:     schema.Example,
		Minimum:     schema.Minimum,
		Maximum:     schema.Maximum,
		MinLength:   schema.MinLength,
		MaxLength:   schema.MaxLength,
		Pattern:     schema.Pattern,
		Required:    schema.Required,
	}

	// Convert properties
	if schema.Properties != nil {
		schemaInfo.Properties = make(map[string]SchemaInfo)
		for name, propSchema := range schema.Properties {
			if propSchema != nil {
				schemaInfo.Properties[name] = od.convertSchema(*propSchema)
			}
		}
	}

	// Convert items for arrays
	if schema.Items != nil {
		itemSchema := od.convertSchema(*schema.Items)
		schemaInfo.Items = &itemSchema
	}

	return schemaInfo
}

// convertRequestBody converts OpenAPI request body to RequestBodyInfo
func (od *OperationDiscovery) convertRequestBody(reqBody RequestBody) *RequestBodyInfo {
	info := &RequestBodyInfo{
		Description: reqBody.Description,
		Required:    reqBody.Required,
		Examples:    make(map[string]ExampleInfo),
	}

	// Extract content types
	for contentType := range reqBody.Content {
		info.ContentTypes = append(info.ContentTypes, contentType)
	}

	// Get schema from first content type (usually application/json)
	for _, mediaType := range reqBody.Content {
		if mediaType.Schema != nil {
			info.Schema = od.convertSchema(*mediaType.Schema)
		}

		// Extract examples
		for name, example := range mediaType.Examples {
			info.Examples[name] = ExampleInfo{
				Summary:     example.Summary,
				Description: example.Description,
				Value:       example.Value,
			}
		}
		break // Use first content type
	}

	return info
}

// convertResponse converts OpenAPI response to ResponseInfo
func (od *OperationDiscovery) convertResponse(response Response) ResponseInfo {
	info := ResponseInfo{
		Description: response.Description,
		Headers:     make(map[string]ParameterInfo),
		Examples:    make(map[string]ExampleInfo),
	}

	// Extract content types
	for contentType := range response.Content {
		info.ContentTypes = append(info.ContentTypes, contentType)
	}

	// Get schema from first content type
	for _, mediaType := range response.Content {
		if mediaType.Schema != nil {
			info.Schema = od.convertSchema(*mediaType.Schema)
		}

		// Extract examples
		for name, example := range mediaType.Examples {
			info.Examples[name] = ExampleInfo{
				Summary:     example.Summary,
				Description: example.Description,
				Value:       example.Value,
			}
		}
		break
	}

	// Convert headers
	for name, header := range response.Headers {
		paramInfo := ParameterInfo{
			Name:        name,
			In:          "header",
			Description: header.Description,
			Required:    header.Required,
			Deprecated:  header.Deprecated,
			Example:     header.Example,
		}
		if header.Schema != nil {
			paramInfo.Schema = od.convertSchema(*header.Schema)
		}
		info.Headers[name] = paramInfo
	}

	return info
}

// getOperationServerURLs gets server URLs for the operation
func (od *OperationDiscovery) getOperationServerURLs(operation *Operation, pathItem PathItem) []string {
	var urls []string

	// Operation-level servers take precedence
	if len(operation.Servers) > 0 {
		for _, server := range operation.Servers {
			urls = append(urls, server.URL)
		}
		return urls
	}

	// Then path-level servers
	if len(pathItem.Servers) > 0 {
		for _, server := range pathItem.Servers {
			urls = append(urls, server.URL)
		}
		return urls
	}

	// Finally, spec-level servers
	if len(od.spec.Servers) > 0 {
		for _, server := range od.spec.Servers {
			urls = append(urls, server.URL)
		}
		return urls
	}

	return urls
}

// getAuthenticationMethods extracts authentication method names
func (od *OperationDiscovery) getAuthenticationMethods(security []SecurityRequirement) []string {
	var methods []string
	seen := make(map[string]bool)

	for _, req := range security {
		for schemeName := range req {
			if !seen[schemeName] {
				methods = append(methods, schemeName)
				seen[schemeName] = true
			}
		}
	}

	// If no operation-level security, check global security
	if len(methods) == 0 && len(od.spec.Security) > 0 {
		for _, req := range od.spec.Security {
			for schemeName := range req {
				if !seen[schemeName] {
					methods = append(methods, schemeName)
					seen[schemeName] = true
				}
			}
		}
	}

	return methods
}

// generateOperationGuidance creates LLM-specific guidance for the operation
func (od *OperationDiscovery) generateOperationGuidance(info EnhancedOperationInfo, operation *Operation) OperationGuidance {
	guidance := OperationGuidance{
		ParameterGuidance: make(map[string]string),
		ErrorGuidance:     make(map[string]string),
	}

	// Generate usage instructions
	guidance.UsageInstructions = od.generateUsageInstructions(info)

	// Generate parameter guidance
	allParams := append(append(append(info.PathParameters, info.QueryParameters...), info.HeaderParameters...), info.CookieParameters...)
	for _, param := range allParams {
		guidance.ParameterGuidance[param.Name] = od.generateParameterGuidance(param)
	}

	// Generate error guidance based on responses
	for statusCode, response := range info.ResponseInfo {
		if statusCode[0] == '4' || statusCode[0] == '5' {
			guidance.ErrorGuidance[statusCode] = od.generateErrorGuidance(statusCode, response)
		}
	}

	// Generate constraints
	guidance.Constraints = od.generateConstraints(info)

	// Generate examples
	guidance.Examples = od.generateOperationExamples(info)

	return guidance
}

// generateUsageInstructions creates usage instructions for LLMs
func (od *OperationDiscovery) generateUsageInstructions(info EnhancedOperationInfo) string {
	var instructions strings.Builder

	instructions.WriteString(fmt.Sprintf("Use this %s operation to %s", info.Method, strings.ToLower(info.Summary)))

	if info.Description != "" {
		instructions.WriteString(fmt.Sprintf(". %s", info.Description))
	}

	if len(info.PathParameters) > 0 {
		instructions.WriteString(". Path parameters are required and must be provided in the endpoint URL.")
	}

	if len(info.QueryParameters) > 0 {
		instructions.WriteString(". Query parameters can be used to filter or modify the request.")
	}

	if info.RequestBodyInfo != nil && info.RequestBodyInfo.Required {
		instructions.WriteString(". This operation requires a request body.")
	}

	if len(info.AuthenticationMethods) > 0 {
		instructions.WriteString(fmt.Sprintf(". Authentication required using: %s.", strings.Join(info.AuthenticationMethods, ", ")))
	}

	return instructions.String()
}

// generateParameterGuidance creates parameter-specific guidance
func (od *OperationDiscovery) generateParameterGuidance(param ParameterInfo) string {
	var guidance strings.Builder

	if param.Required {
		guidance.WriteString("Required parameter. ")
	} else {
		guidance.WriteString("Optional parameter. ")
	}

	if param.Schema.Type != "" {
		guidance.WriteString(fmt.Sprintf("Type: %s. ", param.Schema.Type))
	}

	if param.Schema.Format != "" {
		guidance.WriteString(fmt.Sprintf("Format: %s. ", param.Schema.Format))
	}

	if param.Schema.Minimum != nil || param.Schema.Maximum != nil {
		if param.Schema.Minimum != nil && param.Schema.Maximum != nil {
			guidance.WriteString(fmt.Sprintf("Valid range: %.0f to %.0f. ", *param.Schema.Minimum, *param.Schema.Maximum))
		} else if param.Schema.Minimum != nil {
			guidance.WriteString(fmt.Sprintf("Minimum value: %.0f. ", *param.Schema.Minimum))
		} else {
			guidance.WriteString(fmt.Sprintf("Maximum value: %.0f. ", *param.Schema.Maximum))
		}
	}

	if len(param.Schema.Enum) > 0 {
		guidance.WriteString(fmt.Sprintf("Valid values: %v. ", param.Schema.Enum))
	}

	if param.Schema.Pattern != "" {
		guidance.WriteString(fmt.Sprintf("Must match pattern: %s. ", param.Schema.Pattern))
	}

	return strings.TrimSpace(guidance.String())
}

// generateErrorGuidance creates error-specific guidance
func (od *OperationDiscovery) generateErrorGuidance(statusCode string, response ResponseInfo) string {
	switch statusCode[0] {
	case '4':
		switch statusCode {
		case "400":
			return "Bad Request - Check parameter values and request format. " + response.Description
		case "401":
			return "Unauthorized - Verify authentication credentials are provided and valid. " + response.Description
		case "403":
			return "Forbidden - Check if you have permission to access this resource. " + response.Description
		case "404":
			return "Not Found - Verify the resource path and parameters are correct. " + response.Description
		case "429":
			return "Rate Limited - Wait before making additional requests. " + response.Description
		default:
			return "Client Error - " + response.Description
		}
	case '5':
		return "Server Error - The API server encountered an issue. Retry may resolve temporary problems. " + response.Description
	default:
		return response.Description
	}
}

// generateConstraints creates operation constraints
func (od *OperationDiscovery) generateConstraints(info EnhancedOperationInfo) []string {
	var constraints []string

	if info.Deprecated {
		constraints = append(constraints, "This operation is deprecated and may be removed in future versions")
	}

	if len(info.AuthenticationMethods) > 0 {
		constraints = append(constraints, "Authentication is required")
	}

	requiredParams := 0
	for _, param := range info.PathParameters {
		if param.Required {
			requiredParams++
		}
	}
	for _, param := range info.QueryParameters {
		if param.Required {
			requiredParams++
		}
	}

	if requiredParams > 0 {
		constraints = append(constraints, fmt.Sprintf("%d required parameters must be provided", requiredParams))
	}

	if info.RequestBodyInfo != nil && info.RequestBodyInfo.Required {
		constraints = append(constraints, "Request body is required")
	}

	return constraints
}

// generateOperationExamples creates usage examples for the operation
func (od *OperationDiscovery) generateOperationExamples(info EnhancedOperationInfo) []OperationExample {
	var examples []OperationExample

	// Create a basic example
	basicExample := OperationExample{
		Name:        "Basic Usage",
		Description: fmt.Sprintf("Basic %s request to %s", info.Method, info.Path),
		Parameters:  make(map[string]interface{}),
	}

	// Add example parameter values
	for _, param := range info.PathParameters {
		if param.Example != nil {
			basicExample.Parameters[param.Name] = param.Example
		} else {
			basicExample.Parameters[param.Name] = od.generateExampleValue(param.Schema)
		}
	}

	for _, param := range info.QueryParameters {
		if param.Example != nil {
			basicExample.Parameters[param.Name] = param.Example
		} else if param.Required {
			basicExample.Parameters[param.Name] = od.generateExampleValue(param.Schema)
		}
	}

	// Add example request body
	if info.RequestBodyInfo != nil {
		if len(info.RequestBodyInfo.Examples) > 0 {
			for _, example := range info.RequestBodyInfo.Examples {
				basicExample.RequestBody = example.Value
				break
			}
		} else {
			basicExample.RequestBody = od.generateExampleValue(info.RequestBodyInfo.Schema)
		}
	}

	// Add expected response
	if response, exists := info.ResponseInfo["200"]; exists {
		basicExample.ExpectedResponse = fmt.Sprintf("HTTP 200: %s", response.Description)
	} else if response, exists := info.ResponseInfo["201"]; exists {
		basicExample.ExpectedResponse = fmt.Sprintf("HTTP 201: %s", response.Description)
	}

	examples = append(examples, basicExample)

	return examples
}

// generateExampleValue generates example values based on schema
func (od *OperationDiscovery) generateExampleValue(schema SchemaInfo) interface{} {
	if schema.Example != nil {
		return schema.Example
	}

	if schema.Default != nil {
		return schema.Default
	}

	if len(schema.Enum) > 0 {
		return schema.Enum[0]
	}

	switch schema.Type {
	case "string":
		switch schema.Format {
		case "email":
			return "user@example.com"
		case "date":
			return "2023-01-01"
		case "date-time":
			return "2023-01-01T12:00:00Z"
		}
		return "example"
	case "integer":
		if schema.Minimum != nil {
			return int(*schema.Minimum)
		}
		return 1
	case "number":
		if schema.Minimum != nil {
			return *schema.Minimum
		}
		return 1.0
	case "boolean":
		return true
	case "array":
		if schema.Items != nil {
			return []interface{}{od.generateExampleValue(*schema.Items)}
		}
		return []interface{}{}
	case "object":
		if schema.Properties != nil {
			obj := make(map[string]interface{})
			for name, propSchema := range schema.Properties {
				obj[name] = od.generateExampleValue(propSchema)
			}
			return obj
		}
		return map[string]interface{}{}
	default:
		return nil
	}
}

// FindOperationByID finds an operation by its operationId
func (od *OperationDiscovery) FindOperationByID(operationID string) *EnhancedOperationInfo {
	operations := od.EnumerateOperations()
	for _, op := range operations {
		if op.OperationID == operationID {
			return &op
		}
	}
	return nil
}

// FindOperationsByTag finds operations by tag
func (od *OperationDiscovery) FindOperationsByTag(tag string) []EnhancedOperationInfo {
	var results []EnhancedOperationInfo
	operations := od.EnumerateOperations()

	for _, op := range operations {
		for _, opTag := range op.Tags {
			if opTag == tag {
				results = append(results, op)
				break
			}
		}
	}

	return results
}

// FindOperationsByPath finds operations by path pattern
func (od *OperationDiscovery) FindOperationsByPath(pathPattern string) []EnhancedOperationInfo {
	var results []EnhancedOperationInfo
	operations := od.EnumerateOperations()

	for _, op := range operations {
		if strings.Contains(op.Path, pathPattern) {
			results = append(results, op)
		}
	}

	return results
}

// GetPathToOperationMap creates a mapping from path+method to operation info
func (od *OperationDiscovery) GetPathToOperationMap() map[string]EnhancedOperationInfo {
	operations := od.EnumerateOperations()
	mapping := make(map[string]EnhancedOperationInfo)

	for _, op := range operations {
		key := fmt.Sprintf("%s %s", op.Method, op.Path)
		mapping[key] = op
	}

	return mapping
}

// convertToValidationSchema converts OpenAPI Schema to schema domain Schema for validation
// // func (od *OperationDiscovery) convertToValidationSchema(schema Schema) *sdomain.Schema {
// 	result := &sdomain.Schema{
// 		Type:        schema.Type,
// 		Description: schema.Description,
// 		Title:       schema.Title,
// 	}
//
// 	// Convert properties
// 	if schema.Properties != nil {
// 		result.Properties = make(map[string]sdomain.Property)
// 		for name, prop := range schema.Properties {
// 			if prop != nil {
// 				result.Properties[name] = od.convertToValidationProperty(*prop)
// 			}
// 		}
// 	}
//
// 	// Convert required fields
// 	result.Required = schema.Required
//
// 	// Convert conditional schemas
// 	if schema.AllOf != nil {
// 		result.AllOf = make([]*sdomain.Schema, len(schema.AllOf))
// 		for i, subSchema := range schema.AllOf {
// 			if subSchema != nil {
// 				result.AllOf[i] = od.convertToValidationSchema(*subSchema)
// 			}
// 		}
// 	}
//
// 	if schema.AnyOf != nil {
// 		result.AnyOf = make([]*sdomain.Schema, len(schema.AnyOf))
// 		for i, subSchema := range schema.AnyOf {
// 			if subSchema != nil {
// 				result.AnyOf[i] = od.convertToValidationSchema(*subSchema)
// 			}
// 		}
// 	}
//
// 	if schema.OneOf != nil {
// 		result.OneOf = make([]*sdomain.Schema, len(schema.OneOf))
// 		for i, subSchema := range schema.OneOf {
// 			if subSchema != nil {
// 				result.OneOf[i] = od.convertToValidationSchema(*subSchema)
// 			}
// 		}
// 	}
//
// 	if schema.Not != nil {
// 		result.Not = od.convertToValidationSchema(*schema.Not)
// 	}
//
// 	return result
// }

// convertToValidationProperty converts OpenAPI Schema to schema domain Property for validation
// // func (od *OperationDiscovery) convertToValidationProperty(schema Schema) sdomain.Property {
// 	prop := sdomain.Property{
// 		Type:             schema.Type,
// 		Format:           schema.Format,
// 		Description:      schema.Description,
// 		Minimum:          schema.Minimum,
// 		Maximum:          schema.Maximum,
// 		ExclusiveMinimum: schema.ExclusiveMinimum,
// 		ExclusiveMaximum: schema.ExclusiveMaximum,
// 		MinLength:        schema.MinLength,
// 		MaxLength:        schema.MaxLength,
// 		MinItems:         schema.MinItems,
// 		MaxItems:         schema.MaxItems,
// 		Pattern:          schema.Pattern,
// 	}
//
// 	// Convert enum values to strings
// 	if schema.Enum != nil {
// 		prop.Enum = make([]string, len(schema.Enum))
// 		for i, enumVal := range schema.Enum {
// 			prop.Enum[i] = fmt.Sprintf("%v", enumVal)
// 		}
// 	}
//
// 	// Convert unique items
// 	if schema.UniqueItems {
// 		uniqueItems := true
// 		prop.UniqueItems = &uniqueItems
// 	}
//
// 	// Convert required fields
// 	prop.Required = schema.Required
//
// 	// Convert items schema for arrays
// 	if schema.Items != nil {
// 		itemProp := od.convertToValidationProperty(*schema.Items)
// 		prop.Items = &itemProp
// 	}
//
// 	// Convert nested properties for objects
// 	if schema.Properties != nil {
// 		prop.Properties = make(map[string]sdomain.Property)
// 		for name, nestedSchema := range schema.Properties {
// 			if nestedSchema != nil {
// 				prop.Properties[name] = od.convertToValidationProperty(*nestedSchema)
// 			}
// 		}
// 	}
//
// 	// Convert conditional schemas
// 	if schema.AnyOf != nil {
// 		prop.AnyOf = make([]*sdomain.Schema, len(schema.AnyOf))
// 		for i, subSchema := range schema.AnyOf {
// 			if subSchema != nil {
// 				prop.AnyOf[i] = od.convertToValidationSchema(*subSchema)
// 			}
// 		}
// 	}
//
// 	if schema.OneOf != nil {
// 		prop.OneOf = make([]*sdomain.Schema, len(schema.OneOf))
// 		for i, subSchema := range schema.OneOf {
// 			if subSchema != nil {
// 				prop.OneOf[i] = od.convertToValidationSchema(*subSchema)
// 			}
// 		}
// 	}
//
// 	if schema.Not != nil {
// 		prop.Not = od.convertToValidationSchema(*schema.Not)
// 	}
//
// 	return prop
// }

// ValidateRequestBody validates a request body against an operation's schema
func (od *OperationDiscovery) ValidateRequestBody(operationID string, requestBody interface{}) (*sdomain.ValidationResult, error) {
	// Find the operation
	op := od.FindOperationByID(operationID)
	if op == nil {
		return nil, fmt.Errorf("operation with ID '%s' not found", operationID)
	}

	// Check if operation has a request body
	if op.RequestBodyInfo == nil {
		return &sdomain.ValidationResult{Valid: true}, nil
	}

	// Convert the request body to JSON for validation
	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Convert OpenAPI schema to validation schema
	validationSchema := od.convertSchemaInfoToValidationSchema(op.RequestBodyInfo.Schema)

	// Validate using the schema validator
	return od.validator.Validate(validationSchema, string(requestBodyJSON))
}

// ValidateParameters validates operation parameters against their schemas
func (od *OperationDiscovery) ValidateParameters(operationID string, parameters map[string]interface{}) (map[string]*sdomain.ValidationResult, error) {
	// Find the operation
	op := od.FindOperationByID(operationID)
	if op == nil {
		return nil, fmt.Errorf("operation with ID '%s' not found", operationID)
	}

	results := make(map[string]*sdomain.ValidationResult)

	// Validate all parameter types
	allParams := append(append(append(op.PathParameters, op.QueryParameters...), op.HeaderParameters...), op.CookieParameters...)

	for _, param := range allParams {
		value, exists := parameters[param.Name]

		// Check required parameters
		if param.Required && !exists {
			results[param.Name] = &sdomain.ValidationResult{
				Valid:  false,
				Errors: []string{fmt.Sprintf("required parameter '%s' is missing", param.Name)},
			}
			continue
		}

		// Skip validation for optional missing parameters
		if !exists {
			results[param.Name] = &sdomain.ValidationResult{Valid: true}
			continue
		}

		// For parameter validation, we need to create an object with the parameter as a property
		// This is because the validator is designed to validate objects with properties
		parameterObject := map[string]interface{}{
			param.Name: value,
		}

		// Convert to JSON for validation
		objectJSON, err := json.Marshal(parameterObject)
		if err != nil {
			results[param.Name] = &sdomain.ValidationResult{
				Valid:  false,
				Errors: []string{fmt.Sprintf("failed to marshal parameter object: %v", err)},
			}
			continue
		}

		// Create validation schema with the parameter as a property
		paramProperty := od.convertSchemaInfoToValidationProperty(param.Schema)
		validationSchema := &sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				param.Name: paramProperty,
			},
		}

		// Add required constraint if parameter is required
		if param.Required {
			validationSchema.Required = []string{param.Name}
		}

		// Validate the parameter object
		result, err := od.validator.Validate(validationSchema, string(objectJSON))
		if err != nil {
			results[param.Name] = &sdomain.ValidationResult{
				Valid:  false,
				Errors: []string{fmt.Sprintf("validation error: %v", err)},
			}
			continue
		}

		results[param.Name] = result
	}

	return results, nil
}

// convertSchemaInfoToValidationSchema converts SchemaInfo to validation schema
func (od *OperationDiscovery) convertSchemaInfoToValidationSchema(schemaInfo SchemaInfo) *sdomain.Schema {
	result := &sdomain.Schema{
		Type:        schemaInfo.Type,
		Description: schemaInfo.Description,
		Required:    schemaInfo.Required,
	}

	// Convert properties if present
	if schemaInfo.Properties != nil {
		result.Properties = make(map[string]sdomain.Property)
		for name, prop := range schemaInfo.Properties {
			result.Properties[name] = od.convertSchemaInfoToValidationProperty(prop)
		}
	}

	return result
}

// convertSchemaInfoToValidationProperty converts SchemaInfo to validation property
func (od *OperationDiscovery) convertSchemaInfoToValidationProperty(schemaInfo SchemaInfo) sdomain.Property {
	prop := sdomain.Property{
		Type:        schemaInfo.Type,
		Format:      schemaInfo.Format,
		Description: schemaInfo.Description,
		Minimum:     schemaInfo.Minimum,
		Maximum:     schemaInfo.Maximum,
		MinLength:   schemaInfo.MinLength,
		MaxLength:   schemaInfo.MaxLength,
		Pattern:     schemaInfo.Pattern,
		Required:    schemaInfo.Required,
	}

	// Convert enum values
	if schemaInfo.Enum != nil {
		prop.Enum = make([]string, len(schemaInfo.Enum))
		for i, enumVal := range schemaInfo.Enum {
			prop.Enum[i] = fmt.Sprintf("%v", enumVal)
		}
	}

	// Convert items for arrays
	if schemaInfo.Items != nil {
		itemProp := od.convertSchemaInfoToValidationProperty(*schemaInfo.Items)
		prop.Items = &itemProp
	}

	// Convert nested properties for objects
	if schemaInfo.Properties != nil {
		prop.Properties = make(map[string]sdomain.Property)
		for name, nestedSchema := range schemaInfo.Properties {
			prop.Properties[name] = od.convertSchemaInfoToValidationProperty(nestedSchema)
		}
	}

	return prop
}

// OptimizeSchema applies schema optimization techniques to improve validation performance
func (od *OperationDiscovery) OptimizeSchema(operationID string) error {
	// Find the operation
	op := od.FindOperationByID(operationID)
	if op == nil {
		return fmt.Errorf("operation with ID '%s' not found", operationID)
	}

	// Apply schema optimizations using coercion and validation
	// This is a placeholder for more advanced optimizations

	// Optimize request body schema
	if op.RequestBodyInfo != nil {
		od.optimizeSchemaInfo(&op.RequestBodyInfo.Schema)
	}

	// Optimize parameter schemas
	allParams := append(append(append(op.PathParameters, op.QueryParameters...), op.HeaderParameters...), op.CookieParameters...)
	for i := range allParams {
		od.optimizeSchemaInfo(&allParams[i].Schema)
	}

	return nil
}

// optimizeSchemaInfo applies optimizations to a SchemaInfo structure
func (od *OperationDiscovery) optimizeSchemaInfo(schema *SchemaInfo) {
	// Apply type coercion optimizations
	if schema.Type == "" && schema.Format != "" {
		// Infer type from format
		switch schema.Format {
		case "email", "uri", "url", "uuid", "hostname", "ipv4", "ipv6":
			schema.Type = "string"
		case "date", "date-time":
			schema.Type = "string"
		case "int32", "int64":
			schema.Type = "integer"
		case "float", "double":
			schema.Type = "number"
		}
	}

	// Optimize enum constraints
	if len(schema.Enum) > 0 && schema.Type == "" {
		// Infer type from enum values
		if len(schema.Enum) > 0 {
			switch schema.Enum[0].(type) {
			case string:
				schema.Type = "string"
			case float64:
				schema.Type = "number"
			case bool:
				schema.Type = "boolean"
			}
		}
	}

	// Recursively optimize nested schemas
	if schema.Properties != nil {
		for name := range schema.Properties {
			prop := schema.Properties[name]
			od.optimizeSchemaInfo(&prop)
			schema.Properties[name] = prop
		}
	}

	if schema.Items != nil {
		od.optimizeSchemaInfo(schema.Items)
	}
}

// CoerceParameterValue applies type coercion to a parameter value based on its schema
func (od *OperationDiscovery) CoerceParameterValue(param ParameterInfo, value interface{}) (interface{}, error) {
	// Use the validator's coercion capabilities
	coercedValue, coerced := od.validator.Coerce(param.Schema.Type, value, param.Schema.Format)
	if !coerced {
		return value, fmt.Errorf("failed to coerce value %v to type %s", value, param.Schema.Type)
	}

	return coercedValue, nil
}

// ValidationOptions controls validation behavior for requests
type ValidationOptions struct {
	SkipRequired     bool `json:"skip_required,omitempty"`      // Skip required field validation
	SkipConstraints  bool `json:"skip_constraints,omitempty"`   // Skip min/max/pattern constraints
	SkipTypeChecking bool `json:"skip_type_checking,omitempty"` // Skip type validation
	AllowCoercion    bool `json:"allow_coercion,omitempty"`     // Allow type coercion
	StrictValidation bool `json:"strict_validation,omitempty"`  // Strict mode (all validations)
}

// ValidationReport provides comprehensive validation results with guidance
type ValidationReport struct {
	Valid            bool                        `json:"valid"`
	OperationID      string                      `json:"operation_id"`
	ParameterErrors  map[string]ValidationResult `json:"parameter_errors,omitempty"`
	RequestBodyError *ValidationResult           `json:"request_body_error,omitempty"`
	Guidance         ValidationGuidance          `json:"guidance"`
	Suggestions      []string                    `json:"suggestions,omitempty"`
}

// ValidationResult provides detailed validation result information
type ValidationResult struct {
	Valid        bool     `json:"valid"`
	Errors       []string `json:"errors,omitempty"`
	Warnings     []string `json:"warnings,omitempty"`
	FieldPath    string   `json:"field_path,omitempty"`
	ExpectedType string   `json:"expected_type,omitempty"`
	ActualValue  string   `json:"actual_value,omitempty"`
	Constraints  string   `json:"constraints,omitempty"`
}

// ValidationGuidance provides actionable guidance for validation errors
type ValidationGuidance struct {
	Summary           string            `json:"summary"`
	ParameterGuidance map[string]string `json:"parameter_guidance,omitempty"`
	BodyGuidance      string            `json:"body_guidance,omitempty"`
	Examples          []string          `json:"examples,omitempty"`
	DocumentationURL  string            `json:"documentation_url,omitempty"`
}

// ValidateRequest performs comprehensive request validation with detailed reporting for API operations.
// It validates parameters and request bodies against OpenAPI schemas, provides detailed error messages
// with field-level guidance, generates actionable suggestions for fixing validation errors,
// and supports flexible validation options for different use cases and error tolerance levels.
func (od *OperationDiscovery) ValidateRequest(operationID string, parameters map[string]interface{}, requestBody interface{}, options *ValidationOptions) (*ValidationReport, error) {
	// Find the operation
	op := od.FindOperationByID(operationID)
	if op == nil {
		return nil, fmt.Errorf("operation with ID '%s' not found", operationID)
	}

	// Apply defaults to options
	if options == nil {
		options = &ValidationOptions{}
	}

	report := &ValidationReport{
		Valid:           true,
		OperationID:     operationID,
		ParameterErrors: make(map[string]ValidationResult),
		Guidance: ValidationGuidance{
			ParameterGuidance: make(map[string]string),
		},
	}

	// Validate parameters
	if !options.SkipRequired || !options.SkipConstraints || !options.SkipTypeChecking {
		paramResults, err := od.ValidateParametersEnhanced(operationID, parameters, options)
		if err != nil {
			return nil, fmt.Errorf("parameter validation error: %w", err)
		}

		for paramName, result := range paramResults {
			if !result.Valid {
				report.Valid = false
				report.ParameterErrors[paramName] = ValidationResult{
					Valid:        false,
					Errors:       result.Errors,
					FieldPath:    paramName,
					ExpectedType: od.getParameterType(op, paramName),
					ActualValue:  fmt.Sprintf("%v", parameters[paramName]),
					Constraints:  od.getParameterConstraints(op, paramName),
				}
				report.Guidance.ParameterGuidance[paramName] = od.generateParameterErrorGuidance(op, paramName, result.Errors)
			}
		}
	}

	// Validate request body
	if requestBody != nil && !options.SkipTypeChecking {
		bodyResult, err := od.ValidateRequestBodyEnhanced(operationID, requestBody, options)
		if err != nil {
			return nil, fmt.Errorf("request body validation error: %w", err)
		}

		if !bodyResult.Valid {
			report.Valid = false
			report.RequestBodyError = &ValidationResult{
				Valid:       false,
				Errors:      bodyResult.Errors,
				FieldPath:   "requestBody",
				Constraints: "Must match operation schema",
			}
			report.Guidance.BodyGuidance = od.generateRequestBodyErrorGuidance(op, bodyResult.Errors)
		}
	}

	// Generate overall guidance
	report.Guidance.Summary = od.generateValidationSummary(report)
	report.Suggestions = od.generateValidationSuggestions(report, op)

	return report, nil
}

// ValidateParametersEnhanced provides enhanced parameter validation with detailed error reporting
func (od *OperationDiscovery) ValidateParametersEnhanced(operationID string, parameters map[string]interface{}, options *ValidationOptions) (map[string]*sdomain.ValidationResult, error) {
	// Find the operation
	op := od.FindOperationByID(operationID)
	if op == nil {
		return nil, fmt.Errorf("operation with ID '%s' not found", operationID)
	}

	results := make(map[string]*sdomain.ValidationResult)

	// Validate all parameter types
	allParams := append(append(append(op.PathParameters, op.QueryParameters...), op.HeaderParameters...), op.CookieParameters...)

	for _, param := range allParams {
		value, exists := parameters[param.Name]

		// Check required parameters (unless skipped)
		if !options.SkipRequired && param.Required && !exists {
			results[param.Name] = &sdomain.ValidationResult{
				Valid:  false,
				Errors: []string{fmt.Sprintf("required parameter '%s' is missing", param.Name)},
			}
			continue
		}

		// Skip validation for optional missing parameters
		if !exists {
			results[param.Name] = &sdomain.ValidationResult{Valid: true}
			continue
		}

		// Apply coercion if allowed
		if options.AllowCoercion {
			if coercedValue, err := od.CoerceParameterValue(param, value); err == nil {
				value = coercedValue
				parameters[param.Name] = value // Update the original map
			}
		}

		// Skip type checking if requested
		if options.SkipTypeChecking {
			results[param.Name] = &sdomain.ValidationResult{Valid: true}
			continue
		}

		// Create parameter object for validation
		parameterObject := map[string]interface{}{
			param.Name: value,
		}

		// Convert to JSON for validation
		objectJSON, err := json.Marshal(parameterObject)
		if err != nil {
			results[param.Name] = &sdomain.ValidationResult{
				Valid:  false,
				Errors: []string{fmt.Sprintf("failed to marshal parameter object: %v", err)},
			}
			continue
		}

		// Create validation schema with the parameter as a property
		paramProperty := od.convertSchemaInfoToValidationProperty(param.Schema)

		// Skip constraints if requested
		if options.SkipConstraints {
			paramProperty.Minimum = nil
			paramProperty.Maximum = nil
			paramProperty.MinLength = nil
			paramProperty.MaxLength = nil
			paramProperty.Pattern = ""
		}

		validationSchema := &sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				param.Name: paramProperty,
			},
		}

		// Add required constraint if parameter is required and not skipped
		if !options.SkipRequired && param.Required {
			validationSchema.Required = []string{param.Name}
		}

		// Validate the parameter object
		result, err := od.validator.Validate(validationSchema, string(objectJSON))
		if err != nil {
			results[param.Name] = &sdomain.ValidationResult{
				Valid:  false,
				Errors: []string{fmt.Sprintf("validation error: %v", err)},
			}
			continue
		}

		results[param.Name] = result
	}

	return results, nil
}

// ValidateRequestBodyEnhanced provides enhanced request body validation
func (od *OperationDiscovery) ValidateRequestBodyEnhanced(operationID string, requestBody interface{}, options *ValidationOptions) (*sdomain.ValidationResult, error) {
	// Find the operation
	op := od.FindOperationByID(operationID)
	if op == nil {
		return nil, fmt.Errorf("operation with ID '%s' not found", operationID)
	}

	// Check if operation has a request body
	if op.RequestBodyInfo == nil {
		// No request body expected, but one was provided
		if requestBody != nil {
			return &sdomain.ValidationResult{
				Valid:  false,
				Errors: []string{"operation does not expect a request body"},
			}, nil
		}
		return &sdomain.ValidationResult{Valid: true}, nil
	}

	// Check if required body is missing
	if !options.SkipRequired && op.RequestBodyInfo.Required && requestBody == nil {
		return &sdomain.ValidationResult{
			Valid:  false,
			Errors: []string{"required request body is missing"},
		}, nil
	}

	// Skip validation if requested
	if options.SkipTypeChecking || requestBody == nil {
		return &sdomain.ValidationResult{Valid: true}, nil
	}

	// Convert the request body to JSON for validation
	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Convert OpenAPI schema to validation schema
	validationSchema := od.convertSchemaInfoToValidationSchema(op.RequestBodyInfo.Schema)

	// Skip constraints if requested
	if options.SkipConstraints {
		od.removeConstraintsFromSchema(validationSchema)
	}

	// Validate using the schema validator
	return od.validator.Validate(validationSchema, string(requestBodyJSON))
}

// ValidateResponse validates operation response against schema (optional feature)
func (od *OperationDiscovery) ValidateResponse(operationID string, statusCode string, responseBody interface{}) (*sdomain.ValidationResult, error) {
	// Find the operation
	op := od.FindOperationByID(operationID)
	if op == nil {
		return nil, fmt.Errorf("operation with ID '%s' not found", operationID)
	}

	// Check if response schema exists
	responseInfo, exists := op.ResponseInfo[statusCode]
	if !exists {
		return &sdomain.ValidationResult{
			Valid:  false,
			Errors: []string{fmt.Sprintf("response status code '%s' is not defined for this operation", statusCode)},
		}, nil
	}

	// If no schema is defined for the response, consider it valid
	if responseInfo.Schema.Type == "" {
		return &sdomain.ValidationResult{Valid: true}, nil
	}

	// Convert response body to JSON for validation
	responseJSON, err := json.Marshal(responseBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response body: %w", err)
	}

	// Convert schema and validate
	validationSchema := od.convertSchemaInfoToValidationSchema(responseInfo.Schema)
	return od.validator.Validate(validationSchema, string(responseJSON))
}

// Helper methods for enhanced validation

// getParameterType returns the expected type for a parameter
func (od *OperationDiscovery) getParameterType(op *EnhancedOperationInfo, paramName string) string {
	allParams := append(append(append(op.PathParameters, op.QueryParameters...), op.HeaderParameters...), op.CookieParameters...)
	for _, param := range allParams {
		if param.Name == paramName {
			return param.Schema.Type
		}
	}
	return "unknown"
}

// getParameterConstraints returns constraint information for a parameter
func (od *OperationDiscovery) getParameterConstraints(op *EnhancedOperationInfo, paramName string) string {
	allParams := append(append(append(op.PathParameters, op.QueryParameters...), op.HeaderParameters...), op.CookieParameters...)
	for _, param := range allParams {
		if param.Name == paramName {
			var constraints []string
			if param.Schema.Minimum != nil {
				constraints = append(constraints, fmt.Sprintf("min: %.0f", *param.Schema.Minimum))
			}
			if param.Schema.Maximum != nil {
				constraints = append(constraints, fmt.Sprintf("max: %.0f", *param.Schema.Maximum))
			}
			if param.Schema.MinLength != nil {
				constraints = append(constraints, fmt.Sprintf("minLength: %d", *param.Schema.MinLength))
			}
			if param.Schema.MaxLength != nil {
				constraints = append(constraints, fmt.Sprintf("maxLength: %d", *param.Schema.MaxLength))
			}
			if param.Schema.Pattern != "" {
				constraints = append(constraints, fmt.Sprintf("pattern: %s", param.Schema.Pattern))
			}
			if len(param.Schema.Enum) > 0 {
				constraints = append(constraints, fmt.Sprintf("enum: %v", param.Schema.Enum))
			}
			if len(constraints) > 0 {
				return strings.Join(constraints, ", ")
			}
		}
	}
	return "none"
}

// generateParameterErrorGuidance creates guidance for parameter validation errors
func (od *OperationDiscovery) generateParameterErrorGuidance(op *EnhancedOperationInfo, paramName string, errors []string) string {
	allParams := append(append(append(op.PathParameters, op.QueryParameters...), op.HeaderParameters...), op.CookieParameters...)
	for _, param := range allParams {
		if param.Name == paramName {
			var guidance strings.Builder

			// Analyze errors and provide specific guidance
			for _, errMsg := range errors {
				if strings.Contains(strings.ToLower(errMsg), "required") {
					guidance.WriteString(fmt.Sprintf("Parameter '%s' is required for this operation. ", paramName))
				} else if strings.Contains(strings.ToLower(errMsg), "type") {
					guidance.WriteString(fmt.Sprintf("Parameter '%s' must be of type %s. ", paramName, param.Schema.Type))
				} else if strings.Contains(strings.ToLower(errMsg), "minimum") || strings.Contains(strings.ToLower(errMsg), "at least") {
					if param.Schema.Minimum != nil {
						guidance.WriteString(fmt.Sprintf("Parameter '%s' must be at least %.0f. ", paramName, *param.Schema.Minimum))
					}
				} else if strings.Contains(strings.ToLower(errMsg), "maximum") || strings.Contains(strings.ToLower(errMsg), "at most") {
					if param.Schema.Maximum != nil {
						guidance.WriteString(fmt.Sprintf("Parameter '%s' must be at most %.0f. ", paramName, *param.Schema.Maximum))
					}
				} else if strings.Contains(strings.ToLower(errMsg), "pattern") {
					if param.Schema.Pattern != "" {
						guidance.WriteString(fmt.Sprintf("Parameter '%s' must match pattern: %s. ", paramName, param.Schema.Pattern))
					}
				} else if strings.Contains(strings.ToLower(errMsg), "length") {
					if param.Schema.MinLength != nil && param.Schema.MaxLength != nil {
						guidance.WriteString(fmt.Sprintf("Parameter '%s' must be between %d and %d characters. ", paramName, *param.Schema.MinLength, *param.Schema.MaxLength))
					} else if param.Schema.MinLength != nil {
						guidance.WriteString(fmt.Sprintf("Parameter '%s' must be at least %d characters. ", paramName, *param.Schema.MinLength))
					} else if param.Schema.MaxLength != nil {
						guidance.WriteString(fmt.Sprintf("Parameter '%s' must be at most %d characters. ", paramName, *param.Schema.MaxLength))
					}
				}
			}

			// Add example if available
			if param.Example != nil {
				guidance.WriteString(fmt.Sprintf("Example: %v", param.Example))
			} else if param.Schema.Example != nil {
				guidance.WriteString(fmt.Sprintf("Example: %v", param.Schema.Example))
			}

			return guidance.String()
		}
	}
	return fmt.Sprintf("Parameter '%s' validation failed", paramName)
}

// generateRequestBodyErrorGuidance creates guidance for request body validation errors
func (od *OperationDiscovery) generateRequestBodyErrorGuidance(op *EnhancedOperationInfo, errors []string) string {
	var guidance strings.Builder

	if op.RequestBodyInfo == nil {
		return "This operation does not expect a request body"
	}

	guidance.WriteString("Request body validation failed. ")

	// Analyze errors and provide guidance
	for _, errMsg := range errors {
		if strings.Contains(strings.ToLower(errMsg), "required") {
			guidance.WriteString("Ensure all required fields are included. ")
		} else if strings.Contains(strings.ToLower(errMsg), "type") {
			guidance.WriteString("Check that field types match the expected schema. ")
		} else if strings.Contains(strings.ToLower(errMsg), "format") {
			guidance.WriteString("Verify that field formats (email, date, etc.) are correct. ")
		}
	}

	// Add content type guidance
	if len(op.RequestBodyInfo.ContentTypes) > 0 {
		guidance.WriteString(fmt.Sprintf("Supported content types: %s. ", strings.Join(op.RequestBodyInfo.ContentTypes, ", ")))
	}

	// Add schema type guidance
	if op.RequestBodyInfo.Schema.Type != "" {
		guidance.WriteString(fmt.Sprintf("Expected schema type: %s. ", op.RequestBodyInfo.Schema.Type))
	}

	return guidance.String()
}

// generateValidationSummary creates an overall validation summary
func (od *OperationDiscovery) generateValidationSummary(report *ValidationReport) string {
	if report.Valid {
		return "Request validation passed successfully"
	}

	errorCount := len(report.ParameterErrors)
	if report.RequestBodyError != nil {
		errorCount++
	}

	if errorCount == 1 {
		return "Request validation failed with 1 error"
	}
	return fmt.Sprintf("Request validation failed with %d errors", errorCount)
}

// generateValidationSuggestions creates actionable suggestions for fixing validation errors
func (od *OperationDiscovery) generateValidationSuggestions(report *ValidationReport, op *EnhancedOperationInfo) []string {
	var suggestions []string

	if report.Valid {
		return suggestions
	}

	// Parameter-specific suggestions
	for paramName, result := range report.ParameterErrors {
		if !result.Valid {
			for _, errMsg := range result.Errors {
				if strings.Contains(strings.ToLower(errMsg), "required") {
					suggestions = append(suggestions, fmt.Sprintf("Add the required parameter '%s' to your request", paramName))
				} else if strings.Contains(strings.ToLower(errMsg), "type") {
					suggestions = append(suggestions, fmt.Sprintf("Convert parameter '%s' to the correct type (%s)", paramName, result.ExpectedType))
				} else if strings.Contains(strings.ToLower(errMsg), "range") || strings.Contains(strings.ToLower(errMsg), "minimum") || strings.Contains(strings.ToLower(errMsg), "maximum") || strings.Contains(strings.ToLower(errMsg), "at most") || strings.Contains(strings.ToLower(errMsg), "at least") {
					suggestions = append(suggestions, fmt.Sprintf("Adjust parameter '%s' to meet constraints: %s", paramName, result.Constraints))
				} else {
					// Fallback suggestion for any validation error
					suggestions = append(suggestions, fmt.Sprintf("Fix validation error for parameter '%s': %s", paramName, errMsg))
				}
			}
		}
	}

	// Request body suggestions
	if report.RequestBodyError != nil && !report.RequestBodyError.Valid {
		suggestions = append(suggestions, "Review request body structure and ensure it matches the expected schema")
		if op.RequestBodyInfo != nil && len(op.RequestBodyInfo.Examples) > 0 {
			suggestions = append(suggestions, "Refer to the provided examples for correct request body format")
		}
	}

	// General suggestions
	if len(suggestions) > 2 {
		suggestions = append(suggestions, "Consider using request validation tools or API documentation for guidance")
	}

	return suggestions
}

// removeConstraintsFromSchema removes validation constraints from a schema
func (od *OperationDiscovery) removeConstraintsFromSchema(schema *sdomain.Schema) {
	if schema == nil {
		return
	}

	// Remove schema-level constraints
	schema.Required = nil

	// Remove constraints from properties
	for name, prop := range schema.Properties {
		prop.Minimum = nil
		prop.Maximum = nil
		prop.ExclusiveMinimum = nil
		prop.ExclusiveMaximum = nil
		prop.MinLength = nil
		prop.MaxLength = nil
		prop.MinItems = nil
		prop.MaxItems = nil
		prop.Pattern = ""
		prop.Enum = nil
		prop.UniqueItems = nil
		schema.Properties[name] = prop

		// Recursively remove constraints from nested properties
		for nestedName, nestedProp := range prop.Properties {
			nestedProp.Minimum = nil
			nestedProp.Maximum = nil
			nestedProp.ExclusiveMinimum = nil
			nestedProp.ExclusiveMaximum = nil
			nestedProp.MinLength = nil
			nestedProp.MaxLength = nil
			nestedProp.MinItems = nil
			nestedProp.MaxItems = nil
			nestedProp.Pattern = ""
			nestedProp.Enum = nil
			nestedProp.UniqueItems = nil
			prop.Properties[nestedName] = nestedProp
		}
	}
}
