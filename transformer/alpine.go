package transformer

import (
	"encoding/json"
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

// isJavaScriptExpression checks if a string appears to be a JavaScript expression
// that should be evaluated by Alpine.js (not quoted as a string).
//
// It returns true for:
//   - Arithmetic expressions: age + 50, count * 2, total - discount
//   - Comparison expressions: count > 0, name !== "admin"
//   - Ternary operators: isActive ? "yes" : "no"
//   - Property access: user.name, items[0], obj.nested.prop
//   - Method calls: items.filter(), Math.random()
//   - Logical expressions: isValid && isActive, !disabled || isFree
//
// Cognitive Load: 8
func isJavaScriptExpression(s string) bool {
	trimmed := strings.TrimSpace(s)

	if len(trimmed) == 0 {
		return false
	}

	// If it starts/ends with quotes, it's a string literal, not an expression
	if (strings.HasPrefix(trimmed, `"`) && strings.HasSuffix(trimmed, `"`)) ||
		(strings.HasPrefix(trimmed, `'`) && strings.HasSuffix(trimmed, `'`)) ||
		(strings.HasPrefix(trimmed, "`") && strings.HasSuffix(trimmed, "`")) {
		return false
	}

	// FIX: If it has spaces but no function calls, it's likely a literal string, not an expression
	// Examples: "Custom Template Co." (literal), "Hello World" (literal)
	if strings.Contains(trimmed, " ") && !strings.Contains(trimmed, "(") {
		return false
	}

	// Check for arithmetic operators: + - * / %
	arithmeticOperators := []string{" + ", " - ", " * ", " / ", " % "}
	for _, op := range arithmeticOperators {
		if strings.Contains(trimmed, op) {
			return true
		}
	}

	// Check for comparison operators: > < >= <= == != === !==
	comparisonOperators := []string{" > ", " < ", " >= ", " <= ", " == ", " != ", " === ", " !== "}
	for _, op := range comparisonOperators {
		if strings.Contains(trimmed, op) {
			return true
		}
	}

	// Check for logical operators: && || !
	if strings.Contains(trimmed, " && ") || strings.Contains(trimmed, " || ") {
		return true
	}

	// Check for ternary operator: condition ? value1 : value2
	if strings.Contains(trimmed, "?") && strings.Contains(trimmed, ":") {
		return true
	}

	// CRITICAL FIX: Property/method access should only match valid JavaScript identifiers
	// Must start with identifier, then dot, then another identifier
	// Examples that should match: user.name, obj.prop, items[0].name
	// Examples that should NOT match: ./components/file.html, /path/to/file.js
	propertyAccessPattern := regexp.MustCompile(`^[a-zA-Z_$][a-zA-Z0-9_$]*\.`)
	if propertyAccessPattern.MatchString(trimmed) {
		return true
	}

	// Check for method calls: something()
	if strings.Contains(trimmed, "(") && strings.Contains(trimmed, ")") {
		return true
	}

	return false
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

// isQuotedString checks if a string is already a properly quoted JavaScript string literal.
// Returns true if the string starts and ends with matching quotes (", ', or `).
//
// Examples that return true:
//   - "\"hello\""
//   - "'world'"
//   - "`template`"
//   - "\"./components/file.html\""
//
// Cognitive Load: 3
func isQuotedString(s string) bool {
	trimmed := strings.TrimSpace(s)

	if len(trimmed) < 2 {
		return false
	}

	// Check for double quotes
	if strings.HasPrefix(trimmed, `"`) && strings.HasSuffix(trimmed, `"`) {
		return true
	}

	// Check for single quotes
	if strings.HasPrefix(trimmed, `'`) && strings.HasSuffix(trimmed, `'`) {
		return true
	}

	// Check for backticks (template literals)
	if strings.HasPrefix(trimmed, "`") && strings.HasSuffix(trimmed, "`") {
		return true
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
//   - Already quoted strings → returned as-is (no double-quoting)
//   - Regular strings → quoted and escaped with SINGLE quotes (for x-data attribute safety)
//   - Booleans → "true" or "false"
//   - Numbers → formatted as number string
//   - Arrays → recursively formatted
//   - Maps → formatted as JS object
//
// CRITICAL: JavaScript literals are returned AS-IS without modification.
// Alpine.js accepts BOTH JavaScript object syntax {key: value} and JSON {"key": "value"}
// We preserve the original JavaScript syntax to maintain expressions like ternaries.
//
// CRITICAL FIX: String values are quoted with SINGLE quotes (') instead of double quotes (")
// to prevent breaking HTML attributes when object literals are embedded in x-data="..."
//
// Cognitive Load: 16
func FormatGoValueToJS(value any) string {
	// Handle nil (COGNITIVE LOAD RULE: error handling)
	if value == nil {
		return "null"
	}

	switch v := value.(type) {
	case string:
		log.Printf("[DEBUG formatGoValueToJS] Processing string: %q (len=%d)", v, len(v))

		// CRITICAL FIX: Check if string is already quoted
		// If the fence section parser stored the value with quotes (like `"./components/UserProfile.html"`),
		// we should return it as-is, not add another layer of quotes
		if isQuotedString(v) {
			log.Printf("formatGoValueToJS: Detected already-quoted string, returning as-is: %s", v)
			return v
		}

		// Check if this string is a function definition
		if isFunctionExpression(v) {
			// Return function without quotes
			log.Printf("formatGoValueToJS: Detected function expression, returning as-is (length: %d)", len(v))
			return v
		}

		// CRITICAL FIX: Check if it's a JavaScript literal (array or object)
		// Return AS-IS without any conversion - Alpine.js will handle it
		// This preserves JavaScript syntax including ternaries, property access, etc.
		if isJavaScriptLiteral(v) {
			log.Printf("formatGoValueToJS: Detected JavaScript literal, returning as-is (length: %d, preview: %s...)",
				len(v),
				truncateString(v, 50))
			// Return the JavaScript literal unchanged
			// Alpine.js accepts JavaScript object syntax: {isLoggedIn: false, navItems: isLoggedIn ? [...] : [...]}
			return v
		}

		// CRITICAL FIX: Check if it's a JavaScript expression that needs to be evaluated
		// This includes arithmetic (age + 50), property access (user.name), etc.
		if isJavaScriptExpression(v) {
			log.Printf("formatGoValueToJS: Detected JavaScript expression, returning as-is: %s", v)
			log.Printf("[DEBUG] String '%s' classified as JavaScript expression!", v)
			return v
		}

		// CRITICAL FIX: Use SINGLE quotes instead of double quotes
		// This prevents breaking HTML attributes when embedded in x-data="{ ... }"
		// Regular string - add single quotes and escape (COGNITIVE LOAD RULE: proper escaping)
		escaped := strings.ReplaceAll(v, `\`, `\\`)
		escaped = strings.ReplaceAll(escaped, `'`, `\'`)
		// CRITICAL: Escape newlines and other control characters for HTML attributes
		escaped = strings.ReplaceAll(escaped, "\n", `\n`)
		escaped = strings.ReplaceAll(escaped, "\r", `\r`)
		escaped = strings.ReplaceAll(escaped, "\t", `\t`)
		result := fmt.Sprintf(`'%s'`, escaped)
		log.Printf("formatGoValueToJS: Regular string, quoting with single quotes: %s → %s", v, result)
		return result

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

	case []string:
		// Array of strings - quote each element
		elements := make([]string, 0, len(v))
		for _, item := range v {
			escaped := strings.ReplaceAll(item, `\`, `\\`)
			escaped = strings.ReplaceAll(escaped, `'`, `\'`)
			// Escape newlines and control characters
			escaped = strings.ReplaceAll(escaped, "\n", `\n`)
			escaped = strings.ReplaceAll(escaped, "\r", `\r`)
			escaped = strings.ReplaceAll(escaped, "\t", `\t`)
			elements = append(elements, fmt.Sprintf(`'%s'`, escaped))
		}
		return "[" + strings.Join(elements, ",") + "]"

	case []interface{}:
		// Array - format elements (COGNITIVE LOAD RULE: preallocate slice)
		elements := make([]string, 0, len(v))
		for _, item := range v {
			elements = append(elements, FormatGoValueToJS(item))
		}
		return "[" + strings.Join(elements, ",") + "]"

	case []map[string]any:
		// Array of objects - format each object
		elements := make([]string, 0, len(v))
		for _, item := range v {
			elements = append(elements, FormatGoValueToJS(item))
		}
		return "[" + strings.Join(elements, ",") + "]"

	case map[string]interface{}:
		// Object - format key-value pairs (COGNITIVE LOAD RULE: sorted keys for consistency)
		// NOTE: map[string]any is an alias for map[string]interface{} in Go 1.18+
		// This case handles both since they're the same underlying type
		m := v

		keys := make([]string, 0, len(m))
		for key := range m {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		pairs := make([]string, 0, len(m))
		for _, key := range keys {
			val := m[key]
			formattedValue := FormatGoValueToJS(val)

			// CRITICAL FIX: Quote keys that need quoting (contains special chars)
			// Valid identifiers can be unquoted: {name: "John"}
			// Special chars need quotes: {"pages/_index": "data", "my-key": "value"}
			formattedKey := key
			if !isValidJSIdentifier(key) {
				// Quote the key with single quotes (for HTML attribute safety)
				escaped := strings.ReplaceAll(key, `\`, `\\`)
				escaped = strings.ReplaceAll(escaped, `'`, `\'`)
				// Escape newlines and control characters in keys
				escaped = strings.ReplaceAll(escaped, "\n", `\n`)
				escaped = strings.ReplaceAll(escaped, "\r", `\r`)
				escaped = strings.ReplaceAll(escaped, "\t", `\t`)
				formattedKey = fmt.Sprintf(`'%s'`, escaped)
			}

			pairs = append(pairs, fmt.Sprintf(`%s:%s`, formattedKey, formattedValue))
		}
		return "{" + strings.Join(pairs, ",") + "}"

	default:
		// CRITICAL FIX: Use JSON encoding for unknown types instead of fmt.Sprintf("%v")
		// This prevents Go map syntax (map[key:value]) from appearing in output
		log.Printf("FormatGoValueToJS: WARNING - Unhandled type %T, using JSON encoding", value)
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			log.Printf("FormatGoValueToJS: ERROR - Failed to marshal %T: %v", value, err)
			// Fallback to string conversion as last resort
			str := fmt.Sprintf("%v", value)
			escaped := strings.ReplaceAll(str, `\`, `\\`)
			escaped = strings.ReplaceAll(escaped, `'`, `\'`)
			return fmt.Sprintf(`'%s'`, escaped)
		}
		// Convert double quotes to single quotes for HTML attribute safety
		result := strings.ReplaceAll(string(jsonBytes), `"`, `'`)
		log.Printf("FormatGoValueToJS: JSON-encoded %T to: %s", value, truncateString(result, 100))
		return result
	}
}

// isValidJSIdentifier checks if a string is a valid JavaScript identifier
// Valid identifiers can be used as unquoted object keys
// Pattern: ^[a-zA-Z_$][a-zA-Z0-9_$]*$
func isValidJSIdentifier(s string) bool {
	if s == "" {
		return false
	}

	// First character must be letter, underscore, or dollar sign
	first := rune(s[0])
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_' || first == '$') {
		return false
	}

	// Remaining characters can be letters, digits, underscore, or dollar sign
	for _, ch := range s[1:] {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '$') {
			return false
		}
	}

	return true
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

// extractDependencies extracts variable names referenced in a JavaScript expression
// Returns a map of variable names that this expression depends on
//
// Cognitive Load: 12
func extractDependencies(expr string) map[string]bool {
	deps := make(map[string]bool)

	// Skip simple values that have no dependencies
	trimmed := strings.TrimSpace(expr)
	if len(trimmed) == 0 {
		return deps
	}

	// Skip string literals
	if isQuotedString(trimmed) {
		return deps
	}

	// Skip boolean literals
	if trimmed == "true" || trimmed == "false" || trimmed == "null" {
		return deps
	}

	// Skip number literals
	if regexp.MustCompile(`^-?[0-9]+(\.[0-9]+)?$`).MatchString(trimmed) {
		return deps
	}

	// Extract variable references from JavaScript expressions
	// Patterns to match:
	// - user.name, user.role (object property access)
	// - isLoggedIn ? ... : ... (ternary conditional)
	// - items.filter(Boolean) (method calls)

	// Pattern: identifier followed by . (property access like user.name, user.role)
	propertyPattern := regexp.MustCompile(`\b([a-zA-Z_$][a-zA-Z0-9_$]*)\.[a-zA-Z_$]`)
	matches := propertyPattern.FindAllStringSubmatch(expr, -1)
	for _, match := range matches {
		if len(match) > 1 {
			varName := match[1]
			deps[varName] = true
			log.Printf("extractDependencies: Found property access dependency: %s", varName)
		}
	}

	// Pattern: standalone identifiers (variables like isLoggedIn, items)
	// This pattern matches identifiers that are NOT followed by ( (function calls)
	// and NOT preceded by . (property names)
	identifierPattern := regexp.MustCompile(`(?:^|[^a-zA-Z0-9_$.])\b([a-zA-Z_$][a-zA-Z0-9_$]*)\b(?:[^a-zA-Z0-9_$(.]|$)`)
	matches = identifierPattern.FindAllStringSubmatch(expr, -1)
	for _, match := range matches {
		if len(match) > 1 {
			varName := match[1]
			// Skip JavaScript keywords and built-in objects
			if !isJavaScriptKeyword(varName) {
				deps[varName] = true
				log.Printf("extractDependencies: Found identifier dependency: %s", varName)
			}
		}
	}

	return deps
}

// isJavaScriptKeyword checks if a string is a JavaScript keyword or built-in
func isJavaScriptKeyword(name string) bool {
	keywords := map[string]bool{
		// JavaScript keywords
		"if": true, "else": true, "for": true, "while": true, "do": true,
		"switch": true, "case": true, "default": true, "break": true, "continue": true,
		"return": true, "function": true, "var": true, "let": true, "const": true,
		"class": true, "extends": true, "static": true, "async": true, "await": true,
		"try": true, "catch": true, "finally": true, "throw": true,
		"new": true, "this": true, "super": true, "typeof": true, "instanceof": true,
		"in": true, "of": true, "delete": true, "void": true,

		// Literals
		"true": true, "false": true, "null": true, "undefined": true,

		// Built-in objects
		"Array": true, "Object": true, "String": true, "Number": true, "Boolean": true,
		"Date": true, "Math": true, "JSON": true, "RegExp": true,
		"Map": true, "Set": true, "Promise": true, "Symbol": true,
		"console": true, "window": true, "document": true,
	}
	return keywords[name]
}

// topologicalSort performs a topological sort on properties based on their dependencies
// Returns keys sorted so that dependencies come before dependents
//
// Cognitive Load: 18
func topologicalSort(dataScope map[string]any) []string {
	// Build dependency graph
	dependencies := make(map[string]map[string]bool)

	for key, value := range dataScope {
		// Convert value to string to analyze dependencies
		valueStr := ""
		if str, ok := value.(string); ok {
			valueStr = str
		}

		// Extract dependencies from the value
		deps := extractDependencies(valueStr)

		// Only keep dependencies that actually exist in dataScope
		filteredDeps := make(map[string]bool)
		for dep := range deps {
			if _, exists := dataScope[dep]; exists {
				filteredDeps[dep] = true
			}
		}

		dependencies[key] = filteredDeps

		if len(filteredDeps) > 0 {
			log.Printf("topologicalSort: %s depends on: %v", key, getKeys(filteredDeps))
		}
	}

	// Perform topological sort using Kahn's algorithm
	// 1. Find all nodes with no dependencies (in-degree = 0)
	// 2. Add them to result
	// 3. Remove them from dependency graph
	// 4. Repeat until all nodes processed

	result := make([]string, 0, len(dataScope))
	processed := make(map[string]bool)

	// Keep processing until all keys are added to result
	for len(result) < len(dataScope) {
		// Find keys with no unprocessed dependencies
		var noDeps []string
		for key := range dataScope {
			if processed[key] {
				continue
			}

			// Check if all dependencies are processed
			hasUnprocessedDep := false
			for dep := range dependencies[key] {
				if !processed[dep] {
					hasUnprocessedDep = true
					break
				}
			}

			if !hasUnprocessedDep {
				noDeps = append(noDeps, key)
			}
		}

		// If no keys found, we have a circular dependency
		// Fall back to alphabetical order for remaining keys
		if len(noDeps) == 0 {
			log.Printf("topologicalSort: Circular dependency detected, using alphabetical order for remaining keys")
			for key := range dataScope {
				if !processed[key] {
					noDeps = append(noDeps, key)
				}
			}
			sort.Strings(noDeps)
		} else {
			// Sort noDeps alphabetically for deterministic output
			sort.Strings(noDeps)
		}

		// Add to result and mark as processed
		for _, key := range noDeps {
			result = append(result, key)
			processed[key] = true
		}
	}

	log.Printf("topologicalSort: Final order: %v", result)
	return result
}

// getKeys returns a sorted slice of keys from a map
func getKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// hasSelfReferences checks if any property in dataScope references another property
// Returns true if there are self-references (needs function wrapper)
//
// Cognitive Load: 10
func hasSelfReferences(dataScope map[string]any) bool {
	for key, value := range dataScope {
		// Get value as string
		valueStr := ""
		if str, ok := value.(string); ok {
			valueStr = str
		}

		// Extract dependencies
		deps := extractDependencies(valueStr)

		// Check if this value references other properties in the same scope
		for dep := range deps {
			if dep != key && dataScope[dep] != nil {
				log.Printf("hasSelfReferences: Property '%s' references '%s' - needs function wrapper", key, dep)
				return true
			}
		}
	}

	return false
}

// alpineDataFormatter formats a data scope map into a JavaScript object literal
// suitable for Alpine.js x-data attributes.
//
// This function:
//   1. Filters out loop iterator variables (item, index, etc.) that shouldn't be in root scope
//   2. Ensures critical variables exist in the scope
//   3. Detects self-referencing properties
//   4. Wraps in function syntax if self-referencing detected
//   5. Sorts keys using topological sort (dependencies first)
//   6. Formats values using formatGoValueToJS which preserves JavaScript literals
//
// Pattern: Service Implementation Pattern [Load: 18]
// Cognitive Load: 18 (filtering: 3, self-ref detection: 4, sorting: 2, formatting: 6, string building: 3)
//
// Example without self-reference:
//   dataScope := map[string]any{
//     "name": "John",
//     "age": 30,
//   }
//   alpineDataFormatter(dataScope)
//   // Returns: {age:30,name:'John'}
//
// Example with self-reference:
//   dataScope := map[string]any{
//     "isLoggedIn": false,
//     "navItems": "isLoggedIn ? [...] : [...]",
//   }
//   alpineDataFormatter(dataScope)
//   // Returns: () => { const isLoggedIn = false; const navItems = isLoggedIn ? [...] : [...]; return {isLoggedIn,navItems}; }
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

	// CRITICAL: Detect if we need function wrapper for self-referencing
	needsFunctionWrapper := hasSelfReferences(dataScope)

	// Use topological sort to ensure dependencies come first
	keys := topologicalSort(dataScope)

	if needsFunctionWrapper {
		// Generate function syntax: () => { const k1 = v1; const k2 = v2; return {k1, k2}; }
		log.Printf("alpineDataFormatter: Self-references detected, using function wrapper syntax")

		// Build const declarations (COGNITIVE LOAD RULE: preallocate)
		declarations := make([]string, 0, len(dataScope))
		returnProps := make([]string, 0, len(dataScope))

		for _, key := range keys {
			// Skip internal Alpine.js variables
			if strings.HasPrefix(key, "$") {
				continue
			}

			value := dataScope[key]
			formattedValue := FormatGoValueToJS(value)

			// Build: const key = value;
			declarations = append(declarations, fmt.Sprintf(`const %s = %s;`, key, formattedValue))

			// Add to return object properties
			returnProps = append(returnProps, key)
		}

		// Build function: () => { const x = y; const z = w; return {x, z}; }
		result := fmt.Sprintf(`() => { %s return {%s}; }`,
			strings.Join(declarations, " "),
			strings.Join(returnProps, ","))

		log.Printf("Generated x-data function wrapper: %s", truncateString(result, 200))
		return result
	}

	// No self-references - use object literal syntax
	// Build object literal using formatGoValueToJS (COGNITIVE LOAD RULE: preallocate)
	parts := make([]string, 0, len(dataScope))
	for _, key := range keys {
		// Skip internal Alpine.js variables
		if strings.HasPrefix(key, "$") {
			continue
		}

		value := dataScope[key]

		// Format value based on type - handle dynamic expressions specially
		var formattedValue string
		if strVal, ok := value.(string); ok {
			// CRITICAL: Check if this has the __VAR_REF__ marker
			// This indicates a variable reference that should be output without quotes
			if strings.HasPrefix(strVal, "__VAR_REF__") {
				// Strip the marker and output as unquoted variable reference
				varName := strings.TrimPrefix(strVal, "__VAR_REF__")
				formattedValue = varName
				log.Printf("alpineDataFormatter: Stripped __VAR_REF__ marker, outputting variable reference: %s", varName)
			} else if isDynamicExpression(strVal, dataScope) {
				// Don't quote dynamic expressions - Alpine.js will evaluate them
				formattedValue = strVal
			} else {
				// Regular string - use formatGoValueToJS to properly quote it
				formattedValue = FormatGoValueToJS(value)
			}
		} else {
			// Non-string values - use formatGoValueToJS
			formattedValue = FormatGoValueToJS(value)
		}

		// Build key-value pair with unquoted keys (JavaScript object syntax)
		parts = append(parts, fmt.Sprintf(`%s:%s`, key, formattedValue))
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
	// Format the data scope as JavaScript object literal or function
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
