# Schema Package Documentation

> **[Documentation Home](/docs/README.md) / [Technical Documentation](/docs/technical/README.md) / Schema Package**

The schema package in go-llms provides comprehensive JSON Schema generation and storage capabilities, designed specifically to support scripting engine integration (go-llmspell).

## Overview

The schema package consists of several key components:

1. **Schema Repositories** - Storage and versioning for schemas
2. **Schema Generators** - Generate schemas from Go structs
3. **Schema Validation** - Validate data against schemas (existing)

## Schema Repositories

### InMemorySchemaRepository

Thread-safe in-memory storage for schemas with versioning support.

```go
repo := repository.NewInMemorySchemaRepository()

// Save a schema
schema := &domain.Schema{
    Type: "object",
    Properties: map[string]domain.Property{
        "name": {Type: "string"},
        "age":  {Type: "integer"},
    },
    Required: []string{"name"},
}
err := repo.Save("user-schema", schema)

// Retrieve current version
current, err := repo.Get("user-schema")

// Get specific version
v1, err := repo.GetVersion("user-schema", 1)

// List all versions
versions, err := repo.ListVersions("user-schema")

// Export/Import
data, err := repo.Export()
err = anotherRepo.Import(data)
```

### FileSchemaRepository

File-based persistent storage with directory structure organization.

```go
repo, err := repository.NewFileSchemaRepository("/path/to/schemas")

// Same API as InMemorySchemaRepository
// Schemas are stored as JSON files with version management
```

## Schema Generators

### ReflectionSchemaGenerator

Generates schemas from Go structs using reflection, with support for:
- Struct tags (json, validate, format, etc.)
- Custom type handlers
- Recursive type handling
- Max depth control

```go
gen := generator.NewReflectionSchemaGenerator()

type User struct {
    ID    string `json:"id" validate:"required,uuid" format:"uuid"`
    Name  string `json:"name" validate:"required,min=1,max=100"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age,omitempty" validate:"min=0,max=150"`
}

schema, err := gen.GenerateSchema(User{})
```

### TagSchemaGenerator

Generates schemas primarily from struct tags, with customizable tag parsers.

```go
gen := generator.NewTagSchemaGenerator()

// Set tag priority
gen.SetTagPriority([]string{"schema", "validate", "json"})

// Register custom tag parser
gen.RegisterTagParser("custom", func(tagValue string, prop *domain.Property) error {
    prop.CustomValidator = tagValue
    return nil
})

type Example struct {
    Field string `schema:"type=string,format=email,required"`
}

schema, err := gen.GenerateSchema(Example{})
```

## Custom Type Handlers

Both generators support custom type handlers for specialized types:

```go
gen.RegisterTypeHandler(reflect.TypeOf(time.Duration(0)), 
    func(t reflect.Type, tag reflect.StructTag) (domain.Property, error) {
        return domain.Property{
            Type:        "string",
            Format:      "duration",
            Description: "Duration in string format",
        }, nil
    })
```

## Integration with go-llmspell

The schema package was designed with the following scripting engine requirements in mind:

1. **Dynamic Registration** - Schemas can be registered/unregistered at runtime
2. **Serialization** - All schemas and properties are JSON serializable
3. **Type Conversion** - Bridge-friendly type system for script integration
4. **Metadata Support** - Rich metadata for tool discovery and documentation

## Examples

See the example programs for comprehensive demonstrations:
- `cmd/examples/schema-repository/` - Repository usage examples
- `cmd/examples/schema-generator/` - Generator usage examples

## Best Practices

1. **Use appropriate generator**: 
   - ReflectionSchemaGenerator for automatic schema generation
   - TagSchemaGenerator when you need fine control via tags

2. **Version management**:
   - Always version schemas when they change
   - Use SetCurrentVersion to control which version is active

3. **Custom types**:
   - Register handlers for domain-specific types
   - Ensure serialization compatibility

4. **Testing**:
   - Use the provided mock implementations
   - Validate generated schemas against expected structure