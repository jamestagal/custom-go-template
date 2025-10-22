package transformer

import (
	"log"
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
			element.Attributes = transformAttributeExpressions(element.Attributes, dataScope)

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

		case *ast.DynamicComponentNode:
			// Transform dynamic component nodes
			log.Printf("transformNodes: Transforming DynamicComponent node")
			dcNodes := transformDynamicComponent(n, dataScope)
			transformedNodes = append(transformedNodes, dcNodes...)

		case *ast.DynamicComponentByNameNode:
			// Transform Component:dynamic nodes using runtime resolution system
			log.Printf("transformNodes: Transforming DynamicComponentByName node")
			dcnNodes := TransformDynamicComponentByName(n, dataScope)
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
// Uses alpineDataFormatter from alpine.go which properly formats JavaScript values
func applyAlpineDataWrapper(nodes []ast.Node, dataScope map[string]any) []ast.Node {
	// Build the x-data value from the data scope using proper JavaScript formatting
	// alpineDataFormatter uses FormatGoValueToJS which handles arrays, objects, etc.
	xDataValue := alpineDataFormatter(dataScope)

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
