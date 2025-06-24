# Go-LLMs Built-in Tools

Complete reference for all built-in tools

**Version:** 0.3.5

## Table of Contents

- [data](#data)
- [datetime](#datetime)
- [feed](#feed)
- [file](#file)
- [math](#math)
- [system](#system)
- [web](#web)

## data

### csv_process

Process CSV data: parse, filter, transform, or convert to JSON

Use this tool to process CSV (Comma-Separated Values) data in various ways:

Parse Operation:
- Validates and parses CSV data into a structured format
- Detects headers if has_headers is true
- Returns array of arrays representing rows and columns
- Supports custom delimiters (comma by default)

Filter Operation:
- Filter rows based on conditions using column:operator:value format
- Supported operators:
  - eq, =, ==: Equal to
  - ne, !=, <>: Not equal to
  - contains: Contains substring
  - starts_with: Starts with string
  - ends_with: Ends with string
  - gt, >: Greater than (numeric comparison)
  - lt, <: Less than (numeric comparison)
  - gte, >=: Greater than or equal to
  - lte, <=: Less than or equal to
- Column can be header name (if has_headers) or index (0-based)

Transform Operations:
- select_columns: Select specific columns by name or index
  - Requires params.columns as array or comma-separated string
- sort: Sort records (basic implementation)
- count_rows: Count number of data rows (excluding headers)
- statistics: Calculate statistics for numeric columns
  - Optional params.columns to specify which columns to analyze
  - Returns count, sum, min, max, avg, variance, std_dev

Convert to JSON:
- to_json: Convert CSV to JSON format
  - With headers: Returns array of objects with column names as keys
  - Without headers: Returns array of arrays

Delimiter Support:
- Default is comma (,)
- Can specify any single character delimiter (tab, pipe, semicolon, etc.)
- Common delimiters: "," (comma), "\t" (tab), "|" (pipe), ";" (semicolon)

State Integration:
- csv_default_delimiter: Default delimiter from agent state
- csv_max_rows: Maximum rows to process (for performance limits)

| Property | Value |
|----------|-------|
| **Category** | data |
| **Tags** | data, csv, parse, filter, transform, tabular |
| **Version** | 1.0.0 |
| **package** | github.com/lexlapax/go-llms/pkg/agent/builtins/tools/data |

#### Input Schema

- **Type**: object
- **Required**: data, operation
- **Properties**:
  - **transform**:
    - **Type**: string
    - **Description**: Transform type: select_columns, sort, count_rows, or statistics
    - **Enum**: `select_columns`, `sort`, `count_rows`, `statistics`
  - **data**:
    - **Type**: string
    - **Description**: The CSV data to process
  - **delimiter**:
    - **Type**: string
    - **Description**: Column delimiter character
  - **filter_condition**:
    - **Type**: string
    - **Description**: Filter condition in format column:operator:value
  - **has_headers**:
    - **Type**: boolean
    - **Description**: Whether the first row contains headers
  - **operation**:
    - **Type**: string
    - **Description**: Operation to perform: parse, filter, transform, or to_json
    - **Enum**: `parse`, `filter`, `transform`, `to_json`
  - **params**:
    - **Type**: object
    - **Description**: Additional parameters for transformations

#### Output Schema

- **Type**: object

#### Examples

##### Example 1: Parse CSV with headers

Parse a simple CSV file with column headers

**Input:**
```json
{
  "data": "name,age,city\nJohn,30,New York\nJane,25,Boston\nBob,35,Chicago",
  "has_headers": true,
  "operation": "parse"
}
```

**Output:**
```json
{
  "columns": [
    "name",
    "age",
    "city"
  ],
  "result": [
    [
      "name",
      "age",
      "city"
    ],
    [
      "John",
      "30",
      "New York"
    ],
    [
      "Jane",
      "25",
      "Boston"
    ],
    [
      "Bob",
      "35",
      "Chicago"
    ]
  ],
  "row_count": 3
}
```

##### Example 2: Filter by numeric condition

Filter rows where age is greater than 25

**Input:**
```json
{
  "data": "name,age,city\nJohn,30,New York\nJane,25,Boston\nBob,35,Chicago",
  "filter_condition": "age:gt:25",
  "has_headers": true,
  "operation": "filter"
}
```

**Output:**
```json
{
  "columns": [
    "name",
    "age",
    "city"
  ],
  "result": [
    [
      "name",
      "age",
      "city"
    ],
    [
      "John",
      "30",
      "New York"
    ],
    [
      "Bob",
      "35",
      "Chicago"
    ]
  ],
  "row_count": 2
}
```

##### Example 3: Select specific columns

Extract only name and city columns

**Input:**
```json
{
  "data": "name,age,city,country\nJohn,30,New York,USA\nJane,25,Boston,USA",
  "has_headers": true,
  "operation": "transform",
  "params": {
    "columns": [
      "name",
      "city"
    ]
  },
  "transform": "select_columns"
}
```

**Output:**
```json
{
  "columns": [
    "name",
    "city"
  ],
  "result": [
    [
      "name",
      "city"
    ],
    [
      "John",
      "New York"
    ],
    [
      "Jane",
      "Boston"
    ]
  ],
  "row_count": 2
}
```

##### Example 4: Convert to JSON with headers

Transform CSV into JSON array of objects

**Input:**
```json
{
  "data": "id,name,score\n1,Alice,95\n2,Bob,87",
  "has_headers": true,
  "operation": "to_json"
}
```

**Output:**
```json
{
  "columns": [
    "id",
    "name",
    "score"
  ],
  "result": "[\n  {\n    \"id\": \"1\",\n    \"name\": \"Alice\",\n    \"score\": \"95\"\n  },\n  {\n    \"id\": \"2\",\n    \"name\": \"Bob\",\n    \"score\": \"87\"\n  }\n]",
  "row_count": 2
}
```

##### Example 5: Calculate statistics

Get statistics for numeric columns

**Input:**
```json
{
  "data": "product,price,quantity\nA,10.5,100\nB,20.0,150\nC,15.75,200",
  "has_headers": true,
  "operation": "transform",
  "params": {
    "columns": [
      "price",
      "quantity"
    ]
  },
  "transform": "statistics"
}
```

**Output:**
```json
{
  "result": {
    "column_count": 3,
    "price": {
      "avg": 15.416666666666666,
      "count": 3,
      "max": 20,
      "min": 10.5,
      "std_dev": 19.326388888888886,
      "sum": 46.25,
      "variance": 19.326388888888886
    },
    "quantity": {
      "avg": 150,
      "count": 3,
      "max": 200,
      "min": 100,
      "std_dev": 2500,
      "sum": 450,
      "variance": 2500
    },
    "row_count": 3
  },
  "row_count": 3
}
```

##### Example 6: Parse with custom delimiter

Parse tab-separated values

**Input:**
```json
{
  "data": "name\tage\tcity\nJohn\t30\tNew York\nJane\t25\tBoston",
  "delimiter": "\t",
  "has_headers": true,
  "operation": "parse"
}
```

**Output:**
```json
{
  "columns": [
    "name",
    "age",
    "city"
  ],
  "result": [
    [
      "name",
      "age",
      "city"
    ],
    [
      "John",
      "30",
      "New York"
    ],
    [
      "Jane",
      "25",
      "Boston"
    ]
  ],
  "row_count": 2
}
```

##### Example 7: Filter with string contains

Find all rows where city contains 'New'

**Input:**
```json
{
  "data": "name,city\nJohn,New York\nJane,Boston\nBob,New Orleans",
  "filter_condition": "city:contains:New",
  "has_headers": true,
  "operation": "filter"
}
```

**Output:**
```json
{
  "columns": [
    "name",
    "city"
  ],
  "result": [
    [
      "name",
      "city"
    ],
    [
      "John",
      "New York"
    ],
    [
      "Bob",
      "New Orleans"
    ]
  ],
  "row_count": 2
}
```

---

### data_transform

Transform data: filter, map, reduce, sort, group_by, unique, or reverse

Use this tool to perform common data transformation operations on JSON arrays:

Filter Operation:
- Extract items matching specific conditions
- Condition format: "operator:value"
- Supported operators:
  - eq, =, ==: Equal to
  - ne, !=, <>: Not equal to
  - gt, >: Greater than
  - gte, >=: Greater than or equal to
  - lt, <: Less than
  - lte, <=: Less than or equal to
  - contains: Contains substring
  - starts_with: Starts with string
  - ends_with: Ends with string
  - exists: Field exists (value should be "true" or "false")
- Field can be nested using dots: "address.city"

Map Operation:
- Transform each item in the array
- Map types:
  - extract_field: Extract specific field from objects
  - to_upper: Convert to uppercase
  - to_lower: Convert to lowercase
  - to_number: Convert to numeric value
  - to_string: Convert to string representation

Reduce Operation:
- Aggregate array to a single value
- Reduce types:
  - sum: Sum numeric values
  - count: Count items
  - min: Find minimum value
  - max: Find maximum value
  - average: Calculate average of numeric values
  - concat: Concatenate as comma-separated string

Sort Operation:
- Sort array by value or field
- Order: "asc" (ascending) or "desc" (descending)
- Supports numeric and string sorting

Group By Operation:
- Group items by field value
- Returns object with field values as keys
- Each key contains array of matching items

Unique Operation:
- Remove duplicate items
- Can use field for uniqueness check
- Preserves first occurrence

Reverse Operation:
- Reverse the order of array items
- Simple operation, no parameters needed

Operation Chaining:
- For complex transformations, consider chaining multiple operations
- Example: filter → map → sort → unique

State Integration:
- data_transform_default_sort_order: Default sort order from agent state

| Property | Value |
|----------|-------|
| **Category** | data |
| **Tags** | data, transform, filter, map, reduce, sort, group |
| **Version** | 1.0.0 |
| **package** | github.com/lexlapax/go-llms/pkg/agent/builtins/tools/data |

#### Input Schema

- **Type**: object
- **Required**: data, operation
- **Properties**:
  - **map_type**:
    - **Type**: string
    - **Description**: Type of mapping: extract_field, to_upper, to_lower, to_number, or to_string
    - **Enum**: `extract_field`, `to_upper`, `to_lower`, `to_number`, `to_string`
  - **operation**:
    - **Type**: string
    - **Description**: Operation to perform: filter, map, reduce, sort, group_by, unique, or reverse
    - **Enum**: `filter`, `map`, `reduce`, `sort`, `group_by`, `unique`, `reverse`
  - **reduce_type**:
    - **Type**: string
    - **Description**: Type of reduction: sum, count, min, max, average, or concat
    - **Enum**: `sum`, `count`, `min`, `max`, `average`, `concat`
  - **sort_order**:
    - **Type**: string
    - **Description**: Sort order: asc or desc
    - **Enum**: `asc`, `desc`
  - **condition**:
    - **Type**: string
    - **Description**: Condition for filter operation in format: operator:value
  - **data**:
    - **Type**: string
    - **Description**: The data to transform as JSON string or array
  - **field**:
    - **Type**: string
    - **Description**: Field name or path for the operation

#### Output Schema

- **Type**: object

#### Examples

##### Example 1: Filter by numeric condition

Filter users older than 25

**Input:**
```json
{
  "condition": "gt:25",
  "data": "[{\"name\":\"Alice\",\"age\":30},{\"name\":\"Bob\",\"age\":22},{\"name\":\"Carol\",\"age\":28}]",
  "field": "age",
  "operation": "filter"
}
```

**Output:**
```json
{
  "item_count": 2,
  "result": [
    {
      "age": null,
      "name": "Alice"
    },
    {
      "age": null,
      "name": "Carol"
    }
  ],
  "result_type": "[]interface {}"
}
```

##### Example 2: Map to extract field

Extract names from user objects

**Input:**
```json
{
  "data": "[{\"name\":\"Alice\",\"age\":30},{\"name\":\"Bob\",\"age\":22}]",
  "field": "name",
  "map_type": "extract_field",
  "operation": "map"
}
```

**Output:**
```json
{
  "item_count": 2,
  "result": [
    "Alice",
    "Bob"
  ],
  "result_type": "[]interface {}"
}
```

##### Example 3: Reduce to sum prices

Calculate total price from product list

**Input:**
```json
{
  "data": "[{\"product\":\"A\",\"price\":10.5},{\"product\":\"B\",\"price\":20},{\"product\":\"C\",\"price\":15.5}]",
  "field": "price",
  "operation": "reduce",
  "reduce_type": "sum"
}
```

**Output:**
```json
{
  "item_count": 1,
  "result": null,
  "result_type": "float64"
}
```

##### Example 4: Sort by field descending

Sort products by price from high to low

**Input:**
```json
{
  "data": "[{\"name\":\"A\",\"price\":30},{\"name\":\"B\",\"price\":10},{\"name\":\"C\",\"price\":20}]",
  "field": "price",
  "operation": "sort",
  "sort_order": "desc"
}
```

**Output:**
```json
{
  "item_count": 3,
  "result": [
    {
      "name": "A",
      "price": null
    },
    {
      "name": "C",
      "price": null
    },
    {
      "name": "B",
      "price": null
    }
  ],
  "result_type": "[]interface {}"
}
```

##### Example 5: Group by category

Group products by their category

**Input:**
```json
{
  "data": "[{\"name\":\"Apple\",\"category\":\"fruit\"},{\"name\":\"Carrot\",\"category\":\"vegetable\"},{\"name\":\"Banana\",\"category\":\"fruit\"}]",
  "field": "category",
  "operation": "group_by"
}
```

**Output:**
```json
{
  "item_count": 2,
  "result": {
    "fruit": [
      {
        "category": "fruit",
        "name": "Apple"
      },
      {
        "category": "fruit",
        "name": "Banana"
      }
    ],
    "vegetable": [
      {
        "category": "vegetable",
        "name": "Carrot"
      }
    ]
  },
  "result_type": "map[string]interface {}"
}
```

##### Example 6: Get unique values

Remove duplicate tags

**Input:**
```json
{
  "data": "[\"python\",\"javascript\",\"python\",\"go\",\"javascript\",\"rust\"]",
  "operation": "unique"
}
```

**Output:**
```json
{
  "item_count": 4,
  "result": [
    "python",
    "javascript",
    "go",
    "rust"
  ],
  "result_type": "[]interface {}"
}
```

##### Example 7: Transform strings to uppercase

Convert all strings to uppercase

**Input:**
```json
{
  "data": "[\"hello\",\"world\",\"data\",\"transform\"]",
  "map_type": "to_upper",
  "operation": "map"
}
```

**Output:**
```json
{
  "item_count": 4,
  "result": [
    "HELLO",
    "WORLD",
    "DATA",
    "TRANSFORM"
  ],
  "result_type": "[]interface {}"
}
```

##### Example 8: Filter with nested field

Filter by nested object property

**Input:**
```json
{
  "condition": "eq:NYC",
  "data": "[{\"user\":\"A\",\"profile\":{\"city\":\"NYC\"}},{\"user\":\"B\",\"profile\":{\"city\":\"LA\"}}]",
  "field": "profile.city",
  "operation": "filter"
}
```

**Output:**
```json
{
  "item_count": 1,
  "result": [
    {
      "profile": {
        "city": "NYC"
      },
      "user": "A"
    }
  ],
  "result_type": "[]interface {}"
}
```

##### Example 9: Calculate average

Find average score

**Input:**
```json
{
  "data": "[{\"name\":\"Test1\",\"score\":85},{\"name\":\"Test2\",\"score\":90},{\"name\":\"Test3\",\"score\":78}]",
  "field": "score",
  "operation": "reduce",
  "reduce_type": "average"
}
```

**Output:**
```json
{
  "item_count": 1,
  "result": null,
  "result_type": "float64"
}
```

##### Example 10: Operation chain example

First filter, then map (requires two operations)

**Input:**
```json
{
  "condition": "eq:true",
  "data": "[{\"name\":\"Alice\",\"age\":30,\"active\":true},{\"name\":\"Bob\",\"age\":22,\"active\":false}]",
  "field": "active",
  "operation": "filter"
}
```

**Output:**
```json
{
  "item_count": 1,
  "result": [
    {
      "active": true,
      "age": null,
      "name": "Alice"
    }
  ],
  "result_type": "[]interface {}"
}
```

---

### json_process

Process JSON data: parse, query with JSONPath, or transform

Use this tool to process JSON data in various ways:

Parse Operation:
- Validates JSON syntax and parses the data
- Returns the parsed data structure and its type
- Useful for checking if data is valid JSON

Query Operation (JSONPath):
- Extract specific values using JSONPath expressions
- Supports basic JSONPath syntax:
  - $ or empty: Root object
  - .field: Access object field
  - [n]: Array index access
  - .field[n]: Combination of field and array access
  - Nested paths: $.users[0].address.city

Transform Operations:
- extract_keys: Get all keys from the JSON structure (includes nested paths)
- extract_values: Get all leaf values from the JSON
- flatten: Convert nested JSON to flat key-value pairs
- prettify: Format JSON with indentation for readability
- minify: Remove unnecessary whitespace for compact representation

JSONPath Examples:
- $.name: Get the 'name' field from root
- $.users[0]: Get the first user from 'users' array
- $.users[0].email: Get email of the first user
- $.products[*].price: Get all product prices (Note: [*] not fully supported in basic implementation)

For complex JSONPath queries beyond basic field and array access, consider using the result of a parse operation and processing it further.

| Property | Value |
|----------|-------|
| **Category** | data |
| **Tags** | data, json, parse, query, transform, jsonpath |
| **Version** | 1.0.0 |
| **package** | github.com/lexlapax/go-llms/pkg/agent/builtins/tools/data |

#### Input Schema

- **Type**: object
- **Required**: data, operation
- **Properties**:
  - **operation**:
    - **Type**: string
    - **Description**: Operation to perform: parse, query, or transform
    - **Enum**: `parse`, `query`, `transform`
  - **transform**:
    - **Type**: string
    - **Description**: Transformation type
    - **Enum**: `extract_keys`, `extract_values`, `flatten`, `prettify`, `minify`
  - **data**:
    - **Type**: string
    - **Description**: The JSON data to process
  - **jsonpath**:
    - **Type**: string
    - **Description**: JSONPath expression for query operation

#### Output Schema

- **Type**: object

#### Examples

##### Example 1: Parse and validate JSON

Check if a string is valid JSON and see its structure

**Input:**
```json
{
  "data": "{\"name\": \"John\", \"age\": 30, \"city\": \"New York\"}",
  "operation": "parse"
}
```

**Output:**
```json
{
  "result": {
    "age": null,
    "city": "New York",
    "name": "John"
  },
  "result_type": "map[string]interface {}"
}
```

##### Example 2: Query nested data with JSONPath

Extract specific values from complex JSON structures

**Input:**
```json
{
  "data": "{\"users\": [{\"id\": 1, \"name\": \"Alice\", \"email\": \"alice@example.com\"}, {\"id\": 2, \"name\": \"Bob\", \"email\": \"bob@example.com\"}]}",
  "jsonpath": "users[0].email",
  "operation": "query"
}
```

**Output:**
```json
{
  "result": "alice@example.com",
  "result_type": "string"
}
```

##### Example 3: Extract all keys from JSON

Get a list of all keys including nested paths

**Input:**
```json
{
  "data": "{\"user\": {\"name\": \"John\", \"address\": {\"city\": \"NYC\", \"zip\": \"10001\"}}, \"active\": true}",
  "operation": "transform",
  "transform": "extract_keys"
}
```

**Output:**
```json
{
  "result": [
    "user",
    "user.name",
    "user.address",
    "user.address.city",
    "user.address.zip",
    "active"
  ],
  "result_type": "[]string"
}
```

##### Example 4: Flatten nested JSON

Convert nested structure to flat key-value pairs

**Input:**
```json
{
  "data": "{\"user\": {\"name\": \"John\", \"scores\": [85, 90, 78]}, \"active\": true}",
  "operation": "transform",
  "transform": "flatten"
}
```

**Output:**
```json
{
  "result": {
    "active": true,
    "user.name": "John",
    "user.scores[0]": null,
    "user.scores[1]": null,
    "user.scores[2]": null
  },
  "result_type": "map[string]interface {}"
}
```

##### Example 5: Pretty print JSON

Format JSON with proper indentation

**Input:**
```json
{
  "data": "{\"name\":\"John\",\"age\":30,\"city\":\"New York\"}",
  "operation": "transform",
  "transform": "prettify"
}
```

**Output:**
```json
{
  "result": "{\n  \"name\": \"John\",\n  \"age\": 30,\n  \"city\": \"New York\"\n}",
  "result_type": "string"
}
```

##### Example 6: Handle invalid JSON gracefully

Error handling for malformed JSON

**Input:**
```json
{
  "data": "{\"name\": \"John\", \"age\": 30,}",
  "operation": "parse"
}
```

**Output:**
```json
{
  "error": "invalid JSON: invalid character '}' after object key:value pair",
  "result_type": ""
}
```

##### Example 7: Query array elements

Access specific elements in JSON arrays

**Input:**
```json
{
  "data": "{\"items\": [\"apple\", \"banana\", \"cherry\"], \"count\": 3}",
  "jsonpath": "items[1]",
  "operation": "query"
}
```

**Output:**
```json
{
  "result": "banana",
  "result_type": "string"
}
```

---

### xml_process

Process XML data: parse, query with simplified XPath, or convert to JSON

Use this tool to process XML data in various ways:

Parse Operation:
- Validates XML syntax and parses the data
- Returns a structured representation of the XML
- Preserves element hierarchy and relationships
- Optionally includes attributes (controlled by include_attributes)

Query Operation (Simplified XPath):
- Extract specific elements using path expressions
- Supports basic XPath-like syntax:
  - /: Path separator
  - element: Select elements by name
  - *: Select all child elements
  - @attribute: Select attribute values
  - element/child: Navigate hierarchy
- Returns matching elements or attribute values

Convert to JSON:
- Transforms XML structure to JSON format
- Preserves element names, attributes, and text content
- Useful for working with XML data in JSON-based systems

XML Structure Representation:
- _name: Element name
- _attributes: Element attributes (if include_attributes is true)
- _text: Text content (for leaf elements)
- Child elements are added as properties

Namespace Support:
- Basic namespace handling (local names are used)
- For complex namespace scenarios, consider parsing and processing manually

State Integration:
- xml_include_attributes_default: Default value for include_attributes option

| Property | Value |
|----------|-------|
| **Category** | data |
| **Tags** | data, xml, parse, query, xpath, transform |
| **Version** | 1.0.0 |
| **package** | github.com/lexlapax/go-llms/pkg/agent/builtins/tools/data |

#### Input Schema

- **Type**: object
- **Required**: data, operation
- **Properties**:
  - **operation**:
    - **Type**: string
    - **Description**: Operation to perform: parse, query, or to_json
    - **Enum**: `parse`, `query`, `to_json`
  - **xpath**:
    - **Type**: string
    - **Description**: Simplified XPath query for query operation
  - **data**:
    - **Type**: string
    - **Description**: The XML data to process
  - **include_attributes**:
    - **Type**: boolean
    - **Description**: Whether to include XML attributes in the result

#### Output Schema

- **Type**: object

#### Examples

##### Example 1: Parse simple XML

Parse a basic XML document with attributes

**Input:**
```json
{
  "data": "\u003cbook id=\"123\"\u003e\n  \u003ctitle\u003eGo Programming\u003c/title\u003e\n  \u003cauthor\u003eJohn Doe\u003c/author\u003e\n  \u003cyear\u003e2024\u003c/year\u003e\n\u003c/book\u003e",
  "include_attributes": true,
  "operation": "parse"
}
```

**Output:**
```json
{
  "result": {
    "_attributes": {
      "id": "123"
    },
    "_name": "book",
    "author": {
      "_name": "author",
      "_text": "John Doe"
    },
    "title": {
      "_name": "title",
      "_text": "Go Programming"
    },
    "year": {
      "_name": "year",
      "_text": "2024"
    }
  },
  "root_element": "book"
}
```

##### Example 2: Query specific element

Extract book title using XPath-like query

**Input:**
```json
{
  "data": "\u003ccatalog\u003e\n  \u003cbook\u003e\n    \u003ctitle\u003eXML Processing\u003c/title\u003e\n    \u003cprice\u003e29.99\u003c/price\u003e\n  \u003c/book\u003e\n  \u003cbook\u003e\n    \u003ctitle\u003eData Formats\u003c/title\u003e\n    \u003cprice\u003e34.99\u003c/price\u003e\n  \u003c/book\u003e\n\u003c/catalog\u003e",
  "operation": "query",
  "xpath": "book/title"
}
```

**Output:**
```json
{
  "result": [
    {
      "_name": "title",
      "_text": "XML Processing"
    },
    {
      "_name": "title",
      "_text": "Data Formats"
    }
  ],
  "root_element": "catalog"
}
```

##### Example 3: Query attribute value

Extract attribute values using @ notation

**Input:**
```json
{
  "data": "\u003cusers\u003e\n  \u003cuser id=\"u1\" role=\"admin\"\u003eAlice\u003c/user\u003e\n  \u003cuser id=\"u2\" role=\"user\"\u003eBob\u003c/user\u003e\n\u003c/users\u003e",
  "operation": "query",
  "xpath": "user/@id"
}
```

**Output:**
```json
{
  "result": [
    "u1",
    "u2"
  ],
  "root_element": "users"
}
```

##### Example 4: Convert XML to JSON

Transform XML data into JSON format

**Input:**
```json
{
  "data": "\u003cperson\u003e\n  \u003cname\u003eJane Smith\u003c/name\u003e\n  \u003cemail\u003ejane@example.com\u003c/email\u003e\n  \u003cskills\u003e\n    \u003cskill\u003ePython\u003c/skill\u003e\n    \u003cskill\u003eXML\u003c/skill\u003e\n    \u003cskill\u003eJSON\u003c/skill\u003e\n  \u003c/skills\u003e\n\u003c/person\u003e",
  "include_attributes": false,
  "operation": "to_json"
}
```

**Output:**
```json
{
  "result": "{\n  \"_name\": \"person\",\n  \"name\": {\n    \"_name\": \"name\",\n    \"_text\": \"Jane Smith\"\n  },\n  \"email\": {\n    \"_name\": \"email\",\n    \"_text\": \"jane@example.com\"\n  },\n  \"skills\": {\n    \"_name\": \"skills\",\n    \"skill\": [\n      {\n        \"_name\": \"skill\",\n        \"_text\": \"Python\"\n      },\n      {\n        \"_name\": \"skill\",\n        \"_text\": \"XML\"\n      },\n      {\n        \"_name\": \"skill\",\n        \"_text\": \"JSON\"\n      }\n    ]\n  }\n}",
  "root_element": "person"
}
```

##### Example 5: Handle namespaced XML

Parse XML with namespace declarations

**Input:**
```json
{
  "data": "\u003cns:root xmlns:ns=\"http://example.com/ns\"\u003e\n  \u003cns:item\u003eNamespaced content\u003c/ns:item\u003e\n\u003c/ns:root\u003e",
  "operation": "parse"
}
```

**Output:**
```json
{
  "result": {
    "_name": "root",
    "item": {
      "_name": "item",
      "_text": "Namespaced content"
    }
  },
  "root_element": "root"
}
```

##### Example 6: Query all children with wildcard

Use * to select all child elements

**Input:**
```json
{
  "data": "\u003cconfig\u003e\n  \u003cdatabase\u003eMySQL\u003c/database\u003e\n  \u003ccache\u003eRedis\u003c/cache\u003e\n  \u003cqueue\u003eRabbitMQ\u003c/queue\u003e\n\u003c/config\u003e",
  "operation": "query",
  "xpath": "*"
}
```

**Output:**
```json
{
  "result": [
    {
      "_name": "database",
      "_text": "MySQL"
    },
    {
      "_name": "cache",
      "_text": "Redis"
    },
    {
      "_name": "queue",
      "_text": "RabbitMQ"
    }
  ],
  "root_element": "config"
}
```

##### Example 7: Handle invalid XML gracefully

Error handling for malformed XML

**Input:**
```json
{
  "data": "\u003croot\u003e\u003cunclosed\u003e",
  "operation": "parse"
}
```

**Output:**
```json
{
  "error": "invalid XML: XML syntax error on line 1: unexpected EOF"
}
```

---

## datetime

### datetime_calculate

Perform date/time arithmetic operations including add/subtract, duration, age, and business days

The datetime_calculate tool performs various date/time arithmetic operations:
- add/subtract: Add or subtract years, months, days, hours, minutes, or seconds
- duration: Calculate the duration between two dates
- age: Calculate age from birth date to current date or specified date
- next_weekday/previous_weekday: Find the next or previous occurrence of a specific weekday
- add_business_days/subtract_business_days: Add or subtract business days (excluding weekends)

All operations support timezone specification and use RFC3339 format for dates.

| Property | Value |
|----------|-------|
| **Category** | datetime |
| **Tags** | datetime, arithmetic, duration, age, business-days, weekday, calendar |
| **Version** | 2.0.0 |
| **package** | github.com/lexlapax/go-llms/pkg/agent/builtins/tools/datetime |

#### Input Schema

- **Type**: object
- **Required**: operation, start_date
- **Properties**:
  - **target_weekday**:
    - **Type**: integer
    - **Description**: Target weekday (0 = Sunday, 6 = Saturday)
  - **timezone**:
    - **Type**: string
    - **Description**: Timezone for calculations
  - **unit**:
    - **Type**: string
    - **Description**: Unit for add/subtract operations
    - **Enum**: `years`, `months`, `days`, `hours`, `minutes`, `seconds`
  - **value**:
    - **Type**: integer
    - **Description**: Value to add/subtract
  - **end_date**:
    - **Type**: string
    - **Description**: End date/time for duration calculations
  - **include_weekends**:
    - **Type**: boolean
    - **Description**: Include weekends in business day calculations
  - **operation**:
    - **Type**: string
    - **Description**: Operation to perform
    - **Enum**: `add`, `subtract`, `duration`, `age`, `next_weekday`, `previous_weekday`, `add_business_days`, `subtract_business_days`
  - **start_date**:
    - **Type**: string
    - **Description**: Start date/time (RFC3339 format preferred)

#### Output Schema

- **Type**: object

#### Examples

##### Example 1: Add days

Add 30 days to a date

**Input:**
```json
{
  "operation": "add",
  "start_date": "2024-01-15",
  "unit": "days",
  "value": 30
}
```

**Output:**
```json
{
  "result": "2024-02-14T00:00:00Z"
}
```

##### Example 2: Calculate age

Calculate age from birth date

**Input:**
```json
{
  "operation": "age",
  "start_date": "1990-05-15"
}
```

**Output:**
```json
{
  "age": {
    "days": 25,
    "human_readable": "34 years, 7 months and 25 days",
    "months": 7,
    "total_days": 12653,
    "years": 34
  }
}
```

##### Example 3: Duration between dates

Calculate duration between two dates

**Input:**
```json
{
  "end_date": "2024-12-31",
  "operation": "duration",
  "start_date": "2024-01-01"
}
```

**Output:**
```json
{
  "duration": {
    "days": 365,
    "hours": 0,
    "human_readable": "365 days",
    "milliseconds": 0,
    "minutes": 0,
    "seconds": 0,
    "total_seconds": 31536000
  }
}
```

##### Example 4: Add business days

Add 10 business days to a date

**Input:**
```json
{
  "operation": "add_business_days",
  "start_date": "2024-03-01",
  "value": 10
}
```

**Output:**
```json
{
  "business_days": 10,
  "result": "2024-03-15T00:00:00Z"
}
```

##### Example 5: Next weekday

Find next Monday from a given date

**Input:**
```json
{
  "operation": "next_weekday",
  "start_date": "2024-03-15",
  "target_weekday": 1
}
```

**Output:**
```json
{
  "result": "2024-03-18T00:00:00Z"
}
```

---

### datetime_compare

Compare dates and times with operations like before/after, same period, range checks, and sorting

The datetime_compare tool provides comprehensive date/time comparison capabilities:

Operations:
1. Compare:
   - Check if date1 is before, after, or equal to date2
   - Calculate detailed time difference between dates
   - Provides human-readable difference format
   - Handles timezones correctly

2. Same Period:
   - Check if two dates fall within the same period
   - Supported periods: day, week, month, year
   - Week comparisons use ISO week numbering
   - Timezone-aware comparisons

3. Range Check:
   - Verify if a date falls within a specified range
   - Inclusive of both start and end dates
   - Useful for deadline validation
   - Date range filtering

4. Sort:
   - Sort multiple dates in ascending or descending order
   - Handles any parseable date format
   - Returns dates in RFC3339 format
   - Efficient for large date lists

5. Find Extreme:
   - Find the earliest or latest date from a list
   - Useful for finding min/max dates
   - Deadline calculations
   - Event scheduling

Time Difference Output:
- Days, hours, minutes, seconds breakdown
- Total time in hours, minutes, and seconds
- Human-readable format (e.g., "2 days, 3 hours and 15 minutes")
- Direction indicator (ago/in the future)

State Integration:
- default_timezone: Used when timezone not specified in input

Common Use Cases:
- Deadline validation and monitoring
- Event scheduling and conflict detection
- Age calculations and date arithmetic
- Historical data analysis
- Time period grouping and filtering

| Property | Value |
|----------|-------|
| **Category** | datetime |
| **Tags** | datetime, comparison, before, after, range, sort, period, difference |
| **Version** | 2.0.0 |
| **package** | github.com/lexlapax/go-llms/pkg/agent/builtins/tools/datetime |

#### Input Schema

- **Type**: object

#### Output Schema

- **Type**: object

#### Examples

##### Example 1: Basic date comparison

Compare two dates and get time difference

**Input:**
```json
{
  "date1": "2024-03-15T10:30:00Z",
  "date2": "2024-03-20T15:45:00Z",
  "operation": "compare"
}
```

**Output:**
```json
{
  "after": false,
  "before": true,
  "difference": {
    "days": 5,
    "hours": 5,
    "human_readable": "in 5 days, 5 hours and 15 minutes",
    "minutes": 15,
    "seconds": 0,
    "total_hours": 125.25,
    "total_minutes": 7515,
    "total_seconds": 450900
  },
  "equal": false
}
```

##### Example 2: Check same month

Verify if two dates are in the same month

**Input:**
```json
{
  "date1": "2024-03-05",
  "date2": "2024-03-25",
  "operation": "same_period",
  "period_type": "month"
}
```

**Output:**
```json
{
  "same_period": true
}
```

##### Example 3: Check same week

Check if dates fall in the same ISO week

**Input:**
```json
{
  "date1": "2024-03-18",
  "date2": "2024-03-22",
  "operation": "same_period",
  "period_type": "week"
}
```

**Output:**
```json
{
  "same_period": true
}
```

##### Example 4: Date range validation

Check if a date falls within a range

**Input:**
```json
{
  "date1": "2024-03-15",
  "operation": "range_check",
  "range_end": "2024-03-31",
  "range_start": "2024-03-01"
}
```

**Output:**
```json
{
  "in_range": true
}
```

##### Example 5: Sort dates chronologically

Sort a list of dates in ascending order

**Input:**
```json
{
  "dates": [
    "2024-12-25",
    "2024-01-01",
    "2024-07-04",
    "2024-03-15"
  ],
  "operation": "sort",
  "sort_order": "asc"
}
```

**Output:**
```json
{
  "sorted_dates": [
    "2024-01-01T00:00:00Z",
    "2024-03-15T00:00:00Z",
    "2024-07-04T00:00:00Z",
    "2024-12-25T00:00:00Z"
  ]
}
```

##### Example 6: Sort dates descending

Sort dates from newest to oldest

**Input:**
```json
{
  "dates": [
    "2024-01-15",
    "2024-03-10",
    "2024-02-20"
  ],
  "operation": "sort",
  "sort_order": "desc"
}
```

**Output:**
```json
{
  "sorted_dates": [
    "2024-03-10T00:00:00Z",
    "2024-02-20T00:00:00Z",
    "2024-01-15T00:00:00Z"
  ]
}
```

##### Example 7: Find earliest date

Find the earliest date from a list

**Input:**
```json
{
  "dates": [
    "2024-06-15",
    "2024-03-01",
    "2024-12-31",
    "2024-01-10"
  ],
  "extreme_type": "earliest",
  "operation": "find_extreme"
}
```

**Output:**
```json
{
  "extreme_date": "2024-01-10T00:00:00Z"
}
```

##### Example 8: Find latest date

Find the most recent date

**Input:**
```json
{
  "dates": [
    "2024-02-28",
    "2024-03-15",
    "2024-01-01"
  ],
  "extreme_type": "latest",
  "operation": "find_extreme"
}
```

**Output:**
```json
{
  "extreme_date": "2024-03-15T00:00:00Z"
}
```

##### Example 9: Compare with timezone

Compare dates in specific timezone

**Input:**
```json
{
  "date1": "2024-03-15 10:00:00",
  "date2": "2024-03-15 14:00:00",
  "operation": "compare",
  "timezone": "America/New_York"
}
```

**Output:**
```json
{
  "after": false,
  "before": true,
  "difference": {
    "days": 0,
    "hours": 4,
    "human_readable": "in 4 hours",
    "minutes": 0,
    "seconds": 0,
    "total_hours": 4,
    "total_minutes": 240,
    "total_seconds": 14400
  },
  "equal": false
}
```

##### Example 10: Compare past dates

Compare dates where date1 is after date2

**Input:**
```json
{
  "date1": "2024-03-20",
  "date2": "2024-03-15",
  "operation": "compare"
}
```

**Output:**
```json
{
  "after": true,
  "before": false,
  "difference": {
    "days": 5,
    "hours": 0,
    "human_readable": "5 days ago",
    "minutes": 0,
    "seconds": 0,
    "total_hours": 120,
    "total_minutes": 7200,
    "total_seconds": 432000
  },
  "equal": false
}
```

---

### datetime_convert

Convert date/time between timezones, unix timestamps, and provide timezone information

The datetime_convert tool provides comprehensive timezone and timestamp conversion capabilities:

Operations:
1. Timezone Conversion:
   - Convert dates between any IANA timezones
   - Handles DST transitions automatically
   - Preserves the exact moment in time
   - Supports source timezone specification

2. To Timestamp:
   - Convert datetime to unix timestamps
   - Outputs in seconds, milliseconds, microseconds, and nanoseconds
   - Respects timezone information in the input
   - Uses state timezone or UTC as default

3. From Timestamp:
   - Convert unix timestamps to datetime
   - Supports various units: seconds, milliseconds, microseconds, nanoseconds
   - Can output in any specified timezone
   - Defaults to UTC if no timezone specified

4. List Timezones:
   - Get filtered list of IANA timezone identifiers
   - Supports substring filtering
   - Returns common timezones by default
   - Useful for timezone discovery

Timezone Information:
- Full IANA timezone names (e.g., "America/New_York")
- Timezone abbreviations (e.g., "EST", "EDT")
- UTC offsets in "+/-HH:MM" format
- Offset in seconds for calculations

DST (Daylight Saving Time) Information:
- Current DST status
- Standard and DST timezone names
- Standard and DST offsets
- Automatic detection of DST periods

State Integration:
- default_timezone: Used when from_timezone not specified for to_timestamp

Common Timezone Examples:
- Americas: America/New_York, America/Chicago, America/Los_Angeles, America/Toronto
- Europe: Europe/London, Europe/Paris, Europe/Berlin, Europe/Moscow
- Asia: Asia/Tokyo, Asia/Shanghai, Asia/Kolkata, Asia/Dubai
- Pacific: Australia/Sydney, Pacific/Auckland, Pacific/Honolulu

| Property | Value |
|----------|-------|
| **Category** | datetime |
| **Tags** | datetime, timezone, conversion, timestamp, unix, dst, iana |
| **Version** | 2.0.0 |
| **package** | github.com/lexlapax/go-llms/pkg/agent/builtins/tools/datetime |

#### Input Schema

- **Type**: object

#### Output Schema

- **Type**: object

#### Examples

##### Example 1: Convert between timezones

Convert a New York time to Tokyo time

**Input:**
```json
{
  "datetime": "2024-03-15T10:30:00-04:00",
  "from_timezone": "America/New_York",
  "operation": "timezone",
  "to_timezone": "Asia/Tokyo"
}
```

**Output:**
```json
{
  "converted": "2024-03-15T23:30:00+09:00",
  "timezone_info": {
    "abbreviation": "JST",
    "name": "Asia/Tokyo",
    "offset": "+09:00",
    "offset_seconds": 32400
  }
}
```

##### Example 2: Convert with DST information

Convert time and get DST details

**Input:**
```json
{
  "datetime": "2024-07-15T10:00:00Z",
  "include_dst": true,
  "operation": "timezone",
  "to_timezone": "America/New_York"
}
```

**Output:**
```json
{
  "converted": "2024-07-15T06:00:00-04:00",
  "dst_info": {
    "current_offset": "-04:00",
    "dst_name": "EDT",
    "dst_offset": "-04:00",
    "is_dst": true,
    "standard_name": "EST",
    "standard_offset": "-05:00"
  }
}
```

##### Example 3: Convert datetime to timestamps

Get unix timestamps in various units

**Input:**
```json
{
  "datetime": "2024-03-15T14:30:45.123Z",
  "operation": "to_timestamp"
}
```

**Output:**
```json
{
  "timestamp": null,
  "timestamp_micros": null,
  "timestamp_millis": null,
  "timestamp_nanos": null
}
```

##### Example 4: Convert timestamp to datetime

Convert unix timestamp to readable date

**Input:**
```json
{
  "operation": "from_timestamp",
  "timestamp": null,
  "timestamp_unit": "seconds",
  "to_timezone": "Europe/London"
}
```

**Output:**
```json
{
  "converted": "2024-03-15T14:30:45Z",
  "timezone_info": {
    "abbreviation": "GMT",
    "name": "Europe/London",
    "offset": "+00:00",
    "offset_seconds": 0
  }
}
```

##### Example 5: Convert millisecond timestamp

Convert JavaScript-style millisecond timestamp

**Input:**
```json
{
  "operation": "from_timestamp",
  "timestamp": null,
  "timestamp_unit": "milliseconds",
  "to_timezone": "America/Los_Angeles"
}
```

**Output:**
```json
{
  "converted": "2024-03-15T07:30:45.123-07:00",
  "timezone_info": {
    "abbreviation": "PDT",
    "name": "America/Los_Angeles",
    "offset": "-07:00",
    "offset_seconds": null
  }
}
```

##### Example 6: List European timezones

Find all European timezone identifiers

**Input:**
```json
{
  "operation": "list_timezones",
  "timezone_filter": "Europe"
}
```

**Output:**
```json
{
  "timezones": [
    "Europe/Amsterdam",
    "Europe/Athens",
    "Europe/Berlin",
    "Europe/Brussels",
    "Europe/Dublin",
    "Europe/Helsinki",
    "Europe/Lisbon",
    "Europe/London",
    "Europe/Madrid",
    "Europe/Moscow",
    "Europe/Paris",
    "Europe/Rome",
    "Europe/Stockholm",
    "Europe/Vienna",
    "Europe/Warsaw"
  ]
}
```

##### Example 7: Convert without source timezone

Convert assuming UTC source

**Input:**
```json
{
  "datetime": "2024-03-15T14:30:00Z",
  "operation": "timezone",
  "to_timezone": "Australia/Sydney"
}
```

**Output:**
```json
{
  "converted": "2024-03-16T01:30:00+11:00",
  "timezone_info": {
    "abbreviation": "AEDT",
    "name": "Australia/Sydney",
    "offset": "+11:00",
    "offset_seconds": 39600
  }
}
```

##### Example 8: Timestamp with timezone context

Convert local time to timestamp

**Input:**
```json
{
  "datetime": "2024-03-15 10:30:00",
  "from_timezone": "America/Chicago",
  "operation": "to_timestamp"
}
```

**Output:**
```json
{
  "timestamp": null,
  "timestamp_micros": null,
  "timestamp_millis": null,
  "timestamp_nanos": null
}
```

---

### datetime_format

Format date/time strings with standard formats, custom layouts, relative time, and localization

The datetime_format tool provides flexible date/time formatting capabilities:

Format Types:
1. Standard Formats:
   - RFC3339: "2006-01-02T15:04:05Z07:00" (default)
   - RFC1123: "Mon, 02 Jan 2006 15:04:05 MST"
   - RFC822: "02 Jan 06 15:04 MST"
   - ISO8601: "2006-01-02T15:04:05Z07:00"
   - Kitchen: "3:04PM"
   - Stamp: "Jan _2 15:04:05"
   - UnixDate: "Mon Jan _2 15:04:05 MST 2006"

2. Custom Formats:
   Use Go's time layout syntax with these reference components:
   - Year: "2006" (4 digits), "06" (2 digits)
   - Month: "01" or "1" (number), "Jan" (short name), "January" (full name)
   - Day: "02" or "2" (day of month), "_2" (space-padded)
   - Weekday: "Mon" (short), "Monday" (full)
   - Hour: "15" (24-hour), "03" or "3" (12-hour), "PM" (AM/PM)
   - Minute: "04" or "4"
   - Second: "05" or "5"
   - Millisecond: ".000" (decimal), ".999" (trailing zeros removed)
   - Timezone: "MST" (name), "-0700" (numeric), "Z07:00" (ISO 8601)

3. Relative Time:
   Formats dates as human-readable relative times:
   - "a few seconds ago", "in 3 minutes"
   - "yesterday at 15:04", "tomorrow at 09:30"
   - "3 days ago", "in 2 weeks"
   - "1 month ago", "in 1 year"
   - Optional weekday inclusion: "3 days ago (Monday)"

4. Multiple Formats:
   Output the same date in multiple formats simultaneously.
   Specify format names or custom format strings.

Localization:
Basic support for Spanish (es), French (fr), and German (de):
- Localized month names (full and short)
- Localized weekday names (full and short)  
- AM/PM indicators where applicable

State Integration:
- default_timezone: Default timezone if not specified in input

Examples of Custom Formats:
- "Monday, January 2, 2006": Full date with weekday
- "02/01/2006 15:04:05": European datetime
- "Jan 2 '06 at 3:04pm": Compact friendly format
- "2006-W01-1": ISO week date
- "2006.01.02 AD at 15:04 MST": With era

| Property | Value |
|----------|-------|
| **Category** | datetime |
| **Tags** | datetime, format, localization, relative-time, i18n, display |
| **Version** | 2.0.0 |
| **package** | github.com/lexlapax/go-llms/pkg/agent/builtins/tools/datetime |

#### Output Schema

- **Type**: object

#### Input Schema

- **Type**: object

#### Examples

##### Example 1: Standard RFC3339 format

Format using the default standard format

**Input:**
```json
{
  "datetime": "2024-03-15T14:30:45Z"
}
```

**Output:**
```json
{
  "formatted": "2024-03-15T14:30:45Z"
}
```

##### Example 2: Human-readable custom format

Format date in a friendly, readable way

**Input:**
```json
{
  "custom_format": "Monday, January 2, 2006 at 3:04 PM",
  "datetime": "2024-03-15T14:30:45Z",
  "format_type": "custom"
}
```

**Output:**
```json
{
  "formatted": "Friday, March 15, 2024 at 2:30 PM"
}
```

##### Example 3: Relative time format

Show time relative to now

**Input:**
```json
{
  "datetime": "2024-03-12T10:00:00Z",
  "format_type": "relative",
  "include_weekday": true
}
```

**Output:**
```json
{
  "formatted": "3 days ago (Tuesday)",
  "relative_time": "3 days ago (Tuesday)"
}
```

##### Example 4: Multiple format output

Get the same date in multiple formats

**Input:**
```json
{
  "datetime": "2024-03-15T14:30:45-04:00",
  "format_type": "multiple",
  "formats": [
    "RFC3339",
    "2006-01-02",
    "Kitchen",
    "Monday"
  ]
}
```

**Output:**
```json
{
  "multiple_formats": {
    "2006-01-02": "2024-03-15",
    "Kitchen": "2:30PM",
    "Monday": "Friday",
    "RFC3339": "2024-03-15T14:30:45-04:00"
  }
}
```

##### Example 5: Localized format with Spanish

Format date with Spanish month and weekday names

**Input:**
```json
{
  "custom_format": "Monday, 2 January 2006",
  "datetime": "2024-03-15T14:30:45Z",
  "format_type": "custom",
  "locale": "es"
}
```

**Output:**
```json
{
  "formatted": "Friday, 15 March 2024",
  "localized": {
    "month_name": "marzo",
    "month_name_short": "mar",
    "period": "PM",
    "weekday_name": "viernes",
    "weekday_short": "vie"
  }
}
```

##### Example 6: Format with timezone

Format date in a specific timezone

**Input:**
```json
{
  "custom_format": "Jan 2, 2006 3:04 PM MST",
  "datetime": "2024-03-15T14:30:45Z",
  "format_type": "custom",
  "timezone": "America/New_York"
}
```

**Output:**
```json
{
  "formatted": "Mar 15, 2024 10:30 AM EDT"
}
```

##### Example 7: Kitchen time format

Simple time format for casual display

**Input:**
```json
{
  "datetime": "2024-03-15T14:30:45Z",
  "format_type": "standard",
  "standard_format": "Kitchen"
}
```

**Output:**
```json
{
  "formatted": "2:30PM"
}
```

##### Example 8: Relative time for recent events

Format very recent times

**Input:**
```json
{
  "datetime": "2024-03-15T14:29:30Z",
  "format_type": "relative"
}
```

**Output:**
```json
{
  "formatted": "1 minute ago",
  "relative_time": "1 minute ago"
}
```

---

### datetime_info

Get comprehensive information about a specific date including day/week/month/year info and period boundaries

The datetime_info tool provides comprehensive information about a specific date, including:
- Day information: day of week, day of month, day of year
- Week information: ISO week number and week year
- Month information: month number, name, and days in month
- Quarter information
- Year information including leap year status
- Period boundaries: start/end of week, month, quarter, and year

The tool accepts dates in various formats but prefers RFC3339. It can work with different timezones and allows customizing the week start day.

| Property | Value |
|----------|-------|
| **Category** | datetime |
| **Tags** | datetime, calendar, date-analysis, timezone, week, month, year, iso8601 |
| **Version** | 2.0.0 |
| **package** | github.com/lexlapax/go-llms/pkg/agent/builtins/tools/datetime |

#### Input Schema

- **Type**: object

#### Output Schema

- **Type**: object

#### Examples

##### Example 1: Basic date info

Get information about a specific date

**Input:**
```json
{
  "date": "2024-03-15T10:30:00Z"
}
```

**Output:**
```json
{
  "date": "2024-03-15T10:30:00Z",
  "day_of_month": 15,
  "day_of_week": 5,
  "day_of_week_iso": 5,
  "day_of_week_name": "Friday",
  "day_of_year": 75,
  "is_leap_year": true,
  "month": 3,
  "month_name": "March",
  "year": 2024
}
```

##### Example 2: Date info with timezone

Get date information in a specific timezone

**Input:**
```json
{
  "date": "2024-07-04",
  "timezone": "America/New_York"
}
```

**Output:**
```json
{
  "date": "2024-07-04T00:00:00-04:00",
  "day_of_week": 4,
  "day_of_week_name": "Thursday",
  "month_name": "July",
  "quarter": 3
}
```

##### Example 3: Monday week start

Get date info with Monday as the start of the week

**Input:**
```json
{
  "date": "2024-12-25",
  "week_start_day": 1
}
```

**Output:**
```json
{
  "date": "2024-12-25T00:00:00Z",
  "day_of_week": 3,
  "end_of_week": "2024-12-29T23:59:59.999999999Z",
  "start_of_week": "2024-12-23T00:00:00Z",
  "week_number": 52
}
```

##### Example 4: Leap year check

Check if a year is a leap year and get February info

**Input:**
```json
{
  "date": "2024-02-29"
}
```

**Output:**
```json
{
  "date": "2024-02-29T00:00:00Z",
  "days_in_month": 29,
  "is_leap_year": true,
  "month_name": "February"
}
```

##### Example 5: Quarter boundaries

Get quarter information and boundaries

**Input:**
```json
{
  "date": "2024-05-15"
}
```

**Output:**
```json
{
  "end_of_quarter": "2024-06-30T23:59:59.999999999Z",
  "quarter": 2,
  "start_of_quarter": "2024-04-01T00:00:00Z"
}
```

##### Example 6: ISO week numbering

Get ISO week information for edge cases

**Input:**
```json
{
  "date": "2024-01-01"
}
```

**Output:**
```json
{
  "day_of_year": 1,
  "week_number": 1,
  "week_year": 2024
}
```

##### Example 7: Component extraction

Extract all date components for analysis

**Input:**
```json
{
  "date": "2024-09-30T14:45:30Z"
}
```

**Output:**
```json
{
  "day_of_month": 30,
  "day_of_year": 274,
  "days_in_month": 30,
  "end_of_month": "2024-09-30T23:59:59.999999999Z",
  "end_of_year": "2024-12-31T23:59:59.999999999Z",
  "start_of_year": "2024-01-01T00:00:00Z"
}
```

---

### datetime_now

Get current date/time in various formats with timezone support

Use this tool to get the current date and time in various formats and timezones.

Basic Usage:
- Returns both UTC and local time by default
- Supports any valid IANA timezone (e.g., "America/New_York", "Europe/London", "Asia/Tokyo")
- Can include detailed components, week information, and timestamps

Timezone Support:
- Common US: America/New_York, America/Chicago, America/Denver, America/Los_Angeles
- Europe: Europe/London, Europe/Paris, Europe/Berlin, Europe/Moscow
- Asia: Asia/Tokyo, Asia/Shanghai, Asia/Kolkata, Asia/Dubai
- Pacific: Pacific/Auckland, Australia/Sydney
- Use empty string or omit for UTC and local time only

Format Options:
- Default: RFC3339 (ISO 8601) format
- Custom format using Go time format patterns:
  - "2006-01-02": Date only (YYYY-MM-DD)
  - "15:04:05": Time only (HH:MM:SS)
  - "Mon, 02 Jan 2006 15:04:05 MST": RFC1123 with timezone
  - "January 2, 2006": Human readable date
  - "3:04 PM": 12-hour time
  - "2006-01-02T15:04:05Z07:00": Full ISO 8601

Component Information (include_components):
- Year, Month (number and name), Day
- Hour (24-hour), Minute, Second, Nanosecond
- Weekday (number 0-6, Sunday=0) and name

Week Information (include_week_info):
- ISO week number (1-53)
- ISO day of week (1-7, Monday=1, Sunday=7)
- Day of year (1-366)
- Quarter (1-4)
- Leap year indicator

Timestamps (include_timestamps):
- Unix: Seconds since January 1, 1970 UTC
- UnixMilli: Milliseconds since epoch
- UnixMicro: Microseconds since epoch
- UnixNano: Nanoseconds since epoch

State Integration:
- datetime_default_timezone: Default timezone if not specified
- datetime_default_format: Default format string if not specified

| Property | Value |
|----------|-------|
| **Category** | datetime |
| **Tags** | datetime, current-time, timezone, now, timestamp, utc, local |
| **Version** | 2.0.0 |
| **package** | github.com/lexlapax/go-llms/pkg/agent/builtins/tools/datetime |

#### Output Schema

- **Type**: object

#### Input Schema

- **Type**: object
- **Properties**:
  - **timezone**:
    - **Type**: string
    - **Description**: Timezone to get current time in (e.g., 'America/New_York', 'Europe/London')
  - **format**:
    - **Type**: string
    - **Description**: Custom format string (Go time format)
  - **include_components**:
    - **Type**: boolean
    - **Description**: Include individual date/time components
  - **include_timestamps**:
    - **Type**: boolean
    - **Description**: Include unix timestamps
  - **include_week_info**:
    - **Type**: boolean
    - **Description**: Include week-related information

#### Examples

##### Example 1: Get current time in UTC and local

Basic usage without parameters

**Output:**
```json
{
  "local": "2024-01-15T09:30:45-05:00",
  "utc": "2024-01-15T14:30:45Z"
}
```

##### Example 2: Get time in specific timezone

Request time in New York timezone

**Input:**
```json
{
  "timezone": "America/New_York"
}
```

**Output:**
```json
{
  "local": "2024-01-15T09:30:45-05:00",
  "timezone": "2024-01-15T09:30:45-05:00",
  "timezone_name": "America/New_York",
  "utc": "2024-01-15T14:30:45Z"
}
```

##### Example 3: Get time with components

Include individual date/time components

**Input:**
```json
{
  "include_components": true,
  "timezone": "Europe/London"
}
```

**Output:**
```json
{
  "components": {
    "day": 15,
    "hour": 14,
    "minute": 30,
    "month": 1,
    "month_name": "January",
    "nanosecond": 0,
    "second": 45,
    "weekday": 1,
    "weekday_name": "Monday",
    "year": 2024
  },
  "local": "2024-01-15T09:30:45-05:00",
  "timezone": "2024-01-15T14:30:45Z",
  "timezone_name": "Europe/London",
  "utc": "2024-01-15T14:30:45Z"
}
```

##### Example 4: Get time with week information

Include week-related details

**Input:**
```json
{
  "include_week_info": true
}
```

**Output:**
```json
{
  "local": "2024-03-15T06:00:00-04:00",
  "utc": "2024-03-15T10:00:00Z",
  "week_info": {
    "day_of_week": 5,
    "day_of_year": 75,
    "is_leap_year": true,
    "quarter": 1,
    "week_number": 11
  }
}
```

##### Example 5: Get time with custom format

Format time using custom pattern

**Input:**
```json
{
  "format": "2006年01月02日 15:04:05",
  "timezone": "Asia/Tokyo"
}
```

**Output:**
```json
{
  "formatted": "2024年01月15日 23:30:45",
  "local": "2024-01-15T09:30:45-05:00",
  "timezone": "2024-01-15T23:30:45+09:00",
  "timezone_name": "Asia/Tokyo",
  "utc": "2024-01-15T14:30:45Z"
}
```

##### Example 6: Get Unix timestamps

Include various Unix timestamp formats

**Input:**
```json
{
  "include_timestamps": true
}
```

**Output:**
```json
{
  "local": "2024-01-15T09:30:45-05:00",
  "timestamps": {
    "unix": null,
    "unix_micro": null,
    "unix_milli": null,
    "unix_nano": null
  },
  "utc": "2024-01-15T14:30:45Z"
}
```

##### Example 7: Get all information

Request all available information

**Input:**
```json
{
  "format": "Monday, January 2, 2006 3:04 PM MST",
  "include_components": true,
  "include_timestamps": true,
  "include_week_info": true,
  "timezone": "Australia/Sydney"
}
```

**Output:**
```json
{
  "components": {
    "day": 15,
    "hour": 14,
    "minute": 30,
    "month": 7,
    "month_name": "July",
    "nanosecond": 0,
    "second": 45,
    "weekday": 1,
    "weekday_name": "Monday",
    "year": 2024
  },
  "formatted": "Monday, July 15, 2024 2:30 PM AEST",
  "local": "2024-07-15T00:30:45-04:00",
  "timestamps": {
    "unix": null,
    "unix_micro": null,
    "unix_milli": null,
    "unix_nano": null
  },
  "timezone": "2024-07-15T14:30:45+10:00",
  "timezone_name": "Australia/Sydney",
  "utc": "2024-07-15T04:30:45Z",
  "week_info": {
    "day_of_week": 1,
    "day_of_year": 197,
    "is_leap_year": true,
    "quarter": 3,
    "week_number": 29
  }
}
```

---

### datetime_parse

Parse and validate date/time strings with automatic format detection and relative date support

The datetime_parse tool intelligently parses date/time strings from various formats:

Format Detection:
- Automatic detection of 30+ common date formats (ISO, RFC, US, EU, etc.)
- Custom format specification using Go time layout patterns
- Unix timestamp support (both seconds and milliseconds)
- Relative date parsing (today, tomorrow, yesterday, next Monday, etc.)

Go Time Format Reference:
- Year: "2006" (4 digits), "06" (2 digits)
- Month: "01" or "1" (number), "Jan" (short name), "January" (full name)
- Day: "02" or "2" (day of month), "Mon" (weekday short), "Monday" (weekday full)
- Hour: "15" (24-hour), "03" or "3" (12-hour)
- Minute: "04" or "4"
- Second: "05" or "5"
- AM/PM: "PM" or "pm"
- Timezone: "MST" (name), "-0700" (numeric offset), "Z07:00" (ISO 8601)

Common Format Examples:
- "2006-01-02": ISO date (YYYY-MM-DD)
- "02/01/2006": EU date (DD/MM/YYYY)
- "01/02/2006": US date (MM/DD/YYYY)
- "2006-01-02 15:04:05": Date time with 24-hour format
- "Jan 2, 2006 3:04 PM": Human readable with 12-hour format
- "20060102": Compact date (YYYYMMDD)
- "20060102150405": Compact datetime

Relative Date Support:
- Simple: now, today, yesterday, tomorrow
- Days: "in 3 days", "5 days ago"
- Weeks: "in 2 weeks", "1 week ago", "next week", "last week"
- Months: "in 6 months", "2 months ago", "next month", "last month"
- Years: "next year", "last year"
- Hours: "in 4 hours", "3 hours ago"
- Weekdays: "next Monday", "last Friday"

State Integration:
- default_timezone: Default timezone if not specified in input

Auto-detection tries formats in order of specificity to ensure accurate parsing.

| Property | Value |
|----------|-------|
| **Category** | datetime |
| **Tags** | datetime, parse, validation, format-detection, relative-dates, timestamp |
| **Version** | 2.0.0 |
| **package** | github.com/lexlapax/go-llms/pkg/agent/builtins/tools/datetime |

#### Output Schema

- **Type**: object

#### Input Schema

- **Type**: object

#### Examples

##### Example 1: Parse ISO date

Parse a standard ISO formatted date

**Input:**
```json
{
  "date_string": "2024-03-15"
}
```

**Output:**
```json
{
  "detected_format": "ISO Date",
  "parsed": "2024-03-15T00:00:00Z",
  "unix_timestamp": null,
  "valid": true
}
```

##### Example 2: Parse with custom format

Parse using a specific date format

**Input:**
```json
{
  "date_string": "15/03/2024 14:30",
  "format": "02/01/2006 15:04"
}
```

**Output:**
```json
{
  "detected_format": "custom format",
  "parsed": "2024-03-15T14:30:00Z",
  "unix_timestamp": null,
  "valid": true
}
```

##### Example 3: Parse relative date

Parse natural language relative dates

**Input:**
```json
{
  "date_string": "tomorrow"
}
```

**Output:**
```json
{
  "detected_format": "relative date",
  "parsed": "2024-03-16T00:00:00Z",
  "unix_timestamp": null,
  "valid": true
}
```

##### Example 4: Parse Unix timestamp

Convert Unix timestamp to date

**Input:**
```json
{
  "date_string": "1710460800"
}
```

**Output:**
```json
{
  "detected_format": "Unix timestamp (seconds)",
  "parsed": "2024-03-15T00:00:00Z",
  "unix_timestamp": null,
  "valid": true
}
```

##### Example 5: Parse with timezone

Parse date in specific timezone

**Input:**
```json
{
  "date_string": "2024-03-15 15:30:00",
  "timezone": "America/New_York"
}
```

**Output:**
```json
{
  "detected_format": "DateTime",
  "parsed": "2024-03-15T15:30:00-04:00",
  "unix_timestamp": null,
  "valid": true
}
```

##### Example 6: Parse complex relative date

Parse relative date with reference time

**Input:**
```json
{
  "date_string": "next Monday",
  "reference_time": "2024-03-15T10:00:00Z"
}
```

**Output:**
```json
{
  "detected_format": "relative date",
  "parsed": "2024-03-18T00:00:00Z",
  "unix_timestamp": null,
  "valid": true
}
```

##### Example 7: Parse ambiguous date with auto-detect

Let the tool detect ambiguous date format

**Input:**
```json
{
  "auto_detect": true,
  "date_string": "03/04/2024"
}
```

**Output:**
```json
{
  "detected_format": "US Date",
  "parsed": "2024-03-04T00:00:00Z",
  "unix_timestamp": null,
  "valid": true
}
```

##### Example 8: Parse with validation errors

Handle invalid date string

**Input:**
```json
{
  "date_string": "not a date"
}
```

**Output:**
```json
{
  "detected_format": "",
  "parsed": "",
  "unix_timestamp": null,
  "valid": false,
  "validation_errors": [
    "Unable to parse date string with any known format"
  ]
}
```

---

## feed

### feed_aggregate

Combine multiple feeds into one unified feed with sorting and deduplication

The feed_aggregate tool combines multiple feeds into a single unified feed:

Aggregation Features:
1. Feed Combination:
   - Merges items from all input feeds
   - Preserves all item metadata
   - Maintains feed structure consistency

2. Sorting Options:
   - By date: Published date (falls back to updated date)
   - By title: Alphabetical sorting
   - Ascending or descending order
   - Items without dates sorted to end

3. Duplicate Removal:
   - Detects duplicates by URL (primary)
   - Falls back to content hash (MD5 of title+description+content)
   - Preserves first occurrence

4. Metadata Merging:
   - Combines feed titles: "Aggregated: Feed1, Feed2, ..."
   - Joins descriptions with " | " separator
   - Uses most recent updated timestamp
   - Preserves first feed's other metadata

5. Result Limiting:
   - Apply max_items after sorting and deduplication
   - Useful for creating "top N" feeds

State Integration:
- feed_aggregate_default_sort: Default sort field (date/title)
- feed_aggregate_max_items: Default item limit

Common Use Cases:
- Multi-source news aggregation
- Creating unified podcast feeds
- Combining team/department blogs
- Building curated content feeds
- Cross-source content monitoring

| Property | Value |
|----------|-------|
| **Category** | feed |
| **Tags** | feed, aggregate, combine, merge, sort, deduplicate |
| **Version** | 2.0.0 |
| **package** | github.com/lexlapax/go-llms/pkg/agent/builtins/tools/feed |

#### Output Schema

- **Type**: object

#### Input Schema

- **Type**: object

#### Examples

##### Example 1: Basic feed aggregation

Combine multiple news feeds

**Input:**
```json
{
  "feeds": [
    {
      "items": [
        {
          "link": "https://tech.example.com/ai",
          "published": "2024-03-20T10:00:00Z",
          "title": "AI Breakthrough"
        },
        {
          "link": "https://tech.example.com/phone",
          "published": "2024-03-19T10:00:00Z",
          "title": "New Smartphone"
        }
      ],
      "title": "Tech News"
    },
    {
      "items": [
        {
          "link": "https://science.example.com/mars",
          "published": "2024-03-21T10:00:00Z",
          "title": "Mars Discovery"
        }
      ],
      "title": "Science Daily"
    }
  ]
}
```

**Output:**
```json
{
  "dupes_removed": 0,
  "feed": {
    "items": [
      {
        "link": "https://tech.example.com/phone",
        "published": "2024-03-19T10:00:00Z",
        "title": "New Smartphone"
      },
      {
        "link": "https://tech.example.com/ai",
        "published": "2024-03-20T10:00:00Z",
        "title": "AI Breakthrough"
      },
      {
        "link": "https://science.example.com/mars",
        "published": "2024-03-21T10:00:00Z",
        "title": "Mars Discovery"
      }
    ]
  },
  "source_count": 2,
  "total_items": 3
}
```

##### Example 2: Sort by date descending

Aggregate with most recent items first

**Input:**
```json
{
  "feeds": [
    {
      "items": [
        {
          "published": "2024-01-01T10:00:00Z",
          "title": "Old Post"
        },
        {
          "published": "2024-03-20T10:00:00Z",
          "title": "Recent Post"
        },
        {
          "published": "2024-03-19T10:00:00Z",
          "title": "Yesterday's Post"
        }
      ]
    }
  ],
  "sort_by": "date",
  "sort_descending": true
}
```

**Output:**
```json
{
  "dupes_removed": 0,
  "feed": {
    "items": [
      {
        "published": "2024-03-20T10:00:00Z",
        "title": "Recent Post"
      },
      {
        "published": "2024-03-19T10:00:00Z",
        "title": "Yesterday's Post"
      },
      {
        "published": "2024-01-01T10:00:00Z",
        "title": "Old Post"
      }
    ]
  },
  "source_count": 1,
  "total_items": 3
}
```

##### Example 3: Remove duplicates

Aggregate and remove duplicate items

**Input:**
```json
{
  "feeds": [
    {
      "items": [
        {
          "link": "https://example.com/article1",
          "title": "Shared Article"
        },
        {
          "link": "https://example.com/article2",
          "title": "Unique to A"
        }
      ],
      "title": "Feed A"
    },
    {
      "items": [
        {
          "link": "https://example.com/article1",
          "title": "Shared Article"
        },
        {
          "link": "https://example.com/article3",
          "title": "Unique to B"
        }
      ],
      "title": "Feed B"
    }
  ],
  "remove_dupes": true
}
```

**Output:**
```json
{
  "dupes_removed": 1,
  "feed": {
    "items": [
      {
        "link": "https://example.com/article1",
        "title": "Shared Article"
      },
      {
        "link": "https://example.com/article2",
        "title": "Unique to A"
      },
      {
        "link": "https://example.com/article3",
        "title": "Unique to B"
      }
    ]
  },
  "source_count": 2,
  "total_items": 4
}
```

##### Example 4: Merge metadata

Aggregate with combined feed metadata

**Input:**
```json
{
  "feeds": [
    {
      "description": "Latest technology news",
      "items": [
        {
          "title": "Tech Post"
        }
      ],
      "title": "Tech Blog"
    },
    {
      "description": "Scientific discoveries",
      "items": [
        {
          "title": "Science Post"
        }
      ],
      "title": "Science Blog"
    }
  ],
  "merge_metadata": true
}
```

**Output:**
```json
{
  "feed": {
    "description": "Latest technology news | Scientific discoveries",
    "items": [
      {
        "title": "Tech Post"
      },
      {
        "title": "Science Post"
      }
    ],
    "title": "Aggregated: Tech Blog, Science Blog"
  },
  "source_count": 2,
  "total_items": 2
}
```

##### Example 5: Sort by title

Aggregate and sort alphabetically

**Input:**
```json
{
  "feeds": [
    {
      "items": [
        {
          "title": "Zebra Article"
        },
        {
          "title": "Apple News"
        },
        {
          "title": "Microsoft Update"
        }
      ]
    }
  ],
  "sort_by": "title"
}
```

**Output:**
```json
{
  "feed": {
    "items": [
      {
        "title": "Apple News"
      },
      {
        "title": "Microsoft Update"
      },
      {
        "title": "Zebra Article"
      }
    ]
  },
  "source_count": 1,
  "total_items": 3
}
```

##### Example 6: Limit aggregated items

Create a top-N feed

**Input:**
```json
{
  "feeds": [
    {
      "items": [
        {
          "published": "2024-03-20T10:00:00Z",
          "title": "Post 1"
        },
        {
          "published": "2024-03-19T10:00:00Z",
          "title": "Post 2"
        },
        {
          "published": "2024-03-18T10:00:00Z",
          "title": "Post 3"
        }
      ]
    },
    {
      "items": [
        {
          "published": "2024-03-21T10:00:00Z",
          "title": "Post 4"
        },
        {
          "published": "2024-03-17T10:00:00Z",
          "title": "Post 5"
        }
      ]
    }
  ],
  "max_items": 3,
  "sort_by": "date",
  "sort_descending": true
}
```

**Output:**
```json
{
  "feed": {
    "items": [
      {
        "published": "2024-03-21T10:00:00Z",
        "title": "Post 4"
      },
      {
        "published": "2024-03-20T10:00:00Z",
        "title": "Post 1"
      },
      {
        "published": "2024-03-19T10:00:00Z",
        "title": "Post 2"
      }
    ]
  },
  "source_count": 2,
  "total_items": 5
}
```

##### Example 7: Handle items without dates

Aggregate with mixed date availability

**Input:**
```json
{
  "feeds": [
    {
      "items": [
        {
          "published": "2024-03-20T10:00:00Z",
          "title": "Dated Item"
        },
        {
          "title": "No Date Item"
        },
        {
          "published": "2024-03-19T10:00:00Z",
          "title": "Another Dated"
        }
      ]
    }
  ],
  "sort_by": "date",
  "sort_descending": true
}
```

**Output:**
```json
{
  "feed": {
    "items": [
      {
        "published": "2024-03-20T10:00:00Z",
        "title": "Dated Item"
      },
      {
        "published": "2024-03-19T10:00:00Z",
        "title": "Another Dated"
      },
      {
        "title": "No Date Item"
      }
    ]
  },
  "source_count": 1,
  "total_items": 3
}
```

##### Example 8: Empty feeds handling

Aggregate with some empty feeds

**Input:**
```json
{
  "feeds": [
    {
      "items": [
        {
          "title": "Only Item"
        }
      ],
      "title": "Active Feed"
    },
    {
      "items": null,
      "title": "Empty Feed"
    }
  ]
}
```

**Output:**
```json
{
  "feed": {
    "items": [
      {
        "title": "Only Item"
      }
    ]
  },
  "source_count": 2,
  "total_items": 1
}
```

---

### feed_convert

Convert feeds between RSS, Atom, and JSON Feed formats

The feed_convert tool transforms feeds between different formats:

Supported Conversions:
1. RSS 2.0:
   - Standard RSS format with channel/item structure
   - Wide compatibility with feed readers
   - Best for podcasts and traditional blogs
   - Content included in description field

2. Atom 1.0:
   - IETF standard with better structure
   - Separate content and summary fields
   - Better date/time handling
   - Required unique IDs for entries

3. JSON Feed 1.1:
   - Modern JSON-based format
   - Easy to parse programmatically
   - Native support for attachments
   - Clean separation of content types

Conversion Features:
- Preserves all standard feed elements
- Maps fields appropriately between formats
- Handles enclosures/attachments conversion
- Maintains author information
- Converts dates to appropriate formats
- Optional pretty-printing for readability

State Integration:
- feed_convert_default_format: Default target format
- feed_convert_pretty_print: Default pretty print setting

Common Use Cases:
- Feed format migration
- Cross-platform compatibility
- API format requirements
- Feed validation and testing
- Format modernization

| Property | Value |
|----------|-------|
| **Category** | feed |
| **Tags** | feed, convert, transform, rss, atom, json, format |
| **Version** | 2.0.0 |
| **package** | github.com/lexlapax/go-llms/pkg/agent/builtins/tools/feed |

#### Input Schema

- **Type**: object

#### Output Schema

- **Type**: object

#### Examples

##### Example 1: RSS to JSON Feed

Convert traditional RSS to modern JSON format

**Input:**
```json
{
  "feed": {
    "description": "Latest tech news",
    "items": [
      {
        "content": "\u003cp\u003eFull article content...\u003c/p\u003e",
        "description": "Brief summary",
        "id": "post-1",
        "link": "https://example.com/post-1",
        "published": "2024-03-15T10:00:00Z",
        "title": "New Technology"
      }
    ],
    "link": "https://example.com",
    "title": "Tech Blog"
  },
  "pretty": true,
  "target_type": "json"
}
```

**Output:**
```json
{
  "content": "{\n  \"version\": \"https://jsonfeed.org/version/1.1\",\n  \"title\": \"Tech Blog\",\n  \"home_page_url\": \"https://example.com\",\n  \"description\": \"Latest tech news\",\n  \"items\": [\n    {\n      \"id\": \"post-1\",\n      \"title\": \"New Technology\",\n      \"url\": \"https://example.com/post-1\",\n      \"content_html\": \"\u003cp\u003eFull article content...\u003c/p\u003e\",\n      \"date_published\": \"2024-03-15T10:00:00Z\"\n    }\n  ]\n}",
  "content_type": "application/feed+json",
  "format": "json"
}
```

##### Example 2: Convert to Atom

Convert any feed to Atom format

**Input:**
```json
{
  "feed": {
    "author": {
      "email": "news@example.com",
      "name": "News Team"
    },
    "items": [
      {
        "categories": [
          "urgent",
          "world"
        ],
        "description": "Important update",
        "id": "news-1",
        "link": "https://news.example.com/1",
        "published": "2024-03-15T12:00:00Z",
        "title": "Breaking News"
      }
    ],
    "link": "https://news.example.com",
    "title": "News Feed"
  },
  "include_content": false,
  "target_type": "atom"
}
```

**Output:**
```json
{
  "content": "\u003c?xml version=\"1.0\" encoding=\"UTF-8\"?\u003e\n\u003cfeed xmlns=\"http://www.w3.org/2005/Atom\"\u003e...",
  "content_type": "application/atom+xml",
  "format": "atom"
}
```

##### Example 3: Convert to RSS

Convert modern formats to classic RSS

**Input:**
```json
{
  "feed": {
    "copyright": "© 2024 Example",
    "items": [
      {
        "description": "Our first episode",
        "enclosures": [
          {
            "length": 15000000,
            "type": "audio/mpeg",
            "url": "https://podcast.example.com/ep1.mp3"
          }
        ],
        "id": "episode-1",
        "link": "https://podcast.example.com/1",
        "published": "2024-03-01T09:00:00Z",
        "title": "Episode 1: Introduction"
      }
    ],
    "language": "en-us",
    "link": "https://podcast.example.com",
    "title": "Podcast Feed"
  },
  "target_type": "rss"
}
```

**Output:**
```json
{
  "content": "\u003c?xml version=\"1.0\" encoding=\"UTF-8\"?\u003e\n\u003crss version=\"2.0\"\u003e...",
  "content_type": "application/rss+xml",
  "format": "rss"
}
```

##### Example 4: Pretty print JSON

Convert with human-readable formatting

**Input:**
```json
{
  "feed": {
    "items": [
      {
        "id": "1",
        "title": "Item 1"
      }
    ],
    "title": "Simple Feed"
  },
  "pretty": true,
  "target_type": "json"
}
```

**Output:**
```json
{
  "content": "{\n  \"version\": \"https://jsonfeed.org/version/1.1\",\n  \"title\": \"Simple Feed\",\n  \"items\": [\n    {\n      \"id\": \"1\",\n      \"title\": \"Item 1\"\n    }\n  ]\n}",
  "content_type": "application/feed+json",
  "format": "json"
}
```

##### Example 5: Convert with full content

Include complete article content

**Input:**
```json
{
  "feed": {
    "items": [
      {
        "content": "\u003carticle\u003e\u003cp\u003eThis is the complete article with multiple paragraphs...\u003c/p\u003e\u003c/article\u003e",
        "description": "Summary only",
        "id": "post-1",
        "title": "Full Article"
      }
    ],
    "title": "Blog"
  },
  "include_content": true,
  "target_type": "atom"
}
```

**Output:**
```json
{
  "content": "\u003c?xml version=\"1.0\" encoding=\"UTF-8\"?\u003e\n\u003cfeed xmlns=\"http://www.w3.org/2005/Atom\"\u003e...",
  "content_type": "application/atom+xml",
  "format": "atom"
}
```

##### Example 6: Author preservation

Convert while maintaining author information

**Input:**
```json
{
  "feed": {
    "author": {
      "name": "Editorial Team"
    },
    "items": [
      {
        "author": {
          "email": "john@example.com",
          "name": "John Smith"
        },
        "id": "1",
        "title": "Post by John"
      }
    ],
    "title": "Team Blog"
  },
  "target_type": "json"
}
```

**Output:**
```json
{
  "content": "{\"version\":\"https://jsonfeed.org/version/1.1\",...}",
  "content_type": "application/feed+json",
  "format": "json"
}
```

##### Example 7: Attachment conversion

Convert feeds with media attachments

**Input:**
```json
{
  "feed": {
    "items": [
      {
        "enclosures": [
          {
            "length": 50000000,
            "type": "video/mp4",
            "url": "https://example.com/video.mp4"
          }
        ],
        "id": "video-1",
        "title": "Tutorial Video"
      }
    ],
    "title": "Video Feed"
  },
  "target_type": "json"
}
```

**Output:**
```json
{
  "content": "{...\"attachments\":[{\"url\":\"https://example.com/video.mp4\",\"mime_type\":\"video/mp4\",\"size_in_bytes\":50000000}]...}",
  "content_type": "application/feed+json",
  "format": "json"
}
```

##### Example 8: Date format handling

Convert with proper date formatting

**Input:**
```json
{
  "feed": {
    "items": [
      {
        "id": "event-1",
        "published": "2024-03-20T09:00:00Z",
        "title": "Upcoming Event",
        "updated": "2024-03-21T10:00:00Z"
      }
    ],
    "title": "Event Feed",
    "updated": "2024-03-15T14:30:00Z"
  },
  "target_type": "rss"
}
```

**Output:**
```json
{
  "content": "\u003c?xml version=\"1.0\" encoding=\"UTF-8\"?\u003e\n\u003crss version=\"2.0\"\u003e...\u003cpubDate\u003eMon, 20 Mar 2024 09:00:00 +0000\u003c/pubDate\u003e...",
  "content_type": "application/rss+xml",
  "format": "rss"
}
```

---

### feed_discover

Automatically discover feed URLs from web pages with authentication support

The feed_discover tool automatically finds RSS, Atom, and JSON feed URLs from web pages:

Discovery Methods:
1. HTML Link Tags:
   - Searches for <link rel="alternate"> tags
   - Detects type attributes like application/rss+xml, application/atom+xml
   - Extracts feed titles from link tags
   - Resolves relative URLs to absolute

2. Common Feed Paths:
   - Checks standard feed locations like /feed, /rss, /atom.xml
   - Verifies feed existence with HEAD requests
   - Validates content types

3. Auto-Discovery:
   - Follows feed auto-discovery standards
   - Supports RSS 2.0, Atom 1.0, and JSON Feed
   - Handles multiple feeds per page

Authentication Support:
- Automatic detection from state (api_key, bearer_token, etc.)
- Manual auth configuration for protected sites
- Support for API key, Bearer, Basic, OAuth2, and custom auth
- Auth applied to both discovery and verification requests

State Integration:
- feed_discover_timeout: Default timeout in seconds
- feed_discover_max_size: Default max response size
- feed_discover_follow_redirects: Default redirect behavior
- Authentication auto-detected from state keys

Common Use Cases:
- Find all feeds on a blog or news site
- Discover podcast feeds
- Find JSON feeds for modern applications
- Aggregate feeds from multiple sites
- Verify feed availability before subscription

| Property | Value |
|----------|-------|
| **Category** | feed |
| **Tags** | feed, discover, rss, atom, json, auto-discovery, web |
| **Version** | 2.0.0 |
| **package** | github.com/lexlapax/go-llms/pkg/agent/builtins/tools/feed |

#### Input Schema

- **Type**: object

#### Output Schema

- **Type**: object

#### Examples

##### Example 1: Basic feed discovery

Discover feeds from a blog homepage

**Input:**
```json
{
  "url": "https://blog.example.com"
}
```

**Output:**
```json
{
  "feeds": [
    {
      "source": "link_tag",
      "title": "Example Blog RSS Feed",
      "type": "rss",
      "url": "https://blog.example.com/feed"
    },
    {
      "source": "link_tag",
      "title": "Example Blog Atom Feed",
      "type": "atom",
      "url": "https://blog.example.com/atom.xml"
    }
  ]
}
```

##### Example 2: Discovery with authentication

Discover feeds from a protected site

**Input:**
```json
{
  "auth": {
    "token": "your-access-token",
    "type": "bearer"
  },
  "url": "https://private.example.com"
}
```

**Output:**
```json
{
  "feeds": [
    {
      "source": "link_tag",
      "type": "json",
      "url": "https://private.example.com/api/feed.json"
    }
  ]
}
```

##### Example 3: Discovery with timeout

Set custom timeout for slow sites

**Input:**
```json
{
  "timeout": 60,
  "url": "https://slow.example.com"
}
```

**Output:**
```json
{
  "feeds": [
    {
      "source": "common_path",
      "type": "rss",
      "url": "https://slow.example.com/rss"
    }
  ]
}
```

##### Example 4: Multiple feed types

Discover various feed formats

**Input:**
```json
{
  "url": "https://news.example.com"
}
```

**Output:**
```json
{
  "feeds": [
    {
      "source": "link_tag",
      "title": "News RSS 2.0",
      "type": "rss",
      "url": "https://news.example.com/rss.xml"
    },
    {
      "source": "link_tag",
      "title": "News Atom 1.0",
      "type": "atom",
      "url": "https://news.example.com/feed.atom"
    },
    {
      "source": "link_tag",
      "title": "News JSON Feed",
      "type": "json",
      "url": "https://news.example.com/feed.json"
    }
  ]
}
```

##### Example 5: No redirects follow

Discover without following redirects

**Input:**
```json
{
  "follow_redirects": false,
  "url": "https://redirect.example.com"
}
```

**Output:**
```json
{
  "error": "HTTP error: 301 Moved Permanently",
  "feeds": null
}
```

##### Example 6: Common path discovery

Find feeds via common URL patterns

**Input:**
```json
{
  "url": "https://simple.example.com"
}
```

**Output:**
```json
{
  "feeds": [
    {
      "source": "common_path",
      "type": "rss",
      "url": "https://simple.example.com/feed"
    },
    {
      "source": "common_path",
      "type": "rss",
      "url": "https://simple.example.com/rss"
    }
  ]
}
```

##### Example 7: Custom headers

Include custom headers in discovery request

**Input:**
```json
{
  "headers": {
    "Accept": "text/html,application/xml",
    "X-Client-ID": "my-app"
  },
  "url": "https://api.example.com"
}
```

**Output:**
```json
{
  "feeds": [
    {
      "source": "link_tag",
      "type": "rss",
      "url": "https://api.example.com/v1/feed"
    }
  ]
}
```

##### Example 8: Size-limited discovery

Limit response size for large pages

**Input:**
```json
{
  "max_size": 1048576,
  "url": "https://huge.example.com"
}
```

**Output:**
```json
{
  "feeds": [
    {
      "source": "link_tag",
      "title": "Main Feed",
      "type": "rss",
      "url": "https://huge.example.com/feed.xml"
    }
  ]
}
```

---

### feed_extract

Extract specific fields from feed items for structured data analysis

The feed_extract tool provides selective field extraction from feed data:

Field Extraction:
1. Basic Fields:
   - id, title, description, content, link
   - published, updated (formatted as RFC3339)
   - categories (array of strings)

2. Nested Fields:
   - author.name, author.email, author.url
   - Individual author fields or full author object

3. Media Fields:
   - enclosures (array of media attachments)
   - Each enclosure contains url, type, length

Advanced Features:
- Field Flattening: Convert author.name to author_name
- Metadata Inclusion: Extract feed-level information
- Item Limiting: Control number of items processed
- State Integration: Default fields and limits from state

State Integration:
- feed_extract_max_items: Default maximum items
- feed_extract_default_fields: Default field list

Common Use Cases:
- Data transformation for analytics
- Creating simplified feed summaries
- Extracting specific content types
- Preparing data for external systems
- Content migration and archiving

| Property | Value |
|----------|-------|
| **Category** | feed |
| **Tags** | feed, extract, transform, data, parsing, analysis |
| **Version** | 2.0.0 |
| **package** | github.com/lexlapax/go-llms/pkg/agent/builtins/tools/feed |

#### Input Schema

- **Type**: object

#### Output Schema

- **Type**: object

#### Examples

##### Example 1: Basic field extraction

Extract titles and links from feed items

**Input:**
```json
{
  "feed": {
    "items": [
      {
        "id": "post-1",
        "link": "https://blog.example.com/post-1",
        "title": "Latest Technology Trends"
      },
      {
        "id": "post-2",
        "link": "https://blog.example.com/post-2",
        "title": "AI Developments"
      }
    ],
    "title": "Tech Blog"
  },
  "fields": [
    "title",
    "link"
  ]
}
```

**Output:**
```json
{
  "count": 2,
  "data": [
    {
      "link": "https://blog.example.com/post-1",
      "title": "Latest Technology Trends"
    },
    {
      "link": "https://blog.example.com/post-2",
      "title": "AI Developments"
    }
  ],
  "fields": [
    "title",
    "link"
  ]
}
```

##### Example 2: Extract with author information

Extract nested author fields

**Input:**
```json
{
  "feed": {
    "items": [
      {
        "author": {
          "email": "john@example.com",
          "name": "John Doe"
        },
        "published": "2024-03-15T10:00:00Z",
        "title": "Article by John"
      }
    ]
  },
  "fields": [
    "title",
    "author.name",
    "author.email",
    "published"
  ]
}
```

**Output:**
```json
{
  "count": 1,
  "data": [
    {
      "author.email": "john@example.com",
      "author.name": "John Doe",
      "published": "2024-03-15T10:00:00Z",
      "title": "Article by John"
    }
  ],
  "fields": [
    "title",
    "author.name",
    "author.email",
    "published"
  ]
}
```

##### Example 3: Flattened field extraction

Extract and flatten nested fields

**Input:**
```json
{
  "feed": {
    "items": [
      {
        "author": {
          "name": "Jane Smith",
          "url": "https://jane.example.com"
        },
        "title": "Article with Author"
      }
    ]
  },
  "fields": [
    "title",
    "author.name",
    "author.url"
  ],
  "flatten": true
}
```

**Output:**
```json
{
  "count": 1,
  "data": [
    {
      "author_name": "Jane Smith",
      "author_url": "https://jane.example.com",
      "title": "Article with Author"
    }
  ],
  "fields": [
    "title",
    "author.name",
    "author.url"
  ]
}
```

##### Example 4: Extract with feed metadata

Include feed-level information in results

**Input:**
```json
{
  "feed": {
    "description": "Latest news updates",
    "items": [
      {
        "id": "news-1",
        "title": "Breaking News"
      }
    ],
    "language": "en",
    "link": "https://news.example.com",
    "title": "News Feed"
  },
  "fields": [
    "title",
    "id"
  ],
  "include_metadata": true
}
```

**Output:**
```json
{
  "count": 1,
  "data": [
    {
      "id": "news-1",
      "title": "Breaking News"
    }
  ],
  "fields": [
    "title",
    "id"
  ],
  "metadata": {
    "description": "Latest news updates",
    "language": "en",
    "link": "https://news.example.com",
    "title": "News Feed"
  }
}
```

##### Example 5: Limited item extraction

Extract from first N items only

**Input:**
```json
{
  "feed": {
    "items": [
      {
        "link": "link1",
        "title": "Post 1"
      },
      {
        "link": "link2",
        "title": "Post 2"
      },
      {
        "link": "link3",
        "title": "Post 3"
      },
      {
        "link": "link4",
        "title": "Post 4"
      }
    ]
  },
  "fields": [
    "title",
    "link"
  ],
  "max_items": 2
}
```

**Output:**
```json
{
  "count": 2,
  "data": [
    {
      "link": "link1",
      "title": "Post 1"
    },
    {
      "link": "link2",
      "title": "Post 2"
    }
  ],
  "fields": [
    "title",
    "link"
  ]
}
```

##### Example 6: Extract media enclosures

Extract podcast or media attachments

**Input:**
```json
{
  "feed": {
    "items": [
      {
        "enclosures": [
          {
            "length": 25000000,
            "type": "audio/mpeg",
            "url": "https://podcast.example.com/ep42.mp3"
          }
        ],
        "title": "Episode 42"
      }
    ]
  },
  "fields": [
    "title",
    "enclosures"
  ]
}
```

**Output:**
```json
{
  "count": 1,
  "data": [
    {
      "enclosures": [
        {
          "length": 25000000,
          "type": "audio/mpeg",
          "url": "https://podcast.example.com/ep42.mp3"
        }
      ],
      "title": "Episode 42"
    }
  ],
  "fields": [
    "title",
    "enclosures"
  ]
}
```

##### Example 7: Categories and tags extraction

Extract taxonomies and content classification

**Input:**
```json
{
  "feed": {
    "items": [
      {
        "categories": [
          "technology",
          "programming",
          "web"
        ],
        "content": "Article about web development...",
        "title": "Tech Article"
      }
    ]
  },
  "fields": [
    "title",
    "categories",
    "content"
  ]
}
```

**Output:**
```json
{
  "count": 1,
  "data": [
    {
      "categories": [
        "technology",
        "programming",
        "web"
      ],
      "content": "Article about web development...",
      "title": "Tech Article"
    }
  ],
  "fields": [
    "title",
    "categories",
    "content"
  ]
}
```

##### Example 8: Full author object extraction

Extract complete author information

**Input:**
```json
{
  "feed": {
    "items": [
      {
        "author": {
          "email": "sarah@university.edu",
          "name": "Dr. Sarah Wilson",
          "url": "https://university.edu/faculty/sarah"
        },
        "title": "Expert Opinion"
      }
    ]
  },
  "fields": [
    "title",
    "author"
  ]
}
```

**Output:**
```json
{
  "count": 1,
  "data": [
    {
      "author": {
        "email": "sarah@university.edu",
        "name": "Dr. Sarah Wilson",
        "url": "https://university.edu/faculty/sarah"
      },
      "title": "Expert Opinion"
    }
  ],
  "fields": [
    "title",
    "author"
  ]
}
```

---

### feed_fetch

Fetches and parses feeds in RSS, Atom, or JSON Feed format with authentication support

The feed_fetch tool retrieves and parses web feeds in various formats:

Feed Formats Supported:
1. RSS 2.0:
   - Most common feed format
   - XML-based with channel and item elements
   - Supports enclosures for podcasts/media

2. Atom:
   - IETF standard feed format
   - More structured than RSS
   - Better date/time handling

3. JSON Feed:
   - Modern JSON-based format
   - Easier to parse than XML
   - Native support for content types

Unified Output Format:
- All feeds converted to consistent structure
- Normalized field names across formats
- Proper date/time parsing
- Author information extraction
- Media enclosure support

Authentication Support:
- Automatic detection from state (api_key, bearer_token, etc.)
- Manual auth configuration for protected feeds
- Support for API key, Bearer, Basic, OAuth2, and custom auth
- Works with subscription-based feeds

Conditional Requests:
- ETag support for bandwidth efficiency
- If-Modified-Since header support
- Returns 304 Not Modified when unchanged
- Preserves caching headers in response

State Integration:
- feed_fetch_default_timeout: Default timeout in seconds
- feed_fetch_user_agent: Default User-Agent string
- Authentication auto-detected from state keys

Common Use Cases:
- News aggregation and monitoring
- Podcast feed parsing
- Blog post syndication
- Content change detection
- Feed validation and testing

| Property | Value |
|----------|-------|
| **Category** | feed |
| **Tags** | feed, rss, atom, json, syndication, news, podcast |
| **Version** | 2.0.0 |
| **package** | github.com/lexlapax/go-llms/pkg/agent/builtins/tools/feed |

#### Input Schema

- **Type**: object

#### Output Schema

- **Type**: object

#### Examples

##### Example 1: Basic RSS fetch

Fetch a public RSS feed

**Input:**
```json
{
  "url": "https://example.com/rss"
}
```

**Output:**
```json
{
  "feed": {
    "description": "A blog about examples",
    "items": [
      {
        "description": "This is the first post",
        "id": "https://example.com/post1",
        "link": "https://example.com/post1",
        "published": "2024-03-15T10:00:00Z",
        "title": "First Post"
      }
    ],
    "link": "https://example.com",
    "title": "Example Blog"
  },
  "format": "RSS2",
  "status": 200
}
```

##### Example 2: Fetch with authentication

Fetch a protected feed

**Input:**
```json
{
  "auth": {
    "token": "your-access-token",
    "type": "bearer"
  },
  "url": "https://premium.example.com/feed"
}
```

**Output:**
```json
{
  "feed": {
    "items": [
      {
        "content": "Full premium content here...",
        "id": "premium-1",
        "title": "Exclusive Article"
      }
    ],
    "title": "Premium Content Feed"
  },
  "format": "Atom",
  "status": 200
}
```

##### Example 3: Conditional fetch with ETag

Check if feed has changed

**Input:**
```json
{
  "etag": "W/\"123456789\"",
  "url": "https://news.example.com/feed"
}
```

**Output:**
```json
{
  "headers": {
    "Cache-Control": "max-age=300",
    "ETag": "W/\"123456789\""
  },
  "not_modified": true,
  "status": 304
}
```

##### Example 4: Fetch with item limit

Get only recent items

**Input:**
```json
{
  "max_items": 5,
  "url": "https://blog.example.com/feed"
}
```

**Output:**
```json
{
  "feed": {
    "items": [
      {
        "id": "1",
        "title": "Latest Post"
      },
      {
        "id": "2",
        "title": "Yesterday's Post"
      },
      {
        "id": "3",
        "title": "Previous Post"
      },
      {
        "id": "4",
        "title": "Older Post"
      },
      {
        "id": "5",
        "title": "Fifth Post"
      }
    ],
    "title": "Tech Blog"
  },
  "format": "RSS2",
  "status": 200
}
```

##### Example 5: JSON Feed fetch

Fetch a modern JSON feed

**Input:**
```json
{
  "url": "https://modern.example.com/feed.json"
}
```

**Output:**
```json
{
  "feed": {
    "author": {
      "name": "John Doe",
      "url": "https://modern.example.com/about"
    },
    "description": "A blog using JSON Feed",
    "items": [
      {
        "content": "Here's why JSON feeds are awesome...",
        "id": "2024-03-15",
        "tags": [
          "json",
          "feeds",
          "web"
        ],
        "title": "JSON Feeds are Great"
      }
    ],
    "link": "https://modern.example.com",
    "title": "Modern Blog"
  },
  "format": "JSONFeed",
  "status": 200
}
```

##### Example 6: Podcast feed with enclosures

Fetch podcast RSS with media files

**Input:**
```json
{
  "url": "https://podcast.example.com/rss"
}
```

**Output:**
```json
{
  "feed": {
    "items": [
      {
        "description": "Discussion about feed formats",
        "enclosures": [
          {
            "length": 25000000,
            "type": "audio/mpeg",
            "url": "https://podcast.example.com/episodes/42.mp3"
          }
        ],
        "id": "episode-42",
        "title": "Episode 42: Feed Processing"
      }
    ],
    "title": "Tech Podcast"
  },
  "format": "RSS2",
  "status": 200
}
```

##### Example 7: Custom headers

Fetch with custom HTTP headers

**Input:**
```json
{
  "headers": {
    "Accept": "application/rss+xml",
    "X-API-Version": "2.0"
  },
  "url": "https://api.example.com/feed"
}
```

**Output:**
```json
{
  "feed": {
    "items": null,
    "title": "API Feed"
  },
  "format": "RSS2",
  "status": 200
}
```

##### Example 8: Timeout handling

Set custom timeout for slow feeds

**Input:**
```json
{
  "timeout": 60,
  "url": "https://slow.example.com/feed"
}
```

**Output:**
```json
{
  "feed": {
    "items": null,
    "title": "Slow Feed"
  },
  "format": "Atom",
  "status": 200
}
```

---

### feed_filter

Filter feed items based on multiple criteria including keywords, dates, authors, and categories

The feed_filter tool provides powerful filtering capabilities for feed data:

Filter Types:
1. Keyword Filtering:
   - Searches in title, description, and content
   - Case-insensitive matching
   - Partial match support

2. Date Range Filtering:
   - Filter by published date (falls back to updated date)
   - Supports after/before date ranges
   - RFC3339 date format required

3. Author Filtering:
   - Matches author name field
   - Case-insensitive partial matching
   - Useful for finding posts by specific contributors

4. Category Filtering:
   - Matches against item categories/tags
   - Case-insensitive partial matching
   - Helps find topical content

Matching Modes:
- match_all=false (default): Items matching ANY criteria are included
- match_all=true: Items must match ALL specified criteria
- Date filters are always applied regardless of matching mode

State Integration:
- feed_filter_max_items: Default maximum items limit
- feed_filter_match_all: Default matching mode

Common Use Cases:
- Recent content: Filter by date range
- Topic search: Filter by keywords and categories
- Author archives: Filter by specific authors
- Content curation: Combine multiple filters
- Feed sampling: Use max_items to limit results

| Property | Value |
|----------|-------|
| **Category** | feed |
| **Tags** | feed, filter, search, query, date, keyword, content |
| **Version** | 2.0.0 |
| **package** | github.com/lexlapax/go-llms/pkg/agent/builtins/tools/feed |

#### Output Schema

- **Type**: object

#### Input Schema

- **Type**: object

#### Examples

##### Example 1: Filter by keywords

Find items containing specific keywords

**Input:**
```json
{
  "feed": {
    "items": [
      {
        "description": "How artificial intelligence is transforming medicine",
        "published": "2024-03-15T10:00:00Z",
        "title": "AI Revolution in Healthcare"
      },
      {
        "description": "Latest findings on global warming",
        "published": "2024-03-14T10:00:00Z",
        "title": "Climate Change Report"
      },
      {
        "content": "Deep dive into artificial intelligence companies",
        "published": "2024-03-13T10:00:00Z",
        "title": "Tech Stocks Analysis"
      }
    ]
  },
  "keywords": [
    "artificial intelligence",
    "AI"
  ]
}
```

**Output:**
```json
{
  "filtered_out": 1,
  "items": [
    {
      "description": "How artificial intelligence is transforming medicine",
      "published": "2024-03-15T10:00:00Z",
      "title": "AI Revolution in Healthcare"
    },
    {
      "content": "Deep dive into artificial intelligence companies",
      "published": "2024-03-13T10:00:00Z",
      "title": "Tech Stocks Analysis"
    }
  ],
  "total_items": 3
}
```

##### Example 2: Filter by date range

Get items from the last week

**Input:**
```json
{
  "after": "2024-03-15T00:00:00Z",
  "feed": {
    "items": [
      {
        "published": "2024-03-20T10:00:00Z",
        "title": "Today's News"
      },
      {
        "published": "2024-03-10T10:00:00Z",
        "title": "Last Week's Update"
      },
      {
        "published": "2024-01-01T10:00:00Z",
        "title": "Old Article"
      }
    ]
  }
}
```

**Output:**
```json
{
  "filtered_out": 2,
  "items": [
    {
      "published": "2024-03-20T10:00:00Z",
      "title": "Today's News"
    }
  ],
  "total_items": 3
}
```

##### Example 3: Filter by author

Find posts by specific authors

**Input:**
```json
{
  "authors": [
    "John"
  ],
  "feed": {
    "items": [
      {
        "author": {
          "name": "John Smith"
        },
        "title": "Post by John"
      },
      {
        "author": {
          "name": "Jane Doe"
        },
        "title": "Post by Jane"
      },
      {
        "author": {
          "name": "John Williams"
        },
        "title": "Another John Post"
      }
    ]
  }
}
```

**Output:**
```json
{
  "filtered_out": 1,
  "items": [
    {
      "author": {
        "name": "John Smith"
      },
      "title": "Post by John"
    },
    {
      "author": {
        "name": "John Williams"
      },
      "title": "Another John Post"
    }
  ],
  "total_items": 3
}
```

##### Example 4: Filter by categories

Find items in specific categories

**Input:**
```json
{
  "categories": [
    "tech"
  ],
  "feed": {
    "items": [
      {
        "categories": [
          "technology",
          "news"
        ],
        "title": "Tech News"
      },
      {
        "categories": [
          "sports",
          "news"
        ],
        "title": "Sports Update"
      },
      {
        "categories": [
          "technology",
          "tutorial"
        ],
        "title": "Tech Tutorial"
      }
    ]
  }
}
```

**Output:**
```json
{
  "filtered_out": 1,
  "items": [
    {
      "categories": [
        "technology",
        "news"
      ],
      "title": "Tech News"
    },
    {
      "categories": [
        "technology",
        "tutorial"
      ],
      "title": "Tech Tutorial"
    }
  ],
  "total_items": 3
}
```

##### Example 5: Complex filter with match_all

Apply multiple filters with ALL matching

**Input:**
```json
{
  "after": "2024-03-01T00:00:00Z",
  "authors": [
    "John"
  ],
  "categories": [
    "tech"
  ],
  "feed": {
    "items": [
      {
        "author": {
          "name": "John Smith"
        },
        "categories": [
          "technology",
          "ai"
        ],
        "content": "Artificial intelligence applications",
        "published": "2024-03-20T10:00:00Z",
        "title": "AI in Tech by John"
      },
      {
        "author": {
          "name": "Jane Doe"
        },
        "categories": [
          "news"
        ],
        "content": "Latest AI developments",
        "published": "2024-03-20T10:00:00Z",
        "title": "AI News"
      },
      {
        "author": {
          "name": "John Smith"
        },
        "categories": [
          "technology"
        ],
        "content": "Technology trends",
        "published": "2024-01-01T10:00:00Z",
        "title": "Old Tech Post by John"
      }
    ]
  },
  "keywords": [
    "AI"
  ],
  "match_all": true
}
```

**Output:**
```json
{
  "filtered_out": 2,
  "items": [
    {
      "author": {
        "name": "John Smith"
      },
      "categories": [
        "technology",
        "ai"
      ],
      "content": "Artificial intelligence applications",
      "published": "2024-03-20T10:00:00Z",
      "title": "AI in Tech by John"
    }
  ],
  "total_items": 3
}
```

##### Example 6: Date range with before and after

Filter items within a specific date range

**Input:**
```json
{
  "after": "2024-02-01T00:00:00Z",
  "before": "2024-04-01T00:00:00Z",
  "feed": {
    "items": [
      {
        "published": "2024-01-15T10:00:00Z",
        "title": "January Post"
      },
      {
        "published": "2024-02-15T10:00:00Z",
        "title": "February Post"
      },
      {
        "published": "2024-03-15T10:00:00Z",
        "title": "March Post"
      },
      {
        "published": "2024-04-15T10:00:00Z",
        "title": "April Post"
      }
    ]
  }
}
```

**Output:**
```json
{
  "filtered_out": 2,
  "items": [
    {
      "published": "2024-02-15T10:00:00Z",
      "title": "February Post"
    },
    {
      "published": "2024-03-15T10:00:00Z",
      "title": "March Post"
    }
  ],
  "total_items": 4
}
```

##### Example 7: Limited results

Filter with max_items limit

**Input:**
```json
{
  "categories": [
    "news"
  ],
  "feed": {
    "items": [
      {
        "categories": [
          "news"
        ],
        "title": "News 1"
      },
      {
        "categories": [
          "news"
        ],
        "title": "News 2"
      },
      {
        "categories": [
          "news"
        ],
        "title": "News 3"
      },
      {
        "categories": [
          "news"
        ],
        "title": "News 4"
      },
      {
        "categories": [
          "news"
        ],
        "title": "News 5"
      }
    ]
  },
  "max_items": 3
}
```

**Output:**
```json
{
  "filtered_out": 0,
  "items": [
    {
      "categories": [
        "news"
      ],
      "title": "News 1"
    },
    {
      "categories": [
        "news"
      ],
      "title": "News 2"
    },
    {
      "categories": [
        "news"
      ],
      "title": "News 3"
    }
  ],
  "total_items": 5
}
```

##### Example 8: No date items handling

Handle items without dates gracefully

**Input:**
```json
{
  "after": "2024-03-14T00:00:00Z",
  "feed": {
    "items": [
      {
        "published": "2024-03-15T10:00:00Z",
        "title": "Dated Post"
      },
      {
        "title": "Undated Post"
      },
      {
        "published": "2024-03-16T10:00:00Z",
        "title": "Another Dated"
      }
    ]
  }
}
```

**Output:**
```json
{
  "filtered_out": 1,
  "items": [
    {
      "published": "2024-03-15T10:00:00Z",
      "title": "Dated Post"
    },
    {
      "published": "2024-03-16T10:00:00Z",
      "title": "Another Dated"
    }
  ],
  "total_items": 3
}
```

---

## file

### file_delete

Safely deletes files and directories with confirmation options

Use this tool to safely delete files and directories with multiple confirmation options.

IMPORTANT: This is a DESTRUCTIVE operation that permanently removes data.

Features:
- Multiple safety mechanisms to prevent accidental deletion
- Support for both files and directories
- Recursive directory deletion with safeguards
- Critical system path protection
- Confirmation requirements for dangerous operations
- Path access control via state configuration

Parameters:
- path: File or directory to delete (required)
- force: Skip safety checks (dangerous!)
- recursive: Delete directories and all contents
- require_confirm: Must match the path for deletion to proceed

Safety Mechanisms:
1. Critical system paths are protected by default
2. Non-empty directories require recursive=true
3. Confirmation can be required via require_confirm
4. State configuration can enforce confirmation
5. Path restrictions via allowed/restricted lists

Confirmation Requirements:
- Set require_confirm to the exact path or filename
- State can enforce confirmation via file_require_delete_confirmation
- Non-empty directories need force=true or confirmation

Critical Paths Protected:
- Root directories (/, C:\)
- System directories (/bin, /etc, C:\Windows)
- User home directory
- Program directories

State Configuration:
- file_restricted_paths: Array of paths that cannot be deleted
- file_allowed_paths: Array of allowed path prefixes
- file_require_delete_confirmation: Always require confirmation
- file_use_trash: Prefer trash/recycle bin (note: not implemented)

Best Practices:
- Always use require_confirm for important deletions
- Test with a dry run first (check path without deleting)
- Use recursive=true only when necessary
- Avoid force=true unless absolutely certain
- Consider backups before deletion

Error Handling:
- Non-existent paths return success with deleted=false
- Permission errors are reported clearly
- Directory not empty errors suggest recursive=true

| Property | Value |
|----------|-------|
| **Category** | file |
| **Tags** | filesystem, delete, remove, cleanup |
| **Version** | 1.0.0 |
| **package** | github.com/lexlapax/go-llms/pkg/agent/builtins/tools/file |

#### Input Schema

- **Type**: object
- **Required**: path
- **Properties**:
  - **force**:
    - **Type**: boolean
    - **Description**: Force deletion without safety checks (use with caution)
  - **path**:
    - **Type**: string
    - **Description**: Path to the file or directory to delete
  - **recursive**:
    - **Type**: boolean
    - **Description**: Delete directories and all their contents
  - **require_confirm**:
    - **Type**: string
    - **Description**: Safety confirmation - must match the path being deleted

#### Output Schema

- **Type**: object
- **Required**: path, deleted, was_directory
- **Properties**:
  - **was_directory**:
    - **Type**: boolean
    - **Description**: Whether the deleted item was a directory
  - **deleted**:
    - **Type**: boolean
    - **Description**: Whether the deletion was successful
  - **message**:
    - **Type**: string
    - **Description**: Status message or reason for failure
  - **path**:
    - **Type**: string
    - **Description**: Absolute path that was processed

#### Examples

##### Example 1: Delete a simple file

Remove a temporary file

**Input:**
```json
{
  "path": "/tmp/temp_file.txt"
}
```

**Output:**
```json
{
  "deleted": true,
  "message": "File deleted successfully",
  "path": "/tmp/temp_file.txt",
  "was_directory": false
}
```

##### Example 2: Delete with confirmation

Delete important file with confirmation

**Input:**
```json
{
  "path": "/home/user/important.db",
  "require_confirm": "important.db"
}
```

**Output:**
```json
{
  "deleted": true,
  "message": "File deleted successfully",
  "path": "/home/user/important.db",
  "was_directory": false
}
```

##### Example 3: Delete empty directory

Remove an empty directory

**Input:**
```json
{
  "path": "/tmp/empty_dir"
}
```

**Output:**
```json
{
  "deleted": true,
  "message": "Directory deleted successfully (0 items removed)",
  "path": "/tmp/empty_dir",
  "was_directory": true
}
```

##### Example 4: Delete directory with contents

Recursively delete directory and contents

**Input:**
```json
{
  "path": "/tmp/old_project",
  "recursive": true,
  "require_confirm": "old_project"
}
```

**Output:**
```json
{
  "deleted": true,
  "message": "Directory deleted successfully (42 items removed)",
  "path": "/tmp/old_project",
  "was_directory": true
}
```

##### Example 5: Failed confirmation

Deletion blocked by wrong confirmation

**Input:**
```json
{
  "path": "/home/user/data.db",
  "require_confirm": "wrong_name"
}
```

**Output:**
```json
{
  "deleted": false,
  "message": "Confirmation mismatch: expected '/home/user/data.db' or 'data.db', got 'wrong_name'",
  "path": "/home/user/data.db",
  "was_directory": false
}
```

##### Example 6: Non-empty directory without recursive

Attempt to delete non-empty directory

**Input:**
```json
{
  "path": "/tmp/full_dir"
}
```

**Output:**
```json
{
  "deleted": false,
  "message": "Directory is not empty (15 items). Use recursive=true to delete contents",
  "path": "/tmp/full_dir",
  "was_directory": true
}
```

##### Example 7: Critical path protection

Attempt to delete system directory

**Input:**
```json
{
  "path": "/etc"
}
```

**Output:**
```json
{
  "deleted": false,
  "message": "Cannot delete critical system directory. Use force=true to override (dangerous!)",
  "path": "/etc",
  "was_directory": true
}
```

---

### file_list

Lists files and directories with filtering options

Use this tool to list files and directories with extensive filtering options.

Features:
- Fast directory enumeration
- Flexible pattern matching (glob patterns)
- Recursive directory traversal
- Size-based filtering (min/max)
- Date-based filtering (modified before/after)
- Multiple sort options
- Hidden file control via state

Parameters:
- path: Directory to list (required)
- pattern: Glob pattern (e.g., *.txt, test_*, *.{jpg,png})
- recursive: Search subdirectories (default: false)
- include_dirs: Include directories in results (default: false)
- include_files: Include files in results (default: true)
- min_size/max_size: Filter by file size in bytes
- modified_after/before: Filter by modification time (RFC3339)
- sort_by: Sort by name, size, or modified (default: name)
- sort_reverse: Reverse sort order
- max_results: Limit number of results

Pattern Matching:
- Supports standard glob patterns
- * matches any sequence of characters
- ? matches any single character
- [abc] matches any character in brackets
- [a-z] matches any character in range
- {jpg,png} matches any of the alternatives

Size Filtering:
- Sizes are in bytes
- min_size: 1048576 = 1MB
- max_size: 10485760 = 10MB
- Only applies to files, not directories

Date Filtering:
- Use RFC3339 format: 2024-01-15T10:30:00Z
- Times are compared in UTC
- modified_after: Include files modified after this time
- modified_before: Include files modified before this time

State Configuration:
- file_list_show_hidden: Show hidden files (starting with .)
- file_list_default_sort: Default sort field
- file_list_max_results: Default max results
- file_restricted_paths: Array of restricted paths
- file_allowed_paths: Array of allowed path prefixes

Sorting:
- name: Alphabetical by filename (case-insensitive)
- size: By file size (smallest first)
- modified: By modification time (oldest first)
- Use sort_reverse: true to reverse order

Performance:
- Non-recursive listing is very fast
- Recursive searches may take time for large trees
- Progress events emitted every 100 items
- Context cancellation supported

| Property | Value |
|----------|-------|
| **Category** | file |
| **Tags** | filesystem, directory, list, search |
| **Version** | 1.0.0 |
| **package** | github.com/lexlapax/go-llms/pkg/agent/builtins/tools/file |

#### Input Schema

- **Type**: object
- **Required**: path
- **Properties**:
  - **modified_after**:
    - **Type**: string
    - **Description**: Only files modified after this time (RFC3339)
  - **path**:
    - **Type**: string
    - **Description**: Directory path to list
  - **recursive**:
    - **Type**: boolean
    - **Description**: Search subdirectories recursively
  - **sort_by**:
    - **Type**: string
    - **Description**: Sort results by: name, size, or modified
  - **sort_reverse**:
    - **Type**: boolean
    - **Description**: Reverse sort order
  - **include_dirs**:
    - **Type**: boolean
    - **Description**: Include directories in results
  - **include_files**:
    - **Type**: boolean
    - **Description**: Include files in results (default: true)
  - **min_size**:
    - **Type**: number
    - **Description**: Minimum file size in bytes
  - **modified_before**:
    - **Type**: string
    - **Description**: Only files modified before this time (RFC3339)
  - **pattern**:
    - **Type**: string
    - **Description**: Glob pattern to match files (e.g., '*.txt', 'test_*')
  - **max_results**:
    - **Type**: number
    - **Description**: Maximum number of results to return
  - **max_size**:
    - **Type**: number
    - **Description**: Maximum file size in bytes

#### Output Schema

- **Type**: object
- **Required**: files, total_count, filtered_out, search_path
- **Properties**:
  - **total_count**:
    - **Type**: number
    - **Description**: Total number of items scanned
  - **files**:
    - **Type**: array
    - **Description**: List of files and directories found
  - **filtered_out**:
    - **Type**: number
    - **Description**: Number of items filtered out
  - **pattern**:
    - **Type**: string
    - **Description**: Pattern used for filtering (if any)
  - **search_path**:
    - **Type**: string
    - **Description**: Absolute path that was searched

#### Examples

##### Example 1: List current directory

List all files in current directory

**Input:**
```json
{
  "path": "."
}
```

**Output:**
```json
{
  "files": [
    {
      "extension": "md",
      "is_dir": false,
      "mode": "-rw-r--r--",
      "modified_time": "2024-01-15T10:00:00Z",
      "name": "README.md",
      "path": "./README.md",
      "size": 1024
    },
    {
      "extension": "go",
      "is_dir": false,
      "mode": "-rw-r--r--",
      "modified_time": "2024-01-15T11:00:00Z",
      "name": "main.go",
      "path": "./main.go",
      "size": 2048
    },
    {
      "extension": "mod",
      "is_dir": false,
      "mode": "-rw-r--r--",
      "modified_time": "2024-01-15T09:00:00Z",
      "name": "go.mod",
      "path": "./go.mod",
      "size": 256
    }
  ],
  "filtered_out": 0,
  "search_path": "/home/user/project",
  "total_count": 3
}
```

##### Example 2: Find Go files recursively

Search for all Go source files

**Input:**
```json
{
  "path": ".",
  "pattern": "*.go",
  "recursive": true
}
```

**Output:**
```json
{
  "files": [
    {
      "is_dir": false,
      "name": "main.go",
      "path": "./main.go",
      "size": 2048
    },
    {
      "is_dir": false,
      "name": "utils.go",
      "path": "./pkg/utils.go",
      "size": 1024
    },
    {
      "is_dir": false,
      "name": "test.go",
      "path": "./test/test.go",
      "size": 512
    }
  ],
  "filtered_out": 7,
  "pattern": "*.go",
  "search_path": "/home/user/project",
  "total_count": 10
}
```

##### Example 3: List large files

Find files larger than 10MB

**Input:**
```json
{
  "min_size": 10485760,
  "path": "/home/user/downloads",
  "recursive": true,
  "sort_by": "size",
  "sort_reverse": true
}
```

**Output:**
```json
{
  "files": [
    {
      "name": "video.mp4",
      "size": 104857600
    },
    {
      "name": "backup.zip",
      "size": 52428800
    },
    {
      "name": "dataset.csv",
      "size": 20971520
    }
  ],
  "filtered_out": 47,
  "search_path": "/home/user/downloads",
  "total_count": 50
}
```

##### Example 4: Recent files

Find files modified in last 24 hours

**Input:**
```json
{
  "modified_after": "2024-01-14T10:30:00Z",
  "path": ".",
  "recursive": true,
  "sort_by": "modified",
  "sort_reverse": true
}
```

**Output:**
```json
{
  "files": [
    {
      "modified_time": "2024-01-15T15:30:00Z",
      "name": "report.pdf"
    },
    {
      "modified_time": "2024-01-15T14:00:00Z",
      "name": "data.json"
    },
    {
      "modified_time": "2024-01-15T12:00:00Z",
      "name": "notes.txt"
    }
  ],
  "filtered_out": 97,
  "search_path": "/home/user/project",
  "total_count": 100
}
```

##### Example 5: List directories only

Show only subdirectories

**Input:**
```json
{
  "include_dirs": true,
  "include_files": false,
  "path": "."
}
```

**Output:**
```json
{
  "files": [
    {
      "is_dir": true,
      "name": "src",
      "path": "./src",
      "size": 0
    },
    {
      "is_dir": true,
      "name": "test",
      "path": "./test",
      "size": 0
    },
    {
      "is_dir": true,
      "name": "docs",
      "path": "./docs",
      "size": 0
    }
  ],
  "filtered_out": 2,
  "search_path": "/home/user/project",
  "total_count": 5
}
```

##### Example 6: Image files by extension

Find all image files

**Input:**
```json
{
  "max_results": 100,
  "path": "/home/user/pictures",
  "pattern": "*.{jpg,jpeg,png,gif}",
  "recursive": true
}
```

**Output:**
```json
{
  "files": [
    {
      "extension": "jpg",
      "name": "photo1.jpg"
    },
    {
      "extension": "png",
      "name": "screenshot.png"
    },
    {
      "extension": "gif",
      "name": "animation.gif"
    }
  ],
  "filtered_out": 100,
  "pattern": "*.{jpg,jpeg,png,gif}",
  "search_path": "/home/user/pictures",
  "total_count": 500
}
```

##### Example 7: Complex filter

Recent large Python files

**Input:**
```json
{
  "min_size": 1024,
  "modified_after": "2024-01-01T00:00:00Z",
  "path": ".",
  "pattern": "*.py",
  "recursive": true,
  "sort_by": "size",
  "sort_reverse": true
}
```

**Output:**
```json
{
  "files": [
    {
      "modified_time": "2024-01-10T10:00:00Z",
      "name": "main.py",
      "size": 8192
    },
    {
      "modified_time": "2024-01-05T10:00:00Z",
      "name": "utils.py",
      "size": 4096
    }
  ],
  "filtered_out": 48,
  "total_count": 50
}
```

---

### file_move

Moves or renames files and directories

Use this tool to move or rename files and directories safely.

Features:
- Atomic moves within same filesystem (instant)
- Cross-device transfers (copy-then-delete)
- Directory and file support
- Overwrite protection with explicit flag
- Parent directory creation
- Attribute preservation
- Event tracking for operation progress

Parameters:
- source: Source file or directory path (required)
- destination: Target path or directory (required)
- overwrite: Allow overwriting existing files (optional, default false)
- create_dirs: Create parent directories if needed (optional)
- preserve_attrs: Keep file permissions and timestamps (optional)

Move Operations:
1. Rename: Same directory, different name
2. Move: Different directory, same or different name
3. Cross-device: Automatic copy-then-delete for different filesystems

Destination Behavior:
- If destination is a directory: moves source into it
- If destination is a file path: renames/moves to that path
- Parent directories must exist unless create_dirs is true

Safety Features:
- Won't overwrite without explicit permission
- Checks for same source and destination
- Validates paths before operation
- Restricted path checking via state

State Configuration:
- file.restricted_paths: Array of paths to block
- file.allow_overwrite: Default overwrite permission
- file.prefer_copy_delete: Force copy-delete method

Cross-Device Moves:
- Automatically detected when rename fails
- Only supported for files, not directories
- Preserves content, optionally preserves attributes
- Atomic within filesystem limits

Events Emitted:
- file_move.start: Operation beginning
- file_move.checking_source: Validating source
- file_move.checking_destination: Validating destination
- file_move.moving: Starting move operation
- file_move.copying: Cross-device copy in progress
- file_move.completed: Operation finished

Best Practices:
- Always check if destination exists first
- Use overwrite=true carefully
- Enable preserve_attrs for important files
- Create parent directories when organizing files
- Consider using rename for same-directory operations

| Property | Value |
|----------|-------|
| **Category** | file |
| **Tags** | filesystem, move, rename, transfer |
| **Version** | 1.0.0 |
| **package** | github.com/lexlapax/go-llms/pkg/agent/builtins/tools/file |

#### Input Schema

- **Type**: object
- **Required**: source, destination
- **Properties**:
  - **destination**:
    - **Type**: string
    - **Description**: Destination path (can be a new name or directory)
  - **overwrite**:
    - **Type**: boolean
    - **Description**: Overwrite existing destination file
  - **preserve_attrs**:
    - **Type**: boolean
    - **Description**: Preserve file permissions and timestamps
  - **source**:
    - **Type**: string
    - **Description**: Source file or directory path
  - **create_dirs**:
    - **Type**: boolean
    - **Description**: Create parent directories if they don't exist

#### Output Schema

- **Type**: object
- **Required**: source, destination, moved, was_rename, was_cross_device
- **Properties**:
  - **was_cross_device**:
    - **Type**: boolean
    - **Description**: True if file was moved across filesystems
  - **was_rename**:
    - **Type**: boolean
    - **Description**: True if operation was a rename in same directory
  - **destination**:
    - **Type**: string
    - **Description**: Absolute destination path where file was moved
  - **message**:
    - **Type**: string
    - **Description**: Status message or error description
  - **moved**:
    - **Type**: boolean
    - **Description**: Whether the move operation succeeded
  - **source**:
    - **Type**: string
    - **Description**: Absolute source path that was moved

#### Examples

##### Example 1: Simple rename

Rename a file in the same directory

**Input:**
```json
{
  "destination": "report.txt",
  "source": "document.txt"
}
```

**Output:**
```json
{
  "destination": "/home/user/report.txt",
  "message": "Successfully moved",
  "moved": true,
  "source": "/home/user/document.txt",
  "was_cross_device": false,
  "was_rename": true
}
```

##### Example 2: Move to directory

Move file to another directory

**Input:**
```json
{
  "destination": "/home/user/documents/",
  "source": "/tmp/download.pdf"
}
```

**Output:**
```json
{
  "destination": "/home/user/documents/download.pdf",
  "message": "Successfully moved (cross-device)",
  "moved": true,
  "source": "/tmp/download.pdf",
  "was_cross_device": true,
  "was_rename": false
}
```

##### Example 3: Move with new name

Move and rename in one operation

**Input:**
```json
{
  "create_dirs": true,
  "destination": "projects/final-report.txt",
  "source": "temp/draft.txt"
}
```

**Output:**
```json
{
  "destination": "/home/user/projects/final-report.txt",
  "message": "Successfully moved",
  "moved": true,
  "source": "/home/user/temp/draft.txt",
  "was_cross_device": false,
  "was_rename": false
}
```

##### Example 4: Overwrite existing file

Replace existing file with move

**Input:**
```json
{
  "destination": "config.json",
  "overwrite": true,
  "source": "new-config.json"
}
```

**Output:**
```json
{
  "destination": "/app/config.json",
  "message": "Successfully moved",
  "moved": true,
  "source": "/app/new-config.json",
  "was_cross_device": false,
  "was_rename": true
}
```

##### Example 5: Move directory

Relocate entire directory

**Input:**
```json
{
  "destination": "workspace/active-projects/",
  "source": "old-location/project"
}
```

**Output:**
```json
{
  "destination": "/home/user/workspace/active-projects/project",
  "message": "Successfully moved",
  "moved": true,
  "source": "/home/user/old-location/project",
  "was_cross_device": false,
  "was_rename": false
}
```

##### Example 6: Failed move - exists

Destination already exists

**Input:**
```json
{
  "destination": "existing.txt",
  "source": "update.txt"
}
```

**Output:**
```json
{
  "destination": "/home/user/existing.txt",
  "message": "Destination already exists. Use overwrite=true to replace",
  "moved": false,
  "source": "/home/user/update.txt",
  "was_cross_device": false,
  "was_rename": true
}
```

##### Example 7: Preserve attributes

Move with timestamp preservation

**Input:**
```json
{
  "destination": "/home/user/backups/",
  "preserve_attrs": true,
  "source": "/mnt/backup/archive.tar"
}
```

**Output:**
```json
{
  "destination": "/home/user/backups/archive.tar",
  "message": "Successfully moved (cross-device)",
  "moved": true,
  "source": "/mnt/backup/archive.tar",
  "was_cross_device": true,
  "was_rename": false
}
```

---

### file_read

Reads file contents with support for large files, line ranges, and metadata

Use this tool to read file contents with advanced features.

Features:
- Automatic encoding detection (UTF-8 or binary)
- Line range support for reading specific portions
- File size limits to prevent memory issues
- Metadata retrieval (size, permissions, timestamps)
- Path access control via state configuration
- Progress events for large file operations

Parameters:
- path: File path to read (required)
- max_size: Maximum bytes to read (optional, default 10MB or from state)
- line_start: Start reading from this line (optional, 1-based)
- line_end: Stop reading at this line (optional, inclusive)
- include_meta: Include file metadata (optional, default false)

Line Range Reading:
- Use line_start/line_end for large log files
- Lines are 1-based (first line is 1)
- Only text files support line ranges
- Binary files ignore line parameters

State Configuration:
- file_read_max_size: Default max size in bytes
- file_restricted_paths: Array of paths to block
- file_allowed_paths: Array of allowed path prefixes
- file_preferred_encoding: Override encoding detection

Security:
- Path restrictions can be enforced via state
- Symlinks are followed (be careful with access control)
- Binary files are detected and marked

Performance:
- Uses buffered reading for efficiency
- Streams large files instead of loading all at once
- Emits progress events during read
- Context cancellation supported

| Property | Value |
|----------|-------|
| **Category** | file |
| **Tags** | file, read, filesystem, text, binary |
| **Version** | 1.0.0 |
| **package** | github.com/lexlapax/go-llms/pkg/agent/builtins/tools/file |

#### Input Schema

- **Type**: object
- **Required**: path
- **Properties**:
  - **line_end**:
    - **Type**: number
    - **Description**: Stop reading at this line number (inclusive)
  - **line_start**:
    - **Type**: number
    - **Description**: Start reading from this line number (1-based)
  - **max_size**:
    - **Type**: number
    - **Description**: Maximum bytes to read (0 = unlimited, default: 10MB)
  - **path**:
    - **Type**: string
    - **Description**: The path to the file to read
  - **include_meta**:
    - **Type**: boolean
    - **Description**: Include file metadata in the response

#### Output Schema

- **Type**: object
- **Required**: content, encoding, is_binary
- **Properties**:
  - **metadata**:
    - **Type**: object
    - **Description**: File metadata (if include_meta is true)
  - **warnings**:
    - **Type**: array
    - **Description**: Any warnings generated during read
  - **content**:
    - **Type**: string
    - **Description**: The file content
  - **encoding**:
    - **Type**: string
    - **Description**: Detected file encoding (utf-8 or binary)
  - **is_binary**:
    - **Type**: boolean
    - **Description**: Whether the file is binary
  - **lines**:
    - **Type**: number
    - **Description**: Number of lines read (for text files)

#### Examples

##### Example 1: Read text file

Read a simple text file

**Input:**
```json
{
  "path": "/home/user/config.json"
}
```

**Output:**
```json
{
  "content": "{\"api_key\": \"secret\", \"port\": 8080}",
  "encoding": "utf-8",
  "is_binary": false,
  "lines": 1
}
```

##### Example 2: Read with metadata

Get file content and metadata

**Input:**
```json
{
  "include_meta": true,
  "path": "/var/log/app.log"
}
```

**Output:**
```json
{
  "content": "2024-01-15 10:00:00 INFO Starting application\n2024-01-15 10:00:01 INFO Connected to database",
  "encoding": "utf-8",
  "is_binary": false,
  "lines": 2,
  "metadata": {
    "absolute_path": "/var/log/app.log",
    "extension": ".log",
    "is_dir": false,
    "mod_time": "2024-01-15T10:00:01Z",
    "mode": "-rw-r--r--",
    "size": 2048
  }
}
```

##### Example 3: Read specific lines

Read lines 100-150 from a large log file

**Input:**
```json
{
  "line_end": 150,
  "line_start": 100,
  "path": "/var/log/system.log"
}
```

**Output:**
```json
{
  "content": "[100 lines of log content from line 100 to 150]",
  "encoding": "utf-8",
  "is_binary": false,
  "lines": 51
}
```

##### Example 4: Read with size limit

Read large file with size constraint

**Input:**
```json
{
  "max_size": 1048576,
  "path": "/data/large_dataset.csv"
}
```

**Output:**
```json
{
  "content": "[First 1MB of CSV data]",
  "encoding": "utf-8",
  "is_binary": false,
  "lines": 5000,
  "warnings": [
    "File truncated at 1048576 bytes"
  ]
}
```

##### Example 5: Binary file detection

Read a binary file

**Input:**
```json
{
  "path": "/usr/bin/ls"
}
```

**Output:**
```json
{
  "content": "[Binary content - may appear garbled]",
  "encoding": "binary",
  "is_binary": true
}
```

##### Example 6: Handle missing file

Attempt to read non-existent file

**Input:**
```json
{
  "path": "/tmp/nonexistent.txt"
}
```

**Output:**
```json
{
  "error": "error opening file: open /tmp/nonexistent.txt: no such file or directory"
}
```

##### Example 7: Path restriction

Blocked by security policy

**Input:**
```json
{
  "path": "/etc/shadow"
}
```

**Output:**
```json
{
  "error": "access denied: path /etc/shadow is restricted"
}
```

---

### file_search

Searches for patterns in file contents

Use this tool to search for text patterns within files, similar to grep.

Features:
- Plain text or regex pattern matching
- Case-sensitive or case-insensitive search
- File filtering by name patterns (glob)
- Recursive directory searching
- Context lines before/after matches
- Binary file detection and skipping
- Progress tracking for large searches

Parameters:
- path: File or directory to search (required)
- pattern: Search pattern (required)
- file_pattern: Filter files by name (e.g., *.txt, *.go)
- is_regex: Treat pattern as regular expression
- case_sensitive: Enable case-sensitive matching
- recursive: Search subdirectories
- max_results: Limit number of matches (default: 1000)
- context_lines: Show N lines before/after matches
- include_line_numbers: Show line numbers (default: true)

Pattern Matching:
- Plain text: Exact substring matching
- Regex: Full regular expression support
- Case-insensitive: Controlled by case_sensitive flag
- Special regex chars: . * + ? ^ $ [] {} () | \

File Filtering:
- Use file_pattern for glob matching
- Examples: *.txt, test_*, *.{js,ts}
- Applied to filename only, not path

Context Lines:
- Shows surrounding lines for better understanding
- context_before: Lines preceding the match
- context_after: Lines following the match
- Useful for understanding code context

State Configuration:
- file_access_restrictions: Restricted paths
- file_search_max_results: Default result limit
- file_search_case_sensitive: Default case sensitivity
- file_search_encoding: Default file encoding

Performance:
- Streams files to handle large files efficiently
- Binary files automatically skipped
- Progress reporting for directory searches
- Cancellable via context

Events Emitted:
- Tool call/result events
- Progress events during search
- file_search_complete with statistics
- Error events for invalid patterns

Best Practices:
- Use file_pattern to narrow search scope
- Enable recursive for project-wide searches
- Use context_lines for code searches
- Set reasonable max_results to avoid overload
- Use regex for complex pattern matching

| Property | Value |
|----------|-------|
| **Category** | file |
| **Tags** | filesystem, search, grep, find, pattern |
| **Version** | 1.0.0 |
| **package** | github.com/lexlapax/go-llms/pkg/agent/builtins/tools/file |

#### Input Schema

- **Type**: object
- **Required**: path, pattern
- **Properties**:
  - **max_results**:
    - **Type**: number
    - **Description**: Maximum number of matches to return
  - **path**:
    - **Type**: string
    - **Description**: File or directory path to search
  - **pattern**:
    - **Type**: string
    - **Description**: Search pattern (plain text or regex)
  - **recursive**:
    - **Type**: boolean
    - **Description**: Search subdirectories recursively
  - **is_regex**:
    - **Type**: boolean
    - **Description**: Treat pattern as regular expression
  - **case_sensitive**:
    - **Type**: boolean
    - **Description**: Perform case-sensitive search
  - **context_lines**:
    - **Type**: number
    - **Description**: Number of context lines before/after matches
  - **file_pattern**:
    - **Type**: string
    - **Description**: File name pattern to filter (e.g., '*.txt')
  - **include_line_numbers**:
    - **Type**: boolean
    - **Description**: Include line numbers in results

#### Output Schema

- **Type**: object
- **Required**: matches, total_matches, files_searched, pattern, search_path
- **Properties**:
  - **total_matches**:
    - **Type**: number
    - **Description**: Total number of matches found
  - **files_searched**:
    - **Type**: number
    - **Description**: Number of files searched
  - **matches**:
    - **Type**: array
    - **Description**: List of matches found
  - **pattern**:
    - **Type**: string
    - **Description**: The search pattern used
  - **search_path**:
    - **Type**: string
    - **Description**: The path that was searched

#### Examples

##### Example 1: Simple text search

Find occurrences of a word

**Input:**
```json
{
  "path": "document.txt",
  "pattern": "TODO"
}
```

**Output:**
```json
{
  "files_searched": 1,
  "matches": [
    {
      "file": "/home/user/document.txt",
      "line": "// TODO: Implement error handling",
      "line_number": 15,
      "match_end": 7,
      "match_start": 3
    }
  ],
  "pattern": "TODO",
  "search_path": "/home/user/document.txt",
  "total_matches": 1
}
```

##### Example 2: Recursive code search

Find function definitions in Go files

**Input:**
```json
{
  "file_pattern": "*.go",
  "path": "src/",
  "pattern": "func main",
  "recursive": true
}
```

**Output:**
```json
{
  "files_searched": 45,
  "matches": [
    {
      "file": "/project/src/cmd/app/main.go",
      "line": "func main() {",
      "line_number": 10,
      "match_end": 9,
      "match_start": 0
    },
    {
      "file": "/project/src/examples/demo.go",
      "line": "func main() {",
      "line_number": 8,
      "match_end": 9,
      "match_start": 0
    }
  ],
  "pattern": "func main",
  "search_path": "/project/src",
  "total_matches": 2
}
```

##### Example 3: Regex with context

Find error patterns with surrounding context

**Input:**
```json
{
  "context_lines": 2,
  "is_regex": true,
  "path": "app.log",
  "pattern": "ERROR.*database"
}
```

**Output:**
```json
{
  "files_searched": 1,
  "matches": [
    {
      "context_after": [
        "[2024-01-15 10:30:15] WARN: Retrying connection in 5s",
        "[2024-01-15 10:30:20] INFO: Connection retry attempt 1"
      ],
      "context_before": [
        "[2024-01-15 10:30:14] INFO: Attempting database connection",
        "[2024-01-15 10:30:14] DEBUG: Using connection string: ..."
      ],
      "file": "/var/log/app.log",
      "line": "[2024-01-15 10:30:15] ERROR: database connection failed",
      "line_number": 156,
      "match_end": 38,
      "match_start": 23
    }
  ],
  "pattern": "ERROR.*database",
  "search_path": "/var/log/app.log",
  "total_matches": 1
}
```

##### Example 4: Case-insensitive search

Find variables regardless of case

**Input:**
```json
{
  "case_sensitive": false,
  "path": "config.ini",
  "pattern": "api_key"
}
```

**Output:**
```json
{
  "files_searched": 1,
  "matches": [
    {
      "file": "/app/config.ini",
      "line": "API_KEY=secret123",
      "line_number": 5,
      "match_end": 7,
      "match_start": 0
    },
    {
      "file": "/app/config.ini",
      "line": "backup_api_key=secret456",
      "line_number": 12,
      "match_end": 14,
      "match_start": 7
    }
  ],
  "pattern": "api_key",
  "search_path": "/app/config.ini",
  "total_matches": 2
}
```

##### Example 5: Limited results

Search with result limit

**Input:**
```json
{
  "max_results": 10,
  "path": "logs/",
  "pattern": "INFO",
  "recursive": true
}
```

**Output:**
```json
{
  "files_searched": 3,
  "matches": null,
  "pattern": "INFO",
  "search_path": "/app/logs",
  "total_matches": 10
}
```

##### Example 6: Multiple file types

Search specific file extensions

**Input:**
```json
{
  "file_pattern": "*.{md,txt,rst}",
  "path": "docs/",
  "pattern": "deprecated",
  "recursive": true
}
```

**Output:**
```json
{
  "files_searched": 15,
  "matches": [
    {
      "file": "/project/docs/api.md",
      "line": "**Deprecated**: This endpoint will be removed in v2.0",
      "line_number": 45
    },
    {
      "file": "/project/docs/changelog.txt",
      "line": "- Deprecated old authentication method",
      "line_number": 23
    }
  ],
  "pattern": "deprecated",
  "search_path": "/project/docs",
  "total_matches": 2
}
```

##### Example 7: Complex regex pattern

Extract email addresses

**Input:**
```json
{
  "is_regex": true,
  "path": "contacts.csv",
  "pattern": "[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}"
}
```

**Output:**
```json
{
  "files_searched": 1,
  "matches": [
    {
      "file": "/data/contacts.csv",
      "line": "John Doe,john.doe@example.com,Marketing",
      "line_number": 3,
      "match_end": 29,
      "match_start": 9
    }
  ],
  "pattern": "[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}",
  "search_path": "/data/contacts.csv",
  "total_matches": 1
}
```

---

### file_write

Writes content to files with atomic operations, append mode, and backup support

Use this tool to write content to files with advanced features.

Features:
- Atomic write operations for data integrity
- Append mode for adding to existing files
- Automatic parent directory creation
- File backup before overwriting
- Path access control via state configuration
- Progress events for operation tracking

Parameters:
- path: File path to write (required)
- content: Content to write (required)
- mode: File permissions in octal (optional, default 0644)
- append: Append to file instead of overwrite (optional)
- create_dirs: Create parent directories if needed (optional)
- atomic: Use atomic write operation (optional)
- backup: Create backup of existing file (optional)

Atomic Write:
- Writes to temporary file first
- Renames to target path on success
- Prevents partial writes on failure
- Recommended for critical files

Backup Feature:
- Creates timestamped backup before overwrite
- Format: filename.backup-YYYYMMDD-HHMMSS.ext
- Only backs up existing files
- Can be auto-enabled via state

State Configuration:
- file_restricted_paths: Array of paths to block
- file_allowed_paths: Array of allowed path prefixes
- file_default_permissions: Default file mode
- file_auto_backup: Enable automatic backups

Security:
- Path restrictions enforced via state
- Parent directory creation requires explicit flag
- Atomic writes prevent corruption
- Proper permission setting

Performance:
- Atomic writes may be slower for large files
- Direct writes are fastest
- Progress events emitted for operations
- Context cancellation supported

| Property | Value |
|----------|-------|
| **Category** | file |
| **Tags** | file, write, filesystem, save, create |
| **Version** | 1.0.0 |
| **package** | github.com/lexlapax/go-llms/pkg/agent/builtins/tools/file |

#### Input Schema

- **Type**: object
- **Required**: path, content
- **Properties**:
  - **mode**:
    - **Type**: number
    - **Description**: File permissions in octal (default: 0644)
  - **path**:
    - **Type**: string
    - **Description**: The path to the file to write
  - **append**:
    - **Type**: boolean
    - **Description**: Append to existing file instead of overwriting
  - **atomic**:
    - **Type**: boolean
    - **Description**: Use atomic write operation (safer for important files)
  - **backup**:
    - **Type**: boolean
    - **Description**: Create backup of existing file before writing
  - **content**:
    - **Type**: string
    - **Description**: The content to write to the file
  - **create_dirs**:
    - **Type**: boolean
    - **Description**: Create parent directories if they don't exist

#### Output Schema

- **Type**: object
- **Required**: success, bytes_written, absolute_path, file_existed, mod_time
- **Properties**:
  - **absolute_path**:
    - **Type**: string
    - **Description**: Absolute path to the written file
  - **backup_path**:
    - **Type**: string
    - **Description**: Path to backup file if created
  - **bytes_written**:
    - **Type**: number
    - **Description**: Number of bytes written
  - **file_existed**:
    - **Type**: boolean
    - **Description**: Whether the file existed before writing
  - **mod_time**:
    - **Type**: string
    - **Description**: Modification time after write
  - **success**:
    - **Type**: boolean
    - **Description**: Whether the write operation succeeded

#### Examples

##### Example 1: Simple file write

Write text to a file

**Input:**
```json
{
  "content": "Meeting notes:\n- Discuss project timeline\n- Review budget",
  "path": "/home/user/notes.txt"
}
```

**Output:**
```json
{
  "absolute_path": "/home/user/notes.txt",
  "bytes_written": 48,
  "file_existed": false,
  "mod_time": "2024-01-15T10:30:00Z",
  "success": true
}
```

##### Example 2: Append to log file

Add entry to existing log

**Input:**
```json
{
  "append": true,
  "content": "[2024-01-15 10:30:00] User login successful\n",
  "path": "app.log"
}
```

**Output:**
```json
{
  "absolute_path": "/current/dir/app.log",
  "bytes_written": 46,
  "file_existed": true,
  "mod_time": "2024-01-15T10:30:00Z",
  "success": true
}
```

##### Example 3: Atomic write with backup

Safely update configuration file

**Input:**
```json
{
  "atomic": true,
  "backup": true,
  "content": "{\"version\": \"2.0\", \"port\": 8080, \"debug\": false}",
  "path": "config.json"
}
```

**Output:**
```json
{
  "absolute_path": "/app/config.json",
  "backup_path": "/app/config.backup-20240115-103000.json",
  "bytes_written": 48,
  "file_existed": true,
  "mod_time": "2024-01-15T10:30:00Z",
  "success": true
}
```

##### Example 4: Create file with directories

Write file in non-existent directory

**Input:**
```json
{
  "content": "Date,Product,Amount\n2024-01-15,Widget,100.00\n",
  "create_dirs": true,
  "path": "output/reports/2024/january/sales.csv"
}
```

**Output:**
```json
{
  "absolute_path": "/home/user/output/reports/2024/january/sales.csv",
  "bytes_written": 48,
  "file_existed": false,
  "mod_time": "2024-01-15T10:30:00Z",
  "success": true
}
```

##### Example 5: Write with custom permissions

Create executable script

**Input:**
```json
{
  "content": "#!/bin/bash\necho 'Deploying application...'\n",
  "mode": 755,
  "path": "deploy.sh"
}
```

**Output:**
```json
{
  "absolute_path": "/home/user/deploy.sh",
  "bytes_written": 44,
  "file_existed": false,
  "mod_time": "2024-01-15T10:30:00Z",
  "success": true
}
```

##### Example 6: Handle write errors

Attempt to write to read-only location

**Input:**
```json
{
  "content": "system config",
  "path": "/etc/system.conf"
}
```

**Output:**
```json
{
  "error": "error writing file: open /etc/system.conf: permission denied"
}
```

##### Example 7: Path restriction

Blocked by security policy

**Input:**
```json
{
  "content": "malicious content",
  "path": "/etc/passwd"
}
```

**Output:**
```json
{
  "error": "access denied: path /etc/passwd is restricted"
}
```

---

## math

### calculator

Performs mathematical calculations including arithmetic, trigonometry, and logarithms

Use this tool to perform mathematical calculations. It supports:

Basic Arithmetic:
- add (+): Addition of two numbers
- subtract (-): Subtraction (operand1 - operand2)
- multiply (*): Multiplication
- divide (/): Division (checks for division by zero)
- mod (%): Modulo operation
- power (^, **): Exponentiation
- abs: Absolute value

Roots and Logarithms:
- sqrt: Square root (requires non-negative operand)
- cbrt: Cube root
- log: Natural logarithm (base e) or logarithm with custom base
- log10: Base-10 logarithm
- log2: Base-2 logarithm
- exp: e raised to the power of operand1

Trigonometry (angles in radians):
- sin, cos, tan: Standard trigonometric functions
- asin, acos, atan: Inverse trigonometric functions
- sinh, cosh, tanh: Hyperbolic functions

Rounding:
- floor: Round down to nearest integer
- ceil: Round up to nearest integer
- round: Round to nearest integer

Advanced:
- factorial: Calculate n! (requires non-negative integer ≤ 170)
- gcd: Greatest common divisor (requires positive integers)
- lcm: Least common multiple (requires positive integers)

Mathematical Constants (no operands needed):
- pi (π): 3.14159...
- e: Euler's number (2.71828...)
- phi (φ): Golden ratio
- tau (τ): 2π
- sqrt2, sqrte, sqrtpi, sqrtphi: Square roots of constants
- ln2, ln10, log2e, log10e: Logarithmic constants

Special operand values:
- You can use constant names as operands, e.g., operand1: "pi"
- Numbers can be provided as strings and will be parsed

| Property | Value |
|----------|-------|
| **Category** | math |
| **Tags** | math, calculation, arithmetic, trigonometry |
| **Version** | 1.0.0 |
| **package** | github.com/lexlapax/go-llms/pkg/agent/builtins/tools/math |

#### Input Schema

- **Type**: object
- **Required**: operation
- **Properties**:
  - **operand1**:
    - **Type**: number
    - **Description**: First operand (or single operand for unary operations)
  - **operand2**:
    - **Type**: number
    - **Description**: Second operand (optional for unary operations)
  - **operation**:
    - **Type**: string
    - **Description**: The mathematical operation to perform

#### Output Schema

- **Type**: object

#### Examples

##### Example 1: Basic addition

Add two decimal numbers

**Input:**
```json
{
  "operand1": 10.5,
  "operand2": 5.2,
  "operation": "add"
}
```

**Output:**
```json
{
  "operand1": 10.5,
  "operand2": 5.2,
  "operation": "add",
  "result": 15.7,
  "success": true
}
```

##### Example 2: Square root

Calculate the square root of a number

**Input:**
```json
{
  "operand1": 16,
  "operation": "sqrt"
}
```

**Output:**
```json
{
  "operand1": 16,
  "operation": "sqrt",
  "result": 4,
  "success": true
}
```

##### Example 3: Trigonometry with constants

Calculate sine of π/2

**Input:**
```json
{
  "operand1": "pi",
  "operand2": 2,
  "operation": "sin"
}
```

**Output:**
```json
{
  "operand1": 1.5707963267948966,
  "operation": "sin",
  "result": 1,
  "success": true
}
```

##### Example 4: Get mathematical constant

Retrieve the value of π

**Input:**
```json
{
  "operation": "pi"
}
```

**Output:**
```json
{
  "operation": "pi",
  "result": 3.141592653589793,
  "success": true
}
```

##### Example 5: Division by zero error

Handle division by zero gracefully

**Input:**
```json
{
  "operand1": 10,
  "operand2": 0,
  "operation": "divide"
}
```

**Output:**
```json
{
  "error": "division by zero",
  "operand1": 10,
  "operand2": 0,
  "operation": "divide",
  "success": false
}
```

##### Example 6: Factorial calculation

Calculate 5!

**Input:**
```json
{
  "operand1": 5,
  "operation": "factorial"
}
```

**Output:**
```json
{
  "operand1": 5,
  "operation": "factorial",
  "result": 120,
  "success": true
}
```

##### Example 7: Logarithm with custom base

Calculate log base 2 of 8

**Input:**
```json
{
  "operand1": 8,
  "operand2": 2,
  "operation": "log"
}
```

**Output:**
```json
{
  "operand1": 8,
  "operand2": 2,
  "operation": "log",
  "result": 3,
  "success": true
}
```

---

## system

### execute_command

Executes system commands with enhanced control and security

Use this tool to execute system commands with enhanced control and security features.

Security Features:
- Safe mode (enabled by default) restricts dangerous commands
- Allowlisted commands in safe mode include common utilities
- Custom commands can be allowed via state configuration
- Command validation prevents injection attacks

Parameters:
- command: The command to execute (required)
- working_dir: Directory to execute in (optional, defaults to current)
- environment: Key-value pairs to add to environment (optional)
- timeout: Maximum execution time in seconds (optional, default 30, max 300)
- shell: Shell to use - sh, bash, zsh, or none for direct execution (optional, default sh)
- safe_mode: Enable/disable command safety checks (optional, default true)
- input: Data to provide via stdin (optional)

Output includes:
- stdout: Standard output from the command
- stderr: Standard error output
- exit_code: Command exit code (0 = success)
- success: Boolean indicating successful execution
- timed_out: Whether the command exceeded timeout
- duration_ms: Execution time in milliseconds

Safe Mode:
When safe_mode is true (default), the tool:
1. Blocks dangerous commands (rm, shutdown, format, etc.)
2. Blocks dangerous patterns (sudo, redirects, pipes, etc.)
3. Only allows commands from a safe allowlist
4. Permits full paths to system directories (/usr/bin/, /bin/, etc.)

To allow additional commands in safe mode, set them in state:
state.Set("allowed_commands", []string{"custom-tool", "my-script"})

| Property | Value |
|----------|-------|
| **Category** | system |
| **Tags** | command, shell, execution, system, process |
| **Version** | 1.0.0 |
| **package** | github.com/lexlapax/go-llms/pkg/agent/builtins/tools/system |

#### Input Schema

- **Type**: object
- **Required**: command
- **Properties**:
  - **safe_mode**:
    - **Type**: boolean
    - **Description**: Enable safe mode to restrict dangerous commands (default: true)
  - **shell**:
    - **Type**: string
    - **Description**: Shell to use (sh, bash, zsh, or none for direct execution)
  - **timeout**:
    - **Type**: number
    - **Description**: Timeout in seconds (default: 30, max: 300)
  - **working_dir**:
    - **Type**: string
    - **Description**: Working directory for command execution
  - **command**:
    - **Type**: string
    - **Description**: The command to execute
  - **environment**:
    - **Type**: object
    - **Description**: Environment variables to set (merged with current environment)
  - **input**:
    - **Type**: string
    - **Description**: Input to provide to the command via stdin

#### Output Schema

- **Type**: object
- **Required**: stdout, stderr, exit_code, success, command, working_dir, duration_ms
- **Properties**:
  - **stderr**:
    - **Type**: string
    - **Description**: Standard error output from the command
  - **success**:
    - **Type**: boolean
    - **Description**: Whether the command executed successfully
  - **timed_out**:
    - **Type**: boolean
    - **Description**: Whether the command timed out
  - **stdout**:
    - **Type**: string
    - **Description**: Standard output from the command
  - **command**:
    - **Type**: string
    - **Description**: The command that was executed
  - **duration_ms**:
    - **Type**: number
    - **Description**: Duration of command execution in milliseconds
  - **exit_code**:
    - **Type**: number
    - **Description**: Exit code of the command (0 indicates success)
  - **working_dir**:
    - **Type**: string
    - **Description**: The working directory where the command was executed
  - **environment**:
    - **Type**: object
    - **Description**: Custom environment variables that were set

#### Examples

##### Example 1: Basic command execution

Execute a simple echo command

**Input:**
```json
{
  "command": "echo 'Hello, World!'"
}
```

**Output:**
```json
{
  "exit_code": 0,
  "stderr": "",
  "stdout": "Hello, World!\n",
  "success": true,
  "timed_out": false
}
```

##### Example 2: List directory contents

List files in a specific directory

**Input:**
```json
{
  "command": "ls -la",
  "working_dir": "/tmp"
}
```

**Output:**
```json
{
  "exit_code": 0,
  "stdout": "total 8\ndrwxrwxrwt  2 root root 4096 Jan  1 00:00 .\ndrwxr-xr-x 23 root root 4096 Jan  1 00:00 ..\n",
  "success": true
}
```

##### Example 3: Environment variable usage

Execute command with custom environment variables

**Input:**
```json
{
  "command": "echo $MY_VAR - $HOME",
  "environment": {
    "MY_VAR": "custom value"
  }
}
```

**Output:**
```json
{
  "stdout": "custom value - /home/user\n",
  "success": true
}
```

##### Example 4: Command with timeout

Execute a long-running command with timeout

**Input:**
```json
{
  "command": "sleep 10",
  "timeout": 2
}
```

**Output:**
```json
{
  "duration_ms": 2000,
  "exit_code": null,
  "stderr": "",
  "stdout": "",
  "success": false,
  "timed_out": true
}
```

##### Example 5: Direct command execution

Execute command without shell interpretation

**Input:**
```json
{
  "command": "/usr/bin/ls -la /tmp",
  "shell": "none"
}
```

**Output:**
```json
{
  "success": true
}
```

##### Example 6: Command with input

Provide input to a command via stdin

**Input:**
```json
{
  "command": "wc -l",
  "input": "line1\nline2\nline3\n"
}
```

**Output:**
```json
{
  "stdout": "3\n",
  "success": true
}
```

##### Example 7: Error handling example

Handle command that returns non-zero exit code

**Input:**
```json
{
  "command": "ls /nonexistent-directory"
}
```

**Output:**
```json
{
  "exit_code": 2,
  "stderr": "ls: cannot access '/nonexistent-directory': No such file or directory\n",
  "stdout": "",
  "success": false
}
```

---

### get_environment_variable

Retrieves environment variables safely

Use this tool to safely read environment variables from the system.

Security Features:
- Sensitive variables (containing KEY, SECRET, TOKEN, PASSWORD, etc.) are masked by default
- Use 'sensitive: true' to see unmasked values when necessary
- Add custom sensitive patterns via state: state.Set("sensitive_env_patterns", []string{"*PRIVATE*"})

Parameters:
- name: Retrieve a specific environment variable by exact name (optional)
- pattern: Search for variables matching a pattern (optional)
- no_values: Return only variable names without values (optional, default false)
- sensitive: Allow unmasked retrieval of sensitive variables (optional, default false)

Pattern Matching:
- Use * as wildcard: "GO*" matches all variables starting with GO
- "*_PATH" matches all variables ending with _PATH
- "*API*" matches all variables containing API
- "*" or empty pattern returns all variables

Output:
- variables: Array of found environment variables
- count: Number of variables found
- query: The search term used (name or pattern)

Security Masking:
Sensitive values show only first 3 and last 3 characters:
- Full value: "sk-abc123def456ghi789"
- Masked: "sk-...789"

| Property | Value |
|----------|-------|
| **Category** | system |
| **Tags** | environment, config, system, variables |
| **Version** | 1.0.0 |
| **package** | github.com/lexlapax/go-llms/pkg/agent/builtins/tools/system |

#### Input Schema

- **Type**: object
- **Properties**:
  - **sensitive**:
    - **Type**: boolean
    - **Description**: Allow retrieval of potentially sensitive variables
  - **name**:
    - **Type**: string
    - **Description**: Specific environment variable name to retrieve
  - **no_values**:
    - **Type**: boolean
    - **Description**: Exclude values from results (default: false, meaning values are included)
  - **pattern**:
    - **Type**: string
    - **Description**: Pattern to match variable names (e.g., 'GO_*', '*_PATH')

#### Output Schema

- **Type**: object
- **Required**: variables, count
- **Properties**:
  - **count**:
    - **Type**: number
    - **Description**: Number of variables found
  - **query**:
    - **Type**: string
    - **Description**: The name or pattern that was searched
  - **variables**:
    - **Type**: array
    - **Description**: List of environment variables matching the query

#### Examples

##### Example 1: Get specific variable

Retrieve the HOME environment variable

**Input:**
```json
{
  "name": "HOME"
}
```

**Output:**
```json
{
  "count": 1,
  "query": "HOME",
  "variables": [
    {
      "name": "HOME",
      "value": "/home/user"
    }
  ]
}
```

##### Example 2: Find Go-related variables

List all environment variables starting with GO

**Input:**
```json
{
  "pattern": "GO*"
}
```

**Output:**
```json
{
  "count": 3,
  "query": "GO*",
  "variables": [
    {
      "name": "GOPATH",
      "value": "/home/user/go"
    },
    {
      "name": "GOROOT",
      "value": "/usr/local/go"
    },
    {
      "name": "GO111MODULE",
      "value": "on"
    }
  ]
}
```

##### Example 3: List PATH variables

Find all variables ending with PATH

**Input:**
```json
{
  "pattern": "*PATH"
}
```

**Output:**
```json
{
  "count": 3,
  "query": "*PATH",
  "variables": [
    {
      "name": "PATH",
      "value": "/usr/bin:/bin:/usr/local/bin"
    },
    {
      "name": "GOPATH",
      "value": "/home/user/go"
    },
    {
      "name": "PYTHONPATH",
      "value": "/usr/lib/python3"
    }
  ]
}
```

##### Example 4: List variable names only

Get all variable names without values

**Input:**
```json
{
  "no_values": true,
  "pattern": "*"
}
```

**Output:**
```json
{
  "count": 3,
  "query": "*",
  "variables": [
    {
      "name": "HOME"
    },
    {
      "name": "PATH"
    },
    {
      "name": "USER"
    }
  ]
}
```

##### Example 5: Handle sensitive variables

Retrieve API key with masked value

**Input:**
```json
{
  "name": "OPENAI_API_KEY"
}
```

**Output:**
```json
{
  "count": 1,
  "query": "OPENAI_API_KEY",
  "variables": [
    {
      "masked": true,
      "name": "OPENAI_API_KEY",
      "value": "sk-...789"
    }
  ]
}
```

##### Example 6: Retrieve sensitive unmasked

Get API key value unmasked when needed

**Input:**
```json
{
  "name": "OPENAI_API_KEY",
  "sensitive": true
}
```

**Output:**
```json
{
  "count": 1,
  "query": "OPENAI_API_KEY",
  "variables": [
    {
      "name": "OPENAI_API_KEY",
      "value": "sk-abc123def456ghi789"
    }
  ]
}
```

##### Example 7: Non-existent variable

Request a variable that doesn't exist

**Input:**
```json
{
  "name": "NONEXISTENT_VAR"
}
```

**Output:**
```json
{
  "count": 0,
  "query": "NONEXISTENT_VAR",
  "variables": null
}
```

---

### get_system_info

Retrieves comprehensive system information

Use this tool to retrieve comprehensive information about the system.

By default, returns basic system information:
- Operating system (name, platform, version if available)
- Architecture (e.g., amd64, arm64)
- Number of CPUs
- Hostname
- Timestamp

Optional information can be included:
- include_memory: Memory statistics (allocated, system, GC stats)
- include_runtime: Go runtime information (version, goroutines, GOMAXPROCS)
- include_environment: Environment summary (user, paths, env var count)

Parameters:
- include_environment: Add environment summary (optional, default false)
- include_memory: Add memory statistics (optional, default false)
- include_runtime: Add Go runtime info (optional, default false)

Memory statistics include:
- alloc: Current memory allocated and in use
- total_alloc: Total memory allocated since program start
- sys: Memory obtained from the OS
- num_gc: Number of garbage collection cycles

Runtime information includes:
- Go version and compiler
- Number of logical CPUs
- Current goroutine count
- GOMAXPROCS setting

Environment summary includes:
- Current user and home directory
- Working directory and temp directory
- PATH directory count
- Total environment variables
- Go-specific paths (GOPATH, GOROOT)

Note: You can set defaults via state:
- state.Set("system_info_include_memory", true)
- state.Set("system_info_include_runtime", true)
- state.Set("system_info_include_environment", true)

| Property | Value |
|----------|-------|
| **Category** | system |
| **Tags** | system, info, os, architecture, resources |
| **Version** | 1.0.0 |
| **package** | github.com/lexlapax/go-llms/pkg/agent/builtins/tools/system |

#### Input Schema

- **Type**: object
- **Properties**:
  - **include_environment**:
    - **Type**: boolean
    - **Description**: Include environment summary information
  - **include_memory**:
    - **Type**: boolean
    - **Description**: Include memory statistics
  - **include_runtime**:
    - **Type**: boolean
    - **Description**: Include Go runtime information

#### Output Schema

- **Type**: object
- **Required**: os, architecture, cpus, timestamp
- **Properties**:
  - **memory**:
    - **Type**: object
    - **Description**: Memory statistics (if include_memory is true)
  - **os**:
    - **Type**: object
    - **Description**: Operating system details
  - **runtime**:
    - **Type**: object
    - **Description**: Go runtime information (if include_runtime is true)
  - **timestamp**:
    - **Type**: string
    - **Description**: Timestamp when information was collected (RFC3339)
  - **architecture**:
    - **Type**: string
    - **Description**: System architecture (e.g., amd64, arm64)
  - **cpus**:
    - **Type**: number
    - **Description**: Number of CPU cores
  - **environment**:
    - **Type**: object
    - **Description**: Environment summary (if include_environment is true)
  - **hostname**:
    - **Type**: string
    - **Description**: System hostname

#### Examples

##### Example 1: Basic system information

Get core system details

**Output:**
```json
{
  "architecture": "amd64",
  "cpus": 8,
  "hostname": "dev-machine",
  "os": {
    "name": "linux",
    "platform": "Linux"
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

##### Example 2: System with memory statistics

Include current memory usage

**Input:**
```json
{
  "include_memory": true
}
```

**Output:**
```json
{
  "architecture": "arm64",
  "cpus": 10,
  "memory": {
    "alloc": 52428800,
    "num_gc": 5,
    "sys": 75497472,
    "total_alloc": 104857600
  },
  "os": {
    "name": "darwin",
    "platform": "macOS"
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

##### Example 3: Full system information

Get all available system details

**Input:**
```json
{
  "include_environment": true,
  "include_memory": true,
  "include_runtime": true
}
```

**Output:**
```json
{
  "architecture": "amd64",
  "cpus": 16,
  "environment": {
    "gopath": "/home/appuser/go",
    "goroot": "/usr/local/go",
    "home": "/home/appuser",
    "path_dirs": 12,
    "temp_dir": "/tmp",
    "total_env_vars": 35,
    "user": "appuser",
    "working_dir": "/app"
  },
  "hostname": "prod-server",
  "memory": {
    "alloc": 104857600,
    "num_gc": 10,
    "sys": 150994944,
    "total_alloc": 209715200
  },
  "os": {
    "name": "linux",
    "platform": "Linux",
    "version": "Ubuntu 22.04.3 LTS"
  },
  "runtime": {
    "compiler": "gc",
    "gomaxprocs": 16,
    "num_cpu": 16,
    "num_goroutine": 42,
    "version": "go1.21.5"
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

##### Example 4: Runtime monitoring

Check Go runtime statistics

**Input:**
```json
{
  "include_runtime": true
}
```

**Output:**
```json
{
  "architecture": "amd64",
  "cpus": 12,
  "hostname": "WIN-DEV",
  "os": {
    "name": "windows",
    "platform": "Windows"
  },
  "runtime": {
    "compiler": "gc",
    "gomaxprocs": 12,
    "num_cpu": 12,
    "num_goroutine": 156,
    "version": "go1.21.5"
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

##### Example 5: Environment check

Verify environment configuration

**Input:**
```json
{
  "include_environment": true
}
```

**Output:**
```json
{
  "architecture": "arm64",
  "cpus": 8,
  "environment": {
    "gopath": "/Users/developer/go",
    "goroot": "/opt/homebrew/opt/go/libexec",
    "home": "/Users/developer",
    "path_dirs": 15,
    "temp_dir": "/var/folders/xx/yyyyyy/T",
    "total_env_vars": 52,
    "user": "developer",
    "working_dir": "/Users/developer/projects"
  },
  "os": {
    "name": "darwin",
    "platform": "macOS"
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

---

### process_list

Lists running processes with filtering and sorting

Use this tool to list and analyze running processes on the system.

Cross-Platform Support:
- Unix/Linux/macOS: Uses 'ps aux' command for detailed process information
- Windows: Uses 'tasklist' command (limited CPU info)
- Other platforms: Returns minimal process information

Parameters:
- filter: Search for processes by name (case-insensitive, partial match)
- include_self: Include the current process (default: false)
- sort_by: Order results by pid, name, cpu, or memory
- limit: Maximum processes to return (1-1000)

Output Information:
- pid: Process identifier
- name: Process executable name
- command: Full command line (Unix only)
- cpu_percent: CPU usage percentage (Unix only)
- memory_kb: Memory usage in kilobytes
- user: Process owner
- start_time: When process started

Filtering:
- Searches both process name and command line
- Case-insensitive partial matching
- Example: filter "chrome" matches "Google Chrome Helper"

Sorting:
- pid: Ascending by process ID
- name: Alphabetical by process name
- cpu: Descending by CPU usage (highest first)
- memory: Descending by memory usage (highest first)

State Configuration:
Set default limit via state:
state.Set("process_list_default_limit", 50)

Platform Notes:
- CPU percentage may be 0 on Windows
- Command field may be empty on some platforms
- Memory values are estimates on some systems

| Property | Value |
|----------|-------|
| **Category** | system |
| **Tags** | process, system, monitoring, ps |
| **Version** | 1.0.0 |
| **package** | github.com/lexlapax/go-llms/pkg/agent/builtins/tools/system |

#### Output Schema

- **Type**: object
- **Required**: processes, count, platform, timestamp
- **Properties**:
  - **timestamp**:
    - **Type**: string
    - **Description**: Timestamp when the list was generated (RFC3339)
  - **count**:
    - **Type**: number
    - **Description**: Number of processes returned
  - **platform**:
    - **Type**: string
    - **Description**: Operating system platform
  - **processes**:
    - **Type**: array
    - **Description**: List of running processes

#### Input Schema

- **Type**: object
- **Properties**:
  - **filter**:
    - **Type**: string
    - **Description**: Filter processes by name (case-insensitive contains)
  - **include_self**:
    - **Type**: boolean
    - **Description**: Include the current process in results
  - **limit**:
    - **Type**: number
    - **Description**: Maximum number of processes to return
  - **sort_by**:
    - **Type**: string
    - **Description**: Sort results by: pid, name, cpu, memory

#### Examples

##### Example 1: List all processes

Get a complete process list

**Output:**
```json
{
  "count": 2,
  "platform": "darwin",
  "processes": [
    {
      "command": "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
      "cpu_percent": 5.2,
      "memory_kb": 524288,
      "name": "chrome",
      "pid": 1234,
      "start_time": "10:30AM",
      "user": "john"
    },
    {
      "command": "/usr/local/bin/code",
      "cpu_percent": 2.1,
      "memory_kb": 262144,
      "name": "code",
      "pid": 5678,
      "start_time": "09:15AM",
      "user": "john"
    }
  ],
  "timestamp": "2024-01-15T10:30:00Z"
}
```

##### Example 2: Find specific processes

Search for Chrome processes

**Input:**
```json
{
  "filter": "chrome"
}
```

**Output:**
```json
{
  "count": 2,
  "platform": "linux",
  "processes": [
    {
      "memory_kb": 524288,
      "name": "chrome",
      "pid": 1234
    },
    {
      "memory_kb": 8192,
      "name": "chrome_crashpad",
      "pid": 1235
    }
  ],
  "timestamp": "2024-01-15T10:30:00Z"
}
```

##### Example 3: Top CPU consumers

Find processes using most CPU

**Input:**
```json
{
  "limit": 5,
  "sort_by": "cpu"
}
```

**Output:**
```json
{
  "count": 5,
  "platform": "darwin",
  "processes": [
    {
      "cpu_percent": 95.5,
      "name": "video_encoder",
      "pid": 9999
    },
    {
      "cpu_percent": 45.2,
      "name": "chrome",
      "pid": 8888
    },
    {
      "cpu_percent": 25,
      "name": "spotlight",
      "pid": 7777
    },
    {
      "cpu_percent": 15.3,
      "name": "docker",
      "pid": 6666
    },
    {
      "cpu_percent": 10.1,
      "name": "vscode",
      "pid": 5555
    }
  ],
  "timestamp": "2024-01-15T10:30:00Z"
}
```

##### Example 4: Memory usage analysis

Find memory-hungry processes

**Input:**
```json
{
  "limit": 10,
  "sort_by": "memory"
}
```

**Output:**
```json
{
  "count": 3,
  "platform": "linux",
  "processes": [
    {
      "memory_kb": 2097152,
      "name": "docker",
      "pid": 1111
    },
    {
      "memory_kb": 1048576,
      "name": "chrome",
      "pid": 2222
    },
    {
      "memory_kb": 524288,
      "name": "slack",
      "pid": 3333
    }
  ],
  "timestamp": "2024-01-15T10:30:00Z"
}
```

##### Example 5: Include current process

List processes including self

**Input:**
```json
{
  "filter": "go",
  "include_self": true
}
```

**Output:**
```json
{
  "count": 3,
  "platform": "linux",
  "processes": [
    {
      "command": "go run main.go",
      "name": "go",
      "pid": 12345
    },
    {
      "command": "gopls serve",
      "name": "gopls",
      "pid": 12346
    },
    {
      "command": "./go-llms",
      "name": "go-llms",
      "pid": null
    }
  ],
  "timestamp": "2024-01-15T10:30:00Z"
}
```

##### Example 6: Windows process list

List processes on Windows

**Input:**
```json
{
  "limit": 3
}
```

**Output:**
```json
{
  "count": 3,
  "platform": "windows",
  "processes": [
    {
      "memory_kb": 512000,
      "name": "chrome.exe",
      "pid": 1000,
      "user": "SYSTEM"
    },
    {
      "memory_kb": 64000,
      "name": "svchost.exe",
      "pid": 2000,
      "user": "SYSTEM"
    },
    {
      "memory_kb": 128000,
      "name": "explorer.exe",
      "pid": 3000,
      "user": "User"
    }
  ],
  "timestamp": "2024-01-15T10:30:00Z"
}
```

---

## web

### api_client

Make REST API and GraphQL calls with automatic error handling and authentication support

Use this tool to interact with REST APIs and GraphQL endpoints. It handles:
- Multiple authentication methods (API key, Bearer token, Basic auth, OAuth2, Custom headers)
- Automatic JSON encoding/decoding
- Path parameter substitution
- Helpful error messages and guidance
- Common HTTP methods (GET, POST, PUT, DELETE, etc.)
- OpenAPI/Swagger spec discovery and validation
- GraphQL queries, mutations, and introspection

The tool will automatically set appropriate headers and handle responses intelligently.

Authentication Options:
- API Key: Set key location (header/query/cookie) and key name
- Bearer: Standard Authorization: Bearer token
- Basic: Username/password authentication
- OAuth2: Use access token from OAuth2 flow
- Custom: Set any custom header with optional prefix

Session Management:
- Enable enable_session=true to maintain cookies across requests
- Sessions are preserved in agent state for reuse

REST API Modes:
- Regular Mode: Standard REST API calls with path/query parameters
- OpenAPI Discovery: Set discover_operations=true to explore available endpoints
- OpenAPI Validation: Provide openapi_spec URL to validate requests

GraphQL Modes:
- Query/Mutation Mode: Provide graphql_query to execute GraphQL operations
- Discovery Mode: Set discover_graphql=true to introspect schema
- Variables are passed via graphql_variables parameter
- Automatically handles GraphQL-specific error formatting

| Property | Value |
|----------|-------|
| **Category** | web |
| **Tags** | api, rest, http, graphql, integration, client |
| **Version** | 4.0.0 |
| **package** | github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web |

#### Input Schema

- **Type**: object

#### Output Schema

- **Type**: object

#### Examples

##### Example 1: Simple GET request

Fetch user data from an API

**Input:**
```json
{
  "base_url": "https://api.github.com",
  "endpoint": "/users/octocat",
  "method": "GET"
}
```

**Output:**
```json
{
  "data": {
    "login": "octocat",
    "name": "The Octocat"
  },
  "status_code": 200,
  "success": true
}
```

##### Example 2: POST with authentication

Create a resource with API key authentication

**Input:**
```json
{
  "auth": {
    "api_key": "your-api-key",
    "key_location": "header",
    "key_name": "X-API-Key",
    "type": "api_key"
  },
  "base_url": "https://api.example.com",
  "body": {
    "description": "Created via API",
    "name": "New Item"
  },
  "endpoint": "/items",
  "method": "POST"
}
```

**Output:**
```json
{
  "data": {
    "created": true,
    "id": "12345"
  },
  "status_code": 201,
  "success": true
}
```

##### Example 3: Path parameters

Use path parameters in the endpoint

**Input:**
```json
{
  "base_url": "https://api.example.com",
  "endpoint": "/users/{user_id}/posts/{post_id}",
  "method": "GET",
  "path_params": {
    "post_id": "42",
    "user_id": "alice"
  }
}
```

**Output:**
```json
{
  "data": {
    "content": "Post content here",
    "title": "My Post"
  },
  "status_code": 200,
  "success": true
}
```

##### Example 4: OpenAPI discovery

Discover available operations from OpenAPI spec

**Input:**
```json
{
  "base_url": "https://api.example.com",
  "discover_operations": true,
  "endpoint": "/not-used-in-discovery",
  "openapi_spec": "https://api.example.com/openapi.json"
}
```

**Output:**
```json
{
  "operations": [
    {
      "method": "GET",
      "operationId": "listUsers",
      "path": "/users",
      "summary": "List users"
    },
    {
      "method": "GET",
      "operationId": "getUser",
      "path": "/users/{id}",
      "summary": "Get user by ID"
    }
  ],
  "spec_info": {
    "title": "Example API",
    "version": "1.0.0"
  },
  "success": true,
  "total_operations": 2
}
```

##### Example 5: GraphQL query

Execute a GraphQL query

**Input:**
```json
{
  "auth": {
    "token": "github_token_here",
    "type": "bearer"
  },
  "base_url": "https://api.github.com",
  "endpoint": "/graphql",
  "graphql_query": "query {\n  viewer {\n    login\n    name\n    email\n  }\n}"
}
```

**Output:**
```json
{
  "data": {
    "viewer": {
      "email": "octocat@github.com",
      "login": "octocat",
      "name": "The Octocat"
    }
  },
  "status_code": 200,
  "success": true
}
```

##### Example 6: GraphQL with variables

Execute a GraphQL query with variables

**Input:**
```json
{
  "auth": {
    "token": "github_token_here",
    "type": "bearer"
  },
  "base_url": "https://api.github.com",
  "endpoint": "/graphql",
  "graphql_query": "query GetRepo($owner: String!, $name: String!) {\n  repository(owner: $owner, name: $name) {\n    name\n    description\n    stargazerCount\n  }\n}",
  "graphql_variables": {
    "name": "go",
    "owner": "golang"
  }
}
```

**Output:**
```json
{
  "data": {
    "repository": {
      "description": "The Go programming language",
      "name": "go",
      "stargazerCount": 120000
    }
  },
  "status_code": 200,
  "success": true
}
```

##### Example 7: GraphQL discovery

Discover GraphQL schema

**Input:**
```json
{
  "auth": {
    "token": "github_token_here",
    "type": "bearer"
  },
  "base_url": "https://api.github.com",
  "discover_graphql": true,
  "endpoint": "/graphql"
}
```

**Output:**
```json
{
  "graphql_schema": {
    "endpoint": "https://api.github.com/graphql",
    "operations": {
      "queries": [
        {
          "description": "The currently authenticated user",
          "example": "query { viewer { login name email } }",
          "name": "viewer",
          "returns": "User"
        },
        {
          "arguments": [
            {
              "name": "owner",
              "required": true,
              "type": "String!"
            },
            {
              "name": "name",
              "required": true,
              "type": "String!"
            }
          ],
          "description": "Lookup a repository",
          "example": "query { repository(owner: \"owner\", name: \"repo\") { name } }",
          "name": "repository",
          "returns": "Repository"
        }
      ]
    }
  },
  "status_code": 200,
  "success": true
}
```

---

### http_request

Makes HTTP requests with full method and authentication support

Use this tool to make HTTP requests with full control over method, headers, body, and authentication.

Supported methods:
- GET: Retrieve data
- POST: Create new resources
- PUT: Update existing resources
- DELETE: Remove resources
- PATCH: Partial updates
- HEAD: Get headers only
- OPTIONS: Get allowed methods

Authentication methods:
- Basic: Username/password authentication
- Bearer: Token-based authentication (JWT, OAuth)
- API Key: Key in header or query parameter

Body types:
- json: application/json
- form: application/x-www-form-urlencoded
- xml: application/xml
- text: text/plain
- (default): application/octet-stream

State configuration:
- default_auth_type: Default authentication method
- api_key: Default API key for api_key auth
- bearer_token: Default token for bearer auth
- user_agent: Custom User-Agent header
- http_headers: Default headers as map[string]string

The tool will:
- Automatically handle redirects (unless disabled)
- Add query parameters to URL
- Set appropriate Content-Type for body
- Measure response time
- Return comprehensive response information

| Property | Value |
|----------|-------|
| **Category** | web |
| **Tags** | http, api, rest, post, put, delete, network |
| **Version** | 1.0.0 |
| **package** | github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web |

#### Input Schema

- **Type**: object
- **Required**: url
- **Properties**:
  - **body**:
    - **Type**: string
    - **Description**: Request body content
  - **body_type**:
    - **Type**: string
    - **Description**: Body content type (json, form, text, xml)
  - **auth_username**:
    - **Type**: string
    - **Description**: Username for basic auth
  - **follow_redirects**:
    - **Type**: boolean
    - **Description**: Whether to follow redirects (default: true)
  - **auth_type**:
    - **Type**: string
    - **Description**: Authentication type (basic, bearer, api_key)
  - **query_params**:
    - **Type**: object
    - **Description**: Query parameters to append to the URL
  - **headers**:
    - **Type**: object
    - **Description**: HTTP headers to include in the request
  - **url**:
    - **Type**: string
    - **Description**: The URL to send the request to
  - **auth_key_name**:
    - **Type**: string
    - **Description**: API key name
  - **auth_key_location**:
    - **Type**: string
    - **Description**: Where to place the API key (header or query)
  - **auth_token**:
    - **Type**: string
    - **Description**: Token for bearer auth
  - **method**:
    - **Type**: string
    - **Description**: HTTP method (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS)
  - **auth_key_value**:
    - **Type**: string
    - **Description**: API key value
  - **auth_password**:
    - **Type**: string
    - **Description**: Password for basic auth
  - **timeout**:
    - **Type**: number
    - **Description**: Request timeout in seconds (default: 30)

#### Output Schema

- **Type**: object
- **Required**: status_code, status, headers, body
- **Properties**:
  - **status_code**:
    - **Type**: number
    - **Description**: HTTP response status code
  - **body**:
    - **Type**: string
    - **Description**: Response body content
  - **content_length**:
    - **Type**: number
    - **Description**: Response Content-Length in bytes
  - **content_type**:
    - **Type**: string
    - **Description**: Response Content-Type header
  - **headers**:
    - **Type**: object
    - **Description**: Response headers
  - **redirect_url**:
    - **Type**: string
    - **Description**: Redirect Location header if present
  - **response_time_ms**:
    - **Type**: number
    - **Description**: Response time in milliseconds
  - **status**:
    - **Type**: string
    - **Description**: HTTP response status text

#### Examples

##### Example 1: Simple GET request

Basic data retrieval

**Input:**
```json
{
  "url": "https://api.example.com/users"
}
```

**Output:**
```json
{
  "body": "[{\"id\":1,\"name\":\"John\"},{\"id\":2,\"name\":\"Jane\"}]",
  "content_type": "application/json",
  "headers": {
    "Content-Type": "application/json"
  },
  "response_time_ms": 125,
  "status": "200 OK",
  "status_code": 200
}
```

##### Example 2: POST with JSON body

Create a new resource

**Input:**
```json
{
  "body": "{\"name\":\"Alice\",\"email\":\"alice@example.com\"}",
  "body_type": "json",
  "method": "POST",
  "url": "https://api.example.com/users"
}
```

**Output:**
```json
{
  "body": "{\"id\":3,\"name\":\"Alice\",\"email\":\"alice@example.com\"}",
  "headers": {
    "Location": "https://api.example.com/users/3"
  },
  "response_time_ms": 230,
  "status": "201 Created",
  "status_code": 201
}
```

##### Example 3: PUT with form data

Update resource with form encoding

**Input:**
```json
{
  "body": "name=Bob\u0026city=NYC\u0026age=30",
  "body_type": "form",
  "method": "PUT",
  "url": "https://api.example.com/profile"
}
```

**Output:**
```json
{
  "body": "{\"message\":\"Profile updated\"}",
  "response_time_ms": 150,
  "status": "200 OK",
  "status_code": 200
}
```

##### Example 4: DELETE request

Remove a resource

**Input:**
```json
{
  "method": "DELETE",
  "url": "https://api.example.com/users/123"
}
```

**Output:**
```json
{
  "body": "",
  "response_time_ms": 90,
  "status": "204 No Content",
  "status_code": 204
}
```

##### Example 5: Bearer token auth

Authenticated API request

**Input:**
```json
{
  "auth_token": "eyJhbGciOiJIUzI1NiIs...",
  "auth_type": "bearer",
  "url": "https://api.example.com/me"
}
```

**Output:**
```json
{
  "body": "{\"id\":42,\"username\":\"johndoe\"}",
  "response_time_ms": 100,
  "status_code": 200
}
```

##### Example 6: Basic authentication

Username/password auth

**Input:**
```json
{
  "auth_password": "secret123",
  "auth_type": "basic",
  "auth_username": "admin",
  "url": "https://api.example.com/admin"
}
```

**Output:**
```json
{
  "body": "{\"role\":\"admin\",\"permissions\":[\"read\",\"write\"]}",
  "status_code": 200
}
```

##### Example 7: API key in header

API key authentication

**Input:**
```json
{
  "auth_key_location": "header",
  "auth_key_name": "X-API-Key",
  "auth_key_value": "abc123xyz",
  "auth_type": "api_key",
  "url": "https://api.example.com/data"
}
```

**Output:**
```json
{
  "body": "{\"data\":[1,2,3,4,5]}",
  "status_code": 200
}
```

##### Example 8: Query parameters

Add URL parameters

**Input:**
```json
{
  "query_params": {
    "limit": "10",
    "page": "2",
    "q": "golang"
  },
  "url": "https://api.example.com/search"
}
```

**Output:**
```json
{
  "body": "{\"results\":[...],\"page\":2,\"total\":150}",
  "status_code": 200
}
```

##### Example 9: Custom headers

Add custom HTTP headers

**Input:**
```json
{
  "body": "{\"type\":\"query\"}",
  "headers": {
    "Accept": "application/vnd.api+json",
    "Cache-Control": "no-cache",
    "X-Request-ID": "uuid-123"
  },
  "method": "POST",
  "url": "https://api.example.com/v2/data"
}
```

**Output:**
```json
{
  "headers": {
    "X-Request-ID": "uuid-123"
  },
  "status_code": 200
}
```

##### Example 10: Handle redirects

Control redirect behavior

**Input:**
```json
{
  "follow_redirects": false,
  "url": "http://example.com/old-path"
}
```

**Output:**
```json
{
  "headers": {
    "Location": "https://example.com/new-path"
  },
  "redirect_url": "https://example.com/new-path",
  "status": "301 Moved Permanently",
  "status_code": 301
}
```

---

### web_fetch

Fetches content from a URL with customizable timeout

Use this tool to fetch content from a URL with optional authentication. The tool handles:
- HTTP/HTTPS URLs
- Customizable timeout (default 30 seconds)
- Multiple authentication methods (bearer, basic, API key, OAuth2, custom)
- Automatic content decoding
- Header extraction
- Proper error handling and status codes

Authentication methods:
- bearer: Sends "Authorization: Bearer <token>" header
- basic: Sends HTTP Basic Authentication with username/password
- api_key: Sends API key in header, query, or cookie
- oauth2: Sends OAuth2 access token as bearer token
- custom: Sends custom header with optional prefix

The tool will follow redirects automatically and handle common web server responses.
User agent can be customized via state (user_agent key).
Authentication can be auto-detected from state or provided via parameters.

| Property | Value |
|----------|-------|
| **Category** | web |
| **Tags** | http, fetch, download, web, network |
| **Version** | 1.0.0 |
| **package** | github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web |

#### Input Schema

- **Type**: object
- **Required**: url
- **Properties**:
  - **auth_key_name**:
    - **Type**: string
    - **Description**: API key parameter name (default: X-API-Key)
  - **auth_password**:
    - **Type**: string
    - **Description**: Password for basic authentication
  - **auth_prefix**:
    - **Type**: string
    - **Description**: Optional prefix for custom auth header value (e.g., 'Token')
  - **auth_token**:
    - **Type**: string
    - **Description**: Bearer token or general authentication token
  - **auth_header_name**:
    - **Type**: string
    - **Description**: Custom header name for custom authentication
  - **auth_username**:
    - **Type**: string
    - **Description**: Username for basic authentication
  - **auth_api_key**:
    - **Type**: string
    - **Description**: API key value
  - **auth_type**:
    - **Type**: string
    - **Description**: Authentication type: 'bearer', 'basic', 'api_key', 'oauth2', 'custom'
  - **timeout**:
    - **Type**: number
    - **Description**: Request timeout in seconds (default: 30)
  - **url**:
    - **Type**: string
    - **Description**: The URL to fetch content from
  - **auth_header_value**:
    - **Type**: string
    - **Description**: Custom header value for custom authentication
  - **auth_key_location**:
    - **Type**: string
    - **Description**: Where to place API key: 'header', 'query', or 'cookie' (default: header)

#### Output Schema

- **Type**: object
- **Required**: content, status_code, status_text
- **Properties**:
  - **content**:
    - **Type**: string
    - **Description**: The fetched content from the URL
  - **headers**:
    - **Type**: object
    - **Description**: Response headers
  - **status_code**:
    - **Type**: number
    - **Description**: HTTP status code of the response
  - **status_text**:
    - **Type**: string
    - **Description**: HTTP status text (e.g., '200 OK')

#### Examples

##### Example 1: Fetch a web page

Basic web page retrieval

**Input:**
```json
{
  "url": "https://example.com"
}
```

**Output:**
```json
{
  "content": "\u003c!DOCTYPE html\u003e\n\u003chtml\u003e\n\u003chead\u003e\u003ctitle\u003eExample Domain\u003c/title\u003e...\u003c/html\u003e",
  "headers": {
    "Content-Type": "text/html; charset=UTF-8"
  },
  "status_code": 200,
  "status_text": "200 OK"
}
```

##### Example 2: Fetch API endpoint

Retrieve JSON from an API

**Input:**
```json
{
  "url": "https://api.github.com/users/octocat"
}
```

**Output:**
```json
{
  "content": "{\"login\":\"octocat\",\"id\":1,\"node_id\":\"MDQ6VXNlcjE=\",\"avatar_url\":\"https://github.com/images/error/octocat_happy.gif\"...}",
  "headers": {
    "Content-Type": "application/json; charset=utf-8"
  },
  "status_code": 200,
  "status_text": "200 OK"
}
```

##### Example 3: With custom timeout

Fetch with extended timeout

**Input:**
```json
{
  "timeout": 120,
  "url": "https://slow-server.example.com/large-file"
}
```

**Output:**
```json
{
  "content": "[Large file content...]",
  "status_code": 200,
  "status_text": "200 OK"
}
```

##### Example 4: Handle 404 error

Non-existent page

**Input:**
```json
{
  "url": "https://example.com/does-not-exist"
}
```

**Output:**
```json
{
  "content": "404 page not found",
  "status_code": 404,
  "status_text": "404 Not Found"
}
```

##### Example 5: Handle redirect

Follow redirects automatically

**Input:**
```json
{
  "url": "http://github.com"
}
```

**Output:**
```json
{
  "content": "[GitHub homepage HTML...]",
  "headers": {
    "Content-Type": "text/html; charset=utf-8"
  },
  "status_code": 200,
  "status_text": "200 OK"
}
```

##### Example 6: Timeout error

Request times out

**Input:**
```json
{
  "timeout": 5,
  "url": "https://very-slow-server.example.com"
}
```

**Output:**
```json
{
  "error": "request timeout after 5s"
}
```

##### Example 7: Bearer token authentication

Fetch with bearer token

**Input:**
```json
{
  "auth_token": "ghp_xxxxxxxxxxxxxxxxxxxx",
  "auth_type": "bearer",
  "url": "https://api.github.com/user"
}
```

**Output:**
```json
{
  "content": "{\"login\":\"username\",\"id\":12345,...}",
  "status_code": 200,
  "status_text": "200 OK"
}
```

##### Example 8: API key authentication

Fetch with API key in header

**Input:**
```json
{
  "auth_api_key": "abc123xyz",
  "auth_key_name": "X-API-Key",
  "auth_type": "api_key",
  "url": "https://api.example.com/data"
}
```

**Output:**
```json
{
  "content": "{\"data\":[1,2,3]}",
  "status_code": 200,
  "status_text": "200 OK"
}
```

##### Example 9: Basic authentication

Fetch with username/password

**Input:**
```json
{
  "auth_password": "pass",
  "auth_type": "basic",
  "auth_username": "user",
  "url": "https://api.example.com/protected"
}
```

**Output:**
```json
{
  "content": "{\"message\":\"authenticated\"}",
  "status_code": 200,
  "status_text": "200 OK"
}
```

##### Example 10: Invalid URL

Malformed URL

**Input:**
```json
{
  "url": "not-a-valid-url"
}
```

**Output:**
```json
{
  "error": "invalid URL: not-a-valid-url"
}
```

---

### web_scrape

Extracts structured data from HTML pages

Use this tool to extract structured data from HTML pages with optional authentication. The tool handles:
- HTML parsing and content extraction
- CSS-like selector support (basic tag names)
- Link discovery and classification
- Metadata extraction (title, description, keywords)
- Text content cleaning
- Configurable timeout
- Multiple authentication methods (bearer, basic, API key, OAuth2, custom)

Features:
- extract_text: Get all text content with HTML tags removed
- extract_links: Find all links with type classification (internal/external/anchor)
- extract_meta: Extract metadata from meta tags
- selectors: Extract content matching specific CSS-like selectors (currently supports tag names)

Authentication methods:
- bearer: Sends "Authorization: Bearer <token>" header
- basic: Sends HTTP Basic Authentication with username/password
- api_key: Sends API key in header, query, or cookie
- oauth2: Sends OAuth2 access token as bearer token
- custom: Sends custom header with optional prefix

The tool will:
- Automatically detect content type
- Clean and format extracted text
- Resolve relative URLs to absolute
- Handle common HTML entities
- Filter out script and style content

State configuration:
- user_agent: Custom user agent string
- http_headers: Additional headers as map[string]string
- scrape_selectors: Additional selectors to extract
- respect_robots_txt: Enable robots.txt compliance (future feature)
- Authentication can be auto-detected from state or provided via parameters

| Property | Value |
|----------|-------|
| **Category** | web |
| **Tags** | scrape, html, extract, parse, web, network |
| **Version** | 1.0.0 |
| **package** | github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web |

#### Input Schema

- **Type**: object
- **Required**: url
- **Properties**:
  - **auth_prefix**:
    - **Type**: string
    - **Description**: Optional prefix for custom auth header value (e.g., 'Token')
  - **auth_token**:
    - **Type**: string
    - **Description**: Bearer token or general authentication token
  - **extract_links**:
    - **Type**: boolean
    - **Description**: Extract all links from the page (default: true)
  - **auth_api_key**:
    - **Type**: string
    - **Description**: API key value
  - **auth_username**:
    - **Type**: string
    - **Description**: Username for basic authentication
  - **auth_header_value**:
    - **Type**: string
    - **Description**: Custom header value for custom authentication
  - **extract_text**:
    - **Type**: boolean
    - **Description**: Extract all text content from the page (default: true)
  - **url**:
    - **Type**: string
    - **Description**: The URL to scrape
  - **auth_type**:
    - **Type**: string
    - **Description**: Authentication type: 'bearer', 'basic', 'api_key', 'oauth2', 'custom'
  - **auth_header_name**:
    - **Type**: string
    - **Description**: Custom header name for custom authentication
  - **auth_key_name**:
    - **Type**: string
    - **Description**: API key parameter name (default: X-API-Key)
  - **selectors**:
    - **Type**: array
    - **Description**: CSS-like selectors to extract specific elements (simplified)
  - **auth_key_location**:
    - **Type**: string
    - **Description**: Where to place API key: 'header', 'query', or 'cookie' (default: header)
  - **timeout**:
    - **Type**: number
    - **Description**: Request timeout in seconds (default: 30)
  - **auth_password**:
    - **Type**: string
    - **Description**: Password for basic authentication
  - **max_depth**:
    - **Type**: number
    - **Description**: Maximum depth for following links (0 = current page only, default: 0)
  - **extract_meta**:
    - **Type**: boolean
    - **Description**: Extract metadata (title, description, keywords) (default: true)

#### Output Schema

- **Type**: object
- **Required**: url, status_code, content_type, timestamp
- **Properties**:
  - **url**:
    - **Type**: string
    - **Description**: The URL that was scraped
  - **status_code**:
    - **Type**: number
    - **Description**: HTTP status code
  - **links**:
    - **Type**: array
    - **Description**: Links found on the page
  - **metadata**:
    - **Type**: object
    - **Description**: Page metadata (description, keywords, etc.)
  - **content_type**:
    - **Type**: string
    - **Description**: Content-Type header value
  - **text**:
    - **Type**: string
    - **Description**: Text content extracted from the page
  - **title**:
    - **Type**: string
    - **Description**: Page title extracted from HTML
  - **selectors**:
    - **Type**: object
    - **Description**: Content extracted by CSS selectors
  - **timestamp**:
    - **Type**: string
    - **Description**: ISO 8601 timestamp of when the page was scraped

#### Examples

##### Example 1: Basic web scraping

Extract all content from a webpage

**Input:**
```json
{
  "url": "https://example.com"
}
```

**Output:**
```json
{
  "content_type": "text/html; charset=UTF-8",
  "links": [
    {
      "text": "More information...",
      "type": "external",
      "url": "https://www.iana.org/domains/example"
    }
  ],
  "metadata": {
    "description": "Example domain for documentation"
  },
  "status_code": 200,
  "text": "Example Domain This domain is for use in illustrative examples...",
  "timestamp": "2024-01-15T10:00:00Z",
  "title": "Example Domain",
  "url": "https://example.com"
}
```

##### Example 2: Extract specific elements

Use selectors to extract specific content

**Input:**
```json
{
  "selectors": [
    "h1",
    "h2",
    "p"
  ],
  "url": "https://news.example.com"
}
```

**Output:**
```json
{
  "selectors": {
    "h1": [
      "Breaking News",
      "Top Stories"
    ],
    "h2": [
      "Technology",
      "Business",
      "Sports"
    ],
    "p": [
      "First paragraph...",
      "Second paragraph..."
    ]
  },
  "status_code": 200,
  "timestamp": "2024-01-15T10:00:00Z",
  "url": "https://news.example.com"
}
```

##### Example 3: Extract only links

Get all links from a page

**Input:**
```json
{
  "extract_links": true,
  "extract_meta": false,
  "extract_text": false,
  "url": "https://blog.example.com"
}
```

**Output:**
```json
{
  "links": [
    {
      "text": "First Post",
      "type": "internal",
      "url": "https://blog.example.com/post1"
    },
    {
      "text": "Second Post",
      "type": "internal",
      "url": "https://blog.example.com/post2"
    },
    {
      "text": "Follow us",
      "type": "external",
      "url": "https://twitter.com/blog"
    },
    {
      "text": "Comments",
      "type": "anchor",
      "url": "#comments"
    }
  ],
  "status_code": 200,
  "timestamp": "2024-01-15T10:00:00Z",
  "url": "https://blog.example.com"
}
```

##### Example 4: Extract metadata only

Get page metadata without content

**Input:**
```json
{
  "extract_links": false,
  "extract_meta": true,
  "extract_text": false,
  "url": "https://shop.example.com/product"
}
```

**Output:**
```json
{
  "metadata": {
    "description": "Buy the amazing product for only $99",
    "keywords": "product, amazing, shop",
    "og:description": "The best product you'll ever buy",
    "og:image": "https://shop.example.com/images/product.jpg",
    "og:title": "Amazing Product"
  },
  "status_code": 200,
  "timestamp": "2024-01-15T10:00:00Z",
  "title": "Amazing Product - Shop Example",
  "url": "https://shop.example.com/product"
}
```

##### Example 5: Scrape with timeout

Set custom timeout for slow sites

**Input:**
```json
{
  "timeout": 60,
  "url": "https://slow-site.example.com"
}
```

**Output:**
```json
{
  "status_code": 200,
  "text": "Content that took a while to load...",
  "timestamp": "2024-01-15T10:00:00Z",
  "url": "https://slow-site.example.com"
}
```

##### Example 6: Handle non-HTML content

Attempt to scrape non-HTML

**Input:**
```json
{
  "url": "https://api.example.com/data.json"
}
```

**Output:**
```json
{
  "error": "content type 'application/json' is not HTML/XML"
}
```

##### Example 7: Complex selector extraction

Extract multiple tag types

**Input:**
```json
{
  "extract_text": true,
  "max_depth": 0,
  "selectors": [
    "h1",
    "h2",
    "h3",
    "code",
    "pre"
  ],
  "url": "https://docs.example.com"
}
```

**Output:**
```json
{
  "selectors": {
    "code": [
      "npm install",
      "const api = new API()"
    ],
    "h1": [
      "API Documentation"
    ],
    "h2": [
      "Getting Started",
      "Authentication",
      "Endpoints"
    ],
    "h3": [
      "Installation",
      "Configuration",
      "Examples"
    ],
    "pre": [
      "{ \"status\": \"ok\" }"
    ]
  },
  "status_code": 200,
  "text": "Full text content of the documentation page...",
  "timestamp": "2024-01-15T10:00:00Z",
  "url": "https://docs.example.com"
}
```

##### Example 8: Bearer token authentication

Scrape protected page with bearer token

**Input:**
```json
{
  "auth_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "auth_type": "bearer",
  "url": "https://private.example.com/protected-page"
}
```

**Output:**
```json
{
  "status_code": 200,
  "text": "This is protected content only visible to authenticated users...",
  "timestamp": "2024-01-15T10:00:00Z",
  "title": "Protected Content",
  "url": "https://private.example.com/protected-page"
}
```

##### Example 9: API key authentication

Scrape with API key in header

**Input:**
```json
{
  "auth_api_key": "abc123xyz789",
  "auth_key_name": "X-API-Key",
  "auth_type": "api_key",
  "url": "https://api-docs.example.com/documentation"
}
```

**Output:**
```json
{
  "status_code": 200,
  "text": "Welcome to our API documentation...",
  "timestamp": "2024-01-15T10:00:00Z",
  "title": "API Documentation",
  "url": "https://api-docs.example.com/documentation"
}
```

##### Example 10: Basic authentication

Scrape with username/password

**Input:**
```json
{
  "auth_password": "secret123",
  "auth_type": "basic",
  "auth_username": "admin",
  "url": "https://secure.example.com/admin"
}
```

**Output:**
```json
{
  "status_code": 200,
  "text": "Admin control panel content...",
  "timestamp": "2024-01-15T10:00:00Z",
  "title": "Admin Dashboard",
  "url": "https://secure.example.com/admin"
}
```

---

### web_search

Performs web searches using various search engines (DuckDuckGo, Brave, Tavily, Serpapi, Serper.dev)

Use this tool to search the web using various search engines with optional authentication. The tool automatically selects the best available search engine based on API keys.

Available engines:
- duckduckgo: Free, no API key required, limited results
- brave: Comprehensive web search (requires BRAVE_API_KEY)
- tavily: AI-optimized search with summaries (requires TAVILY_API_KEY) - best for LLM applications
- serpapi: Google search results (requires SERPAPI_API_KEY)
- serperdev: Fast Google search results (requires SERPERDEV_API_KEY)
- searx: Privacy-focused metasearch (requires searx_url in state)

The tool will automatically:
- Select the best available engine based on API keys
- Handle rate limiting and retries
- Filter results based on safe search settings
- Limit results to the requested maximum (up to 50)

API Key Management:
- Set API keys via environment variables (BRAVE_API_KEY, TAVILY_API_KEY, etc.)
- Or provide engine_api_key parameter to override environment variables
- Keys in state (search_api_key) also work for backward compatibility

Authentication methods (for custom search endpoints):
- bearer: Sends "Authorization: Bearer <token>" header
- basic: Sends HTTP Basic Authentication with username/password
- api_key: Sends API key in header, query, or cookie
- oauth2: Sends OAuth2 access token as bearer token
- custom: Sends custom header with optional prefix

Authentication can be auto-detected from state or provided via parameters.

| Property | Value |
|----------|-------|
| **Category** | web |
| **Tags** | search, web, query, internet, network, brave, tavily, duckduckgo, serpapi, serperdev, google |
| **Version** | 2.0.0 |
| **package** | github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web |

#### Input Schema

- **Type**: object
- **Required**: query
- **Properties**:
  - **auth_header_name**:
    - **Type**: string
    - **Description**: Custom header name for custom authentication
  - **query**:
    - **Type**: string
    - **Description**: The search query
  - **safe_search**:
    - **Type**: boolean
    - **Description**: Enable safe search filtering (default: true)
  - **max_results**:
    - **Type**: number
    - **Description**: Maximum number of results to return (default: 10, max: 50)
  - **engine**:
    - **Type**: string
    - **Description**: Search engine to use (duckduckgo, brave, tavily, serpapi, serperdev, searx, or custom)
  - **engine_api_key**:
    - **Type**: string
    - **Description**: Optional API key for the search engine (overrides environment variables)
  - **auth_key_location**:
    - **Type**: string
    - **Description**: Where to place API key: 'header', 'query', or 'cookie' (default: header)
  - **auth_username**:
    - **Type**: string
    - **Description**: Username for basic authentication
  - **timeout**:
    - **Type**: number
    - **Description**: Request timeout in seconds (default: 30)
  - **auth_key_name**:
    - **Type**: string
    - **Description**: API key parameter name (default: X-API-Key)
  - **auth_header_value**:
    - **Type**: string
    - **Description**: Custom header value for custom authentication
  - **auth_password**:
    - **Type**: string
    - **Description**: Password for basic authentication
  - **auth_type**:
    - **Type**: string
    - **Description**: Authentication type: 'bearer', 'basic', 'api_key', 'oauth2', 'custom'
  - **auth_token**:
    - **Type**: string
    - **Description**: Bearer token or general authentication token
  - **auth_api_key**:
    - **Type**: string
    - **Description**: API key value
  - **auth_prefix**:
    - **Type**: string
    - **Description**: Optional prefix for custom auth header value (e.g., 'Token')

#### Output Schema

- **Type**: object
- **Required**: query, engine, results
- **Properties**:
  - **query**:
    - **Type**: string
    - **Description**: The search query that was executed
  - **results**:
    - **Type**: array
    - **Description**: Array of search results
  - **time_ms**:
    - **Type**: number
    - **Description**: Search execution time in milliseconds
  - **total_found**:
    - **Type**: number
    - **Description**: Total number of results found
  - **engine**:
    - **Type**: string
    - **Description**: The search engine that was used

#### Examples

##### Example 1: Basic web search

Search for information using default engine

**Input:**
```json
{
  "query": "latest AI developments 2024"
}
```

**Output:**
```json
{
  "engine": "tavily",
  "query": "latest AI developments 2024",
  "results": [
    {
      "description": "Overview of significant AI advancements...",
      "snippet": "In 2024, artificial intelligence saw unprecedented growth...",
      "title": "Major AI Breakthroughs in 2024",
      "url": "https://example.com/ai-2024"
    }
  ],
  "time_ms": 342,
  "total_found": 10
}
```

##### Example 2: Search with specific engine

Use a specific search engine

**Input:**
```json
{
  "engine": "brave",
  "max_results": 5,
  "query": "python programming tutorials"
}
```

**Output:**
```json
{
  "engine": "brave",
  "query": "python programming tutorials",
  "results": [
    {
      "description": "Well organized and easy to understand Web building tutorials",
      "title": "Python Tutorial - W3Schools",
      "url": "https://www.w3schools.com/python/"
    }
  ],
  "time_ms": 215,
  "total_found": 5
}
```

##### Example 3: Search with API key override

Provide API key directly

**Input:**
```json
{
  "engine": "serpapi",
  "engine_api_key": "your-serpapi-key-here",
  "max_results": 20,
  "query": "climate change research papers"
}
```

**Output:**
```json
{
  "engine": "serpapi",
  "query": "climate change research papers",
  "time_ms": 523,
  "total_found": 20
}
```

##### Example 4: Search with safe search disabled

Search without content filtering

**Input:**
```json
{
  "query": "medical procedures",
  "safe_search": false
}
```

**Output:**
```json
{
  "engine": "duckduckgo",
  "query": "medical procedures",
  "time_ms": 189,
  "total_found": 10
}
```

##### Example 5: Handle missing API keys

Fallback to free engine

**Input:**
```json
{
  "query": "open source projects"
}
```

**Output:**
```json
{
  "engine": "duckduckgo",
  "query": "open source projects",
  "time_ms": 412,
  "total_found": 8
}
```

##### Example 6: Search with custom timeout

Set longer timeout for slow connections

**Input:**
```json
{
  "query": "comprehensive market analysis reports 2024",
  "timeout": 60
}
```

**Output:**
```json
{
  "engine": "tavily",
  "query": "comprehensive market analysis reports 2024",
  "time_ms": 2341,
  "total_found": 15
}
```

##### Example 7: Error: Invalid engine

Handle unsupported engine

**Input:**
```json
{
  "engine": "invalid_engine",
  "query": "test query"
}
```

**Output:**
```json
{
  "error": "unsupported search engine: invalid_engine"
}
```

##### Example 8: Search with bearer token authentication

Search protected custom search endpoint

**Input:**
```json
{
  "auth_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "auth_type": "bearer",
  "engine": "custom",
  "query": "internal documents"
}
```

**Output:**
```json
{
  "engine": "custom",
  "query": "internal documents",
  "time_ms": 452,
  "total_found": 5
}
```

##### Example 9: Search with API key authentication

Search with API key in custom header

**Input:**
```json
{
  "auth_api_key": "abc123xyz789",
  "auth_key_name": "X-Custom-API-Key",
  "auth_type": "api_key",
  "engine": "custom",
  "query": "research papers"
}
```

**Output:**
```json
{
  "engine": "custom",
  "query": "research papers",
  "time_ms": 523,
  "total_found": 12
}
```

---

