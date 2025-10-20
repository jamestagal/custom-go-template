package transformer

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/jimafisk/custom_go_template/ast"
)

// TransformAST transforms the AST to Alpine.js compatible nodes
func TransformAST(template *ast.Template, props map[string]any) *ast.Template {
	// Reset component tracking for each transformation
	resetComponentTracking()

	// Reset the component template registry
	resetComponentTemplateRegistry()

	// Initialize the data scope with the provided props
	dataScope := InitDataScope(props)

	// Find fence section if it exists
	fence := FindFenceSection(template.RootNodes)
	if fence != nil {
		// Initialize store tracking with fence stores (Task 2.4)
		InitStoreTracking(fence.Stores)
		log.Printf("TransformAST: Initialized store tracking with %d store definitions", len(fence.Stores))

		// Collect data from fence section
		CollectFenceData(fence, dataScope)
		log.Printf("TransformAST: Collected fence data, data scope now: %v", dataScope)
	} else {
		// No fence section, initialize empty store tracking
		InitStoreTracking(map[string]string{})
	}

	// Start the transformation process
	log.Printf("TransformAST: Starting node transformation")

	// Transform the root nodes (not in literal context)
	transformedNodes := transformNodes(template.RootNodes, dataScope, true, false)

	// Create a new template with the transformed nodes
	transformedTemplate := &ast.Template{
		RootNodes: transformedNodes,
	}

	// Apply whitespace preservation
	transformedTemplate.RootNodes = preserveWhitespace(transformedTemplate.RootNodes)
	log.Printf("TransformAST: Applied whitespace preservation")

	log.Printf("TransformAST: Transformation complete, generated %d nodes", len(transformedNodes))

	return transformedTemplate
}

// The transformTextWithExpressions function is already implemented in expressions.go

// isLiteralContentElement checks if an element's content should be treated as literal (not transformed)
// Elements like <pre>, <code>, <textarea> should display their content as-is
// Cognitive Load: 3 (simple string comparison)
func isLiteralContentElement(tagName string) bool {
	tag := strings.ToLower(tagName)
	return tag == "pre" || tag == "code" || tag == "textarea" || tag == "script" || tag == "style"
}

// transformNodes recursively transforms AST nodes to their Alpine.js equivalents
// inLiteralContext: when true, text content is not transformed (for <pre>, <code>, etc.)
func transformNodes(nodes []ast.Node, dataScope map[string]any, applyAlpineWrapper bool, inLiteralContext bool) []ast.Node {
	var transformedNodes []ast.Node
	var hasDataScope bool

	// Check if we need to apply Alpine wrapper based on data scope
	if len(dataScope) > 0 {
		hasDataScope = true
	}

	// First pass: transform all nodes except for applying Alpine wrapper
	for i, node := range nodes {
		// DEBUG: Log node type at the beginning
		log.Printf("transformNodes: [%d] Processing node type: %T", i, node)

		switch n := node.(type) {
		case *ast.TextNode:
			// CRITICAL FIX: Skip transformation if we're in a literal content context
			if inLiteralContext {
				// Pass through as-is without any transformation
				transformedNodes = append(transformedNodes, n)
			} else if strings.Contains(n.Content, "{") || strings.Contains(n.Content, "{") {
				// Transform text nodes with expressions
				textNodes := transformTextWithExpressions(n.Content, dataScope)
				transformedNodes = append(transformedNodes, textNodes...)
			} else {
				// No expressions, pass through as is
				transformedNodes = append(transformedNodes, n)
			}

		case *ast.Element:
			// Create a copy of the element to modify
			element := *n

			// Transform attributes (now handles both regular and store expressions)
			element.Attributes = transformAttributes(element.Attributes, dataScope)

			// Create a child scope for the element's children
			// This ensures variables defined in child elements don't leak to siblings
			childScope := CreateChildScope(dataScope)

			// CRITICAL FIX: Check if this element requires literal content handling
			childInLiteralContext := isLiteralContentElement(element.TagName)

			// Recursively transform children with the child scope
			// Pass the literal context flag to children
			element.Children = transformNodes(element.Children, childScope, false, childInLiteralContext)

			// Merge any new variables back to parent scope
			MergeScopes(dataScope, childScope)

			// Add the transformed element
			transformedNodes = append(transformedNodes, &element)

		case *ast.FenceSection:
			// Skip fence sections in the output
			log.Printf("transformNodes: Skipping FenceSection")
			continue

		case *ast.Conditional:
			// Transform conditional nodes (if/else/else-if)
			log.Printf("transformNodes: Transforming Conditional node")
			conditionalNodes := transformConditional(n, dataScope)
			transformedNodes = append(transformedNodes, conditionalNodes...)

		case *ast.Loop:
			// Transform loop nodes
			log.Printf("transformNodes: *** FOUND LOOP NODE *** iterator=%s, collection=%s", n.Iterator, n.Collection)
			loopNodes := transformLoop(n, dataScope)
			transformedNodes = append(transformedNodes, loopNodes...)

		case *ast.ExpressionNode:
			// Transform expression nodes
			log.Printf("transformNodes: Transforming Expression node")
			// Clean the expression by removing any extra curly braces
			cleanedExpr := n.Expression
			cleanedExpr = strings.TrimPrefix(cleanedExpr, "{")
			cleanedExpr = strings.TrimSuffix(cleanedExpr, "}")
			cleanedExpr = strings.TrimSpace(cleanedExpr)

			// Add variables from the expression to the data scope
			extractVariablesFromExpr(cleanedExpr, dataScope)

			// Create an Alpine.js x-text element
			xTextElement := &ast.Element{
				TagName: "span",
				Attributes: []ast.Attribute{
					{
						Name:       "x-text",
						Value:      cleanedExpr,
						Dynamic:    true,
						IsAlpine:   true,
						AlpineType: "text",
					},
				},
			}
			transformedNodes = append(transformedNodes, xTextElement)

		case *ast.ComponentNode:
			// Transform component nodes
			log.Printf("transformNodes: Transforming Component node")
			componentNode := transformComponent(n, dataScope)
			if componentNode != nil {
				transformedNodes = append(transformedNodes, componentNode...)
			}

		case *ast.DynamicComponent:
			// Transform dynamic component nodes
			log.Printf("transformNodes: Transforming DynamicComponent node")
			dcNodes := transformDynamicComponent(n, dataScope)
			transformedNodes = append(transformedNodes, dcNodes...)

		case *ast.DynamicComponentByName:
			// Transform Component:dynamic nodes using runtime resolution system
			log.Printf("transformNodes: Transforming DynamicComponentByName node")
			dcnNodes := transformDynamicComponentByName(n, dataScope)
			transformedNodes = append(transformedNodes, dcnNodes...)

		case *ast.CommentNode:
			// Preserve HTML comments
			transformedNodes = append(transformedNodes, n)

		default:
			// For any other node types, just add them as is
			log.Printf("transformNodes: Unhandled node type: %T", node)
			transformedNodes = append(transformedNodes, node)
		}
	}

	// If we need to apply Alpine wrapper and we have data scope
	if applyAlpineWrapper && hasDataScope {
		return applyAlpineDataWrapper(transformedNodes, dataScope)
	}

	return transformedNodes
}

// applyAlpineDataWrapper wraps the nodes in an Alpine.js x-data wrapper
func applyAlpineDataWrapper(nodes []ast.Node, dataScope map[string]any) []ast.Node {
	// Build the x-data value from the data scope
	xDataValue := buildXDataValue(dataScope)

	// Create a wrapper div with the x-data directive
	wrapperDiv := &ast.Element{
		TagName: "div",
		Attributes: []ast.Attribute{
			{
				Name:       "x-data",
				Value:      xDataValue,
				Dynamic:    true,
				IsAlpine:   true,
				AlpineType: "data",
			},
		},
		Children: nodes,
	}

	return []ast.Node{wrapperDiv}
}

// buildXDataValue creates the Alpine.js x-data attribute value from the data scope
func buildXDataValue(dataScope map[string]any) string {
	if len(dataScope) == 0 {
		return "{}"
	}

	var parts []string
	for key, value := range dataScope {
		parts = append(parts, fmt.Sprintf("%s: %v", key, formatValue(value)))
	}

	return "{" + strings.Join(parts, ", ") + "}"
}

// formatValue formats a Go value for JavaScript
func formatValue(v interface{}) string {
	if v == nil {
		return "null"
	}

	switch val := v.(type) {
	case string:
		return fmt.Sprintf("'%s'", val)
	case bool:
		return fmt.Sprintf("%t", val)
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", val)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", val)
	case float32, float64:
		return fmt.Sprintf("%f", val)
	default:
		// For complex types, just use %v
		return fmt.Sprintf("%v", val)
	}
}

// extractAlpineDataScope extracts the x-data scope from a fence section
func extractAlpineDataScope(fence *ast.FenceSection) map[string]any {
	scope := make(map[string]any)

	// Extract props as data
	for name, value := range fence.Props {
		scope[name] = value
	}

	// Extract variables
	for name, value := range fence.Variables {
		scope[name] = value
	}

	return scope
}

// isWhitespaceOnly checks if a string contains only whitespace
func isWhitespaceOnly(s string) bool {
	return strings.TrimSpace(s) == ""
}

// preserveWhitespace ensures that meaningful whitespace is preserved
func preserveWhitespace(nodes []ast.Node) []ast.Node {
	for i, node := range nodes {
		if elem, ok := node.(*ast.Element); ok {
			// Recursively preserve whitespace in children
			elem.Children = preserveWhitespace(elem.Children)
			nodes[i] = elem
		}
	}
	return nodes
}

// InitDataScope initializes the data scope with props
func InitDataScope(props map[string]any) map[string]any {
	dataScope := make(map[string]any)
	for key, value := range props {
		dataScope[key] = value
	}
	return dataScope
}

// CreateChildScope creates a new scope that inherits from the parent scope
func CreateChildScope(parent map[string]any) map[string]any {
	child := make(map[string]any)
	// Copy all parent variables to child
	for key, value := range parent {
		child[key] = value
	}
	return child
}

// MergeScopes merges variables from the child scope back to the parent
func MergeScopes(parent, child map[string]any) {
	for key, value := range child {
		// Only merge if the key didn't exist in parent
		// This prevents overwriting parent variables
		if _, exists := parent[key]; !exists {
			parent[key] = value
		}
	}
}

// cleanExpression removes extra braces and whitespace from an expression
func cleanExpression(expr string) string {
	expr = strings.TrimPrefix(expr, "{")
	expr = strings.TrimSuffix(expr, "}")
	return strings.TrimSpace(expr)
}

// isOperator checks if a string is a JavaScript operator
func isOperator(s string) bool {
	operators := []string{
		"&&", "||", "!", "==", "!=", "===", "!==",
		"<", ">", "<=", ">=", "+", "-", "*", "/", "%",
		"&", "|", "^", "~", "<<", ">>", ">>>",
	}

	for _, op := range operators {
		if s == op {
			return true
		}
	}

	return false
}

// extractVariablesFromExpr extracts variable names from an expression and adds them to the data scope
func extractVariablesFromExpr(expr string, dataScope map[string]any) {
	// Skip empty expressions
	if strings.TrimSpace(expr) == "" {
		return
	}

	// Use a more sophisticated approach: tokenize the expression
	// Split by common delimiters: spaces, operators, parentheses, etc.
	re := regexp.MustCompile(`[a-zA-Z_$][a-zA-Z0-9_$]*`)
	tokens := re.FindAllString(expr, -1)

	for _, token := range tokens {
		// Skip JavaScript keywords and common functions
		if isJavaScriptKeyword(token) || isFunctionCall(expr, token) {
			continue
		}

		// Add to data scope if not already present
		if _, exists := dataScope[token]; !exists {
			dataScope[token] = nil
		}
	}
}

// isJavaScriptKeyword checks if a string is a JavaScript keyword
func isJavaScriptKeyword(s string) bool {
	keywords := []string{
		"if", "else", "for", "while", "do", "switch", "case", "default",
		"break", "continue", "return", "function", "var", "let", "const",
		"true", "false", "null", "undefined", "typeof", "instanceof",
		"new", "this", "super", "class", "extends", "static", "async", "await",
		"try", "catch", "finally", "throw", "in", "of", "as", "from",
	}

	for _, keyword := range keywords {
		if s == keyword {
			return true
		}
	}

	return false
}

// isFunctionCall checks if a token is followed by an opening parenthesis in the expression
func isFunctionCall(expr, token string) bool {
	// Find the token in the expression
	index := strings.Index(expr, token)
	if index == -1 {
		return false
	}

	// Check if it's followed by '(' (possibly with whitespace)
	remaining := expr[index+len(token):]
	remaining = strings.TrimSpace(remaining)

	return strings.HasPrefix(remaining, "(")
}

// FindFenceSection finds the fence section in the root nodes
func FindFenceSection(nodes []ast.Node) *ast.FenceSection {
	for _, node := range nodes {
		if fence, ok := node.(*ast.FenceSection); ok {
			return fence
		}
	}
	return nil
}

// CollectFenceData extracts data from the fence section and adds it to the data scope
func CollectFenceData(fence *ast.FenceSection, dataScope map[string]any) {
	// Add props to data scope
	for name, value := range fence.Props {
		dataScope[name] = value
	}

	// Add variables to data scope
	for name, value := range fence.Variables {
		dataScope[name] = value
	}

	// Add functions to data scope
	for name := range fence.Functions {
		// Functions are handled differently - we just mark their presence
		dataScope[name] = nil
	}
}
