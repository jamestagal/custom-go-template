package transformer

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"

	"github.com/jimafisk/custom_go_template/ast"
)

// Component tracking to prevent duplicate components
var componentRegistry = make(map[string]bool)

// Reset component tracking for each transformation
func resetComponentTracking() {
	componentRegistry = make(map[string]bool)
}

// isFunctionExpression checks if a string contains a JavaScript function definition.
// It returns true if the string represents any form of JavaScript function.
//
// Supported patterns:
//   - Function declarations: function name() {} or function() {}
//   - Generator functions: function* name() {}
//   - Arrow functions: () => {}, (x) => {}, x => {}
//   - Async functions: async function name() {} or async () => {}
//   - Getters/setters: get name() {} or set name(v) {}
//   - Method shorthand: name() { ... }
//
// Examples that return true:
//   - "function greet() { return 'hello'; }"
//   - "function() { return 42; }"
//   - "() => { return 42; }"
//   - "(x) => x * 2"
//   - "x => x * 2"
//   - "async function fetch() {}"
//   - "async () => {}"
//   - "get value() { return this._value; }"
//   - "set value(v) { this._value = v; }"
//   - "greet() { return 'hello'; }"
//   - "function* generator() {}"
//
// Examples that return false:
//   - "\"hello\""
//   - "42"
//   - "true"
//   - "[1,2,3]"
//   - "{key: 'value'}"
//   - "userName" (simple variable)
//
// Cognitive Load: 8
func isFunctionExpression(expr string) bool {
	expr = strings.TrimSpace(expr)

	if len(expr) == 0 {
		return false
	}

	// Pattern 1: function declarations - function name() {} or function() {}
	if strings.HasPrefix(expr, "function") {
		return true
	}

	// Pattern 2: generator functions - function* name() {}
	if strings.HasPrefix(expr, "function*") {
		return true
	}

	// Pattern 3: arrow functions - () => {} or (x) => {} or x => {}
	if strings.Contains(expr, "=>") {
		return true
	}

	// Pattern 4: async functions - async function name() {} or async () => {}
	if strings.HasPrefix(expr, "async") {
		return true
	}

	// Pattern 5: getters/setters - get name() {} or set name(v) {}
	if strings.HasPrefix(expr, "get ") || strings.HasPrefix(expr, "set ") {
		return true
	}

	// Pattern 6: method shorthand - name() { ... }
	// Look for: identifier followed by ( with balanced braces
	methodShorthandRegex := regexp.MustCompile(`^[a-zA-Z_$][a-zA-Z0-9_$]*\s*\([^)]*\)\s*\{`)
	if methodShorthandRegex.MatchString(expr) {
		return true
	}

	return false
}

// isComplexJSObject checks if a string appears to be a complex JavaScript object
// with methods or nested structures that should be preserved as-is.
//
// Cognitive Load: 4
func isComplexJSObject(s string) bool {
	trimmed := strings.TrimSpace(s)

	// Check if it's an object literal
	if !strings.HasPrefix(trimmed, "{") || !strings.HasSuffix(trimmed, "}") {
		return false
	}

	// Check for function keywords which indicate methods
	if strings.Contains(trimmed, "function") || strings.Contains(trimmed, "=>") {
		return true
	}

	// Check for method shorthand pattern: name() { or name(params) {
	methodPattern := regexp.MustCompile(`[a-zA-Z_$][a-zA-Z0-9_$]*\s*\([^)]*\)\s*\{`)
	return methodPattern.MatchString(trimmed)
}

// isJavaScriptLiteral checks if a string appears to be a JavaScript array or object literal.
// This function determines if a value should be output as raw JavaScript without quotes.
//
// It returns true for:
//   - Arrays: [1, 2, 3] or [{ label: "Home", url: "/" }]
//   - Objects: { key: "value" } or { name: "John", age: 30 }
//   - Expressions: new Date().getFullYear()
//
// Cognitive Load: 6
func isJavaScriptLiteral(s string) bool {
	trimmed := strings.TrimSpace(s)

	if len(trimmed) == 0 {
		return false
	}

	// Check for array literals: starts with [ and ends with ]
	if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
		return true
	}

	// Check for object literals: starts with { and ends with }
	if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
		return true
	}

	// Check for common JavaScript expressions (like new Date())
	jsExpressionPrefixes := []string{
		"new ",
		"Date(",
		"Math.",
		"JSON.",
		"Array.",
		"Object.",
	}
	for _, prefix := range jsExpressionPrefixes {
		if strings.HasPrefix(trimmed, prefix) {
			return true
		}
	}

	return false
}

// formatGoValueToJS converts a Go value to JavaScript literal syntax.
// Functions are returned without quotes, strings are quoted and escaped.
//
// Handles:
//   - nil → "null"
//   - Function strings → returned as-is (no quotes)
//   - JavaScript literals (arrays/objects) → returned as-is (no quotes)
//   - Regular strings → quoted and escaped with double quotes
//   - Booleans → "true" or "false"
//   - Numbers → formatted as number string
//   - Arrays → recursively formatted
//   - Maps → formatted as JS object
//
// CRITICAL FIX: JavaScript arrays and objects (even without functions) are now
// returned as-is without quotes to preserve multi-line prop values.
//
// Cognitive Load: 16
func formatGoValueToJS(value any) string {
	// Handle nil (COGNITIVE LOAD RULE: error handling)
	if value == nil {
		return "null"
	}

	switch v := value.(type) {
	case string:
		// Check if this string is a function definition
		if isFunctionExpression(v) {
			// Return function without quotes
			log.Printf("formatGoValueToJS: Detected function expression, returning as-is (length: %d)", len(v))
			return v
		}

		// CRITICAL FIX: Check if it's a JavaScript literal (array or object)
		// This preserves multi-line prop values like:
		//   links = [{ label: "Home", url: "/" }, ...]
		//   stats = { users: 124, products: 56 }
		if isJavaScriptLiteral(v) {
			log.Printf("formatGoValueToJS: Detected JavaScript literal, returning as-is (length: %d, preview: %s...)",
				len(v),
				truncateString(v, 50))
			return v
		}

		// Regular string - add double quotes and escape (COGNITIVE LOAD RULE: proper escaping)
		escaped := strings.ReplaceAll(v, `\`, `\\`)
		escaped = strings.ReplaceAll(escaped, `"`, `\"`)
		return fmt.Sprintf(`"%s"`, escaped)

	case bool:
		if v {
			return "true"
		}
		return "false"

	case int:
		return fmt.Sprintf("%d", v)
	case int8:
		return fmt.Sprintf("%d", v)
	case int16:
		return fmt.Sprintf("%d", v)
	case int32:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)

	case uint:
		return fmt.Sprintf("%d", v)
	case uint8:
		return fmt.Sprintf("%d", v)
	case uint16:
		return fmt.Sprintf("%d", v)
	case uint32:
		return fmt.Sprintf("%d", v)
	case uint64:
		return fmt.Sprintf("%d", v)

	case float32:
		return fmt.Sprintf("%v", v)
	case float64:
		return fmt.Sprintf("%v", v)

	case []interface{}:
		// Array - format elements (COGNITIVE LOAD RULE: preallocate slice)
		elements := make([]string, 0, len(v))
		for _, item := range v {
			elements = append(elements, formatGoValueToJS(item))
		}
		return "[" + strings.Join(elements, ",") + "]"

	case []map[string]any:
		// Array of objects - format each object
		elements := make([]string, 0, len(v))
		for _, item := range v {
			elements = append(elements, formatGoValueToJS(item))
		}
		return "[" + strings.Join(elements, ",") + "]"

	case map[string]any:
		// Object - format key-value pairs (COGNITIVE LOAD RULE: sorted keys for consistency)
		// Note: map[string]any is the same as map[string]interface{}
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		pairs := make([]string, 0, len(v))
		for _, key := range keys {
			value := v[key]
			formattedValue := formatGoValueToJS(value)
			// Always quote keys for valid JavaScript syntax
			pairs = append(pairs, fmt.Sprintf(`"%s":%s`, key, formattedValue))
		}
		return "{" + strings.Join(pairs, ",") + "}"

	default:
		// Fallback for unknown types - convert to string and quote
		// This should rarely be hit in normal operation
		str := fmt.Sprintf("%v", value)
		escaped := strings.ReplaceAll(str, `\`, `\\`)
		escaped = strings.ReplaceAll(escaped, `"`, `\"`)
		return fmt.Sprintf(`"%s"`, escaped)
	}
}

// truncateString truncates a string to maxLen characters with ellipsis
// Helper function for logging
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// isDefaultPlaceholder checks if a value is a default placeholder that should be removed
func isDefaultPlaceholder(value any) bool {
	if value == nil {
		return true
	}

	// Check for empty string
	if str, ok := value.(string); ok {
		return strings.TrimSpace(str) == ""
	}

	return false
}

// alpineDataFormatter formats a data scope map into a JavaScript object literal
// suitable for Alpine.js x-data attributes.
//
// This function:
//   1. Filters out loop iterator variables (item, index, etc.) that shouldn't be in root scope
//   2. Ensures critical variables exist in the scope
//   3. Sorts keys for deterministic output
//   4. Formats values using formatGoValueToJS which preserves JavaScript literals
//
// Pattern: Service Implementation Pattern [Load: 12]
// Cognitive Load: 12 (filtering: 3, sorting: 2, formatting: 5, string building: 2)
//
// Example:
//   dataScope := map[string]any{
//     "name": "John",
//     "age": 30,
//     "links": "[{ label: \"Home\", url: \"/\" }]",
//   }
//   alpineDataFormatter(dataScope)
//   // Returns: {"age":30,"links":[{ label: "Home", url: "/" }],"name":"John"}
//
// Note: The formatGoValueToJS function now correctly preserves JavaScript arrays
// and objects without quoting them, fixing the multi-line prop truncation issue.
func alpineDataFormatter(dataScope map[string]any) string {
	// Clean up any loop iterator variables that might have leaked
	// Common iterator names that should never be in root scope
	iteratorNames := []string{"item", "index", "key", "value", "i", "idx"}
	for _, name := range iteratorNames {
		// Only remove if the value is nil or looks like a default/placeholder
		if val, exists := dataScope[name]; exists && (val == nil || isDefaultPlaceholder(val)) {
			delete(dataScope, name)
		}
	}

	// Ensure critical variables exist
	ensureCriticalVariables(dataScope)

	// Sort keys for consistent output (COGNITIVE LOAD RULE: deterministic behavior)
	keys := make([]string, 0, len(dataScope))
	for key := range dataScope {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Build object literal using formatGoValueToJS (COGNITIVE LOAD RULE: preallocate)
	parts := make([]string, 0, len(dataScope))
	for _, key := range keys {
		// Skip internal Alpine.js variables
		if strings.HasPrefix(key, "$") {
			continue
		}

		value := dataScope[key]

		// Use the helper to format the value
		// CRITICAL: formatGoValueToJS now preserves JavaScript literals
		formattedValue := formatGoValueToJS(value)

		// Build key-value pair (always quote keys for valid JSON-like syntax)
		parts = append(parts, fmt.Sprintf(`"%s":%s`, key, formattedValue))
	}

	result := "{" + strings.Join(parts, ",") + "}"
	log.Printf("Generated x-data object literal: %s", truncateString(result, 200))
	return result
}

// ensureCriticalVariables ensures that critical variables for conditionals and loops are properly initialized
func ensureCriticalVariables(dataScope map[string]any) {
	// Handle nil dataScope - nothing to ensure
	if dataScope == nil {
		return
	}

	// SIMPLIFIED: Don't add any extra variables
	// Variables should only be added when explicitly referenced in the template
	// This prevents pollution of the data scope with unused variables
}

// wrapWithAlpineData wraps the given nodes with an Alpine.js x-data wrapper
// This creates the wrapper element with formatted data scope
func wrapWithAlpineData(nodes []ast.Node, dataScope map[string]any) *ast.Element {
	// Format the data scope as JavaScript object literal
	dataJSON := alpineDataFormatter(dataScope)

	// Create the wrapper element
	wrapper := &ast.Element{
		TagName: "div",
		Attributes: []ast.Attribute{
			{
				Name:       "x-data",
				Value:      dataJSON,
				IsAlpine:   true,
				AlpineType: "data",
			},
		},
		Children:    nodes,
		SelfClosing: false,
	}

	return wrapper
}

// ensureVariablesInScope ensures all referenced variables exist in the data scope
// This is critical for Alpine.js to work correctly with expressions
func ensureVariablesInScope(nodes []ast.Node, dataScope map[string]any) {
	for _, node := range nodes {
		switch n := node.(type) {
		case *ast.ExpressionNode:
			// Add variables from expressions
			extractVariablesFromExpr(n.Expression, dataScope)

		case *ast.Element:
			// Check attributes for expressions
			for _, attr := range n.Attributes {
				if attr.Dynamic || attr.IsAlpine {
					extractVariablesFromExpr(attr.Value, dataScope)
				}
			}

			// Recursively process children
			ensureVariablesInScope(n.Children, dataScope)

		case *ast.Conditional:
			// Add variables from condition
			extractVariablesFromExpr(n.IfCondition, dataScope)

			// Process branches
			ensureVariablesInScope(n.IfContent, dataScope)
			for i, condition := range n.ElseIfConditions {
				extractVariablesFromExpr(condition, dataScope)
				if i < len(n.ElseIfContent) {
					ensureVariablesInScope(n.ElseIfContent[i], dataScope)
				}
			}
			ensureVariablesInScope(n.ElseContent, dataScope)

		case *ast.Loop:
			// Add loop variables
			dataScope[n.Iterator] = nil
			if n.Value != "" {
				dataScope[n.Value] = nil
			}

			// Add array variable
			extractVariablesFromExpr(n.Collection, dataScope)

			// Process loop body
			ensureVariablesInScope(n.Content, dataScope)
		}
	}
}
