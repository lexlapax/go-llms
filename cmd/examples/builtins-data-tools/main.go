// ABOUTME: Example demonstrating the use of built-in data processing tools
// ABOUTME: Shows JSON, CSV, XML processing and data transformations

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	datatools "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/data"
	agentDomain "github.com/lexlapax/go-llms/pkg/agent/domain"
)

// Helper types for creating a minimal ToolContext for standalone tool execution

// minimalStateReader implements StateReader interface with empty state
type minimalStateReader struct {
	state *agentDomain.State
}

func (m *minimalStateReader) Get(key string) (interface{}, bool) {
	return m.state.Get(key)
}

func (m *minimalStateReader) Values() map[string]interface{} {
	return m.state.Values()
}

func (m *minimalStateReader) GetArtifact(id string) (*agentDomain.Artifact, bool) {
	return m.state.GetArtifact(id)
}

func (m *minimalStateReader) Artifacts() map[string]*agentDomain.Artifact {
	return m.state.Artifacts()
}

func (m *minimalStateReader) Messages() []agentDomain.Message {
	return m.state.Messages()
}

func (m *minimalStateReader) GetMetadata(key string) (interface{}, bool) {
	return m.state.GetMetadata(key)
}

func (m *minimalStateReader) Has(key string) bool {
	return m.state.Has(key)
}

func (m *minimalStateReader) Keys() []string {
	values := m.state.Values()
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	return keys
}

// minimalEventEmitter implements EventEmitter interface with no-op methods
type minimalEventEmitter struct{}

func (m *minimalEventEmitter) Emit(eventType agentDomain.EventType, data interface{}) {}
func (m *minimalEventEmitter) EmitProgress(current, total int, message string)        {}
func (m *minimalEventEmitter) EmitMessage(message string)                             {}
func (m *minimalEventEmitter) EmitError(err error)                                    {}
func (m *minimalEventEmitter) EmitCustom(eventName string, data interface{})          {}

// createToolContext creates a minimal ToolContext for standalone tool execution
func createToolContext(ctx context.Context) *agentDomain.ToolContext {
	state := agentDomain.NewState()
	stateReader := &minimalStateReader{state: state}

	toolCtx := &agentDomain.ToolContext{
		Context:   ctx,
		State:     stateReader,
		RunID:     "standalone-execution",
		Retry:     0,
		StartTime: time.Now(),
		Events:    &minimalEventEmitter{},
		Agent: agentDomain.AgentInfo{
			ID:          "standalone",
			Name:        "standalone-tool-executor",
			Description: "Minimal agent for standalone tool execution",
			Type:        agentDomain.AgentTypeLLM,
			Metadata:    make(map[string]interface{}),
		},
	}

	return toolCtx
}

func main() {
	ctx := context.Background()
	toolCtx := createToolContext(ctx)

	// List all data tools
	fmt.Println("=== Available Data Tools ===")
	fmt.Println()
	dataTools := tools.Tools.ListByCategory("data")
	fmt.Printf("Total data tools: %d\n", len(dataTools))
	for _, entry := range dataTools {
		fmt.Printf("• %s: %s\n", entry.Metadata.Name, entry.Metadata.Description)
	}
	fmt.Println()

	// Example 1: JSON Processing
	fmt.Println("=== Example 1: JSON Processing (json_process) ===")
	fmt.Println()
	jsonTool := tools.MustGetTool("json_process")

	// Sample JSON data
	jsonData := `{
		"users": [
			{"id": 1, "name": "Alice", "age": 30, "city": "New York", "active": true},
			{"id": 2, "name": "Bob", "age": 25, "city": "London", "active": false},
			{"id": 3, "name": "Charlie", "age": 35, "city": "Paris", "active": true},
			{"id": 4, "name": "Diana", "age": 28, "city": "Tokyo", "active": true}
		],
		"metadata": {
			"version": "1.0",
			"last_updated": "2024-01-15",
			"total_users": 4
		}
	}`

	// 1. Parse JSON
	fmt.Println("1. Parse and validate JSON:")
	parseResult, err := jsonTool.Execute(toolCtx, map[string]interface{}{
		"operation": "parse",
		"data":      jsonData,
	})
	if err != nil {
		log.Printf("Failed to parse JSON: %v", err)
	} else if output, ok := parseResult.(*datatools.JSONProcessOutput); ok {
		fmt.Printf("   ✓ Valid JSON parsed successfully\n")
		fmt.Printf("   Result type: %s\n", output.ResultType)
	}

	// 2. Query with JSONPath - simple queries
	fmt.Println("\n2. JSONPath query - get first user:")
	queryResult, err := jsonTool.Execute(toolCtx, map[string]interface{}{
		"operation": "query",
		"data":      jsonData,
		"jsonpath":  "$.users[0]",
	})
	if err != nil {
		log.Printf("Failed to query JSON: %v", err)
	} else if output, ok := queryResult.(*datatools.JSONProcessOutput); ok {
		if user, ok := output.Result.(map[string]interface{}); ok {
			fmt.Printf("   First user: %s (age: %.0f, city: %s)\n",
				user["name"], user["age"], user["city"])
		}
	}

	// 3. Multiple JSONPath queries
	fmt.Println("\n3. Simple JSONPath queries:")
	queries := map[string]string{
		"First user name":  "$.users[0].name",
		"Second user city": "$.users[1].city",
		"Third user age":   "$.users[2].age",
		"Version":          "$.metadata.version",
		"User count":       "$.metadata.total_users",
		"Last updated":     "$.metadata.last_updated",
	}
	for desc, path := range queries {
		result, _ := jsonTool.Execute(toolCtx, map[string]interface{}{
			"operation": "query",
			"data":      jsonData,
			"jsonpath":  path,
		})
		if output, ok := result.(*datatools.JSONProcessOutput); ok {
			fmt.Printf("   %s: %v\n", desc, output.Result)
		}
	}

	// 4. Transform JSON - flatten
	fmt.Println("\n4. Transform JSON - flatten:")
	flattenResult, err := jsonTool.Execute(toolCtx, map[string]interface{}{
		"operation": "transform",
		"data":      jsonData,
		"transform": "flatten",
	})
	if err != nil {
		log.Printf("Failed to flatten JSON: %v", err)
	} else if output, ok := flattenResult.(*datatools.JSONProcessOutput); ok {
		// Pretty print flattened result
		if jsonBytes, err := json.MarshalIndent(output.Result, "   ", "  "); err == nil {
			fmt.Printf("   Flattened structure:\n   %s\n", string(jsonBytes))
		}
	}

	// 5. Query to get users array
	fmt.Println("\n5. Extract user array using JSONPath:")
	extractResult, err := jsonTool.Execute(toolCtx, map[string]interface{}{
		"operation": "query",
		"data":      jsonData,
		"jsonpath":  "$.users",
	})
	if err != nil {
		log.Printf("Failed to extract users: %v", err)
	} else if output, ok := extractResult.(*datatools.JSONProcessOutput); ok {
		if users, ok := output.Result.([]interface{}); ok {
			fmt.Printf("   Extracted %d users\n", len(users))
			// Show user names
			fmt.Print("   User names: ")
			for i, user := range users {
				if u, ok := user.(map[string]interface{}); ok {
					if i > 0 {
						fmt.Print(", ")
					}
					fmt.Printf("%v", u["name"])
				}
			}
			fmt.Println()
		}
	}
	fmt.Println()

	// Example 2: CSV Processing
	fmt.Println("=== Example 2: CSV Processing (csv_process) ===")
	fmt.Println()
	csvTool := tools.MustGetTool("csv_process")

	// Sample CSV data with more fields
	csvData := `name,age,department,salary,years_experience,performance_rating
Alice,30,Engineering,75000,5,4.5
Bob,25,Marketing,55000,2,3.8
Charlie,35,Engineering,85000,8,4.7
Diana,28,Sales,60000,4,4.2
Eve,32,Marketing,65000,6,4.0
Frank,29,Engineering,70000,4,3.9
Grace,31,Sales,62000,5,4.3`

	// 1. Parse CSV
	fmt.Println("1. Parse CSV with headers:")
	parseCSV, err := csvTool.Execute(toolCtx, map[string]interface{}{
		"operation":   "parse",
		"data":        csvData,
		"has_headers": true, // Fixed parameter name (plural)
	})
	if err != nil {
		log.Printf("Failed to parse CSV: %v", err)
	} else if output, ok := parseCSV.(*datatools.CSVProcessOutput); ok {
		fmt.Printf("   ✓ Parsed %d rows\n", output.RowCount)
		if len(output.Columns) > 0 {
			fmt.Printf("   Columns: %v\n", output.Columns)
		}
	}

	// 2. Filter CSV - Engineering department (fixed parameter name)
	fmt.Println("\n2. Filter Engineering employees:")
	filterResult, err := csvTool.Execute(toolCtx, map[string]interface{}{
		"operation":        "filter",
		"data":             csvData,
		"filter_condition": "department:eq:Engineering", // Fixed parameter name
		"has_headers":      true,                        // Fixed parameter name (plural)
	})
	if err != nil {
		log.Printf("Failed to filter CSV: %v", err)
	} else if output, ok := filterResult.(*datatools.CSVProcessOutput); ok {
		fmt.Printf("   Found %d Engineering employees:\n", output.RowCount)
		// Print the filtered results as a table
		if rows, ok := output.Result.([][]string); ok {
			for i, row := range rows {
				if i == 0 {
					fmt.Printf("   %-10s %-4s %-12s %-8s\n", row[0], row[1], row[2], row[3])
					fmt.Println("   " + string(make([]byte, 40)))
				} else {
					fmt.Printf("   %-10s %-4s %-12s $%-7s\n", row[0], row[1], row[2], row[3])
				}
			}
		}
	}

	// 3. Filter high salaries (fixed parameter name)
	fmt.Println("\n3. Filter high salaries (> 65000):")
	highSalaryFilter, err := csvTool.Execute(toolCtx, map[string]interface{}{
		"operation":        "filter",
		"data":             csvData,
		"filter_condition": "salary:gt:65000", // Fixed parameter name
		"has_headers":      true,              // Fixed parameter name (plural)
	})
	if err != nil {
		log.Printf("Failed to filter CSV: %v", err)
	} else if output, ok := highSalaryFilter.(*datatools.CSVProcessOutput); ok {
		fmt.Printf("   Found %d employees with salary > $65,000\n", output.RowCount)
		if rows, ok := output.Result.([][]string); ok && len(rows) > 1 {
			// Skip header, show first few results
			for i := 1; i <= 3 && i < len(rows); i++ {
				row := rows[i]
				if len(row) >= 4 {
					fmt.Printf("   • %s: $%s\n", row[0], row[3])
				}
			}
		}
	}

	// 4. Convert CSV to JSON (fixed operation)
	fmt.Println("\n4. Convert CSV to JSON:")
	csvToJson, err := csvTool.Execute(toolCtx, map[string]interface{}{
		"operation":   "to_json", // Fixed: use to_json operation directly
		"data":        csvData,
		"has_headers": true, // Fixed parameter name (plural)
	})
	if err != nil {
		log.Printf("Failed to convert CSV to JSON: %v", err)
	} else if output, ok := csvToJson.(*datatools.CSVProcessOutput); ok {
		// The result is already a JSON string
		if jsonStr, ok := output.Result.(string); ok {
			fmt.Println("   CSV converted to JSON:")
			// Pretty print just the first 500 characters
			if len(jsonStr) > 500 {
				fmt.Printf("   %s\n   ... (truncated)\n", jsonStr[:500])
			} else {
				fmt.Printf("   %s\n", jsonStr)
			}
		}
	}

	// 5. Get statistics
	fmt.Println("\n5. Calculate statistics for numeric columns:")
	statsResult, err := csvTool.Execute(toolCtx, map[string]interface{}{
		"operation":   "transform",
		"data":        csvData,
		"transform":   "statistics",
		"has_headers": true, // Fixed parameter name (plural)
		"params": map[string]interface{}{ // Added params wrapper
			"columns": []string{"salary", "years_experience", "performance_rating"},
		},
	})
	if err != nil {
		log.Printf("Failed to get CSV statistics: %v", err)
	} else if output, ok := statsResult.(*datatools.CSVProcessOutput); ok {
		fmt.Println("   Statistics:")
		// Pretty print the statistics result
		if jsonBytes, err := json.MarshalIndent(output.Result, "   ", "  "); err == nil {
			fmt.Printf("   %s\n", string(jsonBytes))
		}
	}
	fmt.Println()

	// Example 3: XML Processing
	fmt.Println("=== Example 3: XML Processing (xml_process) ===")
	fmt.Println()
	xmlTool := tools.MustGetTool("xml_process")

	// Sample XML data - more complex structure
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<catalog>
	<metadata>
		<created>2024-01-15</created>
		<version>2.0</version>
	</metadata>
	<book id="1" category="programming">
		<title>Go Programming</title>
		<author>John Doe</author>
		<price currency="USD">39.99</price>
		<rating>4.5</rating>
		<tags>
			<tag>golang</tag>
			<tag>backend</tag>
			<tag>concurrency</tag>
		</tags>
	</book>
	<book id="2" category="web">
		<title>Web Development with Go</title>
		<author>Jane Smith</author>
		<price currency="EUR">34.99</price>
		<rating>4.2</rating>
		<tags>
			<tag>web</tag>
			<tag>http</tag>
			<tag>api</tag>
		</tags>
	</book>
	<book id="3" category="programming">
		<title>Advanced Go Patterns</title>
		<author>Bob Johnson</author>
		<price currency="USD">44.99</price>
		<rating>4.8</rating>
		<tags>
			<tag>patterns</tag>
			<tag>advanced</tag>
		</tags>
	</book>
</catalog>`

	// 1. Parse XML
	fmt.Println("1. Parse and validate XML:")
	parseXML, err := xmlTool.Execute(toolCtx, map[string]interface{}{
		"operation": "parse",
		"data":      xmlData,
	})
	if err != nil {
		log.Printf("Failed to parse XML: %v", err)
	} else if output, ok := parseXML.(*datatools.XMLProcessOutput); ok {
		fmt.Printf("   ✓ Valid XML parsed successfully\n")
		if output.RootElement != "" {
			fmt.Printf("   Root element: %s\n", output.RootElement)
		}
	}

	// 2. XPath queries - use proper XPath syntax
	fmt.Println("\n2. Simplified XPath queries:")
	xpathQueries := map[string]string{
		"Version":           "metadata/version", // Fixed: XPath syntax
		"Created date":      "metadata/created", // Fixed: XPath syntax
		"First book title":  "book/title",       // Fixed: Gets first book's title
		"First book author": "book/author",      // Fixed: Gets first book's author
		"All book titles":   "book/title",       // This will get all titles
		"Book count":        "book",             // This will show all books
	}

	for desc, xpath := range xpathQueries {
		result, err := xmlTool.Execute(toolCtx, map[string]interface{}{
			"operation": "query",
			"data":      xmlData,
			"xpath":     xpath,
		})
		if err != nil {
			log.Printf("   Failed to query '%s': %v", desc, err)
		} else if output, ok := result.(*datatools.XMLProcessOutput); ok {
			fmt.Printf("   %s: %v\n", desc, output.Result)
		}
	}

	// 3. Convert XML to JSON
	fmt.Println("\n3. Convert XML to JSON:")
	xmlToJsonResult, err := xmlTool.Execute(toolCtx, map[string]interface{}{
		"operation":          "to_json",
		"data":               xmlData,
		"include_attributes": true,
	})
	if err != nil {
		log.Printf("Failed to convert XML to JSON: %v", err)
	} else if output, ok := xmlToJsonResult.(*datatools.XMLProcessOutput); ok {
		// Pretty print a portion of the JSON
		fmt.Println("   XML converted to JSON structure:")
		if jsonBytes, err := json.MarshalIndent(output.Result, "   ", "  "); err == nil {
			// Truncate if too long
			jsonStr := string(jsonBytes)
			if len(jsonStr) > 500 {
				jsonStr = jsonStr[:500] + "\n   ... (truncated)"
			}
			fmt.Printf("   %s\n", jsonStr)
		}
	}
	fmt.Println()

	// Example 4: Data Transformations
	fmt.Println("=== Example 4: Data Transformations (data_transform) ===")
	fmt.Println()
	transformTool := tools.MustGetTool("data_transform")

	// Sample data for transformations - student records
	transformData := []map[string]interface{}{
		{"name": "Alice", "score": 85, "grade": "B", "subject": "Math", "semester": "Fall"},
		{"name": "Bob", "score": 92, "grade": "A", "subject": "Math", "semester": "Fall"},
		{"name": "Charlie", "score": 78, "grade": "C", "subject": "Science", "semester": "Fall"},
		{"name": "Diana", "score": 95, "grade": "A", "subject": "Math", "semester": "Spring"},
		{"name": "Eve", "score": 88, "grade": "B", "subject": "Science", "semester": "Spring"},
		{"name": "Frank", "score": 73, "grade": "C", "subject": "Math", "semester": "Spring"},
		{"name": "Grace", "score": 91, "grade": "A", "subject": "Science", "semester": "Fall"},
	}

	// Convert to JSON string for the data_transform tool
	transformDataJSON, err := json.Marshal(transformData)
	if err != nil {
		log.Fatalf("Failed to marshal transform data: %v", err)
	}

	// 1. Filter - high scores (fixed format)
	fmt.Println("1. Filter students with scores > 85:")
	filterHighScores, err := transformTool.Execute(toolCtx, map[string]interface{}{
		"operation": "filter",
		"data":      string(transformDataJSON),
		"field":     "score",
		"condition": "gt:85", // Fixed format: operator:value
	})
	if err != nil {
		log.Printf("Failed to filter data: %v", err)
	} else if output, ok := filterHighScores.(*datatools.DataTransformOutput); ok {
		fmt.Printf("   Found %d students with high scores\n", output.ItemCount)
		if students, ok := output.Result.([]interface{}); ok {
			for _, student := range students {
				if s, ok := student.(map[string]interface{}); ok {
					fmt.Printf("   • %s: %v (%s)\n", s["name"], s["score"], s["grade"])
				}
			}
		}
	}

	// 2. Map - extract names
	fmt.Println("\n2. Map - extract student names:")
	mapNames, err := transformTool.Execute(toolCtx, map[string]interface{}{
		"operation": "map",
		"data":      string(transformDataJSON),
		"map_type":  "extract_field", // Fixed parameter name
		"field":     "name",
	})
	if err != nil {
		log.Printf("Failed to map data: %v", err)
	} else if output, ok := mapNames.(*datatools.DataTransformOutput); ok {
		fmt.Printf("   Student names: %v\n", output.Result)
	}

	// 3. Reduce operations
	fmt.Println("\n3. Reduce operations on scores:")
	reduceOps := []struct {
		reducer string
		desc    string
	}{
		{"average", "Average score"},
		{"sum", "Total score"},
		{"min", "Minimum score"},
		{"max", "Maximum score"},
		{"count", "Student count"},
	}

	for _, op := range reduceOps {
		result, _ := transformTool.Execute(toolCtx, map[string]interface{}{
			"operation":   "reduce",
			"data":        string(transformDataJSON),
			"reduce_type": op.reducer, // Fixed parameter name
			"field":       "score",
		})
		if output, ok := result.(*datatools.DataTransformOutput); ok {
			fmt.Printf("   %s: %v\n", op.desc, output.Result)
		}
	}

	// 4. Group by grade
	fmt.Println("\n4. Group students by grade:")
	groupByGrade, err := transformTool.Execute(toolCtx, map[string]interface{}{
		"operation": "group_by",
		"data":      string(transformDataJSON),
		"field":     "grade",
	})
	if err != nil {
		log.Printf("Failed to group data: %v", err)
	} else if output, ok := groupByGrade.(*datatools.DataTransformOutput); ok {
		if groups, ok := output.Result.(map[string]interface{}); ok {
			for grade, students := range groups {
				if studentList, ok := students.([]interface{}); ok {
					fmt.Printf("   Grade %s: %d students\n", grade, len(studentList))
				}
			}
		}
	}

	// 5. Group by subject
	fmt.Println("\n5. Group by subject:")
	groupBySubject, _ := transformTool.Execute(toolCtx, map[string]interface{}{
		"operation": "group_by",
		"data":      string(transformDataJSON),
		"field":     "subject",
	})
	if output, ok := groupBySubject.(*datatools.DataTransformOutput); ok {
		if groups, ok := output.Result.(map[string]interface{}); ok {
			for subject, students := range groups {
				if studentList, ok := students.([]interface{}); ok {
					fmt.Printf("   %s: %d students\n", subject, len(studentList))
				}
			}
		}
	}

	// 6. Sort by score descending
	fmt.Println("\n6. Sort students by score (descending):")
	sortByScore, err := transformTool.Execute(toolCtx, map[string]interface{}{
		"operation":  "sort",
		"data":       string(transformDataJSON),
		"field":      "score",
		"sort_order": "desc", // Fixed parameter name
	})
	if err != nil {
		log.Printf("Failed to sort data: %v", err)
	} else if output, ok := sortByScore.(*datatools.DataTransformOutput); ok {
		fmt.Printf("   Top 3 students:\n")
		if students, ok := output.Result.([]interface{}); ok {
			for i := 0; i < 3 && i < len(students); i++ {
				if s, ok := students[i].(map[string]interface{}); ok {
					fmt.Printf("   %d. %s: %v points\n", i+1, s["name"], s["score"])
				}
			}
		}
	}

	// 7. Get unique values
	fmt.Println("\n7. Get unique grades:")
	uniqueGrades, _ := transformTool.Execute(toolCtx, map[string]interface{}{
		"operation": "unique",
		"data":      string(transformDataJSON),
		"field":     "grade",
	})
	if output, ok := uniqueGrades.(*datatools.DataTransformOutput); ok {
		fmt.Printf("   Unique grades: %v\n", output.Result)
	}

	// 8. Reverse the list
	fmt.Println("\n8. Reverse student list:")
	firstThree := transformData[:3]
	firstThreeJSON, _ := json.Marshal(firstThree)
	reverseList, _ := transformTool.Execute(toolCtx, map[string]interface{}{
		"operation": "reverse",
		"data":      string(firstThreeJSON),
	})
	if output, ok := reverseList.(*datatools.DataTransformOutput); ok {
		fmt.Println("   Original order → Reversed order:")
		if reversed, ok := output.Result.([]interface{}); ok {
			for i, student := range reversed {
				if s, ok := student.(map[string]interface{}); ok {
					origIndex := len(reversed) - 1 - i
					if origIndex < len(firstThree) {
						orig := firstThree[origIndex]
						fmt.Printf("   %s → %s\n", orig["name"], s["name"])
					}
				}
			}
		}
	}

	fmt.Println("\n=== Summary ===")
	fmt.Println("This example demonstrated:")
	fmt.Println("• JSON processing with JSONPath queries and transformations")
	fmt.Println("• CSV parsing, filtering, statistics, and format conversion")
	fmt.Println("• XML parsing with XPath queries and JSON conversion")
	fmt.Println("• Data transformations: filter, map, reduce, group, sort, unique, reverse")
	fmt.Println("\nAll tools process data without requiring LLM calls, making them fast and efficient.")
}
