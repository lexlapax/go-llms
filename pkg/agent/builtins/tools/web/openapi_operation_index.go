// ABOUTME: Operation index for fast lookup of OpenAPI operations by path and method
// ABOUTME: Improves performance by avoiding linear search through all operations

package web

import (
	"fmt"
	"strings"
)

// OperationIndex provides fast lookup of operations by path and method
type OperationIndex struct {
	// Key format: "METHOD /path" (e.g., "GET /users/{id}")
	operations map[string]*EnhancedOperationInfo

	// Operations grouped by tag for organized discovery
	byTag map[string][]*EnhancedOperationInfo

	// All operations in original order
	allOperations []EnhancedOperationInfo
}

// NewOperationIndex creates an index from enhanced operation info
func NewOperationIndex(operations []EnhancedOperationInfo) *OperationIndex {
	index := &OperationIndex{
		operations:    make(map[string]*EnhancedOperationInfo),
		byTag:         make(map[string][]*EnhancedOperationInfo),
		allOperations: operations,
	}

	// Build indices
	for i := range operations {
		op := &operations[i]

		// Create lookup key
		key := fmt.Sprintf("%s %s", strings.ToUpper(op.Method), op.Path)
		index.operations[key] = op

		// Group by tags
		if len(op.Tags) > 0 {
			for _, tag := range op.Tags {
				index.byTag[tag] = append(index.byTag[tag], op)
			}
		} else {
			// Add to "untagged" group
			index.byTag["untagged"] = append(index.byTag["untagged"], op)
		}
	}

	return index
}

// FindOperation looks up an operation by method and path
func (idx *OperationIndex) FindOperation(method, path string) (*EnhancedOperationInfo, bool) {
	key := fmt.Sprintf("%s %s", strings.ToUpper(method), path)
	op, found := idx.operations[key]
	return op, found
}

// GetOperationsByTag returns all operations with the specified tag
func (idx *OperationIndex) GetOperationsByTag(tag string) []*EnhancedOperationInfo {
	return idx.byTag[tag]
}

// GetAllTags returns all unique tags in the index
func (idx *OperationIndex) GetAllTags() []string {
	tags := make([]string, 0, len(idx.byTag))
	for tag := range idx.byTag {
		tags = append(tags, tag)
	}
	return tags
}

// GetAllOperations returns all operations in original order
func (idx *OperationIndex) GetAllOperations() []EnhancedOperationInfo {
	return idx.allOperations
}

// CountOperations returns the total number of operations
func (idx *OperationIndex) CountOperations() int {
	return len(idx.allOperations)
}

// HasOperation checks if an operation exists for the given method and path
func (idx *OperationIndex) HasOperation(method, path string) bool {
	_, found := idx.FindOperation(method, path)
	return found
}

// GetOperationsByPrefix finds operations with paths starting with the given prefix
func (idx *OperationIndex) GetOperationsByPrefix(pathPrefix string) []*EnhancedOperationInfo {
	var matches []*EnhancedOperationInfo

	for _, op := range idx.allOperations {
		if strings.HasPrefix(op.Path, pathPrefix) {
			matches = append(matches, &idx.allOperations[len(matches)])
		}
	}

	return matches
}

// GetDeprecatedOperations returns all deprecated operations
func (idx *OperationIndex) GetDeprecatedOperations() []*EnhancedOperationInfo {
	var deprecated []*EnhancedOperationInfo

	for i := range idx.allOperations {
		if idx.allOperations[i].Deprecated {
			deprecated = append(deprecated, &idx.allOperations[i])
		}
	}

	return deprecated
}
