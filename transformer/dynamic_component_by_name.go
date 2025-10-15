package transformer

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/analyzer"
)

// TransformDynamicComponentByName transforms <Component:dynamic> nodes into rendered components
//
// Pattern: Service Implementation Pattern [Load: 25]
// Cognitive Load: 25 (name evaluation: 6, component lookup: 5, prop merging: 8, transformation: 6)
//
// This function implements the dynamic component rendering feature for Plenti-style iteration:
//
// PHASE 1: Scope Analysis (COGNITIVE LOAD: 4) - NEW in Phase 2
//   Initialize ScopeAnalyzer to distinguish build-time vs runtime expressions
//
// PHASE 2: Evaluate name expression (COGNITIVE LOAD: 6)
//   Example: "component.name" → resolve from dataScope → "Hero2436"
//
// PHASE 3: Build-time vs Runtime Decision (COGNITIVE LOAD: 5) - NEW in Phase 2
//   if analyzer.IsRuntimeExpression(node.NameExpression):
//     → emit runtime wrapper (Task 2.3)
//   else:
//     → proceed with build-time resolution
//
// PHASE 4: Look up component template (COGNITIVE LOAD: 5)
//   componentTemplate := GetComponentTemplate(resolvedName)
//   If not found: return placeholder with warning
//
// PHASE 5: Build component props (COGNITIVE LOAD: 8)
//   a. Start with empty props map
//   b. Process spread props (left to right)
//   c. Process regular props (left to right)
//   d. Later props override earlier
//
// PHASE 6: Transform component (COGNITIVE LOAD: 6)
//   return transformComponent(componentTemplate, mergedProps, dataScope)
//
// Example transformation:
//   Input:  <Component:dynamic name={component.name} {...component.fields} theme="dark" />
//   Output: <div x-data='{...}' class="hero">...</div> (rendered component)
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

	// PHASE 6: Transform component (COGNITIVE LOAD: 6)
	// Create a regular ComponentNode to reuse existing transformation logic
	componentNode := &ast.ComponentNode{
		Name:  componentName,
		Props: convertPropsMapToComponentProps(mergedProps),
	}

	log.Printf("TransformDynamicComponentByName: transforming component %q with %d props",
		componentName, len(componentNode.Props))

	return transformComponent(componentNode, dataScope)
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
//   <div class="dyn-comp-runtime"
//        x-data="{compName: component.name, compProps: {...}}"
//        x-init="$renderDynamicComponent($el, compName, compProps)">
//   </div>
//
// Example:
//   Input:  nameExpression="component.name", props={theme: "dark"}, spread={...component.fields}
//   Output: <div class="dyn-comp-runtime" x-data="{compName: component.name, compProps: {theme: 'dark', ...}}" x-init="...">
//
// IMPORTANT: The wrapper has NO children - Alpine.js will populate it at runtime
func emitRuntimeWrapper(node *ast.DynamicComponentByNameNode, dataScope map[string]any) []ast.Node {
	log.Printf("emitRuntimeWrapper: creating runtime wrapper for nameExpr=%q", node.NameExpression)

	// PHASE 1: Resolve props for runtime (COGNITIVE LOAD: 5)
	// Merge spread props and regular props
	spreadPropsMap := resolveSpreadProps(node.SpreadProps, dataScope)
	mergedProps := mergeProps(spreadPropsMap, node.Props, dataScope)

	// Serialize props to JSON
	propsJSON := serializePropsForRuntime(mergedProps)

	// PHASE 2: Build x-data attribute value (COGNITIVE LOAD: 4)
	// Format: {compName: expression, compProps: {...}}
	xDataValue := fmt.Sprintf("{compName: %s, compProps: %s}",
		node.NameExpression, // Keep expression as-is for Alpine to evaluate
		propsJSON,           // Serialized props object
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

// serializePropsForRuntime serializes props map to JSON string for x-data attribute
//
// Pattern: Helper Function [Load: 10]
// Cognitive Load: 10 (JSON marshaling: 4, error handling: 2, string formatting: 4)
//
// This function converts a map of props into a JSON string suitable for embedding
// in an HTML attribute (x-data). It properly handles nested objects, arrays,
// and all JSON types.
//
// Example:
//   Input:  map[string]interface{}{"title": "Hello", "count": 42, "active": true}
//   Output: `{"title":"Hello","count":42,"active":true}`
//
// IMPORTANT: The output is valid JSON but Alpine expressions (x-text, x-bind values)
// should NOT be escaped as they need to be evaluated by Alpine.js at runtime.
func serializePropsForRuntime(props map[string]interface{}) string {
	if len(props) == 0 {
		return "{}"
	}

	// Use json.Marshal for proper JSON serialization
	jsonBytes, err := json.Marshal(props)
	if err != nil {
		log.Printf("serializePropsForRuntime: failed to marshal props: %v", err)
		return "{}"
	}

	// Convert to string
	jsonStr := string(jsonBytes)

	log.Printf("serializePropsForRuntime: serialized %d props to JSON: %s", len(props), jsonStr)

	return jsonStr
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
//   dataScope := map[string]any{"component": map[string]any{"name": "Hero2436"}}
//   evaluateNameExpression("component.name", dataScope)
//   // Returns: "Hero2436", nil
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
//   dataScope := map[string]any{
//     "component": map[string]any{
//       "fields": map[string]any{"title": "Hello"},
//     },
//   }
//   resolveNestedPropertyAccess("component.fields.title", dataScope)
//   // Returns: "Hello"
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
//   spreadExprs := []string{"component.fields", "overrides"}
//   dataScope := map[string]any{
//     "component": map[string]any{
//       "fields": map[string]any{"title": "Hello", "link": "/about"},
//     },
//     "overrides": map[string]any{"title": "Override"},
//   }
//   resolveSpreadProps(spreadExprs, dataScope)
//   // Returns: map[string]any{"title": "Override", "link": "/about"}
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
//   spreadProps := map[string]any{"title": "Spread Title", "link": "/spread"}
//   regularProps := []ast.ComponentProp{
//     {Name: "title", Value: "Override Title", IsDynamic: false},
//     {Name: "theme", Value: "dark", IsDynamic: false},
//   }
//   mergeProps(spreadProps, regularProps, dataScope)
//   // Returns: {"title": "Override Title", "link": "/spread", "theme": "dark"}
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
		propValue := resolvePropValue(prop, dataScope)
		result[prop.Name] = propValue
		log.Printf("mergeProps: added/override prop %q = %v", prop.Name, propValue)
	}

	log.Printf("mergeProps: final result has %d props", len(result))
	return result
}

// convertPropsMapToComponentProps converts a map of props to ComponentProp slice
//
// Pattern: Helper Function [Load: 5]
// Cognitive Load: 5 (map iteration: 2, type conversion: 3)
//
// This is needed to reuse the existing transformComponent function which expects
// []ast.ComponentProp rather than map[string]any.
func convertPropsMapToComponentProps(propsMap map[string]any) []ast.ComponentProp {
	props := make([]ast.ComponentProp, 0, len(propsMap))

	for name, value := range propsMap {
		// Convert value to string representation
		var valueStr string
		switch v := value.(type) {
		case string:
			valueStr = v
		case int, int32, int64:
			valueStr = fmt.Sprintf("%d", v)
		case float32, float64:
			valueStr = fmt.Sprintf("%f", v)
		case bool:
			valueStr = fmt.Sprintf("%t", v)
		default:
			valueStr = fmt.Sprintf("%v", v)
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
//   createDynamicByNamePlaceholder(node, "Component 'Foo' not found")
//   // Returns: <div x-component="ERROR: Component 'Foo' not found"></div>
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
