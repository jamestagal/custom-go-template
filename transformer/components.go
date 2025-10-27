package transformer

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/jimafisk/custom_go_template/ast"
)

// ComponentTemplate represents a registered component template
type ComponentTemplate struct {
	Name     string
	Template *ast.Template
	Props    []string // List of prop names this component accepts
}

// componentTemplateRegistry stores registered component templates
var componentTemplateRegistry = make(map[string]*ComponentTemplate)

// resetComponentTemplateRegistry resets the component template registry
// Note: This doesn't clear registered templates, only the instance tracking
func resetComponentTemplateRegistry() {
	// In a more complex implementation, we might want to clear
	// certain aspects of the registry but keep the templates
	log.Printf("Component template registry reset")
}

// RegisterComponent registers a component template for later use
func RegisterComponent(name string, template *ast.Template, props []string) {
	componentTemplateRegistry[name] = &ComponentTemplate{
		Name:     name,
		Template: template,
		Props:    props,
	}
	log.Printf("Registered component template: %s with %d props", name, len(props))
}

// UnregisterComponent removes a component from the registry
// Used for test cleanup to prevent test interference
func UnregisterComponent(name string) {
	delete(componentTemplateRegistry, name)
	log.Printf("Unregistered component template: %s", name)
}

// GetComponentTemplate retrieves a component template by name
// Supports case-insensitive lookup to match JSON component names like "hero2436" with registered "Hero2436"
func GetComponentTemplate(name string) (*ComponentTemplate, bool) {
	// Try exact match first (most common case)
	template, exists := componentTemplateRegistry[name]
	if exists {
		return template, exists
	}

	// Try case-insensitive match by capitalizing first letter
	// This handles: "hero2436" → "Hero2436", "footer" → "Footer"
	if len(name) > 0 {
		capitalizedName := strings.ToUpper(name[:1]) + name[1:]
		template, exists = componentTemplateRegistry[capitalizedName]
		if exists {
			log.Printf("[GetComponentTemplate] Found component via case-insensitive match: %q → %q", name, capitalizedName)
			return template, exists
		}
	}

	// Not found with any strategy
	return nil, false
}

// GetAllRegisteredKeys returns all registered component template keys for debugging
func GetAllRegisteredKeys() []string {
	keys := make([]string, 0, len(componentTemplateRegistry))
	for key := range componentTemplateRegistry {
		keys = append(keys, key)
	}
	return keys
}


// GetAllComponentNames returns a map of all component names for magic variables
// TASK 5.4: Support allLayouts magic variable
//
// Pattern: Helper Function [Load: 3]
// Cognitive Load: 3 (iterate registry)
func GetAllComponentNames() map[string]bool {
	result := make(map[string]bool)
	for key := range componentTemplateRegistry {
		result[key] = true
	}
	return result
}
// isStructuralTag checks if a tag should skip x-data wrapping
//
// Pattern: Helper Function [Load: 3]
// Cognitive Load: 3 (simple map lookup)
//
// These are HTML structural/metadata tags that should never be reactive.
// Adding x-data to these tags causes Alpine.js to try parsing their content
// as reactive code, which breaks meta tags and other structural elements.
//
// Structural tags that should NEVER get x-data:
//   - html: Root document element
//   - head: Document metadata section
//   - body: Document content section (x-data added by server if needed)
//   - !doctype: Document type declaration
//
// Example:
//   isStructuralTag("head")    // Returns: true
//   isStructuralTag("div")     // Returns: false
//   isStructuralTag("header")  // Returns: false (this is a regular component)
func isStructuralTag(tagName string) bool {
	structural := map[string]bool{
		"html":     true,
		"head":     true,
		"body":     true,
		"!doctype": true,
	}
	return structural[strings.ToLower(tagName)]
}

// normalizeComponentPath generates all possible lookup keys for a component path
//
// Pattern: Helper Function [Load: 8]
// Cognitive Load: 8 (path parsing: 5, key generation: 3)
//
// This function addresses the component lookup problem where components might be
// registered with different path variants (e.g., "./components/Header.html" vs "Header")
//
// Example:
//   normalizeComponentPath("./components/UserProfile.html")
//   Returns: ["./components/UserProfile.html", "UserProfile.html", "UserProfile", "./components/UserProfile"]
//
//   normalizeComponentPath("Header")
//   Returns: ["Header"]
func normalizeComponentPath(path string) []string {
	keys := []string{path} // Always include the original path

	// If path contains a directory separator, generate additional keys
	if strings.Contains(path, "/") {
		// Extract filename (everything after last /)
		parts := strings.Split(path, "/")
		filename := parts[len(parts)-1]

		// Add filename as a key
		keys = append(keys, filename)

		// If filename has extension, add version without extension
		if strings.Contains(filename, ".") {
			nameParts := strings.Split(filename, ".")
			nameWithoutExt := strings.Join(nameParts[:len(nameParts)-1], ".")
			keys = append(keys, nameWithoutExt)
		}

		// Add path without extension
		if strings.Contains(path, ".") {
			pathParts := strings.Split(path, ".")
			pathWithoutExt := strings.Join(pathParts[:len(pathParts)-1], ".")
			keys = append(keys, pathWithoutExt)
		}
	} else if strings.Contains(path, ".") {
		// Path is just a filename with extension (e.g., "Header.html")
		nameParts := strings.Split(path, ".")
		nameWithoutExt := strings.Join(nameParts[:len(nameParts)-1], ".")
		keys = append(keys, nameWithoutExt)
	}

	return keys
}

// resolvePropValue resolves a component prop value against the parent data scope
//
// Pattern: Helper Function [Load: 8]
// Cognitive Load: 8 (type checking: 3, scope lookup: 3, value extraction: 2)
//
// This is the public version of extractPropValue, exported for use in tests.
// It handles three types of props:
// 1. Dynamic props (prop={expression}) - resolve against parent scope
// 2. Shorthand props ({prop}) - resolve from parent scope
// 3. Static props (prop="value") - return as string literal
//
// Example:
//   parentScope := map[string]any{"user": "Alice", "count": 42}
//
//   resolvePropValue(ComponentProp{Name: "name", Value: "{user}", IsDynamic: true}, parentScope)
//   // Returns: "Alice"
//
//   resolvePropValue(ComponentProp{Name: "total", Value: "{count}", IsDynamic: true}, parentScope)
//   // Returns: 42
//
//   resolvePropValue(ComponentProp{Name: "label", Value: "Submit", IsDynamic: false}, parentScope)
//   // Returns: "Submit"
func resolvePropValue(prop ast.ComponentProp, parentDataScope map[string]any) any {
	return extractPropValue(prop, parentDataScope)
}

// isSimpleVariableReference checks if a string is a simple variable reference
//
// Pattern: Helper Function [Load: 6]
// Cognitive Load: 6 (validation checks: 4, regex: 2)
//
// Returns true for simple identifiers like: user1, myVar, data
// Returns false for: "string", 123, user.name, true, null, function calls, etc.
//
// This is used to determine if a prop value should be output as an Alpine.js
// expression (user: user1) vs a literal value (user: "string").
func isSimpleVariableReference(s string) bool {
	s = strings.TrimSpace(s)

	// Empty or special literals
	if s == "" || s == "null" || s == "true" || s == "false" {
		return false
	}

	// Quoted strings
	if (strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"")) ||
		(strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'")) {
		return false
	}

	// Numbers
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return false
	}

	// Property access or function calls
	if strings.Contains(s, ".") || strings.Contains(s, "(") {
		return false
	}

	// Must be a valid JavaScript identifier
	matched, _ := regexp.MatchString(`^[a-zA-Z_$][a-zA-Z0-9_$]*$`, s)
	return matched
}

// addComponentDataWrapper wraps component nodes with an x-data attribute
//
// Pattern: Helper Function [Load: 10]
// Cognitive Load: 10 (element type checking: 4, attribute manipulation: 6)
//
// This is an alias for wrapWithXData, provided for backward compatibility with tests.
// It ensures the component output has an x-data attribute containing the data scope.
//
// Requirements:
// 1. Format data scope as Alpine.js object
// 2. Add x-data to single root element if it exists
// 3. Wrap multiple nodes in div with x-data
//
// Example:
//   Single element: <div class="card">content</div>
//     → <div x-data='{ count: 0 }' class="card">content</div>
//
//   Multiple nodes: [<h1>Title</h1>, <p>Content</p>]
//     → <div x-data='{ count: 0 }'><h1>Title</h1><p>Content</p></div>
func addComponentDataWrapper(nodes []ast.Node, dataScope map[string]any) []ast.Node {
	return wrapWithXData(nodes, dataScope)
}

// formatComponentData formats the component data scope for the x-data attribute
// This is a special formatter for components to match the expected output format
func formatComponentData(dataScope map[string]any) string {
	// For test cases, we need to format the data in a specific way
	// Create a simple string representation
	var result strings.Builder
	result.WriteString("{ ")

	// Add each key-value pair
	first := true
	for key, value := range dataScope {
		// DEBUG: Log the type of each value
		log.Printf("[formatComponentData] key=%q, type=%T, value preview=%v", key, value, truncateString(fmt.Sprintf("%v", value), 100))
		// Skip internal Alpine.js variables
		if strings.HasPrefix(key, "$") {
			continue
		}

		// Add comma if not the first item
		if !first {
			result.WriteString(", ")
		}
		first = false

		// Add value based on type
		switch v := value.(type) {
		case string:
			// SPECIAL CASE: If the string starts and ends with quotes, it may be:
			// 1. Fence section value like `prop x = "value"` - quotes should be stripped
			// 2. Component prop expression like `name={"Bo"}` - quotes are part of JS, keep them
			//
			// We detect #2 by checking if this looks like a complete JavaScript string literal
			// (has quotes AND the content is a valid string). For #1, we strip the quotes.
			cleanValue := v

			// Check if this is a JavaScript string literal expression (from component props)
			// These will be like: "Bo", 'Bo', etc. - they should be kept with quotes and
			// output as-is in the x-data (single quotes for HTML attribute safety)
			isJSStringLiteral := false
			if len(v) >= 2 {
				if (strings.HasPrefix(v, "\"") && strings.HasSuffix(v, "\"")) ||
					(strings.HasPrefix(v, "'") && strings.HasSuffix(v, "'")) {
					// This looks like a quoted string
					// If it came from extractPropValue as a string literal expression,
					// we should NOT strip the quotes
					// For now, we'll assume if it's ONLY a quoted string (no other content),
					// it's a JS string literal from a component prop
					innerContent := v[1 : len(v)-1]
					// If the inner content doesn't contain special chars that would indicate
					// it's a fence value, treat it as a JS string literal
					if !strings.Contains(innerContent, "\n") {
						isJSStringLiteral = true
					}
				}
			}

			if !isJSStringLiteral && len(v) >= 2 {
				if (strings.HasPrefix(v, "\"") && strings.HasSuffix(v, "\"")) ||
					(strings.HasPrefix(v, "'") && strings.HasSuffix(v, "'")) ||
					(strings.HasPrefix(v, "`") && strings.HasSuffix(v, "`")) {
					// Strip the outer quotes (fence section value)
					cleanValue = v[1 : len(v)-1]
				}
			}

			// CRITICAL FIX: Check if this is already a function/getter/setter definition
			// If so, output it in method shorthand format (valid for object literals)
			trimmedValue := strings.TrimSpace(cleanValue)
			if isFunctionDefinition(trimmedValue) {
				// Convert function declarations to method shorthand
				// "function name() {}" -> "name() {}"
				// "get name() {}" and "set name() {}" stay as-is
				functionDef := strings.ReplaceAll(cleanValue, `"`, `'`)

				// Remove "function " prefix if present (not for get/set)
				if strings.HasPrefix(trimmedValue, "function ") {
					functionDef = strings.Replace(functionDef, "function ", "", 1)
				}

				result.WriteString(functionDef)
				continue
			}

			// CRITICAL: Check if this value has the variable reference marker
			// Values marked with __VAR_REF__ prefix came from dynamic props that reference parent scope
			if strings.HasPrefix(cleanValue, "__VAR_REF__") {
				// Strip the marker and output as Alpine expression without quotes
				varName := strings.TrimPrefix(cleanValue, "__VAR_REF__")
				result.WriteString(key)
				result.WriteString(": ")
				result.WriteString(varName)
				continue
			}

			// CRITICAL: Check if cleanValue is a quoted string literal (from component props)
			// Like: "Bo" or 'Bo' - these should be output with single quotes
			if (strings.HasPrefix(cleanValue, "\"") && strings.HasSuffix(cleanValue, "\"")) ||
				(strings.HasPrefix(cleanValue, "'") && strings.HasSuffix(cleanValue, "'")) {
				// This is a JavaScript string literal - output with single quotes
				result.WriteString(key)
				result.WriteString(": ")
				// Convert to single quotes for HTML safety
				innerValue := cleanValue[1 : len(cleanValue)-1]
				result.WriteString("'")
				result.WriteString(innerValue)
				result.WriteString("'")
				continue
			}

			// CRITICAL FIX: Check for JavaScript literal BEFORE checking isDynamicExpression
			// This matches the logic in alpineDataFormatter (alpine.go lines 838-841)
			// JavaScript literals (arrays/objects) should be output as-is without quotes
			trimmedValue = strings.TrimSpace(cleanValue)
			if IsJavaScriptLiteral(trimmedValue) {
				// JavaScript literal (array or object) - don't quote it
				result.WriteString(key)
				result.WriteString(": ")
				// Convert double quotes to single quotes for HTML attribute safety
				cleanValue = strings.ReplaceAll(cleanValue, `"`, `'`)
				result.WriteString(cleanValue)
				log.Printf("[formatComponentData] Detected JavaScript literal for key=%q, outputting as-is: %s", key, truncateString(cleanValue, 100))
				continue
			}

			// Check if this is a dynamic expression (no quotes)
			// We need to handle variable references without quotes
			isDynamic := isDynamicExpression(cleanValue, dataScope)
			if isDynamic {
				// CRITICAL FIX: Check if this expression references other variables in the same data scope
				// If it does, we need to use a getter function so it can access 'this'
				if referencesOtherScopeVars(cleanValue, key, dataScope) {
					// Use getter syntax: get navItems() { return this.isLoggedIn ? [...] : [...] }
					result.WriteString("get ")
					result.WriteString(key)
					result.WriteString("() { return ")
					// Replace variable references with this.varName
					getterValue := replaceVarRefsWithThis(cleanValue, dataScope)
					result.WriteString(getterValue)
					result.WriteString(" }")
				} else {
					// This is a variable reference or expression, don't quote it
					result.WriteString(key)
					result.WriteString(": ")
					// CRITICAL FIX: For object literals and arrays, convert double quotes to single quotes
					// to prevent breaking HTML attributes like x-data="{ user: { name: "value" } }"
					trimmedValue := strings.TrimSpace(cleanValue)
					if strings.HasPrefix(trimmedValue, "{") || strings.HasPrefix(trimmedValue, "[") {
						// Replace double quotes with single quotes in object literals and arrays
						cleanValue = strings.ReplaceAll(cleanValue, `"`, `'`)
					}
					result.WriteString(cleanValue)
				}
			} else {
				// This is a literal string, add quotes
				result.WriteString(key)
				result.WriteString(": ")
				result.WriteString("'")
				result.WriteString(cleanValue)
				result.WriteString("'")
			}
		default:
			// For all other types (bool, int, float, etc.), use FormatGoValueToJS for consistency
			result.WriteString(key)
			result.WriteString(": ")
			result.WriteString(FormatGoValueToJS(v))
		}
	}

	result.WriteString(" }")
	return result.String()
}

// isFunctionDefinition checks if a string is a function, getter, or setter definition
// Returns true for patterns like:
//   - function name() { ... }
//   - get name() { ... }
//   - set name(value) { ... }
//   - async function name() { ... }
//   - name() { ... } (method shorthand)
func isFunctionDefinition(value string) bool {
	trimmed := strings.TrimSpace(value)

	// Check for getter: get name() { ... }
	if strings.HasPrefix(trimmed, "get ") {
		return true
	}

	// Check for setter: set name(value) { ... }
	if strings.HasPrefix(trimmed, "set ") {
		return true
	}

	// Check for function keyword: function name() { ... } or async function name() { ... }
	if strings.HasPrefix(trimmed, "function ") || strings.HasPrefix(trimmed, "async function ") {
		return true
	}

	// Check for method shorthand: name() { ... } or async name() { ... }
	// This pattern: word followed by parentheses and a brace
	methodPattern := regexp.MustCompile(`^(?:async\s+)?[a-zA-Z_$][a-zA-Z0-9_$]*\s*\([^)]*\)\s*\{`)
	return methodPattern.MatchString(trimmed)
}

// isValidVariableName checks if a string is a valid JavaScript variable name
func isValidVariableName(s string) bool {
	if len(s) == 0 {
		return false
	}

	// First character must be a letter, underscore, or dollar sign
	firstChar := s[0]
	if !((firstChar >= 'a' && firstChar <= 'z') ||
		(firstChar >= 'A' && firstChar <= 'Z') ||
		firstChar == '_' ||
		firstChar == '$') {
		return false
	}

	// Rest of the characters must be letters, numbers, underscores, or dollar signs
	for i := 1; i < len(s); i++ {
		c := s[i]
		if !((c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') ||
			c == '_' ||
			c == '$') {
			return false
		}
	}

	return true
}

// referencesOtherScopeVars checks if an expression references other variables in the data scope
// (excluding the current variable itself). This is used to determine if we need a getter function.
func referencesOtherScopeVars(expr string, currentVar string, dataScope map[string]any) bool {
	// Check each variable in the data scope (except the current one)
	for varName := range dataScope {
		if varName == currentVar || strings.HasPrefix(varName, "$") {
			continue
		}

		// Check if this variable name appears in the expression as a reference (not a property key)
		// Use word boundaries to avoid false positives (e.g., "user" in "username")
		varPattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(varName) + `\b`)
		matches := varPattern.FindAllStringIndex(expr, -1)

		// Check each match to see if it's a variable reference (not a property key)
		for _, match := range matches {
			endIdx := match[1]
			afterMatch := expr[endIdx:]
			trimmedAfter := strings.TrimSpace(afterMatch)
			// If NOT followed by a colon, it's a variable reference
			if !strings.HasPrefix(trimmedAfter, ":") {
				log.Printf("referencesOtherScopeVars: Expression '%s' references scope var '%s'", expr, varName)
				return true
			}
		}
	}
	return false
}

// replaceVarRefsWithThis replaces variable references in an expression with this.varName
// so they can be used in a getter function that accesses the object's properties
func replaceVarRefsWithThis(expr string, dataScope map[string]any) string {
	result := expr

	// Replace each variable reference with this.varName
	for varName := range dataScope {
		if strings.HasPrefix(varName, "$") {
			continue
		}

		// Replace word-boundary matches: varName -> this.varName
		// But we need to be careful NOT to replace property keys in object literals
		// Since Go's regexp doesn't support lookahead, we'll use FindAllStringIndex and process manually
		varPattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(varName) + `\b`)
		matches := varPattern.FindAllStringIndex(result, -1)

		// Process matches in reverse order to preserve indices
		for i := len(matches) - 1; i >= 0; i-- {
			match := matches[i]
			startIdx := match[0]
			endIdx := match[1]

			// Check if it's followed by a colon (property key pattern)
			afterMatch := result[endIdx:]
			trimmedAfter := strings.TrimSpace(afterMatch)
			if !strings.HasPrefix(trimmedAfter, ":") {
				// Not a property key, replace it
				result = result[:startIdx] + "this." + result[startIdx:endIdx] + result[endIdx:]
			}
		}
	}

	// CRITICAL FIX: Also convert double quotes to single quotes in the result
	// to prevent breaking HTML attributes when this is used in getters
	result = strings.ReplaceAll(result, `"`, `'`)

	log.Printf("replaceVarRefsWithThis: '%s' -> '%s'", expr, result)
	return result
}

// Helper function to determine if a string value is a dynamic expression (not a literal)
// THE ROBUST SOLUTION: Check operators first, then variable existence in scope
func isDynamicExpression(value string, dataScope map[string]any) bool {
	trimmed := strings.TrimSpace(value)

	// STEP 1: Check for obvious operators/syntax that indicate expressions
	// These are ALWAYS dynamic regardless of scope

	// Object literals: { name: "value", key: value }
	if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
		return true
	}

	// CRITICAL: Check for URL paths BEFORE arithmetic operators
	// Paths like "/contact" or "/about" should NOT be treated as division!
	// URL paths start with / and don't contain spaces or other operators
	if strings.HasPrefix(trimmed, "/") && !strings.Contains(trimmed, " ") {
		// This looks like a URL path, not a division expression
		return false
	}

	// Arithmetic operators: age + 50, count * 2
	// Note: Division (/) is checked above to avoid false positives with paths
	if strings.Contains(trimmed, "+") ||
		strings.Contains(trimmed, "-") ||
		strings.Contains(trimmed, "*") ||
		strings.Contains(trimmed, "/") ||
		strings.Contains(trimmed, "%") {
		return true
	}

	// Array access: items[0]
	if strings.Contains(trimmed, "[") && strings.Contains(trimmed, "]") {
		return true
	}

	// Function calls: formatDate()
	if strings.Contains(trimmed, "(") && strings.Contains(trimmed, ")") {
		return true
	}

	// Property access: user.name (must start with valid identifier)
	propertyAccessPattern := regexp.MustCompile(`^[a-zA-Z_$][a-zA-Z0-9_$]*\.`)
	if propertyAccessPattern.MatchString(trimmed) {
		return true
	}

	// STEP 2: Check if it's a valid variable name AND exists in scope
	if isValidVariableName(trimmed) {
		if _, exists := dataScope[trimmed]; exists {
			return true  // It's a variable reference
		}
	}

	// STEP 3: Otherwise it's a literal string
	return false
}

// extractPropValue extracts the value from a component prop, handling dynamic vs static values
//
// CRITICAL FIX (2025-10-06): Return variable references as strings, not resolved values
//
// For component props that reference parent variables (e.g., user={user1}), we need to
// pass the variable NAME as a string so Alpine.js can resolve it at runtime, NOT the
// actual value from the parent scope.
//
// Example:
//   Parent has: user1 = {name: "Alice"}
//   Template: <UserProfile user={user1} />
//
//   BEFORE (wrong): componentDataScope["user"] = {name: "Alice"}
//                   Output: user: { name: 'Alice' } (static object)
//
//   AFTER (correct): componentDataScope["user"] = "user1"
//                    Output: user: user1 (Alpine expression)
func extractPropValue(prop ast.ComponentProp, parentDataScope map[string]any) any {
	if prop.IsDynamic {
		// For dynamic props ({var}), extract the variable name or expression
		// Remove curly braces and whitespace
		varName := strings.TrimSpace(strings.Trim(prop.Value, "{}"))


		// CRITICAL: Check if this is a quoted string literal like {"Bo"}
		// These should be output with quotes in the x-data
		if (strings.HasPrefix(varName, "\"") && strings.HasSuffix(varName, "\"")) ||
			(strings.HasPrefix(varName, "'") && strings.HasSuffix(varName, "'")) {
			// This is a string literal expression - return as-is with quotes
			// The quotes are part of the JavaScript expression, not fence-section quotes
			log.Printf("extractPropValue: String literal expression '%s'", varName)
			return varName
		}

		// CRITICAL FIX: Check if this is a simple variable reference
		// If so, return the variable NAME (as string) for Alpine to resolve
		// Use a special prefix to mark it as a variable reference (not a string literal)
		if isSimpleVariableReference(varName) {
			// Check if the variable exists in parent scope
			if value, exists := parentDataScope[varName]; exists {
				// CRITICAL: Check if the parent's value is ALSO a __VAR_REF__
				// If yes, keep it as a variable reference (for reactive variables)
				// If no, return the actual value (for static data like JSON objects)
				if strVal, ok := value.(string); ok && strings.HasPrefix(strVal, "__VAR_REF__") {
					// Parent has a variable reference, keep the chain
					log.Printf("extractPropValue: Passing variable reference '%s' (parent also has __VAR_REF__)", varName)
					return "__VAR_REF__" + varName
				} else {
					// Parent has actual data, return it directly
					log.Printf("extractPropValue: Resolving '%s' to actual value (type: %T)", varName, value)
					return value
				}
			} else {
				log.Printf("extractPropValue: Variable '%s' NOT FOUND in parent scope!", varName)
			}
		}

		// For expressions (age + 10, user.name), return as-is for Alpine.js to evaluate
		log.Printf("extractPropValue: Passing expression '%s'", varName)
		return varName
	}

	// Static prop value - return as string literal
	return prop.Value
}

// transformComponentProps converts component props to a data scope map
// This handles both static and dynamic prop values
func transformComponentProps(props []ast.ComponentProp, parentDataScope map[string]any) map[string]any {
	propScope := make(map[string]any)

	for _, prop := range props {
		propScope[prop.Name] = extractPropValue(prop, parentDataScope)
	}

	return propScope
}


// TransformComponent is the main entry point for component transformation
// It's called by the main transformer for ComponentNode AST nodes
func TransformComponent(node *ast.ComponentNode, parentDataScope map[string]any) []ast.Node {
	return transformComponent(node, parentDataScope)
}

// transformComponent handles both regular and dynamic component transformation
//
// Pattern: Service Implementation Pattern [Load: 25]
// Cognitive Load: 25 (complex multi-phase transformation)
//
// This function implements the component transformation with prop resolution:
//
// PHASE 1: Component lookup and scope creation (Task 2.4) ✓
//   - Look up component template from registry
//   - Create isolated data scope for this instance
//
// PHASE 2: Process component fence and resolve props (Task 2.5) ✓
//   - Extract fence props and their defaults
//   - Resolve passed props against parent scope
//   - Merge into component data scope
//
// PHASE 3: Transform component body (Task 2.6) ✓
//   - Recursively transform component's AST nodes
//   - Use component data scope (isolated from parent)
//
// PHASE 4: Wrap with x-data (Task 2.4) ✓
//   - Add x-data attribute to root element or wrapper
//   - Format data scope as Alpine.js object
//
// Example transformation:
//   Input:  <UserCard name={user.name} age={user.age} />
//   Output: <div x-data='{name:"John",age:30}' class="card">...</div>
func transformComponent(node *ast.ComponentNode, parentDataScope map[string]any) []ast.Node {
	componentName := node.Name
	log.Printf("Recursively transforming component: %s", componentName)
	log.Printf("DEBUG: parentDataScope keys: %v", getMapKeys(parentDataScope))
	log.Printf("DEBUG: parentDataScope values: %+v", parentDataScope)

	// DYNAMIC COMPONENT SUPPORT (Task 2.7) ✓
	// Handle dynamic component references (e.g., <{componentVar} />)
	isDynamic := strings.HasPrefix(componentName, "{") && strings.HasSuffix(componentName, "}")
	if isDynamic {
		// Extract variable name from braces
		varName := strings.Trim(componentName, "{} ")

		// Add variable to parent scope so Alpine.js can resolve it at runtime
		extractVariablesFromExpr(varName, parentDataScope)

		log.Printf("Dynamic component reference: %s (variable: %s)", componentName, varName)

		// Create x-component directive element for dynamic components
		// Alpine.js will resolve the component name at runtime
		dynamicElement := &ast.Element{
			TagName: "div",
			Attributes: []ast.Attribute{
				{
					Name:    "x-component",
					Value:   varName,
					Dynamic: true,
				},
			},
		}

		// Add passed props as data attributes
		for _, prop := range node.Props {
			propName := prop.Name
			propValue := prop.Value

			if prop.IsDynamic {
				// Dynamic prop value - reference from parent scope
				propValue = strings.Trim(propValue, "{} ")
			}

			dynamicElement.Attributes = append(dynamicElement.Attributes, ast.Attribute{
				Name:    "data-prop-" + propName,
				Value:   propValue,
				Dynamic: prop.IsDynamic,
			})
		}

		return []ast.Node{dynamicElement}
	}

	// PHASE 1: Component lookup and scope creation (Task 2.4) ✓

	// Step 1: Look up component template from registry
	componentTemplate, exists := GetComponentTemplate(componentName)
	if !exists {
		log.Printf("Warning: Component template '%s' not registered. Creating placeholder element.", componentName)

		// Return placeholder element for unregistered components
		// This allows components to be resolved later (e.g., at build time in Plenti)
		// Format: <div x-component="ComponentName" data-prop-propName="propValue"></div>

		placeholderAttrs := []ast.Attribute{
			{
				Name:  "x-component",
				Value: componentName,
			},
		}

		// Add props as data-prop-* attributes
		for _, prop := range node.Props {
			propName := prop.Name
			propValue := prop.Value

			// For dynamic props ({var}), extract the variable name
			// Handle both cases: prop.IsDynamic flag or curly braces in value
			if prop.IsDynamic || (strings.HasPrefix(propValue, "{") && strings.HasSuffix(propValue, "}")) {
				propValue = strings.TrimSpace(strings.Trim(propValue, "{}"))
			}

			placeholderAttrs = append(placeholderAttrs, ast.Attribute{
				Name:  "data-prop-" + propName,
				Value: propValue,
			})
		}

		return []ast.Node{
			&ast.Element{
				TagName:    "div",
				Attributes: placeholderAttrs,
				Children:   []ast.Node{},
			},
		}
	}

	// Step 2: Create isolated data scope for this component instance
	componentDataScope := make(map[string]any)
	log.Printf("Created isolated scope for component '%s'", componentName)

	// PHASE 2: Process component fence and resolve props (Task 2.5) ✓

	// Step 1: Extract component's fence data (props, variables, AND functions including getters/setters)
	// Look for fence section in the component template
	for _, rootNode := range componentTemplate.Template.RootNodes {
		if fence, ok := rootNode.(*ast.FenceSection); ok {
			log.Printf("DEBUG: Found fence section with %d props and %d variables", len(fence.Props), len(fence.Variables))

			// CRITICAL FIX: Use collectComponentFenceData to extract ALL fence data including functions
			// This was the bug - we were manually extracting props and variables but missing functions!
			collectComponentFenceData(fence, componentDataScope)

			log.Printf("DEBUG: After collectComponentFenceData, scope has %d entries", len(componentDataScope))
		}
	}

	// Step 2: Resolve passed props and override defaults
	for _, prop := range node.Props {
		propName := prop.Name
		propValue := extractPropValue(prop, parentDataScope)

		componentDataScope[propName] = propValue
		log.Printf("DEBUG: Override prop '%s' with passed value: %v (type: %T)", propName, propValue, propValue)
	}

	log.Printf("DEBUG: Final component data scope for '%s': %+v", componentName, componentDataScope)

	// PHASE 3: Transform component body (Task 2.6) ✓

	// Filter out fence section and style section from component body
	// Style sections are extracted separately by GetAggregatedStyles()
	componentBodyNodes := []ast.Node{}
	for _, node := range componentTemplate.Template.RootNodes {
		_, isFence := node.(*ast.FenceSection)
		_, isStyle := node.(*ast.StyleSection)
		if !isFence && !isStyle {
			componentBodyNodes = append(componentBodyNodes, node)
		}
	}
	// Recursively transform component body with its isolated scope
	transformedNodes := transformNodes(componentBodyNodes, componentDataScope, false, false)

	// PHASE 4: Wrap with x-data (Task 2.4 + PHASE 2 OPTIMIZATION) ✓

	// Only add x-data wrapper if component has data
	// Components with no props, variables, or functions don't need Alpine.js wrapper
	if len(componentDataScope) == 0 {
		return transformedNodes
	}

	// PHASE 2 OPTIMIZATION: Smart scope diffing to minimize x-data duplication
	//
	// CRITICAL: Components ALWAYS need their own x-data wrapper for isolation
	// The optimization only applies to NON-component scopes (like conditional blocks, etc.)
	//
	// Reason: Components define an API contract via props - they expect specific
	// variables to be in scope. Even if parent happens to have same values, component
	// instances must be isolated for proper reactivity and encapsulation.
	//
	// Example that would break without component isolation:
	//   <Age name={name} age={age} />        <!-- expects name/age in its scope -->
	//   <Age name={"Bo"} age={age + 50} />   <!-- different values, different instance -->
	//
	// Without wrappers, both would share parent scope and show same values!
	if OptimizeXData {
		// Always wrap components with full scope (not diff)
		log.Printf("[X-Data] Component '%s' needs wrapper (component isolation required)", componentName)
		return wrapWithXData(transformedNodes, componentDataScope)
	}

	// Legacy behavior - always wrap with full scope
	log.Printf("[X-Data] Legacy mode: wrapping '%s' with full scope", componentName)
	return wrapWithXData(transformedNodes, componentDataScope)
}

// wrapWithXData ensures the component output has an x-data attribute
//
// Pattern: Helper Function [Load: 12]
// Cognitive Load: 12 (structural tag check: 2, element type checking: 4, attribute manipulation: 6)
//
// CRITICAL CHANGE (Phase A - 2025-10-11):
// Skip adding x-data to structural HTML tags (html, head, body, !doctype).
// These tags should never be reactive as they contain metadata or are managed by the server.
//
// Requirements from Task 2.4:
// 1. Format data scope as Alpine.js object
// 2. Add x-data to single root element if it exists AND it's not structural
// 3. Wrap multiple nodes in div with x-data
// 4. Skip wrapping for structural tags (html, head, body)
//
// Example:
//   Single element: <div class="card">content</div>
//     → <div x-data='{...}' class="card">content</div>
//
//   Structural tag: <head>metadata</head>
//     → <head>metadata</head> (NO x-data added)
//
//   Multiple nodes: [<h1>Title</h1>, <p>Content</p>]
//     → <div x-data='{...}'><h1>Title</h1><p>Content</p></div>
func wrapWithXData(nodes []ast.Node, dataScope map[string]any) []ast.Node {
	// REQUIREMENT 1: Format data scope (COGNITIVE LOAD: 2)
	xDataValue := formatComponentData(dataScope)
	xDataAttr := ast.Attribute{
		Name:       "x-data",
		Value:      xDataValue,
		Dynamic:    true,
		IsAlpine:   true,
		AlpineType: "data",
	}

	// REQUIREMENT 2: Check for single root element (COGNITIVE LOAD: 6)
	// If there's exactly one node and it's an Element, add x-data to it
	if len(nodes) == 0 {
		// No nodes - return empty div with x-data
		return []ast.Node{
			&ast.Element{
				TagName:    "div",
				Attributes: []ast.Attribute{xDataAttr},
				Children:   []ast.Node{},
			},
		}
	}

	// Check for single root element (REQUIREMENT 2)
	if len(nodes) == 1 {
		if element, ok := nodes[0].(*ast.Element); ok {
			// CRITICAL: Check if this is a structural tag (COGNITIVE LOAD: 2)
			if isStructuralTag(element.TagName) {
				log.Printf("wrapWithXData: Skipping x-data for structural tag <%s>", element.TagName)
				// Return as-is without x-data - structural tags should not be reactive
				return nodes
			}

			// Check if x-data already exists to avoid duplicates
			hasXData := false
			for _, attr := range element.Attributes {
				if attr.Name == "x-data" {
					hasXData = true
					break
				}
			}

			// Add x-data to existing element if not present
			if !hasXData {
				element.Attributes = append(element.Attributes, xDataAttr)
			}
			return nodes
		}
	}

	// Wrap multiple nodes or non-element single node in div (REQUIREMENT 3)
	wrapper := &ast.Element{
		TagName:     "div",
		Attributes:  []ast.Attribute{xDataAttr},
		Children:    nodes,
		SelfClosing: false,
	}

	return []ast.Node{wrapper}
}

// TransformDynamicComponent is the public entry point for dynamic component transformation
func TransformDynamicComponent(node *ast.DynamicComponentNode, parentDataScope map[string]any) []ast.Node {
	return transformDynamicComponent(node, parentDataScope)
}

// transformDynamicComponent handles dynamic component transformation (<= syntax)
//
// Pattern: Service Implementation Pattern [Load: 25]
// Cognitive Load: 25 (path resolution: 8, variable extraction: 5, component lookup: 7, transformation: 5)
//
// This function implements Jim's innovative dynamic component feature (<= syntax):
//
// PHASE 1: Extract variables from path expression (COGNITIVE LOAD: 5)
//   Example: "./views/{comp}.html" → extract "comp" variable
//
// PHASE 2: Try to resolve path at transformation time (COGNITIVE LOAD: 8)
//   If path is static or variables have known values, resolve now for build-time optimization
//
// PHASE 3: Look up component template (COGNITIVE LOAD: 7)
//   If found, proceed with transformation. If not, create placeholder.
//
// PHASE 4: Transform like regular component (COGNITIVE LOAD: 5)
//   Reuse transformComponent logic with resolved path
//
// Example transformation:
//   Input:  <='./components/UserProfile.html' name={user.name} age={30} />
//   Output: <div x-data='{name:"John",age:30}' class="profile">...</div>
func transformDynamicComponent(node *ast.DynamicComponentNode, parentDataScope map[string]any) []ast.Node {
	log.Printf("transformDynamicComponent: path=%s, props=%d", node.PathExpression, len(node.Props))

	// PHASE 1: Extract variables from path expression (COGNITIVE LOAD: 5)
	// Example: "./views/{comp}.html" → extract "comp" variable
	extractVariablesFromPath(node.PathExpression, parentDataScope)

	// PHASE 2: Try to resolve path at transformation time (COGNITIVE LOAD: 8)
	// If path is static or variables have known values, resolve now for build-time optimization
	resolvedPath := resolveDynamicPath(node.PathExpression, parentDataScope)
	log.Printf("transformDynamicComponent: resolved path: '%s'", resolvedPath)

	// PHASE 3: Look up component template (COGNITIVE LOAD: 7)
	_, exists := GetComponentTemplate(resolvedPath)
	if !exists {
		// DEBUG: Log all registered keys when lookup fails
		allKeys := GetAllRegisteredKeys()
		log.Printf("WARNING: Dynamic component lookup failed!")
		log.Printf("  Resolved path: '%s'", resolvedPath)
		log.Printf("  All registered keys (%d):", len(allKeys))
		for i, key := range allKeys {
			log.Printf("    [%d] '%s'", i, key)
		}

		// Path couldn't be resolved at build time - check if it still has variables
		if strings.Contains(resolvedPath, "{") {
			log.Printf("transformDynamicComponent: Path contains unresolved variables: %s", resolvedPath)
			// Return placeholder with x-component-dynamic attribute for runtime resolution
			return createDynamicComponentPlaceholder(node, resolvedPath, parentDataScope)
		}

		log.Printf("Warning: Dynamic component path not found: %s", resolvedPath)
		// Return placeholder even for static unresolved paths
		return createDynamicComponentPlaceholder(node, resolvedPath, parentDataScope)
	}

	// PHASE 4: Transform like regular component (COGNITIVE LOAD: 5)
	// We have a resolved component template - transform it using existing logic
	log.Printf("transformDynamicComponent: Found component template for: %s", resolvedPath)

	// Create a ComponentNode to reuse existing transformation logic (DRY principle)
	regularComponentNode := &ast.ComponentNode{
		Name:  resolvedPath,
		Props: node.Props,
	}

	return transformComponent(regularComponentNode, parentDataScope)
}

// extractVariablesFromPath extracts {variable} references from a path expression and adds them to data scope
//
// Pattern: Helper Function [Load: 6]
// Cognitive Load: 6 (regex matching: 3, loop: 2, scope update: 1)
//
// Example:
//   extractVariablesFromPath("./views/{comp}.html", dataScope)
//   // Adds "comp" to dataScope if not present
//
//   extractVariablesFromPath("./views/{section}/{page}.html", dataScope)
//   // Adds "section" and "page" to dataScope
func extractVariablesFromPath(pathExpr string, dataScope map[string]any) {
	// Use regex to find all {variable} patterns (COGNITIVE LOAD: 3)
	varPattern := regexp.MustCompile(`\{([a-zA-Z_$][a-zA-Z0-9_$]*)\}`)
	matches := varPattern.FindAllStringSubmatch(pathExpr, -1)

	// Add each variable to data scope (COGNITIVE LOAD: 3)
	for _, match := range matches {
		if len(match) > 1 {
			varName := match[1]
			if _, exists := dataScope[varName]; !exists {
				dataScope[varName] = nil // Add to scope with nil value (will be provided at runtime)
				log.Printf("extractVariablesFromPath: Added variable '%s' to data scope", varName)
			}
		}
	}
}

// resolveDynamicPath attempts to resolve a dynamic path by substituting known variables
//
// Pattern: Helper Function [Load: 8]
// Cognitive Load: 8 (regex matching: 3, variable lookup: 3, string replacement: 2)
//
// This function enables build-time path resolution when variables have known values.
// If any variable is unknown (nil), the path is returned partially resolved.
//
// Example:
//   dataScope := map[string]any{"comp": "Header"}
//   resolveDynamicPath("./views/{comp}.html", dataScope)
//   // Returns: "./views/Header.html"
//
//   dataScope := map[string]any{"comp": nil}
//   resolveDynamicPath("./views/{comp}.html", dataScope)
//   // Returns: "./views/{comp}.html" (unchanged - variable not resolved)
func resolveDynamicPath(pathExpr string, dataScope map[string]any) string {
	// QUICK FIX: Strip surrounding backticks, single quotes, and double quotes
	resolved := strings.Trim(pathExpr, "`'\"")
	log.Printf("resolveDynamicPath: Cleaned path from '%s' to '%s'", pathExpr, resolved)

	// Find all {variable} patterns (COGNITIVE LOAD: 3)
	varPattern := regexp.MustCompile(`\{([a-zA-Z_$][a-zA-Z0-9_$]*)\}`)
	matches := varPattern.FindAllStringSubmatch(resolved, -1)

	// Substitute variables with their values (COGNITIVE LOAD: 5)
	for _, match := range matches {
		if len(match) > 1 {
			varName := match[1]

			// Check if variable exists in data scope and has a value
			if val, exists := dataScope[varName]; exists && val != nil {
				// Convert value to string
				var strVal string
				switch v := val.(type) {
				case string:
					strVal = v
				default:
					strVal = fmt.Sprintf("%v", v)
				}

				// Replace {varName} with actual value
				resolved = strings.Replace(resolved, match[0], strVal, 1)
				log.Printf("resolveDynamicPath: Resolved {%s} to '%s'", varName, strVal)
			} else {
				log.Printf("resolveDynamicPath: Variable '%s' not resolved (value: %v)", varName, val)
			}
		}
	}

	return resolved
}

// createDynamicComponentPlaceholder creates a placeholder element for unresolved dynamic components
//
// Pattern: Helper Function [Load: 5]
// Cognitive Load: 5 (element creation: 2, prop transformation: 3)
//
// This function creates a special div element with x-component-dynamic attribute
// that can be resolved at runtime (e.g., by Plenti's build system or Alpine.js plugin).
//
// Format: <div x-component-dynamic="path" data-prop-*="value"></div>
//
// Example:
//   node := &DynamicComponentNode{
//     PathExpression: "./views/{comp}.html",
//     Props: []ComponentProp{{Name: "title", Value: "Hello"}},
//   }
//   createDynamicComponentPlaceholder(node, "./views/{comp}.html", dataScope)
//   // Returns: <div x-component-dynamic="./views/{comp}.html" data-prop-title="Hello"></div>
func createDynamicComponentPlaceholder(node *ast.DynamicComponentNode, path string, dataScope map[string]any) []ast.Node {
	// Create base attributes (COGNITIVE LOAD: 2)
	attrs := []ast.Attribute{
		{
			Name:  "x-component",
			Value: path,
		},
	}

	// Add props as data-prop-* attributes (COGNITIVE LOAD: 3)
	for _, prop := range node.Props {
		propValue := resolvePropValueForPlaceholder(prop, dataScope)

		attrs = append(attrs, ast.Attribute{
			Name:  "data-prop-" + prop.Name,
			Value: propValue,
		})
	}

	// Return placeholder element
	return []ast.Node{
		&ast.Element{
			TagName:    "div",
			Attributes: attrs,
			Children:   []ast.Node{},
		},
	}
}

// resolvePropValueForPlaceholder resolves a prop value for use in a placeholder element
//
// Pattern: Helper Function [Load: 5]
// Cognitive Load: 5 (type checking: 2, value extraction: 3)
//
// For dynamic props ({var}), extracts the variable name.
// For static props, returns the value as-is.
func resolvePropValueForPlaceholder(prop ast.ComponentProp, _ map[string]any) string {
	if prop.IsDynamic {
		// For dynamic props ({var}), extract the variable name
		return strings.TrimSpace(strings.Trim(prop.Value, "{}"))
	}

	// For static props, return the value as-is
	return prop.Value
}
