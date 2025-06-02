// ABOUTME: Example demonstrating the use of built-in data processing tools
// ABOUTME: Shows JSON, CSV, XML processing and data transformations

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/data"
)

func main() {
	ctx := context.Background()

	// List all data tools
	fmt.Println("=== Available Data Tools ===")
	dataTools := tools.Tools.ListByCategory("data")
	for _, entry := range dataTools {
		fmt.Printf("- %s: %s\n", entry.Metadata.Name, entry.Metadata.Description)
	}
	fmt.Println()

	// Example 1: JSON Processing
	fmt.Println("=== Example 1: JSON Processing ===")
	jsonTool := tools.MustGetTool("json_process")

	// Sample JSON data
	jsonData := `{
		"users": [
			{"id": 1, "name": "Alice", "age": 30, "city": "New York"},
			{"id": 2, "name": "Bob", "age": 25, "city": "London"},
			{"id": 3, "name": "Charlie", "age": 35, "city": "Paris"}
		],
		"metadata": {
			"version": "1.0",
			"last_updated": "2024-01-15"
		}
	}`

	// Parse and query with JSONPath
	queryResult, err := jsonTool.Execute(ctx, map[string]interface{}{
		"operation": "query",
		"data":      jsonData,
		"jsonpath":  "$.users[?(@.age > 25)].name",
	})
	if err != nil {
		log.Printf("Failed to query JSON: %v", err)
	} else {
		fmt.Printf("Users over 25: %+v\n", queryResult)
	}

	// Transform JSON - flatten
	flattenResult, err := jsonTool.Execute(ctx, map[string]interface{}{
		"operation": "transform",
		"data":      jsonData,
		"transform": "flatten",
	})
	if err != nil {
		log.Printf("Failed to flatten JSON: %v", err)
	} else {
		fmt.Printf("Flattened JSON: %+v\n\n", flattenResult)
	}

	// Example 2: CSV Processing
	fmt.Println("=== Example 2: CSV Processing ===")
	csvTool := tools.MustGetTool("csv_process")

	// Sample CSV data
	csvData := `name,age,department,salary
Alice,30,Engineering,75000
Bob,25,Marketing,55000
Charlie,35,Engineering,85000
Diana,28,Sales,60000
Eve,32,Marketing,65000`

	// Parse and filter CSV
	filterResult, err := csvTool.Execute(ctx, map[string]interface{}{
		"operation": "filter",
		"data":      csvData,
		"filters": []map[string]interface{}{
			{
				"column":   "department",
				"operator": "eq",
				"value":    "Engineering",
			},
		},
		"has_header": true,
	})
	if err != nil {
		log.Printf("Failed to filter CSV: %v", err)
	} else {
		fmt.Printf("Engineering employees:\n%+v\n", filterResult)
	}

	// Get statistics
	statsResult, err := csvTool.Execute(ctx, map[string]interface{}{
		"operation":  "transform",
		"data":       csvData,
		"transform":  "statistics",
		"columns":    []string{"salary"},
		"has_header": true,
	})
	if err != nil {
		log.Printf("Failed to get CSV statistics: %v", err)
	} else {
		fmt.Printf("Salary statistics: %+v\n\n", statsResult)
	}

	// Example 3: XML Processing
	fmt.Println("=== Example 3: XML Processing ===")
	xmlTool := tools.MustGetTool("xml_process")

	// Sample XML data
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<catalog>
	<book id="1">
		<title>Go Programming</title>
		<author>John Doe</author>
		<price currency="USD">39.99</price>
	</book>
	<book id="2">
		<title>Web Development</title>
		<author>Jane Smith</author>
		<price currency="EUR">34.99</price>
	</book>
</catalog>`

	// Query XML with XPath
	xpathResult, err := xmlTool.Execute(ctx, map[string]interface{}{
		"operation": "query",
		"data":      xmlData,
		"xpath":     "//book/title",
	})
	if err != nil {
		log.Printf("Failed to query XML: %v", err)
	} else {
		fmt.Printf("Book titles: %+v\n", xpathResult)
	}

	// Convert XML to JSON
	xmlToJsonResult, err := xmlTool.Execute(ctx, map[string]interface{}{
		"operation":          "to_json",
		"data":               xmlData,
		"include_attributes": true,
	})
	if err != nil {
		log.Printf("Failed to convert XML to JSON: %v", err)
	} else {
		// Pretty print the JSON result
		if jsonBytes, err := json.MarshalIndent(xmlToJsonResult, "", "  "); err == nil {
			fmt.Printf("XML as JSON:\n%s\n\n", string(jsonBytes))
		}
	}

	// Example 4: Data Transformations
	fmt.Println("=== Example 4: Data Transformations ===")
	transformTool := tools.MustGetTool("data_transform")

	// Sample data for transformations
	transformData := []map[string]interface{}{
		{"name": "Alice", "score": 85, "grade": "B"},
		{"name": "Bob", "score": 92, "grade": "A"},
		{"name": "Charlie", "score": 78, "grade": "C"},
		{"name": "Diana", "score": 95, "grade": "A"},
		{"name": "Eve", "score": 88, "grade": "B"},
	}

	// Filter high scores
	filterHighScores, err := transformTool.Execute(ctx, map[string]interface{}{
		"operation": "filter",
		"data":      transformData,
		"condition": map[string]interface{}{
			"field":    "score",
			"operator": "gt",
			"value":    85,
		},
	})
	if err != nil {
		log.Printf("Failed to filter data: %v", err)
	} else {
		fmt.Printf("High scores (>85): %+v\n", filterHighScores)
	}

	// Map to extract names
	mapNames, err := transformTool.Execute(ctx, map[string]interface{}{
		"operation": "map",
		"data":      transformData,
		"transform": "extract_field",
		"field":     "name",
	})
	if err != nil {
		log.Printf("Failed to map data: %v", err)
	} else {
		fmt.Printf("Student names: %+v\n", mapNames)
	}

	// Reduce to calculate average score
	avgScore, err := transformTool.Execute(ctx, map[string]interface{}{
		"operation": "reduce",
		"data":      transformData,
		"reducer":   "average",
		"field":     "score",
	})
	if err != nil {
		log.Printf("Failed to reduce data: %v", err)
	} else {
		fmt.Printf("Average score: %+v\n", avgScore)
	}

	// Group by grade
	groupByGrade, err := transformTool.Execute(ctx, map[string]interface{}{
		"operation": "group_by",
		"data":      transformData,
		"field":     "grade",
	})
	if err != nil {
		log.Printf("Failed to group data: %v", err)
	} else {
		fmt.Printf("Students by grade: %+v\n", groupByGrade)
	}

	// Sort by score descending
	sortByScore, err := transformTool.Execute(ctx, map[string]interface{}{
		"operation": "sort",
		"data":      transformData,
		"field":     "score",
		"order":     "desc",
	})
	if err != nil {
		log.Printf("Failed to sort data: %v", err)
	} else {
		fmt.Printf("Students by score (desc): %+v\n", sortByScore)
	}
}
