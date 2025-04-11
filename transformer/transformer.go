package transformer

import (
	"log"
	"strings"

	"github.com/jimafisk/custom_go_template/ast"
)

// TransformAST is the main entry point for AST transformation
func TransformAST(template *ast.Template, props map[string]any) *ast.Template {
	// Reset component tracking for each transformation
	resetComponentTracking()

	// Initialize data scope with props
	dataScope := InitDataScope(props)

	// Add verbose logging for debugging
	log.Printf("TransformAST: Initialized data scope with props: %v", props)

	// Collect variables from fence section if present
	if fenceNode := FindFenceSection(template.RootNodes); fenceNode != nil {
		CollectFenceData(fenceNode, dataScope)
		log.Printf("TransformAST: Collected fence data, data scope now: %v", dataScope)
	}

	// Transform the nodes
	log.Printf("TransformAST: Starting node transformation")
	transformedNodes := transformNodes(template.RootNodes, dataScope, true)
	log.Printf("TransformAST: Transformation complete, generated %d nodes", len(transformedNodes))

	// Apply whitespace preservation
	transformedNodes = preserveWhitespace(transformedNodes)
	log.Printf("TransformAST: Applied whitespace preservation")

	// Create transformed template
	transformed := &ast.Template{
		RootNodes: transformedNodes,
	}

	return transformed
}

// The transformTextWithExpressions function is already implemented in expressions.go

// transformNodes recursively transforms AST nodes to their Alpine.js equivalents
func transformNodes(nodes []ast.Node, dataScope map[string]any, applyAlpineWrapper bool) []ast.Node {
	var transformedNodes []ast.Node
	var hasDataScope bool

	// Check if we need to apply Alpine wrapper based on data scope
	if len(dataScope) > 0 {
		hasDataScope = true
	}

	// First pass: transform all nodes except for applying Alpine wrapper
	for _, node := range nodes {
		switch n := node.(type) {
		case *ast.TextNode:
			// Check if the text contains double-curly braces or single braces
			if strings.Contains(n.Content, "{") || strings.Contains(n.Content, "{") {
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

			// Transform attributes
			element.Attributes = transformAttributes(element.Attributes, dataScope)

			// Create a child scope for the element's children
			// This ensures variables defined in child elements don't leak to siblings
			childScope := CreateChildScope(dataScope)

			// Recursively transform children with the child scope
			element.Children = transformNodes(element.Children, childScope, false)

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
			log.Printf("transformNodes: Transforming Loop node")
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
			AddExprVarsToScope(cleanedExpr, dataScope)

			// Create a span with x-text for the expression
			exprElement := &ast.Element{
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
				Children:    []ast.Node{},
				SelfClosing: false,
			}

			transformedNodes = append(transformedNodes, exprElement)

		case *ast.ComponentNode:
			// Transform component nodes
			log.Printf("transformNodes: Transforming Component node")
			componentNodes := transformComponent(n, dataScope)
			transformedNodes = append(transformedNodes, componentNodes...)

		default:
			// For any other node types, pass through unchanged
			transformedNodes = append(transformedNodes, node)
		}
	}

	// Fix nested loops and template nesting issues
	transformedNodes = ensureProperNesting(transformedNodes)

	// Check if we need to apply Alpine wrapper
	if applyAlpineWrapper && hasDataScope && needsAlpineWrapper(transformedNodes) {
		log.Printf("transformNodes: Applying Alpine wrapper with data scope: %v", dataScope)

		// Ensure all variables used in expressions are in the data scope
		ensureVariablesInScope(transformedNodes, dataScope)

		// Create Alpine wrapper with the data scope
		alpineWrapper := createAlpineWrapper(dataScope, transformedNodes)

		// Return the wrapped nodes
		return []ast.Node{alpineWrapper}
	}

	// Return the transformed nodes without wrapper
	return transformedNodes
}

// needsAlpineWrapper determines if nodes need Alpine.js data wrapper
func needsAlpineWrapper(nodes []ast.Node) bool {
	// If there are no nodes, no wrapper needed
	if len(nodes) == 0 {
		return false
	}

	// Check if there's already an Alpine.js wrapper
	for _, node := range nodes {
		if element, ok := node.(*ast.Element); ok {
			for _, attr := range element.Attributes {
				if attr.IsAlpine && attr.AlpineType == "data" {
					// If there's already an x-data attribute, no wrapper needed
					return false
				}
			}
		}
	}

	// Check if any node contains expressions or Alpine directives
	for _, node := range nodes {
		switch n := node.(type) {
		case *ast.TextNode:
			// Check if text contains expressions
			if containsExpression(n.Content) {
				return true
			}
		case *ast.Element:
			// Check if element has Alpine directives
			hasAlpineDirective := false
			for _, attr := range n.Attributes {
				if attr.IsAlpine || attr.Dynamic {
					hasAlpineDirective = true
					break
				}
			}

			// If this element has Alpine directives but no x-data,
			// we need a wrapper
			if hasAlpineDirective {
				hasXData := false
				for _, attr := range n.Attributes {
					if attr.IsAlpine && attr.AlpineType == "data" {
						hasXData = true
						break
					}
				}

				if !hasXData {
					return true
				}
			}

			// We don't need to check children if the element itself has x-data
			// as that creates its own Alpine.js scope
			hasXData := false
			for _, attr := range n.Attributes {
				if attr.IsAlpine && attr.AlpineType == "data" {
					hasXData = true
					break
				}
			}

			if !hasXData {
				// Only recursively check children if this element doesn't have x-data
				if needsAlpineWrapper(n.Children) {
					return true
				}
			}
		}
	}

	return false
}

// containsExpression checks if a text contains expressions like {variable}
func containsExpression(text string) bool {
	// Simple check for curly braces
	for i := 0; i < len(text)-1; i++ {
		if text[i] == '{' && i+1 < len(text) && text[i+1] != '{' {
			return true
		}
	}
	return false
}

// createAlpineWrapper creates an Alpine.js data wrapper element
func createAlpineWrapper(dataScope map[string]any, children []ast.Node) *ast.Element {
	// Create the wrapper element using wrapWithAlpineData
	return wrapWithAlpineData(children, dataScope)
}

// transformAttributes transforms element attributes
func transformAttributes(attributes []ast.Attribute, dataScope map[string]any) []ast.Attribute {
	transformedAttributes := make([]ast.Attribute, len(attributes))

	for i, attr := range attributes {
		// Copy the attribute
		transformedAttr := attr

		// If it's an Alpine attribute, process it
		if attr.IsAlpine {
			// For x-data, we don't modify the value
			if attr.AlpineType == "data" {
				// Keep as is
			} else {
				// For other Alpine directives, we might need to transform expressions
				// This is a placeholder for more complex attribute transformation
			}
		}

		transformedAttributes[i] = transformedAttr
	}

	return transformedAttributes
}
