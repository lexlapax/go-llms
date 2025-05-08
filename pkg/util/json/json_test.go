package json

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestJsonCompatibility(t *testing.T) {
	// Test data
	testCases := []struct {
		name   string
		value  interface{}
		target interface{}
	}{
		{
			name:   "simple map",
			value:  map[string]interface{}{"name": "John", "age": 30, "active": true},
			target: &map[string]interface{}{},
		},
		{
			name: "nested map",
			value: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "Jane",
					"address": map[string]interface{}{
						"street": "123 Main St",
						"city":   "Anytown",
					},
				},
				"status": "active",
			},
			target: &map[string]interface{}{},
		},
		{
			name:   "array",
			value:  []string{"apple", "banana", "cherry"},
			target: &[]string{},
		},
		{
			name: "struct",
			value: struct {
				Name    string   `json:"name"`
				Age     int      `json:"age"`
				Hobbies []string `json:"hobbies"`
			}{
				Name:    "Bob",
				Age:     25,
				Hobbies: []string{"reading", "coding"},
			},
			target: &struct {
				Name    string   `json:"name"`
				Age     int      `json:"age"`
				Hobbies []string `json:"hobbies"`
			}{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test Marshal
			stdJson, err1 := json.Marshal(tc.value)
			if err1 != nil {
				t.Fatalf("Standard json.Marshal error: %v", err1)
			}

			optimizedJson, err2 := Marshal(tc.value)
			if err2 != nil {
				t.Fatalf("Optimized json.Marshal error: %v", err2)
			}

			// Compare results
			if !bytes.Equal(stdJson, optimizedJson) {
				t.Errorf("Marshal results differ:\nStandard: %s\nOptimized: %s", stdJson, optimizedJson)
			}

			// Test Unmarshal
			target1 := cloneTarget(tc.target)
			target2 := cloneTarget(tc.target)

			err1 = json.Unmarshal(stdJson, target1)
			if err1 != nil {
				t.Fatalf("Standard json.Unmarshal error: %v", err1)
			}

			err2 = Unmarshal(optimizedJson, target2)
			if err2 != nil {
				t.Fatalf("Optimized json.Unmarshal error: %v", err2)
			}

			// Compare targets after unmarshaling
			// Converting both to JSON for easy comparison
			result1, _ := json.Marshal(target1)
			result2, _ := json.Marshal(target2)

			if !bytes.Equal(result1, result2) {
				t.Errorf("Unmarshal results differ:\nStandard: %s\nOptimized: %s", result1, result2)
			}
		})
	}
}

func TestBufferFunctions(t *testing.T) {
	// Test data
	testData := map[string]interface{}{
		"name": "John Doe",
		"age":  30,
		"address": map[string]interface{}{
			"street": "123 Main St",
			"city":   "Anytown",
		},
	}

	// Test MarshalWithBuffer
	buf := &bytes.Buffer{}
	err := MarshalWithBuffer(testData, buf)
	if err != nil {
		t.Fatalf("MarshalWithBuffer error: %v", err)
	}

	// Verify result
	result := buf.Bytes()
	// Compare with standard Marshal
	expected, _ := json.Marshal(testData)

	// The encoder adds a newline at the end, so we need to adjust for that
	if len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}

	if !bytes.Equal(result, expected) {
		t.Errorf("MarshalWithBuffer result doesn't match standard Marshal:\nExpected: %s\nGot: %s", expected, result)
	}

	// Test MarshalIndentWithBuffer
	buf = &bytes.Buffer{}
	err = MarshalIndentWithBuffer(testData, buf, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndentWithBuffer error: %v", err)
	}

	// Verify result
	result = buf.Bytes()

	// The encoder adds a newline at the end, so we need to adjust for that
	if len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}

	// Compare semantically by unmarshaling both into maps
	var resultMap, expectedMap map[string]interface{}

	// Unmarshal our result
	if err := json.Unmarshal(result, &resultMap); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	// Unmarshal expected
	expected, _ = json.MarshalIndent(testData, "", "  ")
	if err := json.Unmarshal(expected, &expectedMap); err != nil {
		t.Fatalf("Failed to unmarshal expected: %v", err)
	}

	// Compare the maps for semantic equality
	resultJson, _ := json.Marshal(resultMap)
	expectedJson, _ := json.Marshal(expectedMap)

	if !bytes.Equal(resultJson, expectedJson) {
		t.Errorf("MarshalIndentWithBuffer result semantically differs from standard MarshalIndent")
	}
}

// Helper function to clone a target interface{}
func cloneTarget(target interface{}) interface{} {
	switch t := target.(type) {
	case *map[string]interface{}:
		result := make(map[string]interface{})
		return &result
	case *[]string:
		result := make([]string, 0)
		return &result
	case *struct {
		Name    string   `json:"name"`
		Age     int      `json:"age"`
		Hobbies []string `json:"hobbies"`
	}:
		return &struct {
			Name    string   `json:"name"`
			Age     int      `json:"age"`
			Hobbies []string `json:"hobbies"`
		}{}
	default:
		return t
	}
}
