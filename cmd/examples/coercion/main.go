package main

import (
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/schema/validation"
)

// Event represents an event structure with various types that need coercion
type Event struct {
	ID          string    `json:"id" validate:"required" format:"uuid" description:"Unique event identifier"`
	Name        string    `json:"name" validate:"required" description:"Event name"`
	Description string    `json:"description" description:"Event description"`
	StartDate   time.Time `json:"startDate" validate:"required" format:"date-time" description:"Event start date and time"`
	EndDate     time.Time `json:"endDate" validate:"required" format:"date-time" description:"Event end date and time"`
	Location    string    `json:"location" description:"Event location"`
	URL         url.URL   `json:"url" format:"uri" description:"Event website URL"`
	Duration    string    `json:"duration" format:"duration" description:"Event duration"`
	MaxAttendees int      `json:"maxAttendees" validate:"min=1" description:"Maximum number of attendees"`
	Categories  []string  `json:"categories" description:"Event categories"`
	Tags        []string  `json:"tags" description:"Event tags"`
	IsFree      bool      `json:"isFree" description:"Whether the event is free to attend"`
	Organizer   Organizer `json:"organizer" validate:"required" description:"Event organizer"`
}

// Organizer represents the event organizer
type Organizer struct {
	Name  string `json:"name" validate:"required" description:"Organizer name"`
	Email string `json:"email" validate:"required" format:"email" description:"Organizer email"`
}

// createEventSchema manually creates a schema for the Event struct
func createEventSchema() *domain.Schema {
	return &domain.Schema{
		Type: "object",
		Properties: map[string]domain.Property{
			"id": {
				Type:        "string",
				Format:      "uuid",
				Description: "Unique event identifier",
			},
			"name": {
				Type:        "string",
				Description: "Event name",
			},
			"description": {
				Type:        "string",
				Description: "Event description",
			},
			"startDate": {
				Type:        "string",
				Format:      "date-time",
				Description: "Event start date and time",
			},
			"endDate": {
				Type:        "string",
				Format:      "date-time",
				Description: "Event end date and time",
			},
			"location": {
				Type:        "string",
				Description: "Event location",
			},
			"url": {
				Type:        "string",
				Format:      "uri",
				Description: "Event website URL",
			},
			"duration": {
				Type:        "string",
				Format:      "duration",
				Description: "Event duration",
			},
			"maxAttendees": {
				Type:        "integer",
				Minimum:     float64Ptr(1),
				Description: "Maximum number of attendees",
			},
			"categories": {
				Type:        "array",
				Description: "Event categories",
				Items: &domain.Property{
					Type: "string",
				},
			},
			"tags": {
				Type:        "array",
				Description: "Event tags",
				Items: &domain.Property{
					Type: "string",
				},
			},
			"isFree": {
				Type:        "boolean",
				Description: "Whether the event is free to attend",
			},
			"organizer": {
				Type:        "object",
				Description: "Event organizer",
				Properties: map[string]domain.Property{
					"name": {
						Type:        "string",
						Description: "Organizer name",
					},
					"email": {
						Type:        "string",
						Format:      "email",
						Description: "Organizer email",
					},
				},
				Required: []string{"name", "email"},
			},
		},
		Required: []string{"id", "name", "startDate", "endDate", "organizer"},
	}
}

// Helper function for creating float pointers
func float64Ptr(v float64) *float64 {
	return &v
}

func main() {
	// Create a schema for validation
	schema := createEventSchema()

	// Create a validator with coercion enabled
	validator := validation.NewValidator(validation.WithCoercion(true))

	fmt.Println("=== Type Coercion Example ===")
	fmt.Println("\nThis example demonstrates the advanced type coercion capabilities of the schema validator.")
	fmt.Print("It will validate JSON data against a schema and automatically coerce values to their expected types.\n\n")

	// Example 1: Basic validation without coercion
	fmt.Println("Example 1: Validating a well-formed event")
	eventJSON := `{
		"id": "f47ac10b-58cc-0372-8567-0e02b2c3d479",
		"name": "Go Conference 2023",
		"description": "Annual conference for Go developers",
		"startDate": "2023-09-15T09:00:00Z",
		"endDate": "2023-09-16T18:00:00Z",
		"location": "San Francisco, CA",
		"url": "https://goconf.example.com",
		"duration": "2 days",
		"maxAttendees": 500,
		"categories": ["programming", "conference", "go"],
		"tags": ["golang", "developer", "tech"],
		"isFree": false,
		"organizer": {
			"name": "Go Developer Community",
			"email": "organizers@goconf.example.com"
		}
	}`

	result, err := validator.Validate(schema, eventJSON)
	if err != nil {
		log.Fatalf("Validation error: %v", err)
	}

	fmt.Printf("Validation result: %v\n", result.Valid)
	if !result.Valid {
		fmt.Printf("Errors: %v\n", result.Errors)
	}

	// Example 2: Validation with type coercion
	fmt.Println("\nExample 2: Validation with automatic type coercion")
	eventWithTypeMismatch := `{
		"id": "f47ac10b-58cc-0372-8567-0e02b2c3d479",
		"name": "Go Conference 2023",
		"description": "Annual conference for Go developers",
		"startDate": "2023-09-15",
		"endDate": 1694883600,
		"location": "San Francisco, CA",
		"url": "goconf.example.com",
		"duration": "1:30:00",
		"maxAttendees": "500",
		"categories": "programming, conference, go",
		"tags": ["golang", "developer", "tech"],
		"isFree": "false",
		"organizer": {
			"name": "Go Developer Community",
			"email": "organizers@goconf.example.com"
		}
	}`

	result, err = validator.Validate(schema, eventWithTypeMismatch)
	if err != nil {
		log.Fatalf("Validation error: %v", err)
	}

	fmt.Printf("Validation result: %v\n", result.Valid)
	if !result.Valid {
		fmt.Printf("Errors: %v\n", result.Errors)
	}

	// Example 3: Demonstrating specific coercion types
	fmt.Println("\nExample 3: Demonstrating specific coercion types")

	// UUID coercion
	fmt.Println("\nUUID Coercion:")
	uuidStr := "f47ac10b-58cc-0372-8567-0e02b2c3d479"
	uuidVal, ok := validation.CoerceToUUID(uuidStr)
	fmt.Printf("String '%s' to UUID: %v (success: %v)\n", uuidStr, uuidVal, ok)

	// Date coercion
	fmt.Println("\nDate Coercion:")
	dateStr := "2023-09-15"
	dateVal, ok := validation.CoerceToDate(dateStr)
	fmt.Printf("String '%s' to Date: %v (success: %v)\n", dateStr, dateVal.Format(time.RFC3339), ok)

	dateInt := int64(1694883600)
	dateVal, ok = validation.CoerceToDate(dateInt)
	fmt.Printf("Int64 '%d' to Date: %v (success: %v)\n", dateInt, dateVal.Format(time.RFC3339), ok)

	// Email coercion
	fmt.Println("\nEmail Coercion:")
	emailStr := "John Doe <john.doe@example.com>"
	emailVal, ok := validation.CoerceToEmail(emailStr)
	fmt.Printf("String '%s' to Email: %v (success: %v)\n", emailStr, emailVal, ok)

	// URL coercion
	fmt.Println("\nURL Coercion:")
	urlStr := "goconf.example.com"
	urlVal, ok := validation.CoerceToURL(urlStr)
	fmt.Printf("String '%s' to URL: %v (success: %v)\n", urlStr, urlVal, ok)

	// Duration coercion
	fmt.Println("\nDuration Coercion:")
	durationStr := "1h30m"
	durationVal, ok := validation.CoerceToDuration(durationStr)
	fmt.Printf("String '%s' to Duration: %v (success: %v)\n", durationStr, durationVal, ok)

	durationStr = "1:30:00"
	durationVal, ok = validation.CoerceToDuration(durationStr)
	fmt.Printf("String '%s' to Duration: %v (success: %v)\n", durationStr, durationVal, ok)

	durationStr = "2 days"
	durationVal, ok = validation.CoerceToDuration(durationStr)
	fmt.Printf("String '%s' to Duration: %v (success: %v)\n", durationStr, durationVal, ok)

	// Array coercion
	fmt.Println("\nArray Coercion:")
	arrayStr := "programming, conference, go"
	arrayVal, ok := validation.CoerceToArray(arrayStr)
	fmt.Printf("String '%s' to Array: %v (success: %v)\n", arrayStr, arrayVal, ok)

	// Object coercion
	fmt.Println("\nObject Coercion:")
	objectStr := `{"name": "Go Developer Community", "email": "organizers@goconf.example.com"}`
	objectVal, ok := validation.CoerceToObject(objectStr)
	fmt.Printf("JSON String to Object: %v (success: %v)\n", objectVal, ok)

	// Example 4: Note on coercing values when parsing into a struct
	fmt.Println("\nExample 4: Note on struct unmarshaling vs schema validation")
	fmt.Println("The standard Go JSON unmarshaling doesn't have the same coercion capabilities")
	fmt.Println("as our schema validator. For this reason, JSON that passes validation with")
	fmt.Println("coercion might still fail when unmarshaling into a Go struct.")
	fmt.Println("\nTo handle this situation, you could:")
	fmt.Println("1. Use the validated and coerced JSON data directly (as map[string]interface{})")
	fmt.Println("2. Create custom UnmarshalJSON methods for your struct types")
	fmt.Println("3. Use a pre-processing step to clean up the data before unmarshaling")
}