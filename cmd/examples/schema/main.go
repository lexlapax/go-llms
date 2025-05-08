package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/lexlapax/go-llms/pkg/schema/adapter/reflection"
	"github.com/lexlapax/go-llms/pkg/schema/domain"
)

// ProductCategory represents a category enum
type ProductCategory string

// Define product categories as enum values
const (
	CategoryElectronics ProductCategory = "electronics"
	CategoryClothing    ProductCategory = "clothing"
	CategoryBooks       ProductCategory = "books"
	CategoryHomeGoods   ProductCategory = "home_goods"
)

// Address represents a physical address
type Address struct {
	Street     string `json:"street" validate:"required" description:"Street address line"`
	City       string `json:"city" validate:"required" description:"City name"`
	State      string `json:"state" validate:"required" minLength:"2" maxLength:"2" description:"Two-letter state code"`
	PostalCode string `json:"postalCode" validate:"required" pattern:"^\\d{5}(-\\d{4})?$" description:"Postal/ZIP code"`
	Country    string `json:"country" validate:"required" description:"Country name"`
}

// Customer represents a customer in the system
type Customer struct {
	ID          string    `json:"id" validate:"required" description:"Unique customer identifier"`
	FirstName   string    `json:"firstName" validate:"required" description:"Customer's first name"`
	LastName    string    `json:"lastName" validate:"required" description:"Customer's last name"`
	Email       string    `json:"email" validate:"required,email" description:"Customer's email address"`
	PhoneNumber string    `json:"phoneNumber,omitempty" pattern:"^\\+?[0-9]{10,15}$" description:"Customer's phone number"`
	BirthDate   time.Time `json:"birthDate,omitempty" format:"date" description:"Customer's birth date"`
	Address     Address   `json:"address" validate:"required" description:"Customer's primary address"`
}

// ProductReview represents a review for a product
type ProductReview struct {
	Rating      int       `json:"rating" validate:"min=1,max=5" description:"Review rating from 1-5 stars"`
	Comment     string    `json:"comment" description:"Review text"`
	ReviewerID  string    `json:"reviewerId" description:"ID of the customer who wrote the review"`
	ReviewDate  time.Time `json:"reviewDate" format:"date-time" description:"Date when review was submitted"`
	Recommended bool      `json:"recommended" description:"Whether the reviewer recommends this product"`
}

// Product represents a product in the inventory
type Product struct {
	ID          string          `json:"id" validate:"required" description:"Unique product identifier"`
	Name        string          `json:"name" validate:"required" description:"Product name"`
	Description string          `json:"description" description:"Detailed product description"`
	Price       float64         `json:"price" validate:"min=0" description:"Product price in USD"`
	Category    ProductCategory `json:"category" validate:"required,oneof=electronics clothing books home_goods" description:"Product category"`
	InStock     bool            `json:"inStock" description:"Whether the product is in stock"`
	Quantity    int             `json:"quantity" validate:"min=0" description:"Number of units available"`
	Tags        []string        `json:"tags,omitempty" description:"Product tags for search and categorization"`
	Reviews     []ProductReview `json:"reviews,omitempty" description:"Product reviews from customers"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ProductID string  `json:"productId" validate:"required" description:"ID of the ordered product"`
	Quantity  int     `json:"quantity" validate:"min=1" description:"Number of units ordered"`
	UnitPrice float64 `json:"unitPrice" validate:"min=0" description:"Price per unit at time of order"`
}

// PaymentMethod represents payment method types
type PaymentMethod string

// Define payment methods as enum values
const (
	PaymentCreditCard PaymentMethod = "credit_card"
	PaymentDebit      PaymentMethod = "debit"
	PaymentPayPal     PaymentMethod = "paypal"
	PaymentBankWire   PaymentMethod = "bank_wire"
)

// OrderStatus represents the status of an order
type OrderStatus string

// Define order statuses as enum values
const (
	StatusPending     OrderStatus = "pending"
	StatusProcessing  OrderStatus = "processing"
	StatusShipped     OrderStatus = "shipped"
	StatusDelivered   OrderStatus = "delivered"
	StatusCancelled   OrderStatus = "cancelled"
)

// Order represents a customer order
type Order struct {
	ID              string       `json:"id" validate:"required" description:"Unique order identifier"`
	CustomerID      string       `json:"customerId" validate:"required" description:"ID of the customer who placed the order"`
	OrderDate       time.Time    `json:"orderDate" format:"date-time" description:"Date when order was placed"`
	Status          OrderStatus  `json:"status" validate:"required,oneof=pending processing shipped delivered cancelled" description:"Current order status"`
	Items           []OrderItem  `json:"items" validate:"required" description:"Items included in the order"`
	ShippingAddress Address      `json:"shippingAddress" validate:"required" description:"Shipping destination address"`
	BillingAddress  Address      `json:"billingAddress" validate:"required" description:"Billing address for payment"`
	PaymentMethod   PaymentMethod `json:"paymentMethod" validate:"required,oneof=credit_card debit paypal bank_wire" description:"Method of payment"`
	Total           float64      `json:"total" validate:"min=0" description:"Total order amount in USD"`
	Notes           string       `json:"notes,omitempty" description:"Additional order notes or instructions"`
}

// displaySchema prints a schema in a formatted JSON
func displaySchema(name string, schema *domain.Schema) {
	fmt.Printf("\n=== %s Schema ===\n", name)
	
	// Convert schema to JSON
	jsonBytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling schema: %v\n", err)
		return
	}
	
	// Print the JSON schema
	fmt.Println(string(jsonBytes))
	fmt.Println()
}

// showSchemaUsageWithLLM demonstrates how to use a generated schema with an LLM
func showSchemaUsageWithLLM(schema *domain.Schema) {
	fmt.Println("=== Using Generated Schema with LLM ===")
	
	// Example prompt that would generate a product
	prompt := "Generate a detailed product description for a new smartphone with high-end specifications. Include all required fields."
	
	// This is just a simulation - in a real scenario, this would call the LLM
	fmt.Println("Prompt:", prompt)
	fmt.Println("* In a real application, this would send the prompt to an LLM with the schema *")
	fmt.Println("* The LLM would generate a response conforming to the Product schema *")
	fmt.Println("* The processor would validate and parse the response into a Product struct *")
	
	// Example of how you would call the processor in a real application
	fmt.Println("\nExample code for processing with schema:")
	
	exampleCode := `
    // Create processor with LLM provider
    processor := structured.NewProcessor(llmProvider)
    
    // Process response with schema
    var product Product
    err := processor.ProcessTyped(ctx, prompt, &product)
    if err != nil {
        // Handle error
    }
    
    // Use the typed product object
    fmt.Printf("Created product: %s ($%.2f)\n", product.Name, product.Price)
`
	fmt.Println(exampleCode)
}

func main() {
	// Ensure output directory exists
	schemasDir := "schemas"
	if err := os.MkdirAll(schemasDir, 0755); err != nil {
		fmt.Printf("Error creating schemas directory: %v\n", err)
		return
	}

	// Generate and display schemas for various types
	fmt.Println("Generating JSON Schemas from Go structs...")
	
	// Address schema
	addressSchema, err := reflection.GenerateSchema(Address{})
	if err != nil {
		fmt.Printf("Error generating Address schema: %v\n", err)
		return
	}
	displaySchema("Address", addressSchema)
	writeSchemaToFile(schemasDir+"/address_schema.json", addressSchema)
	
	// Customer schema
	customerSchema, err := reflection.GenerateSchema(Customer{})
	if err != nil {
		fmt.Printf("Error generating Customer schema: %v\n", err)
		return
	}
	displaySchema("Customer", customerSchema)
	writeSchemaToFile(schemasDir+"/customer_schema.json", customerSchema)
	
	// Product schema
	productSchema, err := reflection.GenerateSchema(Product{})
	if err != nil {
		fmt.Printf("Error generating Product schema: %v\n", err)
		return
	}
	displaySchema("Product", productSchema)
	writeSchemaToFile(schemasDir+"/product_schema.json", productSchema)
	
	// Order schema
	orderSchema, err := reflection.GenerateSchema(Order{})
	if err != nil {
		fmt.Printf("Error generating Order schema: %v\n", err)
		return
	}
	displaySchema("Order", orderSchema)
	writeSchemaToFile(schemasDir+"/order_schema.json", orderSchema)
	
	// Demonstrate schema usage with LLM
	fmt.Println("\nDemonstrating schema usage with LLM...")
	showSchemaUsageWithLLM(productSchema)
	
	fmt.Println("\nAll schemas have been generated and saved to the 'schemas' directory.")
}

// writeSchemaToFile saves a schema to a JSON file
func writeSchemaToFile(filename string, schema *domain.Schema) {
	jsonBytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling schema for file %s: %v\n", filename, err)
		return
	}
	
	if err := os.WriteFile(filename, jsonBytes, 0644); err != nil {
		fmt.Printf("Error writing schema to file %s: %v\n", filename, err)
		return
	}
}