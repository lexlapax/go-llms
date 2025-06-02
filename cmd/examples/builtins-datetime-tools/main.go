// ABOUTME: Example demonstrating the use of built-in datetime tools
// ABOUTME: Shows various date/time operations like parsing, formatting, and calculations

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/datetime"
)

func main() {
	ctx := context.Background()

	// List all datetime tools
	fmt.Println("=== Available DateTime Tools ===")
	dateTimeTools := tools.Tools.ListByCategory("datetime")
	for _, entry := range dateTimeTools {
		fmt.Printf("- %s: %s\n", entry.Metadata.Name, entry.Metadata.Description)
	}
	fmt.Println()

	// Example 1: Get current date/time
	fmt.Println("=== Example 1: Current Date/Time ===")
	nowTool := tools.MustGetTool("datetime_now")
	result, err := nowTool.Execute(ctx, map[string]interface{}{
		"timezone":           "America/New_York",
		"include_components": true,
		"include_week_info":  true,
		"include_timestamps": true,
		"custom_format":      "Monday, January 2, 2006 at 3:04 PM MST",
	})
	if err != nil {
		log.Fatalf("Failed to get current time: %v", err)
	}
	fmt.Printf("Current time result: %+v\n\n", result)

	// Example 2: Parse a date
	fmt.Println("=== Example 2: Parse Date ===")
	parseTool := tools.MustGetTool("datetime_parse")
	parseResult, err := parseTool.Execute(ctx, map[string]interface{}{
		"date_string": "next Monday at 3pm",
		"timezone":    "America/Los_Angeles",
	})
	if err != nil {
		log.Fatalf("Failed to parse date: %v", err)
	}
	fmt.Printf("Parsed date: %+v\n\n", parseResult)

	// Example 3: Calculate date differences
	fmt.Println("=== Example 3: Date Calculations ===")
	calcTool := tools.MustGetTool("datetime_calculate")

	// Add business days
	calcResult, err := calcTool.Execute(ctx, map[string]interface{}{
		"operation":     "add_business_days",
		"date_time":     "2024-01-15T10:00:00Z",
		"business_days": 5,
	})
	if err != nil {
		log.Fatalf("Failed to calculate date: %v", err)
	}
	fmt.Printf("Add 5 business days: %+v\n", calcResult)

	// Calculate age
	ageResult, err := calcTool.Execute(ctx, map[string]interface{}{
		"operation": "age",
		"date_time": "1990-05-15",
	})
	if err != nil {
		log.Fatalf("Failed to calculate age: %v", err)
	}
	fmt.Printf("Age calculation: %+v\n\n", ageResult)

	// Example 4: Format dates
	fmt.Println("=== Example 4: Format Dates ===")
	formatTool := tools.MustGetTool("datetime_format")
	formatResult, err := formatTool.Execute(ctx, map[string]interface{}{
		"datetime":    "2024-12-25T10:30:00Z",
		"format_type": "multiple",
		"formats":     []string{"RFC3339", "Kitchen", "Monday, January 2, 2006"},
		"locale":      "es", // Spanish
	})
	if err != nil {
		log.Fatalf("Failed to format date: %v", err)
	}
	fmt.Printf("Formatted dates: %+v\n\n", formatResult)

	// Example 5: Timezone conversion
	fmt.Println("=== Example 5: Timezone Conversion ===")
	convertTool := tools.MustGetTool("datetime_convert")
	convertResult, err := convertTool.Execute(ctx, map[string]interface{}{
		"operation":     "timezone",
		"datetime":      "2024-07-15T15:00:00Z",
		"from_timezone": "UTC",
		"to_timezone":   "Asia/Tokyo",
		"include_dst":   true,
	})
	if err != nil {
		log.Fatalf("Failed to convert timezone: %v", err)
	}
	fmt.Printf("Timezone conversion: %+v\n\n", convertResult)

	// Example 6: Get date information
	fmt.Println("=== Example 6: Date Information ===")
	infoTool := tools.MustGetTool("datetime_info")
	infoResult, err := infoTool.Execute(ctx, map[string]interface{}{
		"date_time": "2024-02-29", // Leap year date
	})
	if err != nil {
		log.Fatalf("Failed to get date info: %v", err)
	}
	fmt.Printf("Date information: %+v\n\n", infoResult)

	// Example 7: Compare dates
	fmt.Println("=== Example 7: Compare Dates ===")
	compareTool := tools.MustGetTool("datetime_compare")

	// Compare two dates
	compareResult, err := compareTool.Execute(ctx, map[string]interface{}{
		"operation": "compare",
		"date1":     "2024-01-15T10:00:00Z",
		"date2":     "2024-01-20T15:30:00Z",
	})
	if err != nil {
		log.Fatalf("Failed to compare dates: %v", err)
	}
	fmt.Printf("Date comparison: %+v\n", compareResult)

	// Sort multiple dates
	sortResult, err := compareTool.Execute(ctx, map[string]interface{}{
		"operation": "sort",
		"dates": []string{
			"2024-03-15",
			"2024-01-10",
			"2024-06-20",
			"2024-02-28",
		},
		"sort_order": "desc",
	})
	if err != nil {
		log.Fatalf("Failed to sort dates: %v", err)
	}
	fmt.Printf("Sorted dates: %+v\n", sortResult)
}
