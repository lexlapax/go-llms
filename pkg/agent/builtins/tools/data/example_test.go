// ABOUTME: Example test demonstrating the data processing tools
// ABOUTME: Shows how to use JSONProcess, CSVProcess, XMLProcess, and DataTransform tools

package data

import (
	"context"
	"fmt"
	"log"
)

func ExampleJSONProcess() {
	ctx := context.Background()
	toolCtx := createTestToolContext(ctx)
	tool := JSONProcess()

	// Parse and validate JSON
	input := JSONProcessInput{
		Data:      `{"name": "John", "age": 30, "city": "New York"}`,
		Operation: "parse",
	}

	output, err := tool.Execute(toolCtx, input)
	if err != nil {
		log.Fatal(err)
	}

	result := output.(*JSONProcessOutput)
	fmt.Printf("Parsed successfully. Type: %s\n", result.ResultType)
	// Output: Parsed successfully. Type: map[string]interface {}
}

func ExampleCSVProcess() {
	ctx := context.Background()
	toolCtx := createTestToolContext(ctx)
	tool := CSVProcess()

	// Convert CSV to JSON
	input := CSVProcessInput{
		Data: `name,age,city
John,30,New York
Jane,25,Boston`,
		Operation:  "to_json",
		HasHeaders: true,
	}

	output, err := tool.Execute(toolCtx, input)
	if err != nil {
		log.Fatal(err)
	}

	result := output.(*CSVProcessOutput)
	fmt.Printf("Converted %d rows to JSON\n", result.RowCount)
	// Output: Converted 2 rows to JSON
}

func ExampleDataTransform() {
	ctx := context.Background()
	toolCtx := createTestToolContext(ctx)
	tool := DataTransform()

	// Filter data based on condition
	input := DataTransformInput{
		Data: `[
			{"name": "John", "age": 30},
			{"name": "Jane", "age": 25},
			{"name": "Bob", "age": 35}
		]`,
		Operation: "filter",
		Field:     "age",
		Condition: "gte:30",
	}

	output, err := tool.Execute(toolCtx, input)
	if err != nil {
		log.Fatal(err)
	}

	result := output.(*DataTransformOutput)
	fmt.Printf("Filtered to %d items\n", result.ItemCount)
	// Output: Filtered to 2 items
}
