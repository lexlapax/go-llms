# Schema System: JSON Schema Validation and Type Conversion

> **[Project Root](/) / [Documentation](../..) / [Technical Documentation](../../technical) / [Advanced Topics](../../technical/advanced) / Schema System**

Comprehensive guide to the Go-LLMs schema system, covering JSON Schema validation, type conversion mechanisms, schema inference, custom validators, schema composition, runtime validation, and advanced schema patterns for building robust, type-safe LLM applications.

## Schema Architecture

### 1. Core Schema Interfaces

```go
// Schema represents a validation and conversion schema
type Schema interface {
    // Validation
    Validate(value interface{}) error
    ValidateWithContext(ctx context.Context, value interface{}) error
    
    // Type conversion
    Convert(value interface{}) (interface{}, error)
    ConvertWithOptions(value interface{}, options ConvertOptions) (interface{}, error)
    
    // Schema introspection
    GetType() SchemaType
    GetConstraints() map[string]interface{}
    GetMetadata() SchemaMetadata
    
    // Schema composition
    Merge(other Schema) (Schema, error)
    Extend(extensions map[string]interface{}) (Schema, error)
    
    // Serialization
    ToJSONSchema() (map[string]interface{}, error)
    FromJSONSchema(schema map[string]interface{}) error
}

// SchemaRegistry manages schema registration and discovery
type SchemaRegistry interface {
    // Registration
    Register(name string, schema Schema) error
    Unregister(name string) error
    Get(name string) (Schema, error)
    List() []string
    
    // Discovery
    FindByType(schemaType SchemaType) []Schema
    FindByTag(tag string) []Schema
    Search(query SchemaQuery) []Schema
    
    // Validation
    ValidateValue(schemaName string, value interface{}) error
    ConvertValue(schemaName string, value interface{}) (interface{}, error)
    
    // Schema management
    ImportFromFile(filename string) error
    ExportToFile(filename string) error
    Clear() error
}

type SchemaType string

const (
    SchemaTypeObject  SchemaType = "object"
    SchemaTypeArray   SchemaType = "array"
    SchemaTypeString  SchemaType = "string"
    SchemaTypeNumber  SchemaType = "number"
    SchemaTypeInteger SchemaType = "integer"
    SchemaTypeBoolean SchemaType = "boolean"
    SchemaTypeNull    SchemaType = "null"
    SchemaTypeAny     SchemaType = "any"
    SchemaTypeRef     SchemaType = "$ref"
)

type SchemaMetadata struct {
    ID          string                 `json:"id,omitempty"`
    Title       string                 `json:"title,omitempty"`
    Description string                 `json:"description,omitempty"`
    Version     string                 `json:"version,omitempty"`
    Tags        []string               `json:"tags,omitempty"`
    Examples    []interface{}          `json:"examples,omitempty"`
    Default     interface{}            `json:"default,omitempty"`
    ReadOnly    bool                   `json:"readOnly,omitempty"`
    WriteOnly   bool                   `json:"writeOnly,omitempty"`
    Deprecated  bool                   `json:"deprecated,omitempty"`
}

type ConvertOptions struct {
    StrictMode      bool              `json:"strict_mode"`
    CoerceTypes     bool              `json:"coerce_types"`
    RemoveAdditional bool             `json:"remove_additional"`
    UseDefaults     bool              `json:"use_defaults"`
    CustomConverters map[string]Converter `json:"-"`
}

type SchemaQuery struct {
    Type        SchemaType `json:"type,omitempty"`
    Tags        []string   `json:"tags,omitempty"`
    Keywords    []string   `json:"keywords,omitempty"`
    Properties  []string   `json:"properties,omitempty"`
    Pattern     string     `json:"pattern,omitempty"`
}
```

### 2. JSON Schema Implementation

```go
// JSONSchema implements JSON Schema Draft 7 specification
type JSONSchema struct {
    // Schema metadata
    schema      string                 `json:"$schema,omitempty"`
    id          string                 `json:"$id,omitempty"`
    ref         string                 `json:"$ref,omitempty"`
    title       string                 `json:"title,omitempty"`
    description string                 `json:"description,omitempty"`
    
    // Type constraints
    Type    SchemaType    `json:"type,omitempty"`
    Enum    []interface{} `json:"enum,omitempty"`
    Const   interface{}   `json:"const,omitempty"`
    
    // Numeric constraints
    MultipleOf       *float64 `json:"multipleOf,omitempty"`
    Maximum          *float64 `json:"maximum,omitempty"`
    ExclusiveMaximum *float64 `json:"exclusiveMaximum,omitempty"`
    Minimum          *float64 `json:"minimum,omitempty"`
    ExclusiveMinimum *float64 `json:"exclusiveMinimum,omitempty"`
    
    // String constraints
    MaxLength *int    `json:"maxLength,omitempty"`
    MinLength *int    `json:"minLength,omitempty"`
    Pattern   string  `json:"pattern,omitempty"`
    Format    string  `json:"format,omitempty"`
    
    // Array constraints
    Items           *JSONSchema   `json:"items,omitempty"`
    AdditionalItems *JSONSchema   `json:"additionalItems,omitempty"`
    MaxItems        *int          `json:"maxItems,omitempty"`
    MinItems        *int          `json:"minItems,omitempty"`
    UniqueItems     bool          `json:"uniqueItems,omitempty"`
    Contains        *JSONSchema   `json:"contains,omitempty"`
    
    // Object constraints
    MaxProperties        *int                    `json:"maxProperties,omitempty"`
    MinProperties        *int                    `json:"minProperties,omitempty"`
    Required             []string                `json:"required,omitempty"`
    Properties           map[string]*JSONSchema  `json:"properties,omitempty"`
    PatternProperties    map[string]*JSONSchema  `json:"patternProperties,omitempty"`
    AdditionalProperties interface{}             `json:"additionalProperties,omitempty"`
    Dependencies         map[string]interface{}  `json:"dependencies,omitempty"`
    PropertyNames        *JSONSchema             `json:"propertyNames,omitempty"`
    
    // Logical operators
    AllOf []JSONSchema `json:"allOf,omitempty"`
    AnyOf []JSONSchema `json:"anyOf,omitempty"`
    OneOf []JSONSchema `json:"oneOf,omitempty"`
    Not   *JSONSchema  `json:"not,omitempty"`
    
    // Conditional schemas
    If   *JSONSchema `json:"if,omitempty"`
    Then *JSONSchema `json:"then,omitempty"`
    Else *JSONSchema `json:"else,omitempty"`
    
    // Metadata
    Default     interface{}   `json:"default,omitempty"`
    Examples    []interface{} `json:"examples,omitempty"`
    ReadOnly    bool          `json:"readOnly,omitempty"`
    WriteOnly   bool          `json:"writeOnly,omitempty"`
    Deprecated  bool          `json:"deprecated,omitempty"`
    
    // Custom extensions
    Extensions map[string]interface{} `json:"-"`
}

// Validate validates a value against the schema
func (js *JSONSchema) Validate(value interface{}) error {
    return js.ValidateWithContext(context.Background(), value)
}

// ValidateWithContext validates with context for cancellation
func (js *JSONSchema) ValidateWithContext(ctx context.Context, value interface{}) error {
    validator := &schemaValidator{
        schema:  js,
        context: ctx,
        path:    "",
        visited: make(map[string]bool),
    }
    
    return validator.validate(value)
}

type schemaValidator struct {
    schema  *JSONSchema
    context context.Context
    path    string
    visited map[string]bool
}

func (v *schemaValidator) validate(value interface{}) error {
    // Check for context cancellation
    select {
    case <-v.context.Done():
        return v.context.Err()
    default:
    }
    
    // Handle $ref
    if v.schema.ref != "" {
        return v.validateRef(value)
    }
    
    // Validate type
    if err := v.validateType(value); err != nil {
        return err
    }
    
    // Validate enum
    if err := v.validateEnum(value); err != nil {
        return err
    }
    
    // Validate const
    if err := v.validateConst(value); err != nil {
        return err
    }
    
    // Type-specific validation
    switch v.schema.Type {
    case SchemaTypeString:
        return v.validateString(value)
    case SchemaTypeNumber, SchemaTypeInteger:
        return v.validateNumber(value)
    case SchemaTypeArray:
        return v.validateArray(value)
    case SchemaTypeObject:
        return v.validateObject(value)
    case SchemaTypeBoolean:
        return v.validateBoolean(value)
    case SchemaTypeNull:
        return v.validateNull(value)
    }
    
    // Validate logical operators
    if err := v.validateLogical(value); err != nil {
        return err
    }
    
    // Validate conditional schemas
    if err := v.validateConditional(value); err != nil {
        return err
    }
    
    return nil
}

func (v *schemaValidator) validateType(value interface{}) error {
    if v.schema.Type == "" {
        return nil // Type not specified
    }
    
    actualType := getValueType(value)
    
    // Special handling for integer (subset of number)
    if v.schema.Type == SchemaTypeInteger && actualType == SchemaTypeNumber {
        if num, ok := value.(float64); ok && num == float64(int64(num)) {
            return nil // It's a valid integer
        }
    }
    
    if actualType != v.schema.Type {
        return NewValidationError(v.path, "type", 
            fmt.Sprintf("expected %s, got %s", v.schema.Type, actualType))
    }
    
    return nil
}

func (v *schemaValidator) validateString(value interface{}) error {
    str, ok := value.(string)
    if !ok {
        return nil // Type validation handles this
    }
    
    // Length constraints
    if v.schema.MinLength != nil && len(str) < *v.schema.MinLength {
        return NewValidationError(v.path, "minLength",
            fmt.Sprintf("string length %d is less than minimum %d", len(str), *v.schema.MinLength))
    }
    
    if v.schema.MaxLength != nil && len(str) > *v.schema.MaxLength {
        return NewValidationError(v.path, "maxLength",
            fmt.Sprintf("string length %d exceeds maximum %d", len(str), *v.schema.MaxLength))
    }
    
    // Pattern validation
    if v.schema.Pattern != "" {
        matched, err := regexp.MatchString(v.schema.Pattern, str)
        if err != nil {
            return NewValidationError(v.path, "pattern",
                fmt.Sprintf("invalid pattern: %v", err))
        }
        if !matched {
            return NewValidationError(v.path, "pattern",
                fmt.Sprintf("string does not match pattern %s", v.schema.Pattern))
        }
    }
    
    // Format validation
    if v.schema.Format != "" {
        if err := v.validateFormat(str, v.schema.Format); err != nil {
            return err
        }
    }
    
    return nil
}

func (v *schemaValidator) validateNumber(value interface{}) error {
    var num float64
    var ok bool
    
    switch val := value.(type) {
    case float64:
        num = val
        ok = true
    case int:
        num = float64(val)
        ok = true
    case int64:
        num = float64(val)
        ok = true
    }
    
    if !ok {
        return nil // Type validation handles this
    }
    
    // Numeric constraints
    if v.schema.MultipleOf != nil {
        if math.Mod(num, *v.schema.MultipleOf) != 0 {
            return NewValidationError(v.path, "multipleOf",
                fmt.Sprintf("number %g is not a multiple of %g", num, *v.schema.MultipleOf))
        }
    }
    
    if v.schema.Maximum != nil && num > *v.schema.Maximum {
        return NewValidationError(v.path, "maximum",
            fmt.Sprintf("number %g exceeds maximum %g", num, *v.schema.Maximum))
    }
    
    if v.schema.ExclusiveMaximum != nil && num >= *v.schema.ExclusiveMaximum {
        return NewValidationError(v.path, "exclusiveMaximum",
            fmt.Sprintf("number %g is not less than %g", num, *v.schema.ExclusiveMaximum))
    }
    
    if v.schema.Minimum != nil && num < *v.schema.Minimum {
        return NewValidationError(v.path, "minimum",
            fmt.Sprintf("number %g is less than minimum %g", num, *v.schema.Minimum))
    }
    
    if v.schema.ExclusiveMinimum != nil && num <= *v.schema.ExclusiveMinimum {
        return NewValidationError(v.path, "exclusiveMinimum",
            fmt.Sprintf("number %g is not greater than %g", num, *v.schema.ExclusiveMinimum))
    }
    
    return nil
}

func (v *schemaValidator) validateArray(value interface{}) error {
    arr, ok := value.([]interface{})
    if !ok {
        return nil // Type validation handles this
    }
    
    // Length constraints
    if v.schema.MinItems != nil && len(arr) < *v.schema.MinItems {
        return NewValidationError(v.path, "minItems",
            fmt.Sprintf("array length %d is less than minimum %d", len(arr), *v.schema.MinItems))
    }
    
    if v.schema.MaxItems != nil && len(arr) > *v.schema.MaxItems {
        return NewValidationError(v.path, "maxItems",
            fmt.Sprintf("array length %d exceeds maximum %d", len(arr), *v.schema.MaxItems))
    }
    
    // Unique items
    if v.schema.UniqueItems {
        seen := make(map[string]bool)
        for i, item := range arr {
            key := fmt.Sprintf("%v", item)
            if seen[key] {
                return NewValidationError(fmt.Sprintf("%s[%d]", v.path, i), "uniqueItems",
                    "duplicate item found in array")
            }
            seen[key] = true
        }
    }
    
    // Validate items
    if v.schema.Items != nil {
        for i, item := range arr {
            itemValidator := &schemaValidator{
                schema:  v.schema.Items,
                context: v.context,
                path:    fmt.Sprintf("%s[%d]", v.path, i),
                visited: v.visited,
            }
            
            if err := itemValidator.validate(item); err != nil {
                return err
            }
        }
    }
    
    return nil
}

func (v *schemaValidator) validateObject(value interface{}) error {
    obj, ok := value.(map[string]interface{})
    if !ok {
        return nil // Type validation handles this
    }
    
    // Property count constraints
    if v.schema.MinProperties != nil && len(obj) < *v.schema.MinProperties {
        return NewValidationError(v.path, "minProperties",
            fmt.Sprintf("object has %d properties, minimum is %d", len(obj), *v.schema.MinProperties))
    }
    
    if v.schema.MaxProperties != nil && len(obj) > *v.schema.MaxProperties {
        return NewValidationError(v.path, "maxProperties",
            fmt.Sprintf("object has %d properties, maximum is %d", len(obj), *v.schema.MaxProperties))
    }
    
    // Required properties
    for _, required := range v.schema.Required {
        if _, exists := obj[required]; !exists {
            return NewValidationError(fmt.Sprintf("%s.%s", v.path, required), "required",
                fmt.Sprintf("required property '%s' is missing", required))
        }
    }
    
    // Validate properties
    for propName, propValue := range obj {
        if err := v.validateProperty(propName, propValue); err != nil {
            return err
        }
    }
    
    return nil
}

func (v *schemaValidator) validateProperty(name string, value interface{}) error {
    path := fmt.Sprintf("%s.%s", v.path, name)
    
    // Check exact property match
    if v.schema.Properties != nil {
        if propSchema, exists := v.schema.Properties[name]; exists {
            propValidator := &schemaValidator{
                schema:  propSchema,
                context: v.context,
                path:    path,
                visited: v.visited,
            }
            return propValidator.validate(value)
        }
    }
    
    // Check pattern properties
    if v.schema.PatternProperties != nil {
        for pattern, patternSchema := range v.schema.PatternProperties {
            matched, err := regexp.MatchString(pattern, name)
            if err != nil {
                continue
            }
            if matched {
                propValidator := &schemaValidator{
                    schema:  patternSchema,
                    context: v.context,
                    path:    path,
                    visited: v.visited,
                }
                return propValidator.validate(value)
            }
        }
    }
    
    // Check additional properties
    if v.schema.AdditionalProperties != nil {
        switch ap := v.schema.AdditionalProperties.(type) {
        case bool:
            if !ap {
                return NewValidationError(path, "additionalProperties",
                    fmt.Sprintf("additional property '%s' not allowed", name))
            }
        case *JSONSchema:
            propValidator := &schemaValidator{
                schema:  ap,
                context: v.context,
                path:    path,
                visited: v.visited,
            }
            return propValidator.validate(value)
        }
    }
    
    return nil
}
```

### 3. Type Conversion System

```go
// TypeConverter handles type conversions with schema guidance
type TypeConverter struct {
    converters map[ConversionKey]Converter
    registry   *SchemaRegistry
    options    ConvertOptions
}

type ConversionKey struct {
    From SchemaType `json:"from"`
    To   SchemaType `json:"to"`
}

type Converter interface {
    Convert(value interface{}, schema Schema) (interface{}, error)
    CanConvert(from, to SchemaType) bool
    Priority() int
}

// NewTypeConverter creates a new type converter
func NewTypeConverter(registry *SchemaRegistry) *TypeConverter {
    tc := &TypeConverter{
        converters: make(map[ConversionKey]Converter),
        registry:   registry,
        options:    DefaultConvertOptions(),
    }
    
    // Register built-in converters
    tc.registerBuiltinConverters()
    
    return tc
}

// Convert converts a value according to the target schema
func (tc *TypeConverter) Convert(value interface{}, targetSchema Schema) (interface{}, error) {
    return tc.ConvertWithOptions(value, targetSchema, tc.options)
}

// ConvertWithOptions converts with specific options
func (tc *TypeConverter) ConvertWithOptions(value interface{}, targetSchema Schema, options ConvertOptions) (interface{}, error) {
    if value == nil {
        return tc.handleNullValue(targetSchema, options)
    }
    
    sourceType := getValueType(value)
    targetType := targetSchema.GetType()
    
    // No conversion needed if types match
    if sourceType == targetType {
        return tc.validateAndClean(value, targetSchema, options)
    }
    
    // Find appropriate converter
    converter, err := tc.findConverter(sourceType, targetType)
    if err != nil {
        if !options.CoerceTypes {
            return nil, fmt.Errorf("cannot convert %s to %s: %w", sourceType, targetType, err)
        }
        
        // Try coercion
        return tc.coerceType(value, targetType, options)
    }
    
    // Perform conversion
    converted, err := converter.Convert(value, targetSchema)
    if err != nil {
        return nil, fmt.Errorf("conversion failed: %w", err)
    }
    
    // Validate converted value
    if err := targetSchema.Validate(converted); err != nil {
        return nil, fmt.Errorf("converted value failed validation: %w", err)
    }
    
    return converted, nil
}

// Built-in converters
func (tc *TypeConverter) registerBuiltinConverters() {
    // String converters
    tc.RegisterConverter(&StringToNumberConverter{})
    tc.RegisterConverter(&StringToBooleanConverter{})
    tc.RegisterConverter(&StringToArrayConverter{})
    tc.RegisterConverter(&StringToObjectConverter{})
    
    // Number converters
    tc.RegisterConverter(&NumberToStringConverter{})
    tc.RegisterConverter(&NumberToBooleanConverter{})
    tc.RegisterConverter(&NumberToIntegerConverter{})
    
    // Boolean converters
    tc.RegisterConverter(&BooleanToStringConverter{})
    tc.RegisterConverter(&BooleanToNumberConverter{})
    
    // Array converters
    tc.RegisterConverter(&ArrayToStringConverter{})
    tc.RegisterConverter(&ArrayToObjectConverter{})
    
    // Object converters
    tc.RegisterConverter(&ObjectToStringConverter{})
    tc.RegisterConverter(&ObjectToArrayConverter{})
}

// StringToNumberConverter converts strings to numbers
type StringToNumberConverter struct{}

func (c *StringToNumberConverter) Convert(value interface{}, schema Schema) (interface{}, error) {
    str, ok := value.(string)
    if !ok {
        return nil, fmt.Errorf("expected string, got %T", value)
    }
    
    // Try parsing as integer first if target is integer
    if schema.GetType() == SchemaTypeInteger {
        if intVal, err := strconv.ParseInt(str, 10, 64); err == nil {
            return intVal, nil
        }
    }
    
    // Parse as float
    floatVal, err := strconv.ParseFloat(str, 64)
    if err != nil {
        return nil, fmt.Errorf("cannot parse '%s' as number: %w", str, err)
    }
    
    return floatVal, nil
}

func (c *StringToNumberConverter) CanConvert(from, to SchemaType) bool {
    return from == SchemaTypeString && (to == SchemaTypeNumber || to == SchemaTypeInteger)
}

func (c *StringToNumberConverter) Priority() int {
    return 100
}

// StringToBooleanConverter converts strings to booleans
type StringToBooleanConverter struct{}

func (c *StringToBooleanConverter) Convert(value interface{}, schema Schema) (interface{}, error) {
    str, ok := value.(string)
    if !ok {
        return nil, fmt.Errorf("expected string, got %T", value)
    }
    
    switch strings.ToLower(str) {
    case "true", "yes", "1", "on", "enabled":
        return true, nil
    case "false", "no", "0", "off", "disabled":
        return false, nil
    default:
        return nil, fmt.Errorf("cannot parse '%s' as boolean", str)
    }
}

func (c *StringToBooleanConverter) CanConvert(from, to SchemaType) bool {
    return from == SchemaTypeString && to == SchemaTypeBoolean
}

func (c *StringToBooleanConverter) Priority() int {
    return 100
}

// NumberToStringConverter converts numbers to strings
type NumberToStringConverter struct{}

func (c *NumberToStringConverter) Convert(value interface{}, schema Schema) (interface{}, error) {
    switch val := value.(type) {
    case float64:
        return strconv.FormatFloat(val, 'g', -1, 64), nil
    case int:
        return strconv.Itoa(val), nil
    case int64:
        return strconv.FormatInt(val, 10), nil
    default:
        return nil, fmt.Errorf("expected number, got %T", value)
    }
}

func (c *NumberToStringConverter) CanConvert(from, to SchemaType) bool {
    return (from == SchemaTypeNumber || from == SchemaTypeInteger) && to == SchemaTypeString
}

func (c *NumberToStringConverter) Priority() int {
    return 100
}
```

### 4. Schema Inference

```go
// SchemaInferrer automatically infers schemas from data
type SchemaInferrer struct {
    options InferenceOptions
    samples map[string][]interface{}
}

type InferenceOptions struct {
    MaxSamples      int     `yaml:"max_samples" json:"max_samples"`
    MinConfidence   float64 `yaml:"min_confidence" json:"min_confidence"`
    InferFormats    bool    `yaml:"infer_formats" json:"infer_formats"`
    InferPatterns   bool    `yaml:"infer_patterns" json:"infer_patterns"`
    GenerateExamples bool   `yaml:"generate_examples" json:"generate_examples"`
    MergeSchemas    bool    `yaml:"merge_schemas" json:"merge_schemas"`
}

// InferSchema infers a schema from a set of samples
func (si *SchemaInferrer) InferSchema(samples []interface{}) (*JSONSchema, error) {
    if len(samples) == 0 {
        return nil, fmt.Errorf("no samples provided for inference")
    }
    
    // Analyze samples
    analysis := si.analyzeSamples(samples)
    
    // Generate schema based on analysis
    schema := &JSONSchema{}
    
    // Infer type
    schema.Type = analysis.PrimaryType
    
    // Add type-specific constraints
    switch analysis.PrimaryType {
    case SchemaTypeString:
        si.inferStringConstraints(schema, analysis)
    case SchemaTypeNumber, SchemaTypeInteger:
        si.inferNumericConstraints(schema, analysis)
    case SchemaTypeArray:
        si.inferArrayConstraints(schema, analysis, samples)
    case SchemaTypeObject:
        si.inferObjectConstraints(schema, analysis, samples)
    case SchemaTypeBoolean:
        // Boolean doesn't need additional constraints
    }
    
    // Add metadata
    if si.options.GenerateExamples {
        schema.Examples = si.selectExamples(samples, 3)
    }
    
    return schema, nil
}

type SampleAnalysis struct {
    PrimaryType     SchemaType
    TypeCounts      map[SchemaType]int
    Confidence      float64
    StringAnalysis  *StringAnalysis
    NumericAnalysis *NumericAnalysis
    ArrayAnalysis   *ArrayAnalysis
    ObjectAnalysis  *ObjectAnalysis
}

type StringAnalysis struct {
    MinLength   int
    MaxLength   int
    Patterns    []string
    Formats     []string
    CommonWords []string
    Encoding    string
}

type NumericAnalysis struct {
    MinValue    float64
    MaxValue    float64
    IsInteger   bool
    Precision   int
    MultipleOf  *float64
    Distribution string
}

type ArrayAnalysis struct {
    MinItems     int
    MaxItems     int
    ItemTypes    map[SchemaType]int
    UniqueItems  bool
    ItemSchema   *JSONSchema
}

type ObjectAnalysis struct {
    RequiredProps   []string
    OptionalProps   []string
    PropertyTypes   map[string]SchemaType
    PropertySchemas map[string]*JSONSchema
    AdditionalProps bool
}

// analyzeSamples performs comprehensive analysis of samples
func (si *SchemaInferrer) analyzeSamples(samples []interface{}) *SampleAnalysis {
    analysis := &SampleAnalysis{
        TypeCounts: make(map[SchemaType]int),
    }
    
    // Count types
    for _, sample := range samples {
        sampleType := getValueType(sample)
        analysis.TypeCounts[sampleType]++
    }
    
    // Determine primary type
    maxCount := 0
    for schemaType, count := range analysis.TypeCounts {
        if count > maxCount {
            maxCount = count
            analysis.PrimaryType = schemaType
        }
    }
    
    // Calculate confidence
    analysis.Confidence = float64(maxCount) / float64(len(samples))
    
    // Perform type-specific analysis
    switch analysis.PrimaryType {
    case SchemaTypeString:
        analysis.StringAnalysis = si.analyzeStrings(samples)
    case SchemaTypeNumber, SchemaTypeInteger:
        analysis.NumericAnalysis = si.analyzeNumbers(samples)
    case SchemaTypeArray:
        analysis.ArrayAnalysis = si.analyzeArrays(samples)
    case SchemaTypeObject:
        analysis.ObjectAnalysis = si.analyzeObjects(samples)
    }
    
    return analysis
}

func (si *SchemaInferrer) analyzeStrings(samples []interface{}) *StringAnalysis {
    analysis := &StringAnalysis{
        MinLength: math.MaxInt32,
        MaxLength: 0,
        Patterns:  make([]string, 0),
        Formats:   make([]string, 0),
    }
    
    for _, sample := range samples {
        if str, ok := sample.(string); ok {
            length := len(str)
            if length < analysis.MinLength {
                analysis.MinLength = length
            }
            if length > analysis.MaxLength {
                analysis.MaxLength = length
            }
            
            // Infer formats if enabled
            if si.options.InferFormats {
                if format := si.inferStringFormat(str); format != "" {
                    analysis.Formats = append(analysis.Formats, format)
                }
            }
        }
    }
    
    return analysis
}

func (si *SchemaInferrer) inferStringFormat(str string) string {
    // Email format
    if matched, _ := regexp.MatchString(`^[^@]+@[^@]+\.[^@]+$`, str); matched {
        return "email"
    }
    
    // URI format
    if matched, _ := regexp.MatchString(`^https?://`, str); matched {
        return "uri"
    }
    
    // Date format
    if matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, str); matched {
        return "date"
    }
    
    // Date-time format
    if matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`, str); matched {
        return "date-time"
    }
    
    // UUID format
    if matched, _ := regexp.MatchString(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, str); matched {
        return "uuid"
    }
    
    return ""
}

func (si *SchemaInferrer) analyzeNumbers(samples []interface{}) *NumericAnalysis {
    analysis := &NumericAnalysis{
        MinValue:  math.Inf(1),
        MaxValue:  math.Inf(-1),
        IsInteger: true,
    }
    
    for _, sample := range samples {
        var num float64
        var ok bool
        
        switch val := sample.(type) {
        case float64:
            num = val
            ok = true
            if val != float64(int64(val)) {
                analysis.IsInteger = false
            }
        case int:
            num = float64(val)
            ok = true
        case int64:
            num = float64(val)
            ok = true
        }
        
        if ok {
            if num < analysis.MinValue {
                analysis.MinValue = num
            }
            if num > analysis.MaxValue {
                analysis.MaxValue = num
            }
        }
    }
    
    return analysis
}
```

### 5. Schema Composition and Extension

```go
// SchemaComposer handles schema composition operations
type SchemaComposer struct {
    registry *SchemaRegistry
    resolver *SchemaResolver
}

// MergeSchemas combines multiple schemas using logical operators
func (sc *SchemaComposer) MergeSchemas(schemas []*JSONSchema, operator string) (*JSONSchema, error) {
    if len(schemas) == 0 {
        return nil, fmt.Errorf("no schemas to merge")
    }
    
    if len(schemas) == 1 {
        return schemas[0], nil
    }
    
    merged := &JSONSchema{}
    
    switch operator {
    case "allOf":
        merged.AllOf = make([]JSONSchema, len(schemas))
        for i, schema := range schemas {
            merged.AllOf[i] = *schema
        }
        
    case "anyOf":
        merged.AnyOf = make([]JSONSchema, len(schemas))
        for i, schema := range schemas {
            merged.AnyOf[i] = *schema
        }
        
    case "oneOf":
        merged.OneOf = make([]JSONSchema, len(schemas))
        for i, schema := range schemas {
            merged.OneOf[i] = *schema
        }
        
    case "merge":
        // Intelligent merging
        return sc.intelligentMerge(schemas)
        
    default:
        return nil, fmt.Errorf("unsupported merge operator: %s", operator)
    }
    
    return merged, nil
}

// intelligentMerge performs intelligent schema merging
func (sc *SchemaComposer) intelligentMerge(schemas []*JSONSchema) (*JSONSchema, error) {
    merged := &JSONSchema{}
    
    // Merge types
    types := make(map[SchemaType]bool)
    for _, schema := range schemas {
        if schema.Type != "" {
            types[schema.Type] = true
        }
    }
    
    if len(types) == 1 {
        for schemaType := range types {
            merged.Type = schemaType
        }
    } else if len(types) > 1 {
        // Multiple types - use anyOf
        merged.AnyOf = make([]JSONSchema, len(schemas))
        for i, schema := range schemas {
            merged.AnyOf[i] = *schema
        }
        return merged, nil
    }
    
    // Merge constraints based on type
    switch merged.Type {
    case SchemaTypeString:
        sc.mergeStringConstraints(merged, schemas)
    case SchemaTypeNumber, SchemaTypeInteger:
        sc.mergeNumericConstraints(merged, schemas)
    case SchemaTypeArray:
        sc.mergeArrayConstraints(merged, schemas)
    case SchemaTypeObject:
        sc.mergeObjectConstraints(merged, schemas)
    }
    
    // Merge metadata
    sc.mergeMetadata(merged, schemas)
    
    return merged, nil
}

func (sc *SchemaComposer) mergeStringConstraints(merged *JSONSchema, schemas []*JSONSchema) {
    var minLengths, maxLengths []int
    var patterns []string
    
    for _, schema := range schemas {
        if schema.MinLength != nil {
            minLengths = append(minLengths, *schema.MinLength)
        }
        if schema.MaxLength != nil {
            maxLengths = append(maxLengths, *schema.MaxLength)
        }
        if schema.Pattern != "" {
            patterns = append(patterns, schema.Pattern)
        }
    }
    
    // Use most restrictive constraints
    if len(minLengths) > 0 {
        merged.MinLength = &minLengths[0]
        for _, length := range minLengths {
            if length > *merged.MinLength {
                *merged.MinLength = length
            }
        }
    }
    
    if len(maxLengths) > 0 {
        merged.MaxLength = &maxLengths[0]
        for _, length := range maxLengths {
            if length < *merged.MaxLength {
                *merged.MaxLength = length
            }
        }
    }
    
    // Combine patterns
    if len(patterns) > 0 {
        // For now, just use the first pattern
        // More sophisticated pattern merging could be implemented
        merged.Pattern = patterns[0]
    }
}

func (sc *SchemaComposer) mergeObjectConstraints(merged *JSONSchema, schemas []*JSONSchema) {
    allRequired := make(map[string]int)
    allProperties := make(map[string][]*JSONSchema)
    
    for _, schema := range schemas {
        // Count required properties
        for _, required := range schema.Required {
            allRequired[required]++
        }
        
        // Collect property schemas
        for propName, propSchema := range schema.Properties {
            if allProperties[propName] == nil {
                allProperties[propName] = make([]*JSONSchema, 0)
            }
            allProperties[propName] = append(allProperties[propName], propSchema)
        }
    }
    
    // Property is required if it appears in all schemas
    merged.Required = make([]string, 0)
    for propName, count := range allRequired {
        if count == len(schemas) {
            merged.Required = append(merged.Required, propName)
        }
    }
    
    // Merge property schemas
    merged.Properties = make(map[string]*JSONSchema)
    for propName, propSchemas := range allProperties {
        if len(propSchemas) == 1 {
            merged.Properties[propName] = propSchemas[0]
        } else {
            // Merge property schemas recursively
            mergedProp, err := sc.intelligentMerge(propSchemas)
            if err == nil {
                merged.Properties[propName] = mergedProp
            }
        }
    }
}

// ExtendSchema extends a base schema with additional constraints
func (sc *SchemaComposer) ExtendSchema(base *JSONSchema, extensions map[string]interface{}) (*JSONSchema, error) {
    extended := *base // Copy base schema
    
    for key, value := range extensions {
        switch key {
        case "required":
            if requiredList, ok := value.([]interface{}); ok {
                for _, req := range requiredList {
                    if reqStr, ok := req.(string); ok {
                        extended.Required = append(extended.Required, reqStr)
                    }
                }
            }
            
        case "properties":
            if props, ok := value.(map[string]interface{}); ok {
                if extended.Properties == nil {
                    extended.Properties = make(map[string]*JSONSchema)
                }
                for propName, propDef := range props {
                    if propSchema, err := sc.parseSchemaDefinition(propDef); err == nil {
                        extended.Properties[propName] = propSchema
                    }
                }
            }
            
        case "minLength":
            if minLen, ok := value.(int); ok {
                extended.MinLength = &minLen
            }
            
        case "maxLength":
            if maxLen, ok := value.(int); ok {
                extended.MaxLength = &maxLen
            }
            
        // Add more extension cases as needed
        }
    }
    
    return &extended, nil
}

// parseSchemaDefinition parses a schema definition from interface{}
func (sc *SchemaComposer) parseSchemaDefinition(def interface{}) (*JSONSchema, error) {
    switch definition := def.(type) {
    case map[string]interface{}:
        // Parse as JSON Schema object
        return sc.parseJSONSchemaObject(definition)
    case string:
        // Reference to another schema
        return sc.resolver.ResolveReference(definition)
    default:
        return nil, fmt.Errorf("unsupported schema definition type: %T", def)
    }
}
```

This comprehensive schema system provides robust validation, type conversion, inference, and composition capabilities for building type-safe LLM applications with Go-LLMs.