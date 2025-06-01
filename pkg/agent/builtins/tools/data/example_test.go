// ABOUTME: Example test demonstrating the data processing tools
// ABOUTME: Shows how to use JSONProcess, CSVProcess, XMLProcess, and DataTransform tools

package data_test

import (
	"context"
	"fmt"
	"log"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools/data"
)

func ExampleJSONProcess() {
	ctx := context.Background()
	tool := data.JSONProcess()

	// Parse and validate JSON
	input := data.JSONProcessInput{
		Data:      `{"name": "John", "age": 30, "city": "New York"}`,
		Operation: "parse",
	}

	output, err := tool.Execute(ctx, input)
	if err != nil {
		log.Fatal(err)
	}

	result := output.(*data.JSONProcessOutput)
	fmt.Printf("Parsed successfully. Type: %s\n", result.ResultType)
	// Output: Parsed successfully. Type: map[string]interface {}
}

func ExampleCSVProcess() {
	ctx := context.Background()
	tool := data.CSVProcess()

	// Convert CSV to JSON
	input := data.CSVProcessInput{
		Data: `name,age,city
John,30,New York
Jane,25,Boston`,
		Operation:  "to_json",
		HasHeaders: true,
	}

	output, err := tool.Execute(ctx, input)
	if err != nil {
		log.Fatal(err)
	}

	result := output.(*data.CSVProcessOutput)
	fmt.Printf("Converted %d rows to JSON\n", result.RowCount)
	// Output: Converted 2 rows to JSON
}

func ExampleDataTransform() {
	ctx := context.Background()
	tool := data.DataTransform()

	// Filter data based on condition
	input := data.DataTransformInput{
		Data: `[
			{"name": "John", "age": 30},
			{"name": "Jane", "age": 25},
			{"name": "Bob", "age": 35}
		]`,
		Operation: "filter",
		Field:     "age",
		Condition: "gte:30",
	}

	output, err := tool.Execute(ctx, input)
	if err != nil {
		log.Fatal(err)
	}

	result := output.(*data.DataTransformOutput)
	fmt.Printf("Filtered to %d items\n", result.ItemCount)
	// Output: Filtered to 2 items
}
