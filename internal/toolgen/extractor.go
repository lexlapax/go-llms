package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
)

// ToolMetadata represents extracted tool metadata
type ToolMetadata struct {
	Name                 string            `json:"name"`
	Description          string            `json:"description"`
	Category             string            `json:"category"`
	Tags                 []string          `json:"tags"`
	Version              string            `json:"version"`
	Package              string            `json:"package"`
	ConstructorFunc      string            `json:"constructor_func"`
	ParameterSchema      interface{}       `json:"parameter_schema,omitempty"`
	OutputSchema         interface{}       `json:"output_schema,omitempty"`
	UsageInstructions    string            `json:"usage_instructions,omitempty"`
	Examples             []Example         `json:"examples,omitempty"`
	Constraints          []string          `json:"constraints,omitempty"`
	ErrorGuidance        map[string]string `json:"error_guidance,omitempty"`
	RequiredPermissions  []string          `json:"required_permissions,omitempty"`
	ResourceUsage        ResourceInfo      `json:"resource_usage,omitempty"`
	IsDeterministic      bool              `json:"is_deterministic"`
	IsDestructive        bool              `json:"is_destructive"`
	RequiresConfirmation bool              `json:"requires_confirmation"`
	EstimatedLatency     string            `json:"estimated_latency,omitempty"`
}

// Example represents a tool example
type Example struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Input       interface{} `json:"input,omitempty"`
	Output      interface{} `json:"output,omitempty"`
}

// ResourceInfo represents resource requirements
type ResourceInfo struct {
	Memory      string `json:"memory,omitempty"`
	Network     bool   `json:"network,omitempty"`
	FileSystem  bool   `json:"file_system,omitempty"`
	Concurrency bool   `json:"concurrency,omitempty"`
}

// Extractor extracts tool metadata from Go source files
type Extractor struct {
	// Cache for resolved imports
	imports map[string]string
	// Schema variables found in the file
	schemaVars map[string]interface{}
}

// NewExtractor creates a new metadata extractor
func NewExtractor() *Extractor {
	return &Extractor{
		imports:    make(map[string]string),
		schemaVars: make(map[string]interface{}),
	}
}

// ExtractFromFile extracts tool metadata from a parsed Go file
func (e *Extractor) ExtractFromFile(file *ast.File) ([]ToolMetadata, error) {
	var tools []ToolMetadata

	// First, collect imports
	for _, imp := range file.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		if imp.Name != nil {
			e.imports[imp.Name.Name] = path
		} else {
			// Extract package name from path
			parts := strings.Split(path, "/")
			e.imports[parts[len(parts)-1]] = path
		}
	}

	// Collect schema variable definitions
	for _, decl := range file.Decls {
		if gen, ok := decl.(*ast.GenDecl); ok && gen.Tok == token.VAR {
			for _, spec := range gen.Specs {
				if vs, ok := spec.(*ast.ValueSpec); ok {
					for i, name := range vs.Names {
						if strings.Contains(name.Name, "Schema") || strings.Contains(name.Name, "schema") {
							if i < len(vs.Values) {
								e.schemaVars[name.Name] = e.extractSchema(vs.Values[i])
							}
						}
					}
				}
			}
		}
	}

	// Look for init functions
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "init" {
			// Look for MustRegisterTool calls
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				if call, ok := n.(*ast.CallExpr); ok {
					if e.isMustRegisterToolCall(call) {
						if metadata := e.extractRegistrationMetadata(call, file); metadata != nil {
							tools = append(tools, *metadata)
						}
					}
				}
				return true
			})
		}
	}

	// Builder metadata is now extracted during registration

	return tools, nil
}

// isMustRegisterToolCall checks if a call is to MustRegisterTool
func (e *Extractor) isMustRegisterToolCall(call *ast.CallExpr) bool {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		return sel.Sel.Name == "MustRegisterTool"
	}
	return false
}

// extractRegistrationMetadata extracts metadata from MustRegisterTool call
func (e *Extractor) extractRegistrationMetadata(call *ast.CallExpr, file *ast.File) *ToolMetadata {
	if len(call.Args) < 3 {
		return nil
	}

	metadata := &ToolMetadata{}

	// Extract name (first argument)
	if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
		metadata.Name = e.unquoteString(lit.Value)
	}

	// Extract constructor function name (second argument)
	if fnCall, ok := call.Args[1].(*ast.CallExpr); ok {
		if ident, ok := fnCall.Fun.(*ast.Ident); ok {
			// Store the constructor function name
			metadata.ConstructorFunc = ident.Name

			// Find the function definition and extract builder metadata
			if builderMeta := e.findAndExtractToolFunction(ident.Name, file); builderMeta != nil {
				e.mergeMetadata(metadata, builderMeta)
			}
		}
	}

	// Extract metadata (third argument)
	if comp, ok := call.Args[2].(*ast.CompositeLit); ok {
		e.extractToolMetadataFromComposite(comp, metadata)
	}

	return metadata
}

// extractToolMetadataFromComposite extracts fields from ToolMetadata composite literal
func (e *Extractor) extractToolMetadataFromComposite(comp *ast.CompositeLit, metadata *ToolMetadata) {
	for _, elt := range comp.Elts {
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			if ident, ok := kv.Key.(*ast.Ident); ok {
				switch ident.Name {
				case "Metadata":
					if nested, ok := kv.Value.(*ast.CompositeLit); ok {
						e.extractBuiltinsMetadata(nested, metadata)
					}
				case "RequiredPermissions":
					metadata.RequiredPermissions = e.extractStringSlice(kv.Value)
				case "ResourceUsage":
					if nested, ok := kv.Value.(*ast.CompositeLit); ok {
						e.extractResourceInfo(nested, &metadata.ResourceUsage)
					}
				}
			}
		}
	}
}

// extractBuiltinsMetadata extracts fields from builtins.Metadata
func (e *Extractor) extractBuiltinsMetadata(comp *ast.CompositeLit, metadata *ToolMetadata) {
	for _, elt := range comp.Elts {
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			if ident, ok := kv.Key.(*ast.Ident); ok {
				switch ident.Name {
				case "Name":
					metadata.Name = e.extractString(kv.Value)
				case "Description":
					metadata.Description = e.extractString(kv.Value)
				case "Category":
					metadata.Category = e.extractString(kv.Value)
				case "Tags":
					metadata.Tags = e.extractStringSlice(kv.Value)
				case "Version":
					metadata.Version = e.extractString(kv.Value)
				}
			}
		}
	}
}

// extractResourceInfo extracts ResourceInfo fields
func (e *Extractor) extractResourceInfo(comp *ast.CompositeLit, info *ResourceInfo) {
	for _, elt := range comp.Elts {
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			if ident, ok := kv.Key.(*ast.Ident); ok {
				switch ident.Name {
				case "Memory":
					info.Memory = e.extractString(kv.Value)
				case "Network":
					info.Network = e.extractBool(kv.Value)
				case "FileSystem":
					info.FileSystem = e.extractBool(kv.Value)
				case "Concurrency":
					info.Concurrency = e.extractBool(kv.Value)
				}
			}
		}
	}
}

// mergeMetadata merges builder metadata into registration metadata
func (e *Extractor) mergeMetadata(target, source *ToolMetadata) {
	if source.UsageInstructions != "" {
		target.UsageInstructions = source.UsageInstructions
	}
	// Always prefer examples from builder over metadata
	// Builder examples have Input/Output while metadata examples only have Code
	if len(source.Examples) > 0 {
		target.Examples = source.Examples
	}
	if len(source.Constraints) > 0 {
		target.Constraints = source.Constraints
	}
	if len(source.ErrorGuidance) > 0 {
		target.ErrorGuidance = source.ErrorGuidance
	}
	if source.ParameterSchema != nil {
		target.ParameterSchema = source.ParameterSchema
	}
	if source.OutputSchema != nil {
		target.OutputSchema = source.OutputSchema
	}
	target.IsDeterministic = source.IsDeterministic
	target.IsDestructive = source.IsDestructive
	target.RequiresConfirmation = source.RequiresConfirmation
	target.EstimatedLatency = source.EstimatedLatency
}

// Helper methods for extracting values

func (e *Extractor) extractString(expr ast.Expr) string {
	if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind == token.STRING {
		return e.unquoteString(lit.Value)
	}
	return ""
}

func (e *Extractor) extractBool(expr ast.Expr) bool {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name == "true"
	}
	return false
}

func (e *Extractor) extractStringSlice(expr ast.Expr) []string {
	if comp, ok := expr.(*ast.CompositeLit); ok {
		return e.extractStringArgs(comp.Elts)
	}
	return nil
}

func (e *Extractor) extractStringArgs(args []ast.Expr) []string {
	var result []string
	for _, arg := range args {
		if s := e.extractString(arg); s != "" {
			result = append(result, s)
		}
	}
	return result
}

func (e *Extractor) extractSchema(expr ast.Expr) interface{} {
	// Handle variable references
	if ident, ok := expr.(*ast.Ident); ok {
		if schema, found := e.schemaVars[ident.Name]; found {
			return schema
		}
		// Return placeholder for schema variables
		if strings.Contains(ident.Name, "Schema") || strings.Contains(ident.Name, "schema") {
			return map[string]interface{}{
				"type": "object",
			}
		}
	}

	// Handle &sdomain.Schema{...}
	if unary, ok := expr.(*ast.UnaryExpr); ok && unary.Op == token.AND {
		expr = unary.X
	}

	// Handle composite literal
	if comp, ok := expr.(*ast.CompositeLit); ok {
		schema := make(map[string]interface{})

		for _, elt := range comp.Elts {
			if kv, ok := elt.(*ast.KeyValueExpr); ok {
				if ident, ok := kv.Key.(*ast.Ident); ok {
					switch ident.Name {
					case "Type":
						schema["type"] = e.extractString(kv.Value)
					case "Properties":
						schema["properties"] = e.extractProperties(kv.Value)
					case "Required":
						schema["required"] = e.extractStringSlice(kv.Value)
					case "MinLength":
						if lit, ok := kv.Value.(*ast.BasicLit); ok && lit.Kind == token.INT {
							val, _ := strconv.ParseFloat(lit.Value, 64)
							schema["minLength"] = val
						}
					case "MaxLength":
						if lit, ok := kv.Value.(*ast.BasicLit); ok && lit.Kind == token.INT {
							val, _ := strconv.ParseFloat(lit.Value, 64)
							schema["maxLength"] = val
						}
					case "Minimum":
						if lit, ok := kv.Value.(*ast.BasicLit); ok && lit.Kind == token.INT {
							val, _ := strconv.ParseFloat(lit.Value, 64)
							schema["minimum"] = val
						}
					case "Maximum":
						if lit, ok := kv.Value.(*ast.BasicLit); ok && lit.Kind == token.INT {
							val, _ := strconv.ParseFloat(lit.Value, 64)
							schema["maximum"] = val
						}
					case "Default":
						schema["default"] = e.extractValue(kv.Value)
					case "Enum":
						schema["enum"] = e.extractEnumValues(kv.Value)
					case "Format":
						schema["format"] = e.extractString(kv.Value)
					case "Description":
						schema["description"] = e.extractString(kv.Value)
					}
				}
			}
		}

		return schema
	}

	// Return placeholder for other cases
	return map[string]interface{}{
		"type": "object",
	}
}

func (e *Extractor) extractProperties(expr ast.Expr) map[string]interface{} {
	props := make(map[string]interface{})

	if comp, ok := expr.(*ast.CompositeLit); ok {
		for _, elt := range comp.Elts {
			if kv, ok := elt.(*ast.KeyValueExpr); ok {
				propName := e.extractString(kv.Key)
				if propName != "" {
					props[propName] = e.extractPropertySchema(kv.Value)
				}
			}
		}
	}

	return props
}

func (e *Extractor) extractPropertySchema(expr ast.Expr) map[string]interface{} {
	prop := make(map[string]interface{})

	if comp, ok := expr.(*ast.CompositeLit); ok {
		for _, elt := range comp.Elts {
			if kv, ok := elt.(*ast.KeyValueExpr); ok {
				if ident, ok := kv.Key.(*ast.Ident); ok {
					switch ident.Name {
					case "Type":
						prop["type"] = e.extractString(kv.Value)
					case "Description":
						prop["description"] = e.extractString(kv.Value)
					case "MinLength":
						if lit, ok := kv.Value.(*ast.BasicLit); ok && lit.Kind == token.INT {
							val, _ := strconv.ParseFloat(lit.Value, 64)
							prop["minLength"] = val
						}
					case "MaxLength":
						if lit, ok := kv.Value.(*ast.BasicLit); ok && lit.Kind == token.INT {
							val, _ := strconv.ParseFloat(lit.Value, 64)
							prop["maxLength"] = val
						}
					case "Minimum":
						if lit, ok := kv.Value.(*ast.BasicLit); ok && lit.Kind == token.INT {
							val, _ := strconv.ParseFloat(lit.Value, 64)
							prop["minimum"] = val
						}
					case "Maximum":
						if lit, ok := kv.Value.(*ast.BasicLit); ok && lit.Kind == token.INT {
							val, _ := strconv.ParseFloat(lit.Value, 64)
							prop["maximum"] = val
						}
					case "Default":
						prop["default"] = e.extractValue(kv.Value)
					case "Enum":
						prop["enum"] = e.extractEnumValues(kv.Value)
					case "Format":
						prop["format"] = e.extractString(kv.Value)
					}
				}
			}
		}
	}

	return prop
}

func (e *Extractor) extractValue(expr ast.Expr) interface{} {
	switch v := expr.(type) {
	case *ast.BasicLit:
		switch v.Kind {
		case token.STRING:
			return e.unquoteString(v.Value)
		case token.INT:
			if strings.Contains(v.Value, ".") {
				f, _ := strconv.ParseFloat(v.Value, 64)
				return f
			}
			i, _ := strconv.ParseFloat(v.Value, 64)
			return i
		case token.FLOAT:
			f, _ := strconv.ParseFloat(v.Value, 64)
			return f
		}
	case *ast.Ident:
		// Handle boolean values
		switch v.Name {
		case "true":
			return true
		case "false":
			return false
		default:
			return v.Name
		}
	}
	return nil
}

func (e *Extractor) extractEnumValues(expr ast.Expr) []interface{} {
	var values []interface{}

	if comp, ok := expr.(*ast.CompositeLit); ok {
		for _, elt := range comp.Elts {
			if val := e.extractValue(elt); val != nil {
				values = append(values, val)
			}
		}
	}

	return values
}

func (e *Extractor) extractExamples(args []ast.Expr) []Example {
	var examples []Example

	// WithExamples can take either individual Example structs or a slice
	for _, arg := range args {
		if comp, ok := arg.(*ast.CompositeLit); ok {
			// Check if this is a slice of examples or a single example
			if comp.Type != nil {
				// If it has a type, check if it's []Example or Example
				if _, isSlice := comp.Type.(*ast.ArrayType); isSlice {
					// This is a slice of examples
					for _, elt := range comp.Elts {
						if exComp, ok := elt.(*ast.CompositeLit); ok {
							examples = append(examples, e.extractExample(exComp))
						}
					}
				} else {
					// Single example
					examples = append(examples, e.extractExample(comp))
				}
			} else {
				// No type specified, try to determine by content
				// If it has fields like Name, Description, it's likely a single Example
				hasExampleFields := false
				for _, elt := range comp.Elts {
					if kv, ok := elt.(*ast.KeyValueExpr); ok {
						if ident, ok := kv.Key.(*ast.Ident); ok {
							if ident.Name == "Name" || ident.Name == "Description" {
								hasExampleFields = true
								break
							}
						}
					}
				}

				if hasExampleFields {
					// Single example
					examples = append(examples, e.extractExample(comp))
				} else {
					// Likely a slice of examples
					for _, elt := range comp.Elts {
						if exComp, ok := elt.(*ast.CompositeLit); ok {
							examples = append(examples, e.extractExample(exComp))
						}
					}
				}
			}
		}
	}

	return examples
}

// extractExample extracts a single Example from a composite literal
func (e *Extractor) extractExample(comp *ast.CompositeLit) Example {
	ex := Example{}
	for _, field := range comp.Elts {
		if kv, ok := field.(*ast.KeyValueExpr); ok {
			if ident, ok := kv.Key.(*ast.Ident); ok {
				switch ident.Name {
				case "Name":
					ex.Name = e.extractString(kv.Value)
				case "Description":
					ex.Description = e.extractString(kv.Value)
				case "Input":
					ex.Input = e.extractInterface(kv.Value)
				case "Output":
					ex.Output = e.extractInterface(kv.Value)
				}
			}
		}
	}
	return ex
}

func (e *Extractor) extractInterface(expr ast.Expr) interface{} {
	switch v := expr.(type) {
	case *ast.BasicLit:
		switch v.Kind {
		case token.STRING:
			return e.unquoteString(v.Value)
		case token.INT:
			i, _ := strconv.Atoi(v.Value)
			return i
		case token.FLOAT:
			f, _ := strconv.ParseFloat(v.Value, 64)
			return f
		}
	case *ast.CompositeLit:
		// Handle map[string]interface{} or other composite types
		return e.extractMapOrSlice(v)
	case *ast.Ident:
		// Handle boolean values and nil
		switch v.Name {
		case "true":
			return true
		case "false":
			return false
		case "nil":
			return nil
		default:
			return v.Name
		}
	}
	return nil
}

func (e *Extractor) unquoteString(s string) string {
	if unquoted, err := strconv.Unquote(s); err == nil {
		return unquoted
	}
	return s
}

// extractMapOrSlice extracts a map or slice from a composite literal
func (e *Extractor) extractMapOrSlice(comp *ast.CompositeLit) interface{} {
	// Check if all elements are key-value pairs (map) or just values (slice)
	isMap := false
	for _, elt := range comp.Elts {
		if _, ok := elt.(*ast.KeyValueExpr); ok {
			isMap = true
			break
		}
	}

	if isMap {
		// Extract as map
		result := make(map[string]interface{})
		for _, elt := range comp.Elts {
			if kv, ok := elt.(*ast.KeyValueExpr); ok {
				key := e.extractString(kv.Key)
				if key != "" {
					result[key] = e.extractInterface(kv.Value)
				}
			}
		}
		return result
	} else {
		// Extract as slice
		var result []interface{}
		for _, elt := range comp.Elts {
			result = append(result, e.extractInterface(elt))
		}
		return result
	}
}

// findAndExtractToolFunction finds a function by name and extracts its builder metadata
func (e *Extractor) findAndExtractToolFunction(funcName string, file *ast.File) *ToolMetadata {
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == funcName {
			// Look for ToolBuilder pattern in the function
			metadata := &ToolMetadata{
				IsDeterministic:  true,
				EstimatedLatency: "medium",
			}

			// Find all assignments and method calls in the function
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				switch node := n.(type) {
				case *ast.AssignStmt:
					// Look for builder := atools.NewToolBuilder(...)
					if len(node.Lhs) == 1 && len(node.Rhs) == 1 {
						if _, ok := node.Lhs[0].(*ast.Ident); ok {
							if e.isToolBuilderCreation(node.Rhs[0]) {
								e.extractFromBuilderChain(node.Rhs[0], metadata)
							}
						}
					}
				case *ast.ReturnStmt:
					// Look for return atools.NewToolBuilder(...).Build()
					for _, result := range node.Results {
						if e.isToolBuilderCreation(result) {
							e.extractFromBuilderChain(result, metadata)
						}
					}
				}
				return true
			})

			return metadata
		}
	}
	return nil
}

// isToolBuilderCreation checks if an expression creates a ToolBuilder
func (e *Extractor) isToolBuilderCreation(expr ast.Expr) bool {
	// Handle method chain starting with NewToolBuilder
	if call, ok := expr.(*ast.CallExpr); ok {
		// Check for xxx.NewToolBuilder pattern
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "NewToolBuilder" {
			return true
		}
		// Check if it's a method chain
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			if innerCall, ok := sel.X.(*ast.CallExpr); ok {
				return e.isToolBuilderCreation(innerCall)
			}
		}
		// Check for direct atools.NewToolBuilder call
		if e.isNewToolBuilderCall(call) {
			return true
		}
	}
	return false
}

// isNewToolBuilderCall checks if a call is to NewToolBuilder
func (e *Extractor) isNewToolBuilderCall(call *ast.CallExpr) bool {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "NewToolBuilder" {
		return true
	}
	return false
}

// extractFromBuilderChain extracts metadata from a builder chain expression
func (e *Extractor) extractFromBuilderChain(expr ast.Expr, metadata *ToolMetadata) {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return
	}

	// Process the current call
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		methodName := sel.Sel.Name

		switch methodName {
		case "NewToolBuilder":
			if len(call.Args) >= 2 {
				metadata.Name = e.extractString(call.Args[0])
				metadata.Description = e.extractString(call.Args[1])
			}
		case "WithParameterSchema":
			if len(call.Args) > 0 {
				schema := e.extractSchema(call.Args[0])
				if schema != nil {
					metadata.ParameterSchema = schema
				}
			}
		case "WithOutputSchema":
			if len(call.Args) > 0 {
				schema := e.extractSchema(call.Args[0])
				if schema != nil {
					metadata.OutputSchema = schema
				}
			}
		case "WithCategory":
			if len(call.Args) > 0 {
				metadata.Category = e.extractString(call.Args[0])
			}
		case "WithTags":
			if len(call.Args) > 0 {
				metadata.Tags = e.extractStringSlice(call.Args[0])
			}
		case "WithVersion":
			if len(call.Args) > 0 {
				metadata.Version = e.extractString(call.Args[0])
			}
		case "WithUsageInstructions":
			if len(call.Args) > 0 {
				metadata.UsageInstructions = e.extractString(call.Args[0])
			}
		case "WithConstraints":
			// WithConstraints can take multiple string arguments
			for _, arg := range call.Args {
				if str := e.extractString(arg); str != "" {
					metadata.Constraints = append(metadata.Constraints, str)
				}
			}
		case "WithBehavior":
			if len(call.Args) >= 4 {
				metadata.IsDeterministic = e.extractBool(call.Args[0])
				metadata.IsDestructive = e.extractBool(call.Args[1])
				metadata.RequiresConfirmation = e.extractBool(call.Args[2])
				metadata.EstimatedLatency = e.extractString(call.Args[3])
			}
		case "WithDeterministic":
			if len(call.Args) > 0 {
				metadata.IsDeterministic = e.extractBool(call.Args[0])
			}
		case "WithDestructive":
			if len(call.Args) > 0 {
				metadata.IsDestructive = e.extractBool(call.Args[0])
			}
		case "WithRequiresConfirmation":
			if len(call.Args) > 0 {
				metadata.RequiresConfirmation = e.extractBool(call.Args[0])
			}
		case "WithEstimatedLatency":
			if len(call.Args) > 0 {
				metadata.EstimatedLatency = e.extractString(call.Args[0])
			}
		case "WithExamples":
			if len(call.Args) > 0 {
				metadata.Examples = e.extractExamples(call.Args)
			}
		}

		// Recursively process the receiver
		if innerCall, ok := sel.X.(*ast.CallExpr); ok {
			e.extractFromBuilderChain(innerCall, metadata)
		}
	}

	// Handle direct NewToolBuilder call
	if e.isNewToolBuilderCall(call) {
		if len(call.Args) >= 2 {
			metadata.Name = e.extractString(call.Args[0])
			metadata.Description = e.extractString(call.Args[1])
		}
	}
}

// ParseDirectory parses all Go files in a directory
func ParseDirectory(dir string) ([]ToolMetadata, error) {
	var allMetadata []ToolMetadata

	files, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		return nil, err
	}

	extractor := NewExtractor()
	fset := token.NewFileSet()

	for _, file := range files {
		// Skip test files
		if strings.HasSuffix(file, "_test.go") {
			continue
		}

		parsed, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", file, err)
		}

		metadata, err := extractor.ExtractFromFile(parsed)
		if err != nil {
			return nil, fmt.Errorf("extracting from %s: %w", file, err)
		}

		// Add package information
		pkg := parsed.Name.Name
		for i := range metadata {
			metadata[i].Package = fmt.Sprintf("github.com/lexlapax/go-llms/pkg/agent/builtins/tools/%s", pkg)
		}

		allMetadata = append(allMetadata, metadata...)
	}

	return allMetadata, nil
}
