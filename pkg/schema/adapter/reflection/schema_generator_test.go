package reflection

import (
	"testing"
	"time"
)

// TestPerson is a struct for testing schema generation
type TestPerson struct {
	Name        string    `json:"name" validate:"required" description:"Person's full name"`
	Age         int       `json:"age" validate:"min=0,max=120" description:"Age in years"`
	Email       string    `json:"email" validate:"required,email" description:"Email address"`
	IsActive    bool      `json:"isActive" description:"Whether the person is active"`
	Tags        []string  `json:"tags" description:"Tags associated with the person"`
	BirthDate   time.Time `json:"birthDate" format:"date-time" description:"Date of birth"`
	PhoneNumber string    `json:"phoneNumber,omitempty" pattern:"^\\d{3}-\\d{3}-\\d{4}$" description:"Phone number in xxx-xxx-xxxx format"`
}

// TestAddress is a struct for testing nested schema generation
type TestAddress struct {
	Street  string `json:"street" validate:"required" description:"Street address"`
	City    string `json:"city" validate:"required" description:"City name"`
	State   string `json:"state" validate:"required" description:"State code" minLength:"2" maxLength:"2"`
	ZipCode string `json:"zipCode" validate:"required" pattern:"^\\d{5}(-\\d{4})?$" description:"ZIP code"`
}

// TestNestedPerson is a struct with nested fields for testing
type TestNestedPerson struct {
	Name    string       `json:"name" validate:"required" description:"Person's full name"`
	Address TestAddress  `json:"address" validate:"required" description:"Residential address"`
	Work    *TestAddress `json:"work,omitempty" description:"Work address"`
}

// TestEnumPerson tests enum handling
type TestEnumPerson struct {
	Name   string `json:"name" validate:"required"`
	Status string `json:"status" validate:"required,oneof=active inactive pending" description:"Account status"`
}

func TestGenerateSchema(t *testing.T) {
	t.Run("generate schema for simple struct", func(t *testing.T) {
		schema, err := GenerateSchema(TestPerson{})
		if err != nil {
			t.Fatalf("Error generating schema: %v", err)
		}

		// Check basic schema properties
		if schema.Type != "object" {
			t.Errorf("Expected type 'object', got '%s'", schema.Type)
		}

		// Check if required fields are properly set
		hasName := false
		hasEmail := false
		for _, req := range schema.Required {
			if req == "name" {
				hasName = true
			}
			if req == "email" {
				hasEmail = true
			}
		}
		if !hasName {
			t.Errorf("Expected 'name' to be required")
		}
		if !hasEmail {
			t.Errorf("Expected 'email' to be required")
		}

		// Check if properties have correct types
		if prop, ok := schema.Properties["name"]; ok {
			if prop.Type != "string" {
				t.Errorf("Expected 'name' to have type 'string', got '%s'", prop.Type)
			}
		} else {
			t.Errorf("Expected property 'name' not found")
		}

		if prop, ok := schema.Properties["age"]; ok {
			if prop.Type != "integer" {
				t.Errorf("Expected 'age' to have type 'integer', got '%s'", prop.Type)
			}
			// Check constraints
			if prop.Minimum == nil || *prop.Minimum != 0 {
				t.Errorf("Expected 'age' to have minimum 0")
			}
			if prop.Maximum == nil || *prop.Maximum != 120 {
				t.Errorf("Expected 'age' to have maximum 120")
			}
		} else {
			t.Errorf("Expected property 'age' not found")
		}

		if prop, ok := schema.Properties["tags"]; ok {
			if prop.Type != "array" {
				t.Errorf("Expected 'tags' to have type 'array', got '%s'", prop.Type)
			}
			if prop.Items == nil || prop.Items.Type != "string" {
				t.Errorf("Expected 'tags' items to have type 'string'")
			}
		} else {
			t.Errorf("Expected property 'tags' not found")
		}

		if prop, ok := schema.Properties["email"]; ok {
			if prop.Format != "email" {
				t.Errorf("Expected 'email' to have format 'email', got '%s'", prop.Format)
			}
		} else {
			t.Errorf("Expected property 'email' not found")
		}

		// Check for descriptions
		if prop, ok := schema.Properties["name"]; ok {
			if prop.Description != "Person's full name" {
				t.Errorf("Expected description 'Person's full name', got '%s'", prop.Description)
			}
		}
	})

	t.Run("generate schema for nested struct", func(t *testing.T) {
		schema, err := GenerateSchema(TestNestedPerson{})
		if err != nil {
			t.Fatalf("Error generating schema: %v", err)
		}

		// Check if nested object is properly defined
		if prop, ok := schema.Properties["address"]; ok {
			if prop.Type != "object" {
				t.Errorf("Expected 'address' to have type 'object', got '%s'", prop.Type)
			}

			// Check nested properties
			if nestedProp, ok := prop.Properties["street"]; ok {
				if nestedProp.Type != "string" {
					t.Errorf("Expected 'address.street' to have type 'string', got '%s'", nestedProp.Type)
				}
			} else {
				t.Errorf("Expected property 'address.street' not found")
			}
		} else {
			t.Errorf("Expected property 'address' not found")
		}

		// Check optional nested object
		if prop, ok := schema.Properties["work"]; ok {
			if prop.Type != "object" {
				t.Errorf("Expected 'work' to have type 'object', got '%s'", prop.Type)
			}
		} else {
			t.Errorf("Expected property 'work' not found")
		}
	})

	t.Run("generate schema with enum values", func(t *testing.T) {
		schema, err := GenerateSchema(TestEnumPerson{})
		if err != nil {
			t.Fatalf("Error generating schema: %v", err)
		}

		// Check enum values
		if prop, ok := schema.Properties["status"]; ok {
			if len(prop.Enum) != 3 {
				t.Errorf("Expected 'status' to have 3 enum values, got %d", len(prop.Enum))
			}
			hasActive := false
			hasInactive := false
			hasPending := false
			for _, e := range prop.Enum {
				switch e {
				case "active":
					hasActive = true
				case "inactive":
					hasInactive = true
				case "pending":
					hasPending = true
				}
			}
			if !hasActive || !hasInactive || !hasPending {
				t.Errorf("Missing expected enum values")
			}
		} else {
			t.Errorf("Expected property 'status' not found")
		}
	})
}
