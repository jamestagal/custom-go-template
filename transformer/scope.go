package transformer

import (
	"encoding/json"
	"log"
	"reflect"
	"strings"

	"github.com/jimafisk/custom_go_template/ast"
)

// InitDataScope initializes the data scope with provided props
func InitDataScope(props map[string]any) map[string]any {
	log.Printf("[DIAGNOSTIC] ========== InitDataScope START ==========")
	log.Printf("[DIAGNOSTIC] InitDataScope: received %d props", len(props))

	// Create a new map to avoid modifying the original props
	dataScope := make(map[string]any)

	// Copy all props to the data scope (REVERTED: removed __PROP__ prefix)
	// Props need to be actual values for build-time loop expansion to work
	for key, value := range props {
		dataScope[key] = value
		log.Printf("[DIAGNOSTIC] InitDataScope: copied prop '%s' (type=%T)", key, value)

		// DIAGNOSTIC: Log arrays in detail
		if key == "components" {
			if arr, ok := value.([]interface{}); ok {
				log.Printf("[DIAGNOSTIC] InitDataScope: ✓ 'components' is array with %d items", len(arr))
			}
		}
	}

	log.Printf("[DIAGNOSTIC] InitDataScope: returning dataScope with %d keys", len(dataScope))
	log.Printf("[DIAGNOSTIC] ========== InitDataScope END ==========")
	return dataScope
}

// FindFenceSection locates a fence section in the AST nodes
func FindFenceSection(nodes []ast.Node) *ast.FenceSection {
	for _, node := range nodes {
		if fence, ok := node.(*ast.FenceSection); ok {
			return fence
		}
	}
	return nil
}

// CollectFenceData extracts variables from fence section and adds them to data scope.
// CRITICAL FIX: Now uses parseValue() for consistent handling of JavaScript literals.
// This ensures quoted arrays/objects like let animals = "[...]" are unwrapped properly.
func CollectFenceData(fence *ast.FenceSection, dataScope map[string]any) {
	log.Printf("[DIAGNOSTIC] ========== CollectFenceData START ==========")
	log.Printf("[DIAGNOSTIC] CollectFenceData: dataScope has %d keys BEFORE processing fence", len(dataScope))

	// DIAGNOSTIC: Check if components exists before fence processing
	if componentsRaw, ok := dataScope["components"]; ok {
		if components, ok := componentsRaw.([]interface{}); ok {
			log.Printf("[DIAGNOSTIC] CollectFenceData: ✓ 'components' exists BEFORE fence processing (%d items)", len(components))
		}
	} else {
		log.Printf("[DIAGNOSTIC] CollectFenceData: ✗ 'components' does NOT exist before fence processing")
	}

	// Process variables directly from the FenceSection struct
	for _, variable := range fence.Variables {
		varName := variable.Name
		varValue := variable.Value

		log.Printf("[CollectFenceData] Processing variable: %s = %q", varName, varValue)

		// CRITICAL FIX: Use parseValue() for consistent JavaScript literal handling
		// This handles:
		// - Quoted arrays: let animals = "['dog','cat']" → unwrapped to ['dog','cat']
		// - Quoted objects: let user = "{name:'John'}" → unwrapped to {name:'John'}
		// - Regular values: let name = "John" → unwrapped to John
		parsedValue := parseValue(varValue)
		dataScope[varName] = parsedValue

		log.Printf("[CollectFenceData] Stored: %s = (type=%T) %v", varName, parsedValue, parsedValue)
	}

	// Process props
	for _, prop := range fence.Props {
		if _, exists := dataScope[prop.Name]; !exists {
			// Only add if not already provided in props
			if prop.DefaultValue != "" {
				log.Printf("[CollectFenceData] Processing prop default: %s = %q", prop.Name, prop.DefaultValue)

				// CRITICAL FIX: Use parseValue() for prop defaults too
				parsedValue := parseValue(prop.DefaultValue)
				// REVERTED: Store actual value, not __PROP__ string
				dataScope[prop.Name] = parsedValue

				log.Printf("[CollectFenceData] Stored prop: %s = (type=%T) %v", prop.Name, parsedValue, parsedValue)
			} else {
				dataScope[prop.Name] = nil
			}
		} else {
			log.Printf("[CollectFenceData] Skipping prop '%s' - already in dataScope (from props)", prop.Name)
		}
	}

	// Extract variables from raw content in the fence
	extractVariablesFromExpr(fence.RawContent, dataScope)

	// DIAGNOSTIC: Check if components still exists after fence processing
	if componentsRaw, ok := dataScope["components"]; ok {
		if components, ok := componentsRaw.([]interface{}); ok {
			log.Printf("[DIAGNOSTIC] CollectFenceData: ✓ 'components' STILL exists after fence processing (%d items)", len(components))
		}
	} else {
		log.Printf("[DIAGNOSTIC] CollectFenceData: ✗ 'components' LOST during fence processing!")
	}

	log.Printf("[DIAGNOSTIC] CollectFenceData: dataScope has %d keys AFTER processing fence", len(dataScope))
	log.Printf("[DIAGNOSTIC] ========== CollectFenceData END ==========")
}

// CreateChildScope creates a new scope that inherits from the parent scope
func CreateChildScope(parentScope map[string]any) map[string]any {
	childScope := make(map[string]any)

	// Copy all values from parent scope to child scope
	for key, value := range parentScope {
		childScope[key] = value
	}

	return childScope
}

// MergeScopes merges variables from child scope back to parent scope
// This is useful when we need to track variables defined in nested blocks
func MergeScopes(parentScope, childScope map[string]any) {
	// Only add variables that don't exist in parent scope
	// This prevents overwriting existing values in the parent scope
	for key, value := range childScope {
		if _, exists := parentScope[key]; !exists {
			parentScope[key] = value
		}
	}
}

// cloneScope creates a deep copy of the data scope map
// Used for creating iteration-specific scopes in loop expansion
// Pattern: Scope Cloning [Cognitive Load: 5]
func cloneScope(scope map[string]any) map[string]any {
	// Preallocate clone map (COGNITIVE LOAD RULE)
	clone := make(map[string]any, len(scope))

	// Copy all key-value pairs
	for key, value := range scope {
		clone[key] = value
	}

	return clone
}

// ============================================================================
// PHASE 1: Build-Time Loop Expansion - Collection Resolution
// ============================================================================

// resolveNestedProperty resolves nested property access (e.g., "category.items")
// into the actual value from dataScope.
//
// UPDATED: Now handles map[string]any (exported props) and map[string]interface{} (from JSON)
//
// Example:
//
//	dataScope := map[string]any{
//	    "category": map[string]any{
//	        "items": []string{"a", "b", "c"},
//	    },
//	}
//	value := resolveNestedProperty("category.items", dataScope)
//	// value = []string{"a", "b", "c"}
//
// Error cases:
// - Empty propertyPath → returns nil
// - Any part of the path not found → returns nil
// - Any intermediate value is not a map → returns nil
// - dataScope is nil → returns nil
func resolveNestedProperty(propertyPath string, dataScope map[string]any) any {
	// Split the property path (e.g., "category.items" → ["category", "items"])
	parts := strings.Split(propertyPath, ".")
	if len(parts) == 0 {
		return nil
	}

	// Start with the root value from dataScope
	var current any
	var exists bool

	current, exists = dataScope[parts[0]]
	if !exists {
		log.Printf("resolveNestedProperty: root property '%s' not found in dataScope", parts[0])
		return nil
	}

	// Traverse the nested properties
	for i := 1; i < len(parts); i++ {
		propName := parts[i]

		// Try map[string]any first (most common for exported props)
		if mapVal, ok := current.(map[string]any); ok {
			current, exists = mapVal[propName]
			if !exists {
				log.Printf("resolveNestedProperty: property '%s' not found in map at path '%s'",
					propName, strings.Join(parts[:i+1], "."))
				return nil
			}
			continue
		}

		// Try map[string]interface{} (from JSON unmarshaling)
		if mapVal, ok := current.(map[string]interface{}); ok {
			current, exists = mapVal[propName]
			if !exists {
				log.Printf("resolveNestedProperty: property '%s' not found in map at path '%s'",
					propName, strings.Join(parts[:i+1], "."))
				return nil
			}
			continue
		}

		// Not a map, cannot continue traversal
		log.Printf("resolveNestedProperty: intermediate value at '%s' is not a map (type=%T)",
			strings.Join(parts[:i], "."), current)
		return nil
	}

	return current
}

// resolveCollectionFromScope resolves a collection name (e.g., "components", "items")
// to its actual array value from the dataScope.
//
// This function is used for build-time loop expansion to resolve collection names
// (like "components") to their actual array values from dataScope.
//
// UPDATED: Now handles multiple slice types using reflection ([]string, []interface{}, etc.)
// UPDATED: Now supports nested property access (e.g., "category.items")
//
// Example:
//
//	dataScope := map[string]any{
//	    "components": []interface{}{
//	        map[string]any{"name": "Hero"},
//	        map[string]any{"name": "Footer"},
//	    },
//	    "items": []string{"one", "two", "three"},
//	    "category": map[string]any{
//	        "items": []string{"a", "b"},
//	    },
//	}
//	array, ok := resolveCollectionFromScope("components", dataScope)
//	// array contains the 2 components, ok = true
//	array, ok := resolveCollectionFromScope("items", dataScope)
//	// array contains ["one", "two", "three"] as []interface{}, ok = true
//	array, ok := resolveCollectionFromScope("category.items", dataScope)
//	// array contains ["a", "b"] as []interface{}, ok = true
//
// Error cases:
// - Collection name not found in dataScope → returns (nil, false)
// - Collection exists but is not an array/slice → returns (nil, false)
// - Collection is nil → returns (nil, false)
// - dataScope is nil → returns (nil, false)
func resolveCollectionFromScope(collectionName string, dataScope map[string]any) ([]interface{}, bool) {
	log.Printf("[DIAGNOSTIC] ========== resolveCollectionFromScope START ==========")
	log.Printf("[DIAGNOSTIC] resolveCollectionFromScope: looking for collection '%s'", collectionName)
	log.Printf("[DIAGNOSTIC] resolveCollectionFromScope: dataScope has %d keys", len(dataScope))

	// DIAGNOSTIC: Log all keys
	for key := range dataScope {
		log.Printf("[DIAGNOSTIC] resolveCollectionFromScope: - dataScope key: '%s'", key)
	}

	// Handle nil dataScope gracefully (COGNITIVE LOAD RULE: check nil)
	if dataScope == nil {
		log.Printf("resolveCollectionFromScope: dataScope is nil, cannot resolve '%s'", collectionName)
		return nil, false
	}

	var value any
	var exists bool

	// Check if this is a nested property access (e.g., "category.items")
	if strings.Contains(collectionName, ".") {
		log.Printf("[DIAGNOSTIC] resolveCollectionFromScope: detected nested property path '%s'", collectionName)
		value = resolveNestedProperty(collectionName, dataScope)
		exists = (value != nil)
		if !exists {
			log.Printf("resolveCollectionFromScope: nested property '%s' not found or is nil", collectionName)
			return nil, false
		}
	} else {
		// Simple property lookup
		value, exists = dataScope[collectionName]
		if !exists {
			// Log available keys for debugging
			availableKeys := make([]string, 0, len(dataScope))
			for key := range dataScope {
				availableKeys = append(availableKeys, key)
			}
			log.Printf("resolveCollectionFromScope: collection '%s' not found in dataScope, available keys: %v",
				collectionName, availableKeys)
			log.Printf("[DIAGNOSTIC] ========== resolveCollectionFromScope END (NOT FOUND) ==========")
			return nil, false
		}
	}

	// Handle nil value
	if value == nil {
		log.Printf("resolveCollectionFromScope: collection '%s' is nil", collectionName)
		return nil, false
	}

	// Try direct type assertion first (common case, most efficient)
	if array, ok := value.([]interface{}); ok {
		log.Printf("resolveCollectionFromScope: successfully resolved collection '%s' with %d items ([]interface{})",
			collectionName, len(array))
		log.Printf("[DIAGNOSTIC] ========== resolveCollectionFromScope END (SUCCESS) ==========")
		return array, true
	}

	// Use reflection to handle other slice types ([]string, []map[string]any, etc.)
	// Pattern: Reflection for Type Conversion [Cognitive Load: 8]
	valueType := reflect.TypeOf(value)
	if valueType.Kind() == reflect.Slice {
		// Convert any slice type to []interface{}
		valueSlice := reflect.ValueOf(value)
		length := valueSlice.Len()

		// Preallocate result (COGNITIVE LOAD RULE)
		result := make([]interface{}, length)

		// Copy elements to []interface{}
		for i := 0; i < length; i++ {
			result[i] = valueSlice.Index(i).Interface()
		}

		log.Printf("resolveCollectionFromScope: successfully resolved collection '%s' with %d items (%s)",
			collectionName, length, valueType.String())
		log.Printf("[DIAGNOSTIC] ========== resolveCollectionFromScope END (SUCCESS via reflection) ==========")
		return result, true
	}

	// Not a slice/array type
	log.Printf("resolveCollectionFromScope: collection '%s' is not an array, got type %T",
		collectionName, value)
	log.Printf("[DIAGNOSTIC] ========== resolveCollectionFromScope END (NOT AN ARRAY) ==========")
	return nil, false
}

// ============================================================================
// PHASE 2: Enhanced Scope Diffing Implementation
// ============================================================================

// DiffOptions controls scope diffing behavior
// Pattern: Configuration Struct [Cognitive Load: 3]
type DiffOptions struct {
	// If true, ignore variables that existed in parent scope but were modified in child scope
	IgnoreModifications bool

	// If true, only include variables that are likely to need x-data binding
	// (excludes variables that were never referenced in transformed nodes)
	OnlyUsedVariables bool
}

// DiffScopes compares parent and child scopes to identify newly added variables
// that need to be included in x-data bindings.
//
// This is used after transforming loop/conditional bodies to identify which variables
// from the body scope need to be added to the data scope for Alpine.js reactivity.
//
// Pattern: Scope Diffing for X-Data [Cognitive Load: 10]
//
// Example:
//
//	parentScope := map[string]any{"userName": "John", "userAge": 30}
//	childScope := map[string]any{"userName": "John", "userAge": 30, "greeting": "Hello"}
//	newVars := DiffScopes(parentScope, childScope, DiffOptions{})
//	// newVars = map[string]any{"greeting": "Hello"}
//
// Returns a map containing only variables that:
//  1. Exist in childScope
//  2. Don't exist in parentScope OR have different values (if IgnoreModifications=false)
func DiffScopes(parentScope, childScope map[string]any, options DiffOptions) map[string]any {
	// Preallocate result map (COGNITIVE LOAD RULE)
	newVars := make(map[string]any)

	// Iterate through child scope variables
	for key, childValue := range childScope {
		parentValue, existsInParent := parentScope[key]

		// Case 1: Variable doesn't exist in parent scope → it's new
		if !existsInParent {
			newVars[key] = childValue
			log.Printf("[DiffScopes] New variable found: %s = %v", key, childValue)
			continue
		}

		// Case 2: Variable exists in both scopes
		// If IgnoreModifications is true, skip checking for modifications
		if options.IgnoreModifications {
			continue
		}

		// Case 3: Variable was modified (different value in child scope)
		// Use deep equality check for complex types
		if !deepEqual(parentValue, childValue) {
			newVars[key] = childValue
			log.Printf("[DiffScopes] Modified variable found: %s: %v → %v", key, parentValue, childValue)
		}
	}

	log.Printf("[DiffScopes] Found %d new/modified variables", len(newVars))
	return newVars
}

// deepEqual performs deep equality comparison for scope values
// Uses JSON marshaling for accurate comparison of complex types
// Pattern: Deep Equality Check [Cognitive Load: 6]
func deepEqual(a, b any) bool {
	// Quick check for pointer equality or nil
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Try direct equality first (handles primitives, strings, etc.)
	if a == b {
		return true
	}

	// For complex types (maps, slices), use JSON marshaling for deep comparison
	// This is more reliable than reflect.DeepEqual for our use case
	aJSON, aErr := json.Marshal(a)
	bJSON, bErr := json.Marshal(b)

	// If marshaling fails for either value, fall back to string comparison
	if aErr != nil || bErr != nil {
		return false
	}

	// Compare JSON representations
	return string(aJSON) == string(bJSON)
}
