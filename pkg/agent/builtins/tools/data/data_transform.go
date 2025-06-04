// ABOUTME: DataTransform tool provides common data transformation operations
// ABOUTME: This tool enables agents to filter, map, reduce, and transform data without requiring LLM processing

package data

import (
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
		func(ctx *domain.ToolContext, input DataTransformInput) (*DataTransformOutput, error) {
			// Emit start event
			if ctx.Events != nil {
				ctx.Events.EmitMessage(fmt.Sprintf("Starting data transformation with operation: %s", input.Operation))
			}

			// Check for required parameters based on operation
			switch input.Operation {
			case "filter":
				if input.Condition == "" {
					return nil, fmt.Errorf("condition required for filter operation")
				}
			case "map":
				if input.MapType == "" {
					return nil, fmt.Errorf("map_type required for map operation")
				}
			case "reduce":
				if input.ReduceType == "" {
					return nil, fmt.Errorf("reduce_type required for reduce operation")
				}
			case "group_by":
				if input.Field == "" {
					return nil, fmt.Errorf("field required for group_by operation")
				}
			}

			// Check for any transformation defaults in state
			if ctx.State != nil {
				// Check for default sort order
				if input.Operation == "sort" && input.SortOrder == "" {
					if defaultOrder, exists := ctx.State.Get("data_transform_default_sort_order"); exists {
						if order, ok := defaultOrder.(string); ok && (order == "asc" || order == "desc") {
							input.SortOrder = order
						}
					}
				}
			}

			// Set default sort order if not specified
			if input.Operation == "sort" && input.SortOrder == "" {
				input.SortOrder = "asc"
			}

			// Parse input data
			var data interface{}
			if err := json.Unmarshal([]byte(input.Data), &data); err != nil {
				if ctx.Events != nil {
					ctx.Events.EmitError(err)
				}
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

			// Emit progress event
			if ctx.Events != nil {
				ctx.Events.EmitProgress(1, 2, fmt.Sprintf("Processing %d items", len(dataArray)))
			}

			var result interface{}
			var err error

			switch input.Operation {
			case "filter":
				result, err = filterData(dataArray, input.Field, input.Condition)
			case "map":
				result, err = mapData(dataArray, input.Field, input.MapType)
			case "reduce":
				result, err = reduceData(dataArray, input.Field, input.ReduceType)
			case "sort":
				result, err = sortData(dataArray, input.Field, input.SortOrder)
			case "group_by":
				result, err = groupByData(dataArray, input.Field)
			case "unique":
				result, err = uniqueData(dataArray, input.Field)
			case "reverse":
				result, err = reverseData(dataArray)
			default:
				err = fmt.Errorf("invalid operation: %s", input.Operation)
			}

			// Emit completion or error event
			if ctx.Events != nil {
				if err != nil {
					ctx.Events.EmitError(err)
				} else {
					ctx.Events.EmitProgress(2, 2, "Transformation complete")
				}
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

			// Emit final result details
			if ctx.Events != nil {
				ctx.Events.EmitMessage(fmt.Sprintf("Transformation complete. Result contains %d items", itemCount))
			}

			return &DataTransformOutput{
				Result:     result,
				ItemCount:  itemCount,
				ResultType: fmt.Sprintf("%T", result),
			}, nil
		},
		dataTransformParamSchema,
	)
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

		match := false

		switch operator {
		case "exists":
			// For exists operator, check if field exists (no error)
			match = (err == nil) == (value == "true")
		default:
			// For other operators, skip if field doesn't exist
			if err != nil {
				continue
			}

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
				match = compareNumeric(fieldValue, value) > 0
			case "gte", ">=":
				match = compareNumeric(fieldValue, value) >= 0
			case "lt", "<":
				match = compareNumeric(fieldValue, value) < 0
			case "lte", "<=":
				match = compareNumeric(fieldValue, value) <= 0
			default:
				return nil, fmt.Errorf("unknown operator: %s", operator)
			}
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
		var err error

		switch mapType {
		case "extract_field":
			if field == "" {
				return nil, fmt.Errorf("field required for extract_field mapping")
			}
			mapped, err = getFieldValue(item, field)
			if err != nil {
				continue
			}
		case "to_upper":
			if field != "" {
				fieldValue, err := getFieldValue(item, field)
				if err != nil {
					continue
				}
				mapped = strings.ToUpper(fmt.Sprintf("%v", fieldValue))
			} else {
				mapped = strings.ToUpper(fmt.Sprintf("%v", item))
			}
		case "to_lower":
			if field != "" {
				fieldValue, err := getFieldValue(item, field)
				if err != nil {
					continue
				}
				mapped = strings.ToLower(fmt.Sprintf("%v", fieldValue))
			} else {
				mapped = strings.ToLower(fmt.Sprintf("%v", item))
			}
		case "to_number":
			var valueStr string
			if field != "" {
				fieldValue, err := getFieldValue(item, field)
				if err != nil {
					continue
				}
				valueStr = fmt.Sprintf("%v", fieldValue)
			} else {
				valueStr = fmt.Sprintf("%v", item)
			}
			if f, err := strconv.ParseFloat(valueStr, 64); err == nil {
				mapped = f
			} else {
				// If conversion fails, use 0
				mapped = float64(0)
			}
		case "to_string":
			if field != "" {
				fieldValue, err := getFieldValue(item, field)
				if err != nil {
					continue
				}
				mapped = fmt.Sprintf("%v", fieldValue)
			} else {
				mapped = fmt.Sprintf("%v", item)
			}
		default:
			return nil, fmt.Errorf("unknown map type: %s", mapType)
		}

		result = append(result, mapped)
	}

	return result, nil
}

// reduceData applies reduction to the data
func reduceData(data []interface{}, field, reduceType string) (interface{}, error) {
	if len(data) == 0 {
		return nil, nil
	}

	switch reduceType {
	case "sum":
		sum := 0.0
		for _, item := range data {
			value := item
			if field != "" {
				var err error
				value, err = getFieldValue(item, field)
				if err != nil {
					continue
				}
			}
			if num, err := toNumber(value); err == nil {
				sum += num
			}
		}
		return sum, nil
	case "count":
		return len(data), nil
	case "min":
		var min interface{}
		for i, item := range data {
			value := item
			if field != "" {
				var err error
				value, err = getFieldValue(item, field)
				if err != nil {
					continue
				}
			}
			if i == 0 || compareValues(value, min) < 0 {
				min = value
			}
		}
		return min, nil
	case "max":
		var max interface{}
		for i, item := range data {
			value := item
			if field != "" {
				var err error
				value, err = getFieldValue(item, field)
				if err != nil {
					continue
				}
			}
			if i == 0 || compareValues(value, max) > 0 {
				max = value
			}
		}
		return max, nil
	case "average":
		sum := 0.0
		count := 0
		for _, item := range data {
			value := item
			if field != "" {
				var err error
				value, err = getFieldValue(item, field)
				if err != nil {
					continue
				}
			}
			if num, err := toNumber(value); err == nil {
				sum += num
				count++
			}
		}
		if count == 0 {
			return 0, nil
		}
		return sum / float64(count), nil
	case "concat":
		parts := []string{}
		for _, item := range data {
			value := item
			if field != "" {
				var err error
				value, err = getFieldValue(item, field)
				if err != nil {
					continue
				}
			}
			parts = append(parts, fmt.Sprintf("%v", value))
		}
		return strings.Join(parts, ", "), nil
	default:
		return nil, fmt.Errorf("unknown reduce type: %s", reduceType)
	}
}

// sortData sorts the data array
func sortData(data []interface{}, field, order string) ([]interface{}, error) {
	result := make([]interface{}, len(data))
	copy(result, data)

	sort.Slice(result, func(i, j int) bool {
		valI := result[i]
		valJ := result[j]

		if field != "" {
			var err error
			valI, err = getFieldValue(result[i], field)
			if err != nil {
				return false
			}
			valJ, err = getFieldValue(result[j], field)
			if err != nil {
				return false
			}
		}

		cmp := compareValues(valI, valJ)
		if order == "desc" {
			return cmp > 0
		}
		return cmp < 0
	})

	return result, nil
}

// groupByData groups data by field value
func groupByData(data []interface{}, field string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for _, item := range data {
		key, err := getFieldValue(item, field)
		if err != nil {
			continue
		}

		keyStr := fmt.Sprintf("%v", key)
		if group, exists := result[keyStr]; exists {
			result[keyStr] = append(group.([]interface{}), item)
		} else {
			result[keyStr] = []interface{}{item}
		}
	}

	return result, nil
}

// uniqueData returns unique items
func uniqueData(data []interface{}, field string) ([]interface{}, error) {
	seen := make(map[string]bool)
	result := []interface{}{}

	for _, item := range data {
		var key string
		if field != "" {
			value, err := getFieldValue(item, field)
			if err != nil {
				continue
			}
			key = fmt.Sprintf("%v", value)
		} else {
			key = fmt.Sprintf("%v", item)
		}

		if !seen[key] {
			seen[key] = true
			result = append(result, item)
		}
	}

	return result, nil
}

// reverseData reverses the order of items
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
					return nil, fmt.Errorf("field %s not found", part)
				}
			} else {
				return nil, fmt.Errorf("cannot access field %s on non-map", part)
			}
		}
		return current, nil
	}

	// Handle struct types via reflection
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("cannot access field on non-struct/non-map type")
	}

	fieldVal := v.FieldByName(field)
	if !fieldVal.IsValid() {
		return nil, fmt.Errorf("field %s not found", field)
	}

	return fieldVal.Interface(), nil
}

// compareNumeric compares a value with a string representation of a number
func compareNumeric(value interface{}, strValue string) int {
	num1, err1 := toNumber(value)
	num2, err2 := strconv.ParseFloat(strValue, 64)

	if err1 != nil || err2 != nil {
		// Fall back to string comparison
		return strings.Compare(fmt.Sprintf("%v", value), strValue)
	}

	if num1 < num2 {
		return -1
	} else if num1 > num2 {
		return 1
	}
	return 0
}

// toNumber converts a value to float64
func toNumber(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to number", value)
	}
}

// compareValues compares two values
func compareValues(a, b interface{}) int {
	// Try numeric comparison first
	numA, errA := toNumber(a)
	numB, errB := toNumber(b)
	if errA == nil && errB == nil {
		if numA < numB {
			return -1
		} else if numA > numB {
			return 1
		}
		return 0
	}

	// Fall back to string comparison
	strA := fmt.Sprintf("%v", a)
	strB := fmt.Sprintf("%v", b)
	return strings.Compare(strA, strB)
}

func init() {
	tools.MustRegisterTool("data_transform", DataTransform(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "data_transform",
			Category:    "data",
			Tags:        []string{"data", "transform", "filter", "map", "reduce", "sort", "group"},
			Description: "Transform data: filter, map, reduce, sort, group_by, unique, or reverse",
			Version:     "1.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Filter data",
					Description: "Filter array based on conditions",
					Code:        `DataTransform().Execute(ctx, DataTransformInput{Data: jsonArray, Operation: "filter", Field: "age", Condition: "gt:18"})`,
				},
				{
					Name:        "Map data",
					Description: "Transform array elements",
					Code:        `DataTransform().Execute(ctx, DataTransformInput{Data: jsonArray, Operation: "map", MapType: "to_upper"})`,
				},
				{
					Name:        "Reduce data",
					Description: "Aggregate array to single value",
					Code:        `DataTransform().Execute(ctx, DataTransformInput{Data: jsonArray, Operation: "reduce", Field: "price", ReduceType: "sum"})`,
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

// MustGetDataTransform retrieves the registered DataTransform tool or panics
func MustGetDataTransform() domain.Tool {
	return tools.MustGetTool("data_transform")
}
