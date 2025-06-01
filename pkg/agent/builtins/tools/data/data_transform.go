// ABOUTME: DataTransform tool provides common data transformation operations
// ABOUTME: This tool enables agents to filter, map, reduce, and transform data without requiring LLM processing

package data

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	atools "github.com/lexlapax/go-llms/pkg/agent/tools"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// DataTransformInput represents the input for data transformation operations
type DataTransformInput struct {
	// The data to transform (as JSON string or array)
	Data string `json:"data" jsonschema:"title=Data,description=The data to transform as JSON string or array,required"`

	// Operation to perform: filter, map, reduce, sort, group_by, unique, reverse
	Operation string `json:"operation" jsonschema:"title=Operation,description=Operation to perform: filter map reduce sort group_by unique reverse,enum=filter,enum=map,enum=reduce,enum=sort,enum=group_by,enum=unique,enum=reverse,required"`

	// Field or expression for the operation
	Field string `json:"field,omitempty" jsonschema:"title=Field,description=Field name or path for the operation"`

	// Condition for filter operation
	Condition string `json:"condition,omitempty" jsonschema:"title=Condition,description=Condition for filter operation in format: operator:value"`

	// Mapping type for map operation
	MapType string `json:"map_type,omitempty" jsonschema:"title=Map Type,description=Type of mapping: extract_field to_upper to_lower to_number to_string,enum=extract_field,enum=to_upper,enum=to_lower,enum=to_number,enum=to_string"`

	// Reduce type for reduce operation
	ReduceType string `json:"reduce_type,omitempty" jsonschema:"title=Reduce Type,description=Type of reduction: sum count min max average concat,enum=sum,enum=count,enum=min,enum=max,enum=average,enum=concat"`

	// Sort order for sort operation
	SortOrder string `json:"sort_order,omitempty" jsonschema:"title=Sort Order,description=Sort order: asc or desc,enum=asc,enum=desc,default=asc"`
}

// DataTransformOutput represents the output of data transformation
type DataTransformOutput struct {
	// The transformed result
	Result interface{} `json:"result"`

	// Error message if any
	Error string `json:"error,omitempty"`

	// Number of items in result
	ItemCount int `json:"item_count,omitempty"`

	// Type of the result
	ResultType string `json:"result_type"`
}

// dataTransformParamSchema defines parameters for the DataTransform tool
var dataTransformParamSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"data": {
			Type:        "string",
			Description: "The data to transform as JSON string or array",
		},
		"operation": {
			Type:        "string",
			Description: "Operation to perform: filter, map, reduce, sort, group_by, unique, or reverse",
			Enum:        []string{"filter", "map", "reduce", "sort", "group_by", "unique", "reverse"},
		},
		"field": {
			Type:        "string",
			Description: "Field name or path for the operation",
		},
		"condition": {
			Type:        "string",
			Description: "Condition for filter operation in format: operator:value",
		},
		"map_type": {
			Type:        "string",
			Description: "Type of mapping: extract_field, to_upper, to_lower, to_number, or to_string",
			Enum:        []string{"extract_field", "to_upper", "to_lower", "to_number", "to_string"},
		},
		"reduce_type": {
			Type:        "string",
			Description: "Type of reduction: sum, count, min, max, average, or concat",
			Enum:        []string{"sum", "count", "min", "max", "average", "concat"},
		},
		"sort_order": {
			Type:        "string",
			Description: "Sort order: asc or desc",
			Enum:        []string{"asc", "desc"},
		},
	},
	Required: []string{"data", "operation"},
}

// DataTransform creates a tool for performing data transformations
func DataTransform() domain.Tool {
	return atools.NewTool(
		"data_transform",
		"Transform data: filter, map, reduce, sort, group_by, unique, or reverse",
		func(ctx context.Context, input DataTransformInput) (*DataTransformOutput, error) {
			return executeDataTransform(ctx, input)
		},
		dataTransformParamSchema,
	)
}

// executeDataTransform performs the specified data transformation
func executeDataTransform(ctx context.Context, input DataTransformInput) (*DataTransformOutput, error) {
	// Parse input data
	var data interface{}
	if err := json.Unmarshal([]byte(input.Data), &data); err != nil {
		return &DataTransformOutput{
			Error: fmt.Sprintf("invalid JSON data: %v", err),
		}, nil
	}

	// Ensure data is an array
	dataArray, ok := data.([]interface{})
	if !ok {
		// Try to convert single item to array
		dataArray = []interface{}{data}
	}

	var result interface{}
	var err error

	switch input.Operation {
	case "filter":
		if input.Condition == "" {
			return nil, fmt.Errorf("condition required for filter operation")
		}
		result, err = filterData(dataArray, input.Field, input.Condition)
	case "map":
		if input.MapType == "" {
			return nil, fmt.Errorf("map_type required for map operation")
		}
		result, err = mapData(dataArray, input.Field, input.MapType)
	case "reduce":
		if input.ReduceType == "" {
			return nil, fmt.Errorf("reduce_type required for reduce operation")
		}
		result, err = reduceData(dataArray, input.Field, input.ReduceType)
	case "sort":
		result, err = sortData(dataArray, input.Field, input.SortOrder)
	case "group_by":
		if input.Field == "" {
			return nil, fmt.Errorf("field required for group_by operation")
		}
		result, err = groupByData(dataArray, input.Field)
	case "unique":
		result, err = uniqueData(dataArray, input.Field)
	case "reverse":
		result, err = reverseData(dataArray)
	default:
		return nil, fmt.Errorf("invalid operation: %s", input.Operation)
	}

	if err != nil {
		return &DataTransformOutput{
			Error: err.Error(),
		}, nil
	}

	itemCount := 0
	switch v := result.(type) {
	case []interface{}:
		itemCount = len(v)
	case map[string]interface{}:
		itemCount = len(v)
	default:
		itemCount = 1
	}

	return &DataTransformOutput{
		Result:     result,
		ItemCount:  itemCount,
		ResultType: fmt.Sprintf("%T", result),
	}, nil
}

// filterData applies filtering to the data
func filterData(data []interface{}, field, condition string) ([]interface{}, error) {
	// Parse condition (format: operator:value)
	parts := strings.SplitN(condition, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid condition format. Expected: operator:value")
	}

	operator, value := parts[0], parts[1]
	result := []interface{}{}

	for _, item := range data {
		fieldValue, err := getFieldValue(item, field)
		if err != nil {
			continue
		}

		match := false
		fieldStr := fmt.Sprintf("%v", fieldValue)

		switch operator {
		case "eq", "=", "==":
			match = fieldStr == value
		case "ne", "!=", "<>":
			match = fieldStr != value
		case "contains":
			match = strings.Contains(fieldStr, value)
		case "starts_with":
			match = strings.HasPrefix(fieldStr, value)
		case "ends_with":
			match = strings.HasSuffix(fieldStr, value)
		case "gt", ">":
			match = compareNumeric(fmt.Sprintf("%v", fieldValue), value, ">")
		case "lt", "<":
			match = compareNumeric(fmt.Sprintf("%v", fieldValue), value, "<")
		case "gte", ">=":
			match = compareNumeric(fmt.Sprintf("%v", fieldValue), value, ">=")
		case "lte", "<=":
			match = compareNumeric(fmt.Sprintf("%v", fieldValue), value, "<=")
		case "exists":
			match = fieldValue != nil
		case "not_exists":
			match = fieldValue == nil
		default:
			return nil, fmt.Errorf("unsupported operator: %s", operator)
		}

		if match {
			result = append(result, item)
		}
	}

	return result, nil
}

// mapData applies mapping transformation to the data
func mapData(data []interface{}, field, mapType string) ([]interface{}, error) {
	result := []interface{}{}

	for _, item := range data {
		var mapped interface{}

		switch mapType {
		case "extract_field":
			if field == "" {
				return nil, fmt.Errorf("field required for extract_field mapping")
			}
			fieldValue, _ := getFieldValue(item, field)
			mapped = fieldValue
		case "to_upper":
			if field != "" {
				fieldValue, _ := getFieldValue(item, field)
				mapped = strings.ToUpper(fmt.Sprintf("%v", fieldValue))
			} else {
				mapped = strings.ToUpper(fmt.Sprintf("%v", item))
			}
		case "to_lower":
			if field != "" {
				fieldValue, _ := getFieldValue(item, field)
				mapped = strings.ToLower(fmt.Sprintf("%v", fieldValue))
			} else {
				mapped = strings.ToLower(fmt.Sprintf("%v", item))
			}
		case "to_number":
			var valueToConvert interface{}
			if field != "" {
				valueToConvert, _ = getFieldValue(item, field)
			} else {
				valueToConvert = item
			}
			if num, err := strconv.ParseFloat(fmt.Sprintf("%v", valueToConvert), 64); err == nil {
				mapped = num
			} else {
				mapped = float64(0)
			}
		case "to_string":
			if field != "" {
				fieldValue, _ := getFieldValue(item, field)
				mapped = fmt.Sprintf("%v", fieldValue)
			} else {
				mapped = fmt.Sprintf("%v", item)
			}
		default:
			return nil, fmt.Errorf("unsupported map type: %s", mapType)
		}

		result = append(result, mapped)
	}

	return result, nil
}

// reduceData applies reduction operation to the data
func reduceData(data []interface{}, field, reduceType string) (interface{}, error) {
	if len(data) == 0 {
		return nil, nil
	}

	switch reduceType {
	case "sum":
		sum := 0.0
		for _, item := range data {
			var value interface{}
			if field != "" {
				value, _ = getFieldValue(item, field)
			} else {
				value = item
			}
			if num, err := strconv.ParseFloat(fmt.Sprintf("%v", value), 64); err == nil {
				sum += num
			}
		}
		return sum, nil

	case "count":
		return len(data), nil

	case "min":
		var min interface{}
		for i, item := range data {
			var value interface{}
			if field != "" {
				value, _ = getFieldValue(item, field)
			} else {
				value = item
			}
			if i == 0 || compareValues(value, min, "<") {
				min = value
			}
		}
		return min, nil

	case "max":
		var max interface{}
		for i, item := range data {
			var value interface{}
			if field != "" {
				value, _ = getFieldValue(item, field)
			} else {
				value = item
			}
			if i == 0 || compareValues(value, max, ">") {
				max = value
			}
		}
		return max, nil

	case "average":
		sum := 0.0
		count := 0
		for _, item := range data {
			var value interface{}
			if field != "" {
				value, _ = getFieldValue(item, field)
			} else {
				value = item
			}
			if num, err := strconv.ParseFloat(fmt.Sprintf("%v", value), 64); err == nil {
				sum += num
				count++
			}
		}
		if count > 0 {
			return sum / float64(count), nil
		}
		return 0, nil

	case "concat":
		var result []string
		for _, item := range data {
			var value interface{}
			if field != "" {
				value, _ = getFieldValue(item, field)
			} else {
				value = item
			}
			result = append(result, fmt.Sprintf("%v", value))
		}
		return strings.Join(result, ", "), nil

	default:
		return nil, fmt.Errorf("unsupported reduce type: %s", reduceType)
	}
}

// sortData sorts the data array
func sortData(data []interface{}, field, order string) ([]interface{}, error) {
	if order == "" {
		order = "asc"
	}

	result := make([]interface{}, len(data))
	copy(result, data)

	sort.Slice(result, func(i, j int) bool {
		var val1, val2 interface{}

		if field != "" {
			val1, _ = getFieldValue(result[i], field)
			val2, _ = getFieldValue(result[j], field)
		} else {
			val1 = result[i]
			val2 = result[j]
		}

		less := compareValues(val1, val2, "<")
		if order == "desc" {
			return !less
		}
		return less
	})

	return result, nil
}

// groupByData groups data by a field
func groupByData(data []interface{}, field string) (map[string]interface{}, error) {
	groups := make(map[string]interface{})

	for _, item := range data {
		key, err := getFieldValue(item, field)
		if err != nil {
			key = "undefined"
		}

		keyStr := fmt.Sprintf("%v", key)
		if _, exists := groups[keyStr]; !exists {
			groups[keyStr] = []interface{}{}
		}

		groups[keyStr] = append(groups[keyStr].([]interface{}), item)
	}

	return groups, nil
}

// uniqueData returns unique values from the data
func uniqueData(data []interface{}, field string) ([]interface{}, error) {
	seen := make(map[string]bool)
	result := []interface{}{}

	for _, item := range data {
		var value interface{}
		if field != "" {
			value, _ = getFieldValue(item, field)
		} else {
			value = item
		}

		key := fmt.Sprintf("%v", value)
		if !seen[key] {
			seen[key] = true
			result = append(result, value)
		}
	}

	return result, nil
}

// reverseData reverses the order of elements
func reverseData(data []interface{}) ([]interface{}, error) {
	result := make([]interface{}, len(data))
	for i, item := range data {
		result[len(data)-1-i] = item
	}
	return result, nil
}

// getFieldValue extracts a field value from an item
func getFieldValue(item interface{}, field string) (interface{}, error) {
	if field == "" {
		return item, nil
	}

	// Handle map types
	if m, ok := item.(map[string]interface{}); ok {
		// Support nested field access with dots
		parts := strings.Split(field, ".")
		current := interface{}(m)

		for _, part := range parts {
			if currentMap, ok := current.(map[string]interface{}); ok {
				if val, exists := currentMap[part]; exists {
					current = val
				} else {
					return nil, fmt.Errorf("field not found: %s", part)
				}
			} else {
				return nil, fmt.Errorf("cannot access field %s on non-map", part)
			}
		}

		return current, nil
	}

	// Use reflection for structs
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("cannot extract field from non-struct/non-map type")
	}

	fieldValue := v.FieldByName(field)
	if !fieldValue.IsValid() {
		return nil, fmt.Errorf("field not found: %s", field)
	}

	return fieldValue.Interface(), nil
}

// compareNumeric compares two values numerically
func compareNumeric(a interface{}, b string, op string) bool {
	aNum, aErr := strconv.ParseFloat(fmt.Sprintf("%v", a), 64)
	bNum, bErr := strconv.ParseFloat(b, 64)

	if aErr != nil || bErr != nil {
		return false
	}

	switch op {
	case ">":
		return aNum > bNum
	case "<":
		return aNum < bNum
	case ">=":
		return aNum >= bNum
	case "<=":
		return aNum <= bNum
	}

	return false
}

// compareValues compares two values
func compareValues(a, b interface{}, op string) bool {
	// Try numeric comparison first
	aNum, aErr := strconv.ParseFloat(fmt.Sprintf("%v", a), 64)
	bNum, bErr := strconv.ParseFloat(fmt.Sprintf("%v", b), 64)

	if aErr == nil && bErr == nil {
		switch op {
		case "<":
			return aNum < bNum
		case ">":
			return aNum > bNum
		case "<=":
			return aNum <= bNum
		case ">=":
			return aNum >= bNum
		}
	}

	// Fall back to string comparison
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)

	switch op {
	case "<":
		return aStr < bStr
	case ">":
		return aStr > bStr
	case "<=":
		return aStr <= bStr
	case ">=":
		return aStr >= bStr
	}

	return false
}

func init() {
	tools.MustRegisterTool("data_transform", DataTransform(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "data_transform",
			Category:    "data",
			Tags:        []string{"data", "transform", "filter", "map", "reduce", "sort", "array"},
			Description: "Transform data: filter, map, reduce, sort, group_by, unique, or reverse",
			Version:     "1.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Filter data",
					Description: "Filter array elements based on conditions",
					Code:        `DataTransform().Execute(ctx, DataTransformInput{Data: jsonArray, Operation: "filter", Field: "age", Condition: "gt:25"})`,
				},
				{
					Name:        "Map data",
					Description: "Transform array elements",
					Code:        `DataTransform().Execute(ctx, DataTransformInput{Data: jsonArray, Operation: "map", Field: "name", MapType: "to_upper"})`,
				},
				{
					Name:        "Reduce data",
					Description: "Aggregate array to single value",
					Code:        `DataTransform().Execute(ctx, DataTransformInput{Data: numbers, Operation: "reduce", ReduceType: "sum"})`,
				},
			},
		},
		RequiredPermissions: []string{},
		ResourceUsage: tools.ResourceInfo{
			Memory:      "low",
			Network:     false,
			FileSystem:  false,
			Concurrency: true,
		},
	})
}

// MustGetDataTransform returns the DataTransform tool or panics if not found
func MustGetDataTransform() domain.Tool {
	tool, ok := tools.GetTool("data_transform")
	if !ok {
		panic(fmt.Errorf("data_transform tool not found"))
	}
	return tool
}
