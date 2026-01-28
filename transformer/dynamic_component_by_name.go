package transformer

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/jimafisk/custom_go_template/analyzer"
	"github.com/jimafisk/custom_go_template/ast"
)

// TransformDynamicComponentByName transforms <Component:dynamic> nodes into rendered components
//
// Pattern: Service Implementation Pattern [Load: 25]
// Cognitive Load: 25 (name evaluation: 6, component lookup: 5, prop merging: 8, transformation: 6)
//
// This function implements the dynamic component rendering feature for Plenti-style iteration:
//
// PHASE 1: Scope Analysis (COGNITIVE LOAD: 4) - NEW in Phase 2
//
//	Initialize ScopeAnalyzer to distinguish build-time vs runtime expressions
//
// PHASE 2: Evaluate name expression (COGNITIVE LOAD: 6)
//
//	Example: "component.name" → resolve from dataScope → "Hero2436"
//
// PHASE 3: Build-time vs Runtime Decision (COGNITIVE LOAD: 5) - NEW in Phase 2
//
//	if analyzer.IsRuntimeExpression(node.NameExpression):
//	  → emit runtime wrapper (Task 2.3)
//	else:
//	  → proceed with build-time resolution
//
// PHASE 4: Look up component template (COGNITIVE LOAD: 5)
//
//	componentTemplate := GetComponentTemplate(resolvedName)
//	If not found: return placeholder with warning
//
// PHASE 5: Build component props (COGNITIVE LOAD: 8)
//
//	a. Start with empty props map
//	b. Process spread props (left to right)
//	c. Process regular props (left to right)
//	d. Later props override earlier
//
// PHASE 6: Transform component (COGNITIVE LOAD: 6)
//
//	return transformComponent(componentTemplate, mergedProps, dataScope)
//
// Example transformation:
//
//	Input:  <Component:dynamic name={component.name} {...component.fields} theme="dark" />
//	Output: <div x-data='{...}' class="hero">...</div> (rendered component)
func TransformDynamicComponentByName(node *ast.DynamicComponentByNameNode, dataScope map[string]any) []ast.Node {
	log.Printf("TransformDynamicComponentByName: nameExpr=%q, spreadProps=%d, regularProps=%d",
		node.NameExpression, len(node.SpreadProps), len(node.Props))

	// PHASE 1: Initialize Scope Analyzer (COGNITIVE LOAD: 4) - NEW in Phase 2
	scopeAnalyzer := analyzer.NewScopeAnalyzer(dataScope)

	// PHASE 2: Check if this is a runtime-only expression (COGNITIVE LOAD: 5) - NEW in Phase 2
	if scopeAnalyzer.IsRuntimeExpression(node.NameExpression) {
		log.Printf("TransformDynamicComponentByName: detected RUNTIME expression: %q", node.NameExpression)
		// Emit runtime wrapper for Alpine.js to resolve at runtime
		return emitRuntimeWrapper(node, dataScope)
	}

	// PHASE 3: Evaluate name expression (COGNITIVE LOAD: 6)
	componentName, err := evaluateNameExpression(node.NameExpression, dataScope)
	if err != nil {
		log.Printf("TransformDynamicComponentByName: failed to evaluate name expression %q: %v",
			node.NameExpression, err)
		return createDynamicByNamePlaceholder(node, "ERROR: "+err.Error())
	}

	log.Printf("TransformDynamicComponentByName: resolved component name: %q", componentName)

	// PHASE 4: Look up component template (COGNITIVE LOAD: 5)
	_, exists := GetComponentTemplate(componentName)
	if !exists {
		log.Printf("TransformDynamicComponentByName: component %q not found in registry", componentName)
		return createDynamicByNamePlaceholder(node, fmt.Sprintf("Component '%s' not found", componentName))
	}

	// PHASE 5: Build component props (COGNITIVE LOAD: 8)
	// Step 1: Resolve spread props (left to right)
	spreadPropsMap := resolveSpreadProps(node.SpreadProps, dataScope)
	log.Printf("TransformDynamicComponentByName: resolved %d spread props", len(spreadPropsMap))

	// Step 2: Merge with regular props (regular props override spread props)
	mergedProps := mergeProps(spreadPropsMap, node.Props, dataScope)
	log.Printf("TransformDynamicComponentByName: merged props: %d total", len(mergedProps))

	// PHASE 6: Transform component with resolved props (COGNITIVE LOAD: 6)
	// Use TransformComponentWithResolvedProps to pass ACTUAL values (maps, arrays)
	// This bypasses JSON serialization that would break nested property access
	// for build-time loop expansion (e.g., content.components)
	log.Printf("TransformDynamicComponentByName: transforming component %q with %d resolved props",
		componentName, len(mergedProps))

	return TransformComponentWithResolvedProps(componentName, mergedProps, dataScope)
}

// emitRuntimeWrapper emits a runtime wrapper element for dynamic components
// that cannot be resolved at build-time.
//
// Pattern: Helper Function [Load: 12]
// Cognitive Load: 12 (element creation: 3, JSON serialization: 5, attribute construction: 4)
//
// This function creates an Alpine.js-compatible wrapper that will be resolved
// at runtime by the $renderDynamicComponent magic function.
//
// Runtime Wrapper Structure:
//
//	<div class="dyn-comp-runtime"
//	     x-data="{compName: component.name, compProps: {...component.fields}}"
//	     x-init="$renderDynamicComponent($el, compName, compProps)">
//	</div>
//
// IMPORTANT: This function does NOT auto-inject content/allContent.
// Components access these via Alpine.js scope inheritance from parent x-data.
// If explicit passing is needed, add content={content} to the template.
//
// Example:
//
//	Input:  nameExpression="component.name", spread=["component.fields"], props=[]
//	Output: <div x-data="{compName: component.name, compProps: {...component.fields}}" ...>
//
// IMPORTANT: The wrapper has NO children - Alpine.js will populate it at runtime
func emitRuntimeWrapper(node *ast.DynamicComponentByNameNode, dataScope map[string]any) []ast.Node {
	// Mark that we're using runtime component resolution
	// This tells the registry generator to generate layouts.js
	MarkRuntimeComponentUsed()

	log.Printf("emitRuntimeWrapper: creating runtime wrapper for nameExpr=%q", node.NameExpression)

	// PHASE 1: Build compProps with spread expressions and regular props (COGNITIVE LOAD: 8)
	// For runtime components, we need to keep expressions for Alpine to evaluate

	// Build the compProps object as a string with proper JavaScript syntax
	// Components access content/allContent via Alpine scope inheritance from body x-data
	// Only include explicitly passed props (spread and regular)
	var propsParts []string

	// Step 1: Add spread operators for spread props
	for _, spreadExpr := range node.SpreadProps {
		spreadExpr = strings.TrimSpace(spreadExpr)
		propsParts = append(propsParts, fmt.Sprintf("...%s", spreadExpr))
		log.Printf("emitRuntimeWrapper: adding spread expression: ...%s", spreadExpr)
	}

	// Step 2: Add regular props as key: value pairs
	for _, prop := range node.Props {
		var valueExpr string
		if prop.IsDynamic {
			// Dynamic prop: {var} → extract variable name
			varName := strings.TrimSpace(strings.Trim(prop.Value, "{}"))
			// Check if it's a quoted string literal
			if (strings.HasPrefix(varName, `"`) && strings.HasSuffix(varName, `"`)) ||
				(strings.HasPrefix(varName, `'`) && strings.HasSuffix(varName, `'`)) {
				valueExpr = varName
			} else {
				// Variable reference - keep as identifier
				valueExpr = varName
			}
		} else {
			// Static prop: serialize as JSON string
			jsonBytes, err := json.Marshal(prop.Value)
			if err != nil {
				log.Printf("emitRuntimeWrapper: failed to marshal prop %q: %v", prop.Name, err)
				valueExpr = "null"
			} else {
				// Convert double quotes to single quotes for HTML attribute
				valueExpr = strings.ReplaceAll(string(jsonBytes), `"`, `'`)
			}
		}

		propsParts = append(propsParts, fmt.Sprintf("%s: %s", prop.Name, valueExpr))
		log.Printf("emitRuntimeWrapper: adding regular prop: %s: %s", prop.Name, valueExpr)
	}

	// Build the final compProps object literal
	propsJSON := "{" + strings.Join(propsParts, ", ") + "}"
	log.Printf("emitRuntimeWrapper: compProps = %s", propsJSON)

	// PHASE 2: Build x-data attribute value (COGNITIVE LOAD: 4)
	// Format: {compName: expression, compProps: {...}}
	xDataValue := fmt.Sprintf("{compName: %s, compProps: %s}",
		node.NameExpression, // Keep expression as-is for Alpine to evaluate
		propsJSON,           // JavaScript object with spread operators
	)

	// PHASE 3: Create wrapper element (COGNITIVE LOAD: 3)
	wrapper := &ast.Element{
		TagName: "div",
		Attributes: []ast.Attribute{
			{
				Name:  "class",
				Value: "dyn-comp-runtime",
			},
			{
				Name:       "x-data",
				Value:      xDataValue,
				Dynamic:    true,
				IsAlpine:   true,
				AlpineType: "data",
			},
			{
				Name:       "x-init",
				Value:      "$renderDynamicComponent($el, compName, compProps)",
				Dynamic:    true,
				IsAlpine:   true,
				AlpineType: "init",
			},
		},
		Children:    []ast.Node{}, // Empty - runtime will populate
		SelfClosing: false,
	}

	log.Printf("emitRuntimeWrapper: created wrapper with x-data=%q", xDataValue)

	return []ast.Node{wrapper}
}

// AlpineExpression is a marker type for prop values that should be kept as Alpine expressions
//
// Pattern: Domain Model [Load: 2]
// Cognitive Load: 2 (simple wrapper type)
//
// This type wraps variable names that should be output as JavaScript identifiers
// in the x-data attribute, not as quoted strings.
//
// Example:
//
//	AlpineExpression{Expr: "content"} → serializes to: content (no quotes)
//	String "content"                  → serializes to: 'content' (quoted)
type AlpineExpression struct {
	Expr string // The JavaScript expression (e.g., "content", "component.name", "$store.user")
}

// isComplexObject checks if a value is a complex object (map or slice)
//
// Pattern: Helper Function [Load: 5]
// Cognitive Load: 5 (type checking with reflection: 5)
//
// This function uses reflection to check if a value is a map or slice,
// avoiding duplicate case issues in type switches (since map[string]any and
// map[string]interface{} are the same type, as are []any and []interface{}).
func isComplexObject(value any) bool {
	if value == nil {
		return false
	}

	kind := reflect.TypeOf(value).Kind()
	return kind == reflect.Map || kind == reflect.Slice
}

// serializePropsForRuntime serializes props map to JSON string for x-data attribute
//
// Pattern: Helper Function [Load: 14]
// Cognitive Load: 14 (manual serialization: 6, type switching: 4, string building: 2, dataScope lookup: 2)
//
// This function converts a map of props into a JavaScript object literal suitable
// for embedding in an HTML attribute (x-data). It handles:
//   - Regular values: Serialize as JSON (strings, numbers, bools, objects, arrays)
//   - AlpineExpression values: Check if they reference complex objects in dataScope
//   - If yes: JSON-serialize the actual value (fixes Go map syntax bug)
//   - If no: Output as unquoted JavaScript identifiers
//
// Example:
//
//	Input:  map[string]interface{}{
//	          "title": "Hello",
//	          "content": AlpineExpression{Expr: "content"}, // where content is a map in dataScope
//	          "count": 42,
//	        }
//	Output: `{title: 'Hello', content: {...}, count: 42}`
//
// CRITICAL FIX: AlpineExpression values that reference complex objects (maps/slices)
// are now JSON-serialized to prevent Go map syntax from appearing in the output.
func serializePropsForRuntime(props map[string]interface{}, dataScope map[string]any) string {
	if len(props) == 0 {
		return "{}"
	}

	// Build JavaScript object literal manually to handle AlpineExpression
	var parts []string

	for key, value := range props {
		var valueStr string

		// Check if this is an AlpineExpression (should be output as identifier)
		if alpineExpr, ok := value.(AlpineExpression); ok {
			// CRITICAL FIX: Check if this expression references a complex object in dataScope
			// If yes, we need to JSON-serialize the actual value instead of keeping it as an identifier
			if actualValue, exists := dataScope[alpineExpr.Expr]; exists {
				// Check if the actual value is a complex object (map or slice)
				if isComplexObject(actualValue) {
					// Complex object - JSON-serialize it
					jsonBytes, err := json.Marshal(actualValue)
					if err != nil {
						log.Printf("serializePropsForRuntime: failed to marshal complex object for prop %q: %v", key, err)
						valueStr = "null"
					} else {
						// Convert double quotes to single quotes for HTML attribute compatibility
						valueStr = strings.ReplaceAll(string(jsonBytes), `"`, `'`)
						log.Printf("serializePropsForRuntime: prop %q = complex object (JSON-serialized)", key)
					}
				} else {
					// Simple value or not in scope - output as unquoted identifier
					valueStr = alpineExpr.Expr
					log.Printf("serializePropsForRuntime: prop %q = AlpineExpression(%q) - simple value", key, alpineExpr.Expr)
				}
			} else {
				// Not in dataScope - output as unquoted identifier (will be resolved by Alpine at runtime)
				valueStr = alpineExpr.Expr
				log.Printf("serializePropsForRuntime: prop %q = AlpineExpression(%q) - not in scope", key, alpineExpr.Expr)
			}
		} else {
			// Regular value - serialize as JSON and convert quotes
			jsonBytes, err := json.Marshal(value)
			if err != nil {
				log.Printf("serializePropsForRuntime: failed to marshal prop %q: %v", key, err)
				valueStr = "null"
			} else {
				// Convert double quotes to single quotes for HTML attribute compatibility
				valueStr = strings.ReplaceAll(string(jsonBytes), `"`, `'`)
			}
			log.Printf("serializePropsForRuntime: prop %q = %v (type: %T)", key, valueStr, value)
		}

		// Build key: value pair
		parts = append(parts, fmt.Sprintf("%s: %s", key, valueStr))
	}

	// Join with commas
	result := "{" + strings.Join(parts, ", ") + "}"

	log.Printf("serializePropsForRuntime: serialized %d props to: %s", len(props), result)

	return result
}

// evaluateNameExpression evaluates the name expression to get component name
//
// Pattern: Helper Function [Load: 8]
// Cognitive Load: 8 (type checking: 3, nested resolution: 5)
//
// This function resolves the component name from various expression types:
// 1. String literals: "Hero2436" → "Hero2436"
// 2. Quoted strings: '"Hero2436"' → "Hero2436"
// 3. Simple variables: "componentName" → resolve from dataScope
// 4. Nested properties: "component.name" → resolve nested access
//
// Example:
//
//	dataScope := map[string]any{"component": map[string]any{"name": "Hero2436"}}
//	evaluateNameExpression("component.name", dataScope)
//	// Returns: "Hero2436", nil
func evaluateNameExpression(nameExpr string, dataScope map[string]any) (string, error) {
	nameExpr = strings.TrimSpace(nameExpr)

	// Handle quoted string literals (COGNITIVE LOAD: 2)
	if (strings.HasPrefix(nameExpr, `"`) && strings.HasSuffix(nameExpr, `"`)) ||
		(strings.HasPrefix(nameExpr, `'`) && strings.HasSuffix(nameExpr, `'`)) {
		// Strip quotes
		return nameExpr[1 : len(nameExpr)-1], nil
	}

	// Handle unquoted string literals (no dots, brackets, or operators) (COGNITIVE LOAD: 1)
	if !strings.Contains(nameExpr, ".") &&
		!strings.Contains(nameExpr, "[") &&
		!strings.Contains(nameExpr, "(") {
		// Check if it's a variable in dataScope
		if value, exists := dataScope[nameExpr]; exists {
			// Convert value to string
			if strVal, ok := value.(string); ok {
				return strVal, nil
			}
			return fmt.Sprintf("%v", value), nil
		}
		// Not in dataScope, treat as literal string
		return nameExpr, nil
	}

	// Handle nested property access (COGNITIVE LOAD: 5)
	// Example: "component.name" or "data.component.name"
	result := resolveNestedPropertyAccess(nameExpr, dataScope)
	if result == nil {
		return "", fmt.Errorf("failed to resolve name expression: %q", nameExpr)
	}

	// Convert result to string
	if strVal, ok := result.(string); ok {
		return strVal, nil
	}

	return fmt.Sprintf("%v", result), nil
}

// resolveNestedPropertyAccess resolves nested property access like "component.fields.title"
//
// Pattern: Helper Function [Load: 10]
// Cognitive Load: 10 (string splitting: 2, iterative resolution: 6, type checking: 2)
//
// Example:
//
//	dataScope := map[string]any{
//	  "component": map[string]any{
//	    "fields": map[string]any{"title": "Hello"},
//	  },
//	}
//	resolveNestedPropertyAccess("component.fields.title", dataScope)
//	// Returns: "Hello"
func resolveNestedPropertyAccess(expr string, dataScope map[string]any) any {
	parts := strings.Split(expr, ".")
	if len(parts) == 0 {
		return nil
	}

	// Start with the root variable
	current, exists := dataScope[parts[0]]
	if !exists {
		log.Printf("resolveNestedPropertyAccess: root variable %q not found in dataScope", parts[0])
		return nil
	}

	// Navigate through the property chain
	for i := 1; i < len(parts); i++ {
		part := parts[i]

		// Try to access as map
		if currentMap, ok := current.(map[string]any); ok {
			if value, exists := currentMap[part]; exists {
				current = value
				continue
			}
		}

		// Try to access as map[string]interface{}
		if currentMap, ok := current.(map[string]interface{}); ok {
			if value, exists := currentMap[part]; exists {
				current = value
				continue
			}
		}

		log.Printf("resolveNestedPropertyAccess: property %q not found at path %s",
			part, strings.Join(parts[:i+1], "."))
		return nil
	}

	return current
}

// resolveSpreadProps evaluates spread expressions and returns flattened props
//
// Pattern: Helper Function [Load: 12]
// Cognitive Load: 12 (loop: 3, nested resolution: 5, type checking: 2, merge: 2)
//
// This function expands spread operator syntax like {...component.fields} into individual props.
//
// Example:
//
//	spreadExprs := []string{"component.fields", "overrides"}
//	dataScope := map[string]any{
//	  "component": map[string]any{
//	    "fields": map[string]any{"title": "Hello", "link": "/about"},
//	  },
//	  "overrides": map[string]any{"title": "Override"},
//	}
//	resolveSpreadProps(spreadExprs, dataScope)
//	// Returns: map[string]any{"title": "Override", "link": "/about"}
func resolveSpreadProps(spreadExprs []string, dataScope map[string]any) map[string]any {
	result := make(map[string]any)

	// Process each spread expression in order (left to right)
	for _, expr := range spreadExprs {
		expr = strings.TrimSpace(expr)
		log.Printf("resolveSpreadProps: processing spread expression %q", expr)

		// Resolve the expression to get the object to spread
		var spreadValue any

		// Check if it's a simple variable
		if !strings.Contains(expr, ".") {
			if value, exists := dataScope[expr]; exists {
				spreadValue = value
			} else {
				log.Printf("resolveSpreadProps: variable %q not found in dataScope", expr)
				continue
			}
		} else {
			// Nested property access
			spreadValue = resolveNestedPropertyAccess(expr, dataScope)
			if spreadValue == nil {
				log.Printf("resolveSpreadProps: failed to resolve %q", expr)
				continue
			}
		}

		// Type assert to map and merge
		if spreadMap, ok := spreadValue.(map[string]any); ok {
			// Merge into result (later spreads override earlier)
			for key, value := range spreadMap {
				result[key] = value
			}
			log.Printf("resolveSpreadProps: spread %q added %d props", expr, len(spreadMap))
		} else if spreadMap, ok := spreadValue.(map[string]interface{}); ok {
			// Handle map[string]interface{} as well
			for key, value := range spreadMap {
				result[key] = value
			}
			log.Printf("resolveSpreadProps: spread %q added %d props", expr, len(spreadMap))
		} else {
			log.Printf("resolveSpreadProps: expression %q did not resolve to object (got %T)", expr, spreadValue)
		}
	}

	return result
}

// resolveSpreadPropsForRuntime is like resolveSpreadProps but returns spread EXPRESSIONS for runtime
//
// Pattern: Helper Function [Load: 10]
// Cognitive Load: 10 (expression building: 5, validation: 3, type conversion: 2)
//
// This function processes spread props for runtime components. Instead of resolving
// to actual values, it returns the spread expression itself wrapped as an AlpineExpression.
//
// Example:
//
//	spreadExprs := []string{"component.fields"}
//	// Returns: map with special marker indicating this should be output as "...component.fields"
//
// IMPORTANT: For runtime components, spread props need to be evaluated by Alpine.js,
// not resolved at build-time.
func resolveSpreadPropsForRuntime(spreadExprs []string, dataScope map[string]any) map[string]any {
	result := make(map[string]any)

	// For runtime components, we can't expand spreads at build-time
	// Instead, we need to resolve them to get the prop names, but keep expressions for values
	for _, expr := range spreadExprs {
		expr = strings.TrimSpace(expr)
		log.Printf("resolveSpreadPropsForRuntime: processing spread expression %q", expr)

		// Try to resolve the spread to get the keys
		var spreadValue any
		if !strings.Contains(expr, ".") {
			if value, exists := dataScope[expr]; exists {
				spreadValue = value
			}
		} else {
			spreadValue = resolveNestedPropertyAccess(expr, dataScope)
		}

		// If we got a map, extract the keys and create AlpineExpressions
		if spreadMap, ok := spreadValue.(map[string]any); ok {
			for key := range spreadMap {
				// Create Alpine expression: key -> expr.key
				alpineExpr := AlpineExpression{Expr: expr + "." + key}
				result[key] = alpineExpr
				log.Printf("resolveSpreadPropsForRuntime: spread %q added prop %q as expression %q",
					expr, key, alpineExpr.Expr)
			}
		} else if spreadMap, ok := spreadValue.(map[string]interface{}); ok {
			for key := range spreadMap {
				alpineExpr := AlpineExpression{Expr: expr + "." + key}
				result[key] = alpineExpr
				log.Printf("resolveSpreadPropsForRuntime: spread %q added prop %q as expression %q",
					expr, key, alpineExpr.Expr)
			}
		} else {
			log.Printf("resolveSpreadPropsForRuntime: expression %q did not resolve to object (got %T)", expr, spreadValue)
		}
	}

	return result
}

// mergeProps combines spread props and regular props with override logic
//
// Pattern: Helper Function [Load: 10]
// Cognitive Load: 10 (map operations: 3, loop: 2, prop resolution: 5)
//
// This function implements the prop merging order:
// 1. Spread props are applied first (left to right)
// 2. Regular props are applied next (left to right)
// 3. Later props override earlier (rightmost wins)
//
// Example:
//
//	spreadProps := map[string]any{"title": "Spread Title", "link": "/spread"}
//	regularProps := []ast.ComponentProp{
//	  {Name: "title", Value: "Override Title", IsDynamic: false},
//	  {Name: "theme", Value: "dark", IsDynamic: false},
//	}
//	mergeProps(spreadProps, regularProps, dataScope)
//	// Returns: {"title": "Override Title", "link": "/spread", "theme": "dark"}
func mergeProps(spreadProps map[string]any, regularProps []ast.ComponentProp, dataScope map[string]any) map[string]any {
	// Start with spread props
	result := make(map[string]any, len(spreadProps)+len(regularProps))

	// Copy spread props to result
	for key, value := range spreadProps {
		result[key] = value
	}

	log.Printf("mergeProps: starting with %d spread props", len(spreadProps))

	// Process regular props (these override spread props)
	for _, prop := range regularProps {
		propValue := extractPropValue(prop, dataScope)
		result[prop.Name] = propValue
		log.Printf("mergeProps: added/override prop %q = %v (type: %T)", prop.Name, propValue, propValue)
	}

	log.Printf("mergeProps: final result has %d props", len(result))
	return result
}

// mergePropsForRuntime combines props for runtime components, keeping expressions
//
// Pattern: Helper Function [Load: 10]
// Cognitive Load: 10 (map operations: 3, loop: 2, expression extraction: 5)
//
// This function is similar to mergeProps but designed for runtime dynamic components.
// Key difference: Dynamic prop values are kept as AlpineExpression instead of being
// resolved to their actual values.
//
// Example:
//
//	spreadProps := map[string]any{"title": AlpineExpression{Expr: "component.fields.title"}}
//	regularProps := []ast.ComponentProp{
//	  {Name: "content", Value: "{content}", IsDynamic: true},
//	  {Name: "theme", Value: "dark", IsDynamic: false},
//	}
//	mergePropsForRuntime(spreadProps, regularProps, dataScope)
//	// Returns: {
//	//   "title": AlpineExpression{Expr: "component.fields.title"},
//	//   "content": AlpineExpression{Expr: "content"},
//	//   "theme": "dark"
//	// }
func mergePropsForRuntime(spreadProps map[string]any, regularProps []ast.ComponentProp, dataScope map[string]any) map[string]any {
	// Start with spread props (already contains AlpineExpressions)
	result := make(map[string]any, len(spreadProps)+len(regularProps))

	// Copy spread props to result
	for key, value := range spreadProps {
		result[key] = value
	}

	log.Printf("mergePropsForRuntime: starting with %d spread props", len(spreadProps))

	// Process regular props (these override spread props)
	for _, prop := range regularProps {
		propValue := resolvePropValueForRuntime(prop, dataScope)
		result[prop.Name] = propValue
		log.Printf("mergePropsForRuntime: added/override prop %q = %v (type: %T)", prop.Name, propValue, propValue)
	}

	log.Printf("mergePropsForRuntime: final result has %d props", len(result))
	return result
}

// resolvePropValueForRuntime extracts prop value for runtime components
//
// Pattern: Helper Function [Load: 8]
// Cognitive Load: 8 (expression detection: 4, type checking: 2, wrapping: 2)
//
// This function is similar to resolvePropValue but designed for runtime components.
// Dynamic props ({var}) are returned as AlpineExpression instead of resolved values.
//
// Example:
//
//	prop := ComponentProp{Name: "content", Value: "{content}", IsDynamic: true}
//	resolvePropValueForRuntime(prop, dataScope)
//	// Returns: AlpineExpression{Expr: "content"}
func resolvePropValueForRuntime(prop ast.ComponentProp, dataScope map[string]any) any {
	if prop.IsDynamic {
		// For dynamic props ({var}), extract the variable name or expression
		varName := strings.TrimSpace(strings.Trim(prop.Value, "{}"))

		// Check if this is a quoted string literal like {"Bo"}
		if (strings.HasPrefix(varName, `"`) && strings.HasSuffix(varName, `"`)) ||
			(strings.HasPrefix(varName, `'`) && strings.HasSuffix(varName, `'`)) {
			// This is a string literal expression - return as-is (Alpine will eval)
			return AlpineExpression{Expr: varName}
		}

		// For runtime components, ALL dynamic props should be kept as expressions
		// This allows Alpine.js to evaluate them at runtime from the parent scope
		log.Printf("resolvePropValueForRuntime: prop %q = AlpineExpression(%q)", prop.Name, varName)
		return AlpineExpression{Expr: varName}
	}

	// Static prop value - return as-is (will be JSON serialized)
	return prop.Value
}

// convertPropsMapToComponentProps converts a map of props to ComponentProp slice
//
// Pattern: Helper Function [Load: 8]
// Cognitive Load: 8 (map iteration: 2, type conversion: 3, JSON encoding: 3)
//
// This is needed to reuse the existing transformComponent function which expects
// []ast.ComponentProp rather than map[string]any.
//
// CRITICAL FIX: This function now uses JSON encoding for complex types (maps, slices)
// instead of fmt.Sprintf("%v"), which was causing Go map syntax to appear in the output.
//
// Example:
//
//	Input:  map[string]any{
//	          "title": "Hello",
//	          "content": map[string]any{"components": []any{...}},
//	          "count": 42,
//	        }
//	Output: []ComponentProp{
//	          {Name: "title", Value: "Hello"},
//	          {Name: "content", Value: `{"components":[...]}`},
//	          {Name: "count", Value: "42"},
//	        }
func convertPropsMapToComponentProps(propsMap map[string]any) []ast.ComponentProp {
	props := make([]ast.ComponentProp, 0, len(propsMap))

	for name, value := range propsMap {
		// Add debug logging to track where Go map syntax comes from
		log.Printf("[CONTENT DEBUG] convertPropsMapToComponentProps: prop %q type=%T", name, value)

		// Convert value to string representation
		var valueStr string
		switch v := value.(type) {
		case string:
			valueStr = v
			log.Printf("[CONTENT DEBUG] convertPropsMapToComponentProps: prop %q = string: %q", name, valueStr)
		case int, int32, int64:
			valueStr = fmt.Sprintf("%d", v)
			log.Printf("[CONTENT DEBUG] convertPropsMapToComponentProps: prop %q = int: %q", name, valueStr)
		case float32, float64:
			valueStr = fmt.Sprintf("%f", v)
			log.Printf("[CONTENT DEBUG] convertPropsMapToComponentProps: prop %q = float: %q", name, valueStr)
		case bool:
			valueStr = fmt.Sprintf("%t", v)
			log.Printf("[CONTENT DEBUG] convertPropsMapToComponentProps: prop %q = bool: %q", name, valueStr)
		default:
			// CRITICAL FIX: Use JSON encoding for complex types instead of fmt.Sprintf("%v")
			// This prevents Go map syntax (map[key:value]) from appearing in the output
			if isComplexObject(v) {
				jsonBytes, err := json.Marshal(v)
				if err != nil {
					log.Printf("[CONTENT DEBUG] convertPropsMapToComponentProps: ERROR marshaling prop %q: %v", name, err)
					valueStr = "null"
				} else {
					valueStr = string(jsonBytes)
					log.Printf("[CONTENT DEBUG] convertPropsMapToComponentProps: prop %q = complex object (JSON): %q", name, valueStr)
				}
			} else {
				// For other types (pointers, functions, etc.), use fmt.Sprintf
				valueStr = fmt.Sprintf("%v", v)
				log.Printf("[CONTENT DEBUG] convertPropsMapToComponentProps: prop %q = other type: %q", name, valueStr)
			}
		}

		props = append(props, ast.ComponentProp{
			Name:      name,
			Value:     valueStr,
			IsDynamic: false, // Already resolved
		})
	}

	return props
}

// createDynamicByNamePlaceholder creates a placeholder for unresolved dynamic components
//
// Pattern: Helper Function [Load: 5]
// Cognitive Load: 5 (comment creation: 2, element creation: 3)
//
// Returns a placeholder element with x-component attribute for runtime resolution
// or debugging. This is used when component lookup fails.
//
// Example:
//
//	createDynamicByNamePlaceholder(node, "Component 'Foo' not found")
//	// Returns: <div x-component="ERROR: Component 'Foo' not found"></div>
func createDynamicByNamePlaceholder(node *ast.DynamicComponentByNameNode, message string) []ast.Node {
	log.Printf("createDynamicByNamePlaceholder: %s", message)

	// Create comment node with error message
	comment := &ast.CommentNode{
		Content: fmt.Sprintf(" Dynamic component error: %s (name=%q) ", message, node.NameExpression),
	}

	// Also create a placeholder element for visibility
	placeholder := &ast.Element{
		TagName: "div",
		Attributes: []ast.Attribute{
			{
				Name:  "x-component",
				Value: node.NameExpression,
			},
			{
				Name:  "data-error",
				Value: message,
			},
		},
		Children: []ast.Node{
			&ast.TextNode{
				Content: fmt.Sprintf("[Component Error: %s]", message),
			},
		},
	}

	return []ast.Node{comment, placeholder}
}
