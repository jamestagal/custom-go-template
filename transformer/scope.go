package transformer

import (
	"encoding/json"
	"log"
	"reflect"
	"strings"
	"sync"

	"github.com/jimafisk/custom_go_template/ast"
)

// ============================================================================
// TransformContext: Per-request state for concurrent safety
// ============================================================================

// TransformContext holds per-request transformation state.
// This replaces global variables to enable safe concurrent HTTP request handling.
//
// Pattern: Context Passing for Concurrency Safety [Cognitive Load: 5]
//
// Usage:
//
//	ctx := NewTransformContext(props)
//	transformedAST := TransformASTWithContext(template, ctx)
//	filteredScope := ctx.RuntimeTracker.FilterScope(dataScope)
type TransformContext struct {
	// RuntimeTracker tracks variables that need runtime evaluation by Alpine.js
	RuntimeTracker *RuntimeVarTracker

	// DataScope holds the current variable scope for transformation
	DataScope map[string]any

	// mu protects concurrent access to context fields (if needed for sub-goroutines)
	mu sync.RWMutex
}

// NewTransformContext creates a new transformation context with initialized state
func NewTransformContext(props map[string]any) *TransformContext {
	return &TransformContext{
		RuntimeTracker: NewRuntimeVarTracker(),
		DataScope:      InitDataScope(props),
	}
}

// GetRuntimeTracker returns the runtime tracker from this context.
// Thread-safe for concurrent read access.
func (ctx *TransformContext) GetRuntimeTracker() *RuntimeVarTracker {
	if ctx == nil {
		return nil
	}
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.RuntimeTracker
}

// GetDataScope returns a copy of the current data scope.
// Thread-safe for concurrent read access.
func (ctx *TransformContext) GetDataScope() map[string]any {
	if ctx == nil {
		return nil
	}
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	// Return a copy to prevent concurrent modification
	copy := make(map[string]any, len(ctx.DataScope))
	for k, v := range ctx.DataScope {
		copy[k] = v
	}
	return copy
}

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

// ============================================================================
// PHASE 3: Runtime Variable Tracking for x-data Optimization
// ============================================================================

// RuntimeVarTracker tracks variables that need runtime evaluation by Alpine.js.
// Variables that are only used at build-time (e.g., allContent for loop expansion)
// should NOT be included in x-data to reduce page weight.
//
// Pattern: Runtime Variable Tracking [Cognitive Load: 8]
//
// Usage:
//
//	tracker := NewRuntimeVarTracker()
//	// During transformation, track variables that remain in Alpine directives
//	tracker.Track("photos")      // x-for="photo in photos"
//	tracker.Track("published")   // x-if="published"
//
//	// When serializing x-data, filter to only tracked variables
//	if tracker.IsTracked("allContent") {
//	    // Include in x-data
//	} else {
//	    // Exclude - was only used at build-time
//	}
type RuntimeVarTracker struct {
	vars map[string]bool
}

// NewRuntimeVarTracker creates a new tracker for runtime variables
func NewRuntimeVarTracker() *RuntimeVarTracker {
	return &RuntimeVarTracker{vars: make(map[string]bool)}
}

// Track records a variable as needing runtime evaluation.
// Only the root variable name is tracked (e.g., "photos" from "photos.length").
func (t *RuntimeVarTracker) Track(varName string) {
	if varName == "" || t == nil {
		return
	}

	// Only track the root variable (before any dots)
	// e.g., "photos.length" → "photos"
	root := strings.SplitN(varName, ".", 2)[0]

	// Skip common loop iterator names that shouldn't be in x-data
	// These are temporary variables from x-for directives
	skipNames := map[string]bool{
		"item": true, "index": true, "key": true, "value": true,
		"i": true, "idx": true, "k": true, "v": true,
		"component": true, "post": true, "photo": true,
	}
	if skipNames[root] {
		return
	}

	t.vars[root] = true
	log.Printf("[RuntimeVarTracker] Tracking variable: %s (from %s)", root, varName)
}

// TrackExpression extracts and tracks all variables from an expression.
// Handles expressions like "photos && photos.length > 0" → tracks "photos"
func (t *RuntimeVarTracker) TrackExpression(expr string) {
	if expr == "" || t == nil {
		return
	}

	// Extract variable-like tokens from the expression
	// This is a simplified approach that handles most common cases
	tokens := extractVariableTokens(expr)
	for _, token := range tokens {
		t.Track(token)
	}
}

// IsTracked checks if a variable was tracked for runtime evaluation
func (t *RuntimeVarTracker) IsTracked(varName string) bool {
	if t == nil {
		return true // If no tracker, include everything (safe default)
	}
	return t.vars[varName]
}

// GetTrackedVars returns all tracked variable names
func (t *RuntimeVarTracker) GetTrackedVars() []string {
	if t == nil {
		return nil
	}
	result := make([]string, 0, len(t.vars))
	for v := range t.vars {
		result = append(result, v)
	}
	return result
}

// Snapshot returns a copy of the current tracker state.
// Used to save state before component transformation.
func (t *RuntimeVarTracker) Snapshot() map[string]bool {
	if t == nil {
		return make(map[string]bool)
	}
	snapshot := make(map[string]bool, len(t.vars))
	for k, v := range t.vars {
		snapshot[k] = v
	}
	return snapshot
}

// NewVarsSince returns variables tracked since a snapshot was taken.
// This allows per-component tracking by comparing before/after states.
func (t *RuntimeVarTracker) NewVarsSince(snapshot map[string]bool) []string {
	if t == nil {
		return nil
	}
	var newVars []string
	for v := range t.vars {
		if !snapshot[v] {
			newVars = append(newVars, v)
		}
	}
	return newVars
}

// FilterScopeWithSnapshot filters a dataScope using only variables tracked since snapshot.
// This enables per-component x-data optimization.
func (t *RuntimeVarTracker) FilterScopeWithSnapshot(dataScope map[string]any, snapshot map[string]bool) map[string]any {
	if t == nil {
		return dataScope
	}

	// Find variables tracked since snapshot
	localVars := make(map[string]bool)
	for v := range t.vars {
		if !snapshot[v] {
			localVars[v] = true
		}
	}

	if len(localVars) == 0 {
		// No new variables tracked during this component's transformation
		// This means all expressions were resolved at build-time!
		log.Printf("[RuntimeVarTracker] FilterScopeWithSnapshot: no new variables since snapshot (all build-time resolved)")
		return make(map[string]any) // Return empty scope
	}

	filtered := make(map[string]any)

	// First pass: include directly tracked variables (since snapshot)
	for key, value := range dataScope {
		if localVars[key] {
			filtered[key] = value
		}
	}

	// Second pass: for any getters/setters in filtered, also include their 'this.*' dependencies
	additionalDeps := make(map[string]bool)
	for key, value := range filtered {
		if strVal, ok := value.(string); ok {
			trimmed := strings.TrimSpace(strVal)
			// Check if this is a getter or setter definition
			if (strings.HasPrefix(trimmed, "get ") || strings.HasPrefix(trimmed, "set ")) &&
				strings.Contains(trimmed, "(") && strings.Contains(trimmed, "{") {
				// Extract 'this.*' references from the getter/setter body
				deps := extractThisReferences(strVal)
				for _, dep := range deps {
					if _, exists := dataScope[dep]; exists && !localVars[dep] {
						additionalDeps[dep] = true
						log.Printf("[RuntimeVarTracker] Getter/setter '%s' depends on '%s' - adding to filtered scope", key, dep)
					}
				}
			}
		}
	}

	// Add the additional dependencies
	for dep := range additionalDeps {
		filtered[dep] = dataScope[dep]
	}

	log.Printf("[RuntimeVarTracker] FilterScopeWithSnapshot: %d → %d variables (local: %d, getter deps: %d)",
		len(dataScope), len(filtered), len(localVars), len(additionalDeps))

	return filtered
}

// FilterScope filters a dataScope to only include tracked runtime variables.
// This is called before serializing to x-data to reduce page weight.
// IMPORTANT: Also includes dependencies of getters/setters (things they reference via this.*)
func (t *RuntimeVarTracker) FilterScope(dataScope map[string]any) map[string]any {
	if t == nil || len(t.vars) == 0 {
		// No tracking info, return original scope
		return dataScope
	}

	filtered := make(map[string]any)

	// First pass: include directly tracked variables
	for key, value := range dataScope {
		if t.IsTracked(key) {
			filtered[key] = value
		}
	}

	// Second pass: for any getters/setters in filtered, also include their 'this.*' dependencies
	// This ensures getters like formattedJoinDate get their dependencies (user, formatDate, etc.)
	additionalDeps := make(map[string]bool)
	for key, value := range filtered {
		if strVal, ok := value.(string); ok {
			trimmed := strings.TrimSpace(strVal)
			// Check if this is a getter or setter definition
			if (strings.HasPrefix(trimmed, "get ") || strings.HasPrefix(trimmed, "set ")) &&
				strings.Contains(trimmed, "(") && strings.Contains(trimmed, "{") {
				// Extract 'this.*' references from the getter/setter body
				deps := extractThisReferences(strVal)
				for _, dep := range deps {
					if _, exists := dataScope[dep]; exists && !t.IsTracked(dep) {
						additionalDeps[dep] = true
						log.Printf("[RuntimeVarTracker] Getter/setter '%s' depends on '%s' - adding to filtered scope", key, dep)
					}
				}
			}
		}
	}

	// Add the additional dependencies
	for dep := range additionalDeps {
		filtered[dep] = dataScope[dep]
	}

	log.Printf("[RuntimeVarTracker] FilterScope: %d → %d variables (removed %d build-time-only, added %d getter deps)",
		len(dataScope), len(filtered), len(dataScope)-len(filtered)+len(additionalDeps), len(additionalDeps))

	return filtered
}

// extractThisReferences extracts variable names referenced via 'this.*' in a string
// For example, "this.formatDate(this.user.joinDate)" returns ["formatDate", "user"]
func extractThisReferences(s string) []string {
	var refs []string
	seen := make(map[string]bool)

	// Simple pattern matching for this.varName
	for i := 0; i < len(s)-5; i++ {
		if s[i:i+5] == "this." {
			// Extract the variable name after "this."
			start := i + 5
			end := start
			for end < len(s) && (isAlphaNumeric(s[end]) || s[end] == '_') {
				end++
			}
			if end > start {
				varName := s[start:end]
				if !seen[varName] {
					seen[varName] = true
					refs = append(refs, varName)
				}
			}
		}
	}

	return refs
}

// isAlphaNumeric checks if a byte is a letter or digit
func isAlphaNumeric(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}

// extractVariableTokens extracts variable-like tokens from an expression.
// Returns tokens that look like variable names (start with letter/underscore).
// Pattern: Token Extraction [Cognitive Load: 8]
func extractVariableTokens(expr string) []string {
	var tokens []string
	var current strings.Builder

	for i := 0; i < len(expr); i++ {
		ch := expr[i]

		// Start of a potential identifier
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' || ch == '$' {
			current.WriteByte(ch)

			// Continue reading the identifier
			for i+1 < len(expr) {
				next := expr[i+1]
				if (next >= 'a' && next <= 'z') || (next >= 'A' && next <= 'Z') ||
					(next >= '0' && next <= '9') || next == '_' || next == '.' {
					i++
					current.WriteByte(next)
				} else {
					break
				}
			}

			token := current.String()
			current.Reset()

			// Skip JavaScript keywords and literals
			skipKeywords := map[string]bool{
				"true": true, "false": true, "null": true, "undefined": true,
				"this": true, "new": true, "typeof": true, "instanceof": true,
				"if": true, "else": true, "for": true, "while": true,
				"return": true, "function": true, "const": true, "let": true, "var": true,
				"$store": true, "$el": true, "$refs": true, "$data": true,
			}

			// Skip if it's a keyword or starts with $store (Alpine magic)
			if !skipKeywords[token] && !strings.HasPrefix(token, "$store.") {
				tokens = append(tokens, token)
			}
		}
	}

	return tokens
}
