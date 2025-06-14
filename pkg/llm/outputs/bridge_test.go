package outputs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBridgeAdapter_ConvertSchemaFromBridge(t *testing.T) {
	adapter := NewBridgeAdapter()

	testCases := []struct {
		name         string
		bridgeSchema map[string]interface{}
		validate     func(t *testing.T, schema *OutputSchema)
		wantErr      bool
	}{
		{
			name: "String schema with enum",
			bridgeSchema: map[string]interface{}{
				"type":        "string",
				"description": "Status field",
				"required":    true,
				"enum":        []interface{}{"active", "inactive", "pending"},
			},
			validate: func(t *testing.T, schema *OutputSchema) {
				assert.Equal(t, TypeString, schema.Type)
				assert.Equal(t, "Status field", schema.Description)
				assert.True(t, *schema.Required)
				assert.Equal(t, []string{"active", "inactive", "pending"}, schema.Enum)
			},
		},
		{
			name: "Number schema with constraints",
			bridgeSchema: map[string]interface{}{
				"type":    "number",
				"minimum": 0.0,
				"maximum": 100.0,
			},
			validate: func(t *testing.T, schema *OutputSchema) {
				assert.Equal(t, TypeNumber, schema.Type)
				assert.Equal(t, 0.0, *schema.Minimum)
				assert.Equal(t, 100.0, *schema.Maximum)
			},
		},
		{
			name: "Array schema with items",
			bridgeSchema: map[string]interface{}{
				"type":     "array",
				"minItems": 1,
				"maxItems": 10,
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			validate: func(t *testing.T, schema *OutputSchema) {
				assert.Equal(t, TypeArray, schema.Type)
				assert.Equal(t, 1, *schema.MinItems)
				assert.Equal(t, 10, *schema.MaxItems)
				assert.NotNil(t, schema.Items)
				assert.Equal(t, TypeString, schema.Items.Type)
			},
		},
		{
			name: "Object schema with properties",
			bridgeSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":     "string",
						"required": true,
					},
					"age": map[string]interface{}{
						"type":    "integer",
						"minimum": 0,
					},
				},
				"required":             []interface{}{"name"},
				"additionalProperties": false,
			},
			validate: func(t *testing.T, schema *OutputSchema) {
				assert.Equal(t, TypeObject, schema.Type)
				assert.Len(t, schema.Properties, 2)
				assert.Equal(t, TypeString, schema.Properties["name"].Type)
				assert.Equal(t, TypeInteger, schema.Properties["age"].Type)
				assert.Equal(t, []string{"name"}, schema.RequiredProperties)
				assert.False(t, *schema.AdditionalProperties)
			},
		},
		{
			name: "Schema without type",
			bridgeSchema: map[string]interface{}{
				"description": "No type field",
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			schema, err := adapter.ConvertSchemaFromBridge(tc.bridgeSchema)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, schema)
				tc.validate(t, schema)
			}
		})
	}
}

func TestBridgeAdapter_ParseAndValidate(t *testing.T) {
	ctx := context.Background()
	adapter := NewBridgeAdapter()

	testCases := []struct {
		name     string
		output   string
		schema   *OutputSchema
		validate func(t *testing.T, result *BridgeResult)
	}{
		{
			name:   "Valid JSON without schema",
			output: `{"name": "test", "value": 42}`,
			validate: func(t *testing.T, result *BridgeResult) {
				assert.True(t, result.Success)
				assert.Equal(t, "json", result.Format)
				assert.NotNil(t, result.Data)
				assert.Empty(t, result.Error)
			},
		},
		{
			name:   "Valid JSON with schema validation",
			output: `{"name": "test", "age": 30}`,
			schema: &OutputSchema{
				Type: TypeObject,
				Properties: map[string]*OutputSchema{
					"name": {Type: TypeString},
					"age":  {Type: TypeInteger},
				},
			},
			validate: func(t *testing.T, result *BridgeResult) {
				assert.True(t, result.Success)
				assert.NotNil(t, result.Validation)
				assert.True(t, result.Validation.Valid)
			},
		},
		{
			name:   "Invalid JSON with recovery",
			output: `The result is: {"status": "success", count: 10} done.`,
			validate: func(t *testing.T, result *BridgeResult) {
				assert.True(t, result.Success)
				assert.Equal(t, "json", result.Format)
				data, ok := result.Data.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "success", data["status"])
				assert.Equal(t, float64(10), data["count"])
			},
		},
		{
			name:   "Schema validation failure",
			output: `{"name": 123}`,
			schema: &OutputSchema{
				Type: TypeObject,
				Properties: map[string]*OutputSchema{
					"name": {Type: TypeString},
				},
			},
			validate: func(t *testing.T, result *BridgeResult) {
				assert.False(t, result.Success)
				assert.NotNil(t, result.Validation)
				assert.False(t, result.Validation.Valid)
				assert.NotEmpty(t, result.Validation.Errors)
			},
		},
		{
			name:   "YAML format detection",
			output: "name: test\nvalue: 42",
			validate: func(t *testing.T, result *BridgeResult) {
				assert.True(t, result.Success)
				assert.Equal(t, "yaml", result.Format)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := adapter.ParseAndValidate(ctx, tc.output, tc.schema)
			require.NoError(t, err)
			require.NotNil(t, result)
			tc.validate(t, result)
		})
	}
}

func TestBridgeAdapter_ConvertFormat(t *testing.T) {
	ctx := context.Background()
	adapter := NewBridgeAdapter()

	testData := map[string]interface{}{
		"name":  "test",
		"value": 42,
	}

	testCases := []struct {
		name       string
		data       interface{}
		fromFormat string
		toFormat   string
		validate   func(t *testing.T, result *BridgeResult)
	}{
		{
			name:       "JSON to YAML",
			data:       testData,
			fromFormat: "json",
			toFormat:   "yaml",
			validate: func(t *testing.T, result *BridgeResult) {
				assert.True(t, result.Success)
				assert.Equal(t, "yaml", result.Format)
				yamlStr, ok := result.Data.(string)
				require.True(t, ok)
				assert.Contains(t, yamlStr, "name: test")
				assert.Contains(t, yamlStr, "value: 42")
			},
		},
		{
			name:       "JSON to XML",
			data:       testData,
			fromFormat: "json",
			toFormat:   "xml",
			validate: func(t *testing.T, result *BridgeResult) {
				assert.True(t, result.Success)
				assert.Equal(t, "xml", result.Format)
				xmlStr, ok := result.Data.(string)
				require.True(t, ok)
				assert.Contains(t, xmlStr, "<name>test</name>")
				assert.Contains(t, xmlStr, "<value>42</value>")
			},
		},
		{
			name:       "Invalid format",
			data:       testData,
			fromFormat: "json",
			toFormat:   "invalid",
			validate: func(t *testing.T, result *BridgeResult) {
				assert.False(t, result.Success)
				assert.Contains(t, result.Error, "unknown format")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := adapter.ConvertFormat(ctx, tc.data, tc.fromFormat, tc.toFormat)
			require.NoError(t, err)
			require.NotNil(t, result)
			tc.validate(t, result)
		})
	}
}

func TestBridgeAdapter_FixOutput(t *testing.T) {
	ctx := context.Background()
	adapter := NewBridgeAdapter()

	testCases := []struct {
		name     string
		output   string
		hints    map[string]interface{}
		validate func(t *testing.T, result *BridgeResult)
	}{
		{
			name:   "Fix JSON with trailing comma",
			output: `{"name": "test", "value": 42,}`,
			hints: map[string]interface{}{
				"format": "json",
			},
			validate: func(t *testing.T, result *BridgeResult) {
				assert.True(t, result.Success)
				assert.Equal(t, "json", result.Format)
				data, ok := result.Data.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "test", data["name"])
			},
		},
		{
			name:   "Fix with schema guidance",
			output: `{"person": {"name": "John", "age": "30"}}`,
			hints: map[string]interface{}{
				"schema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"person": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"name": map[string]interface{}{"type": "string"},
								"age":  map[string]interface{}{"type": "integer"},
							},
						},
					},
				},
			},
			validate: func(t *testing.T, result *BridgeResult) {
				assert.True(t, result.Success)
				data, ok := result.Data.(map[string]interface{})
				require.True(t, ok)
				person, ok := data["person"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "John", person["name"])
				assert.Equal(t, "30", person["age"]) // String because no type conversion
			},
		},
		{
			name: "Auto-detect format",
			output: `name: test
value: 42`,
			hints: map[string]interface{}{},
			validate: func(t *testing.T, result *BridgeResult) {
				assert.True(t, result.Success)
				assert.Equal(t, "yaml", result.Format)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := adapter.FixOutput(ctx, tc.output, tc.hints)
			require.NoError(t, err)
			require.NotNil(t, result)
			tc.validate(t, result)
		})
	}
}

func TestBridgeAdapter_GetSupportedFormats(t *testing.T) {
	adapter := NewBridgeAdapter()
	formats := adapter.GetSupportedFormats()

	assert.Contains(t, formats, "json")
	assert.Contains(t, formats, "xml")
	assert.Contains(t, formats, "yaml")
	assert.Len(t, formats, 3)
}

func TestBridgeAdapter_GetParserInfo(t *testing.T) {
	adapter := NewBridgeAdapter()
	info := adapter.GetParserInfo()

	assert.NotNil(t, info["parsers"])
	assert.Equal(t, "json", info["defaultParser"])

	features, ok := info["features"].([]string)
	require.True(t, ok)
	assert.Contains(t, features, "markdown_extraction")
	assert.Contains(t, features, "error_recovery")
	assert.Contains(t, features, "schema_guided_parsing")
	assert.Contains(t, features, "format_auto_detection")
	assert.Contains(t, features, "common_issue_fixing")
}

func TestBridgeAdapter_ComplexScenarios(t *testing.T) {
	ctx := context.Background()
	adapter := NewBridgeAdapter()

	t.Run("Nested schema conversion", func(t *testing.T) {
		bridgeSchema := map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"user": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"profile": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"name": map[string]interface{}{
									"type": "string",
								},
								"tags": map[string]interface{}{
									"type": "array",
									"items": map[string]interface{}{
										"type": "string",
									},
								},
							},
						},
					},
				},
			},
		}

		schema, err := adapter.ConvertSchemaFromBridge(bridgeSchema)
		require.NoError(t, err)
		require.NotNil(t, schema)

		// Verify nested structure
		assert.Equal(t, TypeObject, schema.Type)
		userSchema := schema.Properties["user"]
		assert.NotNil(t, userSchema)
		assert.Equal(t, TypeObject, userSchema.Type)

		profileSchema := userSchema.Properties["profile"]
		assert.NotNil(t, profileSchema)
		assert.Equal(t, TypeObject, profileSchema.Type)

		tagsSchema := profileSchema.Properties["tags"]
		assert.NotNil(t, tagsSchema)
		assert.Equal(t, TypeArray, tagsSchema.Type)
		assert.Equal(t, TypeString, tagsSchema.Items.Type)
	})

	t.Run("End-to-end workflow", func(t *testing.T) {
		// Simulate LLM output with issues
		llmOutput := "Here's the JSON response you requested:\n\n```json\n{\n  \"status\": \"success\",\n  \"data\": {\n    \"users\": [\n      {\"name\": \"Alice\", \"age\": 30},\n      {\"name\": \"Bob\", \"age\": 25}\n    ]\n  },\n  \"timestamp\": \"2023-01-01T12:00:00Z\",\n}\n```"

		// Parse and validate
		result, err := adapter.ParseAndValidate(ctx, llmOutput, nil)
		require.NoError(t, err)
		assert.True(t, result.Success)

		// Convert to YAML
		yamlResult, err := adapter.ConvertFormat(ctx, result.Data, "json", "yaml")
		require.NoError(t, err)
		assert.True(t, yamlResult.Success)

		// Convert to XML
		xmlResult, err := adapter.ConvertFormat(ctx, result.Data, "json", "xml")
		require.NoError(t, err)
		assert.True(t, xmlResult.Success)
	})
}
